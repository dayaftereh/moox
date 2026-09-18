package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"

	"moox/internal/core"
	"moox/internal/protocol"
)

const CommandBuyConstruction = "colony.buy_construction"

type BuyConstructionPayload struct {
	ColonyID core.ID `json:"colony_id"`
}

type ConstructionBuyoutQuote struct {
	ColonyID         core.ID                      `json:"colony_id"`
	ProjectKind      core.ConstructionProjectKind `json:"project_kind"`
	ProjectID        string                       `json:"project_id"`
	ShipDesignID     core.ID                      `json:"ship_design_id,omitempty"`
	DesignRevision   uint32                       `json:"ship_design_revision,omitempty"`
	ProductionCostPP float64                      `json:"production_cost_pp"`
	ProgressPP       float64                      `json:"progress_pp"`
	RemainingPP      float64                      `json:"remaining_pp"`
	CostBC           float64                      `json:"cost_bc"`
	Affordable       bool                         `json:"affordable"`
}

type ConstructionBoughtEvent struct {
	ColonyID           core.ID                      `json:"colony_id"`
	EmpireID           core.ID                      `json:"empire_id"`
	ProjectKind        core.ConstructionProjectKind `json:"project_kind"`
	ProjectID          string                       `json:"project_id"`
	ProductionCostPP   float64                      `json:"production_cost_pp"`
	PreviousProgressPP float64                      `json:"previous_progress_pp"`
	CurrentProgressPP  float64                      `json:"current_progress_pp"`
	CostBC             float64                      `json:"cost_bc"`
	PreviousBalanceBC  float64                      `json:"previous_balance_bc"`
	CurrentBalanceBC   float64                      `json:"current_balance_bc"`
}

func NewBuyConstructionCommand(sequence uint32, payload BuyConstructionPayload) (protocol.Command, error) {
	if payload.ColonyID == 0 {
		return protocol.Command{}, fmt.Errorf("%s requires colony_id", CommandBuyConstruction)
	}
	return protocol.NewCommand(sequence, CommandBuyConstruction, payload)
}

func IsConstructionBuyoutCommand(kind string) bool {
	return kind == CommandBuyConstruction
}

func decodeBuyConstruction(command protocol.Command) (BuyConstructionPayload, error) {
	var payload BuyConstructionPayload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return BuyConstructionPayload{}, fmt.Errorf("decode %s: %w", CommandBuyConstruction, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return BuyConstructionPayload{}, fmt.Errorf("decode %s: trailing JSON value", CommandBuyConstruction)
		}
		return BuyConstructionPayload{}, fmt.Errorf("decode %s trailing data: %w", CommandBuyConstruction, err)
	}
	if payload.ColonyID == 0 {
		return BuyConstructionPayload{}, fmt.Errorf("%s requires colony_id", CommandBuyConstruction)
	}
	return payload, nil
}

// ConstructionBuyoutCostBC implements the original MOO2 1.31 rush-buy curve
// directly in the engine's continuous PP/BC domain:
//
//	 0..10% complete: 4X - 10Y
//	10..50% complete: 3.5X - 5Y
//	50..100% complete: 2X - 2Y
//
// where X is total PP cost and Y is PP already produced.
func ConstructionBuyoutCostBC(productionCostPP, progressPP float64) (float64, error) {
	if math.IsNaN(productionCostPP) || math.IsInf(productionCostPP, 0) || productionCostPP <= 0 {
		return 0, fmt.Errorf("construction buyout requires positive finite production cost, got %v", productionCostPP)
	}
	if math.IsNaN(progressPP) || math.IsInf(progressPP, 0) || progressPP < 0 {
		return 0, fmt.Errorf("construction buyout requires non-negative finite progress, got %v", progressPP)
	}
	y := math.Min(progressPP, productionCostPP)
	var cost float64
	switch {
	case y <= productionCostPP*0.10:
		cost = 4*productionCostPP - 10*y
	case y <= productionCostPP*0.50:
		cost = 3.5*productionCostPP - 5*y
	default:
		cost = 2*productionCostPP - 2*y
	}
	if cost < 1e-9 {
		return 0, nil
	}
	return cost, nil
}

func (r *EconomyResolver) ConstructionBuyoutQuotes(state *core.GameState, empireID core.ID) ([]ConstructionBuyoutQuote, error) {
	if r == nil || r.Rules == nil {
		return nil, fmt.Errorf("construction buyout requires economy resolver")
	}
	if state == nil {
		return nil, fmt.Errorf("construction buyout requires game state")
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return nil, fmt.Errorf("unknown empire %d", empireID)
	}
	quotes := make([]ConstructionBuyoutQuote, 0)
	for i := range state.Colonies {
		colony := &state.Colonies[i]
		if colony.EmpireID != empireID || colony.Construction == nil {
			continue
		}
		quote, err := r.constructionBuyoutQuote(state, colony, empire.Treasury.BalanceBC)
		if err != nil {
			if colony.Construction.ProjectKind == core.ConstructionProjectHousing {
				continue
			}
			return nil, fmt.Errorf("colony %d: %w", colony.ID, err)
		}
		quotes = append(quotes, quote)
	}
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].ColonyID < quotes[j].ColonyID })
	return quotes, nil
}

func (r *EconomyResolver) constructionBuyoutQuote(state *core.GameState, colony *core.Colony, balanceBC float64) (ConstructionBuyoutQuote, error) {
	if colony == nil || colony.Construction == nil {
		return ConstructionBuyoutQuote{}, fmt.Errorf("colony has no active construction")
	}
	if colony.Construction.ProjectKind == core.ConstructionProjectHousing {
		return ConstructionBuyoutQuote{}, fmt.Errorf("housing cannot be bought")
	}
	costPP, err := r.constructionProjectCostPP(state, colony, colony.Construction)
	if err != nil {
		return ConstructionBuyoutQuote{}, err
	}
	progressPP := math.Min(math.Max(0, colony.Construction.ProgressPP), costPP)
	costBC, err := ConstructionBuyoutCostBC(costPP, progressPP)
	if err != nil {
		return ConstructionBuyoutQuote{}, err
	}
	return ConstructionBuyoutQuote{
		ColonyID:         colony.ID,
		ProjectKind:      colony.Construction.ProjectKind,
		ProjectID:        colony.Construction.ProjectID,
		ShipDesignID:     colony.Construction.ShipDesignID,
		DesignRevision:   colony.Construction.ShipDesignRevision,
		ProductionCostPP: costPP,
		ProgressPP:       progressPP,
		RemainingPP:      math.Max(0, costPP-progressPP),
		CostBC:           costBC,
		Affordable:       balanceBC+1e-9 >= costBC,
	}, nil
}

func (r *EconomyResolver) ResolveConstructionBuyoutCommand(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) ([]DomainEvent, error) {
	if r == nil || r.Rules == nil {
		return nil, fmt.Errorf("construction buyout requires economy resolver")
	}
	if command.Kind != CommandBuyConstruction {
		return nil, fmt.Errorf("unsupported construction buyout command %q", command.Kind)
	}
	payload, err := decodeBuyConstruction(command)
	if err != nil {
		return nil, err
	}
	colony := colonyByID(state, payload.ColonyID)
	if colony == nil {
		return nil, fmt.Errorf("unknown colony %d", payload.ColonyID)
	}
	if colony.EmpireID != empireID {
		return nil, fmt.Errorf("seat %d cannot buy construction on colony %d owned by empire %d", seatID, colony.ID, colony.EmpireID)
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return nil, fmt.Errorf("unknown empire %d", empireID)
	}
	quote, err := r.constructionBuyoutQuote(state, colony, empire.Treasury.BalanceBC)
	if err != nil {
		return nil, err
	}
	if quote.CostBC <= 1e-9 {
		return nil, fmt.Errorf("colony %d construction is already fully funded", colony.ID)
	}
	if !quote.Affordable {
		return nil, fmt.Errorf("insufficient treasury: need %.3f BC, have %.3f BC", quote.CostBC, empire.Treasury.BalanceBC)
	}
	previousBalance := empire.Treasury.BalanceBC
	previousProgress := colony.Construction.ProgressPP
	empire.Treasury.BalanceBC -= quote.CostBC
	if math.Abs(empire.Treasury.BalanceBC) < 1e-9 {
		empire.Treasury.BalanceBC = 0
	}
	colony.Construction.ProgressPP = quote.ProductionCostPP
	event, err := NewDomainEvent("colony.construction_bought", seatID, command.Sequence, ConstructionBoughtEvent{
		ColonyID: colony.ID, EmpireID: empire.ID, ProjectKind: quote.ProjectKind, ProjectID: quote.ProjectID,
		ProductionCostPP: quote.ProductionCostPP, PreviousProgressPP: previousProgress, CurrentProgressPP: quote.ProductionCostPP,
		CostBC: quote.CostBC, PreviousBalanceBC: previousBalance, CurrentBalanceBC: empire.Treasury.BalanceBC,
	})
	if err != nil {
		return nil, err
	}
	return []DomainEvent{event}, nil
}
