package ruleset

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestTacticalCombatRulesLoadCanonicalBaseline(t *testing.T) {
	file, err := LoadTacticalCombat(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31", "tactical_combat.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Validate(); err != nil {
		t.Fatal(err)
	}
	if file.SchemaVersion != TacticalCombatSchemaVersion || file.Initiative.BeamOffenseDivisor != 10 {
		t.Fatalf("schema/initiative=%d/%d", file.SchemaVersion, file.Initiative.BeamOffenseDivisor)
	}
	if file.RNG.Multiplier != 0x41C64E6D || file.RNG.Increment != 0x3039 {
		t.Fatalf("rng=%08X/%08X", file.RNG.Multiplier, file.RNG.Increment)
	}
	if !reflect.DeepEqual(file.Beam.ToHitRangeModifiers, []int{0, 0, -10, -20, -30, -40, -55, -70, -85}) {
		t.Fatalf("to-hit table=%v", file.Beam.ToHitRangeModifiers)
	}
	if !reflect.DeepEqual(file.Beam.DamageRangeModifiers, []int{0, 0, -10, -20, -30, -40, -50, -60, -65}) {
		t.Fatalf("damage table=%v", file.Beam.DamageRangeModifiers)
	}
	if file.Weapon.ID != "laser_cannon" || file.Weapon.TechnologyID != 100 || file.Weapon.BaseSpace != 10 || file.Weapon.BaseCostPP != 5 || file.Weapon.MinDamage != 1 || file.Weapon.MaxDamage != 4 {
		t.Fatalf("laser rule=%+v", file.Weapon)
	}
	if file.Frigate.HullID != "frigate" || file.Frigate.ArmorHits != 4 || file.Frigate.Structure != 4 {
		t.Fatalf("frigate rule=%+v", file.Frigate)
	}
}

func TestTacticalCombatRulesRejectMalformedTables(t *testing.T) {
	file := &TacticalCombatFile{
		SchemaVersion: TacticalCombatSchemaVersion,
		Ruleset:       "test",
		Initiative:    TacticalInitiativeRules{BeamOffenseDivisor: 10},
		RNG:           TacticalRNGRules{Multiplier: 1, Increment: 1},
		Beam: TacticalBeamRules{
			BaseHitThreshold: 40, MaxHitThreshold: 95, EffectiveRollCap: 100,
			ToHitRangeModifiers: []int{0}, DamageRangeModifiers: []int{0},
		},
		Weapon:   TacticalWeaponRule{ID: "laser_cannon", TechnologyID: 100, Kind: "beam", BaseSpace: 10, BaseCostPP: 5, MinDamage: 1, MaxDamage: 4},
		Frigate:  TacticalFrigateRule{HullID: "frigate", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", ArmorHits: 4, Structure: 4},
		Drives:   []TacticalDriveSpeedRule{{DriveID: "nuclear_drive", MinSpeed: 10, MaxSpeed: 20}, {DriveID: "fusion_drive", MinSpeed: 12, MaxSpeed: 22}},
		Computer: TacticalComputerRule{ComputerID: "electronic_computer", BeamOffense: 25},
	}
	if err := file.Validate(); err == nil {
		t.Fatal("expected malformed Beam tables to reject")
	}
}
