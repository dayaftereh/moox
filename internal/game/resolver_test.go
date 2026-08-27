package game

import (
	"testing"

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
