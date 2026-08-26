package moo2data

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeBuildingsFindsOriginalNamesAndTechnologyLinks(t *testing.T) {
	root := t.TempDir()
	specs := buildingSpecs()
	helpBlock := syntheticHelpRecords(t, specs)
	if err := os.WriteFile(filepath.Join(root, "HELP.LBX"), buildAssetTestLBX([][]byte{helpBlock}), 0o644); err != nil {
		t.Fatal(err)
	}
	techNameBlock := syntheticBuildingTechNameBlock(t, specs)
	if err := os.WriteFile(filepath.Join(root, "TECHNAME.LBX"), buildAssetTestLBX([][]byte{techNameBlock}), 0o644); err != nil {
		t.Fatal(err)
	}
	writeSyntheticOrion2(t, filepath.Join(root, "Orion2.exe"))

	bundle, err := DecodeBuildings(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Rules.Buildings) != 48 || len(bundle.English.Strings) != 48 {
		t.Fatalf("buildings=%d strings=%d", len(bundle.Rules.Buildings), len(bundle.English.Strings))
	}

	first := bundle.Rules.Buildings[0]
	if first.ID != "alien_management_center" || first.ProductionID != 1 || first.Order != 0 || first.TechnologyID != 5 || first.TechnologyKey != "alien_management_center" || first.ColonyReferenceAssetKey != "building.alien_management_center.colony_reference" {
		t.Fatalf("first=%+v", first)
	}
	if first.ProductionIDVerification != "original-exe-table-id-techname-crosschecked" || first.ProductionIDSource.SourceID != buildingTableSourceID {
		t.Fatalf("first production provenance=%+v", first)
	}
	if first.TechnologyLinkVerification != "original-exe-table-tech-id-and-techname-exact" || first.TechnologySource.SourceID != buildingTableSourceID {
		t.Fatalf("first technology provenance=%+v", first)
	}

	hydro := bundle.Rules.Buildings[20]
	if hydro.ID != "hydroponic_farms" || hydro.TechnologyID != 87 || hydro.TechnologyKey != "hydroponic_farm" || hydro.TechnologyLinkVerification != "original-exe-table-tech-id-and-singular-techname-alias" {
		t.Fatalf("hydroponic farms=%+v", hydro)
	}

	pollution := bundle.Rules.Buildings[31]
	if pollution.NameVerification != "original-techname-string-2-occurrences" || pollution.TechnologyID != 142 || pollution.TechnologyKey != "pollution_processor" || pollution.TechnologyLinkVerification != "original-exe-table-tech-id-and-techname-exact" {
		t.Fatalf("pollution processor=%+v", pollution)
	}

	last := bundle.Rules.Buildings[47]
	if last.ID != "artificial_planet" || last.ProductionID != 48 || last.Order != 47 || last.TechnologyID != 16 || last.TechnologyKey != "planet_construction" || last.TechnologyLinkVerification != "original-exe-table-tech-id-planet-construction" || last.TechnologySource.SourceID != buildingTableSourceID {
		t.Fatalf("artificial planet=%+v", last)
	}
	if last.NameVerification != "original-techname-string" || last.ProductionIDVerification != "original-exe-table-id-secondary-display-name-link" {
		t.Fatalf("artificial planet name/id verification=%+v", last)
	}
	if got := bundle.English.Strings["building.automated_factories.name"]; got != "Automated Factories" {
		t.Fatalf("automated factories name=%q", got)
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

func syntheticBuildingTechnologyIDs(t *testing.T) []int {
	t.Helper()
	ids := []int{
		5, 14, 15, 18, 19, 21, 22, 27, 32, 39, 40, 49,
		50, 52, 61, 68, 74, 75, 76, 86, 87, 103, 129, 130,
		131, 132, 133, 134, 135, 136, 141, 142, 152, 154, 155, 156,
		162, 163, 164, 168, 169, 174, 178, 183, 197, 198, 67, 16,
	}
	if len(ids) != 48 {
		t.Fatalf("technology id fixture count=%d want=48", len(ids))
	}
	return ids
}

func syntheticBuildingTechNameBlock(t *testing.T, specs []buildingSpec) []byte {
	t.Helper()
	techs := make([]string, technologyCount)
	for i := range techs {
		techs[i] = fmt.Sprintf("Synthetic Technology %03d", i+1)
	}
	techs[0] = "Achilles Targeting Unit"
	techs[202] = "Zortrium Armor"

	technologyIDs := syntheticBuildingTechnologyIDs(t)
	if len(technologyIDs) != len(specs) {
		t.Fatalf("technology id fixture count=%d specs=%d", len(technologyIDs), len(specs))
	}
	for i, spec := range specs {
		name := spec.Name
		if spec.ID == "hydroponic_farms" {
			name = "Hydroponic Farm"
		}
		if spec.ID == "artificial_planet" {
			name = "Planet Construction"
		}
		techs[technologyIDs[i]-1] = name
	}

	values := []string{"Starting Tech", "No Tech"}
	values = append(values, techs...)
	values = append(values, "Biology", "Pollution Processor", "Artificial Planet")
	return nullSeparated(values)
}

func writeSyntheticOrion2(t *testing.T, path string) {
	t.Helper()
	const (
		moduleBase    = 0x100
		leRelative    = 0x80
		leOffset      = moduleBase + leRelative
		pageSize      = 0x1000
		dataPages     = 0x400
		objectTable   = 0xC4
		pageMap       = 0xF4
		objectCount   = 2
		object1Pages  = 1
		object2Pages  = 7
		buildingTable = 0x6B3D
		entrySize     = 0x13
	)
	pageCount := object1Pages + object2Pages
	data := make([]byte, moduleBase+dataPages+pageCount*pageSize)
	data[moduleBase] = 'M'
	data[moduleBase+1] = 'Z'
	binary.LittleEndian.PutUint32(data[moduleBase+0x3C:moduleBase+0x40], leRelative)
	data[leOffset] = 'L'
	data[leOffset+1] = 'E'
	binary.LittleEndian.PutUint32(data[leOffset+0x28:leOffset+0x2C], pageSize)
	binary.LittleEndian.PutUint32(data[leOffset+0x40:leOffset+0x44], objectTable)
	binary.LittleEndian.PutUint32(data[leOffset+0x44:leOffset+0x48], objectCount)
	binary.LittleEndian.PutUint32(data[leOffset+0x48:leOffset+0x4C], pageMap)
	binary.LittleEndian.PutUint32(data[leOffset+0x80:leOffset+0x84], dataPages)

	obj := leOffset + objectTable
	binary.LittleEndian.PutUint32(data[obj:obj+4], pageSize)
	binary.LittleEndian.PutUint32(data[obj+12:obj+16], 1)
	binary.LittleEndian.PutUint32(data[obj+16:obj+20], object1Pages)
	obj += 24
	binary.LittleEndian.PutUint32(data[obj:obj+4], object2Pages*pageSize)
	binary.LittleEndian.PutUint32(data[obj+12:obj+16], 2)
	binary.LittleEndian.PutUint32(data[obj+16:obj+20], object2Pages)

	pm := leOffset + pageMap
	for i := 0; i < pageCount; i++ {
		pageNum := i + 1
		data[pm+i*4] = byte(pageNum >> 16)
		data[pm+i*4+1] = byte(pageNum >> 8)
		data[pm+i*4+2] = byte(pageNum)
		data[pm+i*4+3] = 0
	}

	technologyIDs := syntheticBuildingTechnologyIDs(t)
	object2Physical := moduleBase + dataPages + object1Pages*pageSize
	for index := 0; index < 49; index++ {
		offset := object2Physical + buildingTable + index*entrySize
		binary.LittleEndian.PutUint16(data[offset+4:offset+6], uint16(index))
		if index > 0 {
			binary.LittleEndian.PutUint16(data[offset+6:offset+8], uint16(technologyIDs[index-1]))
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
