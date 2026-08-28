package session

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func TestGameSessionFreighterFleetConstructionChoiceAndCompletion(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	rules.MineralIndustryPerWorker["abundant"] = 60
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], game.NewGameTechnologyOptions{Level: game.NewGameTechnologyAverage}); err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("game-freighter-fleet", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	colonyID := state.Colonies[0].ID
	choices, err := s.ConstructionChoices(1, colonyID, rules)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, choice := range choices {
		if choice.ProjectKind != core.ConstructionProjectFreighterFleet {
			continue
		}
		found = true
		if choice.ProjectID != game.FreighterFleetProjectID || choice.ProductionCostPP != 50 || choice.FreightersAdded != 5 {
			t.Fatalf("Freighter Fleet legal action=%+v", choice)
		}
	}
	if !found {
		t.Fatal("seat legal-action surface missing Freighter Fleet")
	}
	if _, err := s.ConstructionChoices(2, colonyID, rules); err == nil {
		t.Fatal("foreign seat unexpectedly received construction choices for colony")
	}

	command, err := game.NewQueueFreighterFleetCommand(1, game.QueueFreighterFleetPayload{ColonyID: colonyID})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-freighter-fleet",
		SeatID:        1,
		Turn:          1,
		BaseRevision:  1,
		Commands:      []protocol.Command{command},
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-freighter-fleet",
		SeatID:        2,
		Turn:          1,
		BaseRevision:  1,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if observer.State.Empires[0].Freighters != 5 {
		t.Fatalf("Observer Freighters=%d want=5", observer.State.Empires[0].Freighters)
	}
	if observer.State.Colonies[0].Construction != nil {
		t.Fatalf("Observer retained completed Freighter Fleet project: %+v", observer.State.Colonies[0].Construction)
	}
	foundCompleted := false
	for _, event := range observer.Events {
		if event.Kind != "colony.freighter_fleet_completed" {
			continue
		}
		foundCompleted = true
		var payload game.FreighterFleetCompletedEvent
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.ColonyID != colonyID || payload.EmpireID != state.Empires[0].ID || payload.FreightersAdded != 5 || payload.TotalFreighters != 5 {
			t.Fatalf("Observer Freighter Fleet completion=%+v", payload)
		}
	}
	if !foundCompleted {
		t.Fatalf("Observer history missing Freighter Fleet completion: %+v", observer.Events)
	}
}
