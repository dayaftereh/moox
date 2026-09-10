package app

import (
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
	SchemaVersion  int             `json:"schema_version"`
	GameID         string          `json:"game_id"`
	ChangeSequence uint64          `json:"change_sequence"`
	Revision       uint64          `json:"revision"`
	Turn           uint64          `json:"turn"`
	Phase          session.Phase   `json:"phase"`
	Result         *session.Result `json:"result,omitempty"`
}

type PlayerSnapshot struct {
	SchemaVersion  int                         `json:"schema_version"`
	ChangeSequence uint64                      `json:"change_sequence"`
	View           session.PlayerView          `json:"view"`
	Decision       *session.PlayerDecisionView `json:"decision,omitempty"`
	Battles        []battle.View               `json:"battles"`
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
	if hosted.immediateResolver != nil {
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
	return PlayerSnapshot{SchemaVersion: SchemaVersion, ChangeSequence: hosted.changeSequence, View: view, Decision: decision, Battles: battles}, nil
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
		return hosted.session.SubmitTurn(batch)
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

func containsSeat(seats []protocol.SeatID, seatID protocol.SeatID) bool {
	index := sort.Search(len(seats), func(i int) bool { return seats[i] >= seatID })
	return index < len(seats) && seats[index] == seatID
}
