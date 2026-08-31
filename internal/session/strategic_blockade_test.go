package session

import (
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func TestObserverViewPreservesAndIsolatesStrategicBlockadeState(t *testing.T) {
	state := core.NewSmallFixture(1701)
	blockaderEmpireID := state.Empires[0].ID
	targetEmpireID := state.NewID()
	targetEmpire := core.Empire{ID: targetEmpireID, Name: "Target Empire", RaceID: "human"}
	targetPlanet := &state.Galaxy.Systems[1].Planets[0]
	targetColonyID := state.NewID()
	targetColony := core.Colony{
		ID:         targetColonyID,
		EmpireID:   targetEmpireID,
		PlanetID:   targetPlanet.ID,
		Population: core.NewAssimilatedPopulation(targetEmpireID, 1, 1, 0),
	}
	targetPlanet.ColonyID = targetColonyID
	state.Empires = append(state.Empires, targetEmpire)
	state.Colonies = append(state.Colonies, targetColony)

	spec := core.ShipDesignSpec{HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "nuclear_drive", FTLSpeed: 2, ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4, HullBaseCostPP: 20, HullSpace: 25, BaseDesignCostPP: 25, ProductionCostPP: 25}
	designID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{ID: designID, EmpireID: blockaderEmpireID, Revision: 1, Name: "Blockader", Spec: spec})
	shipID := state.NewID()
	state.Ships = append(state.Ships, core.Ship{ID: shipID, EmpireID: blockaderEmpireID, SourceDesignID: designID, SourceDesignRevision: 1, Name: "Blockader", Spec: spec})
	fleetID := state.NewID()
	state.StrategicFleets = []core.StrategicFleet{{
		ID:         fleetID,
		EmpireID:   blockaderEmpireID,
		Role:       core.StrategicFleetRoleCombat,
		AtSystemID: state.Galaxy.Systems[1].ID,
		ShipIDs:    []core.ID{shipID},
	}}
	state.DiplomaticRelations = []core.DiplomaticRelation{{
		FromEmpireID: blockaderEmpireID,
		ToEmpireID:   targetEmpireID,
		Stance:       core.DiplomaticStanceHostile,
	}}
	state.Galaxy.Systems[1].BlockadedEmpireIDs = []core.ID{targetEmpireID}

	s, err := NewGameSession("strategic-blockade", state, []Seat{{
		ID:         protocol.SeatID(1),
		EmpireID:   blockaderEmpireID,
		Name:       "Blockader",
		Controller: ControllerLocalHuman,
	}})
	if err != nil {
		t.Fatal(err)
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(observer.State.StrategicFleets, state.StrategicFleets) {
		t.Fatalf("Observer strategic fleets changed: want=%+v got=%+v", state.StrategicFleets, observer.State.StrategicFleets)
	}
	if !reflect.DeepEqual(observer.State.DiplomaticRelations, state.DiplomaticRelations) {
		t.Fatalf("Observer diplomatic relations changed: want=%+v got=%+v", state.DiplomaticRelations, observer.State.DiplomaticRelations)
	}
	if !reflect.DeepEqual(observer.State.Galaxy.Systems[1].BlockadedEmpireIDs, []core.ID{targetEmpireID}) {
		t.Fatalf("Observer blockade state=%v want=[%d]", observer.State.Galaxy.Systems[1].BlockadedEmpireIDs, targetEmpireID)
	}

	observer.State.StrategicFleets[0].AtSystemID = 0
	observer.State.DiplomaticRelations[0].Stance = core.DiplomaticStanceNeutral
	observer.State.Galaxy.Systems[1].BlockadedEmpireIDs = nil
	secondObserver, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if secondObserver.State.StrategicFleets[0].AtSystemID != state.Galaxy.Systems[1].ID {
		t.Fatalf("Observer mutation leaked into authoritative Fleet state: %+v", secondObserver.State.StrategicFleets[0])
	}
	if secondObserver.State.DiplomaticRelations[0].Stance != core.DiplomaticStanceHostile {
		t.Fatalf("Observer mutation leaked into authoritative relation state: %+v", secondObserver.State.DiplomaticRelations[0])
	}
	if !reflect.DeepEqual(secondObserver.State.Galaxy.Systems[1].BlockadedEmpireIDs, []core.ID{targetEmpireID}) {
		t.Fatalf("Observer mutation leaked into authoritative blockade state: %v", secondObserver.State.Galaxy.Systems[1].BlockadedEmpireIDs)
	}
}
