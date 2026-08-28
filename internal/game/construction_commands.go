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
const CommandQueueFreighterFleet = "colony.queue_freighter_fleet"

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
