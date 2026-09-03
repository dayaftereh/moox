package session

import (
	"bytes"
	"encoding/json"
	"fmt"

	"moox/internal/core"
	"moox/internal/protocol"
)

const CompletedSnapshotSchemaVersion = 1

type completedSeatSnapshot struct {
	Seat       Seat                   `json:"seat"`
	Submission *protocol.CommandBatch `json:"submission,omitempty"`
}

type completedSnapshot struct {
	SchemaVersion       int                       `json:"schema_version"`
	GameID              string                    `json:"game_id"`
	Revision            uint64                    `json:"revision"`
	Phase               Phase                     `json:"phase"`
	State               *core.GameState           `json:"state"`
	Seats               []completedSeatSnapshot   `json:"seats"`
	Events              []protocol.DomainEvent    `json:"events"`
	Telemetry           []protocol.DraftTelemetry `json:"telemetry,omitempty"`
	EliminatedEmpireIDs []core.ID                 `json:"eliminated_empire_ids"`
	Result              *Result                   `json:"result"`
}

func (s *GameSession) MarshalCompletedSnapshot() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.phase != PhaseCompleted || (s.result == nil) {
		return nil, fmt.Errorf("completed snapshot requires completed session, phase=%q", s.phase)
	}
	state, err := cloneState(s.state)
	if err != nil {
		return nil, err
	}
	snapshot := completedSnapshot{
		SchemaVersion:       CompletedSnapshotSchemaVersion,
		GameID:              s.gameID,
		Revision:            s.revision,
		Phase:               s.phase,
		State:               state,
		Seats:               make([]completedSeatSnapshot, len(s.seats)),
		Events:              make([]protocol.DomainEvent, len(s.events)),
		Telemetry:           make([]protocol.DraftTelemetry, len(s.telemetry)),
		EliminatedEmpireIDs: append([]core.ID(nil), s.eliminatedEmpires...),
		Result:              cloneResult(s.result),
	}
	for i := range s.seats {
		snapshot.Seats[i].Seat = s.seats[i].seat
		if s.seats[i].submission != nil {
			clone := protocol.CloneCommandBatch(*s.seats[i].submission)
			snapshot.Seats[i].Submission = &clone
		}
	}
	for i := range s.events {
		snapshot.Events[i] = protocol.CloneEvent(s.events[i])
	}
	for i := range s.telemetry {
		snapshot.Telemetry[i] = protocol.CloneDraftTelemetry(s.telemetry[i])
	}
	return json.Marshal(snapshot)
}

func UnmarshalCompletedSnapshot(data []byte) (*GameSession, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var snapshot completedSnapshot
	if err := decoder.Decode(&snapshot); err != nil {
		return nil, fmt.Errorf("decode completed session snapshot: %w", err)
	}
	if snapshot.SchemaVersion != CompletedSnapshotSchemaVersion {
		return nil, fmt.Errorf("completed snapshot schema %d, expected %d", snapshot.SchemaVersion, CompletedSnapshotSchemaVersion)
	}
	if snapshot.Phase != PhaseCompleted || snapshot.Result == nil {
		return nil, fmt.Errorf("completed snapshot must contain completed phase/result")
	}
	if snapshot.State == nil {
		return nil, fmt.Errorf("completed snapshot has nil state")
	}
	if err := snapshot.State.Validate(); err != nil {
		return nil, fmt.Errorf("completed snapshot state: %w", err)
	}
	if snapshot.Revision == 0 || snapshot.Result.CompletedRevision != snapshot.Revision || snapshot.Result.CompletedTurn != snapshot.State.Turn {
		return nil, fmt.Errorf("completed snapshot result boundary does not match state/revision")
	}
	if snapshot.Result.Kind != ResultConquest || snapshot.Result.WinnerEmpireID == 0 || snapshot.Result.WinnerSeatID == 0 {
		return nil, fmt.Errorf("completed snapshot has invalid conquest result")
	}
	if !sortedUniqueCoreIDs(snapshot.EliminatedEmpireIDs) || !sortedUniqueCoreIDs(snapshot.Result.EliminatedEmpireIDs) {
		return nil, fmt.Errorf("completed snapshot eliminated empire IDs must be sorted unique")
	}
	if !equalCoreIDs(snapshot.EliminatedEmpireIDs, snapshot.Result.EliminatedEmpireIDs) {
		return nil, fmt.Errorf("completed snapshot eliminated empire IDs disagree with result")
	}
	if containsCoreID(snapshot.EliminatedEmpireIDs, snapshot.Result.WinnerEmpireID) {
		return nil, fmt.Errorf("completed snapshot winner empire is eliminated")
	}

	seats := make([]Seat, len(snapshot.Seats))
	for i := range snapshot.Seats {
		seats[i] = snapshot.Seats[i].Seat
	}
	s, err := NewGameSession(snapshot.GameID, snapshot.State, seats)
	if err != nil {
		return nil, fmt.Errorf("restore completed session: %w", err)
	}
	winnerSeatID, ok := s.seatForEmpireLocked(snapshot.Result.WinnerEmpireID)
	if !ok || winnerSeatID != snapshot.Result.WinnerSeatID {
		return nil, fmt.Errorf("completed snapshot winner seat/empire mapping is invalid")
	}

	s.revision = snapshot.Revision
	s.phase = PhaseCompleted
	s.eliminatedEmpires = append([]core.ID(nil), snapshot.EliminatedEmpireIDs...)
	s.result = cloneResult(snapshot.Result)
	s.events = make([]protocol.DomainEvent, len(snapshot.Events))
	var maxEvent uint64
	for i := range snapshot.Events {
		event := protocol.CloneEvent(snapshot.Events[i])
		if event.SchemaVersion != protocol.EventSchemaVersion || event.Sequence == 0 || event.Sequence <= maxEvent {
			return nil, fmt.Errorf("completed snapshot event[%d] sequence/schema invalid", i)
		}
		maxEvent = event.Sequence
		s.events[i] = event
	}
	s.nextEventSequence = maxEvent + 1
	if s.nextEventSequence == 1 {
		s.nextEventSequence = 1
	}
	s.telemetry = make([]protocol.DraftTelemetry, len(snapshot.Telemetry))
	var maxTelemetry uint64
	for i := range snapshot.Telemetry {
		event := protocol.CloneDraftTelemetry(snapshot.Telemetry[i])
		if event.Sequence == 0 || event.Sequence <= maxTelemetry {
			return nil, fmt.Errorf("completed snapshot telemetry[%d] sequence invalid", i)
		}
		maxTelemetry = event.Sequence
		s.telemetry[i] = event
	}
	s.nextTelemetrySeq = maxTelemetry + 1
	if s.nextTelemetrySeq == 1 {
		s.nextTelemetrySeq = 1
	}
	for i := range snapshot.Seats {
		if snapshot.Seats[i].Submission == nil {
			continue
		}
		index := s.seatIndexLocked(snapshot.Seats[i].Seat.ID)
		if index < 0 {
			return nil, fmt.Errorf("completed snapshot submission references unknown seat %d", snapshot.Seats[i].Seat.ID)
		}
		clone := protocol.CloneCommandBatch(*snapshot.Seats[i].Submission)
		s.seats[index].submission = &clone
	}
	return s, nil
}

func sortedUniqueCoreIDs(ids []core.ID) bool {
	for i, id := range ids {
		if id == 0 || (i > 0 && id <= ids[i-1]) {
			return false
		}
	}
	return true
}

func equalCoreIDs(a, b []core.ID) bool {
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
