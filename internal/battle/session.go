package battle

import (
	"fmt"
	"sort"
	"sync"

	"moox/internal/protocol"
)

type Phase string

const (
	PhasePending   Phase = "pending"
	PhaseActive    Phase = "active"
	PhaseCompleted Phase = "completed"
)

type Spec struct {
	ID            uint64            `json:"id"`
	GameID        string            `json:"game_id"`
	StrategicTurn uint64            `json:"strategic_turn"`
	Participants  []protocol.SeatID `json:"participants"`
	Seed          uint64            `json:"seed"`
}

type Result struct {
	WinnerSeats []protocol.SeatID `json:"winner_seats"`
	Outcome     string            `json:"outcome"`
}

type View struct {
	Spec   Spec    `json:"spec"`
	Phase  Phase   `json:"phase"`
	Result *Result `json:"result,omitempty"`
}

type Session struct {
	mu     sync.RWMutex
	spec   Spec
	phase  Phase
	result *Result
}

func NewSession(spec Spec) (*Session, error) {
	if err := validateSpec(spec); err != nil {
		return nil, err
	}
	spec.Participants = append([]protocol.SeatID(nil), spec.Participants...)
	sort.Slice(spec.Participants, func(i, j int) bool { return spec.Participants[i] < spec.Participants[j] })
	return &Session{spec: spec, phase: PhasePending}, nil
}

func validateSpec(spec Spec) error {
	if spec.ID == 0 {
		return fmt.Errorf("battle id must be non-zero")
	}
	if spec.GameID == "" {
		return fmt.Errorf("game id must not be empty")
	}
	if spec.StrategicTurn == 0 {
		return fmt.Errorf("strategic turn must be non-zero")
	}
	if len(spec.Participants) < 2 {
		return fmt.Errorf("battle requires at least two participants")
	}
	seen := map[protocol.SeatID]struct{}{}
	for _, participant := range spec.Participants {
		if participant == 0 {
			return fmt.Errorf("battle participant must be non-zero")
		}
		if _, ok := seen[participant]; ok {
			return fmt.Errorf("duplicate battle participant %d", participant)
		}
		seen[participant] = struct{}{}
	}
	return nil
}

func (s *Session) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhasePending {
		return fmt.Errorf("cannot start battle in phase %q", s.phase)
	}
	s.phase = PhaseActive
	return nil
}

func (s *Session) Complete(result Result) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhaseActive {
		return fmt.Errorf("cannot complete battle in phase %q", s.phase)
	}
	if result.Outcome == "" {
		return fmt.Errorf("battle outcome must not be empty")
	}
	winners := append([]protocol.SeatID(nil), result.WinnerSeats...)
	sort.Slice(winners, func(i, j int) bool { return winners[i] < winners[j] })
	for _, winner := range winners {
		if !containsSeat(s.spec.Participants, winner) {
			return fmt.Errorf("winner seat %d is not a participant", winner)
		}
	}
	result.WinnerSeats = winners
	s.result = &result
	s.phase = PhaseCompleted
	return nil
}

func (s *Session) View() View {
	s.mu.RLock()
	defer s.mu.RUnlock()
	view := View{Spec: s.spec, Phase: s.phase}
	view.Spec.Participants = append([]protocol.SeatID(nil), s.spec.Participants...)
	if s.result != nil {
		result := *s.result
		result.WinnerSeats = append([]protocol.SeatID(nil), s.result.WinnerSeats...)
		view.Result = &result
	}
	return view
}

func containsSeat(seats []protocol.SeatID, target protocol.SeatID) bool {
	for _, seat := range seats {
		if seat == target {
			return true
		}
	}
	return false
}

// DeriveSeed provides a deterministic infrastructure seed for an encounter.
// It is not evidence for, or an emulation of, the original MOO2 combat RNG.
func DeriveSeed(gameSeed, strategicTurn, battleID uint64) uint64 {
	x := gameSeed ^ (strategicTurn * 0x9e3779b97f4a7c15) ^ (battleID * 0xbf58476d1ce4e5b9)
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return x
}
