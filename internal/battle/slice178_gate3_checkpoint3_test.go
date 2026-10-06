package battle

import (
	"encoding/json"
	"testing"
)

func slice178Checkpoint3FighterRules() *TacticalFighterRulesSnapshot {
	return &TacticalFighterRulesSnapshot{
		MovementUnitsPerField:      20,
		CraftPerBay:                4,
		FighterBaseSpeed:           10,
		BomberBaseSpeed:            8,
		HeavyFighterBaseSpeed:      8,
		FighterBaseHits:            2,
		BomberBaseHits:             4,
		HeavyFighterBaseHits:       5,
		FighterAttackStock:         4,
		BomberAttackStock:          1,
		HeavyFighterAttackStock:    2,
		WarpDriveSpeedStep:         2,
		TransDimensionalSpeedBonus: 4,
		ArmorDurabilityMultiples:   []int{1, 2, 4, 6, 8, 10},
		RecoveryDivisor:            4,
	}
}

func slice178Checkpoint3Spec(weaponID string, count int) Spec {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Rules.Fighter = slice178Checkpoint3FighterRules()
	spec.Tactical.Ships[0].Weapons = append(spec.Tactical.Ships[0].Weapons, TacticalWeaponSpec{
		Slot: 1, WeaponID: weaponID, WeaponFamily: "fighter", Count: count,
		TacticalSupport: "fighter", CraftPerUnit: 4,
	})
	spec.Tactical.FighterDoctrines = []TacticalFighterDoctrineSpec{{
		SeatID: 1, EmpireID: 1, BestWarpDriveIndex: 1, ArmorDurabilityMultiple: 1,
		FighterBeamOffense: 100,
		BeamPayload: &TacticalFighterPayloadSpec{
			WeaponID: "laser_cannon", MinDamage: 0, MaxDamage: 0, PointDefenseMode: true,
		},
		BombPayload: &TacticalFighterPayloadSpec{
			WeaponID: "nuclear_bomb", MinDamage: 0, MaxDamage: 0,
		},
	}}
	return spec
}

func slice178Checkpoint3Runtime(t *testing.T, spec Spec) tacticalRuntime {
	t.Helper()
	if err := validateSpec(spec); err != nil {
		t.Fatalf("invalid CP3 test spec: %v", err)
	}
	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func slice178Checkpoint3EventData(t *testing.T, event TacticalEvent) map[string]any {
	t.Helper()
	var data map[string]any
	if err := json.Unmarshal(event.Data, &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func slice178Checkpoint3Events(events []TacticalEvent, kind string) []TacticalEvent {
	out := make([]TacticalEvent, 0)
	for _, event := range events {
		if event.Kind == kind {
			out = append(out, event)
		}
	}
	return out
}

func TestSlice178Gate3Checkpoint3VectorALaunchLossRecovery(t *testing.T) {
	spec := slice178Checkpoint3Spec("fighter_bays", 2)
	spec.Tactical.Ships[1].X = 30
	r := slice178Checkpoint3Runtime(t, spec)

	if err := prepareLaunchFighters(spec, &r, 1, 1, LaunchFightersPayload{
		ShipID: 100, WeaponSlot: 1, TargetShipID: 200,
	}); err != nil {
		t.Fatal(err)
	}
	if len(r.fighters.Groups) != 1 {
		t.Fatalf("groups after launch=%+v", r.fighters.Groups)
	}
	group := &r.fighters.Groups[0]
	if group.Multiplicity != 8 || group.Speed != 10 || group.UnitDurability != 2 || group.AttackStock != 4 {
		t.Fatalf("Vector A launch group=%+v", *group)
	}
	bay := tacticalFighterBayState(&r.fighters, 100, 1)
	shipState := r.shipState(100)
	ready := weaponState(shipState, 1)
	if bay == nil || bay.ReadyGroups != 0 || ready == nil || ready.Ready {
		t.Fatalf("Vector A launch Bay/readiness bay=%+v ready=%+v", bay, ready)
	}

	group.Multiplicity = 7
	group.AttackStock = 0
	group.Phase = TacticalFighterPhaseReturning
	group.TargetShipID = group.SourceShipID
	group.TargetProjectileID = 0
	group.TargetFighterGroupID = 0
	group.X, group.Y = shipState.X, shipState.Y
	groupID := group.GroupID
	if err := processTacticalFighterGroupByID(*spec.Tactical, &r, groupID, 1, 2); err != nil {
		t.Fatal(err)
	}
	if len(r.fighters.Groups) != 0 {
		t.Fatalf("Vector A recovered group still active=%+v", r.fighters.Groups)
	}
	if bay.ReadyGroups != 1 || !ready.Ready {
		t.Fatalf("Vector A recovery Bay/readiness bay=%+v ready=%+v", bay, ready)
	}
	recovered := slice178Checkpoint3Events(r.events, "fighter_group_recovered")
	if len(recovered) != 1 {
		t.Fatalf("Vector A recovery events=%+v", r.events)
	}
	data := slice178Checkpoint3EventData(t, recovered[0])
	if data["complete_groups"] != float64(1) || data["recovered_groups"] != float64(1) ||
		data["incomplete_craft_discarded"] != float64(3) {
		t.Fatalf("Vector A recovery event=%+v", data)
	}
}

func TestSlice178Gate3Checkpoint3VectorBHeavyPairedCadence(t *testing.T) {
	spec := slice178Checkpoint3Spec("heavy_fighter_bays", 1)
	r := slice178Checkpoint3Runtime(t, spec)

	if err := prepareLaunchFighters(spec, &r, 1, 1, LaunchFightersPayload{
		ShipID: 100, WeaponSlot: 1, TargetShipID: 200,
	}); err != nil {
		t.Fatal(err)
	}
	if len(r.fighters.Groups) != 1 {
		t.Fatalf("heavy groups after launch=%+v", r.fighters.Groups)
	}
	group := r.fighters.Groups[0]
	if group.Multiplicity != 4 || group.Speed != 8 || group.UnitDurability != 5 || group.AttackStock != 1 || group.Phase != TacticalFighterPhaseOutbound {
		t.Fatalf("Vector B first cycle group=%+v", group)
	}

	if err := processTacticalFighterGroupByID(*spec.Tactical, &r, group.GroupID, 1, 2); err != nil {
		t.Fatal(err)
	}
	if len(r.fighters.Groups) != 1 {
		t.Fatalf("heavy group disappeared after second cycle=%+v", r.fighters.Groups)
	}
	group = r.fighters.Groups[0]
	if group.AttackStock != 0 || group.Phase != TacticalFighterPhaseReturning {
		t.Fatalf("Vector B second cycle group=%+v", group)
	}
	attacks := slice178Checkpoint3Events(r.events, "fighter_attack_resolved")
	if len(attacks) != 4 {
		t.Fatalf("Vector B attack events=%d events=%+v", len(attacks), r.events)
	}
	want := []string{"beam", "bomb", "beam", "bomb"}
	for i, event := range attacks {
		data := slice178Checkpoint3EventData(t, event)
		if data["payload_kind"] != want[i] {
			t.Fatalf("Vector B attack[%d]=%+v want payload=%q", i, data, want[i])
		}
		attempts, ok := data["attempts"].([]any)
		if !ok || len(attempts) != 4 {
			t.Fatalf("Vector B attack[%d] attempts=%+v", i, data["attempts"])
		}
	}
}

func slice178Checkpoint3SeedForTwoLightningLosses(rules TacticalRulesSnapshot) uint32 {
	for seed := uint32(1); seed < 100000; seed++ {
		r := tacticalRuntime{state: TacticalState{RNGState: seed}}
		a, _ := r.random(100, rules)
		b, _ := r.random(100, rules)
		if a-1 < 50 && b-1 < 50 {
			return seed
		}
	}
	return 0
}

func TestSlice178Gate3Checkpoint3VectorGLightningMutableLoopBeforePD(t *testing.T) {
	spec := slice178Checkpoint3Spec("fighter_bays", 1)
	spec.Tactical.Rules.Specials = slice178Checkpoint2SpecialRules()
	spec.Tactical.Ships[1].Specials = []TacticalSpecialSpec{{
		Slot: 0, SpecialID: "lightning_field", TacticalSupport: "tactical_passive",
	}}
	spec.Tactical.Ships[1].Internal = &TacticalInternalSpec{EngineMax: 2, SecondaryMax: 1, Specials: []TacticalInternalSpecialSpec{{Slot: 0, Max: 1}}}
	seed := slice178Checkpoint3SeedForTwoLightningLosses(spec.Tactical.Rules)
	if seed == 0 {
		t.Fatal("could not find deterministic Lightning seed")
	}
	r := slice178Checkpoint3Runtime(t, spec)
	r.state.RNGState = seed
	targetState := r.shipState(200)
	targetState.Internal.Specials = []TacticalInternalSpecialState{{Slot: 0, Current: 1}}

	if err := prepareLaunchFighters(spec, &r, 1, 1, LaunchFightersPayload{
		ShipID: 100, WeaponSlot: 1, TargetShipID: 200,
	}); err != nil {
		t.Fatal(err)
	}
	if len(r.fighters.Groups) != 1 || r.fighters.Groups[0].Multiplicity != 2 {
		t.Fatalf("Vector G post-Lightning group=%+v", r.fighters.Groups)
	}
	intercepts := slice178Checkpoint3Events(r.events, "fighter_interception_resolved")
	if len(intercepts) != 1 {
		t.Fatalf("Vector G interceptions=%+v", intercepts)
	}
	data := slice178Checkpoint3EventData(t, intercepts[0])
	if data["interceptor_source_kind"] != "lightning_field" ||
		data["multiplicity_before"] != float64(4) || data["multiplicity_after"] != float64(2) ||
		data["destroyed_craft"] != float64(2) {
		t.Fatalf("Vector G Lightning event=%+v", data)
	}
	rolls, ok := data["rolls"].([]any)
	if !ok || len(rolls) != 2 {
		t.Fatalf("Vector G Lightning rolls=%+v", data["rolls"])
	}
	firstInterception := -1
	firstAttack := -1
	for i, event := range r.events {
		if firstInterception < 0 && event.Kind == "fighter_interception_resolved" {
			firstInterception = i
		}
		if firstAttack < 0 && event.Kind == "fighter_attack_resolved" {
			firstAttack = i
		}
	}
	if firstInterception < 0 || firstAttack < 0 || firstInterception >= firstAttack {
		t.Fatalf("Vector G event ordering interception=%d attack=%d events=%+v", firstInterception, firstAttack, r.events)
	}
}

func TestSlice178Gate3Checkpoint3FighterMovingActorDCVUsesFrozenDoctrineBonus(t *testing.T) {
	spec := slice178Checkpoint3Spec("fighter_bays", 1)
	spec.Tactical.FighterDoctrines[0].MovingActorDefenseBonus = 25

	got, err := tacticalFighterBeamInterceptorDCV(*spec.Tactical, TacticalFighterGroupState{
		GroupID:      1,
		SourceSeatID: 1,
		Speed:        10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != 75 {
		t.Fatalf("Fighter moving-actor DCV=%d want=75 (speed*5 + frozen defense bonus)", got)
	}
}
