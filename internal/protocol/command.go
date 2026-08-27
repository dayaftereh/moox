package protocol

import (
	"encoding/json"
	"fmt"
)

const CommandSchemaVersion = 1

type SeatID uint32

type Command struct {
	SchemaVersion int             `json:"schema_version"`
	Sequence      uint32          `json:"sequence"`
	Kind          string          `json:"kind"`
	Payload       json.RawMessage `json:"payload,omitempty"`
}

type CommandBatch struct {
	SchemaVersion int       `json:"schema_version"`
	GameID        string    `json:"game_id"`
	SeatID        SeatID    `json:"seat_id"`
	Turn          uint64    `json:"turn"`
	BaseRevision  uint64    `json:"base_revision"`
	Commands      []Command `json:"commands"`
}

func NewCommand(sequence uint32, kind string, payload any) (Command, error) {
	var raw json.RawMessage
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return Command{}, fmt.Errorf("encode command payload: %w", err)
		}
		raw = encoded
	}
	command := Command{SchemaVersion: CommandSchemaVersion, Sequence: sequence, Kind: kind, Payload: raw}
	if err := command.Validate(sequence); err != nil {
		return Command{}, err
	}
	return command, nil
}

func (c Command) Validate(expectedSequence uint32) error {
	if c.SchemaVersion != CommandSchemaVersion {
		return fmt.Errorf("unsupported command schema version %d", c.SchemaVersion)
	}
	if c.Sequence == 0 || c.Sequence != expectedSequence {
		return fmt.Errorf("command sequence %d, expected %d", c.Sequence, expectedSequence)
	}
	if c.Kind == "" {
		return fmt.Errorf("command kind must not be empty")
	}
	if len(c.Payload) > 0 && !json.Valid(c.Payload) {
		return fmt.Errorf("command %d has invalid JSON payload", c.Sequence)
	}
	return nil
}

func (b CommandBatch) Validate() error {
	if b.SchemaVersion != CommandSchemaVersion {
		return fmt.Errorf("unsupported command batch schema version %d", b.SchemaVersion)
	}
	if b.GameID == "" {
		return fmt.Errorf("game_id must not be empty")
	}
	if b.SeatID == 0 {
		return fmt.Errorf("seat_id must be non-zero")
	}
	if b.Turn == 0 {
		return fmt.Errorf("turn must be non-zero")
	}
	if b.BaseRevision == 0 {
		return fmt.Errorf("base_revision must be non-zero")
	}
	for i := range b.Commands {
		if err := b.Commands[i].Validate(uint32(i + 1)); err != nil {
			return fmt.Errorf("commands[%d]: %w", i, err)
		}
	}
	return nil
}

func CloneCommandBatch(batch CommandBatch) CommandBatch {
	clone := batch
	clone.Commands = make([]Command, len(batch.Commands))
	for i := range batch.Commands {
		clone.Commands[i] = batch.Commands[i]
		clone.Commands[i].Payload = append(json.RawMessage(nil), batch.Commands[i].Payload...)
	}
	return clone
}
