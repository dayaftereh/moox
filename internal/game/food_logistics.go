package game

import (
	"fmt"
	"math"
	"sort"

	"moox/internal/core"
)

type ColonyFoodLogistics struct {
	ColonyID     core.ID `json:"colony_id"`
	FoodImported float64 `json:"food_imported"`
	FoodExported float64 `json:"food_exported"`
	FoodSurplus  float64 `json:"food_surplus"`
	FoodShortage float64 `json:"food_shortage"`
}

type FoodLogisticsResolvedEvent struct {
	EmpireID core.ID                  `json:"empire_id"`
	Snapshot core.EmpireFoodLogistics `json:"snapshot"`
	Colonies []ColonyFoodLogistics    `json:"colonies"`
}

func (r *EconomyRules) refreshPopulationProjection(dynamics *core.ColonyPopulationDynamics, population float64, raceID string) error {
	if dynamics == nil {
		return fmt.Errorf("population dynamics must not be nil")
	}
	modifiers, ok := r.RaceModifiers[raceID]
	if !ok {
		return fmt.Errorf("unknown race %q", raceID)
	}
	baseGrowth := 0.0
	if population > 0 && population < dynamics.Capacity {
		baseGrowth = math.Sqrt(r.PopulationGrowthCurveFactor * population * (dynamics.Capacity - population) / dynamics.Capacity)
	}
	naturalGrowth := baseGrowth * modifiers.PopulationGrowthMultiplier
	starvationPenalty := dynamics.FoodShortage * r.StarvationPopulationPerFoodShortage
	if modifiers.Cybernetic {
		starvationPenalty = dynamics.FoodShortage*r.CyberneticStarvationPopulationPerFoodShortage + dynamics.ProductionShortage*r.CyberneticStarvationPopulationPerPPShortage
	}
	net := naturalGrowth - starvationPenalty
	dynamics.BaseGrowth = baseGrowth
	dynamics.GrowthMultiplier = modifiers.PopulationGrowthMultiplier
	dynamics.ProjectedGrowth = 0
	dynamics.ProjectedStarvation = 0
	if net > populationEpsilon {
		dynamics.ProjectedGrowth = math.Min(dynamics.Capacity-population, net)
	} else if net < -populationEpsilon {
		maxLoss := math.Max(0, population-r.MinimumPopulationAfterStarvation)
		dynamics.ProjectedStarvation = math.Min(maxLoss, -net)
	}
	return nil
}

func (r *EconomyResolver) materializeFoodLogistics(state *core.GameState, emit bool) ([]DomainEvent, error) {
	if r == nil || r.Rules == nil {
		return nil, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}

	empireIndexes := make([]int, len(state.Empires))
	for i := range state.Empires {
		empireIndexes[i] = i
	}
	sort.Slice(empireIndexes, func(i, j int) bool { return state.Empires[empireIndexes[i]].ID < state.Empires[empireIndexes[j]].ID })

	var events []DomainEvent
	for _, empireIndex := range empireIndexes {
		empire := &state.Empires[empireIndex]
		modifiers, ok := r.Rules.RaceModifiers[empire.RaceID]
		if !ok {
			return nil, fmt.Errorf("empire %d has unknown race %q", empire.ID, empire.RaceID)
		}
		colonies := coloniesForEmpire(state, empire.ID)
		totalSurplus := 0.0
		totalShortage := 0.0
		for _, colony := range colonies {
			d := &colony.PopulationDynamics
			d.FoodImported = 0
			d.FoodExported = 0
			d.FoodSurplus = d.LocalFoodSurplus
			d.FoodShortage = d.LocalFoodShortage
			totalSurplus += d.LocalFoodSurplus
			totalShortage += d.LocalFoodShortage
		}

		transferPossible := math.Min(totalSurplus, totalShortage)
		capacity := float64(empire.Freighters) * r.Rules.FreighterFoodCapacity
		transfer := math.Min(transferPossible, capacity)
		if transfer > populationEpsilon {
			allocateFoodImports(colonies, transfer, totalShortage)
			allocateFoodExports(colonies, transfer, totalSurplus)
		}

		used := 0
		if transfer > populationEpsilon {
			used = int(math.Ceil(transfer/r.Rules.FreighterFoodCapacity - populationEpsilon))
		}
		required := 0
		if transferPossible > populationEpsilon {
			required = int(math.Ceil(transferPossible/r.Rules.FreighterFoodCapacity - populationEpsilon))
		}
		remainingSurplus := 0.0
		remainingShortage := 0.0
		colonyViews := make([]ColonyFoodLogistics, 0, len(colonies))
		for _, colony := range colonies {
			d := &colony.PopulationDynamics
			d.FoodSurplus = math.Max(0, d.LocalFoodSurplus-d.FoodExported)
			d.FoodShortage = math.Max(0, d.LocalFoodShortage-d.FoodImported)
			if err := r.Rules.refreshPopulationProjection(d, colony.Population.Total, empire.RaceID); err != nil {
				return nil, fmt.Errorf("colony %d population projection: %w", colony.ID, err)
			}
			remainingSurplus += d.FoodSurplus
			remainingShortage += d.FoodShortage
			colonyViews = append(colonyViews, ColonyFoodLogistics{
				ColonyID: colony.ID, FoodImported: d.FoodImported, FoodExported: d.FoodExported,
				FoodSurplus: d.FoodSurplus, FoodShortage: d.FoodShortage,
			})
		}
		saleRate := r.Rules.SurplusFoodBCPerUnit
		if modifiers.FantasticTraders {
			saleRate = r.Rules.FantasticTradersSurplusFoodBCPerUnit
		}
		empire.FoodLogistics = core.EmpireFoodLogistics{
			FreightersRequired:       required,
			FreightersUsed:           used,
			LocalFoodSurplus:         totalSurplus,
			LocalFoodShortage:        totalShortage,
			FoodTransferred:          transfer,
			FoodUnmet:                remainingShortage,
			SurplusFoodSold:          remainingSurplus,
			FreighterOperatingCostBC: float64(used) * r.Rules.FreighterOperatingCostBC,
			SurplusFoodIncomeBC:      remainingSurplus * saleRate,
		}
		if emit && (totalSurplus > populationEpsilon || totalShortage > populationEpsilon || transfer > populationEpsilon) {
			event, err := NewDomainEvent("empire.food_logistics_resolved", 0, 0, FoodLogisticsResolvedEvent{EmpireID: empire.ID, Snapshot: empire.FoodLogistics, Colonies: colonyViews})
			if err != nil {
				return nil, err
			}
			events = append(events, event)
		}
	}
	return events, nil
}

func coloniesForEmpire(state *core.GameState, empireID core.ID) []*core.Colony {
	colonies := make([]*core.Colony, 0)
	for i := range state.Colonies {
		if state.Colonies[i].EmpireID == empireID {
			colonies = append(colonies, &state.Colonies[i])
		}
	}
	sort.Slice(colonies, func(i, j int) bool { return colonies[i].ID < colonies[j].ID })
	return colonies
}

// The exact classic priority when Freighters are insufficient is not yet
// evidenced. MOOX therefore uses deterministic proportional sharing across all
// deficits/surpluses. The policy is isolated here so original ordering can
// replace it without changing state or protocol shapes.
func allocateFoodImports(colonies []*core.Colony, transfer, totalShortage float64) {
	remaining := transfer
	remainingWeight := totalShortage
	for _, colony := range colonies {
		shortage := colony.PopulationDynamics.LocalFoodShortage
		if shortage <= populationEpsilon || remaining <= populationEpsilon {
			continue
		}
		share := remaining
		if remainingWeight > populationEpsilon {
			share = remaining * shortage / remainingWeight
		}
		share = math.Min(shortage, share)
		colony.PopulationDynamics.FoodImported = share
		remaining -= share
		remainingWeight -= shortage
	}
}

func allocateFoodExports(colonies []*core.Colony, transfer, totalSurplus float64) {
	remaining := transfer
	remainingWeight := totalSurplus
	for _, colony := range colonies {
		surplus := colony.PopulationDynamics.LocalFoodSurplus
		if surplus <= populationEpsilon || remaining <= populationEpsilon {
			continue
		}
		share := remaining
		if remainingWeight > populationEpsilon {
			share = remaining * surplus / remainingWeight
		}
		share = math.Min(surplus, share)
		colony.PopulationDynamics.FoodExported = share
		remaining -= share
		remainingWeight -= surplus
	}
}
