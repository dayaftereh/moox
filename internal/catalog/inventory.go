package catalog

import (
	"crypto/sha256"
	"encoding/hex"
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

type Options struct {
	EntryHashes bool
}

type Inventory struct {
	Root          string       `json:"root"`
	GeneratedAt   time.Time    `json:"generated_at"`
	FileCount     int          `json:"file_count"`
	TotalBytes    int64        `json:"total_bytes"`
	LBXContainers int          `json:"lbx_containers"`
	SmackerFiles  int          `json:"smacker_files"`
	UnknownFiles  int          `json:"unknown_files"`
	Files         []FileRecord `json:"files"`
}

type FileRecord struct {
	Path      string     `json:"path"`
	Size      int64      `json:"size"`
	Extension string     `json:"extension,omitempty"`
	SHA256    string     `json:"sha256"`
	Kind      lbx.Kind   `json:"kind"`
	LBX       *LBXRecord `json:"lbx,omitempty"`
	Error     string     `json:"error,omitempty"`
}

type LBXRecord struct {
	EntryCount int           `json:"entry_count"`
	Reserved   uint32        `json:"reserved"`
	DataOffset uint32        `json:"data_offset"`
	Entries    []EntryRecord `json:"entries"`
}

type EntryRecord struct {
	Index  int    `json:"index"`
	Offset int64  `json:"offset"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256,omitempty"`
}

func Build(root string, options Options) (*Inventory, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	stat, err := os.Stat(absRoot)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("inventory root is not a directory: %s", absRoot)
	}

	inv := &Inventory{Root: absRoot, GeneratedAt: time.Now().UTC()}
	err = filepath.WalkDir(absRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}

		record := FileRecord{
			Path:      filepath.ToSlash(rel),
			Size:      info.Size(),
			Extension: strings.ToLower(filepath.Ext(path)),
		}
		record.SHA256, err = hashFile(path)
		if err != nil {
			return err
		}

		record.Kind, err = lbx.Detect(path)
		if err != nil {
			record.Kind = lbx.KindUnknown
			record.Error = err.Error()
		}

		switch record.Kind {
		case lbx.KindLBX:
			parsed, parseErr := lbx.Open(path)
			if parseErr != nil {
				record.Error = parseErr.Error()
				inv.UnknownFiles++
				break
			}
			inv.LBXContainers++
			record.LBX = &LBXRecord{
				EntryCount: len(parsed.Entries),
				Reserved:   parsed.Reserved,
				DataOffset: parsed.DataOffset,
				Entries:    make([]EntryRecord, len(parsed.Entries)),
			}
			for i, block := range parsed.Entries {
				er := EntryRecord{Index: block.Index, Offset: block.Offset, Size: block.Size}
				if options.EntryHashes {
					er.SHA256, err = parsed.EntrySHA256(i)
					if err != nil {
						return err
					}
				}
				record.LBX.Entries[i] = er
			}
		case lbx.KindSmacker:
			inv.SmackerFiles++
		default:
			inv.UnknownFiles++
		}

		inv.FileCount++
		inv.TotalBytes += info.Size()
		inv.Files = append(inv.Files, record)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(inv.Files, func(i, j int) bool { return inv.Files[i].Path < inv.Files[j].Path })
	return inv, nil
}

func hashFile(path string) (string, error) {
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
