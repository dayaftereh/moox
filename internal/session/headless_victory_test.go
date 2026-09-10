package session

import (
	"bytes"
	"path/filepath"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

type canonicalResearchStep struct {
	fieldID      int
	technologyID int
}

func TestCanonicalNewGamePublicCommandHeadlessConquestIsExactReplay(t *testing.T) {
	first := runCanonicalHeadlessConquest(t)
	second := runCanonicalHeadlessConquest(t)
	if !bytes.Equal(first, second) {
		t.Fatalf("canonical headless conquest replay differs\nfirst=%s\nsecond=%s", first, second)
	}
}

func runCanonicalHeadlessConquest(t *testing.T) []byte {
	t.Helper()
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	settings := game.NewGameSettings{
		GalaxySize:      game.GalaxySizeSmall,
		GalaxyAge:       game.GalaxyAgeNormal,
		TechnologyLevel: game.NewGameTechnologyAverage,
		Players: []game.NewGamePlayerSpec{
			{SeatID: 1, EmpireName: "Human", RaceID: "human"},
			{SeatID: 2, EmpireName: "Darlok", RaceID: "darlok"},
		},
	}
	generated, err := rules.NewGame(0x8009, settings)
	if err != nil {
		t.Fatal(err)
	}
	if len(generated.Players) != 2 || generated.Players[0].EmpireID == generated.Players[1].EmpireID {
		t.Fatalf("canonical generated players=%+v", generated.Players)
	}
	humanID := generated.Players[0].EmpireID
	darlokID := generated.Players[1].EmpireID
	generated.State.MarkEmpiresKnown(humanID, darlokID)

	seats := []Seat{
		{ID: 1, EmpireID: humanID, Name: "Human", Controller: ControllerLocalHuman},
		{ID: 2, EmpireID: darlokID, Name: "Darlok", Controller: ControllerLocalHuman},
	}
	s, err := NewGameSession("canonical-headless-victory", generated.State, seats)
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	initial := mustObserver(t, s)
	if initial.Turn != 1 || initial.Phase != PhasePlanning || len(initial.State.Colonies) != 2 {
		t.Fatalf("canonical initial session=%+v", initial)
	}
	humanCapital := empireByIDForVictoryTest(t, initial.State, humanID).Capital
	darlokCapital := empireByIDForVictoryTest(t, initial.State, darlokID).Capital
	if humanCapital == 0 || darlokCapital == 0 {
		t.Fatalf("canonical capitals human=%d darlok=%d", humanCapital, darlokCapital)
	}

	// Seed 0x8009's nearest planet-bearing system is 12 pc from the Human home.
	// Play the ordinary research frontier to Urridium Fuel Cells instead of
	// mutating fuel range or galaxy coordinates in the fixture.
	for _, step := range []canonicalResearchStep{
		{fieldID: 9, technologyID: 51},  // Deuterium Fuel Cells.
		{fieldID: 2, technologyID: 106}, // Required intermediate field.
		{fieldID: 47, technologyID: 98}, // Iridium Fuel Cells.
		{fieldID: 53, technologyID: 108},
		{fieldID: 50, technologyID: 194}, // Urridium Fuel Cells: 12 pc.
	} {
		researchCanonicalField(t, s, resolver, rules, humanID, darlokID, step)
	}

	view := mustObserver(t, s)
	humanCombat := findFleetForVictoryTest(t, view.State, humanID, core.StrategicFleetRoleCombat, "")
	humanColonyShip := findFleetForVictoryTest(t, view.State, humanID, core.StrategicFleetRoleCivilian, core.StrategicFleetSpecialColonyShip)
	darlokCombat := findFleetForVictoryTest(t, view.State, darlokID, core.StrategicFleetRoleCombat, "")
	darlokColonyShip := findFleetForVictoryTest(t, view.State, darlokID, core.StrategicFleetRoleCivilian, core.StrategicFleetSpecialColonyShip)

	// Move Darlok's surviving military/civilian assets off the final Colony. They
	// deliberately remain alive until Empire elimination proves fleets do not save it.
	moveCanonicalFleetsToSystem(t, s, resolver, 2, []core.ID{darlokCombat.ID, darlokColonyShip.ID}, 18)

	// Expansion: use the real starting Colony Ship and establish the first 12-pc
	// supply anchor in system 10 on its arid planet (ID 30 for the pinned seed).
	moveCanonicalFleetsToSystem(t, s, resolver, 1, []core.ID{humanColonyShip.ID}, 10)
	colonize := mustCommand(t, func() (protocol.Command, error) {
		return game.NewColonizePlanetCommand(1, game.ColonizePlanetPayload{FleetID: humanColonyShip.ID, PlanetID: 30})
	})
	view = resolveCanonicalTurn(t, s, resolver, []protocol.Command{colonize}, nil)
	requirePostResolution(t, view)
	if colonyByPlanetForVictoryTest(view.State, 30) == nil || colonyByPlanetForVictoryTest(view.State, 30).EmpireID != humanID {
		t.Fatalf("canonical expansion missing Human colony on planet 30: %+v", view.State.Colonies)
	}
	completeCanonicalTurn(t, s)

	// Extend supply with normal Outpost Ship construction/movement/deployment.
	for _, hop := range []struct {
		systemID core.ID
		planetID core.ID
	}{{11, 32}, {12, 33}, {13, 34}, {17, 43}} {
		outpostFleetID := buildCanonicalSpecialFleet(t, s, resolver, humanID, humanCapital, core.StrategicFleetSpecialOutpostShip)
		moveCanonicalFleetsToSystem(t, s, resolver, 1, []core.ID{outpostFleetID}, hop.systemID)
		deploy := mustCommand(t, func() (protocol.Command, error) {
			return game.NewDeployOutpostCommand(1, game.DeployOutpostPayload{FleetID: outpostFleetID, PlanetID: hop.planetID})
		})
		view = resolveCanonicalTurn(t, s, resolver, []protocol.Command{deploy}, nil)
		requirePostResolution(t, view)
		if !hasOutpostForVictoryTest(view.State, humanID, hop.planetID) {
			t.Fatalf("canonical supply hop system=%d planet=%d missing outpost", hop.systemID, hop.planetID)
		}
		completeCanonicalTurn(t, s)
	}

	transportID := buildCanonicalSpecialFleet(t, s, resolver, humanID, humanCapital, core.StrategicFleetSpecialTroopTransport)

	// War is an immediate public session command before the first turn submission.
	status := s.Status()
	declareWar := mustCommand(t, func() (protocol.Command, error) { return game.NewDeclareWarCommand(1, darlokID) })
	if err := s.ResolveDiplomacyCommand(1, status.Revision, declareWar); err != nil {
		t.Fatal(err)
	}
	view = mustObserver(t, s)
	if view.State.DiplomaticStanceBetween(humanID, darlokID) != core.DiplomaticStanceWar {
		t.Fatalf("canonical war declaration did not activate: %+v", view.State.DiplomaticRelations)
	}

	// Send the original combat fleet plus a real built Troop Transport to Darlok's
	// final Colony. The outpost at system 17 makes system 23 legal at 12 pc range.
	view = moveCanonicalFinalInvasionForce(t, s, resolver, humanCombat.ID, transportID, 23)
	if view.Phase != PhaseInvasionDecisions || view.Invasion == nil {
		t.Fatalf("canonical final arrival phase/invasion=%q/%+v", view.Phase, view.Invasion)
	}
	playerView, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if playerView.Invasion == nil || playerView.Invasion.ColonyID != darlokCapital {
		t.Fatalf("canonical attacker invasion projection=%+v want final colony %d", playerView.Invasion, darlokCapital)
	}
	invade := mustCommand(t, func() (protocol.Command, error) {
		return game.NewInvadeCommand(1, game.InvadePayload{
			ColonyID:          playerView.Invasion.ColonyID,
			TransportFleetIDs: append([]core.ID(nil), playerView.Invasion.EligibleTransportFleetIDs...),
		})
	})
	if err := s.ResolveInvasionCommand(1, playerView.Revision, invade); err != nil {
		t.Fatal(err)
	}

	final := mustObserver(t, s)
	if final.Phase != PhaseCompleted || final.Result == nil || final.Result.Kind != ResultConquest || final.Result.WinnerEmpireID != humanID || final.Result.WinnerSeatID != 1 {
		t.Fatalf("canonical conquest result=%+v phase=%q", final.Result, final.Phase)
	}
	if len(final.Result.EliminatedEmpireIDs) != 1 || final.Result.EliminatedEmpireIDs[0] != darlokID {
		t.Fatalf("canonical eliminated IDs=%v want [%d]", final.Result.EliminatedEmpireIDs, darlokID)
	}
	captured := colonyByIDForVictoryTest(final.State, darlokCapital)
	if captured == nil || captured.EmpireID != humanID {
		t.Fatalf("canonical final Colony owner=%+v", captured)
	}
	for _, fleet := range final.State.StrategicFleets {
		if fleet.EmpireID == darlokID {
			t.Fatalf("eliminated Darlok fleet survived completion: %+v", fleet)
		}
	}
	if final.Turn != final.Result.CompletedTurn || final.Revision != final.Result.CompletedRevision {
		t.Fatalf("canonical completion boundary view=%d/%d result=%d/%d", final.Turn, final.Revision, final.Result.CompletedTurn, final.Result.CompletedRevision)
	}
	encoded, err := s.MarshalCompletedSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := UnmarshalCompletedSnapshot(encoded)
	if err != nil {
		t.Fatal(err)
	}
	reencoded, err := restored.MarshalCompletedSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatal("canonical completed save/load bytes changed")
	}
	return encoded
}

func researchCanonicalField(t *testing.T, s *GameSession, resolver *game.EconomyResolver, rules *game.EconomyRules, humanID, darlokID core.ID, step canonicalResearchStep) {
	t.Helper()
	humanSelect := mustCommand(t, func() (protocol.Command, error) {
		return game.NewSelectResearchCommand(1, game.SelectResearchPayload{TechFieldID: step.fieldID, TechnologyID: step.technologyID})
	})
	darlokSelect := mustCommand(t, func() (protocol.Command, error) {
		return game.NewSelectResearchCommand(1, game.SelectResearchPayload{TechFieldID: step.fieldID, TechnologyID: step.technologyID})
	})
	view := resolveCanonicalTurn(t, s, resolver, []protocol.Command{humanSelect}, []protocol.Command{darlokSelect})
	for turn := 0; turn < 5000; turn++ {
		requirePostResolution(t, view)
		for _, empireID := range []core.ID{humanID, darlokID} {
			empire := empireByIDForVictoryTest(t, view.State, empireID)
			if hasTechnologyForVictoryTest(*empire, step.technologyID) {
				continue
			}
			if empire.Research == nil || empire.Research.TechFieldID != step.fieldID {
				t.Fatalf("empire %d research=%+v want field %d", empireID, empire.Research, step.fieldID)
			}
			if empire.Research.ProgressRP+1e-9 >= rules.TechnologyFieldCostsRP[step.fieldID] {
				if err := s.CompleteResearchField(empireID, resolver); err != nil {
					t.Fatal(err)
				}
				view = mustObserver(t, s)
			}
		}
		if hasTechnologyForVictoryTest(*empireByIDForVictoryTest(t, view.State, humanID), step.technologyID) &&
			hasTechnologyForVictoryTest(*empireByIDForVictoryTest(t, view.State, darlokID), step.technologyID) {
			completeCanonicalTurn(t, s)
			return
		}
		completeCanonicalTurn(t, s)
		view = resolveCanonicalTurn(t, s, resolver, nil, nil)
	}
	t.Fatalf("research field %d technology %d exceeded deterministic turn bound", step.fieldID, step.technologyID)
}

func buildCanonicalSpecialFleet(t *testing.T, s *GameSession, resolver *game.EconomyResolver, empireID, colonyID core.ID, kind core.StrategicFleetSpecialKind) core.ID {
	t.Helper()
	before := mustObserver(t, s)
	existing := make(map[core.ID]struct{})
	for _, fleet := range before.State.StrategicFleets {
		existing[fleet.ID] = struct{}{}
	}
	var command protocol.Command
	var err error
	switch kind {
	case core.StrategicFleetSpecialOutpostShip:
		command, err = game.NewQueueOutpostShipCommand(1, game.QueueOutpostShipPayload{ColonyID: colonyID})
	case core.StrategicFleetSpecialTroopTransport:
		command, err = game.NewQueueTroopTransportCommand(1, game.QueueTroopTransportPayload{ColonyID: colonyID})
	default:
		t.Fatalf("unsupported canonical special fleet kind %q", kind)
	}
	if err != nil {
		t.Fatal(err)
	}
	view := resolveCanonicalTurn(t, s, resolver, []protocol.Command{command}, nil)
	for turn := 0; turn < 200; turn++ {
		requirePostResolution(t, view)
		for _, fleet := range view.State.StrategicFleets {
			if fleet.EmpireID == empireID && fleet.SpecialKind == kind {
				if _, old := existing[fleet.ID]; !old {
					completeCanonicalTurn(t, s)
					return fleet.ID
				}
			}
		}
		completeCanonicalTurn(t, s)
		view = resolveCanonicalTurn(t, s, resolver, nil, nil)
	}
	t.Fatalf("building %q exceeded deterministic turn bound", kind)
	return 0
}

func moveCanonicalFleetsToSystem(t *testing.T, s *GameSession, resolver *game.EconomyResolver, seatID protocol.SeatID, fleetIDs []core.ID, destination core.ID) {
	t.Helper()
	commands := make([]protocol.Command, len(fleetIDs))
	for i, fleetID := range fleetIDs {
		commands[i] = mustCommand(t, func() (protocol.Command, error) {
			return game.NewMoveFleetCommand(uint32(i+1), game.MoveFleetPayload{FleetID: fleetID, DestinationSystemID: destination})
		})
	}
	var seat1, seat2 []protocol.Command
	if seatID == 1 {
		seat1 = commands
	} else {
		seat2 = commands
	}
	view := resolveCanonicalTurn(t, s, resolver, seat1, seat2)
	for turn := 0; turn < 200; turn++ {
		requirePostResolution(t, view)
		if fleetsAtSystemForVictoryTest(view.State, fleetIDs, destination) {
			completeCanonicalTurn(t, s)
			return
		}
		completeCanonicalTurn(t, s)
		view = resolveCanonicalTurn(t, s, resolver, nil, nil)
	}
	t.Fatalf("fleets %v did not reach system %d", fleetIDs, destination)
}

func moveCanonicalFinalInvasionForce(t *testing.T, s *GameSession, resolver *game.EconomyResolver, combatFleetID, transportFleetID, destination core.ID) ObserverView {
	t.Helper()
	commands := []protocol.Command{
		mustCommand(t, func() (protocol.Command, error) {
			return game.NewMoveFleetCommand(1, game.MoveFleetPayload{FleetID: combatFleetID, DestinationSystemID: destination})
		}),
		mustCommand(t, func() (protocol.Command, error) {
			return game.NewMoveFleetCommand(2, game.MoveFleetPayload{FleetID: transportFleetID, DestinationSystemID: destination})
		}),
	}
	view := resolveCanonicalTurn(t, s, resolver, commands, nil)
	for turn := 0; turn < 200; turn++ {
		switch view.Phase {
		case PhaseInvasionDecisions:
			return view
		case PhasePostResolution:
			completeCanonicalTurn(t, s)
			view = resolveCanonicalTurn(t, s, resolver, nil, nil)
		default:
			t.Fatalf("canonical final force entered unexpected phase %q", view.Phase)
		}
	}
	t.Fatalf("canonical final invasion force did not create invasion opportunity")
	return ObserverView{}
}

func resolveCanonicalTurn(t *testing.T, s *GameSession, resolver *game.EconomyResolver, seat1, seat2 []protocol.Command) ObserverView {
	t.Helper()
	before := mustObserver(t, s)
	if before.Phase != PhasePlanning {
		t.Fatalf("canonical turn started in phase %q", before.Phase)
	}
	for _, submission := range []struct {
		seatID   protocol.SeatID
		commands []protocol.Command
	}{{seatID: 1, commands: seat1}, {seatID: 2, commands: seat2}} {
		if err := s.SubmitTurn(protocol.CommandBatch{
			SchemaVersion: protocol.CommandSchemaVersion,
			GameID:        before.GameID,
			SeatID:        submission.seatID,
			Turn:          before.Turn,
			BaseRevision:  before.Revision,
			Commands:      submission.commands,
		}); err != nil {
			t.Fatalf("canonical seat %d submit turn %d: %v", submission.seatID, before.Turn, err)
		}
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatalf("canonical resolve turn %d: %v", before.Turn, err)
	}
	return mustObserver(t, s)
}

func completeCanonicalTurn(t *testing.T, s *GameSession) {
	t.Helper()
	if err := s.CompleteTurn(); err != nil {
		t.Fatal(err)
	}
	if status := s.Status(); status.Phase != PhasePlanning {
		t.Fatalf("canonical CompleteTurn phase=%q", status.Phase)
	}
}

func requirePostResolution(t *testing.T, view ObserverView) {
	t.Helper()
	if view.Phase != PhasePostResolution {
		t.Fatalf("canonical phase=%q want=%q", view.Phase, PhasePostResolution)
	}
}

func mustObserver(t *testing.T, s *GameSession) ObserverView {
	t.Helper()
	view, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	return view
}

func mustCommand(t *testing.T, build func() (protocol.Command, error)) protocol.Command {
	t.Helper()
	command, err := build()
	if err != nil {
		t.Fatal(err)
	}
	return command
}

func empireByIDForVictoryTest(t *testing.T, state *core.GameState, empireID core.ID) *core.Empire {
	t.Helper()
	for i := range state.Empires {
		if state.Empires[i].ID == empireID {
			return &state.Empires[i]
		}
	}
	t.Fatalf("missing empire %d", empireID)
	return nil
}

func colonyByIDForVictoryTest(state *core.GameState, colonyID core.ID) *core.Colony {
	for i := range state.Colonies {
		if state.Colonies[i].ID == colonyID {
			return &state.Colonies[i]
		}
	}
	return nil
}

func colonyByPlanetForVictoryTest(state *core.GameState, planetID core.ID) *core.Colony {
	for i := range state.Colonies {
		if state.Colonies[i].PlanetID == planetID {
			return &state.Colonies[i]
		}
	}
	return nil
}

func findFleetForVictoryTest(t *testing.T, state *core.GameState, empireID core.ID, role core.StrategicFleetRole, special core.StrategicFleetSpecialKind) core.StrategicFleet {
	t.Helper()
	for _, fleet := range state.StrategicFleets {
		if fleet.EmpireID == empireID && fleet.Role == role && fleet.SpecialKind == special {
			return fleet
		}
	}
	t.Fatalf("missing empire %d role=%q special=%q fleet", empireID, role, special)
	return core.StrategicFleet{}
}

func fleetsAtSystemForVictoryTest(state *core.GameState, fleetIDs []core.ID, systemID core.ID) bool {
	found := make(map[core.ID]bool, len(fleetIDs))
	for _, fleet := range state.StrategicFleets {
		for _, fleetID := range fleetIDs {
			if fleet.ID == fleetID && fleet.AtSystemID == systemID && fleet.DestinationSystemID == 0 && fleet.RemainingTurns == 0 {
				found[fleetID] = true
			}
		}
	}
	return len(found) == len(fleetIDs)
}

func hasOutpostForVictoryTest(state *core.GameState, empireID, planetID core.ID) bool {
	for _, outpost := range state.Outposts {
		if outpost.EmpireID == empireID && outpost.PlanetID == planetID {
			return true
		}
	}
	return false
}

func hasTechnologyForVictoryTest(empire core.Empire, technologyID int) bool {
	for _, known := range empire.KnownTechnologyIDs {
		if known == technologyID {
			return true
		}
	}
	return false
}
