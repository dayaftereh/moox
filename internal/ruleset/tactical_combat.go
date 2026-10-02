package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const TacticalCombatSchemaVersion = 1

type TacticalCombatFile struct {
	SchemaVersion                  int                        `json:"schema_version"`
	Ruleset                        string                     `json:"ruleset"`
	Initiative                     TacticalInitiativeRules    `json:"initiative"`
	RNG                            TacticalRNGRules           `json:"rng"`
	Beam                           TacticalBeamRules          `json:"beam"`
	Ordnance                       TacticalOrdnanceRules      `json:"ordnance"`
	Weapon                         TacticalWeaponRule         `json:"weapon"`
	Weapons                        []TacticalWeaponRule       `json:"weapons,omitempty"`
	BeamModifiers                  []TacticalBeamModifierRule `json:"beam_modifiers,omitempty"`
	OrdnanceModifiers              []TacticalBeamModifierRule `json:"ordnance_modifiers,omitempty"`
	Frigate                        TacticalFrigateRule        `json:"frigate"`
	Hulls                          []TacticalHullRule         `json:"hulls"`
	Armors                         []TacticalArmorRule        `json:"armors"`
	Drives                         []TacticalDriveSpeedRule   `json:"drives"`
	Computer                       TacticalComputerRule       `json:"computer"`
	Computers                      []TacticalComputerRule     `json:"computers"`
	Shields                        []TacticalShieldRule       `json:"shields"`
	FuelCells                      []TacticalFuelCellRule     `json:"fuel_cells"`
	MaxUnusedSpaceCombatSpeedBonus int                        `json:"max_unused_space_combat_speed_bonus"`
	BeamDefensePerCombatSpeed      int                        `json:"beam_defense_per_combat_speed"`
	ShieldCapacityPerStrength      int                        `json:"shield_capacity_per_strength"`
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

type TacticalOrdnanceRules struct {
	MovementUnitsPerField     int `json:"movement_units_per_field"`
	FastSpeedBonus            int `json:"fast_speed_bonus"`
	TorpedoCooldownRounds     int `json:"torpedo_cooldown_rounds"`
	MIRVMultiplier            int `json:"mirv_multiplier"`
	PlasmaDissipationPerField int `json:"plasma_dissipation_per_field"`
}

type TacticalWeaponRule struct {
	ID                   string   `json:"id"`
	OriginalIndex        int      `json:"original_index,omitempty"`
	TechnologyID         int      `json:"technology_id"`
	TechnologyFieldID    int      `json:"technology_field_id,omitempty"`
	Kind                 string   `json:"kind"`
	BaseSpace            int      `json:"base_space"`
	BaseCostPP           int      `json:"base_cost_pp"`
	MinDamage            int      `json:"min_damage"`
	MaxDamage            int      `json:"max_damage"`
	AllowedModifierMask  int      `json:"allowed_modifier_mask,omitempty"`
	InnateEffectIDs      []string `json:"innate_effect_ids,omitempty"`
	TacticalSupport      string   `json:"tactical_support,omitempty"`
	CraftPerUnit         int      `json:"craft_per_unit,omitempty"`
	ProjectileSpeed      int      `json:"projectile_speed,omitempty"`
	ProjectileDurability int      `json:"projectile_durability,omitempty"`
	BeamInterceptorDCV   int      `json:"beam_interceptor_dcv,omitempty"`
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

type TacticalHullRule struct {
	HullID           string `json:"hull_id"`
	BaseArmor        int    `json:"base_armor"`
	Structure        int    `json:"structure"`
	BaseCombatSpeed  int    `json:"base_combat_speed"`
	ShieldHullFactor int    `json:"shield_hull_factor"`
	MarineMax        int    `json:"marine_max"`
}

type TacticalArmorRule struct {
	ArmorID      string `json:"armor_id"`
	BonusPercent int    `json:"bonus_percent"`
}

type TacticalDriveSpeedRule struct {
	DriveID          string `json:"drive_id"`
	MinSpeed         int    `json:"min_speed"`
	MaxSpeed         int    `json:"max_speed"`
	CombatSpeedBonus int    `json:"combat_speed_bonus"`
}

type TacticalComputerRule struct {
	ComputerID  string `json:"computer_id"`
	BeamOffense int    `json:"beam_offense"`
}

type TacticalShieldRule struct {
	ShieldID string `json:"shield_id"`
	Strength int    `json:"strength"`
}

type TacticalFuelCellRule struct {
	FuelCellID string `json:"fuel_cell_id"`
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

func (f *TacticalCombatFile) AllWeapons() []TacticalWeaponRule {
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

func (f *TacticalCombatFile) BeamWeapons() []TacticalWeaponRule {
	all := f.AllWeapons()
	out := make([]TacticalWeaponRule, 0, len(all))
	for _, weapon := range all {
		if weapon.Kind == "beam" {
			out = append(out, weapon)
		}
	}
	return out
}

func (f *TacticalCombatFile) OrdnanceWeapons() []TacticalWeaponRule {
	all := f.AllWeapons()
	out := make([]TacticalWeaponRule, 0, len(all))
	for _, weapon := range all {
		switch weapon.Kind {
		case "missile", "torpedo", "bomb":
			out = append(out, weapon)
		}
	}
	return out
}

func (f *TacticalCombatFile) WeaponByID(id string) (TacticalWeaponRule, bool) {
	for _, weapon := range f.AllWeapons() {
		if weapon.ID == id {
			return weapon, true
		}
	}
	return TacticalWeaponRule{}, false
}

func (f *TacticalCombatFile) WeaponModifiers() []TacticalBeamModifierRule {
	if f == nil {
		return nil
	}
	out := make([]TacticalBeamModifierRule, 0, len(f.BeamModifiers)+len(f.OrdnanceModifiers))
	out = append(out, f.BeamModifiers...)
	out = append(out, f.OrdnanceModifiers...)
	return out
}
func (f *TacticalCombatFile) HullByID(id string) (TacticalHullRule, bool) {
	for _, rule := range f.Hulls {
		if rule.HullID == id {
			return rule, true
		}
	}
	return TacticalHullRule{}, false
}

func (f *TacticalCombatFile) ArmorByID(id string) (TacticalArmorRule, bool) {
	for _, rule := range f.Armors {
		if rule.ArmorID == id {
			return rule, true
		}
	}
	return TacticalArmorRule{}, false
}

func (f *TacticalCombatFile) DriveByID(id string) (TacticalDriveSpeedRule, bool) {
	for _, rule := range f.Drives {
		if rule.DriveID == id {
			return rule, true
		}
	}
	return TacticalDriveSpeedRule{}, false
}

func (f *TacticalCombatFile) ComputerByID(id string) (TacticalComputerRule, bool) {
	for _, rule := range f.Computers {
		if rule.ComputerID == id {
			return rule, true
		}
	}
	return TacticalComputerRule{}, false
}

func (f *TacticalCombatFile) ShieldByID(id string) (TacticalShieldRule, bool) {
	for _, rule := range f.Shields {
		if rule.ShieldID == id {
			return rule, true
		}
	}
	return TacticalShieldRule{}, false
}

func (f *TacticalCombatFile) FuelCellByID(id string) (TacticalFuelCellRule, bool) {
	for _, rule := range f.FuelCells {
		if rule.FuelCellID == id {
			return rule, true
		}
	}
	return TacticalFuelCellRule{}, false
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
	if f.Ordnance.MovementUnitsPerField != 20 || f.Ordnance.FastSpeedBonus != 4 || f.Ordnance.TorpedoCooldownRounds != 2 || f.Ordnance.MIRVMultiplier != 4 || f.Ordnance.PlasmaDissipationPerField != 5 {
		return fmt.Errorf("invalid tactical ordnance constants")
	}
	if f.Weapon.ID == "" || f.Weapon.TechnologyID <= 0 || f.Weapon.Kind != "beam" || f.Weapon.BaseSpace <= 0 || f.Weapon.BaseCostPP <= 0 || f.Weapon.MinDamage <= 0 || f.Weapon.MaxDamage < f.Weapon.MinDamage {
		return fmt.Errorf("invalid tactical weapon rule")
	}
	weapons := f.AllWeapons()
	if len(f.BeamWeapons()) == 0 {
		return fmt.Errorf("at least one Beam weapon rule is required")
	}
	seenWeapons := make(map[string]struct{}, len(weapons))
	for i, weapon := range weapons {
		if weapon.ID == "" || weapon.TechnologyID <= 0 {
			return fmt.Errorf("invalid tactical weapon[%d]", i)
		}
		switch weapon.Kind {
		case "beam":
			if weapon.BaseSpace <= 0 || weapon.BaseCostPP <= 0 || weapon.MinDamage <= 0 || weapon.MaxDamage < weapon.MinDamage || weapon.CraftPerUnit != 0 || weapon.ProjectileSpeed != 0 || weapon.ProjectileDurability != 0 || weapon.BeamInterceptorDCV != 0 || (weapon.TacticalSupport != "" && weapon.TacticalSupport != "beam") {
				return fmt.Errorf("invalid Beam weapon[%d]", i)
			}
		case "torpedo":
			if weapon.BaseSpace <= 0 || weapon.BaseCostPP <= 0 || weapon.MinDamage <= 0 || weapon.MaxDamage < weapon.MinDamage || weapon.CraftPerUnit != 0 || weapon.TacticalSupport != "ordnance" || weapon.ProjectileSpeed <= 0 || weapon.ProjectileDurability <= 0 || weapon.BeamInterceptorDCV <= 0 {
				return fmt.Errorf("invalid Torpedo weapon[%d]", i)
			}
		case "bomb":
			if weapon.BaseSpace <= 0 || weapon.BaseCostPP <= 0 || weapon.MinDamage <= 0 || weapon.MaxDamage < weapon.MinDamage || weapon.CraftPerUnit != 0 || weapon.TacticalSupport != "bombardment_only" || weapon.ProjectileSpeed != 0 || weapon.ProjectileDurability != 0 || weapon.BeamInterceptorDCV != 0 {
				return fmt.Errorf("invalid Bomb weapon[%d]", i)
			}
		case "missile":
			if weapon.BaseSpace < 0 || weapon.BaseCostPP < 0 || weapon.MinDamage <= 0 || weapon.MaxDamage < weapon.MinDamage || weapon.CraftPerUnit != 0 || weapon.TacticalSupport != "ordnance" || weapon.ProjectileSpeed <= 0 || weapon.ProjectileDurability <= 0 || weapon.BeamInterceptorDCV <= 0 {
				return fmt.Errorf("invalid Missile weapon[%d]", i)
			}
		case "fighter":
			if weapon.BaseSpace <= 0 || weapon.BaseCostPP <= 0 || weapon.MinDamage != 0 || weapon.MaxDamage != 0 || weapon.CraftPerUnit <= 0 || weapon.TacticalSupport != "metadata_only" {
				return fmt.Errorf("invalid Fighter weapon[%d]", i)
			}
		default:
			return fmt.Errorf("invalid tactical weapon[%d] kind %q", i, weapon.Kind)
		}
		if _, ok := seenWeapons[weapon.ID]; ok {
			return fmt.Errorf("duplicate tactical weapon %q", weapon.ID)
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
	for i, modifier := range f.WeaponModifiers() {
		if modifier.ID == "" || (modifier.Kind != "mount" && modifier.Kind != "special") || modifier.Bit <= 0 || modifier.RequiredMiniLevel < 0 {
			return fmt.Errorf("invalid weapon modifier[%d]", i)
		}
		if _, ok := seenModifierIDs[modifier.ID]; ok {
			return fmt.Errorf("duplicate weapon modifier id %q", modifier.ID)
		}
		if _, ok := seenModifierBits[modifier.Bit]; ok {
			return fmt.Errorf("duplicate weapon modifier bit 0x%X", modifier.Bit)
		}
		seenModifierIDs[modifier.ID] = struct{}{}
		seenModifierBits[modifier.Bit] = struct{}{}
	}
	if f.Frigate.HullID == "" || f.Frigate.ArmorID == "" || f.Frigate.FuelCellID == "" || f.Frigate.ArmorHits <= 0 || f.Frigate.Structure <= 0 {
		return fmt.Errorf("invalid Frigate tactical rule")
	}
	if f.MaxUnusedSpaceCombatSpeedBonus != 10 || f.BeamDefensePerCombatSpeed != 5 || f.ShieldCapacityPerStrength != 5 {
		return fmt.Errorf("invalid tactical foundation constants")
	}
	if len(f.Hulls) != 6 {
		return fmt.Errorf("expected exactly 6 tactical hull rules, got %d", len(f.Hulls))
	}
	seenHull := make(map[string]struct{}, len(f.Hulls))
	for i, hull := range f.Hulls {
		if hull.HullID == "" || hull.BaseArmor <= 0 || hull.Structure <= 0 || hull.BaseCombatSpeed <= 0 || hull.ShieldHullFactor <= 0 || hull.MarineMax <= 0 {
			return fmt.Errorf("invalid tactical hull[%d]", i)
		}
		if _, ok := seenHull[hull.HullID]; ok {
			return fmt.Errorf("duplicate tactical hull %q", hull.HullID)
		}
		seenHull[hull.HullID] = struct{}{}
	}
	if len(f.Armors) != 6 {
		return fmt.Errorf("expected exactly 6 tactical armor rules, got %d", len(f.Armors))
	}
	seenArmor := make(map[string]struct{}, len(f.Armors))
	for i, armor := range f.Armors {
		if armor.ArmorID == "" || armor.BonusPercent < 0 {
			return fmt.Errorf("invalid tactical armor[%d]", i)
		}
		if _, ok := seenArmor[armor.ArmorID]; ok {
			return fmt.Errorf("duplicate tactical armor %q", armor.ArmorID)
		}
		seenArmor[armor.ArmorID] = struct{}{}
	}
	if len(f.Drives) != 6 {
		return fmt.Errorf("expected exactly 6 tactical drive rules, got %d", len(f.Drives))
	}
	seenDrive := make(map[string]struct{}, len(f.Drives))
	for i, drive := range f.Drives {
		if drive.DriveID == "" || drive.MinSpeed <= 0 || drive.MaxSpeed < drive.MinSpeed || drive.CombatSpeedBonus < 0 {
			return fmt.Errorf("invalid tactical drive[%d]", i)
		}
		if _, ok := seenDrive[drive.DriveID]; ok {
			return fmt.Errorf("duplicate tactical drive %q", drive.DriveID)
		}
		seenDrive[drive.DriveID] = struct{}{}
	}
	if f.Computer.ComputerID == "" || f.Computer.BeamOffense <= 0 {
		return fmt.Errorf("invalid baseline tactical computer rule")
	}
	if len(f.Computers) != 5 {
		return fmt.Errorf("expected exactly 5 tactical computer rules, got %d", len(f.Computers))
	}
	seenComputer := make(map[string]struct{}, len(f.Computers))
	for i, computer := range f.Computers {
		if computer.ComputerID == "" || computer.BeamOffense <= 0 {
			return fmt.Errorf("invalid tactical computer[%d]", i)
		}
		if _, ok := seenComputer[computer.ComputerID]; ok {
			return fmt.Errorf("duplicate tactical computer %q", computer.ComputerID)
		}
		seenComputer[computer.ComputerID] = struct{}{}
	}
	if len(f.Shields) != 5 {
		return fmt.Errorf("expected exactly 5 tactical shield rules, got %d", len(f.Shields))
	}
	seenShield := make(map[string]struct{}, len(f.Shields))
	for i, shield := range f.Shields {
		if shield.ShieldID == "" || shield.Strength <= 0 {
			return fmt.Errorf("invalid tactical shield[%d]", i)
		}
		if _, ok := seenShield[shield.ShieldID]; ok {
			return fmt.Errorf("duplicate tactical shield %q", shield.ShieldID)
		}
		seenShield[shield.ShieldID] = struct{}{}
	}
	if len(f.FuelCells) != 5 {
		return fmt.Errorf("expected exactly 5 tactical fuel-cell rules, got %d", len(f.FuelCells))
	}
	seenFuel := make(map[string]struct{}, len(f.FuelCells))
	for i, fuel := range f.FuelCells {
		if fuel.FuelCellID == "" {
			return fmt.Errorf("invalid tactical fuel cell[%d]", i)
		}
		if _, ok := seenFuel[fuel.FuelCellID]; ok {
			return fmt.Errorf("duplicate tactical fuel cell %q", fuel.FuelCellID)
		}
		seenFuel[fuel.FuelCellID] = struct{}{}
	}
	return nil
}
