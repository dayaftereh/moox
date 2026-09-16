# Open slice 16.4 - Preset race portraits and carousel

Status: **Gate 3 complete / Gate 4 next**

Parent: `docs/slices/PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`
Plan: `docs/slices/PLANNED_16_4_PRESET_RACE_PORTRAITS_CAROUSEL.md`
Permanent Gate-1 evidence: `docs/research/SLICE_16_4_GATE1_PRESET_RACE_AUDIT_2026-09-16.md`
Prototype/format evidence: `docs/research/prototypes/SLICE_16_4_RACE_PORTRAITS_2026-09-16/FORMAT_COMPARISON.md`

## Gate 1 checklist

- [x] Inventory all 13 normalized preset races and their authoritative identifiers.
- [x] Classify each as runtime-ready, bounded-fix, or blocked by missing mechanics.
- [x] Extract evidence-backed visual/lore/trait cues useful for original MOOX art direction.
- [x] Write one short visual-DNA record per candidate race.
- [x] Prototype at least three deliberately different race portraits with one common art style.
- [x] Compare PNG/WebP raster quality, file size and crop behavior; use SVG only where appropriate.
- [x] Verify 390 px mobile crop and readable carousel hierarchy.

## Guardrails

- Runtime portraits must be original MOOX art; do not copy/trace/repaint original MOO2 portraits.
- Gate 1 is audit/research/prototype only. Do not freeze Gate 2 or bind the full race selector yet.
- Distinguish authoritative race mechanics/lore from visual interpretation/inference.

## Gate-1 findings

- 13 canonical preset race IDs/traits are normalized and stable.
- Current New Game validation still hard-whitelists Human + Darlok, but underlying mechanics coverage is much broader.
- Klackon is the cleanest current-scope runtime-ready preset; Darlok/Elerian/Gnolam are blocked by missing defining subsystem families; the others are bounded-fix candidates.
- Original HELP records provide sufficient anatomy/culture cues for original MOOX art without copying original portraits.
- Alkari/Meklar/Silicoid research prototypes prove one shared 4:5 visual system can preserve radically different silhouettes/materials and survive centered 1:1/mobile crops.
- WebP at quality 82 used only ~5.6-6.6% of the prototype PNG byte size with ~0.996 SSIM.

## Gate 2 checklist

- [x] Freeze the supported preset-race set for Slice 16 and classify unsupported presets honestly.
- [x] Freeze bounded-fix scope: what Slice 16.4 implements now vs what remains deferred to later gameplay-system slices.
- [x] Freeze server-owned preset-race catalog/fact contract and New Game support semantics.
- [x] Freeze portrait style bible, 4:5 master, safe area and WebP runtime format.
- [x] Freeze semantic asset paths/manifest naming for portraits and emblems.
- [x] Freeze compact race-card facts without promising unavailable mechanics.
- [x] Freeze shared carousel behavior, accessibility and 390 px mobile hierarchy.
- [x] Record representative-portrait visual acceptance as provisional: current Gate-1 prototypes are accepted for integration exploration, but final painterly art may still change.

## Gate-2 freeze

Permanent evidence: `docs/research/SLICE_16_4_GATE2_FREEZE_2026-09-16.md`.

Frozen implementation direction:

- all 13 canonical races are catalog-visible in normalized order;
- Human + Klackon are the first local-player supported set;
- Darlok remains the fixed AI baseline and planned for local-player selection;
- broader race parity mechanics remain deferred;
- runtime portraits use the frozen 4:5 WebP / safe-area / manifest contract;
- Human/Klackon/Darlok need unique runtime portraits in Gate 3;
- current representative portrait direction is provisionally accepted for integration, not permanently art-locked.

## Gate 3 checklist

- [x] Produce/curate one original portrait for every supported race.
- [x] Add race asset manifest and optional emblems.
- [x] Expose authoritative preset race catalog data to the New Game HMI.
- [x] Implement left/right race carousel using Slice 16.1 grammar.
- [x] Reuse the same race portrait identity for opponent composition where possible.
- [x] Add deterministic New Game fixtures for each supported race/start tuple.

Permanent Gate-3 evidence: `docs/research/SLICE_16_4_GATE3_IMPLEMENTATION_2026-09-16.md`.

Gate-3 implementation summary:

- authoritative 13-race server catalog is live;
- Human + Klackon are local-player supported, Darlok remains fixed AI baseline;
- Human/Klackon/Darlok canonical WebP portraits and manifest are integrated;
- shared race VisualSelector renders all 13 profiles, locks Create on planned races and submits the selected supported race;
- live Klackon -> POST -> authoritative snapshot mapping passed;
- full Go/web/browser/external-health regression passed.

## Next

Gate 4 closure acceptance is next. Do not open or execute Gate 4 until explicitly continued.
