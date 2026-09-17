package server

import (
	"net/http"
	"testing"

	"moox/internal/app"
	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

func TestHTTPCompositionCatalogGate3Contract(t *testing.T) {
	server, _ := newNewGameServer(t)
	defer server.Close()
	var catalog compositionCatalogResponse
	getJSON(t, server.URL+"/api/v1/new-game/compositions", &catalog)
	if catalog.SchemaVersion != app.SchemaVersion || catalog.DefaultOpponentCount != 1 || catalog.OriginalOpponentMin != 1 || catalog.OriginalOpponentMax != 7 {
		t.Fatalf("catalog=%+v", catalog)
	}
	if len(catalog.Counts) != 7 || len(catalog.GalaxyLimits) != 4 || len(catalog.Assignments) != 4 {
		t.Fatalf("catalog breadth counts/limits/assignments=%d/%d/%d", len(catalog.Counts), len(catalog.GalaxyLimits), len(catalog.Assignments))
	}
}

func TestHTTPCreateGameThreePlayerComposition(t *testing.T) {
	server, _ := newNewGameServer(t)
	defer server.Close()
	request := serverNewGameRequest("three-player-http", "0x1663")
	request.Settings.Players = append(request.Settings.Players, game.NewGamePlayerSpec{SeatID: protocol.SeatID(3), EmpireName: "Klackon", RaceID: "klackon"})
	request.Controllers = append(request.Controllers, app.PlayerControllerSpec{SeatID: 3, Controller: session.ControllerBuiltinAI})
	var created newGameResponse
	postJSON(t, server.URL+"/api/v1/games", request, "", http.StatusCreated, &created)
	if len(created.Players) != 3 || created.Players[2].RaceID != "klackon" {
		t.Fatalf("created=%+v", created)
	}
	var seat3 app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/three-player-http/seats/3/snapshot", &seat3)
	if seat3.View.Seat.Seat.Controller != session.ControllerBuiltinAI || seat3.View.Empire.RaceID != "klackon" {
		t.Fatalf("seat3=%+v", seat3.View.Seat)
	}
}

func TestHTTPCreateGameRejectsWrongSlice166Controller(t *testing.T) {
	server, _ := newNewGameServer(t)
	defer server.Close()
	request := serverNewGameRequest("wrong-controller", "1")
	request.Controllers[1].Controller = session.ControllerRemoteHuman
	postJSON(t, server.URL+"/api/v1/games", request, "", http.StatusBadRequest, nil)
}
