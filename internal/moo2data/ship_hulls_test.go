package moo2data

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeShipHullsFindsOriginalSequenceAndPictureIDs(t *testing.T) {
	root := t.TempDir()
	values := []string{"Before"}
	for _, spec := range shipHullSpecs {
		values = append(values, spec.Name)
	}
	values = append(values, "After")
	if err := os.WriteFile(filepath.Join(root, "TECHNAME.LBX"), buildAssetTestLBX([][]byte{nullSeparated(values)}), 0o644); err != nil {
		t.Fatal(err)
	}
	writeSyntheticShipHullOrion2(t, filepath.Join(root, "Orion2.exe"))

	bundle, err := DecodeShipHulls(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Rules.Hulls) != 6 || len(bundle.English.Strings) != 6 {
		t.Fatalf("hulls=%d strings=%d", len(bundle.Rules.Hulls), len(bundle.English.Strings))
	}
	frigate := bundle.Rules.Hulls[0]
	if frigate.ID != "frigate" || frigate.SizeIndex != 0 || len(frigate.StrategicPictureIDs) != 8 || frigate.StrategicPictureIDs[0] != 0 || frigate.StrategicPictureIDs[7] != 7 {
		t.Fatalf("frigate=%+v", frigate)
	}
	titan := bundle.Rules.Hulls[4]
	if titan.ID != "titan" || titan.StrategicPictureIDs[0] != 32 || titan.StrategicPictureIDs[7] != 39 {
		t.Fatalf("titan=%+v", titan)
	}
	doom := bundle.Rules.Hulls[5]
	if doom.ID != "doom_star" || doom.SizeIndex != 5 || len(doom.StrategicPictureIDs) != 1 || doom.StrategicPictureIDs[0] != 43 {
		t.Fatalf("doom star=%+v", doom)
	}
	if got := bundle.English.Strings["ship_hull.doom_star.name"]; got != "Doom Star" {
		t.Fatalf("doom star name=%q", got)
	}
}

func writeSyntheticShipHullOrion2(t *testing.T, path string) {
	t.Helper()
	const (
		moduleBase  = 0x100
		leRelative  = 0x80
		leOffset    = moduleBase + leRelative
		pageSize    = 0x1000
		dataPages   = 0x300
		objectTable = 0xC4
		pageMap     = 0xE0
		objectPages = 0x52
		logicOffset = 0x51731
	)
	data := make([]byte, moduleBase+dataPages+objectPages*pageSize)
	data[moduleBase] = 'M'
	data[moduleBase+1] = 'Z'
	binary.LittleEndian.PutUint32(data[moduleBase+0x3C:moduleBase+0x40], leRelative)
	data[leOffset] = 'L'
	data[leOffset+1] = 'E'
	binary.LittleEndian.PutUint32(data[leOffset+0x28:leOffset+0x2C], pageSize)
	binary.LittleEndian.PutUint32(data[leOffset+0x40:leOffset+0x44], objectTable)
	binary.LittleEndian.PutUint32(data[leOffset+0x44:leOffset+0x48], 1)
	binary.LittleEndian.PutUint32(data[leOffset+0x48:leOffset+0x4C], pageMap)
	binary.LittleEndian.PutUint32(data[leOffset+0x80:leOffset+0x84], dataPages)

	obj := leOffset + objectTable
	binary.LittleEndian.PutUint32(data[obj:obj+4], objectPages*pageSize)
	binary.LittleEndian.PutUint32(data[obj+12:obj+16], 1)
	binary.LittleEndian.PutUint32(data[obj+16:obj+20], objectPages)
	pm := leOffset + pageMap
	for i := 0; i < objectPages; i++ {
		pageNum := i + 1
		data[pm+i*4] = byte(pageNum >> 16)
		data[pm+i*4+1] = byte(pageNum >> 8)
		data[pm+i*4+2] = byte(pageNum)
		data[pm+i*4+3] = 0
	}

	pictureLogic := []byte{
		0x66, 0x83, 0x7D, 0xD4, 0x05,
		0x75, 0x09,
		0x8B, 0x45, 0xD8,
		0xC6, 0x40, 0x5C, 0x2B,
		0xEB, 0x1A,
		0x8A, 0x55, 0xD4,
		0xB8, 0x08, 0x00, 0x00, 0x00,
		0xC0, 0xE2, 0x03,
		0xE8, 0x4F, 0x30, 0x0C, 0x00,
		0x00, 0xD0,
		0x8B, 0x55, 0xD8,
		0xFE, 0xC8,
		0x88, 0x42, 0x5C,
	}
	physical := moduleBase + dataPages + logicOffset
	copy(data[physical:physical+len(pictureLogic)], pictureLogic)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
