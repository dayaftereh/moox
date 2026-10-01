package game

import (
	"testing"

	"moox/internal/battle"
	"moox/internal/core"
)

func TestSlice176AllTechBeamsExecuteWhileDeferredOrdnanceRemainsUnavailable(t *testing.T) {
	rules := tacticalMetadataRules(t)

	attacker := checkpoint2Ship(
		38, 1,
		"battleship",
		"interphased_drive",
		"cybertronic_computer",
		"xentronium_armor",
		"class_v_shield",
		"thorium_fuel_cells",
	)
	attacker.Name = "Mm"
	attacker.Spec.HullSpace = 312
	attacker.Spec.SpaceUsed = 259
	attacker.Spec.Specials = []core.ShipSpecialMount{
		{Slot: 0, SpecialID: "battle_pods"},
	}
	attacker.Spec.Weapons = []core.ShipWeaponMount{
		{
			Slot:        0,
			WeaponID:    "mauler_device",
			Count:       2,
			MountType:   core.ShipWeaponMountHeavy,
			FiringArcID: core.ShipWeaponFiringArc360,
		},
		{
			Slot:        1,
			WeaponID:    "pulson_missile",
			Count:       3,
			MountType:   core.ShipWeaponMountStandard,
			ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierFast},
			Ammo:        10,
			FiringArcID: core.ShipWeaponFiringArc360,
		},
		{
			Slot:        2,
			WeaponID:    "fusion_beam",
			Count:       3,
			MountType:   core.ShipWeaponMountHeavy,
			ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierEnveloping},
			FiringArcID: core.ShipWeaponFiringArc360,
		},
	}
	defender := checkpoint2Ship(
		26, 2,
		"frigate",
		"nuclear_drive",
		"electronic_computer",
		"titanium_armor",
		"",
		"standard_fuel_cells",
	)

	encounter := checkpoint2Encounter(attacker, defender)
	tactical, reason, err := tacticalMetadataForEncounter(
		&core.GameState{Ships: []core.Ship{attacker, defender}},
		encounter,
		rules,
	)
	if err != nil {
		t.Fatalf("current All-Tech loadout failed game projection: %v", err)
	}
	if reason != "" || tactical == nil {
		t.Fatalf("current All-Tech loadout blocked Tactical: tactical=%+v reason=%q", tactical, reason)
	}

	session, err := battle.NewSession(checkpoint2BattleSpec(encounter, tactical))
	if err != nil {
		t.Fatalf("current All-Tech loadout failed BattleSession admission: %v", err)
	}
	if err := session.Start(); err != nil {
		t.Fatal(err)
	}
	view, err := session.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if view.Tactical == nil || len(view.Tactical.WeaponActions) != 3 {
		t.Fatalf("weapon action projection=%+v", view.Tactical)
	}
	for _, action := range view.Tactical.WeaponActions {
		switch action.WeaponSlot {
		case 0, 2:
			if !action.Available || action.CommandKind != battle.CommandFireBeam || len(action.Targets) == 0 {
				t.Fatalf("Slice 17.6 Beam slot must be executable: %+v", action)
			}
		case 1:
			if action.Available || action.UnavailableReason != battle.TacticalWeaponUnavailableConsumer || len(action.Targets) != 0 {
				t.Fatalf("deferred Ordnance slot must remain visible but unavailable: %+v", action)
			}
		default:
			t.Fatalf("unexpected weapon action: %+v", action)
		}
	}
}
