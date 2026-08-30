package session

import (
	"path/filepath"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func submitColonyShipSessionTurn(t *testing.T, s *GameSession, resolver *game.EconomyResolver, commands []protocol.Command) ObserverView {
	t.Helper()
	before, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if before.Phase != PhasePlanning {
		t.Fatalf("turn submission started in phase %q", before.Phase)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        before.GameID,
		SeatID:        1,
		Turn:          before.Turn,
		BaseRevision:  before.Revision,
		Commands:      commands,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        before.GameID,
		SeatID:        2,
		Turn:          before.Turn,
		BaseRevision:  before.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	after, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if after.Phase != PhasePostResolution {
		t.Fatalf("resolved turn phase=%q want=%q", after.Phase, PhasePostResolution)
	}
	return after
}

func completeColonyShipSessionTurn(t *testing.T, s *GameSession) {
	t.Helper()
	if err := s.CompleteTurn(); err != nil {
		t.Fatal(err)
	}
}

func observerHasEvent(view ObserverView, kind string) bool {
	for _, event := range view.Events {
		if event.Kind == kind {
			return true
		}
	}
	return false
}

func TestGameSessionColonyShipBuildMoveArriveColonizeObserverFlow(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	rules.MineralIndustryPerWorker["abundant"] = 600
	state.Empires[0].KnownTechnologyIDs = []int{41, 120, 167}
	home := &state.Galaxy.Systems[0]
	target := &state.Galaxy.Systems[1]
	target.X = home.X + 90
	target.Y = home.Y
	targetPlanetID := target.Planets[0].ID
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("game-colony-ship", state, seats)
	if err != nil {
		t.Fatal(err)
	}

	choices, err := s.ConstructionChoices(1, state.Colonies[0].ID, rules)
	if err != nil {
		t.Fatal(err)
	}
	foundChoice := false
	for _, choice := range choices {
		if choice.ProjectKind != core.ConstructionProjectColonyShip {
			continue
		}
		foundChoice = true
		if choice.ProjectID != game.ColonyShipProjectID || choice.TechnologyID != game.ColonyShipTechnologyID || choice.ProductionCostPP != 500 {
			t.Fatalf("Colony Ship legal action=%+v", choice)
		}
	}
	if !foundChoice {
		t.Fatal("seat legal-action surface missing Colony Ship")
	}
	if _, err := s.ConstructionChoices(2, state.Colonies[0].ID, rules); err == nil {
		t.Fatal("foreign seat unexpectedly received Colony Ship construction choices")
	}

	queue, err := game.NewQueueColonyShipCommand(1, game.QueueColonyShipPayload{ColonyID: state.Colonies[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	view := submitColonyShipSessionTurn(t, s, resolver, []protocol.Command{queue})
	if len(view.State.StrategicFleets) != 1 {
		t.Fatalf("Observer strategic fleets after build=%+v", view.State.StrategicFleets)
	}
	fleet := view.State.StrategicFleets[0]
	if fleet.SpecialKind != core.StrategicFleetSpecialColonyShip || fleet.AtSystemID != home.ID || fleet.FTLSpeed != 2 {
		t.Fatalf("Observer built Colony Ship=%+v", fleet)
	}
	if !observerHasEvent(view, "colony.colony_ship_queued") || !observerHasEvent(view, "colony.colony_ship_completed") {
		t.Fatalf("Observer history missing Colony Ship build events: %+v", view.Events)
	}
	fleetID := fleet.ID
	completeColonyShipSessionTurn(t, s)

	move, err := game.NewMoveFleetCommand(1, game.MoveFleetPayload{FleetID: fleetID, DestinationSystemID: target.ID})
	if err != nil {
		t.Fatal(err)
	}
	view = submitColonyShipSessionTurn(t, s, resolver, []protocol.Command{move})
	if len(view.State.StrategicFleets) != 1 || view.State.StrategicFleets[0].AtSystemID != 0 || view.State.StrategicFleets[0].DestinationSystemID != target.ID || view.State.StrategicFleets[0].RemainingTurns != 1 {
		t.Fatalf("Observer transit state=%+v", view.State.StrategicFleets)
	}
	if !observerHasEvent(view, "empire.fleet_movement_started") || !observerHasEvent(view, "empire.fleet_movement_progressed") {
		t.Fatalf("Observer history missing movement events: %+v", view.Events)
	}
	view.State.StrategicFleets[0].RemainingTurns = 99
	isolated, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if isolated.State.StrategicFleets[0].RemainingTurns != 1 {
		t.Fatalf("Observer mutation leaked into authority: %+v", isolated.State.StrategicFleets[0])
	}
	completeColonyShipSessionTurn(t, s)

	view = submitColonyShipSessionTurn(t, s, resolver, nil)
	if len(view.State.StrategicFleets) != 1 || view.State.StrategicFleets[0].AtSystemID != target.ID || view.State.StrategicFleets[0].DestinationSystemID != 0 || view.State.StrategicFleets[0].RemainingTurns != 0 {
		t.Fatalf("Observer arrival state=%+v", view.State.StrategicFleets)
	}
	if !observerHasEvent(view, "empire.fleet_arrived") {
		t.Fatalf("Observer history missing arrival: %+v", view.Events)
	}
	completeColonyShipSessionTurn(t, s)

	colonize, err := game.NewColonizePlanetCommand(1, game.ColonizePlanetPayload{FleetID: fleetID, PlanetID: targetPlanetID})
	if err != nil {
		t.Fatal(err)
	}
	view = submitColonyShipSessionTurn(t, s, resolver, []protocol.Command{colonize})
	if len(view.State.StrategicFleets) != 0 {
		t.Fatalf("Observer retained consumed Colony Ship: %+v", view.State.StrategicFleets)
	}
	if len(view.State.Colonies) != 2 || view.State.Galaxy.Systems[1].Planets[0].ColonyID == 0 {
		t.Fatalf("Observer missing second Colony: colonies=%+v target=%+v", view.State.Colonies, view.State.Galaxy.Systems[1].Planets[0])
	}
	newColony := view.State.Colonies[1]
	if newColony.EmpireID != state.Empires[0].ID || newColony.PlanetID != targetPlanetID || len(newColony.Buildings) != 0 {
		t.Fatalf("Observer new Colony=%+v", newColony)
	}
	if !observerHasEvent(view, "empire.planet_colonized") || !observerHasEvent(view, "empire.colony_ship_consumed") {
		t.Fatalf("Observer history missing colonization lifecycle: %+v", view.Events)
	}
}
