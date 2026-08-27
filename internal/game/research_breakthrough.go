package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

// ResearchBreakthroughChancePercent reproduces the original MOO2 1.31 integer
// chance calculation at code VA 0xE1EC6. Inputs are whole RP, matching the
// original empire research fields rather than the engine's milli-RP economy.
func ResearchBreakthroughChancePercent(baseCostRP, projectedRP int64) int {
	if baseCostRP <= 0 || projectedRP <= baseCostRP {
		return 0
	}
	chance := ((projectedRP - baseCostRP) * 100) / baseCostRP
	if chance > 100 {
		return 100
	}
	if chance == 0 {
		return 1
	}
	return int(chance)
}

func (r *EconomyResolver) advanceResearch(state *core.GameState) ([]DomainEvent, error) {
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}

	empireIDs := make([]core.ID, 0, len(state.Empires))
	for i := range state.Empires {
		if state.Empires[i].Research != nil {
			empireIDs = append(empireIDs, state.Empires[i].ID)
		}
	}
	sort.Slice(empireIDs, func(i, j int) bool { return empireIDs[i] < empireIDs[j] })
	if len(empireIDs) == 0 {
		return nil, nil
	}

	rng := state.RNG()
	var events []DomainEvent
	for _, empireID := range empireIDs {
		empire := empireByID(state, empireID)
		if empire == nil || empire.Research == nil {
			continue
		}
		research := empire.Research
		if research.TechFieldID >= 75 {
			return nil, fmt.Errorf("empire %d research TechField %d uses hyper-advanced cost scaling that is not modeled yet", empireID, research.TechFieldID)
		}
		fieldCostMilli, ok := r.Rules.TechnologyFieldCostsMilli[research.TechFieldID]
		if !ok || fieldCostMilli <= 0 || fieldCostMilli%core.EconomyScale != 0 {
			return nil, fmt.Errorf("invalid base cost for research TechField %d", research.TechFieldID)
		}
		if research.ProgressMilli%core.EconomyScale != 0 {
			return nil, fmt.Errorf("empire %d research progress must use whole RP", empireID)
		}

		baseCostRP := fieldCostMilli / core.EconomyScale
		previousRP := research.ProgressMilli / core.EconomyScale
		turnResearchRP := implementedEmpireResearchRP(state, empireID)
		projectedRP := previousRP + turnResearchRP
		chance := ResearchBreakthroughChancePercent(baseCostRP, projectedRP)

		// MOO2 1.31 computes chance from accumulated + current research, then
		// adds current research and rolls random(100). Even a 0% chance consumes
		// one RNG draw for every active research project.
		research.ProgressMilli += turnResearchRP * core.EconomyScale
		rollZeroBased, err := rng.Intn(100)
		if err != nil {
			return nil, err
		}
		roll := rollZeroBased + 1
		breakthrough := roll <= chance

		ids := append([]int(nil), research.TechnologyIDs...)
		keys, err := r.technologyKeys(ids)
		if err != nil {
			return nil, err
		}
		progressEvent, err := NewDomainEvent("empire.research_progressed", 0, 0, ResearchProgressedEvent{
			EmpireID:       empireID,
			TechFieldID:    research.TechFieldID,
			TechnologyIDs:  ids,
			TechnologyKeys: keys,
			BaseCostRP:     baseCostRP,
			PreviousRP:     previousRP,
			TurnResearchRP: turnResearchRP,
			ProjectedRP:    projectedRP,
			ChancePercent:  chance,
			Roll:           roll,
			Breakthrough:   breakthrough,
		})
		if err != nil {
			return nil, err
		}
		events = append(events, progressEvent)

		if breakthrough {
			completed, err := r.CompleteResearchField(state, empireID)
			if err != nil {
				return nil, err
			}
			events = append(events, completed)
		}
	}
	state.CommitRNG(rng)
	return events, nil
}

// implementedEmpireResearchRP mirrors the proven integer boundary of the MOO2
// 1.31 empire research accumulator: each colony contributes a whole-RP value,
// those integers are summed, and the empire turn value is capped to int16 max.
// Leader/global research sources are intentionally excluded until implemented.
func implementedEmpireResearchRP(state *core.GameState, empireID core.ID) int64 {
	var total int64
	for i := range state.Colonies {
		colony := &state.Colonies[i]
		if colony.EmpireID != empireID {
			continue
		}
		total += colony.AdjustedEconomy.ResearchMilli / core.EconomyScale
		if total >= 32767 {
			return 32767
		}
	}
	return total
}

func (r *EconomyResolver) technologyKeys(ids []int) ([]string, error) {
	keys := make([]string, len(ids))
	for i, id := range ids {
		key, ok := r.Rules.TechnologyKeyByID[id]
		if !ok || key == "" {
			return nil, fmt.Errorf("technology %d has no stable internal key", id)
		}
		keys[i] = key
	}
	return keys, nil
}
