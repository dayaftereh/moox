package session

import (
	"fmt"

	"moox/internal/game"
	"moox/internal/protocol"
)

type PlanningPreview struct {
	GameID       string                         `json:"game_id"`
	Turn         uint64                         `json:"turn"`
	BaseRevision uint64                         `json:"base_revision"`
	Projection   game.PlanningPreviewProjection `json:"projection"`
}

// PlanningPreview validates a normal Planning CommandBatch against the current
// session boundary, applies it only to a deep-cloned GameState and returns a
// server-derived projection. It never records a submission, event, telemetry,
// revision or phase change.
func (s *GameSession) PlanningPreview(batch protocol.CommandBatch, resolver *game.EconomyResolver) (PlanningPreview, error) {
	if resolver == nil {
		return PlanningPreview{}, fmt.Errorf("planning preview resolver must not be nil")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.phase != PhasePlanning {
		return PlanningPreview{}, fmt.Errorf("cannot preview turn in phase %q", s.phase)
	}
	if err := batch.Validate(); err != nil {
		return PlanningPreview{}, fmt.Errorf("invalid command batch: %w", err)
	}
	if batch.GameID != s.gameID {
		return PlanningPreview{}, fmt.Errorf("command batch targets game %q, expected %q", batch.GameID, s.gameID)
	}
	if batch.Turn != s.state.Turn {
		return PlanningPreview{}, fmt.Errorf("command batch targets turn %d, expected %d", batch.Turn, s.state.Turn)
	}
	if batch.BaseRevision != s.revision {
		return PlanningPreview{}, fmt.Errorf("command batch base revision %d, expected %d", batch.BaseRevision, s.revision)
	}
	index := s.seatIndexLocked(batch.SeatID)
	if index < 0 {
		return PlanningPreview{}, fmt.Errorf("unknown seat %d", batch.SeatID)
	}
	if s.empireEliminatedLocked(s.seats[index].seat.EmpireID) {
		return PlanningPreview{}, fmt.Errorf("seat %d controls eliminated empire %d", batch.SeatID, s.seats[index].seat.EmpireID)
	}
	if s.seats[index].submission != nil {
		return PlanningPreview{}, fmt.Errorf("seat %d already submitted turn %d", batch.SeatID, batch.Turn)
	}
	clone, err := cloneState(s.state)
	if err != nil {
		return PlanningPreview{}, err
	}
	empireID := s.seats[index].seat.EmpireID
	if err := resolver.ApplyPlanningDraft(clone, empireID, batch.SeatID, batch.Commands); err != nil {
		return PlanningPreview{}, fmt.Errorf("apply planning draft: %w", err)
	}
	projection, err := resolver.BuildPlanningPreview(clone, empireID)
	if err != nil {
		return PlanningPreview{}, fmt.Errorf("build planning preview: %w", err)
	}
	return PlanningPreview{GameID: s.gameID, Turn: s.state.Turn, BaseRevision: s.revision, Projection: projection}, nil
}
