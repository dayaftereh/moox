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

func TestSlice178Gate3Checkpoint1MigratesSchema1ToCompatibilityDisabledV2(t *testing.T) {
	original := newStartedTacticalSession(t)
	snapshot := original.LiveSnapshot()
	snapshot.SchemaVersion = legacyLiveSnapshotSchemaVersion
	snapshot.Tactical.NextMovingActorOrder = 0
	snapshot.Tactical.Legacy17_8Disabled = false

	restored, err := RestoreLiveSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	got := restored.LiveSnapshot()
	if got.SchemaVersion != LiveSnapshotSchemaVersion {
		t.Fatalf("schema=%d want=%d", got.SchemaVersion, LiveSnapshotSchemaVersion)
	}
	if got.Tactical == nil || !got.Tactical.Legacy17_8Disabled {
		t.Fatalf("legacy migration did not preserve compatibility-disabled state: %+v", got.Tactical)
	}
	if got.Tactical.NextMovingActorOrder != 1 {
		t.Fatalf("next moving-actor order=%d want=1", got.Tactical.NextMovingActorOrder)
	}
}

func TestSlice178Gate3Checkpoint1MigratesLegacyProjectileSpawnOrder(t *testing.T) {
	snapshot := LiveSnapshot{
		SchemaVersion: legacyLiveSnapshotSchemaVersion,
		Tactical: &TacticalLiveSnapshot{
			Ordnance: &TacticalOrdnanceState{
				Projectiles: []TacticalProjectileState{
					{ProjectileID: 2},
					{ProjectileID: 5},
				},
			},
		},
	}
	got, err := migrateLiveSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != LiveSnapshotSchemaVersion || got.Tactical == nil || !got.Tactical.Legacy17_8Disabled {
		t.Fatalf("legacy migration header=%+v", got)
	}
	if got.Tactical.NextMovingActorOrder != 6 {
		t.Fatalf("next moving-actor order=%d want=6", got.Tactical.NextMovingActorOrder)
	}
	if got.Tactical.Ordnance.Projectiles[0].SpawnOrder != 2 || got.Tactical.Ordnance.Projectiles[1].SpawnOrder != 5 {
		t.Fatalf("spawn orders=%d/%d want=2/5", got.Tactical.Ordnance.Projectiles[0].SpawnOrder, got.Tactical.Ordnance.Projectiles[1].SpawnOrder)
	}
	if snapshot.Tactical.Ordnance.Projectiles[0].SpawnOrder != 0 || snapshot.Tactical.Ordnance.Projectiles[1].SpawnOrder != 0 {
		t.Fatal("legacy migration mutated caller-owned snapshot")
	}
}

func TestSlice178Gate3Checkpoint1RejectsMissingV2MovingActorOrder(t *testing.T) {
	s := newStartedTacticalSession(t)
	snapshot := s.LiveSnapshot()
	snapshot.Tactical.NextMovingActorOrder = 0
	if _, err := RestoreLiveSnapshot(snapshot); err == nil {
		t.Fatal("schema-v2 snapshot with zero moving-actor order unexpectedly restored")
	}
}
