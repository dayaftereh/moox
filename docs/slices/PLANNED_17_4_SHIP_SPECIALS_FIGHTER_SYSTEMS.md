# Planned Slice 17.4 - Ship Specials and Fighter Systems

Status: **closed/accepted; Gates 1-4 complete (2026-09-28)**.

## Objective

Add the supported ship-special and fighter-system design breadth needed for meaningful late-game designs while keeping unimplemented Tactical fighter/combat-special consumers explicitly owned by planned Slice 17.8.

## Scope candidates

Gate 1 must inventory and classify original special systems, including families such as:

- Reinforced Hull;
- Heavy Armor and other hull/defensive specials not represented as the primary armor material;
- Battle Pods / space-capacity modifiers;
- inertial/mobility systems;
- cloaking / stealth systems;
- scanners/sensors;
- transport/boarding-related systems;
- repair/regeneration systems;
- offensive/defensive special systems;
- Fighter/Bomber/Interceptor-related bays or carrier systems;
- other normalized specials required by the original design catalog.

This list is an audit starting point, not an automatic promise that every original special becomes fully tactically functional in this slice.

## Required classification

Every accepted special must be labeled as one of:

1. **design + strategic effect complete**;
2. **design + already-supported tactical effect complete**;
3. **design representation complete, Tactical consumer explicitly deferred to Slice 17.8**;
4. **not yet supported/locked**.

No special may appear as fully supported without a real owner for its gameplay effect.

## Gate 1 - specials/fighter audit

- [x] Inventory original/normalized specials and fighter systems.
- [x] Re-audit technology, Space, Cost and miniaturization.
- [x] Re-audit stacking/exclusivity/quantity limits.
- [x] Identify strategic derived-stat effects implementable now.
- [x] Identify existing Tactical consumers versus missing combat subsystems.
- [x] Re-audit fighter/bomber carrier representation and the exact combat dependencies handed to Slice 17.8.
- [x] Define high-tech/Doom-Star reference fixtures in 17.0.

Evidence: docs/research/SLICE_17_4_GATE1_SPECIALS_FIGHTERS_AUDIT_2026-09-26.md - Gate 1 complete.

## Gate 2 - specials contract freeze

- [x] Freeze supported special/fighter matrix.
- [x] Freeze per-special support classification.
- [x] Freeze compatibility, stacking, quantity and rejection rules.
- [x] Freeze cost/space/miniaturization.
- [x] Freeze ShipDesignSpec snapshot representation.
- [x] Freeze Slice-17.8 Tactical ownership for every deferred fighter/combat-special consumer.

Evidence: docs/research/SLICE_17_4_GATE2_SPECIALS_FIGHTERS_CONTRACT_FREEZE_2026-09-27.md - Gate 2 complete.

## Gate 3 - implementation

- [x] Add/finalize special/fighter data and authoritative validation.
- [x] Implement accepted strategic/derived-stat effects.
- [x] Implement bounded existing-Tactical consumers where frozen.
- [x] Extend Designer HMI with supported specials/fighter systems.
- [x] Preserve immutable construction/persistence snapshots.
- [x] Add 17.0 high-tech/Doom-Star lab fixtures and targeted combat handoffs.

Evidence: docs/research/SLICE_17_4_GATE3_IMPLEMENTATION_2026-09-28.md - Gate 3 complete.

## Gate 4 - QA + close

- [x] Every visible supported special has the promised effect/support classification.
- [x] No unsupported tactical effect is silently simulated or ignored.
- [x] Compatibility/space/cost/technology rules pass deterministic matrices.
- [x] Late-game designs are practical to test through reference labs.
- [x] Construction/persistence/revision remain exact.
- [x] Full tests/vet/web/browser checks and `git diff --check` pass.

Evidence: docs/research/SLICE_17_4_GATE4_ACCEPTANCE_CLOSE_2026-09-28.md - Gate 4 accepted/closed.

## Exit criterion

Supported specials and fighter/carrier systems are authoritative design objects with explicit, testable gameplay ownership; every accepted deferred combat consumer is handed unchanged to planned Slice 17.8 rather than being disguised as already complete.
