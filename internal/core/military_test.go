package core

import (
	"bytes"
	"testing"
)

func baselineMilitarySpec() ShipDesignSpec {
	return ShipDesignSpec{
		HullID:             "frigate",
		StrategicPictureID: 0,
		WarpDriveID:        "nuclear_drive",
		FTLSpeed:           2,
		ComputerID:         "electronic_computer",
		ArmorID:            "titanium_armor",
		FuelCellID:         "standard_fuel_cells",
		FuelRangeParsecs:   4,
		HullBaseCostPP:     20,
		HullSpace:          25,
		SpaceUsed:          0,
		BaseDesignCostPP:   25,
		ProductionCostPP:   25,
	}
}

func TestMilitaryStateAllowsMoreThanSixDesignsAndRoundTripsSchema20(t *testing.T) {
	state := NewSmallFixture(1801)
	empireID := state.Empires[0].ID
	for i := 0; i < 8; i++ {
		state.ShipDesigns = append(state.ShipDesigns, ShipDesign{
			ID: state.NewID(), EmpireID: empireID, Revision: 1, Name: "Frigate design", Spec: baselineMilitarySpec(),
		})
	}
	shipID := state.NewID()
	shipSpec := state.ShipDesigns[0].Spec
	state.Ships = append(state.Ships, Ship{
		ID: shipID, EmpireID: empireID, SourceDesignID: state.ShipDesigns[0].ID, SourceDesignRevision: 1, Name: "Frigate", Spec: shipSpec,
	})
	fleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, StrategicFleet{
		ID: fleetID, EmpireID: empireID, Role: StrategicFleetRoleCombat, AtSystemID: state.Galaxy.Systems[0].ID, ShipIDs: []ID{shipID},
	})
	if StateSchemaVersion != 20 || state.SchemaVersion != 20 {
		t.Fatalf("schema=%d constant=%d want=20", state.SchemaVersion, StateSchemaVersion)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("schema20 military state invalid: %v", err)
	}
	encoded, err := MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.ShipDesigns) != 8 {
		t.Fatalf("decoded designs=%d want=8; MOOX must not inherit original six-slot ceiling", len(decoded.ShipDesigns))
	}
	if len(decoded.Ships) != 1 || len(decoded.StrategicFleets) != 1 || len(decoded.StrategicFleets[0].ShipIDs) != 1 || decoded.StrategicFleets[0].ShipIDs[0] != shipID {
		t.Fatalf("decoded military state ships=%+v fleets=%+v", decoded.Ships, decoded.StrategicFleets)
	}
	reencoded, err := MarshalState(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatalf("schema20 military bytes changed across round trip")
	}
}

func TestBuiltShipSnapshotDoesNotFollowLaterDesignRevision(t *testing.T) {
	state := NewSmallFixture(1802)
	empireID := state.Empires[0].ID
	designID := state.NewID()
	original := baselineMilitarySpec()
	state.ShipDesigns = append(state.ShipDesigns, ShipDesign{ID: designID, EmpireID: empireID, Revision: 1, Name: "Alpha", Spec: original})
	shipID := state.NewID()
	state.Ships = append(state.Ships, Ship{ID: shipID, EmpireID: empireID, SourceDesignID: designID, SourceDesignRevision: 1, Name: "Alpha", Spec: original})
	state.ShipDesigns[0].Revision = 2
	state.ShipDesigns[0].Name = "Beta"
	state.ShipDesigns[0].Spec.StrategicPictureID = 1
	fleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, StrategicFleet{ID: fleetID, EmpireID: empireID, Role: StrategicFleetRoleCombat, AtSystemID: state.Galaxy.Systems[0].ID, ShipIDs: []ID{shipID}})
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	if state.Ships[0].Name != "Alpha" || state.Ships[0].Spec.StrategicPictureID != 0 || state.Ships[0].SourceDesignRevision != 1 {
		t.Fatalf("built Ship snapshot followed design edit: ship=%+v design=%+v", state.Ships[0], state.ShipDesigns[0])
	}
}

func TestMilitaryValidationRejectsShipWithoutCombatFleet(t *testing.T) {
	state := NewSmallFixture(1803)
	empireID := state.Empires[0].ID
	designID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, ShipDesign{ID: designID, EmpireID: empireID, Revision: 1, Name: "Frigate", Spec: baselineMilitarySpec()})
	state.Ships = append(state.Ships, Ship{ID: state.NewID(), EmpireID: empireID, SourceDesignID: designID, SourceDesignRevision: 1, Name: "Frigate", Spec: baselineMilitarySpec()})
	if err := state.Validate(); err == nil {
		t.Fatal("unassigned concrete Ship unexpectedly validated")
	}
}
