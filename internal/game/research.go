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
	Level                   NewGameTechnologyLevel
	StrategicCombat         bool
	UncreativeSelectionSeed uint64
}

type ResearchCompletedEvent struct {
	EmpireID       core.ID                    `json:"empire_id"`
	TechFieldID    int                        `json:"tech_field_id"`
	SelectionMode  core.ResearchSelectionMode `json:"selection_mode"`
	TechnologyIDs  []int                      `json:"technology_ids"`
	TechnologyKeys []string                   `json:"technology_keys"`
}

type ResearchProgressedEvent struct {
	EmpireID       core.ID                    `json:"empire_id"`
	TechFieldID    int                        `json:"tech_field_id"`
	SelectionMode  core.ResearchSelectionMode `json:"selection_mode"`
	TechnologyIDs  []int                      `json:"technology_ids"`
	TechnologyKeys []string                   `json:"technology_keys"`
	BaseCostRP     float64                    `json:"base_cost_rp"`
	PreviousRP     float64                    `json:"previous_rp"`
	TurnResearchRP float64                    `json:"turn_research_rp"`
	ProjectedRP    float64                    `json:"projected_rp"`
	ChancePercent  int                        `json:"chance_percent"`
	Roll           int                        `json:"roll"`
	Breakthrough   bool                       `json:"breakthrough"`
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
	if err := r.initializeUncreativeResearchChoices(empire, options); err != nil {
		return err
	}
	return nil
}

// CompleteResearchField materializes only the ownership transition that occurs
// after a research breakthrough has already been decided by the authoritative
// simulation. Breakthrough probability/overflow and research selection are
// intentionally separate from this transition.
func (r *EconomyRules) initializeUncreativeResearchChoices(empire *core.Empire, options NewGameTechnologyOptions) error {
	modifiers, ok := r.RaceModifiers[empire.RaceID]
	if !ok {
		return fmt.Errorf("empire %d has unknown race %q", empire.ID, empire.RaceID)
	}
	empire.UncreativeResearchChoices = nil
	if !modifiers.Uncreative {
		return nil
	}
	if options.UncreativeSelectionSeed == 0 {
		return fmt.Errorf("uncreative empire %d requires a nonzero UncreativeSelectionSeed", empire.ID)
	}

	fieldIDs := make([]int, 0, len(r.TechnologyIDsByField))
	for fieldID := range r.TechnologyIDsByField {
		if fieldID <= 0 || fieldID >= 75 {
			continue
		}
		if _, general := r.GeneralResearchFieldIDs[fieldID]; general {
			continue
		}
		fieldIDs = append(fieldIDs, fieldID)
	}
	sort.Ints(fieldIDs)

	seed := options.UncreativeSelectionSeed ^ (uint64(empire.ID) * 0x9e3779b97f4a7c15)
	rng := core.NewRNG(seed)
	choices := make([]core.FixedResearchChoice, 0, len(fieldIDs))
	for _, fieldID := range fieldIDs {
		allIDs := r.TechnologyIDsByField[fieldID]
		ids := make([]int, 0, len(allIDs))
		for _, technologyID := range allIDs {
			if options.StrategicCombat && !r.TechnologyStrategicAvailable[technologyID] {
				continue
			}
			ids = append(ids, technologyID)
		}
		if len(ids) == 0 {
			continue
		}
		index, err := rng.Intn(len(ids))
		if err != nil {
			return err
		}
		choices = append(choices, core.FixedResearchChoice{TechFieldID: fieldID, TechnologyID: ids[index]})
	}
	empire.UncreativeResearchChoices = choices
	return nil
}

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
	selectionMode := empire.Research.SelectionMode
	fieldCost, ok := r.Rules.TechnologyFieldCostsRP[fieldID]
	if !ok {
		return DomainEvent{}, fmt.Errorf("unknown research tech field %d", fieldID)
	}
	if containsInt(empire.KnownTechnologyFieldIDs, fieldID) {
		return DomainEvent{}, fmt.Errorf("empire %d already completed technology field %d", empireID, fieldID)
	}
	if empire.Research.ProgressRP < fieldCost {
		return DomainEvent{}, fmt.Errorf("empire %d research field %d has progress %.6f RP below base cost %.6f RP", empireID, fieldID, empire.Research.ProgressRP, fieldCost)
	}
	if len(empire.Research.TechnologyIDs) == 0 {
		return DomainEvent{}, fmt.Errorf("empire %d research has no selected technology ids", empireID)
	}
	if err := r.validateActiveResearchSelection(empire); err != nil {
		return DomainEvent{}, err
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
	keys, err := r.technologyKeys(acquired)
	if err != nil {
		return DomainEvent{}, err
	}
	return NewDomainEvent("empire.research_completed", 0, 0, ResearchCompletedEvent{
		EmpireID:       empireID,
		TechFieldID:    fieldID,
		SelectionMode:  selectionMode,
		TechnologyIDs:  acquired,
		TechnologyKeys: keys,
	})
}

func (r *EconomyResolver) validateActiveResearchSelection(empire *core.Empire) error {
	if empire == nil || empire.Research == nil {
		return fmt.Errorf("active research is required")
	}
	modifiers, ok := r.Rules.RaceModifiers[empire.RaceID]
	if !ok {
		return fmt.Errorf("empire %d has unknown race %q", empire.ID, empire.RaceID)
	}
	research := empire.Research
	expectedMode := r.Rules.researchSelectionMode(modifiers, research.TechFieldID)
	if research.SelectionMode != expectedMode {
		return fmt.Errorf("empire %d research field %d selection mode=%q, expected %q", empire.ID, research.TechFieldID, research.SelectionMode, expectedMode)
	}
	switch expectedMode {
	case core.ResearchSelectionChooseOne:
		if len(research.TechnologyIDs) != 1 {
			return fmt.Errorf("empire %d choose-one research field %d has %d technologies, expected 1", empire.ID, research.TechFieldID, len(research.TechnologyIDs))
		}
	case core.ResearchSelectionFixedOne:
		if len(research.TechnologyIDs) != 1 {
			return fmt.Errorf("empire %d fixed-one research field %d has %d technologies, expected 1", empire.ID, research.TechFieldID, len(research.TechnologyIDs))
		}
		fixedID, found := fixedResearchTechnology(empire.UncreativeResearchChoices, research.TechFieldID)
		if !found || research.TechnologyIDs[0] != fixedID {
			return fmt.Errorf("empire %d fixed-one research field %d does not match persisted Uncreative choice", empire.ID, research.TechFieldID)
		}
	case core.ResearchSelectionAll:
		expected := make([]int, 0, len(r.Rules.TechnologyIDsByField[research.TechFieldID]))
		known := make(map[int]struct{}, len(empire.KnownTechnologyIDs))
		for _, id := range empire.KnownTechnologyIDs {
			known[id] = struct{}{}
		}
		for _, id := range r.Rules.TechnologyIDsByField[research.TechFieldID] {
			if _, exists := known[id]; !exists {
				expected = append(expected, id)
			}
		}
		if len(expected) != len(research.TechnologyIDs) {
			return fmt.Errorf("empire %d all-applications research field %d has %d technologies, expected %d", empire.ID, research.TechFieldID, len(research.TechnologyIDs), len(expected))
		}
		for i := range expected {
			if research.TechnologyIDs[i] != expected[i] {
				return fmt.Errorf("empire %d all-applications research field %d technology set does not match server field membership", empire.ID, research.TechFieldID)
			}
		}
	default:
		return fmt.Errorf("unsupported research selection mode %q", expectedMode)
	}
	return nil
}

func containsInt(ids []int, want int) bool {
	index := sort.SearchInts(ids, want)
	return index < len(ids) && ids[index] == want
}
