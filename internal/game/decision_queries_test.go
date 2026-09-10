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
func TestFleetMoveTargetsProjectLegalAndOutOfRangeDestinations(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0x5B)
	empireID := state.Empires[0].ID
	state.Empires[0].KnownTechnologyIDs = []int{standardFuelCellsTechnologyID}
	source := &state.Galaxy.Systems[0]
	legalSystem := &state.Galaxy.Systems[1]
	farSystem := &state.Galaxy.Systems[2]
	source.X, source.Y = 0, 0
	legalSystem.X, legalSystem.Y = 60, 0
	farSystem.X, farSystem.Y = 150, 0
	fleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID: fleetID, EmpireID: empireID, Role: core.StrategicFleetRoleCivilian,
		SpecialKind: core.StrategicFleetSpecialColonyShip, AtSystemID: source.ID, FTLSpeed: 2,
	})

	targets, err := resolver.AvailableFleetMoveTargets(state, empireID)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 2 {
		t.Fatalf("targets=%+v want exactly two non-source systems", targets)
	}
	legal := targets[0]
	illegal := targets[1]
	if legal.DestinationSystemID != legalSystem.ID || !legal.Legal || legal.Reason != "" || legal.DistanceParsecs != 2 || legal.SupplyDistanceParsecs != 2 || legal.FuelRangeParsecs != 4 || legal.ETA != 1 {
		t.Fatalf("2 pc target=%+v", legal)
	}
	if illegal.DestinationSystemID != farSystem.ID || illegal.Legal || illegal.Reason != FleetMoveTargetReasonOutOfFuelRange || illegal.DistanceParsecs != 5 || illegal.SupplyDistanceParsecs != 5 || illegal.FuelRangeParsecs != 4 || illegal.ETA != 3 {
		t.Fatalf("5 pc target=%+v", illegal)
	}
	choices, err := resolver.AvailableFleetMoveChoices(state, empireID)
	if err != nil {
		t.Fatal(err)
	}
	if len(choices) != 1 || choices[0].DestinationSystemID != legalSystem.ID || choices[0].ETA != legal.ETA || choices[0].FuelRangeParsecs != legal.FuelRangeParsecs || choices[0].SupplyDistanceParsecs != legal.SupplyDistanceParsecs {
		t.Fatalf("legal choices=%+v want only the 2 pc target derived from target authority", choices)
	}
	farMove, err := NewMoveFleetCommand(1, MoveFleetPayload{FleetID: fleetID, DestinationSystemID: farSystem.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.moveFleetEvents(state, empireID, 1, farMove); err == nil {
		t.Fatal("5 pc out-of-range target was projected illegal but accepted by move command validation")
	}
	encoded, err := json.Marshal(targets)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte(legalSystem.Name)) || bytes.Contains(encoded, []byte(farSystem.Name)) {
		t.Fatalf("target projection leaked star names: %s", encoded)
	}
}
