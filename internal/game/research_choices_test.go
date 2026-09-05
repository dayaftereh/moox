package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func TestAvailableResearchChoicesReturnsServerAuthoritativeFrontier(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(701)
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}

	choices, err := rules.AvailableResearchChoices(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	gotFields := make([]int, len(choices))
	for i, choice := range choices {
		gotFields[i] = choice.TechFieldID
		if choice.BaseCostRP <= 0 || len(choice.TechnologyIDs) == 0 || len(choice.TechnologyIDs) != len(choice.TechnologyKeys) || len(choice.TechnologyIDs) != len(choice.TechnologyNameKeys) {
			t.Fatalf("incomplete research choice: %+v", choice)
		}
	}
	wantFields := []int{4, 7, 10, 18, 22, 28, 55, 57}
	if !reflect.DeepEqual(gotFields, wantFields) {
		t.Fatalf("research frontier=%v want=%v", gotFields, wantFields)
	}

	choice, ok := researchChoiceByField(choices, 4)
	if !ok {
		t.Fatal("expected TechField 4 in Pre-Warp frontier")
	}
	if choice.PreviousTechFieldID != 29 || choice.BaseCostRP != 80 {
		t.Fatalf("TechField 4 metadata=%+v", choice)
	}
	wantKeys := []string{"anti_missile_rockets", "reinforced_hull", "fighter_bays"}
	if !reflect.DeepEqual(choice.TechnologyKeys, wantKeys) {
		t.Fatalf("TechField 4 keys=%v want=%v", choice.TechnologyKeys, wantKeys)
	}
}

func TestAvailableResearchChoicesRemainAvailableWhileResearchActive(t *testing.T) {
	rules, resolver, state := initializedResearchRace(t, 702, "human")
	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command); err != nil {
		t.Fatal(err)
	}
	state.Empires[0].Research.ProgressRP = 12.5
	choices, err := rules.AvailableResearchChoices(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(choices) == 0 {
		t.Fatal("active research must still expose legal switching choices")
	}
	if _, ok := researchChoiceByField(choices, 4); !ok {
		t.Fatalf("active TechField 4 missing from switching choices: %+v", choices)
	}
}

func TestSelectResearchCommandMaterializesTechnologySetServerSide(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(703)
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != "empire.research_selected" {
		t.Fatalf("event kind=%q", event.Kind)
	}
	if state.Empires[0].Research == nil || state.Empires[0].Research.TechFieldID != 4 || state.Empires[0].Research.ProgressRP != 0 {
		t.Fatalf("research state=%+v", state.Empires[0].Research)
	}
	wantIDs := []int{56}
	if !reflect.DeepEqual(state.Empires[0].Research.TechnologyIDs, wantIDs) {
		t.Fatalf("server selected ids=%v want=%v", state.Empires[0].Research.TechnologyIDs, wantIDs)
	}
	var payload ResearchSelectedEvent
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.TechFieldID != 4 || payload.BaseCostRP != 80 || !reflect.DeepEqual(payload.TechnologyIDs, wantIDs) {
		t.Fatalf("selection event=%+v", payload)
	}
}

func TestSelectResearchRejectsNonFrontierField(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(704)
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 56})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.selectResearch(state, state.Empires[0].ID, 1, command); err == nil {
		t.Fatal("expected TechField 56 to be rejected before prerequisite TechField 28 is completed")
	}
}

func TestStrategicResolverSelectsResearchAndAppliesFractionalTurnRP(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(705)
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	state.Colonies[0].Population = core.NewAssimilatedPopulation(state.Empires[0].ID, 1.25, 1.5, 1.25)
	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	ctx := ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: state.Empires[0].ID}}}
	result, err := resolver.Resolve(ctx, state, []protocol.CommandBatch{{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "research-choice-test",
		SeatID:        1,
		Turn:          1,
		BaseRevision:  1,
		Commands:      []protocol.Command{command},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if result.State.Empires[0].Research == nil {
		t.Fatal("research selection unexpectedly completed")
	}
	// Research consumes the current-turn pre-Population-transition output. The
	// post-resolution colony snapshot may differ after growth/starvation.
	if result.State.Empires[0].Research.ProgressRP != 5.625 {
		t.Fatalf("progress_rp=%v want pre-transition 5.625", result.State.Empires[0].Research.ProgressRP)
	}
}

func TestResearchUsesPreGrowthTurnOutput(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(706)
	if err := rules.InitializeEmpireTechnologies(&state.Empires[0], NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	// The default Human colony is exactly fed and therefore grows at turn end.
	// Its one scientist yields 3 RP base / 4.5 RP with Democracy before growth.
	command, err := NewSelectResearchCommand(1, SelectResearchPayload{TechFieldID: 4, TechnologyID: 56})
	if err != nil {
		t.Fatal(err)
	}
	ctx := ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: state.Empires[0].ID}}}
	result, err := resolver.Resolve(ctx, state, []protocol.CommandBatch{{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        "research-before-growth",
		SeatID:        1,
		Turn:          1,
		BaseRevision:  1,
		Commands:      []protocol.Command{command},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if result.State.Empires[0].Research == nil {
		t.Fatal("research project unexpectedly completed")
	}
	if !closePopulationValue(result.State.Empires[0].Research.ProgressRP, 4.5) {
		t.Fatalf("same-turn research progress=%v want pre-growth 4.5 RP", result.State.Empires[0].Research.ProgressRP)
	}
	colony := result.State.Colonies[0]
	if colony.Population.Total() <= 4 {
		t.Fatalf("expected turn-end population growth, got %+v", colony.Population)
	}
	if colony.AdjustedEconomy.Research <= 4.5 {
		t.Fatalf("post-growth research output=%v should exceed consumed turn RP=4.5", colony.AdjustedEconomy.Research)
	}
	foundGrowth := false
	for _, event := range result.Events {
		if event.Kind == "colony.population_grew" {
			foundGrowth = true
			break
		}
	}
	if !foundGrowth {
		t.Fatalf("missing population growth event: %+v", result.Events)
	}
}

func TestResearchCategoriesCoverEightPlayerFacingChains(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	want := []ResearchCategory{
		{ID: "construction", Order: 0, NameKey: "research.category.construction", RootTechFieldID: 29},
		{ID: "chemistry", Order: 1, NameKey: "research.category.chemistry", RootTechFieldID: 22},
		{ID: "computers", Order: 2, NameKey: "research.category.computers", RootTechFieldID: 28},
		{ID: "physics", Order: 3, NameKey: "research.category.physics", RootTechFieldID: 57},
		{ID: "power", Order: 4, NameKey: "research.category.power", RootTechFieldID: 55},
		{ID: "sociology", Order: 5, NameKey: "research.category.sociology", RootTechFieldID: 10},
		{ID: "biology", Order: 6, NameKey: "research.category.biology", RootTechFieldID: 18},
		{ID: "force_fields", Order: 7, NameKey: "research.category.force_fields", RootTechFieldID: 7},
	}
	got := rules.AvailableResearchCategories()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("research categories=%+v want=%+v", got, want)
	}
	categorized := 0
	for fieldID := range rules.TechnologyFieldCostsRP {
		categoryID := rules.TechnologyFieldCategoryID[fieldID]
		if fieldID == 74 {
			if categoryID != "" {
				t.Fatalf("special Antaran field 74 category=%q want empty", categoryID)
			}
			continue
		}
		if categoryID == "" || rules.TechnologyFieldCategoryNameKey[fieldID] == "" {
			t.Fatalf("ordinary field %d has incomplete category metadata", fieldID)
		}
		order, ok := rules.TechnologyFieldCategoryOrder[fieldID]
		if !ok || order < 0 || order > 7 {
			t.Fatalf("ordinary field %d category order=(%d,%v)", fieldID, order, ok)
		}
		categorized++
	}
	if categorized != 81 {
		t.Fatalf("categorized ordinary/hyper fields=%d want=81", categorized)
	}

	state := core.NewSmallFixture(0xCA7E)
	choices, err := rules.AvailableResearchChoices(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, choice := range choices {
		if choice.TechFieldID == 74 {
			t.Fatalf("special Antaran field 74 exposed as ordinary research choice: %+v", choice)
		}
		if choice.CategoryID != rules.TechnologyFieldCategoryID[choice.TechFieldID] || choice.CategoryOrder != rules.TechnologyFieldCategoryOrder[choice.TechFieldID] || choice.CategoryNameKey != rules.TechnologyFieldCategoryNameKey[choice.TechFieldID] {
			t.Fatalf("choice field %d category=%q/%d/%q does not match rules metadata", choice.TechFieldID, choice.CategoryID, choice.CategoryOrder, choice.CategoryNameKey)
		}
	}
}
