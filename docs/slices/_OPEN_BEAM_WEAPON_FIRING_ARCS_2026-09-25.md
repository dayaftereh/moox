# OPEN Slice 17.2A - Beam Weapon Firing Arcs

Date opened: 2026-09-25

Status: **Gates 1-3 complete - Gate 4 acceptance ready / not started**

Plan: `docs/slices/PLANNED_17_2A_BEAM_WEAPON_FIRING_ARCS.md`

## Scope

This bounded follow-up keeps accepted Slice 17.2 closed while adding the original Beam firing-arc design dimension.

Gate 1 must establish durable original evidence for:
- every original Beam firing-arc bit/value and player-facing label;
- the default Beam arc used by add/design flows;
- `Weapon_Arc_String_`, `Bonus_For_Weapon_Arcs_`, `Get_Weapon_Arc_` and Beam add/edit paths;
- Beam arc legality/compatibility;
- exact Space/PP surcharge arithmetic and ordering;
- current MOOX Core/Ruleset/API/Weapon Selector/BattleSession gaps;
- Baseline/Mid/All-Tech reference vectors including a 360-degree Beam.

## Guardrails

- Slice 17.2 remains closed/accepted and is not reopened.
- Slice 17.3 remains closed/accepted.
- Gate 3 must implement only the frozen Gate-2 Beam arc contract; do not add Slice-17.6 Tactical directional enforcement.
- Do not open Slice 17.4 or later children while 17.2A is open.
- Tactical firing-direction/target-geometry enforcement remains owned by Slice 17.6.
- Missile/Torpedo/Bomb arc policy remains owned by closed Slice 17.3.

## Gate 1 result

Gate 1 is complete. Durable evidence: `docs/research/SLICE_17_2A_GATE1_BEAM_FIRING_ARC_AUDIT_2026-09-25.md`.

Gate 2 contract freeze is complete. Gate 3 implementation is complete with durable evidence at docs/research/SLICE_17_2A_GATE3_BEAM_FIRING_ARC_IMPLEMENTATION_2026-09-25.md. Gate 4 acceptance is ready / not started; Slice 17.6 retains Tactical directional enforcement.
