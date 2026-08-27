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
		base      int64
		projected int64
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
				t.Fatalf("chance(base=%d projected=%d)=%d want=%d", tc.base, tc.projected, got, tc.want)
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
	state.Empires[0].Research = &core.ResearchState{TechFieldID: 56, TechnologyIDs: []int{155}}

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

func TestEmpireResearchQuantizesEachColonyBeforeSumming(t *testing.T) {
	state := core.NewSmallFixture(603)
	empireID := state.Empires[0].ID
	state.Colonies[0].AdjustedEconomy.ResearchMilli = 1900
	second := state.Colonies[0]
	second.ID = state.NextID + 100
	second.AdjustedEconomy.ResearchMilli = 1900
	state.Colonies = append(state.Colonies, second)
	if got := implementedEmpireResearchRP(state, empireID); got != 2 {
		t.Fatalf("integer empire research=%d want=2; summing millis before truncation would incorrectly yield 3", got)
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
		TechnologyIDs: []int{155},
		ProgressMilli: 299 * core.EconomyScale,
	}
	state.Colonies[0].AdjustedEconomy.ResearchMilli = 1000

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
		state.Empires[0].Research = &core.ResearchState{TechFieldID: 56, TechnologyIDs: []int{155}, ProgressMilli: 159 * core.EconomyScale}
		state.Colonies[0].AdjustedEconomy.ResearchMilli = 1000
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

func TestAutomaticResearchRejectsUnmodeledHyperAdvancedCostScaling(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(606)
	state.Empires[0].Research = &core.ResearchState{TechFieldID: 75, TechnologyIDs: []int{155}}
	if _, err := resolver.advanceResearch(state); err == nil {
		t.Fatal("expected hyper-advanced dynamic cost scaling to remain explicitly unsupported")
	}
}
