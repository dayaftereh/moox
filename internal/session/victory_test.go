package session

import (
	"bytes"
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func completedInvasionFixture(t *testing.T, gameID string) (*GameSession, core.ID, core.ID) {
	t.Helper()
	s, resolver, attackerID, defenderID, colonyID, transports := newInvasionSession(t, 0xB221, gameID, 2)
	submitEmptyInvasionTurn(t, s)
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	before := s.Status()
	command, err := game.NewInvadeCommand(1, game.InvadePayload{ColonyID: colonyID, TransportFleetIDs: append([]core.ID(nil), transports...)})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveInvasionCommand(1, before.Revision, command); err != nil {
		t.Fatal(err)
	}
	status := s.Status()
	if status.Phase != PhaseCompleted || status.Result == nil || status.Result.WinnerEmpireID != attackerID {
		t.Fatalf("fixture did not complete: %+v", status)
	}
	return s, attackerID, defenderID
}

func TestEliminatedSeatDoesNotBlockOngoingTurnAndRemainsReadable(t *testing.T) {
	state := core.NewSmallFixture(0xE11A)
	secondID := state.NewID()
	thirdID := state.NewID()
	state.Empires = append(state.Empires,
		core.Empire{ID: secondID, Name: "Second", RaceID: "human"},
		core.Empire{ID: thirdID, Name: "Eliminated", RaceID: "human"},
	)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("three-seat-elimination", state, []Seat{
		{ID: 1, EmpireID: state.Empires[0].ID, Name: "First", Controller: ControllerLocalHuman},
		{ID: 2, EmpireID: secondID, Name: "Second", Controller: ControllerLocalHuman},
		{ID: 3, EmpireID: thirdID, Name: "Eliminated", Controller: ControllerLocalHuman},
	})
	if err != nil {
		t.Fatal(err)
	}
	s.eliminatedEmpires = []core.ID{thirdID}
	status := s.Status()
	if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: status.GameID, SeatID: 3, Turn: status.Turn, BaseRevision: status.Revision}); err == nil {
		t.Fatal("eliminated seat submitted a gameplay turn")
	}
	for _, seatID := range []protocol.SeatID{1, 2} {
		if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: status.GameID, SeatID: seatID, Turn: status.Turn, BaseRevision: status.Revision}); err != nil {
			t.Fatalf("active seat %d submit: %v", seatID, err)
		}
	}
	if s.Status().Phase != PhaseStrategicResolution {
		t.Fatalf("eliminated seat blocked turn readiness: %+v", s.Status())
	}
	batches, err := s.SubmittedBatches()
	if err != nil {
		t.Fatal(err)
	}
	if len(batches) != 2 || batches[0].SeatID != 1 || batches[1].SeatID != 2 {
		t.Fatalf("active submitted batches=%+v", batches)
	}
	view, err := s.PlayerView(3)
	if err != nil {
		t.Fatalf("eliminated seat lost read access: %v", err)
	}
	if !reflect.DeepEqual(view.EliminatedEmpireIDs, []core.ID{thirdID}) {
		t.Fatalf("eliminated seat view IDs=%v", view.EliminatedEmpireIDs)
	}
}
func TestCompletedSessionSnapshotExactRoundTrip(t *testing.T) {
	s, _, _ := completedInvasionFixture(t, "completed-roundtrip")
	beforeView, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := s.MarshalCompletedSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := UnmarshalCompletedSnapshot(encoded)
	if err != nil {
		t.Fatal(err)
	}
	afterView, err := restored.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(beforeView, afterView) {
		t.Fatalf("completed snapshot view differs after roundtrip\nbefore=%+v\nafter=%+v", beforeView, afterView)
	}
	reencoded, err := restored.MarshalCompletedSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatalf("completed snapshot bytes changed after roundtrip\nbefore=%s\nafter=%s", encoded, reencoded)
	}
}

func TestCompletedSessionRejectsGameplayMutationAtomically(t *testing.T) {
	s, _, defenderID := completedInvasionFixture(t, "completed-readonly")
	before := s.Status()
	view, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	beforeEvents := append([]protocol.DomainEvent(nil), view.Events...)

	batch := protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: before.GameID, SeatID: 1, Turn: before.Turn, BaseRevision: before.Revision}
	if err := s.SubmitTurn(batch); err == nil {
		t.Fatal("completed session accepted strategic turn")
	}
	if err := s.CompleteTurn(); err == nil {
		t.Fatal("completed session accepted turn completion")
	}
	if err := s.PublishDraftTelemetry(1, "post_game", "invalid", nil); err == nil {
		t.Fatal("completed session accepted draft telemetry")
	}
	command, err := game.NewDeclareWarCommand(1, defenderID)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveDiplomacyCommand(1, before.Revision, command); err == nil {
		t.Fatal("completed session accepted diplomacy mutation")
	}

	after := s.Status()
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("rejected post-game writes changed status before=%+v after=%+v", before, after)
	}
	afterView, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(beforeEvents, afterView.Events) {
		t.Fatal("rejected post-game writes changed event history")
	}
}
