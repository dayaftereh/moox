# OPEN - Slice 17.3 Missiles, Torpedoes and Bomb Ordnance

Opened: 2026-09-23
Parent: Slice 17 - Military Ship Design + Tactical Combat Program
Status: **Gate 2 complete - Gate 3 implementation/review active**

## Objective

Expand the accepted Slice-17.2 Ship Designer from Beam-only weapon breadth into an original-evidenced ordnance design catalog covering supported missiles, torpedoes and bombs with authoritative technology, damage, Space, Cost, miniaturization, mount/count, ammunition and legal upgrade/modifier rules.

## Gate 1 completed scope

- inventory normalized missile, torpedo and bomb identities required by the supported design catalog;
- map each supported ordnance row to technology unlock, base damage, Space, Cost and any range/speed/guidance metadata required by the design contract;
- re-audit ammunition / shot-count representation and whether counts live per mount, per salvo or elsewhere in the original model;
- re-audit original miniaturization behavior for ordnance Space/Cost;
- audit current `ShipWeaponMount` / Core snapshot shape for lossless missile/torpedo/bomb representation;
- audit legal ordnance modifiers/upgrades and their compatibility restrictions from direct/original evidence;
- separate design-time metadata from projectile flight, interception, ECM/evasion, impact/explosion and bomb-vs-planet runtime state;
- define Baseline/Mid/All-Tech reference vectors in the durable Slice-17.0 labs, including representative missile, torpedo and bomb-carrying designs;
- preserve construction/persistence/revision snapshot authority from closed Slices 17.1-17.2;
- record Slice 17.7 as the complete downstream Tactical ordnance owner.

## HMI boundary

Reuse the accepted compact weapon-selection grammar from Slice 17.2 where it fits, but Gate 1 must first determine what ordnance-specific information is required.

Do not invent missile/torpedo/bomb UI controls for ammo, salvos, guidance or upgrades until Gate 2 freezes the authoritative contract.

## Tactical boundary

Slice 17.3 answers **what ordnance is installed on the ship**.

Complete Tactical ordnance behavior is explicitly downstream in planned **Slice 17.7**, including:

- projectile flight/state;
- missile/torpedo interception;
- Point Defense engagement;
- ECM/evasion resolution;
- impact/explosion timing;
- verified bomb combat/bombardment context.

Gate 1 may inspect current Tactical consumers and handoff fields, but it must not silently implement or approximate those deferred systems.

## Guardrail

Slices 17.0, 17.1 and 17.2 are closed/accepted and remain the foundation.

Gate 1 completed as evidence/audit work. Do not implement the broad ordnance matrix, ammo/upgrades or new Tactical projectile mechanics before Gate 2 freezes the supported contract.

Do not open planned follow-up Slice 17.2A, Slice 17.4 or later children while Slice 17.3 is open unless explicitly requested. On accepted 17.3 close, 17.2A is the next child before 17.4.

## Gate 1 completion target

Gate 1 may close only when the repository contains durable evidence resolving:

- supported missile / torpedo / bomb identity matrix;
- technology ownership;
- damage and any required design-time range/speed/guidance metadata;
- base Space/Cost;
- miniaturization evidence;
- mount/count/ammunition representation;
- legal upgrade/modifier compatibility;
- immutable ShipDesignSpec/BattleSession handoff fields;
- current Tactical consumers versus explicit Slice-17.7 deferrals;
- reference-lab vectors needed for Gate 2/3.

Gate 1 result: **complete**. Durable evidence: `docs/research/SLICE_17_3_GATE1_ORDNANCE_AUDIT_2026-09-23.md`.

## Gate 2 completed scope

Working contract: `docs/research/SLICE_17_3_GATE2_ORDNANCE_CONTRACT_FREEZE_2026-09-23.md`.

Already frozen in Gate 2:

- exact 13-row supported ordnance matrix;
- ordnance Standard-only mount semantics;
- fixed 360-degree Firing Arc behavior with no new arc selector;
- selectable Missile Ammo/Rack sizes `2/5/10/15/20`;
- ordnance modifier IDs/order/legality/refinement/design deltas;
- Missile rack -> modifiers -> miniaturization -> Count quote order;
- Slice-17.7 Tactical ownership and Slice-17.0 reference vectors.

Gate 2 result: **complete**.

Final frozen representation/API decisions:

- canonical snapshots persist Ammo for all ordnance: selectable Missile rack, fixed Torpedo `2`, fixed Bomb/Bio `10`;
- canonical snapshots persist `FiringArcID = "360"` for all 17.3 ordnance;
- `ShipWeaponMount` and `TacticalWeaponSpec` both carry Ammo + FiringArcID;
- catalog/preview fields expose server-owned Ammo choices/defaults/fixed arc;
- stable Ammo/arc rejection reasons and validation order are frozen;
- Beam Ammo/arc behavior remains unchanged until planned 17.2A.

Gate 3 implementation now follows this frozen contract and is available for manual review on the canonical 7171 reference server. Gate 3 is not yet closed; acceptance/browser review and final closeout remain.
