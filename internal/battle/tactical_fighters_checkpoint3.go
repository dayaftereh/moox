package battle

import (
	"fmt"
	"sort"

	"moox/internal/beam"
	"moox/internal/core"
	"moox/internal/protocol"
)

const (
	CommandLaunchFighters = "battle.launch_fighters"

	TacticalWeaponUnavailableFighterBaysEmpty = "fighter_bays_empty"
	TacticalWeaponUnavailableFighterTarget    = "fighter_target_unavailable"

	TacticalFighterPhaseOutbound  = "outbound"
	TacticalFighterPhaseReturning = "returning"
)

type TacticalFighterBayState struct {
	ShipID      core.ID `json:"ship_id"`
	WeaponSlot  int     `json:"weapon_slot"`
	ReadyGroups int     `json:"ready_groups"`
}

type TacticalFighterGroupState struct {
	GroupID                  uint64                      `json:"group_id"`
	SpawnOrder               uint64                      `json:"spawn_order"`
	SourceSeatID             protocol.SeatID             `json:"source_seat_id"`
	SourceShipID             core.ID                     `json:"source_ship_id"`
	SourceWeaponSlot         int                         `json:"source_weapon_slot"`
	Kind                     string                      `json:"kind"`
	TargetShipID             core.ID                     `json:"target_ship_id,omitempty"`
	TargetProjectileID       uint64                      `json:"target_projectile_id,omitempty"`
	TargetFighterGroupID     uint64                      `json:"target_fighter_group_id,omitempty"`
	X                        int                         `json:"x"`
	Y                        int                         `json:"y"`
	Facing                   int                         `json:"facing"`
	Speed                    int                         `json:"speed"`
	MovementRemaining        int                         `json:"movement_remaining"`
	DistanceTraveled         int                         `json:"distance_traveled"`
	Multiplicity             int                         `json:"multiplicity"`
	UnitDurability           int                         `json:"unit_durability"`
	InterceptDamageRemainder int                         `json:"intercept_damage_remainder"`
	AttackStock              int                         `json:"attack_stock"`
	Phase                    string                      `json:"phase"`
	BeamPayload              *TacticalFighterPayloadSpec `json:"beam_payload,omitempty"`
	BombPayload              *TacticalFighterPayloadSpec `json:"bomb_payload,omitempty"`
}

type TacticalFighterState struct {
	NextGroupID uint64                      `json:"next_group_id"`
	BayStates   []TacticalFighterBayState   `json:"bay_states,omitempty"`
	Groups      []TacticalFighterGroupState `json:"groups,omitempty"`
}

type TacticalFighterView struct {
	BayStates []TacticalFighterBayState   `json:"bay_states,omitempty"`
	Groups    []TacticalFighterGroupState `json:"groups,omitempty"`
}

type LaunchFightersPayload struct {
	ShipID               core.ID `json:"ship_id"`
	WeaponSlot           int     `json:"weapon_slot"`
	TargetShipID         core.ID `json:"target_ship_id,omitempty"`
	TargetProjectileID   uint64  `json:"target_projectile_id,omitempty"`
	TargetFighterGroupID uint64  `json:"target_fighter_group_id,omitempty"`
}

func initialTacticalFighterState(spec TacticalSpec) TacticalFighterState {
	out := TacticalFighterState{NextGroupID: 1}
	for _, ship := range spec.Ships {
		for _, weapon := range ship.Weapons {
			if weapon.TacticalSupport != "fighter" {
				continue
			}
			out.BayStates = append(out.BayStates, TacticalFighterBayState{
				ShipID: ship.ShipID, WeaponSlot: weapon.Slot, ReadyGroups: weapon.Count,
			})
		}
	}
	return out
}

func cloneTacticalFighterState(in TacticalFighterState) TacticalFighterState {
	out := TacticalFighterState{
		NextGroupID: in.NextGroupID,
		BayStates:   append([]TacticalFighterBayState(nil), in.BayStates...),
		Groups:      make([]TacticalFighterGroupState, len(in.Groups)),
	}
	for i := range in.Groups {
		out.Groups[i] = in.Groups[i]
		out.Groups[i].BeamPayload = cloneTacticalFighterPayload(in.Groups[i].BeamPayload)
		out.Groups[i].BombPayload = cloneTacticalFighterPayload(in.Groups[i].BombPayload)
	}
	return out
}

func tacticalFighterBayState(state *TacticalFighterState, shipID core.ID, slot int) *TacticalFighterBayState {
	if state == nil {
		return nil
	}
	for i := range state.BayStates {
		if state.BayStates[i].ShipID == shipID && state.BayStates[i].WeaponSlot == slot {
			return &state.BayStates[i]
		}
	}
	return nil
}

func tacticalFighterDoctrine(spec TacticalSpec, seatID protocol.SeatID) *TacticalFighterDoctrineSpec {
	for i := range spec.FighterDoctrines {
		if spec.FighterDoctrines[i].SeatID == seatID {
			return &spec.FighterDoctrines[i]
		}
	}
	return nil
}

func tacticalFighterKind(weapon TacticalWeaponSpec) string {
	switch weapon.WeaponID {
	case "fighter_bays":
		return "fighter"
	case "bomber_bays":
		return "bomber"
	case "heavy_fighter_bays":
		return "heavy_fighter"
	default:
		return ""
	}
}

func tacticalFighterStats(spec TacticalSpec, carrier TacticalShipSpec, weapon TacticalWeaponSpec) (kind string, speed, durability, stock int, doctrine *TacticalFighterDoctrineSpec, err error) {
	if spec.Rules.Fighter == nil {
		return "", 0, 0, 0, nil, fmt.Errorf("tactical Fighter rules are missing")
	}
	kind = tacticalFighterKind(weapon)
	if kind == "" {
		return "", 0, 0, 0, nil, fmt.Errorf("weapon %q is not an executable Fighter row", weapon.WeaponID)
	}
	doctrine = tacticalFighterDoctrine(spec, carrier.SeatID)
	if doctrine == nil {
		return "", 0, 0, 0, nil, fmt.Errorf("seat %d has no Fighter doctrine", carrier.SeatID)
	}
	rules := spec.Rules.Fighter
	baseSpeed, baseHits := 0, 0
	switch kind {
	case "fighter":
		baseSpeed, baseHits, stock = rules.FighterBaseSpeed, rules.FighterBaseHits, rules.FighterAttackStock
		if doctrine.BeamPayload == nil {
			return "", 0, 0, 0, nil, fmt.Errorf("Fighter doctrine for seat %d has no Beam payload", carrier.SeatID)
		}
	case "bomber":
		baseSpeed, baseHits, stock = rules.BomberBaseSpeed, rules.BomberBaseHits, rules.BomberAttackStock
		if doctrine.BombPayload == nil {
			return "", 0, 0, 0, nil, fmt.Errorf("Fighter doctrine for seat %d has no Bomb payload", carrier.SeatID)
		}
	case "heavy_fighter":
		baseSpeed, baseHits, stock = rules.HeavyFighterBaseSpeed, rules.HeavyFighterBaseHits, rules.HeavyFighterAttackStock
		if doctrine.BeamPayload == nil || doctrine.BombPayload == nil {
			return "", 0, 0, 0, nil, fmt.Errorf("Heavy Fighter doctrine for seat %d requires Beam and Bomb payloads", carrier.SeatID)
		}
	}
	speed = baseSpeed + rules.WarpDriveSpeedStep*(doctrine.BestWarpDriveIndex-1)
	if doctrine.TransDimensional {
		speed += rules.TransDimensionalSpeedBonus
	}
	durability = baseHits * doctrine.ArmorDurabilityMultiple
	return kind, speed, durability, stock, doctrine, nil
}

func fighterTargetCount(payload LaunchFightersPayload) int {
	n := 0
	if payload.TargetShipID != 0 {
		n++
	}
	if payload.TargetProjectileID != 0 {
		n++
	}
	if payload.TargetFighterGroupID != 0 {
		n++
	}
	return n
}

func tacticalFighterShipTargets(spec TacticalSpec, r *tacticalRuntime, source TacticalShipSpec) []TacticalFireTarget {
	var targets []TacticalFireTarget
	for _, target := range spec.Ships {
		if target.EmpireID == source.EmpireID {
			continue
		}
		state := r.shipState(target.ShipID)
		if state == nil || state.Destroyed {
			continue
		}
		targets = append(targets, TacticalFireTarget{TargetShipID: target.ShipID, RangeIndex: tacticalRangeIndex(r.shipState(source.ShipID).X, r.shipState(source.ShipID).Y, state.X, state.Y)})
	}
	return targets
}

func tacticalFighterWeaponAction(spec TacticalSpec, r *tacticalRuntime, activeSpec *TacticalShipSpec, activeState *TacticalShipState, weapon TacticalWeaponSpec) TacticalWeaponActionAvailability {
	action := TacticalWeaponActionAvailability{
		ShipID: activeSpec.ShipID, WeaponSlot: weapon.Slot, WeaponID: weapon.WeaponID,
		WeaponFamily: "fighter", CommandKind: CommandLaunchFighters,
	}
	if r.legacy17_8Disabled {
		action.UnavailableReason = TacticalWeaponUnavailableConsumer
		return action
	}
	bay := tacticalFighterBayState(&r.fighters, activeSpec.ShipID, weapon.Slot)
	if bay == nil || bay.ReadyGroups <= 0 {
		action.UnavailableReason = TacticalWeaponUnavailableFighterBaysEmpty
		return action
	}
	ready := weaponState(activeState, weapon.Slot)
	if ready == nil || !ready.Ready {
		action.UnavailableReason = TacticalWeaponUnavailableNotReady
		return action
	}
	if _, _, _, _, _, err := tacticalFighterStats(spec, *activeSpec, weapon); err != nil {
		action.UnavailableReason = TacticalWeaponUnavailableConsumer
		return action
	}
	action.Targets = tacticalFighterLegalTargets(spec, r, *activeSpec, tacticalFighterKind(weapon))
	if len(action.Targets) == 0 {
		action.UnavailableReason = TacticalWeaponUnavailableNoTargets
		return action
	}
	action.Available = true
	return action
}

func refreshTacticalFighterReadiness(spec TacticalSpec, r *tacticalRuntime) {
	if r == nil || r.legacy17_8Disabled {
		return
	}
	for _, ship := range spec.Ships {
		state := r.shipState(ship.ShipID)
		if state == nil || state.Destroyed {
			continue
		}
		for _, weapon := range ship.Weapons {
			if weapon.TacticalSupport != "fighter" {
				continue
			}
			ready := weaponState(state, weapon.Slot)
			bay := tacticalFighterBayState(&r.fighters, ship.ShipID, weapon.Slot)
			if ready != nil && bay != nil {
				ready.Ready = bay.ReadyGroups > 0
			}
		}
	}
}

func prepareLaunchFighters(spec Spec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32, payload LaunchFightersPayload) error {
	if spec.Tactical == nil || r == nil {
		return fmt.Errorf("tactical Fighter launch requires active tactical state")
	}
	if r.legacy17_8Disabled {
		return fmt.Errorf("Fighter launch is unavailable for legacy Slice-17.8 battles")
	}
	active := tacticalShipSpec(*spec.Tactical, r.state.ActiveShipID)
	activeState := r.shipState(r.state.ActiveShipID)
	if active == nil || activeState == nil || activeState.Destroyed || activeState.ActivationComplete {
		return fmt.Errorf("active tactical ship %d is unavailable", r.state.ActiveShipID)
	}
	if seatID == 0 || active.SeatID != seatID {
		return fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != active.ShipID {
		return fmt.Errorf("launch ship %d is not active ship %d", payload.ShipID, active.ShipID)
	}
	if fighterTargetCount(payload) != 1 {
		return fmt.Errorf("Fighter launch requires exactly one target identity")
	}
	weapon := weaponSpec(*active, payload.WeaponSlot)
	if weapon == nil || weapon.TacticalSupport != "fighter" {
		return fmt.Errorf("ship %d slot %d is not an executable Fighter row", active.ShipID, payload.WeaponSlot)
	}
	action := tacticalFighterWeaponAction(*spec.Tactical, r, active, activeState, *weapon)
	if !action.Available {
		reason := action.UnavailableReason
		if reason == "" {
			reason = TacticalWeaponUnavailableConsumer
		}
		return fmt.Errorf("ship %d weapon slot %d cannot launch Fighters: %s", active.ShipID, payload.WeaponSlot, reason)
	}
	allowed := false
	for _, target := range action.Targets {
		if target.TargetShipID == payload.TargetShipID && target.TargetProjectileID == payload.TargetProjectileID && target.TargetFighterGroupID == payload.TargetFighterGroupID {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("Fighter target is not currently legal")
	}
	bay := tacticalFighterBayState(&r.fighters, active.ShipID, weapon.Slot)
	ready := weaponState(activeState, weapon.Slot)
	if bay == nil || ready == nil || bay.ReadyGroups <= 0 || !ready.Ready {
		return fmt.Errorf("Fighter Bay runtime state is unavailable")
	}
	kind, speed, durability, stock, doctrine, err := tacticalFighterStats(*spec.Tactical, *active, *weapon)
	if err != nil {
		return err
	}
	craftPerUnit := weapon.CraftPerUnit
	if craftPerUnit <= 0 {
		craftPerUnit = spec.Tactical.Rules.Fighter.CraftPerBay
	}
	group := TacticalFighterGroupState{
		GroupID: r.fighters.NextGroupID, SpawnOrder: r.nextMovingActorOrder,
		SourceSeatID: active.SeatID, SourceShipID: active.ShipID, SourceWeaponSlot: weapon.Slot,
		Kind: kind, TargetShipID: payload.TargetShipID, TargetProjectileID: payload.TargetProjectileID, TargetFighterGroupID: payload.TargetFighterGroupID,
		X: activeState.X, Y: activeState.Y, Facing: activeState.Facing,
		Speed: speed, MovementRemaining: speed * spec.Tactical.Rules.Fighter.MovementUnitsPerField,
		Multiplicity: bay.ReadyGroups * craftPerUnit, UnitDurability: durability, AttackStock: stock,
		Phase: TacticalFighterPhaseOutbound,
	}
	if kind == "fighter" || kind == "heavy_fighter" {
		group.BeamPayload = cloneTacticalFighterPayload(doctrine.BeamPayload)
	}
	if kind == "bomber" || kind == "heavy_fighter" {
		group.BombPayload = cloneTacticalFighterPayload(doctrine.BombPayload)
	}
	bay.ReadyGroups = 0
	ready.Ready = false
	r.fighters.NextGroupID++
	r.nextMovingActorOrder++
	r.fighters.Groups = append(r.fighters.Groups, group)
	data := map[string]any{
		"group_id": group.GroupID, "spawn_order": group.SpawnOrder,
		"source_seat_id": group.SourceSeatID, "source_ship_id": group.SourceShipID, "source_weapon_slot": group.SourceWeaponSlot,
		"fighter_kind": group.Kind, "target_ship_id": group.TargetShipID, "target_projectile_id": group.TargetProjectileID, "target_fighter_group_id": group.TargetFighterGroupID,
		"start_x": group.X, "start_y": group.Y, "start_facing": group.Facing,
		"speed": group.Speed, "multiplicity": group.Multiplicity, "unit_durability": group.UnitDurability, "attack_stock": group.AttackStock,
		"ready_groups": bay.ReadyGroups,
	}
	if group.BeamPayload != nil {
		data["beam_weapon_id"] = group.BeamPayload.WeaponID
	}
	if group.BombPayload != nil {
		data["bomb_weapon_id"] = group.BombPayload.WeaponID
	}
	if err := r.appendEvent("fighter_group_launched", seatID, commandSequence, data); err != nil {
		return err
	}
	return processTacticalFighterGroupByID(*spec.Tactical, r, group.GroupID, seatID, commandSequence)
}

type tacticalFighterQuantumOutcome struct {
	Keep                 bool
	RemoveWithoutDestroy bool
	DestroyReason        string
}

func processTacticalFighterGroupByID(spec TacticalSpec, r *tacticalRuntime, groupID uint64, eventSeat protocol.SeatID, commandSequence uint32) error {
	index := tacticalFighterGroupIndexByID(r, groupID)
	if index < 0 {
		return fmt.Errorf("tactical Fighter group %d is missing", groupID)
	}
	before := r.fighters.Groups[index].Multiplicity
	outcome, err := moveTacticalFighterQuantum(spec, r, index, eventSeat, commandSequence)
	if err != nil {
		return err
	}
	if outcome.Keep {
		return nil
	}
	index = tacticalFighterGroupIndexByID(r, groupID)
	if index < 0 {
		if outcome.RemoveWithoutDestroy {
			return nil
		}
		return fmt.Errorf("tactical Fighter group %d disappeared before lifecycle removal", groupID)
	}
	r.fighters.Groups = append(r.fighters.Groups[:index], r.fighters.Groups[index+1:]...)
	if outcome.RemoveWithoutDestroy {
		return nil
	}
	reason := outcome.DestroyReason
	if reason == "" {
		reason = "no_target"
	}
	return r.appendEvent("fighter_group_destroyed", eventSeat, commandSequence, map[string]any{
		"group_id": groupID, "reason": reason, "multiplicity_before": before,
	})
}

func moveTacticalFighterQuantum(spec TacticalSpec, r *tacticalRuntime, index int, eventSeat protocol.SeatID, commandSequence uint32) (tacticalFighterQuantumOutcome, error) {
	if index < 0 || index >= len(r.fighters.Groups) {
		return tacticalFighterQuantumOutcome{}, fmt.Errorf("tactical Fighter group index %d is out of range", index)
	}
	group := &r.fighters.Groups[index]
	groupID := group.GroupID
	if group.Phase != TacticalFighterPhaseOutbound && group.Phase != TacticalFighterPhaseReturning {
		return tacticalFighterQuantumOutcome{}, fmt.Errorf("tactical Fighter group %d has invalid phase %q", group.GroupID, group.Phase)
	}

	targetX, targetY, targetOK := tacticalFighterTargetPosition(spec, r, *group)
	if group.Phase == TacticalFighterPhaseOutbound && !targetOK {
		retargeted, err := tacticalFighterRetarget(spec, r, group, eventSeat, commandSequence)
		if err != nil {
			return tacticalFighterQuantumOutcome{}, err
		}
		if !retargeted {
			if !tacticalFighterCarrierFunctional(spec, r, *group) {
				return tacticalFighterQuantumOutcome{DestroyReason: "carrier_unavailable"}, nil
			}
			if err := tacticalFighterBeginReturn(spec, r, group, "no_target", eventSeat, commandSequence); err != nil {
				return tacticalFighterQuantumOutcome{}, err
			}
		}
		targetX, targetY, targetOK = tacticalFighterTargetPosition(spec, r, *group)
	}
	if group.Phase == TacticalFighterPhaseReturning && !targetOK {
		return tacticalFighterQuantumOutcome{DestroyReason: "carrier_unavailable"}, nil
	}
	if !targetOK {
		return tacticalFighterQuantumOutcome{DestroyReason: "no_target"}, nil
	}
	if spec.Rules.Fighter == nil {
		return tacticalFighterQuantumOutcome{}, fmt.Errorf("tactical Fighter group %d has no Fighter rules", groupID)
	}

	budget := group.Speed * spec.Rules.Fighter.MovementUnitsPerField
	group.MovementRemaining = budget
	if group.X == targetX && group.Y == targetY {
		group.MovementRemaining = 0
		return resolveTacticalFighterArrival(spec, r, groupID, eventSeat, commandSequence)
	}

	fromX, fromY := group.X, group.Y
	toX, toY, movedFields := tacticalProjectileDestination(fromX, fromY, targetX, targetY, group.Speed)
	if movedFields <= 0 {
		return tacticalFighterQuantumOutcome{}, fmt.Errorf("tactical Fighter group %d could not advance", groupID)
	}
	used := movedFields * spec.Rules.Fighter.MovementUnitsPerField
	if used > group.MovementRemaining {
		return tacticalFighterQuantumOutcome{}, fmt.Errorf("tactical Fighter group %d movement %d exceeds budget %d", groupID, used, group.MovementRemaining)
	}
	group.X, group.Y = toX, toY
	group.Facing = tacticalFacing(toX-fromX, toY-fromY)
	group.MovementRemaining -= used
	group.DistanceTraveled += used
	data := map[string]any{
		"group_id": groupID, "from_x": fromX, "from_y": fromY, "to_x": group.X, "to_y": group.Y,
		"movement_used": used, "movement_remaining": group.MovementRemaining, "distance_traveled": group.DistanceTraveled,
		"phase": group.Phase,
	}
	for key, value := range tacticalFighterTargetEventFields(*group) {
		data[key] = value
	}
	if err := r.appendEvent("fighter_group_moved", eventSeat, commandSequence, data); err != nil {
		return tacticalFighterQuantumOutcome{}, err
	}
	if group.X == targetX && group.Y == targetY {
		return resolveTacticalFighterArrival(spec, r, groupID, eventSeat, commandSequence)
	}
	return tacticalFighterQuantumOutcome{Keep: true}, nil
}

func processTacticalMovingActors(spec TacticalSpec, r *tacticalRuntime, eventSeat protocol.SeatID, commandSequence uint32) error {
	type actor struct {
		spawn uint64
		kind  string
		id    uint64
	}
	actors := make([]actor, 0, len(r.ordnance.Projectiles)+len(r.fighters.Groups))
	for _, p := range r.ordnance.Projectiles {
		actors = append(actors, actor{spawn: p.SpawnOrder, kind: "projectile", id: p.ProjectileID})
	}
	for _, g := range r.fighters.Groups {
		actors = append(actors, actor{spawn: g.SpawnOrder, kind: "fighter", id: g.GroupID})
	}
	sort.Slice(actors, func(i, j int) bool { return actors[i].spawn < actors[j].spawn })
	for _, a := range actors {
		switch a.kind {
		case "projectile":
			if tacticalProjectileIndexByID(r, a.id) >= 0 {
				if err := processTacticalProjectileByID(spec, r, a.id, eventSeat, commandSequence); err != nil {
					return err
				}
			}
		case "fighter":
			if tacticalFighterGroupIndexByID(r, a.id) >= 0 {
				if err := processTacticalFighterGroupByID(spec, r, a.id, eventSeat, commandSequence); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func tacticalFighterGroupIndexByID(r *tacticalRuntime, groupID uint64) int {
	if r == nil {
		return -1
	}
	for i := range r.fighters.Groups {
		if r.fighters.Groups[i].GroupID == groupID {
			return i
		}
	}
	return -1
}

type tacticalFighterInterceptionResult struct {
	DestroyedUnits     int
	MultiplicityBefore int
	MultiplicityAfter  int
	RemainderBefore    int
	RemainderAfter     int
	Destroyed          bool
}

func tacticalFighterBeamInterceptorDCV(spec TacticalSpec, group TacticalFighterGroupState) (int, error) {
	if group.Speed <= 0 {
		return 0, fmt.Errorf("Fighter group %d has invalid speed %d", group.GroupID, group.Speed)
	}
	doctrine := tacticalFighterDoctrine(spec, group.SourceSeatID)
	if doctrine == nil {
		return 0, fmt.Errorf("Fighter group %d source seat %d has no doctrine", group.GroupID, group.SourceSeatID)
	}
	return group.Speed*5 + doctrine.MovingActorDefenseBonus, nil
}

func tacticalBeamFighterTargets(spec TacticalSpec, r *tacticalRuntime, activeSpec *TacticalShipSpec, activeState *TacticalShipState, weapon TacticalWeaponSpec) []TacticalFireTarget {
	if activeSpec == nil || activeState == nil || canonicalTacticalWeaponMountType(weapon.MountType) == core.ShipWeaponMountHeavy {
		return nil
	}
	configured, err := tacticalBeamConfigured(weapon)
	if err != nil {
		return nil
	}
	targetPoint := TacticalShipState{}
	out := make([]TacticalFireTarget, 0, len(r.fighters.Groups))
	for _, group := range r.fighters.Groups {
		if group.SourceSeatID == activeSpec.SeatID || group.Multiplicity <= 0 {
			continue
		}
		targetPoint.X = group.X
		targetPoint.Y = group.Y
		if !tacticalBeamTargetInArc(*activeState, targetPoint, weapon.FiringArcID) {
			continue
		}
		geometricRangeIndex := tacticalRangeIndex(activeState.X, activeState.Y, group.X, group.Y)
		rangeIndex := tacticalBeamEffectiveRangeIndex(geometricRangeIndex, len(spec.Rules.BeamToHitRangeModifiers), configured.RangeProfile)
		if rangeIndex < 0 || rangeIndex >= len(spec.Rules.BeamToHitRangeModifiers) || rangeIndex >= len(spec.Rules.BeamDamageRangeModifiers) {
			continue
		}
		out = append(out, TacticalFireTarget{TargetFighterGroupID: group.GroupID, RangeIndex: rangeIndex})
	}
	return out
}

func applyTacticalFighterInterception(
	r *tacticalRuntime,
	group *TacticalFighterGroupState,
	damage int,
	sourceKind string,
	interceptorShipID core.ID,
	interceptorWeaponSlot int,
	eventSeat protocol.SeatID,
	commandSequence uint32,
) (tacticalFighterInterceptionResult, error) {
	if group == nil {
		return tacticalFighterInterceptionResult{}, fmt.Errorf("Fighter interception requires group")
	}
	if damage < 0 {
		return tacticalFighterInterceptionResult{}, fmt.Errorf("Fighter interception damage %d is invalid", damage)
	}
	if group.Multiplicity <= 0 || group.UnitDurability <= 0 ||
		group.InterceptDamageRemainder < 0 || group.InterceptDamageRemainder >= group.UnitDurability {
		return tacticalFighterInterceptionResult{}, fmt.Errorf("Fighter group %d interception state is invalid", group.GroupID)
	}
	result := tacticalFighterInterceptionResult{
		MultiplicityBefore: group.Multiplicity,
		RemainderBefore:    group.InterceptDamageRemainder,
	}
	total := group.InterceptDamageRemainder + damage
	destroyed := total / group.UnitDurability
	if destroyed > group.Multiplicity {
		destroyed = group.Multiplicity
	}
	total -= destroyed * group.UnitDurability
	group.Multiplicity -= destroyed
	group.InterceptDamageRemainder = total
	if group.Multiplicity == 0 {
		group.InterceptDamageRemainder = 0
	}
	result.DestroyedUnits = destroyed
	result.MultiplicityAfter = group.Multiplicity
	result.RemainderAfter = group.InterceptDamageRemainder
	result.Destroyed = group.Multiplicity == 0
	if err := r.appendEvent("fighter_interception_resolved", eventSeat, commandSequence, map[string]any{
		"group_id":                group.GroupID,
		"interceptor_source_kind": sourceKind,
		"interceptor_ship_id":     interceptorShipID,
		"interceptor_weapon_slot": interceptorWeaponSlot,
		"damage":                  damage,
		"destroyed_craft":         result.DestroyedUnits,
		"multiplicity_before":     result.MultiplicityBefore,
		"multiplicity_after":      result.MultiplicityAfter,
		"remainder_before":        result.RemainderBefore,
		"remainder_after":         result.RemainderAfter,
	}); err != nil {
		return tacticalFighterInterceptionResult{}, err
	}
	return result, nil
}

func prepareFireBeamAtFighterGroup(spec Spec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32, payload FireBeamPayload) (*Result, error) {
	active := tacticalShipSpec(*spec.Tactical, r.state.ActiveShipID)
	if active == nil {
		return nil, fmt.Errorf("active tactical ship %d is missing from spec", r.state.ActiveShipID)
	}
	if seatID == 0 || active.SeatID != seatID {
		return nil, fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return nil, fmt.Errorf("fire ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	groupIndex := tacticalFighterGroupIndexByID(r, payload.TargetFighterGroupID)
	if groupIndex < 0 {
		return nil, fmt.Errorf("target Fighter group %d is not active", payload.TargetFighterGroupID)
	}
	group := &r.fighters.Groups[groupIndex]
	if group.SourceSeatID == active.SeatID {
		return nil, fmt.Errorf("target Fighter group %d is not an opponent", payload.TargetFighterGroupID)
	}

	weapon := weaponSpec(*active, payload.WeaponSlot)
	activeState := r.shipState(active.ShipID)
	ready := weaponState(activeState, payload.WeaponSlot)
	if weapon == nil || ready == nil {
		return nil, fmt.Errorf("ship %d has no weapon slot %d", active.ShipID, payload.WeaponSlot)
	}
	if !ready.Ready {
		return nil, fmt.Errorf("ship %d weapon slot %d is not ready", active.ShipID, payload.WeaponSlot)
	}
	if !tacticalBeamFireSupported(*weapon) {
		return nil, fmt.Errorf("weapon slot %d is outside the executable Beam firing contract", payload.WeaponSlot)
	}
	if activeState == nil {
		return nil, fmt.Errorf("active tactical ship %d has no runtime position", active.ShipID)
	}
	if canonicalTacticalWeaponMountType(weapon.MountType) == core.ShipWeaponMountHeavy {
		return nil, fmt.Errorf("Heavy Beam weapon slot %d cannot target Fighter groups", payload.WeaponSlot)
	}

	configured, err := tacticalBeamConfigured(*weapon)
	if err != nil {
		return nil, err
	}
	targetPoint := TacticalShipState{X: group.X, Y: group.Y}
	if !tacticalBeamTargetInArc(*activeState, targetPoint, weapon.FiringArcID) {
		return nil, fmt.Errorf("target Fighter group %d is outside weapon slot %d firing arc %q", group.GroupID, payload.WeaponSlot, weapon.FiringArcID)
	}
	geometricRangeIndex := tacticalRangeIndex(activeState.X, activeState.Y, group.X, group.Y)
	rangeIndex := tacticalBeamEffectiveRangeIndex(geometricRangeIndex, len(spec.Tactical.Rules.BeamToHitRangeModifiers), configured.RangeProfile)
	if rangeIndex < 0 || rangeIndex >= len(spec.Tactical.Rules.BeamToHitRangeModifiers) {
		return nil, fmt.Errorf("Beam range index %d is outside the evidenced table", rangeIndex)
	}
	if rangeIndex >= len(spec.Tactical.Rules.BeamDamageRangeModifiers) {
		return nil, fmt.Errorf("Beam damage range index %d is outside the evidenced table", rangeIndex)
	}
	targetBeamDefense, err := tacticalFighterBeamInterceptorDCV(*spec.Tactical, *group)
	if err != nil {
		return nil, err
	}

	rangeHitModifier := spec.Tactical.Rules.BeamToHitRangeModifiers[rangeIndex]
	if tacticalBeamHasEffect(configured, "doubled_range_to_hit") {
		rangeHitModifier *= 2
	}
	threshold := spec.Tactical.Rules.BeamBaseHitThreshold - rangeHitModifier
	if threshold > spec.Tactical.Rules.BeamMaxHitThreshold {
		threshold = spec.Tactical.Rules.BeamMaxHitThreshold
	}
	dependencies := neutralTacticalBeamDependencyContext()
	scaledMinDamage, scaledMaxDamage, damagePercent := beam.EffectiveDamageBounds(
		configured,
		spec.Tactical.Rules.BeamDamageRangeModifiers[rangeIndex],
		dependencies.PlayerLeaderDamagePercent,
		dependencies.HighEnergyFocusBonus,
	)
	attackDifferential := tacticalBeamOffenseForState(*spec.Tactical, *active, *activeState, r.legacy17_8Disabled) - targetBeamDefense + configured.StaticToHitAdjustment
	alwaysHit := tacticalBeamHasEffect(configured, "always_hit")
	liveWeaponCount := tacticalCurrentWeaponCount(*active, *activeState, weapon.Slot)
	totalShots := liveWeaponCount * configured.ShotsPerWeapon
	hitCount := 0
	totalDamage := 0
	flatShotIndex := 0
	groupDestroyed := false
	destroyedMultiplicityBefore := 0
	targetX, targetY := group.X, group.Y
	groupID := group.GroupID

	for physicalIndex := 0; physicalIndex < liveWeaponCount; physicalIndex++ {
		for burstIndex := 0; burstIndex < configured.ShotsPerWeapon; burstIndex++ {
			flatShotIndex++
			rawRoll, err := r.random(100, spec.Tactical.Rules)
			if err != nil {
				return nil, err
			}
			hitRoll := rawRoll
			if alwaysHit {
				hitRoll = 100
			}
			effectiveRoll := hitRoll
			if hitRoll > spec.Tactical.Rules.BeamMaxHitThreshold {
				effectiveRoll = spec.Tactical.Rules.BeamEffectiveRollCap
			} else {
				effectiveRoll += attackDifferential
				if effectiveRoll > spec.Tactical.Rules.BeamEffectiveRollCap {
					effectiveRoll = spec.Tactical.Rules.BeamEffectiveRollCap
				}
			}
			hit := effectiveRoll >= threshold
			damage := 0
			if hit {
				hitCount++
				damage = beam.InterpolateDamage(scaledMinDamage, scaledMaxDamage, effectiveRoll, threshold, spec.Tactical.Rules.BeamEffectiveRollCap)
				totalDamage += damage
			}
			if err := r.appendEvent("beam_fired", seatID, commandSequence, map[string]any{
				"ship_id": active.ShipID, "target_ship_id": core.ID(0), "target_projectile_id": uint64(0), "target_fighter_group_id": groupID,
				"target_x": targetX, "target_y": targetY,
				"weapon_slot": payload.WeaponSlot, "weapon_id": weapon.WeaponID, "mount_type": canonicalTacticalWeaponMountType(weapon.MountType),
				"modifier_ids":          append([]core.ShipWeaponModifierID(nil), weapon.ModifierIDs...),
				"physical_weapon_index": physicalIndex + 1, "physical_weapon_count": liveWeaponCount,
				"burst_shot_index": burstIndex + 1, "burst_shot_count": configured.ShotsPerWeapon,
				"shot_index": flatShotIndex, "shot_count": totalShots,
				"range_index": rangeIndex, "raw_hit_roll": rawRoll, "effective_hit_roll": effectiveRoll,
				"threshold": threshold, "hit": hit, "damage": damage,
				"damage_percent": damagePercent, "scaled_min_damage": scaledMinDamage, "scaled_max_damage": scaledMaxDamage,
			}); err != nil {
				return nil, err
			}
			if hit && damage > 0 && !groupDestroyed {
				applied, err := applyTacticalFighterInterception(
					r, group, damage, "beam_manual", active.ShipID, payload.WeaponSlot, seatID, commandSequence,
				)
				if err != nil {
					return nil, err
				}
				groupDestroyed = applied.Destroyed
				if applied.Destroyed {
					destroyedMultiplicityBefore = applied.MultiplicityBefore
				}
			}
		}
	}

	ready.Ready = false
	if totalShots > 1 {
		if err := r.appendEvent("beam_volley_resolved", seatID, commandSequence, map[string]any{
			"ship_id": active.ShipID, "target_ship_id": core.ID(0), "target_projectile_id": uint64(0), "target_fighter_group_id": groupID,
			"target_x": targetX, "target_y": targetY,
			"weapon_slot": payload.WeaponSlot, "weapon_id": weapon.WeaponID, "mount_type": canonicalTacticalWeaponMountType(weapon.MountType),
			"modifier_ids":          append([]core.ShipWeaponModifierID(nil), weapon.ModifierIDs...),
			"physical_weapon_count": liveWeaponCount, "burst_shot_count": configured.ShotsPerWeapon,
			"shot_count": totalShots, "hit_count": hitCount, "total_damage": totalDamage,
		}); err != nil {
			return nil, err
		}
	}
	if groupDestroyed {
		r.fighters.Groups = append(r.fighters.Groups[:groupIndex], r.fighters.Groups[groupIndex+1:]...)
		if err := r.appendEvent("fighter_group_destroyed", seatID, commandSequence, map[string]any{
			"group_id": groupID, "reason": "intercepted", "multiplicity_before": destroyedMultiplicityBefore,
		}); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func tacticalSpecHasExecutableFighters(spec TacticalSpec) bool {
	for _, ship := range spec.Ships {
		for _, weapon := range ship.Weapons {
			if weapon.TacticalSupport == "fighter" {
				return true
			}
		}
	}
	return false
}

func tacticalFighterPayloadEqual(a, b *TacticalFighterPayloadSpec) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if a.WeaponID != b.WeaponID || a.MinDamage != b.MinDamage || a.MaxDamage != b.MaxDamage || a.PointDefenseMode != b.PointDefenseMode || len(a.InnateEffectIDs) != len(b.InnateEffectIDs) {
		return false
	}
	for i := range a.InnateEffectIDs {
		if a.InnateEffectIDs[i] != b.InnateEffectIDs[i] {
			return false
		}
	}
	return true
}

func validateTacticalFighterState(spec TacticalSpec, state TacticalFighterState, legacyDisabled bool) error {
	if state.NextGroupID == 0 {
		return fmt.Errorf("tactical Fighter next group id must be non-zero")
	}
	if legacyDisabled {
		if len(state.BayStates) != 0 || len(state.Groups) != 0 {
			return fmt.Errorf("legacy Slice-17.8 tactical snapshot must not synthesize Fighter runtime state")
		}
		return nil
	}

	wantBays := initialTacticalFighterState(spec).BayStates
	if len(state.BayStates) != len(wantBays) {
		return fmt.Errorf("tactical Fighter Bay-state count %d, expected %d", len(state.BayStates), len(wantBays))
	}
	for i := range wantBays {
		got, want := state.BayStates[i], wantBays[i]
		if got.ShipID != want.ShipID || got.WeaponSlot != want.WeaponSlot {
			return fmt.Errorf("tactical Fighter Bay-state[%d] identity mismatch", i)
		}
		carrier := tacticalShipSpec(spec, got.ShipID)
		if carrier == nil {
			return fmt.Errorf("tactical Fighter Bay-state[%d] references missing carrier", i)
		}
		weapon := weaponSpec(*carrier, got.WeaponSlot)
		if weapon == nil || weapon.TacticalSupport != "fighter" {
			return fmt.Errorf("tactical Fighter Bay-state[%d] references non-Fighter weapon", i)
		}
		if got.ReadyGroups < 0 || got.ReadyGroups > weapon.Count {
			return fmt.Errorf("tactical Fighter Bay-state[%d] ready groups %d outside 0..%d", i, got.ReadyGroups, weapon.Count)
		}
	}

	var previous uint64
	seen := make(map[uint64]struct{}, len(state.Groups))
	for i := range state.Groups {
		group := state.Groups[i]
		if group.GroupID == 0 || group.GroupID <= previous || group.GroupID >= state.NextGroupID {
			return fmt.Errorf("tactical Fighter group[%d] id/order invalid", i)
		}
		if _, ok := seen[group.GroupID]; ok {
			return fmt.Errorf("tactical Fighter group[%d] duplicates group id %d", i, group.GroupID)
		}
		seen[group.GroupID] = struct{}{}
		previous = group.GroupID

		carrier := tacticalShipSpec(spec, group.SourceShipID)
		if carrier == nil {
			return fmt.Errorf("tactical Fighter group %d references missing source ship %d", group.GroupID, group.SourceShipID)
		}
		weapon := weaponSpec(*carrier, group.SourceWeaponSlot)
		if weapon == nil || weapon.TacticalSupport != "fighter" || group.SourceSeatID != carrier.SeatID {
			return fmt.Errorf("tactical Fighter group %d source identity mismatch", group.GroupID)
		}
		kind, speed, durability, maxStock, doctrine, err := tacticalFighterStats(spec, *carrier, *weapon)
		if err != nil {
			return fmt.Errorf("tactical Fighter group %d source doctrine: %w", group.GroupID, err)
		}
		if group.Kind != kind || group.Speed != speed || group.UnitDurability != durability {
			return fmt.Errorf("tactical Fighter group %d kind/speed/durability mismatch", group.GroupID)
		}
		if group.AttackStock < 0 || group.AttackStock > maxStock {
			return fmt.Errorf("tactical Fighter group %d attack stock %d outside 0..%d", group.GroupID, group.AttackStock, maxStock)
		}
		craftPerUnit := weapon.CraftPerUnit
		if craftPerUnit <= 0 && spec.Rules.Fighter != nil {
			craftPerUnit = spec.Rules.Fighter.CraftPerBay
		}
		maxMultiplicity := weapon.Count * craftPerUnit
		if group.Multiplicity <= 0 || group.Multiplicity > maxMultiplicity {
			return fmt.Errorf("tactical Fighter group %d multiplicity %d outside 1..%d", group.GroupID, group.Multiplicity, maxMultiplicity)
		}
		if group.InterceptDamageRemainder < 0 || group.InterceptDamageRemainder >= group.UnitDurability {
			return fmt.Errorf("tactical Fighter group %d interception remainder invalid", group.GroupID)
		}
		if group.Facing < 0 || group.Facing > 15 || absInt(group.X) > TacticalCoordinateSafetyLimit || absInt(group.Y) > TacticalCoordinateSafetyLimit {
			return fmt.Errorf("tactical Fighter group %d position/facing invalid", group.GroupID)
		}
		maxMovement := group.Speed * spec.Rules.Fighter.MovementUnitsPerField
		if group.MovementRemaining < 0 || group.MovementRemaining > maxMovement || group.DistanceTraveled < 0 {
			return fmt.Errorf("tactical Fighter group %d movement state invalid", group.GroupID)
		}
		switch group.Kind {
		case "fighter":
			if !tacticalFighterPayloadEqual(group.BeamPayload, doctrine.BeamPayload) || group.BombPayload != nil {
				return fmt.Errorf("tactical Fighter group %d payload mismatch", group.GroupID)
			}
		case "bomber":
			if group.BeamPayload != nil || !tacticalFighterPayloadEqual(group.BombPayload, doctrine.BombPayload) {
				return fmt.Errorf("tactical Fighter group %d payload mismatch", group.GroupID)
			}
		case "heavy_fighter":
			if !tacticalFighterPayloadEqual(group.BeamPayload, doctrine.BeamPayload) || !tacticalFighterPayloadEqual(group.BombPayload, doctrine.BombPayload) {
				return fmt.Errorf("tactical Fighter group %d payload mismatch", group.GroupID)
			}
		default:
			return fmt.Errorf("tactical Fighter group %d has invalid kind %q", group.GroupID, group.Kind)
		}
		targetCount := 0
		if group.TargetShipID != 0 {
			targetCount++
		}
		if group.TargetProjectileID != 0 {
			targetCount++
		}
		if group.TargetFighterGroupID != 0 {
			targetCount++
		}
		if targetCount != 1 {
			return fmt.Errorf("tactical Fighter group %d requires exactly one target identity", group.GroupID)
		}
		switch group.Phase {
		case TacticalFighterPhaseOutbound:
			if group.TargetShipID != 0 {
				target := tacticalShipSpec(spec, group.TargetShipID)
				if target == nil || target.SeatID == group.SourceSeatID {
					return fmt.Errorf("tactical Fighter group %d has invalid target ship %d", group.GroupID, group.TargetShipID)
				}
			}
		case TacticalFighterPhaseReturning:
			if group.TargetShipID != group.SourceShipID || group.TargetProjectileID != 0 || group.TargetFighterGroupID != 0 {
				return fmt.Errorf("returning tactical Fighter group %d must target its source carrier", group.GroupID)
			}
		default:
			return fmt.Errorf("tactical Fighter group %d has invalid phase %q", group.GroupID, group.Phase)
		}
	}
	return nil
}
