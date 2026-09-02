package session

import (
	"path/filepath"
	"testing"

	"moox/internal/game"
	"moox/internal/protocol"
)

func TestGeneratedNewGameRunsThroughRealSessionTurn(t *testing.T) {
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	settings := game.NewGameSettings{
		GalaxySize:      game.GalaxySizeSmall,
		GalaxyAge:       game.GalaxyAgeNormal,
		TechnologyLevel: game.NewGameTechnologyAverage,
		Players: []game.NewGamePlayerSpec{
			{SeatID: 1, EmpireName: "Human", RaceID: "human"},
			{SeatID: 2, EmpireName: "Darlok", RaceID: "darlok"},
		},
	}
	generated, err := rules.NewGame(0x8009, settings)
	if err != nil {
		t.Fatal(err)
	}
	seats := make([]Seat, len(generated.Players))
	for i, player := range generated.Players {
		seats[i] = Seat{ID: player.SeatID, EmpireID: player.EmpireID, Name: player.Name, Controller: ControllerLocalHuman}
	}
	s, err := NewGameSession("generated-session", generated.State, seats)
	if err != nil {
		t.Fatal(err)
	}
	for _, player := range generated.Players {
		view, err := s.PlayerView(player.SeatID)
		if err != nil {
			t.Fatal(err)
		}
		if view.Empire.ID != player.EmpireID || len(view.Colonies) != 1 {
			t.Fatalf("player view seat %d=%+v", player.SeatID, view)
		}
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if observer.State == nil || len(observer.State.Galaxy.Systems) != 20 || observer.Turn != 1 || observer.Phase != PhasePlanning {
		t.Fatalf("initial observer=%+v", observer)
	}

	for _, player := range generated.Players {
		batch := protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "generated-session", SeatID: player.SeatID, Turn: 1, BaseRevision: 1}
		if err := s.SubmitTurn(batch); err != nil {
			t.Fatalf("submit seat %d: %v", player.SeatID, err)
		}
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatalf("resolve generated turn: %v", err)
	}
	if err := s.CompleteTurn(); err != nil {
		t.Fatalf("complete generated turn: %v", err)
	}
	status := s.Status()
	if status.Turn != 2 || status.Phase != PhasePlanning || status.Revision != 3 {
		t.Fatalf("post-turn status=%+v", status)
	}
}
