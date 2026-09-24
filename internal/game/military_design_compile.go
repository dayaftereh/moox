package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
	"moox/internal/ruleset"
)

const (
	MilitaryDesignReasonTechnologyRequired       = "technology_required"
	MilitaryDesignReasonMandatorySystemsRequired = "mandatory_systems_required"
	MilitaryDesignReasonInvalidComponent         = "invalid_component_selection"
	MilitaryDesignReasonInvalidStrategicPicture  = "invalid_strategic_picture"
	MilitaryDesignReasonSpaceExceeded            = "space_exceeded"
	MilitaryDesignReasonOrbitalDockRequired      = "orbital_dock_required"
	MilitaryDesignReasonRevisionInUse            = "design_revision_in_use"
)

type militaryDesignSelections struct {
	ComputerID string
	ArmorID    string
	ShieldID   *string
}

func militaryDesignError(reason, format string, args ...any) error {
	return fmt.Errorf("%s: %s", reason, fmt.Sprintf(format, args...))
}

func empireKnowsTechnologyField(empire *core.Empire, fieldID int) bool {
	if empire == nil || fieldID < 0 {
		return false
	}
	i := sort.SearchInts(empire.KnownTechnologyFieldIDs, fieldID)
	return i < len(empire.KnownTechnologyFieldIDs) && empire.KnownTechnologyFieldIDs[i] == fieldID
}

func (r *EconomyRules) shipComponentMiniaturizationLevel(empire *core.Empire, technologyID int) (int, error) {
	if r == nil || empire == nil {
		return 0, fmt.Errorf("miniaturization requires rules and empire")
	}
	fieldID, ok := r.TechnologyFieldByID[technologyID]
	if !ok {
		return 0, fmt.Errorf("technology %d has no research field", technologyID)
	}
	level := 0
	seenFields := map[int]struct{}{fieldID: {}}
	nextID := r.TechnologyFieldNextID[fieldID]
	for nextID != 0 {
		if _, repeated := seenFields[nextID]; repeated {
			break
		}
		seenFields[nextID] = struct{}{}
		if nextID >= 75 && nextID <= 82 {
			completed, _ := hyperAdvancedCompletedLevels(empire, nextID)
			level += completed
			break
		}
		if !empireKnowsTechnologyField(empire, nextID) {
			break
		}
		level++
		nextID = r.TechnologyFieldNextID[nextID]
	}
	return level, nil
}

func miniaturizationCostPercent(level int) int {
	factors := [...]int{100, 75, 55, 40, 30, 25}
	if level < 0 {
		level = 0
	}
	if level >= len(factors) {
		level = len(factors) - 1
	}
	return factors[level]
}

func miniaturizationSpacePerThousand(level, class int) int {
	class0 := [...]int{1000, 800, 650, 500, 350, 250}
	class1 := [...]int{1000, 800, 700, 600, 500, 400}
	factors := class0[:]
	if class == 1 {
		factors = class1[:]
	}
	if level < 0 {
		level = 0
	}
	if level >= len(factors) {
		level = len(factors) - 1
	}
	return factors[level]
}

func miniaturizedComponentCost(base, level int) int {
	if base <= 0 {
		return 0
	}
	value := (base * miniaturizationCostPercent(level)) / 100
	if value < 1 {
		return 1
	}
	return value
}

func miniaturizedComponentSpace(base, level, class int) int {
	if base <= 0 {
		return 0
	}
	value := (base*miniaturizationSpacePerThousand(level, class) + 500) / 1000
	if value < 1 {
		return 1
	}
	return value
}

func shipComputerByID(items []ruleset.ShipComputer, id string) (ruleset.ShipComputer, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return ruleset.ShipComputer{}, false
}

func shipArmorByID(items []ruleset.ShipArmor, id string) (ruleset.ShipArmor, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return ruleset.ShipArmor{}, false
}

func shipShieldByID(items []ruleset.ShipShield, id string) (ruleset.ShipShield, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return ruleset.ShipShield{}, false
}

func (r *EconomyRules) selectedShipComputer(empire *core.Empire, id string) (ruleset.ShipComputer, error) {
	if id == "" {
		item, ok := bestKnownShipComputer(r.ShipComputers, empire)
		if !ok {
			return ruleset.ShipComputer{}, militaryDesignError(MilitaryDesignReasonMandatorySystemsRequired, "empire %d has no supported Computer technology", empire.ID)
		}
		return item, nil
	}
	item, ok := shipComputerByID(r.ShipComputers, id)
	if !ok {
		return ruleset.ShipComputer{}, militaryDesignError(MilitaryDesignReasonInvalidComponent, "unknown Computer %q", id)
	}
	if !empireKnowsTechnology(empire, item.TechnologyID) {
		return ruleset.ShipComputer{}, militaryDesignError(MilitaryDesignReasonTechnologyRequired, "Computer %q requires technology %d", id, item.TechnologyID)
	}
	return item, nil
}

func (r *EconomyRules) selectedShipArmor(empire *core.Empire, id string) (ruleset.ShipArmor, error) {
	if id == "" {
		item, ok := bestKnownShipArmor(r.ShipArmors, empire)
		if !ok {
			return ruleset.ShipArmor{}, militaryDesignError(MilitaryDesignReasonMandatorySystemsRequired, "empire %d has no supported Armor technology", empire.ID)
		}
		return item, nil
	}
	item, ok := shipArmorByID(r.ShipArmors, id)
	if !ok {
		return ruleset.ShipArmor{}, militaryDesignError(MilitaryDesignReasonInvalidComponent, "unknown Armor %q", id)
	}
	if !empireKnowsTechnology(empire, item.TechnologyID) {
		return ruleset.ShipArmor{}, militaryDesignError(MilitaryDesignReasonTechnologyRequired, "Armor %q requires technology %d", id, item.TechnologyID)
	}
	return item, nil
}

func (r *EconomyRules) selectedShipShield(empire *core.Empire, id *string) (*ruleset.ShipShield, error) {
	if id == nil {
		item, ok := bestKnownShipShield(r.ShipShields, empire)
		if !ok {
			return nil, nil
		}
		return &item, nil
	}
	if *id == "" {
		return nil, nil
	}
	item, ok := shipShieldByID(r.ShipShields, *id)
	if !ok {
		return nil, militaryDesignError(MilitaryDesignReasonInvalidComponent, "unknown Shield %q", *id)
	}
	if !empireKnowsTechnology(empire, item.TechnologyID) {
		return nil, militaryDesignError(MilitaryDesignReasonTechnologyRequired, "Shield %q requires technology %d", *id, item.TechnologyID)
	}
	return &item, nil
}

func cloneShipDesignSpec(spec core.ShipDesignSpec) core.ShipDesignSpec {
	spec.Weapons = core.CloneShipWeaponMounts(spec.Weapons)
	return spec
}

func (r *EconomyRules) compileMilitaryDesignSpec(empire *core.Empire, hullID string, strategicPictureID int, selections militaryDesignSelections, weapons []core.ShipWeaponMount) (core.ShipDesignSpec, error) {
	if r == nil || empire == nil {
		return core.ShipDesignSpec{}, fmt.Errorf("military design requires rules and empire")
	}
	weaponSnapshot, err := canonicalWeaponMounts(weapons)
	if err != nil {
		return core.ShipDesignSpec{}, err
	}
	hull, ok := r.ShipHulls[hullID]
	if !ok {
		return core.ShipDesignSpec{}, militaryDesignError(MilitaryDesignReasonInvalidComponent, "ruleset has no ship hull %q", hullID)
	}
	requiredTechnologyID, _, err := r.shipDesignerHullTechnology(hullID)
	if err != nil {
		return core.ShipDesignSpec{}, err
	}
	if requiredTechnologyID != 0 && !empireKnowsTechnology(empire, requiredTechnologyID) {
		return core.ShipDesignSpec{}, militaryDesignError(MilitaryDesignReasonTechnologyRequired, "hull %q requires technology %d", hullID, requiredTechnologyID)
	}
	pictureOK := false
	for _, pictureID := range hull.StrategicPictureIDs {
		if pictureID == strategicPictureID {
			pictureOK = true
			break
		}
	}
	if !pictureOK {
		return core.ShipDesignSpec{}, militaryDesignError(MilitaryDesignReasonInvalidStrategicPicture, "strategic picture %d is not legal for hull %q", strategicPictureID, hullID)
	}

	drive, ok := bestKnownShipDrive(r.ShipDrives, empire)
	if !ok {
		return core.ShipDesignSpec{}, militaryDesignError(MilitaryDesignReasonMandatorySystemsRequired, "empire %d has no supported Warp Drive technology", empire.ID)
	}
	computer, err := r.selectedShipComputer(empire, selections.ComputerID)
	if err != nil {
		return core.ShipDesignSpec{}, err
	}
	armor, err := r.selectedShipArmor(empire, selections.ArmorID)
	if err != nil {
		return core.ShipDesignSpec{}, err
	}
	fuel, ok := bestKnownShipFuelCell(r.ShipFuelCells, empire)
	if !ok {
		return core.ShipDesignSpec{}, militaryDesignError(MilitaryDesignReasonMandatorySystemsRequired, "empire %d has no supported Fuel Cell technology", empire.ID)
	}
	shield, err := r.selectedShipShield(empire, selections.ShieldID)
	if err != nil {
		return core.ShipDesignSpec{}, err
	}

	index := hull.SizeIndex
	if index < 0 || index >= len(drive.CostByHullPP) || index >= len(computer.CostByHullPP) || index >= len(drive.SpaceByHull) {
		return core.ShipDesignSpec{}, militaryDesignError(MilitaryDesignReasonMandatorySystemsRequired, "ship hull %q size index %d has incomplete mandatory component tables", hull.ID, index)
	}

	computerLevel, err := r.shipComponentMiniaturizationLevel(empire, computer.TechnologyID)
	if err != nil {
		return core.ShipDesignSpec{}, err
	}
	baseCost := hull.BaseCostPP + drive.CostByHullPP[index] + miniaturizedComponentCost(computer.CostByHullPP[index], computerLevel) + (hull.BaseCostPP*armor.CostPercent)/100
	spaceUsed := drive.SpaceByHull[index]
	shieldID := ""
	if shield != nil {
		if index >= len(shield.CostByHullPP) || index >= len(shield.SpaceByHull) {
			return core.ShipDesignSpec{}, militaryDesignError(MilitaryDesignReasonMandatorySystemsRequired, "ship shield %q has incomplete hull table", shield.ID)
		}
		shieldLevel, err := r.shipComponentMiniaturizationLevel(empire, shield.TechnologyID)
		if err != nil {
			return core.ShipDesignSpec{}, err
		}
		baseCost += miniaturizedComponentCost(shield.CostByHullPP[index], shieldLevel)
		spaceUsed += miniaturizedComponentSpace(shield.SpaceByHull[index], shieldLevel, 0)
		shieldID = shield.ID
	}

	if len(weaponSnapshot) > 0 {
		if r.TacticalCombat == nil {
			return core.ShipDesignSpec{}, fmt.Errorf("tactical combat rules are unavailable")
		}
		for i, mount := range weaponSnapshot {
			quote, err := r.quoteWeapon(empire, mount)
			if err != nil {
				return core.ShipDesignSpec{}, err
			}
			weaponSnapshot[i] = quote.Mount
			spaceUsed += quote.RowSpace
			baseCost += quote.RowCostPP
		}
	}
	if spaceUsed > hull.BaseSpace {
		return core.ShipDesignSpec{}, militaryDesignError(MilitaryDesignReasonSpaceExceeded, "military design uses %d space but hull %q has only %d", spaceUsed, hull.ID, hull.BaseSpace)
	}
	productionCost := militaryShipProductionCostPP(empire, baseCost, r.RaceModifiers)
	return core.ShipDesignSpec{
		HullID:             hull.ID,
		StrategicPictureID: strategicPictureID,
		WarpDriveID:        drive.ID,
		FTLSpeed:           drive.FTLSpeed,
		ComputerID:         computer.ID,
		ArmorID:            armor.ID,
		ShieldID:           shieldID,
		FuelCellID:         fuel.ID,
		FuelRangeParsecs:   fuel.RangeParsecs,
		HullBaseCostPP:     hull.BaseCostPP,
		HullSpace:          hull.BaseSpace,
		SpaceUsed:          spaceUsed,
		BaseDesignCostPP:   baseCost,
		ProductionCostPP:   productionCost,
		Weapons:            weaponSnapshot,
	}, nil
}

func (r *EconomyRules) effectiveMilitaryDesignSpec(empire *core.Empire, stored core.ShipDesignSpec) (core.ShipDesignSpec, error) {
	shieldID := stored.ShieldID
	return r.compileMilitaryDesignSpec(empire, stored.HullID, stored.StrategicPictureID, militaryDesignSelections{
		ComputerID: stored.ComputerID,
		ArmorID:    stored.ArmorID,
		ShieldID:   &shieldID,
	}, stored.Weapons)
}

func (r *EconomyRules) ProjectMilitaryDesign(design core.ShipDesign, empire *core.Empire) (core.ShipDesign, error) {
	spec, err := r.effectiveMilitaryDesignSpec(empire, design.Spec)
	if err != nil {
		return core.ShipDesign{}, err
	}
	design.Spec = spec
	if design.VisualGenome != nil {
		genome := *design.VisualGenome
		design.VisualGenome = &genome
	}
	return design, nil
}

func (r *EconomyRules) colonyCanBuildMilitaryHull(colony *core.Colony, hullID string) (bool, error) {
	hull, ok := r.ShipHulls[hullID]
	if !ok {
		return false, militaryDesignError(MilitaryDesignReasonInvalidComponent, "ruleset has no ship hull %q", hullID)
	}
	if hull.SizeIndex < 2 {
		return true, nil
	}
	_, tier, err := colonyCommandStation(colony)
	if err != nil {
		return false, err
	}
	return tier >= 1, nil
}
