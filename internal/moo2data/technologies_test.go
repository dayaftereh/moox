package moo2data

import (
	"encoding/binary"
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
	writeSyntheticTechnologyExe(t, filepath.Join(root, "Orion2.exe"))

	bundle, err := DecodeTechnologies(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Rules.Technologies) != 203 || len(bundle.English.Strings) != 203 || len(bundle.Rules.Fields) != 82 {
		t.Fatalf("technologies=%d strings=%d fields=%d", len(bundle.Rules.Technologies), len(bundle.English.Strings), len(bundle.Rules.Fields))
	}
	if got := bundle.Rules.Technologies[0]; got.TechnologyID != 1 || got.ID != "achilles_targeting_unit" || got.TechFieldID != 0 || !got.StrategicCombatAvailable {
		t.Fatalf("first=%+v", got)
	}
	if got := bundle.Rules.Technologies[202]; got.TechnologyID != 203 || got.ID != "zortrium_armor" || got.TechFieldID != 36 {
		t.Fatalf("last=%+v", got)
	}
	wantStart := []int{29, 55, 22, 57, 28, 23}
	for i, want := range wantStart {
		if got := bundle.Rules.NewGameStart.StagedKnownTechFieldIDs[i]; got != want {
			t.Fatalf("start field[%d]=%d want=%d", i, got, want)
		}
	}
}

func writeSyntheticTechnologyExe(t *testing.T, path string) {
	t.Helper()
	data := make([]byte, technologyAIFieldGroupOffset+technologyAIFieldGroupCount*4)
	for i := 0; i < technologyFieldCount; i++ {
		offset := technologyFieldOffset + i*technologyFieldSize
		fieldID := i + 1
		if fieldID > 1 {
			binary.LittleEndian.PutUint16(data[offset:offset+2], uint16(fieldID-1))
		}
		if fieldID < technologyFieldCount {
			binary.LittleEndian.PutUint16(data[offset+2:offset+4], uint16(fieldID+1))
			binary.LittleEndian.PutUint16(data[offset+21:offset+23], uint16(fieldID+1))
		}
		binary.LittleEndian.PutUint32(data[offset+12:offset+16], uint32(50+fieldID))
		data[offset+16] = byte(fieldID % 23)
	}
	for classID := 0; classID < technologyAIClassCount; classID++ {
		offset := technologyAIClassTableOffset + classID*2
		data[offset] = byte(5 + classID%10)
		data[offset+1] = byte(classID % 2)
	}
	for i := 0; i < technologyAIFieldGroupCount; i++ {
		offset := technologyAIFieldGroupOffset + i*4
		binary.LittleEndian.PutUint32(data[offset:offset+4], uint32(i*i))
	}
	for i := 0; i < technologyCount; i++ {
		offset := technologyTableOffset + i*technologyRecordSize
		fieldID := i % 83
		binary.LittleEndian.PutUint16(data[offset:offset+2], uint16(fieldID))
		data[offset+3] = byte(i % technologyAIClassCount)
		data[offset+5] = 1
	}
	start := []uint16{29, 55, 22, 57, 28, 23}
	for i, fieldID := range start {
		binary.LittleEndian.PutUint16(data[newGameFieldsOffset+i*2:newGameFieldsOffset+i*2+2], fieldID)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
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
