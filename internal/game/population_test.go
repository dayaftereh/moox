package game

import (
	"math"
	"testing"

	"moox/internal/core"
)

func closePopulationValue(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9*math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
}

func TestPopulationCapacityUsesSizeClimateAndRaceTraits(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	planet := core.Planet{SizeID: "medium", ClimateID: "terran"}

	tests := []struct {
		name    string
		raceID  string
		climate string
		want    float64
	}{
		{"human terran", "human", "terran", 12},
		{"human barren", "human", "barren", 4},
		{"aquatic ocean", "trilarian", "ocean", 15},
		{"tolerant barren", "silicoid", "barren", 8},
		{"subterranean terran", "sakkra", "terran", 18},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			planet.ClimateID = tc.climate
			got, err := rules.PopulationCapacity(planet, tc.raceID)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("capacity=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestPopulationRaceGrowthMultipliersAreNormalized(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	for raceID, want := range map[string]float64{
		"human":    1,
		"sakkra":   2,
		"silicoid": 0.5,
	} {
		if got := rules.RaceModifiers[raceID].PopulationGrowthMultiplier; got != want {
			t.Fatalf("race %s growth multiplier=%v want=%v", raceID, got, want)
		}
	}
}

func TestPopulationSustenanceUsesDomainNativeFoodAndProduction(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	planet := core.Planet{SizeID: "medium", ClimateID: "terran"}
	colony := core.Colony{Population: core.PopulationState{Total: 4, Farmers: 2, Workers: 1, Scientists: 1}}

	human, err := rules.CalculatePopulationDynamics(colony, planet, "human", core.ColonyEconomy{Food: 4, Production: 3})
	if err != nil {
		t.Fatal(err)
	}
	if human.FoodRequired != 4 || human.FoodShortage != 0 || human.ProductionRequired != 0 || human.ProductionAvailable != 3 {
		t.Fatalf("human sustenance=%+v", human)
	}

	meklar, err := rules.CalculatePopulationDynamics(colony, planet, "meklar", core.ColonyEconomy{Food: 2, Production: 4})
	if err != nil {
		t.Fatal(err)
	}
	if meklar.FoodRequired != 2 || meklar.ProductionRequired != 2 || meklar.ProductionAvailable != 2 || meklar.FoodShortage != 0 || meklar.ProductionShortage != 0 {
		t.Fatalf("Meklar sustenance=%+v", meklar)
	}

	silicoid, err := rules.CalculatePopulationDynamics(colony, planet, "silicoid", core.ColonyEconomy{Food: 0, Production: 3})
	if err != nil {
		t.Fatal(err)
	}
	if silicoid.FoodRequired != 0 || silicoid.FoodShortage != 0 || silicoid.ProductionRequired != 0 || silicoid.ProductionAvailable != 3 {
		t.Fatalf("Silicoid sustenance=%+v", silicoid)
	}
}

func TestPopulationGrowthUsesDirectFloatPopulationUnits(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	planet := core.Planet{SizeID: "medium", ClimateID: "terran"}
	colony := core.Colony{Population: core.PopulationState{Total: 4, Farmers: 2, Workers: 1, Scientists: 1}}
	got, err := rules.CalculatePopulationDynamics(colony, planet, "human", core.ColonyEconomy{Food: 4, Production: 3})
	if err != nil {
		t.Fatal(err)
	}
	want := math.Sqrt(0.002 * 4 * (12 - 4) / 12)
	if !closePopulationValue(got.BaseGrowth, want) || !closePopulationValue(got.ProjectedGrowth, want) {
		t.Fatalf("growth=%+v want=%v", got, want)
	}
	if got.ProjectedGrowth <= 0 || got.ProjectedGrowth >= 1 {
		t.Fatalf("growth should be direct fractional population, got %v", got.ProjectedGrowth)
	}
}

func TestPopulationGrowthAppliesRaceMultiplierAndStopsOnShortage(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	planet := core.Planet{SizeID: "medium", ClimateID: "terran"}
	colony := core.Colony{Population: core.PopulationState{Total: 4, Farmers: 2, Workers: 1, Scientists: 1}}

	sakkra, err := rules.CalculatePopulationDynamics(colony, planet, "sakkra", core.ColonyEconomy{Food: 4, Production: 3})
	if err != nil {
		t.Fatal(err)
	}
	wantBase := math.Sqrt(0.002 * 4 * (18 - 4) / 18)
	if !closePopulationValue(sakkra.BaseGrowth, wantBase) || !closePopulationValue(sakkra.ProjectedGrowth, wantBase*2) {
		t.Fatalf("Sakkra growth=%+v want base=%v projected=%v", sakkra, wantBase, wantBase*2)
	}

	starved, err := rules.CalculatePopulationDynamics(colony, planet, "human", core.ColonyEconomy{Food: 3.5, Production: 3})
	if err != nil {
		t.Fatal(err)
	}
	if starved.FoodShortage != 0.5 || starved.ProjectedGrowth != 0 {
		t.Fatalf("starved population dynamics=%+v", starved)
	}

	cyber, err := rules.CalculatePopulationDynamics(colony, planet, "meklar", core.ColonyEconomy{Food: 2, Production: 1.5})
	if err != nil {
		t.Fatal(err)
	}
	if cyber.ProductionShortage != 0.5 || cyber.ProjectedGrowth != 0 {
		t.Fatalf("production-starved Cybernetic dynamics=%+v", cyber)
	}
}

func TestAdvancePopulationPreservesFractionalJobShares(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(720)
	colony := &state.Colonies[0]
	colony.Population = core.PopulationState{Total: 4, Farmers: 1.25, Workers: 1.5, Scientists: 1.25}
	colony.PopulationDynamics = core.ColonyPopulationDynamics{Capacity: 12, ProjectedGrowth: 0.08, GrowthMultiplier: 1}

	events, err := resolver.advancePopulation(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != "colony.population_grew" {
		t.Fatalf("events=%+v", events)
	}
	if !closePopulationValue(colony.Population.Total, 4.08) {
		t.Fatalf("total=%v want=4.08", colony.Population.Total)
	}
	assigned := colony.Population.Farmers + colony.Population.Workers + colony.Population.Scientists
	if !closePopulationValue(assigned, colony.Population.Total) {
		t.Fatalf("assignments=%v total=%v", assigned, colony.Population.Total)
	}
	if !closePopulationValue(colony.Population.Farmers/colony.Population.Total, 1.25/4) || !closePopulationValue(colony.Population.Workers/colony.Population.Total, 1.5/4) {
		t.Fatalf("job shares changed: %+v", colony.Population)
	}
}
