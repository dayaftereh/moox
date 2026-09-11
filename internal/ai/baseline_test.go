package ai

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"moox/internal/battle"

	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

func TestBaselinePlannerIsDeterministicNonMutatingAndUsesFoodSafeProjection(t *testing.T) {
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := rules.NewGame(0x8009, game.NewGameSettings{
		GalaxySize:      game.GalaxySizeSmall,
		GalaxyAge:       game.GalaxyAgeNormal,
		TechnologyLevel: game.NewGameTechnologyAverage,
		Players: []game.NewGamePlayerSpec{
			{SeatID: 1, EmpireName: "Human", RaceID: "human"},
			{SeatID: 2, EmpireName: "Darlok", RaceID: "darlok"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	seats := []session.Seat{
		{ID: 1, EmpireID: generated.Players[0].EmpireID, Name: "Human", Controller: session.ControllerBuiltinAI},
		{ID: 2, EmpireID: generated.Players[1].EmpireID, Name: "Darlok", Controller: session.ControllerBuiltinAI},
	}
	s, err := session.NewGameSession("planner-determinism", generated.State, seats)
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	view, err := s.DecisionView(2, resolver)
	if err != nil {
		t.Fatal(err)
	}
	before, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}

	first, err := Plan(view)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Plan(view)
	if err != nil {
		t.Fatal(err)
	}
	firstBytes, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondBytes, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("same decision input produced different actions:\nfirst=%s\nsecond=%s", firstBytes, secondBytes)
	}
	after, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("baseline planner mutated its PlayerDecisionView input")
	}
	if first.PolicyVersion != PolicyVersion || first.Kind != ActionSubmitTurn || first.Batch == nil {
		t.Fatalf("initial action=%+v", first)
	}
	if err := first.Batch.Validate(); err != nil {
		t.Fatalf("planner returned invalid batch: %v", err)
	}

	foundResearch := false
	foundFoodSafePopulation := false
	for _, command := range first.Batch.Commands {
		switch command.Kind {
		case game.CommandSelectResearch:
			foundResearch = true
		case game.CommandAssignPopulation:
			var payload game.AssignPopulationPayload
			if err := json.Unmarshal(command.Payload, &payload); err != nil {
				t.Fatal(err)
			}
			for _, colony := range view.Decisions.Population {
				if colony.ColonyID != payload.ColonyID {
					continue
				}
				for _, choice := range colony.Choices {
					if choice.Farmers == payload.Farmers && choice.Workers == payload.Workers && choice.Scientists == payload.Scientists && choice.FoodSafe {
						foundFoodSafePopulation = true
					}
				}
			}
		}
	}
	if !foundResearch {
		t.Fatal("initial baseline AI action omitted research selection")
	}
	if !foundFoodSafePopulation {
		t.Fatalf("initial Darlok baseline AI action did not use a server-projected food-safe assignment: %+v", first.Batch.Commands)
	}
}

func TestBaselinePlannerChoosesFirstAuthoritativeBattleAction(t *testing.T) {
	first, err := battle.NewFireBeamCommand(7, battle.FireBeamPayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	second, err := battle.NewEndActivationCommand(7, battle.EndActivationPayload{ShipID: 100})
	if err != nil {
		t.Fatal(err)
	}
	view := session.PlayerDecisionView{
		Phase:     session.PhaseEncounters,
		Decisions: session.DecisionCatalog{Battles: []session.BattleDecision{{BattleID: 9, Actions: []protocol.Command{first, second}}}},
	}
	action, err := Plan(view)
	if err != nil {
		t.Fatal(err)
	}
	if action.Kind != ActionBattle || action.BattleID != 9 || action.Command == nil {
		t.Fatalf("battle action=%+v", action)
	}
	got, _ := json.Marshal(action.Command)
	want, _ := json.Marshal(first)
	if !bytes.Equal(got, want) {
		t.Fatalf("planner battle command=%s want=%s", got, want)
	}
}
func TestBaselinePlannerOmitsTechnologyIDForAllApplicationsResearch(t *testing.T) {
	view := session.PlayerDecisionView{
		GameID:   "ai-all-research",
		Revision: 1,
		Turn:     1,
		Phase:    session.PhasePlanning,
		Seat: session.SeatView{Seat: session.Seat{
			ID: 3, EmpireID: 3, Name: "Psilon", Controller: session.ControllerBuiltinAI,
		}},
		Empire: core.Empire{ID: 3, Name: "Psilon", RaceID: "psilon"},
		Decisions: session.DecisionCatalog{Research: []game.ResearchChoice{{
			TechFieldID: 9, BaseCostRP: 250, SelectionMode: core.ResearchSelectionAll,
			TechnologyIDs: []int{51, 106},
		}}},
	}
	action, err := Plan(view)
	if err != nil {
		t.Fatal(err)
	}
	if action.Kind != ActionSubmitTurn || action.Batch == nil || len(action.Batch.Commands) == 0 {
		t.Fatalf("all-applications research action=%+v", action)
	}
	command := action.Batch.Commands[0]
	if command.Kind != game.CommandSelectResearch {
		t.Fatalf("first command=%q want %q", command.Kind, game.CommandSelectResearch)
	}
	var payload game.SelectResearchPayload
	if err := json.Unmarshal(command.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.TechFieldID != 9 || payload.TechnologyID != 0 {
		t.Fatalf("all-applications payload=%+v, technology_id must be omitted", payload)
	}
}
