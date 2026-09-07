# Slice 15.3 Gate 1 - art / asset audit

Date: **2026-09-07**

Status: **in progress**.

## Starting point

Slice 15.2 is closed after independent Gate-4 QA. Slice 15.3 is now open at Gate 1 to establish the MOOX visual identity and a reproducible graphics/asset pipeline without moving gameplay authority into the client.

## Initial repository inventory

- `web/src` currently contains one CSS file, two TypeScript files and six TSX files in the visual/runtime source set inspected for Gate 1.
- No shipped `.svg`, `.png`, `.jpg`, `.jpeg`, `.webp` or `.gif` assets were found under `web/` in the initial audit.
- Existing backend/catalog infrastructure includes `graphiccatalog` and `palettecatalog`, plus the broader catalog/support/text/audio catalog families.
- The accepted Slice-15.2 strategic HMI therefore provides real integration targets, but the current browser presentation is still primarily CSS/DOM/icon-glyph driven rather than an established artwork pipeline.

## Gate-1 direction

The first art-direction pass should cover the accepted integration targets rather than inventing new mechanics:

1. OX/Master of Orion X brand and compact application icon.
2. Galaxy background, stars, ownership/selection states and travel overlays.
3. Orbital star-system view: star, rocky/gas planets, asteroid belts, colony/outpost/fleet presence.
4. Colony/planet imagery, Farmer/Worker/Scientist figures, construction/building language.
5. Fleet/ship markers and future Tactical presentation assets.
6. Eight Research-category icons.
7. Diplomacy and Espionage presentation language.
8. Responsive shell ornamentation, hover/selected/disabled states and motion rules.

## Pipeline questions to freeze in Gate 2

- asset source formats and generated runtime formats;
- naming/versioning and resolution/aspect-ratio variants;
- responsive crop/fallback rules;
- provenance/license metadata;
- React asset-catalog integration boundary;
- whether selected 3D-to-2D render workflows add value for ships/planets without coupling runtime gameplay to 3D assets.

## Immediate next Gate-1 work

- inventory current CSS design tokens and glyph usage;
- inventory `graphiccatalog` / `palettecatalog` capabilities and provenance constraints;
- produce multiple representative visual-direction candidates for Galaxy, system, Colony and core iconography;
- compare the candidates against desktop and mobile readability before Gate-2 freeze.

## Gate-1 audit expansion - vector/glyph/catalog findings

- Current shell navigation still uses Unicode glyph strings in `AppShell.tsx` (`Galaxy`, `Colonies`, `Fleets`, `Diplomacy`, `Espionage`). Treat these as temporary functional placeholders rather than final visual identity.
- The current CSS already has a useful token baseline (`--bg`, surfaces, borders, text/muted, accent, semantic success/warning/danger, radii and shadow). Slice 15.3 should evolve these rather than hard-coding independent asset colors.
- `internal/graphiccatalog` schema 2 is a reference/extraction catalog for original LBX graphics. It records source/archive/block hashes, dimensions, frame counts, palette context and optional exported PNG frames. This is valuable evidence/provenance infrastructure, not a license to ship extracted original-game artwork.
- `internal/palettecatalog` schema 1 extracts source palette data to hashed JSON plus PNG swatches. It can support reference analysis and color-distance studies while keeping final MOOX palette decisions original.
- Final icon direction is now **SVG/vector-first**; generic Unicode glyphs and icon fonts should not be the final MOOX icon system.
- Rich textured/painted visuals remain a separate asset class and may use modern compressed raster delivery where vector graphics are a poor fit.

## Procedural ship proof of concept

- `web/src/components/ProceduralShipGlyph.tsx` now provides a dependency-free deterministic React/SVG generator.
- Major silhouette is seeded from stable design identity rather than render-time randomness.
- The initial geometry grammar covers hull envelope/wing station, engines, panel lines and weapon hardpoints.
- The star-system Fleet/Ship roster is the first integration target; this replaces placeholder ship glyphs without changing any server/gameplay authority.
- This is Gate-1 exploration only. Race/faction style grammars, empire palettes and optional per-ship serial variation remain Gate-2 decisions.