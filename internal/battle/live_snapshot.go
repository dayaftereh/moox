package battle

import (
	"encoding/json"
	"fmt"
	"sort"

	"moox/internal/core"
	"moox/internal/protocol"
)

const LiveSnapshotSchemaVersion = 1

type TacticalLiveSnapshot struct {
	State             TacticalState   `json:"state"`
	Events            []TacticalEvent `json:"events"`
	NextEventSequence uint64          `json:"next_event_sequence"`
}

type LiveSnapshot struct {
	SchemaVersion    int                   `json:"schema_version"`
	Spec             Spec                  `json:"spec"`
	Phase            Phase                 `json:"phase"`
	Result           *Result               `json:"result,omitempty"`
	Tactical         *TacticalLiveSnapshot `json:"tactical,omitempty"`
	TacticalRevision uint64                `json:"tactical_revision"`
}

func (s *Session) LiveSnapshot() LiveSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := LiveSnapshot{
		SchemaVersion:    LiveSnapshotSchemaVersion,
		Spec:             cloneSpec(s.spec),
		Phase:            s.phase,
		Result:           cloneLiveResult(s.result),
		TacticalRevision: s.tacticalRevision,
	}
	if s.tactical != nil {
		out.Tactical = &TacticalLiveSnapshot{
			State:             cloneTacticalState(s.tactical.state),
			Events:            cloneTacticalEvents(s.tactical.events),
			NextEventSequence: s.tactical.nextEventSequence,
		}
	}
	return out
}

func RestoreLiveSnapshot(snapshot LiveSnapshot) (*Session, error) {
	if snapshot.SchemaVersion != LiveSnapshotSchemaVersion {
		return nil, fmt.Errorf("battle live snapshot schema %d, expected %d", snapshot.SchemaVersion, LiveSnapshotSchemaVersion)
	}
	base, err := NewSession(snapshot.Spec)
	if err != nil {
		return nil, fmt.Errorf("restore battle spec: %w", err)
	}
	if err := validateLiveSnapshot(snapshot); err != nil {
		return nil, err
	}
	base.phase = snapshot.Phase
	base.result = cloneLiveResult(snapshot.Result)
	base.tacticalRevision = snapshot.TacticalRevision
	if snapshot.Tactical != nil {
		runtime := tacticalRuntime{
			state:             cloneTacticalState(snapshot.Tactical.State),
			events:            cloneTacticalEvents(snapshot.Tactical.Events),
			nextEventSequence: snapshot.Tactical.NextEventSequence,
		}
		base.tactical = &runtime
	}
	return base, nil
}

func validateLiveSnapshot(snapshot LiveSnapshot) error {
	switch snapshot.Phase {
	case PhasePending, PhaseActive, PhaseCompleted:
	default:
		return fmt.Errorf("battle live snapshot has invalid phase %q", snapshot.Phase)
	}
	if snapshot.Phase == PhasePending {
		if snapshot.Result != nil || snapshot.Tactical != nil || snapshot.TacticalRevision != 0 {
			return fmt.Errorf("pending battle live snapshot carries active/completed state")
		}
		return nil
	}
	if snapshot.Phase == PhaseActive && snapshot.Result != nil {
		return fmt.Errorf("active battle live snapshot carries result")
	}
	if snapshot.Phase == PhaseCompleted && snapshot.Result == nil {
		return fmt.Errorf("completed battle live snapshot has no result")
	}
	if snapshot.Result != nil {
		normalized, err := normalizeResult(snapshot.Spec, *snapshot.Result)
		if err != nil {
			return fmt.Errorf("battle live snapshot result: %w", err)
		}
		if !equalLiveResult(normalized, *snapshot.Result) {
			return fmt.Errorf("battle live snapshot result is not canonical")
		}
	}
	if snapshot.Spec.Tactical == nil {
		if snapshot.Tactical != nil || snapshot.TacticalRevision != 0 {
			return fmt.Errorf("non-tactical battle live snapshot carries tactical state")
		}
		return nil
	}
	if snapshot.Tactical == nil {
		return fmt.Errorf("started tactical battle live snapshot has no tactical state")
	}
	if err := validateTacticalLiveSnapshot(*snapshot.Spec.Tactical, *snapshot.Tactical, snapshot.TacticalRevision); err != nil {
		return err
	}
	if snapshot.Phase == PhaseCompleted {
		destroyed := destroyedTacticalShipIDs(snapshot.Tactical.State)
		if !equalCoreIDList(destroyed, snapshot.Result.DestroyedShipIDs) {
			return fmt.Errorf("completed tactical live snapshot result destroyed ships disagree with tactical state")
		}
	}
	return nil
}

func validateTacticalLiveSnapshot(spec TacticalSpec, snapshot TacticalLiveSnapshot, revision uint64) error {
	state := snapshot.State
	if state.Round == 0 || state.NextCommandSequence == 0 {
		return fmt.Errorf("tactical live snapshot has invalid round/command sequence")
	}
	if uint64(state.NextCommandSequence-1) != revision {
		return fmt.Errorf("tactical live snapshot revision %d disagrees with next command sequence %d", revision, state.NextCommandSequence)
	}
	if len(state.Ships) != len(spec.Ships) {
		return fmt.Errorf("tactical live snapshot ship count %d, expected %d", len(state.Ships), len(spec.Ships))
	}
	shipIDs := make(map[core.ID]struct{}, len(spec.Ships))
	for i := range spec.Ships {
		want := spec.Ships[i]
		got := state.Ships[i]
		if got.ShipID != want.ShipID {
			return fmt.Errorf("tactical live snapshot ship[%d] id %d, expected %d", i, got.ShipID, want.ShipID)
		}
		shipIDs[got.ShipID] = struct{}{}
		if got.ArmorCurrent < 0 || got.ArmorCurrent > want.ArmorMax || got.StructureDamage < 0 || got.StructureDamage > want.StructureMax {
			return fmt.Errorf("tactical live snapshot ship %d damage state outside bounds", got.ShipID)
		}
		if got.Destroyed != (got.StructureDamage >= want.StructureMax) {
			return fmt.Errorf("tactical live snapshot ship %d destroyed flag disagrees with structure", got.ShipID)
		}
		if len(got.Weapons) != len(want.Weapons) {
			return fmt.Errorf("tactical live snapshot ship %d weapon-state count mismatch", got.ShipID)
		}
		for j := range want.Weapons {
			if got.Weapons[j].Slot != want.Weapons[j].Slot {
				return fmt.Errorf("tactical live snapshot ship %d weapon slot mismatch", got.ShipID)
			}
		}
	}
	seenInitiative := make(map[core.ID]struct{}, len(state.InitiativeOrder))
	for _, id := range state.InitiativeOrder {
		if _, ok := shipIDs[id]; !ok {
			return fmt.Errorf("tactical live snapshot initiative references unknown ship %d", id)
		}
		if _, ok := seenInitiative[id]; ok {
			return fmt.Errorf("tactical live snapshot initiative duplicates ship %d", id)
		}
		seenInitiative[id] = struct{}{}
	}
	if state.ActiveShipID == 0 {
		return fmt.Errorf("tactical live snapshot has zero active ship")
	}
	if _, ok := shipIDs[state.ActiveShipID]; !ok {
		return fmt.Errorf("tactical live snapshot active ship %d is unknown", state.ActiveShipID)
	}
	var maxEvent uint64
	for i := range snapshot.Events {
		event := snapshot.Events[i]
		if event.Sequence == 0 || event.Sequence <= maxEvent || event.Kind == "" {
			return fmt.Errorf("tactical live snapshot event[%d] sequence/kind invalid", i)
		}
		if len(event.Data) != 0 && !json.Valid(event.Data) {
			return fmt.Errorf("tactical live snapshot event[%d] data invalid", i)
		}
		maxEvent = event.Sequence
	}
	if snapshot.NextEventSequence == 0 || snapshot.NextEventSequence != maxEvent+1 {
		return fmt.Errorf("tactical live snapshot next event sequence %d, expected %d", snapshot.NextEventSequence, maxEvent+1)
	}
	return nil
}

func destroyedTacticalShipIDs(state TacticalState) []core.ID {
	ids := make([]core.ID, 0, len(state.Ships))
	for _, ship := range state.Ships {
		if ship.Destroyed {
			ids = append(ids, ship.ShipID)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func cloneLiveResult(result *Result) *Result {
	if result == nil {
		return nil
	}
	out := *result
	out.WinnerSeats = append([]protocol.SeatID(nil), result.WinnerSeats...)
	out.DestroyedShipIDs = append([]core.ID(nil), result.DestroyedShipIDs...)
	return &out
}

func equalLiveResult(a, b Result) bool {
	if a.WinnerSeat != b.WinnerSeat || a.Outcome != b.Outcome || len(a.WinnerSeats) != len(b.WinnerSeats) || len(a.DestroyedShipIDs) != len(b.DestroyedShipIDs) {
		return false
	}
	for i := range a.WinnerSeats {
		if a.WinnerSeats[i] != b.WinnerSeats[i] {
			return false
		}
	}
	return equalCoreIDList(a.DestroyedShipIDs, b.DestroyedShipIDs)
}

func equalCoreIDList(a, b []core.ID) bool {
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
