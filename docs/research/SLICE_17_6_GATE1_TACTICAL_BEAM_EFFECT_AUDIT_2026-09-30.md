# Slice 17.6 Gate 1 - Tactical Beam Effect Audit

Date: 2026-09-30
Status: **COMPLETE - Gate 1 accepted; Gate 2 contract freeze ready / not started**

Parent: Slice 17 - Military Ship Design + Tactical Combat
Child: Slice 17.6 - Tactical Beam Weapons & Beam Modifiers

## 1. Scope and evidence standard

Gate 1 audits the original Tactical Beam rules required before Gate 2 may freeze the Slice-17.6 contract. No production Beam implementation is allowed in Gate 1.

Authoritative original executable:

- path used for local audit: `C:\ASH\Installation\mastori2\Orion2.exe`
- SHA-256: `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`

Primary executable symbols / object offsets used in this checkpoint:

| Symbol / source | Original offset |
|---|---:|
| `Range_To_Ship_` | `0x1885A` |
| `Apply_Internal_Damage_` | `0x25251` |
| `Select_Internal_` | `0x26810` |
| `Fire_Ship_` | `0x28B5E` |
| `Get_Beam_Range_To_Hit_Bonus_` | `0x29406` |
| `Get_Beam_Weapon_Modifiers_` | `0x29434` |
| `Fire_Beam_Weapon_` | `0x294F7` |
| `Apply_Ship_Damage_` | `0x29985` |
| `Facing_Shield_` | `0x29DE0` |
| `Weapon_In_Range_` | `0x29F1D` |
| original RNG `Random_` | `0x1147A0` |
| Object-2 `_weapons` base used by prior 17.2 audit | `0x7803` records, stride `0x1C` |

External disassembly/extraction helpers remain outside the repository under:

- `C:\ASH\Temp\moox-s173-g1-audit`
- `C:\ASH\Temp\moox-s176-g1-audit`

This audit also reuses already-accepted evidence from:

- `docs/research/TACTICAL_SHIP_COMBAT_BASELINE_2026-09-01.md`
- `docs/research/SLICE_17_2_GATE1_BEAM_WEAPONS_AUDIT_2026-09-20.md`
- `docs/research/SLICE_17_2_GATE2_BEAM_CONTRACT_FREEZE_2026-09-20.md`
- `docs/research/SLICE_17_2A_GATE1_BEAM_FIRING_ARC_AUDIT_2026-09-25.md`

## 2. Reconfirmed normal Beam foundation

The Slice-07 Laser baseline already captured the executable's generic Beam hit/damage foundation. The current 17.6 audit rechecked the relevant symbols rather than inventing a second formula.

### 2.1 Range geometry

The normal ship-to-ship range geometry remains the accepted approximate metric:

`max(abs(dx), abs(dy)) + floor(min(abs(dx), abs(dy)) / 2)`

The resulting Beam range band/index feeds the original To-Hit and damage-dissipation tables.

### 2.2 Range To-Hit table

Normal Beam range indices 0..8 map to:

`0, 0, -10, -20, -30, -40, -55, -70, -85`

### 2.3 Damage-range table

Normal Beam damage range indices map to:

`0, 0, -10, -20, -30, -40, -50, -60, -65`

### 2.4 Hit roll

The accepted generic order remains:

1. derive range / range To-Hit adjustment;
2. apply Beam mount/modifier adjustments;
3. add attacker Beam Offense minus target Beam Defense;
4. draw `Random_(100)`;
5. raw values above 95 normalize to 100, otherwise the attack differential is applied and capped at 100;
6. compare against `min(95, 40 - rangeToHitModifier)`.

The browser must not duplicate this formula.

### 2.5 Damage interpolation

For a hit with configured min/max damage, the executable interpolates damage from the effective hit result and threshold:

`min + ((effectiveRoll - threshold) * (max - min + 1)) / (100 - threshold)`

with the original integer ordering and max cap.

## 3. Mount semantics directly closed

### 3.1 Standard

Standard leaves the ordinary effective range index and base damage multiplier unchanged.

### 3.2 Heavy

Direct `Get_Beam_Weapon_Modifiers_` evidence:

- effective range index is halved using signed integer round-toward-zero behavior;
- configured damage multiplier is increased by 50 percentage points from the base 100 to 150.

`Weapon_In_Range_` independently halves the measured distance for Heavy before range-family legality.

Therefore Heavy has the executable Tactical meaning:

- effectively longer reach in the range calculation;
- **150% configured Beam damage**.

### 3.3 Point Defense

Direct `Get_Beam_Weapon_Modifiers_` evidence:

- effective range index is doubled and capped at 8;
- To-Hit receives **+25**;
- configured damage multiplier is reduced from 100 to **50**.

`Weapon_In_Range_` independently doubles measured distance for Point Defense before range-family legality.

Point Defense is therefore a real ship-targeting Beam mount, not only a future missile-interceptor mode. Its projectile-interception hook is still open and must be frozen for Slice 17.7.

## 4. Selectable modifier semantics directly closed

The 17.2 design legality masks remain authoritative and are not changed here.

### 4.1 Autofire - `0x0080`

`Fire_Beam_Weapon_` sets the volley count to **3**.

Each shot follows the normal hit/damage path. The attack differential receives **-20 per shot**. RNG is therefore consumed per resolved shot in stable shot order.

### 4.2 Continuous - `0x0010`

Continuous adds **+25** to the Beam attack differential.

It does not create additional shots.

### 4.3 No Range Dissipation - `0x0020`

The selected modifier bypasses the normal damage-range dissipation adjustment. It does not remove the ordinary To-Hit range adjustment.

### 4.4 Shield Piercing - `0x0040`

`Apply_Ship_Damage_` treats Shield Piercing as a Shield-layer bypass:

- without Hard Shields, the Shield block amount is zeroed;
- the facing Shield pool is skipped;
- with Hard Shields, the executable neutralizes the Shield-Piercing bit before ordinary Shield absorption.

Particle Beam's innate Shield Piercing uses this same downstream path.

Hard Shields itself remains a Slice-17.8 combat-Special consumer. Slice 17.6 needs only the explicit Beam-side dependency/hook frozen in Gate 2.

### 4.5 Armor Piercing - `0x0008`

`Select_Internal_` proves internal type 7 is the Armor layer at ship field `+0xC2`.

Normal penetrating damage selects Armor first while current Armor remains. Armor Piercing bypasses that early Armor selection unless the original defenses neutralize it.

Original executable defenses observed in this path:

- Heavy Armor special;
- Xentronium armor material;
- Achilles Targeting Unit can force the deeper-internal path.

Those Special behaviors remain owned by Slice 17.8; the Beam API dependency must remain explicit rather than silently simulated.

### 4.6 Enveloping - `0x0100`

`Fire_Beam_Weapon_` enters the four-facing damage-application path when Enveloping is selected. Plasma Cannon's innate Enveloping flag normalizes into this same path.

Direct code at `0x29829..0x2989A` closes the ordering: the loop iterates facing indices **0, 1, 2, 3**. Before every `Apply_Ship_Damage_` call it copies the same original post-hit damage value into a fresh local mutable damage cell. Therefore each facing receives an **independent full application of the original Beam damage**, rather than carrying residual damage from the previous facing. Return/status flags are OR-aggregated, while the loop itself is not terminated by an earlier-facing hit.

## 5. Innate Beam-effect flags directly extracted

Original Object-2 weapon record field `+0x12` for the supported 13 Beam identities:

| Beam | original index | innate flags |
|---|---:|---:|
| Mass Driver | 1 | `0x0004` |
| Gauss Cannon | 2 | `0x0004` |
| Laser Cannon | 3 | `0x0000` |
| Particle Beam | 4 | `0x0100` |
| Fusion Beam | 5 | `0x0002` |
| Ion Pulse Cannon | 6 | `0x0008` |
| Graviton Beam | 7 | `0x0020` |
| Neutron Blaster | 8 | `0x0010` |
| Phasor | 9 | `0x0000` |
| Disrupter Cannon | 10 | `0x0004` |
| Death Ray | 11 | `0x0050` |
| Plasma Cannon | 12 | `0x0202` |
| Mauler Device | 27 | `0x0402` |

### 5.1 Natural no-range-dissipation - innate `0x0004`

Mass Driver, Gauss Cannon and Disrupter Cannon skip ordinary damage-range dissipation through the same effective path as selectable No Range Dissipation.

### 5.2 Particle Beam - innate `0x0100`

`Fire_Beam_Weapon_` normalizes Particle Beam's innate flag into selected Shield Piercing `0x0040`.

### 5.3 Plasma Cannon - innate `0x0200`

The innate high flag normalizes into selected Enveloping `0x0100`.

### 5.4 Mauler Device - innate `0x0400`

The raw Beam hit roll is forced to **100**, giving the executable always-hit behavior before downstream damage processing.

### 5.5 Death Ray - innate `0x0040`

This low innate flag is normalized by `Fire_Beam_Weapon_` into selected Continuous `0x0010`.

This explains the 17.2 HELP-vs-executable design-legality discrepancy: Continuous is a natural Death Ray effect, while the executable selectable-modifier mask only needs to permit Heavy.

## 6. Innate `0x0002`: executable range-To-Hit behavior

Fusion Beam, Plasma Cannon and Mauler Device carry innate `0x0002`.

Direct `Get_Beam_Weapon_Modifiers_` evidence shows this flag doubles the signed **range To-Hit adjustment** returned by `Get_Beam_Range_To_Hit_Bonus_`.

The damage-range table is a separate path and is not doubled by this flag.

For Mauler Device the To-Hit consequence is practically masked by its always-hit path.

This does not match all HELP prose one-to-one. Gate 1 records the discrepancy explicitly and treats the executable path as the gameplay authority for Gate 2.

## 7. Innate damage-path effects

### 7.1 Ion Pulse Cannon - innate `0x0008`

The Ion flag bypasses the ordinary Armor-first branch in `Select_Internal_`, sending damage into the deeper internal-selection path.

For target owner/player bands above 9 (the original monster/Antaran band), Ion damage is reduced to one quarter in the executable.

HELP describes Ion as ineffective against those targets; Gate 1 therefore preserves the exact executable behavior separately from the prose wording.

### 7.2 Graviton Beam - innate `0x0020`

`Apply_Ship_Damage_` adds extra StructureDamage at ship field `+0xC0` based on half of the current post-Shield damage, twice when the first half did not already reach max Structure.

Operationally, a surviving target receives one additional full current-damage amount as StructureDamage before normal downstream processing. If the first half already destroys the target, the second half is unnecessary and skipped.

### 7.3 Neutron Blaster / Death Ray - innate `0x0010`

The executable closes the marine-kill hook exactly:

- Damper Field suppresses the hook;
- the current mutable post-Shield damage is divided by **5** with signed integer division / truncation toward zero;
- that quotient is passed to original routine `0x266AF`;
- `0x266AF` subtracts the amount from tactical-ship byte field `+0xAF`, saturating at zero rather than underflowing;
- a `0xFF` value at `+0xAF` is treated as a sentinel and skips this loss path.

The executable context identifies `+0xAF` as the crew/marine-count field used by this effect; Gate 1 records the offset explicitly rather than inventing a broader crew model. Death Ray combines this hook with its innate Continuous effect.

## 8. Supported Beam range legality

All 13 supported Beam rows dispatch through the same Beam family case in `Weapon_In_Range_`.

For these 13 IDs the audited Beam case adds no weapon-specific hard maximum-distance rejection. Heavy/Point Defense first transform measured distance, after which practical Beam effectiveness is governed by the original To-Hit/damage range tables and battlefield geometry.

The special hard-distance exceptions observed in this function belong to other weapon IDs, not this 13-Beam set.

## 9. Beam Point Defense projectile handoff

Direct original code now closes the Beam-side interceptor contract required by Slice 17.7.

### 9.1 Incoming-actor reservation

Original routine `0x195D0` scans up to 300 Tactical actor records at global base `0x1AB14` with stride `0x1A`. For a candidate ship it detects active actors whose recorded target ship matches that ship and whose origin ship differs. The Beam AI path at `0x19B9E` checks Point Defense mount bit `0x0004`; when an incoming hostile actor exists, that PD mount is withheld from ordinary ship-target firing.

### 9.2 Reactive interceptor firing

Original function beginning at `0x26C9C` receives a ship plus one active `0x1A` Tactical actor index. It snapshots all eight weapon-slot readiness bytes, temporarily masks the slots, then evaluates candidate weapons. For Beam weapon type `0`, Point Defense bit `0x0004` explicitly enters the reactive-fire path. The function compares the actor coordinates against the defending ship position and calls `Fire_Ship_` using the actor X/Y coordinates and a distinct target-mode flag of `1`.

After the reaction pass, the original readiness bytes for all eight weapon slots are restored exactly. Therefore Beam Point Defense interception is a **reactive capability separate from ordinary activation readiness consumption**.

### 9.3 Slice boundary

Slice 17.6 must freeze and later expose a server-authoritative Beam-PD capability sufficient for a projectile actor to request an interception reaction, including the configured Beam identity/mount/modifiers and the fact that the reaction does not consume normal activation readiness. Slice 17.7 still owns projectile creation, flight, target locks, interception timing/eligibility, projectile durability/evasion, interception result, impact and projectile events. No projectile runtime belongs in Core ShipDesignSpec.

## 10. Combat-Special dependencies exposed by the original Beam path

The executable Beam path references several Special identities, including:

- Achilles Targeting Unit;
- Damper Field;
- Hard Shields;
- Heavy Armor;
- High Energy Focus;
- Rangemaster Unit;
- Structural Analyzer.

Examples already directly observed:

- Rangemaster Unit divides the range input by 3 for the Beam range-To-Hit lookup;
- High Energy Focus adds +50 to the damage multiplier;
- Structural Analyzer doubles a downstream post-Shield damage value;
- Hard Shields neutralizes Shield Piercing;
- Heavy Armor participates in Armor-Piercing resistance;
- Damper Field suppresses the Neutron/Death marine hook.

These are **dependencies**, not a claim that Slice 17.6 owns complete combat-Special implementation. Slice 17.8 retains the behavioral consumer scope; Gate 2 must freeze explicit Beam-side hooks so the boundaries compose later.

## 11. Current MOOX parity / gaps

Current Slice-17.5 foundation already provides:

- broad Tactical combatant admission;
- Beam Offense / Beam Defense;
- generic weapon actions;
- deterministic BattleSession RNG;
- Shield -> Armor -> Structure common damage path;
- four Shield-facing pools;
- event-driven Tactical presentation foundations;
- existing Laser firing vertical slice.

17.6 still needs the frozen generic Beam consumer that replaces the Laser-only execution gate without moving formulas into React.

## 12. Exact original damage scaling / rounding and current parity gap

Direct `Fire_Beam_Weapon_` code at `0x2967C..0x296F8` closes a rounding boundary that the original Slice-07 baseline fixture did not exercise.

### 12.1 Original effective damage percent

`Get_Beam_Weapon_Modifiers_` starts the Beam damage multiplier at 100 and applies Beam-side percentage adjustments, including:

- High Energy Focus dependency: +50;
- normal range damage dissipation unless selectable/innate No Range Dissipation suppresses it;
- Point Defense: -50;
- Heavy: +50.

The firing path then reads a player-indexed percentage from original table `0x21224`. That table is initialized as `100 + best applicable ship-leader-derived bonus` for the player. Gate 1 does not invent a broader leader model; the executable dependency is recorded as a player/ship-leader-derived Beam damage percent.

The combined percent used to scale the weapon table is additive:

`combinedDamagePercent = beamDamagePercent + (playerLeaderPercent - 100)`

### 12.2 Original min/max rounding

For positive supported Beam values the executable scales the configured base min/max with nearest-integer percentage rounding:

`scaled = (base * combinedDamagePercent + 50) / 100`

Then:

- scaled minimum is clamped to **at least 1**;
- scaled maximum is clamped to be **at least the scaled minimum**;
- only after those integer bounds are fixed does hit quality interpolate between min and max.

The later interpolation remains:

`damage = min + ((effectiveRoll - threshold) * (max - min + 1)) / (100 - threshold)`

with integer truncation at that division.

### 12.3 Current MOOX parity gap

Current `internal/battle/tactical.go::beamDamage` computes:

- `weapon.MinDamage * multiplier / 100`;
- `weapon.MaxDamage * multiplier / 100`;

without the original `+50` rounding and without the original minimum-damage clamp to 1. It also currently receives only the baseline range modifier because Slice 17.6 has not implemented Heavy/PD/HEF/leader-effective damage yet.

This is a **known Gate-1 parity gap**, not an authorization to patch production code before Gate 2 freezes the shared Beam calculator.

A concrete early-game edge case proves why the boundary matters: Laser Cannon base 1-4 at a -10% damage range band scales in the original to **1-4** (0.9 rounds/clamps to 1; 3.6 rounds to 4), while the current truncating helper would derive **0-3**.

## 13. Weapon Selector configured/effective preview audit

The current design preview already provides the correct server-owned transport path.

`MilitaryDesignWeaponRowPreview` currently returns, per concrete candidate row:

- slot;
- `weapon_id`;
- count;
- mount type;
- modifier IDs;
- ammo;
- firing arc;
- miniaturization level;
- unit/row Space;
- unit/row PP cost.

The row is built from the same authoritative server quote that validates/canonicalizes the candidate mount. `WeaponSelectorDialog` requests `previewMilitaryDesign` for the concrete candidate configuration rather than calculating Space/PP locally.

Therefore Gate 2 should extend this existing row/quote path with **server-derived Tactical configured/effective Beam fields**. It must not add a second React formula implementation.

The required architectural shape is one shared server-side Beam-effect derivation used by:

1. Designer/Weapon Selector preview for configured/effective values;
2. Tactical/BattleSession resolution for the same configured weapon snapshot;
3. Combat/Designer Lab golden vectors.

The exact public field names and which values are configuration-static versus runtime/range/target-dependent remain a Gate-2 contract decision.

## 14. Beam events, readiness and current presentation audit

### 14.1 Current server events

The existing Laser vertical slice already emits one `beam_fired` event per resolved physical shot containing:

- source ship ID;
- target ship ID;
- weapon slot;
- shot index / shot count;
- range index;
- raw and effective hit roll;
- threshold;
- hit flag;
- damage.

For hits, `battle_damage_applied` additionally carries the damage-layer result including Shield arc/current values, Armor/Structure before/after, block/absorption/overflow and the damage-selection roll.

For multi-shot baseline mounts, `beam_volley_resolved` aggregates ship/target/slot, shot count, hit count and total damage.

Weapon readiness is consumed **after the full current mount loop**, not after each individual `beam_fired` event.

### 14.2 Current UI behavior

`TacticalBattlefield` consumes only authoritative events:

- a single-shot `beam_fired` may create a beam animation;
- when shot count is greater than one, per-shot beam animations are suppressed and `beam_volley_resolved` drives one grouped presentation;
- damage layers are derived from matching `battle_damage_applied` events;
- animation lifetime is a browser timer and never gates BattleSession progression.

This is already the correct authority split for 17.6.

### 14.3 Event identity gap for 17.6

Current `beam_fired` / `beam_volley_resolved` payloads identify the slot but do **not** carry Beam identity or configured/effective effect identity. That was sufficient while every executable Beam was Laser.

For multiple Beam identities and effects, Gate 2 must freeze replay-safe server event/snapshot identity sufficient to distinguish at least:

- `weapon_id`;
- configured mount;
- configured modifiers / innate effects as required for presentation;
- grouped versus physical-shot semantics;
- Enveloping/per-facing damage results without client recomputation.

Exact redundancy between event data and immutable Tactical weapon snapshots is a Gate-2 decision, but React must not infer combat rules from weapon names.

## 15. Final original interaction order

The directly audited executable path freezes the following Gate-1 ordering for Gate-2 contract work.

### 15.1 Battery / physical-unit order

For a Beam battery, `Fire_Ship_` keeps a remaining physical weapon-unit count in its firing record. It requires that count to be positive, calls `Fire_Beam_Weapon_` once, decrements the remaining unit count, and can loop back for the next unit while the target remains valid.

Therefore:

- without Autofire: total Beam hit attempts = configured physical `Count`;
- with Autofire: each physical unit performs the inner three-shot Autofire loop, so total hit attempts = `Count * 3`;
- all three Autofire shots of one physical unit resolve before the next physical unit enters `Fire_Beam_Weapon_`.

### 15.2 Per-physical-unit setup order

For each physical unit:

1. load weapon identity and configured modifier/mount flags;
2. normalize innate weapon effects into the downstream effect flags where the executable does so:
   - Particle Beam -> Shield Piercing;
   - Plasma Cannon -> Enveloping;
   - Death Ray -> Continuous;
   - Mass Driver/Gauss/Disrupter -> No Range Dissipation path;
   - Fusion/Plasma/Mauler innate `0x0002` -> doubled range To-Hit adjustment;
   - Mauler -> Always-Hit override after the raw RNG draw;
3. derive tactical range / facing;
4. apply Heavy/Point-Defense range-index transformation;
5. obtain range To-Hit adjustment, including Rangemaster dependency and innate doubled range-To-Hit where applicable;
6. initialize damage percent at 100;
7. add High Energy Focus dependency (+50);
8. add normal range-damage modifier unless selected/innate No Range Dissipation suppresses it;
9. apply Point Defense (-50 damage, +25 To-Hit) or Heavy (+50 damage);
10. combine the resulting Beam damage percent with the player/ship-leader-derived percent from `0x21224`;
11. scale min/max damage with original nearest-integer percentage rounding, clamp minimum to at least 1, clamp maximum >= minimum;
12. form attacker Beam Offense - target Beam Defense;
13. if Autofire, set inner shot count to 3 and apply -20 to the attack differential for every inner shot;
14. if Continuous, add +25 to the attack differential.

### 15.3 Per-shot order

For every inner shot:

1. consume `Random_(100)` for the hit roll;
2. if Mauler innate Always Hit is present, overwrite the drawn raw roll with 100 **after the RNG call**; RNG is still consumed;
3. apply the remaining accepted hit-roll dependencies / later Slice-17.8 special hooks;
4. normalize raw roll >95 to effective 100; otherwise add the attack differential and cap at 100;
5. compute threshold `min(95, 40 - rangeToHitModifier)`;
6. miss if effective roll < threshold;
7. on hit, interpolate damage between already-scaled min/max with the integer hit-quality formula;
8. apply later target-side/Special hooks at their executable positions;
9. dispatch either:
   - one normal facing damage application; or
   - Enveloping facings 0,1,2,3, each with a fresh copy of the same original hit damage;
10. `Apply_Ship_Damage_` handles Shield Piercing / Hard Shields, Shield block/pool, Armor Piercing / Armor selection and lower-layer damage;
11. resolve innate post-layer effects such as Graviton extra StructureDamage and Neutron/Death marine loss;
12. consume any downstream damage-layer/internal-selection RNG in that shot before the next shot begins.

This ordering is the Gate-1 basis for deterministic BattleSession RNG and event sequencing.

### 15.4 Point Defense reaction order

Beam Point Defense projectile interception is separate from normal activation fire:

1. projectile/actor logic identifies an incoming actor;
2. the Beam PD reaction temporarily exposes eligible PD slots;
3. the reaction fires against actor coordinates;
4. the normal weapon readiness bytes are restored afterward.

Slice 17.7 owns the projectile lifecycle and calls into this Beam capability; normal activation readiness must not be consumed by the reactive PD handoff.

## 16. Deterministic original/reference vectors

These vectors use the directly audited executable rules and explicit raw hit inputs. They are intended as Gate-2/3 golden controls, not probability claims.

Common assumptions unless stated:

- player/ship-leader damage percent = 100;
- no combat Specials other than innate weapon effects;
- Beam Offense - Beam Defense = 0;
- no Hard Shields / Heavy Armor / other deferred target-side Special dependency.

### 16.1 Early-game rounding vector - Laser Cannon / Standard

Input:

- Laser base damage 1-4;
- Standard mount;
- effective range index 2;
- range To-Hit = -10;
- range damage = -10%;
- combined damage percent = 90;
- raw hit roll = 100.

Expected:

- scaled min = `max(1, (1*90+50)/100) = 1`;
- scaled max = `(4*90+50)/100 = 4`;
- threshold = 50;
- effective roll = 100;
- hit = true;
- final Beam damage = 4.

This is the canonical control for the current MOOX truncation parity gap: the existing baseline helper would derive 0-3 before interpolation.

### 16.2 Mid-game vector - Particle Beam / Heavy

Input:

- Particle Beam base damage 10-30;
- Heavy mount;
- geometric range index 4 -> Heavy effective index 2;
- effective range To-Hit = -10;
- range damage = -10%;
- Heavy damage adjustment = +50%;
- combined damage percent = 140;
- raw hit roll = 80;
- innate natural Shield Piercing.

Expected:

- scaled min/max = 14-42;
- threshold = 50;
- effective roll = 80;
- hit = true;
- interpolated Beam damage = 31;
- absent Hard Shields, the Shield layer is bypassed by innate Shield Piercing.

### 16.3 Late-game vector - Mauler Device / Heavy

Input:

- Mauler base damage 100-100;
- Heavy mount;
- geometric range index 4 -> Heavy effective index 2;
- normal range To-Hit at index 2 = -10, doubled by innate `0x0002` to -20;
- range damage = -10%;
- Heavy damage adjustment = +50%;
- combined damage percent = 140;
- one `Random_(100)` hit draw is still consumed;
- innate Always Hit overwrites that raw result to 100.

Expected:

- scaled min/max = 140-140;
- threshold = 60;
- effective roll = 100;
- hit = true;
- Beam damage = 140.

### 16.4 Point-Defense ship-target vector - Laser Cannon / PD

Input:

- Laser base damage 1-4;
- Point Defense mount;
- geometric range index 2 -> PD effective index 4;
- range To-Hit at index 4 = -30 plus PD +25 -> -5;
- range damage at index 4 = -30%;
- PD damage adjustment = -50%;
- combined damage percent = 20;
- raw hit roll = 50.

Expected:

- scaled min/max = 1-1 after original rounding/minimum clamp;
- threshold = 45;
- effective roll = 50;
- hit = true;
- Beam damage = 1.

This vector proves Point Defense remains a valid ship-target Beam mode in addition to its later projectile-reaction capability.

### 16.5 Autofire physical-count vector

Input:

- one Beam battery with configured physical `Count = 2`;
- Autofire modifier selected;
- target remains valid for the complete battery.

Expected:

- outer physical-unit calls = 2;
- inner Autofire shots per physical unit = 3;
- total hit attempts = 6;
- every hit attempt receives the Autofire -20 attack-differential adjustment;
- ordering is unit1 shot1/shot2/shot3, then unit2 shot1/shot2/shot3;
- each shot consumes its hit RNG before the next shot; hit-only downstream damage/internal RNG occurs before advancing to the next shot.

## 17. Gate-1 close result

Gate 1 original Beam-effect audit is complete.

Direct/original evidence now covers:

- all supported 13 Beam identities and innate effects required by 17.6;
- Standard / Heavy / Point Defense;
- all supported selectable Beam modifiers;
- range, hit, damage, exact percentage rounding, physical count, Autofire shot order and RNG order;
- Shield/Armor/Structure dependencies and explicit Slice-17.8 Special hooks;
- Beam-side Point Defense handoff to Slice 17.7;
- current server Weapon Selector configured-preview extension path;
- current Beam event / VFX / readiness behavior;
- deterministic early/mid/late plus PD/Autofire reference vectors.

No production behavior was changed during Gate 1. Gate 2 may now freeze the implementation/API/event contract from this evidence.

## 18. Open Gate-1 evidence items

The following are intentionally **not yet frozen**:

None. Gate 1 is complete; remaining work belongs to Gate 2 contract freeze.

## 19. Recovery point

Checkpoint-1 conclusions are safe to resume from without repeating the original audit:

- generic normal Beam formula: reconfirmed;
- Heavy: range index /2, damage 150%;
- Point Defense: range index x2 capped 8, +25 To-Hit, damage 50%;
- Autofire: 3 shots, -20 attack differential per shot;
- Continuous: +25 attack differential;
- No Range Dissipation: bypass damage-range dissipation;
- Shield Piercing: bypass block + facing Shield unless Hard Shields neutralizes it;
- Armor Piercing: bypass Armor-first selection subject to original defenses;
- Particle -> innate Shield Piercing;
- Plasma -> innate Enveloping;
- Mauler -> always hit;
- Death Ray -> innate Continuous + marine hook;
- Ion -> Armor bypass + quarter damage versus monster/Antaran band;
- Graviton -> extra StructureDamage;
- Fusion/Plasma/Mauler `0x0002` -> doubled range To-Hit adjustment;
- supported 13-Beam family: no additional per-weapon hard max-range gate in `Weapon_In_Range_`;
- Neutron/Death marine effect: post-Shield damage / 5, saturating loss at `ship+0xAF`, suppressed by Damper Field;
- Enveloping: four ordered facing applications 0..3, each reset to the same original Beam damage;
- Point Defense projectile handoff: incoming-actor reservation plus reactive `Fire_Ship_` call against actor coordinates, with normal slot readiness restored afterward.

Gate-1 final checkpoint closes physical Count x Autofire ordering, the final effect/RNG interaction order and deterministic early/mid/late plus PD/Autofire reference vectors. Gate 1 is complete. Gate 2 contract freeze is next; no production implementation is authorized yet.
