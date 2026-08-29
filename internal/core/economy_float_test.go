package core

import (
	"math"
	"testing"
)

func TestStateAcceptsFractionalPopulationAndEconomy(t *testing.T) {
	state := NewSmallFixture(710)
	state.Colonies[0].Population = NewAssimilatedPopulation(state.Empires[0].ID, 1.25, 1.5, 1.25)
	state.Colonies[0].Economy = ColonyEconomy{Food: 2.5, Production: 4.5, Research: 3.75, TaxBC: 4}
	state.Colonies[0].AdjustedEconomy = ColonyEconomy{Food: 2.5, Production: 4.5, Research: 5.625, TaxBC: 6}
	state.Colonies[0].Construction = &ConstructionState{ProjectKind: ConstructionProjectBuilding, ProjectID: "holo_simulator", ProgressPP: 2.75}
	if err := state.Validate(); err != nil {
		t.Fatalf("fractional continuous state should validate: %v", err)
	}
}

func TestStateRejectsInvalidContinuousValues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*GameState)
	}{
		{"population nan", func(s *GameState) { s.Colonies[0].Population.Cohorts[0].Workers = math.NaN() }},
		{"population infinity", func(s *GameState) { s.Colonies[0].Population.Cohorts[0].Farmers = math.Inf(1) }},
		{"economy nan", func(s *GameState) { s.Colonies[0].Economy.Food = math.NaN() }},
		{"adjusted economy infinity", func(s *GameState) { s.Colonies[0].AdjustedEconomy.Research = math.Inf(1) }},
		{"population dynamics nan", func(s *GameState) { s.Colonies[0].PopulationDynamics.ProjectedGrowth = math.NaN() }},
		{"population dynamics infinity", func(s *GameState) { s.Colonies[0].PopulationDynamics.FoodRequired = math.Inf(1) }},
		{"population dynamics conflicting food", func(s *GameState) {
			s.Colonies[0].PopulationDynamics.FoodSurplus = 1
			s.Colonies[0].PopulationDynamics.FoodShortage = 1
		}},
		{"population dynamics conflicting production", func(s *GameState) {
			s.Colonies[0].PopulationDynamics.ProductionAvailable = 1
			s.Colonies[0].PopulationDynamics.ProductionShortage = 1
		}},
		{"population dynamics growth beyond capacity", func(s *GameState) {
			s.Colonies[0].PopulationDynamics.Capacity = 4
			s.Colonies[0].PopulationDynamics.ProjectedGrowth = 0.1
		}}, {"construction nan", func(s *GameState) {
			s.Colonies[0].Construction = &ConstructionState{ProjectKind: ConstructionProjectBuilding, ProjectID: "holo_simulator", ProgressPP: math.NaN()}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := NewSmallFixture(711)
			tc.mutate(state)
			if err := state.Validate(); err == nil {
				t.Fatal("expected invalid continuous value to fail validation")
			}
		})
	}
}

func TestStateRejectsInvalidFoodLogistics(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*GameState)
	}{
		{"food logistics nan", func(s *GameState) { s.Empires[0].FoodLogistics.SurplusFoodIncomeBC = math.NaN() }},
		{"negative freighters", func(s *GameState) { s.Empires[0].Freighters = -1 }},
		{"food import exceeds local shortage", func(s *GameState) {
			s.Colonies[0].PopulationDynamics.LocalFoodShortage = 1
			s.Colonies[0].PopulationDynamics.FoodImported = 2
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := NewSmallFixture(735)
			tc.mutate(state)
			if err := state.Validate(); err == nil {
				t.Fatal("expected invalid food logistics state to fail validation")
			}
		})
	}
}
