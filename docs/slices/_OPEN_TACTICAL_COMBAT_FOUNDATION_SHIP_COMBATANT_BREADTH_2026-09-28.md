# OPEN Slice 17.5 - Tactical Combat Foundation / Ship Combatant Breadth
Date opened: 2026-09-28
Status: **Gates 1-2 complete; Gate 3 Checkpoints 1-2 complete; Checkpoint 3 ready / not started**
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

## Gate 1 progress

Initial audit: docs/research/SLICE_17_5_GATE1_TACTICAL_FOUNDATION_AUDIT_2026-09-28.md

Mapped: duplicate Game/Battle Slice-15.5 admission guards, singleton Frigate/Electronic durability/offense substitutions, Nuclear/Fusion-only speed rules, missing Shield/Fuel Tactical snapshot fields, Armor-to-Structure-only damage layering, and the existing BattleSession/RNG/event foundations that can be preserved.

Gate 1 complete: the local 1.31 executable now directly closes all six hull durability rows, all Armor multipliers, all Computer Beam-Offense rows, all Drive min/max matrices, primary Shield block/capacity/facing/recharge semantics, and Fuel tactical neutrality. Current BattleSession RNG/readiness/event authority and the generic unsupported-action replacement surface are audited. Early/mid/late representative vectors cover all six hull classes.

Next: Gate 3 implementation of the frozen foundation. The protected baseline admission guard remains in place until the replacement projection/validation is implemented and regression-green.
## Gate 2 contract freeze

Contract: docs/research/SLICE_17_5_GATE2_TACTICAL_FOUNDATION_CONTRACT_FREEZE_2026-09-29.md

Frozen: all six Hull durability rows; all six Armor multipliers; all five Computer Beam-Offense rows; all six Drive min/max matrices and the accepted concrete-design speed formula; no-Shield plus Class I/III/V/VII/X four-facing Shield capacity/block/recharge semantics; Fuel tactical neutrality; base Beam Defense from current combat speed; replacement semantics for both Slice-15.5 admission guards; generic per-slot action availability with explicit unsupported-action projection; shared RNG/damage/event/presentation authority; deterministic early/mid/late Combat Lab vectors; and compatibility boundaries for Slices 17.6-17.8.

Gate 3 may now implement the frozen replacement. It must not add complete Beam, ordnance or fighter/combat-Special family behavior owned by Slices 17.6-17.8.

## Gate 3 checkpoint execution

Plan: docs/research/SLICE_17_5_GATE3_IMPLEMENTATION_CHECKPOINT_PLAN_2026-09-29.md

Gate 3 is split into five recoverable checkpoints:

1. Tactical Ruleset and Ship Projection.
2. Broad Combatant Admission.
3. Shields and Shared Damage Foundation.
4. Generic Weapon Action Availability and UI Consumption.
5. Integration, Regression and Gate 3 Evidence.

After each checkpoint: test, diff-check, dedicated commit, clean-tree verification, durable session checkpoint, report to the user, then stop until explicit continuation.

Current resume point: **Checkpoint 3 ready / not started**. Checkpoints 1-2 are complete; verify clean HEAD before continuation.
