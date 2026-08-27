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
	wantIndustry := map[string]float64{
		"ultra_poor": 1,
		"poor":       2,
		"abundant":   3,
		"rich":       5,
		"ultra_rich": 8,
	}
	for id, want := range wantIndustry {
		if got := rules.MineralIndustryPerWorker[id]; got != want {
			t.Fatalf("mineral %s industry = %v, want %v", id, got, want)
		}
	}
	if got := rules.ClimateFoodPerFarmer["terran"]; got != 2 {
		t.Fatalf("terran food/farmer = %v", got)
	}
	if rules.BaseResearchPerScientist != 3 || rules.BaseTaxBCPerPopulation != 1 {
		t.Fatalf("unexpected research/tax baselines: %v/%v", rules.BaseResearchPerScientist, rules.BaseTaxBCPerPopulation)
	}
}

func TestRaceEconomyModifiersFromCommittedPresets(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	if got := rules.RaceModifiers["sakkra"].FoodPerFarmer; got != 1 {
		t.Fatalf("Sakkra food modifier = %v", got)
	}
	if got := rules.RaceModifiers["meklar"].ProductionPerWorker; got != 2 {
		t.Fatalf("Meklar production modifier = %v", got)
	}
	if got := rules.RaceModifiers["psilon"].ResearchPerScientist; got != 2 {
		t.Fatalf("Psilon research modifier = %v", got)
	}
	if got := rules.RaceModifiers["gnolam"].TaxBCPerPopulation; got != 1 {
		t.Fatalf("Gnolam tax modifier = %v", got)
	}
	if !rules.RaceModifiers["trilarian"].Aquatic {
		t.Fatal("Trilarian preset did not carry Aquatic economy flag")
	}
}

func TestCalculateBaseEconomyUsesDomainNativeRoleOutputs(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(91)
	colony := state.Colonies[0]
	planet := state.Galaxy.Systems[0].Planets[0]

	got, err := rules.CalculateBaseEconomy(colony, planet, "human")
	if err != nil {
		t.Fatal(err)
	}
	want := core.ColonyEconomy{Food: 4, Production: 3, Research: 3, TaxBC: 4}
	if got != want {
		t.Fatalf("human base economy = %+v, want %+v", got, want)
	}

	colony.Population = core.PopulationState{Total: 4, Farmers: 1, Workers: 1, Scientists: 2}
	psilon, err := rules.CalculateBaseEconomy(colony, planet, "psilon")
	if err != nil {
		t.Fatal(err)
	}
	if psilon.Research != 10 {
		t.Fatalf("Psilon 2-scientist research = %v, want 10", psilon.Research)
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
	if got := result.State.Colonies[0].Population; got != (core.PopulationState{Total: 4, Farmers: 1, Workers: 2, Scientists: 1}) {
		t.Fatalf("population assignment = %+v", got)
	}
	if got := result.State.Colonies[0].Economy; got != (core.ColonyEconomy{Food: 2, Production: 6, Research: 3, TaxBC: 4}) {
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
	colony.Population = core.PopulationState{Total: 4, Farmers: 4}

	got, err := rules.CalculateBaseEconomy(colony, planet, "sakkra")
	if err != nil {
		t.Fatal(err)
	}
	if got.Food != 0 {
		t.Fatalf("Sakkra farming bonus made barren planet farmable: food=%v", got.Food)
	}
}

func TestAquaticFoodCoefficientAppliesOnlyToNormalizedWetClimates(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(93)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Total: 4, Farmers: 1, Workers: 1, Scientists: 2}
	planet := state.Galaxy.Systems[0].Planets[0]

	planet.ClimateID = "tundra"
	tundra, err := rules.CalculateBaseEconomy(colony, planet, "trilarian")
	if err != nil {
		t.Fatal(err)
	}
	if tundra.Food != 2 {
		t.Fatalf("Aquatic tundra food=%v, want 2", tundra.Food)
	}

	planet.ClimateID = "barren"
	barren, err := rules.CalculateBaseEconomy(colony, planet, "trilarian")
	if err != nil {
		t.Fatal(err)
	}
	if barren.Food != 0 {
		t.Fatalf("Aquatic barren food=%v, want 0", barren.Food)
	}
}

func TestPopulationIncomePreservesFractionalBC(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(94)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Total: 1, Scientists: 1}
	planet := state.Galaxy.Systems[0].Planets[0]
	modifiers := rules.RaceModifiers["human"]
	modifiers.TaxBCPerPopulation = 0.5
	rules.RaceModifiers["human"] = modifiers

	got, err := rules.CalculateBaseEconomy(colony, planet, "human")
	if err != nil {
		t.Fatal(err)
	}
	if got.TaxBC != 1.5 {
		t.Fatalf("1 population at 1.5 BC/pop income=%v, want 1.5", got.TaxBC)
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
	want := core.ColonyEconomy{Food: 4, Production: 3, Research: 4.5, TaxBC: 6}
	if adjusted != want {
		t.Fatalf("Human adjusted economy=%+v, want %+v", adjusted, want)
	}
}

func TestContextualEconomyLowGAndHeavyGMatrix(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(102)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Total: 1, Scientists: 1}
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
	if normalWorld.Research != 3.75 {
		t.Fatalf("Low-G Psilon on Normal-G research=%v, want 3.75", normalWorld.Research)
	}

	planet.GravityID = "heavy_g"
	_, heavyWorld, err := rules.CalculateContextualEconomy(basePsilon, colony, planet, "psilon")
	if err != nil {
		t.Fatal(err)
	}
	if heavyWorld.Research != 2.5 {
		t.Fatalf("Low-G Psilon on Heavy-G research=%v, want 2.5", heavyWorld.Research)
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
	if context.GravityPenaltyPercent != 25 || lowWorld.Research != 2.25 {
		t.Fatalf("High-G Bulrathi on Low-G context=%+v research=%v", context, lowWorld.Research)
	}
}

func TestContextualEconomyUnificationAppliesToFoodAndIndustry(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(103)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Total: 2, Farmers: 1, Workers: 1}
	planet := state.Galaxy.Systems[0].Planets[0]
	base, err := rules.CalculateBaseEconomy(colony, planet, "klackon")
	if err != nil {
		t.Fatal(err)
	}
	if base.Food != 3 || base.Production != 4 {
		t.Fatalf("Klackon base economy=%+v", base)
	}
	context, adjusted, err := rules.CalculateContextualEconomy(base, colony, planet, "klackon")
	if err != nil {
		t.Fatal(err)
	}
	if context.GovernmentTraitID != "government_unification" || !context.GovernmentIgnoresMorale {
		t.Fatalf("unexpected Unification context: %+v", context)
	}
	if adjusted.Food != 4.5 || adjusted.Production != 6 {
		t.Fatalf("Klackon adjusted economy=%+v", adjusted)
	}
}

func TestContextualEconomyFeudalResearchPenalty(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(104)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Total: 1, Scientists: 1}
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
	if context.GovernmentTraitID != "government_feudal" || adjusted.Research != 1.5 {
		t.Fatalf("Sakkra Feudal context=%+v adjusted=%+v", context, adjusted)
	}
}

func TestMoraleBarracksPenaltyAndRemoval(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(105)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Total: 1, Scientists: 1}
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
	if adjusted.Research != 4 {
		t.Fatalf("Dictatorship missing-barracks research=%v, want 4", adjusted.Research)
	}

	colony.Buildings = []string{"marine_barracks"}
	context, adjusted, err = rules.CalculateContextualEconomy(base, colony, planet, "psilon")
	if err != nil {
		t.Fatal(err)
	}
	if context.MoraleBarracksPenaltyPercent != 0 || context.MoralePercent != 0 {
		t.Fatalf("Marine Barracks did not remove morale penalty: %+v", context)
	}
	if adjusted.Research != 5 {
		t.Fatalf("Dictatorship with barracks research=%v, want 5", adjusted.Research)
	}
}

func TestMoraleBuildingsAreCumulativeAndAffectMoneySeparately(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(106)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Total: 1, Scientists: 1}
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
	if adjusted.Research != 6 {
		t.Fatalf("Democracy + morale research=%v, want 6", adjusted.Research)
	}
	if adjusted.TaxBC != 1.5 {
		t.Fatalf("Democracy + morale tax=%v, want 1.5", adjusted.TaxBC)
	}
}

func TestUnificationRecordsButIgnoresMorale(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(107)
	colony := state.Colonies[0]
	colony.Population = core.PopulationState{Total: 2, Farmers: 1, Workers: 1}
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
	if adjusted.Food != 4.5 || adjusted.Production != 6 {
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
