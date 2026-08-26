package supportcatalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCopiesNonLBXOnly(t *testing.T) {
	source := t.TempDir()
	out := filepath.Join(t.TempDir(), "out")
	if err := os.WriteFile(filepath.Join(source, "GAME.EXE"), []byte("exe"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "DATA.LBX"), []byte("ignore"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(source, "SB16"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "SB16", "MDI.INI"), []byte("cfg"), 0o644); err != nil {
		t.Fatal(err)
	}

	manifest, err := Build(source, out, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if manifest.FileCount != 2 || manifest.TotalBytes != 6 {
		t.Fatalf("manifest=%+v", manifest)
	}
	for _, record := range manifest.Files {
		if err := VerifyCopy(out, record); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(filepath.Join(out, "files", "DATA.LBX")); !os.IsNotExist(err) {
		t.Fatalf("LBX should not be copied, stat err=%v", err)
	}
}
