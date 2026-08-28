package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

// UncreativeResearchRepairOptions carries only the original repair-path gates
// that have semantic MOOX representations. DimensionalPortalAllowed remains a
// pointer because the original 0x21CAF setting is not normalized yet: callers
// must explicitly provide the eligibility decision before Technology 52 can be
// considered instead of silently guessing the old global setting.
type UncreativeResearchRepairOptions struct {
	StrategicCombat          bool
	DimensionalPortalAllowed *bool
}

// UncreativeResearchRepairResult describes the persisted fixed-choice mutation
// caused by a post-acquisition repair. ReplacementTechnologyID is zero when no
// legal replacement remains for the field.
type UncreativeResearchRepairResult struct {
	TechFieldID             int  `json:"tech_field_id"`
	PreviousTechnologyID    int  `json:"previous_technology_id"`
	ReplacementTechnologyID int  `json:"replacement_technology_id,omitempty"`
	Changed                 bool `json:"changed"`
}

// RepairUncreativeResearchChoiceAfterAcquisition mirrors MOO2 1.31
// Ensure_Uncreative_Field_OK_. The acquired Technology must already be present
// in empire.KnownTechnologyIDs, matching the original Player_Gets_Tech_App_
// ordering. The helper consumes one caller-owned RNG draw per eligible state-0
// replacement candidate in original TechField slot order (which Init_Tech_
// builds by ascending Technology ID).
func (r *EconomyRules) RepairUncreativeResearchChoiceAfterAcquisition(
	empire *core.Empire,
	acquiredTechnologyID int,
	options UncreativeResearchRepairOptions,
	rng *core.RNG,
) (UncreativeResearchRepairResult, error) {
	if r == nil {
		return UncreativeResearchRepairResult{}, fmt.Errorf("economy rules must not be nil")
	}
	if empire == nil {
		return UncreativeResearchRepairResult{}, fmt.Errorf("empire must not be nil")
	}
	if !containsInt(empire.KnownTechnologyIDs, acquiredTechnologyID) {
		return UncreativeResearchRepairResult{}, fmt.Errorf("acquired Technology %d must already be known before Uncreative repair", acquiredTechnologyID)
	}
	modifiers, ok := r.RaceModifiers[empire.RaceID]
	if !ok {
		return UncreativeResearchRepairResult{}, fmt.Errorf("empire %d has unknown race %q", empire.ID, empire.RaceID)
	}
	if !modifiers.Uncreative {
		return UncreativeResearchRepairResult{}, nil
	}
	fieldID, ok := r.TechnologyFieldByID[acquiredTechnologyID]
	if !ok || fieldID <= 0 || fieldID > 82 {
		return UncreativeResearchRepairResult{}, fmt.Errorf("Technology %d has invalid TechField %d", acquiredTechnologyID, fieldID)
	}

	choiceIndex := sort.Search(len(empire.UncreativeResearchChoices), func(i int) bool {
		return empire.UncreativeResearchChoices[i].TechFieldID >= fieldID
	})
	if choiceIndex >= len(empire.UncreativeResearchChoices) || empire.UncreativeResearchChoices[choiceIndex].TechFieldID != fieldID {
		return UncreativeResearchRepairResult{TechFieldID: fieldID}, nil
	}

	previousID := empire.UncreativeResearchChoices[choiceIndex].TechnologyID
	result := UncreativeResearchRepairResult{
		TechFieldID:          fieldID,
		PreviousTechnologyID: previousID,
	}

	// The persisted fixed choice is the semantic equivalent of original status
	// 1 while it is still unknown. Other unknown applications are status 0.
	fixedStillAvailable := !containsInt(empire.KnownTechnologyIDs, previousID)
	candidateCount := 0
	replacementID := 0
	for _, technologyID := range r.TechnologyIDsByField[fieldID] {
		if technologyID == previousID && fixedStillAvailable {
			continue
		}
		if containsInt(empire.KnownTechnologyIDs, technologyID) {
			continue
		}
		eligible, err := r.uncreativeRepairTechnologyEligible(fieldID, technologyID, options)
		if err != nil {
			return UncreativeResearchRepairResult{}, err
		}
		if !eligible {
			continue
		}
		candidateCount++
		if rng == nil {
			return UncreativeResearchRepairResult{}, fmt.Errorf("Uncreative repair for TechField %d requires the caller-owned RNG", fieldID)
		}
		roll, err := rng.Intn(candidateCount)
		if err != nil {
			return UncreativeResearchRepairResult{}, err
		}
		if roll == 0 {
			replacementID = technologyID
		}
	}

	// Original Ensure_Uncreative_Field_OK_ still consumes the reservoir draws
	// above when a status-1 application already exists, but commits no change.
	if fixedStillAvailable {
		result.ReplacementTechnologyID = previousID
		return result, nil
	}

	result.Changed = true
	result.ReplacementTechnologyID = replacementID
	if replacementID == 0 {
		empire.UncreativeResearchChoices = append(
			empire.UncreativeResearchChoices[:choiceIndex],
			empire.UncreativeResearchChoices[choiceIndex+1:]...,
		)
		return result, nil
	}
	empire.UncreativeResearchChoices[choiceIndex].TechnologyID = replacementID
	return result, nil
}

func (r *EconomyRules) uncreativeRepairTechnologyEligible(fieldID, technologyID int, options UncreativeResearchRepairOptions) (bool, error) {
	// Original Ensure_Uncreative_Field_OK_ explicitly refuses to repair the
	// government-evolution family (Confederation/Federation/Unification/Imperium).
	if fieldID == 6 {
		return false, nil
	}
	if technologyID == 52 { // Dimensional Portal.
		if options.DimensionalPortalAllowed == nil {
			return false, fmt.Errorf("Dimensional Portal Uncreative-repair eligibility requires the unresolved original 0x21CAF policy")
		}
		if !*options.DimensionalPortalAllowed {
			return false, nil
		}
	}
	if options.StrategicCombat && !r.TechnologyStrategicAvailable[technologyID] {
		return false, nil
	}
	return true, nil
}
