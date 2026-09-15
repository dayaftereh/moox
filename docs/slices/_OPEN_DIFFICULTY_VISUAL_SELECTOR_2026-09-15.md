# Open Slice 16.2 - Difficulty visual selector

Opened: 2026-09-15
Planned specification: `docs/slices/PLANNED_16_2_DIFFICULTY_VISUAL_SELECTOR.md`
Permanent Gate-1 evidence: `docs/research/SLICE_16_2_GATE1_DIFFICULTY_AUDIT_2026-09-15.md`
Gate-2 working draft: `docs/research/SLICE_16_2_GATE2_DIFFICULTY_FREEZE_DRAFT_2026-09-15.md`

## Recovery state

- Branch: `main`.
- Opened from clean closed Slice-16.1 baseline `4f9046d`.
- Gate 1 audit is complete; Gate 2 is active in a fresh explicit session.
- Slice 16.1 shared visual-selection grammar is the required UI foundation.
- No Difficulty gameplay/server implementation has been made yet.

## Gate 1 checklist

- [x] Re-check original difficulty levels and exact gameplay effects.
- [x] Map existing server/runtime difficulty model, if any.
- [x] Decide whether the five MOOX display labels map 1:1 or represent deliberate modern naming.
- [x] Prototype all five icon/art states in one coherent progression.

## Current findings

- Original MOO2 public labels are Tutor / Easy / Average / Hard / Impossible.
- Original-manual qualitative effects are verified, but the full stock numeric modifier table is not yet authoritative in MOOX.
- Existing local reverse engineering proves a broader game-setting byte at `0x21CB0` affects multiple NPC/difficulty-related branches, but its complete public difficulty mapping remains unresolved.
- Current MOOX New Game/server schema has no authoritative Difficulty field.
- Preferred MOOX display labels remain Easy / Normal / Hard / Very Hard / Impossible as deliberate modern UX naming, not a claim of original naming parity.
- Normal should anchor to original Average baseline semantics; Hard and Impossible retain their original conceptual anchors.
- Easy must not silently absorb Tutor-only behavior/restrictions; Very Hard is a deliberate MOOX modernization requiring explicit mechanics before implementation.
- A five-state original command-emblem progression prototype exists under `docs/research/prototypes/` and is not a runtime asset.

## Next

Gate 2 freeze discussion: supported difficulty set/server IDs, Tutor policy, explicit Very-Hard mechanics, evidence-backed numeric effects, concise modifier-summary text, and final emblem/asset format. Do not start Gate-3 implementation before those decisions are frozen.
