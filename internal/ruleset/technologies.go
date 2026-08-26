package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const TechnologiesSchemaVersion = 1

type TechnologiesFile struct {
	SchemaVersion int          `json:"schema_version"`
	Ruleset       string       `json:"ruleset"`
	Sources       []Source     `json:"sources"`
	Technologies  []Technology `json:"technologies"`
}

type Technology struct {
	ID           string          `json:"id"`
	Order        int             `json:"order"`
	TechnologyID int             `json:"technology_id"`
	NameKey      string          `json:"name_key"`
	NameSource   FieldProvenance `json:"name_source"`
}

func LoadTechnologies(path string) (*TechnologiesFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file TechnologiesFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *TechnologiesFile) Validate() error {
	if f.SchemaVersion != TechnologiesSchemaVersion {
		return fmt.Errorf("unsupported technologies schema version %d", f.SchemaVersion)
	}
	if f.Ruleset == "" {
		return fmt.Errorf("ruleset is required")
	}
	if len(f.Technologies) != 203 {
		return fmt.Errorf("expected 203 technologies, got %d", len(f.Technologies))
	}
	ids := make(map[string]struct{}, len(f.Technologies))
	keys := make(map[string]struct{}, len(f.Technologies))
	for order, tech := range f.Technologies {
		if tech.ID == "" || tech.NameKey == "" || tech.NameSource.SourceID == "" {
			return fmt.Errorf("technology id, name_key and source are required")
		}
		if tech.Order != order || tech.TechnologyID != order+1 {
			return fmt.Errorf("technology %q order/id mismatch: order=%d technology_id=%d expected=%d/%d", tech.ID, tech.Order, tech.TechnologyID, order, order+1)
		}
		if _, exists := ids[tech.ID]; exists {
			return fmt.Errorf("duplicate technology id %q", tech.ID)
		}
		ids[tech.ID] = struct{}{}
		if _, exists := keys[tech.NameKey]; exists {
			return fmt.Errorf("duplicate technology name key %q", tech.NameKey)
		}
		keys[tech.NameKey] = struct{}{}
	}
	return nil
}
