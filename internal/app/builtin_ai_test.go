package app

import (
	"bytes"
	"encoding/json"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

func registerPostContactCanonicalAIGame(t *testing.T, host *Host, gameID string) {
	t.Helper()
	generated, err := host.newGameRules.NewGame(0x8009, appNewGameSettings())
	if err != nil {
		t.Fatal(err)
	}
	// This legacy completion regression predates fog-of-knowledge. Preserve its
	// original post-contact/full-map premise so it continues to test deterministic
	// AI conquest rather than the separate discovery/range integration path.
	for ei := range generated.State.Empires {
		for _, system := range generated.State.Galaxy.Systems {
			generated.State.Empires[ei].MarkSystemVisited(system.ID)
		}
	}
	for i := 0; i < len(generated.State.Empires); i++ {
		for j := i + 1; j < len(generated.State.Empires); j++ {
			generated.State.MarkEmpiresKnown(generated.State.Empires[i].ID, generated.State.Empires[j].ID)
		}
	}
	seats := make([]session.Seat, len(generated.Players))
	for i, player := range generated.Players {
		seats[i] = session.Seat{ID: player.SeatID, EmpireID: player.EmpireID, Name: player.Name, Controller: session.ControllerBuiltinAI}
	}
	gameSession, err := session.NewGameSession(gameID, generated.State, seats)
	if err != nil {
		t.Fatal(err)
	}
	if err := host.Register(Registration{Session: gameSession, Resolver: host.newGameResolver, ImmediateResolver: host.newGameImmediateResolver}); err != nil {
		t.Fatal(err)
	}
}
func TestBuiltinAICanonicalNewGameCompletesAndReplaysExactly(t *testing.T) {
	first := runCanonicalBuiltinAIMatch(t)
	second := runCanonicalBuiltinAIMatch(t)
	if !bytes.Equal(first, second) {
		t.Fatalf("canonical builtin AI completed snapshots differ: first=%d bytes second=%d bytes", len(first), len(second))
	}
}

func runCanonicalBuiltinAIMatch(t *testing.T) []byte {
	t.Helper()
	host := loadNewGameHost(t)
	const gameID = "builtin-ai-canonical"
	registerPostContactCanonicalAIGame(t, host, gameID)

	completed := false
	for step := 0; step < 1000; step++ {
		before := host.ListGames()[0]
		if before.Phase == session.PhaseCompleted {
			completed = true
			break
		}
		receipt, err := host.AdvanceAutomation(gameID)
		if err != nil {
			dumpBuiltinAIDiagnostics(t, host, gameID, step, err)
			t.Fatalf("advance automation step %d: %v", step, err)
		}
		after := host.ListGames()[0]
		if after.Turn > before.Turn+1 {
			t.Fatalf("automation step %d advanced %d turns: before=%d after=%d", step, after.Turn-before.Turn, before.Turn, after.Turn)
		}
		if receipt.GameRevision != after.Revision {
			t.Fatalf("automation receipt revision=%d summary=%d", receipt.GameRevision, after.Revision)
		}
		if after.Phase == session.PhaseCompleted {
			completed = true
			break
		}
		if after.Phase != session.PhasePlanning || after.Turn != before.Turn+1 {
			t.Fatalf("automation step %d stopped at turn=%d phase=%q from turn=%d phase=%q", step, after.Turn, after.Phase, before.Turn, before.Phase)
		}
	}
	if !completed {
		dumpBuiltinAIDiagnostics(t, host, gameID, 1000, nil)
		t.Fatal("canonical builtin AI match did not complete within 1000 turns")
	}

	summary := host.ListGames()[0]
	if summary.Phase != session.PhaseCompleted || summary.Result == nil || summary.Result.Kind != session.ResultConquest {
		t.Fatalf("canonical builtin AI result=%+v summary=%+v", summary.Result, summary)
	}
	if summary.Turn > 1000 {
		t.Fatalf("canonical builtin AI completion turn=%d", summary.Turn)
	}

	hosted, err := host.lookup(gameID)
	if err != nil {
		t.Fatal(err)
	}
	observer, err := hosted.session.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	researchCompletions := 0
	for _, event := range observer.Events {
		if event.Kind == "empire.research_completed" {
			researchCompletions++
		}
	}
	if researchCompletions == 0 {
		t.Fatal("canonical builtin AI match contains no controller-neutral research completion events")
	}
	snapshot, err := hosted.session.MarshalCompletedSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func TestHumanVsBuiltinAIAutomaticTurnsAreExact(t *testing.T) {
	first := runHumanVsBuiltinAI(t)
	second := runHumanVsBuiltinAI(t)
	if !bytes.Equal(first, second) {
		t.Fatalf("human-vs-AI observer snapshots differ: first=%d second=%d", len(first), len(second))
	}
}

func runHumanVsBuiltinAI(t *testing.T) []byte {
	t.Helper()
	host := loadNewGameHost(t)
	const gameID = "human-vs-builtin"
	_, err := host.CreateGame(CreateGameRequest{
		GameID:   gameID,
		Seed:     0x8009,
		Settings: appNewGameSettings(),
		Controllers: []PlayerControllerSpec{
			{SeatID: 2, Controller: session.ControllerBuiltinAI},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	aiSnapshot, err := host.PlayerSnapshot(gameID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if aiSnapshot.View.Seat.Seat.Controller != session.ControllerBuiltinAI {
		t.Fatalf("seat 2 controller=%q", aiSnapshot.View.Seat.Seat.Controller)
	}

	for i := 0; i < 8; i++ {
		before, err := host.PlayerSnapshot(gameID, 1)
		if err != nil {
			t.Fatal(err)
		}
		if before.View.Phase != session.PhasePlanning {
			t.Fatalf("human boundary turn=%d phase=%q", before.View.Turn, before.View.Phase)
		}
		batch := protocol.CommandBatch{
			SchemaVersion: protocol.CommandSchemaVersion,
			GameID:        gameID,
			SeatID:        1,
			Turn:          before.View.Turn,
			BaseRevision:  before.View.Revision,
		}
		if _, err := host.SubmitTurn(gameID, batch); err != nil {
			t.Fatalf("human empty turn %d: %v", before.View.Turn, err)
		}
		after, err := host.PlayerSnapshot(gameID, 1)
		if err != nil {
			t.Fatal(err)
		}
		if after.View.Phase != session.PhasePlanning || after.View.Turn != before.View.Turn+1 {
			t.Fatalf("after human turn=%d phase=%q want turn=%d planning", after.View.Turn, after.View.Phase, before.View.Turn+1)
		}
	}

	aiSnapshot, err = host.PlayerSnapshot(gameID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if aiSnapshot.View.Empire.Research == nil {
		t.Fatal("builtin AI did not select research during automatic human-vs-AI turns")
	}
	observer, err := host.ObserverSnapshot(gameID)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(observer)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func dumpBuiltinAIDiagnostics(t *testing.T, host *Host, gameID string, step int, cause error) {
	t.Helper()
	hosted, err := host.lookup(gameID)
	if err != nil {
		t.Logf("AI diagnostics step=%d cause=%v lookup=%v", step, cause, err)
		return
	}
	observer, err := hosted.session.ObserverView()
	if err != nil {
		t.Logf("AI diagnostics step=%d cause=%v observer=%v", step, cause, err)
		return
	}
	t.Logf("AI diagnostics step=%d cause=%v turn=%d phase=%q revision=%d colonies=%d fleets=%d battles=%d", step, cause, observer.Turn, observer.Phase, observer.Revision, len(observer.State.Colonies), len(observer.State.StrategicFleets), len(observer.Battles))
	for _, empire := range observer.State.Empires {
		t.Logf("empire=%d capital=%d research=%+v tech_count=%d treasury=%.3f", empire.ID, empire.Capital, empire.Research, len(empire.KnownTechnologyIDs), empire.Treasury.BalanceBC)
	}
	for _, colony := range observer.State.Colonies {
		t.Logf("colony=%d empire=%d planet=%d population=%.3f construction=%+v", colony.ID, colony.EmpireID, colony.PlanetID, colony.Population.Total(), colony.Construction)
	}
	for _, fleet := range observer.State.StrategicFleets {
		t.Logf("fleet=%d empire=%d role=%s special=%s at=%d destination=%d remaining=%d ships=%v", fleet.ID, fleet.EmpireID, fleet.Role, fleet.SpecialKind, fleet.AtSystemID, fleet.DestinationSystemID, fleet.RemainingTurns, fleet.ShipIDs)
	}
}
func TestTriangleLaserScoutQueueDoesNotStallStrategicResolution(t *testing.T) {
	host := loadNewGameHost(t)
	const gameID = "triangle-laser-queue"
	generated, err := host.newGameRules.NewReferenceTriangleGame(game.ReferenceTriangleSeed)
	if err != nil {
		t.Fatal(err)
	}
	seats := make([]session.Seat, len(generated.Players))
	for i, player := range generated.Players {
		controller := session.ControllerBuiltinAI
		if player.SeatID == 1 {
			controller = session.ControllerLocalHuman
		}
		seats[i] = session.Seat{ID: player.SeatID, EmpireID: player.EmpireID, Name: player.Name, Controller: controller}
	}
	gameSession, err := session.NewGameSession(gameID, generated.State, seats)
	if err != nil {
		t.Fatal(err)
	}
	if err := host.Register(Registration{Session: gameSession, Resolver: host.newGameResolver, ImmediateResolver: host.newGameImmediateResolver}); err != nil {
		t.Fatal(err)
	}

	before, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if before.Decision == nil || len(before.Decision.Strategic.ShipDesigns) == 0 || len(before.View.Colonies) == 0 {
		t.Fatalf("triangle initial human snapshot incomplete: decision=%v colonies=%d", before.Decision != nil, len(before.View.Colonies))
	}
	design := before.Decision.Strategic.ShipDesigns[0]
	save, err := game.NewSaveMilitaryDesignCommand(1, game.SaveMilitaryDesignPayload{
		DesignID:           design.ID,
		Name:               design.Name,
		HullID:             design.Spec.HullID,
		StrategicPictureID: design.Spec.StrategicPictureID,
		Weapons:            []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := host.SubmitImmediateCommand(gameID, 1, before.View.Revision, save); err != nil {
		t.Fatalf("save Laser Scout design: %v", err)
	}
	afterDesign, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	var revised *core.ShipDesign
	for i := range afterDesign.Decision.Strategic.ShipDesigns {
		if afterDesign.Decision.Strategic.ShipDesigns[i].ID == design.ID {
			revised = &afterDesign.Decision.Strategic.ShipDesigns[i]
			break
		}
	}
	if revised == nil || revised.Revision != 2 || len(revised.Spec.Weapons) != 1 || revised.Spec.Weapons[0].WeaponID != "laser_cannon" {
		t.Fatalf("revised Laser Scout=%+v", revised)
	}

	colonyID := afterDesign.View.Colonies[0].ID
	queue, err := game.NewSetConstructionQueueCommand(1, game.SetConstructionQueuePayload{ColonyID: colonyID, Items: []game.ConstructionQueueItem{{
		ProjectKind: core.ConstructionProjectMilitaryShip, ProjectID: game.MilitaryShipProjectID,
		ShipDesignID: revised.ID, ShipDesignRevision: revised.Revision,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	batch := protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        gameID, SeatID: 1, Turn: afterDesign.View.Turn, BaseRevision: afterDesign.View.Revision,
		Commands: []protocol.Command{queue},
	}
	if _, err := host.SubmitTurn(gameID, batch); err != nil {
		t.Fatalf("submit Laser Scout construction turn: %v", err)
	}
	afterTurn, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if afterTurn.View.Phase == session.PhaseStrategicResolution {
		t.Fatalf("triangle stuck in strategic_resolution after Laser Scout queue: %+v", afterTurn.View)
	}
	var humanColony *core.Colony
	for i := range afterTurn.View.Colonies {
		if afterTurn.View.Colonies[i].ID == colonyID {
			humanColony = &afterTurn.View.Colonies[i]
			break
		}
	}
	if humanColony == nil || humanColony.Construction == nil || humanColony.Construction.ProjectKind != core.ConstructionProjectMilitaryShip || humanColony.Construction.ShipDesignID != revised.ID || humanColony.Construction.ShipDesignRevision != revised.Revision {
		t.Fatalf("Laser Scout construction not committed after turn: colony=%+v", humanColony)
	}
}
