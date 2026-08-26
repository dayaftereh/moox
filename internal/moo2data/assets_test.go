package moo2data

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"moox/internal/ruleset"
)

func TestDecodeAssetsBuildsRaceBuildingAndShipSemantics(t *testing.T) {
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

	races := syntheticAssetRaces()
	buildings := syntheticAssetBuildings()
	shipHulls := syntheticAssetShipHulls()
	assets, err := decodeAssetsWithEvidence(root, races, buildings, shipHulls, buildingGraphicsEvidence{
		ExecutableSHA256: "synthetic-exe",
		EStringsSHA256:   "synthetic-estrings",
		EStringsBlockSHA: "synthetic-estrings-block",
	}, shipStrategicGraphicsEvidence{ExecutableSHA256: "synthetic-exe", ShipsSHA256: "synthetic-ships"})
	if err != nil {
		t.Fatal(err)
	}
	if len(assets.Assets) != 150 {
		t.Fatalf("assets=%d, want 150", len(assets.Assets))
	}
	if err := assets.ValidateAgainstRaces(races); err != nil {
		t.Fatal(err)
	}
	if err := assets.ValidateAgainstBuildings(buildings); err != nil {
		t.Fatal(err)
	}
	if err := assets.ValidateAgainstShipHulls(shipHulls); err != nil {
		t.Fatal(err)
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
	if generic.Status != "pending" || generic.Reference != nil || len(generic.Variants) != 0 {
		t.Fatalf("generic human icon=%+v", generic)
	}
	custom := byKey["race.custom.portrait"]
	if custom.Reference == nil || custom.Reference.Block != 28 {
		t.Fatalf("custom portrait=%+v", custom)
	}

	alien := byKey["building.alien_management_center.colony"]
	if alien.Kind != "building_colony_set" || len(alien.Variants) != 36 {
		t.Fatalf("alien management center asset=%+v", alien)
	}
	assertBuildingVariant(t, alien, "grid_0_0", "BLDG0.LBX", 0, 0)
	assertBuildingVariant(t, alien, "grid_0_1", "BLDG0.LBX", 11, 11)

	artificial := byKey["building.artificial_planet.colony"]
	if len(artificial.Variants) != 36 {
		t.Fatalf("artificial planet variants=%d", len(artificial.Variants))
	}
	assertBuildingVariant(t, artificial, "grid_0_0", "BLDG4.LBX", 252, 0)
	assertBuildingVariant(t, artificial, "grid_5_5", "BLDG4.LBX", 282, 30)

	anchor := byKey["building.alien_management_center.colony_reference"]
	if anchor.Reference == nil || anchor.Reference.Archive != "BLDG0.LBX" || anchor.Reference.Block != 0 {
		t.Fatalf("published building anchor=%+v", anchor)
	}
}

func TestBuildingArchiveBlockFormula(t *testing.T) {
	tests := []struct {
		buildingID int
		x          int
		y          int
		archive    string
		block      int
		effective  int
	}{
		{1, 0, 0, "BLDG0.LBX", 0, 0},
		{1, 5, 0, "BLDG0.LBX", 5, 5},
		{1, 0, 1, "BLDG0.LBX", 11, 11},
		{1, 5, 1, "BLDG0.LBX", 6, 6},
		{10, 0, 0, "BLDG0.LBX", 324, 0},
		{11, 0, 0, "BLDG1.LBX", 0, 0},
		{40, 0, 0, "BLDG3.LBX", 324, 0},
		{41, 0, 0, "BLDG4.LBX", 0, 0},
		{48, 5, 5, "BLDG4.LBX", 282, 30},
	}
	for _, tt := range tests {
		archive, block, effective, err := buildingArchiveBlock(tt.buildingID, tt.x, tt.y)
		if err != nil {
			t.Fatal(err)
		}
		if archive != tt.archive || block != tt.block || effective != tt.effective {
			t.Fatalf("building %d (%d,%d)=(%s,%d,%d), want (%s,%d,%d)", tt.buildingID, tt.x, tt.y, archive, block, effective, tt.archive, tt.block, tt.effective)
		}
	}
	if _, _, _, err := buildingArchiveBlock(49, 0, 0); err == nil {
		t.Fatal("building id 49 should remain outside the standard building formula")
	}
}

func syntheticAssetRaces() *ruleset.RacesFile {
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
	return races
}

func syntheticAssetBuildings() *ruleset.BuildingsFile {
	file := &ruleset.BuildingsFile{SchemaVersion: ruleset.BuildingsSchemaVersion, Ruleset: "moo2-1.31"}
	for order, spec := range buildingSpecs() {
		file.Buildings = append(file.Buildings, ruleset.Building{
			ID:                         spec.ID,
			Order:                      order,
			ProductionID:               order + 1,
			ProductionIDVerification:   "test",
			ProductionIDSource:         ruleset.FieldProvenance{SourceID: "test-building-table"},
			TechnologyID:               order + 1,
			TechnologyKey:              fmt.Sprintf("test_technology_%d", order+1),
			TechnologyLinkVerification: "test",
			TechnologySource:           ruleset.FieldProvenance{SourceID: "test-building-table"},
			NameKey:                    "building." + spec.ID + ".name",
			NameVerification:           "test",
			NameSource:                 ruleset.FieldProvenance{SourceID: "test-name"},
			ColonyReferenceAssetKey:    "building." + spec.ID + ".colony",
		})
	}
	return file
}

func assertBuildingVariant(t *testing.T, asset ruleset.Asset, id, archive string, block, effective int) {
	t.Helper()
	for _, variant := range asset.Variants {
		if variant.ID != id {
			continue
		}
		if variant.Reference.Archive != archive || variant.Reference.Block != block || variant.Reference.Frame != 0 || variant.Reference.Width != 640 || variant.Reference.Height != 480 {
			t.Fatalf("variant %s reference=%+v", id, variant.Reference)
		}
		if got := variant.Metadata["effective_frame"]; got != effective {
			t.Fatalf("variant %s effective_frame=%d want=%d", id, got, effective)
		}
		return
	}
	t.Fatalf("variant %s not found", id)
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

	for archiveIndex := 0; archiveIndex <= 4; archiveIndex++ {
		count := 360
		if archiveIndex == 4 {
			count = 288
		}
		blocks := make([][]byte, count)
		for i := range blocks {
			blocks[i] = syntheticAssetGraphic(640, 480)
		}
		write(fmt.Sprintf("BLDG%d.LBX", archiveIndex), blocks)
	}

	ships := make([][]byte, 449)
	for i := range ships {
		ships[i] = []byte{0}
	}
	for colorIndex := 0; colorIndex < 8; colorIndex++ {
		for pictureID := 0; pictureID <= 48; pictureID++ {
			width, height := 52, 48
			if pictureID == 43 {
				height = 52
			}
			if pictureID == 49 {
				width, height = 2, 1
			}
			ships[colorIndex*50+pictureID] = syntheticAssetGraphic(width, height)
		}
	}
	for i := 400; i < len(ships); i++ {
		ships[i] = syntheticAssetGraphic(52, 48)
	}
	write("SHIPS.LBX", ships)
}

func syntheticAssetShipHulls() *ruleset.ShipHullsFile {
	file := &ruleset.ShipHullsFile{SchemaVersion: ruleset.ShipHullsSchemaVersion, Ruleset: "moo2-1.31"}
	specs := []struct {
		id       string
		pictures []int
	}{
		{id: "frigate", pictures: []int{0, 1, 2, 3, 4, 5, 6, 7}},
		{id: "destroyer", pictures: []int{8, 9, 10, 11, 12, 13, 14, 15}},
		{id: "cruiser", pictures: []int{16, 17, 18, 19, 20, 21, 22, 23}},
		{id: "battleship", pictures: []int{24, 25, 26, 27, 28, 29, 30, 31}},
		{id: "titan", pictures: []int{32, 33, 34, 35, 36, 37, 38, 39}},
		{id: "doom_star", pictures: []int{43}},
	}
	for sizeIndex, spec := range specs {
		file.Hulls = append(file.Hulls, ruleset.ShipHull{
			ID:                       spec.id,
			SizeIndex:                sizeIndex,
			NameKey:                  "ship_hull." + spec.id + ".name",
			NameSource:               ruleset.FieldProvenance{SourceID: "test-techname"},
			StrategicPictureIDs:      spec.pictures,
			PictureLogicVerification: "test",
			PictureLogicSource:       ruleset.FieldProvenance{SourceID: "test-picture-logic"},
			StrategicAssetKey:        "ship_hull." + spec.id + ".strategic",
		})
	}
	return file
}
