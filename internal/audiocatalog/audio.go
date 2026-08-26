package audiocatalog

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"moox/internal/lbx"
)

const SchemaVersion = 1

type Options struct {
	Clean bool
}

type Manifest struct {
	SchemaVersion int           `json:"schema_version"`
	SourceRoot    string        `json:"source_root"`
	GeneratedAt   time.Time     `json:"generated_at"`
	WAVCount      int           `json:"wav_count"`
	TotalBytes    int64         `json:"total_bytes"`
	Files         []AudioRecord `json:"files"`
}

type AudioRecord struct {
	Archive       string `json:"archive"`
	Block         int    `json:"block"`
	Size          int64  `json:"size"`
	SHA256        string `json:"sha256"`
	OutputPath    string `json:"output_path"`
	FormatTag     uint16 `json:"format_tag"`
	Channels      uint16 `json:"channels"`
	SampleRate    uint32 `json:"sample_rate"`
	ByteRate      uint32 `json:"byte_rate"`
	BlockAlign    uint16 `json:"block_align"`
	BitsPerSample uint16 `json:"bits_per_sample"`
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
		archiveDir := strings.ToLower(strings.TrimSuffix(rel, filepath.Ext(rel)))
		for block := range archive.Entries {
			data, err := archive.ReadEntry(block)
			if err != nil {
				return nil, err
			}
			format, ok := parseWAVEFormat(data)
			if !ok {
				continue
			}
			relOut := filepath.ToSlash(filepath.Join(archiveDir, fmt.Sprintf("block_%04d.wav", block)))
			absOut := filepath.Join(outAbs, filepath.FromSlash(relOut))
			if err := os.MkdirAll(filepath.Dir(absOut), 0o755); err != nil {
				return nil, err
			}
			if err := os.WriteFile(absOut, data, 0o644); err != nil {
				return nil, err
			}
			sum := sha256.Sum256(data)
			rec := AudioRecord{
				Archive: rel, Block: block, Size: int64(len(data)), SHA256: hex.EncodeToString(sum[:]), OutputPath: relOut,
				FormatTag: format.FormatTag, Channels: format.Channels, SampleRate: format.SampleRate, ByteRate: format.ByteRate,
				BlockAlign: format.BlockAlign, BitsPerSample: format.BitsPerSample,
			}
			manifest.Files = append(manifest.Files, rec)
			manifest.WAVCount++
			manifest.TotalBytes += int64(len(data))
		}
	}
	sort.Slice(manifest.Files, func(i, j int) bool {
		if manifest.Files[i].Archive != manifest.Files[j].Archive {
			return manifest.Files[i].Archive < manifest.Files[j].Archive
		}
		return manifest.Files[i].Block < manifest.Files[j].Block
	})
	if err := writeJSON(filepath.Join(outAbs, "manifest.json"), manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

type waveFormat struct {
	FormatTag     uint16
	Channels      uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
}

func parseWAVEFormat(data []byte) (waveFormat, bool) {
	var out waveFormat
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return out, false
	}
	for pos := 12; pos+8 <= len(data); {
		id := string(data[pos : pos+4])
		size := int(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		body := pos + 8
		if size < 0 || body+size > len(data) {
			return out, false
		}
		if id == "fmt " {
			if size < 16 {
				return out, false
			}
			out.FormatTag = binary.LittleEndian.Uint16(data[body : body+2])
			out.Channels = binary.LittleEndian.Uint16(data[body+2 : body+4])
			out.SampleRate = binary.LittleEndian.Uint32(data[body+4 : body+8])
			out.ByteRate = binary.LittleEndian.Uint32(data[body+8 : body+12])
			out.BlockAlign = binary.LittleEndian.Uint16(data[body+12 : body+14])
			out.BitsPerSample = binary.LittleEndian.Uint16(data[body+14 : body+16])
			return out, true
		}
		pos = body + size
		if size&1 != 0 {
			pos++
		}
	}
	return out, false
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

func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
