package game

import (
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func TestNewDomainEvent(t *testing.T) {
	event, err := NewDomainEvent("test.resolved", protocol.SeatID(2), 3, map[string]any{"value": 9})
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != "test.resolved" || event.SeatID != 2 || event.CommandSequence != 3 || len(event.Data) == 0 {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestResolveContextMapsServerSeatAuthority(t *testing.T) {
	ctx := ResolveContext{Seats: []SeatAuthority{{SeatID: 2, EmpireID: core.ID(17)}}}
	if empireID, ok := ctx.EmpireForSeat(2); !ok || empireID != 17 {
		t.Fatalf("seat authority lookup = %d, %v", empireID, ok)
	}
	if _, ok := ctx.EmpireForSeat(3); ok {
		t.Fatal("unknown seat unexpectedly resolved")
	}
}
