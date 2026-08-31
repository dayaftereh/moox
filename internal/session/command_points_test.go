package session

import (
	"path/filepath"
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
)

func newCommandPointSession(t *testing.T, seed uint64, gameID string) (*GameSession, *game.EconomyResolver) {
	t.Helper()
	state := core.NewSmallFixture(seed)
	secondID := state.NewID()
	state.Empires = append(state.Empires, core.Empire{ID: secondID, Name: "Second", RaceID: "human"})
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession(gameID, state, []Seat{
		{ID: 1, EmpireID: state.Empires[0].ID, Name: "First", Controller: ControllerLocalHuman},
		{ID: 2, EmpireID: secondID, Name: "Second", Controller: ControllerLocalHuman},
	})
	if err != nil {
		t.Fatal(err)
	}
	return s, resolver
}

func TestCommandPointObserverIsolationAndDeterministicReplay(t *testing.T) {
	sessionA, resolverA := newCommandPointSession(t, 0xA520, "cp")
	viewA := submitColonyShipSessionTurn(t, sessionA, resolverA, nil)
	for i := range viewA.State.Empires {
		if viewA.State.Empires[i].CommandPoints != (core.EmpireCommandPoints{Capacity: 5, Used: 0}) {
			t.Fatalf("observer empire %d Command Points=%+v want=5/0", viewA.State.Empires[i].ID, viewA.State.Empires[i].CommandPoints)
		}
	}
	viewA.State.Empires[0].CommandPoints.Capacity = 999
	viewA.State.Empires[0].Treasury.ShipCommandMaintenanceBC = 999
	isolated, err := sessionA.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if isolated.State.Empires[0].CommandPoints.Capacity != 5 || isolated.State.Empires[0].Treasury.ShipCommandMaintenanceBC != 0 {
		t.Fatalf("Observer CP/Treasury mutation leaked: CP=%+v Treasury=%+v", isolated.State.Empires[0].CommandPoints, isolated.State.Empires[0].Treasury)
	}

	sessionB, resolverB := newCommandPointSession(t, 0xA520, "cp")
	viewB := submitColonyShipSessionTurn(t, sessionB, resolverB, nil)
	if !reflect.DeepEqual(isolated.State, viewB.State) {
		t.Fatalf("identical Command Point sessions diverged:\nA=%+v\nB=%+v", isolated.State.Empires, viewB.State.Empires)
	}
	if !reflect.DeepEqual(isolated.Events, viewB.Events) {
		t.Fatalf("identical Command Point session events diverged:\nA=%+v\nB=%+v", isolated.Events, viewB.Events)
	}
}
