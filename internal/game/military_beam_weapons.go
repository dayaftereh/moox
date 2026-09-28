package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
	"moox/internal/ruleset"
)

const (
	MilitaryDesignReasonWeaponUnknown                    = "weapon_unknown"
	MilitaryDesignReasonWeaponTechnologyRequired         = "weapon_technology_required"
	MilitaryDesignReasonWeaponSlotOutOfRange             = "weapon_slot_out_of_range"
	MilitaryDesignReasonWeaponSlotDuplicate              = "weapon_slot_duplicate"
	MilitaryDesignReasonWeaponCountInvalid               = "weapon_count_invalid"
	MilitaryDesignReasonWeaponMountNotAllowed            = "weapon_mount_not_allowed"
	MilitaryDesignReasonWeaponModifierUnknown            = "weapon_modifier_unknown"
	MilitaryDesignReasonWeaponModifierNotAllowed         = "weapon_modifier_not_allowed"
	MilitaryDesignReasonWeaponModifierRefinementRequired = "weapon_modifier_refinement_required"
	MilitaryDesignReasonWeaponModifierDuplicate          = "weapon_modifier_duplicate"
	MilitaryDesignReasonWeaponAmmoRequired               = "weapon_ammo_required"
	MilitaryDesignReasonWeaponAmmoInvalid                = "weapon_ammo_invalid"
	MilitaryDesignReasonWeaponAmmoNotAllowed             = "weapon_ammo_not_allowed"
	MilitaryDesignReasonWeaponFiringArcUnknown           = "weapon_firing_arc_unknown"
	MilitaryDesignReasonWeaponFiringArcNotAllowed        = "weapon_firing_arc_not_allowed"
)

type militaryBeamWeaponQuote struct {
	Mount                core.ShipWeaponMount
	Weapon               ruleset.TacticalWeaponRule
	MiniaturizationLevel int
	ModifierMask         int
	UnitSpace            int
	UnitCostPP           int
	RowSpace             int
	RowCostPP            int
}

var frozenBeamWeaponIDs = map[string]struct{}{
	"mass_driver":      {},
	"gauss_cannon":     {},
	"laser_cannon":     {},
	"particle_beam":    {},
	"fusion_beam":      {},
	"ion_pulse_cannon": {},
	"graviton_beam":    {},
	"neutron_blaster":  {},
	"phasor":           {},
	"disrupter_cannon": {},
	"death_ray":        {},
	"plasma_cannon":    {},
	"mauler_device":    {},
}

func isFrozenBeamWeaponID(id string) bool {
	_, ok := frozenBeamWeaponIDs[id]
	return ok
}
func canonicalWeaponMountType(value core.ShipWeaponMountType) core.ShipWeaponMountType {
	if value == "" {
		return core.ShipWeaponMountStandard
	}
	return value
}

func canonicalWeaponModifierIDs(ids []core.ShipWeaponModifierID) ([]core.ShipWeaponModifierID, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	out := append([]core.ShipWeaponModifierID(nil), ids...)
	rank := func(id core.ShipWeaponModifierID) int {
		switch id {
		case core.ShipWeaponModifierArmorPiercing:
			return 0
		case core.ShipWeaponModifierContinuous:
			return 1
		case core.ShipWeaponModifierNoRangeDissipation:
			return 2
		case core.ShipWeaponModifierShieldPiercing:
			return 3
		case core.ShipWeaponModifierAutofire:
			return 4
		case core.ShipWeaponModifierEnveloping:
			return 5
		case core.ShipWeaponModifierMIRV:
			return 6
		case core.ShipWeaponModifierECCM:
			return 7
		case core.ShipWeaponModifierArmored:
			return 8
		case core.ShipWeaponModifierFast:
			return 9
		case core.ShipWeaponModifierEmissionsGuidance:
			return 10
		case core.ShipWeaponModifierOverloaded:
			return 11
		default:
			return -1
		}
	}
	for _, id := range out {
		if rank(id) < 0 {
			return nil, militaryDesignError(MilitaryDesignReasonWeaponModifierUnknown, "unknown weapon modifier %q", id)
		}
	}
	sort.Slice(out, func(i, j int) bool { return rank(out[i]) < rank(out[j]) })
	for i := 1; i < len(out); i++ {
		if out[i] == out[i-1] {
			return nil, militaryDesignError(MilitaryDesignReasonWeaponModifierDuplicate, "duplicate weapon modifier %q", out[i])
		}
	}
	return out, nil
}

func canonicalWeaponMount(mount core.ShipWeaponMount) (core.ShipWeaponMount, error) {
	mount.MountType = canonicalWeaponMountType(mount.MountType)
	switch mount.MountType {
	case core.ShipWeaponMountStandard, core.ShipWeaponMountHeavy, core.ShipWeaponMountPointDefense:
	default:
		return core.ShipWeaponMount{}, militaryDesignError(MilitaryDesignReasonWeaponMountNotAllowed, "unknown weapon mount type %q", mount.MountType)
	}
	return mount, nil
}

func canonicalWeaponMounts(mounts []core.ShipWeaponMount) ([]core.ShipWeaponMount, error) {
	if len(mounts) > 8 {
		return nil, fmt.Errorf("military design supports at most 8 weapon mounts")
	}
	out := core.CloneShipWeaponMounts(mounts)
	for i := range out {
		if out[i].Slot < 0 || out[i].Slot > 7 {
			return nil, militaryDesignError(MilitaryDesignReasonWeaponSlotOutOfRange, "weapon slot %d is outside 0..7", out[i].Slot)
		}
		if out[i].Count <= 0 {
			return nil, militaryDesignError(MilitaryDesignReasonWeaponCountInvalid, "weapon count must be positive")
		}
		mount, err := canonicalWeaponMount(out[i])
		if err != nil {
			return nil, err
		}
		out[i] = mount
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Slot < out[j].Slot })
	for i := 1; i < len(out); i++ {
		if out[i].Slot == out[i-1].Slot {
			return nil, militaryDesignError(MilitaryDesignReasonWeaponSlotDuplicate, "duplicate weapon slot %d", out[i].Slot)
		}
	}
	for i := range out {
		if !isFrozenBeamWeaponID(out[i].WeaponID) && !isFrozenOrdnanceWeaponID(out[i].WeaponID) && !isFrozenFighterWeaponID(out[i].WeaponID) {
			return nil, militaryDesignError(MilitaryDesignReasonWeaponUnknown, "unknown supported weapon %q", out[i].WeaponID)
		}
	}
	return out, nil
}

func beamMountBit(mountType core.ShipWeaponMountType) int {
	switch canonicalWeaponMountType(mountType) {
	case core.ShipWeaponMountHeavy:
		return 0x02
	case core.ShipWeaponMountPointDefense:
		return 0x04
	default:
		return 0
	}
}

func beamMountPercent(mountType core.ShipWeaponMountType) int {
	switch canonicalWeaponMountType(mountType) {
	case core.ShipWeaponMountHeavy:
		return 200
	case core.ShipWeaponMountPointDefense:
		return 50
	default:
		return 100
	}
}

func canonicalBeamFiringArc(id core.ShipWeaponFiringArcID) (core.ShipWeaponFiringArcID, int, error) {
	if id == "" {
		id = core.ShipWeaponFiringArcForward
	}
	switch id {
	case core.ShipWeaponFiringArcForward, core.ShipWeaponFiringArcRear:
		return id, 0, nil
	case core.ShipWeaponFiringArcForwardExtended, core.ShipWeaponFiringArcBackExtended:
		return id, 25, nil
	case core.ShipWeaponFiringArc360:
		return id, 50, nil
	default:
		return "", 0, militaryDesignError(MilitaryDesignReasonWeaponFiringArcUnknown, "unknown Beam firing arc %q", id)
	}
}

func beamModifierByID(items []ruleset.TacticalBeamModifierRule, id core.ShipWeaponModifierID) (ruleset.TacticalBeamModifierRule, bool) {
	for _, item := range items {
		if item.ID == string(id) {
			return item, true
		}
	}
	return ruleset.TacticalBeamModifierRule{}, false
}

func (r *EconomyRules) quoteBeamWeapon(empire *core.Empire, input core.ShipWeaponMount) (militaryBeamWeaponQuote, error) {
	if r == nil || r.TacticalCombat == nil || empire == nil {
		return militaryBeamWeaponQuote{}, fmt.Errorf("Beam weapon quote requires rules and empire")
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
	if !ok {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponUnknown, "unknown Beam weapon %q", mount.WeaponID)
	}
	if !empireKnowsTechnology(empire, weapon.TechnologyID) {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponTechnologyRequired, "empire %d has no technology %d for %q", empire.ID, weapon.TechnologyID, weapon.ID)
	}
	mask := beamMountBit(mount.MountType)
	if mask != 0 && weapon.AllowedModifierMask&mask == 0 {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponMountNotAllowed, "%q does not allow mount %q", weapon.ID, mount.MountType)
	}
	if mount.Ammo != 0 {
		return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponAmmoNotAllowed, "Beam weapon %q does not use ordnance ammo/rack state", weapon.ID)
	}
	arcID, arcBonus, err := canonicalBeamFiringArc(mount.FiringArcID)
	if err != nil {
		return militaryBeamWeaponQuote{}, err
	}
	mount.FiringArcID = arcID
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
	for _, modifierID := range mount.ModifierIDs {
		modifier, ok := beamModifierByID(r.TacticalCombat.BeamModifiers, modifierID)
		if !ok || modifier.Kind != "special" {
			return militaryBeamWeaponQuote{}, militaryDesignError(MilitaryDesignReasonWeaponModifierUnknown, "unknown Beam special modifier %q", modifierID)
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

	mountPercent := beamMountPercent(mount.MountType)
	afterMountSpace := (weapon.BaseSpace * mountPercent) / 100
	afterMountCost := (weapon.BaseCostPP * mountPercent) / 100
	afterSpecialSpace := (afterMountSpace * (100 + specialDeltaSpace + arcBonus)) / 100
	afterSpecialCost := (afterMountCost * (100 + specialDeltaCost + arcBonus)) / 100
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
