package game

import (
	"math"
	"path/filepath"
	"testing"

	"moox/internal/core"
)

func loadPlanningBreakdownTestResolver(t *testing.T) *EconomyResolver {
	t.Helper()
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	return resolver
}

func metricComponentValue(breakdown PlanningMetricBreakdown, id string) float64 {
	for _, component := range breakdown.Components {
		if component.ID == id {
			return component.Value
		}
	}
	return 0
}

func metricComponentSum(breakdown PlanningMetricBreakdown) float64 {
	total := 0.0
	for _, component := range breakdown.Components {
		total += component.Value
	}
	return total
}

func TestPlanningColonyBreakdownsMatchAuthoritativeTotals(t *testing.T) {
	resolver := loadPlanningBreakdownTestResolver(t)
	state := core.NewSmallFixture(0xB2B)
	empireID := state.Empires[0].ID

	preview, err := resolver.BuildPlanningPreview(state, empireID)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Colonies) == 0 {
		t.Fatal("expected colony preview")
	}
	colony := preview.Colonies[0]
	if math.Abs(colony.Breakdowns.Growth.Total-colony.PopulationGrowthPerTurn) > 1e-9 {
		t.Fatalf("growth breakdown total %.6f != preview growth %.6f", colony.Breakdowns.Growth.Total, colony.PopulationGrowthPerTurn)
	}
	if math.Abs(metricComponentSum(colony.Breakdowns.Growth)-colony.Breakdowns.Growth.Total) > 1e-9 {
		t.Fatalf("growth components do not sum to total: %+v", colony.Breakdowns.Growth)
	}
	if metricComponentValue(colony.Breakdowns.Growth, "natural") <= 0 {
		t.Fatalf("expected positive natural growth component: %+v", colony.Breakdowns.Growth)
	}
	if math.Abs(colony.Breakdowns.TaxBC.Total-colony.Colony.AdjustedEconomy.TaxBC) > 1e-9 {
		t.Fatalf("tax breakdown total %.6f != adjusted tax %.6f", colony.Breakdowns.TaxBC.Total, colony.Colony.AdjustedEconomy.TaxBC)
	}
	if math.Abs(metricComponentSum(colony.Breakdowns.TaxBC)-colony.Breakdowns.TaxBC.Total) > 1e-9 {
		t.Fatalf("tax components do not sum to total: %+v", colony.Breakdowns.TaxBC)
	}
	if got := metricComponentValue(colony.Breakdowns.TaxBC, "population_tax"); math.Abs(got-4) > 1e-9 {
		t.Fatalf("population tax = %.3f, want 4", got)
	}
	if got := metricComponentValue(colony.Breakdowns.TaxBC, "government"); math.Abs(got-2) > 1e-9 {
		t.Fatalf("government tax contribution = %.3f, want 2", got)
	}
}

func TestPlanningTaxBCBreakdownReflectsMoraleBuilding(t *testing.T) {
	resolver := loadPlanningBreakdownTestResolver(t)
	state := core.NewSmallFixture(0xB2F)
	state.Colonies[0].Buildings = append(state.Colonies[0].Buildings, "holo_simulator")
	preview, err := resolver.BuildPlanningPreview(state, state.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	breakdown := preview.Colonies[0].Breakdowns.TaxBC
	if got := metricComponentValue(breakdown, "morale"); got <= 0 {
		t.Fatalf("expected positive morale tax component with Holo Simulator: %+v", breakdown)
	}
	if math.Abs(metricComponentSum(breakdown)-breakdown.Total) > 1e-9 {
		t.Fatalf("tax components do not sum to total: %+v", breakdown)
	}
}
func TestPlanningGrowthBreakdownReflectsHousingAndCloningCenter(t *testing.T) {
	resolver := loadPlanningBreakdownTestResolver(t)
	baselineState := core.NewSmallFixture(0xB2C)
	empireID := baselineState.Empires[0].ID
	baselinePreview, err := resolver.BuildPlanningPreview(baselineState, empireID)
	if err != nil {
		t.Fatal(err)
	}
	baselineGrowth := baselinePreview.Colonies[0].Breakdowns.Growth.Total

	housingState := core.NewSmallFixture(0xB2D)
	housingState.Colonies[0].Construction = &core.ConstructionState{ProjectKind: core.ConstructionProjectHousing, ProjectID: HousingProjectID}
	housingPreview, err := resolver.BuildPlanningPreview(housingState, housingState.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	housingBreakdown := housingPreview.Colonies[0].Breakdowns.Growth
	if got := metricComponentValue(housingBreakdown, "housing"); got <= 0 {
		t.Fatalf("expected positive Housing component: %+v", housingBreakdown)
	}
	if housingBreakdown.Total <= baselineGrowth {
		t.Fatalf("housing growth %.6f must exceed baseline %.6f", housingBreakdown.Total, baselineGrowth)
	}

	cloningState := core.NewSmallFixture(0xB2E)
	cloningState.Colonies[0].Buildings = append(cloningState.Colonies[0].Buildings, resolver.Rules.CloningCenterBuildingID)
	cloningPreview, err := resolver.BuildPlanningPreview(cloningState, cloningState.Empires[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	cloningBreakdown := cloningPreview.Colonies[0].Breakdowns.Growth
	if got := metricComponentValue(cloningBreakdown, "cloning_center"); got <= 0 {
		t.Fatalf("expected positive Cloning Center component: %+v", cloningBreakdown)
	}
	if cloningBreakdown.Total <= baselineGrowth {
		t.Fatalf("cloning growth %.6f must exceed baseline %.6f", cloningBreakdown.Total, baselineGrowth)
	}
}
