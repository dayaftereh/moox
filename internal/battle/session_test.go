package battle

import (
	"reflect"
	"testing"

	"moox/internal/protocol"
)

func TestBattleSessionLifecycleAndCopies(t *testing.T) {
	session, err := NewSession(Spec{
		ID:            7,
		GameID:        "game-1",
		StrategicTurn: 4,
		Participants:  []protocol.SeatID{2, 1},
		Seed:          99,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := session.View().Spec.Participants; !reflect.DeepEqual(got, []protocol.SeatID{1, 2}) {
		t.Fatalf("participants not normalized: %v", got)
	}
	if err := session.Start(); err != nil {
		t.Fatal(err)
	}
	if err := session.Complete(Result{WinnerSeats: []protocol.SeatID{2}, Outcome: "victory"}); err != nil {
		t.Fatal(err)
	}
	view := session.View()
	if view.Phase != PhaseCompleted || view.Result == nil || view.Result.Outcome != "victory" {
		t.Fatalf("unexpected completed view: %+v", view)
	}
	view.Spec.Participants[0] = 99
	if session.View().Spec.Participants[0] != 1 {
		t.Fatal("battle view leaked mutable participant storage")
	}
}

func TestDeriveSeedIsStableAndScoped(t *testing.T) {
	first := DeriveSeed(123, 8, 1)
	if first != DeriveSeed(123, 8, 1) {
		t.Fatal("same seed inputs produced different result")
	}
	if first == DeriveSeed(123, 8, 2) {
		t.Fatal("different battle IDs produced identical test seed")
	}
}
