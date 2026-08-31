# Military Ship core and design baseline - MOO2 1.31 Gate 1

Date: 2026-08-31

Status: **closed; Gates 1-4 complete**

Recovery marker: removed at Gate 4 closure

Planned slice: `docs/slices/PLANNED_03_MILITARY_SHIP_CORE_DESIGN_BASELINE.md`

## Purpose

Resolve the smallest directly evidenced Master of Orion II 1.31 military Ship/Design model needed for MOOX to construct a concrete combat ship without prematurely implementing generic Fleet movement or tactical combat.

Original copyrighted executable/data are private reference evidence only and are not distributable MOOX content.

## Reference identity and method

Private reference executable:

- `C:\ASH\Temp\mastori2\Orion2.exe`
- SHA-256 `7AE2AC2E5904CA330009AF2827279D889906B0B9B7A8854C38EB707A56E955B5`
- DOS/4GW-bound inner LE image
- LE header file offset `0x292E4`
- object 1 relocation base `0x10000`, virtual size `0x160695`
- object 2 relocation base `0x178000`, virtual size `0x5DCD0`; only its physically mapped leading pages were read because its virtual tail is zero-fill/BSS

Gate 1 used the repository `internal/moo2exe` reader plus local `objdump` against temporary private extracts. All temporary helpers/extracts are removed before Gate 1 closes.

Evidence labels:

- **direct original code/data** - observed directly in the 1.31 executable/data tables;
- **original-derived** - semantic interpretation supported by multiple direct paths;
- **MOOX proposal/divergence** - architecture or intentionally narrow behavior proposed for the headless runtime, not a claim about original storage.

## Existing MOOX baseline

Before this slice:

- Core `StateSchemaVersion` is **17**;
- `data/rulesets/moo2-1.31/ship_hulls.json` normalizes all six player hull identities plus strategic/tactical artwork, but not authoritative runtime space/cost;
- `StrategicFleet` has owner, combat/civilian role, location and fixed Colony/Outpost special-ship transit state;
- ordinary combat Fleets are still abstract strategic blockade-presence markers and have no concrete Ship composition;
- there is no persisted `Ship`, `ShipDesign` or military-design Construction project;
- `internal/battle` is lifecycle/session infrastructure only, not tactical combat rules.

## Relevant original symbols

Object-1 symbol offsets below use executable object-1 offsets; add `0x10000` for the disassembly VA.

| Symbol | object-1 offset | VA |
| --- | ---: | ---: |
| `Init_Ship_Designs_` | `0x44FBF` | `0x54FBF` |
| `Best_Warp_Drive_` | `0x4679E` | `0x5679E` |
| `Best_Ship_Shield_` | `0x4679E` | `0x5679E` |
| `Best_Computer_` | `0x4680D` | `0x5680D` |
| `Best_Armor_` | `0x4685F` | `0x5685F` |
| `Best_Ship_Fuel_` | `0x4709F` | `0x5709F` |
| `Build_Initial_Design_Template_` | `0x5A7C8` | `0x6A7C8` |
| `Clear_Design_` | `0x5B247` | `0x6B247` |
| `Clear_Design_For_Colony_Builds_` | `0x5B342` | `0x6B342` |
| `Cost_Of_Design_` | `0x5B577` | `0x6B577` |
| `Design_Cost_For_Player_` | `0x5B99D` | `0x6B99D` |
| `Design_Template_Costs_` | `0x5C087` | `0x6C087` |
| `Design_Template_Space_Requirements_` | `0x5C287` | `0x6C287` |
| `Total_Design_Cost_` | `0x5E777` | `0x6E777` |
| `Total_Design_Space_` | `0x5E81F` | `0x6E81F` |
| `Update_Player_Design_` | `0x5EA48` | `0x6EA48` |
| `Auto_Design_Ship_` | `0x516A5` | `0x616A5` |
| `Ship_Size_Points_At_Star_` | `0x13B28` | `0x23B28` |
| `Colbldg_Create_Ship_` | `0x9F76F` | `0xAF76F` |
| `Create_Ship_` | `0x9F7B4` | `0xAF7B4` |
| `Delete_Ship_` | `0x9F9EC` | `0xAF9EC` |
| `Colony_Can_Build_Large_Ships_` | `0xD5275` | `0xE5275` |
| `Cost_Reduction_For_Govt_Type_` | `0x5E1A0` | `0x6E1A0` |

## Finding 1 - all six player hull base costs and spaces are directly table-backed

Evidence level: **direct original code/data**.

`Total_Design_Cost_` reads the current hull-size index from the working design, indexes a `0x24`-byte hull record and reads the base hull cost from:

```text
object2 + 0x801E + size_index * 0x24
```

`Total_Design_Space_` reads hull capacity from:

```text
object2 + 0x8020 + size_index * 0x24
```

The six direct records are:

| size index | Hull | Base cost PP | Base space |
| ---: | --- | ---: | ---: |
| 0 | Frigate | 20 | 25 |
| 1 | Destroyer | 70 | 60 |
| 2 | Cruiser | 250 | 120 |
| 3 | Battleship | 600 | 250 |
| 4 | Titan | 1500 | 500 |
| 5 | Doom Star | 4000 | 1200 |

These are **hull-only** values. A military Ship's Construction cost is not merely the hull base cost.

`Total_Design_Cost_` adds the installed shield/drive/armor/computer costs plus weapon and special-system costs, then calls `Cost_Reduction_For_Govt_Type_`. Therefore the same established ship-government cost reducer used for Colony/Outpost ships is part of normal military design cost as well.

For the currently modeled MOOX government surface, the existing Feudal rule remains `ceil(2 * design_cost / 3)`. Evolved Confederation behavior remains outside this slice.

## Finding 2 - the original has exactly six current military design slots

Evidence level: **direct original code**.

`Init_Ship_Designs_` addresses the player design array at:

```text
player + 0x326
```

Each current design record is exactly:

```text
0x63 bytes
```

The initialization loops exactly six slots. In normal combat mode it invokes `Auto_Design_Ship_` for six initial designs; in strategic-combat mode it similarly creates six strategic designs.

Thus the directly supported current-design surface is:

```text
6 mutable current design slots per Empire
```

The original slot storage has no stable GUID-like design identity; slot position is the primary current-design identity.

## Finding 3 - a normal military design stores the core hull/equipment identity directly

Evidence level: **direct original code**.

`Auto_Design_Ship_` writes these normal design fields:

```text
design + 0x10 = hull size index
design + 0x11 = 0                    // normal military design
design + 0x12 = best owned shield
design + 0x13 = best owned warp drive
design + 0x14 = installed/derived FTL speed
design + 0x15 = best owned computer
design + 0x16 = best owned armor
```

The design also contains weapon/special arrays and a picture ID; the existing normalized hull artwork already records the original strategic picture families. Doom Star uses the established special strategic picture 43.

`Build_Initial_Design_Template_` independently confirms the same equipment families and additionally selects the best available Fuel Cell for the working design.

Direct table/technology links observed for the ordinary player path include:

### Drives

| component index | Technology |
| ---: | --- |
| 1 | 120 `nuclear_drive` |
| 2 | 72 `fusion_drive` |
| 3 | 96 `ion_drive` |
| 4 | 11 `anti_matter_drive` |
| 5 | 88 `hyper_drive` |
| 6 | 95 `interphased_drive` |

### Computers

| component index | Technology |
| ---: | --- |
| 1 | 58 `electronic_computer` |
| 2 | 122 `optronic_computer` |
| 3 | 143 `positronic_computer` |
| 4 | 44 `cybertronic_computer` |
| 5 | 110 `moleculartronic_computer` |

### Shields

| component index | Technology |
| ---: | --- |
| 1 | 33 `class_i_shield` |
| 2 | 34 `class_iii_shield` |
| 3 | 35 `class_v_shield` |
| 4 | 36 `class_vii_shield` |
| 5 | 37 `class_x_shield` |

### Armor

Armor component index 1 is the original baseline Titanium level. Higher Best-Armor candidates directly map to:

| component index | Technology |
| ---: | --- |
| 2 | 191 `tritanium_armor` |
| 3 | 203 `zortrium_armor` |
| 4 | 117 `neutronium_armor` |
| 5 | 2 `adamantium_armor` |
| 6 | 201 `xentronium_armor` |

Normalized Technology 187 is `titanium_armor`; the Best-Armor loop starts above the baseline component and uses Titanium as its starting armor value.

### Fuel Cells

| component index | Technology | Original range value |
| ---: | --- | ---: |
| 1 | 167 `standard_fuel_cells` | 4 |
| 2 | 51 `deuterium_fuel_cells` | 6 |
| 3 | 98 `iridium_fuel_cells` | 9 |
| 4 | 194 `urridium_fuel_cells` | 12 |
| 5 | 184 `thorium_fuel_cells` | 255 |

This agrees with the Fuel-range evidence already normalized for Colony/Outpost Ship movement.

## Finding 4 - design cost is a complete installed-design cost

Evidence level: **direct original code**.

`Design_Cost_For_Player_` builds the working template and returns its calculated cost. `Total_Design_Cost_` starts with hull cost, adds mandatory equipment and all installed weapon/special costs, then applies government reduction.

`Design_Template_Costs_` directly uses per-component/per-hull data for shields, drives and computers and an armor multiplier against hull base cost. Therefore Gate 3 cannot correctly construct military Ships by charging only the hull table value.

The minimum implementation needs an original-derived normalized subset for the mandatory equipment families used by the first supported design. Full weapon/special cost normalization is not required if the first supported fixture carries no weapons/specials.

## Finding 5 - an unarmed/cleared current design is a viable narrow baseline

Evidence level: **direct original code, original-derived legality conclusion**.

`Clear_Design_` zeros the weapon and special-system arrays/counts while restoring/selecting best current mandatory systems such as Warp Drive, Computer, Armor and Fuel, then recalculates design data.

The Design Screen's save/commit path calls `Update_Player_Design_`, which copies the working design into the selected `0x63` current slot. No weapon-count presence guard was observed in that direct save path before the update.

Therefore the directly supported narrow conclusion is:

> the original can represent and commit a cleared military design whose weapon/special arrays are empty while the baseline strategic equipment remains installed.

This provides a useful first MOOX fixture without pretending tactical weapons are implemented.

Gate 1 does **not** claim that the original's normal new-game Auto-Designs are unarmed; `Auto_Design_Ship_` goes on to choose weapons/specials. The cleared design is deliberately the smaller supported surface.

## Finding 6 - built Ships copy the current design; they do not live-reference the mutable slot

Evidence level: **direct original code**.

`Create_Ship_` allocates one of up to 500 original Ship records. The full Ship record is `0x81` bytes.

For a normal player design it computes:

```text
player + 0x326 + design_slot * 0x63
```

and copies all `0x63` design bytes directly into the first `0x63` bytes of the new `0x81` Ship record.

It then writes per-instance data after the copied design snapshot, including owner, state/status, system/map location and naming/serial state.

`Colbldg_Create_Ship_` creates the ship for the producing Colony and marks the newly built record with its post-build status.

This proves the key lifecycle invariant:

```text
current ShipDesign slot may later change
!=
already-built Ship design snapshot changes
```

A MOOX model containing only a live mutable `Ship.DesignID` reference without snapshot/revision semantics would therefore be wrong.

## Finding 7 - military completion is located at the producing Colony's System

Evidence level: **direct original code**.

`Colbldg_Create_Ship_` derives the producing Colony's owner and calls `Create_Ship_`. `Create_Ship_` resolves the Colony's Planet/Star relation and initializes the concrete Ship location to that Star.

For the MOOX strategic abstraction, military Construction completion should therefore create the concrete Ship at the producing Colony's `StarSystem`.

Whether MOOX groups that Ship into a newly allocated one-ship Fleet or merges it into an existing Fleet is an architectural choice; the original physical Ship record itself is independent.

## Finding 8 - current design slots are mutable, while refit is a separate operation

Evidence level: **direct original code and symbol/call separation**.

`Update_Player_Design_` writes the current working design back into the selected current slot. Separate original paths exist for Ship refit and refit cost (`Add_Refited_Ship_To_Queue_`, `Refit_Cost_`, refit UI).

Combined with the full design copy in `Create_Ship_`, the correct semantic distinction is:

```text
editing/replacing a current design slot
!=
refitting existing Ships
```

This slice does not need to implement refit.

The exact original interaction between changing a design slot and a Colony that is already constructing that slot is not closed strongly enough here to justify copying a subtle queue-side effect. For the first MOOX implementation, Gate 2 should either snapshot a design revision when queued or reject edits to a design referenced by active Construction. That rule must be labeled as a deterministic MOOX boundary rather than falsely presented as a proven original quirk.

## Finding 9 - hull-size points are size index + 1, but Command-Point accounting remains deferred

Evidence level: **direct original code for the size-point function; insufficient direct link to CP maintenance in this Gate**.

`Ship_Size_Points_At_Star_` scans concrete `0x81` Ship records at a Star, reads each Ship's copied hull-size byte at `ship+0x10`, adds one and accumulates:

```text
Frigate    -> 1
Destroyer  -> 2
Cruiser    -> 3
Battleship -> 4
Titan      -> 5
Doom Star  -> 6
```

This is useful structural evidence that hull size is sufficient to derive a canonical 1..6 weight. Gate 1 did **not** prove that every one of these exact points is the final Command-Point maintenance consumption formula; the dedicated Command Points / ship Maintenance slice remains the correct place to close capacity/overage timing and modifiers.

Therefore this slice should persist hull identity, not a mutable/persisted `CommandPointCost` field.

## Finding 10 - large-hull buildability has additional Colony infrastructure semantics

Evidence level: **direct original code**.

`Colony_Can_Build_Large_Ships_` checks Colony building state rather than merely hull table identity. This confirms that broader hull availability/buildability is not just `all six hulls are always legal`.

Gate 1 intentionally does not expand into every Star Base/Battlestation/Star Fortress/hull unlock rule. The first vertical fixture should use **Frigate**, avoiding unsupported large-hull infrastructure semantics while all six authoritative hull base values are normalized for future work.

## Current MOOX gap

Core schema 17 currently has:

```text
StrategicFleet {
  ID,
  EmpireID,
  Role,
  SpecialKind,
  AtSystemID,
  DestinationSystemID,
  RemainingTurns,
  FTLSpeed
}
```

There is no concrete military `Ship`, no current `ShipDesign`, and no Fleet composition. Ordinary combat Fleets cannot carry semantic transit state yet, which remains appropriate until Slice 04.

Current Construction has semantic project kind/id but no military-design identity/revision snapshot.

## Accepted Gate-2 contract

The recommended narrow fidelity-first contract is:

1. **Core schema 18.** Add global sorted `GameState.ShipDesigns []ShipDesign` and `GameState.Ships []Ship`; extend combat `StrategicFleet` with deterministic `ShipIDs []ID` composition.
2. **Unbounded active MOOX design catalog per Empire.** The original six `0x63` slots remain an evidence fact, not a MOOX gameplay limit. `ShipDesign` contains a stable MOOX ID, `EmpireID`, monotonic `Revision`, name, hull identity, legal picture identity and the supported installed-equipment snapshot. Validation imposes no fixed maximum design count; stable IDs plus deterministic ordering replace the original memory-oriented slot ceiling.
3. **Snapshot semantics.** A built `Ship` stores `SourceDesignID` + `SourceDesignRevision` plus its own immutable build snapshot of the design-relevant fields. Later edits to the current design slot cannot retroactively alter that Ship. This directly mirrors the original `0x63` design copy into the Ship record.
4. **Normalize hull runtime values.** Extend `ship_hulls.json` with the directly proven six base-cost/base-space values. Do not add a persisted CP field; size-index-derived CP/maintenance accounting remains Slice 05.
5. **Normalize only the mandatory component subset needed by this slice.** Record original component index/Technology linkage and per-hull cost/space data for Warp Drive, Computer, Armor, Shield and Fuel Cell families. Full beam/missile/bomb/special-system normalization remains deferred.
6. **First supported design surface = cleared/unarmed Frigate.** Add an authoritative design command that creates/replaces a selected slot with hull `frigate`, a legal Frigate picture, best currently available original mandatory equipment, zero weapon mounts and zero special systems. Other hull construction/design editing can remain explicitly unsupported until their availability rules are closed.
7. **No hidden RNG requirement for the first fixture.** Picture selection is cosmetic; require/validate an explicit legal Frigate picture or choose one deterministic canonical Frigate picture. Do not imitate original Auto-Design weapon/picture RNG while weapon design is intentionally unsupported.
8. **Authoritative design cost.** Calculate and snapshot full supported-design PP cost from hull + mandatory installed equipment using the original table/formula subset, then apply the same government ship-cost adjustment already used by special ships. Do not use hull base cost as final Construction cost.
9. **Construction project.** Add `military_ship` Construction referencing `ShipDesignID` and `ShipDesignRevision`. For the first slice, reject replacing a design revision while any Colony actively constructs that revision; this avoids inventing unresolved original queue-mutation behavior and guarantees deterministic completion.
10. **Completion creates a concrete Ship snapshot.** Allocate a `Ship`, copy the supported design snapshot, and place it at the producing Colony's System.
11. **One-ship combat Fleet on completion.** Allocate a new `StrategicFleet{Role: combat}` containing that Ship ID at the producing System. Do not implement generic combat-Fleet transit, merge or split here; Slice 04 owns those rules.
12. **Minimum Ship fields.** Persist stable Ship ID, owner, source design ID/revision, hull, picture, installed Warp Drive/FTL speed, Shield, Computer, Armor and Fuel Cell plus the supported design cost/space snapshot. Damage, crew experience, tactical position and weapon state remain deferred.
13. **Blockade integration only through existing location semantics.** The new one-ship combat Fleet is real strategic combat presence and therefore participates in the already-authoritative blockade producer when that producer next runs. No new combat/encounter resolution is added.
14. **Save/load/Observer/replay.** Schema-18 exact round-trip, clone isolation, deterministic design/Construction/Ship/Fleet IDs/events and replay equality are required.
15. **Vertical regression.** Prove: Empire -> create cleared Frigate design -> queue normal Construction -> pay exact supported design PP -> completion at source System -> concrete Ship snapshot -> one-ship combat Fleet -> later current-design replacement leaves built Ship unchanged.
16. **Explicit rejects.** Gate 3 should reject unsupported weapon/special payloads, unsupported non-Frigate design creation, generic combat-Fleet movement and refit rather than silently approximating them.

### Accepted deliberate MOOX scope choices

The following are not claims about hidden original behavior:

- allowing an arbitrary number of active MOOX designs per Empire rather than reproducing the original six-slot storage ceiling; the six-slot count is retained solely as original-evidence provenance;

- using a stable MOOX design ID + revision on top of original six slot semantics;
- starting with a cleared/unarmed Frigate instead of reproducing the original randomized weaponized Auto-Design set;
- prohibiting replacement of an actively constructed design revision until queue-mutation/refit behavior is separately modeled;
- creating a one-ship `StrategicFleet` on completion rather than trying to reproduce the original UI stack organization.

These choices preserve the direct gameplay invariants needed by later movement/combat work while keeping unsupported systems explicit.

Gate 2 was accepted on 2026-08-31 with one deliberate modernization: MOOX does not inherit the original six-design storage ceiling. No military Ship/Design gameplay code has started yet; Gate 3 is the next phase.


## Gate 3 implementation - complete

Gate 3 implements the accepted contract while keeping generic combat-Fleet movement, tactical weapons/specials, refit and Command-Point accounting outside this slice.

### Core schema 18

The authoritative persisted state now contains:

```text
GameState.ShipDesigns []ShipDesign
GameState.Ships       []Ship
StrategicFleet.ShipIDs []ID
```

`ShipDesign` is an ID-based, arbitrarily sized active design catalog per Empire. The original six `0x63` design slots remain reverse-engineering provenance only; there is deliberately no six-design MOOX validation ceiling. Dedicated regressions create eight and nine active designs and require successful validation/round-trip.

Each current design has a monotonic revision and one supported `ShipDesignSpec`. A concrete built `Ship` stores the source design ID/revision plus an independent copy of the supported design snapshot. Validation explicitly permits an older built source revision after the current design advances, while the Ship snapshot itself remains immutable unless a future explicit refit mechanic changes it.

Every concrete military Ship must belong to exactly one combat `StrategicFleet`; civilian/special fleets cannot contain concrete military `ShipIDs`. Existing abstract combat presence without concrete Ships remains valid for compatibility with the pre-slice blockade baseline.

### Normalized original hull and mandatory-system data

`ship_hulls.json` advances to ruleset schema 3. The normalizer now reads and validates the original executable tables for:

- all six hull base costs and spaces;
- Warp Drives and their installed FTL speeds;
- Computers;
- Armor;
- Shields;
- Fuel Cells and their strategic ranges.

This data is generated from the private MOO2 1.31 executable through `internal/moo2data` and committed as normalized ruleset data; it is not copied as opaque original binary content.

The directly proven hull values remain:

| Hull | Base PP | Space |
| --- | ---: | ---: |
| Frigate | 20 | 25 |
| Destroyer | 70 | 60 |
| Cruiser | 250 | 120 |
| Battleship | 600 | 250 |
| Titan | 1500 | 500 |
| Doom Star | 4000 | 1200 |

For the first exact cleared Frigate fixture, Nuclear Drive, Titanium Armor and Standard Fuel Cells add no PP/space in the supported subset, while Electronic Computer adds 5 PP. Therefore the directly supported baseline design is:

```text
Hull                    Frigate
Hull PP                 20
Electronic Computer      5
Weapons/Specials         0
Base design PP          25
Hull space              25
Used supported space     0
Normal production PP    25
Current Feudal PP       17  // ceil(2 * 25 / 3)
```

A known Shield is optional in the cleared design. Class I Shield on a Frigate contributes the normalized 3 PP / 5 space before any separately deferred miniaturization behavior.

Gate 1 deliberately deferred detailed miniaturization beyond what is required for the first design. Gate 3 therefore uses the normalized installed-component base tables for the currently supported mandatory-system subset and does not claim that advanced/intermediate technology states reproduce every original miniaturization discount yet.

### Authoritative design creation and revision

New command:

```text
empire.save_military_design {
    design_id?,
    name,
    hull_id,
    strategic_picture_id
}
```

For the accepted first surface:

- `hull_id` must be `frigate`;
- the strategic picture must be one of the original legal Frigate picture IDs;
- the resolver selects the best known supported Drive, Computer, Armor and Fuel Cell plus the best known Shield if available;
- weapons and special-system payloads do not exist yet, so unsupported tactical composition cannot be silently approximated;
- a new design gets a stable deterministic ID and revision 1;
- updating an existing design increments its revision;
- there is no design-count cap;
- an update is rejected while any Colony is actively constructing that design, preserving the accepted deterministic queue boundary.

Events:

```text
empire.military_design_created
empire.military_design_updated
```

### Military Construction

`ConstructionChoice` now carries military design ID/revision/name. Every currently owned active design is surfaced as a distinct authoritative `military_ship` choice; the original six-slot storage limit does not restrict this list.

New queue command:

```text
colony.queue_military_ship {
    colony_id,
    ship_design_id
}
```

The queue snapshots the exact current design ID/revision in `ConstructionState`. Cost resolution requires that exact revision still be authoritative. The normal Production progress path is reused rather than introducing a separate shipyard economy.

Events:

```text
colony.military_ship_queued
colony.construction_progressed
colony.military_ship_completed
```

### Completion -> concrete Ship -> combat Fleet

On completion the resolver:

1. resolves the exact queued current design revision;
2. allocates one deterministic concrete Ship ID;
3. copies the supported design snapshot into the Ship together with source design ID/revision and owner;
4. allocates one deterministic combat Fleet ID;
5. places that Fleet at the producing Colony's StarSystem;
6. stores the concrete Ship ID in `StrategicFleet.ShipIDs`.

The resulting Fleet is intentionally stationary. Slice 04 owns generic combat-Fleet movement, speed/range, merge and split.

A regression then updates the current design to a new revision/picture/name and proves that the already built Ship retains its old source revision, name and design snapshot, matching the original design-copy lifecycle established in Gate 1.

### Observer, save/load and deterministic replay

The GameSession vertical regression exercises the normal authoritative surface:

```text
Empire
-> save cleared Frigate design
-> Observer sees design and legal Construction choice
-> queue military_ship
-> normal Production completes 25-PP design
-> concrete Ship snapshot appears
-> one-Ship combat Fleet appears at source System
```

The returned Observer state is deliberately mutated in the test across `ShipDesigns`, `Ships` and Fleet `ShipIDs`; a fresh Observer read proves no mutation leaks into authority. The final schema-18 state survives exact JSON save/load. Running the same full GameSession scenario twice from the same deterministic fixture produces byte-identical final Core state and identical authoritative event history.

### Gate 3 regression result

Passed after removal of all temporary reverse-engineering/editor helpers:

```text
go test ./internal/game -run Military -count=1
go test ./internal/session -run Military -count=1
go test ./internal/core ./internal/ruleset ./internal/moo2data ./internal/game ./internal/session -count=1
go test ./... -count=1
git diff --check
```

Gate 4 remains intentionally open for a fresh conflict check, final gofmt/test/vet/focused compatibility pass, staged diff verification, documentation/HISTORY review, implementation commit and closing documentation commit that removes the `_OPEN_` marker.
## Gate 4 closure

- Final gofmt verification passed for all changed Go files.
- go test ./... -count=1 passed.
- go vet ./... passed.
- Focused Military/ShipDesign/Construction/StrategicFleet/ColonyShip/Outpost/Blockade/Observer/Save/Load regressions passed.
- Working and staged git diff --check passed.
- Gameplay implementation commit: 2670d36 (game: add military ship design baseline).
- Recovery marker removed in the closing documentation commit; Slice 04 is the next prepared objective.
