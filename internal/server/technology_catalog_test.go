package server

import (
	"testing"

	"moox/internal/app"
	"moox/internal/game"
)

func TestHTTPTechnologyCatalogMatchesFrozenGate3Contract(t *testing.T) {
	server, _ := newNewGameServer(t)
	defer server.Close()
	var catalog technologyCatalogResponse
	getJSON(t, server.URL+"/api/v1/new-game/technologies", &catalog)
	if catalog.SchemaVersion != app.SchemaVersion {
		t.Fatalf("schema_version=%d want=%d", catalog.SchemaVersion, app.SchemaVersion)
	}
	if catalog.DefaultID != game.NewGameTechnologyAverage {
		t.Fatalf("default_id=%q want=%q", catalog.DefaultID, game.NewGameTechnologyAverage)
	}
	want := []struct {
		id           game.NewGameTechnologyLevel
		availability game.NewGameTechnologyAvailability
	}{
		{game.NewGameTechnologyPreWarp, game.NewGameTechnologySupported},
		{game.NewGameTechnologyAverage, game.NewGameTechnologySupported},
		{game.NewGameTechnologyAdvanced, game.NewGameTechnologyPlanned},
	}
	if len(catalog.Profiles) != len(want) {
		t.Fatalf("profiles=%d want=%d", len(catalog.Profiles), len(want))
	}
	for i, expected := range want {
		profile := catalog.Profiles[i]
		if profile.ID != expected.id || profile.Availability != expected.availability || len(profile.Facts) < 3 {
			t.Fatalf("profile[%d]=%+v want id=%q availability=%q", i, profile, expected.id, expected.availability)
		}
	}
}
