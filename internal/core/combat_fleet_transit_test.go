package core

import (
	"reflect"
	"testing"
)

func newCombatTransitState(seed uint64) (*GameState, ID) {
	state := NewSmallFixture(seed)
	empireID := state.Empires[0].ID
	designID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, ShipDesign{
		ID: designID, EmpireID: empireID, Revision: 1, Name: "Transit", Spec: baselineMilitarySpec(),
	})
	shipID := state.NewID()
	state.Ships = append(state.Ships, Ship{
		ID: shipID, EmpireID: empireID, SourceDesignID: designID, SourceDesignRevision: 1, Name: "Transit", Spec: baselineMilitarySpec(),
	})
	return state, shipID
}

func TestCombatFleetTransitValidationAndRoundTripSchema21(t *testing.T) {
	state, shipID := newCombatTransitState(1920)
	fleetID := state.NewID()
	state.StrategicFleets = []StrategicFleet{{
		ID:                  fleetID,
		EmpireID:            state.Empires[0].ID,
		Role:                StrategicFleetRoleCombat,
		DestinationSystemID: state.Galaxy.Systems[1].ID,
		RemainingTurns:      2,
		ShipIDs:             []ID{shipID},
	}}
	if StateSchemaVersion != 21 || state.SchemaVersion != 21 {
		t.Fatalf("schema=%d constant=%d want=21", state.SchemaVersion, StateSchemaVersion)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("valid combat transit rejected: %v", err)
	}
	encoded, err := MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(state.StrategicFleets, loaded.StrategicFleets) {
		t.Fatalf("combat transit changed across roundtrip: want=%+v got=%+v", state.StrategicFleets, loaded.StrategicFleets)
	}
}

func TestCombatFleetTransitValidationRejectsLocationlessAndCachedFTLSpeed(t *testing.T) {
	t.Run("locationless", func(t *testing.T) {
		state, shipID := newCombatTransitState(1921)
		state.StrategicFleets = []StrategicFleet{{
			ID: state.NewID(), EmpireID: state.Empires[0].ID, Role: StrategicFleetRoleCombat, ShipIDs: []ID{shipID},
		}}
		if err := state.Validate(); err == nil {
			t.Fatal("expected locationless combat Fleet to fail validation")
		}
	})

	t.Run("cached ftl speed", func(t *testing.T) {
		state, shipID := newCombatTransitState(1922)
		state.StrategicFleets = []StrategicFleet{{
			ID: state.NewID(), EmpireID: state.Empires[0].ID, Role: StrategicFleetRoleCombat,
			DestinationSystemID: state.Galaxy.Systems[1].ID, RemainingTurns: 1, FTLSpeed: 2, ShipIDs: []ID{shipID},
		}}
		if err := state.Validate(); err == nil {
			t.Fatal("expected ordinary combat Fleet cached ftl_speed to fail validation")
		}
	})
}
