package moo2data

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"moox/internal/ruleset"
)

func TestDecodeAssetsBuildsRacePortraitsAndRoleIcons(t *testing.T) {
	root := t.TempDir()

	raceSelBlocks := make([][]byte, 29)
	for i := range raceSelBlocks {
		raceSelBlocks[i] = []byte{0}
	}
	for i := 15; i <= 28; i++ {
		raceSelBlocks[i] = syntheticAssetGraphic(290, 322)
	}
	if err := os.WriteFile(filepath.Join(root, "RACESEL.LBX"), buildAssetTestLBX(raceSelBlocks), 0o644); err != nil {
		t.Fatal(err)
	}

	raceIconBlocks := make([][]byte, 169)
	for i := range raceIconBlocks {
		raceIconBlocks[i] = []byte{0}
	}
	for order := 0; order < 13; order++ {
		for _, variant := range []int{1, 3, 5, 7} {
			raceIconBlocks[order*13+variant] = syntheticAssetGraphic(28, 28)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "RACEICON.LBX"), buildAssetTestLBX(raceIconBlocks), 0o644); err != nil {
		t.Fatal(err)
	}

	writeKnownAssetTestArchives(t, root)

	races := &ruleset.RacesFile{Ruleset: "moo2-1.31"}
	ids := []string{"alkari", "bulrathi", "darlok", "elerian", "gnolam", "human", "klackon", "meklar", "mrrshan", "psilon", "sakkra", "silicoid", "trilarian"}
	for order, id := range ids {
		races.Races = append(races.Races, ruleset.Race{
			ID:               id,
			Order:            order,
			PortraitAssetKey: "race." + id + ".portrait",
			IconAssetKey:     "race." + id + ".icon",
		})
	}

	assets, err := DecodeAssets(root, races)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets.Assets) != 93 {
		t.Fatalf("assets=%d, want 93", len(assets.Assets))
	}

	byKey := make(map[string]ruleset.Asset, len(assets.Assets))
	for _, asset := range assets.Assets {
		byKey[asset.Key] = asset
	}
	portrait := byKey["race.alkari.portrait"]
	if portrait.Reference == nil || portrait.Reference.Archive != "RACESEL.LBX" || portrait.Reference.Block != 15 || portrait.Reference.Width != 290 || portrait.Reference.Height != 322 {
		t.Fatalf("alkari portrait=%+v", portrait)
	}
	marine := byKey["race.trilarian.icon.marine"]
	if marine.Reference == nil || marine.Reference.Archive != "RACEICON.LBX" || marine.Reference.Block != 163 {
		t.Fatalf("trilarian marine=%+v", marine)
	}
	generic := byKey["race.human.icon"]
	if generic.Status != "pending" || generic.Reference != nil {
		t.Fatalf("generic human icon=%+v", generic)
	}
	custom := byKey["race.custom.portrait"]
	if custom.Reference == nil || custom.Reference.Block != 28 {
		t.Fatalf("custom portrait=%+v", custom)
	}
}

func syntheticAssetGraphic(width, height int) []byte {
	frame := []byte{1, 0, 0, 0, 0, 0, 0, 0, 0xE8, 0x03}
	const headerEnd = 20
	data := make([]byte, headerEnd+len(frame))
	binary.LittleEndian.PutUint16(data[0:2], uint16(width))
	binary.LittleEndian.PutUint16(data[2:4], uint16(height))
	binary.LittleEndian.PutUint16(data[6:8], 1)
	binary.LittleEndian.PutUint32(data[12:16], headerEnd)
	binary.LittleEndian.PutUint32(data[16:20], uint32(len(data)))
	copy(data[headerEnd:], frame)
	return data
}

func buildAssetTestLBX(blocks [][]byte) []byte {
	count := len(blocks)
	headerEnd := 8 + (count+1)*4
	offset := headerEnd
	offsets := make([]int, count+1)
	for i, block := range blocks {
		offsets[i] = offset
		offset += len(block)
	}
	offsets[count] = offset
	data := make([]byte, offset)
	binary.LittleEndian.PutUint16(data[0:2], uint16(count))
	binary.LittleEndian.PutUint16(data[2:4], 0xFEAD)
	for i, value := range offsets {
		binary.LittleEndian.PutUint32(data[8+i*4:12+i*4], uint32(value))
	}
	cursor := headerEnd
	for _, block := range blocks {
		copy(data[cursor:], block)
		cursor += len(block)
	}
	return data
}

func writeKnownAssetTestArchives(t *testing.T, root string) {
	t.Helper()
	write := func(name string, blocks [][]byte) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(root, name), buildAssetTestLBX(blocks), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("MAINMENU.LBX", [][]byte{syntheticAssetGraphic(640, 480)})
	write("BUFFER0.LBX", [][]byte{syntheticAssetGraphic(640, 480)})
	write("COLSUM.LBX", [][]byte{syntheticAssetGraphic(640, 480)})
	colpups := make([][]byte, 6)
	for i := range colpups {
		colpups[i] = []byte{0}
	}
	colpups[5] = syntheticAssetGraphic(640, 480)
	write("COLPUPS.LBX", colpups)
	colony2 := make([][]byte, 50)
	for i := range colony2 {
		colony2[i] = []byte{0}
	}
	for i := 0; i <= 6; i++ {
		colony2[i] = syntheticAssetGraphic(16, 23)
	}
	colony2[49] = syntheticAssetGraphic(640, 480)
	write("COLONY2.LBX", colony2)
	write("BLDG0.LBX", [][]byte{syntheticAssetGraphic(640, 480)})
}
