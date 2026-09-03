package session

import (
	"path/filepath"
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func addSessionSameSystemPlanet(state *core.GameState) core.ID {
	planet := core.Planet{
		ID:        state.NewID(),
		Name:      "Alpha II",
		Orbit:     2,
		SizeID:    "medium",
		MineralID: "abundant",
		GravityID: "normal_g",
		ClimateID: "terran",
	}
	state.Galaxy.Systems[0].Planets = append(state.Galaxy.Systems[0].Planets, planet)
	return planet.ID
}

func sessionColonyOwnsBuilding(colony core.Colony, buildingID string) bool {
	for _, owned := range colony.Buildings {
		if owned == buildingID {
			return true
		}
	}
	return false
}

func loadSessionColonyBaseRules(t *testing.T) *game.EconomyRules {
	t.Helper()
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	return rules
}

func TestGameSessionColonyBaseBuildCreatesMandatoryPostResolutionColonization(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules := loadSessionColonyBaseRules(t)
	rules.MineralIndustryPerWorker["abundant"] = 600
	state.Empires[0].KnownTechnologyIDs = []int{game.ColonyBaseTechnologyID}
	sourceID := state.Colonies[0].ID
	targetID := addSessionSameSystemPlanet(state)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("game-colony-base", state, seats)
	if err != nil {
		t.Fatal(err)
	}

	choices, err := s.ConstructionChoices(1, sourceID, rules)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, choice := range choices {
		if choice.ProjectKind == core.ConstructionProjectBuilding && choice.ProjectID == game.ColonyBaseBuildingID {
			found = true
			if choice.ProductionCostPP != 200 || choice.TechnologyID != 40 {
				t.Fatalf("Colony Base construction choice=%+v", choice)
			}
		}
	}
	if !found {
		t.Fatal("session legal-action surface missing Colony Base")
	}

	queue, err := game.NewQueueBuildingCommand(1, game.QueueBuildingPayload{ColonyID: sourceID, BuildingID: game.ColonyBaseBuildingID})
	if err != nil {
		t.Fatal(err)
	}
	afterBuild := submitColonyShipSessionTurn(t, s, resolver, []protocol.Command{queue})
	if !sessionColonyOwnsBuilding(afterBuild.State.Colonies[0], game.ColonyBaseBuildingID) {
		t.Fatalf("completed Colony Base not materialized: %+v", afterBuild.State.Colonies[0].Buildings)
	}
	if !observerHasEvent(afterBuild, "colony.building_completed") {
		t.Fatal("Observer history missing generic Colony Base building completion")
	}
	sourcePopulation := afterBuild.State.Colonies[0].Population
	if len(afterBuild.ColonyBaseResolutions) != 1 || afterBuild.ColonyBaseResolutions[0].SourceColonyID != sourceID || len(afterBuild.ColonyBaseResolutions[0].TargetPlanetIDs) != 1 || afterBuild.ColonyBaseResolutions[0].TargetPlanetIDs[0] != targetID {
		t.Fatalf("Observer Colony Base resolutions=%+v", afterBuild.ColonyBaseResolutions)
	}
	seatResolutions, err := s.ColonyBaseResolutions(1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(seatResolutions, afterBuild.ColonyBaseResolutions) {
		t.Fatalf("seat resolution surface=%+v observer=%+v", seatResolutions, afterBuild.ColonyBaseResolutions)
	}
	foreign, err := s.ColonyBaseResolutions(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(foreign) != 0 {
		t.Fatalf("foreign seat received Colony Base resolutions: %+v", foreign)
	}
	player, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(player.ColonyBaseResolutions, afterBuild.ColonyBaseResolutions) {
		t.Fatalf("PlayerView Colony Base resolutions=%+v", player.ColonyBaseResolutions)
	}

	// Resolution projection is isolated from authoritative state/client mutation.
	afterBuild.ColonyBaseResolutions[0].TargetPlanetIDs[0] = 999999
	isolated, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if isolated.ColonyBaseResolutions[0].TargetPlanetIDs[0] != targetID {
		t.Fatalf("Observer Colony Base target mutation leaked: %+v", isolated.ColonyBaseResolutions)
	}
	if err := s.CompleteTurn(); err == nil {
		t.Fatal("CompleteTurn unexpectedly advanced with unresolved Colony Base")
	}

	colonize, err := game.NewColonizeWithBaseCommand(1, game.ColonizeWithBasePayload{SourceColonyID: sourceID, PlanetID: targetID})
	if err != nil {
		t.Fatal(err)
	}
	beforeRevision := isolated.Revision
	if err := s.ResolveColonyBaseCommand(1, s.Status().Revision, colonize, resolver); err != nil {
		t.Fatal(err)
	}
	afterColonize, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if afterColonize.Revision != beforeRevision+1 {
		t.Fatalf("post-resolution revision=%d want=%d", afterColonize.Revision, beforeRevision+1)
	}
	if len(afterColonize.ColonyBaseResolutions) != 0 {
		t.Fatalf("resolved Colony Base still pending: %+v", afterColonize.ColonyBaseResolutions)
	}
	if len(afterColonize.State.Colonies) != 2 {
		t.Fatalf("colony count=%d want=2", len(afterColonize.State.Colonies))
	}
	if sessionColonyOwnsBuilding(afterColonize.State.Colonies[0], game.ColonyBaseBuildingID) {
		t.Fatal("source Colony retained consumed Colony Base")
	}
	if !reflect.DeepEqual(afterColonize.State.Colonies[0].Population, sourcePopulation) {
		t.Fatalf("source Population changed: got=%+v want=%+v", afterColonize.State.Colonies[0].Population, sourcePopulation)
	}
	newColony := afterColonize.State.Colonies[1]
	if newColony.PlanetID != targetID || !reflect.DeepEqual(newColony.Population, core.NewAssimilatedPopulation(state.Empires[0].ID, 1, 0, 0)) {
		t.Fatalf("new Colony=%+v", newColony)
	}
	if len(afterColonize.Events) < 2 {
		t.Fatalf("Observer history too short after Colony Base resolution: %d", len(afterColonize.Events))
	}
	last := afterColonize.Events[len(afterColonize.Events)-2:]
	if last[0].Kind != "empire.planet_colonized" || last[1].Kind != "colony.colony_base_colonized" {
		t.Fatalf("post-resolution event order=%q,%q", last[0].Kind, last[1].Kind)
	}
	for _, event := range last {
		if event.Revision != afterColonize.Revision || event.SeatID != 1 || event.CommandSequence != 1 {
			t.Fatalf("Colony Base event metadata=%+v", event)
		}
	}
	if err := s.CompleteTurn(); err != nil {
		t.Fatalf("CompleteTurn rejected after Colony Base resolution: %v", err)
	}
	completed, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if completed.Turn != 2 || completed.Phase != PhasePlanning {
		t.Fatalf("completed Colony Base turn=%d phase=%q", completed.Turn, completed.Phase)
	}
}

func TestGameSessionColonyBaseTrashRefundResolvesTurnGuard(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules := loadSessionColonyBaseRules(t)
	state.Colonies[0].Buildings = append(state.Colonies[0].Buildings, game.ColonyBaseBuildingID)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("game-colony-base-trash", state, seats)
	if err != nil {
		t.Fatal(err)
	}

	post := submitColonyShipSessionTurn(t, s, resolver, nil)
	if len(post.ColonyBaseResolutions) != 1 || len(post.ColonyBaseResolutions[0].TargetPlanetIDs) != 0 {
		t.Fatalf("targetless Colony Base pending projection=%+v", post.ColonyBaseResolutions)
	}
	if err := s.CompleteTurn(); err == nil {
		t.Fatal("CompleteTurn unexpectedly advanced before Colony Base trash resolution")
	}
	balanceBefore := post.State.Empires[0].Treasury.BalanceBC
	trash, err := game.NewTrashColonyBaseCommand(1, game.TrashColonyBasePayload{SourceColonyID: state.Colonies[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveColonyBaseCommand(1, s.Status().Revision, trash, resolver); err != nil {
		t.Fatal(err)
	}
	after, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if len(after.ColonyBaseResolutions) != 0 {
		t.Fatalf("trashed Colony Base still pending: %+v", after.ColonyBaseResolutions)
	}
	if got := after.State.Empires[0].Treasury.BalanceBC; got != balanceBefore+100 {
		t.Fatalf("Treasury after Colony Base trash=%v want=%v", got, balanceBefore+100)
	}
	if !observerHasEvent(after, "colony.colony_base_trashed") {
		t.Fatal("Observer history missing Colony Base trash event")
	}
	if err := s.CompleteTurn(); err != nil {
		t.Fatalf("CompleteTurn rejected after Colony Base trash: %v", err)
	}
}
