package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"moox/internal/core"
	"moox/internal/protocol"
)

const CommandSelectResearch = "empire.select_research"

type SelectResearchPayload struct {
	TechFieldID int `json:"tech_field_id"`
}

type ResearchSelectedEvent struct {
	EmpireID           core.ID  `json:"empire_id"`
	TechFieldID        int      `json:"tech_field_id"`
	BaseCostRP         float64  `json:"base_cost_rp"`
	TechnologyIDs      []int    `json:"technology_ids"`
	TechnologyKeys     []string `json:"technology_keys"`
	TechnologyNameKeys []string `json:"technology_name_keys"`
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
	if empire.Research != nil {
		return DomainEvent{}, fmt.Errorf("empire %d already has active research TechField %d", empireID, empire.Research.TechFieldID)
	}
	choices, err := r.Rules.AvailableResearchChoices(state, empireID)
	if err != nil {
		return DomainEvent{}, err
	}
	choice, ok := researchChoiceByField(choices, payload.TechFieldID)
	if !ok {
		return DomainEvent{}, fmt.Errorf("TechField %d is not a legal research choice for empire %d", payload.TechFieldID, empireID)
	}
	empire.Research = &core.ResearchState{
		TechFieldID:   choice.TechFieldID,
		TechnologyIDs: append([]int(nil), choice.TechnologyIDs...),
		ProgressRP:    0,
	}
	return NewDomainEvent("empire.research_selected", seatID, command.Sequence, ResearchSelectedEvent{
		EmpireID:           empireID,
		TechFieldID:        choice.TechFieldID,
		BaseCostRP:         choice.BaseCostRP,
		TechnologyIDs:      append([]int(nil), choice.TechnologyIDs...),
		TechnologyKeys:     append([]string(nil), choice.TechnologyKeys...),
		TechnologyNameKeys: append([]string(nil), choice.TechnologyNameKeys...),
	})
}
