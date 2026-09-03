package core

import (
	"reflect"
	"testing"
)

func addStrategicTestEmpire(state *GameState, name string) ID {
	id := state.NewID()
	state.Empires = append(state.Empires, Empire{ID: id, Name: name, RaceID: "human"})
	return id
}

func TestStrategicStateRoundTripsAndDefaultsMissingRelationsToNeutral(t *testing.T) {
	state := NewSmallFixture(1501)
	firstEmpireID := state.Empires[0].ID
	secondEmpireID := addStrategicTestEmpire(state, "Second Empire")
	designID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, ShipDesign{ID: designID, EmpireID: firstEmpireID, Revision: 1, Name: "Blockader", Spec: baselineMilitarySpec()})
	shipID := state.NewID()
	state.Ships = append(state.Ships, Ship{ID: shipID, EmpireID: firstEmpireID, SourceDesignID: designID, SourceDesignRevision: 1, Name: "Blockader", Spec: baselineMilitarySpec()})
	fleetID := state.NewID()
	state.StrategicFleets = []StrategicFleet{{
		ID:         fleetID,
		EmpireID:   firstEmpireID,
		Role:       StrategicFleetRoleCombat,
		AtSystemID: state.Galaxy.Systems[1].ID,
		ShipIDs:    []ID{shipID},
	}}
	state.DiplomaticRelations = reciprocalRelations(firstEmpireID, secondEmpireID, DiplomaticStanceWar)

	if got := state.DiplomaticStanceBetween(firstEmpireID, secondEmpireID); got != DiplomaticStanceWar {
		t.Fatalf("directed stance=%q want=%q", got, DiplomaticStanceWar)
	}
	if got := state.DiplomaticStanceBetween(secondEmpireID, firstEmpireID); got != DiplomaticStanceWar {
		t.Fatalf("reciprocal stance=%q want=%q", got, DiplomaticStanceWar)
	}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
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
		t.Fatalf("strategic fleets changed across roundtrip: want=%+v got=%+v", state.StrategicFleets, loaded.StrategicFleets)
	}
	if !reflect.DeepEqual(state.DiplomaticRelations, loaded.DiplomaticRelations) {
		t.Fatalf("diplomatic relations changed across roundtrip: want=%+v got=%+v", state.DiplomaticRelations, loaded.DiplomaticRelations)
	}
}

func TestStrategicFleetValidation(t *testing.T) {
	t.Run("sorted ids", func(t *testing.T) {
		state := NewSmallFixture(1502)
		firstID := state.NewID()
		secondID := state.NewID()
		state.StrategicFleets = []StrategicFleet{
			{ID: secondID, EmpireID: state.Empires[0].ID, Role: StrategicFleetRoleCombat},
			{ID: firstID, EmpireID: state.Empires[0].ID, Role: StrategicFleetRoleCombat},
		}
		if err := state.Validate(); err == nil {
			t.Fatal("expected unsorted strategic fleet ids to fail validation")
		}
	})

	t.Run("known owner", func(t *testing.T) {
		state := NewSmallFixture(1503)
		state.StrategicFleets = []StrategicFleet{{ID: state.NewID(), EmpireID: 999999, Role: StrategicFleetRoleCombat}}
		if err := state.Validate(); err == nil {
			t.Fatal("expected unknown strategic fleet owner to fail validation")
		}
	})

	t.Run("valid role", func(t *testing.T) {
		state := NewSmallFixture(1504)
		state.StrategicFleets = []StrategicFleet{{ID: state.NewID(), EmpireID: state.Empires[0].ID, Role: StrategicFleetRole("transport")}}
		if err := state.Validate(); err == nil {
			t.Fatal("expected unsupported strategic fleet role to fail validation")
		}
	})

	t.Run("known system when stationary", func(t *testing.T) {
		state := NewSmallFixture(1505)
		state.StrategicFleets = []StrategicFleet{{ID: state.NewID(), EmpireID: state.Empires[0].ID, Role: StrategicFleetRoleCombat, AtSystemID: 999999}}
		if err := state.Validate(); err == nil {
			t.Fatal("expected unknown at-system reference to fail validation")
		}
	})
}

func TestDiplomaticRelationValidation(t *testing.T) {
	newState := func(seed uint64) (*GameState, ID, ID) {
		state := NewSmallFixture(seed)
		first := state.Empires[0].ID
		second := addStrategicTestEmpire(state, "Second Empire")
		return state, first, second
	}

	t.Run("directed pair cannot target self", func(t *testing.T) {
		state, first, _ := newState(1506)
		state.DiplomaticRelations = reciprocalRelations(first, first, DiplomaticStanceWar)
		if err := state.Validate(); err == nil {
			t.Fatal("expected self diplomatic relation to fail validation")
		}
	})

	t.Run("known source and target", func(t *testing.T) {
		state, first, _ := newState(1507)
		state.DiplomaticRelations = reciprocalRelations(first, 999999, DiplomaticStanceWar)
		if err := state.Validate(); err == nil {
			t.Fatal("expected unknown diplomatic target to fail validation")
		}
	})

	t.Run("valid stance", func(t *testing.T) {
		state, first, second := newState(1508)
		state.DiplomaticRelations = []DiplomaticRelation{{FromEmpireID: first, ToEmpireID: second, Stance: DiplomaticStance("hostile")}}
		if err := state.Validate(); err == nil {
			t.Fatal("expected unsupported diplomatic stance to fail validation")
		}
	})

	t.Run("strict pair ordering", func(t *testing.T) {
		state, first, second := newState(1509)
		state.DiplomaticRelations = []DiplomaticRelation{
			{FromEmpireID: second, ToEmpireID: first, Stance: DiplomaticStanceWar},
			{FromEmpireID: first, ToEmpireID: second, Stance: DiplomaticStanceWar},
		}
		if err := state.Validate(); err == nil {
			t.Fatal("expected unsorted diplomatic relations to fail validation")
		}
	})

	t.Run("duplicate pair", func(t *testing.T) {
		state, first, second := newState(1510)
		state.DiplomaticRelations = []DiplomaticRelation{
			{FromEmpireID: first, ToEmpireID: second, Stance: DiplomaticStanceNeutral},
			{FromEmpireID: first, ToEmpireID: second, Stance: DiplomaticStanceWar},
		}
		if err := state.Validate(); err == nil {
			t.Fatal("expected duplicate diplomatic relation pair to fail validation")
		}
	})
}

func TestColonyShipTransitStateRoundTripsInSchema23(t *testing.T) {
	state := NewSmallFixture(1510)
	fleetID := state.NewID()
	state.StrategicFleets = []StrategicFleet{{
		ID:                  fleetID,
		EmpireID:            state.Empires[0].ID,
		Role:                StrategicFleetRoleCivilian,
		SpecialKind:         StrategicFleetSpecialColonyShip,
		DestinationSystemID: state.Galaxy.Systems[1].ID,
		RemainingTurns:      2,
		FTLSpeed:            3,
	}}
	if StateSchemaVersion != 23 || state.SchemaVersion != 23 {
		t.Fatalf("schema=%d constant=%d want=22", state.SchemaVersion, StateSchemaVersion)
	}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
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
		t.Fatalf("Colony Ship transit changed across roundtrip: want=%+v got=%+v", state.StrategicFleets, loaded.StrategicFleets)
	}
}

func TestColonyShipTransitValidation(t *testing.T) {
	tests := []struct {
		name  string
		fleet func(*GameState) StrategicFleet
	}{
		{
			name: "must be civilian",
			fleet: func(state *GameState) StrategicFleet {
				return StrategicFleet{ID: state.NewID(), EmpireID: state.Empires[0].ID, Role: StrategicFleetRoleCombat, SpecialKind: StrategicFleetSpecialColonyShip, AtSystemID: state.Galaxy.Systems[0].ID, FTLSpeed: 2}
			},
		},
		{
			name: "requires installed drive",
			fleet: func(state *GameState) StrategicFleet {
				return StrategicFleet{ID: state.NewID(), EmpireID: state.Empires[0].ID, Role: StrategicFleetRoleCivilian, SpecialKind: StrategicFleetSpecialColonyShip, AtSystemID: state.Galaxy.Systems[0].ID}
			},
		},
		{
			name: "stationary cannot also transit",
			fleet: func(state *GameState) StrategicFleet {
				return StrategicFleet{ID: state.NewID(), EmpireID: state.Empires[0].ID, Role: StrategicFleetRoleCivilian, SpecialKind: StrategicFleetSpecialColonyShip, AtSystemID: state.Galaxy.Systems[0].ID, DestinationSystemID: state.Galaxy.Systems[1].ID, RemainingTurns: 1, FTLSpeed: 2}
			},
		},
		{
			name: "transit requires positive eta",
			fleet: func(state *GameState) StrategicFleet {
				return StrategicFleet{ID: state.NewID(), EmpireID: state.Empires[0].ID, Role: StrategicFleetRoleCivilian, SpecialKind: StrategicFleetSpecialColonyShip, DestinationSystemID: state.Galaxy.Systems[1].ID, FTLSpeed: 2}
			},
		},
		{
			name: "requires current or destination system",
			fleet: func(state *GameState) StrategicFleet {
				return StrategicFleet{ID: state.NewID(), EmpireID: state.Empires[0].ID, Role: StrategicFleetRoleCivilian, SpecialKind: StrategicFleetSpecialColonyShip, FTLSpeed: 2}
			},
		},
		{
			name: "ordinary fleet cannot use semantic transit",
			fleet: func(state *GameState) StrategicFleet {
				return StrategicFleet{ID: state.NewID(), EmpireID: state.Empires[0].ID, Role: StrategicFleetRoleCombat, DestinationSystemID: state.Galaxy.Systems[1].ID, RemainingTurns: 1, FTLSpeed: 2}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewSmallFixture(1520)
			state.StrategicFleets = []StrategicFleet{test.fleet(state)}
			if err := state.Validate(); err == nil {
				t.Fatalf("invalid strategic fleet unexpectedly passed validation: %+v", state.StrategicFleets[0])
			}
		})
	}
}
