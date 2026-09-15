package server

import (
	"testing"

	"moox/internal/app"
	"moox/internal/core"
)

func TestHTTPDifficultyCatalogMatchesFrozenServerContract(t *testing.T) {
	server, _ := newNewGameServer(t)
	defer server.Close()

	var catalog difficultyCatalogResponse
	getJSON(t, server.URL+"/api/v1/new-game/difficulties", &catalog)
	if catalog.SchemaVersion != app.SchemaVersion {
		t.Fatalf("schema_version=%d want=%d", catalog.SchemaVersion, app.SchemaVersion)
	}
	if catalog.DefaultID != core.DifficultyNormal {
		t.Fatalf("default_id=%q want=%q", catalog.DefaultID, core.DifficultyNormal)
	}
	if len(catalog.Profiles) != 5 {
		t.Fatalf("profiles=%d want=5", len(catalog.Profiles))
	}
	wantIDs := []core.DifficultyID{core.DifficultyEasy, core.DifficultyNormal, core.DifficultyHard, core.DifficultyVeryHard, core.DifficultyImpossible}
	for i, want := range wantIDs {
		if catalog.Profiles[i].ID != want {
			t.Fatalf("profile[%d].id=%q want=%q", i, catalog.Profiles[i].ID, want)
		}
	}
}
