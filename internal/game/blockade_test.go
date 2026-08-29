package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
)

type blockadeTestFixture struct {
	state             *core.GameState
	blockaderEmpireID core.ID
	targetEmpireID    core.ID
	sourceColonyID    core.ID
	targetColonyID    core.ID
	targetSystemID    core.ID
}

func newBlockadeTestFixture(t *testing.T, seed uint64) blockadeTestFixture {
	t.Helper()
	state := core.NewSmallFixture(seed)
	blockaderEmpireID := state.Empires[0].ID
	targetEmpireID := state.NewID()
	targetEmpire := core.Empire{ID: targetEmpireID, Name: "Target Empire", RaceID: "human", Freighters: 10}

	targetPlanet := &state.Galaxy.Systems[1].Planets[0]
	targetColonyID := state.NewID()
	targetColony := core.Colony{
		ID:         targetColonyID,
		EmpireID:   targetEmpireID,
		PlanetID:   targetPlanet.ID,
		Population: core.NewAssimilatedPopulation(targetEmpireID, 0, 1, 1),
	}
	targetPlanet.ColonyID = targetColonyID

	sourcePlanet := &state.Galaxy.Systems[2].Planets[0]
	sourceColonyID := state.NewID()
	sourceColony := core.Colony{
		ID:         sourceColonyID,
		EmpireID:   targetEmpireID,
		PlanetID:   sourcePlanet.ID,
		Population: core.NewAssimilatedPopulation(targetEmpireID, 3, 0, 0),
	}
	sourcePlanet.ColonyID = sourceColonyID
	targetEmpire.Capital = sourceColonyID

	state.Empires = append(state.Empires, targetEmpire)
	state.Colonies = append(state.Colonies, targetColony, sourceColony)
	return blockadeTestFixture{
		state:             state,
		blockaderEmpireID: blockaderEmpireID,
		targetEmpireID:    targetEmpireID,
		sourceColonyID:    sourceColonyID,
		targetColonyID:    targetColonyID,
		targetSystemID:    state.Galaxy.Systems[1].ID,
	}
}

func addBlockadeTestFleet(state *core.GameState, empireID core.ID, role core.StrategicFleetRole, atSystemID core.ID) core.ID {
	fleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{ID: fleetID, EmpireID: empireID, Role: role, AtSystemID: atSystemID})
	return fleetID
}

func TestRecomputeSystemBlockadesUsesColonyPresenceCombatFleetAndDirectedHostility(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 1601)
	addBlockadeTestFleet(fixture.state, fixture.blockaderEmpireID, core.StrategicFleetRoleCombat, fixture.targetSystemID)
	fixture.state.DiplomaticRelations = []core.DiplomaticRelation{{
		FromEmpireID: fixture.blockaderEmpireID,
		ToEmpireID:   fixture.targetEmpireID,
		Stance:       core.DiplomaticStanceHostile,
	}}
	if err := fixture.state.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := recomputeSystemBlockades(fixture.state); err != nil {
		t.Fatal(err)
	}
	if got := fixture.state.Galaxy.Systems[1].BlockadedEmpireIDs; !reflect.DeepEqual(got, []core.ID{fixture.targetEmpireID}) {
		t.Fatalf("target system blockades=%v want=[%d]", got, fixture.targetEmpireID)
	}
	if got := fixture.state.Galaxy.Systems[0].BlockadedEmpireIDs; len(got) != 0 {
		t.Fatalf("reverse direction unexpectedly blockaded blockader colony: %v", got)
	}
}

func TestRecomputeSystemBlockadesIgnoresCivilianTransitAndSelfPresence(t *testing.T) {
	tests := []struct {
		name       string
		role       core.StrategicFleetRole
		atSystemID func(blockadeTestFixture) core.ID
	}{
		{name: "civilian", role: core.StrategicFleetRoleCivilian, atSystemID: func(f blockadeTestFixture) core.ID { return f.targetSystemID }},
		{name: "not at system", role: core.StrategicFleetRoleCombat, atSystemID: func(blockadeTestFixture) core.ID { return 0 }},
		{name: "combat at own colony only", role: core.StrategicFleetRoleCombat, atSystemID: func(f blockadeTestFixture) core.ID { return f.state.Galaxy.Systems[0].ID }},
	}
	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newBlockadeTestFixture(t, uint64(1610+i))
			addBlockadeTestFleet(fixture.state, fixture.blockaderEmpireID, test.role, test.atSystemID(fixture))
			fixture.state.DiplomaticRelations = []core.DiplomaticRelation{{
				FromEmpireID: fixture.blockaderEmpireID,
				ToEmpireID:   fixture.targetEmpireID,
				Stance:       core.DiplomaticStanceHostile,
			}}
			if err := recomputeSystemBlockades(fixture.state); err != nil {
				t.Fatal(err)
			}
			for _, system := range fixture.state.Galaxy.Systems {
				if len(system.BlockadedEmpireIDs) != 0 {
					t.Fatalf("system %d unexpectedly blockaded: %v", system.ID, system.BlockadedEmpireIDs)
				}
			}
		})
	}
}

func TestRecomputeSystemBlockadesSortsTargetsCombinesBlockadersAndClearsStaleState(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 1620)
	secondTargetID := fixture.state.NewID()
	secondTarget := core.Empire{ID: secondTargetID, Name: "Second Target", RaceID: "human"}
	secondTargetColonyID := fixture.state.NewID()
	secondTargetPlanet := &fixture.state.Galaxy.Systems[1].Planets[1]
	secondTargetColony := core.Colony{
		ID:         secondTargetColonyID,
		EmpireID:   secondTargetID,
		PlanetID:   secondTargetPlanet.ID,
		Population: core.NewAssimilatedPopulation(secondTargetID, 1, 0, 0),
	}
	secondTarget.Capital = secondTargetColonyID
	secondTargetPlanet.ColonyID = secondTargetColonyID
	fixture.state.Empires = append(fixture.state.Empires, secondTarget)
	fixture.state.Colonies = append(fixture.state.Colonies, secondTargetColony)

	thirdBlockaderID := fixture.state.NewID()
	fixture.state.Empires = append(fixture.state.Empires, core.Empire{ID: thirdBlockaderID, Name: "Third Blockader", RaceID: "human"})
	addBlockadeTestFleet(fixture.state, fixture.blockaderEmpireID, core.StrategicFleetRoleCombat, fixture.targetSystemID)
	addBlockadeTestFleet(fixture.state, thirdBlockaderID, core.StrategicFleetRoleCombat, fixture.targetSystemID)
	fixture.state.DiplomaticRelations = []core.DiplomaticRelation{
		{FromEmpireID: fixture.blockaderEmpireID, ToEmpireID: fixture.targetEmpireID, Stance: core.DiplomaticStanceHostile},
		{FromEmpireID: fixture.blockaderEmpireID, ToEmpireID: secondTargetID, Stance: core.DiplomaticStanceHostile},
		{FromEmpireID: thirdBlockaderID, ToEmpireID: fixture.targetEmpireID, Stance: core.DiplomaticStanceHostile},
	}
	if err := fixture.state.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := recomputeSystemBlockades(fixture.state); err != nil {
		t.Fatal(err)
	}
	want := []core.ID{fixture.targetEmpireID, secondTargetID}
	if got := fixture.state.Galaxy.Systems[1].BlockadedEmpireIDs; !reflect.DeepEqual(got, want) {
		t.Fatalf("blockades=%v want=%v", got, want)
	}

	for i := range fixture.state.StrategicFleets {
		fixture.state.StrategicFleets[i].AtSystemID = 0
	}
	fixture.state.Galaxy.Systems[1].BlockadedEmpireIDs = append([]core.ID(nil), want...)
	if err := recomputeSystemBlockades(fixture.state); err != nil {
		t.Fatal(err)
	}
	if got := fixture.state.Galaxy.Systems[1].BlockadedEmpireIDs; len(got) != 0 {
		t.Fatalf("stale blockade state was not cleared: %v", got)
	}
}

func TestResolveRecomputesBlockadesBeforeFinalFoodSnapshot(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	fixture := newBlockadeTestFixture(t, 1630)
	addBlockadeTestFleet(fixture.state, fixture.blockaderEmpireID, core.StrategicFleetRoleCombat, fixture.targetSystemID)
	fixture.state.DiplomaticRelations = []core.DiplomaticRelation{{
		FromEmpireID: fixture.blockaderEmpireID,
		ToEmpireID:   fixture.targetEmpireID,
		Stance:       core.DiplomaticStanceHostile,
	}}

	result, err := resolver.Resolve(ResolveContext{}, fixture.state, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := result.State.Galaxy.Systems[1].BlockadedEmpireIDs; !reflect.DeepEqual(got, []core.ID{fixture.targetEmpireID}) {
		t.Fatalf("final blockade state=%v want=[%d]", got, fixture.targetEmpireID)
	}

	var initialTargetFood *FoodLogisticsResolvedEvent
	for _, event := range result.Events {
		if event.Kind != "empire.food_logistics_resolved" {
			continue
		}
		var payload FoodLogisticsResolvedEvent
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.EmpireID == fixture.targetEmpireID {
			copy := payload
			initialTargetFood = &copy
			break
		}
	}
	if initialTargetFood == nil || initialTargetFood.Snapshot.FoodTransferred <= 0 {
		t.Fatalf("first Colony/Food pass should still use the previously materialized unblocked state: %+v", initialTargetFood)
	}

	target := colonyByID(result.State, fixture.targetColonyID)
	if target == nil {
		t.Fatal("target colony missing after resolution")
	}
	if target.PopulationDynamics.FoodImported != 0 || target.PopulationDynamics.LocalFoodShortage <= 0 {
		t.Fatalf("final Food snapshot did not consume recomputed blockade: %+v", target.PopulationDynamics)
	}
	targetEmpire := empireByID(result.State, fixture.targetEmpireID)
	if targetEmpire == nil || targetEmpire.FoodLogistics.BlockedFoodShortage <= 0 {
		t.Fatalf("target Empire final logistics did not record blocked shortage: %+v", targetEmpire)
	}
}

func TestResolveRecomputesBlockadesBeforePopulationTransferArrival(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	fixture := newBlockadeTestFixture(t, 1640)
	addBlockadeTestFleet(fixture.state, fixture.blockaderEmpireID, core.StrategicFleetRoleCombat, fixture.targetSystemID)
	fixture.state.DiplomaticRelations = []core.DiplomaticRelation{{
		FromEmpireID: fixture.blockaderEmpireID,
		ToEmpireID:   fixture.targetEmpireID,
		Stance:       core.DiplomaticStanceHostile,
	}}
	fixture.state.PopulationTransfers = []core.PopulationTransfer{{
		ID:                  fixture.state.NewID(),
		EmpireID:            fixture.targetEmpireID,
		SourceColonyID:      fixture.sourceColonyID,
		DestinationColonyID: fixture.targetColonyID,
		OriginEmpireID:      fixture.targetEmpireID,
		LoyaltyEmpireID:     fixture.targetEmpireID,
		AssimilationState:   core.PopulationAssimilated,
		Job:                 core.PopulationJobFarmer,
		RemainingTurns:      1,
	}}

	result, err := resolver.Resolve(ResolveContext{}, fixture.state, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.State.PopulationTransfers) != 0 {
		t.Fatalf("arriving Population transfer was not resolved: %+v", result.State.PopulationTransfers)
	}
	var lost *PopulationTransferLostEvent
	for _, event := range result.Events {
		if event.Kind != "empire.population_transfer_lost" {
			continue
		}
		var payload PopulationTransferLostEvent
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.EmpireID == fixture.targetEmpireID && payload.DestinationColonyID == fixture.targetColonyID {
			copy := payload
			lost = &copy
			break
		}
	}
	if lost == nil || lost.Reason != "destination_blockaded" {
		t.Fatalf("Population transfer should be lost to the freshly recomputed blockade: %+v", lost)
	}
}
