package session

import (
	"fmt"

	"moox/internal/game"
	"moox/internal/protocol"
)

func (s *GameSession) ConstructionBuyoutQuotes(seatID protocol.SeatID, resolver *game.EconomyResolver) ([]game.ConstructionBuyoutQuote, error) {
	if resolver == nil {
		return nil, fmt.Errorf("economy resolver must not be nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return nil, fmt.Errorf("unknown seat %d", seatID)
	}
	if s.empireEliminatedLocked(s.seats[index].seat.EmpireID) {
		return nil, fmt.Errorf("seat %d controls eliminated empire %d", seatID, s.seats[index].seat.EmpireID)
	}
	return resolver.ConstructionBuyoutQuotes(s.state, s.seats[index].seat.EmpireID)
}

func (s *GameSession) ResolveConstructionBuyoutCommand(seatID protocol.SeatID, baseRevision uint64, command protocol.Command, resolver *game.EconomyResolver) error {
	if resolver == nil {
		return fmt.Errorf("economy resolver must not be nil")
	}
	if err := command.Validate(1); err != nil {
		return fmt.Errorf("invalid construction buyout command: %w", err)
	}
	if !game.IsConstructionBuyoutCommand(command.Kind) {
		return fmt.Errorf("command kind %q is not a construction buyout command", command.Kind)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhasePlanning {
		return fmt.Errorf("cannot buy construction in phase %q", s.phase)
	}
	if baseRevision != s.revision {
		return fmt.Errorf("immediate command base revision %d, expected %d", baseRevision, s.revision)
	}
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return fmt.Errorf("unknown seat %d", seatID)
	}
	if s.seats[index].submission != nil {
		return fmt.Errorf("construction buyout is closed after seat %d submitted turn %d", seatID, s.state.Turn)
	}
	if s.empireEliminatedLocked(s.seats[index].seat.EmpireID) {
		return fmt.Errorf("seat %d controls eliminated empire %d", seatID, s.seats[index].seat.EmpireID)
	}
	stateInput, err := cloneState(s.state)
	if err != nil {
		return err
	}
	events, err := resolver.ResolveConstructionBuyoutCommand(stateInput, s.seats[index].seat.EmpireID, seatID, command)
	if err != nil {
		return fmt.Errorf("resolve construction buyout command: %w", err)
	}
	if err := s.commitImmediateStateLocked(stateInput, seatID, command, events, "construction buyout"); err != nil {
		return err
	}
	for i := range s.seats {
		if s.seats[i].submission != nil && s.seats[i].submission.Turn == s.state.Turn {
			s.seats[i].submission.BaseRevision = s.revision
		}
	}
	return nil
}
