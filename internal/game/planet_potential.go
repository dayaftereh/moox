package game

import (
	"fmt"
	"math"

	"moox/internal/core"
)

// PlanetPotential is a player-relative, colony-independent projection of the
// economy a population from the viewing Empire can expect from one job unit on
// a planet. It deliberately excludes Colony-local morale/building effects.
type PlanetPotential struct {
	PlanetID                   core.ID `json:"planet_id"`
	FoodPerFarmer              float64 `json:"food_per_farmer"`
	ProductionPerWorker        float64 `json:"production_per_worker"`
	ResearchPerScientist       float64 `json:"research_per_scientist"`
	GravityPenaltyPercent      int     `json:"gravity_penalty_percent"`
	ClimateHabitabilityPercent int     `json:"climate_habitability_percent"`
	SizeBaseCapacity           float64 `json:"size_base_capacity"`
	PopulationCapacity         float64 `json:"population_capacity"`
}

// PlanetPotentialForEmpire derives player-facing planet potential from the same
// authoritative rules used by Colony economy/capacity resolution. Government,
// morale and Colony-local buildings are intentionally excluded because an
// uncolonized planet does not have that context yet.
func (r *EconomyRules) PlanetPotentialForEmpire(planet core.Planet, empire core.Empire) (PlanetPotential, error) {
	if r == nil {
		return PlanetPotential{}, fmt.Errorf("economy rules must not be nil")
	}
	modifiers, ok := r.RaceModifiers[empire.RaceID]
	if !ok {
		return PlanetPotential{}, fmt.Errorf("unknown race %q", empire.RaceID)
	}

	foodBase, ok := r.ClimateFoodPerFarmer[planet.ClimateID]
	if !ok {
		return PlanetPotential{}, fmt.Errorf("unknown climate %q", planet.ClimateID)
	}
	if modifiers.Aquatic {
		if _, aquatic := r.AquaticFoodClimateIDs[planet.ClimateID]; aquatic {
			foodBase += r.AquaticFoodBonus
		}
	}
	foodPerFarmer := 0.0
	if foodBase > 0 {
		foodPerFarmer = math.Max(1, foodBase+modifiers.FoodPerFarmer)
	}

	industryBase, ok := r.MineralIndustryPerWorker[planet.MineralID]
	if !ok {
		return PlanetPotential{}, fmt.Errorf("unknown mineral class %q", planet.MineralID)
	}
	productionPerWorker := math.Max(1, industryBase+modifiers.ProductionPerWorker)
	researchPerScientist := math.Max(1, r.BaseResearchPerScientist+modifiers.ResearchPerScientist)

	gravityPenalty, ok := r.GravityPenaltyPercent[gravityKey(modifiers.GravityID, planet.GravityID)]
	if !ok {
		return PlanetPotential{}, fmt.Errorf("no gravity rule for race %q on planet gravity %q", modifiers.GravityID, planet.GravityID)
	}
	climateHabitability, ok := r.PopulationClimateHabitability[planet.ClimateID]
	if !ok {
		return PlanetPotential{}, fmt.Errorf("unknown population climate %q", planet.ClimateID)
	}
	sizeBaseCapacity, ok := r.PopulationSizeCapacity[planet.SizeID]
	if !ok {
		return PlanetPotential{}, fmt.Errorf("unknown population size %q", planet.SizeID)
	}

	capacity, err := r.PopulationCapacityForEmpire(planet, empire)
	if err != nil {
		return PlanetPotential{}, err
	}

	return PlanetPotential{
		PlanetID:                   planet.ID,
		FoodPerFarmer:              foodPerFarmer,
		ProductionPerWorker:        productionPerWorker,
		ResearchPerScientist:       researchPerScientist,
		GravityPenaltyPercent:      gravityPenalty,
		ClimateHabitabilityPercent: int(math.Round(climateHabitability * 100)),
		SizeBaseCapacity:           sizeBaseCapacity,
		PopulationCapacity:         capacity,
	}, nil
}
