package protocol

import "encoding/json"

type DraftTelemetry struct {
	Sequence uint64          `json:"sequence"`
	Turn     uint64          `json:"turn"`
	SeatID   SeatID          `json:"seat_id"`
	Kind     string          `json:"kind"`
	Summary  string          `json:"summary,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

func CloneDraftTelemetry(event DraftTelemetry) DraftTelemetry {
	clone := event
	clone.Data = append(json.RawMessage(nil), event.Data...)
	return clone
}
