package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

const persistenceCanonicalGameID = "live-ai-canonical"

func newPersistenceCanonicalAIHost(t *testing.T) *Host {
	t.Helper()
	host := loadNewGameHost(t)
	registerPostContactCanonicalAIGame(t, host, persistenceCanonicalGameID)

	return host
}

func advancePersistenceAIToTurn(t *testing.T, host *Host, turn uint64) {
	t.Helper()
	for step := 0; step < 1000; step++ {
		summary := host.ListGames()[0]
		if summary.Turn >= turn {
			if summary.Turn != turn || summary.Phase != session.PhasePlanning {
				t.Fatalf("wanted planning turn %d, got turn=%d phase=%q", turn, summary.Turn, summary.Phase)
			}
			return
		}
		if _, err := host.AdvanceAutomation(persistenceCanonicalGameID); err != nil {
			t.Fatalf("advance to turn %d step %d: %v", turn, step, err)
		}
	}
	t.Fatalf("did not reach turn %d", turn)
}

func completePersistenceAI(t *testing.T, host *Host) []byte {
	t.Helper()
	for step := 0; step < 1000; step++ {
		summary := host.ListGames()[0]
		if summary.Phase == session.PhaseCompleted {
			hosted, err := host.lookup(persistenceCanonicalGameID)
			if err != nil {
				t.Fatal(err)
			}
			data, err := hosted.session.MarshalCompletedSnapshot()
			if err != nil {
				t.Fatal(err)
			}
			return data
		}
		if _, err := host.AdvanceAutomation(persistenceCanonicalGameID); err != nil {
			t.Fatalf("complete persistence AI step %d: %v", step, err)
		}
	}
	t.Fatal("persistence AI match did not complete")
	return nil
}

func TestLivePersistenceCanonicalAIResumeIsExact(t *testing.T) {
	uninterrupted := completePersistenceAI(t, newPersistenceCanonicalAIHost(t))

	beforeSave := newPersistenceCanonicalAIHost(t)
	advancePersistenceAIToTurn(t, beforeSave, 250)
	saved, err := beforeSave.ExportLiveSnapshot(persistenceCanonicalGameID)
	if err != nil {
		t.Fatal(err)
	}

	fresh := loadNewGameHost(t)
	imported, err := fresh.ImportLiveSnapshot(saved)
	if err != nil {
		t.Fatal(err)
	}
	if imported.GameID != persistenceCanonicalGameID || imported.Turn != 250 || imported.Phase != session.PhasePlanning || imported.ChangeSequence != 1 {
		t.Fatalf("imported summary=%+v", imported)
	}
	reexported, err := fresh.ExportLiveSnapshot(persistenceCanonicalGameID)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(saved, reexported) {
		t.Fatalf("fresh-host live re-export changed: saved=%d reexported=%d", len(saved), len(reexported))
	}

	resumed := completePersistenceAI(t, fresh)
	if !bytes.Equal(uninterrupted, resumed) {
		t.Fatalf("save/load AI continuation diverged: uninterrupted=%d resumed=%d", len(uninterrupted), len(resumed))
	}

	completedLive, err := fresh.ExportLiveSnapshot(persistenceCanonicalGameID)
	if err != nil {
		t.Fatal(err)
	}
	completedHost := loadNewGameHost(t)
	completedSummary, err := completedHost.ImportLiveSnapshot(completedLive)
	if err != nil {
		t.Fatal(err)
	}
	if completedSummary.Phase != session.PhaseCompleted || completedSummary.Result == nil {
		t.Fatalf("completed live import summary=%+v", completedSummary)
	}
	completedReexport, err := completedHost.ExportLiveSnapshot(persistenceCanonicalGameID)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(completedLive, completedReexport) {
		t.Fatal("completed live snapshot changed on fresh-host import/re-export")
	}
	completedHosted, err := completedHost.lookup(persistenceCanonicalGameID)
	if err != nil {
		t.Fatal(err)
	}
	completedBytes, err := completedHosted.session.MarshalCompletedSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(uninterrupted, completedBytes) {
		t.Fatal("completed live roundtrip changed existing CompletedSnapshot bytes")
	}
}

func TestRestoreLiveSnapshotIsAtomicAndNotifiesExistingSubscribers(t *testing.T) {
	host := loadNewGameHost(t)
	const gameID = "live-restore"
	if _, err := host.CreateGame(CreateGameRequest{GameID: gameID, Seed: 0x8009, Settings: appNewGameSettings()}); err != nil {
		t.Fatal(err)
	}
	baseline, err := host.ExportLiveSnapshot(gameID)
	if err != nil {
		t.Fatal(err)
	}
	before, err := host.ObserverSnapshot(gameID)
	if err != nil {
		t.Fatal(err)
	}
	ch, cancel, err := host.Subscribe(gameID)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	if _, err := host.RestoreLiveSnapshot(gameID, []byte("{")); !errors.Is(err, ErrInvalidSave) {
		t.Fatalf("invalid restore error=%v", err)
	}
	var corrupt map[string]any
	if err := json.Unmarshal(baseline, &corrupt); err != nil {
		t.Fatal(err)
	}
	corrupt["next_event_sequence"] = float64(999999)
	corruptBytes, _ := json.Marshal(corrupt)
	if _, err := host.RestoreLiveSnapshot(gameID, corruptBytes); !errors.Is(err, ErrInvalidSave) {
		t.Fatalf("counter-corrupt restore error=%v", err)
	}
	var schemaMismatch map[string]any
	if err := json.Unmarshal(baseline, &schemaMismatch); err != nil {
		t.Fatal(err)
	}
	schemaMismatch["schema_version"] = float64(99)
	schemaMismatchBytes, _ := json.Marshal(schemaMismatch)
	if _, err := host.RestoreLiveSnapshot(gameID, schemaMismatchBytes); !errors.Is(err, ErrInvalidSave) {
		t.Fatalf("schema-version restore error=%v", err)
	}
	var partial map[string]any
	if err := json.Unmarshal(baseline, &partial); err != nil {
		t.Fatal(err)
	}
	delete(partial, "next_battle_id")
	partialBytes, _ := json.Marshal(partial)
	if _, err := host.RestoreLiveSnapshot(gameID, partialBytes); !errors.Is(err, ErrInvalidSave) {
		t.Fatalf("partial restore error=%v", err)
	}
	afterRejected, err := host.ObserverSnapshot(gameID)
	if err != nil {
		t.Fatal(err)
	}
	if afterRejected.ChangeSequence != before.ChangeSequence || afterRejected.View.Revision != before.View.Revision {
		t.Fatalf("rejected restore mutated host: before=%+v after=%+v", before, afterRejected)
	}

	receipt, err := host.RestoreLiveSnapshot(gameID, baseline)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.ChangeSequence != before.ChangeSequence+1 || receipt.GameRevision != before.View.Revision {
		t.Fatalf("restore receipt=%+v before=%+v", receipt, before)
	}
	select {
	case notification := <-ch:
		if notification.Reason != "game_loaded" || notification.ChangeSequence != receipt.ChangeSequence || notification.GameRevision != receipt.GameRevision {
			t.Fatalf("restore notification=%+v", notification)
		}
	case <-time.After(time.Second):
		t.Fatal("restore emitted no subscriber invalidation")
	}
	afterRestore, err := host.ExportLiveSnapshot(gameID)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(baseline, afterRestore) {
		t.Fatal("same-save restore changed authoritative live bytes")
	}
}

func TestImportLiveSnapshotRulesMismatchAndDuplicateAreRejected(t *testing.T) {
	source := loadNewGameHost(t)
	const gameID = "live-import"
	if _, err := source.CreateGame(CreateGameRequest{GameID: gameID, Seed: 0x8009, Settings: appNewGameSettings()}); err != nil {
		t.Fatal(err)
	}
	data, err := source.ExportLiveSnapshot(gameID)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	simulationMismatch := make(map[string]any, len(document))
	for key, value := range document {
		simulationMismatch[key] = value
	}
	simulationMismatch["simulation_compat_version"] = float64(99)
	simulationBytes, _ := json.Marshal(simulationMismatch)
	freshSimulation := loadNewGameHost(t)
	if _, err := freshSimulation.ImportLiveSnapshot(simulationBytes); !errors.Is(err, ErrInvalidSave) {
		t.Fatalf("simulation mismatch import error=%v", err)
	}

	ruleset := document["ruleset"].(map[string]any)
	ruleset["sha256"] = "deadbeef"
	mismatched, _ := json.Marshal(document)
	fresh := loadNewGameHost(t)
	if _, err := fresh.ImportLiveSnapshot(mismatched); !errors.Is(err, ErrRulesetMismatch) {
		t.Fatalf("rules mismatch import error=%v", err)
	}
	if games := fresh.ListGames(); len(games) != 0 {
		t.Fatalf("rules mismatch import registered game: %+v", games)
	}
	if _, err := fresh.ImportLiveSnapshot(data); err != nil {
		t.Fatal(err)
	}
	if _, err := fresh.ImportLiveSnapshot(data); !errors.Is(err, ErrGameExists) {
		t.Fatalf("duplicate import error=%v", err)
	}
}
func TestRestoreLiveSnapshotRejectsEmbeddedGameIDMismatch(t *testing.T) {
	host := loadNewGameHost(t)
	const gameID = "live-id-target"
	if _, err := host.CreateGame(CreateGameRequest{GameID: gameID, Seed: 0x8009, Settings: appNewGameSettings()}); err != nil {
		t.Fatal(err)
	}
	data, err := host.ExportLiveSnapshot(gameID)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	document["game_id"] = "other-game"
	mismatched, _ := json.Marshal(document)
	before := host.ListGames()[0]
	if _, err := host.RestoreLiveSnapshot(gameID, mismatched); !errors.Is(err, ErrGameIDMismatch) {
		t.Fatalf("game ID mismatch restore error=%v", err)
	}
	after := host.ListGames()[0]
	if before.ChangeSequence != after.ChangeSequence || before.Revision != after.Revision {
		t.Fatalf("game ID mismatch mutated host before=%+v after=%+v", before, after)
	}
}
func TestLivePersistenceRejectsCustomResolver(t *testing.T) {
	state := core.NewSmallFixture(0x8008)
	gameSession, err := session.NewGameSession("custom-resolver", state, []session.Seat{{ID: 1, EmpireID: state.Empires[0].ID, Name: "Human", Controller: session.ControllerLocalHuman}})
	if err != nil {
		t.Fatal(err)
	}
	custom := game.ResolverFunc(func(ctx game.ResolveContext, input *core.GameState, batches []protocol.CommandBatch) (game.Resolution, error) {
		return game.Resolution{State: input}, nil
	})
	host := NewHost()
	if err := host.Register(Registration{Session: gameSession, Resolver: custom}); err != nil {
		t.Fatal(err)
	}
	if _, err := host.ExportLiveSnapshot("custom-resolver"); !errors.Is(err, ErrPersistenceUnsupported) {
		t.Fatalf("custom resolver persistence error=%v", err)
	}
}
