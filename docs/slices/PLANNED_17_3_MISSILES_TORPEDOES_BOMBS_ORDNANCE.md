# Planned Slice 17.3 - Missiles, Torpedoes and Bomb Ordnance

Status: **active; Gates 1-2 complete; Gate 3 implementation/review active**.

## Objective

Add authoritative design-side ordnance breadth without prematurely pulling the complete projectile/point-defense/planet-bombardment tactical simulation into the Ship Designer program.

## Scope candidates

- missile weapon identities;
- torpedo identities;
- bomb identities;
- technology unlocks;
- mount slots and counts;
- ammunition / shot-count fields where required by the original design;
- base and miniaturized Space/Cost;
- original legal modifiers/upgrades where they are part of the design contract;
- immutable ShipDesignSpec representation;
- construction/persistence and BattleSession handoff metadata.

## Scope guard

17.3 answers **what ordnance is installed on the ship**. Planned Slice 17.7 owns the complete Tactical ordnance consumer, including:

- projectile flight;
- missile interception;
- point-defense engagement;
- evasion/ECM resolution;
- impact/explosion timing;
- complete bomb-vs-planet combat resolution.

Gate 1 may pull in only bounded tactical consumers required to validate the design data honestly; complete projectile/interception/impact behavior remains explicitly owned by 17.7.

## Gate 1 - ordnance audit

- [x] Inventory normalized missile/torpedo/bomb applications and missing data.
- [x] Re-audit technology, Space, Cost, damage, ammo/count and miniaturization.
- [x] Re-audit legal ordnance modifiers and compatibility restrictions.
- [x] Re-audit current ShipWeaponMount/Core shape for ordnance representation.
- [x] Separate design-time metadata from later tactical runtime state.
- [x] Define representative 17.0 labs: missile cruiser, torpedo ship, bomber/bomb carrier where applicable.

Gate-1 evidence: `docs/research/SLICE_17_3_GATE1_ORDNANCE_AUDIT_2026-09-23.md`.

## Gate 2 - ordnance contract freeze

Gate-2 working contract: `docs/research/SLICE_17_3_GATE2_ORDNANCE_CONTRACT_FREEZE_2026-09-23.md`.

- [x] Freeze supported missile/torpedo/bomb matrix.
- [x] Freeze mount/count/ammunition representation.
- [x] Freeze legal modifiers and stable rejection reasons.
- [x] Freeze cost/space/miniaturization formulas.
- [x] Freeze persistence and immutable built-Ship snapshot semantics.
- [x] Freeze Tactical handoff fields and explicit deferred behavior with Slice 17.7 as the downstream ordnance owner.

## Gate 3 - implementation

- [ ] Normalize/complete ordnance data required by the frozen matrix.
- [ ] Expand designer catalog and save validation.
- [ ] Extend ShipDesignSpec weapon snapshots without fake tactical state.
- [ ] Extend Designer HMI for ordnance selection/counts/modifiers.
- [ ] Preserve Colony Construction/persistence/revision behavior.
- [ ] Add reference-lab construction and combat-handoff fixtures.

## Gate 4 - QA + close

- [ ] Legal/illegal ordnance combinations match the frozen contract.
- [ ] Space/Cost/technology/ammo snapshots are deterministic.
- [ ] High-tech ordnance can be tested immediately through 17.0.
- [ ] Built Ships preserve exact ordnance snapshots after design revision.
- [ ] Tactical consumers either use the frozen metadata or report unsupported behavior explicitly.
- [ ] Full tests/vet/web/browser/persistence checks and `git diff --check` pass.

## Exit criterion

Supported missiles, torpedoes and bombs are first-class authoritative design components that can be selected, saved, built and handed unchanged to planned Slice 17.7 without pretending their full tactical projectile behavior is already implemented in 17.3.
