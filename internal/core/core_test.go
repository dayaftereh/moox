package core

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRNGIsDeterministic(t *testing.T) {
	a := NewRNG(42)
	b := NewRNG(42)
	for i := 0; i < 32; i++ {
		if av, bv := a.Uint64(), b.Uint64(); av != bv {
			t.Fatalf("step %d: %d != %d", i, av, bv)
		}
	}
	if a.State() == 42 {
		t.Fatal("RNG state did not advance")
	}
}

func TestStateRNGCanBeCheckpointed(t *testing.T) {
	state := NewSmallFixture(7)
	rng := state.RNG()
	first := rng.Uint64()
	state.CommitRNG(rng)
	resumed := state.RNG()
	second := resumed.Uint64()
	fresh := NewRNG(7)
	if want := fresh.Uint64(); first != want {
		t.Fatalf("first=%d want=%d", first, want)
	}
	if want := fresh.Uint64(); second != want {
		t.Fatalf("second=%d want=%d", second, want)
	}
}

func TestSmallFixtureIsStableAndRoundTripsExactly(t *testing.T) {
	first := NewSmallFixture(12345)
	second := NewSmallFixture(12345)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same seed produced different fixture state")
	}
	if err := first.Validate(); err != nil {
		t.Fatal(err)
	}
	encoded, err := MarshalState(first)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, loaded) {
		t.Fatalf("round trip changed state\nfirst=%+v\nloaded=%+v", first, loaded)
	}
	reencoded, err := MarshalState(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatal("serialized bytes changed after round trip")
	}
}

func TestAdvanceTurnAndEventLog(t *testing.T) {
	state := NewSmallFixture(1)
	state.AdvanceTurn()
	state.AddEvent("turn_started", "turn advanced")
	if state.Turn != 2 || state.Events[len(state.Events)-1].Turn != 2 {
		t.Fatalf("turn=%d events=%+v", state.Turn, state.Events)
	}
}

func TestValidateRejectsDanglingReferences(t *testing.T) {
	state := NewSmallFixture(5)
	state.Colonies[0].PlanetID = 9999
	if err := state.Validate(); err == nil {
		t.Fatal("expected dangling planet reference to fail")
	}
}

func TestSaveLoadFileRoundTrip(t *testing.T) {
	state := NewSmallFixture(987654321)
	path := filepath.Join(t.TempDir(), "saves", "fixture.json")
	if err := SaveState(path, state); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadState(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state, loaded) {
		t.Fatalf("file round trip changed state\nstate=%+v\nloaded=%+v", state, loaded)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := SaveState(path, loaded); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("save bytes changed after load/save round trip")
	}
}

func TestIntnRangeAndValidation(t *testing.T) {
	rng := NewRNG(123)
	for i := 0; i < 1000; i++ {
		value, err := rng.Intn(7)
		if err != nil {
			t.Fatal(err)
		}
		if value < 0 || value >= 7 {
			t.Fatalf("Intn returned %d outside [0,7)", value)
		}
	}
	if _, err := rng.Intn(0); err == nil {
		t.Fatal("expected zero bound to fail")
	}
}

func TestValidateRejectsPopulationAssignmentMismatch(t *testing.T) {
	state := NewSmallFixture(19)
	state.Colonies[0].Population.Workers++
	if err := state.Validate(); err == nil {
		t.Fatal("expected mismatched population assignment to fail validation")
	}
}

func TestValidateRejectsInvalidEconomyContext(t *testing.T) {
	state := NewSmallFixture(20)
	state.Colonies[0].EconomyContext.GravityPenaltyPercent = 101
	if err := state.Validate(); err == nil {
		t.Fatal("expected invalid gravity economy context to fail validation")
	}

	state = NewSmallFixture(21)
	state.Colonies[0].AdjustedEconomy.ProductionMilli = -1
	if err := state.Validate(); err == nil {
		t.Fatal("expected negative adjusted economy to fail validation")
	}
}

func TestValidateRejectsInvalidBuildingList(t *testing.T) {
	state := NewSmallFixture(22)
	state.Colonies[0].Buildings = []string{"marine_barracks", "marine_barracks"}
	if err := state.Validate(); err == nil {
		t.Fatal("expected duplicate colony building to fail validation")
	}

	state = NewSmallFixture(23)
	state.Colonies[0].Buildings = []string{""}
	if err := state.Validate(); err == nil {
		t.Fatal("expected empty colony building id to fail validation")
	}
}

func TestConstructionStateRoundTripsExactly(t *testing.T) {
	state := NewSmallFixture(301)
	state.Colonies[0].Construction = &ConstructionState{BuildingID: "holo_simulator", ProgressMilli: 42000}
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
		t.Fatalf("construction state changed across round-trip:\nwant=%+v\ngot=%+v", state.Colonies[0].Construction, loaded.Colonies[0].Construction)
	}
	reencoded, err := MarshalState(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatal("construction state bytes changed after exact round-trip")
	}
}
