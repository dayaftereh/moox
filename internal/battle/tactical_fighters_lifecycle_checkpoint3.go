package battle

import (
	"fmt"
	"sort"

	"moox/internal/core"
	"moox/internal/protocol"
)

func tacticalFighterSourceEmpire(spec TacticalSpec, group TacticalFighterGroupState) (core.ID, bool) {
	if tacticalShipSpec(spec, group.SourceShipID) == nil {
		return 0, false
	}
	empireID := tacticalEmpireForSeat(spec, group.SourceSeatID)
	if empireID == 0 {
		return 0, false
	}
	return empireID, true
}

func tacticalFighterLegalTargets(spec TacticalSpec, r *tacticalRuntime, source TacticalShipSpec, kind string) []TacticalFireTarget {
	sourceState := r.shipState(source.ShipID)
	if sourceState == nil {
		return nil
	}
	return tacticalFighterLegalTargetsForEmpire(spec, r, source, tacticalOwnerEmpire(source, sourceState), kind)
}

func tacticalFighterLegalTargetsForEmpire(spec TacticalSpec, r *tacticalRuntime, source TacticalShipSpec, sourceEmpire core.ID, kind string) []TacticalFireTarget {
	out := make([]TacticalFireTarget, 0)
	sourceState := r.shipState(source.ShipID)
	if sourceState == nil || sourceEmpire == 0 {
		return nil
	}
	ships := make([]TacticalFireTarget, 0, len(spec.Ships))
	for _, target := range spec.Ships {
		state := r.shipState(target.ShipID)
		if !tacticalShipTargetable(state) || tacticalOwnerEmpire(target, state) == sourceEmpire {
			continue
		}
		ships = append(ships, TacticalFireTarget{
			TargetShipID: target.ShipID,
			RangeIndex:   tacticalRangeIndex(sourceState.X, sourceState.Y, state.X, state.Y),
		})
	}
	sort.Slice(ships, func(i, j int) bool { return ships[i].TargetShipID < ships[j].TargetShipID })
	out = append(out, ships...)
	if kind == "bomber" {
		return out
	}
	projectiles := make([]TacticalFireTarget, 0, len(r.ordnance.Projectiles))
	for _, projectile := range r.ordnance.Projectiles {
		targetEmpire := tacticalEmpireForSeat(spec, projectile.SourceSeatID)
		if targetEmpire == 0 || targetEmpire == sourceEmpire {
			continue
		}
		projectiles = append(projectiles, TacticalFireTarget{
			TargetProjectileID: projectile.ProjectileID,
			RangeIndex:         tacticalRangeIndex(sourceState.X, sourceState.Y, projectile.X, projectile.Y),
		})
	}
	sort.Slice(projectiles, func(i, j int) bool { return projectiles[i].TargetProjectileID < projectiles[j].TargetProjectileID })
	out = append(out, projectiles...)
	groups := make([]TacticalFireTarget, 0, len(r.fighters.Groups))
	for _, group := range r.fighters.Groups {
		targetEmpire, ok := tacticalFighterSourceEmpire(spec, group)
		if !ok || targetEmpire == sourceEmpire || group.Multiplicity <= 0 {
			continue
		}
		groups = append(groups, TacticalFireTarget{
			TargetFighterGroupID: group.GroupID,
			RangeIndex:           tacticalRangeIndex(sourceState.X, sourceState.Y, group.X, group.Y),
		})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].TargetFighterGroupID < groups[j].TargetFighterGroupID })
	out = append(out, groups...)
	return out
}

func tacticalFighterShipAttackPosition(r *tacticalRuntime, group TacticalFighterGroupState, target TacticalShipState) (x, y int, ok bool) {
	if r == nil {
		return 0, 0, false
	}
	shipOccupied := make(map[[2]int]struct{}, len(r.state.Ships))
	for _, ship := range r.state.Ships {
		shipOccupied[[2]int{ship.X, ship.Y}] = struct{}{}
	}
	fighterOccupied := make(map[[2]int]struct{}, len(r.fighters.Groups))
	for _, other := range r.fighters.Groups {
		if other.GroupID == group.GroupID {
			continue
		}
		fighterOccupied[[2]int{other.X, other.Y}] = struct{}{}
	}
	type candidate struct {
		x, y       int
		moveFields int
		distance2  int64
	}
	choose := func(avoidFighters bool) (int, int, bool) {
		candidates := make([]candidate, 0, 8)
		for oy := -1; oy <= 1; oy++ {
			for ox := -1; ox <= 1; ox++ {
				if ox == 0 && oy == 0 {
					continue
				}
				cx, cy := target.X+ox, target.Y+oy
				if _, blocked := shipOccupied[[2]int{cx, cy}]; blocked {
					continue
				}
				if avoidFighters {
					if _, blocked := fighterOccupied[[2]int{cx, cy}]; blocked {
						continue
					}
				}
				dx, dy := cx-group.X, cy-group.Y
				candidates = append(candidates, candidate{
					x: cx, y: cy,
					moveFields: tacticalProjectileFieldDistance(dx, dy),
					distance2:  int64(dx)*int64(dx) + int64(dy)*int64(dy),
				})
			}
		}
		if len(candidates) == 0 {
			return 0, 0, false
		}
		sort.Slice(candidates, func(i, j int) bool {
			if candidates[i].moveFields != candidates[j].moveFields {
				return candidates[i].moveFields < candidates[j].moveFields
			}
			if candidates[i].distance2 != candidates[j].distance2 {
				return candidates[i].distance2 < candidates[j].distance2
			}
			if candidates[i].y != candidates[j].y {
				return candidates[i].y < candidates[j].y
			}
			return candidates[i].x < candidates[j].x
		})
		return candidates[0].x, candidates[0].y, true
	}
	if x, y, ok := choose(true); ok {
		return x, y, true
	}
	// More than eight moving actors may surround one ship. In that case sharing
	// an adjacent Fighter station is preferable to ever occupying a ship cell.
	return choose(false)
}

func tacticalFighterTargetPosition(spec TacticalSpec, r *tacticalRuntime, group TacticalFighterGroupState) (x, y int, ok bool) {
	sourceEmpire, sourceOK := tacticalFighterSourceEmpire(spec, group)
	if !sourceOK {
		return 0, 0, false
	}
	if group.Phase == TacticalFighterPhaseReturning {
		carrier := tacticalShipSpec(spec, group.SourceShipID)
		state := r.shipState(group.SourceShipID)
		if carrier == nil || state == nil || state.Destroyed || tacticalControllerSeat(*carrier, state) != group.SourceSeatID {
			return 0, 0, false
		}
		return state.X, state.Y, true
	}
	if group.TargetShipID != 0 {
		targetSpec := tacticalShipSpec(spec, group.TargetShipID)
		state := r.shipState(group.TargetShipID)
		if targetSpec == nil || !tacticalShipTargetable(state) || tacticalOwnerEmpire(*targetSpec, state) == sourceEmpire {
			return 0, 0, false
		}
		return tacticalFighterShipAttackPosition(r, group, *state)
	}
	if group.TargetProjectileID != 0 {
		if group.Kind == "bomber" {
			return 0, 0, false
		}
		index := tacticalProjectileIndexByID(r, group.TargetProjectileID)
		if index < 0 {
			return 0, 0, false
		}
		projectile := r.ordnance.Projectiles[index]
		targetSource := tacticalShipSpec(spec, projectile.SourceShipID)
		if targetSource == nil || tacticalEmpireForSeat(spec, projectile.SourceSeatID) == sourceEmpire {
			return 0, 0, false
		}
		return projectile.X, projectile.Y, true
	}
	if group.TargetFighterGroupID != 0 {
		if group.Kind == "bomber" || group.TargetFighterGroupID == group.GroupID {
			return 0, 0, false
		}
		index := tacticalFighterGroupIndexByID(r, group.TargetFighterGroupID)
		if index < 0 {
			return 0, 0, false
		}
		target := r.fighters.Groups[index]
		targetEmpire, ok := tacticalFighterSourceEmpire(spec, target)
		if !ok || targetEmpire == sourceEmpire {
			return 0, 0, false
		}
		return target.X, target.Y, true
	}
	return 0, 0, false
}

func tacticalFighterTargetEventFields(group TacticalFighterGroupState) map[string]any {
	return map[string]any{
		"target_ship_id":          group.TargetShipID,
		"target_projectile_id":    group.TargetProjectileID,
		"target_fighter_group_id": group.TargetFighterGroupID,
	}
}

func tacticalFighterSetTarget(group *TacticalFighterGroupState, target TacticalFireTarget) {
	group.TargetShipID = target.TargetShipID
	group.TargetProjectileID = target.TargetProjectileID
	group.TargetFighterGroupID = target.TargetFighterGroupID
}

func tacticalFighterRetarget(spec TacticalSpec, r *tacticalRuntime, group *TacticalFighterGroupState, eventSeat protocol.SeatID, commandSequence uint32) (bool, error) {
	if group == nil {
		return false, fmt.Errorf("Fighter retarget requires group")
	}
	source := tacticalShipSpec(spec, group.SourceShipID)
	if source == nil {
		return false, fmt.Errorf("Fighter group %d source carrier is missing", group.GroupID)
	}
	sourceEmpire, ok := tacticalFighterSourceEmpire(spec, *group)
	if !ok {
		return false, fmt.Errorf("Fighter group %d source seat %d is invalid", group.GroupID, group.SourceSeatID)
	}
	targets := tacticalFighterLegalTargetsForEmpire(spec, r, *source, sourceEmpire, group.Kind)
	if len(targets) == 0 {
		return false, nil
	}
	previous := tacticalFighterTargetEventFields(*group)
	tacticalFighterSetTarget(group, targets[0])
	data := map[string]any{
		"group_id":                         group.GroupID,
		"reason":                           "target_invalid",
		"previous_target_ship_id":          previous["target_ship_id"],
		"previous_target_projectile_id":    previous["target_projectile_id"],
		"previous_target_fighter_group_id": previous["target_fighter_group_id"],
	}
	for key, value := range tacticalFighterTargetEventFields(*group) {
		data["new_"+key] = value
	}
	if err := r.appendEvent("fighter_group_retargeted", eventSeat, commandSequence, data); err != nil {
		return false, err
	}
	return true, nil
}

func tacticalFighterCarrierFunctional(spec TacticalSpec, r *tacticalRuntime, group TacticalFighterGroupState) bool {
	carrier := tacticalShipSpec(spec, group.SourceShipID)
	state := r.shipState(group.SourceShipID)
	return carrier != nil && state != nil && !state.Destroyed && tacticalControllerSeat(*carrier, state) == group.SourceSeatID
}

func tacticalFighterBeginReturn(spec TacticalSpec, r *tacticalRuntime, group *TacticalFighterGroupState, reason string, eventSeat protocol.SeatID, commandSequence uint32) error {
	if group == nil {
		return fmt.Errorf("Fighter return requires group")
	}
	if !tacticalFighterCarrierFunctional(spec, r, *group) {
		return fmt.Errorf("Fighter group %d carrier is unavailable", group.GroupID)
	}
	if group.Phase == TacticalFighterPhaseReturning {
		return nil
	}
	previous := tacticalFighterTargetEventFields(*group)
	group.Phase = TacticalFighterPhaseReturning
	group.TargetShipID = group.SourceShipID
	group.TargetProjectileID = 0
	group.TargetFighterGroupID = 0
	return r.appendEvent("fighter_group_returning", eventSeat, commandSequence, map[string]any{
		"group_id": group.GroupID, "carrier_ship_id": group.SourceShipID, "reason": reason,
		"previous_target_ship_id":          previous["target_ship_id"],
		"previous_target_projectile_id":    previous["target_projectile_id"],
		"previous_target_fighter_group_id": previous["target_fighter_group_id"],
		"multiplicity":                     group.Multiplicity, "attack_stock": group.AttackStock,
	})
}

func tacticalFighterRecover(spec TacticalSpec, r *tacticalRuntime, group *TacticalFighterGroupState, eventSeat protocol.SeatID, commandSequence uint32) error {
	if group == nil || spec.Rules.Fighter == nil {
		return fmt.Errorf("Fighter recovery requires group and rules")
	}
	if !tacticalFighterCarrierFunctional(spec, r, *group) {
		return fmt.Errorf("Fighter group %d carrier is unavailable", group.GroupID)
	}
	carrier := tacticalShipSpec(spec, group.SourceShipID)
	carrierState := r.shipState(group.SourceShipID)
	sourceWeapon := weaponSpec(*carrier, group.SourceWeaponSlot)
	if sourceWeapon == nil || sourceWeapon.TacticalSupport != "fighter" {
		return fmt.Errorf("Fighter group %d source Bay is unavailable", group.GroupID)
	}
	divisor := spec.Rules.Fighter.RecoveryDivisor
	if divisor <= 0 {
		return fmt.Errorf("Fighter recovery divisor %d is invalid", divisor)
	}
	completeGroups := group.Multiplicity / divisor
	discardedCraft := group.Multiplicity % divisor
	remaining := completeGroups

	type bayCandidate struct {
		slot int
		bay  *TacticalFighterBayState
		max  int
	}
	candidates := make([]bayCandidate, 0)
	addCandidate := func(slot int) {
		weapon := weaponSpec(*carrier, slot)
		bay := tacticalFighterBayState(&r.fighters, carrier.ShipID, slot)
		if weapon == nil || bay == nil || weapon.TacticalSupport != "fighter" || weapon.WeaponID != sourceWeapon.WeaponID {
			return
		}
		candidates = append(candidates, bayCandidate{slot: slot, bay: bay, max: weapon.Count})
	}
	addCandidate(group.SourceWeaponSlot)
	var otherSlots []int
	for _, weapon := range carrier.Weapons {
		if weapon.Slot != group.SourceWeaponSlot && weapon.TacticalSupport == "fighter" && weapon.WeaponID == sourceWeapon.WeaponID {
			otherSlots = append(otherSlots, weapon.Slot)
		}
	}
	sort.Ints(otherSlots)
	for _, slot := range otherSlots {
		addCandidate(slot)
	}

	affected := make([]map[string]any, 0, len(candidates))
	recovered := 0
	for _, candidate := range candidates {
		before := candidate.bay.ReadyGroups
		capacity := candidate.max - before
		if capacity < 0 {
			capacity = 0
		}
		add := remaining
		if add > capacity {
			add = capacity
		}
		candidate.bay.ReadyGroups += add
		remaining -= add
		recovered += add
		if ready := weaponState(carrierState, candidate.slot); ready != nil {
			ready.Ready = candidate.bay.ReadyGroups > 0 && !carrierState.ActivationComplete
		}
		affected = append(affected, map[string]any{
			"weapon_slot": candidate.slot, "ready_groups_before": before,
			"ready_groups_after": candidate.bay.ReadyGroups, "capacity": candidate.max,
		})
	}
	return r.appendEvent("fighter_group_recovered", eventSeat, commandSequence, map[string]any{
		"group_id": group.GroupID, "carrier_ship_id": carrier.ShipID, "source_weapon_slot": group.SourceWeaponSlot,
		"multiplicity": group.Multiplicity, "recovery_divisor": divisor,
		"complete_groups": completeGroups, "recovered_groups": recovered,
		"incomplete_craft_discarded": discardedCraft, "unrecovered_complete_groups": remaining,
		"affected_bays": affected,
	})
}

func tacticalFighterPayloadHasEffect(payload *TacticalFighterPayloadSpec, id string) bool {
	if payload == nil {
		return false
	}
	for _, effect := range payload.InnateEffectIDs {
		if effect == id {
			return true
		}
	}
	return false
}

func tacticalFighterDamageOptions(spec TacticalSpec, r *tacticalRuntime, targetSpec TacticalShipSpec, targetState TacticalShipState, payload *TacticalFighterPayloadSpec) tacticalDamageOptions {
	options := tacticalDamageOptions{}
	if tacticalFighterPayloadHasEffect(payload, "natural_shield_piercing") {
		options.ShieldPiercing = true
	}
	if TacticalSpecialFunctional(targetSpec, targetState, r.legacy17_8Disabled, "hard_shields") && spec.Rules.Specials != nil {
		options.ShieldAdditionalBlock = spec.Rules.Specials.HardShieldsAdditionalAbsorption
	}
	return options
}

func tacticalFighterImpactSource(group TacticalFighterGroupState, target TacticalShipState) TacticalShipState {
	projectile := TacticalProjectileState{Facing: group.Facing}
	return tacticalProjectileImpactSource(projectile, target)
}

func tacticalReactiveFighterPointDefenseSlots(spec TacticalSpec, r *tacticalRuntime, target *TacticalShipSpec, group TacticalFighterGroupState) []int {
	if target == nil {
		return nil
	}
	targetState := r.shipState(target.ShipID)
	if targetState == nil || targetState.Destroyed {
		return nil
	}
	point := TacticalShipState{X: group.X, Y: group.Y}
	slots := make([]int, 0)
	for _, weapon := range target.Weapons {
		if !tacticalBeamFireSupported(weapon) || canonicalTacticalWeaponMountType(weapon.MountType) != core.ShipWeaponMountPointDefense {
			continue
		}
		configured, err := tacticalBeamConfigured(weapon)
		if err != nil || !tacticalBeamTargetInArc(*targetState, point, weapon.FiringArcID) {
			continue
		}
		geometricRangeIndex := tacticalRangeIndex(targetState.X, targetState.Y, group.X, group.Y)
		rangeIndex := tacticalBeamEffectiveRangeIndex(geometricRangeIndex, len(spec.Rules.BeamToHitRangeModifiers), configured.RangeProfile)
		if rangeIndex < 0 || rangeIndex >= len(spec.Rules.BeamToHitRangeModifiers) || rangeIndex >= len(spec.Rules.BeamDamageRangeModifiers) {
			continue
		}
		slots = append(slots, weapon.Slot)
	}
	sort.Ints(slots)
	return slots
}

func resolveTacticalFighterPreImpactDefenses(spec TacticalSpec, r *tacticalRuntime, groupID uint64, target *TacticalShipSpec, commandSequence uint32) (bool, error) {
	index := tacticalFighterGroupIndexByID(r, groupID)
	if index < 0 || target == nil {
		return true, nil
	}
	group := &r.fighters.Groups[index]
	targetState := r.shipState(target.ShipID)
	if targetState == nil || targetState.Destroyed {
		return false, nil
	}
	if spec.Rules.Specials != nil && TacticalSpecialFunctional(*target, *targetState, r.legacy17_8Disabled, "lightning_field") {
		before := group.Multiplicity
		rolls := make([]int, 0, before)
		for craftIndex := 0; craftIndex < group.Multiplicity; craftIndex++ {
			sessionRoll, err := r.random(100, spec.Rules)
			if err != nil {
				return false, err
			}
			rawRoll := sessionRoll - 1
			rolls = append(rolls, rawRoll)
			if rawRoll < spec.Rules.Specials.LightningDestroyThreshold {
				group.Multiplicity--
			}
		}
		if err := r.appendEvent("fighter_interception_resolved", target.SeatID, commandSequence, map[string]any{
			"group_id": group.GroupID, "interceptor_source_kind": "lightning_field", "interceptor_ship_id": target.ShipID,
			"multiplicity_before": before, "multiplicity_after": group.Multiplicity,
			"destroyed_craft": before - group.Multiplicity, "rolls": rolls,
			"destroy_threshold": spec.Rules.Specials.LightningDestroyThreshold,
		}); err != nil {
			return false, err
		}
		if group.Multiplicity <= 0 {
			return true, nil
		}
	}

	targetBeamDefense, err := tacticalFighterBeamInterceptorDCV(spec, *group)
	if err != nil {
		return false, err
	}
	for _, slot := range tacticalReactiveFighterPointDefenseSlots(spec, r, target, *group) {
		if group.Multiplicity <= 0 {
			return true, nil
		}
		result, err := resolveReactiveBeamPointDefense(Spec{Tactical: &spec}, r, ReactiveBeamPointDefenseRequest{
			ShipID: target.ShipID, WeaponSlot: slot,
			TargetX: group.X, TargetY: group.Y, TargetBeamDefense: targetBeamDefense,
		})
		if err != nil {
			return false, err
		}
		for _, shot := range result.Shots {
			if !shot.Hit || shot.Damage <= 0 || group.Multiplicity <= 0 {
				continue
			}
			applied, err := applyTacticalFighterInterception(r, group, shot.Damage, "beam_reactive", target.ShipID, slot, target.SeatID, commandSequence)
			if err != nil {
				return false, err
			}
			if applied.Destroyed {
				return true, nil
			}
		}
	}
	return group.Multiplicity <= 0, nil
}

func tacticalFighterBeamTargetDefense(spec TacticalSpec, r *tacticalRuntime, group TacticalFighterGroupState) (int, error) {
	if group.TargetShipID != 0 {
		targetSpec := tacticalShipSpec(spec, group.TargetShipID)
		targetState := r.shipState(group.TargetShipID)
		if targetSpec == nil || targetState == nil || targetState.Destroyed {
			return 0, fmt.Errorf("Fighter group %d target ship is unavailable", group.GroupID)
		}
		return tacticalBeamDefenseForState(spec, *targetSpec, *targetState, r.legacy17_8Disabled), nil
	}
	if group.TargetProjectileID != 0 {
		index := tacticalProjectileIndexByID(r, group.TargetProjectileID)
		if index < 0 {
			return 0, fmt.Errorf("Fighter group %d target projectile is unavailable", group.GroupID)
		}
		return tacticalProjectileBeamInterceptorDCV(spec, r.ordnance.Projectiles[index])
	}
	if group.TargetFighterGroupID != 0 {
		index := tacticalFighterGroupIndexByID(r, group.TargetFighterGroupID)
		if index < 0 {
			return 0, fmt.Errorf("Fighter group %d target Fighter group is unavailable", group.GroupID)
		}
		return tacticalFighterBeamInterceptorDCV(spec, r.fighters.Groups[index])
	}
	return 0, fmt.Errorf("Fighter group %d has no Beam target", group.GroupID)
}

func tacticalFighterApplyShipDamage(spec TacticalSpec, r *tacticalRuntime, group TacticalFighterGroupState, targetSpec *TacticalShipSpec, payload *TacticalFighterPayloadSpec, craftIndex int, payloadKind string, damage int, commandSequence uint32) error {
	if targetSpec == nil || damage <= 0 {
		return nil
	}
	targetState := r.shipState(targetSpec.ShipID)
	if targetState == nil || targetState.Destroyed {
		return nil
	}
	wasDestroyed := targetState.Destroyed
	options := tacticalFighterDamageOptions(spec, r, *targetSpec, *targetState, payload)
	sourceState := tacticalFighterImpactSource(group, *targetState)
	applied := applyTacticalDamageWithSpecials(spec, r, *targetSpec, targetState, sourceState, damage, options)
	event := map[string]any{
		"source_ship_id": group.SourceShipID, "target_ship_id": targetSpec.ShipID,
		"fighter_group_id": group.GroupID, "craft_index": craftIndex, "payload_kind": payloadKind,
		"weapon_slot": group.SourceWeaponSlot, "weapon_id": payload.WeaponID,
		"layer": applied.Layer, "damage": damage, "incoming_damage": applied.IncomingDamage,
		"shield_arc": applied.ShieldArc, "shield_before": applied.ShieldBefore, "shield_after": applied.ShieldAfter,
		"post_block_damage": applied.PostBlockDamage, "shield_absorbed": applied.ShieldAbsorbed,
		"lower_layer_overflow": applied.LowerLayerOverflow,
		"armor_before":         applied.ArmorBefore, "armor_after": applied.ArmorAfter,
		"structure_damage_before": applied.StructureBefore, "structure_damage_after": applied.StructureAfter,
		"extra_structure_damage": applied.ExtraStructureDamage,
	}
	if err := r.appendEvent("battle_damage_applied", group.SourceSeatID, commandSequence, event); err != nil {
		return err
	}
	if !wasDestroyed && targetState.Destroyed {
		if err := tacticalResolveShipDestruction(spec, r, targetSpec.ShipID, group.SourceSeatID, commandSequence, "fighter"); err != nil {
			return err
		}
	}
	return nil
}

func tacticalFighterApplyMovingActorDamage(spec TacticalSpec, r *tacticalRuntime, group TacticalFighterGroupState, damage int, commandSequence uint32) error {
	if damage <= 0 {
		return nil
	}
	if group.TargetProjectileID != 0 {
		index := tacticalProjectileIndexByID(r, group.TargetProjectileID)
		if index < 0 {
			return nil
		}
		projectile := &r.ordnance.Projectiles[index]
		applied, err := applyTacticalProjectileInterception(r, projectile, damage, "fighter_beam", group.SourceShipID, group.SourceWeaponSlot, group.SourceSeatID, commandSequence)
		if err != nil {
			return err
		}
		if applied.Destroyed {
			id := projectile.ProjectileID
			r.ordnance.Projectiles = append(r.ordnance.Projectiles[:index], r.ordnance.Projectiles[index+1:]...)
			return r.appendEvent("projectile_destroyed", group.SourceSeatID, commandSequence, map[string]any{"projectile_id": id, "reason": "intercepted"})
		}
		return nil
	}
	if group.TargetFighterGroupID != 0 {
		index := tacticalFighterGroupIndexByID(r, group.TargetFighterGroupID)
		if index < 0 {
			return nil
		}
		target := &r.fighters.Groups[index]
		before := target.Multiplicity
		applied, err := applyTacticalFighterInterception(r, target, damage, "fighter_beam", group.SourceShipID, group.SourceWeaponSlot, group.SourceSeatID, commandSequence)
		if err != nil {
			return err
		}
		if applied.Destroyed {
			id := target.GroupID
			r.fighters.Groups = append(r.fighters.Groups[:index], r.fighters.Groups[index+1:]...)
			return r.appendEvent("fighter_group_destroyed", group.SourceSeatID, commandSequence, map[string]any{
				"group_id": id, "reason": "intercepted", "multiplicity_before": before,
			})
		}
	}
	return nil
}

func resolveTacticalFighterBeamPass(spec TacticalSpec, r *tacticalRuntime, groupID uint64, commandSequence uint32) error {
	index := tacticalFighterGroupIndexByID(r, groupID)
	if index < 0 {
		return nil
	}
	group := r.fighters.Groups[index]
	payload := group.BeamPayload
	if payload == nil {
		return fmt.Errorf("Fighter group %d has no Beam payload", groupID)
	}
	doctrine := tacticalFighterDoctrine(spec, group.SourceSeatID)
	if doctrine == nil {
		return fmt.Errorf("Fighter group %d has no doctrine", groupID)
	}
	minDamage, maxDamage := payload.MinDamage, payload.MaxDamage
	accuracyBonus := 0
	if payload.PointDefenseMode {
		accuracyBonus = 25
		minDamage = (minDamage + 1) / 2
		maxDamage = (maxDamage + 1) / 2
	}
	craftCount := group.Multiplicity
	attempts := make([]map[string]any, 0, craftCount)
	for craftIndex := 1; craftIndex <= craftCount; craftIndex++ {
		currentIndex := tacticalFighterGroupIndexByID(r, groupID)
		if currentIndex < 0 {
			break
		}
		current := r.fighters.Groups[currentIndex]
		if _, _, ok := tacticalFighterTargetPosition(spec, r, current); !ok {
			break
		}
		defense, err := tacticalFighterBeamTargetDefense(spec, r, current)
		if err != nil {
			break
		}
		rawRoll, err := r.random(100, spec.Rules)
		if err != nil {
			return err
		}
		effectiveRoll := rawRoll
		if rawRoll > spec.Rules.BeamMaxHitThreshold {
			effectiveRoll = spec.Rules.BeamEffectiveRollCap
		} else {
			effectiveRoll += doctrine.FighterBeamOffense - defense + accuracyBonus
			if effectiveRoll > spec.Rules.BeamEffectiveRollCap {
				effectiveRoll = spec.Rules.BeamEffectiveRollCap
			}
		}
		hit := effectiveRoll >= spec.Rules.BeamBaseHitThreshold
		damage := 0
		if hit {
			spread := maxDamage - minDamage
			if spread <= 0 {
				damage = minDamage
			} else {
				damage = minDamage + ((effectiveRoll-spec.Rules.BeamBaseHitThreshold)*(spread+1))/(spec.Rules.BeamEffectiveRollCap-spec.Rules.BeamBaseHitThreshold)
			}
		}
		attempts = append(attempts, map[string]any{
			"craft_index": craftIndex, "payload_kind": "beam", "raw_hit_roll": rawRoll,
			"effective_hit_roll": effectiveRoll, "target_defense": defense,
			"threshold": spec.Rules.BeamBaseHitThreshold, "hit": hit, "damage": damage,
		})
		if hit && damage > 0 {
			if current.TargetShipID != 0 {
				if err := tacticalFighterApplyShipDamage(spec, r, current, tacticalShipSpec(spec, current.TargetShipID), payload, craftIndex, "beam", damage, commandSequence); err != nil {
					return err
				}
			} else if err := tacticalFighterApplyMovingActorDamage(spec, r, current, damage, commandSequence); err != nil {
				return err
			}
		}
	}
	index = tacticalFighterGroupIndexByID(r, groupID)
	if index < 0 {
		return nil
	}
	group = r.fighters.Groups[index]
	data := map[string]any{
		"group_id": groupID, "payload_kind": "beam", "weapon_id": payload.WeaponID,
		"attack_stock_before": group.AttackStock, "attempts": attempts,
	}
	for key, value := range tacticalFighterTargetEventFields(group) {
		data[key] = value
	}
	return r.appendEvent("fighter_attack_resolved", group.SourceSeatID, commandSequence, data)
}

func resolveTacticalFighterBombPass(spec TacticalSpec, r *tacticalRuntime, groupID uint64, commandSequence uint32) error {
	index := tacticalFighterGroupIndexByID(r, groupID)
	if index < 0 {
		return nil
	}
	group := r.fighters.Groups[index]
	payload := group.BombPayload
	if payload == nil || group.TargetShipID == 0 {
		return fmt.Errorf("Fighter group %d has no valid Bomb payload/ship target", groupID)
	}
	craftCount := group.Multiplicity
	attempts := make([]map[string]any, 0, craftCount)
	for craftIndex := 1; craftIndex <= craftCount; craftIndex++ {
		currentIndex := tacticalFighterGroupIndexByID(r, groupID)
		if currentIndex < 0 {
			break
		}
		current := r.fighters.Groups[currentIndex]
		targetSpec := tacticalShipSpec(spec, current.TargetShipID)
		targetState := r.shipState(current.TargetShipID)
		if targetSpec == nil || targetState == nil || targetState.Destroyed {
			break
		}
		sessionRoll, err := r.random(100, spec.Rules)
		if err != nil {
			return err
		}
		rawRoll := sessionRoll - 1
		hit := rawRoll > 0
		damage := 0
		if hit {
			spread := payload.MaxDamage - payload.MinDamage
			if spread <= 0 {
				damage = payload.MinDamage
			} else {
				damage = payload.MinDamage + ((rawRoll+100/(2*spread))*spread)/100
			}
		}
		attempts = append(attempts, map[string]any{
			"craft_index": craftIndex, "payload_kind": "bomb", "raw_hit_roll": rawRoll,
			"threshold": 1, "hit": hit, "damage": damage,
		})
		if hit && damage > 0 {
			if err := tacticalFighterApplyShipDamage(spec, r, current, targetSpec, payload, craftIndex, "bomb", damage, commandSequence); err != nil {
				return err
			}
		}
	}
	index = tacticalFighterGroupIndexByID(r, groupID)
	if index < 0 {
		return nil
	}
	group = r.fighters.Groups[index]
	data := map[string]any{
		"group_id": groupID, "payload_kind": "bomb", "weapon_id": payload.WeaponID,
		"attack_stock_before": group.AttackStock, "attempts": attempts,
	}
	for key, value := range tacticalFighterTargetEventFields(group) {
		data[key] = value
	}
	return r.appendEvent("fighter_attack_resolved", group.SourceSeatID, commandSequence, data)
}

func resolveTacticalFighterAttack(spec TacticalSpec, r *tacticalRuntime, groupID uint64, commandSequence uint32) error {
	index := tacticalFighterGroupIndexByID(r, groupID)
	if index < 0 {
		return nil
	}
	group := r.fighters.Groups[index]
	if group.AttackStock <= 0 {
		if !tacticalFighterCarrierFunctional(spec, r, group) {
			return nil // Arrival resolves carrier loss into a destroyed-group outcome.
		}
		return tacticalFighterBeginReturn(spec, r, &r.fighters.Groups[index], "attack_stock_exhausted", group.SourceSeatID, commandSequence)
	}
	switch group.Kind {
	case "fighter":
		if err := resolveTacticalFighterBeamPass(spec, r, groupID, commandSequence); err != nil {
			return err
		}
	case "bomber":
		if err := resolveTacticalFighterBombPass(spec, r, groupID, commandSequence); err != nil {
			return err
		}
	case "heavy_fighter":
		if err := resolveTacticalFighterBeamPass(spec, r, groupID, commandSequence); err != nil {
			return err
		}
		if current := tacticalFighterGroupIndexByID(r, groupID); current >= 0 {
			g := r.fighters.Groups[current]
			if g.TargetShipID != 0 {
				targetState := r.shipState(g.TargetShipID)
				if targetState != nil && !targetState.Destroyed {
					if err := resolveTacticalFighterBombPass(spec, r, groupID, commandSequence); err != nil {
						return err
					}
				}
			}
		}
	default:
		return fmt.Errorf("Fighter group %d has invalid kind %q", groupID, group.Kind)
	}
	index = tacticalFighterGroupIndexByID(r, groupID)
	if index < 0 {
		return nil
	}
	groupPtr := &r.fighters.Groups[index]
	groupPtr.AttackStock--
	if groupPtr.AttackStock <= 0 {
		if !tacticalFighterCarrierFunctional(spec, r, *groupPtr) {
			return nil // The caller emits fighter_group_destroyed: carrier_unavailable.
		}
		return tacticalFighterBeginReturn(spec, r, groupPtr, "attack_stock_exhausted", groupPtr.SourceSeatID, commandSequence)
	}
	return nil
}

func resolveTacticalFighterArrival(spec TacticalSpec, r *tacticalRuntime, groupID uint64, eventSeat protocol.SeatID, commandSequence uint32) (tacticalFighterQuantumOutcome, error) {
	index := tacticalFighterGroupIndexByID(r, groupID)
	if index < 0 {
		return tacticalFighterQuantumOutcome{RemoveWithoutDestroy: true}, nil
	}
	group := &r.fighters.Groups[index]
	if group.Phase == TacticalFighterPhaseReturning {
		if !tacticalFighterCarrierFunctional(spec, r, *group) {
			return tacticalFighterQuantumOutcome{DestroyReason: "carrier_unavailable"}, nil
		}
		if err := tacticalFighterRecover(spec, r, group, eventSeat, commandSequence); err != nil {
			return tacticalFighterQuantumOutcome{}, err
		}
		return tacticalFighterQuantumOutcome{RemoveWithoutDestroy: true}, nil
	}
	if group.TargetShipID != 0 {
		target := tacticalShipSpec(spec, group.TargetShipID)
		destroyed, err := resolveTacticalFighterPreImpactDefenses(spec, r, groupID, target, commandSequence)
		if err != nil {
			return tacticalFighterQuantumOutcome{}, err
		}
		if destroyed {
			return tacticalFighterQuantumOutcome{DestroyReason: "intercepted"}, nil
		}
	}
	if err := resolveTacticalFighterAttack(spec, r, groupID, commandSequence); err != nil {
		return tacticalFighterQuantumOutcome{}, err
	}
	if index := tacticalFighterGroupIndexByID(r, groupID); index >= 0 {
		group = &r.fighters.Groups[index]
		if group.AttackStock <= 0 && !tacticalFighterCarrierFunctional(spec, r, *group) {
			return tacticalFighterQuantumOutcome{DestroyReason: "carrier_unavailable"}, nil
		}
	}
	return tacticalFighterQuantumOutcome{Keep: true}, nil
}
