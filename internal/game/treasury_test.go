package game

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"moox/internal/core"
)

func TestInitializeNewGameTreasuryUsesOriginalFiftyBC(t *testing.T) {
	state := core.NewSmallFixture(0x9201)
	state.Empires = append(state.Empires, core.Empire{ID: state.NewID(), Name: "Second", RaceID: "human"})
	if err := InitializeNewGameTreasury(state); err != nil {
		t.Fatal(err)
	}
	for i := range state.Empires {
		if state.Empires[i].Treasury.BalanceBC != 50 {
			t.Fatalf("Empire %d starting Treasury=%v want=50 BC", state.Empires[i].ID, state.Empires[i].Treasury.BalanceBC)
		}
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("initialized Treasury state invalid: %v", err)
	}
	if err := InitializeNewGameTreasury(state); err == nil {
		t.Fatal("expected duplicate Treasury initialization to fail")
	}
}

func TestTreasuryStateAllowsFiniteNegativeBalanceButRejectsInvalidComponents(t *testing.T) {
	state := core.NewSmallFixture(0x9202)
	state.Empires[0].Treasury.BalanceBC = -12.5
	if err := state.Validate(); err != nil {
		t.Fatalf("finite negative Treasury balance should remain representable: %v", err)
	}
	state.Empires[0].Treasury.GrossIncomeBC = -1
	if err := state.Validate(); err == nil {
		t.Fatal("expected negative gross modeled income to fail validation")
	}
}
func TestSettleTreasuryAppliesOnlyCurrentlyModeledLedgerComponents(t *testing.T) {
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0x9203)
	empire := &state.Empires[0]
	empire.Treasury.BalanceBC = 50
	state.Colonies[0].AdjustedEconomy.TaxBC = 4.5
	state.Colonies[0].Buildings = []string{"holo_simulator"}
	empire.FoodLogistics.SurplusFoodIncomeBC = 2.25
	empire.FoodLogistics.FreighterOperatingCostBC = 1

	events, err := resolver.settleTreasury(state)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != "empire.treasury_settled" {
		t.Fatalf("Treasury events=%+v", events)
	}
	want := core.EmpireTreasuryState{
		BalanceBC:                 54.75,
		TaxIncomeBC:               4.5,
		SurplusFoodIncomeBC:       2.25,
		GrossIncomeBC:             6.75,
		BuildingMaintenanceBC:     1,
		FreighterOperatingCostBC:  1,
		TotalModeledMaintenanceBC: 2,
		NetModeledIncomeBC:        4.75,
	}
	if empire.Treasury != want {
		t.Fatalf("Treasury=%+v want=%+v", empire.Treasury, want)
	}
	var payload TreasurySettledEvent
	if err := json.Unmarshal(events[0].Data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.EmpireID != empire.ID || payload.PreviousBalanceBC != 50 || payload.Current != want {
		t.Fatalf("Treasury payload=%+v", payload)
	}
}
