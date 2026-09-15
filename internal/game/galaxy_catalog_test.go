package game

import (
	"reflect"
	"testing"
)

func TestGalaxyCatalogMatchesFrozenGate2Contract(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	got, err := GalaxyCatalogFromRules(rules)
	if err != nil {
		t.Fatal(err)
	}
	want := GalaxyCatalog{
		DefaultSizeID: GalaxySizeSmall,
		DefaultAgeID:  GalaxyAgeNormal,
		Sizes: []GalaxySizeProfile{
			{ID: GalaxySizeSmall, StarCount: 20},
			{ID: GalaxySizeMedium, StarCount: 36},
			{ID: GalaxySizeLarge, StarCount: 54},
			{ID: GalaxySizeHuge, StarCount: 71},
		},
		Ages: []GalaxyAgeProfile{
			{ID: GalaxyAgeMineralRich, MineralResourceBias: GalaxyBiasHigher, FoodWorldBias: GalaxyBiasLower},
			{ID: GalaxyAgeNormal, MineralResourceBias: GalaxyBiasBaseline, FoodWorldBias: GalaxyBiasBaseline},
			{ID: GalaxyAgeOrganicRich, MineralResourceBias: GalaxyBiasLower, FoodWorldBias: GalaxyBiasHigher},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("catalog=%+v want=%+v", got, want)
	}
}
