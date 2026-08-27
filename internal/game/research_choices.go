package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

type ResearchChoice struct {
	TechFieldID         int      `json:"tech_field_id"`
	PreviousTechFieldID int      `json:"previous_tech_field_id"`
	NextTechFieldID     int      `json:"next_tech_field_id"`
	BaseCostRP          float64  `json:"base_cost_rp"`
	TechnologyIDs       []int    `json:"technology_ids"`
	TechnologyKeys      []string `json:"technology_keys"`
	TechnologyNameKeys  []string `json:"technology_name_keys"`
}

// AvailableResearchChoices returns the server-authoritative research frontier.
// A field is selectable when it is not already completed and its predecessor
// is known (or it is a root field). Technology membership is materialized by
// the server so Human UI and AI cannot invent arbitrary technology sets.
func (r *EconomyRules) AvailableResearchChoices(state *core.GameState, empireID core.ID) ([]ResearchChoice, error) {
	if r == nil {
		return nil, fmt.Errorf("economy rules must not be nil")
	}
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return nil, fmt.Errorf("unknown empire %d", empireID)
	}
	if empire.Research != nil {
		return []ResearchChoice{}, nil
	}

	knownFields := make(map[int]struct{}, len(empire.KnownTechnologyFieldIDs))
	for _, fieldID := range empire.KnownTechnologyFieldIDs {
		knownFields[fieldID] = struct{}{}
	}
	knownTech := make(map[int]struct{}, len(empire.KnownTechnologyIDs))
	for _, technologyID := range empire.KnownTechnologyIDs {
		knownTech[technologyID] = struct{}{}
	}

	fieldIDs := make([]int, 0, len(r.TechnologyFieldCostsRP))
	for fieldID := range r.TechnologyFieldCostsRP {
		fieldIDs = append(fieldIDs, fieldID)
	}
	sort.Ints(fieldIDs)

	choices := make([]ResearchChoice, 0)
	for _, fieldID := range fieldIDs {
		if _, known := knownFields[fieldID]; known {
			continue
		}
		previousID := r.TechnologyFieldPreviousID[fieldID]
		if previousID != 0 {
			if _, known := knownFields[previousID]; !known {
				continue
			}
		}

		allIDs := r.TechnologyIDsByField[fieldID]
		technologyIDs := make([]int, 0, len(allIDs))
		technologyKeys := make([]string, 0, len(allIDs))
		technologyNameKeys := make([]string, 0, len(allIDs))
		for _, technologyID := range allIDs {
			if _, known := knownTech[technologyID]; known {
				continue
			}
			key, ok := r.TechnologyKeyByID[technologyID]
			if !ok || key == "" {
				return nil, fmt.Errorf("technology %d has no stable internal key", technologyID)
			}
			nameKey, ok := r.TechnologyNameKeyByID[technologyID]
			if !ok || nameKey == "" {
				return nil, fmt.Errorf("technology %d has no name key", technologyID)
			}
			technologyIDs = append(technologyIDs, technologyID)
			technologyKeys = append(technologyKeys, key)
			technologyNameKeys = append(technologyNameKeys, nameKey)
		}
		// Hyper-advanced placeholders currently have no concrete Technology IDs
		// in the normalized table. Do not offer an unmaterializable choice.
		if len(technologyIDs) == 0 {
			continue
		}
		choices = append(choices, ResearchChoice{
			TechFieldID:         fieldID,
			PreviousTechFieldID: previousID,
			NextTechFieldID:     r.TechnologyFieldNextID[fieldID],
			BaseCostRP:          r.TechnologyFieldCostsRP[fieldID],
			TechnologyIDs:       technologyIDs,
			TechnologyKeys:      technologyKeys,
			TechnologyNameKeys:  technologyNameKeys,
		})
	}
	return choices, nil
}

func researchChoiceByField(choices []ResearchChoice, fieldID int) (ResearchChoice, bool) {
	index := sort.Search(len(choices), func(i int) bool { return choices[i].TechFieldID >= fieldID })
	if index >= len(choices) || choices[index].TechFieldID != fieldID {
		return ResearchChoice{}, false
	}
	return choices[index], true
}
