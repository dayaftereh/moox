package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

type ResearchChoice struct {
	TechFieldID         int                        `json:"tech_field_id"`
	PreviousTechFieldID int                        `json:"previous_tech_field_id"`
	NextTechFieldID     int                        `json:"next_tech_field_id"`
	BaseCostRP          float64                    `json:"base_cost_rp"`
	SelectionMode       core.ResearchSelectionMode `json:"selection_mode"`
	TechnologyIDs       []int                      `json:"technology_ids"`
	TechnologyKeys      []string                   `json:"technology_keys"`
	TechnologyNameKeys  []string                   `json:"technology_name_keys"`
}

// AvailableResearchChoices returns the server-authoritative research frontier.
// A field is selectable when it is not already completed and its predecessor
// is known (or it is a root field). The server also projects the race-specific
// application policy so Human UI, built-in AI and remote agents receive the
// same legal actions without inventing Technology ownership.
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
	modifiers, ok := r.RaceModifiers[empire.RaceID]
	if !ok {
		return nil, fmt.Errorf("empire %d has unknown race %q", empireID, empire.RaceID)
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
		for _, technologyID := range allIDs {
			if _, known := knownTech[technologyID]; known {
				continue
			}
			technologyIDs = append(technologyIDs, technologyID)
		}
		// Hyper-advanced placeholders currently have no concrete Technology IDs
		// in the normalized table. Do not offer an unmaterializable choice.
		if len(technologyIDs) == 0 {
			continue
		}

		mode := r.researchSelectionMode(modifiers, fieldID)
		if mode == core.ResearchSelectionFixedOne {
			fixedID, found := fixedResearchTechnology(empire.UncreativeResearchChoices, fieldID)
			if !found {
				return nil, fmt.Errorf("uncreative empire %d has no fixed technology for research field %d", empireID, fieldID)
			}
			if _, known := knownTech[fixedID]; known {
				return nil, fmt.Errorf("uncreative empire %d already knows fixed technology %d for incomplete field %d; external-acquisition semantics are not modeled yet", empireID, fixedID, fieldID)
			}
			if !containsInt(technologyIDs, fixedID) {
				return nil, fmt.Errorf("uncreative empire %d fixed technology %d is not a legal application of field %d", empireID, fixedID, fieldID)
			}
			technologyIDs = []int{fixedID}
		}

		technologyKeys := make([]string, len(technologyIDs))
		technologyNameKeys := make([]string, len(technologyIDs))
		for i, technologyID := range technologyIDs {
			key, ok := r.TechnologyKeyByID[technologyID]
			if !ok || key == "" {
				return nil, fmt.Errorf("technology %d has no stable internal key", technologyID)
			}
			nameKey, ok := r.TechnologyNameKeyByID[technologyID]
			if !ok || nameKey == "" {
				return nil, fmt.Errorf("technology %d has no name key", technologyID)
			}
			technologyKeys[i] = key
			technologyNameKeys[i] = nameKey
		}
		choices = append(choices, ResearchChoice{
			TechFieldID:         fieldID,
			PreviousTechFieldID: previousID,
			NextTechFieldID:     r.TechnologyFieldNextID[fieldID],
			BaseCostRP:          r.TechnologyFieldCostsRP[fieldID],
			SelectionMode:       mode,
			TechnologyIDs:       technologyIDs,
			TechnologyKeys:      technologyKeys,
			TechnologyNameKeys:  technologyNameKeys,
		})
	}
	return choices, nil
}

func (r *EconomyRules) researchSelectionMode(modifiers RaceEconomyModifiers, fieldID int) core.ResearchSelectionMode {
	if _, general := r.GeneralResearchFieldIDs[fieldID]; general {
		return core.ResearchSelectionAll
	}
	if modifiers.Creative {
		return core.ResearchSelectionAll
	}
	if modifiers.Uncreative {
		return core.ResearchSelectionFixedOne
	}
	return core.ResearchSelectionChooseOne
}

func fixedResearchTechnology(choices []core.FixedResearchChoice, fieldID int) (int, bool) {
	index := sort.Search(len(choices), func(i int) bool { return choices[i].TechFieldID >= fieldID })
	if index >= len(choices) || choices[index].TechFieldID != fieldID {
		return 0, false
	}
	return choices[index].TechnologyID, true
}

func researchChoiceByField(choices []ResearchChoice, fieldID int) (ResearchChoice, bool) {
	index := sort.Search(len(choices), func(i int) bool { return choices[i].TechFieldID >= fieldID })
	if index >= len(choices) || choices[index].TechFieldID != fieldID {
		return ResearchChoice{}, false
	}
	return choices[index], true
}
