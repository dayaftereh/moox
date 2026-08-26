package moo2data

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeBuildingsFindsOriginalHelpHeadings(t *testing.T) {
	root := t.TempDir()
	specs := buildingSpecs()
	block := syntheticHelpRecords(t, specs)
	if err := os.WriteFile(filepath.Join(root, "HELP.LBX"), buildAssetTestLBX([][]byte{block}), 0o644); err != nil {
		t.Fatal(err)
	}
	techNames := []byte("Pollution Processor\x00Pollution Processor\x00Artificial Planet\x00")
	if err := os.WriteFile(filepath.Join(root, "TECHNAME.LBX"), buildAssetTestLBX([][]byte{techNames}), 0o644); err != nil {
		t.Fatal(err)
	}

	bundle, err := DecodeBuildings(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Rules.Buildings) != 48 || len(bundle.English.Strings) != 48 {
		t.Fatalf("buildings=%d strings=%d", len(bundle.Rules.Buildings), len(bundle.English.Strings))
	}
	first := bundle.Rules.Buildings[0]
	if first.ID != "alien_management_center" || first.ProductionID != 1 || first.Order != 0 || first.ColonyReferenceAssetKey != "building.alien_management_center.colony_reference" {
		t.Fatalf("first=%+v", first)
	}
	last := bundle.Rules.Buildings[47]
	if last.ID != "artificial_planet" || last.ProductionID != 48 || last.Order != 47 {
		t.Fatalf("last=%+v", last)
	}
	if got := bundle.English.Strings["building.automated_factories.name"]; got != "Automated Factories" {
		t.Fatalf("automated factories name=%q", got)
	}
	if got := bundle.Rules.Buildings[31].NameVerification; got != "original-techname-string-2-occurrences" {
		t.Fatalf("pollution processor verification=%q", got)
	}
	if got := bundle.Rules.Buildings[47].NameVerification; got != "original-techname-string" {
		t.Fatalf("artificial planet verification=%q", got)
	}
}

func syntheticHelpRecords(t *testing.T, specs []buildingSpec) []byte {
	t.Helper()
	const recordSize = 128
	data := make([]byte, 4+len(specs)*recordSize)
	binary.LittleEndian.PutUint16(data[0:2], uint16(len(specs)))
	binary.LittleEndian.PutUint16(data[2:4], recordSize)
	for i, spec := range specs {
		off := 4 + i*recordSize
		heading := spec.Name
		if spec.ID == "pollution_processor" || spec.ID == "artificial_planet" {
			heading = fmt.Sprintf("Synthetic placeholder %d", i)
		}
		if len(heading)+1 >= recordSize {
			t.Fatalf("heading too long: %s", heading)
		}
		copy(data[off:], heading)
		copy(data[off+len(heading)+1:], "Synthetic building help description")
	}
	return data
}
