package battle

import (
	"testing"

	"moox/internal/core"
)

func checkpoint4MixedWeaponSpec() Spec {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Ships[0].Weapons = []TacticalWeaponSpec{
		{Slot: 0, WeaponID: "laser_cannon", WeaponFamily: "beam", Count: 1, MinDamage: 1, MaxDamage: 4, TacticalSupport: "baseline"},
		{Slot: 1, WeaponID: "nuclear_missile", WeaponFamily: "missile", Count: 1, Ammo: 5, FiringArcID: core.ShipWeaponFiringArc360, MinDamage: 8, MaxDamage: 8, TacticalSupport: "metadata_only"},
	}
	return spec
}

func checkpoint4ActionBySlot(t *testing.T, actions []TacticalWeaponActionAvailability, slot int) TacticalWeaponActionAvailability {
	t.Helper()
	for _, action := range actions {
		if action.WeaponSlot == slot {
			return action
		}
	}
	t.Fatalf("weapon action slot %d missing: %+v", slot, actions)
	return TacticalWeaponActionAvailability{}
}

func TestSlice175Checkpoint4ProjectsOneAuthoritativeActionPerInstalledSlot(t *testing.T) {
	spec := checkpoint4MixedWeaponSpec()
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	view := s.View()
	if view.Tactical == nil {
		t.Fatal("tactical view missing")
	}
	if len(view.Tactical.WeaponActions) != 2 {
		t.Fatalf("weapon action count=%d want=2: %+v", len(view.Tactical.WeaponActions), view.Tactical.WeaponActions)
	}

	laser := checkpoint4ActionBySlot(t, view.Tactical.WeaponActions, 0)
	if laser.ShipID != 100 || laser.WeaponID != "laser_cannon" || laser.WeaponFamily != "beam" || laser.CommandKind != CommandFireBeam || !laser.Available || laser.UnavailableReason != "" || len(laser.Targets) != 1 || laser.Targets[0].TargetShipID != 200 {
		t.Fatalf("Laser availability=%+v", laser)
	}

	missile := checkpoint4ActionBySlot(t, view.Tactical.WeaponActions, 1)
	if missile.ShipID != 100 || missile.WeaponID != "nuclear_missile" || missile.WeaponFamily != "missile" || missile.CommandKind != "" || missile.Available || missile.UnavailableReason != TacticalWeaponUnavailableConsumer || len(missile.Targets) != 0 {
		t.Fatalf("missile availability=%+v", missile)
	}

	if len(view.Tactical.LegalFireActions) != 1 || view.Tactical.LegalFireActions[0].WeaponSlot != 0 || view.Tactical.LegalFireActions[0].WeaponID != "laser_cannon" {
		t.Fatalf("legacy fire compatibility projection=%+v", view.Tactical.LegalFireActions)
	}

	attacker, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if attacker.Tactical == nil || len(attacker.Tactical.WeaponActions) != 2 {
		t.Fatalf("controlling player generic actions=%+v", attacker.Tactical)
	}
	defender, err := s.PlayerView(2)
	if err != nil {
		t.Fatal(err)
	}
	if defender.Tactical == nil || len(defender.Tactical.WeaponActions) != 0 {
		t.Fatalf("non-controlling player leaked weapon actions=%+v", defender.Tactical)
	}
}

func TestSlice175Checkpoint4StableUnavailableReasons(t *testing.T) {
	spec := checkpoint4MixedWeaponSpec()
	runtime, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}

	actions := tacticalWeaponActions(*spec.Tactical, &runtime)
	if got := checkpoint4ActionBySlot(t, actions, 1).UnavailableReason; got != TacticalWeaponUnavailableConsumer {
		t.Fatalf("unsupported reason=%q", got)
	}

	laserState := weaponState(runtime.shipState(100), 0)
	if laserState == nil {
		t.Fatal("Laser state missing")
	}
	laserState.Ready = false
	actions = tacticalWeaponActions(*spec.Tactical, &runtime)
	notReady := checkpoint4ActionBySlot(t, actions, 0)
	if notReady.Available || notReady.UnavailableReason != TacticalWeaponUnavailableNotReady || len(notReady.Targets) != 0 {
		t.Fatalf("not-ready action=%+v", notReady)
	}

	laserState.Ready = true
	target := runtime.shipState(200)
	if target == nil {
		t.Fatal("target state missing")
	}
	target.X = 1_000
	actions = tacticalWeaponActions(*spec.Tactical, &runtime)
	noTargets := checkpoint4ActionBySlot(t, actions, 0)
	if noTargets.Available || noTargets.UnavailableReason != TacticalWeaponUnavailableNoTargets || len(noTargets.Targets) != 0 {
		t.Fatalf("no-target action=%+v", noTargets)
	}
}

func TestSlice175Checkpoint4RejectedUnsupportedActionConsumesNoRNGOrReadiness(t *testing.T) {
	spec := checkpoint4MixedWeaponSpec()
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	before := s.View()
	if before.Tactical == nil {
		t.Fatal("tactical view missing")
	}
	beforeLaser := weaponState(&before.Tactical.State.Ships[0], 0)
	beforeMissile := weaponState(&before.Tactical.State.Ships[0], 1)
	if beforeLaser == nil || beforeMissile == nil {
		t.Fatal("weapon state missing before rejection")
	}
	cmd, err := NewFireBeamCommand(1, FireBeamPayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.PrepareCommand(1, cmd); err == nil {
		t.Fatal("unsupported missile unexpectedly prepared as Beam command")
	}

	after := s.View()
	if after.Tactical == nil {
		t.Fatal("tactical view missing after rejection")
	}
	afterLaser := weaponState(&after.Tactical.State.Ships[0], 0)
	afterMissile := weaponState(&after.Tactical.State.Ships[0], 1)
	if after.Tactical.State.RNGState != before.Tactical.State.RNGState || after.Tactical.State.NextCommandSequence != before.Tactical.State.NextCommandSequence || len(after.Tactical.Events) != len(before.Tactical.Events) {
		t.Fatalf("rejected action mutated RNG/sequence/events: before=%+v after=%+v", before.Tactical.State, after.Tactical.State)
	}
	if afterLaser == nil || afterMissile == nil || afterLaser.Ready != beforeLaser.Ready || afterMissile.Ready != beforeMissile.Ready {
		t.Fatalf("rejected action mutated readiness: before=%+v after=%+v", before.Tactical.State.Ships[0].Weapons, after.Tactical.State.Ships[0].Weapons)
	}
}
