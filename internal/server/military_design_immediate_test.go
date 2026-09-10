package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"moox/internal/app"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/session"
)

func newMilitaryDesignImmediateServer(t *testing.T) *httptest.Server {
	t.Helper()
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := rules.NewGame(0x8009, game.NewGameSettings{
		GalaxySize:      game.GalaxySizeSmall,
		GalaxyAge:       game.GalaxyAgeNormal,
		TechnologyLevel: game.NewGameTechnologyAverage,
		Players: []game.NewGamePlayerSpec{
			{SeatID: 1, EmpireName: "Human", RaceID: "human"},
			{SeatID: 2, EmpireName: "Darlok", RaceID: "darlok"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	gameSession, err := session.NewGameSession("military-design-immediate", generated.State, []session.Seat{
		{ID: 1, EmpireID: generated.Players[0].EmpireID, Name: "Human", Controller: session.ControllerLocalHuman},
		{ID: 2, EmpireID: generated.Players[1].EmpireID, Name: "Darlok", Controller: session.ControllerBuiltinAI},
	})
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	host := app.NewHost()
	if err := host.Register(app.Registration{Session: gameSession, Resolver: resolver, ImmediateResolver: resolver}); err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(Config{Host: host, ObserverEnabled: true})
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(handler)
}

func TestHTTPImmediateMilitaryDesignCreatesNamedLaserDesign(t *testing.T) {
	server := newMilitaryDesignImmediateServer(t)
	defer server.Close()

	var before app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/military-design-immediate/seats/1/snapshot", &before)
	if before.View.Phase != session.PhasePlanning || before.Decision == nil {
		t.Fatalf("initial snapshot phase/decision=%q/%v", before.View.Phase, before.Decision != nil)
	}
	if len(before.Decision.Decisions.ShipDesigner.Variants) != 2 {
		t.Fatalf("designer variants=%+v", before.Decision.Decisions.ShipDesigner.Variants)
	}

	command, err := game.NewSaveMilitaryDesignCommand(1, game.SaveMilitaryDesignPayload{
		Name:               "Falcon Mk I",
		HullID:             "frigate",
		StrategicPictureID: 0,
		Weapons:            []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	url := server.URL + "/api/v1/games/military-design-immediate/immediate-commands"
	var receipt app.Receipt
	postJSON(t, url, commandRequest{SchemaVersion: app.SchemaVersion, SeatID: 1, BaseRevision: before.View.Revision, Command: command}, "", http.StatusOK, &receipt)
	if receipt.GameRevision != before.View.Revision+1 {
		t.Fatalf("design receipt revision=%d want=%d", receipt.GameRevision, before.View.Revision+1)
	}

	var after app.PlayerSnapshot
	getJSON(t, server.URL+"/api/v1/games/military-design-immediate/seats/1/snapshot", &after)
	if after.Decision == nil || len(after.Decision.Strategic.ShipDesigns) != len(before.Decision.Strategic.ShipDesigns)+1 {
		t.Fatalf("design catalog before=%d after=%v", len(before.Decision.Strategic.ShipDesigns), after.Decision)
	}
	var saved *core.ShipDesign
	for i := range after.Decision.Strategic.ShipDesigns {
		if after.Decision.Strategic.ShipDesigns[i].Name == "Falcon Mk I" {
			saved = &after.Decision.Strategic.ShipDesigns[i]
			break
		}
	}
	if saved == nil || saved.Spec.HullID != "frigate" || saved.Spec.ProductionCostPP != 30 || len(saved.Spec.Weapons) != 1 || saved.Spec.Weapons[0].WeaponID != "laser_cannon" {
		t.Fatalf("saved Laser design=%+v", saved)
	}

	postJSON(t, url, commandRequest{SchemaVersion: app.SchemaVersion, SeatID: 1, BaseRevision: before.View.Revision, Command: command}, "", http.StatusConflict, nil)
}
