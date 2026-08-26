package ruleset

import (
	"path/filepath"
	"runtime"
	"testing"

	"moox/internal/i18n"
)

func TestCommittedRaceTraitsLoad(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test source path")
	}
	repoRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	rulesPath := filepath.Join(repoRoot, "data", "rulesets", "moo2-1.31", "race_traits.json")
	languagePath := filepath.Join(repoRoot, "data", "languages", "en.json")

	file, err := LoadRaceTraits(rulesPath)
	if err != nil {
		t.Fatal(err)
	}
	language, err := i18n.Load(languagePath)
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
		if _, ok := language.Text(group.NameKey); !ok {
			t.Errorf("missing English key %q", group.NameKey)
		}
		for _, option := range group.Options {
			if _, ok := language.Text(option.NameKey); !ok {
				t.Errorf("missing English key %q", option.NameKey)
			}
		}
	}
	if options != 53 {
		t.Fatalf("options=%d want=53", options)
	}
}
