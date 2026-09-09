package session

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

type Phase string

const (
	PhasePlanning            Phase = "planning"
	PhaseStrategicResolution Phase = "strategic_resolution"
	PhaseEncounters          Phase = "encounters"
	PhaseInvasionDecisions   Phase = "invasion_decisions"
	PhasePostResolution      Phase = "post_resolution"
	PhaseCompleted           Phase = "completed"
)

type ControllerType string

const (
	ControllerLocalHuman  ControllerType = "local_human"
	ControllerRemoteHuman ControllerType = "remote_human"
	ControllerBuiltinAI   ControllerType = "builtin_ai"
	ControllerExternalAI  ControllerType = "external_ai"
	ControllerMCPAI       ControllerType = "mcp_ai"
)

type Seat struct {
	ID         protocol.SeatID `json:"id"`
	EmpireID   core.ID         `json:"empire_id"`
	Name       string          `json:"name"`
	Controller ControllerType  `json:"controller"`
}

type seatState struct {
	seat       Seat
	submission *protocol.CommandBatch
}

type SeatView struct {
	Seat      Seat `json:"seat"`
	Submitted bool `json:"submitted"`
}

type Status struct {
	GameID              string    `json:"game_id"`
	Revision            uint64    `json:"revision"`
	Turn                uint64    `json:"turn"`
	Phase               Phase     `json:"phase"`
	EliminatedEmpireIDs []core.ID `json:"eliminated_empire_ids,omitempty"`
	Result              *Result   `json:"result,omitempty"`
}

type ResultKind string

const ResultConquest ResultKind = "conquest"

type Result struct {
	Kind                ResultKind      `json:"kind"`
	WinnerEmpireID      core.ID         `json:"winner_empire_id"`
	WinnerSeatID        protocol.SeatID `json:"winner_seat_id"`
	EliminatedEmpireIDs []core.ID       `json:"eliminated_empire_ids"`
	CompletedTurn       uint64          `json:"completed_turn"`
	CompletedRevision   uint64          `json:"completed_revision"`
}

type DiplomacyView struct {
	OtherEmpireID      core.ID               `json:"other_empire_id"`
	Stance             core.DiplomaticStance `json:"stance"`
	IncomingPeaceOffer bool                  `json:"incoming_peace_offer,omitempty"`
	OutgoingPeaceOffer bool                  `json:"outgoing_peace_offer,omitempty"`
}

type PlayerView struct {
	GameID                string                      `json:"game_id"`
	Revision              uint64                      `json:"revision"`
	Turn                  uint64                      `json:"turn"`
	Phase                 Phase                       `json:"phase"`
	EliminatedEmpireIDs   []core.ID                   `json:"eliminated_empire_ids,omitempty"`
	Result                *Result                     `json:"result,omitempty"`
	Seat                  SeatView                    `json:"seat"`
	Seats                 []SeatView                  `json:"seats"`
	Empire                core.Empire                 `json:"empire"`
	Colonies              []core.Colony               `json:"colonies"`
	Diplomacy             []DiplomacyView             `json:"diplomacy,omitempty"`
	Invasion              *game.InvasionOpportunity   `json:"invasion,omitempty"`
	ColonyBaseResolutions []game.ColonyBaseResolution `json:"colony_base_resolutions,omitempty"`
	RecentResolutions     []ResolutionSummary         `json:"recent_resolutions,omitempty"`
	OwnSubmission         *protocol.CommandBatch      `json:"own_submission,omitempty"`
}

type ObserverSeatView struct {
	Seat       Seat                   `json:"seat"`
	Submitted  bool                   `json:"submitted"`
	Submission *protocol.CommandBatch `json:"submission,omitempty"`
}

type ObserverView struct {
	GameID                string                      `json:"game_id"`
	Revision              uint64                      `json:"revision"`
	Turn                  uint64                      `json:"turn"`
	Phase                 Phase                       `json:"phase"`
	EliminatedEmpireIDs   []core.ID                   `json:"eliminated_empire_ids,omitempty"`
	Result                *Result                     `json:"result,omitempty"`
	State                 *core.GameState             `json:"state"`
	Seats                 []ObserverSeatView          `json:"seats"`
	Events                []protocol.DomainEvent      `json:"events"`
	Battles               []battle.View               `json:"battles"`
	Telemetry             []protocol.DraftTelemetry   `json:"draft_telemetry,omitempty"`
	ColonyBaseResolutions []game.ColonyBaseResolution `json:"colony_base_resolutions,omitempty"`
	Invasion              *game.InvasionOpportunity   `json:"invasion,omitempty"`
}

type EncounterSpec struct {
	SystemID                  core.ID              `json:"system_id,omitempty"`
	Attacker                  game.EncounterSide   `json:"attacker,omitempty"`
	Defender                  game.EncounterSide   `json:"defender,omitempty"`
	DefenderColonyIDs         []core.ID            `json:"defender_colony_ids,omitempty"`
	Participants              []protocol.SeatID    `json:"participants"`
	Tactical                  *battle.TacticalSpec `json:"tactical,omitempty"`
	TacticalUnsupportedReason string               `json:"tactical_unsupported_reason,omitempty"`
}

type GameSession struct {
	mu                      sync.RWMutex
	gameID                  string
	revision                uint64
	phase                   Phase
	state                   *core.GameState
	seats                   []seatState
	events                  []protocol.DomainEvent
	nextEventSequence       uint64
	telemetry               []protocol.DraftTelemetry
	nextTelemetrySeq        uint64
	battles                 []*battle.Session
	nextBattleID            uint64
	encounterResolver       game.EncounterResolver
	encounterContext        game.ResolveContext
	invasion                *game.InvasionOpportunity
	handledInvasions        []game.InvasionHandledKey
	eliminatedEmpires       []core.ID
	result                  *Result
	pendingEliminationCheck bool
}

func NewGameSession(gameID string, state *core.GameState, seats []Seat) (*GameSession, error) {
	if gameID == "" {
		return nil, fmt.Errorf("game id must not be empty")
	}
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	if err := state.Validate(); err != nil {
		return nil, fmt.Errorf("invalid initial game state: %w", err)
	}
	if len(seats) == 0 {
		return nil, fmt.Errorf("game session requires at least one seat")
	}

	ownedState, err := cloneState(state)
	if err != nil {
		return nil, err
	}
	empireIDs := make(map[core.ID]struct{}, len(ownedState.Empires))
	for _, empire := range ownedState.Empires {
		empireIDs[empire.ID] = struct{}{}
	}

	seatCopy := append([]Seat(nil), seats...)
	sort.Slice(seatCopy, func(i, j int) bool { return seatCopy[i].ID < seatCopy[j].ID })
	seenSeats := make(map[protocol.SeatID]struct{}, len(seatCopy))
	seenEmpires := make(map[core.ID]struct{}, len(seatCopy))
	states := make([]seatState, len(seatCopy))
	for i, seat := range seatCopy {
		if seat.ID == 0 {
			return nil, fmt.Errorf("seat[%d] has zero id", i)
		}
		if seat.Name == "" {
			return nil, fmt.Errorf("seat %d has no name", seat.ID)
		}
		if !validController(seat.Controller) {
			return nil, fmt.Errorf("seat %d has unsupported controller %q", seat.ID, seat.Controller)
		}
		if _, ok := seenSeats[seat.ID]; ok {
			return nil, fmt.Errorf("duplicate seat id %d", seat.ID)
		}
		seenSeats[seat.ID] = struct{}{}
		if _, ok := empireIDs[seat.EmpireID]; !ok {
			return nil, fmt.Errorf("seat %d references unknown empire %d", seat.ID, seat.EmpireID)
		}
		if _, ok := seenEmpires[seat.EmpireID]; ok {
			return nil, fmt.Errorf("empire %d is controlled by more than one seat", seat.EmpireID)
		}
		seenEmpires[seat.EmpireID] = struct{}{}
		states[i] = seatState{seat: seat}
	}

	s := &GameSession{
		gameID:            gameID,
		revision:          1,
		phase:             PhasePlanning,
		state:             ownedState,
		seats:             states,
		nextEventSequence: 1,
		nextTelemetrySeq:  1,
		nextBattleID:      1,
	}
	s.appendEventLocked(protocol.EventScopeSession, "session_created", 0, map[string]any{"seat_count": len(states)})
	return s, nil
}

func validController(controller ControllerType) bool {
	switch controller {
	case ControllerLocalHuman, ControllerRemoteHuman, ControllerBuiltinAI, ControllerExternalAI, ControllerMCPAI:
		return true
	default:
		return false
	}
}

func (s *GameSession) SubmitTurn(batch protocol.CommandBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.phase != PhasePlanning {
		return fmt.Errorf("cannot submit turn in phase %q", s.phase)
	}
	if err := batch.Validate(); err != nil {
		return fmt.Errorf("invalid command batch: %w", err)
	}
	if batch.GameID != s.gameID {
		return fmt.Errorf("command batch targets game %q, expected %q", batch.GameID, s.gameID)
	}
	if batch.Turn != s.state.Turn {
		return fmt.Errorf("command batch targets turn %d, expected %d", batch.Turn, s.state.Turn)
	}
	if batch.BaseRevision != s.revision {
		return fmt.Errorf("command batch base revision %d, expected %d", batch.BaseRevision, s.revision)
	}

	index := s.seatIndexLocked(batch.SeatID)
	if index < 0 {
		return fmt.Errorf("unknown seat %d", batch.SeatID)
	}
	if s.empireEliminatedLocked(s.seats[index].seat.EmpireID) {
		return fmt.Errorf("seat %d controls eliminated empire %d", batch.SeatID, s.seats[index].seat.EmpireID)
	}
	if s.seats[index].submission != nil {
		return fmt.Errorf("seat %d already submitted turn %d", batch.SeatID, batch.Turn)
	}
	clone := protocol.CloneCommandBatch(batch)
	s.seats[index].submission = &clone

	if !s.allSubmittedLocked() {
		return nil
	}

	// Flush accepted turn batches in stable seat order. Transport arrival order is
	// intentionally not part of the authoritative event history.
	for i := range s.seats {
		if s.empireEliminatedLocked(s.seats[i].seat.EmpireID) {
			continue
		}
		s.appendEventLocked(protocol.EventScopeStrategic, "turn_submitted", s.seats[i].seat.ID, *s.seats[i].submission)
	}
	s.phase = PhaseStrategicResolution
	s.appendEventLocked(protocol.EventScopeSession, "phase_changed", 0, map[string]any{"phase": s.phase})
	return nil
}

func (s *GameSession) ResolveStrategic(resolver game.Resolver) error {
	if resolver == nil {
		return fmt.Errorf("strategic resolver must not be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhaseStrategicResolution {
		return fmt.Errorf("cannot resolve strategic turn in phase %q", s.phase)
	}

	stateInput, err := cloneState(s.state)
	if err != nil {
		return err
	}
	batches := make([]protocol.CommandBatch, 0, len(s.seats))
	for i := range s.seats {
		if s.empireEliminatedLocked(s.seats[i].seat.EmpireID) {
			continue
		}
		batches = append(batches, protocol.CloneCommandBatch(*s.seats[i].submission))
	}
	ctx := s.resolveContextLocked()
	resolution, err := resolver.Resolve(ctx, stateInput, batches)
	if err != nil {
		return fmt.Errorf("resolve strategic turn: %w", err)
	}
	if err := s.validateStrategicResolutionLocked(resolution); err != nil {
		return err
	}

	needsContinuation := len(resolution.Encounters) != 0 || resolution.Invasion != nil
	var staged game.EncounterResolver
	if needsContinuation {
		var ok bool
		staged, ok = resolver.(game.EncounterResolver)
		if !ok {
			return fmt.Errorf("strategic resolver returned staged continuation without encounter support")
		}
		if resolution.Invasion != nil {
			if _, ok := resolver.(game.InvasionResolver); !ok {
				return fmt.Errorf("strategic resolver returned invasion without invasion continuation support")
			}
		}
	}
	committed, err := cloneState(resolution.State)
	if err != nil {
		return err
	}
	encounters := make([]EncounterSpec, len(resolution.Encounters))
	for i := range resolution.Encounters {
		encounters[i] = encounterSpecFromGame(resolution.Encounters[i])
	}
	preparedBattles, err := s.prepareEncountersLocked(committed, encounters)
	if err != nil {
		return err
	}

	// Everything above is validation/preparation only. From this point onward the
	// strategic boundary can be committed atomically to the authoritative session.
	s.state = committed
	s.revision++
	s.appendResolvedEventsLocked(resolution.Events)
	if needsContinuation {
		s.encounterResolver = staged
		s.encounterContext = cloneResolveContext(ctx)
	} else {
		s.encounterResolver = nil
		s.encounterContext = game.ResolveContext{}
	}
	if resolution.Invasion != nil {
		s.battles = nil
		s.invasion = game.CloneInvasionOpportunity(resolution.Invasion)
		s.phase = PhaseInvasionDecisions
		s.appendEventLocked(protocol.EventScopeSession, "phase_changed", 0, map[string]any{"phase": s.phase})
		return nil
	}
	s.invasion = nil
	s.commitPreparedEncountersLocked(preparedBattles)
	return nil
}

func (s *GameSession) resolveContextLocked() game.ResolveContext {
	ctx := game.ResolveContext{Seats: make([]game.SeatAuthority, 0, len(s.seats)), HandledInvasions: append([]game.InvasionHandledKey(nil), s.handledInvasions...)}
	for i := range s.seats {
		if s.empireEliminatedLocked(s.seats[i].seat.EmpireID) {
			continue
		}
		ctx.Seats = append(ctx.Seats, game.SeatAuthority{SeatID: s.seats[i].seat.ID, EmpireID: s.seats[i].seat.EmpireID})
	}
	return ctx
}

func cloneResolveContext(ctx game.ResolveContext) game.ResolveContext {
	return game.ResolveContext{Seats: append([]game.SeatAuthority(nil), ctx.Seats...), HandledInvasions: append([]game.InvasionHandledKey(nil), ctx.HandledInvasions...)}
}

func encounterSpecFromGame(encounter game.Encounter) EncounterSpec {
	cloned := game.CloneEncounter(encounter)
	return EncounterSpec{
		SystemID:                  cloned.SystemID,
		Attacker:                  cloned.Attacker,
		Defender:                  cloned.Defender,
		DefenderColonyIDs:         cloned.DefenderColonyIDs,
		Participants:              cloned.Participants,
		Tactical:                  cloned.Tactical,
		TacticalUnsupportedReason: cloned.TacticalUnsupportedReason,
	}
}

func (s *GameSession) validateStrategicResolutionLocked(resolution game.Resolution) error {
	if resolution.State == nil {
		return fmt.Errorf("strategic resolver returned nil state")
	}
	if err := resolution.State.Validate(); err != nil {
		return fmt.Errorf("strategic resolver returned invalid state: %w", err)
	}
	if resolution.State.Turn != s.state.Turn {
		return fmt.Errorf("strategic resolver changed turn from %d to %d", s.state.Turn, resolution.State.Turn)
	}
	if resolution.State.Seed != s.state.Seed {
		return fmt.Errorf("strategic resolver changed immutable game seed")
	}
	if len(resolution.Encounters) != 0 && resolution.Invasion != nil {
		return fmt.Errorf("strategic resolver returned encounters and invasion at the same boundary")
	}
	if err := s.validateInvasionOpportunityLocked(resolution.State, resolution.Invasion); err != nil {
		return fmt.Errorf("strategic resolver returned invalid invasion: %w", err)
	}
	for _, event := range resolution.Events {
		if err := s.validateResolvedEventLocked(event); err != nil {
			return err
		}
	}
	return nil
}

func (s *GameSession) appendResolvedEventsLocked(events []game.DomainEvent) {
	for _, event := range events {
		s.events = append(s.events, protocol.DomainEvent{
			SchemaVersion:   protocol.EventSchemaVersion,
			Sequence:        s.nextEventSequence,
			Turn:            s.state.Turn,
			Revision:        s.revision,
			Scope:           protocol.EventScopeStrategic,
			Kind:            event.Kind,
			SeatID:          event.SeatID,
			CommandSequence: event.CommandSequence,
			Data:            append(json.RawMessage(nil), event.Data...),
		})
		s.nextEventSequence++
	}
}
func (s *GameSession) validateResolvedEventLocked(event game.DomainEvent) error {
	if event.Kind == "" {
		return fmt.Errorf("strategic resolver returned event with empty kind")
	}
	if len(event.Data) > 0 && !json.Valid(event.Data) {
		return fmt.Errorf("strategic event %q has invalid JSON data", event.Kind)
	}
	if event.SeatID == 0 {
		if event.CommandSequence != 0 {
			return fmt.Errorf("strategic event %q has command sequence without seat", event.Kind)
		}
		return nil
	}
	index := s.seatIndexLocked(event.SeatID)
	if index < 0 {
		return fmt.Errorf("strategic event %q references unknown seat %d", event.Kind, event.SeatID)
	}
	if event.CommandSequence > uint32(len(s.seats[index].submission.Commands)) {
		return fmt.Errorf("strategic event %q references unknown command %d for seat %d", event.Kind, event.CommandSequence, event.SeatID)
	}
	return nil
}
func (s *GameSession) SubmittedBatches() ([]protocol.CommandBatch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.phase != PhaseStrategicResolution {
		return nil, fmt.Errorf("submitted batches are not ready in phase %q", s.phase)
	}
	batches := make([]protocol.CommandBatch, 0, len(s.seats))
	for i := range s.seats {
		if s.empireEliminatedLocked(s.seats[i].seat.EmpireID) {
			continue
		}
		batches = append(batches, protocol.CloneCommandBatch(*s.seats[i].submission))
	}
	return batches, nil
}

func (s *GameSession) BeginEncounters(specs []EncounterSpec) ([]battle.View, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.beginEncountersLocked(specs)
}

func (s *GameSession) beginEncountersLocked(specs []EncounterSpec) ([]battle.View, error) {
	if s.phase != PhaseStrategicResolution {
		return nil, fmt.Errorf("cannot begin encounters in phase %q", s.phase)
	}
	if len(s.battles) != 0 {
		return nil, fmt.Errorf("encounters already exist for turn %d", s.state.Turn)
	}
	s.encounterResolver = nil
	s.encounterContext = game.ResolveContext{}
	prepared, err := s.prepareEncountersLocked(s.state, specs)
	if err != nil {
		return nil, err
	}
	s.commitPreparedEncountersLocked(prepared)
	return s.battleViewsLocked(), nil
}

func (s *GameSession) prepareEncountersLocked(state *core.GameState, specs []EncounterSpec) ([]*battle.Session, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	if state == nil {
		return nil, fmt.Errorf("encounter preparation state is nil")
	}
	created := make([]*battle.Session, 0, len(specs))
	seenSystems := make(map[core.ID]struct{}, len(specs))
	for i, encounter := range specs {
		if len(encounter.Participants) < 2 {
			return nil, fmt.Errorf("encounter[%d] requires at least two participants", i)
		}
		for _, seatID := range encounter.Participants {
			if s.seatIndexLocked(seatID) < 0 {
				return nil, fmt.Errorf("encounter[%d] references unknown seat %d", i, seatID)
			}
		}
		if encounter.SystemID != 0 {
			if _, duplicate := seenSystems[encounter.SystemID]; duplicate {
				return nil, fmt.Errorf("encounter wave contains multiple battles for system %d", encounter.SystemID)
			}
			seenSystems[encounter.SystemID] = struct{}{}
		}
		if err := validateEncounterTacticalSnapshot(state, encounter); err != nil {
			return nil, fmt.Errorf("encounter[%d]: %w", i, err)
		}
		id := s.nextBattleID + uint64(i)
		spec := battle.Spec{
			ID:                        id,
			GameID:                    s.gameID,
			StrategicTurn:             state.Turn,
			SystemID:                  encounter.SystemID,
			Attacker:                  battleSideFromGame(encounter.Attacker),
			Defender:                  battleSideFromGame(encounter.Defender),
			DefenderColonyIDs:         append([]core.ID(nil), encounter.DefenderColonyIDs...),
			Participants:              append([]protocol.SeatID(nil), encounter.Participants...),
			Seed:                      battle.DeriveSeed(state.Seed, state.Turn, id),
			TacticalUnsupportedReason: encounter.TacticalUnsupportedReason,
		}
		if encounter.Tactical != nil {
			tactical := battle.CloneTacticalSpec(*encounter.Tactical)
			spec.Tactical = &tactical
		}
		child, err := battle.NewSession(spec)
		if err != nil {
			return nil, fmt.Errorf("encounter[%d]: %w", i, err)
		}
		if err := child.Start(); err != nil {
			return nil, fmt.Errorf("start encounter[%d]: %w", i, err)
		}
		created = append(created, child)
	}
	return created, nil
}

func validateEncounterTacticalSnapshot(state *core.GameState, encounter EncounterSpec) error {
	if encounter.Tactical == nil {
		return nil
	}
	for _, tacticalShip := range encounter.Tactical.Ships {
		var strategic *core.Ship
		for i := range state.Ships {
			if state.Ships[i].ID == tacticalShip.ShipID {
				strategic = &state.Ships[i]
				break
			}
		}
		if strategic == nil {
			return fmt.Errorf("tactical Ship %d is absent from the prepared strategic state", tacticalShip.ShipID)
		}
		if strategic.EmpireID != tacticalShip.EmpireID || strategic.Spec.HullID != tacticalShip.HullID || strategic.Spec.WarpDriveID != tacticalShip.WarpDriveID || strategic.Spec.ComputerID != tacticalShip.ComputerID || strategic.Spec.ArmorID != tacticalShip.ArmorID || strategic.Spec.ShieldID != "" || strategic.Spec.FuelCellID != "standard_fuel_cells" {
			return fmt.Errorf("tactical Ship %d snapshot differs from prepared strategic Ship equipment", tacticalShip.ShipID)
		}
		if len(strategic.Spec.Weapons) != len(tacticalShip.Weapons) {
			return fmt.Errorf("tactical Ship %d weapon snapshot differs from prepared strategic Ship", tacticalShip.ShipID)
		}
		for j := range strategic.Spec.Weapons {
			mount := strategic.Spec.Weapons[j]
			weapon := tacticalShip.Weapons[j]
			if mount.Slot != weapon.Slot || mount.WeaponID != weapon.WeaponID || mount.Count != weapon.Count {
				return fmt.Errorf("tactical Ship %d weapon slot %d differs from prepared strategic Ship", tacticalShip.ShipID, weapon.Slot)
			}
		}
	}
	return nil
}

func battleSideFromGame(side game.EncounterSide) battle.Side {
	return battle.Side{
		EmpireID:         side.EmpireID,
		SeatID:           side.SeatID,
		CombatFleetIDs:   append([]core.ID(nil), side.CombatFleetIDs...),
		ShipIDs:          append([]core.ID(nil), side.ShipIDs...),
		CivilianFleetIDs: append([]core.ID(nil), side.CivilianFleetIDs...),
	}
}

func gameSideFromBattle(side battle.Side) game.EncounterSide {
	return game.EncounterSide{
		EmpireID:         side.EmpireID,
		SeatID:           side.SeatID,
		CombatFleetIDs:   append([]core.ID(nil), side.CombatFleetIDs...),
		ShipIDs:          append([]core.ID(nil), side.ShipIDs...),
		CivilianFleetIDs: append([]core.ID(nil), side.CivilianFleetIDs...),
	}
}

func gameEncounterFromBattleSpec(spec battle.Spec) game.Encounter {
	out := game.Encounter{
		SystemID:                  spec.SystemID,
		Attacker:                  gameSideFromBattle(spec.Attacker),
		Defender:                  gameSideFromBattle(spec.Defender),
		DefenderColonyIDs:         append([]core.ID(nil), spec.DefenderColonyIDs...),
		Participants:              append([]protocol.SeatID(nil), spec.Participants...),
		TacticalUnsupportedReason: spec.TacticalUnsupportedReason,
	}
	if spec.Tactical != nil {
		tactical := battle.CloneTacticalSpec(*spec.Tactical)
		out.Tactical = &tactical
	}
	return out
}

func (s *GameSession) commitPreparedEncountersLocked(created []*battle.Session) {
	s.invasion = nil
	if len(created) == 0 {
		s.phase = PhasePostResolution
		s.encounterResolver = nil
		s.encounterContext = game.ResolveContext{}
		s.appendEventLocked(protocol.EventScopeSession, "phase_changed", 0, map[string]any{"phase": s.phase})
		return
	}
	s.battles = created
	s.nextBattleID += uint64(len(created))
	s.phase = PhaseEncounters
	for _, child := range s.battles {
		s.appendEventLocked(protocol.EventScopeBattle, "battle_created", 0, child.View().Spec)
	}
	s.appendEventLocked(protocol.EventScopeSession, "phase_changed", 0, map[string]any{"phase": s.phase})
}

func (s *GameSession) SubmitBattleCommand(battleID uint64, seatID protocol.SeatID, command protocol.Command) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhaseEncounters {
		return fmt.Errorf("cannot submit tactical command in phase %q", s.phase)
	}
	child := s.battleByIDLocked(battleID)
	if child == nil {
		return fmt.Errorf("unknown battle %d", battleID)
	}
	prepared, err := child.PrepareCommand(seatID, command)
	if err != nil {
		return err
	}
	candidate := prepared.Result()
	if candidate == nil {
		return child.CommitPreparedCommand(prepared)
	}
	if s.encounterResolver == nil {
		if err := child.CommitPreparedCommand(prepared); err != nil {
			return err
		}
		if !s.allBattlesCompletedLocked() {
			return nil
		}
		for _, completed := range s.battles {
			s.appendEventLocked(protocol.EventScopeBattle, "battle_completed", 0, completed.View())
		}
		s.phase = PhasePostResolution
		s.appendEventLocked(protocol.EventScopeSession, "phase_changed", 0, map[string]any{"phase": s.phase})
		return nil
	}
	return s.commitStagedBattleCandidateLocked(child, battleID, *candidate, func() error {
		return child.CommitPreparedCommand(prepared)
	})
}

func (s *GameSession) CompleteBattle(battleID uint64, result battle.Result) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhaseEncounters {
		return fmt.Errorf("cannot complete battle in phase %q", s.phase)
	}
	child := s.battleByIDLocked(battleID)
	if child == nil {
		return fmt.Errorf("unknown battle %d", battleID)
	}
	if s.encounterResolver == nil {
		return s.completeLegacyBattleLocked(child, result)
	}
	return s.completeStagedBattleLocked(child, battleID, result)
}

func (s *GameSession) completeLegacyBattleLocked(child *battle.Session, result battle.Result) error {
	if err := child.Complete(result); err != nil {
		return err
	}
	if !s.allBattlesCompletedLocked() {
		return nil
	}
	for _, completed := range s.battles {
		s.appendEventLocked(protocol.EventScopeBattle, "battle_completed", 0, completed.View())
	}
	s.phase = PhasePostResolution
	s.appendEventLocked(protocol.EventScopeSession, "phase_changed", 0, map[string]any{"phase": s.phase})
	return nil
}

func (s *GameSession) completeStagedBattleLocked(child *battle.Session, battleID uint64, result battle.Result) error {
	normalized, err := child.ValidateResult(result)
	if err != nil {
		return err
	}
	return s.commitStagedBattleCandidateLocked(child, battleID, normalized, func() error {
		return child.Complete(normalized)
	})
}

func (s *GameSession) commitStagedBattleCandidateLocked(child *battle.Session, battleID uint64, normalized battle.Result, commitChild func() error) error {
	if commitChild == nil {
		return fmt.Errorf("battle candidate commit callback is nil")
	}
	active := 0
	for _, current := range s.battles {
		if current.View().Phase == battle.PhaseActive {
			active++
		}
	}
	if active > 1 {
		return commitChild()
	}
	if active != 1 {
		return fmt.Errorf("encounter wave has no active final battle")
	}

	outcomes, err := s.encounterOutcomesWithCandidateLocked(battleID, normalized)
	if err != nil {
		return err
	}
	stateInput, err := cloneState(s.state)
	if err != nil {
		return err
	}
	resolution, err := s.encounterResolver.ResumeAfterEncounters(cloneResolveContext(s.encounterContext), stateInput, outcomes)
	if err != nil {
		return fmt.Errorf("resume strategic encounters: %w", err)
	}
	if err := s.validateStrategicResolutionLocked(resolution); err != nil {
		return err
	}
	committed, err := cloneState(resolution.State)
	if err != nil {
		return err
	}
	nextSpecs := make([]EncounterSpec, len(resolution.Encounters))
	for i := range resolution.Encounters {
		nextSpecs[i] = encounterSpecFromGame(resolution.Encounters[i])
	}
	preparedNext, err := s.prepareEncountersLocked(committed, nextSpecs)
	if err != nil {
		return err
	}
	var finalization *conquestFinalization
	if resolution.Invasion == nil && len(preparedNext) == 0 {
		finalization, err = s.prepareConquestFinalizationLocked(committed, s.pendingEliminationCheck)
		if err != nil {
			return err
		}
	}

	// All strategic continuation and next-boundary preparation succeeded on clones.
	// Only now may the final child and authoritative strategic state mutate.
	if err := commitChild(); err != nil {
		return err
	}
	s.state = committed
	s.revision++
	for _, completed := range s.battles {
		s.appendEventLocked(protocol.EventScopeBattle, "battle_completed", 0, completed.View())
	}
	s.appendResolvedEventsLocked(resolution.Events)

	if resolution.Invasion != nil {
		s.battles = nil
		s.invasion = game.CloneInvasionOpportunity(resolution.Invasion)
		s.phase = PhaseInvasionDecisions
		s.appendEventLocked(protocol.EventScopeSession, "phase_changed", 0, map[string]any{"phase": s.phase})
		return nil
	}
	if len(preparedNext) != 0 {
		s.invasion = nil
		s.battles = preparedNext
		s.nextBattleID += uint64(len(preparedNext))
		for _, next := range s.battles {
			s.appendEventLocked(protocol.EventScopeBattle, "battle_created", 0, next.View().Spec)
		}
		// Remain in PhaseEncounters; no transient PostResolution phase is emitted.
		return nil
	}

	s.invasion = nil
	s.encounterResolver = nil
	s.encounterContext = game.ResolveContext{}
	s.commitPostContinuationLocked(finalization)
	return nil
}

func (s *GameSession) encounterOutcomesWithCandidateLocked(candidateBattleID uint64, candidate battle.Result) ([]game.EncounterOutcome, error) {
	ordered := append([]*battle.Session(nil), s.battles...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].View().Spec.ID < ordered[j].View().Spec.ID })
	outcomes := make([]game.EncounterOutcome, 0, len(ordered))
	for _, child := range ordered {
		view := child.View()
		result := view.Result
		if view.Spec.ID == candidateBattleID {
			copyResult := candidate
			result = &copyResult
		}
		if result == nil {
			return nil, fmt.Errorf("battle %d has no completed result for encounter wave", view.Spec.ID)
		}
		outcomes = append(outcomes, game.EncounterOutcome{
			BattleID:         view.Spec.ID,
			Encounter:        gameEncounterFromBattleSpec(view.Spec),
			WinnerSeat:       result.WinnerSeat,
			Outcome:          result.Outcome,
			DestroyedShipIDs: append([]core.ID(nil), result.DestroyedShipIDs...),
		})
	}
	return outcomes, nil
}
func (s *GameSession) ResolveMilitaryDesignVisualCommand(seatID protocol.SeatID, baseRevision uint64, command protocol.Command) error {
	if err := command.Validate(1); err != nil {
		return fmt.Errorf("invalid military design visual command: %w", err)
	}
	if !game.IsMilitaryDesignVisualCommand(command.Kind) {
		return fmt.Errorf("command kind %q is not a military design visual command", command.Kind)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhasePlanning {
		return fmt.Errorf("cannot update military design visual in phase %q", s.phase)
	}
	if baseRevision != s.revision {
		return fmt.Errorf("immediate command base revision %d, expected %d", baseRevision, s.revision)
	}
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return fmt.Errorf("unknown seat %d", seatID)
	}
	if s.empireEliminatedLocked(s.seats[index].seat.EmpireID) {
		return fmt.Errorf("seat %d controls eliminated empire %d", seatID, s.seats[index].seat.EmpireID)
	}
	stateInput, err := cloneState(s.state)
	if err != nil {
		return err
	}
	events, err := game.ResolveMilitaryDesignVisualCommand(stateInput, s.seats[index].seat.EmpireID, seatID, command)
	if err != nil {
		return fmt.Errorf("resolve military design visual command: %w", err)
	}
	if err := s.commitImmediateStateLocked(stateInput, seatID, command, events, "military design visual"); err != nil {
		return err
	}
	// This command changes presentation state only. Submitted gameplay batches
	// remain semantically valid, so move their revision boundary forward with
	// the visual-only authoritative revision instead of invalidating them.
	for i := range s.seats {
		if s.seats[i].submission != nil && s.seats[i].submission.Turn == s.state.Turn {
			s.seats[i].submission.BaseRevision = s.revision
		}
	}
	return nil
}
func (s *GameSession) ResolveColonyBaseCommand(seatID protocol.SeatID, baseRevision uint64, command protocol.Command, resolver *game.EconomyResolver) error {
	if resolver == nil {
		return fmt.Errorf("economy resolver must not be nil")
	}
	if err := command.Validate(1); err != nil {
		return fmt.Errorf("invalid Colony Base command: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhasePostResolution {
		return fmt.Errorf("cannot resolve Colony Base in phase %q", s.phase)
	}
	if baseRevision != s.revision {
		return fmt.Errorf("immediate command base revision %d, expected %d", baseRevision, s.revision)
	}
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return fmt.Errorf("unknown seat %d", seatID)
	}
	if s.empireEliminatedLocked(s.seats[index].seat.EmpireID) {
		return fmt.Errorf("seat %d controls eliminated empire %d", seatID, s.seats[index].seat.EmpireID)
	}
	stateInput, err := cloneState(s.state)
	if err != nil {
		return err
	}
	events, err := resolver.ResolveColonyBaseCommand(stateInput, s.seats[index].seat.EmpireID, seatID, command)
	if err != nil {
		return fmt.Errorf("resolve Colony Base command: %w", err)
	}
	return s.commitImmediateStateLocked(stateInput, seatID, command, events, "Colony Base")
}

func (s *GameSession) ResolveDiplomacyCommand(seatID protocol.SeatID, baseRevision uint64, command protocol.Command) error {
	if err := command.Validate(1); err != nil {
		return fmt.Errorf("invalid diplomacy command: %w", err)
	}
	if !game.IsDiplomacyCommand(command.Kind) {
		return fmt.Errorf("command kind %q is not a diplomacy command", command.Kind)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhasePlanning {
		return fmt.Errorf("cannot resolve diplomacy in phase %q", s.phase)
	}
	if baseRevision != s.revision {
		return fmt.Errorf("immediate command base revision %d, expected %d", baseRevision, s.revision)
	}
	for _, seat := range s.seats {
		if seat.submission != nil {
			return fmt.Errorf("diplomacy is closed after the first turn submission")
		}
	}
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return fmt.Errorf("unknown seat %d", seatID)
	}
	if s.empireEliminatedLocked(s.seats[index].seat.EmpireID) {
		return fmt.Errorf("seat %d controls eliminated empire %d", seatID, s.seats[index].seat.EmpireID)
	}
	stateInput, err := cloneState(s.state)
	if err != nil {
		return err
	}
	events, err := game.ResolveDiplomacyCommand(stateInput, s.seats[index].seat.EmpireID, seatID, command)
	if err != nil {
		return fmt.Errorf("resolve diplomacy command: %w", err)
	}
	return s.commitImmediateStateLocked(stateInput, seatID, command, events, "diplomacy")
}

func (s *GameSession) commitImmediateStateLocked(stateInput *core.GameState, seatID protocol.SeatID, command protocol.Command, events []game.DomainEvent, label string) error {
	if len(events) == 0 {
		return fmt.Errorf("%s command returned no events", label)
	}
	if err := stateInput.Validate(); err != nil {
		return fmt.Errorf("%s command returned invalid state: %w", label, err)
	}
	for _, event := range events {
		if event.Kind == "" {
			return fmt.Errorf("%s command returned event with empty kind", label)
		}
		if len(event.Data) > 0 && !json.Valid(event.Data) {
			return fmt.Errorf("%s event %q has invalid JSON data", label, event.Kind)
		}
		if event.SeatID != seatID || event.CommandSequence != command.Sequence {
			return fmt.Errorf("%s event %q authority=%d/%d want=%d/%d", label, event.Kind, event.SeatID, event.CommandSequence, seatID, command.Sequence)
		}
	}
	committed, err := cloneState(stateInput)
	if err != nil {
		return err
	}
	s.state = committed
	s.revision++
	for _, event := range events {
		s.events = append(s.events, protocol.DomainEvent{
			SchemaVersion: protocol.EventSchemaVersion, Sequence: s.nextEventSequence, Turn: s.state.Turn, Revision: s.revision,
			Scope: protocol.EventScopeStrategic, Kind: event.Kind, SeatID: event.SeatID, CommandSequence: event.CommandSequence,
			Data: append(json.RawMessage(nil), event.Data...),
		})
		s.nextEventSequence++
	}
	return nil
}

func (s *GameSession) ColonyBaseResolutions(seatID protocol.SeatID) ([]game.ColonyBaseResolution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return nil, fmt.Errorf("unknown seat %d", seatID)
	}
	return game.PendingColonyBaseResolutions(s.state, s.seats[index].seat.EmpireID)
}
func (s *GameSession) CompleteResearchField(empireID core.ID, resolver *game.EconomyResolver) error {
	if resolver == nil {
		return fmt.Errorf("economy resolver must not be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhasePostResolution {
		return fmt.Errorf("cannot complete research in phase %q", s.phase)
	}
	if s.empireEliminatedLocked(empireID) {
		return fmt.Errorf("cannot complete research for eliminated empire %d", empireID)
	}

	stateInput, err := cloneState(s.state)
	if err != nil {
		return err
	}
	event, err := resolver.CompleteResearchField(stateInput, empireID)
	if err != nil {
		return fmt.Errorf("complete research field: %w", err)
	}
	if err := stateInput.Validate(); err != nil {
		return fmt.Errorf("research completion returned invalid state: %w", err)
	}
	if err := s.validateResolvedEventLocked(event); err != nil {
		return err
	}
	committed, err := cloneState(stateInput)
	if err != nil {
		return err
	}

	s.state = committed
	s.revision++
	s.events = append(s.events, protocol.DomainEvent{
		SchemaVersion:   protocol.EventSchemaVersion,
		Sequence:        s.nextEventSequence,
		Turn:            s.state.Turn,
		Revision:        s.revision,
		Scope:           protocol.EventScopeStrategic,
		Kind:            event.Kind,
		SeatID:          event.SeatID,
		CommandSequence: event.CommandSequence,
		Data:            append(json.RawMessage(nil), event.Data...),
	})
	s.nextEventSequence++
	return nil
}
func (s *GameSession) GrantTechnology(empireID core.ID, technologyID int, resolver *game.EconomyResolver, options game.TechnologyGrantOptions) error {
	if resolver == nil {
		return fmt.Errorf("economy resolver must not be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhasePostResolution {
		return fmt.Errorf("cannot grant technology in phase %q", s.phase)
	}

	stateInput, err := cloneState(s.state)
	if err != nil {
		return err
	}
	event, err := resolver.GrantTechnology(stateInput, empireID, technologyID, options)
	if err != nil {
		return fmt.Errorf("grant technology: %w", err)
	}
	if err := stateInput.Validate(); err != nil {
		return fmt.Errorf("technology grant returned invalid state: %w", err)
	}
	if err := s.validateResolvedEventLocked(event); err != nil {
		return err
	}
	committed, err := cloneState(stateInput)
	if err != nil {
		return err
	}

	s.state = committed
	s.revision++
	s.events = append(s.events, protocol.DomainEvent{
		SchemaVersion:   protocol.EventSchemaVersion,
		Sequence:        s.nextEventSequence,
		Turn:            s.state.Turn,
		Revision:        s.revision,
		Scope:           protocol.EventScopeStrategic,
		Kind:            event.Kind,
		SeatID:          event.SeatID,
		CommandSequence: event.CommandSequence,
		Data:            append(json.RawMessage(nil), event.Data...),
	})
	s.nextEventSequence++
	return nil
}
func (s *GameSession) CompleteTurn() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhasePostResolution {
		return fmt.Errorf("cannot complete strategic turn in phase %q", s.phase)
	}
	pending, err := game.PendingColonyBaseResolutions(s.state, 0)
	if err != nil {
		return fmt.Errorf("project pending Colony Base resolutions: %w", err)
	}
	if len(pending) != 0 {
		return fmt.Errorf("cannot complete strategic turn with %d pending Colony Base resolution(s)", len(pending))
	}
	s.state.AdvanceTurn()
	s.revision++
	for i := range s.seats {
		s.seats[i].submission = nil
	}
	s.battles = nil
	s.invasion = nil
	s.handledInvasions = nil
	s.pendingEliminationCheck = false
	s.encounterResolver = nil
	s.encounterContext = game.ResolveContext{}
	s.telemetry = nil
	s.phase = PhasePlanning
	s.appendEventLocked(protocol.EventScopeSession, "turn_advanced", 0, map[string]any{"turn": s.state.Turn})
	s.appendEventLocked(protocol.EventScopeSession, "phase_changed", 0, map[string]any{"phase": s.phase})
	return nil
}

func (s *GameSession) PublishDraftTelemetry(seatID protocol.SeatID, kind, summary string, data any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhasePlanning {
		return fmt.Errorf("draft telemetry is only accepted during planning")
	}
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return fmt.Errorf("unknown seat %d", seatID)
	}
	if s.empireEliminatedLocked(s.seats[index].seat.EmpireID) {
		return fmt.Errorf("seat %d controls eliminated empire %d", seatID, s.seats[index].seat.EmpireID)
	}
	if kind == "" {
		return fmt.Errorf("telemetry kind must not be empty")
	}
	var raw json.RawMessage
	if data != nil {
		encoded, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("encode telemetry data: %w", err)
		}
		raw = encoded
	}
	s.telemetry = append(s.telemetry, protocol.DraftTelemetry{
		Sequence: s.nextTelemetrySeq,
		Turn:     s.state.Turn,
		SeatID:   seatID,
		Kind:     kind,
		Summary:  summary,
		Data:     raw,
	})
	s.nextTelemetrySeq++
	return nil
}

func (s *GameSession) ResearchChoices(seatID protocol.SeatID, rules *game.EconomyRules) ([]game.ResearchChoice, error) {
	if rules == nil {
		return nil, fmt.Errorf("economy rules must not be nil")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return nil, fmt.Errorf("unknown seat %d", seatID)
	}
	return rules.AvailableResearchChoices(s.state, s.seats[index].seat.EmpireID)
}
func (s *GameSession) ConstructionChoices(seatID protocol.SeatID, colonyID core.ID, rules *game.EconomyRules) ([]game.ConstructionChoice, error) {
	if rules == nil {
		return nil, fmt.Errorf("economy rules must not be nil")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return nil, fmt.Errorf("unknown seat %d", seatID)
	}
	return rules.AvailableConstructionChoices(s.state, s.seats[index].seat.EmpireID, colonyID)
}
func (s *GameSession) BuildingChoices(seatID protocol.SeatID, colonyID core.ID, rules *game.EconomyRules) ([]game.BuildingChoice, error) {
	if rules == nil {
		return nil, fmt.Errorf("economy rules must not be nil")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return nil, fmt.Errorf("unknown seat %d", seatID)
	}
	return rules.AvailableBuildingChoices(s.state, s.seats[index].seat.EmpireID, colonyID)
}
func (s *GameSession) Status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Status{
		GameID:              s.gameID,
		Revision:            s.revision,
		Turn:                s.state.Turn,
		Phase:               s.phase,
		EliminatedEmpireIDs: append([]core.ID(nil), s.eliminatedEmpires...),
		Result:              cloneResult(s.result),
	}
}

func (s *GameSession) PlayerBattleViews(seatID protocol.SeatID) ([]battle.View, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.seatIndexLocked(seatID) < 0 {
		return nil, fmt.Errorf("unknown seat %d", seatID)
	}
	views := make([]battle.View, 0, len(s.battles))
	for _, child := range s.battles {
		observer := child.View()
		if !containsSeatID(observer.Spec.Participants, seatID) {
			continue
		}
		view, err := child.PlayerView(seatID)
		if err != nil {
			return nil, fmt.Errorf("project battle %d for seat %d: %w", observer.Spec.ID, seatID, err)
		}
		views = append(views, view)
	}
	sort.Slice(views, func(i, j int) bool { return views[i].Spec.ID < views[j].Spec.ID })
	return views, nil
}

func containsSeatID(seats []protocol.SeatID, seatID protocol.SeatID) bool {
	index := sort.Search(len(seats), func(i int) bool { return seats[i] >= seatID })
	return index < len(seats) && seats[index] == seatID
}

func (s *GameSession) PlayerView(seatID protocol.SeatID) (PlayerView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return PlayerView{}, fmt.Errorf("unknown seat %d", seatID)
	}
	seat := s.seats[index]
	view := PlayerView{
		GameID:              s.gameID,
		Revision:            s.revision,
		Turn:                s.state.Turn,
		Phase:               s.phase,
		EliminatedEmpireIDs: append([]core.ID(nil), s.eliminatedEmpires...),
		Result:              cloneResult(s.result),
		Seat:                seatView(seat),
		Seats:               s.seatViewsLocked(),
	}
	for _, empire := range s.state.Empires {
		if empire.ID == seat.seat.EmpireID {
			view.Empire = empire
			break
		}
	}
	for _, colony := range s.state.Colonies {
		if colony.EmpireID == seat.seat.EmpireID {
			view.Colonies = append(view.Colonies, colony)
		}
	}
	for _, other := range s.state.Empires {
		if other.ID == seat.seat.EmpireID {
			continue
		}
		view.Diplomacy = append(view.Diplomacy, DiplomacyView{
			OtherEmpireID:      other.ID,
			Stance:             s.state.DiplomaticStanceBetween(seat.seat.EmpireID, other.ID),
			IncomingPeaceOffer: s.state.HasDiplomaticPeaceOffer(other.ID, seat.seat.EmpireID),
			OutgoingPeaceOffer: s.state.HasDiplomaticPeaceOffer(seat.seat.EmpireID, other.ID),
		})
	}
	resolutions, err := game.PendingColonyBaseResolutions(s.state, seat.seat.EmpireID)
	if err != nil {
		return PlayerView{}, err
	}
	view.ColonyBaseResolutions = resolutions
	if s.invasion != nil && s.invasion.AttackerSeatID == seatID {
		view.Invasion = game.CloneInvasionOpportunity(s.invasion)
	}
	view.RecentResolutions = s.recentResolutionSummariesLocked(seatID, seat.seat.EmpireID)
	if seat.submission != nil {
		clone := protocol.CloneCommandBatch(*seat.submission)
		view.OwnSubmission = &clone
	}
	return view, nil
}

func (s *GameSession) ObserverView() (ObserverView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, err := cloneState(s.state)
	if err != nil {
		return ObserverView{}, err
	}
	view := ObserverView{
		GameID:              s.gameID,
		Revision:            s.revision,
		Turn:                s.state.Turn,
		Phase:               s.phase,
		EliminatedEmpireIDs: append([]core.ID(nil), s.eliminatedEmpires...),
		Result:              cloneResult(s.result),
		State:               state,
		Battles:             s.battleViewsLocked(),
	}
	view.Invasion = game.CloneInvasionOpportunity(s.invasion)
	resolutions, err := game.PendingColonyBaseResolutions(state, 0)
	if err != nil {
		return ObserverView{}, err
	}
	view.ColonyBaseResolutions = resolutions
	view.Seats = make([]ObserverSeatView, len(s.seats))
	for i := range s.seats {
		view.Seats[i] = ObserverSeatView{Seat: s.seats[i].seat, Submitted: s.seats[i].submission != nil}
		if s.seats[i].submission != nil {
			clone := protocol.CloneCommandBatch(*s.seats[i].submission)
			view.Seats[i].Submission = &clone
		}
	}
	view.Events = make([]protocol.DomainEvent, len(s.events))
	for i := range s.events {
		view.Events[i] = protocol.CloneEvent(s.events[i])
	}
	view.Telemetry = make([]protocol.DraftTelemetry, len(s.telemetry))
	for i := range s.telemetry {
		view.Telemetry[i] = protocol.CloneDraftTelemetry(s.telemetry[i])
	}
	return view, nil
}

func (s *GameSession) appendEventLocked(scope protocol.EventScope, kind string, seatID protocol.SeatID, data any) {
	var raw json.RawMessage
	if data != nil {
		raw, _ = json.Marshal(data)
	}
	s.events = append(s.events, protocol.DomainEvent{
		SchemaVersion: protocol.EventSchemaVersion,
		Sequence:      s.nextEventSequence,
		Turn:          s.state.Turn,
		Revision:      s.revision,
		Scope:         scope,
		Kind:          kind,
		SeatID:        seatID,
		Data:          raw,
	})
	s.nextEventSequence++
}

func (s *GameSession) seatIndexLocked(id protocol.SeatID) int {
	index := sort.Search(len(s.seats), func(i int) bool { return s.seats[i].seat.ID >= id })
	if index < len(s.seats) && s.seats[index].seat.ID == id {
		return index
	}
	return -1
}

func (s *GameSession) empireEliminatedLocked(empireID core.ID) bool {
	return containsCoreID(s.eliminatedEmpires, empireID)
}

func (s *GameSession) allSubmittedLocked() bool {
	for i := range s.seats {
		if s.empireEliminatedLocked(s.seats[i].seat.EmpireID) {
			continue
		}
		if s.seats[i].submission == nil {
			return false
		}
	}
	return true
}

func (s *GameSession) allBattlesCompletedLocked() bool {
	for _, child := range s.battles {
		if child.View().Phase != battle.PhaseCompleted {
			return false
		}
	}
	return true
}

func (s *GameSession) battleByIDLocked(id uint64) *battle.Session {
	for _, child := range s.battles {
		if child.View().Spec.ID == id {
			return child
		}
	}
	return nil
}

func (s *GameSession) battleViewsLocked() []battle.View {
	views := make([]battle.View, len(s.battles))
	for i, child := range s.battles {
		views[i] = child.View()
	}
	sort.Slice(views, func(i, j int) bool { return views[i].Spec.ID < views[j].Spec.ID })
	return views
}

func (s *GameSession) seatViewsLocked() []SeatView {
	views := make([]SeatView, len(s.seats))
	for i := range s.seats {
		views[i] = seatView(s.seats[i])
	}
	return views
}

func seatView(seat seatState) SeatView {
	return SeatView{Seat: seat.seat, Submitted: seat.submission != nil}
}

func cloneState(state *core.GameState) (*core.GameState, error) {
	encoded, err := core.MarshalState(state)
	if err != nil {
		return nil, fmt.Errorf("clone game state: %w", err)
	}
	clone, err := core.UnmarshalState(encoded)
	if err != nil {
		return nil, fmt.Errorf("clone game state: %w", err)
	}
	return clone, nil
}
