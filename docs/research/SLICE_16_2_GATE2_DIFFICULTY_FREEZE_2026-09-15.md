# Slice 16.2 Gate 2 - Difficulty freeze

Date: 2026-09-15
Status: **complete / frozen**

Gate-1 evidence: `docs/research/SLICE_16_2_GATE1_DIFFICULTY_AUDIT_2026-09-15.md`
Shared UI foundation: closed Slice 16.1.

## Architectural ownership decision

Difficulty is a server-owned New Game/gameplay contract in Slice 16.2. It is **not** deferred wholesale to a future AI slice.

Slice 16.2 owns:

- canonical Difficulty identifiers;
- New Game/API schema support;
- authoritative server/catalog profile data;
- validation;
- save/resume persistence of the selected contract;
- deterministic application of the frozen difficulty effects;
- concise player-facing modifier facts supplied from authoritative data.

A later AI/intelligence slice owns:

- decision-policy quality;
- strategic search/lookahead;
- personality behavior;
- expansion/fleet/diplomacy/tactical decision quality;
- other intelligence improvements that are not simply deterministic Difficulty modifiers.

The AI may consume the selected Difficulty/profile, but it does not define the existence or naming of Difficulty.

## Canonical server identifier proposal

Gate-2 starter freezes the **identifier shape** for the five main MOOX levels unless later evidence forces a change:

- `easy`
- `normal`
- `hard`
- `very_hard`
- `impossible`

Recommended Go shape for Gate 3:

```text
type DifficultyID string

const (
    DifficultyEasy       DifficultyID = "easy"
    DifficultyNormal     DifficultyID = "normal"
    DifficultyHard       DifficultyID = "hard"
    DifficultyVeryHard   DifficultyID = "very_hard"
    DifficultyImpossible DifficultyID = "impossible"
)
```

`DifficultyNormal` is the default/no-special-advantage baseline and is the semantic successor to original MOO2 `Average`.

## Tutor policy

Original MOO2 `Tutor` is **not** silently renamed to MOOX `Easy`.

Gate-2 direction:

- the main MOOX Difficulty selector remains the five requested levels Easy / Normal / Hard / Very Hard / Impossible;
- original Tutor-specific restrictions/assistance are treated as tutorial/onboarding semantics rather than an ordinary sixth difficulty level;
- if Tutor behavior is implemented later, it should receive an explicit tutorial/onboarding contract rather than contaminating `easy` with hidden mode-specific rules.

## Display vocabulary

Main MOOX labels remain:

1. Easy
2. Normal
3. Hard
4. Very Hard
5. Impossible

These are deliberate MOOX UX labels. Original-name parity is documented separately and is not exposed as if the original used this exact list.

## Frozen numeric/profile contract

The server owns a single authoritative Difficulty profile/catalog. V1 uses additive per-role bonuses rather than global percentage multipliers, matching both the documented MOO2/1.50 structure and the existing MOOX economy pipeline.

Only built-in-AI empires receive these Difficulty modifiers. Human-controlled empires do not.

To keep the persisted contract exact, values are stored as signed integer eighth-units (1 = 0.125 resource units). Runtime economy code converts them at the calculation edge with value / 8.0.

Frozen profile fields:

```text
DifficultyProfile
  id
  ai_food_per_farmer_eighths
  ai_production_per_worker_eighths
  ai_research_per_scientist_eighths
  ai_tax_bc_per_population_eighths
  ai_command_deficit_bc_per_point_eighths
```

Frozen V1 values:

| ID | Food/farmer | Production/worker | Research/scientist | Tax BC/pop | CP deficit BC / excess CP |
| --- | ---: | ---: | ---: | ---: | ---: |
| `easy` | -2/8 (-0.25) | -4/8 (-0.50) | -4/8 (-0.50) | -2/8 (-0.25) | 88/8 (11.0) |
| `normal` | 0 | 0 | 0 | 0 | 80/8 (10.0) |
| `hard` | +2/8 (+0.25) | +4/8 (+0.50) | +4/8 (+0.50) | +2/8 (+0.25) | 72/8 (9.0) |
| `very_hard` | +4/8 (+0.50) | +8/8 (+1.00) | +8/8 (+1.00) | +3/8 (+0.375) | 68/8 (8.5) |
| `impossible` | +6/8 (+0.75) | +12/8 (+1.50) | +12/8 (+1.50) | +4/8 (+0.50) | 64/8 (8.0) |

### Derivation

The MOO2 1.50 documentation describes the standard AI difficulty table using additive productivity and income bonuses per population unit. For the original Easy/Average/Hard/Impossible anchors it gives productivity bonuses 0 / +0.5 / +1 / +2, income bonuses 0 / +0.25 / +0.5 / +0.75, food bonuses 0 / +0.25 / +0.5 / +1, and command-deficit costs 11 / 10 / 9 / 8 BC.

MOOX deliberately preserves `normal` as the existing no-Difficulty-modifier baseline. Therefore the original Average row is subtracted from the Easy/Average/Hard/Impossible resource bonuses. This yields the MOOX Easy/Normal/Hard/Impossible anchors above without changing the current Normal simulation.

`very_hard` is an explicit MOOX modernization and is frozen as the linear midpoint between Hard and Impossible for each V1 numeric field. That is why tax uses +3/8 BC/pop and command deficit uses 8.5 BC per excess CP.

### Application order and floors

Difficulty deltas are applied to the AI's per-role base output after race-specific per-role modifiers are known and before population multiplication / contextual government-morale-gravity adjustments.

For each built-in-AI colony:

- food per farmer = max(1, normal per-farmer value + Difficulty food delta);
- production per worker = max(1, normal per-worker value + Difficulty production delta);
- research per scientist = max(1, normal per-scientist value + Difficulty research delta);
- tax BC per population = max(0, normal tax-per-pop value + Difficulty tax delta).

Command-point overage uses the profile's BC-per-excess-CP value directly. The existing treasury uses float64 BC, so the exact eighth-unit contract can represent the interpolated 8.5 BC value without changing save semantics.

### Explicit V1 exclusions

The source table also documents difficulty-linked growth, spying, troop/marine and other behavior. Those are not activated by Slice 16.2 V1. Diplomacy hostility, planning quality, spying strategy and tactical intelligence remain downstream AI/system work that may consume the frozen Difficulty ID. Population-growth changes are likewise deferred until the corresponding original/MOOX growth contract is audited rather than mixed into this economy pass.

### Provenance boundary

The structural source is the MOO2 1.50 documentation's standard AI bonus table and config parameters. The original game manual independently confirms the qualitative direction: Easy opponents develop more slowly, Average is the normal-development concept, and Impossible accelerates opponent production/research.

The **Average-normalization and inserted Very-Hard midpoint are deliberate MOOX product decisions**, not claims that stock MOO2 used those exact five MOOX rows.

## Frozen player-facing summary policy

The visual selector and its info dialog may expose only values sourced from the authoritative server Difficulty catalog. The frontend must not maintain a second set of gameplay numbers.

The compact card remains image-led; detailed modifiers stay behind the Slice-16.1 information control. Localized copy may rephrase labels, but the numeric values must come from the server-owned profile.

Frozen factual payload per level:

- Easy: AI food -0.25/farmer; production -0.50/worker; research -0.50/scientist; tax -0.25 BC/pop; command deficit 11 BC/excess CP.
- Normal: baseline food/production/research/tax; command deficit 10 BC/excess CP.
- Hard: AI food +0.25/farmer; production +0.50/worker; research +0.50/scientist; tax +0.25 BC/pop; command deficit 9 BC/excess CP.
- Very Hard: AI food +0.50/farmer; production +1.00/worker; research +1.00/scientist; tax +0.375 BC/pop; command deficit 8.5 BC/excess CP.
- Impossible: AI food +0.75/farmer; production +1.50/worker; research +1.50/scientist; tax +0.50 BC/pop; command deficit 8 BC/excess CP.

The UI may also state that these modifiers apply to built-in AI empires only. It must not claim active diplomacy hostility, spying, growth, troop or tactical-intelligence modifiers in Slice 16.2 V1.

## Asset/motif direction

Gate-1 command-emblem progression remains the preferred production direction:

- one coherent visual family;
- increasing rank/weight/aggression from Easy through Impossible;
- no meme-like muscle iconography;
- no copied/traced original MOO2 art.

Preferred production format: **SVG** for the symbolic command-emblem family, because it is vector-native, small, scalable and consistent with the Slice-16.1 asset policy. Raster/WebP remains available only if the motif becomes painterly enough to justify it.

Prototype reference:

`docs/research/prototypes/SLICE_16_2_DIFFICULTY_ICON_PROGRESSION_2026-09-15.svg`

## Frozen production asset contract

The Gate-1 command-emblem progression is accepted as the production motif family.

- format: SVG;
- canvas: 1200 x 675, viewBox 0 0 1200 675;
- no embedded localized text;
- one coherent crest/shield silhouette family with increasing rank weight/aggression;
- no copied or traced MOO2 artwork;
- provenance: original MOOX procedural/vector artwork;
- runtime IDs: new-game:difficulty:<asset-option-id>.

Frozen runtime paths:

- /assets/new-game/difficulty/easy.svg
- /assets/new-game/difficulty/normal.svg
- /assets/new-game/difficulty/hard.svg
- /assets/new-game/difficulty/very_hard.svg
- /assets/new-game/difficulty/impossible.svg

Gate 3 will create the five production assets and register them in the New Game manifest. The Gate-1 research prototype remains design evidence, not a runtime asset.

## Gate-2 checklist result

- [x] Freeze supported difficulty set and server identifiers.
- [x] Freeze display labels and modifier summary text.
- [x] Freeze final image/icon motif and asset format.

Permanent Gate-2 evidence is this document.

## Gate-3 implementation target after freeze

Once Gate 2 is complete, Gate 3 should add:

- `DifficultyID` to the authoritative New Game settings contract;
- five accepted identifiers and validation;
- server-owned Difficulty catalog/profile data;
- deterministic New Game/gameplay application;
- persistence/save-resume compatibility;
- typed web API contract;
- Slice-16.1 `VisualSelector<DifficultyID>` integration;
- facts sourced from the authoritative catalog;
- deterministic fixtures/tests for all supported levels.

No runtime/server code is added by this Gate-2 starter draft.
