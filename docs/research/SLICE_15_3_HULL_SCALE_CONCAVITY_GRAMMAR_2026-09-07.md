# Slice 15.3 Gate 1 - hull scale and concavity grammar

Date: **2026-09-07**

Status: **implemented Gate-1 visual grammar / ready for user review**.

## Accepted product direction

The procedural Shipbuilder must communicate ship scale through the generated silhouette itself, not merely through a text label or a uniform sprite scale. Small exploration craft should look genuinely small and light; large warships should occupy more visual mass and become structurally more complex, culminating in a very large Doom Star.

The silhouette grammar must also move beyond a mostly convex outer hull. Larger designs should support deliberate **concave side bays, waists/notches and negative internal space** so generated ships can feel more abstract, engineered and varied.

## Authoritative hull evidence versus visual Scout role

Existing MOOX research already normalizes the six original hull records:

| size index | authoritative hull | base cost PP | base space | command points |
| ---: | --- | ---: | ---: | ---: |
| 0 | Frigate | 20 | 25 | 1 |
| 1 | Destroyer | 70 | 60 | 2 |
| 2 | Cruiser | 250 | 120 | 3 |
| 3 | Battleship | 600 | 250 | 4 |
| 4 | Titan | 1500 | 500 | 5 |
| 5 | Doom Star | 4000 | 1200 | 6 |

Source research:

- `docs/research/MILITARY_SHIP_CORE_DESIGN_BASELINE_2026-08-31.md`
- `docs/research/COMMAND_POINTS_SHIP_MAINTENANCE_2026-08-31.md`

**Scout is not introduced as a seventh authoritative gameplay hull.** In the current fresh game the starting authoritative design is named `Scout` and uses the Frigate hull. Slice 15.3 therefore exposes a smaller **Scout visual role** whose gameplay handoff remains Frigate rules until a future authoritative design contract says otherwise.

This keeps the player's desired visual range (small Scout -> Doom Star) without silently changing the server rules.

## Perceptual scale compression

The authoritative base-space progression is extremely wide (25 -> 1200). Rendering that ratio literally would make small ships unreadable or the Doom Star unusably large. The SVG generator therefore uses a monotonic **perceptual compression**: every successive class grows in length, beam, mass and complexity, while staying reviewable in one responsive UI.

Current Gate-1 profile and measured browser body bounds:

| visual class | generated body width | generated body height (representative) | external concave notches | internal cutouts |
| --- | ---: | ---: | ---: | ---: |
| Scout role | 57 | ~18 | 0 | 0 |
| Frigate | 70 | ~31 | 1 | 0 |
| Destroyer | 82 | ~42 | 1 | 0 |
| Cruiser | 92 | ~51 | 2 | 1 |
| Battleship | 101 | ~64 | 2 | 1 |
| Titan | 108 | ~77 | 3 | 2 |
| Doom Star | 114 | ~82 | 3 | 2 |

The exact width/height of a generated silhouette can vary slightly with its deterministic seed, but class envelopes stay ordered.

## New silhouette grammar

`web/src/components/ProceduralShipGlyph.tsx` now uses a hull-class visual profile rather than the previous narrow scalar multiplier.

Each class controls:

- overall hull length;
- beam / visual mass;
- number of hull stations;
- number of forced concave external notches;
- number of transparent internal cutouts;
- engine-count range;
- surface-detail density.

### Concave external bays

The hull remains a deterministic polygon, but selected intermediate stations now expand into a three-point shoulder/notch/shoulder sequence. This produces an actual inward bite in the outer silhouette instead of merely varying the width of a convex envelope.

Large classes deliberately contain more such stations:

- Scout: 0;
- Frigate / Destroyer: 1;
- Cruiser / Battleship: 2;
- Titan / Doom Star: 3.

The notch station and exact depth are still seeded, so repeated `Generate` presses retain class identity but return recognizably different ships.

### Negative internal space

Cruiser and above can contain deterministic transparent internal bays created with SVG masks:

- Cruiser / Battleship: 1;
- Titan / Doom Star: 2.

These openings are true transparent geometry, not background-colored paint, so they remain correct over different backgrounds/themes.

### Mass/readability treatment

The body-fill opacity is also class-aware: Scout remains lighter, while Battleship/Titan/Doom Star receive progressively heavier visual fill. This reinforces perceived mass without changing gameplay values.

## Determinism

The existing visual seed remains the identity input. Hull class is included in the deterministic PRNG input, so:

- the same hull + seed reproduces the same silhouette;
- changing class produces a different class-specific grammar;
- pressing Generate changes the seed/generation and therefore the design while preserving class constraints.

Example Doom Star validation:

- `shipbuilder:game-1:doom_star:1` -> 3 external notches, 2 internal cutouts;
- `shipbuilder:game-1:doom_star:2` -> a visibly different path, still 3 external notches and 2 cutouts.

## Shipbuilder UX changes

The size selector now spans seven visual choices:

1. Scout - labelled **Scout role · Frigate rules**;
2. Frigate - labelled **current ruleset**;
3. Destroyer - visual preview;
4. Cruiser - visual preview;
5. Battleship - visual preview;
6. Titan - visual preview;
7. Doom Star - visual preview.

The Shipbuilder now opens on the Scout visual role so the smallest end of the scale is immediately visible. The authoritative equipment panel still displays the real server-projected Scout/Frigate design separately.

## Browser QA evidence

Fresh `game-1` remains alive on the in-memory development server; the server was **not restarted** for this visual change.

Desktop **995x605**:

- all seven visual classes visible;
- no horizontal page overflow;
- class body bounds increase monotonically Scout -> Doom Star;
- Doom Star fits the SVG coordinate envelope after profile tuning;
- Doom Star candidate generation 1 -> 2 changes the actual polygon path while retaining 3 notches / 2 cutouts.

Mobile **360x646**:

- seven class choices render in a two-column grid;
- no horizontal page overflow;
- default Scout candidate remains usable;
- the existing Generate / recent / Keep / equipment workflow remains responsive.

## Next grammar work

Gate 1 should next explore **art-direction families** on top of this geometry baseline rather than adding gameplay rules:

- faction/race shape families (industrial, angular, organic, crystalline, stealth, etc.);
- optional symmetry/asymmetry rules;
- authored primitive libraries for noses, wings, engine pods and superstructures;
- player locks such as keep nose / keep wings / keep engines / regenerate details;
- empire palette and markings;
- persistence/versioning of the accepted visual seed and generator grammar.

The six-hull gameplay breadth, capacities, component legality and editable weapons remain Slice 17 concerns.
