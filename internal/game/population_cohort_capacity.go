package game

import (
	"fmt"
	"math"
	"sort"

	"moox/internal/core"
)

func (r *EconomyResolver) populationCapacityByOrigin(state *core.GameState, colony core.Colony, planet core.Planet, owner core.Empire, population core.PopulationState) (map[core.ID]float64, error) {
	capacities := make(map[core.ID]float64)
	for originID := range population.TotalsByOrigin() {
		origin := empireByID(state, originID)
		if origin == nil {
			return nil, fmt.Errorf("population origin empire %d not found", originID)
		}
		capacity, err := r.Rules.ColonyPopulationCapacityForOrigin(colony, planet, owner, origin.RaceID)
		if err != nil {
			return nil, err
		}
		capacities[originID] = capacity
	}
	return capacities, nil
}

func populationFitsHeterogeneousCapacity(population core.PopulationState, capacities map[core.ID]float64) bool {
	if population.Total() <= populationEpsilon {
		return true
	}
	thresholds := distinctCapacityThresholds(capacities)
	originTotals := population.TotalsByOrigin()
	for _, threshold := range thresholds {
		constrained := 0.0
		for originID, amount := range originTotals {
			capacity, ok := capacities[originID]
			if !ok {
				return false
			}
			if capacity <= threshold+populationEpsilon {
				constrained += amount
			}
		}
		if constrained > threshold+populationEpsilon {
			return false
		}
	}
	return true
}

func distinctCapacityThresholds(capacities map[core.ID]float64) []float64 {
	thresholds := make([]float64, 0, len(capacities))
	for _, capacity := range capacities {
		found := false
		for _, current := range thresholds {
			if math.Abs(current-capacity) <= populationEpsilon {
				found = true
				break
			}
		}
		if !found {
			thresholds = append(thresholds, capacity)
		}
	}
	sort.Float64s(thresholds)
	return thresholds
}

func trimPopulationToHeterogeneousCapacity(population *core.PopulationState, ownerEmpireID core.ID, capacities map[core.ID]float64) float64 {
	if population == nil || population.Total() <= populationEpsilon {
		return 0
	}
	before := population.Total()
	for _, threshold := range distinctCapacityThresholds(capacities) {
		for {
			constrained := constrainedPopulationAtThreshold(*population, capacities, threshold)
			excess := constrained - threshold
			if excess <= populationEpsilon {
				break
			}
			removed := 0.0
			for priority := 0; priority < 3 && excess > populationEpsilon; priority++ {
				indexes := make([]int, 0)
				classTotal := 0.0
				for i, cohort := range population.Cohorts {
					if capacities[cohort.OriginEmpireID] > threshold+populationEpsilon {
						continue
					}
					if populationTrimPriority(cohort, ownerEmpireID) != priority {
						continue
					}
					indexes = append(indexes, i)
					classTotal += cohort.Total()
				}
				if classTotal <= populationEpsilon {
					continue
				}
				toRemove := math.Min(excess, classTotal)
				factor := math.Max(0, (classTotal-toRemove)/classTotal)
				for _, index := range indexes {
					population.Cohorts[index].Farmers *= factor
					population.Cohorts[index].Workers *= factor
					population.Cohorts[index].Scientists *= factor
				}
				excess -= toRemove
				removed += toRemove
			}
			population.Normalize()
			if removed <= populationEpsilon {
				break
			}
		}
	}
	return math.Max(0, before-population.Total())
}

func constrainedPopulationAtThreshold(population core.PopulationState, capacities map[core.ID]float64, threshold float64) float64 {
	total := 0.0
	for _, cohort := range population.Cohorts {
		if capacities[cohort.OriginEmpireID] <= threshold+populationEpsilon {
			total += cohort.Total()
		}
	}
	return total
}

func populationTrimPriority(cohort core.PopulationCohort, ownerEmpireID core.ID) int {
	if cohort.OriginEmpireID != ownerEmpireID && cohort.AssimilationState == core.PopulationConquered {
		return 0
	}
	if cohort.OriginEmpireID != ownerEmpireID {
		return 1
	}
	return 2
}

func clonePopulationState(population core.PopulationState) core.PopulationState {
	clone := population
	clone.Cohorts = append([]core.PopulationCohort(nil), population.Cohorts...)
	return clone
}

func (r *EconomyResolver) populationCanAdd(state *core.GameState, colony core.Colony, population core.PopulationState, key core.PopulationCohortKey, job core.PopulationJob, amount float64) (bool, error) {
	planet := planetByID(state, colony.PlanetID)
	if planet == nil {
		return false, fmt.Errorf("colony %d references unknown planet %d", colony.ID, colony.PlanetID)
	}
	owner := empireByID(state, colony.EmpireID)
	if owner == nil {
		return false, fmt.Errorf("colony %d references unknown empire %d", colony.ID, colony.EmpireID)
	}
	candidate := clonePopulationState(population)
	if err := candidate.AddToCohortJob(key, job, amount); err != nil {
		return false, err
	}
	capacities, err := r.populationCapacityByOrigin(state, colony, *planet, *owner, candidate)
	if err != nil {
		return false, err
	}
	return populationFitsHeterogeneousCapacity(candidate, capacities), nil
}

func (r *EconomyResolver) trimColonyToHeterogeneousCapacity(state *core.GameState, colony *core.Colony) (float64, error) {
	if colony == nil {
		return 0, fmt.Errorf("colony must not be nil")
	}
	planet := planetByID(state, colony.PlanetID)
	owner := empireByID(state, colony.EmpireID)
	if planet == nil || owner == nil {
		return 0, fmt.Errorf("colony %d has invalid planet/empire references", colony.ID)
	}
	capacities, err := r.populationCapacityByOrigin(state, *colony, *planet, *owner, colony.Population)
	if err != nil {
		return 0, err
	}
	return trimPopulationToHeterogeneousCapacity(&colony.Population, owner.ID, capacities), nil
}
