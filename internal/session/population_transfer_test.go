package session

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func TestPopulationTransferFlowsThroughAuthoritativeSessionAndObserver(t *testing.T) {
	state, seats := twoSeatFixture(t)
	empireID := state.Empires[0].ID
	state.Empires[0].Freighters = 5
	state.Empires[0].KnownTechnologyIDs = []int{120} // Nuclear Drive.

	planet := &state.Galaxy.Systems[1].Planets[0]
	destination := core.Colony{
		ID: state.NewID(), EmpireID: empireID, PlanetID: planet.ID,
		Population: core.NewAssimilatedPopulation(empireID, 0, 1, 0),
	}
	planet.ColonyID = destination.ID
	state.Colonies = append(state.Colonies, destination)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}

	s, err := NewGameSession("game-pop-transfer", state, seats)
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
	command, err := game.NewTransferPopulationCommand(1, game.TransferPopulationPayload{
		SourceColonyID: state.Colonies[0].ID, DestinationColonyID: destination.ID, Job: core.PopulationJobFarmer,
	})
	if err != nil {
		t.Fatal(err)
	}
	seat1 := protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-pop-transfer", SeatID: 1, Turn: 1, BaseRevision: 1, Commands: []protocol.Command{command}}
	seat2 := protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-pop-transfer", SeatID: 2, Turn: 1, BaseRevision: 1}
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
	if len(observer.State.PopulationTransfers) != 0 {
		t.Fatalf("one-turn transfer should have arrived: %+v", observer.State.PopulationTransfers)
	}
	var started, arrived bool
	for _, event := range observer.Events {
		switch event.Kind {
		case "empire.population_transfer_started":
			started = true
			if event.SeatID != 1 || event.CommandSequence != 1 {
				t.Fatalf("start event authority metadata=%+v", event)
			}
			var payload game.PopulationTransferStartedEvent
			if err := json.Unmarshal(event.Data, &payload); err != nil {
				t.Fatal(err)
			}
			if payload.EmpireID != empireID || payload.FreightersReserved != 5 || payload.ETA != 1 {
				t.Fatalf("start payload=%+v", payload)
			}
		case "empire.population_transfer_arrived":
			arrived = true
		}
	}
	if !started || !arrived {
		t.Fatalf("observer missing transfer lifecycle events started=%v arrived=%v events=%+v", started, arrived, observer.Events)
	}
}

func TestObserverPreservesForeignCohortTransferIdentity(t *testing.T) {
	state, seats := twoSeatFixture(t)
	ownerID := state.Empires[0].ID
	foreignID := state.Empires[1].ID
	home := &state.Colonies[0]
	home.Population = core.PopulationState{Cohorts: []core.PopulationCohort{
		{OriginEmpireID: ownerID, LoyaltyEmpireID: ownerID, AssimilationState: core.PopulationAssimilated, Farmers: 2},
		{OriginEmpireID: foreignID, LoyaltyEmpireID: ownerID, AssimilationState: core.PopulationAssimilated, Farmers: 1},
	}}
	home.Population.Normalize()

	system := &state.Galaxy.Systems[0]
	planet := core.Planet{ID: state.NewID(), Name: "Alpha II", Orbit: 2, SizeID: "medium", MineralID: "abundant", GravityID: "normal_g", ClimateID: "terran"}
	destination := core.Colony{ID: state.NewID(), EmpireID: ownerID, PlanetID: planet.ID, Population: core.NewAssimilatedPopulation(ownerID, 0, 2, 0)}
	planet.ColonyID = destination.ID
	system.Planets = append(system.Planets, planet)
	state.Colonies = append(state.Colonies, destination)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}

	s, err := NewGameSession("game-pop-cohort-transfer", state, seats)
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
	key := core.PopulationCohortKey{OriginEmpireID: foreignID, LoyaltyEmpireID: ownerID, AssimilationState: core.PopulationAssimilated}
	command, err := game.NewTransferPopulationCommand(1, game.TransferPopulationPayload{
		SourceColonyID: home.ID, DestinationColonyID: destination.ID, Cohort: &key, Job: core.PopulationJobFarmer,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-pop-cohort-transfer", SeatID: 2, Turn: 1, BaseRevision: 1}); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-pop-cohort-transfer", SeatID: 1, Turn: 1, BaseRevision: 1, Commands: []protocol.Command{command}}); err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := observer.State.Colonies[1].Population.CohortBySemanticKey(key); !ok {
		t.Fatalf("observer lost transferred foreign cohort identity: %+v", observer.State.Colonies[1].Population)
	}
	found := false
	for _, event := range observer.Events {
		if event.Kind != "colony.population_transferred" {
			continue
		}
		var payload game.PopulationTransferredEvent
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			t.Fatal(err)
		}
		if payload.Cohort != key {
			t.Fatalf("observer transfer event cohort=%+v want=%+v", payload.Cohort, key)
		}
		found = true
	}
	if !found {
		t.Fatalf("observer missing cohort-aware transfer event: %+v", observer.Events)
	}
}
