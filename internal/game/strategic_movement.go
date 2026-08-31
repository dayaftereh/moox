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

func colonyShipFuelRangeParsecs(empire core.Empire) int {
	known := make(map[int]struct{}, len(empire.KnownTechnologyIDs))
	for _, technologyID := range empire.KnownTechnologyIDs {
		known[technologyID] = struct{}{}
	}
	rangeParsecs := 0
	for _, fuel := range []struct {
		technologyID int
		rangeParsecs int
	}{
		{standardFuelCellsTechnologyID, 4},
		{deuteriumFuelCellsTechnologyID, 6},
		{iridiumFuelCellsTechnologyID, 9},
		{urridiumFuelCellsTechnologyID, 12},
		{thoriumFuelCellsTechnologyID, 255},
	} {
		if _, ok := known[fuel.technologyID]; ok && fuel.rangeParsecs > rangeParsecs {
			rangeParsecs = fuel.rangeParsecs
		}
	}
	return rangeParsecs
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
	for _, colony := range state.Colonies {
		if colony.EmpireID != empireID {
			continue
		}
		supply := systemForPlanetID(state, colony.PlanetID)
		if supply == nil {
			continue
		}
		distance := strategicDistanceParsecs(*supply, destination)
		if !found || distance < nearest {
			nearest = distance
			found = true
		}
	}
	return nearest, found
}

func (r *EconomyResolver) moveFleet(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeMoveFleet(command)
	if err != nil {
		return DomainEvent{}, err
	}
	_, fleet := strategicFleetByID(state, payload.FleetID)
	if fleet == nil {
		return DomainEvent{}, fmt.Errorf("references unknown strategic fleet %d", payload.FleetID)
	}
	if fleet.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot move fleet %d owned by empire %d", seatID, fleet.ID, fleet.EmpireID)
	}
	if fleet.SpecialKind != core.StrategicFleetSpecialColonyShip {
		return DomainEvent{}, fmt.Errorf("fleet %d is not a Colony Ship", fleet.ID)
	}
	if fleet.AtSystemID == 0 || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 {
		return DomainEvent{}, fmt.Errorf("Colony Ship fleet %d is not stationary at a star system", fleet.ID)
	}
	if fleet.FTLSpeed < 2 {
		return DomainEvent{}, fmt.Errorf("Colony Ship fleet %d has invalid ftl_speed %d", fleet.ID, fleet.FTLSpeed)
	}
	source := systemByID(state, fleet.AtSystemID)
	if source == nil {
		return DomainEvent{}, fmt.Errorf("fleet %d references unknown source system %d", fleet.ID, fleet.AtSystemID)
	}
	destination := systemByID(state, payload.DestinationSystemID)
	if destination == nil {
		return DomainEvent{}, fmt.Errorf("references unknown destination star system %d", payload.DestinationSystemID)
	}
	if destination.ID == source.ID {
		return DomainEvent{}, fmt.Errorf("fleet %d is already at star system %d", fleet.ID, source.ID)
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}
	fuelRange := colonyShipFuelRangeParsecs(*empire)
	supplyDistance, hasSupply := nearestEmpireSupplyDistanceParsecs(state, empireID, *destination)
	if !hasSupply || supplyDistance > fuelRange {
		return DomainEvent{}, fmt.Errorf("destination system %d is outside Colony Ship fuel range: nearest supply %d pc, range %d pc", destination.ID, supplyDistance, fuelRange)
	}
	eta := strategicTravelETA(*source, *destination, fleet.FTLSpeed)
	if eta < 1 {
		return DomainEvent{}, fmt.Errorf("invalid strategic ETA %d from system %d to %d", eta, source.ID, destination.ID)
	}
	event, err := NewDomainEvent("empire.fleet_movement_started", seatID, command.Sequence, FleetMovementStartedEvent{
		FleetID: fleet.ID, EmpireID: fleet.EmpireID, SourceSystemID: source.ID, DestinationSystemID: destination.ID,
		RemainingTurns: eta, FTLSpeed: fleet.FTLSpeed, FuelRangeParsecs: fuelRange, SupplyDistanceParsecs: supplyDistance,
	})
	if err != nil {
		return DomainEvent{}, err
	}
	fleet.AtSystemID = 0
	fleet.DestinationSystemID = destination.ID
	fleet.RemainingTurns = eta
	return event, nil
}

func (r *EconomyResolver) advanceStrategicFleetTransit(state *core.GameState) ([]DomainEvent, error) {
	var events []DomainEvent
	for i := range state.StrategicFleets {
		fleet := &state.StrategicFleets[i]
		if fleet.DestinationSystemID == 0 {
			continue
		}
		if fleet.SpecialKind != core.StrategicFleetSpecialColonyShip || fleet.AtSystemID != 0 || fleet.RemainingTurns <= 0 || fleet.FTLSpeed < 2 {
			return nil, fmt.Errorf("strategic fleet %d has invalid Colony Ship transit state", fleet.ID)
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
		fleet.DestinationSystemID = 0
		fleet.RemainingTurns = 0
		events = append(events, event)
	}
	return events, nil
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
	if allocated := state.NewID(); allocated != newColonyID {
		return nil, fmt.Errorf("allocated Colony ID %d does not match expected next_id %d", allocated, newColonyID)
	}
	planet.ColonyID = newColonyID
	state.Colonies = append(state.Colonies, newColony)
	state.StrategicFleets = append(state.StrategicFleets[:fleetIndex], state.StrategicFleets[fleetIndex+1:]...)
	return []DomainEvent{colonized, consumed}, nil
}
