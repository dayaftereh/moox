package game

import (
	"reflect"
	"testing"

	"moox/internal/core"
)

func testPersistentVisualGenome() core.ShipVisualGenome {
	return core.ShipVisualGenome{
		Version: core.ShipVisualGenomeVersion,
		HullID:  "doom_star", StyleID: "spear", MorphologyID: "chevron", Seed: "game:test:visual:v4",
		Length: 105, Beam: 44, StationCount: 5,
		StationWidths: []float64{8, 24, 35, 18, 0},
		NotchDepths:   []float64{0, 0, 10, 0, 0},
		EngineCount:   4, DetailCount: 5,
		Cutouts:    []core.ShipVisualCutout{{T: .5, Offset: 8, RX: 7, RY: 3, Angle: 15}},
		Primitives: []core.ShipVisualPrimitive{{Kind: "wedge", T: .5, Length: 30, Width: 18, Sweep: -1.1}},
	}
}

func TestMilitaryDesignVisualUpdatesOnlyVisualRevisionAndDeepCopies(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1811)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)
	design := saveTestFrigateDesign(t, resolver, state, empire.ID, 1, "Visual test", 0)
	gameplayRevision := design.Revision

	genome := testPersistentVisualGenome()
	command, err := NewSetMilitaryDesignVisualCommand(1, SetMilitaryDesignVisualPayload{DesignID: design.ID, VisualGenome: genome})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveMilitaryDesignVisualCommand(state, empire.ID, 1, command); err != nil {
		t.Fatal(err)
	}
	stored := &state.ShipDesigns[0]
	if stored.Revision != gameplayRevision || stored.VisualRevision != 1 || stored.VisualGenome == nil {
		t.Fatalf("revisions/gameplay changed unexpectedly: %+v", *stored)
	}
	if !reflect.DeepEqual(*stored.VisualGenome, genome) {
		t.Fatalf("stored visual differs: got=%+v want=%+v", *stored.VisualGenome, genome)
	}
	genome.StationWidths[0] = 999
	genome.Primitives[0].Width = 999
	if stored.VisualGenome.StationWidths[0] == 999 || stored.VisualGenome.Primitives[0].Width == 999 {
		t.Fatal("stored visual aliases caller-owned genome slices")
	}

	second := testPersistentVisualGenome()
	second.Seed = "game:test:visual:v4:second"
	second.MorphologyID = "bulb"
	command, err = NewSetMilitaryDesignVisualCommand(1, SetMilitaryDesignVisualPayload{DesignID: design.ID, VisualGenome: second})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveMilitaryDesignVisualCommand(state, empire.ID, 1, command); err != nil {
		t.Fatal(err)
	}
	if state.ShipDesigns[0].Revision != gameplayRevision || state.ShipDesigns[0].VisualRevision != 2 || state.ShipDesigns[0].VisualGenome.Seed != second.Seed {
		t.Fatalf("second visual revision=%+v", state.ShipDesigns[0])
	}
	gameplayCommand, err := NewSaveMilitaryDesignCommand(1, SaveMilitaryDesignPayload{
		DesignID: design.ID, Name: "Visual test revised", HullID: SupportedMilitaryHullID, StrategicPictureID: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.saveMilitaryDesign(state, empire.ID, 1, gameplayCommand); err != nil {
		t.Fatal(err)
	}
	if state.ShipDesigns[0].Revision != gameplayRevision+1 || state.ShipDesigns[0].VisualRevision != 2 || state.ShipDesigns[0].VisualGenome == nil || state.ShipDesigns[0].VisualGenome.Seed != second.Seed {
		t.Fatalf("gameplay design revision lost visual identity: %+v", state.ShipDesigns[0])
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("visual design state invalid: %v", err)
	}
}

func TestCompletedMilitaryShipFreezesDesignVisualGenome(t *testing.T) {
	state := core.NewSmallFixture(1812)
	genome := testPersistentVisualGenome()
	design := core.ShipDesign{
		ID: state.NewID(), EmpireID: state.Empires[0].ID, Revision: 1, VisualRevision: 3,
		Name: "Frozen visual", Spec: core.ShipDesignSpec{
			HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "nuclear_drive", FTLSpeed: 2,
			ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells",
			FuelRangeParsecs: 4, HullBaseCostPP: 20, HullSpace: 25, BaseDesignCostPP: 25, ProductionCostPP: 25,
		}, VisualGenome: &genome,
	}
	state.ShipDesigns = append(state.ShipDesigns, design)
	colony := &state.Colonies[0]
	if _, err := completeMilitaryShip(state, colony, &state.ShipDesigns[0]); err != nil {
		t.Fatal(err)
	}
	ship := state.Ships[len(state.Ships)-1]
	if ship.SourceVisualRevision != 3 || ship.VisualGenome == nil || !reflect.DeepEqual(*ship.VisualGenome, genome) {
		t.Fatalf("built ship visual=%+v", ship)
	}
	state.ShipDesigns[0].VisualGenome.Seed = "changed-after-build"
	if ship.VisualGenome.Seed != "game:test:visual:v4" {
		t.Fatalf("built ship visual was not frozen: %q", ship.VisualGenome.Seed)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("built visual ship state invalid: %v", err)
	}
}
