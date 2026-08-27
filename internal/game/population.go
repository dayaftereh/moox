package game

import (
	"fmt"
	"math"

	"moox/internal/core"
)

const populationEpsilon = 1e-9

// PopulationCapacity returns the current race-specific population cap for a
// planet. Classic size/climate capacity rounding is retained as an explicit
// gameplay boundary; Population itself remains continuous float64 state.
func (r *EconomyRules) PopulationCapacity(planet core.Planet, raceID string) (float64, error) {
	if r == nil {
		return 0, fmt.Errorf("economy rules must not be nil")
	}
	base, ok := r.PopulationSizeCapacity[planet.SizeID]
	if !ok {
		return 0, fmt.Errorf("unknown population size %q", planet.SizeID)
	}
	sizeClass, ok := r.PopulationSizeClass[planet.SizeID]
	if !ok {
		return 0, fmt.Errorf("unknown population size class %q", planet.SizeID)
	}
	habitability, ok := r.PopulationClimateHabitability[planet.ClimateID]
	if !ok {
		return 0, fmt.Errorf("unknown population climate %q", planet.ClimateID)
	}
	modifiers, ok := r.RaceModifiers[raceID]
	if !ok {
		return 0, fmt.Errorf("unknown race %q", raceID)
	}
	if modifiers.Aquatic {
		switch planet.ClimateID {
		case "tundra", "swamp":
			habitability = math.Max(habitability, 0.80)
		case "ocean", "terran", "gaia":
			habitability = 1
		}
	}
	if modifiers.Tolerant {
		habitability = math.Min(1, habitability+r.TolerantHabitabilityBonus)
	}
	capacity := math.Round(base * habitability)
	if modifiers.Subterranean {
		capacity += r.SubterraneanCapacityPerClass * float64(sizeClass)
	}
	if capacity <= 0 {
		return 0, fmt.Errorf("population capacity for race %q on %s/%s is not positive", raceID, planet.SizeID, planet.ClimateID)
	}
	return capacity, nil
}

func (r *EconomyRules) CalculatePopulationDynamics(colony core.Colony, planet core.Planet, raceID string, adjusted core.ColonyEconomy) (core.ColonyPopulationDynamics, error) {
	capacity, err := r.PopulationCapacity(planet, raceID)
	if err != nil {
		return core.ColonyPopulationDynamics{}, err
	}
	modifiers, ok := r.RaceModifiers[raceID]
	if !ok {
		return core.ColonyPopulationDynamics{}, fmt.Errorf("unknown race %q", raceID)
	}
	population := colony.Population.Total
	foodRequired := population * r.PopulationFoodPerUnit
	productionRequired := 0.0
	if modifiers.Lithovore {
		foodRequired = 0
	} else if modifiers.Cybernetic {
		foodRequired = population * r.CyberneticFoodPerUnit
		productionRequired = population * r.CyberneticProductionPerUnit
	}
	foodDelta := adjusted.Food - foodRequired
	foodSurplus := math.Max(0, foodDelta)
	foodShortage := math.Max(0, -foodDelta)
	productionDelta := adjusted.Production - productionRequired
	productionAvailable := math.Max(0, productionDelta)
	productionShortage := math.Max(0, -productionDelta)

	growthMultiplier := modifiers.PopulationGrowthMultiplier
	baseGrowth := 0.0
	projectedGrowth := 0.0
	if population > 0 && population < capacity && foodShortage <= populationEpsilon && productionShortage <= populationEpsilon {
		baseGrowth = math.Sqrt(r.PopulationGrowthCurveFactor * population * (capacity - population) / capacity)
		projectedGrowth = math.Min(capacity-population, baseGrowth*growthMultiplier)
	}
	return core.ColonyPopulationDynamics{
		Capacity:            capacity,
		FoodRequired:        foodRequired,
		FoodSurplus:         foodSurplus,
		FoodShortage:        foodShortage,
		ProductionRequired:  productionRequired,
		ProductionShortage:  productionShortage,
		ProductionAvailable: productionAvailable,
		BaseGrowth:          baseGrowth,
		GrowthMultiplier:    growthMultiplier,
		ProjectedGrowth:     projectedGrowth,
	}, nil
}

type PopulationGrewEvent struct {
	ColonyID      core.ID                       `json:"colony_id"`
	Previous      core.PopulationState          `json:"previous"`
	Current       core.PopulationState          `json:"current"`
	AppliedGrowth float64                       `json:"applied_growth"`
	Dynamics      core.ColonyPopulationDynamics `json:"dynamics"`
}

func (r *EconomyResolver) advancePopulation(state *core.GameState) ([]DomainEvent, error) {
	var events []DomainEvent
	for i := range state.Colonies {
		colony := &state.Colonies[i]
		growth := colony.PopulationDynamics.ProjectedGrowth
		if growth <= populationEpsilon || colony.Population.Total <= populationEpsilon {
			continue
		}
		previous := colony.Population
		nextTotal := math.Min(colony.PopulationDynamics.Capacity, previous.Total+growth)
		applied := nextTotal - previous.Total
		if applied <= populationEpsilon {
			continue
		}
		ratio := nextTotal / previous.Total
		colony.Population.Total = nextTotal
		colony.Population.Farmers = previous.Farmers * ratio
		colony.Population.Workers = previous.Workers * ratio
		colony.Population.Scientists = nextTotal - colony.Population.Farmers - colony.Population.Workers
		if colony.Population.Scientists < 0 && colony.Population.Scientists > -populationEpsilon {
			colony.Population.Scientists = 0
		}
		event, err := NewDomainEvent("colony.population_grew", 0, 0, PopulationGrewEvent{
			ColonyID:      colony.ID,
			Previous:      previous,
			Current:       colony.Population,
			AppliedGrowth: applied,
			Dynamics:      colony.PopulationDynamics,
		})
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}
