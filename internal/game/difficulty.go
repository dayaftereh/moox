package game

import (
	"fmt"

	"moox/internal/core"
)

// DifficultyProfile is the authoritative server-owned Slice-16.2 V1 contract.
// Resource deltas and command-deficit rates are stored in eighth-units so the
// persisted/wire values remain exact while the economy may calculate in float64.
type DifficultyProfile struct {
	ID                                core.DifficultyID `json:"id"`
	AIFoodPerFarmerEighths            int               `json:"ai_food_per_farmer_eighths"`
	AIProductionPerWorkerEighths      int               `json:"ai_production_per_worker_eighths"`
	AIResearchPerScientistEighths     int               `json:"ai_research_per_scientist_eighths"`
	AITaxBCPerPopulationEighths       int               `json:"ai_tax_bc_per_population_eighths"`
	AICommandDeficitBCPerPointEighths int               `json:"ai_command_deficit_bc_per_point_eighths"`
}

const DefaultDifficultyID core.DifficultyID = core.DifficultyNormal

var difficultyProfiles = [...]DifficultyProfile{
	{ID: core.DifficultyEasy, AIFoodPerFarmerEighths: -2, AIProductionPerWorkerEighths: -4, AIResearchPerScientistEighths: -4, AITaxBCPerPopulationEighths: -2, AICommandDeficitBCPerPointEighths: 88},
	{ID: core.DifficultyNormal, AIFoodPerFarmerEighths: 0, AIProductionPerWorkerEighths: 0, AIResearchPerScientistEighths: 0, AITaxBCPerPopulationEighths: 0, AICommandDeficitBCPerPointEighths: 80},
	{ID: core.DifficultyHard, AIFoodPerFarmerEighths: 2, AIProductionPerWorkerEighths: 4, AIResearchPerScientistEighths: 4, AITaxBCPerPopulationEighths: 2, AICommandDeficitBCPerPointEighths: 72},
	{ID: core.DifficultyVeryHard, AIFoodPerFarmerEighths: 4, AIProductionPerWorkerEighths: 8, AIResearchPerScientistEighths: 8, AITaxBCPerPopulationEighths: 3, AICommandDeficitBCPerPointEighths: 68},
	{ID: core.DifficultyImpossible, AIFoodPerFarmerEighths: 6, AIProductionPerWorkerEighths: 12, AIResearchPerScientistEighths: 12, AITaxBCPerPopulationEighths: 4, AICommandDeficitBCPerPointEighths: 64},
}

func DifficultyProfiles() []DifficultyProfile {
	profiles := make([]DifficultyProfile, len(difficultyProfiles))
	copy(profiles, difficultyProfiles[:])
	return profiles
}

func DifficultyProfileFor(id core.DifficultyID) (DifficultyProfile, error) {
	for _, profile := range difficultyProfiles {
		if profile.ID == id {
			return profile, nil
		}
	}
	return DifficultyProfile{}, fmt.Errorf("unsupported difficulty_id %q", id)
}

func difficultyEighths(value int) float64 {
	return float64(value) / 8.0
}

func difficultyProfileForEmpire(state *core.GameState, empire core.Empire) (DifficultyProfile, bool, error) {
	if !empire.BuiltinAIControlled {
		return DifficultyProfile{}, false, nil
	}
	profile, err := DifficultyProfileFor(state.EffectiveDifficultyID())
	if err != nil {
		return DifficultyProfile{}, false, err
	}
	return profile, true, nil
}
