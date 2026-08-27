package protocol

import "encoding/json"

const EventSchemaVersion = 1

type EventScope string

const (
	EventScopeSession   EventScope = "session"
	EventScopeStrategic EventScope = "strategic"
	EventScopeBattle    EventScope = "battle"
)

type DomainEvent struct {
	SchemaVersion   int             `json:"schema_version"`
	Sequence        uint64          `json:"sequence"`
	Turn            uint64          `json:"turn"`
	Revision        uint64          `json:"revision"`
	Scope           EventScope      `json:"scope"`
	Kind            string          `json:"kind"`
	SeatID          SeatID          `json:"seat_id,omitempty"`
	CommandSequence uint32          `json:"command_sequence,omitempty"`
	Data            json.RawMessage `json:"data,omitempty"`
}

func CloneEvent(event DomainEvent) DomainEvent {
	clone := event
	clone.Data = append(json.RawMessage(nil), event.Data...)
	return clone
}
