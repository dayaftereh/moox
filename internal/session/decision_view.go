package session

import (
	"fmt"
	"sort"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

type StrategicContactKind string

const (
	StrategicContactColony  StrategicContactKind = "colony"
	StrategicContactOutpost StrategicContactKind = "outpost"
	StrategicContactFleet   StrategicContactKind = "fleet"
)

type StrategicContact struct {
	Kind                StrategicContactKind           `json:"kind"`
	EmpireID            core.ID                        `json:"empire_id"`
	ColonyID            core.ID                        `json:"colony_id,omitempty"`
	OutpostID           core.ID                        `json:"outpost_id,omitempty"`
	FleetID             core.ID                        `json:"fleet_id,omitempty"`
	SystemID            core.ID                        `json:"system_id,omitempty"`
	PlanetID            core.ID                        `json:"planet_id,omitempty"`
	DestinationSystemID core.ID                        `json:"destination_system_id,omitempty"`
	RemainingTurns      int                            `json:"remaining_turns,omitempty"`
	Role                core.StrategicFleetRole        `json:"role,omitempty"`
	SpecialKind         core.StrategicFleetSpecialKind `json:"special_kind,omitempty"`
}

type StrategicView struct {
	Galaxy              core.Galaxy               `json:"galaxy"`
	VisitedSystemIDs    []core.ID                 `json:"visited_system_ids,omitempty"`
	PlanetPotentials    []game.PlanetPotential    `json:"planet_potentials,omitempty"`
	Outposts            []core.Outpost            `json:"outposts,omitempty"`
	ShipDesigns         []core.ShipDesign         `json:"ship_designs,omitempty"`
	Ships               []core.Ship               `json:"ships,omitempty"`
	Fleets              []core.StrategicFleet     `json:"fleets,omitempty"`
	PopulationTransfers []core.PopulationTransfer `json:"population_transfers,omitempty"`
	Contacts            []StrategicContact        `json:"contacts,omitempty"`
}

type PublicEmpireIdentity struct {
	ID              core.ID `json:"id"`
	Name            string  `json:"name"`
	RaceID          string  `json:"race_id"`
	PlayerColorSlot int     `json:"player_color_slot"`
}

type ColonyConstructionDecision struct {
	ColonyID core.ID                   `json:"colony_id"`
	Choices  []game.ConstructionChoice `json:"choices"`
}

type ColonyPopulationDecision struct {
	ColonyID core.ID                 `json:"colony_id"`
	Choices  []game.PopulationChoice `json:"choices"`
}

type DiplomacyDecision struct {
	Kind           string  `json:"kind"`
	TargetEmpireID core.ID `json:"target_empire_id"`
}

type BattleDecision struct {
	BattleID uint64             `json:"battle_id"`
	Actions  []protocol.Command `json:"actions"`
}

type DecisionCatalog struct {
	ResearchCategories  []game.ResearchCategory         `json:"research_categories,omitempty"`
	Research            []game.ResearchChoice           `json:"research,omitempty"`
	Construction        []ColonyConstructionDecision    `json:"construction,omitempty"`
	Population          []ColonyPopulationDecision      `json:"population,omitempty"`
	PopulationTransfers []game.PopulationTransferChoice `json:"population_transfers,omitempty"`
	FleetMoveTargets    []game.FleetMoveTarget          `json:"fleet_move_targets,omitempty"`
	FleetMoves          []game.FleetMoveChoice          `json:"fleet_moves,omitempty"`
	Colonization        []game.ColonizationChoice       `json:"colonization,omitempty"`
	OutpostDeployment   []game.OutpostDeploymentChoice  `json:"outpost_deployment,omitempty"`
	Diplomacy           []DiplomacyDecision             `json:"diplomacy,omitempty"`
	ColonyBase          []game.ColonyBaseResolution     `json:"colony_base,omitempty"`
	Invasion            *game.InvasionOpportunity       `json:"invasion,omitempty"`
	ShipDesigner        game.MilitaryDesignerCatalog    `json:"ship_designer"`
	Battles             []BattleDecision                `json:"battles,omitempty"`
}

// PlayerDecisionView is the complete input boundary for built-in AI and future
// strategic HMI clients. It is derived from a deep state clone and intentionally
// excludes observer-only enemy economy/research/queue/submission information.
type PlayerDecisionView struct {
	GameID        string                 `json:"game_id"`
	Revision      uint64                 `json:"revision"`
	Turn          uint64                 `json:"turn"`
	Phase         Phase                  `json:"phase"`
	Seat          SeatView               `json:"seat"`
	Empire        core.Empire            `json:"empire"`
	PublicEmpires []PublicEmpireIdentity `json:"public_empires,omitempty"`
	Colonies      []core.Colony          `json:"colonies"`
	Diplomacy     []DiplomacyView        `json:"diplomacy,omitempty"`
	Strategic     StrategicView          `json:"strategic"`
	Decisions     DecisionCatalog        `json:"decisions"`
}

func playerColorSlotForEmpire(state *core.GameState, empireID core.ID) int {
	if state == nil || empireID == 0 {
		return 1
	}
	for i, empire := range state.Empires {
		if empire.ID != empireID {
			continue
		}
		if empire.PlayerColorSlot >= 1 && empire.PlayerColorSlot <= 8 {
			return empire.PlayerColorSlot
		}
		return (i % 8) + 1
	}
	return 1
}
func (s *GameSession) DecisionView(seatID protocol.SeatID, resolver *game.EconomyResolver) (PlayerDecisionView, error) {
	if resolver == nil || resolver.Rules == nil {
		return PlayerDecisionView{}, fmt.Errorf("economy resolver has no rules")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return PlayerDecisionView{}, fmt.Errorf("unknown seat %d", seatID)
	}
	seat := s.seats[index].seat
	state, err := cloneState(s.state)
	if err != nil {
		return PlayerDecisionView{}, err
	}
	view := PlayerDecisionView{
		GameID:   s.gameID,
		Revision: s.revision,
		Turn:     state.Turn,
		Phase:    s.phase,
		Seat:     seatView(s.seats[index]),
	}
	for _, empire := range state.Empires {
		if empire.ID == seat.EmpireID {
			view.Empire = empire
			view.Empire.PlayerColorSlot = playerColorSlotForEmpire(state, empire.ID)
			break
		}
	}
	if view.Empire.ID == 0 {
		return PlayerDecisionView{}, fmt.Errorf("seat %d references unknown empire %d", seatID, seat.EmpireID)
	}
	for _, empire := range state.Empires {
		if empire.ID != seat.EmpireID && !state.EmpiresHaveContact(seat.EmpireID, empire.ID) {
			continue
		}
		view.PublicEmpires = append(view.PublicEmpires, PublicEmpireIdentity{ID: empire.ID, Name: empire.Name, RaceID: empire.RaceID, PlayerColorSlot: playerColorSlotForEmpire(state, empire.ID)})
	}
	sort.Slice(view.PublicEmpires, func(i, j int) bool { return view.PublicEmpires[i].ID < view.PublicEmpires[j].ID })
	for _, colony := range state.Colonies {
		if colony.EmpireID == seat.EmpireID {
			view.Colonies = append(view.Colonies, colony)
		}
	}
	sort.Slice(view.Colonies, func(i, j int) bool { return view.Colonies[i].ID < view.Colonies[j].ID })
	for _, other := range state.Empires {
		if other.ID == seat.EmpireID || !state.EmpiresHaveContact(seat.EmpireID, other.ID) {
			continue
		}
		view.Diplomacy = append(view.Diplomacy, DiplomacyView{
			OtherEmpireID:      other.ID,
			Stance:             state.DiplomaticStanceBetween(seat.EmpireID, other.ID),
			IncomingPeaceOffer: state.HasDiplomaticPeaceOffer(other.ID, seat.EmpireID),
			OutgoingPeaceOffer: state.HasDiplomaticPeaceOffer(seat.EmpireID, other.ID),
		})
	}
	sort.Slice(view.Diplomacy, func(i, j int) bool { return view.Diplomacy[i].OtherEmpireID < view.Diplomacy[j].OtherEmpireID })
	view.Strategic = buildStrategicView(state, seat.EmpireID)

	for _, system := range view.Strategic.Galaxy.Systems {
		for _, planet := range system.Planets {
			potential, err := resolver.Rules.PlanetPotentialForEmpire(planet, view.Empire)
			if err != nil {
				return PlayerDecisionView{}, fmt.Errorf("planet %d potential: %w", planet.ID, err)
			}
			view.Strategic.PlanetPotentials = append(view.Strategic.PlanetPotentials, potential)
		}
	}

	view.Decisions.ResearchCategories = resolver.Rules.AvailableResearchCategories()
	view.Decisions.Research, err = resolver.Rules.AvailableResearchChoices(state, seat.EmpireID)
	if err != nil {
		return PlayerDecisionView{}, fmt.Errorf("research choices: %w", err)
	}
	view.Decisions.ShipDesigner, err = resolver.Rules.MilitaryDesignerCatalog(&view.Empire)
	if err != nil {
		return PlayerDecisionView{}, fmt.Errorf("ship designer choices: %w", err)
	}
	for _, colony := range view.Colonies {
		construction, err := resolver.Rules.AvailableConstructionQueueChoices(state, seat.EmpireID, colony.ID)
		if err != nil {
			return PlayerDecisionView{}, fmt.Errorf("colony %d construction choices: %w", colony.ID, err)
		}
		population, err := resolver.AvailablePopulationChoices(state, seat.EmpireID, colony.ID)
		if err != nil {
			return PlayerDecisionView{}, fmt.Errorf("colony %d population choices: %w", colony.ID, err)
		}
		view.Decisions.Construction = append(view.Decisions.Construction, ColonyConstructionDecision{ColonyID: colony.ID, Choices: construction})
		view.Decisions.Population = append(view.Decisions.Population, ColonyPopulationDecision{ColonyID: colony.ID, Choices: population})
	}
	view.Decisions.PopulationTransfers, err = resolver.AvailablePopulationTransferChoices(state, seat.EmpireID)
	if err != nil {
		return PlayerDecisionView{}, fmt.Errorf("population transfer choices: %w", err)
	}
	view.Decisions.FleetMoveTargets, err = resolver.AvailableFleetMoveTargets(state, seat.EmpireID)
	if err != nil {
		return PlayerDecisionView{}, fmt.Errorf("fleet move targets: %w", err)
	}
	view.Decisions.FleetMoves, err = resolver.AvailableFleetMoveChoices(state, seat.EmpireID)
	if err != nil {
		return PlayerDecisionView{}, fmt.Errorf("fleet move choices: %w", err)
	}
	view.Decisions.Colonization, err = resolver.AvailableColonizationChoices(state, seat.EmpireID)
	if err != nil {
		return PlayerDecisionView{}, fmt.Errorf("colonization choices: %w", err)
	}
	view.Decisions.OutpostDeployment, err = resolver.AvailableOutpostDeploymentChoices(state, seat.EmpireID)
	if err != nil {
		return PlayerDecisionView{}, fmt.Errorf("outpost deployment choices: %w", err)
	}
	view.Decisions.Diplomacy = diplomacyDecisionCatalog(state, seat.EmpireID)
	view.Decisions.ColonyBase, err = game.PendingColonyBaseResolutions(state, seat.EmpireID)
	if err != nil {
		return PlayerDecisionView{}, fmt.Errorf("colony base choices: %w", err)
	}
	if s.invasion != nil && s.invasion.AttackerSeatID == seatID {
		view.Decisions.Invasion = game.CloneInvasionOpportunity(s.invasion)
	}
	view.Decisions.Battles = s.battleDecisionCatalogLocked(seatID)
	return view, nil
}

func buildStrategicView(state *core.GameState, empireID core.ID) StrategicView {
	visitedSystemIDs := playerVisitedSystemIDs(state, empireID)
	visited := make(map[core.ID]struct{}, len(visitedSystemIDs))
	for _, systemID := range visitedSystemIDs {
		visited[systemID] = struct{}{}
	}
	galaxy := state.Galaxy
	galaxy.Systems = make([]core.StarSystem, len(state.Galaxy.Systems))
	for i, system := range state.Galaxy.Systems {
		galaxy.Systems[i] = system
		galaxy.Systems[i].BlockadedEmpireIDs = nil
		if _, ok := visited[system.ID]; !ok {
			galaxy.Systems[i].Name = ""
			galaxy.Systems[i].Planets = nil
			galaxy.Systems[i].Bodies = nil
			continue
		}
		galaxy.Systems[i].Planets = append([]core.Planet(nil), system.Planets...)
		galaxy.Systems[i].Bodies = append([]core.OrbitalBody(nil), system.Bodies...)
		for pi := range galaxy.Systems[i].Planets {
			galaxy.Systems[i].Planets[pi].ColonyID = 0
			galaxy.Systems[i].Planets[pi].OutpostID = 0
		}
	}
	out := StrategicView{Galaxy: galaxy, VisitedSystemIDs: visitedSystemIDs}
	for _, outpost := range state.Outposts {
		if outpost.EmpireID == empireID {
			out.Outposts = append(out.Outposts, outpost)
		}
	}
	for _, design := range state.ShipDesigns {
		if design.EmpireID == empireID {
			out.ShipDesigns = append(out.ShipDesigns, design)
		}
	}
	for _, ship := range state.Ships {
		if ship.EmpireID == empireID {
			out.Ships = append(out.Ships, ship)
		}
	}
	for _, fleet := range state.StrategicFleets {
		if fleet.EmpireID == empireID {
			fleet.ShipIDs = append([]core.ID(nil), fleet.ShipIDs...)
			out.Fleets = append(out.Fleets, fleet)
		} else if fleet.AtSystemID != 0 && state.EmpiresHaveContact(empireID, fleet.EmpireID) {
			if _, ok := visited[fleet.AtSystemID]; ok {
				out.Contacts = append(out.Contacts, StrategicContact{Kind: StrategicContactFleet, EmpireID: fleet.EmpireID, FleetID: fleet.ID, SystemID: fleet.AtSystemID, Role: fleet.Role, SpecialKind: fleet.SpecialKind})
			}
		}
	}
	for _, transfer := range state.PopulationTransfers {
		if transfer.EmpireID == empireID {
			out.PopulationTransfers = append(out.PopulationTransfers, transfer)
		}
	}
	for _, colony := range state.Colonies {
		if colony.EmpireID == empireID || !state.EmpiresHaveContact(empireID, colony.EmpireID) {
			continue
		}
		if systemID := systemIDForPlanet(state, colony.PlanetID); systemID != 0 {
			if _, ok := visited[systemID]; ok {
				out.Contacts = append(out.Contacts, StrategicContact{Kind: StrategicContactColony, EmpireID: colony.EmpireID, ColonyID: colony.ID, SystemID: systemID, PlanetID: colony.PlanetID})
			}
		}
	}
	for _, outpost := range state.Outposts {
		if outpost.EmpireID == empireID || !state.EmpiresHaveContact(empireID, outpost.EmpireID) {
			continue
		}
		if systemID := systemIDForPlanet(state, outpost.PlanetID); systemID != 0 {
			if _, ok := visited[systemID]; ok {
				out.Contacts = append(out.Contacts, StrategicContact{Kind: StrategicContactOutpost, EmpireID: outpost.EmpireID, OutpostID: outpost.ID, SystemID: systemID, PlanetID: outpost.PlanetID})
			}
		}
	}
	sort.Slice(out.Outposts, func(i, j int) bool { return out.Outposts[i].ID < out.Outposts[j].ID })
	sort.Slice(out.ShipDesigns, func(i, j int) bool { return out.ShipDesigns[i].ID < out.ShipDesigns[j].ID })
	sort.Slice(out.Ships, func(i, j int) bool { return out.Ships[i].ID < out.Ships[j].ID })
	sort.Slice(out.Fleets, func(i, j int) bool { return out.Fleets[i].ID < out.Fleets[j].ID })
	sort.Slice(out.PopulationTransfers, func(i, j int) bool { return out.PopulationTransfers[i].ID < out.PopulationTransfers[j].ID })
	sort.Slice(out.Contacts, func(i, j int) bool {
		if out.Contacts[i].EmpireID != out.Contacts[j].EmpireID {
			return out.Contacts[i].EmpireID < out.Contacts[j].EmpireID
		}
		if out.Contacts[i].Kind != out.Contacts[j].Kind {
			return out.Contacts[i].Kind < out.Contacts[j].Kind
		}
		if out.Contacts[i].SystemID != out.Contacts[j].SystemID {
			return out.Contacts[i].SystemID < out.Contacts[j].SystemID
		}
		if out.Contacts[i].PlanetID != out.Contacts[j].PlanetID {
			return out.Contacts[i].PlanetID < out.Contacts[j].PlanetID
		}
		return out.Contacts[i].FleetID < out.Contacts[j].FleetID
	})
	return out
}

func playerVisitedSystemIDs(state *core.GameState, empireID core.ID) []core.ID {
	visited := make(map[core.ID]struct{})
	for i := range state.Empires {
		if state.Empires[i].ID != empireID {
			continue
		}
		for _, systemID := range state.Empires[i].VisitedSystemIDs {
			visited[systemID] = struct{}{}
		}
		break
	}
	// Compatibility for saves created before visited-system persistence existed:
	// current owned assets prove that their systems have necessarily been visited.
	for _, colony := range state.Colonies {
		if colony.EmpireID == empireID {
			if systemID := systemIDForPlanet(state, colony.PlanetID); systemID != 0 {
				visited[systemID] = struct{}{}
			}
		}
	}
	for _, outpost := range state.Outposts {
		if outpost.EmpireID != empireID {
			continue
		}
		if systemID := systemIDForOutpost(state, outpost); systemID != 0 {
			visited[systemID] = struct{}{}
		}
	}
	for _, fleet := range state.StrategicFleets {
		if fleet.EmpireID == empireID && fleet.AtSystemID != 0 {
			visited[fleet.AtSystemID] = struct{}{}
		}
	}
	ids := make([]core.ID, 0, len(visited))
	for systemID := range visited {
		ids = append(ids, systemID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func systemIDForOutpost(state *core.GameState, outpost core.Outpost) core.ID {
	if outpost.PlanetID != 0 {
		if systemID := systemIDForPlanet(state, outpost.PlanetID); systemID != 0 {
			return systemID
		}
	}
	if outpost.BodyID == 0 {
		return 0
	}
	for _, system := range state.Galaxy.Systems {
		for _, body := range system.Bodies {
			if body.ID == outpost.BodyID {
				return system.ID
			}
		}
	}
	return 0
}

func systemIDForPlanet(state *core.GameState, planetID core.ID) core.ID {
	for _, system := range state.Galaxy.Systems {
		for _, planet := range system.Planets {
			if planet.ID == planetID {
				return system.ID
			}
		}
	}
	return 0
}

func diplomacyDecisionCatalog(state *core.GameState, empireID core.ID) []DiplomacyDecision {
	var out []DiplomacyDecision
	for _, other := range state.Empires {
		if other.ID == empireID || !state.EmpiresHaveContact(empireID, other.ID) {
			continue
		}
		stance := state.DiplomaticStanceBetween(empireID, other.ID)
		if state.HasDiplomaticPeaceOffer(other.ID, empireID) {
			out = append(out, DiplomacyDecision{Kind: game.CommandAcceptPeace, TargetEmpireID: other.ID})
		}
		if stance == core.DiplomaticStanceWar {
			if !state.HasDiplomaticPeaceOffer(empireID, other.ID) {
				out = append(out, DiplomacyDecision{Kind: game.CommandOfferPeace, TargetEmpireID: other.ID})
			}
		} else {
			out = append(out, DiplomacyDecision{Kind: game.CommandDeclareWar, TargetEmpireID: other.ID})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].TargetEmpireID != out[j].TargetEmpireID {
			return out[i].TargetEmpireID < out[j].TargetEmpireID
		}
		return out[i].Kind < out[j].Kind
	})
	return out
}

func (s *GameSession) battleDecisionCatalogLocked(seatID protocol.SeatID) []BattleDecision {
	var out []BattleDecision
	for _, child := range s.battles {
		view := child.View()
		if view.Phase != battle.PhaseActive || view.Tactical == nil {
			continue
		}
		side := view.Spec.Attacker
		enemy := view.Spec.Defender
		if view.Spec.Defender.SeatID == seatID {
			side, enemy = view.Spec.Defender, view.Spec.Attacker
		} else if view.Spec.Attacker.SeatID != seatID {
			continue
		}
		active := view.Tactical.State.ActiveShipID
		ownsActive := false
		for _, id := range side.ShipIDs {
			if id == active {
				ownsActive = true
				break
			}
		}
		if !ownsActive {
			continue
		}
		var activeState *battle.TacticalShipState
		for i := range view.Tactical.State.Ships {
			if view.Tactical.State.Ships[i].ShipID == active {
				activeState = &view.Tactical.State.Ships[i]
				break
			}
		}
		if activeState == nil || activeState.Destroyed {
			continue
		}
		decision := BattleDecision{BattleID: view.Spec.ID}
		sequence := view.Tactical.State.NextCommandSequence
		for _, weapon := range activeState.Weapons {
			if !weapon.Ready {
				continue
			}
			for _, target := range enemy.ShipIDs {
				cmd, err := battle.NewFireBeamCommand(sequence, battle.FireBeamPayload{ShipID: active, TargetShipID: target, WeaponSlot: weapon.Slot})
				if err != nil {
					continue
				}
				if _, err := child.PrepareCommand(seatID, cmd); err == nil {
					decision.Actions = append(decision.Actions, cmd)
				}
			}
		}
		end, err := battle.NewEndActivationCommand(sequence, battle.EndActivationPayload{ShipID: active})
		if err == nil {
			if _, err := child.PrepareCommand(seatID, end); err == nil {
				decision.Actions = append(decision.Actions, end)
			}
		}
		if len(decision.Actions) != 0 {
			out = append(out, decision)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].BattleID < out[j].BattleID })
	return out
}

func (s *GameSession) SeatController(seatID protocol.SeatID) (ControllerType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	index := s.seatIndexLocked(seatID)
	if index < 0 {
		return "", fmt.Errorf("unknown seat %d", seatID)
	}
	return s.seats[index].seat.Controller, nil
}
func (s *GameSession) DueResearchEmpireIDs(rules *game.EconomyRules) ([]core.ID, error) {
	if rules == nil {
		return nil, fmt.Errorf("economy rules must not be nil")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, err := cloneState(s.state)
	if err != nil {
		return nil, err
	}
	var ids []core.ID
	for _, empire := range state.Empires {
		if s.empireEliminatedLocked(empire.ID) {
			continue
		}
		due, err := rules.ResearchDue(state, empire.ID)
		if err != nil {
			return nil, fmt.Errorf("empire %d research due: %w", empire.ID, err)
		}
		if due {
			ids = append(ids, empire.ID)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}
