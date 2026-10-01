package battle

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
)

func TestSlice176Checkpoint2ArcRejectPreservesState(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Ships[0].Weapons[0] = TacticalWeaponSpec{
		Slot: 0, WeaponID: "phasor", WeaponFamily: "beam", Count: 1,
		MountType: core.ShipWeaponMountStandard, FiringArcID: core.ShipWeaponFiringArcForward,
		MinDamage: 5, MaxDamage: 20, TacticalSupport: "beam",
	}
	spec.Tactical.Ships[1].X = 9
	spec.Tactical.Ships[1].Y = 10

	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	before := s.View()
	cmd := mustFireCommand(t, 1, 100, 200, 0)
	if _, err := s.PrepareCommand(1, cmd); err == nil {
		t.Fatal("rear target unexpectedly accepted by forward Beam arc")
	}
	after := s.View()
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("arc-rejected Beam mutated battle before=%+v after=%+v", before.Tactical.State, after.Tactical.State)
	}
	if len(after.Tactical.LegalFireActions) != 0 {
		t.Fatalf("arc-rejected target leaked into legal actions: %+v", after.Tactical.LegalFireActions)
	}
}

func TestSlice176Checkpoint2CountAutofireResolvesSixOrderedShots(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Ships[0].Weapons[0] = TacticalWeaponSpec{
		Slot: 0, WeaponID: "laser_cannon", WeaponFamily: "beam", Count: 2,
		MountType:   core.ShipWeaponMountStandard,
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
	before := s.View()
	submitPrepared(t, s, 1, mustFireCommand(t, 1, 100, 200, 0))
	after := s.View()

	type shotEvent struct {
		ShotIndex int `json:"shot_index"`
		ShotCount int `json:"shot_count"`
	}
	var shots []shotEvent
	for _, event := range after.Tactical.Events {
		if event.Kind != "beam_fired" || event.CommandSequence != 1 {
			continue
		}
		var shot shotEvent
		if err := json.Unmarshal(event.Data, &shot); err != nil {
			t.Fatal(err)
		}
		shots = append(shots, shot)
	}
	if len(shots) != 6 {
		t.Fatalf("beam_fired count=%d want 6 events=%+v", len(shots), shots)
	}
	for i, shot := range shots {
		if shot.ShotIndex != i+1 || shot.ShotCount != 6 {
			t.Fatalf("shot[%d]=%+v want index=%d count=6", i, shot, i+1)
		}
	}
	attacker := tacticalStateForTest(t, after.Tactical.State.Ships, 100)
	if len(attacker.Weapons) != 1 || attacker.Weapons[0].Ready {
		t.Fatalf("accepted Beam battery did not consume readiness once: %+v", attacker.Weapons)
	}
	if after.Tactical.State.RNGState == before.Tactical.State.RNGState {
		t.Fatal("accepted six-shot Beam battery did not consume RNG")
	}
}
