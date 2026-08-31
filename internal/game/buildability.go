package game

import (
	"fmt"
	"math"
	"sort"

	"moox/internal/core"
)

const ColonyShipTechnologyID = 41
const ColonyShipProjectID = "colony_ship"
const ColonyShipBaseCostPP = 500.0
const OutpostShipTechnologyID = 109
const OutpostShipProjectID = "outpost_ship"
const OutpostShipBaseCostPP = 100.0
const FreighterFleetTechnologyID = 69
const FreighterFleetProjectID = "freighter_fleet"
const HousingProjectID = "housing"

type ConstructionChoice struct {
	ProjectKind      core.ConstructionProjectKind `json:"project_kind"`
	ProjectID        string                       `json:"project_id"`
	ProductionCostPP float64                      `json:"production_cost_pp"`
	TechnologyID     int                          `json:"technology_id,omitempty"`
	ProductionID     int                          `json:"production_id,omitempty"`
	MaintenanceBC    int                          `json:"maintenance_bc,omitempty"`
	FreightersAdded  int                          `json:"freighters_added,omitempty"`
}

func (r *EconomyRules) AvailableConstructionChoices(state *core.GameState, empireID, colonyID core.ID) ([]ConstructionChoice, error) {
	buildingChoices, err := r.AvailableBuildingChoices(state, empireID, colonyID)
	if err != nil {
		return nil, err
	}
	colony := colonyByID(state, colonyID)
	if colony == nil {
		return nil, fmt.Errorf("unknown colony %d", colonyID)
	}
	if colony.Construction != nil {
		return []ConstructionChoice{}, nil
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return nil, fmt.Errorf("unknown empire %d", empireID)
	}
	choices := make([]ConstructionChoice, 0, len(buildingChoices)+len(r.PlanetaryTransformations)+3)
	for _, choice := range buildingChoices {
		choices = append(choices, ConstructionChoice{
			ProjectKind:      core.ConstructionProjectBuilding,
			ProjectID:        choice.BuildingID,
			ProductionCostPP: choice.ProductionCostPP,
			TechnologyID:     choice.TechnologyID,
			ProductionID:     choice.ProductionID,
			MaintenanceBC:    choice.MaintenanceBC,
		})
	}
	planet := planetByID(state, colony.PlanetID)
	if planet == nil {
		return nil, fmt.Errorf("colony %d references unknown planet %d", colony.ID, colony.PlanetID)
	}
	transformationChoices := make([]ConstructionChoice, 0, len(r.PlanetaryTransformations))
	for _, definition := range r.PlanetaryTransformations {
		if !empireKnowsTechnology(empire, definition.TechnologyID) {
			continue
		}
		if _, allowed := definition.AllowedClimateIDs[planet.ClimateID]; !allowed {
			continue
		}
		transformationChoices = append(transformationChoices, ConstructionChoice{
			ProjectKind: core.ConstructionProjectPlanetaryTransformation, ProjectID: definition.ProjectID, ProductionCostPP: definition.ProductionCostPP, TechnologyID: definition.TechnologyID, ProductionID: definition.ProductionID,
		})
	}
	sort.Slice(transformationChoices, func(i, j int) bool {
		if transformationChoices[i].ProductionID != transformationChoices[j].ProductionID {
			return transformationChoices[i].ProductionID < transformationChoices[j].ProductionID
		}
		return transformationChoices[i].ProjectID < transformationChoices[j].ProjectID
	})
	choices = append(choices, transformationChoices...)
	hasGrowthRoom, err := r.mixedPopulationHasGrowthRoom(state, *colony, *planet, *empire)
	if err != nil {
		return nil, fmt.Errorf("colony %d mixed population capacity: %w", colony.ID, err)
	}
	if hasGrowthRoom {
		choices = append(choices, ConstructionChoice{
			ProjectKind: core.ConstructionProjectHousing,
			ProjectID:   HousingProjectID,
		})
	}
	if empireKnowsTechnology(empire, ColonyShipTechnologyID) {
		choices = append(choices, ConstructionChoice{
			ProjectKind:      core.ConstructionProjectColonyShip,
			ProjectID:        ColonyShipProjectID,
			ProductionCostPP: r.colonyShipProductionCostPP(empire),
			TechnologyID:     ColonyShipTechnologyID,
		})
	}
	if empireKnowsTechnology(empire, OutpostShipTechnologyID) {
		choices = append(choices, ConstructionChoice{
			ProjectKind:      core.ConstructionProjectOutpostShip,
			ProjectID:        OutpostShipProjectID,
			ProductionCostPP: r.outpostShipProductionCostPP(empire),
			TechnologyID:     OutpostShipTechnologyID,
		})
	}
	if empireKnowsTechnology(empire, FreighterFleetTechnologyID) {
		choices = append(choices, ConstructionChoice{
			ProjectKind:      core.ConstructionProjectFreighterFleet,
			ProjectID:        FreighterFleetProjectID,
			ProductionCostPP: r.FreighterFleetCostPP,
			TechnologyID:     FreighterFleetTechnologyID,
			FreightersAdded:  r.FreightersPerFleet,
		})
	}
	return choices, nil
}

func (r *EconomyRules) colonyShipProductionCostPP(empire *core.Empire) float64 {
	if empire == nil {
		return ColonyShipBaseCostPP
	}
	if modifiers, ok := r.RaceModifiers[empire.RaceID]; ok && modifiers.GovernmentTraitID == "government_feudal" {
		return math.Ceil((2 * ColonyShipBaseCostPP) / 3)
	}
	return ColonyShipBaseCostPP
}

func (r *EconomyRules) outpostShipProductionCostPP(empire *core.Empire) float64 {
	if empire == nil {
		return OutpostShipBaseCostPP
	}
	if modifiers, ok := r.RaceModifiers[empire.RaceID]; ok && modifiers.GovernmentTraitID == "government_feudal" {
		return math.Ceil((2 * OutpostShipBaseCostPP) / 3)
	}
	return OutpostShipBaseCostPP
}

type BuildingChoice struct {
	BuildingID       string  `json:"building_id"`
	ProductionID     int     `json:"production_id"`
	TechnologyID     int     `json:"technology_id"`
	ProductionCostPP float64 `json:"production_cost_pp"`
	MaintenanceBC    int     `json:"maintenance_bc"`
}

// AvailableBuildingChoices projects the currently legal normalized building
// choices for one owned colony. It is intended to be shared by human UI and AI
// controllers so callers do not need to infer buildability client-side.
func (r *EconomyRules) AvailableBuildingChoices(state *core.GameState, empireID, colonyID core.ID) ([]BuildingChoice, error) {
	if r == nil {
		return nil, fmt.Errorf("economy rules must not be nil")
	}
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return nil, fmt.Errorf("unknown empire %d", empireID)
	}
	colony := colonyByID(state, colonyID)
	if colony == nil {
		return nil, fmt.Errorf("unknown colony %d", colonyID)
	}
	if colony.EmpireID != empireID {
		return nil, fmt.Errorf("empire %d does not own colony %d", empireID, colonyID)
	}
	if colony.Construction != nil {
		return []BuildingChoice{}, nil
	}

	owned := make(map[string]struct{}, len(colony.Buildings))
	for _, buildingID := range colony.Buildings {
		owned[buildingID] = struct{}{}
	}
	knownTech := make(map[int]struct{}, len(empire.KnownTechnologyIDs))
	for _, technologyID := range empire.KnownTechnologyIDs {
		knownTech[technologyID] = struct{}{}
	}

	choices := make([]BuildingChoice, 0)
	for _, definition := range r.BuildingDefinitions {
		if _, transformation := r.PlanetaryTransformations[definition.BuildingID]; transformation {
			continue
		}
		if _, exists := owned[definition.BuildingID]; exists {
			continue
		}
		if _, known := knownTech[definition.TechnologyID]; !known {
			continue
		}
		if definition.BuildingID == ColonyBaseBuildingID {
			targets, err := colonyBaseTargetPlanetIDs(state, colony)
			if err != nil {
				return nil, err
			}
			if len(targets) == 0 {
				continue
			}
		}
		choices = append(choices, BuildingChoice{
			BuildingID:       definition.BuildingID,
			ProductionID:     definition.ProductionID,
			TechnologyID:     definition.TechnologyID,
			ProductionCostPP: definition.ProductionCostPP,
			MaintenanceBC:    definition.MaintenanceBC,
		})
	}
	sort.Slice(choices, func(i, j int) bool {
		if choices[i].ProductionID != choices[j].ProductionID {
			return choices[i].ProductionID < choices[j].ProductionID
		}
		return choices[i].BuildingID < choices[j].BuildingID
	})
	return choices, nil
}

func empireKnowsTechnology(empire *core.Empire, technologyID int) bool {
	if empire == nil || technologyID <= 0 {
		return false
	}
	index := sort.SearchInts(empire.KnownTechnologyIDs, technologyID)
	return index < len(empire.KnownTechnologyIDs) && empire.KnownTechnologyIDs[index] == technologyID
}
