package session

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func TestGameSessionGrantTechnologyCommitsObserverEvent(t *testing.T) {
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
	s, err := NewGameSession("game-technology-grant", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	for _, seatID := range []protocol.SeatID{1, 2} {
		if err := s.SubmitTurn(protocol.CommandBatch{
			SchemaVersion: protocol.CommandSchemaVersion,
			GameID:        "game-technology-grant",
			SeatID:        seatID,
			Turn:          1,
			BaseRevision:  1,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}

	empireID := state.Empires[0].ID
	if err := s.GrantTechnology(empireID, 155, resolver, game.TechnologyGrantOptions{SourceKind: "integration_test"}); err != nil {
		t.Fatal(err)
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if observer.Phase != PhasePostResolution || observer.Revision != 3 {
		t.Fatalf("unexpected phase/revision after grant: %q/%d", observer.Phase, observer.Revision)
	}
	empire := empireByIDGrantTest(observer.State, empireID)
	if empire == nil {
		t.Fatalf("missing Empire %d", empireID)
	}
	if !containsIntGrantTest(empire.KnownTechnologyIDs, 155) {
		t.Fatal("Observer state missing granted Research Laboratory Technology")
	}
	if containsIntGrantTest(empire.KnownTechnologyFieldIDs, 56) {
		t.Fatal("Session grant invented TechField 56 completion")
	}

	found := false
	for _, event := range observer.Events {
		if event.Kind != game.EventTechnologyGranted {
			continue
		}
		found = true
		if event.Revision != 3 || event.Scope != protocol.EventScopeStrategic || event.SeatID != 0 || event.CommandSequence != 0 {
			t.Fatalf("technology grant event envelope=%+v", event)
		}
		var payload game.TechnologyGrantedEvent
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.EmpireID != empireID || payload.TechnologyID != 155 || payload.TechnologyKey != "research_laboratory" || payload.SourceKind != "integration_test" {
			t.Fatalf("technology grant payload=%+v", payload)
		}
	}
	if !found {
		t.Fatal("Observer event history missing technology grant")
	}
}

func empireByIDGrantTest(state *core.GameState, empireID core.ID) *core.Empire {
	for i := range state.Empires {
		if state.Empires[i].ID == empireID {
			return &state.Empires[i]
		}
	}
	return nil
}

func containsIntGrantTest(values []int, want int) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
