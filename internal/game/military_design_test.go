package game

import (
	"reflect"
	"testing"

	"moox/internal/core"
)

func addBaselineMilitaryTechnologies(empire *core.Empire) {
	addColonyShipTestTechnology(empire, 58, 120, 167, 187)
}

func saveTestFrigateDesign(t *testing.T, resolver *EconomyResolver, state *core.GameState, empireID core.ID, sequence uint32, name string, pictureID int) core.ShipDesign {
	t.Helper()
	command, err := NewSaveMilitaryDesignCommand(sequence, SaveMilitaryDesignPayload{Name: name, HullID: SupportedMilitaryHullID, StrategicPictureID: pictureID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.saveMilitaryDesign(state, empireID, 1, command); err != nil {
		t.Fatal(err)
	}
	return state.ShipDesigns[len(state.ShipDesigns)-1]
}

func TestSaveMilitaryDesignCreatesOriginalBasedClearedFrigate(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1804)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)

	design := saveTestFrigateDesign(t, resolver, state, empire.ID, 1, "Guardian", 0)
	if design.Revision != 1 || design.Name != "Guardian" || design.EmpireID != empire.ID {
		t.Fatalf("design identity=%+v", design)
	}
	want := core.ShipDesignSpec{
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
	if !reflect.DeepEqual(design.Spec, want) {
		t.Fatalf("cleared Frigate spec=%+v want=%+v", design.Spec, want)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("design state invalid: %v", err)
	}
}

func TestMilitaryDesignCatalogIsNotLimitedToOriginalSixSlots(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1805)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)
	for i := 0; i < 9; i++ {
		saveTestFrigateDesign(t, resolver, state, empire.ID, uint32(i+1), "Frigate", i%8)
	}
	if len(state.ShipDesigns) != 9 {
		t.Fatalf("designs=%d want=9", len(state.ShipDesigns))
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("nine-design catalog invalid: %v", err)
	}
}

func TestMilitaryDesignUsesBestKnownMandatorySystems(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1806)
	empire := &state.Empires[0]
	addColonyShipTestTechnology(empire,
		58, 122,
		120, 72,
		187, 191,
		167, 51,
		33,
	)
	design := saveTestFrigateDesign(t, resolver, state, empire.ID, 1, "Advanced", 7)
	if design.Spec.WarpDriveID != "fusion_drive" || design.Spec.FTLSpeed != 3 {
		t.Fatalf("drive=%q speed=%d", design.Spec.WarpDriveID, design.Spec.FTLSpeed)
	}
	if design.Spec.ComputerID != "optronic_computer" || design.Spec.ArmorID != "tritanium_armor" || design.Spec.FuelCellID != "deuterium_fuel_cells" || design.Spec.FuelRangeParsecs != 6 {
		t.Fatalf("mandatory systems=%+v", design.Spec)
	}
	if design.Spec.ShieldID != "class_i_shield" || design.Spec.SpaceUsed != 5 || design.Spec.BaseDesignCostPP != 28 || design.Spec.ProductionCostPP != 28 {
		t.Fatalf("shield/cost/space=%+v", design.Spec)
	}
}

func TestFeudalMilitaryDesignCostUsesExistingShipGovernmentRule(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1807)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)
	modifier := rules.RaceModifiers[empire.RaceID]
	modifier.GovernmentTraitID = "government_feudal"
	rules.RaceModifiers[empire.RaceID] = modifier
	design := saveTestFrigateDesign(t, resolver, state, empire.ID, 1, "Feudal Frigate", 0)
	if design.Spec.BaseDesignCostPP != 25 || design.Spec.ProductionCostPP != 17 {
		t.Fatalf("Feudal design costs base=%d production=%d want 25/17", design.Spec.BaseDesignCostPP, design.Spec.ProductionCostPP)
	}
}

func TestMilitaryConstructionChoiceQueueCompletionAndSnapshot(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1808)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	addBaselineMilitaryTechnologies(empire)
	design := saveTestFrigateDesign(t, resolver, state, empire.ID, 1, "Sentinel", 3)

	choices, err := rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, choice := range choices {
		if choice.ProjectKind != core.ConstructionProjectMilitaryShip || choice.ShipDesignID != design.ID {
			continue
		}
		found = true
		if choice.ProjectID != MilitaryShipProjectID || choice.ShipDesignRevision != 1 || choice.ShipDesignName != "Sentinel" || choice.ProductionCostPP != 25 {
			t.Fatalf("military Construction choice=%+v", choice)
		}
	}
	if !found {
		t.Fatalf("military design %d missing from Construction choices", design.ID)
	}

	queue, err := NewQueueMilitaryShipCommand(2, QueueMilitaryShipPayload{ColonyID: colony.ID, ShipDesignID: design.ID})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.queueMilitaryShip(state, empire.ID, 1, queue)
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != "colony.military_ship_queued" || colony.Construction == nil || colony.Construction.ProjectKind != core.ConstructionProjectMilitaryShip || colony.Construction.ShipDesignID != design.ID || colony.Construction.ShipDesignRevision != 1 {
		t.Fatalf("queue event/state event=%+v construction=%+v", event, colony.Construction)
	}

	update, err := NewSaveMilitaryDesignCommand(3, SaveMilitaryDesignPayload{DesignID: design.ID, Name: "Changed", HullID: SupportedMilitaryHullID, StrategicPictureID: 4})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.saveMilitaryDesign(state, empire.ID, 1, update); err == nil {
		t.Fatal("active military Construction unexpectedly allowed design revision")
	}

	colony.PopulationDynamics.ProductionAvailable = 25
	events, err := resolver.advanceConstruction(state)
	if err != nil {
		t.Fatal(err)
	}
	if colony.Construction != nil {
		t.Fatalf("completed military Construction remains: %+v", colony.Construction)
	}
	if len(state.Ships) != 1 {
		t.Fatalf("ships=%d want=1", len(state.Ships))
	}
	ship := state.Ships[0]
	if ship.SourceDesignID != design.ID || ship.SourceDesignRevision != 1 || ship.Name != "Sentinel" || ship.Spec.StrategicPictureID != 3 || ship.Spec.ProductionCostPP != 25 {
		t.Fatalf("built Ship snapshot=%+v", ship)
	}
	var fleet *core.StrategicFleet
	for i := range state.StrategicFleets {
		if len(state.StrategicFleets[i].ShipIDs) == 1 && state.StrategicFleets[i].ShipIDs[0] == ship.ID {
			fleet = &state.StrategicFleets[i]
			break
		}
	}
	if fleet == nil || fleet.Role != core.StrategicFleetRoleCombat || fleet.SpecialKind != core.StrategicFleetSpecialNone || fleet.AtSystemID != state.Galaxy.Systems[0].ID {
		t.Fatalf("built Ship combat Fleet=%+v fleets=%+v", fleet, state.StrategicFleets)
	}
	if findDomainEvent(events, "colony.military_ship_completed") == nil {
		t.Fatalf("completion events=%+v", events)
	}

	updatedEvent, err := resolver.saveMilitaryDesign(state, empire.ID, 1, update)
	if err != nil {
		t.Fatal(err)
	}
	if updatedEvent.Kind != "empire.military_design_updated" || state.ShipDesigns[0].Revision != 2 || state.ShipDesigns[0].Spec.StrategicPictureID != 4 {
		t.Fatalf("updated design=%+v event=%+v", state.ShipDesigns[0], updatedEvent)
	}
	if state.Ships[0].SourceDesignRevision != 1 || state.Ships[0].Name != "Sentinel" || state.Ships[0].Spec.StrategicPictureID != 3 {
		t.Fatalf("built Ship changed after current design update: %+v", state.Ships[0])
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("post-build/update military state invalid: %v", err)
	}
}

func TestMilitaryDesignRejectsUnsupportedHullAndPicture(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1809)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)

	wrongHull, _ := NewSaveMilitaryDesignCommand(1, SaveMilitaryDesignPayload{Name: "Cruiser", HullID: "cruiser", StrategicPictureID: 16})
	if _, err := resolver.saveMilitaryDesign(state, empire.ID, 1, wrongHull); err == nil {
		t.Fatal("unsupported Cruiser design unexpectedly accepted")
	}
	wrongPicture, _ := NewSaveMilitaryDesignCommand(2, SaveMilitaryDesignPayload{Name: "Bad picture", HullID: "frigate", StrategicPictureID: 8})
	if _, err := resolver.saveMilitaryDesign(state, empire.ID, 1, wrongPicture); err == nil {
		t.Fatal("Frigate design with non-Frigate picture unexpectedly accepted")
	}
}

func TestSaveMilitaryDesignSupportsOneStandardLaser(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0xB702)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)
	addColonyShipTestTechnology(empire, 100)
	command, err := NewSaveMilitaryDesignCommand(1, SaveMilitaryDesignPayload{
		Name: "Laser Frigate", HullID: SupportedMilitaryHullID, StrategicPictureID: 0,
		Weapons: []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.saveMilitaryDesign(state, empire.ID, 1, command); err != nil {
		t.Fatal(err)
	}
	design := state.ShipDesigns[len(state.ShipDesigns)-1]
	if !reflect.DeepEqual(design.Spec.Weapons, []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}}) {
		t.Fatalf("weapons=%+v", design.Spec.Weapons)
	}
	if design.Spec.SpaceUsed != 10 || design.Spec.BaseDesignCostPP != 30 || design.Spec.ProductionCostPP != 30 {
		t.Fatalf("armed costs space=%d base=%d production=%d", design.Spec.SpaceUsed, design.Spec.BaseDesignCostPP, design.Spec.ProductionCostPP)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("armed design state invalid: %v", err)
	}
}

func TestSaveMilitaryDesignLaserRequiresTechnologyWithoutMutation(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0xB703)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)
	beforeNextID := state.NextID
	command, err := NewSaveMilitaryDesignCommand(1, SaveMilitaryDesignPayload{
		Name: "Illegal Laser", HullID: SupportedMilitaryHullID, StrategicPictureID: 0,
		Weapons: []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.saveMilitaryDesign(state, empire.ID, 1, command); err == nil {
		t.Fatal("expected missing Laser technology to reject")
	}
	if len(state.ShipDesigns) != 0 || state.NextID != beforeNextID {
		t.Fatalf("failed armed save mutated state: designs=%d next_id=%d want=%d", len(state.ShipDesigns), state.NextID, beforeNextID)
	}
}

func TestSaveMilitaryDesignRejectsUnsupportedWeaponSurface(t *testing.T) {
	cases := []SaveMilitaryDesignPayload{
		{Name: "wrong slot", HullID: SupportedMilitaryHullID, Weapons: []core.ShipWeaponMount{{Slot: 1, WeaponID: "laser_cannon", Count: 1}}},
		{Name: "wrong id", HullID: SupportedMilitaryHullID, Weapons: []core.ShipWeaponMount{{Slot: 0, WeaponID: "nuclear_missile", Count: 1}}},
		{Name: "wrong count", HullID: SupportedMilitaryHullID, Weapons: []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 2}}},
		{Name: "two mounts", HullID: SupportedMilitaryHullID, Weapons: []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}, {Slot: 1, WeaponID: "laser_cannon", Count: 1}}},
	}
	for _, payload := range cases {
		payload.StrategicPictureID = 0
		if _, err := NewSaveMilitaryDesignCommand(1, payload); err == nil {
			t.Fatalf("expected unsupported weapon payload to reject: %+v", payload.Weapons)
		}
	}
}

func TestBuiltMilitaryShipKeepsWeaponSnapshotAcrossDesignRevision(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0xB704)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)
	addColonyShipTestTechnology(empire, 100)
	command, err := NewSaveMilitaryDesignCommand(1, SaveMilitaryDesignPayload{
		Name: "Armed", HullID: SupportedMilitaryHullID, StrategicPictureID: 0,
		Weapons: []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.saveMilitaryDesign(state, empire.ID, 1, command); err != nil {
		t.Fatal(err)
	}
	designID := state.ShipDesigns[0].ID
	if _, err := completeMilitaryShip(state, &state.Colonies[0], &state.ShipDesigns[0]); err != nil {
		t.Fatal(err)
	}
	if len(state.Ships) != 1 || len(state.Ships[0].Spec.Weapons) != 1 {
		t.Fatalf("built armed Ship=%+v", state.Ships)
	}
	update, err := NewSaveMilitaryDesignCommand(2, SaveMilitaryDesignPayload{DesignID: designID, Name: "Unarmed revision", HullID: SupportedMilitaryHullID, StrategicPictureID: 0})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.saveMilitaryDesign(state, empire.ID, 1, update); err != nil {
		t.Fatal(err)
	}
	if len(state.ShipDesigns[0].Spec.Weapons) != 0 {
		t.Fatalf("updated design remains armed: %+v", state.ShipDesigns[0].Spec.Weapons)
	}
	if !reflect.DeepEqual(state.Ships[0].Spec.Weapons, []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}}) || state.Ships[0].SourceDesignRevision != 1 {
		t.Fatalf("built Ship weapon snapshot changed: %+v", state.Ships[0])
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("revised design/built Ship state invalid: %v", err)
	}
}
