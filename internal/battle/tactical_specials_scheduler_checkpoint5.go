package battle

import (
	"moox/internal/core"
	"moox/internal/protocol"
)

const (
	TacticalPassKindNormal   = "normal"
	TacticalPassKindTimeWarp = "time_warp"
)

func tacticalActivationOrder(r *tacticalRuntime) []core.ID {
	if r == nil {
		return nil
	}
	if r.state.PassKind == TacticalPassKindTimeWarp && len(r.state.TimeWarpQueue) != 0 {
		return r.state.TimeWarpQueue
	}
	return r.state.InitiativeOrder
}

func tacticalSeatPassComplete(spec TacticalSpec, r *tacticalRuntime, seatID protocol.SeatID, timeWarpOnly bool) bool {
	if r == nil || seatID == 0 {
		return false
	}
	seen := false
	for _, ship := range spec.Ships {
		state := r.shipState(ship.ShipID)
		if state == nil || state.Destroyed || tacticalControllerSeat(ship, state) != seatID {
			continue
		}
		if timeWarpOnly && !TacticalSpecialFunctional(ship, *state, r.legacy17_8Disabled, "time_warp_facilitator") {
			continue
		}
		seen = true
		if !state.ActivationComplete {
			return false
		}
	}
	return seen
}

func tacticalTimeWarpEligibleShips(spec TacticalSpec, r *tacticalRuntime, seatID protocol.SeatID) []core.ID {
	if r == nil || r.legacy17_8Disabled || seatID == 0 {
		return nil
	}
	out := make([]core.ID, 0)
	for _, shipID := range r.state.InitiativeOrder {
		ship := tacticalShipSpec(spec, shipID)
		state := r.shipState(shipID)
		if ship == nil || state == nil || state.Destroyed || tacticalControllerSeat(*ship, state) != seatID {
			continue
		}
		if TacticalSpecialFunctional(*ship, *state, false, "time_warp_facilitator") {
			out = append(out, shipID)
		}
	}
	return out
}

func tacticalHasAnyTimeWarpShip(spec TacticalSpec, r *tacticalRuntime) bool {
	if r == nil || r.legacy17_8Disabled {
		return false
	}
	for _, ship := range spec.Ships {
		state := r.shipState(ship.ShipID)
		if state != nil && !state.Destroyed && TacticalSpecialFunctional(ship, *state, false, "time_warp_facilitator") {
			return true
		}
	}
	return false
}

func tacticalBeginNextTimeWarpPass(spec Spec, r *tacticalRuntime, afterSeat protocol.SeatID, commandSequence uint32) (bool, error) {
	if spec.Tactical == nil || r == nil || r.legacy17_8Disabled || !tacticalHasAnyTimeWarpShip(*spec.Tactical, r) {
		return false, nil
	}
	start := 0
	if afterSeat != 0 {
		for i, seatID := range spec.Participants {
			if seatID == afterSeat {
				start = i + 1
				break
			}
		}
	}
	for i := start; i < len(spec.Participants); i++ {
		seatID := spec.Participants[i]
		eligible := tacticalTimeWarpEligibleShips(*spec.Tactical, r, seatID)
		if err := r.appendEvent("time_warp_pass_started", seatID, commandSequence, map[string]any{
			"round": r.state.Round, "side_seat_id": seatID, "eligible_ship_ids": append([]core.ID(nil), eligible...),
		}); err != nil {
			return false, err
		}
		if len(eligible) == 0 {
			continue
		}
		r.state.PassKind = TacticalPassKindTimeWarp
		r.state.TimeWarpSeatID = seatID
		r.state.TimeWarpQueue = append(r.state.TimeWarpQueue[:0], eligible...)
		for _, shipID := range eligible {
			ship := tacticalShipSpec(*spec.Tactical, shipID)
			state := r.shipState(shipID)
			if ship == nil || state == nil {
				continue
			}
			state.ActivationComplete = false
			state.MovementCurrent = state.MovementMax
			for j := range state.Weapons {
				state.Weapons[j].Ready = true
			}
			tacticalStartShipTurnSpecials(*spec.Tactical, *ship, state)
		}
		refreshTacticalOrdnanceReadiness(*spec.Tactical, r)
		refreshTacticalFighterReadiness(*spec.Tactical, r)
		r.state.ActiveShipID = eligible[0]
		return true, nil
	}
	return false, nil
}

func tacticalEndOwnPassSpecials(spec TacticalSpec, r *tacticalRuntime, seatID protocol.SeatID, timeWarpOnly bool, commandSequence uint32) error {
	if r == nil || seatID == 0 || r.legacy17_8Disabled {
		return nil
	}
	timeWarpSet := map[core.ID]bool{}
	if timeWarpOnly {
		for _, id := range r.state.TimeWarpQueue {
			timeWarpSet[id] = true
		}
	}
	for i := range spec.Ships {
		ship := spec.Ships[i]
		state := r.shipState(ship.ShipID)
		if state == nil || state.Destroyed || tacticalControllerSeat(ship, state) != seatID {
			continue
		}
		if timeWarpOnly && !timeWarpSet[ship.ShipID] {
			continue
		}

		if state.EnergyAbsorberStored > 0 {
			before := state.EnergyAbsorberStored
			state.EnergyAbsorberStored = 0
			if err := r.appendEvent("special_state_changed", seatID, commandSequence, map[string]any{
				"ship_id": ship.ShipID, "special_id": "energy_absorber", "effect": "stored_energy_expired",
				"before": before, "after": 0,
			}); err != nil {
				return err
			}
		}

		beforeCloak := state.CloakState
		beforeTurns := state.PhasingTurnsRemaining
		switch state.CloakState {
		case TacticalCloakStateRecovering:
			state.CloakState = TacticalCloakStateCloaked
		case TacticalCloakStatePhased:
			if state.PhasingTurnsRemaining > 0 {
				state.PhasingTurnsRemaining--
			}
			if state.PhasingTurnsRemaining <= 0 {
				state.PhasingTurnsRemaining = 0
				state.CloakState = TacticalCloakStateCloaked
			}
		case TacticalCloakStatePhaseRecovering:
			state.CloakState = TacticalCloakStatePhased
			if err := reconcileTacticalUntargetableShipActors(spec, r, ship.ShipID, seatID, commandSequence); err != nil {
				return err
			}
		}
		if state.CloakState != beforeCloak || state.PhasingTurnsRemaining != beforeTurns {
			if err := r.appendEvent("special_state_changed", seatID, commandSequence, map[string]any{
				"ship_id": ship.ShipID, "special_id": "phasing_cloak", "effect": "cloak_lifecycle",
				"cloak_state_before": beforeCloak, "cloak_state_after": state.CloakState,
				"phasing_turns_before": beforeTurns, "phasing_turns_after": state.PhasingTurnsRemaining,
			}); err != nil {
				return err
			}
		}

		for j := range state.Weapons {
			weapon := weaponSpec(ship, state.Weapons[j].Slot)
			if weapon == nil || !tacticalWeaponEligibleForSpecialReadiness(ship, *state, *weapon) {
				continue
			}
			if state.Weapons[j].SpecialReadiness == TacticalSpecialReadinessRearmPending {
				state.Weapons[j].SpecialReadiness = TacticalSpecialReadinessExtraVolleyReady
				if err := r.appendEvent("special_state_changed", seatID, commandSequence, map[string]any{
					"ship_id": ship.ShipID, "special_id": tacticalWeaponReadinessSpecialID(*weapon),
					"effect": "weapon_special_readiness", "weapon_slot": weapon.Slot,
					"before": TacticalSpecialReadinessRearmPending, "after": TacticalSpecialReadinessExtraVolleyReady,
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func tacticalWeaponReadinessSpecialID(weapon TacticalWeaponSpec) string {
	if tacticalBeamFireSupported(weapon) {
		return "hyper_x_capacitors"
	}
	if tacticalWeaponFamily(weapon) == "missile" {
		return "fast_missile_racks"
	}
	return ""
}

func tacticalCompleteWeaponFire(ship TacticalShipSpec, shipState *TacticalShipState, weapon TacticalWeaponSpec, ready *TacticalWeaponState) bool {
	if shipState == nil || ready == nil {
		return false
	}
	extraVolley := false
	if tacticalWeaponEligibleForSpecialReadiness(ship, *shipState, weapon) {
		switch ready.SpecialReadiness {
		case TacticalSpecialReadinessExtraVolleyReady:
			ready.SpecialReadiness = TacticalSpecialReadinessNone
			extraVolley = true
		case TacticalSpecialReadinessRearmPending:
			ready.SpecialReadiness = TacticalSpecialReadinessNone
		}
	}
	ready.Ready = extraVolley
	return extraVolley
}

func tacticalResolveEnergyAbsorberPreWeapon(spec TacticalSpec, r *tacticalRuntime, source TacticalShipSpec, sourceState *TacticalShipState, target TacticalShipSpec, targetState *TacticalShipState, seatID protocol.SeatID, commandSequence uint32) (bool, error) {
	if r == nil || sourceState == nil || targetState == nil || r.legacy17_8Disabled || sourceState.EnergyAbsorberStored <= 0 ||
		!TacticalSpecialFunctional(source, *sourceState, false, "energy_absorber") {
		return false, nil
	}
	if tacticalRangeIndex(sourceState.X, sourceState.Y, targetState.X, targetState.Y) >= 8 {
		return false, nil
	}

	stored := sourceState.EnergyAbsorberStored
	forcedMiss := false
	displacementRoll := 0
	if spec.Rules.Specials != nil && TacticalSpecialFunctional(target, *targetState, false, "displacement_device") {
		roll, err := r.random(100, spec.Rules)
		if err != nil {
			return false, err
		}
		displacementRoll = roll
		// The audited Energy_Absorber_Attack_ branch is intentionally different
		// from ordinary Beam/ordnance Displacement: values > 30 zero this attack.
		forcedMiss = displacementRoll > spec.Rules.Specials.DisplacementBeamMissThreshold
	}

	wasDestroyed := targetState.Destroyed
	if !forcedMiss {
		options := tacticalDamageOptions{}
		if spec.Rules.Specials != nil && TacticalSpecialFunctional(target, *targetState, false, "hard_shields") {
			options.ShieldAdditionalBlock = spec.Rules.Specials.HardShieldsAdditionalAbsorption
		}
		applied := applyTacticalDamageWithSpecials(spec, r, target, targetState, *sourceState, stored, options)
		if err := r.appendEvent("battle_damage_applied", seatID, commandSequence, map[string]any{
			"source_ship_id": source.ShipID, "target_ship_id": target.ShipID,
			"special_id": "energy_absorber", "damage_kind": "energy_absorber", "damage": stored,
			"incoming_damage": applied.IncomingDamage, "post_special_damage": applied.PostSpecialDamage,
			"energy_absorbed": applied.EnergyAbsorbed, "damper_reduced_damage": applied.DamperReducedDamage,
			"shield_arc": applied.ShieldArc, "shield_before": applied.ShieldBefore, "shield_after": applied.ShieldAfter,
			"post_block_damage": applied.PostBlockDamage, "shield_absorbed": applied.ShieldAbsorbed,
			"lower_layer_overflow": applied.LowerLayerOverflow,
			"armor_before":         applied.ArmorBefore, "armor_after": applied.ArmorAfter,
			"structure_damage_before": applied.StructureBefore, "structure_damage_after": applied.StructureAfter,
		}); err != nil {
			return false, err
		}
	}

	sourceState.EnergyAbsorberStored = 0
	if !wasDestroyed && targetState.Destroyed {
		if err := tacticalResolveShipDestruction(spec, r, target.ShipID, seatID, commandSequence, "energy_absorber"); err != nil {
			return false, err
		}
	}
	if err := r.appendEvent("special_state_changed", seatID, commandSequence, map[string]any{
		"ship_id": source.ShipID, "special_id": "energy_absorber", "effect": "stored_energy_discharged",
		"target_ship_id": target.ShipID, "before": stored, "after": 0,
		"forced_miss": forcedMiss, "displacement_roll": displacementRoll,
	}); err != nil {
		return false, err
	}
	return targetState.Destroyed, nil
}
func tacticalRepairPercent(spec TacticalSpec, ship TacticalShipSpec, state TacticalShipState) int {
	percent := 0
	if doctrine := tacticalBoardingDoctrine(spec, tacticalControllerSeat(ship, &state)); doctrine != nil {
		percent += doctrine.FleetEngineerBonus
	}
	if spec.Rules.Specials != nil && TacticalSpecialFunctional(ship, state, false, "automated_repair_unit") {
		percent += spec.Rules.Specials.AutomatedRepairBasePercent
	}
	return percent
}

func tacticalMaxInternalSpace(spec *TacticalInternalSpec) int {
	if spec == nil {
		return 0
	}
	total := spec.EngineMax + spec.SecondaryMax + spec.ShieldGeneratorMax
	for _, weapon := range spec.Weapons {
		total += weapon.Max
	}
	for _, special := range spec.Specials {
		total += special.Max
	}
	return total
}

func tacticalRepairPool(current *int, max int, budget *int) {
	if current == nil || budget == nil || *budget <= 0 || max <= *current {
		return
	}
	need := max - *current
	if need > *budget {
		need = *budget
	}
	*current += need
	*budget -= need
}

func tacticalRepairShips(spec TacticalSpec, r *tacticalRuntime, commandSequence uint32) error {
	if r == nil || r.legacy17_8Disabled {
		return nil
	}
	for i := range spec.Ships {
		ship := spec.Ships[i]
		state := r.shipState(ship.ShipID)
		if state == nil || state.Destroyed || state.PostCaptureState {
			continue
		}
		percent := tacticalRepairPercent(spec, ship, *state)
		if percent <= 0 {
			continue
		}

		armorBefore := state.ArmorCurrent
		structureBefore := state.StructureDamage
		internalBefore := cloneTacticalInternalState(state.Internal)

		durabilityBudget := percent * (ship.StructureMax + ship.ArmorMax) / 100
		if durabilityBudget < 1 {
			durabilityBudget = 1
		}
		if state.StructureDamage > 0 {
			used := state.StructureDamage
			if used > durabilityBudget {
				used = durabilityBudget
			}
			state.StructureDamage -= used
			durabilityBudget -= used
		}
		if durabilityBudget > 0 && state.ArmorCurrent < ship.ArmorMax {
			tacticalRepairPool(&state.ArmorCurrent, ship.ArmorMax, &durabilityBudget)
		}

		internalBudget := percent * tacticalMaxInternalSpace(ship.Internal) / 200
		if internalBudget < 1 && ship.Internal != nil {
			internalBudget = 1
		}
		if ship.Internal != nil && internalBudget > 0 {
			tacticalRepairPool(&state.Internal.EngineCurrent, ship.Internal.EngineMax, &internalBudget)
			tacticalRepairPool(&state.Internal.SecondaryCurrent, ship.Internal.SecondaryMax, &internalBudget)
			for _, immutable := range ship.Internal.Weapons {
				for j := range state.Internal.Weapons {
					if state.Internal.Weapons[j].Slot == immutable.Slot {
						tacticalRepairPool(&state.Internal.Weapons[j].Current, immutable.Max, &internalBudget)
						break
					}
				}
			}
			tacticalRepairPool(&state.Internal.ShieldGeneratorCurrent, ship.Internal.ShieldGeneratorMax, &internalBudget)
			for _, immutable := range ship.Internal.Specials {
				for j := range state.Internal.Specials {
					if state.Internal.Specials[j].Slot == immutable.Slot {
						tacticalRepairPool(&state.Internal.Specials[j].Current, immutable.Max, &internalBudget)
						break
					}
				}
			}
		}

		if ship.Internal != nil {
			for _, immutable := range ship.Internal.Weapons {
				for j := range state.Internal.Weapons {
					if state.Internal.Weapons[j].Slot != immutable.Slot {
						continue
					}
					liveCount := 0
					if state.Internal.Weapons[j].Current > 0 {
						liveCount = (state.Internal.Weapons[j].Current + immutable.UnitSize - 1) / immutable.UnitSize
					}
					if weapon := weaponState(state, immutable.Slot); weapon != nil {
						weapon.CountCurrent = liveCount
						if liveCount == 0 {
							weapon.Ready = false
						}
					}
					break
				}
			}
		}
		reconcileTacticalShipDerivedState(spec, r, ship.ShipID)

		if armorBefore != state.ArmorCurrent || structureBefore != state.StructureDamage || !tacticalInternalStateEqual(internalBefore, state.Internal) {
			if err := r.appendEvent("ship_repaired", tacticalControllerSeat(ship, state), commandSequence, map[string]any{
				"ship_id": ship.ShipID, "repair_percent": percent,
				"armor_before": armorBefore, "armor_after": state.ArmorCurrent,
				"structure_damage_before": structureBefore, "structure_damage_after": state.StructureDamage,
				"internal_before": internalBefore, "internal_after": cloneTacticalInternalState(state.Internal),
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func tacticalInternalStateEqual(a, b TacticalInternalState) bool {
	if a.EngineCurrent != b.EngineCurrent || a.SecondaryCurrent != b.SecondaryCurrent || a.ShieldGeneratorCurrent != b.ShieldGeneratorCurrent ||
		len(a.Weapons) != len(b.Weapons) || len(a.Specials) != len(b.Specials) {
		return false
	}
	for i := range a.Weapons {
		if a.Weapons[i] != b.Weapons[i] {
			return false
		}
	}
	for i := range a.Specials {
		if a.Specials[i] != b.Specials[i] {
			return false
		}
	}
	return true
}

func rechargeTacticalShieldsWithEvents(spec TacticalSpec, r *tacticalRuntime, commandSequence uint32) error {
	if r == nil {
		return nil
	}
	before := make(map[core.ID][4]int, len(r.state.Ships))
	for _, state := range r.state.Ships {
		before[state.ShipID] = state.ShieldArcCurrent
	}
	rechargeTacticalShields(spec, r)
	for i := range spec.Ships {
		ship := spec.Ships[i]
		state := r.shipState(ship.ShipID)
		if state == nil {
			continue
		}
		old := before[ship.ShipID]
		if old == state.ShieldArcCurrent {
			continue
		}
		consumed := 0
		for arc := range old {
			if state.ShieldArcCurrent[arc] > old[arc] {
				consumed += state.ShieldArcCurrent[arc] - old[arc]
			}
		}
		if err := r.appendEvent("shields_recharged", tacticalControllerSeat(ship, state), commandSequence, map[string]any{
			"ship_id": ship.ShipID, "before": old, "after": state.ShieldArcCurrent, "budget_consumed": consumed,
		}); err != nil {
			return err
		}
	}
	return nil
}

func reconcileTacticalUntargetableShipActors(spec TacticalSpec, r *tacticalRuntime, shipID core.ID, eventSeat protocol.SeatID, commandSequence uint32) error {
	for i := 0; i < len(r.ordnance.Projectiles); {
		projectile := &r.ordnance.Projectiles[i]
		if projectile.TargetShipID != shipID {
			i++
			continue
		}
		sourceEmpire := tacticalEmpireForSeat(spec, projectile.SourceSeatID)
		targets := tacticalOrdnanceRetargetIDs(spec, r, sourceEmpire)
		if len(targets) == 0 {
			projectileID := projectile.ProjectileID
			r.ordnance.Projectiles = append(r.ordnance.Projectiles[:i], r.ordnance.Projectiles[i+1:]...)
			if err := r.appendEvent("projectile_destroyed", eventSeat, commandSequence, map[string]any{
				"projectile_id": projectileID, "reason": "no_target",
			}); err != nil {
				return err
			}
			continue
		}
		previous := projectile.TargetShipID
		projectile.TargetShipID = targets[0]
		if err := r.appendEvent("projectile_retargeted", eventSeat, commandSequence, map[string]any{
			"projectile_id": projectile.ProjectileID, "previous_target_ship_id": previous,
			"new_target_ship_id": projectile.TargetShipID, "reason": "target_invalid",
		}); err != nil {
			return err
		}
		i++
	}

	for i := 0; i < len(r.fighters.Groups); {
		group := &r.fighters.Groups[i]
		if group.Phase != TacticalFighterPhaseOutbound || group.TargetShipID != shipID {
			i++
			continue
		}
		retargeted, err := tacticalFighterRetarget(spec, r, group, eventSeat, commandSequence)
		if err != nil {
			return err
		}
		if retargeted {
			i++
			continue
		}
		if tacticalFighterCarrierFunctional(spec, r, *group) {
			if err := tacticalFighterBeginReturn(spec, r, group, "no_target", eventSeat, commandSequence); err != nil {
				return err
			}
			i++
			continue
		}
		groupID := group.GroupID
		multiplicity := group.Multiplicity
		r.fighters.Groups = append(r.fighters.Groups[:i], r.fighters.Groups[i+1:]...)
		if err := r.appendEvent("fighter_group_destroyed", eventSeat, commandSequence, map[string]any{
			"group_id": groupID, "reason": "carrier_unavailable", "multiplicity_before": multiplicity,
		}); err != nil {
			return err
		}
	}
	return nil
}
