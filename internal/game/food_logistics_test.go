package game

import (
	"encoding/json"
	"math"
	"testing"

	"moox/internal/core"
)

func twoColonyFoodFixture(t *testing.T, seed uint64) (*core.GameState, *core.Colony, *core.Colony) {
	t.Helper()
	state := core.NewSmallFixture(seed)
	home := &state.Colonies[0]
	home.Population = core.PopulationState{Total: 4, Farmers: 3, Scientists: 1}

	planet := &state.Galaxy.Systems[1].Planets[0] // small desert
	second := core.Colony{
		ID:         state.NewID(),
		EmpireID:   state.Empires[0].ID,
		PlanetID:   planet.ID,
		Population: core.PopulationState{Total: 2, Workers: 1, Scientists: 1},
	}
	planet.ColonyID = second.ID
	state.Colonies = append(state.Colonies, second)
	return state, &state.Colonies[0], &state.Colonies[1]
}

func findDomainEvent(events []DomainEvent, kind string) *DomainEvent {
	for i := range events {
		if events[i].Kind == kind {
			return &events[i]
		}
	}
	return nil
}

func TestFoodLogisticsTransfersFoodWithFreighters(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state, _, _ := twoColonyFoodFixture(t, 730)
	state.Empires[0].Freighters = 2

	result, err := resolver.Resolve(ResolveContext{}, state, nil)
	if err != nil {
		t.Fatal(err)
	}
	event := findDomainEvent(result.Events, "empire.food_logistics_resolved")
	if event == nil {
		t.Fatalf("missing food logistics event: %+v", result.Events)
	}
	var payload FoodLogisticsResolvedEvent
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Snapshot.FreightersRequired != 2 || payload.Snapshot.FreightersUsed != 2 {
		t.Fatalf("freighters required/used=%d/%d want 2/2", payload.Snapshot.FreightersRequired, payload.Snapshot.FreightersUsed)
	}
	if !closePopulationValue(payload.Snapshot.FoodTransferred, 2) || payload.Snapshot.FoodUnmet != 0 {
		t.Fatalf("food transfer snapshot=%+v", payload.Snapshot)
	}
	if payload.Snapshot.FreighterOperatingCostBC != 1 {
		t.Fatalf("operating cost=%v want=1 BC", payload.Snapshot.FreighterOperatingCostBC)
	}
	if len(payload.Colonies) != 2 || payload.Colonies[0].FoodExported != 2 || payload.Colonies[1].FoodImported != 2 {
		t.Fatalf("colony logistics=%+v", payload.Colonies)
	}
	if findDomainEvent(result.Events, "colony.population_starved") != nil {
		t.Fatalf("fully supplied colony should not starve: %+v", result.Events)
	}
}

func TestAllocateFoodImportsFollowsOriginalRoundRobinPriority(t *testing.T) {
	colonies := []*core.Colony{
		{ID: 10, PopulationDynamics: core.ColonyPopulationDynamics{LocalFoodShortage: 3}},
		{ID: 20, PopulationDynamics: core.ColonyPopulationDynamics{LocalFoodShortage: 1}},
		{ID: 30, PopulationDynamics: core.ColonyPopulationDynamics{LocalFoodShortage: 2}},
	}

	allocateFoodImports(colonies, 4, 1)

	want := []float64{2, 1, 1}
	for i, colony := range colonies {
		if !closePopulationValue(colony.PopulationDynamics.FoodImported, want[i]) {
			t.Fatalf("colony %d imported=%v want=%v", colony.ID, colony.PopulationDynamics.FoodImported, want[i])
		}
	}
}
func TestInsufficientFreightersLeaveShortageAndCauseStarvation(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state, _, sink := twoColonyFoodFixture(t, 731)
	state.Empires[0].Freighters = 1
	previous := sink.Population.Total

	result, err := resolver.Resolve(ResolveContext{}, state, nil)
	if err != nil {
		t.Fatal(err)
	}
	foodEvent := findDomainEvent(result.Events, "empire.food_logistics_resolved")
	if foodEvent == nil {
		t.Fatal("missing food logistics event")
	}
	var logistics FoodLogisticsResolvedEvent
	if err := json.Unmarshal(foodEvent.Data, &logistics); err != nil {
		t.Fatal(err)
	}
	if logistics.Snapshot.FreightersRequired != 2 || logistics.Snapshot.FreightersUsed != 1 || !closePopulationValue(logistics.Snapshot.FoodUnmet, 1) {
		t.Fatalf("unexpected constrained logistics: %+v", logistics.Snapshot)
	}
	starved := findDomainEvent(result.Events, "colony.population_starved")
	if starved == nil {
		t.Fatalf("expected starvation event: %+v", result.Events)
	}
	var payload PopulationStarvedEvent
	if err := json.Unmarshal(starved.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.ColonyID != sink.ID || payload.AppliedLoss <= 0 || payload.Current.Total >= previous {
		t.Fatalf("unexpected starvation payload: %+v", payload)
	}
	assigned := payload.Current.Farmers + payload.Current.Workers + payload.Current.Scientists
	if !closePopulationValue(assigned, payload.Current.Total) {
		t.Fatalf("starvation broke population allocation: %+v", payload.Current)
	}
}

func TestStarvationCannotEliminateLastPopulationUnit(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(732)
	state.Colonies[0].Population = core.PopulationState{Total: 1, Workers: 1}
	state.Empires[0].Freighters = 0
	result, err := resolver.Resolve(ResolveContext{}, state, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.State.Colonies[0].Population.Total != 1 {
		t.Fatalf("last population unit starved away: %+v", result.State.Colonies[0].Population)
	}
}

func TestCyberneticStarvationUsesFoodAndProductionShortage(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	planet := core.Planet{SizeID: "medium", ClimateID: "terran"}
	colony := core.Colony{Population: core.PopulationState{Total: 4, Farmers: 1, Scientists: 3}}
	adjusted := core.ColonyEconomy{Food: 1, Production: 1}
	d, err := rules.CalculatePopulationDynamics(colony, planet, "meklar", adjusted)
	if err != nil {
		t.Fatal(err)
	}
	if d.FoodShortage != 1 || d.ProductionShortage != 1 {
		t.Fatalf("cybernetic shortages=%+v", d)
	}
	base := math.Sqrt(rules.PopulationGrowthCurveFactor * 4 * (12 - 4) / 12)
	wantPenalty := 0.025 + 0.025
	wantStarvation := math.Max(0, wantPenalty-base)
	if !closePopulationValue(d.ProjectedStarvation, wantStarvation) {
		t.Fatalf("projected starvation=%v want=%v dynamics=%+v", d.ProjectedStarvation, wantStarvation, d)
	}
}

func TestSurplusFoodSaleAndFantasticTradersRate(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}

	regularState := core.NewSmallFixture(733)
	regularState.Colonies[0].Population = core.PopulationState{Total: 4, Farmers: 3, Scientists: 1}
	regular, err := resolver.Resolve(ResolveContext{}, regularState, nil)
	if err != nil {
		t.Fatal(err)
	}
	regularEvent := findDomainEvent(regular.Events, "empire.food_logistics_resolved")
	if regularEvent == nil {
		t.Fatal("missing regular food logistics event")
	}
	var regularPayload FoodLogisticsResolvedEvent
	if err := json.Unmarshal(regularEvent.Data, &regularPayload); err != nil {
		t.Fatal(err)
	}
	if !closePopulationValue(regularPayload.Snapshot.SurplusFoodIncomeBC, 1) {
		t.Fatalf("regular surplus income=%+v", regularPayload.Snapshot)
	}

	m := rules.RaceModifiers["human"]
	m.FantasticTraders = true
	rules.RaceModifiers["human"] = m
	tradersState := core.NewSmallFixture(734)
	tradersState.Colonies[0].Population = core.PopulationState{Total: 4, Farmers: 3, Scientists: 1}
	traders, err := resolver.Resolve(ResolveContext{}, tradersState, nil)
	if err != nil {
		t.Fatal(err)
	}
	tradersEvent := findDomainEvent(traders.Events, "empire.food_logistics_resolved")
	if tradersEvent == nil {
		t.Fatal("missing Fantastic Traders food logistics event")
	}
	var tradersPayload FoodLogisticsResolvedEvent
	if err := json.Unmarshal(tradersEvent.Data, &tradersPayload); err != nil {
		t.Fatal(err)
	}
	if !closePopulationValue(tradersPayload.Snapshot.SurplusFoodIncomeBC, 2) {
		t.Fatalf("Fantastic Traders surplus income=%+v", tradersPayload.Snapshot)
	}
}

func TestFoodLogisticsPreservesFractionalContinuousValues(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(736)
	state.Empires[0].Freighters = 2
	state.Colonies[0].PopulationDynamics.LocalFoodSurplus = 1.375
	state.Colonies[0].PopulationDynamics.FoodSurplus = 1.375

	planet := &state.Galaxy.Systems[1].Planets[0]
	second := core.Colony{
		ID:         state.NewID(),
		EmpireID:   state.Empires[0].ID,
		PlanetID:   planet.ID,
		Population: core.PopulationState{Total: 2, Workers: 1, Scientists: 1},
		PopulationDynamics: core.ColonyPopulationDynamics{
			Capacity:          10,
			LocalFoodShortage: 1.375,
			FoodShortage:      1.375,
		},
	}
	planet.ColonyID = second.ID
	state.Colonies = append(state.Colonies, second)

	events, err := resolver.materializeFoodLogistics(state, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != "empire.food_logistics_resolved" {
		t.Fatalf("unexpected logistics events: %+v", events)
	}
	var payload FoodLogisticsResolvedEvent
	if err := json.Unmarshal(events[0].Data, &payload); err != nil {
		t.Fatal(err)
	}
	if !closePopulationValue(payload.Snapshot.FoodTransferred, 1.375) || payload.Snapshot.FreightersUsed != 2 {
		t.Fatalf("fractional transfer was truncated: %+v", payload.Snapshot)
	}
	if payload.Snapshot.FreighterOperatingCostBC != 1 {
		t.Fatalf("freighter operating cost=%v want=1 BC", payload.Snapshot.FreighterOperatingCostBC)
	}
	if !closePopulationValue(payload.Colonies[0].FoodExported, 1.375) || !closePopulationValue(payload.Colonies[1].FoodImported, 1.375) {
		t.Fatalf("fractional colony transfer was truncated: %+v", payload.Colonies)
	}
}

func TestStarvationPreservesFractionalPopulationLoss(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(737)
	planet := state.Galaxy.Systems[0].Planets[0]
	colony := &state.Colonies[0]
	colony.Population = core.PopulationState{Total: 4, Farmers: 2, Workers: 1, Scientists: 1}

	dynamics, err := rules.CalculatePopulationDynamics(*colony, planet, "human", core.ColonyEconomy{Food: 2, Production: 3})
	if err != nil {
		t.Fatal(err)
	}
	if dynamics.ProjectedStarvation <= 0 || dynamics.ProjectedStarvation == float64(int(dynamics.ProjectedStarvation)) {
		t.Fatalf("expected fractional starvation loss, got %+v", dynamics)
	}
	colony.PopulationDynamics = dynamics
	before := colony.Population.Total
	events, err := resolver.advancePopulation(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != "colony.population_starved" {
		t.Fatalf("unexpected starvation events: %+v", events)
	}
	want := before - dynamics.ProjectedStarvation
	if !closePopulationValue(colony.Population.Total, want) {
		t.Fatalf("fractional starvation total=%v want=%v loss=%v", colony.Population.Total, want, dynamics.ProjectedStarvation)
	}
}
