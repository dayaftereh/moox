package battle

import (
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func strategicBattleTestSpec() Spec {
	return Spec{
		ID: 11, GameID: "strategic", StrategicTurn: 3, SystemID: 7,
		Attacker:          Side{EmpireID: 20, SeatID: 2, CombatFleetIDs: []core.ID{30}, ShipIDs: []core.ID{31, 32}, CivilianFleetIDs: []core.ID{33}},
		Defender:          Side{EmpireID: 40, SeatID: 1, CombatFleetIDs: []core.ID{50}, ShipIDs: []core.ID{51, 52}},
		DefenderColonyIDs: []core.ID{60}, Participants: []protocol.SeatID{2, 1}, Seed: 99,
	}
}

func TestStrategicBattleSpecPreservesDirectionAndDetachedIdentity(t *testing.T) {
	s, err := NewSession(strategicBattleTestSpec())
	if err != nil {
		t.Fatal(err)
	}
	view := s.View()
	if view.Spec.Attacker.SeatID != 2 || view.Spec.Defender.SeatID != 1 || !reflect.DeepEqual(view.Spec.Participants, []protocol.SeatID{1, 2}) {
		t.Fatalf("direction/participants=%+v", view.Spec)
	}
	view.Spec.Attacker.ShipIDs[0] = 999
	view.Spec.DefenderColonyIDs[0] = 999
	fresh := s.View().Spec
	if fresh.Attacker.ShipIDs[0] != 31 || fresh.DefenderColonyIDs[0] != 60 {
		t.Fatal("battle strategic identity leaked mutable slice storage")
	}
}

func TestStrategicBattleResultRequiresOneWinnerAndSnapshotCasualties(t *testing.T) {
	s, err := NewSession(strategicBattleTestSpec())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ValidateResult(Result{Outcome: "none"}); err == nil {
		t.Fatal("zero-winner result unexpectedly accepted")
	}
	if _, err := s.ValidateResult(Result{WinnerSeats: []protocol.SeatID{1, 2}, Outcome: "draw"}); err == nil {
		t.Fatal("multi-winner result unexpectedly accepted")
	}
	if _, err := s.ValidateResult(Result{WinnerSeat: 1, Outcome: "victory", DestroyedShipIDs: []core.ID{999}}); err == nil {
		t.Fatal("out-of-snapshot casualty unexpectedly accepted")
	}
	normalized, err := s.ValidateResult(Result{WinnerSeats: []protocol.SeatID{2}, Outcome: "victory", DestroyedShipIDs: []core.ID{52, 31}})
	if err != nil {
		t.Fatal(err)
	}
	if normalized.WinnerSeat != 2 || !reflect.DeepEqual(normalized.WinnerSeats, []protocol.SeatID{2}) || !reflect.DeepEqual(normalized.DestroyedShipIDs, []core.ID{31, 52}) {
		t.Fatalf("normalized result=%+v", normalized)
	}
	if s.View().Phase != PhaseActive {
		t.Fatal("pure ValidateResult mutated battle phase")
	}
}
