# Planned Slice 17.3 - Missiles, Torpedoes and Bomb Ordnance

Status: **active; Gate 1 ordnance audit in progress (opened 2026-09-23)**.

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

- [ ] Inventory normalized missile/torpedo/bomb applications and missing data.
- [ ] Re-audit technology, Space, Cost, damage, ammo/count and miniaturization.
- [ ] Re-audit legal ordnance modifiers and compatibility restrictions.
- [ ] Re-audit current ShipWeaponMount/Core shape for ordnance representation.
- [ ] Separate design-time metadata from later tactical runtime state.
- [ ] Define representative 17.0 labs: missile cruiser, torpedo ship, bomber/bomb carrier where applicable.

## Gate 2 - ordnance contract freeze

- [ ] Freeze supported missile/torpedo/bomb matrix.
- [ ] Freeze mount/count/ammunition representation.
- [ ] Freeze legal modifiers and stable rejection reasons.
- [ ] Freeze cost/space/miniaturization formulas.
- [ ] Freeze persistence and immutable built-Ship snapshot semantics.
- [ ] Freeze Tactical handoff fields and explicit deferred behavior with Slice 17.7 as the downstream ordnance owner.

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
