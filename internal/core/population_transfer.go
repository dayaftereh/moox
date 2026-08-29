package core

type PopulationTransfer struct {
	ID                  ID                          `json:"id"`
	EmpireID            ID                          `json:"empire_id"`
	SourceColonyID      ID                          `json:"source_colony_id"`
	DestinationColonyID ID                          `json:"destination_colony_id"`
	OriginEmpireID      ID                          `json:"origin_empire_id"`
	LoyaltyEmpireID     ID                          `json:"loyalty_empire_id"`
	AssimilationState   PopulationAssimilationState `json:"assimilation_state"`
	Job                 PopulationJob               `json:"job"`
	RemainingTurns      int                         `json:"remaining_turns"`
}

func (t PopulationTransfer) CohortKey() PopulationCohortKey {
	return PopulationCohortKey{
		OriginEmpireID:    t.OriginEmpireID,
		LoyaltyEmpireID:   t.LoyaltyEmpireID,
		AssimilationState: t.AssimilationState,
	}
}
