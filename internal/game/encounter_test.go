package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func encounterTestContext(state *core.GameState) ResolveContext {
	ctx := ResolveContext{Seats: make([]SeatAuthority, len(state.Empires))}
	for i := range state.Empires {
		ctx.Seats[i] = SeatAuthority{SeatID: protocol.SeatID(i + 1), EmpireID: state.Empires[i].ID}
	}
	return ctx
}

func addEncounterTestEmpire(state *core.GameState, name string) core.ID {
	id := state.NewID()
	state.Empires = append(state.Empires, core.Empire{ID: id, Name: name, RaceID: "human"})
	setCombatFleetTestTech(&state.Empires[len(state.Empires)-1])
	return id
}

func TestPrepareEncounterBoundaryAggregatesSameEmpireFleetsAndDirection(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 2201)
	state := fixture.state
	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	firstFleet, firstShips := addCombatFleetTestFleet(state, fixture.blockaderEmpireID, fixture.targetSystemID, 2)
	secondFleet, secondShips := addCombatFleetTestFleet(state, fixture.blockaderEmpireID, fixture.targetSystemID, 1)
	defenderFleet, defenderShips := addCombatFleetTestFleet(state, fixture.targetEmpireID, fixture.targetSystemID, 2)
	state.DiplomaticRelations = []core.DiplomaticRelation{{
		FromEmpireID: fixture.blockaderEmpireID, ToEmpireID: fixture.targetEmpireID, Stance: core.DiplomaticStanceHostile,
	}}
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	events, encounters, err := resolver.prepareEncounterBoundary(encounterTestContext(state), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 || len(encounters) != 1 {
		t.Fatalf("boundary events=%v encounters=%+v", events, encounters)
	}
	got := encounters[0]
	if got.SystemID != fixture.targetSystemID || got.Attacker.EmpireID != fixture.blockaderEmpireID || got.Defender.EmpireID != fixture.targetEmpireID {
		t.Fatalf("directed encounter=%+v", got)
	}
	wantFleetIDs := []core.ID{firstFleet, secondFleet}
	wantShipIDs := append(append([]core.ID(nil), firstShips...), secondShips...)
	if !reflect.DeepEqual(got.Attacker.CombatFleetIDs, wantFleetIDs) || !reflect.DeepEqual(got.Attacker.ShipIDs, wantShipIDs) {
		t.Fatalf("aggregated attacker=%+v want fleets=%v ships=%v", got.Attacker, wantFleetIDs, wantShipIDs)
	}
	if !reflect.DeepEqual(got.Defender.CombatFleetIDs, []core.ID{defenderFleet}) || !reflect.DeepEqual(got.Defender.ShipIDs, defenderShips) {
		t.Fatalf("defender=%+v", got.Defender)
	}
	if !reflect.DeepEqual(got.DefenderColonyIDs, []core.ID{fixture.targetColonyID}) {
		t.Fatalf("defender colony context=%v", got.DefenderColonyIDs)
	}
	if !reflect.DeepEqual(got.Participants, []protocol.SeatID{1, 2}) {
		t.Fatalf("participants=%v", got.Participants)
	}
}

func TestPrepareEncounterBoundaryCivilianOnlyOverrunAndColonySuppression(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 2202)
	state := fixture.state
	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	attackerSystem := state.Galaxy.Systems[0].ID
	addCombatFleetTestFleet(state, fixture.blockaderEmpireID, attackerSystem, 1)
	civilianID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID: civilianID, EmpireID: fixture.targetEmpireID, Role: core.StrategicFleetRoleCivilian,
		SpecialKind: core.StrategicFleetSpecialOutpostShip, AtSystemID: attackerSystem, FTLSpeed: 2,
	})
	state.DiplomaticRelations = []core.DiplomaticRelation{{
		FromEmpireID: fixture.blockaderEmpireID, ToEmpireID: fixture.targetEmpireID, Stance: core.DiplomaticStanceHostile,
	}}
	resolver, err := NewEconomyResolver(loadCommittedEconomyRules(t))
	if err != nil {
		t.Fatal(err)
	}
	events, encounters, err := resolver.prepareEncounterBoundary(encounterTestContext(state), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(encounters) != 0 || findDomainEvent(events, "empire.civilian_fleets_overrun") == nil {
		t.Fatalf("overrun events=%v encounters=%v", events, encounters)
	}
	if _, fleet := strategicFleetByID(state, civilianID); fleet != nil {
		t.Fatalf("civilian fleet %d survived unescorted hostile overrun", civilianID)
	}

	// A defender Colony is an unresolved original combat-defense boundary. Do not
	// silently treat its civilian ship as defenseless in Slice 06.
	state = fixture.state
	// The first boundary mutated fixture.state; rebuild a fresh fixture.
	fixture = newBlockadeTestFixture(t, 2203)
	state = fixture.state
	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	addCombatFleetTestFleet(state, fixture.blockaderEmpireID, fixture.targetSystemID, 1)
	civilianID = state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID: civilianID, EmpireID: fixture.targetEmpireID, Role: core.StrategicFleetRoleCivilian,
		SpecialKind: core.StrategicFleetSpecialColonyShip, AtSystemID: fixture.targetSystemID, FTLSpeed: 2,
	})
	state.DiplomaticRelations = []core.DiplomaticRelation{{
		FromEmpireID: fixture.blockaderEmpireID, ToEmpireID: fixture.targetEmpireID, Stance: core.DiplomaticStanceHostile,
	}}
	events, encounters, err = resolver.prepareEncounterBoundary(encounterTestContext(state), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 || len(encounters) != 0 {
		t.Fatalf("colony-only defense should defer events=%v encounters=%v", events, encounters)
	}
	if _, fleet := strategicFleetByID(state, civilianID); fleet == nil {
		t.Fatal("defender Colony context incorrectly allowed civilian auto-overrun")
	}
}

func TestResumeAfterEncountersAppliesCasualtiesRetreatThenPostEncounterWork(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 2204)
	state := fixture.state
	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	_, attackerShips := addCombatFleetTestFleet(state, fixture.blockaderEmpireID, fixture.targetSystemID, 1)
	defenderFleet, defenderShips := addCombatFleetTestFleet(state, fixture.targetEmpireID, fixture.targetSystemID, 2)
	civilianFleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID: civilianFleetID, EmpireID: fixture.targetEmpireID, Role: core.StrategicFleetRoleCivilian,
		SpecialKind: core.StrategicFleetSpecialOutpostShip, AtSystemID: fixture.targetSystemID, FTLSpeed: 2,
	})
	state.DiplomaticRelations = []core.DiplomaticRelation{{
		FromEmpireID: fixture.blockaderEmpireID, ToEmpireID: fixture.targetEmpireID, Stance: core.DiplomaticStanceHostile,
	}}
	ctx := encounterTestContext(state)
	resolver, err := NewEconomyResolver(loadCommittedEconomyRules(t))
	if err != nil {
		t.Fatal(err)
	}
	_, encounters, err := resolver.prepareEncounterBoundary(ctx, state)
	if err != nil || len(encounters) != 1 {
		t.Fatalf("prepare encounter: %v encounters=%v", err, encounters)
	}
	outcome := EncounterOutcome{
		BattleID: 1, Encounter: encounters[0], WinnerSeat: encounters[0].Attacker.SeatID, Outcome: "attacker_victory",
		DestroyedShipIDs: []core.ID{defenderShips[0]},
	}
	resolution, err := resolver.ResumeAfterEncounters(ctx, state, []EncounterOutcome{outcome})
	if err != nil {
		t.Fatal(err)
	}
	if len(resolution.Encounters) != 0 {
		t.Fatalf("unexpected follow-up encounters=%+v", resolution.Encounters)
	}
	if shipByID(state, defenderShips[0]) != nil {
		t.Fatalf("destroyed ship %d still exists", defenderShips[0])
	}
	_, fleet := strategicFleetByID(state, defenderFleet)
	if fleet == nil || fleet.AtSystemID != 0 || fleet.DestinationSystemID != state.Galaxy.Systems[2].ID || fleet.RemainingTurns <= 0 {
		t.Fatalf("loser fleet did not retreat to other owned Colony system: %+v", fleet)
	}
	if !reflect.DeepEqual(fleet.ShipIDs, []core.ID{defenderShips[1]}) {
		t.Fatalf("retreat fleet ships=%v", fleet.ShipIDs)
	}
	_, civilianFleet := strategicFleetByID(state, civilianFleetID)
	if civilianFleet == nil || civilianFleet.AtSystemID != 0 || civilianFleet.DestinationSystemID != state.Galaxy.Systems[2].ID || civilianFleet.RemainingTurns <= 0 {
		t.Fatalf("escorted civilian did not retreat with losing side: %+v", civilianFleet)
	}
	if shipByID(state, attackerShips[0]) == nil {
		t.Fatal("winning ship disappeared")
	}
	if got := state.Galaxy.Systems[1].BlockadedEmpireIDs; !reflect.DeepEqual(got, []core.ID{fixture.targetEmpireID}) {
		t.Fatalf("post-battle blockade=%v want target empire", got)
	}
	if findDomainEvent(resolution.Events, "empire.battle_casualties_applied") == nil || findDomainEvent(resolution.Events, "empire.fleet_retreated_after_battle") == nil {
		t.Fatalf("post-battle events=%v", resolution.Events)
	}
}

func TestResumeAfterEncountersUsesClosestColonyTieBreakAndDestroysWithoutRetreat(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 2205)
	state := fixture.state
	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	battleSystem := &state.Galaxy.Systems[1]
	// Add a second target Colony System at the same squared distance as system 2,
	// then verify lowest SystemID wins the deterministic MOOX tie.
	thirdSystemID := state.NewID()
	thirdPlanetID := state.NewID()
	state.Galaxy.Systems = append(state.Galaxy.Systems, core.StarSystem{
		ID: thirdSystemID, Name: "Delta", X: battleSystem.X - 10, Y: battleSystem.Y,
		Planets: []core.Planet{{ID: thirdPlanetID, Name: "Delta I", Orbit: 1, SizeID: "medium", MineralID: "abundant", GravityID: "normal_g", ClimateID: "terran"}},
	})
	thirdSystem := &state.Galaxy.Systems[len(state.Galaxy.Systems)-1]
	secondRetreatColonyID := state.NewID()
	thirdSystem.Planets[0].ColonyID = secondRetreatColonyID
	state.Colonies = append(state.Colonies, core.Colony{
		ID: secondRetreatColonyID, EmpireID: fixture.targetEmpireID, PlanetID: thirdSystem.Planets[0].ID,
		Population: core.NewAssimilatedPopulation(fixture.targetEmpireID, 1, 0, 0),
	})
	// Force equal geometry without changing IDs.
	state.Galaxy.Systems[2].X, state.Galaxy.Systems[2].Y = battleSystem.X+10, battleSystem.Y
	_, attackerShips := addCombatFleetTestFleet(state, fixture.blockaderEmpireID, battleSystem.ID, 1)
	defenderFleet, _ := addCombatFleetTestFleet(state, fixture.targetEmpireID, battleSystem.ID, 1)
	state.DiplomaticRelations = []core.DiplomaticRelation{{FromEmpireID: fixture.blockaderEmpireID, ToEmpireID: fixture.targetEmpireID, Stance: core.DiplomaticStanceHostile}}
	ctx := encounterTestContext(state)
	resolver, err := NewEconomyResolver(loadCommittedEconomyRules(t))
	if err != nil {
		t.Fatal(err)
	}
	_, encounters, err := resolver.prepareEncounterBoundary(ctx, state)
	if err != nil || len(encounters) != 1 {
		t.Fatalf("prepare=%v encounters=%v", err, encounters)
	}
	if _, err := resolver.ResumeAfterEncounters(ctx, state, []EncounterOutcome{{BattleID: 1, Encounter: encounters[0], WinnerSeat: encounters[0].Attacker.SeatID, Outcome: "win"}}); err != nil {
		t.Fatal(err)
	}
	_, retreated := strategicFleetByID(state, defenderFleet)
	wantDestination := state.Galaxy.Systems[2].ID
	if state.Galaxy.Systems[3].ID < wantDestination {
		wantDestination = state.Galaxy.Systems[3].ID
	}
	if retreated == nil || retreated.DestinationSystemID != wantDestination {
		t.Fatalf("tie retreat=%+v want destination=%d", retreated, wantDestination)
	}
	if shipByID(state, attackerShips[0]) == nil {
		t.Fatal("attacker missing after tie-break test")
	}

	// No other Colony means the losing combat fleet and its concrete ships are destroyed.
	fixture = newBlockadeTestFixture(t, 2206)
	state = fixture.state
	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	// Remove the target Empire's other Colony, leaving only the battle-system Colony.
	keep := state.Colonies[:0]
	for _, colony := range state.Colonies {
		if colony.ID != fixture.sourceColonyID {
			keep = append(keep, colony)
		}
	}
	state.Colonies = keep
	state.Galaxy.Systems[2].Planets[0].ColonyID = 0
	addCombatFleetTestFleet(state, fixture.blockaderEmpireID, fixture.targetSystemID, 1)
	defenderFleet, defenderShips := addCombatFleetTestFleet(state, fixture.targetEmpireID, fixture.targetSystemID, 1)
	state.DiplomaticRelations = []core.DiplomaticRelation{{FromEmpireID: fixture.blockaderEmpireID, ToEmpireID: fixture.targetEmpireID, Stance: core.DiplomaticStanceHostile}}
	ctx = encounterTestContext(state)
	_, encounters, err = resolver.prepareEncounterBoundary(ctx, state)
	if err != nil || len(encounters) != 1 {
		t.Fatalf("prepare no-retreat=%v encounters=%v", err, encounters)
	}
	resolution, err := resolver.ResumeAfterEncounters(ctx, state, []EncounterOutcome{{BattleID: 2, Encounter: encounters[0], WinnerSeat: encounters[0].Attacker.SeatID, Outcome: "win"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, fleet := strategicFleetByID(state, defenderFleet); fleet != nil || shipByID(state, defenderShips[0]) != nil {
		t.Fatalf("no-destination loser survived fleet=%+v ship=%+v", fleet, shipByID(state, defenderShips[0]))
	}
	if findDomainEvent(resolution.Events, "empire.fleet_destroyed_after_battle") == nil {
		t.Fatalf("missing retreat-destruction event: %v", resolution.Events)
	}
}

func TestEncounterWavesArePairwiseWithinOneSystem(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 2207)
	state := fixture.state
	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	thirdEmpireID := addEncounterTestEmpire(state, "Third")
	addCombatFleetTestFleet(state, fixture.blockaderEmpireID, fixture.targetSystemID, 1)
	addCombatFleetTestFleet(state, fixture.targetEmpireID, fixture.targetSystemID, 1)
	addCombatFleetTestFleet(state, thirdEmpireID, fixture.targetSystemID, 1)
	state.DiplomaticRelations = []core.DiplomaticRelation{
		{FromEmpireID: fixture.blockaderEmpireID, ToEmpireID: fixture.targetEmpireID, Stance: core.DiplomaticStanceHostile},
		{FromEmpireID: fixture.blockaderEmpireID, ToEmpireID: thirdEmpireID, Stance: core.DiplomaticStanceHostile},
	}
	ctx := encounterTestContext(state)
	resolver, err := NewEconomyResolver(loadCommittedEconomyRules(t))
	if err != nil {
		t.Fatal(err)
	}
	_, encounters, err := resolver.prepareEncounterBoundary(ctx, state)
	if err != nil || len(encounters) != 1 {
		t.Fatalf("first wave=%v encounters=%+v", err, encounters)
	}
	if encounters[0].Defender.EmpireID != fixture.targetEmpireID {
		t.Fatalf("canonical first defender=%d want=%d", encounters[0].Defender.EmpireID, fixture.targetEmpireID)
	}
	resolution, err := resolver.ResumeAfterEncounters(ctx, state, []EncounterOutcome{{
		BattleID: 1, Encounter: encounters[0], WinnerSeat: encounters[0].Attacker.SeatID, Outcome: "first_win",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(resolution.Encounters) != 1 || resolution.Encounters[0].SystemID != fixture.targetSystemID || resolution.Encounters[0].Defender.EmpireID != thirdEmpireID {
		t.Fatalf("second wave encounters=%+v", resolution.Encounters)
	}
}

func TestEncounterCasualtiesRemoveEmptyFleetAndRetreatMovementFailureDestroysLoser(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 2208)
	state := fixture.state
	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	addCombatFleetTestFleet(state, fixture.blockaderEmpireID, fixture.targetSystemID, 1)
	defenderFleet, defenderShips := addCombatFleetTestFleet(state, fixture.targetEmpireID, fixture.targetSystemID, 1)
	state.DiplomaticRelations = []core.DiplomaticRelation{{FromEmpireID: fixture.blockaderEmpireID, ToEmpireID: fixture.targetEmpireID, Stance: core.DiplomaticStanceHostile}}
	ctx := encounterTestContext(state)
	resolver, err := NewEconomyResolver(loadCommittedEconomyRules(t))
	if err != nil {
		t.Fatal(err)
	}
	_, encounters, err := resolver.prepareEncounterBoundary(ctx, state)
	if err != nil || len(encounters) != 1 {
		t.Fatalf("prepare casualty cleanup=%v encounters=%v", err, encounters)
	}
	if _, err := resolver.ResumeAfterEncounters(ctx, state, []EncounterOutcome{{
		BattleID: 1, Encounter: encounters[0], WinnerSeat: encounters[0].Attacker.SeatID, Outcome: "destroyed",
		DestroyedShipIDs: defenderShips,
	}}); err != nil {
		t.Fatal(err)
	}
	if shipByID(state, defenderShips[0]) != nil {
		t.Fatalf("destroyed concrete ship %d survived", defenderShips[0])
	}
	if _, fleet := strategicFleetByID(state, defenderFleet); fleet != nil {
		t.Fatalf("empty combat fleet %d survived casualty cleanup: %+v", defenderFleet, fleet)
	}

	fixture = newBlockadeTestFixture(t, 2209)
	state = fixture.state
	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	addCombatFleetTestFleet(state, fixture.blockaderEmpireID, fixture.targetSystemID, 1)
	defenderFleet, defenderShips = addCombatFleetTestFleet(state, fixture.targetEmpireID, fixture.targetSystemID, 1)
	state.DiplomaticRelations = []core.DiplomaticRelation{{FromEmpireID: fixture.blockaderEmpireID, ToEmpireID: fixture.targetEmpireID, Stance: core.DiplomaticStanceHostile}}
	ctx = encounterTestContext(state)
	_, encounters, err = resolver.prepareEncounterBoundary(ctx, state)
	if err != nil || len(encounters) != 1 {
		t.Fatalf("prepare movement-failure=%v encounters=%v", err, encounters)
	}
	// An owned retreat Colony exists, but removing current strategic drive/fuel
	// knowledge makes the supported movement profile unavailable. Contract says
	// destroy the asset rather than searching a second retreat destination.
	state.Empires[1].KnownTechnologyIDs = nil
	resolution, err := resolver.ResumeAfterEncounters(ctx, state, []EncounterOutcome{{
		BattleID: 2, Encounter: encounters[0], WinnerSeat: encounters[0].Attacker.SeatID, Outcome: "retreat_unavailable",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, fleet := strategicFleetByID(state, defenderFleet); fleet != nil || shipByID(state, defenderShips[0]) != nil {
		t.Fatalf("movement-unavailable loser survived fleet=%+v ship=%+v", fleet, shipByID(state, defenderShips[0]))
	}
	destroyed := findDomainEvent(resolution.Events, "empire.fleet_destroyed_after_battle")
	if destroyed == nil {
		t.Fatalf("missing movement-unavailable destruction event: %+v", resolution.Events)
	}
	var payload FleetDestroyedAfterBattleEvent
	if err := json.Unmarshal(destroyed.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Reason != "retreat_movement_unavailable" {
		t.Fatalf("movement failure reason=%q", payload.Reason)
	}
}

func TestPrepareEncounterBoundaryRejectsUnseatedHostileParticipant(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 2214)
	state := fixture.state
	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	addCombatFleetTestFleet(state, fixture.blockaderEmpireID, fixture.targetSystemID, 1)
	addCombatFleetTestFleet(state, fixture.targetEmpireID, fixture.targetSystemID, 1)
	state.DiplomaticRelations = []core.DiplomaticRelation{{FromEmpireID: fixture.blockaderEmpireID, ToEmpireID: fixture.targetEmpireID, Stance: core.DiplomaticStanceHostile}}
	resolver, err := NewEconomyResolver(loadCommittedEconomyRules(t))
	if err != nil {
		t.Fatal(err)
	}
	ctx := ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: fixture.blockaderEmpireID}}}
	if _, _, err := resolver.prepareEncounterBoundary(ctx, state); err == nil {
		t.Fatal("hostile encounter requiring an unseated defender unexpectedly succeeded")
	}
}
