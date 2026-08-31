package session

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func runOutpostSessionScenario(t *testing.T, gameID string) ObserverView {
	t.Helper()
	state, seats := twoSeatFixture(t)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	rules.MineralIndustryPerWorker["abundant"] = 600
	state.Empires[0].KnownTechnologyIDs = []int{109, 120, 167}
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
	s, err := NewGameSession(gameID, state, seats)
	if err != nil {
		t.Fatal(err)
	}

	choices, err := s.ConstructionChoices(1, state.Colonies[0].ID, rules)
	if err != nil {
		t.Fatal(err)
	}
	foundChoice := false
	for _, choice := range choices {
		if choice.ProjectKind != core.ConstructionProjectOutpostShip {
			continue
		}
		foundChoice = true
		if choice.ProjectID != game.OutpostShipProjectID || choice.TechnologyID != game.OutpostShipTechnologyID || choice.ProductionCostPP != 100 {
			t.Fatalf("Outpost Ship legal action=%+v", choice)
		}
	}
	if !foundChoice {
		t.Fatal("seat legal-action surface missing Outpost Ship")
	}

	queue, err := game.NewQueueOutpostShipCommand(1, game.QueueOutpostShipPayload{ColonyID: state.Colonies[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	view := submitColonyShipSessionTurn(t, s, resolver, []protocol.Command{queue})
	if len(view.State.StrategicFleets) != 1 {
		t.Fatalf("Observer strategic fleets after Outpost build=%+v", view.State.StrategicFleets)
	}
	fleet := view.State.StrategicFleets[0]
	if fleet.SpecialKind != core.StrategicFleetSpecialOutpostShip || fleet.AtSystemID != home.ID || fleet.FTLSpeed != 2 {
		t.Fatalf("Observer built Outpost Ship=%+v", fleet)
	}
	if !observerHasEvent(view, "colony.outpost_ship_queued") || !observerHasEvent(view, "colony.outpost_ship_completed") {
		t.Fatalf("Observer history missing Outpost Ship build events: %+v", view.Events)
	}
	fleetID := fleet.ID
	completeColonyShipSessionTurn(t, s)

	move, err := game.NewMoveFleetCommand(1, game.MoveFleetPayload{FleetID: fleetID, DestinationSystemID: target.ID})
	if err != nil {
		t.Fatal(err)
	}
	view = submitColonyShipSessionTurn(t, s, resolver, []protocol.Command{move})
	if len(view.State.StrategicFleets) != 1 || view.State.StrategicFleets[0].AtSystemID != 0 || view.State.StrategicFleets[0].DestinationSystemID != target.ID || view.State.StrategicFleets[0].RemainingTurns != 1 {
		t.Fatalf("Observer Outpost transit state=%+v", view.State.StrategicFleets)
	}
	if !observerHasEvent(view, "empire.fleet_movement_started") || !observerHasEvent(view, "empire.fleet_movement_progressed") {
		t.Fatalf("Observer history missing Outpost movement events: %+v", view.Events)
	}
	completeColonyShipSessionTurn(t, s)

	view = submitColonyShipSessionTurn(t, s, resolver, nil)
	if len(view.State.StrategicFleets) != 1 || view.State.StrategicFleets[0].AtSystemID != target.ID || view.State.StrategicFleets[0].DestinationSystemID != 0 || view.State.StrategicFleets[0].RemainingTurns != 0 {
		t.Fatalf("Observer Outpost arrival state=%+v", view.State.StrategicFleets)
	}
	if !observerHasEvent(view, "empire.fleet_arrived") {
		t.Fatalf("Observer history missing Outpost arrival: %+v", view.Events)
	}
	completeColonyShipSessionTurn(t, s)

	deploy, err := game.NewDeployOutpostCommand(1, game.DeployOutpostPayload{FleetID: fleetID, PlanetID: targetPlanetID})
	if err != nil {
		t.Fatal(err)
	}
	view = submitColonyShipSessionTurn(t, s, resolver, []protocol.Command{deploy})
	if len(view.State.StrategicFleets) != 0 {
		t.Fatalf("Observer retained consumed Outpost Ship: %+v", view.State.StrategicFleets)
	}
	if len(view.State.Outposts) != 1 || view.State.Outposts[0].EmpireID != state.Empires[0].ID || view.State.Outposts[0].PlanetID != targetPlanetID {
		t.Fatalf("Observer Outposts=%+v", view.State.Outposts)
	}
	if view.State.Galaxy.Systems[1].Planets[0].OutpostID != view.State.Outposts[0].ID {
		t.Fatalf("Observer Planet Outpost link=%+v outposts=%+v", view.State.Galaxy.Systems[1].Planets[0], view.State.Outposts)
	}
	if !observerHasEvent(view, "empire.outpost_deployed") || !observerHasEvent(view, "empire.outpost_ship_consumed") {
		t.Fatalf("Observer history missing Outpost deployment lifecycle: %+v", view.Events)
	}

	outpostID := view.State.Outposts[0].ID
	view.State.Outposts[0].EmpireID = 0
	view.State.Galaxy.Systems[1].Planets[0].OutpostID = 0
	isolated, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if len(isolated.State.Outposts) != 1 || isolated.State.Outposts[0].ID != outpostID || isolated.State.Outposts[0].EmpireID != state.Empires[0].ID || isolated.State.Galaxy.Systems[1].Planets[0].OutpostID != outpostID {
		t.Fatalf("Observer Outpost mutation leaked into authority: outposts=%+v planet=%+v", isolated.State.Outposts, isolated.State.Galaxy.Systems[1].Planets[0])
	}
	return isolated
}

func TestGameSessionOutpostShipBuildMoveDeployObserverFlow(t *testing.T) {
	view := runOutpostSessionScenario(t, "game-outpost-flow")
	encoded, err := core.MarshalState(view.State)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := core.UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	reencoded, err := core.MarshalState(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatalf("Observer Outpost state changed across save/load round trip\nfirst=%s\nsecond=%s", encoded, reencoded)
	}
}

func TestGameSessionOutpostFlowReplaysDeterministically(t *testing.T) {
	first := runOutpostSessionScenario(t, "game-outpost-replay")
	second := runOutpostSessionScenario(t, "game-outpost-replay")
	firstState, err := core.MarshalState(first.State)
	if err != nil {
		t.Fatal(err)
	}
	secondState, err := core.MarshalState(second.State)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstState, secondState) {
		t.Fatalf("identical Outpost session replay diverged\nfirst=%s\nsecond=%s", firstState, secondState)
	}
	if !reflect.DeepEqual(first.Events, second.Events) {
		t.Fatalf("identical Outpost session event history diverged\nfirst=%+v\nsecond=%+v", first.Events, second.Events)
	}
}
