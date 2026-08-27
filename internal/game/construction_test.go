package game

import (
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
	colonyID := state.Colonies[0].ID
	command, err := NewQueueBuildingCommand(1, QueueBuildingPayload{ColonyID: colonyID, BuildingID: "holo_simulator"})
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
	if colony.Construction.ProgressMilli != colony.AdjustedEconomy.ProductionMilli {
		t.Fatalf("progress=%d adjusted production=%d", colony.Construction.ProgressMilli, colony.AdjustedEconomy.ProductionMilli)
	}
	if colony.Construction.ProgressMilli <= 0 || colony.Construction.ProgressMilli >= rules.BuildingDefinitions["holo_simulator"].ProductionCostMilli {
		t.Fatalf("unexpected first-turn construction progress: %+v", colony.Construction)
	}
	if len(result.Events) != 2 || result.Events[0].Kind != "colony.construction_queued" || result.Events[1].Kind != "colony.construction_progressed" {
		t.Fatalf("unexpected construction events: %+v", result.Events)
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
		state.Colonies[0].Construction = &core.ConstructionState{BuildingID: "research_lab", ProgressMilli: 1000}
		command := makeCommand(t, state, "holo_simulator")
		ctx, batches := constructionBatch(t, state, command)
		if _, err := resolver.Resolve(ctx, state, batches); err == nil {
			t.Fatal("expected second active construction to fail")
		}
	})
}

func TestCompletedBuildingAffectsEconomyOnNextRecalculation(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	// Test-only acceleration: the original Holo Simulator still costs 120 PP,
	// but one abundant-world worker produces enough to finish it this fixture turn.
	rules.MineralIndustryPerWorkerMilli["abundant"] = 120 * core.EconomyScale
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
	if colony.EconomyContext.MoraleBuildingBonusPercent != 0 {
		t.Fatalf("newly completed Holo affected same-turn economy: %+v", colony.EconomyContext)
	}
	if colony.AdjustedEconomy.ProductionMilli != 120*core.EconomyScale {
		t.Fatalf("same-turn production=%d", colony.AdjustedEconomy.ProductionMilli)
	}
	if len(first.Events) != 3 || first.Events[2].Kind != "colony.building_completed" {
		t.Fatalf("unexpected completion events: %+v", first.Events)
	}

	second, err := resolver.Resolve(ctx, first.State, nil)
	if err != nil {
		t.Fatal(err)
	}
	colony = second.State.Colonies[0]
	if colony.EconomyContext.MoraleBuildingBonusPercent != 20 || colony.EconomyContext.MoralePercent != 20 {
		t.Fatalf("completed Holo missing on next recalculation: %+v", colony.EconomyContext)
	}
	if colony.AdjustedEconomy.ProductionMilli != 144*core.EconomyScale {
		t.Fatalf("next-turn morale-adjusted production=%d, want %d", colony.AdjustedEconomy.ProductionMilli, 144*core.EconomyScale)
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
	firstProgress := first.State.Colonies[0].Construction.ProgressMilli
	second, err := resolver.Resolve(ctx, first.State, nil)
	if err != nil {
		t.Fatal(err)
	}
	secondProgress := second.State.Colonies[0].Construction.ProgressMilli
	if secondProgress != firstProgress+second.State.Colonies[0].AdjustedEconomy.ProductionMilli {
		t.Fatalf("progress did not advance deterministically: first=%d second=%d production=%d", firstProgress, secondProgress, second.State.Colonies[0].AdjustedEconomy.ProductionMilli)
	}
}
