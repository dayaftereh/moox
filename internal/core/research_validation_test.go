package core

import (
	"math"
	"testing"
)

func TestResearchProgressAllowsFractionalRP(t *testing.T) {
	state := NewSmallFixture(607)
	state.Empires[0].Research = &ResearchState{TechFieldID: 56, SelectionMode: ResearchSelectionChooseOne, TechnologyIDs: []int{155}, ProgressRP: 150.001}
	if err := state.Validate(); err != nil {
		t.Fatalf("fractional research RP should be valid: %v", err)
	}
}

func TestResearchProgressRejectsInvalidFloatValues(t *testing.T) {
	for _, progress := range []float64{-0.001, math.NaN(), math.Inf(1), math.Inf(-1)} {
		state := NewSmallFixture(608)
		state.Empires[0].Research = &ResearchState{TechFieldID: 56, SelectionMode: ResearchSelectionChooseOne, TechnologyIDs: []int{155}, ProgressRP: progress}
		if err := state.Validate(); err == nil {
			t.Fatalf("expected progress %v to fail validation", progress)
		}
	}
}
