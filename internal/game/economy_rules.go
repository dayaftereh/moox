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
	FoodPerFarmer              float64
	ProductionPerWorker        float64
	ResearchPerScientist       float64
	TaxBCPerPopulation         float64
	PopulationGrowthMultiplier float64
	Aquatic                    bool
	Tolerant                   bool
	Subterranean               bool
	Cybernetic                 bool
	Lithovore                  bool
	FantasticTraders           bool
	TransDimensional           bool
	Creative                   bool
	Uncreative                 bool
	GravityID                  string
	GovernmentTraitID          string
}

type BuildingDefinition struct {
	BuildingID       string
	ProductionID     int
	TechnologyID     int
	ProductionCostPP float64
	MaintenanceBC    int
}

type BarrenOrbitTransformationRule struct {
	InnerOrbitMax int
	MiddleOrbit   int
	OuterOrbitMin int
	InnerResult   string
	OuterResult   string
	MiddleResults []string
}

type PlanetaryTransformationDefinition struct {
	ProjectID             string
	ProductionID          int
	TechnologyID          int
	ProductionCostPP      float64
	AllowedClimateIDs     map[string]struct{}
	ResultClimateBySource map[string]string
	BarrenOrbitRule       *BarrenOrbitTransformationRule
}

type EconomyRules struct {
	ClimateFoodPerFarmer                          map[string]float64
	MineralIndustryPerWorker                      map[string]float64
	RaceModifiers                                 map[string]RaceEconomyModifiers
	RaceResearchModifiers                         map[string]RaceResearchModifiers
	AquaticFoodBonus                              float64
	AquaticFoodClimateIDs                         map[string]struct{}
	BaseResearchPerScientist                      float64
	BaseTaxBCPerPopulation                        float64
	PopulationFoodPerUnit                         float64
	CyberneticFoodPerUnit                         float64
	CyberneticProductionPerUnit                   float64
	PopulationGrowthCurveFactor                   float64
	HousingGrowthPercentPerPPPerPopulation        float64
	HousingGrowthPercentRounding                  string
	CloningCenterBuildingID                       string
	CloningCenterFlatGrowth                       float64
	PopulationGrowthTechnologyBonusByID           map[int]float64
	AdvancedCityPlanningTechnologyID              int
	AdvancedCityPlanningCapacityBonus             float64
	BiospheresBuildingID                          string
	BiospheresCapacityBonus                       float64
	PlanetaryTransformations                      map[string]PlanetaryTransformationDefinition
	TolerantHabitabilityBonus                     float64
	SubterraneanCapacityPerClass                  float64
	PopulationSizeCapacity                        map[string]float64
	PopulationSizeClass                           map[string]int
	PopulationClimateHabitability                 map[string]float64
	FreighterFoodCapacity                         float64
	FreightersPerFleet                            int
	FreighterFleetCostPP                          float64
	FreighterOperatingCostBC                      float64
	SurplusFoodBCPerUnit                          float64
	FantasticTradersSurplusFoodBCPerUnit          float64
	StarvationPopulationPerFoodShortage           float64
	CyberneticStarvationPopulationPerFoodShortage float64
	CyberneticStarvationPopulationPerPPShortage   float64
	MinimumPopulationAfterStarvation              float64
	GravityPenaltyPercent                         map[string]int
	GovernmentModifiers                           map[string]GovernmentEconomyModifier
	MoraleBarracksPenaltyPercent                  int
	MoraleBarracksGovernments                     map[string]struct{}
	MoraleBarracksBuildingIDs                     map[string]struct{}
	MoraleBuildingBonusPercent                    map[string]int
	KnownBuildingIDs                              map[string]struct{}
	BuildingDefinitions                           map[string]BuildingDefinition
	TechnologyFieldCostsRP                        map[int]float64
	TechnologyFieldPreviousID                     map[int]int
	TechnologyFieldNextID                         map[int]int
	TechnologyFieldAIGroup                        map[int]int
	TechnologyIDsByField                          map[int][]int
	TechnologyFieldByID                           map[int]int
	TechnologyKeyByID                             map[int]string
	TechnologyNameKeyByID                         map[int]string
	TechnologyStrategicAvailable                  map[int]bool
	TechnologyAIClassByID                         map[int]int
	TechnologyAIClasses                           map[int]TechnologyAIClassDefinition
	TechnologyAIFieldGroupValues                  []int
	NewGameAlwaysKnownFieldID                     int
	NewGameStagedKnownFieldIDs                    []int
	GeneralResearchFieldIDs                       map[int]struct{}
	HyperAdvancedResearchFieldIDs                 map[int]struct{}
	HyperAdvancedCostIncrementRP                  float64
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
	technologies, err := ruleset.LoadTechnologies(filepath.Join(rulesetDir, "technologies.json"))
	if err != nil {
		return nil, fmt.Errorf("load technologies: %w", err)
	}
	if err := technologies.Validate(); err != nil {
		return nil, fmt.Errorf("validate technologies: %w", err)
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

	technologyFieldCosts := make(map[int]float64, len(technologies.Fields))
	technologyFieldPreviousID := make(map[int]int, len(technologies.Fields))
	technologyFieldNextID := make(map[int]int, len(technologies.Fields))
	technologyFieldAIGroup := make(map[int]int, len(technologies.Fields))
	for _, field := range technologies.Fields {
		technologyFieldCosts[field.FieldID] = float64(field.ResearchCost)
		technologyFieldPreviousID[field.FieldID] = field.PreviousID
		technologyFieldNextID[field.FieldID] = field.NextID
		technologyFieldAIGroup[field.FieldID] = field.AIGroup
	}
	technologyIDsByField := make(map[int][]int)
	technologyFieldByID := make(map[int]int, len(technologies.Technologies))
	technologyKeyByID := make(map[int]string, len(technologies.Technologies))
	technologyIDByKey := make(map[string]int, len(technologies.Technologies))
	technologyNameKeyByID := make(map[int]string, len(technologies.Technologies))
	technologyStrategicAvailable := make(map[int]bool, len(technologies.Technologies))
	technologyAIClassByID := make(map[int]int, len(technologies.Technologies))
	for _, technology := range technologies.Technologies {
		technologyFieldByID[technology.TechnologyID] = technology.TechFieldID
		technologyKeyByID[technology.TechnologyID] = technology.ID
		technologyIDByKey[technology.ID] = technology.TechnologyID
		technologyNameKeyByID[technology.TechnologyID] = technology.NameKey
		technologyStrategicAvailable[technology.TechnologyID] = technology.StrategicCombatAvailable
		technologyAIClassByID[technology.TechnologyID] = technology.AIClass
		technologyIDsByField[technology.TechFieldID] = append(technologyIDsByField[technology.TechFieldID], technology.TechnologyID)
	}
	technologyAIClasses := make(map[int]TechnologyAIClassDefinition, len(technologies.AIResearch.TechnologyClasses))
	for _, class := range technologies.AIResearch.TechnologyClasses {
		technologyAIClasses[class.ClassID] = TechnologyAIClassDefinition{BaseWeight: class.BaseWeight, CompetitionSensitive: class.CompetitionSensitive}
	}
	technologyAIFieldGroupValues := append([]int(nil), technologies.AIResearch.FieldGroupValues...)
	buildingIDs := make(map[string]struct{}, len(buildings.Buildings))
	buildingDefinitions := make(map[string]BuildingDefinition, len(buildings.Buildings))
	for _, building := range buildings.Buildings {
		buildingIDs[building.ID] = struct{}{}
		buildingDefinitions[building.ID] = BuildingDefinition{BuildingID: building.ID, ProductionID: building.ProductionID, TechnologyID: building.TechnologyID, ProductionCostPP: float64(building.ProductionCostPP), MaintenanceBC: building.MaintenanceBC}
	}
	growthRules := economy.Population.GrowthModifiers
	if _, ok := buildingIDs[growthRules.CloningCenterBuildingID]; !ok {
		return nil, fmt.Errorf("population growth rule references unknown Cloning Center building %q", growthRules.CloningCenterBuildingID)
	}
	populationGrowthTechnologyBonusByID := make(map[int]float64, len(growthRules.TechnologyBonuses))
	for _, bonus := range growthRules.TechnologyBonuses {
		technologyID, ok := technologyIDByKey[bonus.TechnologyKey]
		if !ok {
			return nil, fmt.Errorf("population growth rule references unknown technology %q", bonus.TechnologyKey)
		}
		populationGrowthTechnologyBonusByID[technologyID] = bonus.Bonus
	}
	capacityRules := economy.Population.CapacityModifiers
	advancedCityPlanningTechnologyID, ok := technologyIDByKey[capacityRules.AdvancedCityPlanningTechnologyKey]
	if !ok {
		return nil, fmt.Errorf("population capacity rule references unknown technology %q", capacityRules.AdvancedCityPlanningTechnologyKey)
	}
	if _, ok := buildingIDs[capacityRules.BiospheresBuildingID]; !ok {
		return nil, fmt.Errorf("population capacity rule references unknown Biospheres building %q", capacityRules.BiospheresBuildingID)
	}
	planetaryTransformations := make(map[string]PlanetaryTransformationDefinition, len(capacityRules.PlanetaryTransformations))
	for _, transformation := range capacityRules.PlanetaryTransformations {
		building, ok := buildingDefinitions[transformation.ProjectID]
		if !ok {
			return nil, fmt.Errorf("population capacity rule references unknown transformation production identity %q", transformation.ProjectID)
		}
		allowed := make(map[string]struct{}, len(transformation.AllowedClimateIDs))
		for _, climateID := range transformation.AllowedClimateIDs {
			if _, ok := climateIDs[climateID]; !ok {
				return nil, fmt.Errorf("planetary transformation %q references unknown allowed climate %q", transformation.ProjectID, climateID)
			}
			allowed[climateID] = struct{}{}
		}
		results := make(map[string]string, len(transformation.ResultClimateBySource))
		for sourceClimate, resultClimate := range transformation.ResultClimateBySource {
			if _, ok := climateIDs[sourceClimate]; !ok {
				return nil, fmt.Errorf("planetary transformation %q references unknown source climate %q", transformation.ProjectID, sourceClimate)
			}
			if _, ok := climateIDs[resultClimate]; !ok {
				return nil, fmt.Errorf("planetary transformation %q references unknown result climate %q", transformation.ProjectID, resultClimate)
			}
			results[sourceClimate] = resultClimate
		}
		var barren *BarrenOrbitTransformationRule
		if rule := transformation.BarrenOrbitRule; rule != nil {
			for _, climateID := range append([]string{rule.InnerResult, rule.OuterResult}, rule.MiddleResults...) {
				if _, ok := climateIDs[climateID]; !ok {
					return nil, fmt.Errorf("planetary transformation %q barren rule references unknown result climate %q", transformation.ProjectID, climateID)
				}
			}
			barren = &BarrenOrbitTransformationRule{InnerOrbitMax: rule.InnerOrbitMax, MiddleOrbit: rule.MiddleOrbit, OuterOrbitMin: rule.OuterOrbitMin, InnerResult: rule.InnerResult, OuterResult: rule.OuterResult, MiddleResults: append([]string(nil), rule.MiddleResults...)}
		}
		planetaryTransformations[transformation.ProjectID] = PlanetaryTransformationDefinition{
			ProjectID: transformation.ProjectID, ProductionID: building.ProductionID, TechnologyID: building.TechnologyID, ProductionCostPP: building.ProductionCostPP, AllowedClimateIDs: allowed, ResultClimateBySource: results, BarrenOrbitRule: barren,
		}
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

	populationSizeCapacity := make(map[string]float64, len(economy.Population.SizeCapacity))
	populationSizeClass := make(map[string]int, len(planetClasses.Sizes))
	for _, size := range economy.Population.SizeCapacity {
		populationSizeCapacity[size.SizeID] = size.Capacity
	}
	for _, size := range planetClasses.Sizes {
		if _, ok := populationSizeCapacity[size.ID]; !ok {
			return nil, fmt.Errorf("population rules missing planet size %q", size.ID)
		}
		populationSizeClass[size.ID] = size.Index + 1
	}
	populationClimateHabitability := make(map[string]float64, len(economy.Population.ClimateHabitability))
	for _, climate := range economy.Population.ClimateHabitability {
		populationClimateHabitability[climate.ClimateID] = climate.Fraction
	}
	for _, climate := range planetClasses.Climates {
		if _, ok := populationClimateHabitability[climate.ID]; !ok {
			return nil, fmt.Errorf("population rules missing planet climate %q", climate.ID)
		}
	}

	hyperAdvancedResearchFieldIDs := make(map[int]struct{}, len(technologies.HyperAdvanced.TechFieldIDs))
	for _, fieldID := range technologies.HyperAdvanced.TechFieldIDs {
		hyperAdvancedResearchFieldIDs[fieldID] = struct{}{}
	}

	generalResearchFieldIDs := make(map[int]struct{}, len(technologies.NewGameStart.StagedKnownTechFieldIDs)+1)
	generalResearchFieldIDs[technologies.NewGameStart.AlwaysKnownTechFieldID] = struct{}{}
	for _, fieldID := range technologies.NewGameStart.StagedKnownTechFieldIDs {
		generalResearchFieldIDs[fieldID] = struct{}{}
	}

	rules := &EconomyRules{
		ClimateFoodPerFarmer:                          make(map[string]float64, len(planetClasses.Climates)),
		MineralIndustryPerWorker:                      make(map[string]float64, len(planetClasses.MineralClasses)),
		RaceModifiers:                                 make(map[string]RaceEconomyModifiers, len(races.Races)),
		RaceResearchModifiers:                         make(map[string]RaceResearchModifiers, len(races.Races)),
		AquaticFoodBonus:                              float64(economy.AquaticFoodBonus.Value),
		AquaticFoodClimateIDs:                         aquaticClimates,
		BaseResearchPerScientist:                      float64(economy.BaseResearchPerScientist.Value),
		BaseTaxBCPerPopulation:                        float64(economy.BaseTaxBCPerPopulation.Value),
		PopulationFoodPerUnit:                         economy.Population.FoodPerPopulation,
		CyberneticFoodPerUnit:                         economy.Population.CyberneticFoodPerPopulation,
		CyberneticProductionPerUnit:                   economy.Population.CyberneticProductionPerPopulation,
		PopulationGrowthCurveFactor:                   economy.Population.GrowthCurveFactor,
		HousingGrowthPercentPerPPPerPopulation:        growthRules.HousingPercentPerProductionPerPopulation,
		HousingGrowthPercentRounding:                  growthRules.HousingPercentRounding,
		CloningCenterBuildingID:                       growthRules.CloningCenterBuildingID,
		CloningCenterFlatGrowth:                       growthRules.CloningCenterFlatGrowth,
		PopulationGrowthTechnologyBonusByID:           populationGrowthTechnologyBonusByID,
		AdvancedCityPlanningTechnologyID:              advancedCityPlanningTechnologyID,
		AdvancedCityPlanningCapacityBonus:             capacityRules.AdvancedCityPlanningFlatCapacity,
		BiospheresBuildingID:                          capacityRules.BiospheresBuildingID,
		BiospheresCapacityBonus:                       capacityRules.BiospheresFlatCapacity,
		PlanetaryTransformations:                      planetaryTransformations,
		TolerantHabitabilityBonus:                     economy.Population.TolerantHabitabilityBonus,
		SubterraneanCapacityPerClass:                  economy.Population.SubterraneanCapacityPerSizeClass,
		PopulationSizeCapacity:                        populationSizeCapacity,
		PopulationSizeClass:                           populationSizeClass,
		PopulationClimateHabitability:                 populationClimateHabitability,
		FreighterFoodCapacity:                         economy.FoodLogistics.FreighterFoodCapacity,
		FreightersPerFleet:                            economy.FoodLogistics.FreightersPerFleet,
		FreighterFleetCostPP:                          economy.FoodLogistics.FreighterFleetCostPP,
		FreighterOperatingCostBC:                      economy.FoodLogistics.FreighterOperatingCostBC,
		SurplusFoodBCPerUnit:                          economy.FoodLogistics.SurplusFoodBCPerUnit,
		FantasticTradersSurplusFoodBCPerUnit:          economy.FoodLogistics.FantasticTradersSurplusFoodBCPerUnit,
		StarvationPopulationPerFoodShortage:           economy.FoodLogistics.StarvationPopulationPerFoodShortage,
		CyberneticStarvationPopulationPerFoodShortage: economy.FoodLogistics.CyberneticStarvationPopulationPerFoodShortage,
		CyberneticStarvationPopulationPerPPShortage:   economy.FoodLogistics.CyberneticStarvationPopulationPerPPShortage,
		MinimumPopulationAfterStarvation:              economy.FoodLogistics.MinimumPopulationAfterStarvation,
		GravityPenaltyPercent:                         gravityPenalties,
		GovernmentModifiers:                           governmentModifiers,
		MoraleBarracksPenaltyPercent:                  economy.Morale.BarracksPenaltyPercent,
		MoraleBarracksGovernments:                     moraleBarracksGovernments,
		MoraleBarracksBuildingIDs:                     moraleBarracksBuildingIDs,
		MoraleBuildingBonusPercent:                    moraleBuildingBonusPercent,
		KnownBuildingIDs:                              buildingIDs,
		BuildingDefinitions:                           buildingDefinitions,
		TechnologyFieldCostsRP:                        technologyFieldCosts,
		TechnologyFieldPreviousID:                     technologyFieldPreviousID,
		TechnologyFieldNextID:                         technologyFieldNextID,
		TechnologyFieldAIGroup:                        technologyFieldAIGroup,
		TechnologyIDsByField:                          technologyIDsByField,
		TechnologyFieldByID:                           technologyFieldByID,
		TechnologyKeyByID:                             technologyKeyByID,
		TechnologyNameKeyByID:                         technologyNameKeyByID,
		TechnologyStrategicAvailable:                  technologyStrategicAvailable,
		TechnologyAIClassByID:                         technologyAIClassByID,
		TechnologyAIClasses:                           technologyAIClasses,
		TechnologyAIFieldGroupValues:                  technologyAIFieldGroupValues,
		NewGameAlwaysKnownFieldID:                     technologies.NewGameStart.AlwaysKnownTechFieldID,
		NewGameStagedKnownFieldIDs:                    append([]int(nil), technologies.NewGameStart.StagedKnownTechFieldIDs...),
		GeneralResearchFieldIDs:                       generalResearchFieldIDs,
		HyperAdvancedResearchFieldIDs:                 hyperAdvancedResearchFieldIDs,
		HyperAdvancedCostIncrementRP:                  float64(technologies.HyperAdvanced.CostIncrementRP),
	}
	for _, climate := range planetClasses.Climates {
		rules.ClimateFoodPerFarmer[climate.ID] = float64(climate.BaseFoodPerFarmer)
	}
	for _, mineral := range planetClasses.MineralClasses {
		value, ok := industryByMineral[mineral.ID]
		if !ok {
			return nil, fmt.Errorf("economy rules missing mineral class %q", mineral.ID)
		}
		rules.MineralIndustryPerWorker[mineral.ID] = float64(value)
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
		modifiers := RaceEconomyModifiers{GravityID: "normal_g", PopulationGrowthMultiplier: 1}
		researchModifiers := RaceResearchModifiers{}
		for _, selection := range race.TraitSelections {
			option, ok := options[selection.TraitID]
			if !ok {
				continue
			}
			switch option.Ability {
			case "aquatic":
				modifiers.Aquatic = true
				researchModifiers.Aquatic = true
			case "tolerant":
				modifiers.Tolerant = true
				researchModifiers.Tolerant = true
			case "subterranean":
				modifiers.Subterranean = true
				researchModifiers.Subterranean = true
			case "cybernetic":
				modifiers.Cybernetic = true
				researchModifiers.Cybernetic = true
			case "lithovore":
				modifiers.Lithovore = true
				researchModifiers.Lithovore = true
			case "fantastic_traders":
				modifiers.FantasticTraders = true
			case "trans_dimensional":
				modifiers.TransDimensional = true
			case "creative":
				modifiers.Creative = true
			case "uncreative":
				modifiers.Uncreative = true
			case "low_g_world":
				modifiers.GravityID = "low_g"
				researchModifiers.LowGWorld = true
			case "high_g_world":
				modifiers.GravityID = "heavy_g"
				researchModifiers.HighGWorld = true
			case "government_feudal", "government_dictatorship", "government_democracy", "government_unification":
				modifiers.GovernmentTraitID = option.Ability
				researchModifiers.GovernmentTraitID = option.Ability
			case "telepathic":
				researchModifiers.Telepathic = true
			case "stealthy_ships":
				researchModifiers.StealthyShips = true
			}
			if option.Value == nil {
				continue
			}
			delta := *option.Value
			switch option.ValueKind {
			case "food_per_farmer_delta":
				modifiers.FoodPerFarmer += delta
				researchModifiers.FarmingDelta = delta
			case "production_per_worker_delta":
				modifiers.ProductionPerWorker += delta
				researchModifiers.IndustryDelta = delta
			case "research_per_scientist_delta":
				modifiers.ResearchPerScientist += delta
				researchModifiers.ScienceDelta = delta
			case "tax_bc_per_population_delta":
				modifiers.TaxBCPerPopulation += delta
				researchModifiers.MoneyDelta = delta
			case "population_growth_multiplier":
				modifiers.PopulationGrowthMultiplier = delta
				researchModifiers.PopulationGrowthPercent = int(math.Round((delta - 1) * 100))
			case "ship_defense_bonus":
				researchModifiers.ShipDefenseBonus = int(math.Round(delta))
			case "ship_attack_bonus":
				researchModifiers.ShipAttackBonus = int(math.Round(delta))
			case "ground_combat_bonus":
				researchModifiers.GroundCombatBonus = int(math.Round(delta))
			case "spying_bonus":
				researchModifiers.SpyingBonus = int(math.Round(delta))
			}
		}
		if modifiers.Creative && modifiers.Uncreative {
			return nil, fmt.Errorf("race %q cannot be both creative and uncreative", race.ID)
		}
		if modifiers.PopulationGrowthMultiplier <= 0 || math.IsNaN(modifiers.PopulationGrowthMultiplier) || math.IsInf(modifiers.PopulationGrowthMultiplier, 0) {
			return nil, fmt.Errorf("race %q has invalid population growth multiplier %g", race.ID, modifiers.PopulationGrowthMultiplier)
		}
		if modifiers.GovernmentTraitID == "" {
			return nil, fmt.Errorf("race %q has no normalized starting government", race.ID)
		}
		if _, ok := governmentModifiers[modifiers.GovernmentTraitID]; !ok {
			return nil, fmt.Errorf("race %q uses unsupported government %q", race.ID, modifiers.GovernmentTraitID)
		}
		rules.RaceModifiers[race.ID] = modifiers
		rules.RaceResearchModifiers[race.ID] = researchModifiers
	}
	return rules, nil
}

func (r *EconomyRules) CalculateBaseEconomy(colony core.Colony, planet core.Planet, raceID string) (core.ColonyEconomy, error) {
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
	taxPerPopulation := math.Max(0, r.BaseTaxBCPerPopulation+modifiers.TaxBCPerPopulation)
	population := colony.Population
	return core.ColonyEconomy{
		Food:       population.Farmers * foodPerFarmer,
		Production: population.Workers * productionPerWorker,
		Research:   population.Scientists * researchPerScientist,
		TaxBC:      population.Total * taxPerPopulation,
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
		Food:       adjustedRoleOutput(base.Food, government.FoodPercent, moralePercent, gravityPenalty),
		Production: adjustedRoleOutput(base.Production, government.ProductionPercent, moralePercent, gravityPenalty),
		Research:   adjustedRoleOutput(base.Research, government.ResearchPercent, moralePercent, gravityPenalty),
		TaxBC:      adjustedTaxIncome(base.TaxBC, government, moralePercent),
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
func adjustedRoleOutput(base float64, governmentPercent, moralePercent, gravityPenaltyPercent int) float64 {
	if base <= 0 {
		return 0
	}
	percent := 100 + governmentPercent + moralePercent - gravityPenaltyPercent
	if percent <= 0 {
		return 0
	}
	return base * float64(percent) / 100
}

func adjustedTaxIncome(base float64, government GovernmentEconomyModifier, moralePercent int) float64 {
	if base <= 0 {
		return 0
	}
	governmentBonus := base * float64(government.TaxPercent) / 100
	if government.TaxPercent != 0 {
		switch government.TaxBonusRounding {
		case "down":
			governmentBonus = math.Floor(governmentBonus)
		case "nearest":
			governmentBonus = math.Round(governmentBonus)
		}
	}
	moraleBonus := base * float64(moralePercent) / 100
	return math.Max(0, base+governmentBonus+moraleBonus)
}

func gravityKey(raceGravityID, planetGravityID string) string {
	return raceGravityID + "/" + planetGravityID
}
