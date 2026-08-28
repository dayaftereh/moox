package game

import (
	"encoding/json"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func TestStrategicApplyOrderUsesPreGrowthResearchAndConstructionSnapshots(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(760)
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}

	queue, err := NewQueueBuildingCommand(1, QueueBuildingPayload{
		ColonyID:   state.Colonies[0].ID,
		BuildingID: "marine_barracks",
	})
	if err != nil {
		t.Fatal(err)
	}
	selectResearch, err := NewSelectResearchCommand(2, SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	ctx := ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: state.Empires[0].ID}}}
	result, err := resolver.Resolve(ctx, state, []protocol.CommandBatch{{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "turn-order",
		SeatID:        1,
		Turn:          1,
		BaseRevision:  1,
		Commands:      []protocol.Command{queue, selectResearch},
	}})
	if err != nil {
		t.Fatal(err)
	}

	wantKinds := []string{
		"colony.construction_queued",
		"empire.research_selected",
		"empire.treasury_settled",
		"empire.research_progressed",
		"colony.population_grew",
		"colony.construction_progressed",
	}
	if len(result.Events) != len(wantKinds) {
		t.Fatalf("events=%d want=%d: %+v", len(result.Events), len(wantKinds), result.Events)
	}
	for i, want := range wantKinds {
		if result.Events[i].Kind != want {
			t.Fatalf("event[%d]=%q want=%q; events=%+v", i, result.Events[i].Kind, want, result.Events)
		}
	}

	var research ResearchProgressedEvent
	if err := json.Unmarshal(result.Events[3].Data, &research); err != nil {
		t.Fatal(err)
	}
	if research.TurnResearchRP != 4.5 || research.ProjectedRP != 4.5 {
		t.Fatalf("research used non-pre-growth output: %+v", research)
	}

	var construction ConstructionProgressedEvent
	if err := json.Unmarshal(result.Events[5].Data, &construction); err != nil {
		t.Fatal(err)
	}
	if construction.AppliedPP != 3 {
		t.Fatalf("construction used non-pre-growth output: %+v", construction)
	}
	colony := result.State.Colonies[0]
	if colony.Population.Total <= 4 {
		t.Fatalf("expected Population growth before construction apply, got %+v", colony.Population)
	}
	if colony.AdjustedEconomy.Research <= research.TurnResearchRP || colony.AdjustedEconomy.Production <= construction.AppliedPP {
		t.Fatalf("post-turn snapshot should exceed consumed pre-growth output: adjusted=%+v research=%v production=%v", colony.AdjustedEconomy, research.TurnResearchRP, construction.AppliedPP)
	}
}
