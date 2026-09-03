package game

import (
	"sort"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func TestPopulationCapacityStacksAdvancedCityPlanningAndBiospheresAfterRaceLayers(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	modifiers := rules.RaceModifiers["human"]
	modifiers.Aquatic = true
	modifiers.Tolerant = true
	modifiers.Subterranean = true
	rules.RaceModifiers["human"] = modifiers

	planet := core.Planet{SizeID: "medium", ClimateID: "tundra"}
	empire := core.Empire{RaceID: "human", KnownTechnologyIDs: []int{rules.AdvancedCityPlanningTechnologyID}}
	colony := core.Colony{Buildings: []string{rules.BiospheresBuildingID}}

	base, err := rules.PopulationCapacity(planet, empire.RaceID)
	if err != nil {
		t.Fatal(err)
	}
	if base != 21 { // Aquatic Tundra -> Terran effective, Tolerant -> 100%, medium 15 + Subterranean 6.
		t.Fatalf("base capacity=%v want=22", base)
	}
	planetCap, err := rules.PopulationCapacityForEmpire(planet, empire)
	if err != nil {
		t.Fatal(err)
	}
	if planetCap != 26 {
		t.Fatalf("ACP capacity=%v want=26", planetCap)
	}
	colonyCap, err := rules.ColonyPopulationCapacity(colony, planet, empire)
	if err != nil {
		t.Fatal(err)
	}
	if colonyCap != 28 {
		t.Fatalf("Biospheres capacity=%v want=28", colonyCap)
	}
}

func TestPlanetaryTransformationChoiceQueuesAndCompletesWithoutPersistentBuilding(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0xA210)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	planet := planetByID(state, colony.PlanetID)
	planet.ClimateID = "desert"
	planet.Orbit = 2
	empire.KnownTechnologyIDs = []int{rules.PlanetaryTransformations["terraforming"].TechnologyID}
	sort.Ints(empire.KnownTechnologyIDs)

	buildingChoices, err := rules.AvailableBuildingChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, choice := range buildingChoices {
		if choice.BuildingID == "terraforming" || choice.BuildingID == "gaia_transformation" {
			t.Fatalf("planetary transformation leaked into persistent building choices: %+v", choice)
		}
	}
	buildingCommand, err := NewQueueBuildingCommand(6, QueueBuildingPayload{ColonyID: colony.ID, BuildingID: "terraforming"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.queueBuilding(state, empire.ID, protocol.SeatID(1), buildingCommand); err == nil {
		t.Fatal("Terraforming unexpectedly accepted through persistent building command")
	}

	choices, err := rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, choice := range choices {
		if choice.ProjectID == "terraforming" {
			found = choice.ProjectKind == core.ConstructionProjectPlanetaryTransformation && choice.ProductionCostPP == 250
		}
	}
	if !found {
		t.Fatalf("Terraforming missing from legal construction choices: %+v", choices)
	}

	command, err := NewQueuePlanetaryTransformationCommand(7, QueuePlanetaryTransformationPayload{ColonyID: colony.ID, ProjectID: "terraforming"})
	if err != nil {
		t.Fatal(err)
	}
	queued, err := resolver.queuePlanetaryTransformation(state, empire.ID, protocol.SeatID(1), command)
	if err != nil {
		t.Fatal(err)
	}
	if queued.Kind != "colony.planetary_transformation_queued" || colony.Construction == nil || colony.Construction.ProjectKind != core.ConstructionProjectPlanetaryTransformation {
		t.Fatalf("queue result event=%+v construction=%+v", queued, colony.Construction)
	}
	definition := rules.PlanetaryTransformations["terraforming"]
	colony.Construction.ProgressPP = definition.ProductionCostPP - 1
	colony.PopulationDynamics.ProductionAvailable = 1
	events, err := resolver.advanceConstruction(state)
	if err != nil {
		t.Fatal(err)
	}
	if colony.Construction != nil {
		t.Fatalf("completed Terraforming still has construction state: %+v", colony.Construction)
	}
	if planet.ClimateID != "arid" {
		t.Fatalf("Terraforming climate=%q want arid", planet.ClimateID)
	}
	for _, buildingID := range colony.Buildings {
		if buildingID == "terraforming" {
			t.Fatalf("Terraforming incorrectly persisted as building: %v", colony.Buildings)
		}
	}
	if len(events) != 2 || events[0].Kind != "colony.construction_progressed" || events[1].Kind != "colony.planetary_transformation_completed" {
		t.Fatalf("Terraforming completion events=%+v", events)
	}
}

func TestPlanetaryTransformationLegalityUsesPhysicalClimate(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(0xA216)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	planet := planetByID(state, colony.PlanetID)
	planet.ClimateID = "terran"
	empire.KnownTechnologyIDs = []int{rules.PlanetaryTransformations["gaia_transformation"].TechnologyID, rules.PlanetaryTransformations["terraforming"].TechnologyID}
	sort.Ints(empire.KnownTechnologyIDs)
	choices, err := rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	gaiaFound, terraformFound := false, false
	for _, choice := range choices {
		switch choice.ProjectID {
		case "gaia_transformation":
			gaiaFound = choice.ProjectKind == core.ConstructionProjectPlanetaryTransformation
		case "terraforming":
			terraformFound = true
		}
	}
	if !gaiaFound || terraformFound {
		t.Fatalf("Terran transformation choices=%+v want Gaia only", choices)
	}
}

func TestPlanetaryTransformationClimateMappingsAndBarrenOrbitRNG(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	terraforming := rules.PlanetaryTransformations["terraforming"]
	gaia := rules.PlanetaryTransformations["gaia_transformation"]

	for _, tc := range []struct {
		definition PlanetaryTransformationDefinition
		climate    string
		want       string
	}{
		{terraforming, "desert", "arid"},
		{terraforming, "tundra", "swamp"},
		{terraforming, "ocean", "terran"},
		{terraforming, "swamp", "terran"},
		{terraforming, "arid", "terran"},
		{gaia, "terran", "gaia"},
	} {
		state := core.NewSmallFixture(0xA211)
		planet := *planetByID(state, state.Colonies[0].PlanetID)
		planet.ClimateID = tc.climate
		got, err := resolver.planetaryTransformationTargetClimate(state, planet, tc.definition)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("%s on %s -> %s want %s", tc.definition.ProjectID, tc.climate, got, tc.want)
		}
	}

	for _, tc := range []struct {
		orbit int
		want  string
	}{{1, "desert"}, {2, "desert"}, {4, "tundra"}, {5, "tundra"}} {
		state := core.NewSmallFixture(0xA212)
		before := state.RNGState
		planet := *planetByID(state, state.Colonies[0].PlanetID)
		planet.ClimateID = "barren"
		planet.Orbit = tc.orbit
		got, err := resolver.planetaryTransformationTargetClimate(state, planet, terraforming)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want || state.RNGState != before {
			t.Fatalf("barren orbit %d -> %s rng=%d want %s unchanged-rng=%d", tc.orbit, got, state.RNGState, tc.want, before)
		}
	}

	stateA := core.NewSmallFixture(0xA213)
	stateB := core.NewSmallFixture(0xA213)
	planetA := *planetByID(stateA, stateA.Colonies[0].PlanetID)
	planetB := *planetByID(stateB, stateB.Colonies[0].PlanetID)
	planetA.ClimateID, planetB.ClimateID = "barren", "barren"
	planetA.Orbit, planetB.Orbit = 3, 3
	before := stateA.RNGState
	gotA, err := resolver.planetaryTransformationTargetClimate(stateA, planetA, terraforming)
	if err != nil {
		t.Fatal(err)
	}
	gotB, err := resolver.planetaryTransformationTargetClimate(stateB, planetB, terraforming)
	if err != nil {
		t.Fatal(err)
	}
	if gotA != gotB || stateA.RNGState != stateB.RNGState || stateA.RNGState == before {
		t.Fatalf("middle barren orbit is not deterministic: A=%s/%d B=%s/%d before=%d", gotA, stateA.RNGState, gotB, stateB.RNGState, before)
	}
	if gotA != "desert" && gotA != "tundra" {
		t.Fatalf("middle barren orbit result=%q", gotA)
	}
}

func TestAdvancedCityPlanningBreakthroughDoesNotRetroactivelyGrowPopulation(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0xA214)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	planet := planetByID(state, colony.PlanetID)
	planet.ClimateID = "terran"
	colony.Population = core.NewAssimilatedPopulation(colony.EmpireID, 6, 5, 1)
	if err := resolver.recalculateColony(state, colony); err != nil {
		t.Fatal(err)
	}
	oldCapacity := colony.PopulationDynamics.Capacity
	if oldCapacity != 12 || colony.PopulationDynamics.ProjectedGrowth != 0 {
		t.Fatalf("pre-research dynamics=%+v want capacity=12/no growth", colony.PopulationDynamics)
	}
	fieldID := rules.TechnologyFieldByID[rules.AdvancedCityPlanningTechnologyID]
	cost := rules.TechnologyFieldCostsRP[fieldID]
	currentRP := colony.AdjustedEconomy.Research
	empire.Research = &core.ResearchState{
		TechFieldID: fieldID, SelectionMode: core.ResearchSelectionChooseOne,
		TechnologyIDs: []int{rules.AdvancedCityPlanningTechnologyID}, ProgressRP: 2*cost - currentRP,
	}
	if _, err := resolver.Resolve(ResolveContext{}, state, nil); err != nil {
		t.Fatal(err)
	}
	if !empireKnowsTechnology(empire, rules.AdvancedCityPlanningTechnologyID) {
		t.Fatalf("Advanced City Planning was not acquired: %v", empire.KnownTechnologyIDs)
	}
	if colony.Population.Total() != 12 {
		t.Fatalf("ACP breakthrough retroactively grew Population: %v", colony.Population.Total())
	}
	if colony.PopulationDynamics.Capacity != 17 || colony.PopulationDynamics.ProjectedGrowth <= 0 {
		t.Fatalf("post-research dynamics=%+v want capacity=17 and next-state growth", colony.PopulationDynamics)
	}
}

func TestBiospheresAndGaiaCompletionAffectNextStateCapacityOnly(t *testing.T) {
	for _, tc := range []struct {
		name        string
		projectKind core.ConstructionProjectKind
		projectID   string
		climate     string
		wantClimate string
		wantCap     float64
	}{
		{"Biospheres", core.ConstructionProjectBuilding, "biospheres", "terran", "terran", 14},
		{"Gaia", core.ConstructionProjectPlanetaryTransformation, "gaia_transformation", "terran", "gaia", 15},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rules := loadCommittedEconomyRules(t)
			resolver, err := NewEconomyResolver(rules)
			if err != nil {
				t.Fatal(err)
			}
			state := core.NewSmallFixture(0xA215)
			colony := &state.Colonies[0]
			planet := planetByID(state, colony.PlanetID)
			planet.ClimateID = tc.climate
			colony.Population = core.NewAssimilatedPopulation(colony.EmpireID, 6, 6, 0)
			if err := resolver.recalculateColony(state, colony); err != nil {
				t.Fatal(err)
			}
			if colony.PopulationDynamics.Capacity != 12 || colony.PopulationDynamics.ProjectedGrowth != 0 {
				t.Fatalf("pre-completion dynamics=%+v", colony.PopulationDynamics)
			}
			cost := 0.0
			if tc.projectKind == core.ConstructionProjectBuilding {
				cost = rules.BuildingDefinitions[tc.projectID].ProductionCostPP
			} else {
				cost = rules.PlanetaryTransformations[tc.projectID].ProductionCostPP
			}
			available := colony.PopulationDynamics.ProductionAvailable
			if available <= 0 || available >= cost {
				t.Fatalf("unexpected production available=%v cost=%v", available, cost)
			}
			colony.Construction = &core.ConstructionState{ProjectKind: tc.projectKind, ProjectID: tc.projectID, ProgressPP: cost - available}
			if _, err := resolver.Resolve(ResolveContext{}, state, nil); err != nil {
				t.Fatal(err)
			}
			if colony.Population.Total() != 12 {
				t.Fatalf("completion retroactively grew Population: %v", colony.Population.Total())
			}
			if planet.ClimateID != tc.wantClimate {
				t.Fatalf("post-completion climate=%q want=%q", planet.ClimateID, tc.wantClimate)
			}
			if colony.PopulationDynamics.Capacity != tc.wantCap || colony.PopulationDynamics.ProjectedGrowth <= 0 {
				t.Fatalf("post-completion dynamics=%+v want capacity=%v and next-state growth", colony.PopulationDynamics, tc.wantCap)
			}
		})
	}
}

func TestAggregateCapacityClampPreservesRoleProportions(t *testing.T) {
	colony := core.Colony{Population: core.NewAssimilatedPopulation(1, 4, 3, 3)}
	removed := clampAggregatePopulationToCapacity(&colony, 5)
	if removed != 5 || colony.Population.Total() != 5 || colony.Population.Farmers() != 2 || colony.Population.Workers() != 1.5 || colony.Population.Scientists() != 1.5 {
		t.Fatalf("clamped aggregate population=%+v removed=%v", colony.Population, removed)
	}
}
