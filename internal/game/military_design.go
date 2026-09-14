package game

import (
	"fmt"
	"math"
	"strings"

	"moox/internal/core"
	"moox/internal/protocol"
	"moox/internal/ruleset"
)

const (
	CommandSaveMilitaryDesign = "empire.save_military_design"
	MilitaryShipProjectID     = "military_ship"
	SupportedMilitaryHullID   = "frigate"
)

type SaveMilitaryDesignPayload struct {
	DesignID           core.ID                `json:"design_id,omitempty"`
	Name               string                 `json:"name"`
	HullID             string                 `json:"hull_id"`
	StrategicPictureID int                    `json:"strategic_picture_id"`
	Weapons            []core.ShipWeaponMount `json:"weapons,omitempty"`
}

type MilitaryDesignSavedEvent struct {
	Design  core.ShipDesign `json:"design"`
	Created bool            `json:"created"`
}

func IsMilitaryDesignCommand(kind string) bool {
	return kind == CommandSaveMilitaryDesign
}

func NewSaveMilitaryDesignCommand(sequence uint32, payload SaveMilitaryDesignPayload) (protocol.Command, error) {
	if err := validateSaveMilitaryDesignPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandSaveMilitaryDesign, payload)
}

func decodeSaveMilitaryDesign(command protocol.Command) (SaveMilitaryDesignPayload, error) {
	var payload SaveMilitaryDesignPayload
	if err := decodeStrictCommandPayload(command, CommandSaveMilitaryDesign, &payload); err != nil {
		return SaveMilitaryDesignPayload{}, err
	}
	if err := validateSaveMilitaryDesignPayload(payload); err != nil {
		return SaveMilitaryDesignPayload{}, err
	}
	return payload, nil
}

func validateSaveMilitaryDesignPayload(payload SaveMilitaryDesignPayload) error {
	if strings.TrimSpace(payload.Name) == "" {
		return fmt.Errorf("name must not be empty")
	}
	if payload.HullID == "" {
		return fmt.Errorf("hull_id must not be empty")
	}
	if payload.StrategicPictureID < 0 {
		return fmt.Errorf("strategic_picture_id must be non-negative")
	}
	if err := validateBaselineMilitaryWeapons(payload.Weapons); err != nil {
		return err
	}
	return nil
}

func validateBaselineMilitaryWeapons(weapons []core.ShipWeaponMount) error {
	if len(weapons) > 8 {
		return fmt.Errorf("military design supports at most 8 weapon mounts")
	}
	previousSlot := -1
	for i, mount := range weapons {
		if mount.Slot < 0 || mount.Slot > 7 {
			return fmt.Errorf("weapon[%d] slot %d is outside 0..7", i, mount.Slot)
		}
		if i > 0 && mount.Slot <= previousSlot {
			return fmt.Errorf("weapon mounts must be strictly ascending by slot")
		}
		if mount.WeaponID != "laser_cannon" {
			return fmt.Errorf("weapon[%d] %q is not supported in the current slice", i, mount.WeaponID)
		}
		if mount.Count <= 0 {
			return fmt.Errorf("weapon[%d] count must be positive", i)
		}
		previousSlot = mount.Slot
	}
	return nil
}

func shipDesignByID(state *core.GameState, id core.ID) (int, *core.ShipDesign) {
	if state == nil || id == 0 {
		return -1, nil
	}
	for i := range state.ShipDesigns {
		if state.ShipDesigns[i].ID == id {
			return i, &state.ShipDesigns[i]
		}
	}
	return -1, nil
}

func (r *EconomyResolver) ResolveMilitaryDesignCommand(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) ([]DomainEvent, error) {
	if r == nil || r.Rules == nil {
		return nil, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	if !IsMilitaryDesignCommand(command.Kind) {
		return nil, fmt.Errorf("command kind %q is not a military design command", command.Kind)
	}
	event, err := r.saveMilitaryDesign(state, empireID, seatID, command)
	if err != nil {
		return nil, err
	}
	return []DomainEvent{event}, nil
}

func (r *EconomyResolver) saveMilitaryDesign(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeSaveMilitaryDesign(command)
	if err != nil {
		return DomainEvent{}, err
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}
	name := strings.TrimSpace(payload.Name)
	spec, err := r.Rules.clearedMilitaryDesignSpec(empire, payload.HullID, payload.StrategicPictureID, payload.Weapons)
	if err != nil {
		return DomainEvent{}, err
	}
	created := payload.DesignID == 0
	var design core.ShipDesign
	if created {
		design = core.ShipDesign{ID: state.NewID(), EmpireID: empireID, Revision: 1, Name: name, Spec: spec}
		state.ShipDesigns = append(state.ShipDesigns, design)
	} else {
		_, current := shipDesignByID(state, payload.DesignID)
		if current == nil {
			return DomainEvent{}, fmt.Errorf("references unknown ship design %d", payload.DesignID)
		}
		if current.EmpireID != empireID {
			return DomainEvent{}, fmt.Errorf("seat %d cannot update ship design %d owned by empire %d", seatID, current.ID, current.EmpireID)
		}
		for i := range state.Colonies {
			construction := state.Colonies[i].Construction
			if construction != nil && construction.ProjectKind == core.ConstructionProjectMilitaryShip && construction.ShipDesignID == current.ID {
				return DomainEvent{}, fmt.Errorf("ship design %d cannot be revised while colony %d constructs revision %d", current.ID, state.Colonies[i].ID, construction.ShipDesignRevision)
			}
		}
		if current.Revision == ^uint32(0) {
			return DomainEvent{}, fmt.Errorf("ship design %d revision overflow", current.ID)
		}
		current.Revision++
		current.Name = name
		current.Spec = spec
		design = *current
	}
	kind := "empire.military_design_updated"
	if created {
		kind = "empire.military_design_created"
	}
	return NewDomainEvent(kind, seatID, command.Sequence, MilitaryDesignSavedEvent{Design: design, Created: created})
}

func (r *EconomyRules) clearedMilitaryDesignSpec(empire *core.Empire, hullID string, strategicPictureID int, weapons []core.ShipWeaponMount) (core.ShipDesignSpec, error) {
	if empire == nil {
		return core.ShipDesignSpec{}, fmt.Errorf("empire must not be nil")
	}
	if err := validateBaselineMilitaryWeapons(weapons); err != nil {
		return core.ShipDesignSpec{}, err
	}
	if hullID != SupportedMilitaryHullID {
		return core.ShipDesignSpec{}, fmt.Errorf("military design hull %q is not supported in the current slice; only %q is available", hullID, SupportedMilitaryHullID)
	}
	hull, ok := r.ShipHulls[hullID]
	if !ok {
		return core.ShipDesignSpec{}, fmt.Errorf("ruleset has no ship hull %q", hullID)
	}
	pictureOK := false
	for _, pictureID := range hull.StrategicPictureIDs {
		if pictureID == strategicPictureID {
			pictureOK = true
			break
		}
	}
	if !pictureOK {
		return core.ShipDesignSpec{}, fmt.Errorf("strategic picture %d is not legal for hull %q", strategicPictureID, hullID)
	}
	drive, ok := bestKnownShipDrive(r.ShipDrives, empire)
	if !ok {
		return core.ShipDesignSpec{}, fmt.Errorf("empire %d has no supported Warp Drive technology for military design", empire.ID)
	}
	computer, ok := bestKnownShipComputer(r.ShipComputers, empire)
	if !ok {
		return core.ShipDesignSpec{}, fmt.Errorf("empire %d has no supported Computer technology for military design", empire.ID)
	}
	armor, ok := bestKnownShipArmor(r.ShipArmors, empire)
	if !ok {
		return core.ShipDesignSpec{}, fmt.Errorf("empire %d has no supported Armor technology for military design", empire.ID)
	}
	fuel, ok := bestKnownShipFuelCell(r.ShipFuelCells, empire)
	if !ok {
		return core.ShipDesignSpec{}, fmt.Errorf("empire %d has no supported Fuel Cell technology for military design", empire.ID)
	}
	shield, hasShield := bestKnownShipShield(r.ShipShields, empire)

	index := hull.SizeIndex
	if index < 0 || index >= len(drive.CostByHullPP) || index >= len(computer.CostByHullPP) || index >= len(drive.SpaceByHull) {
		return core.ShipDesignSpec{}, fmt.Errorf("ship hull %q size index %d has incomplete mandatory component tables", hull.ID, index)
	}
	baseCost := hull.BaseCostPP + drive.CostByHullPP[index] + computer.CostByHullPP[index] + (hull.BaseCostPP*armor.CostPercent)/100
	spaceUsed := drive.SpaceByHull[index]
	shieldID := ""
	if hasShield {
		if index >= len(shield.CostByHullPP) || index >= len(shield.SpaceByHull) {
			return core.ShipDesignSpec{}, fmt.Errorf("ship shield %q has incomplete hull table", shield.ID)
		}
		baseCost += shield.CostByHullPP[index]
		spaceUsed += shield.SpaceByHull[index]
		shieldID = shield.ID
	}
	weaponSnapshot := append([]core.ShipWeaponMount(nil), weapons...)
	if len(weaponSnapshot) > 0 {
		if r.TacticalCombat == nil {
			return core.ShipDesignSpec{}, fmt.Errorf("tactical combat rules are unavailable")
		}
		weapon := r.TacticalCombat.Weapon
		if weapon.ID != "laser_cannon" || weapon.TechnologyID != 100 {
			return core.ShipDesignSpec{}, fmt.Errorf("tactical rules do not define the supported Laser Cannon")
		}
		if !empireKnowsTechnology(empire, weapon.TechnologyID) {
			return core.ShipDesignSpec{}, fmt.Errorf("empire %d has no Laser Cannon technology %d", empire.ID, weapon.TechnologyID)
		}
		for i, mount := range weaponSnapshot {
			if mount.WeaponID != weapon.ID {
				return core.ShipDesignSpec{}, fmt.Errorf("weapon[%d] %q is not the supported Laser Cannon", i, mount.WeaponID)
			}
			spaceUsed += weapon.BaseSpace * mount.Count
			baseCost += weapon.BaseCostPP * mount.Count
		}
	}
	if spaceUsed > hull.BaseSpace {
		return core.ShipDesignSpec{}, fmt.Errorf("military design uses %d space but hull %q has only %d", spaceUsed, hull.ID, hull.BaseSpace)
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

func militaryShipProductionCostPP(empire *core.Empire, baseCostPP int, modifiers map[string]RaceEconomyModifiers) int {
	if baseCostPP <= 0 {
		return 0
	}
	if empire != nil {
		if modifier, ok := modifiers[empire.RaceID]; ok && modifier.GovernmentTraitID == "government_feudal" {
			return int(math.Ceil((2 * float64(baseCostPP)) / 3))
		}
	}
	return baseCostPP
}

func bestKnownShipDrive(items []ruleset.ShipDrive, empire *core.Empire) (ruleset.ShipDrive, bool) {
	for i := len(items) - 1; i >= 0; i-- {
		if empireKnowsTechnology(empire, items[i].TechnologyID) {
			return items[i], true
		}
	}
	return ruleset.ShipDrive{}, false
}

func bestKnownShipComputer(items []ruleset.ShipComputer, empire *core.Empire) (ruleset.ShipComputer, bool) {
	for i := len(items) - 1; i >= 0; i-- {
		if empireKnowsTechnology(empire, items[i].TechnologyID) {
			return items[i], true
		}
	}
	return ruleset.ShipComputer{}, false
}

func bestKnownShipArmor(items []ruleset.ShipArmor, empire *core.Empire) (ruleset.ShipArmor, bool) {
	for i := len(items) - 1; i >= 0; i-- {
		if empireKnowsTechnology(empire, items[i].TechnologyID) {
			return items[i], true
		}
	}
	return ruleset.ShipArmor{}, false
}

func bestKnownShipShield(items []ruleset.ShipShield, empire *core.Empire) (ruleset.ShipShield, bool) {
	for i := len(items) - 1; i >= 0; i-- {
		if empireKnowsTechnology(empire, items[i].TechnologyID) {
			return items[i], true
		}
	}
	return ruleset.ShipShield{}, false
}

func bestKnownShipFuelCell(items []ruleset.ShipFuelCell, empire *core.Empire) (ruleset.ShipFuelCell, bool) {
	for i := len(items) - 1; i >= 0; i-- {
		if empireKnowsTechnology(empire, items[i].TechnologyID) {
			return items[i], true
		}
	}
	return ruleset.ShipFuelCell{}, false
}

type MilitaryShipQueuedEvent struct {
	ColonyID           core.ID `json:"colony_id"`
	ShipDesignID       core.ID `json:"ship_design_id"`
	ShipDesignRevision uint32  `json:"ship_design_revision"`
	ShipDesignName     string  `json:"ship_design_name"`
	ProductionCostPP   int     `json:"production_cost_pp"`
}

type MilitaryShipCompletedEvent struct {
	ColonyID           core.ID `json:"colony_id"`
	EmpireID           core.ID `json:"empire_id"`
	ShipDesignID       core.ID `json:"ship_design_id"`
	ShipDesignRevision uint32  `json:"ship_design_revision"`
	ShipID             core.ID `json:"ship_id"`
	FleetID            core.ID `json:"fleet_id"`
	SystemID           core.ID `json:"system_id"`
	HullID             string  `json:"hull_id"`
	ProductionCostPP   int     `json:"production_cost_pp"`
}

func (r *EconomyResolver) queueMilitaryShip(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeQueueMilitaryShip(command)
	if err != nil {
		return DomainEvent{}, err
	}
	colony := colonyByID(state, payload.ColonyID)
	if colony == nil {
		return DomainEvent{}, fmt.Errorf("references unknown colony %d", payload.ColonyID)
	}
	if colony.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot queue military Ship on colony %d owned by empire %d", seatID, colony.ID, colony.EmpireID)
	}
	if colony.Construction != nil {
		return DomainEvent{}, fmt.Errorf("colony %d already constructs %s %q", colony.ID, colony.Construction.ProjectKind, colony.Construction.ProjectID)
	}
	_, design := shipDesignByID(state, payload.ShipDesignID)
	if design == nil {
		return DomainEvent{}, fmt.Errorf("references unknown ship design %d", payload.ShipDesignID)
	}
	if design.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot construct ship design %d owned by empire %d", seatID, design.ID, design.EmpireID)
	}
	colony.Construction = &core.ConstructionState{
		ProjectKind:        core.ConstructionProjectMilitaryShip,
		ProjectID:          MilitaryShipProjectID,
		ShipDesignID:       design.ID,
		ShipDesignRevision: design.Revision,
	}
	return NewDomainEvent("colony.military_ship_queued", seatID, command.Sequence, MilitaryShipQueuedEvent{
		ColonyID: colony.ID, ShipDesignID: design.ID, ShipDesignRevision: design.Revision, ShipDesignName: design.Name, ProductionCostPP: design.Spec.ProductionCostPP,
	})
}

func militaryConstructionDesign(state *core.GameState, empireID core.ID, designID core.ID, revision uint32) (*core.ShipDesign, error) {
	_, design := shipDesignByID(state, designID)
	if design == nil {
		return nil, fmt.Errorf("references unknown ship design %d", designID)
	}
	if design.EmpireID != empireID {
		return nil, fmt.Errorf("ship design %d is owned by empire %d, expected %d", design.ID, design.EmpireID, empireID)
	}
	if revision == 0 || design.Revision != revision {
		return nil, fmt.Errorf("ship design %d revision=%d, construction requires revision %d", design.ID, design.Revision, revision)
	}
	return design, nil
}

func completeMilitaryShip(state *core.GameState, colony *core.Colony, design *core.ShipDesign) (DomainEvent, error) {
	if state == nil || colony == nil || design == nil {
		return DomainEvent{}, fmt.Errorf("military Ship completion requires state, colony and design")
	}
	system := systemForPlanetID(state, colony.PlanetID)
	if system == nil {
		return DomainEvent{}, fmt.Errorf("colony %d planet %d is not assigned to a star system", colony.ID, colony.PlanetID)
	}
	shipSpec := design.Spec
	shipSpec.Weapons = append([]core.ShipWeaponMount(nil), design.Spec.Weapons...)
	var visual *core.ShipVisualGenome
	if design.VisualGenome != nil {
		clone := core.CloneShipVisualGenome(*design.VisualGenome)
		visual = &clone
	}
	ship := core.Ship{
		ID:                   state.NewID(),
		EmpireID:             colony.EmpireID,
		SourceDesignID:       design.ID,
		SourceDesignRevision: design.Revision,
		SourceVisualRevision: design.VisualRevision,
		Name:                 design.Name,
		Spec:                 shipSpec,
		VisualGenome:         visual,
	}
	fleet := core.StrategicFleet{
		ID:         state.NewID(),
		EmpireID:   colony.EmpireID,
		Role:       core.StrategicFleetRoleCombat,
		AtSystemID: system.ID,
		ShipIDs:    []core.ID{ship.ID},
	}
	state.Ships = append(state.Ships, ship)
	state.StrategicFleets = append(state.StrategicFleets, fleet)
	return NewDomainEvent("colony.military_ship_completed", 0, 0, MilitaryShipCompletedEvent{
		ColonyID: colony.ID, EmpireID: colony.EmpireID, ShipDesignID: design.ID, ShipDesignRevision: design.Revision,
		ShipID: ship.ID, FleetID: fleet.ID, SystemID: system.ID, HullID: ship.Spec.HullID, ProductionCostPP: ship.Spec.ProductionCostPP,
	})
}
