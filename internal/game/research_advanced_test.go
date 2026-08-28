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

func TestAdvancedProfileWeightsMatchOriginalBranches(t *testing.T) {
	for _, tc := range []struct {
		name string
		got  int
		want int
	}{
		{"objective0-class25", advancedObjectiveWeight(7, 25, 0), 100},
		{"objective1-class26", advancedObjectiveWeight(7, 26, 1), 50},
		{"objective1-class25", advancedObjectiveWeight(7, 25, 1), 20},
		{"objective2-class29", advancedObjectiveWeight(7, 29, 2), 100},
		{"objective3-class36", advancedObjectiveWeight(7, 36, 3), 50},
		{"theme0-class18", advancedThemeWeight(7, 18, 0), 50},
		{"theme1-class15", advancedThemeWeight(7, 15, 1), 20},
		{"theme1-class35", advancedThemeWeight(7, 35, 1), 100},
		{"theme3-class26", advancedThemeWeight(7, 26, 3), 20},
		{"theme6-class34", advancedThemeWeight(7, 34, 6), 100},
		{"personality0-class12", advancedPersonalityWeight(7, 12, 0), 100},
		{"personality1-class17", advancedPersonalityWeight(7, 17, 1), 100},
		{"personality2-class39", advancedPersonalityWeight(7, 39, 2), 50},
		{"personality3-class2", advancedPersonalityWeight(7, 2, 3), 100},
		{"personality4-class4", advancedPersonalityWeight(7, 4, 4), 100},
		{"personality5-class0", advancedPersonalityWeight(7, 0, 5), 100},
		{"unchanged", advancedPersonalityWeight(7, 40, 5), 7},
	} {
		if tc.got != tc.want {
			t.Fatalf("%s=%d want=%d", tc.name, tc.got, tc.want)
		}
	}
}

func TestAdvancedRaceWeightsMatchOriginalBranches(t *testing.T) {
	for _, tc := range []struct {
		name  string
		class int
		race  RaceResearchModifiers
		want  int
	}{
		{"farming-negative", 0, RaceResearchModifiers{FarmingDelta: -0.5}, 100},
		{"farming-positive", 0, RaceResearchModifiers{FarmingDelta: 1}, 10},
		{"lithovore-overrides-farming", 0, RaceResearchModifiers{FarmingDelta: 1, Lithovore: true}, 1},
		{"cybernetic-overrides-farming", 0, RaceResearchModifiers{FarmingDelta: 1, Cybernetic: true}, 20},
		{"industry-negative", 1, RaceResearchModifiers{IndustryDelta: -1}, 100},
		{"science-nonzero", 2, RaceResearchModifiers{ScienceDelta: 2}, 100},
		{"money-positive", 3, RaceResearchModifiers{MoneyDelta: 1}, 20},
		{"tolerant-overrides-industry", 4, RaceResearchModifiers{IndustryDelta: 2, Tolerant: true}, 1},
		{"subterranean", 6, RaceResearchModifiers{Subterranean: true}, 20},
		{"population-growth-positive-overrides-subterranean", 6, RaceResearchModifiers{Subterranean: true, PopulationGrowthPercent: 100}, 5},
		{"population-growth-negative", 6, RaceResearchModifiers{PopulationGrowthPercent: -50}, 100},
		{"spying", 12, RaceResearchModifiers{SpyingBonus: 20}, 50},
		{"democracy", 12, RaceResearchModifiers{GovernmentTraitID: "government_democracy"}, 50},
		{"ground-combat-negative", 16, RaceResearchModifiers{GroundCombatBonus: -10}, 20},
		{"ship-defense-negative", 18, RaceResearchModifiers{ShipDefenseBonus: -20}, 50},
		{"ship-attack-negative", 25, RaceResearchModifiers{ShipAttackBonus: -20}, 100},
		{"ship-attack-positive", 27, RaceResearchModifiers{ShipAttackBonus: 20}, 100},
		{"ship-defense-positive", 28, RaceResearchModifiers{ShipDefenseBonus: 25}, 100},
		{"stealthy", 37, RaceResearchModifiers{StealthyShips: true}, 1},
		{"unification", 40, RaceResearchModifiers{GovernmentTraitID: "government_unification"}, 1},
	} {
		if got := advancedRaceWeight(7, tc.class, tc.race); got != tc.want {
			t.Fatalf("%s=%d want=%d", tc.name, got, tc.want)
		}
	}
}

func TestAdvancedSpecialTechnologyWeights(t *testing.T) {
	if got := advancedSpecialTechnologyWeight(7, 5, RaceResearchModifiers{Telepathic: true}); got != 1 {
		t.Fatalf("Telepathic Alien Management Center weight=%d want=1", got)
	}
	if got := advancedSpecialTechnologyWeight(7, 131, RaceResearchModifiers{LowGWorld: true}); got != 50 {
		t.Fatalf("Low-G Planetary Gravity Generator weight=%d want=50", got)
	}
	if got := advancedSpecialTechnologyWeight(7, 131, RaceResearchModifiers{HighGWorld: true}); got != 1 {
		t.Fatalf("High-G Planetary Gravity Generator weight=%d want=1", got)
	}
}

func TestAdvancedTechnologyEligibilityMatchesGovernmentFamilies(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		government string
		allowed    int
		rejected   []int
	}{
		{"government_feudal", 42, []int{65, 77, 92}},
		{"government_democracy", 65, []int{42, 77, 92}},
		{"government_unification", 77, []int{42, 65, 92}},
		{"government_dictatorship", 92, []int{42, 65, 77}},
	} {
		raceID := "probe_" + tc.government
		rules.RaceModifiers[raceID] = RaceEconomyModifiers{GovernmentTraitID: tc.government}
		empire := &core.Empire{ID: 1, RaceID: raceID}
		if !rules.advancedTechnologyEligible(empire, tc.allowed, false) {
			t.Fatalf("%s should allow Technology %d", tc.government, tc.allowed)
		}
		for _, technologyID := range tc.rejected {
			if rules.advancedTechnologyEligible(empire, technologyID, false) {
				t.Fatalf("%s unexpectedly allows Technology %d", tc.government, technologyID)
			}
		}
	}
}

func TestCompetitionAIGroupUsesOriginalFirstSeventyFiveTechnologyLimit(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	classID := -1
	lowID := 0
	highID := 0
	lowGroup := -1
	highGroup := -1
	for candidateClass := 0; candidateClass < 41 && highID == 0; candidateClass++ {
		for technologyID := 1; technologyID <= 203; technologyID++ {
			if rules.TechnologyAIClassByID[technologyID] != candidateClass {
				continue
			}
			group := rules.TechnologyFieldAIGroup[rules.TechnologyFieldByID[technologyID]]
			if technologyID <= 75 && group > lowGroup {
				lowID = technologyID
				lowGroup = group
			}
			if technologyID > 75 && group > highGroup {
				highID = technologyID
				highGroup = group
			}
		}
		if lowID != 0 && highID != 0 && highGroup > lowGroup {
			classID = candidateClass
			break
		}
		lowID, highID, lowGroup, highGroup = 0, 0, -1, -1
	}
	if classID < 0 {
		t.Fatal("could not find AI class with a higher >75 Technology group for limit regression")
	}
	other := core.Empire{ID: 2, KnownTechnologyIDs: []int{lowID, highID}}
	if got := rules.maxKnownAIGroup(&other, classID, 75); got != lowGroup {
		t.Fatalf("first-75 max AI group=%d want=%d (low=%d high=%d)", got, lowGroup, lowID, highID)
	}
	if got := rules.maxKnownAIGroup(&other, classID, 0); got != highGroup {
		t.Fatalf("unlimited max AI group=%d want=%d", got, highGroup)
	}
	state := &core.GameState{Empires: []core.Empire{{ID: 1}, other}}
	if got := rules.competitionAIGroup(state, 1, classID); got != lowGroup {
		t.Fatalf("competition AI group=%d want original first-75 value %d", got, lowGroup)
	}
}

func TestAdvancedProgressionWeightMatchesOriginalIntegerArithmetic(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		weight               int
		classID              int
		targetAIGroup        int
		targetProgression    int
		knownAIGroup         int
		competitionSensitive bool
		want                 int
	}{
		{"non-sensitive", 10, 7, 5, 22, 0, false, 220},
		{"sensitive-far-ahead", 10, 7, 5, 22, 1, true, 293},
		{"sensitive-near-zero-class", 10, 18, 5, 22, 4, true, 0},
		{"sensitive-near-ordinary", 10, 7, 5, 22, 4, true, 220},
	} {
		got := advancedProgressionWeight(tc.weight, tc.classID, tc.targetAIGroup, tc.targetProgression, tc.knownAIGroup, tc.competitionSensitive)
		if got != tc.want {
			t.Fatalf("%s=%d want=%d", tc.name, got, tc.want)
		}
	}
}

func TestAdvancedCompetitionWeightUsesOriginalClassPairing(t *testing.T) {
	rules := &EconomyRules{
		TechnologyAIClassByID:  map[int]int{1: 18, 2: 33},
		TechnologyFieldByID:    map[int]int{1: 1, 2: 2},
		TechnologyFieldAIGroup: map[int]int{1: 10, 2: 8},
	}
	current := core.Empire{ID: 1}
	other := core.Empire{ID: 2, KnownTechnologyIDs: []int{1, 2}}
	state := &core.GameState{Empires: []core.Empire{current, other}}
	if got := rules.advancedCompetitionWeight(state, &state.Empires[0], 25, 3, 2, AdvancedResearchPreferenceProfile{}); got != 14 {
		t.Fatalf("class25 vs competition class18 weight=%d want=14", got)
	}
	if got := rules.advancedCompetitionWeight(state, &state.Empires[0], 21, 3, 2, AdvancedResearchPreferenceProfile{Objective: 2}); got != 10 {
		t.Fatalf("objective2 class21 vs competition class33 weight=%d want=10", got)
	}
	if got := advancedGapMultiplier(7, 12, 8); got != 7 {
		t.Fatalf("gap multiplier without gap=%d want=7", got)
	}
}

func TestAdvancedCalcTechnologyWeightSpecialCases(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	state := &core.GameState{Empires: []core.Empire{{ID: 1, RaceID: "human"}}}
	profile := AdvancedResearchPreferenceProfile{Personality: 1, Objective: 0, Theme: 0}
	if got, err := rules.advancedCalcTechnologyWeight(state, 1, 52, profile, false); err != nil || got != 0 {
		t.Fatalf("Dimensional Portal Calc_Tech_Value weight=%d err=%v want 0,nil", got, err)
	}

	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyAverage}); err != nil {
		t.Fatal(err)
	}
	candidates, err := rules.advancedCandidateTechnologyIDs(state, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	class18Tech := 0
	for _, technologyID := range candidates {
		if rules.TechnologyAIClassByID[technologyID] == 18 {
			class18Tech = technologyID
			break
		}
	}
	if class18Tech == 0 {
		t.Fatal("no open Advanced candidate with AI class 18")
	}
	empire := &state.Empires[0]
	base, err := rules.advancedProfileRaceBaseWeight(empire, class18Tech, profile)
	if err != nil {
		t.Fatal(err)
	}
	fieldID := rules.TechnologyFieldByID[class18Tech]
	targetAIGroup := rules.TechnologyFieldAIGroup[fieldID]
	if targetAIGroup > 22 {
		targetAIGroup = 22
	}
	progression := rules.TechnologyAIFieldGroupValues[targetAIGroup]
	known := rules.maxKnownAIGroup(empire, 18, 0)
	class := rules.TechnologyAIClasses[18]
	preEarly := advancedProgressionWeight(base, 18, targetAIGroup, progression, known, class.CompetitionSensitive)
	preEarly = rules.advancedCompetitionWeight(state, empire, 18, progression, preEarly, profile)
	want := preEarly * 2
	if want == 0 && rules.allOtherFieldTechnologiesKnown(empire, fieldID, class18Tech) {
		want = progression * 10
	}
	got, err := rules.advancedCalcTechnologyWeight(state, 1, class18Tech, profile, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("class18 Advanced weight=%d want=%d (pre-early=%d)", got, want, preEarly)
	}

	empire.KnownTechnologyIDs = appendUniqueSortedInt(empire.KnownTechnologyIDs, class18Tech)
	if got, err := rules.advancedCalcTechnologyWeight(state, 1, class18Tech, profile, false); err != nil || got != 0 {
		t.Fatalf("known Technology weight=%d err=%v want 0,nil", got, err)
	}
}
