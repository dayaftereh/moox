# Slice 16.3 Gate 3 - Galaxy visual selector implementation

Date: 2026-09-16
Status: **complete**

Gate-1 evidence: `docs/research/SLICE_16_3_GATE1_GALAXY_VISUAL_SELECTOR_AUDIT_2026-09-15.md`
Gate-2 freeze: `docs/research/SLICE_16_3_GATE2_GALAXY_VISUAL_SELECTOR_FREEZE_2026-09-15.md`

## Objective

Implement the frozen Slice-16.3 Galaxy Size / Galaxy Age contract end to end without reopening the Gate-2 semantics:

- four authoritative Size IDs;
- three original-derived Galaxy Age IDs;
- deterministic data-driven generation for all 12 Size/Age tuples;
- authoritative server-owned Galaxy catalog/facts;
- production Galaxy Age artwork;
- two shared typed visual selectors;
- deterministic/browser regression coverage.

Gate 4 remains separate.

## Implementation commits

- `f2f16e4 feat: generalize galaxy size and age generation`
  - ruleset schema v2 with self-contained per-age spectral/climate profiles;
  - four Size IDs and three Age IDs supported in New Game validation;
  - data-driven `moox_grid_jitter_v1` for all frozen grids;
  - selected Age controls spectral and climate weighted generation;
  - deterministic tests for all 12 Size/Age tuples;
  - Small+Normal backward deterministic baseline preserved.
- `57b5ccd feat: add authoritative galaxy catalog`
  - server-owned Galaxy catalog derived from validated rules;
  - `GET /api/v1/new-game/galaxy`;
  - exact Size star-count facts and qualitative Age bias fields;
  - exact endpoint contract tests.
- `a4e17f6 ui: add galaxy age composition assets`
  - deterministic production Galaxy Age SVG generator;
  - three manifest/provenance entries;
  - build-time deterministic regeneration checks.
- `f2d2285 ui: integrate galaxy size and age selectors`
  - typed web Galaxy catalog API;
  - authoritative four-option Galaxy Size selector;
  - three-option Galaxy Age selector;
  - server-derived info facts;
  - selected Size/Age submitted to Create Game;
  - mobile/desktop browser smoke for Size + Age + Difficulty.

## Ruleset normalization

`data/rulesets/moo2-1.31/new_game_galaxy.json` is now New Game galaxy schema version 2.

Each Galaxy Age owns its complete generator data:

```text
NewGameGalaxyAge
  id
  index
  spectral_weights[7]
  climate_weights[4][10]
```

The previous global 7x3 spectral matrix plus one global Normal climate matrix has been replaced by explicit per-age profiles.

Frozen Age data is implemented as audited:

- `mineral_rich`: original spectral column 0 + object-2 `0x57A7` climate table;
- `normal`: original spectral column 1 + object-2 `0x57A7` climate table;
- `organic_rich`: original spectral column 2 + object-2 `0x57CF` climate table.

Validation now requires the exact frozen Size/Age ID order, valid star/grid capacity, seven non-negative spectral weights totaling 100, and four non-empty 10-weight climate groups per Age.

## Authoritative Size generation

Supported IDs are:

| ID | Systems | Grid | Logical extent |
| --- | ---: | ---: | ---: |
| `small` | 20 | 5 x 4 | 1000 x 800 |
| `medium` | 36 | 6 x 6 | 1200 x 1200 |
| `large` | 54 | 9 x 6 | 1800 x 1200 |
| `huge` | 71 | 9 x 8 | 1800 x 1600 |

`tiny` is not a gameplay/server ID.

The frozen generic `moox_grid_jitter_v1` implementation uses:

- 200 logical units per cell;
- cell center offset 100;
- `Intn(81)-40` jitter for X and Y;
- exactly two coordinate RNG draws per generated star;
- row-major star-index placement;
- no rejection loop or unused-cell random choice.

For Huge, the 72nd grid cell is simply unused because the authoritative count is 71.

## Backward deterministic compatibility

Gate 3 preserved the existing pre-Slice-16.3 Small+Normal deterministic baseline.

The existing golden-seed state fingerprint remained unchanged after the generator became data-driven. This proves the generic implementation did not silently change the established Small+Normal RNG consumption/order or resulting authoritative state.

Permanent tests also prove, for every one of the 12 Size/Age tuples:

- same seed + same complete settings -> byte-identical authoritative state;
- exact system count matches 20/36/54/71;
- every coordinate remains within the selected grid cell's frozen +/-40 jitter bounds;
- invalid `tiny` / unknown-age values are rejected.

The frozen spectral and climate rows are asserted in permanent tests, including the Organic-Rich `0x57CF` climate table.

## Galaxy catalog

Gate 3 implements:

```text
GET /api/v1/new-game/galaxy
```

The live response exposes only player-facing authoritative data:

```text
schema_version
default_size_id = small
default_age_id = normal
sizes[] { id, star_count }
ages[]  { id, mineral_resource_bias, food_world_bias }
```

Live Size rows:

```text
small  -> 20
medium -> 36
large  -> 54
huge   -> 71
```

Live Age rows:

```text
mineral_rich -> mineral higher,  food lower
normal       -> mineral baseline, food baseline
organic_rich -> mineral lower,   food higher
```

The star counts are derived from the loaded, validated ruleset rather than duplicated in the HTTP layer.

Raw spectral/climate matrices, grid dimensions and RNG internals remain generator/ruleset data and are not exposed as player-facing balance facts.

## Invalid-setting contract

Gate-2 policy removed Tiny from the bound gameplay selector rather than presenting it as a disabled pseudo-option.

The stale pre-freeze Gate-3 checklist wording "visibly disabled" is therefore normalized during Gate-3 closeout to the frozen behavior:

- invalid/unknown Size/Age IDs remain server-rejected;
- Tiny is absent from the authoritative catalog and bound selector;
- current player composition remains exactly the existing two-player Human/Darlok contract and invalid player settings continue to be server-rejected by the existing validation/tests;
- Create Game is locked if authoritative Difficulty or Galaxy catalogs are unavailable or do not contain the selected IDs.

No size-dependent player-capacity rule is invented by Slice 16.3.

## Production Galaxy Age artwork

Gate-1 research prototypes were promoted to:

`web/scripts/generate-galaxy-age-art.mjs`

Runtime assets:

- `/assets/new-game/galaxy-age/mineral-rich.svg`
- `/assets/new-game/galaxy-age/normal.svg`
- `/assets/new-game/galaxy-age/organic-rich.svg`

Server/asset mapping preserves the established snake_case -> kebab-case boundary:

- `mineral_rich` -> `mineral-rich`;
- `normal` -> `normal`;
- `organic_rich` -> `organic-rich`.

All three are deterministic 1200 x 675 SVGs with no embedded localized text.

The visual contract remains intentionally honest:

- identical galaxy silhouette and decorative background-star layout across all three Ages;
- only the abstract mineral-to-biosphere composition emphasis changes;
- no young/old temporal-age implication;
- no generated-map preview;
- no exact planet-ratio claim;
- no frontend-owned spectral-colour probability claim.

The static New Game asset guard now checks 13 deterministic SVG files total. This count still includes historical/prototype `tiny.svg`; the production bound selector itself uses 12 authoritative options across Difficulty (5), Galaxy Size (4) and Galaxy Age (3).

## Web/API integration

Web API types now include:

```text
GalaxySizeID = small | medium | large | huge
GalaxyAgeID  = mineral_rich | normal | organic_rich
GalaxyBias   = lower | baseline | higher
```

The New Game page loads both authoritative catalogs:

- Difficulty;
- Galaxy.

Galaxy Size uses `VisualSelector<GalaxySizeID>` with only the four supported IDs.

Galaxy Age uses `VisualSelector<GalaxyAgeID>` with the three supported IDs.

The Size information surface formats `star_count` from server data. The Age information surface formats the two qualitative bias fields from server data. There is no frontend-owned gameplay fact table.

Create Game sends the currently selected:

```text
galaxy_size
galaxy_age
```

rather than hardcoded `small` / `normal` values.

## Browser/live QA

The reusable Chrome smoke now validates three semantically scoped selectors:

- `data-setting-id="difficulty"`;
- `data-setting-id="galaxy-size"`;
- `data-setting-id="galaxy-age"`.

390 px mobile coverage includes:

- no horizontal overflow;
- >=44 px touch targets;
- real CDP touch input;
- Home/End/Arrow keyboard navigation;
- all four Size labels/assets marked supported;
- all three Age labels/assets marked supported;
- all five Difficulty labels/assets still supported;
- Create Game remains enabled for every authoritative Size/Age/Difficulty choice;
- Huge info dialog shows the server-derived `71 Sternsysteme` fact;
- Mineral Rich info shows server-derived higher-mineral / lower-food semantics;
- info-dialog Escape/focus-return behavior remains intact.

Desktop coverage validates responsive artwork sizing/radius and >=52 px arrow controls for Difficulty, Galaxy Size and Galaxy Age.

## Final Gate-3 regression

The final implementation regression passed before documentation closeout:

- `go test ./... -count=1`;
- `npm run build`;
- UTF-8/mojibake guard;
- deterministic 13-SVG New Game asset check;
- TypeScript + Vite production build;
- `npm run check:new-game-selector:browser -- http://127.0.0.1:7171`;
- `git diff --check`;
- `/healthz` through NetBird `100.120.252.216:7171`;
- `/healthz` through LAN `192.168.5.27:7171`;
- live `/api/v1/new-game/galaxy` exact catalog response.

## Gate-3 checklist result

- [x] Extend authoritative New Game validation/generation for the accepted settings.
- [x] Add visual carousel selectors for size and age.
- [x] Add deterministic seed fixtures per supported tuple.
- [x] Keep invalid Size/Age/player settings server-rejected and unavailable in the bound selector contract.

## Next

Gate 4 remains separate and unopened. It should repeat equal-seed/settings hash/equality acceptance, verify visual selections map exactly to authoritative generated settings, perform final desktop/mobile browser review and run the complete tests/build/diff closeout before Slice 16.3 is closed.
