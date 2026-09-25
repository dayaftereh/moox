# Slice 17.2A Gate 2 - Beam Firing-Arc Contract Freeze
Date: 2026-09-25
Status: **Gate 2 complete; Gate 3 implementation ready / not started**
Parent: Slice 17 - Military Ship Design + Tactical Combat
Follow-up: Slice 17.2A - Beam Weapon Firing Arcs
## 1. Purpose and authority
This document freezes the MOOX Beam firing-arc contract before product implementation.
Evidence authority is the completed Gate-1 audit:
docs/research/SLICE_17_2A_GATE1_BEAM_FIRING_ARC_AUDIT_2026-09-25.md
Slice 17.2 remains closed/accepted. Slice 17.3 remains closed/accepted. This contract does not reopen either slice and does not authorize Tactical target-direction enforcement; that remains owned by Slice 17.6.
## 2. Canonical firing-arc identity
core.ShipWeaponFiringArcID remains the persistent/API type.
| ID | Original bit | EN label | DE label | Original short | Surcharge |
| --- | ---: | --- | --- | --- | ---: |
| forward | 0x01 | Forward | Vorne | F | 0% |
| forward_extended | 0x02 | Forward Extended | Vorne erweitert | Fx | +25% |
| back_extended | 0x04 | Back Extended | Hinten erweitert | Bx | +25% |
| rear | 0x08 | Rear | Hinten | B | 0% |
| 360 | 0x10 | 360 | 360┬░ | 360 | +50% |
Canonical selector order:
forward, forward_extended, back_extended, rear, 360.
The persisted design value is one choice, not a combinable bitmask.
The existing ShipWeaponFiringArc360 = "360" identity is retained unchanged.
## 3. Beam default and normalization
For a newly added Beam weapon, the authoritative default is forward.
Contract rules:
- Beam quote/preview may accept an empty FiringArcID only as legacy/add-flow shorthand and canonicalize it to forward.
- A successful authoritative quote returns and persists the canonical non-empty arc ID.
- Existing saved Beam rows with a valid non-empty arc preserve that arc on edit.
- Gate 3 must not silently map an unknown non-empty arc to Forward.
- Fixed-360 Ordnance behavior from Slice 17.3 is unchanged.
## 4. Compatibility and validation
Every supported Beam weapon may use all five Beam arcs whenever its selected Beam mount is itself legal.
There is no additional Beam-ID x arc or mount x arc compatibility matrix.
Stable Beam firing-arc rejection reasons:
- weapon_firing_arc_unknown - supplied non-empty arc ID is not one of the five canonical IDs.
- weapon_firing_arc_not_allowed - reserved for a known arc that is not permitted for the selected weapon family/rule. Beam rows should not hit this once Gate 3 implements this contract; fixed-360 Ordnance continues to enforce its own family rule.
Existing mount and modifier rejection reasons remain authoritative and are not renamed by 17.2A.
## 5. Space / PP surcharge and quote ordering
Arc surcharge percentages:
- forward: 0
- forward_extended: 25
- back_extended: 25
- rear: 0
- 360: 50
For Beam Space and PP independently, the frozen ordering is:
1. accepted Slice-17.2 base weapon value;
2. accepted Slice-17.2 mount factor;
3. one additive configuration factor containing 100 + selected weapon-modifier percentage delta + Beam firing-arc surcharge percentage;
4. accepted Slice-17.2 miniaturization calculation;
5. accepted Slice-17.2 Count/minimum semantics.
Frozen formula:
configured = floor(after_mount * (100 + weapon_modifier_bonus + arc_bonus) / 100)
The arc surcharge is not applied as a later independent multiplier.
17.2A does not reopen Count/grouped-row rounding behavior. Gate-3 tests use Count=1 reference vectors for arc parity plus regression coverage proving existing Count behavior remains unchanged.
## 6. Persistence and immutable handoff
core.ShipWeaponMount.FiringArcID is the sole persisted Beam arc field.
A successful Beam add/edit must carry its canonical FiringArcID through:
ShipDesignSpec -> design revision/save -> construction -> built Ship snapshot -> BattleSession TacticalWeaponSpec/metadata.
No layer may infer or recalculate the Beam arc from UI state after authoritative quote/save.
Built-Ship and BattleSession snapshots are immutable with respect to later ruleset/catalog changes.
## 7. Server catalog / preview contract
Beam weapon catalog rows gain server-owned selectable arc metadata:
- firing_arc_choices: ordered list of canonical arc choices;
- default_firing_arc_id: forward.
Each arc choice exposes:
- id;
- label_en;
- label_de;
- space_cost_percent;
- production_cost_percent.
For the frozen original rule, Space and PP percentages are identical per arc.
fixed_firing_arc_id remains reserved for weapon families such as Slice-17.3 Ordnance that have no player-selectable arc.
Beam preview/quote responses return the canonical selected firing_arc_id in the quoted mount.
The web client must render Beam choices from server catalog metadata and must not embed its own legality or surcharge table.
## 8. Ship Designer / HMI contract
The Weapon Selector exposes one compact Beam firing-arc selector for Beam rows only.
Behavior:
- initial Beam add shows Forward selected;
- all five choices are available for every otherwise legal Beam row;
- changing arc requotes through the authoritative server path;
- preview Space/PP updates from server output;
- edit reopens with the persisted arc selected;
- fixed-360 Ordnance retains the existing read-only behavior.
Labels are bilingual as frozen in section 2. Original F/Fx/B/Bx/360 may be used as compact secondary notation, but canonical API IDs remain the strings frozen above.
## 9. Tactical boundary / Slice 17.6 deferral
17.2A owns only design-side arc identity, quote, persistence and immutable metadata handoff.
BattleSession must carry the selected Beam FiringArcID, but 17.2A does not implement:
- target-bearing geometry;
- ship facing/rotation checks;
- fire eligibility by direction;
- Tactical arc rendering;
- Tactical rejection because a target lies outside an arc.
Those behaviors remain Slice 17.6 - Tactical Beam.
Until Slice 17.6, Beam Tactical support remains metadata-only with respect to directional arc enforcement. No fake 360 fallback is permitted.
## 10. Frozen reference vectors
### A. Baseline - Laser Cannon
Count 1, Standard, no modifiers.
Forward:
- Space 10
- PP 5
360:
- Space 15
- PP 7
### B. Mid-Tech - Neutron Blaster
Count 1, Standard, no modifiers, miniaturization level 1.
Forward/current base path:
- Space 8
- PP 6
360:
- configured base 15/12 before miniaturization
- expected Space 12
- expected PP 9
### C. All-Tech - Phasor
Count 1, Point Defense, Shield Piercing + Autofire, miniaturization level 4.
Without arc surcharge:
- expected Space 5
- expected PP 3
With 360:
- mounted base 5/5
- second factor 100 + 100 + 50 = 250%
- configured pre-mini 12/12
- expected Space 6
- expected PP 3
These vectors are binding Gate-3 regression and browser/reference-lab expectations.
## 11. Gate-3 acceptance surface
Gate 3 must implement this frozen contract across:
- Core constants/types;
- authoritative Beam quote/validation;
- ruleset/catalog projection;
- preview/API;
- Weapon Selector;
- save/revision;
- construction;
- built Ship;
- BattleSession metadata;
- focused unit/integration tests;
- Baseline/Mid/All-Tech reference-lab/browser checks.
Gate 3 must not add Tactical directional enforcement.
## 12. Gate-2 checklist
- [x] Freeze canonical arc IDs/order/labels.
- [x] Freeze default arc.
- [x] Freeze per-arc Space/PP deltas.
- [x] Freeze compatibility and stable rejection reasons.
- [x] Freeze exact server quote ordering with Standard/Heavy/Point Defense, modifiers and miniaturization.
- [x] Freeze the persistent ShipWeaponMount arc field and immutable BattleSession handoff field.
- [x] Freeze the downstream Slice-17.6 Tactical support/defer metadata.
- [x] Freeze reference-lab and browser expected vectors.
Gate 2 is complete. Gate 3 implementation is ready / not started.
