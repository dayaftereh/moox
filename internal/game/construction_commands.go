package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"moox/internal/core"
	"moox/internal/protocol"
)

const CommandQueueBuilding = "colony.queue_building"
const CommandQueueColonyShip = "colony.queue_colony_ship"
const CommandQueueOutpostShip = "colony.queue_outpost_ship"
const CommandQueueTroopTransport = "colony.queue_troop_transport"
const CommandQueueMilitaryShip = "colony.queue_military_ship"
const CommandQueueFreighterFleet = "colony.queue_freighter_fleet"
const CommandQueueHousing = "colony.queue_housing"
const CommandQueuePlanetaryTransformation = "colony.queue_planetary_transformation"

type QueueBuildingPayload struct {
	ColonyID   core.ID `json:"colony_id"`
	BuildingID string  `json:"building_id"`
}

func NewQueueBuildingCommand(sequence uint32, payload QueueBuildingPayload) (protocol.Command, error) {
	if err := validateQueueBuildingPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandQueueBuilding, payload)
}

func decodeQueueBuilding(command protocol.Command) (QueueBuildingPayload, error) {
	var payload QueueBuildingPayload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return QueueBuildingPayload{}, fmt.Errorf("decode %s: %w", CommandQueueBuilding, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return QueueBuildingPayload{}, fmt.Errorf("decode %s: trailing JSON value", CommandQueueBuilding)
		}
		return QueueBuildingPayload{}, fmt.Errorf("decode %s trailing data: %w", CommandQueueBuilding, err)
	}
	if err := validateQueueBuildingPayload(payload); err != nil {
		return QueueBuildingPayload{}, err
	}
	return payload, nil
}

func validateQueueBuildingPayload(payload QueueBuildingPayload) error {
	if payload.ColonyID == 0 {
		return fmt.Errorf("colony_id must be non-zero")
	}
	if payload.BuildingID == "" {
		return fmt.Errorf("building_id must not be empty")
	}
	return nil
}

type QueueHousingPayload struct {
	ColonyID core.ID `json:"colony_id"`
}

func NewQueueHousingCommand(sequence uint32, payload QueueHousingPayload) (protocol.Command, error) {
	if payload.ColonyID == 0 {
		return protocol.Command{}, fmt.Errorf("colony_id must be non-zero")
	}
	return protocol.NewCommand(sequence, CommandQueueHousing, payload)
}

func decodeQueueHousing(command protocol.Command) (QueueHousingPayload, error) {
	var payload QueueHousingPayload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return QueueHousingPayload{}, fmt.Errorf("decode %s: %w", CommandQueueHousing, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return QueueHousingPayload{}, fmt.Errorf("decode %s: trailing JSON value", CommandQueueHousing)
		}
		return QueueHousingPayload{}, fmt.Errorf("decode %s trailing data: %w", CommandQueueHousing, err)
	}
	if payload.ColonyID == 0 {
		return QueueHousingPayload{}, fmt.Errorf("colony_id must be non-zero")
	}
	return payload, nil
}

type QueuePlanetaryTransformationPayload struct {
	ColonyID  core.ID `json:"colony_id"`
	ProjectID string  `json:"project_id"`
}

func NewQueuePlanetaryTransformationCommand(sequence uint32, payload QueuePlanetaryTransformationPayload) (protocol.Command, error) {
	if err := validateQueuePlanetaryTransformationPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandQueuePlanetaryTransformation, payload)
}

func decodeQueuePlanetaryTransformation(command protocol.Command) (QueuePlanetaryTransformationPayload, error) {
	var payload QueuePlanetaryTransformationPayload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return QueuePlanetaryTransformationPayload{}, fmt.Errorf("decode %s: %w", CommandQueuePlanetaryTransformation, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return QueuePlanetaryTransformationPayload{}, fmt.Errorf("decode %s: trailing JSON value", CommandQueuePlanetaryTransformation)
		}
		return QueuePlanetaryTransformationPayload{}, fmt.Errorf("decode %s trailing data: %w", CommandQueuePlanetaryTransformation, err)
	}
	if err := validateQueuePlanetaryTransformationPayload(payload); err != nil {
		return QueuePlanetaryTransformationPayload{}, err
	}
	return payload, nil
}

func validateQueuePlanetaryTransformationPayload(payload QueuePlanetaryTransformationPayload) error {
	if payload.ColonyID == 0 {
		return fmt.Errorf("colony_id must be non-zero")
	}
	if payload.ProjectID == "" {
		return fmt.Errorf("project_id must not be empty")
	}
	return nil
}

type QueueColonyShipPayload struct {
	ColonyID core.ID `json:"colony_id"`
}

func NewQueueColonyShipCommand(sequence uint32, payload QueueColonyShipPayload) (protocol.Command, error) {
	if payload.ColonyID == 0 {
		return protocol.Command{}, fmt.Errorf("colony_id must be non-zero")
	}
	return protocol.NewCommand(sequence, CommandQueueColonyShip, payload)
}

func decodeQueueColonyShip(command protocol.Command) (QueueColonyShipPayload, error) {
	var payload QueueColonyShipPayload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return QueueColonyShipPayload{}, fmt.Errorf("decode %s: %w", CommandQueueColonyShip, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return QueueColonyShipPayload{}, fmt.Errorf("decode %s: trailing JSON value", CommandQueueColonyShip)
		}
		return QueueColonyShipPayload{}, fmt.Errorf("decode %s trailing data: %w", CommandQueueColonyShip, err)
	}
	if payload.ColonyID == 0 {
		return QueueColonyShipPayload{}, fmt.Errorf("colony_id must be non-zero")
	}
	return payload, nil
}

type QueueOutpostShipPayload struct {
	ColonyID core.ID `json:"colony_id"`
}

func NewQueueOutpostShipCommand(sequence uint32, payload QueueOutpostShipPayload) (protocol.Command, error) {
	if payload.ColonyID == 0 {
		return protocol.Command{}, fmt.Errorf("colony_id must be non-zero")
	}
	return protocol.NewCommand(sequence, CommandQueueOutpostShip, payload)
}

func decodeQueueOutpostShip(command protocol.Command) (QueueOutpostShipPayload, error) {
	var payload QueueOutpostShipPayload
	if err := decodeStrictCommandPayload(command, CommandQueueOutpostShip, &payload); err != nil {
		return QueueOutpostShipPayload{}, err
	}
	if payload.ColonyID == 0 {
		return QueueOutpostShipPayload{}, fmt.Errorf("colony_id must be non-zero")
	}
	return payload, nil
}

type QueueTroopTransportPayload struct {
	ColonyID core.ID `json:"colony_id"`
}

func NewQueueTroopTransportCommand(sequence uint32, payload QueueTroopTransportPayload) (protocol.Command, error) {
	if payload.ColonyID == 0 {
		return protocol.Command{}, fmt.Errorf("colony_id must be non-zero")
	}
	return protocol.NewCommand(sequence, CommandQueueTroopTransport, payload)
}

func decodeQueueTroopTransport(command protocol.Command) (QueueTroopTransportPayload, error) {
	var payload QueueTroopTransportPayload
	if err := decodeStrictCommandPayload(command, CommandQueueTroopTransport, &payload); err != nil {
		return QueueTroopTransportPayload{}, err
	}
	if payload.ColonyID == 0 {
		return QueueTroopTransportPayload{}, fmt.Errorf("colony_id must be non-zero")
	}
	return payload, nil
}

type QueueFreighterFleetPayload struct {
	ColonyID core.ID `json:"colony_id"`
}

func NewQueueFreighterFleetCommand(sequence uint32, payload QueueFreighterFleetPayload) (protocol.Command, error) {
	if payload.ColonyID == 0 {
		return protocol.Command{}, fmt.Errorf("colony_id must be non-zero")
	}
	return protocol.NewCommand(sequence, CommandQueueFreighterFleet, payload)
}

func decodeQueueFreighterFleet(command protocol.Command) (QueueFreighterFleetPayload, error) {
	var payload QueueFreighterFleetPayload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return QueueFreighterFleetPayload{}, fmt.Errorf("decode %s: %w", CommandQueueFreighterFleet, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return QueueFreighterFleetPayload{}, fmt.Errorf("decode %s: trailing JSON value", CommandQueueFreighterFleet)
		}
		return QueueFreighterFleetPayload{}, fmt.Errorf("decode %s trailing data: %w", CommandQueueFreighterFleet, err)
	}
	if payload.ColonyID == 0 {
		return QueueFreighterFleetPayload{}, fmt.Errorf("colony_id must be non-zero")
	}
	return payload, nil
}

type QueueMilitaryShipPayload struct {
	ColonyID     core.ID `json:"colony_id"`
	ShipDesignID core.ID `json:"ship_design_id"`
}

func NewQueueMilitaryShipCommand(sequence uint32, payload QueueMilitaryShipPayload) (protocol.Command, error) {
	if err := validateQueueMilitaryShipPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandQueueMilitaryShip, payload)
}

func decodeQueueMilitaryShip(command protocol.Command) (QueueMilitaryShipPayload, error) {
	var payload QueueMilitaryShipPayload
	if err := decodeStrictCommandPayload(command, CommandQueueMilitaryShip, &payload); err != nil {
		return QueueMilitaryShipPayload{}, err
	}
	if err := validateQueueMilitaryShipPayload(payload); err != nil {
		return QueueMilitaryShipPayload{}, err
	}
	return payload, nil
}

func validateQueueMilitaryShipPayload(payload QueueMilitaryShipPayload) error {
	if payload.ColonyID == 0 {
		return fmt.Errorf("colony_id must be non-zero")
	}
	if payload.ShipDesignID == 0 {
		return fmt.Errorf("ship_design_id must be non-zero")
	}
	return nil
}
