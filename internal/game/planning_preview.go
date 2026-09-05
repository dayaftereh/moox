package game

import (
	"fmt"
	"math"
	"sort"

	"moox/internal/core"
	"moox/internal/protocol"
)

type PlanningFreighterPreview struct {
	Total            int `json:"total"`
	FoodUsed         int `json:"food_used"`
	TransferReserved int `json:"transfer_reserved"`
	Available        int `json:"available"`
}

type PlanningResearchPreview struct {
	TechFieldID   int                        `json:"tech_field_id,omitempty"`
	SelectionMode core.ResearchSelectionMode `json:"selection_mode,omitempty"`
	TechnologyIDs []int                      `json:"technology_ids,omitempty"`
	ProgressRP    float64                    `json:"progress_rp"`
	CostRP        float64                    `json:"cost_rp"`
	RemainingRP   float64                    `json:"remaining_rp"`
	RPPerTurn     float64                    `json:"rp_per_turn"`
	ETATurns      *int                       `json:"eta_turns,omitempty"`
}

type PlanningConstructionPreview struct {
	Project     core.ConstructionState `json:"project"`
	CostPP      float64                `json:"cost_pp"`
	RemainingPP float64                `json:"remaining_pp"`
	ETATurns    *int                   `json:"eta_turns,omitempty"`
}

type PlanningColonyPreview struct {
	Colony                  core.Colony                   `json:"colony"`
	FreePopulationCapacity  float64                       `json:"free_population_capacity"`
	PopulationGrowthPerTurn float64                       `json:"population_growth_per_turn"`
	PopulationLossPerTurn   float64                       `json:"population_loss_per_turn"`
	NextPopulationETATurns  *int                          `json:"next_population_eta_turns,omitempty"`
	Construction            []PlanningConstructionPreview `json:"construction"`
}

type PlanningPreviewProjection struct {
	EmpireID            core.ID                    `json:"empire_id"`
	TreasuryBalanceBC   float64                    `json:"treasury_balance_bc"`
	NetModeledIncomeBC  float64                    `json:"net_modeled_income_bc"`
	Freighters          PlanningFreighterPreview   `json:"freighters"`
	CommandPoints       core.EmpireCommandPoints   `json:"command_points"`
	CommandPointOverage int                        `json:"command_point_overage"`
	Research            PlanningResearchPreview    `json:"research"`
	Colonies            []PlanningColonyPreview    `json:"colonies"`
	PopulationTransfers []PopulationTransferChoice `json:"population_transfers,omitempty"`
}

// ApplyPlanningDraft applies legal Planning commands to a disposable State
// clone. It intentionally does not run strategic resolution, advance the turn,
// emit session events or touch a hosted GameSession revision.
func (r *EconomyResolver) ApplyPlanningDraft(state *core.GameState, empireID core.ID, seatID protocol.SeatID, commands []protocol.Command) error {
	if r == nil || r.Rules == nil {
		return fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return fmt.Errorf("game state must not be nil")
	}
	if empireByID(state, empireID) == nil {
		return fmt.Errorf("unknown empire %d", empireID)
	}
	for _, command := range commands {
		var err error
		switch command.Kind {
		case CommandAssignPopulation:
			err = r.applyPlanningPopulationAssignment(state, empireID, command)
		case CommandSaveMilitaryDesign:
			_, err = r.saveMilitaryDesign(state, empireID, seatID, command)
		case CommandSetConstructionQueue:
			_, err = r.setConstructionQueue(state, empireID, seatID, command)
		case CommandQueueBuilding:
			_, err = r.queueBuilding(state, empireID, seatID, command)
		case CommandQueueColonyShip:
			_, err = r.queueColonyShip(state, empireID, seatID, command)
		case CommandQueueOutpostShip:
			_, err = r.queueOutpostShip(state, empireID, seatID, command)
		case CommandQueueTroopTransport:
			_, err = r.queueTroopTransport(state, empireID, seatID, command)
		case CommandQueueMilitaryShip:
			_, err = r.queueMilitaryShip(state, empireID, seatID, command)
		case CommandQueueHousing:
			_, err = r.queueHousing(state, empireID, seatID, command)
		case CommandQueuePlanetaryTransformation:
			_, err = r.queuePlanetaryTransformation(state, empireID, seatID, command)
		case CommandQueueFreighterFleet:
			_, err = r.queueFreighterFleet(state, empireID, seatID, command)
		case CommandTransferPopulation:
			_, err = r.transferPopulation(state, empireID, seatID, command)
		case CommandMoveFleet:
			_, err = r.moveFleetEvents(state, empireID, seatID, command)
		case CommandSplitFleet:
			_, err = r.splitFleet(state, empireID, seatID, command)
		case CommandMergeFleets:
			_, err = r.mergeFleets(state, empireID, seatID, command)
		case CommandColonizePlanet:
			_, err = r.colonizePlanet(state, empireID, seatID, command)
		case CommandDeployOutpost:
			_, err = r.deployOutpost(state, empireID, seatID, command)
		case CommandSelectResearch:
			_, err = r.selectResearch(state, empireID, seatID, command)
		default:
			err = fmt.Errorf("unsupported Planning preview command kind %q", command.Kind)
		}
		if err != nil {
			return fmt.Errorf("command %d (%s): %w", command.Sequence, command.Kind, err)
		}
	}
	for i := range state.Colonies {
		if err := r.recalculateColony(state, &state.Colonies[i]); err != nil {
			return fmt.Errorf("recalculate colony %d: %w", state.Colonies[i].ID, err)
		}
	}
	if _, err := r.materializeFoodLogistics(state, false); err != nil {
		return fmt.Errorf("materialize food logistics: %w", err)
	}
	return nil
}

func (r *EconomyResolver) applyPlanningPopulationAssignment(state *core.GameState, empireID core.ID, command protocol.Command) error {
	payload, err := decodeAssignPopulation(command)
	if err != nil {
		return err
	}
	colony := colonyByID(state, payload.ColonyID)
	if colony == nil {
		return fmt.Errorf("references unknown colony %d", payload.ColonyID)
	}
	if colony.EmpireID != empireID {
		return fmt.Errorf("cannot assign population on colony %d owned by empire %d", colony.ID, colony.EmpireID)
	}
	assigned := payload.Farmers + payload.Workers + payload.Scientists
	if math.Abs(assigned-colony.Population.Total()) > 1e-9*math.Max(1, math.Max(math.Abs(assigned), math.Abs(colony.Population.Total()))) {
		return fmt.Errorf("colony %d assignment total %g does not equal population total %g", colony.ID, assigned, colony.Population.Total())
	}
	if err := colony.Population.SetAggregateJobs(payload.Farmers, payload.Workers, payload.Scientists); err != nil {
		return fmt.Errorf("colony %d assignment: %w", colony.ID, err)
	}
	return r.recalculateColony(state, colony)
}

func (r *EconomyResolver) BuildPlanningPreview(state *core.GameState, empireID core.ID) (PlanningPreviewProjection, error) {
	if r == nil || r.Rules == nil {
		return PlanningPreviewProjection{}, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return PlanningPreviewProjection{}, fmt.Errorf("game state must not be nil")
	}
	empireIndex := -1
	for i := range state.Empires {
		if state.Empires[i].ID == empireID {
			empireIndex = i
			break
		}
	}
	if empireIndex < 0 {
		return PlanningPreviewProjection{}, fmt.Errorf("unknown empire %d", empireID)
	}
	for i := range state.Colonies {
		if err := r.recalculateColony(state, &state.Colonies[i]); err != nil {
			return PlanningPreviewProjection{}, err
		}
	}
	if _, err := r.materializeFoodLogistics(state, false); err != nil {
		return PlanningPreviewProjection{}, err
	}
	empire := &state.Empires[empireIndex]
	treasury, err := r.prepareTreasurySettlement(state, empireIndex)
	if err != nil {
		return PlanningPreviewProjection{}, err
	}
	availableFreighters := empire.Freighters - empire.FoodLogistics.FreightersUsed - empire.FoodLogistics.PopulationTransportFreightersReserved
	if availableFreighters < 0 {
		availableFreighters = 0
	}
	preview := PlanningPreviewProjection{
		EmpireID:           empire.ID,
		TreasuryBalanceBC:  empire.Treasury.BalanceBC,
		NetModeledIncomeBC: treasury.Treasury.NetModeledIncomeBC,
		Freighters: PlanningFreighterPreview{
			Total: empire.Freighters, FoodUsed: empire.FoodLogistics.FreightersUsed,
			TransferReserved: empire.FoodLogistics.PopulationTransportFreightersReserved,
			Available:        availableFreighters,
		},
		CommandPoints:       treasury.CommandPoints,
		CommandPointOverage: commandPointOverage(treasury.CommandPoints),
	}

	for i := range state.Colonies {
		colony := &state.Colonies[i]
		if colony.EmpireID != empireID {
			continue
		}
		preview.Research.RPPerTurn += colony.AdjustedEconomy.Research
		cp, err := r.buildPlanningColonyPreview(state, colony)
		if err != nil {
			return PlanningPreviewProjection{}, err
		}
		preview.Colonies = append(preview.Colonies, cp)
	}
	sort.Slice(preview.Colonies, func(i, j int) bool { return preview.Colonies[i].Colony.ID < preview.Colonies[j].Colony.ID })
	if empire.Research != nil {
		preview.Research.TechFieldID = empire.Research.TechFieldID
		preview.Research.SelectionMode = empire.Research.SelectionMode
		preview.Research.TechnologyIDs = append([]int(nil), empire.Research.TechnologyIDs...)
		preview.Research.ProgressRP = empire.Research.ProgressRP
		cost, err := r.Rules.researchFieldCostRP(empire, empire.Research.TechFieldID)
		if err != nil {
			return PlanningPreviewProjection{}, err
		}
		preview.Research.CostRP = cost
		preview.Research.RemainingRP = math.Max(0, cost-empire.Research.ProgressRP)
		preview.Research.ETATurns = turnsAtRate(preview.Research.RemainingRP, preview.Research.RPPerTurn)
	}
	preview.PopulationTransfers, err = r.AvailablePopulationTransferChoices(state, empireID)
	if err != nil {
		return PlanningPreviewProjection{}, err
	}
	return preview, nil
}

func (r *EconomyResolver) buildPlanningColonyPreview(state *core.GameState, colony *core.Colony) (PlanningColonyPreview, error) {
	out := PlanningColonyPreview{
		Colony:                  *colony,
		FreePopulationCapacity:  math.Max(0, colony.PopulationDynamics.Capacity-colony.Population.Total()),
		PopulationGrowthPerTurn: math.Max(0, colony.PopulationDynamics.ProjectedGrowth),
		PopulationLossPerTurn:   math.Max(0, colony.PopulationDynamics.ProjectedStarvation),
	}
	out.NextPopulationETATurns = nextWholePopulationETA(colony.Population.Total(), colony.PopulationDynamics.Capacity, out.PopulationGrowthPerTurn)
	projects := make([]core.ConstructionState, 0, 1+len(colony.ConstructionQueue))
	if colony.Construction != nil {
		projects = append(projects, *colony.Construction)
	}
	projects = append(projects, colony.ConstructionQueue...)
	rate := colony.PopulationDynamics.ProductionAvailable
	cumulative := 0.0
	blocked := false
	for _, project := range projects {
		entry := PlanningConstructionPreview{Project: project}
		if project.ProjectKind == core.ConstructionProjectHousing {
			blocked = true
			out.Construction = append(out.Construction, entry)
			continue
		}
		cost, err := r.constructionProjectCostPP(state, colony, &project)
		if err != nil {
			return PlanningColonyPreview{}, err
		}
		entry.CostPP = cost
		entry.RemainingPP = math.Max(0, cost-project.ProgressPP)
		if !blocked {
			cumulative += entry.RemainingPP
			entry.ETATurns = turnsAtRate(cumulative, rate)
		}
		out.Construction = append(out.Construction, entry)
	}
	return out, nil
}

func turnsAtRate(remaining, rate float64) *int {
	if remaining <= populationEpsilon {
		zero := 0
		return &zero
	}
	if rate <= populationEpsilon {
		return nil
	}
	turns := int(math.Ceil(remaining/rate - populationEpsilon))
	if turns < 1 {
		turns = 1
	}
	return &turns
}

func nextWholePopulationETA(population, capacity, growth float64) *int {
	if growth <= populationEpsilon || population >= capacity-populationEpsilon {
		return nil
	}
	nextWhole := math.Floor(population+populationEpsilon) + 1
	if nextWhole > capacity+populationEpsilon {
		return nil
	}
	return turnsAtRate(nextWhole-population, growth)
}
