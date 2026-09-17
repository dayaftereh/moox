package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"moox/internal/app"
	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

func newNewGameServer(t *testing.T) (*httptest.Server, *app.Host) {
	t.Helper()
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	host, err := app.NewHostWithNewGame(rules)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(Config{Host: host, ObserverEnabled: true})
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(handler), host
}

func serverNewGameRequest(gameID, seed string) newGameRequest {
	return newGameRequest{
		SchemaVersion: app.SchemaVersion,
		GameID:        gameID,
		Seed:          seed,
		Settings: game.NewGameSettings{
			GalaxySize:      game.GalaxySizeSmall,
			GalaxyAge:       game.GalaxyAgeNormal,
			TechnologyLevel: game.NewGameTechnologyAverage,
			Players: []game.NewGamePlayerSpec{
				{SeatID: protocol.SeatID(1), EmpireName: "Human", RaceID: "human"},
				{SeatID: protocol.SeatID(2), EmpireName: "Darlok", RaceID: "darlok"},
			},
		},
		Controllers: []app.PlayerControllerSpec{
			{SeatID: 1, Controller: session.ControllerLocalHuman},
			{SeatID: 2, Controller: session.ControllerBuiltinAI},
		},
	}
}

func TestHTTPCreateGameAllocatesServerGameID(t *testing.T) {
	server, _ := newNewGameServer(t)
	defer server.Close()

	firstRequest := serverNewGameRequest("", "0x8009")
	var first newGameResponse
	postJSON(t, server.URL+"/api/v1/games", firstRequest, "", http.StatusCreated, &first)
	if first.Game.GameID != "game-1" {
		t.Fatalf("first server-assigned game ID=%q want game-1", first.Game.GameID)
	}

	secondRequest := serverNewGameRequest("", "0x8010")
	var second newGameResponse
	postJSON(t, server.URL+"/api/v1/games", secondRequest, "", http.StatusCreated, &second)
	if second.Game.GameID != "game-2" {
		t.Fatalf("second server-assigned game ID=%q want game-2", second.Game.GameID)
	}
}

func TestHTTPCreateGameRejectsWhitespaceGameID(t *testing.T) {
	server, _ := newNewGameServer(t)
	defer server.Close()
	postJSON(t, server.URL+"/api/v1/games", serverNewGameRequest(" game-custom ", "0x8009"), "", http.StatusBadRequest, nil)
}

func TestHTTPCreateGameThenPlayerSnapshot(t *testing.T) {
	server, host := newNewGameServer(t)
	defer server.Close()
	if games := host.ListGames(); len(games) != 0 {
		t.Fatalf("production host not empty: %+v", games)
	}

	var created newGameResponse
	postJSON(t, server.URL+"/api/v1/games", serverNewGameRequest("seeded", "0x8009"), "", http.StatusCreated, &created)
	if created.SchemaVersion != app.SchemaVersion || created.Game.GameID != "seeded" || created.Game.ChangeSequence != 1 || len(created.Players) != 2 {
		t.Fatalf("created=%+v", created)
	}
	var snapshot app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/seeded/seats/1/snapshot", &snapshot)
	if snapshot.View.Empire.ID != created.Players[0].EmpireID || snapshot.View.Empire.RaceID != "human" || snapshot.ChangeSequence != 1 {
		t.Fatalf("player snapshot=%+v", snapshot)
	}
	batch := protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "seeded",
		SeatID:        1,
		Turn:          snapshot.View.Turn,
		BaseRevision:  snapshot.View.Revision,
	}
	var receipt app.Receipt
	postJSON(t, server.URL+"/api/v1/games/seeded/turn-submissions", batch, "", http.StatusOK, &receipt)
	if receipt.GameID != "seeded" || receipt.ChangeSequence != 2 {
		t.Fatalf("turn submission receipt=%+v", receipt)
	}
	var games []app.GameSummary
	getJSON(t, server.URL+"/api/v1/games", &games)
	if len(games) != 1 || games[0].GameID != "seeded" {
		t.Fatalf("games=%+v", games)
	}
}

func TestHTTPCreateGameAcceptsBuiltinAIControllerAssignment(t *testing.T) {
	server, _ := newNewGameServer(t)
	defer server.Close()
	request := serverNewGameRequest("human-ai", "0x8009")
	request.Controllers = []app.PlayerControllerSpec{{SeatID: 1, Controller: session.ControllerLocalHuman}, {SeatID: 2, Controller: session.ControllerBuiltinAI}}
	var created newGameResponse
	postJSON(t, server.URL+"/api/v1/games", request, "", http.StatusCreated, &created)
	var human app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/human-ai/seats/1/snapshot", &human)
	if human.View.Seat.Seat.Controller != session.ControllerLocalHuman {
		t.Fatalf("human controller=%q", human.View.Seat.Seat.Controller)
	}
	var computer app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/human-ai/seats/2/snapshot", &computer)
	if computer.View.Seat.Seat.Controller != session.ControllerBuiltinAI {
		t.Fatalf("AI controller=%q", computer.View.Seat.Seat.Controller)
	}
}

func TestHTTPCreateGameDuplicateAndInputErrors(t *testing.T) {
	server, _ := newNewGameServer(t)
	defer server.Close()
	request := serverNewGameRequest("seeded", "32777")
	postJSON(t, server.URL+"/api/v1/games", request, "", http.StatusCreated, &newGameResponse{})

	var duplicate apiErrorEnvelope
	postJSON(t, server.URL+"/api/v1/games", request, "", http.StatusConflict, &duplicate)
	if duplicate.Error.Code != "game_exists" {
		t.Fatalf("duplicate error=%+v", duplicate)
	}

	badSeed := serverNewGameRequest("bad-seed", "0x")
	var bad apiErrorEnvelope
	postJSON(t, server.URL+"/api/v1/games", badSeed, "", http.StatusBadRequest, &bad)
	if bad.Error.Code != "bad_request" {
		t.Fatalf("bad seed error=%+v", bad)
	}

	badSettings := serverNewGameRequest("bad-settings", "1")
	badSettings.Settings.GalaxySize = "tiny"
	postJSON(t, server.URL+"/api/v1/games", badSettings, "", http.StatusBadRequest, &bad)
	if bad.Error.Code != "bad_request" {
		t.Fatalf("bad settings error=%+v", bad)
	}

	badSchema := serverNewGameRequest("bad-schema", "1")
	badSchema.SchemaVersion = 99
	postJSON(t, server.URL+"/api/v1/games", badSchema, "", http.StatusBadRequest, &bad)
	if bad.Error.Code != "bad_request" {
		t.Fatalf("bad schema error=%+v", bad)
	}

	crossOrigin := serverNewGameRequest("cross-origin", "1")
	postJSON(t, server.URL+"/api/v1/games", crossOrigin, "https://example.invalid", http.StatusForbidden, &bad)
	if bad.Error.Code != "forbidden" {
		t.Fatalf("origin error=%+v", bad)
	}
}

func TestParseNewGameSeed(t *testing.T) {
	cases := map[string]uint64{"0": 0, "32777": 32777, "0x8009": 0x8009, "0XFFFFFFFFFFFFFFFF": ^uint64(0)}
	for raw, want := range cases {
		got, err := parseNewGameSeed(raw)
		if err != nil || got != want {
			t.Fatalf("parse %q=%d,%v want=%d", raw, got, err, want)
		}
	}
	for _, raw := range []string{"", " 1", "1 ", "0x", "-1", "18446744073709551616", "xyz"} {
		if _, err := parseNewGameSeed(raw); err == nil {
			t.Fatalf("invalid seed %q accepted", raw)
		}
	}
}
