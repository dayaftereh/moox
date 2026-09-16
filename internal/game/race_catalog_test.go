package game

import (
	"reflect"
	"testing"
)

func TestPresetRaceCatalogMatchesFrozenGate2Contract(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	catalog, err := PresetRaceCatalogFromRules(rules)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.DefaultPlayerRaceID != "human" || catalog.FixedOpponentRaceID != "darlok" {
		t.Fatalf("defaults=%q/%q", catalog.DefaultPlayerRaceID, catalog.FixedOpponentRaceID)
	}
	if len(catalog.Profiles) != 13 {
		t.Fatalf("profiles=%d want=13", len(catalog.Profiles))
	}
	gotIDs := make([]string, 0, len(catalog.Profiles))
	for _, profile := range catalog.Profiles {
		gotIDs = append(gotIDs, profile.ID)
		wantAvailability := PresetRaceAvailabilityPlanned
		if profile.ID == "human" || profile.ID == "klackon" {
			wantAvailability = PresetRaceAvailabilitySupported
		}
		if profile.PlayerAvailability != wantAvailability {
			t.Fatalf("race %q availability=%q want=%q", profile.ID, profile.PlayerAvailability, wantAvailability)
		}
		if len(profile.TraitIDs) == 0 || len(profile.CardFactTraitIDs) == 0 {
			t.Fatalf("race %q missing trait/card facts", profile.ID)
		}
	}
	if !reflect.DeepEqual(gotIDs, frozenPresetRaceOrder) {
		t.Fatalf("race order=%v want=%v", gotIDs, frozenPresetRaceOrder)
	}
}

func TestNewGameAllowsKlackonAgainstDarlokDeterministically(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	settings := canonicalNewGameSettings()
	settings.Players[0].RaceID = "klackon"
	settings.Players[0].EmpireName = "Klackon"

	first, err := rules.NewGame(0x8009, settings)
	if err != nil {
		t.Fatal(err)
	}
	second, err := rules.NewGame(0x8009, settings)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same Klackon seed/settings did not produce identical NewGameResult")
	}
}

func TestNewGameRejectsCatalogVisiblePlannedPlayerRace(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	settings := canonicalNewGameSettings()
	settings.Players[0].RaceID = "alkari"
	settings.Players[0].EmpireName = "Alkari"
	if _, err := rules.NewGame(0x8009, settings); err == nil {
		t.Fatal("planned Alkari player race unexpectedly accepted")
	}
}
