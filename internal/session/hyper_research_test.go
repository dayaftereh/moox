package session

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func TestGameSessionHyperAdvancedResearchUsesSharedLegalActionSurface(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	state.Empires[0].KnownTechnologyFieldIDs = []int{70}

	s, err := NewGameSession("game-hyper-research", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	choices, err := s.ResearchChoices(1, rules)
	if err != nil {
		t.Fatal(err)
	}
	var hyper *game.ResearchChoice
	for i := range choices {
		if choices[i].TechFieldID == 75 {
			hyper = &choices[i]
			break
		}
	}
	if hyper == nil || hyper.SelectionMode != core.ResearchSelectionRepeatField || hyper.BaseCostRP != 15000 || hyper.ResearchLevel != 1 || len(hyper.TechnologyIDs) != 0 {
		t.Fatalf("session Hyper-Advanced legal action=%+v", hyper)
	}

	command, err := game.NewSelectResearchCommand(1, game.SelectResearchPayload{TechFieldID: 75})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-hyper-research", SeatID: 1, Turn: 1, BaseRevision: 1, Commands: []protocol.Command{command}}); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-hyper-research", SeatID: 2, Turn: 1, BaseRevision: 1}); err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}

	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	research := observer.State.Empires[0].Research
	if research == nil || research.TechFieldID != 75 || research.SelectionMode != core.ResearchSelectionRepeatField || len(research.TechnologyIDs) != 0 {
		t.Fatalf("authoritative Hyper-Advanced research=%+v", research)
	}
	found := false
	for _, event := range observer.Events {
		if event.Kind != "empire.research_selected" {
			continue
		}
		var selected game.ResearchSelectedEvent
		if err := json.Unmarshal(event.Data, &selected); err != nil {
			t.Fatal(err)
		}
		if selected.TechFieldID == 75 && selected.SelectionMode == core.ResearchSelectionRepeatField && selected.BaseCostRP == 15000 && selected.ResearchLevel == 1 && len(selected.TechnologyIDs) == 0 {
			found = true
		}
	}
	if !found {
		t.Fatal("Observer did not receive Hyper-Advanced research_selected metadata")
	}
}
