# Planned slice 16.4 - Preset race portraits and carousel

Status: **open; Gate 3 active (2026-09-16)**.

Parent: `PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`.
Depends on: Slice 16.1 shared visual selection grammar and Gate-1 compatibility audit of the normalized preset races.

## Objective

Give every supported preset race a memorable visual identity in New Game. Race choice must no longer be a plain text/dropdown decision: the selected race is represented by a large original MOOX portrait, race name and a compact set of meaningful facts, browsed with left/right arrows.

The normalized ruleset currently contains **13 preset races**. Slice 16 Gate 1 determines which are runtime-compatible immediately and what mechanics are still blockers. Every race accepted into the Slice-16 runtime contract must receive its own portrait asset before Gate 4 closure.

## Binding product direction

- One unique, original portrait per supported preset race.
- Race selection uses the common carousel with previous/next arrows.
- The portrait is the visual focus; text supplements rather than replaces it.
- The user should be able to imagine what the species/civilization looks and feels like from the portrait alone.
- Portraits may evoke classic 1990s space-4X painted sci-fi art direction, but **must not copy, trace, repaint or reuse original Master of Orion II artwork**.
- Race silhouettes, anatomy, clothing, architecture motifs and palettes should be original MOOX interpretations informed by evidence-backed race traits/lore, not one-to-one reproductions of original portraits.

## Runtime asset strategy

Default direction for race portraits:

- painterly/organic portraits ship as optimized **PNG or WebP raster assets**;
- keep a larger editable/generation source or documented generation recipe in the art workflow;
- **SVG is preferred for the portrait frame, race emblem, decorative geometry, badges and simple symbolic overlays**, not forced for complex faces/fur/skin/material painting;
- a full-vector race portrait is allowed only if the chosen race/style genuinely benefits from it and remains visually consistent with the raster portrait family.

This deliberately avoids turning complex character illustration into brittle hand-authored SVG solely for implementation convenience.

## Portrait art bible

All race portraits should share one composition system so the carousel feels like one game, not unrelated generated pictures.

Recommended baseline:

- portrait-oriented master composition, approximately 4:5;
- head/upper-torso or equivalent species focal anatomy;
- eyes/facial focal point kept inside a safe central region for mobile crops;
- three-quarter or strong frontal pose;
- dark space/civilization background with subtle race-specific environmental cues;
- cinematic rim/key lighting consistent across the family;
- clear silhouette at 128-256 px preview size;
- no text baked into the artwork;
- no real-world logos, copyrighted franchise insignia or recognizable borrowed characters.

Each race gets a small **visual DNA record** before image generation, for example:

- body/anatomy language;
- material/clothing language;
- technology/culture cues;
- dominant/subordinate palette;
- emotional/leadership tone;
- forbidden similarities/cliches to avoid.

The generation/review process should keep the family stylistically coherent while making silhouettes visibly distinct.

## Suggested repository contract

Candidate semantic IDs/paths (final convention frozen in Gate 2):

```text
web/public/assets/races/<race-id>/portrait.webp
web/public/assets/races/<race-id>/portrait.png        # optional lossless/master runtime variant
web/public/assets/races/<race-id>/emblem.svg
web/public/assets/races/race-portrait-manifest.json
```

Manifest metadata should include at least:

- stable race ID;
- portrait asset path;
- emblem asset path when present;
- alt/accessibility name key;
- crop/safe-area metadata if required;
- asset/version provenance note.

## Race card facts

Under the portrait, show a compact evidence-backed summary from the authoritative preset catalog, such as:

- government;
- strongest defining traits/modifiers;
- homeworld/environment cue if authoritative;
- starting technologies or special start rules where applicable.

Do not overload the card with the full Race Designer trait sheet. A details panel may show more, but the primary carousel remains visually scannable.

## Player vs opponent usage

The same portrait assets should be reusable for:

- Human race selection;
- opponent race assignment in Slice 16.6;
- later Diplomacy/Intelligence/Race information surfaces.

This creates one canonical race visual identity instead of separate New Game-only art.

## Gate 1 - race/runtime/art audit

- [x] Inventory all 13 normalized preset races and their authoritative identifiers.
- [x] Classify each as runtime-ready, bounded-fix, or blocked by missing mechanics.
- [x] Extract the evidence-backed visual/lore/trait cues useful for original MOOX art direction.
- [x] Write one short visual-DNA record per candidate race.
- [x] Prototype at least three deliberately different race portraits with one common art style.
- [x] Compare PNG/WebP raster quality, file size and crop behavior; use SVG only where appropriate.
- [x] Verify 390 px mobile crop and readable carousel hierarchy.

Permanent Gate-1 evidence: `docs/research/SLICE_16_4_GATE1_PRESET_RACE_AUDIT_2026-09-16.md`.
Prototype/format evidence: `docs/research/prototypes/SLICE_16_4_RACE_PORTRAITS_2026-09-16/FORMAT_COMPARISON.md`.

## Gate 2 - freeze

- [x] Freeze the supported preset-race set for Slice 16.
- [x] Freeze portrait style bible, aspect ratio, safe area and runtime format.
- [x] Freeze manifest/path naming.
- [x] Freeze the compact race-card facts.
- [x] User visual review of representative portraits before generating/finalizing the complete supported set.

Permanent Gate-2 evidence: `docs/research/SLICE_16_4_GATE2_FREEZE_2026-09-16.md`.

## Gate 3 - implementation

- [ ] Produce/curate one original portrait for every supported race.
- [ ] Add race asset manifest and optional emblems.
- [ ] Expose authoritative preset race catalog data to the New Game HMI.
- [ ] Implement left/right race carousel using Slice 16.1 grammar.
- [ ] Reuse the same race portrait identity for opponent composition where possible.
- [ ] Add deterministic New Game fixtures for each supported race/start tuple.

## Gate 4 - close

- [ ] Every supported race has a non-placeholder portrait.
- [ ] No copied/traced original-game art remains in runtime assets.
- [ ] Portrait family passes desktop and real/mobile-width visual review.
- [ ] Text/ARIA remains sufficient with images disabled.
- [ ] Equal seed/settings/race tuples remain deterministic.
- [ ] Full tests/build/diff checks.

## Downstream race-mechanics follow-up

Slice 16.4 deliberately freezes only an honest early supported subset. Full preset-race mechanics completion and the Custom Race Designer are reserved in `PLANNED_22_PRESET_RACE_COMPLETION_CUSTOM_RACE_DESIGNER.md`, so deferred traits/races are not lost when this portrait/catalog slice closes.

## Exit criterion

Preset race selection feels like choosing a civilization, not selecting a database row: every supported race has its own original MOOX portrait and stable reusable visual identity, while the underlying preset mechanics remain server-authoritative and evidence-backed.