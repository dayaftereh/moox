package session

import (
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func TestStatusProjectsLightweightSessionState(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("game-status", state, seats)
	if err != nil {
		t.Fatal(err)
	}

	status := s.Status()
	if status.GameID != "game-status" || status.Revision != 1 || status.Turn != 1 || status.Phase != PhasePlanning {
		t.Fatalf("unexpected status: %+v", status)
	}

	if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-status", SeatID: 1, Turn: 1, BaseRevision: 1}); err != nil {
		t.Fatal(err)
	}
	if got := s.Status(); got.Phase != PhasePlanning || got.Revision != 1 {
		t.Fatalf("partial submission must remain planning/revision1: %+v", got)
	}
}

func TestPlayerBattleViewsExposeOnlyParticipantBattles(t *testing.T) {
	state := core.NewSmallFixture(0x8008)
	second := core.Empire{ID: state.NewID(), Name: "Second", RaceID: "alkari"}
	third := core.Empire{ID: state.NewID(), Name: "Third", RaceID: "bulrathi"}
	state.Empires = append(state.Empires, second, third)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	seats := []Seat{
		{ID: 1, EmpireID: state.Empires[0].ID, Name: "One", Controller: ControllerLocalHuman},
		{ID: 2, EmpireID: second.ID, Name: "Two", Controller: ControllerRemoteHuman},
		{ID: 3, EmpireID: third.ID, Name: "Three", Controller: ControllerRemoteHuman},
	}
	s, err := NewGameSession("game-battle-projection", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	for _, seatID := range []protocol.SeatID{1, 2, 3} {
		if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-battle-projection", SeatID: seatID, Turn: 1, BaseRevision: 1}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.BeginEncounters([]EncounterSpec{
		{Participants: []protocol.SeatID{1, 2}},
		{Participants: []protocol.SeatID{2, 3}},
	}); err != nil {
		t.Fatal(err)
	}

	one, err := s.PlayerBattleViews(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(one) != 1 || one[0].Spec.ID != 1 {
		t.Fatalf("seat1 battle projection=%+v", one)
	}
	two, err := s.PlayerBattleViews(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(two) != 2 || two[0].Spec.ID != 1 || two[1].Spec.ID != 2 {
		t.Fatalf("seat2 battle projection=%+v", two)
	}
	three, err := s.PlayerBattleViews(3)
	if err != nil {
		t.Fatal(err)
	}
	if len(three) != 1 || three[0].Spec.ID != 2 {
		t.Fatalf("seat3 battle projection=%+v", three)
	}
	if _, err := s.PlayerBattleViews(99); err == nil {
		t.Fatal("unknown seat should be rejected")
	}
}
