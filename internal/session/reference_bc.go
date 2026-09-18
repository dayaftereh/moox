package session

import (
	"fmt"

	"moox/internal/game"
	"moox/internal/protocol"
)

func (s *GameSession) ResolveReferenceGrantBCCommand(seatID protocol.SeatID, baseRevision uint64, command protocol.Command) error {
	if err := command.Validate(1); err != nil {
		return fmt.Errorf("invalid reference BC command: %w", err)
	}
	if command.Kind != game.CommandReferenceGrantBC {
		return fmt.Errorf("command kind %q is not a reference BC command", command.Kind)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhasePlanning {
		return fmt.Errorf("cannot grant reference BC in phase %q", s.phase)
	}
	if baseRevision != s.revision {
		return fmt.Errorf("reference BC base revision %d, expected %d", baseRevision, s.revision)
	}
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return fmt.Errorf("unknown seat %d", seatID)
	}
	if s.seats[index].submission != nil {
		return fmt.Errorf("reference BC grant is closed after seat %d submitted turn %d", seatID, s.state.Turn)
	}
	if s.empireEliminatedLocked(s.seats[index].seat.EmpireID) {
		return fmt.Errorf("seat %d controls eliminated empire %d", seatID, s.seats[index].seat.EmpireID)
	}
	stateInput, err := cloneState(s.state)
	if err != nil {
		return err
	}
	events, err := game.ResolveReferenceGrantBCCommand(stateInput, s.seats[index].seat.EmpireID, seatID, command)
	if err != nil {
		return fmt.Errorf("resolve reference BC command: %w", err)
	}
	if err := s.commitImmediateStateLocked(stateInput, seatID, command, events, "reference BC"); err != nil {
		return err
	}
	for i := range s.seats {
		if s.seats[i].submission != nil && s.seats[i].submission.Turn == s.state.Turn {
			s.seats[i].submission.BaseRevision = s.revision
		}
	}
	return nil
}
