package game

import (
	"reflect"
	"testing"

	"moox/internal/core"
)

func initializedResearchRace(t *testing.T, seed uint64, raceID string) (*EconomyRules, *EconomyResolver, *core.GameState) {
	t.Helper()
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(seed)
	state.Empires[0].RaceID = raceID
	options := NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}
	var newGameRNG *core.RNG
	if modifiers := rules.RaceModifiers[raceID]; modifiers.Uncreative {
		newGameRNG = state.RNG()
		options.NewGameRNG = newGameRNG
	}
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], options); err != nil {
		t.Fatal(err)
	}
	if newGameRNG != nil {
		state.CommitRNG(newGameRNG)
	}
	return rules, resolver, state
}
func TestOrdinaryResearchChoosesOneApplicationAtSelectionTime(t *testing.T) {
	rules, resolver, state := initializedResearchRace(t, 740, "human")
	choices, err := rules.AvailableResearchChoices(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	field4, ok := researchChoiceByField(choices, 4)
	if !ok {
		t.Fatal("missing TechField 4")
	}
	if field4.SelectionMode != core.ResearchSelectionChooseOne {
		t.Fatalf("Human field 4 mode=%q want choose_one", field4.SelectionMode)
	}
	if want := []int{13, 56, 66}; !reflect.DeepEqual(field4.TechnologyIDs, want) {
		t.Fatalf("Human field 4 applications=%v want=%v", field4.TechnologyIDs, want)
	}

	missing, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, missing); err == nil {
		t.Fatal("ordinary research accepted a field without choosing an application")
	}
	illegal, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4, TechnologyID: 155})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, illegal); err == nil {
		t.Fatal("ordinary research accepted a technology outside the field")
	}

	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != "empire.research_selected" {
		t.Fatalf("selection event kind=%q", event.Kind)
	}
	if state.Empires[0].Research.SelectionMode != core.ResearchSelectionChooseOne || !reflect.DeepEqual(state.Empires[0].Research.TechnologyIDs, []int{56}) {
		t.Fatalf("ordinary active research=%+v", state.Empires[0].Research)
	}
	state.Empires[0].Research.ProgressRP = rules.TechnologyFieldCostsRP[4]
	if _, err := resolver.CompleteResearchField(state, state.Empires[0].ID); err != nil {
		t.Fatal(err)
	}
	if !containsTechnology(state.Empires[0].KnownTechnologyIDs, 56) || containsTechnology(state.Empires[0].KnownTechnologyIDs, 13) || containsTechnology(state.Empires[0].KnownTechnologyIDs, 66) {
		t.Fatalf("ordinary completion granted wrong application set: %v", state.Empires[0].KnownTechnologyIDs)
	}
}

func TestGeneralResearchFieldGrantsAllApplicationsForOrdinaryRace(t *testing.T) {
	rules, resolver, state := initializedResearchRace(t, 741, "human")
	choices, err := rules.AvailableResearchChoices(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	field55, ok := researchChoiceByField(choices, 55)
	if !ok {
		t.Fatal("missing General TechField 55")
	}
	if field55.SelectionMode != core.ResearchSelectionAll {
		t.Fatalf("General field 55 mode=%q want all", field55.SelectionMode)
	}
	if !reflect.DeepEqual(field55.TechnologyIDs, rules.TechnologyIDsByField[55]) {
		t.Fatalf("General field 55 applications=%v want=%v", field55.TechnologyIDs, rules.TechnologyIDsByField[55])
	}
	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 55})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command); err != nil {
		t.Fatal(err)
	}
	if state.Empires[0].Research.SelectionMode != core.ResearchSelectionAll || !reflect.DeepEqual(state.Empires[0].Research.TechnologyIDs, rules.TechnologyIDsByField[55]) {
		t.Fatalf("General field active research=%+v", state.Empires[0].Research)
	}
}

func TestCreativeResearchAcquiresAllFieldApplications(t *testing.T) {
	rules, resolver, state := initializedResearchRace(t, 742, "psilon")
	choices, err := rules.AvailableResearchChoices(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	field4, ok := researchChoiceByField(choices, 4)
	if !ok {
		t.Fatal("missing Psilon TechField 4")
	}
	if field4.SelectionMode != core.ResearchSelectionAll || !reflect.DeepEqual(field4.TechnologyIDs, []int{13, 56, 66}) {
		t.Fatalf("Creative field 4=%+v", field4)
	}
	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command); err != nil {
		t.Fatal(err)
	}
	state.Empires[0].Research.ProgressRP = rules.TechnologyFieldCostsRP[4]
	if _, err := resolver.CompleteResearchField(state, state.Empires[0].ID); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{13, 56, 66} {
		if !containsTechnology(state.Empires[0].KnownTechnologyIDs, id) {
			t.Fatalf("Creative completion missing technology %d: %v", id, state.Empires[0].KnownTechnologyIDs)
		}
	}
}

func TestUncreativeResearchUsesPersistedFixedApplicationWithoutQueryRNG(t *testing.T) {
	rules, resolver, state := initializedResearchRace(t, 743, "klackon")
	fixedID, ok := fixedResearchTechnology(state.Empires[0].UncreativeResearchChoices, 4)
	if !ok {
		t.Fatal("Klackon has no persisted fixed application for TechField 4")
	}
	before, err := core.MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	choices, err := rules.AvailableResearchChoices(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	after, err := core.MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("ResearchChoices mutated authoritative state or RNG")
	}
	field4, ok := researchChoiceByField(choices, 4)
	if !ok {
		t.Fatal("missing Klackon TechField 4")
	}
	if field4.SelectionMode != core.ResearchSelectionFixedOne || !reflect.DeepEqual(field4.TechnologyIDs, []int{fixedID}) {
		t.Fatalf("Uncreative field 4=%+v fixed=%d", field4, fixedID)
	}
	field55, ok := researchChoiceByField(choices, 55)
	if !ok || field55.SelectionMode != core.ResearchSelectionAll {
		t.Fatalf("Uncreative General field 55=%+v", field55)
	}

	clientChoice, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4, TechnologyID: fixedID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, clientChoice); err == nil {
		t.Fatal("Uncreative client was allowed to submit its own technology choice")
	}
	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command); err != nil {
		t.Fatal(err)
	}
	if state.Empires[0].Research.SelectionMode != core.ResearchSelectionFixedOne || !reflect.DeepEqual(state.Empires[0].Research.TechnologyIDs, []int{fixedID}) {
		t.Fatalf("Uncreative active research=%+v", state.Empires[0].Research)
	}
}

func TestUncreativeResearchPlanUsesSharedNewGameRNGDeterministically(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	missing := core.NewSmallFixture(744)
	missing.Empires[0].RaceID = "klackon"
	if err := rules.InitializeEmpireTechnologies(&missing.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err == nil {
		t.Fatal("Uncreative initialization accepted a missing NewGameRNG")
	}
	if len(missing.Empires[0].KnownTechnologyFieldIDs) != 0 || len(missing.Empires[0].KnownTechnologyIDs) != 0 || len(missing.Empires[0].UncreativeResearchChoices) != 0 {
		t.Fatalf("failed NewGameRNG preflight partially initialized empire: %+v", missing.Empires[0])
	}

	first := core.NewSmallFixture(745)
	second := core.NewSmallFixture(745)
	first.Empires[0].RaceID = "klackon"
	second.Empires[0].RaceID = "klackon"
	firstRNG := first.RNG()
	secondRNG := second.RNG()
	firstInitialRNGState := firstRNG.State()
	if err := rules.InitializeEmpireTechnologies(&first.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp, NewGameRNG: firstRNG}); err != nil {
		t.Fatal(err)
	}
	if err := rules.InitializeEmpireTechnologies(&second.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp, NewGameRNG: secondRNG}); err != nil {
		t.Fatal(err)
	}
	first.CommitRNG(firstRNG)
	second.CommitRNG(secondRNG)
	if first.RNGState == firstInitialRNGState {
		t.Fatal("Uncreative initialization did not consume the shared new-game RNG")
	}
	if first.RNGState != second.RNGState {
		t.Fatalf("same new-game RNG state diverged: first=%d second=%d", first.RNGState, second.RNGState)
	}
	if !reflect.DeepEqual(first.Empires[0].UncreativeResearchChoices, second.Empires[0].UncreativeResearchChoices) {
		t.Fatalf("same new-game RNG produced different Uncreative plans:\nfirst=%v\nsecond=%v", first.Empires[0].UncreativeResearchChoices, second.Empires[0].UncreativeResearchChoices)
	}
	if len(first.Empires[0].UncreativeResearchChoices) == 0 {
		t.Fatal("Uncreative initialization produced no fixed research choices")
	}
	for _, fixed := range first.Empires[0].UncreativeResearchChoices {
		if fixed.TechFieldID < 1 || fixed.TechFieldID > 73 {
			t.Fatalf("Uncreative plan contains non-original initial field %+v", fixed)
		}
		if _, general := rules.GeneralResearchFieldIDs[fixed.TechFieldID]; general {
			t.Fatalf("Uncreative plan incorrectly fixes General field %+v", fixed)
		}
	}
	if _, found := fixedResearchTechnology(first.Empires[0].UncreativeResearchChoices, 74); found {
		t.Fatal("Uncreative plan incorrectly contains special Antaran TechField 74")
	}
}

func TestUncreativePlansConsumeOneSharedNewGameRNGAcrossPlayers(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	buildPair := func(seed uint64) (core.Empire, core.Empire, uint64) {
		first := core.Empire{ID: 1, RaceID: "klackon"}
		second := core.Empire{ID: 2, RaceID: "klackon"}
		rng := core.NewRNG(seed)
		options := NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp, NewGameRNG: rng}
		if err := rules.InitializeEmpireTechnologies(&first, options); err != nil {
			t.Fatal(err)
		}
		if err := rules.InitializeEmpireTechnologies(&second, options); err != nil {
			t.Fatal(err)
		}
		return first, second, rng.State()
	}

	firstA, secondA, stateA := buildPair(0x1BADB002)
	firstB, secondB, stateB := buildPair(0x1BADB002)
	if stateA != stateB || !reflect.DeepEqual(firstA.UncreativeResearchChoices, firstB.UncreativeResearchChoices) || !reflect.DeepEqual(secondA.UncreativeResearchChoices, secondB.UncreativeResearchChoices) {
		t.Fatalf("shared new-game RNG sequence is not deterministic:\nA=%v / %v state=%d\nB=%v / %v state=%d", firstA.UncreativeResearchChoices, secondA.UncreativeResearchChoices, stateA, firstB.UncreativeResearchChoices, secondB.UncreativeResearchChoices, stateB)
	}

	human := core.Empire{ID: 3, RaceID: "human"}
	humanRNG := core.NewRNG(0xDEADBEEF)
	before := humanRNG.State()
	if err := rules.InitializeEmpireTechnologies(&human, NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp, NewGameRNG: humanRNG}); err != nil {
		t.Fatal(err)
	}
	if humanRNG.State() != before {
		t.Fatalf("ordinary player consumed Uncreative new-game RNG: before=%d after=%d", before, humanRNG.State())
	}
}
func TestUncreativeInitialTechnologyEligibilityMatchesOriginalTraitFilters(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	modifiers := RaceEconomyModifiers{
		GovernmentTraitID: "government_unification",
		Tolerant:          true,
		Lithovore:         true,
	}
	options := NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}
	for _, technologyID := range []int{86, 141, 195, 19, 50, 113, 142, 6, 29, 68, 87, 162, 178} {
		if rules.uncreativeInitialTechnologyEligible(modifiers, options, technologyID) {
			t.Fatalf("original trait filter incorrectly accepted technology %d", technologyID)
		}
	}
	if !rules.uncreativeInitialTechnologyEligible(modifiers, options, 13) {
		t.Fatal("trait filter rejected unrelated Anti-Missile Rockets")
	}
	if rules.uncreativeInitialTechnologyEligible(RaceEconomyModifiers{}, NewGameTechnologyOptions{StrategicCombat: true}, 56) {
		t.Fatal("strategic-combat filter accepted non-strategic Reinforced Hull")
	}
	if !rules.uncreativeInitialTechnologyEligible(RaceEconomyModifiers{}, NewGameTechnologyOptions{StrategicCombat: false}, 56) {
		t.Fatal("tactical mode rejected Reinforced Hull")
	}
	if rules.uncreativeInitialTechnologyEligible(RaceEconomyModifiers{}, NewGameTechnologyOptions{RandomEventsDisabled: true}, 52) {
		t.Fatal("No Random Events accepted Dimensional Portal")
	}
	if !rules.uncreativeInitialTechnologyEligible(RaceEconomyModifiers{}, NewGameTechnologyOptions{}, 52) {
		t.Fatal("default Random Events setting rejected Dimensional Portal")
	}
}
func TestResearchCompletionRejectsRaceSelectionModeMismatch(t *testing.T) {
	rules, resolver, state := initializedResearchRace(t, 746, "human")
	state.Empires[0].Research = &core.ResearchState{
		TechFieldID:   4,
		SelectionMode: core.ResearchSelectionAll,
		TechnologyIDs: append([]int(nil), rules.TechnologyIDsByField[4]...),
		ProgressRP:    rules.TechnologyFieldCostsRP[4],
	}
	if _, err := resolver.CompleteResearchField(state, state.Empires[0].ID); err == nil {
		t.Fatal("Human research completion accepted Creative all-applications mode")
	}
}

func TestResearchCompletionRejectsUncreativeFixedChoiceMismatch(t *testing.T) {
	rules, resolver, state := initializedResearchRace(t, 747, "klackon")
	fixedID, ok := fixedResearchTechnology(state.Empires[0].UncreativeResearchChoices, 4)
	if !ok {
		t.Fatal("missing Uncreative fixed TechField 4 application")
	}
	wrongID := 13
	if wrongID == fixedID {
		wrongID = 56
	}
	state.Empires[0].Research = &core.ResearchState{
		TechFieldID:   4,
		SelectionMode: core.ResearchSelectionFixedOne,
		TechnologyIDs: []int{wrongID},
		ProgressRP:    rules.TechnologyFieldCostsRP[4],
	}
	if _, err := resolver.CompleteResearchField(state, state.Empires[0].ID); err == nil {
		t.Fatalf("Uncreative research completion accepted %d instead of fixed %d", wrongID, fixedID)
	}
}
