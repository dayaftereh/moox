package battle

import "testing"

func slice178Checkpoint4BattleSpec() Spec {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Rules.Boarding = &TacticalBoardingRulesSnapshot{
		PhysicalRange:                1,
		TransporterRange:             4,
		DamperTransporterDivisor:     3,
		SecurityStationsDefenseBonus: 20,
		RaidInternalDamageThreshold:  20,
		RaidDefenderLossThreshold:    50,
		RaidAttackerLossThreshold:    -50,
		CaptureQuantumRandomBound:    10,
		CaptureQuantumDestroyMax:     5,
		CrewQualityBonuses:           []int{0, 5, 10, 15},
	}
	spec.Tactical.BoardingDoctrines = []TacticalBoardingDoctrineSpec{
		{SeatID: 1, EmpireID: 1, GroundCombatBonus: 1000},
		{SeatID: 2, EmpireID: 2},
	}
	spec.Tactical.Ships[0].MarineMax = 10
	spec.Tactical.Ships[1].MarineMax = 1
	spec.Tactical.Ships[1].CurrentCombatSpeed = 0
	return spec
}

func slice178Checkpoint4QuantumSeed(t *testing.T, rules TacticalRulesSnapshot, want int) uint32 {
	t.Helper()
	for seed := uint32(1); seed < 100000; seed++ {
		r := tacticalRuntime{state: TacticalState{RNGState: seed}}
		if _, err := r.random(100, rules); err != nil {
			t.Fatal(err)
		}
		if _, err := r.random(100, rules); err != nil {
			t.Fatal(err)
		}
		got, err := r.random(10, rules)
		if err != nil {
			t.Fatal(err)
		}
		if got == want {
			return seed
		}
	}
	t.Fatalf("no deterministic seed found for third Random(10)=%d", want)
	return 0
}

func TestSlice178Gate3Checkpoint4VectorCRaidRestoresMarinesAndKeepsInternalDamage(t *testing.T) {
	spec := slice178Checkpoint4BattleSpec()
	spec.Tactical.Ships[1].Internal = &TacticalInternalSpec{EngineMax: 1, SecondaryMax: 1}
	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	source := spec.Tactical.Ships[0]
	target := spec.Tactical.Ships[1]
	sourceState := r.shipState(source.ShipID)
	targetState := r.shipState(target.ShipID)
	sourceState.MarineCurrent = 6 // selected party 4 is already dispatched
	targetState.MarineCurrent = 1
	beforeEngine := targetState.Internal.EngineCurrent
	beforeSecondary := targetState.Internal.SecondaryCurrent

	if err := resolveTacticalRaid(*spec.Tactical, &r, 1, 1, source, sourceState, target, targetState, 4, 4, TacticalBoardingDeliveryPhysical); err != nil {
		t.Fatal(err)
	}
	if sourceState.MarineCurrent != 10 {
		t.Fatalf("Raid source Marines=%d want=10", sourceState.MarineCurrent)
	}
	if targetState.MarineCurrent != 1 {
		t.Fatalf("Raid target persistent Marines=%d want=1", targetState.MarineCurrent)
	}
	if targetState.Internal.EngineCurrent == beforeEngine && targetState.Internal.SecondaryCurrent == beforeSecondary {
		t.Fatalf("Raid did not persist any internal damage: before engine/secondary=%d/%d after=%d/%d",
			beforeEngine, beforeSecondary, targetState.Internal.EngineCurrent, targetState.Internal.SecondaryCurrent)
	}
}

func TestSlice178Gate3Checkpoint4CaptureTransfersAuthorityAndReconcilesActors(t *testing.T) {
	spec := slice178Checkpoint4BattleSpec()
	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	source := spec.Tactical.Ships[0]
	target := spec.Tactical.Ships[1]
	sourceState := r.shipState(source.ShipID)
	targetState := r.shipState(target.ShipID)
	sourceState.MarineCurrent = 7 // selected party 3 is already dispatched
	targetState.MarineCurrent = 1
	targetState.TeleporterUsed = true
	targetState.AlreadyBoarded = true
	targetState.BeingCaptured = true

	r.ordnance.Projectiles = []TacticalProjectileState{{
		ProjectileID: 1, SourceSeatID: 1, SourceShipID: source.ShipID, TargetShipID: target.ShipID,
	}}
	r.fighters.Groups = []TacticalFighterGroupState{{
		GroupID: 1, SourceSeatID: 1, SourceShipID: source.ShipID, TargetShipID: target.ShipID,
		Multiplicity: 4, Phase: TacticalFighterPhaseOutbound,
	}}

	if err := resolveTacticalCapture(*spec.Tactical, &r, 1, 2, source, sourceState, target, targetState, 3, TacticalBoardingDeliveryPhysical); err != nil {
		t.Fatal(err)
	}
	if targetState.OwnerEmpireID != 1 || targetState.ControllerSeatID != 1 {
		t.Fatalf("captured authority empire/seat=%d/%d want=1/1", targetState.OwnerEmpireID, targetState.ControllerSeatID)
	}
	if targetState.MarineCurrent != 3 {
		t.Fatalf("captured Marines=%d want=3", targetState.MarineCurrent)
	}
	if !targetState.PostCaptureState {
		t.Fatal("non-Telepathic captured ship did not enter post-capture state")
	}
	if targetState.TeleporterUsed || targetState.AlreadyBoarded || targetState.BeingCaptured {
		t.Fatalf("captured Boarding transients were not cleared: %+v", *targetState)
	}
	if len(r.ordnance.Projectiles) != 0 {
		t.Fatalf("friendly inbound projectile survived capture reconciliation: %+v", r.ordnance.Projectiles)
	}
	if len(r.fighters.Groups) != 1 || r.fighters.Groups[0].Phase != TacticalFighterPhaseReturning || r.fighters.Groups[0].TargetShipID != source.ShipID {
		t.Fatalf("Fighter capture reconciliation=%+v", r.fighters.Groups)
	}
	if r.fighters.Groups[0].SourceSeatID != 1 {
		t.Fatalf("launched Fighter source seat mutated=%d", r.fighters.Groups[0].SourceSeatID)
	}
}

func TestSlice178Gate3Checkpoint4VectorDQuantumThreshold(t *testing.T) {
	for _, tc := range []struct {
		roll          int
		wantDestroyed bool
		wantOwner     int
	}{
		{roll: 5, wantDestroyed: true, wantOwner: 2},
		{roll: 6, wantDestroyed: false, wantOwner: 1},
	} {
		t.Run(string(rune('0'+tc.roll)), func(t *testing.T) {
			spec := slice178Checkpoint4BattleSpec()
			spec.Tactical.Ships[1].Specials = []TacticalSpecialSpec{{
				Slot: 0, SpecialID: "quantum_detonator", TacticalSupport: "tactical_passive",
			}}
			spec.Tactical.Ships[1].Internal = &TacticalInternalSpec{
				Specials: []TacticalInternalSpecialSpec{{Slot: 0, Max: 1}},
			}
			r, err := newTacticalRuntime(*spec.Tactical)
			if err != nil {
				t.Fatal(err)
			}
			r.state.RNGState = slice178Checkpoint4QuantumSeed(t, spec.Tactical.Rules, tc.roll)
			source := spec.Tactical.Ships[0]
			target := spec.Tactical.Ships[1]
			sourceState := r.shipState(source.ShipID)
			targetState := r.shipState(target.ShipID)
			sourceState.MarineCurrent = 7
			targetState.MarineCurrent = 1
			if err := resolveTacticalCapture(*spec.Tactical, &r, 1, 3, source, sourceState, target, targetState, 3, TacticalBoardingDeliveryPhysical); err != nil {
				t.Fatal(err)
			}
			if targetState.Destroyed != tc.wantDestroyed || int(targetState.OwnerEmpireID) != tc.wantOwner {
				t.Fatalf("Quantum roll %d destroyed/owner=%v/%d want=%v/%d", tc.roll, targetState.Destroyed, targetState.OwnerEmpireID, tc.wantDestroyed, tc.wantOwner)
			}
		})
	}
}

func TestSlice178Gate3Checkpoint4MutableAuthorityPersistsAndControlsView(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	captured := s.tactical.shipState(200)
	captured.OwnerEmpireID = 1
	captured.ControllerSeatID = 1
	s.tactical.state.ActiveShipID = 200

	seat1, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	seat2, err := s.PlayerView(2)
	if err != nil {
		t.Fatal(err)
	}
	if seat1.Tactical == nil || !seat1.Tactical.CanEndActivation {
		t.Fatalf("new controller lacks captured-ship action surface: %+v", seat1.Tactical)
	}
	if seat2.Tactical == nil || seat2.Tactical.CanEndActivation || len(seat2.Tactical.LegalMoves) != 0 || len(seat2.Tactical.BoardingActions) != 0 {
		t.Fatalf("old controller retained captured-ship action surface: %+v", seat2.Tactical)
	}
	if got := remainingWinner(spec, s.tactical); got != 1 {
		t.Fatalf("remaining winner after ownership convergence=%d want=1", got)
	}
	if err := validateLiveSnapshot(s.LiveSnapshot()); err != nil {
		t.Fatalf("mutable authority live snapshot rejected: %v", err)
	}
}

func TestSlice178Gate3Checkpoint4WarpDissipatorRetreatLegality(t *testing.T) {
	spec := slice178Checkpoint4BattleSpec()
	spec.Tactical.Ships[1].Specials = []TacticalSpecialSpec{{
		Slot: 0, SpecialID: "warp_dissipator", TacticalSupport: "tactical_passive",
	}}
	spec.Tactical.Ships[1].Internal = &TacticalInternalSpec{
		Specials: []TacticalInternalSpecialSpec{{Slot: 0, Max: 1}},
	}
	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	if !tacticalRetreatBlockedByWarpDissipator(*spec.Tactical, &r, 1) {
		t.Fatal("functional enemy Warp Dissipator did not block retreat")
	}
	if _, err := prepareRetreat(spec, &r, 1, 1, RetreatPayload{ShipID: 100}); err == nil {
		t.Fatal("battle.retreat accepted despite functional enemy Warp Dissipator")
	}

	carrier := r.shipState(200)
	carrier.PostCaptureState = true
	if tacticalRetreatBlockedByWarpDissipator(*spec.Tactical, &r, 1) {
		t.Fatal("post-capture-invalid Warp Dissipator still blocks retreat")
	}
	carrier.PostCaptureState = false
	carrier.OwnerEmpireID = 1
	carrier.ControllerSeatID = 1
	if tacticalRetreatBlockedByWarpDissipator(*spec.Tactical, &r, 1) {
		t.Fatal("friendly captured Warp Dissipator still blocks retreat")
	}
}
