package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"moox/internal/core"
	"moox/internal/protocol"
)

const CommandMoveFleet = "empire.move_fleet"
const CommandColonizePlanet = "empire.colonize_planet"

type MoveFleetPayload struct {
	FleetID             core.ID `json:"fleet_id"`
	DestinationSystemID core.ID `json:"destination_system_id"`
}

func NewMoveFleetCommand(sequence uint32, payload MoveFleetPayload) (protocol.Command, error) {
	if err := validateMoveFleetPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandMoveFleet, payload)
}

func decodeMoveFleet(command protocol.Command) (MoveFleetPayload, error) {
	var payload MoveFleetPayload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return MoveFleetPayload{}, fmt.Errorf("decode %s: %w", CommandMoveFleet, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return MoveFleetPayload{}, fmt.Errorf("decode %s: trailing JSON value", CommandMoveFleet)
		}
		return MoveFleetPayload{}, fmt.Errorf("decode %s trailing data: %w", CommandMoveFleet, err)
	}
	if err := validateMoveFleetPayload(payload); err != nil {
		return MoveFleetPayload{}, err
	}
	return payload, nil
}

func validateMoveFleetPayload(payload MoveFleetPayload) error {
	if payload.FleetID == 0 {
		return fmt.Errorf("fleet_id must be non-zero")
	}
	if payload.DestinationSystemID == 0 {
		return fmt.Errorf("destination_system_id must be non-zero")
	}
	return nil
}

type ColonizePlanetPayload struct {
	FleetID  core.ID `json:"fleet_id"`
	PlanetID core.ID `json:"planet_id"`
}

func NewColonizePlanetCommand(sequence uint32, payload ColonizePlanetPayload) (protocol.Command, error) {
	if err := validateColonizePlanetPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandColonizePlanet, payload)
}

func decodeColonizePlanet(command protocol.Command) (ColonizePlanetPayload, error) {
	var payload ColonizePlanetPayload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return ColonizePlanetPayload{}, fmt.Errorf("decode %s: %w", CommandColonizePlanet, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return ColonizePlanetPayload{}, fmt.Errorf("decode %s: trailing JSON value", CommandColonizePlanet)
		}
		return ColonizePlanetPayload{}, fmt.Errorf("decode %s trailing data: %w", CommandColonizePlanet, err)
	}
	if err := validateColonizePlanetPayload(payload); err != nil {
		return ColonizePlanetPayload{}, err
	}
	return payload, nil
}

func validateColonizePlanetPayload(payload ColonizePlanetPayload) error {
	if payload.FleetID == 0 {
		return fmt.Errorf("fleet_id must be non-zero")
	}
	if payload.PlanetID == 0 {
		return fmt.Errorf("planet_id must be non-zero")
	}
	return nil
}
