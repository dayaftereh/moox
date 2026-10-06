package battle

import (
	"fmt"
	"sort"

	"moox/internal/beam"
	"moox/internal/core"
	"moox/internal/protocol"
)

type tacticalProjectileInterceptionResult struct {
	DestroyedUnits     int
	MultiplicityBefore int
	MultiplicityAfter  int
	RemainderBefore    int
	RemainderAfter     int
	Destroyed          bool
}

func tacticalProjectileIndexByID(r *tacticalRuntime, projectileID uint64) int {
	for i := range r.ordnance.Projectiles {
		if r.ordnance.Projectiles[i].ProjectileID == projectileID {
			return i
		}
	}
	return -1
}

func tacticalProjectileBeamInterceptorDCV(spec TacticalSpec, projectile TacticalProjectileState) (int, error) {
	rule := tacticalOrdnanceRule(spec, projectile.WeaponID)
	if rule == nil || rule.BeamInterceptorDCV <= 0 {
		return 0, fmt.Errorf("projectile %d weapon %q has no Beam interceptor DCV", projectile.ProjectileID, projectile.WeaponID)
	}
	return rule.BeamInterceptorDCV, nil
}

func tacticalBeamProjectileTargets(spec TacticalSpec, r *tacticalRuntime, activeSpec *TacticalShipSpec, activeState *TacticalShipState, weapon TacticalWeaponSpec) []TacticalFireTarget {
	if activeSpec == nil || activeState == nil || canonicalTacticalWeaponMountType(weapon.MountType) == core.ShipWeaponMountHeavy {
		return nil
	}
	configured, err := tacticalBeamConfigured(weapon)
	if err != nil {
		return nil
	}
	targetPoint := TacticalShipState{}
	out := make([]TacticalFireTarget, 0, len(r.ordnance.Projectiles))
	for _, projectile := range r.ordnance.Projectiles {
		projectileEmpire := tacticalEmpireForSeat(spec, projectile.SourceSeatID)
		if projectileEmpire == 0 || projectileEmpire == tacticalOwnerEmpire(*activeSpec, activeState) {
			continue
		}
		targetPoint.X = projectile.X
		targetPoint.Y = projectile.Y
		if !tacticalBeamTargetInArc(*activeState, targetPoint, weapon.FiringArcID) {
			continue
		}
		geometricRangeIndex := tacticalRangeIndex(activeState.X, activeState.Y, projectile.X, projectile.Y)
		rangeIndex := tacticalBeamEffectiveRangeIndex(geometricRangeIndex, len(spec.Rules.BeamToHitRangeModifiers), configured.RangeProfile)
		if rangeIndex < 0 || rangeIndex >= len(spec.Rules.BeamToHitRangeModifiers) || rangeIndex >= len(spec.Rules.BeamDamageRangeModifiers) {
			continue
		}
		out = append(out, TacticalFireTarget{TargetProjectileID: projectile.ProjectileID, RangeIndex: rangeIndex})
	}
	return out
}

func applyTacticalProjectileInterception(
	r *tacticalRuntime,
	projectile *TacticalProjectileState,
	damage int,
	sourceKind string,
	interceptorShipID core.ID,
	interceptorWeaponSlot int,
	eventSeat protocol.SeatID,
	commandSequence uint32,
) (tacticalProjectileInterceptionResult, error) {
	if projectile == nil {
		return tacticalProjectileInterceptionResult{}, fmt.Errorf("projectile interception requires projectile")
	}
	if damage < 0 {
		return tacticalProjectileInterceptionResult{}, fmt.Errorf("projectile interception damage %d is invalid", damage)
	}
	if projectile.Multiplicity <= 0 || projectile.UnitDurability <= 0 ||
		projectile.InterceptDamageRemainder < 0 || projectile.InterceptDamageRemainder >= projectile.UnitDurability {
		return tacticalProjectileInterceptionResult{}, fmt.Errorf("projectile %d interception state is invalid", projectile.ProjectileID)
	}

	result := tacticalProjectileInterceptionResult{
		MultiplicityBefore: projectile.Multiplicity,
		RemainderBefore:    projectile.InterceptDamageRemainder,
	}
	total := projectile.InterceptDamageRemainder + damage
	destroyed := total / projectile.UnitDurability
	if destroyed > projectile.Multiplicity {
		destroyed = projectile.Multiplicity
	}
	total -= destroyed * projectile.UnitDurability
	projectile.Multiplicity -= destroyed
	projectile.InterceptDamageRemainder = total
	if projectile.Multiplicity == 0 {
		projectile.InterceptDamageRemainder = 0
	}

	result.DestroyedUnits = destroyed
	result.MultiplicityAfter = projectile.Multiplicity
	result.RemainderAfter = projectile.InterceptDamageRemainder
	result.Destroyed = projectile.Multiplicity == 0

	if err := r.appendEvent("projectile_interception_resolved", eventSeat, commandSequence, map[string]any{
		"projectile_id":           projectile.ProjectileID,
		"interceptor_source_kind": sourceKind,
		"interceptor_ship_id":     interceptorShipID,
		"interceptor_weapon_slot": interceptorWeaponSlot,
		"damage":                  damage,
		"destroyed_units":         result.DestroyedUnits,
		"multiplicity_before":     result.MultiplicityBefore,
		"multiplicity_after":      result.MultiplicityAfter,
		"remainder_before":        result.RemainderBefore,
		"remainder_after":         result.RemainderAfter,
	}); err != nil {
		return tacticalProjectileInterceptionResult{}, err
	}
	return result, nil
}

func prepareFireBeamAtProjectile(spec Spec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32, payload FireBeamPayload) (*Result, error) {
	active := tacticalShipSpec(*spec.Tactical, r.state.ActiveShipID)
	if active == nil {
		return nil, fmt.Errorf("active tactical ship %d is missing from spec", r.state.ActiveShipID)
	}
	activeState := r.shipState(active.ShipID)
	if activeState == nil || seatID == 0 || tacticalControllerSeat(*active, activeState) != seatID {
		return nil, fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return nil, fmt.Errorf("fire ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	projectileIndex := tacticalProjectileIndexByID(r, payload.TargetProjectileID)
	if projectileIndex < 0 {
		return nil, fmt.Errorf("target projectile %d is not active", payload.TargetProjectileID)
	}
	projectile := &r.ordnance.Projectiles[projectileIndex]
	projectileEmpire := tacticalEmpireForSeat(*spec.Tactical, projectile.SourceSeatID)
	if projectileEmpire == 0 || projectileEmpire == tacticalOwnerEmpire(*active, activeState) {
		return nil, fmt.Errorf("target projectile %d is not an opponent", payload.TargetProjectileID)
	}

	weapon := weaponSpec(*active, payload.WeaponSlot)
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
		return nil, fmt.Errorf("Heavy Beam weapon slot %d cannot target projectile actors", payload.WeaponSlot)
	}

	configured, err := tacticalBeamConfigured(*weapon)
	if err != nil {
		return nil, err
	}
	targetPoint := TacticalShipState{X: projectile.X, Y: projectile.Y}
	if !tacticalBeamTargetInArc(*activeState, targetPoint, weapon.FiringArcID) {
		return nil, fmt.Errorf("target projectile %d is outside weapon slot %d firing arc %q", projectile.ProjectileID, payload.WeaponSlot, weapon.FiringArcID)
	}

	geometricRangeIndex := tacticalRangeIndex(activeState.X, activeState.Y, projectile.X, projectile.Y)
	rangeIndex := tacticalBeamEffectiveRangeIndex(geometricRangeIndex, len(spec.Tactical.Rules.BeamToHitRangeModifiers), configured.RangeProfile)
	if rangeIndex < 0 || rangeIndex >= len(spec.Tactical.Rules.BeamToHitRangeModifiers) {
		return nil, fmt.Errorf("Beam range index %d is outside the evidenced table", rangeIndex)
	}
	if rangeIndex >= len(spec.Tactical.Rules.BeamDamageRangeModifiers) {
		return nil, fmt.Errorf("Beam damage range index %d is outside the evidenced table", rangeIndex)
	}
	targetBeamDefense, err := tacticalProjectileBeamInterceptorDCV(*spec.Tactical, *projectile)
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
	projectileDestroyed := false

	for physicalIndex := 0; physicalIndex < liveWeaponCount; physicalIndex++ {
		for burstIndex := 0; burstIndex < configured.ShotsPerWeapon; burstIndex++ {
			flatShotIndex++
			tacticalBreakCloakOnAction(activeState)
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
				"ship_id": active.ShipID, "target_ship_id": core.ID(0), "target_projectile_id": projectile.ProjectileID,
				"target_x": projectile.X, "target_y": projectile.Y,
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
			if hit && damage > 0 && !projectileDestroyed {
				applied, err := applyTacticalProjectileInterception(
					r, projectile, damage, "beam_manual", active.ShipID, payload.WeaponSlot, seatID, commandSequence,
				)
				if err != nil {
					return nil, err
				}
				projectileDestroyed = applied.Destroyed
			}
		}
	}

	tacticalCompleteWeaponFire(*active, activeState, *weapon, ready)
	if totalShots > 1 {
		if err := r.appendEvent("beam_volley_resolved", seatID, commandSequence, map[string]any{
			"ship_id": active.ShipID, "target_ship_id": core.ID(0), "target_projectile_id": projectile.ProjectileID,
			"target_x": projectile.X, "target_y": projectile.Y,
			"weapon_slot": payload.WeaponSlot, "weapon_id": weapon.WeaponID, "mount_type": canonicalTacticalWeaponMountType(weapon.MountType),
			"modifier_ids":          append([]core.ShipWeaponModifierID(nil), weapon.ModifierIDs...),
			"physical_weapon_count": liveWeaponCount, "burst_shot_count": configured.ShotsPerWeapon,
			"shot_count": totalShots, "hit_count": hitCount, "total_damage": totalDamage,
		}); err != nil {
			return nil, err
		}
	}
	if projectileDestroyed {
		projectileID := projectile.ProjectileID
		r.ordnance.Projectiles = append(r.ordnance.Projectiles[:projectileIndex], r.ordnance.Projectiles[projectileIndex+1:]...)
		if err := r.appendEvent("projectile_destroyed", seatID, commandSequence, map[string]any{
			"projectile_id": projectileID,
			"reason":        "intercepted",
		}); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func tacticalReactivePointDefenseSlots(spec TacticalSpec, r *tacticalRuntime, target *TacticalShipSpec, projectile TacticalProjectileState) []int {
	if target == nil {
		return nil
	}
	targetState := r.shipState(target.ShipID)
	if targetState == nil || targetState.Destroyed {
		return nil
	}
	point := TacticalShipState{X: projectile.X, Y: projectile.Y}
	slots := make([]int, 0)
	for _, weapon := range target.Weapons {
		if !tacticalBeamFireSupported(weapon) || canonicalTacticalWeaponMountType(weapon.MountType) != core.ShipWeaponMountPointDefense {
			continue
		}
		configured, err := tacticalBeamConfigured(weapon)
		if err != nil || !tacticalBeamTargetInArc(*targetState, point, weapon.FiringArcID) {
			continue
		}
		geometricRangeIndex := tacticalRangeIndex(targetState.X, targetState.Y, projectile.X, projectile.Y)
		rangeIndex := tacticalBeamEffectiveRangeIndex(geometricRangeIndex, len(spec.Rules.BeamToHitRangeModifiers), configured.RangeProfile)
		if rangeIndex < 0 || rangeIndex >= len(spec.Rules.BeamToHitRangeModifiers) || rangeIndex >= len(spec.Rules.BeamDamageRangeModifiers) {
			continue
		}
		slots = append(slots, weapon.Slot)
	}
	sort.Ints(slots)
	return slots
}

func resolveReactiveProjectilePointDefense(spec TacticalSpec, r *tacticalRuntime, projectile *TacticalProjectileState, target *TacticalShipSpec) (bool, error) {
	if projectile == nil || target == nil {
		return false, nil
	}
	targetBeamDefense, err := tacticalProjectileBeamInterceptorDCV(spec, *projectile)
	if err != nil {
		return false, err
	}
	for _, slot := range tacticalReactivePointDefenseSlots(spec, r, target, *projectile) {
		if projectile.Multiplicity <= 0 {
			return true, nil
		}
		result, err := resolveReactiveBeamPointDefense(Spec{Tactical: &spec}, r, ReactiveBeamPointDefenseRequest{
			ShipID: target.ShipID, WeaponSlot: slot,
			TargetX: projectile.X, TargetY: projectile.Y,
			TargetBeamDefense: targetBeamDefense,
		})
		if err != nil {
			return false, err
		}
		for _, shot := range result.Shots {
			if projectile.Multiplicity <= 0 {
				return true, nil
			}
			if !shot.Hit || shot.Damage <= 0 {
				continue
			}
			applied, err := applyTacticalProjectileInterception(
				r, projectile, shot.Damage, "beam_reactive", target.ShipID, slot, target.SeatID, 0,
			)
			if err != nil {
				return false, err
			}
			if applied.Destroyed {
				return true, nil
			}
		}
	}
	return projectile.Multiplicity <= 0, nil
}
