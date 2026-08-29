package game

import (
	"fmt"
	"math"
	"sort"

	"moox/internal/core"
)

type ColonyFoodLogistics struct {
	ColonyID     core.ID `json:"colony_id"`
	Blockaded    bool    `json:"blockaded"`
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

func (r *EconomyRules) refreshPopulationProjection(dynamics *core.ColonyPopulationDynamics, colony core.Colony, empire core.Empire) error {
	if dynamics == nil {
		return fmt.Errorf("population dynamics must not be nil")
	}
	population := colony.Population.Total
	modifiers, ok := r.RaceModifiers[empire.RaceID]
	if !ok {
		return fmt.Errorf("unknown race %q", empire.RaceID)
	}
	baseGrowth := 0.0
	if population > 0 && population < dynamics.Capacity {
		baseGrowth = math.Sqrt(r.PopulationGrowthCurveFactor * population * (dynamics.Capacity - population) / dynamics.Capacity)
	}
	growthMultiplier := modifiers.PopulationGrowthMultiplier
	for _, technologyID := range empire.KnownTechnologyIDs {
		if bonus, ok := r.PopulationGrowthTechnologyBonusByID[technologyID]; ok {
			growthMultiplier += bonus
		}
	}
	if colony.Construction != nil && colony.Construction.ProjectKind == core.ConstructionProjectHousing {
		if colony.Construction.ProjectID != HousingProjectID {
			return fmt.Errorf("colony %d has invalid Housing project id %q", colony.ID, colony.Construction.ProjectID)
		}
		if population > populationEpsilon && population < dynamics.Capacity && dynamics.ProductionAvailable > 0 {
			housingPercent := r.HousingGrowthPercentPerPPPerPopulation * dynamics.ProductionAvailable / population
			if r.HousingGrowthPercentRounding == "down" {
				housingPercent = math.Floor(housingPercent + populationEpsilon)
			}
			growthMultiplier += housingPercent / 100
		}
	}
	flatGrowth := 0.0
	if population > 0 && population < dynamics.Capacity && colonyHasBuilding(colony, r.CloningCenterBuildingID) {
		flatGrowth = r.CloningCenterFlatGrowth
	}
	naturalGrowth := baseGrowth * growthMultiplier
	starvationPenalty := dynamics.FoodShortage * r.StarvationPopulationPerFoodShortage
	if modifiers.Cybernetic {
		starvationPenalty = dynamics.FoodShortage*r.CyberneticStarvationPopulationPerFoodShortage + dynamics.ProductionShortage*r.CyberneticStarvationPopulationPerPPShortage
	}
	net := naturalGrowth + flatGrowth - starvationPenalty
	dynamics.BaseGrowth = baseGrowth
	dynamics.GrowthMultiplier = growthMultiplier
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

func colonyHasBuilding(colony core.Colony, buildingID string) bool {
	for _, owned := range colony.Buildings {
		if owned == buildingID {
			return true
		}
	}
	return false
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

	systemByPlanetID := make(map[core.ID]*core.StarSystem)
	for si := range state.Galaxy.Systems {
		system := &state.Galaxy.Systems[si]
		for pi := range system.Planets {
			systemByPlanetID[system.Planets[pi].ID] = system
		}
	}

	var events []DomainEvent
	for _, empireIndex := range empireIndexes {
		empire := &state.Empires[empireIndex]
		modifiers, ok := r.Rules.RaceModifiers[empire.RaceID]
		if !ok {
			return nil, fmt.Errorf("empire %d has unknown race %q", empire.ID, empire.RaceID)
		}
		colonies := coloniesForEmpire(state, empire.ID)
		eligibleColonies := make([]*core.Colony, 0, len(colonies))
		blockadedByColonyID := make(map[core.ID]bool, len(colonies))
		totalSurplus := 0.0
		totalShortage := 0.0
		eligibleSurplus := 0.0
		eligibleShortage := 0.0
		blockedSurplus := 0.0
		blockedShortage := 0.0
		for _, colony := range colonies {
			d := &colony.PopulationDynamics
			d.FoodImported = 0
			d.FoodExported = 0
			d.FoodSurplus = d.LocalFoodSurplus
			d.FoodShortage = d.LocalFoodShortage
			totalSurplus += d.LocalFoodSurplus
			totalShortage += d.LocalFoodShortage

			system, ok := systemByPlanetID[colony.PlanetID]
			if !ok {
				return nil, fmt.Errorf("colony %d planet %d is not attached to a star system", colony.ID, colony.PlanetID)
			}
			blockaded := systemBlockadesEmpire(system, empire.ID)
			blockadedByColonyID[colony.ID] = blockaded
			if blockaded {
				blockedSurplus += d.LocalFoodSurplus
				blockedShortage += d.LocalFoodShortage
				continue
			}
			eligibleColonies = append(eligibleColonies, colony)
			eligibleSurplus += d.LocalFoodSurplus
			eligibleShortage += d.LocalFoodShortage
		}

		transferPossible := math.Min(eligibleSurplus, eligibleShortage)
		populationReserved := populationTransferFreightersReserved(state, empire.ID)
		availableForFood := empire.Freighters - populationReserved
		if availableForFood < 0 {
			availableForFood = 0
		}
		capacity := float64(availableForFood) * r.Rules.FreighterFoodCapacity
		transfer := math.Min(transferPossible, capacity)
		if transfer > populationEpsilon {
			allocateFoodImports(eligibleColonies, transfer, r.Rules.FreighterFoodCapacity)
			allocateFoodExports(eligibleColonies, transfer, eligibleSurplus)
		}

		used := 0
		if transfer > populationEpsilon {
			used = int(math.Ceil(transfer/r.Rules.FreighterFoodCapacity - populationEpsilon))
		}
		required := 0
		if transferPossible > populationEpsilon {
			required = int(math.Ceil(transferPossible/r.Rules.FreighterFoodCapacity - populationEpsilon))
		}
		sellableSurplus := 0.0
		remainingShortage := 0.0
		colonyViews := make([]ColonyFoodLogistics, 0, len(colonies))
		for _, colony := range colonies {
			d := &colony.PopulationDynamics
			d.FoodSurplus = math.Max(0, d.LocalFoodSurplus-d.FoodExported)
			d.FoodShortage = math.Max(0, d.LocalFoodShortage-d.FoodImported)
			if err := r.Rules.refreshPopulationProjection(d, *colony, *empire); err != nil {
				return nil, fmt.Errorf("colony %d population projection: %w", colony.ID, err)
			}
			if !blockadedByColonyID[colony.ID] {
				sellableSurplus += d.FoodSurplus
			}
			remainingShortage += d.FoodShortage
			colonyViews = append(colonyViews, ColonyFoodLogistics{
				ColonyID: colony.ID, Blockaded: blockadedByColonyID[colony.ID], FoodImported: d.FoodImported, FoodExported: d.FoodExported,
				FoodSurplus: d.FoodSurplus, FoodShortage: d.FoodShortage,
			})
		}
		saleRate := r.Rules.SurplusFoodBCPerUnit
		if modifiers.FantasticTraders {
			saleRate = r.Rules.FantasticTradersSurplusFoodBCPerUnit
		}
		empire.FoodLogistics = core.EmpireFoodLogistics{
			FreightersRequired:                    required,
			FreightersUsed:                        used,
			PopulationTransportFreightersReserved: populationReserved,
			FreightersAvailableForFood:            availableForFood,
			LocalFoodSurplus:                      totalSurplus,
			LocalFoodShortage:                     totalShortage,
			BlockedFoodSurplus:                    blockedSurplus,
			BlockedFoodShortage:                   blockedShortage,
			FoodTransferred:                       transfer,
			FoodUnmet:                             remainingShortage,
			SurplusFoodSold:                       sellableSurplus,
			FreighterOperatingCostBC:              float64(used) * r.Rules.FreighterOperatingCostBC,
			SurplusFoodIncomeBC:                   sellableSurplus * saleRate,
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

func systemBlockadesEmpire(system *core.StarSystem, empireID core.ID) bool {
	if system == nil || empireID == 0 {
		return false
	}
	for _, blockadedEmpireID := range system.BlockadedEmpireIDs {
		if blockadedEmpireID == empireID {
			return true
		}
	}
	return false
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

// Original MOO2 1.31 Pass_Out_Imports_ builds its deficit-colony list in
// colony-array order and, when Food/Freighters are insufficient, repeatedly
// gives each eligible colony one Freighter-load of Food before wrapping to the
// first colony. The original has additional passes for mixed population
// cohorts. MOOX currently models one aggregate race cohort per Colony, so those
// later cohort passes collapse to this single round-robin threshold. The caller
// supplies Colonies in stable simulation-ID order as MOOX's deterministic
// analogue of the original colony-array order.
func allocateFoodImports(colonies []*core.Colony, transfer, foodPerFreighter float64) {
	remaining := transfer
	for remaining > populationEpsilon {
		allocated := false
		for _, colony := range colonies {
			if remaining <= populationEpsilon {
				break
			}
			dynamics := &colony.PopulationDynamics
			shortage := dynamics.LocalFoodShortage - dynamics.FoodImported
			if shortage <= populationEpsilon {
				continue
			}
			share := math.Min(shortage, foodPerFreighter)
			share = math.Min(share, remaining)
			if share <= populationEpsilon {
				continue
			}
			dynamics.FoodImported += share
			remaining -= share
			allocated = true
		}
		if !allocated {
			break
		}
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
