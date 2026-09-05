package game

import (
	"fmt"

	"moox/internal/core"
	"moox/internal/protocol"
)

const CommandSetConstructionQueue = "colony.set_construction_queue"

type ConstructionQueueItem struct {
	ProjectKind        core.ConstructionProjectKind `json:"project_kind"`
	ProjectID          string                       `json:"project_id"`
	ShipDesignID       core.ID                      `json:"ship_design_id,omitempty"`
	ShipDesignRevision uint32                       `json:"ship_design_revision,omitempty"`
}

type SetConstructionQueuePayload struct {
	ColonyID core.ID                 `json:"colony_id"`
	Items    []ConstructionQueueItem `json:"items"`
}

type ConstructionQueueSetEvent struct {
	ColonyID  core.ID `json:"colony_id"`
	ItemCount int     `json:"item_count"`
	StockPP   float64 `json:"stock_pp"`
}

func NewSetConstructionQueueCommand(sequence uint32, payload SetConstructionQueuePayload) (protocol.Command, error) {
	if err := validateSetConstructionQueuePayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandSetConstructionQueue, payload)
}

func decodeSetConstructionQueue(command protocol.Command) (SetConstructionQueuePayload, error) {
	var payload SetConstructionQueuePayload
	if err := decodeStrictCommandPayload(command, CommandSetConstructionQueue, &payload); err != nil {
		return SetConstructionQueuePayload{}, err
	}
	if err := validateSetConstructionQueuePayload(payload); err != nil {
		return SetConstructionQueuePayload{}, err
	}
	return payload, nil
}

func validateSetConstructionQueuePayload(payload SetConstructionQueuePayload) error {
	if payload.ColonyID == 0 {
		return fmt.Errorf("colony_id must be non-zero")
	}
	for i, item := range payload.Items {
		if err := validateConstructionQueueItem(item); err != nil {
			return fmt.Errorf("items[%d]: %w", i, err)
		}
	}
	return nil
}

func validateConstructionQueueItem(item ConstructionQueueItem) error {
	if item.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}
	switch item.ProjectKind {
	case core.ConstructionProjectBuilding,
		core.ConstructionProjectColonyShip,
		core.ConstructionProjectOutpostShip,
		core.ConstructionProjectTroopTransport,
		core.ConstructionProjectMilitaryShip,
		core.ConstructionProjectFreighterFleet,
		core.ConstructionProjectHousing,
		core.ConstructionProjectPlanetaryTransformation:
	default:
		return fmt.Errorf("unsupported project_kind %q", item.ProjectKind)
	}
	if item.ProjectKind == core.ConstructionProjectMilitaryShip {
		if item.ShipDesignID == 0 || item.ShipDesignRevision == 0 {
			return fmt.Errorf("military Ship requires ship_design_id and ship_design_revision")
		}
	} else if item.ShipDesignID != 0 || item.ShipDesignRevision != 0 {
		return fmt.Errorf("non-military project cannot reference a ship design")
	}
	return nil
}

func constructionQueueItemFromChoice(choice ConstructionChoice) ConstructionQueueItem {
	return ConstructionQueueItem{
		ProjectKind:        choice.ProjectKind,
		ProjectID:          choice.ProjectID,
		ShipDesignID:       choice.ShipDesignID,
		ShipDesignRevision: choice.ShipDesignRevision,
	}
}

func constructionStateFromQueueItem(item ConstructionQueueItem) core.ConstructionState {
	return core.ConstructionState{
		ProjectKind:        item.ProjectKind,
		ProjectID:          item.ProjectID,
		ShipDesignID:       item.ShipDesignID,
		ShipDesignRevision: item.ShipDesignRevision,
	}
}

func constructionQueueItemMatchesChoice(item ConstructionQueueItem, choice ConstructionChoice) bool {
	return item.ProjectKind == choice.ProjectKind &&
		item.ProjectID == choice.ProjectID &&
		item.ShipDesignID == choice.ShipDesignID &&
		item.ShipDesignRevision == choice.ShipDesignRevision
}

// AvailableConstructionQueueChoices returns the legal buildable catalog even
// while a Colony already has an active construction project. It never mutates
// the authoritative state.
func (r *EconomyRules) AvailableConstructionQueueChoices(state *core.GameState, empireID, colonyID core.ID) ([]ConstructionChoice, error) {
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	clone := *state
	clone.Colonies = append([]core.Colony(nil), state.Colonies...)
	found := false
	for i := range clone.Colonies {
		if clone.Colonies[i].ID == colonyID {
			clone.Colonies[i].Construction = nil
			clone.Colonies[i].ConstructionQueue = nil
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("unknown colony %d", colonyID)
	}
	return r.AvailableConstructionChoices(&clone, empireID, colonyID)
}

func (r *EconomyResolver) setConstructionQueue(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeSetConstructionQueue(command)
	if err != nil {
		return DomainEvent{}, err
	}
	colony := colonyByID(state, payload.ColonyID)
	if colony == nil {
		return DomainEvent{}, fmt.Errorf("references unknown colony %d", payload.ColonyID)
	}
	if colony.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot edit construction on colony %d owned by empire %d", seatID, colony.ID, colony.EmpireID)
	}
	choices, err := r.Rules.AvailableConstructionQueueChoices(state, empireID, colony.ID)
	if err != nil {
		return DomainEvent{}, err
	}
	seenUnique := map[string]struct{}{}
	for i, item := range payload.Items {
		legal := false
		for _, choice := range choices {
			if constructionQueueItemMatchesChoice(item, choice) {
				legal = true
				break
			}
		}
		if !legal {
			return DomainEvent{}, fmt.Errorf("queue item %d project %s/%q is not currently buildable", i, item.ProjectKind, item.ProjectID)
		}
		if item.ProjectKind == core.ConstructionProjectBuilding || item.ProjectKind == core.ConstructionProjectPlanetaryTransformation {
			key := string(item.ProjectKind) + "\x00" + item.ProjectID
			if _, exists := seenUnique[key]; exists {
				return DomainEvent{}, fmt.Errorf("queue item %d duplicates non-repeatable project %s/%q", i, item.ProjectKind, item.ProjectID)
			}
			seenUnique[key] = struct{}{}
		}
	}

	stockPP := colony.ConstructionReservePP
	if colony.Construction != nil {
		stockPP += colony.Construction.ProgressPP
	}
	colony.Construction = nil
	colony.ConstructionQueue = nil
	colony.ConstructionReservePP = 0
	if len(payload.Items) == 0 {
		colony.ConstructionReservePP = stockPP
	} else {
		head := constructionStateFromQueueItem(payload.Items[0])
		head.ProgressPP = stockPP
		colony.Construction = &head
		if len(payload.Items) > 1 {
			colony.ConstructionQueue = make([]core.ConstructionState, 0, len(payload.Items)-1)
			for _, item := range payload.Items[1:] {
				colony.ConstructionQueue = append(colony.ConstructionQueue, constructionStateFromQueueItem(item))
			}
		}
	}
	return NewDomainEvent("colony.construction_queue_set", seatID, command.Sequence, ConstructionQueueSetEvent{ColonyID: colony.ID, ItemCount: len(payload.Items), StockPP: stockPP})
}

func promoteConstructionQueue(colony *core.Colony) bool {
	if colony == nil || len(colony.ConstructionQueue) == 0 {
		if colony != nil {
			colony.Construction = nil
		}
		return false
	}
	next := colony.ConstructionQueue[0]
	colony.ConstructionQueue = append([]core.ConstructionState(nil), colony.ConstructionQueue[1:]...)
	colony.Construction = &next
	return true
}
