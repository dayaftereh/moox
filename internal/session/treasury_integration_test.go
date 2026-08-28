package session

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"moox/internal/game"
	"moox/internal/protocol"
)

func TestGameSessionTreasurySettlementAppearsInObserverHistory(t *testing.T) {
	state, seats := twoSeatFixture(t)
	if err := game.InitializeNewGameTreasury(state); err != nil {
		t.Fatal(err)
	}
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("game-treasury", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	for _, seatID := range []protocol.SeatID{1, 2} {
		if err := s.SubmitTurn(protocol.CommandBatch{
			SchemaVersion: protocol.CommandSchemaVersion,
			GameID:        "game-treasury",
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
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if observer.State.Empires[0].Treasury.BalanceBC <= 50 {
		t.Fatalf("Treasury did not settle current-turn income: %+v", observer.State.Empires[0].Treasury)
	}
	found := false
	for _, event := range observer.Events {
		if event.Kind != "empire.treasury_settled" {
			continue
		}
		var payload game.TreasurySettledEvent
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.EmpireID != observer.State.Empires[0].ID {
			continue
		}
		found = true
		if payload.PreviousBalanceBC != 50 || payload.Current != observer.State.Empires[0].Treasury {
			t.Fatalf("Treasury Observer payload=%+v state=%+v", payload, observer.State.Empires[0].Treasury)
		}
	}
	if !found {
		t.Fatal("Observer history missing Treasury settlement event")
	}
}
