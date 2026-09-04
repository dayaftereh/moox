package app

import (
	"errors"
	"fmt"
	"sort"

	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

var (
	ErrInvalidSave            = errors.New("invalid save")
	ErrRulesetMismatch        = errors.New("ruleset mismatch")
	ErrGameIDMismatch         = errors.New("game id mismatch")
	ErrPersistenceUnsupported = errors.New("persistence unsupported")
)

func (h *Host) ExportLiveSnapshot(gameID string) ([]byte, error) {
	if h == nil {
		return nil, fmt.Errorf("host is nil")
	}
	hosted, err := h.lookup(gameID)
	if err != nil {
		return nil, err
	}
	hosted.mu.Lock()
	defer hosted.mu.Unlock()
	rules, _, err := persistenceDependencies(hosted.resolver, hosted.immediateResolver)
	if err != nil {
		return nil, err
	}
	data, err := hosted.session.MarshalLiveSnapshot(rules)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidSave, err)
	}
	return data, nil
}

func (h *Host) ImportLiveSnapshot(data []byte) (GameSummary, error) {
	if h == nil {
		return GameSummary{}, fmt.Errorf("host is nil")
	}
	rules, continuation, err := persistenceDependencies(h.newGameResolver, h.newGameImmediateResolver)
	if err != nil {
		return GameSummary{}, err
	}
	restored, err := decodeLiveSnapshot(data, rules, continuation)
	if err != nil {
		return GameSummary{}, err
	}
	status := restored.Status()
	if err := h.Register(Registration{Session: restored, Resolver: h.newGameResolver, ImmediateResolver: h.newGameImmediateResolver}); err != nil {
		return GameSummary{}, err
	}
	return GameSummary{
		SchemaVersion:  SchemaVersion,
		GameID:         status.GameID,
		ChangeSequence: 1,
		Revision:       status.Revision,
		Turn:           status.Turn,
		Phase:          status.Phase,
		Result:         status.Result,
	}, nil
}

func (h *Host) RestoreLiveSnapshot(gameID string, data []byte) (Receipt, error) {
	if h == nil {
		return Receipt{}, fmt.Errorf("host is nil")
	}
	hosted, err := h.lookup(gameID)
	if err != nil {
		return Receipt{}, err
	}

	// Resolver/rules dependencies are immutable for the hosted-game lifetime.
	// Read them before decoding so invalid saves cannot alter hosted state.
	hosted.mu.Lock()
	rules, continuation, depErr := persistenceDependencies(hosted.resolver, hosted.immediateResolver)
	resolver := hosted.resolver
	immediate := hosted.immediateResolver
	hosted.mu.Unlock()
	if depErr != nil {
		return Receipt{}, depErr
	}
	restored, err := decodeLiveSnapshot(data, rules, continuation)
	if err != nil {
		return Receipt{}, err
	}
	status := restored.Status()
	if status.GameID != gameID {
		return Receipt{}, fmt.Errorf("%w: snapshot game %q, target %q", ErrGameIDMismatch, status.GameID, gameID)
	}
	seats, seatInfo, err := persistenceSeatCache(restored)
	if err != nil {
		return Receipt{}, fmt.Errorf("%w: restored seat cache: %v", ErrInvalidSave, err)
	}

	hosted.mu.Lock()
	defer hosted.mu.Unlock()
	// Re-check the dependency identity before the single atomic assignment block.
	currentRules, _, depErr := persistenceDependencies(hosted.resolver, hosted.immediateResolver)
	if depErr != nil {
		return Receipt{}, depErr
	}
	if currentRules.Identity() != rules.Identity() || hosted.resolver != resolver || hosted.immediateResolver != immediate {
		return Receipt{}, fmt.Errorf("%w: hosted persistence dependencies changed during restore", ErrPersistenceUnsupported)
	}
	hosted.session = restored
	hosted.seats = seats
	hosted.seatInfo = seatInfo
	hosted.changeSequence++
	notification := Notification{
		SchemaVersion:  SchemaVersion,
		Kind:           "snapshot_invalidated",
		GameID:         gameID,
		ChangeSequence: hosted.changeSequence,
		GameRevision:   status.Revision,
		Scope:          "session",
		Reason:         "game_loaded",
	}
	hosted.publish(notification)
	return Receipt{
		SchemaVersion:  SchemaVersion,
		GameID:         gameID,
		ChangeSequence: hosted.changeSequence,
		GameRevision:   status.Revision,
	}, nil
}

func decodeLiveSnapshot(data []byte, rules *game.EconomyRules, continuation game.EncounterResolver) (*session.GameSession, error) {
	restored, err := session.UnmarshalLiveSnapshot(data, rules, continuation)
	if err == nil {
		return restored, nil
	}
	if errors.Is(err, session.ErrLiveSnapshotRulesMismatch) {
		return nil, fmt.Errorf("%w: %v", ErrRulesetMismatch, err)
	}
	return nil, fmt.Errorf("%w: %v", ErrInvalidSave, err)
}

func persistenceDependencies(resolver game.Resolver, immediate *game.EconomyResolver) (*game.EconomyRules, game.EncounterResolver, error) {
	economy, ok := resolver.(*game.EconomyResolver)
	if !ok || economy == nil || economy.Rules == nil {
		return nil, nil, fmt.Errorf("%w: live snapshot v1 requires production EconomyResolver", ErrPersistenceUnsupported)
	}
	if immediate == nil || immediate.Rules == nil {
		return nil, nil, fmt.Errorf("%w: live snapshot v1 requires immediate EconomyResolver", ErrPersistenceUnsupported)
	}
	if economy.Rules.Identity() != immediate.Rules.Identity() || economy.Rules.RulesetID == "" || economy.Rules.RulesetSHA256 == "" {
		return nil, nil, fmt.Errorf("%w: strategic/immediate rules identity mismatch", ErrPersistenceUnsupported)
	}
	return economy.Rules, economy, nil
}

func persistenceSeatCache(restored *session.GameSession) ([]protocol.SeatID, map[protocol.SeatID]session.Seat, error) {
	observer, err := restored.ObserverView()
	if err != nil {
		return nil, nil, err
	}
	seats := make([]protocol.SeatID, len(observer.Seats))
	seatInfo := make(map[protocol.SeatID]session.Seat, len(observer.Seats))
	for i := range observer.Seats {
		seat := observer.Seats[i].Seat
		seats[i] = seat.ID
		seatInfo[seat.ID] = seat
	}
	sort.Slice(seats, func(i, j int) bool { return seats[i] < seats[j] })
	return seats, seatInfo, nil
}
