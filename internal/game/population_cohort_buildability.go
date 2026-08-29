package game

import (
	"fmt"

	"moox/internal/core"
)

func (r *EconomyRules) mixedPopulationHasGrowthRoom(state *core.GameState, colony core.Colony, planet core.Planet, owner core.Empire) (bool, error) {
	originTotals := colony.Population.TotalsByOrigin()
	if len(originTotals) == 0 {
		return true, nil
	}
	capacities := make(map[core.ID]float64, len(originTotals))
	for originID := range originTotals {
		origin := empireByID(state, originID)
		if origin == nil {
			return false, fmt.Errorf("population origin empire %d not found", originID)
		}
		capacity, err := r.ColonyPopulationCapacityForOrigin(colony, planet, owner, origin.RaceID)
		if err != nil {
			return false, err
		}
		capacities[originID] = capacity
	}
	for originID := range originTotals {
		candidate := clonePopulationState(colony.Population)
		key := core.PopulationCohortKey{OriginEmpireID: originID, LoyaltyEmpireID: owner.ID, AssimilationState: core.PopulationAssimilated}
		if err := candidate.AddToCohortJob(key, core.PopulationJobFarmer, populationEpsilon*100); err != nil {
			return false, err
		}
		if populationFitsHeterogeneousCapacity(candidate, capacities) {
			return true, nil
		}
	}
	return false, nil
}
