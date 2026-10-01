# OPEN Slice 17.7 - Tactical Missiles / Torpedoes / Bombs
Date opened: 2026-10-01
Status: **Gate 1 ordnance Tactical audit active**
Plan: `docs/slices/PLANNED_17_7_TACTICAL_ORDNANCE_MISSILES_TORPEDOES_BOMBS.md`
Evidence: `docs/research/SLICE_17_7_GATE1_ORDNANCE_TACTICAL_AUDIT_2026-10-01.md`

## Scope

Slice 17.7 turns the authoritative missile/torpedo/bomb design snapshots from closed Slice 17.3 into supported server-authoritative Tactical ordnance behavior on the closed 17.5 Tactical foundation, integrating the frozen Slice-17.6 Beam Point Defense reaction contract where projectile interception applies.

Gate 1 must establish direct/frozen evidence for:

- every supported missile, torpedo and bomb identity and its correct Tactical or bombardment context;
- ammo, salvo, readiness, range, travel, guidance, hit and damage semantics;
- authoritative projectile lifecycle/state and deterministic movement/ETA requirements;
- Point Defense/interceptor target eligibility, ordering and RNG interaction with Slice 17.6;
- defensive missile/torpedo countermeasures represented by current snapshots;
- bomb ownership: ship Tactical versus planetary bombardment or another explicit path;
- BattleSession save/load/replay requirements for active projectile state;
- event-driven projectile, interception, impact and explosion presentation;
- deterministic original/reference vectors for representative ordnance cases.

## Guardrails

- Slices 17.0-17.6 and follow-up 17.2A remain closed/accepted and are not reopened.
- Gate 1 is audit/evidence only; do not implement Tactical ordnance behavior before Gate 2 freezes the combat contract.
- BattleSession remains sole Tactical gameplay authority; React does not simulate projectile movement, interception, hit or damage.
- Slice 17.6 Beam Point Defense is a frozen dependency; Slice 17.7 owns projectile-side eligibility, timing and outcome.
- Slice 17.8 retains fighter/bomber/combat-Special behavior.
- Exactly one Slice-17 child is open.

## Gate 1 progress

Gate 1 is active in `docs/research/SLICE_17_7_GATE1_ORDNANCE_TACTICAL_AUDIT_2026-10-01.md`. No Gate-2 contract or gameplay implementation has started.
