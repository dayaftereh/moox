package session

import (
	"encoding/json"
	"fmt"
	"sort"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func (s *GameSession) validateInvasionOpportunityLocked(state *core.GameState, opportunity *game.InvasionOpportunity) error {
	if opportunity == nil {
		return nil
	}
	if state == nil {
		return fmt.Errorf("invasion opportunity state is nil")
	}
	if opportunity.SystemID == 0 || opportunity.ColonyID == 0 || opportunity.AttackerEmpireID == 0 || opportunity.DefenderEmpireID == 0 || opportunity.AttackerSeatID == 0 {
		return fmt.Errorf("invasion opportunity has zero identity field")
	}
	if opportunity.AttackerEmpireID == opportunity.DefenderEmpireID {
		return fmt.Errorf("invasion opportunity attacker and defender must differ")
	}
	seatIndex := s.seatIndexLocked(opportunity.AttackerSeatID)
	if seatIndex < 0 || s.seats[seatIndex].seat.EmpireID != opportunity.AttackerEmpireID {
		return fmt.Errorf("invasion opportunity seat %d does not authorize empire %d", opportunity.AttackerSeatID, opportunity.AttackerEmpireID)
	}
	var colony *core.Colony
	for i := range state.Colonies {
		if state.Colonies[i].ID == opportunity.ColonyID {
			colony = &state.Colonies[i]
			break
		}
	}
	if colony == nil || colony.EmpireID != opportunity.DefenderEmpireID {
		return fmt.Errorf("invasion opportunity references unavailable defender colony %d", opportunity.ColonyID)
	}
	if len(opportunity.EligibleTransportFleetIDs) == 0 {
		return fmt.Errorf("invasion opportunity has no eligible transports")
	}
	for i, fleetID := range opportunity.EligibleTransportFleetIDs {
		if i > 0 && fleetID <= opportunity.EligibleTransportFleetIDs[i-1] {
			return fmt.Errorf("invasion opportunity transport IDs must be sorted unique")
		}
		found := false
		for fi := range state.StrategicFleets {
			fleet := &state.StrategicFleets[fi]
			if fleet.ID != fleetID {
				continue
			}
			found = fleet.EmpireID == opportunity.AttackerEmpireID && fleet.Role == core.StrategicFleetRoleCivilian && fleet.SpecialKind == core.StrategicFleetSpecialTroopTransport && fleet.AtSystemID == opportunity.SystemID && fleet.DestinationSystemID == 0 && fleet.RemainingTurns == 0
			break
		}
		if !found {
			return fmt.Errorf("invasion opportunity transport %d is not an eligible stationary attacker transport", fleetID)
		}
	}
	return nil
}

func validateInvasionImmediateEvents(events []game.DomainEvent, seatID protocol.SeatID, command protocol.Command) error {
	if len(events) == 0 {
		return fmt.Errorf("invasion command returned no events")
	}
	for _, event := range events {
		if event.Kind == "" {
			return fmt.Errorf("invasion command returned event with empty kind")
		}
		if len(event.Data) > 0 && !json.Valid(event.Data) {
			return fmt.Errorf("invasion event %q has invalid JSON data", event.Kind)
		}
		if event.SeatID != seatID || event.CommandSequence != command.Sequence {
			return fmt.Errorf("invasion event %q authority=%d/%d want=%d/%d", event.Kind, event.SeatID, event.CommandSequence, seatID, command.Sequence)
		}
	}
	return nil
}

func appendHandledInvasion(keys []game.InvasionHandledKey, key game.InvasionHandledKey) []game.InvasionHandledKey {
	out := append([]game.InvasionHandledKey(nil), keys...)
	out = append(out, key)
	sort.Slice(out, func(i, j int) bool {
		if out[i].ColonyID != out[j].ColonyID {
			return out[i].ColonyID < out[j].ColonyID
		}
		return out[i].AttackerEmpireID < out[j].AttackerEmpireID
	})
	unique := out[:0]
	for _, current := range out {
		if len(unique) != 0 && unique[len(unique)-1] == current {
			continue
		}
		unique = append(unique, current)
	}
	return unique
}

func (s *GameSession) appendImmediateDomainEventsLocked(events []game.DomainEvent) {
	for _, event := range events {
		s.events = append(s.events, protocol.DomainEvent{
			SchemaVersion:   protocol.EventSchemaVersion,
			Sequence:        s.nextEventSequence,
			Turn:            s.state.Turn,
			Revision:        s.revision,
			Scope:           protocol.EventScopeStrategic,
			Kind:            event.Kind,
			SeatID:          event.SeatID,
			CommandSequence: event.CommandSequence,
			Data:            append(json.RawMessage(nil), event.Data...),
		})
		s.nextEventSequence++
	}
}

func (s *GameSession) ResolveInvasionCommand(seatID protocol.SeatID, baseRevision uint64, command protocol.Command) error {
	if err := command.Validate(1); err != nil {
		return fmt.Errorf("invalid invasion command: %w", err)
	}
	if !game.IsInvasionCommand(command.Kind) {
		return fmt.Errorf("command kind %q is not an invasion command", command.Kind)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhaseInvasionDecisions {
		return fmt.Errorf("cannot resolve invasion in phase %q", s.phase)
	}
	if baseRevision != s.revision {
		return fmt.Errorf("immediate command base revision %d, expected %d", baseRevision, s.revision)
	}
	if s.invasion == nil {
		return fmt.Errorf("invasion decision phase has no pending opportunity")
	}
	if s.seatIndexLocked(seatID) < 0 {
		return fmt.Errorf("unknown seat %d", seatID)
	}
	resolver, ok := s.encounterResolver.(game.InvasionResolver)
	if !ok || s.encounterResolver == nil {
		return fmt.Errorf("invasion continuation resolver is not configured")
	}

	stateInput, err := cloneState(s.state)
	if err != nil {
		return err
	}
	ctx := cloneResolveContext(s.encounterContext)
	opportunity := game.CloneInvasionOpportunity(s.invasion)
	immediateEvents, err := resolver.ResolveInvasionCommand(ctx, stateInput, *opportunity, seatID, command)
	if err != nil {
		return fmt.Errorf("resolve invasion command: %w", err)
	}
	if err := validateInvasionImmediateEvents(immediateEvents, seatID, command); err != nil {
		return err
	}
	candidateHandled := appendHandledInvasion(ctx.HandledInvasions, game.InvasionHandledKey{ColonyID: opportunity.ColonyID, AttackerEmpireID: opportunity.AttackerEmpireID})
	ctx.HandledInvasions = append([]game.InvasionHandledKey(nil), candidateHandled...)
	continuation, err := resolver.ResumeAfterInvasion(ctx, stateInput)
	if err != nil {
		return fmt.Errorf("resume strategic invasion: %w", err)
	}
	if err := s.validateStrategicResolutionLocked(continuation); err != nil {
		return err
	}
	committed, err := cloneState(continuation.State)
	if err != nil {
		return err
	}
	nextSpecs := make([]EncounterSpec, len(continuation.Encounters))
	for i := range continuation.Encounters {
		nextSpecs[i] = encounterSpecFromGame(continuation.Encounters[i])
	}
	preparedNext, err := s.prepareEncountersLocked(committed, nextSpecs)
	if err != nil {
		return err
	}

	// All state, RNG, handled-key and next-boundary work above is candidate-only.
	s.state = committed
	s.revision++
	s.appendImmediateDomainEventsLocked(immediateEvents)
	s.appendResolvedEventsLocked(continuation.Events)
	s.handledInvasions = candidateHandled
	s.encounterContext = cloneResolveContext(ctx)

	if continuation.Invasion != nil {
		s.battles = nil
		s.invasion = game.CloneInvasionOpportunity(continuation.Invasion)
		// Remain in PhaseInvasionDecisions for the next independent opportunity.
		return nil
	}
	if len(preparedNext) != 0 {
		s.invasion = nil
		s.battles = preparedNext
		s.nextBattleID += uint64(len(preparedNext))
		s.phase = PhaseEncounters
		for _, next := range s.battles {
			s.appendEventLocked(protocol.EventScopeBattle, "battle_created", 0, next.View().Spec)
		}
		s.appendEventLocked(protocol.EventScopeSession, "phase_changed", 0, map[string]any{"phase": s.phase})
		return nil
	}

	s.invasion = nil
	s.battles = nil
	s.encounterResolver = nil
	s.encounterContext = game.ResolveContext{}
	s.phase = PhasePostResolution
	s.appendEventLocked(protocol.EventScopeSession, "phase_changed", 0, map[string]any{"phase": s.phase})
	return nil
}
