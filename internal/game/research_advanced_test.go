package game

import (
	"fmt"
	"path/filepath"
	"testing"

	"moox/internal/core"
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

func TestAdvancedCandidateAndGrantSemantics(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}

	creativeState := &core.GameState{Empires: []core.Empire{{ID: 1, RaceID: "psilon"}}}
	if err := rules.InitializeEmpireTechnologies(&creativeState.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyAverage}); err != nil {
		t.Fatal(err)
	}
	creativeCandidates, err := rules.advancedCandidateTechnologyIDs(creativeState, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	creativeTech := 0
	creativeField := 0
	for _, technologyID := range creativeCandidates {
		fieldID := rules.TechnologyFieldByID[technologyID]
		if len(rules.TechnologyIDsByField[fieldID]) > 1 {
			creativeTech = technologyID
			creativeField = fieldID
			break
		}
	}
	if creativeTech == 0 {
		t.Fatal("no multi-application Creative Advanced candidate found")
	}
	if err := rules.grantAdvancedStartingTechnology(creativeState, 1, creativeTech, false); err != nil {
		t.Fatal(err)
	}
	for _, technologyID := range rules.TechnologyIDsByField[creativeField] {
		if !containsInt(creativeState.Empires[0].KnownTechnologyIDs, technologyID) {
			t.Fatalf("Creative grant TechField %d missing Technology %d", creativeField, technologyID)
		}
	}

	uncreativeState := &core.GameState{Empires: []core.Empire{{ID: 2, RaceID: "klackon"}}}
	rng := core.NewRNG(0x1234)
	if err := rules.InitializeEmpireTechnologies(&uncreativeState.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyAverage, NewGameRNG: rng}); err != nil {
		t.Fatal(err)
	}
	uncreativeCandidates, err := rules.advancedCandidateTechnologyIDs(uncreativeState, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	seenFields := map[int]struct{}{}
	fixedByField := map[int]int{}
	for _, fixed := range uncreativeState.Empires[0].UncreativeResearchChoices {
		fixedByField[fixed.TechFieldID] = fixed.TechnologyID
	}
	for _, technologyID := range uncreativeCandidates {
		fieldID := rules.TechnologyFieldByID[technologyID]
		if _, duplicate := seenFields[fieldID]; duplicate {
			t.Fatalf("Uncreative frontier exposes multiple applications for TechField %d", fieldID)
		}
		seenFields[fieldID] = struct{}{}
		if want := fixedByField[fieldID]; technologyID != want {
			t.Fatalf("Uncreative TechField %d candidate=%d want fixed=%d", fieldID, technologyID, want)
		}
	}
}

func TestAdvancedNewGameOrchestrationIsOrderedAndNineteenGrants(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	state := &core.GameState{Empires: []core.Empire{
		{ID: 2, RaceID: "human"},
		{ID: 1, RaceID: "psilon"},
	}}
	profiles := map[core.ID]AdvancedResearchPreferenceProfile{
		1: {Personality: 1, Objective: 0, Theme: 0},
		2: {Personality: 1, Objective: 0, Theme: 0},
	}
	calls := map[core.ID]int{}
	secondEmpireObservedFirstComplete := false
	chooser := func(current *core.GameState, empireID core.ID, _ AdvancedResearchPreferenceProfile, strategicCombat bool, rng *core.RNG) (int, error) {
		calls[empireID]++
		if empireID == 2 && calls[empireID] == 1 {
			first := empireByID(current, 1)
			secondEmpireObservedFirstComplete = first != nil && len(first.KnownTechnologyFieldIDs) == 7+advancedNewGameExtraGrantCount
		}
		candidates, err := rules.advancedCandidateTechnologyIDs(current, empireID, strategicCombat)
		if err != nil {
			return 0, err
		}
		if len(candidates) == 0 {
			return 0, fmt.Errorf("no Advanced candidates")
		}
		index, err := rng.Intn(len(candidates))
		if err != nil {
			return 0, err
		}
		return candidates[index], nil
	}
	if err := rules.initializeAdvancedNewGameTechnologies(state, NewGameTechnologyStateOptions{
		Level:               NewGameTechnologyAdvanced,
		NewGameRNG:          core.NewRNG(0xBEEF),
		AdvancedPreferences: profiles,
	}, chooser); err != nil {
		t.Fatal(err)
	}
	if calls[1] != advancedNewGameExtraGrantCount || calls[2] != advancedNewGameExtraGrantCount {
		t.Fatalf("Advanced chooser calls=%v want %d each", calls, advancedNewGameExtraGrantCount)
	}
	if !secondEmpireObservedFirstComplete {
		t.Fatal("second Empire did not observe first Empire's completed Advanced initialization")
	}
	for _, empire := range state.Empires {
		if got := len(empire.KnownTechnologyFieldIDs); got != 7+advancedNewGameExtraGrantCount {
			t.Fatalf("Empire %d known fields=%d want=%d", empire.ID, got, 7+advancedNewGameExtraGrantCount)
		}
	}
}

func TestInitializeNewGameTechnologiesAverageUsesStableEmpireOrder(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	state := &core.GameState{Empires: []core.Empire{
		{ID: 9, RaceID: "human"},
		{ID: 3, RaceID: "psilon"},
	}}
	if err := rules.InitializeNewGameTechnologies(state, NewGameTechnologyStateOptions{Level: NewGameTechnologyAverage}); err != nil {
		t.Fatal(err)
	}
	for _, empire := range state.Empires {
		if got := len(empire.KnownTechnologyFieldIDs); got != 7 {
			t.Fatalf("Empire %d Average known fields=%d want=7", empire.ID, got)
		}
	}
}
