package game

import (
	"fmt"

	"moox/internal/core"
	"moox/internal/protocol"
)

type ConstructionQueuedEvent struct {
	ColonyID         core.ID `json:"colony_id"`
	BuildingID       string  `json:"building_id"`
	ProductionCostPP float64 `json:"production_cost_pp"`
}

type ConstructionProgressedEvent struct {
	ColonyID    core.ID                      `json:"colony_id"`
	ProjectKind core.ConstructionProjectKind `json:"project_kind"`
	ProjectID   string                       `json:"project_id"`
	AppliedPP   float64                      `json:"applied_pp"`
	ProgressPP  float64                      `json:"progress_pp"`
	RemainingPP float64                      `json:"remaining_pp"`
}

type BuildingCompletedEvent struct {
	ColonyID   core.ID `json:"colony_id"`
	BuildingID string  `json:"building_id"`
}

func (r *EconomyResolver) queueBuilding(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeQueueBuilding(command)
	if err != nil {
		return DomainEvent{}, err
	}
	colony := colonyByID(state, payload.ColonyID)
	if colony == nil {
		return DomainEvent{}, fmt.Errorf("references unknown colony %d", payload.ColonyID)
	}
	if colony.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot queue construction on colony %d owned by empire %d", seatID, colony.ID, colony.EmpireID)
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}
	definition, ok := r.Rules.BuildingDefinitions[payload.BuildingID]
	if !ok {
		return DomainEvent{}, fmt.Errorf("unknown building %q", payload.BuildingID)
	}
	if !empireKnowsTechnology(empire, definition.TechnologyID) {
		return DomainEvent{}, fmt.Errorf("empire %d does not know technology %d required for building %q", empireID, definition.TechnologyID, payload.BuildingID)
	}
	for _, buildingID := range colony.Buildings {
		if buildingID == payload.BuildingID {
			return DomainEvent{}, fmt.Errorf("colony %d already owns building %q", colony.ID, payload.BuildingID)
		}
	}
	if colony.Construction != nil {
		return DomainEvent{}, fmt.Errorf("colony %d already constructs %q", colony.ID, colony.Construction.ProjectID)
	}
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectBuilding, ProjectID: payload.BuildingID}
	return NewDomainEvent("colony.construction_queued", seatID, command.Sequence, ConstructionQueuedEvent{
		ColonyID:         colony.ID,
		BuildingID:       payload.BuildingID,
		ProductionCostPP: definition.ProductionCostPP,
	})
}

func (r *EconomyResolver) advanceConstruction(state *core.GameState) ([]DomainEvent, error) {
	var events []DomainEvent
	for i := range state.Colonies {
		colony := &state.Colonies[i]
		if colony.Construction == nil {
			continue
		}
		projectKind := colony.Construction.ProjectKind
		projectID := colony.Construction.ProjectID
		costPP, err := r.constructionProjectCostPP(colony.Construction)
		if err != nil {
			return nil, fmt.Errorf("colony %d: %w", colony.ID, err)
		}
		remaining := costPP - colony.Construction.ProgressPP
		if remaining <= 1e-9 {
			return nil, fmt.Errorf("colony %d construction %s %q has invalid completed progress %g", colony.ID, projectKind, projectID, colony.Construction.ProgressPP)
		}
		applied := colony.PopulationDynamics.ProductionAvailable
		if applied < 0 {
			return nil, fmt.Errorf("colony %d has negative adjusted production %g", colony.ID, applied)
		}
		if applied > remaining {
			applied = remaining
		}
		colony.Construction.ProgressPP += applied
		remaining -= applied
		progress, err := NewDomainEvent("colony.construction_progressed", 0, 0, ConstructionProgressedEvent{
			ColonyID:    colony.ID,
			ProjectKind: projectKind,
			ProjectID:   projectID,
			AppliedPP:   applied,
			ProgressPP:  colony.Construction.ProgressPP,
			RemainingPP: remaining,
		})
		if err != nil {
			return nil, err
		}
		events = append(events, progress)
		if remaining > 1e-9 {
			continue
		}
		colony.Construction = nil
		switch projectKind {
		case core.ConstructionProjectBuilding:
			colony.Buildings = append(colony.Buildings, projectID)
			completed, err := NewDomainEvent("colony.building_completed", 0, 0, BuildingCompletedEvent{ColonyID: colony.ID, BuildingID: projectID})
			if err != nil {
				return nil, err
			}
			events = append(events, completed)
		case core.ConstructionProjectFreighterFleet:
			empire := empireByID(state, colony.EmpireID)
			if empire == nil {
				return nil, fmt.Errorf("colony %d references unknown empire %d", colony.ID, colony.EmpireID)
			}
			empire.Freighters += r.Rules.FreightersPerFleet
			completed, err := NewDomainEvent("colony.freighter_fleet_completed", 0, 0, FreighterFleetCompletedEvent{
				ColonyID:        colony.ID,
				EmpireID:        empire.ID,
				FreightersAdded: r.Rules.FreightersPerFleet,
				TotalFreighters: empire.Freighters,
			})
			if err != nil {
				return nil, err
			}
			events = append(events, completed)
		default:
			return nil, fmt.Errorf("colony %d completed unsupported project kind %q", colony.ID, projectKind)
		}
	}
	return events, nil
}

func (r *EconomyResolver) constructionProjectCostPP(project *core.ConstructionState) (float64, error) {
	if project == nil {
		return 0, fmt.Errorf("construction project must not be nil")
	}
	switch project.ProjectKind {
	case core.ConstructionProjectBuilding:
		definition, ok := r.Rules.BuildingDefinitions[project.ProjectID]
		if !ok {
			return 0, fmt.Errorf("constructs unknown building %q", project.ProjectID)
		}
		return definition.ProductionCostPP, nil
	case core.ConstructionProjectFreighterFleet:
		if project.ProjectID != FreighterFleetProjectID {
			return 0, fmt.Errorf("constructs unknown Freighter Fleet project %q", project.ProjectID)
		}
		return r.Rules.FreighterFleetCostPP, nil
	default:
		return 0, fmt.Errorf("constructs unsupported project kind %q", project.ProjectKind)
	}
}

type FreighterFleetCompletedEvent struct {
	ColonyID        core.ID `json:"colony_id"`
	EmpireID        core.ID `json:"empire_id"`
	FreightersAdded int     `json:"freighters_added"`
	TotalFreighters int     `json:"total_freighters"`
}
type FreighterFleetQueuedEvent struct {
	ColonyID         core.ID `json:"colony_id"`
	ProductionCostPP float64 `json:"production_cost_pp"`
	FreightersAdded  int     `json:"freighters_added"`
}

func (r *EconomyResolver) queueFreighterFleet(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeQueueFreighterFleet(command)
	if err != nil {
		return DomainEvent{}, err
	}
	colony := colonyByID(state, payload.ColonyID)
	if colony == nil {
		return DomainEvent{}, fmt.Errorf("references unknown colony %d", payload.ColonyID)
	}
	if colony.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot queue Freighter Fleet on colony %d owned by empire %d", seatID, colony.ID, colony.EmpireID)
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}
	if !empireKnowsTechnology(empire, FreighterFleetTechnologyID) {
		return DomainEvent{}, fmt.Errorf("empire %d does not know Technology %d required for Freighter Fleet", empireID, FreighterFleetTechnologyID)
	}
	if colony.Construction != nil {
		return DomainEvent{}, fmt.Errorf("colony %d already constructs %s %q", colony.ID, colony.Construction.ProjectKind, colony.Construction.ProjectID)
	}
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectFreighterFleet, ProjectID: FreighterFleetProjectID}
	return NewDomainEvent("colony.freighter_fleet_queued", seatID, command.Sequence, FreighterFleetQueuedEvent{
		ColonyID:         colony.ID,
		ProductionCostPP: r.Rules.FreighterFleetCostPP,
		FreightersAdded:  r.Rules.FreightersPerFleet,
	})
}
