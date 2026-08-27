package core

import (
	"math"
	"testing"
)

func TestStateAcceptsFractionalPopulationAndEconomy(t *testing.T) {
	state := NewSmallFixture(710)
	state.Colonies[0].Population = PopulationState{Total: 4, Farmers: 1.25, Workers: 1.5, Scientists: 1.25}
	state.Colonies[0].Economy = ColonyEconomy{Food: 2.5, Production: 4.5, Research: 3.75, TaxBC: 4}
	state.Colonies[0].AdjustedEconomy = ColonyEconomy{Food: 2.5, Production: 4.5, Research: 5.625, TaxBC: 6}
	state.Colonies[0].Construction = &ConstructionState{BuildingID: "holo_simulator", ProgressPP: 2.75}
	if err := state.Validate(); err != nil {
		t.Fatalf("fractional continuous state should validate: %v", err)
	}
}

func TestStateRejectsInvalidContinuousValues(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*GameState)
	}{
		{"population nan", func(s *GameState) { s.Colonies[0].Population.Total = math.NaN() }},
		{"population infinity", func(s *GameState) { s.Colonies[0].Population.Farmers = math.Inf(1) }},
		{"economy nan", func(s *GameState) { s.Colonies[0].Economy.Food = math.NaN() }},
		{"adjusted economy infinity", func(s *GameState) { s.Colonies[0].AdjustedEconomy.Research = math.Inf(1) }},
		{"construction nan", func(s *GameState) {
			s.Colonies[0].Construction = &ConstructionState{BuildingID: "holo_simulator", ProgressPP: math.NaN()}
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
