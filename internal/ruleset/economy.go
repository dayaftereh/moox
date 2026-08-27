package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const EconomySchemaVersion = 1

type EconomyFile struct {
	SchemaVersion            int                        `json:"schema_version"`
	Ruleset                  string                     `json:"ruleset"`
	Sources                  []Source                   `json:"sources"`
	BaseResearchPerScientist EconomyScalar              `json:"base_research_per_scientist"`
	BaseTaxBCPerPopulation   EconomyScalar              `json:"base_tax_bc_per_population"`
	MineralIndustryPerWorker []MineralIndustryPerWorker `json:"mineral_industry_per_worker"`
	AquaticFoodBonus         EconomyClimateBonus        `json:"aquatic_food_bonus"`
}

type EconomyScalar struct {
	Value    int    `json:"value"`
	SourceID string `json:"source_id"`
}

type MineralIndustryPerWorker struct {
	MineralID string `json:"mineral_id"`
	Value     int    `json:"value"`
	SourceID  string `json:"source_id"`
}

type EconomyClimateBonus struct {
	Value      int      `json:"value"`
	ClimateIDs []string `json:"climate_ids"`
	SourceID   string   `json:"source_id"`
}

func LoadEconomy(path string) (*EconomyFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file EconomyFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	if err := file.Validate(); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *EconomyFile) Validate() error {
	if f.SchemaVersion != EconomySchemaVersion {
		return fmt.Errorf("unsupported economy schema version %d", f.SchemaVersion)
	}
	if f.Ruleset == "" {
		return fmt.Errorf("ruleset is required")
	}
	sources := make(map[string]struct{}, len(f.Sources))
	for _, source := range f.Sources {
		if source.ID == "" || source.Type == "" {
			return fmt.Errorf("economy source id and type are required")
		}
		if _, exists := sources[source.ID]; exists {
			return fmt.Errorf("duplicate economy source %q", source.ID)
		}
		sources[source.ID] = struct{}{}
	}
	validateScalar := func(name string, scalar EconomyScalar, min int) error {
		if scalar.Value < min {
			return fmt.Errorf("%s value %d below %d", name, scalar.Value, min)
		}
		if _, ok := sources[scalar.SourceID]; !ok {
			return fmt.Errorf("%s references unknown source %q", name, scalar.SourceID)
		}
		return nil
	}
	if err := validateScalar("base_research_per_scientist", f.BaseResearchPerScientist, 1); err != nil {
		return err
	}
	if err := validateScalar("base_tax_bc_per_population", f.BaseTaxBCPerPopulation, 0); err != nil {
		return err
	}
	if len(f.MineralIndustryPerWorker) != 5 {
		return fmt.Errorf("mineral industry count=%d, expected 5", len(f.MineralIndustryPerWorker))
	}
	seen := make(map[string]struct{}, len(f.MineralIndustryPerWorker))
	for i, item := range f.MineralIndustryPerWorker {
		if item.MineralID == "" || item.Value < 1 {
			return fmt.Errorf("mineral_industry_per_worker[%d] has invalid id/value", i)
		}
		if _, exists := seen[item.MineralID]; exists {
			return fmt.Errorf("duplicate mineral industry id %q", item.MineralID)
		}
		seen[item.MineralID] = struct{}{}
		if _, ok := sources[item.SourceID]; !ok {
			return fmt.Errorf("mineral industry %q references unknown source %q", item.MineralID, item.SourceID)
		}
	}
	if f.AquaticFoodBonus.Value <= 0 || len(f.AquaticFoodBonus.ClimateIDs) == 0 {
		return fmt.Errorf("aquatic_food_bonus requires positive value and climates")
	}
	if _, ok := sources[f.AquaticFoodBonus.SourceID]; !ok {
		return fmt.Errorf("aquatic_food_bonus references unknown source %q", f.AquaticFoodBonus.SourceID)
	}
	climates := make(map[string]struct{}, len(f.AquaticFoodBonus.ClimateIDs))
	for _, climateID := range f.AquaticFoodBonus.ClimateIDs {
		if climateID == "" {
			return fmt.Errorf("aquatic_food_bonus has empty climate id")
		}
		if _, exists := climates[climateID]; exists {
			return fmt.Errorf("aquatic_food_bonus has duplicate climate %q", climateID)
		}
		climates[climateID] = struct{}{}
	}
	return nil
}
