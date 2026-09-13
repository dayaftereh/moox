package battle

import (
	"fmt"
	"sort"
	"sync"

	"moox/internal/core"
	"moox/internal/protocol"
)

type Phase string

const (
	PhasePending   Phase = "pending"
	PhaseActive    Phase = "active"
	PhaseCompleted Phase = "completed"
)

type Side struct {
	EmpireID         core.ID         `json:"empire_id"`
	SeatID           protocol.SeatID `json:"seat_id"`
	CombatFleetIDs   []core.ID       `json:"combat_fleet_ids"`
	ShipIDs          []core.ID       `json:"ship_ids"`
	CivilianFleetIDs []core.ID       `json:"civilian_fleet_ids,omitempty"`
}

type Spec struct {
	ID                        uint64            `json:"id"`
	GameID                    string            `json:"game_id"`
	StrategicTurn             uint64            `json:"strategic_turn"`
	SystemID                  core.ID           `json:"system_id,omitempty"`
	Attacker                  Side              `json:"attacker,omitempty"`
	Defender                  Side              `json:"defender,omitempty"`
	DefenderColonyIDs         []core.ID         `json:"defender_colony_ids,omitempty"`
	Participants              []protocol.SeatID `json:"participants"`
	Seed                      uint64            `json:"seed"`
	Tactical                  *TacticalSpec     `json:"tactical,omitempty"`
	TacticalUnsupportedReason string            `json:"tactical_unsupported_reason,omitempty"`
}

type Result struct {
	WinnerSeat       protocol.SeatID   `json:"winner_seat"`
	WinnerSeats      []protocol.SeatID `json:"winner_seats,omitempty"`
	Outcome          string            `json:"outcome"`
	DestroyedShipIDs []core.ID         `json:"destroyed_ship_ids,omitempty"`
}

type View struct {
	Spec     Spec          `json:"spec"`
	Phase    Phase         `json:"phase"`
	Result   *Result       `json:"result,omitempty"`
	Tactical *TacticalView `json:"tactical,omitempty"`
}

type Session struct {
	mu               sync.RWMutex
	spec             Spec
	phase            Phase
	result           *Result
	tactical         *tacticalRuntime
	tacticalRevision uint64
}

func NewSession(spec Spec) (*Session, error) {
	if err := validateSpec(spec); err != nil {
		return nil, err
	}
	spec = cloneSpec(spec)
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

	strategic := spec.SystemID != 0 || spec.Attacker.EmpireID != 0 || spec.Defender.EmpireID != 0
	if !strategic {
		if spec.Tactical != nil || spec.TacticalUnsupportedReason != "" {
			return fmt.Errorf("non-strategic battle cannot carry tactical encounter metadata")
		}
		return nil
	}
	if spec.SystemID == 0 {
		return fmt.Errorf("strategic battle requires non-zero system id")
	}
	if len(spec.Participants) != 2 {
		return fmt.Errorf("strategic battle requires exactly two participants")
	}
	if err := validateSide("attacker", spec.Attacker); err != nil {
		return err
	}
	if err := validateSide("defender", spec.Defender); err != nil {
		return err
	}
	if spec.Attacker.EmpireID == spec.Defender.EmpireID {
		return fmt.Errorf("strategic battle sides cannot share empire %d", spec.Attacker.EmpireID)
	}
	if spec.Attacker.SeatID == spec.Defender.SeatID {
		return fmt.Errorf("strategic battle sides cannot share seat %d", spec.Attacker.SeatID)
	}
	if !containsSeat(spec.Participants, spec.Attacker.SeatID) || !containsSeat(spec.Participants, spec.Defender.SeatID) {
		return fmt.Errorf("strategic battle participants do not match side seats")
	}
	if err := validateSortedUniqueIDs("defender colony ids", spec.DefenderColonyIDs, false); err != nil {
		return err
	}
	if spec.Tactical != nil {
		if err := validateTacticalSpec(spec); err != nil {
			return err
		}
	}
	return nil
}

func validateSide(label string, side Side) error {
	if side.EmpireID == 0 || side.SeatID == 0 {
		return fmt.Errorf("%s side requires non-zero empire and seat", label)
	}
	if len(side.CombatFleetIDs) == 0 || len(side.ShipIDs) == 0 {
		return fmt.Errorf("%s side requires combat fleet and ship ids", label)
	}
	if err := validateSortedUniqueIDs(label+" combat fleet ids", side.CombatFleetIDs, true); err != nil {
		return err
	}
	if err := validateSortedUniqueIDs(label+" ship ids", side.ShipIDs, true); err != nil {
		return err
	}
	if err := validateSortedUniqueIDs(label+" civilian fleet ids", side.CivilianFleetIDs, false); err != nil {
		return err
	}
	return nil
}

func validateSortedUniqueIDs(label string, ids []core.ID, required bool) error {
	if required && len(ids) == 0 {
		return fmt.Errorf("%s must not be empty", label)
	}
	var previous core.ID
	for i, id := range ids {
		if id == 0 {
			return fmt.Errorf("%s contains zero id", label)
		}
		if i > 0 && id <= previous {
			return fmt.Errorf("%s must be strictly ascending", label)
		}
		previous = id
	}
	return nil
}

func (s *Session) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhasePending {
		return fmt.Errorf("cannot start battle in phase %q", s.phase)
	}
	if s.spec.Tactical != nil {
		runtime, err := newTacticalRuntime(*s.spec.Tactical)
		if err != nil {
			return err
		}
		s.tactical = &runtime
	}
	s.phase = PhaseActive
	return nil
}

// ValidateResult normalizes and validates one candidate without mutating the
// battle session. GameSession uses this for the final child of an encounter wave
// so the strategic continuation can be prepared atomically before completion.
func (s *Session) ValidateResult(result Result) (Result, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.phase != PhaseActive {
		return Result{}, fmt.Errorf("cannot complete battle in phase %q", s.phase)
	}
	if s.spec.Tactical != nil {
		return Result{}, fmt.Errorf("tactical-enabled battle results must be produced by tactical commands")
	}
	return normalizeResult(s.spec, result)
}

func (s *Session) Complete(result Result) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhaseActive {
		return fmt.Errorf("cannot complete battle in phase %q", s.phase)
	}
	if s.spec.Tactical != nil {
		return fmt.Errorf("tactical-enabled battle results must be produced by tactical commands")
	}
	normalized, err := normalizeResult(s.spec, result)
	if err != nil {
		return err
	}
	s.result = &normalized
	s.phase = PhaseCompleted
	return nil
}

func normalizeResult(spec Spec, result Result) (Result, error) {
	if result.Outcome == "" {
		return Result{}, fmt.Errorf("battle outcome must not be empty")
	}
	winners := append([]protocol.SeatID(nil), result.WinnerSeats...)
	if result.WinnerSeat != 0 {
		if len(winners) == 0 {
			winners = []protocol.SeatID{result.WinnerSeat}
		} else if len(winners) != 1 || winners[0] != result.WinnerSeat {
			return Result{}, fmt.Errorf("winner seat and winner seats disagree")
		}
	} else {
		if len(winners) != 1 {
			return Result{}, fmt.Errorf("battle result requires exactly one winner")
		}
		result.WinnerSeat = winners[0]
	}
	if len(winners) != 1 {
		return Result{}, fmt.Errorf("battle result requires exactly one winner")
	}
	if !containsSeat(spec.Participants, result.WinnerSeat) {
		return Result{}, fmt.Errorf("winner seat %d is not a participant", result.WinnerSeat)
	}
	result.WinnerSeats = []protocol.SeatID{result.WinnerSeat}

	destroyed := append([]core.ID(nil), result.DestroyedShipIDs...)
	sort.Slice(destroyed, func(i, j int) bool { return destroyed[i] < destroyed[j] })
	for i, id := range destroyed {
		if id == 0 {
			return Result{}, fmt.Errorf("destroyed ship id must be non-zero")
		}
		if i > 0 && id == destroyed[i-1] {
			return Result{}, fmt.Errorf("duplicate destroyed ship id %d", id)
		}
	}
	if len(destroyed) != 0 {
		allowed := make(map[core.ID]struct{}, len(spec.Attacker.ShipIDs)+len(spec.Defender.ShipIDs))
		for _, id := range spec.Attacker.ShipIDs {
			allowed[id] = struct{}{}
		}
		for _, id := range spec.Defender.ShipIDs {
			allowed[id] = struct{}{}
		}
		if len(allowed) == 0 {
			return Result{}, fmt.Errorf("generic battle cannot report destroyed strategic ships")
		}
		for _, id := range destroyed {
			if _, ok := allowed[id]; !ok {
				return Result{}, fmt.Errorf("destroyed ship %d is outside the battle snapshot", id)
			}
		}
	}
	result.DestroyedShipIDs = destroyed
	return result, nil
}

func (s *Session) PlayerView(seatID protocol.SeatID) (View, error) {
	view := s.View()
	if !containsSeat(view.Spec.Participants, seatID) {
		return View{}, fmt.Errorf("seat %d is not a battle participant", seatID)
	}
	if view.Spec.Tactical != nil {
		view.Spec.Tactical.InitialRNGState = 0
	}
	if view.Tactical != nil {
		view.Tactical.State.RNGState = 0
		active := tacticalShipSpec(*view.Spec.Tactical, view.Tactical.State.ActiveShipID)
		if active == nil || active.SeatID != seatID || view.Phase != PhaseActive {
			view.Tactical.LegalMoves = nil
			view.Tactical.LegalFireActions = nil
			view.Tactical.CanEndActivation = false
		}
	}
	return view, nil
}

func (s *Session) View() View {
	s.mu.RLock()
	defer s.mu.RUnlock()
	view := View{Spec: cloneSpec(s.spec), Phase: s.phase}
	if s.result != nil {
		result := *s.result
		result.WinnerSeats = append([]protocol.SeatID(nil), s.result.WinnerSeats...)
		result.DestroyedShipIDs = append([]core.ID(nil), s.result.DestroyedShipIDs...)
		view.Result = &result
	}
	if s.tactical != nil && s.spec.Tactical != nil {
		tactical := buildTacticalView(*s.spec.Tactical, *s.tactical)
		if s.phase != PhaseActive {
			tactical.LegalMoves = nil
			tactical.LegalFireActions = nil
			tactical.CanEndActivation = false
			tactical.CanWaitActivation = false
			tactical.WaitTargetShipIDs = nil
		}
		view.Tactical = &tactical
	}
	return view
}

func cloneSpec(spec Spec) Spec {
	out := spec
	out.Participants = append([]protocol.SeatID(nil), spec.Participants...)
	out.DefenderColonyIDs = append([]core.ID(nil), spec.DefenderColonyIDs...)
	out.Attacker = cloneSide(spec.Attacker)
	out.Defender = cloneSide(spec.Defender)
	if spec.Tactical != nil {
		tactical := cloneTacticalSpec(*spec.Tactical)
		out.Tactical = &tactical
	}
	return out
}

func cloneSide(side Side) Side {
	out := side
	out.CombatFleetIDs = append([]core.ID(nil), side.CombatFleetIDs...)
	out.ShipIDs = append([]core.ID(nil), side.ShipIDs...)
	out.CivilianFleetIDs = append([]core.ID(nil), side.CivilianFleetIDs...)
	return out
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
