package core

import (
	"reflect"
	"testing"
)

func baselineWeaponTestSpec() ShipDesignSpec {
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

func TestShipWeaponMountValidationAndSchema23RoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name    string
		weapons []ShipWeaponMount
	}{
		{name: "unarmed"},
		{name: "one laser", weapons: []ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := NewSmallFixture(0xB701)
			spec := baselineWeaponTestSpec()
			spec.Weapons = append([]ShipWeaponMount(nil), tc.weapons...)
			if len(spec.Weapons) != 0 {
				spec.SpaceUsed += 10
				spec.BaseDesignCostPP += 5
				spec.ProductionCostPP += 5
			}
			state.ShipDesigns = append(state.ShipDesigns, ShipDesign{ID: state.NewID(), EmpireID: state.Empires[0].ID, Revision: 1, Name: "Fixture", Spec: spec})
			if err := state.Validate(); err != nil {
				t.Fatalf("valid schema21 weapon state rejected: %v", err)
			}
			encoded, err := MarshalState(state)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := UnmarshalState(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if loaded.SchemaVersion != 23 || !reflect.DeepEqual(loaded.ShipDesigns[0].Spec.Weapons, spec.Weapons) {
				t.Fatalf("weapon round-trip schema=%d weapons=%+v want=%+v", loaded.SchemaVersion, loaded.ShipDesigns[0].Spec.Weapons, spec.Weapons)
			}
		})
	}
}

func TestShipWeaponMountStructuralValidationRejectsMalformed(t *testing.T) {
	cases := []struct {
		name    string
		weapons []ShipWeaponMount
	}{
		{name: "slot negative", weapons: []ShipWeaponMount{{Slot: -1, WeaponID: "laser_cannon", Count: 1}}},
		{name: "slot high", weapons: []ShipWeaponMount{{Slot: 8, WeaponID: "laser_cannon", Count: 1}}},
		{name: "empty id", weapons: []ShipWeaponMount{{Slot: 0, Count: 1}}},
		{name: "zero count", weapons: []ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon"}}},
		{name: "duplicate slot", weapons: []ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}, {Slot: 0, WeaponID: "laser_cannon", Count: 1}}},
		{name: "descending slot", weapons: []ShipWeaponMount{{Slot: 1, WeaponID: "laser_cannon", Count: 1}, {Slot: 0, WeaponID: "laser_cannon", Count: 1}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec := baselineWeaponTestSpec()
			spec.Weapons = tc.weapons
			if err := validateShipDesignSpec(spec, "fixture"); err == nil {
				t.Fatal("expected malformed weapon mount to reject")
			}
		})
	}
}
