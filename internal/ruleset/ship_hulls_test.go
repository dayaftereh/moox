package ruleset

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestCommittedShipHullsLoadAndValidate(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test source path")
	}
	root := filepath.Join(filepath.Dir(currentFile), "..", "..")
	file, err := LoadShipHulls(filepath.Join(root, "data", "rulesets", "moo2-1.31", "ship_hulls.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Validate(); err != nil {
		t.Fatal(err)
	}
	if file.Hulls[0].ID != "frigate" || file.Hulls[0].StrategicPictureIDs[0] != 0 || file.Hulls[0].StrategicPictureIDs[7] != 7 {
		t.Fatalf("unexpected frigate: %+v", file.Hulls[0])
	}
	if file.Hulls[5].ID != "doom_star" || len(file.Hulls[5].StrategicPictureIDs) != 1 || file.Hulls[5].StrategicPictureIDs[0] != 43 || file.Hulls[5].TacticalAssetKey != "ship_hull.doom_star.tactical" {
		t.Fatalf("unexpected doom star: %+v", file.Hulls[5])
	}
}
