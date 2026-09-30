# Slice 17.6 Gate 2 - Tactical Beam Contract Freeze

Date: 2026-09-30
Status: **FROZEN / COMPLETE - Gate 3 implementation ready / not started**

Parent: Slice 17 - Military Ship Design + Tactical Combat
Child: Slice 17.6 - Tactical Beam Weapons & Beam Modifiers

Evidence authority:

- `docs/research/SLICE_17_6_GATE1_TACTICAL_BEAM_EFFECT_AUDIT_2026-09-30.md`
- closed Slice 17.2 Beam identity/mount/modifier contract
- closed Slice 17.2A Beam firing-arc contract
- closed Slice 17.5 Tactical combat foundation contract

This document freezes the implementation/API/event contract for Gate 3. Production Beam behavior may change only to match this contract; Gate 3 must not invent new gameplay semantics.

## 1. Authority and scope

### 1.1 Sole gameplay authority

`BattleSession` remains the sole mutable Tactical gameplay authority.

The browser:

- never computes hit chance;
- never computes effective range;
- never computes Beam damage;
- never decides Shield/Armor bypass;
- never advances RNG;
- never decides weapon readiness;
- never decides target legality;
- never simulates Point Defense interception.

React may display server-projected configured/effective facts and authoritative events only.

### 1.2 Shared server Beam derivation

Gate 3 must implement one shared server-owned Beam rule derivation that is reused by:

1. Ship Designer / Weapon Selector preview;
2. `BattleSession` Tactical Beam resolution;
3. deterministic Combat/Designer Lab tests and reference vectors.

A second independent preview formula is prohibited.

### 1.3 Slice boundaries

Slice 17.6 owns:

- supported Beam execution;
- Standard / Heavy / Point Defense Beam mount semantics;
- supported selected Beam modifiers;
- innate Beam effects;
- Beam range/hit/damage/RNG/readiness;
- Beam firing-arc legality;
- Beam-side reactive Point Defense capability;
- configured/effective Selector preview;
- Beam event payloads and presentation contract.

Slice 17.7 continues to own:

- missile/torpedo/bomb actor creation;
- projectile movement/flight;
- projectile locks and timing;
- projectile durability/evasion;
- deciding when a projectile requests Point Defense interception;
- interception outcome against projectile actors;
- projectile impact/destruction events.

Slice 17.8 continues to own active combat-Special behavior. Slice 17.6 freezes explicit dependency hooks for those effects but does not implement their broader Special mechanics.

## 2. Supported Tactical Beam matrix

All 13 Beam identities frozen by Slice 17.2 become executable Tactical Beam weapons in Gate 3.

| Weapon | Base damage | Frozen innate Tactical effect |
|---|---:|---|
| Mass Driver | 6-6 | no range damage dissipation |
| Gauss Cannon | 18-18 | no range damage dissipation |
| Laser Cannon | 1-4 | none |
| Particle Beam | 10-30 | natural Shield Piercing |
| Fusion Beam | 2-6 | doubled signed range To-Hit adjustment |
| Ion Pulse Cannon | 2-10 | Armor-first bypass; quarter damage versus original monster/Antaran band |
| Graviton Beam | 3-15 | extra StructureDamage |
| Neutron Blaster | 3-12 | marine kill |
| Phasor | 5-20 | none |
| Disrupter Cannon | 40-40 | no range damage dissipation |
| Death Ray | 50-100 | natural Continuous + marine kill |
| Plasma Cannon | 6-30 | natural Enveloping + doubled signed range To-Hit adjustment |
| Mauler Device | 100-100 | Always Hit after RNG draw + doubled signed range To-Hit adjustment |

The exact legal mount/modifier combinations and firing arcs remain those already frozen by Slices 17.2 and 17.2A. Slice 17.6 must neither widen nor narrow design legality.

Non-Beam families remain metadata-visible and non-executable until their owner slices.

## 3. Tactical support mode

The tactical support vocabulary is extended with:

- `beam` - executable by the full Slice-17.6 Beam consumer;
- `metadata_only` - visible but unavailable because the family consumer belongs to a later slice;
- existing Special support modes remain unchanged.

Gate 3 migrates all 13 accepted Beam identities to `TacticalSupport="beam"`.

The old `baseline` Laser-only execution predicate is removed from gameplay authority. Compatibility handling may accept old baseline fixtures during migration, but authoritative new snapshots expose executable Beam support uniformly.

## 4. Immutable Beam snapshot contract

`TacticalWeaponSpec` remains the immutable configured weapon snapshot and continues to carry:

- slot;
- `weapon_id`;
- weapon family;
- physical Count;
- mount type;
- canonical modifier IDs;
- ammo where relevant;
- firing arc;
- base MinDamage / MaxDamage from the weapon table;
- innate effect IDs;
- Tactical support mode.

`MinDamage` / `MaxDamage` remain **base damage**, not range- or target-dependent effective damage.

No runtime-derived damage value is persisted into `ShipDesignSpec` or `TacticalWeaponSpec`.

## 5. Base / configured / effective terminology

### 5.1 Base damage

**Base damage / Basisschaden** is the raw weapon-table MinDamage / MaxDamage.

It is independent of:

- mount;
- range;
- attacker/target;
- leader/race/crew context;
- combat Specials.

The catalog continues to display base damage.

### 5.2 Configured damage

**Configured damage** is a design-static preview under the frozen neutral context:

- player/ship-leader damage percent = 100;
- range damage modifier = 0;
- no runtime target;
- no deferred Slice-17.8 Special contribution.

Configured damage includes only design-static Beam-row effects that alter raw damage:

- Standard: 100%;
- Heavy: 150%;
- Point Defense: 50%;

using the original percentage rounding:

`configured = max(1, (base * percent + 50) / 100)`

and `configuredMax >= configuredMin`.

Other configured effects are displayed separately rather than folded into a fake damage number.

### 5.3 Effective runtime damage

**Effective damage** is contextual and belongs to a concrete Tactical/Lab evaluation. It includes, when active:

- mount damage adjustment;
- range damage dissipation or No Range Dissipation;
- player/ship-leader Beam damage percent;
- active dependency hooks such as High Energy Focus;
- the original rounding/order;
- target-side hooks where they affect the eventual application.

The ordinary Weapon Selector must not present one context-free effective damage number as universal truth.

## 6. Weapon Selector preview contract

The existing server-owned `MilitaryDesignWeaponRowPreview` remains the transport for a concrete candidate weapon row.

Gate 3 extends it with a server-derived Beam Tactical preview object. The JSON semantics are frozen as:

- `family = "beam"`;
- `base_min_damage`;
- `base_max_damage`;
- `configured_min_damage`;
- `configured_max_damage`;
- `shots_per_weapon` - 1 or 3 for Autofire;
- `range_profile` - `standard`, `heavy`, or `point_defense`;
- `static_to_hit_adjustment` - mount/modifier adjustment independent of attacker/target;
- `effect_ids` - canonical configured/innate Tactical effect identifiers needed for truthful display.

The preview is optional for non-Beam families.

The Selector must visually distinguish base damage from configured damage when different and explicitly display non-damage effects such as:

- Autofire shot behavior;
- Continuous To-Hit adjustment;
- No Range Dissipation;
- Armor Piercing;
- Shield Piercing;
- Enveloping;
- innate Always Hit / extra Structure / marine-kill facts where appropriate.

React performs no formula.

## 7. Frozen range and hit order

For each physical Beam unit:

1. obtain geometric Tactical range;
2. apply Heavy / Point Defense range transformation;
3. obtain the original range To-Hit adjustment;
4. apply Rangemaster dependency if later active;
5. apply innate doubled signed range To-Hit where present;
6. form attacker Beam Offense - target Beam Defense;
7. Autofire contributes -20 attack differential per inner shot;
8. Continuous contributes +25 attack differential;
9. Point Defense contributes +25 To-Hit through the frozen mount path;
10. consume `Random(100)`;
11. Mauler still consumes that RNG draw, then overwrites the raw roll to 100;
12. raw roll >95 normalizes to effective 100; otherwise add attack differential and cap at 100;
13. threshold = `min(95, 40 - rangeToHitModifier)`;
14. effective roll below threshold misses.

The server owns every step.

## 8. Frozen damage percentage and rounding order

For each physical Beam unit before its shot loop:

1. start Beam damage percent at 100;
2. add active High Energy Focus dependency +50 when later supplied by Slice 17.8;
3. add normal range damage modifier unless selected/innate No Range Dissipation suppresses it;
4. Point Defense subtracts 50 or Heavy adds 50;
5. combine player/ship-leader-derived percentage:

`combinedPercent = beamPercent + (playerLeaderPercent - 100)`

6. scale base min and max independently with original nearest-integer percentage rounding:

`scaled = (base * combinedPercent + 50) / 100`

7. clamp scaled minimum to at least 1;
8. clamp scaled maximum to at least scaled minimum;
9. on a hit, interpolate:

`damage = min + ((effectiveRoll - threshold) * (max-min+1)) / (100-threshold)`

using integer truncation at this final division and clamping to the scaled bounds.

Gate 3 must replace the current truncating baseline helper with this frozen shared derivation.

## 9. Physical Count, Autofire and RNG ordering

A configured Beam battery contains `Count` physical weapon units.

Without Autofire:

- total hit attempts = `Count`.

With Autofire:

- each physical unit resolves three inner shots;
- total hit attempts = `Count * 3`;
- order is unit1 shot1/2/3, unit2 shot1/2/3, etc.

Per inner shot:

1. consume hit RNG;
2. apply Mauler Always-Hit override after the draw;
3. resolve hit/miss;
4. if hit, compute damage;
5. consume downstream damage/internal-selection RNG for that shot;
6. finish all damage applications/effects for that shot;
7. only then advance to the next shot.

A miss consumes only the hit RNG required by the frozen path.

Rejected/unavailable/illegal commands consume zero RNG, readiness, state or events.

## 10. Normal activation readiness

A normal Beam command resolves the entire configured battery atomically.

- readiness is checked before command execution;
- one accepted command resolves all physical units / Autofire inner shots in frozen order;
- normal weapon readiness is consumed once after the whole accepted battery command;
- a command that fails validation before commit consumes no readiness;
- new round readiness reset remains authoritative server state.

This preserves command atomicity and deterministic replay.

## 11. Firing-arc legality

The persisted Slice-17.2A firing arc is enforced server-side before RNG is consumed.

Target projection and command preparation must reject a target outside the configured arc with:

- zero RNG consumption;
- zero readiness consumption;
- zero mutation;
- zero combat events.

React consumes projected legal targets only.

## 12. Damage application order

On a hit:

1. compute the single hit damage value;
2. apply target-side/dependency hooks at their frozen original positions;
3. dispatch one normal facing application, or Enveloping applications to facings 0,1,2,3;
4. every Enveloping facing receives a fresh copy of the same original hit damage;
5. `Apply_Ship_Damage` resolves Shield Piercing / Shield block / facing Shield;
6. resolve Armor Piercing / Armor-first selection;
7. resolve lower-layer Armor/Structure effects;
8. resolve innate post-layer effects such as Graviton extra StructureDamage and Neutron/Death marine loss;
9. destruction/winner resolution follows the fully resolved shot.

Enveloping does not carry residual damage from one facing into the next.

## 13. Minimal Tactical marine state

Neutron Blaster and Death Ray cannot be implemented truthfully without server-owned marine state.

Gate 3 adds:

### Immutable ship spec

- `MarineMax`

### Mutable Tactical ship state/view

- `MarineCurrent`

Base `MarineMax` by hull is frozen from the original hull table:

| Hull | Marines |
|---|---:|
| Frigate | 5 |
| Destroyer | 8 |
| Cruiser | 12 |
| Battleship | 20 |
| Titan | 30 |
| Doom Star | 50 |

`MarineCurrent` initializes to `MarineMax`.

For Neutron Blaster / Death Ray:

- use the current post-Shield damage at the original effect point;
- `marineKilled = damage / 5` with integer truncation;
- subtract saturating at zero;
- Damper Field hook may later suppress the effect.

Original `0xFF` sentinel behavior is not represented by magic 255 in normal MOOX combatants; future monster/Antaran handling must use explicit metadata.

### Troop Pods boundary

Original Troop Pods double the marine capacity at Tactical load time.

Because full combat-Special behavior belongs to Slice 17.8:

- 17.6 freezes an explicit marine-capacity Special dependency hook;
- Gate 3 consumes `troop_pods` only as a passive initial MarineMax multiplier of **2** so Neutron/Death resolution remains truthful for already-built ships;
- no boarding, shuttle, transport or active Troop Pods action is implemented in 17.6;
- all other Troop Pods behavior remains Slice 17.8.

This narrow dependency is allowed because it is an input to a 17.6-owned Beam damage effect, not an active Special action.

## 14. Combat-Special dependency hooks

The shared Beam calculator/damage pipeline exposes explicit neutral hooks for later Slice 17.8 consumers:

- Rangemaster Unit - range To-Hit input;
- High Energy Focus - Beam damage percent;
- Hard Shields - Shield Piercing cancellation;
- Heavy Armor / Xentronium / Achilles Targeting Unit - Armor/internal-selection behavior;
- Structural Analyzer - downstream post-Shield damage behavior;
- Damper Field - marine-kill suppression;
- Troop Pods - initial MarineMax multiplier.

Gate 3 may implement the hook shape and neutral/default path. It must not silently implement unrelated 17.8 Special behavior.

## 15. Beam Point Defense handoff

Point Defense has two frozen roles:

### 15.1 Normal ship-target Beam mode

It remains a legal ship-target Beam configuration with:

- transformed range index;
- +25 To-Hit;
- -50 damage percent;
- normal activation readiness behavior.

### 15.2 Reactive projectile capability

The Beam runtime exposes a server-authoritative reactive capability that Slice 17.7 can invoke for an incoming projectile actor.

The capability must preserve:

- configured Beam identity;
- Point Defense mount;
- canonical modifiers/innate effects;
- frozen Beam hit/RNG semantics;
- actor-coordinate targeting;
- restoration of normal activation readiness after the reaction.

Slice 17.7 owns whether/when a projectile is eligible and what interception means to projectile state.

## 16. Beam event contract

### 16.1 Physical shot event

Each authoritative physical/Autofire inner shot emits `beam_fired` with at least:

- source `ship_id`;
- `target_ship_id`;
- `weapon_slot`;
- `weapon_id`;
- `mount_type`;
- canonical `modifier_ids`;
- `physical_weapon_index`;
- `physical_weapon_count`;
- `burst_shot_index`;
- `burst_shot_count`;
- flattened `shot_index`;
- flattened `shot_count`;
- effective `range_index`;
- raw hit roll;
- effective hit roll;
- threshold;
- hit flag;
- computed hit damage.

Event fields are replay facts, not inputs back into gameplay.

### 16.2 Damage application event

`battle_damage_applied` remains the authoritative layer/result event and gains enough identity to correlate with the exact Beam shot and Enveloping application.

It includes or preserves:

- source/target/slot/weapon identity;
- flattened shot index/count;
- `application_index` / `application_count` for Enveloping;
- Shield facing and before/after;
- block / absorbed / overflow;
- Armor before/after;
- Structure before/after;
- optional `marine_before`, `marine_after`, `marine_killed` when the shot has a marine effect;
- existing damage/internal-selection evidence needed for replay/debug.

### 16.3 Battery aggregate

`beam_volley_resolved` represents the complete accepted battery command and includes at least:

- source/target/slot/weapon identity;
- total shot count;
- hit count;
- total direct Beam damage.

It does not replace the per-shot authority events.

## 17. Presentation / VFX contract

Beam presentation is event-driven only.

- hit and miss are visibly distinguishable;
- multiple physical/Autofire shots remain understandable as multiple authoritative shots/burst count;
- presentation may group timing visually, but may not collapse or invent gameplay results;
- Shield / Armor / Structure / marine feedback consumes authoritative result events;
- Beam identity may select MOOX-native visual treatment from `weapon_id`;
- browser timers never block or advance BattleSession;
- replayed events reconstruct the same semantic presentation facts even if animation timing differs.

## 18. Save/load/replay/Observer contract

Active Tactical persistence/replay must preserve:

- immutable Beam configuration;
- current weapon readiness;
- `MarineCurrent`;
- RNG state;
- event sequence;
- all mutable Shield/Armor/Structure state already frozen by 17.5.

Identical Tactical spec + initial RNG + command sequence must produce identical:

- hit/miss results;
- shot ordering;
- damage;
- marine changes;
- events;
- final result.

Observer/public views expose authoritative facts but no mutable authority.

## 19. Frozen deterministic vectors

Gate 3 and Gate 4 must encode the Gate-1 vectors as deterministic controls.

### Vector A - Laser / Standard rounding

- base 1-4;
- effective range index 2;
- damage percent 90;
- raw roll 100;
- expected scaled 1-4;
- threshold 50;
- damage 4.

### Vector B - Particle Beam / Heavy

- base 10-30;
- geometric index 4 -> Heavy index 2;
- damage percent 140;
- raw roll 80;
- expected scaled 14-42;
- threshold 50;
- damage 31;
- natural Shield Piercing.

### Vector C - Mauler / Heavy

- base 100;
- geometric index 4 -> Heavy index 2;
- signed range To-Hit -10 doubled to -20;
- damage percent 140;
- hit RNG still consumed;
- raw overwritten to 100;
- threshold 60;
- expected damage 140.

### Vector D - Laser / Point Defense ship target

- base 1-4;
- geometric index 2 -> PD index 4;
- To-Hit -30 +25 = -5;
- range damage -30 plus PD -50 => 20%;
- raw 50;
- expected scaled 1-1;
- threshold 45;
- damage 1.

### Vector E - Count 2 + Autofire

- physical Count 2;
- 3 shots per physical weapon;
- total 6 hit attempts;
- -20 attack differential on every attempt;
- order unit1 s1/s2/s3 then unit2 s1/s2/s3.

Selector/Combat Lab configured facts and Tactical results must come from the same shared Beam derivation.

## 20. Gate 3 implementation constraints

Gate 3 must proceed in recoverable checkpoints.

Recommended order:

1. shared Beam rule model + exact rounding + preview data;
2. generic Beam action/readiness/count/Autofire execution + arcs;
3. damage modifiers and innate Beam effects + minimal marine state;
4. PD reactive hook + event/VFX identity;
5. persistence/replay/Observer/integration/golden-vector regression pass.

Every checkpoint must keep the full Go/web baseline green before moving on.

## 21. Gate 2 close result

The Tactical Beam contract is frozen.

Gate 3 may implement exactly this contract.

It may not:

- move formulas into React;
- silently ignore an accepted Beam innate effect;
- silently substitute Laser behavior for another Beam;
- consume RNG on rejected commands;
- consume normal activation readiness during the reactive PD handoff;
- widen 17.6 into projectile lifecycle, boarding, Fighters or unrelated combat-Special mechanics.

Any new evidence that contradicts this freeze requires an explicit contract revision before implementation continues.
