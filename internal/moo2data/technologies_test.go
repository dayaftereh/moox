package moo2data

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeTechnologiesUsesOriginalBoundedSequence(t *testing.T) {
	root := t.TempDir()
	values := []string{"Starting Tech", "No Tech", "Achilles Targeting Unit"}
	for i := 2; i < 203; i++ {
		values = append(values, fmt.Sprintf("Synthetic Tech %03d", i+1))
	}
	values = append(values, "Zortrium Armor", "Biology", "Power")
	block := nullSeparated(values)
	if err := os.WriteFile(filepath.Join(root, "TECHNAME.LBX"), buildAssetTestLBX([][]byte{block}), 0o644); err != nil {
		t.Fatal(err)
	}

	bundle, err := DecodeTechnologies(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Rules.Technologies) != 203 || len(bundle.English.Strings) != 203 {
		t.Fatalf("technologies=%d strings=%d", len(bundle.Rules.Technologies), len(bundle.English.Strings))
	}
	if got := bundle.Rules.Technologies[0]; got.TechnologyID != 1 || got.ID != "achilles_targeting_unit" {
		t.Fatalf("first=%+v", got)
	}
	if got := bundle.Rules.Technologies[202]; got.TechnologyID != 203 || got.ID != "zortrium_armor" {
		t.Fatalf("last=%+v", got)
	}
}

func TestStableTechnologyID(t *testing.T) {
	cases := map[string]string{
		"Anti-Matter Drive":     "anti_matter_drive",
		"Robo-Miners":           "robo_miners",
		"Class VII Shield":      "class_vii_shield",
		"Multi-Wave ECM Jammer": "multi_wave_ecm_jammer",
	}
	for input, want := range cases {
		if got := stableTechnologyID(input); got != want {
			t.Fatalf("stableTechnologyID(%q)=%q want=%q", input, got, want)
		}
	}
}

func nullSeparated(values []string) []byte {
	var data []byte
	for _, value := range values {
		data = append(data, []byte(value)...)
		data = append(data, 0)
	}
	return data
}
