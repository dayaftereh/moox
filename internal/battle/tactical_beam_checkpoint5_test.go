package battle

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"moox/internal/core"
)

type checkpoint5BeamEvent struct {
	RangeIndex          int                         `json:"range_index"`
	RawHitRoll          int                         `json:"raw_hit_roll"`
	EffectiveHitRoll    int                         `json:"effective_hit_roll"`
	Threshold           int                         `json:"threshold"`
	Damage              int                         `json:"damage"`
	DamagePercent       int                         `json:"damage_percent"`
	ScaledMinDamage     int                         `json:"scaled_min_damage"`
	ScaledMaxDamage     int                         `json:"scaled_max_damage"`
	PhysicalWeaponIndex int                         `json:"physical_weapon_index"`
	PhysicalWeaponCount int                         `json:"physical_weapon_count"`
	BurstShotIndex      int                         `json:"burst_shot_index"`
	BurstShotCount      int                         `json:"burst_shot_count"`
	ShotIndex           int                         `json:"shot_index"`
	ShotCount           int                         `json:"shot_count"`
	WeaponID            string                      `json:"weapon_id"`
	MountType           core.ShipWeaponMountType    `json:"mount_type"`
	ModifierIDs         []core.ShipWeaponModifierID `json:"modifier_ids"`
}

func checkpoint5SetNextRaw(t *testing.T, s *Session, want int) uint32 {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	for seed := uint32(1); seed < 1_000_000; seed++ {
		probe := cloneTacticalRuntime(*s.tactical)
		probe.state.RNGState = seed
		raw, err := probe.random(100, s.spec.Tactical.Rules)
		if err == nil && raw == want {
			s.tactical.state.RNGState = seed
			return seed
		}
	}
	t.Fatalf("failed to find Tactical RNG seed for raw=%d", want)
	return 0
}

func checkpoint5VectorFire(t *testing.T, weapon TacticalWeaponSpec, targetX, raw int, withShield bool) View {
	t.Helper()
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Ships[0].BeamOffense = 25
	weapon.Slot = 0
	if weapon.Count == 0 {
		weapon.Count = 1
	}
	weapon.WeaponFamily = "beam"
	weapon.TacticalSupport = "beam"
	if weapon.MountType == "" {
		weapon.MountType = core.ShipWeaponMountStandard
	}
	if weapon.FiringArcID == "" {
		weapon.FiringArcID = core.ShipWeaponFiringArc360
	}
	spec.Tactical.Ships[0].Weapons = []TacticalWeaponSpec{weapon}
	target := checkpoint3DoomStarTarget(withShield)
	target.X = targetX
	target.CurrentCombatSpeed = 5
	target.BeamDefense = 25
	spec.Tactical.Ships[1] = target

	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	checkpoint5SetNextRaw(t, s, raw)
	submitPrepared(t, s, 1, mustFireCommand(t, 1, 100, 200, 0))
	return s.View()
}

func checkpoint5BeamEvents(t *testing.T, view View) []checkpoint5BeamEvent {
	t.Helper()
	var out []checkpoint5BeamEvent
	for _, event := range view.Tactical.Events {
		if event.Kind != "beam_fired" || event.CommandSequence != 1 {
			continue
		}
		var decoded checkpoint5BeamEvent
		if err := json.Unmarshal(event.Data, &decoded); err != nil {
			t.Fatal(err)
		}
		out = append(out, decoded)
	}
	return out
}

func TestSlice176Checkpoint5GoldenVectorAStandardLaserRounding(t *testing.T) {
	view := checkpoint5VectorFire(t, TacticalWeaponSpec{
		WeaponID: "laser_cannon", MinDamage: 1, MaxDamage: 4,
	}, 16, 100, false)
	events := checkpoint5BeamEvents(t, view)
	if len(events) != 1 {
		t.Fatalf("vector A shot count=%d want 1", len(events))
	}
	got := events[0]
	if got.RangeIndex != 2 || got.RawHitRoll != 100 || got.EffectiveHitRoll != 100 ||
		got.Threshold != 50 || got.DamagePercent != 90 || got.ScaledMinDamage != 1 ||
		got.ScaledMaxDamage != 4 || got.Damage != 4 {
		t.Fatalf("vector A=%+v", got)
	}
}

func TestSlice176Checkpoint5GoldenVectorBParticleHeavy(t *testing.T) {
	view := checkpoint5VectorFire(t, TacticalWeaponSpec{
		WeaponID: "particle_beam", MinDamage: 10, MaxDamage: 30,
		MountType:       core.ShipWeaponMountHeavy,
		InnateEffectIDs: []string{"natural_shield_piercing"},
	}, 22, 80, true)
	events := checkpoint5BeamEvents(t, view)
	if len(events) != 1 {
		t.Fatalf("vector B shot count=%d want 1", len(events))
	}
	got := events[0]
	if got.RangeIndex != 2 || got.RawHitRoll != 80 || got.EffectiveHitRoll != 80 ||
		got.Threshold != 50 || got.DamagePercent != 140 || got.ScaledMinDamage != 14 ||
		got.ScaledMaxDamage != 42 || got.Damage != 31 {
		t.Fatalf("vector B=%+v", got)
	}
	target := tacticalStateForTest(t, view.Tactical.State.Ships, 200)
	if target.ShieldArcCurrent != [4]int{35, 35, 35, 35} || target.ArmorCurrent != 119 {
		t.Fatalf("vector B natural Shield Piercing target=%+v", target)
	}
}

func TestSlice176Checkpoint5GoldenVectorCMaulerHeavy(t *testing.T) {
	view := checkpoint5VectorFire(t, TacticalWeaponSpec{
		WeaponID: "mauler_device", MinDamage: 100, MaxDamage: 100,
		MountType:       core.ShipWeaponMountHeavy,
		InnateEffectIDs: []string{"always_hit", "double_range_dissipation"},
	}, 22, 10, false)
	events := checkpoint5BeamEvents(t, view)
	if len(events) != 1 {
		t.Fatalf("vector C shot count=%d want 1", len(events))
	}
	got := events[0]
	if got.RangeIndex != 2 || got.RawHitRoll != 10 || got.EffectiveHitRoll != 100 ||
		got.Threshold != 60 || got.DamagePercent != 140 || got.Damage != 140 {
		t.Fatalf("vector C=%+v", got)
	}
}

func TestSlice176Checkpoint5GoldenVectorDPointDefenseShipTarget(t *testing.T) {
	view := checkpoint5VectorFire(t, TacticalWeaponSpec{
		WeaponID: "laser_cannon", MinDamage: 1, MaxDamage: 4,
		MountType: core.ShipWeaponMountPointDefense,
	}, 16, 50, false)
	events := checkpoint5BeamEvents(t, view)
	if len(events) != 1 {
		t.Fatalf("vector D shot count=%d want 1", len(events))
	}
	got := events[0]
	// Runtime records the range-only effective threshold (70) and the PD +25
	// in EffectiveHitRoll. The frozen vector's raw-roll threshold is therefore
	// 70 - 25 = 45.
	rawThreshold := got.Threshold - (got.EffectiveHitRoll - got.RawHitRoll)
	if got.RangeIndex != 4 || got.RawHitRoll != 50 || got.EffectiveHitRoll != 75 ||
		rawThreshold != 45 || got.DamagePercent != 20 || got.ScaledMinDamage != 1 ||
		got.ScaledMaxDamage != 1 || got.Damage != 1 {
		t.Fatalf("vector D=%+v rawThreshold=%d", got, rawThreshold)
	}
}

func TestSlice176Checkpoint5GoldenVectorECountTwoAutofireOrdering(t *testing.T) {
	view := checkpoint5VectorFire(t, TacticalWeaponSpec{
		WeaponID: "laser_cannon", Count: 2, MinDamage: 1, MaxDamage: 4,
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierAutofire},
	}, 11, 50, false)
	events := checkpoint5BeamEvents(t, view)
	if len(events) != 6 {
		t.Fatalf("vector E shot count=%d want 6", len(events))
	}
	for i, got := range events {
		if got.PhysicalWeaponIndex != i/3+1 || got.PhysicalWeaponCount != 2 ||
			got.BurstShotIndex != i%3+1 || got.BurstShotCount != 3 ||
			got.ShotIndex != i+1 || got.ShotCount != 6 {
			t.Fatalf("vector E shot[%d]=%+v", i, got)
		}
		if got.RawHitRoll <= 95 {
			if got.EffectiveHitRoll != got.RawHitRoll-20 {
				t.Fatalf("vector E shot[%d] raw/effective=%d/%d want -20 differential", i, got.RawHitRoll, got.EffectiveHitRoll)
			}
		} else if got.EffectiveHitRoll != 100 {
			t.Fatalf("vector E shot[%d] high raw effective=%d want 100", i, got.EffectiveHitRoll)
		}
	}
}

func checkpoint5PersistenceSpec() Spec {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Ships[0].BeamOffense = 25
	spec.Tactical.Ships[0].Weapons = []TacticalWeaponSpec{
		{
			Slot: 0, WeaponID: "neutron_blaster", WeaponFamily: "beam", Count: 1,
			MountType: core.ShipWeaponMountStandard, FiringArcID: core.ShipWeaponFiringArc360,
			MinDamage: 3, MaxDamage: 12, InnateEffectIDs: []string{"marine_kill"}, TacticalSupport: "beam",
		},
		{
			Slot: 1, WeaponID: "phasor", WeaponFamily: "beam", Count: 3,
			MountType: core.ShipWeaponMountPointDefense, FiringArcID: core.ShipWeaponFiringArc360,
			ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierShieldPiercing, core.ShipWeaponModifierAutofire},
			MinDamage:   5, MaxDamage: 20, TacticalSupport: "beam",
		},
	}
	target := checkpoint3DoomStarTarget(false)
	target.CurrentCombatSpeed = 5
	target.BeamDefense = 25
	spec.Tactical.Ships[1] = target
	return spec
}

func TestSlice176Checkpoint5ConfiguredBeamLiveSnapshotReplayPreservesMarineRNGAndEvents(t *testing.T) {
	s, err := NewSession(checkpoint5PersistenceSpec())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	checkpoint5SetNextRaw(t, s, 100)
	submitPrepared(t, s, 1, mustFireCommand(t, 1, 100, 200, 0))

	beforeView := s.View()
	target := tacticalStateForTest(t, beforeView.Tactical.State.Ships, 200)
	if target.MarineCurrent != 48 {
		t.Fatalf("pre-save MarineCurrent=%d want 48", target.MarineCurrent)
	}
	before, err := json.Marshal(s.LiveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreLiveSnapshot(s.LiveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	after, err := json.Marshal(restored.LiveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("configured Beam live snapshot changed across restore\nbefore=%s\nafter =%s", before, after)
	}

	restoredView := restored.View()
	gotWeapon := restoredView.Spec.Tactical.Ships[0].Weapons[1]
	if gotWeapon.WeaponID != "phasor" || gotWeapon.MountType != core.ShipWeaponMountPointDefense ||
		gotWeapon.FiringArcID != core.ShipWeaponFiringArc360 ||
		!reflect.DeepEqual(gotWeapon.ModifierIDs, []core.ShipWeaponModifierID{core.ShipWeaponModifierShieldPiercing, core.ShipWeaponModifierAutofire}) ||
		gotWeapon.TacticalSupport != "beam" {
		t.Fatalf("restored configured Beam=%+v", gotWeapon)
	}
	if tacticalStateForTest(t, restoredView.Tactical.State.Ships, 200).MarineCurrent != 48 ||
		restoredView.Tactical.State.RNGState != beforeView.Tactical.State.RNGState ||
		!reflect.DeepEqual(restoredView.Tactical.Events, beforeView.Tactical.Events) {
		t.Fatalf("restored mutable Beam state diverged")
	}

	continueBattle := func(target *Session) {
		endA, _ := NewEndActivationCommand(2, EndActivationPayload{ShipID: 100})
		submitPrepared(t, target, 1, endA)
		endD, _ := NewEndActivationCommand(3, EndActivationPayload{ShipID: 200})
		submitPrepared(t, target, 2, endD)
		fire2, _ := NewFireBeamCommand(4, FireBeamPayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
		submitPrepared(t, target, 1, fire2)
	}
	continueBattle(s)
	continueBattle(restored)
	left, _ := json.Marshal(s.LiveSnapshot())
	right, _ := json.Marshal(restored.LiveSnapshot())
	if !bytes.Equal(left, right) {
		t.Fatalf("configured Beam replay diverged after restore\nleft =%s\nright=%s", left, right)
	}
}

func TestSlice176Checkpoint5LiveSnapshotRejectsMarineCurrentCorruption(t *testing.T) {
	s, err := NewSession(checkpoint5PersistenceSpec())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		value int
	}{
		{name: "negative", value: -1},
		{name: "above max", value: 51},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := s.LiveSnapshot()
			snapshot.Tactical.State.Ships[1].MarineCurrent = tc.value
			_, err := RestoreLiveSnapshot(snapshot)
			if err == nil || !strings.Contains(err.Error(), "MarineCurrent") {
				t.Fatalf("MarineCurrent=%d restore error=%v", tc.value, err)
			}
		})
	}
}

func TestSlice176Checkpoint5ObserverPlayerProjectionIsIsolatedForBeamAndMarineFacts(t *testing.T) {
	s, err := NewSession(checkpoint5PersistenceSpec())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	observer := s.View()
	if observer.Tactical == nil || len(observer.Tactical.WeaponActions) != 2 {
		t.Fatalf("observer Tactical=%+v", observer.Tactical)
	}
	if observer.Tactical.Ships[1].MarineCurrent != 50 || observer.Tactical.Ships[1].MarineMax != 50 {
		t.Fatalf("observer marine facts=%+v", observer.Tactical.Ships[1])
	}
	weapon := observer.Tactical.Ships[0].Weapons[1]
	if weapon.WeaponID != "phasor" || weapon.MountType != core.ShipWeaponMountPointDefense ||
		!reflect.DeepEqual(weapon.ModifierIDs, []core.ShipWeaponModifierID{core.ShipWeaponModifierShieldPiercing, core.ShipWeaponModifierAutofire}) {
		t.Fatalf("observer configured Beam=%+v", weapon)
	}

	active, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if active.Tactical == nil || active.Tactical.State.RNGState != 0 || len(active.Tactical.WeaponActions) != 2 {
		t.Fatalf("active player projection=%+v", active.Tactical)
	}
	inactive, err := s.PlayerView(2)
	if err != nil {
		t.Fatal(err)
	}
	if inactive.Tactical == nil || inactive.Tactical.State.RNGState != 0 || len(inactive.Tactical.WeaponActions) != 0 {
		t.Fatalf("inactive player projection=%+v", inactive.Tactical)
	}

	observer.Spec.Tactical.Ships[0].Weapons[1].ModifierIDs[0] = core.ShipWeaponModifierArmorPiercing
	observer.Tactical.Ships[0].Weapons[1].ModifierIDs[0] = core.ShipWeaponModifierArmorPiercing
	observer.Tactical.State.Ships[1].MarineCurrent = 0
	fresh := s.View()
	if fresh.Spec.Tactical.Ships[0].Weapons[1].ModifierIDs[0] != core.ShipWeaponModifierShieldPiercing ||
		fresh.Tactical.Ships[0].Weapons[1].ModifierIDs[0] != core.ShipWeaponModifierShieldPiercing ||
		fresh.Tactical.State.Ships[1].MarineCurrent != 50 {
		t.Fatalf("observer/public mutation leaked into authority")
	}
}
