package core

import (
	"fmt"
	"math"
)

const StateSchemaVersion = 20

type ID uint64

type GameState struct {
	SchemaVersion       int                  `json:"schema_version"`
	Seed                uint64               `json:"seed"`
	RNGState            uint64               `json:"rng_state"`
	Turn                uint64               `json:"turn"`
	NextID              ID                   `json:"next_id"`
	Galaxy              Galaxy               `json:"galaxy"`
	Empires             []Empire             `json:"empires"`
	Colonies            []Colony             `json:"colonies"`
	Outposts            []Outpost            `json:"outposts,omitempty"`
	ShipDesigns         []ShipDesign         `json:"ship_designs,omitempty"`
	Ships               []Ship               `json:"ships,omitempty"`
	StrategicFleets     []StrategicFleet     `json:"strategic_fleets,omitempty"`
	DiplomaticRelations []DiplomaticRelation `json:"diplomatic_relations,omitempty"`
	PopulationTransfers []PopulationTransfer `json:"population_transfers,omitempty"`
	Events              []Event              `json:"events"`
}

type Galaxy struct {
	ID      ID           `json:"id"`
	Systems []StarSystem `json:"systems"`
}

type StarSystem struct {
	ID                 ID       `json:"id"`
	Name               string   `json:"name"`
	X                  int      `json:"x"`
	Y                  int      `json:"y"`
	BlockadedEmpireIDs []ID     `json:"blockaded_empire_ids,omitempty"`
	Planets            []Planet `json:"planets"`
}

type Planet struct {
	ID        ID     `json:"id"`
	Name      string `json:"name"`
	Orbit     int    `json:"orbit"`
	SizeID    string `json:"size_id"`
	MineralID string `json:"mineral_id"`
	GravityID string `json:"gravity_id"`
	ClimateID string `json:"climate_id"`
	ColonyID  ID     `json:"colony_id,omitempty"`
	OutpostID ID     `json:"outpost_id,omitempty"`
}

type Outpost struct {
	ID       ID `json:"id"`
	EmpireID ID `json:"empire_id"`
	PlanetID ID `json:"planet_id"`
}

type Empire struct {
	ID                        ID                           `json:"id"`
	Name                      string                       `json:"name"`
	RaceID                    string                       `json:"race_id"`
	Capital                   ID                           `json:"capital_colony_id,omitempty"`
	Freighters                int                          `json:"freighters"`
	CommandPoints             EmpireCommandPoints          `json:"command_points"`
	Treasury                  EmpireTreasuryState          `json:"treasury"`
	FoodLogistics             EmpireFoodLogistics          `json:"food_logistics"`
	UncreativeResearchChoices []FixedResearchChoice        `json:"uncreative_research_choices,omitempty"`
	HyperAdvancedResearch     []HyperAdvancedResearchLevel `json:"hyper_advanced_research,omitempty"`
	KnownTechnologyIDs        []int                        `json:"known_technology_ids,omitempty"`
	KnownTechnologyFieldIDs   []int                        `json:"known_technology_field_ids,omitempty"`
	Research                  *ResearchState               `json:"research,omitempty"`
}

type FixedResearchChoice struct {
	TechFieldID  int `json:"tech_field_id"`
	TechnologyID int `json:"technology_id"`
}

type HyperAdvancedResearchLevel struct {
	TechFieldID     int `json:"tech_field_id"`
	CompletedLevels int `json:"completed_levels"`
}

type EmpireCommandPoints struct {
	Capacity int `json:"capacity"`
	Used     int `json:"used"`
}

type EmpireTreasuryState struct {
	BalanceBC                 float64 `json:"balance_bc"`
	TaxIncomeBC               float64 `json:"tax_income_bc"`
	SurplusFoodIncomeBC       float64 `json:"surplus_food_income_bc"`
	GrossIncomeBC             float64 `json:"gross_income_bc"`
	BuildingMaintenanceBC     float64 `json:"building_maintenance_bc"`
	FreighterOperatingCostBC  float64 `json:"freighter_operating_cost_bc"`
	ShipCommandMaintenanceBC  float64 `json:"ship_command_maintenance_bc"`
	TotalModeledMaintenanceBC float64 `json:"total_modeled_maintenance_bc"`
	NetModeledIncomeBC        float64 `json:"net_modeled_income_bc"`
}
type EmpireFoodLogistics struct {
	FreightersRequired                    int     `json:"freighters_required"`
	FreightersUsed                        int     `json:"freighters_used"`
	PopulationTransportFreightersReserved int     `json:"population_transport_freighters_reserved"`
	FreightersAvailableForFood            int     `json:"freighters_available_for_food"`
	LocalFoodSurplus                      float64 `json:"local_food_surplus"`
	LocalFoodShortage                     float64 `json:"local_food_shortage"`
	BlockedFoodSurplus                    float64 `json:"blocked_food_surplus"`
	BlockedFoodShortage                   float64 `json:"blocked_food_shortage"`
	FoodTransferred                       float64 `json:"food_transferred"`
	FoodUnmet                             float64 `json:"food_unmet"`
	SurplusFoodSold                       float64 `json:"surplus_food_sold"`
	FreighterOperatingCostBC              float64 `json:"freighter_operating_cost_bc"`
	SurplusFoodIncomeBC                   float64 `json:"surplus_food_income_bc"`
}

type ResearchSelectionMode string

const (
	ResearchSelectionAll         ResearchSelectionMode = "all"
	ResearchSelectionChooseOne   ResearchSelectionMode = "choose_one"
	ResearchSelectionFixedOne    ResearchSelectionMode = "fixed_one"
	ResearchSelectionRepeatField ResearchSelectionMode = "repeat_field"
)

type ResearchState struct {
	TechFieldID   int                   `json:"tech_field_id"`
	SelectionMode ResearchSelectionMode `json:"selection_mode"`
	TechnologyIDs []int                 `json:"technology_ids"`
	ProgressRP    float64               `json:"progress_rp"`
}
type Colony struct {
	ID                 ID                       `json:"id"`
	EmpireID           ID                       `json:"empire_id"`
	PlanetID           ID                       `json:"planet_id"`
	Population         PopulationState          `json:"population"`
	Buildings          []string                 `json:"buildings,omitempty"`
	Economy            ColonyEconomy            `json:"economy"`
	EconomyContext     ColonyEconomyContext     `json:"economy_context"`
	AdjustedEconomy    ColonyEconomy            `json:"adjusted_economy"`
	PopulationDynamics ColonyPopulationDynamics `json:"population_dynamics"`
	Construction       *ConstructionState       `json:"construction,omitempty"`
}

type ConstructionProjectKind string

const (
	ConstructionProjectBuilding                ConstructionProjectKind = "building"
	ConstructionProjectColonyShip              ConstructionProjectKind = "colony_ship"
	ConstructionProjectOutpostShip             ConstructionProjectKind = "outpost_ship"
	ConstructionProjectMilitaryShip            ConstructionProjectKind = "military_ship"
	ConstructionProjectFreighterFleet          ConstructionProjectKind = "freighter_fleet"
	ConstructionProjectHousing                 ConstructionProjectKind = "housing"
	ConstructionProjectPlanetaryTransformation ConstructionProjectKind = "planetary_transformation"
)

type ConstructionState struct {
	ProjectKind        ConstructionProjectKind `json:"project_kind"`
	ProjectID          string                  `json:"project_id"`
	ProgressPP         float64                 `json:"progress_pp"`
	ShipDesignID       ID                      `json:"ship_design_id,omitempty"`
	ShipDesignRevision uint32                  `json:"ship_design_revision,omitempty"`
}

// ColonyEconomy stores domain-native continuous resource output. Values are
// expressed directly as Food, PP, RP and BC rather than scaled integer units.
type ColonyEconomy struct {
	Food       float64 `json:"food"`
	Production float64 `json:"production"`
	Research   float64 `json:"research"`
	TaxBC      float64 `json:"tax_bc"`
}

// ColonyPopulationDynamics stores the materialized per-turn capacity,
// sustenance, available-production and projected-growth values derived from
// authoritative Colony, Planet and Race state.
type PopulationOriginDynamics struct {
	OriginEmpireID      ID      `json:"origin_empire_id"`
	Capacity            float64 `json:"capacity"`
	Population          float64 `json:"population"`
	Food2Required       float64 `json:"food2_required"`
	BaseGrowth          float64 `json:"base_growth"`
	GrowthMultiplier    float64 `json:"growth_multiplier"`
	ProjectedGrowth     float64 `json:"projected_growth"`
	ProjectedStarvation float64 `json:"projected_starvation"`
}

type ColonyPopulationDynamics struct {
	Capacity                float64                    `json:"capacity"`
	FoodRequired            float64                    `json:"food_required"`
	OwnerOriginFood2        float64                    `json:"owner_origin_food2"`
	AssimilatedForeignFood2 float64                    `json:"assimilated_foreign_food2"`
	ConqueredForeignFood2   float64                    `json:"conquered_foreign_food2"`
	RemainingFood2          float64                    `json:"remaining_food2"`
	WholeFoodRequired       float64                    `json:"whole_food_required"`
	LocalFoodSurplus        float64                    `json:"local_food_surplus"`
	LocalFoodShortage       float64                    `json:"local_food_shortage"`
	FoodImported            float64                    `json:"food_imported"`
	FoodExported            float64                    `json:"food_exported"`
	FoodSurplus             float64                    `json:"food_surplus"`
	FoodShortage            float64                    `json:"food_shortage"`
	ProductionRequired      float64                    `json:"production_required"`
	ProductionShortage      float64                    `json:"production_shortage"`
	ProductionAvailable     float64                    `json:"production_available"`
	BaseGrowth              float64                    `json:"base_growth"`
	GrowthMultiplier        float64                    `json:"growth_multiplier"`
	ProjectedGrowth         float64                    `json:"projected_growth"`
	ProjectedStarvation     float64                    `json:"projected_starvation"`
	Origins                 []PopulationOriginDynamics `json:"origins,omitempty"`
}

// ColonyEconomyContext records the currently implemented percentage layers used
// to recompute AdjustedEconomy from Economy. Future morale/leader modifiers must
// be added to this context and recalculated from the base snapshot, never
// multiplied sequentially onto already-adjusted output.
type ColonyEconomyContext struct {
	RaceGravityID                string `json:"race_gravity_id,omitempty"`
	PlanetGravityID              string `json:"planet_gravity_id,omitempty"`
	GravityPenaltyPercent        int    `json:"gravity_penalty_percent"`
	GovernmentTraitID            string `json:"government_trait_id,omitempty"`
	GovernmentFoodPercent        int    `json:"government_food_percent"`
	GovernmentProductionPercent  int    `json:"government_production_percent"`
	GovernmentResearchPercent    int    `json:"government_research_percent"`
	GovernmentTaxPercent         int    `json:"government_tax_percent"`
	GovernmentIgnoresMorale      bool   `json:"government_ignores_morale"`
	MoraleBarracksPenaltyPercent int    `json:"morale_barracks_penalty_percent"`
	MoraleBuildingBonusPercent   int    `json:"morale_building_bonus_percent"`
	MoralePercent                int    `json:"morale_percent"`
}

type Event struct {
	Turn    uint64 `json:"turn"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

func NewGameState(seed uint64) *GameState {
	return &GameState{SchemaVersion: StateSchemaVersion, Seed: seed, RNGState: seed, Turn: 1, NextID: 1}
}

func (s *GameState) NewID() ID {
	id := s.NextID
	s.NextID++
	return id
}

func (s *GameState) RNG() *RNG {
	return &RNG{state: s.RNGState}
}

func (s *GameState) CommitRNG(rng *RNG) {
	s.RNGState = rng.State()
}

func (s *GameState) AdvanceTurn() {
	s.Turn++
}

func (s *GameState) AddEvent(kind, message string) {
	s.Events = append(s.Events, Event{Turn: s.Turn, Kind: kind, Message: message})
}

func (s *GameState) Validate() error {
	if s.SchemaVersion != StateSchemaVersion {
		return fmt.Errorf("unsupported state schema version %d", s.SchemaVersion)
	}
	if s.NextID == 0 {
		return fmt.Errorf("next_id must be non-zero")
	}
	if s.Turn == 0 {
		return fmt.Errorf("turn must start at 1 or later")
	}

	seen := make(map[ID]string)
	planetIDs := make(map[ID]struct{})
	systemIDs := make(map[ID]struct{})
	empireIDs := make(map[ID]struct{})
	colonyIDs := make(map[ID]struct{})
	outpostIDs := make(map[ID]struct{})
	colonyPlanetIDs := make(map[ID]ID)
	outpostPlanetIDs := make(map[ID]ID)
	checkID := func(id ID, label string) error {
		if id == 0 {
			return fmt.Errorf("%s has zero id", label)
		}
		if previous, ok := seen[id]; ok {
			return fmt.Errorf("duplicate id %d used by %s and %s", id, previous, label)
		}
		seen[id] = label
		return nil
	}

	if err := checkID(s.Galaxy.ID, "galaxy"); err != nil {
		return err
	}
	for si := range s.Galaxy.Systems {
		system := &s.Galaxy.Systems[si]
		if system.Name == "" {
			return fmt.Errorf("system %d has no name", si)
		}
		if err := checkID(system.ID, fmt.Sprintf("system[%d]", si)); err != nil {
			return err
		}
		systemIDs[system.ID] = struct{}{}
		lastBlockadedEmpireID := ID(0)
		for bi, empireID := range system.BlockadedEmpireIDs {
			if empireID == 0 {
				return fmt.Errorf("system[%d] blockaded empire id must be non-zero", si)
			}
			if bi > 0 && empireID <= lastBlockadedEmpireID {
				return fmt.Errorf("system[%d] blockaded empire ids must be strictly ascending", si)
			}
			lastBlockadedEmpireID = empireID
		}
		for pi := range system.Planets {
			planet := &system.Planets[pi]
			if planet.Name == "" || planet.SizeID == "" || planet.MineralID == "" || planet.GravityID == "" || planet.ClimateID == "" {
				return fmt.Errorf("system[%d] planet[%d] has incomplete identity", si, pi)
			}
			if err := checkID(planet.ID, fmt.Sprintf("system[%d].planet[%d]", si, pi)); err != nil {
				return err
			}
			planetIDs[planet.ID] = struct{}{}
		}
	}
	for i := range s.Empires {
		empire := &s.Empires[i]
		if empire.Name == "" || empire.RaceID == "" {
			return fmt.Errorf("empire[%d] has incomplete identity", i)
		}
		if err := checkID(empire.ID, fmt.Sprintf("empire[%d]", i)); err != nil {
			return err
		}
		if empire.Freighters < 0 {
			return fmt.Errorf("empire[%d] freighters must be non-negative", i)
		}
		if empire.CommandPoints.Capacity < 0 || empire.CommandPoints.Used < 0 {
			return fmt.Errorf("empire[%d] command points must be non-negative", i)
		}
		for _, item := range []struct {
			label string
			value float64
		}{
			{"treasury.balance_bc", empire.Treasury.BalanceBC},
			{"treasury.tax_income_bc", empire.Treasury.TaxIncomeBC},
			{"treasury.surplus_food_income_bc", empire.Treasury.SurplusFoodIncomeBC},
			{"treasury.gross_income_bc", empire.Treasury.GrossIncomeBC},
			{"treasury.building_maintenance_bc", empire.Treasury.BuildingMaintenanceBC},
			{"treasury.freighter_operating_cost_bc", empire.Treasury.FreighterOperatingCostBC},
			{"treasury.ship_command_maintenance_bc", empire.Treasury.ShipCommandMaintenanceBC},
			{"treasury.total_modeled_maintenance_bc", empire.Treasury.TotalModeledMaintenanceBC},
			{"treasury.net_modeled_income_bc", empire.Treasury.NetModeledIncomeBC},
		} {
			if math.IsNaN(item.value) || math.IsInf(item.value, 0) {
				return fmt.Errorf("empire[%d] %s must be finite", i, item.label)
			}
		}
		for _, item := range []struct {
			label string
			value float64
		}{
			{"treasury.tax_income_bc", empire.Treasury.TaxIncomeBC},
			{"treasury.surplus_food_income_bc", empire.Treasury.SurplusFoodIncomeBC},
			{"treasury.gross_income_bc", empire.Treasury.GrossIncomeBC},
			{"treasury.building_maintenance_bc", empire.Treasury.BuildingMaintenanceBC},
			{"treasury.freighter_operating_cost_bc", empire.Treasury.FreighterOperatingCostBC},
			{"treasury.total_modeled_maintenance_bc", empire.Treasury.TotalModeledMaintenanceBC},
		} {
			if item.value < 0 {
				return fmt.Errorf("empire[%d] %s must be non-negative", i, item.label)
			}
		}
		if empire.FoodLogistics.FreightersRequired < 0 || empire.FoodLogistics.FreightersUsed < 0 || empire.FoodLogistics.FreightersUsed > empire.Freighters {
			return fmt.Errorf("empire[%d] food logistics freighter counts are invalid", i)
		}
		for _, item := range []struct {
			label string
			value float64
		}{
			{"local_food_surplus", empire.FoodLogistics.LocalFoodSurplus},
			{"local_food_shortage", empire.FoodLogistics.LocalFoodShortage},
			{"blocked_food_surplus", empire.FoodLogistics.BlockedFoodSurplus},
			{"blocked_food_shortage", empire.FoodLogistics.BlockedFoodShortage},
			{"food_transferred", empire.FoodLogistics.FoodTransferred},
			{"food_unmet", empire.FoodLogistics.FoodUnmet},
			{"surplus_food_sold", empire.FoodLogistics.SurplusFoodSold},
			{"freighter_operating_cost_bc", empire.FoodLogistics.FreighterOperatingCostBC},
			{"surplus_food_income_bc", empire.FoodLogistics.SurplusFoodIncomeBC},
		} {
			if !finiteNonNegative(item.value) {
				return fmt.Errorf("empire[%d] food logistics %s must be finite and non-negative", i, item.label)
			}
		}
		seenTech := make(map[int]struct{}, len(empire.KnownTechnologyIDs))
		lastTech := 0
		for ti, technologyID := range empire.KnownTechnologyIDs {
			if technologyID < 1 || technologyID > 203 {
				return fmt.Errorf("empire[%d] known technology id %d outside [1,203]", i, technologyID)
			}
			if _, exists := seenTech[technologyID]; exists {
				return fmt.Errorf("empire[%d] duplicate known technology id %d", i, technologyID)
			}
			if ti > 0 && technologyID <= lastTech {
				return fmt.Errorf("empire[%d] known technology ids must be strictly ascending", i)
			}
			seenTech[technologyID] = struct{}{}
			lastTech = technologyID
		}
		seenField := make(map[int]struct{}, len(empire.KnownTechnologyFieldIDs))
		lastField := -1
		for fi, fieldID := range empire.KnownTechnologyFieldIDs {
			if fieldID < 0 || fieldID > 82 {
				return fmt.Errorf("empire[%d] known technology field id %d outside [0,82]", i, fieldID)
			}
			if fieldID >= 75 {
				return fmt.Errorf("empire[%d] repeatable Hyper-Advanced field %d must not be stored as permanently known", i, fieldID)
			}
			if _, exists := seenField[fieldID]; exists {
				return fmt.Errorf("empire[%d] duplicate known technology field id %d", i, fieldID)
			}
			if fi > 0 && fieldID <= lastField {
				return fmt.Errorf("empire[%d] known technology field ids must be strictly ascending", i)
			}
			seenField[fieldID] = struct{}{}
			lastField = fieldID
		}
		lastHyperField := 0
		for hi, level := range empire.HyperAdvancedResearch {
			if level.TechFieldID < 75 || level.TechFieldID > 82 || level.CompletedLevels <= 0 {
				return fmt.Errorf("empire[%d] invalid hyper-advanced research level field=%d completed_levels=%d", i, level.TechFieldID, level.CompletedLevels)
			}
			if hi > 0 && level.TechFieldID <= lastHyperField {
				return fmt.Errorf("empire[%d] hyper-advanced research levels must be strictly ascending by field", i)
			}
			lastHyperField = level.TechFieldID
		}
		lastFixedField := 0
		for fi, fixed := range empire.UncreativeResearchChoices {
			if fixed.TechFieldID < 1 || fixed.TechFieldID > 82 || fixed.TechnologyID < 1 || fixed.TechnologyID > 203 {
				return fmt.Errorf("empire[%d] invalid uncreative research choice field=%d technology=%d", i, fixed.TechFieldID, fixed.TechnologyID)
			}
			if fi > 0 && fixed.TechFieldID <= lastFixedField {
				return fmt.Errorf("empire[%d] uncreative research choices must be strictly ascending by field", i)
			}
			lastFixedField = fixed.TechFieldID
		}
		if empire.Research != nil {
			if empire.Research.TechFieldID < 1 || empire.Research.TechFieldID > 82 {
				return fmt.Errorf("empire[%d] research tech_field_id %d outside [1,82]", i, empire.Research.TechFieldID)
			}
			if _, known := seenField[empire.Research.TechFieldID]; known {
				return fmt.Errorf("empire[%d] researches already-known technology field %d", i, empire.Research.TechFieldID)
			}
			switch empire.Research.SelectionMode {
			case ResearchSelectionAll, ResearchSelectionChooseOne, ResearchSelectionFixedOne, ResearchSelectionRepeatField:
			default:
				return fmt.Errorf("empire[%d] research selection_mode %q is invalid", i, empire.Research.SelectionMode)
			}
			if math.IsNaN(empire.Research.ProgressRP) || math.IsInf(empire.Research.ProgressRP, 0) || empire.Research.ProgressRP < 0 {
				return fmt.Errorf("empire[%d] research progress_rp must be a finite non-negative number", i)
			}
			if empire.Research.SelectionMode == ResearchSelectionRepeatField {
				if empire.Research.TechFieldID < 75 || empire.Research.TechFieldID > 82 || len(empire.Research.TechnologyIDs) != 0 {
					return fmt.Errorf("empire[%d] repeat-field research requires Hyper-Advanced field 75..82 and no technology_ids", i)
				}
			} else if len(empire.Research.TechnologyIDs) == 0 {
				return fmt.Errorf("empire[%d] research technology_ids are required", i)
			}
			lastResearchTech := 0
			for ri, technologyID := range empire.Research.TechnologyIDs {
				if technologyID < 1 || technologyID > 203 {
					return fmt.Errorf("empire[%d] research technology id %d outside [1,203]", i, technologyID)
				}
				if _, known := seenTech[technologyID]; known {
					return fmt.Errorf("empire[%d] researches already-known technology id %d", i, technologyID)
				}
				if ri > 0 && technologyID <= lastResearchTech {
					return fmt.Errorf("empire[%d] research technology ids must be strictly ascending", i)
				}
				lastResearchTech = technologyID
			}
		}
		empireIDs[empire.ID] = struct{}{}
	}
	ships, err := validateMilitaryState(s, empireIDs, checkID)
	if err != nil {
		return err
	}
	for i := range s.Colonies {
		construction := s.Colonies[i].Construction
		if construction == nil || construction.ProjectKind != ConstructionProjectMilitaryShip {
			continue
		}
		designFound := false
		for _, design := range s.ShipDesigns {
			if design.ID == construction.ShipDesignID {
				designFound = true
				if design.EmpireID != s.Colonies[i].EmpireID || construction.ShipDesignRevision != design.Revision {
					return fmt.Errorf("colony[%d] military construction references incompatible design %d revision %d", i, construction.ShipDesignID, construction.ShipDesignRevision)
				}
				break
			}
		}
		if !designFound {
			return fmt.Errorf("colony[%d] military construction references unknown design %d", i, construction.ShipDesignID)
		}
	}
	if err := validateStrategicState(s, empireIDs, systemIDs, ships, checkID); err != nil {
		return err
	}
	for si := range s.Galaxy.Systems {
		for _, empireID := range s.Galaxy.Systems[si].BlockadedEmpireIDs {
			if _, ok := empireIDs[empireID]; !ok {
				return fmt.Errorf("system[%d] references unknown blockaded empire %d", si, empireID)
			}
		}
	}
	for i := range s.Colonies {
		colony := &s.Colonies[i]
		if colony.EmpireID == 0 || colony.PlanetID == 0 {
			return fmt.Errorf("colony[%d] has incomplete references", i)
		}
		if colony.Construction != nil {
			if colony.Construction.ProjectKind != ConstructionProjectBuilding && colony.Construction.ProjectKind != ConstructionProjectColonyShip && colony.Construction.ProjectKind != ConstructionProjectOutpostShip && colony.Construction.ProjectKind != ConstructionProjectMilitaryShip && colony.Construction.ProjectKind != ConstructionProjectFreighterFleet && colony.Construction.ProjectKind != ConstructionProjectHousing && colony.Construction.ProjectKind != ConstructionProjectPlanetaryTransformation {
				return fmt.Errorf("colony[%d] construction project_kind %q is invalid", i, colony.Construction.ProjectKind)
			}
			if colony.Construction.ProjectID == "" {
				return fmt.Errorf("colony[%d] construction project_id is required", i)
			}
			if !finiteNonNegative(colony.Construction.ProgressPP) {
				return fmt.Errorf("colony[%d] construction progress_pp must be finite and non-negative", i)
			}
			if colony.Construction.ProjectKind == ConstructionProjectColonyShip && colony.Construction.ProjectID != "colony_ship" {
				return fmt.Errorf("colony[%d] Colony Ship construction project_id must be %q", i, "colony_ship")
			}
			if colony.Construction.ProjectKind == ConstructionProjectOutpostShip && colony.Construction.ProjectID != "outpost_ship" {
				return fmt.Errorf("colony[%d] Outpost Ship construction project_id must be %q", i, "outpost_ship")
			}
			if colony.Construction.ProjectKind == ConstructionProjectMilitaryShip {
				if colony.Construction.ProjectID != "military_ship" || colony.Construction.ShipDesignID == 0 || colony.Construction.ShipDesignRevision == 0 {
					return fmt.Errorf("colony[%d] military Ship construction requires project_id %q and positive design id/revision", i, "military_ship")
				}
			} else if colony.Construction.ShipDesignID != 0 || colony.Construction.ShipDesignRevision != 0 {
				return fmt.Errorf("colony[%d] non-military construction cannot reference a ship design", i)
			}
			if colony.Construction.ProjectKind == ConstructionProjectHousing {
				if colony.Construction.ProjectID != "housing" {
					return fmt.Errorf("colony[%d] housing construction project_id must be %q", i, "housing")
				}
				if !nearlyEqual(colony.Construction.ProgressPP, 0) {
					return fmt.Errorf("colony[%d] housing construction must not accumulate progress_pp", i)
				}
			}
			if colony.Construction.ProjectKind == ConstructionProjectPlanetaryTransformation {
				if colony.Construction.ProjectID != "terraforming" && colony.Construction.ProjectID != "gaia_transformation" {
					return fmt.Errorf("colony[%d] planetary transformation project_id %q is invalid", i, colony.Construction.ProjectID)
				}
			}
		}
		population := colony.Population
		if err := population.Validate(colony.EmpireID, func(id ID) bool {
			_, ok := empireIDs[id]
			return ok
		}); err != nil {
			return fmt.Errorf("colony[%d] population: %w", i, err)
		}
		for _, item := range []struct {
			label string
			value float64
		}{
			{"food", colony.Economy.Food},
			{"production", colony.Economy.Production},
			{"research", colony.Economy.Research},
			{"tax_bc", colony.Economy.TaxBC},
			{"adjusted_food", colony.AdjustedEconomy.Food},
			{"adjusted_production", colony.AdjustedEconomy.Production},
			{"adjusted_research", colony.AdjustedEconomy.Research},
			{"adjusted_tax_bc", colony.AdjustedEconomy.TaxBC},
		} {
			if !finiteNonNegative(item.value) {
				return fmt.Errorf("colony[%d] economy %s must be finite and non-negative", i, item.label)
			}
		}
		for _, item := range []struct {
			label string
			value float64
		}{
			{"population_capacity", colony.PopulationDynamics.Capacity},
			{"food_required", colony.PopulationDynamics.FoodRequired},
			{"owner_origin_food2", colony.PopulationDynamics.OwnerOriginFood2},
			{"assimilated_foreign_food2", colony.PopulationDynamics.AssimilatedForeignFood2},
			{"conquered_foreign_food2", colony.PopulationDynamics.ConqueredForeignFood2},
			{"remaining_food2", colony.PopulationDynamics.RemainingFood2},
			{"whole_food_required", colony.PopulationDynamics.WholeFoodRequired},
			{"local_food_surplus", colony.PopulationDynamics.LocalFoodSurplus},
			{"local_food_shortage", colony.PopulationDynamics.LocalFoodShortage},
			{"food_imported", colony.PopulationDynamics.FoodImported},
			{"food_exported", colony.PopulationDynamics.FoodExported},
			{"food_surplus", colony.PopulationDynamics.FoodSurplus},
			{"food_shortage", colony.PopulationDynamics.FoodShortage},
			{"production_required", colony.PopulationDynamics.ProductionRequired},
			{"production_shortage", colony.PopulationDynamics.ProductionShortage},
			{"production_available", colony.PopulationDynamics.ProductionAvailable},
			{"base_growth", colony.PopulationDynamics.BaseGrowth},
			{"growth_multiplier", colony.PopulationDynamics.GrowthMultiplier},
			{"projected_growth", colony.PopulationDynamics.ProjectedGrowth},
			{"projected_starvation", colony.PopulationDynamics.ProjectedStarvation},
		} {
			if !finiteNonNegative(item.value) {
				return fmt.Errorf("colony[%d] %s must be finite and non-negative", i, item.label)
			}
		}
		originSeen := make(map[ID]struct{}, len(colony.PopulationDynamics.Origins))
		for oi, origin := range colony.PopulationDynamics.Origins {
			if origin.OriginEmpireID == 0 {
				return fmt.Errorf("colony[%d] population origin dynamics[%d] missing origin empire", i, oi)
			}
			if _, ok := empireIDs[origin.OriginEmpireID]; !ok {
				return fmt.Errorf("colony[%d] population origin dynamics[%d] references unknown empire %d", i, oi, origin.OriginEmpireID)
			}
			if _, ok := originSeen[origin.OriginEmpireID]; ok {
				return fmt.Errorf("colony[%d] population origin dynamics duplicates empire %d", i, origin.OriginEmpireID)
			}
			originSeen[origin.OriginEmpireID] = struct{}{}
			for _, item := range []float64{origin.Capacity, origin.Population, origin.Food2Required, origin.BaseGrowth, origin.GrowthMultiplier, origin.ProjectedGrowth, origin.ProjectedStarvation} {
				if !finiteNonNegative(item) {
					return fmt.Errorf("colony[%d] population origin dynamics[%d] contains invalid value", i, oi)
				}
			}
		}
		if colony.PopulationDynamics.FoodImported > colony.PopulationDynamics.LocalFoodShortage+1e-9 || colony.PopulationDynamics.FoodExported > colony.PopulationDynamics.LocalFoodSurplus+1e-9 {
			return fmt.Errorf("colony[%d] food import/export exceeds local shortage/surplus", i)
		}
		if !nearlyEqual(colony.PopulationDynamics.FoodShortage, math.Max(0, colony.PopulationDynamics.LocalFoodShortage-colony.PopulationDynamics.FoodImported)) || !nearlyEqual(colony.PopulationDynamics.FoodSurplus, math.Max(0, colony.PopulationDynamics.LocalFoodSurplus-colony.PopulationDynamics.FoodExported)) {
			return fmt.Errorf("colony[%d] food logistics snapshot is inconsistent", i)
		}
		if colony.PopulationDynamics.ProjectedGrowth > 1e-9 && colony.PopulationDynamics.ProjectedStarvation > 1e-9 {
			return fmt.Errorf("colony[%d] cannot project growth and starvation simultaneously", i)
		}
		if colony.PopulationDynamics.FoodSurplus > 1e-9 && colony.PopulationDynamics.FoodShortage > 1e-9 {
			return fmt.Errorf("colony[%d] cannot have food surplus and shortage simultaneously", i)
		}
		if colony.PopulationDynamics.ProductionAvailable > 1e-9 && colony.PopulationDynamics.ProductionShortage > 1e-9 {
			return fmt.Errorf("colony[%d] cannot have production available and shortage simultaneously", i)
		}
		if colony.PopulationDynamics.Capacity > 0 && colony.Population.Total() <= colony.PopulationDynamics.Capacity+1e-9 {
			remaining := math.Max(0, colony.PopulationDynamics.Capacity-colony.Population.Total())
			if colony.PopulationDynamics.ProjectedGrowth > remaining+1e-9 {
				return fmt.Errorf("colony[%d] projected growth %g exceeds remaining capacity %g", i, colony.PopulationDynamics.ProjectedGrowth, remaining)
			}
		}
		seenBuildings := make(map[string]struct{}, len(colony.Buildings))
		for bi, buildingID := range colony.Buildings {
			if buildingID == "" {
				return fmt.Errorf("colony[%d] building[%d] has empty id", i, bi)
			}
			if _, exists := seenBuildings[buildingID]; exists {
				return fmt.Errorf("colony[%d] has duplicate building %q", i, buildingID)
			}
			seenBuildings[buildingID] = struct{}{}
		}
		context := colony.EconomyContext
		if context.GravityPenaltyPercent < 0 || context.GravityPenaltyPercent > 100 {
			return fmt.Errorf("colony[%d] gravity penalty percent %d outside 0..100", i, context.GravityPenaltyPercent)
		}
		for _, percent := range []int{context.GovernmentFoodPercent, context.GovernmentProductionPercent, context.GovernmentResearchPercent, context.GovernmentTaxPercent, context.MoraleBarracksPenaltyPercent, context.MoraleBuildingBonusPercent, context.MoralePercent} {
			if percent < -100 || percent > 200 {
				return fmt.Errorf("colony[%d] government economy percent %d outside -100..200", i, percent)
			}
		}
		if _, ok := empireIDs[colony.EmpireID]; !ok {
			return fmt.Errorf("colony[%d] references unknown empire %d", i, colony.EmpireID)
		}
		if _, ok := planetIDs[colony.PlanetID]; !ok {
			return fmt.Errorf("colony[%d] references unknown planet %d", i, colony.PlanetID)
		}
		if previous, exists := colonyPlanetIDs[colony.PlanetID]; exists {
			return fmt.Errorf("colony[%d] shares planet %d with colony %d", i, colony.PlanetID, previous)
		}
		colonyPlanetIDs[colony.PlanetID] = colony.ID
		if err := checkID(colony.ID, fmt.Sprintf("colony[%d]", i)); err != nil {
			return err
		}
		colonyIDs[colony.ID] = struct{}{}
	}
	for i := range s.Outposts {
		outpost := &s.Outposts[i]
		if outpost.EmpireID == 0 || outpost.PlanetID == 0 {
			return fmt.Errorf("outpost[%d] has incomplete references", i)
		}
		if _, ok := empireIDs[outpost.EmpireID]; !ok {
			return fmt.Errorf("outpost[%d] references unknown empire %d", i, outpost.EmpireID)
		}
		if _, ok := planetIDs[outpost.PlanetID]; !ok {
			return fmt.Errorf("outpost[%d] references unknown planet %d", i, outpost.PlanetID)
		}
		if colonyID, occupied := colonyPlanetIDs[outpost.PlanetID]; occupied {
			return fmt.Errorf("outpost[%d] shares planet %d with colony %d", i, outpost.PlanetID, colonyID)
		}
		if previous, exists := outpostPlanetIDs[outpost.PlanetID]; exists {
			return fmt.Errorf("outpost[%d] shares planet %d with outpost %d", i, outpost.PlanetID, previous)
		}
		if err := checkID(outpost.ID, fmt.Sprintf("outpost[%d]", i)); err != nil {
			return err
		}
		outpostPlanetIDs[outpost.PlanetID] = outpost.ID
		outpostIDs[outpost.ID] = struct{}{}
	}
	for i := range s.Empires {
		if capital := s.Empires[i].Capital; capital != 0 {
			if _, ok := colonyIDs[capital]; !ok {
				return fmt.Errorf("empire[%d] references unknown capital colony %d", i, capital)
			}
		}
	}
	for si := range s.Galaxy.Systems {
		for pi := range s.Galaxy.Systems[si].Planets {
			planet := &s.Galaxy.Systems[si].Planets[pi]
			if planet.ColonyID != 0 && planet.OutpostID != 0 {
				return fmt.Errorf("system[%d] planet[%d] cannot contain both colony %d and outpost %d", si, pi, planet.ColonyID, planet.OutpostID)
			}
			if colonyID := planet.ColonyID; colonyID != 0 {
				if _, ok := colonyIDs[colonyID]; !ok {
					return fmt.Errorf("system[%d] planet[%d] references unknown colony %d", si, pi, colonyID)
				}
				if expected := colonyPlanetIDs[planet.ID]; expected != colonyID {
					return fmt.Errorf("system[%d] planet[%d] colony link %d does not match colony on planet %d", si, pi, colonyID, expected)
				}
			} else if expected := colonyPlanetIDs[planet.ID]; expected != 0 {
				return fmt.Errorf("system[%d] planet[%d] is missing reciprocal colony link %d", si, pi, expected)
			}
			if outpostID := planet.OutpostID; outpostID != 0 {
				if _, ok := outpostIDs[outpostID]; !ok {
					return fmt.Errorf("system[%d] planet[%d] references unknown outpost %d", si, pi, outpostID)
				}
				if expected := outpostPlanetIDs[planet.ID]; expected != outpostID {
					return fmt.Errorf("system[%d] planet[%d] outpost link %d does not match outpost on planet %d", si, pi, outpostID, expected)
				}
			} else if expected := outpostPlanetIDs[planet.ID]; expected != 0 {
				return fmt.Errorf("system[%d] planet[%d] is missing reciprocal outpost link %d", si, pi, expected)
			}
		}
	}
	for i, transfer := range s.PopulationTransfers {
		if transfer.EmpireID == 0 || transfer.SourceColonyID == 0 || transfer.DestinationColonyID == 0 {
			return fmt.Errorf("population_transfer[%d] has incomplete references", i)
		}
		if transfer.SourceColonyID == transfer.DestinationColonyID {
			return fmt.Errorf("population_transfer[%d] source and destination colony are identical", i)
		}
		if transfer.OriginEmpireID == 0 || transfer.LoyaltyEmpireID == 0 || transfer.AssimilationState != PopulationAssimilated {
			return fmt.Errorf("population_transfer[%d] must identify assimilated organic cohort", i)
		}
		if _, ok := empireIDs[transfer.OriginEmpireID]; !ok {
			return fmt.Errorf("population_transfer[%d] references unknown origin empire %d", i, transfer.OriginEmpireID)
		}
		if _, ok := empireIDs[transfer.LoyaltyEmpireID]; !ok {
			return fmt.Errorf("population_transfer[%d] references unknown loyalty empire %d", i, transfer.LoyaltyEmpireID)
		}
		if transfer.LoyaltyEmpireID != transfer.EmpireID {
			return fmt.Errorf("population_transfer[%d] assimilated cohort loyalty must match transfer empire", i)
		}
		if transfer.Job != PopulationJobFarmer && transfer.Job != PopulationJobWorker && transfer.Job != PopulationJobScientist {
			return fmt.Errorf("population_transfer[%d] has invalid job %q", i, transfer.Job)
		}
		if transfer.RemainingTurns < 1 || transfer.RemainingTurns > 15 {
			return fmt.Errorf("population_transfer[%d] remaining_turns %d outside [1,15]", i, transfer.RemainingTurns)
		}
		if _, ok := empireIDs[transfer.EmpireID]; !ok {
			return fmt.Errorf("population_transfer[%d] references unknown empire %d", i, transfer.EmpireID)
		}
		if _, ok := colonyIDs[transfer.SourceColonyID]; !ok {
			return fmt.Errorf("population_transfer[%d] references unknown source colony %d", i, transfer.SourceColonyID)
		}
		if _, ok := colonyIDs[transfer.DestinationColonyID]; !ok {
			return fmt.Errorf("population_transfer[%d] references unknown destination colony %d", i, transfer.DestinationColonyID)
		}
		if err := checkID(transfer.ID, fmt.Sprintf("population_transfer[%d]", i)); err != nil {
			return err
		}
	}
	for id := range seen {
		if id >= s.NextID {
			return fmt.Errorf("id %d is not below next_id %d", id, s.NextID)
		}
	}
	return nil
}

func finiteNonNegative(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0
}

func nearlyEqual(a, b float64) bool {
	scale := math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
	return math.Abs(a-b) <= 1e-9*scale
}
