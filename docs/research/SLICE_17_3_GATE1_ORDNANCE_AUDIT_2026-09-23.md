# Slice 17.3 Gate 1 - Missiles, Torpedoes and Bomb Ordnance Audit

Date: 2026-09-23

Status: **Gate 1 complete; Gate 2 contract freeze ready / not started**

Parent: `docs/slices/PLANNED_17_3_MISSILES_TORPEDOES_BOMBS_ORDNANCE.md`

Open marker: `docs/slices/_OPEN_MISSILES_TORPEDOES_BOMBS_ORDNANCE_2026-09-23.md`

## 1. Gate-1 conclusion

The original MOO2 1.31 executable contains a contiguous ordnance block immediately after the 13 Beam rows:

- 4 Missile rows;
- 3 Torpedo rows;
- 6 Bomb/Biological-bomb rows.

Gate 1 closes the design-side evidence required to freeze Slice 17.3 in Gate 2. It does **not** implement ordnance and does not pull projectile/interception/bombardment simulation forward from planned Slice 17.7.

The most important schema finding is that original Ship Design weapon rows carry **Count and Ammo as separate values**. MOOX currently persists Count but has no ordnance Ammo/Rack field. Gate 2 therefore must freeze an explicit persistent ammo/rack representation before Gate 3 can implement missiles honestly.

## 2. Primary evidence and method

Primary original source:

- `reference/original/support/files/Orion2.exe`;
- normalized original HELP/technology evidence in `data/rulesets/moo2-1.31/technologies.json`;
- the existing 17.2 executable/table audit in `docs/research/SLICE_17_2_GATE1_BEAM_WEAPONS_AUDIT_2026-09-20.md`.

Watcom symbols in Orion2.exe directly anchor:

- `_weapons`: LE object 2, offset `0x7803`;
- `_weapon_modifications`: LE object 2, offset `0x7D0B`;
- `Weapon_Cost_`: object-1 `0x5EC74` / VA `0x6EC74`;
- `Weapon_Needs_Racks_`: object-1 `0x5EDAE` / VA `0x6EDAE`;
- `One_Weapon_Space_`: object-1 `0x5EDC6` / VA `0x6EDC6`;
- `Weapon_Space_`: object-1 `0x5EE8E` / VA `0x6EE8E`;
- `Weapon_Type_`: object-1 `0x5EFEB` / VA `0x6EFEB`;
- `Weapon_Mod_String_`: object-1 `0x5F07F` / VA `0x6F07F`;
- `Miniaturization_Category_`: object-1 `0x5F11C` / VA `0x6F11C`;
- `Weapon_Rack_Size_To_Value_`: object-1 `0x5F1CC` / VA `0x6F1CC`;
- `Mod_Allowed_On_Weapon_Type_`: object-1 `0x5678C` / VA `0x6678C`;
- `Design_Weapon_`: object-1 `0x67D69` / VA `0x77D69`;
- `Design_Weapon_Ammo_`: object-1 `0x67D8E` / VA `0x77D8E`;
- `Design_Weapon_Count_`: object-1 `0x67DB2` / VA `0x77DB2`;
- `Design_Weapon_Firing_Arc_`: object-1 `0x67DD6` / VA `0x77DD6`;
- `Design_Weapon_Specials_`: object-1 `0x67DFA` / VA `0x77DFA`;
- `Missile_Speed_`: object-1 `0x2CD21` / VA `0x3CD21`;
- `Get_Missile_Hits_`: object-1 `0x2E095` / VA `0x3E095`.

Temporary disassembly tooling was kept outside the repository under `C:\ASH\Temp`; no analysis dependency is added to MOOX.

## 3. Original weapon-record structure reused by 17.3

The 17.2 audit already proved the 46-record, 28-byte `_weapons` table and these generic fields:

| Record offset | Meaning |
| ---: | --- |
| `+0x0A` | Technology ID |
| `+0x0C` | weapon type |
| `+0x0E` | base Space |
| `+0x10` | base PP Cost |
| `+0x14` | minimum/base damage |
| `+0x16` | maximum damage |
| `+0x1A` | per-weapon allowed-modifier mask |

For Gate 1, raw word `+0x12` and bytes `+0x18/+0x19` remain raw unless an effect is independently evidenced. The new audit does not invent generic names for them.

Weapon types relevant here are directly identified by the executable:

- type `1` = Missile;
- type `2` = Torpedo;
- type `3` = Bomb/Biological bomb.

## 4. Frozen Gate-1 ordnance identity/base-data matrix

Direct read of original weapon rows 14-26:

| Tbl | Canonical ID | Tech | Field | RP | Type | Base Space | Base PP | Damage | Raw +0x12 | Raw +0x0D | Record mask |
| ---: | --- | ---: | ---: | ---: | --- | ---: | ---: | --- | ---: | ---: | --- |
| 14 | `nuclear_missile` | 121 | 22 | 50 | Missile | 0 | 0 | 8 | 0 | 5 | `0x3E00` |
| 15 | `merculite_missile` | 106 | 2 | 650 | Missile | 0 | 0 | 14 | 0 | 5 | `0x3E00` |
| 16 | `pulson_missile` | 149 | 47 | 1150 | Missile | 0 | 0 | 20 | 0 | 5 | `0x3E00` |
| 17 | `zeon_missile` | 202 | 50 | 4500 | Missile | 0 | 0 | 30 | 0 | 5 | `0x3E00` |
| 18 | `anti_matter_torpedoes` | 12 | 13 | 2000 | Torpedo | 20 | 15 | 25 | 0 | 2 | `0x5500` |
| 19 | `proton_torpedoes` | 146 | 38 | 4500 | Torpedo | 30 | 20 | 40 | 0 | 2 | `0x4500` |
| 20 | `plasma_torpedoes` | 139 | 40 | 10000 | Torpedo | 40 | 75 | 120 | 2048 | 2 | `0x4520` |
| 21 | `nuclear_bomb` | 119 | 55 | 50 | Bomb | 5 | 3 | 3-12 | 0 | 10 | `0x0000` |
| 22 | `fusion_bomb` | 71 | 5 | 250 | Bomb | 7 | 5 | 4-24 | 0 | 10 | `0x0000` |
| 23 | `anti_matter_bomb` | 10 | 13 | 2000 | Bomb | 7 | 6 | 5-40 | 0 | 10 | `0x0000` |
| 24 | `neutronium_bomb` | 118 | 40 | 10000 | Bomb | 10 | 9 | 10-60 | 0 | 10 | `0x0000` |
| 25 | `death_spores` | 48 | 1 | 400 | Bio bomb | 5 | 5 | 10 | 128 | 10 | `0x0000` |
| 26 | `bio_terminator` | 28 | 17 | 4500 | Bio bomb | 7 | 8 | 20 | 128 | 10 | `0x0000` |

Notes:

- Missile record Space/PP being zero is intentional. Missile rack size is priced separately by the original cost/space functions.
- `death_spores` and `bio_terminator` are original type-3 weapon rows and therefore belong to the design-side bomb-family audit, while their population/diplomatic bombardment effects remain downstream.
- Normalized technology metadata marks the four conventional bombs/three torpedoes/four missiles strategic-combat-available. Death Spores and Bio Terminator are marked false; Gate 1 treats that as runtime-context metadata, not a reason to erase their original design rows.

## 5. Original 8-byte Ship Design weapon row

The accessor functions close the original row layout used by every installed weapon row.

Relative to the row base:

| Row offset | Meaning | Original accessor |
| ---: | --- | --- |
| `+0`.. `+1` | weapon index/ID word | `Design_Weapon_` |
| `+2` | Count byte | `Design_Weapon_Count_` |
| `+4` | Firing Arc byte | `Design_Weapon_Firing_Arc_` |
| `+5`.. `+6` | Specials / modifier mask word | `Design_Weapon_Specials_` |
| `+7` | Ammo byte | `Design_Weapon_Ammo_` |

The row stride is exactly 8 bytes.

The byte at `+3` is not needed by the current Gate-1 contract and is not assigned an invented meaning.

### Design consequence

`Count` and `Ammo` are **not the same field**.

For Slice 17.3:

- Count remains the number of installed weapon/rack units represented by the current MOOX `ShipWeaponMount.Count`;
- Missile Ammo/Rack Size requires a separate persistent field;
- Tactical projectile instances, remaining in-flight missiles and interception state are **not** design-snapshot fields and remain Slice 17.7 runtime state.

Firing Arc is proven as an original structural row field. Gate 2 must explicitly decide whether Slice 17.3 supports an ordnance firing-arc choice or preserves a fixed/default arc; it must not silently overload Ammo/Count or invent browser-only arc state.

## 6. Missile rack / ammo Space and Cost

`Weapon_Needs_Racks_` returns true exactly for weapon type 1 (Missile).

The original rack values accepted by `Weapon_Rack_Size_To_Value_`, `Weapon_Space_` and `Weapon_Cost_` are:

| Design Ammo/Rack Size | Rack value used in Space/Cost base |
| ---: | ---: |
| 2 | 10 |
| 5 | 20 |
| 10 | 30 |
| 15 | 35 |
| 20 | 40 |

Important:

- these values are part of the **design** formula, not the number of live projectile entities;
- Missile weapon records intentionally contribute base Space/PP `0/0`;
- the rack value is folded into the original Space/Cost path before final miniaturization;
- `One_Weapon_Space_` has a Missile-specific minimum-rack path and must not be substituted for the authoritative full-row `Weapon_Space_` calculation in saved-design validation.

Gate 2 must freeze one server-owned quote formula around these original functions and use it for catalog preview, save validation and immutable construction snapshots.

## 7. Miniaturization

All 13 ordnance Technology IDs directly evaluate through original `Miniaturization_Category_` to **Space class 1**.

Therefore the same class-1 Space factors already frozen for Beams apply:

| Mini level | Space factor |
| ---: | ---: |
| 0 | 100% |
| 1 | 80% |
| 2 | 70% |
| 3 | 60% |
| 4 | 50% |
| 5+ | 40% |

The universal positive Cost factors remain:

| Mini level | Cost factor |
| ---: | ---: |
| 0 | 100% |
| 1 | 75% |
| 2 | 55% |
| 3 | 40% |
| 4 | 30% |
| 5+ | 25% |

As in 17.2:

- modifier/rack design deltas are applied before final miniaturization;
- final positive values must preserve the original minimum-one behavior;
- Gate 2 must reuse the already-audited miniaturization implementation instead of introducing a parallel approximate formula.

## 8. Ordnance modifier table

The original 15-record `_weapon_modifications` table uses record index as the modifier bit position.

The ordnance-relevant records are:

| Bit | Mask | Modifier | Required mini level | Space delta | Cost delta | Evidence |
| ---: | ---: | --- | ---: | ---: | ---: | --- |
| 9 | `0x0200` | MIRV | 2 | +100% | +100% | only remaining Missile-family modifier after direct consumers below |
| 10 | `0x0400` | ECCM | 1 | +25% | +25% | shared Missile/Torpedo legal modifier; HELP family intersection |
| 11 | `0x0800` | Armored / Heavily Armored | 1 | +25% | +25% | `Get_Missile_Hits_` tests high-byte `0x08` and doubles missile hits |
| 12 | `0x1000` | Fast | 1 | +25% | +25% | `Missile_Speed_` tests high-byte `0x10` and adds +4 speed |
| 13 | `0x2000` | Emissions Guidance | 0 | +300% | +300% | original modifier record 13 directly encodes +300% Space and +300% Cost |
| 14 | `0x4000` | Overloaded | 1 | +50% | +50% | Torpedo modifier after Enveloping/ECCM intersection |

Existing shared Beam-era bits still apply where the Torpedo family allows them:

- `0x0020` = No Range Dissipation;
- `0x0100` = Enveloping.

### Family legality mask

`Mod_Allowed_On_Weapon_Type_` independently returns:

| Weapon family | Family modifier mask |
| --- | ---: |
| Beam | `0x01FE` |
| Missile | `0x3E00` |
| Torpedo | `0x6520` |
| Bomb/Bio bomb | no normal weapon modifiers |

Effective legality must be the intersection of the family mask and the per-weapon record mask.

This resolves a useful original-data quirk:

- Anti-Matter Torpedo record mask is `0x5500`, which contains the `0x1000` Fast bit;
- Torpedo family mask `0x6520` does **not** contain `0x1000`;
- effective Anti-Matter Torpedo legality is therefore `0x4500`: Enveloping + ECCM + Overloaded, exactly matching original HELP.

Per-weapon effective modifier sets:

| Weapon | Effective design modifiers |
| --- | --- |
| all 4 Missiles | MIRV, ECCM, Armored, Fast, Emissions Guidance |
| Anti-Matter Torpedo | Enveloping, ECCM, Overloaded |
| Proton Torpedo | Enveloping, ECCM, Overloaded |
| Plasma Torpedo | No Range Dissipation, Enveloping, ECCM, Overloaded |
| Bomb/Bio bomb rows | none |

## 9. Player-facing innate/runtime metadata to hand downstream

These are original-evidenced behaviors needed for honest metadata/handoff. They are **not** implemented in Slice 17.3.

### Missiles

- Nuclear Missile: 8 damage.
- Merculite Missile: 14 damage.
- Pulson Missile: 20 damage.
- Zeon Missile: 30 damage.
- `Missile_Speed_` gives the four Missile rows base combat speed 12; Fast adds +4.
- MIRV, guidance/ECCM, armored-hit behavior, targeting, interception and impact are Slice 17.7 consumers.

### Torpedoes

Original HELP:

- Anti-Matter Torpedoes: speed 20, 25 damage, fire every other turn; Overloaded/Enveloping/ECCM.
- Proton Torpedoes: 40 damage, trans-warp/same-turn impact semantics, fire every other turn, max range 24; Overloaded/Enveloping/ECCM.
- Plasma Torpedoes: speed 24, 120 starting strength/damage, loses 5 strength per square, fires every other turn; Overloaded/Enveloping/ECCM/No Range Dissipation.

All projectile movement/readiness/impact semantics remain Slice 17.7.

### Bombs / biological weapons

Original HELP identifies conventional bombs as orbital bombardment devices:

- Nuclear Bomb: 3-12.
- Fusion Bomb: 4-24.
- Anti-Matter Bomb: 5-40.
- Neutronium Bomb: 10-60.

Death Spores and Bio Terminator are biological bombardment weapons. Their population kill/diplomatic effects are not Ship Designer formulas and remain in the downstream bombardment/combat owner.

## 10. Current MOOX representation gap

### Ruleset

Current `data/rulesets/moo2-1.31/tactical_combat.json` exposes the 13 Beam weapons from Slice 17.2 and Beam modifiers only. No Missile/Torpedo/Bomb catalog is normalized yet.

### Core ShipDesignSpec

Current `core.ShipWeaponMount` persists:

- Slot;
- WeaponID;
- Count;
- MountType;
- ModifierIDs.

It does **not** persist:

- Missile Ammo/Rack Size;
- Firing Arc.

Therefore current Core shape is not lossless for original Missile design rows.

### Compile / preview

Current military compile and preview paths call `quoteBeamWeapon(...)` for installed weapons. They cannot honestly price/validate ordnance yet.

### Battle handoff

Current `battle.TacticalWeaponSpec` carries:

- Slot;
- WeaponID;
- Count;
- MountType;
- ModifierIDs;
- MinDamage/MaxDamage;
- innate-effect IDs;
- TacticalSupport.

It has no Ammo/Rack field and no projectile/readiness state.

Gate 2 must freeze the immutable design/built-Ship/BattleSession ordnance metadata shape. Gate 3 may populate that metadata. Slice 17.7 owns live projectile/ammo-consumption/interception state.

## 11. Design-time versus Tactical ownership

### Slice 17.3 owns

- identity/category;
- Technology unlock;
- Count;
- Missile Ammo/Rack Size;
- accepted firing-arc policy if Gate 2 includes it;
- legal modifier IDs;
- Space/PP quote;
- miniaturization level;
- immutable saved/built Ship snapshot;
- immutable BattleSession handoff metadata;
- explicit tactical support/defer classification.

### Slice 17.7 owns

- projectile entities/flight;
- missile/torpedo speed during battle;
- target tracking/retargeting;
- hit/evasion/ECCM resolution;
- Point Defense and interception;
- missile hit points and destruction;
- MIRV warhead behavior;
- Emissions Guidance target-system behavior;
- Torpedo cooldown/readiness;
- Plasma range dissipation;
- impact/explosion/damage application;
- bomb/bombardment combat context.

Fast Missile Racks is a separate ship-special Technology, not one of the Missile weapon-row modifier bits. Its two-volley/reload behavior belongs to the later combat-special/ordnance interaction owner and must not be encoded as a Missile modifier.

## 12. Reference-lab vectors for Gate 2/3

The existing Slice-17.0 harness remains authoritative.

### Baseline - `game-triangle-2pc`

Purpose:

- prove ordinary New Game lock/unlock behavior is not changed by 17.3;
- verify earliest ordnance catalog entries appear only when their ordinary starting/research state permits them;
- use Nuclear Missile / Nuclear Bomb as the first candidate pair where the baseline state has their technologies;
- prove unsupported/unresearched later Missiles/Torpedoes/Bombs remain locked.

No debug-only free-tech logic is added to baseline.

### Mid-Tech - `game-triangle-mid-tech`

The existing mid profile includes fields needed for:

- Nuclear Missile;
- Merculite Missile;
- Pulson Missile;
- Nuclear Bomb;
- Fusion Bomb;
- Death Spores.

Representative design vector:

- missile-heavy Cruiser/Battleship;
- at least two Missile rack sizes so Count and Ammo are independently exercised;
- one Missile modifier available at lower refinement and one requiring higher refinement;
- a conventional bomb row;
- Torpedoes remain locked if their fields are outside the profile.

### All-Tech - `game-triangle-all-tech`

All concrete fields 0-74 are granted.

Representative coverage:

- Zeon Missile with large rack and multiple legal Missile modifiers;
- Anti-Matter and/or Proton Torpedo for base Torpedo modifier coverage;
- Plasma Torpedo for No Range Dissipation coverage;
- Neutronium Bomb;
- Death Spores and Bio Terminator as bomb-family design rows with explicit downstream bombardment support status;
- save -> revise -> build -> BattleSession handoff proving immutable Count/Ammo/modifier metadata.

## 13. Gate-2 freeze requirements

Gate 2 must explicitly freeze, without guessing:

1. supported canonical IDs for all 13 audited ordnance rows;
2. stable weapon-family/category representation;
3. persistent Missile Ammo/Rack field name/type and valid values `2/5/10/15/20`;
4. Count semantics versus Ammo semantics;
5. firing-arc support or an explicit fixed/default ordnance arc policy;
6. Missile rack Space/Cost formulas and exact modifier stacking using the original quote order;
7. class-1 miniaturization reuse;
8. modifier IDs:
   - `mirv`;
   - `eccm`;
   - `armored`;
   - `fast`;
   - `emissions_guidance`;
   - `overloaded`;
   - existing `enveloping`;
   - existing `no_range_dissipation`;
9. family-mask AND per-weapon-mask legality;
10. stable validation/rejection reasons;
11. immutable ShipDesignSpec/built-Ship/BattleSession handoff fields;
12. explicit support status for conventional bombs and biological bombardment rows;
13. reference-lab expected vectors.

Gate 2 must not freeze live projectile structures into Core design snapshots.

## 14. Gate-1 checklist result

- [x] Inventory normalized missile/torpedo/bomb applications and missing data.
- [x] Re-audit technology, Space, Cost, damage, ammo/count and miniaturization.
- [x] Re-audit legal ordnance modifiers and compatibility restrictions.
- [x] Re-audit current ShipWeaponMount/Core shape for ordnance representation.
- [x] Separate design-time metadata from later tactical runtime state.
- [x] Define representative 17.0 labs: missile, torpedo and bomb-family vectors.

## 15. Gate result

**Slice 17.3 Gate 1 is complete.**

No production code was changed in Gate 1.

The repository now has enough direct original evidence to begin Gate 2 contract freeze without inventing:

- ordnance identities;
- technology ownership;
- damage/base Space/base PP;
- Missile rack/ammo semantics;
- miniaturization class;
- modifier bits/refinement/design deltas;
- Core snapshot gaps;
- Tactical/17.7 boundaries;
- reference test vectors.

Gate 2 is next and remains **not started** until explicitly opened/continued.
