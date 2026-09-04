package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

var ErrLiveSnapshotRulesMismatch = errors.New("live snapshot ruleset mismatch")

const (
	LiveSnapshotSchemaVersion      = 1
	SimulationCompatibilityVersion = 1
)

type liveSeatSnapshot struct {
	Seat       Seat                   `json:"seat"`
	Submission *protocol.CommandBatch `json:"submission,omitempty"`
}

type liveSeatAuthority struct {
	SeatID   protocol.SeatID `json:"seat_id"`
	EmpireID core.ID         `json:"empire_id"`
}

type liveResolveContext struct {
	Seats            []liveSeatAuthority       `json:"seats,omitempty"`
	HandledInvasions []game.InvasionHandledKey `json:"handled_invasions,omitempty"`
}

type liveSnapshot struct {
	SchemaVersion           int                       `json:"schema_version"`
	SimulationCompatVersion int                       `json:"simulation_compat_version"`
	Ruleset                 game.RulesIdentity        `json:"ruleset"`
	GameID                  string                    `json:"game_id"`
	Revision                uint64                    `json:"revision"`
	Phase                   Phase                     `json:"phase"`
	State                   *core.GameState           `json:"state"`
	Seats                   []liveSeatSnapshot        `json:"seats"`
	Events                  []protocol.DomainEvent    `json:"events"`
	NextEventSequence       uint64                    `json:"next_event_sequence"`
	Telemetry               []protocol.DraftTelemetry `json:"telemetry,omitempty"`
	NextTelemetrySequence   uint64                    `json:"next_telemetry_sequence"`
	Battles                 []battle.LiveSnapshot     `json:"battles,omitempty"`
	NextBattleID            uint64                    `json:"next_battle_id"`
	EncounterContext        liveResolveContext        `json:"encounter_context"`
	Invasion                *game.InvasionOpportunity `json:"invasion,omitempty"`
	HandledInvasions        []game.InvasionHandledKey `json:"handled_invasions,omitempty"`
	EliminatedEmpireIDs     []core.ID                 `json:"eliminated_empire_ids,omitempty"`
	Result                  *Result                   `json:"result,omitempty"`
	PendingEliminationCheck bool                      `json:"pending_elimination_check"`
}

func (s *GameSession) MarshalLiveSnapshot(rules *game.EconomyRules) ([]byte, error) {
	if rules == nil {
		return nil, fmt.Errorf("live snapshot requires economy rules")
	}
	identity := rules.Identity()
	if identity.ID == "" || identity.SHA256 == "" {
		return nil, fmt.Errorf("live snapshot requires rules identity")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := s.validateLiveSaveBoundaryLocked(rules); err != nil {
		return nil, err
	}
	state, err := cloneState(s.state)
	if err != nil {
		return nil, err
	}
	snapshot := liveSnapshot{
		SchemaVersion:           LiveSnapshotSchemaVersion,
		SimulationCompatVersion: SimulationCompatibilityVersion,
		Ruleset:                 identity,
		GameID:                  s.gameID,
		Revision:                s.revision,
		Phase:                   s.phase,
		State:                   state,
		Seats:                   make([]liveSeatSnapshot, len(s.seats)),
		Events:                  make([]protocol.DomainEvent, len(s.events)),
		NextEventSequence:       s.nextEventSequence,
		Telemetry:               make([]protocol.DraftTelemetry, len(s.telemetry)),
		NextTelemetrySequence:   s.nextTelemetrySeq,
		NextBattleID:            s.nextBattleID,
		EncounterContext:        makeLiveResolveContext(s.encounterContext),
		Invasion:                game.CloneInvasionOpportunity(s.invasion),
		HandledInvasions:        append([]game.InvasionHandledKey(nil), s.handledInvasions...),
		EliminatedEmpireIDs:     append([]core.ID(nil), s.eliminatedEmpires...),
		Result:                  cloneResult(s.result),
		PendingEliminationCheck: s.pendingEliminationCheck,
	}
	for i := range s.seats {
		if !supportedLiveController(s.seats[i].seat.Controller) {
			return nil, fmt.Errorf("live snapshot v1 does not support controller %q", s.seats[i].seat.Controller)
		}
		snapshot.Seats[i].Seat = s.seats[i].seat
		if s.seats[i].submission != nil {
			cloned := protocol.CloneCommandBatch(*s.seats[i].submission)
			snapshot.Seats[i].Submission = &cloned
		}
	}
	for i := range s.events {
		snapshot.Events[i] = protocol.CloneEvent(s.events[i])
	}
	for i := range s.telemetry {
		snapshot.Telemetry[i] = protocol.CloneDraftTelemetry(s.telemetry[i])
	}
	if len(s.battles) != 0 {
		snapshot.Battles = make([]battle.LiveSnapshot, len(s.battles))
		for i := range s.battles {
			snapshot.Battles[i] = s.battles[i].LiveSnapshot()
		}
		sort.Slice(snapshot.Battles, func(i, j int) bool { return snapshot.Battles[i].Spec.ID < snapshot.Battles[j].Spec.ID })
	}
	return json.Marshal(snapshot)
}

func UnmarshalLiveSnapshot(data []byte, rules *game.EconomyRules, continuation game.EncounterResolver) (*GameSession, error) {
	if rules == nil {
		return nil, fmt.Errorf("restore live snapshot requires economy rules")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var snapshot liveSnapshot
	if err := decoder.Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("decode live session snapshot: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("decode live session snapshot: trailing JSON document")
		}
		return nil, fmt.Errorf("decode live session snapshot trailing data: %w", err)
	}
	if snapshot.SchemaVersion != LiveSnapshotSchemaVersion {
		return nil, fmt.Errorf("live snapshot schema %d, expected %d", snapshot.SchemaVersion, LiveSnapshotSchemaVersion)
	}
	if snapshot.SimulationCompatVersion != SimulationCompatibilityVersion {
		return nil, fmt.Errorf("live snapshot simulation compatibility %d, expected %d", snapshot.SimulationCompatVersion, SimulationCompatibilityVersion)
	}
	if snapshot.Ruleset != rules.Identity() {
		return nil, fmt.Errorf("%w: snapshot %q/%s active %q/%s", ErrLiveSnapshotRulesMismatch, snapshot.Ruleset.ID, snapshot.Ruleset.SHA256, rules.RulesetID, rules.RulesetSHA256)
	}
	if snapshot.State == nil {
		return nil, fmt.Errorf("live snapshot has nil state")
	}
	if err := snapshot.State.Validate(); err != nil {
		return nil, fmt.Errorf("live snapshot state: %w", err)
	}
	if snapshot.GameID == "" || snapshot.Revision == 0 || snapshot.NextEventSequence == 0 || snapshot.NextTelemetrySequence == 0 || snapshot.NextBattleID == 0 {
		return nil, fmt.Errorf("live snapshot has invalid identity/revision/sequence counter")
	}
	if len(snapshot.Seats) == 0 {
		return nil, fmt.Errorf("live snapshot has no seats")
	}
	seats := make([]Seat, len(snapshot.Seats))
	for i := range snapshot.Seats {
		if i > 0 && snapshot.Seats[i-1].Seat.ID >= snapshot.Seats[i].Seat.ID {
			return nil, fmt.Errorf("live snapshot seats must be strictly ascending")
		}
		if !supportedLiveController(snapshot.Seats[i].Seat.Controller) {
			return nil, fmt.Errorf("live snapshot v1 does not support controller %q", snapshot.Seats[i].Seat.Controller)
		}
		seats[i] = snapshot.Seats[i].Seat
	}
	s, err := NewGameSession(snapshot.GameID, snapshot.State, seats)
	if err != nil {
		return nil, fmt.Errorf("restore live session: %w", err)
	}

	s.revision = snapshot.Revision
	s.phase = snapshot.Phase
	s.events = make([]protocol.DomainEvent, len(snapshot.Events))
	var lastEvent uint64
	for i := range snapshot.Events {
		event := protocol.CloneEvent(snapshot.Events[i])
		if event.SchemaVersion != protocol.EventSchemaVersion || event.Sequence == 0 || event.Sequence != lastEvent+1 || event.Kind == "" || event.Turn == 0 || event.Turn > snapshot.State.Turn || event.Revision > snapshot.Revision {
			return nil, fmt.Errorf("live snapshot event[%d] sequence/schema/revision invalid", i)
		}
		if len(event.Data) != 0 && !json.Valid(event.Data) {
			return nil, fmt.Errorf("live snapshot event[%d] data invalid", i)
		}
		lastEvent = event.Sequence
		s.events[i] = event
	}
	if len(snapshot.Events) == 0 || snapshot.NextEventSequence != lastEvent+1 {
		return nil, fmt.Errorf("live snapshot next event sequence %d, expected %d", snapshot.NextEventSequence, lastEvent+1)
	}
	s.nextEventSequence = snapshot.NextEventSequence

	s.telemetry = make([]protocol.DraftTelemetry, len(snapshot.Telemetry))
	var lastTelemetry uint64
	for i := range snapshot.Telemetry {
		event := protocol.CloneDraftTelemetry(snapshot.Telemetry[i])
		if event.Sequence == 0 || (i > 0 && event.Sequence != lastTelemetry+1) || event.Kind == "" || event.Turn != snapshot.State.Turn || s.seatIndexLocked(event.SeatID) < 0 {
			return nil, fmt.Errorf("live snapshot telemetry[%d] invalid", i)
		}
		if len(event.Data) != 0 && !json.Valid(event.Data) {
			return nil, fmt.Errorf("live snapshot telemetry[%d] data invalid", i)
		}
		lastTelemetry = event.Sequence
		s.telemetry[i] = event
	}
	if len(snapshot.Telemetry) != 0 && snapshot.NextTelemetrySequence != lastTelemetry+1 {
		return nil, fmt.Errorf("live snapshot next telemetry sequence %d, expected %d", snapshot.NextTelemetrySequence, lastTelemetry+1)
	}
	if len(snapshot.Telemetry) == 0 && snapshot.NextTelemetrySequence == 0 {
		return nil, fmt.Errorf("live snapshot next telemetry sequence is zero")
	}
	s.nextTelemetrySeq = snapshot.NextTelemetrySequence

	for i := range snapshot.Seats {
		if snapshot.Seats[i].Submission == nil {
			continue
		}
		batch := protocol.CloneCommandBatch(*snapshot.Seats[i].Submission)
		if err := batch.Validate(); err != nil {
			return nil, fmt.Errorf("live snapshot seat %d submission: %w", snapshot.Seats[i].Seat.ID, err)
		}
		if batch.GameID != snapshot.GameID || batch.SeatID != snapshot.Seats[i].Seat.ID || batch.Turn != snapshot.State.Turn || batch.BaseRevision > snapshot.Revision || (snapshot.Phase == PhasePlanning && batch.BaseRevision != snapshot.Revision) {
			return nil, fmt.Errorf("live snapshot seat %d submission authority/boundary invalid", snapshot.Seats[i].Seat.ID)
		}
		index := s.seatIndexLocked(batch.SeatID)
		cloned := protocol.CloneCommandBatch(batch)
		s.seats[index].submission = &cloned
	}

	if !sortedUniqueCoreIDs(snapshot.EliminatedEmpireIDs) {
		return nil, fmt.Errorf("live snapshot eliminated empire IDs must be sorted unique")
	}
	s.eliminatedEmpires = append([]core.ID(nil), snapshot.EliminatedEmpireIDs...)
	s.result = cloneResult(snapshot.Result)
	s.pendingEliminationCheck = snapshot.PendingEliminationCheck
	s.handledInvasions = append([]game.InvasionHandledKey(nil), snapshot.HandledInvasions...)
	if !sortedUniqueInvasionKeys(s.handledInvasions) {
		return nil, fmt.Errorf("live snapshot handled invasion keys must be sorted unique")
	}
	s.encounterContext = snapshot.EncounterContext.toGame()
	if !sortedUniqueInvasionKeys(s.encounterContext.HandledInvasions) {
		return nil, fmt.Errorf("live snapshot encounter handled invasion keys must be sorted unique")
	}
	s.invasion = game.CloneInvasionOpportunity(snapshot.Invasion)

	if len(snapshot.Battles) != 0 {
		s.battles = make([]*battle.Session, len(snapshot.Battles))
		var lastBattleID uint64
		for i := range snapshot.Battles {
			if snapshot.Battles[i].Spec.ID == 0 || snapshot.Battles[i].Spec.ID <= lastBattleID {
				return nil, fmt.Errorf("live snapshot battles must be strictly ascending by ID")
			}
			if snapshot.Battles[i].Spec.GameID != snapshot.GameID || snapshot.Battles[i].Spec.StrategicTurn != snapshot.State.Turn {
				return nil, fmt.Errorf("live snapshot battle %d game/turn mismatch", snapshot.Battles[i].Spec.ID)
			}
			restored, err := battle.RestoreLiveSnapshot(snapshot.Battles[i])
			if err != nil {
				return nil, fmt.Errorf("restore live battle %d: %w", snapshot.Battles[i].Spec.ID, err)
			}
			s.battles[i] = restored
			lastBattleID = snapshot.Battles[i].Spec.ID
		}
		if snapshot.NextBattleID <= lastBattleID {
			return nil, fmt.Errorf("live snapshot next battle ID %d is not above current battle %d", snapshot.NextBattleID, lastBattleID)
		}
	}
	historicalBattleID, err := maxCreatedBattleID(snapshot.Events)
	if err != nil {
		return nil, err
	}
	if snapshot.NextBattleID <= historicalBattleID {
		return nil, fmt.Errorf("live snapshot next battle ID %d is not above historical battle %d", snapshot.NextBattleID, historicalBattleID)
	}
	s.nextBattleID = snapshot.NextBattleID

	if err := s.validateRestoredLiveBoundaryLocked(rules, continuation); err != nil {
		return nil, err
	}
	if snapshot.Phase == PhaseEncounters || snapshot.Phase == PhaseInvasionDecisions {
		s.encounterResolver = continuation
	}
	return s, nil
}

func (s *GameSession) validateLiveSaveBoundaryLocked(rules *game.EconomyRules) error {
	switch s.phase {
	case PhasePlanning:
		if len(s.battles) != 0 || s.invasion != nil || s.encounterResolver != nil || !emptyResolveContext(s.encounterContext) || s.pendingEliminationCheck {
			return fmt.Errorf("planning live snapshot has transient continuation state")
		}
	case PhaseEncounters:
		if len(s.battles) == 0 || s.invasion != nil || s.encounterResolver == nil {
			return fmt.Errorf("encounters live snapshot lacks active battle continuation")
		}
	case PhaseInvasionDecisions:
		if s.invasion == nil || len(s.battles) != 0 || s.encounterResolver == nil {
			return fmt.Errorf("invasion live snapshot lacks invasion continuation")
		}
		if _, ok := s.encounterResolver.(game.InvasionResolver); !ok {
			return fmt.Errorf("invasion live snapshot resolver lacks invasion continuation")
		}
	case PhasePostResolution:
		if len(s.battles) != 0 || s.invasion != nil || s.encounterResolver != nil || !emptyResolveContext(s.encounterContext) || s.pendingEliminationCheck {
			return fmt.Errorf("post-resolution live snapshot has transient continuation state")
		}
		pending, err := game.PendingColonyBaseResolutions(s.state, 0)
		if err != nil {
			return fmt.Errorf("project post-resolution Colony Base decisions: %w", err)
		}
		if len(pending) == 0 {
			return fmt.Errorf("post-resolution live snapshot requires a pending Colony Base decision")
		}
		if due, err := researchDueLocked(s, rules); err != nil {
			return err
		} else if due {
			return fmt.Errorf("post-resolution live snapshot cannot precede due Research completion")
		}
	case PhaseCompleted:
		if s.result == nil || len(s.battles) != 0 || s.invasion != nil || s.encounterResolver != nil || !emptyResolveContext(s.encounterContext) || len(s.handledInvasions) != 0 || s.pendingEliminationCheck {
			return fmt.Errorf("completed live snapshot has invalid continuation/result state")
		}
	default:
		return fmt.Errorf("live snapshot is not supported in phase %q", s.phase)
	}
	return nil
}

func (s *GameSession) validateRestoredLiveBoundaryLocked(rules *game.EconomyRules, continuation game.EncounterResolver) error {
	if err := validateResultBoundary(s); err != nil {
		return err
	}
	if err := validateEncounterContextLocked(s); err != nil {
		return err
	}
	switch s.phase {
	case PhasePlanning:
		if len(s.battles) != 0 || s.invasion != nil || !emptyResolveContext(s.encounterContext) || s.pendingEliminationCheck {
			return fmt.Errorf("restored planning snapshot has transient continuation state")
		}
	case PhaseEncounters:
		if continuation == nil || len(s.battles) == 0 || s.invasion != nil {
			return fmt.Errorf("restored encounters snapshot lacks compatible continuation")
		}
	case PhaseInvasionDecisions:
		if continuation == nil || s.invasion == nil || len(s.battles) != 0 {
			return fmt.Errorf("restored invasion snapshot lacks compatible continuation")
		}
		if _, ok := continuation.(game.InvasionResolver); !ok {
			return fmt.Errorf("restored invasion snapshot continuation lacks invasion support")
		}
	case PhasePostResolution:
		if len(s.battles) != 0 || s.invasion != nil || !emptyResolveContext(s.encounterContext) || s.pendingEliminationCheck {
			return fmt.Errorf("restored post-resolution snapshot has transient continuation state")
		}
		pending, err := game.PendingColonyBaseResolutions(s.state, 0)
		if err != nil || len(pending) == 0 {
			return fmt.Errorf("restored post-resolution snapshot requires pending Colony Base decision")
		}
		if due, err := researchDueLocked(s, rules); err != nil {
			return err
		} else if due {
			return fmt.Errorf("restored post-resolution snapshot precedes due Research completion")
		}
	case PhaseCompleted:
		if len(s.battles) != 0 || s.invasion != nil || !emptyResolveContext(s.encounterContext) || len(s.handledInvasions) != 0 || s.pendingEliminationCheck {
			return fmt.Errorf("restored completed snapshot has transient continuation state")
		}
	default:
		return fmt.Errorf("restored live snapshot phase %q is unsupported", s.phase)
	}
	return nil
}

func validateResultBoundary(s *GameSession) error {
	if s.phase != PhaseCompleted {
		if s.result != nil {
			return fmt.Errorf("non-completed live snapshot carries game result")
		}
		return nil
	}
	if s.result == nil || s.result.Kind != ResultConquest || s.result.WinnerEmpireID == 0 || s.result.WinnerSeatID == 0 {
		return fmt.Errorf("completed live snapshot has invalid conquest result")
	}
	if s.result.CompletedRevision != s.revision || s.result.CompletedTurn != s.state.Turn || !equalCoreIDs(s.result.EliminatedEmpireIDs, s.eliminatedEmpires) {
		return fmt.Errorf("completed live snapshot result boundary disagrees with session")
	}
	if containsCoreID(s.eliminatedEmpires, s.result.WinnerEmpireID) {
		return fmt.Errorf("completed live snapshot winner is eliminated")
	}
	winnerSeat, ok := s.seatForEmpireLocked(s.result.WinnerEmpireID)
	if !ok || winnerSeat != s.result.WinnerSeatID {
		return fmt.Errorf("completed live snapshot winner seat/empire mapping invalid")
	}
	return nil
}

func validateEncounterContextLocked(s *GameSession) error {
	if s.phase != PhaseEncounters && s.phase != PhaseInvasionDecisions {
		if !emptyResolveContext(s.encounterContext) {
			return fmt.Errorf("live snapshot has encounter context outside continuation phase")
		}
		return nil
	}
	if !equalInvasionKeys(s.encounterContext.HandledInvasions, s.handledInvasions) {
		return fmt.Errorf("live snapshot encounter handled invasions disagree with session")
	}
	expected := make([]game.SeatAuthority, 0, len(s.seats))
	for _, seat := range s.seats {
		if s.empireEliminatedLocked(seat.seat.EmpireID) {
			continue
		}
		expected = append(expected, game.SeatAuthority{SeatID: seat.seat.ID, EmpireID: seat.seat.EmpireID})
	}
	if len(expected) != len(s.encounterContext.Seats) {
		return fmt.Errorf("live snapshot encounter seat authority count mismatch")
	}
	for i := range expected {
		if expected[i] != s.encounterContext.Seats[i] {
			return fmt.Errorf("live snapshot encounter seat authority[%d] mismatch", i)
		}
	}
	return nil
}

func researchDueLocked(s *GameSession, rules *game.EconomyRules) (bool, error) {
	for _, seat := range s.seats {
		if s.empireEliminatedLocked(seat.seat.EmpireID) {
			continue
		}
		due, err := rules.ResearchDue(s.state, seat.seat.EmpireID)
		if err != nil {
			return false, fmt.Errorf("project due Research for empire %d: %w", seat.seat.EmpireID, err)
		}
		if due {
			return true, nil
		}
	}
	return false, nil
}

func supportedLiveController(controller ControllerType) bool {
	switch controller {
	case ControllerLocalHuman, ControllerRemoteHuman, ControllerBuiltinAI:
		return true
	default:
		return false
	}
}

func makeLiveResolveContext(ctx game.ResolveContext) liveResolveContext {
	out := liveResolveContext{
		Seats:            make([]liveSeatAuthority, len(ctx.Seats)),
		HandledInvasions: append([]game.InvasionHandledKey(nil), ctx.HandledInvasions...),
	}
	for i := range ctx.Seats {
		out.Seats[i] = liveSeatAuthority{SeatID: ctx.Seats[i].SeatID, EmpireID: ctx.Seats[i].EmpireID}
	}
	return out
}

func (c liveResolveContext) toGame() game.ResolveContext {
	out := game.ResolveContext{
		Seats:            make([]game.SeatAuthority, len(c.Seats)),
		HandledInvasions: append([]game.InvasionHandledKey(nil), c.HandledInvasions...),
	}
	for i := range c.Seats {
		out.Seats[i] = game.SeatAuthority{SeatID: c.Seats[i].SeatID, EmpireID: c.Seats[i].EmpireID}
	}
	return out
}

func emptyResolveContext(ctx game.ResolveContext) bool {
	return len(ctx.Seats) == 0 && len(ctx.HandledInvasions) == 0
}

func sortedUniqueInvasionKeys(keys []game.InvasionHandledKey) bool {
	for i, key := range keys {
		if key.ColonyID == 0 || key.AttackerEmpireID == 0 {
			return false
		}
		if i > 0 {
			previous := keys[i-1]
			if key.ColonyID < previous.ColonyID || (key.ColonyID == previous.ColonyID && key.AttackerEmpireID <= previous.AttackerEmpireID) {
				return false
			}
		}
	}
	return true
}

func equalInvasionKeys(a, b []game.InvasionHandledKey) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func maxCreatedBattleID(events []protocol.DomainEvent) (uint64, error) {
	var maxID uint64
	for i := range events {
		if events[i].Kind != "battle_created" {
			continue
		}
		var spec battle.Spec
		if err := json.Unmarshal(events[i].Data, &spec); err != nil {
			return 0, fmt.Errorf("live snapshot battle_created event[%d]: %w", i, err)
		}
		if spec.ID == 0 {
			return 0, fmt.Errorf("live snapshot battle_created event[%d] has zero battle ID", i)
		}
		if spec.ID > maxID {
			maxID = spec.ID
		}
	}
	return maxID, nil
}
