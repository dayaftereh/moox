package ruleset

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestCommittedRaceTraitsLoad(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test source path")
	}
	path := filepath.Join(filepath.Dir(currentFile), "..", "..", "data", "rulesets", "moo2-1.31", "race_traits.json")
	file, err := LoadRaceTraits(path)
	if err != nil {
		t.Fatal(err)
	}
	if file.Ruleset != "moo2-1.31" {
		t.Fatalf("ruleset=%q", file.Ruleset)
	}
	if len(file.Groups) != 11 {
		t.Fatalf("groups=%d want=11", len(file.Groups))
	}
	options := 0
	for _, group := range file.Groups {
		options += len(group.Options)
	}
	if options != 53 {
		t.Fatalf("options=%d want=53", options)
	}
}
