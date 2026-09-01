package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const TacticalCombatSchemaVersion = 1

type TacticalCombatFile struct {
	SchemaVersion int                      `json:"schema_version"`
	Ruleset       string                   `json:"ruleset"`
	Initiative    TacticalInitiativeRules  `json:"initiative"`
	RNG           TacticalRNGRules         `json:"rng"`
	Beam          TacticalBeamRules        `json:"beam"`
	Weapon        TacticalWeaponRule       `json:"weapon"`
	Frigate       TacticalFrigateRule      `json:"frigate"`
	Drives        []TacticalDriveSpeedRule `json:"drives"`
	Computer      TacticalComputerRule     `json:"computer"`
}

type TacticalInitiativeRules struct {
	BeamOffenseDivisor int `json:"beam_offense_divisor"`
}

type TacticalRNGRules struct {
	Multiplier uint32 `json:"multiplier"`
	Increment  uint32 `json:"increment"`
}

type TacticalBeamRules struct {
	BaseHitThreshold     int   `json:"base_hit_threshold"`
	MaxHitThreshold      int   `json:"max_hit_threshold"`
	EffectiveRollCap     int   `json:"effective_roll_cap"`
	ToHitRangeModifiers  []int `json:"to_hit_range_modifiers"`
	DamageRangeModifiers []int `json:"damage_range_modifiers"`
}

type TacticalWeaponRule struct {
	ID           string `json:"id"`
	TechnologyID int    `json:"technology_id"`
	Kind         string `json:"kind"`
	BaseSpace    int    `json:"base_space"`
	BaseCostPP   int    `json:"base_cost_pp"`
	MinDamage    int    `json:"min_damage"`
	MaxDamage    int    `json:"max_damage"`
}

type TacticalFrigateRule struct {
	HullID     string `json:"hull_id"`
	ArmorID    string `json:"armor_id"`
	FuelCellID string `json:"fuel_cell_id"`
	ArmorHits  int    `json:"armor_hits"`
	Structure  int    `json:"structure"`
}

type TacticalDriveSpeedRule struct {
	DriveID  string `json:"drive_id"`
	MinSpeed int    `json:"min_speed"`
	MaxSpeed int    `json:"max_speed"`
}

type TacticalComputerRule struct {
	ComputerID  string `json:"computer_id"`
	BeamOffense int    `json:"beam_offense"`
}

func LoadTacticalCombat(path string) (*TacticalCombatFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file TacticalCombatFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *TacticalCombatFile) Validate() error {
	if f == nil {
		return fmt.Errorf("tactical combat rules are nil")
	}
	if f.SchemaVersion != TacticalCombatSchemaVersion {
		return fmt.Errorf("unsupported tactical combat schema version %d", f.SchemaVersion)
	}
	if f.Ruleset == "" {
		return fmt.Errorf("tactical combat ruleset is required")
	}
	if f.Initiative.BeamOffenseDivisor <= 0 {
		return fmt.Errorf("initiative beam offense divisor must be positive")
	}
	if f.RNG.Multiplier == 0 {
		return fmt.Errorf("tactical RNG multiplier must be non-zero")
	}
	if f.Beam.BaseHitThreshold <= 0 || f.Beam.MaxHitThreshold < f.Beam.BaseHitThreshold || f.Beam.EffectiveRollCap < f.Beam.MaxHitThreshold {
		return fmt.Errorf("invalid Beam hit thresholds")
	}
	if len(f.Beam.ToHitRangeModifiers) != 9 || len(f.Beam.DamageRangeModifiers) != 9 {
		return fmt.Errorf("Beam range tables must contain exactly 9 entries")
	}
	if f.Weapon.ID == "" || f.Weapon.TechnologyID <= 0 || f.Weapon.Kind != "beam" || f.Weapon.BaseSpace <= 0 || f.Weapon.BaseCostPP <= 0 || f.Weapon.MinDamage <= 0 || f.Weapon.MaxDamage < f.Weapon.MinDamage {
		return fmt.Errorf("invalid tactical weapon rule")
	}
	if f.Frigate.HullID == "" || f.Frigate.ArmorID == "" || f.Frigate.FuelCellID == "" || f.Frigate.ArmorHits <= 0 || f.Frigate.Structure <= 0 {
		return fmt.Errorf("invalid Frigate tactical rule")
	}
	if len(f.Drives) != 2 {
		return fmt.Errorf("expected exactly 2 baseline tactical drive rules, got %d", len(f.Drives))
	}
	seenDrive := make(map[string]struct{}, len(f.Drives))
	for i, drive := range f.Drives {
		if drive.DriveID == "" || drive.MinSpeed <= 0 || drive.MaxSpeed < drive.MinSpeed {
			return fmt.Errorf("invalid tactical drive[%d]", i)
		}
		if _, ok := seenDrive[drive.DriveID]; ok {
			return fmt.Errorf("duplicate tactical drive %q", drive.DriveID)
		}
		seenDrive[drive.DriveID] = struct{}{}
	}
	if f.Computer.ComputerID == "" || f.Computer.BeamOffense <= 0 {
		return fmt.Errorf("invalid tactical computer rule")
	}
	return nil
}
