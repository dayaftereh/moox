package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

const (
	ShipDesignerLockTechnologyRequired = "technology_required"
	ShipDesignerLockCurrentScope       = "current_design_scope"
)

var shipDesignerHullTechnologyKeys = map[string]string{
	"titan":     "titan_construction",
	"doom_star": "doom_star_construction",
}

type MilitaryDesignerHullChoice struct {
	ID                    string `json:"id"`
	SizeIndex             int    `json:"size_index"`
	NameKey               string `json:"name_key"`
	BaseCostPP            int    `json:"base_cost_pp"`
	BaseSpace             int    `json:"base_space"`
	CommandPointCost      int    `json:"command_point_cost"`
	StrategicPictureIDs   []int  `json:"strategic_picture_ids"`
	RequiredTechnologyID  int    `json:"required_technology_id,omitempty"`
	RequiredTechnologyKey string `json:"required_technology_key,omitempty"`
	TechnologyKnown       bool   `json:"technology_known"`
	SaveAvailable         bool   `json:"save_available"`
	LockReason            string `json:"lock_reason,omitempty"`
}

type MilitaryDesignerWeaponChoice struct {
	ID              string `json:"id"`
	NameKey         string `json:"name_key"`
	Kind            string `json:"kind"`
	TechnologyID    int    `json:"technology_id"`
	TechnologyKnown bool   `json:"technology_known"`
	Available       bool   `json:"available"`
	LockReason      string `json:"lock_reason,omitempty"`
	BaseSpace       int    `json:"base_space"`
	BaseCostPP      int    `json:"base_cost_pp"`
}

type MilitaryDesignerVariant struct {
	Key              string                 `json:"key"`
	HullID           string                 `json:"hull_id"`
	Weapons          []core.ShipWeaponMount `json:"weapons,omitempty"`
	Spec             core.ShipDesignSpec    `json:"spec"`
	CommandPointCost int                    `json:"command_point_cost"`
}

type MilitaryDesignerCatalog struct {
	Hulls    []MilitaryDesignerHullChoice   `json:"hulls"`
	Weapons  []MilitaryDesignerWeaponChoice `json:"weapons"`
	Variants []MilitaryDesignerVariant      `json:"variants"`
}

func (r *EconomyRules) MilitaryDesignerCatalog(empire *core.Empire) (MilitaryDesignerCatalog, error) {
	if r == nil || empire == nil {
		return MilitaryDesignerCatalog{}, fmt.Errorf("military designer requires rules and empire")
	}

	hulls := make([]MilitaryDesignerHullChoice, 0, len(r.ShipHulls))
	for _, hull := range r.ShipHulls {
		technologyID, technologyKey, err := r.shipDesignerHullTechnology(hull.ID)
		if err != nil {
			return MilitaryDesignerCatalog{}, err
		}
		technologyKnown := technologyID == 0 || empireKnowsTechnology(empire, technologyID)
		saveAvailable := technologyKnown && hull.ID == SupportedMilitaryHullID
		lockReason := ""
		if !technologyKnown {
			lockReason = ShipDesignerLockTechnologyRequired
		} else if !saveAvailable {
			lockReason = ShipDesignerLockCurrentScope
		}
		hulls = append(hulls, MilitaryDesignerHullChoice{
			ID: hull.ID, SizeIndex: hull.SizeIndex, NameKey: hull.NameKey,
			BaseCostPP: hull.BaseCostPP, BaseSpace: hull.BaseSpace,
			CommandPointCost:     hull.SizeIndex + 1,
			StrategicPictureIDs:  append([]int(nil), hull.StrategicPictureIDs...),
			RequiredTechnologyID: technologyID, RequiredTechnologyKey: technologyKey,
			TechnologyKnown: technologyKnown, SaveAvailable: saveAvailable, LockReason: lockReason,
		})
	}
	sort.Slice(hulls, func(i, j int) bool {
		if hulls[i].SizeIndex != hulls[j].SizeIndex {
			return hulls[i].SizeIndex < hulls[j].SizeIndex
		}
		return hulls[i].ID < hulls[j].ID
	})

	catalog := MilitaryDesignerCatalog{Hulls: hulls}
	if r.TacticalCombat != nil {
		weapon := r.TacticalCombat.Weapon
		technologyKnown := empireKnowsTechnology(empire, weapon.TechnologyID)
		lockReason := ""
		if !technologyKnown {
			lockReason = ShipDesignerLockTechnologyRequired
		}
		catalog.Weapons = append(catalog.Weapons, MilitaryDesignerWeaponChoice{
			ID: weapon.ID, NameKey: r.TechnologyNameKeyByID[weapon.TechnologyID], Kind: weapon.Kind,
			TechnologyID: weapon.TechnologyID, TechnologyKnown: technologyKnown, Available: technologyKnown,
			LockReason: lockReason, BaseSpace: weapon.BaseSpace, BaseCostPP: weapon.BaseCostPP,
		})
	}

	frigate, ok := r.ShipHulls[SupportedMilitaryHullID]
	if !ok || len(frigate.StrategicPictureIDs) == 0 {
		return MilitaryDesignerCatalog{}, fmt.Errorf("supported military hull %q has no strategic picture", SupportedMilitaryHullID)
	}
	pictureID := frigate.StrategicPictureIDs[0]
	baseSpec, err := r.clearedMilitaryDesignSpec(empire, SupportedMilitaryHullID, pictureID, nil)
	if err != nil {
		return MilitaryDesignerCatalog{}, fmt.Errorf("military designer base preview: %w", err)
	}
	catalog.Variants = append(catalog.Variants, MilitaryDesignerVariant{
		Key: "frigate:none", HullID: SupportedMilitaryHullID, Spec: baseSpec,
		CommandPointCost: frigate.SizeIndex + 1,
	})
	if len(catalog.Weapons) != 0 && catalog.Weapons[0].Available {
		mounts := []core.ShipWeaponMount{{Slot: 0, WeaponID: catalog.Weapons[0].ID, Count: 1}}
		spec, err := r.clearedMilitaryDesignSpec(empire, SupportedMilitaryHullID, pictureID, mounts)
		if err != nil {
			return MilitaryDesignerCatalog{}, fmt.Errorf("military designer weapon preview: %w", err)
		}
		catalog.Variants = append(catalog.Variants, MilitaryDesignerVariant{
			Key: "frigate:" + catalog.Weapons[0].ID, HullID: SupportedMilitaryHullID,
			Weapons: append([]core.ShipWeaponMount(nil), mounts...), Spec: spec,
			CommandPointCost: frigate.SizeIndex + 1,
		})
	}
	return catalog, nil
}

func (r *EconomyRules) shipDesignerHullTechnology(hullID string) (int, string, error) {
	key := shipDesignerHullTechnologyKeys[hullID]
	if key == "" {
		return 0, "", nil
	}
	for id, candidate := range r.TechnologyKeyByID {
		if candidate == key {
			return id, key, nil
		}
	}
	return 0, key, fmt.Errorf("ship designer hull %q references missing technology %q", hullID, key)
}
