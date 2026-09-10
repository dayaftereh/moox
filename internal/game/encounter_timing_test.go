package game

import (
	"encoding/json"
	"testing"

	"moox/internal/core"
)

func TestHostileEncounterIncludesSameTurnCombatFleetArrivalBeforeBlockade(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 2210)
	state := fixture.state
	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	attackerFleetID, _ := addCombatFleetTestFleet(state, fixture.blockaderEmpireID, state.Galaxy.Systems[0].ID, 1)
	_, attackerFleet := strategicFleetByID(state, attackerFleetID)
	attackerFleet.AtSystemID = 0
	attackerFleet.DestinationSystemID = fixture.targetSystemID
	attackerFleet.RemainingTurns = 1
	attackerFleet.FTLSpeed = 0
	addCombatFleetTestFleet(state, fixture.targetEmpireID, fixture.targetSystemID, 1)
	state.DiplomaticRelations = reciprocalWarRelations(fixture.blockaderEmpireID, fixture.targetEmpireID)
	resolver, err := NewEconomyResolver(loadCommittedEconomyRules(t))
	if err != nil {
		t.Fatal(err)
	}
	resolution, err := resolver.Resolve(encounterTestContext(state), state, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolution.Encounters) != 1 {
		t.Fatalf("arrival encounters=%+v", resolution.Encounters)
	}
	if resolution.Encounters[0].SystemID != fixture.targetSystemID || !containsCoreID(resolution.Encounters[0].Attacker.CombatFleetIDs, attackerFleetID) {
		t.Fatalf("arrival encounter=%+v", resolution.Encounters[0])
	}
	_, arrivedFleet := strategicFleetByID(state, attackerFleetID)
	if arrivedFleet == nil || arrivedFleet.AtSystemID != fixture.targetSystemID || arrivedFleet.DestinationSystemID != 0 || arrivedFleet.RemainingTurns != 0 {
		t.Fatalf("arrival not materialized before encounter: %+v", arrivedFleet)
	}
	if len(state.Galaxy.Systems[1].BlockadedEmpireIDs) != 0 {
		t.Fatalf("blockade materialized before encounter result: %v", state.Galaxy.Systems[1].BlockadedEmpireIDs)
	}
}

func TestHostileEncounterIncludesMilitaryShipCompletedThisTurn(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 2211)
	state := fixture.state
	attacker := &state.Empires[0]
	setCombatFleetTestTech(attacker)
	setCombatFleetTestTech(&state.Empires[1])
	battleSystemID := state.Galaxy.Systems[0].ID
	addCombatFleetTestFleet(state, fixture.targetEmpireID, battleSystemID, 1)
	state.DiplomaticRelations = reciprocalWarRelations(attacker.ID, fixture.targetEmpireID)

	spec := combatFleetTestSpec()
	designID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{
		ID: designID, EmpireID: attacker.ID, Revision: 1, Name: "Current-turn Frigate", Spec: spec,
	})
	state.Colonies[0].Construction = &core.ConstructionState{
		ProjectKind:        core.ConstructionProjectMilitaryShip,
		ProjectID:          MilitaryShipProjectID,
		ProgressPP:         float64(spec.ProductionCostPP) - 0.1,
		ShipDesignID:       designID,
		ShipDesignRevision: 1,
	}
	resolver, err := NewEconomyResolver(loadCommittedEconomyRules(t))
	if err != nil {
		t.Fatal(err)
	}
	resolution, err := resolver.Resolve(encounterTestContext(state), state, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolution.Encounters) != 1 {
		t.Fatalf("construction encounters=%+v construction=%+v ships=%v fleets=%v", resolution.Encounters, state.Colonies[0].Construction, state.Ships, state.StrategicFleets)
	}
	if len(resolution.Encounters[0].Attacker.ShipIDs) != 1 || len(resolution.Encounters[0].Attacker.CombatFleetIDs) != 1 {
		t.Fatalf("new military ship did not join current-turn encounter: %+v", resolution.Encounters[0].Attacker)
	}
	if state.Colonies[0].Construction != nil {
		t.Fatalf("military construction did not complete before encounter: %+v", state.Colonies[0].Construction)
	}
}

func TestEncounterBoundaryCanCreateIndependentSystemBattlesAndCanonicalReciprocalPair(t *testing.T) {
	state := core.NewSmallFixture(2212)
	setCombatFleetTestTech(&state.Empires[0])
	empireB := addEncounterTestEmpire(state, "B")
	empireC := addEncounterTestEmpire(state, "C")
	empireD := addEncounterTestEmpire(state, "D")
	addCombatFleetTestFleet(state, state.Empires[0].ID, state.Galaxy.Systems[0].ID, 1)
	addCombatFleetTestFleet(state, empireB, state.Galaxy.Systems[0].ID, 1)
	addCombatFleetTestFleet(state, empireC, state.Galaxy.Systems[1].ID, 1)
	addCombatFleetTestFleet(state, empireD, state.Galaxy.Systems[1].ID, 1)
	state.DiplomaticRelations = warRelations([2]core.ID{state.Empires[0].ID, empireB}, [2]core.ID{empireC, empireD})
	resolver, err := NewEconomyResolver(loadCommittedEconomyRules(t))
	if err != nil {
		t.Fatal(err)
	}
	_, encounters, err := resolver.prepareEncounterBoundary(encounterTestContext(state), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(encounters) != 2 {
		t.Fatalf("independent-system encounters=%+v", encounters)
	}
	if encounters[0].SystemID >= encounters[1].SystemID {
		t.Fatalf("encounters not ordered by system: %+v", encounters)
	}
	if encounters[0].Attacker.EmpireID != state.Empires[0].ID || encounters[0].Defender.EmpireID != empireB {
		t.Fatalf("reciprocal hostility canonical direction=%+v", encounters[0])
	}
}

func containsCoreID(ids []core.ID, target core.ID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func TestPopulationTransferWaitsForFinalEncounterWaveAndSeesPostBattleBlockade(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 2213)
	state := fixture.state
	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	addCombatFleetTestFleet(state, fixture.blockaderEmpireID, fixture.targetSystemID, 1)
	addCombatFleetTestFleet(state, fixture.targetEmpireID, fixture.targetSystemID, 1)
	state.DiplomaticRelations = reciprocalWarRelations(fixture.blockaderEmpireID, fixture.targetEmpireID)
	transferID := state.NewID()
	state.PopulationTransfers = append(state.PopulationTransfers, core.PopulationTransfer{
		ID: transferID, EmpireID: fixture.targetEmpireID,
		SourceColonyID: fixture.sourceColonyID, DestinationColonyID: fixture.targetColonyID,
		OriginEmpireID: fixture.targetEmpireID, LoyaltyEmpireID: fixture.targetEmpireID,
		AssimilationState: core.PopulationAssimilated, Job: core.PopulationJobFarmer, RemainingTurns: 1,
	})
	ctx := encounterTestContext(state)
	resolver, err := NewEconomyResolver(loadCommittedEconomyRules(t))
	if err != nil {
		t.Fatal(err)
	}
	pre, err := resolver.Resolve(ctx, state, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(pre.Encounters) != 1 || len(state.PopulationTransfers) != 1 || state.PopulationTransfers[0].RemainingTurns != 1 {
		t.Fatalf("transfer advanced before encounter completion: encounters=%v transfers=%+v", pre.Encounters, state.PopulationTransfers)
	}
	if len(state.Galaxy.Systems[1].BlockadedEmpireIDs) != 0 {
		t.Fatalf("blockade materialized before encounter completion: %v", state.Galaxy.Systems[1].BlockadedEmpireIDs)
	}
	post, err := resolver.ResumeAfterEncounters(ctx, state, []EncounterOutcome{{
		BattleID: 1, Encounter: pre.Encounters[0], WinnerSeat: pre.Encounters[0].Attacker.SeatID, Outcome: "attacker_victory",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(state.PopulationTransfers) != 0 {
		t.Fatalf("transfer remained after final encounter wave: %+v", state.PopulationTransfers)
	}
	lost := findDomainEvent(post.Events, "empire.population_transfer_lost")
	if lost == nil {
		t.Fatalf("post-encounter transfer event missing: %+v", post.Events)
	}
	var payload PopulationTransferLostEvent
	if err := json.Unmarshal(lost.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.TransferID != transferID || payload.Reason != "destination_blockaded" {
		t.Fatalf("post-battle transfer result=%+v", payload)
	}
}

func TestDiplomacyWarEnablesSameSystemEncounterAtNextBoundary(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 2214)
	state := fixture.state
	state.MarkEmpiresKnown(fixture.blockaderEmpireID, fixture.targetEmpireID)

	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	addCombatFleetTestFleet(state, fixture.blockaderEmpireID, fixture.targetSystemID, 1)
	addCombatFleetTestFleet(state, fixture.targetEmpireID, fixture.targetSystemID, 1)
	resolver, err := NewEconomyResolver(loadCommittedEconomyRules(t))
	if err != nil {
		t.Fatal(err)
	}
	_, encounters, err := resolver.prepareEncounterBoundary(encounterTestContext(state), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(encounters) != 0 {
		t.Fatalf("neutral same-system fleets created encounter: %+v", encounters)
	}
	declare, _ := NewDeclareWarCommand(1, fixture.targetEmpireID)
	if _, err := ResolveDiplomacyCommand(state, fixture.blockaderEmpireID, 1, declare); err != nil {
		t.Fatal(err)
	}
	_, encounters, err = resolver.prepareEncounterBoundary(encounterTestContext(state), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(encounters) != 1 {
		t.Fatalf("war did not enable same-system encounter: %+v", encounters)
	}
}

func TestAcceptedPeaceBeforeEncounterBoundaryKeepsFleetsButSuppressesBattle(t *testing.T) {
	fixture := newBlockadeTestFixture(t, 2215)
	state := fixture.state
	setCombatFleetTestTech(&state.Empires[0])
	setCombatFleetTestTech(&state.Empires[1])
	firstFleetID, _ := addCombatFleetTestFleet(state, fixture.blockaderEmpireID, fixture.targetSystemID, 1)
	secondFleetID, _ := addCombatFleetTestFleet(state, fixture.targetEmpireID, fixture.targetSystemID, 1)
	state.DiplomaticRelations = reciprocalWarRelations(fixture.blockaderEmpireID, fixture.targetEmpireID)
	offer, _ := NewOfferPeaceCommand(1, fixture.targetEmpireID)
	if _, err := ResolveDiplomacyCommand(state, fixture.blockaderEmpireID, 1, offer); err != nil {
		t.Fatal(err)
	}
	accept, _ := NewAcceptPeaceCommand(1, fixture.blockaderEmpireID)
	if _, err := ResolveDiplomacyCommand(state, fixture.targetEmpireID, 2, accept); err != nil {
		t.Fatal(err)
	}
	resolver, err := NewEconomyResolver(loadCommittedEconomyRules(t))
	if err != nil {
		t.Fatal(err)
	}
	_, encounters, err := resolver.prepareEncounterBoundary(encounterTestContext(state), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(encounters) != 0 {
		t.Fatalf("accepted peace still created encounter: %+v", encounters)
	}
	for _, id := range []core.ID{firstFleetID, secondFleetID} {
		_, fleet := strategicFleetByID(state, id)
		if fleet == nil || fleet.AtSystemID != fixture.targetSystemID {
			t.Fatalf("peace altered fleet %d: %+v", id, fleet)
		}
	}
}
