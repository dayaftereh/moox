package i18n

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestCommittedRaceTraitLanguagesLoad(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test source path")
	}
	root := filepath.Join(filepath.Dir(currentFile), "..", "..", "data", "languages")
	cases := map[string]struct {
		key       string
		want      string
		wantCount int
	}{
		"en": {"ship_hull.doom_star.name", "Doom Star", 334},
		"de": {"race_traits.group.population_growth.name", "Bev\u00f6lkerung", 64},
		"fr": {"race_traits.option.government_democracy.name", "D\u00e9mocratie", 64},
		"es": {"race_traits.group.population_growth.name", "Poblaci\u00f3n", 64},
		"it": {"race_traits.group.special_abilities.name", "Abilit\u00e0 speciali", 64},
	}
	for locale, tc := range cases {
		file, err := Load(filepath.Join(root, locale+".json"))
		if err != nil {
			t.Fatalf("%s: %v", locale, err)
		}
		if file.Locale != locale {
			t.Fatalf("%s: locale=%q", locale, file.Locale)
		}
		if len(file.Strings) != tc.wantCount {
			t.Fatalf("%s: strings=%d want=%d", locale, len(file.Strings), tc.wantCount)
		}
		got, ok := file.Text(tc.key)
		if !ok || got != tc.want {
			t.Fatalf("%s %s=%q ok=%v want=%q", locale, tc.key, got, ok, tc.want)
		}
	}
}

func TestMerge(t *testing.T) {
	base := &File{SchemaVersion: SchemaVersion, Locale: "en", Strings: map[string]string{"a": "A"}, Sources: []Source{{ID: "one"}}}
	fragment := &File{SchemaVersion: SchemaVersion, Locale: "en", Strings: map[string]string{"b": "B"}, Sources: []Source{{ID: "two"}}}
	if err := base.Merge(fragment); err != nil {
		t.Fatal(err)
	}
	if base.Strings["b"] != "B" || len(base.Sources) != 2 {
		t.Fatalf("merged=%+v", base)
	}
	conflict := &File{SchemaVersion: SchemaVersion, Locale: "en", Strings: map[string]string{"a": "other"}}
	if err := base.Merge(conflict); err == nil {
		t.Fatal("expected conflicting key to fail")
	}
}
