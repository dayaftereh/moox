package game

import (
	"reflect"
	"testing"

	"moox/internal/core"
)

func commandPointTestResolver(t *testing.T) (*EconomyRules, *EconomyResolver) {
	t.Helper()
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	return rules, resolver
}

func TestCommittedCommandPointRulesMatchOriginalEvidence(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	cp := rules.CommandPoints
	if cp.BaseCapacity != 5 || cp.FixedSpecialShipPoints != 1 || cp.WarlordPointsPerColony != 2 || cp.StandardOverageBCPerPoint != 10 {
		t.Fatalf("unexpected Command Point scalar rules: %+v", cp)
	}
	wantStations := map[string]int{"star_base": 1, "battlestation": 2, "star_fortress": 3}
	if !reflect.DeepEqual(cp.StationPoints, wantStations) {
		t.Fatalf("station points=%v want=%v", cp.StationPoints, wantStations)
	}
	wantCommunications := map[int]int{180: 1, 176: 2, 91: 3}
	if !reflect.DeepEqual(cp.CommunicationsPointsByTechnologyID, wantCommunications) {
		t.Fatalf("communications points=%v want=%v", cp.CommunicationsPointsByTechnologyID, wantCommunications)
	}
	if cp.ImperiumTechnologyID != 92 || cp.ImperiumRequiredGovernmentTraitID != "government_dictatorship" || cp.ImperiumBonusNumerator != 1 || cp.ImperiumBonusDenominator != 2 {
		t.Fatalf("Imperium rules=%+v", cp)
	}
}

func TestCommandPointUsageUsesHullSizeAndFixedSpecialShips(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	for _, test := range []struct {
		hullID string
		want   int
	}{
		{"frigate", 1},
		{"destroyer", 2},
		{"cruiser", 3},
		{"battleship", 4},
		{"titan", 5},
		{"doom_star", 6},
	} {
		t.Run(test.hullID, func(t *testing.T) {
			state := core.NewSmallFixture(0xA501)
			empire := &state.Empires[0]
			state.Ships = append(state.Ships, core.Ship{ID: state.NewID(), EmpireID: empire.ID, Spec: core.ShipDesignSpec{HullID: test.hullID}})
			points, err := rules.deriveEmpireCommandPoints(state, empire)
			if err != nil {
				t.Fatal(err)
			}
			if points.Used != test.want {
				t.Fatalf("hull %s used=%d want=%d", test.hullID, points.Used, test.want)
			}
		})
	}

	state := core.NewSmallFixture(0xA502)
	empire := &state.Empires[0]
	state.StrategicFleets = append(state.StrategicFleets,
		core.StrategicFleet{ID: state.NewID(), EmpireID: empire.ID, Role: core.StrategicFleetRoleCivilian, SpecialKind: core.StrategicFleetSpecialColonyShip, AtSystemID: state.Galaxy.Systems[0].ID, FTLSpeed: 2},
		core.StrategicFleet{ID: state.NewID(), EmpireID: empire.ID, Role: core.StrategicFleetRoleCivilian, SpecialKind: core.StrategicFleetSpecialOutpostShip, DestinationSystemID: state.Galaxy.Systems[1].ID, RemainingTurns: 2, FTLSpeed: 2},
	)
	empire.Freighters = 20
	points, err := rules.deriveEmpireCommandPoints(state, empire)
	if err != nil {
		t.Fatal(err)
	}
	if points.Used != 2 {
		t.Fatalf("fixed special usage=%d want=2", points.Used)
	}
}

func TestCommandPointCapacityStationsCommunicationsWarlordAndImperium(t *testing.T) {
	rules := loadCommittedEconomyRules(t)

	state := core.NewSmallFixture(0xA503)
	empire := &state.Empires[0]
	points, err := rules.deriveEmpireCommandPoints(state, empire)
	if err != nil {
		t.Fatal(err)
	}
	if points.Capacity != 5 {
		t.Fatalf("base capacity=%d want=5", points.Capacity)
	}

	for _, test := range []struct {
		buildingID string
		want       int
	}{
		{"star_base", 6},
		{"battlestation", 7},
		{"star_fortress", 8},
	} {
		t.Run(test.buildingID, func(t *testing.T) {
			state := core.NewSmallFixture(0xA504)
			state.Colonies[0].Buildings = []string{test.buildingID}
			points, err := rules.deriveEmpireCommandPoints(state, &state.Empires[0])
			if err != nil {
				t.Fatal(err)
			}
			if points.Capacity != test.want {
				t.Fatalf("capacity=%d want=%d", points.Capacity, test.want)
			}
		})
	}

	for _, test := range []struct {
		techIDs []int
		want    int
	}{
		{[]int{180}, 7},
		{[]int{176}, 8},
		{[]int{91}, 9},
		{[]int{91, 176, 180}, 9},
	} {
		state := core.NewSmallFixture(0xA505)
		state.Colonies[0].Buildings = []string{"star_base"}
		state.Empires[0].KnownTechnologyIDs = append([]int(nil), test.techIDs...)
		points, err := rules.deriveEmpireCommandPoints(state, &state.Empires[0])
		if err != nil {
			t.Fatal(err)
		}
		if points.Capacity != test.want {
			t.Fatalf("communications techs %v capacity=%d want=%d", test.techIDs, points.Capacity, test.want)
		}
	}

	state = core.NewSmallFixture(0xA506)
	state.Empires[0].RaceID = "mrrshan"
	points, err = rules.deriveEmpireCommandPoints(state, &state.Empires[0])
	if err != nil {
		t.Fatal(err)
	}
	if points.Capacity != 7 {
		t.Fatalf("Warlord capacity=%d want=7", points.Capacity)
	}

	state = core.NewSmallFixture(0xA507)
	state.Empires[0].RaceID = "darlok"
	state.Empires[0].KnownTechnologyIDs = []int{92}
	points, err = rules.deriveEmpireCommandPoints(state, &state.Empires[0])
	if err != nil {
		t.Fatal(err)
	}
	if points.Capacity != 7 {
		t.Fatalf("odd Imperium base capacity=%d want=7", points.Capacity)
	}
	state.Colonies[0].Buildings = []string{"star_base"}
	points, err = rules.deriveEmpireCommandPoints(state, &state.Empires[0])
	if err != nil {
		t.Fatal(err)
	}
	if points.Capacity != 9 {
		t.Fatalf("Imperium with Star Base capacity=%d want=9", points.Capacity)
	}

	state = core.NewSmallFixture(0xA508)
	state.Empires[0].KnownTechnologyIDs = []int{92}
	points, err = rules.deriveEmpireCommandPoints(state, &state.Empires[0])
	if err != nil {
		t.Fatal(err)
	}
	if points.Capacity != 5 {
		t.Fatalf("non-Dictatorship Imperium-tech capacity=%d want=5", points.Capacity)
	}
}

func TestTreasuryCommandPointOverageAndAtomicValidation(t *testing.T) {
	_, resolver := commandPointTestResolver(t)
	state := core.NewSmallFixture(0xA509)
	empire := &state.Empires[0]
	for i := 0; i < 6; i++ {
		state.Ships = append(state.Ships, core.Ship{ID: state.NewID(), EmpireID: empire.ID, Spec: core.ShipDesignSpec{HullID: "frigate"}})
	}
	events, err := resolver.settleTreasury(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || empire.CommandPoints != (core.EmpireCommandPoints{Capacity: 5, Used: 6}) {
		t.Fatalf("settled Command Points=%+v events=%+v", empire.CommandPoints, events)
	}
	if empire.Treasury.ShipCommandMaintenanceBC != 10 || empire.Treasury.TotalModeledMaintenanceBC != 10 || empire.Treasury.BalanceBC != -10 {
		t.Fatalf("overage Treasury=%+v", empire.Treasury)
	}

	state = core.NewSmallFixture(0xA50A)
	empire = &state.Empires[0]
	for i := 0; i < 5; i++ {
		state.Ships = append(state.Ships, core.Ship{ID: state.NewID(), EmpireID: empire.ID, Spec: core.ShipDesignSpec{HullID: "frigate"}})
	}
	if _, err := resolver.settleTreasury(state); err != nil {
		t.Fatal(err)
	}
	if empire.Treasury.ShipCommandMaintenanceBC != 0 {
		t.Fatalf("at-capacity ship Maintenance=%v want=0", empire.Treasury.ShipCommandMaintenanceBC)
	}

	state = core.NewSmallFixture(0xA50B)
	state.Empires[0].Treasury.BalanceBC = 50
	state.Empires[0].CommandPoints = core.EmpireCommandPoints{Capacity: 11, Used: 4}
	secondID := state.NewID()
	state.Empires = append(state.Empires, core.Empire{ID: secondID, Name: "Second", RaceID: "human", Treasury: core.EmpireTreasuryState{BalanceBC: 60}, CommandPoints: core.EmpireCommandPoints{Capacity: 12, Used: 5}})
	state.Ships = append(state.Ships, core.Ship{ID: state.NewID(), EmpireID: secondID, Spec: core.ShipDesignSpec{HullID: "unknown_hull"}})
	before := append([]core.Empire(nil), state.Empires...)
	if _, err := resolver.settleTreasury(state); err == nil {
		t.Fatal("expected invalid hull settlement to fail")
	}
	if !reflect.DeepEqual(state.Empires, before) {
		t.Fatalf("failed settlement partially mutated Empires:\nbefore=%+v\nafter=%+v", before, state.Empires)
	}
}

func TestCommandStationReplacementAndBuildability(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	colony := core.Colony{ID: 1, Buildings: []string{"holo_simulator", "star_base"}}
	if err := applyCompletedCommandStation(&colony, "battlestation"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(colony.Buildings, []string{"holo_simulator", "battlestation"}) {
		t.Fatalf("Battlestation replacement=%v", colony.Buildings)
	}
	if err := applyCompletedCommandStation(&colony, "star_fortress"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(colony.Buildings, []string{"holo_simulator", "star_fortress"}) {
		t.Fatalf("Star Fortress replacement=%v", colony.Buildings)
	}
	before := append([]string(nil), colony.Buildings...)
	if err := applyCompletedCommandStation(&colony, "star_base"); err == nil {
		t.Fatal("expected lower-tier replacement to fail")
	}
	if !reflect.DeepEqual(colony.Buildings, before) {
		t.Fatalf("rejected lower-tier replacement mutated buildings=%v", colony.Buildings)
	}

	state := core.NewSmallFixture(0xA50C)
	empire := &state.Empires[0]
	colonyState := &state.Colonies[0]
	colonyState.Buildings = []string{"battlestation"}
	empire.KnownTechnologyIDs = []int{
		rules.BuildingDefinitions["star_base"].TechnologyID,
		rules.BuildingDefinitions["battlestation"].TechnologyID,
		rules.BuildingDefinitions["star_fortress"].TechnologyID,
	}
	// AvailableBuildingChoices relies on sorted technology IDs.
	for i := 0; i < len(empire.KnownTechnologyIDs); i++ {
		for j := i + 1; j < len(empire.KnownTechnologyIDs); j++ {
			if empire.KnownTechnologyIDs[j] < empire.KnownTechnologyIDs[i] {
				empire.KnownTechnologyIDs[i], empire.KnownTechnologyIDs[j] = empire.KnownTechnologyIDs[j], empire.KnownTechnologyIDs[i]
			}
		}
	}
	choices, err := rules.AvailableBuildingChoices(state, empire.ID, colonyState.ID)
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, choice := range choices {
		seen[choice.BuildingID] = true
	}
	if seen["star_base"] || seen["battlestation"] || !seen["star_fortress"] {
		t.Fatalf("Battlestation build choices=%v", seen)
	}
}

func TestCommandPointsShipAndStationCompletionApplyNextSettlement(t *testing.T) {
	rules, resolver := commandPointTestResolver(t)
	state := core.NewSmallFixture(0xA50D)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]

	addCombatFleetTestFleet(state, empire.ID, state.Galaxy.Systems[0].ID, 5)
	if _, err := resolver.settleTreasury(state); err != nil {
		t.Fatal(err)
	}
	if empire.CommandPoints.Used != 5 || empire.Treasury.ShipCommandMaintenanceBC != 0 {
		t.Fatalf("pre-construction snapshot CP=%+v Treasury=%+v", empire.CommandPoints, empire.Treasury)
	}

	spec := combatFleetTestSpec()
	designID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{ID: designID, EmpireID: empire.ID, Revision: 1, Name: "Sixth", Spec: spec})
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectMilitaryShip, ProjectID: MilitaryShipProjectID, ShipDesignID: designID, ShipDesignRevision: 1}
	colony.PopulationDynamics.ProductionAvailable = float64(spec.ProductionCostPP)
	if _, err := resolver.advanceConstruction(state); err != nil {
		t.Fatal(err)
	}
	if empire.CommandPoints.Used != 5 || empire.Treasury.ShipCommandMaintenanceBC != 0 {
		t.Fatalf("construction retroactively changed settlement CP=%+v Treasury=%+v", empire.CommandPoints, empire.Treasury)
	}
	if _, err := resolver.settleTreasury(state); err != nil {
		t.Fatal(err)
	}
	if empire.CommandPoints.Used != 6 || empire.Treasury.ShipCommandMaintenanceBC != 10 {
		t.Fatalf("next settlement CP=%+v Treasury=%+v", empire.CommandPoints, empire.Treasury)
	}

	state = core.NewSmallFixture(0xA50E)
	empire = &state.Empires[0]
	colony = &state.Colonies[0]
	colony.Buildings = []string{"star_base"}
	if _, err := resolver.settleTreasury(state); err != nil {
		t.Fatal(err)
	}
	if empire.CommandPoints.Capacity != 6 || empire.Treasury.BuildingMaintenanceBC != float64(rules.BuildingDefinitions["star_base"].MaintenanceBC) {
		t.Fatalf("pre-upgrade CP=%+v Treasury=%+v", empire.CommandPoints, empire.Treasury)
	}
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectBuilding, ProjectID: "battlestation"}
	colony.PopulationDynamics.ProductionAvailable = rules.BuildingDefinitions["battlestation"].ProductionCostPP
	if _, err := resolver.advanceConstruction(state); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(colony.Buildings, []string{"battlestation"}) {
		t.Fatalf("station completion did not replace lower tier: %v", colony.Buildings)
	}
	if empire.CommandPoints.Capacity != 6 {
		t.Fatalf("station completion retroactively changed CP=%+v", empire.CommandPoints)
	}
	if _, err := resolver.settleTreasury(state); err != nil {
		t.Fatal(err)
	}
	if empire.CommandPoints.Capacity != 7 || empire.Treasury.BuildingMaintenanceBC != float64(rules.BuildingDefinitions["battlestation"].MaintenanceBC) {
		t.Fatalf("next station settlement CP=%+v Treasury=%+v", empire.CommandPoints, empire.Treasury)
	}
}

func TestCommandPointSnapshotRoundTripsSchema20(t *testing.T) {
	_, resolver := commandPointTestResolver(t)
	state := core.NewSmallFixture(0xA50F)
	state.Colonies[0].Buildings = []string{"star_base"}
	if _, err := resolver.settleTreasury(state); err != nil {
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
	if loaded.SchemaVersion != 20 || loaded.Empires[0].CommandPoints != state.Empires[0].CommandPoints || loaded.Empires[0].Treasury != state.Empires[0].Treasury {
		t.Fatalf("round-trip changed CP/Treasury: before=%+v/%+v after=%+v/%+v", state.Empires[0].CommandPoints, state.Empires[0].Treasury, loaded.Empires[0].CommandPoints, loaded.Empires[0].Treasury)
	}
}

func TestInvalidMultipleCommandStationsRejectSettlementWithoutMutation(t *testing.T) {
	_, resolver := commandPointTestResolver(t)
	state := core.NewSmallFixture(0xA510)
	state.Empires[0].Treasury.BalanceBC = 25
	state.Empires[0].CommandPoints = core.EmpireCommandPoints{Capacity: 8, Used: 2}
	state.Colonies[0].Buildings = []string{"star_base", "battlestation"}
	beforeEmpire := state.Empires[0]
	beforeBuildings := append([]string(nil), state.Colonies[0].Buildings...)
	if _, err := resolver.settleTreasury(state); err == nil {
		t.Fatal("expected multiple command station tiers to reject settlement")
	}
	if !reflect.DeepEqual(state.Empires[0], beforeEmpire) || !reflect.DeepEqual(state.Colonies[0].Buildings, beforeBuildings) {
		t.Fatalf("rejected settlement mutated state empire=%+v buildings=%v", state.Empires[0], state.Colonies[0].Buildings)
	}
}

func TestCommandPointUsageIsInvariantAcrossCombatFleetSplitMergeAndMovement(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0xA511)
	empire := &state.Empires[0]
	setCombatFleetTestTech(empire)
	fleetID, shipIDs := addCombatFleetTestFleet(state, empire.ID, state.Galaxy.Systems[0].ID, 3)

	assertUsed := func(stage string) {
		t.Helper()
		points, err := rules.deriveEmpireCommandPoints(state, empire)
		if err != nil {
			t.Fatal(err)
		}
		if points.Used != 3 {
			t.Fatalf("%s used=%d want=3", stage, points.Used)
		}
	}
	assertUsed("before split")

	newFleetID := state.NextID
	split, err := NewSplitFleetCommand(1, SplitFleetPayload{FleetID: fleetID, ShipIDs: []core.ID{shipIDs[2]}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.splitFleet(state, empire.ID, 1, split); err != nil {
		t.Fatal(err)
	}
	assertUsed("after split")

	merge, err := NewMergeFleetsCommand(2, MergeFleetsPayload{TargetFleetID: fleetID, SourceFleetID: newFleetID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.mergeFleets(state, empire.ID, 1, merge); err != nil {
		t.Fatal(err)
	}
	assertUsed("after merge")

	move, err := NewMoveFleetCommand(3, MoveFleetPayload{FleetID: fleetID, DestinationSystemID: state.Galaxy.Systems[1].ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.moveFleetEvents(state, empire.ID, 1, move); err != nil {
		t.Fatal(err)
	}
	assertUsed("after movement start")
}

func TestCommandPointUsageDropsWhenOutpostShipIsLegallyConsumed(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0xA512)
	empire := &state.Empires[0]
	targetSystem := &state.Galaxy.Systems[1]
	targetPlanet := &targetSystem.Planets[0]
	fleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID: fleetID, EmpireID: empire.ID, Role: core.StrategicFleetRoleCivilian,
		SpecialKind: core.StrategicFleetSpecialOutpostShip, AtSystemID: targetSystem.ID, FTLSpeed: 2,
	})

	before, err := rules.deriveEmpireCommandPoints(state, empire)
	if err != nil {
		t.Fatal(err)
	}
	if before.Used != 1 {
		t.Fatalf("Outpost Ship pre-deployment used=%d want=1", before.Used)
	}
	command, err := NewDeployOutpostCommand(1, DeployOutpostPayload{FleetID: fleetID, PlanetID: targetPlanet.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.deployOutpost(state, empire.ID, 1, command); err != nil {
		t.Fatal(err)
	}
	after, err := rules.deriveEmpireCommandPoints(state, empire)
	if err != nil {
		t.Fatal(err)
	}
	if after.Used != 0 {
		t.Fatalf("consumed Outpost Ship used=%d want=0", after.Used)
	}
}

func TestCommandPointUsageIgnoresPopulationTransfers(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(0xA513)
	empire := &state.Empires[0]
	targetPlanet := &state.Galaxy.Systems[1].Planets[0]
	targetColonyID := state.NewID()
	targetPlanet.ColonyID = targetColonyID
	state.Colonies = append(state.Colonies, core.Colony{
		ID: targetColonyID, EmpireID: empire.ID, PlanetID: targetPlanet.ID,
		Population: core.NewAssimilatedPopulation(empire.ID, 1, 0, 0),
	})
	state.PopulationTransfers = append(state.PopulationTransfers, core.PopulationTransfer{
		ID: state.NewID(), EmpireID: empire.ID, SourceColonyID: state.Colonies[0].ID,
		DestinationColonyID: targetColonyID, OriginEmpireID: empire.ID, LoyaltyEmpireID: empire.ID,
		AssimilationState: core.PopulationAssimilated, Job: core.PopulationJobFarmer, RemainingTurns: 2,
	})
	points, err := rules.deriveEmpireCommandPoints(state, empire)
	if err != nil {
		t.Fatal(err)
	}
	if points.Used != 0 {
		t.Fatalf("Population transfer used=%d want=0", points.Used)
	}
}
