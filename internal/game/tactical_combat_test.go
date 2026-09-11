package game

import (
	"path/filepath"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
	"moox/internal/ruleset"
)

func tacticalMetadataRules(t *testing.T) *ruleset.TacticalCombatFile {
	t.Helper()
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	return rules.TacticalCombat
}

func tacticalMetadataShip(id, empireID core.ID, drive string, armed bool) core.Ship {
	spec := core.ShipDesignSpec{
		HullID: "frigate", WarpDriveID: drive, ComputerID: "electronic_computer",
		ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4,
	}
	if armed {
		spec.Weapons = []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}}
	}
	return core.Ship{ID: id, EmpireID: empireID, Name: "Tactical Test", Spec: spec}
}

func TestTacticalMetadataSupportsTwoByTwoWithOpenFieldDeployment(t *testing.T) {
	state := &core.GameState{Ships: []core.Ship{
		tacticalMetadataShip(101, 1, "fusion_drive", true),
		tacticalMetadataShip(102, 1, "fusion_drive", true),
		tacticalMetadataShip(201, 2, "nuclear_drive", false),
		tacticalMetadataShip(202, 2, "nuclear_drive", false),
	}}
	encounter := Encounter{
		SystemID:     7,
		Attacker:     EncounterSide{EmpireID: 1, SeatID: 1, CombatFleetIDs: []core.ID{11}, ShipIDs: []core.ID{101, 102}},
		Defender:     EncounterSide{EmpireID: 2, SeatID: 2, CombatFleetIDs: []core.ID{22}, ShipIDs: []core.ID{201, 202}},
		Participants: []protocol.SeatID{1, 2},
	}
	tactical, reason, err := tacticalMetadataForEncounter(state, encounter, tacticalMetadataRules(t))
	if err != nil {
		t.Fatal(err)
	}
	if reason != "" || tactical == nil {
		t.Fatalf("2v2 tactical metadata unsupported: tactical=%+v reason=%q", tactical, reason)
	}
	if len(tactical.Ships) != 4 {
		t.Fatalf("tactical ships=%d want 4", len(tactical.Ships))
	}
	want := map[core.ID]struct{ x, y, facing, speed int }{
		101: {10, 9, 0, 22},
		102: {10, 11, 0, 22},
		201: {14, 9, 8, 20},
		202: {14, 11, 8, 20},
	}
	occupied := map[[2]int]core.ID{}
	for _, ship := range tactical.Ships {
		expected, ok := want[ship.ShipID]
		if !ok {
			t.Fatalf("unexpected tactical ship %d", ship.ShipID)
		}
		if ship.X != expected.x || ship.Y != expected.y || ship.Facing != expected.facing || ship.CurrentCombatSpeed != expected.speed {
			t.Fatalf("ship %d deployment=(%d,%d) facing=%d speed=%d want (%d,%d) facing=%d speed=%d", ship.ShipID, ship.X, ship.Y, ship.Facing, ship.CurrentCombatSpeed, expected.x, expected.y, expected.facing, expected.speed)
		}
		if ship.TurningMode != "normal" {
			t.Fatalf("ship %d turning mode=%q want normal", ship.ShipID, ship.TurningMode)
		}
		key := [2]int{ship.X, ship.Y}
		if prior := occupied[key]; prior != 0 {
			t.Fatalf("ships %d and %d share deployment %v", prior, ship.ShipID, key)
		}
		occupied[key] = ship.ShipID
	}
}

func TestTacticalMetadataAllowsAllUnarmedBattle(t *testing.T) {
	state := &core.GameState{Ships: []core.Ship{
		tacticalMetadataShip(101, 1, "nuclear_drive", false),
		tacticalMetadataShip(201, 2, "nuclear_drive", false),
	}}
	encounter := Encounter{
		SystemID:     7,
		Attacker:     EncounterSide{EmpireID: 1, SeatID: 1, CombatFleetIDs: []core.ID{11}, ShipIDs: []core.ID{101}},
		Defender:     EncounterSide{EmpireID: 2, SeatID: 2, CombatFleetIDs: []core.ID{22}, ShipIDs: []core.ID{201}},
		Participants: []protocol.SeatID{1, 2},
	}
	tactical, reason, err := tacticalMetadataForEncounter(state, encounter, tacticalMetadataRules(t))
	if err != nil {
		t.Fatal(err)
	}
	if tactical == nil || reason != "" || len(tactical.Ships) != 2 {
		t.Fatalf("all-unarmed battle tactical=%+v reason=%q", tactical, reason)
	}
	for _, ship := range tactical.Ships {
		if len(ship.Weapons) != 0 {
			t.Fatalf("unarmed tactical ship %d unexpectedly has weapons %+v", ship.ShipID, ship.Weapons)
		}
	}
}

func TestTacticalMetadataAllowsThreeByTwoAndIgnoresCivilianAndColonyContext(t *testing.T) {
	state := &core.GameState{Ships: []core.Ship{
		tacticalMetadataShip(101, 1, "fusion_drive", true),
		tacticalMetadataShip(102, 1, "fusion_drive", false),
		tacticalMetadataShip(103, 1, "fusion_drive", false),
		tacticalMetadataShip(201, 2, "nuclear_drive", false),
		tacticalMetadataShip(202, 2, "nuclear_drive", false),
	}}
	encounter := Encounter{
		SystemID:          7,
		Attacker:          EncounterSide{EmpireID: 1, SeatID: 1, CombatFleetIDs: []core.ID{11}, CivilianFleetIDs: []core.ID{12}, ShipIDs: []core.ID{101, 102, 103}},
		Defender:          EncounterSide{EmpireID: 2, SeatID: 2, CombatFleetIDs: []core.ID{22}, CivilianFleetIDs: []core.ID{23}, ShipIDs: []core.ID{201, 202}},
		DefenderColonyIDs: []core.ID{31},
		Participants:      []protocol.SeatID{1, 2},
	}
	tactical, reason, err := tacticalMetadataForEncounter(state, encounter, tacticalMetadataRules(t))
	if err != nil {
		t.Fatal(err)
	}
	if tactical == nil || reason != "" || len(tactical.Ships) != 5 {
		t.Fatalf("3v2 battle tactical=%+v reason=%q", tactical, reason)
	}
	want := []core.ID{101, 102, 103, 201, 202}
	for i, ship := range tactical.Ships {
		if ship.ShipID != want[i] {
			t.Fatalf("tactical ship[%d]=%d want %d", i, ship.ShipID, want[i])
		}
	}
}
