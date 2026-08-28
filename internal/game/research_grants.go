package game

import (
	"fmt"

	"moox/internal/core"
)

const EventTechnologyGranted = "empire.technology_granted"

type TechnologyGrantSource string

type TechnologyGrantOptions struct {
	SourceKind           TechnologyGrantSource
	StrategicCombat      bool
	RandomEventsDisabled bool
}

type TechnologyGrantedEvent struct {
	EmpireID          core.ID                         `json:"empire_id"`
	TechnologyID      int                             `json:"technology_id"`
	TechnologyKey     string                          `json:"technology_key"`
	TechnologyNameKey string                          `json:"technology_name_key"`
	TechFieldID       int                             `json:"tech_field_id"`
	SourceKind        TechnologyGrantSource           `json:"source_kind"`
	UncreativeRepair  *UncreativeResearchRepairResult `json:"uncreative_repair,omitempty"`
}

// GrantTechnology applies a concrete non-research Technology acquisition to
// authoritative state. External acquisition does not itself complete the
// owning TechField. For Uncreative Empires, acquiring the persisted fixed
// application of an incomplete field repairs that fixed choice using the
// authoritative GameState RNG, matching the verified original ordering.
func (r *EconomyResolver) GrantTechnology(state *core.GameState, empireID core.ID, technologyID int, options TechnologyGrantOptions) (DomainEvent, error) {
	if r == nil || r.Rules == nil {
		return DomainEvent{}, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return DomainEvent{}, fmt.Errorf("game state must not be nil")
	}
	if options.SourceKind == "" {
		return DomainEvent{}, fmt.Errorf("technology grant source_kind must not be empty")
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("unknown empire %d", empireID)
	}
	if technologyID < 1 || technologyID > 203 {
		return DomainEvent{}, fmt.Errorf("technology id %d outside [1,203]", technologyID)
	}
	if empire.Research != nil && containsInt(empire.Research.TechnologyIDs, technologyID) {
		return DomainEvent{}, fmt.Errorf("cannot externally grant Technology %d while empire %d actively researches it: active-project acquisition semantics are not normalized", technologyID, empireID)
	}
	if containsInt(empire.KnownTechnologyIDs, technologyID) {
		return DomainEvent{}, fmt.Errorf("empire %d already knows Technology %d", empireID, technologyID)
	}
	technologyKey, ok := r.Rules.TechnologyKeyByID[technologyID]
	if !ok || technologyKey == "" {
		return DomainEvent{}, fmt.Errorf("Technology %d has no stable internal key", technologyID)
	}
	technologyNameKey, ok := r.Rules.TechnologyNameKeyByID[technologyID]
	if !ok || technologyNameKey == "" {
		return DomainEvent{}, fmt.Errorf("Technology %d has no name key", technologyID)
	}
	fieldID, ok := r.Rules.TechnologyFieldByID[technologyID]
	if !ok {
		return DomainEvent{}, fmt.Errorf("Technology %d has no normalized TechField mapping", technologyID)
	}

	// Work on a detached Empire copy so a failed repair cannot leave partial
	// ownership changes behind when this transition is called below Session.
	updated := *empire
	updated.KnownTechnologyIDs = append([]int(nil), empire.KnownTechnologyIDs...)
	updated.UncreativeResearchChoices = append([]core.FixedResearchChoice(nil), empire.UncreativeResearchChoices...)
	updated.KnownTechnologyIDs = appendUniqueSortedInt(updated.KnownTechnologyIDs, technologyID)

	rng := state.RNG()
	var repair *UncreativeResearchRepairResult
	modifiers, ok := r.Rules.RaceModifiers[updated.RaceID]
	if !ok {
		return DomainEvent{}, fmt.Errorf("empire %d has unknown race %q", empireID, updated.RaceID)
	}
	if modifiers.Uncreative && fieldID > 0 && fieldID <= 82 && !containsInt(updated.KnownTechnologyFieldIDs, fieldID) {
		result, err := r.Rules.RepairUncreativeResearchChoiceAfterAcquisition(
			&updated,
			technologyID,
			UncreativeResearchRepairOptions{
				StrategicCombat:      options.StrategicCombat,
				RandomEventsDisabled: options.RandomEventsDisabled,
			},
			rng,
		)
		if err != nil {
			return DomainEvent{}, err
		}
		if result.Changed {
			copy := result
			repair = &copy
		}
	}

	*empire = updated
	state.CommitRNG(rng)
	return NewDomainEvent(EventTechnologyGranted, 0, 0, TechnologyGrantedEvent{
		EmpireID:          empireID,
		TechnologyID:      technologyID,
		TechnologyKey:     technologyKey,
		TechnologyNameKey: technologyNameKey,
		TechFieldID:       fieldID,
		SourceKind:        options.SourceKind,
		UncreativeRepair:  repair,
	})
}
