# Slice 17.3 Gate 2 - Ordnance Contract Freeze

Date: 2026-09-23

Status: **active; contract freeze in progress / not complete**

Parent: `docs/slices/PLANNED_17_3_MISSILES_TORPEDOES_BOMBS_ORDNANCE.md`

Gate-1 authority: `docs/research/SLICE_17_3_GATE1_ORDNANCE_AUDIT_2026-09-23.md`

## 1. Purpose

Freeze one server-authoritative design contract for the 13 audited Missile, Torpedo and Bomb/Biological-bomb rows before Gate 3 changes production code.

This contract owns design identity, legality, Count/Ammo representation, Space/PP quotes, immutable construction snapshots and metadata handoff only. Planned Slice 17.7 still owns live projectile, interception, ECCM/evasion, Point Defense, impact/explosion, Torpedo readiness and bombardment resolution.

## 2. Frozen supported matrix

Gate 2 accepts exactly the 13 Gate-1 rows:

### Missiles

- `nuclear_missile`
- `merculite_missile`
- `pulson_missile`
- `zeon_missile`

### Torpedoes

- `anti_matter_torpedoes`
- `proton_torpedoes`
- `plasma_torpedoes`

### Bomb / biological-bomb family

- `nuclear_bomb`
- `fusion_bomb`
- `anti_matter_bomb`
- `neutronium_bomb`
- `death_spores`
- `bio_terminator`

Technology IDs, fields/RP, base Space, base PP, damage and record masks are frozen by the Gate-1 matrix and must not be duplicated as a second hand-maintained source.

## 3. Weapon-family contract

Canonical design categories:

- `beam` - existing Slice 17.2 behavior;
- `missile`;
- `torpedo`;
- `bomb`.

Death Spores and Bio Terminator use design category `bomb` plus explicit biological/bombardment metadata. They are not silently treated as ordinary Tactical ship-to-ship projectiles.

Gate 3 may normalize category metadata in a generic weapon catalog, but one canonical server rules source must own it.

## 4. Slot / Count / mount contract

The accepted eight-row weapon contract remains:

- slots `0..7`;
- slots unique;
- persisted rows sorted ascending by Slot;
- Count must be positive;
- one row is one independently configured weapon group;
- identical WeaponIDs in separate slots remain separate rows.

For all Slice-17.3 ordnance rows:

- `MountType` is canonically `standard`;
- Heavy and Point Defense mount types are Beam mount concepts and are rejected for Missile/Torpedo/Bomb rows;
- ordnance-specific behavior is expressed by Ammo/Rack and ModifierIDs, not by inventing packed WeaponIDs or overloading MountType.

## 5. Firing Arc contract - frozen

Direct original evidence closes the Gate-1 question:

- `Add_Missile_Weapon_To_Design_` writes original design-row arc byte `0x10`;
- `Add_Bomb_Weapon_To_Design_` also writes `0x10`;
- `Weapon_Arc_String_` maps `0x10` directly to the string `360`.

Therefore Slice 17.3 freezes all supported ordnance rows to **360-degree firing arc**.

HMI consequence:

- no Firing Arc selector is introduced by 17.3;
- no browser-owned arc value exists;
- Gate 3 must either persist a canonical fixed-360 metadata value or derive it deterministically from the ordnance contract, but there is no user-editable alternative in 17.3.

The persist-vs-derived representation is still being finalized before Gate 2 close; the gameplay value itself is frozen.

## 6. Ammo / rack contract

Original Ship Design rows keep Count and Ammo as distinct bytes. MOOX must therefore add a distinct authoritative ordnance Ammo/Rack concept rather than overloading Count.

### 6.1 Missiles - frozen selectable rack sizes

Legal Missile Ammo/Rack values are exactly:

- 2
- 5
- 10
- 15
- 20

The original rack value used by Space/PP is:

| Ammo/Rack | Rack base value |
| ---: | ---: |
| 2 | 10 |
| 5 | 20 |
| 10 | 30 |
| 15 | 35 |
| 20 | 40 |

Missile Ammo/Rack is design state and must survive save -> revision -> construction -> built Ship -> BattleSession metadata unchanged.

### 6.2 Torpedo / Bomb fixed values - pending representation close

Gate-1 evidence shows:

- Torpedo weapon rows carry original auxiliary/family value 2 and their runtime model is not a selectable Missile rack;
- `Add_Bomb_Weapon_To_Design_` writes fixed design Ammo byte 10;
- Bomb rows carry no normal weapon modifiers.

Before Gate 2 closes, the contract must make one explicit representation choice:

1. persist fixed Torpedo/Bomb Ammo values in the generic Ammo field, or
2. persist Ammo only for selectable Missile racks and derive fixed Torpedo/Bomb values from immutable weapon metadata.

No Gate-3 implementation starts until this representation choice is frozen.

## 7. Modifier identity and canonical order - frozen

Existing canonical IDs retain their Slice-17.2 meaning.

New ordnance IDs:

- `mirv` -> original bit 9 / `0x0200`;
- `eccm` -> bit 10 / `0x0400`;
- `armored` -> bit 11 / `0x0800`;
- `fast` -> bit 12 / `0x1000`;
- `emissions_guidance` -> bit 13 / `0x2000`;
- `overloaded` -> bit 14 / `0x4000`.

Canonical ModifierIDs ordering follows original bit order. The existing Slice-17.2 IDs therefore remain before these new IDs where their original bits are lower.

No duplicate ModifierID is legal.

## 8. Modifier legality - frozen

Effective legality is:

`family_mask AND weapon_record_mask`

Family masks:

- Missile: `0x3E00`;
- Torpedo: `0x6520`;
- Bomb: no normal weapon modifiers.

Effective sets:

- all four Missiles: MIRV, ECCM, Armored, Fast, Emissions Guidance;
- Anti-Matter Torpedo: Enveloping, ECCM, Overloaded;
- Proton Torpedo: Enveloping, ECCM, Overloaded;
- Plasma Torpedo: No Range Dissipation, Enveloping, ECCM, Overloaded;
- Bomb/Bio rows: none.

This family-mask intersection is authoritative even when a raw weapon record contains an otherwise invalid bit.

## 9. Modifier refinement and design deltas - frozen

| Modifier | Required mini level | Space delta | PP delta |
| --- | ---: | ---: | ---: |
| MIRV | 2 | +100% | +100% |
| ECCM | 1 | +25% | +25% |
| Armored | 1 | +25% | +25% |
| Fast | 1 | +25% | +25% |
| Emissions Guidance | 0 | +300% | +300% |
| Overloaded | 1 | +50% | +50% |
| existing Enveloping | 2 | +100% | +100% |
| existing No Range Dissipation | 1 | +25% | +25% |

Design legality checks required-mini-level independently of mask legality.

## 10. Space / PP quote order - frozen

All 13 ordnance Technology IDs use original Space miniaturization class 1.

For one physical row, let:

- `Bspace` = original base Space;
- `Bcost` = original base PP;
- `Rack` = Missile rack base value above, otherwise 0;
- `Dspecial` = sum of selected legal modifier percentage deltas;
- `M` = current miniaturization level;
- `N` = Count.

The original `Weapon_Space_` / `Weapon_Cost_` path freezes this order:

1. **Base/rack stage**
   - Missile: `base = B + Rack`;
   - Torpedo/Bomb: `base = B`.
2. **Special-modifier stage**
   - modifier deltas are additive, not compounded;
   - `after_special = floor(base * (100 + Dspecial) / 100)`.
3. **Regular/arc stage**
   - Slice-17.3 ordnance has no Beam Heavy/PD mount factor;
   - original Beam arc bonuses are skipped for non-Beam weapon types;
   - fixed 360 ordnance arc therefore adds no separate design penalty.
4. **Miniaturization stage**
   - reuse the already accepted class-1 Space and universal PP factors from 17.2.
5. **Count stage**
   - row totals multiply the frozen per-unit result by `N`.

Frozen miniaturization factors:

Space class 1:

- M0 100%
- M1 80%
- M2 70%
- M3 60%
- M4 50%
- M5+ 40%

PP:

- M0 100%
- M1 75%
- M2 55%
- M3 40%
- M4 30%
- M5+ 25%

Gate 3 must implement one server-owned quote function used by catalog preview, save validation and construction snapshots. No browser-local parallel formula is authoritative.

## 11. Stable rejection-reason set - initial freeze

Reuse existing generic reasons where applicable:

- `weapon_unknown`
- `weapon_technology_required`
- `weapon_slot_out_of_range`
- `weapon_slot_duplicate`
- `weapon_count_invalid`
- `weapon_mount_not_allowed`
- `weapon_modifier_unknown`
- `weapon_modifier_not_allowed`
- `weapon_modifier_refinement_required`
- `weapon_modifier_duplicate`
- existing global `space_exceeded`

Ordnance-specific reasons to freeze before Gate 2 close:

- `weapon_ammo_required`
- `weapon_ammo_invalid`
- `weapon_ammo_not_allowed`

Translated UI text is downstream from these machine-readable reasons.

## 12. Immutable snapshot / handoff contract - active freeze

Conceptual target:

```text
ShipWeaponMount {
    Slot
    WeaponID
    Count
    MountType
    ModifierIDs
    Ammo
}
```

Frozen semantics already:

- Ammo is not Count;
- Missile selectable Ammo/Rack is immutable saved design state;
- ordnance MountType is Standard;
- ModifierIDs are canonical sorted/unique;
- constructed/built Ship snapshots do not re-read later research or later design rows;
- BattleSession receives the same frozen ordnance configuration metadata;
- live remaining ammo, projectile positions, target locks, interception results and readiness counters are not design snapshot fields.

Before Gate 2 closes, fixed Torpedo/Bomb Ammo representation and fixed-360 persist-vs-derived representation must be resolved.

## 13. Tactical support boundary - frozen

Slice 17.3 may hand complete immutable metadata to BattleSession, but new ordnance rows remain explicit non-simulated metadata until their consumer exists.

Planned Slice 17.7 owns:

- projectile entities/flight;
- Missile/Torpedo speed;
- target tracking/retargeting;
- hit/evasion/ECCM;
- Point Defense/interception;
- Missile hit points;
- MIRV;
- Emissions Guidance runtime effect;
- Torpedo cooldown/readiness;
- Plasma range dissipation;
- impact/explosion/damage;
- bomb/bombardment runtime resolution.

Gate 3 must never fake these behaviors just to make a design appear Tactical-ready.

## 14. Reference contract - frozen

Use the existing Slice-17.0 labs:

### Baseline

`game-triangle-2pc`

- Nuclear Missile;
- Nuclear Bomb;
- later ordnance locked.

### Mid-Tech

`game-triangle-mid-tech`

- Nuclear/Merculite/Pulson Missile;
- Nuclear/Fusion Bomb;
- Death Spores;
- Torpedoes locked by current profile.

### All-Tech

`game-triangle-all-tech`

- all 13 rows;
- large Missile rack + multiple modifiers;
- Torpedo modifier coverage;
- Plasma No Range Dissipation;
- Neutronium Bomb;
- Death Spores + Bio Terminator;
- save -> revise -> build -> BattleSession immutable metadata checks.

## 15. Remaining work before Gate 2 close

- [x] Supported 13-row matrix frozen.
- [x] Slot/Count and standard-only ordnance mount semantics frozen.
- [x] Firing Arc gameplay policy frozen to fixed 360 degrees.
- [x] Missile selectable Ammo/Rack values frozen.
- [x] Modifier IDs, order, legality, refinement and deltas frozen.
- [x] Space/PP quote order and miniaturization reuse frozen.
- [x] Slice-17.7 Tactical boundary frozen.
- [x] Reference-lab vectors frozen.
- [ ] Freeze fixed Torpedo/Bomb Ammo representation in snapshots/API.
- [ ] Freeze fixed-360 persist-vs-derived representation.
- [ ] Finalize machine-readable ordnance rejection reasons.
- [ ] Finalize exact ShipWeaponMount / TacticalWeaponSpec field shapes.

Gate 2 remains **active** until the final four representation/API items are closed.
