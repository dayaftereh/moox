package session

import (
	"path/filepath"
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func sessionCombatFleetSpec() core.ShipDesignSpec {
	return core.ShipDesignSpec{
		HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "nuclear_drive", FTLSpeed: 2,
		ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4,
		HullBaseCostPP: 20, HullSpace: 25, BaseDesignCostPP: 25, ProductionCostPP: 25,
	}
}

func runCombatFleetSessionScenario(t *testing.T, gameID string, seed uint64) (*GameSession, ObserverView, core.ID, []core.ID) {
	t.Helper()
	state := core.NewSmallFixture(seed)
	empire := &state.Empires[0]
	empire.KnownTechnologyIDs = []int{51, 72, 120, 167}
	source := &state.Galaxy.Systems[0]
	destination := &state.Galaxy.Systems[1]
	destination.X = source.X + 150
	destination.Y = source.Y

	spec := sessionCombatFleetSpec()
	designID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{
		ID: designID, EmpireID: empire.ID, Revision: 1, Name: "Session combat", Spec: spec,
	})
	shipIDs := make([]core.ID, 0, 2)
	for i := 0; i < 2; i++ {
		shipID := state.NewID()
		state.Ships = append(state.Ships, core.Ship{
			ID: shipID, EmpireID: empire.ID, SourceDesignID: designID, SourceDesignRevision: 1, Name: "Session combat", Spec: spec,
		})
		shipIDs = append(shipIDs, shipID)
	}
	fleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID: fleetID, EmpireID: empire.ID, Role: core.StrategicFleetRoleCombat,
		AtSystemID: source.ID, ShipIDs: append([]core.ID(nil), shipIDs...),
	})
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}

	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession(gameID, state, []Seat{{
		ID: 1, EmpireID: empire.ID, Name: "Player", Controller: ControllerLocalHuman,
	}})
	if err != nil {
		t.Fatal(err)
	}
	before, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	move, err := game.NewMoveFleetCommand(1, game.MoveFleetPayload{
		FleetID: fleetID, DestinationSystemID: destination.ID, ShipIDs: []core.ID{shipIDs[1]},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        gameID,
		SeatID:        1,
		Turn:          before.Turn,
		BaseRevision:  before.Revision,
		Commands:      []protocol.Command{move},
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	view, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	return s, view, fleetID, shipIDs
}

func observerEventIndex(view ObserverView, kind string) int {
	for i := range view.Events {
		if view.Events[i].Kind == kind {
			return i
		}
	}
	return -1
}

func observerFleetByID(view ObserverView, id core.ID) *core.StrategicFleet {
	for i := range view.State.StrategicFleets {
		if view.State.StrategicFleets[i].ID == id {
			return &view.State.StrategicFleets[i]
		}
	}
	return nil
}

func TestGameSessionCombatFleetSubsetMoveObserverIsolationAndDeterministicReplay(t *testing.T) {
	sessionA, viewA, sourceFleetID, shipIDs := runCombatFleetSessionScenario(t, "combat-fleet-replay", 1951)
	_, viewB, _, _ := runCombatFleetSessionScenario(t, "combat-fleet-replay", 1951)

	if !reflect.DeepEqual(viewA.State, viewB.State) {
		t.Fatalf("identical combat Fleet sessions diverged in state:\nA=%+v\nB=%+v", viewA.State.StrategicFleets, viewB.State.StrategicFleets)
	}
	if !reflect.DeepEqual(viewA.Events, viewB.Events) {
		t.Fatalf("identical combat Fleet sessions diverged in events:\nA=%+v\nB=%+v", viewA.Events, viewB.Events)
	}
	if core.StateSchemaVersion != 20 || viewA.State.SchemaVersion != 20 {
		t.Fatalf("observer schema=%d constant=%d want=20", viewA.State.SchemaVersion, core.StateSchemaVersion)
	}

	splitIndex := observerEventIndex(viewA, "empire.fleet_split")
	moveIndex := observerEventIndex(viewA, "empire.fleet_movement_started")
	progressIndex := observerEventIndex(viewA, "empire.fleet_movement_progressed")
	if splitIndex < 0 || moveIndex < 0 || progressIndex < 0 || !(splitIndex < moveIndex && moveIndex < progressIndex) {
		t.Fatalf("strategic event order split=%d move=%d progress=%d events=%+v", splitIndex, moveIndex, progressIndex, viewA.Events)
	}

	sourceFleet := observerFleetByID(viewA, sourceFleetID)
	if sourceFleet == nil || sourceFleet.AtSystemID == 0 || !reflect.DeepEqual(sourceFleet.ShipIDs, []core.ID{shipIDs[0]}) {
		t.Fatalf("observer source Fleet=%+v", sourceFleet)
	}
	var movingFleet *core.StrategicFleet
	for i := range viewA.State.StrategicFleets {
		fleet := &viewA.State.StrategicFleets[i]
		if fleet.ID != sourceFleetID && reflect.DeepEqual(fleet.ShipIDs, []core.ID{shipIDs[1]}) {
			movingFleet = fleet
			break
		}
	}
	if movingFleet == nil || movingFleet.AtSystemID != 0 || movingFleet.DestinationSystemID == 0 || movingFleet.RemainingTurns != 1 || movingFleet.FTLSpeed != 0 {
		t.Fatalf("observer moving Fleet=%+v", movingFleet)
	}

	viewA.State.StrategicFleets[0].ShipIDs[0] = 999999
	isolated, err := sessionA.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	isolatedSource := observerFleetByID(isolated, sourceFleetID)
	if isolatedSource == nil || !reflect.DeepEqual(isolatedSource.ShipIDs, []core.ID{shipIDs[0]}) {
		t.Fatalf("Observer mutation leaked into authoritative Fleet state: %+v", isolatedSource)
	}
}
