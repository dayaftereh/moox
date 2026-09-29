package game

import (
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func TestSlice175Checkpoint1TacticalFoundationMatrix(t *testing.T) {
	rules := tacticalMetadataRules(t)

	type hullVector struct {
		id        string
		structure int
		armor     [3]int
		shieldArc [3]int
		minSpeed  [3]int
	}
	hulls := []hullVector{
		{id: "frigate", structure: 4, armor: [3]int{4, 16, 40}, shieldArc: [3]int{0, 15, 50}, minSpeed: [3]int{10, 14, 20}},
		{id: "destroyer", structure: 10, armor: [3]int{10, 40, 100}, shieldArc: [3]int{0, 30, 100}, minSpeed: [3]int{8, 12, 18}},
		{id: "cruiser", structure: 30, armor: [3]int{30, 120, 300}, shieldArc: [3]int{0, 45, 150}, minSpeed: [3]int{6, 10, 16}},
		{id: "battleship", structure: 50, armor: [3]int{50, 200, 500}, shieldArc: [3]int{0, 60, 200}, minSpeed: [3]int{5, 9, 15}},
		{id: "titan", structure: 80, armor: [3]int{80, 320, 800}, shieldArc: [3]int{0, 75, 250}, minSpeed: [3]int{4, 8, 14}},
		{id: "doom_star", structure: 150, armor: [3]int{150, 600, 1500}, shieldArc: [3]int{0, 105, 350}, minSpeed: [3]int{3, 7, 13}},
	}
	tiers := []struct {
		name           string
		driveID        string
		computerID     string
		armorID        string
		shieldID       string
		fuelCellID     string
		beamOffense    int
		shieldStrength int
	}{
		{name: "early", driveID: "nuclear_drive", computerID: "electronic_computer", armorID: "titanium_armor", fuelCellID: "standard_fuel_cells", beamOffense: 25},
		{name: "mid", driveID: "ion_drive", computerID: "positronic_computer", armorID: "zortrium_armor", shieldID: "class_iii_shield", fuelCellID: "iridium_fuel_cells", beamOffense: 75, shieldStrength: 3},
		{name: "late", driveID: "interphased_drive", computerID: "moleculartronic_computer", armorID: "xentronium_armor", shieldID: "class_x_shield", fuelCellID: "thorium_fuel_cells", beamOffense: 125, shieldStrength: 10},
	}

	for tierIndex, tier := range tiers {
		for _, hull := range hulls {
			t.Run(tier.name+"/"+hull.id, func(t *testing.T) {
				ship := core.Ship{
					ID: 77, EmpireID: 1,
					Spec: core.ShipDesignSpec{
						HullID: hull.id, WarpDriveID: tier.driveID, ComputerID: tier.computerID,
						ArmorID: tier.armorID, ShieldID: tier.shieldID, FuelCellID: tier.fuelCellID,
						HullSpace: 100, SpaceUsed: 100,
					},
				}
				full, err := projectTacticalShip(ship, protocol.SeatID(1), 10, 10, 0, rules, nil)
				if err != nil {
					t.Fatal(err)
				}
				if full.MinimumCombatSpeed != hull.minSpeed[tierIndex] || full.CurrentCombatSpeed != hull.minSpeed[tierIndex] {
					t.Fatalf("full design speed=%d/%d want=%d", full.MinimumCombatSpeed, full.CurrentCombatSpeed, hull.minSpeed[tierIndex])
				}
				if full.BeamOffense != tier.beamOffense || full.BeamDefense != hull.minSpeed[tierIndex]*5 {
					t.Fatalf("full design offense/defense=%d/%d want=%d/%d", full.BeamOffense, full.BeamDefense, tier.beamOffense, hull.minSpeed[tierIndex]*5)
				}
				if full.ArmorMax != hull.armor[tierIndex] || full.StructureMax != hull.structure {
					t.Fatalf("durability armor/structure=%d/%d want=%d/%d", full.ArmorMax, full.StructureMax, hull.armor[tierIndex], hull.structure)
				}
				if full.ShieldStrength != tier.shieldStrength || full.ShieldArcMax != hull.shieldArc[tierIndex] {
					t.Fatalf("shield strength/arc=%d/%d want=%d/%d", full.ShieldStrength, full.ShieldArcMax, tier.shieldStrength, hull.shieldArc[tierIndex])
				}
				if full.HullID != hull.id || full.WarpDriveID != tier.driveID || full.ComputerID != tier.computerID ||
					full.ArmorID != tier.armorID || full.ShieldID != tier.shieldID || full.FuelCellID != tier.fuelCellID {
					t.Fatalf("projection provenance=%+v", full)
				}

				ship.Spec.SpaceUsed = 0
				empty, err := projectTacticalShip(ship, protocol.SeatID(1), 10, 10, 0, rules, nil)
				if err != nil {
					t.Fatal(err)
				}
				wantMax := hull.minSpeed[tierIndex] + 10
				if empty.MinimumCombatSpeed != hull.minSpeed[tierIndex] || empty.CurrentCombatSpeed != wantMax || empty.BeamDefense != wantMax*5 {
					t.Fatalf("empty design speed/defense=%d/%d/%d want=%d/%d/%d", empty.MinimumCombatSpeed, empty.CurrentCombatSpeed, empty.BeamDefense, hull.minSpeed[tierIndex], wantMax, wantMax*5)
				}
			})
		}
	}
}

func TestSlice175Checkpoint1UnusedSpaceSpeedUsesCeiling(t *testing.T) {
	rules := tacticalMetadataRules(t)
	ship := core.Ship{Spec: core.ShipDesignSpec{
		HullID: "cruiser", WarpDriveID: "ion_drive", ComputerID: "positronic_computer",
		ArmorID: "zortrium_armor", ShieldID: "class_iii_shield", FuelCellID: "iridium_fuel_cells",
		HullSpace: 100, SpaceUsed: 55,
	}}
	got, err := tacticalFoundationForShip(ship, rules)
	if err != nil {
		t.Fatal(err)
	}
	if got.MinimumCombatSpeed != 10 || got.CurrentCombatSpeed != 15 || got.BeamDefense != 75 {
		t.Fatalf("partial design speed/defense=%d/%d/%d want=10/15/75", got.MinimumCombatSpeed, got.CurrentCombatSpeed, got.BeamDefense)
	}
}

func TestSlice175Checkpoint1BaselineEncounterProjectsTruthfulFoundation(t *testing.T) {
	state := &core.GameState{Ships: []core.Ship{
		tacticalMetadataShip(101, 1, "nuclear_drive", false),
		tacticalMetadataShip(201, 2, "fusion_drive", false),
	}}
	encounter := Encounter{
		SystemID:     7,
		Attacker:     EncounterSide{EmpireID: 1, SeatID: 1, ShipIDs: []core.ID{101}},
		Defender:     EncounterSide{EmpireID: 2, SeatID: 2, ShipIDs: []core.ID{201}},
		Participants: []protocol.SeatID{1, 2},
	}
	tactical, reason, err := tacticalMetadataForEncounter(state, encounter, tacticalMetadataRules(t))
	if err != nil {
		t.Fatal(err)
	}
	if reason != "" || tactical == nil || len(tactical.Ships) != 2 {
		t.Fatalf("tactical projection=%+v reason=%q", tactical, reason)
	}
	for _, ship := range tactical.Ships {
		wantMin, wantCurrent := 10, 20
		if ship.WarpDriveID == "fusion_drive" {
			wantMin, wantCurrent = 12, 22
		}
		if ship.MinimumCombatSpeed != wantMin || ship.CurrentCombatSpeed != wantCurrent || ship.BeamDefense != wantCurrent*5 {
			t.Fatalf("ship %d foundation min/current/defense=%d/%d/%d want=%d/%d/%d", ship.ShipID, ship.MinimumCombatSpeed, ship.CurrentCombatSpeed, ship.BeamDefense, wantMin, wantCurrent, wantCurrent*5)
		}
		if ship.FuelCellID != "standard_fuel_cells" || ship.ShieldID != "" || ship.ShieldStrength != 0 || ship.ShieldArcMax != 0 {
			t.Fatalf("ship %d shield/fuel provenance=%+v", ship.ShipID, ship)
		}
	}
}
