package game

import (
	"fmt"
	"math"

	"moox/internal/core"
)

const populationEpsilon = 1e-9

// PopulationCapacity returns the race/planet population cap before Empire-wide
// Technologies and Colony-local Buildings. It mirrors the original helper layer
// used underneath Planet_Max_Population_For_Player_ and Colony_Race_Pop_Limit_.
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

// PopulationCapacityForEmpire adds Empire-wide capacity effects to the
// race/planet capacity. Original 1.31 applies Advanced City Planning here.
func (r *EconomyRules) PopulationCapacityForEmpire(planet core.Planet, empire core.Empire) (float64, error) {
	capacity, err := r.PopulationCapacity(planet, empire.RaceID)
	if err != nil {
		return 0, err
	}
	if empireKnowsTechnology(&empire, r.AdvancedCityPlanningTechnologyID) {
		capacity += r.AdvancedCityPlanningCapacityBonus
	}
	return capacity, nil
}

// ColonyPopulationCapacity adds Colony-local effects after the Empire-wide
// capacity layer. Original 1.31 adds Biospheres after Advanced City Planning.
func (r *EconomyRules) ColonyPopulationCapacity(colony core.Colony, planet core.Planet, empire core.Empire) (float64, error) {
	capacity, err := r.PopulationCapacityForEmpire(planet, empire)
	if err != nil {
		return 0, err
	}
	if colonyHasBuilding(colony, r.BiospheresBuildingID) {
		capacity += r.BiospheresCapacityBonus
	}
	return capacity, nil
}

// clampAggregatePopulationToCapacity is the current single-cohort fallback for
// an authoritative capacity decrease. Original 1.31 removes Population
// immediately when a capacity source disappears. Until race/job cohorts are
// normalized, MOOX preserves the aggregate role proportions while clamping.
func clampAggregatePopulationToCapacity(colony *core.Colony, capacity float64) float64 {
	if colony == nil || colony.Population.Total() <= capacity+populationEpsilon {
		return 0
	}
	previous := colony.Population.Total()
	if capacity < 0 {
		capacity = 0
	}
	factor := 0.0
	if previous > populationEpsilon {
		factor = capacity / previous
	}
	for i := range colony.Population.Cohorts {
		colony.Population.Cohorts[i].Farmers *= factor
		colony.Population.Cohorts[i].Workers *= factor
		colony.Population.Cohorts[i].Scientists *= factor
	}
	colony.Population.Normalize()
	return previous - capacity
}

func (r *EconomyRules) CalculatePopulationDynamics(colony core.Colony, planet core.Planet, raceID string, adjusted core.ColonyEconomy) (core.ColonyPopulationDynamics, error) {
	return r.calculatePopulationDynamics(colony, planet, core.Empire{RaceID: raceID}, adjusted)
}

func (r *EconomyRules) CalculatePopulationDynamicsForEmpire(colony core.Colony, planet core.Planet, empire core.Empire, adjusted core.ColonyEconomy) (core.ColonyPopulationDynamics, error) {
	return r.calculatePopulationDynamics(colony, planet, empire, adjusted)
}

func (r *EconomyRules) calculatePopulationDynamics(colony core.Colony, planet core.Planet, empire core.Empire, adjusted core.ColonyEconomy) (core.ColonyPopulationDynamics, error) {
	raceID := empire.RaceID
	capacity, err := r.ColonyPopulationCapacity(colony, planet, empire)
	if err != nil {
		return core.ColonyPopulationDynamics{}, err
	}
	modifiers, ok := r.RaceModifiers[raceID]
	if !ok {
		return core.ColonyPopulationDynamics{}, fmt.Errorf("unknown race %q", raceID)
	}
	population := colony.Population.Total()
	foodRequired := population * r.PopulationFoodPerUnit
	productionRequired := 0.0
	if modifiers.Lithovore {
		foodRequired = 0
	} else if modifiers.Cybernetic {
		foodRequired = population * r.CyberneticFoodPerUnit
		productionRequired = population * r.CyberneticProductionPerUnit
	}
	foodDelta := adjusted.Food - foodRequired
	productionDelta := adjusted.Production - productionRequired
	dynamics := core.ColonyPopulationDynamics{
		Capacity:            capacity,
		FoodRequired:        foodRequired,
		LocalFoodSurplus:    math.Max(0, foodDelta),
		LocalFoodShortage:   math.Max(0, -foodDelta),
		FoodSurplus:         math.Max(0, foodDelta),
		FoodShortage:        math.Max(0, -foodDelta),
		ProductionRequired:  productionRequired,
		ProductionShortage:  math.Max(0, -productionDelta),
		ProductionAvailable: math.Max(0, productionDelta),
	}
	if err := r.refreshPopulationProjection(&dynamics, colony, empire); err != nil {
		return core.ColonyPopulationDynamics{}, err
	}
	return dynamics, nil
}

type PopulationGrewEvent struct {
	ColonyID      core.ID                       `json:"colony_id"`
	Previous      core.PopulationState          `json:"previous"`
	Current       core.PopulationState          `json:"current"`
	AppliedGrowth float64                       `json:"applied_growth"`
	Dynamics      core.ColonyPopulationDynamics `json:"dynamics"`
}

type PopulationStarvedEvent struct {
	ColonyID    core.ID                       `json:"colony_id"`
	Previous    core.PopulationState          `json:"previous"`
	Current     core.PopulationState          `json:"current"`
	AppliedLoss float64                       `json:"applied_loss"`
	Dynamics    core.ColonyPopulationDynamics `json:"dynamics"`
}

func (r *EconomyResolver) advancePopulation(state *core.GameState) ([]DomainEvent, error) {
	var events []DomainEvent
	for i := range state.Colonies {
		colony := &state.Colonies[i]
		if colony.Population.Total() <= populationEpsilon {
			continue
		}
		previous := clonePopulationState(colony.Population)
		if loss := colony.PopulationDynamics.ProjectedStarvation; loss > populationEpsilon {
			if len(colony.PopulationDynamics.Origins) == 0 {
				nextTotal := math.Max(r.Rules.MinimumPopulationAfterStarvation, previous.Total()-loss)
				scalePopulation(&colony.Population, previous, nextTotal)
			} else {
				for _, origin := range colony.PopulationDynamics.Origins {
					if err := applyOriginStarvation(&colony.Population, origin.OriginEmpireID, origin.ProjectedStarvation); err != nil {
						return nil, fmt.Errorf("colony %d starvation origin %d: %w", colony.ID, origin.OriginEmpireID, err)
					}
				}
			}
			applied := previous.Total() - colony.Population.Total()
			if applied <= populationEpsilon {
				continue
			}
			event, err := NewDomainEvent("colony.population_starved", 0, 0, PopulationStarvedEvent{
				ColonyID: colony.ID, Previous: previous, Current: colony.Population,
				AppliedLoss: applied, Dynamics: colony.PopulationDynamics,
			})
			if err != nil {
				return nil, err
			}
			events = append(events, event)
			continue
		}
		if growth := colony.PopulationDynamics.ProjectedGrowth; growth > populationEpsilon {
			if len(colony.PopulationDynamics.Origins) == 0 {
				nextTotal := math.Min(colony.PopulationDynamics.Capacity, previous.Total()+growth)
				scalePopulation(&colony.Population, previous, nextTotal)
			} else {
				for _, origin := range colony.PopulationDynamics.Origins {
					if err := applyOriginGrowth(&colony.Population, colony.EmpireID, origin); err != nil {
						return nil, fmt.Errorf("colony %d growth origin %d: %w", colony.ID, origin.OriginEmpireID, err)
					}
				}
				if _, err := r.trimColonyToHeterogeneousCapacity(state, colony); err != nil {
					return nil, err
				}
			}
			applied := colony.Population.Total() - previous.Total()
			if applied <= populationEpsilon {
				continue
			}
			event, err := NewDomainEvent("colony.population_grew", 0, 0, PopulationGrewEvent{
				ColonyID: colony.ID, Previous: previous, Current: colony.Population,
				AppliedGrowth: applied, Dynamics: colony.PopulationDynamics,
			})
			if err != nil {
				return nil, err
			}
			events = append(events, event)
		}
	}
	return events, nil
}
func scalePopulation(target *core.PopulationState, previous core.PopulationState, nextTotal float64) {
	if previous.Total() <= populationEpsilon {
		target.Cohorts = nil
		return
	}
	ratio := nextTotal / previous.Total()
	*target = previous
	target.Cohorts = append([]core.PopulationCohort(nil), previous.Cohorts...)
	for i := range target.Cohorts {
		target.Cohorts[i].Farmers *= ratio
		target.Cohorts[i].Workers *= ratio
		target.Cohorts[i].Scientists *= ratio
	}
	target.Normalize()
}
