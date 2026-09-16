package server

import (
	"testing"

	"moox/internal/app"
	"moox/internal/game"
)

func TestHTTPRaceCatalogMatchesFrozenServerContract(t *testing.T) {
	server, _ := newNewGameServer(t)
	defer server.Close()

	var catalog raceCatalogResponse
	getJSON(t, server.URL+"/api/v1/new-game/races", &catalog)
	if catalog.SchemaVersion != app.SchemaVersion {
		t.Fatalf("schema_version=%d want=%d", catalog.SchemaVersion, app.SchemaVersion)
	}
	if catalog.DefaultPlayerRaceID != game.DefaultPlayerRaceID || catalog.FixedOpponentRaceID != game.FixedOpponentRaceID {
		t.Fatalf("race defaults=%q/%q", catalog.DefaultPlayerRaceID, catalog.FixedOpponentRaceID)
	}
	if len(catalog.Profiles) != 13 {
		t.Fatalf("profiles=%d want=13", len(catalog.Profiles))
	}
	supported := map[string]bool{}
	for i, profile := range catalog.Profiles {
		if profile.Order != i {
			t.Fatalf("profile[%d] order=%d", i, profile.Order)
		}
		if profile.PlayerAvailability == game.PresetRaceAvailabilitySupported {
			supported[profile.ID] = true
		}
	}
	if len(supported) != 2 || !supported["human"] || !supported["klackon"] {
		t.Fatalf("supported races=%v want human+klackon", supported)
	}
}
