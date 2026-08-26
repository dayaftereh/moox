package ruleset

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestBuildingsValidate(t *testing.T) {
	file := &BuildingsFile{SchemaVersion: BuildingsSchemaVersion, Ruleset: "moo2-1.31"}
	for i := 0; i < 48; i++ {
		offset := i * 100
		file.Buildings = append(file.Buildings, Building{
			ID:                       "building_id_" + string(rune('a'+i%26)) + string(rune('A'+i/26)),
			Order:                    i,
			ProductionID:             i + 1,
			ProductionIDVerification: "test",
			ProductionIDSource:       FieldProvenance{SourceID: "secondary"},
			NameKey:                  "building.name." + string(rune('a'+i%26)) + string(rune('A'+i/26)),
			NameVerification:         "test",
			NameSource:               FieldProvenance{SourceID: "help", Offset: &offset},
		})
	}
	if err := file.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestCommittedBuildingsLoadAndValidate(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test source path")
	}
	root := filepath.Join(filepath.Dir(currentFile), "..", "..")
	file, err := LoadBuildings(filepath.Join(root, "data", "rulesets", "moo2-1.31", "buildings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Validate(); err != nil {
		t.Fatal(err)
	}
	if len(file.Buildings) != 48 {
		t.Fatalf("buildings=%d want=48", len(file.Buildings))
	}
	if file.Buildings[0].ID != "alien_management_center" || file.Buildings[47].ID != "artificial_planet" {
		t.Fatalf("unexpected building endpoints: %s .. %s", file.Buildings[0].ID, file.Buildings[47].ID)
	}
}
