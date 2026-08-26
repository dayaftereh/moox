package moo2data

import (
	"crypto/sha256"
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
		return nil, fmt.Errorf("open Orion2.exe for ship hull picture logic: %w", err)
	}
	const (
		pictureLogicObject = 1
		pictureLogicOffset = 0x51731
		pictureLogicSize   = 0x2A
	)
	logic, err := exe.ReadObject(pictureLogicObject, pictureLogicOffset, pictureLogicSize)
	if err != nil {
		return nil, fmt.Errorf("read Orion2.exe Auto_Design_Ship_ picture logic: %w", err)
	}
	logicSum := sha256.Sum256(logic)
	if got := hex.EncodeToString(logicSum[:]); got != shipHullPictureLogicSHA256 {
		return nil, fmt.Errorf("Orion2.exe Auto_Design_Ship_ picture logic hash=%s, expected %s", got, shipHullPictureLogicSHA256)
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
		out.Hulls = append(out.Hulls, ruleset.ShipHull{
			ID:                       spec.ID,
			SizeIndex:                sizeIndex,
			NameKey:                  nameKey,
			NameSource:               ruleset.FieldProvenance{SourceID: shipHullNamesSourceID, Offset: &nameOffset},
			StrategicPictureIDs:      pictureIDs,
			PictureLogicVerification: "original-techname-sequence-plus-hash-verified-auto-design-picture-logic",
			PictureLogicSource:       ruleset.FieldProvenance{SourceID: shipHullPictureLogicSourceID},
			StrategicAssetKey:        "ship_hull." + spec.ID + ".strategic",
			TacticalAssetKey:         "ship_hull." + spec.ID + ".tactical",
		})
		english.Strings[nameKey] = spec.Name
	}

	if err := out.Validate(); err != nil {
		return nil, err
	}
	if err := english.Validate(); err != nil {
		return nil, err
	}
	return &ShipHullsBundle{Rules: out, English: english}, nil
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
