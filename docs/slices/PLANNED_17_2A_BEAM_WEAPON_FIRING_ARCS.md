# Planned Slice 17.2A - Beam Weapon Firing Arcs

Status: **active; Gates 1-2 complete; Gate 3 implementation ready / not started**.

## Objective

Add the original Beam Weapon Firing Arc design dimension that was not included in the accepted Slice-17.2 Beam contract, without reopening or rewriting the already accepted 17.2 scope.

Slice 17.2A is a bounded design-side follow-up. It owns authoritative Beam firing-arc identity, legality, Space/PP impact, Ship Designer selection, persistence and immutable handoff metadata. Complete Tactical enforcement of firing direction/target geometry remains downstream in Slice 17.6.

## Sequence / dependency

Binding Slice-17 child order around this follow-up:

`17.2 closed -> 17.3 Ordnance -> 17.2A Beam Firing Arcs -> 17.4 Specials/Fighters`

17.2A may open only after:

- Slice 17.2 remains closed/accepted;
- Slice 17.3 is closed/accepted;
- no other 17.x child is open;
- the existing Beam catalog / Weapon Selector baseline remains green.

## Known original evidence at planning time

Direct original executable evidence already establishes:

- Beam weapons use a real firing-arc field in the original design row;
- arc bit `0x10` maps to **360** in `Weapon_Arc_String_`;
- `Bonus_For_Weapon_Arcs_` returns **+50%** for `0x10`;
- Beam arc cost/space is therefore a design rule, not merely Tactical presentation;
- the accepted Slice-17.2 MOOX snapshot currently carries WeaponID, Count, MountType and ModifierIDs but no frozen Beam firing-arc field.

Gate 1 must re-audit the complete original arc matrix before implementation. Do not extrapolate the names/meaning of the other arc bits from `0x10` alone.

## Ownership boundary

### Slice 17.2A owns

- complete supported Beam firing-arc identity matrix;
- canonical Beam arc IDs and labels;
- original default arc;
- original per-arc Space/PP surcharge;
- Beam arc legality / compatibility;
- authoritative server quote integration;
- Ship Designer / Weapon Selector arc choice;
- saved ShipDesignSpec arc metadata;
- immutable built-Ship and BattleSession arc metadata;
- reference-lab and browser coverage for arc choices.

### Slice 17.2A does not own

- Missile/Torpedo/Bomb arc policy, which remains part of Slice 17.3;
- broad Tactical combatant admission, owned by Slice 17.5;
- Tactical Beam hit/range/damage or target-direction enforcement, owned by Slice 17.6;
- projectile/fighter mechanics.

## Gate 1 - original firing-arc audit

- [x] Enumerate every original Beam firing-arc bit/value and player-facing label.
- [x] Prove the original default Beam arc used by normal add/design flows.
- [x] Re-audit `Weapon_Arc_String_`, `Bonus_For_Weapon_Arcs_`, `Get_Weapon_Arc_` and Beam add/edit paths.
- [x] Freeze which arcs are legal for which Beam rows/mounts, if compatibility restrictions exist.
- [x] Freeze exact Space/PP bonus arithmetic and ordering relative to mount/modifier/miniaturization.
- [x] Audit the current MOOX Core/Ruleset/API/Weapon Selector/BattleSession gaps.
- [x] Define Baseline/Mid/All-Tech reference vectors including at least one 360-degree Beam row.

Evidence: `docs/research/SLICE_17_2A_GATE1_BEAM_FIRING_ARC_AUDIT_2026-09-25.md`.

## Gate 2 - Beam firing-arc contract freeze

- [x] Freeze canonical arc IDs/order/labels.
- [x] Freeze default arc.
- [x] Freeze per-arc Space/PP deltas.
- [x] Freeze compatibility and stable rejection reasons.
- [x] Freeze exact server quote ordering with Standard/Heavy/Point Defense, modifiers and miniaturization.
- [x] Freeze the persistent ShipWeaponMount arc field and immutable BattleSession handoff field.
- [x] Freeze the downstream Slice-17.6 Tactical support/defer metadata.
- [x] Freeze reference-lab and browser expected vectors.

Evidence: docs/research/SLICE_17_2A_GATE2_BEAM_FIRING_ARC_CONTRACT_2026-09-25.md.

## Gate 3 - implementation

- [ ] Normalize the frozen Beam firing-arc catalog/rules.
- [ ] Extend authoritative Beam quote/validation with firing arcs.
- [ ] Persist arc metadata through save/revision/construction/built Ship/BattleSession.
- [ ] Add a compact Weapon Selector arc control consistent with the accepted 17.2 HMI.
- [ ] Make base/configured Space/PP previews update from the server-owned rule path.
- [ ] Preserve Count, MountType and ModifierIDs when editing only the arc.
- [ ] Add Go/ruleset/API/persistence/browser regressions across representative Beam rows and mobile widths.
- [ ] Keep Tactical arc enforcement explicitly deferred to 17.6 until its combat contract is implemented.

## Gate 4 - acceptance

- [ ] Original-supported Beam arcs can be selected without reopening 17.2.
- [ ] 360-degree Beam configuration applies the verified original design surcharge.
- [ ] Save -> reload -> revise -> build -> BattleSession preserves the chosen arc exactly.
- [ ] Weapon Selector remains compact/non-overlapping on desktop and mobile.
- [ ] No browser-local pricing/legality formula diverges from server authority.
- [ ] Tactical support status is honest until Slice 17.6 consumes the arc metadata.

## Exit criterion

Slice 17.2A closes when supported Beam firing arcs are original-evidenced, server-authoritative, selectable, costed, persisted and handed off losslessly, with Tactical direction enforcement explicitly owned by Slice 17.6.

On accepted close, the next child is **Slice 17.4 - Ship Specials and Fighter Systems**.

