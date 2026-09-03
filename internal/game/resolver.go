package game

import (
	"encoding/json"
	"fmt"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/protocol"
)

type SeatAuthority struct {
	SeatID   protocol.SeatID
	EmpireID core.ID
}

type ResolveContext struct {
	Seats            []SeatAuthority
	HandledInvasions []InvasionHandledKey
}

func (c ResolveContext) EmpireForSeat(seatID protocol.SeatID) (core.ID, bool) {
	for _, seat := range c.Seats {
		if seat.SeatID == seatID {
			return seat.EmpireID, true
		}
	}
	return 0, false
}

func (c ResolveContext) SeatForEmpire(empireID core.ID) (protocol.SeatID, bool) {
	for _, seat := range c.Seats {
		if seat.EmpireID == empireID {
			return seat.SeatID, true
		}
	}
	return 0, false
}

type DomainEvent struct {
	Kind            string
	SeatID          protocol.SeatID
	CommandSequence uint32
	Data            json.RawMessage
}

type EncounterSide struct {
	EmpireID         core.ID         `json:"empire_id"`
	SeatID           protocol.SeatID `json:"seat_id"`
	CombatFleetIDs   []core.ID       `json:"combat_fleet_ids"`
	ShipIDs          []core.ID       `json:"ship_ids"`
	CivilianFleetIDs []core.ID       `json:"civilian_fleet_ids,omitempty"`
}

type Encounter struct {
	SystemID                  core.ID              `json:"system_id"`
	Attacker                  EncounterSide        `json:"attacker"`
	Defender                  EncounterSide        `json:"defender"`
	DefenderColonyIDs         []core.ID            `json:"defender_colony_ids,omitempty"`
	Participants              []protocol.SeatID    `json:"participants"`
	Tactical                  *battle.TacticalSpec `json:"tactical,omitempty"`
	TacticalUnsupportedReason string               `json:"tactical_unsupported_reason,omitempty"`
}

type EncounterOutcome struct {
	BattleID         uint64          `json:"battle_id"`
	Encounter        Encounter       `json:"encounter"`
	WinnerSeat       protocol.SeatID `json:"winner_seat"`
	Outcome          string          `json:"outcome"`
	DestroyedShipIDs []core.ID       `json:"destroyed_ship_ids,omitempty"`
}

type Resolution struct {
	State      *core.GameState
	Events     []DomainEvent
	Encounters []Encounter
	Invasion   *InvasionOpportunity
}

type Resolver interface {
	Resolve(ctx ResolveContext, state *core.GameState, batches []protocol.CommandBatch) (Resolution, error)
}

// EncounterResolver extends a normal strategic Resolver with the in-memory
// continuation needed when the strategic turn pauses at the encounter boundary.
// Active encounter lifecycle remains session state.
type EncounterResolver interface {
	Resolver
	ResumeAfterEncounters(ctx ResolveContext, state *core.GameState, outcomes []EncounterOutcome) (Resolution, error)
}

type InvasionResolver interface {
	ResolveInvasionCommand(ctx ResolveContext, state *core.GameState, opportunity InvasionOpportunity, seatID protocol.SeatID, command protocol.Command) ([]DomainEvent, error)
	ResumeAfterInvasion(ctx ResolveContext, state *core.GameState) (Resolution, error)
}

type ResolverFunc func(ctx ResolveContext, state *core.GameState, batches []protocol.CommandBatch) (Resolution, error)

func (f ResolverFunc) Resolve(ctx ResolveContext, state *core.GameState, batches []protocol.CommandBatch) (Resolution, error) {
	return f(ctx, state, batches)
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
