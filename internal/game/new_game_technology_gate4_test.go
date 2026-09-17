package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
)

func TestNewGameTechnologyGate4FrozenInitialStates(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	for _, raceID := range []string{"human", "klackon"} {
		for _, level := range []NewGameTechnologyLevel{NewGameTechnologyPreWarp, NewGameTechnologyAverage} {
			t.Run(raceID+"/"+string(level), func(t *testing.T) {
				settings := canonicalNewGameSettings()
				settings.Players[0].RaceID = raceID
				settings.Players[0].EmpireName = raceID
				settings.TechnologyLevel = level
				result, err := rules.NewGame(0x1654, settings)
				if err != nil {
					t.Fatal(err)
				}
				assertFrozenTechnologyStart(t, result.State, level)
			})
		}
	}
}

func TestNewGameTechnologyGate4RepeatedDeterminism(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	for _, seed := range []uint64{0x1653, 0x1654, 0x1655} {
		for _, raceID := range []string{"human", "klackon"} {
			for _, level := range []NewGameTechnologyLevel{NewGameTechnologyPreWarp, NewGameTechnologyAverage} {
				settings := canonicalNewGameSettings()
				settings.Players[0].RaceID = raceID
				settings.Players[0].EmpireName = raceID
				settings.TechnologyLevel = level
				first, err := rules.NewGame(seed, settings)
				if err != nil {
					t.Fatal(err)
				}
				second, err := rules.NewGame(seed, settings)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(first, second) {
					t.Fatalf("seed=%#x race=%s level=%s produced different states", seed, raceID, level)
				}
				firstJSON, err := json.Marshal(first.State)
				if err != nil {
					t.Fatal(err)
				}
				secondJSON, err := json.Marshal(second.State)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(firstJSON, secondJSON) {
					t.Fatalf("seed=%#x race=%s level=%s produced different JSON", seed, raceID, level)
				}
			}
		}
	}
}

func assertFrozenTechnologyStart(t *testing.T, state *core.GameState, level NewGameTechnologyLevel) {
	t.Helper()
	if state == nil {
		t.Fatal("state is nil")
	}
	if len(state.Colonies) != len(state.Empires) {
		t.Fatalf("colonies=%d empires=%d want one home colony per empire", len(state.Colonies), len(state.Empires))
	}

	preWarpTech := []int{32, 40, 103, 145, 166, 168}
	preWarpFields := []int{0, 29}
	averageFields := []int{0, 22, 23, 28, 29, 55, 57}

	for _, empire := range state.Empires {
		if empire.Freighters != 0 {
			t.Fatalf("empire %d freighters=%d want=0", empire.ID, empire.Freighters)
		}
		if empire.Treasury.BalanceBC != 50 {
			t.Fatalf("empire %d treasury=%v want=50", empire.ID, empire.Treasury.BalanceBC)
		}

		var colonies []core.Colony
		for _, colony := range state.Colonies {
			if colony.EmpireID == empire.ID {
				colonies = append(colonies, colony)
			}
		}
		if len(colonies) != 1 {
			t.Fatalf("empire %d colonies=%d want=1", empire.ID, len(colonies))
		}
		colony := colonies[0]
		if empire.Capital != colony.ID {
			t.Fatalf("empire %d capital=%d home colony=%d", empire.ID, empire.Capital, colony.ID)
		}
		if got := colony.Population.Total(); got != 8 {
			t.Fatalf("empire %d population=%v want=8", empire.ID, got)
		}
		if got := colony.Population.Farmers(); got != 4 {
			t.Fatalf("empire %d farmers=%v want=4", empire.ID, got)
		}
		if got := colony.Population.Workers(); got != 2 {
			t.Fatalf("empire %d workers=%v want=2", empire.ID, got)
		}
		if got := colony.Population.Scientists(); got != 2 {
			t.Fatalf("empire %d scientists=%v want=2", empire.ID, got)
		}

		switch level {
		case NewGameTechnologyPreWarp:
			if !reflect.DeepEqual(empire.KnownTechnologyFieldIDs, preWarpFields) {
				t.Fatalf("empire %d Pre-Warp fields=%v want=%v", empire.ID, empire.KnownTechnologyFieldIDs, preWarpFields)
			}
			if !reflect.DeepEqual(empire.KnownTechnologyIDs, preWarpTech) {
				t.Fatalf("empire %d Pre-Warp technologies=%v want=%v", empire.ID, empire.KnownTechnologyIDs, preWarpTech)
			}
		case NewGameTechnologyAverage:
			if !reflect.DeepEqual(empire.KnownTechnologyFieldIDs, averageFields) {
				t.Fatalf("empire %d Average fields=%v want=%v", empire.ID, empire.KnownTechnologyFieldIDs, averageFields)
			}
			if got := len(empire.KnownTechnologyIDs); got != 20 {
				t.Fatalf("empire %d Average known applications=%d want=20", empire.ID, got)
			}
		default:
			t.Fatalf("unsupported test level %q", level)
		}
	}

	if level == NewGameTechnologyPreWarp {
		if len(state.Ships) != 0 || len(state.ShipDesigns) != 0 || len(state.StrategicFleets) != 0 {
			t.Fatalf("Pre-Warp ships/designs/fleets=%d/%d/%d want=0/0/0", len(state.Ships), len(state.ShipDesigns), len(state.StrategicFleets))
		}
		return
	}

	if len(state.Ships) != 2*len(state.Empires) {
		t.Fatalf("Average ships=%d want=%d scouts", len(state.Ships), 2*len(state.Empires))
	}
	if len(state.ShipDesigns) != len(state.Empires) {
		t.Fatalf("Average ship designs=%d want=%d", len(state.ShipDesigns), len(state.Empires))
	}
	if len(state.StrategicFleets) != 2*len(state.Empires) {
		t.Fatalf("Average strategic fleets=%d want=%d", len(state.StrategicFleets), 2*len(state.Empires))
	}
	for _, empire := range state.Empires {
		scoutShips := 0
		for _, ship := range state.Ships {
			if ship.EmpireID == empire.ID {
				scoutShips++
			}
		}
		if scoutShips != 2 {
			t.Fatalf("empire %d scouts=%d want=2", empire.ID, scoutShips)
		}
		combatFleets := 0
		colonyFleets := 0
		for _, fleet := range state.StrategicFleets {
			if fleet.EmpireID != empire.ID {
				continue
			}
			if fleet.Role == core.StrategicFleetRoleCombat {
				combatFleets++
				if len(fleet.ShipIDs) != 2 {
					t.Fatalf("empire %d scout fleet ships=%d want=2", empire.ID, len(fleet.ShipIDs))
				}
			}
			if fleet.SpecialKind == core.StrategicFleetSpecialColonyShip {
				colonyFleets++
			}
		}
		if combatFleets != 1 || colonyFleets != 1 {
			t.Fatalf("empire %d combat/colony fleets=%d/%d want=1/1", empire.ID, combatFleets, colonyFleets)
		}
	}
}
