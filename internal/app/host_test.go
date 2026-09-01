package app

import (
	"errors"
	"testing"
	"time"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

type passResolver struct{}

func (passResolver) Resolve(_ game.ResolveContext, state *core.GameState, _ []protocol.CommandBatch) (game.Resolution, error) {
	return game.Resolution{State: state}, nil
}

func newTwoSeatHost(t *testing.T) (*Host, *session.GameSession) {
	t.Helper()
	state := core.NewSmallFixture(0x80080001)
	second := core.Empire{ID: state.NewID(), Name: "Second", RaceID: "alkari"}
	state.Empires = append(state.Empires, second)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	gameSession, err := session.NewGameSession("game-app", state, []session.Seat{
		{ID: 1, EmpireID: state.Empires[0].ID, Name: "One", Controller: session.ControllerLocalHuman},
		{ID: 2, EmpireID: second.ID, Name: "Two", Controller: session.ControllerRemoteHuman},
	})
	if err != nil {
		t.Fatal(err)
	}
	host := NewHost()
	if err := host.Register(Registration{Session: gameSession, Resolver: passResolver{}}); err != nil {
		t.Fatal(err)
	}
	return host, gameSession
}

func TestRegisterListAndPlayerSnapshotStartAtChangeSequenceOne(t *testing.T) {
	host, _ := newTwoSeatHost(t)
	games := host.ListGames()
	if len(games) != 1 || games[0].GameID != "game-app" || games[0].ChangeSequence != 1 || games[0].Revision != 1 || games[0].Turn != 1 {
		t.Fatalf("unexpected game list: %+v", games)
	}
	snapshot, err := host.PlayerSnapshot("game-app", 1)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.SchemaVersion != 1 || snapshot.ChangeSequence != 1 || snapshot.View.GameID != "game-app" || len(snapshot.Battles) != 0 {
		t.Fatalf("unexpected player snapshot: %+v", snapshot)
	}
	if _, err := host.PlayerSnapshot("game-app", 99); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown seat error=%v", err)
	}
	if _, err := host.ObserverSnapshot("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown game error=%v", err)
	}
}

func TestChangeSequenceTracksProjectionChangesIndependentFromGameRevision(t *testing.T) {
	host, _ := newTwoSeatHost(t)
	notifications, cancel, err := host.Subscribe("game-app")
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	first := protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-app", SeatID: 1, Turn: 1, BaseRevision: 1}
	receipt, err := host.SubmitTurn("game-app", first)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.ChangeSequence != 2 || receipt.GameRevision != 1 {
		t.Fatalf("partial submission receipt=%+v", receipt)
	}
	note := receiveNotification(t, notifications)
	if note.ChangeSequence != 2 || note.GameRevision != 1 || note.Reason != "submission" || note.Scope != "session" {
		t.Fatalf("partial submission notification=%+v", note)
	}

	if _, err := host.SubmitTurn("game-app", first); !errors.Is(err, ErrSessionRejected) {
		t.Fatalf("duplicate submission error=%v", err)
	}
	afterReject, err := host.PlayerSnapshot("game-app", 1)
	if err != nil {
		t.Fatal(err)
	}
	if afterReject.ChangeSequence != 2 || afterReject.View.Revision != 1 {
		t.Fatalf("rejection advanced state: %+v", afterReject)
	}

	second := protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-app", SeatID: 2, Turn: 1, BaseRevision: 1}
	receipt, err = host.SubmitTurn("game-app", second)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.ChangeSequence != 3 || receipt.GameRevision != 3 {
		t.Fatalf("final submission receipt=%+v", receipt)
	}
	note = receiveNotification(t, notifications)
	if note.ChangeSequence != 3 || note.GameRevision != 3 || note.Reason != "turn_advanced" {
		t.Fatalf("final submission notification=%+v", note)
	}
	finalSnapshot, err := host.PlayerSnapshot("game-app", 1)
	if err != nil {
		t.Fatal(err)
	}
	if finalSnapshot.ChangeSequence != 3 || finalSnapshot.View.Turn != 2 || finalSnapshot.View.Revision != 3 || finalSnapshot.View.Phase != session.PhasePlanning {
		t.Fatalf("automatic phase driver result=%+v", finalSnapshot)
	}
}

func TestSubscriberKeepsNewestInvalidationForSlowClient(t *testing.T) {
	host, _ := newTwoSeatHost(t)
	notifications, cancel, err := host.Subscribe("game-app")
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	if _, err := host.SubmitTurn("game-app", protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-app", SeatID: 1, Turn: 1, BaseRevision: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := host.SubmitTurn("game-app", protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-app", SeatID: 2, Turn: 1, BaseRevision: 1}); err != nil {
		t.Fatal(err)
	}
	note := receiveNotification(t, notifications)
	if note.ChangeSequence != 3 || note.GameRevision != 3 {
		t.Fatalf("slow subscriber did not retain newest notification: %+v", note)
	}
}

func receiveNotification(t *testing.T, notifications <-chan Notification) Notification {
	t.Helper()
	select {
	case note, ok := <-notifications:
		if !ok {
			t.Fatal("notification channel closed unexpectedly")
		}
		return note
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for notification")
		return Notification{}
	}
}
