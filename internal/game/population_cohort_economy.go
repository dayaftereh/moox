package game

import (
	"fmt"
	"math"

	"moox/internal/core"
)

const conqueredOrganicOutputFactor = 0.75

func applyDifficultyToCohortBase(cohort core.PopulationCohort, base core.ColonyEconomy, profile DifficultyProfile) core.ColonyEconomy {
	factor := 1.0
	if cohort.AssimilationState == core.PopulationConquered {
		factor = conqueredOrganicOutputFactor
	}
	return core.ColonyEconomy{
		Food:       difficultyAdjustedRoleOutput(base.Food, cohort.Farmers, factor, difficultyEighths(profile.AIFoodPerFarmerEighths), true),
		Production: difficultyAdjustedRoleOutput(base.Production, cohort.Workers, factor, difficultyEighths(profile.AIProductionPerWorkerEighths), false),
		Research:   difficultyAdjustedRoleOutput(base.Research, cohort.Scientists, factor, difficultyEighths(profile.AIResearchPerScientistEighths), false),
	}
}

func difficultyAdjustedRoleOutput(base, workers, factor, delta float64, preserveZero bool) float64 {
	if workers <= 0 || factor <= 0 {
		return base
	}
	normalPerWorker := base / (workers * factor)
	if preserveZero && normalPerWorker <= 0 {
		return base
	}
	return workers * factor * math.Max(1, normalPerWorker+delta)
}

func applyDifficultyToTaxBase(population, baseTax float64, profile DifficultyProfile) float64 {
	if population <= 0 {
		return baseTax
	}
	normalPerPopulation := baseTax / population
	return population * math.Max(0, normalPerPopulation+difficultyEighths(profile.AITaxBCPerPopulationEighths))
}

// CalculateCohortBaseEconomy computes only job-derived Food/PP/RP for one
// organic cohort using the cohort origin race. Tax remains owner-wide.
func (r *EconomyRules) CalculateCohortBaseEconomy(cohort core.PopulationCohort, planet core.Planet, raceID string) (core.ColonyEconomy, error) {
	if r == nil {
		return core.ColonyEconomy{}, fmt.Errorf("economy rules must not be nil")
	}
	foodBase, ok := r.ClimateFoodPerFarmer[planet.ClimateID]
	if !ok {
		return core.ColonyEconomy{}, fmt.Errorf("unknown climate %q", planet.ClimateID)
	}
	industryBase, ok := r.MineralIndustryPerWorker[planet.MineralID]
	if !ok {
		return core.ColonyEconomy{}, fmt.Errorf("unknown mineral class %q", planet.MineralID)
	}
	modifiers, ok := r.RaceModifiers[raceID]
	if !ok {
		return core.ColonyEconomy{}, fmt.Errorf("unknown race %q", raceID)
	}
	if modifiers.Aquatic {
		if _, ok := r.AquaticFoodClimateIDs[planet.ClimateID]; ok {
			foodBase += r.AquaticFoodBonus
		}
	}
	foodPerFarmer := 0.0
	if foodBase > 0 {
		foodPerFarmer = math.Max(1, foodBase+modifiers.FoodPerFarmer)
	}
	productionPerWorker := math.Max(1, industryBase+modifiers.ProductionPerWorker)
	researchPerScientist := math.Max(1, r.BaseResearchPerScientist+modifiers.ResearchPerScientist)
	factor := 1.0
	if cohort.AssimilationState == core.PopulationConquered {
		factor = conqueredOrganicOutputFactor
	}
	return core.ColonyEconomy{
		Food:       cohort.Farmers * foodPerFarmer * factor,
		Production: cohort.Workers * productionPerWorker * factor,
		Research:   cohort.Scientists * researchPerScientist * factor,
	}, nil
}

func (r *EconomyRules) CalculateOwnerTaxBase(population float64, raceID string) (float64, error) {
	modifiers, ok := r.RaceModifiers[raceID]
	if !ok {
		return 0, fmt.Errorf("unknown race %q", raceID)
	}
	return population * math.Max(0, r.BaseTaxBCPerPopulation+modifiers.TaxBCPerPopulation), nil
}

// CalculateContextualCohortEconomy applies source-race gravity but owner-race
// government/morale. This is the mixed-population split proven by Gate 1.
func (r *EconomyRules) CalculateContextualCohortEconomy(base core.ColonyEconomy, colony core.Colony, planet core.Planet, sourceRaceID, ownerRaceID string) (core.ColonyEconomy, error) {
	sourceModifiers, ok := r.RaceModifiers[sourceRaceID]
	if !ok {
		return core.ColonyEconomy{}, fmt.Errorf("unknown source race %q", sourceRaceID)
	}
	ownerModifiers, ok := r.RaceModifiers[ownerRaceID]
	if !ok {
		return core.ColonyEconomy{}, fmt.Errorf("unknown owner race %q", ownerRaceID)
	}
	gravityPenalty, ok := r.GravityPenaltyPercent[gravityKey(sourceModifiers.GravityID, planet.GravityID)]
	if !ok {
		return core.ColonyEconomy{}, fmt.Errorf("no gravity rule for source race %q on planet gravity %q", sourceModifiers.GravityID, planet.GravityID)
	}
	government, ok := r.GovernmentModifiers[ownerModifiers.GovernmentTraitID]
	if !ok {
		return core.ColonyEconomy{}, fmt.Errorf("unknown government %q for owner race %q", ownerModifiers.GovernmentTraitID, ownerRaceID)
	}
	_, _, moralePercent, err := r.calculateMorale(colony, ownerModifiers.GovernmentTraitID, government.IgnoresMorale)
	if err != nil {
		return core.ColonyEconomy{}, err
	}
	return core.ColonyEconomy{
		Food:       adjustedRoleOutput(base.Food, government.FoodPercent, moralePercent, gravityPenalty),
		Production: adjustedRoleOutput(base.Production, government.ProductionPercent, moralePercent, gravityPenalty),
		Research:   adjustedRoleOutput(base.Research, government.ResearchPercent, moralePercent, gravityPenalty),
	}, nil
}

func (r *EconomyRules) CalculateOwnerAdjustedTax(baseTax float64, colony core.Colony, ownerRaceID string) (float64, error) {
	ownerModifiers, ok := r.RaceModifiers[ownerRaceID]
	if !ok {
		return 0, fmt.Errorf("unknown owner race %q", ownerRaceID)
	}
	government, ok := r.GovernmentModifiers[ownerModifiers.GovernmentTraitID]
	if !ok {
		return 0, fmt.Errorf("unknown government %q for owner race %q", ownerModifiers.GovernmentTraitID, ownerRaceID)
	}
	_, _, moralePercent, err := r.calculateMorale(colony, ownerModifiers.GovernmentTraitID, government.IgnoresMorale)
	if err != nil {
		return 0, err
	}
	return adjustedTaxIncome(baseTax, government, moralePercent), nil
}

func (r *EconomyResolver) calculateRaceAwareColonyEconomy(state *core.GameState, colony core.Colony, planet core.Planet, owner core.Empire) (core.ColonyEconomy, core.ColonyEconomyContext, core.ColonyEconomy, error) {
	profile, applyDifficulty, err := difficultyProfileForEmpire(state, owner)
	if err != nil {
		return core.ColonyEconomy{}, core.ColonyEconomyContext{}, core.ColonyEconomy{}, err
	}
	base := core.ColonyEconomy{}
	adjusted := core.ColonyEconomy{}
	for _, cohort := range colony.Population.Cohorts {
		origin := empireByID(state, cohort.OriginEmpireID)
		if origin == nil {
			return core.ColonyEconomy{}, core.ColonyEconomyContext{}, core.ColonyEconomy{}, fmt.Errorf("cohort origin empire %d not found", cohort.OriginEmpireID)
		}
		cohortBase, err := r.Rules.CalculateCohortBaseEconomy(cohort, planet, origin.RaceID)
		if err != nil {
			return core.ColonyEconomy{}, core.ColonyEconomyContext{}, core.ColonyEconomy{}, err
		}
		if applyDifficulty {
			cohortBase = applyDifficultyToCohortBase(cohort, cohortBase, profile)
		}
		cohortAdjusted, err := r.Rules.CalculateContextualCohortEconomy(cohortBase, colony, planet, origin.RaceID, owner.RaceID)
		if err != nil {
			return core.ColonyEconomy{}, core.ColonyEconomyContext{}, core.ColonyEconomy{}, err
		}
		base.Food += cohortBase.Food
		base.Production += cohortBase.Production
		base.Research += cohortBase.Research
		adjusted.Food += cohortAdjusted.Food
		adjusted.Production += cohortAdjusted.Production
		adjusted.Research += cohortAdjusted.Research
	}
	baseTax, err := r.Rules.CalculateOwnerTaxBase(colony.Population.Total(), owner.RaceID)
	if err != nil {
		return core.ColonyEconomy{}, core.ColonyEconomyContext{}, core.ColonyEconomy{}, err
	}
	if applyDifficulty {
		baseTax = applyDifficultyToTaxBase(colony.Population.Total(), baseTax, profile)
	}
	adjustedTax, err := r.Rules.CalculateOwnerAdjustedTax(baseTax, colony, owner.RaceID)
	if err != nil {
		return core.ColonyEconomy{}, core.ColonyEconomyContext{}, core.ColonyEconomy{}, err
	}
	base.TaxBC = baseTax
	adjusted.TaxBC = adjustedTax

	// Preserve the owner-wide context metadata for UI/observer compatibility.
	// Per-origin gravity is already applied above and cannot be represented by
	// the legacy single GravityPenaltyPercent field on a mixed colony.
	context, _, err := r.Rules.CalculateContextualEconomy(core.ColonyEconomy{}, colony, planet, owner.RaceID)
	if err != nil {
		return core.ColonyEconomy{}, core.ColonyEconomyContext{}, core.ColonyEconomy{}, err
	}
	return base, context, adjusted, nil
}
