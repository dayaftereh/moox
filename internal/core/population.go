package core

import (
	"fmt"
	"math"
	"sort"
)

const populationStateEpsilon = 1e-9

type PopulationAssimilationState string

const (
	PopulationAssimilated PopulationAssimilationState = "assimilated"
	PopulationConquered   PopulationAssimilationState = "conquered"
)

type PopulationJob string

const (
	PopulationJobFarmer    PopulationJob = "farmer"
	PopulationJobWorker    PopulationJob = "worker"
	PopulationJobScientist PopulationJob = "scientist"
)

type PopulationCohort struct {
	OriginEmpireID    ID                          `json:"origin_empire_id"`
	LoyaltyEmpireID   ID                          `json:"loyalty_empire_id"`
	AssimilationState PopulationAssimilationState `json:"assimilation_state"`
	Farmers           float64                     `json:"farmers"`
	Workers           float64                     `json:"workers"`
	Scientists        float64                     `json:"scientists"`
}

type PopulationCohortKey struct {
	OriginEmpireID    ID                          `json:"origin_empire_id"`
	LoyaltyEmpireID   ID                          `json:"loyalty_empire_id"`
	AssimilationState PopulationAssimilationState `json:"assimilation_state"`
}

type PopulationState struct {
	Cohorts []PopulationCohort `json:"cohorts,omitempty"`
}

func NewAssimilatedPopulation(empireID ID, farmers, workers, scientists float64) PopulationState {
	p := PopulationState{Cohorts: []PopulationCohort{{
		OriginEmpireID:    empireID,
		LoyaltyEmpireID:   empireID,
		AssimilationState: PopulationAssimilated,
		Farmers:           farmers,
		Workers:           workers,
		Scientists:        scientists,
	}}}
	p.Normalize()
	return p
}

func (c PopulationCohort) Key() PopulationCohortKey {
	return PopulationCohortKey{OriginEmpireID: c.OriginEmpireID, LoyaltyEmpireID: c.LoyaltyEmpireID, AssimilationState: c.AssimilationState}
}

func (c PopulationCohort) Total() float64 { return c.Farmers + c.Workers + c.Scientists }

func (p PopulationState) Total() float64 {
	total := 0.0
	for _, c := range p.Cohorts {
		total += c.Total()
	}
	return total
}

func (p PopulationState) Farmers() float64 {
	total := 0.0
	for _, c := range p.Cohorts {
		total += c.Farmers
	}
	return total
}

func (p PopulationState) Workers() float64 {
	total := 0.0
	for _, c := range p.Cohorts {
		total += c.Workers
	}
	return total
}

func (p PopulationState) Scientists() float64 {
	total := 0.0
	for _, c := range p.Cohorts {
		total += c.Scientists
	}
	return total
}

func (p PopulationState) IsZero() bool { return p.Total() <= populationStateEpsilon }

func (p PopulationState) TotalsByOrigin() map[ID]float64 {
	out := make(map[ID]float64)
	for _, c := range p.Cohorts {
		out[c.OriginEmpireID] += c.Total()
	}
	return out
}

func (p PopulationState) CohortBySemanticKey(key PopulationCohortKey) (PopulationCohort, bool) {
	for _, c := range p.Cohorts {
		if c.Key() == key {
			return c, true
		}
	}
	return PopulationCohort{}, false
}

func (p *PopulationState) Normalize() {
	if p == nil {
		return
	}
	merged := make(map[PopulationCohortKey]PopulationCohort, len(p.Cohorts))
	for _, c := range p.Cohorts {
		if math.Abs(c.Farmers) <= populationStateEpsilon {
			c.Farmers = 0
		}
		if math.Abs(c.Workers) <= populationStateEpsilon {
			c.Workers = 0
		}
		if math.Abs(c.Scientists) <= populationStateEpsilon {
			c.Scientists = 0
		}
		if c.Total() <= populationStateEpsilon {
			continue
		}
		key := c.Key()
		current := merged[key]
		if current.OriginEmpireID == 0 {
			current = PopulationCohort{OriginEmpireID: key.OriginEmpireID, LoyaltyEmpireID: key.LoyaltyEmpireID, AssimilationState: key.AssimilationState}
		}
		current.Farmers += c.Farmers
		current.Workers += c.Workers
		current.Scientists += c.Scientists
		merged[key] = current
	}
	p.Cohorts = p.Cohorts[:0]
	for _, c := range merged {
		p.Cohorts = append(p.Cohorts, c)
	}
	sort.Slice(p.Cohorts, func(i, j int) bool {
		a, b := p.Cohorts[i], p.Cohorts[j]
		if a.OriginEmpireID != b.OriginEmpireID {
			return a.OriginEmpireID < b.OriginEmpireID
		}
		if a.LoyaltyEmpireID != b.LoyaltyEmpireID {
			return a.LoyaltyEmpireID < b.LoyaltyEmpireID
		}
		return assimilationOrder(a.AssimilationState) < assimilationOrder(b.AssimilationState)
	})
}

func assimilationOrder(state PopulationAssimilationState) int {
	if state == PopulationAssimilated {
		return 0
	}
	if state == PopulationConquered {
		return 1
	}
	return 2
}

func (p PopulationState) Validate(ownerEmpireID ID, empireExists func(ID) bool) error {
	seen := make(map[PopulationCohortKey]struct{}, len(p.Cohorts))
	for i, c := range p.Cohorts {
		if c.OriginEmpireID == 0 || c.LoyaltyEmpireID == 0 {
			return fmt.Errorf("cohort[%d] origin and loyalty empire IDs are required", i)
		}
		if empireExists != nil && (!empireExists(c.OriginEmpireID) || !empireExists(c.LoyaltyEmpireID)) {
			return fmt.Errorf("cohort[%d] references unknown origin/loyalty empire", i)
		}
		if c.AssimilationState != PopulationAssimilated && c.AssimilationState != PopulationConquered {
			return fmt.Errorf("cohort[%d] assimilation state %q is invalid", i, c.AssimilationState)
		}
		if c.AssimilationState == PopulationConquered && c.OriginEmpireID == ownerEmpireID {
			return fmt.Errorf("cohort[%d] owner-origin population cannot be conquered", i)
		}
		if c.AssimilationState == PopulationConquered && c.LoyaltyEmpireID != c.OriginEmpireID {
			return fmt.Errorf("cohort[%d] conquered population must retain origin loyalty", i)
		}
		if c.AssimilationState == PopulationAssimilated && c.LoyaltyEmpireID != ownerEmpireID {
			return fmt.Errorf("cohort[%d] assimilated population must be loyal to colony owner", i)
		}
		for _, item := range []struct {
			name  string
			value float64
		}{{"farmers", c.Farmers}, {"workers", c.Workers}, {"scientists", c.Scientists}} {
			if math.IsNaN(item.value) || math.IsInf(item.value, 0) || item.value < 0 {
				return fmt.Errorf("cohort[%d] %s must be finite and non-negative", i, item.name)
			}
		}
		if c.Total() <= populationStateEpsilon {
			return fmt.Errorf("cohort[%d] must contain positive population", i)
		}
		key := c.Key()
		if _, ok := seen[key]; ok {
			return fmt.Errorf("cohort[%d] duplicates semantic key %+v", i, key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

// SetAggregateJobs preserves cohort identities and totals while applying the
// existing colony-wide assignment command deterministically across cohorts.
func (p *PopulationState) SetAggregateJobs(farmers, workers, scientists float64) error {
	if p == nil {
		return fmt.Errorf("population state must not be nil")
	}
	total := p.Total()
	assigned := farmers + workers + scientists
	if math.Abs(assigned-total) > populationStateEpsilon*math.Max(1, math.Max(math.Abs(assigned), math.Abs(total))) {
		return fmt.Errorf("assignment total %g does not equal population total %g", assigned, total)
	}
	if total <= populationStateEpsilon {
		p.Cohorts = nil
		return nil
	}
	remainingFarmers, remainingWorkers, remainingScientists := farmers, workers, scientists
	remainingTotal := total
	for i := range p.Cohorts {
		cohortTotal := p.Cohorts[i].Total()
		if i == len(p.Cohorts)-1 || remainingTotal <= populationStateEpsilon {
			p.Cohorts[i].Farmers = math.Max(0, remainingFarmers)
			p.Cohorts[i].Workers = math.Max(0, remainingWorkers)
			p.Cohorts[i].Scientists = math.Max(0, remainingScientists)
			break
		}
		share := cohortTotal / remainingTotal
		f := remainingFarmers * share
		w := remainingWorkers * share
		s := cohortTotal - f - w
		if s < 0 && s > -populationStateEpsilon {
			s = 0
		}
		p.Cohorts[i].Farmers, p.Cohorts[i].Workers, p.Cohorts[i].Scientists = f, w, s
		remainingFarmers -= f
		remainingWorkers -= w
		remainingScientists -= s
		remainingTotal -= cohortTotal
	}
	p.Normalize()
	return nil
}

func (p *PopulationState) AddToCohortJob(key PopulationCohortKey, job PopulationJob, amount float64) error {
	if p == nil || amount < 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return fmt.Errorf("population addition must be finite and non-negative")
	}
	for i := range p.Cohorts {
		if p.Cohorts[i].Key() == key {
			return addPopulationJob(&p.Cohorts[i], job, amount)
		}
	}
	c := PopulationCohort{OriginEmpireID: key.OriginEmpireID, LoyaltyEmpireID: key.LoyaltyEmpireID, AssimilationState: key.AssimilationState}
	if err := addPopulationJob(&c, job, amount); err != nil {
		return err
	}
	p.Cohorts = append(p.Cohorts, c)
	p.Normalize()
	return nil
}

func (p *PopulationState) RemoveFromCohortJob(key PopulationCohortKey, job PopulationJob, amount float64) error {
	if p == nil || amount < 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return fmt.Errorf("population removal must be finite and non-negative")
	}
	for i := range p.Cohorts {
		if p.Cohorts[i].Key() != key {
			continue
		}
		current, err := populationJobValue(p.Cohorts[i], job)
		if err != nil {
			return err
		}
		if current+populationStateEpsilon < amount {
			return fmt.Errorf("cohort job %s has %g population, cannot remove %g", job, current, amount)
		}
		if err := addPopulationJob(&p.Cohorts[i], job, -amount); err != nil {
			return err
		}
		p.Normalize()
		return nil
	}
	return fmt.Errorf("population cohort %+v not found", key)
}

func (p PopulationState) JobAmount(key PopulationCohortKey, job PopulationJob) (float64, bool) {
	c, ok := p.CohortBySemanticKey(key)
	if !ok {
		return 0, false
	}
	v, err := populationJobValue(c, job)
	return v, err == nil
}

func addPopulationJob(c *PopulationCohort, job PopulationJob, delta float64) error {
	switch job {
	case PopulationJobFarmer:
		c.Farmers += delta
	case PopulationJobWorker:
		c.Workers += delta
	case PopulationJobScientist:
		c.Scientists += delta
	default:
		return fmt.Errorf("unknown population job %q", job)
	}
	return nil
}

func populationJobValue(c PopulationCohort, job PopulationJob) (float64, error) {
	switch job {
	case PopulationJobFarmer:
		return c.Farmers, nil
	case PopulationJobWorker:
		return c.Workers, nil
	case PopulationJobScientist:
		return c.Scientists, nil
	default:
		return 0, fmt.Errorf("unknown population job %q", job)
	}
}
