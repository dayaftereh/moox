package graphiccatalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"moox/internal/lbx"
	"moox/internal/moo2gfx"
)

const SchemaVersion = 2

type Options struct {
	ExportPNG bool
	Clean     bool
	Progress  func(Progress)
}

type Progress struct {
	ArchiveIndex int
	ArchiveCount int
	Archive      string
	Graphics     int
	Frames       int
}

type Manifest struct {
	SchemaVersion           int             `json:"schema_version"`
	SourceRoot              string          `json:"source_root"`
	GeneratedAt             time.Time       `json:"generated_at"`
	LBXArchives             int             `json:"lbx_archives"`
	GraphicBlocks           int             `json:"graphic_blocks"`
	InternalPalette         int             `json:"internal_palette_blocks"`
	ExternalPalette         int             `json:"external_palette_blocks"`
	PaletteContextsResolved int             `json:"palette_contexts_resolved"`
	PaletteContextsPending  int             `json:"palette_contexts_pending"`
	FramesTotal             int             `json:"frames_total"`
	FramesExported          int             `json:"frames_exported"`
	FramesComplete          int             `json:"frames_complete"`
	FramesPartial           int             `json:"frames_partial_palette"`
	FramesFailed            int             `json:"frames_failed"`
	Archives                []ArchiveRecord `json:"archives"`
}

type ArchiveRecord struct {
	Path          string          `json:"path"`
	SHA256        string          `json:"sha256"`
	GraphicBlocks int             `json:"graphic_blocks"`
	Graphics      []GraphicRecord `json:"graphics"`
}

type GraphicRecord struct {
	Block                    int           `json:"block"`
	BlockSHA256              string        `json:"block_sha256"`
	Width                    int           `json:"width"`
	Height                   int           `json:"height"`
	FrameCount               int           `json:"frame_count"`
	Delay                    uint16        `json:"delay"`
	Flags                    uint16        `json:"flags"`
	Palette                  string        `json:"palette"`
	PaletteShift             int           `json:"palette_shift,omitempty"`
	PaletteEntries           int           `json:"palette_entries,omitempty"`
	PaletteContextStatus     string        `json:"palette_context_status,omitempty"`
	PaletteContextSource     string        `json:"palette_context_source,omitempty"`
	PaletteContextEvidence   string        `json:"palette_context_evidence,omitempty"`
	PaletteContextConfidence string        `json:"palette_context_confidence,omitempty"`
	Frames                   []FrameRecord `json:"frames,omitempty"`
}

type FrameRecord struct {
	Frame                int    `json:"frame"`
	Status               string `json:"status"`
	PNGPath              string `json:"png_path,omitempty"`
	PNGSHA256            string `json:"png_sha256,omitempty"`
	DrawnPixels          int    `json:"drawn_pixels,omitempty"`
	MissingPalettePixels int    `json:"missing_palette_pixels,omitempty"`
	Error                string `json:"error,omitempty"`
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
	stat, err := os.Stat(sourceAbs)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("source root is not a directory: %s", sourceAbs)
	}
	if options.Clean {
		if err := os.RemoveAll(outAbs); err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(outAbs, 0o755); err != nil {
		return nil, err
	}

	archives, err := findLBXArchives(sourceAbs)
	if err != nil {
		return nil, err
	}
	manifest := &Manifest{SchemaVersion: SchemaVersion, SourceRoot: sourceAbs, GeneratedAt: time.Now().UTC()}
	resolver := newPaletteResolver(sourceAbs)

	for archiveIndex, archivePath := range archives {
		rel, err := filepath.Rel(sourceAbs, archivePath)
		if err != nil {
			return nil, err
		}
		rel = filepath.ToSlash(rel)
		kind, err := lbx.Detect(archivePath)
		if err != nil || kind != lbx.KindLBX {
			continue
		}
		parsed, err := lbx.Open(archivePath)
		if err != nil {
			continue
		}
		archiveHash, err := sha256File(archivePath)
		if err != nil {
			return nil, err
		}
		record := ArchiveRecord{Path: rel, SHA256: archiveHash}
		for blockIndex := range parsed.Entries {
			block, err := parsed.ReadEntry(blockIndex)
			if err != nil {
				return nil, fmt.Errorf("read %s block %d: %w", rel, blockIndex, err)
			}
			graphic, err := moo2gfx.Parse(block)
			if err != nil {
				continue
			}
			sum := sha256.Sum256(block)
			gr := GraphicRecord{
				Block:       blockIndex,
				BlockSHA256: hex.EncodeToString(sum[:]),
				Width:       graphic.Width,
				Height:      graphic.Height,
				FrameCount:  graphic.FrameCount,
				Delay:       graphic.Delay,
				Flags:       graphic.Flags,
			}
			manifest.GraphicBlocks++
			manifest.FramesTotal += graphic.FrameCount
			record.GraphicBlocks++
			hasInternalPalette := graphic.Flags&moo2gfx.FlagInternalPalette != 0
			embeddedEntries := paletteEntries(graphic)
			if hasInternalPalette {
				gr.Palette = "internal"
				gr.PaletteShift = graphic.PaletteShift
				gr.PaletteEntries = embeddedEntries
				manifest.InternalPalette++
			} else {
				gr.Palette = "external"
				manifest.ExternalPalette++
			}

			needsPaletteContext := embeddedEntries < 256
			resolution := paletteResolution{}
			if needsPaletteContext {
				resolution, err = resolver.Resolve(rel, blockIndex, parsed, graphic)
				if err != nil {
					return nil, fmt.Errorf("resolve palette for %s block %d: %w", rel, blockIndex, err)
				}
			}
			if resolution.Resolved {
				gr.PaletteContextStatus = "resolved"
				gr.PaletteContextSource = resolution.Source
				gr.PaletteContextEvidence = resolution.Evidence
				gr.PaletteContextConfidence = resolution.Confidence
				manifest.PaletteContextsResolved++
			} else if needsPaletteContext {
				gr.PaletteContextStatus = "pending"
				manifest.PaletteContextsPending++
			} else {
				gr.PaletteContextStatus = "self"
			}

			canExport := hasInternalPalette || resolution.Resolved
			if canExport && options.ExportPNG {
				gr.Frames = exportFrames(outAbs, rel, blockIndex, graphic, manifest)
			} else if !canExport {
				gr.Frames = make([]FrameRecord, graphic.FrameCount)
				for frame := range gr.Frames {
					gr.Frames[frame] = FrameRecord{Frame: frame, Status: "external_palette_pending"}
				}
			}
			record.Graphics = append(record.Graphics, gr)
		}
		if record.GraphicBlocks > 0 {
			manifest.Archives = append(manifest.Archives, record)
		}
		manifest.LBXArchives++
		if options.Progress != nil {
			options.Progress(Progress{ArchiveIndex: archiveIndex + 1, ArchiveCount: len(archives), Archive: rel, Graphics: manifest.GraphicBlocks, Frames: manifest.FramesExported})
		}
	}

	sort.Slice(manifest.Archives, func(i, j int) bool { return manifest.Archives[i].Path < manifest.Archives[j].Path })
	if err := writeManifest(filepath.Join(outAbs, "manifest.json"), manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func exportFrames(outRoot, archiveRel string, blockIndex int, graphic *moo2gfx.Graphic, manifest *Manifest) []FrameRecord {
	archiveDir := strings.TrimSuffix(filepath.ToSlash(archiveRel), filepath.Ext(archiveRel))
	archiveDir = strings.ToLower(archiveDir)
	records := make([]FrameRecord, graphic.FrameCount)
	decoder := graphic.NewDisplayDecoder()
	for frame := 0; frame < graphic.FrameCount; frame++ {
		rec := FrameRecord{Frame: frame}
		img, info, err := decoder.DecodeNext()
		if err != nil {
			rec.Status = "decode_failed"
			rec.Error = err.Error()
			manifest.FramesFailed++
			records[frame] = rec
			continue
		}
		rec.DrawnPixels = info.DrawnPixels
		rec.MissingPalettePixels = info.MissingPalettePixels
		if info.MissingPalettePixels == 0 {
			rec.Status = "complete"
		} else {
			rec.Status = "partial_palette"
		}
		relPNG := filepath.ToSlash(filepath.Join(archiveDir, fmt.Sprintf("block_%04d_frame_%03d.png", blockIndex, frame)))
		absPNG := filepath.Join(outRoot, filepath.FromSlash(relPNG))
		if err := os.MkdirAll(filepath.Dir(absPNG), 0o755); err != nil {
			rec.Status = "write_failed"
			rec.Error = err.Error()
			manifest.FramesFailed++
			records[frame] = rec
			continue
		}
		file, err := os.Create(absPNG)
		if err != nil {
			rec.Status = "write_failed"
			rec.Error = err.Error()
			manifest.FramesFailed++
			records[frame] = rec
			continue
		}
		encodeErr := png.Encode(file, img)
		closeErr := file.Close()
		if encodeErr != nil || closeErr != nil {
			rec.Status = "write_failed"
			if encodeErr != nil {
				rec.Error = encodeErr.Error()
			} else {
				rec.Error = closeErr.Error()
			}
			manifest.FramesFailed++
			records[frame] = rec
			continue
		}
		rec.PNGPath = relPNG
		rec.PNGSHA256, err = sha256File(absPNG)
		if err != nil {
			rec.Status = "hash_failed"
			rec.Error = err.Error()
			manifest.FramesFailed++
			records[frame] = rec
			continue
		}
		manifest.FramesExported++
		if rec.Status == "complete" {
			manifest.FramesComplete++
		} else if rec.Status == "partial_palette" {
			manifest.FramesPartial++
		}
		records[frame] = rec
	}
	return records
}

func paletteEntries(g *moo2gfx.Graphic) int {
	count := 0
	for _, valid := range g.PaletteValid {
		if valid {
			count++
		}
	}
	return count
}

func findLBXArchives(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".lbx") {
			paths = append(paths, path)
		}
		return nil
	})
	sort.Strings(paths)
	return paths, err
}

func writeManifest(path string, manifest *Manifest) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(manifest); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
