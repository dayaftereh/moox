package game

import (
	"fmt"

	"moox/internal/core"
	"moox/internal/protocol"
)

type OutpostDeployedEvent struct {
	EmpireID  core.ID `json:"empire_id"`
	FleetID   core.ID `json:"fleet_id"`
	SystemID  core.ID `json:"system_id"`
	BodyID    core.ID `json:"body_id"`
	PlanetID  core.ID `json:"planet_id,omitempty"`
	OutpostID core.ID `json:"outpost_id"`
}

type OutpostShipConsumedEvent struct {
	EmpireID  core.ID `json:"empire_id"`
	FleetID   core.ID `json:"fleet_id"`
	BodyID    core.ID `json:"body_id"`
	PlanetID  core.ID `json:"planet_id,omitempty"`
	OutpostID core.ID `json:"outpost_id"`
}

type OutpostConvertedEvent struct {
	EmpireID  core.ID `json:"empire_id"`
	OutpostID core.ID `json:"outpost_id"`
	PlanetID  core.ID `json:"planet_id"`
	ColonyID  core.ID `json:"colony_id"`
}

func outpostByID(state *core.GameState, id core.ID) (int, *core.Outpost) {
	if state == nil || id == 0 {
		return -1, nil
	}
	for i := range state.Outposts {
		if state.Outposts[i].ID == id {
			return i, &state.Outposts[i]
		}
	}
	return -1, nil
}

func outpostReferencesPlanet(state *core.GameState, planetID core.ID) bool {
	if state == nil || planetID == 0 {
		return false
	}
	for i := range state.Outposts {
		if state.Outposts[i].PlanetID == planetID {
			return true
		}
	}
	return false
}

// commitFoundedColony applies the shared authoritative Planet occupancy change
// after prepareFoundedColony has produced a fully initialized Colony. A
// same-owner Outpost is removed atomically from the target Planet; foreign
// Outposts are rejected before any state mutation.
func commitFoundedColony(state *core.GameState, empireID core.ID, planet *core.Planet, colony core.Colony) (core.ID, error) {
	if state == nil || planet == nil {
		return 0, fmt.Errorf("state and target planet must not be nil")
	}
	if colony.ID == 0 || colony.EmpireID != empireID || colony.PlanetID != planet.ID {
		return 0, fmt.Errorf("prepared colony %d does not match empire %d planet %d", colony.ID, empireID, planet.ID)
	}
	if planet.ColonyID != 0 {
		return 0, fmt.Errorf("planet %d is already colonized by colony %d", planet.ID, planet.ColonyID)
	}
	convertedOutpostID := planet.OutpostID
	outpostIndex := -1
	if convertedOutpostID != 0 {
		index, outpost := outpostByID(state, convertedOutpostID)
		if outpost == nil {
			return 0, fmt.Errorf("planet %d references unknown outpost %d", planet.ID, convertedOutpostID)
		}
		if outpost.EmpireID != empireID {
			return 0, fmt.Errorf("planet %d outpost %d is owned by empire %d", planet.ID, outpost.ID, outpost.EmpireID)
		}
		if outpost.PlanetID != 0 && outpost.PlanetID != planet.ID {
			return 0, fmt.Errorf("outpost %d references planet %d, not target planet %d", outpost.ID, outpost.PlanetID, planet.ID)
		}
		if targetID := outpostTargetBodyID(*outpost); targetID != planet.ID {
			return 0, fmt.Errorf("outpost %d targets body %d, not target planet body %d", outpost.ID, targetID, planet.ID)
		}
		outpostIndex = index
	} else if outpostReferencesPlanet(state, planet.ID) {
		return 0, fmt.Errorf("planet %d is referenced by an Outpost but has no outpost_id link", planet.ID)
	}
	if allocated := state.NewID(); allocated != colony.ID {
		return 0, fmt.Errorf("allocated Colony ID %d does not match expected next_id %d", allocated, colony.ID)
	}
	if outpostIndex >= 0 {
		state.Outposts = append(state.Outposts[:outpostIndex], state.Outposts[outpostIndex+1:]...)
		planet.OutpostID = 0
		target := orbitalBodyTargetByID(state, planet.ID)
		if target.Body != nil {
			target.Body.OutpostID = 0
		}
	}
	planet.ColonyID = colony.ID
	state.Colonies = append(state.Colonies, colony)
	return convertedOutpostID, nil
}

func (r *EconomyResolver) deployOutpost(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) ([]DomainEvent, error) {
	payload, err := decodeDeployOutpost(command)
	if err != nil {
		return nil, err
	}
	fleetIndex, fleet := strategicFleetByID(state, payload.FleetID)
	if fleet == nil {
		return nil, fmt.Errorf("references unknown strategic fleet %d", payload.FleetID)
	}
	if fleet.EmpireID != empireID {
		return nil, fmt.Errorf("seat %d cannot deploy Outpost with fleet %d owned by empire %d", seatID, fleet.ID, fleet.EmpireID)
	}
	if fleet.SpecialKind != core.StrategicFleetSpecialOutpostShip {
		return nil, fmt.Errorf("fleet %d is not an Outpost Ship", fleet.ID)
	}
	if fleet.AtSystemID == 0 || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 {
		return nil, fmt.Errorf("Outpost Ship fleet %d must be stationary before deployment", fleet.ID)
	}

	bodyID := payload.TargetBodyID()
	target := orbitalBodyTargetByID(state, bodyID)
	if target.System == nil {
		return nil, fmt.Errorf("references unknown orbital body %d", bodyID)
	}
	if target.System.ID != fleet.AtSystemID {
		return nil, fmt.Errorf("body %d is in system %d but Outpost Ship fleet %d is at system %d", bodyID, target.System.ID, fleet.ID, fleet.AtSystemID)
	}
	if target.Planet != nil && (target.Planet.ColonyID != 0 || colonyReferencesPlanet(state, target.Planet.ID)) {
		return nil, fmt.Errorf("planet body %d is already occupied by a Colony", bodyID)
	}
	if orbitalBodyOccupiedByOutpost(state, target) {
		return nil, fmt.Errorf("body %d already has an Outpost", bodyID)
	}
	if empireByID(state, empireID) == nil {
		return nil, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}

	planetID := core.ID(0)
	if target.Planet != nil {
		planetID = target.Planet.ID
	}
	outpostID := state.NextID
	if outpostID == 0 {
		return nil, fmt.Errorf("cannot allocate Outpost ID from zero next_id")
	}
	deployed, err := NewDomainEvent("empire.outpost_deployed", seatID, command.Sequence, OutpostDeployedEvent{EmpireID: empireID, FleetID: fleet.ID, SystemID: target.System.ID, BodyID: bodyID, PlanetID: planetID, OutpostID: outpostID})
	if err != nil {
		return nil, err
	}
	consumed, err := NewDomainEvent("empire.outpost_ship_consumed", seatID, command.Sequence, OutpostShipConsumedEvent{EmpireID: empireID, FleetID: fleet.ID, BodyID: bodyID, PlanetID: planetID, OutpostID: outpostID})
	if err != nil {
		return nil, err
	}
	if allocated := state.NewID(); allocated != outpostID {
		return nil, fmt.Errorf("allocated Outpost ID %d does not match expected next_id %d", allocated, outpostID)
	}
	state.Outposts = append(state.Outposts, core.Outpost{ID: outpostID, EmpireID: empireID, BodyID: bodyID, PlanetID: planetID})
	setOrbitalBodyOutpost(target, outpostID)
	state.StrategicFleets = append(state.StrategicFleets[:fleetIndex], state.StrategicFleets[fleetIndex+1:]...)
	return []DomainEvent{deployed, consumed}, nil
}
