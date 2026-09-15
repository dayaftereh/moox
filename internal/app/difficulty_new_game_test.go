package app

import (
	"testing"

	"moox/internal/session"
)

func TestHostCreateGamePersistsBuiltinAIControllerMarker(t *testing.T) {
	host := loadNewGameHost(t)
	created, err := host.CreateGame(CreateGameRequest{
		GameID:   "difficulty-controller-marker",
		Seed:     0x8009,
		Settings: appNewGameSettings(),
		Controllers: []PlayerControllerSpec{
			{SeatID: 2, Controller: session.ControllerBuiltinAI},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(created.Players) != 2 {
		t.Fatalf("players=%d want=2", len(created.Players))
	}
	human, err := host.PlayerSnapshot(created.Game.GameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	computer, err := host.PlayerSnapshot(created.Game.GameID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if human.View.Empire.BuiltinAIControlled {
		t.Fatal("local-human empire incorrectly marked builtin AI")
	}
	if !computer.View.Empire.BuiltinAIControlled || computer.View.Seat.Seat.Controller != session.ControllerBuiltinAI {
		t.Fatalf("builtin AI marker/controller mismatch: empire=%+v seat=%+v", computer.View.Empire, computer.View.Seat.Seat)
	}
}
