package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const PlanetClassesSchemaVersion = 1

type PlanetClassesFile struct {
	SchemaVersion  int             `json:"schema_version"`
	Ruleset        string          `json:"ruleset"`
	Sources        []Source        `json:"sources"`
	Sizes          []PlanetSize    `json:"sizes"`
	MineralClasses []MineralClass  `json:"mineral_classes"`
	GravityClasses []GravityClass  `json:"gravity_classes"`
	Climates       []PlanetClimate `json:"climates"`
}

type PlanetSize struct {
	ID                           string          `json:"id"`
	Index                        int             `json:"index"`
	NameKey                      string          `json:"name_key"`
	NameSource                   FieldProvenance `json:"name_source"`
	GenerationRollUpperThreshold int             `json:"generation_roll_upper_threshold"`
	GenerationSource             FieldProvenance `json:"generation_source"`
}

type MineralClass struct {
	ID               string          `json:"id"`
	Index            int             `json:"index"`
	NameKey          string          `json:"name_key"`
	NameSource       FieldProvenance `json:"name_source"`
	BaseExtraction   int             `json:"base_extraction"`
	ExtractionSource FieldProvenance `json:"extraction_source"`
}

type GravityClass struct {
	ID         string          `json:"id"`
	Index      int             `json:"index"`
	NameKey    string          `json:"name_key"`
	NameSource FieldProvenance `json:"name_source"`
}

type PlanetClimate struct {
	ID                string          `json:"id"`
	Index             int             `json:"index"`
	NameKey           string          `json:"name_key"`
	NameSource        FieldProvenance `json:"name_source"`
	BaseFoodPerFarmer int             `json:"base_food_per_farmer"`
	FoodSource        FieldProvenance `json:"food_source"`
}

func LoadPlanetClasses(path string) (*PlanetClassesFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file PlanetClassesFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *PlanetClassesFile) Validate() error {
	if f.SchemaVersion != PlanetClassesSchemaVersion {
		return fmt.Errorf("unsupported planet classes schema version %d", f.SchemaVersion)
	}
	if f.Ruleset == "" {
		return fmt.Errorf("ruleset is required")
	}
	if len(f.Sizes) != 5 || len(f.MineralClasses) != 5 || len(f.GravityClasses) != 3 || len(f.Climates) != 10 {
		return fmt.Errorf("planet class counts sizes/minerals/gravity/climates=%d/%d/%d/%d, expected 5/5/3/10", len(f.Sizes), len(f.MineralClasses), len(f.GravityClasses), len(f.Climates))
	}
	seen := make(map[string]struct{}, 23)
	lastThreshold := 0
	for index, item := range f.Sizes {
		if item.Index != index || item.ID == "" || item.NameKey == "" || item.NameSource.SourceID == "" || item.GenerationSource.SourceID == "" {
			return fmt.Errorf("planet size %d has invalid identity/provenance", index)
		}
		if _, ok := seen[item.ID]; ok {
			return fmt.Errorf("duplicate planet class id %q", item.ID)
		}
		seen[item.ID] = struct{}{}
		if item.GenerationRollUpperThreshold <= lastThreshold || item.GenerationRollUpperThreshold > 10 {
			return fmt.Errorf("planet size %q generation threshold=%d is not an increasing value in [1,10]", item.ID, item.GenerationRollUpperThreshold)
		}
		lastThreshold = item.GenerationRollUpperThreshold
	}
	if lastThreshold != 10 {
		return fmt.Errorf("planet size generation thresholds do not cover roll 10")
	}
	for index, item := range f.MineralClasses {
		if item.Index != index || item.ID == "" || item.NameKey == "" || item.NameSource.SourceID == "" || item.ExtractionSource.SourceID == "" {
			return fmt.Errorf("mineral class %d has invalid identity/provenance", index)
		}
		if item.BaseExtraction != index+1 {
			return fmt.Errorf("mineral class %q base_extraction=%d, expected %d", item.ID, item.BaseExtraction, index+1)
		}
		if _, ok := seen[item.ID]; ok {
			return fmt.Errorf("duplicate planet class id %q", item.ID)
		}
		seen[item.ID] = struct{}{}
	}
	for index, item := range f.GravityClasses {
		if item.Index != index || item.ID == "" || item.NameKey == "" || item.NameSource.SourceID == "" {
			return fmt.Errorf("gravity class %d has invalid identity/provenance", index)
		}
		if _, ok := seen[item.ID]; ok {
			return fmt.Errorf("duplicate planet class id %q", item.ID)
		}
		seen[item.ID] = struct{}{}
	}
	for index, item := range f.Climates {
		if item.Index != index || item.ID == "" || item.NameKey == "" || item.NameSource.SourceID == "" || item.FoodSource.SourceID == "" {
			return fmt.Errorf("planet climate %d has invalid identity/provenance", index)
		}
		if item.BaseFoodPerFarmer < 0 || item.BaseFoodPerFarmer > 3 {
			return fmt.Errorf("planet climate %q base_food_per_farmer=%d outside [0,3]", item.ID, item.BaseFoodPerFarmer)
		}
		if _, ok := seen[item.ID]; ok {
			return fmt.Errorf("duplicate planet class id %q", item.ID)
		}
		seen[item.ID] = struct{}{}
	}
	return nil
}
