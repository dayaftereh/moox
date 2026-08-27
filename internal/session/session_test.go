package session

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func twoSeatFixture(t *testing.T) (*core.GameState, []Seat) {
	t.Helper()
	state := core.NewSmallFixture(1234)
	secondEmpire := core.Empire{ID: state.NewID(), Name: "Second Empire", RaceID: "alkari"}
	state.Empires = append(state.Empires, secondEmpire)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	return state, []Seat{
		{ID: 2, EmpireID: secondEmpire.ID, Name: "AI", Controller: ControllerBuiltinAI},
		{ID: 1, EmpireID: state.Empires[0].ID, Name: "Human", Controller: ControllerLocalHuman},
	}
}

func makeBatch(t *testing.T, gameID string, seatID protocol.SeatID, turn, revision uint64, kind string) protocol.CommandBatch {
	t.Helper()
	command, err := protocol.NewCommand(1, kind, map[string]any{"seat": seatID})
	if err != nil {
		t.Fatal(err)
	}
	return protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        gameID,
		SeatID:        seatID,
		Turn:          turn,
		BaseRevision:  revision,
		Commands:      []protocol.Command{command},
	}
}

func TestParallelSubmissionsBecomeDeterministicReplayOrder(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("game-1", state, seats)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.SubmitTurn(makeBatch(t, "game-1", 2, 1, 1, "test.ai_action")); err != nil {
		t.Fatal(err)
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if observer.Phase != PhasePlanning || !observer.Seats[1].Submitted {
		t.Fatalf("first parallel submission not visible to observer: %+v", observer.Seats)
	}
	if len(observer.Events) != 1 {
		t.Fatalf("arrival order leaked into authoritative events: %v", observer.Events)
	}

	if err := s.SubmitTurn(makeBatch(t, "game-1", 1, 1, 1, "test.human_action")); err != nil {
		t.Fatal(err)
	}
	if sView, _ := s.ObserverView(); sView.Phase != PhaseStrategicResolution {
		t.Fatalf("expected strategic resolution, got %q", sView.Phase)
	}
	batches, err := s.SubmittedBatches()
	if err != nil {
		t.Fatal(err)
	}
	if batches[0].SeatID != 1 || batches[1].SeatID != 2 {
		t.Fatalf("batches are not stable seat order: %d, %d", batches[0].SeatID, batches[1].SeatID)
	}
	observer, _ = s.ObserverView()
	if len(observer.Events) != 4 {
		t.Fatalf("unexpected event count: %d", len(observer.Events))
	}
	if observer.Events[1].Kind != "turn_submitted" || observer.Events[1].SeatID != 1 || observer.Events[2].SeatID != 2 {
		t.Fatalf("submission events are not deterministic seat order: %+v", observer.Events)
	}
}

func TestPlayerViewDoesNotExposeEnemyEmpireState(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("game-1", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	view, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if view.Empire.ID != state.Empires[0].ID {
		t.Fatalf("wrong empire projected: %d", view.Empire.ID)
	}
	for _, colony := range view.Colonies {
		if colony.EmpireID != view.Empire.ID {
			t.Fatalf("enemy colony leaked into player view: %+v", colony)
		}
	}
	if len(view.Seats) != 2 || view.Seats[0].Seat.ID != 1 || view.Seats[1].Seat.ID != 2 {
		t.Fatalf("seat status projection not stable: %+v", view.Seats)
	}
}

func TestObserverAndPlayerViewsAreDetachedCopies(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("game-1", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	observer.State.Empires[0].Name = "tampered"
	fresh, _ := s.ObserverView()
	if fresh.State.Empires[0].Name == "tampered" {
		t.Fatal("observer view mutated authoritative state")
	}
}

func TestParallelBattlesCompleteInStableReplayOrder(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("game-1", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	for _, seatID := range []protocol.SeatID{1, 2} {
		if err := s.SubmitTurn(makeBatch(t, "game-1", seatID, 1, 1, "test.action")); err != nil {
			t.Fatal(err)
		}
	}
	views, err := s.BeginEncounters([]EncounterSpec{
		{Participants: []protocol.SeatID{1, 2}},
		{Participants: []protocol.SeatID{2, 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 2 || views[0].Spec.ID != 1 || views[1].Spec.ID != 2 || views[0].Spec.Seed == views[1].Spec.Seed {
		t.Fatalf("unexpected encounter identities: %+v", views)
	}

	if err := s.CompleteBattle(2, battle.Result{WinnerSeats: []protocol.SeatID{2}, Outcome: "victory"}); err != nil {
		t.Fatal(err)
	}
	mid, _ := s.ObserverView()
	if mid.Phase != PhaseEncounters {
		t.Fatalf("session advanced before all battles completed: %q", mid.Phase)
	}
	for _, event := range mid.Events {
		if event.Kind == "battle_completed" {
			t.Fatal("wall-clock battle completion order leaked into authoritative events")
		}
	}

	if err := s.CompleteBattle(1, battle.Result{WinnerSeats: []protocol.SeatID{1}, Outcome: "victory"}); err != nil {
		t.Fatal(err)
	}
	finished, _ := s.ObserverView()
	if finished.Phase != PhasePostResolution {
		t.Fatalf("expected post resolution, got %q", finished.Phase)
	}
	var completedIDs []uint64
	for _, event := range finished.Events {
		if event.Kind != "battle_completed" {
			continue
		}
		var view battle.View
		if err := json.Unmarshal(event.Data, &view); err != nil {
			t.Fatal(err)
		}
		completedIDs = append(completedIDs, view.Spec.ID)
	}
	if !reflect.DeepEqual(completedIDs, []uint64{1, 2}) {
		t.Fatalf("battle completion replay order = %v", completedIDs)
	}

	if err := s.CompleteTurn(); err != nil {
		t.Fatal(err)
	}
	next, _ := s.ObserverView()
	if next.Turn != 2 || next.Revision != 2 || next.Phase != PhasePlanning {
		t.Fatalf("unexpected next turn state: turn=%d rev=%d phase=%q", next.Turn, next.Revision, next.Phase)
	}
	for _, seat := range next.Seats {
		if seat.Submitted {
			t.Fatal("submission was not reset for next turn")
		}
	}
}

func TestDraftTelemetryIsObserverOnlyAndNonAuthoritative(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("game-1", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.PublishDraftTelemetry(2, "ai.plan", "consider colony defense", map[string]any{"goal": "defense"}); err != nil {
		t.Fatal(err)
	}
	observer, _ := s.ObserverView()
	if len(observer.Telemetry) != 1 || observer.Telemetry[0].SeatID != 2 {
		t.Fatalf("observer did not receive telemetry: %+v", observer.Telemetry)
	}
	if len(observer.Events) != 1 {
		t.Fatal("draft telemetry entered authoritative event history")
	}
	player, _ := s.PlayerView(1)
	encoded, err := json.Marshal(player)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) == "" || json.Valid(encoded) == false {
		t.Fatal("player view is not serializable")
	}
}

func TestSubmissionRejectsStaleRevision(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("game-1", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	batch := makeBatch(t, "game-1", 1, 1, 2, "test.action")
	if err := s.SubmitTurn(batch); err == nil {
		t.Fatal("expected stale/future revision to be rejected")
	}
}

func TestConcurrentParallelSubmissionsReachResolution(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("game-concurrent", state, seats)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 2)
	for _, seatID := range []protocol.SeatID{1, 2} {
		seatID := seatID
		wg.Add(1)
		go func() {
			defer wg.Done()
			errors <- s.SubmitTurn(makeBatch(t, "game-concurrent", seatID, 1, 1, "test.concurrent"))
		}()
	}
	wg.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}

	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if observer.Phase != PhaseStrategicResolution {
		t.Fatalf("concurrent submissions left session in %q", observer.Phase)
	}
	if !observer.Seats[0].Submitted || !observer.Seats[1].Submitted {
		t.Fatalf("not all concurrent submissions were retained: %+v", observer.Seats)
	}
	if observer.Events[1].SeatID != 1 || observer.Events[2].SeatID != 2 {
		t.Fatalf("concurrent arrival order affected replay: %+v", observer.Events)
	}
}

func TestResolveStrategicCommitsDetachedStateAndDomainEvents(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("game-resolve", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	for _, seatID := range []protocol.SeatID{2, 1} {
		if err := s.SubmitTurn(makeBatch(t, "game-resolve", seatID, 1, 1, "test.resolve")); err != nil {
			t.Fatal(err)
		}
	}

	var resolverState *core.GameState
	resolver := game.ResolverFunc(func(ctx game.ResolveContext, state *core.GameState, batches []protocol.CommandBatch) (game.Resolution, error) {
		if batches[0].SeatID != 1 || batches[1].SeatID != 2 {
			t.Fatalf("resolver received unstable batch order: %d, %d", batches[0].SeatID, batches[1].SeatID)
		}
		state.Empires[0].Name = "Resolved Empire"
		resolverState = state
		event, err := game.NewDomainEvent("test.resolved", 1, 1, map[string]any{"ok": true})
		if err != nil {
			return game.Resolution{}, err
		}
		return game.Resolution{State: state, Events: []game.DomainEvent{event}}, nil
	})
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}

	view, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if view.Phase != PhasePostResolution || view.Revision != 2 {
		t.Fatalf("unexpected resolved session: phase=%q revision=%d", view.Phase, view.Revision)
	}
	if view.State.Empires[0].Name != "Resolved Empire" {
		t.Fatalf("resolved state not committed: %q", view.State.Empires[0].Name)
	}
	found := false
	for _, event := range view.Events {
		if event.Kind == "test.resolved" {
			found = true
			if event.Revision != 2 || event.SeatID != 1 || event.CommandSequence != 1 {
				t.Fatalf("unexpected domain event metadata: %+v", event)
			}
		}
	}
	if !found {
		t.Fatal("resolved domain event missing from observer history")
	}

	resolverState.Empires[0].Name = "mutated after resolve"
	fresh, _ := s.ObserverView()
	if fresh.State.Empires[0].Name != "Resolved Empire" {
		t.Fatal("resolver retained mutable access to authoritative state")
	}
}

func TestResolveStrategicRejectsInvalidEncounterAtomically(t *testing.T) {
	state, seats := twoSeatFixture(t)
	originalName := state.Empires[0].Name
	s, err := NewGameSession("game-atomic", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	for _, seatID := range []protocol.SeatID{1, 2} {
		if err := s.SubmitTurn(makeBatch(t, "game-atomic", seatID, 1, 1, "test.resolve")); err != nil {
			t.Fatal(err)
		}
	}
	before, _ := s.ObserverView()

	resolver := game.ResolverFunc(func(ctx game.ResolveContext, state *core.GameState, batches []protocol.CommandBatch) (game.Resolution, error) {
		state.Empires[0].Name = "must not commit"
		return game.Resolution{
			State:      state,
			Encounters: []game.Encounter{{Participants: []protocol.SeatID{1, 99}}},
		}, nil
	})
	if err := s.ResolveStrategic(resolver); err == nil {
		t.Fatal("expected invalid encounter to reject resolution")
	}
	after, _ := s.ObserverView()
	if after.Revision != before.Revision || after.Phase != PhaseStrategicResolution {
		t.Fatalf("failed resolution changed session metadata: before=%d/%q after=%d/%q", before.Revision, before.Phase, after.Revision, after.Phase)
	}
	if after.State.Empires[0].Name != originalName {
		t.Fatalf("failed resolution partially committed state: %q", after.State.Empires[0].Name)
	}
	if len(after.Events) != len(before.Events) {
		t.Fatalf("failed resolution partially committed events: before=%d after=%d", len(before.Events), len(after.Events))
	}
}

func TestGameSessionResolvesConstructionCommand(t *testing.T) {
	state, seats := twoSeatFixture(t)
	state.Empires[0].KnownTechnologyIDs = []int{86}
	s, err := NewGameSession("game-construction", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	command, err := game.NewQueueBuildingCommand(1, game.QueueBuildingPayload{ColonyID: state.Colonies[0].ID, BuildingID: "holo_simulator"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-construction",
		SeatID:        1,
		Turn:          1,
		BaseRevision:  1,
		Commands:      []protocol.Command{command},
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-construction",
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
	if observer.Phase != PhasePostResolution || observer.Revision != 2 {
		t.Fatalf("unexpected resolved session: phase=%q revision=%d", observer.Phase, observer.Revision)
	}
	construction := observer.State.Colonies[0].Construction
	if construction == nil || construction.BuildingID != "holo_simulator" || construction.ProgressPP <= 0 {
		t.Fatalf("construction not materialized through GameSession: %+v", construction)
	}
	var queued, progressed bool
	for _, event := range observer.Events {
		switch event.Kind {
		case "colony.construction_queued":
			queued = true
			if event.SeatID != 1 || event.CommandSequence != 1 {
				t.Fatalf("queue event lost command attribution: %+v", event)
			}
		case "colony.construction_progressed":
			progressed = true
			if event.SeatID != 0 || event.CommandSequence != 0 {
				t.Fatalf("system progress event has player attribution: %+v", event)
			}
		}
	}
	if !queued || !progressed {
		t.Fatalf("observer missing construction events: queued=%v progressed=%v", queued, progressed)
	}
}

func TestGameSessionBuildingChoicesUseSeatAuthority(t *testing.T) {
	state, seats := twoSeatFixture(t)
	state.Empires[0].KnownTechnologyIDs = []int{86, 141}
	s, err := NewGameSession("game-choices", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	choices, err := s.BuildingChoices(1, state.Colonies[0].ID, rules)
	if err != nil {
		t.Fatal(err)
	}
	if len(choices) != 2 || choices[0].BuildingID != "holo_simulator" || choices[1].BuildingID != "pleasure_dome" {
		t.Fatalf("unexpected seat building choices: %+v", choices)
	}
	if _, err := s.BuildingChoices(2, state.Colonies[0].ID, rules); err == nil {
		t.Fatal("expected other seat to be denied production choices for foreign colony")
	}
}

func TestGameSessionCompletesResearchThroughAuthoritativeBoundary(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], game.NewGameTechnologyOptions{Level: game.NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	state.Empires[0].Research = &core.ResearchState{TechFieldID: 56, SelectionMode: core.ResearchSelectionChooseOne, TechnologyIDs: []int{155}, ProgressRP: 150}
	s, err := NewGameSession("game-research", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	for _, seatID := range []protocol.SeatID{1, 2} {
		if err := s.SubmitTurn(protocol.CommandBatch{
			SchemaVersion: protocol.CommandSchemaVersion,
			GameID:        "game-research",
			SeatID:        seatID,
			Turn:          1,
			BaseRevision:  1,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	before, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if before.Phase != PhasePostResolution {
		t.Fatalf("phase before research completion=%q", before.Phase)
	}
	if err := s.CompleteResearchField(state.Empires[0].ID, resolver); err != nil {
		t.Fatal(err)
	}
	after, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if after.Revision != before.Revision+1 {
		t.Fatalf("research completion revision=%d want=%d", after.Revision, before.Revision+1)
	}
	if after.State.Empires[0].Research != nil {
		t.Fatalf("research still active: %+v", after.State.Empires[0].Research)
	}
	var known bool
	for _, technologyID := range after.State.Empires[0].KnownTechnologyIDs {
		if technologyID == 155 {
			known = true
			break
		}
	}
	if !known {
		t.Fatalf("technology 155 not acquired: %v", after.State.Empires[0].KnownTechnologyIDs)
	}
	last := after.Events[len(after.Events)-1]
	if last.Kind != "empire.research_completed" || last.Scope != protocol.EventScopeStrategic || last.SeatID != 0 || last.CommandSequence != 0 {
		t.Fatalf("unexpected research completion event: %+v", last)
	}
}

func TestGameSessionAutomaticallyResolvesGuaranteedResearchBreakthrough(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], game.NewGameTechnologyOptions{Level: game.NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	state.Empires[0].Research = &core.ResearchState{
		TechFieldID:   56,
		SelectionMode: core.ResearchSelectionChooseOne,
		TechnologyIDs: []int{155},
		ProgressRP:    299,
	}
	s, err := NewGameSession("game-auto-research", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	for _, seatID := range []protocol.SeatID{1, 2} {
		if err := s.SubmitTurn(protocol.CommandBatch{
			SchemaVersion: protocol.CommandSchemaVersion,
			GameID:        "game-auto-research",
			SeatID:        seatID,
			Turn:          1,
			BaseRevision:  1,
		}); err != nil {
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
	if observer.Phase != PhasePostResolution {
		t.Fatalf("automatic research left session in phase %q", observer.Phase)
	}
	if observer.State.Empires[0].Research != nil {
		t.Fatalf("guaranteed research still active: %+v", observer.State.Empires[0].Research)
	}
	var progressed *game.ResearchProgressedEvent
	var completed *game.ResearchCompletedEvent
	for _, event := range observer.Events {
		switch event.Kind {
		case "empire.research_progressed":
			var payload game.ResearchProgressedEvent
			if err := json.Unmarshal(event.Data, &payload); err != nil {
				t.Fatal(err)
			}
			progressed = &payload
		case "empire.research_completed":
			var payload game.ResearchCompletedEvent
			if err := json.Unmarshal(event.Data, &payload); err != nil {
				t.Fatal(err)
			}
			completed = &payload
		}
	}
	if progressed == nil || !progressed.Breakthrough || progressed.ChancePercent != 100 {
		t.Fatalf("observer missing guaranteed research progress event: %+v", progressed)
	}
	if len(progressed.TechnologyKeys) != 1 || progressed.TechnologyKeys[0] != "research_laboratory" {
		t.Fatalf("observer progress lacks speaking technology key: %+v", progressed)
	}
	if completed == nil || len(completed.TechnologyKeys) != 1 || completed.TechnologyKeys[0] != "research_laboratory" {
		t.Fatalf("observer completion lacks Technology 155 (research_laboratory): %+v", completed)
	}
}

func TestGameSessionResearchChoicesAndSelectResearchShareAuthority(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], game.NewGameTechnologyOptions{Level: game.NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("game-research-choice", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	choices, err := s.ResearchChoices(1, rules)
	if err != nil {
		t.Fatal(err)
	}
	var field4 *game.ResearchChoice
	for i := range choices {
		if choices[i].TechFieldID == 4 {
			field4 = &choices[i]
			break
		}
	}
	if field4 == nil || field4.BaseCostRP != 80 || len(field4.TechnologyKeys) != 3 {
		t.Fatalf("authoritative ResearchChoices missing TechField 4: %+v", choices)
	}
	command, err := game.NewSelectResearchCommand(1, game.SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-research-choice",
		SeatID:        1,
		Turn:          1,
		BaseRevision:  1,
		Commands:      []protocol.Command{command},
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-research-choice",
		SeatID:        2,
		Turn:          1,
		BaseRevision:  1,
	}); err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if observer.State.Empires[0].Research == nil || observer.State.Empires[0].Research.TechFieldID != 4 {
		t.Fatalf("authoritative select_research not materialized: %+v", observer.State.Empires[0].Research)
	}
	if !reflect.DeepEqual(observer.State.Empires[0].Research.TechnologyIDs, []int{56}) {
		t.Fatalf("server-authoritative Technology IDs=%v want=[56]", observer.State.Empires[0].Research.TechnologyIDs)
	}
	found := false
	for _, event := range observer.Events {
		if event.Kind != "empire.research_selected" {
			continue
		}
		var selected game.ResearchSelectedEvent
		if err := json.Unmarshal(event.Data, &selected); err != nil {
			t.Fatal(err)
		}
		if selected.TechFieldID == 4 && selected.SelectionMode == core.ResearchSelectionChooseOne && reflect.DeepEqual(selected.TechnologyKeys, []string{"reinforced_hull"}) {
			found = true
		}
	}
	if !found {
		t.Fatal("Observer did not receive speaking empire.research_selected event")
	}
	if choices, err := s.ResearchChoices(1, rules); err != nil || len(choices) != 0 {
		t.Fatalf("active research should leave no new ResearchChoices: choices=%+v err=%v", choices, err)
	}
}
