package session

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func TestPlayerViewProjectsPlayerSafeRecentResolutionSummaries(t *testing.T) {
	state, seats := twoSeatFixture(t)
	humanEmpireID := state.Empires[0].ID
	enemyEmpireID := state.Empires[1].ID
	s, err := NewGameSession("summary-game", state, seats)
	if err != nil {
		t.Fatal(err)
	}

	humanResearch, err := game.NewDomainEvent("empire.research_completed", 0, 0, game.ResearchCompletedEvent{
		EmpireID:       humanEmpireID,
		TechFieldID:    17,
		SelectionMode:  core.ResearchSelectionMode("creative_choice"),
		TechnologyIDs:  []int{41},
		TechnologyKeys: []string{"automated_factory"},
	})
	if err != nil {
		t.Fatal(err)
	}
	enemyResearch, err := game.NewDomainEvent("empire.research_completed", 0, 0, game.ResearchCompletedEvent{
		EmpireID:       enemyEmpireID,
		TechFieldID:    99,
		SelectionMode:  core.ResearchSelectionMode("creative_choice"),
		TechnologyIDs:  []int{199},
		TechnologyKeys: []string{"enemy_secret"},
	})
	if err != nil {
		t.Fatal(err)
	}
	humanGrant, err := game.NewDomainEvent(game.EventTechnologyGranted, 0, 0, game.TechnologyGrantedEvent{
		EmpireID:          humanEmpireID,
		TechnologyID:      55,
		TechnologyKey:     "reinforced_hull",
		TechnologyNameKey: "technology.reinforced_hull",
		TechFieldID:       18,
		SourceKind:        game.TechnologyGrantSource("capture"),
	})
	if err != nil {
		t.Fatal(err)
	}
	enemyGrant, err := game.NewDomainEvent(game.EventTechnologyGranted, 0, 0, game.TechnologyGrantedEvent{
		EmpireID:          enemyEmpireID,
		TechnologyID:      200,
		TechnologyKey:     "enemy_grant",
		TechnologyNameKey: "technology.enemy_grant",
		TechFieldID:       100,
		SourceKind:        game.TechnologyGrantSource("capture"),
	})
	if err != nil {
		t.Fatal(err)
	}

	s.mu.Lock()
	s.appendResolvedEventsLocked([]game.DomainEvent{humanResearch, enemyResearch, humanGrant, enemyGrant})
	s.appendEventLocked(protocol.EventScopeBattle, "battle_completed", 0, battle.View{
		Spec: battle.Spec{
			ID:            7,
			GameID:        "summary-game",
			StrategicTurn: state.Turn,
			SystemID:      12,
			Attacker:      battle.Side{EmpireID: humanEmpireID, SeatID: 1, ShipIDs: []core.ID{101, 102}},
			Defender:      battle.Side{EmpireID: enemyEmpireID, SeatID: 2, ShipIDs: []core.ID{201}},
			Participants:  []protocol.SeatID{1, 2},
		},
		Phase:  battle.PhaseCompleted,
		Result: &battle.Result{WinnerSeat: 1, WinnerSeats: []protocol.SeatID{1}, Outcome: "victory", DestroyedShipIDs: []core.ID{102, 201}},
	})
	s.appendEventLocked(protocol.EventScopeStrategic, "empire.invasion_resolved", 1, game.InvasionResolvedEvent{
		SystemID:                  12,
		ColonyID:                  33,
		AttackerEmpireID:          humanEmpireID,
		DefenderEmpireID:          enemyEmpireID,
		SelectedTransportFleetIDs: []core.ID{77},
		InitialAttackerInfantry:   5,
		InitialDefenderInfantry:   2,
		SurvivingAttackerInfantry: 3,
		Captured:                  true,
	})
	s.appendEventLocked(protocol.EventScopeSession, "empire_eliminated", 0, EmpireEliminatedEvent{EmpireID: enemyEmpireID})
	s.appendEventLocked(protocol.EventScopeSession, "game_completed", 0, Result{
		Kind:                ResultConquest,
		WinnerEmpireID:      humanEmpireID,
		WinnerSeatID:        1,
		EliminatedEmpireIDs: []core.ID{enemyEmpireID},
		CompletedTurn:       state.Turn,
		CompletedRevision:   s.revision,
	})
	s.mu.Unlock()

	human, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if got := summaryKinds(human.RecentResolutions); !reflect.DeepEqual(got, []ResolutionSummaryKind{
		ResolutionSummaryResearchBreakthrough,
		ResolutionSummaryTechnologyGranted,
		ResolutionSummaryBattleCompleted,
		ResolutionSummaryInvasionResolved,
		ResolutionSummaryEmpireEliminated,
		ResolutionSummaryGameCompleted,
	}) {
		t.Fatalf("human recent resolution kinds=%v", got)
	}
	if human.RecentResolutions[0].Research == nil || human.RecentResolutions[0].Research.EmpireID != humanEmpireID || containsIntSummary(human.RecentResolutions[0].Research.TechnologyIDs, 199) {
		t.Fatalf("human research summary leaked or missing data: %+v", human.RecentResolutions[0])
	}
	if human.RecentResolutions[1].TechnologyGrant == nil || human.RecentResolutions[1].TechnologyGrant.TechnologyID != 55 {
		t.Fatalf("human technology grant summary=%+v", human.RecentResolutions[1])
	}
	battleSummary := human.RecentResolutions[2].Battle
	if battleSummary == nil || battleSummary.WinnerEmpireID != humanEmpireID || !reflect.DeepEqual(battleSummary.DestroyedShipIDs, []core.ID{102, 201}) || !reflect.DeepEqual(battleSummary.SurvivingShipIDs, []core.ID{101}) {
		t.Fatalf("human battle summary=%+v", battleSummary)
	}
	if human.RecentResolutions[3].Invasion == nil || human.RecentResolutions[3].Invasion.Outcome != "captured" {
		t.Fatalf("human invasion summary=%+v", human.RecentResolutions[3])
	}
	firstIDs := summaryIDs(human.RecentResolutions)
	humanAgain, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(firstIDs, summaryIDs(humanAgain.RecentResolutions)) {
		t.Fatalf("summary IDs are not stable: first=%v second=%v", firstIDs, summaryIDs(humanAgain.RecentResolutions))
	}

	enemy, err := s.PlayerView(2)
	if err != nil {
		t.Fatal(err)
	}
	if got := summaryKinds(enemy.RecentResolutions); !reflect.DeepEqual(got, []ResolutionSummaryKind{
		ResolutionSummaryResearchBreakthrough,
		ResolutionSummaryTechnologyGranted,
		ResolutionSummaryBattleCompleted,
		ResolutionSummaryInvasionResolved,
		ResolutionSummaryEmpireEliminated,
		ResolutionSummaryGameCompleted,
	}) {
		t.Fatalf("enemy recent resolution kinds=%v", got)
	}
	if enemy.RecentResolutions[0].Research == nil || enemy.RecentResolutions[0].Research.EmpireID != enemyEmpireID || enemy.RecentResolutions[0].Research.TechFieldID != 99 {
		t.Fatalf("enemy research summary=%+v", enemy.RecentResolutions[0])
	}
	if enemy.RecentResolutions[1].TechnologyGrant == nil || enemy.RecentResolutions[1].TechnologyGrant.TechnologyID != 200 {
		t.Fatalf("enemy technology grant summary=%+v", enemy.RecentResolutions[1])
	}
}

func TestRecentResolutionSummariesExcludeNonParticipantsAndOldEvents(t *testing.T) {
	state, seats := twoSeatFixture(t)
	thirdEmpire := core.Empire{ID: state.NewID(), Name: "Third", RaceID: "human"}
	state.Empires = append(state.Empires, thirdEmpire)
	seats = append(seats, Seat{ID: 3, EmpireID: thirdEmpire.ID, Name: "Third", Controller: ControllerRemoteHuman})
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("summary-private", state, seats)
	if err != nil {
		t.Fatal(err)
	}

	s.mu.Lock()
	s.state.Turn = 4
	s.events = append(s.events, protocol.DomainEvent{
		SchemaVersion: protocol.EventSchemaVersion,
		Sequence:      s.nextEventSequence,
		Turn:          1,
		Revision:      s.revision,
		Scope:         protocol.EventScopeSession,
		Kind:          "empire_eliminated",
		Data:          mustSummaryJSON(t, EmpireEliminatedEvent{EmpireID: state.Empires[1].ID}),
	})
	s.nextEventSequence++
	s.appendEventLocked(protocol.EventScopeBattle, "battle_completed", 0, battle.View{
		Spec: battle.Spec{
			ID:           9,
			GameID:       "summary-private",
			SystemID:     5,
			Attacker:     battle.Side{EmpireID: state.Empires[0].ID, SeatID: 1, ShipIDs: []core.ID{1}},
			Defender:     battle.Side{EmpireID: state.Empires[1].ID, SeatID: 2, ShipIDs: []core.ID{2}},
			Participants: []protocol.SeatID{1, 2},
		},
		Phase:  battle.PhaseCompleted,
		Result: &battle.Result{WinnerSeat: 1, Outcome: "victory"},
	})
	s.mu.Unlock()

	third, err := s.PlayerView(3)
	if err != nil {
		t.Fatal(err)
	}
	if len(third.RecentResolutions) != 0 {
		t.Fatalf("nonparticipant received private/old summaries: %+v", third.RecentResolutions)
	}
}

func summaryKinds(summaries []ResolutionSummary) []ResolutionSummaryKind {
	out := make([]ResolutionSummaryKind, len(summaries))
	for i := range summaries {
		out[i] = summaries[i].Kind
	}
	return out
}

func summaryIDs(summaries []ResolutionSummary) []string {
	out := make([]string, len(summaries))
	for i := range summaries {
		out[i] = summaries[i].ID
	}
	return out
}

func containsIntSummary(values []int, target int) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func mustSummaryJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
