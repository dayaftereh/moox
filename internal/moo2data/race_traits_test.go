package moo2data

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeRaceTraits(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "RACESTUF.LBX")
	block := make([]byte, 0, 1024)
	for _, group := range raceTraitSpecs() {
		block = append(block, []byte(group.Label)...)
		block = append(block, 0)
		for _, option := range group.Options {
			block = append(block, []byte(option.Label)...)
			block = append(block, 0)
		}
	}

	const dataOffset = 0x800
	file := make([]byte, dataOffset+len(block))
	binary.LittleEndian.PutUint16(file[0:2], 1)
	binary.LittleEndian.PutUint16(file[2:4], 0xFEAD)
	binary.LittleEndian.PutUint32(file[4:8], 0)
	binary.LittleEndian.PutUint32(file[8:12], dataOffset)
	binary.LittleEndian.PutUint32(file[12:16], uint32(dataOffset+len(block)))
	copy(file[dataOffset:], block)
	if err := os.WriteFile(path, file, 0o644); err != nil {
		t.Fatal(err)
	}

	decoded, err := DecodeRaceTraits(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Groups) != 11 {
		t.Fatalf("groups=%d want=11", len(decoded.Groups))
	}
	if decoded.PickBudget.StartingPicks != 10 || decoded.PickBudget.MaxNegativePicks != 10 {
		t.Fatalf("unexpected pick budget: %+v", decoded.PickBudget)
	}

	var optionCount int
	var creativeCost int
	var foundLithovoreMutex bool
	for _, group := range decoded.Groups {
		optionCount += len(group.Options)
		for _, option := range group.Options {
			if option.ID == "creative" {
				creativeCost = option.PickCost
			}
			if option.ID == "lithovore" {
				for _, id := range option.MutexWith {
					if id == "farming_plus_1" {
						foundLithovoreMutex = true
					}
				}
			}
		}
	}
	if optionCount != 53 {
		t.Fatalf("options=%d want=53", optionCount)
	}
	if creativeCost != 8 {
		t.Fatalf("creative cost=%d want=8", creativeCost)
	}
	if !foundLithovoreMutex {
		t.Fatal("lithovore/farming incompatibility not encoded")
	}
}
