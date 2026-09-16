# Open slice 16.5 - Starting technology visual selector

Status: **Gate 1 active**

Parent: `docs/slices/PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`
Plan: `docs/slices/PLANNED_16_5_STARTING_TECH_VISUAL_SELECTOR.md`

## Gate 1 checklist

- [ ] Re-check original and normalized New Game technology-level options.
- [ ] Confirm whether the supported target is exactly three levels and their authoritative IDs.
- [ ] Inventory exact start-state effects/granted technologies for each level.
- [ ] Identify race-specific interactions and RNG/order dependencies.
- [ ] Prototype the three-image visual progression using the Slice-16.1 selection grammar.
- [ ] Verify the proposed progression remains distinct and readable at 390 px mobile.

## Guardrails

- Gate 1 is audit/research/prototype only; do not widen New Game validation or freeze Gate 2 yet.
- Separate original MOO2 evidence, normalized MOOX contracts and current runtime limitations.
- Do not hard-code hidden technology grants in the frontend.
- Advanced start must remain server-owned because its 19 extra grants are RNG-, race- and cross-Empire-order dependent.

## Next

Audit the existing Pre-Warp/Average/Advanced research evidence and runtime paths, document the exact supported semantics/gaps, then build a representative three-stage visual prototype before Gate 2.
