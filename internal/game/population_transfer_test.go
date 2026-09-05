package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func TestPopulationTransferReservesFiveFreightersBeforeFoodAndReleasesOnArrival(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state, home, destination := twoColonyFoodFixture(t, 910)
	empire := &state.Empires[0]
	empire.Freighters = 5
	empire.KnownTechnologyIDs = []int{120} // Nuclear Drive: original FTL speed 2.

	command, err := NewTransferPopulationCommand(1, TransferPopulationPayload{
		SourceColonyID: home.ID, DestinationColonyID: destination.ID, Job: core.PopulationJobFarmer,
	})
	if err != nil {
		t.Fatal(err)
	}
	started, err := resolver.transferPopulation(state, empire.ID, protocol.SeatID(1), command)
	if err != nil {
		t.Fatal(err)
	}
	if started.Kind != "empire.population_transfer_started" || len(state.PopulationTransfers) != 1 {
		t.Fatalf("unexpected start event/state: kind=%s transfers=%+v", started.Kind, state.PopulationTransfers)
	}
	if state.PopulationTransfers[0].RemainingTurns != 1 {
		t.Fatalf("ETA=%d want=1 for Alpha->Beta at Nuclear Drive speed", state.PopulationTransfers[0].RemainingTurns)
	}
	if home.Population.Total() != 3 || home.Population.Farmers() != 2 {
		t.Fatalf("source population after launch=%+v", home.Population)
	}

	for i := range state.Colonies {
		if err := resolver.recalculateColony(state, &state.Colonies[i]); err != nil {
			t.Fatal(err)
		}
	}
	foodEvents, err := resolver.materializeFoodLogistics(state, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(foodEvents) == 0 {
		t.Fatal("expected food logistics event")
	}
	var food FoodLogisticsResolvedEvent
	if err := json.Unmarshal(foodEvents[0].Data, &food); err != nil {
		t.Fatal(err)
	}
	if food.Snapshot.PopulationTransportFreightersReserved != 5 || food.Snapshot.FreightersAvailableForFood != 0 || food.Snapshot.FreightersUsed != 0 {
		t.Fatalf("food Freighter reservation snapshot=%+v", food.Snapshot)
	}
	if food.Snapshot.FreighterOperatingCostBC != 2 {
		t.Fatalf("reserved Freighter operating cost=%v want=2 BC", food.Snapshot.FreighterOperatingCostBC)
	}

	transferEvents, err := resolver.advancePopulationTransfers(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(state.PopulationTransfers) != 0 {
		t.Fatalf("arrived transfer still active: %+v", state.PopulationTransfers)
	}
	if findDomainEvent(transferEvents, "empire.population_transfer_arrived") == nil {
		t.Fatalf("missing arrival event: %+v", transferEvents)
	}
	if destination.Population.Total() != 3 || destination.Population.Farmers() != 1 {
		t.Fatalf("destination population after arrival=%+v", destination.Population)
	}
	if _, err := resolver.materializeFoodLogistics(state, false); err != nil {
		t.Fatal(err)
	}
	if empire.FoodLogistics.PopulationTransportFreightersReserved != 0 || empire.FoodLogistics.FreightersAvailableForFood != 5 {
		t.Fatalf("Freighters not released after arrival: %+v", empire.FoodLogistics)
	}
}

func TestPopulationTransferReservationAndFoodUseShareFreighterMaintenanceBucket(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state, home, destination := twoColonyFoodFixture(t, 0xA300)
	empire := &state.Empires[0]
	empire.Freighters = 6
	empire.KnownTechnologyIDs = []int{120} // Nuclear Drive: original FTL speed 2.

	command, err := NewTransferPopulationCommand(1, TransferPopulationPayload{
		SourceColonyID: home.ID, DestinationColonyID: destination.ID, Job: core.PopulationJobFarmer,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.transferPopulation(state, empire.ID, protocol.SeatID(1), command); err != nil {
		t.Fatal(err)
	}
	for i := range state.Colonies {
		if err := resolver.recalculateColony(state, &state.Colonies[i]); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := resolver.materializeFoodLogistics(state, false); err != nil {
		t.Fatal(err)
	}

	snapshot := empire.FoodLogistics
	if snapshot.PopulationTransportFreightersReserved != 5 || snapshot.FreightersAvailableForFood != 1 || snapshot.FreightersUsed != 1 {
		t.Fatalf("combined Freighter activity snapshot=%+v", snapshot)
	}
	if snapshot.FreighterOperatingCostBC != 3 {
		t.Fatalf("combined Freighter operating cost=%v want=3 BC snapshot=%+v", snapshot.FreighterOperatingCostBC, snapshot)
	}

	empire.Treasury.BalanceBC = 50
	if _, err := resolver.settleTreasury(state); err != nil {
		t.Fatal(err)
	}
	if empire.Treasury.FreighterOperatingCostBC != 3 {
		t.Fatalf("Treasury FreighterOperatingCostBC=%v want=3 BC Treasury=%+v", empire.Treasury.FreighterOperatingCostBC, empire.Treasury)
	}
}
func TestPopulationTransferWithinSystemIsImmediateAndNeedsNoFreighters(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, _ := NewEconomyResolver(rules)
	state := core.NewSmallFixture(911)
	home := &state.Colonies[0]
	system := &state.Galaxy.Systems[0]
	planet := core.Planet{ID: state.NewID(), Name: "Alpha II", Orbit: 2, SizeID: "medium", MineralID: "abundant", GravityID: "normal_g", ClimateID: "terran"}
	destination := core.Colony{ID: state.NewID(), EmpireID: state.Empires[0].ID, PlanetID: planet.ID, Population: core.NewAssimilatedPopulation(state.Empires[0].ID, 0, 2, 0)}
	planet.ColonyID = destination.ID
	system.Planets = append(system.Planets, planet)
	state.Colonies = append(state.Colonies, destination)

	command, err := NewTransferPopulationCommand(1, TransferPopulationPayload{SourceColonyID: home.ID, DestinationColonyID: destination.ID, Job: core.PopulationJobFarmer})
	if err != nil {
		t.Fatal(err)
	}
	event, err := resolver.transferPopulation(state, state.Empires[0].ID, 1, command)
	if err != nil {
		t.Fatal(err)
	}
	if event.Kind != "colony.population_transferred" || len(state.PopulationTransfers) != 0 {
		t.Fatalf("same-system transfer kind=%s active=%+v", event.Kind, state.PopulationTransfers)
	}
	if state.Colonies[0].Population.Farmers() != 1 || state.Colonies[1].Population.Farmers() != 1 {
		t.Fatalf("same-system job preservation failed source=%+v destination=%+v", state.Colonies[0].Population, state.Colonies[1].Population)
	}
}

func TestPopulationTransferIsLostWhenDestinationIsBlockadedAtArrival(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, _ := NewEconomyResolver(rules)
	state, home, destination := twoColonyFoodFixture(t, 912)
	empireID := state.Empires[0].ID
	state.Empires[0].Freighters = 5
	state.Empires[0].KnownTechnologyIDs = []int{120}
	state.Galaxy.Systems[1].BlockadedEmpireIDs = []core.ID{empireID}
	beforeDestination := destination.Population

	command, _ := NewTransferPopulationCommand(1, TransferPopulationPayload{SourceColonyID: home.ID, DestinationColonyID: destination.ID, Job: core.PopulationJobFarmer})
	if _, err := resolver.transferPopulation(state, empireID, 1, command); err != nil {
		t.Fatal(err)
	}
	events, err := resolver.advancePopulationTransfers(state)
	if err != nil {
		t.Fatal(err)
	}
	lost := findDomainEvent(events, "empire.population_transfer_lost")
	if lost == nil {
		t.Fatalf("missing lost event: %+v", events)
	}
	var payload PopulationTransferLostEvent
	if err := json.Unmarshal(lost.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Reason != "destination_blockaded" || payload.FreightersReleased != 5 {
		t.Fatalf("lost payload=%+v", payload)
	}
	if !reflect.DeepEqual(destination.Population, beforeDestination) || len(state.PopulationTransfers) != 0 {
		t.Fatalf("blockaded arrival changed destination or retained transfer: pop=%+v transfers=%+v", destination.Population, state.PopulationTransfers)
	}
}

func TestPopulationTransferCapacityCountsInboundSettlers(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, _ := NewEconomyResolver(rules)
	state, home, destination := twoColonyFoodFixture(t, 913)
	empireID := state.Empires[0].ID
	state.Empires[0].Freighters = 10
	if err := resolver.recalculateColony(state, destination); err != nil {
		t.Fatal(err)
	}
	capacity := destination.PopulationDynamics.Capacity
	if capacity < 2 {
		t.Fatalf("unexpected destination capacity %g", capacity)
	}
	destination.Population = core.NewAssimilatedPopulation(empireID, 0, capacity-1, 0)
	state.PopulationTransfers = append(state.PopulationTransfers, core.PopulationTransfer{
		ID: state.NewID(), EmpireID: empireID, SourceColonyID: home.ID, DestinationColonyID: destination.ID, OriginEmpireID: empireID, LoyaltyEmpireID: empireID, AssimilationState: core.PopulationAssimilated, Job: core.PopulationJobWorker, RemainingTurns: 2,
	})
	command, _ := NewTransferPopulationCommand(1, TransferPopulationPayload{SourceColonyID: home.ID, DestinationColonyID: destination.ID, Job: core.PopulationJobFarmer})
	if _, err := resolver.transferPopulation(state, empireID, 1, command); err == nil {
		t.Fatal("expected inbound Settler reservation to make destination capacity unavailable")
	}
}

func TestPopulationTransferOriginalDriveAndTraitETA(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	if got := rules.populationTransferFTLSpeed(core.Empire{RaceID: "human", KnownTechnologyIDs: []int{120}}); got != 2 {
		t.Fatalf("Nuclear Drive speed=%d want=2", got)
	}
	if got := rules.populationTransferFTLSpeed(core.Empire{RaceID: "human", KnownTechnologyIDs: []int{120, 72, 96, 11, 88, 95}}); got != 7 {
		t.Fatalf("Interphased Drive speed=%d want=7", got)
	}
	if got := rules.populationTransferFTLSpeed(core.Empire{RaceID: "trilarian", KnownTechnologyIDs: []int{120}}); got != 4 {
		t.Fatalf("Trans Dimensional Nuclear Drive speed=%d want=4", got)
	}
	source := core.StarSystem{X: 0, Y: 0}
	destination := core.StarSystem{X: 90, Y: 0}
	if got := populationTransferETA(source, destination, 2); got != 2 {
		t.Fatalf("3 parsec ETA at speed 2=%d want=2", got)
	}
	if got := populationTransferETA(source, core.StarSystem{X: 10000, Y: 0}, 2); got != 15 {
		t.Fatalf("long-range ETA=%d want original cap 15", got)
	}
}

func TestPopulationTransferCanChangeDestinationJob(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state, home, destination := twoColonyFoodFixture(t, 916)
	empire := &state.Empires[0]
	empire.Freighters = 5
	empire.KnownTechnologyIDs = []int{120}
	beforeScientists := destination.Population.Scientists()
	beforeFarmers := destination.Population.Farmers()

	choices, err := resolver.AvailablePopulationTransferChoices(state, empire.ID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, choice := range choices {
		if choice.SourceColonyID == home.ID && choice.DestinationColonyID == destination.ID && choice.SourceJob == core.PopulationJobFarmer && choice.DestinationJob == core.PopulationJobScientist {
			if choice.SameSystem || choice.ETA != 1 || choice.FreightersRequired != 5 {
				t.Fatalf("transfer choice metadata=%+v", choice)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no Farmer->Scientist transfer choice in %+v", choices)
	}

	command, err := NewTransferPopulationCommand(1, TransferPopulationPayload{
		SourceColonyID: home.ID, DestinationColonyID: destination.ID,
		SourceJob: core.PopulationJobFarmer, DestinationJob: core.PopulationJobScientist,
	})
	if err != nil {
		t.Fatal(err)
	}
	started, err := resolver.transferPopulation(state, empire.ID, protocol.SeatID(1), command)
	if err != nil {
		t.Fatal(err)
	}
	if started.Kind != "empire.population_transfer_started" || len(state.PopulationTransfers) != 1 {
		t.Fatalf("start event/state=%+v/%+v", started, state.PopulationTransfers)
	}
	transfer := state.PopulationTransfers[0]
	if transfer.SourcePopulationJob() != core.PopulationJobFarmer || transfer.DestinationPopulationJob() != core.PopulationJobScientist {
		t.Fatalf("transfer jobs=%+v", transfer)
	}
	events, err := resolver.advancePopulationTransfers(state)
	if err != nil {
		t.Fatal(err)
	}
	if eventIndex(events, "empire.population_transfer_arrived") < 0 {
		t.Fatalf("arrival events=%+v", events)
	}
	if got := destination.Population.Scientists(); got != beforeScientists+1 {
		t.Fatalf("destination scientists=%v want=%v", got, beforeScientists+1)
	}
	if got := destination.Population.Farmers(); got != beforeFarmers {
		t.Fatalf("destination farmers=%v want unchanged %v", got, beforeFarmers)
	}
}
