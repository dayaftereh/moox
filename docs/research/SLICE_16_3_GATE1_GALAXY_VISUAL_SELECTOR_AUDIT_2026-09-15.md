# Slice 16.3 Gate 1 - Galaxy visual selector audit

Date: 2026-09-15
Status: **Gate 1 complete / Gate 2 next**

## Objective

Audit the original and current MOOX galaxy-size / galaxy-age contracts before binding new visual selector options. Gate 1 is evidence and prototype work only: no new generator setting is enabled here.

The main outcome is a correction to the planned visual premise: original MOO2 `Galaxy Age` is **not** an astronomical young/normal/old control. It is a three-state world-composition bias named Mineral Rich / Normal / Organic Rich.

## Sources used

Primary local/owned evidence:

- `reference/original/support/files/Orion2.exe` - owned original executable, read through `internal/moo2exe` LE object mapping;
- `reference/original/text/help/block_0000.ascii.txt` - extracted original HELP records 544 and 547;
- `docs/research/NEW_GAME_GALAXY_GENERATION_BASELINE_2026-09-02.md` - prior direct disassembly and Slice-09 generation audit;
- `data/rulesets/moo2-1.31/new_game_galaxy.json` - normalized original-derived New Game tables;
- `internal/ruleset/new_game_galaxy.go` - ruleset schema/validation;
- `internal/game/new_game.go` - current authoritative New Game validation and generation;
- `docs/research/SLICE_16_1_GATE2_GALAXY_SIZE_ART_2026-09-15.md` - accepted visual-size research/prototype family.

No community table is required for the Gate-1 conclusions below.

## Original Galaxy Size evidence

Prior direct disassembly of `Universe_Generation_` proves four original size indices and their star/grid rows:

| Original index | Normalized repo ID | Stars | Grid columns | Grid rows |
| ---: | --- | ---: | ---: | ---: |
| 0 | `small` | **20** | **5** | **4** |
| 1 | `medium` | **36** | **6** | **6** |
| 2 | `large` | **54** | **9** | **6** |
| 3 | `huge` | **71** | **9** | **8** |

Original HELP record 547 says the Galaxy Size control changes both the number of stars and the overall volume of space in the galaxy being created.

Evidence precision note:

- `Small = index 0 = 20 stars / 5x4` is directly documented by the earlier executable audit.
- The repository's normalized four IDs are `small / medium / large / huge` and are validated in `internal/ruleset/new_game_galaxy.go` against the exact four rows above.
- The extracted HELP prose names the control and its semantics but does not enumerate all four textual option labels in that record. Therefore this Gate-1 document treats `medium/large/huge` as the established normalized repo IDs for original indices 1/2/3, without claiming HELP record 547 itself spelled them out.

### Tiny is not an authoritative fifth size

Slice 16.1 intentionally created a five-image browsing prototype `Tiny / Small / Medium / Large / Huge` while only Small was server-supported.

There is no fifth `tiny` row in the original four-row size table or the current normalized ruleset. Therefore:

- the existing `tiny.svg` remains useful Slice-16.1 visual/prototype evidence;
- it is **not** a binding Slice-16.3 gameplay size candidate unless a later explicit MOOX-extension decision adds a fifth size;
- the fidelity-oriented Gate-2 candidate size set is the four-row `small / medium / large / huge` sequence.

## Original Galaxy Age evidence

Original HELP record 544 states the control cycles through, in order:

1. **Mineral Rich**;
2. **Normal**;
3. **Organic Rich**.

The HELP description states that Mineral Rich increases the chance of planets with mineral resources and decreases the chance of worlds that create food; Organic Rich does the opposite.

The original global `_g_galaxy_age` is three-way and the normalized mapping is:

| Original index | Repo ID |
| ---: | --- |
| 0 | `mineral_rich` |
| 1 | `normal` |
| 2 | `organic_rich` |

This means the previous planned "younger / normal / older" astronomy cue is rejected. It would misstate the original setting.

## Original spectral weights by Galaxy Age

`Generate_Spectral_Class_` uses a 7x3 weight matrix. Rows are original spectral classes 0..6 and columns are Mineral Rich / Normal / Organic Rich:

```text
20 10  5
25 15  5
10 16 30
10 16 21
32 37 30
 1  2  3
 2  4  6
```

Equivalently, per age:

- `mineral_rich`: `20, 25, 10, 10, 32, 1, 2`
- `normal`: `10, 15, 16, 16, 37, 2, 4`
- `organic_rich`: `5, 5, 30, 21, 30, 3, 6`

Each spectral column sums to 100.

Gate 1 does **not** convert these class indexes into a new frontend-owned star-colour rule. The selector art must not invent a spectral-class colour mapping that the generator/server does not expose as a player-facing contract.

## Climate-weight recheck and new Organic-Rich evidence

The prior Slice-09 audit established:

- object 2 offset `0x57A7` is the 4-group x 10-climate table used by Galaxy Ages 0 and 1;
- object 2 offset `0x57CF` is the adjacent 4-group x 10-climate table used by Galaxy Age 2 (`organic_rich`).

For this Gate-1 audit, `Orion2.exe` was read directly through `internal/moo2exe.ReadObject(2, offset, 40)`.

The original bytes are stored climate-major (four group bytes per climate). Transposing `0x57A7` reproduces the already committed `normal_climate_weights` exactly, validating the extraction method.

### Age 0/1 climate weights (`0x57A7`)

Climate column order remains:

```text
toxic, radiated, barren, desert, tundra,
ocean, swamp, arid, terran, gaia
```

Group rows:

```text
group 0: 15 55 25  5  0  0  0  0  0 0
group 1: 15 50 25 10  5  0  0  0  0 0
group 2: 10 15 10 10 10 10 11 11 11 2
group 3: 20  0 70  0  8  2  0  0  0 0
```

These are **weights**, not frontend percentages; for example group 1 sums to 105 and is still valid weighted-choice input.

### Organic-Rich climate weights (`0x57CF`) - newly rechecked in Gate 1

```text
group 0: 15 40 20 25  0  0  0  0  0 0
group 1:  5 30 20 25 20  0  0  0  0 0
group 2:  5  8  8 13 13 13 13 13 10 4
group 3: 20  0 50  0 30  0  0  0  0 0
```

All four Organic-Rich rows sum to 100.

This table is Gate-1 evidence only. It is **not** added to the runtime ruleset in Gate 1; Gate 2 must decide the normalized schema and Gate 3 would implement it.

## Current MOOX authoritative runtime

Current public game-layer types expose only:

```text
GalaxySizeSmall = "small"
GalaxyAgeNormal = "normal"
```

`validateNewGameSettings` rejects any other size or age.

The current generator is still the narrow Slice-09 implementation:

- `const stars = 20`;
- spectral generation always selects column `1` (`normal`);
- coordinate placement is hardcoded as `column = i % 5`, `row = i / 5`;
- the accepted `moox_grid_jitter_v1` Small coordinate baseline is 5x4 in a logical 1000x800 map;
- planet climate generation always uses `NormalClimateWeights[group]`;
- New Game still requires exactly two players;
- the supported race pair remains one Human plus one Darlok.

Therefore there is currently **no size-dependent player-capacity contract** and no player-cap fact should be shown in the 16.3 selector yet.

## Data already normalized but not runtime-enabled

The ruleset already carries more evidence than the generator currently consumes:

- all four size rows and exact 20/36/54/71 star counts;
- all four original grid shapes;
- all three Galaxy Age IDs;
- all three spectral-weight columns.

But runtime schema currently has only one climate matrix field, `NormalClimateWeights`, and runtime code selects it unconditionally.

## Gate-1 runtime-readiness classification

| Setting | Evidence/data state | Runtime-ready now without generator changes? | Gate-1 conclusion |
| --- | --- | --- | --- |
| `small` | fully normalized and implemented | **yes** | current authoritative size |
| `medium` | count/grid normalized | no | needs data-driven star/grid/coordinate implementation + deterministic fixtures |
| `large` | count/grid normalized | no | same |
| `huge` | count/grid normalized | no | same |
| `tiny` | no original/runtime row | no | visual prototype only; exclude from fidelity binding candidate |
| `normal` age | fully normalized and implemented | **yes** | current authoritative age |
| `mineral_rich` | ID + spectral column proven; shares age-0/1 climate table | no | needs runtime age selection and explicit normalized climate-contract shape |
| `organic_rich` | ID + spectral column proven; exact separate climate table now rechecked | no | needs ruleset normalization + runtime age selection + fixtures |

Thus **Small + Normal is the only tuple that is runtime-ready with zero new generator work today**.

Medium/Large/Huge look low-risk from an evidence perspective because their star/grid rows are already normalized, but they still require a Gate-2 coordinate/generalization freeze and Gate-3 implementation.

Mineral/Organic Age support is also evidence-ready enough for Gate 2 now, but is not currently runtime-ready.

## Size-art audit

The accepted Slice-16.1 procedural Galaxy Size assets are:

- `/assets/new-game/galaxy-size/tiny.svg`
- `/assets/new-game/galaxy-size/small.svg`
- `/assets/new-game/galaxy-size/medium.svg`
- `/assets/new-game/galaxy-size/large.svg`
- `/assets/new-game/galaxy-size/huge.svg`

For Slice 16.3, the coherent fidelity-oriented size prototype is the existing **Small / Medium / Large / Huge** subset.

The existing art deliberately communicates relative scale through footprint, density, disk extent, core luminosity and structural richness rather than arm count. That remains a good visual language because original HELP says Galaxy Size changes star count and overall volume.

Honesty boundary:

- the SVG's decorative/star-circle count is **not** the authoritative generated star count;
- morphology and arm count are illustrative;
- the art is not a generated-map preview;
- exact facts such as `20 / 36 / 54 / 71 stars` belong in server/ruleset-derived fact text once those sizes are actually supported;
- no image may imply exact generated star positions.

All four Small/Medium/Large/Huge SVGs were revalidated as XML during this Gate-1 audit.

## Galaxy Age visual prototype

Research-only deterministic prototypes were created under:

`docs/research/prototypes/SLICE_16_3_GALAXY_AGE_2026-09-15/`

Files:

- `mineral_rich.svg`
- `normal.svg`
- `organic_rich.svg`
- `generate.mjs`

The prototype intentionally keeps the same:

- 1200x675 framing;
- galaxy silhouette;
- decorative background-star positions/count;
- overall map footprint.

It changes only an abstract **mineral <-> biosphere composition cue**:

- Mineral Rich emphasizes a faceted mineral/crystal glyph and places the composition marker toward that side;
- Normal balances both cues;
- Organic Rich emphasizes a biosphere/planet glyph and places the marker toward that side.

This avoids all of the misleading implications rejected by Gate 1:

- no young/old temporal astronomy claim;
- no different star count by Age;
- no exact planet ratio;
- no generated star positions;
- no claim that green/red/blue galaxy colour directly maps to an original spectral class.

The three SVGs contain no embedded localized `<text>` elements and parse as valid XML.

## Recommended Gate-2 questions

Gate 1 intentionally stops short of freezing these decisions, but the evidence points to a narrow set of Gate-2 choices:

1. Freeze the fidelity-oriented Size IDs as `small / medium / large / huge`; decide explicitly whether `tiny` is retired from the bound selector or retained only as a clearly non-authoritative MOOX extension.
2. Freeze Galaxy Age as `mineral_rich / normal / organic_rich`; do **not** use young/old labels.
3. Freeze a data-driven coordinate generalization for 5x4 / 6x6 / 9x6 / 9x8 while preserving deterministic `moox_grid_jitter_v1` semantics or versioning a successor algorithm.
4. Extend the normalized ruleset schema to own the age-2 Organic-Rich climate table and make the age-to-climate-table selection explicit.
5. Decide server-owned New Game metadata/catalog shape for size star counts and Age summaries, following the Slice-16.2 Difficulty-catalog precedent.
6. Keep player-count capacity out of Size facts until a separate authoritative player-composition contract exists.
7. Decide whether the accepted existing four size images and the new composition-based Age prototype are production-ready directions or need another art pass before Gate 3.

## Gate-1 checklist result

- [x] Re-check original galaxy size and age tables.
- [x] Map current generator support and star/system counts.
- [x] Identify settings that are runtime-ready without new generator mechanics.
- [x] Prototype one coherent image series for size and one for age.
- [x] Verify no illustration implies data the generator does not guarantee.

Gate 1 is complete. Gate 2 is next; no new galaxy setting was enabled by this audit.
