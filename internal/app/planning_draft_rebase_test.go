package app

import (
	"encoding/json"
	"fmt"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

func newPlanningDraftRebaseTriangle(t *testing.T, gameID string) (*Host, PlayerSnapshot) {
	t.Helper()
	host := loadNewGameHost(t)
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
	snapshot, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	return host, snapshot
}

func populationDraftOrder(t *testing.T, colonyID core.ID, farmers, workers, scientists float64) PlanningDraftOrder {
	t.Helper()
	command, err := game.NewAssignPopulationCommand(1, game.AssignPopulationPayload{
		ColonyID: colonyID, Farmers: farmers, Workers: workers, Scientists: scientists,
	})
	if err != nil {
		t.Fatal(err)
	}
	return PlanningDraftOrder{Key: fmt.Sprintf("population:%d", colonyID), Kind: command.Kind, Payload: append(json.RawMessage(nil), command.Payload...)}
}

func TestPlanningDraftRebasesAcrossImmediateMilitaryDesignRevision(t *testing.T) {
	const gameID = "planning-draft-rebase"
	host, before := newPlanningDraftRebaseTriangle(t, gameID)
	if before.Decision == nil || len(before.View.Colonies) == 0 || len(before.Decision.Strategic.ShipDesigns) == 0 {
		t.Fatalf("initial snapshot incomplete")
	}
	colony := before.View.Colonies[0]
	draft := PlanningDraft{
		SchemaVersion: SchemaVersion,
		GameID:        gameID,
		SeatID:        1,
		Turn:          before.View.Turn,
		BaseRevision:  before.View.Revision,
		DraftRevision: 1,
		Orders:        []PlanningDraftOrder{populationDraftOrder(t, colony.ID, 5, 3, 0)},
	}
	saved, err := host.SavePlanningDraft(gameID, 1, draft)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Draft.BaseRevision != before.View.Revision {
		t.Fatalf("initial draft base revision=%d want %d", saved.Draft.BaseRevision, before.View.Revision)
	}

	design := before.Decision.Strategic.ShipDesigns[0]
	saveDesign, err := game.NewSaveMilitaryDesignCommand(1, game.SaveMilitaryDesignPayload{
		DesignID:           design.ID,
		Name:               design.Name,
		HullID:             design.Spec.HullID,
		StrategicPictureID: design.Spec.StrategicPictureID,
		Weapons:            []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := host.SubmitImmediateCommand(gameID, 1, before.View.Revision, saveDesign)
	if err != nil {
		t.Fatalf("immediate design save: %v", err)
	}
	if receipt.GameRevision <= before.View.Revision {
		t.Fatalf("design save did not advance revision: before=%d after=%d", before.View.Revision, receipt.GameRevision)
	}

	afterDesign, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if afterDesign.PlanningDraft == nil || afterDesign.PlanningDraft.BaseRevision != afterDesign.View.Revision {
		t.Fatalf("planning draft was not rebased after gameplay design save: %+v", afterDesign.PlanningDraft)
	}
	var revised *core.ShipDesign
	for i := range afterDesign.Decision.Strategic.ShipDesigns {
		if afterDesign.Decision.Strategic.ShipDesigns[i].ID == design.ID {
			revised = &afterDesign.Decision.Strategic.ShipDesigns[i]
			break
		}
	}
	if revised == nil || revised.VisualGenome == nil {
		t.Fatalf("revised design missing visual genome: %+v", revised)
	}
	visual, err := game.NewSetMilitaryDesignVisualCommand(1, game.SetMilitaryDesignVisualPayload{DesignID: revised.ID, VisualGenome: *revised.VisualGenome})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := host.SubmitImmediateCommand(gameID, 1, afterDesign.View.Revision, visual); err != nil {
		t.Fatalf("immediate visual save: %v", err)
	}

	after, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if after.PlanningDraft == nil {
		t.Fatal("planning draft was lost across gameplay + visual design revisions")
	}
	if after.PlanningDraft.BaseRevision != after.View.Revision {
		t.Fatalf("rebased draft base revision=%d want current %d", after.PlanningDraft.BaseRevision, after.View.Revision)
	}
	if len(after.PlanningDraft.Orders) != 1 || after.PlanningDraft.Orders[0].Key != draft.Orders[0].Key {
		t.Fatalf("rebased draft orders=%+v want population order", after.PlanningDraft.Orders)
	}
	preview, err := host.PlanningPreview(gameID, 1, mustPlanningBatch(t, *after.PlanningDraft))
	if err != nil {
		t.Fatalf("rebased draft no longer previews: %v", err)
	}
	got := preview.Preview.Projection.Colonies[0].Colony.Population.Cohorts[0]
	if got.Farmers != 5 || got.Workers != 3 || got.Scientists != 0 {
		t.Fatalf("rebased population preview=%+v want 5/3/0", got)
	}
}

func TestStalePlanningDraftSaveRebasesAndInvalidStaleDraftDoesNotOverwrite(t *testing.T) {
	const gameID = "planning-draft-stale-save"
	host, before := newPlanningDraftRebaseTriangle(t, gameID)
	colony := before.View.Colonies[0]
	design := before.Decision.Strategic.ShipDesigns[0]

	saveDesign, err := game.NewSaveMilitaryDesignCommand(1, game.SaveMilitaryDesignPayload{
		DesignID: design.ID, Name: design.Name, HullID: design.Spec.HullID, StrategicPictureID: design.Spec.StrategicPictureID,
		Weapons: []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := host.SubmitImmediateCommand(gameID, 1, before.View.Revision, saveDesign)
	if err != nil {
		t.Fatal(err)
	}

	stale := PlanningDraft{
		SchemaVersion: SchemaVersion, GameID: gameID, SeatID: 1, Turn: before.View.Turn,
		BaseRevision: before.View.Revision, DraftRevision: 2,
		Orders: []PlanningDraftOrder{populationDraftOrder(t, colony.ID, 5, 3, 0)},
	}
	saved, err := host.SavePlanningDraft(gameID, 1, stale)
	if err != nil {
		t.Fatalf("valid stale planning draft should rebase: %v", err)
	}
	if saved.Draft.BaseRevision != receipt.GameRevision {
		t.Fatalf("stale draft rebased to %d want %d", saved.Draft.BaseRevision, receipt.GameRevision)
	}

	zeroBase := stale
	zeroBase.BaseRevision = 0
	zeroBase.DraftRevision = 3
	if _, err := host.SavePlanningDraft(gameID, 1, zeroBase); err == nil {
		t.Fatal("zero-base planning draft unexpectedly accepted through rebase")
	}

	invalidOrder := populationDraftOrder(t, core.ID(999999), 5, 3, 0)
	invalid := stale
	invalid.DraftRevision = 3
	invalid.Orders = []PlanningDraftOrder{invalidOrder}
	if _, err := host.SavePlanningDraft(gameID, 1, invalid); err == nil {
		t.Fatal("invalid stale draft unexpectedly accepted")
	}
	after, err := host.PlayerSnapshot(gameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if after.PlanningDraft == nil || after.PlanningDraft.DraftRevision != 2 || len(after.PlanningDraft.Orders) != 1 || after.PlanningDraft.Orders[0].Key != stale.Orders[0].Key {
		t.Fatalf("invalid stale draft overwrote valid rebased draft: %+v", after.PlanningDraft)
	}
}

func mustPlanningBatch(t *testing.T, draft PlanningDraft) protocol.CommandBatch {
	t.Helper()
	batch, err := draft.commandBatch()
	if err != nil {
		t.Fatal(err)
	}
	return batch
}
