package ai

import (
	"fmt"
	"math"
	"sort"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

const PolicyVersion = "baseline_v1"

type ActionKind string

const (
	ActionNone       ActionKind = "none"
	ActionSubmitTurn ActionKind = "submit_turn"
	ActionImmediate  ActionKind = "immediate"
	ActionBattle     ActionKind = "battle"
	ActionInvasion   ActionKind = "invasion"
	ActionColonyBase ActionKind = "colony_base"
)

type Action struct {
	PolicyVersion string                 `json:"policy_version"`
	Kind          ActionKind             `json:"kind"`
	Batch         *protocol.CommandBatch `json:"batch,omitempty"`
	Command       *protocol.Command      `json:"command,omitempty"`
	BattleID      uint64                 `json:"battle_id,omitempty"`
}

var expansionTechPriority = []int{51, 106, 98, 108, 194}

func Plan(view session.PlayerDecisionView) (Action, error) {
	action := Action{PolicyVersion: PolicyVersion, Kind: ActionNone}
	switch view.Phase {
	case session.PhaseCompleted:
		return action, nil
	case session.PhaseEncounters:
		if len(view.Decisions.Battles) == 0 {
			return action, nil
		}
		decision := view.Decisions.Battles[0]
		if len(decision.Actions) == 0 {
			return action, nil
		}
		command := decision.Actions[0]
		return Action{PolicyVersion: PolicyVersion, Kind: ActionBattle, BattleID: decision.BattleID, Command: &command}, nil
	case session.PhaseInvasionDecisions:
		if view.Decisions.Invasion == nil {
			return action, nil
		}
		command, err := game.NewInvadeCommand(1, game.InvadePayload{ColonyID: view.Decisions.Invasion.ColonyID, TransportFleetIDs: append([]core.ID(nil), view.Decisions.Invasion.EligibleTransportFleetIDs...)})
		if err != nil {
			return Action{}, err
		}
		return Action{PolicyVersion: PolicyVersion, Kind: ActionInvasion, Command: &command}, nil
	case session.PhasePostResolution:
		if len(view.Decisions.ColonyBase) == 0 {
			return action, nil
		}
		resolution := view.Decisions.ColonyBase[0]
		var command protocol.Command
		var err error
		if len(resolution.TargetPlanetIDs) != 0 {
			command, err = game.NewColonizeWithBaseCommand(1, game.ColonizeWithBasePayload{SourceColonyID: resolution.SourceColonyID, PlanetID: resolution.TargetPlanetIDs[0]})
		} else {
			command, err = game.NewTrashColonyBaseCommand(1, game.TrashColonyBasePayload{SourceColonyID: resolution.SourceColonyID})
		}
		if err != nil {
			return Action{}, err
		}
		return Action{PolicyVersion: PolicyVersion, Kind: ActionColonyBase, Command: &command}, nil
	case session.PhasePlanning:
		if command, ok, err := planImmediateDiplomacy(view); err != nil {
			return Action{}, err
		} else if ok {
			return Action{PolicyVersion: PolicyVersion, Kind: ActionImmediate, Command: &command}, nil
		}
		batch, err := planTurn(view)
		if err != nil {
			return Action{}, err
		}
		return Action{PolicyVersion: PolicyVersion, Kind: ActionSubmitTurn, Batch: &batch}, nil
	default:
		return action, nil
	}
}

func planImmediateDiplomacy(view session.PlayerDecisionView) (protocol.Command, bool, error) {
	target, ok := targetEnemyColony(view)
	if !ok || !invasionDepartureReady(view, target.SystemID) {
		return protocol.Command{}, false, nil
	}
	for _, diplomacy := range view.Diplomacy {
		if diplomacy.OtherEmpireID != target.EmpireID {
			continue
		}
		if diplomacy.Stance == core.DiplomaticStanceWar {
			return protocol.Command{}, false, nil
		}
		for _, decision := range view.Decisions.Diplomacy {
			if decision.TargetEmpireID == target.EmpireID && decision.Kind == game.CommandDeclareWar {
				command, err := game.NewDeclareWarCommand(1, target.EmpireID)
				return command, err == nil, err
			}
		}
	}
	return protocol.Command{}, false, nil
}

func planTurn(view session.PlayerDecisionView) (protocol.CommandBatch, error) {
	commands := make([]protocol.Command, 0, 12)
	appendCommand := func(build func(uint32) (protocol.Command, error)) error {
		command, err := build(uint32(len(commands) + 1))
		if err != nil {
			return err
		}
		commands = append(commands, command)
		return nil
	}

	hasExpansionRange := empireKnows(view.Empire, 194)
	if view.Empire.Research == nil && !hasExpansionRange {
		if choice, ok := chooseResearch(view); ok {
			techID := 0
			if len(choice.TechnologyIDs) != 0 {
				techID = choice.TechnologyIDs[0]
			}
			for _, priority := range expansionTechPriority {
				for _, candidate := range choice.TechnologyIDs {
					if candidate == priority {
						techID = candidate
					}
				}
				if techID == priority {
					break
				}
			}
			if techID != 0 {
				if err := appendCommand(func(seq uint32) (protocol.Command, error) {
					return game.NewSelectResearchCommand(seq, game.SelectResearchPayload{TechFieldID: choice.TechFieldID, TechnologyID: techID})
				}); err != nil {
					return protocol.CommandBatch{}, err
				}
			}
		}
	}

	for _, population := range view.Decisions.Population {
		choice, ok := choosePopulation(population.Choices, hasExpansionRange)
		if !ok {
			continue
		}
		colony := colonyByID(view.Colonies, population.ColonyID)
		if colony == nil {
			continue
		}
		if nearlyEqual(colony.Population.Farmers(), choice.Farmers) && nearlyEqual(colony.Population.Workers(), choice.Workers) && nearlyEqual(colony.Population.Scientists(), choice.Scientists) {
			continue
		}
		selected := choice
		if err := appendCommand(func(seq uint32) (protocol.Command, error) {
			return game.NewAssignPopulationCommand(seq, game.AssignPopulationPayload{ColonyID: selected.ColonyID, Farmers: selected.Farmers, Workers: selected.Workers, Scientists: selected.Scientists})
		}); err != nil {
			return protocol.CommandBatch{}, err
		}
	}

	stagingDone := false
	if !hasExpansionRange {
		if staging, ok := earlyCivilianStagingMove(view); ok {
			selected := staging
			if err := appendCommand(func(seq uint32) (protocol.Command, error) {
				return game.NewMoveFleetCommand(seq, game.MoveFleetPayload{FleetID: selected.FleetID, DestinationSystemID: selected.DestinationSystemID, ShipIDs: append([]core.ID(nil), selected.ShipIDs...)})
			}); err != nil {
				return protocol.CommandBatch{}, err
			}
			stagingDone = true
		}
	}

	patrolDone := false
	if patrol, ok := defensivePatrolMove(view); ok {
		selected := patrol
		if err := appendCommand(func(seq uint32) (protocol.Command, error) {
			return game.NewMoveFleetCommand(seq, game.MoveFleetPayload{FleetID: selected.FleetID, DestinationSystemID: selected.DestinationSystemID, ShipIDs: append([]core.ID(nil), selected.ShipIDs...)})
		}); err != nil {
			return protocol.CommandBatch{}, err
		}
		patrolDone = true
	}

	if hasExpansionRange && !patrolDone && !stagingDone {
		// Consume civilian support ships before considering another move/build.
		for _, choice := range bestColonizations(view) {
			selected := choice
			if err := appendCommand(func(seq uint32) (protocol.Command, error) {
				return game.NewColonizePlanetCommand(seq, game.ColonizePlanetPayload{FleetID: selected.FleetID, PlanetID: selected.PlanetID})
			}); err != nil {
				return protocol.CommandBatch{}, err
			}
		}
		for _, choice := range bestOutpostDeployments(view) {
			selected := choice
			if err := appendCommand(func(seq uint32) (protocol.Command, error) {
				return game.NewDeployOutpostCommand(seq, game.DeployOutpostPayload{FleetID: selected.FleetID, BodyID: selected.BodyID, PlanetID: selected.PlanetID})
			}); err != nil {
				return protocol.CommandBatch{}, err
			}
		}

		target, hasTarget := targetEnemyColony(view)
		if !hasTarget {
			if move, ok := bestExplorationMove(view); ok {
				selected := move
				if err := appendCommand(func(seq uint32) (protocol.Command, error) {
					return game.NewMoveFleetCommand(seq, game.MoveFleetPayload{FleetID: selected.FleetID, DestinationSystemID: selected.DestinationSystemID, ShipIDs: append([]core.ID(nil), selected.ShipIDs...)})
				}); err != nil {
					return protocol.CommandBatch{}, err
				}
			}
		}
		if hasTarget {
			stanceWar := false
			for _, diplomacy := range view.Diplomacy {
				if diplomacy.OtherEmpireID == target.EmpireID && diplomacy.Stance == core.DiplomaticStanceWar {
					stanceWar = true
				}
			}
			attackMoved := false
			if stanceWar {
				if move, ok := nextAttackMove(view, target.SystemID); ok {
					selected := move
					if err := appendCommand(func(seq uint32) (protocol.Command, error) {
						return game.NewMoveFleetCommand(seq, game.MoveFleetPayload{FleetID: selected.FleetID, DestinationSystemID: selected.DestinationSystemID, ShipIDs: append([]core.ID(nil), selected.ShipIDs...)})
					}); err != nil {
						return protocol.CommandBatch{}, err
					}
					attackMoved = true
				}
			}
			if !attackMoved {
				for _, fleet := range view.Strategic.Fleets {
					switch fleet.SpecialKind {
					case core.StrategicFleetSpecialColonyShip:
						if hasColonizationForFleet(view.Decisions.Colonization, fleet.ID) {
							continue
						}
						if move, ok := bestExpansionMove(view, fleet.ID, target.SystemID, true); ok {
							selected := move
							if err := appendCommand(func(seq uint32) (protocol.Command, error) {
								return game.NewMoveFleetCommand(seq, game.MoveFleetPayload{FleetID: selected.FleetID, DestinationSystemID: selected.DestinationSystemID, ShipIDs: append([]core.ID(nil), selected.ShipIDs...)})
							}); err != nil {
								return protocol.CommandBatch{}, err
							}
						}
					case core.StrategicFleetSpecialOutpostShip:
						if hasOutpostDeploymentForFleet(view.Decisions.OutpostDeployment, fleet.ID) {
							continue
						}
						if move, ok := bestExpansionMove(view, fleet.ID, target.SystemID, false); ok {
							selected := move
							if err := appendCommand(func(seq uint32) (protocol.Command, error) {
								return game.NewMoveFleetCommand(seq, game.MoveFleetPayload{FleetID: selected.FleetID, DestinationSystemID: selected.DestinationSystemID, ShipIDs: append([]core.ID(nil), selected.ShipIDs...)})
							}); err != nil {
								return protocol.CommandBatch{}, err
							}
						}
					}
				}
			}
			if construction, ok := chooseStrategicConstruction(view, target.SystemID); ok {
				selected := construction
				if err := appendCommand(func(seq uint32) (protocol.Command, error) { return constructionCommand(seq, selected) }); err != nil {
					return protocol.CommandBatch{}, err
				}
			}
		}
	}

	batch := protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: view.GameID, SeatID: view.Seat.Seat.ID, Turn: view.Turn, BaseRevision: view.Revision, Commands: commands}
	if err := batch.Validate(); err != nil {
		return protocol.CommandBatch{}, fmt.Errorf("baseline_v1 batch: %w", err)
	}
	return batch, nil
}

func chooseResearch(view session.PlayerDecisionView) (game.ResearchChoice, bool) {
	known := make(map[int]struct{}, len(view.Empire.KnownTechnologyIDs))
	for _, id := range view.Empire.KnownTechnologyIDs {
		known[id] = struct{}{}
	}
	for _, priority := range expansionTechPriority {
		if _, ok := known[priority]; ok {
			continue
		}
		for _, choice := range view.Decisions.Research {
			for _, technologyID := range choice.TechnologyIDs {
				if technologyID == priority {
					return choice, true
				}
			}
		}
	}
	if len(view.Decisions.Research) == 0 {
		return game.ResearchChoice{}, false
	}
	choices := append([]game.ResearchChoice(nil), view.Decisions.Research...)
	sort.Slice(choices, func(i, j int) bool {
		if choices[i].BaseCostRP != choices[j].BaseCostRP {
			return choices[i].BaseCostRP < choices[j].BaseCostRP
		}
		return choices[i].TechFieldID < choices[j].TechFieldID
	})
	return choices[0], true
}

func choosePopulation(choices []game.PopulationChoice, production bool) (game.PopulationChoice, bool) {
	var best game.PopulationChoice
	found := false
	for _, choice := range choices {
		if !choice.FoodSafe {
			continue
		}
		if !found {
			best, found = choice, true
			continue
		}
		if production {
			if choice.AdjustedEconomy.Production > best.AdjustedEconomy.Production+1e-9 || (nearlyEqual(choice.AdjustedEconomy.Production, best.AdjustedEconomy.Production) && choice.Farmers < best.Farmers) {
				best = choice
			}
		} else {
			if choice.AdjustedEconomy.Research > best.AdjustedEconomy.Research+1e-9 || (nearlyEqual(choice.AdjustedEconomy.Research, best.AdjustedEconomy.Research) && choice.Farmers < best.Farmers) {
				best = choice
			}
		}
	}
	if found {
		return best, true
	}
	for _, choice := range choices {
		if !found || choice.AdjustedEconomy.Food > best.AdjustedEconomy.Food {
			best, found = choice, true
		}
	}
	return best, found
}

func bestColonizations(view session.PlayerDecisionView) []game.ColonizationChoice {
	byFleet := map[core.ID]game.ColonizationChoice{}
	for _, choice := range view.Decisions.Colonization {
		current, ok := byFleet[choice.FleetID]
		if !ok || planetScore(view, choice.PlanetID) > planetScore(view, current.PlanetID) || (planetScore(view, choice.PlanetID) == planetScore(view, current.PlanetID) && choice.PlanetID < current.PlanetID) {
			byFleet[choice.FleetID] = choice
		}
	}
	ids := make([]core.ID, 0, len(byFleet))
	for id := range byFleet {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	out := make([]game.ColonizationChoice, 0, len(ids))
	for _, id := range ids {
		out = append(out, byFleet[id])
	}
	return out
}

func bestOutpostDeployments(view session.PlayerDecisionView) []game.OutpostDeploymentChoice {
	byFleet := map[core.ID]game.OutpostDeploymentChoice{}
	for _, choice := range view.Decisions.OutpostDeployment {
		current, ok := byFleet[choice.FleetID]
		if !ok || choice.PlanetID < current.PlanetID {
			byFleet[choice.FleetID] = choice
		}
	}
	ids := make([]core.ID, 0, len(byFleet))
	for id := range byFleet {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	out := make([]game.OutpostDeploymentChoice, 0, len(ids))
	for _, id := range ids {
		out = append(out, byFleet[id])
	}
	return out
}

func bestExplorationMove(view session.PlayerDecisionView) (game.FleetMoveChoice, bool) {
	visited := make(map[core.ID]struct{}, len(view.Strategic.VisitedSystemIDs))
	for _, systemID := range view.Strategic.VisitedSystemIDs {
		visited[systemID] = struct{}{}
	}
	fleetByID := make(map[core.ID]core.StrategicFleet, len(view.Strategic.Fleets))
	for _, fleet := range view.Strategic.Fleets {
		fleetByID[fleet.ID] = fleet
	}
	priority := func(fleet core.StrategicFleet) int {
		if fleet.Role == core.StrategicFleetRoleCombat && fleet.SpecialKind == core.StrategicFleetSpecialNone {
			return 0
		}
		switch fleet.SpecialKind {
		case core.StrategicFleetSpecialColonyShip:
			return 1
		case core.StrategicFleetSpecialOutpostShip:
			return 2
		default:
			return 3
		}
	}
	var best game.FleetMoveChoice
	bestPriority := math.MaxInt
	bestDistance := int64(math.MaxInt64)
	found := false
	for _, move := range view.Decisions.FleetMoves {
		if len(move.ShipIDs) != 0 {
			continue
		}
		if _, known := visited[move.DestinationSystemID]; known {
			continue
		}
		fleet, ok := fleetByID[move.FleetID]
		if !ok || fleet.AtSystemID == 0 {
			continue
		}
		if fleet.SpecialKind == core.StrategicFleetSpecialColonyShip && hasColonizationForFleet(view.Decisions.Colonization, fleet.ID) {
			continue
		}
		if fleet.SpecialKind == core.StrategicFleetSpecialOutpostShip && hasOutpostDeploymentForFleet(view.Decisions.OutpostDeployment, fleet.ID) {
			continue
		}
		distance, ok := systemDistanceSquared(view, fleet.AtSystemID, move.DestinationSystemID)
		if !ok {
			continue
		}
		p := priority(fleet)
		if !found || p < bestPriority || (p == bestPriority && (distance < bestDistance || (distance == bestDistance && (move.DestinationSystemID < best.DestinationSystemID || (move.DestinationSystemID == best.DestinationSystemID && move.FleetID < best.FleetID))))) {
			best, bestPriority, bestDistance, found = move, p, distance, true
		}
	}
	return best, found
}
func targetEnemyColony(view session.PlayerDecisionView) (session.StrategicContact, bool) {
	supplySystems := ownSupplySystems(view)
	var best session.StrategicContact
	bestDistance := int64(math.MaxInt64)
	found := false
	for _, contact := range view.Strategic.Contacts {
		if contact.Kind != session.StrategicContactColony {
			continue
		}
		distance := int64(math.MaxInt64)
		for _, sourceID := range supplySystems {
			if d, ok := systemDistanceSquared(view, sourceID, contact.SystemID); ok && d < distance {
				distance = d
			}
		}
		if !found || distance < bestDistance || (distance == bestDistance && contact.SystemID < best.SystemID) {
			best, bestDistance, found = contact, distance, true
		}
	}
	return best, found
}

func invasionDepartureReady(view session.PlayerDecisionView, targetSystemID core.ID) bool {
	hasCombat, hasTransport := false, false
	for _, fleet := range view.Strategic.Fleets {
		if _, ok := moveChoiceForFleetTo(view.Decisions.FleetMoves, fleet.ID, targetSystemID); !ok {
			continue
		}
		if fleet.Role == core.StrategicFleetRoleCombat && fleet.SpecialKind == core.StrategicFleetSpecialNone {
			hasCombat = true
		}
		if fleet.SpecialKind == core.StrategicFleetSpecialTroopTransport {
			hasTransport = true
		}
	}
	return hasCombat && hasTransport
}

func invasionFleets(view session.PlayerDecisionView) []core.StrategicFleet {
	var combat, transports []core.StrategicFleet
	for _, fleet := range view.Strategic.Fleets {
		if fleet.Role == core.StrategicFleetRoleCombat && fleet.SpecialKind == core.StrategicFleetSpecialNone {
			combat = append(combat, fleet)
		}
		if fleet.SpecialKind == core.StrategicFleetSpecialTroopTransport {
			transports = append(transports, fleet)
		}
	}
	sort.Slice(combat, func(i, j int) bool { return combat[i].ID < combat[j].ID })
	sort.Slice(transports, func(i, j int) bool { return transports[i].ID < transports[j].ID })
	out := make([]core.StrategicFleet, 0, 2)
	if len(combat) > 0 {
		out = append(out, combat[0])
	}
	if len(transports) > 0 {
		out = append(out, transports[0])
	}
	return out
}

func bestExpansionMove(view session.PlayerDecisionView, fleetID, targetSystemID core.ID, colonize bool) (game.FleetMoveChoice, bool) {
	occupied := occupiedPlanets(view)
	var best game.FleetMoveChoice
	bestDistance := int64(math.MaxInt64)
	found := false
	for _, move := range view.Decisions.FleetMoves {
		if move.FleetID != fleetID {
			continue
		}
		system := systemByID(view, move.DestinationSystemID)
		if system == nil || len(system.Planets) == 0 {
			continue
		}
		hasEmpty := false
		for _, planet := range system.Planets {
			if _, used := occupied[planet.ID]; !used {
				hasEmpty = true
				break
			}
		}
		if !hasEmpty {
			continue
		}
		// Colony Ships never deliberately enter a system with an enemy Colony.
		if colonize && enemyColonyAt(view, system.ID) {
			continue
		}
		distance, ok := systemDistanceSquared(view, system.ID, targetSystemID)
		if !ok {
			continue
		}
		if !found || distance < bestDistance || (distance == bestDistance && move.DestinationSystemID < best.DestinationSystemID) {
			best, bestDistance, found = move, distance, true
		}
	}
	return best, found
}

func chooseStrategicConstruction(view session.PlayerDecisionView, targetSystemID core.ID) (constructionSelection, bool) {
	if anySpecialFleet(view, core.StrategicFleetSpecialColonyShip) {
		return constructionSelection{}, false
	}
	hasCombat := anyCombatFleet(view)
	targetCombatReachable := false
	for _, fleet := range view.Strategic.Fleets {
		if fleet.Role == core.StrategicFleetRoleCombat && fleet.SpecialKind == core.StrategicFleetSpecialNone {
			if _, ok := moveChoiceForFleetTo(view.Decisions.FleetMoves, fleet.ID, targetSystemID); ok {
				targetCombatReachable = true
			}
		}
	}
	needKind := core.ConstructionProjectKind("")
	if !hasCombat {
		needKind = core.ConstructionProjectMilitaryShip
	} else if !targetCombatReachable {
		if !anySpecialFleet(view, core.StrategicFleetSpecialOutpostShip) && !constructionInProgress(view, core.ConstructionProjectOutpostShip) {
			needKind = core.ConstructionProjectOutpostShip
		}
	} else if !anySpecialFleet(view, core.StrategicFleetSpecialTroopTransport) && !constructionInProgress(view, core.ConstructionProjectTroopTransport) {
		needKind = core.ConstructionProjectTroopTransport
	}
	if needKind == "" {
		return constructionSelection{}, false
	}
	for _, colonyDecision := range view.Decisions.Construction {
		for _, choice := range colonyDecision.Choices {
			if choice.ProjectKind == needKind {
				return constructionSelection{ColonyID: colonyDecision.ColonyID, Choice: choice}, true
			}
		}
	}
	return constructionSelection{}, false
}

type constructionSelection struct {
	ColonyID core.ID
	Choice   game.ConstructionChoice
}

func constructionCommand(sequence uint32, selected constructionSelection) (protocol.Command, error) {
	switch selected.Choice.ProjectKind {
	case core.ConstructionProjectBuilding:
		return game.NewQueueBuildingCommand(sequence, game.QueueBuildingPayload{ColonyID: selected.ColonyID, BuildingID: selected.Choice.ProjectID})
	case core.ConstructionProjectColonyShip:
		return game.NewQueueColonyShipCommand(sequence, game.QueueColonyShipPayload{ColonyID: selected.ColonyID})
	case core.ConstructionProjectOutpostShip:
		return game.NewQueueOutpostShipCommand(sequence, game.QueueOutpostShipPayload{ColonyID: selected.ColonyID})
	case core.ConstructionProjectTroopTransport:
		return game.NewQueueTroopTransportCommand(sequence, game.QueueTroopTransportPayload{ColonyID: selected.ColonyID})
	case core.ConstructionProjectMilitaryShip:
		return game.NewQueueMilitaryShipCommand(sequence, game.QueueMilitaryShipPayload{ColonyID: selected.ColonyID, ShipDesignID: selected.Choice.ShipDesignID})
	case core.ConstructionProjectFreighterFleet:
		return game.NewQueueFreighterFleetCommand(sequence, game.QueueFreighterFleetPayload{ColonyID: selected.ColonyID})
	case core.ConstructionProjectHousing:
		return game.NewQueueHousingCommand(sequence, game.QueueHousingPayload{ColonyID: selected.ColonyID})
	case core.ConstructionProjectPlanetaryTransformation:
		return game.NewQueuePlanetaryTransformationCommand(sequence, game.QueuePlanetaryTransformationPayload{ColonyID: selected.ColonyID, ProjectID: selected.Choice.ProjectID})
	default:
		return protocol.Command{}, fmt.Errorf("unsupported construction project %q", selected.Choice.ProjectKind)
	}
}

func empireKnows(empire core.Empire, tech int) bool {
	for _, id := range empire.KnownTechnologyIDs {
		if id == tech {
			return true
		}
	}
	return false
}
func colonyByID(colonies []core.Colony, id core.ID) *core.Colony {
	for i := range colonies {
		if colonies[i].ID == id {
			return &colonies[i]
		}
	}
	return nil
}
func nearlyEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9*math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
}
func moveChoiceForFleetTo(choices []game.FleetMoveChoice, fleetID, systemID core.ID) (game.FleetMoveChoice, bool) {
	var best game.FleetMoveChoice
	found := false
	for _, c := range choices {
		if c.FleetID != fleetID || c.DestinationSystemID != systemID {
			continue
		}
		if !found || (len(c.ShipIDs) == 1 && len(best.ShipIDs) != 1) || (len(c.ShipIDs) == len(best.ShipIDs) && lessShipIDs(c.ShipIDs, best.ShipIDs)) {
			best, found = c, true
		}
	}
	return best, found
}

func earlyCivilianStagingMove(view session.PlayerDecisionView) (game.FleetMoveChoice, bool) {
	ownedColonySystems := map[core.ID]struct{}{}
	for _, colony := range view.Colonies {
		if systemID := systemForPlanet(view, colony.PlanetID); systemID != 0 {
			ownedColonySystems[systemID] = struct{}{}
		}
	}
	var best game.FleetMoveChoice
	bestHasPlanets := true
	bestDistance := int64(math.MaxInt64)
	found := false
	for _, fleet := range view.Strategic.Fleets {
		if fleet.SpecialKind != core.StrategicFleetSpecialColonyShip || fleet.AtSystemID == 0 {
			continue
		}
		if _, ok := ownedColonySystems[fleet.AtSystemID]; !ok {
			continue
		}
		for _, move := range view.Decisions.FleetMoves {
			if move.FleetID != fleet.ID || enemyColonyAt(view, move.DestinationSystemID) {
				continue
			}
			destination := systemByID(view, move.DestinationSystemID)
			if destination == nil {
				continue
			}
			hasPlanets := len(destination.Planets) != 0
			distance, ok := systemDistanceSquared(view, fleet.AtSystemID, move.DestinationSystemID)
			if !ok {
				continue
			}
			if !found || (bestHasPlanets && !hasPlanets) || (bestHasPlanets == hasPlanets && (distance < bestDistance || (distance == bestDistance && move.DestinationSystemID < best.DestinationSystemID))) {
				best, bestHasPlanets, bestDistance, found = move, hasPlanets, distance, true
			}
		}
	}
	return best, found
}

func nextAttackMove(view session.PlayerDecisionView, targetSystemID core.ID) (game.FleetMoveChoice, bool) {
	combatAtTarget := false
	for _, fleet := range view.Strategic.Fleets {
		if fleet.Role == core.StrategicFleetRoleCombat && fleet.SpecialKind == core.StrategicFleetSpecialNone && fleet.AtSystemID == targetSystemID {
			combatAtTarget = true
			break
		}
	}
	if combatAtTarget {
		for _, fleet := range view.Strategic.Fleets {
			if fleet.SpecialKind != core.StrategicFleetSpecialTroopTransport {
				continue
			}
			if move, ok := moveChoiceForFleetTo(view.Decisions.FleetMoves, fleet.ID, targetSystemID); ok {
				return move, true
			}
		}
		return game.FleetMoveChoice{}, false
	}
	for _, fleet := range invasionFleets(view) {
		if fleet.Role != core.StrategicFleetRoleCombat || fleet.SpecialKind != core.StrategicFleetSpecialNone {
			continue
		}
		if move, ok := moveChoiceForFleetTo(view.Decisions.FleetMoves, fleet.ID, targetSystemID); ok {
			return move, true
		}
	}
	return game.FleetMoveChoice{}, false
}

func defensivePatrolMove(view session.PlayerDecisionView) (game.FleetMoveChoice, bool) {
	ownedColonySystems := map[core.ID]struct{}{}
	for _, colony := range view.Colonies {
		if systemID := systemForPlanet(view, colony.PlanetID); systemID != 0 {
			ownedColonySystems[systemID] = struct{}{}
		}
	}
	visited := make(map[core.ID]struct{}, len(view.Strategic.VisitedSystemIDs))
	for _, systemID := range view.Strategic.VisitedSystemIDs {
		visited[systemID] = struct{}{}
	}
	var best game.FleetMoveChoice
	bestVisited := true
	bestDistance := int64(math.MaxInt64)
	found := false
	for _, fleet := range view.Strategic.Fleets {
		if fleet.Role != core.StrategicFleetRoleCombat || fleet.SpecialKind != core.StrategicFleetSpecialNone || fleet.AtSystemID == 0 {
			continue
		}
		if _, ok := ownedColonySystems[fleet.AtSystemID]; !ok {
			continue
		}
		for _, move := range view.Decisions.FleetMoves {
			if move.FleetID != fleet.ID || len(move.ShipIDs) != 0 || enemyColonyAt(view, move.DestinationSystemID) {
				continue
			}
			distance, ok := systemDistanceSquared(view, fleet.AtSystemID, move.DestinationSystemID)
			if !ok {
				continue
			}
			_, destinationVisited := visited[move.DestinationSystemID]
			if !found || (bestVisited && !destinationVisited) || (bestVisited == destinationVisited && (distance < bestDistance || (distance == bestDistance && (move.DestinationSystemID < best.DestinationSystemID || (move.DestinationSystemID == best.DestinationSystemID && lessShipIDs(move.ShipIDs, best.ShipIDs)))))) {
				best, bestVisited, bestDistance, found = move, destinationVisited, distance, true
			}
		}
	}
	return best, found
}

func lessShipIDs(a, b []core.ID) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return len(a) < len(b)
}
func hasColonizationForFleet(choices []game.ColonizationChoice, id core.ID) bool {
	for _, c := range choices {
		if c.FleetID == id {
			return true
		}
	}
	return false
}
func hasOutpostDeploymentForFleet(choices []game.OutpostDeploymentChoice, id core.ID) bool {
	for _, c := range choices {
		if c.FleetID == id {
			return true
		}
	}
	return false
}
func anySpecialFleet(view session.PlayerDecisionView, kind core.StrategicFleetSpecialKind) bool {
	for _, f := range view.Strategic.Fleets {
		if f.SpecialKind == kind {
			return true
		}
	}
	return false
}
func anyCombatFleet(view session.PlayerDecisionView) bool {
	for _, f := range view.Strategic.Fleets {
		if f.Role == core.StrategicFleetRoleCombat && f.SpecialKind == core.StrategicFleetSpecialNone {
			return true
		}
	}
	return false
}
func constructionInProgress(view session.PlayerDecisionView, kind core.ConstructionProjectKind) bool {
	for _, c := range view.Colonies {
		if c.Construction != nil && c.Construction.ProjectKind == kind {
			return true
		}
	}
	return false
}
func ownSupplySystems(view session.PlayerDecisionView) []core.ID {
	set := map[core.ID]struct{}{}
	for _, c := range view.Colonies {
		if id := systemForPlanet(view, c.PlanetID); id != 0 {
			set[id] = struct{}{}
		}
	}
	for _, o := range view.Strategic.Outposts {
		if id := systemForPlanet(view, o.PlanetID); id != 0 {
			set[id] = struct{}{}
		}
	}
	ids := make([]core.ID, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
func systemForPlanet(view session.PlayerDecisionView, planetID core.ID) core.ID {
	for _, s := range view.Strategic.Galaxy.Systems {
		for _, p := range s.Planets {
			if p.ID == planetID {
				return s.ID
			}
		}
	}
	return 0
}
func systemByID(view session.PlayerDecisionView, id core.ID) *core.StarSystem {
	for i := range view.Strategic.Galaxy.Systems {
		if view.Strategic.Galaxy.Systems[i].ID == id {
			return &view.Strategic.Galaxy.Systems[i]
		}
	}
	return nil
}
func systemDistanceSquared(view session.PlayerDecisionView, a, b core.ID) (int64, bool) {
	sa, sb := systemByID(view, a), systemByID(view, b)
	if sa == nil || sb == nil {
		return 0, false
	}
	dx := int64(sa.X - sb.X)
	dy := int64(sa.Y - sb.Y)
	return dx*dx + dy*dy, true
}
func occupiedPlanets(view session.PlayerDecisionView) map[core.ID]struct{} {
	m := map[core.ID]struct{}{}
	for _, c := range view.Colonies {
		m[c.PlanetID] = struct{}{}
	}
	for _, o := range view.Strategic.Outposts {
		m[o.PlanetID] = struct{}{}
	}
	for _, c := range view.Strategic.Contacts {
		if c.PlanetID != 0 && (c.Kind == session.StrategicContactColony || c.Kind == session.StrategicContactOutpost) {
			m[c.PlanetID] = struct{}{}
		}
	}
	return m
}
func enemyColonyAt(view session.PlayerDecisionView, systemID core.ID) bool {
	for _, c := range view.Strategic.Contacts {
		if c.Kind == session.StrategicContactColony && c.SystemID == systemID {
			return true
		}
	}
	return false
}
func planetScore(view session.PlayerDecisionView, planetID core.ID) int {
	for _, s := range view.Strategic.Galaxy.Systems {
		for _, p := range s.Planets {
			if p.ID == planetID {
				switch p.ClimateID {
				case "gaia":
					return 9
				case "terran":
					return 8
				case "ocean":
					return 7
				case "swamp":
					return 6
				case "arid":
					return 5
				case "tundra":
					return 4
				case "barren":
					return 3
				case "radiated":
					return 2
				case "toxic":
					return 1
				default:
					return 0
				}
			}
		}
	}
	return 0
}

var _ = battle.CommandFireBeam
