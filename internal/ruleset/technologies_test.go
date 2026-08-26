package ruleset

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestCommittedTechnologiesLoadAndValidate(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test source path")
	}
	root := filepath.Join(filepath.Dir(currentFile), "..", "..")
	file, err := LoadTechnologies(filepath.Join(root, "data", "rulesets", "moo2-1.31", "technologies.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Validate(); err != nil {
		t.Fatal(err)
	}
	checks := map[int]string{1: "achilles_targeting_unit", 5: "alien_management_center", 16: "planet_construction", 87: "hydroponic_farm", 203: "zortrium_armor"}
	for id, want := range checks {
		if got := file.Technologies[id-1].ID; got != want {
			t.Fatalf("technology %d id=%q want=%q", id, got, want)
		}
	}
}
