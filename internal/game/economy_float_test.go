package game

import (
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
	command, err := NewAssignPopulationCommand(1, AssignPopulationPayload{
		ColonyID:   state.Colonies[0].ID,
		Farmers:    1.25,
		Workers:    1.5,
		Scientists: 1.25,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := resolver.Resolve(
		ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: state.Empires[0].ID}}},
		state,
		[]protocol.CommandBatch{{
			SchemaVersion: protocol.CommandSchemaVersion,
			GameID:        "fractional-population",
			SeatID:        1,
			Turn:          1,
			BaseRevision:  1,
			Commands:      []protocol.Command{command},
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	colony := result.State.Colonies[0]
	wantPopulation := core.PopulationState{Total: 4, Farmers: 1.25, Workers: 1.5, Scientists: 1.25}
	if colony.Population != wantPopulation {
		t.Fatalf("population=%+v want=%+v", colony.Population, wantPopulation)
	}
	wantBase := core.ColonyEconomy{Food: 2.5, Production: 4.5, Research: 3.75, TaxBC: 4}
	if colony.Economy != wantBase {
		t.Fatalf("base economy=%+v want=%+v", colony.Economy, wantBase)
	}
	if colony.AdjustedEconomy.Research != 5.625 {
		t.Fatalf("fractional adjusted research=%v want=5.625", colony.AdjustedEconomy.Research)
	}
}

func TestConstructionPreservesFractionalProductionPP(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(713)
	state.Colonies[0].Construction = &core.ConstructionState{BuildingID: "holo_simulator"}
	state.Colonies[0].AdjustedEconomy.Production = 3.75
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
