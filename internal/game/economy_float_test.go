package game

import (
	"encoding/json"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func TestFractionalPopulationAssignmentProducesFractionalEconomy(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(712)
	command, err := NewAssignPopulationCommand(1, AssignPopulationPayload{ColonyID: state.Colonies[0].ID, Farmers: 1.25, Workers: 1.5, Scientists: 1.25})
	if err != nil {
		t.Fatal(err)
	}
	result, err := resolver.Resolve(ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: state.Empires[0].ID}}}, state, []protocol.CommandBatch{{
		SchemaVersion: protocol.CommandSchemaVersion, GameID: "fractional-population", SeatID: 1, Turn: 1, BaseRevision: 1, Commands: []protocol.Command{command},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Events) == 0 || result.Events[0].Kind != "colony.population_assigned" {
		t.Fatalf("missing assignment event: %+v", result.Events)
	}
	var assigned PopulationAssignedEvent
	if err := json.Unmarshal(result.Events[0].Data, &assigned); err != nil {
		t.Fatal(err)
	}
	wantPopulation := core.PopulationState{Total: 4, Farmers: 1.25, Workers: 1.5, Scientists: 1.25}
	if assigned.Current != wantPopulation {
		t.Fatalf("assigned population=%+v want=%+v", assigned.Current, wantPopulation)
	}
	wantBase := core.ColonyEconomy{Food: 2.5, Production: 4.5, Research: 3.75, TaxBC: 4}
	if assigned.BaseEconomy != wantBase {
		t.Fatalf("assignment base economy=%+v want=%+v", assigned.BaseEconomy, wantBase)
	}
	if assigned.AdjustedEconomy.Research != 5.625 {
		t.Fatalf("assignment fractional adjusted research=%v want=5.625", assigned.AdjustedEconomy.Research)
	}
	if findDomainEvent(result.Events, "colony.population_starved") == nil {
		t.Fatalf("food-deficit assignment should resolve starvation separately: %+v", result.Events)
	}
}

func TestConstructionPreservesFractionalProductionPP(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(713)
	state.Colonies[0].Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectBuilding, ProjectID: "holo_simulator"}
	state.Colonies[0].AdjustedEconomy.Production = 3.75
	state.Colonies[0].PopulationDynamics.ProductionAvailable = 3.75
	events, err := resolver.advanceConstruction(state)
	if err != nil {
		t.Fatal(err)
	}
	if state.Colonies[0].Construction == nil || state.Colonies[0].Construction.ProgressPP != 3.75 {
		t.Fatalf("fractional construction progress=%+v", state.Colonies[0].Construction)
	}
	if len(events) != 1 || events[0].Kind != "colony.construction_progressed" {
		t.Fatalf("unexpected construction events=%+v", events)
	}
}
