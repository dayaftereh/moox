# Planned Slice 17.4 - Ship Specials and Fighter Systems

Status: **planned; child of Slice 17; not open**.

## Objective

Add the supported ship-special and fighter-system design breadth needed for meaningful late-game designs while keeping unimplemented tactical mechanics explicitly downstream.

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
3. **design representation complete, tactical consumer explicitly deferred**;
4. **not yet supported/locked**.

No special may appear as fully supported without a real owner for its gameplay effect.

## Gate 1 - specials/fighter audit

- [ ] Inventory original/normalized specials and fighter systems.
- [ ] Re-audit technology, Space, Cost and miniaturization.
- [ ] Re-audit stacking/exclusivity/quantity limits.
- [ ] Identify strategic derived-stat effects implementable now.
- [ ] Identify existing Tactical consumers versus missing combat subsystems.
- [ ] Re-audit fighter/bomber carrier representation and downstream combat dependencies.
- [ ] Define high-tech/Doom-Star reference fixtures in 17.0.

## Gate 2 - specials contract freeze

- [ ] Freeze supported special/fighter matrix.
- [ ] Freeze per-special support classification.
- [ ] Freeze compatibility, stacking, quantity and rejection rules.
- [ ] Freeze cost/space/miniaturization.
- [ ] Freeze ShipDesignSpec snapshot representation.
- [ ] Freeze downstream Tactical ownership for deferred consumers.

## Gate 3 - implementation

- [ ] Add/finalize special/fighter data and authoritative validation.
- [ ] Implement accepted strategic/derived-stat effects.
- [ ] Implement bounded existing-Tactical consumers where frozen.
- [ ] Extend Designer HMI with supported specials/fighter systems.
- [ ] Preserve immutable construction/persistence snapshots.
- [ ] Add 17.0 high-tech/Doom-Star lab fixtures and targeted combat handoffs.

## Gate 4 - QA + close

- [ ] Every visible supported special has the promised effect/support classification.
- [ ] No unsupported tactical effect is silently simulated or ignored.
- [ ] Compatibility/space/cost/technology rules pass deterministic matrices.
- [ ] Late-game designs are practical to test through reference labs.
- [ ] Construction/persistence/revision remain exact.
- [ ] Full tests/vet/web/browser checks and `git diff --check` pass.

## Exit criterion

Supported specials and fighter/carrier systems are authoritative design objects with explicit, testable gameplay ownership, enabling meaningful advanced ship designs without disguising later Tactical work as already complete.
