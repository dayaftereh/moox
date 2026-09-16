# Open slice 16.4 - Preset race portraits and carousel

Status: **Gate 1 complete / Gate 2 next**

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

## Next

Gate 2 should freeze the supported-race policy and bounded-fix scope, the authoritative race-catalog/fact contract, final art style/safe-zone, WebP runtime format and carousel behavior. Do not start Gate 2 until explicitly continued.
