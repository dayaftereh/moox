package game

import (
	"math"
	"testing"

	"moox/internal/core"
)

func queueItemFromChoice(choice ConstructionChoice) ConstructionQueueItem {
	return ConstructionQueueItem{
		ProjectKind:        choice.ProjectKind,
		ProjectID:          choice.ProjectID,
		ShipDesignID:       choice.ShipDesignID,
		ShipDesignRevision: choice.ShipDesignRevision,
	}
}

func repeatableConstructionChoice(t *testing.T, choices []ConstructionChoice) ConstructionChoice {
	t.Helper()
	for _, choice := range choices {
		if choice.ProjectKind != core.ConstructionProjectBuilding && choice.ProjectKind != core.ConstructionProjectPlanetaryTransformation {
			return choice
		}
	}
	t.Fatalf("no repeatable construction choice in %+v", choices)
	return ConstructionChoice{}
}

func TestSetConstructionQueueAcceptsMoreThanOriginalSevenSlots(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(211)
	colony := &state.Colonies[0]
	choices, err := rules.AvailableConstructionQueueChoices(state, colony.EmpireID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	choice := repeatableConstructionChoice(t, choices)
	items := make([]ConstructionQueueItem, 8)
	for i := range items {
		items[i] = queueItemFromChoice(choice)
	}
	command, err := NewSetConstructionQueueCommand(1, SetConstructionQueuePayload{ColonyID: colony.ID, Items: items})
	if err != nil {
		t.Fatal(err)
	}
	ctx, batches := constructionBatch(t, state, command)
	result, err := resolver.Resolve(ctx, state, batches)
	if err != nil {
		t.Fatal(err)
	}
	got := result.State.Colonies[0]
	if got.Construction == nil || got.Construction.ProjectKind != choice.ProjectKind || got.Construction.ProjectID != choice.ProjectID {
		t.Fatalf("queue head=%+v want %s/%q", got.Construction, choice.ProjectKind, choice.ProjectID)
	}
	if len(got.ConstructionQueue) != 7 {
		t.Fatalf("upcoming queue len=%d want=7 for eight total items", len(got.ConstructionQueue))
	}
}

func TestSetConstructionQueuePreservesAccumulatedProductionStock(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(212)
	colony := &state.Colonies[0]
	choices, err := rules.AvailableConstructionQueueChoices(state, colony.EmpireID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(choices) < 2 {
		t.Fatalf("need at least two legal construction choices, got %+v", choices)
	}
	old := constructionStateFromQueueItem(queueItemFromChoice(choices[1]))
	old.ProgressPP = 12
	colony.Construction = &old
	colony.ConstructionReservePP = 3

	command, err := NewSetConstructionQueueCommand(1, SetConstructionQueuePayload{
		ColonyID: colony.ID,
		Items: []ConstructionQueueItem{
			queueItemFromChoice(choices[0]),
			queueItemFromChoice(choices[1]),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.setConstructionQueue(state, colony.EmpireID, 1, command); err != nil {
		t.Fatal(err)
	}
	if colony.Construction == nil || colony.Construction.ProjectKind != choices[0].ProjectKind || colony.Construction.ProjectID != choices[0].ProjectID {
		t.Fatalf("reordered head=%+v want first legal choice %+v", colony.Construction, choices[0])
	}
	if math.Abs(colony.Construction.ProgressPP-15) > 1e-9 {
		t.Fatalf("reordered head progress=%v want accumulated 15 PP", colony.Construction.ProgressPP)
	}
	if colony.ConstructionReservePP != 0 || len(colony.ConstructionQueue) != 1 {
		t.Fatalf("reserve=%v queue=%+v want reserve=0 and one upcoming item", colony.ConstructionReservePP, colony.ConstructionQueue)
	}
}

func TestConstructionQueueCarriesOverflowAcrossCompletedItems(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(213)
	addColonyShipTestTechnology(&state.Empires[0], 120, standardFuelCellsTechnologyID)
	colony := &state.Colonies[0]
	choices, err := rules.AvailableConstructionQueueChoices(state, colony.EmpireID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	var choice ConstructionChoice
	found := false
	for _, candidate := range choices {
		if candidate.ProductionCostPP > 0 && candidate.ProjectKind != core.ConstructionProjectBuilding && candidate.ProjectKind != core.ConstructionProjectPlanetaryTransformation && candidate.ProjectKind != core.ConstructionProjectHousing {
			choice = candidate
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("need one repeatable finite construction choice, got %+v", choices)
	}
	head := constructionStateFromQueueItem(queueItemFromChoice(choice))
	next := constructionStateFromQueueItem(queueItemFromChoice(choice))
	colony.Construction = &head
	colony.ConstructionQueue = []core.ConstructionState{next}
	colony.ConstructionReservePP = 0
	colony.PopulationDynamics.ProductionAvailable = choice.ProductionCostPP*2 + 5

	events, err := resolver.advanceConstruction(state)
	if err != nil {
		t.Fatal(err)
	}
	if colony.Construction != nil || len(colony.ConstructionQueue) != 0 {
		t.Fatalf("construction=%+v queue=%+v want both completed", colony.Construction, colony.ConstructionQueue)
	}
	if math.Abs(colony.ConstructionReservePP-5) > 1e-9 {
		t.Fatalf("overflow reserve=%v want=5 PP", colony.ConstructionReservePP)
	}
	completed := 0
	for _, event := range events {
		if event.Kind != "colony.construction_progressed" {
			completed++
		}
	}
	if completed != 2 {
		t.Fatalf("completion events=%d want=2 events=%+v", completed, events)
	}
}
