package battle

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestLiveSnapshotActiveTacticalRoundTripAndContinuation(t *testing.T) {
	original := newStartedTacticalSession(t)
	fire1, _ := NewFireBeamCommand(1, FireBeamPayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	submitPrepared(t, original, 1, fire1)

	before, err := json.Marshal(original.LiveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreLiveSnapshot(original.LiveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	after, err := json.Marshal(restored.LiveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("battle live snapshot re-export changed\nbefore=%s\nafter =%s", before, after)
	}

	continueBattle := func(s *Session) {
		endAttacker, _ := NewEndActivationCommand(2, EndActivationPayload{ShipID: 100})
		submitPrepared(t, s, 1, endAttacker)
		endDefender, _ := NewEndActivationCommand(3, EndActivationPayload{ShipID: 200})
		submitPrepared(t, s, 2, endDefender)
		fire2, _ := NewFireBeamCommand(4, FireBeamPayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
		submitPrepared(t, s, 1, fire2)
	}
	continueBattle(original)
	continueBattle(restored)
	left, _ := json.Marshal(original.LiveSnapshot())
	right, _ := json.Marshal(restored.LiveSnapshot())
	if !bytes.Equal(left, right) {
		t.Fatalf("continued battle diverged\noriginal=%s\nrestored=%s", left, right)
	}
	if original.View().Phase != PhaseCompleted {
		t.Fatalf("continued battle phase=%q", original.View().Phase)
	}
}

func TestLiveSnapshotRejectsTacticalCounterCorruption(t *testing.T) {
	s := newStartedTacticalSession(t)
	fire1, _ := NewFireBeamCommand(1, FireBeamPayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	submitPrepared(t, s, 1, fire1)
	snapshot := s.LiveSnapshot()
	snapshot.TacticalRevision++
	if _, err := RestoreLiveSnapshot(snapshot); err == nil {
		t.Fatal("corrupt tactical revision unexpectedly restored")
	}
}
