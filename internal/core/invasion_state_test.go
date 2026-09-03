package core

import "testing"

func TestSchema23GroundForcesCapitalOwnershipAndTroopTransportRoundTrip(t *testing.T) {
	state := NewSmallFixture(0xB310)
	if StateSchemaVersion != 23 || state.SchemaVersion != 23 {
		t.Fatalf("schema=%d state=%d", StateSchemaVersion, state.SchemaVersion)
	}
	state.Colonies[0].GroundForces.Infantry = 3
	transportID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, StrategicFleet{
		ID: transportID, EmpireID: state.Empires[0].ID, Role: StrategicFleetRoleCivilian,
		SpecialKind: StrategicFleetSpecialTroopTransport, AtSystemID: state.Galaxy.Systems[0].ID, FTLSpeed: 2,
	})
	encoded, err := MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SchemaVersion != 23 || loaded.Colonies[0].GroundForces.Infantry != 3 {
		t.Fatalf("roundtrip schema/ground=%d/%+v", loaded.SchemaVersion, loaded.Colonies[0].GroundForces)
	}
	if len(loaded.StrategicFleets) == 0 || loaded.StrategicFleets[len(loaded.StrategicFleets)-1].SpecialKind != StrategicFleetSpecialTroopTransport {
		t.Fatalf("transport roundtrip=%+v", loaded.StrategicFleets)
	}
}

func TestSchema23RejectsNegativeGroundInfantry(t *testing.T) {
	state := NewSmallFixture(0xB311)
	state.Colonies[0].GroundForces.Infantry = -1
	if err := state.Validate(); err == nil {
		t.Fatal("negative ground Infantry accepted")
	}
}

func TestSchema23RejectsForeignOwnedCapital(t *testing.T) {
	state := NewSmallFixture(0xB312)
	secondID := state.NewID()
	state.Empires = append(state.Empires, Empire{ID: secondID, Name: "Second", RaceID: "human", Capital: state.Colonies[0].ID})
	if err := state.Validate(); err == nil {
		t.Fatal("foreign-owned Capital accepted")
	}
}
