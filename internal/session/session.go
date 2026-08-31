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
	PhasePostResolution      Phase = "post_resolution"
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

type PlayerView struct {
	GameID                string                      `json:"game_id"`
	Revision              uint64                      `json:"revision"`
	Turn                  uint64                      `json:"turn"`
	Phase                 Phase                       `json:"phase"`
	Seat                  SeatView                    `json:"seat"`
	Seats                 []SeatView                  `json:"seats"`
	Empire                core.Empire                 `json:"empire"`
	Colonies              []core.Colony               `json:"colonies"`
	ColonyBaseResolutions []game.ColonyBaseResolution `json:"colony_base_resolutions,omitempty"`
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
	State                 *core.GameState             `json:"state"`
	Seats                 []ObserverSeatView          `json:"seats"`
	Events                []protocol.DomainEvent      `json:"events"`
	Battles               []battle.View               `json:"battles"`
	Telemetry             []protocol.DraftTelemetry   `json:"draft_telemetry,omitempty"`
	ColonyBaseResolutions []game.ColonyBaseResolution `json:"colony_base_resolutions,omitempty"`
}

type EncounterSpec struct {
	Participants []protocol.SeatID `json:"participants"`
}

type GameSession struct {
	mu                sync.RWMutex
	gameID            string
	revision          uint64
	phase             Phase
	state             *core.GameState
	seats             []seatState
	events            []protocol.DomainEvent
	nextEventSequence uint64
	telemetry         []protocol.DraftTelemetry
	nextTelemetrySeq  uint64
	battles           []*battle.Session
	nextBattleID      uint64
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
	batches := make([]protocol.CommandBatch, len(s.seats))
	for i := range s.seats {
		batches[i] = protocol.CloneCommandBatch(*s.seats[i].submission)
	}

	ctx := game.ResolveContext{Seats: make([]game.SeatAuthority, len(s.seats))}
	for i := range s.seats {
		ctx.Seats[i] = game.SeatAuthority{SeatID: s.seats[i].seat.ID, EmpireID: s.seats[i].seat.EmpireID}
	}
	resolution, err := resolver.Resolve(ctx, stateInput, batches)
	if err != nil {
		return fmt.Errorf("resolve strategic turn: %w", err)
	}
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

	committed, err := cloneState(resolution.State)
	if err != nil {
		return err
	}
	for _, event := range resolution.Events {
		if err := s.validateResolvedEventLocked(event); err != nil {
			return err
		}
	}
	encounters := make([]EncounterSpec, len(resolution.Encounters))
	for i := range resolution.Encounters {
		encounters[i] = EncounterSpec{Participants: append([]protocol.SeatID(nil), resolution.Encounters[i].Participants...)}
	}
	preparedBattles, err := s.prepareEncountersLocked(encounters)
	if err != nil {
		return err
	}

	// Everything above is validation/preparation only. From this point onward the
	// resolution can be committed atomically to the authoritative session.
	s.state = committed
	s.revision++
	for _, event := range resolution.Events {
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
	s.commitPreparedEncountersLocked(preparedBattles)
	return nil
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
	batches := make([]protocol.CommandBatch, len(s.seats))
	for i := range s.seats {
		batches[i] = protocol.CloneCommandBatch(*s.seats[i].submission)
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
	prepared, err := s.prepareEncountersLocked(specs)
	if err != nil {
		return nil, err
	}
	s.commitPreparedEncountersLocked(prepared)
	return s.battleViewsLocked(), nil
}

func (s *GameSession) prepareEncountersLocked(specs []EncounterSpec) ([]*battle.Session, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	created := make([]*battle.Session, 0, len(specs))
	for i, encounter := range specs {
		if len(encounter.Participants) < 2 {
			return nil, fmt.Errorf("encounter[%d] requires at least two participants", i)
		}
		for _, seatID := range encounter.Participants {
			if s.seatIndexLocked(seatID) < 0 {
				return nil, fmt.Errorf("encounter[%d] references unknown seat %d", i, seatID)
			}
		}
		id := s.nextBattleID + uint64(i)
		spec := battle.Spec{
			ID:            id,
			GameID:        s.gameID,
			StrategicTurn: s.state.Turn,
			Participants:  append([]protocol.SeatID(nil), encounter.Participants...),
			Seed:          battle.DeriveSeed(s.state.Seed, s.state.Turn, id),
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

func (s *GameSession) commitPreparedEncountersLocked(created []*battle.Session) {
	if len(created) == 0 {
		s.phase = PhasePostResolution
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
	if err := child.Complete(result); err != nil {
		return err
	}
	if !s.allBattlesCompletedLocked() {
		return nil
	}

	// As with turn submissions, authoritative completion events are emitted in
	// battle-ID order, not wall-clock completion order.
	for _, completed := range s.battles {
		view := completed.View()
		s.appendEventLocked(protocol.EventScopeBattle, "battle_completed", 0, view)
	}
	s.phase = PhasePostResolution
	s.appendEventLocked(protocol.EventScopeSession, "phase_changed", 0, map[string]any{"phase": s.phase})
	return nil
}

func (s *GameSession) ResolveColonyBaseCommand(seatID protocol.SeatID, command protocol.Command, resolver *game.EconomyResolver) error {
	if resolver == nil {
		return fmt.Errorf("economy resolver must not be nil")
	}
	if err := command.Validate(command.Sequence); err != nil {
		return fmt.Errorf("invalid Colony Base command: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhasePostResolution {
		return fmt.Errorf("cannot resolve Colony Base in phase %q", s.phase)
	}
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return fmt.Errorf("unknown seat %d", seatID)
	}

	stateInput, err := cloneState(s.state)
	if err != nil {
		return err
	}
	events, err := resolver.ResolveColonyBaseCommand(stateInput, s.seats[index].seat.EmpireID, seatID, command)
	if err != nil {
		return fmt.Errorf("resolve Colony Base command: %w", err)
	}
	if len(events) == 0 {
		return fmt.Errorf("Colony Base command returned no events")
	}
	if err := stateInput.Validate(); err != nil {
		return fmt.Errorf("Colony Base command returned invalid state: %w", err)
	}
	for _, event := range events {
		if event.Kind == "" {
			return fmt.Errorf("Colony Base command returned event with empty kind")
		}
		if len(event.Data) > 0 && !json.Valid(event.Data) {
			return fmt.Errorf("Colony Base event %q has invalid JSON data", event.Kind)
		}
		if event.SeatID != seatID || event.CommandSequence != command.Sequence {
			return fmt.Errorf("Colony Base event %q authority=%d/%d want=%d/%d", event.Kind, event.SeatID, event.CommandSequence, seatID, command.Sequence)
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
	if s.seatIndexLocked(seatID) < 0 {
		return fmt.Errorf("unknown seat %d", seatID)
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
func (s *GameSession) PlayerView(seatID protocol.SeatID) (PlayerView, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return PlayerView{}, fmt.Errorf("unknown seat %d", seatID)
	}
	seat := s.seats[index]
	view := PlayerView{
		GameID:   s.gameID,
		Revision: s.revision,
		Turn:     s.state.Turn,
		Phase:    s.phase,
		Seat:     seatView(seat),
		Seats:    s.seatViewsLocked(),
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
	resolutions, err := game.PendingColonyBaseResolutions(s.state, seat.seat.EmpireID)
	if err != nil {
		return PlayerView{}, err
	}
	view.ColonyBaseResolutions = resolutions
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
		GameID:   s.gameID,
		Revision: s.revision,
		Turn:     s.state.Turn,
		Phase:    s.phase,
		State:    state,
		Battles:  s.battleViewsLocked(),
	}
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

func (s *GameSession) allSubmittedLocked() bool {
	for i := range s.seats {
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
