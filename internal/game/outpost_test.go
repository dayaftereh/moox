package game

import (
	"testing"

	"moox/internal/core"
)

func outpostShipChoice(choices []ConstructionChoice) *ConstructionChoice {
	for i := range choices {
		if choices[i].ProjectKind == core.ConstructionProjectOutpostShip {
			return &choices[i]
		}
	}
	return nil
}

func addFixedSpecialShipTestFleet(state *core.GameState, empireID, atSystemID core.ID, kind core.StrategicFleetSpecialKind, ftlSpeed int) core.ID {
	fleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID:          fleetID,
		EmpireID:    empireID,
		Role:        core.StrategicFleetRoleCivilian,
		SpecialKind: kind,
		AtSystemID:  atSystemID,
		FTLSpeed:    ftlSpeed,
	})
	return fleetID
}

func addOutpostTestState(state *core.GameState, empireID, planetID core.ID) core.ID {
	planet := planetByID(state, planetID)
	if planet == nil {
		panic("test planet not found")
	}
	outpostID := state.NewID()
	state.Outposts = append(state.Outposts, core.Outpost{ID: outpostID, EmpireID: empireID, PlanetID: planetID})
	planet.OutpostID = outpostID
	return outpostID
}

func TestOutpostShipConstructionChoiceRequiresTechnologyAndUsesOriginalCost(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(1720)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	empire.KnownTechnologyIDs = nil

	choices, err := rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	if choice := outpostShipChoice(choices); choice != nil {
		t.Fatalf("Outpost Ship unexpectedly available without Technology %d: %+v", OutpostShipTechnologyID, *choice)
	}

	addColonyShipTestTechnology(empire, OutpostShipTechnologyID)
	choices, err = rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	choice := outpostShipChoice(choices)
	if choice == nil || choice.ProjectID != OutpostShipProjectID || choice.TechnologyID != OutpostShipTechnologyID || choice.ProductionCostPP != 100 {
		t.Fatalf("ordinary Outpost Ship choice=%+v", choice)
	}

	modifiers := rules.RaceModifiers[empire.RaceID]
	modifiers.GovernmentTraitID = "government_feudal"
	rules.RaceModifiers[empire.RaceID] = modifiers
	choices, err = rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	choice = outpostShipChoice(choices)
	if choice == nil || choice.ProductionCostPP != 67 {
		t.Fatalf("Feudal Outpost Ship choice=%+v want cost 67", choice)
	}
}

func TestQueueOutpostShipUsesGenericConstructionStateAndRequiresTechnology(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1721)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	empire.KnownTechnologyIDs = nil
	command, err := NewQueueOutpostShipCommand(5, QueueOutpostShipPayload{ColonyID: colony.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.queueOutpostShip(state, empire.ID, 1, command); err == nil {
		t.Fatal("Outpost Ship queue unexpectedly succeeded without Technology 109")
	}
	addColonyShipTestTechnology(empire, OutpostShipTechnologyID)
	event, err := resolver.queueOutpostShip(state, empire.ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if colony.Construction == nil || colony.Construction.ProjectKind != core.ConstructionProjectOutpostShip || colony.Construction.ProjectID != OutpostShipProjectID || colony.Construction.ProgressPP != 0 {
		t.Fatalf("Outpost Ship construction state=%+v", colony.Construction)
	}
	if event.Kind != "colony.outpost_ship_queued" || event.SeatID != 1 || event.CommandSequence != 5 {
		t.Fatalf("queue event=%+v", event)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("queued Outpost Ship state invalid: %v", err)
	}
}

func TestOutpostShipCompletionCreatesStationaryCivilianFleetWithInstalledDrive(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1722)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	addColonyShipTestTechnology(empire, OutpostShipTechnologyID, 120)
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectOutpostShip, ProjectID: OutpostShipProjectID}
	colony.PopulationDynamics.ProductionAvailable = 100

	events, err := resolver.advanceConstruction(state)
	if err != nil {
		t.Fatal(err)
	}
	if colony.Construction != nil {
		t.Fatalf("completed Outpost Ship retained construction state: %+v", colony.Construction)
	}
	if len(state.StrategicFleets) != 1 {
		t.Fatalf("strategic fleets=%d want=1", len(state.StrategicFleets))
	}
	fleet := state.StrategicFleets[0]
	if fleet.EmpireID != empire.ID || fleet.Role != core.StrategicFleetRoleCivilian || fleet.SpecialKind != core.StrategicFleetSpecialOutpostShip || fleet.AtSystemID != state.Galaxy.Systems[0].ID || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 || fleet.FTLSpeed != 2 {
		t.Fatalf("completed Outpost Ship fleet=%+v", fleet)
	}
	if findDomainEvent(events, "colony.outpost_ship_completed") == nil {
		t.Fatalf("completion events=%+v", events)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("completed Outpost Ship state invalid: %v", err)
	}
}

func TestOutpostDeployImmediatelyExtendsSupplyEvenWhileBlockaded(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1723)
	empire := &state.Empires[0]
	addColonyShipTestTechnology(empire, standardFuelCellsTechnologyID)
	alpha := &state.Galaxy.Systems[0]
	beta := &state.Galaxy.Systems[1]
	gamma := &state.Galaxy.Systems[2]
	alpha.X, alpha.Y = 0, 0
	beta.X, beta.Y = 90, 0
	gamma.X, gamma.Y = 210, 0

	outpostFleetID := addFixedSpecialShipTestFleet(state, empire.ID, alpha.ID, core.StrategicFleetSpecialOutpostShip, 2)
	colonyFleetID := addFixedSpecialShipTestFleet(state, empire.ID, alpha.ID, core.StrategicFleetSpecialColonyShip, 2)
	moveColony, _ := NewMoveFleetCommand(1, MoveFleetPayload{FleetID: colonyFleetID, DestinationSystemID: gamma.ID})
	if _, err := resolver.moveFleet(state, empire.ID, 1, moveColony); err == nil {
		t.Fatal("7 pc Gamma destination unexpectedly reachable before Outpost deployment")
	}

	moveOutpost, _ := NewMoveFleetCommand(2, MoveFleetPayload{FleetID: outpostFleetID, DestinationSystemID: beta.ID})
	if _, err := resolver.moveFleet(state, empire.ID, 1, moveOutpost); err != nil {
		t.Fatalf("Outpost Ship could not start legal 3 pc transit: %v", err)
	}
	if _, err := resolver.advanceStrategicFleetTransit(state); err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.advanceStrategicFleetTransit(state); err != nil {
		t.Fatal(err)
	}
	_, outpostFleet := strategicFleetByID(state, outpostFleetID)
	if outpostFleet == nil || outpostFleet.AtSystemID != beta.ID {
		t.Fatalf("Outpost Ship did not arrive at Beta: %+v", outpostFleet)
	}

	target := &beta.Planets[0]
	deploy, err := NewDeployOutpostCommand(3, DeployOutpostPayload{FleetID: outpostFleetID, PlanetID: target.ID})
	if err != nil {
		t.Fatal(err)
	}
	events, err := resolver.deployOutpost(state, empire.ID, 1, deploy)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Kind != "empire.outpost_deployed" || events[1].Kind != "empire.outpost_ship_consumed" {
		t.Fatalf("deployment events=%+v", events)
	}
	if len(state.Outposts) != 1 || target.OutpostID != state.Outposts[0].ID || state.Outposts[0].EmpireID != empire.ID {
		t.Fatalf("Outpost state=%+v target=%+v", state.Outposts, target)
	}
	if _, fleet := strategicFleetByID(state, outpostFleetID); fleet != nil {
		t.Fatalf("consumed Outpost Ship remains: %+v", *fleet)
	}

	beta.BlockadedEmpireIDs = []core.ID{empire.ID}
	if _, err := resolver.moveFleet(state, empire.ID, 1, moveColony); err != nil {
		t.Fatalf("Beta Outpost did not immediately extend 4 pc supply to Gamma while blockaded: %v", err)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("post-Outpost supply state invalid: %v", err)
	}
}

func TestColonyShipConvertsSameOwnerOutpost(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1724)
	empire := &state.Empires[0]
	targetSystem := &state.Galaxy.Systems[1]
	target := &targetSystem.Planets[0]
	outpostID := addOutpostTestState(state, empire.ID, target.ID)
	fleetID := addFixedSpecialShipTestFleet(state, empire.ID, targetSystem.ID, core.StrategicFleetSpecialColonyShip, 2)
	command, _ := NewColonizePlanetCommand(4, ColonizePlanetPayload{FleetID: fleetID, PlanetID: target.ID})
	events, err := resolver.colonizePlanet(state, empire.ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 || events[0].Kind != "empire.planet_colonized" || events[1].Kind != "empire.outpost_converted" || events[2].Kind != "empire.colony_ship_consumed" {
		t.Fatalf("conversion events=%+v", events)
	}
	if len(state.Outposts) != 0 || target.OutpostID != 0 || target.ColonyID == 0 || len(state.Colonies) != 2 {
		t.Fatalf("conversion state outposts=%+v target=%+v colonies=%+v", state.Outposts, target, state.Colonies)
	}
	if outpostID == 0 || state.Colonies[1].PlanetID != target.ID || state.Colonies[1].EmpireID != empire.ID {
		t.Fatalf("converted Colony=%+v old outpost=%d", state.Colonies[1], outpostID)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("converted state invalid: %v", err)
	}
}

func TestColonyShipRejectsForeignOutpostConversion(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1725)
	owner := &state.Empires[0]
	foreign := core.Empire{ID: state.NewID(), Name: "Foreign", RaceID: "alkari"}
	state.Empires = append(state.Empires, foreign)
	targetSystem := &state.Galaxy.Systems[1]
	target := &targetSystem.Planets[0]
	outpostID := addOutpostTestState(state, foreign.ID, target.ID)
	fleetID := addFixedSpecialShipTestFleet(state, owner.ID, targetSystem.ID, core.StrategicFleetSpecialColonyShip, 2)
	command, _ := NewColonizePlanetCommand(1, ColonizePlanetPayload{FleetID: fleetID, PlanetID: target.ID})
	if _, err := resolver.colonizePlanet(state, owner.ID, 1, command); err == nil {
		t.Fatal("Colony Ship unexpectedly converted foreign Outpost")
	}
	if len(state.Outposts) != 1 || state.Outposts[0].ID != outpostID || target.OutpostID != outpostID || target.ColonyID != 0 {
		t.Fatalf("failed foreign conversion mutated state: outposts=%+v target=%+v", state.Outposts, target)
	}
}

func TestColonyBaseConvertsSameOwnerOutpost(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1726)
	empire := &state.Empires[0]
	source := &state.Colonies[0]
	home := &state.Galaxy.Systems[0]
	targetID := state.NewID()
	home.Planets = append(home.Planets, core.Planet{ID: targetID, Name: "Alpha II", Orbit: 2, SizeID: "small", MineralID: "poor", GravityID: "normal_g", ClimateID: "barren"})
	target := &home.Planets[len(home.Planets)-1]
	outpostID := addOutpostTestState(state, empire.ID, target.ID)
	source.Buildings = append(source.Buildings, ColonyBaseBuildingID)
	command, err := NewColonizeWithBaseCommand(6, ColonizeWithBasePayload{SourceColonyID: source.ID, PlanetID: target.ID})
	if err != nil {
		t.Fatal(err)
	}
	events, err := resolver.colonizeWithBase(state, empire.ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 || events[0].Kind != "empire.planet_colonized" || events[1].Kind != "empire.outpost_converted" || events[2].Kind != "colony.colony_base_colonized" {
		t.Fatalf("Colony Base conversion events=%+v", events)
	}
	if len(state.Outposts) != 0 || target.OutpostID != 0 || target.ColonyID == 0 || colonyOwnsBuilding(source, ColonyBaseBuildingID) {
		t.Fatalf("Colony Base conversion state oldOutpost=%d outposts=%+v target=%+v buildings=%v", outpostID, state.Outposts, target, source.Buildings)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("Colony Base converted state invalid: %v", err)
	}
}
