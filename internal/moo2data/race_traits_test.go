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
	labels := make([]byte, 0, 1024)
	costs := make([]byte, 0, 57)
	for _, group := range raceTraitSpecs() {
		labels = append(labels, []byte(group.Label)...)
		labels = append(labels, 0)
		for _, option := range group.Options {
			labels = append(labels, []byte(option.Label)...)
			labels = append(labels, 0)
			costs = append(costs, byte(int8(option.PickCost)))
		}
	}
	costs = append(costs, 0, 0, 0, 0)

	blocks := make([][]byte, 7)
	blocks[0] = labels
	blocks[6] = costs
	file := buildTestLBX(blocks)
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
	var creativeCostSource string
	var foundLithovoreMutex bool
	for _, group := range decoded.Groups {
		optionCount += len(group.Options)
		for _, option := range group.Options {
			if option.ID == "creative" {
				creativeCost = option.PickCost
				creativeCostSource = option.PickCostSource.SourceID
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
	if creativeCostSource != raceStuffCostsSourceID {
		t.Fatalf("creative cost source=%q", creativeCostSource)
	}
	if !foundLithovoreMutex {
		t.Fatal("lithovore/farming incompatibility not encoded")
	}
}

func buildTestLBX(blocks [][]byte) []byte {
	const dataOffset = 0x800
	count := len(blocks)
	offsets := make([]uint32, count+1)
	cursor := dataOffset
	for i, block := range blocks {
		offsets[i] = uint32(cursor)
		cursor += len(block)
	}
	offsets[count] = uint32(cursor)
	file := make([]byte, cursor)
	binary.LittleEndian.PutUint16(file[0:2], uint16(count))
	binary.LittleEndian.PutUint16(file[2:4], 0xFEAD)
	for i, offset := range offsets {
		binary.LittleEndian.PutUint32(file[8+i*4:12+i*4], offset)
	}
	cursor = dataOffset
	for _, block := range blocks {
		copy(file[cursor:], block)
		cursor += len(block)
	}
	return file
}
