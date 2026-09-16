# Open slice 16.5 - Starting technology visual selector

Status: **Gate 2 active**

Parent: `docs/slices/PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`
Plan: `docs/slices/PLANNED_16_5_STARTING_TECH_VISUAL_SELECTOR.md`
Permanent Gate-1 evidence: `docs/research/SLICE_16_5_GATE1_STARTING_TECH_AUDIT_2026-09-16.md`
Prototype evidence: `docs/research/prototypes/SLICE_16_5_STARTING_TECH_2026-09-16/`

## Gate 1 checklist

- [x] Re-check original and normalized New Game technology-level options.
- [x] Confirm whether the supported target is exactly three levels and their authoritative IDs.
- [x] Inventory exact start-state effects/granted technologies for each level.
- [x] Identify race-specific interactions and RNG/order dependencies.
- [x] Prototype the three-image visual progression using the Slice-16.1 selection grammar.
- [x] Verify the proposed progression remains distinct and readable at 390 px mobile.

## Gate 1 result

- Canonical IDs are exactly `pre_warp`, `average` and `advanced`.
- Average is the only complete production New Game start currently accepted; Pre-Warp needs its no-ship start bootstrap and Advanced needs preference generation plus larger-empire/fleet/population bootstrap parity.
- Advanced technology grants remain server-owned because the extra 19 research-field grants vary with seed, race and cross-Empire initialization order.
- The three-stage visual progression and 390 px preview are suitable to carry forward into Gate 2.

## Guardrails

- Gate 1 was audit/research/prototype only; no New Game validation/runtime breadth was widened before the Gate-2 freeze.
- Separate original MOO2 evidence, normalized MOOX contracts and current runtime limitations.
- Do not hard-code hidden technology grants in the frontend.
- Advanced start must remain server-owned because its 19 extra grants are RNG-, race- and cross-Empire-order dependent.

## Gate 2 checklist

- [ ] Freeze the supported technology-start set and player-facing labels for `pre_warp`, `average` and `advanced`.
- [ ] Freeze the complete authoritative start-state contract for each level, not only technology ownership.
- [ ] Freeze the Slice-16.5 implementation boundary for Pre-Warp and Advanced bootstrap parity.
- [ ] Freeze server-owned explanatory facts and prohibit fixed frontend Advanced grant lists/counts beyond stable evidence-backed facts.
- [ ] Freeze Advanced seed/race/cross-Empire initialization semantics and deterministic fixture expectations.
- [ ] Freeze original three-stage art direction, asset format, safe area and 390 px behavior.
- [ ] Freeze shared selector/card accessibility and navigation behavior inherited from Slice 16.1.

## Next

Write permanent Gate-2 freeze evidence from the Gate-1 audit before changing runtime behavior. Gate 3 must implement only the resulting frozen full-start/server/UI contract.
