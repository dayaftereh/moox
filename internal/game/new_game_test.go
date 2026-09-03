package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func canonicalNewGameSettings() NewGameSettings {
	return NewGameSettings{
		GalaxySize:      GalaxySizeSmall,
		GalaxyAge:       GalaxyAgeNormal,
		TechnologyLevel: NewGameTechnologyAverage,
		Players: []NewGamePlayerSpec{
			{SeatID: protocol.SeatID(1), EmpireName: "Human", RaceID: "human"},
			{SeatID: protocol.SeatID(2), EmpireName: "Darlok", RaceID: "darlok"},
		},
	}
}

func TestNewGameGoldenSeedStateFingerprint(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	result, err := rules.NewGame(0x8009, canonicalNewGameSettings())
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(result.State)
	if err != nil {
		t.Fatal(err)
	}
	got := fmt.Sprintf("%x", sha256.Sum256(data))
	const want = "83614740b216409877b03c536a16328c7fa7031867f58dbeb70857a8953f5762"
	if got != want {
		t.Fatalf("golden seed state sha256=%s want=%s", got, want)
	}
}
func TestNewGameDeterministicAndRoundTrips(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	settings := canonicalNewGameSettings()
	first, err := rules.NewGame(0x8009, settings)
	if err != nil {
		t.Fatal(err)
	}
	second, err := rules.NewGame(0x8009, settings)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same seed/settings did not produce an identical NewGameResult")
	}

	data, err := json.Marshal(first.State)
	if err != nil {
		t.Fatal(err)
	}
	secondData, err := json.Marshal(second.State)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, secondData) {
		t.Fatal("same seed/settings did not produce byte-identical JSON state")
	}
	var loaded core.GameState
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}
	if err := loaded.Validate(); err != nil {
		t.Fatalf("round-tripped state invalid: %v", err)
	}
	if !reflect.DeepEqual(first.State, &loaded) {
		t.Fatal("state changed across JSON round trip")
	}

	different, err := rules.NewGame(0x800A, settings)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(first.State.Galaxy, different.State.Galaxy) || first.State.RNGState == different.State.RNGState {
		t.Fatal("different seed did not change generated galaxy/RNG state")
	}
}

func TestNewGameCanonicalBaselineInvariants(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	result, err := rules.NewGame(0x8009, canonicalNewGameSettings())
	if err != nil {
		t.Fatal(err)
	}
	state := result.State
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	if state.Galaxy.ID != 1 {
		t.Fatalf("galaxy id=%d want=1", state.Galaxy.ID)
	}
	if len(state.Empires) != 2 || state.Empires[0].ID != 2 || state.Empires[1].ID != 3 {
		t.Fatalf("empire IDs=%v", []core.ID{state.Empires[0].ID, state.Empires[1].ID})
	}
	if len(state.Galaxy.Systems) != 20 {
		t.Fatalf("systems=%d want=20", len(state.Galaxy.Systems))
	}
	coords := map[[2]int]struct{}{}
	for i, system := range state.Galaxy.Systems {
		if want := core.ID(4 + i); system.ID != want {
			t.Fatalf("system[%d] id=%d want=%d", i, system.ID, want)
		}
		if system.X < 60 || system.X > 940 || system.Y < 60 || system.Y > 740 {
			t.Fatalf("system %d coords=(%d,%d) outside envelope", system.ID, system.X, system.Y)
		}
		key := [2]int{system.X, system.Y}
		if _, exists := coords[key]; exists {
			t.Fatalf("duplicate coordinates %v", key)
		}
		coords[key] = struct{}{}
		lastOrbit := -1
		for _, planet := range system.Planets {
			if planet.Orbit <= lastOrbit || planet.Orbit < 0 || planet.Orbit > 4 {
				t.Fatalf("system %d planet orbits not sorted/unique", system.ID)
			}
			lastOrbit = planet.Orbit
		}
	}

	homeSystems := make([]*core.StarSystem, 2)
	for i := range state.Empires {
		empire := &state.Empires[i]
		if empire.Treasury.BalanceBC != 50 || empire.Freighters != 0 {
			t.Fatalf("empire %d start treasury/freighters=%v/%d", empire.ID, empire.Treasury.BalanceBC, empire.Freighters)
		}
		wantTech := []int{32, 40, 41, 58, 63, 69, 100, 101, 103, 109, 119, 120, 121, 145, 157, 166, 167, 168, 187, 189}
		if !reflect.DeepEqual(empire.KnownTechnologyIDs, wantTech) {
			t.Fatalf("empire %d tech=%v want=%v", empire.ID, empire.KnownTechnologyIDs, wantTech)
		}
		colony := colonyByID(state, empire.Capital)
		if colony == nil {
			t.Fatalf("empire %d has no capital colony", empire.ID)
		}
		if colony.Population.Total() != 8 || colony.Population.Farmers() != 4 || colony.Population.Workers() != 2 || colony.Population.Scientists() != 2 {
			t.Fatalf("empire %d population=%+v", empire.ID, colony.Population)
		}
		if len(colony.Buildings) != 0 {
			t.Fatalf("empire %d starts with buildings %v", empire.ID, colony.Buildings)
		}
		planet := planetByID(state, colony.PlanetID)
		if planet == nil {
			t.Fatalf("capital planet missing")
		}
		if planet.SizeID != "medium" || planet.MineralID != "abundant" || planet.GravityID != "normal_g" || planet.ClimateID != "terran" {
			t.Fatalf("homeworld=%+v", *planet)
		}
		for si := range state.Galaxy.Systems {
			for pi := range state.Galaxy.Systems[si].Planets {
				if state.Galaxy.Systems[si].Planets[pi].ID == planet.ID {
					homeSystems[i] = &state.Galaxy.Systems[si]
				}
			}
		}
		if homeSystems[i] == nil || len(homeSystems[i].Planets) < 3 {
			t.Fatalf("empire %d home system has <3 planets", empire.ID)
		}
		if planet.Orbit != homeSystems[i].Planets[0].Orbit {
			t.Fatalf("homeworld is not lowest orbit")
		}
	}
	if homeSystems[0].ID >= homeSystems[1].ID {
		t.Fatalf("player home IDs not ascending: %d,%d", homeSystems[0].ID, homeSystems[1].ID)
	}
	maxDistance := int64(-1)
	var best [2]core.ID
	for i := 0; i < len(state.Galaxy.Systems); i++ {
		for j := i + 1; j < len(state.Galaxy.Systems); j++ {
			a, b := state.Galaxy.Systems[i], state.Galaxy.Systems[j]
			dx, dy := int64(a.X-b.X), int64(a.Y-b.Y)
			d := dx*dx + dy*dy
			pair := [2]core.ID{a.ID, b.ID}
			if d > maxDistance || (d == maxDistance && (pair[0] < best[0] || (pair[0] == best[0] && pair[1] < best[1]))) {
				maxDistance = d
				best = pair
			}
		}
	}
	if got := [2]core.ID{homeSystems[0].ID, homeSystems[1].ID}; got != best {
		t.Fatalf("home pair=%v want farthest=%v", got, best)
	}

	if len(state.ShipDesigns) != 2 || len(state.Ships) != 4 || len(state.StrategicFleets) != 4 {
		t.Fatalf("designs/ships/fleets=%d/%d/%d", len(state.ShipDesigns), len(state.Ships), len(state.StrategicFleets))
	}
	for i, empire := range state.Empires {
		design := state.ShipDesigns[i]
		if design.EmpireID != empire.ID || design.Name != "Scout" || design.Spec.HullID != "frigate" || design.Spec.WarpDriveID != "nuclear_drive" || design.Spec.ComputerID != "electronic_computer" || design.Spec.ArmorID != "titanium_armor" || design.Spec.FuelCellID != "standard_fuel_cells" || design.Spec.ShieldID != "" || len(design.Spec.Weapons) != 0 {
			t.Fatalf("empire %d scout design=%+v", empire.ID, design)
		}
		combat := 0
		colonyShips := 0
		for _, fleet := range state.StrategicFleets {
			if fleet.EmpireID != empire.ID {
				continue
			}
			if fleet.Role == core.StrategicFleetRoleCombat {
				combat++
				if len(fleet.ShipIDs) != 2 {
					t.Fatalf("combat fleet ships=%v", fleet.ShipIDs)
				}
			}
			if fleet.SpecialKind == core.StrategicFleetSpecialColonyShip {
				colonyShips++
			}
		}
		if combat != 1 || colonyShips != 1 {
			t.Fatalf("empire %d combat/colony fleets=%d/%d", empire.ID, combat, colonyShips)
		}
	}
	ids := append([]core.ID(nil), state.ShipDesigns[0].ID, state.ShipDesigns[1].ID)
	if !sort.SliceIsSorted(ids, func(i, j int) bool { return ids[i] < ids[j] }) {
		t.Fatal("design IDs not ordered")
	}
}

func TestNewGameRejectsUnsupportedSettings(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	cases := []NewGameSettings{
		{},
		func() NewGameSettings { s := canonicalNewGameSettings(); s.GalaxySize = "medium"; return s }(),
		func() NewGameSettings { s := canonicalNewGameSettings(); s.GalaxyAge = "organic_rich"; return s }(),
		func() NewGameSettings {
			s := canonicalNewGameSettings()
			s.TechnologyLevel = NewGameTechnologyPreWarp
			return s
		}(),
		func() NewGameSettings { s := canonicalNewGameSettings(); s.StrategicCombat = true; return s }(),
		func() NewGameSettings { s := canonicalNewGameSettings(); s.Players[1].RaceID = "human"; return s }(),
		func() NewGameSettings { s := canonicalNewGameSettings(); s.Players[1].SeatID = 1; return s }(),
		func() NewGameSettings { s := canonicalNewGameSettings(); s.Players[1].EmpireName = "human"; return s }(),
	}
	for i, settings := range cases {
		if _, err := rules.NewGame(7, settings); err == nil {
			t.Fatalf("case %d unexpectedly accepted", i)
		}
	}
}
