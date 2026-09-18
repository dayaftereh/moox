package game

import (
	"fmt"

	"moox/internal/core"
	"moox/internal/protocol"
)

type ReferenceMilitaryShipFixture struct {
	DesignID       core.ID
	DesignRevision uint32
	ShipID         core.ID
	FleetID        core.ID
	SystemID       core.ID
}

func (r *EconomyResolver) MaterializeReferenceMilitaryShipFixture(
	state *core.GameState,
	empireID core.ID,
	seatID protocol.SeatID,
	systemID core.ID,
	payload SaveMilitaryDesignPayload,
) (ReferenceMilitaryShipFixture, error) {
	if r == nil || r.Rules == nil {
		return ReferenceMilitaryShipFixture{}, fmt.Errorf("reference military Ship fixture requires economy resolver")
	}
	if state == nil {
		return ReferenceMilitaryShipFixture{}, fmt.Errorf("reference military Ship fixture requires state")
	}
	if payload.DesignID != 0 {
		return ReferenceMilitaryShipFixture{}, fmt.Errorf("reference military Ship fixture requires a new design")
	}
	command, err := NewSaveMilitaryDesignCommand(1, payload)
	if err != nil {
		return ReferenceMilitaryShipFixture{}, err
	}
	beforeDesigns := len(state.ShipDesigns)
	if _, err := r.ResolveMilitaryDesignCommand(state, empireID, seatID, command); err != nil {
		return ReferenceMilitaryShipFixture{}, err
	}
	if len(state.ShipDesigns) != beforeDesigns+1 {
		return ReferenceMilitaryShipFixture{}, fmt.Errorf("reference military Ship fixture did not create exactly one design")
	}
	design := &state.ShipDesigns[len(state.ShipDesigns)-1]
	ship, fleet, err := materializeMilitaryShip(state, empireID, systemID, design)
	if err != nil {
		return ReferenceMilitaryShipFixture{}, err
	}
	if err := state.Validate(); err != nil {
		return ReferenceMilitaryShipFixture{}, fmt.Errorf("validate reference military Ship fixture: %w", err)
	}
	return ReferenceMilitaryShipFixture{
		DesignID: design.ID, DesignRevision: design.Revision,
		ShipID: ship.ID, FleetID: fleet.ID, SystemID: systemID,
	}, nil
}
