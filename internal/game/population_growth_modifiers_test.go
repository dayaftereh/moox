package game

import (
	"math"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func TestPopulationGrowthMedicineModifiersAddPercentagePoints(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	planet := core.Planet{SizeID: "medium", ClimateID: "terran"}
	colony := core.Colony{Population: core.NewAssimilatedPopulation(1, 2, 1, 1)}
	adjusted := core.ColonyEconomy{Food: 4, Production: 3}
	base := math.Sqrt(0.002 * 4 * (12 - 4) / 12)

	micro, err := rules.CalculatePopulationDynamicsForEmpire(colony, planet, core.Empire{RaceID: "human", KnownTechnologyIDs: []int{107}}, adjusted)
	if err != nil {
		t.Fatal(err)
	}
	if !closePopulationValue(micro.GrowthMultiplier, 1.25) || !closePopulationValue(micro.ProjectedGrowth, base*1.25) {
		t.Fatalf("Microbiotics growth=%+v want multiplier=1.25 projected=%v", micro, base*1.25)
	}

	both, err := rules.CalculatePopulationDynamicsForEmpire(colony, planet, core.Empire{RaceID: "human", KnownTechnologyIDs: []int{107, 193}}, adjusted)
	if err != nil {
		t.Fatal(err)
	}
	if !closePopulationValue(both.GrowthMultiplier, 1.75) || !closePopulationValue(both.ProjectedGrowth, base*1.75) {
		t.Fatalf("medicine stacking=%+v want multiplier=1.75 projected=%v", both, base*1.75)
	}
}

func TestPopulationGrowthHousingUsesAvailableProductionAndOriginalFloor(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	planet := core.Planet{SizeID: "medium", ClimateID: "terran"}
	colony := core.Colony{
		Population:   core.NewAssimilatedPopulation(1, 2, 1, 1),
		Construction: &core.ConstructionState{ProjectKind: core.ConstructionProjectHousing, ProjectID: HousingProjectID},
	}
	got, err := rules.CalculatePopulationDynamicsForEmpire(colony, planet, core.Empire{RaceID: "human"}, core.ColonyEconomy{Food: 4, Production: 3.25})
	if err != nil {
		t.Fatal(err)
	}
	// Original Housing adds integer percentage points: floor(40 * 3.25 / 4) = 32.
	if !closePopulationValue(got.GrowthMultiplier, 1.32) {
		t.Fatalf("Housing multiplier=%v want=1.32 dynamics=%+v", got.GrowthMultiplier, got)
	}

	cyber, err := rules.CalculatePopulationDynamicsForEmpire(colony, planet, core.Empire{RaceID: "meklar"}, core.ColonyEconomy{Food: 2, Production: 4})
	if err != nil {
		t.Fatal(err)
	}
	wantMultiplier := rules.RaceModifiers["meklar"].PopulationGrowthMultiplier + 0.20
	if !closePopulationValue(cyber.ProductionRequired, 2) || !closePopulationValue(cyber.ProductionAvailable, 2) || !closePopulationValue(cyber.GrowthMultiplier, wantMultiplier) {
		t.Fatalf("Cybernetic Housing dynamics=%+v want available=2 multiplier=%v", cyber, wantMultiplier)
	}
}

func TestPopulationGrowthCloningCenterIsFlatAndNotMultiplied(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	modifiers := rules.RaceModifiers["human"]
	modifiers.PopulationGrowthMultiplier = 2
	rules.RaceModifiers["human"] = modifiers
	planet := core.Planet{SizeID: "medium", ClimateID: "terran"}
	colony := core.Colony{
		Population: core.NewAssimilatedPopulation(1, 2, 1, 1),
		Buildings:  []string{rules.CloningCenterBuildingID},
	}
	got, err := rules.CalculatePopulationDynamicsForEmpire(colony, planet, core.Empire{RaceID: "human"}, core.ColonyEconomy{Food: 4, Production: 3})
	if err != nil {
		t.Fatal(err)
	}
	base := math.Sqrt(0.002 * 4 * (12 - 4) / 12)
	want := base*2 + 0.1
	if !closePopulationValue(got.ProjectedGrowth, want) {
		t.Fatalf("Cloning Center projected growth=%v want=%v dynamics=%+v", got.ProjectedGrowth, want, got)
	}
}

func TestPopulationGrowthCombinedOriginalStackingAndCapacityClamp(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	modifiers := rules.RaceModifiers["human"]
	modifiers.PopulationGrowthMultiplier = 1.5
	rules.RaceModifiers["human"] = modifiers
	planet := core.Planet{SizeID: "medium", ClimateID: "terran"}
	colony := core.Colony{
		Population:   core.NewAssimilatedPopulation(1, 2, 1, 1),
		Buildings:    []string{rules.CloningCenterBuildingID},
		Construction: &core.ConstructionState{ProjectKind: core.ConstructionProjectHousing, ProjectID: HousingProjectID},
	}
	got, err := rules.CalculatePopulationDynamicsForEmpire(colony, planet, core.Empire{RaceID: "human", KnownTechnologyIDs: []int{107, 193}}, core.ColonyEconomy{Food: 4, Production: 3.25})
	if err != nil {
		t.Fatal(err)
	}
	base := math.Sqrt(0.002 * 4 * (12 - 4) / 12)
	// 1.50 race + .25 Microbiotics + .50 Universal Antidote + .32 Housing.
	wantMultiplier := 2.57
	want := base*wantMultiplier + 0.1
	if !closePopulationValue(got.GrowthMultiplier, wantMultiplier) || !closePopulationValue(got.ProjectedGrowth, want) {
		t.Fatalf("combined growth=%+v want multiplier=%v projected=%v", got, wantMultiplier, want)
	}

	colony.Population = core.NewAssimilatedPopulation(colony.EmpireID, 4, 4, 3.95)
	colony.Construction = nil
	clamped, err := rules.CalculatePopulationDynamicsForEmpire(colony, planet, core.Empire{RaceID: "human"}, core.ColonyEconomy{Food: 12, Production: 3})
	if err != nil {
		t.Fatal(err)
	}
	if !closePopulationValue(clamped.ProjectedGrowth, 0.05) {
		t.Fatalf("capacity-clamped growth=%v want=0.05 dynamics=%+v", clamped.ProjectedGrowth, clamped)
	}
}

func TestHousingChoiceQueueAndCapacityAutoStop(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0xA130)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	choices, err := rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, choice := range choices {
		if choice.ProjectKind == core.ConstructionProjectHousing {
			found = choice.ProjectID == HousingProjectID && choice.ProductionCostPP == 0
		}
	}
	if !found {
		t.Fatalf("Housing missing from legal construction choices: %+v", choices)
	}

	command, err := NewQueueHousingCommand(9, QueueHousingPayload{ColonyID: colony.ID})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.queueHousing(state, empire.ID, protocol.SeatID(1), command)
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != "colony.housing_queued" || colony.Construction == nil || colony.Construction.ProjectKind != core.ConstructionProjectHousing {
		t.Fatalf("Housing queue result event=%+v construction=%+v", event, colony.Construction)
	}

	planet := planetByID(state, colony.PlanetID)
	capacity, err := rules.PopulationCapacity(*planet, empire.RaceID)
	if err != nil {
		t.Fatal(err)
	}
	colony.Population = core.NewAssimilatedPopulation(colony.EmpireID, capacity, 0, 0)
	colony.PopulationDynamics.Capacity = capacity
	events, err := resolver.advanceConstruction(state)
	if err != nil {
		t.Fatal(err)
	}
	if colony.Construction != nil {
		t.Fatalf("Housing was not cleared at population capacity: %+v", colony.Construction)
	}
	if len(events) != 1 || events[0].Kind != "colony.housing_stopped" {
		t.Fatalf("Housing stop events=%+v", events)
	}
}
