package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const ShipHullsSchemaVersion = 3

type ShipHullsFile struct {
	SchemaVersion       int                     `json:"schema_version"`
	Ruleset             string                  `json:"ruleset"`
	Sources             []Source                `json:"sources"`
	Hulls               []ShipHull              `json:"hulls"`
	MandatoryComponents ShipMandatoryComponents `json:"mandatory_components"`
}

type ShipHull struct {
	ID                       string          `json:"id"`
	SizeIndex                int             `json:"size_index"`
	NameKey                  string          `json:"name_key"`
	NameSource               FieldProvenance `json:"name_source"`
	BaseCostPP               int             `json:"base_cost_pp"`
	BaseSpace                int             `json:"base_space"`
	RuntimeSource            FieldProvenance `json:"runtime_source"`
	StrategicPictureIDs      []int           `json:"strategic_picture_ids"`
	PictureLogicVerification string          `json:"picture_logic_verification"`
	PictureLogicSource       FieldProvenance `json:"picture_logic_source"`
	StrategicAssetKey        string          `json:"strategic_asset_key"`
	TacticalAssetKey         string          `json:"tactical_asset_key"`
}

type ShipMandatoryComponents struct {
	Drives    []ShipDrive    `json:"drives"`
	Computers []ShipComputer `json:"computers"`
	Armors    []ShipArmor    `json:"armors"`
	Shields   []ShipShield   `json:"shields"`
	FuelCells []ShipFuelCell `json:"fuel_cells"`
}

type ShipDrive struct {
	ID             string          `json:"id"`
	ComponentIndex int             `json:"component_index"`
	TechnologyID   int             `json:"technology_id"`
	FTLSpeed       int             `json:"ftl_speed"`
	SpaceByHull    []int           `json:"space_by_hull"`
	CostByHullPP   []int           `json:"cost_by_hull_pp"`
	Source         FieldProvenance `json:"source"`
}

type ShipComputer struct {
	ID             string          `json:"id"`
	ComponentIndex int             `json:"component_index"`
	TechnologyID   int             `json:"technology_id"`
	CostByHullPP   []int           `json:"cost_by_hull_pp"`
	Source         FieldProvenance `json:"source"`
}

type ShipArmor struct {
	ID             string          `json:"id"`
	ComponentIndex int             `json:"component_index"`
	TechnologyID   int             `json:"technology_id"`
	CostPercent    int             `json:"cost_percent"`
	Source         FieldProvenance `json:"source"`
}

type ShipShield struct {
	ID             string          `json:"id"`
	ComponentIndex int             `json:"component_index"`
	TechnologyID   int             `json:"technology_id"`
	SpaceByHull    []int           `json:"space_by_hull"`
	CostByHullPP   []int           `json:"cost_by_hull_pp"`
	Source         FieldProvenance `json:"source"`
}

type ShipFuelCell struct {
	ID             string          `json:"id"`
	ComponentIndex int             `json:"component_index"`
	TechnologyID   int             `json:"technology_id"`
	RangeParsecs   int             `json:"range_parsecs"`
	Source         FieldProvenance `json:"source"`
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
		if hull.BaseCostPP <= 0 || hull.BaseSpace <= 0 {
			return fmt.Errorf("ship hull %q has invalid runtime cost/space %d/%d", hull.ID, hull.BaseCostPP, hull.BaseSpace)
		}
		if hull.NameSource.SourceID == "" || hull.RuntimeSource.SourceID == "" || hull.PictureLogicSource.SourceID == "" {
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
	if err := validateMandatoryComponents(f.MandatoryComponents, len(f.Hulls)); err != nil {
		return err
	}
	return nil
}

func validateMandatoryComponents(c ShipMandatoryComponents, hullCount int) error {
	if len(c.Drives) != 6 || len(c.Computers) != 5 || len(c.Armors) != 6 || len(c.Shields) != 5 || len(c.FuelCells) != 5 {
		return fmt.Errorf("mandatory ship component counts drives=%d computers=%d armors=%d shields=%d fuel_cells=%d", len(c.Drives), len(c.Computers), len(c.Armors), len(c.Shields), len(c.FuelCells))
	}
	seen := make(map[string]struct{})
	check := func(id string, index, technologyID int, source FieldProvenance) error {
		if id == "" || index <= 0 || technologyID <= 0 || source.SourceID == "" {
			return fmt.Errorf("ship component %q has incomplete identity/provenance", id)
		}
		if _, ok := seen[id]; ok {
			return fmt.Errorf("duplicate ship component id %q", id)
		}
		seen[id] = struct{}{}
		return nil
	}
	checkHullValues := func(id string, values []int) error {
		if len(values) != hullCount {
			return fmt.Errorf("ship component %q has %d hull values, expected %d", id, len(values), hullCount)
		}
		for _, value := range values {
			if value < 0 {
				return fmt.Errorf("ship component %q has negative hull value %d", id, value)
			}
		}
		return nil
	}
	for i, item := range c.Drives {
		if item.ComponentIndex != i+1 || item.FTLSpeed < 2 {
			return fmt.Errorf("ship drive %q has invalid index/speed", item.ID)
		}
		if err := check(item.ID, item.ComponentIndex, item.TechnologyID, item.Source); err != nil {
			return err
		}
		if err := checkHullValues(item.ID, item.SpaceByHull); err != nil {
			return err
		}
		if err := checkHullValues(item.ID, item.CostByHullPP); err != nil {
			return err
		}
	}
	for i, item := range c.Computers {
		if item.ComponentIndex != i+1 {
			return fmt.Errorf("ship computer %q has invalid index", item.ID)
		}
		if err := check(item.ID, item.ComponentIndex, item.TechnologyID, item.Source); err != nil {
			return err
		}
		if err := checkHullValues(item.ID, item.CostByHullPP); err != nil {
			return err
		}
	}
	for i, item := range c.Armors {
		if item.ComponentIndex != i+1 || item.CostPercent < 0 {
			return fmt.Errorf("ship armor %q has invalid index/cost", item.ID)
		}
		if err := check(item.ID, item.ComponentIndex, item.TechnologyID, item.Source); err != nil {
			return err
		}
	}
	for i, item := range c.Shields {
		if item.ComponentIndex != i+1 {
			return fmt.Errorf("ship shield %q has invalid index", item.ID)
		}
		if err := check(item.ID, item.ComponentIndex, item.TechnologyID, item.Source); err != nil {
			return err
		}
		if err := checkHullValues(item.ID, item.SpaceByHull); err != nil {
			return err
		}
		if err := checkHullValues(item.ID, item.CostByHullPP); err != nil {
			return err
		}
	}
	for i, item := range c.FuelCells {
		if item.ComponentIndex != i+1 || item.RangeParsecs <= 0 {
			return fmt.Errorf("ship fuel cell %q has invalid index/range", item.ID)
		}
		if err := check(item.ID, item.ComponentIndex, item.TechnologyID, item.Source); err != nil {
			return err
		}
	}
	return nil
}
