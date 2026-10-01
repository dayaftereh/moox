package battle

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
)

type checkpoint4BeamIdentityEvent struct {
	ShipID              core.ID                     `json:"ship_id"`
	SourceShipID        core.ID                     `json:"source_ship_id"`
	TargetShipID        core.ID                     `json:"target_ship_id"`
	WeaponSlot          int                         `json:"weapon_slot"`
	WeaponID            string                      `json:"weapon_id"`
	MountType           core.ShipWeaponMountType    `json:"mount_type"`
	ModifierIDs         []core.ShipWeaponModifierID `json:"modifier_ids"`
	PhysicalWeaponIndex int                         `json:"physical_weapon_index"`
	PhysicalWeaponCount int                         `json:"physical_weapon_count"`
	BurstShotIndex      int                         `json:"burst_shot_index"`
	BurstShotCount      int                         `json:"burst_shot_count"`
	ShotIndex           int                         `json:"shot_index"`
	ShotCount           int                         `json:"shot_count"`
	ApplicationIndex    int                         `json:"application_index"`
	ApplicationCount    int                         `json:"application_count"`
	ShieldArc           int                         `json:"shield_arc"`
	ReactivePD          bool                        `json:"reactive_point_defense"`
	TargetX             int                         `json:"target_x"`
	TargetY             int                         `json:"target_y"`
}

func checkpoint4FireWeapon(t *testing.T, weapon TacticalWeaponSpec, target TacticalShipSpec) View {
	t.Helper()
	spec := baselineTacticalBattleSpec()
	weapon.Slot = 0
	weapon.WeaponFamily = "beam"
	weapon.TacticalSupport = "beam"
	if weapon.FiringArcID == "" {
		weapon.FiringArcID = core.ShipWeaponFiringArc360
	}
	if weapon.MountType == "" {
		weapon.MountType = core.ShipWeaponMountStandard
	}
	spec.Tactical.Ships[0].Weapons[0] = weapon
	spec.Tactical.Ships[1] = target
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	submitPrepared(t, s, 1, mustFireCommand(t, 1, 100, target.ShipID, 0))
	return s.View()
}

func checkpoint4Events(t *testing.T, view View, kind string) []checkpoint4BeamIdentityEvent {
	t.Helper()
	var out []checkpoint4BeamIdentityEvent
	for _, event := range view.Tactical.Events {
		if event.Kind != kind {
			continue
		}
		var decoded checkpoint4BeamIdentityEvent
		if err := json.Unmarshal(event.Data, &decoded); err != nil {
			t.Fatal(err)
		}
		out = append(out, decoded)
	}
	return out
}

func TestSlice176Checkpoint4BeamEventsCarryFrozenShotIdentity(t *testing.T) {
	view := checkpoint4FireWeapon(t, TacticalWeaponSpec{
		WeaponID: "laser_cannon", Count: 2, MinDamage: 1, MaxDamage: 4,
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierAutofire},
	}, checkpoint3DoomStarTarget(false))

	shots := checkpoint4Events(t, view, "beam_fired")
	if len(shots) != 6 {
		t.Fatalf("beam_fired count=%d want 6", len(shots))
	}
	for i, shot := range shots {
		wantPhysical := i/3 + 1
		wantBurst := i%3 + 1
		if shot.ShipID != 100 || shot.TargetShipID != 200 || shot.WeaponSlot != 0 ||
			shot.WeaponID != "laser_cannon" || shot.MountType != core.ShipWeaponMountStandard ||
			!reflect.DeepEqual(shot.ModifierIDs, []core.ShipWeaponModifierID{core.ShipWeaponModifierAutofire}) ||
			shot.PhysicalWeaponIndex != wantPhysical || shot.PhysicalWeaponCount != 2 ||
			shot.BurstShotIndex != wantBurst || shot.BurstShotCount != 3 ||
			shot.ShotIndex != i+1 || shot.ShotCount != 6 {
			t.Fatalf("shot[%d]=%+v", i, shot)
		}
	}
	volleys := checkpoint4Events(t, view, "beam_volley_resolved")
	if len(volleys) != 1 {
		t.Fatalf("volley count=%d want 1", len(volleys))
	}
	volley := volleys[0]
	if volley.ShipID != 100 || volley.TargetShipID != 200 || volley.WeaponID != "laser_cannon" ||
		volley.PhysicalWeaponCount != 2 || volley.BurstShotCount != 3 || volley.ShotCount != 6 {
		t.Fatalf("volley identity=%+v", volley)
	}
}

func TestSlice176Checkpoint4EnvelopingDamageEventsCarryPerFacingIdentity(t *testing.T) {
	view := checkpoint4FireWeapon(t, TacticalWeaponSpec{
		WeaponID: "plasma_cannon", Count: 1, MinDamage: 6, MaxDamage: 30,
		InnateEffectIDs: []string{"enveloping", "doubled_range_to_hit"},
	}, checkpoint3DoomStarTarget(true))

	damage := checkpoint4Events(t, view, "battle_damage_applied")
	if len(damage) != 4 {
		t.Fatalf("enveloping damage event count=%d want 4", len(damage))
	}
	for i, event := range damage {
		if event.SourceShipID != 100 || event.TargetShipID != 200 || event.WeaponID != "plasma_cannon" ||
			event.ApplicationIndex != i+1 || event.ApplicationCount != 4 || event.ShieldArc != i ||
			event.PhysicalWeaponIndex != 1 || event.PhysicalWeaponCount != 1 ||
			event.BurstShotIndex != 1 || event.BurstShotCount != 1 ||
			event.ShotIndex != 1 || event.ShotCount != 1 {
			t.Fatalf("damage[%d]=%+v", i, event)
		}
	}
}

func TestSlice176Checkpoint4ReactivePointDefensePreservesReadinessAndUsesActorCoordinates(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Ships[0].Weapons[0] = TacticalWeaponSpec{
		Slot: 0, WeaponID: "laser_cannon", WeaponFamily: "beam", Count: 2,
		MountType:   core.ShipWeaponMountPointDefense,
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierAutofire},
		FiringArcID: core.ShipWeaponFiringArc360,
		MinDamage:   1, MaxDamage: 4, TacticalSupport: "beam",
	}
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	s.mu.Lock()
	source := s.tactical.shipState(100)
	source.Weapons[0].Ready = false
	beforeRNG := s.tactical.state.RNGState
	beforeSequence := s.tactical.state.NextCommandSequence
	beforeEvents := len(s.tactical.events)
	s.mu.Unlock()

	result, err := s.ResolveReactiveBeamPointDefense(ReactiveBeamPointDefenseRequest{
		ShipID: 100, WeaponSlot: 0, TargetX: 11, TargetY: 10, TargetBeamDefense: 25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.WeaponID != "laser_cannon" || result.MountType != core.ShipWeaponMountPointDefense ||
		!reflect.DeepEqual(result.ModifierIDs, []core.ShipWeaponModifierID{core.ShipWeaponModifierAutofire}) ||
		result.TargetX != 11 || result.TargetY != 10 || len(result.Shots) != 6 {
		t.Fatalf("reactive result=%+v", result)
	}
	for i, shot := range result.Shots {
		if shot.PhysicalWeaponIndex != i/3+1 || shot.PhysicalWeaponCount != 2 ||
			shot.BurstShotIndex != i%3+1 || shot.BurstShotCount != 3 ||
			shot.ShotIndex != i+1 || shot.ShotCount != 6 {
			t.Fatalf("reactive shot[%d]=%+v", i, shot)
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	afterSource := s.tactical.shipState(100)
	if afterSource.Weapons[0].Ready {
		t.Fatal("reactive PD changed normal readiness false -> true")
	}
	if s.tactical.state.RNGState == beforeRNG {
		t.Fatal("reactive PD did not consume Beam hit RNG")
	}
	if s.tactical.state.NextCommandSequence != beforeSequence {
		t.Fatalf("reactive PD changed player command sequence %d -> %d", beforeSequence, s.tactical.state.NextCommandSequence)
	}
	newEvents := s.tactical.events[beforeEvents:]
	if len(newEvents) != 7 {
		t.Fatalf("reactive events=%d want 7", len(newEvents))
	}
	for i, event := range newEvents[:6] {
		if event.Kind != "beam_fired" || event.CommandSequence != 0 {
			t.Fatalf("reactive event[%d]=%+v", i, event)
		}
		var decoded checkpoint4BeamIdentityEvent
		if err := json.Unmarshal(event.Data, &decoded); err != nil {
			t.Fatal(err)
		}
		if !decoded.ReactivePD || decoded.TargetShipID != 0 || decoded.TargetX != 11 || decoded.TargetY != 10 ||
			decoded.WeaponID != "laser_cannon" || decoded.MountType != core.ShipWeaponMountPointDefense {
			t.Fatalf("reactive event[%d] identity=%+v", i, decoded)
		}
	}
	if newEvents[6].Kind != "beam_volley_resolved" {
		t.Fatalf("reactive aggregate=%+v", newEvents[6])
	}
}

func TestSlice176Checkpoint4ReactivePointDefenseRejectsNormalBeamWithoutMutation(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	before := s.View()
	_, err = s.ResolveReactiveBeamPointDefense(ReactiveBeamPointDefenseRequest{
		ShipID: 100, WeaponSlot: 0, TargetX: 11, TargetY: 10,
	})
	if err == nil {
		t.Fatal("standard Beam unexpectedly accepted as reactive PD")
	}
	after := s.View()
	if !reflect.DeepEqual(before.Tactical.State, after.Tactical.State) || !reflect.DeepEqual(before.Tactical.Events, after.Tactical.Events) {
		t.Fatalf("rejected reactive PD mutated state/events")
	}
}
