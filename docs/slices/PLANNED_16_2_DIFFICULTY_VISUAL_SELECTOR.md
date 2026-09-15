# Planned slice 16.2 - Difficulty visual selector

Status: **open; Gates 1-3 complete, Gate 4 next (2026-09-15)**.

Parent: `PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`.
Depends on: Slice 16.1 shared visual selection grammar.

## Objective

Turn difficulty selection into a highly readable visual carousel rather than a small dropdown. Difficulty should feel immediately understandable before the player reads any detail text.

## Binding visual direction

Preferred MOOX presentation labels:

1. **Easy**
2. **Normal**
3. **Hard**
4. **Very Hard**
5. **Impossible**

Gate 1 must still audit the original/runtime difficulty semantics and modifier tables before gameplay values are bound to those display labels. If evidence requires different internal names, the MOOX presentation may still use the five labels above as a deliberate UX vocabulary, but the mapping must be documented explicitly.

## Artwork concept

Each difficulty gets its own original MOOX image/icon family. The visual progression should read without text.

Suggested motif:

- Easy: friendly approval / open hand / subtle thumbs-up or relaxed pilot cue.
- Normal: balanced neutral crest / steady fist / centered command insignia.
- Hard: flexed arm / stronger military silhouette.
- Very Hard: heavily flexed/armored arm or more aggressive commander emblem.
- Impossible: exaggerated elite/alien power silhouette, cracked-metal badge, or maximal-strength visual.

The motif can use a stylized arm/muscle progression as the user suggested, but it must remain tasteful and legible at small sizes rather than becoming a meme icon.

## Information strip

Under the artwork, show only concise evidence-backed consequences, for example once Gate 1 confirms them:

- AI advantage/handicap;
- economy/production modifiers;
- research modifiers;
- any start-state difference.

Do not invent numbers in the frontend. Every displayed fact must come from normalized/runtime settings data or a server-owned catalog.

## Interaction

- Previous/next arrows use the shared 16.1 selector.
- Difficulty changes are preview-only until game creation.
- Current selection is visually unmistakable.
- Disabled/unsupported levels must be visibly unavailable rather than silently coerced.

## Gate 1 - audit

- [x] Re-check original difficulty levels and exact gameplay effects.
- [x] Map existing server/runtime difficulty model, if any.
- [x] Decide whether the five MOOX display labels map 1:1 or represent deliberate modern naming.
- [x] Prototype all five icon/art states in one coherent progression.

Permanent Gate-1 evidence: `docs/research/SLICE_16_2_GATE1_DIFFICULTY_AUDIT_2026-09-15.md`.

## Gate 2 - freeze

Gate-2 working draft (superseded): `docs/research/SLICE_16_2_GATE2_DIFFICULTY_FREEZE_DRAFT_2026-09-15.md`.

Permanent Gate-2 freeze evidence: `docs/research/SLICE_16_2_GATE2_DIFFICULTY_FREEZE_2026-09-15.md`.

- [x] Freeze supported difficulty set and server identifiers.
- [x] Freeze display labels and modifier summary text.
- [x] Freeze final image/icon motif and asset format.

## Gate 3 - implementation

- [x] Add authoritative difficulty setting/catalog support.
- [x] Add the visual carousel selector.
- [x] Bind concise modifier facts from server/normalized data.
- [x] Add deterministic New Game fixtures for each supported level.

## Gate 4 - close

- [ ] Equal seed + equal settings remains byte/deterministically equal.
- [ ] Difficulty changes produce only their frozen intended effects.
- [ ] Desktop/mobile carousel QA.
- [ ] Full tests/build/diff checks.

## Exit criterion

Difficulty is understandable from image + title at a glance, has evidence-backed explanatory facts, and creates the exact authoritative difficulty contract selected by the player.