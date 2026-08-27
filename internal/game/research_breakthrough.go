package game

import (
	"fmt"
	"math"
	"sort"

	"moox/internal/core"
)

// ResearchBreakthroughChancePercent keeps the proven MOO2 1.31 breakthrough
// curve while allowing MOOX to accumulate research in fractional RP. The
// floating-point percentage is floored only at the final 1..100 roll boundary.
func ResearchBreakthroughChancePercent(baseCostRP, projectedRP float64) int {
	if math.IsNaN(baseCostRP) || math.IsNaN(projectedRP) || math.IsInf(baseCostRP, 0) || math.IsInf(projectedRP, 0) || baseCostRP <= 0 || projectedRP <= baseCostRP {
		return 0
	}
	chance := math.Floor(((projectedRP - baseCostRP) * 100) / baseCostRP)
	if chance > 100 {
		return 100
	}
	if chance < 1 {
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
		baseCostRP, ok := r.Rules.TechnologyFieldCostsRP[research.TechFieldID]
		if !ok || baseCostRP <= 0 || math.IsNaN(baseCostRP) || math.IsInf(baseCostRP, 0) {
			return nil, fmt.Errorf("invalid base cost for research TechField %d", research.TechFieldID)
		}

		previousRP := research.ProgressRP
		turnResearchRP := implementedEmpireResearchRP(state, empireID)
		projectedRP := previousRP + turnResearchRP
		chance := ResearchBreakthroughChancePercent(baseCostRP, projectedRP)

		// MOOX deliberately keeps fractional RP instead of reproducing the old
		// integer storage constraint. Rounding happens only where the rule needs
		// a discrete 1..100 breakthrough comparison.
		research.ProgressRP = projectedRP
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

// implementedEmpireResearchRP bridges the fixed-point colony economy into the
// research domain without per-colony truncation. Research therefore preserves
// fractional RP all the way until a rule explicitly requires rounding.
func implementedEmpireResearchRP(state *core.GameState, empireID core.ID) float64 {
	var totalMilli int64
	for i := range state.Colonies {
		colony := &state.Colonies[i]
		if colony.EmpireID != empireID {
			continue
		}
		totalMilli += colony.AdjustedEconomy.ResearchMilli
	}
	return float64(totalMilli) / float64(core.EconomyScale)
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
