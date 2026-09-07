package game

import (
	"fmt"
	"math"

	"moox/internal/core"
)

const metricEpsilon = 1e-9

type PlanningMetricComponent struct {
	ID    string  `json:"id"`
	Value float64 `json:"value"`
}

type PlanningMetricBreakdown struct {
	Total      float64                   `json:"total"`
	Components []PlanningMetricComponent `json:"components,omitempty"`
}

type PlanningColonyBreakdowns struct {
	Growth PlanningMetricBreakdown `json:"growth"`
	TaxBC  PlanningMetricBreakdown `json:"tax_bc"`
}

func appendMetricComponent(components []PlanningMetricComponent, id string, value float64) []PlanningMetricComponent {
	if math.Abs(value) <= metricEpsilon {
		return components
	}
	return append(components, PlanningMetricComponent{ID: id, Value: value})
}

func (r *EconomyResolver) planningColonyBreakdowns(state *core.GameState, colony *core.Colony) (PlanningColonyBreakdowns, error) {
	if state == nil || colony == nil {
		return PlanningColonyBreakdowns{}, fmt.Errorf("state and colony must not be nil")
	}
	owner := empireByID(state, colony.EmpireID)
	if owner == nil {
		return PlanningColonyBreakdowns{}, fmt.Errorf("colony %d owner empire %d not found", colony.ID, colony.EmpireID)
	}
	growth, err := r.planningGrowthBreakdown(state, *colony, *owner)
	if err != nil {
		return PlanningColonyBreakdowns{}, err
	}
	taxBC, err := r.planningTaxBCBreakdown(*colony, *owner)
	if err != nil {
		return PlanningColonyBreakdowns{}, err
	}
	return PlanningColonyBreakdowns{Growth: growth, TaxBC: taxBC}, nil
}

func (r *EconomyResolver) planningGrowthBreakdown(state *core.GameState, colony core.Colony, owner core.Empire) (PlanningMetricBreakdown, error) {
	dynamics := colony.PopulationDynamics
	if dynamics.ProjectedStarvation > metricEpsilon {
		total := -dynamics.ProjectedStarvation
		return PlanningMetricBreakdown{
			Total:      total,
			Components: []PlanningMetricComponent{{ID: "starvation", Value: total}},
		}, nil
	}

	finalGrowth := dynamics.ProjectedGrowth
	if finalGrowth <= metricEpsilon {
		return PlanningMetricBreakdown{Total: finalGrowth}, nil
	}

	technologyBonus := 0.0
	for _, technologyID := range owner.KnownTechnologyIDs {
		technologyBonus += r.Rules.PopulationGrowthTechnologyBonusByID[technologyID]
	}
	housingBonus, err := r.housingPopulationGrowthBonus(colony, dynamics.ProductionAvailable, colony.Population.Total())
	if err != nil {
		return PlanningMetricBreakdown{}, err
	}

	eligibleForCloning := 0.0
	totalPopulation := colony.Population.Total()
	for _, originDynamics := range dynamics.Origins {
		if totalPopulation < originDynamics.Capacity-populationEpsilon {
			eligibleForCloning += originDynamics.Population
		}
	}

	natural := 0.0
	technology := 0.0
	housing := 0.0
	cloning := 0.0
	hasCloningCenter := colonyHasBuilding(colony, r.Rules.CloningCenterBuildingID)
	for _, originDynamics := range dynamics.Origins {
		origin := empireByID(state, originDynamics.OriginEmpireID)
		if origin == nil {
			return PlanningMetricBreakdown{}, fmt.Errorf("population origin empire %d not found", originDynamics.OriginEmpireID)
		}
		modifiers, ok := r.Rules.RaceModifiers[origin.RaceID]
		if !ok {
			return PlanningMetricBreakdown{}, fmt.Errorf("unknown origin race %q", origin.RaceID)
		}
		natural += originDynamics.BaseGrowth * modifiers.PopulationGrowthMultiplier
		technology += originDynamics.BaseGrowth * technologyBonus
		housing += originDynamics.BaseGrowth * housingBonus
		if hasCloningCenter && eligibleForCloning > populationEpsilon && totalPopulation < originDynamics.Capacity-populationEpsilon {
			cloning += r.Rules.CloningCenterFlatGrowth * originDynamics.Population / eligibleForCloning
		}
	}

	rawTotal := natural + technology + housing + cloning
	if rawTotal > metricEpsilon && finalGrowth < rawTotal-metricEpsilon {
		scale := finalGrowth / rawTotal
		natural *= scale
		technology *= scale
		housing *= scale
		cloning *= scale
	}

	components := make([]PlanningMetricComponent, 0, 5)
	components = appendMetricComponent(components, "natural", natural)
	components = appendMetricComponent(components, "technology", technology)
	components = appendMetricComponent(components, "housing", housing)
	components = appendMetricComponent(components, "cloning_center", cloning)
	explained := natural + technology + housing + cloning
	components = appendMetricComponent(components, "other", finalGrowth-explained)
	return PlanningMetricBreakdown{Total: finalGrowth, Components: components}, nil
}

func (r *EconomyResolver) planningTaxBCBreakdown(colony core.Colony, owner core.Empire) (PlanningMetricBreakdown, error) {
	baseTax, err := r.Rules.CalculateOwnerTaxBase(colony.Population.Total(), owner.RaceID)
	if err != nil {
		return PlanningMetricBreakdown{}, err
	}
	modifiers, ok := r.Rules.RaceModifiers[owner.RaceID]
	if !ok {
		return PlanningMetricBreakdown{}, fmt.Errorf("unknown owner race %q", owner.RaceID)
	}
	government, ok := r.Rules.GovernmentModifiers[modifiers.GovernmentTraitID]
	if !ok {
		return PlanningMetricBreakdown{}, fmt.Errorf("unknown government %q", modifiers.GovernmentTraitID)
	}
	_, _, _, err = r.Rules.calculateMorale(colony, modifiers.GovernmentTraitID, government.IgnoresMorale)
	if err != nil {
		return PlanningMetricBreakdown{}, err
	}

	governmentAdjusted := adjustedTaxIncome(baseTax, government, 0)
	finalTax := colony.AdjustedEconomy.TaxBC
	components := make([]PlanningMetricComponent, 0, 4)
	components = appendMetricComponent(components, "population_tax", baseTax)
	components = appendMetricComponent(components, "government", governmentAdjusted-baseTax)
	components = appendMetricComponent(components, "morale", finalTax-governmentAdjusted)
	explained := baseTax + (governmentAdjusted - baseTax) + (finalTax - governmentAdjusted)
	components = appendMetricComponent(components, "other", finalTax-explained)
	return PlanningMetricBreakdown{Total: finalTax, Components: components}, nil
}
