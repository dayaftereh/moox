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
			t.Fatalf("step %v: %v != %v", i, av, bv)
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
		t.Fatalf("first=%v want=%v", first, want)
	}
	if want := fresh.Uint64(); second != want {
		t.Fatalf("second=%v want=%v", second, want)
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
		t.Fatalf("turn=%v events=%+v", state.Turn, state.Events)
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
			t.Fatalf("Intn returned %v outside [0,7)", value)
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
	state.Colonies[0].AdjustedEconomy.Production = -1
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
	state.Colonies[0].Construction = &ConstructionState{BuildingID: "holo_simulator", ProgressPP: 42.5}
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

func TestKnownTechnologyIDsValidateAndRoundTrip(t *testing.T) {
	state := NewSmallFixture(403)
	state.Empires[0].KnownTechnologyIDs = []int{22, 86, 141}
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
	if !reflect.DeepEqual(loaded.Empires[0].KnownTechnologyIDs, []int{22, 86, 141}) {
		t.Fatalf("known technologies changed on round-trip: %v", loaded.Empires[0].KnownTechnologyIDs)
	}

	state.Empires[0].KnownTechnologyIDs = []int{86, 22}
	if err := state.Validate(); err == nil {
		t.Fatal("expected unsorted known technologies to fail validation")
	}
	state.Empires[0].KnownTechnologyIDs = []int{86, 86}
	if err := state.Validate(); err == nil {
		t.Fatal("expected duplicate known technologies to fail validation")
	}
	state.Empires[0].KnownTechnologyIDs = []int{204}
	if err := state.Validate(); err == nil {
		t.Fatal("expected out-of-range known technology to fail validation")
	}
}

func TestResearchStateRoundTripsExactly(t *testing.T) {
	state := NewSmallFixture(505)
	state.Empires[0].KnownTechnologyIDs = []int{32, 40, 103, 145, 166, 168}
	state.Empires[0].Research = &ResearchState{TechFieldID: 56, SelectionMode: ResearchSelectionChooseOne, TechnologyIDs: []int{155}, ProgressRP: 42.125}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	encoded, err := MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(encoded, []byte(`"progress_rp": 42.125`)) && !bytes.Contains(encoded, []byte(`"progress_rp":42.125`)) {
		t.Fatalf("schema-6 research JSON does not expose progress_rp: %s", encoded)
	}
	if bytes.Contains(encoded, []byte("progress_milli")) {
		t.Fatalf("legacy research progress_milli leaked into schema-6 JSON: %s", encoded)
	}
	loaded, err := UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state, loaded) {
		t.Fatalf("research state changed across round-trip: want=%+v got=%+v", state.Empires[0].Research, loaded.Empires[0].Research)
	}
	reencoded, err := MarshalState(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatal("research state bytes changed after exact round-trip")
	}
}

func TestKnownTechnologyFieldIDsValidation(t *testing.T) {
	state := NewSmallFixture(509)
	state.Empires[0].KnownTechnologyFieldIDs = []int{0, 22, 29}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	state.Empires[0].KnownTechnologyFieldIDs = []int{29, 22}
	if err := state.Validate(); err == nil {
		t.Fatal("expected unsorted known technology fields to fail validation")
	}
	state.Empires[0].KnownTechnologyFieldIDs = []int{22, 22}
	if err := state.Validate(); err == nil {
		t.Fatal("expected duplicate known technology fields to fail validation")
	}
	state.Empires[0].KnownTechnologyFieldIDs = []int{83}
	if err := state.Validate(); err == nil {
		t.Fatal("expected out-of-range known technology field to fail validation")
	}
}

func TestResearchStateRejectsAlreadyKnownTechnologyState(t *testing.T) {
	state := NewSmallFixture(510)
	state.Empires[0].KnownTechnologyIDs = []int{155}
	state.Empires[0].Research = &ResearchState{TechFieldID: 56, SelectionMode: ResearchSelectionChooseOne, TechnologyIDs: []int{155}}
	if err := state.Validate(); err == nil {
		t.Fatal("expected active research of an already-known technology to fail validation")
	}

	state = NewSmallFixture(511)
	state.Empires[0].KnownTechnologyFieldIDs = []int{56}
	state.Empires[0].Research = &ResearchState{TechFieldID: 56, SelectionMode: ResearchSelectionChooseOne, TechnologyIDs: []int{155}}
	if err := state.Validate(); err == nil {
		t.Fatal("expected active research of an already-known technology field to fail validation")
	}
}
