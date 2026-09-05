package game

import (
	"fmt"
	"math"
	"sort"

	"moox/internal/core"
)

// PopulationChoice is a server-authoritative preview of one legal aggregate
// population assignment. Controllers can rank these choices without copying
// colony economy/food formulas into their policy layer.
type PopulationChoice struct {
	ColonyID           core.ID                       `json:"colony_id"`
	Farmers            float64                       `json:"farmers"`
	Workers            float64                       `json:"workers"`
	Scientists         float64                       `json:"scientists"`
	AdjustedEconomy    core.ColonyEconomy            `json:"adjusted_economy"`
	PopulationDynamics core.ColonyPopulationDynamics `json:"population_dynamics"`
	FoodSafe           bool                          `json:"food_safe"`
}

// FleetMoveChoice is one server-authoritative legal strategic move. ShipIDs is
// empty for a whole/fixed fleet move and carries a combat subset when the move
// would deterministically split a fleet.
type FleetMoveChoice struct {
	FleetID               core.ID   `json:"fleet_id"`
	SourceSystemID        core.ID   `json:"source_system_id"`
	DestinationSystemID   core.ID   `json:"destination_system_id"`
	ETA                   int       `json:"eta"`
	FuelRangeParsecs      int       `json:"fuel_range_parsecs"`
	SupplyDistanceParsecs int       `json:"supply_distance_parsecs"`
	ShipIDs               []core.ID `json:"ship_ids,omitempty"`
}

type ColonizationChoice struct {
	FleetID  core.ID `json:"fleet_id"`
	SystemID core.ID `json:"system_id"`
	PlanetID core.ID `json:"planet_id"`
}

type OutpostDeploymentChoice struct {
	FleetID  core.ID `json:"fleet_id"`
	SystemID core.ID `json:"system_id"`
	BodyID   core.ID `json:"body_id"`
	PlanetID core.ID `json:"planet_id,omitempty"`
}

// ResearchDue reports whether the active selected field has accumulated enough
// RP for the existing authoritative completion path. It does not mutate state.
func (r *EconomyRules) ResearchDue(state *core.GameState, empireID core.ID) (bool, error) {
	if r == nil {
		return false, fmt.Errorf("economy rules must not be nil")
	}
	if state == nil {
		return false, fmt.Errorf("game state must not be nil")
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return false, fmt.Errorf("unknown empire %d", empireID)
	}
	if empire.Research == nil {
		return false, nil
	}
	cost, err := r.researchFieldCostRP(empire, empire.Research.TechFieldID)
	if err != nil {
		return false, err
	}
	return empire.Research.ProgressRP+1e-9 >= cost, nil
}

func (r *EconomyResolver) AvailablePopulationChoices(state *core.GameState, empireID, colonyID core.ID) ([]PopulationChoice, error) {
	if r == nil || r.Rules == nil {
		return nil, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	colony := colonyByID(state, colonyID)
	if colony == nil {
		return nil, fmt.Errorf("unknown colony %d", colonyID)
	}
	if colony.EmpireID != empireID {
		return nil, fmt.Errorf("empire %d does not own colony %d", empireID, colonyID)
	}
	total := colony.Population.Total()
	if total < 0 || math.IsNaN(total) || math.IsInf(total, 0) {
		return nil, fmt.Errorf("colony %d has invalid population total %g", colonyID, total)
	}

	type jobs struct{ f, w, s float64 }
	candidates := []jobs{{colony.Population.Farmers(), colony.Population.Workers(), colony.Population.Scientists()}}
	maxFarmers := int(math.Ceil(total - 1e-12))
	for i := 0; i <= maxFarmers; i++ {
		farmers := math.Min(float64(i), total)
		remaining := total - farmers
		if remaining < 0 && remaining > -1e-9 {
			remaining = 0
		}
		candidates = append(candidates, jobs{farmers, remaining, 0}, jobs{farmers, 0, remaining})
	}

	seen := make(map[string]struct{}, len(candidates))
	choices := make([]PopulationChoice, 0, len(candidates))
	for _, candidateJobs := range candidates {
		key := fmt.Sprintf("%.9f/%.9f/%.9f", candidateJobs.f, candidateJobs.w, candidateJobs.s)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		candidate := *colony
		candidate.Population.Cohorts = append([]core.PopulationCohort(nil), colony.Population.Cohorts...)
		if err := candidate.Population.SetAggregateJobs(candidateJobs.f, candidateJobs.w, candidateJobs.s); err != nil {
			return nil, fmt.Errorf("preview colony %d population: %w", colonyID, err)
		}
		if err := r.recalculateColony(state, &candidate); err != nil {
			return nil, fmt.Errorf("preview colony %d economy: %w", colonyID, err)
		}
		choices = append(choices, PopulationChoice{
			ColonyID:           colonyID,
			Farmers:            candidateJobs.f,
			Workers:            candidateJobs.w,
			Scientists:         candidateJobs.s,
			AdjustedEconomy:    candidate.AdjustedEconomy,
			PopulationDynamics: candidate.PopulationDynamics,
			FoodSafe:           candidate.AdjustedEconomy.Food+1e-9 >= candidate.PopulationDynamics.FoodRequired,
		})
	}
	sort.Slice(choices, func(i, j int) bool {
		if choices[i].Farmers != choices[j].Farmers {
			return choices[i].Farmers < choices[j].Farmers
		}
		if choices[i].Workers != choices[j].Workers {
			return choices[i].Workers < choices[j].Workers
		}
		return choices[i].Scientists < choices[j].Scientists
	})
	return choices, nil
}

func (r *EconomyResolver) AvailableFleetMoveChoices(state *core.GameState, empireID core.ID) ([]FleetMoveChoice, error) {
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
	type profile struct {
		ftlSpeed  int
		fuelRange int
		shipIDs   []core.ID
	}
	choices := make([]FleetMoveChoice, 0)
	for i := range state.StrategicFleets {
		fleet := &state.StrategicFleets[i]
		if fleet.EmpireID != empireID || fleet.AtSystemID == 0 || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 {
			continue
		}
		source := systemByID(state, fleet.AtSystemID)
		if source == nil {
			return nil, fmt.Errorf("fleet %d references unknown source system %d", fleet.ID, fleet.AtSystemID)
		}
		profiles := make([]profile, 0, len(fleet.ShipIDs)+1)
		if _, fixed := fixedSpecialShipName(fleet.SpecialKind); fixed {
			if fleet.FTLSpeed < 2 {
				continue
			}
			profiles = append(profiles, profile{ftlSpeed: fleet.FTLSpeed, fuelRange: colonyShipFuelRangeParsecs(*empire)})
		} else if fleet.Role == core.StrategicFleetRoleCombat && fleet.SpecialKind == core.StrategicFleetSpecialNone {
			ftlSpeed, fuelRange, err := r.combatFleetMovementProfile(state, *fleet, empire)
			if err == nil {
				profiles = append(profiles, profile{ftlSpeed: ftlSpeed, fuelRange: fuelRange})
			}
			if len(fleet.ShipIDs) > 1 {
				for _, shipID := range fleet.ShipIDs {
					one := *fleet
					one.ShipIDs = []core.ID{shipID}
					ftlSpeed, fuelRange, err := r.combatFleetMovementProfile(state, one, empire)
					if err != nil {
						continue
					}
					profiles = append(profiles, profile{ftlSpeed: ftlSpeed, fuelRange: fuelRange, shipIDs: []core.ID{shipID}})
				}
			}
		} else {
			continue
		}
		for si := range state.Galaxy.Systems {
			destination := &state.Galaxy.Systems[si]
			if destination.ID == source.ID {
				continue
			}
			supplyDistance, ok := nearestEmpireSupplyDistanceParsecs(state, empireID, *destination)
			if !ok {
				continue
			}
			for _, movement := range profiles {
				if supplyDistance > movement.fuelRange {
					continue
				}
				eta := strategicTravelETA(*source, *destination, movement.ftlSpeed)
				if eta < 1 {
					continue
				}
				choices = append(choices, FleetMoveChoice{
					FleetID: fleet.ID, SourceSystemID: source.ID, DestinationSystemID: destination.ID,
					ETA: eta, FuelRangeParsecs: movement.fuelRange, SupplyDistanceParsecs: supplyDistance,
					ShipIDs: append([]core.ID(nil), movement.shipIDs...),
				})
			}
		}
	}
	sort.Slice(choices, func(i, j int) bool {
		if choices[i].FleetID != choices[j].FleetID {
			return choices[i].FleetID < choices[j].FleetID
		}
		if choices[i].DestinationSystemID != choices[j].DestinationSystemID {
			return choices[i].DestinationSystemID < choices[j].DestinationSystemID
		}
		if len(choices[i].ShipIDs) != len(choices[j].ShipIDs) {
			return len(choices[i].ShipIDs) < len(choices[j].ShipIDs)
		}
		for k := range choices[i].ShipIDs {
			if choices[i].ShipIDs[k] != choices[j].ShipIDs[k] {
				return choices[i].ShipIDs[k] < choices[j].ShipIDs[k]
			}
		}
		return false
	})
	return choices, nil
}

func (r *EconomyResolver) AvailableColonizationChoices(state *core.GameState, empireID core.ID) ([]ColonizationChoice, error) {
	if r == nil || r.Rules == nil {
		return nil, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	if empireByID(state, empireID) == nil {
		return nil, fmt.Errorf("unknown empire %d", empireID)
	}
	var choices []ColonizationChoice
	for _, fleet := range state.StrategicFleets {
		if fleet.EmpireID != empireID || fleet.SpecialKind != core.StrategicFleetSpecialColonyShip || fleet.AtSystemID == 0 || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 {
			continue
		}
		system := systemByID(state, fleet.AtSystemID)
		if system == nil {
			return nil, fmt.Errorf("fleet %d references unknown system %d", fleet.ID, fleet.AtSystemID)
		}
		for _, planet := range system.Planets {
			if _, _, _, err := r.prepareFoundedColony(state, empireID, planet.ID); err != nil {
				continue
			}
			choices = append(choices, ColonizationChoice{FleetID: fleet.ID, SystemID: system.ID, PlanetID: planet.ID})
		}
	}
	sort.Slice(choices, func(i, j int) bool {
		if choices[i].FleetID != choices[j].FleetID {
			return choices[i].FleetID < choices[j].FleetID
		}
		return choices[i].PlanetID < choices[j].PlanetID
	})
	return choices, nil
}

func (r *EconomyResolver) AvailableOutpostDeploymentChoices(state *core.GameState, empireID core.ID) ([]OutpostDeploymentChoice, error) {
	if r == nil || r.Rules == nil {
		return nil, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	if empireByID(state, empireID) == nil {
		return nil, fmt.Errorf("unknown empire %d", empireID)
	}
	var choices []OutpostDeploymentChoice
	for _, fleet := range state.StrategicFleets {
		if fleet.EmpireID != empireID || fleet.SpecialKind != core.StrategicFleetSpecialOutpostShip || fleet.AtSystemID == 0 || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 {
			continue
		}
		system := systemByID(state, fleet.AtSystemID)
		if system == nil {
			return nil, fmt.Errorf("fleet %d references unknown system %d", fleet.ID, fleet.AtSystemID)
		}
		if len(system.Bodies) > 0 {
			for bi := range system.Bodies {
				body := &system.Bodies[bi]
				target := orbitalBodyTargetByID(state, body.ID)
				if orbitalBodyOccupiedByOutpost(state, target) {
					continue
				}
				planetID := core.ID(0)
				if target.Planet != nil {
					planetID = target.Planet.ID
					if target.Planet.ColonyID != 0 || colonyReferencesPlanet(state, target.Planet.ID) {
						continue
					}
				}
				choices = append(choices, OutpostDeploymentChoice{FleetID: fleet.ID, SystemID: system.ID, BodyID: body.ID, PlanetID: planetID})
			}
		} else {
			for pi := range system.Planets {
				planet := &system.Planets[pi]
				if planet.ColonyID != 0 || colonyReferencesPlanet(state, planet.ID) || planet.OutpostID != 0 || outpostReferencesPlanet(state, planet.ID) {
					continue
				}
				choices = append(choices, OutpostDeploymentChoice{FleetID: fleet.ID, SystemID: system.ID, BodyID: planet.ID, PlanetID: planet.ID})
			}
		}
	}
	sort.Slice(choices, func(i, j int) bool {
		if choices[i].FleetID != choices[j].FleetID {
			return choices[i].FleetID < choices[j].FleetID
		}
		if choices[i].SystemID != choices[j].SystemID {
			return choices[i].SystemID < choices[j].SystemID
		}
		return choices[i].BodyID < choices[j].BodyID
	})
	return choices, nil
}
