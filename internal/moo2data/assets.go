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

func DecodeAssets(installationRoot string, races *ruleset.RacesFile, buildings *ruleset.BuildingsFile) (*ruleset.AssetsFile, error) {
	evidence, err := verifyOriginalBuildingGraphicsFormula(installationRoot)
	if err != nil {
		return nil, err
	}
	return decodeAssetsWithEvidence(installationRoot, races, buildings, evidence)
}

func decodeAssetsWithEvidence(installationRoot string, races *ruleset.RacesFile, buildings *ruleset.BuildingsFile, buildingEvidence buildingGraphicsEvidence) (*ruleset.AssetsFile, error) {
	if races == nil {
		return nil, fmt.Errorf("races are required")
	}
	if buildings == nil {
		return nil, fmt.Errorf("buildings are required")
	}
	if err := buildings.Validate(); err != nil {
		return nil, fmt.Errorf("validate buildings: %w", err)
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
	if err := out.ValidateAgainstRaces(races); err != nil {
		return nil, err
	}
	if err := out.ValidateAgainstBuildings(buildings); err != nil {
		return nil, err
	}
	return out, nil
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
