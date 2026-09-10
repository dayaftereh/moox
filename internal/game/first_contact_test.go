package game

import (
	"testing"

	"moox/internal/core"
)

func TestFleetArrivalEstablishesFirstContactAtPreviouslyVisitedSystem(t *testing.T) {
	state := core.NewSmallFixture(0xC017AC7)
	visitorID := state.Empires[0].ID
	otherID := state.NewID()
	state.Empires = append(state.Empires, core.Empire{ID: otherID, Name: "Darlok", RaceID: "darlok"})
	if len(state.Galaxy.Systems) < 2 {
		t.Fatal("small fixture needs at least two systems")
	}
	systemID := state.Galaxy.Systems[1].ID
	state.Empires[1].MarkSystemVisited(systemID)
	fleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID: fleetID, EmpireID: visitorID, Role: core.StrategicFleetRoleCivilian,
		SpecialKind: core.StrategicFleetSpecialColonyShip, DestinationSystemID: systemID, RemainingTurns: 1, FTLSpeed: 2,
	})
	events, err := (&EconomyResolver{}).advanceStrategicFleetTransit(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != "empire.fleet_arrived" {
		t.Fatalf("arrival events=%+v", events)
	}
	if !state.Empires[0].HasVisitedSystem(systemID) {
		t.Fatalf("arrival did not persist visited system %d", systemID)
	}
	if !state.EmpiresHaveContact(visitorID, otherID) || !state.Empires[0].HasKnownEmpire(otherID) || !state.Empires[1].HasKnownEmpire(visitorID) {
		t.Fatalf("arrival did not persist symmetric first contact: %+v", state.Empires)
	}
}
