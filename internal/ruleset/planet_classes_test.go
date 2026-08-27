package ruleset

import "testing"

func TestPlanetClassesValidate(t *testing.T) {
	file := &PlanetClassesFile{
		SchemaVersion: PlanetClassesSchemaVersion,
		Ruleset:       "moo2-1.31",
	}
	for i, threshold := range []int{1, 3, 7, 9, 10} {
		file.Sizes = append(file.Sizes, PlanetSize{ID: []string{"tiny", "small", "medium", "large", "huge"}[i], Index: i, NameKey: "size", NameSource: FieldProvenance{SourceID: "names"}, GenerationRollUpperThreshold: threshold, GenerationSource: FieldProvenance{SourceID: "size"}})
	}
	for i, id := range []string{"ultra_poor", "poor", "abundant", "rich", "ultra_rich"} {
		file.MineralClasses = append(file.MineralClasses, MineralClass{ID: id, Index: i, NameKey: "mineral", NameSource: FieldProvenance{SourceID: "names"}, BaseExtraction: i + 1, ExtractionSource: FieldProvenance{SourceID: "mineral"}})
	}
	for i, id := range []string{"low_g", "normal_g", "heavy_g"} {
		file.GravityClasses = append(file.GravityClasses, GravityClass{ID: id, Index: i, NameKey: "gravity", NameSource: FieldProvenance{SourceID: "names"}})
	}
	for i, id := range []string{"toxic", "radiated", "barren", "desert", "tundra", "ocean", "swamp", "arid", "terran", "gaia"} {
		food := []int{0, 0, 0, 1, 1, 2, 2, 1, 2, 3}[i]
		file.Climates = append(file.Climates, PlanetClimate{ID: id, Index: i, NameKey: "climate", NameSource: FieldProvenance{SourceID: "names"}, BaseFoodPerFarmer: food, FoodSource: FieldProvenance{SourceID: "food"}})
	}
	if err := file.Validate(); err != nil {
		t.Fatal(err)
	}
	file.Sizes[2].GenerationRollUpperThreshold = 2
	if err := file.Validate(); err == nil {
		t.Fatal("expected invalid threshold to fail")
	}
}
