package game

import (
	"testing"

	"moox/internal/core"
)

func troopTransportChoice(choices []ConstructionChoice) *ConstructionChoice {
	for i := range choices {
		if choices[i].ProjectKind == core.ConstructionProjectTroopTransport {
			return &choices[i]
		}
	}
	return nil
}

func TestTroopTransportConstructionIsBaselineBuildableAndUsesOriginalCost(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	state := core.NewSmallFixture(0xB410)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]
	empire.KnownTechnologyIDs = nil
	choices, err := rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	choice := troopTransportChoice(choices)
	if choice == nil || choice.ProjectID != TroopTransportProjectID || choice.TechnologyID != 0 || choice.ProductionCostPP != 100 {
		t.Fatalf("ordinary Troop Transport choice=%+v", choice)
	}
	modifiers := rules.RaceModifiers[empire.RaceID]
	modifiers.GovernmentTraitID = "government_feudal"
	rules.RaceModifiers[empire.RaceID] = modifiers
	choices, err = rules.AvailableConstructionChoices(state, empire.ID, colony.ID)
	if err != nil {
		t.Fatal(err)
	}
	choice = troopTransportChoice(choices)
	if choice == nil || choice.ProductionCostPP != 67 {
		t.Fatalf("Feudal Troop Transport choice=%+v want cost 67", choice)
	}
}

func TestTroopTransportQueueCompletionMovementAndCommandPointIdentity(t *testing.T) {
	rules, resolver := commandPointTestResolver(t)
	state := core.NewSmallFixture(0xB411)
	empire := &state.Empires[0]
	addColonyShipTestTechnology(empire, 120, standardFuelCellsTechnologyID)
	colony := &state.Colonies[0]
	command, err := NewQueueTroopTransportCommand(1, QueueTroopTransportPayload{ColonyID: colony.ID})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.queueTroopTransport(state, empire.ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != "colony.troop_transport_queued" || colony.Construction == nil || colony.Construction.ProjectKind != core.ConstructionProjectTroopTransport || colony.Construction.ProjectID != TroopTransportProjectID {
		t.Fatalf("queued transport event/state=%+v/%+v", event, colony.Construction)
	}
	colony.PopulationDynamics.ProductionAvailable = 100
	events, err := resolver.advanceConstruction(state)
	if err != nil {
		t.Fatal(err)
	}
	if colony.Construction != nil {
		t.Fatalf("completed transport retained construction=%+v", colony.Construction)
	}
	if eventIndex(events, "colony.troop_transport_completed") < 0 {
		t.Fatalf("completion events=%+v", events)
	}
	if len(state.StrategicFleets) != 1 {
		t.Fatalf("fleets=%+v", state.StrategicFleets)
	}
	fleet := &state.StrategicFleets[0]
	if fleet.Role != core.StrategicFleetRoleCivilian || fleet.SpecialKind != core.StrategicFleetSpecialTroopTransport || fleet.AtSystemID != state.Galaxy.Systems[0].ID || fleet.FTLSpeed != 2 || len(fleet.ShipIDs) != 0 {
		t.Fatalf("completed transport=%+v", *fleet)
	}
	points, err := rules.deriveEmpireCommandPoints(state, empire)
	if err != nil {
		t.Fatal(err)
	}
	if points.Used != 1 {
		t.Fatalf("transport command point use=%d want 1", points.Used)
	}
	move, err := NewMoveFleetCommand(1, MoveFleetPayload{FleetID: fleet.ID, DestinationSystemID: state.Galaxy.Systems[1].ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.moveFleet(state, empire.ID, 1, move); err != nil {
		t.Fatal(err)
	}
	if fleet.AtSystemID != 0 || fleet.DestinationSystemID != state.Galaxy.Systems[1].ID || fleet.RemainingTurns <= 0 {
		t.Fatalf("moving transport=%+v", *fleet)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("transport state invalid: %v", err)
	}
}
