package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"

	"moox/internal/core"
	"moox/internal/protocol"
)

const CommandAssignPopulation = "colony.assign_population"

type AssignPopulationPayload struct {
	ColonyID   core.ID `json:"colony_id"`
	Farmers    float64 `json:"farmers"`
	Workers    float64 `json:"workers"`
	Scientists float64 `json:"scientists"`
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
	for _, value := range []float64{payload.Farmers, payload.Workers, payload.Scientists} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return fmt.Errorf("population assignments must be finite and non-negative")
		}
	}
	return nil
}
