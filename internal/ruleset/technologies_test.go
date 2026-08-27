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
	if file.NewGameStart.AlwaysKnownTechFieldID != 0 {
		t.Fatalf("always-known tech field=%d want=0", file.NewGameStart.AlwaysKnownTechFieldID)
	}
	wantStartFields := []int{29, 55, 22, 57, 28, 23}
	if len(file.NewGameStart.StagedKnownTechFieldIDs) != len(wantStartFields) {
		t.Fatalf("staged start fields=%v", file.NewGameStart.StagedKnownTechFieldIDs)
	}
	for i, want := range wantStartFields {
		if got := file.NewGameStart.StagedKnownTechFieldIDs[i]; got != want {
			t.Fatalf("staged start field[%d]=%d want=%d", i, got, want)
		}
	}
	if len(file.Fields) != 82 {
		t.Fatalf("technology fields=%d want=82", len(file.Fields))
	}
	field56 := file.Fields[55]
	if field56.FieldID != 56 || field56.ResearchCost != 150 || field56.PreviousID != 28 || field56.NextID != 15 {
		t.Fatalf("unexpected field 56: %+v", field56)
	}
	if file.Technologies[62].TechnologyID != 63 || file.Technologies[62].TechFieldID != 22 || file.Technologies[62].StrategicCombatAvailable {
		t.Fatalf("unexpected Extended Fuel Tanks metadata: %+v", file.Technologies[62])
	}
	if file.Technologies[120].TechnologyID != 121 || file.Technologies[120].TechFieldID != 22 || !file.Technologies[120].StrategicCombatAvailable {
		t.Fatalf("unexpected Nuclear Missile metadata: %+v", file.Technologies[120])
	}
	for id, want := range checks {
		if got := file.Technologies[id-1].ID; got != want {
			t.Fatalf("technology %d id=%q want=%q", id, got, want)
		}
	}
}
