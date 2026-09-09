package session

import (
	"encoding/json"
	"fmt"
	"sort"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

const recentResolutionSummaryLimit = 12

type ResolutionSummaryKind string

const (
	ResolutionSummaryResearchBreakthrough ResolutionSummaryKind = "research_breakthrough"
	ResolutionSummaryTechnologyGranted    ResolutionSummaryKind = "technology_granted"
	ResolutionSummaryBattleCompleted      ResolutionSummaryKind = "battle_completed"
	ResolutionSummaryInvasionResolved     ResolutionSummaryKind = "invasion_resolved"
	ResolutionSummaryInvasionDeclined     ResolutionSummaryKind = "invasion_declined"
	ResolutionSummaryEmpireEliminated     ResolutionSummaryKind = "empire_eliminated"
	ResolutionSummaryGameCompleted        ResolutionSummaryKind = "game_completed"
)

type ResearchResolutionSummary struct {
	EmpireID        core.ID                    `json:"empire_id"`
	TechFieldID     int                        `json:"tech_field_id"`
	SelectionMode   core.ResearchSelectionMode `json:"selection_mode"`
	TechnologyIDs   []int                      `json:"technology_ids,omitempty"`
	TechnologyKeys  []string                   `json:"technology_keys,omitempty"`
	CompletedLevels int                        `json:"completed_levels,omitempty"`
	ResearchLevel   int                        `json:"research_level,omitempty"`
}

type TechnologyGrantResolutionSummary struct {
	EmpireID          core.ID                    `json:"empire_id"`
	TechnologyID      int                        `json:"technology_id"`
	TechnologyKey     string                     `json:"technology_key"`
	TechnologyNameKey string                     `json:"technology_name_key"`
	TechFieldID       int                        `json:"tech_field_id"`
	SourceKind        game.TechnologyGrantSource `json:"source_kind"`
}

type BattleResolutionSummary struct {
	BattleID          uint64          `json:"battle_id"`
	SystemID          core.ID         `json:"system_id,omitempty"`
	AttackerEmpireID  core.ID         `json:"attacker_empire_id,omitempty"`
	DefenderEmpireID  core.ID         `json:"defender_empire_id,omitempty"`
	WinnerSeatID      protocol.SeatID `json:"winner_seat_id"`
	WinnerEmpireID    core.ID         `json:"winner_empire_id,omitempty"`
	Outcome           string          `json:"outcome"`
	DestroyedShipIDs  []core.ID       `json:"destroyed_ship_ids,omitempty"`
	SurvivingShipIDs  []core.ID       `json:"surviving_ship_ids,omitempty"`
	DefenderColonyIDs []core.ID       `json:"defender_colony_ids,omitempty"`
}

type InvasionResolutionSummary struct {
	SystemID                   core.ID   `json:"system_id"`
	ColonyID                   core.ID   `json:"colony_id"`
	AttackerEmpireID           core.ID   `json:"attacker_empire_id"`
	DefenderEmpireID           core.ID   `json:"defender_empire_id"`
	Outcome                    string    `json:"outcome"`
	SelectedTransportFleetIDs  []core.ID `json:"selected_transport_fleet_ids,omitempty"`
	InitialAttackerInfantry    int       `json:"initial_attacker_infantry,omitempty"`
	InitialDefenderInfantry    int       `json:"initial_defender_infantry,omitempty"`
	InitialDefenderMilitia     int       `json:"initial_defender_militia,omitempty"`
	SurvivingAttackerInfantry  int       `json:"surviving_attacker_infantry,omitempty"`
	SurvivingDefenderInfantry  int       `json:"surviving_defender_infantry,omitempty"`
	SurvivingDefenderMilitia   int       `json:"surviving_defender_militia,omitempty"`
	ConsumedTransportFleetIDs  []core.ID `json:"consumed_transport_fleet_ids,omitempty"`
	SurvivingTransportFleetIDs []core.ID `json:"surviving_transport_fleet_ids,omitempty"`
}

type ResolutionSummary struct {
	ID                 string                            `json:"id"`
	EventSequence      uint64                            `json:"event_sequence"`
	Turn               uint64                            `json:"turn"`
	Revision           uint64                            `json:"revision"`
	Kind               ResolutionSummaryKind             `json:"kind"`
	Research           *ResearchResolutionSummary        `json:"research,omitempty"`
	TechnologyGrant    *TechnologyGrantResolutionSummary `json:"technology_grant,omitempty"`
	Battle             *BattleResolutionSummary          `json:"battle,omitempty"`
	Invasion           *InvasionResolutionSummary        `json:"invasion,omitempty"`
	EliminatedEmpireID core.ID                           `json:"eliminated_empire_id,omitempty"`
	Result             *Result                           `json:"result,omitempty"`
}

func (s *GameSession) recentResolutionSummariesLocked(seatID protocol.SeatID, empireID core.ID) []ResolutionSummary {
	if recentResolutionSummaryLimit <= 0 || len(s.events) == 0 {
		return nil
	}
	minimumTurn := uint64(0)
	if s.state != nil && s.state.Turn > 1 {
		minimumTurn = s.state.Turn - 1
	}
	out := make([]ResolutionSummary, 0, recentResolutionSummaryLimit)
	for i := len(s.events) - 1; i >= 0 && len(out) < recentResolutionSummaryLimit; i-- {
		event := s.events[i]
		if event.Turn < minimumTurn && event.Kind != "game_completed" {
			continue
		}
		summary, ok := resolutionSummaryForEvent(event, seatID, empireID)
		if !ok {
			continue
		}
		out = append(out, summary)
	}
	for left, right := 0, len(out)-1; left < right; left, right = left+1, right-1 {
		out[left], out[right] = out[right], out[left]
	}
	return out
}

func resolutionSummaryForEvent(event protocol.DomainEvent, seatID protocol.SeatID, empireID core.ID) (ResolutionSummary, bool) {
	base := ResolutionSummary{
		ID:            fmt.Sprintf("event-%d", event.Sequence),
		EventSequence: event.Sequence,
		Turn:          event.Turn,
		Revision:      event.Revision,
	}
	switch event.Kind {
	case "empire.research_completed":
		var data game.ResearchCompletedEvent
		if json.Unmarshal(event.Data, &data) != nil || data.EmpireID != empireID {
			return ResolutionSummary{}, false
		}
		base.Kind = ResolutionSummaryResearchBreakthrough
		base.Research = &ResearchResolutionSummary{
			EmpireID:        data.EmpireID,
			TechFieldID:     data.TechFieldID,
			SelectionMode:   data.SelectionMode,
			TechnologyIDs:   append([]int(nil), data.TechnologyIDs...),
			TechnologyKeys:  append([]string(nil), data.TechnologyKeys...),
			CompletedLevels: data.CompletedLevels,
			ResearchLevel:   data.ResearchLevel,
		}
		return base, true
	case game.EventTechnologyGranted:
		var data game.TechnologyGrantedEvent
		if json.Unmarshal(event.Data, &data) != nil || data.EmpireID != empireID {
			return ResolutionSummary{}, false
		}
		base.Kind = ResolutionSummaryTechnologyGranted
		base.TechnologyGrant = &TechnologyGrantResolutionSummary{
			EmpireID:          data.EmpireID,
			TechnologyID:      data.TechnologyID,
			TechnologyKey:     data.TechnologyKey,
			TechnologyNameKey: data.TechnologyNameKey,
			TechFieldID:       data.TechFieldID,
			SourceKind:        data.SourceKind,
		}
		return base, true
	case "battle_completed":
		var view battle.View
		if json.Unmarshal(event.Data, &view) != nil || view.Result == nil || !seatParticipated(view.Spec.Participants, seatID) {
			return ResolutionSummary{}, false
		}
		winnerEmpireID := core.ID(0)
		if view.Result.WinnerSeat == view.Spec.Attacker.SeatID {
			winnerEmpireID = view.Spec.Attacker.EmpireID
		} else if view.Result.WinnerSeat == view.Spec.Defender.SeatID {
			winnerEmpireID = view.Spec.Defender.EmpireID
		}
		base.Kind = ResolutionSummaryBattleCompleted
		base.Battle = &BattleResolutionSummary{
			BattleID:          view.Spec.ID,
			SystemID:          view.Spec.SystemID,
			AttackerEmpireID:  view.Spec.Attacker.EmpireID,
			DefenderEmpireID:  view.Spec.Defender.EmpireID,
			WinnerSeatID:      view.Result.WinnerSeat,
			WinnerEmpireID:    winnerEmpireID,
			Outcome:           view.Result.Outcome,
			DestroyedShipIDs:  append([]core.ID(nil), view.Result.DestroyedShipIDs...),
			SurvivingShipIDs:  survivingBattleShipIDs(view),
			DefenderColonyIDs: append([]core.ID(nil), view.Spec.DefenderColonyIDs...),
		}
		return base, true
	case "empire.invasion_resolved":
		var data game.InvasionResolvedEvent
		if json.Unmarshal(event.Data, &data) != nil || (data.AttackerEmpireID != empireID && data.DefenderEmpireID != empireID) {
			return ResolutionSummary{}, false
		}
		outcome := "repelled"
		if data.Captured {
			outcome = "captured"
		}
		base.Kind = ResolutionSummaryInvasionResolved
		base.Invasion = &InvasionResolutionSummary{
			SystemID:                   data.SystemID,
			ColonyID:                   data.ColonyID,
			AttackerEmpireID:           data.AttackerEmpireID,
			DefenderEmpireID:           data.DefenderEmpireID,
			Outcome:                    outcome,
			SelectedTransportFleetIDs:  append([]core.ID(nil), data.SelectedTransportFleetIDs...),
			InitialAttackerInfantry:    data.InitialAttackerInfantry,
			InitialDefenderInfantry:    data.InitialDefenderInfantry,
			InitialDefenderMilitia:     data.InitialDefenderMilitia,
			SurvivingAttackerInfantry:  data.SurvivingAttackerInfantry,
			SurvivingDefenderInfantry:  data.SurvivingDefenderInfantry,
			SurvivingDefenderMilitia:   data.SurvivingDefenderMilitia,
			ConsumedTransportFleetIDs:  append([]core.ID(nil), data.ConsumedTransportFleetIDs...),
			SurvivingTransportFleetIDs: append([]core.ID(nil), data.SurvivingTransportFleetIDs...),
		}
		return base, true
	case "empire.invasion_declined":
		var data game.InvasionDeclinedEvent
		if json.Unmarshal(event.Data, &data) != nil || (data.AttackerEmpireID != empireID && data.DefenderEmpireID != empireID) {
			return ResolutionSummary{}, false
		}
		base.Kind = ResolutionSummaryInvasionDeclined
		base.Invasion = &InvasionResolutionSummary{
			SystemID:         data.SystemID,
			ColonyID:         data.ColonyID,
			AttackerEmpireID: data.AttackerEmpireID,
			DefenderEmpireID: data.DefenderEmpireID,
			Outcome:          "declined",
		}
		return base, true
	case "empire_eliminated":
		var data EmpireEliminatedEvent
		if json.Unmarshal(event.Data, &data) != nil || data.EmpireID == 0 {
			return ResolutionSummary{}, false
		}
		base.Kind = ResolutionSummaryEmpireEliminated
		base.EliminatedEmpireID = data.EmpireID
		return base, true
	case "game_completed":
		var data Result
		if json.Unmarshal(event.Data, &data) != nil || data.Kind == "" {
			return ResolutionSummary{}, false
		}
		base.Kind = ResolutionSummaryGameCompleted
		base.Result = cloneResult(&data)
		return base, true
	default:
		return ResolutionSummary{}, false
	}
}

func seatParticipated(participants []protocol.SeatID, seatID protocol.SeatID) bool {
	for _, participant := range participants {
		if participant == seatID {
			return true
		}
	}
	return false
}

func survivingBattleShipIDs(view battle.View) []core.ID {
	destroyed := make(map[core.ID]struct{}, len(view.Result.DestroyedShipIDs))
	for _, shipID := range view.Result.DestroyedShipIDs {
		destroyed[shipID] = struct{}{}
	}
	all := make([]core.ID, 0, len(view.Spec.Attacker.ShipIDs)+len(view.Spec.Defender.ShipIDs))
	all = append(all, view.Spec.Attacker.ShipIDs...)
	all = append(all, view.Spec.Defender.ShipIDs...)
	surviving := all[:0]
	for _, shipID := range all {
		if _, gone := destroyed[shipID]; !gone {
			surviving = append(surviving, shipID)
		}
	}
	sort.Slice(surviving, func(i, j int) bool { return surviving[i] < surviving[j] })
	return surviving
}
