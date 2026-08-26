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
		key  string
		want string
	}{
		"en": {"race_traits.option.creative.name", "Creative"},
		"de": {"race_traits.group.population_growth.name", "Bevölkerung"},
		"fr": {"race_traits.option.government_democracy.name", "Démocratie"},
		"es": {"race_traits.group.population_growth.name", "Población"},
		"it": {"race_traits.group.special_abilities.name", "Abilità speciali"},
	}
	for locale, tc := range cases {
		file, err := Load(filepath.Join(root, locale+".json"))
		if err != nil {
			t.Fatalf("%s: %v", locale, err)
		}
		if file.Locale != locale {
			t.Fatalf("%s: locale=%q", locale, file.Locale)
		}
		if len(file.Strings) != 64 {
			t.Fatalf("%s: strings=%d want=64", locale, len(file.Strings))
		}
		got, ok := file.Text(tc.key)
		if !ok || got != tc.want {
			t.Fatalf("%s %s=%q ok=%v want=%q", locale, tc.key, got, ok, tc.want)
		}
	}
}
