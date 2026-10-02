# Slice 17.7 Gate 3 - Checkpoint 4 Interception Evidence

Date: 2026-10-02
Status: COMPLETE

## Scope

Checkpoint 4 implements the frozen Gate-2 Beam interception contract for active Missile/Torpedo projectile actors on top of the completed Checkpoints 1-3.

Implemented here:

- additive manual Beam projectile targeting through `target_projectile_id`;
- exactly-one ship/projectile Beam target validation;
- Standard Beam and Point Defense Beam manual projectile targets;
- explicit Heavy Beam exclusion for projectile targets;
- projectile interceptor DCV derived from the frozen ordnance weapon rule;
- existing Slice-17.6 Beam hit/range/damage derivation reused for manual projectile fire;
- normal manual Beam readiness consumption;
- deterministic projectile multiplicity / durability / interception-remainder mutation;
- authoritative `projectile_interception_resolved` events;
- authoritative `projectile_destroyed { reason: "intercepted" }` lifecycle;
- reactive pre-impact Point Defense using the frozen Slice-17.6 resolver;
- ascending reactive Point Defense slot ordering;
- reactive Point Defense normal-readiness preservation;
- reactive interception before MIRV split, Plasma dissipation and ordinary impact;
- neutral Slice-17.8 interceptor/Special hook boundary retained.

Checkpoint 4 intentionally does not implement Anti-Missile Rockets, Lightning Field, Dauntless Guidance, Fast Missile Racks, fighter/bomber interception, or any other active Slice-17.8 Special semantics.

## Manual Beam target transport

The existing Beam target transport is extended additively.

`TacticalFireTarget` now exposes:

- optional `target_ship_id`;
- optional `target_projectile_id`;
- existing `range_index`.

`FireBeamPayload` now accepts the same optional projectile identity.

The server rejects a Beam command unless exactly one target identity is non-zero. Existing ship-target Beam commands therefore remain unchanged, while projectile-target commands use the same command kind and weapon slot transport.

The legal-fire projection adds active opponent projectile actors to supported Beam mounts only when the authoritative projectile coordinate is inside the weapon firing arc and produces a valid frozen Beam range index.

Manual projectile-target eligibility is:

- Standard Beam: legal;
- Point Defense Beam: legal;
- Heavy Beam: illegal.

Heavy Beam mounts remain fully usable against normal ship targets; only projectile actors are excluded.

## Interceptor DCV

The frozen `TacticalProjectileState` does not duplicate interceptor DCV. Checkpoint 4 therefore derives the target Beam Defense deterministically from the frozen ordnance weapon rule identified by the projectile's `weapon_id`.

This preserves the Gate-2 actor schema and locks the already-frozen values:

- Nuclear Missile: 50;
- Merculite Missile: 75;
- Pulson Missile: 100;
- Zeon Missile: 130;
- Anti-Matter Torpedo: 100;
- Proton Torpedo: 100;
- Plasma Torpedo: 120.

Fast Missile configuration does not alter Beam-interceptor DCV.

## Manual Beam resolution

Manual projectile fire reuses the accepted Slice-17.6 Beam configuration and deterministic resolution rules:

- canonical mount profile;
- firing-arc test;
- geometric/effective range index;
- range to-hit modifier;
- Beam threshold and effective-roll cap;
- source Beam Offense versus projectile interceptor DCV;
- configured static hit adjustment and always-hit behavior;
- frozen damage-range scaling;
- configured physical/burst shot order;
- exactly the Slice-17.6 Beam hit RNG consumed per shot;
- Beam damage interpolation from the returned hit result.

Projectile targets do not execute Ship Shield/Armor/Structure selection or damage logic. A successful Beam shot passes its returned Beam damage directly to the projectile interception arithmetic.

Normal manual fire consumes the mount's ordinary activation readiness exactly like ship-target Beam fire.

`beam_fired` / `beam_volley_resolved` remain the authoritative Beam-resolution events. Manual projectile shots add:

- `target_ship_id = 0`;
- `target_projectile_id`;
- authoritative projectile x/y.

## Projectile interception arithmetic

Checkpoint 4 implements the frozen grouped-actor durability formula exactly.

For incoming damage `D`:

`total = intercept_damage_remainder + D`

`destroyed = min(multiplicity, floor(total / unit_durability))`

Then:

- subtract `destroyed * unit_durability` from total;
- decrement multiplicity by destroyed;
- if units remain, persist total as the new interception remainder;
- if multiplicity reaches zero, the remainder is canonicalized to zero and the actor is terminal.

No Shield, Armor, Structure, Marine or ship-damage selection path participates.

A dedicated vector starts from:

- multiplicity 3;
- durability 4;
- remainder 2.

Applying damage 7 yields:

- destroyed 2;
- multiplicity 1;
- remainder 1.

Applying damage 3 then yields:

- destroyed 1;
- multiplicity 0;
- remainder 0;
- terminal interception.

## Projectile-side event contract

Every applied positive Beam interception mutation emits `projectile_interception_resolved` with:

- `projectile_id`;
- `interceptor_source_kind` = `beam_manual` or `beam_reactive`;
- interceptor Ship ID;
- interceptor weapon slot;
- damage;
- destroyed units;
- multiplicity before/after;
- remainder before/after.

When multiplicity reaches zero, the actor is removed by its owning lifecycle and emits:

`projectile_destroyed { reason: "intercepted" }`

Beam shot identity remains in the existing Beam events; the new interception event records only projectile-side state mutation.

## Reactive pre-impact Point Defense

The Checkpoint-3 reach/impact handoff now invokes Checkpoint 4 before MIRV split and before impact-strength / ordinary-impact resolution.

For a live target Ship:

1. enumerate executable Point Defense Beam slots;
2. reject Standard and Heavy mounts from the automatic reaction;
3. filter by authoritative projectile x/y firing arc and valid frozen Beam range;
4. sort remaining slots by ascending weapon slot;
5. derive projectile interceptor DCV from the frozen ordnance rule;
6. call the already-frozen internal Slice-17.6 `resolveReactiveBeamPointDefense` for each slot while the actor survives;
7. consume returned Beam shots in their frozen order;
8. apply each positive hit's returned damage to projectile multiplicity/remainder;
9. stop immediately when multiplicity reaches zero.

The frozen reactive resolver restores the source Ship's ordinary weapon readiness after its call. Checkpoint 4 does not add a readiness check or readiness mutation around that contract.

Reactive Beam events retain their Slice-17.6 identity:

- `target_ship_id = 0`;
- target x/y = authoritative projectile coordinate;
- `reactive_point_defense = true`;
- command sequence 0.

The following `projectile_interception_resolved` event provides the projectile identity and projectile-side mutation.

## MIRV ordering

The frozen pre-impact order is now executable:

1. target validity / retarget;
2. neutral Slice-17.8 pre-impact hook position;
3. reactive Beam Point Defense;
4. if the carrier survives, MIRV multiplicity expansion;
5. MIRV split event;
6. impact-strength / Plasma dissipation derivation;
7. per-unit ordinary impact resolution.

A dedicated MIRV-carrier regression uses a Point Defense Phasor on the target Ship. The carrier is destroyed by reactive Beam damage before the split:

- reactive Beam RNG is consumed;
- ordinary readiness remains unchanged even when it was already false;
- `projectile_interception_resolved` records the terminal projectile mutation;
- `projectile_destroyed` reason is `intercepted`;
- no `projectile_mirv_split` occurs;
- no `projectile_impact_resolved` occurs;
- no ordinary impact RNG is consumed after interception.

## Regression coverage

Dedicated Checkpoint-4 regressions cover:

- interception remainder arithmetic across multiple damage applications;
- projectile-side event before/after values;
- Standard manual projectile target projection;
- Point Defense manual projectile target projection;
- Heavy projectile-target omission while preserving Heavy ship-target fire;
- manual projectile Beam destruction and readiness consumption;
- explicit Heavy projectile command rejection without mutation;
- empty/ambiguous Beam target payload rejection without mutation;
- reactive Point Defense carrier kill before MIRV;
- reactive readiness preservation;
- no post-interception MIRV/impact;
- ascending reactive Point Defense slot order;
- automatic Standard Beam exclusion.

Existing Slice-17.6 reactive Beam tests and all prior Slice-17.7 Checkpoint regressions remain green.

## Scope boundary

A focused production scan confirms Checkpoint 4 does not activate Slice-17.8-owned mechanics. There is no Anti-Missile Rockets, Fast Missile Racks, Lightning Field, Dauntless Guidance, fighter or bomber behavior added to the 17.7 core.

Checkpoint 5 remains responsible for final persistence/replay/UI/integration acceptance and browser presentation of projectile targeting/interception.

## QA

Focused Battle suite:

`go test ./internal/battle -count=1`

- PASS.

Focused integration suite:

`go test ./internal/ruleset ./internal/battle ./internal/game ./internal/session -count=1`

- PASS in all four packages.

`git diff --check`

- PASS; only the repository's known LF-to-CRLF warnings are emitted on Windows.

`go vet ./...`

- PASS on the current Checkpoint-4 implementation.

Final repository-wide `go test ./... -count=1`:

- PASS on the exact Checkpoint-4 commit candidate.

## Next checkpoint

Gate 3 Checkpoint 5 owns final persistence/UI/integration acceptance:

- active projectile/remainder save-load and deterministic continuation verification;
- replay/projection acceptance for manual/reactive interception;
- browser legal-target transport for `target_projectile_id`;
- authoritative projectile marker/event presentation;
- launch/move/retarget/interception/MIRV/impact/destruction animation handoff;
- live Missile ammo / Torpedo cooldown presentation;
- final full-stack Slice-17.7 acceptance without client-side simulation.
