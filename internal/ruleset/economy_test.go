package ruleset

import (
	"path/filepath"
	"testing"
)

func TestCommittedEconomyRulesLoadAndValidate(t *testing.T) {
	file, err := LoadEconomy(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31", "economy.json"))
	if err != nil {
		t.Fatal(err)
	}
	if file.BaseResearchPerScientist.Value != 3 || file.BaseTaxBCPerPopulation.Value != 1 {
		t.Fatalf("unexpected research/tax baseline: %d/%d", file.BaseResearchPerScientist.Value, file.BaseTaxBCPerPopulation.Value)
	}
	if file.AquaticFoodBonus.Value != 1 || len(file.AquaticFoodBonus.ClimateIDs) != 3 {
		t.Fatalf("unexpected aquatic food rule: %+v", file.AquaticFoodBonus)
	}
	want := []int{1, 2, 3, 5, 8}
	for i, item := range file.MineralIndustryPerWorker {
		if item.Value != want[i] {
			t.Fatalf("mineral industry[%d]=%d, want %d", i, item.Value, want[i])
		}
	}
}

func TestCommittedEconomyContextRules(t *testing.T) {
	file, err := LoadEconomy(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31", "economy.json"))
	if err != nil {
		t.Fatal(err)
	}
	gravity := make(map[string]int, len(file.GravityPenalties))
	for _, rule := range file.GravityPenalties {
		gravity[rule.RaceGravityID+"/"+rule.PlanetGravityID] = rule.Percent
	}
	wantGravity := map[string]int{
		"low_g/low_g":       0,
		"low_g/normal_g":    25,
		"low_g/heavy_g":     50,
		"normal_g/low_g":    25,
		"normal_g/normal_g": 0,
		"normal_g/heavy_g":  50,
		"heavy_g/low_g":     25,
		"heavy_g/normal_g":  0,
		"heavy_g/heavy_g":   0,
	}
	for key, want := range wantGravity {
		if got := gravity[key]; got != want {
			t.Fatalf("gravity %s=%d, want %d", key, got, want)
		}
	}

	government := make(map[string]GovernmentEconomyRule, len(file.GovernmentModifiers))
	for _, rule := range file.GovernmentModifiers {
		government[rule.TraitID] = rule
	}
	if got := government["government_feudal"].ResearchPercent; got != -50 {
		t.Fatalf("Feudal research percent=%d", got)
	}
	democracy := government["government_democracy"]
	if democracy.ResearchPercent != 50 || democracy.TaxPercent != 50 || democracy.TaxBonusRounding != "down" {
		t.Fatalf("unexpected Democracy economy rule: %+v", democracy)
	}
	unification := government["government_unification"]
	if unification.FoodPercent != 50 || unification.ProductionPercent != 50 || !unification.IgnoresMorale {
		t.Fatalf("unexpected Unification economy rule: %+v", unification)
	}
	dictatorship := government["government_dictatorship"]
	if dictatorship.FoodPercent != 0 || dictatorship.ProductionPercent != 0 || dictatorship.ResearchPercent != 0 || dictatorship.TaxPercent != 0 {
		t.Fatalf("unexpected Dictatorship economy rule: %+v", dictatorship)
	}
}

func TestCommittedMoraleRules(t *testing.T) {
	file, err := LoadEconomy(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31", "economy.json"))
	if err != nil {
		t.Fatal(err)
	}
	if file.Morale.BarracksPenaltyPercent != -20 {
		t.Fatalf("barracks morale penalty=%d, want -20", file.Morale.BarracksPenaltyPercent)
	}
	wantGovernments := map[string]bool{"government_feudal": true, "government_dictatorship": true}
	for _, traitID := range file.Morale.BarracksGovernmentTraitIDs {
		delete(wantGovernments, traitID)
	}
	if len(wantGovernments) != 0 {
		t.Fatalf("missing barracks morale governments: %v", wantGovernments)
	}
	wantBarracks := map[string]bool{"marine_barracks": true, "armor_barracks": true}
	for _, buildingID := range file.Morale.BarracksBuildingIDs {
		delete(wantBarracks, buildingID)
	}
	if len(wantBarracks) != 0 {
		t.Fatalf("missing barracks building ids: %v", wantBarracks)
	}
	bonuses := make(map[string]int)
	for _, bonus := range file.Morale.BuildingBonuses {
		bonuses[bonus.BuildingID] = bonus.Percent
	}
	if bonuses["holo_simulator"] != 20 || bonuses["pleasure_dome"] != 30 {
		t.Fatalf("unexpected morale building bonuses: %v", bonuses)
	}
}

func TestCommittedPopulationEconomyRules(t *testing.T) {
	file, err := LoadEconomy(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31", "economy.json"))
	if err != nil {
		t.Fatal(err)
	}
	population := file.Population
	if population.FoodPerPopulation != 1 || population.CyberneticFoodPerPopulation != 0.5 || population.CyberneticProductionPerPopulation != 0.5 {
		t.Fatalf("unexpected sustenance rules: %+v", population)
	}
	if population.GrowthCurveFactor != 0.002 || population.TolerantHabitabilityBonus != 0.25 || population.SubterraneanCapacityPerSizeClass != 2 {
		t.Fatalf("unexpected population growth/capacity rules: %+v", population)
	}
	if len(population.SizeCapacity) != 5 || len(population.ClimateHabitability) != 10 || len(population.SourceIDs) < 3 {
		t.Fatalf("population rule coverage incomplete: %+v", population)
	}
}

func TestCommittedFoodLogisticsRules(t *testing.T) {
	file, err := LoadEconomy(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31", "economy.json"))
	if err != nil {
		t.Fatal(err)
	}
	l := file.FoodLogistics
	if l.FreighterFoodCapacity != 1 || l.FreightersPerFleet != 5 || l.FreighterFleetCostPP != 50 || l.FreighterOperatingCostBC != 0.5 {
		t.Fatalf("unexpected freighter rules: %+v", l)
	}
	if l.SurplusFoodBCPerUnit != 0.5 || l.FantasticTradersSurplusFoodBCPerUnit != 1 {
		t.Fatalf("unexpected surplus-food sale rules: %+v", l)
	}
	if l.StarvationPopulationPerFoodShortage != 0.05 || l.CyberneticStarvationPopulationPerFoodShortage != 0.025 || l.CyberneticStarvationPopulationPerPPShortage != 0.025 || l.MinimumPopulationAfterStarvation != 1 {
		t.Fatalf("unexpected starvation rules: %+v", l)
	}
	if len(l.SourceIDs) < 4 {
		t.Fatalf("food logistics sources incomplete: %+v", l.SourceIDs)
	}
}
