# Slice 17.7 Gate 3 - Checkpoint 1 Implementation Evidence

Date: 2026-10-02
Status: COMPLETE

## Scope

Checkpoint 1 implements only the frozen Gate-2 foundation for Tactical ordnance:

- executable Missile/Torpedo rules and Tactical support identities;
- replay-stable ordnance rules snapshots;
- authoritative live ammo/cooldown/projectile state;
- deterministic initial state;
- live-snapshot persistence and legacy migration;
- observer/player/Tactical-view projection;
- deep-clone / alias isolation;
- rules and runtime validation.

The checkpoint intentionally does not implement launch actions, projectile movement, retargeting, impact, damage, ECCM RNG, MIRV resolution, Beam interception, reactive Point Defense, or browser projectile animation.

## Ruleset and identities

The moo2-1.31 Tactical ruleset now carries the frozen ordnance constants:

- movement units per field: 20
- Fast speed bonus: 4
- torpedo cooldown rounds: 2
- MIRV multiplier: 4
- Plasma dissipation per field: 5

Executable Missile/Torpedo weapons expose frozen projectile speed, durability and Beam interceptor DCV values. Missile/Torpedo Tactical support is ordnance; Bomb/Bio support is bombardment_only. Beam rules remain separate and reject projectile-only fields.

## Battle state and rules snapshot

The battle package now contains detached TacticalOrdnanceRulesSnapshot and TacticalOrdnanceState structures, including per-weapon rules, per-slot ammo/cooldown state, projectile DTO/state and deterministic NextProjectileID handling.

Initial state is deterministic:

- NextProjectileID starts at 1;
- Missile ammo comes from TacticalWeaponSpec.Ammo;
- Torpedo ammo state is 0 and readiness is represented by cooldown;
- cooldowns start at 0;
- weapon states are ordered by ShipID / WeaponSlot;
- no projectile exists before a later launch checkpoint.

Runtime and snapshot validation reject family/support/rule mismatches.

## Projection and alias safety

TacticalRulesSnapshot carries an optional Ordnance block for backward compatibility. Tactical runtime owns the authoritative ordnance state, while observer/player/Tactical views receive detached clones. MissileDefense and MissileScannerReduction are projected as neutral fields for the later Slice-17.8 consumer.

Checkpoint tests mutate projected/view/snapshot data and verify authoritative runtime state and sibling projections remain unchanged.

## Live snapshot and legacy migration

TacticalLiveSnapshot now carries an optional Ordnance state.

New battles write the block explicitly. Restore clones and validates a present block. Legacy snapshots without the block remain loadable only when they do not already contain executable ordnance support; their ordnance state is initialized deterministically from TacticalSpec. A snapshot that already declares TacticalSupport=ordnance but omits the authoritative ordnance state is rejected.

LiveSnapshotSchemaVersion remains 1, matching the additive Gate-2 compatibility contract.

## Regression coverage

Dedicated Checkpoint-1 tests cover:

- deterministic initialization and detached projection;
- live-snapshot round trip and legacy migration;
- support/family mismatch rejection.

Existing ruleset, military designer, Tactical combat and rules-identity baselines were updated for the intentional support-identity/ruleset change.

## QA

Passed on the final Checkpoint-1 working tree:

`gofmt` on changed Go production/test files.

`go test ./internal/ruleset ./internal/battle ./internal/game ./internal/session -count=1`

- internal/ruleset: PASS
- internal/battle: PASS
- internal/game: PASS
- internal/session: PASS

`go test ./... -count=1`

- PASS across the repository.

`go vet ./...`

- PASS.

`git diff --check`

- PASS; only the repository's known LF-to-CRLF warnings are emitted on Windows.

## Gate boundary

Checkpoint 1 is complete. Checkpoint 2 owns launch/readiness/movement and therefore remains the first place where battle.fire_ordnance, projectile spawn/movement, ammo consumption, torpedo cooldown progression and retarget ordering may be implemented.
