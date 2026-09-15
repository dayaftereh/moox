# Open Slice 16.2 - Difficulty visual selector

Opened: 2026-09-15
Planned specification: `docs/slices/PLANNED_16_2_DIFFICULTY_VISUAL_SELECTOR.md`
Permanent Gate-1 evidence: `docs/research/SLICE_16_2_GATE1_DIFFICULTY_AUDIT_2026-09-15.md`
Gate-2 working draft (superseded): `docs/research/SLICE_16_2_GATE2_DIFFICULTY_FREEZE_DRAFT_2026-09-15.md`
Permanent Gate-2 freeze evidence: `docs/research/SLICE_16_2_GATE2_DIFFICULTY_FREEZE_2026-09-15.md`
Permanent Gate-3 implementation evidence: `docs/research/SLICE_16_2_GATE3_IMPLEMENTATION_2026-09-15.md`
Gate-4 visual redesign review candidate: `docs/research/SLICE_16_2_GATE4_DIFFICULTY_ICON_REDESIGN_V2_2026-09-15.md`

## Recovery state

- Branch: `main`.
- Opened from clean closed Slice-16.1 baseline `4f9046d`.
- Gates 1-3 are complete; Gate 4 is active at visual-review candidate stage.
- Slice 16.1 shared visual-selection grammar is the required UI foundation.
- Difficulty server contract, built-in-AI V1 gameplay effects, deterministic fixtures, production SVGs and shared web selector are implemented and regression-tested.

## Gate 1 checklist

- [x] Re-check original difficulty levels and exact gameplay effects.
- [x] Map existing server/runtime difficulty model, if any.
- [x] Decide whether the five MOOX display labels map 1:1 or represent deliberate modern naming.
- [x] Prototype all five icon/art states in one coherent progression.

## Current findings

- Original MOO2 public labels are Tutor / Easy / Average / Hard / Impossible.
- Original-manual qualitative effects are verified, but the full stock numeric modifier table is not yet authoritative in MOOX.
- Existing local reverse engineering proves a broader game-setting byte at `0x21CB0` affects multiple NPC/difficulty-related branches, but its complete public difficulty mapping remains unresolved.
- Current MOOX New Game/server contract now owns `difficulty_id` plus a read-only Difficulty catalog endpoint.
- Preferred MOOX display labels Easy / Normal / Hard / Very Hard / Impossible are implemented over stable server IDs; all five are currently supported.
- Normal should anchor to original Average baseline semantics; Hard and Impossible retain their original conceptual anchors.
- Tutor remains separate from main Difficulty; Very Hard uses the explicitly frozen MOOX midpoint mechanics.
- The first command-crest production family was replaced during Gate 4 by a cleaner deterministic threat-sigil V2 family after user feedback about overlay density; Gate 4 remains open for visual acceptance.

## Next

User reviews the V2 Difficulty threat-sigil family. If accepted, complete the remaining Gate-4 deterministic/intended-effect/full-regression closure checks and close Slice 16.2. If visual changes are requested, tune/regenerate the five SVGs without changing the frozen Difficulty mechanics.
