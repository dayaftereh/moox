package rawextract

import (
	"crypto/sha256"
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
	Clean    bool
	Progress func(Progress)
}

type Progress struct {
	FileIndex    int
	FileCount    int
	Source       string
	Blocks       int
	BytesWritten int64
}

type Manifest struct {
	SchemaVersion int             `json:"schema_version"`
	SourceRoot    string          `json:"source_root"`
	GeneratedAt   time.Time       `json:"generated_at"`
	LBXArchives   int             `json:"lbx_archives"`
	SmackerFiles  int             `json:"smacker_files"`
	BlockCount    int             `json:"block_count"`
	PayloadBytes  int64           `json:"payload_bytes"`
	CopiedBytes   int64           `json:"copied_bytes"`
	Archives      []ArchiveRecord `json:"archives"`
	Smacker       []SmackerRecord `json:"smacker"`
}

type ArchiveRecord struct {
	Path       string        `json:"path"`
	SHA256     string        `json:"sha256"`
	BlockCount int           `json:"block_count"`
	Blocks     []BlockRecord `json:"blocks"`
}

type BlockRecord struct {
	Index      int    `json:"index"`
	Offset     int64  `json:"offset"`
	Size       int64  `json:"size"`
	SHA256     string `json:"sha256"`
	OutputPath string `json:"output_path"`
}

type SmackerRecord struct {
	Path       string `json:"path"`
	Size       int64  `json:"size"`
	SHA256     string `json:"sha256"`
	OutputPath string `json:"output_path"`
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

	files, err := findLBXFiles(sourceAbs)
	if err != nil {
		return nil, err
	}
	manifest := &Manifest{SchemaVersion: SchemaVersion, SourceRoot: sourceAbs, GeneratedAt: time.Now().UTC()}
	for fileIndex, path := range files {
		rel, err := filepath.Rel(sourceAbs, path)
		if err != nil {
			return nil, err
		}
		rel = filepath.ToSlash(rel)
		kind, err := lbx.Detect(path)
		if err != nil {
			return nil, fmt.Errorf("detect %s: %w", rel, err)
		}
		switch kind {
		case lbx.KindLBX:
			record, err := extractArchive(path, rel, outAbs, manifest)
			if err != nil {
				return nil, err
			}
			manifest.LBXArchives++
			manifest.Archives = append(manifest.Archives, *record)
		case lbx.KindSmacker:
			record, err := copySmacker(path, rel, outAbs)
			if err != nil {
				return nil, err
			}
			manifest.SmackerFiles++
			manifest.CopiedBytes += record.Size
			manifest.Smacker = append(manifest.Smacker, *record)
		}
		if options.Progress != nil {
			options.Progress(Progress{FileIndex: fileIndex + 1, FileCount: len(files), Source: rel, Blocks: manifest.BlockCount, BytesWritten: manifest.CopiedBytes})
		}
	}
	sort.Slice(manifest.Archives, func(i, j int) bool { return manifest.Archives[i].Path < manifest.Archives[j].Path })
	sort.Slice(manifest.Smacker, func(i, j int) bool { return manifest.Smacker[i].Path < manifest.Smacker[j].Path })
	if err := writeJSON(filepath.Join(outAbs, "manifest.json"), manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func extractArchive(sourcePath, rel string, outRoot string, manifest *Manifest) (*ArchiveRecord, error) {
	archive, err := lbx.Open(sourcePath)
	if err != nil {
		return nil, err
	}
	archiveHash, err := sha256File(sourcePath)
	if err != nil {
		return nil, err
	}
	dir := strings.ToLower(strings.TrimSuffix(filepath.ToSlash(rel), filepath.Ext(rel)))
	record := &ArchiveRecord{Path: rel, SHA256: archiveHash, BlockCount: len(archive.Entries), Blocks: make([]BlockRecord, 0, len(archive.Entries))}
	for _, entry := range archive.Entries {
		relOut := filepath.ToSlash(filepath.Join("lbx", dir, fmt.Sprintf("block_%04d.bin", entry.Index)))
		absOut := filepath.Join(outRoot, filepath.FromSlash(relOut))
		if err := os.MkdirAll(filepath.Dir(absOut), 0o755); err != nil {
			return nil, err
		}
		dst, err := os.Create(absOut)
		if err != nil {
			return nil, err
		}
		h := sha256.New()
		mw := io.MultiWriter(dst, h)
		if err := archive.CopyEntry(entry.Index, mw); err != nil {
			_ = dst.Close()
			return nil, err
		}
		if err := dst.Close(); err != nil {
			return nil, err
		}
		record.Blocks = append(record.Blocks, BlockRecord{Index: entry.Index, Offset: entry.Offset, Size: entry.Size, SHA256: hex.EncodeToString(h.Sum(nil)), OutputPath: relOut})
		manifest.BlockCount++
		manifest.PayloadBytes += entry.Size
		manifest.CopiedBytes += entry.Size
	}
	return record, nil
}

func copySmacker(sourcePath, rel, outRoot string) (*SmackerRecord, error) {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return nil, err
	}
	base := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel)) + ".smk"
	relOut := filepath.ToSlash(filepath.Join("smacker", base))
	absOut := filepath.Join(outRoot, filepath.FromSlash(relOut))
	if err := os.MkdirAll(filepath.Dir(absOut), 0o755); err != nil {
		return nil, err
	}
	src, err := os.Open(sourcePath)
	if err != nil {
		return nil, err
	}
	defer src.Close()
	dst, err := os.Create(absOut)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(dst, h), src); err != nil {
		_ = dst.Close()
		return nil, err
	}
	if err := dst.Close(); err != nil {
		return nil, err
	}
	return &SmackerRecord{Path: rel, Size: info.Size(), SHA256: hex.EncodeToString(h.Sum(nil)), OutputPath: relOut}, nil
}

func findLBXFiles(root string) ([]string, error) {
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
