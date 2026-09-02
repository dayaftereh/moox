package app

import (
	"errors"
	"path/filepath"
	"testing"

	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

func loadNewGameHost(t *testing.T) *Host {
	t.Helper()
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	host, err := NewHostWithNewGame(rules)
	if err != nil {
		t.Fatal(err)
	}
	return host
}

func appNewGameSettings() game.NewGameSettings {
	return game.NewGameSettings{
		GalaxySize:      game.GalaxySizeSmall,
		GalaxyAge:       game.GalaxyAgeNormal,
		TechnologyLevel: game.NewGameTechnologyAverage,
		Players: []game.NewGamePlayerSpec{
			{SeatID: protocol.SeatID(1), EmpireName: "Human", RaceID: "human"},
			{SeatID: protocol.SeatID(2), EmpireName: "Darlok", RaceID: "darlok"},
		},
	}
}

func TestHostCreateGameRegistersGeneratedSessionAtomically(t *testing.T) {
	host := loadNewGameHost(t)
	created, err := host.CreateGame(CreateGameRequest{GameID: "new-game", Seed: 0x8009, Settings: appNewGameSettings()})
	if err != nil {
		t.Fatal(err)
	}
	if created.Game.GameID != "new-game" || created.Game.SchemaVersion != SchemaVersion || created.Game.ChangeSequence != 1 || created.Game.Turn != 1 || created.Game.Phase != session.PhasePlanning {
		t.Fatalf("created summary=%+v", created.Game)
	}
	if len(created.Players) != 2 || created.Players[0].SeatID != 1 || created.Players[1].SeatID != 2 {
		t.Fatalf("created players=%+v", created.Players)
	}
	for _, player := range created.Players {
		snapshot, err := host.PlayerSnapshot("new-game", player.SeatID)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.ChangeSequence != 1 || snapshot.View.Empire.ID != player.EmpireID || snapshot.View.Seat.Seat.Controller != session.ControllerLocalHuman {
			t.Fatalf("seat %d snapshot=%+v", player.SeatID, snapshot)
		}
	}
	observer, err := host.ObserverSnapshot("new-game")
	if err != nil {
		t.Fatal(err)
	}
	if observer.View.State == nil || len(observer.View.State.Galaxy.Systems) != 20 || len(observer.View.Seats) != 2 {
		t.Fatalf("observer=%+v", observer.View)
	}

	before := host.ListGames()
	_, err = host.CreateGame(CreateGameRequest{GameID: "new-game", Seed: 0x800A, Settings: appNewGameSettings()})
	if !errors.Is(err, ErrGameExists) {
		t.Fatalf("duplicate error=%v want ErrGameExists", err)
	}
	after := host.ListGames()
	if len(before) != 1 || len(after) != 1 || before[0] != after[0] {
		t.Fatalf("duplicate changed registry before=%+v after=%+v", before, after)
	}
}

func TestHostCreateGameRejectsInvalidSettingsWithoutRegistration(t *testing.T) {
	host := loadNewGameHost(t)
	settings := appNewGameSettings()
	settings.GalaxySize = "medium"
	_, err := host.CreateGame(CreateGameRequest{GameID: "bad-game", Seed: 1, Settings: settings})
	if !errors.Is(err, game.ErrInvalidNewGameSettings) {
		t.Fatalf("error=%v want ErrInvalidNewGameSettings", err)
	}
	if games := host.ListGames(); len(games) != 0 {
		t.Fatalf("invalid game partially registered: %+v", games)
	}
}
