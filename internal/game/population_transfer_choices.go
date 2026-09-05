package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

type PopulationTransferChoice struct {
	SourceColonyID      core.ID                  `json:"source_colony_id"`
	DestinationColonyID core.ID                  `json:"destination_colony_id"`
	Cohort              core.PopulationCohortKey `json:"cohort"`
	SourceJob           core.PopulationJob       `json:"source_job"`
	DestinationJob      core.PopulationJob       `json:"destination_job"`
	Amount              float64                  `json:"amount"`
	SameSystem          bool                     `json:"same_system"`
	ETA                 int                      `json:"eta"`
	FreightersRequired  int                      `json:"freighters_required"`
}

func (r *EconomyResolver) AvailablePopulationTransferChoices(state *core.GameState, empireID core.ID) ([]PopulationTransferChoice, error) {
	if r == nil || r.Rules == nil {
		return nil, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return nil, fmt.Errorf("unknown empire %d", empireID)
	}
	jobs := []core.PopulationJob{core.PopulationJobFarmer, core.PopulationJobWorker, core.PopulationJobScientist}
	active := activePopulationTransfers(state, empireID)
	canLaunchInterstellar := active < maxActivePopulationTransfers && empire.Freighters >= (active+1)*populationTransferFreighters
	var choices []PopulationTransferChoice

	for di := range state.Colonies {
		destination := &state.Colonies[di]
		if destination.EmpireID != empireID {
			continue
		}
		if err := r.recalculateColony(state, destination); err != nil {
			return nil, err
		}
		capacityPopulation := clonePopulationState(destination.Population)
		for _, inbound := range state.PopulationTransfers {
			if inbound.EmpireID != empireID || inbound.DestinationColonyID != destination.ID {
				continue
			}
			if err := capacityPopulation.AddToCohortJob(inbound.CohortKey(), inbound.DestinationPopulationJob(), 1); err != nil {
				return nil, err
			}
		}
		destinationSystem := systemForPlanetID(state, destination.PlanetID)
		if destinationSystem == nil {
			return nil, fmt.Errorf("destination colony %d is not attached to a star system", destination.ID)
		}

		for si := range state.Colonies {
			source := &state.Colonies[si]
			if source.EmpireID != empireID || source.ID == destination.ID || source.Population.Total() < 2-populationEpsilon {
				continue
			}
			sourceSystem := systemForPlanetID(state, source.PlanetID)
			if sourceSystem == nil {
				return nil, fmt.Errorf("source colony %d is not attached to a star system", source.ID)
			}
			sameSystem := sourceSystem.ID == destinationSystem.ID
			if !sameSystem && !canLaunchInterstellar {
				continue
			}
			eta := 0
			freighters := 0
			if !sameSystem {
				eta = populationTransferETA(*sourceSystem, *destinationSystem, r.Rules.populationTransferFTLSpeed(*empire))
				if eta < 1 {
					continue
				}
				freighters = populationTransferFreighters
			}
			for _, cohort := range source.Population.Cohorts {
				key := cohort.Key()
				if key.AssimilationState != core.PopulationAssimilated || key.LoyaltyEmpireID != empireID {
					continue
				}
				for _, sourceJob := range jobs {
					amount, ok := source.Population.JobAmount(key, sourceJob)
					if !ok || amount < 1-populationEpsilon {
						continue
					}
					for _, destinationJob := range jobs {
						canAdd, err := r.populationCanAdd(state, *destination, capacityPopulation, key, destinationJob, 1)
						if err != nil {
							return nil, err
						}
						if !canAdd {
							continue
						}
						choices = append(choices, PopulationTransferChoice{
							SourceColonyID:      source.ID,
							DestinationColonyID: destination.ID,
							Cohort:              key,
							SourceJob:           sourceJob,
							DestinationJob:      destinationJob,
							Amount:              1,
							SameSystem:          sameSystem,
							ETA:                 eta,
							FreightersRequired:  freighters,
						})
					}
				}
			}
		}
	}

	sort.Slice(choices, func(i, j int) bool {
		a, b := choices[i], choices[j]
		if a.SourceColonyID != b.SourceColonyID {
			return a.SourceColonyID < b.SourceColonyID
		}
		if a.DestinationColonyID != b.DestinationColonyID {
			return a.DestinationColonyID < b.DestinationColonyID
		}
		if a.Cohort.OriginEmpireID != b.Cohort.OriginEmpireID {
			return a.Cohort.OriginEmpireID < b.Cohort.OriginEmpireID
		}
		if a.Cohort.LoyaltyEmpireID != b.Cohort.LoyaltyEmpireID {
			return a.Cohort.LoyaltyEmpireID < b.Cohort.LoyaltyEmpireID
		}
		if a.SourceJob != b.SourceJob {
			return a.SourceJob < b.SourceJob
		}
		return a.DestinationJob < b.DestinationJob
	})
	return choices, nil
}
