package game

import (
	"testing"

	"moox/internal/battle"
	"moox/internal/core"
)

func TestSlice175Checkpoint4MixedWeaponProjectionCarriesAuthoritativeFamilies(t *testing.T) {
	rules := tacticalMetadataRules(t)
	attacker := checkpoint2Ship(101, 1, "cruiser", "ion_drive", "positronic_computer", "zortrium_armor", "class_iii_shield", "iridium_fuel_cells")
	attacker.Spec.Weapons = []core.ShipWeaponMount{
		{Slot: 0, WeaponID: "laser_cannon", Count: 1},
		{Slot: 1, WeaponID: "nuclear_missile", Count: 1, Ammo: 5, FiringArcID: core.ShipWeaponFiringArc360},
		{Slot: 2, WeaponID: "fighter_bays", Count: 1},
	}
	defender := checkpoint2Ship(201, 2, "frigate", "nuclear_drive", "electronic_computer", "titanium_armor", "", "standard_fuel_cells")
	encounter := checkpoint2Encounter(attacker, defender)
	tactical, reason, err := tacticalMetadataForEncounter(&core.GameState{Ships: []core.Ship{attacker, defender}}, encounter, rules)
	if err != nil {
		t.Fatal(err)
	}
	if reason != "" || tactical == nil {
		t.Fatalf("mixed weapon design blocked encounter: tactical=%+v reason=%q", tactical, reason)
	}
	if len(tactical.Ships) == 0 || len(tactical.Ships[0].Weapons) != 3 {
		t.Fatalf("mixed weapon snapshot missing: %+v", tactical.Ships)
	}
	wantFamilies := []string{"beam", "missile", "fighter"}
	for i, want := range wantFamilies {
		if got := tactical.Ships[0].Weapons[i].WeaponFamily; got != want {
			t.Fatalf("slot %d family=%q want=%q", i, got, want)
		}
	}

	session, err := battle.NewSession(checkpoint2BattleSpec(encounter, tactical))
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Start(); err != nil {
		t.Fatal(err)
	}
	view, err := session.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if view.Tactical == nil || len(view.Tactical.WeaponActions) != 3 {
		t.Fatalf("generic mixed-weapon action projection=%+v", view.Tactical)
	}
	for i, action := range view.Tactical.WeaponActions {
		if action.WeaponSlot != i || action.WeaponFamily != wantFamilies[i] {
			t.Fatalf("action[%d]=%+v", i, action)
		}
	}
	if !view.Tactical.WeaponActions[0].Available || view.Tactical.WeaponActions[0].CommandKind != battle.CommandFireBeam {
		t.Fatalf("Laser action not executable: %+v", view.Tactical.WeaponActions[0])
	}
	if action := view.Tactical.WeaponActions[1]; !action.Available || action.CommandKind != battle.CommandFireOrdnance || len(action.Targets) == 0 {
		t.Fatalf("Missile action not executable after Slice 17.7 CP2: %+v", action)
	}
	if action := view.Tactical.WeaponActions[2]; action.Available || action.UnavailableReason != battle.TacticalWeaponUnavailableConsumer || len(action.Targets) != 0 {
		t.Fatalf("Fighter slot not explicitly unavailable: %+v", action)
	}
}
