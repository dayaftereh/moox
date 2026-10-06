package battle

import (
	"fmt"

	"moox/internal/core"
)

const (
	TacticalCloakStateNone            = "none"
	TacticalCloakStateCloaked         = "cloaked"
	TacticalCloakStateBroken          = "cloak_broken"
	TacticalCloakStateRecovering      = "cloak_recovering"
	TacticalCloakStatePhased          = "phased"
	TacticalCloakStatePhaseBroken     = "phase_broken"
	TacticalCloakStatePhaseRecovering = "phase_recovering"

	TacticalSpecialReadinessNone             = "none"
	TacticalSpecialReadinessExtraVolleyReady = "extra_volley_ready"
	TacticalSpecialReadinessRearmPending     = "rearm_pending"

	TacticalMovementKindNormal   = "normal"
	TacticalMovementKindTeleport = "teleport"
)

func tacticalInitialCloakState(spec TacticalSpec, ship TacticalShipSpec, state TacticalShipState) (string, int) {
	if TacticalSpecialFunctional(ship, state, false, "phasing_cloak") {
		lifetime := 10
		if spec.Rules.Specials != nil && spec.Rules.Specials.PhasingInitialLifetime > 0 {
			lifetime = spec.Rules.Specials.PhasingInitialLifetime
		}
		return TacticalCloakStatePhased, lifetime
	}
	if TacticalSpecialFunctional(ship, state, false, "cloaking_device") {
		return TacticalCloakStateCloaked, 0
	}
	return TacticalCloakStateNone, 0
}

func tacticalWeaponEligibleForSpecialReadiness(ship TacticalShipSpec, state TacticalShipState, weapon TacticalWeaponSpec) bool {
	if tacticalBeamFireSupported(weapon) && TacticalSpecialFunctional(ship, state, false, "hyper_x_capacitors") {
		return true
	}
	return tacticalWeaponFamily(weapon) == "missile" && TacticalSpecialFunctional(ship, state, false, "fast_missile_racks")
}

func tacticalInitialSpecialReadiness(ship TacticalShipSpec, state TacticalShipState, weapon TacticalWeaponSpec) string {
	if tacticalWeaponEligibleForSpecialReadiness(ship, state, weapon) {
		return TacticalSpecialReadinessExtraVolleyReady
	}
	return TacticalSpecialReadinessNone
}

func tacticalSelfTeleportFunctional(spec TacticalSpec, r *tacticalRuntime, ship TacticalShipSpec, state TacticalShipState) bool {
	if r == nil || r.legacy17_8Disabled || state.Destroyed || state.ActivationComplete || state.TeleporterUsed || spec.Rules.Specials == nil {
		return false
	}
	return TacticalSpecialFunctional(ship, state, false, "phase_shifter") ||
		TacticalSpecialFunctional(ship, state, false, "sub_space_teleporter")
}

func tacticalTeleportMoveOption(spec TacticalSpec, r *tacticalRuntime, ship TacticalShipSpec, state TacticalShipState, x, y int) (TacticalMoveOption, error) {
	if !tacticalSelfTeleportFunctional(spec, r, ship, state) {
		return TacticalMoveOption{}, fmt.Errorf("ship %d has no available self-teleport", state.ShipID)
	}
	if x == state.X && y == state.Y {
		return TacticalMoveOption{}, fmt.Errorf("same-coordinate Tactical move is not a teleport")
	}
	if absInt(x) > TacticalCoordinateSafetyLimit || absInt(y) > TacticalCoordinateSafetyLimit {
		return TacticalMoveOption{}, fmt.Errorf("Tactical destination (%d,%d) exceeds technical coordinate safety envelope", x, y)
	}
	distance := tacticalRangeIndex(state.X, state.Y, x, y)
	if distance > spec.Rules.Specials.SelfTeleportMaxRange {
		return TacticalMoveOption{}, fmt.Errorf("Tactical teleport destination (%d,%d) exceeds range %d", x, y, spec.Rules.Specials.SelfTeleportMaxRange)
	}
	for _, other := range r.state.Ships {
		if other.ShipID != state.ShipID && !other.Destroyed && other.X == x && other.Y == y {
			return TacticalMoveOption{}, fmt.Errorf("Tactical destination (%d,%d) is occupied by ship %d", x, y, other.ShipID)
		}
	}
	return TacticalMoveOption{
		X:                      x,
		Y:                      y,
		MoveCost:               0,
		ResultingFacing:        state.Facing,
		MovementRemainingAfter: state.MovementCurrent,
		MovementKind:           TacticalMovementKindTeleport,
	}, nil
}

func tacticalSpecialStateValid(state TacticalShipState) bool {
	switch state.CloakState {
	case TacticalCloakStateNone, TacticalCloakStateCloaked, TacticalCloakStateBroken,
		TacticalCloakStateRecovering, TacticalCloakStatePhased, TacticalCloakStatePhaseBroken,
		TacticalCloakStatePhaseRecovering:
	default:
		return false
	}
	if state.PhasingTurnsRemaining < 0 || state.EnergyAbsorberStored < 0 {
		return false
	}
	if state.CloakState != TacticalCloakStatePhased && state.CloakState != TacticalCloakStatePhaseBroken && state.CloakState != TacticalCloakStatePhaseRecovering {
		return state.PhasingTurnsRemaining == 0
	}
	return true
}

func tacticalSpecialReadinessValid(value string) bool {
	switch value {
	case TacticalSpecialReadinessNone, TacticalSpecialReadinessExtraVolleyReady, TacticalSpecialReadinessRearmPending:
		return true
	default:
		return false
	}
}

func rechargeTacticalShieldArcsWithBudget(arcs *[4]int, effectiveMax, baseMax, factor int) {
	if arcs == nil || effectiveMax <= 0 || baseMax <= 0 || factor <= 0 {
		return
	}
	budget := baseMax * factor / 10
	for budget > 0 {
		changed := false
		for arc := 0; arc < len(arcs) && budget > 0; arc++ {
			if arcs[arc] < 0 {
				arcs[arc] = 0
			}
			if arcs[arc] >= effectiveMax {
				continue
			}
			arcs[arc]++
			budget--
			changed = true
		}
		if !changed {
			break
		}
	}
}

func tacticalSpecialHasID(ship TacticalShipSpec, id string) bool {
	for _, special := range ship.Specials {
		if special.SpecialID == id {
			return true
		}
	}
	return false
}

func tacticalSpecialShipByID(spec TacticalSpec, shipID core.ID) *TacticalShipSpec {
	return tacticalShipSpec(spec, shipID)
}

func applyTacticalDamageWithSpecials(spec TacticalSpec, r *tacticalRuntime, targetSpec TacticalShipSpec, targetState *TacticalShipState, sourceState TacticalShipState, incomingDamage int, options tacticalDamageOptions) tacticalDamageApplication {
	originalDamage := incomingDamage
	adjustedDamage := incomingDamage
	absorbed := 0
	damperReduction := 0
	if r != nil && targetState != nil && adjustedDamage > 0 && !r.legacy17_8Disabled && spec.Rules.Specials != nil {
		rules := spec.Rules.Specials
		if TacticalSpecialFunctional(targetSpec, *targetState, false, "energy_absorber") && rules.EnergyAbsorberDivisor > 0 {
			absorbed = adjustedDamage / rules.EnergyAbsorberDivisor
			if absorbed < 1 {
				absorbed = 1
			}
			if absorbed > adjustedDamage {
				absorbed = adjustedDamage
			}
			targetState.EnergyAbsorberStored += absorbed
			adjustedDamage -= absorbed
		}
		if adjustedDamage > 0 && TacticalSpecialFunctional(targetSpec, *targetState, false, "damper_field") && rules.DamperDamageDivisor > 0 {
			before := adjustedDamage
			adjustedDamage /= rules.DamperDamageDivisor
			damperReduction = before - adjustedDamage
			options.MarineKillSuppressed = true
		}
	}
	result := applyTacticalDamageWithOptions(targetSpec, targetState, sourceState, adjustedDamage, options)
	result.IncomingDamage = originalDamage
	result.PostSpecialDamage = adjustedDamage
	result.EnergyAbsorbed = absorbed
	result.DamperReducedDamage = damperReduction
	return result
}

func tacticalResolveBeamReflection(spec TacticalSpec, r *tacticalRuntime, targetSpec TacticalShipSpec, targetState TacticalShipState, damage int) (remainingDamage, reflectedDamage, roll int, err error) {
	if r == nil || damage <= 0 || r.legacy17_8Disabled || spec.Rules.Specials == nil ||
		!TacticalSpecialFunctional(targetSpec, targetState, false, "reflection_field") {
		return damage, 0, 0, nil
	}
	roll, err = r.random(uint32(damage+15), spec.Rules)
	if err != nil {
		return 0, 0, 0, err
	}
	if roll > damage {
		return 0, damage, roll, nil
	}
	return damage, 0, roll, nil
}

func tacticalBeamForcedMiss(spec TacticalSpec, r *tacticalRuntime, targetSpec TacticalShipSpec, targetState TacticalShipState) (bool, int, error) {
	if r == nil || r.legacy17_8Disabled || spec.Rules.Specials == nil ||
		!TacticalSpecialFunctional(targetSpec, targetState, false, "displacement_device") {
		return false, 0, nil
	}
	roll, err := r.random(100, spec.Rules)
	if err != nil {
		return false, 0, err
	}
	return roll <= spec.Rules.Specials.DisplacementBeamMissThreshold, roll, nil
}

func tacticalOrdnanceForcedMiss(spec TacticalSpec, r *tacticalRuntime, targetSpec TacticalShipSpec, targetState TacticalShipState) (forced bool, displacementRoll, cloakRoll int, err error) {
	if r == nil || r.legacy17_8Disabled || spec.Rules.Specials == nil {
		return false, 0, 0, nil
	}
	if TacticalSpecialFunctional(targetSpec, targetState, false, "displacement_device") {
		displacementRoll, err = r.random(100, spec.Rules)
		if err != nil {
			return false, 0, 0, err
		}
		if displacementRoll < spec.Rules.Specials.DisplacementOrdnanceMissThreshold {
			return true, displacementRoll, 0, nil
		}
	}
	if targetState.CloakState == TacticalCloakStateCloaked {
		cloakRoll, err = r.random(100, spec.Rules)
		if err != nil {
			return false, displacementRoll, 0, err
		}
		if cloakRoll < spec.Rules.Specials.CloakMissileForcedMissThreshold {
			return true, displacementRoll, cloakRoll, nil
		}
	}
	return false, displacementRoll, cloakRoll, nil
}

func tacticalShipTargetable(state *TacticalShipState) bool {
	return state != nil && !state.Destroyed && state.CloakState != TacticalCloakStatePhased
}

func tacticalBreakCloakOnAction(state *TacticalShipState) {
	if state == nil {
		return
	}
	switch state.CloakState {
	case TacticalCloakStateCloaked, TacticalCloakStateRecovering:
		state.CloakState = TacticalCloakStateBroken
	case TacticalCloakStatePhased, TacticalCloakStatePhaseRecovering:
		state.CloakState = TacticalCloakStatePhaseBroken
	}
}

func tacticalStartShipTurnSpecials(spec TacticalSpec, ship TacticalShipSpec, state *TacticalShipState) {
	if state == nil {
		return
	}
	state.TeleporterUsed = false
	switch state.CloakState {
	case TacticalCloakStateBroken:
		state.CloakState = TacticalCloakStateRecovering
	case TacticalCloakStatePhaseBroken:
		state.CloakState = TacticalCloakStatePhaseRecovering
	}
	for i := range state.Weapons {
		weapon := weaponSpec(ship, state.Weapons[i].Slot)
		if weapon != nil && tacticalWeaponEligibleForSpecialReadiness(ship, *state, *weapon) &&
			state.Weapons[i].SpecialReadiness != TacticalSpecialReadinessExtraVolleyReady {
			state.Weapons[i].SpecialReadiness = TacticalSpecialReadinessRearmPending
		}
	}
}

func tacticalConsumeSpecialReadinessOnFire(state *TacticalWeaponState) bool {
	if state == nil {
		return false
	}
	switch state.SpecialReadiness {
	case TacticalSpecialReadinessExtraVolleyReady:
		state.SpecialReadiness = TacticalSpecialReadinessNone
		return true
	case TacticalSpecialReadinessRearmPending:
		state.SpecialReadiness = TacticalSpecialReadinessNone
	}
	return false
}
