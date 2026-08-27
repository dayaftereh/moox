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
