# Slice 17.6 Gate 1 - Tactical Beam Effect Audit

Date: 2026-09-30
Status: **IN PROGRESS - checkpoint 1 evidence committed; Gate 1 not complete**

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

## 12. Open Gate-1 evidence items

The following are intentionally **not yet frozen**:

1. the current server Weapon Selector preview/API shape and the minimum shared configured/effective Beam-stat model;
2. the existing Beam fire/damage event payload versus fields needed for misses, Autofire/grouped shots, Beam identity and per-facing effects;
3. deterministic representative early/mid/late reference vectors;
4. final interaction-order table for all accepted selected + innate effects.

Gate 1 remains active until those items are closed or explicitly evidenced as downstream dependencies.

## 13. Recovery point

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

Checkpoint 2 closes the marine-loss, Enveloping-order and Beam-PD handoff items. Selector preview, Beam event/VFX payloads, final interaction ordering and deterministic vectors remain open. No production implementation is authorized by this checkpoint.
