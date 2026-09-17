package app

import (
	"errors"
	"testing"

	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

func TestHostCreateGameThreePlayerFrozenCompositionAndControllers(t *testing.T) {
	host := loadNewGameHost(t)
	settings := appNewGameSettings()
	settings.Players = append(settings.Players, game.NewGamePlayerSpec{SeatID: protocol.SeatID(3), EmpireName: "Klackon", RaceID: "klackon"})
	created, err := host.CreateGame(CreateGameRequest{
		GameID:   "three-player",
		Seed:     0x1663,
		Settings: settings,
		Controllers: []PlayerControllerSpec{
			{SeatID: 1, Controller: session.ControllerLocalHuman},
			{SeatID: 2, Controller: session.ControllerBuiltinAI},
			{SeatID: 3, Controller: session.ControllerBuiltinAI},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(created.Players) != 3 {
		t.Fatalf("players=%d want=3", len(created.Players))
	}
	wantControllers := []session.ControllerType{session.ControllerLocalHuman, session.ControllerBuiltinAI, session.ControllerBuiltinAI}
	wantRaces := []string{"human", "darlok", "klackon"}
	for i, player := range created.Players {
		snapshot, err := host.PlayerSnapshot("three-player", player.SeatID)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.View.Seat.Seat.Controller != wantControllers[i] || snapshot.View.Empire.RaceID != wantRaces[i] {
			t.Fatalf("seat %d controller/race=%s/%s want=%s/%s", player.SeatID, snapshot.View.Seat.Seat.Controller, snapshot.View.Empire.RaceID, wantControllers[i], wantRaces[i])
		}
	}
}

func TestHostCreateGameRejectsExplicitWrongSlice166Controller(t *testing.T) {
	host := loadNewGameHost(t)
	_, err := host.CreateGame(CreateGameRequest{
		GameID:   "wrong-controller",
		Seed:     1,
		Settings: appNewGameSettings(),
		Controllers: []PlayerControllerSpec{
			{SeatID: 1, Controller: session.ControllerLocalHuman},
			{SeatID: 2, Controller: session.ControllerRemoteHuman},
		},
	})
	if !errors.Is(err, game.ErrInvalidNewGameSettings) {
		t.Fatalf("err=%v want ErrInvalidNewGameSettings", err)
	}
}
