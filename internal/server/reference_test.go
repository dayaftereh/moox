package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"moox/internal/app"
	"moox/internal/game"
	"moox/internal/session"
)

func newReferenceControlServer(t *testing.T, controlsEnabled bool, trustedReference bool) (*httptest.Server, *app.Host, string) {
	t.Helper()
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	host, err := app.NewHostWithNewGame(rules)
	if err != nil {
		t.Fatal(err)
	}
	generated, err := rules.NewReferenceTriangleGame(game.ReferenceTriangleSeed)
	if err != nil {
		t.Fatal(err)
	}
	player := generated.Players[0]
	const gameID = "reference-http"
	gameSession, err := session.NewGameSession(gameID, generated.State, []session.Seat{{
		ID: player.SeatID, EmpireID: player.EmpireID, Name: player.Name, Controller: session.ControllerLocalHuman,
	}})
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	registration := app.Registration{Session: gameSession, Resolver: resolver, ImmediateResolver: resolver}
	if trustedReference {
		registration.Reference = &app.ReferenceGameInfo{ScenarioID: game.ReferenceTriangleScenarioID, ProfileID: string(game.ReferenceTriangleProfileBaseline), ControlSeatID: 1}
	}
	if err := host.Register(registration); err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(Config{Host: host, ReferenceControlsEnabled: controlsEnabled})
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(handler), host, gameID
}

func postReferenceAdvance(t *testing.T, serverURL, gameID string, request app.ReferenceAdvanceRequest) (*http.Response, []byte) {
	t.Helper()
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.Post(serverURL+"/api/v1/games/"+gameID+"/reference/advance", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var body bytes.Buffer
	if _, err := body.ReadFrom(response.Body); err != nil {
		t.Fatal(err)
	}
	return response, body.Bytes()
}

func TestReferenceAdvanceRouteAbsentWhenDevelopmentCapabilityDisabled(t *testing.T) {
	server, host, gameID := newReferenceControlServer(t, false, true)
	defer server.Close()
	snapshot, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	response, _ := postReferenceAdvance(t, server.URL, gameID, app.ReferenceAdvanceRequest{
		SchemaVersion: app.SchemaVersion, SeatID: 1, BaseRevision: snapshot.View.Revision,
		Mode: app.ReferenceAdvanceTurns, Turns: 1,
	})
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("status=%d want 404", response.StatusCode)
	}
}

func TestReferenceAdvanceRejectsUntrustedGameOnEnabledServer(t *testing.T) {
	server, host, gameID := newReferenceControlServer(t, true, false)
	defer server.Close()
	snapshot, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	response, body := postReferenceAdvance(t, server.URL, gameID, app.ReferenceAdvanceRequest{
		SchemaVersion: app.SchemaVersion, SeatID: 1, BaseRevision: snapshot.View.Revision,
		Mode: app.ReferenceAdvanceTurns, Turns: 1,
	})
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", response.StatusCode, body)
	}
	var envelope apiErrorEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error.Code != "reference_control_forbidden" {
		t.Fatalf("error=%+v", envelope.Error)
	}
}

func TestReferenceAdvanceHTTPUsesTrustedReferenceMetadata(t *testing.T) {
	server, host, gameID := newReferenceControlServer(t, true, true)
	defer server.Close()
	snapshot, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	response, body := postReferenceAdvance(t, server.URL, gameID, app.ReferenceAdvanceRequest{
		SchemaVersion: app.SchemaVersion, SeatID: 1, BaseRevision: snapshot.View.Revision,
		Mode: app.ReferenceAdvanceTurns, Turns: 1,
	})
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.StatusCode, body)
	}
	var result app.ReferenceAdvanceResult
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	if result.TurnsAdvanced != 1 || result.StopReason != "requested_turns_reached" || result.FinalPhase != session.PhasePlanning {
		t.Fatalf("result=%+v", result)
	}
}

func postReferenceGrantBC(t *testing.T, serverURL, gameID string, request app.ReferenceGrantBCRequest) (*http.Response, []byte) {
	t.Helper()
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.Post(serverURL+"/api/v1/games/"+gameID+"/reference/grant-bc", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var body bytes.Buffer
	if _, err := body.ReadFrom(response.Body); err != nil {
		t.Fatal(err)
	}
	return response, body.Bytes()
}

func TestReferenceGrantBCRouteIsolationAndTrustedGrant(t *testing.T) {
	disabled, disabledHost, gameID := newReferenceControlServer(t, false, true)
	snapshot, err := disabledHost.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	response, _ := postReferenceGrantBC(t, disabled.URL, gameID, app.ReferenceGrantBCRequest{
		SchemaVersion: app.SchemaVersion, SeatID: 1, BaseRevision: snapshot.View.Revision, AmountBC: 100,
	})
	disabled.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("disabled grant route status=%d want 404", response.StatusCode)
	}

	enabled, host, gameID := newReferenceControlServer(t, true, true)
	defer enabled.Close()
	before, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	start := before.View.Empire.Treasury.BalanceBC
	response, body := postReferenceGrantBC(t, enabled.URL, gameID, app.ReferenceGrantBCRequest{
		SchemaVersion: app.SchemaVersion, SeatID: 1, BaseRevision: before.View.Revision, AmountBC: 10000,
	})
	if response.StatusCode != http.StatusOK {
		t.Fatalf("grant status=%d body=%s", response.StatusCode, body)
	}
	var result app.ReferenceGrantBCResult
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	if result.AmountBC != 10000 || result.BalanceBC != start+10000 {
		t.Fatalf("grant result=%+v start=%v", result, start)
	}
}
