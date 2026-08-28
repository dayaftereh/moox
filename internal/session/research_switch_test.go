package session

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"moox/internal/game"
	"moox/internal/protocol"
)

func TestGameSessionResearchSwitchPreservesPriorProgressAndObserverEvent(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], game.NewGameTechnologyOptions{Level: game.NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("game-research-switch", state, seats)
	if err != nil {
		t.Fatal(err)
	}

	selectField4, err := game.NewSelectResearchCommand(1, game.SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-research-switch",
		SeatID:        1,
		Turn:          1,
		BaseRevision:  1,
		Commands:      []protocol.Command{selectField4},
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-research-switch", SeatID: 2, Turn: 1, BaseRevision: 1}); err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	first, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if first.State.Empires[0].Research == nil || first.State.Empires[0].Research.ProgressRP <= 0 {
		t.Fatalf("first project has no accumulated research: %+v", first.State.Empires[0].Research)
	}
	transferred := first.State.Empires[0].Research.ProgressRP
	if err := s.CompleteTurn(); err != nil {
		t.Fatal(err)
	}

	switchTo55, err := game.NewSelectResearchCommand(1, game.SelectResearchPayload{TechFieldID: 55})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-research-switch",
		SeatID:        1,
		Turn:          2,
		BaseRevision:  3,
		Commands:      []protocol.Command{switchTo55},
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-research-switch", SeatID: 2, Turn: 2, BaseRevision: 3}); err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if observer.State.Empires[0].Research == nil || observer.State.Empires[0].Research.TechFieldID != 55 {
		t.Fatalf("switched project not authoritative: %+v", observer.State.Empires[0].Research)
	}
	if observer.State.Empires[0].Research.ProgressRP <= transferred {
		t.Fatalf("current-turn RP did not advance after transferred pool: transferred=%v current=%v", transferred, observer.State.Empires[0].Research.ProgressRP)
	}

	found := false
	for _, event := range observer.Events {
		if event.Kind != "empire.research_switched" {
			continue
		}
		var payload game.ResearchSwitchedEvent
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.Previous.TechFieldID == 4 && payload.Current.TechFieldID == 55 && payload.TransferredRP == transferred {
			found = true
		}
	}
	if !found {
		t.Fatalf("Observer missing research_switched with transferred RP=%v", transferred)
	}
}
