package game

import (
	"fmt"
	"math"

	"moox/internal/core"
	"moox/internal/protocol"
)

type ConstructionQueuedEvent struct {
	ColonyID         core.ID `json:"colony_id"`
	BuildingID       string  `json:"building_id"`
	ProductionCostPP float64 `json:"production_cost_pp"`
}

type ConstructionProgressedEvent struct {
	ColonyID    core.ID                      `json:"colony_id"`
	ProjectKind core.ConstructionProjectKind `json:"project_kind"`
	ProjectID   string                       `json:"project_id"`
	AppliedPP   float64                      `json:"applied_pp"`
	ProgressPP  float64                      `json:"progress_pp"`
	RemainingPP float64                      `json:"remaining_pp"`
}

type BuildingCompletedEvent struct {
	ColonyID   core.ID `json:"colony_id"`
	BuildingID string  `json:"building_id"`
}

type ColonyShipQueuedEvent struct {
	ColonyID         core.ID `json:"colony_id"`
	ProductionCostPP float64 `json:"production_cost_pp"`
}

type ColonyShipCompletedEvent struct {
	ColonyID core.ID `json:"colony_id"`
	EmpireID core.ID `json:"empire_id"`
	FleetID  core.ID `json:"fleet_id"`
	SystemID core.ID `json:"system_id"`
	FTLSpeed int     `json:"ftl_speed"`
}

type OutpostShipQueuedEvent struct {
	ColonyID         core.ID `json:"colony_id"`
	ProductionCostPP float64 `json:"production_cost_pp"`
}

type OutpostShipCompletedEvent struct {
	ColonyID core.ID `json:"colony_id"`
	EmpireID core.ID `json:"empire_id"`
	FleetID  core.ID `json:"fleet_id"`
	SystemID core.ID `json:"system_id"`
	FTLSpeed int     `json:"ftl_speed"`
}

type TroopTransportQueuedEvent struct {
	ColonyID         core.ID `json:"colony_id"`
	ProductionCostPP float64 `json:"production_cost_pp"`
}

type TroopTransportCompletedEvent struct {
	ColonyID core.ID `json:"colony_id"`
	EmpireID core.ID `json:"empire_id"`
	FleetID  core.ID `json:"fleet_id"`
	SystemID core.ID `json:"system_id"`
	FTLSpeed int     `json:"ftl_speed"`
}

type HousingQueuedEvent struct {
	ColonyID core.ID `json:"colony_id"`
}

type HousingStoppedEvent struct {
	ColonyID core.ID `json:"colony_id"`
	Reason   string  `json:"reason"`
}

type PlanetaryTransformationQueuedEvent struct {
	ColonyID         core.ID `json:"colony_id"`
	PlanetID         core.ID `json:"planet_id"`
	ProjectID        string  `json:"project_id"`
	ProductionCostPP float64 `json:"production_cost_pp"`
}

type PlanetaryTransformationCompletedEvent struct {
	ColonyID          core.ID `json:"colony_id"`
	PlanetID          core.ID `json:"planet_id"`
	ProjectID         string  `json:"project_id"`
	PreviousClimateID string  `json:"previous_climate_id"`
	CurrentClimateID  string  `json:"current_climate_id"`
	PopulationRemoved float64 `json:"population_removed,omitempty"`
}

func (r *EconomyResolver) queueBuilding(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeQueueBuilding(command)
	if err != nil {
		return DomainEvent{}, err
	}
	colony := colonyByID(state, payload.ColonyID)
	if colony == nil {
		return DomainEvent{}, fmt.Errorf("references unknown colony %d", payload.ColonyID)
	}
	if colony.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot queue construction on colony %d owned by empire %d", seatID, colony.ID, colony.EmpireID)
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}
	definition, ok := r.Rules.BuildingDefinitions[payload.BuildingID]
	if !ok {
		return DomainEvent{}, fmt.Errorf("unknown building %q", payload.BuildingID)
	}
	if _, transformation := r.Rules.PlanetaryTransformations[payload.BuildingID]; transformation {
		return DomainEvent{}, fmt.Errorf("%q is a planetary transformation, not a persistent building", payload.BuildingID)
	}
	if !empireKnowsTechnology(empire, definition.TechnologyID) {
		return DomainEvent{}, fmt.Errorf("empire %d does not know technology %d required for building %q", empireID, definition.TechnologyID, payload.BuildingID)
	}
	for _, buildingID := range colony.Buildings {
		if buildingID == payload.BuildingID {
			return DomainEvent{}, fmt.Errorf("colony %d already owns building %q", colony.ID, payload.BuildingID)
		}
	}
	if payload.BuildingID == ColonyBaseBuildingID {
		targets, err := colonyBaseTargetPlanetIDs(state, colony)
		if err != nil {
			return DomainEvent{}, err
		}
		if len(targets) == 0 {
			return DomainEvent{}, fmt.Errorf("colony %d has no empty same-system planet for %q", colony.ID, ColonyBaseBuildingID)
		}
	}
	if colony.Construction != nil {
		return DomainEvent{}, fmt.Errorf("colony %d already constructs %q", colony.ID, colony.Construction.ProjectID)
	}
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectBuilding, ProjectID: payload.BuildingID}
	return NewDomainEvent("colony.construction_queued", seatID, command.Sequence, ConstructionQueuedEvent{
		ColonyID:         colony.ID,
		BuildingID:       payload.BuildingID,
		ProductionCostPP: definition.ProductionCostPP,
	})
}

func (r *EconomyResolver) queueHousing(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeQueueHousing(command)
	if err != nil {
		return DomainEvent{}, err
	}
	colony := colonyByID(state, payload.ColonyID)
	if colony == nil {
		return DomainEvent{}, fmt.Errorf("references unknown colony %d", payload.ColonyID)
	}
	if colony.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot queue Housing on colony %d owned by empire %d", seatID, colony.ID, colony.EmpireID)
	}
	if colony.Construction != nil {
		return DomainEvent{}, fmt.Errorf("colony %d already constructs %s %q", colony.ID, colony.Construction.ProjectKind, colony.Construction.ProjectID)
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}
	planet := planetByID(state, colony.PlanetID)
	if planet == nil {
		return DomainEvent{}, fmt.Errorf("colony %d references unknown planet %d", colony.ID, colony.PlanetID)
	}
	capacity, err := r.Rules.ColonyPopulationCapacity(*colony, *planet, *empire)
	if err != nil {
		return DomainEvent{}, err
	}
	if colony.Population.Total() >= capacity-populationEpsilon {
		return DomainEvent{}, fmt.Errorf("colony %d is already at population capacity %g", colony.ID, capacity)
	}
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectHousing, ProjectID: HousingProjectID}
	return NewDomainEvent("colony.housing_queued", seatID, command.Sequence, HousingQueuedEvent{ColonyID: colony.ID})
}

func (r *EconomyResolver) queuePlanetaryTransformation(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeQueuePlanetaryTransformation(command)
	if err != nil {
		return DomainEvent{}, err
	}
	colony := colonyByID(state, payload.ColonyID)
	if colony == nil {
		return DomainEvent{}, fmt.Errorf("references unknown colony %d", payload.ColonyID)
	}
	if colony.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot queue planetary transformation on colony %d owned by empire %d", seatID, colony.ID, colony.EmpireID)
	}
	if colony.Construction != nil {
		return DomainEvent{}, fmt.Errorf("colony %d already constructs %s %q", colony.ID, colony.Construction.ProjectKind, colony.Construction.ProjectID)
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}
	definition, ok := r.Rules.PlanetaryTransformations[payload.ProjectID]
	if !ok {
		return DomainEvent{}, fmt.Errorf("unknown planetary transformation %q", payload.ProjectID)
	}
	if !empireKnowsTechnology(empire, definition.TechnologyID) {
		return DomainEvent{}, fmt.Errorf("empire %d does not know technology %d required for planetary transformation %q", empireID, definition.TechnologyID, payload.ProjectID)
	}
	planet := planetByID(state, colony.PlanetID)
	if planet == nil {
		return DomainEvent{}, fmt.Errorf("colony %d references unknown planet %d", colony.ID, colony.PlanetID)
	}
	if _, allowed := definition.AllowedClimateIDs[planet.ClimateID]; !allowed {
		return DomainEvent{}, fmt.Errorf("planetary transformation %q is not legal on climate %q", payload.ProjectID, planet.ClimateID)
	}
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectPlanetaryTransformation, ProjectID: payload.ProjectID}
	return NewDomainEvent("colony.planetary_transformation_queued", seatID, command.Sequence, PlanetaryTransformationQueuedEvent{
		ColonyID: colony.ID, PlanetID: planet.ID, ProjectID: payload.ProjectID, ProductionCostPP: definition.ProductionCostPP,
	})
}

func (r *EconomyResolver) advanceConstruction(state *core.GameState) ([]DomainEvent, error) {
	var events []DomainEvent
	for i := range state.Colonies {
		colony := &state.Colonies[i]
		if colony.Construction == nil {
			continue
		}
		available := colony.PopulationDynamics.ProductionAvailable
		if available < 0 {
			return nil, fmt.Errorf("colony %d has negative adjusted production %g", colony.ID, available)
		}
		reserveApplied := false
		completedPrior := false
		for colony.Construction != nil {
			project := *colony.Construction
			if project.ProjectKind == core.ConstructionProjectHousing {
				if project.ProjectID != HousingProjectID || math.Abs(project.ProgressPP) > populationEpsilon {
					return nil, fmt.Errorf("colony %d has invalid Housing construction state", colony.ID)
				}
				if colony.PopulationDynamics.Capacity > 0 && colony.Population.Total() >= colony.PopulationDynamics.Capacity-populationEpsilon {
					colony.Construction = nil
					stopped, err := NewDomainEvent("colony.housing_stopped", 0, 0, HousingStoppedEvent{ColonyID: colony.ID, Reason: "population_capacity"})
					if err != nil {
						return nil, err
					}
					events = append(events, stopped)
					promoteConstructionQueue(colony)
					continue
				}
				if completedPrior && available > 0 {
					colony.ConstructionReservePP += available
				}
				break
			}
			if !reserveApplied {
				available += colony.ConstructionReservePP
				colony.ConstructionReservePP = 0
				reserveApplied = true
			}
			costPP, err := r.constructionProjectCostPP(state, colony, colony.Construction)
			if err != nil {
				return nil, fmt.Errorf("colony %d: %w", colony.ID, err)
			}
			remaining := math.Max(0, costPP-colony.Construction.ProgressPP)
			applied := math.Min(available, remaining)
			colony.Construction.ProgressPP += applied
			available -= applied
			remaining -= applied
			progress, err := NewDomainEvent("colony.construction_progressed", 0, 0, ConstructionProgressedEvent{
				ColonyID:    colony.ID,
				ProjectKind: project.ProjectKind,
				ProjectID:   project.ProjectID,
				AppliedPP:   applied,
				ProgressPP:  colony.Construction.ProgressPP,
				RemainingPP: remaining,
			})
			if err != nil {
				return nil, err
			}
			events = append(events, progress)
			if remaining > 1e-9 {
				break
			}
			colony.Construction = nil
			completedEvents, err := r.completeConstructionProject(state, colony, project)
			if err != nil {
				return nil, err
			}
			events = append(events, completedEvents...)
			completedPrior = true
			if !promoteConstructionQueue(colony) {
				if available > 0 {
					colony.ConstructionReservePP += available
				}
				break
			}
		}
	}
	return events, nil
}

func (r *EconomyResolver) completeConstructionProject(state *core.GameState, colony *core.Colony, project core.ConstructionState) ([]DomainEvent, error) {
	projectKind := project.ProjectKind
	projectID := project.ProjectID
	shipDesignID := project.ShipDesignID
	shipDesignRevision := project.ShipDesignRevision
	var events []DomainEvent
	switch projectKind {
	case core.ConstructionProjectBuilding:
		if err := applyCompletedCommandStation(colony, projectID); err != nil {
			return nil, err
		}
		completed, err := NewDomainEvent("colony.building_completed", 0, 0, BuildingCompletedEvent{ColonyID: colony.ID, BuildingID: projectID})
		if err != nil {
			return nil, err
		}
		events = append(events, completed)
	case core.ConstructionProjectPlanetaryTransformation:
		completed, err := r.completePlanetaryTransformation(state, colony, projectID)
		if err != nil {
			return nil, err
		}
		events = append(events, completed)
	case core.ConstructionProjectColonyShip:
		if projectID != ColonyShipProjectID {
			return nil, fmt.Errorf("colony %d completed unknown Colony Ship project %q", colony.ID, projectID)
		}
		empire := empireByID(state, colony.EmpireID)
		if empire == nil {
			return nil, fmt.Errorf("colony %d references unknown empire %d", colony.ID, colony.EmpireID)
		}
		system := systemForPlanetID(state, colony.PlanetID)
		if system == nil {
			return nil, fmt.Errorf("colony %d planet %d is not assigned to a star system", colony.ID, colony.PlanetID)
		}
		ftlSpeed := r.Rules.populationTransferFTLSpeed(*empire)
		if ftlSpeed < 2 {
			return nil, fmt.Errorf("empire %d cannot complete Colony Ship without an installed strategic drive", empire.ID)
		}
		fleet := core.StrategicFleet{ID: state.NewID(), EmpireID: empire.ID, Role: core.StrategicFleetRoleCivilian, SpecialKind: core.StrategicFleetSpecialColonyShip, AtSystemID: system.ID, FTLSpeed: ftlSpeed}
		state.StrategicFleets = append(state.StrategicFleets, fleet)
		completed, err := NewDomainEvent("colony.colony_ship_completed", 0, 0, ColonyShipCompletedEvent{ColonyID: colony.ID, EmpireID: empire.ID, FleetID: fleet.ID, SystemID: system.ID, FTLSpeed: fleet.FTLSpeed})
		if err != nil {
			return nil, err
		}
		events = append(events, completed)
	case core.ConstructionProjectOutpostShip:
		if projectID != OutpostShipProjectID {
			return nil, fmt.Errorf("colony %d completed unknown Outpost Ship project %q", colony.ID, projectID)
		}
		empire := empireByID(state, colony.EmpireID)
		if empire == nil {
			return nil, fmt.Errorf("colony %d references unknown empire %d", colony.ID, colony.EmpireID)
		}
		system := systemForPlanetID(state, colony.PlanetID)
		if system == nil {
			return nil, fmt.Errorf("colony %d planet %d is not assigned to a star system", colony.ID, colony.PlanetID)
		}
		ftlSpeed := r.Rules.populationTransferFTLSpeed(*empire)
		if ftlSpeed < 2 {
			return nil, fmt.Errorf("empire %d cannot complete Outpost Ship without an installed strategic drive", empire.ID)
		}
		fleet := core.StrategicFleet{ID: state.NewID(), EmpireID: empire.ID, Role: core.StrategicFleetRoleCivilian, SpecialKind: core.StrategicFleetSpecialOutpostShip, AtSystemID: system.ID, FTLSpeed: ftlSpeed}
		state.StrategicFleets = append(state.StrategicFleets, fleet)
		completed, err := NewDomainEvent("colony.outpost_ship_completed", 0, 0, OutpostShipCompletedEvent{ColonyID: colony.ID, EmpireID: empire.ID, FleetID: fleet.ID, SystemID: system.ID, FTLSpeed: fleet.FTLSpeed})
		if err != nil {
			return nil, err
		}
		events = append(events, completed)
	case core.ConstructionProjectTroopTransport:
		if projectID != TroopTransportProjectID {
			return nil, fmt.Errorf("colony %d completed unknown Troop Transport project %q", colony.ID, projectID)
		}
		empire := empireByID(state, colony.EmpireID)
		if empire == nil {
			return nil, fmt.Errorf("colony %d references unknown empire %d", colony.ID, colony.EmpireID)
		}
		system := systemForPlanetID(state, colony.PlanetID)
		if system == nil {
			return nil, fmt.Errorf("colony %d planet %d is not assigned to a star system", colony.ID, colony.PlanetID)
		}
		ftlSpeed := r.Rules.populationTransferFTLSpeed(*empire)
		if ftlSpeed < 2 {
			return nil, fmt.Errorf("empire %d cannot complete Troop Transport without an installed strategic drive", empire.ID)
		}
		fleet := core.StrategicFleet{ID: state.NewID(), EmpireID: empire.ID, Role: core.StrategicFleetRoleCivilian, SpecialKind: core.StrategicFleetSpecialTroopTransport, AtSystemID: system.ID, FTLSpeed: ftlSpeed}
		state.StrategicFleets = append(state.StrategicFleets, fleet)
		completed, err := NewDomainEvent("colony.troop_transport_completed", 0, 0, TroopTransportCompletedEvent{ColonyID: colony.ID, EmpireID: empire.ID, FleetID: fleet.ID, SystemID: system.ID, FTLSpeed: fleet.FTLSpeed})
		if err != nil {
			return nil, err
		}
		events = append(events, completed)
	case core.ConstructionProjectMilitaryShip:
		if projectID != MilitaryShipProjectID {
			return nil, fmt.Errorf("colony %d completed unknown military Ship project %q", colony.ID, projectID)
		}
		design, err := militaryConstructionDesign(state, colony.EmpireID, shipDesignID, shipDesignRevision)
		if err != nil {
			return nil, fmt.Errorf("colony %d military Ship completion: %w", colony.ID, err)
		}
		completed, err := completeMilitaryShip(state, colony, design)
		if err != nil {
			return nil, err
		}
		events = append(events, completed)
	case core.ConstructionProjectFreighterFleet:
		empire := empireByID(state, colony.EmpireID)
		if empire == nil {
			return nil, fmt.Errorf("colony %d references unknown empire %d", colony.ID, colony.EmpireID)
		}
		empire.Freighters += r.Rules.FreightersPerFleet
		completed, err := NewDomainEvent("colony.freighter_fleet_completed", 0, 0, FreighterFleetCompletedEvent{ColonyID: colony.ID, EmpireID: empire.ID, FreightersAdded: r.Rules.FreightersPerFleet, TotalFreighters: empire.Freighters})
		if err != nil {
			return nil, err
		}
		events = append(events, completed)
	default:
		return nil, fmt.Errorf("colony %d completed unsupported project kind %q", colony.ID, projectKind)
	}
	return events, nil
}

func (r *EconomyResolver) constructionProjectCostPP(state *core.GameState, colony *core.Colony, project *core.ConstructionState) (float64, error) {
	if project == nil {
		return 0, fmt.Errorf("construction project must not be nil")
	}
	switch project.ProjectKind {
	case core.ConstructionProjectBuilding:
		definition, ok := r.Rules.BuildingDefinitions[project.ProjectID]
		if !ok {
			return 0, fmt.Errorf("constructs unknown building %q", project.ProjectID)
		}
		return definition.ProductionCostPP, nil
	case core.ConstructionProjectPlanetaryTransformation:
		definition, ok := r.Rules.PlanetaryTransformations[project.ProjectID]
		if !ok {
			return 0, fmt.Errorf("constructs unknown planetary transformation %q", project.ProjectID)
		}
		return definition.ProductionCostPP, nil
	case core.ConstructionProjectColonyShip:
		if project.ProjectID != ColonyShipProjectID {
			return 0, fmt.Errorf("constructs unknown Colony Ship project %q", project.ProjectID)
		}
		if state == nil || colony == nil {
			return 0, fmt.Errorf("Colony Ship construction requires authoritative colony state")
		}
		empire := empireByID(state, colony.EmpireID)
		if empire == nil {
			return 0, fmt.Errorf("colony %d references unknown empire %d", colony.ID, colony.EmpireID)
		}
		return r.Rules.colonyShipProductionCostPP(empire), nil
	case core.ConstructionProjectOutpostShip:
		if project.ProjectID != OutpostShipProjectID {
			return 0, fmt.Errorf("constructs unknown Outpost Ship project %q", project.ProjectID)
		}
		if state == nil || colony == nil {
			return 0, fmt.Errorf("Outpost Ship construction requires authoritative colony state")
		}
		empire := empireByID(state, colony.EmpireID)
		if empire == nil {
			return 0, fmt.Errorf("colony %d references unknown empire %d", colony.ID, colony.EmpireID)
		}
		return r.Rules.outpostShipProductionCostPP(empire), nil
	case core.ConstructionProjectTroopTransport:
		if project.ProjectID != TroopTransportProjectID {
			return 0, fmt.Errorf("constructs unknown Troop Transport project %q", project.ProjectID)
		}
		if state == nil || colony == nil {
			return 0, fmt.Errorf("Troop Transport construction requires authoritative colony state")
		}
		empire := empireByID(state, colony.EmpireID)
		if empire == nil {
			return 0, fmt.Errorf("colony %d references unknown empire %d", colony.ID, colony.EmpireID)
		}
		return r.Rules.troopTransportProductionCostPP(empire), nil
	case core.ConstructionProjectMilitaryShip:
		if project.ProjectID != MilitaryShipProjectID {
			return 0, fmt.Errorf("constructs unknown military Ship project %q", project.ProjectID)
		}
		if state == nil || colony == nil {
			return 0, fmt.Errorf("military Ship construction requires authoritative colony state")
		}
		design, err := militaryConstructionDesign(state, colony.EmpireID, project.ShipDesignID, project.ShipDesignRevision)
		if err != nil {
			return 0, err
		}
		return float64(design.Spec.ProductionCostPP), nil
	case core.ConstructionProjectFreighterFleet:
		if project.ProjectID != FreighterFleetProjectID {
			return 0, fmt.Errorf("constructs unknown Freighter Fleet project %q", project.ProjectID)
		}
		return r.Rules.FreighterFleetCostPP, nil
	default:
		return 0, fmt.Errorf("constructs unsupported project kind %q", project.ProjectKind)
	}
}

func (r *EconomyResolver) completePlanetaryTransformation(state *core.GameState, colony *core.Colony, projectID string) (DomainEvent, error) {
	definition, ok := r.Rules.PlanetaryTransformations[projectID]
	if !ok {
		return DomainEvent{}, fmt.Errorf("colony %d completed unknown planetary transformation %q", colony.ID, projectID)
	}
	planet := planetByID(state, colony.PlanetID)
	if planet == nil {
		return DomainEvent{}, fmt.Errorf("colony %d references unknown planet %d", colony.ID, colony.PlanetID)
	}
	previousClimate := planet.ClimateID
	if _, allowed := definition.AllowedClimateIDs[previousClimate]; !allowed {
		return DomainEvent{}, fmt.Errorf("planetary transformation %q is no longer legal on climate %q", projectID, previousClimate)
	}
	empire := empireByID(state, colony.EmpireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("colony %d references unknown empire %d", colony.ID, colony.EmpireID)
	}
	currentClimate, err := r.planetaryTransformationTargetClimate(state, *planet, definition)
	if err != nil {
		return DomainEvent{}, err
	}
	planet.ClimateID = currentClimate
	populationRemoved, err := r.trimColonyToHeterogeneousCapacity(state, colony)
	if err != nil {
		return DomainEvent{}, err
	}
	return NewDomainEvent("colony.planetary_transformation_completed", 0, 0, PlanetaryTransformationCompletedEvent{
		ColonyID: colony.ID, PlanetID: planet.ID, ProjectID: projectID, PreviousClimateID: previousClimate, CurrentClimateID: currentClimate, PopulationRemoved: populationRemoved,
	})
}

func (r *EconomyResolver) planetaryTransformationTargetClimate(state *core.GameState, planet core.Planet, definition PlanetaryTransformationDefinition) (string, error) {
	if result, ok := definition.ResultClimateBySource[planet.ClimateID]; ok {
		return result, nil
	}
	if planet.ClimateID != "barren" || definition.BarrenOrbitRule == nil {
		return "", fmt.Errorf("planetary transformation %q has no result for climate %q", definition.ProjectID, planet.ClimateID)
	}
	rule := definition.BarrenOrbitRule
	switch {
	case planet.Orbit > 0 && planet.Orbit <= rule.InnerOrbitMax:
		return rule.InnerResult, nil
	case planet.Orbit == rule.MiddleOrbit:
		rng := state.RNG()
		index, err := rng.Intn(len(rule.MiddleResults))
		if err != nil {
			return "", err
		}
		state.CommitRNG(rng)
		return rule.MiddleResults[index], nil
	case planet.Orbit >= rule.OuterOrbitMin:
		return rule.OuterResult, nil
	default:
		return "", fmt.Errorf("planet %d has unsupported orbit %d for barren transformation", planet.ID, planet.Orbit)
	}
}

func (r *EconomyResolver) queueColonyShip(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeQueueColonyShip(command)
	if err != nil {
		return DomainEvent{}, err
	}
	colony := colonyByID(state, payload.ColonyID)
	if colony == nil {
		return DomainEvent{}, fmt.Errorf("references unknown colony %d", payload.ColonyID)
	}
	if colony.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot queue Colony Ship on colony %d owned by empire %d", seatID, colony.ID, colony.EmpireID)
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}
	if !empireKnowsTechnology(empire, ColonyShipTechnologyID) {
		return DomainEvent{}, fmt.Errorf("empire %d does not know Technology %d required for Colony Ship", empireID, ColonyShipTechnologyID)
	}
	if colony.Construction != nil {
		return DomainEvent{}, fmt.Errorf("colony %d already constructs %s %q", colony.ID, colony.Construction.ProjectKind, colony.Construction.ProjectID)
	}
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectColonyShip, ProjectID: ColonyShipProjectID}
	return NewDomainEvent("colony.colony_ship_queued", seatID, command.Sequence, ColonyShipQueuedEvent{
		ColonyID: colony.ID, ProductionCostPP: r.Rules.colonyShipProductionCostPP(empire),
	})
}

func (r *EconomyResolver) queueOutpostShip(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeQueueOutpostShip(command)
	if err != nil {
		return DomainEvent{}, err
	}
	colony := colonyByID(state, payload.ColonyID)
	if colony == nil {
		return DomainEvent{}, fmt.Errorf("references unknown colony %d", payload.ColonyID)
	}
	if colony.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot queue Outpost Ship on colony %d owned by empire %d", seatID, colony.ID, colony.EmpireID)
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}
	if !empireKnowsTechnology(empire, OutpostShipTechnologyID) {
		return DomainEvent{}, fmt.Errorf("empire %d does not know Technology %d required for Outpost Ship", empireID, OutpostShipTechnologyID)
	}
	if colony.Construction != nil {
		return DomainEvent{}, fmt.Errorf("colony %d already constructs %s %q", colony.ID, colony.Construction.ProjectKind, colony.Construction.ProjectID)
	}
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectOutpostShip, ProjectID: OutpostShipProjectID}
	return NewDomainEvent("colony.outpost_ship_queued", seatID, command.Sequence, OutpostShipQueuedEvent{
		ColonyID: colony.ID, ProductionCostPP: r.Rules.outpostShipProductionCostPP(empire),
	})
}

func (r *EconomyResolver) queueTroopTransport(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeQueueTroopTransport(command)
	if err != nil {
		return DomainEvent{}, err
	}
	colony := colonyByID(state, payload.ColonyID)
	if colony == nil {
		return DomainEvent{}, fmt.Errorf("references unknown colony %d", payload.ColonyID)
	}
	if colony.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot queue Troop Transport on colony %d owned by empire %d", seatID, colony.ID, colony.EmpireID)
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}
	if colony.Construction != nil {
		return DomainEvent{}, fmt.Errorf("colony %d already constructs %s %q", colony.ID, colony.Construction.ProjectKind, colony.Construction.ProjectID)
	}
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectTroopTransport, ProjectID: TroopTransportProjectID}
	return NewDomainEvent("colony.troop_transport_queued", seatID, command.Sequence, TroopTransportQueuedEvent{ColonyID: colony.ID, ProductionCostPP: r.Rules.troopTransportProductionCostPP(empire)})
}

type FreighterFleetCompletedEvent struct {
	ColonyID        core.ID `json:"colony_id"`
	EmpireID        core.ID `json:"empire_id"`
	FreightersAdded int     `json:"freighters_added"`
	TotalFreighters int     `json:"total_freighters"`
}
type FreighterFleetQueuedEvent struct {
	ColonyID         core.ID `json:"colony_id"`
	ProductionCostPP float64 `json:"production_cost_pp"`
	FreightersAdded  int     `json:"freighters_added"`
}

func (r *EconomyResolver) queueFreighterFleet(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeQueueFreighterFleet(command)
	if err != nil {
		return DomainEvent{}, err
	}
	colony := colonyByID(state, payload.ColonyID)
	if colony == nil {
		return DomainEvent{}, fmt.Errorf("references unknown colony %d", payload.ColonyID)
	}
	if colony.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot queue Freighter Fleet on colony %d owned by empire %d", seatID, colony.ID, colony.EmpireID)
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}
	if !empireKnowsTechnology(empire, FreighterFleetTechnologyID) {
		return DomainEvent{}, fmt.Errorf("empire %d does not know Technology %d required for Freighter Fleet", empireID, FreighterFleetTechnologyID)
	}
	if colony.Construction != nil {
		return DomainEvent{}, fmt.Errorf("colony %d already constructs %s %q", colony.ID, colony.Construction.ProjectKind, colony.Construction.ProjectID)
	}
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectFreighterFleet, ProjectID: FreighterFleetProjectID}
	return NewDomainEvent("colony.freighter_fleet_queued", seatID, command.Sequence, FreighterFleetQueuedEvent{
		ColonyID:         colony.ID,
		ProductionCostPP: r.Rules.FreighterFleetCostPP,
		FreightersAdded:  r.Rules.FreightersPerFleet,
	})
}
