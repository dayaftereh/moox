# Slice 16.5 Gate 3 - Starting technology implementation

Date: 2026-09-17

Status: **complete; Gate 4 next**.

## Implemented runtime boundary

- `pre_warp` and `average` are accepted New Game technology levels.
- `advanced` remains present in the authoritative catalog as `planned`, is visible in the selector, and cannot be submitted to Create Game.
- Full Advanced initialization remains explicitly deferred to Slice 16.7.

## Backend and server

- New Game validation accepts only the Gate-2 support matrix: Pre-Warp and Average.
- Pre-Warp creates the common home-colony baseline and stops before creating Average starting ship designs, Scouts, Colony Ship or strategic fleets.
- Average keeps the established starting-asset path unchanged.
- A server-owned `/api/v1/new-game/technologies` catalog exposes the three frozen entries, support/planned status and compact authoritative facts.

## HMI and assets

- New Game loads the technology catalog and renders a typed three-option `VisualSelector`.
- Create Game submits the selected supported `technology_level` instead of a hard-coded Average value.
- The Create Game action is locked whenever technology catalog data is unavailable or the selected option is not server-supported.
- Advanced remains selectable for inspection but is visibly planned/locked.
- Three deterministic 1200x675 original SVG technology-progression assets were added for Pre-Warp, Average and Advanced, together with their generator and manifest entries.

## Determinism and tests

- Added deterministic Human/Klackon x Pre-Warp/Average fixtures using an identical seed twice.
- Pre-Warp fixtures assert no starting ships/designs/fleets and the frozen two completed starting fields.
- Average fixtures preserve the existing starting-fleet shape.
- Advanced has an explicit rejection test.
- HTTP catalog coverage freezes the server support boundary.
- Frontend contract coverage verifies all three catalog-bound options, Advanced locking and byte-for-byte reproducible SVG generation.

Verification completed on 2026-09-17:

- `go test ./... -count=1` - pass.
- `npm run build` in `web/` - pass, including UTF-8, New Game selector, race assets, technology-start contract, TypeScript and Vite build.
- Gate 4 remains responsible for exact frozen-state closure, repeated determinism acceptance, desktop/390 px visual review and final full diff/build closeout.
## Live visual refinement after user review

User review on 2026-09-17 found two HMI problems after Gate 3 went live:

- the three technology illustrations were technically separate SVG files but shared almost the same central-core composition, so they looked effectively identical in the selector;
- the localized option titles repeated the word "technology" and could wrap instead of staying as a compact single-line choice.

Refinement applied:

- rebuilt the deterministic SVG generator around three deliberately distinct compositions while keeping one coherent MOOX visual language:
  - Pre-Warp: planetary surface / launchpad / primitive orbital satellite;
  - Average: orbital station plus two scouts and one colony ship;
  - Advanced: multi-node hyperlane network and expanded fleet presence;
- added stable per-art `data-art-signature` markers and extended `check:technology-start` so the three compositions are contract-tested as distinct and reproducible;
- shortened selector labels to `Pre-Warp`, `Average`, `Advanced` in English and `Pre-Warp`, `Durchschnittlich`, `Fortschrittlich` in German;
- added a technology-selector-specific no-wrap title rule for compact desktop/mobile presentation;
- added a technology-art asset version query so browsers do not keep the previous visually-similar SVGs from cache after the update;
- rendered all three live-served SVGs through headless Chrome at 1200x675 and confirmed different rendered outputs before removing temporary QA screenshots.

Verification after refinement:

- `npm run build` - pass;
- `check:technology-start` - pass with three distinct/reproducible SVG composition signatures and compact one-line labels;
- live server returns HTTP 200 for all three updated SVGs;
- live JS bundle contains the new art cache-busting version and compact German labels.