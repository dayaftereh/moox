package game

import (
	"encoding/json"
	"math"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func TestConstructionBuyoutCostBCOriginalCurve(t *testing.T) {
	for _, tc := range []struct {
		name     string
		progress float64
		want     float64
	}{
		{"zero", 0, 400},
		{"ten-percent", 10, 300},
		{"thirty-percent", 30, 200},
		{"half", 50, 100},
		{"three-quarters", 75, 50},
		{"complete", 100, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ConstructionBuyoutCostBC(100, tc.progress)
			if err != nil {
				t.Fatal(err)
			}
			if math.Abs(got-tc.want) > 1e-9 {
				t.Fatalf("cost=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestConstructionBuyoutDeductsBCAndCompletesOnNormalTurn(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(17801)
	empire := &state.Empires[0]
	empire.Treasury.BalanceBC = 10_000
	colony := &state.Colonies[0]
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectBuilding, ProjectID: "holo_simulator", ProgressPP: 8}

	quotes, err := resolver.ConstructionBuyoutQuotes(state, empire.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 || quotes[0].ColonyID != colony.ID || quotes[0].CostBC <= 0 || !quotes[0].Affordable {
		t.Fatalf("quote=%+v", quotes)
	}
	quote := quotes[0]
	beforeBC := empire.Treasury.BalanceBC
	command, err := NewBuyConstructionCommand(1, BuyConstructionPayload{ColonyID: colony.ID})
	if err != nil {
		t.Fatal(err)
	}
	events, err := resolver.ResolveConstructionBuyoutCommand(state, empire.ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != "colony.construction_bought" {
		t.Fatalf("events=%+v", events)
	}
	if math.Abs(empire.Treasury.BalanceBC-(beforeBC-quote.CostBC)) > 1e-9 {
		t.Fatalf("treasury=%v want=%v", empire.Treasury.BalanceBC, beforeBC-quote.CostBC)
	}
	if colony.Construction == nil || math.Abs(colony.Construction.ProgressPP-quote.ProductionCostPP) > 1e-9 {
		t.Fatalf("bought construction=%+v quote=%+v", colony.Construction, quote)
	}
	var bought ConstructionBoughtEvent
	if err := json.Unmarshal(events[0].Data, &bought); err != nil {
		t.Fatal(err)
	}
	if bought.CostBC != quote.CostBC || bought.PreviousProgressPP != 8 || bought.CurrentProgressPP != quote.ProductionCostPP {
		t.Fatalf("bought event=%+v quote=%+v", bought, quote)
	}

	ctx := ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: empire.ID}}}
	result, err := resolver.Resolve(ctx, state, []protocol.CommandBatch{{
		SchemaVersion: protocol.CommandSchemaVersion, GameID: "game-buyout", SeatID: 1, Turn: state.Turn, BaseRevision: 1,
		Commands: []protocol.Command{},
	}})
	if err != nil {
		t.Fatal(err)
	}
	completed := false
	for _, event := range result.Events {
		if event.Kind == "colony.building_completed" {
			completed = true
		}
	}
	if !completed {
		t.Fatalf("normal turn did not complete bought construction: %+v", result.Events)
	}
	if result.State.Colonies[0].Construction != nil {
		t.Fatalf("construction remains after normal turn: %+v", result.State.Colonies[0].Construction)
	}
	found := false
	for _, building := range result.State.Colonies[0].Buildings {
		if building == "holo_simulator" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("bought holo_simulator not completed: %+v", result.State.Colonies[0].Buildings)
	}
}

func TestConstructionBuyoutRejectsInsufficientTreasuryWithoutMutation(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(17802)
	empire := &state.Empires[0]
	empire.Treasury.BalanceBC = 1
	colony := &state.Colonies[0]
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectBuilding, ProjectID: "holo_simulator"}
	command, err := NewBuyConstructionCommand(1, BuyConstructionPayload{ColonyID: colony.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.ResolveConstructionBuyoutCommand(state, empire.ID, 1, command); err == nil {
		t.Fatal("insufficient treasury buyout unexpectedly succeeded")
	}
	if empire.Treasury.BalanceBC != 1 || colony.Construction.ProgressPP != 0 {
		t.Fatalf("failed buyout mutated state treasury=%v construction=%+v", empire.Treasury.BalanceBC, colony.Construction)
	}
}

func TestConstructionBuyoutSkipsHousing(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(17803)
	state.Colonies[0].Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectHousing, ProjectID: HousingProjectID}
	quotes, err := resolver.ConstructionBuyoutQuotes(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 0 {
		t.Fatalf("housing buyout quotes=%+v", quotes)
	}
}
