package game

import (
	"fmt"
	"math"

	"moox/internal/core"
)

func applyOriginGrowth(population *core.PopulationState, ownerEmpireID core.ID, origin core.PopulationOriginDynamics) error {
	if population == nil || origin.ProjectedGrowth <= populationEpsilon {
		return nil
	}
	farmers, workers, scientists := 0.0, 0.0, 0.0
	for _, cohort := range population.Cohorts {
		if cohort.OriginEmpireID != origin.OriginEmpireID {
			continue
		}
		farmers += cohort.Farmers
		workers += cohort.Workers
		scientists += cohort.Scientists
	}
	total := farmers + workers + scientists
	if total <= populationEpsilon {
		return fmt.Errorf("population origin %d has no current population for growth", origin.OriginEmpireID)
	}
	key := core.PopulationCohortKey{
		OriginEmpireID:    origin.OriginEmpireID,
		LoyaltyEmpireID:   ownerEmpireID,
		AssimilationState: core.PopulationAssimilated,
	}
	growth := origin.ProjectedGrowth
	parts := []struct {
		job    core.PopulationJob
		amount float64
	}{
		{core.PopulationJobFarmer, growth * farmers / total},
		{core.PopulationJobWorker, growth * workers / total},
		{core.PopulationJobScientist, growth * scientists / total},
	}
	for _, part := range parts {
		if part.amount <= populationEpsilon {
			continue
		}
		if err := population.AddToCohortJob(key, part.job, part.amount); err != nil {
			return err
		}
	}
	return nil
}

func applyOriginStarvation(population *core.PopulationState, originEmpireID core.ID, loss float64) error {
	if population == nil || loss <= populationEpsilon {
		return nil
	}
	workers, scientists, farmers := 0.0, 0.0, 0.0
	for _, cohort := range population.Cohorts {
		if cohort.OriginEmpireID != originEmpireID {
			continue
		}
		workers += cohort.Workers
		scientists += cohort.Scientists
		farmers += cohort.Farmers
	}
	remaining := math.Min(loss, workers+scientists+farmers)
	nonFarmers := workers + scientists
	if nonFarmers > populationEpsilon && remaining > populationEpsilon {
		remove := math.Min(remaining, nonFarmers)
		workerLoss := remove * workers / nonFarmers
		scientistLoss := remove - workerLoss
		if err := removeOriginJobProportionally(population, originEmpireID, core.PopulationJobWorker, workerLoss); err != nil {
			return err
		}
		if err := removeOriginJobProportionally(population, originEmpireID, core.PopulationJobScientist, scientistLoss); err != nil {
			return err
		}
		remaining -= remove
	}
	if remaining > populationEpsilon {
		if err := removeOriginJobProportionally(population, originEmpireID, core.PopulationJobFarmer, remaining); err != nil {
			return err
		}
	}
	population.Normalize()
	return nil
}

func removeOriginJobProportionally(population *core.PopulationState, originEmpireID core.ID, job core.PopulationJob, amount float64) error {
	if amount <= populationEpsilon {
		return nil
	}
	total := 0.0
	for _, cohort := range population.Cohorts {
		if cohort.OriginEmpireID != originEmpireID {
			continue
		}
		value, ok := population.JobAmount(cohort.Key(), job)
		if ok {
			total += value
		}
	}
	if amount > total+populationEpsilon {
		return fmt.Errorf("origin %d job %s has %g population, cannot remove %g", originEmpireID, job, total, amount)
	}
	if total <= populationEpsilon {
		return nil
	}
	remaining := amount
	indexes := make([]int, 0)
	for i := range population.Cohorts {
		if population.Cohorts[i].OriginEmpireID == originEmpireID {
			indexes = append(indexes, i)
		}
	}
	for n, index := range indexes {
		cohort := &population.Cohorts[index]
		current, err := cohortJobValue(*cohort, job)
		if err != nil {
			return err
		}
		if current <= populationEpsilon {
			continue
		}
		remove := amount * current / total
		if n == len(indexes)-1 || remove > remaining {
			remove = remaining
		}
		if remove > current {
			remove = current
		}
		setCohortJobValue(cohort, job, current-remove)
		remaining -= remove
		if remaining <= populationEpsilon {
			break
		}
	}
	population.Normalize()
	return nil
}

func cohortJobValue(cohort core.PopulationCohort, job core.PopulationJob) (float64, error) {
	switch job {
	case core.PopulationJobFarmer:
		return cohort.Farmers, nil
	case core.PopulationJobWorker:
		return cohort.Workers, nil
	case core.PopulationJobScientist:
		return cohort.Scientists, nil
	default:
		return 0, fmt.Errorf("unknown population job %q", job)
	}
}

func setCohortJobValue(cohort *core.PopulationCohort, job core.PopulationJob, value float64) {
	value = math.Max(0, value)
	switch job {
	case core.PopulationJobFarmer:
		cohort.Farmers = value
	case core.PopulationJobWorker:
		cohort.Workers = value
	case core.PopulationJobScientist:
		cohort.Scientists = value
	}
}
