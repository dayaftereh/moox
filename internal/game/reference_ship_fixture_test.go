package game

import (
	"reflect"
	"testing"

	"moox/internal/core"
)

func TestReferenceMilitaryShipFixtureUsesProductionDesignAndImmutableSnapshot(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(17100)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)
	addColonyShipTestTechnology(empire, 100)
	if len(state.Galaxy.Systems) == 0 {
		t.Fatal("fixture has no star systems")
	}
	result, err := resolver.MaterializeReferenceMilitaryShipFixture(
		state, empire.ID, 1, state.Galaxy.Systems[0].ID,
		SaveMilitaryDesignPayload{
			Name: "Reference Laser", HullID: SupportedMilitaryHullID, StrategicPictureID: 0,
			Weapons: []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 2}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.DesignID == 0 || result.DesignRevision != 1 || result.ShipID == 0 || result.FleetID == 0 {
		t.Fatalf("fixture result=%+v", result)
	}
	_, design := shipDesignByID(state, result.DesignID)
	if design == nil {
		t.Fatal("fixture design missing")
	}
	ship := shipByID(state, result.ShipID)
	if ship == nil {
		t.Fatal("fixture Ship missing")
	}
	if ship.SourceDesignID != design.ID || ship.SourceDesignRevision != design.Revision || !reflect.DeepEqual(ship.Spec, design.Spec) {
		t.Fatalf("Ship snapshot=%+v design=%+v", ship, design)
	}
	if len(design.Spec.Weapons) == 0 || len(ship.Spec.Weapons) == 0 {
		t.Fatalf("expected weapon snapshots design=%+v ship=%+v", design.Spec.Weapons, ship.Spec.Weapons)
	}
	design.Spec.Weapons[0].Count = 7
	if ship.Spec.Weapons[0].Count != 2 {
		t.Fatalf("built Ship weapon snapshot mutated with design: %+v", ship.Spec.Weapons)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("reference fixture invalid after design change: %v", err)
	}
}
