package moo2data

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"

	"moox/internal/lbx"
	"moox/internal/moo2exe"
	"moox/internal/moo2gfx"
	"moox/internal/ruleset"
)

const (
	racePortraitsSourceID              = "moo2-1.31-racesel-portraits"
	raceIconsSourceID                  = "moo2-1.31-raceicon-matrix"
	raceIconUsageSourceID              = "openmoo2-race-role-usage"
	openMOO2GraphicMapSourceID         = "openmoo2-graphic-map"
	moo2BuildingAnchorSourceID         = "moo2-graphics-building-anchor"
	buildingGraphicsFormulaSourceID    = "moo2-1.31-orion2-building-graphics-formula"
	buildingArchiveFormatSourceID      = "moo2-1.31-estrings-building-archive-format"
	cacheLoadBldgCodeSHA256            = "c3f6d4f657563a83c6f9b2643dbd5baa711f1019eaef66576de1fcc8296e7162"
	bldgCoordsEffectiveFrameCodeSHA256 = "4a3023201766b2797af90d45149739dc912498cee72f7ba929e05ce7b724c678"
	shipStrategicGraphicsSourceID      = "moo2-1.31-orion2-strategic-ship-graphics"
	shipStrategicArchiveSourceID       = "moo2-1.31-ships-strategic-graphics"
	shipGetPictureCodeSHA256           = "af7430b189af870cec1848fea4649c27dfb00d8af6c7f737a5c611f6c54075ef"
	shipPaletteCodeSHA256              = "a78aa3af973e4f20a1bbc4c844dc094d24b8cfc75a711e845cc9d34215b3c1bb"
	colonyShipDesignCodeSHA256         = "3e468fc441a98fccc8214b88c60daf661ab01b31431d988155f5d785d3ee56cc"
	outpostShipDesignCodeSHA256        = "b5a6c0dbe3e1a8aa471bc3ba71d07f05631ebfc67fb9974b1e17812878156c9d"
	transportShipDesignCodeSHA256      = "bddf60805ca545e5ee0580426fcb34c76a840d3c14279be4780c03ebb6a0e17b"
	shipTacticalGraphicsSourceID       = "moo2-1.31-orion2-tactical-ship-graphics"
	shipTacticalArchiveSourceID        = "moo2-1.31-cmbtshp-tactical-graphics"
	loadCombatShipCodeSHA256           = "27ca9b7efcbc52930d68c4de81fc42d4f7b27aaf6942e4f096ff1a143bd20a3d"
	drawShipCodeSHA256                 = "d7ac88d365eba8857bf93fb1e7c7b4bab587828502343346ac878ee0c87bce0e"
	drawShipToBitmapCodeSHA256         = "61ea32e4db51f981e77f30b73cc412e8e83ac933d0fecc3ac5f3e031603511ff"
	combatShipPaletteCodeSHA256        = "5676f51a6b7233a14a6d1efed4d10d9158fa30cab740c042039e8fb4083ccc64"
	loadIndividualShipPicturesSHA256   = "73532630ee0c7225343a07ec2c061f690f6a94b0f3f068612980ca28fbaf3784"
)

type raceIconRoleSpec struct {
	Role    string
	Variant int
}

var confirmedRaceIconRoles = []raceIconRoleSpec{
	{Role: "farmer", Variant: 1},
	{Role: "worker", Variant: 3},
	{Role: "scientist", Variant: 5},
	{Role: "marine", Variant: 7},
}

type buildingGraphicsEvidence struct {
	ExecutableSHA256 string
	EStringsSHA256   string
	EStringsBlockSHA string
}

type shipStrategicGraphicsEvidence struct {
	ExecutableSHA256  string
	ShipsSHA256       string
	CombatShipsSHA256 string
}

func DecodeAssets(installationRoot string, races *ruleset.RacesFile, buildings *ruleset.BuildingsFile, shipHulls *ruleset.ShipHullsFile) (*ruleset.AssetsFile, error) {
	buildingEvidence, err := verifyOriginalBuildingGraphicsFormula(installationRoot)
	if err != nil {
		return nil, err
	}
	shipEvidence, err := verifyOriginalShipStrategicGraphics(installationRoot)
	if err != nil {
		return nil, err
	}
	return decodeAssetsWithEvidence(installationRoot, races, buildings, shipHulls, buildingEvidence, shipEvidence)
}

func decodeAssetsWithEvidence(installationRoot string, races *ruleset.RacesFile, buildings *ruleset.BuildingsFile, shipHulls *ruleset.ShipHullsFile, buildingEvidence buildingGraphicsEvidence, shipEvidence shipStrategicGraphicsEvidence) (*ruleset.AssetsFile, error) {
	if races == nil {
		return nil, fmt.Errorf("races are required")
	}
	if buildings == nil {
		return nil, fmt.Errorf("buildings are required")
	}
	if err := buildings.Validate(); err != nil {
		return nil, fmt.Errorf("validate buildings: %w", err)
	}
	if shipHulls == nil {
		return nil, fmt.Errorf("ship hulls are required")
	}
	if err := shipHulls.Validate(); err != nil {
		return nil, fmt.Errorf("validate ship hulls: %w", err)
	}
	if len(races.Races) != 13 {
		return nil, fmt.Errorf("expected 13 preset races, got %d", len(races.Races))
	}

	ordered := append([]ruleset.Race(nil), races.Races...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Order < ordered[j].Order })
	for order, race := range ordered {
		if race.Order != order {
			return nil, fmt.Errorf("race %q has order %d, expected %d", race.ID, race.Order, order)
		}
	}

	raceSelPath := filepath.Join(installationRoot, "RACESEL.LBX")
	raceSel, err := lbx.Open(raceSelPath)
	if err != nil {
		return nil, fmt.Errorf("open RACESEL.LBX: %w", err)
	}
	raceIconPath := filepath.Join(installationRoot, "RACEICON.LBX")
	raceIcon, err := lbx.Open(raceIconPath)
	if err != nil {
		return nil, fmt.Errorf("open RACEICON.LBX: %w", err)
	}
	if len(raceSel.Entries) <= 28 {
		return nil, fmt.Errorf("RACESEL.LBX has %d blocks, expected at least 29", len(raceSel.Entries))
	}
	if len(raceIcon.Entries) < 169 {
		return nil, fmt.Errorf("RACEICON.LBX has %d blocks, expected at least 169", len(raceIcon.Entries))
	}

	raceSelHash, err := sha256File(raceSelPath)
	if err != nil {
		return nil, err
	}
	raceIconHash, err := sha256File(raceIconPath)
	if err != nil {
		return nil, err
	}

	out := &ruleset.AssetsFile{
		SchemaVersion: ruleset.AssetsSchemaVersion,
		Ruleset:       races.Ruleset,
		Sources: []ruleset.Source{
			{
				ID:          racePortraitsSourceID,
				Type:        "original-observed",
				Description: "RACESEL.LBX blocks 15..27 visually match the 13 preset races in canonical race order; block 28 is the custom-race portrait",
				Archive:     "RACESEL.LBX",
				SHA256:      raceSelHash,
			},
			{
				ID:          raceIconsSourceID,
				Type:        "original-observed",
				Description: "RACEICON.LBX blocks 0..168 form a 13 race x 13 variant matrix in canonical race order",
				Archive:     "RACEICON.LBX",
				SHA256:      raceIconHash,
			},
			{
				ID:           raceIconUsageSourceID,
				Type:         "secondary-community",
				Description:  "OpenMOO2 colony screens use RACEICON variants 1, 3, 5 and 7 for farmer, worker, scientist and marine population icons respectively; used as factual usage evidence only",
				URL:          "https://github.com/mimi1vx/openmoo2/blob/2cd3c344aed24380390caaaa819bf7a010b8f4a2/oldmess/gui/colony_screen.py",
				AccessedDate: "2026-08-26",
			},
			{
				ID:           openMOO2GraphicMapSourceID,
				Type:         "secondary-community",
				Description:  "OpenMOO2 graphic.ini semantic-to-LBX mappings; selected mappings are independently revalidated against the local 1.31 archive structure and block hashes",
				URL:          "https://github.com/mimi1vx/openmoo2/blob/2cd3c344aed24380390caaaa819bf7a010b8f4a2/data/graphic.ini",
				AccessedDate: "2026-08-26",
			},
			{
				ID:           moo2BuildingAnchorSourceID,
				Type:         "secondary-reference",
				Description:  "Published MOO2 graphics-format notes identify BLDG0.LBX block 0 as the Alien Management Center on the colony screen",
				URL:          "https://masteroforion2.blogspot.com/2008/04/moo2-graphics.html",
				AccessedDate: "2026-08-26",
			},
			{
				ID:          buildingGraphicsFormulaSourceID,
				Type:        "original-observed",
				Description: "Orion2.exe 1.31 Cache_Load_Bldg_ (LE object 1 offset 0x9F6DC) computes (building_id-1)/10 for BLDG archive selection and remainder*36 for the building group, then calls Bldg_Coords_To_Effective_Frame_ (offset 0xAC8A6), which maps a 6x6 colony grid to frames 0..35 in serpentine row order. Function byte ranges are hash-validated before generation.",
				Archive:     "Orion2.exe",
				SHA256:      buildingEvidence.ExecutableSHA256,
			},
			{
				ID:          buildingArchiveFormatSourceID,
				Type:        "original-observed",
				Description: "ESTRINGS.LBX block 0 offset 0xE0F contains the archive format string BLDG%D.LBX referenced by Cache_Load_Bldg_",
				Archive:     "ESTRINGS.LBX",
				SHA256:      buildingEvidence.EStringsSHA256,
				BlockSHA256: buildingEvidence.EStringsBlockSHA,
			},
		},
	}

	out.Sources = append(out.Sources,
		ruleset.Source{
			ID:          shipStrategicGraphicsSourceID,
			Type:        "original-observed",
			Description: "Orion2.exe 1.31 Get_Ship_Picture_Seg_ selects SHIPS.LBX block color_index*50+picture_id and Load_Player_Ship_Palette_ selects color_index*50+49. Colony, Outpost and Transport design functions assign strategic picture IDs 45, 46 and 47. All referenced function ranges are hash-validated before generation.",
			Archive:     "Orion2.exe",
			SHA256:      shipEvidence.ExecutableSHA256,
		},
		ruleset.Source{
			ID:          shipStrategicArchiveSourceID,
			Type:        "original-observed",
			Description: "SHIPS.LBX strategic ship graphics; player colors 0..7 occupy 50-slot groups. Standard strategic picture blocks are one-frame graphics; slot 49 is the per-color palette selected by the original executable.",
			Archive:     "SHIPS.LBX",
			SHA256:      shipEvidence.ShipsSHA256,
		},
	)

	out.Sources = append(out.Sources,
		ruleset.Source{
			ID:          shipTacticalGraphicsSourceID,
			Type:        "original-observed",
			Description: "Orion2.exe 1.31 copies the design picture ID into the combat-ship picture field (Doom Star hard-coded to 43), selects CMBTSHP as color_index*45+picture_id, and draws the 20 frames as five folded orientation indices times four animation phases. All relevant code ranges are hash-validated before generation.",
			Archive:     "Orion2.exe",
			SHA256:      shipEvidence.ExecutableSHA256,
		},
		ruleset.Source{
			ID:          shipTacticalArchiveSourceID,
			Type:        "original-observed",
			Description: "CMBTSHP.LBX contains 8 player-color groups of 45 entries; picture slots 0..43 are 59x60 graphics with 20 frames and slot 44 is the per-color combat palette.",
			Archive:     "CMBTSHP.LBX",
			SHA256:      shipEvidence.CombatShipsSHA256,
		},
	)

	for _, race := range ordered {
		portraitBlock := 15 + race.Order
		portraitRef, err := assetGraphicReference(raceSel, "RACESEL.LBX", portraitBlock, 0, 290, 322)
		if err != nil {
			return nil, fmt.Errorf("race %s portrait: %w", race.ID, err)
		}
		out.Assets = append(out.Assets, ruleset.Asset{
			Key:          race.PortraitAssetKey,
			Kind:         "race_portrait",
			Status:       "confirmed",
			Verification: "original-order-visual-confirmed",
			Reference:    portraitRef,
		})

		out.Assets = append(out.Assets, ruleset.Asset{
			Key:              race.IconAssetKey,
			Kind:             "race_icon",
			Status:           "pending",
			Verification:     "matrix-confirmed-role-unspecified",
			UnresolvedReason: "RACEICON contains 13 context/role variants per race; no current evidence identifies one variant as the canonical generic icon",
		})

		for _, role := range confirmedRaceIconRoles {
			block := race.Order*13 + role.Variant
			ref, err := assetGraphicReference(raceIcon, "RACEICON.LBX", block, 0, 28, 28)
			if err != nil {
				return nil, fmt.Errorf("race %s %s icon: %w", race.ID, role.Role, err)
			}
			out.Assets = append(out.Assets, ruleset.Asset{
				Key:          race.IconAssetKey + "." + role.Role,
				Kind:         "race_role_icon",
				Status:       "confirmed",
				Verification: "original-matrix-plus-secondary-usage-confirmed",
				Reference:    ref,
			})
		}
	}

	customRef, err := assetGraphicReference(raceSel, "RACESEL.LBX", 28, 0, 290, 322)
	if err != nil {
		return nil, fmt.Errorf("custom race portrait: %w", err)
	}
	out.Assets = append(out.Assets, ruleset.Asset{
		Key:          "race.custom.portrait",
		Kind:         "race_portrait",
		Status:       "confirmed",
		Verification: "original-sequence-visual-confirmed",
		Reference:    customRef,
	})

	if err := appendKnownSemanticReferences(installationRoot, out); err != nil {
		return nil, err
	}
	if err := appendBuildingColonyAssets(installationRoot, out, buildings); err != nil {
		return nil, err
	}
	if err := appendShipStrategicAssets(installationRoot, out, shipHulls); err != nil {
		return nil, err
	}
	if err := appendShipTacticalAssets(installationRoot, out, shipHulls); err != nil {
		return nil, err
	}
	if err := out.ValidateAgainstRaces(races); err != nil {
		return nil, err
	}
	if err := out.ValidateAgainstBuildings(buildings); err != nil {
		return nil, err
	}
	if err := out.ValidateAgainstShipHulls(shipHulls); err != nil {
		return nil, err
	}
	return out, nil
}

func verifyOriginalShipStrategicGraphics(installationRoot string) (shipStrategicGraphicsEvidence, error) {
	var evidence shipStrategicGraphicsEvidence
	exe, err := moo2exe.Open(filepath.Join(installationRoot, "Orion2.exe"))
	if err != nil {
		return evidence, fmt.Errorf("open Orion2.exe for ship graphics: %w", err)
	}
	checks := []struct {
		Name   string
		Offset int
		Size   int
		SHA256 string
	}{
		{Name: "Get_Ship_Picture_Seg_", Offset: 0x48697, Size: 0x3C, SHA256: shipGetPictureCodeSHA256},
		{Name: "Load_Player_Ship_Palette_", Offset: 0x485E0, Size: 0x98, SHA256: shipPaletteCodeSHA256},
		{Name: "Load_Colony_Ship_Design_", Offset: 0x464CD, Size: 0xB9, SHA256: colonyShipDesignCodeSHA256},
		{Name: "Load_Outpost_Ship_Design_", Offset: 0x46586, Size: 0xB8, SHA256: outpostShipDesignCodeSHA256},
		{Name: "Load_Transport_Ship_Design_", Offset: 0x4663E, Size: 0xE8, SHA256: transportShipDesignCodeSHA256},
		{Name: "Load_Combat_Ship_", Offset: 0x3954A, Size: 0x4F7, SHA256: loadCombatShipCodeSHA256},
		{Name: "Draw_Ship_", Offset: 0x20062, Size: 0x5CF, SHA256: drawShipCodeSHA256},
		{Name: "Draw_Ship_To_Bitmap_", Offset: 0x22B26, Size: 0x26E, SHA256: drawShipToBitmapCodeSHA256},
		{Name: "Load_Combat_Ship_Palette_", Offset: 0x39F99, Size: 0x191, SHA256: combatShipPaletteCodeSHA256},
		{Name: "Load_Individual_Ship_Pictures_", Offset: 0x3A12A, Size: 0x3BA, SHA256: loadIndividualShipPicturesSHA256},
	}
	for _, check := range checks {
		code, err := exe.ReadObject(1, check.Offset, check.Size)
		if err != nil {
			return evidence, fmt.Errorf("read Orion2.exe %s: %w", check.Name, err)
		}
		sum := sha256.Sum256(code)
		got := hex.EncodeToString(sum[:])
		if got != check.SHA256 {
			return evidence, fmt.Errorf("Orion2.exe %s code hash=%s, expected %s for 1.31 ship graphics", check.Name, got, check.SHA256)
		}
	}
	evidence.ExecutableSHA256 = exe.SHA256()

	shipsPath := filepath.Join(installationRoot, "SHIPS.LBX")
	ships, err := lbx.Open(shipsPath)
	if err != nil {
		return evidence, fmt.Errorf("open SHIPS.LBX: %w", err)
	}
	if len(ships.Entries) != 449 {
		return evidence, fmt.Errorf("SHIPS.LBX has %d entries, expected 449 for 1.31", len(ships.Entries))
	}
	evidence.ShipsSHA256, err = sha256File(shipsPath)
	if err != nil {
		return evidence, err
	}

	combatPath := filepath.Join(installationRoot, "CMBTSHP.LBX")
	combat, err := lbx.Open(combatPath)
	if err != nil {
		return evidence, fmt.Errorf("open CMBTSHP.LBX: %w", err)
	}
	if len(combat.Entries) != 360 {
		return evidence, fmt.Errorf("CMBTSHP.LBX has %d entries, expected 360 for 1.31", len(combat.Entries))
	}
	evidence.CombatShipsSHA256, err = sha256File(combatPath)
	if err != nil {
		return evidence, err
	}
	return evidence, nil
}

func assetGraphicReferenceAnySize(archive *lbx.File, archiveName string, blockIndex, frameIndex int) (*ruleset.AssetReference, int, error) {
	if blockIndex < 0 || blockIndex >= len(archive.Entries) {
		return nil, 0, fmt.Errorf("block %d outside archive", blockIndex)
	}
	block, err := archive.ReadEntry(blockIndex)
	if err != nil {
		return nil, 0, fmt.Errorf("read block %d: %w", blockIndex, err)
	}
	graphic, err := moo2gfx.Parse(block)
	if err != nil {
		return nil, 0, fmt.Errorf("parse graphic block %d: %w", blockIndex, err)
	}
	if frameIndex < 0 || frameIndex >= graphic.FrameCount {
		return nil, 0, fmt.Errorf("frame %d outside [0,%d) for block %d", frameIndex, graphic.FrameCount, blockIndex)
	}
	sum := sha256.Sum256(block)
	return &ruleset.AssetReference{
		Archive:     archiveName,
		Block:       blockIndex,
		Frame:       frameIndex,
		BlockSHA256: hex.EncodeToString(sum[:]),
		Width:       graphic.Width,
		Height:      graphic.Height,
	}, graphic.FrameCount, nil
}

func appendShipStrategicAssets(installationRoot string, out *ruleset.AssetsFile, hulls *ruleset.ShipHullsFile) error {
	ships, err := lbx.Open(filepath.Join(installationRoot, "SHIPS.LBX"))
	if err != nil {
		return fmt.Errorf("open SHIPS.LBX for strategic ship assets: %w", err)
	}
	if len(ships.Entries) != 449 {
		return fmt.Errorf("SHIPS.LBX has %d entries, expected 449", len(ships.Entries))
	}
	for _, hull := range hulls.Hulls {
		asset := ruleset.Asset{Key: hull.StrategicAssetKey, Kind: "ship_hull_strategic_set", Status: "confirmed", Verification: "original-executable-player-color-50-slot-picture-formula", Variants: make([]ruleset.AssetVariant, 0, len(hull.StrategicPictureIDs)*8)}
		for colorIndex := 0; colorIndex < 8; colorIndex++ {
			for styleIndex, pictureID := range hull.StrategicPictureIDs {
				blockIndex := colorIndex*50 + pictureID
				ref, frameCount, err := assetGraphicReferenceAnySize(ships, "SHIPS.LBX", blockIndex, 0)
				if err != nil {
					return fmt.Errorf("ship hull %s color %d picture %d: %w", hull.ID, colorIndex, pictureID, err)
				}
				if frameCount != 1 {
					return fmt.Errorf("ship hull %s color %d picture %d has %d frames, expected 1 strategic frame", hull.ID, colorIndex, pictureID, frameCount)
				}
				wantHeight := 48
				if pictureID == 43 {
					wantHeight = 52
				}
				if ref.Width != 52 || ref.Height != wantHeight {
					return fmt.Errorf("ship hull %s color %d picture %d is %dx%d, expected 52x%d", hull.ID, colorIndex, pictureID, ref.Width, ref.Height, wantHeight)
				}
				asset.Variants = append(asset.Variants, ruleset.AssetVariant{ID: fmt.Sprintf("color_%d_style_%d", colorIndex, styleIndex), Reference: *ref, Metadata: map[string]int{"color_index": colorIndex, "style_index": styleIndex, "picture_id": pictureID}})
			}
		}
		out.Assets = append(out.Assets, asset)
	}
	civilian := []struct {
		Key       string
		PictureID int
	}{{"ship.colony.strategic", 45}, {"ship.outpost.strategic", 46}, {"ship.transport.strategic", 47}}
	for _, spec := range civilian {
		asset := ruleset.Asset{Key: spec.Key, Kind: "ship_civilian_strategic_set", Status: "confirmed", Verification: "original-executable-civilian-picture-id-plus-player-color-50-slot-formula", Variants: make([]ruleset.AssetVariant, 0, 8)}
		for colorIndex := 0; colorIndex < 8; colorIndex++ {
			blockIndex := colorIndex*50 + spec.PictureID
			ref, frameCount, err := assetGraphicReferenceAnySize(ships, "SHIPS.LBX", blockIndex, 0)
			if err != nil {
				return fmt.Errorf("civilian ship %s color %d: %w", spec.Key, colorIndex, err)
			}
			if frameCount != 1 || ref.Width != 52 || ref.Height != 48 {
				return fmt.Errorf("civilian ship %s color %d graphic is %dx%d frames=%d, expected 52x48 frames=1", spec.Key, colorIndex, ref.Width, ref.Height, frameCount)
			}
			asset.Variants = append(asset.Variants, ruleset.AssetVariant{ID: fmt.Sprintf("color_%d", colorIndex), Reference: *ref, Metadata: map[string]int{"color_index": colorIndex, "picture_id": spec.PictureID}})
		}
		out.Assets = append(out.Assets, asset)
	}
	return nil
}

func appendShipTacticalAssets(installationRoot string, out *ruleset.AssetsFile, hulls *ruleset.ShipHullsFile) error {
	combat, err := lbx.Open(filepath.Join(installationRoot, "CMBTSHP.LBX"))
	if err != nil {
		return fmt.Errorf("open CMBTSHP.LBX for tactical ship assets: %w", err)
	}
	if len(combat.Entries) != 360 {
		return fmt.Errorf("CMBTSHP.LBX has %d entries, expected 360", len(combat.Entries))
	}
	for _, hull := range hulls.Hulls {
		asset := ruleset.Asset{Key: hull.TacticalAssetKey, Kind: "ship_hull_tactical_set", Status: "confirmed", Verification: "original-executable-combat-picture-link-plus-45-slot-color-and-5x4-frame-formula", Variants: make([]ruleset.AssetVariant, 0, len(hull.StrategicPictureIDs)*8*20)}
		for colorIndex := 0; colorIndex < 8; colorIndex++ {
			for styleIndex, pictureID := range hull.StrategicPictureIDs {
				if pictureID < 0 || pictureID > 43 {
					return fmt.Errorf("ship hull %s tactical picture id %d outside CMBTSHP picture range [0,43]", hull.ID, pictureID)
				}
				blockIndex := colorIndex*45 + pictureID
				block, err := combat.ReadEntry(blockIndex)
				if err != nil {
					return fmt.Errorf("ship hull %s tactical block %d: %w", hull.ID, blockIndex, err)
				}
				graphic, err := moo2gfx.Parse(block)
				if err != nil {
					return fmt.Errorf("ship hull %s tactical block %d parse: %w", hull.ID, blockIndex, err)
				}
				if graphic.Width != 59 || graphic.Height != 60 || graphic.FrameCount != 20 {
					return fmt.Errorf("ship hull %s tactical block %d is %dx%d frames=%d, expected 59x60 frames=20", hull.ID, blockIndex, graphic.Width, graphic.Height, graphic.FrameCount)
				}
				sum := sha256.Sum256(block)
				blockSHA := hex.EncodeToString(sum[:])
				for orientationIndex := 0; orientationIndex < 5; orientationIndex++ {
					for phase := 0; phase < 4; phase++ {
						frame := orientationIndex*4 + phase
						asset.Variants = append(asset.Variants, ruleset.AssetVariant{
							ID:        fmt.Sprintf("color_%d_style_%d_orientation_%d_phase_%d", colorIndex, styleIndex, orientationIndex, phase),
							Reference: ruleset.AssetReference{Archive: "CMBTSHP.LBX", Block: blockIndex, Frame: frame, BlockSHA256: blockSHA, Width: graphic.Width, Height: graphic.Height},
							Metadata:  map[string]int{"color_index": colorIndex, "style_index": styleIndex, "picture_id": pictureID, "orientation_index": orientationIndex, "animation_phase": phase},
						})
					}
				}
			}
		}
		out.Assets = append(out.Assets, asset)
	}
	return nil
}

func verifyOriginalBuildingGraphicsFormula(installationRoot string) (buildingGraphicsEvidence, error) {
	var evidence buildingGraphicsEvidence
	exe, err := moo2exe.Open(filepath.Join(installationRoot, "Orion2.exe"))
	if err != nil {
		return evidence, fmt.Errorf("open Orion2.exe for building graphics formula: %w", err)
	}
	checks := []struct {
		Name   string
		Offset int
		Size   int
		SHA256 string
	}{
		{Name: "Cache_Load_Bldg_", Offset: 0x9F6DC, Size: 0x66, SHA256: cacheLoadBldgCodeSHA256},
		{Name: "Bldg_Coords_To_Effective_Frame_", Offset: 0xAC8A6, Size: 0x29, SHA256: bldgCoordsEffectiveFrameCodeSHA256},
	}
	for _, check := range checks {
		code, err := exe.ReadObject(1, check.Offset, check.Size)
		if err != nil {
			return evidence, fmt.Errorf("read Orion2.exe %s: %w", check.Name, err)
		}
		sum := sha256.Sum256(code)
		got := hex.EncodeToString(sum[:])
		if got != check.SHA256 {
			return evidence, fmt.Errorf("Orion2.exe %s code hash=%s, expected %s for 1.31 building graphics formula", check.Name, got, check.SHA256)
		}
	}
	evidence.ExecutableSHA256 = exe.SHA256()

	estringsPath := filepath.Join(installationRoot, "ESTRINGS.LBX")
	estrings, err := lbx.Open(estringsPath)
	if err != nil {
		return evidence, fmt.Errorf("open ESTRINGS.LBX for building archive format: %w", err)
	}
	block, err := estrings.ReadEntry(0)
	if err != nil {
		return evidence, fmt.Errorf("read ESTRINGS.LBX block 0: %w", err)
	}
	const formatOffset = 0xE0F
	const formatString = "BLDG%D.LBX"
	if formatOffset+len(formatString) > len(block) || string(block[formatOffset:formatOffset+len(formatString)]) != formatString {
		return evidence, fmt.Errorf("ESTRINGS.LBX block 0 does not contain %q at 0x%X", formatString, formatOffset)
	}
	evidence.EStringsSHA256, err = sha256File(estringsPath)
	if err != nil {
		return evidence, err
	}
	blockSum := sha256.Sum256(block)
	evidence.EStringsBlockSHA = hex.EncodeToString(blockSum[:])
	return evidence, nil
}

func buildingEffectiveFrame(x, y int) (int, error) {
	if x < 0 || x >= 6 || y < 0 || y >= 6 {
		return 0, fmt.Errorf("building grid coordinate (%d,%d) outside 6x6", x, y)
	}
	base := y * 6
	if y%2 == 0 {
		return base + x, nil
	}
	return base + 5 - x, nil
}

func buildingArchiveBlock(buildingID, x, y int) (string, int, int, error) {
	if buildingID < 1 || buildingID > 48 {
		return "", 0, 0, fmt.Errorf("building id %d outside [1,48]", buildingID)
	}
	effectiveFrame, err := buildingEffectiveFrame(x, y)
	if err != nil {
		return "", 0, 0, err
	}
	zero := buildingID - 1
	archiveIndex := zero / 10
	groupWithinArchive := zero % 10
	return fmt.Sprintf("BLDG%d.LBX", archiveIndex), groupWithinArchive*36 + effectiveFrame, effectiveFrame, nil
}

func appendBuildingColonyAssets(installationRoot string, out *ruleset.AssetsFile, buildings *ruleset.BuildingsFile) error {
	archives := make(map[string]*lbx.File)
	for _, building := range buildings.Buildings {
		asset := ruleset.Asset{
			Key:          building.ColonyReferenceAssetKey,
			Kind:         "building_colony_set",
			Status:       "confirmed",
			Verification: "original-executable-building-archive-group-and-serpentine-frame-formula",
			Variants:     make([]ruleset.AssetVariant, 0, 36),
		}
		for y := 0; y < 6; y++ {
			for x := 0; x < 6; x++ {
				archiveName, blockIndex, effectiveFrame, err := buildingArchiveBlock(building.ProductionID, x, y)
				if err != nil {
					return fmt.Errorf("building %s: %w", building.ID, err)
				}
				archive := archives[archiveName]
				if archive == nil {
					opened, err := lbx.Open(filepath.Join(installationRoot, archiveName))
					if err != nil {
						return fmt.Errorf("open %s for building %s: %w", archiveName, building.ID, err)
					}
					archive = opened
					archives[archiveName] = archive
				}
				ref, err := assetGraphicReference(archive, archiveName, blockIndex, 0, 640, 480)
				if err != nil {
					return fmt.Errorf("building %s grid (%d,%d): %w", building.ID, x, y, err)
				}
				asset.Variants = append(asset.Variants, ruleset.AssetVariant{
					ID:        fmt.Sprintf("grid_%d_%d", x, y),
					Reference: *ref,
					Metadata: map[string]int{
						"grid_x":          x,
						"grid_y":          y,
						"effective_frame": effectiveFrame,
					},
				})
			}
		}
		out.Assets = append(out.Assets, asset)
	}
	return nil
}
func appendKnownSemanticReferences(installationRoot string, out *ruleset.AssetsFile) error {
	type knownAsset struct {
		Key     string
		Kind    string
		Archive string
		Block   int
		Frame   int
		Width   int
		Height  int
		Verify  string
	}
	known := []knownAsset{
		{Key: "ui.main_menu.splash", Kind: "ui_reference", Archive: "MAINMENU.LBX", Block: 0, Frame: 0, Width: 640, Height: 480, Verify: "secondary-render-map-local-structure-confirmed"},
		{Key: "ui.main_screen.panel", Kind: "ui_panel", Archive: "BUFFER0.LBX", Block: 0, Frame: 0, Width: 640, Height: 480, Verify: "secondary-render-map-local-structure-confirmed"},
		{Key: "ui.colonies.panel", Kind: "ui_panel", Archive: "COLSUM.LBX", Block: 0, Frame: 0, Width: 640, Height: 480, Verify: "secondary-render-map-local-structure-confirmed"},
		{Key: "ui.colony.panel", Kind: "ui_panel", Archive: "COLPUPS.LBX", Block: 5, Frame: 0, Width: 640, Height: 480, Verify: "secondary-render-map-local-structure-confirmed"},
		{Key: "ui.background.starfield", Kind: "ui_background", Archive: "COLONY2.LBX", Block: 49, Frame: 0, Width: 640, Height: 480, Verify: "secondary-render-map-local-structure-confirmed"},
		{Key: "production.food.1", Kind: "production_icon", Archive: "COLONY2.LBX", Block: 0, Frame: 0, Width: 16, Height: 23, Verify: "secondary-render-map-local-structure-confirmed"},
		{Key: "production.industry.1", Kind: "production_icon", Archive: "COLONY2.LBX", Block: 1, Frame: 0, Width: 16, Height: 23, Verify: "secondary-render-map-local-structure-confirmed"},
		{Key: "production.research.1", Kind: "production_icon", Archive: "COLONY2.LBX", Block: 2, Frame: 0, Width: 16, Height: 23, Verify: "secondary-render-map-local-structure-confirmed"},
		{Key: "production.money.1", Kind: "production_icon", Archive: "COLONY2.LBX", Block: 3, Frame: 0, Width: 16, Height: 23, Verify: "secondary-render-map-local-structure-confirmed"},
		{Key: "production.food.10", Kind: "production_icon", Archive: "COLONY2.LBX", Block: 4, Frame: 0, Width: 16, Height: 23, Verify: "secondary-render-map-local-structure-confirmed"},
		{Key: "production.industry.10", Kind: "production_icon", Archive: "COLONY2.LBX", Block: 5, Frame: 0, Width: 16, Height: 23, Verify: "secondary-render-map-local-structure-confirmed"},
		{Key: "production.research.10", Kind: "production_icon", Archive: "COLONY2.LBX", Block: 6, Frame: 0, Width: 16, Height: 23, Verify: "secondary-render-map-local-structure-confirmed"},
		{Key: "production.money.10", Kind: "production_icon", Archive: "COLONY2.LBX", Block: 3, Frame: 0, Width: 16, Height: 23, Verify: "secondary-render-map-local-structure-confirmed"},
		{Key: "building.alien_management_center.colony_reference", Kind: "building_colony_reference", Archive: "BLDG0.LBX", Block: 0, Frame: 0, Width: 640, Height: 480, Verify: "published-building-anchor-local-structure-confirmed"},
	}

	archives := make(map[string]*lbx.File)
	for _, spec := range known {
		archive := archives[spec.Archive]
		if archive == nil {
			opened, err := lbx.Open(filepath.Join(installationRoot, spec.Archive))
			if err != nil {
				return fmt.Errorf("open %s for semantic asset %s: %w", spec.Archive, spec.Key, err)
			}
			archive = opened
			archives[spec.Archive] = archive
		}
		ref, err := assetGraphicReference(archive, spec.Archive, spec.Block, spec.Frame, spec.Width, spec.Height)
		if err != nil {
			return fmt.Errorf("semantic asset %s: %w", spec.Key, err)
		}
		out.Assets = append(out.Assets, ruleset.Asset{Key: spec.Key, Kind: spec.Kind, Status: "confirmed", Verification: spec.Verify, Reference: ref})
	}
	return nil
}

func assetGraphicReference(archive *lbx.File, archiveName string, blockIndex, frameIndex, expectedWidth, expectedHeight int) (*ruleset.AssetReference, error) {
	if blockIndex < 0 || blockIndex >= len(archive.Entries) {
		return nil, fmt.Errorf("block %d outside archive", blockIndex)
	}
	block, err := archive.ReadEntry(blockIndex)
	if err != nil {
		return nil, fmt.Errorf("read block %d: %w", blockIndex, err)
	}
	graphic, err := moo2gfx.Parse(block)
	if err != nil {
		return nil, fmt.Errorf("parse graphic block %d: %w", blockIndex, err)
	}
	if graphic.Width != expectedWidth || graphic.Height != expectedHeight {
		return nil, fmt.Errorf("block %d is %dx%d, expected %dx%d", blockIndex, graphic.Width, graphic.Height, expectedWidth, expectedHeight)
	}
	if frameIndex < 0 || frameIndex >= graphic.FrameCount {
		return nil, fmt.Errorf("frame %d outside [0,%d) for block %d", frameIndex, graphic.FrameCount, blockIndex)
	}
	sum := sha256.Sum256(block)
	return &ruleset.AssetReference{
		Archive:     archiveName,
		Block:       blockIndex,
		Frame:       frameIndex,
		BlockSHA256: hex.EncodeToString(sum[:]),
		Width:       graphic.Width,
		Height:      graphic.Height,
	}, nil
}
