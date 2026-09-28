package game

import (
	"sort"

	"moox/internal/core"
	"moox/internal/ruleset"
)

const (
	MilitaryDesignReasonSpecialUnknown            = "special_unknown"
	MilitaryDesignReasonSpecialTechnologyRequired = "special_technology_required"
	MilitaryDesignReasonSpecialSlotOutOfRange     = "special_slot_out_of_range"
	MilitaryDesignReasonSpecialSlotDuplicate      = "special_slot_duplicate"
	MilitaryDesignReasonSpecialDuplicate          = "special_duplicate"
	MilitaryDesignReasonSpecialExclusionConflict  = "special_exclusion_conflict"
	MilitaryDesignReasonSpecialRequiresShield     = "special_requires_shield"
	MilitaryDesignReasonSpecialNotSupported       = "special_not_supported"
)

type militarySpecialQuote struct {
	Mount                core.ShipSpecialMount
	Special              ruleset.ShipSpecialRule
	MiniaturizationLevel int
	UnitSpace            int
	UnitCostPP           int
}

func canonicalShipSpecialMounts(in []core.ShipSpecialMount) ([]core.ShipSpecialMount, error) {
	if len(in) == 0 {
		return nil, nil
	}
	if len(in) > 8 {
		return nil, militaryDesignError(MilitaryDesignReasonSpecialSlotOutOfRange, "special mount count %d exceeds 8", len(in))
	}
	out := core.CloneShipSpecialMounts(in)
	sort.Slice(out, func(i, j int) bool { return out[i].Slot < out[j].Slot })
	seenIDs := make(map[string]struct{}, len(out))
	for i := range out {
		if out[i].Slot < 0 || out[i].Slot > 7 {
			return nil, militaryDesignError(MilitaryDesignReasonSpecialSlotOutOfRange, "special slot %d is outside 0..7", out[i].Slot)
		}
		if i > 0 && out[i].Slot == out[i-1].Slot {
			return nil, militaryDesignError(MilitaryDesignReasonSpecialSlotDuplicate, "duplicate special slot %d", out[i].Slot)
		}
		if out[i].SpecialID == "" {
			return nil, militaryDesignError(MilitaryDesignReasonSpecialUnknown, "special_id must not be empty")
		}
		if _, ok := seenIDs[out[i].SpecialID]; ok {
			return nil, militaryDesignError(MilitaryDesignReasonSpecialDuplicate, "duplicate special %q", out[i].SpecialID)
		}
		seenIDs[out[i].SpecialID] = struct{}{}
	}
	return out, nil
}

func (r *EconomyRules) quoteShipSpecials(empire *core.Empire, hullIndex int, shieldID string, in []core.ShipSpecialMount) ([]core.ShipSpecialMount, []militarySpecialQuote, int, int, error) {
	mounts, err := canonicalShipSpecialMounts(in)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	if len(mounts) == 0 {
		return nil, nil, 0, 0, nil
	}
	if r == nil || r.ShipSpecials == nil {
		return nil, nil, 0, 0, militaryDesignError(MilitaryDesignReasonSpecialNotSupported, "ship special rules are unavailable")
	}
	if hullIndex < 0 || hullIndex >= 6 {
		return nil, nil, 0, 0, militaryDesignError(MilitaryDesignReasonSpecialNotSupported, "hull index %d is outside supported special table", hullIndex)
	}

	installed := make(map[int]ruleset.ShipSpecialRule, len(mounts))
	rules := make([]ruleset.ShipSpecialRule, len(mounts))
	for i, mount := range mounts {
		if mount.SpecialID == "regeneration" {
			return nil, nil, 0, 0, militaryDesignError(MilitaryDesignReasonSpecialNotSupported, "special %q is not a normal Designer choice", mount.SpecialID)
		}
		special, ok := r.ShipSpecials.ByID(mount.SpecialID)
		if !ok {
			return nil, nil, 0, 0, militaryDesignError(MilitaryDesignReasonSpecialUnknown, "unknown special %q", mount.SpecialID)
		}
		if !empireKnowsTechnology(empire, special.TechnologyID) {
			return nil, nil, 0, 0, militaryDesignError(MilitaryDesignReasonSpecialTechnologyRequired, "empire %d has no technology %d for %q", empire.ID, special.TechnologyID, special.ID)
		}
		if special.RequiresShield && shieldID == "" {
			return nil, nil, 0, 0, militaryDesignError(MilitaryDesignReasonSpecialRequiresShield, "%q requires a primary shield", special.ID)
		}
		rules[i] = special
		installed[special.OriginalID] = special
	}
	for _, special := range rules {
		for _, excludedID := range special.ExcludedOriginalIDs {
			if excludedID == special.OriginalID {
				continue
			}
			if other, ok := installed[excludedID]; ok {
				return nil, nil, 0, 0, militaryDesignError(MilitaryDesignReasonSpecialExclusionConflict, "%q conflicts with %q", special.ID, other.ID)
			}
		}
	}

	quotes := make([]militarySpecialQuote, 0, len(mounts))
	totalSpace, totalCost := 0, 0
	for i, mount := range mounts {
		special := rules[i]
		level := 0
		space := special.BaseSpaceByHull[hullIndex]
		cost := special.BaseCostByHullPP[hullIndex]
		if special.Miniaturizes {
			level, err = r.shipComponentMiniaturizationLevel(empire, special.TechnologyID)
			if err != nil {
				return nil, nil, 0, 0, err
			}
			space = miniaturizedComponentSpace(space, level, special.MiniaturizationClass)
			cost = miniaturizedComponentCost(cost, level)
		}
		totalSpace += space
		totalCost += cost
		quotes = append(quotes, militarySpecialQuote{
			Mount: mount, Special: special, MiniaturizationLevel: level,
			UnitSpace: space, UnitCostPP: cost,
		})
	}
	return mounts, quotes, totalSpace, totalCost, nil
}

func shipSpecialInstalled(mounts []core.ShipSpecialMount, id string) bool {
	for _, mount := range mounts {
		if mount.SpecialID == id {
			return true
		}
	}
	return false
}
