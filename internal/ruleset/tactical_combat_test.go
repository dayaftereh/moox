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
	if got := file.BeamWeapons(); len(got) != 13 {
		t.Fatalf("Beam weapon count=%d want=13", len(got))
	}
	if got := file.Beam.SpaceMiniaturizationFactors; !reflect.DeepEqual(got, []int{100, 80, 70, 60, 50, 40}) {
		t.Fatalf("Beam Space miniaturization=%v", got)
	}
	if got := file.Beam.CostMiniaturizationFactors; !reflect.DeepEqual(got, []int{100, 75, 55, 40, 30, 25}) {
		t.Fatalf("Beam Cost miniaturization=%v", got)
	}
	phasor, ok := file.WeaponByID("phasor")
	if !ok || phasor.TechnologyID != 127 || phasor.TechnologyFieldID != 52 || phasor.BaseSpace != 10 || phasor.BaseCostPP != 10 || phasor.MinDamage != 5 || phasor.MaxDamage != 20 || phasor.AllowedModifierMask != 0xD6 {
		t.Fatalf("phasor rule=%+v ok=%v", phasor, ok)
	}
	deathRay, ok := file.WeaponByID("death_ray")
	if !ok || deathRay.AllowedModifierMask != 0x02 {
		t.Fatalf("death ray executable modifier mask=%#x ok=%v want=0x02", deathRay.AllowedModifierMask, ok)
	}
	if len(file.BeamModifiers) != 8 {
		t.Fatalf("Beam modifier count=%d want=8", len(file.BeamModifiers))
	}
	if file.BeamModifiers[0].ID != "heavy" || file.BeamModifiers[0].Bit != 0x02 || file.BeamModifiers[0].SpacePercentDelta != 100 {
		t.Fatalf("heavy modifier=%+v", file.BeamModifiers[0])
	}
	if file.BeamModifiers[7].ID != "enveloping" || file.BeamModifiers[7].RequiredMiniLevel != 2 || file.BeamModifiers[7].Bit != 0x100 {
		t.Fatalf("enveloping modifier=%+v", file.BeamModifiers[7])
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
