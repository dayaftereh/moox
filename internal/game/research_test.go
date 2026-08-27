package game

import (
	"reflect"
	"testing"

	"moox/internal/core"
)

func TestInitializeEmpireTechnologiesFromOriginalStartFields(t *testing.T) {
	rules := loadCommittedEconomyRules(t)

	preWarp := core.NewSmallFixture(501)
	if err := rules.InitializeEmpireTechnologies(&preWarp.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	wantPreWarp := []int{32, 40, 103, 145, 166, 168}
	if !reflect.DeepEqual(preWarp.Empires[0].KnownTechnologyIDs, wantPreWarp) {
		t.Fatalf("pre-warp known technologies=%v want=%v", preWarp.Empires[0].KnownTechnologyIDs, wantPreWarp)
	}
	if wantFields := []int{0, 29}; !reflect.DeepEqual(preWarp.Empires[0].KnownTechnologyFieldIDs, wantFields) {
		t.Fatalf("pre-warp known fields=%v want=%v", preWarp.Empires[0].KnownTechnologyFieldIDs, wantFields)
	}

	averageTactical := core.NewSmallFixture(502)
	if err := rules.InitializeEmpireTechnologies(&averageTactical.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyAverage}); err != nil {
		t.Fatal(err)
	}
	wantAverageTactical := []int{32, 40, 41, 58, 63, 69, 100, 101, 103, 109, 119, 120, 121, 145, 157, 166, 167, 168, 187, 189}
	if !reflect.DeepEqual(averageTactical.Empires[0].KnownTechnologyIDs, wantAverageTactical) {
		t.Fatalf("average tactical known technologies=%v want=%v", averageTactical.Empires[0].KnownTechnologyIDs, wantAverageTactical)
	}
	if wantFields := []int{0, 22, 23, 28, 29, 55, 57}; !reflect.DeepEqual(averageTactical.Empires[0].KnownTechnologyFieldIDs, wantFields) {
		t.Fatalf("average known fields=%v want=%v", averageTactical.Empires[0].KnownTechnologyFieldIDs, wantFields)
	}

	averageStrategic := core.NewSmallFixture(503)
	if err := rules.InitializeEmpireTechnologies(&averageStrategic.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyAverage, StrategicCombat: true}); err != nil {
		t.Fatal(err)
	}
	wantAverageStrategic := []int{32, 40, 41, 58, 69, 100, 101, 103, 109, 119, 120, 121, 145, 157, 166, 167, 168, 187, 189}
	if !reflect.DeepEqual(averageStrategic.Empires[0].KnownTechnologyIDs, wantAverageStrategic) {
		t.Fatalf("average strategic known technologies=%v want original SAVE10 observation=%v", averageStrategic.Empires[0].KnownTechnologyIDs, wantAverageStrategic)
	}
	if containsTechnology(averageStrategic.Empires[0].KnownTechnologyIDs, 63) {
		t.Fatal("strategic Average start incorrectly grants Extended Fuel Tanks")
	}

	advanced := core.NewSmallFixture(504)
	if err := rules.InitializeEmpireTechnologies(&advanced.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyAdvanced}); err == nil {
		t.Fatal("expected unresolved Advanced start generator to remain unsupported")
	}
}

func TestCompleteResearchFieldTransitionsTechnologyAndFieldOwnership(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(505)
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	// Field 56 is Optronics (150 RP), and technology 155 is Research Lab.
	state.Empires[0].Research = &core.ResearchState{TechFieldID: 56, SelectionMode: core.ResearchSelectionChooseOne, TechnologyIDs: []int{155}, ProgressRP: 150}
	event, err := resolver.CompleteResearchField(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != "empire.research_completed" || event.SeatID != 0 || event.CommandSequence != 0 {
		t.Fatalf("unexpected research event: %+v", event)
	}
	if state.Empires[0].Research != nil {
		t.Fatalf("research state not cleared: %+v", state.Empires[0].Research)
	}
	if !containsTechnology(state.Empires[0].KnownTechnologyIDs, 155) || !containsInt(state.Empires[0].KnownTechnologyFieldIDs, 56) {
		t.Fatalf("research completion not materialized: techs=%v fields=%v", state.Empires[0].KnownTechnologyIDs, state.Empires[0].KnownTechnologyFieldIDs)
	}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestCompleteResearchFieldUnlocksBuildingChoice(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(506)
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	colonyID := state.Colonies[0].ID
	before, err := rules.AvailableBuildingChoices(state, state.Empires[0].ID, colonyID)
	if err != nil {
		t.Fatal(err)
	}
	if buildingChoiceContains(before, "research_laboratory") {
		t.Fatal("Research Lab was buildable before Optronics completion")
	}
	state.Empires[0].Research = &core.ResearchState{TechFieldID: 56, SelectionMode: core.ResearchSelectionChooseOne, TechnologyIDs: []int{155}, ProgressRP: 150}
	if _, err := resolver.CompleteResearchField(state, state.Empires[0].ID); err != nil {
		t.Fatal(err)
	}
	after, err := rules.AvailableBuildingChoices(state, state.Empires[0].ID, colonyID)
	if err != nil {
		t.Fatal(err)
	}
	if !buildingChoiceContains(after, "research_laboratory") {
		t.Fatal("Research Lab did not become a legal building choice after technology 155 acquisition")
	}
}

func TestCompleteResearchFieldRejectsPrematureAndCrossFieldCompletion(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}

	premature := core.NewSmallFixture(507)
	premature.Empires[0].Research = &core.ResearchState{TechFieldID: 56, SelectionMode: core.ResearchSelectionChooseOne, TechnologyIDs: []int{155}, ProgressRP: 149}
	if _, err := resolver.CompleteResearchField(premature, premature.Empires[0].ID); err == nil {
		t.Fatal("expected completion below original field base cost to fail")
	}

	crossField := core.NewSmallFixture(508)
	crossField.Empires[0].Research = &core.ResearchState{TechFieldID: 56, SelectionMode: core.ResearchSelectionChooseOne, TechnologyIDs: []int{22}, ProgressRP: 150}
	if _, err := resolver.CompleteResearchField(crossField, crossField.Empires[0].ID); err == nil {
		t.Fatal("expected cross-field research completion to fail")
	}
}

func containsTechnology(ids []int, want int) bool {
	return containsInt(ids, want)
}

func buildingChoiceContains(choices []BuildingChoice, id string) bool {
	for _, choice := range choices {
		if choice.BuildingID == id {
			return true
		}
	}
	return false
}
