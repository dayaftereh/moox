package core

import (
	"reflect"
	"testing"
)

func testShipVisualGenome() ShipVisualGenome {
	return ShipVisualGenome{
		Version: ShipVisualGenomeVersion,
		HullID:  "cruiser", StyleID: "sleek", MorphologyID: "manta", Seed: "test:visual:v4",
		Length: 82.5, Beam: 24.25, StationCount: 5,
		StationWidths: []float64{5, 15, 20, 12, 0},
		NotchDepths:   []float64{0, 0, 7, 0, 0},
		EngineCount:   2, DetailCount: 3,
		Cutouts: []ShipVisualCutout{{T: .55, Offset: 4.5, RX: 5, RY: 2.5, Angle: 12}},
		Primitives: []ShipVisualPrimitive{
			{Kind: "wedge", T: .42, Length: 18, Width: 9, Sweep: -.6},
			{Kind: "pod", T: .66, Length: 10, Width: 6, Sweep: .2},
		},
	}
}

func TestShipVisualGenomeV4ValidatesSymmetricHalfSchema(t *testing.T) {
	genome := testShipVisualGenome()
	if err := ValidateShipVisualGenome(genome); err != nil {
		t.Fatalf("valid v4 genome rejected: %v", err)
	}
	genome.MorphologyID = "asymmetric"
	if err := ValidateShipVisualGenome(genome); err == nil {
		t.Fatal("asymmetric morphology must not be accepted by the v4 default persistence contract")
	}
}

func TestShipVisualGenomeRoundTripsThroughStateSchema23(t *testing.T) {
	state := NewSmallFixture(1810)
	genome := testShipVisualGenome()
	state.ShipDesigns = append(state.ShipDesigns, ShipDesign{
		ID: state.NewID(), EmpireID: state.Empires[0].ID, Revision: 1, VisualRevision: 1,
		Name: "Persistent visual", Spec: baselineMilitarySpec(), VisualGenome: &genome,
	})
	if StateSchemaVersion != 23 || state.SchemaVersion != 23 {
		t.Fatalf("optional visual persistence must not require a state-schema bump: state=%d const=%d", state.SchemaVersion, StateSchemaVersion)
	}
	encoded, err := MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.ShipDesigns) != 1 || decoded.ShipDesigns[0].VisualRevision != 1 || decoded.ShipDesigns[0].VisualGenome == nil {
		t.Fatalf("decoded visual design=%+v", decoded.ShipDesigns)
	}
	if !reflect.DeepEqual(*decoded.ShipDesigns[0].VisualGenome, genome) {
		t.Fatalf("visual genome changed across state round-trip\ngot=%+v\nwant=%+v", *decoded.ShipDesigns[0].VisualGenome, genome)
	}
}
