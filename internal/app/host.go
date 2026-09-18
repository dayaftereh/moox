package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"

	"moox/internal/ai"
	"moox/internal/battle"
	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

const SchemaVersion = 1

var (
	ErrNotFound        = errors.New("not found")
	ErrSessionRejected = errors.New("session rejected")
	ErrGameExists      = errors.New("game already exists")
)

type GameSummary struct {
	SchemaVersion  int                `json:"schema_version"`
	GameID         string             `json:"game_id"`
	ChangeSequence uint64             `json:"change_sequence"`
	Revision       uint64             `json:"revision"`
	Turn           uint64             `json:"turn"`
	Phase          session.Phase      `json:"phase"`
	Result         *session.Result    `json:"result,omitempty"`
	Reference      *ReferenceGameInfo `json:"reference,omitempty"`
}

type PlayerSnapshot struct {
	SchemaVersion       int                            `json:"schema_version"`
	ChangeSequence      uint64                         `json:"change_sequence"`
	View                session.PlayerView             `json:"view"`
	Decision            *session.PlayerDecisionView    `json:"decision,omitempty"`
	Battles             []battle.View                  `json:"battles"`
	PlanningDraft       *PlanningDraft                 `json:"planning_draft,omitempty"`
	Reference           *ReferenceGameInfo             `json:"reference,omitempty"`
	ConstructionBuyouts []game.ConstructionBuyoutQuote `json:"construction_buyouts,omitempty"`
}

type PlanningDraftOrder struct {
	Key     string          `json:"key"`
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type PlanningDraft struct {
	SchemaVersion int                  `json:"schema_version"`
	GameID        string               `json:"game_id"`
	SeatID        protocol.SeatID      `json:"seat_id"`
	Turn          uint64               `json:"turn"`
	BaseRevision  uint64               `json:"base_revision"`
	DraftRevision uint64               `json:"draft_revision"`
	Orders        []PlanningDraftOrder `json:"orders"`
}

type PlanningDraftSnapshot struct {
	SchemaVersion  int           `json:"schema_version"`
	ChangeSequence uint64        `json:"change_sequence"`
	Draft          PlanningDraft `json:"draft"`
}

type PlanningPreviewSnapshot struct {
	SchemaVersion  int                     `json:"schema_version"`
	ChangeSequence uint64                  `json:"change_sequence"`
	Preview        session.PlanningPreview `json:"preview"`
}

type ObserverSnapshot struct {
	SchemaVersion  int                  `json:"schema_version"`
	ChangeSequence uint64               `json:"change_sequence"`
	View           session.ObserverView `json:"view"`
}

type Receipt struct {
	SchemaVersion  int    `json:"schema_version"`
	GameID         string `json:"game_id"`
	ChangeSequence uint64 `json:"change_sequence"`
	GameRevision   uint64 `json:"game_revision"`
}

type Notification struct {
	SchemaVersion  int    `json:"schema_version"`
	Kind           string `json:"kind"`
	GameID         string `json:"game_id"`
	ChangeSequence uint64 `json:"change_sequence"`
	GameRevision   uint64 `json:"game_revision"`
	Scope          string `json:"scope"`
	BattleID       uint64 `json:"battle_id,omitempty"`
	Reason         string `json:"reason"`
}

type Registration struct {
	Session           *session.GameSession
	Resolver          game.Resolver
	ImmediateResolver *game.EconomyResolver
	Reference         *ReferenceGameInfo
}

type Host struct {
	mu                       sync.RWMutex
	games                    map[string]*hostedGame
	newGameRules             *game.EconomyRules
	newGameResolver          game.Resolver
	newGameImmediateResolver *game.EconomyResolver
}

type hostedGame struct {
	mu                sync.Mutex
	session           *session.GameSession
	resolver          game.Resolver
	immediateResolver *game.EconomyResolver
	seats             []protocol.SeatID
	seatInfo          map[protocol.SeatID]session.Seat
	changeSequence    uint64
	nextSubscriberID  uint64
	subscribers       map[uint64]chan Notification
	planningDrafts    map[protocol.SeatID]PlanningDraft
	reference         *ReferenceGameInfo
}

func NewHost() *Host {
	return &Host{games: make(map[string]*hostedGame)}
}

func (h *Host) Register(reg Registration) error {
	if h == nil {
		return fmt.Errorf("host is nil")
	}
	if reg.Session == nil {
		return fmt.Errorf("session must not be nil")
	}
	if reg.Resolver == nil {
		return fmt.Errorf("resolver must not be nil")
	}
	status := reg.Session.Status()
	if status.GameID == "" {
		return fmt.Errorf("session game ID must not be empty")
	}
	observer, err := reg.Session.ObserverView()
	if err != nil {
		return fmt.Errorf("read session seats: %w", err)
	}
	seats := make([]protocol.SeatID, len(observer.Seats))
	seatInfo := make(map[protocol.SeatID]session.Seat, len(observer.Seats))
	for i := range observer.Seats {
		seats[i] = observer.Seats[i].Seat.ID
		seatInfo[observer.Seats[i].Seat.ID] = observer.Seats[i].Seat
	}
	sort.Slice(seats, func(i, j int) bool { return seats[i] < seats[j] })

	reference, err := normalizeReferenceInfo(reg.Reference)
	if err != nil {
		return err
	}
	if reference != nil && !containsSeat(seats, reference.ControlSeatID) {
		return fmt.Errorf("reference control seat %d is not registered in game %q", reference.ControlSeatID, status.GameID)
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if _, exists := h.games[status.GameID]; exists {
		return fmt.Errorf("%w: game %q", ErrGameExists, status.GameID)
	}
	h.games[status.GameID] = &hostedGame{
		session:           reg.Session,
		resolver:          reg.Resolver,
		immediateResolver: reg.ImmediateResolver,
		seats:             seats,
		seatInfo:          seatInfo,
		changeSequence:    1,
		nextSubscriberID:  1,
		subscribers:       make(map[uint64]chan Notification),
		planningDrafts:    make(map[protocol.SeatID]PlanningDraft),
		reference:         reference,
	}
	return nil
}

func (h *Host) ListGames() []GameSummary {
	if h == nil {
		return nil
	}
	h.mu.RLock()
	games := make([]*hostedGame, 0, len(h.games))
	for _, hosted := range h.games {
		games = append(games, hosted)
	}
	h.mu.RUnlock()

	summaries := make([]GameSummary, 0, len(games))
	for _, hosted := range games {
		hosted.mu.Lock()
		status := hosted.session.Status()
		summaries = append(summaries, GameSummary{
			SchemaVersion:  SchemaVersion,
			GameID:         status.GameID,
			ChangeSequence: hosted.changeSequence,
			Revision:       status.Revision,
			Turn:           status.Turn,
			Phase:          status.Phase,
			Result:         status.Result,
			Reference:      cloneReferenceInfo(hosted.reference),
		})
		hosted.mu.Unlock()
	}
	sort.Slice(summaries, func(i, j int) bool { return summaries[i].GameID < summaries[j].GameID })
	return summaries
}

func (h *Host) PlayerSnapshot(gameID string, seatID protocol.SeatID) (PlayerSnapshot, error) {
	hosted, err := h.lookup(gameID)
	if err != nil {
		return PlayerSnapshot{}, err
	}
	hosted.mu.Lock()
	defer hosted.mu.Unlock()
	if !containsSeat(hosted.seats, seatID) {
		return PlayerSnapshot{}, fmt.Errorf("%w: seat %d", ErrNotFound, seatID)
	}
	view, err := hosted.session.PlayerView(seatID)
	if err != nil {
		return PlayerSnapshot{}, fmt.Errorf("project player view: %w", err)
	}
	var decision *session.PlayerDecisionView
	var constructionBuyouts []game.ConstructionBuyoutQuote
	if hosted.immediateResolver != nil {
		quotes, err := hosted.session.ConstructionBuyoutQuotes(seatID, hosted.immediateResolver)
		if err != nil {
			return PlayerSnapshot{}, fmt.Errorf("project construction buyouts: %w", err)
		}
		constructionBuyouts = quotes
		projected, err := hosted.session.DecisionView(seatID, hosted.immediateResolver)
		if err != nil {
			return PlayerSnapshot{}, fmt.Errorf("project player decision view: %w", err)
		}
		decision = &projected
	}
	battles, err := hosted.session.PlayerBattleViews(seatID)
	if err != nil {
		return PlayerSnapshot{}, fmt.Errorf("project player battles: %w", err)
	}
	var planningDraft *PlanningDraft
	if draft, ok := hosted.planningDrafts[seatID]; ok {
		if draft.GameID == view.GameID && draft.Turn == view.Turn && draft.BaseRevision == view.Revision && view.Phase == session.PhasePlanning && !view.Seat.Submitted {
			cloned := clonePlanningDraft(draft)
			planningDraft = &cloned
		} else {
			delete(hosted.planningDrafts, seatID)
		}
	}
	return PlayerSnapshot{SchemaVersion: SchemaVersion, ChangeSequence: hosted.changeSequence, View: view, Decision: decision, Battles: battles, PlanningDraft: planningDraft, Reference: cloneReferenceInfo(hosted.reference), ConstructionBuyouts: constructionBuyouts}, nil
}

func (h *Host) PlanningPreview(gameID string, seatID protocol.SeatID, batch protocol.CommandBatch) (PlanningPreviewSnapshot, error) {
	hosted, err := h.lookup(gameID)
	if err != nil {
		return PlanningPreviewSnapshot{}, err
	}
	hosted.mu.Lock()
	defer hosted.mu.Unlock()
	if !containsSeat(hosted.seats, seatID) {
		return PlanningPreviewSnapshot{}, fmt.Errorf("%w: seat %d", ErrNotFound, seatID)
	}
	if batch.GameID != gameID || batch.SeatID != seatID {
		return PlanningPreviewSnapshot{}, fmt.Errorf("planning preview batch identity does not match route game/seat")
	}
	if hosted.immediateResolver == nil {
		return PlanningPreviewSnapshot{}, fmt.Errorf("planning preview is unavailable without an economy resolver")
	}
	preview, err := hosted.session.PlanningPreview(batch, hosted.immediateResolver)
	if err != nil {
		if errors.Is(err, session.ErrPlanningPreviewRejected) {
			return PlanningPreviewSnapshot{}, fmt.Errorf("%w: %v", ErrSessionRejected, err)
		}
		return PlanningPreviewSnapshot{}, err
	}
	return PlanningPreviewSnapshot{SchemaVersion: SchemaVersion, ChangeSequence: hosted.changeSequence, Preview: preview}, nil
}

func (h *Host) SavePlanningDraft(gameID string, seatID protocol.SeatID, draft PlanningDraft) (PlanningDraftSnapshot, error) {
	hosted, err := h.lookup(gameID)
	if err != nil {
		return PlanningDraftSnapshot{}, err
	}
	hosted.mu.Lock()
	defer hosted.mu.Unlock()
	if !containsSeat(hosted.seats, seatID) {
		return PlanningDraftSnapshot{}, fmt.Errorf("%w: seat %d", ErrNotFound, seatID)
	}
	if draft.GameID != gameID || draft.SeatID != seatID {
		return PlanningDraftSnapshot{}, fmt.Errorf("planning draft identity does not match route game/seat")
	}
	if _, err := draft.commandBatch(); err != nil {
		return PlanningDraftSnapshot{}, fmt.Errorf("%w: %v", ErrSessionRejected, err)
	}
	status := hosted.session.Status()
	if draft.Turn != status.Turn || draft.BaseRevision > status.Revision || status.Phase != session.PhasePlanning {
		return PlanningDraftSnapshot{}, fmt.Errorf("%w: planning draft targets turn/revision %d/%d while session is %d/%d phase %s", ErrSessionRejected, draft.Turn, draft.BaseRevision, status.Turn, status.Revision, status.Phase)
	}
	if draft.BaseRevision != status.Revision {
		draft = clonePlanningDraft(draft)
		draft.BaseRevision = status.Revision
	}
	batch, err := draft.commandBatch()
	if err != nil {
		return PlanningDraftSnapshot{}, fmt.Errorf("%w: %v", ErrSessionRejected, err)
	}
	if existing, ok := hosted.planningDrafts[seatID]; ok {
		if existing.Turn != status.Turn || existing.BaseRevision != status.Revision || existing.GameID != gameID {
			delete(hosted.planningDrafts, seatID)
		} else if draft.DraftRevision <= existing.DraftRevision {
			return PlanningDraftSnapshot{SchemaVersion: SchemaVersion, ChangeSequence: hosted.changeSequence, Draft: clonePlanningDraft(existing)}, nil
		}
	}
	if hosted.immediateResolver == nil {
		return PlanningDraftSnapshot{}, fmt.Errorf("planning draft validation is unavailable without an economy resolver")
	}
	if _, err := hosted.session.PlanningPreview(batch, hosted.immediateResolver); err != nil {
		if errors.Is(err, session.ErrPlanningPreviewRejected) {
			return PlanningDraftSnapshot{}, fmt.Errorf("%w: %v", ErrSessionRejected, err)
		}
		return PlanningDraftSnapshot{}, err
	}
	stored := clonePlanningDraft(draft)
	hosted.planningDrafts[seatID] = stored
	return PlanningDraftSnapshot{SchemaVersion: SchemaVersion, ChangeSequence: hosted.changeSequence, Draft: clonePlanningDraft(stored)}, nil
}

func (h *Host) ObserverSnapshot(gameID string) (ObserverSnapshot, error) {
	hosted, err := h.lookup(gameID)
	if err != nil {
		return ObserverSnapshot{}, err
	}
	hosted.mu.Lock()
	defer hosted.mu.Unlock()
	view, err := hosted.session.ObserverView()
	if err != nil {
		return ObserverSnapshot{}, fmt.Errorf("project observer view: %w", err)
	}
	return ObserverSnapshot{SchemaVersion: SchemaVersion, ChangeSequence: hosted.changeSequence, View: view}, nil
}

func (h *Host) SubmitTurn(gameID string, batch protocol.CommandBatch) (Receipt, error) {
	hosted, err := h.lookup(gameID)
	if err != nil {
		return Receipt{}, err
	}
	if batch.GameID != gameID {
		return Receipt{}, fmt.Errorf("%w: command batch targets game %q, expected %q", ErrSessionRejected, batch.GameID, gameID)
	}
	return hosted.mutate("session", 0, "submission", func() error {
		if err := hosted.session.SubmitTurn(batch); err != nil {
			return err
		}
		delete(hosted.planningDrafts, batch.SeatID)
		return nil
	})
}

func (h *Host) SubmitImmediateCommand(gameID string, seatID protocol.SeatID, baseRevision uint64, command protocol.Command) (Receipt, error) {
	hosted, err := h.lookup(gameID)
	if err != nil {
		return Receipt{}, err
	}
	return hosted.mutate("session", 0, "immediate_command", func() error {
		if game.IsDiplomacyCommand(command.Kind) {
			return hosted.session.ResolveDiplomacyCommand(seatID, baseRevision, command)
		}
		if game.IsInvasionCommand(command.Kind) {
			return hosted.session.ResolveInvasionCommand(seatID, baseRevision, command)
		}
		if game.IsConstructionBuyoutCommand(command.Kind) {
			if hosted.immediateResolver == nil {
				return fmt.Errorf("immediate command resolver is not configured")
			}
			resolvedCommand, draftKey, err := hosted.constructionBuyoutCommandWithDraft(seatID, command)
			if err != nil {
				return err
			}
			if err := hosted.session.ResolveConstructionBuyoutCommand(seatID, baseRevision, resolvedCommand, hosted.immediateResolver); err != nil {
				return err
			}
			if draftKey != "" {
				hosted.removePlanningDraftOrder(seatID, draftKey)
			}
			return nil
		}
		if game.IsMilitaryDesignCommand(command.Kind) {
			if hosted.immediateResolver == nil {
				return fmt.Errorf("immediate command resolver is not configured")
			}
			return hosted.session.ResolveMilitaryDesignCommand(seatID, baseRevision, command, hosted.immediateResolver)
		}
		if game.IsMilitaryDesignVisualCommand(command.Kind) {
			return hosted.session.ResolveMilitaryDesignVisualCommand(seatID, baseRevision, command)
		}
		if hosted.immediateResolver == nil {
			return fmt.Errorf("immediate command resolver is not configured")
		}
		return hosted.session.ResolveColonyBaseCommand(seatID, baseRevision, command, hosted.immediateResolver)
	})
}

func (h *Host) SubmitBattleCommand(gameID string, battleID uint64, seatID protocol.SeatID, command protocol.Command) (Receipt, error) {
	hosted, err := h.lookup(gameID)
	if err != nil {
		return Receipt{}, err
	}
	return hosted.mutate("battle", battleID, "battle_command", func() error {
		return hosted.session.SubmitBattleCommand(battleID, seatID, command)
	})
}

// AdvanceAutomation drives a game whose active empires are all controlled by
// built-in AI through at most one strategic turn. It is intentionally bounded
// so callers/tests retain control over long-running autonomous matches.
func (h *Host) AdvanceAutomation(gameID string) (Receipt, error) {
	hosted, err := h.lookup(gameID)
	if err != nil {
		return Receipt{}, err
	}
	return hosted.mutate("automation", 0, "automation_advanced", func() error {
		return hosted.requireAllActiveBuiltin()
	})
}

func (h *Host) Subscribe(gameID string) (<-chan Notification, func(), error) {
	hosted, err := h.lookup(gameID)
	if err != nil {
		return nil, nil, err
	}
	hosted.mu.Lock()
	id := hosted.nextSubscriberID
	hosted.nextSubscriberID++
	ch := make(chan Notification, 1)
	hosted.subscribers[id] = ch
	hosted.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			hosted.mu.Lock()
			if current, ok := hosted.subscribers[id]; ok {
				delete(hosted.subscribers, id)
				close(current)
			}
			hosted.mu.Unlock()
		})
	}
	return ch, cancel, nil
}

func (h *Host) lookup(gameID string) (*hostedGame, error) {
	if h == nil {
		return nil, fmt.Errorf("host is nil")
	}
	if gameID == "" {
		return nil, fmt.Errorf("%w: empty game ID", ErrNotFound)
	}
	h.mu.RLock()
	hosted := h.games[gameID]
	h.mu.RUnlock()
	if hosted == nil {
		return nil, fmt.Errorf("%w: game %q", ErrNotFound, gameID)
	}
	return hosted, nil
}

func (g *hostedGame) mutate(scope string, battleID uint64, reason string, fn func() error) (Receipt, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	before := g.session.Status()
	if err := fn(); err != nil {
		return Receipt{}, fmt.Errorf("%w: %v", ErrSessionRejected, err)
	}
	driveErr := g.driveToInteractiveBoundary()
	after := g.session.Status()
	g.rebasePlanningDrafts(before, after)
	g.changeSequence++
	if after.Phase == session.PhaseCompleted && before.Phase != session.PhaseCompleted {
		reason = "game_completed"
	} else if after.Turn > before.Turn {
		reason = "turn_advanced"
	}
	notification := Notification{
		SchemaVersion:  SchemaVersion,
		Kind:           "snapshot_invalidated",
		GameID:         after.GameID,
		ChangeSequence: g.changeSequence,
		GameRevision:   after.Revision,
		Scope:          scope,
		BattleID:       battleID,
		Reason:         reason,
	}
	g.publish(notification)
	receipt := Receipt{SchemaVersion: SchemaVersion, GameID: after.GameID, ChangeSequence: g.changeSequence, GameRevision: after.Revision}
	if driveErr != nil {
		return receipt, driveErr
	}
	return receipt, nil
}

func (g *hostedGame) rebasePlanningDrafts(before, after session.Status) {
	if g.immediateResolver == nil || before.GameID != after.GameID || before.Turn != after.Turn || before.Phase != session.PhasePlanning || after.Phase != session.PhasePlanning || before.Revision == after.Revision {
		return
	}
	for seatID, draft := range g.planningDrafts {
		if draft.GameID != after.GameID || draft.Turn != after.Turn || draft.BaseRevision == after.Revision {
			continue
		}
		rebased := clonePlanningDraft(draft)
		rebased.BaseRevision = after.Revision
		batch, err := rebased.commandBatch()
		if err != nil {
			continue
		}
		if _, err := g.session.PlanningPreview(batch, g.immediateResolver); err != nil {
			continue
		}
		g.planningDrafts[seatID] = rebased
	}
}

func (g *hostedGame) driveToInteractiveBoundary() error {
	startingTurn := g.session.Status().Turn
	for step := 0; step < 1024; step++ {
		status := g.session.Status()
		if status.Phase == session.PhasePlanning && status.Turn > startingTurn {
			return nil
		}
		switch status.Phase {
		case session.PhasePlanning:
			progressed, humanPending, err := g.drivePlanningControllers(status)
			if err != nil {
				return err
			}
			if humanPending {
				return nil
			}
			if !progressed {
				return fmt.Errorf("automatic planning made no progress")
			}
		case session.PhaseEncounters:
			progressed, humanPending, err := g.driveEncounterControllers(status)
			if err != nil {
				return err
			}
			if humanPending {
				return nil
			}
			if !progressed {
				return fmt.Errorf("built-in AI cannot progress active encounter")
			}
		case session.PhaseInvasionDecisions:
			progressed, humanPending, err := g.driveInvasionControllers(status)
			if err != nil {
				return err
			}
			if humanPending {
				return nil
			}
			if !progressed {
				return fmt.Errorf("built-in AI cannot progress invasion decision")
			}
		case session.PhaseCompleted:
			return nil
		case session.PhaseStrategicResolution:
			if g.resolver == nil {
				return fmt.Errorf("strategic resolver is not configured")
			}
			if err := g.session.ResolveStrategic(g.resolver); err != nil {
				return fmt.Errorf("drive strategic resolution: %w", err)
			}
		case session.PhasePostResolution:
			progressed, humanPending, err := g.drivePostResolutionControllers(status)
			if err != nil {
				return err
			}
			if humanPending {
				return nil
			}
			if !progressed {
				if err := g.session.CompleteTurn(); err != nil {
					return fmt.Errorf("complete turn: %w", err)
				}
			}
		default:
			return fmt.Errorf("unsupported session phase %q", status.Phase)
		}
	}
	return fmt.Errorf("automatic phase driver exceeded step limit")
}

func (g *hostedGame) drivePlanningControllers(status session.Status) (bool, bool, error) {
	humanPending := false
	for _, seatID := range g.seats {
		seat := g.seatInfo[seatID]
		if statusHasEliminatedEmpire(status, seat) {
			continue
		}
		player, err := g.session.PlayerView(seatID)
		if err != nil {
			return false, false, fmt.Errorf("project planning seat %d: %w", seatID, err)
		}
		if player.Seat.Submitted {
			continue
		}
		if seat.Controller != session.ControllerBuiltinAI {
			humanPending = true
			continue
		}
		view, decision, err := g.builtinDecisionWithView(seatID)
		if err != nil {
			return false, false, err
		}
		switch decision.Kind {
		case ai.ActionImmediate:
			if decision.Command == nil || !game.IsDiplomacyCommand(decision.Command.Kind) {
				return false, false, fmt.Errorf("builtin AI seat %d returned invalid planning immediate action", seatID)
			}
			if err := g.session.ResolveDiplomacyCommand(seatID, view.Revision, *decision.Command); err != nil {
				return false, false, fmt.Errorf("builtin AI seat %d diplomacy: %w", seatID, err)
			}
			return true, false, nil
		case ai.ActionSubmitTurn:
			if decision.Batch == nil {
				return false, false, fmt.Errorf("builtin AI seat %d returned nil turn batch", seatID)
			}
			if err := g.session.SubmitTurn(*decision.Batch); err != nil {
				return false, false, fmt.Errorf("builtin AI seat %d submit turn: %w", seatID, err)
			}
			return true, false, nil
		default:
			return false, false, fmt.Errorf("builtin AI seat %d returned planning action %q", seatID, decision.Kind)
		}
	}
	return false, humanPending, nil
}

func (g *hostedGame) drivePostResolutionControllers(status session.Status) (bool, bool, error) {
	if g.immediateResolver != nil && g.immediateResolver.Rules != nil {
		due, err := g.session.DueResearchEmpireIDs(g.immediateResolver.Rules)
		if err != nil {
			return false, false, fmt.Errorf("project due research: %w", err)
		}
		if len(due) != 0 {
			if err := g.session.CompleteResearchField(due[0], g.immediateResolver); err != nil {
				return false, false, fmt.Errorf("complete due research for empire %d: %w", due[0], err)
			}
			return true, false, nil
		}
	}
	humanPending := false
	for _, seatID := range g.seats {
		seat := g.seatInfo[seatID]
		if statusHasEliminatedEmpire(status, seat) {
			continue
		}
		pending, err := g.session.ColonyBaseResolutions(seatID)
		if err != nil {
			return false, false, fmt.Errorf("project pending immediate decisions for seat %d: %w", seatID, err)
		}
		if len(pending) == 0 {
			continue
		}
		if seat.Controller != session.ControllerBuiltinAI {
			humanPending = true
			continue
		}
		if g.immediateResolver == nil {
			return false, false, fmt.Errorf("builtin AI colony-base resolution requires immediate resolver")
		}
		view, decision, err := g.builtinDecisionWithView(seatID)
		if err != nil {
			return false, false, err
		}
		if decision.Kind != ai.ActionColonyBase || decision.Command == nil {
			return false, false, fmt.Errorf("builtin AI seat %d returned colony-base action %q", seatID, decision.Kind)
		}
		if err := g.session.ResolveColonyBaseCommand(seatID, view.Revision, *decision.Command, g.immediateResolver); err != nil {
			return false, false, fmt.Errorf("builtin AI seat %d colony-base resolution: %w", seatID, err)
		}
		return true, false, nil
	}
	return false, humanPending, nil
}

func (g *hostedGame) driveEncounterControllers(status session.Status) (bool, bool, error) {
	for _, seatID := range g.seats {
		seat := g.seatInfo[seatID]
		if statusHasEliminatedEmpire(status, seat) || seat.Controller != session.ControllerBuiltinAI {
			continue
		}
		_, decision, err := g.builtinDecisionWithView(seatID)
		if err != nil {
			return false, false, err
		}
		if decision.Kind == ai.ActionBattle && decision.Command != nil {
			if err := g.session.SubmitBattleCommand(decision.BattleID, seatID, *decision.Command); err != nil {
				return false, false, fmt.Errorf("builtin AI seat %d battle %d: %w", seatID, decision.BattleID, err)
			}
			return true, false, nil
		}
	}
	for _, seatID := range g.seats {
		seat := g.seatInfo[seatID]
		if statusHasEliminatedEmpire(status, seat) || seat.Controller == session.ControllerBuiltinAI {
			continue
		}
		views, err := g.session.PlayerBattleViews(seatID)
		if err != nil {
			return false, false, fmt.Errorf("project participant battles for seat %d: %w", seatID, err)
		}
		for _, view := range views {
			if view.Phase == battle.PhaseActive {
				return false, true, nil
			}
		}
	}
	return false, false, nil
}

func (g *hostedGame) driveInvasionControllers(status session.Status) (bool, bool, error) {
	for _, seatID := range g.seats {
		seat := g.seatInfo[seatID]
		if statusHasEliminatedEmpire(status, seat) {
			continue
		}
		player, err := g.session.PlayerView(seatID)
		if err != nil {
			return false, false, fmt.Errorf("project invasion seat %d: %w", seatID, err)
		}
		if player.Invasion == nil {
			continue
		}
		if seat.Controller != session.ControllerBuiltinAI {
			return false, true, nil
		}
		view, decision, err := g.builtinDecisionWithView(seatID)
		if err != nil {
			return false, false, err
		}
		if decision.Kind != ai.ActionInvasion || decision.Command == nil {
			return false, false, fmt.Errorf("builtin AI seat %d returned invasion action %q", seatID, decision.Kind)
		}
		if err := g.session.ResolveInvasionCommand(seatID, view.Revision, *decision.Command); err != nil {
			return false, false, fmt.Errorf("builtin AI seat %d invasion: %w", seatID, err)
		}
		return true, false, nil
	}
	return false, false, nil
}

func (g *hostedGame) builtinDecisionWithView(seatID protocol.SeatID) (session.PlayerDecisionView, ai.Action, error) {
	if g.immediateResolver == nil {
		return session.PlayerDecisionView{}, ai.Action{}, fmt.Errorf("builtin AI seat %d requires immediate economy resolver", seatID)
	}
	view, err := g.session.DecisionView(seatID, g.immediateResolver)
	if err != nil {
		return session.PlayerDecisionView{}, ai.Action{}, fmt.Errorf("project builtin AI seat %d decision view: %w", seatID, err)
	}
	decision, err := ai.Plan(view)
	if err != nil {
		return session.PlayerDecisionView{}, ai.Action{}, fmt.Errorf("plan builtin AI seat %d: %w", seatID, err)
	}
	return view, decision, nil
}

func (g *hostedGame) requireAllActiveBuiltin() error {
	status := g.session.Status()
	if status.Phase == session.PhaseCompleted {
		return fmt.Errorf("game %q is completed", status.GameID)
	}
	for _, seatID := range g.seats {
		seat := g.seatInfo[seatID]
		if statusHasEliminatedEmpire(status, seat) {
			continue
		}
		if seat.Controller != session.ControllerBuiltinAI {
			return fmt.Errorf("seat %d controller %q is not builtin_ai", seatID, seat.Controller)
		}
	}
	return nil
}

func statusHasEliminatedEmpire(status session.Status, seat session.Seat) bool {
	for _, eliminated := range status.EliminatedEmpireIDs {
		if eliminated == seat.EmpireID {
			return true
		}
	}
	return false
}

func (g *hostedGame) publish(notification Notification) {
	for _, ch := range g.subscribers {
		select {
		case ch <- notification:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- notification:
			default:
			}
		}
	}
}

func (draft PlanningDraft) commandBatch() (protocol.CommandBatch, error) {
	if draft.SchemaVersion != SchemaVersion {
		return protocol.CommandBatch{}, fmt.Errorf("unsupported planning draft schema version %d", draft.SchemaVersion)
	}
	if draft.GameID == "" || draft.SeatID == 0 || draft.Turn == 0 || draft.BaseRevision == 0 || draft.DraftRevision == 0 {
		return protocol.CommandBatch{}, fmt.Errorf("planning draft requires game_id, seat_id, turn, base_revision and positive draft_revision")
	}
	commands := make([]protocol.Command, len(draft.Orders))
	keys := make(map[string]struct{}, len(draft.Orders))
	for i := range draft.Orders {
		order := draft.Orders[i]
		if order.Key == "" {
			return protocol.CommandBatch{}, fmt.Errorf("orders[%d].key must not be empty", i)
		}
		if _, exists := keys[order.Key]; exists {
			return protocol.CommandBatch{}, fmt.Errorf("orders[%d].key %q is duplicated", i, order.Key)
		}
		keys[order.Key] = struct{}{}
		if order.Kind == "" {
			return protocol.CommandBatch{}, fmt.Errorf("orders[%d].kind must not be empty", i)
		}
		if len(order.Payload) > 0 && !json.Valid(order.Payload) {
			return protocol.CommandBatch{}, fmt.Errorf("orders[%d].payload is invalid JSON", i)
		}
		commands[i] = protocol.Command{SchemaVersion: protocol.CommandSchemaVersion, Sequence: uint32(i + 1), Kind: order.Kind, Payload: append(json.RawMessage(nil), order.Payload...)}
	}
	batch := protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: draft.GameID, SeatID: draft.SeatID, Turn: draft.Turn, BaseRevision: draft.BaseRevision, Commands: commands}
	if err := batch.Validate(); err != nil {
		return protocol.CommandBatch{}, err
	}
	return batch, nil
}

func clonePlanningDraft(draft PlanningDraft) PlanningDraft {
	clone := draft
	clone.Orders = make([]PlanningDraftOrder, len(draft.Orders))
	for i := range draft.Orders {
		clone.Orders[i] = draft.Orders[i]
		clone.Orders[i].Payload = append(json.RawMessage(nil), draft.Orders[i].Payload...)
	}
	return clone
}

func (g *hostedGame) constructionBuyoutCommandWithDraft(seatID protocol.SeatID, command protocol.Command) (protocol.Command, string, error) {
	if command.Kind != game.CommandBuyConstruction {
		return command, "", nil
	}
	var buy game.BuyConstructionPayload
	if err := json.Unmarshal(command.Payload, &buy); err != nil || buy.ColonyID == 0 {
		return command, "", nil
	}
	draft, ok := g.planningDrafts[seatID]
	if !ok {
		return command, "", nil
	}
	key := fmt.Sprintf("construction:%d", buy.ColonyID)
	for _, order := range draft.Orders {
		if order.Key != key {
			continue
		}
		if order.Kind != game.CommandSetConstructionQueue {
			return protocol.Command{}, "", fmt.Errorf("planning draft %q has unexpected kind %q", key, order.Kind)
		}
		var queue game.SetConstructionQueuePayload
		if err := json.Unmarshal(order.Payload, &queue); err != nil {
			return protocol.Command{}, "", fmt.Errorf("decode planning draft %q: %w", key, err)
		}
		if queue.ColonyID != buy.ColonyID || len(queue.Items) == 0 {
			return protocol.Command{}, "", fmt.Errorf("planning draft %q has no buyable current construction", key)
		}
		planned, err := game.NewBuyPlannedConstructionCommand(command.Sequence, game.BuyPlannedConstructionPayload{ColonyID: buy.ColonyID, Items: queue.Items})
		if err != nil {
			return protocol.Command{}, "", err
		}
		return planned, key, nil
	}
	return command, "", nil
}

func (g *hostedGame) removePlanningDraftOrder(seatID protocol.SeatID, key string) {
	draft, ok := g.planningDrafts[seatID]
	if !ok {
		return
	}
	orders := make([]PlanningDraftOrder, 0, len(draft.Orders))
	for _, order := range draft.Orders {
		if order.Key != key {
			orders = append(orders, order)
		}
	}
	if len(orders) == len(draft.Orders) {
		return
	}
	if len(orders) == 0 {
		delete(g.planningDrafts, seatID)
		return
	}
	draft.Orders = orders
	draft.DraftRevision++
	g.planningDrafts[seatID] = draft
}

func containsSeat(seats []protocol.SeatID, seatID protocol.SeatID) bool {
	index := sort.Search(len(seats), func(i int) bool { return seats[i] >= seatID })
	return index < len(seats) && seats[index] == seatID
}
