package session

import (
	"path/filepath"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func TestEconomyCommandFlowsThroughAuthoritativeSessionAndObserver(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("game-economy", state, seats)
	if err != nil {
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

	colonyID := state.Colonies[0].ID
	command, err := game.NewAssignPopulationCommand(1, game.AssignPopulationPayload{ColonyID: colonyID, Farmers: 1, Workers: 2, Scientists: 1})
	if err != nil {
		t.Fatal(err)
	}
	seat1 := protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-economy",
		SeatID:        1,
		Turn:          1,
		BaseRevision:  1,
		Commands:      []protocol.Command{command},
	}
	seat2 := protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-economy",
		SeatID:        2,
		Turn:          1,
		BaseRevision:  1,
	}
	if err := s.SubmitTurn(seat2); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(seat1); err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}

	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if observer.Phase != PhasePostResolution || observer.Revision != 2 {
		t.Fatalf("unexpected resolved phase/revision: %q/%d", observer.Phase, observer.Revision)
	}
	colony := observer.State.Colonies[0]
	if colony.Population != (core.PopulationState{Units: 4, Farmers: 1, Workers: 2, Scientists: 1}) {
		t.Fatalf("observer population = %+v", colony.Population)
	}
	if colony.Economy != (core.ColonyEconomy{FoodMilli: 2000, ProductionMilli: 6000, ResearchMilli: 3000, TaxBCMilli: 4000}) {
		t.Fatalf("observer base economy = %+v", colony.Economy)
	}
	found := false
	for _, event := range observer.Events {
		if event.Kind == "colony.population_assigned" {
			found = true
			if event.SeatID != 1 || event.CommandSequence != 1 || event.Revision != 2 {
				t.Fatalf("population event metadata = %+v", event)
			}
		}
	}
	if !found {
		t.Fatal("observer missing population assignment domain event")
	}

	player, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(player.Colonies) != 1 || player.Colonies[0].Economy != colony.Economy {
		t.Fatalf("player projection did not receive resolved own-colony economy: %+v", player.Colonies)
	}
}

func TestEconomyCommandCannotCrossSeatEmpireBoundaryThroughSession(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("game-economy-guard", state, seats)
	if err != nil {
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
	command, err := game.NewAssignPopulationCommand(1, game.AssignPopulationPayload{ColonyID: state.Colonies[0].ID, Farmers: 1, Workers: 2, Scientists: 1})
	if err != nil {
		t.Fatal(err)
	}
	seat1 := protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-economy-guard", SeatID: 1, Turn: 1, BaseRevision: 1}
	seat2 := protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-economy-guard", SeatID: 2, Turn: 1, BaseRevision: 1, Commands: []protocol.Command{command}}
	if err := s.SubmitTurn(seat1); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(seat2); err != nil {
		t.Fatal(err)
	}
	before, _ := s.ObserverView()
	if err := s.ResolveStrategic(resolver); err == nil {
		t.Fatal("expected cross-empire population command to fail")
	}
	after, _ := s.ObserverView()
	if after.Revision != before.Revision || after.Phase != before.Phase || after.State.Colonies[0].Population != before.State.Colonies[0].Population {
		t.Fatalf("failed malicious resolve changed authoritative session")
	}
}
