package battle

import (
	"encoding/json"
	"fmt"
	"sort"

	"moox/internal/core"
	"moox/internal/protocol"
)

const (
	LiveSnapshotSchemaVersion       = 2
	legacyLiveSnapshotSchemaVersion = 1
)

type TacticalLiveSnapshot struct {
	State                TacticalState          `json:"state"`
	Ordnance             *TacticalOrdnanceState `json:"ordnance,omitempty"`
	Fighters             *TacticalFighterState  `json:"fighters,omitempty"`
	Events               []TacticalEvent        `json:"events"`
	NextEventSequence    uint64                 `json:"next_event_sequence"`
	NextMovingActorOrder uint64                 `json:"next_moving_actor_order,omitempty"`
	Legacy17_8Disabled   bool                   `json:"legacy_17_8_disabled,omitempty"`
	Slice178StateVersion int                    `json:"slice_17_8_state_version,omitempty"`
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
		ordnance := cloneTacticalOrdnanceState(s.tactical.ordnance)
		fighters := cloneTacticalFighterState(s.tactical.fighters)
		out.Tactical = &TacticalLiveSnapshot{
			State:                cloneTacticalState(s.tactical.state),
			Ordnance:             &ordnance,
			Fighters:             &fighters,
			Events:               cloneTacticalEvents(s.tactical.events),
			NextEventSequence:    s.tactical.nextEventSequence,
			NextMovingActorOrder: s.tactical.nextMovingActorOrder,
			Legacy17_8Disabled:   s.tactical.legacy17_8Disabled,
			Slice178StateVersion: 3,
		}
	}
	return out
}

func RestoreLiveSnapshot(snapshot LiveSnapshot) (*Session, error) {
	migrated, err := migrateLiveSnapshot(snapshot)
	if err != nil {
		return nil, err
	}
	snapshot = migrated
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
		ordnance := initialTacticalOrdnanceState(*snapshot.Spec.Tactical)
		if snapshot.Tactical.Ordnance != nil {
			ordnance = cloneTacticalOrdnanceState(*snapshot.Tactical.Ordnance)
		}
		fighters := TacticalFighterState{NextGroupID: 1}
		if snapshot.Tactical.Fighters != nil {
			fighters = cloneTacticalFighterState(*snapshot.Tactical.Fighters)
		} else if !snapshot.Tactical.Legacy17_8Disabled {
			fighters = initialTacticalFighterState(*snapshot.Spec.Tactical)
		}
		runtime := tacticalRuntime{
			state:                cloneTacticalState(snapshot.Tactical.State),
			ordnance:             ordnance,
			fighters:             fighters,
			events:               cloneTacticalEvents(snapshot.Tactical.Events),
			nextEventSequence:    snapshot.Tactical.NextEventSequence,
			nextMovingActorOrder: snapshot.Tactical.NextMovingActorOrder,
			legacy17_8Disabled:   snapshot.Tactical.Legacy17_8Disabled,
		}
		base.tactical = &runtime
	}
	return base, nil
}

func migrateLiveSnapshot(snapshot LiveSnapshot) (LiveSnapshot, error) {
	switch snapshot.SchemaVersion {
	case LiveSnapshotSchemaVersion:
		if snapshot.Tactical != nil && snapshot.Tactical.Slice178StateVersion == 0 {
			if err := migrateSlice178Checkpoint2State(&snapshot, false); err != nil {
				return LiveSnapshot{}, err
			}
		}
		if snapshot.Tactical != nil && snapshot.Tactical.Slice178StateVersion == 1 {
			if err := migrateSlice178Checkpoint3State(&snapshot, snapshot.Tactical.Legacy17_8Disabled); err != nil {
				return LiveSnapshot{}, err
			}
		}
		if snapshot.Tactical != nil && snapshot.Tactical.Slice178StateVersion == 2 {
			if err := migrateSlice178Checkpoint4State(&snapshot); err != nil {
				return LiveSnapshot{}, err
			}
		}
		return snapshot, nil
	case legacyLiveSnapshotSchemaVersion:
		snapshot.SchemaVersion = LiveSnapshotSchemaVersion
		if snapshot.Tactical == nil {
			return snapshot, nil
		}
		tactical := *snapshot.Tactical
		tactical.State = cloneTacticalState(snapshot.Tactical.State)
		tactical.Events = cloneTacticalEvents(snapshot.Tactical.Events)
		if snapshot.Tactical.Ordnance != nil {
			ordnance := cloneTacticalOrdnanceState(*snapshot.Tactical.Ordnance)
			tactical.Ordnance = &ordnance
		}
		snapshot.Tactical = &tactical
		snapshot.Tactical.Legacy17_8Disabled = true
		var maxSpawnOrder uint64
		if snapshot.Tactical.Ordnance != nil {
			for i := range snapshot.Tactical.Ordnance.Projectiles {
				projectile := &snapshot.Tactical.Ordnance.Projectiles[i]
				projectile.SpawnOrder = projectile.ProjectileID
				if projectile.SpawnOrder > maxSpawnOrder {
					maxSpawnOrder = projectile.SpawnOrder
				}
			}
		}
		nextOrder := maxSpawnOrder + 1
		if snapshot.Tactical.Ordnance != nil && snapshot.Tactical.Ordnance.NextProjectileID > nextOrder {
			nextOrder = snapshot.Tactical.Ordnance.NextProjectileID
		}
		snapshot.Tactical.NextMovingActorOrder = nextOrder
		if err := migrateSlice178Checkpoint2State(&snapshot, true); err != nil {
			return LiveSnapshot{}, err
		}
		if err := migrateSlice178Checkpoint3State(&snapshot, true); err != nil {
			return LiveSnapshot{}, err
		}
		if err := migrateSlice178Checkpoint4State(&snapshot); err != nil {
			return LiveSnapshot{}, err
		}
		return snapshot, nil
	default:
		return LiveSnapshot{}, fmt.Errorf("battle live snapshot schema %d, expected %d or %d", snapshot.SchemaVersion, legacyLiveSnapshotSchemaVersion, LiveSnapshotSchemaVersion)
	}
}

func migrateSlice178Checkpoint2State(snapshot *LiveSnapshot, legacyDisabled bool) error {
	if snapshot == nil || snapshot.Tactical == nil || snapshot.Spec.Tactical == nil {
		return nil
	}
	if snapshot.Tactical.Slice178StateVersion != 0 {
		return nil
	}
	tactical := *snapshot.Tactical
	tactical.State = cloneTacticalState(snapshot.Tactical.State)
	snapshot.Tactical = &tactical
	if len(snapshot.Tactical.State.Ships) != len(snapshot.Spec.Tactical.Ships) {
		return fmt.Errorf("Slice 17.8 checkpoint-2 migration ship count mismatch")
	}
	for i := range snapshot.Tactical.State.Ships {
		state := &snapshot.Tactical.State.Ships[i]
		spec := snapshot.Spec.Tactical.Ships[i]
		if state.ShipID != spec.ShipID {
			return fmt.Errorf("Slice 17.8 checkpoint-2 migration ship[%d] id mismatch", i)
		}
		state.CurrentCombatSpeed = spec.CurrentCombatSpeed
		state.MovementMax = spec.CurrentCombatSpeed
		if state.MovementCurrent > state.MovementMax {
			state.MovementCurrent = state.MovementMax
		}
		if len(state.Weapons) != len(spec.Weapons) {
			return fmt.Errorf("Slice 17.8 checkpoint-2 migration ship %d weapon count mismatch", state.ShipID)
		}
		for j := range state.Weapons {
			if state.Weapons[j].Slot != spec.Weapons[j].Slot {
				return fmt.Errorf("Slice 17.8 checkpoint-2 migration ship %d weapon slot mismatch", state.ShipID)
			}
			state.Weapons[j].CountCurrent = spec.Weapons[j].Count
		}
		if !legacyDisabled {
			state.Internal = initialTacticalInternalState(spec.Internal)
		}
	}
	snapshot.Tactical.Slice178StateVersion = 1
	return nil
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
	if snapshot.Slice178StateVersion != 3 {
		return fmt.Errorf("tactical live snapshot has unsupported Slice-17.8 state version %d", snapshot.Slice178StateVersion)
	}
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
		expectedSpeed := tacticalCurrentCombatSpeed(spec, want, got, snapshot.Legacy17_8Disabled)
		if got.CurrentCombatSpeed != expectedSpeed || got.MovementMax != expectedSpeed || got.MovementCurrent < 0 || got.MovementCurrent > got.MovementMax {
			return fmt.Errorf("tactical live snapshot ship %d combat-speed/movement state disagrees with derived speed %d", got.ShipID, expectedSpeed)
		}
		if got.ArmorCurrent < 0 || got.ArmorCurrent > want.ArmorMax || got.StructureDamage < 0 || got.StructureDamage > want.StructureMax {
			return fmt.Errorf("tactical live snapshot ship %d damage state outside bounds", got.ShipID)
		}
		if got.MarineCurrent < 0 || got.MarineCurrent > want.MarineMax {
			return fmt.Errorf("tactical live snapshot ship %d MarineCurrent %d outside 0..%d", got.ShipID, got.MarineCurrent, want.MarineMax)
		}
		shieldArcMax := tacticalShieldArcMaxForState(spec, want, got, snapshot.Legacy17_8Disabled)
		for arc, current := range got.ShieldArcCurrent {
			if current < 0 || current > shieldArcMax {
				return fmt.Errorf("tactical live snapshot ship %d shield arc %d current %d outside 0..%d", got.ShipID, arc, current, shieldArcMax)
			}
		}
		if got.Destroyed != (got.StructureDamage >= want.StructureMax) {
			return fmt.Errorf("tactical live snapshot ship %d destroyed flag disagrees with structure", got.ShipID)
		}
		if len(got.Weapons) != len(want.Weapons) {
			return fmt.Errorf("tactical live snapshot ship %d weapon-state count mismatch", got.ShipID)
		}
		for j := range want.Weapons {
			if got.Weapons[j].Slot != want.Weapons[j].Slot || got.Weapons[j].CountCurrent < 0 || got.Weapons[j].CountCurrent > want.Weapons[j].Count {
				return fmt.Errorf("tactical live snapshot ship %d weapon slot mismatch", got.ShipID)
			}
		}
		if !snapshot.Legacy17_8Disabled {
			if err := validateTacticalInternalState(want.Internal, got.Internal); err != nil {
				return fmt.Errorf("tactical live snapshot ship %d internal state: %w", got.ShipID, err)
			}
			if want.Internal != nil {
				for j, immutable := range want.Internal.Weapons {
					liveCount := 0
					if got.Internal.Weapons[j].Current > 0 {
						liveCount = (got.Internal.Weapons[j].Current + immutable.UnitSize - 1) / immutable.UnitSize
					}
					if got.Weapons[j].CountCurrent != liveCount {
						return fmt.Errorf("tactical live snapshot ship %d weapon slot %d live count disagrees with internal pool", got.ShipID, immutable.Slot)
					}
				}
			}
		}
	}
	if snapshot.Ordnance == nil {
		if tacticalSpecHasExecutableOrdnance(spec) {
			return fmt.Errorf("tactical live snapshot with executable ordnance is missing ordnance state")
		}
	} else if err := validateTacticalOrdnanceState(spec, *snapshot.Ordnance); err != nil {
		return fmt.Errorf("tactical live snapshot ordnance state: %w", err)
	}
	if snapshot.Fighters == nil {
		if !snapshot.Legacy17_8Disabled && tacticalSpecHasExecutableFighters(spec) {
			return fmt.Errorf("tactical live snapshot with executable Fighters is missing Fighter state")
		}
	} else if err := validateTacticalFighterState(spec, *snapshot.Fighters, snapshot.Legacy17_8Disabled); err != nil {
		return fmt.Errorf("tactical live snapshot Fighter state: %w", err)
	}
	if snapshot.NextMovingActorOrder == 0 {
		return fmt.Errorf("tactical live snapshot next moving-actor order must be non-zero")
	}
	seenSpawnOrders := make(map[uint64]struct{})
	projectilesByID := make(map[uint64]TacticalProjectileState)
	if snapshot.Ordnance != nil {
		for i, projectile := range snapshot.Ordnance.Projectiles {
			if projectile.SpawnOrder == 0 || projectile.SpawnOrder >= snapshot.NextMovingActorOrder {
				return fmt.Errorf("tactical projectile[%d] spawn order %d is outside 1..%d", i, projectile.SpawnOrder, snapshot.NextMovingActorOrder-1)
			}
			if _, exists := seenSpawnOrders[projectile.SpawnOrder]; exists {
				return fmt.Errorf("tactical projectile[%d] duplicates spawn order %d", i, projectile.SpawnOrder)
			}
			seenSpawnOrders[projectile.SpawnOrder] = struct{}{}
			projectilesByID[projectile.ProjectileID] = projectile
		}
	}
	fighterGroupsByID := make(map[uint64]TacticalFighterGroupState)
	if snapshot.Fighters != nil {
		for i, group := range snapshot.Fighters.Groups {
			if group.SpawnOrder == 0 || group.SpawnOrder >= snapshot.NextMovingActorOrder {
				return fmt.Errorf("tactical Fighter group[%d] spawn order %d is outside 1..%d", i, group.SpawnOrder, snapshot.NextMovingActorOrder-1)
			}
			if _, exists := seenSpawnOrders[group.SpawnOrder]; exists {
				return fmt.Errorf("tactical Fighter group[%d] duplicates moving-actor spawn order %d", i, group.SpawnOrder)
			}
			seenSpawnOrders[group.SpawnOrder] = struct{}{}
			fighterGroupsByID[group.GroupID] = group
		}
		for _, group := range snapshot.Fighters.Groups {
			source := tacticalShipSpec(spec, group.SourceShipID)
			if group.TargetProjectileID != 0 {
				target, ok := projectilesByID[group.TargetProjectileID]
				if !ok {
					return fmt.Errorf("tactical Fighter group %d references missing target projectile %d", group.GroupID, group.TargetProjectileID)
				}
				targetSource := tacticalShipSpec(spec, target.SourceShipID)
				if source == nil || targetSource == nil || source.EmpireID == targetSource.EmpireID {
					return fmt.Errorf("tactical Fighter group %d target projectile %d is not hostile", group.GroupID, group.TargetProjectileID)
				}
			}
			if group.TargetFighterGroupID != 0 {
				target, ok := fighterGroupsByID[group.TargetFighterGroupID]
				if !ok {
					return fmt.Errorf("tactical Fighter group %d references missing target Fighter group %d", group.GroupID, group.TargetFighterGroupID)
				}
				if target.GroupID == group.GroupID || target.SourceSeatID == group.SourceSeatID {
					return fmt.Errorf("tactical Fighter group %d target Fighter group %d is not hostile", group.GroupID, group.TargetFighterGroupID)
				}
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

func migrateSlice178Checkpoint3State(snapshot *LiveSnapshot, legacyDisabled bool) error {
	if snapshot == nil || snapshot.Tactical == nil || snapshot.Spec.Tactical == nil {
		return nil
	}
	if snapshot.Tactical.Slice178StateVersion != 1 {
		return nil
	}
	tactical := *snapshot.Tactical
	tactical.State = cloneTacticalState(snapshot.Tactical.State)
	tactical.Events = cloneTacticalEvents(snapshot.Tactical.Events)
	if snapshot.Tactical.Ordnance != nil {
		ordnance := cloneTacticalOrdnanceState(*snapshot.Tactical.Ordnance)
		tactical.Ordnance = &ordnance
	}
	fighters := TacticalFighterState{NextGroupID: 1}
	if !legacyDisabled {
		fighters = initialTacticalFighterState(*snapshot.Spec.Tactical)
	}
	tactical.Fighters = &fighters
	tactical.Legacy17_8Disabled = legacyDisabled || tactical.Legacy17_8Disabled
	tactical.Slice178StateVersion = 2
	snapshot.Tactical = &tactical
	return nil
}

func migrateSlice178Checkpoint4State(snapshot *LiveSnapshot) error {
	if snapshot == nil || snapshot.Tactical == nil || snapshot.Spec.Tactical == nil {
		return nil
	}
	if snapshot.Tactical.Slice178StateVersion != 2 {
		return nil
	}
	tactical := *snapshot.Tactical
	tactical.State = cloneTacticalState(snapshot.Tactical.State)
	if len(tactical.State.Ships) != len(snapshot.Spec.Tactical.Ships) {
		return fmt.Errorf("Slice 17.8 checkpoint-4 migration ship count mismatch")
	}
	for i := range tactical.State.Ships {
		state := &tactical.State.Ships[i]
		spec := snapshot.Spec.Tactical.Ships[i]
		if state.ShipID != spec.ShipID {
			return fmt.Errorf("Slice 17.8 checkpoint-4 migration ship[%d] id mismatch", i)
		}
		state.OwnerEmpireID = spec.EmpireID
		state.ControllerSeatID = spec.SeatID
		state.TeleporterUsed = false
		state.AlreadyBoarded = false
		state.BeingCaptured = false
		state.PostCaptureState = false
	}
	tactical.Slice178StateVersion = 3
	snapshot.Tactical = &tactical
	return nil
}
