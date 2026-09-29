package battle

import (
	"encoding/json"
	"testing"

	"moox/internal/core"
)

func checkpoint3ShieldBattleSpec(shieldID string, strength, arcMax int) Spec {
	spec := baselineTacticalBattleSpec()
	defender := &spec.Tactical.Ships[1]
	defender.ShieldID = shieldID
	defender.ShieldStrength = strength
	defender.ShieldArcMax = arcMax
	return spec
}

func checkpoint3ShipView(t *testing.T, view View, id core.ID) TacticalShipView {
	t.Helper()
	if view.Tactical == nil {
		t.Fatal("tactical view missing")
	}
	for _, ship := range view.Tactical.Ships {
		if ship.ShipID == id {
			return ship
		}
	}
	t.Fatalf("tactical ship view %d missing", id)
	return TacticalShipView{}
}

func TestSlice175Checkpoint3ShieldClassesInitializeFourFacingPoolsAndView(t *testing.T) {
	tests := []struct {
		name     string
		shieldID string
		strength int
		arcMax   int
	}{
		{name: "none", strength: 0, arcMax: 0},
		{name: "class I", shieldID: "class_i_shield", strength: 1, arcMax: 5},
		{name: "class III", shieldID: "class_iii_shield", strength: 3, arcMax: 15},
		{name: "class V", shieldID: "class_v_shield", strength: 5, arcMax: 25},
		{name: "class VII", shieldID: "class_vii_shield", strength: 7, arcMax: 35},
		{name: "class X", shieldID: "class_x_shield", strength: 10, arcMax: 50},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := checkpoint3ShieldBattleSpec(tc.shieldID, tc.strength, tc.arcMax)
			s, err := NewSession(spec)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.Start(); err != nil {
				t.Fatal(err)
			}
			view := s.View()
			state := tacticalStateForTest(t, view.Tactical.State.Ships, 200)
			wantArcs := [4]int{tc.arcMax, tc.arcMax, tc.arcMax, tc.arcMax}
			if state.ShieldArcCurrent != wantArcs {
				t.Fatalf("state shield arcs=%v want=%v", state.ShieldArcCurrent, wantArcs)
			}
			shipView := checkpoint3ShipView(t, view, 200)
			if shipView.ShieldID != tc.shieldID || shipView.ShieldStrength != tc.strength || shipView.ShieldArcMax != tc.arcMax || shipView.ShieldArcCurrent != wantArcs {
				t.Fatalf("shield view=%+v want id=%q strength=%d arcs=%v max=%d", shipView, tc.shieldID, tc.strength, wantArcs, tc.arcMax)
			}

			view.Tactical.State.Ships[1].ShieldArcCurrent[0] = -1
			view.Tactical.Ships[1].ShieldArcCurrent[0] = -1
			fresh := s.View()
			if got := tacticalStateForTest(t, fresh.Tactical.State.Ships, 200).ShieldArcCurrent[0]; got != tc.arcMax {
				t.Fatalf("mutating detached state view leaked into session: got=%d want=%d", got, tc.arcMax)
			}
			if got := checkpoint3ShipView(t, fresh, 200).ShieldArcCurrent[0]; got != tc.arcMax {
				t.Fatalf("mutating detached ship view leaked into session: got=%d want=%d", got, tc.arcMax)
			}
		})
	}
}

func TestSlice175Checkpoint3ShieldImpactFacingBoundaries(t *testing.T) {
	target := TacticalShipState{X: 0, Y: 0, Facing: 0}
	tests := []struct {
		name string
		x, y int
		want int
	}{
		{name: "front", x: 1, y: 0, want: 0},
		{name: "front boundary 45", x: 1, y: 1, want: 0},
		{name: "arc3 boundary 135", x: -1, y: 1, want: 3},
		{name: "arc2 boundary 225", x: -1, y: -1, want: 2},
		{name: "arc1 boundary 315", x: 1, y: -1, want: 1},
		{name: "wrapped front", x: 2, y: -1, want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := TacticalShipState{X: tc.x, Y: tc.y}
			if got := tacticalShieldImpactArc(target, source); got != tc.want {
				t.Fatalf("impact arc=%d want=%d", got, tc.want)
			}
		})
	}

	target.Facing = 4
	if got := tacticalShieldImpactArc(target, TacticalShipState{X: 0, Y: 2}); got != 0 {
		t.Fatalf("rotated front arc=%d want=0", got)
	}
	if got := tacticalShieldImpactArc(target, TacticalShipState{X: -2, Y: 0}); got != 3 {
		t.Fatalf("rotated port arc=%d want=3", got)
	}
}

func TestSlice175Checkpoint3FrozenSharedDamageVectors(t *testing.T) {
	source := TacticalShipState{ShipID: 100, X: 1, Y: 0}

	t.Run("no shield armor then structure", func(t *testing.T) {
		spec := TacticalShipSpec{ShipID: 200, ArmorMax: 4, StructureMax: 4}
		state := TacticalShipState{ShipID: 200, X: 0, Y: 0, ArmorCurrent: 4}
		got := applyTacticalDamage(spec, &state, source, 6)
		if got.ShieldArc != 0 || got.PostBlockDamage != 6 || got.ShieldAbsorbed != 0 || got.LowerLayerOverflow != 6 {
			t.Fatalf("no-shield envelope=%+v", got)
		}
		if state.ArmorCurrent != 0 || state.StructureDamage != 2 || state.Destroyed {
			t.Fatalf("no-shield state=%+v", state)
		}
		if got.ArmorBefore != 4 || got.ArmorAfter != 0 || got.StructureBefore != 0 || got.StructureAfter != 2 || got.Layer != "armor_structure" {
			t.Fatalf("no-shield damage result=%+v", got)
		}
	})

	t.Run("class III frigate overflow two to armor", func(t *testing.T) {
		spec := TacticalShipSpec{ShipID: 200, ShieldID: "class_iii_shield", ShieldStrength: 3, ShieldArcMax: 15, ArmorMax: 4, StructureMax: 4}
		state := TacticalShipState{ShipID: 200, X: 0, Y: 0, ArmorCurrent: 4, ShieldArcCurrent: [4]int{15, 15, 15, 15}}
		got := applyTacticalDamage(spec, &state, source, 20)
		if got.PostBlockDamage != 17 || got.ShieldBefore != 15 || got.ShieldAfter != 0 || got.ShieldAbsorbed != 15 || got.LowerLayerOverflow != 2 {
			t.Fatalf("frigate shield result=%+v", got)
		}
		if state.ArmorCurrent != 2 || state.StructureDamage != 0 || state.Destroyed {
			t.Fatalf("frigate state=%+v", state)
		}
	})

	t.Run("class III cruiser absorbs entire post-block hit", func(t *testing.T) {
		spec := TacticalShipSpec{ShipID: 200, ShieldID: "class_iii_shield", ShieldStrength: 3, ShieldArcMax: 45, ArmorMax: 30, StructureMax: 30}
		state := TacticalShipState{ShipID: 200, X: 0, Y: 0, ArmorCurrent: 30, ShieldArcCurrent: [4]int{45, 45, 45, 45}}
		got := applyTacticalDamage(spec, &state, source, 20)
		if got.PostBlockDamage != 17 || got.ShieldBefore != 45 || got.ShieldAfter != 28 || got.ShieldAbsorbed != 17 || got.LowerLayerOverflow != 0 {
			t.Fatalf("cruiser shield result=%+v", got)
		}
		if state.ArmorCurrent != 30 || state.StructureDamage != 0 || state.Destroyed || got.Layer != "shield" {
			t.Fatalf("cruiser state=%+v result=%+v", state, got)
		}
	})

	t.Run("overflow destroys deterministically", func(t *testing.T) {
		spec := TacticalShipSpec{ShipID: 200, ShieldID: "class_i_shield", ShieldStrength: 1, ShieldArcMax: 5, ArmorMax: 1, StructureMax: 4}
		state := TacticalShipState{ShipID: 200, X: 0, Y: 0, ArmorCurrent: 1, ShieldArcCurrent: [4]int{5, 5, 5, 5}}
		got := applyTacticalDamage(spec, &state, source, 20)
		if got.PostBlockDamage != 19 || got.ShieldAbsorbed != 5 || got.LowerLayerOverflow != 14 {
			t.Fatalf("overflow envelope=%+v", got)
		}
		if state.ArmorCurrent != 0 || state.StructureDamage != 4 || !state.Destroyed {
			t.Fatalf("destroyed state=%+v", state)
		}
		if got.ArmorBefore != 1 || got.ArmorAfter != 0 || got.StructureBefore != 0 || got.StructureAfter != 4 || got.Layer != "shield_armor_structure" {
			t.Fatalf("overflow result=%+v", got)
		}
	})
}

func TestSlice175Checkpoint3DeterministicRechargeAndRoundTransition(t *testing.T) {
	arcs := [4]int{}
	rechargeTacticalShieldArcs(&arcs, 45)
	if want := [4]int{4, 3, 3, 3}; arcs != want {
		t.Fatalf("empty Class III Cruiser recharge=%v want=%v", arcs, want)
	}

	arcs = [4]int{44, 45, 45, 45}
	rechargeTacticalShieldArcs(&arcs, 45)
	if want := [4]int{45, 45, 45, 45}; arcs != want {
		t.Fatalf("recharge exceeded/full handling=%v want=%v", arcs, want)
	}

	spec := checkpoint3ShieldBattleSpec("class_iii_shield", 3, 15)
	runtime, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	for i := range runtime.state.Ships {
		runtime.state.Ships[i].ActivationComplete = true
	}
	defender := runtime.shipState(200)
	defender.ShieldArcCurrent = [4]int{}
	if err := advanceAfterActivationCompletion(spec, &runtime); err != nil {
		t.Fatal(err)
	}
	if runtime.state.Round != 2 {
		t.Fatalf("round=%d want=2", runtime.state.Round)
	}
	if want := [4]int{1, 1, 1, 1}; defender.ShieldArcCurrent != want {
		t.Fatalf("round transition recharge=%v want=%v", defender.ShieldArcCurrent, want)
	}
}

func TestSlice175Checkpoint3BattleDamageEventCarriesShieldEnvelope(t *testing.T) {
	spec := checkpoint3ShieldBattleSpec("class_iii_shield", 3, 15)
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	s.mu.Lock()
	var chosen uint32
	for seed := uint32(1); seed < 100000; seed++ {
		probe := cloneTacticalRuntime(*s.tactical)
		probe.state.RNGState = seed
		raw, err := probe.random(100, s.spec.Tactical.Rules)
		if err != nil {
			s.mu.Unlock()
			t.Fatal(err)
		}
		if raw > s.spec.Tactical.Rules.BeamMaxHitThreshold {
			chosen = seed
			break
		}
	}
	if chosen == 0 {
		s.mu.Unlock()
		t.Fatal("failed to find deterministic critical-hit seed")
	}
	s.tactical.state.RNGState = chosen
	s.mu.Unlock()

	submitPrepared(t, s, 1, mustFireCommand(t, 1, 100, 200, 0))
	view := s.View()
	defender := tacticalStateForTest(t, view.Tactical.State.Ships, 200)
	if defender.ShieldArcCurrent[2] != 14 {
		t.Fatalf("rear shield after Laser=%v want arc2=14", defender.ShieldArcCurrent)
	}
	if defender.ArmorCurrent != 4 || defender.StructureDamage != 0 {
		t.Fatalf("shielded Laser leaked to lower layers: %+v", defender)
	}

	var data map[string]any
	found := false
	for _, event := range view.Tactical.Events {
		if event.Kind != "battle_damage_applied" {
			continue
		}
		if err := json.Unmarshal(event.Data, &data); err != nil {
			t.Fatal(err)
		}
		found = true
		break
	}
	if !found {
		t.Fatal("battle_damage_applied event missing")
	}
	wantNumbers := map[string]float64{
		"source_ship_id":          100,
		"target_ship_id":          200,
		"weapon_slot":             0,
		"incoming_damage":         4,
		"shield_arc":              2,
		"shield_before":           15,
		"shield_after":            14,
		"post_block_damage":       1,
		"shield_absorbed":         1,
		"lower_layer_overflow":    0,
		"armor_before":            4,
		"armor_after":             4,
		"structure_damage_before": 0,
		"structure_damage_after":  0,
	}
	for key, want := range wantNumbers {
		if got, ok := data[key].(float64); !ok || got != want {
			t.Fatalf("damage event %s=%v want=%v; data=%v", key, data[key], want, data)
		}
	}
	if got, _ := data["layer"].(string); got != "shield" {
		t.Fatalf("damage event layer=%q want shield; data=%v", got, data)
	}
}

func TestSlice175Checkpoint3ShieldStateLiveSnapshotBounds(t *testing.T) {
	spec := checkpoint3ShieldBattleSpec("class_iii_shield", 3, 15)
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	s.mu.Lock()
	s.tactical.shipState(200).ShieldArcCurrent = [4]int{2, 4, 6, 8}
	s.mu.Unlock()

	snapshot := s.LiveSnapshot()
	restored, err := RestoreLiveSnapshot(snapshot)
	if err != nil {
		t.Fatalf("valid shield state did not restore: %v", err)
	}
	if got := tacticalStateForTest(t, restored.View().Tactical.State.Ships, 200).ShieldArcCurrent; got != [4]int{2, 4, 6, 8} {
		t.Fatalf("restored shield state=%v", got)
	}

	corrupt := s.LiveSnapshot()
	for i := range corrupt.Tactical.State.Ships {
		if corrupt.Tactical.State.Ships[i].ShipID == 200 {
			corrupt.Tactical.State.Ships[i].ShieldArcCurrent[0] = 16
		}
	}
	if _, err := RestoreLiveSnapshot(corrupt); err == nil {
		t.Fatal("out-of-bounds shield arc unexpectedly restored")
	}
}
