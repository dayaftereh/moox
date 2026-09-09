package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"moox/internal/app"
	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

func newServerFixture(t *testing.T, observerEnabled bool, staticFS fs.FS) (*httptest.Server, *core.GameState) {
	t.Helper()
	state := core.NewSmallFixture(0x80080002)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	gameSession, err := session.NewGameSession("demo", state, []session.Seat{{
		ID: 1, EmpireID: state.Empires[0].ID, Name: "Developer", Controller: session.ControllerLocalHuman,
	}})
	if err != nil {
		t.Fatal(err)
	}
	host := app.NewHost()
	if err := host.Register(app.Registration{Session: gameSession, Resolver: resolver, ImmediateResolver: resolver}); err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(Config{Host: host, ObserverEnabled: observerEnabled, StaticFS: staticFS})
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(handler), state
}

func TestHTTPPlayerSnapshotAndConcretePopulationCommandFlow(t *testing.T) {
	server, state := newServerFixture(t, true, nil)
	defer server.Close()

	var snapshot app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/demo/seats/1/snapshot", &snapshot)
	if snapshot.ChangeSequence != 1 || snapshot.View.Revision != 1 || snapshot.View.Turn != 1 || len(snapshot.View.Colonies) != 1 {
		t.Fatalf("initial snapshot=%+v", snapshot)
	}

	command, err := game.NewAssignPopulationCommand(1, game.AssignPopulationPayload{
		ColonyID: state.Colonies[0].ID, Farmers: 1, Workers: 2, Scientists: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	batch := protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "demo",
		SeatID:        1,
		Turn:          1,
		BaseRevision:  1,
		Commands:      []protocol.Command{command},
	}
	var receipt app.Receipt
	postJSON(t, server.URL+"/api/v1/games/demo/turn-submissions", batch, "", http.StatusOK, &receipt)
	if receipt.ChangeSequence != 2 || receipt.GameRevision != 3 {
		t.Fatalf("receipt=%+v", receipt)
	}

	getJSON(t, server.URL+"/api/v1/games/demo/seats/1/snapshot", &snapshot)
	if snapshot.ChangeSequence != 2 || snapshot.View.Revision != 3 || snapshot.View.Turn != 2 || snapshot.View.Phase != session.PhasePlanning {
		t.Fatalf("post-command snapshot=%+v", snapshot)
	}

	var observer app.ObserverSnapshot
	getJSON(t, server.URL+"/api/v1/games/demo/observer/snapshot", &observer)
	found := false
	for _, event := range observer.View.Events {
		if event.Kind == "colony.population_assigned" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("real population assignment command did not reach authoritative event history")
	}
}

func TestHTTPPlanningPreviewIsPureAndCarriesAtomicDecisionProjection(t *testing.T) {
	server, state := newServerFixture(t, true, nil)
	defer server.Close()

	var before app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/demo/seats/1/snapshot", &before)
	if before.Decision == nil {
		t.Fatal("player snapshot is missing atomic DecisionView projection")
	}
	if before.Decision.GameID != before.View.GameID || before.Decision.Revision != before.View.Revision || before.Decision.Turn != before.View.Turn {
		t.Fatalf("snapshot/decision boundary mismatch: view=%+v decision=%+v", before.View, before.Decision)
	}
	beforeColony, err := json.Marshal(before.View.Colonies[0])
	if err != nil {
		t.Fatal(err)
	}
	command, err := game.NewAssignPopulationCommand(1, game.AssignPopulationPayload{
		ColonyID: state.Colonies[0].ID, Farmers: 1, Workers: 2, Scientists: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	batch := protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "demo",
		SeatID:        1,
		Turn:          before.View.Turn,
		BaseRevision:  before.View.Revision,
		Commands:      []protocol.Command{command},
	}
	var preview app.PlanningPreviewSnapshot
	postJSON(t, server.URL+"/api/v1/games/demo/seats/1/planning-preview", batch, "", http.StatusOK, &preview)
	if preview.ChangeSequence != before.ChangeSequence || preview.Preview.GameID != "demo" || preview.Preview.Turn != before.View.Turn || preview.Preview.BaseRevision != before.View.Revision {
		t.Fatalf("planning preview boundary mismatch: before=%+v preview=%+v", before, preview)
	}
	if len(preview.Preview.Projection.Colonies) != 1 {
		t.Fatalf("planning preview colonies=%d, want 1", len(preview.Preview.Projection.Colonies))
	}

	var after app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/demo/seats/1/snapshot", &after)
	if after.ChangeSequence != before.ChangeSequence || after.View.Revision != before.View.Revision || after.View.Turn != before.View.Turn {
		t.Fatalf("planning preview mutated hosted session: before=%+v after=%+v", before, after)
	}
	afterColony, err := json.Marshal(after.View.Colonies[0])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterColony, beforeColony) {
		t.Fatalf("planning preview mutated authoritative Colony: before=%s after=%s", beforeColony, afterColony)
	}
}
func TestWebSocketInvalidatesSnapshotAndReconnectUsesHTTPTruth(t *testing.T) {
	server, state := newServerFixture(t, false, nil)
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/games/demo/stream"
	conn, _, err := websocket.Dial(context.Background(), wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	command, err := game.NewAssignPopulationCommand(1, game.AssignPopulationPayload{
		ColonyID: state.Colonies[0].ID, Farmers: 1, Workers: 2, Scientists: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	batch := protocol.CommandBatch{SchemaVersion: 1, GameID: "demo", SeatID: 1, Turn: 1, BaseRevision: 1, Commands: []protocol.Command{command}}
	postJSON(t, server.URL+"/api/v1/games/demo/turn-submissions", batch, "", http.StatusOK, &app.Receipt{})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var note app.Notification
	if err := wsjson.Read(ctx, conn, &note); err != nil {
		t.Fatal(err)
	}
	if note.Kind != "snapshot_invalidated" || note.ChangeSequence != 2 || note.GameRevision != 3 || note.Reason != "turn_advanced" {
		t.Fatalf("notification=%+v", note)
	}
	_ = conn.Close(websocket.StatusNormalClosure, "reconnect")

	var snapshot app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/demo/seats/1/snapshot", &snapshot)
	if snapshot.ChangeSequence != note.ChangeSequence || snapshot.View.Revision != note.GameRevision || snapshot.View.Turn != 2 {
		t.Fatalf("HTTP resync does not match notification: note=%+v snapshot=%+v", note, snapshot)
	}
}

func TestHTTPValidationObserverBoundaryAndBattleEndpoint(t *testing.T) {
	server, _ := newServerFixture(t, false, nil)
	defer server.Close()

	response, err := http.Get(server.URL + "/api/v1/games/missing/seats/1/snapshot")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown game status=%d", response.StatusCode)
	}
	response.Body.Close()

	response, err = http.Get(server.URL + "/api/v1/games/demo/observer/snapshot")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("disabled observer status=%d", response.StatusCode)
	}
	response.Body.Close()

	request, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/games/demo/turn-submissions", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing content type status=%d", response.StatusCode)
	}
	response.Body.Close()

	request, err = http.NewRequest(http.MethodPost, server.URL+"/api/v1/games/demo/turn-submissions", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://evil.invalid")
	response, err = http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin mutation status=%d", response.StatusCode)
	}
	response.Body.Close()

	battleCommand, err := protocol.NewCommand(1, "battle.end_activation", map[string]any{"ship_id": 1})
	if err != nil {
		t.Fatal(err)
	}
	var apiErr apiErrorEnvelope
	postJSON(t, server.URL+"/api/v1/games/demo/battles/99/commands", commandRequest{SchemaVersion: 1, SeatID: 1, Command: battleCommand}, "", http.StatusConflict, &apiErr)
	if apiErr.Error.Code != "session_rejected" {
		t.Fatalf("battle endpoint error=%+v", apiErr)
	}
}

func TestStaticSPAUsesAssetAndIndexFallback(t *testing.T) {
	staticFS := fstest.MapFS{
		"index.html":     {Data: []byte("<html>MOOX</html>")},
		"assets/main.js": {Data: []byte("console.log('moox')")},
	}
	server, _ := newServerFixture(t, false, staticFS)
	defer server.Close()

	assertBodyContains(t, server.URL+"/", "MOOX")
	assertBodyContains(t, server.URL+"/assets/main.js", "console.log")
	assertBodyContains(t, server.URL+"/strategic/map", "MOOX")

	response, err := http.Get(server.URL + "/api/v1/unknown")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown API route status=%d", response.StatusCode)
	}
}

func getJSON(t *testing.T, url string, dst any) {
	t.Helper()
	response, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("GET %s status=%d body=%s", url, response.StatusCode, body)
	}
	if err := json.NewDecoder(response.Body).Decode(dst); err != nil {
		t.Fatal(err)
	}
}

func postJSON(t *testing.T, url string, value any, origin string, wantStatus int, dst any) {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	if origin != "" {
		request.Header.Set("Origin", origin)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != wantStatus {
		responseBody, _ := io.ReadAll(response.Body)
		t.Fatalf("POST %s status=%d want=%d body=%s", url, response.StatusCode, wantStatus, responseBody)
	}
	if dst != nil {
		if err := json.NewDecoder(response.Body).Decode(dst); err != nil {
			t.Fatal(err)
		}
	}
}

func assertBodyContains(t *testing.T, url, want string) {
	t.Helper()
	response, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || !strings.Contains(string(body), want) {
		t.Fatalf("GET %s status=%d body=%s want substring=%q", url, response.StatusCode, body, want)
	}
}

func TestBattleCommandEndpointAcceptsParticipantTacticalCommand(t *testing.T) {
	server, attackerShipID, defenderShipID := newTacticalServerFixture(t)
	defer server.Close()

	postJSON(t, server.URL+"/api/v1/games/battle-demo/turn-submissions", protocol.CommandBatch{
		SchemaVersion: 1, GameID: "battle-demo", SeatID: 1, Turn: 1, BaseRevision: 1,
	}, "", http.StatusOK, &app.Receipt{})
	var secondReceipt app.Receipt
	postJSON(t, server.URL+"/api/v1/games/battle-demo/turn-submissions", protocol.CommandBatch{
		SchemaVersion: 1, GameID: "battle-demo", SeatID: 2, Turn: 1, BaseRevision: 1,
	}, "", http.StatusOK, &secondReceipt)
	if secondReceipt.ChangeSequence != 3 || secondReceipt.GameRevision != 2 {
		t.Fatalf("battle creation receipt=%+v", secondReceipt)
	}

	var snapshot app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/battle-demo/seats/1/snapshot", &snapshot)
	if snapshot.View.Phase != session.PhaseEncounters || len(snapshot.Battles) != 1 || snapshot.Battles[0].Tactical == nil {
		t.Fatalf("expected participant tactical battle: %+v", snapshot)
	}
	battleID := snapshot.Battles[0].Spec.ID
	if battleID != 1 {
		t.Fatalf("expected first battle ID 1, got %d", battleID)
	}
	tactical := snapshot.Battles[0].Tactical
	if tactical.State.RNGState != 0 || snapshot.Battles[0].Spec.Tactical.InitialRNGState != 0 {
		t.Fatal("player snapshot leaked Tactical RNG authority")
	}
	if len(tactical.Ships) != 2 || len(tactical.LegalMoves) == 0 || len(tactical.LegalFireActions) != 1 || !tactical.CanEndActivation {
		t.Fatalf("participant Tactical projection incomplete: %+v", tactical)
	}
	moveOption := tactical.LegalMoves[0]
	move, err := battle.NewMoveShipCommand(1, battle.MoveShipPayload{ShipID: attackerShipID, X: moveOption.X, Y: moveOption.Y})
	if err != nil {
		t.Fatal(err)
	}
	var receipt app.Receipt
	postJSON(t, server.URL+"/api/v1/games/battle-demo/battles/1/commands", commandRequest{
		SchemaVersion: 1, SeatID: 1, Command: move,
	}, "", http.StatusOK, &receipt)
	if receipt.ChangeSequence != 4 || receipt.GameRevision != 2 {
		t.Fatalf("battle move receipt=%+v", receipt)
	}
	getJSON(t, server.URL+"/api/v1/games/battle-demo/seats/1/snapshot", &snapshot)
	tactical = snapshot.Battles[0].Tactical
	if tactical.State.NextCommandSequence != 2 || tactical.State.RNGState != 0 {
		t.Fatalf("accepted battle move not visible/redacted correctly: %+v", tactical.State)
	}
	var moved *battle.TacticalShipView
	for i := range tactical.Ships {
		if tactical.Ships[i].ShipID == attackerShipID {
			moved = &tactical.Ships[i]
			break
		}
	}
	if moved == nil || moved.X != moveOption.X || moved.Y != moveOption.Y || moved.Facing != moveOption.ResultingFacing || moved.MovementCurrent != moveOption.MovementRemainingAfter {
		t.Fatalf("player scan projection did not reflect move: moved=%+v option=%+v", moved, moveOption)
	}
	if len(tactical.LegalFireActions) != 1 || len(tactical.LegalFireActions[0].Targets) == 0 {
		t.Fatalf("post-move fire projection missing: %+v", tactical.LegalFireActions)
	}
	fireAction := tactical.LegalFireActions[0]
	fire, err := battle.NewFireBeamCommand(2, battle.FireBeamPayload{ShipID: attackerShipID, TargetShipID: defenderShipID, WeaponSlot: fireAction.WeaponSlot})
	if err != nil {
		t.Fatal(err)
	}
	postJSON(t, server.URL+"/api/v1/games/battle-demo/battles/1/commands", commandRequest{
		SchemaVersion: 1, SeatID: 1, Command: fire,
	}, "", http.StatusOK, &receipt)
	if receipt.ChangeSequence != 5 || receipt.GameRevision != 2 {
		t.Fatalf("battle fire receipt=%+v", receipt)
	}
	getJSON(t, server.URL+"/api/v1/games/battle-demo/seats/1/snapshot", &snapshot)
	if len(snapshot.Battles) != 1 || snapshot.Battles[0].Tactical == nil || snapshot.Battles[0].Tactical.State.NextCommandSequence != 3 {
		t.Fatalf("accepted battle commands not visible in player snapshot: %+v", snapshot.Battles)
	}
}

func newTacticalServerFixture(t *testing.T) (*httptest.Server, core.ID, core.ID) {
	t.Helper()
	state := core.NewSmallFixture(0x80080003)
	attackerEmpireID := state.Empires[0].ID
	defenderEmpireID := state.NewID()
	state.Empires = append(state.Empires, core.Empire{ID: defenderEmpireID, Name: "Defender", RaceID: "alkari"})

	attackerSpec := core.ShipDesignSpec{
		HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "fusion_drive", FTLSpeed: 3,
		ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4,
		HullBaseCostPP: 20, HullSpace: 25, SpaceUsed: 10, BaseDesignCostPP: 30, ProductionCostPP: 30,
		Weapons: []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}},
	}
	defenderSpec := core.ShipDesignSpec{
		HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "nuclear_drive", FTLSpeed: 2,
		ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4,
		HullBaseCostPP: 20, HullSpace: 25, SpaceUsed: 0, BaseDesignCostPP: 25, ProductionCostPP: 25,
	}
	attackerDesignID := state.NewID()
	defenderDesignID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns,
		core.ShipDesign{ID: attackerDesignID, EmpireID: attackerEmpireID, Revision: 1, Name: "Fusion Laser", Spec: attackerSpec},
		core.ShipDesign{ID: defenderDesignID, EmpireID: defenderEmpireID, Revision: 1, Name: "Nuclear Target", Spec: defenderSpec},
	)
	attackerShipID := state.NewID()
	defenderShipID := state.NewID()
	state.Ships = append(state.Ships,
		core.Ship{ID: attackerShipID, EmpireID: attackerEmpireID, SourceDesignID: attackerDesignID, SourceDesignRevision: 1, Name: "Fusion Laser", Spec: attackerSpec},
		core.Ship{ID: defenderShipID, EmpireID: defenderEmpireID, SourceDesignID: defenderDesignID, SourceDesignRevision: 1, Name: "Nuclear Target", Spec: defenderSpec},
	)
	systemID := state.Galaxy.Systems[0].ID
	state.StrategicFleets = append(state.StrategicFleets,
		core.StrategicFleet{ID: state.NewID(), EmpireID: attackerEmpireID, Role: core.StrategicFleetRoleCombat, AtSystemID: systemID, ShipIDs: []core.ID{attackerShipID}},
		core.StrategicFleet{ID: state.NewID(), EmpireID: defenderEmpireID, Role: core.StrategicFleetRoleCombat, AtSystemID: systemID, ShipIDs: []core.ID{defenderShipID}},
	)
	state.DiplomaticRelations = reciprocalWarRelations(attackerEmpireID, defenderEmpireID)
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
	gameSession, err := session.NewGameSession("battle-demo", state, []session.Seat{
		{ID: 1, EmpireID: attackerEmpireID, Name: "Attacker", Controller: session.ControllerLocalHuman},
		{ID: 2, EmpireID: defenderEmpireID, Name: "Defender", Controller: session.ControllerRemoteHuman},
	})
	if err != nil {
		t.Fatal(err)
	}
	host := app.NewHost()
	if err := host.Register(app.Registration{Session: gameSession, Resolver: resolver, ImmediateResolver: resolver}); err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(Config{Host: host})
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(handler), attackerShipID, defenderShipID
}

func TestHTTPAdapterPreservesDirectSessionReplayResult(t *testing.T) {
	server, serverState := newServerFixture(t, true, nil)
	defer server.Close()

	directState := core.NewSmallFixture(0x80080002)
	if directState.Colonies[0].ID != serverState.Colonies[0].ID {
		t.Fatal("deterministic fixture IDs diverged")
	}
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	directSession, err := session.NewGameSession("demo", directState, []session.Seat{{
		ID: 1, EmpireID: directState.Empires[0].ID, Name: "Developer", Controller: session.ControllerLocalHuman,
	}})
	if err != nil {
		t.Fatal(err)
	}
	command, err := game.NewAssignPopulationCommand(1, game.AssignPopulationPayload{
		ColonyID: directState.Colonies[0].ID, Farmers: 1, Workers: 2, Scientists: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	batch := protocol.CommandBatch{SchemaVersion: 1, GameID: "demo", SeatID: 1, Turn: 1, BaseRevision: 1, Commands: []protocol.Command{command}}
	if err := directSession.SubmitTurn(batch); err != nil {
		t.Fatal(err)
	}
	if err := directSession.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	if directSession.Status().Phase == session.PhasePostResolution {
		if err := directSession.CompleteTurn(); err != nil {
			t.Fatal(err)
		}
	}
	directObserver, err := directSession.ObserverView()
	if err != nil {
		t.Fatal(err)
	}

	postJSON(t, server.URL+"/api/v1/games/demo/turn-submissions", batch, "", http.StatusOK, &app.Receipt{})
	var remote app.ObserverSnapshot
	getJSON(t, server.URL+"/api/v1/games/demo/observer/snapshot", &remote)
	directJSON, err := json.Marshal(directObserver)
	if err != nil {
		t.Fatal(err)
	}
	remoteJSON, err := json.Marshal(remote.View)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(directJSON, remoteJSON) {
		t.Fatalf("HTTP adapter changed authoritative replay result\ndirect=%s\nremote=%s", directJSON, remoteJSON)
	}
}
