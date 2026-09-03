package session

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func newInvasionSession(t *testing.T, seed uint64, gameID string, transportCount int) (*GameSession, *game.EconomyResolver, core.ID, core.ID, core.ID, []core.ID) {
	t.Helper()
	state := core.NewSmallFixture(seed)
	attackerID := state.Empires[0].ID
	defenderID := state.NewID()
	targetSystem := &state.Galaxy.Systems[1]
	targetPlanet := &targetSystem.Planets[0]
	targetColonyID := state.NewID()
	defender := core.Empire{ID: defenderID, Name: "Defender", RaceID: "human", Capital: targetColonyID}
	state.Empires = append(state.Empires, defender)
	targetPlanet.ColonyID = targetColonyID
	state.Colonies = append(state.Colonies, core.Colony{
		ID: targetColonyID, EmpireID: defenderID, PlanetID: targetPlanet.ID,
		Population: core.NewAssimilatedPopulation(defenderID, 2, 1, 1),
	})
	state.DiplomaticRelations = []core.DiplomaticRelation{
		{FromEmpireID: attackerID, ToEmpireID: defenderID, Stance: core.DiplomaticStanceWar},
		{FromEmpireID: defenderID, ToEmpireID: attackerID, Stance: core.DiplomaticStanceWar},
	}
	sort.Slice(state.DiplomaticRelations, func(i, j int) bool {
		if state.DiplomaticRelations[i].FromEmpireID != state.DiplomaticRelations[j].FromEmpireID {
			return state.DiplomaticRelations[i].FromEmpireID < state.DiplomaticRelations[j].FromEmpireID
		}
		return state.DiplomaticRelations[i].ToEmpireID < state.DiplomaticRelations[j].ToEmpireID
	})

	spec := core.ShipDesignSpec{
		HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "nuclear_drive", FTLSpeed: 2,
		ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4,
		HullBaseCostPP: 20, HullSpace: 25, BaseDesignCostPP: 25, ProductionCostPP: 25,
		Weapons: []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}},
	}
	designID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{ID: designID, EmpireID: attackerID, Revision: 1, Name: "Escort", Spec: spec})
	shipID := state.NewID()
	state.Ships = append(state.Ships, core.Ship{ID: shipID, EmpireID: attackerID, SourceDesignID: designID, SourceDesignRevision: 1, Name: "Escort", Spec: spec})
	combatFleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{ID: combatFleetID, EmpireID: attackerID, Role: core.StrategicFleetRoleCombat, AtSystemID: targetSystem.ID, ShipIDs: []core.ID{shipID}})
	transportIDs := make([]core.ID, 0, transportCount)
	for i := 0; i < transportCount; i++ {
		fleetID := state.NewID()
		transportIDs = append(transportIDs, fleetID)
		state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{ID: fleetID, EmpireID: attackerID, Role: core.StrategicFleetRoleCivilian, SpecialKind: core.StrategicFleetSpecialTroopTransport, AtSystemID: targetSystem.ID, FTLSpeed: 2})
	}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}

	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession(gameID, state, []Seat{
		{ID: 1, EmpireID: attackerID, Name: "Attacker", Controller: ControllerLocalHuman},
		{ID: 2, EmpireID: defenderID, Name: "Defender", Controller: ControllerLocalHuman},
	})
	if err != nil {
		t.Fatal(err)
	}
	return s, resolver, attackerID, defenderID, targetColonyID, transportIDs
}

func empireFinanceSnapshot(state *core.GameState, empireID core.ID) (core.EmpireCommandPoints, core.EmpireTreasuryState) {
	for i := range state.Empires {
		if state.Empires[i].ID == empireID {
			return state.Empires[i].CommandPoints, state.Empires[i].Treasury
		}
	}
	return core.EmpireCommandPoints{}, core.EmpireTreasuryState{}
}

func submitEmptyInvasionTurn(t *testing.T, s *GameSession) {
	t.Helper()
	status := s.Status()
	for _, seatID := range []protocol.SeatID{1, 2} {
		if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: status.GameID, SeatID: seatID, Turn: status.Turn, BaseRevision: status.Revision, Commands: []protocol.Command{}}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInvasionSessionProjectionStaleRevisionAndDeclineNoReoffer(t *testing.T) {
	s, resolver, attackerID, defenderID, colonyID, _ := newInvasionSession(t, 0xB210, "invasion-decline", 1)
	submitEmptyInvasionTurn(t, s)
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	status := s.Status()
	if status.Phase != PhaseInvasionDecisions {
		t.Fatalf("phase=%q", status.Phase)
	}
	attackerView, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	defenderView, err := s.PlayerView(2)
	if err != nil {
		t.Fatal(err)
	}
	if attackerView.Invasion == nil || attackerView.Invasion.ColonyID != colonyID || attackerView.Invasion.AttackerEmpireID != attackerID || attackerView.Invasion.DefenderEmpireID != defenderID {
		t.Fatalf("attacker invasion projection=%+v", attackerView.Invasion)
	}
	if defenderView.Invasion != nil {
		t.Fatalf("defender leaked invasion projection=%+v", defenderView.Invasion)
	}

	command, err := game.NewDeclineInvasionCommand(1, game.DeclineInvasionPayload{ColonyID: colonyID})
	if err != nil {
		t.Fatal(err)
	}
	beforeObserver, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveInvasionCommand(1, status.Revision-1, command); err == nil {
		t.Fatal("stale invasion revision accepted")
	}
	afterStale, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if afterStale.Revision != beforeObserver.Revision || afterStale.Phase != beforeObserver.Phase || !reflect.DeepEqual(afterStale.State, beforeObserver.State) || !reflect.DeepEqual(afterStale.Invasion, beforeObserver.Invasion) {
		t.Fatal("stale invasion command mutated authoritative session")
	}
	if err := s.ResolveInvasionCommand(1, status.Revision, command); err != nil {
		t.Fatal(err)
	}
	after := s.Status()
	if after.Revision != status.Revision+1 || after.Phase != PhasePostResolution {
		t.Fatalf("decline status=%+v before=%+v", after, status)
	}
	if len(s.handledInvasions) != 1 || s.handledInvasions[0] != (game.InvasionHandledKey{ColonyID: colonyID, AttackerEmpireID: attackerID}) {
		t.Fatalf("handled invasions=%+v", s.handledInvasions)
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if observer.Invasion != nil {
		t.Fatalf("declined invasion was reoffered: %+v", observer.Invasion)
	}
}

func TestInvasionSessionCaptureCommitsOnceAndProjectsCapturedColony(t *testing.T) {
	s, resolver, attackerID, defenderID, colonyID, transports := newInvasionSession(t, 0xB211, "invasion-capture", 2)
	submitEmptyInvasionTurn(t, s)
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	before := s.Status()
	beforeObserver, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	beforeAttackerCP, beforeAttackerTreasury := empireFinanceSnapshot(beforeObserver.State, attackerID)
	beforeDefenderCP, beforeDefenderTreasury := empireFinanceSnapshot(beforeObserver.State, defenderID)
	attackerView, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if attackerView.Invasion == nil {
		t.Fatal("missing invasion projection")
	}
	command, err := game.NewInvadeCommand(1, game.InvadePayload{ColonyID: colonyID, TransportFleetIDs: append([]core.ID(nil), transports...)})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveInvasionCommand(1, before.Revision, command); err != nil {
		t.Fatal(err)
	}
	after := s.Status()
	if after.Revision != before.Revision+1 || after.Phase != PhasePostResolution {
		t.Fatalf("capture status=%+v before=%+v", after, before)
	}
	player, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	var captured *core.Colony
	for i := range player.Colonies {
		if player.Colonies[i].ID == colonyID {
			captured = &player.Colonies[i]
			break
		}
	}
	if captured == nil || captured.EmpireID != attackerID || captured.GroundForces.Infantry != 4 {
		t.Fatalf("captured player colony=%+v", captured)
	}
	defender, err := s.PlayerView(2)
	if err != nil {
		t.Fatal(err)
	}
	for _, colony := range defender.Colonies {
		if colony.ID == colonyID {
			t.Fatalf("defender still projects captured colony %+v", colony)
		}
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	afterAttackerCP, afterAttackerTreasury := empireFinanceSnapshot(observer.State, attackerID)
	afterDefenderCP, afterDefenderTreasury := empireFinanceSnapshot(observer.State, defenderID)
	if !reflect.DeepEqual(afterAttackerCP, beforeAttackerCP) || !reflect.DeepEqual(afterAttackerTreasury, beforeAttackerTreasury) || !reflect.DeepEqual(afterDefenderCP, beforeDefenderCP) || !reflect.DeepEqual(afterDefenderTreasury, beforeDefenderTreasury) {
		t.Fatalf("invasion continuation re-settled Treasury/CP attacker=%+v/%+v defender=%+v/%+v", afterAttackerCP, afterAttackerTreasury, afterDefenderCP, afterDefenderTreasury)
	}
	if observer.State.DiplomaticStanceBetween(attackerID, defenderID) != core.DiplomaticStanceWar {
		t.Fatal("capture changed war stance")
	}
	resolvedIndex, conqueredIndex := -1, -1
	for i, event := range observer.Events {
		if event.Kind == "empire.invasion_resolved" {
			resolvedIndex = i
		}
		if event.Kind == "empire.colony_conquered" {
			conqueredIndex = i
		}
	}
	if resolvedIndex < 0 || conqueredIndex != resolvedIndex+1 {
		t.Fatalf("invasion event order resolved=%d conquered=%d", resolvedIndex, conqueredIndex)
	}
	if len(s.handledInvasions) != 1 {
		t.Fatalf("handled invasions=%+v", s.handledInvasions)
	}
}

func TestDeclaredWarTransitArrivalFlowsIntoInvasionCapture(t *testing.T) {
	s, resolver, attackerID, defenderID, colonyID, transports := newInvasionSession(t, 0xB212, "invasion-arrival", 2)
	// Gate-2 integration starts neutral and with the attacker support group already
	// in ordinary strategic transit toward the enemy Colony system.
	s.state.DiplomaticRelations = nil
	var targetSystemID core.ID
	for si := range s.state.Galaxy.Systems {
		for pi := range s.state.Galaxy.Systems[si].Planets {
			if s.state.Galaxy.Systems[si].Planets[pi].ColonyID == colonyID {
				targetSystemID = s.state.Galaxy.Systems[si].ID
			}
		}
	}
	if targetSystemID == 0 {
		t.Fatal("target system not found")
	}
	for i := range s.state.StrategicFleets {
		fleet := &s.state.StrategicFleets[i]
		if fleet.EmpireID != attackerID || fleet.AtSystemID != targetSystemID {
			continue
		}
		fleet.AtSystemID = 0
		fleet.DestinationSystemID = targetSystemID
		fleet.RemainingTurns = 1
		if fleet.SpecialKind == core.StrategicFleetSpecialTroopTransport {
			fleet.FTLSpeed = 2
		} else {
			fleet.FTLSpeed = 0
		}
	}
	if err := s.state.Validate(); err != nil {
		t.Fatalf("transit setup invalid: %v", err)
	}

	declare, err := game.NewDeclareWarCommand(1, defenderID)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveDiplomacyCommand(1, s.Status().Revision, declare); err != nil {
		t.Fatal(err)
	}
	if s.state.DiplomaticStanceBetween(attackerID, defenderID) != core.DiplomaticStanceWar {
		t.Fatal("war declaration did not commit")
	}

	submitEmptyInvasionTurn(t, s)
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	if s.Status().Phase != PhaseInvasionDecisions {
		t.Fatalf("arrival did not reach invasion decision, phase=%q", s.Status().Phase)
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	arrivals := 0
	for _, event := range observer.Events {
		if event.Kind == "empire.fleet_arrived" {
			arrivals++
		}
	}
	if arrivals < 3 {
		t.Fatalf("arrival events=%d want combat fleet + two transports", arrivals)
	}
	if observer.Invasion == nil || observer.Invasion.ColonyID != colonyID {
		t.Fatalf("arrival invasion=%+v", observer.Invasion)
	}

	command, err := game.NewInvadeCommand(1, game.InvadePayload{ColonyID: colonyID, TransportFleetIDs: transports})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveInvasionCommand(1, s.Status().Revision, command); err != nil {
		t.Fatal(err)
	}
	view, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	var captured *core.Colony
	for i := range view.State.Colonies {
		if view.State.Colonies[i].ID == colonyID {
			captured = &view.State.Colonies[i]
			break
		}
	}
	if captured == nil || captured.EmpireID != attackerID {
		t.Fatalf("arrival conquest state=%+v", captured)
	}
	if view.State.DiplomaticStanceBetween(attackerID, defenderID) != core.DiplomaticStanceWar {
		t.Fatal("capture ended declared war")
	}
}

func TestInvasionSessionFailureConsumesSelectedAndDoesNotReoffer(t *testing.T) {
	s, resolver, attackerID, defenderID, colonyID, transports := newInvasionSession(t, 0xB213, "invasion-failure", 2)
	for i := range s.state.Colonies {
		if s.state.Colonies[i].ID == colonyID {
			s.state.Colonies[i].GroundForces.Infantry = 100
			break
		}
	}
	submitEmptyInvasionTurn(t, s)
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	before := s.Status()
	command, err := game.NewInvadeCommand(1, game.InvadePayload{ColonyID: colonyID, TransportFleetIDs: []core.ID{transports[0]}})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveInvasionCommand(1, before.Revision, command); err != nil {
		t.Fatal(err)
	}
	if s.Status().Phase != PhasePostResolution {
		t.Fatalf("failed invasion phase=%q", s.Status().Phase)
	}
	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if observer.Invasion != nil {
		t.Fatalf("failed invasion reoffered same pair: %+v", observer.Invasion)
	}
	var target *core.Colony
	for i := range observer.State.Colonies {
		if observer.State.Colonies[i].ID == colonyID {
			target = &observer.State.Colonies[i]
			break
		}
	}
	if target == nil || target.EmpireID != defenderID {
		t.Fatalf("failed invasion changed owner: %+v", target)
	}
	if observer.State.DiplomaticStanceBetween(attackerID, defenderID) != core.DiplomaticStanceWar {
		t.Fatal("failed invasion changed war")
	}
	selectedExists, unselectedExists := false, false
	for _, fleet := range observer.State.StrategicFleets {
		if fleet.ID == transports[0] {
			selectedExists = true
		}
		if fleet.ID == transports[1] {
			unselectedExists = true
		}
	}
	if selectedExists || !unselectedExists {
		t.Fatalf("transport survival selected=%t unselected=%t", selectedExists, unselectedExists)
	}
	if len(s.handledInvasions) != 1 || s.handledInvasions[0] != (game.InvasionHandledKey{ColonyID: colonyID, AttackerEmpireID: attackerID}) {
		t.Fatalf("handled invasions=%+v", s.handledInvasions)
	}
}

func TestInvasionHandledKeysClearOnTurnCompletion(t *testing.T) {
	s, resolver, _, _, colonyID, _ := newInvasionSession(t, 0xB214, "invasion-handled-reset", 1)
	submitEmptyInvasionTurn(t, s)
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	command, err := game.NewDeclineInvasionCommand(1, game.DeclineInvasionPayload{ColonyID: colonyID})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveInvasionCommand(1, s.Status().Revision, command); err != nil {
		t.Fatal(err)
	}
	if len(s.handledInvasions) != 1 || s.Status().Phase != PhasePostResolution {
		t.Fatalf("pre-completion handled/phase=%+v/%q", s.handledInvasions, s.Status().Phase)
	}
	if err := s.CompleteTurn(); err != nil {
		t.Fatal(err)
	}
	if len(s.handledInvasions) != 0 || s.Status().Phase != PhasePlanning {
		t.Fatalf("post-completion handled/phase=%+v/%q", s.handledInvasions, s.Status().Phase)
	}
}
