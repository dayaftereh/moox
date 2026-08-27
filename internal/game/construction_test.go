package game

import (
	"encoding/json"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func constructionBatch(t *testing.T, state *core.GameState, command protocol.Command) (ResolveContext, []protocol.CommandBatch) {
	t.Helper()
	return ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: state.Empires[0].ID}}}, []protocol.CommandBatch{{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "game-1",
		SeatID:        1,
		Turn:          state.Turn,
		BaseRevision:  1,
		Commands:      []protocol.Command{command},
	}}
}

func TestQueueBuildingAppliesCurrentTurnProduction(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(201)
	state.Empires[0].KnownTechnologyIDs = []int{86}
	command, err := NewQueueBuildingCommand(1, QueueBuildingPayload{ColonyID: state.Colonies[0].ID, BuildingID: "holo_simulator"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, batches := constructionBatch(t, state, command)
	result, err := resolver.Resolve(ctx, state, batches)
	if err != nil {
		t.Fatal(err)
	}
	colony := result.State.Colonies[0]
	if colony.Construction == nil || colony.Construction.BuildingID != "holo_simulator" {
		t.Fatalf("construction not queued: %+v", colony.Construction)
	}
	if colony.Construction.ProgressPP != 3 {
		t.Fatalf("current-turn progress=%v want=3 PP", colony.Construction.ProgressPP)
	}
	if colony.AdjustedEconomy.Production <= colony.Construction.ProgressPP {
		t.Fatalf("post-growth production=%v should exceed consumed current-turn PP=%v", colony.AdjustedEconomy.Production, colony.Construction.ProgressPP)
	}
	if len(result.Events) != 3 || result.Events[0].Kind != "colony.construction_queued" || result.Events[1].Kind != "colony.construction_progressed" || result.Events[2].Kind != "colony.population_grew" {
		t.Fatalf("unexpected construction/growth events: %+v", result.Events)
	}
	var progressed ConstructionProgressedEvent
	if err := json.Unmarshal(result.Events[1].Data, &progressed); err != nil {
		t.Fatal(err)
	}
	if progressed.AppliedPP != 3 {
		t.Fatalf("applied current-turn PP=%v want=3", progressed.AppliedPP)
	}
}

func TestQueueBuildingRejectsForeignUnknownOwnedAndBusy(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	makeCommand := func(t *testing.T, state *core.GameState, buildingID string) protocol.Command {
		t.Helper()
		command, err := NewQueueBuildingCommand(1, QueueBuildingPayload{ColonyID: state.Colonies[0].ID, BuildingID: buildingID})
		if err != nil {
			t.Fatal(err)
		}
		return command
	}

	t.Run("foreign colony", func(t *testing.T) {
		state := core.NewSmallFixture(202)
		command := makeCommand(t, state, "holo_simulator")
		ctx := ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: state.Empires[0].ID + 100}}}
		_, err := resolver.Resolve(ctx, state, []protocol.CommandBatch{{SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-1", SeatID: 1, Turn: 1, BaseRevision: 1, Commands: []protocol.Command{command}}})
		if err == nil {
			t.Fatal("expected foreign colony construction to fail")
		}
	})

	t.Run("missing technology", func(t *testing.T) {
		state := core.NewSmallFixture(208)
		command := makeCommand(t, state, "holo_simulator")
		ctx, batches := constructionBatch(t, state, command)
		if _, err := resolver.Resolve(ctx, state, batches); err == nil {
			t.Fatal("expected unknown technology to block construction")
		}
	})
	t.Run("unknown building", func(t *testing.T) {
		state := core.NewSmallFixture(203)
		command := makeCommand(t, state, "not_a_building")
		ctx, batches := constructionBatch(t, state, command)
		if _, err := resolver.Resolve(ctx, state, batches); err == nil {
			t.Fatal("expected unknown building to fail")
		}
	})

	t.Run("already owned", func(t *testing.T) {
		state := core.NewSmallFixture(204)
		state.Empires[0].KnownTechnologyIDs = []int{86}
		state.Colonies[0].Buildings = []string{"holo_simulator"}
		command := makeCommand(t, state, "holo_simulator")
		ctx, batches := constructionBatch(t, state, command)
		if _, err := resolver.Resolve(ctx, state, batches); err == nil {
			t.Fatal("expected already-owned building to fail")
		}
	})

	t.Run("busy colony", func(t *testing.T) {
		state := core.NewSmallFixture(205)
		state.Empires[0].KnownTechnologyIDs = []int{86}
		state.Colonies[0].Construction = &core.ConstructionState{BuildingID: "research_lab", ProgressPP: 1}
		command := makeCommand(t, state, "holo_simulator")
		ctx, batches := constructionBatch(t, state, command)
		if _, err := resolver.Resolve(ctx, state, batches); err == nil {
			t.Fatal("expected second active construction to fail")
		}
	})
}

func TestCompletedBuildingAffectsPostTurnSnapshotWithoutRetroactivePP(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	rules.MineralIndustryPerWorker["abundant"] = 120
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(206)
	state.Empires[0].KnownTechnologyIDs = []int{86}
	command, err := NewQueueBuildingCommand(1, QueueBuildingPayload{ColonyID: state.Colonies[0].ID, BuildingID: "holo_simulator"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, batches := constructionBatch(t, state, command)
	first, err := resolver.Resolve(ctx, state, batches)
	if err != nil {
		t.Fatal(err)
	}
	colony := first.State.Colonies[0]
	if colony.Construction != nil {
		t.Fatalf("completed building left construction active: %+v", colony.Construction)
	}
	if len(colony.Buildings) != 1 || colony.Buildings[0] != "holo_simulator" {
		t.Fatalf("completed building missing: %+v", colony.Buildings)
	}
	if colony.EconomyContext.MoraleBuildingBonusPercent != 20 || colony.EconomyContext.MoralePercent != 20 {
		t.Fatalf("completed Holo missing from post-turn snapshot: %+v", colony.EconomyContext)
	}
	if colony.AdjustedEconomy.Production <= 144 {
		t.Fatalf("post-growth/post-building production=%v should exceed 144", colony.AdjustedEconomy.Production)
	}
	if len(first.Events) != 4 || first.Events[2].Kind != "colony.building_completed" || first.Events[3].Kind != "colony.population_grew" {
		t.Fatalf("unexpected completion/growth events: %+v", first.Events)
	}
	var progress ConstructionProgressedEvent
	if err := json.Unmarshal(first.Events[1].Data, &progress); err != nil {
		t.Fatal(err)
	}
	if progress.AppliedPP != 120 {
		t.Fatalf("completed building consumed %v PP, want exactly current-turn 120", progress.AppliedPP)
	}
	second, err := resolver.Resolve(ctx, first.State, nil)
	if err != nil {
		t.Fatal(err)
	}
	colony = second.State.Colonies[0]
	if colony.EconomyContext.MoraleBuildingBonusPercent != 20 || colony.EconomyContext.MoralePercent != 20 {
		t.Fatalf("completed Holo missing on subsequent recalculation: %+v", colony.EconomyContext)
	}
}

func TestConstructionProgressIsDeterministicAcrossResolutions(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(207)
	state.Empires[0].KnownTechnologyIDs = []int{86}
	command, err := NewQueueBuildingCommand(1, QueueBuildingPayload{ColonyID: state.Colonies[0].ID, BuildingID: "holo_simulator"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, batches := constructionBatch(t, state, command)
	first, err := resolver.Resolve(ctx, state, batches)
	if err != nil {
		t.Fatal(err)
	}
	firstProgress := first.State.Colonies[0].Construction.ProgressPP
	second, err := resolver.Resolve(ctx, first.State, nil)
	if err != nil {
		t.Fatal(err)
	}
	secondProgress := second.State.Colonies[0].Construction.ProgressPP
	if len(second.Events) == 0 || second.Events[0].Kind != "colony.construction_progressed" {
		t.Fatalf("second resolution missing construction progress event: %+v", second.Events)
	}
	var secondEvent ConstructionProgressedEvent
	if err := json.Unmarshal(second.Events[0].Data, &secondEvent); err != nil {
		t.Fatal(err)
	}
	if !closePopulationValue(secondProgress, firstProgress+secondEvent.AppliedPP) {
		t.Fatalf("progress did not advance deterministically: first=%v second=%v applied=%v", firstProgress, secondProgress, secondEvent.AppliedPP)
	}
	if second.State.Colonies[0].AdjustedEconomy.Production <= secondEvent.AppliedPP {
		t.Fatalf("post-growth output=%v should exceed already-consumed turn PP=%v", second.State.Colonies[0].AdjustedEconomy.Production, secondEvent.AppliedPP)
	}
}

func TestCyberneticSustenanceReducesConstructionProduction(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(721)
	state.Empires[0].RaceID = "meklar"
	state.Empires[0].KnownTechnologyIDs = []int{86}
	command, err := NewQueueBuildingCommand(1, QueueBuildingPayload{ColonyID: state.Colonies[0].ID, BuildingID: "holo_simulator"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, batches := constructionBatch(t, state, command)
	result, err := resolver.Resolve(ctx, state, batches)
	if err != nil {
		t.Fatal(err)
	}
	progressEvent := findDomainEvent(result.Events, "colony.construction_progressed")
	if progressEvent == nil {
		t.Fatalf("missing construction progress event: %+v", result.Events)
	}
	var progress ConstructionProgressedEvent
	if err := json.Unmarshal(progressEvent.Data, &progress); err != nil {
		t.Fatal(err)
	}
	// One Meklar worker produces 5 PP base. Missing-Barracks Dictatorship morale
	// reduces that to 4 PP; Cybernetic sustenance consumes 0.5 PP * 4 pop = 2 PP.
	if progress.AppliedPP != 2 {
		t.Fatalf("Cybernetic construction applied_pp=%v want=2", progress.AppliedPP)
	}

}
