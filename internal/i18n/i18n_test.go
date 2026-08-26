package i18n

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestCommittedEnglishLanguageLoads(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test source path")
	}
	path := filepath.Join(filepath.Dir(currentFile), "..", "..", "data", "languages", "en.json")
	file, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if file.Locale != "en" {
		t.Fatalf("locale=%q", file.Locale)
	}
	if len(file.Strings) != 64 {
		t.Fatalf("strings=%d want=64", len(file.Strings))
	}
	got, ok := file.Text("race_traits.option.creative.name")
	if !ok || got != "Creative" {
		t.Fatalf("creative name=%q ok=%v", got, ok)
	}
}
