package battle

import "testing"

func slice178Checkpoint5SpecWithSpecial(id string) TacticalSpec {
	spec := baselineTacticalBattleSpec().Tactical
	spec.Rules.Specials = slice178Checkpoint2SpecialRules()
	spec.Ships[0].Specials = []TacticalSpecialSpec{{Slot: 0, SpecialID: id, TacticalSupport: "tactical_passive"}}
	spec.Ships[0].Internal = &TacticalInternalSpec{
		Specials: []TacticalInternalSpecialSpec{{Slot: 0, Max: 1}},
	}
	return *spec
}

func TestSlice178Gate3Checkpoint5InitialCloakAndPhasingState(t *testing.T) {
	t.Run("cloaking device", func(t *testing.T) {
		spec := slice178Checkpoint5SpecWithSpecial("cloaking_device")
		r, err := newTacticalRuntime(spec)
		if err != nil {
			t.Fatal(err)
		}
		state := r.shipState(100)
		if state.CloakState != TacticalCloakStateCloaked || state.PhasingTurnsRemaining != 0 {
			t.Fatalf("Cloaking Device initial state=%q/%d want=%q/0", state.CloakState, state.PhasingTurnsRemaining, TacticalCloakStateCloaked)
		}
	})

	t.Run("phasing cloak", func(t *testing.T) {
		spec := slice178Checkpoint5SpecWithSpecial("phasing_cloak")
		r, err := newTacticalRuntime(spec)
		if err != nil {
			t.Fatal(err)
		}
		state := r.shipState(100)
		if state.CloakState != TacticalCloakStatePhased || state.PhasingTurnsRemaining != 10 {
			t.Fatalf("Phasing Cloak initial state=%q/%d want=%q/10", state.CloakState, state.PhasingTurnsRemaining, TacticalCloakStatePhased)
		}
	})
}

func TestSlice178Gate3Checkpoint5InitialSpecialReadiness(t *testing.T) {
	t.Run("Hyper-X Beam", func(t *testing.T) {
		spec := slice178Checkpoint5SpecWithSpecial("hyper_x_capacitors")
		r, err := newTacticalRuntime(spec)
		if err != nil {
			t.Fatal(err)
		}
		state := r.shipState(100)
		if got := state.Weapons[0].SpecialReadiness; got != TacticalSpecialReadinessExtraVolleyReady {
			t.Fatalf("Hyper-X initial special readiness=%q want=%q", got, TacticalSpecialReadinessExtraVolleyReady)
		}
	})

	t.Run("Fast Missile Racks", func(t *testing.T) {
		ship := TacticalShipSpec{
			Specials: []TacticalSpecialSpec{{Slot: 0, SpecialID: "fast_missile_racks", TacticalSupport: "tactical_passive"}},
			Internal: &TacticalInternalSpec{Specials: []TacticalInternalSpecialSpec{{Slot: 0, Max: 1}}},
		}
		state := TacticalShipState{Internal: initialTacticalInternalState(ship.Internal)}
		weapon := TacticalWeaponSpec{WeaponID: "test_missile", WeaponFamily: "missile"}
		if got := tacticalInitialSpecialReadiness(ship, state, weapon); got != TacticalSpecialReadinessExtraVolleyReady {
			t.Fatalf("FMR initial special readiness=%q want=%q", got, TacticalSpecialReadinessExtraVolleyReady)
		}
	})
}

func TestSlice178Gate3Checkpoint5VectorHShieldRecharge(t *testing.T) {
	t.Run("normal factor pointwise", func(t *testing.T) {
		arcs := [4]int{16, 16, 16, 16}
		rechargeTacticalShieldArcsWithBudget(&arcs, 20, 20, 3)
		if want := [4]int{18, 18, 17, 17}; arcs != want {
			t.Fatalf("normal recharge=%v want=%v", arcs, want)
		}
	})

	t.Run("Shield Capacitors restore all missing points", func(t *testing.T) {
		spec := slice178Checkpoint5SpecWithSpecial("shield_capacitors")
		spec.Ships[0].ShieldStrength = 4
		spec.Ships[0].ShieldArcMax = 20
		r, err := newTacticalRuntime(spec)
		if err != nil {
			t.Fatal(err)
		}
		state := r.shipState(100)
		state.ShieldArcCurrent = [4]int{16, 16, 16, 16}
		rechargeTacticalShields(spec, &r)
		if want := [4]int{20, 20, 20, 20}; state.ShieldArcCurrent != want {
			t.Fatalf("Shield Capacitors recharge=%v want=%v", state.ShieldArcCurrent, want)
		}
	})
}

func TestSlice178Gate3Checkpoint5VectorITeleportMove(t *testing.T) {
	spec := slice178Checkpoint5SpecWithSpecial("phase_shifter")
	r, err := newTacticalRuntime(spec)
	if err != nil {
		t.Fatal(err)
	}
	state := r.shipState(100)
	state.MovementCurrent = 0
	facing := state.Facing

	var atRange20 *TacticalMoveOption
	for _, option := range legalMoves(spec, &r) {
		if option.X == state.X+60 && option.Y == state.Y {
			copy := option
			atRange20 = &copy
			break
		}
	}
	if atRange20 == nil || atRange20.MovementKind != TacticalMovementKindTeleport || atRange20.MoveCost != 0 || atRange20.MovementRemainingAfter != 0 || atRange20.ResultingFacing != facing {
		t.Fatalf("Range-20 teleport projection=%+v", atRange20)
	}

	if err := prepareMoveShip(Spec{Tactical: &spec}, &r, 1, 1, MoveShipPayload{ShipID: 100, X: 70, Y: 10}); err != nil {
		t.Fatal(err)
	}
	if state.X != 70 || state.Y != 10 || state.Facing != facing || state.MovementCurrent != 0 || !state.TeleporterUsed {
		t.Fatalf("accepted teleport state=%+v", *state)
	}
	if err := prepareMoveShip(Spec{Tactical: &spec}, &r, 1, 2, MoveShipPayload{ShipID: 100, X: 69, Y: 10}); err == nil {
		t.Fatal("second teleport in same ship turn was accepted")
	}

	r2, err := newTacticalRuntime(spec)
	if err != nil {
		t.Fatal(err)
	}
	r2.shipState(100).MovementCurrent = 0
	if err := prepareMoveShip(Spec{Tactical: &spec}, &r2, 1, 1, MoveShipPayload{ShipID: 100, X: 73, Y: 10}); err == nil {
		t.Fatal("Range-21 teleport was accepted")
	}
}

func TestSlice178Gate3Checkpoint5LiveSnapshotV4AndV3Migration(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	snapshot := s.LiveSnapshot()
	if snapshot.Tactical == nil || snapshot.Tactical.Slice178StateVersion != 4 {
		t.Fatalf("emitted Slice-17.8 state version=%+v want=4", snapshot.Tactical)
	}
	for _, ship := range snapshot.Tactical.State.Ships {
		if ship.CloakState != TacticalCloakStateNone {
			t.Fatalf("baseline emitted cloak state=%q want none", ship.CloakState)
		}
		for _, weapon := range ship.Weapons {
			if weapon.SpecialReadiness != TacticalSpecialReadinessNone {
				t.Fatalf("baseline emitted special readiness=%q want none", weapon.SpecialReadiness)
			}
		}
	}

	snapshot.Tactical.Slice178StateVersion = 3
	for i := range snapshot.Tactical.State.Ships {
		snapshot.Tactical.State.Ships[i].CloakState = ""
		snapshot.Tactical.State.Ships[i].PhasingTurnsRemaining = 0
		snapshot.Tactical.State.Ships[i].EnergyAbsorberStored = 0
		for j := range snapshot.Tactical.State.Ships[i].Weapons {
			snapshot.Tactical.State.Ships[i].Weapons[j].SpecialReadiness = ""
		}
	}
	restored, err := RestoreLiveSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	after := restored.LiveSnapshot()
	if after.Tactical == nil || after.Tactical.Slice178StateVersion != 4 {
		t.Fatalf("restored Slice-17.8 state version=%+v want=4", after.Tactical)
	}
	for _, ship := range after.Tactical.State.Ships {
		if ship.CloakState != TacticalCloakStateNone || ship.EnergyAbsorberStored != 0 {
			t.Fatalf("migrated baseline Special state=%+v", ship)
		}
		for _, weapon := range ship.Weapons {
			if weapon.SpecialReadiness != TacticalSpecialReadinessNone {
				t.Fatalf("migrated baseline special readiness=%q want none", weapon.SpecialReadiness)
			}
		}
	}
}

func TestSlice178Gate3Checkpoint5RestoreRejectsInvalidSpecialState(t *testing.T) {
	baselineSnapshot := func(t *testing.T) LiveSnapshot {
		t.Helper()
		s, err := NewSession(baselineTacticalBattleSpec())
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Start(); err != nil {
			t.Fatal(err)
		}
		return s.LiveSnapshot()
	}

	t.Run("readiness on non-eligible row", func(t *testing.T) {
		snapshot := baselineSnapshot(t)
		snapshot.Tactical.State.Ships[0].Weapons[0].SpecialReadiness = TacticalSpecialReadinessExtraVolleyReady
		if _, err := RestoreLiveSnapshot(snapshot); err == nil {
			t.Fatal("restore accepted special readiness on non-eligible Beam row")
		}
	})

	t.Run("cloak state without installed Special", func(t *testing.T) {
		snapshot := baselineSnapshot(t)
		snapshot.Tactical.State.Ships[0].CloakState = TacticalCloakStateCloaked
		if _, err := RestoreLiveSnapshot(snapshot); err == nil {
			t.Fatal("restore accepted cloak lifecycle without installed cloak Special")
		}
	})

	t.Run("Absorber store without installed Special", func(t *testing.T) {
		snapshot := baselineSnapshot(t)
		snapshot.Tactical.State.Ships[0].EnergyAbsorberStored = 1
		if _, err := RestoreLiveSnapshot(snapshot); err == nil {
			t.Fatal("restore accepted Energy-Absorber store without installed Special")
		}
	})

	t.Run("phasing counter above frozen lifetime", func(t *testing.T) {
		snapshot := baselineSnapshot(t)
		snapshot.Spec.Tactical.Rules.Specials = slice178Checkpoint2SpecialRules()
		snapshot.Tactical.State.Ships[0].CloakState = TacticalCloakStatePhased
		snapshot.Tactical.State.Ships[0].PhasingTurnsRemaining = 11
		if _, err := RestoreLiveSnapshot(snapshot); err == nil {
			t.Fatal("restore accepted Phasing-Cloak counter above frozen lifetime 10")
		}
	})
}
func slice178Checkpoint5SeedForFirstRoll(t *testing.T, rules TacticalRulesSnapshot, bound uint32, want int) uint32 {
	t.Helper()
	for seed := uint32(1); seed < 100000; seed++ {
		r := tacticalRuntime{state: TacticalState{RNGState: seed}}
		got, err := r.random(bound, rules)
		if err != nil {
			t.Fatal(err)
		}
		if got == want {
			return seed
		}
	}
	t.Fatalf("no deterministic seed found for Random(%d)=%d", bound, want)
	return 0
}

func TestSlice178Gate3Checkpoint5VectorEEnergyAbsorberBeforeDamper(t *testing.T) {
	spec := *baselineTacticalBattleSpec().Tactical
	spec.Rules.Specials = slice178Checkpoint2SpecialRules()
	target := &spec.Ships[1]
	target.Specials = []TacticalSpecialSpec{
		{Slot: 0, SpecialID: "energy_absorber", TacticalSupport: "tactical_passive"},
		{Slot: 1, SpecialID: "damper_field", TacticalSupport: "tactical_passive"},
	}
	target.Internal = &TacticalInternalSpec{
		Specials: []TacticalInternalSpecialSpec{{Slot: 0, Max: 1}, {Slot: 1, Max: 1}},
	}
	r, err := newTacticalRuntime(spec)
	if err != nil {
		t.Fatal(err)
	}
	targetState := r.shipState(target.ShipID)
	sourceState := r.shipState(spec.Ships[0].ShipID)
	targetState.ArmorCurrent = 0
	targetState.ShieldArcCurrent = [4]int{}

	applied := applyTacticalDamageWithSpecials(spec, &r, *target, targetState, *sourceState, 20, tacticalDamageOptions{})
	if targetState.EnergyAbsorberStored != 5 {
		t.Fatalf("Energy Absorber stored=%d want=5", targetState.EnergyAbsorberStored)
	}
	if applied.IncomingDamage != 20 || applied.EnergyAbsorbed != 5 || applied.PostSpecialDamage != 3 || applied.DamperReducedDamage != 12 {
		t.Fatalf("Vector E preprocessing=%+v", applied)
	}
	if targetState.StructureDamage != 3 {
		t.Fatalf("Vector E downstream structure damage=%d want=3", targetState.StructureDamage)
	}
}

func TestSlice178Gate3Checkpoint5EnergyAbsorberPreWeaponDischarge(t *testing.T) {
	t.Run("range seven auto-discharge", func(t *testing.T) {
		spec := slice178Checkpoint5SpecWithSpecial("energy_absorber")
		spec.Ships[1].X = 31 // source X=10 -> Tactical range 7
		r, err := newTacticalRuntime(spec)
		if err != nil {
			t.Fatal(err)
		}
		source := spec.Ships[0]
		target := spec.Ships[1]
		sourceState := r.shipState(source.ShipID)
		targetState := r.shipState(target.ShipID)
		sourceState.EnergyAbsorberStored = 5
		targetState.ArmorCurrent = 0

		destroyed, err := tacticalResolveEnergyAbsorberPreWeapon(spec, &r, source, sourceState, target, targetState, 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if !destroyed || !targetState.Destroyed || targetState.StructureDamage != target.StructureMax || sourceState.EnergyAbsorberStored != 0 {
			t.Fatalf("range-7 Absorber discharge source=%+v target=%+v", *sourceState, *targetState)
		}
	})

	t.Run("range eight keeps store", func(t *testing.T) {
		spec := slice178Checkpoint5SpecWithSpecial("energy_absorber")
		spec.Ships[1].X = 32 // source X=10 -> Tactical range 8
		r, err := newTacticalRuntime(spec)
		if err != nil {
			t.Fatal(err)
		}
		source := spec.Ships[0]
		target := spec.Ships[1]
		sourceState := r.shipState(source.ShipID)
		targetState := r.shipState(target.ShipID)
		sourceState.EnergyAbsorberStored = 5
		targetState.ArmorCurrent = 0

		destroyed, err := tacticalResolveEnergyAbsorberPreWeapon(spec, &r, source, sourceState, target, targetState, 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if destroyed || targetState.StructureDamage != 0 || sourceState.EnergyAbsorberStored != 5 {
			t.Fatalf("range-8 Absorber state source=%+v target=%+v", *sourceState, *targetState)
		}
	})

	t.Run("direct Displacement branch uses greater-than-30 miss", func(t *testing.T) {
		spec := slice178Checkpoint5SpecWithSpecial("energy_absorber")
		spec.Ships[1].X = 31
		spec.Ships[1].Specials = []TacticalSpecialSpec{{Slot: 0, SpecialID: "displacement_device", TacticalSupport: "tactical_passive"}}
		spec.Ships[1].Internal = &TacticalInternalSpec{Specials: []TacticalInternalSpecialSpec{{Slot: 0, Max: 1}}}
		r, err := newTacticalRuntime(spec)
		if err != nil {
			t.Fatal(err)
		}
		source := spec.Ships[0]
		target := spec.Ships[1]
		sourceState := r.shipState(source.ShipID)
		targetState := r.shipState(target.ShipID)
		sourceState.EnergyAbsorberStored = 5
		targetState.ArmorCurrent = 0
		r.state.RNGState = slice178Checkpoint5SeedForFirstRoll(t, spec.Rules, 100, 31)

		destroyed, err := tacticalResolveEnergyAbsorberPreWeapon(spec, &r, source, sourceState, target, targetState, 1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if destroyed || targetState.StructureDamage != 0 || sourceState.EnergyAbsorberStored != 0 {
			t.Fatalf("Displacement-missed Absorber source=%+v target=%+v", *sourceState, *targetState)
		}
	})
}
func TestSlice178Gate3Checkpoint5VectorFReflectionBoundary(t *testing.T) {
	spec := slice178Checkpoint5SpecWithSpecial("reflection_field")
	target := spec.Ships[0]
	r, err := newTacticalRuntime(spec)
	if err != nil {
		t.Fatal(err)
	}
	state := r.shipState(target.ShipID)

	for _, tc := range []struct {
		roll          int
		wantRemaining int
		wantReflected int
	}{
		{roll: 20, wantRemaining: 20, wantReflected: 0},
		{roll: 21, wantRemaining: 0, wantReflected: 20},
	} {
		r.state.RNGState = slice178Checkpoint5SeedForFirstRoll(t, spec.Rules, 35, tc.roll)
		remaining, reflected, gotRoll, err := tacticalResolveBeamReflection(spec, &r, target, *state, 20)
		if err != nil {
			t.Fatal(err)
		}
		if gotRoll != tc.roll || remaining != tc.wantRemaining || reflected != tc.wantReflected {
			t.Fatalf("Reflection roll=%d got roll/remain/reflected=%d/%d/%d want=%d/%d/%d", tc.roll, gotRoll, remaining, reflected, tc.roll, tc.wantRemaining, tc.wantReflected)
		}
	}
}

func TestSlice178Gate3Checkpoint5DisplacementAndCloakForcedMissBoundaries(t *testing.T) {
	t.Run("Displacement Beam <=30", func(t *testing.T) {
		spec := slice178Checkpoint5SpecWithSpecial("displacement_device")
		r, err := newTacticalRuntime(spec)
		if err != nil {
			t.Fatal(err)
		}
		target := spec.Ships[0]
		state := r.shipState(target.ShipID)
		for _, tc := range []struct {
			roll int
			miss bool
		}{{30, true}, {31, false}} {
			r.state.RNGState = slice178Checkpoint5SeedForFirstRoll(t, spec.Rules, 100, tc.roll)
			miss, roll, err := tacticalBeamForcedMiss(spec, &r, target, *state)
			if err != nil {
				t.Fatal(err)
			}
			if roll != tc.roll || miss != tc.miss {
				t.Fatalf("Beam Displacement roll=%d miss=%v want=%v", roll, miss, tc.miss)
			}
		}
	})

	t.Run("Displacement Ordnance <30", func(t *testing.T) {
		spec := slice178Checkpoint5SpecWithSpecial("displacement_device")
		r, err := newTacticalRuntime(spec)
		if err != nil {
			t.Fatal(err)
		}
		target := spec.Ships[0]
		state := r.shipState(target.ShipID)
		for _, tc := range []struct {
			roll int
			miss bool
		}{{29, true}, {30, false}} {
			r.state.RNGState = slice178Checkpoint5SeedForFirstRoll(t, spec.Rules, 100, tc.roll)
			miss, displacementRoll, _, err := tacticalOrdnanceForcedMiss(spec, &r, target, *state)
			if err != nil {
				t.Fatal(err)
			}
			if displacementRoll != tc.roll || miss != tc.miss {
				t.Fatalf("Ordnance Displacement roll=%d miss=%v want=%v", displacementRoll, miss, tc.miss)
			}
		}
	})

	t.Run("Cloak Missile <50", func(t *testing.T) {
		spec := slice178Checkpoint5SpecWithSpecial("cloaking_device")
		r, err := newTacticalRuntime(spec)
		if err != nil {
			t.Fatal(err)
		}
		target := spec.Ships[0]
		state := r.shipState(target.ShipID)
		for _, tc := range []struct {
			roll int
			miss bool
		}{{49, true}, {50, false}} {
			r.state.RNGState = slice178Checkpoint5SeedForFirstRoll(t, spec.Rules, 100, tc.roll)
			miss, _, cloakRoll, err := tacticalOrdnanceForcedMiss(spec, &r, target, *state)
			if err != nil {
				t.Fatal(err)
			}
			if cloakRoll != tc.roll || miss != tc.miss {
				t.Fatalf("Cloak Missile roll=%d miss=%v want=%v", cloakRoll, miss, tc.miss)
			}
		}
	})
}

func TestSlice178Gate3Checkpoint5CloakAndPhasingLifecycle(t *testing.T) {
	t.Run("Cloaking Device break recover recloak", func(t *testing.T) {
		spec := slice178Checkpoint5SpecWithSpecial("cloaking_device")
		r, err := newTacticalRuntime(spec)
		if err != nil {
			t.Fatal(err)
		}
		state := r.shipState(100)
		tacticalBreakCloakOnAction(state)
		if state.CloakState != TacticalCloakStateBroken {
			t.Fatalf("broken cloak state=%q", state.CloakState)
		}
		tacticalStartShipTurnSpecials(spec, spec.Ships[0], state)
		if state.CloakState != TacticalCloakStateRecovering {
			t.Fatalf("start-turn cloak state=%q want recovering", state.CloakState)
		}
		if err := tacticalEndOwnPassSpecials(spec, &r, 1, false, 1); err != nil {
			t.Fatal(err)
		}
		if state.CloakState != TacticalCloakStateCloaked {
			t.Fatalf("end-pass cloak state=%q want cloaked", state.CloakState)
		}
	})

	t.Run("Phasing break recovery and lifetime", func(t *testing.T) {
		spec := slice178Checkpoint5SpecWithSpecial("phasing_cloak")
		r, err := newTacticalRuntime(spec)
		if err != nil {
			t.Fatal(err)
		}
		state := r.shipState(100)
		tacticalBreakCloakOnAction(state)
		if state.CloakState != TacticalCloakStatePhaseBroken || state.PhasingTurnsRemaining != 10 {
			t.Fatalf("broken phasing state=%q/%d", state.CloakState, state.PhasingTurnsRemaining)
		}
		tacticalStartShipTurnSpecials(spec, spec.Ships[0], state)
		if state.CloakState != TacticalCloakStatePhaseRecovering {
			t.Fatalf("start-turn phasing state=%q want phase_recovering", state.CloakState)
		}
		if err := tacticalEndOwnPassSpecials(spec, &r, 1, false, 1); err != nil {
			t.Fatal(err)
		}
		if state.CloakState != TacticalCloakStatePhased || state.PhasingTurnsRemaining != 10 {
			t.Fatalf("recovered phasing state=%q/%d want phased/10", state.CloakState, state.PhasingTurnsRemaining)
		}
		if err := tacticalEndOwnPassSpecials(spec, &r, 1, false, 2); err != nil {
			t.Fatal(err)
		}
		if state.CloakState != TacticalCloakStatePhased || state.PhasingTurnsRemaining != 9 {
			t.Fatalf("active phasing cleanup=%q/%d want phased/9", state.CloakState, state.PhasingTurnsRemaining)
		}
		state.PhasingTurnsRemaining = 1
		if err := tacticalEndOwnPassSpecials(spec, &r, 1, false, 3); err != nil {
			t.Fatal(err)
		}
		if state.CloakState != TacticalCloakStateCloaked || state.PhasingTurnsRemaining != 0 {
			t.Fatalf("expired phasing state=%q/%d want cloaked/0", state.CloakState, state.PhasingTurnsRemaining)
		}
	})
}

func TestSlice178Gate3Checkpoint5HyperXExtraVolleyAndRearm(t *testing.T) {
	spec := slice178Checkpoint5SpecWithSpecial("hyper_x_capacitors")
	r, err := newTacticalRuntime(spec)
	if err != nil {
		t.Fatal(err)
	}
	ship := spec.Ships[0]
	state := r.shipState(ship.ShipID)
	weapon := ship.Weapons[0]
	ready := weaponState(state, weapon.Slot)
	if ready == nil {
		t.Fatal("missing Beam readiness state")
	}
	if !tacticalCompleteWeaponFire(ship, state, weapon, ready) || !ready.Ready || ready.SpecialReadiness != TacticalSpecialReadinessNone {
		t.Fatalf("first Hyper-X volley state=%+v", *ready)
	}
	if tacticalCompleteWeaponFire(ship, state, weapon, ready) || ready.Ready {
		t.Fatalf("second Hyper-X volley did not exhaust ordinary readiness: %+v", *ready)
	}

	ready.Ready = true
	tacticalStartShipTurnSpecials(spec, ship, state)
	if ready.SpecialReadiness != TacticalSpecialReadinessRearmPending {
		t.Fatalf("next-turn Hyper-X state=%q want rearm_pending", ready.SpecialReadiness)
	}
	if err := tacticalEndOwnPassSpecials(spec, &r, 1, false, 1); err != nil {
		t.Fatal(err)
	}
	if ready.SpecialReadiness != TacticalSpecialReadinessExtraVolleyReady {
		t.Fatalf("non-firing end-pass Hyper-X state=%q want extra_volley_ready", ready.SpecialReadiness)
	}
}

func TestSlice178Gate3Checkpoint5AutomatedRepairOrder(t *testing.T) {
	spec := slice178Checkpoint5SpecWithSpecial("automated_repair_unit")
	ship := &spec.Ships[0]
	ship.ArmorMax = 10
	ship.StructureMax = 10
	ship.Internal = &TacticalInternalSpec{
		EngineMax: 5, SecondaryMax: 5,
		Weapons:            []TacticalInternalWeaponSpec{{Slot: 0, Max: 5, UnitSize: 1}},
		ShieldGeneratorMax: 5,
		Specials:           []TacticalInternalSpecialSpec{{Slot: 0, Max: 5}},
	}
	r, err := newTacticalRuntime(spec)
	if err != nil {
		t.Fatal(err)
	}
	state := r.shipState(ship.ShipID)
	state.StructureDamage = 5
	state.ArmorCurrent = 5
	state.Internal.EngineCurrent = 3
	state.Internal.SecondaryCurrent = 3
	state.Internal.Weapons[0].Current = 3
	state.Internal.ShieldGeneratorCurrent = 3

	if err := tacticalRepairShips(spec, &r, 1); err != nil {
		t.Fatal(err)
	}
	if state.StructureDamage != 1 || state.ArmorCurrent != 5 {
		t.Fatalf("durability repair structure/armor=%d/%d want=1/5", state.StructureDamage, state.ArmorCurrent)
	}
	if state.Internal.EngineCurrent != 5 || state.Internal.SecondaryCurrent != 3 || state.Internal.Weapons[0].Current != 3 || state.Internal.ShieldGeneratorCurrent != 3 {
		t.Fatalf("internal repair order state=%+v", state.Internal)
	}
}

func TestSlice178Gate3Checkpoint5VectorJTimeWarpScheduling(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Rules.Specials = slice178Checkpoint2SpecialRules()
	spec.Tactical.Ships[0].Specials = []TacticalSpecialSpec{{Slot: 0, SpecialID: "time_warp_facilitator", TacticalSupport: "tactical_passive"}}
	spec.Tactical.Ships[0].Internal = &TacticalInternalSpec{Specials: []TacticalInternalSpecialSpec{{Slot: 0, Max: 1}}}
	spec.Tactical.Ships[1].ShieldID = "class_i_shield"
	spec.Tactical.Ships[1].ShieldStrength = 4
	spec.Tactical.Ships[1].ShieldArcMax = 20
	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}

	// Damage a shield so the once-per-round recharge position is observable.
	defender := r.shipState(200)
	defender.ShieldArcCurrent[0] = 0

	first := r.shipState(100)
	first.ActivationComplete = true
	if err := advanceAfterActivationCompletion(spec, &r); err != nil {
		t.Fatal(err)
	}
	if r.state.Round != 1 || r.state.PassKind != TacticalPassKindNormal || r.state.ActiveShipID != 200 {
		t.Fatalf("after normal side A state=%+v", r.state)
	}

	defender.ActivationComplete = true
	if err := advanceAfterActivationCompletion(spec, &r); err != nil {
		t.Fatal(err)
	}
	if r.state.Round != 1 || r.state.PassKind != TacticalPassKindTimeWarp || r.state.TimeWarpSeatID != 1 || r.state.ActiveShipID != 100 {
		t.Fatalf("Time-Warp pass did not start after both normal sides: %+v", r.state)
	}
	if len(r.state.TimeWarpQueue) != 1 || r.state.TimeWarpQueue[0] != 100 {
		t.Fatalf("Time-Warp queue=%v want [100]", r.state.TimeWarpQueue)
	}
	if defender.ShieldArcCurrent[0] != 0 {
		t.Fatalf("shield recharged before Time-Warp pass completed: %v", defender.ShieldArcCurrent)
	}

	first.ActivationComplete = true
	if err := advanceAfterActivationCompletion(spec, &r); err != nil {
		t.Fatal(err)
	}
	if r.state.Round != 2 || r.state.PassKind != TacticalPassKindNormal || r.state.TimeWarpSeatID != 0 || len(r.state.TimeWarpQueue) != 0 {
		t.Fatalf("round did not finalize after Time-Warp passes: %+v", r.state)
	}
	if defender.ShieldArcCurrent[0] <= 0 {
		t.Fatalf("global shield recharge did not run after Time-Warp pass: %v", defender.ShieldArcCurrent)
	}

	var passes []TacticalEvent
	for _, event := range r.events {
		if event.Kind == "time_warp_pass_started" {
			passes = append(passes, event)
		}
	}
	if len(passes) != 2 || passes[0].SeatID != 1 || passes[1].SeatID != 2 {
		t.Fatalf("Time-Warp side-pass events=%+v", passes)
	}
}
