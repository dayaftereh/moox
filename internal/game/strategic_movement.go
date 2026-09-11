package game

import (
	"fmt"
	"math"

	"moox/internal/core"
	"moox/internal/protocol"
)

const (
	standardFuelCellsTechnologyID  = 167
	deuteriumFuelCellsTechnologyID = 51
	iridiumFuelCellsTechnologyID   = 98
	urridiumFuelCellsTechnologyID  = 194
	thoriumFuelCellsTechnologyID   = 184
)

type FleetMovementStartedEvent struct {
	FleetID               core.ID `json:"fleet_id"`
	EmpireID              core.ID `json:"empire_id"`
	SourceSystemID        core.ID `json:"source_system_id"`
	DestinationSystemID   core.ID `json:"destination_system_id"`
	RemainingTurns        int     `json:"remaining_turns"`
	FTLSpeed              int     `json:"ftl_speed"`
	FuelRangeParsecs      int     `json:"fuel_range_parsecs"`
	SupplyDistanceParsecs int     `json:"supply_distance_parsecs"`
}

type FleetMovementProgressedEvent struct {
	FleetID             core.ID `json:"fleet_id"`
	EmpireID            core.ID `json:"empire_id"`
	DestinationSystemID core.ID `json:"destination_system_id"`
	RemainingTurns      int     `json:"remaining_turns"`
}

type FleetArrivedEvent struct {
	FleetID  core.ID `json:"fleet_id"`
	EmpireID core.ID `json:"empire_id"`
	SystemID core.ID `json:"system_id"`
}

type PlanetColonizedEvent struct {
	EmpireID core.ID `json:"empire_id"`
	FleetID  core.ID `json:"fleet_id"`
	SystemID core.ID `json:"system_id"`
	PlanetID core.ID `json:"planet_id"`
	ColonyID core.ID `json:"colony_id"`
}

type ColonyShipConsumedEvent struct {
	EmpireID core.ID `json:"empire_id"`
	FleetID  core.ID `json:"fleet_id"`
	ColonyID core.ID `json:"colony_id"`
	PlanetID core.ID `json:"planet_id"`
}

func strategicDistanceParsecs(source, destination core.StarSystem) int {
	dx := float64(source.X - destination.X)
	dy := float64(source.Y - destination.Y)
	distance := math.Hypot(dx, dy)
	if distance <= 0 {
		return 0
	}
	return int(math.Ceil(distance / populationTransferCoordinateUnitsParsec))
}

func strategicTravelETA(source, destination core.StarSystem, ftlSpeed int) int {
	parsecs := strategicDistanceParsecs(source, destination)
	if parsecs <= 0 || ftlSpeed < 2 {
		return 0
	}
	return (parsecs + ftlSpeed - 1) / ftlSpeed
}

type strategicFuelTechnology struct {
	technologyID int
	fuelCellID   string
	rangeParsecs int
}

var strategicFuelTechnologies = []strategicFuelTechnology{
	{standardFuelCellsTechnologyID, "standard_fuel_cells", 4},
	{deuteriumFuelCellsTechnologyID, "deuterium_fuel_cells", 6},
	{iridiumFuelCellsTechnologyID, "iridium_fuel_cells", 9},
	{urridiumFuelCellsTechnologyID, "urridium_fuel_cells", 12},
	{thoriumFuelCellsTechnologyID, "thorium_fuel_cells", 255},
}

func strategicFuelCellProfile(empire core.Empire) (string, int) {
	known := make(map[int]struct{}, len(empire.KnownTechnologyIDs))
	for _, technologyID := range empire.KnownTechnologyIDs {
		known[technologyID] = struct{}{}
	}
	fuelCellID := ""
	rangeParsecs := 0
	for _, fuel := range strategicFuelTechnologies {
		if _, ok := known[fuel.technologyID]; ok && fuel.rangeParsecs > rangeParsecs {
			fuelCellID = fuel.fuelCellID
			rangeParsecs = fuel.rangeParsecs
		}
	}
	return fuelCellID, rangeParsecs
}

func colonyShipFuelRangeParsecs(empire core.Empire) int {
	_, rangeParsecs := strategicFuelCellProfile(empire)
	return rangeParsecs
}

// ProjectSpecialFleetMovementMetadata adds the movement equipment currently effective
// for fixed Colony/Outpost/Transport fleets to a player-view copy. The fleet FTL speed
// remains the per-instance strategic speed; fuel cell/range intentionally follow current
// empire technology because that is the authoritative movement rule.
func (r *EconomyRules) ProjectSpecialFleetMovementMetadata(fleet core.StrategicFleet, empire core.Empire) core.StrategicFleet {
	if _, fixed := fixedSpecialShipName(fleet.SpecialKind); !fixed {
		return fleet
	}
	fleet.WarpDriveID = r.populationTransferDriveIDForFTLSpeed(empire, fleet.FTLSpeed)
	if fleet.WarpDriveID == "" {
		fleet.WarpDriveID, _ = r.populationTransferDriveProfile(empire)
	}
	fleet.FuelCellID, fleet.FuelRangeParsecs = strategicFuelCellProfile(empire)
	return fleet
}

func systemByID(state *core.GameState, id core.ID) *core.StarSystem {
	if state == nil || id == 0 {
		return nil
	}
	for i := range state.Galaxy.Systems {
		if state.Galaxy.Systems[i].ID == id {
			return &state.Galaxy.Systems[i]
		}
	}
	return nil
}

func strategicFleetByID(state *core.GameState, id core.ID) (int, *core.StrategicFleet) {
	if state == nil || id == 0 {
		return -1, nil
	}
	for i := range state.StrategicFleets {
		if state.StrategicFleets[i].ID == id {
			return i, &state.StrategicFleets[i]
		}
	}
	return -1, nil
}

func nearestEmpireSupplyDistanceParsecs(state *core.GameState, empireID core.ID, destination core.StarSystem) (int, bool) {
	nearest := 0
	found := false
	considerPlanet := func(planetID core.ID) {
		supply := systemForPlanetID(state, planetID)
		if supply == nil {
			return
		}
		distance := strategicDistanceParsecs(*supply, destination)
		if !found || distance < nearest {
			nearest = distance
			found = true
		}
	}
	for _, colony := range state.Colonies {
		if colony.EmpireID == empireID {
			considerPlanet(colony.PlanetID)
		}
	}
	for _, outpost := range state.Outposts {
		if outpost.EmpireID != empireID {
			continue
		}
		target := orbitalBodyTargetByID(state, outpostTargetBodyID(outpost))
		if target.System == nil {
			continue
		}
		distance := strategicDistanceParsecs(*target.System, destination)
		if !found || distance < nearest {
			nearest = distance
			found = true
		}
	}
	return nearest, found
}

func fixedSpecialShipName(kind core.StrategicFleetSpecialKind) (string, bool) {
	switch kind {
	case core.StrategicFleetSpecialColonyShip:
		return "Colony Ship", true
	case core.StrategicFleetSpecialOutpostShip:
		return "Outpost Ship", true
	case core.StrategicFleetSpecialTroopTransport:
		return "Troop Transport", true
	default:
		return "", false
	}
}

func (r *EconomyResolver) moveFleetEvents(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) ([]DomainEvent, error) {
	payload, err := decodeMoveFleet(command)
	if err != nil {
		return nil, err
	}
	_, fleet := strategicFleetByID(state, payload.FleetID)
	if fleet == nil {
		return nil, fmt.Errorf("references unknown strategic fleet %d", payload.FleetID)
	}
	if fleet.EmpireID != empireID {
		return nil, fmt.Errorf("seat %d cannot move fleet %d owned by empire %d", seatID, fleet.ID, fleet.EmpireID)
	}
	if _, fixedSpecial := fixedSpecialShipName(fleet.SpecialKind); !fixedSpecial {
		if fleet.Role == core.StrategicFleetRoleCombat && fleet.SpecialKind == core.StrategicFleetSpecialNone {
			return r.moveCombatFleet(state, empireID, seatID, command, payload, fleet)
		}
		return nil, fmt.Errorf("fleet %d is not a movable fixed support ship or ordinary combat fleet", fleet.ID)
	}
	if len(payload.ShipIDs) != 0 {
		return nil, fmt.Errorf("fixed special fleet %d does not support ship_ids selection", fleet.ID)
	}
	shipName, _ := fixedSpecialShipName(fleet.SpecialKind)
	if fleet.AtSystemID == 0 || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 {
		return nil, fmt.Errorf("%s fleet %d is not stationary at a star system", shipName, fleet.ID)
	}
	if fleet.FTLSpeed < 2 {
		return nil, fmt.Errorf("%s fleet %d has invalid ftl_speed %d", shipName, fleet.ID, fleet.FTLSpeed)
	}
	source := systemByID(state, fleet.AtSystemID)
	if source == nil {
		return nil, fmt.Errorf("fleet %d references unknown source system %d", fleet.ID, fleet.AtSystemID)
	}
	destination := systemByID(state, payload.DestinationSystemID)
	if destination == nil {
		return nil, fmt.Errorf("references unknown destination star system %d", payload.DestinationSystemID)
	}
	if destination.ID == source.ID {
		return nil, fmt.Errorf("fleet %d is already at star system %d", fleet.ID, source.ID)
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return nil, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}
	fuelRange := colonyShipFuelRangeParsecs(*empire)
	supplyDistance, hasSupply := nearestEmpireSupplyDistanceParsecs(state, empireID, *destination)
	if !hasSupply || supplyDistance > fuelRange {
		return nil, fmt.Errorf("destination system %d is outside %s fuel range: nearest supply %d pc, range %d pc", destination.ID, shipName, supplyDistance, fuelRange)
	}
	eta := strategicTravelETA(*source, *destination, fleet.FTLSpeed)
	if eta < 1 {
		return nil, fmt.Errorf("invalid strategic ETA %d from system %d to %d", eta, source.ID, destination.ID)
	}
	event, err := NewDomainEvent("empire.fleet_movement_started", seatID, command.Sequence, FleetMovementStartedEvent{
		FleetID: fleet.ID, EmpireID: fleet.EmpireID, SourceSystemID: source.ID, DestinationSystemID: destination.ID,
		RemainingTurns: eta, FTLSpeed: fleet.FTLSpeed, FuelRangeParsecs: fuelRange, SupplyDistanceParsecs: supplyDistance,
	})
	if err != nil {
		return nil, err
	}
	fleet.AtSystemID = 0
	fleet.DestinationSystemID = destination.ID
	fleet.RemainingTurns = eta
	return []DomainEvent{event}, nil
}

// moveFleet keeps the pre-Slice-04 helper shape for focused Colony/Outpost tests.
// The resolver uses moveFleetEvents so subset combat movement can emit split then move.
func (r *EconomyResolver) moveFleet(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	events, err := r.moveFleetEvents(state, empireID, seatID, command)
	if err != nil {
		return DomainEvent{}, err
	}
	if len(events) == 0 {
		return DomainEvent{}, fmt.Errorf("fleet movement produced no events")
	}
	return events[len(events)-1], nil
}

func (r *EconomyResolver) advanceStrategicFleetTransit(state *core.GameState) ([]DomainEvent, error) {
	var events []DomainEvent
	for i := range state.StrategicFleets {
		fleet := &state.StrategicFleets[i]
		if fleet.DestinationSystemID == 0 {
			continue
		}
		if fleet.AtSystemID != 0 || fleet.RemainingTurns <= 0 {
			return nil, fmt.Errorf("strategic fleet %d has invalid transit state", fleet.ID)
		}
		_, fixedSpecial := fixedSpecialShipName(fleet.SpecialKind)
		combatTransit := fleet.Role == core.StrategicFleetRoleCombat && fleet.SpecialKind == core.StrategicFleetSpecialNone && fleet.FTLSpeed == 0 && len(fleet.ShipIDs) > 0
		if (!fixedSpecial || fleet.FTLSpeed < 2) && !combatTransit {
			return nil, fmt.Errorf("strategic fleet %d has invalid transit kind/drive state", fleet.ID)
		}
		destination := systemByID(state, fleet.DestinationSystemID)
		if destination == nil {
			return nil, fmt.Errorf("strategic fleet %d references unknown destination system %d", fleet.ID, fleet.DestinationSystemID)
		}
		nextRemaining := fleet.RemainingTurns - 1
		if nextRemaining > 0 {
			event, err := NewDomainEvent("empire.fleet_movement_progressed", 0, 0, FleetMovementProgressedEvent{
				FleetID: fleet.ID, EmpireID: fleet.EmpireID, DestinationSystemID: destination.ID, RemainingTurns: nextRemaining,
			})
			if err != nil {
				return nil, err
			}
			fleet.RemainingTurns = nextRemaining
			events = append(events, event)
			continue
		}
		event, err := NewDomainEvent("empire.fleet_arrived", 0, 0, FleetArrivedEvent{
			FleetID: fleet.ID, EmpireID: fleet.EmpireID, SystemID: destination.ID,
		})
		if err != nil {
			return nil, err
		}
		fleet.AtSystemID = destination.ID
		if empire := empireByID(state, fleet.EmpireID); empire != nil {
			empire.MarkSystemVisited(destination.ID)
		}
		establishFirstContactsAtSystem(state, fleet.EmpireID, destination.ID)
		fleet.DestinationSystemID = 0
		fleet.RemainingTurns = 0
		events = append(events, event)
	}
	return events, nil
}

func establishFirstContactsAtSystem(state *core.GameState, visitingEmpireID, systemID core.ID) {
	visitor := empireByID(state, visitingEmpireID)
	if visitor == nil || systemID == 0 {
		return
	}
	for i := range state.Empires {
		other := &state.Empires[i]
		if other.ID == visitingEmpireID || state.EmpiresHaveContact(visitingEmpireID, other.ID) {
			continue
		}
		if other.HasVisitedSystem(systemID) || empireHasPresenceAtSystem(state, other.ID, systemID) {
			state.MarkEmpiresKnown(visitingEmpireID, other.ID)
		}
	}
}

func empireHasPresenceAtSystem(state *core.GameState, empireID, systemID core.ID) bool {
	for _, fleet := range state.StrategicFleets {
		if fleet.EmpireID == empireID && fleet.AtSystemID == systemID {
			return true
		}
	}
	system := systemByID(state, systemID)
	if system == nil {
		return false
	}
	planetIDs := make(map[core.ID]struct{}, len(system.Planets))
	bodyIDs := make(map[core.ID]struct{}, len(system.Bodies))
	for _, planet := range system.Planets {
		planetIDs[planet.ID] = struct{}{}
	}
	for _, body := range system.Bodies {
		bodyIDs[body.ID] = struct{}{}
	}
	for _, colony := range state.Colonies {
		if colony.EmpireID == empireID {
			if _, ok := planetIDs[colony.PlanetID]; ok {
				return true
			}
		}
	}
	for _, outpost := range state.Outposts {
		if outpost.EmpireID != empireID {
			continue
		}
		if _, ok := planetIDs[outpost.PlanetID]; ok && outpost.PlanetID != 0 {
			return true
		}
		if _, ok := bodyIDs[outpost.BodyID]; ok && outpost.BodyID != 0 {
			return true
		}
	}
	return false
}

func (r *EconomyResolver) colonizePlanet(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) ([]DomainEvent, error) {
	payload, err := decodeColonizePlanet(command)
	if err != nil {
		return nil, err
	}
	fleetIndex, fleet := strategicFleetByID(state, payload.FleetID)
	if fleet == nil {
		return nil, fmt.Errorf("references unknown strategic fleet %d", payload.FleetID)
	}
	if fleet.EmpireID != empireID {
		return nil, fmt.Errorf("seat %d cannot colonize with fleet %d owned by empire %d", seatID, fleet.ID, fleet.EmpireID)
	}
	if fleet.SpecialKind != core.StrategicFleetSpecialColonyShip {
		return nil, fmt.Errorf("fleet %d is not a Colony Ship", fleet.ID)
	}
	if fleet.AtSystemID == 0 || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 {
		return nil, fmt.Errorf("Colony Ship fleet %d must be stationary before colonization", fleet.ID)
	}
	planet := planetByID(state, payload.PlanetID)
	if planet == nil {
		return nil, fmt.Errorf("references unknown planet %d", payload.PlanetID)
	}
	targetSystem := systemForPlanetID(state, planet.ID)
	if targetSystem == nil {
		return nil, fmt.Errorf("planet %d is not assigned to a star system", planet.ID)
	}
	if targetSystem.ID != fleet.AtSystemID {
		return nil, fmt.Errorf("planet %d is in system %d but Colony Ship fleet %d is at system %d", planet.ID, targetSystem.ID, fleet.ID, fleet.AtSystemID)
	}
	if planet.ColonyID != 0 {
		return nil, fmt.Errorf("planet %d is already colonized by colony %d", planet.ID, planet.ColonyID)
	}
	for _, existing := range state.Colonies {
		if existing.PlanetID == planet.ID {
			return nil, fmt.Errorf("planet %d is already referenced by colony %d", planet.ID, existing.ID)
		}
	}
	newColony, _, _, err := r.prepareFoundedColony(state, empireID, planet.ID)
	if err != nil {
		return nil, fmt.Errorf("seat %d Colony Ship founding: %w", seatID, err)
	}
	newColonyID := newColony.ID
	colonized, err := NewDomainEvent("empire.planet_colonized", seatID, command.Sequence, PlanetColonizedEvent{
		EmpireID: empireID, FleetID: fleet.ID, SystemID: targetSystem.ID, PlanetID: planet.ID, ColonyID: newColonyID,
	})
	if err != nil {
		return nil, err
	}
	consumed, err := NewDomainEvent("empire.colony_ship_consumed", seatID, command.Sequence, ColonyShipConsumedEvent{
		EmpireID: empireID, FleetID: fleet.ID, ColonyID: newColonyID, PlanetID: planet.ID,
	})
	if err != nil {
		return nil, err
	}
	convertedOutpostID := planet.OutpostID
	var converted *DomainEvent
	if convertedOutpostID != 0 {
		event, err := NewDomainEvent("empire.outpost_converted", seatID, command.Sequence, OutpostConvertedEvent{
			EmpireID: empireID, OutpostID: convertedOutpostID, PlanetID: planet.ID, ColonyID: newColonyID,
		})
		if err != nil {
			return nil, err
		}
		converted = &event
	}
	if _, err := commitFoundedColony(state, empireID, planet, newColony); err != nil {
		return nil, err
	}
	state.StrategicFleets = append(state.StrategicFleets[:fleetIndex], state.StrategicFleets[fleetIndex+1:]...)
	events := []DomainEvent{colonized}
	if converted != nil {
		events = append(events, *converted)
	}
	events = append(events, consumed)
	return events, nil
}
