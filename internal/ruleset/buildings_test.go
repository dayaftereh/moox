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
			ID:                         "building_id_" + string(rune('a'+i%26)) + string(rune('A'+i/26)),
			Order:                      i,
			ProductionID:               i + 1,
			ProductionIDVerification:   "test",
			ProductionIDSource:         FieldProvenance{SourceID: "secondary"},
			TechnologyID:               i + 1,
			TechnologyKey:              "technology_id_" + string(rune('a'+i%26)) + string(rune('A'+i/26)),
			TechnologyLinkVerification: "test",
			TechnologySource:           FieldProvenance{SourceID: "technology"},
			NameKey:                    "building.name." + string(rune('a'+i%26)) + string(rune('A'+i/26)),
			NameVerification:           "test",
			NameSource:                 FieldProvenance{SourceID: "help", Offset: &offset},
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
	technologies, err := LoadTechnologies(filepath.Join(root, "data", "rulesets", "moo2-1.31", "technologies.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := file.ValidateAgainstTechnologies(technologies); err != nil {
		t.Fatal(err)
	}
	if len(file.Buildings) != 48 {
		t.Fatalf("buildings=%d want=48", len(file.Buildings))
	}
	if file.Buildings[0].ID != "alien_management_center" || file.Buildings[0].TechnologyID != 5 || file.Buildings[0].TechnologyKey != "alien_management_center" {
		t.Fatalf("unexpected first building: %+v", file.Buildings[0])
	}
	if file.Buildings[20].TechnologyID != 87 || file.Buildings[20].TechnologyKey != "hydroponic_farm" {
		t.Fatalf("unexpected hydroponic farms link: %+v", file.Buildings[20])
	}
	if file.Buildings[47].ID != "artificial_planet" || file.Buildings[47].TechnologyID != 16 || file.Buildings[47].TechnologyKey != "planet_construction" {
		t.Fatalf("unexpected last building: %+v", file.Buildings[47])
	}
}
