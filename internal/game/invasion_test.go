package game

import (
	"bytes"
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func addInvasionTransport(state *core.GameState, empireID, systemID core.ID) core.ID {
	id := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID: id, EmpireID: empireID, Role: core.StrategicFleetRoleCivilian,
		SpecialKind: core.StrategicFleetSpecialTroopTransport, AtSystemID: systemID, FTLSpeed: 2,
	})
	return id
}

func invasionTestContext(attackerEmpireID, defenderEmpireID core.ID) ResolveContext {
	return ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: attackerEmpireID}, {SeatID: 2, EmpireID: defenderEmpireID}}}
}

func TestInvasionOpportunityRequiresOrbitalControlAndNoModeledStation(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 0xB110)
	fixture.state.DiplomaticRelations = reciprocalWarRelations(fixture.blockaderEmpireID, fixture.targetEmpireID)
	addBlockadeTestFleet(fixture.state, fixture.blockaderEmpireID, core.StrategicFleetRoleCombat, fixture.targetSystemID)
	transportID := addInvasionTransport(fixture.state, fixture.blockaderEmpireID, fixture.targetSystemID)
	ctx := invasionTestContext(fixture.blockaderEmpireID, fixture.targetEmpireID)
	rules, _ := commandPointTestResolver(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}

	opportunity, err := resolver.prepareInvasionBoundary(ctx, fixture.state)
	if err != nil {
		t.Fatal(err)
	}
	if opportunity == nil || opportunity.ColonyID != fixture.targetColonyID || !reflect.DeepEqual(opportunity.EligibleTransportFleetIDs, []core.ID{transportID}) {
		t.Fatalf("unexpected invasion opportunity: %+v", opportunity)
	}

	colony := colonyByID(fixture.state, fixture.targetColonyID)
	colony.Buildings = []string{commandStationStarBase}
	blocked, err := resolver.prepareInvasionBoundary(ctx, fixture.state)
	if err != nil {
		t.Fatal(err)
	}
	if blocked != nil {
		t.Fatalf("modeled orbital station must block invasion: %+v", blocked)
	}
	colony.Buildings = nil

	addBlockadeTestFleet(fixture.state, fixture.targetEmpireID, core.StrategicFleetRoleCombat, fixture.targetSystemID)
	blocked, err = resolver.prepareInvasionBoundary(ctx, fixture.state)
	if err != nil {
		t.Fatal(err)
	}
	if blocked != nil {
		t.Fatalf("remaining hostile combat fleet must block invasion: %+v", blocked)
	}
}

func TestColonyMilitiaCountsOnlyAssimilatedOrganicPopulation(t *testing.T) {
	state := core.NewSmallFixture(0xB111)
	owner := state.Empires[0].ID
	other := state.NewID()
	state.Empires = append(state.Empires, core.Empire{ID: other, Name: "Other", RaceID: "human"})
	colony := &state.Colonies[0]
	colony.Population = core.PopulationState{Cohorts: []core.PopulationCohort{
		{OriginEmpireID: owner, LoyaltyEmpireID: owner, AssimilationState: core.PopulationAssimilated, Farmers: 9},
		{OriginEmpireID: other, LoyaltyEmpireID: other, AssimilationState: core.PopulationConquered, Workers: 20},
	}}
	colony.Population.Normalize()
	if got := colonyMilitia(colony); got != 1 {
		t.Fatalf("militia=%d want 1 from assimilated 9 only", got)
	}
}

func TestGroundInvasionCapturePacksSurvivingTransportsAndConquersPopulation(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 0xB112)
	fixture.state.DiplomaticRelations = reciprocalWarRelations(fixture.blockaderEmpireID, fixture.targetEmpireID)
	addBlockadeTestFleet(fixture.state, fixture.blockaderEmpireID, core.StrategicFleetRoleCombat, fixture.targetSystemID)
	firstTransport := addInvasionTransport(fixture.state, fixture.blockaderEmpireID, fixture.targetSystemID)
	secondTransport := addInvasionTransport(fixture.state, fixture.blockaderEmpireID, fixture.targetSystemID)
	colony := colonyByID(fixture.state, fixture.targetColonyID)
	colony.GroundForces.Infantry = 0
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectHousing, ProjectID: HousingProjectID}
	colony.Population = core.PopulationState{Cohorts: []core.PopulationCohort{
		{OriginEmpireID: fixture.targetEmpireID, LoyaltyEmpireID: fixture.targetEmpireID, AssimilationState: core.PopulationAssimilated, Farmers: 2},
		{OriginEmpireID: fixture.blockaderEmpireID, LoyaltyEmpireID: fixture.blockaderEmpireID, AssimilationState: core.PopulationConquered, Workers: 2},
	}}
	colony.Population.Normalize()
	ctx := invasionTestContext(fixture.blockaderEmpireID, fixture.targetEmpireID)
	rules, _ := commandPointTestResolver(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	opportunity, err := resolver.prepareInvasionBoundary(ctx, fixture.state)
	if err != nil {
		t.Fatal(err)
	}
	if opportunity == nil {
		t.Fatal("expected invasion opportunity")
	}
	command, err := NewInvadeCommand(1, InvadePayload{ColonyID: colony.ID, TransportFleetIDs: []core.ID{firstTransport, secondTransport}})
	if err != nil {
		t.Fatal(err)
	}
	events, err := resolver.ResolveInvasionCommand(ctx, fixture.state, *opportunity, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Kind != "empire.invasion_resolved" || events[1].Kind != "empire.colony_conquered" {
		t.Fatalf("events=%+v", events)
	}
	if colony.EmpireID != fixture.blockaderEmpireID || colony.GroundForces.Infantry != 4 || colony.Construction != nil {
		t.Fatalf("captured colony state=%+v", *colony)
	}
	if _, fleet := strategicFleetByID(fixture.state, firstTransport); fleet != nil {
		t.Fatalf("lower transport %d should be consumed", firstTransport)
	}
	if _, fleet := strategicFleetByID(fixture.state, secondTransport); fleet == nil {
		t.Fatalf("higher transport %d should survive reverse packing", secondTransport)
	}
	var sawDefenderConquered, sawAttackerAssimilated bool
	for _, cohort := range colony.Population.Cohorts {
		if cohort.OriginEmpireID == fixture.targetEmpireID && cohort.AssimilationState == core.PopulationConquered && cohort.LoyaltyEmpireID == fixture.targetEmpireID {
			sawDefenderConquered = true
		}
		if cohort.OriginEmpireID == fixture.blockaderEmpireID && cohort.AssimilationState == core.PopulationAssimilated && cohort.LoyaltyEmpireID == fixture.blockaderEmpireID {
			sawAttackerAssimilated = true
		}
	}
	if !sawDefenderConquered || !sawAttackerAssimilated {
		t.Fatalf("captured cohorts=%+v", colony.Population.Cohorts)
	}
	if fixture.state.DiplomaticStanceBetween(fixture.blockaderEmpireID, fixture.targetEmpireID) != core.DiplomaticStanceWar {
		t.Fatal("capture must leave war active")
	}
	if err := fixture.state.Validate(); err != nil {
		t.Fatalf("captured state invalid: %v", err)
	}
}

func TestGroundInvasionFailureConsumesSelectedTransportOnly(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 0xB113)
	fixture.state.DiplomaticRelations = reciprocalWarRelations(fixture.blockaderEmpireID, fixture.targetEmpireID)
	addBlockadeTestFleet(fixture.state, fixture.blockaderEmpireID, core.StrategicFleetRoleCombat, fixture.targetSystemID)
	selected := addInvasionTransport(fixture.state, fixture.blockaderEmpireID, fixture.targetSystemID)
	unselected := addInvasionTransport(fixture.state, fixture.blockaderEmpireID, fixture.targetSystemID)
	colony := colonyByID(fixture.state, fixture.targetColonyID)
	colony.GroundForces.Infantry = 100
	beforePopulation := colony.Population
	beforeConstruction := colony.Construction
	ctx := invasionTestContext(fixture.blockaderEmpireID, fixture.targetEmpireID)
	rules, _ := commandPointTestResolver(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	opportunity, err := resolver.prepareInvasionBoundary(ctx, fixture.state)
	if err != nil {
		t.Fatal(err)
	}
	command, err := NewInvadeCommand(1, InvadePayload{ColonyID: colony.ID, TransportFleetIDs: []core.ID{selected}})
	if err != nil {
		t.Fatal(err)
	}
	events, err := resolver.ResolveInvasionCommand(ctx, fixture.state, *opportunity, protocol.SeatID(1), command)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != "empire.invasion_resolved" {
		t.Fatalf("events=%+v", events)
	}
	if colony.EmpireID != fixture.targetEmpireID || !reflect.DeepEqual(colony.Population, beforePopulation) || colony.Construction != beforeConstruction {
		t.Fatalf("failed invasion changed ownership/population/construction")
	}
	if _, fleet := strategicFleetByID(fixture.state, selected); fleet != nil {
		t.Fatal("selected transport survived failed invasion")
	}
	if _, fleet := strategicFleetByID(fixture.state, unselected); fleet == nil {
		t.Fatal("unselected transport was consumed")
	}
	if colony.GroundForces.Infantry >= 100 {
		t.Fatalf("defender should have taken deterministic casualties, infantry=%d", colony.GroundForces.Infantry)
	}
	if err := fixture.state.Validate(); err != nil {
		t.Fatalf("failed-invasion state invalid: %v", err)
	}
}

func TestInvasionDeterministicReplaySaveLoadAndInvalidAtomicity(t *testing.T) {
	run := func(seed uint64) ([]byte, []DomainEvent) {
		fixture := newBlockadeTestFixture(t, seed)
		fixture.state.DiplomaticRelations = reciprocalWarRelations(fixture.blockaderEmpireID, fixture.targetEmpireID)
		addBlockadeTestFleet(fixture.state, fixture.blockaderEmpireID, core.StrategicFleetRoleCombat, fixture.targetSystemID)
		first := addInvasionTransport(fixture.state, fixture.blockaderEmpireID, fixture.targetSystemID)
		second := addInvasionTransport(fixture.state, fixture.blockaderEmpireID, fixture.targetSystemID)
		colony := colonyByID(fixture.state, fixture.targetColonyID)
		colony.GroundForces.Infantry = 1
		colony.Population = core.NewAssimilatedPopulation(fixture.targetEmpireID, 2, 1, 1)
		ctx := invasionTestContext(fixture.blockaderEmpireID, fixture.targetEmpireID)
		_, resolver := commandPointTestResolver(t)
		opportunity, err := resolver.prepareInvasionBoundary(ctx, fixture.state)
		if err != nil {
			t.Fatal(err)
		}
		command, err := NewInvadeCommand(1, InvadePayload{ColonyID: colony.ID, TransportFleetIDs: []core.ID{first, second}})
		if err != nil {
			t.Fatal(err)
		}
		events, err := resolver.ResolveInvasionCommand(ctx, fixture.state, *opportunity, 1, command)
		if err != nil {
			t.Fatal(err)
		}
		if colony.EmpireID != fixture.blockaderEmpireID {
			t.Fatal("replay fixture did not capture colony")
		}
		encoded, err := core.MarshalState(fixture.state)
		if err != nil {
			t.Fatal(err)
		}
		loaded, err := core.UnmarshalState(encoded)
		if err != nil {
			t.Fatal(err)
		}
		roundTrip, err := core.MarshalState(loaded)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(encoded, roundTrip) {
			t.Fatal("captured Core23 state changed across save/load")
		}
		return encoded, events
	}
	stateA, eventsA := run(0xB114)
	stateB, eventsB := run(0xB114)
	if !bytes.Equal(stateA, stateB) || !reflect.DeepEqual(eventsA, eventsB) {
		t.Fatalf("invasion replay diverged state_equal=%t events_equal=%t", bytes.Equal(stateA, stateB), reflect.DeepEqual(eventsA, eventsB))
	}

	fixture := newBlockadeTestFixture(t, 0xB115)
	fixture.state.DiplomaticRelations = reciprocalWarRelations(fixture.blockaderEmpireID, fixture.targetEmpireID)
	addBlockadeTestFleet(fixture.state, fixture.blockaderEmpireID, core.StrategicFleetRoleCombat, fixture.targetSystemID)
	first := addInvasionTransport(fixture.state, fixture.blockaderEmpireID, fixture.targetSystemID)
	second := addInvasionTransport(fixture.state, fixture.blockaderEmpireID, fixture.targetSystemID)
	ctx := invasionTestContext(fixture.blockaderEmpireID, fixture.targetEmpireID)
	_, resolver := commandPointTestResolver(t)
	opportunity, err := resolver.prepareInvasionBoundary(ctx, fixture.state)
	if err != nil {
		t.Fatal(err)
	}
	before, err := core.MarshalState(fixture.state)
	if err != nil {
		t.Fatal(err)
	}
	invalid, err := NewInvadeCommand(1, InvadePayload{ColonyID: fixture.targetColonyID, TransportFleetIDs: []core.ID{second, first}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.ResolveInvasionCommand(ctx, fixture.state, *opportunity, 1, invalid); err == nil {
		t.Fatal("unsorted transport IDs accepted")
	}
	after, err := core.MarshalState(fixture.state)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("invalid invasion mutated state or RNG")
	}
}

func TestCapturedOwnershipAndCohortsRoundTripExactly(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 0xB116)
	fixture.state.DiplomaticRelations = reciprocalWarRelations(fixture.blockaderEmpireID, fixture.targetEmpireID)
	addBlockadeTestFleet(fixture.state, fixture.blockaderEmpireID, core.StrategicFleetRoleCombat, fixture.targetSystemID)
	first := addInvasionTransport(fixture.state, fixture.blockaderEmpireID, fixture.targetSystemID)
	second := addInvasionTransport(fixture.state, fixture.blockaderEmpireID, fixture.targetSystemID)
	colony := colonyByID(fixture.state, fixture.targetColonyID)
	colony.GroundForces.Infantry = 0
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectHousing, ProjectID: HousingProjectID}
	colony.Population = core.PopulationState{Cohorts: []core.PopulationCohort{
		{OriginEmpireID: fixture.targetEmpireID, LoyaltyEmpireID: fixture.targetEmpireID, AssimilationState: core.PopulationAssimilated, Farmers: 2},
		{OriginEmpireID: fixture.blockaderEmpireID, LoyaltyEmpireID: fixture.blockaderEmpireID, AssimilationState: core.PopulationConquered, Workers: 2},
	}}
	colony.Population.Normalize()
	attacker := empireByID(fixture.state, fixture.blockaderEmpireID)
	defender := empireByID(fixture.state, fixture.targetEmpireID)
	attacker.Capital = 0
	defender.Capital = colony.ID

	ctx := invasionTestContext(fixture.blockaderEmpireID, fixture.targetEmpireID)
	_, resolver := commandPointTestResolver(t)
	opportunity, err := resolver.prepareInvasionBoundary(ctx, fixture.state)
	if err != nil {
		t.Fatal(err)
	}
	command, err := NewInvadeCommand(1, InvadePayload{ColonyID: colony.ID, TransportFleetIDs: []core.ID{first, second}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.ResolveInvasionCommand(ctx, fixture.state, *opportunity, 1, command); err != nil {
		t.Fatal(err)
	}

	encoded, err := core.MarshalState(fixture.state)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := core.UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	loadedColony := colonyByID(loaded, fixture.targetColonyID)
	if loadedColony == nil {
		t.Fatal("captured colony missing after round-trip")
	}
	if loadedColony.EmpireID != fixture.blockaderEmpireID || loadedColony.Construction != nil || loadedColony.GroundForces.Infantry != 4 {
		t.Fatalf("round-tripped captured colony=%+v", *loadedColony)
	}
	wantCohorts := []core.PopulationCohort{
		{OriginEmpireID: fixture.blockaderEmpireID, LoyaltyEmpireID: fixture.blockaderEmpireID, AssimilationState: core.PopulationAssimilated, Workers: 2},
		{OriginEmpireID: fixture.targetEmpireID, LoyaltyEmpireID: fixture.targetEmpireID, AssimilationState: core.PopulationConquered, Farmers: 2},
	}
	if !reflect.DeepEqual(loadedColony.Population.Cohorts, wantCohorts) {
		t.Fatalf("round-tripped cohorts=%+v want=%+v", loadedColony.Population.Cohorts, wantCohorts)
	}
	loadedAttacker := empireByID(loaded, fixture.blockaderEmpireID)
	loadedDefender := empireByID(loaded, fixture.targetEmpireID)
	wantDefenderCapital := replacementCapital(loaded, fixture.targetEmpireID)
	if loadedAttacker == nil || loadedDefender == nil || loadedAttacker.Capital != loadedColony.ID || loadedDefender.Capital != wantDefenderCapital || loadedDefender.Capital == loadedColony.ID {
		t.Fatalf("round-tripped capitals attacker=%+v defender=%+v want_defender=%d", loadedAttacker, loadedDefender, wantDefenderCapital)
	}
	if loaded.DiplomaticStanceBetween(fixture.blockaderEmpireID, fixture.targetEmpireID) != core.DiplomaticStanceWar {
		t.Fatal("round-tripped capture changed war stance")
	}
}
