package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

const advancedNewGameExtraGrantCount = 19

// TechnologyAIClassDefinition is normalized original MOO2 AI/research metadata.
// It is runtime rules data, not authoritative game state.
type TechnologyAIClassDefinition struct {
	BaseWeight           int
	CompetitionSensitive bool
}

// RaceResearchModifiers is the semantic MOOX projection of the original
// player race block consumed by Calc_Tech_Value_. It deliberately stores game
// concepts rather than the original player+0x8A0 byte layout.
type RaceResearchModifiers struct {
	PopulationGrowthPercent int
	FarmingDelta            float64
	IndustryDelta           float64
	ScienceDelta            float64
	MoneyDelta              float64
	ShipDefenseBonus        int
	ShipAttackBonus         int
	GroundCombatBonus       int
	SpyingBonus             int
	LowGWorld               bool
	HighGWorld              bool
	Aquatic                 bool
	Subterranean            bool
	Cybernetic              bool
	Lithovore               bool
	Tolerant                bool
	Telepathic              bool
	StealthyShips           bool
	GovernmentTraitID       string
}

// AdvancedResearchPreferenceProfile represents the three original preference
// selectors consumed by Calc_Tech_Value_. These are kept explicit until the
// New Game personality/objective/theme generator is normalized separately.
type AdvancedResearchPreferenceProfile struct {
	Personality int
	Objective   int
	Theme       int
}

func (p AdvancedResearchPreferenceProfile) Validate() error {
	if p.Personality < 0 || p.Personality > 5 {
		return fmt.Errorf("advanced research personality %d outside [0,5]", p.Personality)
	}
	if p.Objective < 0 || p.Objective > 3 {
		return fmt.Errorf("advanced research objective %d outside [0,3]", p.Objective)
	}
	if p.Theme < 0 || p.Theme > 6 {
		return fmt.Errorf("advanced research theme %d outside [0,6]", p.Theme)
	}
	return nil
}

func advancedObjectiveWeight(weight, classID, objective int) int {
	switch objective {
	case 0:
		switch classID {
		case 25, 26, 27:
			return 100
		}
	case 1:
		switch classID {
		case 26:
			return 50
		case 19, 30:
			return 100
		case 25:
			return 20
		}
	case 2:
		switch classID {
		case 21, 24, 29:
			return 100
		}
	case 3:
		switch classID {
		case 36:
			return 50
		case 38:
			return 100
		}
	}
	return weight
}

func advancedThemeWeight(weight, classID, theme int) int {
	switch theme {
	case 0:
		if classID == 18 || classID == 23 {
			return 50
		}
	case 1:
		switch classID {
		case 15, 16:
			return 20
		case 32:
			return 50
		case 35:
			return 100
		}
	case 2:
		if classID == 18 {
			return 50
		}
		if classID == 20 {
			return 100
		}
	case 3:
		if classID == 28 {
			return 100
		}
		if classID == 26 {
			return 20
		}
	case 4:
		if classID == 37 {
			return 100
		}
	case 5:
		if classID == 31 || classID == 32 {
			return 100
		}
	case 6:
		if classID == 33 || classID == 34 {
			return 100
		}
	}
	return weight
}

func advancedPersonalityWeight(weight, classID, personality int) int {
	switch personality {
	case 0:
		if classID == 3 || classID == 11 || classID == 12 {
			return 100
		}
	case 1:
		if classID == 11 || classID == 17 || classID == 33 {
			return 100
		}
	case 2:
		if classID == 39 {
			return 50
		}
		if classID == 9 || classID == 10 {
			return 100
		}
	case 3:
		if classID == 2 {
			return 100
		}
	case 4:
		if classID == 1 || classID == 4 {
			return 100
		}
	case 5:
		if classID == 0 || classID == 4 {
			return 100
		}
	}
	return weight
}

func advancedRaceWeight(weight, classID int, race RaceResearchModifiers) int {
	switch classID {
	case 0:
		if race.FarmingDelta < 0 {
			weight = 100
		} else if race.FarmingDelta > 0 {
			weight = 10
		}
		if race.Lithovore {
			return 1
		}
		if race.Cybernetic {
			return 20
		}
	case 1:
		if race.IndustryDelta < 0 {
			return 100
		}
	case 2:
		if race.ScienceDelta != 0 {
			return 100
		}
	case 3:
		if race.MoneyDelta < 0 {
			return 100
		}
		if race.MoneyDelta > 0 {
			return 20
		}
	case 4:
		if race.IndustryDelta > 0 {
			weight = 100
		}
		if race.Tolerant {
			return 1
		}
	case 6:
		if race.Subterranean {
			weight = 20
		}
		if race.PopulationGrowthPercent < 0 {
			return 100
		}
		if race.PopulationGrowthPercent > 0 {
			return 5
		}
	case 12:
		if race.SpyingBonus != 0 || race.GovernmentTraitID == "government_democracy" {
			return 50
		}
	case 16:
		if race.GroundCombatBonus < 0 {
			return 20
		}
	case 18:
		if race.ShipDefenseBonus < 0 {
			return 50
		}
	case 25:
		if race.ShipAttackBonus < 0 {
			return 100
		}
	case 27:
		if race.ShipAttackBonus > 0 {
			return 100
		}
	case 28:
		if race.ShipDefenseBonus > 0 {
			return 100
		}
	case 37:
		if race.StealthyShips {
			return 1
		}
	case 40:
		if race.GovernmentTraitID == "government_unification" {
			return 1
		}
	}
	return weight
}

func advancedSpecialTechnologyWeight(weight, technologyID int, race RaceResearchModifiers) int {
	switch technologyID {
	case 5: // Alien Management Center is redundant for Telepathic races.
		if race.Telepathic {
			return 1
		}
	case 131: // Planetary Gravity Generator value depends on native gravity.
		if race.HighGWorld {
			return 1
		}
		if race.LowGWorld {
			return 50
		}
	}
	return weight
}

func (r *EconomyRules) advancedProfileRaceBaseWeight(empire *core.Empire, technologyID int, profile AdvancedResearchPreferenceProfile) (int, error) {
	if empire == nil {
		return 0, fmt.Errorf("empire must not be nil")
	}
	if err := profile.Validate(); err != nil {
		return 0, err
	}
	classID, ok := r.TechnologyAIClassByID[technologyID]
	if !ok {
		return 0, fmt.Errorf("Technology %d has no AI class", technologyID)
	}
	class, ok := r.TechnologyAIClasses[classID]
	if !ok {
		return 0, fmt.Errorf("Technology %d references unknown AI class %d", technologyID, classID)
	}
	race, ok := r.RaceResearchModifiers[empire.RaceID]
	if !ok {
		return 0, fmt.Errorf("empire %d has no research modifiers for race %q", empire.ID, empire.RaceID)
	}
	weight := class.BaseWeight
	weight = advancedObjectiveWeight(weight, classID, profile.Objective)
	weight = advancedThemeWeight(weight, classID, profile.Theme)
	weight = advancedPersonalityWeight(weight, classID, profile.Personality)
	weight = advancedRaceWeight(weight, classID, race)
	weight = advancedSpecialTechnologyWeight(weight, technologyID, race)
	return weight, nil
}

func (r *EconomyRules) advancedTechnologyEligible(empire *core.Empire, technologyID int, strategicCombat bool) bool {
	if empire == nil {
		return false
	}
	fieldID, ok := r.TechnologyFieldByID[technologyID]
	if !ok || fieldID <= 0 || fieldID >= 75 {
		return false
	}
	modifiers, ok := r.RaceModifiers[empire.RaceID]
	if !ok {
		return false
	}
	if !r.uncreativeInitialTechnologyEligible(modifiers, NewGameTechnologyOptions{StrategicCombat: strategicCombat}, technologyID) {
		return false
	}
	switch technologyID {
	case 42: // Confederation.
		return modifiers.GovernmentTraitID == "government_feudal"
	case 65: // Federation.
		return modifiers.GovernmentTraitID == "government_democracy"
	case 77: // Galactic Unification.
		return modifiers.GovernmentTraitID == "government_unification"
	case 92: // Imperium.
		return modifiers.GovernmentTraitID == "government_dictatorship"
	}
	return true
}

func (r *EconomyRules) maxKnownAIGroup(empire *core.Empire, classID int, technologyLimit int) int {
	if empire == nil {
		return 0
	}
	maxGroup := 0
	for _, technologyID := range empire.KnownTechnologyIDs {
		if technologyLimit > 0 && technologyID > technologyLimit {
			continue
		}
		if r.TechnologyAIClassByID[technologyID] != classID {
			continue
		}
		fieldID := r.TechnologyFieldByID[technologyID]
		if group := r.TechnologyFieldAIGroup[fieldID]; group > maxGroup {
			maxGroup = group
		}
	}
	return maxGroup
}

func (r *EconomyRules) competitionAIGroup(state *core.GameState, empireID core.ID, classID int) int {
	maxGroup := 0
	for i := range state.Empires {
		empire := &state.Empires[i]
		if empire.ID == empireID {
			continue
		}
		if group := r.maxKnownAIGroup(empire, classID, 75); group > maxGroup {
			maxGroup = group
		}
	}
	return maxGroup
}

func advancedProgressionWeight(weight, classID, targetAIGroup, targetProgression, knownAIGroup int, competitionSensitive bool) int {
	if !competitionSensitive {
		return weight * targetProgression
	}
	if knownAIGroup+3 <= targetAIGroup {
		return (targetAIGroup - knownAIGroup) * weight * targetProgression / 3
	}
	switch classID {
	case 14, 16, 18, 24, 25, 32, 39:
		return 0
	}
	denominator := knownAIGroup + 3 - targetAIGroup
	if denominator <= 0 {
		return 0
	}
	return weight * targetProgression * 2 / denominator
}

func advancedGapMultiplier(weight, targetProgression, competitionAIGroup int) int {
	if targetProgression >= competitionAIGroup {
		return weight
	}
	return weight * (competitionAIGroup - targetProgression)
}

func (r *EconomyRules) advancedCompetitionWeight(state *core.GameState, empire *core.Empire, classID, targetProgression, weight int, profile AdvancedResearchPreferenceProfile) int {
	competition := func(aiClass int) int {
		return r.competitionAIGroup(state, empire.ID, aiClass)
	}
	switch classID {
	case 25:
		weight = advancedGapMultiplier(weight, targetProgression, competition(18))
	case 18:
		weight = advancedGapMultiplier(weight, targetProgression, competition(25))
	case 19:
		weight = advancedGapMultiplier(weight, targetProgression, competition(10))
	case 10:
		weight = advancedGapMultiplier(weight, targetProgression, competition(19))
	case 12:
		weight = advancedGapMultiplier(weight, targetProgression, competition(12))
	case 15:
		weight = advancedGapMultiplier(weight, targetProgression, competition(15))
	case 8:
		own := r.maxKnownAIGroup(empire, 15, 75) + r.maxKnownAIGroup(empire, 16, 75)
		if own < 2*competition(15) {
			weight *= 2
		}
	}
	if profile.Objective == 2 && classID == 24 {
		weight = advancedGapMultiplier(weight, targetProgression, competition(28))
	}
	switch profile.Objective {
	case 0:
		if classID == 19 {
			weight = advancedGapMultiplier(weight, targetProgression, competition(33))
		}
	case 1:
		if classID == 26 || classID == 19 {
			weight = advancedGapMultiplier(weight, targetProgression, competition(33))
		}
	case 2:
		if classID == 21 {
			weight = advancedGapMultiplier(weight, targetProgression, competition(33))
		}
	}
	return weight
}

func (r *EconomyRules) allOtherFieldTechnologiesKnown(empire *core.Empire, fieldID, technologyID int) bool {
	for _, otherID := range r.TechnologyIDsByField[fieldID] {
		if otherID == technologyID {
			continue
		}
		if !containsInt(empire.KnownTechnologyIDs, otherID) {
			return false
		}
	}
	return true
}

func (r *EconomyRules) advancedCalcTechnologyWeight(state *core.GameState, empireID core.ID, technologyID int, profile AdvancedResearchPreferenceProfile, strategicCombat bool) (int, error) {
	empire := empireByID(state, empireID)
	if empire == nil {
		return 0, fmt.Errorf("unknown empire %d", empireID)
	}
	if !r.advancedTechnologyEligible(empire, technologyID, strategicCombat) || containsInt(empire.KnownTechnologyIDs, technologyID) {
		return 0, nil
	}
	// Calc_Tech_Value_ suppresses Dimensional Portal for every non-sentinel
	// personality. Init_Player_Tech_ temporarily replaces the Human sentinel
	// 100 with profile 1, so Advanced start reaches the same zero-weight path.
	if technologyID == 52 {
		return 0, nil
	}
	weight, err := r.advancedProfileRaceBaseWeight(empire, technologyID, profile)
	if err != nil {
		return 0, err
	}
	classID := r.TechnologyAIClassByID[technologyID]
	class, ok := r.TechnologyAIClasses[classID]
	if !ok {
		return 0, fmt.Errorf("Technology %d references unknown AI class %d", technologyID, classID)
	}
	fieldID := r.TechnologyFieldByID[technologyID]
	targetAIGroup := r.TechnologyFieldAIGroup[fieldID]
	if targetAIGroup < 0 {
		targetAIGroup = 0
	}
	if targetAIGroup > 22 {
		targetAIGroup = 22
	}
	if targetAIGroup >= len(r.TechnologyAIFieldGroupValues) {
		return 0, fmt.Errorf("TechField %d AI group %d has no progression value", fieldID, targetAIGroup)
	}
	targetProgression := r.TechnologyAIFieldGroupValues[targetAIGroup]
	knownAIGroup := r.maxKnownAIGroup(empire, classID, 0)
	weight = advancedProgressionWeight(weight, classID, targetAIGroup, targetProgression, knownAIGroup, class.CompetitionSensitive)
	weight = r.advancedCompetitionWeight(state, empire, classID, targetProgression, weight, profile)

	// Advanced starting grants run at game creation and therefore always take
	// the original early-game (<150 turn-equivalent) class-18 multiplier.
	if classID == 18 {
		weight *= 2
	}
	if weight == 0 && r.allOtherFieldTechnologiesKnown(empire, fieldID, technologyID) {
		weight = targetProgression * 10
	}
	return weight, nil
}

type NewGameTechnologyStateOptions struct {
	Level               NewGameTechnologyLevel
	StrategicCombat     bool
	NewGameRNG          *core.RNG
	AdvancedPreferences map[core.ID]AdvancedResearchPreferenceProfile
}

type advancedTechnologyChooser func(state *core.GameState, empireID core.ID, profile AdvancedResearchPreferenceProfile, strategicCombat bool, rng *core.RNG) (int, error)

// InitializeNewGameTechnologies is the full-state technology initialization
// boundary. Pre-Warp and Average delegate to the deterministic per-Empire
// initializer. Advanced requires cross-Empire competition weighting and is
// enabled once the verified Calc_Tech_Value_ translation is supplied.
func (r *EconomyRules) InitializeNewGameTechnologies(state *core.GameState, options NewGameTechnologyStateOptions) error {
	if options.Level == NewGameTechnologyAdvanced {
		return fmt.Errorf("advanced new-game technology weighting is not implemented yet; generator orchestration is available but Calc_Tech_Value_ translation is still required")
	}
	if state == nil {
		return fmt.Errorf("game state must not be nil")
	}
	empireIndexes := sortedEmpireIndexes(state)
	for _, index := range empireIndexes {
		if err := r.InitializeEmpireTechnologies(&state.Empires[index], NewGameTechnologyOptions{
			Level:           options.Level,
			StrategicCombat: options.StrategicCombat,
			NewGameRNG:      options.NewGameRNG,
		}); err != nil {
			return err
		}
	}
	return nil
}

// initializeAdvancedNewGameTechnologies contains the fully verified state and
// ordering mechanics around the still-separate weighting function. Keeping the
// chooser injectable makes the 19-grant orchestration testable without
// pretending an approximate weight function is original behavior.
func (r *EconomyRules) initializeAdvancedNewGameTechnologies(state *core.GameState, options NewGameTechnologyStateOptions, chooser advancedTechnologyChooser) error {
	if r == nil {
		return fmt.Errorf("economy rules must not be nil")
	}
	if state == nil {
		return fmt.Errorf("game state must not be nil")
	}
	if options.Level != NewGameTechnologyAdvanced {
		return fmt.Errorf("advanced initializer requires level %q", NewGameTechnologyAdvanced)
	}
	if options.NewGameRNG == nil {
		return fmt.Errorf("advanced new-game technology generation requires the caller-owned NewGameRNG")
	}
	if chooser == nil {
		return fmt.Errorf("advanced technology chooser must not be nil")
	}

	empireIndexes := sortedEmpireIndexes(state)
	for _, index := range empireIndexes {
		empire := &state.Empires[index]
		if _, ok := r.RaceModifiers[empire.RaceID]; !ok {
			return fmt.Errorf("empire %d has unknown race %q", empire.ID, empire.RaceID)
		}
		profile, ok := options.AdvancedPreferences[empire.ID]
		if !ok {
			return fmt.Errorf("advanced research preference profile is required for empire %d", empire.ID)
		}
		if err := profile.Validate(); err != nil {
			return fmt.Errorf("empire %d: %w", empire.ID, err)
		}
		if len(empire.KnownTechnologyIDs) != 0 || len(empire.KnownTechnologyFieldIDs) != 0 || empire.Research != nil || len(empire.UncreativeResearchChoices) != 0 {
			return fmt.Errorf("empire %d technology state is already initialized", empire.ID)
		}
	}

	for _, index := range empireIndexes {
		empire := &state.Empires[index]
		profile := options.AdvancedPreferences[empire.ID]
		if err := r.InitializeEmpireTechnologies(empire, NewGameTechnologyOptions{
			Level:           NewGameTechnologyAverage,
			StrategicCombat: options.StrategicCombat,
			NewGameRNG:      options.NewGameRNG,
		}); err != nil {
			return err
		}
		for grant := 0; grant < advancedNewGameExtraGrantCount; grant++ {
			technologyID, err := chooser(state, empire.ID, profile, options.StrategicCombat, options.NewGameRNG)
			if err != nil {
				return fmt.Errorf("advanced grant %d for empire %d: %w", grant+1, empire.ID, err)
			}
			if err := r.grantAdvancedStartingTechnology(state, empire.ID, technologyID, options.StrategicCombat); err != nil {
				return fmt.Errorf("advanced grant %d for empire %d: %w", grant+1, empire.ID, err)
			}
		}
	}
	return nil
}

func sortedEmpireIndexes(state *core.GameState) []int {
	indexes := make([]int, len(state.Empires))
	for i := range indexes {
		indexes[i] = i
	}
	sort.Slice(indexes, func(i, j int) bool {
		return state.Empires[indexes[i]].ID < state.Empires[indexes[j]].ID
	})
	return indexes
}

func (r *EconomyRules) advancedCandidateTechnologyIDs(state *core.GameState, empireID core.ID, strategicCombat bool) ([]int, error) {
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
	fixed := make(map[int]int, len(empire.UncreativeResearchChoices))
	for _, choice := range empire.UncreativeResearchChoices {
		fixed[choice.TechFieldID] = choice.TechnologyID
	}

	var candidates []int
	for fieldID := 1; fieldID < 75; fieldID++ {
		if _, known := knownFields[fieldID]; known {
			continue
		}
		previousID, ok := r.TechnologyFieldPreviousID[fieldID]
		if !ok {
			continue
		}
		if previousID != 0 {
			if _, known := knownFields[previousID]; !known {
				continue
			}
		}
		ids := r.TechnologyIDsByField[fieldID]
		if modifiers.Uncreative {
			technologyID, ok := fixed[fieldID]
			if !ok {
				return nil, fmt.Errorf("uncreative empire %d has no fixed application for open TechField %d", empireID, fieldID)
			}
			ids = []int{technologyID}
		}
		for _, technologyID := range ids {
			if !r.advancedTechnologyEligible(empire, technologyID, strategicCombat) {
				continue
			}
			if _, known := knownTech[technologyID]; known {
				continue
			}
			if strategicCombat && !r.TechnologyStrategicAvailable[technologyID] {
				continue
			}
			candidates = append(candidates, technologyID)
		}
	}
	sort.Ints(candidates)
	return candidates, nil
}

func (r *EconomyRules) grantAdvancedStartingTechnology(state *core.GameState, empireID core.ID, technologyID int, strategicCombat bool) error {
	empire := empireByID(state, empireID)
	if empire == nil {
		return fmt.Errorf("unknown empire %d", empireID)
	}
	candidates, err := r.advancedCandidateTechnologyIDs(state, empireID, strategicCombat)
	if err != nil {
		return err
	}
	if !containsInt(candidates, technologyID) {
		return fmt.Errorf("Technology %d is not an open Advanced-start candidate", technologyID)
	}
	fieldID, ok := r.TechnologyFieldByID[technologyID]
	if !ok || fieldID <= 0 || fieldID >= 75 {
		return fmt.Errorf("Technology %d has invalid Advanced-start TechField %d", technologyID, fieldID)
	}
	modifiers := r.RaceModifiers[empire.RaceID]
	acquired := []int{technologyID}
	if modifiers.Creative {
		acquired = acquired[:0]
		for _, id := range r.TechnologyIDsByField[fieldID] {
			if strategicCombat && !r.TechnologyStrategicAvailable[id] {
				continue
			}
			acquired = append(acquired, id)
		}
	}
	if modifiers.Uncreative {
		fixedID := 0
		for _, choice := range empire.UncreativeResearchChoices {
			if choice.TechFieldID == fieldID {
				fixedID = choice.TechnologyID
				break
			}
		}
		if fixedID == 0 || fixedID != technologyID {
			return fmt.Errorf("Technology %d does not match Uncreative fixed application %d for TechField %d", technologyID, fixedID, fieldID)
		}
	}

	empire.KnownTechnologyFieldIDs = appendUniqueSortedInt(empire.KnownTechnologyFieldIDs, fieldID)
	for _, id := range acquired {
		empire.KnownTechnologyIDs = appendUniqueSortedInt(empire.KnownTechnologyIDs, id)
	}
	return nil
}

func appendUniqueSortedInt(values []int, value int) []int {
	index := sort.SearchInts(values, value)
	if index < len(values) && values[index] == value {
		return values
	}
	values = append(values, 0)
	copy(values[index+1:], values[index:])
	values[index] = value
	return values
}
