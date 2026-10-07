# Slice 17.7 Gate 2 - Tactical Ordnance Contract Freeze

Date: 2026-10-02
Status: **FROZEN / COMPLETE - Gate 3 implementation ready / not started**

Parent: Slice 17 - Military Ship Design + Tactical Combat
Child: Slice 17.7 - Tactical Missiles / Torpedoes / Bombs

Evidence authority:

- `docs/research/SLICE_17_7_GATE1_ORDNANCE_TACTICAL_AUDIT_2026-10-01.md`
- closed Slice 17.3 ordnance identity/design contract
- closed Slice 17.5 Tactical foundation contract
- closed Slice 17.6 Tactical Beam / reactive Point Defense contract
- current BattleSession command/event/live-snapshot conventions

This document freezes the implementation/API/event contract for Gate 3. Gate 3 may implement production Tactical ordnance only to match this contract; it must not invent new gameplay semantics.

## 1. Authority and slice boundary

### 1.1 Sole gameplay authority

`BattleSession` remains the sole mutable ship-Tactical gameplay authority.

The browser:

- never advances projectile movement;
- never chooses or changes a projectile target;
- never decrements ammo or torpedo cooldown;
- never computes projectile durability, DCV, hit chance or damage;
- never invokes Point Defense;
- never consumes Tactical RNG;
- never decides MIRV split, Plasma dissipation, interception or impact outcome.

React may render detached server state and animate authoritative events only.

### 1.2 Slice ownership

Slice 17.7 owns:

- Missile/Torpedo launch commands and live rack/cooldown state;
- Missile/Torpedo projectile actor creation and stable identity;
- movement, target lock, retarget-capable state and deterministic actor order;
- core impact accuracy, ECCM, supported warhead modifiers and ship damage;
- projectile durability / interceptor DCV;
- application of Beam damage to projectile actors;
- the call site for the frozen 17.6 reactive Beam-PD capability;
- projectile persistence/replay/projection and event-driven presentation;
- the explicit ship-Tactical exclusion boundary for Bomb/Bio rows.

Slice 17.6 continues to own Beam hit/damage/RNG semantics and the reactive Point Defense Beam resolver.

Slice 17.8 owns active combat-Special behavior, including Lightning Field, Anti-Missile Rockets, Dauntless Guidance System, Fast Missile Racks, ECM/scanner Special contributions and forced-miss Special branches. Slice 17.7 freezes deterministic hook points only.

## 2. Supported runtime matrix

### 2.1 Executable ship-Tactical ordnance

Gate 3 makes these seven rows executable with `TacticalSupport="ordnance"`:

| Weapon | Family | Base damage/strength | Speed | Durability | Beam-interceptor DCV |
| --- | --- | ---: | ---: | ---: | ---: |
| Nuclear Missile | missile | 8 | 12 | 4 | 50 |
| Merculite Missile | missile | 14 | 12 | 8 | 75 |
| Pulson Missile | missile | 20 | 12 | 12 | 100 |
| Zeon Missile | missile | 30 | 12 | 16 | 130 |
| Anti-Matter Torpedo | torpedo | 25 | 20 | 5 | 100 |
| Proton Torpedo | torpedo | 40 | 20 | 5 | 100 |
| Plasma Torpedo | torpedo | 120 | 24 | 5 | 120 |

All seven retain fixed firing arc `360`.

### 2.2 Bomb/Bio rows

These six rows become `TacticalSupport="bombardment_only"`:

- Nuclear Bomb;
- Fusion Bomb;
- Anti-Matter Bomb;
- Neutronium Bomb;
- Death Spores;
- Bio Terminator.

They do **not** receive `battle.fire_ordnance`, legal ship targets, projectiles or ship-Tactical damage events.

The current repository has no authoritative MOOX colony/planet bombardment resolver. Gate 3 must therefore preserve these rows as truthful bombardment-only metadata and must not invent population/planetary damage formulas. A future bombardment owner may consume the frozen 17.3 snapshots separately.

## 3. Tactical ruleset contract

Gate 3 extends the existing tactical rules file additively. It does not scatter original ordnance constants through BattleSession code.

The existing tactical-combat schema number remains **1** for this additive extension so current active snapshots remain restorable. New fields are validated when present/required by `TacticalSupport="ordnance"`.

Freeze one server-owned ordnance rule block equivalent to:

- `movement_units_per_field = 20`;
- `fast_speed_bonus = 4`;
- `torpedo_cooldown_rounds = 2`;
- `mirv_multiplier = 4`;
- `plasma_dissipation_per_field = 5`.

Each executable Missile/Torpedo weapon rule additionally freezes:

- base projectile speed;
- base durability;
- Beam-interceptor DCV.

`TacticalRulesSnapshot` must carry the ordnance rule values required to replay an active battle without consulting mutable external rules.

## 4. Immutable snapshot versus mutable runtime state

### 4.1 Immutable `TacticalWeaponSpec`

The existing immutable snapshot remains authoritative for:

- slot;
- `weapon_id`;
- family;
- physical installed `Count`;
- canonical modifier IDs;
- designed `Ammo`/rack capacity;
- fixed firing arc;
- base MinDamage / MaxDamage;
- Tactical support identity.

Designed `Ammo` is never mutated.

### 4.2 Live installed-ordnance state

Gate 3 adds an ordnance-specific live state record per executable Missile/Torpedo slot:

`TacticalOrdnanceWeaponState`

with JSON content frozen as:

- `ship_id`;
- `weapon_slot`;
- `ammo_remaining`;
- `cooldown_rounds`.

Records are sorted by `ship_id`, then `weapon_slot`.

Missiles initialize `ammo_remaining = TacticalWeaponSpec.Ammo`.
Torpedoes do not consume rack ammo; `ammo_remaining = 0`.

### 4.3 Active projectile state

Gate 3 adds:

`TacticalProjectileState`

with these frozen JSON fields:

- `projectile_id` - non-zero monotone battle-local ID;
- `source_seat_id`;
- `source_ship_id`;
- `weapon_slot`;
- `target_ship_id`;
- `weapon_id`;
- `weapon_family`;
- `modifier_ids`;
- `x`;
- `y`;
- `facing`;
- `speed`;
- `movement_remaining`;
- `distance_traveled`;
- `multiplicity`;
- `unit_durability`;
- `intercept_damage_remainder`;
- `warhead_strength`.

Only active unresolved actors remain in the active projectile list.

### 4.4 Ordnance runtime container

Gate 3 adds one authoritative ordnance runtime container with:

- `next_projectile_id`;
- ordered `weapon_states`;
- ordered active `projectiles`.

`next_projectile_id` starts at 1 and IDs are never reused during a BattleSession.

Projectile processing, projection and persistence order is ascending `projectile_id`.

## 5. Action / command contract

### 5.1 Player launch command

Freeze:

`CommandFireOrdnance = "battle.fire_ordnance"`

and payload:

`FireOrdnancePayload`

- `ship_id`;
- `target_ship_id`;
- `weapon_slot`.

Only Missile/Torpedo `TacticalSupport="ordnance"` rows may receive this command.

### 5.2 Action availability

Existing `TacticalWeaponActionAvailability` remains the server projection.

For executable Missile/Torpedo rows:

- `command_kind = "battle.fire_ordnance"`;
- targets are live opponent Ships legal under the current BattleSession target rules;
- no browser-local target filtering is authoritative.

Freeze additive unavailable reasons:

- `ammo_exhausted` - Missile rack has zero live volleys;
- `weapon_cooldown` - Torpedo cooldown is non-zero;
- `ordnance_effect_unavailable` - selected design-valid ordnance effect has no frozen Tactical consumer.

Existing `weapon_not_ready`, `no_legal_targets` and `tactical_consumer_unavailable` remain valid where applicable.

Rejected commands consume zero RNG, mutate no ammo/cooldown/projectile/readiness state and append no gameplay event.

### 5.3 Emissions Guidance boundary

`emissions_guidance` remains design-valid for the four Missiles, but its original drive/subsystem damage consumer has no current MOOX subsystem authority.

Gate 3 therefore must **not** approximate it as Armor Piercing, Structure damage or ordinary damage.

A Missile row carrying `emissions_guidance` projects `ordnance_effect_unavailable` and cannot be launched until a dedicated subsystem-damage contract exists.

This is the explicit Gate-2 unsupported-behavior freeze required by the Slice plan.

## 6. Ammo, readiness and cooldown

### 6.1 Missile launch

One successful Missile command:

1. requires `ammo_remaining > 0`;
2. consumes exactly **one** live rack volley;
3. creates one grouped projectile actor;
4. actor `multiplicity = installed Count`;
5. marks the weapon unavailable for the rest of the current normal activation through the existing readiness path.

Installed Count never multiplies rack-ammo consumption.

### 6.2 Torpedo launch

One successful Torpedo command:

1. requires `cooldown_rounds == 0`;
2. creates one grouped projectile actor with `multiplicity = installed Count`;
3. sets `cooldown_rounds = 2`;
4. marks the weapon unavailable for the rest of the current activation.

At each deterministic new-round transition, positive Torpedo cooldown decrements once.

Therefore a fired Torpedo follows `2 -> 1 -> 0` over subsequent round starts and is available every other turn.

### 6.3 Normal round readiness

At new-round reset:

- Beam readiness keeps its existing behavior;
- Missile readiness becomes true only when live rack ammo remains;
- Torpedo readiness becomes true only when its decremented cooldown is zero;
- later Fast Missile Racks adjustments are a Slice-17.8 hook and are not applied by Gate 3.

## 7. Projectile movement and deterministic scheduling

### 7.1 MOOX movement unit

One Tactical grid field consumes **20 movement units**.

For one projectile movement quantum:

`movement_remaining = speed * 20`.

`distance_traveled` accumulates actual movement units.

The actor moves server-authoritatively toward its current locked target; the browser only animates the resulting authoritative move event.

### 7.2 BattleSession scheduling adaptation

MOOX freezes one projectile movement quantum:

1. immediately after successful launch;
2. once for each still-active projectile at the deterministic new-round transition before the first Ship activation of that round.

This gives every projectile exactly one movement quantum in its launch round and one per later round, independent of the number of Ships or UI commands.

At a round transition:

1. apply existing round-transition state reset/recharge semantics;
2. decrement Torpedo cooldowns and derive Missile/Torpedo readiness;
3. process active projectiles in ascending `projectile_id`;
4. remove/resolve destroyed Ships and projectiles;
5. recompute initiative / first active Ship from surviving combatants.

No projectile movement is tied to browser animation completion.

### 7.3 Reach versus accuracy

Projectile geometry and combat accuracy remain separate.

The movement resolver first decides whether the actor reaches/collides with the locked target.

Only a reached target enters impact/interception resolution and can consume impact RNG.

## 8. Target lock and retargeting

Every active actor carries authoritative `target_ship_id`.

If the locked target is no longer valid before impact:

- base Slice-17.7 behavior deterministically selects the first legal opponent by ascending Ship ID;
- if no legal target exists, destroy the projectile with reason `no_target`.

The deterministic target list/order is server-owned.

Dauntless Guidance may later override/extend retarget eligibility through the frozen Slice-17.8 hook; Gate 3 does not implement its friend-or-foe behavior.

## 9. Configured projectile derivation

One shared server helper must derive configured projectile facts for BattleSession and tests.

### 9.1 Missile modifiers

Supported core runtime modifiers:

- MIRV;
- ECCM;
- Armored;
- Fast.

Frozen effects:

- `Fast`: speed 12 -> 16;
- `Armored`: unit durability doubles;
- `ECCM`: impact-defense rule in section 11;
- `MIRV`: section 12.

Original-parity quirk: `Fast` does **not** change Beam-interceptor DCV.

### 9.2 Torpedo modifiers

Supported core runtime modifiers:

- ECCM;
- Overloaded;
- Enveloping;
- Plasma-only No Range Dissipation.

`Overloaded` freezes configured warhead strength as:

`base + floor(base / 2)`.

For Plasma Torpedo this gives 180 before dissipation.

`Enveloping` reuses the already-frozen four-facing application identity from Slice 17.6: one successful impact applies a fresh copy of the impact damage to all four shield facings in stable arc order 0,1,2,3.

## 10. Beam interception against projectile actors

### 10.1 Manual Beam targeting

Slice 17.7 extends existing Beam target transport additively.

`TacticalFireTarget` gains optional `target_projectile_id`.

`FireBeamPayload` gains optional `target_projectile_id`.

Exactly one of `target_ship_id` or `target_projectile_id` must be non-zero.

For projectile targets:

- Standard Beam mounts are legal;
- Point Defense Beam mounts are legal;
- Heavy Beam mounts are illegal;
- target Beam Defense input is the projectile's frozen interceptor DCV;
- normal activation Beam fire consumes ordinary weapon readiness;
- Slice 17.6 still owns Beam RNG/hit/damage derivation;
- Slice 17.7 applies returned Beam damage to projectile multiplicity/remainder.

### 10.2 Reactive pre-impact Point Defense

Reactive pre-impact fire is distinct from manual Beam targeting.

Immediately before MIRV split / ordinary impact:

1. Slice 17.7 enumerates the target Ship's executable Point Defense Beam slots in ascending slot order;
2. for each slot while the actor still exists, call the already-frozen
   `ResolveReactiveBeamPointDefense`;
3. request coordinates are the authoritative projectile `x/y`;
4. `TargetBeamDefense = projectile interceptor DCV`;
5. consume the returned Beam shots in their frozen shot order;
6. apply each hit's damage to projectile state;
7. ordinary activation readiness remains unchanged by the reactive call, exactly as frozen in Slice 17.6.

Standard Beam mounts do not participate in this automatic pre-impact reaction.

## 11. Projectile interception damage

For one projectile actor:

- `unit_durability` is the configured per-unit HP;
- `multiplicity` is the number of surviving units;
- `intercept_damage_remainder` is accumulated damage already applied to the next surviving unit.

For incoming interceptor damage `D`:

1. `total = intercept_damage_remainder + D`;
2. `destroyed = min(multiplicity, floor(total / unit_durability))`;
3. subtract `destroyed * unit_durability` from `total`;
4. decrement multiplicity by `destroyed`;
5. if multiplicity remains positive, persist `intercept_damage_remainder = total`;
6. if multiplicity reaches zero, remove the actor and emit destruction/interception events.

No Shield/Armor/Structure logic applies to projectile durability.

## 12. MIRV timing

A MIRV Missile launches as one grouped carrier actor.

Pre-impact order is:

1. target validity / retarget;
2. Slice-17.8 pre-impact hook position (inactive in Gate 3 base);
3. reactive Beam-PD;
4. if the carrier survives, MIRV expands
   `multiplicity *= 4`;
5. emit authoritative MIRV split event;
6. iterate impact attempts over the expanded multiplicity.

No separately flying child actors are created.

## 13. Impact accuracy and RNG order

For each surviving impact unit:

1. start from target `MissileDefense`;
2. subtract source `MissileScannerReduction`;
3. clamp to at least zero;
4. when ECCM is selected, integer-divide the remaining defense by 2;
5. later Slice-17.8 forced-miss hooks occupy their frozen pre-roll positions;
6. consume exactly one Tactical `Random(100)` for the core impact attempt;
7. hit only when `raw_roll > effective_missile_defense`;
8. on hit, apply the frozen warhead damage/application behavior;
9. proceed to the next impact unit in ascending impact index.

Gate 3 adds additive replay-stable ship snapshot/view fields:

- `missile_defense`;
- `missile_scanner_reduction`.

Until Slice 17.8 activates the relevant Special consumers, Gate 3 projects these contributions as zero. They are explicit hooks, not browser formulas.

Projectile impact RNG is separate from:

- geometric reach/collision;
- Beam/manual interceptor RNG;
- reactive Beam-PD RNG;
- later Special RNG.

## 14. Damage / Plasma dissipation

### 14.1 Ordinary Missile/Torpedo hit

A successful non-Enveloping warhead passes its current `warhead_strength` into the already-frozen shared Shield -> Armor -> Structure damage application.

Existing `battle_damage_applied` remains the authoritative ship-damage event and gains projectile identity fields where the source is an ordnance impact.

### 14.2 Plasma Torpedo

Configured Plasma base strength is 120, or 180 with Overloaded.

Before impact:

- derive completed fields as `floor(distance_traveled / 20)`;
- unless No Range Dissipation is selected, subtract `5 * completed_fields`;
- if resulting strength <= 0, destroy the actor with reason `dissipated` before impact RNG/damage.

No Range Dissipation suppresses this subtraction only; it does not alter speed, durability or DCV.

## 15. Event contract

All ordnance events use the existing `TacticalEvent` envelope and monotonically increasing event sequence.

Freeze event kinds:

### `ordnance_launched`

Required data:

- `projectile_id`;
- source Ship/seat/weapon slot;
- target Ship ID;
- weapon/family/modifier identity;
- start x/y/facing;
- speed;
- multiplicity;
- unit durability;
- interceptor DCV;
- warhead strength;
- ammo remaining or cooldown after launch.

### `projectile_moved`

Required data:

- `projectile_id`;
- from x/y;
- to x/y;
- movement used;
- movement remaining;
- distance traveled;
- target Ship ID.

### `projectile_retargeted`

Required data:

- `projectile_id`;
- previous target Ship ID;
- new target Ship ID;
- reason.

### `projectile_interception_resolved`

Required data:

- `projectile_id`;
- interceptor source kind (`beam_manual`, `beam_reactive`, or later Special hook identity);
- interceptor Ship/weapon slot where applicable;
- damage;
- destroyed units;
- multiplicity before/after;
- remainder before/after.

Beam-specific shot identity remains in the existing `beam_fired` / `beam_volley_resolved` events. This event records only the projectile-side state result.

### `projectile_mirv_split`

Required data:

- `projectile_id`;
- multiplicity before;
- multiplicity after.

### `projectile_impact_resolved`

Required data:

- `projectile_id`;
- `impact_index`;
- `impact_count`;
- target Ship ID;
- effective Missile Defense;
- raw hit roll;
- hit boolean;
- warhead strength;
- damage applied.

### `projectile_destroyed`

Required data:

- `projectile_id`;
- reason: `intercepted`, `dissipated`, `no_target`, or `impact_complete`.

Events are emitted in the same order as authoritative mutations. React never synthesizes missing intermediate gameplay facts.

## 16. Presentation contract

The Tactical UI:

- renders projectile markers from authoritative projected projectile state;
- animates launch/move/retarget/interception/MIRV/impact/destruction from events;
- shows live Missile ammo and Torpedo cooldown from server projection;
- may show projectile multiplicity/durability/target facts already exposed by the server;
- never extrapolates projectile position after the last authoritative event;
- never delays state mutation until animation completion.

Manual Beam targeting may surface projectile actors as targetable entities only when the server action projection includes their IDs.

## 17. Persistence / replay / legacy migration

### 17.1 New live snapshot content

`TacticalLiveSnapshot` gains optional ordnance live state containing:

- next projectile ID;
- live ordnance weapon states;
- active projectiles.

Every mutable field required by sections 4-14 is persisted.

Tactical RNG and event sequence continue to use the existing live-snapshot fields.

### 17.2 Legacy compatibility

`LiveSnapshotSchemaVersion` remains **1** for this additive change.

A missing ordnance live-state object means a pre-17.7 snapshot.

Because pre-17.7 production could not launch Missile/Torpedo actors, deterministic legacy migration is:

- active projectile list empty;
- `next_projectile_id = 1`;
- Missile `ammo_remaining = TacticalWeaponSpec.Ammo`;
- Torpedo `cooldown_rounds = 0`.

After migration, new snapshots always emit the explicit ordnance live-state object.

This preserves current reference/live games without pretending that an old snapshot could contain hidden ordnance actions.

### 17.3 Projection safety

Observer and participant views receive detached projectile/ordnance state.

PlayerView continues to zero authoritative RNG state.

No view may expose mutable aliases into BattleSession state.

Replay of the same snapshot + command stream must reproduce projectile IDs, events, RNG, impacts and destruction byte-for-byte/canonically.

## 18. Slice-17.8 hook boundary

Gate 3 implements deterministic hook positions but not active Special semantics.

Frozen hook phases:

1. `ordnance_readiness` - Fast Missile Racks;
2. `retarget` - Dauntless Guidance;
3. `pre_impact_defense` - Lightning Field / forced-miss defenses;
4. `projectile_interceptor` - Anti-Missile Rockets and later Special interceptors;
5. `impact_defense` - ECM/scanner/Special contributions.

Absent an active Slice-17.8 consumer, hooks are neutral and consume zero RNG.

Gate 3 must not hard-code 17.8 Special effects into the 17.7 core.

## 19. Bombardment boundary

Bomb/Bio design snapshots remain fully preserved.

Inside ordinary ship Tactical:

- no Bomb launch command;
- no Bomb projectile;
- no ship-target Bomb action;
- no Bomb RNG or damage consumption.

The `bombardment_only` support identity is the explicit handoff for a future colony/planet combat owner.

Gate 3 acceptance for Bomb/Bio is truthful exclusion and persistence, not invented planetary simulation.

## 20. Deterministic frozen reference vectors

Gate 3 must encode at least these vectors.

### A - Nuclear Missile baseline

- Count N, designed Ammo A;
- live ammo initializes A;
- launch consumes one volley -> A-1;
- actor multiplicity N;
- speed 12;
- durability 4;
- DCV 50;
- warhead 8;
- no MIRV.

### B - Zeon / Fast / Armored / ECCM

- speed 16;
- durability 32;
- DCV remains 130;
- target defense derives scanner reduction/clamp then ECCM /2;
- one RNG per impact unit.

### C - Nuclear / MIRV

- launch actor multiplicity = Count;
- pre-impact reactive defenses occur before split;
- surviving multiplicity becomes Count*4;
- one impact attempt/RNG per expanded unit.

### D - Plasma Torpedo variants

- speed 24;
- durability 5;
- DCV 120;
- cooldown 2 -> 1 -> 0 across round starts;
- base strength 120;
- Overloaded strength 180;
- Plasma loses 5 per completed field;
- No Range Dissipation suppresses that loss.

### E - reactive Beam-PD

- projectile DCV is passed as `TargetBeamDefense`;
- Point Defense Beam slots process ascending;
- Slice 17.6 Beam RNG/events occur first;
- Slice 17.7 applies damage to multiplicity/remainder;
- ordinary activation readiness is unchanged;
- surviving carrier reaches MIRV/impact afterward.

### F - manual Beam projectile target

- Standard and Point Defense Beam mounts may target actor;
- Heavy cannot;
- normal Beam readiness is consumed;
- Beam event carries `target_projectile_id`;
- projectile-side interception event follows Beam event(s).

### G - Bomb exclusion

- Bomb/Bio metadata survives Tactical snapshot/view;
- support is `bombardment_only`;
- no `battle.fire_ordnance` action or projectile is produced.

## 21. Gate-3 implementation checkpoints

Gate 3 should be executed as bounded checkpoints:

1. **Rules/state/projection** - ordnance rules, support identities, live weapon/projectile DTOs, legacy migration, detached views.
2. **Launch/movement/readiness** - command/action availability, ammo, Torpedo cooldown, stable IDs, movement/retarget ordering.
3. **Impact/modifiers** - accuracy/ECCM, MIRV, Armored/Fast, Torpedo Overloaded/Enveloping/NRD, shared ship damage.
4. **Interception** - manual Beam projectile targets, reactive 17.6 PD handoff, projectile damage/remainder, neutral 17.8 hooks.
5. **Persistence/UI/integration** - live save/load/replay, Observer/player projections, event-driven browser presentation, Selector/Build -> BattleSession ordnance execution and vectors A-G.

Gate 3 does not start until this Gate-2 freeze is committed and the repository is clean.

## 22. Gate 2 result

**FROZEN / PASS - Gate 2 complete; Gate 3 implementation ready / not started.**

No production Tactical ordnance gameplay behavior is changed by this document.

### 2026-10-07 user-review refinement - partial trajectory

For a projectile quantum that cannot reach the locked target, the shared moving-actor geometry spends the full available field budget and chooses the straightest in-budget grid endpoint toward the target before using remaining distance and deterministic coordinate order as tie-breakers. This preserves the existing ceil-Euclidean field metric while preventing axis-heavy partial movement that visually travels straight on X/Y and corrects toward the target only in a later round.
