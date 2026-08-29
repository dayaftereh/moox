package game

import (
	"encoding/json"
	"testing"

	"moox/internal/core"
)

func addPopulationTestEmpire(state *core.GameState, raceID string) core.ID {
	id := state.NewID()
	state.Empires = append(state.Empires, core.Empire{ID: id, Name: "Foreign " + raceID, RaceID: raceID})
	return id
}

func TestRaceAwareCohortEconomyUsesOriginRaceAndConqueredFactor(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1501)
	owner := &state.Empires[0]
	foreignID := addPopulationTestEmpire(state, "sakkra")
	colony := &state.Colonies[0]
	planet := planetByID(state, colony.PlanetID)
	colony.Population = core.PopulationState{Cohorts: []core.PopulationCohort{
		{OriginEmpireID: owner.ID, LoyaltyEmpireID: owner.ID, AssimilationState: core.PopulationAssimilated, Farmers: 1, Workers: 1, Scientists: 1},
		{OriginEmpireID: foreignID, LoyaltyEmpireID: owner.ID, AssimilationState: core.PopulationAssimilated, Farmers: 1},
		{OriginEmpireID: foreignID, LoyaltyEmpireID: foreignID, AssimilationState: core.PopulationConquered, Workers: 1},
	}}
	colony.Population.Normalize()
	if err := resolver.recalculateColony(state, colony); err != nil {
		t.Fatal(err)
	}

	ownerBase, err := rules.CalculateCohortBaseEconomy(colony.Population.Cohorts[0], *planet, owner.RaceID)
	if err != nil {
		t.Fatal(err)
	}
	foreignAssim, err := rules.CalculateCohortBaseEconomy(colony.Population.Cohorts[1], *planet, "sakkra")
	if err != nil {
		t.Fatal(err)
	}
	foreignConquered, err := rules.CalculateCohortBaseEconomy(colony.Population.Cohorts[2], *planet, "sakkra")
	if err != nil {
		t.Fatal(err)
	}
	if foreignConquered.Production <= 0 || !closePopulationValue(foreignConquered.Production, foreignAssimulatedProductionForTest(rules, *planet)*conqueredOrganicOutputFactor) {
		t.Fatalf("conquered production=%v did not use 0.75 factor", foreignConquered.Production)
	}
	if !closePopulationValue(colony.Economy.Food, ownerBase.Food+foreignAssim.Food+foreignConquered.Food) ||
		!closePopulationValue(colony.Economy.Production, ownerBase.Production+foreignAssim.Production+foreignConquered.Production) ||
		!closePopulationValue(colony.Economy.Research, ownerBase.Research+foreignAssim.Research+foreignConquered.Research) {
		t.Fatalf("mixed base economy=%+v owner=%+v assimilated=%+v conquered=%+v", colony.Economy, ownerBase, foreignAssim, foreignConquered)
	}
	wantTax, err := rules.CalculateOwnerTaxBase(colony.Population.Total(), owner.RaceID)
	if err != nil {
		t.Fatal(err)
	}
	if !closePopulationValue(colony.Economy.TaxBC, wantTax) {
		t.Fatalf("owner-wide tax=%v want=%v", colony.Economy.TaxBC, wantTax)
	}
}

func foreignAssimulatedProductionForTest(rules *EconomyRules, planet core.Planet) float64 {
	cohort := core.PopulationCohort{OriginEmpireID: 2, LoyaltyEmpireID: 1, AssimilationState: core.PopulationAssimilated, Workers: 1}
	base, err := rules.CalculateCohortBaseEconomy(cohort, planet, "sakkra")
	if err != nil {
		return 0
	}
	return base.Production
}

func TestFourPassFoodTargetsPrioritizeOwnerAssimilatedConqueredAndRounding(t *testing.T) {
	d := core.ColonyPopulationDynamics{
		OwnerOriginFood2:        2,
		AssimilatedForeignFood2: 2,
		ConqueredForeignFood2:   2,
		WholeFoodRequired:       4,
	}
	want := []float64{1, 2, 3, 4}
	for pass, expected := range want {
		if got := foodImportPassTarget(d, pass+1); got != expected {
			t.Fatalf("pass %d target=%v want=%v", pass+1, got, expected)
		}
	}

	colonies := []*core.Colony{
		{ID: 10, AdjustedEconomy: core.ColonyEconomy{}, PopulationDynamics: core.ColonyPopulationDynamics{OwnerOriginFood2: 2, AssimilatedForeignFood2: 2, ConqueredForeignFood2: 2, WholeFoodRequired: 3, LocalFoodShortage: 3}},
		{ID: 20, AdjustedEconomy: core.ColonyEconomy{}, PopulationDynamics: core.ColonyPopulationDynamics{OwnerOriginFood2: 2, AssimilatedForeignFood2: 2, ConqueredForeignFood2: 2, WholeFoodRequired: 3, LocalFoodShortage: 3}},
	}
	allocateFoodImports(colonies, 3, 1)
	if colonies[0].PopulationDynamics.FoodImported != 2 || colonies[1].PopulationDynamics.FoodImported != 1 {
		t.Fatalf("three loads must feed both owner passes before first foreign pass: %+v %+v", colonies[0].PopulationDynamics, colonies[1].PopulationDynamics)
	}

	cyberHalf := []*core.Colony{{
		ID:                 30,
		PopulationDynamics: core.ColonyPopulationDynamics{OwnerOriginFood2: 1, WholeFoodRequired: 1, LocalFoodShortage: 1},
	}}
	allocateFoodImports(cyberHalf, 1, 1)
	if cyberHalf[0].PopulationDynamics.FoodImported != 1 {
		t.Fatalf("pass 4 must fill final half-Food rounding to one whole Food: %+v", cyberHalf[0].PopulationDynamics)
	}
}

func TestHeterogeneousCapacityTrimUsesOriginalOrganicPriority(t *testing.T) {
	ownerID := core.ID(1)
	foreignID := core.ID(2)
	population := core.PopulationState{Cohorts: []core.PopulationCohort{
		{OriginEmpireID: ownerID, LoyaltyEmpireID: ownerID, AssimilationState: core.PopulationAssimilated, Farmers: 4},
		{OriginEmpireID: foreignID, LoyaltyEmpireID: ownerID, AssimilationState: core.PopulationAssimilated, Workers: 3},
		{OriginEmpireID: foreignID, LoyaltyEmpireID: foreignID, AssimilationState: core.PopulationConquered, Scientists: 3},
	}}
	population.Normalize()
	capacities := map[core.ID]float64{ownerID: 10, foreignID: 5}
	removed := trimPopulationToHeterogeneousCapacity(&population, ownerID, capacities)
	if !closePopulationValue(removed, 1) {
		t.Fatalf("removed=%v want=1", removed)
	}
	conquered, ok := population.CohortBySemanticKey(core.PopulationCohortKey{OriginEmpireID: foreignID, LoyaltyEmpireID: foreignID, AssimilationState: core.PopulationConquered})
	if !ok || !closePopulationValue(conquered.Total(), 2) {
		t.Fatalf("conquered foreign must be trimmed first: %+v", population)
	}
	assimilated, _ := population.CohortBySemanticKey(core.PopulationCohortKey{OriginEmpireID: foreignID, LoyaltyEmpireID: ownerID, AssimilationState: core.PopulationAssimilated})
	if !closePopulationValue(assimilated.Total(), 3) {
		t.Fatalf("assimilated foreign changed before conquered exhausted: %+v", population)
	}
	if !populationFitsHeterogeneousCapacity(population, capacities) {
		t.Fatalf("trimmed population still violates heterogeneous capacity: %+v", population)
	}
}

func TestRaceAwareGrowthCreatesAssimilatedForeignCohort(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, _ := NewEconomyResolver(rules)
	state := core.NewSmallFixture(1502)
	owner := &state.Empires[0]
	foreignID := addPopulationTestEmpire(state, "sakkra")
	colony := &state.Colonies[0]
	colony.Population = core.PopulationState{Cohorts: []core.PopulationCohort{
		{OriginEmpireID: owner.ID, LoyaltyEmpireID: owner.ID, AssimilationState: core.PopulationAssimilated, Farmers: 2},
		{OriginEmpireID: foreignID, LoyaltyEmpireID: foreignID, AssimilationState: core.PopulationConquered, Workers: 2},
	}}
	colony.Population.Normalize()
	planet := planetByID(state, colony.PlanetID)
	dynamics, err := resolver.calculateRaceAwarePopulationDynamics(state, *colony, *planet, *owner, core.ColonyEconomy{Food: 100, Production: 100})
	if err != nil {
		t.Fatal(err)
	}
	var foreignDynamics *core.PopulationOriginDynamics
	for i := range dynamics.Origins {
		if dynamics.Origins[i].OriginEmpireID == foreignID {
			foreignDynamics = &dynamics.Origins[i]
			break
		}
	}
	if foreignDynamics == nil || foreignDynamics.ProjectedGrowth <= 0 {
		t.Fatalf("missing foreign projected growth: %+v", dynamics.Origins)
	}
	if want := rules.RaceModifiers["sakkra"].PopulationGrowthMultiplier; !closePopulationValue(foreignDynamics.GrowthMultiplier, want) {
		t.Fatalf("foreign growth multiplier=%v want source-race=%v", foreignDynamics.GrowthMultiplier, want)
	}
	colony.PopulationDynamics = dynamics
	beforeConquered, _ := colony.Population.CohortBySemanticKey(core.PopulationCohortKey{OriginEmpireID: foreignID, LoyaltyEmpireID: foreignID, AssimilationState: core.PopulationConquered})
	if _, err := resolver.advancePopulation(state); err != nil {
		t.Fatal(err)
	}
	assimilated, ok := colony.Population.CohortBySemanticKey(core.PopulationCohortKey{OriginEmpireID: foreignID, LoyaltyEmpireID: owner.ID, AssimilationState: core.PopulationAssimilated})
	if !ok || assimilated.Total() <= 0 {
		t.Fatalf("foreign growth did not land in assimilated cohort: %+v", colony.Population)
	}
	afterConquered, _ := colony.Population.CohortBySemanticKey(core.PopulationCohortKey{OriginEmpireID: foreignID, LoyaltyEmpireID: foreignID, AssimilationState: core.PopulationConquered})
	if !closePopulationValue(beforeConquered.Total(), afterConquered.Total()) {
		t.Fatalf("growth must not enlarge conquered cohort: before=%+v after=%+v", beforeConquered, afterConquered)
	}
}

func TestOriginStarvationRemovesNonFarmersBeforeFarmersAcrossStatuses(t *testing.T) {
	originID := core.ID(2)
	population := core.PopulationState{Cohorts: []core.PopulationCohort{
		{OriginEmpireID: originID, LoyaltyEmpireID: 1, AssimilationState: core.PopulationAssimilated, Farmers: 1, Workers: 1},
		{OriginEmpireID: originID, LoyaltyEmpireID: originID, AssimilationState: core.PopulationConquered, Scientists: 1},
	}}
	population.Normalize()
	if err := applyOriginStarvation(&population, originID, 1); err != nil {
		t.Fatal(err)
	}
	if !closePopulationValue(population.Farmers(), 1) || !closePopulationValue(population.Workers()+population.Scientists(), 1) {
		t.Fatalf("starvation did not remove non-Farmers first: %+v", population)
	}
}

func TestTransferSelectorRejectsAmbiguousAndConqueredCohorts(t *testing.T) {
	ownerID := core.ID(1)
	foreignID := core.ID(2)
	population := core.PopulationState{Cohorts: []core.PopulationCohort{
		{OriginEmpireID: ownerID, LoyaltyEmpireID: ownerID, AssimilationState: core.PopulationAssimilated, Farmers: 2},
		{OriginEmpireID: foreignID, LoyaltyEmpireID: ownerID, AssimilationState: core.PopulationAssimilated, Farmers: 1},
		{OriginEmpireID: foreignID, LoyaltyEmpireID: foreignID, AssimilationState: core.PopulationConquered, Workers: 1},
	}}
	population.Normalize()
	if _, err := resolveTransferCohortKey(population, TransferPopulationPayload{Job: core.PopulationJobFarmer}, ownerID); err == nil {
		t.Fatal("same-job mixed-origin transfer without cohort selector must be rejected as ambiguous")
	}
	foreignAssim := core.PopulationCohortKey{OriginEmpireID: foreignID, LoyaltyEmpireID: ownerID, AssimilationState: core.PopulationAssimilated}
	if got, err := resolveTransferCohortKey(population, TransferPopulationPayload{Cohort: &foreignAssim, Job: core.PopulationJobFarmer}, ownerID); err != nil || got != foreignAssim {
		t.Fatalf("explicit assimilated foreign selector failed: got=%+v err=%v", got, err)
	}
	foreignConquered := core.PopulationCohortKey{OriginEmpireID: foreignID, LoyaltyEmpireID: foreignID, AssimilationState: core.PopulationConquered}
	if _, err := resolveTransferCohortKey(population, TransferPopulationPayload{Cohort: &foreignConquered, Job: core.PopulationJobWorker}, ownerID); err == nil {
		t.Fatal("conquered cohort transfer must be rejected")
	}
}

func TestSameSystemForeignTransferPreservesOriginAndEventIdentity(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, _ := NewEconomyResolver(rules)
	state := core.NewSmallFixture(1503)
	ownerID := state.Empires[0].ID
	foreignID := addPopulationTestEmpire(state, "sakkra")
	home := &state.Colonies[0]
	home.Population = core.PopulationState{Cohorts: []core.PopulationCohort{
		{OriginEmpireID: ownerID, LoyaltyEmpireID: ownerID, AssimilationState: core.PopulationAssimilated, Farmers: 2},
		{OriginEmpireID: foreignID, LoyaltyEmpireID: ownerID, AssimilationState: core.PopulationAssimilated, Farmers: 1},
	}}
	home.Population.Normalize()
	system := &state.Galaxy.Systems[0]
	planet := core.Planet{ID: state.NewID(), Name: "Alpha II", Orbit: 2, SizeID: "medium", MineralID: "abundant", GravityID: "normal_g", ClimateID: "terran"}
	destination := core.Colony{ID: state.NewID(), EmpireID: ownerID, PlanetID: planet.ID, Population: core.NewAssimilatedPopulation(ownerID, 0, 2, 0)}
	planet.ColonyID = destination.ID
	system.Planets = append(system.Planets, planet)
	state.Colonies = append(state.Colonies, destination)
	key := core.PopulationCohortKey{OriginEmpireID: foreignID, LoyaltyEmpireID: ownerID, AssimilationState: core.PopulationAssimilated}
	command, err := NewTransferPopulationCommand(1, TransferPopulationPayload{SourceColonyID: home.ID, DestinationColonyID: destination.ID, Cohort: &key, Job: core.PopulationJobFarmer})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.transferPopulation(state, ownerID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	arrived, ok := state.Colonies[1].Population.CohortBySemanticKey(key)
	if !ok || !closePopulationValue(arrived.Farmers, 1) {
		t.Fatalf("destination did not preserve foreign origin: %+v", state.Colonies[1].Population)
	}
	var payload PopulationTransferredEvent
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Cohort != key {
		t.Fatalf("transfer event cohort=%+v want=%+v", payload.Cohort, key)
	}
}
