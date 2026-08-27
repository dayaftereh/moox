package game

import (
	"encoding/json"
	"fmt"

	"moox/internal/core"
	"moox/internal/protocol"
)

type DomainEvent struct {
	Kind            string
	SeatID          protocol.SeatID
	CommandSequence uint32
	Data            json.RawMessage
}

type Encounter struct {
	Participants []protocol.SeatID
}

type Resolution struct {
	State      *core.GameState
	Events     []DomainEvent
	Encounters []Encounter
}

type Resolver interface {
	Resolve(state *core.GameState, batches []protocol.CommandBatch) (Resolution, error)
}

type ResolverFunc func(state *core.GameState, batches []protocol.CommandBatch) (Resolution, error)

func (f ResolverFunc) Resolve(state *core.GameState, batches []protocol.CommandBatch) (Resolution, error) {
	return f(state, batches)
}

func NewDomainEvent(kind string, seatID protocol.SeatID, commandSequence uint32, data any) (DomainEvent, error) {
	if kind == "" {
		return DomainEvent{}, fmt.Errorf("domain event kind must not be empty")
	}
	var raw json.RawMessage
	if data != nil {
		encoded, err := json.Marshal(data)
		if err != nil {
			return DomainEvent{}, fmt.Errorf("encode domain event data: %w", err)
		}
		raw = encoded
	}
	return DomainEvent{Kind: kind, SeatID: seatID, CommandSequence: commandSequence, Data: raw}, nil
}
