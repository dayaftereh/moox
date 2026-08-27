package moo2data

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodePlanetClassesWithVerifiedEvidence(t *testing.T) {
	root := t.TempDir()
	values := make([]string, 0x32C)
	for i := range values {
		values[i] = "unused"
	}
	for _, spec := range append(append(append([]planetClassSpec{}, planetSizeSpecs...), mineralClassSpecs...), append(gravityClassSpecs, planetClimateSpecs...)...) {
		values[spec.EStringIndex] = spec.Name
	}
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint16(payload[0:2], 1)
	for _, value := range values {
		payload = append(payload, []byte(value)...)
		payload = append(payload, 0)
	}
	binary.LittleEndian.PutUint16(payload[2:4], uint16(len(payload)-4))
	if err := os.WriteFile(filepath.Join(root, "ESTRINGS.LBX"), buildAssetTestLBX([][]byte{payload}), 0o644); err != nil {
		t.Fatal(err)
	}

	evidence := planetClassEvidence{
		ExecutableSHA256:  "synthetic-orion2",
		SizeThresholds:    []int{1, 3, 7, 9, 10},
		MineralExtraction: []int{1, 2, 3, 4, 5},
		FoodPerFarmer:     []int{0, 0, 0, 1, 1, 2, 2, 1, 2, 3},
	}
	bundle, err := decodePlanetClassesWithEvidence(root, evidence)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(bundle.Rules.Sizes); got != 5 {
		t.Fatalf("sizes=%d", got)
	}
	if got := len(bundle.Rules.MineralClasses); got != 5 {
		t.Fatalf("minerals=%d", got)
	}
	if got := len(bundle.Rules.GravityClasses); got != 3 {
		t.Fatalf("gravity=%d", got)
	}
	if got := len(bundle.Rules.Climates); got != 10 {
		t.Fatalf("climates=%d", got)
	}
	if got := bundle.Rules.Sizes[2].GenerationRollUpperThreshold; got != 7 {
		t.Fatalf("medium threshold=%d", got)
	}
	if got := bundle.Rules.MineralClasses[4].BaseExtraction; got != 5 {
		t.Fatalf("ultra rich extraction=%d", got)
	}
	if got := bundle.Rules.Climates[9].BaseFoodPerFarmer; got != 3 {
		t.Fatalf("gaia food=%d", got)
	}
	if got := bundle.English.Strings["planet_climate.terran.name"]; got != "Terran" {
		t.Fatalf("terran name=%q", got)
	}
	if got := len(bundle.English.Strings); got != 23 {
		t.Fatalf("English strings=%d", got)
	}
}

func TestParseEStringsBlockRejectsMalformedPayload(t *testing.T) {
	_, err := parseEStringsBlock([]byte{1, 0, 5, 0, 'x', 0})
	if err == nil {
		t.Fatal("expected payload-size error")
	}
}
