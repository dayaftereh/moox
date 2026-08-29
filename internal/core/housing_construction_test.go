package core

import (
	"reflect"
	"testing"
)

func TestHousingConstructionStateRoundTripsInSchema12(t *testing.T) {
	state := NewSmallFixture(0xA120)
	state.Colonies[0].Construction = &ConstructionState{
		ProjectKind: ConstructionProjectHousing,
		ProjectID:   "housing",
	}
	if state.SchemaVersion != StateSchemaVersion {
		t.Fatalf("schema=%d constant=%d", state.SchemaVersion, StateSchemaVersion)
	}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	encoded, err := MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state, loaded) {
		t.Fatalf("Housing state changed across round trip\nstate=%+v\nloaded=%+v", state.Colonies[0].Construction, loaded.Colonies[0].Construction)
	}
}

func TestHousingConstructionStateRejectsInvalidIdentityOrProgress(t *testing.T) {
	for _, tc := range []struct {
		name    string
		project *ConstructionState
	}{
		{"wrong id", &ConstructionState{ProjectKind: ConstructionProjectHousing, ProjectID: "cloning_center"}},
		{"progress", &ConstructionState{ProjectKind: ConstructionProjectHousing, ProjectID: "housing", ProgressPP: 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := NewSmallFixture(0xA121)
			state.Colonies[0].Construction = tc.project
			if err := state.Validate(); err == nil {
				t.Fatalf("expected invalid Housing state to fail: %+v", tc.project)
			}
		})
	}
}
