package graphiccatalog

import (
	"fmt"
	"path/filepath"
	"strings"

	"moox/internal/lbx"
	"moox/internal/moo2gfx"
)

type paletteResolution struct {
	Resolved   bool
	Source     string
	Evidence   string
	Confidence string
}

type paletteResolver struct {
	sourceRoot     string
	fontsBlock2    *moo2gfx.ExternalPalette
	councilPalette *moo2gfx.Graphic
}

func newPaletteResolver(sourceRoot string) *paletteResolver {
	return &paletteResolver{sourceRoot: sourceRoot}
}

func (r *paletteResolver) Resolve(archiveRel string, blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	name := strings.ToUpper(filepath.Base(filepath.FromSlash(archiveRel)))
	switch name {
	case "BLDG0.LBX":
		palette, err := r.loadFontsBlock2()
		if err != nil {
			return paletteResolution{}, err
		}
		palette.Apply(graphic)
		return paletteResolution{
			Resolved:   true,
			Source:     "FONTS.LBX#2",
			Evidence:   "https://masteroforion2.blogspot.com/2008/04/moo2-graphics.html",
			Confidence: "confirmed",
		}, nil

	case "COUNCIL.LBX":
		if blockIndex == 0 {
			return paletteResolution{}, nil
		}
		paletteGraphic, err := r.loadCouncilPalette(archive)
		if err != nil {
			return paletteResolution{}, err
		}
		moo2gfx.ApplyPaletteFromGraphic(paletteGraphic, graphic)
		return paletteResolution{
			Resolved:   true,
			Source:     "COUNCIL.LBX#0",
			Evidence:   "https://www.spheriumnorth.com/orion-forum/nfphpbb/viewtopic.php?t=377",
			Confidence: "confirmed",
		}, nil
	}
	return paletteResolution{}, nil
}

func (r *paletteResolver) loadFontsBlock2() (*moo2gfx.ExternalPalette, error) {
	if r.fontsBlock2 != nil {
		return r.fontsBlock2, nil
	}
	archive, err := lbx.Open(filepath.Join(r.sourceRoot, "FONTS.LBX"))
	if err != nil {
		return nil, fmt.Errorf("open FONTS.LBX for BLDG0 palette: %w", err)
	}
	if len(archive.Entries) <= 2 {
		return nil, fmt.Errorf("FONTS.LBX does not contain palette block 2")
	}
	data, err := archive.ReadEntry(2)
	if err != nil {
		return nil, fmt.Errorf("read FONTS.LBX block 2: %w", err)
	}
	palette, err := moo2gfx.ParseExternalPalette(data)
	if err != nil {
		return nil, fmt.Errorf("parse FONTS.LBX block 2 palette: %w", err)
	}
	r.fontsBlock2 = palette
	return palette, nil
}

func (r *paletteResolver) loadCouncilPalette(archive *lbx.File) (*moo2gfx.Graphic, error) {
	if r.councilPalette != nil {
		return r.councilPalette, nil
	}
	if len(archive.Entries) == 0 {
		return nil, fmt.Errorf("COUNCIL.LBX does not contain block 0")
	}
	data, err := archive.ReadEntry(0)
	if err != nil {
		return nil, fmt.Errorf("read COUNCIL.LBX block 0 palette carrier: %w", err)
	}
	graphic, err := moo2gfx.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse COUNCIL.LBX block 0 palette carrier: %w", err)
	}
	if graphic.Flags&moo2gfx.FlagInternalPalette == 0 || paletteEntries(graphic) != 256 {
		return nil, fmt.Errorf("COUNCIL.LBX block 0 is not the expected full internal palette carrier")
	}
	r.councilPalette = graphic
	return graphic, nil
}
