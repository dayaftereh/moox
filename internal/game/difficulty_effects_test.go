package game

import (
	"encoding/json"
	"math"
	"testing"

	"moox/internal/core"
)

func generatedDifficultyGame(t *testing.T, id core.DifficultyID) NewGameResult {
	t.Helper()
	rules := loadCommittedEconomyRules(t)
	settings := canonicalNewGameSettings()
	settings.DifficultyID = id
	settings.Players[1].BuiltinAIControlled = true
	result, err := rules.NewGame(0x8009, settings)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func colonyForEmpire(t *testing.T, state *core.GameState, empireID core.ID) core.Colony {
	t.Helper()
	for _, colony := range state.Colonies {
		if colony.EmpireID == empireID {
			return colony
		}
	}
	t.Fatalf("empire %d has no colony", empireID)
	return core.Colony{}
}

func nearlyDifficultyEqual(got, want float64) bool {
	return math.Abs(got-want) <= 1e-9
}

func TestDifficultyAppliesFrozenPerRoleDeltasOnlyToBuiltinAI(t *testing.T) {
	normal := generatedDifficultyGame(t, core.DifficultyNormal)
	normalHuman := colonyForEmpire(t, normal.State, normal.Players[0].EmpireID)
	normalAI := colonyForEmpire(t, normal.State, normal.Players[1].EmpireID)

	for _, id := range core.SupportedDifficultyIDs {
		result := generatedDifficultyGame(t, id)
		human := colonyForEmpire(t, result.State, result.Players[0].EmpireID)
		ai := colonyForEmpire(t, result.State, result.Players[1].EmpireID)
		profile, err := DifficultyProfileFor(id)
		if err != nil {
			t.Fatal(err)
		}
		if human.Economy != normalHuman.Economy || human.AdjustedEconomy != normalHuman.AdjustedEconomy {
			t.Fatalf("difficulty %q changed human economy: got=%+v/%+v normal=%+v/%+v", id, human.Economy, human.AdjustedEconomy, normalHuman.Economy, normalHuman.AdjustedEconomy)
		}
		if ai.Population.Total() != normalAI.Population.Total() || ai.Population.Farmers() != normalAI.Population.Farmers() || ai.Population.Workers() != normalAI.Population.Workers() || ai.Population.Scientists() != normalAI.Population.Scientists() {
			t.Fatalf("difficulty %q changed initial population assignment", id)
		}
		wantFood := normalAI.Economy.Food + ai.Population.Farmers()*difficultyEighths(profile.AIFoodPerFarmerEighths)
		wantProduction := normalAI.Economy.Production + ai.Population.Workers()*difficultyEighths(profile.AIProductionPerWorkerEighths)
		wantResearch := normalAI.Economy.Research + ai.Population.Scientists()*difficultyEighths(profile.AIResearchPerScientistEighths)
		wantTax := normalAI.Economy.TaxBC + ai.Population.Total()*difficultyEighths(profile.AITaxBCPerPopulationEighths)
		if !nearlyDifficultyEqual(ai.Economy.Food, wantFood) || !nearlyDifficultyEqual(ai.Economy.Production, wantProduction) || !nearlyDifficultyEqual(ai.Economy.Research, wantResearch) || !nearlyDifficultyEqual(ai.Economy.TaxBC, wantTax) {
			t.Fatalf("difficulty %q AI base economy=%+v want food=%v production=%v research=%v tax=%v", id, ai.Economy, wantFood, wantProduction, wantResearch, wantTax)
		}
	}
}

func TestDifficultyNewGameFixturesAreDeterministicForEverySupportedLevel(t *testing.T) {
	for _, id := range core.SupportedDifficultyIDs {
		t.Run(string(id), func(t *testing.T) {
			first := generatedDifficultyGame(t, id)
			second := generatedDifficultyGame(t, id)
			firstJSON, err := json.Marshal(first.State)
			if err != nil {
				t.Fatal(err)
			}
			secondJSON, err := json.Marshal(second.State)
			if err != nil {
				t.Fatal(err)
			}
			if string(firstJSON) != string(secondJSON) {
				t.Fatalf("same seed/settings/difficulty %q produced different state bytes", id)
			}
			if first.State.DifficultyID != id {
				t.Fatalf("persisted difficulty=%q want=%q", first.State.DifficultyID, id)
			}
		})
	}
}

func TestDifficultyCommandDeficitRateAppliesOnlyToBuiltinAI(t *testing.T) {
	_, resolver := commandPointTestResolver(t)
	cases := []struct {
		id   core.DifficultyID
		want float64
	}{
		{core.DifficultyEasy, 11},
		{core.DifficultyNormal, 10},
		{core.DifficultyHard, 9},
		{core.DifficultyVeryHard, 8.5},
		{core.DifficultyImpossible, 8},
	}
	for _, tc := range cases {
		t.Run(string(tc.id), func(t *testing.T) {
			state := core.NewSmallFixture(0xA509)
			state.DifficultyID = tc.id
			empire := &state.Empires[0]
			empire.BuiltinAIControlled = true
			for i := 0; i < 6; i++ {
				state.Ships = append(state.Ships, core.Ship{ID: state.NewID(), EmpireID: empire.ID, Spec: core.ShipDesignSpec{HullID: "frigate"}})
			}
			if _, err := resolver.settleTreasury(state); err != nil {
				t.Fatal(err)
			}
			if !nearlyDifficultyEqual(empire.Treasury.ShipCommandMaintenanceBC, tc.want) {
				t.Fatalf("difficulty %q command maintenance=%v want=%v", tc.id, empire.Treasury.ShipCommandMaintenanceBC, tc.want)
			}
		})
	}

	state := core.NewSmallFixture(0xA50C)
	state.DifficultyID = core.DifficultyImpossible
	empire := &state.Empires[0]
	for i := 0; i < 6; i++ {
		state.Ships = append(state.Ships, core.Ship{ID: state.NewID(), EmpireID: empire.ID, Spec: core.ShipDesignSpec{HullID: "frigate"}})
	}
	if _, err := resolver.settleTreasury(state); err != nil {
		t.Fatal(err)
	}
	if empire.Treasury.ShipCommandMaintenanceBC != 10 {
		t.Fatalf("human command maintenance=%v want standard 10", empire.Treasury.ShipCommandMaintenanceBC)
	}
}

func normalizedInitialDifficultyState(t *testing.T, state *core.GameState) []byte {
	t.Helper()
	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	var clone core.GameState
	if err := json.Unmarshal(raw, &clone); err != nil {
		t.Fatal(err)
	}
	clone.DifficultyID = core.DifficultyNormal
	aiEmpires := map[core.ID]bool{}
	for _, empire := range clone.Empires {
		if empire.BuiltinAIControlled {
			aiEmpires[empire.ID] = true
		}
	}
	for i := range clone.Empires {
		if aiEmpires[clone.Empires[i].ID] {
			clone.Empires[i].FoodLogistics = core.EmpireFoodLogistics{}
		}
	}
	for i := range clone.Colonies {
		if aiEmpires[clone.Colonies[i].EmpireID] {
			clone.Colonies[i].Economy = core.ColonyEconomy{}
			clone.Colonies[i].AdjustedEconomy = core.ColonyEconomy{}
			clone.Colonies[i].PopulationDynamics = core.ColonyPopulationDynamics{}
		}
	}
	normalized, err := json.Marshal(&clone)
	if err != nil {
		t.Fatal(err)
	}
	return normalized
}

func TestDifficultyChangesOnlyFrozenInitialStateEffects(t *testing.T) {
	normal := generatedDifficultyGame(t, core.DifficultyNormal)
	want := string(normalizedInitialDifficultyState(t, normal.State))
	for _, id := range core.SupportedDifficultyIDs {
		t.Run(string(id), func(t *testing.T) {
			result := generatedDifficultyGame(t, id)
			got := string(normalizedInitialDifficultyState(t, result.State))
			if got != want {
				t.Fatalf("difficulty %q changed initial authoritative state outside difficulty_id and direct/derived built-in-AI food/economy fields", id)
			}
		})
	}
}
