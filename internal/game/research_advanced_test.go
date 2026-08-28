package game

import (
	"path/filepath"
	"testing"
)

func TestAdvancedResearchRuntimeMetadata(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(rules.TechnologyAIClasses); got != 41 {
		t.Fatalf("technology AI classes=%d want=41", got)
	}
	if got := len(rules.TechnologyAIFieldGroupValues); got != 23 {
		t.Fatalf("field-group values=%d want=23", got)
	}
	if got := rules.TechnologyAIClasses[0]; got.BaseWeight != 50 || got.CompetitionSensitive {
		t.Fatalf("class0=%+v", got)
	}
	if got := rules.TechnologyAIClasses[10]; got.BaseWeight != 20 || !got.CompetitionSensitive {
		t.Fatalf("class10=%+v", got)
	}
	if got := rules.TechnologyAIFieldGroupValues[22]; got != 12000 {
		t.Fatalf("field-group[22]=%d want=12000", got)
	}
	if got := rules.TechnologyAIClassByID[1]; got != 27 {
		t.Fatalf("Technology 1 AI class=%d want=27", got)
	}
	if got := rules.TechnologyAIClassByID[203]; got != 32 {
		t.Fatalf("Technology 203 AI class=%d want=32", got)
	}
}

func TestRaceResearchModifiersUseSemanticTraits(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}

	sakkra := rules.RaceResearchModifiers["sakkra"]
	if sakkra.PopulationGrowthPercent != 100 || sakkra.FarmingDelta != 1 || !sakkra.Subterranean || sakkra.SpyingBonus != -10 || sakkra.GovernmentTraitID != "government_feudal" {
		t.Fatalf("sakkra research modifiers=%+v", sakkra)
	}
	psilon := rules.RaceResearchModifiers["psilon"]
	if psilon.ScienceDelta != 2 || !psilon.LowGWorld || psilon.GovernmentTraitID != "government_dictatorship" {
		t.Fatalf("psilon research modifiers=%+v", psilon)
	}
	darlok := rules.RaceResearchModifiers["darlok"]
	if darlok.SpyingBonus != 20 || !darlok.StealthyShips {
		t.Fatalf("darlok research modifiers=%+v", darlok)
	}
	elerian := rules.RaceResearchModifiers["elerian"]
	if elerian.ShipAttackBonus != 20 || elerian.ShipDefenseBonus != 25 || !elerian.Telepathic {
		t.Fatalf("elerian research modifiers=%+v", elerian)
	}
}

func TestAdvancedResearchPreferenceProfileValidation(t *testing.T) {
	if err := (AdvancedResearchPreferenceProfile{Personality: 1, Objective: 3, Theme: 6}).Validate(); err != nil {
		t.Fatalf("valid profile rejected: %v", err)
	}
	for _, profile := range []AdvancedResearchPreferenceProfile{
		{Personality: -1},
		{Personality: 6},
		{Objective: 4},
		{Theme: 7},
	} {
		if err := profile.Validate(); err == nil {
			t.Fatalf("invalid profile accepted: %+v", profile)
		}
	}
}
