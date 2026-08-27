package game

import (
	"path/filepath"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func loadCommittedEconomyRules(t *testing.T) *EconomyRules {
	t.Helper()
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	return rules
}

func TestCommittedEconomyRulesBaseValues(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	wantIndustry := map[string]int64{
		"ultra_poor": 1000,
		"poor":       2000,
		"abundant":   3000,
		"rich":       5000,
		"ultra_rich": 8000,
	}
	for id, want := range wantIndustry {
		if got := rules.MineralIndustryPerWorkerMilli[id]; got != want {
			t.Fatalf("mineral %s industry = %d, want %d", id, got, want)
		}
	}
	if got := rules.ClimateFoodPerFarmerMilli["terran"]; got != 2000 {
		t.Fatalf("terran food/farmer = %d", got)
	}
	if rules.BaseResearchPerScientistMilli != 3000 || rules.BaseTaxBCPerPopulationMilli != 1000 {
		t.Fatalf("unexpected research/tax baselines: %d/%d", rules.BaseResearchPerScientistMilli, rules.BaseTaxBCPerPopulationMilli)
	}
}

func TestRaceEconomyModifiersFromCommittedPresets(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	if got := rules.RaceModifiers["sakkra"].FoodPerFarmerMilli; got != 1000 {
		t.Fatalf("Sakkra food modifier = %d", got)
	}
	if got := rules.RaceModifiers["meklar"].ProductionPerWorkerMilli; got != 2000 {
		t.Fatalf("Meklar production modifier = %d", got)
	}
	if got := rules.RaceModifiers["psilon"].ResearchPerScientistMilli; got != 2000 {
		t.Fatalf("Psilon research modifier = %d", got)
	}
	if got := rules.RaceModifiers["gnolam"].TaxBCPerPopulationMilli; got != 1000 {
		t.Fatalf("Gnolam tax modifier = %d", got)
	}
	if !rules.RaceModifiers["trilarian"].Aquatic {
		t.Fatal("Trilarian preset did not carry Aquatic economy flag")
	}
}

func TestCalculateBaseEconomyUsesFixedPointRoleOutputs(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(91)
	colony := state.Colonies[0]
	planet := state.Galaxy.Systems[0].Planets[0]

	got, err := rules.CalculateBaseEconomy(colony, planet, "human")
	if err != nil {
		t.Fatal(err)
	}
	want := core.ColonyEconomy{FoodMilli: 4000, ProductionMilli: 3000, ResearchMilli: 3000, TaxBCMilli: 4000}
	if got != want {
		t.Fatalf("human base economy = %+v, want %+v", got, want)
	}

	colony.Population = core.PopulationState{Units: 4, Farmers: 1, Workers: 1, Scientists: 2}
	psilon, err := rules.CalculateBaseEconomy(colony, planet, "psilon")
	if err != nil {
		t.Fatal(err)
	}
	if psilon.ResearchMilli != 10000 {
		t.Fatalf("Psilon 2-scientist research = %d, want 10000", psilon.ResearchMilli)
	}
}

func TestEconomyResolverAssignsOwnedPopulationAndEmitsEvent(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(44)
	colonyID := state.Colonies[0].ID
	command, err := NewAssignPopulationCommand(1, AssignPopulationPayload{ColonyID: colonyID, Farmers: 1, Workers: 2, Scientists: 1})
	if err != nil {
		t.Fatal(err)
	}
	batch := protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-1",
		SeatID:        1,
		Turn:          1,
		BaseRevision:  1,
		Commands:      []protocol.Command{command},
	}
	result, err := resolver.Resolve(
		ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: state.Empires[0].ID}}},
		state,
		[]protocol.CommandBatch{batch},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := result.State.Colonies[0].Population; got != (core.PopulationState{Units: 4, Farmers: 1, Workers: 2, Scientists: 1}) {
		t.Fatalf("population assignment = %+v", got)
	}
	if got := result.State.Colonies[0].Economy; got != (core.ColonyEconomy{FoodMilli: 2000, ProductionMilli: 6000, ResearchMilli: 3000, TaxBCMilli: 4000}) {
		t.Fatalf("base economy = %+v", got)
	}
	if len(result.Events) != 1 || result.Events[0].Kind != "colony.population_assigned" || result.Events[0].SeatID != 1 || result.Events[0].CommandSequence != 1 {
		t.Fatalf("unexpected domain events: %+v", result.Events)
	}
}

func TestEconomyResolverRejectsForeignColonyAndBadTotals(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(45)
	colonyID := state.Colonies[0].ID
	command, err := NewAssignPopulationCommand(1, AssignPopulationPayload{ColonyID: colonyID, Farmers: 2, Workers: 1, Scientists: 1})
	if err != nil {
		t.Fatal(err)
	}
	batch := protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-1", SeatID: 2, Turn: 1, BaseRevision: 1, Commands: []protocol.Command{command}}
	if _, err := resolver.Resolve(ResolveContext{Seats: []SeatAuthority{{SeatID: 2, EmpireID: core.ID(999)}}}, state, []protocol.CommandBatch{batch}); err == nil {
		t.Fatal("expected foreign colony assignment to fail")
	}

	bad, err := NewAssignPopulationCommand(1, AssignPopulationPayload{ColonyID: colonyID, Farmers: 1, Workers: 1, Scientists: 1})
	if err != nil {
		t.Fatal(err)
	}
	batch.SeatID = 1
	batch.Commands = []protocol.Command{bad}
	if _, err := resolver.Resolve(ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: state.Empires[0].ID}}}, state, []protocol.CommandBatch{batch}); err == nil {
		t.Fatal("expected incomplete population assignment to fail")
	}
}

func TestFoodBonusesDoNotMakeNoFarmingPlanetFarmable(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(92)
	colony := state.Colonies[0]
	planet := state.Galaxy.Systems[0].Planets[0]
	planet.ClimateID = "barren"
	colony.Population = core.PopulationState{Units: 4, Farmers: 4}

	got, err := rules.CalculateBaseEconomy(colony, planet, "sakkra")
	if err != nil {
		t.Fatal(err)
	}
	if got.FoodMilli != 0 {
		t.Fatalf("Sakkra farming bonus made barren planet farmable: food=%d", got.FoodMilli)
	}
}

func TestAquaticFoodCoefficientAppliesOnlyToNormalizedWetClimates(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(93)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Units: 4, Farmers: 1, Workers: 1, Scientists: 2}
	planet := state.Galaxy.Systems[0].Planets[0]

	planet.ClimateID = "tundra"
	tundra, err := rules.CalculateBaseEconomy(colony, planet, "trilarian")
	if err != nil {
		t.Fatal(err)
	}
	if tundra.FoodMilli != 2000 {
		t.Fatalf("Aquatic tundra food=%d, want 2000", tundra.FoodMilli)
	}

	planet.ClimateID = "barren"
	barren, err := rules.CalculateBaseEconomy(colony, planet, "trilarian")
	if err != nil {
		t.Fatal(err)
	}
	if barren.FoodMilli != 0 {
		t.Fatalf("Aquatic barren food=%d, want 0", barren.FoodMilli)
	}
}

func TestPopulationIncomeRoundsToWholeBC(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(94)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Units: 1, Scientists: 1}
	planet := state.Galaxy.Systems[0].Planets[0]
	modifiers := rules.RaceModifiers["human"]
	modifiers.TaxBCPerPopulationMilli = 500
	rules.RaceModifiers["human"] = modifiers

	got, err := rules.CalculateBaseEconomy(colony, planet, "human")
	if err != nil {
		t.Fatal(err)
	}
	if got.TaxBCMilli != 2000 {
		t.Fatalf("1 population at 1.5 BC/pop rounded income=%d, want 2000", got.TaxBCMilli)
	}
}

func TestContextualEconomyHumanDemocracy(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(101)
	colony := state.Colonies[0]
	planet := state.Galaxy.Systems[0].Planets[0]
	base, err := rules.CalculateBaseEconomy(colony, planet, "human")
	if err != nil {
		t.Fatal(err)
	}
	context, adjusted, err := rules.CalculateContextualEconomy(base, colony, planet, "human")
	if err != nil {
		t.Fatal(err)
	}
	if context.RaceGravityID != "normal_g" || context.PlanetGravityID != "normal_g" || context.GravityPenaltyPercent != 0 {
		t.Fatalf("unexpected Human gravity context: %+v", context)
	}
	if context.GovernmentTraitID != "government_democracy" || context.GovernmentResearchPercent != 50 || context.GovernmentTaxPercent != 50 {
		t.Fatalf("unexpected Human government context: %+v", context)
	}
	want := core.ColonyEconomy{FoodMilli: 4000, ProductionMilli: 3000, ResearchMilli: 5000, TaxBCMilli: 6000}
	if adjusted != want {
		t.Fatalf("Human adjusted economy=%+v, want %+v", adjusted, want)
	}
}

func TestContextualEconomyLowGAndHeavyGMatrix(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(102)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Units: 1, Scientists: 1}
	colony.Buildings = []string{"marine_barracks"}
	planet := state.Galaxy.Systems[0].Planets[0]

	basePsilon, err := rules.CalculateBaseEconomy(colony, planet, "psilon")
	if err != nil {
		t.Fatal(err)
	}
	_, normalWorld, err := rules.CalculateContextualEconomy(basePsilon, colony, planet, "psilon")
	if err != nil {
		t.Fatal(err)
	}
	if normalWorld.ResearchMilli != 4000 {
		t.Fatalf("Low-G Psilon on Normal-G research=%d, want 4000", normalWorld.ResearchMilli)
	}

	planet.GravityID = "heavy_g"
	_, heavyWorld, err := rules.CalculateContextualEconomy(basePsilon, colony, planet, "psilon")
	if err != nil {
		t.Fatal(err)
	}
	if heavyWorld.ResearchMilli != 3000 {
		t.Fatalf("Low-G Psilon on Heavy-G research=%d, want 3000", heavyWorld.ResearchMilli)
	}

	planet.GravityID = "low_g"
	baseBulrathi, err := rules.CalculateBaseEconomy(colony, planet, "bulrathi")
	if err != nil {
		t.Fatal(err)
	}
	context, lowWorld, err := rules.CalculateContextualEconomy(baseBulrathi, colony, planet, "bulrathi")
	if err != nil {
		t.Fatal(err)
	}
	if context.GravityPenaltyPercent != 25 || lowWorld.ResearchMilli != 2000 {
		t.Fatalf("High-G Bulrathi on Low-G context=%+v research=%d", context, lowWorld.ResearchMilli)
	}
}

func TestContextualEconomyUnificationAppliesToFoodAndIndustry(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(103)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Units: 2, Farmers: 1, Workers: 1}
	planet := state.Galaxy.Systems[0].Planets[0]
	base, err := rules.CalculateBaseEconomy(colony, planet, "klackon")
	if err != nil {
		t.Fatal(err)
	}
	if base.FoodMilli != 3000 || base.ProductionMilli != 4000 {
		t.Fatalf("Klackon base economy=%+v", base)
	}
	context, adjusted, err := rules.CalculateContextualEconomy(base, colony, planet, "klackon")
	if err != nil {
		t.Fatal(err)
	}
	if context.GovernmentTraitID != "government_unification" || !context.GovernmentIgnoresMorale {
		t.Fatalf("unexpected Unification context: %+v", context)
	}
	if adjusted.FoodMilli != 5000 || adjusted.ProductionMilli != 6000 {
		t.Fatalf("Klackon adjusted economy=%+v", adjusted)
	}
}

func TestContextualEconomyFeudalResearchPenalty(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(104)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Units: 1, Scientists: 1}
	colony.Buildings = []string{"marine_barracks"}
	planet := state.Galaxy.Systems[0].Planets[0]
	base, err := rules.CalculateBaseEconomy(colony, planet, "sakkra")
	if err != nil {
		t.Fatal(err)
	}
	context, adjusted, err := rules.CalculateContextualEconomy(base, colony, planet, "sakkra")
	if err != nil {
		t.Fatal(err)
	}
	if context.GovernmentTraitID != "government_feudal" || adjusted.ResearchMilli != 2000 {
		t.Fatalf("Sakkra Feudal context=%+v adjusted=%+v", context, adjusted)
	}
}

func TestMoraleBarracksPenaltyAndRemoval(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(105)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Units: 1, Scientists: 1}
	planet := state.Galaxy.Systems[0].Planets[0]
	planet.GravityID = "low_g"
	base, err := rules.CalculateBaseEconomy(colony, planet, "psilon")
	if err != nil {
		t.Fatal(err)
	}
	context, adjusted, err := rules.CalculateContextualEconomy(base, colony, planet, "psilon")
	if err != nil {
		t.Fatal(err)
	}
	if context.MoraleBarracksPenaltyPercent != -20 || context.MoralePercent != -20 {
		t.Fatalf("unexpected missing-barracks morale context: %+v", context)
	}
	if adjusted.ResearchMilli != 4000 {
		t.Fatalf("Dictatorship missing-barracks research=%d, want 4000", adjusted.ResearchMilli)
	}

	colony.Buildings = []string{"marine_barracks"}
	context, adjusted, err = rules.CalculateContextualEconomy(base, colony, planet, "psilon")
	if err != nil {
		t.Fatal(err)
	}
	if context.MoraleBarracksPenaltyPercent != 0 || context.MoralePercent != 0 {
		t.Fatalf("Marine Barracks did not remove morale penalty: %+v", context)
	}
	if adjusted.ResearchMilli != 5000 {
		t.Fatalf("Dictatorship with barracks research=%d, want 5000", adjusted.ResearchMilli)
	}
}

func TestMoraleBuildingsAreCumulativeAndAffectMoneySeparately(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(106)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Units: 1, Scientists: 1}
	colony.Buildings = []string{"holo_simulator", "pleasure_dome"}
	planet := state.Galaxy.Systems[0].Planets[0]
	base, err := rules.CalculateBaseEconomy(colony, planet, "human")
	if err != nil {
		t.Fatal(err)
	}
	context, adjusted, err := rules.CalculateContextualEconomy(base, colony, planet, "human")
	if err != nil {
		t.Fatal(err)
	}
	if context.MoraleBuildingBonusPercent != 50 || context.MoralePercent != 50 {
		t.Fatalf("unexpected cumulative morale context: %+v", context)
	}
	if adjusted.ResearchMilli != 6000 {
		t.Fatalf("Democracy + morale research=%d, want 6000", adjusted.ResearchMilli)
	}
	if adjusted.TaxBCMilli != 2000 {
		t.Fatalf("Democracy + morale tax=%d, want 2000", adjusted.TaxBCMilli)
	}
}

func TestUnificationRecordsButIgnoresMorale(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(107)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Units: 2, Farmers: 1, Workers: 1}
	colony.Buildings = []string{"holo_simulator", "pleasure_dome"}
	planet := state.Galaxy.Systems[0].Planets[0]
	base, err := rules.CalculateBaseEconomy(colony, planet, "klackon")
	if err != nil {
		t.Fatal(err)
	}
	context, adjusted, err := rules.CalculateContextualEconomy(base, colony, planet, "klackon")
	if err != nil {
		t.Fatal(err)
	}
	if !context.GovernmentIgnoresMorale || context.MoraleBuildingBonusPercent != 50 || context.MoralePercent != 0 {
		t.Fatalf("Unification morale bypass context=%+v", context)
	}
	if adjusted.FoodMilli != 5000 || adjusted.ProductionMilli != 6000 {
		t.Fatalf("Unification applied ignored morale: %+v", adjusted)
	}
}

func TestContextualEconomyRejectsUnknownBuilding(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(108)
	colony := state.Colonies[0]
	colony.Buildings = []string{"not_a_real_building"}
	planet := state.Galaxy.Systems[0].Planets[0]
	base, err := rules.CalculateBaseEconomy(colony, planet, "human")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := rules.CalculateContextualEconomy(base, colony, planet, "human"); err == nil {
		t.Fatal("expected unknown colony building to fail contextual economy")
	}
}
