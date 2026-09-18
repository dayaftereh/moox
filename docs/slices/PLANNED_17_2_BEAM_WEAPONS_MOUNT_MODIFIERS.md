# Planned Slice 17.2 - Beam Weapons and Beam Mount Modifiers

Status: **planned; child of Slice 17; not open**.

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

- [ ] Inventory normalized Beam weapons and missing original data.
- [ ] Re-audit technology, damage, Space, Cost, range and miniaturization.
- [ ] Re-audit modifier availability/compatibility per Beam.
- [ ] Re-audit weapon slot/count limits and original design-array semantics.
- [ ] Map supported Beam metadata onto current ShipWeaponMount without lossy hacks.
- [ ] Identify tactical consumers already implemented versus deferred.

## Gate 2 - Beam contract freeze

- [ ] Freeze supported Beam matrix.
- [ ] Freeze slot/count model.
- [ ] Freeze modifier compatibility/rejection reasons.
- [ ] Freeze cost/space/miniaturization rules.
- [ ] Freeze ShipDesignSpec snapshot shape.
- [ ] Freeze 17.0 Designer/Combat Lab fixtures.

## Gate 3 - implementation

- [ ] Expand Beam catalog/rules.
- [ ] Expand save/design validation and weapon mounts.
- [ ] Implement selected modifier metadata and bounded tactical consumers.
- [ ] Update Designer HMI for multiple Beam slots/counts/modifiers.
- [ ] Preserve construction/persistence/snapshot semantics.
- [ ] Add deterministic Tactical/Designer Lab regressions.

## Gate 4 - QA + close

- [ ] Legal/illegal Beam combinations match the frozen matrix.
- [ ] Space/Cost and technology locks are exact for tested fixtures.
- [ ] Existing canonical Laser tactical fixture remains green.
- [ ] New supported Beam snapshots reach BattleSession without translation hacks.
- [ ] Construction/persistence/browser QA passes.
- [ ] Full tests/vet and `git diff --check` pass.

## Exit criterion

Supported Beam weapons can be selected, modified, costed, saved, built and delivered as authoritative immutable ship snapshots, with explicit tactical support/defer status for every frozen Beam feature.
