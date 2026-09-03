package game

import (
	"fmt"
	"math"
	"sort"

	"moox/internal/core"
	"moox/internal/protocol"
)

const (
	CommandInvade          = "invasion.invade"
	CommandDeclineInvasion = "invasion.decline"
)

type InvasionHandledKey struct {
	ColonyID         core.ID `json:"colony_id"`
	AttackerEmpireID core.ID `json:"attacker_empire_id"`
}

type InvasionOpportunity struct {
	SystemID                  core.ID         `json:"system_id"`
	ColonyID                  core.ID         `json:"colony_id"`
	AttackerEmpireID          core.ID         `json:"attacker_empire_id"`
	DefenderEmpireID          core.ID         `json:"defender_empire_id"`
	AttackerSeatID            protocol.SeatID `json:"attacker_seat_id"`
	EligibleTransportFleetIDs []core.ID       `json:"eligible_transport_fleet_ids"`
}

type InvadePayload struct {
	ColonyID          core.ID   `json:"colony_id"`
	TransportFleetIDs []core.ID `json:"transport_fleet_ids"`
}

type DeclineInvasionPayload struct {
	ColonyID core.ID `json:"colony_id"`
}

type InvasionResolvedEvent struct {
	SystemID                   core.ID   `json:"system_id"`
	ColonyID                   core.ID   `json:"colony_id"`
	AttackerEmpireID           core.ID   `json:"attacker_empire_id"`
	DefenderEmpireID           core.ID   `json:"defender_empire_id"`
	SelectedTransportFleetIDs  []core.ID `json:"selected_transport_fleet_ids"`
	InitialAttackerInfantry    int       `json:"initial_attacker_infantry"`
	InitialDefenderInfantry    int       `json:"initial_defender_infantry"`
	InitialDefenderMilitia     int       `json:"initial_defender_militia"`
	SurvivingAttackerInfantry  int       `json:"surviving_attacker_infantry"`
	SurvivingDefenderInfantry  int       `json:"surviving_defender_infantry"`
	SurvivingDefenderMilitia   int       `json:"surviving_defender_militia"`
	ConsumedTransportFleetIDs  []core.ID `json:"consumed_transport_fleet_ids,omitempty"`
	SurvivingTransportFleetIDs []core.ID `json:"surviving_transport_fleet_ids,omitempty"`
	Captured                   bool      `json:"captured"`
}

type ColonyConqueredEvent struct {
	SystemID         core.ID `json:"system_id"`
	ColonyID         core.ID `json:"colony_id"`
	PreviousEmpireID core.ID `json:"previous_empire_id"`
	CurrentEmpireID  core.ID `json:"current_empire_id"`
	GarrisonInfantry int     `json:"garrison_infantry"`
}

type InvasionDeclinedEvent struct {
	SystemID         core.ID `json:"system_id"`
	ColonyID         core.ID `json:"colony_id"`
	AttackerEmpireID core.ID `json:"attacker_empire_id"`
	DefenderEmpireID core.ID `json:"defender_empire_id"`
}

func IsInvasionCommand(kind string) bool {
	return kind == CommandInvade || kind == CommandDeclineInvasion
}

func NewInvadeCommand(sequence uint32, payload InvadePayload) (protocol.Command, error) {
	if payload.ColonyID == 0 {
		return protocol.Command{}, fmt.Errorf("colony_id must be non-zero")
	}
	if len(payload.TransportFleetIDs) == 0 {
		return protocol.Command{}, fmt.Errorf("transport_fleet_ids must not be empty")
	}
	return protocol.NewCommand(sequence, CommandInvade, payload)
}

func NewDeclineInvasionCommand(sequence uint32, payload DeclineInvasionPayload) (protocol.Command, error) {
	if payload.ColonyID == 0 {
		return protocol.Command{}, fmt.Errorf("colony_id must be non-zero")
	}
	return protocol.NewCommand(sequence, CommandDeclineInvasion, payload)
}

func decodeInvade(command protocol.Command) (InvadePayload, error) {
	var payload InvadePayload
	if err := decodeStrictCommandPayload(command, CommandInvade, &payload); err != nil {
		return InvadePayload{}, err
	}
	if payload.ColonyID == 0 {
		return InvadePayload{}, fmt.Errorf("colony_id must be non-zero")
	}
	if len(payload.TransportFleetIDs) == 0 {
		return InvadePayload{}, fmt.Errorf("transport_fleet_ids must not be empty")
	}
	if !idsStrictlyIncreasing(payload.TransportFleetIDs) {
		return InvadePayload{}, fmt.Errorf("transport_fleet_ids must be sorted unique")
	}
	return payload, nil
}

func decodeDeclineInvasion(command protocol.Command) (DeclineInvasionPayload, error) {
	var payload DeclineInvasionPayload
	if err := decodeStrictCommandPayload(command, CommandDeclineInvasion, &payload); err != nil {
		return DeclineInvasionPayload{}, err
	}
	if payload.ColonyID == 0 {
		return DeclineInvasionPayload{}, fmt.Errorf("colony_id must be non-zero")
	}
	return payload, nil
}

func idsStrictlyIncreasing(ids []core.ID) bool {
	if len(ids) == 0 {
		return false
	}
	for i := 1; i < len(ids); i++ {
		if ids[i] <= ids[i-1] {
			return false
		}
	}
	return true
}

func CloneInvasionOpportunity(in *InvasionOpportunity) *InvasionOpportunity {
	if in == nil {
		return nil
	}
	out := *in
	out.EligibleTransportFleetIDs = append([]core.ID(nil), in.EligibleTransportFleetIDs...)
	return &out
}

func (c ResolveContext) invasionHandled(colonyID, attackerEmpireID core.ID) bool {
	for _, handled := range c.HandledInvasions {
		if handled.ColonyID == colonyID && handled.AttackerEmpireID == attackerEmpireID {
			return true
		}
	}
	return false
}

func (r *EconomyResolver) prepareInvasionBoundary(ctx ResolveContext, state *core.GameState) (*InvasionOpportunity, error) {
	if r == nil || r.Rules == nil {
		return nil, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	candidates := make([]InvasionOpportunity, 0)
	for ci := range state.Colonies {
		colony := &state.Colonies[ci]
		system := systemForPlanetID(state, colony.PlanetID)
		if system == nil {
			return nil, fmt.Errorf("colony %d planet %d is not assigned to a star system", colony.ID, colony.PlanetID)
		}
		if defenderHasModeledOrbitalStation(state, colony.EmpireID, system.ID) {
			continue
		}
		byAttacker := make(map[core.ID][]core.ID)
		for fi := range state.StrategicFleets {
			fleet := &state.StrategicFleets[fi]
			if fleet.AtSystemID != system.ID || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 || fleet.SpecialKind != core.StrategicFleetSpecialTroopTransport {
				continue
			}
			if fleet.EmpireID == colony.EmpireID || !state.MayAttackEmpire(fleet.EmpireID, colony.EmpireID) || ctx.invasionHandled(colony.ID, fleet.EmpireID) {
				continue
			}
			if _, ok := ctx.SeatForEmpire(fleet.EmpireID); !ok {
				continue
			}
			byAttacker[fleet.EmpireID] = append(byAttacker[fleet.EmpireID], fleet.ID)
		}
		colonyCandidates := make([]InvasionOpportunity, 0, len(byAttacker))
		for attackerEmpireID, transportIDs := range byAttacker {
			if !hasStationaryCombatFleet(state, attackerEmpireID, system.ID) || hasModeledHostileCombatPresence(state, attackerEmpireID, system.ID) {
				continue
			}
			seatID, _ := ctx.SeatForEmpire(attackerEmpireID)
			colonyCandidates = append(colonyCandidates, InvasionOpportunity{
				SystemID: system.ID, ColonyID: colony.ID, AttackerEmpireID: attackerEmpireID, DefenderEmpireID: colony.EmpireID,
				AttackerSeatID: seatID, EligibleTransportFleetIDs: sortedUniqueIDs(transportIDs),
			})
		}
		// Same-colony multi-attacker resolution is explicitly deferred in Slice 11.
		if len(colonyCandidates) == 1 {
			candidates = append(candidates, colonyCandidates[0])
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].SystemID != candidates[j].SystemID {
			return candidates[i].SystemID < candidates[j].SystemID
		}
		if candidates[i].ColonyID != candidates[j].ColonyID {
			return candidates[i].ColonyID < candidates[j].ColonyID
		}
		return candidates[i].AttackerEmpireID < candidates[j].AttackerEmpireID
	})
	return CloneInvasionOpportunity(&candidates[0]), nil
}

func hasStationaryCombatFleet(state *core.GameState, empireID, systemID core.ID) bool {
	for i := range state.StrategicFleets {
		fleet := &state.StrategicFleets[i]
		if fleet.EmpireID == empireID && fleet.AtSystemID == systemID && fleet.DestinationSystemID == 0 && fleet.RemainingTurns == 0 && fleet.Role == core.StrategicFleetRoleCombat && fleet.SpecialKind == core.StrategicFleetSpecialNone && len(fleet.ShipIDs) != 0 {
			return true
		}
	}
	return false
}

func hasModeledHostileCombatPresence(state *core.GameState, attackerEmpireID, systemID core.ID) bool {
	for i := range state.StrategicFleets {
		fleet := &state.StrategicFleets[i]
		if fleet.EmpireID == attackerEmpireID || fleet.AtSystemID != systemID || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 || fleet.Role != core.StrategicFleetRoleCombat || fleet.SpecialKind != core.StrategicFleetSpecialNone || len(fleet.ShipIDs) == 0 {
			continue
		}
		if state.MayAttackEmpire(attackerEmpireID, fleet.EmpireID) {
			return true
		}
	}
	return false
}

func defenderHasModeledOrbitalStation(state *core.GameState, defenderEmpireID, systemID core.ID) bool {
	for i := range state.Colonies {
		colony := &state.Colonies[i]
		if colony.EmpireID != defenderEmpireID {
			continue
		}
		system := systemForPlanetID(state, colony.PlanetID)
		if system == nil || system.ID != systemID {
			continue
		}
		if colonyOwnsBuilding(colony, commandStationStarBase) || colonyOwnsBuilding(colony, commandStationBattlestation) || colonyOwnsBuilding(colony, commandStationStarFortress) {
			return true
		}
	}
	return false
}

func sameInvasionOpportunity(a, b *InvasionOpportunity) bool {
	if a == nil || b == nil || a.SystemID != b.SystemID || a.ColonyID != b.ColonyID || a.AttackerEmpireID != b.AttackerEmpireID || a.DefenderEmpireID != b.DefenderEmpireID || a.AttackerSeatID != b.AttackerSeatID || len(a.EligibleTransportFleetIDs) != len(b.EligibleTransportFleetIDs) {
		return false
	}
	for i := range a.EligibleTransportFleetIDs {
		if a.EligibleTransportFleetIDs[i] != b.EligibleTransportFleetIDs[i] {
			return false
		}
	}
	return true
}

func (r *EconomyResolver) ResolveInvasionCommand(ctx ResolveContext, state *core.GameState, opportunity InvasionOpportunity, seatID protocol.SeatID, command protocol.Command) ([]DomainEvent, error) {
	if err := command.Validate(1); err != nil {
		return nil, fmt.Errorf("invalid invasion command: %w", err)
	}
	if !IsInvasionCommand(command.Kind) {
		return nil, fmt.Errorf("command kind %q is not an invasion command", command.Kind)
	}
	actorEmpireID, ok := ctx.EmpireForSeat(seatID)
	if !ok || actorEmpireID != opportunity.AttackerEmpireID || seatID != opportunity.AttackerSeatID {
		return nil, fmt.Errorf("seat %d has no authority for invasion of colony %d", seatID, opportunity.ColonyID)
	}
	current, err := r.prepareInvasionBoundary(ctx, state)
	if err != nil {
		return nil, err
	}
	if !sameInvasionOpportunity(current, &opportunity) {
		return nil, fmt.Errorf("invasion opportunity is no longer authoritative")
	}
	if command.Kind == CommandDeclineInvasion {
		payload, err := decodeDeclineInvasion(command)
		if err != nil {
			return nil, err
		}
		if payload.ColonyID != opportunity.ColonyID {
			return nil, fmt.Errorf("decline colony_id %d does not match pending colony %d", payload.ColonyID, opportunity.ColonyID)
		}
		event, err := NewDomainEvent("empire.invasion_declined", seatID, command.Sequence, InvasionDeclinedEvent{SystemID: opportunity.SystemID, ColonyID: opportunity.ColonyID, AttackerEmpireID: opportunity.AttackerEmpireID, DefenderEmpireID: opportunity.DefenderEmpireID})
		if err != nil {
			return nil, err
		}
		return []DomainEvent{event}, nil
	}
	payload, err := decodeInvade(command)
	if err != nil {
		return nil, err
	}
	if payload.ColonyID != opportunity.ColonyID {
		return nil, fmt.Errorf("invade colony_id %d does not match pending colony %d", payload.ColonyID, opportunity.ColonyID)
	}
	eligible := make(map[core.ID]struct{}, len(opportunity.EligibleTransportFleetIDs))
	for _, id := range opportunity.EligibleTransportFleetIDs {
		eligible[id] = struct{}{}
	}
	for _, id := range payload.TransportFleetIDs {
		if _, ok := eligible[id]; !ok {
			return nil, fmt.Errorf("transport fleet %d is not eligible for pending invasion", id)
		}
	}
	return r.resolveGroundInvasion(state, opportunity, seatID, command.Sequence, payload.TransportFleetIDs)
}

func (r *EconomyResolver) resolveGroundInvasion(state *core.GameState, opportunity InvasionOpportunity, seatID protocol.SeatID, commandSequence uint32, selected []core.ID) ([]DomainEvent, error) {
	colony := colonyByID(state, opportunity.ColonyID)
	if colony == nil || colony.EmpireID != opportunity.DefenderEmpireID {
		return nil, fmt.Errorf("pending invasion colony is unavailable")
	}
	initialAttacker := len(selected) * 4
	attacker := initialAttacker
	initialDefenderInfantry := colony.GroundForces.Infantry
	defenderInfantry := initialDefenderInfantry
	initialMilitia := colonyMilitia(colony)
	defenderMilitia := initialMilitia
	rng := state.RNG()
	for attacker > 0 && defenderInfantry+defenderMilitia > 0 {
		aRoll, err := rng.Intn(100)
		if err != nil {
			return nil, err
		}
		dRoll, err := rng.Intn(100)
		if err != nil {
			return nil, err
		}
		if defenderInfantry == 0 {
			dRoll -= 10
		}
		loseAttacker := aRoll <= dRoll
		loseDefender := aRoll >= dRoll
		if loseAttacker {
			attacker--
		}
		if loseDefender {
			if defenderInfantry > 0 {
				defenderInfantry--
			} else if defenderMilitia > 0 {
				defenderMilitia--
			}
		}
	}
	state.CommitRNG(rng)
	captured := attacker > 0
	consumed := make([]core.ID, 0, len(selected))
	survivingTransports := make([]core.ID, 0, len(selected))
	if !captured {
		colony.GroundForces.Infantry = defenderInfantry
		consumed = append(consumed, selected...)
		removeStrategicFleetsByID(state, idSet(consumed))
	} else {
		remaining := attacker - 1
		for i := len(selected) - 1; i >= 0; i-- {
			if remaining >= 4 {
				remaining -= 4
				survivingTransports = append(survivingTransports, selected[i])
			} else {
				consumed = append(consumed, selected[i])
			}
		}
		sort.Slice(survivingTransports, func(i, j int) bool { return survivingTransports[i] < survivingTransports[j] })
		sort.Slice(consumed, func(i, j int) bool { return consumed[i] < consumed[j] })
		removeStrategicFleetsByID(state, idSet(consumed))
		colony.GroundForces.Infantry = 1 + remaining
		if err := captureColony(state, colony, opportunity.AttackerEmpireID, opportunity.DefenderEmpireID); err != nil {
			return nil, err
		}
	}
	resolvedData := InvasionResolvedEvent{
		SystemID: opportunity.SystemID, ColonyID: opportunity.ColonyID, AttackerEmpireID: opportunity.AttackerEmpireID, DefenderEmpireID: opportunity.DefenderEmpireID,
		SelectedTransportFleetIDs: append([]core.ID(nil), selected...), InitialAttackerInfantry: initialAttacker, InitialDefenderInfantry: initialDefenderInfantry, InitialDefenderMilitia: initialMilitia,
		SurvivingAttackerInfantry: attacker, SurvivingDefenderInfantry: defenderInfantry, SurvivingDefenderMilitia: defenderMilitia,
		ConsumedTransportFleetIDs: consumed, SurvivingTransportFleetIDs: survivingTransports, Captured: captured,
	}
	resolved, err := NewDomainEvent("empire.invasion_resolved", seatID, commandSequence, resolvedData)
	if err != nil {
		return nil, err
	}
	events := []DomainEvent{resolved}
	if captured {
		conquered, err := NewDomainEvent("empire.colony_conquered", seatID, commandSequence, ColonyConqueredEvent{SystemID: opportunity.SystemID, ColonyID: opportunity.ColonyID, PreviousEmpireID: opportunity.DefenderEmpireID, CurrentEmpireID: opportunity.AttackerEmpireID, GarrisonInfantry: colony.GroundForces.Infantry})
		if err != nil {
			return nil, err
		}
		events = append(events, conquered)
	}
	return events, nil
}

func colonyMilitia(colony *core.Colony) int {
	if colony == nil {
		return 0
	}
	total := 0.0
	for _, cohort := range colony.Population.Cohorts {
		if cohort.AssimilationState == core.PopulationAssimilated {
			total += cohort.Total()
		}
	}
	return int(math.Floor(total / 5.0))
}

func captureColony(state *core.GameState, colony *core.Colony, attackerEmpireID, defenderEmpireID core.ID) error {
	if colony == nil {
		return fmt.Errorf("captured colony is nil")
	}
	colony.EmpireID = attackerEmpireID
	colony.Construction = nil
	for i := range colony.Population.Cohorts {
		cohort := &colony.Population.Cohorts[i]
		if cohort.OriginEmpireID == attackerEmpireID {
			cohort.LoyaltyEmpireID = attackerEmpireID
			cohort.AssimilationState = core.PopulationAssimilated
		} else {
			cohort.LoyaltyEmpireID = cohort.OriginEmpireID
			cohort.AssimilationState = core.PopulationConquered
		}
	}
	colony.Population.Normalize()
	attacker := empireByID(state, attackerEmpireID)
	defender := empireByID(state, defenderEmpireID)
	if attacker == nil || defender == nil {
		return fmt.Errorf("capture references unknown attacker/defender empire")
	}
	if defender.Capital == colony.ID {
		defender.Capital = replacementCapital(state, defenderEmpireID)
	}
	if attacker.Capital == 0 {
		attacker.Capital = colony.ID
	}
	return nil
}

func replacementCapital(state *core.GameState, empireID core.ID) core.ID {
	var bestID core.ID
	bestPop := -1.0
	for i := range state.Colonies {
		colony := &state.Colonies[i]
		if colony.EmpireID != empireID {
			continue
		}
		population := colony.Population.Total()
		if population > bestPop || (population == bestPop && (bestID == 0 || colony.ID < bestID)) {
			bestPop = population
			bestID = colony.ID
		}
	}
	return bestID
}

func (r *EconomyResolver) ResumeAfterInvasion(ctx ResolveContext, state *core.GameState) (Resolution, error) {
	if state == nil {
		return Resolution{}, fmt.Errorf("game state must not be nil")
	}
	if invasion, err := r.prepareInvasionBoundary(ctx, state); err != nil {
		return Resolution{}, err
	} else if invasion != nil {
		return Resolution{State: state, Invasion: invasion}, nil
	}
	boundaryEvents, encounters, err := r.prepareEncounterBoundary(ctx, state)
	if err != nil {
		return Resolution{}, err
	}
	if len(encounters) != 0 {
		return Resolution{State: state, Events: boundaryEvents, Encounters: encounters}, nil
	}
	postEvents, err := r.finishPostEncounter(state)
	if err != nil {
		return Resolution{}, err
	}
	boundaryEvents = append(boundaryEvents, postEvents...)
	return Resolution{State: state, Events: boundaryEvents}, nil
}
