package session

import (
	"errors"
	"fmt"

	"moox/internal/game"
	"moox/internal/protocol"
)

var ErrMilitaryDesignPreviewRejected = errors.New("military design preview rejected")

type MilitaryDesignPreviewRequest struct {
	BaseRevision uint64                   `json:"base_revision"`
	Draft        game.MilitaryDesignDraft `json:"draft"`
}

type MilitaryDesignPreview struct {
	GameID       string                          `json:"game_id"`
	Turn         uint64                          `json:"turn"`
	BaseRevision uint64                          `json:"base_revision"`
	Projection   game.MilitaryDesignDraftPreview `json:"projection"`
}

func (s *GameSession) MilitaryDesignPreview(seatID protocol.SeatID, request MilitaryDesignPreviewRequest, resolver *game.EconomyResolver) (MilitaryDesignPreview, error) {
	if resolver == nil || resolver.Rules == nil {
		return MilitaryDesignPreview{}, fmt.Errorf("military design preview resolver must not be nil")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.phase != PhasePlanning {
		return MilitaryDesignPreview{}, fmt.Errorf("%w: cannot preview military design in phase %q", ErrMilitaryDesignPreviewRejected, s.phase)
	}
	if request.BaseRevision != s.revision {
		return MilitaryDesignPreview{}, fmt.Errorf("%w: base revision %d, expected %d", ErrMilitaryDesignPreviewRejected, request.BaseRevision, s.revision)
	}
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return MilitaryDesignPreview{}, fmt.Errorf("%w: unknown seat %d", ErrMilitaryDesignPreviewRejected, seatID)
	}
	empireID := s.seats[index].seat.EmpireID
	if s.empireEliminatedLocked(empireID) {
		return MilitaryDesignPreview{}, fmt.Errorf("%w: seat %d controls eliminated empire %d", ErrMilitaryDesignPreviewRejected, seatID, empireID)
	}
	if s.seats[index].submission != nil {
		return MilitaryDesignPreview{}, fmt.Errorf("%w: seat %d already submitted turn %d", ErrMilitaryDesignPreviewRejected, seatID, s.state.Turn)
	}
	var empireIndex = -1
	for i := range s.state.Empires {
		if s.state.Empires[i].ID == empireID {
			empireIndex = i
			break
		}
	}
	if empireIndex < 0 {
		return MilitaryDesignPreview{}, fmt.Errorf("military design preview empire %d is missing", empireID)
	}
	projection := resolver.Rules.PreviewMilitaryDesign(&s.state.Empires[empireIndex], request.Draft)
	return MilitaryDesignPreview{
		GameID: s.gameID, Turn: s.state.Turn, BaseRevision: s.revision, Projection: projection,
	}, nil
}
