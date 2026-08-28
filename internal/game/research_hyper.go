package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

func (r *EconomyRules) isHyperAdvancedField(fieldID int) bool {
	if r == nil {
		return false
	}
	_, ok := r.HyperAdvancedResearchFieldIDs[fieldID]
	return ok
}

func hyperAdvancedCompletedLevels(empire *core.Empire, fieldID int) (int, bool) {
	if empire == nil {
		return 0, false
	}
	i := sort.Search(len(empire.HyperAdvancedResearch), func(i int) bool {
		return empire.HyperAdvancedResearch[i].TechFieldID >= fieldID
	})
	if i >= len(empire.HyperAdvancedResearch) || empire.HyperAdvancedResearch[i].TechFieldID != fieldID {
		return 0, false
	}
	return empire.HyperAdvancedResearch[i].CompletedLevels, true
}

func incrementHyperAdvancedCompletedLevels(empire *core.Empire, fieldID int) (int, error) {
	if empire == nil {
		return 0, fmt.Errorf("empire must not be nil")
	}
	if fieldID < 75 || fieldID > 82 {
		return 0, fmt.Errorf("TechField %d is not Hyper-Advanced", fieldID)
	}
	i := sort.Search(len(empire.HyperAdvancedResearch), func(i int) bool {
		return empire.HyperAdvancedResearch[i].TechFieldID >= fieldID
	})
	if i < len(empire.HyperAdvancedResearch) && empire.HyperAdvancedResearch[i].TechFieldID == fieldID {
		empire.HyperAdvancedResearch[i].CompletedLevels++
		return empire.HyperAdvancedResearch[i].CompletedLevels, nil
	}
	empire.HyperAdvancedResearch = append(empire.HyperAdvancedResearch, core.HyperAdvancedResearchLevel{})
	copy(empire.HyperAdvancedResearch[i+1:], empire.HyperAdvancedResearch[i:])
	empire.HyperAdvancedResearch[i] = core.HyperAdvancedResearchLevel{TechFieldID: fieldID, CompletedLevels: 1}
	return 1, nil
}

// researchFieldCostRP returns the authoritative simulation threshold. For
// Hyper-Advanced fields this intentionally follows Player_Research_Cost_ in
// MOO2 1.31: static field cost + persisted completed-level counter * increment.
// The original technology-selection UI temporarily promotes counters by one
// while drawing its preview; MOOX does not reproduce that UI-only off-by-one.
func (r *EconomyRules) researchFieldCostRP(empire *core.Empire, fieldID int) (float64, error) {
	if r == nil {
		return 0, fmt.Errorf("economy rules must not be nil")
	}
	base, ok := r.TechnologyFieldCostsRP[fieldID]
	if !ok {
		return 0, fmt.Errorf("unknown research tech field %d", fieldID)
	}
	if !r.isHyperAdvancedField(fieldID) {
		return base, nil
	}
	completed, _ := hyperAdvancedCompletedLevels(empire, fieldID)
	return base + float64(completed)*r.HyperAdvancedCostIncrementRP, nil
}
