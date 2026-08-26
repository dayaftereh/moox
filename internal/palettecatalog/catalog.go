package palettecatalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"moox/internal/lbx"
	"moox/internal/moo2gfx"
)

const SchemaVersion = 1

type Options struct {
	Clean bool
}

type Manifest struct {
	SchemaVersion int             `json:"schema_version"`
	SourceRoot    string          `json:"source_root"`
	GeneratedAt   time.Time       `json:"generated_at"`
	PaletteCount  int             `json:"palette_count"`
	Palettes      []PaletteRecord `json:"palettes"`
}

type PaletteRecord struct {
	ID            string `json:"id"`
	SourceArchive string `json:"source_archive"`
	SourceBlock   int    `json:"source_block"`
	ArchiveSHA256 string `json:"archive_sha256"`
	BlockSHA256   string `json:"block_sha256"`
	PaletteSHA256 string `json:"palette_sha256"`
	BlockBytes    int    `json:"block_bytes"`
	PaletteBytes  int    `json:"palette_bytes"`
	TailBytes     int    `json:"tail_bytes"`
	TailSHA256    string `json:"tail_sha256,omitempty"`
	JSONPath      string `json:"json_path"`
	SwatchPath    string `json:"swatch_path"`
}

type PaletteFile struct {
	SchemaVersion int                       `json:"schema_version"`
	ID            string                    `json:"id"`
	SourceArchive string                    `json:"source_archive"`
	SourceBlock   int                       `json:"source_block"`
	Colors        [256]moo2gfx.PaletteColor `json:"colors"`
}

type sourceSpec struct {
	Archive string
	First   int
	Last    int
}

var sources = []sourceSpec{
	{Archive: "FONTS.LBX", First: 1, Last: 13},
	{Archive: "IFONTS.LBX", First: 1, Last: 4},
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

	manifest := &Manifest{SchemaVersion: SchemaVersion, SourceRoot: sourceAbs, GeneratedAt: time.Now().UTC()}
	for _, spec := range sources {
		archivePath := filepath.Join(sourceAbs, spec.Archive)
		archive, err := lbx.Open(archivePath)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", spec.Archive, err)
		}
		archiveHash, err := sha256File(archivePath)
		if err != nil {
			return nil, err
		}
		for blockIndex := spec.First; blockIndex <= spec.Last; blockIndex++ {
			if blockIndex >= len(archive.Entries) {
				return nil, fmt.Errorf("%s block %d missing; archive has %d blocks", spec.Archive, blockIndex, len(archive.Entries))
			}
			block, err := archive.ReadEntry(blockIndex)
			if err != nil {
				return nil, err
			}
			palette, err := moo2gfx.ParseExternalPalette(block)
			if err != nil {
				return nil, fmt.Errorf("parse %s block %d: %w", spec.Archive, blockIndex, err)
			}
			id := fmt.Sprintf("%s_%03d", strings.ToLower(strings.TrimSuffix(spec.Archive, filepath.Ext(spec.Archive))), blockIndex)
			jsonRel := id + ".json"
			pngRel := id + ".png"
			paletteFile := PaletteFile{SchemaVersion: SchemaVersion, ID: id, SourceArchive: spec.Archive, SourceBlock: blockIndex, Colors: palette.Colors}
			if err := writeJSON(filepath.Join(outAbs, jsonRel), paletteFile); err != nil {
				return nil, err
			}
			if err := writePNG(filepath.Join(outAbs, pngRel), palette.Swatch(12)); err != nil {
				return nil, err
			}

			blockSum := sha256.Sum256(block)
			paletteSum := sha256.Sum256(block[:moo2gfx.ExternalPaletteBytes])
			record := PaletteRecord{
				ID:            id,
				SourceArchive: spec.Archive,
				SourceBlock:   blockIndex,
				ArchiveSHA256: archiveHash,
				BlockSHA256:   hex.EncodeToString(blockSum[:]),
				PaletteSHA256: hex.EncodeToString(paletteSum[:]),
				BlockBytes:    len(block),
				PaletteBytes:  moo2gfx.ExternalPaletteBytes,
				TailBytes:     len(block) - moo2gfx.ExternalPaletteBytes,
				JSONPath:      filepath.ToSlash(jsonRel),
				SwatchPath:    filepath.ToSlash(pngRel),
			}
			if record.TailBytes > 0 {
				tailSum := sha256.Sum256(block[moo2gfx.ExternalPaletteBytes:])
				record.TailSHA256 = hex.EncodeToString(tailSum[:])
			}
			manifest.Palettes = append(manifest.Palettes, record)
		}
	}
	manifest.PaletteCount = len(manifest.Palettes)
	if err := writeJSON(filepath.Join(outAbs, "manifest.json"), manifest); err != nil {
		return nil, err
	}
	return manifest, nil
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

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func sha256File(path string) (string, error) {
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
