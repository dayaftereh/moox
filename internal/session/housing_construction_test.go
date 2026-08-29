package session

import (
	"path/filepath"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func TestGameSessionHousingChoiceAuthorityAndObserverLifecycle(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("game-housing", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	colonyID := state.Colonies[0].ID
	choices, err := s.ConstructionChoices(1, colonyID, rules)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, choice := range choices {
		if choice.ProjectKind == core.ConstructionProjectHousing {
			found = choice.ProjectID == game.HousingProjectID
		}
	}
	if !found {
		t.Fatalf("seat legal-action surface missing Housing: %+v", choices)
	}
	if _, err := s.ConstructionChoices(2, colonyID, rules); err == nil {
		t.Fatal("foreign seat unexpectedly received Housing construction choices")
	}

	command, err := game.NewQueueHousingCommand(1, game.QueueHousingPayload{ColonyID: colonyID})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-housing",
		SeatID:        1,
		Turn:          1,
		BaseRevision:  1,
		Commands:      []protocol.Command{command},
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-housing",
		SeatID:        2,
		Turn:          1,
		BaseRevision:  1,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	construction := observer.State.Colonies[0].Construction
	if construction == nil || construction.ProjectKind != core.ConstructionProjectHousing || construction.ProjectID != game.HousingProjectID || construction.ProgressPP != 0 {
		t.Fatalf("Observer Housing state=%+v", construction)
	}
	baseRaceMultiplier := rules.RaceModifiers[observer.State.Empires[0].RaceID].PopulationGrowthMultiplier
	if observer.State.Colonies[0].PopulationDynamics.GrowthMultiplier <= baseRaceMultiplier {
		t.Fatalf("Observer next-state growth multiplier=%v did not retain Housing bonus above race multiplier=%v", observer.State.Colonies[0].PopulationDynamics.GrowthMultiplier, baseRaceMultiplier)
	}
	foundQueued := false
	for _, event := range observer.Events {
		if event.Kind == "colony.housing_queued" {
			foundQueued = true
			break
		}
	}
	if !foundQueued {
		t.Fatalf("Observer history missing Housing queue event: %+v", observer.Events)
	}
}
