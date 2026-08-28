package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"moox/internal/core"
	"moox/internal/protocol"
)

const CommandSelectResearch = "empire.select_research"

type SelectResearchPayload struct {
	TechFieldID  int `json:"tech_field_id"`
	TechnologyID int `json:"technology_id,omitempty"`
}

type ResearchSelectedEvent struct {
	EmpireID           core.ID                    `json:"empire_id"`
	TechFieldID        int                        `json:"tech_field_id"`
	BaseCostRP         float64                    `json:"base_cost_rp"`
	SelectionMode      core.ResearchSelectionMode `json:"selection_mode"`
	TechnologyIDs      []int                      `json:"technology_ids"`
	TechnologyKeys     []string                   `json:"technology_keys"`
	TechnologyNameKeys []string                   `json:"technology_name_keys"`
}

type ResearchProjectSnapshot struct {
	TechFieldID    int                        `json:"tech_field_id"`
	SelectionMode  core.ResearchSelectionMode `json:"selection_mode"`
	TechnologyIDs  []int                      `json:"technology_ids"`
	TechnologyKeys []string                   `json:"technology_keys"`
}

type ResearchSwitchedEvent struct {
	EmpireID      core.ID                 `json:"empire_id"`
	Previous      ResearchProjectSnapshot `json:"previous"`
	Current       ResearchProjectSnapshot `json:"current"`
	TransferredRP float64                 `json:"transferred_rp"`
}

func NewSelectResearchCommand(sequence uint32, payload SelectResearchPayload) (protocol.Command, error) {
	if err := validateSelectResearchPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandSelectResearch, payload)
}

func decodeSelectResearch(command protocol.Command) (SelectResearchPayload, error) {
	var payload SelectResearchPayload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return SelectResearchPayload{}, fmt.Errorf("decode %s: %w", CommandSelectResearch, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return SelectResearchPayload{}, fmt.Errorf("decode %s: trailing JSON value", CommandSelectResearch)
		}
		return SelectResearchPayload{}, fmt.Errorf("decode %s trailing data: %w", CommandSelectResearch, err)
	}
	if err := validateSelectResearchPayload(payload); err != nil {
		return SelectResearchPayload{}, err
	}
	return payload, nil
}

func validateSelectResearchPayload(payload SelectResearchPayload) error {
	if payload.TechFieldID <= 0 {
		return fmt.Errorf("tech_field_id must be positive")
	}
	if payload.TechnologyID < 0 {
		return fmt.Errorf("technology_id must not be negative")
	}
	return nil
}

func (r *EconomyResolver) selectResearch(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeSelectResearch(command)
	if err != nil {
		return DomainEvent{}, err
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("unknown empire %d", empireID)
	}
	choices, err := r.Rules.AvailableResearchChoices(state, empireID)
	if err != nil {
		return DomainEvent{}, err
	}
	choice, ok := researchChoiceByField(choices, payload.TechFieldID)
	if !ok {
		return DomainEvent{}, fmt.Errorf("TechField %d is not a legal research choice for empire %d", payload.TechFieldID, empireID)
	}

	acquiredIDs, err := researchAcquisitionSet(choice, payload.TechnologyID)
	if err != nil {
		return DomainEvent{}, fmt.Errorf("select research TechField %d: %w", payload.TechFieldID, err)
	}
	keys, nameKeys, err := r.Rules.researchTechnologyMetadata(acquiredIDs)
	if err != nil {
		return DomainEvent{}, err
	}

	progressRP := 0.0
	var previous *core.ResearchState
	if empire.Research != nil {
		copy := cloneResearchState(empire.Research)
		previous = &copy
		progressRP = empire.Research.ProgressRP
	}
	current := core.ResearchState{
		TechFieldID:   choice.TechFieldID,
		SelectionMode: choice.SelectionMode,
		TechnologyIDs: append([]int(nil), acquiredIDs...),
		ProgressRP:    progressRP,
	}
	if previous != nil && sameResearchSelection(*previous, current) {
		return DomainEvent{}, fmt.Errorf("empire %d already researches TechField %d with the same application selection", empireID, current.TechFieldID)
	}

	if previous == nil {
		empire.Research = &current
		return NewDomainEvent("empire.research_selected", seatID, command.Sequence, ResearchSelectedEvent{
			EmpireID:           empireID,
			TechFieldID:        choice.TechFieldID,
			BaseCostRP:         choice.BaseCostRP,
			SelectionMode:      choice.SelectionMode,
			TechnologyIDs:      append([]int(nil), acquiredIDs...),
			TechnologyKeys:     keys,
			TechnologyNameKeys: nameKeys,
		})
	}

	previousSnapshot, err := r.Rules.researchProjectSnapshot(*previous)
	if err != nil {
		return DomainEvent{}, err
	}
	currentSnapshot := ResearchProjectSnapshot{
		TechFieldID:    current.TechFieldID,
		SelectionMode:  current.SelectionMode,
		TechnologyIDs:  append([]int(nil), current.TechnologyIDs...),
		TechnologyKeys: append([]string(nil), keys...),
	}
	empire.Research = &current
	return NewDomainEvent("empire.research_switched", seatID, command.Sequence, ResearchSwitchedEvent{
		EmpireID:      empireID,
		Previous:      previousSnapshot,
		Current:       currentSnapshot,
		TransferredRP: progressRP,
	})
}

func researchAcquisitionSet(choice ResearchChoice, requestedTechnologyID int) ([]int, error) {
	switch choice.SelectionMode {
	case core.ResearchSelectionAll:
		if requestedTechnologyID != 0 {
			return nil, fmt.Errorf("technology_id must be omitted for all-applications research")
		}
		return append([]int(nil), choice.TechnologyIDs...), nil
	case core.ResearchSelectionChooseOne:
		if requestedTechnologyID == 0 {
			return nil, fmt.Errorf("technology_id is required when one application must be chosen")
		}
		if !containsInt(choice.TechnologyIDs, requestedTechnologyID) {
			return nil, fmt.Errorf("technology_id %d is not a legal application of TechField %d", requestedTechnologyID, choice.TechFieldID)
		}
		return []int{requestedTechnologyID}, nil
	case core.ResearchSelectionFixedOne:
		if requestedTechnologyID != 0 {
			return nil, fmt.Errorf("technology_id must be omitted for fixed Uncreative research")
		}
		if len(choice.TechnologyIDs) != 1 {
			return nil, fmt.Errorf("fixed research choice for TechField %d has %d applications, expected 1", choice.TechFieldID, len(choice.TechnologyIDs))
		}
		return append([]int(nil), choice.TechnologyIDs...), nil
	default:
		return nil, fmt.Errorf("unsupported research selection mode %q", choice.SelectionMode)
	}
}

func (r *EconomyRules) researchTechnologyMetadata(ids []int) ([]string, []string, error) {
	keys := make([]string, len(ids))
	nameKeys := make([]string, len(ids))
	for i, technologyID := range ids {
		key, ok := r.TechnologyKeyByID[technologyID]
		if !ok || key == "" {
			return nil, nil, fmt.Errorf("technology %d has no stable internal key", technologyID)
		}
		nameKey, ok := r.TechnologyNameKeyByID[technologyID]
		if !ok || nameKey == "" {
			return nil, nil, fmt.Errorf("technology %d has no name key", technologyID)
		}
		keys[i] = key
		nameKeys[i] = nameKey
	}
	return keys, nameKeys, nil
}

func (r *EconomyRules) researchProjectSnapshot(state core.ResearchState) (ResearchProjectSnapshot, error) {
	keys, _, err := r.researchTechnologyMetadata(state.TechnologyIDs)
	if err != nil {
		return ResearchProjectSnapshot{}, err
	}
	return ResearchProjectSnapshot{
		TechFieldID:    state.TechFieldID,
		SelectionMode:  state.SelectionMode,
		TechnologyIDs:  append([]int(nil), state.TechnologyIDs...),
		TechnologyKeys: keys,
	}, nil
}

func cloneResearchState(state *core.ResearchState) core.ResearchState {
	if state == nil {
		return core.ResearchState{}
	}
	copy := *state
	copy.TechnologyIDs = append([]int(nil), state.TechnologyIDs...)
	return copy
}

func sameResearchSelection(a, b core.ResearchState) bool {
	return a.TechFieldID == b.TechFieldID && a.SelectionMode == b.SelectionMode && reflect.DeepEqual(a.TechnologyIDs, b.TechnologyIDs)
}
