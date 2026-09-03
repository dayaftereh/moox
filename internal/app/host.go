package app

import (
	"errors"
	"fmt"
	"sort"
	"sync"

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
	SchemaVersion  int           `json:"schema_version"`
	GameID         string        `json:"game_id"`
	ChangeSequence uint64        `json:"change_sequence"`
	Revision       uint64        `json:"revision"`
	Turn           uint64        `json:"turn"`
	Phase          session.Phase `json:"phase"`
}

type PlayerSnapshot struct {
	SchemaVersion  int                `json:"schema_version"`
	ChangeSequence uint64             `json:"change_sequence"`
	View           session.PlayerView `json:"view"`
	Battles        []battle.View      `json:"battles"`
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
	for i := range observer.Seats {
		seats[i] = observer.Seats[i].Seat.ID
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
	battles, err := hosted.session.PlayerBattleViews(seatID)
	if err != nil {
		return PlayerSnapshot{}, fmt.Errorf("project player battles: %w", err)
	}
	return PlayerSnapshot{SchemaVersion: SchemaVersion, ChangeSequence: hosted.changeSequence, View: view, Battles: battles}, nil
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
	if after.Turn > before.Turn {
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
	for step := 0; step < 16; step++ {
		status := g.session.Status()
		switch status.Phase {
		case session.PhasePlanning, session.PhaseEncounters, session.PhaseInvasionDecisions:
			return nil
		case session.PhaseStrategicResolution:
			if g.resolver == nil {
				return fmt.Errorf("strategic resolver is not configured")
			}
			if err := g.session.ResolveStrategic(g.resolver); err != nil {
				return fmt.Errorf("drive strategic resolution: %w", err)
			}
		case session.PhasePostResolution:
			pending, err := g.hasPendingImmediateDecision()
			if err != nil {
				return err
			}
			if pending {
				return nil
			}
			if err := g.session.CompleteTurn(); err != nil {
				return fmt.Errorf("complete turn: %w", err)
			}
		default:
			return fmt.Errorf("unsupported session phase %q", status.Phase)
		}
	}
	return fmt.Errorf("automatic phase driver exceeded step limit")
}

func (g *hostedGame) hasPendingImmediateDecision() (bool, error) {
	for _, seatID := range g.seats {
		pending, err := g.session.ColonyBaseResolutions(seatID)
		if err != nil {
			return false, fmt.Errorf("project pending immediate decisions for seat %d: %w", seatID, err)
		}
		if len(pending) != 0 {
			return true, nil
		}
	}
	return false, nil
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
