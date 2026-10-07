package main

import (
	"path/filepath"
	"testing"

	"moox/internal/app"
	"moox/internal/game"
	"moox/internal/session"
)

func TestDefaultHTTPAddress(t *testing.T) {
	if defaultHTTPAddress != "127.0.0.1:7171" {
		t.Fatalf("default HTTP address=%q, want 127.0.0.1:7171", defaultHTTPAddress)
	}
}

func TestLoopbackAddressPolicy(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:7171", "[::1]:7171", "localhost:7171"} {
		if !isLoopbackAddress(addr) {
			t.Fatalf("expected loopback address %q", addr)
		}
	}
	for _, addr := range []string{"0.0.0.0:7171", ":7171", "192.168.1.20:7171", "bad"} {
		if isLoopbackAddress(addr) {
			t.Fatalf("unexpected loopback approval for %q", addr)
		}
	}
}

func TestProductionHostStartsEmpty(t *testing.T) {
	host, err := newServerHost(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	if games := host.ListGames(); len(games) != 0 {
		t.Fatalf("production host started with games: %+v", games)
	}
}

func TestDemoFixtureRequiresExplicitFlag(t *testing.T) {
	host, err := newServerHost(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"), true, false)
	if err != nil {
		t.Fatal(err)
	}
	games := host.ListGames()
	if len(games) != 1 || games[0].GameID != demoGameID {
		t.Fatalf("explicit demo fixture games=%+v", games)
	}
}
func TestReferenceGamesRequireExplicitFlag(t *testing.T) {
	host, err := newServerHost(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"), false, true)
	if err != nil {
		t.Fatal(err)
	}
	games := host.ListGames()
	if len(games) != 5 {
		t.Fatalf("reference games=%+v, want exactly five", games)
	}
	seen := map[string]bool{}
	for _, summary := range games {
		seen[summary.GameID] = true
	}
	for _, gameID := range []string{standardReferenceGameID, triangleReferenceGameID, triangleMidTechReferenceGameID, triangleAllTechReferenceGameID, triangleCombatReviewReferenceGameID} {
		if !seen[gameID] {
			t.Fatalf("reference game IDs=%v missing %q", seen, gameID)
		}
	}

	standardHuman, err := host.PlayerSnapshot(standardReferenceGameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if standardHuman.Reference != nil {
		t.Fatalf("legacy standard reference unexpectedly has reference controls: %+v", standardHuman.Reference)
	}
	standardAI, err := host.PlayerSnapshot(standardReferenceGameID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if standardHuman.View.Seat.Seat.Controller != "local_human" || standardAI.View.Seat.Seat.Controller != "builtin_ai" {
		t.Fatalf("standard controllers human=%q ai=%q", standardHuman.View.Seat.Seat.Controller, standardAI.View.Seat.Seat.Controller)
	}

	for _, tc := range []struct {
		gameID     string
		scenarioID string
		profileID  string
		fields     int
		techs      int
	}{
		{triangleReferenceGameID, game.ReferenceTriangleScenarioID, string(game.ReferenceTriangleProfileBaseline), 7, 19},
		{triangleMidTechReferenceGameID, game.ReferenceTriangleMidTechScenarioID, string(game.ReferenceTriangleProfileMidTech), 36, 94},
		{triangleAllTechReferenceGameID, game.ReferenceTriangleAllTechScenarioID, string(game.ReferenceTriangleProfileAllTech), 75, 202},
		{triangleCombatReviewReferenceGameID, game.ReferenceTriangleCombatReviewScenarioID, string(game.ReferenceTriangleProfileAllTech), 75, 202},
	} {
		human, err := host.PlayerSnapshot(tc.gameID, 1)
		if err != nil {
			t.Fatal(err)
		}
		ai1, err := host.PlayerSnapshot(tc.gameID, 2)
		if err != nil {
			t.Fatal(err)
		}
		ai2, err := host.PlayerSnapshot(tc.gameID, 3)
		if err != nil {
			t.Fatal(err)
		}
		if human.View.Seat.Seat.Controller != "local_human" || ai1.View.Seat.Seat.Controller != "builtin_ai" || ai2.View.Seat.Seat.Controller != "builtin_ai" {
			t.Fatalf("%s controllers human=%q ai1=%q ai2=%q", tc.gameID, human.View.Seat.Seat.Controller, ai1.View.Seat.Seat.Controller, ai2.View.Seat.Seat.Controller)
		}
		if human.Reference == nil || human.Reference.ScenarioID != tc.scenarioID || human.Reference.ProfileID != tc.profileID ||
			human.Reference.ControlSeatID != 1 || human.Reference.MaxTurnsPerRequest != app.ReferenceMaxTurnsPerRequest ||
			human.Reference.MaxConstructionTurns != app.ReferenceMaxConstructionTurns {
			t.Fatalf("%s reference=%+v", tc.gameID, human.Reference)
		}
		if tc.profileID != string(game.ReferenceTriangleProfileBaseline) {
			if len(human.View.Empire.KnownTechnologyFieldIDs) != tc.fields || len(human.View.Empire.KnownTechnologyIDs) != tc.techs {
				t.Fatalf("%s fields=%d techs=%d", tc.gameID, len(human.View.Empire.KnownTechnologyFieldIDs), len(human.View.Empire.KnownTechnologyIDs))
			}
		}
	}
}

func TestCombatReviewReferenceStopsAtSingleFightSystemBattle(t *testing.T) {
	host, err := newServerHost(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"), false, true)
	if err != nil {
		t.Fatal(err)
	}
	before, err := host.PlayerSnapshot(triangleCombatReviewReferenceGameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if before.Reference == nil || !before.Reference.HoldBuiltinAIPlanning {
		t.Fatalf("combat review reference=%+v want held builtin planning", before.Reference)
	}
	var fightSystemID uint64
	if before.Decision == nil {
		t.Fatal("combat review has no decision view")
	}
	for _, system := range before.Decision.Strategic.Galaxy.Systems {
		if system.Name == "Fight System" {
			fightSystemID = uint64(system.ID)
			break
		}
	}
	if fightSystemID == 0 {
		t.Fatal("Fight System missing from combat review")
	}
	result, err := host.AdvanceReference(triangleCombatReviewReferenceGameID, app.ReferenceAdvanceRequest{
		SchemaVersion: app.SchemaVersion,
		SeatID:        1,
		BaseRevision:  before.View.Revision,
		Mode:          app.ReferenceAdvanceTurns,
		Turns:         1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.StopReason != "interactive_boundary" || result.FinalPhase != session.PhaseEncounters || result.TurnsAdvanced != 0 {
		t.Fatalf("advance=%+v want immediate encounter boundary", result)
	}
	after, err := host.PlayerSnapshot(triangleCombatReviewReferenceGameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Battles) != 1 {
		t.Fatalf("battles=%d want exactly one: %+v", len(after.Battles), after.Battles)
	}
	if uint64(after.Battles[0].Spec.SystemID) != fightSystemID {
		t.Fatalf("battle system=%d want Fight System %d", after.Battles[0].Spec.SystemID, fightSystemID)
	}
}
