package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

const ReferenceTriangleMidTechScenarioID = "triangle-2pc-mid-tech-v1"
const ReferenceTriangleAllTechScenarioID = "triangle-2pc-all-tech-v1"

type ReferenceTriangleTechnologyProfile string

const (
	ReferenceTriangleProfileBaseline ReferenceTriangleTechnologyProfile = "baseline"
	ReferenceTriangleProfileMidTech  ReferenceTriangleTechnologyProfile = "mid_tech"
	ReferenceTriangleProfileAllTech  ReferenceTriangleTechnologyProfile = "all_tech"
)

var referenceTriangleMidTechFieldIDs = []int{
	0, 1, 2, 3, 4, 5, 7, 9, 10, 15, 16, 18, 19, 20, 21, 22, 23, 28,
	29, 31, 34, 35, 36, 41, 43, 45, 47, 54, 55, 56, 57, 60, 62, 63, 66, 73,
}

func ReferenceTriangleMidTechFieldIDs() []int {
	return append([]int(nil), referenceTriangleMidTechFieldIDs...)
}

func ReferenceTriangleAllTechFieldIDs() []int {
	ids := make([]int, 75)
	for i := range ids {
		ids[i] = i
	}
	return ids
}

func (r *EconomyRules) ApplyReferenceTriangleTechnologyProfile(state *core.GameState, profile ReferenceTriangleTechnologyProfile) error {
	if r == nil {
		return fmt.Errorf("reference technology profile requires economy rules")
	}
	if state == nil {
		return fmt.Errorf("reference technology profile requires game state")
	}
	if profile == ReferenceTriangleProfileBaseline {
		return state.Validate()
	}

	var fieldIDs []int
	switch profile {
	case ReferenceTriangleProfileMidTech:
		fieldIDs = ReferenceTriangleMidTechFieldIDs()
	case ReferenceTriangleProfileAllTech:
		fieldIDs = ReferenceTriangleAllTechFieldIDs()
	default:
		return fmt.Errorf("unknown reference triangle technology profile %q", profile)
	}

	technologyIDs, err := r.referenceTechnologyIDsForFields(fieldIDs)
	if err != nil {
		return err
	}
	for i := range state.Empires {
		empire := &state.Empires[i]
		empire.KnownTechnologyFieldIDs = append([]int(nil), fieldIDs...)
		empire.KnownTechnologyIDs = append([]int(nil), technologyIDs...)
		empire.Research = nil
		empire.HyperAdvancedResearch = nil
	}
	if err := state.Validate(); err != nil {
		return fmt.Errorf("validate reference triangle %s technology profile: %w", profile, err)
	}
	return nil
}

func (r *EconomyRules) referenceTechnologyIDsForFields(fieldIDs []int) ([]int, error) {
	seenFields := make(map[int]struct{}, len(fieldIDs))
	seenTechnologies := make(map[int]struct{})
	for _, fieldID := range fieldIDs {
		if fieldID < 0 || fieldID > 74 {
			return nil, fmt.Errorf("reference profile field %d outside concrete field range [0,74]", fieldID)
		}
		if _, duplicate := seenFields[fieldID]; duplicate {
			return nil, fmt.Errorf("reference profile duplicates field %d", fieldID)
		}
		seenFields[fieldID] = struct{}{}
		ids, ok := r.TechnologyIDsByField[fieldID]
		if !ok && fieldID != 0 {
			return nil, fmt.Errorf("reference profile field %d is absent from normalized technology rules", fieldID)
		}
		for _, technologyID := range ids {
			if technologyID < 1 || technologyID > 203 {
				return nil, fmt.Errorf("reference profile field %d contains invalid technology %d", fieldID, technologyID)
			}
			seenTechnologies[technologyID] = struct{}{}
		}
	}
	technologyIDs := make([]int, 0, len(seenTechnologies))
	for technologyID := range seenTechnologies {
		technologyIDs = append(technologyIDs, technologyID)
	}
	sort.Ints(technologyIDs)
	return technologyIDs, nil
}
