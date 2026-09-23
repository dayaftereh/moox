# OPEN - Slice 17.3 Missiles, Torpedoes and Bomb Ordnance

Opened: 2026-09-23
Parent: Slice 17 - Military Ship Design + Tactical Combat Program
Status: **Gate 1 active**

## Objective

Expand the accepted Slice-17.2 Ship Designer from Beam-only weapon breadth into an original-evidenced ordnance design catalog covering supported missiles, torpedoes and bombs with authoritative technology, damage, Space, Cost, miniaturization, mount/count, ammunition and legal upgrade/modifier rules.

## Gate 1 active scope

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

Gate 1 is evidence/audit work. Do not implement the broad ordnance matrix, ammo/upgrades or new Tactical projectile mechanics before Gate 2 freezes the supported contract.

Do not open Slice 17.4 or later children while Slice 17.3 is open unless explicitly requested.

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
