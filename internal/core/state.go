package core

import (
	"fmt"
	"math"
)

const StateSchemaVersion = 4

type ID uint64

type GameState struct {
	SchemaVersion int      `json:"schema_version"`
	Seed          uint64   `json:"seed"`
	RNGState      uint64   `json:"rng_state"`
	Turn          uint64   `json:"turn"`
	NextID        ID       `json:"next_id"`
	Galaxy        Galaxy   `json:"galaxy"`
	Empires       []Empire `json:"empires"`
	Colonies      []Colony `json:"colonies"`
	Events        []Event  `json:"events"`
}

type Galaxy struct {
	ID      ID           `json:"id"`
	Systems []StarSystem `json:"systems"`
}

type StarSystem struct {
	ID      ID       `json:"id"`
	Name    string   `json:"name"`
	X       int      `json:"x"`
	Y       int      `json:"y"`
	Planets []Planet `json:"planets"`
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
}

type Empire struct {
	ID                      ID             `json:"id"`
	Name                    string         `json:"name"`
	RaceID                  string         `json:"race_id"`
	Capital                 ID             `json:"capital_colony_id,omitempty"`
	KnownTechnologyIDs      []int          `json:"known_technology_ids,omitempty"`
	KnownTechnologyFieldIDs []int          `json:"known_technology_field_ids,omitempty"`
	Research                *ResearchState `json:"research,omitempty"`
}

type ResearchState struct {
	TechFieldID   int     `json:"tech_field_id"`
	TechnologyIDs []int   `json:"technology_ids"`
	ProgressRP    float64 `json:"progress_rp"`
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

type ConstructionState struct {
	BuildingID string  `json:"building_id"`
	ProgressPP float64 `json:"progress_pp"`
}

type PopulationState struct {
	Total      float64 `json:"total"`
	Farmers    float64 `json:"farmers"`
	Workers    float64 `json:"workers"`
	Scientists float64 `json:"scientists"`
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
type ColonyPopulationDynamics struct {
	Capacity            float64 `json:"capacity"`
	FoodRequired        float64 `json:"food_required"`
	FoodSurplus         float64 `json:"food_surplus"`
	FoodShortage        float64 `json:"food_shortage"`
	ProductionRequired  float64 `json:"production_required"`
	ProductionShortage  float64 `json:"production_shortage"`
	ProductionAvailable float64 `json:"production_available"`
	BaseGrowth          float64 `json:"base_growth"`
	GrowthMultiplier    float64 `json:"growth_multiplier"`
	ProjectedGrowth     float64 `json:"projected_growth"`
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
	empireIDs := make(map[ID]struct{})
	colonyIDs := make(map[ID]struct{})
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
			if _, exists := seenField[fieldID]; exists {
				return fmt.Errorf("empire[%d] duplicate known technology field id %d", i, fieldID)
			}
			if fi > 0 && fieldID <= lastField {
				return fmt.Errorf("empire[%d] known technology field ids must be strictly ascending", i)
			}
			seenField[fieldID] = struct{}{}
			lastField = fieldID
		}
		if empire.Research != nil {
			if empire.Research.TechFieldID < 1 || empire.Research.TechFieldID > 82 {
				return fmt.Errorf("empire[%d] research tech_field_id %d outside [1,82]", i, empire.Research.TechFieldID)
			}
			if _, known := seenField[empire.Research.TechFieldID]; known {
				return fmt.Errorf("empire[%d] researches already-known technology field %d", i, empire.Research.TechFieldID)
			}
			if math.IsNaN(empire.Research.ProgressRP) || math.IsInf(empire.Research.ProgressRP, 0) || empire.Research.ProgressRP < 0 {
				return fmt.Errorf("empire[%d] research progress_rp must be a finite non-negative number", i)
			}
			if len(empire.Research.TechnologyIDs) == 0 {
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
	for i := range s.Colonies {
		colony := &s.Colonies[i]
		if colony.EmpireID == 0 || colony.PlanetID == 0 {
			return fmt.Errorf("colony[%d] has incomplete references", i)
		}
		if colony.Construction != nil {
			if colony.Construction.BuildingID == "" {
				return fmt.Errorf("colony[%d] construction building_id is required", i)
			}
			if !finiteNonNegative(colony.Construction.ProgressPP) {
				return fmt.Errorf("colony[%d] construction progress_pp must be finite and non-negative", i)
			}
		}
		population := colony.Population
		for _, item := range []struct {
			label string
			value float64
		}{
			{"total", population.Total},
			{"farmers", population.Farmers},
			{"workers", population.Workers},
			{"scientists", population.Scientists},
		} {
			if !finiteNonNegative(item.value) {
				return fmt.Errorf("colony[%d] population %s must be finite and non-negative", i, item.label)
			}
		}
		if assigned := population.Farmers + population.Workers + population.Scientists; !nearlyEqual(assigned, population.Total) {
			return fmt.Errorf("colony[%d] assigned population=%g does not equal total=%g", i, assigned, population.Total)
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
			{"food_surplus", colony.PopulationDynamics.FoodSurplus},
			{"food_shortage", colony.PopulationDynamics.FoodShortage},
			{"production_required", colony.PopulationDynamics.ProductionRequired},
			{"production_shortage", colony.PopulationDynamics.ProductionShortage},
			{"production_available", colony.PopulationDynamics.ProductionAvailable},
			{"base_growth", colony.PopulationDynamics.BaseGrowth},
			{"growth_multiplier", colony.PopulationDynamics.GrowthMultiplier},
			{"projected_growth", colony.PopulationDynamics.ProjectedGrowth},
		} {
			if !finiteNonNegative(item.value) {
				return fmt.Errorf("colony[%d] %s must be finite and non-negative", i, item.label)
			}
		}
		if colony.PopulationDynamics.FoodSurplus > 1e-9 && colony.PopulationDynamics.FoodShortage > 1e-9 {
			return fmt.Errorf("colony[%d] cannot have food surplus and shortage simultaneously", i)
		}
		if colony.PopulationDynamics.ProductionAvailable > 1e-9 && colony.PopulationDynamics.ProductionShortage > 1e-9 {
			return fmt.Errorf("colony[%d] cannot have production available and shortage simultaneously", i)
		}
		if colony.PopulationDynamics.Capacity > 0 && colony.Population.Total <= colony.PopulationDynamics.Capacity+1e-9 {
			remaining := math.Max(0, colony.PopulationDynamics.Capacity-colony.Population.Total)
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
		if err := checkID(colony.ID, fmt.Sprintf("colony[%d]", i)); err != nil {
			return err
		}
		colonyIDs[colony.ID] = struct{}{}
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
			if colonyID := s.Galaxy.Systems[si].Planets[pi].ColonyID; colonyID != 0 {
				if _, ok := colonyIDs[colonyID]; !ok {
					return fmt.Errorf("system[%d] planet[%d] references unknown colony %d", si, pi, colonyID)
				}
			}
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
