package moo2data

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"moox/internal/i18n"
	"moox/internal/lbx"
	"moox/internal/moo2exe"
	"moox/internal/ruleset"
	"moox/internal/textscan"
)

const (
	shipHullNamesSourceID        = "moo2-1.31-techname-ship-hull-names"
	shipHullPictureLogicSourceID = "moo2-1.31-orion2-player-ship-picture-logic"
	shipHullRuntimeSourceID      = "moo2-1.31-orion2-ship-design-tables"
	shipHullPictureLogicSHA256   = "6daa37d29c2818c73add925fd145b49728a015dcd16c9d00e87486eee6605de2"
)

type ShipHullsBundle struct {
	Rules   *ruleset.ShipHullsFile
	English *i18n.File
}

type shipHullSpec struct {
	ID   string
	Name string
}

var shipHullSpecs = []shipHullSpec{
	{ID: "frigate", Name: "Frigate"},
	{ID: "destroyer", Name: "Destroyer"},
	{ID: "cruiser", Name: "Cruiser"},
	{ID: "battleship", Name: "Battleship"},
	{ID: "titan", Name: "Titan"},
	{ID: "doom_star", Name: "Doom Star"},
}

func DecodeShipHulls(installationRoot string) (*ShipHullsBundle, error) {
	techNamePath := filepath.Join(installationRoot, "TECHNAME.LBX")
	archive, err := lbx.Open(techNamePath)
	if err != nil {
		return nil, fmt.Errorf("open TECHNAME.LBX: %w", err)
	}
	block, err := archive.ReadEntry(0)
	if err != nil {
		return nil, fmt.Errorf("read TECHNAME.LBX block 0: %w", err)
	}
	fileHash, err := sha256File(techNamePath)
	if err != nil {
		return nil, err
	}
	blockSum := sha256.Sum256(block)
	blockIndex := 0
	runs := textscan.ASCII(block, 3)
	start, err := findShipHullNameSequence(runs)
	if err != nil {
		return nil, err
	}

	exe, err := moo2exe.Open(filepath.Join(installationRoot, "Orion2.exe"))
	if err != nil {
		return nil, fmt.Errorf("open Orion2.exe for ship hull data: %w", err)
	}
	const (
		pictureLogicObject = 1
		pictureLogicOffset = 0x51731
		pictureLogicSize   = 0x2A
		designDataObject   = 2
		designDataOffset   = 0x7600
		designDataSize     = 0x0B00
	)
	logic, err := exe.ReadObject(pictureLogicObject, pictureLogicOffset, pictureLogicSize)
	if err != nil {
		return nil, fmt.Errorf("read Orion2.exe Auto_Design_Ship_ picture logic: %w", err)
	}
	logicSum := sha256.Sum256(logic)
	if got := hex.EncodeToString(logicSum[:]); got != shipHullPictureLogicSHA256 {
		return nil, fmt.Errorf("Orion2.exe Auto_Design_Ship_ picture logic hash=%s, expected %s", got, shipHullPictureLogicSHA256)
	}
	designData, err := exe.ReadObject(designDataObject, designDataOffset, designDataSize)
	if err != nil {
		return nil, fmt.Errorf("read Orion2.exe ship design tables: %w", err)
	}
	readWord := func(objectOffset int) (int, error) {
		offset := objectOffset - designDataOffset
		if offset < 0 || offset+2 > len(designData) {
			return 0, fmt.Errorf("ship design table offset 0x%X outside extracted range", objectOffset)
		}
		return int(int16(binary.LittleEndian.Uint16(designData[offset : offset+2]))), nil
	}
	readHullValues := func(base int) ([]int, error) {
		values := make([]int, len(shipHullSpecs))
		for i := range values {
			value, err := readWord(base + i*2)
			if err != nil {
				return nil, err
			}
			values[i] = value
		}
		return values, nil
	}

	out := &ruleset.ShipHullsFile{
		SchemaVersion: ruleset.ShipHullsSchemaVersion,
		Ruleset:       "moo2-1.31",
		Sources: []ruleset.Source{
			{
				ID:          shipHullNamesSourceID,
				Type:        "original-observed",
				Description: "English TECHNAME.LBX block 0 contains the consecutive player hull-name sequence Frigate, Destroyer, Cruiser, Battleship, Titan, Doom Star",
				Archive:     "TECHNAME.LBX",
				Block:       &blockIndex,
				SHA256:      fileHash,
				BlockSHA256: hex.EncodeToString(blockSum[:]),
			},
			{
				ID:          shipHullPictureLogicSourceID,
				Type:        "original-observed",
				Description: "Orion2.exe 1.31 Auto_Design_Ship_ object 1 range 0x51731..0x5175A uses Random_(8), size_index<<3 and -1 for sizes 0..4, producing eight picture IDs per size; size 5 is hard-coded to picture ID 43. Code range SHA-256 is verified before normalization.",
				Archive:     "Orion2.exe",
				SHA256:      exe.SHA256(),
			},
			{
				ID:          shipHullRuntimeSourceID,
				Type:        "original-observed",
				Description: "Orion2.exe 1.31 LE object 2 ship-design tables: hull cost/space records and mandatory drive/computer/armor/shield/fuel component records used by Total_Design_Cost_, Design_Template_Costs_ and Design_Template_Space_Requirements_. Offsets in field provenance are object-2-relative offsets.",
				Archive:     "Orion2.exe",
				SHA256:      exe.SHA256(),
			},
		},
	}
	english := &i18n.File{
		SchemaVersion: i18n.SchemaVersion,
		Locale:        "en",
		Sources: []i18n.Source{{
			ID:          shipHullNamesSourceID,
			Type:        "original-observed",
			Description: "Player ship hull names from TECHNAME.LBX block 0",
			Archive:     "TECHNAME.LBX",
			Block:       &blockIndex,
			SHA256:      fileHash,
			BlockSHA256: hex.EncodeToString(blockSum[:]),
		}},
		Strings: make(map[string]string, len(shipHullSpecs)),
	}

	for sizeIndex, spec := range shipHullSpecs {
		run := runs[start+sizeIndex]
		nameOffset := run.Offset
		pictureIDs := make([]int, 0, 8)
		if sizeIndex < 5 {
			for style := 0; style < 8; style++ {
				pictureIDs = append(pictureIDs, sizeIndex*8+style)
			}
		} else {
			pictureIDs = append(pictureIDs, 43)
		}
		nameKey := "ship_hull." + spec.ID + ".name"
		runtimeOffset := 0x801E + sizeIndex*0x24
		baseCost, err := readWord(runtimeOffset)
		if err != nil {
			return nil, err
		}
		baseSpace, err := readWord(runtimeOffset + 2)
		if err != nil {
			return nil, err
		}
		out.Hulls = append(out.Hulls, ruleset.ShipHull{
			ID:                       spec.ID,
			SizeIndex:                sizeIndex,
			NameKey:                  nameKey,
			NameSource:               ruleset.FieldProvenance{SourceID: shipHullNamesSourceID, Offset: &nameOffset},
			BaseCostPP:               baseCost,
			BaseSpace:                baseSpace,
			RuntimeSource:            ruleset.FieldProvenance{SourceID: shipHullRuntimeSourceID, Offset: intPointer(runtimeOffset)},
			StrategicPictureIDs:      pictureIDs,
			PictureLogicVerification: "original-techname-sequence-plus-hash-verified-auto-design-picture-logic",
			PictureLogicSource:       ruleset.FieldProvenance{SourceID: shipHullPictureLogicSourceID},
			StrategicAssetKey:        "ship_hull." + spec.ID + ".strategic",
			TacticalAssetKey:         "ship_hull." + spec.ID + ".tactical",
		})
		english.Strings[nameKey] = spec.Name
	}

	componentSource := func(offset int) ruleset.FieldProvenance {
		return ruleset.FieldProvenance{SourceID: shipHullRuntimeSourceID, Offset: intPointer(offset)}
	}
	driveSpecs := []struct {
		id    string
		speed int
	}{
		{"nuclear_drive", 2}, {"fusion_drive", 3}, {"ion_drive", 4}, {"anti_matter_drive", 5}, {"hyper_drive", 6}, {"interphased_drive", 7},
	}
	for i, spec := range driveSpecs {
		index := i + 1
		base := 0x7E76 + index*0x2E
		technologyID, err := readWord(base)
		if err != nil {
			return nil, err
		}
		spaces, err := readHullValues(base + 2)
		if err != nil {
			return nil, err
		}
		costs, err := readHullValues(base + 0x0E)
		if err != nil {
			return nil, err
		}
		out.MandatoryComponents.Drives = append(out.MandatoryComponents.Drives, ruleset.ShipDrive{ID: spec.id, ComponentIndex: index, TechnologyID: technologyID, FTLSpeed: spec.speed, SpaceByHull: spaces, CostByHullPP: costs, Source: componentSource(base)})
	}
	computerIDs := []string{"electronic_computer", "optronic_computer", "positronic_computer", "cybertronic_computer", "moleculartronic_computer"}
	for i, id := range computerIDs {
		index := i + 1
		base := 0x7DF2 + index*0x16
		technologyID, err := readWord(base)
		if err != nil {
			return nil, err
		}
		costs, err := readHullValues(base + 2)
		if err != nil {
			return nil, err
		}
		out.MandatoryComponents.Computers = append(out.MandatoryComponents.Computers, ruleset.ShipComputer{ID: id, ComponentIndex: index, TechnologyID: technologyID, CostByHullPP: costs, Source: componentSource(base)})
	}
	armorIDs := []string{"titanium_armor", "tritanium_armor", "zortrium_armor", "neutronium_armor", "adamantium_armor", "xentronium_armor"}
	for i, id := range armorIDs {
		index := i + 1
		base := 0x763E + index*0x0F
		technologyID, err := readWord(base)
		if err != nil {
			return nil, err
		}
		costPercent, err := readWord(base + 6)
		if err != nil {
			return nil, err
		}
		out.MandatoryComponents.Armors = append(out.MandatoryComponents.Armors, ruleset.ShipArmor{ID: id, ComponentIndex: index, TechnologyID: technologyID, CostPercent: costPercent, Source: componentSource(base)})
	}
	shieldIDs := []string{"class_i_shield", "class_iii_shield", "class_v_shield", "class_vii_shield", "class_x_shield"}
	for i, id := range shieldIDs {
		index := i + 1
		base := 0x76A7 + index*0x3B
		technologyID, err := readWord(base)
		if err != nil {
			return nil, err
		}
		spaces, err := readHullValues(base + 2)
		if err != nil {
			return nil, err
		}
		costs, err := readHullValues(base + 0x0E)
		if err != nil {
			return nil, err
		}
		out.MandatoryComponents.Shields = append(out.MandatoryComponents.Shields, ruleset.ShipShield{ID: id, ComponentIndex: index, TechnologyID: technologyID, SpaceByHull: spaces, CostByHullPP: costs, Source: componentSource(base)})
	}
	fuelIDs := []string{"standard_fuel_cells", "deuterium_fuel_cells", "iridium_fuel_cells", "urridium_fuel_cells", "thorium_fuel_cells"}
	for i, id := range fuelIDs {
		index := i + 1
		base := 0x7FE8 + index*0x0A
		technologyID, err := readWord(base)
		if err != nil {
			return nil, err
		}
		rangeParsecs, err := readWord(base + 2)
		if err != nil {
			return nil, err
		}
		out.MandatoryComponents.FuelCells = append(out.MandatoryComponents.FuelCells, ruleset.ShipFuelCell{ID: id, ComponentIndex: index, TechnologyID: technologyID, RangeParsecs: rangeParsecs, Source: componentSource(base)})
	}

	if err := out.Validate(); err != nil {
		return nil, err
	}
	if err := english.Validate(); err != nil {
		return nil, err
	}
	return &ShipHullsBundle{Rules: out, English: english}, nil
}

func intPointer(value int) *int {
	return &value
}

func findShipHullNameSequence(runs []textscan.String) (int, error) {
	for start := 0; start+len(shipHullSpecs) <= len(runs); start++ {
		match := true
		for index, spec := range shipHullSpecs {
			if runs[start+index].Value != spec.Name {
				match = false
				break
			}
		}
		if match {
			return start, nil
		}
	}
	return 0, fmt.Errorf("TECHNAME block 0 consecutive player hull-name sequence not found")
}
