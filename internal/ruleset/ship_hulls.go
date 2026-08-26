package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const ShipHullsSchemaVersion = 2

type ShipHullsFile struct {
	SchemaVersion int        `json:"schema_version"`
	Ruleset       string     `json:"ruleset"`
	Sources       []Source   `json:"sources"`
	Hulls         []ShipHull `json:"hulls"`
}

type ShipHull struct {
	ID                       string          `json:"id"`
	SizeIndex                int             `json:"size_index"`
	NameKey                  string          `json:"name_key"`
	NameSource               FieldProvenance `json:"name_source"`
	StrategicPictureIDs      []int           `json:"strategic_picture_ids"`
	PictureLogicVerification string          `json:"picture_logic_verification"`
	PictureLogicSource       FieldProvenance `json:"picture_logic_source"`
	StrategicAssetKey        string          `json:"strategic_asset_key"`
	TacticalAssetKey         string          `json:"tactical_asset_key"`
}

func LoadShipHulls(path string) (*ShipHullsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file ShipHullsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *ShipHullsFile) Validate() error {
	if f.SchemaVersion != ShipHullsSchemaVersion {
		return fmt.Errorf("unsupported ship hulls schema version %d", f.SchemaVersion)
	}
	if f.Ruleset == "" {
		return fmt.Errorf("ruleset is required")
	}
	if len(f.Hulls) != 6 {
		return fmt.Errorf("expected 6 player ship hulls, got %d", len(f.Hulls))
	}
	ids := make(map[string]struct{}, len(f.Hulls))
	keys := make(map[string]struct{}, len(f.Hulls))
	pictureIDs := make(map[int]string)
	for index, hull := range f.Hulls {
		if hull.ID == "" || hull.NameKey == "" || hull.StrategicAssetKey == "" || hull.TacticalAssetKey == "" || hull.PictureLogicVerification == "" {
			return fmt.Errorf("ship hull %d has incomplete identity fields", index)
		}
		if hull.SizeIndex != index {
			return fmt.Errorf("ship hull %q size_index=%d, expected %d", hull.ID, hull.SizeIndex, index)
		}
		if hull.NameSource.SourceID == "" || hull.PictureLogicSource.SourceID == "" {
			return fmt.Errorf("ship hull %q has incomplete provenance", hull.ID)
		}
		if _, exists := ids[hull.ID]; exists {
			return fmt.Errorf("duplicate ship hull id %q", hull.ID)
		}
		ids[hull.ID] = struct{}{}
		if _, exists := keys[hull.NameKey]; exists {
			return fmt.Errorf("duplicate ship hull name key %q", hull.NameKey)
		}
		keys[hull.NameKey] = struct{}{}

		wantCount := 8
		if hull.SizeIndex == 5 {
			wantCount = 1
		}
		if len(hull.StrategicPictureIDs) != wantCount {
			return fmt.Errorf("ship hull %q has %d picture ids, expected %d", hull.ID, len(hull.StrategicPictureIDs), wantCount)
		}
		for style, pictureID := range hull.StrategicPictureIDs {
			if pictureID < 0 || pictureID > 48 {
				return fmt.Errorf("ship hull %q picture id %d outside [0,48]", hull.ID, pictureID)
			}
			if previous, exists := pictureIDs[pictureID]; exists {
				return fmt.Errorf("ship hull %q reuses picture id %d already used by %s", hull.ID, pictureID, previous)
			}
			pictureIDs[pictureID] = hull.ID
			if hull.SizeIndex < 5 {
				want := hull.SizeIndex*8 + style
				if pictureID != want {
					return fmt.Errorf("ship hull %q style %d picture id=%d, expected %d", hull.ID, style, pictureID, want)
				}
			} else if pictureID != 43 {
				return fmt.Errorf("doom star picture id=%d, expected 43", pictureID)
			}
		}
	}
	return nil
}
