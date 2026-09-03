package game

import (
	"reflect"
	"testing"

	"moox/internal/core"
)

func TestCombatFleetTransitRoundTripsSchema23(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1910)
	empire := &state.Empires[0]
	setCombatFleetTestTech(empire)
	source := &state.Galaxy.Systems[0]
	destination := &state.Galaxy.Systems[1]
	destination.X = source.X + 150
	destination.Y = source.Y
	fleetID, shipIDs := addCombatFleetTestFleet(state, empire.ID, source.ID, 2)

	command, err := NewMoveFleetCommand(1, MoveFleetPayload{
		FleetID:             fleetID,
		DestinationSystemID: destination.ID,
		ShipIDs:             []core.ID{shipIDs[1]},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.moveFleetEvents(state, empire.ID, 1, command); err != nil {
		t.Fatal(err)
	}
	if core.StateSchemaVersion != 23 || state.SchemaVersion != 23 {
		t.Fatalf("schema=%d constant=%d want=22", state.SchemaVersion, core.StateSchemaVersion)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("schema21 combat transit invalid before save: %v", err)
	}

	encoded, err := core.MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := core.UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.StrategicFleets, loaded.StrategicFleets) {
		t.Fatalf("combat fleets changed across save/load: want=%+v got=%+v", state.StrategicFleets, loaded.StrategicFleets)
	}
	if !reflect.DeepEqual(state.Ships, loaded.Ships) {
		t.Fatalf("ships changed across save/load: want=%+v got=%+v", state.Ships, loaded.Ships)
	}
	if state.NextID != loaded.NextID {
		t.Fatalf("next_id changed across save/load: want=%d got=%d", state.NextID, loaded.NextID)
	}
}
