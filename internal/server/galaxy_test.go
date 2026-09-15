package server

import (
	"testing"

	"moox/internal/app"
	"moox/internal/game"
)

func TestHTTPGalaxyCatalogMatchesFrozenServerContract(t *testing.T) {
	server, _ := newNewGameServer(t)
	defer server.Close()

	var catalog galaxyCatalogResponse
	getJSON(t, server.URL+"/api/v1/new-game/galaxy", &catalog)
	if catalog.SchemaVersion != app.SchemaVersion {
		t.Fatalf("schema_version=%d want=%d", catalog.SchemaVersion, app.SchemaVersion)
	}
	if catalog.DefaultSizeID != game.GalaxySizeSmall {
		t.Fatalf("default_size_id=%q want=%q", catalog.DefaultSizeID, game.GalaxySizeSmall)
	}
	if catalog.DefaultAgeID != game.GalaxyAgeNormal {
		t.Fatalf("default_age_id=%q want=%q", catalog.DefaultAgeID, game.GalaxyAgeNormal)
	}
	wantSizes := []game.GalaxySizeProfile{
		{ID: game.GalaxySizeSmall, StarCount: 20},
		{ID: game.GalaxySizeMedium, StarCount: 36},
		{ID: game.GalaxySizeLarge, StarCount: 54},
		{ID: game.GalaxySizeHuge, StarCount: 71},
	}
	if len(catalog.Sizes) != len(wantSizes) {
		t.Fatalf("sizes=%d want=%d", len(catalog.Sizes), len(wantSizes))
	}
	for i, want := range wantSizes {
		if catalog.Sizes[i] != want {
			t.Fatalf("size[%d]=%+v want=%+v", i, catalog.Sizes[i], want)
		}
	}
	wantAges := []game.GalaxyAgeProfile{
		{ID: game.GalaxyAgeMineralRich, MineralResourceBias: game.GalaxyBiasHigher, FoodWorldBias: game.GalaxyBiasLower},
		{ID: game.GalaxyAgeNormal, MineralResourceBias: game.GalaxyBiasBaseline, FoodWorldBias: game.GalaxyBiasBaseline},
		{ID: game.GalaxyAgeOrganicRich, MineralResourceBias: game.GalaxyBiasLower, FoodWorldBias: game.GalaxyBiasHigher},
	}
	if len(catalog.Ages) != len(wantAges) {
		t.Fatalf("ages=%d want=%d", len(catalog.Ages), len(wantAges))
	}
	for i, want := range wantAges {
		if catalog.Ages[i] != want {
			t.Fatalf("age[%d]=%+v want=%+v", i, catalog.Ages[i], want)
		}
	}
}
