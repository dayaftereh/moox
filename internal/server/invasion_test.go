package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"moox/internal/app"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

func newInvasionServerFixture(t *testing.T) (*httptest.Server, *session.GameSession, core.ID, core.ID, []core.ID) {
	t.Helper()
	state := core.NewSmallFixture(0xD1A11)
	attackerID := state.Empires[0].ID
	defenderID := state.NewID()
	targetSystem := &state.Galaxy.Systems[1]
	targetPlanet := &targetSystem.Planets[0]
	colonyID := state.NewID()
	state.Empires = append(state.Empires, core.Empire{ID: defenderID, Name: "Defender", RaceID: "human", Capital: colonyID})
	targetPlanet.ColonyID = colonyID
	state.Colonies = append(state.Colonies, core.Colony{ID: colonyID, EmpireID: defenderID, PlanetID: targetPlanet.ID, Population: core.NewAssimilatedPopulation(defenderID, 2, 1, 1)})
	state.DiplomaticRelations = []core.DiplomaticRelation{
		{FromEmpireID: attackerID, ToEmpireID: defenderID, Stance: core.DiplomaticStanceWar},
		{FromEmpireID: defenderID, ToEmpireID: attackerID, Stance: core.DiplomaticStanceWar},
	}
	sort.Slice(state.DiplomaticRelations, func(i, j int) bool {
		if state.DiplomaticRelations[i].FromEmpireID != state.DiplomaticRelations[j].FromEmpireID {
			return state.DiplomaticRelations[i].FromEmpireID < state.DiplomaticRelations[j].FromEmpireID
		}
		return state.DiplomaticRelations[i].ToEmpireID < state.DiplomaticRelations[j].ToEmpireID
	})
	spec := core.ShipDesignSpec{HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "nuclear_drive", FTLSpeed: 2, ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4, HullBaseCostPP: 20, HullSpace: 25, BaseDesignCostPP: 25, ProductionCostPP: 25, Weapons: []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}}}
	designID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{ID: designID, EmpireID: attackerID, Revision: 1, Name: "Escort", Spec: spec})
	shipID := state.NewID()
	state.Ships = append(state.Ships, core.Ship{ID: shipID, EmpireID: attackerID, SourceDesignID: designID, SourceDesignRevision: 1, Name: "Escort", Spec: spec})
	combatFleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{ID: combatFleetID, EmpireID: attackerID, Role: core.StrategicFleetRoleCombat, AtSystemID: targetSystem.ID, ShipIDs: []core.ID{shipID}})
	transports := make([]core.ID, 0, 2)
	for i := 0; i < 2; i++ {
		id := state.NewID()
		transports = append(transports, id)
		state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{ID: id, EmpireID: attackerID, Role: core.StrategicFleetRoleCivilian, SpecialKind: core.StrategicFleetSpecialTroopTransport, AtSystemID: targetSystem.ID, FTLSpeed: 2})
	}
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
	gameSession, err := session.NewGameSession("invasion", state, []session.Seat{
		{ID: 1, EmpireID: attackerID, Name: "Attacker", Controller: session.ControllerLocalHuman},
		{ID: 2, EmpireID: defenderID, Name: "Defender", Controller: session.ControllerLocalHuman},
	})
	if err != nil {
		t.Fatal(err)
	}
	status := gameSession.Status()
	for _, seatID := range []protocol.SeatID{1, 2} {
		if err := gameSession.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "invasion", SeatID: seatID, Turn: status.Turn, BaseRevision: status.Revision, Commands: []protocol.Command{}}); err != nil {
			t.Fatal(err)
		}
	}
	if err := gameSession.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	if gameSession.Status().Phase != session.PhaseInvasionDecisions {
		t.Fatalf("prepared phase=%q", gameSession.Status().Phase)
	}
	host := app.NewHost()
	if err := host.Register(app.Registration{Session: gameSession, Resolver: resolver, ImmediateResolver: resolver}); err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(Config{Host: host, ObserverEnabled: true})
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(handler), gameSession, colonyID, attackerID, transports
}

func TestWebSocketGameCompletedNotificationAndReconnectProjection(t *testing.T) {
	server, _, colonyID, attackerID, transports := newInvasionServerFixture(t)
	defer server.Close()
	var before app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/invasion/seats/1/snapshot", &before)
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/games/invasion/stream"
	conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	command, err := game.NewInvadeCommand(1, game.InvadePayload{ColonyID: colonyID, TransportFleetIDs: transports})
	if err != nil {
		t.Fatal(err)
	}
	var receipt app.Receipt
	postJSON(t, server.URL+"/api/v1/games/invasion/immediate-commands", commandRequest{SchemaVersion: app.SchemaVersion, SeatID: 1, BaseRevision: before.View.Revision, Command: command}, "", http.StatusOK, &receipt)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var note app.Notification
	if err := wsjson.Read(ctx, conn, &note); err != nil {
		t.Fatal(err)
	}
	if note.Kind != "snapshot_invalidated" || note.Reason != "game_completed" || note.ChangeSequence != receipt.ChangeSequence || note.GameRevision != receipt.GameRevision {
		t.Fatalf("game-completed notification=%+v receipt=%+v", note, receipt)
	}
	_ = conn.Close(websocket.StatusNormalClosure, "reconnect")

	var reconnected app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/invasion/seats/1/snapshot", &reconnected)
	if reconnected.ChangeSequence != note.ChangeSequence || reconnected.View.Revision != note.GameRevision || reconnected.View.Phase != session.PhaseCompleted || reconnected.View.Result == nil || reconnected.View.Result.WinnerEmpireID != attackerID {
		t.Fatalf("completed reconnect projection note=%+v snapshot=%+v", note, reconnected)
	}
}
func TestHTTPImmediateInvasionCaptureAndRevisionContract(t *testing.T) {
	server, _, colonyID, attackerID, transports := newInvasionServerFixture(t)
	defer server.Close()
	var attacker app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/invasion/seats/1/snapshot", &attacker)
	if attacker.View.Phase != session.PhaseInvasionDecisions || attacker.View.Invasion == nil || attacker.View.Invasion.ColonyID != colonyID {
		t.Fatalf("prepared invasion snapshot=%+v", attacker.View)
	}
	command, err := game.NewInvadeCommand(1, game.InvadePayload{ColonyID: colonyID, TransportFleetIDs: transports})
	if err != nil {
		t.Fatal(err)
	}
	url := server.URL + "/api/v1/games/invasion/immediate-commands"
	postJSON(t, url, commandRequest{SchemaVersion: app.SchemaVersion, SeatID: 1, BaseRevision: attacker.View.Revision - 1, Command: command}, "", http.StatusConflict, nil)
	postJSON(t, url, commandRequest{SchemaVersion: app.SchemaVersion, SeatID: 1, Command: command}, "", http.StatusBadRequest, nil)
	var receipt app.Receipt
	postJSON(t, url, commandRequest{SchemaVersion: app.SchemaVersion, SeatID: 1, BaseRevision: attacker.View.Revision, Command: command}, "", http.StatusOK, &receipt)
	if receipt.ChangeSequence != attacker.ChangeSequence+1 {
		t.Fatalf("receipt change_sequence=%d want=%d", receipt.ChangeSequence, attacker.ChangeSequence+1)
	}
	if receipt.GameRevision != attacker.View.Revision+2 {
		t.Fatalf("receipt revision=%d want=%d: invasion commit + elimination/completion boundary", receipt.GameRevision, attacker.View.Revision+2)
	}
	attacker = app.PlayerSnapshot{}
	getJSON(t, server.URL+"/api/v1/games/invasion/seats/1/snapshot", &attacker)
	if attacker.View.Phase != session.PhaseCompleted || attacker.View.Invasion != nil || attacker.View.Result == nil || attacker.View.Result.Kind != session.ResultConquest || attacker.View.Result.WinnerEmpireID != attackerID {
		t.Fatalf("post-capture completed view phase/invasion/result=%q/%+v/%+v", attacker.View.Phase, attacker.View.Invasion, attacker.View.Result)
	}
	found := false
	for _, colony := range attacker.View.Colonies {
		if colony.ID == colonyID {
			found = true
			if colony.EmpireID != attackerID || colony.GroundForces.Infantry != 4 {
				t.Fatalf("captured colony=%+v", colony)
			}
		}
	}
	if !found {
		t.Fatal("captured colony missing from attacker snapshot")
	}
}
