package supportcatalog

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
)

const SchemaVersion = 1

type Options struct {
	Clean bool
}

type Manifest struct {
	SchemaVersion int          `json:"schema_version"`
	SourceRoot    string       `json:"source_root"`
	GeneratedAt   time.Time    `json:"generated_at"`
	FileCount     int          `json:"file_count"`
	TotalBytes    int64        `json:"total_bytes"`
	Files         []FileRecord `json:"files"`
}

type FileRecord struct {
	Path       string `json:"path"`
	Extension  string `json:"extension,omitempty"`
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

	manifest := &Manifest{SchemaVersion: SchemaVersion, SourceRoot: sourceAbs, GeneratedAt: time.Now().UTC()}
	err = filepath.WalkDir(sourceAbs, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || strings.EqualFold(filepath.Ext(entry.Name()), ".lbx") {
			return nil
		}
		rel, err := filepath.Rel(sourceAbs, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		relOut := filepath.ToSlash(filepath.Join("files", filepath.FromSlash(rel)))
		absOut := filepath.Join(outAbs, filepath.FromSlash(relOut))
		if err := os.MkdirAll(filepath.Dir(absOut), 0o755); err != nil {
			return err
		}
		hash, err := copyAndHash(path, absOut)
		if err != nil {
			return err
		}
		manifest.Files = append(manifest.Files, FileRecord{Path: rel, Extension: strings.ToLower(filepath.Ext(rel)), Size: info.Size(), SHA256: hash, OutputPath: relOut})
		manifest.FileCount++
		manifest.TotalBytes += info.Size()
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(manifest.Files, func(i, j int) bool { return manifest.Files[i].Path < manifest.Files[j].Path })
	if err := writeJSON(filepath.Join(outAbs, "manifest.json"), manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func copyAndHash(source, destination string) (string, error) {
	src, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer src.Close()
	dst, err := os.Create(destination)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(dst, h), src); err != nil {
		_ = dst.Close()
		return "", err
	}
	if err := dst.Close(); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
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

func VerifyCopy(root string, record FileRecord) error {
	path := filepath.Join(root, filepath.FromSlash(record.OutputPath))
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != record.SHA256 {
		return fmt.Errorf("hash mismatch for %s: %s != %s", record.Path, got, record.SHA256)
	}
	return nil
}
