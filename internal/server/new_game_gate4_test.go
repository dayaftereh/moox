package server

import (
	"fmt"
	"net/http"
	"testing"

	"moox/internal/app"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

func TestHTTPGate4SupportedClassCreatePlaySmoke(t *testing.T) {
	cases := []struct {
		name       string
		difficulty core.DifficultyID
		size       game.GalaxySize
		age        game.GalaxyAge
		tech       game.NewGameTechnologyLevel
		localRace  string
		opponents  int
	}{
		{"easy-small-mineral-prewarp-human-2p", "easy", "small", "mineral_rich", "pre_warp", "human", 1},
		{"normal-medium-normal-average-klackon-3p", "normal", "medium", "normal", "average", "klackon", 2},
		{"hard-large-organic-prewarp-human-3p", "hard", "large", "organic_rich", "pre_warp", "human", 2},
		{"veryhard-huge-mineral-average-klackon-2p", "very_hard", "huge", "mineral_rich", "average", "klackon", 1},
		{"impossible-small-normal-average-human-3p", "impossible", "small", "normal", "average", "human", 2},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server, _ := newNewGameServer(t)
			defer server.Close()

			request := serverNewGameRequest(fmt.Sprintf("gate4-%d", i+1), fmt.Sprintf("0x%04x", 0x9100+i))
			request.Settings.DifficultyID = tc.difficulty
			request.Settings.GalaxySize = tc.size
			request.Settings.GalaxyAge = tc.age
			request.Settings.TechnologyLevel = tc.tech
			request.Settings.StrategicCombat = false

			localName := "Human"
			if tc.localRace == "klackon" {
				localName = "Klackon"
			}
			request.Settings.Players = []game.NewGamePlayerSpec{
				{SeatID: 1, EmpireName: localName, RaceID: tc.localRace},
				{SeatID: 2, EmpireName: "Darlok", RaceID: "darlok"},
			}
			request.Controllers = []app.PlayerControllerSpec{
				{SeatID: 1, Controller: session.ControllerLocalHuman},
				{SeatID: 2, Controller: session.ControllerBuiltinAI},
			}
			if tc.opponents == 2 {
				thirdRace, thirdName := "klackon", "Klackon"
				if tc.localRace == "klackon" {
					thirdRace, thirdName = "human", "Human"
				}
				request.Settings.Players = append(request.Settings.Players, game.NewGamePlayerSpec{SeatID: 3, EmpireName: thirdName, RaceID: thirdRace})
				request.Controllers = append(request.Controllers, app.PlayerControllerSpec{SeatID: 3, Controller: session.ControllerBuiltinAI})
			}

			var created newGameResponse
			postJSON(t, server.URL+"/api/v1/games", request, "", http.StatusCreated, &created)
			if len(created.Players) != tc.opponents+1 {
				t.Fatalf("created player count=%d want %d", len(created.Players), tc.opponents+1)
			}

			var before app.PlayerSnapshot
			getJSON(t, server.URL+"/api/v1/games/"+request.GameID+"/seats/1/snapshot", &before)
			if before.View.Turn != 1 {
				t.Fatalf("initial turn=%d want 1", before.View.Turn)
			}

			batch := protocol.CommandBatch{
				SchemaVersion: protocol.CommandSchemaVersion,
				GameID:        request.GameID,
				SeatID:        1,
				Turn:          before.View.Turn,
				BaseRevision:  before.View.Revision,
			}
			var receipt app.Receipt
			postJSON(t, server.URL+"/api/v1/games/"+request.GameID+"/turn-submissions", batch, "", http.StatusOK, &receipt)

			var after app.PlayerSnapshot
			getJSON(t, server.URL+"/api/v1/games/"+request.GameID+"/seats/1/snapshot", &after)
			if after.View.Turn <= before.View.Turn {
				t.Fatalf("turn did not advance through built-in AI resolution: before=%d after=%d receipt=%+v", before.View.Turn, after.View.Turn, receipt)
			}
		})
	}
}

func TestHTTPGate4UnsupportedCombinationsRejected(t *testing.T) {
	server, _ := newNewGameServer(t)
	defer server.Close()

	fourPlayers := serverNewGameRequest("gate4-four-player", "0x9201")
	fourPlayers.Settings.Players = append(fourPlayers.Settings.Players,
		game.NewGamePlayerSpec{SeatID: 3, EmpireName: "Klackon", RaceID: "klackon"},
		game.NewGamePlayerSpec{SeatID: 4, EmpireName: "Psilon", RaceID: "psilon"},
	)
	fourPlayers.Controllers = append(fourPlayers.Controllers,
		app.PlayerControllerSpec{SeatID: 3, Controller: session.ControllerBuiltinAI},
		app.PlayerControllerSpec{SeatID: 4, Controller: session.ControllerBuiltinAI},
	)
	postJSON(t, server.URL+"/api/v1/games", fourPlayers, "", http.StatusBadRequest, nil)

	advanced := serverNewGameRequest("gate4-advanced", "0x9202")
	advanced.Settings.TechnologyLevel = "advanced"
	postJSON(t, server.URL+"/api/v1/games", advanced, "", http.StatusBadRequest, nil)

	darlokLocal := serverNewGameRequest("gate4-darlok-local", "0x9203")
	darlokLocal.Settings.Players[0].RaceID = "darlok"
	darlokLocal.Settings.Players[0].EmpireName = "Darlok Local"
	postJSON(t, server.URL+"/api/v1/games", darlokLocal, "", http.StatusBadRequest, nil)

	var games []app.GameSummary
	getJSON(t, server.URL+"/api/v1/games", &games)
	if len(games) != 0 {
		t.Fatalf("unsupported combinations created games: %+v", games)
	}
}
