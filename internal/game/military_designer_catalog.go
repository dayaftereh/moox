package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

const (
	ShipDesignerLockTechnologyRequired = MilitaryDesignReasonTechnologyRequired
	ShipDesignerLockCurrentScope       = "current_design_scope"
	ShipDesignerLockMandatorySystems   = MilitaryDesignReasonMandatorySystemsRequired
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

type MilitaryDesignerComponentChoice struct {
	ID                         string `json:"id"`
	NameKey                    string `json:"name_key,omitempty"`
	TechnologyID               int    `json:"technology_id"`
	TechnologyKnown            bool   `json:"technology_known"`
	Available                  bool   `json:"available"`
	LockReason                 string `json:"lock_reason,omitempty"`
	MiniaturizationLevel       int    `json:"miniaturization_level"`
	EffectiveCostByHullPP      []int  `json:"effective_cost_by_hull_pp,omitempty"`
	EffectiveSpaceByHull       []int  `json:"effective_space_by_hull,omitempty"`
	BeamAttackBonus            int    `json:"beam_attack_bonus,omitempty"`
	ShieldDamageBlockPerHit    int    `json:"shield_damage_block_per_hit,omitempty"`
	ShieldCapacityPerHullClass int    `json:"shield_capacity_per_hull_class,omitempty"`
}

type MilitaryDesignerAutomaticSystems struct {
	WarpDriveID             string `json:"warp_drive_id,omitempty"`
	FTLSpeed                int    `json:"ftl_speed,omitempty"`
	TacticalSpeedBonus      int    `json:"tactical_speed_bonus,omitempty"`
	ArmorID                 string `json:"armor_id,omitempty"`
	ArmorDurabilityMultiple int    `json:"armor_durability_multiple,omitempty"`
	FuelCellID              string `json:"fuel_cell_id,omitempty"`
	FuelRangeParsecs        int    `json:"fuel_range_parsecs,omitempty"`
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
	MinDamage       int    `json:"min_damage"`
	MaxDamage       int    `json:"max_damage"`
}

type MilitaryDesignerVariant struct {
	Key              string                 `json:"key"`
	HullID           string                 `json:"hull_id"`
	Weapons          []core.ShipWeaponMount `json:"weapons,omitempty"`
	Spec             core.ShipDesignSpec    `json:"spec"`
	CommandPointCost int                    `json:"command_point_cost"`
}

type MilitaryDesignerCatalog struct {
	Hulls                     []MilitaryDesignerHullChoice      `json:"hulls"`
	Computers                 []MilitaryDesignerComponentChoice `json:"computers"`
	Shields                   []MilitaryDesignerComponentChoice `json:"shields"`
	AutomaticSystems          MilitaryDesignerAutomaticSystems  `json:"automatic_systems"`
	Weapons                   []MilitaryDesignerWeaponChoice    `json:"weapons"`
	Variants                  []MilitaryDesignerVariant         `json:"variants"`
	ProductionCostNumerator   int                               `json:"production_cost_numerator"`
	ProductionCostDenominator int                               `json:"production_cost_denominator"`
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
		lockReason := ""
		if !technologyKnown {
			lockReason = ShipDesignerLockTechnologyRequired
		}
		hulls = append(hulls, MilitaryDesignerHullChoice{
			ID: hull.ID, SizeIndex: hull.SizeIndex, NameKey: hull.NameKey,
			BaseCostPP: hull.BaseCostPP, BaseSpace: hull.BaseSpace,
			CommandPointCost:     hull.SizeIndex + 1,
			StrategicPictureIDs:  append([]int(nil), hull.StrategicPictureIDs...),
			RequiredTechnologyID: technologyID, RequiredTechnologyKey: technologyKey,
			TechnologyKnown: technologyKnown, SaveAvailable: technologyKnown, LockReason: lockReason,
		})
	}
	sort.Slice(hulls, func(i, j int) bool {
		if hulls[i].SizeIndex != hulls[j].SizeIndex {
			return hulls[i].SizeIndex < hulls[j].SizeIndex
		}
		return hulls[i].ID < hulls[j].ID
	})

	catalog := MilitaryDesignerCatalog{
		Hulls: hulls, ProductionCostNumerator: 1, ProductionCostDenominator: 1,
	}
	if modifier, ok := r.RaceModifiers[empire.RaceID]; ok && modifier.GovernmentTraitID == "government_feudal" {
		catalog.ProductionCostNumerator = 2
		catalog.ProductionCostDenominator = 3
	}

	drive, hasDrive := bestKnownShipDrive(r.ShipDrives, empire)
	armor, hasArmor := bestKnownShipArmor(r.ShipArmors, empire)
	fuel, hasFuel := bestKnownShipFuelCell(r.ShipFuelCells, empire)
	if hasDrive {
		catalog.AutomaticSystems.WarpDriveID = drive.ID
		catalog.AutomaticSystems.FTLSpeed = drive.FTLSpeed
		catalog.AutomaticSystems.TacticalSpeedBonus = shipDriveTacticalSpeedBonus(drive.ComponentIndex)
	}
	if hasArmor {
		catalog.AutomaticSystems.ArmorID = armor.ID
		catalog.AutomaticSystems.ArmorDurabilityMultiple = shipArmorDurabilityMultiple(armor.ComponentIndex)
	}
	if hasFuel {
		catalog.AutomaticSystems.FuelCellID = fuel.ID
		catalog.AutomaticSystems.FuelRangeParsecs = fuel.RangeParsecs
	}
	if !hasDrive || !hasArmor || !hasFuel {
		for i := range catalog.Hulls {
			if catalog.Hulls[i].SaveAvailable {
				catalog.Hulls[i].SaveAvailable = false
				catalog.Hulls[i].LockReason = ShipDesignerLockMandatorySystems
			}
		}
	}

	for _, computer := range r.ShipComputers {
		known := empireKnowsTechnology(empire, computer.TechnologyID)
		level := 0
		if known {
			var err error
			level, err = r.shipComponentMiniaturizationLevel(empire, computer.TechnologyID)
			if err != nil {
				return MilitaryDesignerCatalog{}, err
			}
		}
		effective := make([]int, len(computer.CostByHullPP))
		for i, base := range computer.CostByHullPP {
			effective[i] = miniaturizedComponentCost(base, level)
		}
		lockReason := ""
		if !known {
			lockReason = ShipDesignerLockTechnologyRequired
		}
		catalog.Computers = append(catalog.Computers, MilitaryDesignerComponentChoice{
			ID: computer.ID, NameKey: r.TechnologyNameKeyByID[computer.TechnologyID],
			TechnologyID: computer.TechnologyID, TechnologyKnown: known, Available: known,
			LockReason: lockReason, MiniaturizationLevel: level, EffectiveCostByHullPP: effective,
			BeamAttackBonus: shipComputerBeamAttackBonus(computer.ComponentIndex),
		})
	}
	for _, shield := range r.ShipShields {
		known := empireKnowsTechnology(empire, shield.TechnologyID)
		level := 0
		if known {
			var err error
			level, err = r.shipComponentMiniaturizationLevel(empire, shield.TechnologyID)
			if err != nil {
				return MilitaryDesignerCatalog{}, err
			}
		}
		effectiveCost := make([]int, len(shield.CostByHullPP))
		effectiveSpace := make([]int, len(shield.SpaceByHull))
		for i, base := range shield.CostByHullPP {
			effectiveCost[i] = miniaturizedComponentCost(base, level)
		}
		for i, base := range shield.SpaceByHull {
			effectiveSpace[i] = miniaturizedComponentSpace(base, level, 0)
		}
		lockReason := ""
		if !known {
			lockReason = ShipDesignerLockTechnologyRequired
		}
		catalog.Shields = append(catalog.Shields, MilitaryDesignerComponentChoice{
			ID: shield.ID, NameKey: r.TechnologyNameKeyByID[shield.TechnologyID],
			TechnologyID: shield.TechnologyID, TechnologyKnown: known, Available: known,
			LockReason: lockReason, MiniaturizationLevel: level,
			EffectiveCostByHullPP: effectiveCost, EffectiveSpaceByHull: effectiveSpace,
			ShieldDamageBlockPerHit:    shipShieldDamageBlock(shield.ComponentIndex),
			ShieldCapacityPerHullClass: shipShieldCapacityPerHullClass(shield.ComponentIndex),
		})
	}

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
			MinDamage: weapon.MinDamage, MaxDamage: weapon.MaxDamage,
		})
	}

	for i := range catalog.Hulls {
		hullChoice := &catalog.Hulls[i]
		if !hullChoice.SaveAvailable {
			continue
		}
		hull := r.ShipHulls[hullChoice.ID]
		if len(hull.StrategicPictureIDs) == 0 {
			hullChoice.SaveAvailable = false
			hullChoice.LockReason = ShipDesignerLockMandatorySystems
			continue
		}
		pictureID := hull.StrategicPictureIDs[0]
		baseSpec, err := r.clearedMilitaryDesignSpec(empire, hull.ID, pictureID, nil)
		if err != nil {
			hullChoice.SaveAvailable = false
			hullChoice.LockReason = ShipDesignerLockMandatorySystems
			continue
		}
		catalog.Variants = append(catalog.Variants, MilitaryDesignerVariant{
			Key: hull.ID + ":none", HullID: hull.ID, Spec: baseSpec, CommandPointCost: hull.SizeIndex + 1,
		})
		if len(catalog.Weapons) != 0 && catalog.Weapons[0].Available {
			mounts := []core.ShipWeaponMount{{Slot: 0, WeaponID: catalog.Weapons[0].ID, Count: 1}}
			spec, err := r.clearedMilitaryDesignSpec(empire, hull.ID, pictureID, mounts)
			if err == nil {
				catalog.Variants = append(catalog.Variants, MilitaryDesignerVariant{
					Key: hull.ID + ":" + catalog.Weapons[0].ID, HullID: hull.ID,
					Weapons: append([]core.ShipWeaponMount(nil), mounts...), Spec: spec,
					CommandPointCost: hull.SizeIndex + 1,
				})
			}
		}
	}
	return catalog, nil
}

func shipComputerBeamAttackBonus(componentIndex int) int {
	if componentIndex <= 0 {
		return 0
	}
	return componentIndex * 25
}

func shipShieldDamageBlock(componentIndex int) int {
	switch componentIndex {
	case 1:
		return 1
	case 2:
		return 3
	case 3:
		return 5
	case 4:
		return 7
	case 5:
		return 10
	default:
		return 0
	}
}

func shipShieldCapacityPerHullClass(componentIndex int) int {
	switch componentIndex {
	case 1:
		return 5
	case 2:
		return 15
	case 3:
		return 25
	case 4:
		return 35
	case 5:
		return 50
	default:
		return 0
	}
}

func shipDriveTacticalSpeedBonus(componentIndex int) int {
	if componentIndex <= 1 {
		return 0
	}
	return (componentIndex - 1) * 2
}

func shipArmorDurabilityMultiple(componentIndex int) int {
	switch componentIndex {
	case 1:
		return 1
	case 2:
		return 2
	case 3:
		return 4
	case 4:
		return 6
	case 5:
		return 8
	case 6:
		return 10
	default:
		return 0
	}
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
