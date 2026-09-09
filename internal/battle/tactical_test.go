package battle

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func baselineTacticalBattleSpec() Spec {
	rules := TacticalRulesSnapshot{
		SchemaVersion:                1,
		InitiativeBeamOffenseDivisor: 10,
		RNGMultiplier:                0x41C64E6D,
		RNGIncrement:                 0x3039,
		BeamBaseHitThreshold:         40,
		BeamMaxHitThreshold:          95,
		BeamEffectiveRollCap:         100,
		BeamToHitRangeModifiers:      []int{0, 0, -10, -20, -30, -40, -55, -70, -85},
		BeamDamageRangeModifiers:     []int{0, 0, -10, -20, -30, -40, -50, -60, -65},
	}
	return Spec{
		ID:            1,
		GameID:        "tactical-test",
		StrategicTurn: 7,
		SystemID:      99,
		Attacker:      Side{EmpireID: 1, SeatID: 1, CombatFleetIDs: []core.ID{10}, ShipIDs: []core.ID{100}},
		Defender:      Side{EmpireID: 2, SeatID: 2, CombatFleetIDs: []core.ID{20}, ShipIDs: []core.ID{200}},
		Participants:  []protocol.SeatID{1, 2},
		Seed:          0xDEADBEEF,
		Tactical: &TacticalSpec{
			Rules:             rules,
			InitiativeEnabled: true,
			InitialRNGState:   BaselineTacticalRNGState,
			Ships: []TacticalShipSpec{
				{
					ShipID: 100, EmpireID: 1, SeatID: 1, X: 10, Y: 10,
					HullID: "frigate", WarpDriveID: "fusion_drive", ComputerID: "electronic_computer", ArmorID: "titanium_armor",
					Weapons:            []TacticalWeaponSpec{{Slot: 0, WeaponID: "laser_cannon", Count: 1, MinDamage: 1, MaxDamage: 4}},
					CurrentCombatSpeed: 22, BeamOffense: 25, BeamDefense: 0, ArmorMax: 4, StructureMax: 4,
				},
				{
					ShipID: 200, EmpireID: 2, SeatID: 2, X: 11, Y: 10,
					HullID: "frigate", WarpDriveID: "nuclear_drive", ComputerID: "electronic_computer", ArmorID: "titanium_armor",
					CurrentCombatSpeed: 20, BeamOffense: 25, BeamDefense: 0, ArmorMax: 4, StructureMax: 4,
				},
			},
		},
	}
}

func newStartedTacticalSession(t *testing.T) *Session {
	t.Helper()
	s, err := NewSession(baselineTacticalBattleSpec())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	return s
}

func submitPrepared(t *testing.T, s *Session, seat protocol.SeatID, command protocol.Command) {
	t.Helper()
	prepared, err := s.PrepareCommand(seat, command)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CommitPreparedCommand(prepared); err != nil {
		t.Fatal(err)
	}
}

func runGoldenTacticalBattle(t *testing.T) View {
	t.Helper()
	s := newStartedTacticalSession(t)
	fire1, _ := NewFireBeamCommand(1, FireBeamPayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	submitPrepared(t, s, 1, fire1)
	endAttacker, _ := NewEndActivationCommand(2, EndActivationPayload{ShipID: 100})
	submitPrepared(t, s, 1, endAttacker)
	endDefender, _ := NewEndActivationCommand(3, EndActivationPayload{ShipID: 200})
	submitPrepared(t, s, 2, endDefender)
	fire2, _ := NewFireBeamCommand(4, FireBeamPayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	submitPrepared(t, s, 1, fire2)
	return s.View()
}

func TestTacticalGoldenLaserBattle(t *testing.T) {
	s := newStartedTacticalSession(t)
	initial := s.View()
	if initial.Tactical == nil || initial.Tactical.State.Round != 1 || initial.Tactical.State.ActiveShipID != 100 || !reflect.DeepEqual(initial.Tactical.State.InitiativeOrder, []core.ID{100, 200}) {
		t.Fatalf("initial tactical state=%+v", initial.Tactical)
	}
	if initial.Tactical.State.NextCommandSequence != 1 || initial.Tactical.State.RNGState != BaselineTacticalRNGState {
		t.Fatalf("initial sequence/rng=%d/%08X", initial.Tactical.State.NextCommandSequence, initial.Tactical.State.RNGState)
	}

	fire1, _ := NewFireBeamCommand(1, FireBeamPayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	submitPrepared(t, s, 1, fire1)
	afterFirst := s.View()
	defender := afterFirst.Tactical.State.Ships[1]
	if defender.ShipID != 200 || defender.ArmorCurrent != 0 || defender.StructureDamage != 0 || defender.Destroyed {
		t.Fatalf("after first Laser defender=%+v", defender)
	}
	if afterFirst.Tactical.State.Ships[0].Weapons[0].Ready {
		t.Fatal("Laser remained ready after firing")
	}
	if afterFirst.Tactical.State.NextCommandSequence != 2 || len(afterFirst.Tactical.Events) != 3 {
		t.Fatalf("after first seq/events=%d/%d", afterFirst.Tactical.State.NextCommandSequence, len(afterFirst.Tactical.Events))
	}

	endAttacker, _ := NewEndActivationCommand(2, EndActivationPayload{ShipID: 100})
	submitPrepared(t, s, 1, endAttacker)
	if got := s.View().Tactical.State.ActiveShipID; got != 200 {
		t.Fatalf("defender not activated, got %d", got)
	}
	endDefender, _ := NewEndActivationCommand(3, EndActivationPayload{ShipID: 200})
	submitPrepared(t, s, 2, endDefender)
	round2 := s.View()
	if round2.Tactical.State.Round != 2 || round2.Tactical.State.ActiveShipID != 100 || !round2.Tactical.State.Ships[0].Weapons[0].Ready {
		t.Fatalf("round2 state=%+v", round2.Tactical.State)
	}

	fire2, _ := NewFireBeamCommand(4, FireBeamPayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	prepared, err := s.PrepareCommand(1, fire2)
	if err != nil {
		t.Fatal(err)
	}
	candidate := prepared.Result()
	if candidate == nil || candidate.WinnerSeat != 1 || candidate.Outcome != TacticalOutcomeVictory || !reflect.DeepEqual(candidate.DestroyedShipIDs, []core.ID{200}) {
		t.Fatalf("terminal candidate=%+v", candidate)
	}
	if err := s.CommitPreparedCommand(prepared); err != nil {
		t.Fatal(err)
	}
	finished := s.View()
	if finished.Phase != PhaseCompleted || finished.Result == nil || finished.Result.WinnerSeat != 1 || finished.Result.Outcome != TacticalOutcomeVictory {
		t.Fatalf("finished=%+v", finished)
	}
	if finished.Tactical.State.RNGState != 0xFD95EBB9 || finished.Tactical.State.NextCommandSequence != 5 {
		t.Fatalf("final rng/sequence=%08X/%d", finished.Tactical.State.RNGState, finished.Tactical.State.NextCommandSequence)
	}
	finalDefender := finished.Tactical.State.Ships[1]
	if finalDefender.ArmorCurrent != 0 || finalDefender.StructureDamage != 4 || !finalDefender.Destroyed {
		t.Fatalf("final defender=%+v", finalDefender)
	}
	wantKinds := []string{"round_started", "beam_fired", "battle_damage_applied", "activation_ended", "activation_ended", "round_started", "beam_fired", "battle_damage_applied", "ship_destroyed", "winner_determined"}
	gotKinds := make([]string, len(finished.Tactical.Events))
	for i := range finished.Tactical.Events {
		gotKinds[i] = finished.Tactical.Events[i].Kind
	}
	if !reflect.DeepEqual(gotKinds, wantKinds) {
		t.Fatalf("event kinds=%v want=%v", gotKinds, wantKinds)
	}
	var firstBeam, firstDamage, secondBeam, secondDamage map[string]any
	for payload, dst := range map[int]*map[string]any{1: &firstBeam, 2: &firstDamage, 6: &secondBeam, 7: &secondDamage} {
		if err := json.Unmarshal(finished.Tactical.Events[payload].Data, dst); err != nil {
			t.Fatal(err)
		}
	}
	if firstBeam["raw_hit_roll"] != float64(100) || firstDamage["selection_roll"] != float64(19) || secondBeam["raw_hit_roll"] != float64(100) || secondDamage["selection_roll"] != float64(100) {
		t.Fatalf("golden RNG event trace=%v/%v/%v/%v", firstBeam, firstDamage, secondBeam, secondDamage)
	}
	if firstDamage["layer"] != "armor" || secondDamage["layer"] != "structure" {
		t.Fatalf("golden damage layers=%v/%v", firstDamage, secondDamage)
	}
}

func TestTacticalReplayAndViewAreDeterministicDetached(t *testing.T) {
	first := runGoldenTacticalBattle(t)
	second := runGoldenTacticalBattle(t)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same tactical input diverged:\nfirst=%+v\nsecond=%+v", first, second)
	}

	s := newStartedTacticalSession(t)
	view := s.View()
	view.Spec.Tactical.Rules.BeamToHitRangeModifiers[0] = 999
	view.Spec.Tactical.Ships[0].Weapons[0].WeaponID = "mutated"
	view.Tactical.State.InitiativeOrder[0] = 999
	view.Tactical.State.Ships[0].Weapons[0].Ready = false
	view.Tactical.Events[0].Kind = "mutated"
	fresh := s.View()
	if fresh.Spec.Tactical.Rules.BeamToHitRangeModifiers[0] != 0 || fresh.Spec.Tactical.Ships[0].Weapons[0].WeaponID != "laser_cannon" || fresh.Tactical.State.InitiativeOrder[0] != 100 || !fresh.Tactical.State.Ships[0].Weapons[0].Ready || fresh.Tactical.Events[0].Kind != "round_started" {
		t.Fatalf("mutating View aliased authoritative battle: %+v", fresh)
	}
}

func TestTacticalRejectedCommandsDoNotMutate(t *testing.T) {
	tests := []struct {
		name string
		seat protocol.SeatID
		cmd  protocol.Command
	}{
		{name: "wrong seat", seat: 2, cmd: mustFireCommand(t, 1, 100, 200, 0)},
		{name: "wrong sequence", seat: 1, cmd: mustFireCommand(t, 2, 100, 200, 0)},
		{name: "wrong active ship", seat: 1, cmd: mustFireCommand(t, 1, 200, 100, 0)},
		{name: "friendly target", seat: 1, cmd: mustFireCommand(t, 1, 100, 100, 0)},
		{name: "wrong slot", seat: 1, cmd: mustFireCommand(t, 1, 100, 200, 1)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newStartedTacticalSession(t)
			before := s.View()
			if _, err := s.PrepareCommand(tc.seat, tc.cmd); err == nil {
				t.Fatal("expected command rejection")
			}
			after := s.View()
			if !reflect.DeepEqual(before, after) {
				t.Fatalf("rejected command mutated battle:\nbefore=%+v\nafter=%+v", before, after)
			}
		})
	}

	s := newStartedTacticalSession(t)
	fire := mustFireCommand(t, 1, 100, 200, 0)
	submitPrepared(t, s, 1, fire)
	before := s.View()
	doubleFire := mustFireCommand(t, 2, 100, 200, 0)
	if _, err := s.PrepareCommand(1, doubleFire); err == nil {
		t.Fatal("expected double fire to reject")
	}
	if after := s.View(); !reflect.DeepEqual(before, after) {
		t.Fatal("double-fire rejection mutated battle")
	}
}

func TestUnsupportedInternalSelectionRejectsAtomically(t *testing.T) {
	s := newStartedTacticalSession(t)
	fire := mustFireCommand(t, 1, 100, 200, 0)
	submitPrepared(t, s, 1, fire)

	// Test-only setup: return the attacker Laser to ready and choose a deterministic
	// RNG state whose next shot hits but whose following Select_Internal roll is not
	// 100. The public PrepareCommand must still reject without mutating this state.
	s.mu.Lock()
	s.tactical.state.Ships[0].Weapons[0].Ready = true
	chosen := uint32(0)
	for seed := uint32(0); seed < 100000; seed++ {
		probe := cloneTacticalRuntime(*s.tactical)
		probe.state.RNGState = seed
		raw, _ := probe.random(100, s.spec.Tactical.Rules)
		selection, _ := probe.random(100, s.spec.Tactical.Rules)
		effective := raw
		if raw > 95 {
			effective = 100
		} else {
			effective += 25
			if effective > 100 {
				effective = 100
			}
		}
		if effective >= 40 && selection != 100 {
			chosen = seed
			break
		}
	}
	if chosen == 0 {
		s.mu.Unlock()
		t.Fatal("failed to find unsupported internal-selection seed")
	}
	s.tactical.state.RNGState = chosen
	s.mu.Unlock()

	before := s.View()
	second := mustFireCommand(t, 2, 100, 200, 0)
	_, err := s.PrepareCommand(1, second)
	if err == nil || !strings.Contains(err.Error(), "unsupported tactical internal subsystem selection") {
		t.Fatalf("unexpected unsupported-internal error: %v", err)
	}
	after := s.View()
	if !reflect.DeepEqual(before, after) {
		t.Fatal("unsupported internal damage mutated authoritative battle")
	}
}

func TestTacticalEnabledBattleRejectsManualResultAndUnsupportedLifecycleRemainsManual(t *testing.T) {
	s := newStartedTacticalSession(t)
	before := s.View()
	if err := s.Complete(Result{WinnerSeat: 1, Outcome: "manual"}); err == nil {
		t.Fatal("expected tactical-enabled manual result to reject")
	}
	if !reflect.DeepEqual(before, s.View()) {
		t.Fatal("manual result rejection mutated tactical battle")
	}

	spec := baselineTacticalBattleSpec()
	spec.Tactical = nil
	spec.TacticalUnsupportedReason = "fixture unsupported"
	legacy, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := legacy.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := legacy.PrepareCommand(1, mustFireCommand(t, 1, 100, 200, 0)); err == nil || !strings.Contains(err.Error(), "fixture unsupported") {
		t.Fatalf("unsupported tactical command error=%v", err)
	}
	if err := legacy.Complete(Result{WinnerSeat: 1, Outcome: "manual_victory", DestroyedShipIDs: []core.ID{200}}); err != nil {
		t.Fatalf("lifecycle-only battle rejected manual result: %v", err)
	}
}

func mustFireCommand(t *testing.T, sequence uint32, shipID, targetID core.ID, slot int) protocol.Command {
	t.Helper()
	cmd, err := NewFireBeamCommand(sequence, FireBeamPayload{ShipID: shipID, TargetShipID: targetID, WeaponSlot: slot})
	if err != nil {
		t.Fatal(err)
	}
	return cmd
}

func TestTacticalMoveCostFacingAndTurningModes(t *testing.T) {
	cost, facing := tacticalMoveCost(10, 10, 0, 13, 10, 2)
	if cost != 3 || facing != 0 {
		t.Fatalf("straight move cost/facing=%d/%d want 3/0", cost, facing)
	}
	cost, facing = tacticalMoveCost(10, 10, 0, 10, 13, 2)
	if cost != 11 || facing != 4 {
		t.Fatalf("normal quarter-turn cost/facing=%d/%d want 11/4", cost, facing)
	}
	cost, facing = tacticalMoveCost(10, 10, 0, 10, 13, 1)
	if cost != 7 || facing != 4 {
		t.Fatalf("stabilizer quarter-turn cost/facing=%d/%d want 7/4", cost, facing)
	}
	cost, facing = tacticalMoveCost(10, 10, 0, 10, 13, 0)
	if cost != 3 || facing != 4 {
		t.Fatalf("nullifier quarter-turn cost/facing=%d/%d want 3/4", cost, facing)
	}
	cost, facing = tacticalMoveCost(10, 10, 15, 13, 9, 2)
	if facing != 15 || cost != 4 {
		t.Fatalf("diagonal move cost/facing=%d/%d want 4/15", cost, facing)
	}
}

func TestTacticalMoveCommandProjectsAndCommitsAuthoritatively(t *testing.T) {
	s := newStartedTacticalSession(t)
	initial := s.View()
	if initial.Tactical == nil || !initial.Tactical.CanEndActivation {
		t.Fatalf("initial tactical view=%+v", initial.Tactical)
	}
	var straight *TacticalMoveOption
	for i := range initial.Tactical.LegalMoves {
		move := &initial.Tactical.LegalMoves[i]
		if move.X == 13 && move.Y == 10 {
			straight = move
			break
		}
	}
	if straight == nil || straight.MoveCost != 3 || straight.ResultingFacing != 0 || straight.MovementRemainingAfter != 19 {
		t.Fatalf("straight legal move=%+v", straight)
	}
	for _, move := range initial.Tactical.LegalMoves {
		if move.X == 11 && move.Y == 10 {
			t.Fatal("occupied defender coordinate was projected as legal")
		}
	}
	beforeRNG := initial.Tactical.State.RNGState
	move := mustMoveCommand(t, 1, 100, 13, 10)
	submitPrepared(t, s, 1, move)
	after := s.View()
	ship := after.Tactical.State.Ships[0]
	if ship.X != 13 || ship.Y != 10 || ship.Facing != 0 || ship.MovementCurrent != 19 || ship.MovementMax != 22 {
		t.Fatalf("moved ship=%+v", ship)
	}
	if after.Tactical.State.RNGState != beforeRNG {
		t.Fatalf("movement consumed RNG: before=%08X after=%08X", beforeRNG, after.Tactical.State.RNGState)
	}
	if after.Tactical.State.NextCommandSequence != 2 || after.Tactical.Events[len(after.Tactical.Events)-1].Kind != "ship_moved" {
		t.Fatalf("movement sequence/events=%d/%v", after.Tactical.State.NextCommandSequence, after.Tactical.Events)
	}
}

func TestTacticalRejectedMovesDoNotMutate(t *testing.T) {
	tests := []struct {
		name string
		seat protocol.SeatID
		cmd  protocol.Command
	}{
		{name: "wrong seat", seat: 2, cmd: mustMoveCommand(t, 1, 100, 13, 10)},
		{name: "wrong sequence", seat: 1, cmd: mustMoveCommand(t, 2, 100, 13, 10)},
		{name: "wrong active ship", seat: 1, cmd: mustMoveCommand(t, 1, 200, 13, 10)},
		{name: "same coordinate", seat: 1, cmd: mustMoveCommand(t, 1, 100, 10, 10)},
		{name: "occupied", seat: 1, cmd: mustMoveCommand(t, 1, 100, 11, 10)},
		{name: "over budget", seat: 1, cmd: mustMoveCommand(t, 1, 100, 100, 10)},
		{name: "technical envelope", seat: 1, cmd: mustMoveCommand(t, 1, 100, TacticalCoordinateSafetyLimit+1, 10)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newStartedTacticalSession(t)
			before := s.View()
			if _, err := s.PrepareCommand(tc.seat, tc.cmd); err == nil {
				t.Fatal("expected move rejection")
			}
			if after := s.View(); !reflect.DeepEqual(before, after) {
				t.Fatalf("rejected move mutated battle: before=%+v after=%+v", before, after)
			}
		})
	}
}

func TestTacticalTwoByTwoActivationAndRoundReset(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	spec.Attacker.ShipIDs = []core.ID{100, 101}
	spec.Defender.ShipIDs = []core.ID{200, 201}
	attackerLead := spec.Tactical.Ships[0]
	defenderLead := spec.Tactical.Ships[1]
	attackerWing := defenderLead
	attackerWing.ShipID = 101
	attackerWing.EmpireID = 1
	attackerWing.SeatID = 1
	attackerWing.X = 10
	attackerWing.Y = 12
	attackerWing.Facing = 0
	defenderLead.X = 20
	defenderLead.Y = 10
	defenderLead.Facing = 8
	defenderWing := attackerLead
	defenderWing.ShipID = 201
	defenderWing.EmpireID = 2
	defenderWing.SeatID = 2
	defenderWing.X = 20
	defenderWing.Y = 12
	defenderWing.Facing = 8
	spec.Tactical.Ships = []TacticalShipSpec{attackerLead, attackerWing, defenderLead, defenderWing}
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	view := s.View()
	wantOrder := []core.ID{100, 201, 101, 200}
	if !reflect.DeepEqual(view.Tactical.State.InitiativeOrder, wantOrder) || view.Tactical.State.ActiveShipID != 100 {
		t.Fatalf("2v2 initiative=%v active=%d", view.Tactical.State.InitiativeOrder, view.Tactical.State.ActiveShipID)
	}
	sequence := uint32(1)
	for _, step := range []struct {
		seat       protocol.SeatID
		ship, next core.ID
	}{{1, 100, 201}, {2, 201, 101}, {1, 101, 200}, {2, 200, 100}} {
		cmd, _ := NewEndActivationCommand(sequence, EndActivationPayload{ShipID: step.ship})
		submitPrepared(t, s, step.seat, cmd)
		sequence++
		if got := s.View().Tactical.State.ActiveShipID; got != step.next {
			t.Fatalf("after ship %d active=%d want %d", step.ship, got, step.next)
		}
	}
	round2 := s.View().Tactical.State
	if round2.Round != 2 {
		t.Fatalf("round=%d want 2", round2.Round)
	}
	for _, ship := range round2.Ships {
		if !ship.Destroyed && (ship.ActivationComplete || ship.MovementCurrent != ship.MovementMax) {
			t.Fatalf("round reset ship=%+v", ship)
		}
	}
}

func TestTacticalTwoByTwoRejectsThirdShipPerSide(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	spec.Attacker.ShipIDs = []core.ID{100, 101, 102}
	if _, err := NewSession(spec); err == nil || !strings.Contains(err.Error(), "one or two combat Ships per side") {
		t.Fatalf("unexpected 3-ship validation error: %v", err)
	}
}

func mustMoveCommand(t *testing.T, sequence uint32, shipID core.ID, x, y int) protocol.Command {
	t.Helper()
	cmd, err := NewMoveShipCommand(sequence, MoveShipPayload{ShipID: shipID, X: x, Y: y})
	if err != nil {
		t.Fatal(err)
	}
	return cmd
}
