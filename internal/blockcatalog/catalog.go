package blockcatalog

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

const SchemaVersion = 1

type Manifest struct {
	SchemaVersion int            `json:"schema_version"`
	SourceRoot    string         `json:"source_root"`
	GeneratedAt   time.Time      `json:"generated_at"`
	BlockCount    int            `json:"block_count"`
	Classes       map[string]int `json:"classes"`
	Blocks        []BlockRecord  `json:"blocks"`
}

type BlockRecord struct {
	Archive        string   `json:"archive"`
	Block          int      `json:"block"`
	Size           int64    `json:"size"`
	SHA256         string   `json:"sha256"`
	Class          string   `json:"class"`
	Evidence       []string `json:"evidence,omitempty"`
	HeadHex        string   `json:"head_hex,omitempty"`
	PrintableRatio float64  `json:"printable_ratio,omitempty"`
	ASCIIStrings   int      `json:"ascii_strings,omitempty"`
	FixedBodyBytes int      `json:"fixed_body_bytes,omitempty"`
	ArrayCount     int      `json:"array_count,omitempty"`
	RecordSize     int      `json:"record_size,omitempty"`
	SlotSize       int      `json:"slot_size,omitempty"`
	NonemptySlots  int      `json:"nonempty_slots,omitempty"`
	Graphic        *Graphic `json:"graphic,omitempty"`
}

type Graphic struct {
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Frames      int    `json:"frames"`
	Flags       uint16 `json:"flags"`
	PaletteMode string `json:"palette_mode"`
}

func Build(sourceRoot string) (*Manifest, error) {
	abs, err := filepath.Abs(sourceRoot)
	if err != nil {
		return nil, err
	}
	files, err := findLBX(abs)
	if err != nil {
		return nil, err
	}
	manifest := &Manifest{SchemaVersion: SchemaVersion, SourceRoot: abs, GeneratedAt: time.Now().UTC(), Classes: make(map[string]int)}
	for _, path := range files {
		kind, err := lbx.Detect(path)
		if err != nil || kind != lbx.KindLBX {
			continue
		}
		archive, err := lbx.Open(path)
		if err != nil {
			return nil, err
		}
		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return nil, err
		}
		rel = filepath.ToSlash(rel)
		for i := range archive.Entries {
			data, err := archive.ReadEntry(i)
			if err != nil {
				return nil, fmt.Errorf("read %s block %d: %w", rel, i, err)
			}
			rec := classify(rel, i, data)
			manifest.Blocks = append(manifest.Blocks, rec)
			manifest.BlockCount++
			manifest.Classes[rec.Class]++
		}
	}
	sort.Slice(manifest.Blocks, func(i, j int) bool {
		if manifest.Blocks[i].Archive != manifest.Blocks[j].Archive {
			return manifest.Blocks[i].Archive < manifest.Blocks[j].Archive
		}
		return manifest.Blocks[i].Block < manifest.Blocks[j].Block
	})
	return manifest, nil
}

func classify(archive string, block int, data []byte) BlockRecord {
	sum := sha256.Sum256(data)
	rec := BlockRecord{Archive: archive, Block: block, Size: int64(len(data)), SHA256: hex.EncodeToString(sum[:]), HeadHex: headHex(data, 16)}
	rec.PrintableRatio, rec.ASCIIStrings = textMetrics(data)

	if len(data) == 0 {
		rec.Class = "empty"
		rec.Evidence = []string{"zero-length block"}
		return rec
	}

	if isKnownExternalPalette(archive, block) {
		rec.Class = "external_palette"
		rec.Evidence = []string{"known FONTS/IFONTS palette block position", "at least 1024 bytes for 256 DAC-RGB entries"}
		return rec
	}
	if isKnownFontData(archive, block) {
		rec.Class = "font_data"
		rec.Evidence = []string{"documented FONTS/IFONTS block 0 font-data position"}
		return rec
	}

	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE" {
		rec.Class = "riff_wave"
		rec.Evidence = []string{"RIFF signature", "WAVE form type"}
		return rec
	}

	if g, err := moo2gfx.Parse(data); err == nil {
		paletteMode := "external"
		if g.Flags&moo2gfx.FlagInternalPalette != 0 {
			paletteMode = "internal_or_mixed"
		}
		rec.Class = "graphic"
		rec.Evidence = []string{"valid MOO2 graphic header and monotonic frame offset table"}
		rec.Graphic = &Graphic{Width: g.Width, Height: g.Height, Frames: g.FrameCount, Flags: g.Flags, PaletteMode: paletteMode}
		return rec
	}

	if len(data) >= 4 && binary.LittleEndian.Uint16(data[0:2]) == 1 {
		body := int(binary.LittleEndian.Uint16(data[2:4]))
		if body == len(data)-4 {
			rec.Class = "fixed_record_v1"
			rec.FixedBodyBytes = body
			rec.Evidence = []string{"uint16 type=1", "uint16 body_size exactly matches block_size-4"}
			if rec.ASCIIStrings > 0 {
				rec.Evidence = append(rec.Evidence, "contains printable ASCII string runs")
			}
			return rec
		}
	}
	if len(data) >= 4 {
		count := int(binary.LittleEndian.Uint16(data[0:2]))
		recordSize := int(binary.LittleEndian.Uint16(data[2:4]))
		if count > 0 && recordSize > 0 && 4+count*recordSize == len(data) {
			rec.Class = "fixed_array_v1"
			rec.ArrayCount = count
			rec.RecordSize = recordSize
			rec.Evidence = []string{"uint16 count", "uint16 record_size", "4 + count*record_size exactly matches block size"}
			if rec.ASCIIStrings > 0 {
				rec.Evidence = append(rec.Evidence, "contains printable ASCII string runs")
			}
			return rec
		}
	}

	if slotSize, nonempty := detectFixedASCIISlots(data); slotSize > 0 {
		rec.Class = "fixed_ascii_slots_candidate"
		rec.SlotSize = slotSize
		rec.NonemptySlots = nonempty
		rec.Evidence = []string{"block divides into fixed-size ASCII/NUL slots", "multiple non-empty slots"}
		return rec
	}

	if rec.PrintableRatio >= 0.85 && rec.ASCIIStrings >= 1 {
		rec.Class = "ascii_blob_candidate"
		rec.Evidence = []string{"at least 85% printable ASCII", "contains a printable ASCII run"}
		return rec
	}

	if looksLikeSignedByteTable(data) {
		rec.Class = "signed_byte_table_candidate"
		rec.Evidence = []string{"small binary block", "most bytes interpreted as signed int8 are in range -32..32", "little/no printable text"}
		return rec
	}

	if rec.PrintableRatio >= 0.60 && rec.ASCIIStrings >= 2 {
		rec.Class = "string_table_candidate"
		rec.Evidence = []string{"at least 60% printable ASCII", "at least two printable ASCII runs"}
		return rec
	}

	rec.Class = "unknown"
	return rec
}

func isKnownExternalPalette(archive string, block int) bool {
	name := strings.ToUpper(filepath.Base(archive))
	return (name == "FONTS.LBX" && block >= 1 && block <= 13) || (name == "IFONTS.LBX" && block >= 1 && block <= 4)
}
func isKnownFontData(archive string, block int) bool {
	name := strings.ToUpper(filepath.Base(archive))
	return block == 0 && (name == "FONTS.LBX" || name == "IFONTS.LBX")
}

func detectFixedASCIISlots(data []byte) (slotSize, nonempty int) {
	runs := textscan.ASCII(data, 2)
	if len(runs) < 2 {
		return 0, 0
	}
	for _, b := range data {
		if b != 0 && (b < 0x20 || b > 0x7e) {
			return 0, 0
		}
	}
	for _, size := range []int{8, 10, 12, 15, 16, 20, 24, 32, 40, 64} {
		if len(data) < size*4 || len(data)%size != 0 {
			continue
		}
		aligned := 0
		maxRun := 0
		for _, run := range runs {
			if run.Offset%size == 0 {
				aligned++
			}
			if len(run.Value) > maxRun {
				maxRun = len(run.Value)
			}
		}
		if maxRun > size || float64(aligned)/float64(len(runs)) < 0.90 {
			continue
		}
		used := 0
		for offset := 0; offset < len(data); offset += size {
			nonzero := false
			for _, b := range data[offset : offset+size] {
				if b != 0 {
					nonzero = true
					break
				}
			}
			if nonzero {
				used++
			}
		}
		if used >= 2 {
			return size, used
		}
	}
	return 0, 0
}
func looksLikeSignedByteTable(data []byte) bool {
	if len(data) < 16 || len(data) > 256 {
		return false
	}
	if len(textscan.ASCII(data, 4)) != 0 {
		return false
	}
	inRange := 0
	nonzero := 0
	for _, b := range data {
		v := int(int8(b))
		if v >= -32 && v <= 32 {
			inRange++
		}
		if b != 0 {
			nonzero++
		}
	}
	return nonzero >= 8 && float64(inRange)/float64(len(data)) >= 0.90
}

func textMetrics(data []byte) (float64, int) {
	if len(data) == 0 {
		return 0, 0
	}
	printable := 0
	for _, b := range data {
		if b >= 0x20 && b <= 0x7e {
			printable++
		}
	}
	return float64(printable) / float64(len(data)), len(textscan.ASCII(data, 4))
}

func headHex(data []byte, n int) string {
	if len(data) < n {
		n = len(data)
	}
	return strings.ToUpper(hex.EncodeToString(data[:n]))
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

func WriteJSON(path string, manifest *Manifest) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	f, err := os.Create(abs)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(manifest); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
