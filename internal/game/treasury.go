package game

import (
	"fmt"

	"moox/internal/core"
)

const OriginalNewGameStartingTreasuryBC = 50.0

// InitializeNewGameTreasury applies the directly verified MOO2 1.31 starting
// Treasury value. It is intentionally separate from technology initialization
// so future New Game orchestration can compose economy and technology domains.
func InitializeNewGameTreasury(state *core.GameState) error {
	if state == nil {
		return fmt.Errorf("game state must not be nil")
	}
	for i := range state.Empires {
		if state.Empires[i].Treasury != (core.EmpireTreasuryState{}) {
			return fmt.Errorf("empire %d treasury state is already initialized", state.Empires[i].ID)
		}
		state.Empires[i].Treasury.BalanceBC = OriginalNewGameStartingTreasuryBC
	}
	return nil
}

type TreasurySettledEvent struct {
	EmpireID          core.ID                  `json:"empire_id"`
	PreviousBalanceBC float64                  `json:"previous_balance_bc"`
	Current           core.EmpireTreasuryState `json:"current"`
}

type treasurySettlementPlan struct {
	EmpireIndex   int
	CommandPoints core.EmpireCommandPoints
	Treasury      core.EmpireTreasuryState
	Event         DomainEvent
}

func (r *EconomyResolver) prepareTreasurySettlement(state *core.GameState, empireIndex int) (treasurySettlementPlan, error) {
	empire := &state.Empires[empireIndex]
	commandPoints, err := r.Rules.deriveEmpireCommandPoints(state, empire)
	if err != nil {
		return treasurySettlementPlan{}, err
	}

	previousBalanceBC := empire.Treasury.BalanceBC
	var taxIncomeBC float64
	var buildingMaintenanceBC float64
	for ci := range state.Colonies {
		colony := &state.Colonies[ci]
		if colony.EmpireID != empire.ID {
			continue
		}
		taxIncomeBC += colony.AdjustedEconomy.TaxBC
		for _, buildingID := range colony.Buildings {
			definition, ok := r.Rules.BuildingDefinitions[buildingID]
			if !ok {
				return treasurySettlementPlan{}, fmt.Errorf("empire %d colony %d owns unknown building %q", empire.ID, colony.ID, buildingID)
			}
			buildingMaintenanceBC += float64(definition.MaintenanceBC)
		}
	}

	surplusFoodIncomeBC := empire.FoodLogistics.SurplusFoodIncomeBC
	freighterOperatingCostBC := empire.FoodLogistics.FreighterOperatingCostBC
	shipCommandMaintenanceBC := float64(commandPointOverage(commandPoints)) * r.Rules.CommandPoints.StandardOverageBCPerPoint
	grossIncomeBC := taxIncomeBC + surplusFoodIncomeBC
	totalModeledMaintenanceBC := buildingMaintenanceBC + freighterOperatingCostBC + shipCommandMaintenanceBC
	netModeledIncomeBC := grossIncomeBC - totalModeledMaintenanceBC
	treasury := core.EmpireTreasuryState{
		BalanceBC:                 previousBalanceBC + netModeledIncomeBC,
		TaxIncomeBC:               taxIncomeBC,
		SurplusFoodIncomeBC:       surplusFoodIncomeBC,
		GrossIncomeBC:             grossIncomeBC,
		BuildingMaintenanceBC:     buildingMaintenanceBC,
		FreighterOperatingCostBC:  freighterOperatingCostBC,
		ShipCommandMaintenanceBC:  shipCommandMaintenanceBC,
		TotalModeledMaintenanceBC: totalModeledMaintenanceBC,
		NetModeledIncomeBC:        netModeledIncomeBC,
	}
	event, err := NewDomainEvent("empire.treasury_settled", 0, 0, TreasurySettledEvent{
		EmpireID:          empire.ID,
		PreviousBalanceBC: previousBalanceBC,
		Current:           treasury,
	})
	if err != nil {
		return treasurySettlementPlan{}, err
	}
	return treasurySettlementPlan{
		EmpireIndex:   empireIndex,
		CommandPoints: commandPoints,
		Treasury:      treasury,
		Event:         event,
	}, nil
}

func (r *EconomyResolver) settleTreasury(state *core.GameState) ([]DomainEvent, error) {
	if r == nil || r.Rules == nil {
		return nil, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}

	empireIndexes := sortedEmpireIndexes(state)
	plans := make([]treasurySettlementPlan, 0, len(empireIndexes))
	for _, index := range empireIndexes {
		plan, err := r.prepareTreasurySettlement(state, index)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	events := make([]DomainEvent, 0, len(plans))
	for _, plan := range plans {
		empire := &state.Empires[plan.EmpireIndex]
		empire.CommandPoints = plan.CommandPoints
		empire.Treasury = plan.Treasury
		events = append(events, plan.Event)
	}
	return events, nil
}
