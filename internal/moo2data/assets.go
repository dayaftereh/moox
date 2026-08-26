package moo2data

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"

	"moox/internal/lbx"
	"moox/internal/moo2gfx"
	"moox/internal/ruleset"
)

const (
	racePortraitsSourceID      = "moo2-1.31-racesel-portraits"
	raceIconsSourceID          = "moo2-1.31-raceicon-matrix"
	raceIconUsageSourceID      = "openmoo2-race-role-usage"
	openMOO2GraphicMapSourceID = "openmoo2-graphic-map"
	moo2BuildingAnchorSourceID = "moo2-graphics-building-anchor"
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

func DecodeAssets(installationRoot string, races *ruleset.RacesFile) (*ruleset.AssetsFile, error) {
	if races == nil {
		return nil, fmt.Errorf("races are required")
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
	if err := out.ValidateAgainstRaces(races); err != nil {
		return nil, err
	}
	return out, nil
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
