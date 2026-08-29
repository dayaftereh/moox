package core

import (
	"reflect"
	"testing"
)

func TestPlanetaryTransformationStateRoundTripsInCurrentSchema(t *testing.T) {
	state := NewSmallFixture(0xA200)
	state.Colonies[0].Construction = &ConstructionState{
		ProjectKind: ConstructionProjectPlanetaryTransformation,
		ProjectID:   "terraforming",
		ProgressPP:  123.5,
	}
	if StateSchemaVersion != 15 || state.SchemaVersion != 15 {
		t.Fatalf("schema=%d constant=%d want=15", state.SchemaVersion, StateSchemaVersion)
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
		t.Fatalf("planetary transformation changed across round trip\nstate=%+v\nloaded=%+v", state.Colonies[0].Construction, loaded.Colonies[0].Construction)
	}
}

func TestPlanetaryTransformationStateRejectsUnknownProject(t *testing.T) {
	state := NewSmallFixture(0xA201)
	state.Colonies[0].Construction = &ConstructionState{
		ProjectKind: ConstructionProjectPlanetaryTransformation,
		ProjectID:   "unknown_transformation",
	}
	if err := state.Validate(); err == nil {
		t.Fatal("unknown planetary transformation unexpectedly validated")
	}
}
