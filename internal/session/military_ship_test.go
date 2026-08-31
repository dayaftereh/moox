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

func runMilitaryShipSessionScenario(t *testing.T, gameID string) ObserverView {
	t.Helper()
	state, seats := twoSeatFixture(t)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	rules.MineralIndustryPerWorker["abundant"] = 600
	state.Empires[0].KnownTechnologyIDs = []int{58, 120, 167, 187}
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

	save, err := game.NewSaveMilitaryDesignCommand(1, game.SaveMilitaryDesignPayload{Name: "Sentinel", HullID: game.SupportedMilitaryHullID, StrategicPictureID: 2})
	if err != nil {
		t.Fatal(err)
	}
	view := submitColonyShipSessionTurn(t, s, resolver, []protocol.Command{save})
	if len(view.State.ShipDesigns) != 1 {
		t.Fatalf("Observer designs=%+v", view.State.ShipDesigns)
	}
	design := view.State.ShipDesigns[0]
	if design.Revision != 1 || design.Name != "Sentinel" || design.Spec.HullID != "frigate" || design.Spec.ProductionCostPP != 25 {
		t.Fatalf("Observer military design=%+v", design)
	}
	if !observerHasEvent(view, "empire.military_design_created") {
		t.Fatalf("Observer history missing design creation: %+v", view.Events)
	}
	completeColonyShipSessionTurn(t, s)

	choices, err := s.ConstructionChoices(1, state.Colonies[0].ID, rules)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, choice := range choices {
		if choice.ProjectKind == core.ConstructionProjectMilitaryShip && choice.ShipDesignID == design.ID {
			found = true
			if choice.ShipDesignRevision != 1 || choice.ProductionCostPP != 25 {
				t.Fatalf("session military legal action=%+v", choice)
			}
		}
	}
	if !found {
		t.Fatalf("session legal actions missing military design %d", design.ID)
	}

	queue, err := game.NewQueueMilitaryShipCommand(1, game.QueueMilitaryShipPayload{ColonyID: state.Colonies[0].ID, ShipDesignID: design.ID})
	if err != nil {
		t.Fatal(err)
	}
	view = submitColonyShipSessionTurn(t, s, resolver, []protocol.Command{queue})
	if len(view.State.Ships) != 1 {
		t.Fatalf("Observer Ships after build=%+v", view.State.Ships)
	}
	ship := view.State.Ships[0]
	if ship.SourceDesignID != design.ID || ship.SourceDesignRevision != 1 || ship.Name != "Sentinel" || ship.Spec.StrategicPictureID != 2 {
		t.Fatalf("Observer built Ship=%+v", ship)
	}
	var fleet *core.StrategicFleet
	for i := range view.State.StrategicFleets {
		if len(view.State.StrategicFleets[i].ShipIDs) == 1 && view.State.StrategicFleets[i].ShipIDs[0] == ship.ID {
			fleet = &view.State.StrategicFleets[i]
			break
		}
	}
	if fleet == nil || fleet.Role != core.StrategicFleetRoleCombat || fleet.AtSystemID != state.Galaxy.Systems[0].ID {
		t.Fatalf("Observer military Fleet=%+v fleets=%+v", fleet, view.State.StrategicFleets)
	}
	if !observerHasEvent(view, "colony.military_ship_queued") || !observerHasEvent(view, "colony.military_ship_completed") {
		t.Fatalf("Observer history missing military Construction lifecycle: %+v", view.Events)
	}

	shipID := ship.ID
	fleetID := fleet.ID
	view.State.ShipDesigns[0].Name = "mutated observer design"
	view.State.Ships[0].Name = "mutated observer ship"
	for i := range view.State.StrategicFleets {
		if view.State.StrategicFleets[i].ID == fleetID {
			view.State.StrategicFleets[i].ShipIDs[0] = 0
		}
	}
	isolated, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if isolated.State.ShipDesigns[0].Name != "Sentinel" || isolated.State.Ships[0].Name != "Sentinel" {
		t.Fatalf("Observer military mutation leaked into authority: design=%+v ship=%+v", isolated.State.ShipDesigns[0], isolated.State.Ships[0])
	}
	fleet = nil
	for i := range isolated.State.StrategicFleets {
		if isolated.State.StrategicFleets[i].ID == fleetID {
			fleet = &isolated.State.StrategicFleets[i]
			break
		}
	}
	if fleet == nil || len(fleet.ShipIDs) != 1 || fleet.ShipIDs[0] != shipID {
		t.Fatalf("Observer Fleet mutation leaked into authority: %+v", fleet)
	}
	return isolated
}

func TestGameSessionMilitaryDesignBuildObserverFlow(t *testing.T) {
	view := runMilitaryShipSessionScenario(t, "game-military-flow")
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
		t.Fatalf("military Observer state changed across save/load round trip\nfirst=%s\nsecond=%s", encoded, reencoded)
	}
}

func TestGameSessionMilitaryDesignBuildReplaysDeterministically(t *testing.T) {
	first := runMilitaryShipSessionScenario(t, "game-military-replay")
	second := runMilitaryShipSessionScenario(t, "game-military-replay")
	firstState, err := core.MarshalState(first.State)
	if err != nil {
		t.Fatal(err)
	}
	secondState, err := core.MarshalState(second.State)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstState, secondState) {
		t.Fatalf("identical military session replay diverged\nfirst=%s\nsecond=%s", firstState, secondState)
	}
	if !reflect.DeepEqual(first.Events, second.Events) {
		t.Fatalf("identical military event history diverged\nfirst=%+v\nsecond=%+v", first.Events, second.Events)
	}
}
