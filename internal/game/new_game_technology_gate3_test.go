package game

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTechnologyLevelCatalogFreezesGate3SupportBoundary(t *testing.T) {
	catalog := TechnologyLevelCatalog()
	if catalog.DefaultID != NewGameTechnologyAverage {
		t.Fatalf("default=%q want=%q", catalog.DefaultID, NewGameTechnologyAverage)
	}
	want := []struct {
		id           NewGameTechnologyLevel
		availability NewGameTechnologyAvailability
	}{
		{NewGameTechnologyPreWarp, NewGameTechnologySupported},
		{NewGameTechnologyAverage, NewGameTechnologySupported},
		{NewGameTechnologyAdvanced, NewGameTechnologyPlanned},
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

func TestNewGamePreWarpAndAverageRaceMatrixDeterministic(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	for _, raceID := range []string{"human", "klackon"} {
		for _, level := range []NewGameTechnologyLevel{NewGameTechnologyPreWarp, NewGameTechnologyAverage} {
			t.Run(raceID+"/"+string(level), func(t *testing.T) {
				settings := canonicalNewGameSettings()
				settings.Players[0].RaceID = raceID
				settings.Players[0].EmpireName = raceID
				settings.TechnologyLevel = level
				first, err := rules.NewGame(0x1653, settings)
				if err != nil {
					t.Fatal(err)
				}
				second, err := rules.NewGame(0x1653, settings)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(first, second) {
					t.Fatal("same race/level/seed produced different states")
				}
				firstJSON, _ := json.Marshal(first.State)
				secondJSON, _ := json.Marshal(second.State)
				if !reflect.DeepEqual(firstJSON, secondJSON) {
					t.Fatal("same race/level/seed produced different JSON")
				}
				if level == NewGameTechnologyPreWarp {
					if len(first.State.Ships) != 0 || len(first.State.ShipDesigns) != 0 || len(first.State.StrategicFleets) != 0 {
						t.Fatalf("Pre-Warp ships/designs/fleets=%d/%d/%d want 0/0/0", len(first.State.Ships), len(first.State.ShipDesigns), len(first.State.StrategicFleets))
					}
					for _, empire := range first.State.Empires {
						if got := len(empire.KnownTechnologyFieldIDs); got != 2 {
							t.Fatalf("Pre-Warp known fields=%d want=2", got)
						}
					}
				} else {
					if len(first.State.Ships) != 4 || len(first.State.ShipDesigns) != 2 || len(first.State.StrategicFleets) != 4 {
						t.Fatalf("Average ships/designs/fleets=%d/%d/%d want 4/2/4", len(first.State.Ships), len(first.State.ShipDesigns), len(first.State.StrategicFleets))
					}
				}
			})
		}
	}
}

func TestNewGameAdvancedRemainsRejected(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	settings := canonicalNewGameSettings()
	settings.TechnologyLevel = NewGameTechnologyAdvanced
	if _, err := rules.NewGame(0x1657, settings); err == nil {
		t.Fatal("Advanced unexpectedly accepted by Gate 3")
	}
}
