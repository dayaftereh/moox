package server

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"moox/internal/app"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/session"
)

func newDiplomacyServerFixture(t *testing.T, staticFS fs.FS) (*httptest.Server, *core.GameState) {
	t.Helper()
	state := core.NewSmallFixture(0xD1A10)
	secondEmpire := core.Empire{ID: state.NewID(), Name: "Darlok", RaceID: "darlok"}
	state.Empires = append(state.Empires, secondEmpire)
	state.MarkEmpiresKnown(state.Empires[0].ID, secondEmpire.ID)

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
	gameSession, err := session.NewGameSession("diplomacy", state, []session.Seat{
		{ID: 1, EmpireID: state.Empires[0].ID, Name: "Human", Controller: session.ControllerLocalHuman},
		{ID: 2, EmpireID: secondEmpire.ID, Name: "Darlok", Controller: session.ControllerLocalHuman},
	})
	if err != nil {
		t.Fatal(err)
	}
	host := app.NewHost()
	if err := host.Register(app.Registration{Session: gameSession, Resolver: resolver, ImmediateResolver: resolver}); err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(Config{Host: host, ObserverEnabled: true, StaticFS: staticFS})
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(handler), state
}

func TestHTTPImmediateDiplomacyLifecycleAndRevisionContract(t *testing.T) {
	server, state := newDiplomacyServerFixture(t, nil)
	defer server.Close()
	url := server.URL + "/api/v1/games/diplomacy/immediate-commands"

	var human app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/diplomacy/seats/1/snapshot", &human)
	declare, _ := game.NewDeclareWarCommand(1, state.Empires[1].ID)
	var receipt app.Receipt
	postJSON(t, url, commandRequest{SchemaVersion: app.SchemaVersion, SeatID: 1, BaseRevision: human.View.Revision, Command: declare}, "", http.StatusOK, &receipt)
	if receipt.GameRevision != human.View.Revision+1 {
		t.Fatalf("war receipt revision=%d want=%d", receipt.GameRevision, human.View.Revision+1)
	}
	getJSON(t, server.URL+"/api/v1/games/diplomacy/seats/1/snapshot", &human)
	if len(human.View.Diplomacy) != 1 || human.View.Diplomacy[0].Stance != core.DiplomaticStanceWar {
		t.Fatalf("war snapshot diplomacy=%+v", human.View.Diplomacy)
	}

	offer, _ := game.NewOfferPeaceCommand(1, state.Empires[1].ID)
	postJSON(t, url, commandRequest{SchemaVersion: app.SchemaVersion, SeatID: 1, BaseRevision: human.View.Revision - 1, Command: offer}, "", http.StatusConflict, nil)
	postJSON(t, url, commandRequest{SchemaVersion: app.SchemaVersion, SeatID: 1, Command: offer}, "", http.StatusBadRequest, nil)
	postJSON(t, url, commandRequest{SchemaVersion: app.SchemaVersion, SeatID: 1, BaseRevision: human.View.Revision, Command: offer}, "", http.StatusOK, &receipt)

	var darlok app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/diplomacy/seats/2/snapshot", &darlok)
	if len(darlok.View.Diplomacy) != 1 || !darlok.View.Diplomacy[0].IncomingPeaceOffer {
		t.Fatalf("incoming peace snapshot=%+v", darlok.View.Diplomacy)
	}
	accept, _ := game.NewAcceptPeaceCommand(1, state.Empires[0].ID)
	postJSON(t, url, commandRequest{SchemaVersion: app.SchemaVersion, SeatID: 2, BaseRevision: darlok.View.Revision, Command: accept}, "", http.StatusOK, &receipt)
	human = app.PlayerSnapshot{}
	darlok = app.PlayerSnapshot{}
	getJSON(t, server.URL+"/api/v1/games/diplomacy/seats/1/snapshot", &human)
	getJSON(t, server.URL+"/api/v1/games/diplomacy/seats/2/snapshot", &darlok)
	if human.View.Diplomacy[0].Stance != core.DiplomaticStancePeace || darlok.View.Diplomacy[0].Stance != core.DiplomaticStancePeace {
		t.Fatalf("peace snapshots human=%+v darlok=%+v", human.View.Diplomacy, darlok.View.Diplomacy)
	}
	if human.View.Diplomacy[0].OutgoingPeaceOffer || darlok.View.Diplomacy[0].IncomingPeaceOffer {
		t.Fatal("peace acceptance did not clear projected offer")
	}
}
