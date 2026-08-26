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
	"CMBTFGTR.LBX": {Archive: "FONTS.LBX", PaletteBlock: 4, Evidence: moo2WorkshopEvidence, ExternalOnly: true},
	"COLGCBT.LBX":  {Archive: "FONTS.LBX", PaletteBlock: 2, Evidence: moo2WorkshopEvidence, ExternalOnly: true},
	"COLROADS.LBX": {Archive: "FONTS.LBX", PaletteBlock: 2, Evidence: moo2WorkshopEvidence, ExternalOnly: true},
	"COLVEGGI.LBX": {Archive: "FONTS.LBX", PaletteBlock: 2, Evidence: moo2WorkshopEvidence, ExternalOnly: true},
	"STARBG.LBX":   {Archive: "FONTS.LBX", PaletteBlock: 1, Evidence: moo2WorkshopEvidence, ExternalOnly: true},
}

type paletteResolver struct {
	sourceRoot         string
	externalPalettes   map[string]*moo2gfx.ExternalPalette
	councilPalette     *moo2gfx.Graphic
	shipCarriers       map[int]*moo2gfx.Graphic
	combatShipCarriers map[int]*moo2gfx.Graphic
	beamsCarrier       *moo2gfx.Graphic
	localCarriers      map[string]*moo2gfx.Graphic
}

func newPaletteResolver(sourceRoot string) *paletteResolver {
	return &paletteResolver{
		sourceRoot:         sourceRoot,
		externalPalettes:   make(map[string]*moo2gfx.ExternalPalette),
		shipCarriers:       make(map[int]*moo2gfx.Graphic),
		combatShipCarriers: make(map[int]*moo2gfx.Graphic),
		localCarriers:      make(map[string]*moo2gfx.Graphic),
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

	case "BUFFER0.LBX":
		return r.resolveBuffer0(blockIndex, graphic)

	case "OFFICER.LBX":
		return r.resolveOfficer(blockIndex, graphic)

	case "FLEET.LBX":
		return r.resolveFleet(blockIndex, archive, graphic)

	case "DIPLOMAT.LBX":
		return r.resolveDiplomat(blockIndex, archive, graphic)

	case "MONSTER.LBX":
		return r.resolveMonster(blockIndex, archive, graphic)

	case "GSTAR.LBX":
		return r.resolveGStar(blockIndex, graphic)

	case "COLONY.LBX":
		return r.resolveColony(blockIndex, graphic)

	case "MULTIGM.LBX":
		return r.resolveMultiGame(blockIndex, archive, graphic)

	case "PLNTSUM.LBX":
		return r.resolveLocalFullPaletteRange("PLNTSUM.LBX", blockIndex, 1, 83, 0, archive, graphic)

	case "RACES.LBX":
		return r.resolveLocalFullPaletteRange("RACES.LBX", blockIndex, 1, 63, 0, archive, graphic)

	case "INFO.LBX":
		return r.resolveInfo(blockIndex, archive, graphic)

	case "COLSUM.LBX":
		return r.resolveLocalFullPaletteRange("COLSUM.LBX", blockIndex, 1, 20, 21, archive, graphic)

	case "COMBAT.LBX":
		return r.resolveCombatUI(blockIndex, archive, graphic)

	case "DIPSTARS.LBX":
		return r.resolveDipStars(blockIndex, graphic)

	case "APP_PICS.LBX":
		return r.resolveAppPics(blockIndex, graphic)

	case "MAINPUPS.LBX":
		return r.resolveMainPups(blockIndex, graphic)

	case "COLSYSDI.LBX":
		return r.resolveTwoRangeFonts(blockIndex, graphic, 5, 1, 65, 2)

	case "SYSDISP.LBX":
		return r.resolveTwoRangeFonts(blockIndex, graphic, 5, 1, 61, 2)

	case "GAME.LBX":
		return r.resolveGame(blockIndex, graphic)

	case "COLONY2.LBX":
		return r.resolveColony2(blockIndex, graphic)

	case "RACESEL.LBX":
		return r.resolveRaceSel(blockIndex, archive, graphic)

	case "TECHSEL.LBX":
		return r.resolveTechSel(blockIndex, archive, graphic)

	case "NEWGAME.LBX":
		return r.resolveNewGame(blockIndex, archive, graphic)
	}
	return paletteResolution{}, nil
}

func (r *paletteResolver) resolveLocalFullPaletteRange(archiveName string, blockIndex, first, last, carrierBlock int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	if blockIndex < first || blockIndex > last {
		return paletteResolution{}, nil
	}
	carrier, err := r.loadLocalCarrier(archiveName, archive, carrierBlock, 0, 256)
	if err != nil {
		return paletteResolution{}, err
	}
	moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
	return paletteResolution{Resolved: true, Source: fmt.Sprintf("%s#%d", archiveName, carrierBlock), Evidence: moo2WorkshopEvidence, Confidence: "confirmed"}, nil
}

func (r *paletteResolver) resolveInfo(blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	if blockIndex == 1 {
		return paletteResolution{}, nil
	}
	if blockIndex != 0 && (blockIndex < 2 || blockIndex > 25) {
		return paletteResolution{}, nil
	}
	carrier, err := r.loadLocalCarrier("INFO.LBX", archive, 1, 0, 256)
	if err != nil {
		return paletteResolution{}, err
	}
	moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
	return paletteResolution{Resolved: true, Source: "INFO.LBX#1", Evidence: moo2WorkshopEvidence, Confidence: "confirmed"}, nil
}

func (r *paletteResolver) resolveCombatUI(blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	if blockIndex >= 45 && blockIndex <= 49 {
		return r.applyExternalPalette(graphic, 4)
	}
	if (blockIndex >= 0 && blockIndex <= 10) || (blockIndex >= 12 && blockIndex <= 44) || (blockIndex >= 50 && blockIndex <= 89) {
		carrier, err := r.loadLocalCarrier("COMBAT.LBX", archive, 11, 0, 256)
		if err != nil {
			return paletteResolution{}, err
		}
		moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
		return paletteResolution{Resolved: true, Source: "COMBAT.LBX#11", Evidence: moo2WorkshopEvidence, Confidence: "confirmed"}, nil
	}
	return paletteResolution{}, nil
}

func (r *paletteResolver) resolveDipStars(blockIndex int, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	if blockIndex < 0 || blockIndex > 11 {
		return paletteResolution{}, nil
	}
	carrier, err := r.loadArchiveCarrier("DIPLOMAT.LBX", 0, 0, 256)
	if err != nil {
		return paletteResolution{}, err
	}
	moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
	return paletteResolution{Resolved: true, Source: "DIPLOMAT.LBX#0", Evidence: moo2WorkshopEvidence, Confidence: "confirmed"}, nil
}

func (r *paletteResolver) resolveAppPics(blockIndex int, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	if blockIndex != 0 {
		return paletteResolution{}, nil
	}
	carrier, err := r.loadArchiveCarrier("INFO.LBX", 1, 0, 256)
	if err != nil {
		return paletteResolution{}, err
	}
	moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
	return paletteResolution{Resolved: true, Source: "INFO.LBX#1", Evidence: moo2WorkshopEvidence, Confidence: "confirmed"}, nil
}

func (r *paletteResolver) resolveMainPups(blockIndex int, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	switch {
	case blockIndex >= 0 && blockIndex <= 51:
		return r.applyExternalPalette(graphic, 1)
	case blockIndex >= 53 && blockIndex <= 55:
		return r.applyExternalPalette(graphic, 1)
	case blockIndex == 56:
		return r.applyExternalPalette(graphic, 2)
	case blockIndex >= 57 && blockIndex <= 60:
		return r.applyExternalPalette(graphic, 1)
	case blockIndex == 61:
		return r.applyExternalPalette(graphic, 2)
	case blockIndex >= 62 && blockIndex <= 72:
		return r.applyExternalPalette(graphic, 1)
	case blockIndex == 73:
		return r.applyExternalPalette(graphic, 2)
	case blockIndex >= 74 && blockIndex <= 83:
		return r.applyExternalPalette(graphic, 1)
	default:
		return paletteResolution{}, nil
	}
}

func (r *paletteResolver) resolveTwoRangeFonts(blockIndex int, graphic *moo2gfx.Graphic, firstEnd, firstPalette, lastEnd, lastPalette int) (paletteResolution, error) {
	switch {
	case blockIndex >= 0 && blockIndex <= firstEnd:
		return r.applyExternalPalette(graphic, firstPalette)
	case blockIndex > firstEnd && blockIndex <= lastEnd:
		return r.applyExternalPalette(graphic, lastPalette)
	default:
		return paletteResolution{}, nil
	}
}

func (r *paletteResolver) resolveGame(blockIndex int, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	switch {
	case blockIndex >= 0 && blockIndex <= 19:
		return r.applyExternalPalette(graphic, 1)
	case blockIndex >= 20 && blockIndex <= 26:
		return r.applyExternalPalette(graphic, 6)
	case blockIndex >= 27 && blockIndex <= 31:
		return r.applyExternalPalette(graphic, 1)
	default:
		return paletteResolution{}, nil
	}
}

func (r *paletteResolver) resolveColony2(blockIndex int, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	if (blockIndex >= 0 && blockIndex <= 49) || blockIndex == 52 {
		return r.applyExternalPalette(graphic, 2)
	}
	return paletteResolution{}, nil
}

func (r *paletteResolver) resolveRaceSel(blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	if (blockIndex >= 0 && blockIndex <= 14) || (blockIndex >= 29 && blockIndex <= 33) {
		return r.applyExternalPalette(graphic, 10)
	}
	if blockIndex < 34 || blockIndex > 137 {
		return paletteResolution{}, nil
	}
	carrier, err := r.loadLocalCarrier("RACESEL.LBX", archive, 32, 128, 128)
	if err != nil {
		return paletteResolution{}, err
	}
	base, err := r.loadExternalPalette("FONTS.LBX", 10)
	if err != nil {
		return paletteResolution{}, err
	}
	base.Apply(carrier)
	moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
	return paletteResolution{Resolved: true, Source: "FONTS.LBX#10 + RACESEL.LBX#32", Evidence: moo2WorkshopEvidence, Confidence: "confirmed"}, nil
}

func (r *paletteResolver) resolveTechSel(blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	if blockIndex >= 14 && blockIndex <= 22 {
		return r.applyExternalPalette(graphic, 1)
	}
	if blockIndex < 23 || blockIndex > 27 {
		return paletteResolution{}, nil
	}
	carrier, err := r.loadLocalCarrier("TECHSEL.LBX", archive, 14, 224, 16)
	if err != nil {
		return paletteResolution{}, err
	}
	base, err := r.loadExternalPalette("FONTS.LBX", 1)
	if err != nil {
		return paletteResolution{}, err
	}
	base.Apply(carrier)
	moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
	return paletteResolution{Resolved: true, Source: "FONTS.LBX#1 + TECHSEL.LBX#14", Evidence: moo2WorkshopEvidence, Confidence: "confirmed"}, nil
}

func (r *paletteResolver) resolveNewGame(blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	carrier, err := r.loadLocalCarrier("NEWGAME.LBX", archive, 1, 128, 128)
	if err != nil {
		return paletteResolution{}, err
	}
	if blockIndex == 0 || (blockIndex >= 4 && blockIndex <= 22) {
		moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
		return paletteResolution{Resolved: true, Source: "NEWGAME.LBX#1", Evidence: moo2WorkshopEvidence, Confidence: "confirmed"}, nil
	}
	if blockIndex >= 23 && blockIndex <= 29 {
		moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
		base, err := r.loadExternalPalette("FONTS.LBX", 10)
		if err != nil {
			return paletteResolution{}, err
		}
		base.Apply(graphic)
		return paletteResolution{Resolved: true, Source: "FONTS.LBX#10 + NEWGAME.LBX#1", Evidence: moo2WorkshopEvidence, Confidence: "confirmed"}, nil
	}
	return paletteResolution{}, nil
}

func (r *paletteResolver) loadArchiveCarrier(archiveName string, block, expectedShift, expectedEntries int) (*moo2gfx.Graphic, error) {
	archive, err := lbx.Open(filepath.Join(r.sourceRoot, archiveName))
	if err != nil {
		return nil, fmt.Errorf("open %s palette source: %w", archiveName, err)
	}
	return r.loadLocalCarrier(archiveName, archive, block, expectedShift, expectedEntries)
}
func (r *paletteResolver) resolveGStar(blockIndex int, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	switch {
	case blockIndex >= 0 && blockIndex <= 22:
		return r.applyExternalPalette(graphic, 1)
	case blockIndex >= 23 && blockIndex <= 32:
		return r.applyExternalPalette(graphic, 2)
	default:
		return paletteResolution{}, nil
	}
}

func (r *paletteResolver) resolveColony(blockIndex int, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	if blockIndex < 5 || blockIndex > 18 {
		return paletteResolution{}, nil
	}
	return r.applyExternalPalette(graphic, 2)
}

func (r *paletteResolver) resolveMultiGame(blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	carrierBlock := multiGamePaletteCarrier(blockIndex)
	if carrierBlock < 0 {
		return paletteResolution{}, nil
	}
	expectedEntries := 256
	if carrierBlock == 42 {
		expectedEntries = 192
	}
	carrier, err := r.loadLocalCarrier("MULTIGM.LBX", archive, carrierBlock, 0, expectedEntries)
	if err != nil {
		return paletteResolution{}, err
	}
	moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
	return paletteResolution{
		Resolved:   true,
		Source:     fmt.Sprintf("MULTIGM.LBX#%d", carrierBlock),
		Evidence:   moo2WorkshopEvidence,
		Confidence: "confirmed",
	}, nil
}

func multiGamePaletteCarrier(block int) int {
	switch {
	case block >= 1 && block <= 39:
		return 0
	case block == 40:
		return 42
	case block == 41:
		return 0
	case block >= 43 && block <= 45:
		return 42
	case block >= 46 && block <= 149:
		return 0
	case block >= 150 && block <= 253:
		return 42
	case block >= 254 && block <= 260:
		return 0
	default:
		return -1
	}
}
func (r *paletteResolver) resolveBuffer0(blockIndex int, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	if !buffer0UsesFonts1(blockIndex) {
		return paletteResolution{}, nil
	}
	return r.applyExternalPalette(graphic, 1)
}

func buffer0UsesFonts1(block int) bool {
	return (block >= 1 && block <= 12) ||
		(block >= 15 && block <= 91) ||
		(block >= 112 && block <= 121) ||
		(block >= 132 && block <= 136) ||
		(block >= 142 && block <= 287)
}

func (r *paletteResolver) resolveOfficer(blockIndex int, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	paletteBlock := officerPaletteBlock(blockIndex)
	if paletteBlock < 0 {
		return paletteResolution{}, nil
	}
	return r.applyExternalPalette(graphic, paletteBlock)
}

func officerPaletteBlock(block int) int {
	switch {
	case block >= 0 && block <= 209:
		return 1
	case block >= 210 && block <= 276:
		return 2
	case block >= 277 && block <= 343:
		return 4
	default:
		return -1
	}
}

func (r *paletteResolver) resolveFleet(blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	paletteBlock, useCarrier := fleetPaletteRule(blockIndex)
	if paletteBlock < 0 && !useCarrier {
		return paletteResolution{}, nil
	}
	sources := make([]string, 0, 2)
	if useCarrier {
		carrier, err := r.loadLocalCarrier("FLEET.LBX", archive, 111, 0, 176)
		if err != nil {
			return paletteResolution{}, err
		}
		moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
		sources = append(sources, "FLEET.LBX#111")
	}
	if paletteBlock >= 0 {
		palette, err := r.loadExternalPalette("FONTS.LBX", paletteBlock)
		if err != nil {
			return paletteResolution{}, err
		}
		palette.Apply(graphic)
		sources = append([]string{fmt.Sprintf("FONTS.LBX#%d", paletteBlock)}, sources...)
	}
	return paletteResolution{Resolved: true, Source: strings.Join(sources, " + "), Evidence: moo2WorkshopEvidence, Confidence: "confirmed"}, nil
}

func fleetPaletteRule(block int) (paletteBlock int, useCarrier bool) {
	switch {
	case block >= 0 && block <= 44:
		return 1, false
	case block >= 45 && block <= 81:
		return 1, true
	case block >= 83 && block <= 110:
		return 1, true
	default:
		return -1, false
	}
}

func (r *paletteResolver) resolveDiplomat(blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	if blockIndex < 13 || blockIndex > 38 {
		return paletteResolution{}, nil
	}
	carrierBlock := (blockIndex - 13) / 2
	carrier, err := r.loadLocalCarrier("DIPLOMAT.LBX", archive, carrierBlock, 0, 256)
	if err != nil {
		return paletteResolution{}, err
	}
	moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
	return paletteResolution{
		Resolved:   true,
		Source:     fmt.Sprintf("DIPLOMAT.LBX#%d", carrierBlock),
		Evidence:   moo2WorkshopEvidence,
		Confidence: "confirmed",
	}, nil
}

func (r *paletteResolver) resolveMonster(blockIndex int, archive *lbx.File, graphic *moo2gfx.Graphic) (paletteResolution, error) {
	paletteBlock, carrierBlock, ok := monsterPaletteRule(blockIndex)
	if !ok {
		return paletteResolution{}, nil
	}
	sources := make([]string, 0, 2)
	if carrierBlock >= 0 {
		carrier, err := r.loadLocalCarrier("MONSTER.LBX", archive, carrierBlock, 32, 32)
		if err != nil {
			return paletteResolution{}, err
		}
		moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
		sources = append(sources, fmt.Sprintf("MONSTER.LBX#%d", carrierBlock))
	}
	if paletteBlock >= 0 {
		palette, err := r.loadExternalPalette("FONTS.LBX", paletteBlock)
		if err != nil {
			return paletteResolution{}, err
		}
		palette.Apply(graphic)
		sources = append([]string{fmt.Sprintf("FONTS.LBX#%d", paletteBlock)}, sources...)
	}
	return paletteResolution{Resolved: true, Source: strings.Join(sources, " + "), Evidence: moo2WorkshopEvidence, Confidence: "confirmed"}, nil
}

func monsterPaletteRule(block int) (paletteBlock, carrierBlock int, ok bool) {
	switch {
	case block == 7:
		return 1, 14, true
	case block == 8:
		return 1, -1, true
	case block == 9:
		return 1, 14, true
	case block == 12:
		return 1, -1, true
	case block >= 20 && block <= 21:
		return 1, -1, true
	case block == 24:
		return 1, -1, true
	case block == 25:
		return 1, 13, true
	default:
		return -1, -1, false
	}
}

func (r *paletteResolver) applyExternalPalette(graphic *moo2gfx.Graphic, paletteBlock int) (paletteResolution, error) {
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

func (r *paletteResolver) loadLocalCarrier(archiveName string, archive *lbx.File, block, expectedShift, expectedEntries int) (*moo2gfx.Graphic, error) {
	key := fmt.Sprintf("%s#%d", archiveName, block)
	if graphic := r.localCarriers[key]; graphic != nil {
		return graphic, nil
	}
	if block < 0 || block >= len(archive.Entries) {
		return nil, fmt.Errorf("%s does not contain palette carrier block %d", archiveName, block)
	}
	data, err := archive.ReadEntry(block)
	if err != nil {
		return nil, fmt.Errorf("read %s palette carrier block %d: %w", archiveName, block, err)
	}
	graphic, err := moo2gfx.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse %s palette carrier block %d: %w", archiveName, block, err)
	}
	if graphic.Flags&moo2gfx.FlagInternalPalette == 0 || paletteEntries(graphic) != expectedEntries || graphic.PaletteShift != expectedShift {
		return nil, fmt.Errorf("%s block %d is not the expected palette carrier shift=%d entries=%d", archiveName, block, expectedShift, expectedEntries)
	}
	r.localCarriers[key] = graphic
	return graphic, nil
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
	if paletteBlock < 0 && !useCarrier {
		return paletteResolution{}, nil
	}
	if useCarrier {
		carrier, err := r.loadBeamsCarrier(archive)
		if err != nil {
			return paletteResolution{}, err
		}
		moo2gfx.ApplyPaletteFromGraphic(carrier, graphic)
	}
	source := ""
	if paletteBlock >= 0 {
		palette, err := r.loadExternalPalette("FONTS.LBX", paletteBlock)
		if err != nil {
			return paletteResolution{}, err
		}
		palette.Apply(graphic)
		source = fmt.Sprintf("FONTS.LBX#%d", paletteBlock)
	}
	if useCarrier {
		if source != "" {
			source += " + "
		}
		source += "BEAMS.LBX#67"
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
