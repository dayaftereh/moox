package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"moox/internal/core"
	"moox/internal/protocol"
)

const CommandReferenceGrantBC = "reference.grant_bc"

type ReferenceGrantBCPayload struct {
	AmountBC int `json:"amount_bc"`
}

type ReferenceBCGrantedEvent struct {
	EmpireID          core.ID `json:"empire_id"`
	AmountBC          int     `json:"amount_bc"`
	PreviousBalanceBC float64 `json:"previous_balance_bc"`
	CurrentBalanceBC  float64 `json:"current_balance_bc"`
}

func NewReferenceGrantBCCommand(sequence uint32, payload ReferenceGrantBCPayload) (protocol.Command, error) {
	if !ReferenceGrantBCAmountAllowed(payload.AmountBC) {
		return protocol.Command{}, fmt.Errorf("%s amount_bc must be one of 100, 1000, 10000", CommandReferenceGrantBC)
	}
	return protocol.NewCommand(sequence, CommandReferenceGrantBC, payload)
}

func ReferenceGrantBCAmountAllowed(amount int) bool {
	return amount == 100 || amount == 1000 || amount == 10000
}

func decodeReferenceGrantBC(command protocol.Command) (ReferenceGrantBCPayload, error) {
	if command.Kind != CommandReferenceGrantBC {
		return ReferenceGrantBCPayload{}, fmt.Errorf("unsupported reference BC command %q", command.Kind)
	}
	var payload ReferenceGrantBCPayload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return ReferenceGrantBCPayload{}, fmt.Errorf("decode %s: %w", CommandReferenceGrantBC, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return ReferenceGrantBCPayload{}, fmt.Errorf("decode %s: trailing JSON value", CommandReferenceGrantBC)
		}
		return ReferenceGrantBCPayload{}, fmt.Errorf("decode %s trailing data: %w", CommandReferenceGrantBC, err)
	}
	if !ReferenceGrantBCAmountAllowed(payload.AmountBC) {
		return ReferenceGrantBCPayload{}, fmt.Errorf("%s amount_bc must be one of 100, 1000, 10000", CommandReferenceGrantBC)
	}
	return payload, nil
}

func ResolveReferenceGrantBCCommand(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) ([]DomainEvent, error) {
	if state == nil {
		return nil, fmt.Errorf("reference BC grant requires game state")
	}
	payload, err := decodeReferenceGrantBC(command)
	if err != nil {
		return nil, err
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return nil, fmt.Errorf("unknown empire %d", empireID)
	}
	previous := empire.Treasury.BalanceBC
	empire.Treasury.BalanceBC += float64(payload.AmountBC)
	event, err := NewDomainEvent("reference.bc_granted", seatID, command.Sequence, ReferenceBCGrantedEvent{
		EmpireID: empire.ID, AmountBC: payload.AmountBC,
		PreviousBalanceBC: previous, CurrentBalanceBC: empire.Treasury.BalanceBC,
	})
	if err != nil {
		return nil, err
	}
	return []DomainEvent{event}, nil
}
