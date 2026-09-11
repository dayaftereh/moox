package game

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func loadColonyShipRules(t *testing.T) *EconomyRules {
	t.Helper()
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	return rules
}

func addColonyShipTestTechnology(empire *core.Empire, technologyIDs ...int) {
	for _, technologyID := range technologyIDs {
		empire.KnownTechnologyIDs = appendUniqueSortedInt(empire.KnownTechnologyIDs, technologyID)
	}
}

func colonyShipChoice(choices []ConstructionChoice) *ConstructionChoice {
	for i := range choices {
		if choices[i].ProjectKind == core.ConstructionProjectColonyShip {
			return &choices[i]
		}
	}
	return nil
}

func addColonyShipTestFleet(state *core.GameState, empireID, atSystemID core.ID, ftlSpeed int) core.ID {
	fleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID:          fleetID,
		EmpireID:    empireID,
		Role:        core.StrategicFleetRoleCivilian,
		SpecialKind: core.StrategicFleetSpecialColonyShip,
		AtSystemID:  atSystemID,
		FTLSpeed:    ftlSpeed,
	})
	return fleetID
}

func eventIndex(events []DomainEvent, kind string) int {
	for i := range events {
		if events[i].Kind == kind {
			return i
		}
	}
	return -1
}

func TestColonyShipConstructionChoiceRequiresTechnologyAndUsesOriginalCost(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(1701)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	empire.KnownTechnologyIDs = nil

	choices, err := rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	if choice := colonyShipChoice(choices); choice != nil {
		t.Fatalf("Colony Ship unexpectedly available without Technology %d: %+v", ColonyShipTechnologyID, *choice)
	}

	addColonyShipTestTechnology(empire, ColonyShipTechnologyID)
	choices, err = rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	choice := colonyShipChoice(choices)
	if choice == nil {
		t.Fatal("Technology 41 did not expose Colony Ship construction")
	}
	if choice.ProjectID != ColonyShipProjectID || choice.TechnologyID != ColonyShipTechnologyID || choice.ProductionCostPP != 500 {
		t.Fatalf("ordinary Colony Ship choice=%+v", *choice)
	}

	modifiers := rules.RaceModifiers[empire.RaceID]
	modifiers.GovernmentTraitID = "government_feudal"
	rules.RaceModifiers[empire.RaceID] = modifiers
	choices, err = rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	choice = colonyShipChoice(choices)
	if choice == nil || choice.ProductionCostPP != 334 {
		t.Fatalf("Feudal Colony Ship choice=%+v want cost 334", choice)
	}
}

func TestQueueColonyShipUsesGenericConstructionStateAndRequiresTechnology(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1702)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	empire.KnownTechnologyIDs = nil
	command, err := NewQueueColonyShipCommand(7, QueueColonyShipPayload{ColonyID: colony.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.queueColonyShip(state, empire.ID, 1, command); err == nil {
		t.Fatal("Colony Ship queue unexpectedly succeeded without Technology 41")
	}
	addColonyShipTestTechnology(empire, ColonyShipTechnologyID)
	event, err := resolver.queueColonyShip(state, empire.ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if colony.Construction == nil || colony.Construction.ProjectKind != core.ConstructionProjectColonyShip || colony.Construction.ProjectID != ColonyShipProjectID || colony.Construction.ProgressPP != 0 {
		t.Fatalf("Colony Ship construction state=%+v", colony.Construction)
	}
	if event.Kind != "colony.colony_ship_queued" || event.SeatID != 1 || event.CommandSequence != 7 {
		t.Fatalf("queue event=%+v", event)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("queued Colony Ship state invalid: %v", err)
	}
}

func TestColonyShipCompletionCreatesStationaryCivilianFleetWithInstalledDrive(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1703)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	addColonyShipTestTechnology(empire, ColonyShipTechnologyID, 120)
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectColonyShip, ProjectID: ColonyShipProjectID}
	colony.PopulationDynamics.ProductionAvailable = 500

	events, err := resolver.advanceConstruction(state)
	if err != nil {
		t.Fatal(err)
	}
	if colony.Construction != nil {
		t.Fatalf("completed Colony Ship retained construction state: %+v", colony.Construction)
	}
	if len(state.StrategicFleets) != 1 {
		t.Fatalf("strategic fleets=%d want=1", len(state.StrategicFleets))
	}
	fleet := state.StrategicFleets[0]
	if fleet.EmpireID != empire.ID || fleet.Role != core.StrategicFleetRoleCivilian || fleet.SpecialKind != core.StrategicFleetSpecialColonyShip || fleet.AtSystemID != state.Galaxy.Systems[0].ID || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 || fleet.FTLSpeed != 2 {
		t.Fatalf("completed Colony Ship fleet=%+v", fleet)
	}
	completed := findDomainEvent(events, "colony.colony_ship_completed")
	if completed == nil {
		t.Fatalf("completion events=%+v", events)
	}
	var payload ColonyShipCompletedEvent
	if err := json.Unmarshal(completed.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.FleetID != fleet.ID || payload.SystemID != fleet.AtSystemID || payload.FTLSpeed != 2 {
		t.Fatalf("completion payload=%+v", payload)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("completed Colony Ship state invalid: %v", err)
	}
}

func TestStrategicFleetTransitRunsBeforeCurrentTurnColonyShipCompletion(t *testing.T) {
	rules := loadColonyShipRules(t)
	rules.MineralIndustryPerWorker["abundant"] = 600
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1704)
	empire := &state.Empires[0]
	addColonyShipTestTechnology(empire, ColonyShipTechnologyID, 120)
	colony := &state.Colonies[0]
	colony.Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectColonyShip, ProjectID: ColonyShipProjectID, ProgressPP: 499}

	oldFleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID: oldFleetID, EmpireID: empire.ID, Role: core.StrategicFleetRoleCivilian, SpecialKind: core.StrategicFleetSpecialColonyShip,
		DestinationSystemID: state.Galaxy.Systems[1].ID, RemainingTurns: 1, FTLSpeed: 2,
	})
	result, err := resolver.Resolve(ResolveContext{}, state, nil)
	if err != nil {
		t.Fatal(err)
	}
	arrivalIndex := eventIndex(result.Events, "empire.fleet_arrived")
	completionIndex := eventIndex(result.Events, "colony.colony_ship_completed")
	if arrivalIndex < 0 || completionIndex < 0 || arrivalIndex >= completionIndex {
		t.Fatalf("event order arrival=%d completion=%d events=%+v", arrivalIndex, completionIndex, result.Events)
	}
	if len(state.StrategicFleets) != 2 {
		t.Fatalf("strategic fleets=%+v", state.StrategicFleets)
	}
	if state.StrategicFleets[0].ID != oldFleetID || state.StrategicFleets[0].AtSystemID != state.Galaxy.Systems[1].ID {
		t.Fatalf("pre-existing Fleet did not arrive before Production: %+v", state.StrategicFleets[0])
	}
	newFleet := state.StrategicFleets[1]
	if newFleet.AtSystemID != state.Galaxy.Systems[0].ID || newFleet.DestinationSystemID != 0 || newFleet.RemainingTurns != 0 {
		t.Fatalf("newly completed Colony Ship incorrectly moved in same resolution: %+v", newFleet)
	}
}

func TestColonyShipFuelRangeUsesOriginalFuelCellTable(t *testing.T) {
	tests := []struct {
		name string
		tech []int
		want int
	}{
		{name: "none", want: 0},
		{name: "standard", tech: []int{167}, want: 4},
		{name: "deuterium", tech: []int{51}, want: 6},
		{name: "iridium", tech: []int{98}, want: 9},
		{name: "urridium", tech: []int{194}, want: 12},
		{name: "thorium", tech: []int{184}, want: 255},
		{name: "best known", tech: []int{167, 51, 98, 194}, want: 12},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			empire := core.Empire{}
			for _, technologyID := range test.tech {
				empire.KnownTechnologyIDs = appendUniqueSortedInt(empire.KnownTechnologyIDs, technologyID)
			}
			if got := colonyShipFuelRangeParsecs(empire); got != test.want {
				t.Fatalf("range=%d want=%d technologies=%v", got, test.want, empire.KnownTechnologyIDs)
			}
		})
	}
}

func TestMoveColonyShipChecksSupplyRangeAndAdvancesDeterministicETA(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1705)
	empire := &state.Empires[0]
	addColonyShipTestTechnology(empire, standardFuelCellsTechnologyID)
	source := &state.Galaxy.Systems[0]
	destination := &state.Galaxy.Systems[1]
	destination.X = source.X + 90
	destination.Y = source.Y
	fleetID := addColonyShipTestFleet(state, empire.ID, source.ID, 2)

	command, err := NewMoveFleetCommand(3, MoveFleetPayload{FleetID: fleetID, DestinationSystemID: destination.ID})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.moveFleet(state, empire.ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	fleet := &state.StrategicFleets[0]
	if fleet.AtSystemID != 0 || fleet.SourceSystemID != source.ID || fleet.DestinationSystemID != destination.ID || fleet.RemainingTurns != 2 || fleet.TransitTurnsTotal != 2 {
		t.Fatalf("started transit=%+v", *fleet)
	}
	projected := ProjectStrategicFleetTransitMetadata(*fleet, state)
	if projected.RouteDistanceParsecs != 3 || projected.RemainingDistanceParsecs != 3 {
		t.Fatalf("started transit projection=%+v", projected)
	}
	cancel, _ := NewMoveFleetCommand(4, MoveFleetPayload{FleetID: fleetID, DestinationSystemID: source.ID})
	before := *fleet
	if _, err := resolver.moveFleet(state, empire.ID, 1, cancel); err == nil {
		t.Fatal("started fleet movement was unexpectedly cancellable")
	}
	if !reflect.DeepEqual(*fleet, before) {
		t.Fatalf("rejected in-transit reroute mutated fleet: before=%+v after=%+v", before, *fleet)
	}
	var started FleetMovementStartedEvent
	if err := json.Unmarshal(event.Data, &started); err != nil {
		t.Fatal(err)
	}
	if started.RemainingTurns != 2 || started.FuelRangeParsecs != 4 || started.SupplyDistanceParsecs != 3 || started.FTLSpeed != 2 {
		t.Fatalf("movement started=%+v", started)
	}

	events, err := resolver.advanceStrategicFleetTransit(state)
	if err != nil {
		t.Fatal(err)
	}
	if fleet.SourceSystemID != source.ID || fleet.RemainingTurns != 1 || fleet.TransitTurnsTotal != 2 || findDomainEvent(events, "empire.fleet_movement_progressed") == nil {
		t.Fatalf("first transit advance fleet=%+v events=%+v", *fleet, events)
	}
	projected = ProjectStrategicFleetTransitMetadata(*fleet, state)
	if projected.RouteDistanceParsecs != 3 || projected.RemainingDistanceParsecs != 2 {
		t.Fatalf("progressed transit projection=%+v", projected)
	}
	events, err = resolver.advanceStrategicFleetTransit(state)
	if err != nil {
		t.Fatal(err)
	}
	if fleet.AtSystemID != destination.ID || fleet.SourceSystemID != 0 || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 || fleet.TransitTurnsTotal != 0 || findDomainEvent(events, "empire.fleet_arrived") == nil {
		t.Fatalf("arrival fleet=%+v events=%+v", *fleet, events)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("arrived state invalid: %v", err)
	}
}

func TestMoveColonyShipRejectsDestinationOutsideSupplyRange(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1706)
	empire := &state.Empires[0]
	addColonyShipTestTechnology(empire, standardFuelCellsTechnologyID)
	source := &state.Galaxy.Systems[0]
	destination := &state.Galaxy.Systems[1]
	destination.X = source.X + 180
	destination.Y = source.Y
	fleetID := addColonyShipTestFleet(state, empire.ID, source.ID, 2)
	command, _ := NewMoveFleetCommand(1, MoveFleetPayload{FleetID: fleetID, DestinationSystemID: destination.ID})
	if _, err := resolver.moveFleet(state, empire.ID, 1, command); err == nil {
		t.Fatal("Standard Fuel Cells unexpectedly reached a 6 pc destination")
	}
	if state.StrategicFleets[0].AtSystemID != source.ID {
		t.Fatalf("failed movement mutated fleet: %+v", state.StrategicFleets[0])
	}
	addColonyShipTestTechnology(empire, deuteriumFuelCellsTechnologyID)
	if _, err := resolver.moveFleet(state, empire.ID, 1, command); err != nil {
		t.Fatalf("Deuterium Fuel Cells should reach 6 pc destination: %v", err)
	}
}

func TestColonizePlanetCreatesSecondColonyAndConsumesColonyShip(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1707)
	empire := &state.Empires[0]
	targetSystem := &state.Galaxy.Systems[1]
	targetPlanet := &targetSystem.Planets[0]
	fleetID := addColonyShipTestFleet(state, empire.ID, targetSystem.ID, 2)
	expectedColonyID := state.NextID
	command, err := NewColonizePlanetCommand(9, ColonizePlanetPayload{FleetID: fleetID, PlanetID: targetPlanet.ID})
	if err != nil {
		t.Fatal(err)
	}
	events, err := resolver.colonizePlanet(state, empire.ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Kind != "empire.planet_colonized" || events[1].Kind != "empire.colony_ship_consumed" {
		t.Fatalf("colonization events=%+v", events)
	}
	if targetPlanet.ColonyID != expectedColonyID || len(state.Colonies) != 2 {
		t.Fatalf("target Planet/Colonies link=%d colonies=%+v", targetPlanet.ColonyID, state.Colonies)
	}
	newColony := state.Colonies[1]
	if newColony.ID != expectedColonyID || newColony.EmpireID != empire.ID || newColony.PlanetID != targetPlanet.ID {
		t.Fatalf("new Colony identity=%+v", newColony)
	}
	if newColony.Population.Total() != 1 || newColony.Population.Farmers() != 1 || newColony.Population.Workers() != 0 || newColony.Population.Scientists() != 0 {
		t.Fatalf("founding Population=%+v", newColony.Population)
	}
	if len(newColony.Buildings) != 0 {
		t.Fatalf("new Colony received invented free Buildings: %v", newColony.Buildings)
	}
	if len(state.StrategicFleets) != 0 {
		t.Fatalf("consumed Colony Ship still present: %+v", state.StrategicFleets)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("colonized state invalid: %v", err)
	}
}

func TestColonizePlanetRejectsInvalidFleetAndPlanetStates(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("wrong owner", func(t *testing.T) {
		state := core.NewSmallFixture(1710)
		secondEmpireID := state.NewID()
		state.Empires = append(state.Empires, core.Empire{ID: secondEmpireID, Name: "Other", RaceID: "human"})
		fleetID := addColonyShipTestFleet(state, state.Empires[0].ID, state.Galaxy.Systems[1].ID, 2)
		command, _ := NewColonizePlanetCommand(1, ColonizePlanetPayload{FleetID: fleetID, PlanetID: state.Galaxy.Systems[1].Planets[0].ID})
		if _, err := resolver.colonizePlanet(state, secondEmpireID, 2, command); err == nil {
			t.Fatal("foreign Empire unexpectedly colonized with another Empire's Colony Ship")
		}
	})

	t.Run("ordinary combat fleet", func(t *testing.T) {
		state := core.NewSmallFixture(1711)
		fleetID := state.NewID()
		state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{ID: fleetID, EmpireID: state.Empires[0].ID, Role: core.StrategicFleetRoleCombat, AtSystemID: state.Galaxy.Systems[1].ID})
		command, _ := NewColonizePlanetCommand(1, ColonizePlanetPayload{FleetID: fleetID, PlanetID: state.Galaxy.Systems[1].Planets[0].ID})
		if _, err := resolver.colonizePlanet(state, state.Empires[0].ID, 1, command); err == nil {
			t.Fatal("combat Fleet unexpectedly colonized a Planet")
		}
	})

	t.Run("moving Colony Ship", func(t *testing.T) {
		state := core.NewSmallFixture(1712)
		fleetID := state.NewID()
		state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
			ID: fleetID, EmpireID: state.Empires[0].ID, Role: core.StrategicFleetRoleCivilian, SpecialKind: core.StrategicFleetSpecialColonyShip,
			DestinationSystemID: state.Galaxy.Systems[1].ID, RemainingTurns: 1, FTLSpeed: 2,
		})
		command, _ := NewColonizePlanetCommand(1, ColonizePlanetPayload{FleetID: fleetID, PlanetID: state.Galaxy.Systems[1].Planets[0].ID})
		if _, err := resolver.colonizePlanet(state, state.Empires[0].ID, 1, command); err == nil {
			t.Fatal("moving Colony Ship unexpectedly colonized before arrival")
		}
	})

	t.Run("different system", func(t *testing.T) {
		state := core.NewSmallFixture(1713)
		fleetID := addColonyShipTestFleet(state, state.Empires[0].ID, state.Galaxy.Systems[0].ID, 2)
		command, _ := NewColonizePlanetCommand(1, ColonizePlanetPayload{FleetID: fleetID, PlanetID: state.Galaxy.Systems[1].Planets[0].ID})
		if _, err := resolver.colonizePlanet(state, state.Empires[0].ID, 1, command); err == nil {
			t.Fatal("Colony Ship unexpectedly colonized Planet in another System")
		}
	})

	t.Run("already colonized", func(t *testing.T) {
		state := core.NewSmallFixture(1714)
		fleetID := addColonyShipTestFleet(state, state.Empires[0].ID, state.Galaxy.Systems[0].ID, 2)
		command, _ := NewColonizePlanetCommand(1, ColonizePlanetPayload{FleetID: fleetID, PlanetID: state.Galaxy.Systems[0].Planets[0].ID})
		if _, err := resolver.colonizePlanet(state, state.Empires[0].ID, 1, command); err == nil {
			t.Fatal("Colony Ship unexpectedly colonized occupied Planet")
		}
	})
}

func TestColonyShipHeadlessVerticalLoopBuildMoveArriveColonize(t *testing.T) {
	rules := loadColonyShipRules(t)
	rules.MineralIndustryPerWorker["abundant"] = 600
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1720)
	empire := &state.Empires[0]
	addColonyShipTestTechnology(empire, ColonyShipTechnologyID, 120, standardFuelCellsTechnologyID)
	home := &state.Galaxy.Systems[0]
	target := &state.Galaxy.Systems[1]
	target.X = home.X + 90
	target.Y = home.Y
	targetPlanetID := target.Planets[0].ID
	ctx := ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: empire.ID}}}

	queue, err := NewQueueColonyShipCommand(1, QueueColonyShipPayload{ColonyID: state.Colonies[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	result, err := resolver.Resolve(ctx, state, []protocol.CommandBatch{{SeatID: 1, Commands: []protocol.Command{queue}}})
	if err != nil {
		t.Fatal(err)
	}
	if findDomainEvent(result.Events, "colony.colony_ship_completed") == nil || len(state.StrategicFleets) != 1 {
		t.Fatalf("build step state=%+v events=%+v", state.StrategicFleets, result.Events)
	}
	fleetID := state.StrategicFleets[0].ID
	if state.StrategicFleets[0].AtSystemID != home.ID {
		t.Fatalf("built Colony Ship=%+v", state.StrategicFleets[0])
	}

	move, err := NewMoveFleetCommand(2, MoveFleetPayload{FleetID: fleetID, DestinationSystemID: target.ID})
	if err != nil {
		t.Fatal(err)
	}
	result, err = resolver.Resolve(ctx, state, []protocol.CommandBatch{{SeatID: 1, Commands: []protocol.Command{move}}})
	if err != nil {
		t.Fatal(err)
	}
	if findDomainEvent(result.Events, "empire.fleet_movement_started") == nil || state.StrategicFleets[0].RemainingTurns != 1 {
		t.Fatalf("move step fleet=%+v events=%+v", state.StrategicFleets[0], result.Events)
	}

	result, err = resolver.Resolve(ctx, state, nil)
	if err != nil {
		t.Fatal(err)
	}
	if findDomainEvent(result.Events, "empire.fleet_arrived") == nil || state.StrategicFleets[0].AtSystemID != target.ID {
		t.Fatalf("arrival step fleet=%+v events=%+v", state.StrategicFleets[0], result.Events)
	}

	colonize, err := NewColonizePlanetCommand(3, ColonizePlanetPayload{FleetID: fleetID, PlanetID: targetPlanetID})
	if err != nil {
		t.Fatal(err)
	}
	result, err = resolver.Resolve(ctx, state, []protocol.CommandBatch{{SeatID: 1, Commands: []protocol.Command{colonize}}})
	if err != nil {
		t.Fatal(err)
	}
	if findDomainEvent(result.Events, "empire.planet_colonized") == nil || findDomainEvent(result.Events, "empire.colony_ship_consumed") == nil {
		t.Fatalf("colonization events=%+v", result.Events)
	}
	if len(state.Colonies) != 2 || state.Galaxy.Systems[1].Planets[0].ColonyID == 0 || len(state.StrategicFleets) != 0 {
		t.Fatalf("vertical-loop final colonies=%+v fleets=%+v target=%+v", state.Colonies, state.StrategicFleets, state.Galaxy.Systems[1].Planets[0])
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("vertical-loop final state invalid: %v", err)
	}
	encoded, err := core.MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := core.UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Colonies) != 2 || loaded.Galaxy.Systems[1].Planets[0].ColonyID != state.Galaxy.Systems[1].Planets[0].ColonyID || len(loaded.StrategicFleets) != 0 {
		t.Fatalf("vertical-loop roundtrip state=%+v", loaded)
	}
}
