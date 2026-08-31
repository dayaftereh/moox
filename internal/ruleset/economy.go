package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const EconomySchemaVersion = 8

type EconomyFile struct {
	SchemaVersion            int                        `json:"schema_version"`
	Ruleset                  string                     `json:"ruleset"`
	Sources                  []Source                   `json:"sources"`
	BaseResearchPerScientist EconomyScalar              `json:"base_research_per_scientist"`
	BaseTaxBCPerPopulation   EconomyScalar              `json:"base_tax_bc_per_population"`
	MineralIndustryPerWorker []MineralIndustryPerWorker `json:"mineral_industry_per_worker"`
	AquaticFoodBonus         EconomyClimateBonus        `json:"aquatic_food_bonus"`
	GravityPenalties         []GravityPenaltyRule       `json:"gravity_penalties"`
	GovernmentModifiers      []GovernmentEconomyRule    `json:"government_modifiers"`
	Population               PopulationEconomyRule      `json:"population"`
	FoodLogistics            FoodLogisticsEconomyRule   `json:"food_logistics"`
	CommandPoints            CommandPointsEconomyRule   `json:"command_points"`
	Morale                   MoraleEconomyRule          `json:"morale"`
}

type EconomyScalar struct {
	Value    int    `json:"value"`
	SourceID string `json:"source_id"`
}

type MineralIndustryPerWorker struct {
	MineralID string `json:"mineral_id"`
	Value     int    `json:"value"`
	SourceID  string `json:"source_id"`
}

type EconomyClimateBonus struct {
	Value      int      `json:"value"`
	ClimateIDs []string `json:"climate_ids"`
	SourceID   string   `json:"source_id"`
}

type PopulationSizeCapacity struct {
	SizeID   string  `json:"size_id"`
	Capacity float64 `json:"capacity"`
}

type PopulationClimateHabitability struct {
	ClimateID string  `json:"climate_id"`
	Fraction  float64 `json:"fraction"`
}

type PopulationEconomyRule struct {
	FoodPerPopulation                 float64                         `json:"food_per_population"`
	CyberneticFoodPerPopulation       float64                         `json:"cybernetic_food_per_population"`
	CyberneticProductionPerPopulation float64                         `json:"cybernetic_production_per_population"`
	GrowthCurveFactor                 float64                         `json:"growth_curve_factor"`
	GrowthModifiers                   PopulationGrowthModifierRule    `json:"growth_modifiers"`
	CapacityModifiers                 PopulationCapacityModifierRule  `json:"capacity_modifiers"`
	TolerantHabitabilityBonus         float64                         `json:"tolerant_habitability_bonus"`
	SubterraneanCapacityPerSizeClass  float64                         `json:"subterranean_capacity_per_size_class"`
	SizeCapacity                      []PopulationSizeCapacity        `json:"size_capacity"`
	ClimateHabitability               []PopulationClimateHabitability `json:"climate_habitability"`
	SourceIDs                         []string                        `json:"source_ids"`
}

type PopulationGrowthModifierRule struct {
	HousingPercentPerProductionPerPopulation float64                           `json:"housing_percent_per_production_per_population"`
	HousingPercentRounding                   string                            `json:"housing_percent_rounding"`
	CloningCenterBuildingID                  string                            `json:"cloning_center_building_id"`
	CloningCenterFlatGrowth                  float64                           `json:"cloning_center_flat_growth"`
	TechnologyBonuses                        []PopulationGrowthTechnologyBonus `json:"technology_bonuses"`
	SourceIDs                                []string                          `json:"source_ids"`
}

type PopulationGrowthTechnologyBonus struct {
	TechnologyKey string  `json:"technology_key"`
	Bonus         float64 `json:"bonus"`
}

type PopulationCapacityModifierRule struct {
	AdvancedCityPlanningTechnologyKey string                        `json:"advanced_city_planning_technology_key"`
	AdvancedCityPlanningFlatCapacity  float64                       `json:"advanced_city_planning_flat_capacity"`
	BiospheresBuildingID              string                        `json:"biospheres_building_id"`
	BiospheresFlatCapacity            float64                       `json:"biospheres_flat_capacity"`
	PlanetaryTransformations          []PlanetaryTransformationRule `json:"planetary_transformations"`
	SourceIDs                         []string                      `json:"source_ids"`
}

type PlanetaryTransformationRule struct {
	ProjectID             string                       `json:"project_id"`
	AllowedClimateIDs     []string                     `json:"allowed_climate_ids"`
	ResultClimateBySource map[string]string            `json:"result_climate_by_source"`
	BarrenOrbitRule       *BarrenTerraformingOrbitRule `json:"barren_orbit_rule,omitempty"`
	SourceIDs             []string                     `json:"source_ids"`
}

type BarrenTerraformingOrbitRule struct {
	InnerOrbitMax int      `json:"inner_orbit_max"`
	MiddleOrbit   int      `json:"middle_orbit"`
	OuterOrbitMin int      `json:"outer_orbit_min"`
	InnerResult   string   `json:"inner_result"`
	OuterResult   string   `json:"outer_result"`
	MiddleResults []string `json:"middle_results"`
}

type FoodLogisticsEconomyRule struct {
	FreighterFoodCapacity                         float64  `json:"freighter_food_capacity"`
	FreightersPerFleet                            int      `json:"freighters_per_fleet"`
	FreighterFleetCostPP                          float64  `json:"freighter_fleet_cost_pp"`
	FreighterOperatingCostBC                      float64  `json:"freighter_operating_cost_bc"`
	SurplusFoodBCPerUnit                          float64  `json:"surplus_food_bc_per_unit"`
	FantasticTradersSurplusFoodBCPerUnit          float64  `json:"fantastic_traders_surplus_food_bc_per_unit"`
	StarvationPopulationPerFoodShortage           float64  `json:"starvation_population_per_food_shortage"`
	CyberneticStarvationPopulationPerFoodShortage float64  `json:"cybernetic_starvation_population_per_food_shortage"`
	CyberneticStarvationPopulationPerPPShortage   float64  `json:"cybernetic_starvation_population_per_pp_shortage"`
	MinimumPopulationAfterStarvation              float64  `json:"minimum_population_after_starvation"`
	SourceIDs                                     []string `json:"source_ids"`
}

type CommandPointsEconomyRule struct {
	BaseCapacity              int                              `json:"base_capacity"`
	FixedSpecialShipPoints    int                              `json:"fixed_special_ship_points"`
	WarlordPointsPerColony    int                              `json:"warlord_points_per_colony"`
	StandardOverageBCPerPoint float64                          `json:"standard_overage_bc_per_point"`
	Stations                  []CommandPointStationRule        `json:"stations"`
	Communications            []CommandPointCommunicationsRule `json:"communications"`
	Imperium                  CommandPointImperiumRule         `json:"imperium"`
	SourceIDs                 []string                         `json:"source_ids"`
}

type CommandPointStationRule struct {
	BuildingID string `json:"building_id"`
	Points     int    `json:"points"`
}

type CommandPointCommunicationsRule struct {
	TechnologyID int `json:"technology_id"`
	Points       int `json:"points_per_station"`
}

type CommandPointImperiumRule struct {
	TechnologyID              int    `json:"technology_id"`
	RequiredGovernmentTraitID string `json:"required_government_trait_id"`
	BonusNumerator            int    `json:"bonus_numerator"`
	BonusDenominator          int    `json:"bonus_denominator"`
	Rounding                  string `json:"rounding"`
}

type GravityPenaltyRule struct {
	RaceGravityID   string   `json:"race_gravity_id"`
	PlanetGravityID string   `json:"planet_gravity_id"`
	Percent         int      `json:"percent"`
	SourceIDs       []string `json:"source_ids"`
}

type GovernmentEconomyRule struct {
	TraitID           string `json:"trait_id"`
	FoodPercent       int    `json:"food_percent"`
	ProductionPercent int    `json:"production_percent"`
	ResearchPercent   int    `json:"research_percent"`
	TaxPercent        int    `json:"tax_percent"`
	TaxBonusRounding  string `json:"tax_bonus_rounding"`
	IgnoresMorale     bool   `json:"ignores_morale"`
	SourceID          string `json:"source_id"`
}

type MoraleEconomyRule struct {
	BarracksPenaltyPercent     int                   `json:"barracks_penalty_percent"`
	BarracksGovernmentTraitIDs []string              `json:"barracks_government_trait_ids"`
	BarracksBuildingIDs        []string              `json:"barracks_building_ids"`
	BuildingBonuses            []MoraleBuildingBonus `json:"building_bonuses"`
	SourceID                   string                `json:"source_id"`
}

type MoraleBuildingBonus struct {
	BuildingID string `json:"building_id"`
	Percent    int    `json:"percent"`
}

func LoadEconomy(path string) (*EconomyFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file EconomyFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	if err := file.Validate(); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *EconomyFile) Validate() error {
	if f.SchemaVersion != EconomySchemaVersion {
		return fmt.Errorf("unsupported economy schema version %d", f.SchemaVersion)
	}
	if f.Ruleset == "" {
		return fmt.Errorf("ruleset is required")
	}
	sources := make(map[string]struct{}, len(f.Sources))
	for _, source := range f.Sources {
		if source.ID == "" || source.Type == "" {
			return fmt.Errorf("economy source id and type are required")
		}
		if _, exists := sources[source.ID]; exists {
			return fmt.Errorf("duplicate economy source %q", source.ID)
		}
		sources[source.ID] = struct{}{}
	}
	validateSource := func(name, sourceID string) error {
		if _, ok := sources[sourceID]; !ok {
			return fmt.Errorf("%s references unknown source %q", name, sourceID)
		}
		return nil
	}
	validateScalar := func(name string, scalar EconomyScalar, min int) error {
		if scalar.Value < min {
			return fmt.Errorf("%s value %d below %d", name, scalar.Value, min)
		}
		return validateSource(name, scalar.SourceID)
	}
	if err := validateScalar("base_research_per_scientist", f.BaseResearchPerScientist, 1); err != nil {
		return err
	}
	if err := validateScalar("base_tax_bc_per_population", f.BaseTaxBCPerPopulation, 0); err != nil {
		return err
	}
	if len(f.MineralIndustryPerWorker) != 5 {
		return fmt.Errorf("mineral industry count=%d, expected 5", len(f.MineralIndustryPerWorker))
	}
	seenMinerals := make(map[string]struct{}, len(f.MineralIndustryPerWorker))
	for i, item := range f.MineralIndustryPerWorker {
		if item.MineralID == "" || item.Value < 1 {
			return fmt.Errorf("mineral_industry_per_worker[%d] has invalid id/value", i)
		}
		if _, exists := seenMinerals[item.MineralID]; exists {
			return fmt.Errorf("duplicate mineral industry id %q", item.MineralID)
		}
		seenMinerals[item.MineralID] = struct{}{}
		if err := validateSource("mineral industry "+item.MineralID, item.SourceID); err != nil {
			return err
		}
	}
	if f.AquaticFoodBonus.Value <= 0 || len(f.AquaticFoodBonus.ClimateIDs) == 0 {
		return fmt.Errorf("aquatic_food_bonus requires positive value and climates")
	}
	if err := validateSource("aquatic_food_bonus", f.AquaticFoodBonus.SourceID); err != nil {
		return err
	}
	climates := make(map[string]struct{}, len(f.AquaticFoodBonus.ClimateIDs))
	for _, climateID := range f.AquaticFoodBonus.ClimateIDs {
		if climateID == "" {
			return fmt.Errorf("aquatic_food_bonus has empty climate id")
		}
		if _, exists := climates[climateID]; exists {
			return fmt.Errorf("aquatic_food_bonus has duplicate climate %q", climateID)
		}
		climates[climateID] = struct{}{}
	}
	if err := f.validatePopulation(sources); err != nil {
		return err
	}
	if err := f.validateFoodLogistics(sources); err != nil {
		return err
	}
	if err := f.validateCommandPoints(sources); err != nil {
		return err
	}
	if len(f.GravityPenalties) != 9 {
		return fmt.Errorf("gravity penalty count=%d, expected 9", len(f.GravityPenalties))
	}
	seenGravity := make(map[string]struct{}, len(f.GravityPenalties))
	for i, rule := range f.GravityPenalties {
		if rule.RaceGravityID == "" || rule.PlanetGravityID == "" || rule.Percent < 0 || rule.Percent > 100 {
			return fmt.Errorf("gravity_penalties[%d] has invalid ids/percent", i)
		}
		key := rule.RaceGravityID + "/" + rule.PlanetGravityID
		if _, exists := seenGravity[key]; exists {
			return fmt.Errorf("duplicate gravity penalty %q", key)
		}
		seenGravity[key] = struct{}{}
		if len(rule.SourceIDs) == 0 {
			return fmt.Errorf("gravity penalty %q has no sources", key)
		}
		for _, sourceID := range rule.SourceIDs {
			if err := validateSource("gravity penalty "+key, sourceID); err != nil {
				return err
			}
		}
	}
	if len(f.GovernmentModifiers) != 4 {
		return fmt.Errorf("government modifier count=%d, expected 4", len(f.GovernmentModifiers))
	}
	seenGovernments := make(map[string]struct{}, len(f.GovernmentModifiers))
	for i, rule := range f.GovernmentModifiers {
		if rule.TraitID == "" {
			return fmt.Errorf("government_modifiers[%d] missing trait_id", i)
		}
		if _, exists := seenGovernments[rule.TraitID]; exists {
			return fmt.Errorf("duplicate government modifier %q", rule.TraitID)
		}
		seenGovernments[rule.TraitID] = struct{}{}
		for _, percent := range []int{rule.FoodPercent, rule.ProductionPercent, rule.ResearchPercent, rule.TaxPercent} {
			if percent < -100 || percent > 200 {
				return fmt.Errorf("government modifier %q has out-of-range percent %d", rule.TraitID, percent)
			}
		}
		if rule.TaxBonusRounding != "none" && rule.TaxBonusRounding != "down" {
			return fmt.Errorf("government modifier %q has unsupported tax rounding %q", rule.TraitID, rule.TaxBonusRounding)
		}
		if err := validateSource("government modifier "+rule.TraitID, rule.SourceID); err != nil {
			return err
		}
	}
	if err := f.validateMorale(sources, seenGovernments); err != nil {
		return err
	}
	return nil
}

func (f *EconomyFile) validatePopulation(sources map[string]struct{}) error {
	p := f.Population
	if p.FoodPerPopulation <= 0 || p.CyberneticFoodPerPopulation < 0 || p.CyberneticProductionPerPopulation < 0 || p.GrowthCurveFactor <= 0 || p.TolerantHabitabilityBonus < 0 || p.SubterraneanCapacityPerSizeClass < 0 {
		return fmt.Errorf("population economy scalar values are invalid")
	}
	growth := p.GrowthModifiers
	if growth.HousingPercentPerProductionPerPopulation <= 0 || growth.HousingPercentRounding != "down" || growth.CloningCenterBuildingID == "" || growth.CloningCenterFlatGrowth <= 0 {
		return fmt.Errorf("population growth modifier scalars are invalid")
	}
	if len(growth.TechnologyBonuses) == 0 || len(growth.SourceIDs) == 0 {
		return fmt.Errorf("population growth modifiers require technology bonuses and sources")
	}
	seenGrowthTechnologies := make(map[string]struct{}, len(growth.TechnologyBonuses))
	for i, bonus := range growth.TechnologyBonuses {
		if bonus.TechnologyKey == "" || bonus.Bonus <= 0 {
			return fmt.Errorf("population growth technology_bonuses[%d] is invalid", i)
		}
		if _, exists := seenGrowthTechnologies[bonus.TechnologyKey]; exists {
			return fmt.Errorf("duplicate population growth technology %q", bonus.TechnologyKey)
		}
		seenGrowthTechnologies[bonus.TechnologyKey] = struct{}{}
	}
	for _, sourceID := range growth.SourceIDs {
		if _, ok := sources[sourceID]; !ok {
			return fmt.Errorf("population growth modifiers reference unknown source %q", sourceID)
		}
	}
	capacity := p.CapacityModifiers
	if capacity.AdvancedCityPlanningTechnologyKey == "" || capacity.AdvancedCityPlanningFlatCapacity <= 0 || capacity.BiospheresBuildingID == "" || capacity.BiospheresFlatCapacity <= 0 {
		return fmt.Errorf("population capacity modifier scalars are invalid")
	}
	if len(capacity.PlanetaryTransformations) == 0 || len(capacity.SourceIDs) == 0 {
		return fmt.Errorf("population capacity modifiers require transformations and sources")
	}
	seenTransformations := make(map[string]struct{}, len(capacity.PlanetaryTransformations))
	for i, transformation := range capacity.PlanetaryTransformations {
		if transformation.ProjectID == "" || len(transformation.AllowedClimateIDs) == 0 || len(transformation.SourceIDs) == 0 {
			return fmt.Errorf("population planetary_transformations[%d] is invalid", i)
		}
		if _, exists := seenTransformations[transformation.ProjectID]; exists {
			return fmt.Errorf("duplicate population planetary transformation %q", transformation.ProjectID)
		}
		seenTransformations[transformation.ProjectID] = struct{}{}
		if len(transformation.ResultClimateBySource) == 0 && transformation.BarrenOrbitRule == nil {
			return fmt.Errorf("population planetary transformation %q has no results", transformation.ProjectID)
		}
		if rule := transformation.BarrenOrbitRule; rule != nil {
			if rule.InnerOrbitMax <= 0 || rule.MiddleOrbit <= rule.InnerOrbitMax || rule.OuterOrbitMin <= rule.MiddleOrbit || rule.InnerResult == "" || rule.OuterResult == "" || len(rule.MiddleResults) < 2 {
				return fmt.Errorf("population planetary transformation %q has invalid barren orbit rule", transformation.ProjectID)
			}
		}
		for _, sourceID := range transformation.SourceIDs {
			if _, ok := sources[sourceID]; !ok {
				return fmt.Errorf("population planetary transformation %q references unknown source %q", transformation.ProjectID, sourceID)
			}
		}
	}
	for _, sourceID := range capacity.SourceIDs {
		if _, ok := sources[sourceID]; !ok {
			return fmt.Errorf("population capacity modifiers reference unknown source %q", sourceID)
		}
	}
	if len(p.SizeCapacity) != 5 || len(p.ClimateHabitability) != 10 {
		return fmt.Errorf("population size/climate rule counts=%d/%d, expected 5/10", len(p.SizeCapacity), len(p.ClimateHabitability))
	}
	seenSizes := make(map[string]struct{}, len(p.SizeCapacity))
	for i, item := range p.SizeCapacity {
		if item.SizeID == "" || item.Capacity <= 0 {
			return fmt.Errorf("population size_capacity[%d] is invalid", i)
		}
		if _, exists := seenSizes[item.SizeID]; exists {
			return fmt.Errorf("duplicate population size capacity %q", item.SizeID)
		}
		seenSizes[item.SizeID] = struct{}{}
	}
	seenClimates := make(map[string]struct{}, len(p.ClimateHabitability))
	for i, item := range p.ClimateHabitability {
		if item.ClimateID == "" || item.Fraction <= 0 || item.Fraction > 1 {
			return fmt.Errorf("population climate_habitability[%d] is invalid", i)
		}
		if _, exists := seenClimates[item.ClimateID]; exists {
			return fmt.Errorf("duplicate population climate habitability %q", item.ClimateID)
		}
		seenClimates[item.ClimateID] = struct{}{}
	}
	if len(p.SourceIDs) == 0 {
		return fmt.Errorf("population economy rule has no sources")
	}
	for _, sourceID := range p.SourceIDs {
		if _, ok := sources[sourceID]; !ok {
			return fmt.Errorf("population economy rule references unknown source %q", sourceID)
		}
	}
	return nil
}

func (f *EconomyFile) validateFoodLogistics(sources map[string]struct{}) error {
	l := f.FoodLogistics
	if l.FreighterFoodCapacity <= 0 || l.FreightersPerFleet <= 0 || l.FreighterFleetCostPP <= 0 || l.FreighterOperatingCostBC < 0 || l.SurplusFoodBCPerUnit < 0 || l.FantasticTradersSurplusFoodBCPerUnit < l.SurplusFoodBCPerUnit || l.StarvationPopulationPerFoodShortage <= 0 || l.CyberneticStarvationPopulationPerFoodShortage <= 0 || l.CyberneticStarvationPopulationPerPPShortage <= 0 || l.MinimumPopulationAfterStarvation <= 0 {
		return fmt.Errorf("food logistics scalar values are invalid")
	}
	if len(l.SourceIDs) == 0 {
		return fmt.Errorf("food logistics rule has no sources")
	}
	for _, sourceID := range l.SourceIDs {
		if _, ok := sources[sourceID]; !ok {
			return fmt.Errorf("food logistics rule references unknown source %q", sourceID)
		}
	}
	return nil
}

func (f *EconomyFile) validateCommandPoints(sources map[string]struct{}) error {
	rule := f.CommandPoints
	if rule.BaseCapacity <= 0 || rule.FixedSpecialShipPoints <= 0 || rule.WarlordPointsPerColony <= 0 || rule.StandardOverageBCPerPoint < 0 {
		return fmt.Errorf("command points scalar values are invalid")
	}
	if len(rule.Stations) == 0 || len(rule.Communications) == 0 {
		return fmt.Errorf("command points station/communications rules are required")
	}
	seenStations := make(map[string]struct{}, len(rule.Stations))
	for index, station := range rule.Stations {
		if station.BuildingID == "" || station.Points <= 0 {
			return fmt.Errorf("command_points.stations[%d] is invalid", index)
		}
		if _, exists := seenStations[station.BuildingID]; exists {
			return fmt.Errorf("duplicate command point station %q", station.BuildingID)
		}
		seenStations[station.BuildingID] = struct{}{}
	}
	seenTech := make(map[int]struct{}, len(rule.Communications))
	for index, communication := range rule.Communications {
		if communication.TechnologyID <= 0 || communication.Points <= 0 {
			return fmt.Errorf("command_points.communications[%d] is invalid", index)
		}
		if _, exists := seenTech[communication.TechnologyID]; exists {
			return fmt.Errorf("duplicate command point communications technology %d", communication.TechnologyID)
		}
		seenTech[communication.TechnologyID] = struct{}{}
	}
	imperium := rule.Imperium
	if imperium.TechnologyID <= 0 || imperium.RequiredGovernmentTraitID == "" || imperium.BonusNumerator <= 0 || imperium.BonusDenominator <= 0 || imperium.Rounding != "floor" {
		return fmt.Errorf("command points imperium rule is invalid")
	}
	if len(rule.SourceIDs) == 0 {
		return fmt.Errorf("command points rule has no sources")
	}
	for _, sourceID := range rule.SourceIDs {
		if _, ok := sources[sourceID]; !ok {
			return fmt.Errorf("command points rule references unknown source %q", sourceID)
		}
	}
	return nil
}

func (f *EconomyFile) validateMorale(sources, governments map[string]struct{}) error {
	if f.Morale.BarracksPenaltyPercent >= 0 || f.Morale.BarracksPenaltyPercent < -100 {
		return fmt.Errorf("morale barracks penalty must be negative and >= -100, got %d", f.Morale.BarracksPenaltyPercent)
	}
	if _, ok := sources[f.Morale.SourceID]; !ok {
		return fmt.Errorf("morale references unknown source %q", f.Morale.SourceID)
	}
	if len(f.Morale.BarracksGovernmentTraitIDs) == 0 || len(f.Morale.BarracksBuildingIDs) == 0 {
		return fmt.Errorf("morale barracks rules require governments and building ids")
	}
	seenGovernments := make(map[string]struct{}, len(f.Morale.BarracksGovernmentTraitIDs))
	for _, traitID := range f.Morale.BarracksGovernmentTraitIDs {
		if _, ok := governments[traitID]; !ok {
			return fmt.Errorf("morale barracks rule references unknown government %q", traitID)
		}
		if _, exists := seenGovernments[traitID]; exists {
			return fmt.Errorf("duplicate morale barracks government %q", traitID)
		}
		seenGovernments[traitID] = struct{}{}
	}
	seenBuildings := make(map[string]struct{})
	for _, buildingID := range f.Morale.BarracksBuildingIDs {
		if buildingID == "" {
			return fmt.Errorf("morale barracks building id must not be empty")
		}
		if _, exists := seenBuildings[buildingID]; exists {
			return fmt.Errorf("duplicate morale building %q", buildingID)
		}
		seenBuildings[buildingID] = struct{}{}
	}
	for _, bonus := range f.Morale.BuildingBonuses {
		if bonus.BuildingID == "" || bonus.Percent <= 0 || bonus.Percent > 100 {
			return fmt.Errorf("invalid morale building bonus %+v", bonus)
		}
		if _, exists := seenBuildings[bonus.BuildingID]; exists {
			return fmt.Errorf("duplicate morale building %q", bonus.BuildingID)
		}
		seenBuildings[bonus.BuildingID] = struct{}{}
	}
	return nil
}
