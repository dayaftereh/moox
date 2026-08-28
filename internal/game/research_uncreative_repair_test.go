package game

import (
	"path/filepath"
	"sort"
	"testing"

	"moox/internal/core"
)

func uncreativeRepairFixture(t *testing.T) (*EconomyRules, core.Empire) {
	t.Helper()
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	empire := core.Empire{ID: 1, RaceID: "klackon"}
	newGameRNG := core.NewRNG(0x12345678)
	if err := rules.InitializeEmpireTechnologies(&empire, NewGameTechnologyOptions{
		Level:      NewGameTechnologyPreWarp,
		NewGameRNG: newGameRNG,
	}); err != nil {
		t.Fatal(err)
	}
	return rules, empire
}

func addKnownTechnologyForRepairTest(empire *core.Empire, technologyID int) {
	if containsInt(empire.KnownTechnologyIDs, technologyID) {
		return
	}
	empire.KnownTechnologyIDs = append(empire.KnownTechnologyIDs, technologyID)
	sort.Ints(empire.KnownTechnologyIDs)
}

func setFixedResearchTechnologyForRepairTest(t *testing.T, empire *core.Empire, fieldID, technologyID int) {
	t.Helper()
	index := sort.Search(len(empire.UncreativeResearchChoices), func(i int) bool {
		return empire.UncreativeResearchChoices[i].TechFieldID >= fieldID
	})
	if index >= len(empire.UncreativeResearchChoices) || empire.UncreativeResearchChoices[index].TechFieldID != fieldID {
		t.Fatalf("missing fixed Uncreative choice for TechField %d", fieldID)
	}
	empire.UncreativeResearchChoices[index].TechnologyID = technologyID
}

func TestUncreativeRepairReservoirMatchesOriginalSlotOrderAndDrawCount(t *testing.T) {
	rules, empire := uncreativeRepairFixture(t)
	const fieldID = 4
	fixedID, ok := fixedResearchTechnology(empire.UncreativeResearchChoices, fieldID)
	if !ok {
		t.Fatalf("missing fixed Uncreative choice for TechField %d", fieldID)
	}
	addKnownTechnologyForRepairTest(&empire, fixedID)

	const seed = 0xBEEF
	rng := core.NewRNG(seed)
	expectedRNG := core.NewRNG(seed)
	expectedID := 0
	candidateCount := 0
	for _, technologyID := range rules.TechnologyIDsByField[fieldID] {
		if technologyID == fixedID {
			continue
		}
		candidateCount++
		roll, err := expectedRNG.Intn(candidateCount)
		if err != nil {
			t.Fatal(err)
		}
		if roll == 0 {
			expectedID = technologyID
		}
	}

	result, err := rules.RepairUncreativeResearchChoiceAfterAcquisition(
		&empire,
		fixedID,
		UncreativeResearchRepairOptions{},
		rng,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || result.PreviousTechnologyID != fixedID || result.ReplacementTechnologyID != expectedID {
		t.Fatalf("repair result=%+v want previous=%d replacement=%d changed=true", result, fixedID, expectedID)
	}
	if rng.State() != expectedRNG.State() {
		t.Fatalf("repair RNG state=%d want=%d after %d reservoir draws", rng.State(), expectedRNG.State(), candidateCount)
	}
	if got, ok := fixedResearchTechnology(empire.UncreativeResearchChoices, fieldID); !ok || got != expectedID {
		t.Fatalf("fixed TechField %d choice=%d found=%v want=%d", fieldID, got, ok, expectedID)
	}

	state := &core.GameState{Empires: []core.Empire{empire}}
	choices, err := rules.AvailableResearchChoices(state, empire.ID)
	if err != nil {
		t.Fatal(err)
	}
	choice, ok := researchChoiceByField(choices, fieldID)
	if !ok || choice.SelectionMode != core.ResearchSelectionFixedOne || len(choice.TechnologyIDs) != 1 || choice.TechnologyIDs[0] != expectedID {
		t.Fatalf("repaired legal choice=%+v found=%v want fixed Technology %d", choice, ok, expectedID)
	}
}

func TestUncreativeRepairConsumesReservoirDrawWithoutReplacingLiveFixedChoice(t *testing.T) {
	rules, empire := uncreativeRepairFixture(t)
	const fieldID = 4
	fixedID, ok := fixedResearchTechnology(empire.UncreativeResearchChoices, fieldID)
	if !ok {
		t.Fatalf("missing fixed Uncreative choice for TechField %d", fieldID)
	}
	externalID := 0
	remainingID := 0
	for _, technologyID := range rules.TechnologyIDsByField[fieldID] {
		if technologyID == fixedID {
			continue
		}
		if externalID == 0 {
			externalID = technologyID
		} else {
			remainingID = technologyID
		}
	}
	if externalID == 0 || remainingID == 0 {
		t.Fatalf("TechField %d needs two non-fixed applications for the fixture", fieldID)
	}
	addKnownTechnologyForRepairTest(&empire, externalID)

	const seed = 0xCAFE
	rng := core.NewRNG(seed)
	expectedRNG := core.NewRNG(seed)
	if _, err := expectedRNG.Intn(1); err != nil {
		t.Fatal(err)
	}
	result, err := rules.RepairUncreativeResearchChoiceAfterAcquisition(
		&empire,
		externalID,
		UncreativeResearchRepairOptions{},
		rng,
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Changed || result.ReplacementTechnologyID != fixedID {
		t.Fatalf("repair with live fixed choice=%+v want unchanged Technology %d", result, fixedID)
	}
	if rng.State() != expectedRNG.State() {
		t.Fatalf("repair RNG state=%d want=%d after discarded one-candidate reservoir", rng.State(), expectedRNG.State())
	}
	if got, _ := fixedResearchTechnology(empire.UncreativeResearchChoices, fieldID); got != fixedID {
		t.Fatalf("fixed Technology changed to %d want=%d", got, fixedID)
	}
}

func TestUncreativeRepairDoesNotRandomlyReplaceGovernmentEvolution(t *testing.T) {
	rules, empire := uncreativeRepairFixture(t)
	const fieldID = 6
	fixedID, ok := fixedResearchTechnology(empire.UncreativeResearchChoices, fieldID)
	if !ok {
		t.Fatalf("missing fixed Uncreative choice for TechField %d", fieldID)
	}
	addKnownTechnologyForRepairTest(&empire, fixedID)

	rng := core.NewRNG(123)
	before := rng.State()
	result, err := rules.RepairUncreativeResearchChoiceAfterAcquisition(
		&empire,
		fixedID,
		UncreativeResearchRepairOptions{},
		rng,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || result.ReplacementTechnologyID != 0 {
		t.Fatalf("government repair=%+v want removal with no replacement", result)
	}
	if rng.State() != before {
		t.Fatalf("government repair consumed RNG: before=%d after=%d", before, rng.State())
	}
	if got, found := fixedResearchTechnology(empire.UncreativeResearchChoices, fieldID); found {
		t.Fatalf("government TechField %d retained fixed Technology %d", fieldID, got)
	}

	// Make the field frontier-visible and verify a missing fixed application is
	// represented as no legal action rather than a state-projection error.
	empire.KnownTechnologyFieldIDs = append(empire.KnownTechnologyFieldIDs, 12)
	sort.Ints(empire.KnownTechnologyFieldIDs)
	state := &core.GameState{Empires: []core.Empire{empire}}
	choices, err := rules.AvailableResearchChoices(state, empire.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, found := researchChoiceByField(choices, fieldID); found {
		t.Fatalf("government TechField %d should be unselectable without an eligible fixed application", fieldID)
	}
}

func TestUncreativeRepairKeepsDimensionalPortalGateExplicit(t *testing.T) {
	rules, empire := uncreativeRepairFixture(t)
	const fieldID = 51
	setFixedResearchTechnologyForRepairTest(t, &empire, fieldID, 54)
	addKnownTechnologyForRepairTest(&empire, 54)

	if _, err := rules.RepairUncreativeResearchChoiceAfterAcquisition(
		&empire,
		54,
		UncreativeResearchRepairOptions{},
		core.NewRNG(1),
	); err == nil {
		t.Fatal("expected unresolved Dimensional Portal repair gate to require an explicit policy")
	}

	allow := true
	rules, empire = uncreativeRepairFixture(t)
	setFixedResearchTechnologyForRepairTest(t, &empire, fieldID, 54)
	addKnownTechnologyForRepairTest(&empire, 54)
	rng := core.NewRNG(1)
	expected := core.NewRNG(1)
	if _, err := expected.Intn(1); err != nil {
		t.Fatal(err)
	}
	result, err := rules.RepairUncreativeResearchChoiceAfterAcquisition(
		&empire,
		54,
		UncreativeResearchRepairOptions{DimensionalPortalAllowed: &allow},
		rng,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed || result.ReplacementTechnologyID != 52 {
		t.Fatalf("Dimensional Portal repair=%+v want Technology 52", result)
	}
	if rng.State() != expected.State() {
		t.Fatalf("Dimensional Portal one-candidate reservoir RNG=%d want=%d", rng.State(), expected.State())
	}
}
