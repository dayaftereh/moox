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

func TestShipWeaponMountModifierSchemaRoundTripAndClone(t *testing.T) {
	state := NewSmallFixture(0xB702)
	spec := baselineWeaponTestSpec()
	spec.Weapons = []ShipWeaponMount{{
		Slot: 0, WeaponID: "laser_cannon", Count: 2,
		MountType:   ShipWeaponMountHeavy,
		ModifierIDs: []ShipWeaponModifierID{ShipWeaponModifierArmorPiercing, ShipWeaponModifierAutofire},
	}}
	spec.SpaceUsed = 20
	spec.BaseDesignCostPP = 35
	spec.ProductionCostPP = 35
	state.ShipDesigns = append(state.ShipDesigns, ShipDesign{ID: state.NewID(), EmpireID: state.Empires[0].ID, Revision: 1, Name: "Modified", Spec: spec})
	if err := state.Validate(); err != nil {
		t.Fatalf("modified weapon state rejected: %v", err)
	}
	encoded, err := MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.ShipDesigns[0].Spec.Weapons, spec.Weapons) {
		t.Fatalf("modified weapon round-trip=%+v want=%+v", loaded.ShipDesigns[0].Spec.Weapons, spec.Weapons)
	}
	cloned := CloneShipWeaponMounts(spec.Weapons)
	cloned[0].ModifierIDs[0] = ShipWeaponModifierContinuous
	if spec.Weapons[0].ModifierIDs[0] != ShipWeaponModifierArmorPiercing {
		t.Fatal("CloneShipWeaponMounts aliased modifier slice")
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
		{name: "invalid mount", weapons: []ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1, MountType: "super_heavy"}}},
		{name: "unknown modifier", weapons: []ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1, ModifierIDs: []ShipWeaponModifierID{"unknown"}}}},
		{name: "duplicate modifier", weapons: []ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1, ModifierIDs: []ShipWeaponModifierID{ShipWeaponModifierContinuous, ShipWeaponModifierContinuous}}}},
		{name: "noncanonical modifier order", weapons: []ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1, ModifierIDs: []ShipWeaponModifierID{ShipWeaponModifierAutofire, ShipWeaponModifierArmorPiercing}}}},
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
