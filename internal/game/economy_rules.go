package game

import (
	"fmt"
	"math"
	"path/filepath"

	"moox/internal/core"
	"moox/internal/ruleset"
)

type GovernmentEconomyModifier struct {
	FoodPercent       int
	ProductionPercent int
	ResearchPercent   int
	TaxPercent        int
	TaxBonusRounding  string
	IgnoresMorale     bool
}

type RaceEconomyModifiers struct {
	FoodPerFarmerMilli        int64
	ProductionPerWorkerMilli  int64
	ResearchPerScientistMilli int64
	TaxBCPerPopulationMilli   int64
	Aquatic                   bool
	GravityID                 string
	GovernmentTraitID         string
}

type BuildingDefinition struct {
	BuildingID          string
	ProductionID        int
	TechnologyID        int
	ProductionCostMilli int64
	MaintenanceBC       int
}
type EconomyRules struct {
	ClimateFoodPerFarmerMilli     map[string]int64
	MineralIndustryPerWorkerMilli map[string]int64
	RaceModifiers                 map[string]RaceEconomyModifiers
	AquaticFoodBonusMilli         int64
	AquaticFoodClimateIDs         map[string]struct{}
	BaseResearchPerScientistMilli int64
	BaseTaxBCPerPopulationMilli   int64
	GravityPenaltyPercent         map[string]int
	GovernmentModifiers           map[string]GovernmentEconomyModifier
	MoraleBarracksPenaltyPercent  int
	MoraleBarracksGovernments     map[string]struct{}
	MoraleBarracksBuildingIDs     map[string]struct{}
	MoraleBuildingBonusPercent    map[string]int
	KnownBuildingIDs              map[string]struct{}
	BuildingDefinitions           map[string]BuildingDefinition
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
	buildings, err := ruleset.LoadBuildings(filepath.Join(rulesetDir, "buildings.json"))
	if err != nil {
		return nil, fmt.Errorf("load buildings: %w", err)
	}
	if err := buildings.Validate(); err != nil {
		return nil, fmt.Errorf("validate buildings: %w", err)
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

	gravityIDs := make(map[string]struct{}, len(planetClasses.GravityClasses))
	for _, gravity := range planetClasses.GravityClasses {
		gravityIDs[gravity.ID] = struct{}{}
	}
	gravityPenalties := make(map[string]int, len(economy.GravityPenalties))
	for _, rule := range economy.GravityPenalties {
		if _, ok := gravityIDs[rule.RaceGravityID]; !ok {
			return nil, fmt.Errorf("gravity rule references unknown race gravity %q", rule.RaceGravityID)
		}
		if _, ok := gravityIDs[rule.PlanetGravityID]; !ok {
			return nil, fmt.Errorf("gravity rule references unknown planet gravity %q", rule.PlanetGravityID)
		}
		gravityPenalties[gravityKey(rule.RaceGravityID, rule.PlanetGravityID)] = rule.Percent
	}

	governmentModifiers := make(map[string]GovernmentEconomyModifier, len(economy.GovernmentModifiers))
	for _, rule := range economy.GovernmentModifiers {
		governmentModifiers[rule.TraitID] = GovernmentEconomyModifier{
			FoodPercent:       rule.FoodPercent,
			ProductionPercent: rule.ProductionPercent,
			ResearchPercent:   rule.ResearchPercent,
			TaxPercent:        rule.TaxPercent,
			TaxBonusRounding:  rule.TaxBonusRounding,
			IgnoresMorale:     rule.IgnoresMorale,
		}
	}

	buildingIDs := make(map[string]struct{}, len(buildings.Buildings))
	buildingDefinitions := make(map[string]BuildingDefinition, len(buildings.Buildings))
	for _, building := range buildings.Buildings {
		buildingIDs[building.ID] = struct{}{}
		buildingDefinitions[building.ID] = BuildingDefinition{BuildingID: building.ID, ProductionID: building.ProductionID, TechnologyID: building.TechnologyID, ProductionCostMilli: int64(building.ProductionCostPP) * core.EconomyScale, MaintenanceBC: building.MaintenanceBC}
	}
	moraleBarracksGovernments := make(map[string]struct{}, len(economy.Morale.BarracksGovernmentTraitIDs))
	for _, traitID := range economy.Morale.BarracksGovernmentTraitIDs {
		moraleBarracksGovernments[traitID] = struct{}{}
	}
	moraleBarracksBuildingIDs := make(map[string]struct{}, len(economy.Morale.BarracksBuildingIDs))
	for _, buildingID := range economy.Morale.BarracksBuildingIDs {
		if _, ok := buildingIDs[buildingID]; !ok {
			return nil, fmt.Errorf("morale barracks rule references unknown building %q", buildingID)
		}
		moraleBarracksBuildingIDs[buildingID] = struct{}{}
	}
	moraleBuildingBonusPercent := make(map[string]int, len(economy.Morale.BuildingBonuses))
	for _, bonus := range economy.Morale.BuildingBonuses {
		if _, ok := buildingIDs[bonus.BuildingID]; !ok {
			return nil, fmt.Errorf("morale bonus references unknown building %q", bonus.BuildingID)
		}
		moraleBuildingBonusPercent[bonus.BuildingID] = bonus.Percent
	}

	rules := &EconomyRules{
		ClimateFoodPerFarmerMilli:     make(map[string]int64, len(planetClasses.Climates)),
		MineralIndustryPerWorkerMilli: make(map[string]int64, len(planetClasses.MineralClasses)),
		RaceModifiers:                 make(map[string]RaceEconomyModifiers, len(races.Races)),
		AquaticFoodBonusMilli:         int64(economy.AquaticFoodBonus.Value) * core.EconomyScale,
		AquaticFoodClimateIDs:         aquaticClimates,
		BaseResearchPerScientistMilli: int64(economy.BaseResearchPerScientist.Value) * core.EconomyScale,
		BaseTaxBCPerPopulationMilli:   int64(economy.BaseTaxBCPerPopulation.Value) * core.EconomyScale,
		GravityPenaltyPercent:         gravityPenalties,
		GovernmentModifiers:           governmentModifiers,
		MoraleBarracksPenaltyPercent:  economy.Morale.BarracksPenaltyPercent,
		MoraleBarracksGovernments:     moraleBarracksGovernments,
		MoraleBarracksBuildingIDs:     moraleBarracksBuildingIDs,
		MoraleBuildingBonusPercent:    moraleBuildingBonusPercent,
		KnownBuildingIDs:              buildingIDs,
		BuildingDefinitions:           buildingDefinitions,
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
	for governmentTraitID := range governmentModifiers {
		option, ok := options[governmentTraitID]
		if !ok || option.Ability != governmentTraitID {
			return nil, fmt.Errorf("economy government rule %q does not match normalized race trait ability", governmentTraitID)
		}
	}

	for _, race := range races.Races {
		modifiers := RaceEconomyModifiers{GravityID: "normal_g"}
		for _, selection := range race.TraitSelections {
			option, ok := options[selection.TraitID]
			if !ok {
				continue
			}
			switch option.Ability {
			case "aquatic":
				modifiers.Aquatic = true
			case "low_g_world":
				modifiers.GravityID = "low_g"
			case "high_g_world":
				modifiers.GravityID = "heavy_g"
			case "government_feudal", "government_dictatorship", "government_democracy", "government_unification":
				modifiers.GovernmentTraitID = option.Ability
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
		if modifiers.GovernmentTraitID == "" {
			return nil, fmt.Errorf("race %q has no normalized starting government", race.ID)
		}
		if _, ok := governmentModifiers[modifiers.GovernmentTraitID]; !ok {
			return nil, fmt.Errorf("race %q uses unsupported government %q", race.ID, modifiers.GovernmentTraitID)
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

func (r *EconomyRules) CalculateContextualEconomy(base core.ColonyEconomy, colony core.Colony, planet core.Planet, raceID string) (core.ColonyEconomyContext, core.ColonyEconomy, error) {
	if r == nil {
		return core.ColonyEconomyContext{}, core.ColonyEconomy{}, fmt.Errorf("economy rules must not be nil")
	}
	modifiers, ok := r.RaceModifiers[raceID]
	if !ok {
		return core.ColonyEconomyContext{}, core.ColonyEconomy{}, fmt.Errorf("unknown race %q", raceID)
	}
	gravityPenalty, ok := r.GravityPenaltyPercent[gravityKey(modifiers.GravityID, planet.GravityID)]
	if !ok {
		return core.ColonyEconomyContext{}, core.ColonyEconomy{}, fmt.Errorf("no gravity rule for race %q on planet gravity %q", modifiers.GravityID, planet.GravityID)
	}
	government, ok := r.GovernmentModifiers[modifiers.GovernmentTraitID]
	if !ok {
		return core.ColonyEconomyContext{}, core.ColonyEconomy{}, fmt.Errorf("unknown government %q for race %q", modifiers.GovernmentTraitID, raceID)
	}
	barracksPenalty, buildingBonus, moralePercent, err := r.calculateMorale(colony, modifiers.GovernmentTraitID, government.IgnoresMorale)
	if err != nil {
		return core.ColonyEconomyContext{}, core.ColonyEconomy{}, err
	}

	context := core.ColonyEconomyContext{
		RaceGravityID:                modifiers.GravityID,
		PlanetGravityID:              planet.GravityID,
		GravityPenaltyPercent:        gravityPenalty,
		GovernmentTraitID:            modifiers.GovernmentTraitID,
		GovernmentFoodPercent:        government.FoodPercent,
		GovernmentProductionPercent:  government.ProductionPercent,
		GovernmentResearchPercent:    government.ResearchPercent,
		GovernmentTaxPercent:         government.TaxPercent,
		GovernmentIgnoresMorale:      government.IgnoresMorale,
		MoraleBarracksPenaltyPercent: barracksPenalty,
		MoraleBuildingBonusPercent:   buildingBonus,
		MoralePercent:                moralePercent,
	}
	adjusted := core.ColonyEconomy{
		FoodMilli:       adjustedRoleOutput(base.FoodMilli, government.FoodPercent, moralePercent, gravityPenalty),
		ProductionMilli: adjustedRoleOutput(base.ProductionMilli, government.ProductionPercent, moralePercent, gravityPenalty),
		ResearchMilli:   adjustedRoleOutput(base.ResearchMilli, government.ResearchPercent, moralePercent, gravityPenalty),
		TaxBCMilli:      adjustedTaxIncome(base.TaxBCMilli, government, moralePercent),
	}
	return context, adjusted, nil
}

func (r *EconomyRules) calculateMorale(colony core.Colony, governmentTraitID string, ignoresMorale bool) (int, int, int, error) {
	barracksPenalty := 0
	buildingBonus := 0
	hasBarracks := false
	for _, buildingID := range colony.Buildings {
		if _, ok := r.KnownBuildingIDs[buildingID]; !ok {
			return 0, 0, 0, fmt.Errorf("colony %d has unknown building %q", colony.ID, buildingID)
		}
		if _, ok := r.MoraleBarracksBuildingIDs[buildingID]; ok {
			hasBarracks = true
		}
		buildingBonus += r.MoraleBuildingBonusPercent[buildingID]
	}
	if _, needsBarracks := r.MoraleBarracksGovernments[governmentTraitID]; needsBarracks && !hasBarracks {
		barracksPenalty = r.MoraleBarracksPenaltyPercent
	}
	rawMorale := barracksPenalty + buildingBonus
	if ignoresMorale {
		return barracksPenalty, buildingBonus, 0, nil
	}
	return barracksPenalty, buildingBonus, rawMorale, nil
}
func adjustedRoleOutput(baseMilli int64, governmentPercent, moralePercent, gravityPenaltyPercent int) int64 {
	if baseMilli <= 0 {
		return 0
	}
	percent := 100 + governmentPercent + moralePercent - gravityPenaltyPercent
	if percent <= 0 {
		return 0
	}
	return roundMilliToWhole(baseMilli * int64(percent) / 100)
}

func adjustedTaxIncome(baseMilli int64, government GovernmentEconomyModifier, moralePercent int) int64 {
	if baseMilli <= 0 {
		return 0
	}
	governmentBonusMilli := int64(0)
	if government.TaxPercent != 0 {
		governmentBonusMilli = baseMilli * int64(government.TaxPercent) / 100
		switch government.TaxBonusRounding {
		case "down":
			governmentBonusMilli = floorMilliToWhole(governmentBonusMilli)
		default:
			governmentBonusMilli = roundSignedMilliToWhole(governmentBonusMilli)
		}
	}
	moraleBonusMilli := roundSignedMilliToWhole(baseMilli * int64(moralePercent) / 100)
	return max64(0, baseMilli+governmentBonusMilli+moraleBonusMilli)
}
func gravityKey(raceGravityID, planetGravityID string) string {
	return raceGravityID + "/" + planetGravityID
}

func roundSignedMilliToWhole(value int64) int64 {
	if value < 0 {
		return -roundMilliToWhole(-value)
	}
	return roundMilliToWhole(value)
}
func roundMilliToWhole(value int64) int64 {
	if value <= 0 {
		return 0
	}
	return ((value + core.EconomyScale/2) / core.EconomyScale) * core.EconomyScale
}

func floorMilliToWhole(value int64) int64 {
	if value <= 0 {
		return 0
	}
	return (value / core.EconomyScale) * core.EconomyScale
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
