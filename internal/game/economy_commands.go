package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"moox/internal/core"
	"moox/internal/protocol"
)

const CommandAssignPopulation = "colony.assign_population"

type AssignPopulationPayload struct {
	ColonyID   core.ID `json:"colony_id"`
	Farmers    int     `json:"farmers"`
	Workers    int     `json:"workers"`
	Scientists int     `json:"scientists"`
}

func NewAssignPopulationCommand(sequence uint32, payload AssignPopulationPayload) (protocol.Command, error) {
	if err := validateAssignPopulationPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandAssignPopulation, payload)
}

func decodeAssignPopulation(command protocol.Command) (AssignPopulationPayload, error) {
	var payload AssignPopulationPayload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return AssignPopulationPayload{}, fmt.Errorf("decode %s: %w", CommandAssignPopulation, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return AssignPopulationPayload{}, fmt.Errorf("decode %s: trailing JSON value", CommandAssignPopulation)
		}
		return AssignPopulationPayload{}, fmt.Errorf("decode %s trailing data: %w", CommandAssignPopulation, err)
	}
	if err := validateAssignPopulationPayload(payload); err != nil {
		return AssignPopulationPayload{}, err
	}
	return payload, nil
}

func validateAssignPopulationPayload(payload AssignPopulationPayload) error {
	if payload.ColonyID == 0 {
		return fmt.Errorf("colony_id must be non-zero")
	}
	if payload.Farmers < 0 || payload.Workers < 0 || payload.Scientists < 0 {
		return fmt.Errorf("population assignments must be non-negative")
	}
	return nil
}
