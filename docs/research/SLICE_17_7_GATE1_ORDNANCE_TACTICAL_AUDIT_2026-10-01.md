# Slice 17.7 Gate 1 - Tactical Ordnance Audit
Date opened: 2026-10-01
Status: **COMPLETE - Gate 1 evidence closed; Gate 2 contract freeze ready (2026-10-01)**

## Objective

Re-audit the supported Slice-17.3 missile/torpedo/bomb matrix for server-authoritative Tactical execution before freezing any Slice-17.7 combat contract.

## Audit checklist

- [x] Supported missile/torpedo/bomb identities and design-snapshot inputs.
- [x] Ammo, salvo, readiness and launch semantics.
- [x] Range, travel, guidance, target tracking and projectile lifecycle.
- [x] Hit, damage, Shield/Armor/Structure application and RNG ordering.
- [x] Beam Point Defense target eligibility/interception ordering and projectile-side outcome.
- [x] Defensive countermeasures represented by current design snapshots.
- [x] Bomb combat context and planetary-bombardment ownership.
- [x] BattleSession projectile state, persistence, replay and Observer/player projection requirements.
- [x] Tactical UI/event/VFX requirements without client-side simulation.
- [x] Deterministic reference vectors and explicit unknowns/blockers.

## Gate boundary

This document is Gate-1 evidence only. No Tactical ordnance gameplay implementation is authorized until Gate 2 freezes the supported contract.


## Evidence update - original Tactical runtime (2026-10-01)

### Primary executable method

The Slice-17.3 design audit already anchored the original MOO2 1.31 LE executable and the ordnance table. Slice 17.7 now extends that audit into the original Tactical runtime.

Watcom debug records in `Orion2.exe` expose Object-1 code offsets directly. The relevant original symbols are:

- `Fire_Missile_` - `0x2C892`;
- `Move_Missile_` - `0x2CEB7`;
- `Resolve_Missile_` - `0x2D2DF`;
- `Did_Missile_Hit_` - `0x2BCCD`;
- `Kill_Missile_` - `0x2BC80`;
- `Missile_Speed_` - `0x2CD21`;
- `Missile_Dcv_` - `0x2DFE0`;
- `Get_Missile_Hits_` - `0x2E095`;
- `Retarget_Missile_` - `0x2DDD8`;
- `Apply_Damage_To_Missile_` - `0x2A0B9`;
- `Weapon_Can_Fire_At_Missiles_` - `0x2A7AA`;
- `Anti_Missile_Rockets_` - `0x9BA09`;
- `Fire_Bomb_` - `0x2A6CA`;
- `Strategic_Bombardment_` - `0x3257E`;
- `Tactical_Bombardment_` - `0x32781`.

Temporary disassembly/callgraph tooling remains outside the repository under `C:\ASH\Temp\moox-s177-g1-*`.

### Missile / Torpedo projectile lifecycle - directly evidenced

The original runtime uses an explicit missile-object lifecycle shared by missile/torpedo combat ordnance.

`Fire_Missile_` calls, in its launch path:

1. `Next_Missile_Id_`;
2. structure initialization;
3. `Missile_Facing_`;
4. `Missile_Speed_`;
5. `Missile_Launch_FX_`;
6. `Resolve_Missile_`;
7. `Kill_Missile_` on terminal launch-path branches.

`Move_Missile_` contains direct calls to:

- `Coord_Movement_`;
- `Move_Missile_` for continued/sub-step movement;
- `Resolve_Missile_`;
- `Did_Missile_Hit_`;
- `Missile_Speed_`;
- `Kill_Missile_`.

`Resolve_Missile_` is the central impact/resolution path. Its direct callgraph includes:

- `Is_Missile_Target_Alive_`;
- `Does_Combat_Ship_Have_Special_`;
- deterministic `Random_` consumers;
- `Lightning_Field_`;
- `Final_Fire_`;
- `Relative_Bearing_XY_`;
- `Find_Weakest_Shield_` / `Facing_Shield_`;
- `Missile_Speed_`;
- `Apply_Ship_Damage_`;
- `Kill_Missile_`;
- `Retarget_Missile_`.

This establishes that projectile existence/movement, target validity, impact resolution, defensive-special interaction, ship damage, retargeting and projectile destruction are one server-side combat concern in the original game. It does **not** yet freeze the exact branch conditions or ordering of every special.

`Retarget_Missile_` directly checks:

- `Target_Is_Legal_For_Weapon_`;
- `Does_Combat_Planet_Have_Defenses_`;
- `Ship_Is_In_Stasis_`;
- `Range_To_Missile_`;
- `Kill_Missile_` when no acceptable continuation target remains.

Therefore Slice 17.7 needs authoritative target-lock/retarget state; a browser-local projectile animation cannot own targeting.

### Collision versus combat accuracy

`Did_Missile_Hit_` is a geometric collision-radius test. The function computes squared coordinate distance between the projectile and target and compares it with a target/hit radius. It is distinct from the later offensive/defensive combat-value calculation.

Gate 2 must therefore not collapse:

- projectile reaches/collides with target;
- missile attack / target missile-defense calculation;
- Point Defense interception;

into one generic "hit roll".

### Missile / Torpedo speed, durability and interceptor DCV - directly evidenced

`Missile_Speed_` and `Get_Missile_Hits_` now close the supported projectile baseline:

- Nuclear / Merculite / Pulson / Zeon Missile: base combat speed **12**;
- Missile modifier `Fast`: actual flight speed **+4**, therefore 16;
- Missile base durability: **4 / 8 / 12 / 16** respectively;
- Missile modifier `Armored`: doubles that durability;
- Anti-Matter / Proton Torpedo speed: **20**;
- Plasma Torpedo speed: **24**;
- supported Torpedoes use the separate base durability path **5**.

`Missile_Dcv_` is not the ordinary missile-versus-ship accuracy formula. The Beam fire path calls it when a Beam attacks a projectile actor, and subtracts the returned value from the Beam attack derivation.

For the supported Slice-17.7 ordnance rows the original 1.31 executable computes:

| Projectile | Interceptor DCV |
| --- | ---: |
| Nuclear Missile | 50 |
| Merculite Missile | 75 |
| Pulson Missile | 100 |
| Zeon Missile | 130 |
| Anti-Matter Torpedo | 100 |
| Proton Torpedo | 100 |
| Plasma Torpedo | 120 |

The base is `5 * Missile_Speed_(weapon, 0)` plus the original fixed Missile-type correction. Supported Missile/Torpedo rows do **not** take the `Get_Fleet_Pilot_Bonus_` branch present for other original actor/weapon types.

Important original-parity quirk: `Missile_Dcv_` explicitly calls `Missile_Speed_` with modifier bits **zero**. Consequently `Fast` changes actual Missile flight speed 12 -> 16 but does **not** raise this Beam-interceptor DCV in the audited 1.31 executable. Gate 2 must freeze the executable behavior rather than silently replacing it with the HELP wording about a corresponding defense increase.

These are live projectile properties, not immutable design Count/Ammo fields.

### Normal impact accuracy and ECCM - directly evidenced

`Did_Missile_Hit_` remains the geometric collision/radius test. Once an actor reaches impact resolution, normal Missile/Torpedo accuracy uses the target combat ship's `+0x38` Missile Defense/Evasion:

1. start from target Missile Defense;
2. apply the best attacker scanner reduction observed by the original path: **-20 / -40 / -70**;
3. clamp at zero;
4. if ordnance modifier `ECCM` (`0x0400`, high-byte bit `0x04`) is present, halve that remaining defense with signed integer shift;
5. consume `Random_(100)` for the normal impact check;
6. the normal hit path continues only when the random result is **strictly greater** than the resulting Missile Defense/Evasion.

Separate combat Specials can force an earlier miss through their own RNG branches. Those Special identities/effects remain Slice 17.8 dependencies and are not absorbed into the 17.7 core formula.

Therefore Gate 2 must keep three independent concepts: geometric collision, projectile impact accuracy/evasion, and interceptor fire against the projectile actor.

### MIRV split and projectile multiplicity - directly evidenced

`Fire_Missile_` stores installed weapon `Count` in actor field `+0x0F`. The actor therefore represents a grouped volley with an authoritative multiplicity, not one actor per physical warhead.

In `Resolve_Missile_`:

1. pre-impact defensive handling runs first, including Lightning Field and the reactive `Final_Fire_`/Beam-PD opportunity;
2. only after those carrier-level defenses, MIRV bit `0x0200` multiplies the surviving actor multiplicity by **4** (`actor+0x0F <<= 2`);
3. the subsequent impact loop iterates that final multiplicity and performs separate hit/damage attempts.

This matches HELP: four warheads split immediately before impact. The original does **not** spawn four independently flying actors, so there is no per-child flight durability inheritance to invent.

### Beam Point Defense / projectile-side handoff - directly evidenced

The original Beam fire path directly consumes projectile state:

- `Fire_Beam_Weapon_` calls `Missile_Dcv_` for projectile targets;
- successful projectile-target Beam fire passes damage to `Apply_Damage_To_Missile_`;
- `Apply_Damage_To_Missile_` uses `Get_Missile_Hits_`, actor modifier state and actor partial-damage remainder to determine how many projectile units are destroyed;
- actor `+0x0D` is accumulated partial intercept damage/remainder; actor `+0x0F` is grouped projectile multiplicity.

`Weapon_Can_Fire_At_Missiles_` closes the Slice-17.7 relevant eligibility rule:

- Beam type `0`: eligible unless Heavy bit `0x0002` is set; therefore Standard and Point Defense Beam mounts can engage projectile actors, Heavy Beam mounts cannot;
- Missile type `1` and Torpedo type `2`: cannot be used as projectile interceptors;
- several explicit original non-Beam IDs are also allowed, but they belong to Fighter/Special consumers outside the current 17.7 Beam/Ordnance family and remain 17.8-owned.

`Resolve_Missile_` calls reactive `Final_Fire_` **before** MIRV split and before normal impact hit/damage. This is the exact Slice-17.6/17.7 handoff point: 17.6 derives the Beam reaction; 17.7 owns the actor, DCV input, damage application, multiplicity and destruction.

### Defensive Special hooks - evidenced, behavior owned outside 17.7

Original HELP and executable paths establish the hooks that Slice 17.7 must expose without implementing the Specials themselves:

- Lightning Field: pre-impact 50% destruction opportunity for missile/torpedo/fighter actors; active Special behavior remains 17.8;
- Anti-Missile Rockets: `Fire_Special_Weapon_` consumer, only ordinary Missile actors (plus a separate original non-17.7 actor class), max range 15, hit chance **85% - 5% per square**, projectile HP/durability damage and `Kill_Missile_` at zero; active Special behavior remains 17.8;
- Dauntless Guidance System: friend-or-foe retarget behavior for missiles/torpedoes; Special behavior remains 17.8 while 17.7 owns the target-lock/retarget-capable actor state;
- Fast Missile Racks: readiness/volley behavior remains an active Special dependency in 17.8.

Because Anti-Missile Rockets are dispatched from original `Fire_Special_Weapon_`, their active ordering relative to other Specials is not made a new 17.7-owned rule. Gate 2 needs only a deterministic projectile-interception hook for later Special consumers.

### Bomb ownership - directly evidenced

Bombs are not on the ordinary ship-to-ship missile projectile path.

`Fire_Bomb_` resolves against a planet and directly calls `Apply_Damage_To_Planet_`.

`Strategic_Bombardment_` separately calls:

- `Missile_Attack_`;
- `Beam_Attack_`;
- `Strat_Bomb_Attack_`.

`Tactical_Bombardment_` constructs a colony-bombardment combat context and directly calls:

- `Load_Combat_Ship_`;
- `Load_Tactical_Colony_`;
- `Set_Combat_Ordnance_Values_`;
- offensive/defensive combat bonus setup;
- `Init_Special_Devices_`;
- `Fire_Ship_`.

Together with HELP text describing the conventional bombs as orbital bombardment weapons, this is sufficient to freeze one Gate-1 ownership fact:

**Bomb/Bio rows must not become ordinary ship-to-ship Tactical weapon actions merely because they are present in a ship design snapshot. Their runtime owner is the planet/colony bombardment path.**

Gate 2 still needs to decide how much of that existing bombardment owner belongs inside Slice 17.7 versus a later dedicated bombardment slice, but ordinary ship-target bomb firing is not supported by current evidence.

## Current MOOX runtime boundary - verified

Current production/runtime behavior was checked against the closed Slice-17.3 handoff:

- `ShipWeaponMount.Ammo` and fixed ordnance `FiringArcID="360"` survive into `battle.TacticalWeaponSpec`;
- modifier/configuration metadata survives into the Tactical view;
- current ordnance rules remain `tactical_support="metadata_only"`;
- Tactical `weapon_actions` intentionally reports the later-family consumer as unavailable;
- no live projectile collection exists in `tacticalRuntime`; current runtime state is ship Tactical state + events + event sequence.

Focused regressions passed:

- `TestSlice173TacticalMetadataPreservesOrdnanceSnapshotWithoutEnablingDeferredFire`;
- `TestSlice175Checkpoint4StableUnavailableReasons`;
- `TestSlice173OrdnanceQuoteCanonicalAmmoArcAndBaseArithmetic`;
- `TestSlice173AllTechCatalogExposesHighTechOrdnance`.

This is the correct pre-17.7 baseline. Gate 2 may replace `metadata_only` only after the Tactical contract is frozen.

## Derived MOOX runtime state requirements

The following are **MOOX architectural requirements derived from the original runtime plus existing save/replay rules**, not claims that MOOX must copy the original memory layout byte-for-byte.

### Per installed ordnance group

- source Ship ID and weapon slot;
- immutable weapon/modifier identity, Count and designed rack Ammo/capacity;
- live remaining Missile volleys;
- live Torpedo cooldown/readiness counter;
- stable consumer/support identity for projection and replay.

Original launch/readiness semantics now evidenced:

- one Missile activation consumes **one rack volley**, independent of installed Count;
- installed `Count` becomes projectile multiplicity for that volley;
- a Torpedo shot writes cooldown value **2** to the live slot state;
- `Init_Ship_For_Start_Of_Turn_` decrements positive Torpedo cooldown once per turn start;
- therefore after firing, the next turn begins at 1 (still unavailable) and the following turn reaches 0 (ready): exact every-other-turn behavior.

Fast Missile Racks may alter readiness/volley opportunities later, but that active Special remains Slice 17.8-owned.

### Per active projectile actor

The original fields directly motivate authoritative MOOX facts for:

- stable projectile ID;
- source seat/Ship/weapon slot and target Ship ID;
- weapon identity and modifier/effect identity;
- authoritative X/Y position and facing;
- speed;
- remaining movement budget;
- elapsed/age state required for range dissipation/lifetime;
- grouped projectile multiplicity (`Count`, and MIRV-expanded multiplicity immediately pre-impact);
- accumulated partial interception damage/remainder;
- current strength/damage for dissipating Torpedoes;
- active/target-lock/retarget state;
- deterministic creation/event ordering.

### Movement unit and Plasma dissipation

`Fire_Missile_` sets actor speed from `Missile_Speed_` and initializes movement budget to `speed * 20`. `Move_Missile_` caps movement to the remaining budget, applies coordinate movement, then subtracts the actual moved amount. Therefore one original combat-speed point corresponds to **20 coordinate movement units**, and one full combat field/square is the unit used by the dissipation arithmetic.

For Plasma Torpedo:

- start strength/damage is 120 before modifiers;
- `Overloaded` adds +50% warhead strength before range dissipation;
- traveled full fields are derived from elapsed movement and remaining budget;
- Plasma loses **5 strength per completed field traveled**;
- No Range Dissipation (`0x0020`) skips that loss block;
- if remaining strength reaches zero or below, the actor is killed before impact damage.

### Persistence / replay / projection boundary

BattleSession save/load/replay must persist the complete active projectile actor list, per-group live ammo/cooldown, next projectile/event identity and Tactical RNG state. Observer/player projections must be detached copies and must not leak authoritative RNG internals. React may animate server events but must never advance movement, choose retargets, run interception, roll hits or apply damage.

Gate 2 will freeze exact DTO and event field names. Gate 1 freezes the required information content and authority boundary.

## Deterministic Gate-1 reference vectors

These vectors are now sufficiently evidenced to serve as Gate-2 contract inputs.

### A - Nuclear Missile baseline

- weapon: Nuclear Missile;
- damage: 8;
- flight speed: 12;
- Beam-interceptor DCV: 50;
- per-projectile durability: 4;
- one activation consumes one live rack volley;
- installed Count becomes grouped actor multiplicity;
- no MIRV: final impact multiplicity remains Count.

### B - Zeon Missile / Fast / Armored / ECCM

- weapon damage: 30;
- flight speed: 12 -> **16** from Fast;
- Beam-interceptor DCV remains **130** in the audited executable because `Missile_Dcv_` ignores modifier bits;
- durability: 16 -> **32** from Armored;
- ECCM does not alter interceptor DCV; at ship impact it halves the target Missile Defense remaining after scanner reduction/clamp.

### C - Nuclear Missile / MIRV

- launch produces one grouped carrier actor with multiplicity = installed Count;
- Lightning Field / reactive Beam-PD act on the carrier before split;
- immediately pre-impact MIRV multiplies surviving multiplicity by 4;
- per-warhead impact attempts then iterate the expanded multiplicity;
- no separately flying MIRV child actors exist in the original.

### D - Plasma Torpedo / Overloaded / No Range Dissipation variants

- base speed: 24;
- base durability: 5;
- base strength: 120;
- firing writes cooldown 2; subsequent turn starts decrement 2 -> 1 -> 0;
- Overloaded starts from 150% warhead strength;
- without No Range Dissipation, subtract 5 for each completed field traveled before impact;
- with No Range Dissipation, bypass that subtraction;
- Beam-interceptor DCV: 120.

### E - reactive Beam interceptor handoff

1. Slice 17.7 owns incoming actor coordinates, multiplicity, durability/remainder and DCV;
2. original Resolve order reaches pre-impact defenses and reactive `Final_Fire_` before MIRV split;
3. eligible supported Beam interceptors are non-Heavy Beam mounts;
4. Slice 17.6 consumes its Beam RNG/damage contract and returns the reaction result without consuming ordinary activation readiness;
5. Slice 17.7 converts returned damage into destroyed projectile units/remaining partial damage and owns actor destruction;
6. normal ship-impact RNG is separate and happens only for surviving impact multiplicity.

### F - Nuclear Bomb / planet context

- damage: 3-12;
- `Fire_Bomb_` targets the planet damage path;
- strategic and tactical colony bombardment have dedicated original entry points;
- no ordinary ship-target Tactical Bomb action is created by Slice 17.7.

## Original deterministic ordering relevant to Gate 2

For the supported core path the audited order is:

1. launch actor/group and initialize movement/target state;
2. move actor using authoritative movement budget;
3. geometric collision/reach test;
4. validate/retarget target state as allowed by the owning rules;
5. pre-impact Special hook(s), including Lightning Field when present;
6. reactive `Final_Fire_` / Beam-PD opportunity;
7. MIRV multiplicity expansion when selected;
8. derive impact damage/strength, including Plasma range dissipation and ECCM-adjusted target Missile Defense;
9. consume per-warhead impact RNG / forced-miss Special hooks;
10. apply ship damage for successful surviving impacts;
11. kill or retarget the projectile actor as required.

Original actor-processing loops observed in the Tactical path scan the fixed 300-record projectile array in ascending actor index. MOOX Gate 2 may represent this as ascending stable projectile ID/order; the required property is deterministic replay-stable ordering, not reproduction of a 300-slot memory array.

## Gate-1 closure: facts versus Gate-2 choices

Gate 1 is **complete**. No unresolved original fact blocks the Slice-17.7 contract freeze.

Facts closed by Gate 1 include:

- supported Missile/Torpedo identities, speed, damage/strength, durability and modifier legality;
- Count versus rack Ammo/live volley semantics;
- grouped projectile actor/multiplicity and MIRV split timing;
- movement-budget unit and Plasma dissipation ordering;
- Torpedo every-other-turn cooldown semantics;
- target Missile Defense / scanner / ECCM impact ordering;
- projectile DCV used by Beam interception and the Fast/DCV executable quirk;
- relevant Beam interceptor eligibility;
- reactive Beam-PD placement before MIRV/impact;
- Bomb/Bio bombardment-only ownership;
- explicit 17.8 hooks for active Specials such as Lightning Field, Anti-Missile Rockets, Dauntless Guidance and Fast Missile Racks.

Gate 2 still must **choose and freeze MOOX representation**, including:

- exact projectile DTO/JSON field names and schema-version migration;
- exact command/rejection/event names;
- stable projectile-ID allocation and event emission order;
- how the existing 17.6 reactive Beam-PD API is invoked from the 17.7 projectile resolver;
- the explicit handoff API exposed later to 17.8 Special consumers;
- the bombardment API boundary while keeping Bombs out of ordinary ship-target Tactical actions.

Those are contract-design decisions, not missing Gate-1 evidence.

## Gate 1 result

**PASS - Gate 1 evidence complete; Gate 2 contract freeze is ready.**

No Tactical ordnance gameplay implementation is authorized by this document. Production behavior remains `metadata_only` until Gate 2 freezes the contract and Gate 3 implements it.
