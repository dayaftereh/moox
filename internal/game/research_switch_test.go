package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
)

func TestResearchChoicesRemainAvailableDuringActiveProject(t *testing.T) {
	rules, resolver, state := initializedResearchRace(t, 750, "human")
	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command); err != nil {
		t.Fatal(err)
	}
	state.Empires[0].Research.ProgressRP = 20

	choices, err := rules.AvailableResearchChoices(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	field4, ok := researchChoiceByField(choices, 4)
	if !ok || field4.SelectionMode != core.ResearchSelectionChooseOne {
		t.Fatalf("active project should keep TechField 4 switch choices available: %+v", field4)
	}
	if _, ok := researchChoiceByField(choices, 55); !ok {
		t.Fatalf("active project should allow switching to another legal field: %+v", choices)
	}
}

func TestResearchSwitchPreservesProgressAcrossFields(t *testing.T) {
	rules, resolver, state := initializedResearchRace(t, 751, "human")
	selectField4, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, selectField4); err != nil {
		t.Fatal(err)
	}
	state.Empires[0].Research.ProgressRP = 70

	switchTo55, err := NewSelectResearchCommand(2, SelectResearchPayload{TechFieldID: 55})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.selectResearch(state, state.Empires[0].ID, 1, switchTo55)
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != "empire.research_switched" || event.CommandSequence != 2 {
		t.Fatalf("switch event=%+v", event)
	}
	active := state.Empires[0].Research
	if active == nil || active.TechFieldID != 55 || active.SelectionMode != core.ResearchSelectionAll || active.ProgressRP != 70 {
		t.Fatalf("switched research=%+v", active)
	}
	if !reflect.DeepEqual(active.TechnologyIDs, rules.TechnologyIDsByField[55]) {
		t.Fatalf("switched General field technologies=%v want=%v", active.TechnologyIDs, rules.TechnologyIDsByField[55])
	}

	var payload ResearchSwitchedEvent
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.EmpireID != state.Empires[0].ID || payload.TransferredRP != 70 {
		t.Fatalf("switch payload=%+v", payload)
	}
	if payload.Previous.TechFieldID != 4 || payload.Previous.SelectionMode != core.ResearchSelectionChooseOne || !reflect.DeepEqual(payload.Previous.TechnologyIDs, []int{56}) {
		t.Fatalf("previous project snapshot=%+v", payload.Previous)
	}
	if payload.Current.TechFieldID != 55 || payload.Current.SelectionMode != core.ResearchSelectionAll || !reflect.DeepEqual(payload.Current.TechnologyIDs, rules.TechnologyIDsByField[55]) {
		t.Fatalf("current project snapshot=%+v", payload.Current)
	}
}

func TestResearchSwitchPreservesProgressWithinFieldApplication(t *testing.T) {
	_, resolver, state := initializedResearchRace(t, 752, "human")
	selectReinforcedHull, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, selectReinforcedHull); err != nil {
		t.Fatal(err)
	}
	state.Empires[0].Research.ProgressRP = 40

	switchApplication, err := NewSelectResearchCommand(2, SelectResearchPayload{TechFieldID: 4, TechnologyID: 13})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.selectResearch(state, state.Empires[0].ID, 1, switchApplication)
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != "empire.research_switched" {
		t.Fatalf("event kind=%q", event.Kind)
	}
	active := state.Empires[0].Research
	if active.TechFieldID != 4 || active.SelectionMode != core.ResearchSelectionChooseOne || active.ProgressRP != 40 || !reflect.DeepEqual(active.TechnologyIDs, []int{13}) {
		t.Fatalf("application switch state=%+v", active)
	}
}

func TestResearchSwitchRejectsSameSelectionWithoutMutation(t *testing.T) {
	_, resolver, state := initializedResearchRace(t, 753, "human")
	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command); err != nil {
		t.Fatal(err)
	}
	state.Empires[0].Research.ProgressRP = 30
	before, err := core.MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	same, err := NewSelectResearchCommand(2, SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, same); err == nil {
		t.Fatal("same research selection should be rejected as a no-op")
	}
	after, err := core.MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("rejected same-selection switch mutated authoritative state")
	}
}

func TestResearchSwitchDoesNotClampProgressToCheaperTarget(t *testing.T) {
	_, resolver, state := initializedResearchRace(t, 754, "human")
	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command); err != nil {
		t.Fatal(err)
	}
	state.Empires[0].Research.ProgressRP = 79
	switchTo55, err := NewSelectResearchCommand(2, SelectResearchPayload{TechFieldID: 55})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, switchTo55); err != nil {
		t.Fatal(err)
	}
	if state.Empires[0].Research.ProgressRP != 79 {
		t.Fatalf("switch clamped/transformed RP=%v want=79", state.Empires[0].Research.ProgressRP)
	}
}

func TestResearchSwitchPreservesFractionalProgressRP(t *testing.T) {
	_, resolver, state := initializedResearchRace(t, 755, "human")
	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command); err != nil {
		t.Fatal(err)
	}
	state.Empires[0].Research.ProgressRP = 12.875

	switchTo55, err := NewSelectResearchCommand(2, SelectResearchPayload{TechFieldID: 55})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.selectResearch(state, state.Empires[0].ID, 1, switchTo55)
	if err != nil {
		t.Fatal(err)
	}
	if state.Empires[0].Research.ProgressRP != 12.875 {
		t.Fatalf("fractional RP was rounded/truncated on switch: %v", state.Empires[0].Research.ProgressRP)
	}
	var payload ResearchSwitchedEvent
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.TransferredRP != 12.875 {
		t.Fatalf("fractional transferred_rp=%v want=12.875", payload.TransferredRP)
	}
}
