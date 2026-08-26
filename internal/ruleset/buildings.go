package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const BuildingsSchemaVersion = 1

type BuildingsFile struct {
	SchemaVersion int        `json:"schema_version"`
	Ruleset       string     `json:"ruleset"`
	Sources       []Source   `json:"sources"`
	Buildings     []Building `json:"buildings"`
}

type Building struct {
	ID                       string          `json:"id"`
	Order                    int             `json:"order"`
	ProductionID             int             `json:"production_id"`
	ProductionIDVerification string          `json:"production_id_verification"`
	ProductionIDSource       FieldProvenance `json:"production_id_source"`
	NameKey                  string          `json:"name_key"`
	NameVerification         string          `json:"name_verification"`
	NameSource               FieldProvenance `json:"name_source"`
	ColonyReferenceAssetKey  string          `json:"colony_reference_asset_key,omitempty"`
}

func LoadBuildings(path string) (*BuildingsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file BuildingsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *BuildingsFile) Validate() error {
	if f.SchemaVersion != BuildingsSchemaVersion {
		return fmt.Errorf("unsupported buildings schema version %d", f.SchemaVersion)
	}
	if f.Ruleset == "" {
		return fmt.Errorf("ruleset is required")
	}
	if len(f.Buildings) != 48 {
		return fmt.Errorf("expected 48 colony buildings, got %d", len(f.Buildings))
	}
	ids := make(map[string]struct{}, len(f.Buildings))
	nameKeys := make(map[string]struct{}, len(f.Buildings))
	productionIDs := make(map[int]struct{}, len(f.Buildings))
	orders := make(map[int]struct{}, len(f.Buildings))
	for _, building := range f.Buildings {
		if building.ID == "" || building.NameKey == "" || building.NameVerification == "" || building.ProductionIDVerification == "" {
			return fmt.Errorf("building id, name_key and verification fields are required")
		}
		if _, exists := ids[building.ID]; exists {
			return fmt.Errorf("duplicate building id %q", building.ID)
		}
		ids[building.ID] = struct{}{}
		if _, exists := nameKeys[building.NameKey]; exists {
			return fmt.Errorf("duplicate building name key %q", building.NameKey)
		}
		nameKeys[building.NameKey] = struct{}{}
		if _, exists := productionIDs[building.ProductionID]; exists {
			return fmt.Errorf("duplicate building production id %d", building.ProductionID)
		}
		productionIDs[building.ProductionID] = struct{}{}
		if _, exists := orders[building.Order]; exists {
			return fmt.Errorf("duplicate building order %d", building.Order)
		}
		orders[building.Order] = struct{}{}
		if building.Order < 0 || building.Order >= 48 {
			return fmt.Errorf("building %q order %d outside [0,48)", building.ID, building.Order)
		}
		if building.ProductionID != building.Order+1 {
			return fmt.Errorf("building %q production_id=%d, expected order+1=%d", building.ID, building.ProductionID, building.Order+1)
		}
		if building.NameSource.SourceID == "" || building.ProductionIDSource.SourceID == "" {
			return fmt.Errorf("building %q has incomplete field provenance", building.ID)
		}
	}
	return nil
}
