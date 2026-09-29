package game

import (
	"strings"
	"testing"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/protocol"
)

func checkpoint2Encounter(attacker, defender core.Ship) Encounter {
	return Encounter{
		SystemID:     7,
		Attacker:     EncounterSide{EmpireID: attacker.EmpireID, SeatID: 1, CombatFleetIDs: []core.ID{11}, ShipIDs: []core.ID{attacker.ID}},
		Defender:     EncounterSide{EmpireID: defender.EmpireID, SeatID: 2, CombatFleetIDs: []core.ID{22}, ShipIDs: []core.ID{defender.ID}},
		Participants: []protocol.SeatID{1, 2},
	}
}

func checkpoint2BattleSpec(encounter Encounter, tactical *battle.TacticalSpec) battle.Spec {
	return battle.Spec{
		ID: 17502, GameID: "slice-17-5-cp2", StrategicTurn: 1, SystemID: encounter.SystemID,
		Attacker: battle.Side{
			EmpireID: encounter.Attacker.EmpireID, SeatID: encounter.Attacker.SeatID,
			CombatFleetIDs: append([]core.ID(nil), encounter.Attacker.CombatFleetIDs...),
			ShipIDs:        append([]core.ID(nil), encounter.Attacker.ShipIDs...),
		},
		Defender: battle.Side{
			EmpireID: encounter.Defender.EmpireID, SeatID: encounter.Defender.SeatID,
			CombatFleetIDs: append([]core.ID(nil), encounter.Defender.CombatFleetIDs...),
			ShipIDs:        append([]core.ID(nil), encounter.Defender.ShipIDs...),
		},
		Participants: append([]protocol.SeatID(nil), encounter.Participants...),
		Seed:         17502, Tactical: tactical,
	}
}

func checkpoint2Ship(id, empireID core.ID, hullID, driveID, computerID, armorID, shieldID, fuelID string) core.Ship {
	ship := tacticalMetadataShip(id, empireID, driveID, false)
	ship.Spec.HullID = hullID
	ship.Spec.ComputerID = computerID
	ship.Spec.ArmorID = armorID
	ship.Spec.ShieldID = shieldID
	ship.Spec.FuelCellID = fuelID
	ship.Spec.HullSpace = 100
	ship.Spec.SpaceUsed = 100
	return ship
}

func TestSlice175Checkpoint2BroadCombatantAdmissionMatrix(t *testing.T) {
	rules := tacticalMetadataRules(t)
	hulls := []string{"frigate", "destroyer", "cruiser", "battleship", "titan", "doom_star"}
	tiers := []struct {
		name, drive, computer, armor, shield, fuel string
	}{
		{name: "early", drive: "nuclear_drive", computer: "electronic_computer", armor: "titanium_armor", fuel: "standard_fuel_cells"},
		{name: "mid", drive: "ion_drive", computer: "positronic_computer", armor: "zortrium_armor", shield: "class_iii_shield", fuel: "iridium_fuel_cells"},
		{name: "late", drive: "interphased_drive", computer: "moleculartronic_computer", armor: "xentronium_armor", shield: "class_x_shield", fuel: "thorium_fuel_cells"},
	}
	defender := checkpoint2Ship(201, 2, "frigate", "nuclear_drive", "electronic_computer", "titanium_armor", "", "standard_fuel_cells")

	for _, tier := range tiers {
		for _, hull := range hulls {
			t.Run(tier.name+"/"+hull, func(t *testing.T) {
				attacker := checkpoint2Ship(101, 1, hull, tier.drive, tier.computer, tier.armor, tier.shield, tier.fuel)
				encounter := checkpoint2Encounter(attacker, defender)
				tactical, reason, err := tacticalMetadataForEncounter(&core.GameState{Ships: []core.Ship{attacker, defender}}, encounter, rules)
				if err != nil {
					t.Fatal(err)
				}
				if reason != "" || tactical == nil {
					t.Fatalf("legal combatant was not admitted: tactical=%+v reason=%q", tactical, reason)
				}
				if _, err := battle.NewSession(checkpoint2BattleSpec(encounter, tactical)); err != nil {
					t.Fatalf("BattleSession rejected legal %s/%s combatant: %v", tier.name, hull, err)
				}
			})
		}
	}
}

func TestSlice175Checkpoint2UnknownMandatoryIdentitiesAreEncounterErrors(t *testing.T) {
	rules := tacticalMetadataRules(t)
	tests := []struct {
		name  string
		apply func(*core.Ship)
		want  string
	}{
		{name: "hull", apply: func(s *core.Ship) { s.Spec.HullID = "unknown_hull" }, want: "hull"},
		{name: "drive", apply: func(s *core.Ship) { s.Spec.WarpDriveID = "unknown_drive" }, want: "drive"},
		{name: "computer", apply: func(s *core.Ship) { s.Spec.ComputerID = "unknown_computer" }, want: "computer"},
		{name: "armor", apply: func(s *core.Ship) { s.Spec.ArmorID = "unknown_armor" }, want: "armor"},
		{name: "shield", apply: func(s *core.Ship) { s.Spec.ShieldID = "unknown_shield" }, want: "shield"},
		{name: "fuel", apply: func(s *core.Ship) { s.Spec.FuelCellID = "unknown_fuel" }, want: "fuel"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			attacker := checkpoint2Ship(101, 1, "frigate", "nuclear_drive", "electronic_computer", "titanium_armor", "", "standard_fuel_cells")
			defender := checkpoint2Ship(201, 2, "frigate", "nuclear_drive", "electronic_computer", "titanium_armor", "", "standard_fuel_cells")
			tc.apply(&attacker)
			tactical, reason, err := tacticalMetadataForEncounter(&core.GameState{Ships: []core.Ship{attacker, defender}}, checkpoint2Encounter(attacker, defender), rules)
			if err == nil || tactical != nil || reason != "" {
				t.Fatalf("unknown %s did not fail structurally: tactical=%+v reason=%q err=%v", tc.name, tactical, reason, err)
			}
			if !strings.Contains(strings.ToLower(err.Error()), tc.want) {
				t.Fatalf("unknown %s error=%v", tc.name, err)
			}
		})
	}
}

func TestSlice175Checkpoint2MalformedWeaponAndSpecialSnapshotsAreEncounterErrors(t *testing.T) {
	rules := tacticalMetadataRules(t)
	tests := []struct {
		name  string
		apply func(*core.Ship)
	}{
		{name: "duplicate weapon slot", apply: func(s *core.Ship) {
			s.Spec.Weapons = []core.ShipWeaponMount{
				{Slot: 0, WeaponID: "laser_cannon", Count: 1},
				{Slot: 0, WeaponID: "laser_cannon", Count: 1},
			}
		}},
		{name: "weapon slot out of range", apply: func(s *core.Ship) {
			s.Spec.Weapons = []core.ShipWeaponMount{{Slot: 8, WeaponID: "laser_cannon", Count: 1}}
		}},
		{name: "unknown weapon", apply: func(s *core.Ship) {
			s.Spec.Weapons = []core.ShipWeaponMount{{Slot: 0, WeaponID: "unknown_weapon", Count: 1}}
		}},
		{name: "duplicate special", apply: func(s *core.Ship) {
			s.Spec.Specials = []core.ShipSpecialMount{{Slot: 0, SpecialID: "battle_pods"}, {Slot: 1, SpecialID: "battle_pods"}}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			attacker := checkpoint2Ship(101, 1, "frigate", "nuclear_drive", "electronic_computer", "titanium_armor", "", "standard_fuel_cells")
			defender := checkpoint2Ship(201, 2, "frigate", "nuclear_drive", "electronic_computer", "titanium_armor", "", "standard_fuel_cells")
			tc.apply(&attacker)
			tactical, reason, err := tacticalMetadataForEncounter(&core.GameState{Ships: []core.Ship{attacker, defender}}, checkpoint2Encounter(attacker, defender), rules)
			if err == nil || tactical != nil || reason != "" {
				t.Fatalf("malformed combatant was not rejected: tactical=%+v reason=%q err=%v", tactical, reason, err)
			}
		})
	}
}

func TestSlice175Checkpoint2LaterFamilyWeaponsDoNotBlockAdmission(t *testing.T) {
	rules := tacticalMetadataRules(t)
	for _, weaponID := range []string{"nuclear_missile", "fighter_bays"} {
		t.Run(weaponID, func(t *testing.T) {
			attacker := checkpoint2Ship(101, 1, "cruiser", "ion_drive", "positronic_computer", "zortrium_armor", "class_iii_shield", "iridium_fuel_cells")
			attacker.Spec.Weapons = []core.ShipWeaponMount{{Slot: 0, WeaponID: weaponID, Count: 1}}
			if weaponID == "nuclear_missile" {
				attacker.Spec.Weapons[0].Ammo = 5
				attacker.Spec.Weapons[0].FiringArcID = core.ShipWeaponFiringArc360
			}
			defender := checkpoint2Ship(201, 2, "frigate", "nuclear_drive", "electronic_computer", "titanium_armor", "", "standard_fuel_cells")
			encounter := checkpoint2Encounter(attacker, defender)
			tactical, reason, err := tacticalMetadataForEncounter(&core.GameState{Ships: []core.Ship{attacker, defender}}, encounter, rules)
			if err != nil {
				t.Fatal(err)
			}
			if reason != "" || tactical == nil {
				t.Fatalf("later-family weapon blocked encounter: tactical=%+v reason=%q", tactical, reason)
			}
			session, err := battle.NewSession(checkpoint2BattleSpec(encounter, tactical))
			if err != nil {
				t.Fatalf("BattleSession rejected later-family metadata: %v", err)
			}
			if err := session.Start(); err != nil {
				t.Fatalf("BattleSession start rejected later-family metadata: %v", err)
			}
			view := session.View()
			for _, action := range view.Tactical.LegalFireActions {
				if action.WeaponSlot == 0 && action.WeaponID == weaponID {
					t.Fatalf("later-family weapon unexpectedly received fire action: %+v", action)
				}
			}
		})
	}
}
