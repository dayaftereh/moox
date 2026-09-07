package session

import (
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func sessionVisualGenome() core.ShipVisualGenome {
	return core.ShipVisualGenome{
		Version: core.ShipVisualGenomeVersion,
		HullID:  "titan", StyleID: "organic", MorphologyID: "bulb", Seed: "session:visual:v4",
		Length: 96, Beam: 48, StationCount: 5,
		StationWidths: []float64{9, 30, 42, 24, 0},
		NotchDepths:   []float64{0, 0, 8, 0, 0},
		EngineCount:   3, DetailCount: 4,
		Cutouts:    []core.ShipVisualCutout{{T: .58, Offset: 7, RX: 6, RY: 4, Angle: -10}},
		Primitives: []core.ShipVisualPrimitive{{Kind: "pod", T: .52, Length: 22, Width: 17, Sweep: .15}},
	}
}

func sessionVisualDesignSpec() core.ShipDesignSpec {
	return core.ShipDesignSpec{
		HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "nuclear_drive", FTLSpeed: 2,
		ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells",
		FuelRangeParsecs: 4, HullBaseCostPP: 20, HullSpace: 25, BaseDesignCostPP: 25, ProductionCostPP: 25,
	}
}

func TestMilitaryDesignVisualImmediatePersistsAfterSubmittedSeatAndLiveSnapshot(t *testing.T) {
	state, seats := twoSeatFixture(t)
	designID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{
		ID: designID, EmpireID: state.Empires[0].ID, Revision: 1, Name: "Visual persistent", Spec: sessionVisualDesignSpec(),
	})
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}

	rules := loadSessionColonyBaseRules(t)
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("visual-persistence", state, seats)
	if err != nil {
		t.Fatal(err)
	}

	status := s.Status()
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion, GameID: status.GameID, SeatID: 1,
		Turn: status.Turn, BaseRevision: status.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	playerAfterSubmit, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if !playerAfterSubmit.Seat.Submitted || s.Status().Phase != PhasePlanning {
		t.Fatalf("seat should be submitted while seat 2 is pending: seat=%+v status=%+v", playerAfterSubmit.Seat, s.Status())
	}

	genome := sessionVisualGenome()
	command, err := game.NewSetMilitaryDesignVisualCommand(1, game.SetMilitaryDesignVisualPayload{DesignID: designID, VisualGenome: genome})
	if err != nil {
		t.Fatal(err)
	}
	before := s.Status().Revision
	if err := s.ResolveMilitaryDesignVisualCommand(1, before, command); err != nil {
		t.Fatal(err)
	}
	if s.Status().Revision != before+1 {
		t.Fatalf("revision=%d want=%d", s.Status().Revision, before+1)
	}

	decision, err := s.DecisionView(1, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if len(decision.Strategic.ShipDesigns) != 1 || decision.Strategic.ShipDesigns[0].VisualGenome == nil || decision.Strategic.ShipDesigns[0].VisualRevision != 1 {
		t.Fatalf("decision visual design=%+v", decision.Strategic.ShipDesigns)
	}
	if decision.Strategic.ShipDesigns[0].Revision != 1 || !reflect.DeepEqual(*decision.Strategic.ShipDesigns[0].VisualGenome, genome) {
		t.Fatalf("decision visual/gameplay revision changed: %+v", decision.Strategic.ShipDesigns[0])
	}

	encoded, err := s.MarshalLiveSnapshot(rules)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := UnmarshalLiveSnapshot(encoded, rules, nil)
	if err != nil {
		t.Fatal(err)
	}
	restoredDecision, err := restored.DecisionView(1, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if len(restoredDecision.Strategic.ShipDesigns) != 1 || restoredDecision.Strategic.ShipDesigns[0].VisualGenome == nil || !reflect.DeepEqual(*restoredDecision.Strategic.ShipDesigns[0].VisualGenome, genome) {
		t.Fatalf("restored visual genome=%+v", restoredDecision.Strategic.ShipDesigns)
	}
}
