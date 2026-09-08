# Slice 15.3 - Spectral stars, building-art coverage and Construction layout correction

Date: **2026-09-08**

Status: **implemented as the second Gate-2 art-fidelity review revision; Gate 2 remains unfrozen**.

## Review feedback

The previous semantic star marker still read as an icon rather than a star. The requested direction is closer to the original game's strategic reading:

- a genuinely luminous stellar core;
- soft light fading outward;
- visible star-color variation by primary/spectral class;
- color/glow rather than a generic outlined star symbol.

The same review also identified two Construction issues:

1. Housing and later buildings should not remain visually generic simply because they are not in the starting tech set.
2. The larger BuildingArt introduced a layout mismatch: the new image could exceed the old icon-sized grid column, separating image, production-point facts and description.

## Original-game reference check

The historical MOO2 manual states that every visitable star system is displayed on the Galaxy Map in the color of its primary star. Its documented star-color sequence includes:

- Blue-White / Class B;
- White / Class F;
- Yellow / Class G;
- Orange / Class K;
- Red / Class M;
- Brown dwarfs;
- the strategic map also contains special objects such as black holes (already documented in `docs/research/MOO2_GAME_REFERENCE.md`).

Historical screenshots were checked only for visual behavior/reference: small bright stellar cores, colored coronas and soft outward glow on a dark starfield. No original pixels, sprite geometry or palettes are copied into MOOX.

The normalized current MOOX generator uses integer spectral values 0..6. The visual mapping is:

| normalized value | visual identity | display class |
| ---: | --- | --- |
| 0 | blue-white | B |
| 1 | white | F |
| 2 | yellow | G |
| 3 | orange | K |
| 4 | red | M |
| 5 | brown | BD |
| 6 | black hole / special object | BH |

Value 6 is already special in the generator: it produces no normal satellite system in `generateNewGamePlanets`, matching its non-standard object role.

## `StarArt`

New runtime component:

`web/src/components/StarArt.tsx`

Normal stellar rendering uses original MOOX SVG/vector effects:

- multiple radial gradients;
- bright white/hot inner core;
- class-colored stellar body;
- broad class-colored corona fading to alpha zero;
- secondary blur halo;
- four subtle diffraction/light rays with deterministic rotation;
- deterministic low-opacity `feTurbulence` surface texture;
- small hot highlight;
- stable per-system visual variation from system ID.

The component remains resolution independent and contains no raster copy of original artwork.

### Black hole

Spectral value 6 has a separate composition:

- nearly black central event-horizon disk;
- blue/violet outer halo;
- thin luminous accretion ellipse;
- soft disk blur;
- no fake stellar surface.

## Product integration

### Galaxy Map

The previous 18px `star-system` semantic icon is no longer the primary star representation.

Each Galaxy system now renders `StarArt` using:

- authoritative `system.spectral_class`;
- stable `system.id` visual seed.

The interaction button is a transparent 44x44 desktop / 42x42 compact hit target. The actual luminous star art is:

- 48x48 desktop;
- 44x44 at the 320px floor.

Ownership/fleet/contact markers remain separate semantic overlays. A selected system gets a restrained selection ring around the star rather than replacing the stellar color with the generic UI accent.

### System dialog

The central system star now uses the same `StarArt` component at:

- 104x104 desktop;
- 92x92 compact.

The class badge now displays the recognizable class token rather than a raw normalized integer. Example from live `game-1`:

- System 01: `Spektralklasse B`, title `Blue-White`, 104x104 blue-white star.

## Live spectral coverage

Current `game-1` contains normalized spectral values:

- 0: 3 systems;
- 1: 3 systems;
- 3: 6 systems;
- 4: 6 systems;
- 5: 1 system;
- 6: 1 system.

This gives live coverage for six distinct visual identities without modifying game state. Spectral value 2 / Yellow is additionally rendered in the Art Fidelity Lab.

## Building ruleset audit

Authoritative source:

`data/rulesets/moo2-1.31/buildings.json`

Current count: **48 building/transformation definitions**.

This includes early buildings plus later facilities such as Automated Factories, Research Laboratory, Biospheres, planetary defenses, economic buildings, pollution/industry facilities, the Star Base/Battlestation/Star Fortress chain, Terraforming/Gaia transformation and late strategic structures.

### Housing is not a persistent building

`Housing` is intentionally different:

- it is a repeatable Construction project;
- it must have its own rich project illustration;
- it must **not** appear later as a persistent built building on the Colony/planet surface.

The current Construction UI therefore renders Housing with a dedicated habitat/dome illustration while preserving its project semantics.

## Rich building-art coverage

The earlier six explicitly illustrated objects remain dedicated:

- Capitol;
- Colony Base;
- Marine Barracks;
- Star Base;
- Battlestation;
- Star Fortress.

All other authoritative building IDs now fall through a deterministic rich facility generator instead of one identical generic-building scene.

Facility families:

- industry/mining;
- research;
- agriculture;
- defense;
- civic/economic;
- environment/transformation;
- space/special.

Each family has a distinct motif/palette, and each building ID deterministically changes tower count, heights, widths, layout and detail lights. This is a fidelity baseline rather than a claim that every late-game facility has final hand-authored art; it guarantees that late buildings no longer collapse onto one identical placeholder image and gives them a stable route for later art replacement.

The Art Fidelity Lab contains **49 cards** in the coverage section:

- 48 persistent building/transformation IDs;
- 1 explicitly labelled Housing repeatable-project card.

`web/src/building-art-catalog.ts` is generated from the authoritative ruleset for this review coverage.

## Construction layout bug and fix

### Root cause

The first rich BuildingArt pass retained old icon-era layout assumptions:

- `construction-project-hero` still used two columns;
- rich art had a 190px minimum height plus `aspect-ratio`;
- that minimum height forced the visual wider than the first grid column;
- the 68px rich catalog image was still placed in a 34px catalog grid column;
- `margin-top: auto` on project actions exaggerated vertical separation.

This is why the image, PP data and description appeared not to line up.

### Corrected detail layout

Rich projects (building, Housing and planetary transformation) now use:

- full-width 16:7 hero illustration;
- title/type directly below the image;
- production cost / maintenance / remaining / ETA directly below title;
- description immediately after facts;
- actions in normal flow rather than being pushed to the bottom.

Measured desktop result at 791px:

- detail width: 434px;
- rich hero: 434x189.8;
- title: full 434px below hero;
- facts: 434px, directly after title;
- description: 434px;
- actions: 434px;
- no horizontal overflow.

Measured 320px result:

- detail width: 290px;
- rich hero: 290x126.9;
- title: 290px;
- facts: 290px;
- description: 290px;
- actions: 290px;
- document scroll width: exactly 320px;
- no horizontal overflow.

### Corrected catalog layout

Rich catalog rows now reserve the width of their art:

- desktop: 68x48 art in a 68px column;
- compact: 60x42 art in a 60px column.

Housing participates in the rich catalog layout. Ship/freighter projects retain compact semantic icon rows.

## Art Fidelity Lab

`/art-fidelity.html` now adds:

- all seven spectral identities side by side;
- all 48 current ruleset building/transformation art routes;
- Housing explicitly called out as non-persistent repeatable project.

At 320px:

- stars resolve to two columns;
- building coverage resolves to two columns;
- 49 building/project cards render;
- no document-level horizontal overflow.

## Gate status

This remains a **Gate-2 review revision**. Gate 2 is intentionally not frozen until the richer star/planet/building direction is visually accepted.
