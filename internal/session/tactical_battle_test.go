package session

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

type tacticalStrategicFixture struct {
	attackerEmpireID core.ID
	defenderEmpireID core.ID
	attackerShipIDs  []core.ID
	defenderShipIDs  []core.ID
	attackerFleetIDs []core.ID
	defenderFleetIDs []core.ID
}

func loadTacticalEconomyResolver(t *testing.T) *game.EconomyResolver {
	t.Helper()
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	return resolver
}

func makeTacticalStrategicFixture(t *testing.T, systemCount int) (*core.GameState, []Seat, tacticalStrategicFixture) {
	t.Helper()
	if systemCount < 1 || systemCount > 2 {
		t.Fatalf("unsupported test system count %d", systemCount)
	}
	state, seats := twoSeatFixture(t)
	attackerEmpireID := state.Empires[0].ID
	defenderEmpireID := state.Empires[1].ID
	attackerSpec := core.ShipDesignSpec{
		HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "fusion_drive", FTLSpeed: 3,
		ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4,
		HullBaseCostPP: 20, HullSpace: 25, SpaceUsed: 10, BaseDesignCostPP: 30, ProductionCostPP: 30,
		Weapons: []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}},
	}
	defenderSpec := core.ShipDesignSpec{
		HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "nuclear_drive", FTLSpeed: 2,
		ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4,
		HullBaseCostPP: 20, HullSpace: 25, SpaceUsed: 10, BaseDesignCostPP: 25, ProductionCostPP: 25,
	}
	attackerDesignID := state.NewID()
	defenderDesignID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns,
		core.ShipDesign{ID: attackerDesignID, EmpireID: attackerEmpireID, Revision: 1, Name: "Fusion Laser", Spec: attackerSpec},
		core.ShipDesign{ID: defenderDesignID, EmpireID: defenderEmpireID, Revision: 1, Name: "Nuclear Target", Spec: defenderSpec},
	)
	fixture := tacticalStrategicFixture{attackerEmpireID: attackerEmpireID, defenderEmpireID: defenderEmpireID}
	for i := 0; i < systemCount; i++ {
		attackerShipID := state.NewID()
		defenderShipID := state.NewID()
		state.Ships = append(state.Ships,
			core.Ship{ID: attackerShipID, EmpireID: attackerEmpireID, SourceDesignID: attackerDesignID, SourceDesignRevision: 1, Name: "Fusion Laser", Spec: attackerSpec},
			core.Ship{ID: defenderShipID, EmpireID: defenderEmpireID, SourceDesignID: defenderDesignID, SourceDesignRevision: 1, Name: "Nuclear Target", Spec: defenderSpec},
		)
		attackerFleetID := state.NewID()
		defenderFleetID := state.NewID()
		systemID := state.Galaxy.Systems[i].ID
		state.StrategicFleets = append(state.StrategicFleets,
			core.StrategicFleet{ID: attackerFleetID, EmpireID: attackerEmpireID, Role: core.StrategicFleetRoleCombat, AtSystemID: systemID, ShipIDs: []core.ID{attackerShipID}},
			core.StrategicFleet{ID: defenderFleetID, EmpireID: defenderEmpireID, Role: core.StrategicFleetRoleCombat, AtSystemID: systemID, ShipIDs: []core.ID{defenderShipID}},
		)
		fixture.attackerShipIDs = append(fixture.attackerShipIDs, attackerShipID)
		fixture.defenderShipIDs = append(fixture.defenderShipIDs, defenderShipID)
		fixture.attackerFleetIDs = append(fixture.attackerFleetIDs, attackerFleetID)
		fixture.defenderFleetIDs = append(fixture.defenderFleetIDs, defenderFleetID)
	}
	state.DiplomaticRelations = reciprocalWarRelations(attackerEmpireID, defenderEmpireID)
	if err := state.Validate(); err != nil {
		t.Fatalf("tactical strategic fixture invalid: %v", err)
	}
	return state, seats, fixture
}

func submitGoldenBattleCommands(t *testing.T, s *GameSession, view battle.View) error {
	t.Helper()
	attackerShipID := view.Spec.Attacker.ShipIDs[0]
	defenderShipID := view.Spec.Defender.ShipIDs[0]
	fire1, err := battle.NewFireBeamCommand(1, battle.FireBeamPayload{ShipID: attackerShipID, TargetShipID: defenderShipID, WeaponSlot: 0})
	if err != nil {
		return err
	}
	if err := s.SubmitBattleCommand(view.Spec.ID, view.Spec.Attacker.SeatID, fire1); err != nil {
		return err
	}
	endAttacker, err := battle.NewEndActivationCommand(2, battle.EndActivationPayload{ShipID: attackerShipID})
	if err != nil {
		return err
	}
	if err := s.SubmitBattleCommand(view.Spec.ID, view.Spec.Attacker.SeatID, endAttacker); err != nil {
		return err
	}
	endDefender, err := battle.NewEndActivationCommand(3, battle.EndActivationPayload{ShipID: defenderShipID})
	if err != nil {
		return err
	}
	if err := s.SubmitBattleCommand(view.Spec.ID, view.Spec.Defender.SeatID, endDefender); err != nil {
		return err
	}
	fire2, err := battle.NewFireBeamCommand(4, battle.FireBeamPayload{ShipID: attackerShipID, TargetShipID: defenderShipID, WeaponSlot: 0})
	if err != nil {
		return err
	}
	return s.SubmitBattleCommand(view.Spec.ID, view.Spec.Attacker.SeatID, fire2)
}

func TestEconomyResolverTacticalBattleCompletesThroughStrategicHandoff(t *testing.T) {
	state, seats, fixture := makeTacticalStrategicFixture(t, 1)
	resolver := loadTacticalEconomyResolver(t)
	s, err := NewGameSession("tactical-strategic", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "tactical-strategic")
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	started, _ := s.ObserverView()
	if started.Phase != PhaseEncounters || len(started.Battles) != 1 {
		t.Fatalf("started tactical encounter=%+v", started)
	}
	child := started.Battles[0]
	if child.Spec.Tactical == nil || child.Spec.TacticalUnsupportedReason != "" || child.Tactical == nil {
		t.Fatalf("tactical fixture not attached: %+v", child)
	}
	if child.Spec.Tactical.InitialRNGState != battle.BaselineTacticalRNGState || child.Tactical.State.ActiveShipID != fixture.attackerShipIDs[0] {
		t.Fatalf("tactical start=%+v", child.Tactical)
	}
	if err := s.CompleteBattle(child.Spec.ID, battle.Result{WinnerSeat: child.Spec.Attacker.SeatID, Outcome: "manual"}); err == nil {
		t.Fatal("expected manual result injection to tactical battle to reject")
	}
	if err := submitGoldenBattleCommands(t, s, child); err != nil {
		t.Fatal(err)
	}
	finished, _ := s.ObserverView()
	if finished.Phase != PhasePostResolution || len(finished.Battles) != 1 || finished.Battles[0].Phase != battle.PhaseCompleted {
		t.Fatalf("finished tactical session=%+v", finished)
	}
	if finished.Battles[0].Result == nil || finished.Battles[0].Result.WinnerSeat != 1 || finished.Battles[0].Result.Outcome != battle.TacticalOutcomeVictory || !reflect.DeepEqual(finished.Battles[0].Result.DestroyedShipIDs, fixture.defenderShipIDs) {
		t.Fatalf("tactical result=%+v", finished.Battles[0].Result)
	}
	if finished.Battles[0].Tactical == nil || finished.Battles[0].Tactical.State.RNGState != 0xFD95EBB9 || len(finished.Battles[0].Tactical.Events) != 10 {
		t.Fatalf("completed tactical view=%+v", finished.Battles[0].Tactical)
	}
	for _, ship := range finished.State.Ships {
		if ship.ID == fixture.defenderShipIDs[0] {
			t.Fatalf("destroyed defender Ship %d remains in strategic state", ship.ID)
		}
	}
	for _, fleet := range finished.State.StrategicFleets {
		if fleet.ID == fixture.defenderFleetIDs[0] {
			t.Fatalf("empty destroyed defender Fleet %d remains in strategic state", fleet.ID)
		}
	}
	foundAttacker := false
	for _, ship := range finished.State.Ships {
		if ship.ID == fixture.attackerShipIDs[0] {
			foundAttacker = true
		}
	}
	if !foundAttacker {
		t.Fatal("winning attacker Ship disappeared")
	}
}

type failOnceTacticalResolver struct {
	inner      *game.EconomyResolver
	resumeErr  error
	resumeCall int
}

func (r *failOnceTacticalResolver) Resolve(ctx game.ResolveContext, state *core.GameState, batches []protocol.CommandBatch) (game.Resolution, error) {
	return r.inner.Resolve(ctx, state, batches)
}

func (r *failOnceTacticalResolver) ResumeAfterEncounters(ctx game.ResolveContext, state *core.GameState, outcomes []game.EncounterOutcome) (game.Resolution, error) {
	r.resumeCall++
	if r.resumeErr != nil {
		return game.Resolution{}, r.resumeErr
	}
	return r.inner.ResumeAfterEncounters(ctx, state, outcomes)
}

func TestTerminalTacticalCommandStrategicContinuationFailureIsAtomicAndRetryable(t *testing.T) {
	state, seats, fixture := makeTacticalStrategicFixture(t, 1)
	resolver := &failOnceTacticalResolver{inner: loadTacticalEconomyResolver(t), resumeErr: errors.New("resume rejected")}
	s, err := NewGameSession("tactical-retry", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "tactical-retry")
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	started, _ := s.ObserverView()
	child := started.Battles[0]
	attackerShipID, defenderShipID := child.Spec.Attacker.ShipIDs[0], child.Spec.Defender.ShipIDs[0]
	fire1, _ := battle.NewFireBeamCommand(1, battle.FireBeamPayload{ShipID: attackerShipID, TargetShipID: defenderShipID, WeaponSlot: 0})
	endA, _ := battle.NewEndActivationCommand(2, battle.EndActivationPayload{ShipID: attackerShipID})
	endD, _ := battle.NewEndActivationCommand(3, battle.EndActivationPayload{ShipID: defenderShipID})
	for _, step := range []struct {
		seat protocol.SeatID
		cmd  protocol.Command
	}{{child.Spec.Attacker.SeatID, fire1}, {child.Spec.Attacker.SeatID, endA}, {child.Spec.Defender.SeatID, endD}} {
		if err := s.SubmitBattleCommand(child.Spec.ID, step.seat, step.cmd); err != nil {
			t.Fatal(err)
		}
	}
	before, _ := s.ObserverView()
	fire2, _ := battle.NewFireBeamCommand(4, battle.FireBeamPayload{ShipID: attackerShipID, TargetShipID: defenderShipID, WeaponSlot: 0})
	if err := s.SubmitBattleCommand(child.Spec.ID, child.Spec.Attacker.SeatID, fire2); err == nil {
		t.Fatal("expected terminal tactical continuation failure")
	}
	afterFailure, _ := s.ObserverView()
	if afterFailure.Revision != before.Revision || afterFailure.Phase != PhaseEncounters || !reflect.DeepEqual(afterFailure.State, before.State) || !reflect.DeepEqual(afterFailure.Battles, before.Battles) {
		t.Fatalf("failed terminal command mutated session:\nbefore=%+v\nafter=%+v", before, afterFailure)
	}
	resolver.resumeErr = nil
	if err := s.SubmitBattleCommand(child.Spec.ID, child.Spec.Attacker.SeatID, fire2); err != nil {
		t.Fatal(err)
	}
	finished, _ := s.ObserverView()
	if finished.Phase != PhasePostResolution || finished.Battles[0].Result == nil || !reflect.DeepEqual(finished.Battles[0].Result.DestroyedShipIDs, fixture.defenderShipIDs) || resolver.resumeCall != 2 {
		t.Fatalf("retry did not complete cleanly: phase=%s result=%+v calls=%d", finished.Phase, finished.Battles[0].Result, resolver.resumeCall)
	}
}

type armingEconomyResolver struct {
	inner          *game.EconomyResolver
	attackerShipID core.ID
}

func (r *armingEconomyResolver) Resolve(ctx game.ResolveContext, state *core.GameState, batches []protocol.CommandBatch) (game.Resolution, error) {
	for i := range state.Ships {
		if state.Ships[i].ID == r.attackerShipID {
			state.Ships[i].Spec.WarpDriveID = "fusion_drive"
			state.Ships[i].Spec.FTLSpeed = 3
			state.Ships[i].Spec.SpaceUsed = 10
			state.Ships[i].Spec.BaseDesignCostPP = 30
			state.Ships[i].Spec.ProductionCostPP = 30
			state.Ships[i].Spec.Weapons = []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}}
			break
		}
	}
	return r.inner.Resolve(ctx, state, batches)
}

func (r *armingEconomyResolver) ResumeAfterEncounters(ctx game.ResolveContext, state *core.GameState, outcomes []game.EncounterOutcome) (game.Resolution, error) {
	return r.inner.ResumeAfterEncounters(ctx, state, outcomes)
}

func TestTacticalEncounterPreparationUsesReturnedPreEncounterState(t *testing.T) {
	state, seats, fixture := makeTacticalStrategicFixture(t, 1)
	// Make the authoritative pre-resolution Ship deliberately unsupported. The
	// resolver arms/upgrades only its private resolution-state clone.
	for i := range state.Ships {
		if state.Ships[i].ID == fixture.attackerShipIDs[0] {
			state.Ships[i].Spec.WarpDriveID = "nuclear_drive"
			state.Ships[i].Spec.FTLSpeed = 2
			state.Ships[i].Spec.SpaceUsed = 0
			state.Ships[i].Spec.BaseDesignCostPP = 25
			state.Ships[i].Spec.ProductionCostPP = 25
			state.Ships[i].Spec.Weapons = nil
		}
	}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	resolver := &armingEconomyResolver{inner: loadTacticalEconomyResolver(t), attackerShipID: fixture.attackerShipIDs[0]}
	s, err := NewGameSession("tactical-returned-state", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "tactical-returned-state")
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatalf("encounter preparation did not use returned pre-encounter state: %v", err)
	}
	view, _ := s.ObserverView()
	if len(view.Battles) != 1 || view.Battles[0].Spec.Tactical == nil {
		t.Fatalf("returned-state tactical fixture missing: %+v", view.Battles)
	}
	found := false
	for _, ship := range view.State.Ships {
		if ship.ID == fixture.attackerShipIDs[0] {
			found = ship.Spec.WarpDriveID == "fusion_drive" && len(ship.Spec.Weapons) == 1
		}
	}
	if !found {
		t.Fatal("committed pre-encounter State did not contain resolver-only armed Ship snapshot")
	}
}

func TestParallelTacticalBattlesKeepLocalEventsAndGlobalStableCompletionOrder(t *testing.T) {
	state, seats, _ := makeTacticalStrategicFixture(t, 2)
	resolver := loadTacticalEconomyResolver(t)
	s, err := NewGameSession("tactical-parallel", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "tactical-parallel")
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	started, _ := s.ObserverView()
	if len(started.Battles) != 2 || started.Battles[0].Spec.ID != 1 || started.Battles[1].Spec.ID != 2 || started.Battles[0].Spec.Tactical == nil || started.Battles[1].Spec.Tactical == nil {
		t.Fatalf("parallel tactical battles=%+v", started.Battles)
	}
	if err := submitGoldenBattleCommands(t, s, started.Battles[1]); err != nil {
		t.Fatal(err)
	}
	mid, _ := s.ObserverView()
	if mid.Phase != PhaseEncounters || mid.Battles[1].Phase != battle.PhaseCompleted || mid.Battles[0].Phase != battle.PhaseActive {
		t.Fatalf("reverse first completion state=%+v", mid.Battles)
	}
	for _, event := range mid.Events {
		if event.Kind == "battle_completed" || event.Kind == "beam_fired" || event.Kind == "battle_damage_applied" {
			t.Fatalf("local tactical/wall-clock completion leaked globally before wave completion: %+v", event)
		}
	}
	if err := submitGoldenBattleCommands(t, s, started.Battles[0]); err != nil {
		t.Fatal(err)
	}
	finished, _ := s.ObserverView()
	if finished.Phase != PhasePostResolution {
		t.Fatalf("parallel tactical phase=%s", finished.Phase)
	}
	var completedIDs []uint64
	for _, event := range finished.Events {
		if event.Kind == "beam_fired" || event.Kind == "battle_damage_applied" || event.Kind == "ship_destroyed" {
			t.Fatalf("local tactical event leaked into global stream: %+v", event)
		}
		if event.Kind == "battle_completed" {
			var completed battle.View
			if err := json.Unmarshal(event.Data, &completed); err != nil {
				t.Fatal(err)
			}
			completedIDs = append(completedIDs, completed.Spec.ID)
			if completed.Tactical == nil || len(completed.Tactical.Events) != 10 {
				t.Fatalf("completed Battle %d lost local tactical events", completed.Spec.ID)
			}
		}
	}
	if !reflect.DeepEqual(completedIDs, []uint64{1, 2}) {
		t.Fatalf("reverse wall-clock tactical completion leaked globally: %v", completedIDs)
	}
}

func TestSlice175Checkpoint2ShieldedStrategicEncounterStartsTacticalBattle(t *testing.T) {
	state, seats, fixture := makeTacticalStrategicFixture(t, 1)
	for i := range state.Ships {
		if state.Ships[i].ID == fixture.defenderShipIDs[0] {
			state.Ships[i].Spec.ShieldID = "class_i_shield"
		}
	}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	resolver := loadTacticalEconomyResolver(t)
	s, err := NewGameSession("tactical-shielded", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "tactical-shielded")
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	started, _ := s.ObserverView()
	if len(started.Battles) != 1 || started.Battles[0].Spec.Tactical == nil || started.Battles[0].Spec.TacticalUnsupportedReason != "" || started.Battles[0].Tactical == nil {
		t.Fatalf("shielded legal encounter did not start Tactical: %+v", started.Battles)
	}
	var defender *battle.TacticalShipSpec
	for i := range started.Battles[0].Spec.Tactical.Ships {
		ship := &started.Battles[0].Spec.Tactical.Ships[i]
		if ship.ShipID == fixture.defenderShipIDs[0] {
			defender = ship
			break
		}
	}
	if defender == nil || defender.ShieldID != "class_i_shield" || defender.ShieldStrength != 1 || defender.ShieldArcMax != 5 {
		t.Fatalf("shielded defender projection=%+v", defender)
	}
}

func TestSlice177Checkpoint2TacticalEncounterProjectsExecutableBeamAndOrdnance(t *testing.T) {
	state, seats, fixture := makeTacticalStrategicFixture(t, 1)
	attackerID := fixture.attackerShipIDs[0]
	for i := range state.Ships {
		if state.Ships[i].ID != attackerID {
			continue
		}
		state.Ships[i].Name = "Mm"
		state.Ships[i].Spec = core.ShipDesignSpec{
			HullID:             "battleship",
			StrategicPictureID: 24,
			WarpDriveID:        "interphased_drive",
			FTLSpeed:           7,
			ComputerID:         "cybertronic_computer",
			ArmorID:            "xentronium_armor",
			ShieldID:           "class_v_shield",
			FuelCellID:         "thorium_fuel_cells",
			FuelRangeParsecs:   255,
			HullBaseCostPP:     600,
			HullSpace:          312,
			SpaceUsed:          259,
			BaseDesignCostPP:   1456,
			ProductionCostPP:   1456,
			Specials: []core.ShipSpecialMount{
				{Slot: 0, SpecialID: "battle_pods"},
			},
			Weapons: []core.ShipWeaponMount{
				{
					Slot:        0,
					WeaponID:    "mauler_device",
					Count:       2,
					MountType:   core.ShipWeaponMountHeavy,
					FiringArcID: core.ShipWeaponFiringArc360,
				},
				{
					Slot:        1,
					WeaponID:    "pulson_missile",
					Count:       3,
					MountType:   core.ShipWeaponMountStandard,
					ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierFast},
					Ammo:        10,
					FiringArcID: core.ShipWeaponFiringArc360,
				},
				{
					Slot:        2,
					WeaponID:    "fusion_beam",
					Count:       3,
					MountType:   core.ShipWeaponMountHeavy,
					ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierEnveloping},
					FiringArcID: core.ShipWeaponFiringArc360,
				},
			},
		}
		break
	}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}

	resolver := loadTacticalEconomyResolver(t)
	s, err := NewGameSession("tactical-live-loadout", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "tactical-live-loadout")
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatalf("strategic resolution rejected live All-Tech loadout: %v", err)
	}

	started, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if started.Phase != PhaseEncounters || len(started.Battles) != 1 {
		t.Fatalf("live All-Tech loadout did not reach encounter boundary: phase=%q battles=%d", started.Phase, len(started.Battles))
	}
	view := started.Battles[0]
	if view.Spec.Tactical == nil || view.Spec.TacticalUnsupportedReason != "" || view.Tactical == nil {
		t.Fatalf("live All-Tech loadout did not start Tactical: %+v", view)
	}
	if len(view.Tactical.WeaponActions) != 3 {
		t.Fatalf("live All-Tech weapon actions=%+v", view.Tactical.WeaponActions)
	}
	for _, action := range view.Tactical.WeaponActions {
		switch action.WeaponSlot {
		case 0, 2:
			if !action.Available || action.CommandKind != battle.CommandFireBeam || len(action.Targets) == 0 {
				t.Fatalf("Slice 17.6 Beam slot must be executable: %+v", action)
			}
		case 1:
			if !action.Available || action.CommandKind != battle.CommandFireOrdnance || len(action.Targets) == 0 {
				t.Fatalf("Slice 17.7 Ordnance slot must be executable: %+v", action)
			}
		default:
			t.Fatalf("unexpected weapon action: %+v", action)
		}
	}
}
