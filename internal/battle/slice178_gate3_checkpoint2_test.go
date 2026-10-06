package battle

import "testing"

func slice178Checkpoint2SpecialRules() *TacticalSpecialRulesSnapshot {
	return &TacticalSpecialRulesSnapshot{
		AugmentedEnginesSpeedBonus:        5,
		BattleScannerOffenseBonus:         50,
		ECMJammerDefenseBonus:             70,
		MultiWaveECMJammerDefenseBonus:    100,
		WideAreaJammerSelfDefenseBonus:    130,
		WideAreaJammerFleetDefenseBonus:   70,
		CloakBeamDefenseBonus:             80,
		CloakMissileForcedMissThreshold:   50,
		PhasingInitialLifetime:            10,
		EnergyAbsorberDivisor:             4,
		DamperDamageDivisor:               4,
		DisplacementOrdnanceMissThreshold: 30,
		DisplacementBeamMissThreshold:     30,
		ShieldRechargeNormalFactor:        3,
		ShieldRechargeCapacitorFactor:     10,
		HardShieldsAdditionalAbsorption:   3,
		MultiPhasedShieldNumerator:        3,
		MultiPhasedShieldDenominator:      2,
		LightningDestroyThreshold:         50,
		QuantumExplosionMultiplier:        3,
		AutomatedRepairBasePercent:        20,
		SelfTeleportMaxRange:              20,
	}
}

func slice178Checkpoint2BattleSpec() Spec {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Rules.Specials = slice178Checkpoint2SpecialRules()
	ship := &spec.Tactical.Ships[0]
	ship.ShieldID = "class_i_shield"
	ship.ShieldStrength = 1
	ship.ShieldArcMax = 7
	ship.CurrentCombatSpeed = 27
	ship.BeamOffense = 75
	ship.BeamDefense = 135
	ship.Specials = []TacticalSpecialSpec{
		{Slot: 0, SpecialID: "augmented_engines", TacticalSupport: "tactical_passive"},
		{Slot: 1, SpecialID: "battle_scanner", TacticalSupport: "tactical_passive"},
		{Slot: 2, SpecialID: "ecm_jammer", TacticalSupport: "tactical_passive"},
		{Slot: 3, SpecialID: "multi_wave_ecm_jammer", TacticalSupport: "tactical_passive"},
		{Slot: 4, SpecialID: "wide_area_jammer", TacticalSupport: "tactical_passive"},
		{Slot: 5, SpecialID: "hard_shields", TacticalSupport: "tactical_passive"},
		{Slot: 6, SpecialID: "multi_phased_shields", TacticalSupport: "tactical_passive"},
	}
	ship.Internal = &TacticalInternalSpec{
		EngineMax:    2,
		SecondaryMax: 1,
		Weapons:      []TacticalInternalWeaponSpec{{Slot: 0, Max: 1, UnitSize: 1}},
		Specials: []TacticalInternalSpecialSpec{
			{Slot: 0, Max: 1}, {Slot: 1, Max: 1}, {Slot: 2, Max: 1}, {Slot: 3, Max: 1},
			{Slot: 4, Max: 1}, {Slot: 5, Max: 1}, {Slot: 6, Max: 1},
		},
		ShieldGeneratorMax: 1,
	}
	return spec
}

func TestSlice178Gate3Checkpoint2InternalDamageDrivesPassiveConsumers(t *testing.T) {
	spec := slice178Checkpoint2BattleSpec()
	if _, err := NewSession(spec); err != nil {
		t.Fatalf("CP2 spec rejected: %v", err)
	}
	runtime, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	shipSpec := spec.Tactical.Ships[0]
	state := runtime.shipState(shipSpec.ShipID)
	if state == nil {
		t.Fatal("CP2 ship state missing")
	}
	if state.CurrentCombatSpeed != 27 || tacticalBeamOffenseForState(*spec.Tactical, shipSpec, *state, false) != 75 {
		t.Fatalf("initial passive derivation speed/offense=%d/%d", state.CurrentCombatSpeed, tacticalBeamOffenseForState(*spec.Tactical, shipSpec, *state, false))
	}
	if got := tacticalMissileDefenseForState(*spec.Tactical, &runtime, shipSpec, *state); got != 130 {
		t.Fatalf("ECM best-only defense=%d want=130", got)
	}
	if got := tacticalShieldArcMaxForState(*spec.Tactical, shipSpec, *state, false); got != 7 {
		t.Fatalf("Multi-Phased shield max=%d want=7", got)
	}

	for _, tc := range []struct {
		slot int
		want int
	}{
		{4, 100},
		{3, 70},
		{2, 0},
	} {
		if err := applyTacticalInternalDamage(*spec.Tactical, &runtime, shipSpec.ShipID, TacticalInternalPoolSpecial, tc.slot, 1, "test"); err != nil {
			t.Fatal(err)
		}
		if got := tacticalMissileDefenseForState(*spec.Tactical, &runtime, shipSpec, *state); got != tc.want {
			t.Fatalf("ECM defense after special slot %d loss=%d want=%d", tc.slot, got, tc.want)
		}
	}

	if err := applyTacticalInternalDamage(*spec.Tactical, &runtime, shipSpec.ShipID, TacticalInternalPoolSpecial, 1, 1, "test"); err != nil {
		t.Fatal(err)
	}
	if got := tacticalBeamOffenseForState(*spec.Tactical, shipSpec, *state, false); got != 25 {
		t.Fatalf("Battle Scanner loss offense=%d want=25", got)
	}

	if err := applyTacticalInternalDamage(*spec.Tactical, &runtime, shipSpec.ShipID, TacticalInternalPoolSpecial, 0, 1, "test"); err != nil {
		t.Fatal(err)
	}
	if state.CurrentCombatSpeed != 22 || state.MovementMax != 22 {
		t.Fatalf("Augmented Engines loss speed/movement=%d/%d want=22/22", state.CurrentCombatSpeed, state.MovementMax)
	}

	if err := applyTacticalInternalDamage(*spec.Tactical, &runtime, shipSpec.ShipID, TacticalInternalPoolSpecial, 6, 1, "test"); err != nil {
		t.Fatal(err)
	}
	if got := tacticalShieldArcMaxForState(*spec.Tactical, shipSpec, *state, false); got != 5 {
		t.Fatalf("Multi-Phased loss shield max=%d want=5", got)
	}
	if state.ShieldArcCurrent != [4]int{5, 5, 5, 5} {
		t.Fatalf("Multi-Phased loss did not clamp live shields: %+v", state.ShieldArcCurrent)
	}

	if err := applyTacticalInternalDamage(*spec.Tactical, &runtime, shipSpec.ShipID, TacticalInternalPoolSpecial, 5, 1, "test"); err != nil {
		t.Fatal(err)
	}
	if TacticalSpecialFunctional(shipSpec, *state, false, "hard_shields") {
		t.Fatal("destroyed Hard Shields remained functional")
	}

	if err := applyTacticalInternalDamage(*spec.Tactical, &runtime, shipSpec.ShipID, TacticalInternalPoolWeapon, 0, 1, "test"); err != nil {
		t.Fatal(err)
	}
	if got := tacticalCurrentWeaponCount(shipSpec, *state, 0); got != 0 || state.Weapons[0].Ready {
		t.Fatalf("destroyed weapon pool count/ready=%d/%v want=0/false", got, state.Weapons[0].Ready)
	}

	if err := applyTacticalInternalDamage(*spec.Tactical, &runtime, shipSpec.ShipID, TacticalInternalPoolShieldGenerator, 0, 1, "test"); err != nil {
		t.Fatal(err)
	}
	if state.ShieldArcCurrent != [4]int{} {
		t.Fatalf("destroyed shield generator left live facings: %+v", state.ShieldArcCurrent)
	}

	if err := applyTacticalInternalDamage(*spec.Tactical, &runtime, shipSpec.ShipID, TacticalInternalPoolEngine, 0, 1, "test"); err != nil {
		t.Fatal(err)
	}
	if state.CurrentCombatSpeed != 0 || state.MovementMax != 0 {
		t.Fatalf("engine threshold speed/movement=%d/%d want=0/0", state.CurrentCombatSpeed, state.MovementMax)
	}
	if len(runtime.events) == 0 || runtime.events[len(runtime.events)-1].Kind != "internal_damage_applied" {
		t.Fatalf("internal damage event missing: %+v", runtime.events)
	}
}

func TestSlice178Gate3Checkpoint2MigratesPreCheckpoint2Schema2State(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Ships[0].Internal = &TacticalInternalSpec{
		EngineMax:    2,
		SecondaryMax: 1,
		Weapons:      []TacticalInternalWeaponSpec{{Slot: 0, Max: 1, UnitSize: 1}},
	}
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	snapshot := s.LiveSnapshot()
	if snapshot.Tactical == nil {
		t.Fatal("tactical live snapshot missing")
	}
	snapshot.Tactical.Slice178StateVersion = 0
	legacy := &snapshot.Tactical.State.Ships[0]
	legacy.CurrentCombatSpeed = 0
	legacy.MovementMax = 0
	legacy.Weapons[0].CountCurrent = 0
	legacy.Internal = TacticalInternalState{}

	restored, err := RestoreLiveSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	got := restored.LiveSnapshot()
	if got.Tactical == nil || got.Tactical.Slice178StateVersion != 4 {
		t.Fatalf("migrated Slice-17.8 state version=%+v", got.Tactical)
	}
	state := got.Tactical.State.Ships[0]
	if state.CurrentCombatSpeed != spec.Tactical.Ships[0].CurrentCombatSpeed || state.MovementMax != spec.Tactical.Ships[0].CurrentCombatSpeed {
		t.Fatalf("migrated speed/movement=%d/%d", state.CurrentCombatSpeed, state.MovementMax)
	}
	if state.Weapons[0].CountCurrent != spec.Tactical.Ships[0].Weapons[0].Count {
		t.Fatalf("migrated weapon count=%d want=%d", state.Weapons[0].CountCurrent, spec.Tactical.Ships[0].Weapons[0].Count)
	}
	if state.Internal.EngineCurrent != 2 || state.Internal.SecondaryCurrent != 1 || len(state.Internal.Weapons) != 1 || state.Internal.Weapons[0].Current != 1 {
		t.Fatalf("migrated internal state=%+v", state.Internal)
	}
	if legacy.CurrentCombatSpeed != 0 || legacy.Weapons[0].CountCurrent != 0 || legacy.Internal.EngineCurrent != 0 {
		t.Fatal("CP2 migration mutated caller-owned snapshot")
	}
}
