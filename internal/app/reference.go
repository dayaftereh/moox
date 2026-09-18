package app

import (
	"encoding/json"
	"errors"
	"fmt"

	"moox/internal/core"
	"moox/internal/protocol"
	"moox/internal/session"
)

const (
	ReferenceMaxTurnsPerRequest   = 25
	ReferenceMaxConstructionTurns = 512
)

var (
	ErrReferenceForbidden            = errors.New("reference control forbidden")
	ErrReferenceNotReady             = errors.New("reference control not ready")
	ErrReferenceNoActiveConstruction = errors.New("reference no active construction")
	ErrReferenceStaleRevision        = errors.New("reference stale revision")
)

type ReferenceGameInfo struct {
	ScenarioID           string          `json:"scenario_id"`
	ProfileID            string          `json:"profile_id"`
	ControlSeatID        protocol.SeatID `json:"control_seat_id"`
	MaxTurnsPerRequest   int             `json:"max_turns_per_request"`
	MaxConstructionTurns int             `json:"max_construction_turns"`
}

type ReferenceAdvanceMode string

const (
	ReferenceAdvanceTurns             ReferenceAdvanceMode = "turns"
	ReferenceAdvanceUntilConstruction ReferenceAdvanceMode = "until_construction_complete"
)

type ReferenceAdvanceRequest struct {
	SchemaVersion int                  `json:"schema_version"`
	SeatID        protocol.SeatID      `json:"seat_id"`
	BaseRevision  uint64               `json:"base_revision"`
	Mode          ReferenceAdvanceMode `json:"mode"`
	Turns         int                  `json:"turns,omitempty"`
	ColonyID      core.ID              `json:"colony_id,omitempty"`
}

type ReferenceAdvanceResult struct {
	SchemaVersion  int           `json:"schema_version"`
	GameID         string        `json:"game_id"`
	StartTurn      uint64        `json:"start_turn"`
	EndTurn        uint64        `json:"end_turn"`
	TurnsAdvanced  uint64        `json:"turns_advanced"`
	StopReason     string        `json:"stop_reason"`
	FinalPhase     session.Phase `json:"final_phase"`
	GameRevision   uint64        `json:"game_revision"`
	ChangeSequence uint64        `json:"change_sequence"`
}

type referenceConstructionSignature struct {
	Kind               core.ConstructionProjectKind
	ProjectID          string
	ShipDesignID       core.ID
	ShipDesignRevision uint32
}

func normalizeReferenceInfo(info *ReferenceGameInfo) (*ReferenceGameInfo, error) {
	if info == nil {
		return nil, nil
	}
	if info.ScenarioID == "" || info.ProfileID == "" || info.ControlSeatID == 0 {
		return nil, fmt.Errorf("reference registration requires scenario_id, profile_id and control_seat_id")
	}
	clone := *info
	if clone.MaxTurnsPerRequest == 0 {
		clone.MaxTurnsPerRequest = ReferenceMaxTurnsPerRequest
	}
	if clone.MaxConstructionTurns == 0 {
		clone.MaxConstructionTurns = ReferenceMaxConstructionTurns
	}
	if clone.MaxTurnsPerRequest != ReferenceMaxTurnsPerRequest || clone.MaxConstructionTurns != ReferenceMaxConstructionTurns {
		return nil, fmt.Errorf("reference registration bounds differ from frozen Slice 17.0 contract")
	}
	return &clone, nil
}

func cloneReferenceInfo(info *ReferenceGameInfo) *ReferenceGameInfo {
	if info == nil {
		return nil
	}
	clone := *info
	return &clone
}

func (h *Host) AdvanceReference(gameID string, request ReferenceAdvanceRequest) (ReferenceAdvanceResult, error) {
	hosted, err := h.lookup(gameID)
	if err != nil {
		return ReferenceAdvanceResult{}, err
	}
	hosted.mu.Lock()
	defer hosted.mu.Unlock()

	info := hosted.reference
	if info == nil || request.SeatID != info.ControlSeatID {
		return ReferenceAdvanceResult{}, fmt.Errorf("%w: game %q seat %d", ErrReferenceForbidden, gameID, request.SeatID)
	}
	status := hosted.session.Status()
	if request.BaseRevision != status.Revision {
		return ReferenceAdvanceResult{}, fmt.Errorf("%w: base revision %d, current %d", ErrReferenceStaleRevision, request.BaseRevision, status.Revision)
	}
	if request.SchemaVersion != SchemaVersion {
		return ReferenceAdvanceResult{}, fmt.Errorf("%w: unsupported schema_version %d", ErrReferenceNotReady, request.SchemaVersion)
	}
	if status.Phase != session.PhasePlanning || status.Result != nil {
		return ReferenceAdvanceResult{}, fmt.Errorf("%w: phase %q", ErrReferenceNotReady, status.Phase)
	}
	controlView, err := hosted.session.PlayerView(info.ControlSeatID)
	if err != nil {
		return ReferenceAdvanceResult{}, err
	}
	if controlView.Seat.Submitted {
		return ReferenceAdvanceResult{}, fmt.Errorf("%w: control seat already submitted turn %d", ErrReferenceNotReady, status.Turn)
	}
	if draft, ok := hosted.planningDrafts[info.ControlSeatID]; ok && len(draft.Orders) != 0 {
		return ReferenceAdvanceResult{}, fmt.Errorf("%w: control seat has a non-empty planning draft", ErrReferenceNotReady)
	}

	limit := 0
	var construction *referenceConstructionSignature
	switch request.Mode {
	case ReferenceAdvanceTurns:
		if request.Turns < 1 || request.Turns > info.MaxTurnsPerRequest {
			return ReferenceAdvanceResult{}, fmt.Errorf("%w: turns must be within [1,%d]", ErrReferenceNotReady, info.MaxTurnsPerRequest)
		}
		if request.ColonyID != 0 {
			return ReferenceAdvanceResult{}, fmt.Errorf("%w: colony_id is invalid for turns mode", ErrReferenceNotReady)
		}
		limit = request.Turns
	case ReferenceAdvanceUntilConstruction:
		if request.Turns != 0 || request.ColonyID == 0 {
			return ReferenceAdvanceResult{}, fmt.Errorf("%w: construction mode requires colony_id and no turns", ErrReferenceNotReady)
		}
		signature, err := referenceConstructionForView(controlView, request.ColonyID)
		if err != nil {
			return ReferenceAdvanceResult{}, err
		}
		construction = &signature
		limit = info.MaxConstructionTurns
	default:
		return ReferenceAdvanceResult{}, fmt.Errorf("%w: unsupported mode %q", ErrReferenceNotReady, request.Mode)
	}

	start := status
	eventCursor, err := hosted.referenceEventCount()
	if err != nil {
		return ReferenceAdvanceResult{}, err
	}
	stopReason := "hard_limit_reached"

	for step := 0; step < limit; step++ {
		before := hosted.session.Status()
		if before.Phase == session.PhaseCompleted || before.Result != nil {
			stopReason = "game_completed"
			break
		}
		if before.Phase != session.PhasePlanning {
			stopReason = "interactive_boundary"
			break
		}
		view, err := hosted.session.PlayerView(info.ControlSeatID)
		if err != nil {
			return ReferenceAdvanceResult{}, err
		}
		if view.Seat.Submitted {
			return ReferenceAdvanceResult{}, fmt.Errorf("%w: control seat already submitted turn %d", ErrReferenceNotReady, before.Turn)
		}
		if draft, ok := hosted.planningDrafts[info.ControlSeatID]; ok && len(draft.Orders) != 0 {
			return ReferenceAdvanceResult{}, fmt.Errorf("%w: control seat has a non-empty planning draft", ErrReferenceNotReady)
		}

		batch := protocol.CommandBatch{
			SchemaVersion: protocol.CommandSchemaVersion,
			GameID:        before.GameID,
			SeatID:        info.ControlSeatID,
			Turn:          before.Turn,
			BaseRevision:  before.Revision,
			Commands:      []protocol.Command{},
		}
		if err := hosted.session.SubmitTurn(batch); err != nil {
			return ReferenceAdvanceResult{}, fmt.Errorf("%w: submit reference turn: %v", ErrReferenceNotReady, err)
		}
		delete(hosted.planningDrafts, info.ControlSeatID)
		driveErr := hosted.driveToInteractiveBoundary()
		after := hosted.session.Status()
		hosted.changeSequence++
		reason := "reference_advanced"
		if after.Phase == session.PhaseCompleted && before.Phase != session.PhaseCompleted {
			reason = "game_completed"
		} else if after.Turn > before.Turn {
			reason = "turn_advanced"
		}
		hosted.publish(Notification{
			SchemaVersion: SchemaVersion, Kind: "snapshot_invalidated", GameID: after.GameID,
			ChangeSequence: hosted.changeSequence, GameRevision: after.Revision, Scope: "reference", Reason: reason,
		})
		if driveErr != nil {
			return ReferenceAdvanceResult{}, fmt.Errorf("%w: %v", ErrSessionRejected, driveErr)
		}

		if construction != nil {
			completed, nextCursor, err := hosted.referenceConstructionCompleted(eventCursor, request.ColonyID, *construction)
			if err != nil {
				return ReferenceAdvanceResult{}, err
			}
			eventCursor = nextCursor
			if completed {
				stopReason = "construction_completed"
				break
			}
			currentView, err := hosted.session.PlayerView(info.ControlSeatID)
			if err != nil {
				return ReferenceAdvanceResult{}, err
			}
			same, err := referenceConstructionStillCurrent(currentView, request.ColonyID, *construction)
			if err != nil {
				return ReferenceAdvanceResult{}, err
			}
			if !same {
				stopReason = "construction_changed"
				break
			}
		}

		if after.Phase == session.PhaseCompleted || after.Result != nil {
			stopReason = "game_completed"
			break
		}
		if after.Phase != session.PhasePlanning || after.Turn == before.Turn {
			stopReason = "interactive_boundary"
			break
		}
		if request.Mode == ReferenceAdvanceTurns && step+1 == limit {
			stopReason = "requested_turns_reached"
			break
		}
	}

	final := hosted.session.Status()
	return ReferenceAdvanceResult{
		SchemaVersion:  SchemaVersion,
		GameID:         final.GameID,
		StartTurn:      start.Turn,
		EndTurn:        final.Turn,
		TurnsAdvanced:  final.Turn - start.Turn,
		StopReason:     stopReason,
		FinalPhase:     final.Phase,
		GameRevision:   final.Revision,
		ChangeSequence: hosted.changeSequence,
	}, nil
}

func referenceConstructionForView(view session.PlayerView, colonyID core.ID) (referenceConstructionSignature, error) {
	for i := range view.Colonies {
		colony := view.Colonies[i]
		if colony.ID != colonyID {
			continue
		}
		project := colony.Construction
		if project == nil || project.ProjectKind == "" {
			return referenceConstructionSignature{}, fmt.Errorf("%w: colony %d has no active construction", ErrReferenceNoActiveConstruction, colonyID)
		}
		if referenceCompletionEventKind(project.ProjectKind) == "" {
			return referenceConstructionSignature{}, fmt.Errorf("%w: colony %d project kind %q has no finite completion event", ErrReferenceNoActiveConstruction, colonyID, project.ProjectKind)
		}
		return referenceConstructionSignature{
			Kind: project.ProjectKind, ProjectID: project.ProjectID,
			ShipDesignID: project.ShipDesignID, ShipDesignRevision: project.ShipDesignRevision,
		}, nil
	}
	return referenceConstructionSignature{}, fmt.Errorf("%w: colony %d is not controlled by seat %d", ErrReferenceNoActiveConstruction, colonyID, view.Seat.Seat.ID)
}

func referenceConstructionStillCurrent(view session.PlayerView, colonyID core.ID, want referenceConstructionSignature) (bool, error) {
	for i := range view.Colonies {
		colony := view.Colonies[i]
		if colony.ID != colonyID {
			continue
		}
		if colony.Construction == nil {
			return false, nil
		}
		got := colony.Construction
		return got.ProjectKind == want.Kind && got.ProjectID == want.ProjectID &&
			got.ShipDesignID == want.ShipDesignID && got.ShipDesignRevision == want.ShipDesignRevision, nil
	}
	return false, fmt.Errorf("reference colony %d disappeared from control-seat view", colonyID)
}

func referenceCompletionEventKind(kind core.ConstructionProjectKind) string {
	switch kind {
	case core.ConstructionProjectBuilding:
		return "colony.building_completed"
	case core.ConstructionProjectColonyShip:
		return "colony.colony_ship_completed"
	case core.ConstructionProjectOutpostShip:
		return "colony.outpost_ship_completed"
	case core.ConstructionProjectTroopTransport:
		return "colony.troop_transport_completed"
	case core.ConstructionProjectMilitaryShip:
		return "colony.military_ship_completed"
	case core.ConstructionProjectFreighterFleet:
		return "colony.freighter_fleet_completed"
	case core.ConstructionProjectPlanetaryTransformation:
		return "colony.planetary_transformation_completed"
	default:
		return ""
	}
}

func (g *hostedGame) referenceEventCount() (int, error) {
	observer, err := g.session.ObserverView()
	if err != nil {
		return 0, err
	}
	return len(observer.Events), nil
}

func (g *hostedGame) referenceConstructionCompleted(cursor int, colonyID core.ID, want referenceConstructionSignature) (bool, int, error) {
	observer, err := g.session.ObserverView()
	if err != nil {
		return false, cursor, err
	}
	if cursor < 0 || cursor > len(observer.Events) {
		cursor = 0
	}
	kind := referenceCompletionEventKind(want.Kind)
	for _, event := range observer.Events[cursor:] {
		if event.Kind != kind {
			continue
		}
		var payload struct {
			ColonyID           core.ID `json:"colony_id"`
			BuildingID         string  `json:"building_id,omitempty"`
			ShipDesignID       core.ID `json:"ship_design_id,omitempty"`
			ShipDesignRevision uint32  `json:"ship_design_revision,omitempty"`
		}
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			return false, len(observer.Events), fmt.Errorf("decode reference completion event %q: %w", event.Kind, err)
		}
		if payload.ColonyID != colonyID {
			continue
		}
		if want.Kind == core.ConstructionProjectBuilding && payload.BuildingID != want.ProjectID {
			continue
		}
		if want.Kind == core.ConstructionProjectMilitaryShip &&
			(payload.ShipDesignID != want.ShipDesignID || payload.ShipDesignRevision != want.ShipDesignRevision) {
			continue
		}
		return true, len(observer.Events), nil
	}
	return false, len(observer.Events), nil
}
