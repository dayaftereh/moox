package core

import (
	"bytes"
	"reflect"
	"testing"
)

func TestHyperAdvancedResearchStateRoundTripsExactly(t *testing.T) {
	state := NewSmallFixture(780)
	state.Empires[0].KnownTechnologyFieldIDs = []int{70}
	state.Empires[0].HyperAdvancedResearch = []HyperAdvancedResearchLevel{
		{TechFieldID: 75, CompletedLevels: 2},
		{TechFieldID: 82, CompletedLevels: 5},
	}
	state.Empires[0].Research = &ResearchState{TechFieldID: 75, SelectionMode: ResearchSelectionRepeatField, ProgressRP: 1234.5}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	encoded, err := MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(encoded, []byte(`"hyper_advanced_research"`)) || !bytes.Contains(encoded, []byte(`"completed_levels": 2`)) && !bytes.Contains(encoded, []byte(`"completed_levels":2`)) {
		t.Fatalf("schema-7 Hyper-Advanced state missing from JSON: %s", encoded)
	}
	loaded, err := UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state, loaded) {
		t.Fatalf("Hyper-Advanced state changed across round-trip:\nwant=%+v\ngot=%+v", state.Empires[0], loaded.Empires[0])
	}
}

func TestHyperAdvancedResearchStateValidation(t *testing.T) {
	cases := []struct {
		name   string
		levels []HyperAdvancedResearchLevel
	}{
		{"field below range", []HyperAdvancedResearchLevel{{TechFieldID: 74, CompletedLevels: 1}}},
		{"field above range", []HyperAdvancedResearchLevel{{TechFieldID: 83, CompletedLevels: 1}}},
		{"zero level", []HyperAdvancedResearchLevel{{TechFieldID: 75, CompletedLevels: 0}}},
		{"negative level", []HyperAdvancedResearchLevel{{TechFieldID: 75, CompletedLevels: -1}}},
		{"duplicate field", []HyperAdvancedResearchLevel{{TechFieldID: 75, CompletedLevels: 1}, {TechFieldID: 75, CompletedLevels: 2}}},
		{"unsorted fields", []HyperAdvancedResearchLevel{{TechFieldID: 76, CompletedLevels: 1}, {TechFieldID: 75, CompletedLevels: 2}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := NewSmallFixture(781)
			state.Empires[0].HyperAdvancedResearch = tc.levels
			if err := state.Validate(); err == nil {
				t.Fatalf("expected invalid Hyper-Advanced state to fail: %+v", tc.levels)
			}
		})
	}
}

func TestRepeatFieldResearchValidation(t *testing.T) {
	valid := NewSmallFixture(782)
	valid.Empires[0].Research = &ResearchState{TechFieldID: 75, SelectionMode: ResearchSelectionRepeatField, ProgressRP: 1.25}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid repeat-field research rejected: %v", err)
	}

	withTechnology := NewSmallFixture(783)
	withTechnology.Empires[0].Research = &ResearchState{TechFieldID: 75, SelectionMode: ResearchSelectionRepeatField, TechnologyIDs: []int{155}}
	if err := withTechnology.Validate(); err == nil {
		t.Fatal("repeat-field research accepted a concrete technology")
	}

	normalField := NewSmallFixture(784)
	normalField.Empires[0].Research = &ResearchState{TechFieldID: 56, SelectionMode: ResearchSelectionRepeatField}
	if err := normalField.Validate(); err == nil {
		t.Fatal("repeat-field mode accepted a normal TechField")
	}
}
