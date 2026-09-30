# OPEN Slice 17.6 - Tactical Beam Weapons & Beam Modifiers
Date opened: 2026-09-30
Status: **Gate 1 original Beam-effect audit active**
Plan: docs/slices/PLANNED_17_6_TACTICAL_BEAM_WEAPONS_MODIFIERS.md

## Scope

Slice 17.6 turns the authoritative Beam identity/mount/modifier/firing-arc snapshots from closed Slices 17.2 and 17.2A into complete supported server-authoritative Tactical Beam combat on top of the closed Slice-17.5 Tactical foundation.

Gate 1 must establish direct/frozen original evidence for:

- every supported Beam identity's Tactical range, hit and damage behavior;
- Standard, Heavy and Point Defense mount semantics and incompatibilities;
- every supported 17.2 Beam modifier and interaction order;
- hit-roll inputs, distance/range effects, damage rolls, shot counts, defense interaction and exact rounding/order;
- Beam-side Point Defense hooks required by later Slice 17.7 projectile interception;
- one server-owned configured/effective Weapon Selector preview path;
- current Beam fire/damage event and Laser animation/readiness behavior;
- deterministic early/mid/late reference vectors.

## Guardrails

- Slices 17.0-17.5 and follow-up 17.2A remain closed/accepted and are not reopened.
- Gate 1 is audit/evidence only; do not implement full Beam behavior before Gate 2 freezes the Tactical Beam contract.
- BattleSession remains sole Tactical gameplay authority; React must not gain hit, range, damage or readiness formulas.
- Slice 17.7 retains missile/torpedo/bomb flight and end-to-end projectile interception.
- Slice 17.8 retains fighter/bomber/combat-Special behavior.
- Exactly one Slice-17 child is open.

## Gate 1 progress

Opening checkpoint only. Direct original-evidence audit is ready to begin; no 17.6 production implementation has started.
