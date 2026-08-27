package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

type NewGameTechnologyLevel string

const (
	NewGameTechnologyPreWarp  NewGameTechnologyLevel = "pre_warp"
	NewGameTechnologyAverage  NewGameTechnologyLevel = "average"
	NewGameTechnologyAdvanced NewGameTechnologyLevel = "advanced"
)

type NewGameTechnologyOptions struct {
	Level           NewGameTechnologyLevel
	StrategicCombat bool
}

type ResearchCompletedEvent struct {
	EmpireID      core.ID `json:"empire_id"`
	TechFieldID   int     `json:"tech_field_id"`
	TechnologyIDs []int   `json:"technology_ids"`
}

// InitializeEmpireTechnologies applies only new-game starts whose field set is
// deterministic from the normalized original tables. Advanced remains deferred
// because its extra research fields depend on the original randomized/race-aware
// start generator.
func (r *EconomyRules) InitializeEmpireTechnologies(empire *core.Empire, options NewGameTechnologyOptions) error {
	if r == nil {
		return fmt.Errorf("economy rules must not be nil")
	}
	if empire == nil {
		return fmt.Errorf("empire must not be nil")
	}
	if len(empire.KnownTechnologyIDs) != 0 || len(empire.KnownTechnologyFieldIDs) != 0 || empire.Research != nil {
		return fmt.Errorf("empire %d technology state is already initialized", empire.ID)
	}
	fields := []int{r.NewGameAlwaysKnownFieldID}
	switch options.Level {
	case NewGameTechnologyPreWarp:
		if len(r.NewGameStagedKnownFieldIDs) == 0 {
			return fmt.Errorf("new-game staged technology fields are missing")
		}
		fields = append(fields, r.NewGameStagedKnownFieldIDs[0])
	case NewGameTechnologyAverage:
		fields = append(fields, r.NewGameStagedKnownFieldIDs...)
	case NewGameTechnologyAdvanced:
		return fmt.Errorf("advanced new-game technology generation is not implemented: randomized/race-aware extra fields are not yet normalized")
	default:
		return fmt.Errorf("unsupported new-game technology level %q", options.Level)
	}

	fieldSet := make(map[int]struct{}, len(fields))
	for _, fieldID := range fields {
		fieldSet[fieldID] = struct{}{}
	}
	empire.KnownTechnologyFieldIDs = make([]int, 0, len(fieldSet))
	for fieldID := range fieldSet {
		empire.KnownTechnologyFieldIDs = append(empire.KnownTechnologyFieldIDs, fieldID)
	}
	sort.Ints(empire.KnownTechnologyFieldIDs)

	known := make(map[int]struct{})
	for _, fieldID := range empire.KnownTechnologyFieldIDs {
		for _, technologyID := range r.TechnologyIDsByField[fieldID] {
			if options.StrategicCombat && !r.TechnologyStrategicAvailable[technologyID] {
				continue
			}
			known[technologyID] = struct{}{}
		}
	}
	empire.KnownTechnologyIDs = make([]int, 0, len(known))
	for technologyID := range known {
		empire.KnownTechnologyIDs = append(empire.KnownTechnologyIDs, technologyID)
	}
	sort.Ints(empire.KnownTechnologyIDs)
	return nil
}

// CompleteResearchField materializes only the ownership transition that occurs
// after a research breakthrough has already been decided by the authoritative
// simulation. Breakthrough probability/overflow and research selection are
// intentionally separate from this transition.
func (r *EconomyResolver) CompleteResearchField(state *core.GameState, empireID core.ID) (DomainEvent, error) {
	if r == nil || r.Rules == nil {
		return DomainEvent{}, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return DomainEvent{}, fmt.Errorf("game state must not be nil")
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("unknown empire %d", empireID)
	}
	if empire.Research == nil {
		return DomainEvent{}, fmt.Errorf("empire %d has no active research", empireID)
	}
	fieldID := empire.Research.TechFieldID
	fieldCost, ok := r.Rules.TechnologyFieldCostsMilli[fieldID]
	if !ok {
		return DomainEvent{}, fmt.Errorf("unknown research tech field %d", fieldID)
	}
	if containsInt(empire.KnownTechnologyFieldIDs, fieldID) {
		return DomainEvent{}, fmt.Errorf("empire %d already completed technology field %d", empireID, fieldID)
	}
	if empire.Research.ProgressMilli < fieldCost {
		return DomainEvent{}, fmt.Errorf("empire %d research field %d has progress %d below base cost %d", empireID, fieldID, empire.Research.ProgressMilli, fieldCost)
	}
	if len(empire.Research.TechnologyIDs) == 0 {
		return DomainEvent{}, fmt.Errorf("empire %d research has no selected technology ids", empireID)
	}
	known := make(map[int]struct{}, len(empire.KnownTechnologyIDs))
	for _, technologyID := range empire.KnownTechnologyIDs {
		known[technologyID] = struct{}{}
	}
	acquired := append([]int(nil), empire.Research.TechnologyIDs...)
	if !sort.IntsAreSorted(acquired) {
		return DomainEvent{}, fmt.Errorf("empire %d research technology ids must be sorted", empireID)
	}
	last := 0
	for _, technologyID := range acquired {
		if technologyID == last {
			return DomainEvent{}, fmt.Errorf("empire %d research contains duplicate technology %d", empireID, technologyID)
		}
		last = technologyID
		techFieldID, ok := r.Rules.TechnologyFieldByID[technologyID]
		if !ok {
			return DomainEvent{}, fmt.Errorf("unknown technology id %d", technologyID)
		}
		if techFieldID != fieldID {
			return DomainEvent{}, fmt.Errorf("technology %d belongs to tech field %d, not active field %d", technologyID, techFieldID, fieldID)
		}
		if _, exists := known[technologyID]; exists {
			return DomainEvent{}, fmt.Errorf("empire %d already knows technology %d", empireID, technologyID)
		}
		known[technologyID] = struct{}{}
	}
	for _, technologyID := range acquired {
		empire.KnownTechnologyIDs = append(empire.KnownTechnologyIDs, technologyID)
	}
	sort.Ints(empire.KnownTechnologyIDs)
	empire.KnownTechnologyFieldIDs = append(empire.KnownTechnologyFieldIDs, fieldID)
	sort.Ints(empire.KnownTechnologyFieldIDs)
	empire.Research = nil
	return NewDomainEvent("empire.research_completed", 0, 0, ResearchCompletedEvent{
		EmpireID:      empireID,
		TechFieldID:   fieldID,
		TechnologyIDs: acquired,
	})
}

func containsInt(ids []int, want int) bool {
	index := sort.SearchInts(ids, want)
	return index < len(ids) && ids[index] == want
}
