package game

import (
	"reflect"
	"testing"

	"moox/internal/core"
)

func TestAvailableBuildingChoicesRequireKnownTechnologyAndOwnership(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(401)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]

	choices, err := rules.AvailableBuildingChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(choices) != 0 {
		t.Fatalf("empire with no known technology got choices: %+v", choices)
	}

	empire.KnownTechnologyIDs = []int{86, 141}
	choices, err = rules.AvailableBuildingChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, choice := range choices {
		ids = append(ids, choice.BuildingID)
	}
	if !reflect.DeepEqual(ids, []string{"holo_simulator", "pleasure_dome"}) {
		t.Fatalf("known technology choices=%v", ids)
	}
	if choices[0].TechnologyID != 86 || choices[0].ProductionCostPP != 120 || choices[0].MaintenanceBC != 1 {
		t.Fatalf("unexpected Holo choice: %+v", choices[0])
	}

	colony.Buildings = []string{"holo_simulator"}
	choices, err = rules.AvailableBuildingChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(choices) != 1 || choices[0].BuildingID != "pleasure_dome" {
		t.Fatalf("owned building not filtered: %+v", choices)
	}

	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectBuilding, ProjectID: "pleasure_dome", ProgressPP: 1}
	choices, err = rules.AvailableBuildingChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(choices) != 0 {
		t.Fatalf("busy colony should expose no queue choices: %+v", choices)
	}
}

func TestAvailableBuildingChoicesRejectForeignColony(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(402)
	state.Empires = append(state.Empires, core.Empire{ID: state.NewID(), Name: "Other", RaceID: "alkari"})
	if _, err := rules.AvailableBuildingChoices(state, state.Empires[1].ID, state.Colonies[0].ID); err == nil {
		t.Fatal("expected foreign-colony choice projection to fail")
	}
}
