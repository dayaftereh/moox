# Slice 17.7 Gate 3 - Checkpoint 3 Impact / Modifiers Evidence

Date: 2026-10-02
Status: COMPLETE

## Scope

Checkpoint 3 implements the frozen Gate-2 projectile configuration, impact accuracy and ship-damage contract on top of the completed Checkpoint-1 state foundation and Checkpoint-2 launch/movement/readiness runtime.

Implemented here:

- shared configured-projectile derivation used by launch and impact;
- Missile Fast, Armored, ECCM and MIRV;
- Torpedo ECCM, Overloaded, Enveloping and Plasma-only No Range Dissipation;
- geometric reach -> authoritative impact handoff;
- target Missile Defense -> source scanner reduction -> zero clamp -> ECCM /2 ordering;
- exactly one Tactical `Random(100)` per surviving impact unit;
- hit rule `raw_roll > effective_missile_defense`;
- MIRV grouped-carrier split immediately before core impact attempts;
- Plasma Torpedo distance dissipation and pre-RNG `dissipated` destruction;
- reuse of the accepted shared Shield -> Armor -> Structure ship-damage path;
- Enveloping application in stable shield-arc order 0,1,2,3;
- projectile identity on authoritative `battle_damage_applied` events;
- projectile impact/destruction events and terminal BattleSession result propagation.

Checkpoint 3 intentionally does not implement manual Beam projectile targeting, reactive Beam Point Defense handoff, projectile interception damage/remainder, Anti-Missile Rockets, or active Slice-17.8 combat-Special behavior. Those remain owned by Checkpoint 4 / Slice 17.8.

## Configured projectile derivation

A single server helper now derives the runtime projectile configuration from the frozen ordnance rule row plus the weapon's persisted modifier IDs.

Missiles:

- base speed/durability/DCV come from the frozen rule snapshot;
- Fast adds the frozen +4 speed bonus;
- Armored doubles per-unit durability;
- ECCM marks the impact-defense rule;
- MIRV marks grouped pre-impact expansion;
- Fast deliberately leaves Beam-interceptor DCV unchanged;
- Emissions Guidance remains unavailable because its active behavior belongs outside the Slice-17.7 Gate-3 core.

Torpedoes:

- ECCM uses the same impact-defense rule;
- Overloaded sets warhead strength to `base + floor(base/2)`;
- Enveloping enables four stable damage applications;
- No Range Dissipation is executable only for Plasma Torpedoes and suppresses Plasma range-strength loss.

This derivation is used at launch, so the authoritative actor already carries configured speed, durability and warhead strength before movement begins.

## Frozen reference vectors

### Zeon Missile / Fast / Armored / ECCM

The reference vector is covered end-to-end:

- base speed 12 -> configured speed 16;
- base durability 16 -> configured durability 32;
- Beam-interceptor DCV remains 130;
- warhead remains 30;
- a launch from x=10 advances to x=26 on its immediate movement quantum;
- distance traveled is 320 movement units.

### Nuclear Missile / MIRV

MIRV launches as one grouped actor. At impact, after the neutral Checkpoint-4 reaction-hook position, surviving multiplicity expands by the frozen multiplier 4.

For Count=2:

- multiplicity before = 2;
- multiplicity after = 8;
- one `projectile_mirv_split` event is emitted before the first impact;
- exactly eight ordered impact attempts consume eight core impact RNG rolls.

No separately flying child actors are created.

### Plasma Torpedo

The frozen Plasma actor remains speed 24 / durability 5 / DCV 120.

Without No Range Dissipation:

`impact_strength = configured_warhead - 5 * floor(distance_traveled / 20)`

If the resulting strength is zero or negative, the actor is removed with reason `dissipated` before any impact RNG or ship damage is consumed.

With Overloaded:

- base 120 -> configured warhead 180.

With No Range Dissipation:

- the 5-per-field subtraction is suppressed;
- speed, durability and DCV are unchanged.

## Impact defense and RNG

For each impact unit, the server performs exactly the frozen order:

1. start with target `MissileDefense`;
2. subtract source `MissileScannerReduction`;
3. clamp to at least zero;
4. if ECCM is present, integer-divide the remainder by two;
5. retain the Slice-17.8 forced-miss hook position as neutral;
6. consume exactly one Tactical `Random(100)`;
7. hit only when `raw_roll > effective_missile_defense`;
8. apply warhead damage on hit;
9. continue with the next impact index.

A focused ECCM vector uses target defense 80 and source scanner reduction 20, producing 60 then ECCM 30. The deterministic first two core rolls are 100 and 19: the first unit hits and the second misses, and the post-two-roll Tactical RNG state is regression-locked.

No RNG is consumed by geometric movement, retargeting, Plasma dissipation, or the inactive Slice-17.8 hooks.

## Ship damage and Enveloping

Ordinary Missile/Torpedo hits pass current impact strength into the existing accepted shared ship-damage function. This preserves the established Shield -> Armor -> Structure behavior and destruction state.

`battle_damage_applied` remains the authoritative ship-damage event and now carries additive ordnance identity:

- `projectile_id`;
- source Ship / weapon slot / weapon ID / modifiers;
- impact index / count;
- application index / count;
- existing shield, armor and structure mutation facts.

Enveloping applies a fresh copy of the successful impact strength four times in stable shield-facing order:

`0, 1, 2, 3`

The core impact still consumes only the single impact RNG roll for that projectile unit; the four damage applications do not create extra hit rolls.

## Impact lifecycle and terminal results

Checkpoint 2 movement now hands a projectile to Checkpoint 3 immediately when the authoritative actor reaches the target coordinate, including:

- immediate launch-quantum reach;
- later round-transition reach;
- actors already located on the target coordinate.

After core impact resolution the actor is removed and emits:

- `projectile_destroyed { reason: "impact_complete" }`, or
- `projectile_destroyed { reason: "dissipated" }` for exhausted Plasma strength.

A newly destroyed target emits `ship_destroyed`. BattleSession then uses the existing strategic/tactical winner derivation to emit `winner_determined` and return a normalized `TacticalOutcomeVictory` result. This works both for immediate launch impacts and impacts triggered during a round transition.

## Events

Checkpoint 3 adds the frozen authoritative events:

### `projectile_mirv_split`

- projectile ID;
- multiplicity before;
- multiplicity after.

### `projectile_impact_resolved`

Per impact unit:

- projectile ID;
- impact index/count;
- target Ship ID;
- effective Missile Defense;
- raw hit roll;
- hit boolean;
- current warhead strength;
- damage applied.

Existing `battle_damage_applied`, `ship_destroyed`, `winner_determined` and `projectile_destroyed` events remain authoritative and replay-ordered.

## Regression coverage

Dedicated Checkpoint-3 regressions cover:

- configured Zeon Fast / Armored / ECCM vector;
- configured Plasma Overloaded / NRD vector;
- real launch actor configuration for Zeon and Plasma;
- ECCM scanner/clamp/divide ordering;
- exact per-impact RNG order;
- hit/miss behavior;
- MIRV split ordering and expanded multiplicity;
- Plasma dissipation before RNG;
- NRD suppression of Plasma dissipation;
- Overloaded 180-strength Plasma impact;
- Enveloping arc order 0,1,2,3;
- additive projectile identity on `battle_damage_applied`;
- `impact_complete` / `dissipated` actor removal;
- immediate-impact terminal BattleSession result;
- round-transition-impact terminal BattleSession result.

## QA

Checkpoint-3 focused package tests:

`go test ./internal/battle -count=1`

- PASS.

Focused integration suite:

`go test ./internal/ruleset ./internal/battle ./internal/game ./internal/session -count=1`

- PASS in all four packages.

Final repository-wide QA:

`go test ./... -count=1`

- PASS across the repository.

`go vet ./...`

- PASS.

`git diff --check`

- PASS; only the repository's known LF-to-CRLF warnings are emitted on Windows.

A dedicated scope scan confirms Checkpoint 4 has not been pulled forward: no `target_projectile_id`, manual/reactive Beam interception resolution, `projectile_interception_resolved`, or interception-remainder mutation is introduced by Checkpoint 3.

## Next checkpoint

Gate 3 Checkpoint 4 owns Interception:

- manual Standard/Point-Defense Beam projectile targets;
- Heavy Beam exclusion;
- reactive pre-impact Point Defense handoff to the frozen Slice-17.6 resolver;
- projectile multiplicity/remainder damage application;
- `projectile_interception_resolved` and intercepted destruction;
- neutral Slice-17.8 interceptor hook positions.

Checkpoint 5 remains persistence/UI/integration acceptance across live save/load/replay and browser presentation.
