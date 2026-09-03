package game

import (
	"bytes"
	"encoding/json"
	"sort"
	"testing"

	"moox/internal/core"
)

func TestDecisionQueriesAreAuthoritativeDeterministicAndNonMutating(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	generated, err := rules.NewGame(0x8009, canonicalNewGameSettings())
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	humanID := generated.Players[0].EmpireID
	darlokID := generated.Players[1].EmpireID
	for i := range generated.State.Empires {
		if generated.State.Empires[i].ID == humanID {
			generated.State.Empires[i].KnownTechnologyIDs = append(generated.State.Empires[i].KnownTechnologyIDs, 194)
			sort.Ints(generated.State.Empires[i].KnownTechnologyIDs)
		}
	}
	before, err := json.Marshal(generated.State)
	if err != nil {
		t.Fatal(err)
	}
	var humanColonyID, darlokColonyID core.ID
	for _, colony := range generated.State.Colonies {
		switch colony.EmpireID {
		case humanID:
			humanColonyID = colony.ID
		case darlokID:
			darlokColonyID = colony.ID
		}
	}
	if humanColonyID == 0 || darlokColonyID == 0 {
		t.Fatalf("canonical colonies human=%d darlok=%d", humanColonyID, darlokColonyID)
	}

	population, err := resolver.AvailablePopulationChoices(generated.State, darlokID, darlokColonyID)
	if err != nil {
		t.Fatal(err)
	}
	foodSafe := 0
	for _, choice := range population {
		if choice.FoodSafe {
			foodSafe++
		}
	}
	if foodSafe == 0 {
		t.Fatalf("Darlok canonical colony has no server-projected food-safe population choice: %+v", population)
	}
	var darlokColony *core.Colony
	for i := range generated.State.Colonies {
		if generated.State.Colonies[i].ID == darlokColonyID {
			darlokColony = &generated.State.Colonies[i]
			break
		}
	}
	if darlokColony == nil || darlokColony.AdjustedEconomy.Food+1e-9 >= darlokColony.PopulationDynamics.FoodRequired {
		t.Fatalf("canonical Darlok start unexpectedly already food-safe: %+v", darlokColony)
	}

	moves, err := resolver.AvailableFleetMoveChoices(generated.State, humanID)
	if err != nil {
		t.Fatal(err)
	}
	wholeCombat, splitCombat := 0, 0
	for _, move := range moves {
		if move.FleetID != 59 {
			continue
		}
		if len(move.ShipIDs) == 0 {
			wholeCombat++
		}
		if len(move.ShipIDs) == 1 {
			splitCombat++
		}
	}
	if wholeCombat == 0 || splitCombat == 0 {
		t.Fatalf("combat move catalog lacks whole/subset legal moves: whole=%d split=%d choices=%+v", wholeCombat, splitCombat, moves)
	}

	colonization, err := resolver.AvailableColonizationChoices(generated.State, humanID)
	if err != nil {
		t.Fatal(err)
	}
	if len(colonization) == 0 {
		t.Fatal("canonical Human Colony Ship has no legal same-system colonization choice")
	}

	after, err := json.Marshal(generated.State)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("decision queries mutated authoritative GameState")
	}

	population2, err := resolver.AvailablePopulationChoices(generated.State, darlokID, darlokColonyID)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, _ := json.Marshal(population)
	secondJSON, _ := json.Marshal(population2)
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatal("population decision catalog is not deterministic")
	}
}
