package moo2data

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeShipHullsFindsOriginalSequenceAndRuntimeTables(t *testing.T) {
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
	if frigate.ID != "frigate" || frigate.SizeIndex != 0 || frigate.BaseCostPP != 20 || frigate.BaseSpace != 25 || len(frigate.StrategicPictureIDs) != 8 || frigate.StrategicPictureIDs[0] != 0 || frigate.StrategicPictureIDs[7] != 7 {
		t.Fatalf("frigate=%+v", frigate)
	}
	titan := bundle.Rules.Hulls[4]
	if titan.ID != "titan" || titan.BaseCostPP != 1500 || titan.BaseSpace != 500 || titan.StrategicPictureIDs[0] != 32 || titan.StrategicPictureIDs[7] != 39 {
		t.Fatalf("titan=%+v", titan)
	}
	doom := bundle.Rules.Hulls[5]
	if doom.ID != "doom_star" || doom.SizeIndex != 5 || doom.BaseCostPP != 4000 || doom.BaseSpace != 1200 || len(doom.StrategicPictureIDs) != 1 || doom.StrategicPictureIDs[0] != 43 || doom.TacticalAssetKey != "ship_hull.doom_star.tactical" {
		t.Fatalf("doom star=%+v", doom)
	}
	if got := bundle.English.Strings["ship_hull.doom_star.name"]; got != "Doom Star" {
		t.Fatalf("doom star name=%q", got)
	}
	components := bundle.Rules.MandatoryComponents
	if components.Drives[0].ID != "nuclear_drive" || components.Drives[0].TechnologyID != 120 || components.Drives[0].FTLSpeed != 2 {
		t.Fatalf("nuclear drive=%+v", components.Drives[0])
	}
	if components.Computers[0].CostByHullPP[0] != 5 || components.Shields[0].SpaceByHull[0] != 5 || components.Shields[0].CostByHullPP[0] != 3 {
		t.Fatalf("baseline components computer=%+v shield=%+v", components.Computers[0], components.Shields[0])
	}
	if components.Armors[0].TechnologyID != 187 || components.Armors[0].CostPercent != 0 || components.FuelCells[0].TechnologyID != 167 || components.FuelCells[0].RangeParsecs != 4 {
		t.Fatalf("baseline armor/fuel armor=%+v fuel=%+v", components.Armors[0], components.FuelCells[0])
	}
}

func writeSyntheticShipHullOrion2(t *testing.T, path string) {
	t.Helper()
	const (
		moduleBase   = 0x100
		leRelative   = 0x80
		leOffset     = moduleBase + leRelative
		pageSize     = 0x1000
		dataPages    = 0x400
		objectTable  = 0xC4
		pageMap      = 0x100
		object1Pages = 0x52
		object2Pages = 0x09
		logicOffset  = 0x51731
	)
	totalObjectPages := object1Pages + object2Pages
	data := make([]byte, moduleBase+dataPages+totalObjectPages*pageSize)
	data[moduleBase] = 'M'
	data[moduleBase+1] = 'Z'
	binary.LittleEndian.PutUint32(data[moduleBase+0x3C:moduleBase+0x40], leRelative)
	data[leOffset] = 'L'
	data[leOffset+1] = 'E'
	binary.LittleEndian.PutUint32(data[leOffset+0x28:leOffset+0x2C], pageSize)
	binary.LittleEndian.PutUint32(data[leOffset+0x40:leOffset+0x44], objectTable)
	binary.LittleEndian.PutUint32(data[leOffset+0x44:leOffset+0x48], 2)
	binary.LittleEndian.PutUint32(data[leOffset+0x48:leOffset+0x4C], pageMap)
	binary.LittleEndian.PutUint32(data[leOffset+0x80:leOffset+0x84], dataPages)

	obj1 := leOffset + objectTable
	binary.LittleEndian.PutUint32(data[obj1:obj1+4], object1Pages*pageSize)
	binary.LittleEndian.PutUint32(data[obj1+12:obj1+16], 1)
	binary.LittleEndian.PutUint32(data[obj1+16:obj1+20], object1Pages)
	obj2 := obj1 + 24
	binary.LittleEndian.PutUint32(data[obj2:obj2+4], object2Pages*pageSize)
	binary.LittleEndian.PutUint32(data[obj2+12:obj2+16], object1Pages+1)
	binary.LittleEndian.PutUint32(data[obj2+16:obj2+20], object2Pages)
	pm := leOffset + pageMap
	for i := 0; i < totalObjectPages; i++ {
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
	object1Physical := moduleBase + dataPages
	copy(data[object1Physical+logicOffset:object1Physical+logicOffset+len(pictureLogic)], pictureLogic)
	object2Physical := object1Physical + object1Pages*pageSize
	putWord := func(offset, value int) {
		binary.LittleEndian.PutUint16(data[object2Physical+offset:object2Physical+offset+2], uint16(int16(value)))
	}
	hullCosts := []int{20, 70, 250, 600, 1500, 4000}
	hullSpaces := []int{25, 60, 120, 250, 500, 1200}
	for i := 0; i < 6; i++ {
		putWord(0x801E+i*0x24, hullCosts[i])
		putWord(0x8020+i*0x24, hullSpaces[i])
	}
	writeHullValues := func(base int, values []int) {
		for i, value := range values {
			putWord(base+i*2, value)
		}
	}
	driveTechs := []int{120, 72, 96, 11, 88, 95}
	for i, tech := range driveTechs {
		base := 0x7E76 + (i+1)*0x2E
		putWord(base, tech)
		writeHullValues(base+2, []int{0, 0, 0, 0, 0, 0})
		writeHullValues(base+0x0E, []int{0, 0, 0, 0, 0, 0})
	}
	computerTechs := []int{58, 122, 143, 44, 110}
	for i, tech := range computerTechs {
		base := 0x7DF2 + (i+1)*0x16
		putWord(base, tech)
		writeHullValues(base+2, []int{5, 15, 50, 125, 300, 800})
	}
	armorTechs := []int{187, 191, 203, 117, 2, 201}
	for i, tech := range armorTechs {
		base := 0x763E + (i+1)*0x0F
		putWord(base, tech)
		putWord(base+6, 0)
	}
	shieldTechs := []int{33, 34, 35, 36, 37}
	for i, tech := range shieldTechs {
		base := 0x76A7 + (i+1)*0x3B
		putWord(base, tech)
		writeHullValues(base+2, []int{5, 10, 20, 50, 100, 200})
		writeHullValues(base+0x0E, []int{3, 5, 10, 25, 50, 100})
	}
	fuelTechs := []int{167, 51, 98, 194, 184}
	fuelRanges := []int{4, 6, 9, 12, 255}
	for i, tech := range fuelTechs {
		base := 0x7FE8 + (i+1)*0x0A
		putWord(base, tech)
		putWord(base+2, fuelRanges[i])
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
