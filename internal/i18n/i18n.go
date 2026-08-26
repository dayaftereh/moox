package i18n

import (
	"encoding/json"
	"fmt"
	"os"
)

const SchemaVersion = 1

type Source struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Description  string `json:"description,omitempty"`
	Archive      string `json:"archive,omitempty"`
	Block        *int   `json:"block,omitempty"`
	SHA256       string `json:"sha256,omitempty"`
	BlockSHA256  string `json:"block_sha256,omitempty"`
	AccessedDate string `json:"accessed_date,omitempty"`
}

type File struct {
	SchemaVersion int               `json:"schema_version"`
	Locale        string            `json:"locale"`
	Fallback      string            `json:"fallback,omitempty"`
	Sources       []Source          `json:"sources,omitempty"`
	Strings       map[string]string `json:"strings"`
}

func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file File
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	if err := file.Validate(); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *File) Validate() error {
	if f.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported language schema version %d", f.SchemaVersion)
	}
	if f.Locale == "" {
		return fmt.Errorf("locale is required")
	}
	if len(f.Strings) == 0 {
		return fmt.Errorf("language file %q has no strings", f.Locale)
	}
	for key, value := range f.Strings {
		if key == "" {
			return fmt.Errorf("language file %q contains an empty key", f.Locale)
		}
		if value == "" {
			return fmt.Errorf("language key %q is empty", key)
		}
	}
	return nil
}

func (f *File) Text(key string) (string, bool) {
	value, ok := f.Strings[key]
	return value, ok
}
