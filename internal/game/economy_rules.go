package game

import (
	"fmt"
	"math"
	"path/filepath"

	"moox/internal/core"
	"moox/internal/ruleset"
)

type RaceEconomyModifiers struct {
	FoodPerFarmerMilli        int64
	ProductionPerWorkerMilli  int64
	ResearchPerScientistMilli int64
	TaxBCPerPopulationMilli   int64
	Aquatic                   bool
}

type EconomyRules struct {
	ClimateFoodPerFarmerMilli     map[string]int64
	MineralIndustryPerWorkerMilli map[string]int64
	RaceModifiers                 map[string]RaceEconomyModifiers
	AquaticFoodBonusMilli         int64
	AquaticFoodClimateIDs         map[string]struct{}
	BaseResearchPerScientistMilli int64
	BaseTaxBCPerPopulationMilli   int64
}

func LoadEconomyRules(rulesetDir string) (*EconomyRules, error) {
	planetClasses, err := ruleset.LoadPlanetClasses(filepath.Join(rulesetDir, "planet_classes.json"))
	if err != nil {
		return nil, fmt.Errorf("load planet classes: %w", err)
	}
	if err := planetClasses.Validate(); err != nil {
		return nil, fmt.Errorf("validate planet classes: %w", err)
	}
	economy, err := ruleset.LoadEconomy(filepath.Join(rulesetDir, "economy.json"))
	if err != nil {
		return nil, fmt.Errorf("load economy: %w", err)
	}
	races, err := ruleset.LoadRaces(filepath.Join(rulesetDir, "races.json"))
	if err != nil {
		return nil, fmt.Errorf("load races: %w", err)
	}
	traits, err := ruleset.LoadRaceTraits(filepath.Join(rulesetDir, "race_traits.json"))
	if err != nil {
		return nil, fmt.Errorf("load race traits: %w", err)
	}
	if err := races.ValidateAgainstTraits(traits); err != nil {
		return nil, fmt.Errorf("validate races against traits: %w", err)
	}

	industryByMineral := make(map[string]int, len(economy.MineralIndustryPerWorker))
	for _, item := range economy.MineralIndustryPerWorker {
		industryByMineral[item.MineralID] = item.Value
	}
	if len(industryByMineral) != len(planetClasses.MineralClasses) {
		return nil, fmt.Errorf("economy mineral count %d does not match planet class count %d", len(industryByMineral), len(planetClasses.MineralClasses))
	}

	climateIDs := make(map[string]struct{}, len(planetClasses.Climates))
	for _, climate := range planetClasses.Climates {
		climateIDs[climate.ID] = struct{}{}
	}
	aquaticClimates := make(map[string]struct{}, len(economy.AquaticFoodBonus.ClimateIDs))
	for _, climateID := range economy.AquaticFoodBonus.ClimateIDs {
		if _, ok := climateIDs[climateID]; !ok {
			return nil, fmt.Errorf("aquatic food rule references unknown climate %q", climateID)
		}
		aquaticClimates[climateID] = struct{}{}
	}

	rules := &EconomyRules{
		ClimateFoodPerFarmerMilli:     make(map[string]int64, len(planetClasses.Climates)),
		MineralIndustryPerWorkerMilli: make(map[string]int64, len(planetClasses.MineralClasses)),
		RaceModifiers:                 make(map[string]RaceEconomyModifiers, len(races.Races)),
		AquaticFoodBonusMilli:         int64(economy.AquaticFoodBonus.Value) * core.EconomyScale,
		AquaticFoodClimateIDs:         aquaticClimates,
		BaseResearchPerScientistMilli: int64(economy.BaseResearchPerScientist.Value) * core.EconomyScale,
		BaseTaxBCPerPopulationMilli:   int64(economy.BaseTaxBCPerPopulation.Value) * core.EconomyScale,
	}
	for _, climate := range planetClasses.Climates {
		rules.ClimateFoodPerFarmerMilli[climate.ID] = int64(climate.BaseFoodPerFarmer) * core.EconomyScale
	}
	for _, mineral := range planetClasses.MineralClasses {
		value, ok := industryByMineral[mineral.ID]
		if !ok {
			return nil, fmt.Errorf("economy rules missing mineral class %q", mineral.ID)
		}
		rules.MineralIndustryPerWorkerMilli[mineral.ID] = int64(value) * core.EconomyScale
	}

	options := make(map[string]ruleset.RaceTraitOption)
	for _, group := range traits.Groups {
		for _, option := range group.Options {
			options[option.ID] = option
		}
	}
	for _, race := range races.Races {
		var modifiers RaceEconomyModifiers
		for _, selection := range race.TraitSelections {
			option, ok := options[selection.TraitID]
			if !ok {
				continue
			}
			if option.Ability == "aquatic" {
				modifiers.Aquatic = true
			}
			if option.Value == nil {
				continue
			}
			delta := int64(math.Round(*option.Value * float64(core.EconomyScale)))
			switch option.ValueKind {
			case "food_per_farmer_delta":
				modifiers.FoodPerFarmerMilli += delta
			case "production_per_worker_delta":
				modifiers.ProductionPerWorkerMilli += delta
			case "research_per_scientist_delta":
				modifiers.ResearchPerScientistMilli += delta
			case "tax_bc_per_population_delta":
				modifiers.TaxBCPerPopulationMilli += delta
			}
		}
		rules.RaceModifiers[race.ID] = modifiers
	}
	return rules, nil
}

func (r *EconomyRules) CalculateBaseEconomy(colony core.Colony, planet core.Planet, raceID string) (core.ColonyEconomy, error) {
	if r == nil {
		return core.ColonyEconomy{}, fmt.Errorf("economy rules must not be nil")
	}
	foodBase, ok := r.ClimateFoodPerFarmerMilli[planet.ClimateID]
	if !ok {
		return core.ColonyEconomy{}, fmt.Errorf("unknown climate %q", planet.ClimateID)
	}
	industryBase, ok := r.MineralIndustryPerWorkerMilli[planet.MineralID]
	if !ok {
		return core.ColonyEconomy{}, fmt.Errorf("unknown mineral class %q", planet.MineralID)
	}
	modifiers, ok := r.RaceModifiers[raceID]
	if !ok {
		return core.ColonyEconomy{}, fmt.Errorf("unknown race %q", raceID)
	}

	// Farming racial bonuses do not make a no-farming climate farmable. Aquatic
	// modifies only the explicitly normalized wet-climate coefficients.
	if modifiers.Aquatic {
		if _, ok := r.AquaticFoodClimateIDs[planet.ClimateID]; ok {
			foodBase += r.AquaticFoodBonusMilli
		}
	}
	foodPerFarmer := int64(0)
	if foodBase > 0 {
		foodPerFarmer = max64(core.EconomyScale, foodBase+modifiers.FoodPerFarmerMilli)
	}
	productionPerWorker := max64(core.EconomyScale, industryBase+modifiers.ProductionPerWorkerMilli)
	researchPerScientist := max64(core.EconomyScale, r.BaseResearchPerScientistMilli+modifiers.ResearchPerScientistMilli)
	taxPerPopulation := max64(0, r.BaseTaxBCPerPopulationMilli+modifiers.TaxBCPerPopulationMilli)

	population := colony.Population
	rawTaxMilli := int64(population.Units) * taxPerPopulation
	roundedTaxMilli := roundMilliToWhole(rawTaxMilli)
	return core.ColonyEconomy{
		FoodMilli:       int64(population.Farmers) * foodPerFarmer,
		ProductionMilli: int64(population.Workers) * productionPerWorker,
		ResearchMilli:   int64(population.Scientists) * researchPerScientist,
		TaxBCMilli:      roundedTaxMilli,
	}, nil
}

func roundMilliToWhole(value int64) int64 {
	if value <= 0 {
		return 0
	}
	return ((value + core.EconomyScale/2) / core.EconomyScale) * core.EconomyScale
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
