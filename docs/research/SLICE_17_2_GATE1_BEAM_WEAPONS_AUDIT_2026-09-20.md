# Slice 17.2 Gate 1 - Beam Weapons and Beam Mount Modifiers Audit

Date: 2026-09-20  
Result: **PASS for Gate 1 evidence/audit**  
Implementation status: **no broad Beam implementation in Gate 1**

## 1. Purpose

Slice 17.2 expands the accepted Slice-17.1 Ship Designer from the current single standard Laser baseline into the original Master of Orion II Beam-weapon breadth.

Gate 1 does not implement that breadth. It closes the evidence needed for Gate 2 to freeze:

- Beam identities and technology ownership;
- base damage, Space and PP cost;
- Beam miniaturization;
- original Beam modifier compatibility;
- original weapon-slot/count/snapshot semantics;
- current MOOX representation gaps;
- current Tactical consumers versus behavior that must remain explicitly deferred;
- durable Slice-17.0 reference-lab acceptance vectors.

The user-approved HMI direction is also preserved: the main Designer stays compact; broad weapon selection belongs in a dedicated dialog/sheet; installed weapon rows become the later configuration/modifier entry point.

## 2. Primary evidence

Primary/local original reference:

- legal MOO2 1.31 executable: `C:\ASH\Installation\mastori2\Orion2.exe`;
- local ignored HELP extraction: `reference/original/text/help/block_0000.ascii.txt`;
- normalized original technology provenance: `data/rulesets/moo2-1.31/technologies.json`.

Existing permanent evidence reused:

- `docs/research/TACTICAL_SHIP_COMBAT_BASELINE_2026-09-01.md`;
- `docs/research/SLICE_17_1_GATE1_HULLS_MANDATORY_SYSTEMS_AUDIT_2026-09-19.md`;
- `docs/research/SLICE_15_6_GATE3_WEAPON_SLOTS_BLOCK_A_CONTRACT_AMENDMENT_2026-09-14.md`.

The executable retains Watcom symbols, allowing the Gate-1 audit to anchor the actual data tables and functions instead of inferring values from prose alone.

## 3. Original weapon-table anchor

Watcom symbols identify:

- `_weapons`: LE object 2, offset `0x7803`;
- `_weapon_modifications`: LE object 2, offset `0x7D0B`.

The difference is:

`0x7D0B - 0x7803 = 0x508 = 46 * 0x1C`

Therefore the original weapon table contains **46 records of 0x1C / 28 bytes each**.

Existing Slice-07/Slice-15.5 evidence independently identifies weapon record index 3 as the standard Laser Cannon, which is a useful cross-check for the decoded layout.

### Directly proven weapon-record fields

Original functions close these fields:

| Record offset | Meaning | Direct consumer |
| ---: | --- | --- |
| `+0x0A` | technology ID | `One_Weapon_Space_`, miniaturization path |
| `+0x0C` | weapon type/kind | `Weapon_Type_` |
| `+0x0E` | base Space | `One_Weapon_Space_`, `Weapon_Space_` |
| `+0x10` | base PP Cost | `Weapon_Cost_` |
| `+0x14` | minimum damage | original Beam damage/display paths |
| `+0x16` | maximum damage | original Beam damage/display paths |
| `+0x1A` | allowed-modifier bit mask | `Determine_Beam_Weapon_Modifications_` |

The exact generic names of raw word `+0x12` and bytes `+0x18/+0x19` are **not frozen in Gate 1**. They are kept raw rather than given invented meanings. Their currently relevant player-facing effects are independently evidenced by original HELP and executable behavior.

## 4. Frozen Beam identity/base-data matrix

Filtering the original table to weapon type `0` (Beam) with a real technology ID yields **13 Beam weapons**.

| Tbl | Canonical ID | Tech | Field | Field RP | Space | PP | Damage | Mod mask |
| ---: | --- | ---: | ---: | ---: | ---: | ---: | --- | --- |
| 1 | `mass_driver` | 104 | 7 | 250 | 10 | 7 | 6 | `0x008E` |
| 2 | `gauss_cannon` | 78 | 64 | 2750 | 10 | 10 | 18 | `0x008A` |
| 3 | `laser_cannon` | 100 | 57 | 50 | 10 | 5 | 1-4 | `0x00BE` |
| 4 | `particle_beam` | 123 | 74 | 15000 | 15 | 35 | 10-30 | `0x0016` |
| 5 | `fusion_beam` | 70 | 31 | 150 | 10 | 6 | 2-6 | `0x0116` |
| 6 | `ion_pulse_cannon` | 97 | 41 | 900 | 30 | 15 | 2-10 | `0x0090` |
| 7 | `graviton_beam` | 79 | 16 | 1150 | 15 | 12 | 3-15 | `0x0012` |
| 8 | `neutron_blaster` | 115 | 54 | 900 | 10 | 8 | 3-12 | `0x0012` |
| 9 | `phasor` | 127 | 52 | 2000 | 10 | 10 | 5-20 | `0x00D6` |
| 10 | `disrupter_cannon` | 54 | 51 | 4500 | 20 | 25 | 40 | `0x0082` |
| 11 | `death_ray` | 47 | 74 | 15000 | 30 | 75 | 50-100 | `0x0002` |
| 12 | `plasma_cannon` | 137 | 59 | 3500 | 25 | 15 | 6-30 | `0x0012` |
| 27 | `mauler_device` | 105 | 39 | 6000 | 50 | 75 | 100 | `0x0002` |

Notes:

- `disrupter_cannon` is the existing normalized technology ID spelling and must remain stable unless a separately planned migration renames it.
- Technology field 74 is a self-loop field with RP cost 15000 and has no normalized category label. Gate 1 does not invent one.
- Fixed-damage weapons have equal minimum/maximum values in the original table.

## 5. Original innate Beam behavior metadata

Original technology HELP gives the following player-facing behavior. This is evidence metadata; it does **not** mean all effects are implemented in current Tactical runtime.

| Beam | Original-evidenced innate behavior relevant to 17.2 |
| --- | --- |
| Mass Driver | 6 damage; damage not reduced by range |
| Gauss Cannon | 18 damage regardless of range |
| Laser Cannon | normal 1-4 Beam damage |
| Particle Beam | 10-30; naturally shield-piercing |
| Fusion Beam | normal 2-6 Beam damage |
| Ion Pulse Cannon | 2-10; bypasses armor/structure systems path; ineffective against monsters/Antaran ships |
| Graviton Beam | 3-15; penetrating damage adds extra structural damage |
| Neutron Blaster | 3-12 to shields; penetrating internal damage kills marines |
| Phasor | normal 5-20 Beam damage |
| Disrupter Cannon | 40; damage not reduced by range |
| Death Ray | 50-100; penetrating internal damage kills marines |
| Plasma Cannon | 6-30 to all four shield facings; doubled range penalties |
| Mauler Device | always hits; 100 damage; doubled range dissipation |

These properties must be represented explicitly if Gate 2 freezes them into design/tactical metadata. They must not be reconstructed later from display strings.

## 6. Beam modifier table and exact compatibility bits

The original `_weapon_modifications` table begins at object-2 `0x7D0B` with 15-byte records.

For the Beam modifier family, executable table values and original HELP align on the following modifier identities:

| Bit | Mask | Modifier | Refinement / mini level | Space delta | Cost delta |
| ---: | ---: | --- | ---: | ---: | ---: |
| 1 | `0x0002` | Heavy Mount | 0 | +100% | +100% |
| 2 | `0x0004` | Point Defense | 0 | -50% | -50% |
| 3 | `0x0008` | Armor Piercing | 1 | +50% | +50% |
| 4 | `0x0010` | Continuous | 1 | +50% | +50% |
| 5 | `0x0020` | No Range Penalty / no range dissipation | 1 | +25% | +25% |
| 6 | `0x0040` | Shield Piercing | 1 | +50% | +50% |
| 7 | `0x0080` | Autofire | 2 | +50% | +50% |
| 8 | `0x0100` | Enveloping | 2 | +100% | +100% |

Heavy Mount and Point Defense are the original regular mount choices; the remaining entries are special Beam modifiers unlocked by refinement/miniaturization level.

`Weapon_Cost_` and `Weapon_Space_` apply modifier deltas **before** final miniaturization.

### Per-Beam executable compatibility matrix

The executable `+0x1A` mask is authoritative for selection legality.

| Beam | Heavy | PD | AP | Continuous | No Range | Shield Piercing | Autofire | Enveloping |
| --- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| Mass Driver | Y | Y | Y | - | - | - | Y | - |
| Gauss Cannon | Y | - | Y | - | - | - | Y | - |
| Laser Cannon | Y | Y | Y | Y | Y | - | Y | - |
| Particle Beam | Y | Y | - | Y | - | - | - | - |
| Fusion Beam | Y | Y | - | Y | - | - | - | Y |
| Ion Pulse Cannon | - | - | - | Y | - | - | Y | - |
| Graviton Beam | Y | - | - | Y | - | - | - | - |
| Neutron Blaster | Y | - | - | Y | - | - | - | - |
| Phasor | Y | Y | - | Y | - | Y | Y | - |
| Disrupter Cannon | Y | - | - | - | - | - | Y | - |
| Death Ray | Y | - | - | - | - | - | - | - |
| Plasma Cannon | Y | - | - | Y | - | - | - | - |
| Mauler Device | Y | - | - | - | - | - | - | - |

### Death Ray HELP discrepancy

Normalized original HELP says Death Ray supports Heavy Mount and Continuous, but the executable allowed-modifier mask is `0x0002`, i.e. Heavy Mount only.

For authoritative design legality, Gate 1 chooses the **executable mask** over conflicting HELP prose. The discrepancy remains documented instead of silently “fixing” the original.

Natural/innate properties are separate from selectable modifier bits:

- Particle Beam is naturally shield-piercing but does not offer the selectable Shield Piercing modifier.
- Mass Driver/Gauss/Disrupter have innate range behavior and therefore do not need the selectable No Range modifier to obtain that behavior.

## 7. Beam miniaturization

The original call path is direct:

`Weapon_Cost_` / `Weapon_Space_`
-> modifier delta
-> `Miniaturization_Category_`
-> `Tech_Level_`
-> `Cost_Given_M_Level_` / `Space_Given_M_Level_`.

### Cost

The already-direct Slice-17.1 audit closed the universal positive cost reducer:

| Mini level | Cost factor |
| ---: | ---: |
| 0 | 100% |
| 1 | 75% |
| 2 | 55% |
| 3 | 40% |
| 4 | 30% |
| 5+ | 25% |

`mini_cost = max(1, floor(modified_base_cost * factor / 100))`

### Beam Space class

Direct evaluation of the original `Miniaturization_Category_` path for all 13 Beam technology IDs returns **Space class 1**.

Therefore Beam Space uses:

| Mini level | Beam Space factor |
| ---: | ---: |
| 0 | 100% |
| 1 | 80% |
| 2 | 70% |
| 3 | 60% |
| 4 | 50% |
| 5+ | 40% |

The exact positive-Space rule remains:

`mini_space = max(1, floor((modified_base_space * factor_per_thousand + 500) / 1000))`

This is nearest-integer rounding with positive half boundaries upward.

This is intentionally **different** from Shield Space class 0 in Slice 17.1, whose later-level factors are 65/50/35/25%.

## 8. Original weapon-array / slot semantics

Original accessor functions close the design-array structure used by Beam weapons:

- `Design_Weapon_` indexes a weapon row with **8-byte stride** and reads the weapon ID/index word;
- `Design_Weapon_Count_` reads the count byte from that row;
- `Design_Weapon_Specials_` reads a modifier/special word from that same row;
- original design/update code explicitly iterates/compares against **8 weapon rows**.

Thus an original design has eight weapon rows, each preserving at least:

- weapon identity;
- quantity;
- firing-arc/ammo-related row data;
- modifier/special mask.

The current accepted MOOX Block-A baseline intentionally models:

`ShipWeaponMount { Slot, WeaponID, Count }`

with:

- slots `0..7`;
- maximum 8 mount rows;
- strictly ascending unique slots;
- positive physical count;
- one mount row = one independent Tactical firing group.

That existing slot/count shape is compatible with unmodified Beam identity/count.

### Required Gate-2 schema change

The current `ShipWeaponMount` is **not sufficient** for 17.2 modifiers because it has no authoritative modifier field.

Gate 2 must freeze an explicit modifier representation in the immutable design/built-Ship snapshot. It must **not** encode modifiers into `WeaponID` strings or infer them from UI state.

The exact Go field shape remains a Gate-2 decision; Gate 1 only closes the need for explicit lossless modifier metadata.

## 9. Current MOOX implementation gap

Current rules/runtime remain intentionally Laser-only:

- `data/rulesets/moo2-1.31/tactical_combat.json` exposes one `weapon` record: `laser_cannon`;
- `MilitaryDesignerCatalog` projects that one Tactical weapon;
- `compileMilitaryDesignSpec` explicitly rejects any mount whose `WeaponID` is not the supported Laser;
- the current compiler adds Laser base Space/Cost without weapon miniaturization/modifier metadata;
- `TacticalWeaponSpec` contains only `Slot, WeaponID, Count, MinDamage, MaxDamage`.

This is a clean Gate-1 finding, not a bug in the accepted 17.1 baseline.

## 10. Tactical support/defer matrix

The current Tactical runtime has one generic **normal Beam** hit/damage path:

- global normal Beam range-to-hit table;
- global normal Beam range-damage table;
- attacker Beam Attack versus defender Beam Defense;
- per-weapon min/max damage;
- deterministic firing groups by slot/count.

It does **not** currently carry per-weapon innate flags or selectable modifier metadata into `TacticalWeaponSpec`.

Therefore Gate 1 classifies support as follows:

### Existing/bounded consumer

- standard Laser Cannon, no modifiers: implemented and regression-covered;
- generic min/max Beam damage is structurally reusable once a broader catalog is frozen.

### Not yet supported and must not be silently simulated

Selectable modifiers:

- Heavy Mount;
- Point Defense;
- Armor Piercing;
- Continuous;
- No Range Penalty;
- Shield Piercing;
- Autofire;
- Enveloping.

Innate/special Beam behavior:

- innate no-range-dissipation families;
- natural shield piercing;
- armor/structure bypass;
- extra structure damage;
- marine-kill effects;
- all-four-shield damage;
- doubled range penalties;
- always-hit behavior;
- monster/Antaran immunity exceptions.

Gate 2 must decide which of these effects are bounded enough to implement in 17.2 Tactical and which remain explicit metadata/deferred behavior. Design legality/snapshot support can be broader than Tactical consumption as long as every feature carries a clear support/defer status.

## 11. Slice-17.0 reference-lab vectors

The durable Triangle profiles give deterministic technology surfaces for 17.2.

### Baseline / Average

Reference Triangle initialization uses Average starting technologies:

- always-known field 0;
- staged fields 29, 55, 22, **57**, 28, 23.

Among the frozen Beam matrix this exposes:

- Laser Cannon (field 57).

Expected Gate-3 vector:

- one legal baseline Beam;
- Laser cost/Space at mini level 0 before modifiers;
- Heavy/Point-Defense regular mount choices can be tested separately once Gate 2 freezes mount semantics.

### Mid-Tech

The existing deterministic Mid-Tech field set includes Beam fields:

- 7, 16, 31, 41, 54, 57.

Expected Beam identities:

- Mass Driver;
- Laser Cannon;
- Fusion Beam;
- Ion Pulse Cannon;
- Graviton Beam;
- Neutron Blaster.

This is the primary modifier/miniaturization integration laboratory because it contains multiple normal and innate-special Beam families without granting the entire tree.

### All-Tech

The All-Tech profile grants concrete fields 0..74.

Expected Beam identities:

- all 13 frozen Beam weapons.

This is the required late-game Designer matrix for Phasor, Plasma Cannon, Disrupter Cannon, Mauler Device, Gauss Cannon, Particle Beam and Death Ray without research grinding.

## 12. HMI data implications for Gate 2

The audit supports the user-requested weapon-selection dialog without inventing data.

Useful filter/sort fields are now factual rather than speculative:

- available / locked;
- weapon name;
- technology field/tier;
- min/max or fixed damage;
- base/current Space;
- base/current PP Cost;
- innate behavior tags;
- allowed modifier tags;
- miniaturization/refinement level.

Recommended interaction grammar remains:

1. compact installed-weapon list in the main Designer;
2. `+` opens a dedicated weapon selector;
3. selector can use search/filter/sort over the frozen Beam data;
4. selecting a Beam creates/updates an installed mount;
5. clicking an installed mount opens weapon configuration/modifiers;
6. Special Systems later reuse the same high-level dialog grammar but remain Slice 17.4 scope.

Gate 1 does not freeze exact tabs, filter widgets or desktop/mobile dialog layout. That belongs in the Gate-2 HMI contract after the data model is frozen.

## 13. Gate-1 conclusions

Gate-1 checklist result:

- [x] Inventory normalized Beam weapons and missing original data.
- [x] Re-audit technology, damage, Space, Cost, range and miniaturization.
- [x] Re-audit modifier availability/compatibility per Beam.
- [x] Re-audit weapon slot/count limits and original design-array semantics.
- [x] Map supported Beam metadata onto current `ShipWeaponMount` without lossy hacks.
- [x] Identify Tactical consumers already implemented versus deferred.

### Gate-2 inputs now closed

Gate 2 can freeze:

- exactly 13 Beam identities;
- exact technology IDs/fields;
- exact base Space/Cost and min/max damage;
- Beam Space miniaturization class 1;
- exact cost/Space miniaturization arithmetic;
- exact 8-bit Beam modifier compatibility family;
- exact per-Beam allowed modifier masks;
- original 8-row weapon design-array semantics;
- requirement for explicit modifier snapshot data;
- Baseline/Mid/All-Tech acceptance matrices;
- explicit Tactical support/defer boundary.

### Residual non-blocking research notes

The generic names/meaning of weapon-record raw `+0x12` and `+0x18/+0x19` remain intentionally unresolved. Gate 2 must not depend on guessed names for them.

Death Ray HELP conflicts with executable modifier legality on Continuous. The executable mask is authoritative for design legality.

No Slice-17.2 gameplay implementation was performed in Gate 1.

**Gate 1 result: PASS. Gate 2 is next but not started.**
