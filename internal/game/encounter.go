package game

import (
	"fmt"
	"sort"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/protocol"
)

type CivilianFleetsOverrunEvent struct {
	SystemID         core.ID   `json:"system_id"`
	AttackerEmpireID core.ID   `json:"attacker_empire_id"`
	DefenderEmpireID core.ID   `json:"defender_empire_id"`
	CivilianFleetIDs []core.ID `json:"civilian_fleet_ids"`
}

type BattleCasualtiesAppliedEvent struct {
	BattleID         uint64    `json:"battle_id"`
	SystemID         core.ID   `json:"system_id"`
	DestroyedShipIDs []core.ID `json:"destroyed_ship_ids"`
}

type FleetRetreatedAfterBattleEvent struct {
	BattleID            uint64  `json:"battle_id"`
	FleetID             core.ID `json:"fleet_id"`
	EmpireID            core.ID `json:"empire_id"`
	SourceSystemID      core.ID `json:"source_system_id"`
	DestinationSystemID core.ID `json:"destination_system_id"`
	RemainingTurns      int     `json:"remaining_turns"`
}

type FleetDestroyedAfterBattleEvent struct {
	BattleID         uint64    `json:"battle_id"`
	FleetID          core.ID   `json:"fleet_id"`
	EmpireID         core.ID   `json:"empire_id"`
	SystemID         core.ID   `json:"system_id"`
	DestroyedShipIDs []core.ID `json:"destroyed_ship_ids,omitempty"`
	Reason           string    `json:"reason"`
}

type encounterPresence struct {
	side      EncounterSide
	colonyIDs []core.ID
}

func cloneEncounterSide(side EncounterSide) EncounterSide {
	return EncounterSide{
		EmpireID:         side.EmpireID,
		SeatID:           side.SeatID,
		CombatFleetIDs:   append([]core.ID(nil), side.CombatFleetIDs...),
		ShipIDs:          append([]core.ID(nil), side.ShipIDs...),
		CivilianFleetIDs: append([]core.ID(nil), side.CivilianFleetIDs...),
	}
}

func CloneEncounter(encounter Encounter) Encounter {
	out := Encounter{
		SystemID:                  encounter.SystemID,
		Attacker:                  cloneEncounterSide(encounter.Attacker),
		Defender:                  cloneEncounterSide(encounter.Defender),
		DefenderColonyIDs:         append([]core.ID(nil), encounter.DefenderColonyIDs...),
		Participants:              append([]protocol.SeatID(nil), encounter.Participants...),
		TacticalUnsupportedReason: encounter.TacticalUnsupportedReason,
	}
	if encounter.Tactical != nil {
		tactical := battle.CloneTacticalSpec(*encounter.Tactical)
		out.Tactical = &tactical
	}
	return out
}

func (r *EconomyResolver) prepareEncounterBoundary(ctx ResolveContext, state *core.GameState) ([]DomainEvent, []Encounter, error) {
	if r == nil || r.Rules == nil {
		return nil, nil, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return nil, nil, fmt.Errorf("game state must not be nil")
	}

	var events []DomainEvent
	presence, systemIDs, err := buildEncounterPresence(state)
	if err != nil {
		return nil, nil, err
	}

	// Direct original HELP/executable evidence makes unescorted Colony/Outpost
	// ships non-combat assets. Resolve those deterministic no-colony overruns before
	// constructing tactical BattleSessions.
	for _, systemID := range systemIDs {
		byEmpire := presence[systemID]
		empireIDs := sortedEncounterEmpireIDs(byEmpire)
		for _, attackerID := range empireIDs {
			attacker := byEmpire[attackerID]
			if attacker == nil || len(attacker.side.ShipIDs) == 0 {
				continue
			}
			for _, defenderID := range empireIDs {
				if defenderID == attackerID || state.DiplomaticStanceBetween(attackerID, defenderID) != core.DiplomaticStanceHostile {
					continue
				}
				defender := byEmpire[defenderID]
				if defender == nil || len(defender.side.ShipIDs) != 0 || len(defender.side.CivilianFleetIDs) == 0 || len(defender.colonyIDs) != 0 {
					continue
				}
				if _, ok := ctx.SeatForEmpire(attackerID); !ok {
					return nil, nil, fmt.Errorf("hostile encounter attacker empire %d has no session seat", attackerID)
				}
				if _, ok := ctx.SeatForEmpire(defenderID); !ok {
					return nil, nil, fmt.Errorf("hostile encounter defender empire %d has no session seat", defenderID)
				}
				fleetIDs := append([]core.ID(nil), defender.side.CivilianFleetIDs...)
				removeStrategicFleetsByID(state, idSet(fleetIDs))
				event, err := NewDomainEvent("empire.civilian_fleets_overrun", 0, 0, CivilianFleetsOverrunEvent{
					SystemID: systemID, AttackerEmpireID: attackerID, DefenderEmpireID: defenderID,
					CivilianFleetIDs: fleetIDs,
				})
				if err != nil {
					return nil, nil, err
				}
				events = append(events, event)
				defender.side.CivilianFleetIDs = nil
			}
		}
	}

	// Rebuild from authoritative state after immediate overruns so every encounter
	// identity list is an exact frozen reference set.
	presence, systemIDs, err = buildEncounterPresence(state)
	if err != nil {
		return nil, nil, err
	}
	encounters := make([]Encounter, 0)
	for _, systemID := range systemIDs {
		byEmpire := presence[systemID]
		empireIDs := sortedEncounterEmpireIDs(byEmpire)
		var chosen *Encounter
		for _, attackerID := range empireIDs {
			attacker := byEmpire[attackerID]
			if attacker == nil || len(attacker.side.ShipIDs) == 0 {
				continue
			}
			for _, defenderID := range empireIDs {
				if defenderID == attackerID || state.DiplomaticStanceBetween(attackerID, defenderID) != core.DiplomaticStanceHostile {
					continue
				}
				defender := byEmpire[defenderID]
				if defender == nil || len(defender.side.ShipIDs) == 0 {
					continue
				}
				attackerSeat, ok := ctx.SeatForEmpire(attackerID)
				if !ok || attackerSeat == 0 {
					return nil, nil, fmt.Errorf("hostile encounter attacker empire %d has no session seat", attackerID)
				}
				defenderSeat, ok := ctx.SeatForEmpire(defenderID)
				if !ok || defenderSeat == 0 {
					return nil, nil, fmt.Errorf("hostile encounter defender empire %d has no session seat", defenderID)
				}
				attackerSide := cloneEncounterSide(attacker.side)
				defenderSide := cloneEncounterSide(defender.side)
				attackerSide.SeatID = attackerSeat
				defenderSide.SeatID = defenderSeat
				participants := []protocol.SeatID{attackerSeat, defenderSeat}
				sort.Slice(participants, func(i, j int) bool { return participants[i] < participants[j] })
				candidate := Encounter{
					SystemID:          systemID,
					Attacker:          attackerSide,
					Defender:          defenderSide,
					DefenderColonyIDs: append([]core.ID(nil), defender.colonyIDs...),
					Participants:      participants,
				}
				tactical, unsupportedReason, err := tacticalMetadataForEncounter(state, candidate, r.Rules.TacticalCombat)
				if err != nil {
					return nil, nil, err
				}
				candidate.Tactical = tactical
				candidate.TacticalUnsupportedReason = unsupportedReason
				chosen = &candidate
				break
			}
			if chosen != nil {
				break
			}
		}
		if chosen != nil {
			encounters = append(encounters, *chosen)
		}
	}
	return events, encounters, nil
}

func buildEncounterPresence(state *core.GameState) (map[core.ID]map[core.ID]*encounterPresence, []core.ID, error) {
	presence := make(map[core.ID]map[core.ID]*encounterPresence)
	ensure := func(systemID, empireID core.ID) *encounterPresence {
		byEmpire := presence[systemID]
		if byEmpire == nil {
			byEmpire = make(map[core.ID]*encounterPresence)
			presence[systemID] = byEmpire
		}
		entry := byEmpire[empireID]
		if entry == nil {
			entry = &encounterPresence{side: EncounterSide{EmpireID: empireID}}
			byEmpire[empireID] = entry
		}
		return entry
	}

	for i := range state.StrategicFleets {
		fleet := &state.StrategicFleets[i]
		if fleet.AtSystemID == 0 {
			continue
		}
		entry := ensure(fleet.AtSystemID, fleet.EmpireID)
		switch {
		case fleet.Role == core.StrategicFleetRoleCombat && fleet.SpecialKind == core.StrategicFleetSpecialNone:
			if len(fleet.ShipIDs) == 0 {
				return nil, nil, fmt.Errorf("stationary combat fleet %d has no concrete ships", fleet.ID)
			}
			entry.side.CombatFleetIDs = append(entry.side.CombatFleetIDs, fleet.ID)
			entry.side.ShipIDs = append(entry.side.ShipIDs, fleet.ShipIDs...)
		case fleet.Role == core.StrategicFleetRoleCivilian && (fleet.SpecialKind == core.StrategicFleetSpecialColonyShip || fleet.SpecialKind == core.StrategicFleetSpecialOutpostShip):
			entry.side.CivilianFleetIDs = append(entry.side.CivilianFleetIDs, fleet.ID)
		}
	}
	for i := range state.Colonies {
		colony := &state.Colonies[i]
		system := systemForPlanetID(state, colony.PlanetID)
		if system == nil {
			return nil, nil, fmt.Errorf("colony %d references planet %d outside the galaxy", colony.ID, colony.PlanetID)
		}
		entry := ensure(system.ID, colony.EmpireID)
		entry.colonyIDs = append(entry.colonyIDs, colony.ID)
	}

	systemIDs := make([]core.ID, 0, len(presence))
	for systemID, byEmpire := range presence {
		systemIDs = append(systemIDs, systemID)
		for _, entry := range byEmpire {
			entry.side.CombatFleetIDs = sortedUniqueIDs(entry.side.CombatFleetIDs)
			entry.side.ShipIDs = sortedUniqueIDs(entry.side.ShipIDs)
			entry.side.CivilianFleetIDs = sortedUniqueIDs(entry.side.CivilianFleetIDs)
			entry.colonyIDs = sortedUniqueIDs(entry.colonyIDs)
		}
	}
	sort.Slice(systemIDs, func(i, j int) bool { return systemIDs[i] < systemIDs[j] })
	return presence, systemIDs, nil
}

func sortedEncounterEmpireIDs(byEmpire map[core.ID]*encounterPresence) []core.ID {
	ids := make([]core.ID, 0, len(byEmpire))
	for id := range byEmpire {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func sortedUniqueIDs(ids []core.ID) []core.ID {
	if len(ids) == 0 {
		return nil
	}
	out := append([]core.ID(nil), ids...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	write := 0
	for _, id := range out {
		if id == 0 || (write > 0 && out[write-1] == id) {
			continue
		}
		out[write] = id
		write++
	}
	return out[:write]
}

func idSet(ids []core.ID) map[core.ID]struct{} {
	set := make(map[core.ID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}

func removeStrategicFleetsByID(state *core.GameState, remove map[core.ID]struct{}) {
	if len(remove) == 0 {
		return
	}
	kept := state.StrategicFleets[:0]
	for _, fleet := range state.StrategicFleets {
		if _, drop := remove[fleet.ID]; drop {
			continue
		}
		kept = append(kept, fleet)
	}
	state.StrategicFleets = kept
}

func (r *EconomyResolver) applyEncounterOutcomes(state *core.GameState, outcomes []EncounterOutcome) ([]DomainEvent, error) {
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	ordered := append([]EncounterOutcome(nil), outcomes...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].BattleID < ordered[j].BattleID })
	seenBattle := make(map[uint64]struct{}, len(ordered))
	seenSystem := make(map[core.ID]struct{}, len(ordered))
	var events []DomainEvent
	for _, outcome := range ordered {
		if outcome.BattleID == 0 {
			return nil, fmt.Errorf("encounter outcome has zero battle id")
		}
		if _, duplicate := seenBattle[outcome.BattleID]; duplicate {
			return nil, fmt.Errorf("duplicate encounter outcome battle id %d", outcome.BattleID)
		}
		seenBattle[outcome.BattleID] = struct{}{}
		if outcome.Encounter.SystemID == 0 {
			return nil, fmt.Errorf("battle %d encounter has zero system id", outcome.BattleID)
		}
		if _, duplicate := seenSystem[outcome.Encounter.SystemID]; duplicate {
			return nil, fmt.Errorf("multiple outcomes for system %d in one encounter wave", outcome.Encounter.SystemID)
		}
		seenSystem[outcome.Encounter.SystemID] = struct{}{}
		if outcome.WinnerSeat != outcome.Encounter.Attacker.SeatID && outcome.WinnerSeat != outcome.Encounter.Defender.SeatID {
			return nil, fmt.Errorf("battle %d winner seat %d is not an encounter side", outcome.BattleID, outcome.WinnerSeat)
		}
		if outcome.Outcome == "" {
			return nil, fmt.Errorf("battle %d outcome must not be empty", outcome.BattleID)
		}
		destroyed := sortedUniqueIDs(outcome.DestroyedShipIDs)
		if len(destroyed) != len(outcome.DestroyedShipIDs) {
			return nil, fmt.Errorf("battle %d destroyed ship ids must be unique and non-zero", outcome.BattleID)
		}
		allowed := idSet(append(append([]core.ID(nil), outcome.Encounter.Attacker.ShipIDs...), outcome.Encounter.Defender.ShipIDs...))
		for _, shipID := range destroyed {
			if _, ok := allowed[shipID]; !ok {
				return nil, fmt.Errorf("battle %d destroyed ship %d is outside the encounter snapshot", outcome.BattleID, shipID)
			}
		}
		if len(destroyed) != 0 {
			removeDestroyedShipsAndEmptyFleets(state, idSet(destroyed))
			event, err := NewDomainEvent("empire.battle_casualties_applied", 0, 0, BattleCasualtiesAppliedEvent{
				BattleID: outcome.BattleID, SystemID: outcome.Encounter.SystemID, DestroyedShipIDs: destroyed,
			})
			if err != nil {
				return nil, err
			}
			events = append(events, event)
		}

		loser := outcome.Encounter.Defender
		if outcome.WinnerSeat == outcome.Encounter.Defender.SeatID {
			loser = outcome.Encounter.Attacker
		}
		retreatEvents, err := r.applyLosingSideRetreat(state, outcome.BattleID, outcome.Encounter.SystemID, loser)
		if err != nil {
			return nil, err
		}
		events = append(events, retreatEvents...)
	}
	return events, nil
}

func removeDestroyedShipsAndEmptyFleets(state *core.GameState, destroyed map[core.ID]struct{}) {
	if len(destroyed) == 0 {
		return
	}
	ships := state.Ships[:0]
	for _, ship := range state.Ships {
		if _, drop := destroyed[ship.ID]; drop {
			continue
		}
		ships = append(ships, ship)
	}
	state.Ships = ships

	fleets := state.StrategicFleets[:0]
	for _, fleet := range state.StrategicFleets {
		if fleet.Role == core.StrategicFleetRoleCombat && fleet.SpecialKind == core.StrategicFleetSpecialNone {
			keptIDs := fleet.ShipIDs[:0]
			for _, shipID := range fleet.ShipIDs {
				if _, drop := destroyed[shipID]; !drop {
					keptIDs = append(keptIDs, shipID)
				}
			}
			fleet.ShipIDs = keptIDs
			if len(fleet.ShipIDs) == 0 {
				continue
			}
		}
		fleets = append(fleets, fleet)
	}
	state.StrategicFleets = fleets
}

func (r *EconomyResolver) applyLosingSideRetreat(state *core.GameState, battleID uint64, sourceSystemID core.ID, loser EncounterSide) ([]DomainEvent, error) {
	source := systemByID(state, sourceSystemID)
	if source == nil {
		return nil, fmt.Errorf("battle %d references unknown source system %d", battleID, sourceSystemID)
	}
	destination := closestOtherOwnedColonySystem(state, loser.EmpireID, sourceSystemID)
	fleetIDs := append(append([]core.ID(nil), loser.CombatFleetIDs...), loser.CivilianFleetIDs...)
	fleetIDs = sortedUniqueIDs(fleetIDs)
	var events []DomainEvent
	for _, fleetID := range fleetIDs {
		_, fleet := strategicFleetByID(state, fleetID)
		if fleet == nil {
			continue
		}
		if fleet.EmpireID != loser.EmpireID {
			return nil, fmt.Errorf("battle %d losing fleet %d belongs to empire %d, want %d", battleID, fleet.ID, fleet.EmpireID, loser.EmpireID)
		}
		if fleet.AtSystemID != sourceSystemID {
			return nil, fmt.Errorf("battle %d losing fleet %d is no longer stationary at system %d", battleID, fleet.ID, sourceSystemID)
		}
		if destination != nil {
			eta, ok := r.establishBattleRetreat(state, fleet, *source, *destination)
			if ok {
				event, err := NewDomainEvent("empire.fleet_retreated_after_battle", 0, 0, FleetRetreatedAfterBattleEvent{
					BattleID: battleID, FleetID: fleet.ID, EmpireID: fleet.EmpireID,
					SourceSystemID: sourceSystemID, DestinationSystemID: destination.ID, RemainingTurns: eta,
				})
				if err != nil {
					return nil, err
				}
				events = append(events, event)
				continue
			}
		}

		destroyedShipIDs := append([]core.ID(nil), fleet.ShipIDs...)
		removeFleetAndOwnedShips(state, fleet.ID)
		reason := "no_retreat_destination"
		if destination != nil {
			reason = "retreat_movement_unavailable"
		}
		event, err := NewDomainEvent("empire.fleet_destroyed_after_battle", 0, 0, FleetDestroyedAfterBattleEvent{
			BattleID: battleID, FleetID: fleetID, EmpireID: loser.EmpireID, SystemID: sourceSystemID,
			DestroyedShipIDs: destroyedShipIDs, Reason: reason,
		})
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (r *EconomyResolver) establishBattleRetreat(state *core.GameState, fleet *core.StrategicFleet, source, destination core.StarSystem) (int, bool) {
	if fleet == nil || source.ID == destination.ID {
		return 0, false
	}
	switch {
	case fleet.Role == core.StrategicFleetRoleCombat && fleet.SpecialKind == core.StrategicFleetSpecialNone:
		empire := empireByID(state, fleet.EmpireID)
		ftlSpeed, _, err := r.combatFleetMovementProfile(state, *fleet, empire)
		if err != nil {
			return 0, false
		}
		eta := strategicTravelETA(source, destination, ftlSpeed)
		if eta < 1 {
			return 0, false
		}
		fleet.AtSystemID = 0
		fleet.DestinationSystemID = destination.ID
		fleet.RemainingTurns = eta
		fleet.FTLSpeed = 0
		return eta, true
	case fleet.Role == core.StrategicFleetRoleCivilian && (fleet.SpecialKind == core.StrategicFleetSpecialColonyShip || fleet.SpecialKind == core.StrategicFleetSpecialOutpostShip):
		eta := strategicTravelETA(source, destination, fleet.FTLSpeed)
		if eta < 1 {
			return 0, false
		}
		fleet.AtSystemID = 0
		fleet.DestinationSystemID = destination.ID
		fleet.RemainingTurns = eta
		return eta, true
	default:
		return 0, false
	}
}

func closestOtherOwnedColonySystem(state *core.GameState, empireID, sourceSystemID core.ID) *core.StarSystem {
	source := systemByID(state, sourceSystemID)
	if source == nil {
		return nil
	}
	seen := make(map[core.ID]struct{})
	var best *core.StarSystem
	var bestDistance int64
	for i := range state.Colonies {
		colony := &state.Colonies[i]
		if colony.EmpireID != empireID {
			continue
		}
		system := systemForPlanetID(state, colony.PlanetID)
		if system == nil || system.ID == sourceSystemID {
			continue
		}
		if _, duplicate := seen[system.ID]; duplicate {
			continue
		}
		seen[system.ID] = struct{}{}
		dx := int64(source.X - system.X)
		dy := int64(source.Y - system.Y)
		distance := dx*dx + dy*dy
		if best == nil || distance < bestDistance || (distance == bestDistance && system.ID < best.ID) {
			best = system
			bestDistance = distance
		}
	}
	return best
}

func removeFleetAndOwnedShips(state *core.GameState, fleetID core.ID) {
	_, fleet := strategicFleetByID(state, fleetID)
	if fleet == nil {
		return
	}
	shipIDs := idSet(fleet.ShipIDs)
	if len(shipIDs) != 0 {
		ships := state.Ships[:0]
		for _, ship := range state.Ships {
			if _, drop := shipIDs[ship.ID]; !drop {
				ships = append(ships, ship)
			}
		}
		state.Ships = ships
	}
	removeStrategicFleetsByID(state, map[core.ID]struct{}{fleetID: {}})
}
