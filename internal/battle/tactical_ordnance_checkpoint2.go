package battle

import (
	"fmt"
	"math"
	"sort"

	"moox/internal/core"
	"moox/internal/protocol"
)

const (
	CommandFireOrdnance = "battle.fire_ordnance"

	TacticalWeaponUnavailableAmmoExhausted  = "ammo_exhausted"
	TacticalWeaponUnavailableCooldown       = "weapon_cooldown"
	TacticalWeaponUnavailableOrdnanceEffect = "ordnance_effect_unavailable"
)

type FireOrdnancePayload struct {
	ShipID       core.ID `json:"ship_id"`
	TargetShipID core.ID `json:"target_ship_id"`
	WeaponSlot   int     `json:"weapon_slot"`
}

func NewFireOrdnanceCommand(sequence uint32, payload FireOrdnancePayload) (protocol.Command, error) {
	return protocol.NewCommand(sequence, CommandFireOrdnance, payload)
}

func tacticalOrdnanceWeaponState(state *TacticalOrdnanceState, shipID core.ID, slot int) *TacticalOrdnanceWeaponState {
	if state == nil {
		return nil
	}
	for i := range state.WeaponStates {
		if state.WeaponStates[i].ShipID == shipID && state.WeaponStates[i].WeaponSlot == slot {
			return &state.WeaponStates[i]
		}
	}
	return nil
}

func tacticalOrdnanceRule(spec TacticalSpec, weaponID string) *TacticalOrdnanceWeaponRuleSnapshot {
	if spec.Rules.Ordnance == nil {
		return nil
	}
	for i := range spec.Rules.Ordnance.Weapons {
		if spec.Rules.Ordnance.Weapons[i].WeaponID == weaponID {
			return &spec.Rules.Ordnance.Weapons[i]
		}
	}
	return nil
}

func tacticalWeaponHasModifier(weapon TacticalWeaponSpec, modifier core.ShipWeaponModifierID) bool {
	for _, id := range weapon.ModifierIDs {
		if id == modifier {
			return true
		}
	}
	return false
}

func tacticalOrdnanceEffectUnavailable(weapon TacticalWeaponSpec) bool {
	return tacticalWeaponHasModifier(weapon, core.ShipWeaponModifierEmissionsGuidance)
}

func refreshTacticalOrdnanceReadiness(spec TacticalSpec, r *tacticalRuntime) {
	if r == nil {
		return
	}
	for _, ship := range spec.Ships {
		shipState := r.shipState(ship.ShipID)
		if shipState == nil {
			continue
		}
		for _, weapon := range ship.Weapons {
			if weapon.TacticalSupport != "ordnance" {
				continue
			}
			state := weaponState(shipState, weapon.Slot)
			live := tacticalOrdnanceWeaponState(&r.ordnance, ship.ShipID, weapon.Slot)
			if state == nil || live == nil {
				continue
			}
			switch tacticalWeaponFamily(weapon) {
			case "missile":
				state.Ready = live.AmmoRemaining > 0
			case "torpedo":
				state.Ready = live.CooldownRounds == 0
			default:
				state.Ready = false
			}
		}
	}
}

func tacticalOrdnanceTargets(spec TacticalSpec, r *tacticalRuntime, source *TacticalShipSpec) []TacticalFireTarget {
	if r == nil || source == nil {
		return nil
	}
	sourceState := r.shipState(source.ShipID)
	if sourceState == nil || sourceState.Destroyed {
		return nil
	}
	sourceOwner := tacticalOwnerEmpire(*source, sourceState)
	targets := make([]TacticalFireTarget, 0, len(spec.Ships))
	for _, target := range spec.Ships {
		state := r.shipState(target.ShipID)
		if !tacticalShipTargetable(state) || tacticalOwnerEmpire(target, state) == sourceOwner {
			continue
		}
		targets = append(targets, TacticalFireTarget{
			TargetShipID: target.ShipID,
			RangeIndex:   tacticalRangeIndex(sourceState.X, sourceState.Y, state.X, state.Y),
		})
	}
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].TargetShipID < targets[j].TargetShipID
	})
	return targets
}

func tacticalOrdnanceWeaponAction(spec TacticalSpec, r *tacticalRuntime, activeSpec *TacticalShipSpec, activeState *TacticalShipState, weaponSpec TacticalWeaponSpec) TacticalWeaponActionAvailability {
	action := TacticalWeaponActionAvailability{
		ShipID:       activeSpec.ShipID,
		WeaponSlot:   weaponSpec.Slot,
		WeaponID:     weaponSpec.WeaponID,
		WeaponFamily: tacticalWeaponFamily(weaponSpec),
		CommandKind:  CommandFireOrdnance,
	}
	if activeState == nil || weaponSpec.TacticalSupport != "ordnance" || (action.WeaponFamily != "missile" && action.WeaponFamily != "torpedo") {
		action.CommandKind = ""
		action.UnavailableReason = TacticalWeaponUnavailableConsumer
		return action
	}
	if tacticalOrdnanceEffectUnavailable(weaponSpec) {
		action.UnavailableReason = TacticalWeaponUnavailableOrdnanceEffect
		return action
	}
	ready := weaponState(activeState, weaponSpec.Slot)
	live := tacticalOrdnanceWeaponState(&r.ordnance, activeSpec.ShipID, weaponSpec.Slot)
	if ready == nil || live == nil || tacticalOrdnanceRule(spec, weaponSpec.WeaponID) == nil {
		action.CommandKind = ""
		action.UnavailableReason = TacticalWeaponUnavailableConsumer
		return action
	}
	switch action.WeaponFamily {
	case "missile":
		if live.AmmoRemaining <= 0 {
			action.UnavailableReason = TacticalWeaponUnavailableAmmoExhausted
			return action
		}
	case "torpedo":
		if live.CooldownRounds > 0 {
			action.UnavailableReason = TacticalWeaponUnavailableCooldown
			return action
		}
	}
	if !ready.Ready {
		action.UnavailableReason = TacticalWeaponUnavailableNotReady
		return action
	}
	action.Targets = tacticalOrdnanceTargets(spec, r, activeSpec)
	if len(action.Targets) == 0 {
		action.UnavailableReason = TacticalWeaponUnavailableNoTargets
		return action
	}
	action.Available = true
	return action
}

func prepareFireOrdnance(spec Spec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32, payload FireOrdnancePayload) error {
	active := tacticalShipSpec(*spec.Tactical, r.state.ActiveShipID)
	activeState := r.shipState(r.state.ActiveShipID)
	if active == nil || activeState == nil {
		return fmt.Errorf("active tactical ship %d is missing", r.state.ActiveShipID)
	}
	if seatID == 0 || tacticalControllerSeat(*active, activeState) != seatID {
		return fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return fmt.Errorf("fire ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	weapon := weaponSpec(*active, payload.WeaponSlot)
	if weapon == nil {
		return fmt.Errorf("ship %d has no weapon slot %d", active.ShipID, payload.WeaponSlot)
	}
	action := tacticalOrdnanceWeaponAction(*spec.Tactical, r, active, activeState, *weapon)
	if !action.Available {
		if action.UnavailableReason == "" {
			action.UnavailableReason = TacticalWeaponUnavailableConsumer
		}
		return fmt.Errorf("ship %d weapon slot %d cannot fire ordnance: %s", active.ShipID, payload.WeaponSlot, action.UnavailableReason)
	}
	targetAllowed := false
	for _, target := range action.Targets {
		if target.TargetShipID == payload.TargetShipID {
			targetAllowed = true
			break
		}
	}
	if !targetAllowed {
		return fmt.Errorf("target ship %d is not a legal ordnance target", payload.TargetShipID)
	}
	targetSpec := tacticalShipSpec(*spec.Tactical, payload.TargetShipID)
	targetState := r.shipState(payload.TargetShipID)
	if targetSpec == nil || !tacticalShipTargetable(targetState) {
		return fmt.Errorf("target ship %d is not a live tactical ship", payload.TargetShipID)
	}
	configured, err := deriveConfiguredTacticalProjectile(*spec.Tactical, *weapon)
	if err != nil {
		return err
	}
	live := tacticalOrdnanceWeaponState(&r.ordnance, active.ShipID, weapon.Slot)
	ready := weaponState(activeState, weapon.Slot)
	if live == nil || ready == nil {
		return fmt.Errorf("ordnance runtime state missing for ship %d slot %d", active.ShipID, weapon.Slot)
	}

	liveWeaponCount := tacticalCurrentWeaponCount(*active, *activeState, weapon.Slot)
	if liveWeaponCount <= 0 {
		return fmt.Errorf("ordnance weapon slot %d has no live physical weapons", weapon.Slot)
	}
	if destroyedByAbsorber, err := tacticalResolveEnergyAbsorberPreWeapon(*spec.Tactical, r, *active, activeState, *targetSpec, targetState, seatID, commandSequence); err != nil {
		return err
	} else if destroyedByAbsorber {
		tacticalBreakCloakOnAction(activeState)
		return nil
	}
	facing := activeState.Facing
	if targetState.X != activeState.X || targetState.Y != activeState.Y {
		facing = tacticalFacing(targetState.X-activeState.X, targetState.Y-activeState.Y)
	}
	projectileID := r.ordnance.NextProjectileID
	spawnOrder := r.nextMovingActorOrder
	tacticalBreakCloakOnAction(activeState)
	projectile := TacticalProjectileState{
		ProjectileID:      projectileID,
		SpawnOrder:        spawnOrder,
		SourceSeatID:      tacticalControllerSeat(*active, activeState),
		SourceShipID:      active.ShipID,
		WeaponSlot:        weapon.Slot,
		TargetShipID:      payload.TargetShipID,
		WeaponID:          weapon.WeaponID,
		WeaponFamily:      tacticalWeaponFamily(*weapon),
		ModifierIDs:       append([]core.ShipWeaponModifierID(nil), weapon.ModifierIDs...),
		X:                 activeState.X,
		Y:                 activeState.Y,
		Facing:            facing,
		Speed:             configured.Speed,
		MovementRemaining: configured.Speed * spec.Tactical.Rules.Ordnance.MovementUnitsPerField,
		Multiplicity:      liveWeaponCount,
		UnitDurability:    configured.UnitDurability,
		WarheadStrength:   configured.WarheadStrength,
	}

	switch projectile.WeaponFamily {
	case "missile":
		live.AmmoRemaining--
	case "torpedo":
		live.CooldownRounds = spec.Tactical.Rules.Ordnance.TorpedoCooldownRounds
	default:
		return fmt.Errorf("weapon slot %d is outside the executable ordnance contract", weapon.Slot)
	}
	ready.Ready = false
	r.ordnance.NextProjectileID++
	r.nextMovingActorOrder++
	r.ordnance.Projectiles = append(r.ordnance.Projectiles, projectile)

	launchData := map[string]any{
		"projectile_id":        projectile.ProjectileID,
		"source_seat_id":       projectile.SourceSeatID,
		"source_ship_id":       projectile.SourceShipID,
		"weapon_slot":          projectile.WeaponSlot,
		"target_ship_id":       projectile.TargetShipID,
		"weapon_id":            projectile.WeaponID,
		"weapon_family":        projectile.WeaponFamily,
		"modifier_ids":         projectile.ModifierIDs,
		"start_x":              projectile.X,
		"start_y":              projectile.Y,
		"start_facing":         projectile.Facing,
		"speed":                projectile.Speed,
		"multiplicity":         projectile.Multiplicity,
		"unit_durability":      projectile.UnitDurability,
		"beam_interceptor_dcv": configured.BeamInterceptorDCV,
		"warhead_strength":     projectile.WarheadStrength,
	}
	if projectile.WeaponFamily == "missile" {
		launchData["ammo_remaining"] = live.AmmoRemaining
	} else {
		launchData["cooldown_rounds"] = live.CooldownRounds
	}
	if err := r.appendEvent("ordnance_launched", seatID, commandSequence, launchData); err != nil {
		return err
	}
	return processTacticalProjectileByID(*spec.Tactical, r, projectileID, seatID, commandSequence)
}

func advanceTacticalOrdnanceRound(spec TacticalSpec, r *tacticalRuntime) error {
	for i := range r.ordnance.WeaponStates {
		state := &r.ordnance.WeaponStates[i]
		_, weapon := ordnanceWeaponSpec(spec, state.ShipID, state.WeaponSlot)
		if weapon == nil || tacticalWeaponFamily(*weapon) != "torpedo" {
			continue
		}
		if state.CooldownRounds > 0 {
			state.CooldownRounds--
		}
	}
	refreshTacticalOrdnanceReadiness(spec, r)
	refreshTacticalFighterReadiness(spec, r)
	return processTacticalMovingActors(spec, r, 0, 0)
}

func processTacticalProjectiles(spec TacticalSpec, r *tacticalRuntime, eventSeat protocol.SeatID, commandSequence uint32) error {
	sort.Slice(r.ordnance.Projectiles, func(i, j int) bool {
		return r.ordnance.Projectiles[i].ProjectileID < r.ordnance.Projectiles[j].ProjectileID
	})
	for i := 0; i < len(r.ordnance.Projectiles); {
		projectileID := r.ordnance.Projectiles[i].ProjectileID
		outcome, err := moveTacticalProjectileQuantum(spec, r, i, eventSeat, commandSequence)
		if err != nil {
			return err
		}
		if outcome.Keep {
			i++
			continue
		}
		r.ordnance.Projectiles = append(r.ordnance.Projectiles[:i], r.ordnance.Projectiles[i+1:]...)
		reason := outcome.DestroyReason
		if reason == "" {
			reason = "no_target"
		}
		if err := r.appendEvent("projectile_destroyed", eventSeat, commandSequence, map[string]any{
			"projectile_id": projectileID,
			"reason":        reason,
		}); err != nil {
			return err
		}
	}
	return nil
}

func processTacticalProjectileByID(spec TacticalSpec, r *tacticalRuntime, projectileID uint64, eventSeat protocol.SeatID, commandSequence uint32) error {
	for i := range r.ordnance.Projectiles {
		if r.ordnance.Projectiles[i].ProjectileID != projectileID {
			continue
		}
		outcome, err := moveTacticalProjectileQuantum(spec, r, i, eventSeat, commandSequence)
		if err != nil {
			return err
		}
		if outcome.Keep {
			return nil
		}
		r.ordnance.Projectiles = append(r.ordnance.Projectiles[:i], r.ordnance.Projectiles[i+1:]...)
		reason := outcome.DestroyReason
		if reason == "" {
			reason = "no_target"
		}
		return r.appendEvent("projectile_destroyed", eventSeat, commandSequence, map[string]any{
			"projectile_id": projectileID,
			"reason":        reason,
		})
	}
	return fmt.Errorf("tactical projectile %d is missing", projectileID)
}

func moveTacticalProjectileQuantum(spec TacticalSpec, r *tacticalRuntime, index int, eventSeat protocol.SeatID, commandSequence uint32) (tacticalProjectileQuantumOutcome, error) {
	if index < 0 || index >= len(r.ordnance.Projectiles) {
		return tacticalProjectileQuantumOutcome{}, fmt.Errorf("tactical projectile index %d is out of range", index)
	}
	projectile := &r.ordnance.Projectiles[index]
	source := tacticalShipSpec(spec, projectile.SourceShipID)
	if source == nil {
		return tacticalProjectileQuantumOutcome{}, fmt.Errorf("tactical projectile %d source ship is missing", projectile.ProjectileID)
	}
	sourceEmpire := tacticalEmpireForSeat(spec, projectile.SourceSeatID)
	if sourceEmpire == 0 {
		return tacticalProjectileQuantumOutcome{}, fmt.Errorf("tactical projectile %d source seat %d is invalid", projectile.ProjectileID, projectile.SourceSeatID)
	}
	targetState := r.shipState(projectile.TargetShipID)
	targetSpec := tacticalShipSpec(spec, projectile.TargetShipID)
	if targetSpec == nil || !tacticalShipTargetable(targetState) || tacticalOwnerEmpire(*targetSpec, targetState) == sourceEmpire {
		targets := tacticalOrdnanceRetargetIDs(spec, r, sourceEmpire)
		if len(targets) == 0 {
			return tacticalProjectileQuantumOutcome{DestroyReason: "no_target"}, nil
		}
		previous := projectile.TargetShipID
		projectile.TargetShipID = targets[0]
		targetState = r.shipState(projectile.TargetShipID)
		if err := r.appendEvent("projectile_retargeted", eventSeat, commandSequence, map[string]any{
			"projectile_id":           projectile.ProjectileID,
			"previous_target_ship_id": previous,
			"new_target_ship_id":      projectile.TargetShipID,
			"reason":                  "target_invalid",
		}); err != nil {
			return tacticalProjectileQuantumOutcome{}, err
		}
	}

	if spec.Rules.Ordnance == nil {
		return tacticalProjectileQuantumOutcome{}, fmt.Errorf("tactical projectile %d has no ordnance rules", projectile.ProjectileID)
	}
	budget := projectile.Speed * spec.Rules.Ordnance.MovementUnitsPerField
	projectile.MovementRemaining = budget
	if targetState.X == projectile.X && targetState.Y == projectile.Y {
		projectile.MovementRemaining = 0
		return resolveTacticalProjectileImpact(spec, r, projectile, commandSequence)
	}
	fromX, fromY := projectile.X, projectile.Y
	toX, toY, movedFields := tacticalProjectileDestination(fromX, fromY, targetState.X, targetState.Y, projectile.Speed)
	if movedFields <= 0 {
		return tacticalProjectileQuantumOutcome{}, fmt.Errorf("tactical projectile %d could not advance toward target %d", projectile.ProjectileID, projectile.TargetShipID)
	}
	movementUsed := movedFields * spec.Rules.Ordnance.MovementUnitsPerField
	if movementUsed > projectile.MovementRemaining {
		return tacticalProjectileQuantumOutcome{}, fmt.Errorf("tactical projectile %d movement %d exceeds budget %d", projectile.ProjectileID, movementUsed, projectile.MovementRemaining)
	}
	projectile.X = toX
	projectile.Y = toY
	projectile.Facing = tacticalFacing(toX-fromX, toY-fromY)
	projectile.MovementRemaining -= movementUsed
	projectile.DistanceTraveled += movementUsed

	if err := r.appendEvent("projectile_moved", eventSeat, commandSequence, map[string]any{
		"projectile_id":      projectile.ProjectileID,
		"from_x":             fromX,
		"from_y":             fromY,
		"to_x":               projectile.X,
		"to_y":               projectile.Y,
		"movement_used":      movementUsed,
		"movement_remaining": projectile.MovementRemaining,
		"distance_traveled":  projectile.DistanceTraveled,
		"target_ship_id":     projectile.TargetShipID,
	}); err != nil {
		return tacticalProjectileQuantumOutcome{}, err
	}
	if projectile.X == targetState.X && projectile.Y == targetState.Y {
		return resolveTacticalProjectileImpact(spec, r, projectile, commandSequence)
	}
	return tacticalProjectileQuantumOutcome{Keep: true}, nil
}

func tacticalOrdnanceRetargetIDs(spec TacticalSpec, r *tacticalRuntime, sourceEmpireID core.ID) []core.ID {
	ids := make([]core.ID, 0, len(spec.Ships))
	for _, ship := range spec.Ships {
		state := r.shipState(ship.ShipID)
		if state == nil || state.Destroyed || tacticalOwnerEmpire(ship, state) == sourceEmpireID {
			continue
		}
		ids = append(ids, ship.ShipID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func tacticalProjectileDestination(fromX, fromY, targetX, targetY, maxFields int) (int, int, int) {
	if maxFields <= 0 {
		return fromX, fromY, 0
	}
	dx, dy := targetX-fromX, targetY-fromY
	targetDistance := tacticalProjectileFieldDistance(dx, dy)
	if targetDistance <= maxFields {
		return targetX, targetY, targetDistance
	}

	bestX, bestY := fromX, fromY
	bestMove := 0
	bestRemaining := int64(dx)*int64(dx) + int64(dy)*int64(dy)
	bestCross := int64(math.MaxInt64)
	for oy := -maxFields; oy <= maxFields; oy++ {
		for ox := -maxFields; ox <= maxFields; ox++ {
			if ox == 0 && oy == 0 {
				continue
			}
			move := tacticalProjectileFieldDistance(ox, oy)
			if move <= 0 || move > maxFields {
				continue
			}
			rx, ry := targetX-(fromX+ox), targetY-(fromY+oy)
			remaining := int64(rx)*int64(rx) + int64(ry)*int64(ry)
			cross := int64(ox)*int64(dy) - int64(oy)*int64(dx)
			if cross < 0 {
				cross = -cross
			}
			if move > bestMove ||
				(move == bestMove && cross < bestCross) ||
				(move == bestMove && cross == bestCross && remaining < bestRemaining) ||
				(move == bestMove && cross == bestCross && remaining == bestRemaining && (fromY+oy < bestY || (fromY+oy == bestY && fromX+ox < bestX))) {
				bestX, bestY = fromX+ox, fromY+oy
				bestMove = move
				bestRemaining = remaining
				bestCross = cross
			}
		}
	}
	return bestX, bestY, bestMove
}

func tacticalProjectileFieldDistance(dx, dy int) int {
	if dx == 0 && dy == 0 {
		return 0
	}
	distanceSquared := float64(int64(dx)*int64(dx) + int64(dy)*int64(dy))
	return int(math.Ceil(math.Sqrt(distanceSquared)))
}
