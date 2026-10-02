# Slice 17.7 Gate 3 - Checkpoint 2 Launch / Movement / Readiness Evidence

Date: 2026-10-02
Status: COMPLETE

## Scope

Checkpoint 2 implements the frozen Gate-2 launch, movement and readiness contract on top of the completed Checkpoint-1 rules/state/projection foundation.

Implemented here:

- `battle.fire_ordnance` command and payload;
- server-projected Missile/Torpedo action availability;
- explicit `ammo_exhausted`, `weapon_cooldown` and `ordnance_effect_unavailable` reasons;
- one-volley Missile ammo consumption;
- Torpedo cooldown `2 -> 1 -> 0`;
- normal-activation ordnance readiness;
- stable monotonic projectile IDs;
- grouped projectile spawn with installed weapon Count as multiplicity;
- immediate launch-round projectile movement quantum;
- one deterministic movement quantum at each later round transition;
- ascending-projectile-ID scheduling;
- deterministic ascending-Ship-ID retargeting;
- `no_target` projectile destruction;
- authoritative launch/move/retarget/destroy events;
- rejected-command atomicity and RNG neutrality.

Checkpoint 2 intentionally does not implement impact accuracy, impact RNG, MIRV expansion, Fast/Armored/ECCM/Overloaded/Enveloping/No-Range-Dissipation effects, ship damage, manual/reactive Beam interception, active Slice-17.8 Specials, or browser projectile presentation. Those remain owned by later Gate-3 checkpoints.

## Command and action projection

The BattleSession now accepts:

`CommandFireOrdnance = "battle.fire_ordnance"`

with `FireOrdnancePayload`:

- `ship_id`;
- `target_ship_id`;
- `weapon_slot`.

Only executable Missile/Torpedo rows with `TacticalSupport="ordnance"` project this command. Live opponent Ships are server-owned targets.

Availability distinguishes:

- `ammo_exhausted` for an empty Missile rack;
- `weapon_cooldown` for a Torpedo with positive cooldown;
- `ordnance_effect_unavailable` for Emissions Guidance;
- existing `weapon_not_ready`, `no_legal_targets` and `tactical_consumer_unavailable` where applicable.

The legacy `LegalFireActions` surface remains Beam-only in this checkpoint. Ordnance execution is exposed through `TacticalWeaponActionAvailability`; browser/UI integration remains Checkpoint 5.

## Launch state

A successful Missile launch:

- consumes exactly one live rack volley;
- does not multiply ammo consumption by installed Count;
- creates one grouped projectile actor;
- sets actor multiplicity to installed Count;
- consumes the weapon's normal activation readiness.

A successful Torpedo launch:

- requires zero live cooldown;
- creates one grouped actor with installed Count multiplicity;
- sets cooldown to the frozen value 2;
- consumes normal activation readiness.

Projectile IDs are allocated from `NextProjectileID` and remain monotonic/replay-stable.

Checkpoint 2 carries the frozen base actor values from Checkpoint 1: base speed, base durability, interceptor DCV identity and base warhead strength. Modifier execution is deliberately deferred to Checkpoint 3.

## Movement and retarget ordering

One Tactical grid field consumes 20 projectile movement units.

Each movement quantum resets the actor budget to:

`speed * movement_units_per_field`

and accumulates actual used movement units in `distance_traveled`.

A successful launch immediately processes one movement quantum. At each deterministic new-round transition:

1. existing round reset/recharge runs;
2. positive Torpedo cooldowns decrement and Missile/Torpedo readiness is derived;
3. active projectiles process in ascending `projectile_id`;
4. initiative is recomputed;
5. `round_started` is emitted.

Movement and targeting are server-authoritative and do not consume Tactical RNG in Checkpoint 2.

If a locked target is no longer valid, the actor selects the first live opponent by ascending Ship ID and emits `projectile_retargeted`. If no legal opponent remains, the actor is removed and emits `projectile_destroyed` with reason `no_target`.

## Events

Checkpoint 2 emits:

- `ordnance_launched`;
- `projectile_moved`;
- `projectile_retargeted`;
- `projectile_destroyed` for `no_target`.

Launch events include stable projectile/source/target/weapon identity, modifier IDs, starting pose, speed, multiplicity, durability, Beam interceptor DCV, warhead strength, and post-launch ammo or cooldown.

Movement events include from/to coordinates, movement used/remaining, accumulated distance and target identity.

## Regression coverage

Dedicated Checkpoint-2 tests cover:

- executable Missile/Torpedo action projection;
- Nuclear Missile one-volley launch, stable ID and grouped multiplicity;
- Anti-Matter Torpedo launch/cooldown;
- immediate launch movement;
- later round movement;
- cooldown `2 -> 1 -> 0`;
- explicit `ammo_exhausted`;
- successful launch/movement RNG neutrality;
- rejected-command atomicity;
- ascending-Ship-ID retarget;
- `no_target` destruction;
- Emissions Guidance remaining unavailable.

Historical Slice-17.3/17.5/17.6 tests that explicitly asserted "Ordnance deferred" were updated only where Checkpoint 2 intentionally supersedes that boundary. Beam behavior remains executable and Fighter/later-family behavior remains deferred.

## QA

Final Checkpoint-2 working tree:

`go test ./internal/battle -count=1`

- PASS.

`go test ./internal/ruleset ./internal/battle ./internal/game ./internal/session -count=1`

- PASS in all four packages.

`go test ./... -count=1`

- PASS across the repository.

`go vet ./...`

- PASS.

`git diff --check`

- PASS; only the repository's known LF-to-CRLF warnings are emitted on Windows.

A dedicated scope scan confirms no Checkpoint-3/4 behavior was introduced: no impact RNG, MIRV expansion, ordnance modifier execution, impact events or Beam interception resolution exists in the Checkpoint-2 implementation.

## Next checkpoint

Gate 3 Checkpoint 3 owns Impact / Modifiers:

- impact reach/collision handoff and accuracy;
- Missile Defense / scanner / ECCM ordering;
- MIRV split timing;
- Fast and Armored Missile effects;
- Torpedo Overloaded / Enveloping;
- Plasma dissipation / No Range Dissipation;
- shared authoritative ship damage.

Manual/reactive Beam projectile interception remains Checkpoint 4.
