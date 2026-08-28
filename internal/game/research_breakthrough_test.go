package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
)

func TestResearchBreakthroughChancePercentOriginalIntegerCurve(t *testing.T) {
	cases := []struct {
		name      string
		base      float64
		projected float64
		want      int
	}{
		{"invalid base", 0, 100, 0},
		{"below base", 50, 49, 0},
		{"exact base", 50, 50, 0},
		{"one over fifty", 50, 51, 2},
		{"minimum one percent after floor", 150, 151, 1},
		{"halfway to double", 150, 225, 50},
		{"double cost", 150, 300, 100},
		{"above double capped", 150, 450, 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResearchBreakthroughChancePercent(tc.base, tc.projected); got != tc.want {
				t.Fatalf("chance(base=%g projected=%g)=%d want=%d", tc.base, tc.projected, got, tc.want)
			}
		})
	}
}

func TestActiveResearchConsumesOneRollEvenAtZeroPercent(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(601)
	state.Empires[0].Research = &core.ResearchState{TechFieldID: 56, SelectionMode: core.ResearchSelectionChooseOne, TechnologyIDs: []int{155}}

	expected := state.RNG()
	zeroBased, err := expected.Intn(100)
	if err != nil {
		t.Fatal(err)
	}
	events, err := resolver.advanceResearch(state)
	if err != nil {
		t.Fatal(err)
	}
	if state.RNGState != expected.State() {
		t.Fatalf("RNG state=%d want=%d after active zero-percent research", state.RNGState, expected.State())
	}
	if len(events) != 1 || events[0].Kind != "empire.research_progressed" {
		t.Fatalf("unexpected events: %+v", events)
	}
	var progress ResearchProgressedEvent
	if err := json.Unmarshal(events[0].Data, &progress); err != nil {
		t.Fatal(err)
	}
	if progress.ChancePercent != 0 || progress.Roll != zeroBased+1 || progress.Breakthrough {
		t.Fatalf("unexpected zero-percent roll event: %+v", progress)
	}
	if !reflect.DeepEqual(progress.TechnologyKeys, []string{"research_laboratory"}) {
		t.Fatalf("speaking technology keys=%v", progress.TechnologyKeys)
	}
}

func TestNoActiveResearchDoesNotConsumeRNG(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(602)
	before := state.RNGState
	events, err := resolver.advanceResearch(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 || state.RNGState != before {
		t.Fatalf("idle research changed RNG/events: rng=%d want=%d events=%+v", state.RNGState, before, events)
	}
}

func TestEmpireResearchPreservesFractionalColonyOutput(t *testing.T) {
	state := core.NewSmallFixture(603)
	empireID := state.Empires[0].ID
	state.Colonies[0].AdjustedEconomy.Research = 1.9
	second := state.Colonies[0]
	second.ID = state.NextID + 100
	second.AdjustedEconomy.Research = 1.9
	state.Colonies = append(state.Colonies, second)
	if got := implementedEmpireResearchRP(state, empireID); got != 3.8 {
		t.Fatalf("fractional empire research=%v want=3.8", got)
	}
}

func TestGuaranteedBreakthroughAcquiresSpeakingTechnologyAndUnlocksBuilding(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(604)
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	state.Empires[0].Research = &core.ResearchState{
		TechFieldID:   56,
		SelectionMode: core.ResearchSelectionChooseOne,
		TechnologyIDs: []int{155},
		ProgressRP:    299,
	}
	state.Colonies[0].AdjustedEconomy.Research = 1

	events, err := resolver.advanceResearch(state)
	if err != nil {
		t.Fatal(err)
	}
	if state.Empires[0].Research != nil {
		t.Fatalf("100%% breakthrough did not clear ResearchState: %+v", state.Empires[0].Research)
	}
	if !containsTechnology(state.Empires[0].KnownTechnologyIDs, 155) {
		t.Fatalf("Technology 155 (research_laboratory) not acquired: %v", state.Empires[0].KnownTechnologyIDs)
	}
	if len(events) != 2 || events[0].Kind != "empire.research_progressed" || events[1].Kind != "empire.research_completed" {
		t.Fatalf("unexpected breakthrough events: %+v", events)
	}
	var progress ResearchProgressedEvent
	if err := json.Unmarshal(events[0].Data, &progress); err != nil {
		t.Fatal(err)
	}
	if progress.BaseCostRP != 150 || progress.PreviousRP != 299 || progress.TurnResearchRP != 1 || progress.ProjectedRP != 300 || progress.ChancePercent != 100 || !progress.Breakthrough {
		t.Fatalf("unexpected guaranteed breakthrough: %+v", progress)
	}
	if !reflect.DeepEqual(progress.TechnologyKeys, []string{"research_laboratory"}) {
		t.Fatalf("progress technology keys=%v", progress.TechnologyKeys)
	}
	var completed ResearchCompletedEvent
	if err := json.Unmarshal(events[1].Data, &completed); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(completed.TechnologyIDs, []int{155}) || !reflect.DeepEqual(completed.TechnologyKeys, []string{"research_laboratory"}) {
		t.Fatalf("completion identifiers=%+v", completed)
	}

	choices, err := rules.AvailableBuildingChoices(state, state.Empires[0].ID, state.Colonies[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if !buildingChoiceContains(choices, "research_laboratory") {
		t.Fatal("Building research_laboratory (Production ID 35) did not become legal after Technology 155")
	}
}

func TestResearchResolutionIsReplayDeterministic(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	makeState := func() *core.GameState {
		state := core.NewSmallFixture(605)
		state.Empires[0].Research = &core.ResearchState{TechFieldID: 56, SelectionMode: core.ResearchSelectionChooseOne, TechnologyIDs: []int{155}, ProgressRP: 159}
		state.Colonies[0].AdjustedEconomy.Research = 1
		return state
	}
	first := makeState()
	second := makeState()
	firstEvents, err := resolver.advanceResearch(first)
	if err != nil {
		t.Fatal(err)
	}
	secondEvents, err := resolver.advanceResearch(second)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(firstEvents, secondEvents) {
		t.Fatalf("same seed/state produced different research events:\nfirst=%+v\nsecond=%+v", firstEvents, secondEvents)
	}
	firstJSON, err := core.MarshalState(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := core.MarshalState(second)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatal("same seed/state produced different research state/RNG serialization")
	}
}

func TestAutomaticResearchRepeatsHyperAdvancedFieldWithDynamicCost(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(606)
	state.Empires[0].RaceID = "human"
	state.Empires[0].KnownTechnologyFieldIDs = []int{70}
	state.Empires[0].Research = &core.ResearchState{TechFieldID: 75, SelectionMode: core.ResearchSelectionRepeatField}
	state.Colonies[0].AdjustedEconomy.Research = 50000

	first, err := resolver.advanceResearch(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 2 {
		t.Fatalf("first Hyper-Advanced turn events=%d want=2: %+v", len(first), first)
	}
	var firstProgress ResearchProgressedEvent
	if err := json.Unmarshal(first[0].Data, &firstProgress); err != nil {
		t.Fatal(err)
	}
	if firstProgress.BaseCostRP != 15000 || firstProgress.CompletedLevels != 0 || firstProgress.ResearchLevel != 1 || !firstProgress.Breakthrough {
		t.Fatalf("first Hyper-Advanced progress=%+v", firstProgress)
	}
	var firstComplete ResearchCompletedEvent
	if err := json.Unmarshal(first[1].Data, &firstComplete); err != nil {
		t.Fatal(err)
	}
	if firstComplete.CompletedLevels != 1 || firstComplete.ResearchLevel != 1 || len(firstComplete.TechnologyIDs) != 0 {
		t.Fatalf("first Hyper-Advanced completion=%+v", firstComplete)
	}
	if state.Empires[0].Research == nil || state.Empires[0].Research.TechFieldID != 75 || state.Empires[0].Research.SelectionMode != core.ResearchSelectionRepeatField || state.Empires[0].Research.ProgressRP != 0 || len(state.Empires[0].Research.TechnologyIDs) != 0 {
		t.Fatalf("Hyper-Advanced research did not remain active/reset: %+v", state.Empires[0].Research)
	}
	if level, ok := hyperAdvancedCompletedLevels(&state.Empires[0], 75); !ok || level != 1 {
		t.Fatalf("Hyper-Advanced field 75 completed level=%d ok=%v state=%v", level, ok, state.Empires[0].HyperAdvancedResearch)
	}
	if containsInt(state.Empires[0].KnownTechnologyFieldIDs, 75) {
		t.Fatalf("repeatable Hyper-Advanced field was marked permanently known: %v", state.Empires[0].KnownTechnologyFieldIDs)
	}

	second, err := resolver.advanceResearch(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 2 {
		t.Fatalf("second Hyper-Advanced turn events=%d want=2: %+v", len(second), second)
	}
	var secondProgress ResearchProgressedEvent
	if err := json.Unmarshal(second[0].Data, &secondProgress); err != nil {
		t.Fatal(err)
	}
	if secondProgress.BaseCostRP != 25000 || secondProgress.CompletedLevels != 1 || secondProgress.ResearchLevel != 2 || !secondProgress.Breakthrough {
		t.Fatalf("second Hyper-Advanced progress=%+v", secondProgress)
	}
	if level, ok := hyperAdvancedCompletedLevels(&state.Empires[0], 75); !ok || level != 2 {
		t.Fatalf("Hyper-Advanced field 75 second level=%d ok=%v", level, ok)
	}
}
