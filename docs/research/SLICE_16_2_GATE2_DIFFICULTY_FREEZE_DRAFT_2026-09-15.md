# Slice 16.2 Gate 2 - Difficulty freeze draft

Date: 2026-09-15
Status: **superseded by final Gate-2 freeze**

Gate-1 evidence: `docs/research/SLICE_16_2_GATE1_DIFFICULTY_AUDIT_2026-09-15.md`
Shared UI foundation: closed Slice 16.1.
Final freeze: `docs/research/SLICE_16_2_GATE2_DIFFICULTY_FREEZE_2026-09-15.md`.

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

## Modifier/profile contract direction

The server should own a Difficulty profile/catalog rather than scattering `if difficulty == ...` values through UI and gameplay code.

The profile must distinguish deterministic modifiers from AI policy intelligence. Candidate authoritative fields are grouped by semantics, not yet frozen as exact numeric values:

- economy/production modifier;
- research modifier;
- income/treasury modifier if original evidence supports it;
- command-point deficit/maintenance behavior where supported by original evidence;
- diplomacy/initial-attitude modifier where supported;
- start-state changes/restrictions where supported;
- explicit evidence/provenance metadata for each player-facing fact.

The final Gate-2 schema should use deterministic integer/fixed-point values rather than floating-point gameplay math where practical.

## Evidence status by level

### Easy

Evidence-backed qualitative direction:

- opponents develop more slowly in production/research than the player;
- opposing races are friendlier than baseline.

Exact numeric stock values are not frozen yet.

### Normal

Frozen semantic direction:

- baseline / no Difficulty advantage or handicap;
- corresponds conceptually to original `Average` normal-development semantics.

### Hard

Original public label exists, but exact stock numeric effect table remains unresolved in current authoritative MOOX research.

Gate-2 must not invent the numbers merely to fill the catalog.

### Very Hard

This is a deliberate MOOX modernization with no original display-label counterpart.

Gate-2 rule:

- it must be a first-class server profile/identifier;
- it must not be calculated in the frontend;
- it must not be an undocumented midpoint interpolation;
- its exact deterministic mechanics must be chosen/frozen explicitly before Gate 3.

### Impossible

Evidence-backed qualitative direction:

- significantly accelerated opponent production/research;
- hostile initial diplomatic posture.

Exact numeric stock values still require authoritative freeze.

## Player-facing summary policy

The visual selector may show only facts that come from the authoritative server/catalog contract.

Until numeric values are frozen, production UI must not claim percentage bonuses/penalties. Qualitative labels may be used only if they map directly to frozen catalog semantics.

Potential concise summary grammar after freeze:

- Easy: `AI economy/research handicap` / `friendlier rivals`
- Normal: `baseline rules` / `no difficulty modifiers`
- Hard: only frozen effects once confirmed
- Very Hard: only explicit MOOX profile facts once designed
- Impossible: `strong AI economy/research advantage` / `hostile rivals`

This wording remains draft until the underlying authoritative fields are frozen.

## Asset/motif direction

Gate-1 command-emblem progression remains the preferred production direction:

- one coherent visual family;
- increasing rank/weight/aggression from Easy through Impossible;
- no meme-like muscle iconography;
- no copied/traced original MOO2 art.

Preferred production format: **SVG** for the symbolic command-emblem family, because it is vector-native, small, scalable and consistent with the Slice-16.1 asset policy. Raster/WebP remains available only if the motif becomes painterly enough to justify it.

Prototype reference:

`docs/research/prototypes/SLICE_16_2_DIFFICULTY_ICON_PROGRESSION_2026-09-15.svg`

## Gate-2 open items before freeze

The following are intentionally **not** marked frozen yet:

1. exact evidence-backed numeric modifier table for Easy/Hard/Impossible;
2. explicit deterministic mechanics for MOOX Very Hard;
3. final server profile field list and units;
4. final concise modifier-summary strings once the authoritative fields are known;
5. production-quality five-asset SVG set derived from the accepted motif.

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
