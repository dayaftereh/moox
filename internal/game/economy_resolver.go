package game

import (
	"fmt"
	"math"

	"moox/internal/core"
	"moox/internal/protocol"
)

type EconomyResolver struct {
	Rules *EconomyRules
}

type PopulationAssignedEvent struct {
	ColonyID           core.ID                       `json:"colony_id"`
	Previous           core.PopulationState          `json:"previous"`
	Current            core.PopulationState          `json:"current"`
	BaseEconomy        core.ColonyEconomy            `json:"base_economy"`
	EconomyContext     core.ColonyEconomyContext     `json:"economy_context"`
	AdjustedEconomy    core.ColonyEconomy            `json:"adjusted_economy"`
	PopulationDynamics core.ColonyPopulationDynamics `json:"population_dynamics"`
}

func NewEconomyResolver(rules *EconomyRules) (*EconomyResolver, error) {
	if rules == nil {
		return nil, fmt.Errorf("economy rules must not be nil")
	}
	return &EconomyResolver{Rules: rules}, nil
}

func (r *EconomyResolver) Resolve(ctx ResolveContext, state *core.GameState, batches []protocol.CommandBatch) (Resolution, error) {
	if r == nil || r.Rules == nil {
		return Resolution{}, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return Resolution{}, fmt.Errorf("game state must not be nil")
	}

	var events []DomainEvent
	for _, batch := range batches {
		empireID, ok := ctx.EmpireForSeat(batch.SeatID)
		if !ok {
			return Resolution{}, fmt.Errorf("seat %d has no server authority", batch.SeatID)
		}
		for _, command := range batch.Commands {
			switch command.Kind {
			case CommandAssignPopulation:
				payload, err := decodeAssignPopulation(command)
				if err != nil {
					return Resolution{}, fmt.Errorf("seat %d command %d: %w", batch.SeatID, command.Sequence, err)
				}
				colony := colonyByID(state, payload.ColonyID)
				if colony == nil {
					return Resolution{}, fmt.Errorf("seat %d command %d references unknown colony %d", batch.SeatID, command.Sequence, payload.ColonyID)
				}
				if colony.EmpireID != empireID {
					return Resolution{}, fmt.Errorf("seat %d cannot assign population on colony %d owned by empire %d", batch.SeatID, colony.ID, colony.EmpireID)
				}
				assigned := payload.Farmers + payload.Workers + payload.Scientists
				if math.Abs(assigned-colony.Population.Total()) > 1e-9*math.Max(1, math.Max(math.Abs(assigned), math.Abs(colony.Population.Total()))) {
					return Resolution{}, fmt.Errorf("colony %d assignment total %g does not equal population total %g", colony.ID, assigned, colony.Population.Total())
				}
				previous := colony.Population
				if err := colony.Population.SetAggregateJobs(payload.Farmers, payload.Workers, payload.Scientists); err != nil {
					return Resolution{}, fmt.Errorf("colony %d assignment: %w", colony.ID, err)
				}
				if err := r.recalculateColony(state, colony); err != nil {
					return Resolution{}, fmt.Errorf("seat %d command %d: %w", batch.SeatID, command.Sequence, err)
				}
				event, err := NewDomainEvent("colony.population_assigned", batch.SeatID, command.Sequence, PopulationAssignedEvent{
					ColonyID:           colony.ID,
					Previous:           previous,
					Current:            colony.Population,
					BaseEconomy:        colony.Economy,
					EconomyContext:     colony.EconomyContext,
					AdjustedEconomy:    colony.AdjustedEconomy,
					PopulationDynamics: colony.PopulationDynamics,
				})
				if err != nil {
					return Resolution{}, err
				}
				events = append(events, event)
			case CommandQueueBuilding:
				event, err := r.queueBuilding(state, empireID, batch.SeatID, command)
				if err != nil {
					return Resolution{}, fmt.Errorf("seat %d command %d: %w", batch.SeatID, command.Sequence, err)
				}
				events = append(events, event)
			case CommandQueueColonyShip:
				event, err := r.queueColonyShip(state, empireID, batch.SeatID, command)
				if err != nil {
					return Resolution{}, fmt.Errorf("seat %d command %d: %w", batch.SeatID, command.Sequence, err)
				}
				events = append(events, event)
			case CommandQueueHousing:
				event, err := r.queueHousing(state, empireID, batch.SeatID, command)
				if err != nil {
					return Resolution{}, fmt.Errorf("seat %d command %d: %w", batch.SeatID, command.Sequence, err)
				}
				events = append(events, event)
			case CommandQueuePlanetaryTransformation:
				event, err := r.queuePlanetaryTransformation(state, empireID, batch.SeatID, command)
				if err != nil {
					return Resolution{}, fmt.Errorf("seat %d command %d: %w", batch.SeatID, command.Sequence, err)
				}
				events = append(events, event)
			case CommandQueueFreighterFleet:
				event, err := r.queueFreighterFleet(state, empireID, batch.SeatID, command)
				if err != nil {
					return Resolution{}, fmt.Errorf("seat %d command %d: %w", batch.SeatID, command.Sequence, err)
				}
				events = append(events, event)
			case CommandTransferPopulation:
				event, err := r.transferPopulation(state, empireID, batch.SeatID, command)
				if err != nil {
					return Resolution{}, fmt.Errorf("seat %d command %d: %w", batch.SeatID, command.Sequence, err)
				}
				events = append(events, event)
			case CommandMoveFleet:
				event, err := r.moveFleet(state, empireID, batch.SeatID, command)
				if err != nil {
					return Resolution{}, fmt.Errorf("seat %d command %d: %w", batch.SeatID, command.Sequence, err)
				}
				events = append(events, event)
			case CommandColonizePlanet:
				colonizationEvents, err := r.colonizePlanet(state, empireID, batch.SeatID, command)
				if err != nil {
					return Resolution{}, fmt.Errorf("seat %d command %d: %w", batch.SeatID, command.Sequence, err)
				}
				events = append(events, colonizationEvents...)
			case CommandSelectResearch:
				event, err := r.selectResearch(state, empireID, batch.SeatID, command)
				if err != nil {
					return Resolution{}, fmt.Errorf("seat %d command %d: %w", batch.SeatID, command.Sequence, err)
				}
				events = append(events, event)
			default:
				return Resolution{}, fmt.Errorf("seat %d command %d has unsupported strategic command kind %q", batch.SeatID, command.Sequence, command.Kind)
			}
		}
	}

	// Keep the materialized base-economy snapshot consistent even for colonies
	// with no assignment command in this turn.
	for i := range state.Colonies {
		if err := r.recalculateColony(state, &state.Colonies[i]); err != nil {
			return Resolution{}, err
		}
	}
	foodEvents, err := r.materializeFoodLogistics(state, true)
	if err != nil {
		return Resolution{}, err
	}
	events = append(events, foodEvents...)
	treasuryEvents, err := r.settleTreasury(state)
	if err != nil {
		return Resolution{}, err
	}
	events = append(events, treasuryEvents...)
	// Original 1.31 keeps current-turn colony resource values materialized before
	// Next_Turn_Calc applies state changes. Player research consumes that snapshot
	// first; Population growth/starvation is then applied; Construction consumes
	// the already-materialized pre-growth PP snapshot afterwards. None of these
	// phases retroactively recompute current-turn RP/PP from newly changed Population.
	researchEvents, err := r.advanceResearch(state)
	if err != nil {
		return Resolution{}, err
	}
	events = append(events, researchEvents...)
	populationEvents, err := r.advancePopulation(state)
	if err != nil {
		return Resolution{}, err
	}
	events = append(events, populationEvents...)
	fleetTransitEvents, err := r.advanceStrategicFleetTransit(state)
	if err != nil {
		return Resolution{}, err
	}
	events = append(events, fleetTransitEvents...)
	constructionEvents, err := r.advanceConstruction(state)
	if err != nil {
		return Resolution{}, err
	}
	events = append(events, constructionEvents...)
	// Original 1.31 recomputes strategic blockades after the first Colony pass and
	// immediately before Settler movement. Rebuild the derived system blockade state
	// here so Population-transfer arrivals and the final Food snapshot consume the same
	// current Fleet/Diplomacy result.
	if err := recomputeSystemBlockades(state); err != nil {
		return Resolution{}, err
	}
	populationTransferEvents, err := r.advancePopulationTransfers(state)
	if err != nil {
		return Resolution{}, err
	}
	events = append(events, populationTransferEvents...)
	// Recalculate the next-state local economy only after the original-order
	// apply phases, then rematerialize logistics without emitting a second turn event.
	for i := range state.Colonies {
		if err := r.recalculateColony(state, &state.Colonies[i]); err != nil {
			return Resolution{}, err
		}
	}
	if _, err := r.materializeFoodLogistics(state, false); err != nil {
		return Resolution{}, err
	}
	return Resolution{State: state, Events: events}, nil
}

func (r *EconomyResolver) recalculateColony(state *core.GameState, colony *core.Colony) error {
	planet := planetByID(state, colony.PlanetID)
	if planet == nil {
		return fmt.Errorf("colony %d references unknown planet %d", colony.ID, colony.PlanetID)
	}
	empire := empireByID(state, colony.EmpireID)
	if empire == nil {
		return fmt.Errorf("colony %d references unknown empire %d", colony.ID, colony.EmpireID)
	}
	base, context, adjusted, err := r.calculateRaceAwareColonyEconomy(state, *colony, *planet, *empire)
	if err != nil {
		return fmt.Errorf("calculate colony %d race-aware economy: %w", colony.ID, err)
	}
	dynamics, err := r.calculateRaceAwarePopulationDynamics(state, *colony, *planet, *empire, adjusted)
	if err != nil {
		return fmt.Errorf("calculate colony %d race-aware population dynamics: %w", colony.ID, err)
	}
	colony.Economy = base
	colony.EconomyContext = context
	colony.AdjustedEconomy = adjusted
	colony.PopulationDynamics = dynamics
	return nil
}

func colonyByID(state *core.GameState, id core.ID) *core.Colony {
	for i := range state.Colonies {
		if state.Colonies[i].ID == id {
			return &state.Colonies[i]
		}
	}
	return nil
}

func empireByID(state *core.GameState, id core.ID) *core.Empire {
	for i := range state.Empires {
		if state.Empires[i].ID == id {
			return &state.Empires[i]
		}
	}
	return nil
}

func planetByID(state *core.GameState, id core.ID) *core.Planet {
	for si := range state.Galaxy.Systems {
		for pi := range state.Galaxy.Systems[si].Planets {
			if state.Galaxy.Systems[si].Planets[pi].ID == id {
				return &state.Galaxy.Systems[si].Planets[pi]
			}
		}
	}
	return nil
}
