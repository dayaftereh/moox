package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const TacticalCombatSchemaVersion = 1

type TacticalCombatFile struct {
	SchemaVersion int                        `json:"schema_version"`
	Ruleset       string                     `json:"ruleset"`
	Initiative    TacticalInitiativeRules    `json:"initiative"`
	RNG           TacticalRNGRules           `json:"rng"`
	Beam          TacticalBeamRules          `json:"beam"`
	Weapon        TacticalWeaponRule         `json:"weapon"`
	Weapons       []TacticalWeaponRule       `json:"weapons,omitempty"`
	BeamModifiers []TacticalBeamModifierRule `json:"beam_modifiers,omitempty"`
	Frigate       TacticalFrigateRule        `json:"frigate"`
	Drives        []TacticalDriveSpeedRule   `json:"drives"`
	Computer      TacticalComputerRule       `json:"computer"`
}

type TacticalInitiativeRules struct {
	BeamOffenseDivisor int `json:"beam_offense_divisor"`
}

type TacticalRNGRules struct {
	Multiplier uint32 `json:"multiplier"`
	Increment  uint32 `json:"increment"`
}

type TacticalBeamRules struct {
	BaseHitThreshold            int   `json:"base_hit_threshold"`
	MaxHitThreshold             int   `json:"max_hit_threshold"`
	EffectiveRollCap            int   `json:"effective_roll_cap"`
	ToHitRangeModifiers         []int `json:"to_hit_range_modifiers"`
	DamageRangeModifiers        []int `json:"damage_range_modifiers"`
	SpaceMiniaturizationFactors []int `json:"space_miniaturization_factors,omitempty"`
	CostMiniaturizationFactors  []int `json:"cost_miniaturization_factors,omitempty"`
}

type TacticalWeaponRule struct {
	ID                  string   `json:"id"`
	OriginalIndex       int      `json:"original_index,omitempty"`
	TechnologyID        int      `json:"technology_id"`
	TechnologyFieldID   int      `json:"technology_field_id,omitempty"`
	Kind                string   `json:"kind"`
	BaseSpace           int      `json:"base_space"`
	BaseCostPP          int      `json:"base_cost_pp"`
	MinDamage           int      `json:"min_damage"`
	MaxDamage           int      `json:"max_damage"`
	AllowedModifierMask int      `json:"allowed_modifier_mask,omitempty"`
	InnateEffectIDs     []string `json:"innate_effect_ids,omitempty"`
	TacticalSupport     string   `json:"tactical_support,omitempty"`
}

type TacticalBeamModifierRule struct {
	ID                string `json:"id"`
	Bit               int    `json:"bit"`
	Kind              string `json:"kind"`
	RequiredMiniLevel int    `json:"required_mini_level"`
	SpacePercentDelta int    `json:"space_percent_delta"`
	CostPercentDelta  int    `json:"cost_percent_delta"`
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

func (f *TacticalCombatFile) BeamWeapons() []TacticalWeaponRule {
	if f == nil {
		return nil
	}
	if len(f.Weapons) != 0 {
		return f.Weapons
	}
	if f.Weapon.ID == "" {
		return nil
	}
	return []TacticalWeaponRule{f.Weapon}
}

func (f *TacticalCombatFile) WeaponByID(id string) (TacticalWeaponRule, bool) {
	for _, weapon := range f.BeamWeapons() {
		if weapon.ID == id {
			return weapon, true
		}
	}
	return TacticalWeaponRule{}, false
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
	weapons := f.BeamWeapons()
	if len(weapons) == 0 {
		return fmt.Errorf("at least one Beam weapon rule is required")
	}
	seenWeapons := make(map[string]struct{}, len(weapons))
	for i, weapon := range weapons {
		if weapon.ID == "" || weapon.TechnologyID <= 0 || weapon.Kind != "beam" || weapon.BaseSpace <= 0 || weapon.BaseCostPP <= 0 || weapon.MinDamage <= 0 || weapon.MaxDamage < weapon.MinDamage {
			return fmt.Errorf("invalid Beam weapon[%d]", i)
		}
		if _, ok := seenWeapons[weapon.ID]; ok {
			return fmt.Errorf("duplicate Beam weapon %q", weapon.ID)
		}
		seenWeapons[weapon.ID] = struct{}{}
	}
	if len(f.Beam.SpaceMiniaturizationFactors) != 0 && len(f.Beam.SpaceMiniaturizationFactors) != 6 {
		return fmt.Errorf("Beam Space miniaturization table must contain exactly 6 entries")
	}
	if len(f.Beam.CostMiniaturizationFactors) != 0 && len(f.Beam.CostMiniaturizationFactors) != 6 {
		return fmt.Errorf("Beam Cost miniaturization table must contain exactly 6 entries")
	}
	seenModifierBits := map[int]struct{}{}
	seenModifierIDs := map[string]struct{}{}
	for i, modifier := range f.BeamModifiers {
		if modifier.ID == "" || (modifier.Kind != "mount" && modifier.Kind != "special") || modifier.Bit <= 0 || modifier.RequiredMiniLevel < 0 {
			return fmt.Errorf("invalid Beam modifier[%d]", i)
		}
		if _, ok := seenModifierIDs[modifier.ID]; ok {
			return fmt.Errorf("duplicate Beam modifier id %q", modifier.ID)
		}
		if _, ok := seenModifierBits[modifier.Bit]; ok {
			return fmt.Errorf("duplicate Beam modifier bit 0x%X", modifier.Bit)
		}
		seenModifierIDs[modifier.ID] = struct{}{}
		seenModifierBits[modifier.Bit] = struct{}{}
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
