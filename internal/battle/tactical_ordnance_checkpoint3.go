package battle

import (
	"fmt"
	"math"

	"moox/internal/core"
)

type tacticalConfiguredProjectile struct {
	Speed              int
	UnitDurability     int
	BeamInterceptorDCV int
	WarheadStrength    int
	MIRV               bool
	ECCM               bool
	Enveloping         bool
	NoRangeDissipation bool
}

type tacticalProjectileQuantumOutcome struct {
	Keep            bool
	DestroyReason   string
	DestroyedShipID core.ID
}

func deriveConfiguredTacticalProjectile(spec TacticalSpec, weapon TacticalWeaponSpec) (tacticalConfiguredProjectile, error) {
	rule := tacticalOrdnanceRule(spec, weapon.WeaponID)
	if rule == nil || spec.Rules.Ordnance == nil {
		return tacticalConfiguredProjectile{}, fmt.Errorf("ordnance rule missing for weapon %q", weapon.WeaponID)
	}
	family := tacticalWeaponFamily(weapon)
	if family != "missile" && family != "torpedo" {
		return tacticalConfiguredProjectile{}, fmt.Errorf("weapon %q family %q is outside projectile derivation", weapon.WeaponID, family)
	}

	configured := tacticalConfiguredProjectile{
		Speed:              rule.BaseSpeed,
		UnitDurability:     rule.BaseDurability,
		BeamInterceptorDCV: rule.BeamInterceptorDCV,
		WarheadStrength:    weapon.MaxDamage,
	}
	for _, modifier := range weapon.ModifierIDs {
		switch family {
		case "missile":
			switch modifier {
			case core.ShipWeaponModifierMIRV:
				configured.MIRV = true
			case core.ShipWeaponModifierECCM:
				configured.ECCM = true
			case core.ShipWeaponModifierArmored:
				configured.UnitDurability *= 2
			case core.ShipWeaponModifierFast:
				configured.Speed += spec.Rules.Ordnance.FastSpeedBonus
			case core.ShipWeaponModifierEmissionsGuidance:
				return tacticalConfiguredProjectile{}, fmt.Errorf("weapon %q Emissions Guidance remains outside Slice 17.7 Gate 3", weapon.WeaponID)
			default:
				return tacticalConfiguredProjectile{}, fmt.Errorf("Missile weapon %q has unsupported runtime modifier %q", weapon.WeaponID, modifier)
			}
		case "torpedo":
			switch modifier {
			case core.ShipWeaponModifierECCM:
				configured.ECCM = true
			case core.ShipWeaponModifierOverloaded:
				configured.WarheadStrength += configured.WarheadStrength / 2
			case core.ShipWeaponModifierEnveloping:
				configured.Enveloping = true
			case core.ShipWeaponModifierNoRangeDissipation:
				if weapon.WeaponID != "plasma_torpedoes" {
					return tacticalConfiguredProjectile{}, fmt.Errorf("No Range Dissipation is only executable for Plasma Torpedoes")
				}
				configured.NoRangeDissipation = true
			default:
				return tacticalConfiguredProjectile{}, fmt.Errorf("Torpedo weapon %q has unsupported runtime modifier %q", weapon.WeaponID, modifier)
			}
		}
	}
	if configured.Speed <= 0 || configured.UnitDurability <= 0 || configured.WarheadStrength <= 0 {
		return tacticalConfiguredProjectile{}, fmt.Errorf("weapon %q derived invalid projectile values", weapon.WeaponID)
	}
	return configured, nil
}

func tacticalProjectileImpactSource(projectile TacticalProjectileState, target TacticalShipState) TacticalShipState {
	angle := float64(projectile.Facing) * (2 * math.Pi / 16)
	dx := int(math.Round(math.Cos(angle) * 1000))
	dy := int(math.Round(math.Sin(angle) * 1000))
	if dx == 0 && dy == 0 {
		dx = 1
	}
	return TacticalShipState{X: target.X - dx, Y: target.Y - dy}
}

func tacticalProjectileEffectiveMissileDefense(source, target TacticalShipSpec, projectile TacticalProjectileState) int {
	defense := target.MissileDefense - source.MissileScannerReduction
	if defense < 0 {
		defense = 0
	}
	for _, modifier := range projectile.ModifierIDs {
		if modifier == core.ShipWeaponModifierECCM {
			defense /= 2
			break
		}
	}
	return defense
}

func tacticalProjectileImpactWarhead(spec TacticalSpec, projectile TacticalProjectileState) (int, bool, error) {
	if projectile.WeaponID != "plasma_torpedoes" {
		return projectile.WarheadStrength, false, nil
	}
	if spec.Rules.Ordnance == nil || spec.Rules.Ordnance.MovementUnitsPerField <= 0 {
		return 0, false, fmt.Errorf("Plasma projectile %d has no movement-unit rule", projectile.ProjectileID)
	}
	noRangeDissipation := false
	for _, modifier := range projectile.ModifierIDs {
		if modifier == core.ShipWeaponModifierNoRangeDissipation {
			noRangeDissipation = true
			break
		}
	}
	strength := projectile.WarheadStrength
	if !noRangeDissipation {
		completedFields := projectile.DistanceTraveled / spec.Rules.Ordnance.MovementUnitsPerField
		strength -= completedFields * spec.Rules.Ordnance.PlasmaDissipationPerField
	}
	return strength, strength <= 0, nil
}

func resolveTacticalProjectileImpact(spec TacticalSpec, r *tacticalRuntime, projectile *TacticalProjectileState, commandSequence uint32) (tacticalProjectileQuantumOutcome, error) {
	if projectile == nil {
		return tacticalProjectileQuantumOutcome{}, fmt.Errorf("projectile impact requires tactical spec and projectile")
	}
	source := tacticalShipSpec(spec, projectile.SourceShipID)
	target := tacticalShipSpec(spec, projectile.TargetShipID)
	targetState := r.shipState(projectile.TargetShipID)
	if source == nil || target == nil || targetState == nil || targetState.Destroyed {
		return tacticalProjectileQuantumOutcome{}, fmt.Errorf("projectile %d impact target/source is unavailable", projectile.ProjectileID)
	}
	intercepted, err := resolveReactiveProjectilePointDefense(spec, r, projectile, target)
	if err != nil {
		return tacticalProjectileQuantumOutcome{}, err
	}
	if intercepted {
		return tacticalProjectileQuantumOutcome{DestroyReason: "intercepted"}, nil
	}

	weapon := weaponSpec(*source, projectile.WeaponSlot)
	if weapon == nil {
		return tacticalProjectileQuantumOutcome{}, fmt.Errorf("projectile %d source weapon is unavailable", projectile.ProjectileID)
	}
	configured, err := deriveConfiguredTacticalProjectile(spec, *weapon)
	if err != nil {
		return tacticalProjectileQuantumOutcome{}, err
	}

	if configured.MIRV {
		before := projectile.Multiplicity
		projectile.Multiplicity *= spec.Rules.Ordnance.MIRVMultiplier
		if err := r.appendEvent("projectile_mirv_split", projectile.SourceSeatID, commandSequence, map[string]any{
			"projectile_id":       projectile.ProjectileID,
			"multiplicity_before": before,
			"multiplicity_after":  projectile.Multiplicity,
		}); err != nil {
			return tacticalProjectileQuantumOutcome{}, err
		}
	}

	impactWarhead, dissipated, err := tacticalProjectileImpactWarhead(spec, *projectile)
	if err != nil {
		return tacticalProjectileQuantumOutcome{}, err
	}
	if dissipated {
		return tacticalProjectileQuantumOutcome{DestroyReason: "dissipated"}, nil
	}

	effectiveDefense := tacticalProjectileEffectiveMissileDefense(*source, *target, *projectile)
	impactCount := projectile.Multiplicity
	targetWasDestroyed := targetState.Destroyed
	damageSource := tacticalProjectileImpactSource(*projectile, *targetState)

	for impactIndex := 1; impactIndex <= impactCount; impactIndex++ {
		rawRoll, err := r.random(100, spec.Rules)
		if err != nil {
			return tacticalProjectileQuantumOutcome{}, err
		}
		hit := rawRoll > effectiveDefense
		damageApplied := 0
		var applications []tacticalDamageApplication
		if hit {
			applicationCount := 1
			if configured.Enveloping {
				applicationCount = 4
			}
			applications = make([]tacticalDamageApplication, 0, applicationCount)
			for applicationIndex := 0; applicationIndex < applicationCount; applicationIndex++ {
				options := tacticalDamageOptions{}
				if configured.Enveloping {
					options.HasShieldArcOverride = true
					options.ShieldArcOverride = applicationIndex
				}
				applied := applyTacticalDamageWithOptions(*target, targetState, damageSource, impactWarhead, options)
				applications = append(applications, applied)
				damageApplied += impactWarhead
			}
		}

		if err := r.appendEvent("projectile_impact_resolved", projectile.SourceSeatID, commandSequence, map[string]any{
			"projectile_id":             projectile.ProjectileID,
			"impact_index":              impactIndex,
			"impact_count":              impactCount,
			"target_ship_id":            target.ShipID,
			"effective_missile_defense": effectiveDefense,
			"raw_hit_roll":              rawRoll,
			"hit":                       hit,
			"warhead_strength":          impactWarhead,
			"damage_applied":            damageApplied,
		}); err != nil {
			return tacticalProjectileQuantumOutcome{}, err
		}

		for applicationIndex, applied := range applications {
			event := map[string]any{
				"source_ship_id": projectile.SourceShipID, "target_ship_id": target.ShipID,
				"weapon_slot": projectile.WeaponSlot, "weapon_id": projectile.WeaponID,
				"modifier_ids":  append([]core.ShipWeaponModifierID(nil), projectile.ModifierIDs...),
				"projectile_id": projectile.ProjectileID,
				"impact_index":  impactIndex, "impact_count": impactCount,
				"application_index": applicationIndex + 1, "application_count": len(applications),
				"layer": applied.Layer, "damage": impactWarhead, "incoming_damage": applied.IncomingDamage,
				"shield_arc": applied.ShieldArc, "shield_before": applied.ShieldBefore, "shield_after": applied.ShieldAfter,
				"post_block_damage": applied.PostBlockDamage, "shield_absorbed": applied.ShieldAbsorbed,
				"lower_layer_overflow": applied.LowerLayerOverflow,
				"armor_before":         applied.ArmorBefore, "armor_after": applied.ArmorAfter,
				"structure_damage_before": applied.StructureBefore, "structure_damage_after": applied.StructureAfter,
				"extra_structure_damage": applied.ExtraStructureDamage,
			}
			if err := r.appendEvent("battle_damage_applied", projectile.SourceSeatID, commandSequence, event); err != nil {
				return tacticalProjectileQuantumOutcome{}, err
			}
		}
	}

	outcome := tacticalProjectileQuantumOutcome{DestroyReason: "impact_complete"}
	if !targetWasDestroyed && targetState.Destroyed {
		outcome.DestroyedShipID = target.ShipID
		if err := r.appendEvent("ship_destroyed", projectile.SourceSeatID, commandSequence, map[string]any{"ship_id": target.ShipID}); err != nil {
			return tacticalProjectileQuantumOutcome{}, err
		}
	}
	return outcome, nil
}
