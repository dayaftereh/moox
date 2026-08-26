package graphiccatalog

import (
	"fmt"
	"path/filepath"
	"strings"

	"moox/internal/lbx"
	"moox/internal/moo2gfx"
)

const (
	wolverineGraphicsEvidence = "https://masteroforion2.blogspot.com/2008/04/moo2-graphics.html"
	orionNebulaEvidence       = "https://www.spheriumnorth.com/orion-forum/nfphpbb/viewtopic.php?t=377"
	moo2WorkshopEvidence      = "https://moo2mod.com/doc/history/moo2_workshop.html"
)

type paletteResolution struct {
	Resolved   bool
	Source     string
	Evidence   string
	Confidence string
}

type externalPaletteRule struct {
	Archive      string
	PaletteBlock int
	Evidence     string
	ExternalOnly bool
}

var confirmedExternalPaletteRules = map[string]externalPaletteRule{
	"BLDG0.LBX":    {Archive: "FONTS.LBX", PaletteBlock: 2, Evidence: wolverineGraphicsEvidence, ExternalOnly: true},
	"BLDG1.LBX":    {Archive: "FONTS.LBX", PaletteBlock: 2, Evidence: moo2WorkshopEvidence, ExternalOnly: true},
	"BLDG2.LBX":    {Archive: "FONTS.LBX", PaletteBlock: 2, Evidence: moo2WorkshopEvidence, ExternalOnly: true},
	"BLDG3.LBX":    {Archive: "FONTS.LBX", PaletteBlock: 2, Evidence: moo2WorkshopEvidence, ExternalOnly: true},
	"BLDG4.LBX":    {Archive: "FONTS.LBX", PaletteBlock: 2, Evidence: moo2WorkshopEvidence, ExternalOnly: true},
	"RACEICON.LBX": {Archive: "FONTS.LBX", PaletteBlock: 2, Evidence: moo2WorkshopEvidence, ExternalOnly: true},
	"DESIGN.LBX":   {Archive: "FONTS.LBX", PaletteBlock: 5, Evidence: moo2WorkshopEvidence, ExternalOnly: true},
	"MAINMENU.LBX": {Archive: "FONTS.LBX", PaletteBlock: 6, Evidence: moo2WorkshopEvidence, ExternalOnly: true},
	"CMBTMISL.LBX": {Archive: "FONTS.LBX", PaletteBlock: 4, Evidence: moo2WorkshopEvidence, ExternalOnly: true},
}

type paletteResolver struct {
	sourceRoot         string
	externalPalettes   map[string]*moo2gfx.ExternalPalette
	councilPalette     *moo2gfx.Graphic
	shipCarriers       map[int]*moo2gfx.Graphic
	combatShipCarriers map[int]*moo2gfx.Graphic
	beamsCarrier       *moo2gfx.Graphic
}

func newPaletteResolver(sourceRoot string) *paletteResolver {
	return &paletteResolver{
		sourceRoot:         sourceRoot,
		externalPalettes:   make(map[string]*moo2gfx.ExternalPalette),
		shipCarriers:       make(map[int]*moo2gfx.Graphic),
		combatShipCarriers: make(map[int]*moo2gfx.Graphic),
	}
}

func (r *paletteResolver) Resolve(archiveRel string, blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	name := strings.ToUpper(filepath.Base(filepath.FromSlash(archiveRel)))

	if rule, ok := confirmedExternalPaletteRules[name]; ok {
		if !rule.ExternalOnly || graphic.Flags&moo2gfx.FlagInternalPalette == 0 {
			palette, err := r.loadExternalPalette(rule.Archive, rule.PaletteBlock)
			if err != nil {
				return paletteResolution{}, err
			}
			palette.Apply(graphic)
			return paletteResolution{
				Resolved:   true,
				Source:     fmt.Sprintf("%s#%d", rule.Archive, rule.PaletteBlock),
				Evidence:   rule.Evidence,
				Confidence: "confirmed",
			}, nil
		}
	}

	switch name {
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
			Evidence:   orionNebulaEvidence,
			Confidence: "confirmed",
		}, nil

	case "SHIPS.LBX":
		return r.resolveShips(blockIndex, archive, graphic)

	case "CMBTSHP.LBX":
		return r.resolveCombatShips(blockIndex, archive, graphic)

	case "CMBTSFX.LBX":
		return r.resolveCombatSFX(blockIndex, graphic)

	case "BEAMS.LBX":
		return r.resolveBeams(blockIndex, archive, graphic)
	}
	return paletteResolution{}, nil
}

func (r *paletteResolver) resolveCombatSFX(blockIndex int, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	paletteBlock := combatSFXPaletteBlock(blockIndex)
	if paletteBlock < 0 {
		return paletteResolution{}, nil
	}
	palette, err := r.loadExternalPalette("FONTS.LBX", paletteBlock)
	if err != nil {
		return paletteResolution{}, err
	}
	palette.Apply(graphic)
	return paletteResolution{
		Resolved:   true,
		Source:     fmt.Sprintf("FONTS.LBX#%d", paletteBlock),
		Evidence:   moo2WorkshopEvidence,
		Confidence: "confirmed",
	}, nil
}

func combatSFXPaletteBlock(block int) int {
	switch {
	case block >= 2 && block <= 7:
		return 4
	case block == 8:
		return 1
	case block >= 9 && block <= 13:
		return 2
	case block >= 14 && block <= 15:
		return 4
	case block >= 16 && block <= 39:
		return 2
	case block >= 41 && block <= 42:
		return 2
	case block >= 43 && block <= 46:
		return 4
	case block >= 48 && block <= 51:
		return 4
	case block >= 52 && block <= 67:
		return 2
	case block >= 69 && block <= 78:
		return 1
	default:
		return -1
	}
}

func (r *paletteResolver) resolveBeams(blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	paletteBlock, useCarrier := beamsPaletteRule(blockIndex)
	if paletteBlock < 0 {
		return paletteResolution{}, nil
	}
	if useCarrier {
		carrier, err := r.loadBeamsCarrier(archive)
		if err != nil {
			return paletteResolution{}, err
		}
		moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
	}
	palette, err := r.loadExternalPalette("FONTS.LBX", paletteBlock)
	if err != nil {
		return paletteResolution{}, err
	}
	palette.Apply(graphic)
	source := fmt.Sprintf("FONTS.LBX#%d", paletteBlock)
	if useCarrier {
		source += " + BEAMS.LBX#67"
	}
	return paletteResolution{
		Resolved:   true,
		Source:     source,
		Evidence:   moo2WorkshopEvidence,
		Confidence: "confirmed",
	}, nil
}

func beamsPaletteRule(block int) (paletteBlock int, useCarrier bool) {
	switch {
	case block >= 1 && block <= 16:
		return 3, false
	case block >= 17 && block <= 32:
		return 1, false
	case block >= 33 && block <= 48:
		return 3, false
	case block >= 49 && block <= 64:
		return 1, false
	case block >= 65 && block <= 66:
		return 2, false
	case block == 67:
		return 4, false
	case block == 68:
		return 4, true
	case block == 69:
		return 4, false
	case block >= 70 && block <= 87:
		return 4, true
	case block >= 88 && block <= 108:
		return 4, false
	case block >= 109 && block <= 129:
		return 4, true
	case block >= 131 && block <= 152:
		return 4, true
	default:
		return -1, false
	}
}

func (r *paletteResolver) loadBeamsCarrier(archive *lbx.File) (*moo2gfx.Graphic, error) {
	if r.beamsCarrier != nil {
		return r.beamsCarrier, nil
	}
	const block = 67
	if block >= len(archive.Entries) {
		return nil, fmt.Errorf("BEAMS.LBX does not contain palette carrier block %d", block)
	}
	data, err := archive.ReadEntry(block)
	if err != nil {
		return nil, fmt.Errorf("read BEAMS.LBX palette carrier block %d: %w", block, err)
	}
	graphic, err := moo2gfx.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse BEAMS.LBX palette carrier block %d: %w", block, err)
	}
	if graphic.Flags&moo2gfx.FlagInternalPalette == 0 || paletteEntries(graphic) != 15 || graphic.PaletteShift != 241 {
		return nil, fmt.Errorf("BEAMS.LBX block %d is not the expected 15-color palette carrier at shift 241", block)
	}
	r.beamsCarrier = graphic
	return graphic, nil
}
func (r *paletteResolver) resolveCombatShips(blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	if graphic.Flags&moo2gfx.FlagInternalPalette != 0 {
		return paletteResolution{}, nil
	}
	carrier := combatShipPaletteCarrier(blockIndex)
	if carrier < 0 {
		return paletteResolution{}, nil
	}
	carrierGraphic, err := r.loadCombatShipCarrier(archive, carrier)
	if err != nil {
		return paletteResolution{}, err
	}
	moo2gfx.ApplyPaletteFromGraphic(carrierGraphic, graphic)
	base, err := r.loadExternalPalette("FONTS.LBX", 4)
	if err != nil {
		return paletteResolution{}, err
	}
	base.Apply(graphic)
	return paletteResolution{
		Resolved:   true,
		Source:     fmt.Sprintf("FONTS.LBX#4 + CMBTSHP.LBX#%d", carrier),
		Evidence:   moo2WorkshopEvidence,
		Confidence: "confirmed",
	}, nil
}

func combatShipPaletteCarrier(block int) int {
	if block < 0 || block >= 360 {
		return -1
	}
	group := block / 45
	within := block % 45
	if group >= 8 || within >= 44 {
		return -1
	}
	return group*45 + 44
}
func (r *paletteResolver) resolveShips(blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	// Internal 2x1 graphics in SHIPS are the palette carriers themselves.
	// They are useful as sources but do not need another display context here.
	if graphic.Flags&moo2gfx.FlagInternalPalette != 0 {
		return paletteResolution{}, nil
	}

	carrier := shipPaletteCarrier(blockIndex)
	if carrier < 0 {
		return paletteResolution{}, nil
	}

	carrierGraphic, err := r.loadShipCarrier(archive, carrier)
	if err != nil {
		return paletteResolution{}, err
	}
	// The local carrier supplies the color-specific high palette range first.
	// FONTS#1 then fills every still-missing base entry without overwriting it.
	moo2gfx.ApplyPaletteFromGraphic(carrierGraphic, graphic)
	base, err := r.loadExternalPalette("FONTS.LBX", 1)
	if err != nil {
		return paletteResolution{}, err
	}
	base.Apply(graphic)

	return paletteResolution{
		Resolved:   true,
		Source:     fmt.Sprintf("FONTS.LBX#1 + SHIPS.LBX#%d", carrier),
		Evidence:   moo2WorkshopEvidence,
		Confidence: "confirmed",
	}, nil
}

func shipPaletteCarrier(block int) int {
	// MoO2 Workshop's 1.31 description files explicitly pair each 49-image
	// banner-color group with the palette holder at the end of that group.
	if block >= 0 && block < 400 {
		group := block / 50
		within := block % 50
		if group < 8 && within < 49 {
			return group*50 + 49
		}
		return -1
	}

	// Special ship/monster blocks with dependencies explicitly described by
	// MoO2 Workshop. Blocks without a described dependency remain pending.
	switch block {
	case 400, 401, 402, 403, 404:
		return 413
	case 407:
		return 419
	case 408, 412, 420, 424:
		return 414
	case 409, 421:
		return 416
	case 410, 422:
		return 418
	case 411, 423:
		return 415
	default:
		return -1
	}
}

func (r *paletteResolver) loadExternalPalette(archiveName string, block int) (*moo2gfx.ExternalPalette, error) {
	key := strings.ToUpper(fmt.Sprintf("%s#%d", archiveName, block))
	if palette := r.externalPalettes[key]; palette != nil {
		return palette, nil
	}

	archive, err := lbx.Open(filepath.Join(r.sourceRoot, archiveName))
	if err != nil {
		return nil, fmt.Errorf("open %s palette source: %w", archiveName, err)
	}
	if block < 0 || block >= len(archive.Entries) {
		return nil, fmt.Errorf("%s does not contain palette block %d", archiveName, block)
	}
	data, err := archive.ReadEntry(block)
	if err != nil {
		return nil, fmt.Errorf("read %s block %d: %w", archiveName, block, err)
	}
	palette, err := moo2gfx.ParseExternalPalette(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s block %d palette: %w", archiveName, block, err)
	}
	r.externalPalettes[key] = palette
	return palette, nil
}

func (r *paletteResolver) loadCombatShipCarrier(archive *lbx.File, block int) (*moo2gfx.Graphic, error) {
	if graphic := r.combatShipCarriers[block]; graphic != nil {
		return graphic, nil
	}
	if block < 0 || block >= len(archive.Entries) {
		return nil, fmt.Errorf("CMBTSHP.LBX does not contain palette carrier block %d", block)
	}
	data, err := archive.ReadEntry(block)
	if err != nil {
		return nil, fmt.Errorf("read CMBTSHP.LBX palette carrier block %d: %w", block, err)
	}
	graphic, err := moo2gfx.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse CMBTSHP.LBX palette carrier block %d: %w", block, err)
	}
	if graphic.Flags&moo2gfx.FlagInternalPalette == 0 || paletteEntries(graphic) != 32 || graphic.PaletteShift != 32 {
		return nil, fmt.Errorf("CMBTSHP.LBX block %d is not the expected 32-color palette carrier at shift 32", block)
	}
	r.combatShipCarriers[block] = graphic
	return graphic, nil
}
func (r *paletteResolver) loadShipCarrier(archive *lbx.File, block int) (*moo2gfx.Graphic, error) {
	if graphic := r.shipCarriers[block]; graphic != nil {
		return graphic, nil
	}
	if block < 0 || block >= len(archive.Entries) {
		return nil, fmt.Errorf("SHIPS.LBX does not contain palette carrier block %d", block)
	}
	data, err := archive.ReadEntry(block)
	if err != nil {
		return nil, fmt.Errorf("read SHIPS.LBX palette carrier block %d: %w", block, err)
	}
	graphic, err := moo2gfx.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse SHIPS.LBX palette carrier block %d: %w", block, err)
	}
	if graphic.Flags&moo2gfx.FlagInternalPalette == 0 || paletteEntries(graphic) == 0 {
		return nil, fmt.Errorf("SHIPS.LBX block %d is not an internal palette carrier", block)
	}
	r.shipCarriers[block] = graphic
	return graphic, nil
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
