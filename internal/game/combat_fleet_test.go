package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func combatFleetTestSpec() core.ShipDesignSpec {
	return core.ShipDesignSpec{
		HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "nuclear_drive", FTLSpeed: 2,
		ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4,
		HullBaseCostPP: 20, HullSpace: 25, BaseDesignCostPP: 25, ProductionCostPP: 25,
	}
}

func addCombatFleetTestFleet(state *core.GameState, empireID, systemID core.ID, shipCount int) (core.ID, []core.ID) {
	spec := combatFleetTestSpec()
	designID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{ID: designID, EmpireID: empireID, Revision: 1, Name: "Combat test", Spec: spec})
	shipIDs := make([]core.ID, 0, shipCount)
	for i := 0; i < shipCount; i++ {
		shipID := state.NewID()
		state.Ships = append(state.Ships, core.Ship{ID: shipID, EmpireID: empireID, SourceDesignID: designID, SourceDesignRevision: 1, Name: "Combat test", Spec: spec})
		shipIDs = append(shipIDs, shipID)
	}
	fleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID: fleetID, EmpireID: empireID, Role: core.StrategicFleetRoleCombat, SpecialKind: core.StrategicFleetSpecialNone,
		AtSystemID: systemID, ShipIDs: append([]core.ID(nil), shipIDs...),
	})
	return fleetID, shipIDs
}

func setCombatFleetTestTech(empire *core.Empire) {
	empire.KnownTechnologyIDs = []int{51, 72, 120, 167} // Deuterium, Fusion, Nuclear, Standard Fuel Cells.
}

func TestCombatFleetMovementProfileUsesCurrentEmpireTechAndTransDimensional(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(1901)
	empire := &state.Empires[0]
	setCombatFleetTestTech(empire)
	fleetID, _ := addCombatFleetTestFleet(state, empire.ID, state.Galaxy.Systems[0].ID, 2)
	_, fleet := strategicFleetByID(state, fleetID)
	if fleet == nil {
		t.Fatal("combat fleet missing")
	}
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	speed, fuelRange, err := resolver.combatFleetMovementProfile(state, *fleet, empire)
	if err != nil {
		t.Fatal(err)
	}
	if speed != 3 || fuelRange != 6 {
		t.Fatalf("movement profile speed=%d range=%d want=3/6; build snapshots must not pin Nuclear/Standard values", speed, fuelRange)
	}
	mods := rules.RaceModifiers[empire.RaceID]
	mods.TransDimensional = true
	rules.RaceModifiers[empire.RaceID] = mods
	speed, fuelRange, err = resolver.combatFleetMovementProfile(state, *fleet, empire)
	if err != nil {
		t.Fatal(err)
	}
	if speed != 5 || fuelRange != 6 {
		t.Fatalf("Trans-Dimensional movement profile speed=%d range=%d want=5/6", speed, fuelRange)
	}
}

func TestCombatFleetSubsetMoveSplitsBeforeMovementAndUsesDerivedProfile(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1902)
	empire := &state.Empires[0]
	setCombatFleetTestTech(empire)
	source := &state.Galaxy.Systems[0]
	destination := &state.Galaxy.Systems[1]
	destination.X = source.X + 150 // 5 pc.
	destination.Y = source.Y
	fleetID, shipIDs := addCombatFleetTestFleet(state, empire.ID, source.ID, 2)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	nextID := state.NextID
	command, err := NewMoveFleetCommand(7, MoveFleetPayload{FleetID: fleetID, DestinationSystemID: destination.ID, ShipIDs: []core.ID{shipIDs[1]}})
	if err != nil {
		t.Fatal(err)
	}
	events, err := resolver.moveFleetEvents(state, empire.ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Kind != "empire.fleet_split" || events[1].Kind != "empire.fleet_movement_started" {
		t.Fatalf("subset movement event order=%v", []string{events[0].Kind, events[1].Kind})
	}
	var split FleetSplitEvent
	if err := json.Unmarshal(events[0].Data, &split); err != nil {
		t.Fatal(err)
	}
	if split.NewFleetID != nextID || !reflect.DeepEqual(split.MovedShipIDs, []core.ID{shipIDs[1]}) {
		t.Fatalf("split event=%+v nextID=%d", split, nextID)
	}
	var started FleetMovementStartedEvent
	if err := json.Unmarshal(events[1].Data, &started); err != nil {
		t.Fatal(err)
	}
	if started.FleetID != nextID || started.RemainingTurns != 2 || started.FTLSpeed != 3 || started.FuelRangeParsecs != 6 || started.SupplyDistanceParsecs != 5 {
		t.Fatalf("movement event=%+v", started)
	}
	_, sourceFleet := strategicFleetByID(state, fleetID)
	_, movingFleet := strategicFleetByID(state, nextID)
	if sourceFleet == nil || movingFleet == nil {
		t.Fatalf("source/moving fleet missing: source=%+v moving=%+v", sourceFleet, movingFleet)
	}
	if !reflect.DeepEqual(sourceFleet.ShipIDs, []core.ID{shipIDs[0]}) || sourceFleet.AtSystemID != source.ID {
		t.Fatalf("source fleet after subset move=%+v", *sourceFleet)
	}
	if !reflect.DeepEqual(movingFleet.ShipIDs, []core.ID{shipIDs[1]}) || movingFleet.AtSystemID != 0 || movingFleet.SourceSystemID != source.ID || movingFleet.DestinationSystemID != destination.ID || movingFleet.RemainingTurns != 2 || movingFleet.TransitTurnsTotal != 2 || movingFleet.FTLSpeed != 0 {
		t.Fatalf("moving split fleet=%+v", *movingFleet)
	}
	beforeTransitReroute := *movingFleet
	reroute, _ := NewMoveFleetCommand(8, MoveFleetPayload{FleetID: movingFleet.ID, DestinationSystemID: source.ID})
	if _, err := resolver.moveFleetEvents(state, empire.ID, 1, reroute); err == nil {
		t.Fatal("started combat fleet movement was unexpectedly cancellable")
	}
	_, afterTransitReroute := strategicFleetByID(state, movingFleet.ID)
	if afterTransitReroute == nil || !reflect.DeepEqual(*afterTransitReroute, beforeTransitReroute) {
		t.Fatalf("rejected combat reroute mutated transit: before=%+v after=%+v", beforeTransitReroute, afterTransitReroute)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("schema23 subset transit invalid: %v", err)
	}
}

func TestCombatFleetRejectedSubsetMoveIsAtomicAndDoesNotConsumeNextID(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(1903)
	empire := &state.Empires[0]
	setCombatFleetTestTech(empire)
	source := &state.Galaxy.Systems[0]
	destination := &state.Galaxy.Systems[1]
	destination.X = source.X + 210 // 7 pc, beyond Deuterium 6 pc range.
	destination.Y = source.Y
	fleetID, shipIDs := addCombatFleetTestFleet(state, empire.ID, source.ID, 2)
	beforeNextID := state.NextID
	beforeFleets := append([]core.StrategicFleet(nil), state.StrategicFleets...)
	beforeFleets[0].ShipIDs = append([]core.ID(nil), beforeFleets[0].ShipIDs...)
	command, err := NewMoveFleetCommand(8, MoveFleetPayload{FleetID: fleetID, DestinationSystemID: destination.ID, ShipIDs: []core.ID{shipIDs[1]}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.moveFleetEvents(state, empire.ID, 1, command); err == nil {
		t.Fatal("expected out-of-range subset move to fail")
	}
	if state.NextID != beforeNextID || !reflect.DeepEqual(state.StrategicFleets, beforeFleets) {
		t.Fatalf("rejected subset move mutated state: next=%d/%d fleets=%+v want=%+v", state.NextID, beforeNextID, state.StrategicFleets, beforeFleets)
	}
}

func TestCombatFleetExplicitSplitAndMergePreserveComposition(t *testing.T) {
	state := core.NewSmallFixture(1904)
	empire := &state.Empires[0]
	fleetID, shipIDs := addCombatFleetTestFleet(state, empire.ID, state.Galaxy.Systems[0].ID, 3)
	resolver := &EconomyResolver{}
	nextID := state.NextID
	splitCommand, err := NewSplitFleetCommand(1, SplitFleetPayload{FleetID: fleetID, ShipIDs: []core.ID{shipIDs[2]}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.splitFleet(state, empire.ID, 1, splitCommand); err != nil {
		t.Fatal(err)
	}
	_, source := strategicFleetByID(state, fleetID)
	_, split := strategicFleetByID(state, nextID)
	if source == nil || split == nil || !reflect.DeepEqual(source.ShipIDs, shipIDs[:2]) || !reflect.DeepEqual(split.ShipIDs, []core.ID{shipIDs[2]}) {
		t.Fatalf("split composition source=%+v new=%+v", source, split)
	}
	mergeCommand, err := NewMergeFleetsCommand(2, MergeFleetsPayload{TargetFleetID: fleetID, SourceFleetID: nextID})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.mergeFleets(state, empire.ID, 1, mergeCommand)
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != "empire.fleets_merged" {
		t.Fatalf("merge event kind=%q", event.Kind)
	}
	_, merged := strategicFleetByID(state, fleetID)
	_, removed := strategicFleetByID(state, nextID)
	if merged == nil || removed != nil || !reflect.DeepEqual(merged.ShipIDs, shipIDs) {
		t.Fatalf("merged composition=%+v removed=%+v", merged, removed)
	}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestCombatFleetArrivalDrivesSameResolutionBlockade(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	fixture := newBlockadeTestFixture(t, 1905)
	state := fixture.state
	empire := &state.Empires[0]
	setCombatFleetTestTech(empire)
	source := &state.Galaxy.Systems[0]
	destination := &state.Galaxy.Systems[1]
	destination.X = source.X + 30 // 1 pc, ETA 1 with Fusion.
	destination.Y = source.Y
	fleetID, _ := addCombatFleetTestFleet(state, empire.ID, source.ID, 1)
	state.DiplomaticRelations = reciprocalWarRelations(empire.ID, fixture.targetEmpireID)
	command, err := NewMoveFleetCommand(1, MoveFleetPayload{FleetID: fleetID, DestinationSystemID: destination.ID})
	if err != nil {
		t.Fatal(err)
	}
	ctx := ResolveContext{Seats: []SeatAuthority{{SeatID: 1, EmpireID: empire.ID}}}
	result, err := resolver.Resolve(ctx, state, []protocol.CommandBatch{{SeatID: 1, Commands: []protocol.Command{command}}})
	if err != nil {
		t.Fatal(err)
	}
	if eventIndex(result.Events, "empire.fleet_movement_started") < 0 || eventIndex(result.Events, "empire.fleet_arrived") < 0 || eventIndex(result.Events, "empire.fleet_movement_started") >= eventIndex(result.Events, "empire.fleet_arrived") {
		t.Fatalf("movement/arrival events=%+v", result.Events)
	}
	_, arrived := strategicFleetByID(result.State, fleetID)
	if arrived == nil || arrived.AtSystemID != destination.ID || arrived.DestinationSystemID != 0 || arrived.RemainingTurns != 0 {
		t.Fatalf("arrived combat fleet=%+v", arrived)
	}
	if !reflect.DeepEqual(result.State.Galaxy.Systems[1].BlockadedEmpireIDs, []core.ID{fixture.targetEmpireID}) {
		t.Fatalf("arrival blockade=%v want target empire %d", result.State.Galaxy.Systems[1].BlockadedEmpireIDs, fixture.targetEmpireID)
	}
	if !result.State.Empires[0].HasVisitedSystem(destination.ID) {
		t.Fatalf("arrival did not persist visited system %d: %v", destination.ID, result.State.Empires[0].VisitedSystemIDs)
	}
}
