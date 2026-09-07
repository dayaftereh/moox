package game

import (
	"fmt"

	"moox/internal/core"
	"moox/internal/protocol"
)

const CommandSetMilitaryDesignVisual = "empire.set_military_design_visual"

type SetMilitaryDesignVisualPayload struct {
	DesignID     core.ID               `json:"design_id"`
	VisualGenome core.ShipVisualGenome `json:"visual_genome"`
}

type MilitaryDesignVisualUpdatedEvent struct {
	DesignID       core.ID               `json:"design_id"`
	VisualRevision uint32                `json:"visual_revision"`
	VisualGenome   core.ShipVisualGenome `json:"visual_genome"`
}

func IsMilitaryDesignVisualCommand(kind string) bool {
	return kind == CommandSetMilitaryDesignVisual
}

func NewSetMilitaryDesignVisualCommand(sequence uint32, payload SetMilitaryDesignVisualPayload) (protocol.Command, error) {
	if err := validateSetMilitaryDesignVisualPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandSetMilitaryDesignVisual, payload)
}

func decodeSetMilitaryDesignVisual(command protocol.Command) (SetMilitaryDesignVisualPayload, error) {
	var payload SetMilitaryDesignVisualPayload
	if err := decodeStrictCommandPayload(command, CommandSetMilitaryDesignVisual, &payload); err != nil {
		return SetMilitaryDesignVisualPayload{}, err
	}
	if err := validateSetMilitaryDesignVisualPayload(payload); err != nil {
		return SetMilitaryDesignVisualPayload{}, err
	}
	return payload, nil
}

func validateSetMilitaryDesignVisualPayload(payload SetMilitaryDesignVisualPayload) error {
	if payload.DesignID == 0 {
		return fmt.Errorf("design_id must be non-zero")
	}
	if err := core.ValidateShipVisualGenome(payload.VisualGenome); err != nil {
		return fmt.Errorf("visual_genome: %w", err)
	}
	return nil
}

// ResolveMilitaryDesignVisualCommand mutates only presentation identity. It
// deliberately leaves ShipDesign.Revision and ShipDesign.Spec untouched so a
// visual reroll cannot invalidate an in-progress gameplay construction order.
func ResolveMilitaryDesignVisualCommand(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) ([]DomainEvent, error) {
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	payload, err := decodeSetMilitaryDesignVisual(command)
	if err != nil {
		return nil, err
	}
	_, design := shipDesignByID(state, payload.DesignID)
	if design == nil {
		return nil, fmt.Errorf("references unknown ship design %d", payload.DesignID)
	}
	if design.EmpireID != empireID {
		return nil, fmt.Errorf("seat %d cannot update ship design %d owned by empire %d", seatID, design.ID, design.EmpireID)
	}
	if design.VisualRevision == ^uint32(0) {
		return nil, fmt.Errorf("ship design %d visual revision overflow", design.ID)
	}
	genome := core.CloneShipVisualGenome(payload.VisualGenome)
	design.VisualRevision++
	design.VisualGenome = &genome
	event, err := NewDomainEvent("empire.military_design_visual_updated", seatID, command.Sequence, MilitaryDesignVisualUpdatedEvent{
		DesignID:       design.ID,
		VisualRevision: design.VisualRevision,
		VisualGenome:   core.CloneShipVisualGenome(genome),
	})
	if err != nil {
		return nil, err
	}
	return []DomainEvent{event}, nil
}
