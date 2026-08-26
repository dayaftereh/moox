package ruleset

import (
	"path/filepath"
	"runtime"
	"testing"

	"moox/internal/i18n"
)

func TestCommittedRacesLoadAndValidate(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test source path")
	}
	root := filepath.Join(filepath.Dir(currentFile), "..", "..")
	traits, err := LoadRaceTraits(filepath.Join(root, "data", "rulesets", "moo2-1.31", "race_traits.json"))
	if err != nil {
		t.Fatal(err)
	}
	races, err := LoadRaces(filepath.Join(root, "data", "rulesets", "moo2-1.31", "races.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := races.ValidateAgainstTraits(traits); err != nil {
		t.Fatal(err)
	}
	if len(races.Races) != 13 {
		t.Fatalf("races=%d want=13", len(races.Races))
	}
	english, err := i18n.Load(filepath.Join(root, "data", "languages", "en.json"))
	if err != nil {
		t.Fatal(err)
	}
	seenGnolam := false
	seenKlackon := false
	for _, race := range races.Races {
		if _, ok := english.Text(race.NameKey); !ok {
			t.Errorf("missing English name key %q", race.NameKey)
		}
		if race.ID == "gnolam" {
			seenGnolam = true
			if race.DerivedPickTotal != 8 {
				t.Errorf("Gnolam derived picks=%d want=8", race.DerivedPickTotal)
			}
		}
		if race.ID == "klackon" {
			seenKlackon = true
			if race.DerivedPickTotal != 9 {
				t.Errorf("Klackon derived picks=%d want=9", race.DerivedPickTotal)
			}
		}
	}
	if !seenGnolam || !seenKlackon {
		t.Fatal("expected Gnolam and Klackon in committed races")
	}
}
