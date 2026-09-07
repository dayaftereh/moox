# Slice 15.3 Gate 1 - Core SVG icon language implementation

Date: **2026-09-07**

Status: **first implementation landed; Gate-1 visual grammar candidate, not yet Gate-2 freeze**.

## Purpose

MOOX previously used a mix of Unicode symbols and text abbreviations as visible artwork in the strategic shell, including navigation glyphs and resource labels such as `F`, `CP` and `RP`.

The accepted Slice-15.3 direction already requires:

- vector-first navigation/action icons;
- vector-first resource/status icons;
- stylized Farmer/Worker/Scientist figures where readable;
- native trusted React/SVG rather than arbitrary SVG-string injection;
- `viewBox` scaling, `currentColor`, themeability and stable strokes;
- no final icon-font dependency.

This block turns that research direction into the first reusable runtime implementation.

## Central registry

New component:

`web/src/components/GameIcon.tsx`

The shell no longer owns ad-hoc character glyphs. Callers pass a typed `GameIconName`, and `GameIcon` renders trusted inline SVG geometry.

Current registry contains 26 named icons:

### Navigation / feature

- `galaxy`
- `colonies`
- `fleets`
- `research`
- `diplomacy`
- `espionage`
- `ship-designer`

### Shell actions

- `menu`
- `more`
- `home`
- `close`
- `info`

### Economy / resources

- `credits`
- `food`
- `freighter`
- `command`
- `population`
- `production`
- `science`

### World / military primitives

- `planet`
- `star`
- `ship`
- `transport`

### Population roles

- `farmer`
- `worker`
- `scientist`

Not every registered icon is exposed in a screen yet. The unused names are deliberate foundation for the next integration passes, avoiding new one-off SVG implementations in each screen.

## First icon grammar

This is a Gate-1 candidate grammar and can still be tuned after visual review.

Current rules:

- canonical **24 x 24** `viewBox`;
- default **1.55** outline weight;
- rounded line caps and joins;
- `fill="none"` by default;
- sparse `currentColor` solid accents only where a center/pupil/node benefits from fill;
- no hard-coded icon fill palette inside the component;
- parent UI owns semantic color through CSS/currentColor;
- direct React/SVG elements only;
- decorative use is `aria-hidden`; optional labelled icons can expose `role="img"` + title;
- runtime `data-icon` attribute exists for QA/inspection;
- CSS applies `vector-effect: non-scaling-stroke` to icon geometry so the grammar can scale between topbar/mobile/detail sizes without visibly changing line weight.

The style aims for a compact technical/sci-fi language rather than generic emoji or platform-dependent Unicode artwork.

## Live replacement scope

### Strategic navigation

The primary navigation now maps typed sections to SVG icons:

- Galaxy -> `galaxy`
- Colonies -> `colonies`
- Fleets -> `fleets`
- Diplomacy -> `diplomacy`
- Espionage -> `espionage`

The same registry/component is used by both side navigation and bottom/mobile navigation.

### Global resource bar

The resource model is now type-safe (`GameIconName`) rather than arbitrary strings.

Live mappings:

- BC -> `credits`
- Food -> `food`
- Freighters -> `freighter`
- Command Points -> `command`
- Research Points -> `research`

The same icon is reused in the compact topbar chip and its expanded detail popover.

### Shell/menu actions

Replaced text/Unicode artwork with SVG for:

- main menu trigger -> `menu`
- advanced/more -> `more`
- main menu/home -> `home`
- menu close -> `close`
- resource detail close -> `close`

The existing procedural ship silhouette remains the Ship Designer menu visual; it is already a vector/procedural asset and does not need replacement by a generic static icon.

### Colony population jobs

Population assignment now has dedicated small role figures:

- Farmer -> `farmer`
- Worker -> `worker`
- Scientist -> `scientist`

These are used in the job headings and empty-state affordance. Role colors remain UI/CSS concerns rather than being baked into the SVG:

- Farmer receives the existing green/food direction;
- Worker receives a warm production tone;
- Scientist receives a cool research tone.

The Colony planet-information affordance also uses the common `info` icon and its close action uses the common `close` icon.

## Responsive sizing

Observed live CSS sizes:

### Desktop review

- navigation glyph: **17 x 17 px**;
- topbar resource icon: **13 x 13 px**;
- expanded resource-detail icon: **19 x 19 px**;
- resource/menu close: **15 x 15 px**;
- population job figures: **17 x 17 px**.

### 320 px mobile review

- bottom navigation glyph: **16 x 16 px**;
- topbar resource icon: **12 x 12 px**;
- population job figures: **17 x 17 px**.

The SVG source remains one 24x24 registry; callers do not need mobile-specific artwork.

## Browser QA

Validated against the current Slice-15.3 review game on port 7172 without restarting the server.

### Galaxy / desktop

- no danger/error banner;
- all five resource chips expose typed SVG names `credits`, `food`, `freighter`, `command`, `research`;
- all five primary navigation items expose `galaxy`, `colonies`, `fleets`, `diplomacy`, `espionage`;
- expanded BC resource popover reuses `credits` at 19x19 and common `close` at 15x15;
- no horizontal page overflow.

### Mobile 320 x 646

- bottom nav renders all five icons at 16x16;
- resource icons render at 12x12;
- menu popover exposes common `close`, `more`, `home` icons at expected sizes;
- `documentElement.scrollWidth == window.innerWidth` in the tested layouts, i.e. no actual horizontal document overflow.

### Colony Detail

- Farmer, Worker and Scientist job headings expose their distinct SVG role icons;
- planet-info affordance exposes the common `info` icon;
- no danger/error banner;
- desktop and 320-px mobile render the same registry without alternate assets.

## Architectural boundary

This block deliberately does **not** attempt to draw all final MOOX artwork at once.

What is now fixed enough to build on:

- one typed registry/component rather than per-screen Unicode;
- one vector technical baseline;
- one responsive scaling model;
- `currentColor` ownership by screen/theme;
- consistent accessibility behavior.

What remains Gate-1 review material:

- exact final stroke weight/radius vocabulary;
- degree of solid fill vs outline;
- final faction-specific variants;
- icon treatment inside larger illustrated cards;
- Research category icon family;
- building/construction icon family;
- richer star/planet/system presentation;
- final OX brand/app icon.

## Next integration pass

Recommended next block:

1. use the registry in Galaxy/star-system object markers and Fleet/entity affordances;
2. add dedicated Research-category icons rather than one generic atom for every category;
3. add construction/building category primitives;
4. replace remaining visible placeholder boxes/glyphs in Colony and strategic detail screens;
5. only then review the whole family for Gate-2 grammar freeze.

## Research/action refinement follow-up (2026-09-07)

The shared `research` identity now renders a microscope, while `test-tube` is reserved for Chemistry/Biology/Pharma-style category use. Farmer/Worker/Scientist icons now appear on each draggable population marker, not only the role heading. Important actions gained small leading SVGs under a selective button-icon policy. See `docs/research/SLICE_15_3_ICON_REFINEMENT_RESEARCH_ACTIONS_2026-09-07.md`.
