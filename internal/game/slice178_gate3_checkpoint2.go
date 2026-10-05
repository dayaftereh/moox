package game

import (
	"fmt"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/ruleset"
)

var slice178NonDamageableSpecialOriginalIDs = map[int]struct{}{
	4: {}, 11: {}, 14: {}, 24: {}, 27: {},
}

func tacticalSlice178SpecialInstalled(ship core.Ship, id string) bool {
	for _, special := range ship.Spec.Specials {
		if special.SpecialID == id {
			return true
		}
	}
	return false
}

func applySlice178PassiveFoundation(ship core.Ship, rules *ruleset.TacticalCombatFile, in tacticalFoundationValues) tacticalFoundationValues {
	out := in
	if tacticalSlice178SpecialInstalled(ship, "reinforced_hull") {
		out.StructureMax *= 3
	}
	if tacticalSlice178SpecialInstalled(ship, "heavy_armor") {
		out.ArmorMax *= 3
	}
	if tacticalSlice178SpecialInstalled(ship, "augmented_engines") {
		out.CurrentCombatSpeed += rules.Specials.AugmentedEnginesSpeedBonus
	}
	if tacticalSlice178SpecialInstalled(ship, "battle_scanner") {
		out.BeamOffense += rules.Specials.BattleScannerOffenseBonus
	}
	if tacticalSlice178SpecialInstalled(ship, "multi_phased_shields") && out.ShieldArcMax > 0 {
		out.ShieldArcMax = (out.ShieldArcMax * rules.Specials.MultiPhasedShieldNumerator) / rules.Specials.MultiPhasedShieldDenominator
	}
	out.BeamDefense = out.CurrentCombatSpeed * rules.BeamDefensePerCombatSpeed
	return out
}

func tacticalInternalSpecForShip(ship core.Ship, rules *ruleset.TacticalCombatFile, specialRules *ruleset.ShipSpecialsFile) (*battle.TacticalInternalSpec, error) {
	if rules == nil {
		return nil, fmt.Errorf("tactical internal projection requires tactical rules")
	}
	hullIndex := -1
	var hull ruleset.TacticalHullRule
	for i, candidate := range rules.Hulls {
		if candidate.HullID == ship.Spec.HullID {
			hullIndex = i
			hull = candidate
			break
		}
	}
	if hullIndex < 0 {
		return nil, fmt.Errorf("tactical internal projection has no hull %q", ship.Spec.HullID)
	}
	out := &battle.TacticalInternalSpec{
		EngineMax:    hull.InternalEngineMax,
		SecondaryMax: hull.InternalSecondaryMax,
	}
	if tacticalSlice178SpecialInstalled(ship, "reinforced_hull") {
		out.EngineMax *= 3
	}
	if ship.Spec.ShieldID != "" {
		out.ShieldGeneratorMax = hull.InternalShieldGeneratorMax
	}
	mounts, err := canonicalWeaponMounts(ship.Spec.Weapons)
	if err != nil {
		return nil, err
	}
	for _, mount := range mounts {
		unitSize, err := tacticalInternalWeaponUnitSize(rules, mount)
		if err != nil {
			return nil, err
		}
		out.Weapons = append(out.Weapons, battle.TacticalInternalWeaponSpec{
			Slot: mount.Slot, Max: mount.Count * unitSize, UnitSize: unitSize,
		})
	}
	for _, mount := range ship.Spec.Specials {
		if specialRules == nil {
			continue
		}
		special, ok := specialRules.ByID(mount.SpecialID)
		if !ok {
			return nil, fmt.Errorf("tactical internal projection has unknown special %q", mount.SpecialID)
		}
		if _, excluded := slice178NonDamageableSpecialOriginalIDs[special.OriginalID]; excluded {
			continue
		}
		if hullIndex >= len(special.BaseSpaceByHull) {
			return nil, fmt.Errorf("tactical internal projection special %q has no hull %d space", special.ID, hullIndex)
		}
		maxValue := special.BaseSpaceByHull[hullIndex] / 10
		if maxValue < 1 {
			maxValue = 1
		}
		out.Specials = append(out.Specials, battle.TacticalInternalSpecialSpec{Slot: mount.Slot, Max: maxValue})
	}
	return out, nil
}

func tacticalInternalWeaponUnitSize(rules *ruleset.TacticalCombatFile, input core.ShipWeaponMount) (int, error) {
	mount, err := canonicalWeaponMount(input)
	if err != nil {
		return 0, err
	}
	weapon, ok := rules.WeaponByID(mount.WeaponID)
	if !ok {
		return 0, fmt.Errorf("tactical internal projection has unknown weapon %q", mount.WeaponID)
	}
	configuredSpace := weapon.BaseSpace
	switch weapon.Kind {
	case "beam":
		_, arcBonus, err := canonicalBeamFiringArc(mount.FiringArcID)
		if err != nil {
			return 0, err
		}
		delta := 0
		for _, modifierID := range mount.ModifierIDs {
			modifier, ok := beamModifierByID(rules.BeamModifiers, modifierID)
			if !ok || modifier.Kind != "special" {
				return 0, fmt.Errorf("tactical internal projection Beam %q has unknown modifier %q", weapon.ID, modifierID)
			}
			delta += modifier.SpacePercentDelta
		}
		configuredSpace = (weapon.BaseSpace * beamMountPercent(mount.MountType)) / 100
		configuredSpace = (configuredSpace * (100 + delta + arcBonus)) / 100
	case "missile", "torpedo", "bomb":
		rackBase := 0
		if weapon.Kind == "missile" {
			var legal bool
			rackBase, legal = ordnanceRackBase(mount.Ammo)
			if !legal {
				return 0, fmt.Errorf("tactical internal projection missile %q has invalid rack ammo %d", weapon.ID, mount.Ammo)
			}
		}
		delta := 0
		for _, modifierID := range mount.ModifierIDs {
			modifier, ok := beamModifierByID(rules.WeaponModifiers(), modifierID)
			if !ok || modifier.Kind != "special" {
				return 0, fmt.Errorf("tactical internal projection ordnance %q has unknown modifier %q", weapon.ID, modifierID)
			}
			delta += modifier.SpacePercentDelta
		}
		configuredSpace = ((weapon.BaseSpace + rackBase) * (100 + delta)) / 100
	case "fighter":
		configuredSpace = weapon.BaseSpace
	default:
		return 0, fmt.Errorf("tactical internal projection weapon %q has unsupported kind %q", weapon.ID, weapon.Kind)
	}
	unitSize := configuredSpace / 10
	if unitSize < 1 {
		unitSize = 1
	}
	return unitSize, nil
}
