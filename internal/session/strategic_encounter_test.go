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

type stagedEncounterResolverStub struct {
	initial      []game.Encounter
	resumeErr    error
	resumeCalls  int
	seenOutcomes [][]game.EncounterOutcome
	nextWaves    [][]game.Encounter
}

func (r *stagedEncounterResolverStub) Resolve(_ game.ResolveContext, state *core.GameState, _ []protocol.CommandBatch) (game.Resolution, error) {
	return game.Resolution{State: state, Encounters: cloneGameEncounters(r.initial)}, nil
}

func (r *stagedEncounterResolverStub) ResumeAfterEncounters(_ game.ResolveContext, state *core.GameState, outcomes []game.EncounterOutcome) (game.Resolution, error) {
	copyOutcomes := make([]game.EncounterOutcome, len(outcomes))
	for i := range outcomes {
		copyOutcomes[i] = outcomes[i]
		copyOutcomes[i].Encounter = game.CloneEncounter(outcomes[i].Encounter)
		copyOutcomes[i].DestroyedShipIDs = append([]core.ID(nil), outcomes[i].DestroyedShipIDs...)
	}
	r.seenOutcomes = append(r.seenOutcomes, copyOutcomes)
	if r.resumeErr != nil {
		return game.Resolution{}, r.resumeErr
	}
	var next []game.Encounter
	if r.resumeCalls < len(r.nextWaves) {
		next = cloneGameEncounters(r.nextWaves[r.resumeCalls])
	}
	r.resumeCalls++
	event, err := game.NewDomainEvent("test.encounter_wave_resumed", 0, 0, map[string]int{"wave": r.resumeCalls})
	if err != nil {
		return game.Resolution{}, err
	}
	return game.Resolution{State: state, Events: []game.DomainEvent{event}, Encounters: next}, nil
}

func cloneGameEncounters(in []game.Encounter) []game.Encounter {
	out := make([]game.Encounter, len(in))
	for i := range in {
		out[i] = game.CloneEncounter(in[i])
	}
	return out
}

func fakeStrategicEncounter(systemID, attackerEmpireID, defenderEmpireID core.ID, attackerSeat, defenderSeat protocol.SeatID, base core.ID) game.Encounter {
	participants := []protocol.SeatID{attackerSeat, defenderSeat}
	if participants[0] > participants[1] {
		participants[0], participants[1] = participants[1], participants[0]
	}
	return game.Encounter{
		SystemID:     systemID,
		Attacker:     game.EncounterSide{EmpireID: attackerEmpireID, SeatID: attackerSeat, CombatFleetIDs: []core.ID{base}, ShipIDs: []core.ID{base + 1}},
		Defender:     game.EncounterSide{EmpireID: defenderEmpireID, SeatID: defenderSeat, CombatFleetIDs: []core.ID{base + 2}, ShipIDs: []core.ID{base + 3}},
		Participants: participants,
	}
}

func submitStubTurn(t *testing.T, s *GameSession, gameID string) {
	t.Helper()
	view, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	for _, seat := range view.Seats {
		batch := protocol.CommandBatch{
			SchemaVersion: protocol.CommandSchemaVersion,
			GameID:        gameID,
			SeatID:        seat.Seat.ID,
			Turn:          view.Turn,
			BaseRevision:  view.Revision,
		}
		if err := s.SubmitTurn(batch); err != nil {
			t.Fatal(err)
		}
	}
}

func TestStagedEncounterWaveCommitsParallelResultsInBattleIDOrder(t *testing.T) {
	state, seats := twoSeatFixture(t)
	encounter1 := fakeStrategicEncounter(state.Galaxy.Systems[0].ID, seats[0].EmpireID, seats[1].EmpireID, seats[0].ID, seats[1].ID, 100)
	encounter2 := fakeStrategicEncounter(state.Galaxy.Systems[1].ID, seats[1].EmpireID, seats[0].EmpireID, seats[1].ID, seats[0].ID, 200)
	resolver := &stagedEncounterResolverStub{initial: []game.Encounter{encounter1, encounter2}}
	s, err := NewGameSession("staged-parallel", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "staged-parallel")
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	started, _ := s.ObserverView()
	if started.Phase != PhaseEncounters || len(started.Battles) != 2 || started.Battles[0].Spec.ID != 1 || started.Battles[1].Spec.ID != 2 {
		t.Fatalf("started encounters=%+v", started.Battles)
	}
	if started.Battles[0].Spec.Seed == started.Battles[1].Spec.Seed {
		t.Fatal("parallel battle seeds collided")
	}
	if started.Battles[0].Spec.Attacker.SeatID != seats[0].ID || started.Battles[0].Spec.Defender.SeatID != seats[1].ID {
		t.Fatalf("battle direction lost: %+v", started.Battles[0].Spec)
	}

	if err := s.CompleteBattle(2, battle.Result{WinnerSeat: seats[1].ID, Outcome: "second_done"}); err != nil {
		t.Fatal(err)
	}
	mid, _ := s.ObserverView()
	if mid.Phase != PhaseEncounters || resolver.resumeCalls != 0 {
		t.Fatalf("non-final result advanced staged continuation: phase=%s calls=%d", mid.Phase, resolver.resumeCalls)
	}
	for _, event := range mid.Events {
		if event.Kind == "battle_completed" {
			t.Fatal("wall-clock completion leaked before whole wave was ready")
		}
	}

	if err := s.CompleteBattle(1, battle.Result{WinnerSeat: seats[0].ID, Outcome: "first_done"}); err != nil {
		t.Fatal(err)
	}
	finished, _ := s.ObserverView()
	if finished.Phase != PhasePostResolution || resolver.resumeCalls != 1 {
		t.Fatalf("finished phase=%s calls=%d", finished.Phase, resolver.resumeCalls)
	}
	if len(resolver.seenOutcomes) != 1 || len(resolver.seenOutcomes[0]) != 2 || resolver.seenOutcomes[0][0].BattleID != 1 || resolver.seenOutcomes[0][1].BattleID != 2 {
		t.Fatalf("resume outcomes order=%+v", resolver.seenOutcomes)
	}
	var completedIDs []uint64
	var kinds []string
	for _, event := range finished.Events {
		kinds = append(kinds, event.Kind)
		if event.Kind == "battle_completed" {
			var view battle.View
			if err := jsonUnmarshal(event.Data, &view); err != nil {
				t.Fatal(err)
			}
			completedIDs = append(completedIDs, view.Spec.ID)
		}
	}
	if !reflect.DeepEqual(completedIDs, []uint64{1, 2}) {
		t.Fatalf("completion order=%v", completedIDs)
	}
	firstCompleted, resumed := indexOfKind(kinds, "battle_completed"), indexOfKind(kinds, "test.encounter_wave_resumed")
	if firstCompleted < 0 || resumed < firstCompleted {
		t.Fatalf("event ordering=%v", kinds)
	}
}

func TestStagedEncounterFinalResultFailureIsAtomicAndRetryable(t *testing.T) {
	state, seats := twoSeatFixture(t)
	resolver := &stagedEncounterResolverStub{
		initial:   []game.Encounter{fakeStrategicEncounter(state.Galaxy.Systems[0].ID, seats[0].EmpireID, seats[1].EmpireID, seats[0].ID, seats[1].ID, 300)},
		resumeErr: errors.New("continuation rejected"),
	}
	s, err := NewGameSession("staged-atomic", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "staged-atomic")
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	before, _ := s.ObserverView()
	if err := s.CompleteBattle(1, battle.Result{WinnerSeat: seats[0].ID, Outcome: "invalid", DestroyedShipIDs: []core.ID{999999}}); err == nil {
		t.Fatal("expected out-of-snapshot casualty to reject final result")
	}
	invalidAfter, _ := s.ObserverView()
	if invalidAfter.Revision != before.Revision || invalidAfter.Battles[0].Phase != battle.PhaseActive || resolver.resumeCalls != 0 || !reflect.DeepEqual(invalidAfter.State, before.State) {
		t.Fatalf("invalid casualty result mutated or resumed session: before=%d after=%+v calls=%d", before.Revision, invalidAfter, resolver.resumeCalls)
	}
	if err := s.CompleteBattle(1, battle.Result{WinnerSeat: seats[0].ID, Outcome: "candidate"}); err == nil {
		t.Fatal("expected staged continuation failure")
	}
	after, _ := s.ObserverView()
	if after.Revision != before.Revision || after.Phase != PhaseEncounters || !reflect.DeepEqual(after.State, before.State) || after.Battles[0].Phase != battle.PhaseActive {
		t.Fatalf("failed final result mutated authoritative session: before rev=%d after=%+v", before.Revision, after)
	}
	resolver.resumeErr = nil
	if err := s.CompleteBattle(1, battle.Result{WinnerSeat: seats[0].ID, Outcome: "candidate"}); err != nil {
		t.Fatal(err)
	}
	finished, _ := s.ObserverView()
	if finished.Phase != PhasePostResolution || finished.Battles[0].Phase != battle.PhaseCompleted {
		t.Fatalf("retry did not complete: %+v", finished.Battles)
	}
}

func TestStagedEncounterSequentialSameSystemWavesKeepPhaseAndIDs(t *testing.T) {
	state, seats := twoSeatFixture(t)
	first := fakeStrategicEncounter(state.Galaxy.Systems[0].ID, seats[0].EmpireID, seats[1].EmpireID, seats[0].ID, seats[1].ID, 400)
	second := fakeStrategicEncounter(state.Galaxy.Systems[0].ID, seats[1].EmpireID, seats[0].EmpireID, seats[1].ID, seats[0].ID, 500)
	resolver := &stagedEncounterResolverStub{initial: []game.Encounter{first}, nextWaves: [][]game.Encounter{{second}}}
	s, err := NewGameSession("staged-waves", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "staged-waves")
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	if err := s.CompleteBattle(1, battle.Result{WinnerSeat: seats[0].ID, Outcome: "wave1"}); err != nil {
		t.Fatal(err)
	}
	wave2, _ := s.ObserverView()
	if wave2.Phase != PhaseEncounters || len(wave2.Battles) != 1 || wave2.Battles[0].Spec.ID != 2 {
		t.Fatalf("second wave=%+v phase=%s", wave2.Battles, wave2.Phase)
	}
	for _, event := range wave2.Events {
		if event.Kind != "phase_changed" {
			continue
		}
		var phase struct {
			Phase Phase `json:"phase"`
		}
		if err := jsonUnmarshal(event.Data, &phase); err != nil {
			t.Fatal(err)
		}
		if phase.Phase == PhasePostResolution {
			t.Fatal("same-system next wave briefly exposed PostResolution")
		}
	}
	if err := s.CompleteBattle(2, battle.Result{WinnerSeat: seats[1].ID, Outcome: "wave2"}); err != nil {
		t.Fatal(err)
	}
	finished, _ := s.ObserverView()
	if finished.Phase != PhasePostResolution || resolver.resumeCalls != 2 {
		t.Fatalf("final wave phase=%s calls=%d", finished.Phase, resolver.resumeCalls)
	}
}

func indexOfKind(kinds []string, target string) int {
	for i, kind := range kinds {
		if kind == target {
			return i
		}
	}
	return -1
}

// Keep JSON decoding local so test call sites remain terse without changing the
// production event API.
func jsonUnmarshal(data []byte, target any) error {
	return json.Unmarshal(data, target)
}

func strategicEncounterRules(t *testing.T) *game.EconomyResolver {
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

func sessionEncounterShipSpec() core.ShipDesignSpec {
	return core.ShipDesignSpec{
		HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "nuclear_drive", FTLSpeed: 2,
		ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4,
		HullBaseCostPP: 20, HullSpace: 25, BaseDesignCostPP: 25, ProductionCostPP: 25,
	}
}

func addSessionEncounterCombatFleet(state *core.GameState, empireID, systemID core.ID) (core.ID, core.ID) {
	spec := sessionEncounterShipSpec()
	designID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{
		ID: designID, EmpireID: empireID, Revision: 1, Name: "Encounter test", Spec: spec,
	})
	shipID := state.NewID()
	state.Ships = append(state.Ships, core.Ship{
		ID: shipID, EmpireID: empireID, SourceDesignID: designID, SourceDesignRevision: 1, Name: "Encounter test", Spec: spec,
	})
	fleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID: fleetID, EmpireID: empireID, Role: core.StrategicFleetRoleCombat, SpecialKind: core.StrategicFleetSpecialNone,
		AtSystemID: systemID, ShipIDs: []core.ID{shipID},
	})
	return fleetID, shipID
}

type realEncounterFixture struct {
	state            *core.GameState
	seats            []Seat
	attackerEmpireID core.ID
	defenderEmpireID core.ID
	battleSystemID   core.ID
	retreatSystemID  core.ID
	defenderFleetID  core.ID
	defenderShipID   core.ID
}

func newRealEncounterFixture(seed uint64) realEncounterFixture {
	state := core.NewSmallFixture(seed)
	attacker := &state.Empires[0]
	attacker.KnownTechnologyIDs = []int{51, 72, 120, 167}
	defenderEmpireID := state.NewID()
	defender := core.Empire{
		ID: defenderEmpireID, Name: "Defender", RaceID: "human", Freighters: 10,
		KnownTechnologyIDs: []int{51, 72, 120, 167},
	}

	battleSystem := &state.Galaxy.Systems[1]
	battlePlanet := &battleSystem.Planets[0]
	defenderColonyID := state.NewID()
	battlePlanet.ColonyID = defenderColonyID
	defenderColony := core.Colony{
		ID: defenderColonyID, EmpireID: defenderEmpireID, PlanetID: battlePlanet.ID,
		Population: core.NewAssimilatedPopulation(defenderEmpireID, 0, 1, 1),
	}

	retreatSystem := &state.Galaxy.Systems[2]
	retreatPlanet := &retreatSystem.Planets[0]
	retreatColonyID := state.NewID()
	retreatPlanet.ColonyID = retreatColonyID
	retreatColony := core.Colony{
		ID: retreatColonyID, EmpireID: defenderEmpireID, PlanetID: retreatPlanet.ID,
		Population: core.NewAssimilatedPopulation(defenderEmpireID, 2, 0, 0),
	}
	defender.Capital = retreatColonyID
	state.Empires = append(state.Empires, defender)
	state.Colonies = append(state.Colonies, defenderColony, retreatColony)

	addSessionEncounterCombatFleet(state, attacker.ID, battleSystem.ID)
	defenderFleetID, defenderShipID := addSessionEncounterCombatFleet(state, defenderEmpireID, battleSystem.ID)
	state.DiplomaticRelations = []core.DiplomaticRelation{{
		FromEmpireID: attacker.ID, ToEmpireID: defenderEmpireID, Stance: core.DiplomaticStanceHostile,
	}}
	return realEncounterFixture{
		state: state,
		seats: []Seat{
			{ID: 1, EmpireID: attacker.ID, Name: "Attacker", Controller: ControllerLocalHuman},
			{ID: 2, EmpireID: defenderEmpireID, Name: "Defender", Controller: ControllerBuiltinAI},
		},
		attackerEmpireID: attacker.ID, defenderEmpireID: defenderEmpireID,
		battleSystemID: battleSystem.ID, retreatSystemID: retreatSystem.ID,
		defenderFleetID: defenderFleetID, defenderShipID: defenderShipID,
	}
}

func runRealEncounterSession(t *testing.T, seed uint64, gameID string) ObserverView {
	t.Helper()
	fixture := newRealEncounterFixture(seed)
	s, err := NewGameSession(gameID, fixture.state, fixture.seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, gameID)
	if err := s.ResolveStrategic(strategicEncounterRules(t)); err != nil {
		t.Fatal(err)
	}
	pre, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if pre.Phase != PhaseEncounters || len(pre.Battles) != 1 {
		t.Fatalf("real resolver did not pause at encounter: phase=%s battles=%+v", pre.Phase, pre.Battles)
	}
	if len(pre.State.Galaxy.Systems[1].BlockadedEmpireIDs) != 0 {
		t.Fatalf("blockade resolved before battle: %v", pre.State.Galaxy.Systems[1].BlockadedEmpireIDs)
	}
	if pre.Battles[0].Spec.SystemID != fixture.battleSystemID || pre.Battles[0].Spec.Attacker.EmpireID != fixture.attackerEmpireID || pre.Battles[0].Spec.Defender.EmpireID != fixture.defenderEmpireID {
		t.Fatalf("real encounter spec=%+v", pre.Battles[0].Spec)
	}
	if len(pre.Battles[0].Spec.DefenderColonyIDs) != 1 {
		t.Fatalf("defender Colony context=%v", pre.Battles[0].Spec.DefenderColonyIDs)
	}
	// Observer BattleSpecs are detached copies.
	pre.Battles[0].Spec.Attacker.ShipIDs[0] = 999999
	fresh, _ := s.ObserverView()
	if fresh.Battles[0].Spec.Attacker.ShipIDs[0] == 999999 {
		t.Fatal("Observer battle spec mutation leaked into authoritative child")
	}
	if err := s.CompleteBattle(fresh.Battles[0].Spec.ID, battle.Result{WinnerSeat: 1, Outcome: "attacker_victory"}); err != nil {
		t.Fatal(err)
	}
	final, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if final.Phase != PhasePostResolution {
		t.Fatalf("real encounter did not resume post-resolution: %s", final.Phase)
	}
	var defenderFleet *core.StrategicFleet
	for i := range final.State.StrategicFleets {
		if final.State.StrategicFleets[i].ID == fixture.defenderFleetID {
			defenderFleet = &final.State.StrategicFleets[i]
			break
		}
	}
	if defenderFleet == nil || defenderFleet.AtSystemID != 0 || defenderFleet.DestinationSystemID != fixture.retreatSystemID || defenderFleet.RemainingTurns <= 0 {
		t.Fatalf("defender did not retreat before post resolution: %+v", defenderFleet)
	}
	if len(final.State.Galaxy.Systems[1].BlockadedEmpireIDs) != 1 || final.State.Galaxy.Systems[1].BlockadedEmpireIDs[0] != fixture.defenderEmpireID {
		t.Fatalf("post-battle blockade=%v", final.State.Galaxy.Systems[1].BlockadedEmpireIDs)
	}
	battleCompletedIndex := -1
	retreatIndex := -1
	for i, event := range final.Events {
		switch event.Kind {
		case "battle_completed":
			if battleCompletedIndex < 0 {
				battleCompletedIndex = i
			}
		case "empire.fleet_retreated_after_battle":
			if retreatIndex < 0 {
				retreatIndex = i
			}
		}
	}
	if battleCompletedIndex < 0 || retreatIndex <= battleCompletedIndex {
		t.Fatalf("battle/retreat event order battle=%d retreat=%d", battleCompletedIndex, retreatIndex)
	}
	return final
}

func TestEconomyResolverSessionEncounterBoundaryRetreatBlockadeObserverAndReplay(t *testing.T) {
	first := runRealEncounterSession(t, 2230, "real-encounter")
	second := runRealEncounterSession(t, 2230, "real-encounter")
	if !reflect.DeepEqual(first.State, second.State) {
		t.Fatalf("identical encounter sessions diverged State:\nfirst=%+v\nsecond=%+v", first.State.StrategicFleets, second.State.StrategicFleets)
	}
	if !reflect.DeepEqual(first.Events, second.Events) {
		t.Fatalf("identical encounter sessions diverged Events")
	}
	if !reflect.DeepEqual(first.Battles, second.Battles) {
		t.Fatalf("identical encounter sessions diverged Battle views")
	}
}

func TestEconomyResolverNoEncounterTurnRemainsOneShotCompatible(t *testing.T) {
	state := core.NewSmallFixture(2231)
	state.Empires[0].KnownTechnologyIDs = []int{51, 72, 120, 167}
	seats := []Seat{{ID: 1, EmpireID: state.Empires[0].ID, Name: "Solo", Controller: ControllerLocalHuman}}
	s, err := NewGameSession("no-encounter", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "no-encounter")
	if err := s.ResolveStrategic(strategicEncounterRules(t)); err != nil {
		t.Fatal(err)
	}
	view, _ := s.ObserverView()
	if view.Phase != PhasePostResolution || len(view.Battles) != 0 {
		t.Fatalf("no-encounter turn changed behavior: phase=%s battles=%v", view.Phase, view.Battles)
	}
}
