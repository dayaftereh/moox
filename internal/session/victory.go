package session

import (
	"fmt"
	"sort"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

type EmpireEliminatedEvent struct {
	EmpireID core.ID `json:"empire_id"`
}

type conquestFinalization struct {
	state           *core.GameState
	newlyEliminated []core.ID
	allEliminated   []core.ID
	winnerEmpireID  core.ID
	winnerSeatID    protocol.SeatID
	complete        bool
}

func cloneResult(result *Result) *Result {
	if result == nil {
		return nil
	}
	clone := *result
	clone.EliminatedEmpireIDs = append([]core.ID(nil), result.EliminatedEmpireIDs...)
	return &clone
}

func containsConquestEvent(events []game.DomainEvent) bool {
	for _, event := range events {
		if event.Kind == "empire.colony_conquered" {
			return true
		}
	}
	return false
}

func (s *GameSession) prepareConquestFinalizationLocked(state *core.GameState, pending bool) (*conquestFinalization, error) {
	if !pending {
		return nil, nil
	}
	participants := s.participatingEmpireIDsLocked()
	if len(participants) < 2 {
		return &conquestFinalization{state: state}, nil
	}
	candidate, err := cloneState(state)
	if err != nil {
		return nil, err
	}
	owners := make(map[core.ID]struct{}, len(candidate.Colonies))
	for _, colony := range candidate.Colonies {
		owners[colony.EmpireID] = struct{}{}
	}

	plan := &conquestFinalization{
		state:         candidate,
		allEliminated: append([]core.ID(nil), s.eliminatedEmpires...),
	}
	for _, empireID := range participants {
		if _, alive := owners[empireID]; alive {
			continue
		}
		if containsCoreID(plan.allEliminated, empireID) {
			continue
		}
		plan.newlyEliminated = append(plan.newlyEliminated, empireID)
		plan.allEliminated = append(plan.allEliminated, empireID)
	}
	sort.Slice(plan.newlyEliminated, func(i, j int) bool { return plan.newlyEliminated[i] < plan.newlyEliminated[j] })
	sort.Slice(plan.allEliminated, func(i, j int) bool { return plan.allEliminated[i] < plan.allEliminated[j] })

	if len(plan.newlyEliminated) != 0 {
		cleanupEliminatedEmpires(candidate, plan.newlyEliminated)
		if err := candidate.Validate(); err != nil {
			return nil, fmt.Errorf("validate post-elimination state: %w", err)
		}
	}

	alive := make([]core.ID, 0, len(participants))
	for _, empireID := range participants {
		if _, ownsColony := owners[empireID]; ownsColony {
			alive = append(alive, empireID)
		}
	}
	if len(alive) == 1 {
		winnerSeatID, ok := s.seatForEmpireLocked(alive[0])
		if !ok {
			return nil, fmt.Errorf("conquest winner empire %d has no controlling seat", alive[0])
		}
		plan.complete = true
		plan.winnerEmpireID = alive[0]
		plan.winnerSeatID = winnerSeatID
	}
	return plan, nil
}

func (s *GameSession) commitPostContinuationLocked(plan *conquestFinalization) {
	if s.pendingEliminationCheck {
		s.pendingEliminationCheck = false
	}
	if plan != nil && (len(plan.newlyEliminated) != 0 || plan.complete) {
		s.state = plan.state
		s.eliminatedEmpires = append([]core.ID(nil), plan.allEliminated...)
		// Elimination/game completion is its own authoritative boundary after the
		// invasion/encounter state commit, so it owns one additional revision.
		s.revision++
		for _, empireID := range plan.newlyEliminated {
			s.appendEventLocked(protocol.EventScopeSession, "empire_eliminated", 0, EmpireEliminatedEvent{EmpireID: empireID})
		}
	}
	if plan != nil && plan.complete {
		result := &Result{
			Kind:                ResultConquest,
			WinnerEmpireID:      plan.winnerEmpireID,
			WinnerSeatID:        plan.winnerSeatID,
			EliminatedEmpireIDs: append([]core.ID(nil), s.eliminatedEmpires...),
			CompletedTurn:       s.state.Turn,
			CompletedRevision:   s.revision,
		}
		s.result = result
		s.battles = nil
		s.invasion = nil
		s.handledInvasions = nil
		s.encounterResolver = nil
		s.encounterContext = game.ResolveContext{}
		s.phase = PhaseCompleted
		s.appendEventLocked(protocol.EventScopeSession, "game_completed", 0, result)
		s.appendEventLocked(protocol.EventScopeSession, "phase_changed", 0, map[string]any{"phase": s.phase})
		return
	}
	s.phase = PhasePostResolution
	s.appendEventLocked(protocol.EventScopeSession, "phase_changed", 0, map[string]any{"phase": s.phase})
}

func (s *GameSession) participatingEmpireIDsLocked() []core.ID {
	ids := make([]core.ID, 0, len(s.seats))
	for _, seat := range s.seats {
		ids = append(ids, seat.seat.EmpireID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (s *GameSession) seatForEmpireLocked(empireID core.ID) (protocol.SeatID, bool) {
	for _, seat := range s.seats {
		if seat.seat.EmpireID == empireID {
			return seat.seat.ID, true
		}
	}
	return 0, false
}

func containsCoreID(ids []core.ID, id core.ID) bool {
	index := sort.Search(len(ids), func(i int) bool { return ids[i] >= id })
	return index < len(ids) && ids[index] == id
}

func cleanupEliminatedEmpires(state *core.GameState, eliminated []core.ID) {
	if state == nil || len(eliminated) == 0 {
		return
	}
	set := make(map[core.ID]struct{}, len(eliminated))
	for _, empireID := range eliminated {
		set[empireID] = struct{}{}
	}

	removedOutposts := make(map[core.ID]struct{})
	outposts := state.Outposts[:0]
	for _, outpost := range state.Outposts {
		if _, drop := set[outpost.EmpireID]; drop {
			removedOutposts[outpost.ID] = struct{}{}
			continue
		}
		outposts = append(outposts, outpost)
	}
	state.Outposts = outposts
	for si := range state.Galaxy.Systems {
		system := &state.Galaxy.Systems[si]
		blockades := system.BlockadedEmpireIDs[:0]
		for _, empireID := range system.BlockadedEmpireIDs {
			if _, drop := set[empireID]; !drop {
				blockades = append(blockades, empireID)
			}
		}
		system.BlockadedEmpireIDs = blockades
		for pi := range system.Planets {
			planet := &system.Planets[pi]
			if _, removed := removedOutposts[planet.OutpostID]; removed {
				planet.OutpostID = 0
			}
		}
	}

	ships := state.Ships[:0]
	for _, ship := range state.Ships {
		if _, drop := set[ship.EmpireID]; !drop {
			ships = append(ships, ship)
		}
	}
	state.Ships = ships
	fleets := state.StrategicFleets[:0]
	for _, fleet := range state.StrategicFleets {
		if _, drop := set[fleet.EmpireID]; !drop {
			fleets = append(fleets, fleet)
		}
	}
	state.StrategicFleets = fleets

	transfers := state.PopulationTransfers[:0]
	for _, transfer := range state.PopulationTransfers {
		if _, drop := set[transfer.EmpireID]; !drop {
			transfers = append(transfers, transfer)
		}
	}
	state.PopulationTransfers = transfers
	relations := state.DiplomaticRelations[:0]
	for _, relation := range state.DiplomaticRelations {
		_, fromEliminated := set[relation.FromEmpireID]
		_, toEliminated := set[relation.ToEmpireID]
		if !fromEliminated && !toEliminated {
			relations = append(relations, relation)
		}
	}
	state.DiplomaticRelations = relations
	offers := state.DiplomaticPeaceOffers[:0]
	for _, offer := range state.DiplomaticPeaceOffers {
		_, fromEliminated := set[offer.FromEmpireID]
		_, toEliminated := set[offer.ToEmpireID]
		if !fromEliminated && !toEliminated {
			offers = append(offers, offer)
		}
	}
	state.DiplomaticPeaceOffers = offers

	for i := range state.Empires {
		if _, eliminated := set[state.Empires[i].ID]; eliminated {
			state.Empires[i].Capital = 0
			state.Empires[i].Freighters = 0
		}
	}
}
