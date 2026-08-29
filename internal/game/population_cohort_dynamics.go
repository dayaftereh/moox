package game

import (
	"fmt"
	"math"
	"sort"

	"moox/internal/core"
)

func (r *EconomyRules) ColonyPopulationCapacityForOrigin(colony core.Colony, planet core.Planet, owner core.Empire, originRaceID string) (float64, error) {
	capacity, err := r.PopulationCapacity(planet, originRaceID)
	if err != nil {
		return 0, err
	}
	if empireKnowsTechnology(&owner, r.AdvancedCityPlanningTechnologyID) {
		capacity += r.AdvancedCityPlanningCapacityBonus
	}
	if colonyHasBuilding(colony, r.BiospheresBuildingID) {
		capacity += r.BiospheresCapacityBonus
	}
	return capacity, nil
}

func (r *EconomyResolver) calculateRaceAwarePopulationDynamics(state *core.GameState, colony core.Colony, planet core.Planet, owner core.Empire, adjusted core.ColonyEconomy) (core.ColonyPopulationDynamics, error) {
	if r == nil || r.Rules == nil {
		return core.ColonyPopulationDynamics{}, fmt.Errorf("economy resolver has no rules")
	}
	totalPopulation := colony.Population.Total()
	originTotals := colony.Population.TotalsByOrigin()
	originIDs := make([]core.ID, 0, len(originTotals))
	for originID := range originTotals {
		originIDs = append(originIDs, originID)
	}
	sort.Slice(originIDs, func(i, j int) bool { return originIDs[i] < originIDs[j] })

	originFood2 := make(map[core.ID]float64, len(originIDs))
	originProductionRequired := make(map[core.ID]float64, len(originIDs))
	dynamics := core.ColonyPopulationDynamics{}
	for _, cohort := range colony.Population.Cohorts {
		origin := empireByID(state, cohort.OriginEmpireID)
		if origin == nil {
			return core.ColonyPopulationDynamics{}, fmt.Errorf("population origin empire %d not found", cohort.OriginEmpireID)
		}
		mods, ok := r.Rules.RaceModifiers[origin.RaceID]
		if !ok {
			return core.ColonyPopulationDynamics{}, fmt.Errorf("unknown origin race %q", origin.RaceID)
		}
		foodPerPopulation := r.Rules.PopulationFoodPerUnit
		if mods.Lithovore {
			foodPerPopulation = 0
		} else if mods.Cybernetic {
			foodPerPopulation = r.Rules.CyberneticFoodPerUnit
			originProductionRequired[cohort.OriginEmpireID] += cohort.Total() * r.Rules.CyberneticProductionPerUnit
		}
		food2 := cohort.Total() * foodPerPopulation * 2
		originFood2[cohort.OriginEmpireID] += food2
		switch {
		case cohort.OriginEmpireID == owner.ID:
			dynamics.OwnerOriginFood2 += food2
		case cohort.AssimilationState == core.PopulationAssimilated:
			dynamics.AssimilatedForeignFood2 += food2
		case cohort.AssimilationState == core.PopulationConquered:
			dynamics.ConqueredForeignFood2 += food2
		}
	}

	totalFood2 := dynamics.OwnerOriginFood2 + dynamics.AssimilatedForeignFood2 + dynamics.ConqueredForeignFood2 + dynamics.RemainingFood2
	dynamics.WholeFoodRequired = math.Ceil(totalFood2/2 - populationEpsilon)
	if dynamics.WholeFoodRequired < 0 {
		dynamics.WholeFoodRequired = 0
	}
	dynamics.FoodRequired = dynamics.WholeFoodRequired
	for _, required := range originProductionRequired {
		dynamics.ProductionRequired += required
	}
	foodDelta := adjusted.Food - dynamics.FoodRequired
	productionDelta := adjusted.Production - dynamics.ProductionRequired
	dynamics.LocalFoodSurplus = math.Max(0, foodDelta)
	dynamics.LocalFoodShortage = math.Max(0, -foodDelta)
	dynamics.FoodSurplus = dynamics.LocalFoodSurplus
	dynamics.FoodShortage = dynamics.LocalFoodShortage
	dynamics.ProductionShortage = math.Max(0, -productionDelta)
	dynamics.ProductionAvailable = math.Max(0, productionDelta)

	technologyGrowthBonus := 0.0
	for _, technologyID := range owner.KnownTechnologyIDs {
		technologyGrowthBonus += r.Rules.PopulationGrowthTechnologyBonusByID[technologyID]
	}
	housingGrowthBonus := 0.0
	if colony.Construction != nil && colony.Construction.ProjectKind == core.ConstructionProjectHousing {
		if colony.Construction.ProjectID != HousingProjectID {
			return core.ColonyPopulationDynamics{}, fmt.Errorf("colony %d has invalid Housing project id %q", colony.ID, colony.Construction.ProjectID)
		}
		if totalPopulation > populationEpsilon && dynamics.ProductionAvailable > 0 {
			housingPercent := r.Rules.HousingGrowthPercentPerPPPerPopulation * dynamics.ProductionAvailable / totalPopulation
			if r.Rules.HousingGrowthPercentRounding == "down" {
				housingPercent = math.Floor(housingPercent + populationEpsilon)
			}
			housingGrowthBonus = housingPercent / 100
		}
	}

	eligibleForCloning := 0.0
	originCapacity := make(map[core.ID]float64, len(originIDs))
	for _, originID := range originIDs {
		origin := empireByID(state, originID)
		if origin == nil {
			return core.ColonyPopulationDynamics{}, fmt.Errorf("population origin empire %d not found", originID)
		}
		capacity, err := r.Rules.ColonyPopulationCapacityForOrigin(colony, planet, owner, origin.RaceID)
		if err != nil {
			return core.ColonyPopulationDynamics{}, err
		}
		originCapacity[originID] = capacity
		if capacity > dynamics.Capacity {
			dynamics.Capacity = capacity
		}
		if totalPopulation < capacity-populationEpsilon {
			eligibleForCloning += originTotals[originID]
		}
	}

	for _, originID := range originIDs {
		origin := empireByID(state, originID)
		mods := r.Rules.RaceModifiers[origin.RaceID]
		capacity := originCapacity[originID]
		originPopulation := originTotals[originID]
		baseGrowth := 0.0
		if originPopulation > populationEpsilon && totalPopulation < capacity-populationEpsilon {
			baseGrowth = math.Sqrt(r.Rules.PopulationGrowthCurveFactor * originPopulation * (capacity - totalPopulation) / capacity)
		}
		growthMultiplier := mods.PopulationGrowthMultiplier + technologyGrowthBonus + housingGrowthBonus
		projectedGrowth := baseGrowth * growthMultiplier
		if colonyHasBuilding(colony, r.Rules.CloningCenterBuildingID) && eligibleForCloning > populationEpsilon && totalPopulation < capacity-populationEpsilon {
			projectedGrowth += r.Rules.CloningCenterFlatGrowth * originPopulation / eligibleForCloning
		}
		dynamics.Origins = append(dynamics.Origins, core.PopulationOriginDynamics{
			OriginEmpireID:   originID,
			Capacity:         capacity,
			Population:       originPopulation,
			Food2Required:    originFood2[originID],
			BaseGrowth:       baseGrowth,
			GrowthMultiplier: growthMultiplier,
			ProjectedGrowth:  projectedGrowth,
		})
		dynamics.BaseGrowth += baseGrowth
		dynamics.ProjectedGrowth += projectedGrowth
		if totalPopulation > populationEpsilon {
			dynamics.GrowthMultiplier += growthMultiplier * originPopulation / totalPopulation
		}
	}
	if len(dynamics.Origins) == 1 {
		dynamics.GrowthMultiplier = dynamics.Origins[0].GrowthMultiplier
	}

	if err := r.refreshRaceAwarePopulationProjection(state, colony, owner, adjusted.Food, &dynamics); err != nil {
		return core.ColonyPopulationDynamics{}, err
	}
	return dynamics, nil
}

func (r *EconomyResolver) refreshRaceAwarePopulationProjection(state *core.GameState, colony core.Colony, owner core.Empire, adjustedFood float64, dynamics *core.ColonyPopulationDynamics) error {
	if dynamics == nil {
		return fmt.Errorf("population dynamics must not be nil")
	}
	unmetFood2, err := r.unmetFood2ByOrigin(state, colony, owner, adjustedFood+dynamics.FoodImported)
	if err != nil {
		return err
	}
	productionRequiredByOrigin := make(map[core.ID]float64)
	totalCyberneticProductionRequired := 0.0
	for _, cohort := range colony.Population.Cohorts {
		origin := empireByID(state, cohort.OriginEmpireID)
		if origin == nil {
			return fmt.Errorf("population origin empire %d not found", cohort.OriginEmpireID)
		}
		mods := r.Rules.RaceModifiers[origin.RaceID]
		if mods.Cybernetic {
			required := cohort.Total() * r.Rules.CyberneticProductionPerUnit
			productionRequiredByOrigin[cohort.OriginEmpireID] += required
			totalCyberneticProductionRequired += required
		}
	}

	totalLoss := 0.0
	hasStarvationPressure := false
	for i := range dynamics.Origins {
		originDynamics := &dynamics.Origins[i]
		origin := empireByID(state, originDynamics.OriginEmpireID)
		if origin == nil {
			return fmt.Errorf("population origin empire %d not found", originDynamics.OriginEmpireID)
		}
		mods := r.Rules.RaceModifiers[origin.RaceID]
		foodShortage := unmetFood2[originDynamics.OriginEmpireID] / 2
		loss := foodShortage * r.Rules.StarvationPopulationPerFoodShortage
		if mods.Cybernetic {
			productionShortage := 0.0
			if totalCyberneticProductionRequired > populationEpsilon && dynamics.ProductionShortage > 0 {
				productionShortage = dynamics.ProductionShortage * productionRequiredByOrigin[originDynamics.OriginEmpireID] / totalCyberneticProductionRequired
			}
			loss = foodShortage*r.Rules.CyberneticStarvationPopulationPerFoodShortage + productionShortage*r.Rules.CyberneticStarvationPopulationPerPPShortage
		}
		if loss > populationEpsilon {
			hasStarvationPressure = true
		}
		originDynamics.ProjectedStarvation = math.Min(originDynamics.Population, math.Max(0, loss))
		totalLoss += originDynamics.ProjectedStarvation
	}
	maxLoss := math.Max(0, colony.Population.Total()-r.Rules.MinimumPopulationAfterStarvation)
	if totalLoss > maxLoss+populationEpsilon && totalLoss > populationEpsilon {
		scale := maxLoss / totalLoss
		totalLoss = 0
		for i := range dynamics.Origins {
			dynamics.Origins[i].ProjectedStarvation *= scale
			totalLoss += dynamics.Origins[i].ProjectedStarvation
		}
	}
	dynamics.ProjectedStarvation = totalLoss
	if hasStarvationPressure {
		dynamics.ProjectedGrowth = 0
		for i := range dynamics.Origins {
			dynamics.Origins[i].ProjectedGrowth = 0
		}
	} else {
		dynamics.ProjectedGrowth = 0
		for i := range dynamics.Origins {
			dynamics.ProjectedGrowth += dynamics.Origins[i].ProjectedGrowth
		}
	}
	return nil
}

func (r *EconomyResolver) unmetFood2ByOrigin(state *core.GameState, colony core.Colony, owner core.Empire, foodAvailable float64) (map[core.ID]float64, error) {
	tiers := []map[core.ID]float64{{}, {}, {}}
	for _, cohort := range colony.Population.Cohorts {
		origin := empireByID(state, cohort.OriginEmpireID)
		if origin == nil {
			return nil, fmt.Errorf("population origin empire %d not found", cohort.OriginEmpireID)
		}
		mods := r.Rules.RaceModifiers[origin.RaceID]
		foodPerPopulation := r.Rules.PopulationFoodPerUnit
		if mods.Lithovore {
			foodPerPopulation = 0
		} else if mods.Cybernetic {
			foodPerPopulation = r.Rules.CyberneticFoodPerUnit
		}
		food2 := cohort.Total() * foodPerPopulation * 2
		tier := 0
		if cohort.OriginEmpireID != owner.ID {
			if cohort.AssimilationState == core.PopulationConquered {
				tier = 2
			} else {
				tier = 1
			}
		}
		tiers[tier][cohort.OriginEmpireID] += food2
	}

	unmet := make(map[core.ID]float64)
	availableFood2 := math.Max(0, foodAvailable*2)
	for _, tier := range tiers {
		tierTotal := 0.0
		for _, food2 := range tier {
			tierTotal += food2
		}
		if tierTotal <= populationEpsilon {
			continue
		}
		if availableFood2 >= tierTotal-populationEpsilon {
			availableFood2 -= tierTotal
			continue
		}
		missing := tierTotal - availableFood2
		availableFood2 = 0
		for originID, food2 := range tier {
			unmet[originID] += missing * food2 / tierTotal
		}
	}
	return unmet, nil
}
