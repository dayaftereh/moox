package game

import (
	"testing"

	"moox/internal/battle"
	"moox/internal/core"
)

func TestSlice176Gate4AllSupportedBeamsProjectAndExecute(t *testing.T) {
	rules := tacticalMetadataRules(t)
	weaponIDs := []string{
		"mass_driver",
		"gauss_cannon",
		"laser_cannon",
		"particle_beam",
		"fusion_beam",
		"ion_pulse_cannon",
		"graviton_beam",
		"neutron_blaster",
		"phasor",
		"disrupter_cannon",
		"death_ray",
		"plasma_cannon",
		"mauler_device",
	}

	for _, weaponID := range weaponIDs {
		t.Run(weaponID, func(t *testing.T) {
			attacker := checkpoint2Ship(
				101, 1,
				"battleship",
				"interphased_drive",
				"moleculartronic_computer",
				"xentronium_armor",
				"class_x_shield",
				"thorium_fuel_cells",
			)
			attacker.Spec.Weapons = []core.ShipWeaponMount{{
				Slot:        0,
				WeaponID:    weaponID,
				Count:       1,
				MountType:   core.ShipWeaponMountStandard,
				FiringArcID: core.ShipWeaponFiringArc360,
			}}
			defender := checkpoint2Ship(
				201, 2,
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
				t.Fatal(err)
			}
			if reason != "" || tactical == nil {
				t.Fatalf("Beam %q failed Tactical projection: tactical=%+v reason=%q", weaponID, tactical, reason)
			}
			if len(tactical.Ships) == 0 || len(tactical.Ships[0].Weapons) != 1 {
				t.Fatalf("Beam %q Tactical snapshot=%+v", weaponID, tactical.Ships)
			}
			projected := tactical.Ships[0].Weapons[0]
			if projected.WeaponID != weaponID || projected.TacticalSupport != "beam" || projected.WeaponFamily != "beam" {
				t.Fatalf("Beam %q projected identity=%+v", weaponID, projected)
			}

			session, err := battle.NewSession(checkpoint2BattleSpec(encounter, tactical))
			if err != nil {
				t.Fatal(err)
			}
			if err := session.Start(); err != nil {
				t.Fatal(err)
			}
			view, err := session.PlayerView(encounter.Attacker.SeatID)
			if err != nil {
				t.Fatal(err)
			}
			var action *battle.TacticalWeaponActionAvailability
			for i := range view.Tactical.WeaponActions {
				if view.Tactical.WeaponActions[i].WeaponSlot == 0 {
					action = &view.Tactical.WeaponActions[i]
					break
				}
			}
			if action == nil || !action.Available || action.CommandKind != battle.CommandFireBeam ||
				len(action.Targets) != 1 || action.Targets[0].TargetShipID != defender.ID {
				t.Fatalf("Beam %q legal action=%+v", weaponID, action)
			}

			fire, err := battle.NewFireBeamCommand(1, battle.FireBeamPayload{
				ShipID: attacker.ID, TargetShipID: defender.ID, WeaponSlot: 0,
			})
			if err != nil {
				t.Fatal(err)
			}
			prepared, err := session.PrepareCommand(encounter.Attacker.SeatID, fire)
			if err != nil {
				t.Fatalf("Beam %q prepare: %v", weaponID, err)
			}
			if err := session.CommitPreparedCommand(prepared); err != nil {
				t.Fatalf("Beam %q commit: %v", weaponID, err)
			}
			after := session.View()
			found := false
			for _, event := range after.Tactical.Events {
				if event.Kind == "beam_fired" && event.CommandSequence == 1 {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("Beam %q executed without authoritative beam_fired event: %+v", weaponID, after.Tactical.Events)
			}
		})
	}
}
