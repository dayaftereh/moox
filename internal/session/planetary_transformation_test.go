package session

import (
	"path/filepath"
	"sort"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func sessionPlanetByID(state *core.GameState, id core.ID) *core.Planet {
	for i := range state.Galaxy.Systems {
		for j := range state.Galaxy.Systems[i].Planets {
			if state.Galaxy.Systems[i].Planets[j].ID == id {
				return &state.Galaxy.Systems[i].Planets[j]
			}
		}
	}
	return nil
}

func TestGameSessionPlanetaryTransformationChoiceAuthorityAndObserverLifecycle(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	colony := &state.Colonies[0]
	planet := sessionPlanetByID(state, colony.PlanetID)
	planet.ClimateID = "desert"
	definition := rules.PlanetaryTransformations["terraforming"]
	state.Empires[0].KnownTechnologyIDs = append(state.Empires[0].KnownTechnologyIDs, definition.TechnologyID)
	sort.Ints(state.Empires[0].KnownTechnologyIDs)

	s, err := NewGameSession("game-terraforming", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	choices, err := s.ConstructionChoices(1, colony.ID, rules)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, choice := range choices {
		if choice.ProjectID == "terraforming" {
			found = choice.ProjectKind == core.ConstructionProjectPlanetaryTransformation
		}
	}
	if !found {
		t.Fatalf("seat legal-action surface missing Terraforming: %+v", choices)
	}
	if _, err := s.ConstructionChoices(2, colony.ID, rules); err == nil {
		t.Fatal("foreign seat unexpectedly received planetary transformation choices")
	}

	command, err := game.NewQueuePlanetaryTransformationCommand(1, game.QueuePlanetaryTransformationPayload{ColonyID: colony.ID, ProjectID: "terraforming"})
	if err != nil {
		t.Fatal(err)
	}
	for _, batch := range []protocol.CommandBatch{
		{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-terraforming", SeatID: 1, Turn: 1, BaseRevision: 1, Commands: []protocol.Command{command}},
		{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-terraforming", SeatID: 2, Turn: 1, BaseRevision: 1},
	} {
		if err := s.SubmitTurn(batch); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	construction := observer.State.Colonies[0].Construction
	if construction == nil || construction.ProjectKind != core.ConstructionProjectPlanetaryTransformation || construction.ProjectID != "terraforming" || construction.ProgressPP <= 0 {
		t.Fatalf("Observer Terraforming state=%+v", construction)
	}
	observerPlanet := sessionPlanetByID(observer.State, observer.State.Colonies[0].PlanetID)
	if observerPlanet.ClimateID != "desert" {
		t.Fatalf("incomplete Terraforming changed climate early: %q", observerPlanet.ClimateID)
	}
	for _, buildingID := range observer.State.Colonies[0].Buildings {
		if buildingID == "terraforming" {
			t.Fatalf("Terraforming leaked into Observer persistent buildings: %v", observer.State.Colonies[0].Buildings)
		}
	}
	foundQueued := false
	for _, event := range observer.Events {
		if event.Kind == "colony.planetary_transformation_queued" {
			foundQueued = true
			break
		}
	}
	if !foundQueued {
		t.Fatalf("Observer history missing transformation queue event: %+v", observer.Events)
	}
}

func TestGameSessionBarrenMiddleOrbitTransformationReplayIsDeterministic(t *testing.T) {
	run := func() (string, uint64) {
		state, seats := twoSeatFixture(t)
		rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
		if err != nil {
			t.Fatal(err)
		}
		resolver, err := game.NewEconomyResolver(rules)
		if err != nil {
			t.Fatal(err)
		}
		colony := &state.Colonies[0]
		planet := sessionPlanetByID(state, colony.PlanetID)
		planet.ClimateID = "barren"
		planet.Orbit = 3
		cost := rules.PlanetaryTransformations["terraforming"].ProductionCostPP
		colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectPlanetaryTransformation, ProjectID: "terraforming", ProgressPP: cost - 1}
		s, err := NewGameSession("game-barren-replay", state, seats)
		if err != nil {
			t.Fatal(err)
		}
		for _, seatID := range []protocol.SeatID{1, 2} {
			if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-barren-replay", SeatID: seatID, Turn: 1, BaseRevision: 1}); err != nil {
				t.Fatal(err)
			}
		}
		if err := s.ResolveStrategic(resolver); err != nil {
			t.Fatal(err)
		}
		observer, err := s.ObserverView()
		if err != nil {
			t.Fatal(err)
		}
		if observer.State.Colonies[0].Construction != nil {
			t.Fatalf("near-complete Terraforming did not finish: %+v", observer.State.Colonies[0].Construction)
		}
		observerPlanet := sessionPlanetByID(observer.State, observer.State.Colonies[0].PlanetID)
		if observerPlanet.ClimateID != "desert" && observerPlanet.ClimateID != "tundra" {
			t.Fatalf("middle-orbit Barren transformed to %q", observerPlanet.ClimateID)
		}
		foundCompleted := false
		for _, event := range observer.Events {
			if event.Kind == "colony.planetary_transformation_completed" {
				foundCompleted = true
				break
			}
		}
		if !foundCompleted {
			t.Fatalf("Observer history missing transformation completion: %+v", observer.Events)
		}
		return observerPlanet.ClimateID, observer.State.RNGState
	}

	climateA, rngA := run()
	climateB, rngB := run()
	if climateA != climateB || rngA != rngB {
		t.Fatalf("same replay diverged: A=%s/%d B=%s/%d", climateA, rngA, climateB, rngB)
	}
}
