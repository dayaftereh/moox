# Slice 15.3 - Gate 2 visual identity / asset-pipeline contract draft

Date: **2026-09-07**

Status: **Gate-1 proposal ready for user review; not frozen until Gate 2 is explicitly approved**.

## 0. Visual goals

1. **Strategy-first readability** - state, actions and hierarchy must remain understandable before decorative art is considered; 320 CSS px is a real acceptance floor.
2. **Recognizable original MOOX identity** - OX branding, Deep Space Command surfaces, vector language and procedural ships should make the product identifiable without copying original-game artwork.
3. **Modern science-fiction tone** - precise orbital/vector geometry, controlled glow and dark-space depth; avoid generic fantasy ornament and avoid turning the strategic layer into a dense engineering dashboard.
4. **Nostalgia through function, not pixels** - preserve recognizable strategic concepts, category structure and pacing cues where useful, but recreate all distributable visual expression as original/licensed MOOX work.
5. **Semantic consistency over decoration** - the same meaning should reuse the same icon/color identity across Galaxy, System, Colony, Fleet, Research and Diplomacy.
6. **Rich art is additive** - future portraits, planet paintings and backgrounds enrich the UI but may not obscure, replace or become authority for gameplay state.

Acceptable reference distance: MOOX may study private/original material for functional roles and historical context, but final distributable art must be independently authored and must not trace/copy original pixels, logos, compositions or distinctive asset geometry without explicit rights.

## 1. Proposed visual direction

Strategic baseline: **Deep Space Command (Styleboard Candidate A)**.

Core characteristics:

- deep-space navy/black surfaces;
- electric-blue player/action/selection accent;
- restrained glow;
- vector-first semantic UI language;
- moderate radii and clear panel hierarchy;
- rich illustration may be added later without replacing the functional vector layer.

Candidate B (Orbital Glass) remains a reference for atmospheric/showcase treatments. Candidate C (Industrial Tactical) remains a reference for Tactical Slice 15.5.

## 2. Proposed brand contract

- Product wordmark: **Master of Orion X / MOOX**.
- Compact brand motif: accepted **OX** lettermark plus a simple original spacecraft/star/orbit application mark when a pictorial mark is required.
- Brand marks must remain original MOOX work; no original-game logo pixels or traced artwork.
- The OX mark may use the primary accent but must also work in monochrome/currentColor contexts.
- Do not encode gameplay meaning in the brand mark.

## 3. Palette / semantic role contract

Exact token values remain implementation tokens; the semantic roles are the contract.

| Role | Strategic baseline | Meaning |
| --- | --- | --- |
| background | very dark navy/black | space / canvas |
| raised surface | dark blue-gray | cards, dialogs, controls |
| primary accent | electric blue | player focus, navigation, primary action |
| vector highlight | pale blue/cyan | ships, stars, selected geometry |
| success/friendly | green | owned colony, peace/friendly, positive status |
| warning | amber | outpost, caution, pending warning |
| danger/hostile | red/pink-red | war, hostile, destructive action |
| neutral/unknown | slate | unknown/neutral contact, unavailable secondary state |
| Research/category accent | green/cyan-compatible | category recognition without replacing semantic danger/warning colors |

Rule: semantic state color takes priority over decorative palette variation.

## 4. Typography contract

- Browser/system sans stack remains the default strategic UI family unless a distributable project font is explicitly approved later.
- Headings: strong weight, compact line height, limited negative tracking only at large title sizes.
- Eyebrows/status labels: uppercase, high weight, increased letter spacing.
- Body/status values must remain legible at the 320px floor without relying on hover.
- No text may be baked into raster game art when it can remain localized DOM text.

## 5. Core SVG icon grammar

- canonical viewBox: **24 x 24**;
- `currentColor` by default;
- fill primarily `none`;
- rounded caps/joins;
- nominal stroke: approximately **1.55**;
- filled dots/centers are allowed when they improve recognition;
- icons describe a semantic role, not a screen-local decoration;
- reuse an existing semantic icon before adding a near-duplicate.

Current size tiers:

| Context | Target rendered size |
| --- | ---: |
| compact button | 13px |
| standard button | 15px |
| Research field bar | 15px |
| population role marker | 18px |
| Galaxy system center | 18px |
| Construction catalog | 19px |
| Research category card mobile | 22px |
| Research category card desktop | 25px |
| large project/hero icon | 48px mobile / ~64-66px larger layouts |

Touch controls remain at least 44px where they are primary interactive controls; tiny map/status markers are informational/secondary and must not be the sole touch target.

## 6. Responsive layout contract

Canonical browser floor: **320 CSS px**.

Proposed tiers:

| Tier | Width | Rule |
| --- | --- | --- |
| compact mobile | 320-479px | single-column detail surfaces; bottom nav; no horizontal document scroll |
| wide mobile / compact tablet | 480-759px | opportunistic two-column groups where readable |
| tablet / desktop | 760-1199px | side-by-side strategic panels and richer dialogs |
| wide desktop | 1200px+ | use space for context, not uncontrolled density growth |

Rules:

- document-level horizontal scrolling is a failure;
- dedicated pannable/zoomable canvases (Galaxy, future Tactical) may pan internally without widening the document;
- critical actions must remain visible without hover-only discovery;
- text and icons must survive browser zoom and system font metrics;
- 320px is a hard QA target, not a design afterthought.

## 7. Raster / illustration target matrix

Vector remains preferred for functional marks. These raster targets apply only where illustration/texturing is valuable.

| Asset class | Authoring target | Runtime variants | Crop/safe-area rule |
| --- | --- | --- | --- |
| application icon | SVG master | 32, 64, 128, 256, 512 PNG as required by packaging | centered mark; 12% outer breathing room |
| race portrait | 1024x1280, 4:5 | 256x320 and 512x640 WebP/PNG | face/eyes inside central 65%; do not crop identity features |
| planet portrait/detail | 1024x1024, 1:1 | 256 and 512 square WebP/PNG | planet disk within central 82% |
| Colony environment/hero | 1600x900, 16:9 | 800x450 / 1600x900 WebP; optional 4:5 mobile crop | key structures inside central 60% width / 72% height safe zone |
| Galaxy background | 2560x1440, 16:9 | 1280x720 / 2560x1440 WebP | non-semantic; cover/crop allowed; no important labels baked in |
| Tactical background | 2560x1440, 16:9 | 1280x720 / 2560x1440 WebP | non-authoritative background; Tactical overlays stay vector/DOM |
| building illustration | SVG preferred or 512x512 | 128/256 square WebP/PNG | recognizable silhouette at 64px |
| technology illustration | SVG preferred or 512x512 | 128/256 square WebP/PNG | no text baked in |
| UI texture/ornament | vector or tileable raster | size-specific only when needed | decorative only; must not encode state |

Runtime raster preference: WebP for opaque/illustrative browser assets, PNG for alpha-heavy/fallback/package needs. Keep the source/master format separate from runtime delivery.

## 8. Procedural ship visual contract

Current proposal:

- full resolved **Visual Genome v4** is authoritative optional visual state;
- generator grammar is symmetric half-hull/mirroring with the accepted morphology/style fields;
- appearance revision is separate from gameplay/equipment revision;
- a built ship freezes its accepted visual genome;
- player `Generate` in the eventual full Ship Designer changes appearance only;
- presets get a fresh resolved visual genome once at new-game creation, then persist exactly;
- gameplay rules must not depend on rendered SVG geometry;
- future faction/race visual weighting may choose genome probabilities but must not require a new save schema for every art tweak if full genome is already resolved.

No raster ship catalog is required for the canonical strategic match. Cached raster exports may be added later as an optimization, never as gameplay authority.

## 9. Asset naming and semantic IDs

Game/runtime references use stable semantic IDs, not filenames.

Examples:

- `race.alkari.portrait`
- `planet.terran.background`
- `building.marine_barracks.icon`
- `research.physics.icon`
- `ui.action.colonize`

Proposed distributable layout after Gate-2 approval:

```text
assets/
  manifest.json
  races/
  planets/
  buildings/
  technologies/
  ui/
  backgrounds/
  source/
web/public/assets/        # generated/copied runtime mirror; do not hand-author
reference/original/       # private local reference; ignored by Git
```

The exact physical layout can change without changing semantic IDs.

File naming proposal:

`<semantic-id>.<variant>.vNN.<ext>`

Examples:

- `race.alkari.portrait.default.v01.webp`
- `planet.terran.background.default.v02.webp`
- `ui.brand.ox.mono.v01.svg`

Build hashes remain a delivery concern and are not semantic IDs.

## 10. Provenance / license contract

Every committed distributable non-trivial art asset must have provenance metadata. Allowed provenance classes:

1. `moox_original` - authored specifically for MOOX;
2. `generated` - generated from an original MOOX brief with recorded tool/model/workflow metadata;
3. `commissioned` - created by a contributor/artist with explicit distribution rights assigned/licensed to MOOX;
4. `third_party_licensed` - only when the exact license permits the intended redistribution/modification;
5. `reference_only` - **never distributable**; may exist only in ignored/private reference locations.

Required metadata fields for distributable art:

- semantic asset ID;
- provenance class;
- creator/owner;
- creation/acquisition date;
- source/master path;
- runtime derived paths;
- SHA-256 of the committed master and/or runtime asset;
- license identifier/terms or MOOX ownership statement;
- source URL/reference when third-party licensed;
- generation tool/model/version and brief/prompt reference when generated;
- derivative chain when the runtime file is produced from another master;
- reviewer/approval state.

Rules:

- original Master of Orion II artwork is reference-only unless a separate explicit redistribution right is proven;
- extracted LBX/graphiccatalog/palettecatalog data is evidence/reference, not automatic shipping permission;
- do not commit private reference material under `assets/` or `web/public/`;
- no asset is accepted merely because it is technically downloadable;
- unclear license/provenance means **do not ship**.

## 11. Fallback / missing-asset policy

The browser must remain functional when rich art is missing.

Fallback order:

1. accepted semantic rich asset;
2. typed `GameIcon` / procedural vector representation;
3. neutral geometric placeholder that preserves layout;
4. localized text label.

A broken image icon, missing alt/label or layout collapse is never an acceptable fallback.

Loading rich assets must not block gameplay state or command submission.

## 12. Source/runtime boundary

- server/game state stores semantic IDs and visual genomes, not browser file paths;
- React/browser maps semantic IDs to runtime presentation;
- rich visuals remain non-authoritative presentation unless a future slice explicitly defines otherwise;
- source masters are not necessarily served to the browser;
- runtime derivatives are reproducible from approved masters/workflow metadata;
- generated build output under `web/dist` is never the source of truth.

## 13. Representative Gate-2 acceptance surfaces

The Gate-2 freeze should be judged against these concrete review surfaces:

1. strategic shell/navigation/resource row;
2. Galaxy + selected system markers;
3. star-system dialog with planet/gas/asteroid and colony/fleet presence;
4. Colony list + Colony Detail population/building presentation;
5. Construction catalog/project/queue;
6. location/fleet cards plus procedural ship visual;
7. eight-category Research full page + compact overlay;
8. Diplomacy neutral/friendly/hostile visual states;
9. Espionage empty-state only until Slice 20;
10. Ship Designer appearance subfunction as currently represented by Shipbuilder research, without claiming Slice-17 equipment editing is complete.

The live Styleboard `/styleboard.html` supplies a compact cross-surface direction comparison. The functional routes on 7171 remain the primary acceptance evidence.

## 14. Gate-2 decisions requested

To freeze Gate 2, explicitly accept or revise:

- Candidate A / Deep Space Command as strategic baseline;
- OX brand treatment direction;
- semantic palette roles;
- 24x24 vector icon grammar;
- 320px responsive floor and target tiers;
- raster target matrix;
- semantic asset ID/naming policy;
- provenance/license metadata requirements;
- fallback policy;
- Visual Genome v4 procedural ship contract;
- representative acceptance surfaces.

Until that approval, this document is a **draft contract**, not a frozen specification.
