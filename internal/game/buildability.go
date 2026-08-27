package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

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
		if _, exists := owned[definition.BuildingID]; exists {
			continue
		}
		if _, known := knownTech[definition.TechnologyID]; !known {
			continue
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
