package game

import (
	"fmt"

	"moox/internal/core"
	"moox/internal/protocol"
)

type ConstructionQueuedEvent struct {
	ColonyID            core.ID `json:"colony_id"`
	BuildingID          string  `json:"building_id"`
	ProductionCostMilli int64   `json:"production_cost_milli"`
}

type ConstructionProgressedEvent struct {
	ColonyID       core.ID `json:"colony_id"`
	BuildingID     string  `json:"building_id"`
	AppliedMilli   int64   `json:"applied_milli"`
	ProgressMilli  int64   `json:"progress_milli"`
	RemainingMilli int64   `json:"remaining_milli"`
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
	definition, ok := r.Rules.BuildingDefinitions[payload.BuildingID]
	if !ok {
		return DomainEvent{}, fmt.Errorf("unknown building %q", payload.BuildingID)
	}
	for _, buildingID := range colony.Buildings {
		if buildingID == payload.BuildingID {
			return DomainEvent{}, fmt.Errorf("colony %d already owns building %q", colony.ID, payload.BuildingID)
		}
	}
	if colony.Construction != nil {
		return DomainEvent{}, fmt.Errorf("colony %d already constructs %q", colony.ID, colony.Construction.BuildingID)
	}
	colony.Construction = &core.ConstructionState{BuildingID: payload.BuildingID}
	return NewDomainEvent("colony.construction_queued", seatID, command.Sequence, ConstructionQueuedEvent{
		ColonyID:            colony.ID,
		BuildingID:          payload.BuildingID,
		ProductionCostMilli: definition.ProductionCostMilli,
	})
}

func (r *EconomyResolver) advanceConstruction(state *core.GameState) ([]DomainEvent, error) {
	var events []DomainEvent
	for i := range state.Colonies {
		colony := &state.Colonies[i]
		if colony.Construction == nil {
			continue
		}
		definition, ok := r.Rules.BuildingDefinitions[colony.Construction.BuildingID]
		if !ok {
			return nil, fmt.Errorf("colony %d constructs unknown building %q", colony.ID, colony.Construction.BuildingID)
		}
		remaining := definition.ProductionCostMilli - colony.Construction.ProgressMilli
		if remaining <= 0 {
			return nil, fmt.Errorf("colony %d construction %q has invalid completed progress %d", colony.ID, colony.Construction.BuildingID, colony.Construction.ProgressMilli)
		}
		applied := colony.AdjustedEconomy.ProductionMilli
		if applied < 0 {
			return nil, fmt.Errorf("colony %d has negative adjusted production %d", colony.ID, applied)
		}
		if applied > remaining {
			applied = remaining
		}
		colony.Construction.ProgressMilli += applied
		remaining -= applied
		progress, err := NewDomainEvent("colony.construction_progressed", 0, 0, ConstructionProgressedEvent{
			ColonyID:       colony.ID,
			BuildingID:     colony.Construction.BuildingID,
			AppliedMilli:   applied,
			ProgressMilli:  colony.Construction.ProgressMilli,
			RemainingMilli: remaining,
		})
		if err != nil {
			return nil, err
		}
		events = append(events, progress)
		if remaining != 0 {
			continue
		}
		buildingID := colony.Construction.BuildingID
		colony.Buildings = append(colony.Buildings, buildingID)
		colony.Construction = nil
		completed, err := NewDomainEvent("colony.building_completed", 0, 0, BuildingCompletedEvent{ColonyID: colony.ID, BuildingID: buildingID})
		if err != nil {
			return nil, err
		}
		events = append(events, completed)
	}
	return events, nil
}
