package app

import (
	"errors"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/session"
)

func registerReferenceRunnerTestGame(t *testing.T, host *Host, gameID string, profile game.ReferenceTriangleTechnologyProfile) {
	t.Helper()
	generated, err := host.newGameRules.NewReferenceTriangleGame(game.ReferenceTriangleSeed)
	if err != nil {
		t.Fatal(err)
	}
	if err := host.newGameRules.ApplyReferenceTriangleTechnologyProfile(generated.State, profile); err != nil {
		t.Fatal(err)
	}
	player := generated.Players[0]
	gameSession, err := session.NewGameSession(gameID, generated.State, []session.Seat{{
		ID: player.SeatID, EmpireID: player.EmpireID, Name: player.Name, Controller: session.ControllerLocalHuman,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := host.Register(Registration{
		Session: gameSession, Resolver: host.newGameResolver, ImmediateResolver: host.newGameImmediateResolver,
		Reference: &ReferenceGameInfo{ScenarioID: "runner-test", ProfileID: string(profile), ControlSeatID: 1},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReferenceAdvanceUsesNormalTurnPipeline(t *testing.T) {
	host := loadNewGameHost(t)
	const gameID = "reference-runner-turn"
	registerReferenceRunnerTestGame(t, host, gameID, game.ReferenceTriangleProfileBaseline)
	before, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := host.AdvanceReference(gameID, ReferenceAdvanceRequest{
		SchemaVersion: SchemaVersion, SeatID: 1, BaseRevision: before.View.Revision,
		Mode: ReferenceAdvanceTurns, Turns: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.StartTurn != 1 || result.EndTurn != 2 || result.TurnsAdvanced != 1 || result.StopReason != "requested_turns_reached" || result.FinalPhase != session.PhasePlanning {
		t.Fatalf("reference advance result=%+v", result)
	}
	after, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if after.View.Turn != 2 || after.View.Revision != result.GameRevision || after.ChangeSequence != result.ChangeSequence {
		t.Fatalf("after snapshot turn=%d revision=%d change=%d result=%+v", after.View.Turn, after.View.Revision, after.ChangeSequence, result)
	}
}

func TestReferenceAdvanceRejectsOrdinaryGameAndStaleRevision(t *testing.T) {
	host := loadNewGameHost(t)
	if _, err := host.CreateGame(CreateGameRequest{GameID: "ordinary", Seed: 0x8123, Settings: appNewGameSettings()}); err != nil {
		t.Fatal(err)
	}
	ordinary, err := host.PlayerSnapshot("ordinary", 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = host.AdvanceReference("ordinary", ReferenceAdvanceRequest{
		SchemaVersion: SchemaVersion, SeatID: 1, BaseRevision: ordinary.View.Revision,
		Mode: ReferenceAdvanceTurns, Turns: 1,
	})
	if !errors.Is(err, ErrReferenceForbidden) {
		t.Fatalf("ordinary reference advance err=%v", err)
	}

	const gameID = "reference-stale"
	registerReferenceRunnerTestGame(t, host, gameID, game.ReferenceTriangleProfileBaseline)
	snapshot, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = host.AdvanceReference(gameID, ReferenceAdvanceRequest{
		SchemaVersion: SchemaVersion, SeatID: 1, BaseRevision: snapshot.View.Revision + 1,
		Mode: ReferenceAdvanceTurns, Turns: 1,
	})
	if !errors.Is(err, ErrReferenceStaleRevision) {
		t.Fatalf("stale reference advance err=%v", err)
	}
}

func TestReferenceAdvanceRejectsNonEmptyDraft(t *testing.T) {
	host := loadNewGameHost(t)
	const gameID = "reference-draft"
	registerReferenceRunnerTestGame(t, host, gameID, game.ReferenceTriangleProfileBaseline)
	snapshot, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	hosted, err := host.lookup(gameID)
	if err != nil {
		t.Fatal(err)
	}
	hosted.mu.Lock()
	hosted.planningDrafts[1] = PlanningDraft{
		SchemaVersion: SchemaVersion, GameID: gameID, SeatID: 1, Turn: snapshot.View.Turn,
		BaseRevision: snapshot.View.Revision, DraftRevision: 1,
		Orders: []PlanningDraftOrder{{Key: "test", Kind: "test"}},
	}
	hosted.mu.Unlock()
	_, err = host.AdvanceReference(gameID, ReferenceAdvanceRequest{
		SchemaVersion: SchemaVersion, SeatID: 1, BaseRevision: snapshot.View.Revision,
		Mode: ReferenceAdvanceTurns, Turns: 1,
	})
	if !errors.Is(err, ErrReferenceNotReady) {
		t.Fatalf("draft reference advance err=%v", err)
	}
}

func TestReferenceAdvanceUntilConstructionRequiresRealCompletionEvent(t *testing.T) {
	host := loadNewGameHost(t)
	const gameID = "reference-construction"
	generated, err := host.newGameRules.NewReferenceTriangleGame(game.ReferenceTriangleSeed)
	if err != nil {
		t.Fatal(err)
	}
	if err := host.newGameRules.ApplyReferenceTriangleTechnologyProfile(generated.State, game.ReferenceTriangleProfileAllTech); err != nil {
		t.Fatal(err)
	}
	if len(generated.State.Colonies) == 0 {
		t.Fatal("triangle has no colonies")
	}
	colony := &generated.State.Colonies[0]
	colony.Construction = &core.ConstructionState{
		ProjectKind: core.ConstructionProjectBuilding, ProjectID: "holo_simulator", ProgressPP: 1_000_000,
	}
	player := generated.Players[0]
	gameSession, err := session.NewGameSession(gameID, generated.State, []session.Seat{{
		ID: player.SeatID, EmpireID: player.EmpireID, Name: player.Name, Controller: session.ControllerLocalHuman,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := host.Register(Registration{
		Session: gameSession, Resolver: host.newGameResolver, ImmediateResolver: host.newGameImmediateResolver,
		Reference: &ReferenceGameInfo{ScenarioID: "construction-test", ProfileID: string(game.ReferenceTriangleProfileAllTech), ControlSeatID: 1},
	}); err != nil {
		t.Fatal(err)
	}
	before, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := host.AdvanceReference(gameID, ReferenceAdvanceRequest{
		SchemaVersion: SchemaVersion, SeatID: 1, BaseRevision: before.View.Revision,
		Mode: ReferenceAdvanceUntilConstruction, ColonyID: colony.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.StopReason != "construction_completed" || result.TurnsAdvanced != 1 {
		t.Fatalf("construction advance=%+v", result)
	}
	after, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, built := range after.View.Colonies[0].Buildings {
		if built == "holo_simulator" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("holo_simulator missing after completion: %+v", after.View.Colonies[0].Buildings)
	}
}

func TestReferencePersistenceCannotEscalateImportAndRestoreRetainsTrustedMetadata(t *testing.T) {
	const gameID = "reference-persistence"
	host := loadNewGameHost(t)
	registerReferenceRunnerTestGame(t, host, gameID, game.ReferenceTriangleProfileMidTech)
	snapshot, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Reference == nil {
		t.Fatal("trusted reference metadata missing before export")
	}
	data, err := host.ExportLiveSnapshot(gameID)
	if err != nil {
		t.Fatal(err)
	}

	importHost := loadNewGameHost(t)
	summary, err := importHost.ImportLiveSnapshot(data)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Reference != nil {
		t.Fatalf("generic import escalated reference metadata: %+v", summary.Reference)
	}
	imported, err := importHost.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if imported.Reference != nil {
		t.Fatalf("generic imported game has reference controls: %+v", imported.Reference)
	}

	advanced, err := host.AdvanceReference(gameID, ReferenceAdvanceRequest{
		SchemaVersion: SchemaVersion, SeatID: 1, BaseRevision: snapshot.View.Revision,
		Mode: ReferenceAdvanceTurns, Turns: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if advanced.EndTurn != 2 {
		t.Fatalf("advanced=%+v", advanced)
	}
	if _, err := host.RestoreLiveSnapshot(gameID, data); err != nil {
		t.Fatal(err)
	}
	restored, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if restored.View.Turn != 1 || restored.Reference == nil || restored.Reference.ProfileID != string(game.ReferenceTriangleProfileMidTech) {
		t.Fatalf("restored reference snapshot=%+v reference=%+v", restored.View, restored.Reference)
	}
}
