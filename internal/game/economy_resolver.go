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
				if math.Abs(assigned-colony.Population.Total) > 1e-9*math.Max(1, math.Max(math.Abs(assigned), math.Abs(colony.Population.Total))) {
					return Resolution{}, fmt.Errorf("colony %d assignment total %g does not equal population total %g", colony.ID, assigned, colony.Population.Total)
				}
				previous := colony.Population
				colony.Population.Farmers = payload.Farmers
				colony.Population.Workers = payload.Workers
				colony.Population.Scientists = payload.Scientists
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
	constructionEvents, err := r.advanceConstruction(state)
	if err != nil {
		return Resolution{}, err
	}
	events = append(events, constructionEvents...)
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
	// Current MOOX ordering keeps Population as a turn-end transition until the
	// conflicting secondary evidence about original 1.31 turn ordering is resolved.
	// Recalculate the next-state local economy after Population changes, then
	// rematerialize logistics without emitting a second authoritative turn event.
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
	base, err := r.Rules.CalculateBaseEconomy(*colony, *planet, empire.RaceID)
	if err != nil {
		return fmt.Errorf("calculate colony %d base economy: %w", colony.ID, err)
	}
	context, adjusted, err := r.Rules.CalculateContextualEconomy(base, *colony, *planet, empire.RaceID)
	if err != nil {
		return fmt.Errorf("calculate colony %d contextual economy: %w", colony.ID, err)
	}
	dynamics, err := r.Rules.CalculatePopulationDynamics(*colony, *planet, empire.RaceID, adjusted)
	if err != nil {
		return fmt.Errorf("calculate colony %d population dynamics: %w", colony.ID, err)
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
