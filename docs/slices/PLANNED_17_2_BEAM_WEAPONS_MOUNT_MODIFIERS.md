# Planned Slice 17.2 - Beam Weapons and Beam Mount Modifiers

Status: **active; Gates 1-3 complete / Gate 4 next (2026-09-21)**.

## Objective

Expand the current single standard Laser mount into an original-evidenced Beam-weapon design system that later Tactical combat can consume.

## Scope candidates

- normalized Beam weapon identities;
- technology unlocks;
- mount slots/counts;
- base and miniaturized Space/Cost;
- min/max damage metadata;
- range/dissipation metadata;
- applicable original Beam mount modifiers such as Point Defense, Heavy Mount, Autofire, Armor Piercing, Continuous and No Range Dissipation where evidence supports them;
- tactical compatibility metadata with the existing Slice-07 Laser baseline.

Gate 1 must decide the exact supported modifier subset rather than assuming every HELP-listed modifier applies to every Beam.

## Scope guard

17.2 owns Beam **design legality and authoritative snapshots**. It only extends Tactical runtime behavior where the required implementation is bounded and explicitly frozen. Full broad Tactical combat remains downstream.

## Gate 1 - Beam audit

- [x] Inventory normalized Beam weapons and missing original data.
- [x] Re-audit technology, damage, Space, Cost, range and miniaturization.
- [x] Re-audit modifier availability/compatibility per Beam.
- [x] Re-audit weapon slot/count limits and original design-array semantics.
- [x] Map supported Beam metadata onto current ShipWeaponMount without lossy hacks.
- [x] Identify tactical consumers already implemented versus deferred.

Gate-1 evidence: docs/research/SLICE_17_2_GATE1_BEAM_WEAPONS_AUDIT_2026-09-20.md.

## Gate 2 - Beam contract freeze

- [x] Freeze supported Beam matrix.
- [x] Freeze slot/count model.
- [x] Freeze modifier compatibility/rejection reasons.
- [x] Freeze cost/space/miniaturization rules.
- [x] Freeze ShipDesignSpec snapshot shape.
- [x] Freeze 17.0 Designer/Combat Lab fixtures.

Gate-2 contract: docs/research/SLICE_17_2_GATE2_BEAM_CONTRACT_FREEZE_2026-09-20.md.
Gate-2 HMI add-flow amendment: docs/research/SLICE_17_2_GATE2_HMI_ADD_FLOW_AMENDMENT_2026-09-20.md.

## Gate 3 - implementation

- [x] Expand Beam catalog/rules.
- [x] Expand save/design validation and weapon mounts.
- [x] Implement selected modifier metadata and bounded tactical consumers.
- [x] Update Designer HMI for multiple Beam slots/counts/modifiers.
- [x] Preserve construction/persistence/snapshot semantics.
- [x] Add deterministic Tactical/Designer Lab regressions.

Gate-3 implementation: docs/research/SLICE_17_2_GATE3_IMPLEMENTATION_2026-09-21.md.
Gate-3 mobile HMI amendment: docs/research/SLICE_17_2_GATE3_MOBILE_WEAPON_SELECTOR_AMENDMENT_2026-09-21.md.

## Gate 4 - QA + close

- [ ] Legal/illegal Beam combinations match the frozen matrix.
- [ ] Space/Cost and technology locks are exact for tested fixtures.
- [ ] Existing canonical Laser tactical fixture remains green.
- [ ] New supported Beam snapshots reach BattleSession without translation hacks.
- [ ] Construction/persistence/browser QA passes.
- [ ] Full tests/vet and `git diff --check` pass.

## Exit criterion

Supported Beam weapons can be selected, modified, costed, saved, built and delivered as authoritative immutable ship snapshots, with explicit tactical support/defer status for every frozen Beam feature.
