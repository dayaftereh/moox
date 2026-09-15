package game

import (
	"errors"
	"reflect"
	"testing"

	"moox/internal/core"
)

func TestDifficultyProfilesMatchFrozenGate2Contract(t *testing.T) {
	got := DifficultyProfiles()
	want := []DifficultyProfile{
		{ID: core.DifficultyEasy, AIFoodPerFarmerEighths: -2, AIProductionPerWorkerEighths: -4, AIResearchPerScientistEighths: -4, AITaxBCPerPopulationEighths: -2, AICommandDeficitBCPerPointEighths: 88},
		{ID: core.DifficultyNormal, AIFoodPerFarmerEighths: 0, AIProductionPerWorkerEighths: 0, AIResearchPerScientistEighths: 0, AITaxBCPerPopulationEighths: 0, AICommandDeficitBCPerPointEighths: 80},
		{ID: core.DifficultyHard, AIFoodPerFarmerEighths: 2, AIProductionPerWorkerEighths: 4, AIResearchPerScientistEighths: 4, AITaxBCPerPopulationEighths: 2, AICommandDeficitBCPerPointEighths: 72},
		{ID: core.DifficultyVeryHard, AIFoodPerFarmerEighths: 4, AIProductionPerWorkerEighths: 8, AIResearchPerScientistEighths: 8, AITaxBCPerPopulationEighths: 3, AICommandDeficitBCPerPointEighths: 68},
		{ID: core.DifficultyImpossible, AIFoodPerFarmerEighths: 6, AIProductionPerWorkerEighths: 12, AIResearchPerScientistEighths: 12, AITaxBCPerPopulationEighths: 4, AICommandDeficitBCPerPointEighths: 64},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("profiles=\n%+v\nwant=\n%+v", got, want)
	}
	got[0].AIFoodPerFarmerEighths = 999
	again, err := DifficultyProfileFor(core.DifficultyEasy)
	if err != nil {
		t.Fatal(err)
	}
	if again.AIFoodPerFarmerEighths != -2 {
		t.Fatal("DifficultyProfiles returned mutable catalog storage")
	}
}

func TestNewGameDifficultyDefaultsPersistsAndRejectsUnknown(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	settings := canonicalNewGameSettings()
	settings.DifficultyID = ""
	generated, err := rules.NewGame(0x8009, settings)
	if err != nil {
		t.Fatal(err)
	}
	if generated.State.DifficultyID != core.DifficultyNormal || generated.State.EffectiveDifficultyID() != core.DifficultyNormal {
		t.Fatalf("default difficulty=%q effective=%q", generated.State.DifficultyID, generated.State.EffectiveDifficultyID())
	}

	settings.DifficultyID = core.DifficultyVeryHard
	generated, err = rules.NewGame(0x8009, settings)
	if err != nil {
		t.Fatal(err)
	}
	if generated.State.DifficultyID != core.DifficultyVeryHard {
		t.Fatalf("explicit difficulty=%q", generated.State.DifficultyID)
	}

	settings.DifficultyID = core.DifficultyID("nightmare")
	if _, err := rules.NewGame(0x8009, settings); !errors.Is(err, ErrInvalidNewGameSettings) {
		t.Fatalf("unknown difficulty error=%v want ErrInvalidNewGameSettings", err)
	}
}
