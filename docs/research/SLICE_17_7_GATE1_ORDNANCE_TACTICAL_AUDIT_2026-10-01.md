# Slice 17.7 Gate 1 - Tactical Ordnance Audit
Date opened: 2026-10-01
Status: **ACTIVE - original/reference and current-runtime audit in progress**

## Objective

Re-audit the supported Slice-17.3 missile/torpedo/bomb matrix for server-authoritative Tactical execution before freezing any Slice-17.7 combat contract.

## Audit checklist

- [ ] Supported missile/torpedo/bomb identities and design-snapshot inputs.
- [ ] Ammo, salvo, readiness and launch semantics.
- [ ] Range, travel, guidance, target tracking and projectile lifecycle.
- [ ] Hit, damage, Shield/Armor/Structure application and RNG ordering.
- [ ] Beam Point Defense target eligibility/interception ordering and projectile-side outcome.
- [ ] Defensive countermeasures represented by current design snapshots.
- [ ] Bomb combat context and planetary-bombardment ownership.
- [ ] BattleSession projectile state, persistence, replay and Observer/player projection requirements.
- [ ] Tactical UI/event/VFX requirements without client-side simulation.
- [ ] Deterministic reference vectors and explicit unknowns/blockers.

## Gate boundary

This document is Gate-1 evidence only. No Tactical ordnance gameplay implementation is authorized until Gate 2 freezes the supported contract.
