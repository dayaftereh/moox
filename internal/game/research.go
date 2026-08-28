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
	NewGameRNG      *core.RNG
}

type ResearchCompletedEvent struct {
	EmpireID        core.ID                    `json:"empire_id"`
	TechFieldID     int                        `json:"tech_field_id"`
	SelectionMode   core.ResearchSelectionMode `json:"selection_mode"`
	TechnologyIDs   []int                      `json:"technology_ids"`
	TechnologyKeys  []string                   `json:"technology_keys"`
	CompletedLevels int                        `json:"completed_levels,omitempty"`
	ResearchLevel   int                        `json:"research_level,omitempty"`
}

type ResearchProgressedEvent struct {
	EmpireID        core.ID                    `json:"empire_id"`
	TechFieldID     int                        `json:"tech_field_id"`
	SelectionMode   core.ResearchSelectionMode `json:"selection_mode"`
	TechnologyIDs   []int                      `json:"technology_ids"`
	TechnologyKeys  []string                   `json:"technology_keys"`
	BaseCostRP      float64                    `json:"base_cost_rp"`
	PreviousRP      float64                    `json:"previous_rp"`
	TurnResearchRP  float64                    `json:"turn_research_rp"`
	ProjectedRP     float64                    `json:"projected_rp"`
	ChancePercent   int                        `json:"chance_percent"`
	Roll            int                        `json:"roll"`
	Breakthrough    bool                       `json:"breakthrough"`
	CompletedLevels int                        `json:"completed_levels,omitempty"`
	ResearchLevel   int                        `json:"research_level,omitempty"`
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
	modifiers, ok := r.RaceModifiers[empire.RaceID]
	if !ok {
		return fmt.Errorf("empire %d has unknown race %q", empire.ID, empire.RaceID)
	}
	if modifiers.Uncreative && options.NewGameRNG == nil {
		return fmt.Errorf("uncreative empire %d requires the caller-owned NewGameRNG", empire.ID)
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

// initializeUncreativeResearchChoices materializes the original new-game
// fixed-application plan for an Uncreative empire using the caller-owned RNG.
func (r *EconomyRules) initializeUncreativeResearchChoices(empire *core.Empire, options NewGameTechnologyOptions) error {
	modifiers, ok := r.RaceModifiers[empire.RaceID]
	if !ok {
		return fmt.Errorf("empire %d has unknown race %q", empire.ID, empire.RaceID)
	}
	empire.UncreativeResearchChoices = nil
	if !modifiers.Uncreative {
		return nil
	}
	if options.NewGameRNG == nil {
		return fmt.Errorf("uncreative empire %d requires the caller-owned NewGameRNG", empire.ID)
	}

	// Original MOO2 1.31 Init_Player_Tech_ selects fixed Uncreative
	// applications during Init_Players_. Its field loop is exactly 1..73;
	// field 74 is the special Antaran technology field and is not part of the
	// normal Uncreative research plan. General/start fields already have
	// available applications and therefore consume no Uncreative selection draw.
	const originalUncreativeFieldMax = 73
	choices := make([]core.FixedResearchChoice, 0, originalUncreativeFieldMax)
	for fieldID := 1; fieldID <= originalUncreativeFieldMax; fieldID++ {
		if _, general := r.GeneralResearchFieldIDs[fieldID]; general {
			continue
		}
		allIDs := r.TechnologyIDsByField[fieldID]
		if len(allIDs) == 0 {
			return fmt.Errorf("uncreative research field %d has no normalized technology applications", fieldID)
		}

		eligible := 0
		for _, technologyID := range allIDs {
			if r.uncreativeInitialTechnologyEligible(modifiers, options, technologyID) {
				eligible++
			}
		}
		if eligible == 0 {
			return fmt.Errorf("uncreative research field %d has no eligible technology applications", fieldID)
		}

		// The original draws from the field's application slots and retries when
		// the selected application is illegal for the race/game mode. Keep that
		// gameplay-level RNG consumption pattern while retaining MOOX's own
		// deterministic SplitMix64 generator rather than claiming LCG identity.
		for {
			index, err := options.NewGameRNG.Intn(len(allIDs))
			if err != nil {
				return err
			}
			technologyID := allIDs[index]
			if !r.uncreativeInitialTechnologyEligible(modifiers, options, technologyID) {
				continue
			}
			choices = append(choices, core.FixedResearchChoice{TechFieldID: fieldID, TechnologyID: technologyID})
			break
		}
	}
	empire.UncreativeResearchChoices = choices
	return nil
}

// uncreativeInitialTechnologyEligible mirrors the original Init_Player_Tech_
// race/mode exclusions that are already normalized in MOOX. The original also
// gates Technology 52 (Dimensional Portal) on global byte 0x21CAF; the meaning
// of that game-setting byte is not normalized yet, so that one gate remains
// deliberately deferred rather than guessed here.
func (r *EconomyRules) uncreativeInitialTechnologyEligible(modifiers RaceEconomyModifiers, options NewGameTechnologyOptions, technologyID int) bool {
	if options.StrategicCombat && !r.TechnologyStrategicAvailable[technologyID] {
		return false
	}
	if modifiers.GovernmentTraitID == "government_unification" {
		switch technologyID {
		case 86, 141, 195: // Holo Simulator, Pleasure Dome, Virtual Reality Network.
			return false
		}
	}
	if modifiers.Tolerant {
		switch technologyID {
		case 19, 50, 113, 142: // Pollution-management applications made redundant by Tolerant.
			return false
		}
	}
	if modifiers.Lithovore {
		switch technologyID {
		case 6, 29, 68, 87, 162, 178: // Food/farming applications made redundant by Lithovore.
			return false
		}
	}
	return true
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
	selectionMode := empire.Research.SelectionMode
	fieldCost, err := r.Rules.researchFieldCostRP(empire, fieldID)
	if err != nil {
		return DomainEvent{}, err
	}
	if empire.Research.ProgressRP < fieldCost {
		return DomainEvent{}, fmt.Errorf("empire %d research field %d has progress %.6f RP below cost %.6f RP", empireID, fieldID, empire.Research.ProgressRP, fieldCost)
	}
	if err := r.validateActiveResearchSelection(empire); err != nil {
		return DomainEvent{}, err
	}
	if r.Rules.isHyperAdvancedField(fieldID) {
		completedLevels, err := incrementHyperAdvancedCompletedLevels(empire, fieldID)
		if err != nil {
			return DomainEvent{}, err
		}
		empire.Research.ProgressRP = 0
		return NewDomainEvent("empire.research_completed", 0, 0, ResearchCompletedEvent{
			EmpireID:        empireID,
			TechFieldID:     fieldID,
			SelectionMode:   selectionMode,
			CompletedLevels: completedLevels,
			ResearchLevel:   completedLevels,
		})
	}
	if containsInt(empire.KnownTechnologyFieldIDs, fieldID) {
		return DomainEvent{}, fmt.Errorf("empire %d already completed technology field %d", empireID, fieldID)
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
	case core.ResearchSelectionRepeatField:
		if !r.Rules.isHyperAdvancedField(research.TechFieldID) || len(research.TechnologyIDs) != 0 {
			return fmt.Errorf("empire %d repeat-field research TechField %d must be Hyper-Advanced with no concrete technologies", empire.ID, research.TechFieldID)
		}
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
