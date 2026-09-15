# Slice 16.3 Gate 2 - Galaxy visual selector freeze

Date: 2026-09-15
Status: **complete / frozen**

Gate-1 evidence: `docs/research/SLICE_16_3_GATE1_GALAXY_VISUAL_SELECTOR_AUDIT_2026-09-15.md`
Shared selector foundation: closed Slice 16.1.
Difficulty catalog/server-owned-fact precedent: closed Slice 16.2.

## Architectural ownership

Slice 16.3 owns the authoritative New Game contract for:

- supported Galaxy Size identifiers;
- supported Galaxy Age identifiers;
- deterministic size/age generation inputs;
- normalized original-derived size/age ruleset data;
- the player-facing Galaxy New Game catalog;
- exact star-count facts;
- evidence-backed qualitative Galaxy Age facts;
- deterministic selector assets and responsive presentation.

Slice 16.3 does **not** own broad player/opponent composition. Player count remains the current exactly-two-player Human/Darlok baseline until the later player-composition slice.

## Frozen Galaxy Size contract

The authoritative V1 gameplay size set is exactly:

1. `small`
2. `medium`
3. `large`
4. `huge`

Default: `small`.

Frozen generator facts:

| ID | Original index | Stars / star systems | Grid | Derived logical extent |
| --- | ---: | ---: | ---: | ---: |
| `small` | 0 | **20** | 5 x 4 | 1000 x 800 |
| `medium` | 1 | **36** | 6 x 6 | 1200 x 1200 |
| `large` | 2 | **54** | 9 x 6 | 1800 x 1200 |
| `huge` | 3 | **71** | 9 x 8 | 1800 x 1600 |

The star counts and grid dimensions are original-derived. The logical extents are the explicit MOOX coordinate normalization frozen below, not original MOO2 pixel-coordinate claims.

### Tiny policy

`tiny` is **not** a Galaxy Size server ID in Slice 16.3.

The existing Slice-16.1 Tiny image was an intentional presentation/browsing prototype before the original/runtime size table was fully audited. Gate 1 confirmed there is no fifth Tiny row in the authoritative four-row original-derived table.

Frozen policy:

- do not add `GalaxySizeTiny`;
- do not return Tiny in the authoritative Galaxy catalog;
- do not show Tiny in the bound production Galaxy Size selector;
- the existing `tiny.svg` may remain as a historical/prototype asset until a later cleanup, but its presence in the asset tree must not imply gameplay support;
- adding Tiny later would require an explicit MOOX-extension design decision and a new generator profile rather than silently inventing original parity.

## Frozen Galaxy Age contract

The authoritative V1 Galaxy Age set is exactly:

1. `mineral_rich`
2. `normal`
3. `organic_rich`

Default: `normal`.

These names preserve the original setting semantics. They are **not** temporal astronomy labels.

Frozen player-facing semantics from original HELP:

| ID | Mineral-resource bias | Food-world bias |
| --- | --- | --- |
| `mineral_rich` | `higher` | `lower` |
| `normal` | `baseline` | `baseline` |
| `organic_rich` | `lower` | `higher` |

No stronger numeric probability claim is exposed in the UI.

## Frozen spectral contract

Each Galaxy Age profile owns exactly seven spectral weights, in original spectral-class index order 0..6:

```text
mineral_rich: 20, 25, 10, 10, 32, 1, 2
normal:       10, 15, 16, 16, 37, 2, 4
organic_rich:  5,  5, 30, 21, 30, 3, 6
```

Every row sums to 100.

These values remain generator data. Gate 3 must not expose a frontend-owned spectral-class-to-colour probability table.

## Frozen climate contract

Climate order remains:

```text
toxic, radiated, barren, desert, tundra,
ocean, swamp, arid, terran, gaia
```

Galaxy Ages 0 and 1 use the original object-2 `0x57A7` climate table:

```text
group 0: 15 55 25  5  0  0  0  0  0 0
group 1: 15 50 25 10  5  0  0  0  0 0
group 2: 10 15 10 10 10 10 11 11 11 2
group 3: 20  0 70  0  8  2  0  0  0 0
```

`organic_rich` uses the directly rechecked object-2 `0x57CF` table:

```text
group 0: 15 40 20 25  0  0  0  0  0 0
group 1:  5 30 20 25 20  0  0  0  0 0
group 2:  5  8  8 13 13 13 13 13 10 4
group 3: 20  0 50  0 30  0  0  0  0 0
```

These are weighted-choice inputs, not UI percentages. Rows are valid when their total weight is positive; they are not required to sum to 100. The existing Age0/1 group-1 row sums to 105 and is preserved exactly.

## Frozen normalized ruleset shape

Gate 3 should bump the normalized New Game galaxy ruleset schema and make every age self-contained rather than relying on implicit global matrix columns.

Frozen conceptual shape:

```text
NewGameGalaxySize
  id
  index
  stars
  grid_columns
  grid_rows

NewGameGalaxyAge
  id
  index
  spectral_weights[7]
  climate_weights[4][10]
```

The existing global `spectral_weights[7][3]` and single `normal_climate_weights[4][10]` representation may be migrated into these per-age profiles in Gate 3.

Validation must require:

- exactly four unique size IDs in frozen order;
- exactly three unique age IDs in frozen order;
- size index matches list position;
- age index matches list position;
- `stars > 0` and `stars <= grid_columns * grid_rows`;
- every size grid dimension is positive;
- every age has exactly 7 spectral weights;
- every age spectral row sums to 100;
- every age has exactly 4 climate groups x 10 climate weights;
- every climate weight is non-negative;
- every climate group has a positive total weight.

## Frozen deterministic coordinate normalization

The existing `moox_grid_jitter_v1` Small algorithm is generalized data-driven without changing Small behavior.

For selected size profile `S`:

```text
cell_size = 200
center_offset = 100
jitter_roll = Intn(81) - 40
logical_width  = S.grid_columns * 200
logical_height = S.grid_rows * 200
```

For star index `i` from `0 .. S.stars-1`:

```text
column = i % S.grid_columns
row    = i / S.grid_columns
center_x = 100 + 200 * column
center_y = 100 + 200 * row
x = center_x + Intn(81) - 40
y = center_y + Intn(81) - 40
```

Frozen RNG/order rules:

1. choose spectral class for all `N` stars, in star-index order;
2. allocate star IDs and consume exactly two coordinate jitter draws per star, in row-major star-index order;
3. generate planets/bodies after all star spectral/coordinate planning, preserving the existing generation ordering;
4. no coordinate rejection loop is added in this normalization;
5. no extra RNG draw is consumed to choose unused cells.

For Huge, `9 x 8 = 72` cells exist but only 71 stars are generated. The final row-major cell is therefore deterministically unused.

### Backward-compatibility requirement

For `small + normal`, Gate 3 must prove the generalized/data-driven implementation produces the same deterministic authoritative state as the pre-Slice-16.3 implementation for the existing fixtures/seeds.

No Small coordinate, RNG-consumption or star-order change is permitted merely as a side effect of making the algorithm generic.

## Frozen generation selection

Gate 3 generation must resolve both selected profiles before RNG-driven universe generation:

```text
sizeProfile = GalaxySizeProfileFor(settings.galaxy_size)
ageProfile  = GalaxyAgeProfileFor(settings.galaxy_age)
```

Then:

- star count/grid comes only from `sizeProfile`;
- spectral weighted choice comes only from `ageProfile.spectral_weights`;
- planet climate weighted choice comes only from `ageProfile.climate_weights[group]`;
- all other existing planet/body/homeworld/start rules remain unchanged unless separately required for generic size support;
- unsupported/unknown IDs are rejected before generation begins.

All 4 x 3 = **12 Size/Age tuples** are intended to be authoritative Gate-3 supported combinations with the current exactly-two-player baseline.

## Frozen server Galaxy catalog

Follow the Slice-16.2 Difficulty pattern: the frontend must not duplicate authoritative star counts or gameplay-effect semantics in option constants.

Frozen endpoint:

```text
GET /api/v1/new-game/galaxy
```

Frozen response shape:

```text
GalaxyCatalogResponse
  schema_version
  default_size_id
  default_age_id
  sizes[]
  ages[]

GalaxySizeProfile
  id
  star_count

GalaxyAgeProfile
  id
  mineral_resource_bias
  food_world_bias
```

Frozen bias wire values:

```text
lower
baseline
higher
```

Frozen catalog rows:

```text
sizes:
  small   -> 20
  medium  -> 36
  large   -> 54
  huge    -> 71

ages:
  mineral_rich -> mineral higher,  food lower
  normal       -> mineral baseline, food baseline
  organic_rich -> mineral lower,   food higher
```

The server/catalog may derive these profiles from validated ruleset/game data, but the response values are server-owned.

Not exposed in the player-facing catalog:

- raw 7-element spectral weights;
- raw 4x10 climate weights;
- coordinate grid dimensions;
- RNG algorithm internals;
- invented size/player-capacity limits.

## Frozen player-facing facts

### Size

Compact fact:

- exact star-system count from `star_count`.

Example localized rendering:

```text
20 star systems
36 star systems
54 star systems
71 star systems
```

The frontend may localize wording but must not own the numeric values.

### Age

Compact facts are derived from the two server bias fields.

Allowed semantic rendering:

- `higher`: more likely / increased tendency;
- `baseline`: normal/baseline tendency;
- `lower`: less likely / decreased tendency.

For example, Mineral Rich may render as “More mineral-rich worlds” and “Fewer food-friendly worlds”. Organic Rich renders the opposite. Normal renders the baseline distribution.

Do not render raw percentages because the original contract is weighted generation with interacting spectral/climate/body rules, not one simple player-facing percentage.

## Frozen TypeScript identifier shapes

Gate 3 should widen the current narrow API types to:

```text
GalaxySizeID = 'small' | 'medium' | 'large' | 'huge'
GalaxyAgeID  = 'mineral_rich' | 'normal' | 'organic_rich'
GalaxyBias   = 'lower' | 'baseline' | 'higher'
```

`NewGameSettings.galaxy_size` and `.galaxy_age` use those types directly.

## Frozen Size assets

Production selector variants are exactly:

```text
new-game:galaxy-size:small
new-game:galaxy-size:medium
new-game:galaxy-size:large
new-game:galaxy-size:huge
```

Runtime paths remain:

```text
/assets/new-game/galaxy-size/small.svg
/assets/new-game/galaxy-size/medium.svg
/assets/new-game/galaxy-size/large.svg
/assets/new-game/galaxy-size/huge.svg
```

The accepted Slice-16.1 deterministic 1200 x 675 procedural images remain the frozen production direction.

Visual semantics remain illustrative:

- increasing footprint/density/richness communicates larger scale;
- arm count/morphology is not a numeric size promise;
- decorative SVG star-circle count is not the generator star count;
- exact generated star positions are never depicted.

## Frozen Galaxy Age assets

Server IDs use snake_case, but asset option IDs follow the existing semantic-path kebab-case convention.

Frozen mapping:

| Server ID | Asset option ID | Runtime path |
| --- | --- | --- |
| `mineral_rich` | `mineral-rich` | `/assets/new-game/galaxy-age/mineral-rich.svg` |
| `normal` | `normal` | `/assets/new-game/galaxy-age/normal.svg` |
| `organic_rich` | `organic-rich` | `/assets/new-game/galaxy-age/organic-rich.svg` |

Semantic manifest IDs:

```text
new-game:galaxy-age:mineral-rich
new-game:galaxy-age:normal
new-game:galaxy-age:organic-rich
```

Gate 3 should promote the research prototype into a deterministic production generator, conventionally:

`web/scripts/generate-galaxy-age-art.mjs`

All three assets remain 1200 x 675 SVGs with no embedded localized `<text>`.

Frozen visual language:

- identical galaxy silhouette across all three Age variants;
- identical decorative background-star positions/count across all three;
- identical overall footprint;
- only the abstract mineral-to-biosphere composition emphasis changes;
- Mineral Rich emphasizes the mineral/crystal cue;
- Normal balances the two cues;
- Organic Rich emphasizes the biosphere/food-world cue.

This prevents the image from implying temporal age, different galaxy size, exact star positions, exact planet ratios or frontend-owned spectral-colour mechanics.

## Frozen responsive presentation

Reuse the closed Slice-16.1 `VisualSelector<TId>` grammar for **two separate selectors**:

1. Galaxy Size;
2. Galaxy Age.

Presentation rules:

- one selected 16:9 visual card per selector;
- previous/next arrows and position dots retain existing shared behavior;
- labels remain localized UI text outside the SVG;
- the `?` information dialog/bottom sheet shows server-derived compact facts;
- desktop uses the existing New Game settings-grid/card rhythm;
- mobile keeps one-column selector cards, >=44 px controls and no horizontal overflow;
- keyboard ArrowLeft/ArrowRight/Home/End behavior remains inherited from `VisualSelector`;
- info-dialog Escape/focus-return behavior remains inherited from the shared selector;
- Create Game stays locked if the authoritative Galaxy catalog is unavailable or the selected IDs are absent from it.

After Gate-3 implementation, all four Size and all three Age options are supported; there is no preview-only Tiny option in the bound selector.

## Frozen validation and deterministic test obligations for Gate 3

Gate 3 must add permanent coverage for:

1. exact catalog IDs/defaults/facts;
2. rejection of unknown Size/Age IDs;
3. all 12 Size/Age tuples generating successfully with current valid two-player settings;
4. same seed + same full settings tuple producing byte-identical authoritative state;
5. exact star count for each size;
6. coordinates remaining inside the derived size extent and within the frozen per-cell jitter bounds;
7. Small+Normal backward deterministic equality with the pre-16.3 baseline;
8. age selection consuming the correct spectral/climate profile;
9. `mineral_rich`/`normal` sharing the frozen `0x57A7` climate data while using different spectral rows;
10. `organic_rich` using the frozen `0x57CF` climate data;
11. frontend/server catalog agreement and create-lock behavior;
12. deterministic production generation of the four selected Size assets plus three Age assets;
13. mobile/desktop mouse/touch/keyboard selector QA.

## Explicit non-goals

Gate 2 does not freeze or implement:

- more than two players;
- size-dependent maximum player counts;
- extra galaxy shapes/morphologies as generator settings;
- exact original MOO2 pixel-coordinate parity;
- original coordinate rejection-loop parity;
- Tiny as an original fifth size;
- raw spectral/climate probabilities as player-facing data;
- astronomical young/old Galaxy Age semantics.

## Gate-2 checklist result

- [x] Freeze supported galaxy sizes and ages.
- [x] Freeze exact runtime IDs and displayed facts.
- [x] Freeze asset variants and responsive presentation.

Gate 2 is complete. Gate 3 implementation is next and must start from this frozen contract rather than reopening product semantics.
