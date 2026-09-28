package game

import (
	"fmt"

	"moox/internal/core"
)

var frozenOrdnanceWeaponIDs = map[string]struct{}{
	"nuclear_missile":       {},
	"merculite_missile":     {},
	"pulson_missile":        {},
	"zeon_missile":          {},
	"anti_matter_torpedoes": {},
	"proton_torpedoes":      {},
	"plasma_torpedoes":      {},
	"nuclear_bomb":          {},
	"fusion_bomb":           {},
	"anti_matter_bomb":      {},
	"neutronium_bomb":       {},
	"death_spores":          {},
	"bio_terminator":        {},
}

func isFrozenOrdnanceWeaponID(id string) bool {
	_, ok := frozenOrdnanceWeaponIDs[id]
	return ok
}

func ordnanceRackBase(ammo int) (int, bool) {
	switch ammo {
	case 2:
		return 10, true
	case 5:
		return 20, true
	case 10:
		return 30, true
	case 15:
		return 35, true
	case 20:
		return 40, true
	default:
		return 0, false
	}
}

func (r *EconomyRules) quoteOrdnanceWeapon(empire *core.Empire, input core.ShipWeaponMount) (militaryBeamWeaponQuote, error) {
	if r == nil || r.TacticalCombat == nil || empire == nil {
		return militaryBeamWeaponQuote{}, fmt.Errorf("ordnance weapon quote requires rules and empire")
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
	if !ok || !isFrozenOrdnanceWeaponID(mount.WeaponID) {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponUnknown, "unknown ordnance weapon %q", mount.WeaponID)
	}
	switch weapon.Kind {
	case "missile", "torpedo", "bomb":
	default:
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponUnknown, "%q is not supported ordnance", mount.WeaponID)
	}
	if !empireKnowsTechnology(empire, weapon.TechnologyID) {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponTechnologyRequired, "empire %d has no technology %d for %q", empire.ID, weapon.TechnologyID, weapon.ID)
	}
	if mount.MountType != core.ShipWeaponMountStandard {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponMountNotAllowed, "%q requires standard ordnance mount", weapon.ID)
	}

	rackBase := 0
	switch weapon.Kind {
	case "missile":
		if mount.Ammo == 0 {
			return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponAmmoRequired, "Missile %q requires rack Ammo", weapon.ID)
		}
		var legal bool
		rackBase, legal = ordnanceRackBase(mount.Ammo)
		if !legal {
			return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponAmmoInvalid, "Missile %q Ammo %d is not one of 2,5,10,15,20", weapon.ID, mount.Ammo)
		}
	case "torpedo":
		if mount.Ammo == 0 {
			mount.Ammo = 2
		} else if mount.Ammo != 2 {
			return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponAmmoNotAllowed, "Torpedo %q uses fixed Ammo 2", weapon.ID)
		}
	case "bomb":
		if mount.Ammo == 0 {
			mount.Ammo = 10
		} else if mount.Ammo != 10 {
			return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponAmmoNotAllowed, "Bomb %q uses fixed Ammo 10", weapon.ID)
		}
	}

	switch mount.FiringArcID {
	case "":
		mount.FiringArcID = core.ShipWeaponFiringArc360
	case core.ShipWeaponFiringArc360:
	default:
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponFiringArcNotAllowed, "%q supports only firing arc 360 in Slice 17.3", weapon.ID)
	}
	ids, err := canonicalWeaponModifierIDs(mount.ModifierIDs)
	if err != nil {
		return militaryBeamWeaponQuote{}, err
	}
	mount.ModifierIDs = ids

	level, err := r.shipComponentMiniaturizationLevel(empire, weapon.TechnologyID)
	if err != nil {
		return militaryBeamWeaponQuote{}, err
	}

	specialDeltaSpace := 0
	specialDeltaCost := 0
	mask := 0
	for _, modifierID := range mount.ModifierIDs {
		modifier, ok := beamModifierByID(r.TacticalCombat.WeaponModifiers(), modifierID)
		if !ok || modifier.Kind != "special" {
			return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponModifierUnknown, "unknown ordnance modifier %q", modifierID)
		}
		if weapon.AllowedModifierMask&modifier.Bit == 0 {
			return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponModifierNotAllowed, "%q does not allow modifier %q", weapon.ID, modifierID)
		}
		if level < modifier.RequiredMiniLevel {
			return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponModifierRefinementRequired, "%q requires miniaturization level %d for modifier %q", weapon.ID, modifier.RequiredMiniLevel, modifierID)
		}
		mask |= modifier.Bit
		specialDeltaSpace += modifier.SpacePercentDelta
		specialDeltaCost += modifier.CostPercentDelta
	}

	baseSpace := weapon.BaseSpace + rackBase
	baseCost := weapon.BaseCostPP + rackBase
	afterSpecialSpace := (baseSpace * (100 + specialDeltaSpace)) / 100
	afterSpecialCost := (baseCost * (100 + specialDeltaCost)) / 100
	unitSpace := miniaturizedComponentSpace(afterSpecialSpace, level, 1)
	unitCost := miniaturizedComponentCost(afterSpecialCost, level)

	return militaryBeamWeaponQuote{
		Mount:                mount,
		Weapon:               weapon,
		MiniaturizationLevel: level,
		ModifierMask:         mask,
		UnitSpace:            unitSpace,
		UnitCostPP:           unitCost,
		RowSpace:             unitSpace * mount.Count,
		RowCostPP:            unitCost * mount.Count,
	}, nil
}

func (r *EconomyRules) quoteWeapon(empire *core.Empire, input core.ShipWeaponMount) (militaryBeamWeaponQuote, error) {
	if r == nil || r.TacticalCombat == nil {
		return militaryBeamWeaponQuote{}, fmt.Errorf("weapon quote requires tactical combat rules")
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
	switch {
	case isFrozenBeamWeaponID(mount.WeaponID):
		return r.quoteBeamWeapon(empire, mount)
	case isFrozenOrdnanceWeaponID(mount.WeaponID):
		return r.quoteOrdnanceWeapon(empire, mount)
	case isFrozenFighterWeaponID(mount.WeaponID):
		return r.quoteFighterWeapon(empire, mount)
	default:
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponUnknown, "unknown supported weapon %q", mount.WeaponID)
	}
}
