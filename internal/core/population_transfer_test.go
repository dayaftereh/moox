package core

import (
	"reflect"
	"testing"
)

func TestPopulationTransferStateRoundTripsExactly(t *testing.T) {
	state := NewSmallFixture(914)
	planet := &state.Galaxy.Systems[1].Planets[0]
	second := Colony{
		ID: state.NewID(), EmpireID: state.Empires[0].ID, PlanetID: planet.ID,
		Population: PopulationState{Total: 1, Workers: 1},
	}
	planet.ColonyID = second.ID
	state.Colonies = append(state.Colonies, second)
	state.PopulationTransfers = []PopulationTransfer{{
		ID: state.NewID(), EmpireID: state.Empires[0].ID, SourceColonyID: state.Colonies[0].ID,
		DestinationColonyID: second.ID, Job: PopulationJobWorker, RemainingTurns: 3,
	}}
	encoded, err := MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state, loaded) {
		t.Fatalf("Population transfer state changed across roundtrip\nstate=%+v\nloaded=%+v", state.PopulationTransfers, loaded.PopulationTransfers)
	}
}
