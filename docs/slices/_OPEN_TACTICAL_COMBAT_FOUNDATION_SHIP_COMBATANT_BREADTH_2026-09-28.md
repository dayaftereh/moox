# OPEN Slice 17.5 - Tactical Combat Foundation / Ship Combatant Breadth
Date opened: 2026-09-28
Status: **Gate 1 foundation audit active**
Plan: docs/slices/PLANNED_17_5_TACTICAL_COMBAT_FOUNDATION_SHIP_COMBATANT_BREADTH.md

## Scope

Slice 17.5 replaces the remaining Slice-15.5 baseline-only Tactical combatant boundary with a broad, server-authoritative Tactical foundation that can admit the accepted Slice-17.1 hull/mandatory-system matrix without hidden Frigate/Electronic/Titanium substitutions.

Gate 1 must establish durable evidence for:

- the exact Slice-15.5 Tactical admission guard and every hard-coded baseline value it protects;
- original hull durability semantics for Frigate through Doom Star;
- Armor, Computer, Shield, Drive and Fuel-Cell Tactical effects;
- current BattleSession RNG, damage-layer, slot/action readiness and event authority;
- current Tactical UI assumptions that are Laser-specific versus already generic;
- representative early/mid/late combatant vectors across all six hull classes;
- explicit per-weapon unsupported-action behavior that does not reject an otherwise legal encounter.

## Guardrails

- Slices 17.0-17.4 and follow-up 17.2A remain closed/accepted and are not reopened.
- Gate 1 is audit/evidence only; do not remove the protected baseline admission guard before Gate 2 freezes its replacement contract.
- Do not implement complete Beam, ordnance or fighter/special combat rules here; those remain owned by Slices 17.6, 17.7 and 17.8.
- BattleSession remains the sole Tactical authority; browser presentation must not invent hit, damage, readiness or legal-fire rules.
- Unsupported later-family weapons must become explicit per-action unavailability, not whole-encounter rejection.
- Only one Slice-17 child is open.
