# Planned slice 17 - Military ship design component / weapon breadth

Status: **planned / queued; not open**.

Queue position: **17 of 17**.

## Objective

Expand the intentionally minimal Slice-07 military design model into a useful canonical multi-hull/component/multi-weapon design baseline that deeper Tactical Combat can consume.

## Existing narrow boundary

The persisted design/Ship snapshot already carries hull, drive, computer, armor, shield, fuel, costs/space and weapon mounts, but runtime design validation explicitly permits only the current supported hull and slot 0 with one `laser_cannon`. Weapon modifiers remain future work.

## Dependencies

- Existing Research/technology unlock model.
- Slice 15.6 browser vertical slice establishes the presentation shell, including the first interactive Slice-15.5 Tactical baseline, while this slice remains core/server-authoritative first.

## Scope guard

In scope candidates to freeze in Gate 1:

- multiple standard hull identities;
- multiple researched drives/computers/armor/shields/fuel components;
- multiple Beam/weapon mounts and counts;
- deterministic space/cost/unlock validation;
- production/ref snapshot semantics for revised designs;
- legal-action projection usable by future Ship Designer UI.

Defer unless Gate 1 proves they fit cleanly:

- missiles/fighters/bombs full tactical behavior;
- weapon modifications and complete miniaturization;
- specials;
- refit workflow;
- tactical movement/boarding/planetary defense.

## Gate 1 - Original design-system audit

- [ ] Inventory normalized hull/component/weapon evidence and missing normalization.
- [ ] Re-check original space/cost/technology/miniaturization behavior needed for the bounded baseline.
- [ ] Define immutable built-Ship snapshot vs mutable design-slot semantics.
- [ ] Define supported component/weapon set and explicit deferred categories.
- [ ] Define deterministic design/production/combat handoff fixtures.
- [ ] Present Gate-2 design contract.

## Gate 2 - Implementation decision

- [ ] Freeze supported hull/component/weapon matrix.
- [ ] Freeze space/cost/unlock formulas.
- [ ] Freeze design revision/snapshot behavior.
- [ ] Freeze tactical compatibility contract.

## Gate 3 - Implementation

- [ ] Expand data/rules and design validation.
- [ ] Expand legal design/save/queue surfaces.
- [ ] Preserve exact Ship snapshots across later design revisions.
- [ ] Add deterministic construction and basic tactical handoff regressions.

## Gate 4 - Follow-up QA + commit + close

- [ ] Verify legal/illegal design matrix and costs.
- [ ] Verify save/load/replay and built-Ship snapshot immutability.
- [ ] Run full tests/vet/web checks and `git diff --check`.
- [ ] Update evidence/status/HISTORY and close marker.