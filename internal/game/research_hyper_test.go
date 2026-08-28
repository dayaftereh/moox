package game

import (
	"encoding/json"
	"testing"

	"moox/internal/core"
)

func TestHyperAdvancedResearchChoiceAndSelection(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(760)
	state.Empires[0].RaceID = "human"
	state.Empires[0].KnownTechnologyFieldIDs = []int{70}

	choices, err := rules.AvailableResearchChoices(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	field75, ok := researchChoiceByField(choices, 75)
	if !ok {
		t.Fatal("Hyper-Advanced TechField 75 was not offered after predecessor 70")
	}
	if field75.SelectionMode != core.ResearchSelectionRepeatField || field75.BaseCostRP != 15000 || field75.CompletedLevels != 0 || field75.ResearchLevel != 1 || len(field75.TechnologyIDs) != 0 {
		t.Fatalf("initial Hyper-Advanced choice=%+v", field75)
	}

	illegal, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 75, TechnologyID: 155})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, illegal); err == nil {
		t.Fatal("Hyper-Advanced selection accepted a concrete technology_id")
	}

	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 75})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if state.Empires[0].Research == nil || state.Empires[0].Research.TechFieldID != 75 || state.Empires[0].Research.SelectionMode != core.ResearchSelectionRepeatField || len(state.Empires[0].Research.TechnologyIDs) != 0 {
		t.Fatalf("active Hyper-Advanced research=%+v", state.Empires[0].Research)
	}
	var selected ResearchSelectedEvent
	if err := json.Unmarshal(event.Data, &selected); err != nil {
		t.Fatal(err)
	}
	if selected.BaseCostRP != 15000 || selected.CompletedLevels != 0 || selected.ResearchLevel != 1 || len(selected.TechnologyIDs) != 0 {
		t.Fatalf("Hyper-Advanced selected event=%+v", selected)
	}
}

func TestHyperAdvancedChoiceUsesPersistedCompletedLevel(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(761)
	state.Empires[0].RaceID = "psilon"
	state.Empires[0].KnownTechnologyFieldIDs = []int{70}
	state.Empires[0].HyperAdvancedResearch = []core.HyperAdvancedResearchLevel{{TechFieldID: 75, CompletedLevels: 3}}

	choices, err := rules.AvailableResearchChoices(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	field75, ok := researchChoiceByField(choices, 75)
	if !ok {
		t.Fatal("repeatable Hyper-Advanced TechField 75 disappeared after completed levels")
	}
	if field75.SelectionMode != core.ResearchSelectionRepeatField || field75.BaseCostRP != 45000 || field75.CompletedLevels != 3 || field75.ResearchLevel != 4 {
		t.Fatalf("level-4 Hyper-Advanced choice=%+v", field75)
	}
}

func TestHyperAdvancedReselectTreatsNilAndEmptyTechnologyListsAsSameProject(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(763)
	state.Empires[0].RaceID = "human"
	state.Empires[0].KnownTechnologyFieldIDs = []int{70}
	state.Empires[0].Research = &core.ResearchState{
		TechFieldID:   75,
		SelectionMode: core.ResearchSelectionRepeatField,
		TechnologyIDs: []int{},
		ProgressRP:    12.5,
	}
	command, err := NewSelectResearchCommand(3, SelectResearchPayload{TechFieldID: 75})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command); err == nil {
		t.Fatal("reselecting the same Hyper-Advanced project was accepted because [] and nil technology lists differed")
	}
	if state.Empires[0].Research.ProgressRP != 12.5 {
		t.Fatalf("rejected Hyper reselect changed progress: %+v", state.Empires[0].Research)
	}
}
func TestHyperAdvancedSwitchPreservesAccumulatedRP(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(762)
	state.Empires[0].RaceID = "human"
	state.Empires[0].KnownTechnologyFieldIDs = []int{70}
	state.Empires[0].Research = &core.ResearchState{
		TechFieldID: 4, SelectionMode: core.ResearchSelectionChooseOne,
		TechnologyIDs: []int{56}, ProgressRP: 123.5,
	}
	command, err := NewSelectResearchCommand(2, SelectResearchPayload{TechFieldID: 75})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if state.Empires[0].Research == nil || state.Empires[0].Research.TechFieldID != 75 || state.Empires[0].Research.ProgressRP != 123.5 {
		t.Fatalf("switched Hyper-Advanced research=%+v", state.Empires[0].Research)
	}
	var switched ResearchSwitchedEvent
	if err := json.Unmarshal(event.Data, &switched); err != nil {
		t.Fatal(err)
	}
	if switched.TransferredRP != 123.5 || switched.Current.SelectionMode != core.ResearchSelectionRepeatField || switched.Current.ResearchLevel != 1 {
		t.Fatalf("Hyper-Advanced switch event=%+v", switched)
	}
}
