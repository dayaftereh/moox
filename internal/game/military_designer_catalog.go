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
	BaseCombatSpeed       int    `json:"base_combat_speed"`
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
	CostReductionPercent       int    `json:"cost_reduction_percent,omitempty"`
	SpaceReductionPercent      int    `json:"space_reduction_percent,omitempty"`
	EffectiveCostByHullPP      []int  `json:"effective_cost_by_hull_pp,omitempty"`
	EffectiveSpaceByHull       []int  `json:"effective_space_by_hull,omitempty"`
	BeamAttackBonus            int    `json:"beam_attack_bonus,omitempty"`
	ShieldDamageBlockPerHit    int    `json:"shield_damage_block_per_hit,omitempty"`
	ShieldCapacityPerHullClass int    `json:"shield_capacity_per_hull_class,omitempty"`
}

type MilitaryDesignerSpecialChoice struct {
	ID                    string   `json:"id"`
	NameKey               string   `json:"name_key"`
	OriginalID            int      `json:"original_id"`
	TechnologyID          int      `json:"technology_id"`
	TechnologyFieldID     int      `json:"technology_field_id"`
	TechnologyKnown       bool     `json:"technology_known"`
	Available             bool     `json:"available"`
	LockReason            string   `json:"lock_reason,omitempty"`
	SupportClass          int      `json:"support_class"`
	TacticalSupport       string   `json:"tactical_support"`
	MiniaturizationLevel  int      `json:"miniaturization_level"`
	CostReductionPercent  int      `json:"cost_reduction_percent"`
	SpaceReductionPercent int      `json:"space_reduction_percent"`
	EffectiveCostByHullPP []int    `json:"effective_cost_by_hull_pp"`
	EffectiveSpaceByHull  []int    `json:"effective_space_by_hull"`
	ExcludedSpecialIDs    []string `json:"excluded_special_ids,omitempty"`
	RequiresShield        bool     `json:"requires_shield,omitempty"`
}

type MilitaryDesignerAutomaticSystems struct {
	WarpDriveID             string `json:"warp_drive_id,omitempty"`
	FTLSpeed                int    `json:"ftl_speed,omitempty"`
	TacticalSpeedBonus      int    `json:"tactical_speed_bonus,omitempty"`
	ArmorID                 string `json:"armor_id,omitempty"`
	ArmorDurabilityMultiple int    `json:"armor_durability_multiple,omitempty"`
	FuelCellID              string `json:"fuel_cell_id,omitempty"`
	FuelRangeParsecs        int    `json:"fuel_range_parsecs,omitempty"`
	MegafluxersKnown        bool   `json:"megafluxers_known"`
	HullSpacePercent        int    `json:"hull_space_percent"`
}

type MilitaryDesignerWeaponMountChoice struct {
	ID                string `json:"id"`
	Available         bool   `json:"available"`
	LockReason        string `json:"lock_reason,omitempty"`
	SpacePercentDelta int    `json:"space_percent_delta"`
	CostPercentDelta  int    `json:"cost_percent_delta"`
}

type MilitaryDesignerWeaponModifierChoice struct {
	ID                string `json:"id"`
	RequiredMiniLevel int    `json:"required_mini_level"`
	Available         bool   `json:"available"`
	LockReason        string `json:"lock_reason,omitempty"`
	SpacePercentDelta int    `json:"space_percent_delta"`
	CostPercentDelta  int    `json:"cost_percent_delta"`
}

type MilitaryDesignerWeaponFiringArcChoice struct {
	ID                    core.ShipWeaponFiringArcID `json:"id"`
	LabelEN               string                     `json:"label_en"`
	LabelDE               string                     `json:"label_de"`
	SpaceCostPercent      int                        `json:"space_cost_percent"`
	ProductionCostPercent int                        `json:"production_cost_percent"`
}

type MilitaryDesignerWeaponChoice struct {
	ID                      string                                  `json:"id"`
	NameKey                 string                                  `json:"name_key"`
	Kind                    string                                  `json:"kind"`
	OriginalIndex           int                                     `json:"original_index"`
	TechnologyID            int                                     `json:"technology_id"`
	TechnologyFieldID       int                                     `json:"technology_field_id"`
	TechnologyKnown         bool                                    `json:"technology_known"`
	Available               bool                                    `json:"available"`
	LockReason              string                                  `json:"lock_reason,omitempty"`
	BaseSpace               int                                     `json:"base_space"`
	BaseCostPP              int                                     `json:"base_cost_pp"`
	MinDamage               int                                     `json:"min_damage"`
	MaxDamage               int                                     `json:"max_damage"`
	MiniaturizationLevel    int                                     `json:"miniaturization_level"`
	CostReductionPercent    int                                     `json:"cost_reduction_percent"`
	SpaceReductionPercent   int                                     `json:"space_reduction_percent"`
	EffectiveStandardSpace  int                                     `json:"effective_standard_space"`
	EffectiveStandardCostPP int                                     `json:"effective_standard_cost_pp"`
	AllowedModifierMask     int                                     `json:"allowed_modifier_mask"`
	MountChoices            []MilitaryDesignerWeaponMountChoice     `json:"mount_choices"`
	ModifierChoices         []MilitaryDesignerWeaponModifierChoice  `json:"modifier_choices"`
	AmmoChoices             []int                                   `json:"ammo_choices,omitempty"`
	DefaultAmmo             int                                     `json:"default_ammo,omitempty"`
	FiringArcChoices        []MilitaryDesignerWeaponFiringArcChoice `json:"firing_arc_choices,omitempty"`
	DefaultFiringArcID      core.ShipWeaponFiringArcID              `json:"default_firing_arc_id,omitempty"`
	FixedFiringArcID        core.ShipWeaponFiringArcID              `json:"fixed_firing_arc_id,omitempty"`
	InnateEffectIDs         []string                                `json:"innate_effect_ids,omitempty"`
	TacticalSupport         string                                  `json:"tactical_support"`
	CraftPerUnit            int                                     `json:"craft_per_unit,omitempty"`
}

func beamFiringArcChoices() []MilitaryDesignerWeaponFiringArcChoice {
	return []MilitaryDesignerWeaponFiringArcChoice{
		{ID: core.ShipWeaponFiringArcForward, LabelEN: "Forward", LabelDE: "Vorne"},
		{ID: core.ShipWeaponFiringArcForwardExtended, LabelEN: "Forward Extended", LabelDE: "Vorne erweitert", SpaceCostPercent: 25, ProductionCostPercent: 25},
		{ID: core.ShipWeaponFiringArcBackExtended, LabelEN: "Back Extended", LabelDE: "Hinten erweitert", SpaceCostPercent: 25, ProductionCostPercent: 25},
		{ID: core.ShipWeaponFiringArcRear, LabelEN: "Rear", LabelDE: "Hinten"},
		{ID: core.ShipWeaponFiringArc360, LabelEN: "360", LabelDE: "360", SpaceCostPercent: 50, ProductionCostPercent: 50},
	}
}

type MilitaryDesignerVariant struct {
	Key              string                 `json:"key"`
	HullID           string                 `json:"hull_id"`
	Weapons          []core.ShipWeaponMount `json:"weapons,omitempty"`
	Spec             core.ShipDesignSpec    `json:"spec"`
	CommandPointCost int                    `json:"command_point_cost"`
}

type MilitaryDesignerCatalog struct {
	Hulls                          []MilitaryDesignerHullChoice      `json:"hulls"`
	Computers                      []MilitaryDesignerComponentChoice `json:"computers"`
	Shields                        []MilitaryDesignerComponentChoice `json:"shields"`
	AutomaticSystems               MilitaryDesignerAutomaticSystems  `json:"automatic_systems"`
	Weapons                        []MilitaryDesignerWeaponChoice    `json:"weapons"`
	Specials                       []MilitaryDesignerSpecialChoice   `json:"specials"`
	Variants                       []MilitaryDesignerVariant         `json:"variants"`
	ProductionCostNumerator        int                               `json:"production_cost_numerator"`
	ProductionCostDenominator      int                               `json:"production_cost_denominator"`
	MaxUnusedSpaceCombatSpeedBonus int                               `json:"max_unused_space_combat_speed_bonus"`
	BeamDefensePerCombatSpeed      int                               `json:"beam_defense_per_combat_speed"`
	BaseMissileEvasion             int                               `json:"base_missile_evasion"`
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
			BaseCombatSpeed:      shipHullBaseCombatSpeed(hull.SizeIndex),
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
		Hulls: hulls, ProductionCostNumerator: 1, ProductionCostDenominator: 1, MaxUnusedSpaceCombatSpeedBonus: 10, BeamDefensePerCombatSpeed: 5, BaseMissileEvasion: 0,
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
	catalog.AutomaticSystems.HullSpacePercent = 100
	if r.ShipSpecials != nil {
		catalog.AutomaticSystems.MegafluxersKnown = empireKnowsTechnology(empire, r.ShipSpecials.Megafluxers.TechnologyID)
		if catalog.AutomaticSystems.MegafluxersKnown {
			catalog.AutomaticSystems.HullSpacePercent = r.ShipSpecials.Megafluxers.HullSpacePercent
		}
		for _, special := range r.ShipSpecials.Specials {
			known := empireKnowsTechnology(empire, special.TechnologyID)
			level := 0
			if known && special.Miniaturizes {
				var err error
				level, err = r.shipComponentMiniaturizationLevel(empire, special.TechnologyID)
				if err != nil {
					return MilitaryDesignerCatalog{}, err
				}
			}
			effectiveSpace := append([]int(nil), special.BaseSpaceByHull...)
			effectiveCost := append([]int(nil), special.BaseCostByHullPP...)
			if special.Miniaturizes {
				for i := range effectiveSpace {
					effectiveSpace[i] = miniaturizedComponentSpace(special.BaseSpaceByHull[i], level, special.MiniaturizationClass)
					effectiveCost[i] = miniaturizedComponentCost(special.BaseCostByHullPP[i], level)
				}
			}
			excludedIDs := make([]string, 0, len(special.ExcludedOriginalIDs))
			for _, originalID := range special.ExcludedOriginalIDs {
				excluded, ok := r.ShipSpecials.ByOriginalID(originalID)
				if ok && excluded.ID != special.ID {
					excludedIDs = append(excludedIDs, excluded.ID)
				}
			}
			lockReason := ""
			if !known {
				lockReason = MilitaryDesignReasonSpecialTechnologyRequired
			}
			costReduction := 0
			spaceReduction := 0
			if special.Miniaturizes {
				costReduction = 100 - miniaturizationCostPercent(level)
				spaceReduction = 100 - miniaturizationSpacePerThousand(level, special.MiniaturizationClass)/10
			}
			catalog.Specials = append(catalog.Specials, MilitaryDesignerSpecialChoice{
				ID: special.ID, NameKey: r.TechnologyNameKeyByID[special.TechnologyID],
				OriginalID: special.OriginalID, TechnologyID: special.TechnologyID, TechnologyFieldID: r.TechnologyFieldByID[special.TechnologyID],
				TechnologyKnown: known, Available: known, LockReason: lockReason,
				SupportClass: special.SupportClass, TacticalSupport: special.TacticalSupport,
				MiniaturizationLevel: level, CostReductionPercent: costReduction, SpaceReductionPercent: spaceReduction,
				EffectiveCostByHullPP: effectiveCost, EffectiveSpaceByHull: effectiveSpace,
				ExcludedSpecialIDs: excludedIDs, RequiresShield: special.RequiresShield,
			})
		}
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
			LockReason: lockReason, MiniaturizationLevel: level,
			CostReductionPercent:  100 - miniaturizationCostPercent(level),
			EffectiveCostByHullPP: effective,
			BeamAttackBonus:       shipComputerBeamAttackBonus(computer.ComponentIndex),
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
			CostReductionPercent:  100 - miniaturizationCostPercent(level),
			SpaceReductionPercent: 100 - miniaturizationSpacePerThousand(level, 0)/10,
			EffectiveCostByHullPP: effectiveCost, EffectiveSpaceByHull: effectiveSpace,
			ShieldDamageBlockPerHit:    shipShieldDamageBlock(shield.ComponentIndex),
			ShieldCapacityPerHullClass: shipShieldCapacityPerHullClass(shield.ComponentIndex),
		})
	}

	if r.TacticalCombat != nil {
		for _, weapon := range r.TacticalCombat.AllWeapons() {
			technologyKnown := empireKnowsTechnology(empire, weapon.TechnologyID)
			level := 0
			if technologyKnown {
				var err error
				level, err = r.shipComponentMiniaturizationLevel(empire, weapon.TechnologyID)
				if err != nil {
					return MilitaryDesignerCatalog{}, err
				}
			}
			lockReason := ""
			if !technologyKnown {
				lockReason = ShipDesignerLockTechnologyRequired
			}

			mountChoices := []MilitaryDesignerWeaponMountChoice{{
				ID: string(core.ShipWeaponMountStandard), Available: technologyKnown,
				LockReason: lockReason,
			}}
			if weapon.Kind == "beam" {
				for _, modifier := range r.TacticalCombat.BeamModifiers {
					if modifier.Kind != "mount" || weapon.AllowedModifierMask&modifier.Bit == 0 {
						continue
					}
					mountID := ""
					switch modifier.ID {
					case "heavy":
						mountID = string(core.ShipWeaponMountHeavy)
					case "point_defense":
						mountID = string(core.ShipWeaponMountPointDefense)
					default:
						continue
					}
					mountChoices = append(mountChoices, MilitaryDesignerWeaponMountChoice{
						ID: mountID, Available: technologyKnown, LockReason: lockReason,
						SpacePercentDelta: modifier.SpacePercentDelta, CostPercentDelta: modifier.CostPercentDelta,
					})
				}
			}

			modifierSource := r.TacticalCombat.BeamModifiers
			if weapon.Kind == "fighter" {
				modifierSource = nil
			} else if weapon.Kind != "beam" {
				modifierSource = r.TacticalCombat.WeaponModifiers()
			}
			modifierChoices := make([]MilitaryDesignerWeaponModifierChoice, 0, 6)
			for _, modifier := range modifierSource {
				if modifier.Kind != "special" || weapon.AllowedModifierMask&modifier.Bit == 0 {
					continue
				}
				available := technologyKnown && level >= modifier.RequiredMiniLevel
				modifierLock := ""
				if !technologyKnown {
					modifierLock = ShipDesignerLockTechnologyRequired
				} else if level < modifier.RequiredMiniLevel {
					modifierLock = MilitaryDesignReasonWeaponModifierRefinementRequired
				}
				modifierChoices = append(modifierChoices, MilitaryDesignerWeaponModifierChoice{
					ID: modifier.ID, RequiredMiniLevel: modifier.RequiredMiniLevel, Available: available, LockReason: modifierLock,
					SpacePercentDelta: modifier.SpacePercentDelta, CostPercentDelta: modifier.CostPercentDelta,
				})
			}

			var ammoChoices []int
			defaultAmmo := 0
			fixedArc := core.ShipWeaponFiringArcID("")
			var firingArcChoices []MilitaryDesignerWeaponFiringArcChoice
			defaultFiringArc := core.ShipWeaponFiringArcID("")
			defaultMount := core.ShipWeaponMount{
				Slot: 0, WeaponID: weapon.ID, Count: 1, MountType: core.ShipWeaponMountStandard,
			}
			if weapon.Kind == "beam" {
				firingArcChoices = beamFiringArcChoices()
				defaultFiringArc = core.ShipWeaponFiringArcForward
				defaultMount.FiringArcID = defaultFiringArc
			} else if weapon.Kind != "fighter" {
				fixedArc = core.ShipWeaponFiringArc360
				defaultMount.FiringArcID = fixedArc
				switch weapon.Kind {
				case "missile":
					ammoChoices = []int{2, 5, 10, 15, 20}
					defaultAmmo = 5
				case "torpedo":
					defaultAmmo = 2
				case "bomb":
					defaultAmmo = 10
				}
				defaultMount.Ammo = defaultAmmo
			}

			effectiveSpace := miniaturizedComponentSpace(weapon.BaseSpace, level, 1)
			effectiveCost := miniaturizedComponentCost(weapon.BaseCostPP, level)
			if technologyKnown {
				quote, err := r.quoteWeapon(empire, defaultMount)
				if err != nil {
					return MilitaryDesignerCatalog{}, err
				}
				effectiveSpace = quote.UnitSpace
				effectiveCost = quote.UnitCostPP
			} else if weapon.Kind == "missile" {
				rackBase, _ := ordnanceRackBase(defaultAmmo)
				effectiveSpace = miniaturizedComponentSpace(weapon.BaseSpace+rackBase, level, 1)
				effectiveCost = miniaturizedComponentCost(weapon.BaseCostPP+rackBase, level)
			}

			catalog.Weapons = append(catalog.Weapons, MilitaryDesignerWeaponChoice{
				ID: weapon.ID, NameKey: r.TechnologyNameKeyByID[weapon.TechnologyID], Kind: weapon.Kind,
				OriginalIndex: weapon.OriginalIndex, TechnologyID: weapon.TechnologyID, TechnologyFieldID: weapon.TechnologyFieldID,
				TechnologyKnown: technologyKnown, Available: technologyKnown, LockReason: lockReason,
				BaseSpace: weapon.BaseSpace, BaseCostPP: weapon.BaseCostPP, MinDamage: weapon.MinDamage, MaxDamage: weapon.MaxDamage,
				MiniaturizationLevel:    level,
				CostReductionPercent:    100 - miniaturizationCostPercent(level),
				SpaceReductionPercent:   100 - miniaturizationSpacePerThousand(level, 1)/10,
				EffectiveStandardSpace:  effectiveSpace,
				EffectiveStandardCostPP: effectiveCost,
				AllowedModifierMask:     weapon.AllowedModifierMask,
				MountChoices:            mountChoices,
				ModifierChoices:         modifierChoices,
				AmmoChoices:             ammoChoices,
				DefaultAmmo:             defaultAmmo,
				FiringArcChoices:        firingArcChoices,
				DefaultFiringArcID:      defaultFiringArc,
				FixedFiringArcID:        fixedArc,
				InnateEffectIDs:         append([]string(nil), weapon.InnateEffectIDs...), TacticalSupport: weapon.TacticalSupport, CraftPerUnit: weapon.CraftPerUnit,
			})
		}
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
		laserIndex := -1
		for weaponIndex := range catalog.Weapons {
			if catalog.Weapons[weaponIndex].ID == "laser_cannon" && catalog.Weapons[weaponIndex].Available {
				laserIndex = weaponIndex
				break
			}
		}
		if laserIndex >= 0 {
			weapon := catalog.Weapons[laserIndex]
			mounts := []core.ShipWeaponMount{{Slot: 0, WeaponID: weapon.ID, Count: 1, MountType: core.ShipWeaponMountStandard}}
			spec, err := r.clearedMilitaryDesignSpec(empire, hull.ID, pictureID, mounts)
			if err == nil {
				catalog.Variants = append(catalog.Variants, MilitaryDesignerVariant{
					Key: hull.ID + ":" + weapon.ID, HullID: hull.ID,
					Weapons: core.CloneShipWeaponMounts(spec.Weapons), Spec: spec,
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

func shipHullBaseCombatSpeed(sizeIndex int) int {
	switch sizeIndex {
	case 0:
		return 10
	case 1:
		return 8
	case 2:
		return 6
	case 3:
		return 5
	case 4:
		return 4
	case 5:
		return 3
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
