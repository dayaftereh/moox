package game

import "fmt"

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
