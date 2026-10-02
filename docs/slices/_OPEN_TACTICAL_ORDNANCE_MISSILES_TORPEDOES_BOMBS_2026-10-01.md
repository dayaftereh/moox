# OPEN Slice 17.7 - Tactical Missiles / Torpedoes / Bombs
Date opened: 2026-10-01
Status: **Gates 1-2 complete; Gate 3 implementation ready / not started**
Plan: `docs/slices/PLANNED_17_7_TACTICAL_ORDNANCE_MISSILES_TORPEDOES_BOMBS.md`
Gate 1 evidence: `docs/research/SLICE_17_7_GATE1_ORDNANCE_TACTICAL_AUDIT_2026-10-01.md`
Gate 2 contract: `docs/research/SLICE_17_7_GATE2_ORDNANCE_CONTRACT_FREEZE_2026-10-02.md`

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

Gate 1 is complete in `docs/research/SLICE_17_7_GATE1_ORDNANCE_TACTICAL_AUDIT_2026-10-01.md`. Gate 2 is frozen/complete in `docs/research/SLICE_17_7_GATE2_ORDNANCE_CONTRACT_FREEZE_2026-10-02.md`; Gate 3 implementation is ready but not started.

Current evidenced progress:

- original Missile/Torpedo projectile lifecycle symbols and callgraph mapped;
- Missile speed and projectile durability paths revalidated;
- Beam-vs-missile PD handoff tied directly to `Missile_Dcv_` + `Apply_Damage_To_Missile_`;
- Anti-Missile-Rockets separate interception path identified;
- Bombs proven to use planet/colony bombardment ownership rather than ordinary ship-target Tactical actions;
- current MOOX metadata-only/runtime gap verified by focused regressions;
- draft deterministic vectors A-E recorded.

Gate-1 closure now directly evidences launch multiplicity and live rack/cooldown state, movement budgets, interceptor DCV, impact Missile Defense/scanner/ECCM ordering, MIRV split timing, Plasma dissipation, reactive Beam-PD placement, relevant interceptor eligibility and bombardment ownership. Active combat-Special effects remain explicit Slice-17.8 hooks.

Gate 2 contract freeze is complete. No Gate-3 production implementation has started.

## Gate 2 progress

Gate 2 is **FROZEN / COMPLETE**. The contract freezes executable Missile/Torpedo support, bombardment-only Bomb/Bio ownership, live ammo/cooldown/projectile state, deterministic movement/impact/interception ordering, manual/reactive Beam interception, event/persistence/projection schemas, Slice-17.8 neutral hooks and deterministic vectors A-G.

Gate 3 is ready and remains not started.
