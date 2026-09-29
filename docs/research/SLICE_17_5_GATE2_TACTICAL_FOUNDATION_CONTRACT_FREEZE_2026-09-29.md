# Slice 17.5 Gate 2 - Tactical Combat Foundation Contract Freeze

Date: 2026-09-29
Status: **Frozen for Gate 3 implementation**
Gate 1 evidence: `docs/research/SLICE_17_5_GATE1_TACTICAL_FOUNDATION_AUDIT_2026-09-28.md`
OPEN slice: `docs/slices/_OPEN_TACTICAL_COMBAT_FOUNDATION_SHIP_COMBATANT_BREADTH_2026-09-28.md`

## 1. Contract goal

Slice 17.5 replaces the remaining Slice-15.5 Tactical combatant fixture boundary with a broad, server-authoritative Tactical foundation for the accepted Slice-17.1 hull and mandatory-system matrix.

Gate 3 may remove the Frigate / Electronic / Titanium / no-Shield / Standard-Fuel guard only after the replacement projection and BattleSession validation described here exist.

This contract does **not** implement complete Beam, missile, torpedo, bomb, fighter or combat-Special behavior. Those remain owned by Slices 17.6, 17.7 and 17.8.

## 2. Authority and compatibility invariants

1. `BattleSession` remains the sole Tactical gameplay authority.
2. The browser never computes hit chance, damage, Shield absorption, readiness, legal targets or RNG.
3. Tactical RNG remains one ordered mutable server stream.
4. A rejected or unavailable action consumes **zero** RNG, changes no readiness/state and appends no gameplay event.
5. Prepared command application remains atomic: state, RNG and events become visible together only for a successful command.
6. `TacticalEvent{sequence, kind, seat_id?, command_sequence?, data?}` remains the common event envelope.
7. Existing Laser combat, movement, turning, wait/end-activation, retreat, Observer snapshots and event ordering remain regression contracts.
8. New fields are additive to existing snapshots/events wherever practical; Gate 3 must not gratuitously rename accepted Slice-15.5 fields.
9. Presentation and animation are consequences of authoritative state/events. Animation completion never advances gameplay.

## 3. Frozen Tactical combatant matrix

### 3.1 Hull durability

Base Armor and Structure are frozen directly from the local 1.31 executable:

| Hull | base Armor | Structure |
| --- | ---: | ---: |
| Frigate | 4 | 4 |
| Destroyer | 10 | 10 |
| Cruiser | 30 | 30 |
| Battleship | 50 | 50 |
| Titan | 80 | 80 |
| Doom Star | 150 | 150 |

`StructureMax` is the hull Structure value above.

### 3.2 Armor material durability

Installed Armor modifies the hull base Armor only:

| Armor | bonus | multiplier |
| --- | ---: | ---: |
| Titanium | 0% | 1x |
| Tritanium | 100% | 2x |
| Zortrium | 300% | 4x |
| Neutronium | 500% | 6x |
| Adamantium | 700% | 8x |
| Xentronium | 900% | 10x |

Frozen formula:

`ArmorMax = baseArmor + floor(baseArmor * armorBonusPercent / 100)`

Equivalent multipliers above are exact for the accepted rows.

Heavy Armor and Armor-Piercing interactions are **not** folded into this foundation.

### 3.3 Computer Beam Offense

| Computer | Beam Offense |
| --- | ---: |
| Electronic | 25 |
| Optronic | 50 |
| Positronic | 75 |
| Cybertronic | 100 |
| Moleculartronic | 125 |

These are the base Tactical targeting inputs. Later weapon/modifier/crew/race effects may extend them only in their owner slices.

### 3.4 Drive / combat-speed matrix

Values are ordered Frigate / Destroyer / Cruiser / Battleship / Titan / Doom Star:

| Drive | minimum / damage floor | empty-design maximum |
| --- | --- | --- |
| Nuclear | 10 / 8 / 6 / 5 / 4 / 3 | 20 / 18 / 16 / 15 / 14 / 13 |
| Fusion | 12 / 10 / 8 / 7 / 6 / 5 | 22 / 20 / 18 / 17 / 16 / 15 |
| Ion | 14 / 12 / 10 / 9 / 8 / 7 | 24 / 22 / 20 / 19 / 18 / 17 |
| Anti-Matter | 16 / 14 / 12 / 11 / 10 / 9 | 26 / 24 / 22 / 21 / 20 / 19 |
| Hyper | 18 / 16 / 14 / 13 / 12 / 11 | 28 / 26 / 24 / 23 / 22 / 21 |
| Interphased | 20 / 18 / 16 / 15 / 14 / 13 | 30 / 28 / 26 / 25 / 24 / 23 |

Frozen pristine design formula:

`minimumCombatSpeed = hullBaseSpeed + driveBonus`

where hull base speed is `10/8/6/5/4/3` and drive bonus is `0/2/4/6/8/10`.

`unusedSpace = clamp(hullSpace - spaceUsed, 0, hullSpace)`

`unusedSpaceBonus = clamp(ceil(10 * unusedSpace / hullSpace), 0, 10)`

`currentCombatSpeed = minimumCombatSpeed + unusedSpaceBonus`

Gate 3 initializes Tactical movement current/max from `currentCombatSpeed`.

The foundation preserves `minimumCombatSpeed` as the proven lower bound for future engine-damage consumers, but Slice 17.5 does not invent a new internal-engine damage system.

### 3.5 Base Beam Defense

Frozen base formula:

`BeamDefense = currentCombatSpeed * 5`

This is before later ECM, Inertial, crew, race or other owner-slice modifiers. Existing Inertial turning-mode behavior from Slice 17.4 remains intact; 17.5 does not silently add deferred Inertial defense bonuses.

### 3.6 Fuel Cells

Fuel Cells are Tactical-admission-neutral.

- `FuelCellID` may be carried in the immutable Tactical snapshot for truthful provenance / Observer display.
- Fuel choice does not change movement, initiative, damage, RNG or action availability in Slice 17.5.
- No accepted Fuel Cell may cause whole-encounter rejection.

## 4. Primary Shield contract

### 4.1 Shield strength and per-facing capacity

Shield strength / per-hit block:

| Shield | strength/block |
| --- | ---: |
| none | 0 |
| Class I | 1 |
| Class III | 3 |
| Class V | 5 |
| Class VII | 7 |
| Class X | 10 |

Hull factor:

| Hull | factor |
| --- | ---: |
| Frigate | 1 |
| Destroyer | 2 |
| Cruiser | 3 |
| Battleship | 4 |
| Titan | 5 |
| Doom Star | 7 |

Frozen capacity formula:

`shieldArcMax = shieldStrength * 5 * hullFactor`

Each installed Shield owns four independent facing pools. At Tactical load all four current pools initialize to `shieldArcMax`.

No-Shield initializes strength, max and all four pools to zero.

### 4.2 Facing mapping

Relative bearing is normalized to `0..359` degrees.

- `<=45` or `>315` -> arc 0
- `46..135` -> arc 3
- `136..225` -> arc 2
- `226..315` -> arc 1

The server determines the facing arc. The browser never chooses or recomputes it.

### 4.3 Shared Shield -> Armor -> Structure damage sequence

For normal foundation damage:

1. `incomingDamage` is the authoritative raw damage supplied by the weapon-family resolver.
2. `postBlockDamage = max(0, incomingDamage - shieldStrength)`.
3. The selected facing Shield pool absorbs `min(postBlockDamage, shieldArcCurrent)`.
4. Only Shield overflow continues to Armor.
5. Armor absorbs until `ArmorCurrent == 0`.
6. Remaining damage increments `StructureDamage`, capped at `StructureMax`.
7. The ship is destroyed when `StructureDamage >= StructureMax`.

Hard Shields, Multi-Phased Shields, Shield Capacitors and other Special hooks remain deferred to Slice 17.8. Armor-Piercing and family-specific bypass semantics remain owner-slice behavior.

### 4.4 Base Shield recharge

Base recharge budget is:

`rechargeBudget = floor(shieldArcMax * 3 / 10)`

The server walks arcs in stable order `0,1,2,3`, adding one point to a deficient arc per visit until the budget is exhausted or all four arcs are full.

MOOX round mapping for Gate 3 is frozen as one recharge operation at the deterministic round transition before the first activation of the new round. This is the MOOX BattleSession adaptation of the proven once-per-turn original recharge routine.

Slice 17.5 does not invent shield-generator subsystem damage. All admitted installed Shields start operational. A future subsystem-damage owner may gate recharge without changing the base capacity/block/recharge math.

## 5. Tactical snapshot / state contract

Gate 3 replaces hidden baseline substitutions with truthful derived data.

### 5.1 `TacticalShipSpec`

The immutable ship spec must continue to carry existing design provenance and add/derive at least:

- `HullID`
- `WarpDriveID`
- `ComputerID`
- `ArmorID`
- `ShieldID`
- `FuelCellID`
- `MinimumCombatSpeed`
- `CurrentCombatSpeed`
- `BeamOffense`
- `BeamDefense`
- `ArmorMax`
- `StructureMax`
- `ShieldStrength`
- `ShieldArcMax`
- immutable weapon and Special snapshots already accepted by earlier slices.

### 5.2 `TacticalShipState`

Mutable state must additionally own the four current Shield arc pools. Existing Armor, Structure, movement, readiness and destroyed state remain authoritative.

Observer/public Tactical views expose Shield identity plus current/max facing pools as detached data. They do not expose mutable aliases.

## 6. Replacement semantics for the old baseline-only guard

Gate 3 removes both protected fixture checks together:

- Game-layer `baselineCombatantUnsupportedReason`
- Battle-layer `validateBaselineCombatant`

They are replaced by broad structural/derived validation.

A legal Slice-17.1 combatant is admitted when:

1. its Hull/Drive/Computer/Armor/Shield/Fuel identities are known accepted rows;
2. derived durability/offense/defense/speed/Shield values equal this frozen contract;
3. weapon slots are unique, ascending and within `0..7`;
4. immutable weapon/Special metadata is structurally valid.

The encounter is **not** rejected because an installed later-family weapon lacks an implemented Tactical consumer.

Unknown/malformed identities, impossible derived values or malformed slots remain encounter-level errors.

No accepted combatant may be admitted while silently substituting Frigate, Electronic, Titanium, no-Shield, Nuclear/Fusion-only or zero-Defense values.

## 7. Generic weapon-slot action availability

### 7.1 Authoritative projection

For the currently active ship, the server projects exactly one action-availability descriptor per installed weapon slot.

Frozen logical fields:

- `ship_id`
- `weapon_slot`
- `weapon_id`
- `weapon_family` (from authoritative weapon kind)
- optional implemented `command_kind`
- `available`
- optional stable `unavailable_reason`
- authoritative target list when the action is available.

Existing `LegalFireActions` may remain as a compatibility projection during Gate 3, but generic action availability becomes the UI-facing architectural contract.

### 7.2 Base unavailable reasons

The foundation freezes these shared reasons:

- `tactical_consumer_unavailable`
- `weapon_not_ready`
- `no_legal_targets`

Owner slices may add family-specific reasons (for example arc/range/ammo conditions) without changing these meanings.

### 7.3 Unsupported later-family behavior

An installed weapon whose Tactical consumer is not implemented:

- remains present in `TacticalShipSpec` and the ship view;
- does not block encounter creation;
- receives `available=false`;
- receives `unavailable_reason=tactical_consumer_unavailable`;
- exposes no authoritative targets;
- consumes zero RNG;
- does not consume readiness;
- cannot be turned into a client-side pseudo-action.

Current supported Laser/Beam behavior remains executable and regression-green.

## 8. RNG and command sequencing

1. `TacticalState.RNGState` remains the sole mutable Tactical RNG state.
2. RNG draws occur only inside authoritative successful action resolution.
3. Structural validation, unsupported-action rejection and unavailable-action rejection consume zero draws.
4. Current Beam draw order remains a regression baseline until Slice 17.6 deliberately extends Beam behavior.
5. Child slices define the exact draw count/order for their family, but must use this shared RNG stream.
6. A successful command commits state, readiness, RNG state and its ordered events atomically under one command sequence.
7. Replay/Observer see the same resulting event order and public state.

## 9. Common damage/event envelope

### 9.1 Event envelope

The accepted envelope remains:

`TacticalEvent{sequence, kind, seat_id?, command_sequence?, data?}`

Child slices add family-specific event kinds/payload fields but do not fork the envelope or event sequence.

### 9.2 Shared damage event

`battle_damage_applied` remains the common lower-layer damage event.

Existing fields remain compatible where present. Gate 3 adds sufficient authoritative Shield data to represent the frozen sequence, including:

- source ship identity where available;
- target ship identity;
- source weapon slot;
- raw/incoming damage;
- selected Shield arc;
- Shield pool before/after;
- damage after Shield block;
- Shield absorbed amount / lower-layer overflow;
- Armor before/after;
- Structure damage before/after;
- existing shot/count context for Beam compatibility.

The numeric before/after fields are authoritative. A textual `layer` label is presentation metadata and must not become a client gameplay input.

## 10. Tactical UI / presentation contract

1. React consumes server-projected weapon actions and their availability.
2. Unsupported slots remain visible and explicitly disabled; the entire battle is not replaced by an unsupported screen.
3. UI target selection comes only from authoritative target lists.
4. Shield pools, Armor, Structure, movement and readiness are rendered from Tactical views.
5. Hit/damage/Shield calculations are never duplicated in TypeScript.
6. Effects/animations are driven by ordered Tactical events.
7. Animation completion cannot mutate BattleSession state or advance the activation.
8. Existing Laser/movement/retreat presentation remains a regression baseline.

## 11. Deterministic Combat Lab freeze

The following vectors are mandatory Gate-3/4 reference coverage.

### 11.1 Early tier - Nuclear / Electronic / Titanium / no Shield / Standard Fuel

| Hull | Structure | Armor | Speed min..max | Beam Offense | Beam Defense range |
| --- | ---: | ---: | --- | ---: | --- |
| Frigate | 4 | 4 | 10..20 | 25 | 50..100 |
| Destroyer | 10 | 10 | 8..18 | 25 | 40..90 |
| Cruiser | 30 | 30 | 6..16 | 25 | 30..80 |
| Battleship | 50 | 50 | 5..15 | 25 | 25..75 |
| Titan | 80 | 80 | 4..14 | 25 | 20..70 |
| Doom Star | 150 | 150 | 3..13 | 25 | 15..65 |

### 11.2 Mid tier - Ion / Positronic / Zortrium / Class III / Iridium Fuel

| Hull | Structure | Armor | Shield/arc | recharge | Speed min..max | Offense | Defense range |
| --- | ---: | ---: | ---: | ---: | --- | ---: | --- |
| Frigate | 4 | 16 | 15 | 4 | 14..24 | 75 | 70..120 |
| Destroyer | 10 | 40 | 30 | 9 | 12..22 | 75 | 60..110 |
| Cruiser | 30 | 120 | 45 | 13 | 10..20 | 75 | 50..100 |
| Battleship | 50 | 200 | 60 | 18 | 9..19 | 75 | 45..95 |
| Titan | 80 | 320 | 75 | 22 | 8..18 | 75 | 40..90 |
| Doom Star | 150 | 600 | 105 | 31 | 7..17 | 75 | 35..85 |

Class III per-hit block = 3.

### 11.3 Late tier - Interphased / Moleculartronic / Xentronium / Class X / Thorium Fuel

| Hull | Structure | Armor | Shield/arc | recharge | Speed min..max | Offense | Defense range |
| --- | ---: | ---: | ---: | ---: | --- | ---: | --- |
| Frigate | 4 | 40 | 50 | 15 | 20..30 | 125 | 100..150 |
| Destroyer | 10 | 100 | 100 | 30 | 18..28 | 125 | 90..140 |
| Cruiser | 30 | 300 | 150 | 45 | 16..26 | 125 | 80..130 |
| Battleship | 50 | 500 | 200 | 60 | 15..25 | 125 | 75..125 |
| Titan | 80 | 800 | 250 | 75 | 14..24 | 125 | 70..120 |
| Doom Star | 150 | 1500 | 350 | 105 | 13..23 | 125 | 65..115 |

Class X per-hit block = 10. Thorium remains Tactical-neutral.

### 11.4 Shared damage vectors

Gate 3 must additionally freeze tests for:

- no-Shield damage flows directly Armor -> Structure;
- Class III Frigate at full arc 15 receiving raw damage 20: block 3, Shield absorbs 15, exactly 2 reaches Armor;
- Class III Cruiser arc max 45 receiving raw damage 20: block 3, facing arc becomes 28, Armor unchanged;
- four empty Class III Cruiser arcs with recharge budget 13 and stable order become `[4,3,3,3]`;
- Shield overflow, Armor exhaustion and Structure destruction produce deterministic before/after event payloads.

### 11.5 Unsupported-action vector

A legal broad combatant carrying at least one not-yet-implemented later-family weapon must:

- enter BattleSession successfully;
- retain the slot metadata;
- expose that slot as unavailable with `tactical_consumer_unavailable`;
- keep supported slots usable;
- consume no RNG/readiness when the unavailable slot is selected/rejected.

## 12. Child-slice compatibility boundaries

### Slice 17.6 - Tactical Beams

Must consume:

- broad combatant projection;
- Beam Offense / Beam Defense foundation;
- generic action availability;
- shared RNG;
- Shield -> Armor -> Structure application;
- common event envelope.

It may add Beam/modifier-specific hit, range, arc, damage and effect rules, but may not fork those foundations.

### Slice 17.7 - Tactical Ordnance

Must consume the same combatant/action/RNG/damage/event foundations. Projectile/rack/ammo/flight/interception behavior is additive family logic.

### Slice 17.8 - Tactical Fighters / Combat Specials

Must consume the same foundations. Fighter launch/return and combat-Special consumers are additive. Existing Special metadata and Inertial turning modes remain compatible.

## 13. Gate 3 implementation order

1. Extend authoritative Tactical ruleset rows for all Hull/Armor/Computer/Drive/Shield/Fuel foundation data.
2. Extend Tactical immutable spec/state/view with truthful Shield/Fuel/minimum-speed fields.
3. Replace Game and Battle baseline combatant guards together.
4. Initialize broad durability, speed, offense, defense and Shield state from frozen rules.
5. Implement shared Shield -> Armor -> Structure application and deterministic recharge.
6. Add generic active-ship weapon action availability while preserving current Beam compatibility.
7. Extend common damage/event payloads additively.
8. Update Tactical UI to consume generic action availability and Shield views without client formulas.
9. Add 18 matrix vectors plus shared damage/unsupported-action persistence/replay/Observer regressions.
10. Only after those tests pass remove any remaining Slice-15.5 fixture assumptions.

## 14. Gate result

Gate 2 is complete when this contract, the Slice plan and the OPEN marker agree on the matrix, admission replacement, action availability, RNG, damage/event envelope, presentation boundary and child-slice ownership.

No production implementation is part of Gate 2.
