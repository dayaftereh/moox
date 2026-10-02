package battle

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func checkpoint5FlyingOrdnanceSpec() Spec {
	spec := checkpoint1OrdnanceSpec()
	spec.Tactical.Ships[0].Weapons = spec.Tactical.Ships[0].Weapons[:1]
	spec.Tactical.Ships[0].Weapons[0].Count = 1
	spec.Tactical.Ships[1].X = 30
	spec.Tactical.Ships[1].Y = 10
	return spec
}

func checkpoint5CommitCommand(t *testing.T, s *Session, seatID protocol.SeatID, command protocol.Command) {
	t.Helper()
	prepared, err := s.PrepareCommand(seatID, command)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CommitPreparedCommand(prepared); err != nil {
		t.Fatal(err)
	}
}

func TestSlice177Gate3Checkpoint5ActiveProjectileSnapshotReplayContinuation(t *testing.T) {
	spec := checkpoint5FlyingOrdnanceSpec()
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	fire, err := NewFireOrdnanceCommand(1, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint5CommitCommand(t, s, 1, fire)

	s.mu.Lock()
	if len(s.tactical.ordnance.Projectiles) != 1 {
		s.mu.Unlock()
		t.Fatalf("active projectiles=%+v", s.tactical.ordnance.Projectiles)
	}
	s.tactical.ordnance.Projectiles[0].InterceptDamageRemainder = 2
	s.mu.Unlock()

	before := s.LiveSnapshot()
	if before.Tactical == nil || before.Tactical.Ordnance == nil || len(before.Tactical.Ordnance.Projectiles) != 1 {
		t.Fatalf("active projectile snapshot=%+v", before.Tactical)
	}
	projectile := before.Tactical.Ordnance.Projectiles[0]
	if projectile.ProjectileID != 1 || projectile.X != 22 || projectile.Y != 10 ||
		projectile.Multiplicity != 1 || projectile.UnitDurability != 4 ||
		projectile.InterceptDamageRemainder != 2 || projectile.TargetShipID != 200 {
		t.Fatalf("persisted active projectile=%+v", projectile)
	}

	encoded, err := json.Marshal(before)
	if err != nil {
		t.Fatal(err)
	}
	var decoded LiveSnapshot
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreLiveSnapshot(decoded)
	if err != nil {
		t.Fatal(err)
	}
	restoredEncoded, err := json.Marshal(restored.LiveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, restoredEncoded) {
		t.Fatalf("active ordnance snapshot changed across restore\nbefore=%s\nafter =%s", encoded, restoredEncoded)
	}

	continueBattle := func(target *Session) {
		endAttacker, err := NewEndActivationCommand(2, EndActivationPayload{ShipID: 100})
		if err != nil {
			t.Fatal(err)
		}
		checkpoint5CommitCommand(t, target, 1, endAttacker)
		endDefender, err := NewEndActivationCommand(3, EndActivationPayload{ShipID: 200})
		if err != nil {
			t.Fatal(err)
		}
		checkpoint5CommitCommand(t, target, 2, endDefender)
	}
	continueBattle(s)
	continueBattle(restored)

	left, err := json.Marshal(s.LiveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	right, err := json.Marshal(restored.LiveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		t.Fatalf("ordnance continuation diverged after restore\nleft =%s\nright=%s", left, right)
	}
}

func TestSlice177Gate3Checkpoint5ActiveProjectileObserverPlayerProjectionDetached(t *testing.T) {
	spec := checkpoint5FlyingOrdnanceSpec()
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	fire, err := NewFireOrdnanceCommand(1, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint5CommitCommand(t, s, 1, fire)

	s.mu.Lock()
	s.tactical.ordnance.Projectiles[0].InterceptDamageRemainder = 2
	s.mu.Unlock()

	observer := s.View()
	player, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if observer.Tactical == nil || player.Tactical == nil || len(observer.Tactical.Ordnance.Projectiles) != 1 {
		t.Fatalf("ordnance projection missing observer=%+v player=%+v", observer.Tactical, player.Tactical)
	}
	if player.Tactical.State.RNGState != 0 || player.Spec.Tactical.InitialRNGState != 0 {
		t.Fatalf("player RNG redaction changed: %+v", player.Tactical.State)
	}
	if !reflect.DeepEqual(observer.Tactical.Ordnance, player.Tactical.Ordnance) {
		t.Fatalf("player/observer ordnance differs\nobserver=%+v\nplayer=%+v", observer.Tactical.Ordnance, player.Tactical.Ordnance)
	}

	observer.Tactical.Ordnance.Projectiles[0].X = 999
	observer.Tactical.Ordnance.Projectiles[0].InterceptDamageRemainder = 0
	observer.Tactical.Ordnance.WeaponStates[0].AmmoRemaining = 999
	player.Tactical.Ordnance.Projectiles[0].Y = 999
	fresh := s.View()
	if got := fresh.Tactical.Ordnance.Projectiles[0]; got.X != 22 || got.Y != 10 || got.InterceptDamageRemainder != 2 {
		t.Fatalf("projectile projection aliases authority: %+v", got)
	}
	if got := fresh.Tactical.Ordnance.WeaponStates[0].AmmoRemaining; got != 4 {
		t.Fatalf("weapon-state projection aliases authority: ammo=%d", got)
	}
}

func TestSlice177Gate3Checkpoint5RejectsCorruptProjectileRemainderSnapshot(t *testing.T) {
	spec := checkpoint5FlyingOrdnanceSpec()
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	fire, err := NewFireOrdnanceCommand(1, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint5CommitCommand(t, s, 1, fire)
	snapshot := s.LiveSnapshot()
	projectile := &snapshot.Tactical.Ordnance.Projectiles[0]
	projectile.InterceptDamageRemainder = projectile.UnitDurability
	if _, err := RestoreLiveSnapshot(snapshot); err == nil {
		t.Fatal("corrupt projectile remainder unexpectedly restored")
	}
}

func TestSlice177Gate3Checkpoint5BombardmentOnlyBombRemainsExcludedFromShipTactical(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Ships[0].Weapons = []TacticalWeaponSpec{{
		Slot: 0, WeaponID: "nuclear_bomb", WeaponFamily: "bomb", Count: 1, Ammo: 10,
		FiringArcID: core.ShipWeaponFiringArc360, MinDamage: 3, MaxDamage: 12, TacticalSupport: "bombardment_only",
	}}
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	before := s.View()
	if before.Spec.Tactical == nil || before.Tactical == nil ||
		len(before.Spec.Tactical.Ships[0].Weapons) != 1 || len(before.Tactical.Ships[0].Weapons) != 1 {
		t.Fatalf("bomb metadata projection missing: %+v", before.Tactical)
	}
	specWeapon := before.Spec.Tactical.Ships[0].Weapons[0]
	viewWeapon := before.Tactical.Ships[0].Weapons[0]
	if specWeapon.WeaponID != "nuclear_bomb" || specWeapon.WeaponFamily != "bomb" ||
		specWeapon.TacticalSupport != "bombardment_only" ||
		viewWeapon.WeaponID != "nuclear_bomb" || viewWeapon.TacticalSupport != "bombardment_only" {
		t.Fatalf("bomb metadata changed spec=%+v view=%+v", specWeapon, viewWeapon)
	}
	if len(before.Tactical.Ordnance.Projectiles) != 0 || len(before.Tactical.Ordnance.WeaponStates) != 0 {
		t.Fatalf("bomb unexpectedly materialized ordnance runtime=%+v", before.Tactical.Ordnance)
	}
	for _, action := range before.Tactical.WeaponActions {
		if action.WeaponSlot == 0 && (action.Available || action.CommandKind == CommandFireOrdnance) {
			t.Fatalf("bomb unexpectedly exposed ship Tactical action=%+v", action)
		}
	}

	fire, err := NewFireOrdnanceCommand(1, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.PrepareCommand(1, fire); err == nil {
		t.Fatal("bomb unexpectedly prepared as battle.fire_ordnance")
	}
	after := s.View()
	if !reflect.DeepEqual(before.Tactical.State, after.Tactical.State) ||
		!reflect.DeepEqual(before.Tactical.Ordnance, after.Tactical.Ordnance) ||
		!reflect.DeepEqual(before.Tactical.Events, after.Tactical.Events) {
		t.Fatal("rejected bombardment-only bomb mutated Tactical authority")
	}
}
