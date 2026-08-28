package game

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func TestFreighterFleetConstructionChoiceRequiresFreightersTechnology(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0x9101)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	if err := rules.InitializeEmpireTechnologies(empire, NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	choices, err := rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, choice := range choices {
		if choice.ProjectKind == core.ConstructionProjectFreighterFleet {
			t.Fatal("Pre-Warp Empire unexpectedly can build Freighter Fleet before Technology 69")
		}
	}
	empire.KnownTechnologyIDs = appendUniqueSortedInt(empire.KnownTechnologyIDs, FreighterFleetTechnologyID)
	choices, err = rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, choice := range choices {
		if choice.ProjectKind != core.ConstructionProjectFreighterFleet {
			continue
		}
		found = true
		if choice.ProjectID != FreighterFleetProjectID || choice.ProductionCostPP != 50 || choice.TechnologyID != 69 || choice.FreightersAdded != 5 {
			t.Fatalf("Freighter Fleet choice=%+v", choice)
		}
	}
	if !found {
		t.Fatal("Technology 69 did not expose Freighter Fleet construction choice")
	}
}

func TestQueueFreighterFleetUsesGenericConstructionState(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0x9102)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	if err := rules.InitializeEmpireTechnologies(empire, NewGameTechnologyOptions{Level: NewGameTechnologyAverage}); err != nil {
		t.Fatal(err)
	}
	command, err := NewQueueFreighterFleetCommand(7, QueueFreighterFleetPayload{ColonyID: colony.ID})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.queueFreighterFleet(state, empire.ID, protocol.SeatID(1), command)
	if err != nil {
		t.Fatal(err)
	}
	if colony.Construction == nil || colony.Construction.ProjectKind != core.ConstructionProjectFreighterFleet || colony.Construction.ProjectID != FreighterFleetProjectID || colony.Construction.ProgressPP != 0 {
		t.Fatalf("Freighter Fleet construction state=%+v", colony.Construction)
	}
	if event.Kind != "colony.freighter_fleet_queued" || event.SeatID != 1 || event.CommandSequence != 7 {
		t.Fatalf("queue event=%+v", event)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("queued Freighter Fleet state invalid: %v", err)
	}
}
func TestQueueFreighterFleetResolvesProductionAndAddsFiveFreighters(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	rules.MineralIndustryPerWorker["abundant"] = 60
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0x9103)
	empire := &state.Empires[0]
	if err := rules.InitializeEmpireTechnologies(empire, NewGameTechnologyOptions{Level: NewGameTechnologyAverage}); err != nil {
		t.Fatal(err)
	}
	command, err := NewQueueFreighterFleetCommand(1, QueueFreighterFleetPayload{ColonyID: state.Colonies[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	ctx, batches := constructionBatch(t, state, command)
	result, err := resolver.Resolve(ctx, state, batches)
	if err != nil {
		t.Fatal(err)
	}
	if result.State.Colonies[0].Construction != nil {
		t.Fatalf("completed Freighter Fleet left construction active: %+v", result.State.Colonies[0].Construction)
	}
	if result.State.Empires[0].Freighters != 5 {
		t.Fatalf("Freighters=%d want=5", result.State.Empires[0].Freighters)
	}
	progressEvent := findDomainEvent(result.Events, "colony.construction_progressed")
	if progressEvent == nil {
		t.Fatalf("missing construction progress event: %+v", result.Events)
	}
	var progress ConstructionProgressedEvent
	if err := json.Unmarshal(progressEvent.Data, &progress); err != nil {
		t.Fatal(err)
	}
	if progress.ProjectKind != core.ConstructionProjectFreighterFleet || progress.ProjectID != FreighterFleetProjectID || progress.AppliedPP != 50 || progress.RemainingPP != 0 {
		t.Fatalf("Freighter Fleet progress=%+v", progress)
	}
	completedEvent := findDomainEvent(result.Events, "colony.freighter_fleet_completed")
	if completedEvent == nil {
		t.Fatalf("missing Freighter Fleet completion event: %+v", result.Events)
	}
	var completed FreighterFleetCompletedEvent
	if err := json.Unmarshal(completedEvent.Data, &completed); err != nil {
		t.Fatal(err)
	}
	if completed.EmpireID != empire.ID || completed.ColonyID != state.Colonies[0].ID || completed.FreightersAdded != 5 || completed.TotalFreighters != 5 {
		t.Fatalf("Freighter Fleet completion=%+v", completed)
	}
}
