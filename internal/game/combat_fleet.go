package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
	"moox/internal/protocol"
)

type FleetSplitEvent struct {
	EmpireID      core.ID   `json:"empire_id"`
	SourceFleetID core.ID   `json:"source_fleet_id"`
	NewFleetID    core.ID   `json:"new_fleet_id"`
	SystemID      core.ID   `json:"system_id"`
	MovedShipIDs  []core.ID `json:"moved_ship_ids"`
}

type FleetsMergedEvent struct {
	EmpireID      core.ID   `json:"empire_id"`
	TargetFleetID core.ID   `json:"target_fleet_id"`
	SourceFleetID core.ID   `json:"source_fleet_id"`
	SystemID      core.ID   `json:"system_id"`
	ShipIDs       []core.ID `json:"ship_ids"`
}

func shipByID(state *core.GameState, id core.ID) *core.Ship {
	if state == nil || id == 0 {
		return nil
	}
	for i := range state.Ships {
		if state.Ships[i].ID == id {
			return &state.Ships[i]
		}
	}
	return nil
}

func (r *EconomyResolver) combatFleetMovementProfile(state *core.GameState, fleet core.StrategicFleet, empire *core.Empire) (int, int, error) {
	if r == nil || r.Rules == nil {
		return 0, 0, fmt.Errorf("economy resolver has no rules")
	}
	if empire == nil {
		return 0, 0, fmt.Errorf("combat fleet has no owning empire")
	}
	if fleet.Role != core.StrategicFleetRoleCombat || fleet.SpecialKind != core.StrategicFleetSpecialNone {
		return 0, 0, fmt.Errorf("fleet %d is not an ordinary combat fleet", fleet.ID)
	}
	if len(fleet.ShipIDs) == 0 {
		return 0, 0, fmt.Errorf("combat fleet %d has no concrete ships", fleet.ID)
	}
	drive, ok := bestKnownShipDrive(r.Rules.ShipDrives, empire)
	if !ok || drive.FTLSpeed < 2 {
		return 0, 0, fmt.Errorf("empire %d has no supported strategic Warp Drive", empire.ID)
	}
	fuel, ok := bestKnownShipFuelCell(r.Rules.ShipFuelCells, empire)
	if !ok || fuel.RangeParsecs <= 0 {
		return 0, 0, fmt.Errorf("empire %d has no supported strategic Fuel Cell", empire.ID)
	}
	ftlSpeed := drive.FTLSpeed
	if modifiers, ok := r.Rules.RaceModifiers[empire.RaceID]; ok && modifiers.TransDimensional {
		ftlSpeed += 2
	}
	effectiveRange := fuel.RangeParsecs
	for _, shipID := range fleet.ShipIDs {
		ship := shipByID(state, shipID)
		if ship == nil {
			return 0, 0, fmt.Errorf("combat fleet %d references unknown ship %d", fleet.ID, shipID)
		}
		if ship.EmpireID != fleet.EmpireID {
			return 0, 0, fmt.Errorf("combat fleet %d ship %d belongs to empire %d", fleet.ID, shipID, ship.EmpireID)
		}
		// Slice 04 has no supported per-Ship strategic range special yet. Keeping
		// this member loop makes the future Extended Fuel Tanks hook explicit.
		memberRange := fuel.RangeParsecs
		if memberRange < effectiveRange {
			effectiveRange = memberRange
		}
	}
	return ftlSpeed, effectiveRange, nil
}

func partitionFleetShipIDs(fleetShipIDs, selected []core.ID) (remaining []core.ID, moved []core.ID, all bool, err error) {
	if len(selected) == 0 {
		return nil, append([]core.ID(nil), fleetShipIDs...), true, nil
	}
	membership := make(map[core.ID]struct{}, len(fleetShipIDs))
	for _, shipID := range fleetShipIDs {
		membership[shipID] = struct{}{}
	}
	for _, shipID := range selected {
		if _, ok := membership[shipID]; !ok {
			return nil, nil, false, fmt.Errorf("selected ship %d is not a member of the fleet", shipID)
		}
	}
	moved = append([]core.ID(nil), selected...)
	if len(selected) == len(fleetShipIDs) {
		return nil, moved, true, nil
	}
	selectedSet := make(map[core.ID]struct{}, len(selected))
	for _, shipID := range selected {
		selectedSet[shipID] = struct{}{}
	}
	remaining = make([]core.ID, 0, len(fleetShipIDs)-len(selected))
	for _, shipID := range fleetShipIDs {
		if _, selected := selectedSet[shipID]; !selected {
			remaining = append(remaining, shipID)
		}
	}
	return remaining, moved, false, nil
}

func (r *EconomyResolver) moveCombatFleet(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command, payload MoveFleetPayload, fleet *core.StrategicFleet) ([]DomainEvent, error) {
	if fleet.Role != core.StrategicFleetRoleCombat || fleet.SpecialKind != core.StrategicFleetSpecialNone {
		return nil, fmt.Errorf("fleet %d is not an ordinary combat fleet", fleet.ID)
	}
	if fleet.AtSystemID == 0 || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 || fleet.FTLSpeed != 0 {
		return nil, fmt.Errorf("combat fleet %d is not stationary at a star system", fleet.ID)
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
	remaining, moved, all, err := partitionFleetShipIDs(fleet.ShipIDs, payload.ShipIDs)
	if err != nil {
		return nil, err
	}
	if len(moved) == 0 {
		return nil, fmt.Errorf("combat fleet %d has no ships to move", fleet.ID)
	}
	profileFleet := *fleet
	profileFleet.ShipIDs = moved
	ftlSpeed, fuelRange, err := r.combatFleetMovementProfile(state, profileFleet, empire)
	if err != nil {
		return nil, err
	}
	supplyDistance, hasSupply := nearestEmpireSupplyDistanceParsecs(state, empireID, *destination)
	if !hasSupply || supplyDistance > fuelRange {
		return nil, fmt.Errorf("destination system %d is outside combat fleet fuel range: nearest supply %d pc, range %d pc", destination.ID, supplyDistance, fuelRange)
	}
	eta := strategicTravelETA(*source, *destination, ftlSpeed)
	if eta < 1 {
		return nil, fmt.Errorf("invalid strategic ETA %d from system %d to %d", eta, source.ID, destination.ID)
	}

	movingFleetID := fleet.ID
	predictedFleetID := core.ID(0)
	var splitEvent DomainEvent
	if !all {
		if len(remaining) == 0 {
			return nil, fmt.Errorf("combat fleet %d split would leave source empty", fleet.ID)
		}
		predictedFleetID = state.NextID
		if predictedFleetID == 0 {
			return nil, fmt.Errorf("cannot allocate split fleet id from next_id 0")
		}
		movingFleetID = predictedFleetID
		splitEvent, err = NewDomainEvent("empire.fleet_split", seatID, command.Sequence, FleetSplitEvent{
			EmpireID: empireID, SourceFleetID: fleet.ID, NewFleetID: predictedFleetID, SystemID: source.ID,
			MovedShipIDs: append([]core.ID(nil), moved...),
		})
		if err != nil {
			return nil, err
		}
	}
	startedEvent, err := NewDomainEvent("empire.fleet_movement_started", seatID, command.Sequence, FleetMovementStartedEvent{
		FleetID: movingFleetID, EmpireID: empireID, SourceSystemID: source.ID, DestinationSystemID: destination.ID,
		RemainingTurns: eta, FTLSpeed: ftlSpeed, FuelRangeParsecs: fuelRange, SupplyDistanceParsecs: supplyDistance,
	})
	if err != nil {
		return nil, err
	}

	targetFleet := fleet
	events := make([]DomainEvent, 0, 2)
	if !all {
		fleet.ShipIDs = append([]core.ID(nil), remaining...)
		allocated := state.NewID()
		if allocated != predictedFleetID {
			panic("GameState.NewID returned an unexpected id")
		}
		state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
			ID: predictedFleetID, EmpireID: empireID, Role: core.StrategicFleetRoleCombat, SpecialKind: core.StrategicFleetSpecialNone,
			AtSystemID: source.ID, ShipIDs: append([]core.ID(nil), moved...),
		})
		sort.Slice(state.StrategicFleets, func(i, j int) bool { return state.StrategicFleets[i].ID < state.StrategicFleets[j].ID })
		_, targetFleet = strategicFleetByID(state, predictedFleetID)
		if targetFleet == nil {
			panic("new split fleet disappeared after insertion")
		}
		events = append(events, splitEvent)
	}
	targetFleet.AtSystemID = 0
	targetFleet.SourceSystemID = source.ID
	targetFleet.DestinationSystemID = destination.ID
	targetFleet.RemainingTurns = eta
	targetFleet.TransitTurnsTotal = eta
	targetFleet.FTLSpeed = 0
	events = append(events, startedEvent)
	return events, nil
}

func (r *EconomyResolver) splitFleet(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeSplitFleet(command)
	if err != nil {
		return DomainEvent{}, err
	}
	_, fleet := strategicFleetByID(state, payload.FleetID)
	if fleet == nil {
		return DomainEvent{}, fmt.Errorf("references unknown strategic fleet %d", payload.FleetID)
	}
	if fleet.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot split fleet %d owned by empire %d", seatID, fleet.ID, fleet.EmpireID)
	}
	if fleet.Role != core.StrategicFleetRoleCombat || fleet.SpecialKind != core.StrategicFleetSpecialNone {
		return DomainEvent{}, fmt.Errorf("fleet %d is not an ordinary combat fleet", fleet.ID)
	}
	if fleet.AtSystemID == 0 || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 || fleet.FTLSpeed != 0 {
		return DomainEvent{}, fmt.Errorf("combat fleet %d must be stationary before split", fleet.ID)
	}
	if systemByID(state, fleet.AtSystemID) == nil {
		return DomainEvent{}, fmt.Errorf("fleet %d references unknown source system %d", fleet.ID, fleet.AtSystemID)
	}
	remaining, moved, all, err := partitionFleetShipIDs(fleet.ShipIDs, payload.ShipIDs)
	if err != nil {
		return DomainEvent{}, err
	}
	if all || len(remaining) == 0 {
		return DomainEvent{}, fmt.Errorf("split_fleet ship_ids must be a proper non-empty subset of fleet %d", fleet.ID)
	}
	predictedFleetID := state.NextID
	if predictedFleetID == 0 {
		return DomainEvent{}, fmt.Errorf("cannot allocate split fleet id from next_id 0")
	}
	event, err := NewDomainEvent("empire.fleet_split", seatID, command.Sequence, FleetSplitEvent{
		EmpireID: empireID, SourceFleetID: fleet.ID, NewFleetID: predictedFleetID, SystemID: fleet.AtSystemID,
		MovedShipIDs: append([]core.ID(nil), moved...),
	})
	if err != nil {
		return DomainEvent{}, err
	}

	fleet.ShipIDs = append([]core.ID(nil), remaining...)
	allocated := state.NewID()
	if allocated != predictedFleetID {
		panic("GameState.NewID returned an unexpected id")
	}
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID: predictedFleetID, EmpireID: empireID, Role: core.StrategicFleetRoleCombat, SpecialKind: core.StrategicFleetSpecialNone,
		AtSystemID: fleet.AtSystemID, ShipIDs: append([]core.ID(nil), moved...),
	})
	sort.Slice(state.StrategicFleets, func(i, j int) bool { return state.StrategicFleets[i].ID < state.StrategicFleets[j].ID })
	return event, nil
}

func (r *EconomyResolver) mergeFleets(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeMergeFleets(command)
	if err != nil {
		return DomainEvent{}, err
	}
	targetIndex, target := strategicFleetByID(state, payload.TargetFleetID)
	sourceIndex, source := strategicFleetByID(state, payload.SourceFleetID)
	if target == nil {
		return DomainEvent{}, fmt.Errorf("references unknown target strategic fleet %d", payload.TargetFleetID)
	}
	if source == nil {
		return DomainEvent{}, fmt.Errorf("references unknown source strategic fleet %d", payload.SourceFleetID)
	}
	if target.EmpireID != empireID || source.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d can only merge fleets owned by empire %d", seatID, empireID)
	}
	for _, fleet := range []*core.StrategicFleet{target, source} {
		if fleet.Role != core.StrategicFleetRoleCombat || fleet.SpecialKind != core.StrategicFleetSpecialNone {
			return DomainEvent{}, fmt.Errorf("fleet %d is not an ordinary combat fleet", fleet.ID)
		}
		if fleet.AtSystemID == 0 || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 || fleet.FTLSpeed != 0 {
			return DomainEvent{}, fmt.Errorf("combat fleet %d must be stationary before merge", fleet.ID)
		}
	}
	if target.AtSystemID != source.AtSystemID {
		return DomainEvent{}, fmt.Errorf("fleets %d and %d are not at the same star system", target.ID, source.ID)
	}
	if systemByID(state, target.AtSystemID) == nil {
		return DomainEvent{}, fmt.Errorf("fleets reference unknown star system %d", target.AtSystemID)
	}
	seen := make(map[core.ID]struct{}, len(target.ShipIDs)+len(source.ShipIDs))
	combined := make([]core.ID, 0, len(target.ShipIDs)+len(source.ShipIDs))
	for _, shipID := range append(append([]core.ID(nil), target.ShipIDs...), source.ShipIDs...) {
		if _, exists := seen[shipID]; exists {
			return DomainEvent{}, fmt.Errorf("merge would duplicate ship %d", shipID)
		}
		seen[shipID] = struct{}{}
		combined = append(combined, shipID)
	}
	sort.Slice(combined, func(i, j int) bool { return combined[i] < combined[j] })
	if len(combined) == 0 {
		return DomainEvent{}, fmt.Errorf("merged combat fleet would be empty")
	}
	event, err := NewDomainEvent("empire.fleets_merged", seatID, command.Sequence, FleetsMergedEvent{
		EmpireID: empireID, TargetFleetID: target.ID, SourceFleetID: source.ID, SystemID: target.AtSystemID,
		ShipIDs: append([]core.ID(nil), combined...),
	})
	if err != nil {
		return DomainEvent{}, err
	}

	state.StrategicFleets[targetIndex].ShipIDs = combined
	state.StrategicFleets = append(state.StrategicFleets[:sourceIndex], state.StrategicFleets[sourceIndex+1:]...)
	return event, nil
}
