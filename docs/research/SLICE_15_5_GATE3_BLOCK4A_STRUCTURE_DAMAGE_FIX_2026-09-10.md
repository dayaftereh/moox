# Slice 15.5 Gate 3 Block 4a - Armor overflow to aggregate Structure

Date: 2026-09-10

## Problem found by full browser battle QA

A real 2v2 browser battle could not complete after Armor was depleted. The retained Slice-07 safety boundary rejected either:

- Laser damage larger than the target's remaining Armor; or
- a post-Armor historical internal-selection roll other than 100.

This left the new Slice-15.5 interactive battle unable to destroy ships reliably.

## Frozen 15.5 baseline fix

Internal subsystem damage remains deferred. The supported first Tactical damage model is now:

1. apply as much beam damage as possible to current Armor;
2. apply any remaining damage to aggregate Structure;
3. clamp StructureDamage to StructureMax;
4. destroy the ship when StructureDamage reaches StructureMax.

The historical internal-selection RNG roll is still consumed in the same place for deterministic RNG compatibility, but it no longer causes an unsupported rejection in the aggregate-Structure baseline and does not invent subsystem damage.

Damage events distinguish `armor`, `structure`, and `armor_structure` when one hit crosses the Armor boundary.

## Regression coverage

Added/updated tests prove:

- a deterministic 4-damage Laser against Armor=1 produces Armor=0 and StructureDamage=3;
- that crossing hit records `armor_structure`;
- a target with Armor=0 can take aggregate Structure damage and be destroyed;
- a deliberately selected historical internal-selection roll !=100 no longer rejects and still advances aggregate Structure;
- existing golden deterministic battle expectations remain valid.

## Validation

Passed:

- `go test ./internal/battle ./internal/session ./internal/server`.

This fix is intentionally separate from the still-uncommitted Tactical ship-art polish so it can be checkpointed and session-rotated independently.
