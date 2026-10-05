package game

import (
	"path/filepath"
	"testing"

	"moox/internal/core"
)

func TestSlice178Gate3Checkpoint2ProjectsPassiveFoundationAndInternalMaxima(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	baseShip := tacticalMetadataShip(101, 1, "fusion_drive", true)
	baseShip.Spec.ShieldID = "class_i_shield"
	base, err := tacticalFoundationForShip(baseShip, rules.TacticalCombat)
	if err != nil {
		t.Fatal(err)
	}

	ship := baseShip
	ship.Spec.Specials = []core.ShipSpecialMount{
		{Slot: 0, SpecialID: "reinforced_hull"},
		{Slot: 1, SpecialID: "heavy_armor"},
		{Slot: 2, SpecialID: "augmented_engines"},
		{Slot: 3, SpecialID: "battle_scanner"},
		{Slot: 4, SpecialID: "hard_shields"},
		{Slot: 5, SpecialID: "multi_phased_shields"},
	}
	got, err := tacticalFoundationForShip(ship, rules.TacticalCombat)
	if err != nil {
		t.Fatal(err)
	}
	if got.StructureMax != base.StructureMax*3 {
		t.Fatalf("Reinforced Hull structure=%d want=%d", got.StructureMax, base.StructureMax*3)
	}
	if got.ArmorMax != base.ArmorMax*3 {
		t.Fatalf("Heavy Armor armor=%d want=%d", got.ArmorMax, base.ArmorMax*3)
	}
	if got.CurrentCombatSpeed != base.CurrentCombatSpeed+5 {
		t.Fatalf("Augmented Engines speed=%d want=%d", got.CurrentCombatSpeed, base.CurrentCombatSpeed+5)
	}
	if got.BeamOffense != base.BeamOffense+50 {
		t.Fatalf("Battle Scanner offense=%d want=%d", got.BeamOffense, base.BeamOffense+50)
	}
	if got.BeamDefense != got.CurrentCombatSpeed*rules.TacticalCombat.BeamDefensePerCombatSpeed {
		t.Fatalf("derived Beam Defense=%d", got.BeamDefense)
	}
	if got.ShieldArcMax != (base.ShieldArcMax*3)/2 {
		t.Fatalf("Multi-Phased shield max=%d want=%d", got.ShieldArcMax, (base.ShieldArcMax*3)/2)
	}

	internal, err := tacticalInternalSpecForShip(ship, rules.TacticalCombat, rules.ShipSpecials)
	if err != nil {
		t.Fatal(err)
	}
	if internal == nil {
		t.Fatal("internal snapshot missing")
	}
	if internal.EngineMax != 6 || internal.SecondaryMax != 1 || internal.ShieldGeneratorMax != 1 {
		t.Fatalf("fixed internal maxima=%d/%d/%d want=6/1/1", internal.EngineMax, internal.SecondaryMax, internal.ShieldGeneratorMax)
	}
	if len(internal.Weapons) != 1 || internal.Weapons[0].Slot != 0 || internal.Weapons[0].Max != 1 || internal.Weapons[0].UnitSize != 1 {
		t.Fatalf("weapon internal pools=%+v", internal.Weapons)
	}
	wantSpecialSlots := []int{2, 3, 4, 5}
	if len(internal.Specials) != len(wantSpecialSlots) {
		t.Fatalf("special internal pools=%+v", internal.Specials)
	}
	for i, slot := range wantSpecialSlots {
		if internal.Specials[i].Slot != slot || internal.Specials[i].Max != 1 {
			t.Fatalf("special internal pool[%d]=%+v want slot=%d max=1", i, internal.Specials[i], slot)
		}
	}

	specials := tacticalSpecialRulesSnapshot(rules.TacticalCombat)
	if specials == nil ||
		specials.ECMJammerDefenseBonus != 70 ||
		specials.MultiWaveECMJammerDefenseBonus != 100 ||
		specials.WideAreaJammerSelfDefenseBonus != 130 ||
		specials.WideAreaJammerFleetDefenseBonus != 70 {
		t.Fatalf("ECM replay constants=%+v", specials)
	}
}
