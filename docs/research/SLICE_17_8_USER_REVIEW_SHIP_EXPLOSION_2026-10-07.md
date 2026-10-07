# Slice 17.8 User Review - Ship Destruction Explosion

Date: 2026-10-07
Status: IN PROGRESS
Owner: Slice 17.8 manual review

## Trigger

Manual Combat Review found that destroyed ships currently transition directly to the destroyed visual state without a dedicated ship-destruction explosion presentation.

## Verified original contract

The original tactical path has a shared ship-explosion consumer:

- ordinary explosion power is 5 * hull_explosion_value;
- Quantum Detonator multiplies ordinary ship-explosion power by 3;
- Destroy_Ship_FX_ feeds that value into the shared spherical ship-explosion damage/effect path;
- Quantum Detonator capture self-destruction remains a separate 60 percent branch before owner transfer.

## Current MOOX gap

- Tactical runtime already emits authoritative ship_destroyed events from Beam, ordnance, Fighter and Boarding destruction paths.
- Tactical UI currently has no dedicated ship_destroyed explosion animation; destroyed ships are only rendered dimmed/grayscale.
- No shared ordinary ship-explosion damage consumer has yet been identified in production code.
- QuantumExplosionMultiplier = 3 already exists in the frozen 17.8 rules and must be consumed by the shared explosion path rather than implemented as visual-only behavior.

## Review-fix scope

1. Re-audit the original hull explosion value source and spherical damage targeting/range semantics before inventing any values.
2. Implement one shared authoritative ship-explosion resolution path used by normal destruction and Quantum Detonator death.
3. Preserve existing ship_destroyed ordering and deterministic replay; add explicit explosion event data only where needed for authoritative damage/presentation.
4. Apply any secondary explosion damage through existing Tactical damage semantics and handle chained ship destruction deterministically.
5. Add event-driven ship explosion VFX and combat-camera focus; the browser must visualize server outcomes only and must not simulate explosion damage.
6. Cover Beam, ordnance, Fighter and Quantum-detonator destruction with focused regressions.
7. Run full Go/Web QA, commit, then rebuild/reset canonical 7171 for manual review.

## Guardrails

- No client-side damage simulation.
- Do not weaken the existing ship_destroyed event contract.
- Do not conflate Quantum Detonator's capture-time 60 percent self-destruct branch with its x3 death-explosion multiplier.
- Do not open Slice 17.9 until Slice 17.8 manual review is explicitly accepted.
- No push unless explicitly requested.
