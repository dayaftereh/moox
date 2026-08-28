package game

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"moox/internal/core"
)

func TestGrantTechnologyAddsTechnologyWithoutCompletingField(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0x1234)
	empire := &state.Empires[0]
	if err := rules.InitializeEmpireTechnologies(empire, NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}

	const technologyID = 155 // research_laboratory, TechField 56.
	event, err := resolver.GrantTechnology(state, empire.ID, technologyID, TechnologyGrantOptions{SourceKind: "external_test"})
	if err != nil {
		t.Fatal(err)
	}
	if !containsInt(empire.KnownTechnologyIDs, technologyID) {
		t.Fatalf("Technology %d was not granted", technologyID)
	}
	if containsInt(empire.KnownTechnologyFieldIDs, 56) {
		t.Fatal("external Technology grant must not invent TechField completion")
	}
	if event.Kind != EventTechnologyGranted || event.SeatID != 0 || event.CommandSequence != 0 {
		t.Fatalf("unexpected event envelope: %+v", event)
	}
	var payload TechnologyGrantedEvent
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.EmpireID != empire.ID || payload.TechnologyID != technologyID || payload.TechnologyKey != "research_laboratory" || payload.TechFieldID != 56 || payload.SourceKind != "external_test" || payload.UncreativeRepair != nil {
		t.Fatalf("unexpected grant payload: %+v", payload)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("grant produced invalid state: %v", err)
	}
}

func TestGrantTechnologyRepairsUncreativeFixedChoice(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0x3344)
	empire := &state.Empires[0]
	empire.RaceID = "klackon"
	newGameRNG := core.NewRNG(0x7788)
	if err := rules.InitializeEmpireTechnologies(empire, NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp, NewGameRNG: newGameRNG}); err != nil {
		t.Fatal(err)
	}

	var fixed core.FixedResearchChoice
	found := false
	for _, choice := range empire.UncreativeResearchChoices {
		if len(rules.TechnologyIDsByField[choice.TechFieldID]) > 1 && !containsInt(empire.KnownTechnologyFieldIDs, choice.TechFieldID) {
			fixed = choice
			found = true
			break
		}
	}
	if !found {
		t.Fatal("fixture has no incomplete multi-application Uncreative field")
	}
	beforeRNG := state.RNGState
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.GrantTechnology(state, empire.ID, fixed.TechnologyID, TechnologyGrantOptions{SourceKind: "external_test"})
	if err != nil {
		t.Fatal(err)
	}
	if !containsInt(empire.KnownTechnologyIDs, fixed.TechnologyID) {
		t.Fatalf("acquired fixed Technology %d is not known", fixed.TechnologyID)
	}
	if containsInt(empire.KnownTechnologyFieldIDs, fixed.TechFieldID) {
		t.Fatal("Uncreative external acquisition must not complete the TechField")
	}
	if state.RNGState == beforeRNG {
		t.Fatal("Uncreative repair did not commit authoritative RNG consumption")
	}
	var payload TechnologyGrantedEvent
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.UncreativeRepair == nil || !payload.UncreativeRepair.Changed || payload.UncreativeRepair.PreviousTechnologyID != fixed.TechnologyID {
		t.Fatalf("missing Uncreative repair metadata: %+v", payload)
	}
	if replacement, ok := fixedResearchTechnology(empire.UncreativeResearchChoices, fixed.TechFieldID); ok && replacement == fixed.TechnologyID {
		t.Fatalf("Uncreative fixed choice was not repaired away from acquired Technology %d", fixed.TechnologyID)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("Uncreative grant produced invalid state: %v", err)
	}
}

func TestGrantTechnologyRejectsDuplicateWithoutMutation(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0x1234)
	empire := &state.Empires[0]
	if err := rules.InitializeEmpireTechnologies(empire, NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	technologyID := empire.KnownTechnologyIDs[0]
	beforeRNG := state.RNGState
	beforeCount := len(empire.KnownTechnologyIDs)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.GrantTechnology(state, empire.ID, technologyID, TechnologyGrantOptions{SourceKind: "external_test"}); err == nil {
		t.Fatal("expected duplicate Technology grant to fail")
	}
	if len(empire.KnownTechnologyIDs) != beforeCount || state.RNGState != beforeRNG {
		t.Fatal("failed duplicate grant mutated authoritative state")
	}
}
func TestGrantTechnologyRejectsActiveResearchConflict(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0x5566)
	empire := &state.Empires[0]
	if err := rules.InitializeEmpireTechnologies(empire, NewGameTechnologyOptions{Level: NewGameTechnologyPreWarp}); err != nil {
		t.Fatal(err)
	}
	empire.Research = &core.ResearchState{
		TechFieldID:   4,
		SelectionMode: core.ResearchSelectionChooseOne,
		TechnologyIDs: []int{56},
	}
	beforeRNG := state.RNGState
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.GrantTechnology(state, empire.ID, 56, TechnologyGrantOptions{SourceKind: "external_test"}); err == nil {
		t.Fatal("expected active Research conflict to be rejected")
	}
	if containsInt(empire.KnownTechnologyIDs, 56) || state.RNGState != beforeRNG {
		t.Fatal("failed active-Research grant mutated authoritative state")
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("rejected grant left invalid state: %v", err)
	}
}
