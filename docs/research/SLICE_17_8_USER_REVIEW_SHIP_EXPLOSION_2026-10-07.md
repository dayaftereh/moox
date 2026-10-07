# Slice 17.8 User Review - Ship Destruction Explosion

Date: 2026-10-07
Status: READY FOR MANUAL REVIEW
Owner: Slice 17.8 manual review

## Trigger

Manual Combat Review found that destroyed ships currently transition directly to the destroyed visual state without a dedicated ship-destruction explosion presentation.

## Verified original contract

Direct 1.31 executable audit now distinguishes normal destruction presentation from Quantum damage:

- Is_Combat_Ship_Dead_ returns death mode 1 for ordinary Structure destruction and overrides that mode to 4 when functional Quantum Detonator ID 24 is present.
- Normal death mode 1 goes through Play_Ship_Explosion_SFX_. That helper selects/plays the hull explosion SFX and does not call Do_Spherical_SFX_. Ordinary ship destruction is therefore an explosion presentation, not a generic spherical AoE-damage trigger.
- Quantum death mode 4 instead calls Damage_From_Ship_Exploding_ and then Do_Spherical_SFX_.
- Damage_From_Ship_Exploding_ reads the original hull-table explosion value at object-2 offset 0x801E + size_index*0x24 + 0x0C, multiplies by 5, then multiplies by 3 for Quantum Detonator.
- Original hull explosion values are Frigate 2, Destroyer 5, Cruiser 10, Battleship 15, Titan 20, Doom Star 40; Quantum powers are therefore 30, 75, 150, 225, 300 and 600.
- Quantum mode 4 loads SPHERSFX.LBX block 1 (270px wide), so Do_Spherical_SFX_ uses radius floor(270/2)=135 pixels. Combat coordinates use 20 pixels per tactical field.
- Eligible nearby ships are live, non-stasis ships other than the exploding source, independent of side, with squared pixel distance <= 135^2.
- Quantum raw blast at a target is max(1, power * (100 - floor(distance_squared * 100 / 135^2)) / 100). The original then converts this through target-hull-size stochastic blast rolls before normal Facing_Shield_ -> Apply_Ship_Damage_ handling.
- Quantum Detonator capture self-destruction remains a separate capture branch; once destruction occurs it uses the same mode-4 death explosion semantics.

## Current MOOX gap

- Tactical runtime already emits authoritative ship_destroyed events from Beam, ordnance, Fighter and Boarding destruction paths.
- Tactical UI currently has no dedicated ship_destroyed explosion animation; destroyed ships are only rendered dimmed/grayscale.
- No shared ordinary ship-explosion damage consumer has yet been identified in production code.
- QuantumExplosionMultiplier = 3 already exists in the frozen 17.8 rules and must be consumed by the shared explosion path rather than implemented as visual-only behavior.

## Review-fix scope

1. Preserve the verified normal-death behavior: every authoritative ship_destroyed gets explosion presentation, but ordinary mode-1 death does not invent spherical AoE damage.
2. Implement one shared authoritative Quantum death-explosion resolver using the recovered hull powers, 135px radius, radial falloff and target-size blast conversion.
3. Preserve existing ship_destroyed ordering and deterministic replay; add explicit explosion event data only where needed for authoritative damage/presentation.
4. Apply Quantum secondary explosion damage through existing Tactical damage semantics and handle chained ship destruction deterministically.
5. Add event-driven ship explosion VFX and combat-camera focus; the browser must visualize server outcomes only and must not simulate explosion damage.
6. Cover Beam, ordnance, Fighter and Quantum-detonator destruction with focused regressions.
7. Run full Go/Web QA, commit, then rebuild/reset canonical 7171 for manual review.

## Guardrails

- No client-side damage simulation.
- Do not weaken the existing ship_destroyed event contract.
- Do not conflate Quantum Detonator's capture-time 60 percent self-destruct branch with its x3 death-explosion multiplier.
- Do not open Slice 17.9 until Slice 17.8 manual review is explicitly accepted.
- No push unless explicitly requested.

## Implementation result

- All authoritative ship destruction paths now converge on one tacticalResolveShipDestruction resolver.
- Ordinary mode-1 destruction emits ship_destroyed with explosion_visual=true and no invented spherical AoE damage.
- Functional Quantum Detonator mode-4 destruction adds recovered hull power/radius metadata, resolves friendly/enemy splash damage server-side through the existing shield/armor/structure pipeline, emits ship_explosion_resolved, and handles deterministic chained destruction through the same resolver.
- Existing capture-time Quantum self-destruction now converges on the same death-explosion resolver after boarding_resolved.
- Browser presentation consumes authoritative ship_destroyed only: camera centers on the destroyed ship, normal deaths render a multi-stage core/halo/shockwave/debris effect, and Quantum deaths additionally render the authoritative 135px / 6.75-field spherical field.
- Browser code never computes explosion damage.
- Focused internal/battle tests cover original hull explosion values, ordinary visual-only destruction, Quantum power/radius/friendly fire, and squared-distance falloff.
- Full go test ./... -count=1, go vet ./..., web npm run build, Tactical ordnance browser contract and git diff --check passed before commit.
