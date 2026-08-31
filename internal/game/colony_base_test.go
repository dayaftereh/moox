package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
)

func addSameSystemColonyBaseTarget(state *core.GameState) core.ID {
	planet := core.Planet{
		ID:        state.NewID(),
		Name:      "Alpha II",
		Orbit:     2,
		SizeID:    "medium",
		MineralID: "abundant",
		GravityID: "normal_g",
		ClimateID: "terran",
	}
	state.Galaxy.Systems[0].Planets = append(state.Galaxy.Systems[0].Planets, planet)
	return planet.ID
}

func findBuildingChoice(choices []BuildingChoice, buildingID string) *BuildingChoice {
	for i := range choices {
		if choices[i].BuildingID == buildingID {
			return &choices[i]
		}
	}
	return nil
}

func TestColonyBaseBuildabilityRequiresEmptySameSystemPlanet(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(7001)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	empire.KnownTechnologyIDs = []int{ColonyBaseTechnologyID}

	choices, err := rules.AvailableBuildingChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	if choice := findBuildingChoice(choices, ColonyBaseBuildingID); choice != nil {
		t.Fatalf("Colony Base unexpectedly buildable without same-system target: %+v", *choice)
	}

	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	queue, err := NewQueueBuildingCommand(1, QueueBuildingPayload{ColonyID: colony.ID, BuildingID: ColonyBaseBuildingID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.queueBuilding(state, empire.ID, 1, queue); err == nil {
		t.Fatal("direct Colony Base queue unexpectedly bypassed same-system target legality")
	}

	targetID := addSameSystemColonyBaseTarget(state)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	choices, err = rules.AvailableBuildingChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	choice := findBuildingChoice(choices, ColonyBaseBuildingID)
	if choice == nil {
		t.Fatal("Colony Base missing after adding empty same-system target")
	}
	if choice.ProductionID != ColonyBaseProductionID || choice.TechnologyID != ColonyBaseTechnologyID || choice.ProductionCostPP != ColonyBaseBaseCostPP || choice.MaintenanceBC != 0 {
		t.Fatalf("Colony Base choice=%+v", *choice)
	}
	if targetID == 0 {
		t.Fatal("same-system target has zero ID")
	}
	if _, err := resolver.queueBuilding(state, empire.ID, 1, queue); err != nil {
		t.Fatalf("legal Colony Base queue rejected: %v", err)
	}
	if colony.Construction == nil || colony.Construction.ProjectID != ColonyBaseBuildingID || colony.Construction.ProjectKind != core.ConstructionProjectBuilding {
		t.Fatalf("queued Colony Base construction=%+v", colony.Construction)
	}
}

func TestPendingColonyBaseResolutionColonizesWithoutSourcePopulationTransfer(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(7002)
	empire := &state.Empires[0]
	source := &state.Colonies[0]
	targetID := addSameSystemColonyBaseTarget(state)
	source.Buildings = append(source.Buildings, ColonyBaseBuildingID)
	sourcePopulation := source.Population

	pending, err := PendingColonyBaseResolutions(state, empire.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || pending[0].SourceColonyID != source.ID || len(pending[0].TargetPlanetIDs) != 1 || pending[0].TargetPlanetIDs[0] != targetID || pending[0].TrashRefundBC != ColonyBaseTrashRefundBC {
		t.Fatalf("pending Colony Base resolution=%+v", pending)
	}

	invalid, err := NewColonizeWithBaseCommand(1, ColonizeWithBasePayload{SourceColonyID: source.ID, PlanetID: state.Galaxy.Systems[1].Planets[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.ResolveColonyBaseCommand(state, empire.ID, 1, invalid); err == nil {
		t.Fatal("cross-system Colony Base target unexpectedly accepted")
	}
	if !colonyOwnsBuilding(source, ColonyBaseBuildingID) || len(state.Colonies) != 1 {
		t.Fatal("rejected Colony Base command mutated authoritative state")
	}

	command, err := NewColonizeWithBaseCommand(2, ColonizeWithBasePayload{SourceColonyID: source.ID, PlanetID: targetID})
	if err != nil {
		t.Fatal(err)
	}
	events, err := resolver.ResolveColonyBaseCommand(state, empire.ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Kind != "empire.planet_colonized" || events[1].Kind != "colony.colony_base_colonized" {
		t.Fatalf("Colony Base colonization events=%+v", events)
	}
	if len(state.Colonies) != 2 {
		t.Fatalf("colony count=%d want=2", len(state.Colonies))
	}
	if colonyOwnsBuilding(&state.Colonies[0], ColonyBaseBuildingID) {
		t.Fatal("source Colony retained consumed Colony Base")
	}
	if !reflect.DeepEqual(state.Colonies[0].Population, sourcePopulation) {
		t.Fatalf("source Population changed: got=%+v want=%+v", state.Colonies[0].Population, sourcePopulation)
	}
	newColony := state.Colonies[1]
	if newColony.PlanetID != targetID || newColony.EmpireID != empire.ID || !reflect.DeepEqual(newColony.Population, core.NewAssimilatedPopulation(empire.ID, 1, 0, 0)) {
		t.Fatalf("new Colony=%+v", newColony)
	}
	target := planetByID(state, targetID)
	if target == nil || target.ColonyID != newColony.ID {
		t.Fatalf("target Planet after colonization=%+v", target)
	}
	pending, err = PendingColonyBaseResolutions(state, empire.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 0 {
		t.Fatalf("consumed Colony Base still pending: %+v", pending)
	}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestTrashColonyBaseRefundsOneHundredBCEvenWithoutTarget(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(7003)
	empire := &state.Empires[0]
	source := &state.Colonies[0]
	source.Buildings = append(source.Buildings, ColonyBaseBuildingID)
	empire.Treasury.BalanceBC = 50

	pending, err := PendingColonyBaseResolutions(state, empire.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || len(pending[0].TargetPlanetIDs) != 0 || pending[0].TrashRefundBC != 100 {
		t.Fatalf("targetless pending Colony Base=%+v", pending)
	}

	command, err := NewTrashColonyBaseCommand(1, TrashColonyBasePayload{SourceColonyID: source.ID})
	if err != nil {
		t.Fatal(err)
	}
	events, err := resolver.ResolveColonyBaseCommand(state, empire.ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != "colony.colony_base_trashed" {
		t.Fatalf("trash events=%+v", events)
	}
	if colonyOwnsBuilding(source, ColonyBaseBuildingID) {
		t.Fatal("trashed Colony Base still owned")
	}
	if empire.Treasury.BalanceBC != 150 {
		t.Fatalf("Treasury=%v want=150", empire.Treasury.BalanceBC)
	}
	if len(state.Colonies) != 1 {
		t.Fatalf("trash created/removed Colony unexpectedly: %d", len(state.Colonies))
	}
	var payload ColonyBaseTrashedEvent
	if err := json.Unmarshal(events[0].Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.RefundBC != 100 || payload.PreviousBalanceBC != 50 || payload.CurrentBalanceBC != 150 || payload.SourceColonyID != source.ID {
		t.Fatalf("trash payload=%+v", payload)
	}
	if _, err := resolver.ResolveColonyBaseCommand(state, empire.ID, 1, command); err == nil {
		t.Fatal("second trash unexpectedly accepted after Colony Base was consumed")
	}
}

func TestPendingColonyBaseResolutionSurvivesSchema19RoundTrip(t *testing.T) {
	state := core.NewSmallFixture(7004)
	targetID := addSameSystemColonyBaseTarget(state)
	state.Colonies[0].Buildings = append(state.Colonies[0].Buildings, ColonyBaseBuildingID)
	if state.SchemaVersion != core.StateSchemaVersion || core.StateSchemaVersion != 19 {
		t.Fatalf("schema=%d constant=%d want=19", state.SchemaVersion, core.StateSchemaVersion)
	}
	before, err := PendingColonyBaseResolutions(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := core.MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := core.UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	after, err := PendingColonyBaseResolutions(loaded, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("pending Colony Base changed across schema-18 round trip: before=%+v after=%+v", before, after)
	}
	if len(after) != 1 || len(after[0].TargetPlanetIDs) != 1 || after[0].TargetPlanetIDs[0] != targetID {
		t.Fatalf("round-tripped pending Colony Base=%+v", after)
	}
	reencoded, err := core.MarshalState(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(reencoded, encoded) {
		t.Fatalf("schema-18 bytes changed after Colony Base round trip\nfirst=%s\nsecond=%s", encoded, reencoded)
	}
}

func TestColonyBaseColonizationReplayIsDeterministic(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	original := core.NewSmallFixture(7005)
	targetID := addSameSystemColonyBaseTarget(original)
	original.Colonies[0].Buildings = append(original.Colonies[0].Buildings, ColonyBaseBuildingID)
	encoded, err := core.MarshalState(original)
	if err != nil {
		t.Fatal(err)
	}
	stateA, err := core.UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	stateB, err := core.UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	resolverA, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	resolverB, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	command, err := NewColonizeWithBaseCommand(3, ColonizeWithBasePayload{SourceColonyID: original.Colonies[0].ID, PlanetID: targetID})
	if err != nil {
		t.Fatal(err)
	}
	eventsA, err := resolverA.ResolveColonyBaseCommand(stateA, original.Empires[0].ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	eventsB, err := resolverB.ResolveColonyBaseCommand(stateB, original.Empires[0].ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(eventsA, eventsB) {
		t.Fatalf("same Colony Base replay produced different events: A=%+v B=%+v", eventsA, eventsB)
	}
	stateBytesA, err := core.MarshalState(stateA)
	if err != nil {
		t.Fatal(err)
	}
	stateBytesB, err := core.MarshalState(stateB)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stateBytesA, stateBytesB) {
		t.Fatalf("same Colony Base replay produced different state\nA=%s\nB=%s", stateBytesA, stateBytesB)
	}
}
