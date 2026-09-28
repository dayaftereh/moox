package game

import (
	"fmt"

	"moox/internal/core"
)

var frozenFighterWeaponIDs = map[string]struct{}{
	"heavy_fighter_bays": {},
	"bomber_bays":        {},
	"fighter_bays":       {},
}

func isFrozenFighterWeaponID(id string) bool {
	_, ok := frozenFighterWeaponIDs[id]
	return ok
}

func (r *EconomyRules) quoteFighterWeapon(empire *core.Empire, input core.ShipWeaponMount) (militaryBeamWeaponQuote, error) {
	if r == nil || r.TacticalCombat == nil {
		return militaryBeamWeaponQuote{}, fmt.Errorf("fighter quote requires tactical combat rules")
	}
	mount, err := canonicalWeaponMount(input)
	if err != nil {
		return militaryBeamWeaponQuote{}, err
	}
	if mount.Slot < 0 || mount.Slot > 7 {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponSlotOutOfRange, "weapon slot %d is outside 0..7", mount.Slot)
	}
	if mount.Count <= 0 {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponCountInvalid, "weapon count must be positive")
	}
	weapon, ok := r.TacticalCombat.WeaponByID(mount.WeaponID)
	if !ok || weapon.Kind != "fighter" || !isFrozenFighterWeaponID(mount.WeaponID) {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponUnknown, "unknown Fighter weapon %q", mount.WeaponID)
	}
	if !empireKnowsTechnology(empire, weapon.TechnologyID) {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponTechnologyRequired, "empire %d has no technology %d for %q", empire.ID, weapon.TechnologyID, weapon.ID)
	}
	if mount.MountType != core.ShipWeaponMountStandard {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponMountNotAllowed, "%q requires standard Fighter mount", weapon.ID)
	}
	if len(mount.ModifierIDs) != 0 {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponModifierNotAllowed, "%q does not allow weapon modifiers", weapon.ID)
	}
	if mount.Ammo != 0 {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponAmmoNotAllowed, "%q does not use ordnance ammo/rack state", weapon.ID)
	}
	if mount.FiringArcID != "" {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponFiringArcNotAllowed, "%q does not use a firing arc", weapon.ID)
	}
	level, err := r.shipComponentMiniaturizationLevel(empire, weapon.TechnologyID)
	if err != nil {
		return militaryBeamWeaponQuote{}, err
	}
	unitSpace := miniaturizedComponentSpace(weapon.BaseSpace, level, 1)
	unitCost := miniaturizedComponentCost(weapon.BaseCostPP, level)
	return militaryBeamWeaponQuote{
		Mount:                mount,
		Weapon:               weapon,
		MiniaturizationLevel: level,
		UnitSpace:            unitSpace,
		UnitCostPP:           unitCost,
		RowSpace:             unitSpace * mount.Count,
		RowCostPP:            unitCost * mount.Count,
	}, nil
}
