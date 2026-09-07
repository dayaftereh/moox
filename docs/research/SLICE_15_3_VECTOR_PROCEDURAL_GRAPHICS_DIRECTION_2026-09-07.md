# Slice 15.3 - vector-first and procedural graphics direction

Date: **2026-09-07**

Status: **Gate-1 direction candidate / implementation proof of concept started**.

## Product decision captured

Slice 15.3 should prefer modern, resolution-independent graphics wherever that is a good fit. The primary direction is **vector-first** rather than a large fixed raster-icon library.

The browser should therefore move away from generic Unicode glyphs as final UI artwork and toward an original MOOX icon/illustration system built primarily from SVG/vector geometry.

## Proposed asset technology split

### Vector-first

Use SVG/vector graphics as the default for:

- navigation and action icons;
- resource/status icons;
- Research category symbols;
- Diplomacy/Espionage symbols;
- Farmer/Worker/Scientist figures where a readable stylized figure is sufficient;
- fleet/ship strategic silhouettes;
- star-system markers, orbit overlays, selection/range/route overlays;
- UI ornamentation, badges, frames and scalable state indicators;
- procedural 2D ship geometry.

Preferred runtime integration:

- native inline SVG/React components for dynamic/procedural visuals;
- `viewBox`-based scaling;
- CSS design tokens / `currentColor` for themeable icons;
- SVG `vector-effect: non-scaling-stroke` where line weight should remain visually stable;
- no icon-font dependency for final MOOX artwork;
- decorative vectors hidden from assistive technology, semantic icons labelled where required;
- avoid runtime injection of arbitrary SVG strings; construct trusted SVG elements/components directly.

### Raster/painted imagery remains valid

Vector-first does **not** mean every visual must be SVG. Rich textured artwork is better kept as modern compressed raster imagery when appropriate, especially:

- large planet paintings/textures;
- race/faction portraits;
- nebula/starfield texture layers if procedural/CSS treatment is insufficient;
- highly detailed Colony/building illustrations;
- cinematic or painterly presentation.

For these asset families Gate 2 should evaluate modern runtime formats such as AVIF/WebP plus a defined fallback policy. Source artwork and runtime delivery formats should remain separate concerns.

## Procedural ship direction

A fixed catalog of ship pictures is not the preferred long-term direction. The proposed MOOX approach is a **deterministic procedural 2D ship generator**.

### Why 2D first

- directly usable in the current browser HMI;
- extremely cheap to render and scale;
- natural fit for SVG;
- supports mobile and desktop without separate resolution assets;
- can provide large visual variety without storing thousands of images;
- generated geometry can later become an input to a 3D/extrusion pipeline without requiring runtime 3D now.

### Determinism rule

Visual randomness must never be frame-time randomness. A ship/design receives a stable seed derived from authoritative identity data. Re-rendering, saving/resuming and replaying the same design must reproduce the same visual.

Initial proof-of-concept seed contract:

`empire_id : source_design_id : source_design_revision : strategic_picture_id`

The seed is **presentation-only** and does not affect gameplay authority.

### Geometry grammar

The first SVG proof of concept generates a top-down silhouette from a seeded geometry grammar:

- mirrored primary hull silhouette;
- deterministic hull width/length profile;
- seeded wing/body stations;
- engine count/placement;
- panel-line placement;
- visible hardpoints derived from weapon count;
- hull-size scaling hooks for future Frigate/Destroyer/Cruiser/Battleship/Titan breadth.

The current ruleset only exposes the narrow Frigate military baseline, but the generator is intentionally prepared for later Slice-17 hull breadth without inventing new gameplay rules.

## Recommended future ship grammar

Gate 2 should freeze a more expressive generator contract with independent layers:

1. **Design seed** - stable major silhouette for a ship design.
2. **Faction/race style grammar** - angular, organic, industrial, crystalline, asymmetric, etc.; visual only.
3. **Hull class** - controls scale, mass, complexity and available visual zones.
4. **Functional modules** - engines, weapon hardpoints, hangars/sensors where authoritative components exist.
5. **Cosmetic modules** - fins, armor panels, antennae, lights and markings.
6. **Empire palette** - themeable faction colors with contrast/accessibility constraints.
7. **Optional serial variation** - very small per-ship differences while keeping a design immediately recognizable. This should be opt-in; the design silhouette itself should remain stable.

## 2D -> future 3D path

Do not introduce runtime 3D in Slice 15.3 solely for novelty. Instead preserve geometry semantics so a later generator can:

- extrude the 2D hull polygon;
- generate layered height zones;
- place engine/weapon/module sockets from the same seed;
- render sprites/portraits offline or eventually support a native 3D tactical presentation.

This keeps the current implementation simple while avoiding a dead-end 2D asset model.

## Current implementation proof of concept

Added `web/src/components/ProceduralShipGlyph.tsx`.

Properties:

- dependency-free React/SVG;
- deterministic seeded PRNG;
- `viewBox` scalable geometry;
- hull-class scaling hooks;
- weapon hardpoint count input;
- CSS/current-color compatible;
- used initially in the star-system Fleet/Ship roster to replace generic ship glyphs.

This is a **Gate-1 exploratory implementation**, not the final frozen art style.

## Gate-2 decisions still required

- exact MOOX icon grammar and line/fill style;
- palette and faction-color rules;
- final ship silhouette grammar and race-style families;
- whether individual ships receive subtle serial variation or only per-design variation;
- raster image runtime formats/fallbacks;
- asset catalog schema and provenance metadata;
- SVG source/storage conventions and optimization/build policy;
- representative acceptance mockups for Galaxy, system, Colony, Fleet and Research screens.

## Shipbuilder interaction contract added 2026-09-07

The procedural generator is now exposed through the in-game Shipbuilder prototype using the accepted sequence **choose hull size -> Generate repeatedly -> Keep design -> equipment handoff**. The current Frigate is server-authoritative; Destroyer/Cruiser/Battleship/Titan remain explicitly labelled visual previews until Slice 17 supplies authoritative multi-hull/component/weapon breadth. The accepted visual seed must eventually be persisted with the ship design so save/resume/replay preserve the chosen appearance.

See `docs/research/SLICE_15_3_SHIPBUILDER_GENERATION_UX_2026-09-07.md` for the complete browser/product contract and QA evidence.
