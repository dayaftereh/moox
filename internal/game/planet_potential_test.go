package game

import (
	"path/filepath"
	"testing"

	"moox/internal/core"
)

func TestPlanetPotentialForEmpireUsesAuthoritativePlanetAndRaceRules(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0x8008)
	empire := state.Empires[0]
	planet := state.Galaxy.Systems[0].Planets[0]

	potential, err := rules.PlanetPotentialForEmpire(planet, empire)
	if err != nil {
		t.Fatal(err)
	}
	if potential.PlanetID != planet.ID {
		t.Fatalf("planet id = %d, want %d", potential.PlanetID, planet.ID)
	}
	if potential.FoodPerFarmer != 2 {
		t.Fatalf("food/farmer = %.2f, want 2", potential.FoodPerFarmer)
	}
	if potential.ProductionPerWorker != 3 {
		t.Fatalf("production/worker = %.2f, want 3", potential.ProductionPerWorker)
	}
	if potential.ResearchPerScientist != 3 {
		t.Fatalf("research/scientist = %.2f, want 3", potential.ResearchPerScientist)
	}
	if potential.GravityPenaltyPercent != 0 {
		t.Fatalf("gravity penalty = %d, want 0", potential.GravityPenaltyPercent)
	}
	if potential.ClimateHabitabilityPercent != 80 {
		t.Fatalf("climate habitability = %d, want 80", potential.ClimateHabitabilityPercent)
	}
	if potential.SizeBaseCapacity != 15 {
		t.Fatalf("size base capacity = %.2f, want 15", potential.SizeBaseCapacity)
	}
	if potential.PopulationCapacity != 12 {
		t.Fatalf("population capacity = %.2f, want 12", potential.PopulationCapacity)
	}
}
