package battle

import (
	"fmt"

	"moox/internal/core"
	"moox/internal/protocol"
)

const (
	tacticalExplosionPixelsPerField      = 20
	tacticalQuantumExplosionRadiusPixels = 135
)

func tacticalHullExplosionProfile(hullID string) (sizeIndex int, explosionValue int, ok bool) {
	switch hullID {
	case "frigate":
		return 0, 2, true
	case "destroyer":
		return 1, 5, true
	case "cruiser":
		return 2, 10, true
	case "battleship":
		return 3, 15, true
	case "titan":
		return 4, 20, true
	case "doom_star":
		return 5, 40, true
	default:
		return 0, 0, false
	}
}

func tacticalQuantumDeathExplosionEnabled(ship TacticalShipSpec, state TacticalShipState, legacy17_8Disabled bool) bool {
	if legacy17_8Disabled {
		return false
	}
	probe := state
	probe.Destroyed = false
	return TacticalSpecialFunctional(ship, probe, false, "quantum_detonator")
}

func tacticalShipDestructionEventData(spec TacticalSpec, r *tacticalRuntime, ship TacticalShipSpec, reason string) (map[string]any, bool, int, error) {
	data := map[string]any{
		"ship_id":          ship.ShipID,
		"explosion_visual": true,
	}
	if reason != "" {
		data["reason"] = reason
	}
	if !tacticalQuantumDeathExplosionEnabled(ship, *r.shipState(ship.ShipID), r.legacy17_8Disabled) {
		return data, false, 0, nil
	}
	if spec.Rules.Specials == nil || spec.Rules.Specials.QuantumExplosionMultiplier <= 0 {
		return nil, false, 0, fmt.Errorf("Quantum Detonator ship %d lacks explosion multiplier", ship.ShipID)
	}
	_, hullExplosionValue, ok := tacticalHullExplosionProfile(ship.HullID)
	if !ok {
		return nil, false, 0, fmt.Errorf("Quantum Detonator ship %d has unsupported hull %q", ship.ShipID, ship.HullID)
	}
	power := 5 * hullExplosionValue * spec.Rules.Specials.QuantumExplosionMultiplier
	data["quantum_detonator"] = true
	data["explosion_power"] = power
	data["explosion_radius_pixels"] = tacticalQuantumExplosionRadiusPixels
	data["explosion_pixels_per_field"] = tacticalExplosionPixelsPerField
	return data, true, power, nil
}

type tacticalQuantumExplosionTarget struct {
	shipID                core.ID
	distanceSquaredPixels int
}

func tacticalQuantumExplosionTargets(spec TacticalSpec, r *tacticalRuntime, source TacticalShipSpec, sourceState TacticalShipState) []tacticalQuantumExplosionTarget {
	radiusSquared := tacticalQuantumExplosionRadiusPixels * tacticalQuantumExplosionRadiusPixels
	out := make([]tacticalQuantumExplosionTarget, 0)
	for _, candidate := range spec.Ships {
		if candidate.ShipID == source.ShipID {
			continue
		}
		state := r.shipState(candidate.ShipID)
		if state == nil || state.Destroyed {
			continue
		}
		dx := (state.X - sourceState.X) * tacticalExplosionPixelsPerField
		dy := (state.Y - sourceState.Y) * tacticalExplosionPixelsPerField
		distanceSquared := dx*dx + dy*dy
		if distanceSquared > radiusSquared {
			continue
		}
		out = append(out, tacticalQuantumExplosionTarget{shipID: candidate.ShipID, distanceSquaredPixels: distanceSquared})
	}
	return out
}

func tacticalQuantumExplosionRawBlast(power, distanceSquaredPixels int) int {
	radiusSquared := tacticalQuantumExplosionRadiusPixels * tacticalQuantumExplosionRadiusPixels
	distancePercent := (distanceSquaredPixels * 100) / radiusSquared
	remainingPercent := 100 - distancePercent
	if remainingPercent < 0 {
		remainingPercent = 0
	}
	blast := (power * remainingPercent) / 100
	if blast < 1 {
		blast = 1
	}
	return blast
}

func tacticalQuantumExplosionIncomingDamage(r *tacticalRuntime, rules TacticalRulesSnapshot, target TacticalShipSpec, rawBlast int) (int, int, error) {
	sizeIndex, _, ok := tacticalHullExplosionProfile(target.HullID)
	if !ok {
		return 0, 0, fmt.Errorf("ship %d has unsupported hull %q for Quantum explosion", target.ShipID, target.HullID)
	}
	if rawBlast <= 0 {
		return 0, 0, fmt.Errorf("Quantum explosion raw blast must be positive")
	}
	rollCount := sizeIndex + 1
	total := 0
	for i := 0; i < rollCount; i++ {
		first, err := r.random(uint32(rawBlast), rules)
		if err != nil {
			return 0, 0, err
		}
		damage := 1
		if first > 1 {
			damage, err = r.random(uint32(rawBlast), rules)
			if err != nil {
				return 0, 0, err
			}
		}
		total += damage
	}
	return total, rollCount, nil
}

func tacticalResolveQuantumDeathExplosion(spec TacticalSpec, r *tacticalRuntime, source TacticalShipSpec, sourceState TacticalShipState, eventSeat protocol.SeatID, commandSequence uint32, power int) error {
	targets := tacticalQuantumExplosionTargets(spec, r, source, sourceState)
	affected := make([]core.ID, 0, len(targets))
	newlyDestroyed := make([]core.ID, 0)

	for _, candidate := range targets {
		targetSpec := tacticalShipSpec(spec, candidate.shipID)
		targetState := r.shipState(candidate.shipID)
		if targetSpec == nil || targetState == nil || targetState.Destroyed {
			continue
		}
		rawBlast := tacticalQuantumExplosionRawBlast(power, candidate.distanceSquaredPixels)
		incomingDamage, rollCount, err := tacticalQuantumExplosionIncomingDamage(r, spec.Rules, *targetSpec, rawBlast)
		if err != nil {
			return err
		}
		wasDestroyed := targetState.Destroyed
		options := tacticalDamageOptions{}
		if spec.Rules.Specials != nil && TacticalSpecialFunctional(*targetSpec, *targetState, r.legacy17_8Disabled, "hard_shields") {
			options.ShieldAdditionalBlock = spec.Rules.Specials.HardShieldsAdditionalAbsorption
		}
		applied := applyTacticalDamageWithSpecials(spec, r, *targetSpec, targetState, sourceState, incomingDamage, options)
		if err := r.appendEvent("battle_damage_applied", eventSeat, commandSequence, map[string]any{
			"source_ship_id": source.ShipID, "target_ship_id": targetSpec.ShipID,
			"special_id": "quantum_detonator", "damage_kind": "ship_explosion",
			"explosion_power": power, "explosion_radius_pixels": tacticalQuantumExplosionRadiusPixels,
			"distance_squared_pixels": candidate.distanceSquaredPixels, "raw_blast": rawBlast,
			"target_hull_roll_count": rollCount, "damage": incomingDamage,
			"incoming_damage": applied.IncomingDamage, "post_special_damage": applied.PostSpecialDamage,
			"energy_absorbed": applied.EnergyAbsorbed, "damper_reduced_damage": applied.DamperReducedDamage,
			"shield_arc": applied.ShieldArc, "shield_before": applied.ShieldBefore, "shield_after": applied.ShieldAfter,
			"post_block_damage": applied.PostBlockDamage, "shield_absorbed": applied.ShieldAbsorbed,
			"lower_layer_overflow": applied.LowerLayerOverflow,
			"armor_before":         applied.ArmorBefore, "armor_after": applied.ArmorAfter,
			"structure_damage_before": applied.StructureBefore, "structure_damage_after": applied.StructureAfter,
		}); err != nil {
			return err
		}
		affected = append(affected, targetSpec.ShipID)
		if !wasDestroyed && targetState.Destroyed {
			newlyDestroyed = append(newlyDestroyed, targetSpec.ShipID)
		}
	}

	if err := r.appendEvent("ship_explosion_resolved", eventSeat, commandSequence, map[string]any{
		"ship_id": source.ShipID, "special_id": "quantum_detonator",
		"explosion_power": power, "explosion_radius_pixels": tacticalQuantumExplosionRadiusPixels,
		"explosion_pixels_per_field": tacticalExplosionPixelsPerField,
		"affected_ship_ids":          affected, "destroyed_ship_ids": newlyDestroyed,
	}); err != nil {
		return err
	}

	for _, shipID := range newlyDestroyed {
		if err := tacticalResolveShipDestruction(spec, r, shipID, eventSeat, commandSequence, "ship_explosion"); err != nil {
			return err
		}
	}
	return nil
}

func tacticalResolveShipDestruction(spec TacticalSpec, r *tacticalRuntime, shipID core.ID, eventSeat protocol.SeatID, commandSequence uint32, reason string) error {
	if r == nil {
		return fmt.Errorf("ship destruction requires tactical runtime")
	}
	ship := tacticalShipSpec(spec, shipID)
	state := r.shipState(shipID)
	if ship == nil || state == nil {
		return fmt.Errorf("ship destruction references unknown ship %d", shipID)
	}
	if !state.Destroyed {
		return fmt.Errorf("ship destruction event for live ship %d", shipID)
	}
	data, quantum, power, err := tacticalShipDestructionEventData(spec, r, *ship, reason)
	if err != nil {
		return err
	}
	if err := r.appendEvent("ship_destroyed", eventSeat, commandSequence, data); err != nil {
		return err
	}
	if !quantum {
		return nil
	}
	return tacticalResolveQuantumDeathExplosion(spec, r, *ship, *state, eventSeat, commandSequence, power)
}
