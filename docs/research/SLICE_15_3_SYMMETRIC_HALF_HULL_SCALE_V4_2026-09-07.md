# Slice 15.3 Gate 1 - symmetric half-hull generator and hull footprint v4

Date: **2026-09-07**

Status: **implemented first draft, browser-validated**.

## Product decision

The standard MOOX procedural ship generator now uses **strict bilateral symmetry around the longitudinal X axis**.

The generator stores and randomizes only one canonical lateral half of the ship. The renderer mirrors that half exactly to the opposite side. This applies to the core hull contour as well as lateral primitives and cutouts.

The primary Shipbuilder remains intentionally simple:

`choose hull size -> see one generated ship -> click ship / Generate -> full random reroll -> Use this design`

There is still only one player-facing generation point. Directed evolution and edit locks remain historical/advanced generator research and are not reintroduced into the normal flow.

## Coordinate / symmetry contract

The ship flies along the X axis.

Visual Genome v4 stores positive **distance from the X axis** for the generated half geometry. In SVG screen coordinates the canonical upper half is rendered as `centerY - offset`; the opposite side is generated only by reflection:

`y_mirror = 2 * centerY - y`

The important contract is not SVG's screen-coordinate sign convention but that only one half is generated and the second half is a deterministic mirror.

## Visual Genome v4

The symmetry change is a generator-semantics change, so the visual genome version advances from **v3 -> v4**.

### Primitive genes

A primitive gene now stores only:

- kind (`wedge`, `spike`, `pod`);
- longitudinal position;
- length;
- lateral width;
- sweep.

The old per-primitive `side` and probabilistic `mirrored` fields are removed from new genomes.

Every stored primitive is rendered twice:

1. canonical half;
2. exact X-axis mirror.

This removes accidental whole-ship asymmetry while preserving the large topology/randomness space.

### Cutout genes

Cutouts are also stored once. A normal lateral cutout becomes an upper/lower mirrored pair. A centerline cutout such as the Fork morphology's central opening remains a single centered opening.

Mirrored ellipse rotations use opposite angles, preserving reflection rather than simple duplication.

### Core hull

The core body already uses station half-widths. v4 makes that contract explicit:

- generate the canonical half contour;
- include concave shoulder/notch/shoulder features on that half;
- reverse and mirror the resulting point list;
- close one symmetric SVG hull path.

### Engines and hardpoints

Engine positions are centered around the axis and therefore stay symmetric for both even and odd engine counts.

Weapon/hardpoint markers are emitted as mirrored pairs; an odd final marker sits on the centerline rather than breaking the silhouette.

## Asymmetry policy

`asymmetric` is removed from the standard full-random morphology pool.

The active primary pool is now:

1. Needle / Cigarette;
2. Barge / Heavy;
3. Manta / Widewing;
4. Fork / Twin boom;
5. Chevron / A-shape;
6. Hammer / Front-heavy;
7. Bulb / Blob.

This retains radical differences in topology, aspect ratio and mass distribution while making the results cleaner and more intentional.

Asymmetry can still return later as an explicit advanced/alien/faction-specific rule, but it is not part of the default random generator.

## Hull Space values

The Shipbuilder now exposes the existing authoritative hull-space progression directly in the size selector and current generated-ship metadata:

| visual selection | gameplay Space | note |
| --- | ---: | --- |
| Scout | 25 | visual Scout role currently uses Frigate rules |
| Frigate | 25 | authoritative baseline hull |
| Destroyer | 60 | authoritative hull |
| Cruiser | 120 | authoritative hull |
| Battleship | 250 | authoritative hull |
| Titan | 500 | authoritative hull |
| Doom Star | 1200 | authoritative hull |

Source baseline remains `docs/research/MILITARY_SHIP_CORE_DESIGN_BASELINE_2026-08-31.md`.

## Visible preview footprint

A second issue was that dynamic SVG viewBoxes made every hull class appear roughly the same physical size in the Shipbuilder even though their internal geometry dimensions differed.

v4 therefore separates:

- **shape topology / SVG viewBox**, which keeps extreme forms unclipped;
- **class footprint**, which controls how much of the common preview stage the ship visually occupies.

Current first-draft footprint factors:

| class | footprint |
| --- | ---: |
| Scout | 0.40 |
| Frigate | 0.48 |
| Destroyer | 0.58 |
| Cruiser | 0.68 |
| Battleship | 0.78 |
| Titan | 0.90 |
| Doom Star | 1.00 |

These factors are applied to the large random ship and the size-selector reference thumbnails. Fleet/map scaling is deliberately not changed in this Gate-1 block.

## Browser validation

The existing in-memory `game-1` remained alive; the preview server was not restarted.

### Desktop 995 x 605

No error banner and no horizontal page overflow.

Measured in the same **512 x 390 px** random-ship stage:

| class | footprint | rendered SVG box | Space |
| --- | ---: | --- | ---: |
| Scout | 0.40 | ~184 x 136 px | 25 |
| Cruiser | 0.68 | ~313 x 231 px | 120 |
| Doom Star | 1.00 | ~461 x 340 px | 1200 |

The selector thumbnails are also monotonic from Scout 0.40 through Doom Star 1.00.

### Symmetry evidence

The live SVGs expose `data-symmetry="x-axis"` plus stored-half/rendered counts for QA.

Representative random rerolls:

- Fork/Spear: 9 stored half-primitives -> 18 rendered primitive paths; coordinate-pair reflection check passed;
- Needle/Spear: 11 stored half-primitives -> 22 rendered primitive paths;
- Doom Star/Hammer: 2 stored half-cutouts -> 4 rendered cutout ellipses; center positions/radii mirror exactly.

The core hull point list also passes direct coordinate reflection checks around the SVG centerline.

### Mobile 360 x 646

Same **311 x 220 px** preview stage:

- Scout footprint 0.40 -> ~120 x 78 px;
- Doom Star footprint 1.00 -> ~299 x 195 px;
- Space metadata remains visible;
- no horizontal overflow;
- no persistent server-error banner.

## Next work

The next useful review is visual rather than architectural:

- tune the footprint factors if Scout/Frigate feel too small or Titan/Doom Star too dominant;
- sample each of the seven symmetric morphology families across Scout, Cruiser and Doom Star;
- refine signature half-primitives so Manta/Chevron/Fork remain clearly different after mandatory mirroring;
- then continue with accepted Visual Genome v4 persistence/versioning and faction/race weighting.