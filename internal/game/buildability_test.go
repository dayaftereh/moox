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
func TestBuildingConstructionEffectsExposeNormalizedRuntimeEffects(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	if got := rules.buildingConstructionEffects("cloning_center"); len(got) != 1 || got[0].Kind != "population_growth_flat" || got[0].Value != 0.1 {
		t.Fatalf("cloning center effects=%+v", got)
	}
	if got := rules.buildingConstructionEffects("biospheres"); len(got) != 1 || got[0].Kind != "population_capacity_flat" || got[0].Value != 2 {
		t.Fatalf("biospheres effects=%+v", got)
	}
	if got := rules.buildingConstructionEffects("holo_simulator"); len(got) != 1 || got[0].Kind != "morale_percent" || got[0].Value != 20 {
		t.Fatalf("holo simulator effects=%+v", got)
	}
	if got := rules.buildingConstructionEffects("star_base"); len(got) != 1 || got[0].Kind != "command_points" || got[0].Value != 1 {
		t.Fatalf("star base effects=%+v", got)
	}
	if got := rules.buildingConstructionEffects("marine_barracks"); len(got) != 1 || got[0].Kind != "morale_barracks_relief_percent" || got[0].Value != 20 {
		t.Fatalf("marine barracks effects=%+v", got)
	}

	// The projection is an ordered list, not a single effect slot.
	rules.MoraleBuildingBonusPercent["cloning_center"] = 15
	if got := rules.buildingConstructionEffects("cloning_center"); len(got) != 2 || got[0].Kind != "population_growth_flat" || got[1].Kind != "morale_percent" || got[1].Value != 15 {
		t.Fatalf("cloning center multi-effects=%+v", got)
	}
}

func TestConstructionChoiceIncludesBuildingDescriptionAndEffects(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(403)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	empire.KnownTechnologyIDs = []int{86}
	choices, err := rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	var holo *ConstructionChoice
	for i := range choices {
		if choices[i].ProjectKind == core.ConstructionProjectBuilding && choices[i].ProjectID == "holo_simulator" {
			holo = &choices[i]
			break
		}
	}
	if holo == nil {
		t.Fatalf("Holo Simulator choice missing: %+v", choices)
	}
	if holo.ProductionCostPP != 120 || holo.MaintenanceBC != 1 || holo.OriginalDescription == "" {
		t.Fatalf("Holo Simulator metadata=%+v", *holo)
	}
	if len(holo.Effects) != 1 || holo.Effects[0].Kind != "morale_percent" || holo.Effects[0].Value != 20 {
		t.Fatalf("Holo Simulator effects=%+v", holo.Effects)
	}
}
