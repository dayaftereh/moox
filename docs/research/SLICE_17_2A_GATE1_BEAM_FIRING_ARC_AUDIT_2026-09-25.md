# Slice 17.2A Gate 1 - Beam Weapon Firing Arc Audit

Date: 2026-09-25

Status: **Gate 1 complete; Gate 2 contract freeze ready / not started**

Parent plan: `docs/slices/PLANNED_17_2A_BEAM_WEAPON_FIRING_ARCS.md`

Open marker: `docs/slices/_OPEN_BEAM_WEAPON_FIRING_ARCS_2026-09-25.md`

## 1. Scope and result

Slice 17.2A is a bounded follow-up to accepted Slice 17.2. Gate 1 re-audits only the missing original Beam firing-arc design dimension.

Gate 1 proves:
- the five original Beam firing-arc identities;
- the Human Ship Designer default;
- original arc labels and geometry masks;
- arc legality/compatibility;
- exact arc surcharge placement relative to mount, weapon modifiers and miniaturization;
- the persistent original design-row arc field;
- current MOOX representation gaps;
- deterministic Baseline/Mid/All-Tech reference vectors.

No production Beam-arc support is implemented by Gate 1.

Complete Tactical direction/target-geometry enforcement remains owned by Slice 17.6.

## 2. Original source identity

Direct evidence was re-read from the installed MOO2 1.31 sources used by the earlier Slice-17 weapon audits:

- `C:\ASH\Installation\mastori2\Orion2.exe`
  - SHA-256: `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`
- `C:\ASH\Installation\mastori2\KENTEXT.LBX`
  - SHA-256: `ba5952ad3bcb2484d5dff376eba564f7d8c2f71f26168922710f7892c1d87914`
- `C:\ASH\Installation\mastori2\HESTRNGS.LBX`
  - SHA-256: `a3193f56d12ae512ab8a78cc50aba9df44d6d614a4cc23b14ec7b9f1b09cbf30`

The x86 audit used the existing external temporary disassembly module under
`C:\ASH\Temp\moox-s173-g1-audit`.
It is outside the MOOX repository and did not alter `go.mod` or `go.sum`.

## 3. Original executable anchors

Watcom symbols resolve the relevant Object-1 offsets:

| Symbol | Object-1 offset | Gate-1 use |
| --- | ---: | --- |
| `Get_Weapon_Arc_` | `0x22AD1` | combat geometry -> arc-mask mapping |
| `Add_Beam_Weapon_To_Design_` | `0x5147F` | original 8-byte design-row write |
| `Set_New_Weapon_Modifications_` | `0x57F5C` | Human Designer edit-state initialization |
| `Update_Weapon_Replacement_` | `0x58E78` | installs Human Designer weapon selection |
| `Weapon_Replacement_Help_` | `0x5953B` | Human Designer arc choices/default |
| `Bonus_For_Weapon_Arcs_` | `0x5A680` | original arc surcharge |
| `Build_Initial_Design_Template_` | `0x5A7C8` | persisted-row -> editable template projection |
| `Modify_Weapon_By_Arcs_` | `0x5D0DE` | directional-bit helper |
| `Weapon_Cost_` | `0x5EC74` | authoritative original PP path |
| `One_Weapon_Space_` | `0x5EDC6` | one-weapon Space helper |
| `Weapon_Space_` | `0x5EE8E` | authoritative original Space path |
| `Weapon_Type_` | `0x5EFEB` | reads original record type byte |
| `Weapon_Arc_String_` | `0x5EFF8` | original player-facing short labels |
| `Design_Weapon_Firing_Arc_` | `0x67DD6` | persistent design arc accessor |

## 4. Canonical original Beam arc matrix

`Weapon_Arc_String_` maps five single-bit values.

`KENTEXT.LBX` group-6 records resolve the English short labels:
- message 3 -> `F`;
- message 4 -> `Fx`;
- message 5 -> `B`;
- message 6 -> `Bx`.

`Weapon_Arc_String_` maps the bits in this order:

| Original bit | Short label | Gate-2 canonical meaning | Geometry from `Get_Weapon_Arc_` | Arc surcharge |
| ---: | --- | --- | --- | ---: |
| `0x01` | `F` | Forward | <=60 or >=300 degrees; 120-degree forward sector | 0% |
| `0x02` | `Fx` | Forward Extended | <=120 or >=240 degrees; 240-degree forward sector | +25% |
| `0x04` | `Bx` | Back Extended | 60..300 degrees; 240-degree rear sector | +25% |
| `0x08` | `B` | Back / Rear | 120..240 degrees; 120-degree rear sector | 0% |
| `0x10` | `360` | 360 degrees | unrestricted design arc | +50% |

`HESTRNGS.LBX` independently contains the expanded Ship Designer wording
`Forward Extended`, `Back Extended`, `Rear` and `360`.

The English short-label pairing `F/Fx` and `B/Bx`, together with the proven geometry intervals, supports the canonical Gate-2 identity names above.

### 4.1 These are five choices, not arbitrary bitmask combinations

The Human Designer iterates exactly five arc choice buttons.
When a Beam arc button is selected it stores:

`arc = 1 << buttonIndex`

Therefore the legal design values are the five single-bit values
`{0x01,0x02,0x04,0x08,0x10}`.

`Get_Weapon_Arc_` returns a runtime geometry mask whose bits can overlap at a target angle.
That runtime mask must not be confused with the persisted single-choice design arc.

## 5. Original Human Designer default

The normal Human Designer path was traced independently from the AI/auto-design path.

Important correction from the audit:
`Determine_Beam_Weapon_Modifications_` is used by auto-design logic and includes random choices.
It is not valid evidence for the Human Designer default.

The Human Designer path is:

1. `Set_New_Weapon_Modifications_`
   - initializes global current arc `0x23976` to zero for a new empty slot;
   - when editing an installed row, copies the existing saved arc into `0x23976`.

2. `Weapon_Replacement_Help_`
   - calls `Weapon_Type_(selectedWeapon)`;
   - stores that result in `cx`;
   - `Weapon_Type_` directly reads the original weapon-record type byte at record offset `+0x0C`;
   - if type == 0, it writes `0x23976 = 0x01`;
   - otherwise it writes `0x23976 = 0x10`.

3. `Update_Weapon_Replacement_`
   - copies `byte[0x23976]` into the editable template firing-arc field.

Original type 0 is Beam.

Therefore the normal original Human Designer default for a newly added Beam weapon is:

**Forward / F / `0x01`.**

Non-Beam weapon families use the separate 360-degree behavior already handled by closed Slice 17.3.

## 6. Persistence proof

The original weapon design row remains the previously proven 8-byte structure.

For Beam add:
- row `+0..+1`: weapon identity word;
- row `+2`: Count;
- row `+4`: Firing Arc byte;
- row `+5..+6`: specials/modifier word;
- row `+7`: Ammo, written as zero for Beam.

`Add_Beam_Weapon_To_Design_` writes the selected arc byte to row `+4`.

`Design_Weapon_Firing_Arc_` later reads that same persistent field from the player design structure.

Therefore Beam firing arc is immutable design metadata and is not browser-only or Tactical-only state.

## 7. Arc legality and compatibility

The Human Designer replacement-field path first separates weapon families:

- `Weapon_Needs_Racks_` routes Missile rack choices separately;
- the arc-choice branch requires `Weapon_Type_ == 0`, i.e. Beam.

For a Beam, the arc-choice path:
- loops exactly five times;
- creates all five arc controls;
- marks every one of the five controls active;
- contains no Beam-ID-specific condition;
- contains no Heavy/Point-Defense/Standard condition.

The only Beam mount restrictions remain the already accepted Slice-17.2 mount legality.

Gate-1 conclusion:

**Every supported Beam row may use all five firing arcs whenever the selected Beam mount itself is legal. There is no additional arc x mount or arc x Beam-ID compatibility mask.**

## 8. Original Space/PP surcharge arithmetic

`Bonus_For_Weapon_Arcs_` proves:
- `0x02` or `0x04` -> +25%;
- `0x10` -> +50%;
- otherwise -> 0%.

`Weapon_Cost_` and `Weapon_Space_` independently inline the same arc logic in the authoritative quote path.

The important ordering is:

1. original base weapon value;
2. existing Slice-17.2 mount factor;
3. a second additive configuration factor containing:
   - 100%;
   - arc surcharge;
   - selected weapon-modifier delta;
4. existing miniaturization path.

Thus the core configuration stage is:

`configured = floor(after_mount * (100 + arc_bonus + weapon_modifier_bonus) / 100)`

The arc surcharge is **additive with the weapon-modifier percentage**, not a later multiplicative factor.

Examples:
- no modifier + 360 -> 150% of mounted base before miniaturization;
- +100% selected modifiers + 360 -> 250% of mounted base before miniaturization.

### 8.1 Preserve accepted Slice-17.2 count semantics

This bounded follow-up does not reopen the accepted Slice-17.2 unit/count contract.

Gate 2 must integrate the arc surcharge into the existing authoritative MOOX Beam quote while preserving the accepted Count contract unless a separately scoped parity correction is explicitly opened.

Gate-1 reference vectors below therefore use `Count=1` to isolate the firing-arc contract from unrelated grouped-row rounding.

## 9. Current MOOX gap audit

### 9.1 Core

Already present from Slice 17.3:
- `type ShipWeaponFiringArcID string`;
- `ShipWeaponMount.FiringArcID`;
- canonical `ShipWeaponFiringArc360 = "360"`.

Missing for 17.2A:
- canonical Beam arc IDs for Forward, Forward Extended, Back Extended and Back/Rear;
- Beam canonicalization/default rules.

### 9.2 Authoritative Beam quote

Current `quoteBeamWeapon` explicitly rejects any non-empty `FiringArcID` with
`weapon_firing_arc_not_allowed` and the message that Beam arcs are deferred to Slice 17.2A.

Missing:
- Beam arc legality;
- default Forward canonicalization;
- +25/+50 arc surcharge in the server quote;
- stable arc validation reasons.

### 9.3 Designer catalog/API

The current catalog exposes `FixedFiringArcID` only for fixed-360 ordnance.

Missing:
- Beam arc choices;
- Beam default arc;
- labels/order;
- server-owned preview metadata for selectable Beam arcs.

### 9.4 Web client

Current TypeScript narrows `ShipWeaponFiringArcID` to only `'360'`.

Current Weapon Selector:
- can persist an already supplied fixed arc;
- shows only the read-only fixed-360 label for ordnance;
- has no selectable Beam arc control.

Gate 3 must add a compact selector driven only by server-owned catalog/preview values.

### 9.5 BattleSession / Tactical boundary

Already present:
- `TacticalWeaponSpec.FiringArcID`;
- `TacticalWeaponView.FiringArcID`;
- Core -> Tactical metadata copy.

Therefore 17.2A does not need a second Tactical arc field.

Slice 17.6 remains responsible for:
- ship facing;
- target-relative geometry;
- whether a chosen arc can bear on a target;
- Tactical action gating/rendering based on the persisted arc.

17.2A only guarantees lossless authoritative design metadata.

## 10. Reference-profile evidence

Fresh `-reference-games` snapshots were queried through the canonical server path.

### 10.1 Available Beam catalog by profile

Baseline `game-triangle-2pc`:
- Laser Cannon only.

Mid-Tech `game-triangle-mid-tech`:
- Laser Cannon;
- Mass Driver;
- Fusion Beam;
- Ion Pulse Cannon;
- Graviton Beam;
- Neutron Blaster.

All-Tech `game-triangle-all-tech`:
- all 13 accepted Slice-17.2 Beam rows.

Live catalog checks also confirmed:
- Baseline Laser: mini level 0, Standard 10 Space / 5 PP;
- Mid Neutron Blaster: mini level 1, Standard 8 Space / 6 PP;
- All-Tech Phasor: mini level 4, Standard 5 Space / 3 PP before optional mount/modifier configuration.

## 11. Gate-2 reference vectors

These are Gate-1 expected vectors to freeze formally in Gate 2.
They describe the original arc arithmetic applied to the already accepted Slice-17.2 quote pipeline.

### Vector A - Baseline default and 360 surcharge

Profile: Baseline
Weapon: Laser Cannon
Count: 1
Mount: Standard
Modifiers: none

Default:
- arc: Forward / `0x01`;
- Space: 10;
- PP: 5.

360 variant:
- arc: `0x10`;
- pre-mini factor: 150%;
- expected Space: 15;
- expected PP: 7.

This is the minimum reference proving the original +50% 360 surcharge.

### Vector B - Mid-Tech miniaturization

Profile: Mid-Tech
Weapon: Neutron Blaster
Count: 1
Mount: Standard
Modifiers: none
Mini level: 1

Current accepted no-arc quote:
- Space: 8;
- PP: 6.

360 variant:
- base 10/8;
- configured at +50% -> 15/12 before miniaturization;
- M1 Space 80% -> 12;
- M1 Cost 75% -> 9.

Expected:
- Space: 12;
- PP: 9.

### Vector C - All-Tech mount + modifiers + arc

Profile: All-Tech
Weapon: Phasor
Count: 1
Mount: Point Defense
Modifiers:
- Shield Piercing;
- Autofire.
Mini level: 4

Accepted 17.2 arithmetic:
- base 10/10;
- Point Defense -> 5/5 mounted base;
- Shield Piercing + Autofire -> +100% weapon-modifier delta.

Without arc:
- configured before mini -> 10/10;
- expected accepted quote -> 5 Space / 3 PP.

With 360:
- second factor becomes `100 + 100 + 50 = 250%`;
- configured before mini -> floor(5 * 250 / 100) = 12;
- M4 Space 50% -> 6;
- M4 Cost 30% -> 3.

Expected:
- Space: 6;
- PP: 3.

This vector proves that 360 is additive with selected weapon-modifier percentage after the mount factor and before miniaturization.

## 12. Gate-1 checklist

- [x] Enumerate every original Beam firing-arc bit/value and player-facing label.
- [x] Prove the original default Beam arc used by normal add/design flows.
- [x] Re-audit `Weapon_Arc_String_`, `Bonus_For_Weapon_Arcs_`, `Get_Weapon_Arc_` and Beam add/edit paths.
- [x] Freeze which arcs are legal for which Beam rows/mounts, if compatibility restrictions exist.
- [x] Freeze exact Space/PP bonus arithmetic and ordering relative to mount/modifier/miniaturization.
- [x] Audit the current MOOX Core/Ruleset/API/Weapon Selector/BattleSession gaps.
- [x] Define Baseline/Mid/All-Tech reference vectors including at least one 360-degree Beam row.

Gate 1 is complete.

## 13. Gate-2 inputs

Gate 2 must freeze, before any implementation:

1. canonical MOOX IDs for the five Beam arcs;
2. canonical order and DE/EN labels;
3. default Beam arc = Forward;
4. exact +0/+25/+25/+0/+50 surcharge table;
5. all-five-arcs-for-all-legal-Beams compatibility rule;
6. stable invalid/unknown arc rejection reasons;
7. exact integration into the accepted 17.2 mount/modifier/miniaturization quote;
8. persistent `ShipWeaponMount.FiringArcID` and immutable built-Ship/BattleSession handoff;
9. server catalog/preview fields for selectable arcs;
10. explicit Slice-17.6 Tactical enforcement deferral;
11. Baseline/Mid/All-Tech expected vectors and browser coverage.

No Gate-3 implementation is authorized by this Gate-1 result.
