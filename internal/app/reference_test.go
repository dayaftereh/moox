package app

import (
	"errors"
	"math"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
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

func TestReferenceGrantBCIsTrustedAndServerAuthoritative(t *testing.T) {
	host := loadNewGameHost(t)
	const gameID = "reference-grant-bc"
	registerReferenceRunnerTestGame(t, host, gameID, game.ReferenceTriangleProfileAllTech)
	before, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	start := before.View.Empire.Treasury.BalanceBC
	result, err := host.GrantReferenceBC(gameID, ReferenceGrantBCRequest{
		SchemaVersion: SchemaVersion, SeatID: 1, BaseRevision: before.View.Revision, AmountBC: 1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.AmountBC != 1000 || result.BalanceBC != start+1000 {
		t.Fatalf("grant result=%+v start=%v", result, start)
	}
	after, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if after.View.Empire.Treasury.BalanceBC != start+1000 {
		t.Fatalf("treasury=%v want=%v", after.View.Empire.Treasury.BalanceBC, start+1000)
	}

	if _, err := host.CreateGame(CreateGameRequest{GameID: "ordinary-grant", Seed: 0x8124, Settings: appNewGameSettings()}); err != nil {
		t.Fatal(err)
	}
	ordinary, err := host.PlayerSnapshot("ordinary-grant", 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = host.GrantReferenceBC("ordinary-grant", ReferenceGrantBCRequest{
		SchemaVersion: SchemaVersion, SeatID: 1, BaseRevision: ordinary.View.Revision, AmountBC: 100,
	})
	if !errors.Is(err, ErrReferenceForbidden) {
		t.Fatalf("ordinary BC grant err=%v", err)
	}
}

func TestReferenceGrantBCRejectsUnapprovedAmount(t *testing.T) {
	host := loadNewGameHost(t)
	const gameID = "reference-grant-bc-invalid"
	registerReferenceRunnerTestGame(t, host, gameID, game.ReferenceTriangleProfileBaseline)
	before, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = host.GrantReferenceBC(gameID, ReferenceGrantBCRequest{
		SchemaVersion: SchemaVersion, SeatID: 1, BaseRevision: before.View.Revision, AmountBC: 500,
	})
	if !errors.Is(err, ErrReferenceNotReady) {
		t.Fatalf("invalid BC amount err=%v", err)
	}
}

func TestHostConstructionBuyoutUsesNormalImmediateAuthorityAndNextTurnCompletion(t *testing.T) {
	host := loadNewGameHost(t)
	const gameID = "reference-normal-buyout"
	registerReferenceRunnerTestGame(t, host, gameID, game.ReferenceTriangleProfileAllTech)

	before, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	colonyID := before.View.Colonies[0].ID
	queue, err := game.NewSetConstructionQueueCommand(1, game.SetConstructionQueuePayload{
		ColonyID: colonyID,
		Items:    []game.ConstructionQueueItem{{ProjectKind: core.ConstructionProjectBuilding, ProjectID: "holo_simulator"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := host.SubmitTurn(gameID, protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion, GameID: gameID, SeatID: 1,
		Turn: before.View.Turn, BaseRevision: before.View.Revision, Commands: []protocol.Command{queue},
	}); err != nil {
		t.Fatal(err)
	}

	queued, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	var quote *game.ConstructionBuyoutQuote
	for i := range queued.ConstructionBuyouts {
		if queued.ConstructionBuyouts[i].ColonyID == colonyID {
			quote = &queued.ConstructionBuyouts[i]
			break
		}
	}
	if quote == nil || quote.CostBC <= 0 {
		t.Fatalf("buyout quote=%+v", quote)
	}

	grant, err := host.GrantReferenceBC(gameID, ReferenceGrantBCRequest{
		SchemaVersion: SchemaVersion, SeatID: 1, BaseRevision: queued.View.Revision, AmountBC: 10000,
	})
	if err != nil {
		t.Fatal(err)
	}
	funded, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	buy, err := game.NewBuyConstructionCommand(1, game.BuyConstructionPayload{ColonyID: colonyID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := host.SubmitImmediateCommand(gameID, 1, funded.View.Revision, buy); err != nil {
		t.Fatal(err)
	}
	bought, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := bought.View.Empire.Treasury.BalanceBC, grant.BalanceBC-quote.CostBC; math.Abs(got-want) > 1e-9 {
		t.Fatalf("buyout treasury=%v want=%v", got, want)
	}
	if bought.View.Colonies[0].Construction == nil || math.Abs(bought.View.Colonies[0].Construction.ProgressPP-quote.ProductionCostPP) > 1e-9 {
		t.Fatalf("bought construction=%+v quote=%+v", bought.View.Colonies[0].Construction, quote)
	}

	result, err := host.AdvanceReference(gameID, ReferenceAdvanceRequest{
		SchemaVersion: SchemaVersion, SeatID: 1, BaseRevision: bought.View.Revision, Mode: ReferenceAdvanceTurns, Turns: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.TurnsAdvanced != 1 {
		t.Fatalf("advance after buyout=%+v", result)
	}
	completed, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, building := range completed.View.Colonies[0].Buildings {
		if building == "holo_simulator" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("bought holo_simulator did not complete on normal turn: %+v", completed.View.Colonies[0].Buildings)
	}
}
