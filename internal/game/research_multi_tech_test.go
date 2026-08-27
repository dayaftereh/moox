package game

import (
	"reflect"
	"testing"

	"moox/internal/core"
)

func initializedResearchRace(t *testing.T, seed uint64, raceID string, uncreativeSeed uint64) (*EconomyRules, *EconomyResolver, *core.GameState) {
	t.Helper()
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(seed)
	state.Empires[0].RaceID = raceID
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], NewGameTechnologyOptions{
		Level:                   NewGameTechnologyPreWarp,
		UncreativeSelectionSeed: uncreativeSeed,
	}); err != nil {
		t.Fatal(err)
	}
	return rules, resolver, state
}

func TestOrdinaryResearchChoosesOneApplicationAtSelectionTime(t *testing.T) {
	rules, resolver, state := initializedResearchRace(t, 740, "human", 0)
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
	rules, resolver, state := initializedResearchRace(t, 741, "human", 0)
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
	rules, resolver, state := initializedResearchRace(t, 742, "psilon", 0)
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
	rules, resolver, state := initializedResearchRace(t, 743, "klackon", 0x12345678)
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

func TestUncreativeResearchPlanIsDeterministicAndRequiresSeed(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	missing := core.NewSmallFixture(744)
	missing.Empires[0].RaceID = "klackon"
	if err := rules.InitializeEmpireTechnologies(&missing.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err == nil {
		t.Fatal("Uncreative initialization accepted missing selection seed")
	}

	first := core.NewSmallFixture(745)
	second := core.NewSmallFixture(745)
	first.Empires[0].RaceID = "klackon"
	second.Empires[0].RaceID = "klackon"
	options := NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp, UncreativeSelectionSeed: 0xC0FFEE}
	if err := rules.InitializeEmpireTechnologies(&first.Empires[0], options); err != nil {
		t.Fatal(err)
	}
	if err := rules.InitializeEmpireTechnologies(&second.Empires[0], options); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Empires[0].UncreativeResearchChoices, second.Empires[0].UncreativeResearchChoices) {
		t.Fatalf("same seed produced different Uncreative plan:\nfirst=%v\nsecond=%v", first.Empires[0].UncreativeResearchChoices, second.Empires[0].UncreativeResearchChoices)
	}
	if len(first.Empires[0].UncreativeResearchChoices) == 0 {
		t.Fatal("Uncreative initialization produced no fixed research choices")
	}
}

func TestResearchCompletionRejectsRaceSelectionModeMismatch(t *testing.T) {
	rules, resolver, state := initializedResearchRace(t, 746, "human", 0)
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
	rules, resolver, state := initializedResearchRace(t, 747, "klackon", 0xA11CE)
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
