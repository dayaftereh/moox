package core

import "testing"

func TestResearchProgressRequiresWholeRP(t *testing.T) {
	state := NewSmallFixture(607)
	state.Empires[0].Research = &ResearchState{TechFieldID: 56, TechnologyIDs: []int{155}, ProgressMilli: 150001}
	if err := state.Validate(); err == nil {
		t.Fatal("expected fractional milli-RP research progress to fail validation")
	}
}
