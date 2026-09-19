package game

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"moox/internal/core"
)

func stringPointer(value string) *string { return &value }

func applyMilitaryReferenceProfile(t *testing.T, rules *EconomyRules, state *core.GameState, profile ReferenceTriangleTechnologyProfile) {
	t.Helper()
	if err := rules.ApplyReferenceTriangleTechnologyProfile(state, profile); err != nil {
		t.Fatalf("apply %s profile: %v", profile, err)
	}
}

func hullPictureForTest(t *testing.T, rules *EconomyRules, hullID string) int {
	t.Helper()
	hull, ok := rules.ShipHulls[hullID]
	if !ok || len(hull.StrategicPictureIDs) == 0 {
		t.Fatalf("hull %q has no strategic picture", hullID)
	}
	return hull.StrategicPictureIDs[0]
}

func TestSlice171MiniaturizationGoldenVectors(t *testing.T) {
	rules := loadColonyShipRules(t)

	for _, tc := range []struct {
		name       string
		profile    ReferenceTriangleTechnologyProfile
		wantCost   int
		wantSpace  int
		wantCompLv int
		wantShLv   int
	}{
		{name: "mid", profile: ReferenceTriangleProfileMidTech, wantCost: 275, wantSpace: 13, wantCompLv: 3, wantShLv: 2},
		{name: "all", profile: ReferenceTriangleProfileAllTech, wantCost: 264, wantSpace: 5, wantCompLv: 8, wantShLv: 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := core.NewSmallFixture(0x171100 + uint64(len(tc.name)))
			applyMilitaryReferenceProfile(t, rules, state, tc.profile)
			empire := &state.Empires[0]
			shield := "class_i_shield"
			spec, err := rules.compileMilitaryDesignSpec(empire, "cruiser", hullPictureForTest(t, rules, "cruiser"), militaryDesignSelections{
				ComputerID: "electronic_computer",
				ShieldID:   &shield,
			}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if spec.BaseDesignCostPP != tc.wantCost || spec.ProductionCostPP != tc.wantCost || spec.SpaceUsed != tc.wantSpace || spec.HullSpace != 120 {
				t.Fatalf("profile=%s Cruiser spec=%+v want cost=%d space=%d", tc.profile, spec, tc.wantCost, tc.wantSpace)
			}
			compLevel, err := rules.shipComponentMiniaturizationLevel(empire, 58)
			if err != nil {
				t.Fatal(err)
			}
			shieldLevel, err := rules.shipComponentMiniaturizationLevel(empire, 33)
			if err != nil {
				t.Fatal(err)
			}
			if compLevel != tc.wantCompLv || shieldLevel != tc.wantShLv {
				t.Fatalf("profile=%s levels computer=%d shield=%d want=%d/%d", tc.profile, compLevel, shieldLevel, tc.wantCompLv, tc.wantShLv)
			}
		})
	}
}

func TestSlice171CatalogProfilesAndLateHullLocks(t *testing.T) {
	rules := loadColonyShipRules(t)

	mid := core.NewSmallFixture(0x171120)
	applyMilitaryReferenceProfile(t, rules, mid, ReferenceTriangleProfileMidTech)
	midCatalog, err := rules.MilitaryDesignerCatalog(&mid.Empires[0])
	if err != nil {
		t.Fatal(err)
	}
	if !midCatalog.Hulls[4].SaveAvailable || midCatalog.Hulls[5].SaveAvailable || midCatalog.Hulls[5].LockReason != MilitaryDesignReasonTechnologyRequired {
		t.Fatalf("Mid-Tech Titan/Doom locks=%+v / %+v", midCatalog.Hulls[4], midCatalog.Hulls[5])
	}
	if midCatalog.AutomaticSystems.WarpDriveID != "ion_drive" || midCatalog.AutomaticSystems.ArmorID != "tritanium_armor" || midCatalog.AutomaticSystems.FuelCellID != "iridium_fuel_cells" {
		t.Fatalf("Mid-Tech automatic systems=%+v", midCatalog.AutomaticSystems)
	}
	if midCatalog.AutomaticSystems.TacticalSpeedBonus != 4 || midCatalog.AutomaticSystems.ArmorDurabilityMultiple != 2 {
		t.Fatalf("Mid-Tech automatic effects=%+v", midCatalog.AutomaticSystems)
	}
	if len(midCatalog.Computers) != 5 || midCatalog.Computers[0].MiniaturizationLevel != 3 || !midCatalog.Computers[0].Available {
		t.Fatalf("Mid-Tech computers=%+v", midCatalog.Computers)
	}
	if midCatalog.Computers[0].BeamAttackBonus != 25 || midCatalog.Computers[4].BeamAttackBonus != 125 {
		t.Fatalf("Computer Beam Attack effects=%+v", midCatalog.Computers)
	}
	if len(midCatalog.Shields) != 5 || midCatalog.Shields[0].MiniaturizationLevel != 2 || !midCatalog.Shields[0].Available {
		t.Fatalf("Mid-Tech shields=%+v", midCatalog.Shields)
	}
	if midCatalog.Shields[0].ShieldDamageBlockPerHit != 1 || midCatalog.Shields[0].ShieldCapacityPerHullClass != 5 ||
		midCatalog.Shields[4].ShieldDamageBlockPerHit != 10 || midCatalog.Shields[4].ShieldCapacityPerHullClass != 50 {
		t.Fatalf("Shield effects=%+v", midCatalog.Shields)
	}

	all := core.NewSmallFixture(0x171121)
	applyMilitaryReferenceProfile(t, rules, all, ReferenceTriangleProfileAllTech)
	allCatalog, err := rules.MilitaryDesignerCatalog(&all.Empires[0])
	if err != nil {
		t.Fatal(err)
	}
	if !allCatalog.Hulls[4].SaveAvailable || !allCatalog.Hulls[5].SaveAvailable {
		t.Fatalf("All-Tech late hulls=%+v / %+v", allCatalog.Hulls[4], allCatalog.Hulls[5])
	}
	if allCatalog.AutomaticSystems.WarpDriveID != "interphased_drive" || allCatalog.AutomaticSystems.ArmorID != "xentronium_armor" || allCatalog.AutomaticSystems.FuelCellID != "thorium_fuel_cells" {
		t.Fatalf("All-Tech automatic systems=%+v", allCatalog.AutomaticSystems)
	}
	if allCatalog.AutomaticSystems.TacticalSpeedBonus != 10 || allCatalog.AutomaticSystems.ArmorDurabilityMultiple != 10 {
		t.Fatalf("All-Tech automatic effects=%+v", allCatalog.AutomaticSystems)
	}
	doomFound := false
	for _, variant := range allCatalog.Variants {
		if variant.Key != "doom_star:none" {
			continue
		}
		doomFound = true
		if variant.Spec.ComputerID != "moleculartronic_computer" || variant.Spec.ArmorID != "xentronium_armor" || variant.Spec.ShieldID != "class_x_shield" || variant.Spec.WarpDriveID != "interphased_drive" || variant.Spec.FuelCellID != "thorium_fuel_cells" {
			t.Fatalf("All-Tech Doom Star default=%+v", variant.Spec)
		}
	}
	if !doomFound {
		t.Fatal("All-Tech Doom Star preview missing")
	}
}

func TestSlice171ExplicitComputerAndNoShieldSelection(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0x171130)
	applyMilitaryReferenceProfile(t, rules, state, ReferenceTriangleProfileAllTech)
	empire := &state.Empires[0]
	noShield := ""

	command, err := NewSaveMilitaryDesignCommand(1, SaveMilitaryDesignPayload{
		Name:               "Lean Cruiser",
		HullID:             "cruiser",
		StrategicPictureID: hullPictureForTest(t, rules, "cruiser"),
		ComputerID:         "electronic_computer",
		ShieldID:           &noShield,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.saveMilitaryDesign(state, empire.ID, 1, command); err != nil {
		t.Fatal(err)
	}
	design := state.ShipDesigns[len(state.ShipDesigns)-1]
	if design.Spec.ComputerID != "electronic_computer" || design.Spec.ShieldID != "" || design.Spec.SpaceUsed != 0 || design.Spec.BaseDesignCostPP != 262 {
		t.Fatalf("explicit low-computer/no-shield design=%+v", design.Spec)
	}

	bogus, _ := NewSaveMilitaryDesignCommand(2, SaveMilitaryDesignPayload{
		Name: "Bogus", HullID: "frigate", StrategicPictureID: hullPictureForTest(t, rules, "frigate"), ComputerID: "not_a_computer",
	})
	if _, err := resolver.saveMilitaryDesign(state, empire.ID, 1, bogus); err == nil || !strings.Contains(err.Error(), MilitaryDesignReasonInvalidComponent) {
		t.Fatalf("bogus Computer error=%v", err)
	}

	baseline := core.NewSmallFixture(0x171131)
	baselineEmpire := &baseline.Empires[0]
	addBaselineMilitaryTechnologies(baselineEmpire)
	locked, _ := NewSaveMilitaryDesignCommand(1, SaveMilitaryDesignPayload{
		Name: "Locked", HullID: "frigate", StrategicPictureID: hullPictureForTest(t, rules, "frigate"), ComputerID: "moleculartronic_computer",
	})
	if _, err := resolver.saveMilitaryDesign(baseline, baselineEmpire.ID, 1, locked); err == nil || !strings.Contains(err.Error(), MilitaryDesignReasonTechnologyRequired) {
		t.Fatalf("unowned Computer error=%v", err)
	}
}

func TestSlice171OrbitalDockBuildabilityAndQueueSnapshot(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0x171140)
	applyMilitaryReferenceProfile(t, rules, state, ReferenceTriangleProfileMidTech)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	shield := "class_i_shield"

	save, err := NewSaveMilitaryDesignCommand(1, SaveMilitaryDesignPayload{
		Name:               "Reference Cruiser",
		HullID:             "cruiser",
		StrategicPictureID: hullPictureForTest(t, rules, "cruiser"),
		ComputerID:         "electronic_computer",
		ShieldID:           &shield,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.saveMilitaryDesign(state, empire.ID, 1, save); err != nil {
		t.Fatal(err)
	}
	design := state.ShipDesigns[len(state.ShipDesigns)-1]
	if design.Spec.BaseDesignCostPP != 275 || design.Spec.SpaceUsed != 13 {
		t.Fatalf("Mid-Tech saved design=%+v", design.Spec)
	}

	choices, err := rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, choice := range choices {
		if choice.ProjectKind == core.ConstructionProjectMilitaryShip && choice.ShipDesignID == design.ID {
			t.Fatalf("stationless Colony exposes Cruiser choice: %+v", choice)
		}
	}
	queue, _ := NewQueueMilitaryShipCommand(2, QueueMilitaryShipPayload{ColonyID: colony.ID, ShipDesignID: design.ID})
	if _, err := resolver.queueMilitaryShip(state, empire.ID, 1, queue); err == nil || !strings.Contains(err.Error(), MilitaryDesignReasonOrbitalDockRequired) {
		t.Fatalf("stationless Cruiser queue error=%v", err)
	}

	for _, station := range []string{"star_base", "battlestation", "star_fortress"} {
		probe := *colony
		probe.Buildings = []string{station}
		ok, err := rules.colonyCanBuildMilitaryHull(&probe, "cruiser")
		if err != nil || !ok {
			t.Fatalf("station %s buildability=%v err=%v", station, ok, err)
		}
	}
	colony.Buildings = append(colony.Buildings, "star_base")
	choices, err = rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, choice := range choices {
		if choice.ProjectKind == core.ConstructionProjectMilitaryShip && choice.ShipDesignID == design.ID {
			found = true
			if choice.ProductionCostPP != 275 {
				t.Fatalf("Mid-Tech effective construction cost=%v want=275", choice.ProductionCostPP)
			}
		}
	}
	if !found {
		t.Fatal("Star Base Colony does not expose Cruiser construction")
	}
	if _, err := resolver.queueMilitaryShip(state, empire.ID, 1, queue); err != nil {
		t.Fatal(err)
	}
	if colony.Construction == nil || colony.Construction.MilitaryShipSpec == nil {
		t.Fatalf("military queue has no compiled snapshot: %+v", colony.Construction)
	}
	queuedSpec := cloneShipDesignSpec(*colony.Construction.MilitaryShipSpec)
	if queuedSpec.ProductionCostPP != 275 || queuedSpec.SpaceUsed != 13 {
		t.Fatalf("queued Mid-Tech snapshot=%+v", queuedSpec)
	}

	encoded, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	var restored core.GameState
	if err := json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	if err := restored.Validate(); err != nil {
		t.Fatalf("restored queue state invalid: %v", err)
	}
	restoredConstruction := restored.Colonies[0].Construction
	if restoredConstruction == nil || restoredConstruction.MilitaryShipSpec == nil || !reflect.DeepEqual(*restoredConstruction.MilitaryShipSpec, queuedSpec) {
		t.Fatalf("restored queue snapshot=%+v want=%+v", restoredConstruction, queuedSpec)
	}

	applyMilitaryReferenceProfile(t, rules, state, ReferenceTriangleProfileAllTech)
	projected, err := rules.ProjectMilitaryDesign(state.ShipDesigns[len(state.ShipDesigns)-1], empire)
	if err != nil {
		t.Fatal(err)
	}
	if projected.Revision != design.Revision || projected.Spec.BaseDesignCostPP != 264 || projected.Spec.SpaceUsed != 5 {
		t.Fatalf("All-Tech effective saved-design projection=%+v", projected)
	}
	cost, err := resolver.constructionProjectCostPP(state, colony, colony.Construction)
	if err != nil {
		t.Fatal(err)
	}
	if cost != 275 {
		t.Fatalf("queued construction repriced after research: %v want=275", cost)
	}
	colony.PopulationDynamics.ProductionAvailable = 275
	if _, err := resolver.advanceConstruction(state); err != nil {
		t.Fatal(err)
	}
	if len(state.Ships) != 1 {
		t.Fatalf("built ships=%d want=1", len(state.Ships))
	}
	built := state.Ships[0]
	if built.Spec.ProductionCostPP != 275 || built.Spec.SpaceUsed != 13 || built.Spec.ComputerID != "electronic_computer" || built.Spec.ShieldID != "class_i_shield" {
		t.Fatalf("built Ship did not preserve queue snapshot: %+v", built.Spec)
	}
}

func TestSlice171FrigateDestroyerNeedNoOrbitalDock(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(0x171150)
	colony := &state.Colonies[0]
	for _, hullID := range []string{"frigate", "destroyer"} {
		ok, err := rules.colonyCanBuildMilitaryHull(colony, hullID)
		if err != nil || !ok {
			t.Fatalf("%s stationless buildability=%v err=%v", hullID, ok, err)
		}
	}
	for _, hullID := range []string{"cruiser", "battleship", "titan", "doom_star"} {
		ok, err := rules.colonyCanBuildMilitaryHull(colony, hullID)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatalf("%s unexpectedly buildable without orbital dock", hullID)
		}
	}
}
