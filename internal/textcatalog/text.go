package textcatalog

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"moox/internal/lbx"
	"moox/internal/moo2gfx"
	"moox/internal/textscan"
)

const SchemaVersion = 2

type Options struct {
	Clean bool
}

type Manifest struct {
	SchemaVersion         int          `json:"schema_version"`
	SourceRoot            string       `json:"source_root"`
	GeneratedAt           time.Time    `json:"generated_at"`
	FixedRecords          int          `json:"fixed_records"`
	FixedArrays           int          `json:"fixed_arrays"`
	StringTableCandidates int          `json:"string_table_candidates"`
	ASCIIRuns             int          `json:"ascii_runs"`
	Records               []TextRecord `json:"records"`
}

type TextRecord struct {
	Archive        string            `json:"archive"`
	Block          int               `json:"block"`
	Kind           string            `json:"kind"`
	BlockSize      int               `json:"block_size"`
	BlockSHA256    string            `json:"block_sha256"`
	BodyOffset     int               `json:"body_offset,omitempty"`
	BodySize       int               `json:"body_size,omitempty"`
	BodySHA256     string            `json:"body_sha256,omitempty"`
	ArrayCount     int               `json:"array_count,omitempty"`
	RecordSize     int               `json:"record_size,omitempty"`
	NonZeroBytes   int               `json:"nonzero_bytes"`
	HighBytes      int               `json:"high_bytes"`
	PrintableRatio float64           `json:"printable_ratio"`
	Runs           []textscan.String `json:"ascii_runs,omitempty"`
	ArrayRecords   []ArrayRecord     `json:"array_records,omitempty"`
	PreviewPath    string            `json:"preview_path,omitempty"`
}

type ArrayRecord struct {
	Index          int               `json:"index"`
	Offset         int               `json:"offset"`
	Size           int               `json:"size"`
	SHA256         string            `json:"sha256"`
	NonZeroBytes   int               `json:"nonzero_bytes"`
	HighBytes      int               `json:"high_bytes"`
	PrintableRatio float64           `json:"printable_ratio"`
	Runs           []textscan.String `json:"ascii_runs,omitempty"`
}

func Build(sourceRoot, outDir string, options Options) (*Manifest, error) {
	sourceAbs, err := filepath.Abs(sourceRoot)
	if err != nil {
		return nil, err
	}
	outAbs, err := filepath.Abs(outDir)
	if err != nil {
		return nil, err
	}
	if options.Clean {
		if err := os.RemoveAll(outAbs); err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(outAbs, 0o755); err != nil {
		return nil, err
	}
	files, err := findLBX(sourceAbs)
	if err != nil {
		return nil, err
	}
	manifest := &Manifest{SchemaVersion: SchemaVersion, SourceRoot: sourceAbs, GeneratedAt: time.Now().UTC()}
	for _, path := range files {
		kind, err := lbx.Detect(path)
		if err != nil || kind != lbx.KindLBX {
			continue
		}
		archive, err := lbx.Open(path)
		if err != nil {
			return nil, err
		}
		rel, err := filepath.Rel(sourceAbs, path)
		if err != nil {
			return nil, err
		}
		rel = filepath.ToSlash(rel)
		for block := range archive.Entries {
			data, err := archive.ReadEntry(block)
			if err != nil {
				return nil, err
			}
			rec, ok := analyze(rel, block, data)
			if !ok {
				continue
			}
			switch rec.Kind {
			case "fixed_record_v1":
				manifest.FixedRecords++
			case "fixed_array_v1":
				manifest.FixedArrays++
			case "string_table_candidate":
				manifest.StringTableCandidates++
			}
			manifest.ASCIIRuns += countRuns(rec)
			if countRuns(rec) > 0 {
				previewRel := filepath.ToSlash(filepath.Join(strings.ToLower(strings.TrimSuffix(rel, filepath.Ext(rel))), fmt.Sprintf("block_%04d.ascii.txt", block)))
				if err := writePreview(filepath.Join(outAbs, filepath.FromSlash(previewRel)), rec); err != nil {
					return nil, err
				}
				rec.PreviewPath = previewRel
			}
			manifest.Records = append(manifest.Records, rec)
		}
	}
	sort.Slice(manifest.Records, func(i, j int) bool {
		if manifest.Records[i].Archive != manifest.Records[j].Archive {
			return manifest.Records[i].Archive < manifest.Records[j].Archive
		}
		return manifest.Records[i].Block < manifest.Records[j].Block
	})
	if err := writeJSON(filepath.Join(outAbs, "manifest.json"), manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func analyze(archive string, block int, data []byte) (TextRecord, bool) {
	if len(data) == 0 || isKnownExternalPalette(archive, block) || isWave(data) {
		return TextRecord{}, false
	}
	if _, err := moo2gfx.Parse(data); err == nil {
		return TextRecord{}, false
	}
	blockSum := sha256.Sum256(data)
	rec := TextRecord{Archive: archive, Block: block, BlockSize: len(data), BlockSHA256: hex.EncodeToString(blockSum[:])}

	if len(data) >= 4 && binary.LittleEndian.Uint16(data[0:2]) == 1 {
		bodySize := int(binary.LittleEndian.Uint16(data[2:4]))
		if bodySize == len(data)-4 {
			body := data[4:]
			rec.Kind = "fixed_record_v1"
			rec.BodyOffset = 4
			rec.BodySize = bodySize
			bodySum := sha256.Sum256(body)
			rec.BodySHA256 = hex.EncodeToString(bodySum[:])
			rec.Runs = textscan.ASCII(body, 4)
			rec.NonZeroBytes, rec.HighBytes, rec.PrintableRatio = byteMetrics(body)
			return rec, true
		}
	}

	if len(data) >= 4 {
		count := int(binary.LittleEndian.Uint16(data[0:2]))
		recordSize := int(binary.LittleEndian.Uint16(data[2:4]))
		if count > 0 && recordSize > 0 && 4+count*recordSize == len(data) {
			rec.Kind = "fixed_array_v1"
			rec.ArrayCount = count
			rec.RecordSize = recordSize
			rec.NonZeroBytes, rec.HighBytes, rec.PrintableRatio = byteMetrics(data[4:])
			rec.ArrayRecords = make([]ArrayRecord, 0, count)
			totalRuns := 0
			for i := 0; i < count; i++ {
				offset := 4 + i*recordSize
				row := data[offset : offset+recordSize]
				rowSum := sha256.Sum256(row)
				ar := ArrayRecord{Index: i, Offset: offset, Size: recordSize, SHA256: hex.EncodeToString(rowSum[:])}
				ar.NonZeroBytes, ar.HighBytes, ar.PrintableRatio = byteMetrics(row)
				ar.Runs = textscan.ASCII(row, 4)
				totalRuns += len(ar.Runs)
				rec.ArrayRecords = append(rec.ArrayRecords, ar)
			}
			if totalRuns == 0 {
				return TextRecord{}, false
			}
			return rec, true
		}
	}

	rec.Runs = textscan.ASCII(data, 4)
	rec.NonZeroBytes, rec.HighBytes, rec.PrintableRatio = byteMetrics(data)
	if rec.PrintableRatio >= 0.60 && len(rec.Runs) >= 2 {
		rec.Kind = "string_table_candidate"
		return rec, true
	}
	return TextRecord{}, false
}

func countRuns(rec TextRecord) int {
	n := len(rec.Runs)
	for _, row := range rec.ArrayRecords {
		n += len(row.Runs)
	}
	return n
}

func byteMetrics(data []byte) (nonzero, high int, printableRatio float64) {
	if len(data) == 0 {
		return 0, 0, 0
	}
	printable := 0
	for _, b := range data {
		if b != 0 {
			nonzero++
		}
		if b >= 0x80 {
			high++
		}
		if b >= 0x20 && b <= 0x7e {
			printable++
		}
	}
	return nonzero, high, float64(printable) / float64(len(data))
}

func writePreview(path string, rec TextRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s block %d (%s)\n", rec.Archive, rec.Block, rec.Kind)
	fmt.Fprintf(&b, "# This is an ASCII-only research preview. Raw bytes remain authoritative.\n")
	fmt.Fprintf(&b, "# High bytes are not decoded here.\n\n")
	if len(rec.ArrayRecords) > 0 {
		for _, row := range rec.ArrayRecords {
			if len(row.Runs) == 0 {
				continue
			}
			fmt.Fprintf(&b, "## record %d offset 0x%04X size %d\n", row.Index, row.Offset, row.Size)
			for _, run := range row.Runs {
				fmt.Fprintf(&b, "0x%04X\t%s\n", row.Offset+run.Offset, run.Value)
			}
			b.WriteByte('\n')
		}
	} else {
		for _, run := range rec.Runs {
			fmt.Fprintf(&b, "0x%04X\t%s\n", run.Offset+rec.BodyOffset, run.Value)
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func isKnownExternalPalette(archive string, block int) bool {
	name := strings.ToUpper(filepath.Base(archive))
	return (name == "FONTS.LBX" && block >= 1 && block <= 13) || (name == "IFONTS.LBX" && block >= 1 && block <= 4)
}

func isWave(data []byte) bool {
	return len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WAVE"
}

func findLBX(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".lbx") {
			paths = append(paths, path)
		}
		return nil
	})
	sort.Strings(paths)
	return paths, err
}

func writeJSON(path string, value any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
