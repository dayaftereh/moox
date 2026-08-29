package core

import (
	"reflect"
	"testing"
)

func TestPopulationCohortsRoundTripDeterministically(t *testing.T) {
	state := NewSmallFixture(1401)
	owner := state.Empires[0].ID
	foreign := Empire{ID: state.NewID(), Name: "Foreign", RaceID: "alkari"}
	state.Empires = append(state.Empires, foreign)
	state.Colonies[0].Population = PopulationState{Cohorts: []PopulationCohort{
		{OriginEmpireID: foreign.ID, LoyaltyEmpireID: foreign.ID, AssimilationState: PopulationConquered, Workers: 0.5},
		{OriginEmpireID: owner, LoyaltyEmpireID: owner, AssimilationState: PopulationAssimilated, Farmers: 2, Workers: 1},
		{OriginEmpireID: foreign.ID, LoyaltyEmpireID: owner, AssimilationState: PopulationAssimilated, Scientists: 0.5},
	}}
	state.Colonies[0].Population.Normalize()

	if StateSchemaVersion != 14 || state.SchemaVersion != 14 {
		t.Fatalf("schema=%d constant=%d want=14", state.SchemaVersion, StateSchemaVersion)
	}
	if got := state.Colonies[0].Population.Total(); got != 4 {
		t.Fatalf("total=%g want=4", got)
	}
	if got := state.Colonies[0].Population.Farmers(); got != 2 {
		t.Fatalf("farmers=%g want=2", got)
	}
	if got := state.Colonies[0].Population.Workers(); got != 1.5 {
		t.Fatalf("workers=%g want=1.5", got)
	}
	if got := state.Colonies[0].Population.Scientists(); got != 0.5 {
		t.Fatalf("scientists=%g want=0.5", got)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("mixed cohort state should validate: %v", err)
	}

	encoded, err := MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.Colonies[0].Population, loaded.Colonies[0].Population) {
		t.Fatalf("population changed across roundtrip\nstate=%+v\nloaded=%+v", state.Colonies[0].Population, loaded.Colonies[0].Population)
	}
}

func TestPopulationCohortValidationRejectsInvalidSemanticState(t *testing.T) {
	t.Run("duplicate", func(t *testing.T) {
		state := NewSmallFixture(1402)
		state.Colonies[0].Population.Cohorts = append(state.Colonies[0].Population.Cohorts, state.Colonies[0].Population.Cohorts[0])
		if err := state.Validate(); err == nil {
			t.Fatal("expected duplicate cohort key to fail")
		}
	})

	t.Run("owner conquered", func(t *testing.T) {
		state := NewSmallFixture(1403)
		state.Colonies[0].Population.Cohorts[0].AssimilationState = PopulationConquered
		if err := state.Validate(); err == nil {
			t.Fatal("expected owner-origin conquered population to fail")
		}
	})

	t.Run("unknown origin", func(t *testing.T) {
		state := NewSmallFixture(1404)
		state.Colonies[0].Population.Cohorts[0].OriginEmpireID = 999999
		if err := state.Validate(); err == nil {
			t.Fatal("expected unknown origin empire to fail")
		}
	})
}

func TestSetAggregateJobsPreservesCohortIdentitiesAndTotals(t *testing.T) {
	owner := ID(1)
	foreign := ID(2)
	population := PopulationState{Cohorts: []PopulationCohort{
		{OriginEmpireID: owner, LoyaltyEmpireID: owner, AssimilationState: PopulationAssimilated, Farmers: 2},
		{OriginEmpireID: foreign, LoyaltyEmpireID: owner, AssimilationState: PopulationAssimilated, Workers: 2},
	}}
	population.Normalize()
	before := population.TotalsByOrigin()
	if err := population.SetAggregateJobs(1, 2, 1); err != nil {
		t.Fatal(err)
	}
	if got := population.TotalsByOrigin(); !reflect.DeepEqual(before, got) {
		t.Fatalf("origin totals changed: before=%v after=%v", before, got)
	}
	if population.Farmers() != 1 || population.Workers() != 2 || population.Scientists() != 1 {
		t.Fatalf("unexpected aggregate jobs: %+v", population)
	}
}
