# Slice 17.1 Gate 1 - Hulls and Mandatory Ship Systems Audit

Date: 2026-09-19
Slice: 17.1 - Hulls and Mandatory Ship Systems
Gate: 1 - original/system audit
Baseline: `685691b` (`docs(slice): close slice 17.0 reference harness`)
Session: `ses-20260919T122602-63762fe1c392`

## Purpose

Re-audit the original-valid military hull and mandatory-system breadth before Gate 2 freezes the design contract.

This Gate is evidence/audit only. It does not widen Designer save legality, Colony buildability, Tactical combat or miniaturization behavior.

## Evidence classes

This audit distinguishes:

1. **direct normalized original data** - committed ruleset values generated from the private MOO2 1.31 executable/data tables and carrying source provenance;
2. **prior direct original-code audit** - reverse-engineering findings already preserved in repository research from the private 1.31 executable;
3. **official manual corroboration** - original MicroProse manual text available publicly;
4. **secondary technical corroboration** - community technical references retained as cross-checks where useful, but not used in place of the restored primary executable for miniaturization;
5. **current MOOX runtime** - what the repository actually enforces today.

The former `C:\ASH\Temp\mastori2\Orion2.exe` path is historical. The same hash-identified executable was re-located at `C:\ASH\Installation\mastori2\Orion2.exe` during this Gate, allowing the previously missing miniaturization routines to be freshly disassembled and closed as primary evidence.

## 1. Hull matrix

### Direct normalized original data

`data/rulesets/moo2-1.31/ship_hulls.json` is sourced from `moo2-1.31-orion2-ship-design-tables`.

| Hull | Size index | Base PP | Base Space | Command Points | original-table source offset |
| --- | ---: | ---: | ---: | ---: | ---: |
| Frigate | 0 | 20 | 25 | 1 | 32798 |
| Destroyer | 1 | 70 | 60 | 2 | 32834 |
| Cruiser | 2 | 250 | 120 | 3 | 32870 |
| Battleship | 3 | 600 | 250 | 4 | 32906 |
| Titan | 4 | 1500 | 500 | 5 | 32942 |
| Doom Star | 5 | 4000 | 1200 | 6 | 32978 |

Prior direct original audit in `MILITARY_SHIP_CORE_DESIGN_BASELINE_2026-08-31.md` established that `Total_Design_Cost_` and `Total_Design_Space_` index the original hull table by size index.

The values are hull-only. Original total design cost adds mandatory systems plus weapons/specials, then applies the government cost reducer.

### Command Point authority

Prior direct CP audit and current runtime agree that normal military Ship Command Point usage is:

`command_points = hull.size_index + 1`

The runtime derives this from the concrete built Ship's `HullID`; it does not persist a second mutable CP field in `ShipDesignSpec` or `Ship`.

This is already the correct architectural direction for 17.1.

## 2. Large-hull Colony buildability

### Prior direct original finding

The earlier executable audit identified `Colony_Can_Build_Large_Ships_` at original object-1 offset `0xD5275` / VA `0xE5275` and proved that large-hull legality checks Colony building state rather than only hull identity.

That earlier slice deliberately did not resolve the exact threshold.

### Official-manual corroboration

The original manual states that a planet without a Star Base cannot build ships larger than Medium, and separately states that Large ships or larger cannot be refitted without a Star Base or better orbital dock.

The Design Dock labels the six size classes in order:

`Small -> Medium -> Large -> Huge -> Titanic -> Death Star`

The normalized military hull order is:

`Frigate -> Destroyer -> Cruiser -> Battleship -> Titan -> Doom Star`

Therefore the Gate-1 buildability conclusion is:

| Hull | Orbital dock requirement |
| --- | --- |
| Frigate | none from this large-hull rule |
| Destroyer | none from this large-hull rule |
| Cruiser | Star Base or better |
| Battleship | Star Base or better |
| Titan | Star Base or better |
| Doom Star | Star Base or better |

Current MOOX has a normalized exclusive station family:

`Star Base -> Battlestation -> Star Fortress`

so “Star Base or better” can be evaluated as ownership of any one legal member of that family.

### Current MOOX gap

`AvailableConstructionChoices` currently adds every saved military `ShipDesign` owned by the Empire to every idle Colony without checking the design hull against Colony orbital infrastructure.

That behavior is harmless while only Frigate designs can be saved, but becomes wrong as soon as 17.1 enables Cruiser+.

Gate 3 will need authoritative buildability validation at both construction-choice projection and queue command resolution, not merely an HMI lock.

## 3. Titan and Doom Star technology locks

Current normalized technology data and current Designer catalog agree:

| Hull | Technology ID | Technology key | Field | Research cost |
| --- | ---: | --- | ---: | ---: |
| Titan | 186 | `titan_construction` | 19 Construction | 1150 RP |
| Doom Star | 55 | `doom_star_construction` | 58 Construction | 7500 RP |

Frigate, Destroyer, Cruiser and Battleship have no dedicated hull-unlock technology in the current original-valid catalog.

Consequences for the durable 17.0 labs:

- Baseline does not imply Titan/Doom Star.
- Mid-Tech's <=1150-RP profile includes Titan Construction.
- All-Tech includes both Titan and Doom Star Construction.
- infrastructure restrictions remain separate from technology ownership.

Current `MilitaryDesignerCatalog` already projects the Titan/Doom Star technology locks, but save legality remains Frigate-only due to `current_design_scope`.

## 4. Mandatory-system identity and technology catalog

Prior direct original audit established the normal design fields for Hull, Shield, Warp Drive, FTL speed, Computer and Armor, while `Build_Initial_Design_Template_` also selects Fuel Cells.

The committed schema validates exactly:

- 6 Drives;
- 5 Computers;
- 6 Armors;
- 5 Shields;
- 5 Fuel Cells.

### Drives

Direct normalized original values:

| Drive | Tech ID | Intro field / RP | Strategic FTL | Base cost by hull | Base space by hull |
| --- | ---: | --- | ---: | --- | --- |
| Nuclear | 120 | 55 / 50 | 2 | all 0 | all 0 |
| Fusion | 72 | 5 / 250 | 3 | all 0 | all 0 |
| Ion | 96 | 41 / 900 | 4 | all 0 | all 0 |
| Anti-Matter | 11 | 13 / 2000 | 5 | all 0 | all 0 |
| Hyper | 88 | 38 / 4500 | 6 | all 0 | all 0 |
| Interphased | 95 | 40 / 10000 | 7 | all 0 | all 0 |

All rows carry `moo2-1.31-orion2-ship-design-tables` provenance.

### Computers

Direct normalized original base costs:

| Computer | Tech ID | Intro field / RP | Base PP Frg/Des/Cru/BB/Titan/DS |
| --- | ---: | --- | --- |
| Electronic | 58 | 28 / 50 | 5 / 15 / 50 / 125 / 300 / 800 |
| Optronic | 122 | 56 / 150 | 5 / 15 / 50 / 125 / 300 / 800 |
| Positronic | 143 | 60 / 900 | 5 / 15 / 50 / 125 / 300 / 800 |
| Cybertronic | 44 | 25 / 2750 | 5 / 15 / 50 / 125 / 300 / 800 |
| Moleculartronic | 110 | 49 / 6000 | 5 / 15 / 50 / 125 / 300 / 800 |

The normalized mandatory-component schema does not assign optional hull Space to Computers.

### Armor

| Armor | Tech ID | Intro field / RP | normalized base cost % |
| --- | ---: | --- | ---: |
| Titanium | 187 | 22 / 50 | 0 |
| Tritanium | 191 | 9 / 250 | 0 |
| Zortrium | 203 | 53 / 2000 | 0 |
| Neutronium | 117 | 50 / 4500 | 0 |
| Adamantium | 2 | 48 / 10000 | 0 |
| Xentronium | 201 | 74 / 15000 | 0 |

The prior direct original audit also established Titanium as the baseline component and the higher Best-Armor sequence.

### Shields

All normalized Shield rows share the same unminiaturized hull tables:

- Space: `5 / 10 / 20 / 50 / 100 / 200`
- PP: `3 / 5 / 10 / 25 / 50 / 100`

| Shield | Tech ID | Intro field / RP |
| --- | ---: | --- |
| Class I | 33 | 7 / 250 |
| Class III | 34 | 45 / 900 |
| Class V | 35 | 64 / 2750 |
| Class VII | 36 | 61 / 4500 |
| Class X | 37 | 68 / 15000 |

### Fuel Cells

| Fuel Cell | Tech ID | Intro field / RP | Strategic range |
| --- | ---: | --- | ---: |
| Standard | 167 | 22 / 50 | 4 pc |
| Deuterium | 51 | 9 / 250 | 6 pc |
| Iridium | 98 | 47 / 1150 | 9 pc |
| Urridium | 194 | 50 / 4500 | 12 pc |
| Thorium | 184 | 48 / 10000 | 255 (runtime unlimited-range sentinel) |

Fuel cells carry no normalized optional Space/PP table in the mandatory-system schema.

## 5. Original Automatics semantics matter for the Gate-2 UI contract

The original manual's Design Dock describes Engines, Shields, Armor, Fuel Cells and Computer as “Automatics”.

The important distinctions are:

- the best available system is installed by default;
- **Shield and Computer may be deliberately changed to a lesser available system** to save build cost;
- **Drive and Fuel Cell capability updates all existing ships when improved technology is acquired**, without a refit;
- Engines, Armor and Fuel Cells do not consume optional design Space.

This means 17.1 should not blindly model all five families as identical freely selectable component slots.

Gate 2 must freeze a category-specific selection contract.

The evidence currently supports:

- Drive: automatic/current-best strategic capability;
- Fuel Cell: automatic/current-best strategic capability;
- Armor: automatic best available for a newly authored/refitted design;
- Computer: selectable among legal known levels, default best;
- Shield: optional/selectable among legal known levels, default best when using the original auto-template behavior.

The exact UX for “no shield” versus a lesser shield still needs an explicit Gate-2 freeze. MOOX already represents no Shield as empty `ShieldID`, and prior direct audit proved a cleared design with empty weapon/special arrays is legal, but Gate 1 will not invent an unsupported original UI rule.

## 6. Strategic semantics: Drive/Fuel current-tech behavior is already aligned

MOOX's accepted Slice-04 movement contract derives ordinary combat-fleet movement from the Empire's **current best** Drive and Fuel Cell rather than pinning movement to the built Ship's historical snapshot.

`combatFleetMovementProfile` currently derives:

- FTL = best currently known Warp Drive, plus Trans-Dimensional when present;
- range = current best Fuel Cell range for every current military member.

Built `ShipDesignSpec.FTLSpeed` and `FuelRangeParsecs` remain immutable design/build provenance.

The original manual's explicit statement that Drives and Fuel Cells update existing ships without refit corroborates this architecture.

Therefore 17.1 must **not** regress strategic movement to historical per-Ship Drive/Fuel capability.

## 7. Miniaturization audit

### Restored primary executable evidence

The private reference executable is available at:

`C:\ASH\Installation\mastori2\Orion2.exe`

SHA-256:

`7AE2AC2E5904CA330009AF2827279D889906B0B9B7A8854C38EB707A56E955B5`

This is byte-identical to the MOO2 1.31 executable used by the earlier direct audits. A fresh LE-object disassembly was therefore performed for the missing miniaturization evidence. Temporary extracts/disassembly helpers were removed after the audit.

Relevant object-1 routines/call sites:

- `0x5E70A` - validates the player/technology and dispatches to the miniaturization-level calculation for an owned component technology;
- `0x5D048` - walks the technology field successor chain and calculates miniaturization level;
- `0x5B519` - applies production-cost miniaturization;
- `0x5E60E` - applies Space miniaturization;
- `Design_Template_Costs_` `0x5C087` calls the level helper and cost reducer for Shields and Computers;
- `Design_Template_Space_Requirements_` `0x5C287` calls the level helper and Space reducer; the Shield path at `0x5C38D..0x5C39A` explicitly clears `EBX`, selecting Space class 0 below.

### Exact miniaturization-level derivation

The original routine starts from the installed technology's research field and follows that field's **NextID** chain, not its PreviousID chain. This was cross-checked directly against object-2 technology data: Electronic Computer technology 58 maps to field 28 and the word read by the routine at the successor-table location is 56, exactly normalized field 28's `next_id`.

For each ordinary successor field below 75:

- player field state byte `player + 0xC4 + field_id` must equal `3`;
- each consecutive completed successor adds one miniaturization level;
- traversal stops at the first successor field that is not complete or at field 0.

When the successor is a Hyper-Advanced field 75..82, the routine adds the repeat counter at:

`player + 0x21C + (field_id - 75)`

and then stops.

Therefore the exact effective level is:

`consecutive completed later ordinary fields + Hyper-Advanced repeat count at the chain end`

The reduction helpers saturate at level 5+, so additional Hyper-Advanced repeats beyond the saturation point do not reduce cost/Space further.

### Exact production-cost reducer

Object-1 `0x5B519` selects these integer percentages:

| Mini level | Cost factor |
| ---: | ---: |
| 0 | 100% |
| 1 | 75% |
| 2 | 55% |
| 3 | 40% |
| 4 | 30% |
| 5+ | 25% |

For a positive base cost the original performs integer multiply then signed integer division by 100, which is floor/truncation for these positive values, and then forces a minimum result of 1. Zero base cost stays zero.

Exact positive-cost rule:

`mini_cost = max(1, floor(base_cost * percent / 100))`

This reducer is directly used by the mandatory Shield cost path at `0x5C0D1..0x5C0DC` and the Computer cost path at `0x5C267..0x5C272`.

### Exact Space reducer

Object-1 `0x5E60E` contains two miniaturizing Space classes plus a no-reduction fallback. Factors are stored as per-thousand integers.

| Mini level | Space class 0 | Space class 1 |
| ---: | ---: | ---: |
| 0 | 1000 / 1000 | 1000 / 1000 |
| 1 | 800 / 1000 | 800 / 1000 |
| 2 | 650 / 1000 | 700 / 1000 |
| 3 | 500 / 1000 | 600 / 1000 |
| 4 | 350 / 1000 | 500 / 1000 |
| 5+ | 250 / 1000 | 400 / 1000 |

The exact positive-Space calculation is:

`mini_space = max(1, floor((base_space * factor + 500) / 1000))`

Thus the original uses nearest-integer rounding with the half boundary rounded upward for positive values. It does **not** use a generic ceiling rule.

The mandatory Shield path explicitly supplies Space class 0, so Shields use the 100/80/65/50/35/25-percent Space sequence. Computer has no optional Space table in the normalized mandatory-system schema. Current Drive Space is zero, so Drive Space miniaturization has no numerical effect on the normalized original row.

### Direct golden vectors for Gate 2/3

These vectors follow the executable exactly and are suitable for later regression tests.

Electronic Computer on a Cruiser has base cost 50 PP:

| Mini level | PP |
| ---: | ---: |
| 0 | 50 |
| 1 | 37 |
| 2 | 27 |
| 3 | 20 |
| 4 | 15 |
| 5+ | 12 |

Class I Shield on a Cruiser has base cost 10 PP and base Space 20:

| Mini level | PP | Space |
| ---: | ---: | ---: |
| 0 | 10 | 20 |
| 1 | 7 | 16 |
| 2 | 5 | 13 |
| 3 | 4 | 10 |
| 4 | 3 | 7 |
| 5+ | 2 | 5 |

### Gate-1 conclusion on miniaturization

The prior secondary evidence is now superseded where it was ambiguous. For the 17.1 mandatory-system surface, Gate 1 now has primary executable evidence for:

- level derivation through successor fields;
- Hyper-Advanced repeat contribution;
- saturation at level 5+;
- exact cost percentages;
- exact Shield Space percentages;
- exact integer rounding/minimum behavior;
- direct Shield and Computer call-site application.

Gate 2 therefore no longer needs to guess or choose between 0.65/0.66, 0.35/0.36, floor/ceil variants, or Hyper-Advanced handling. It only needs to freeze the implementation contract around these now-direct rules.

## 8. Derived strategic/tactical stats

### Strategic - supported/authoritative now

- Hull -> Command Point usage: `size_index + 1`.
- Drive -> strategic FTL 2/3/4/5/6/7.
- Fuel -> strategic range 4/6/9/12/unlimited.
- Drive/Fuel improvement applies to existing strategic combat ships through current Empire capability.
- Design `FTLSpeed` and `FuelRangeParsecs` remain immutable source-design provenance.

### Tactical - original evidence, but downstream runtime support

Secondary technical/original-manual corroboration identifies these original effects:

Computers:
- Electronic +25 Beam Attack;
- Optronic +50;
- Positronic +75;
- Cybertronic +100;
- Moleculartronic +125.

Armor material durability multiplier:
- Titanium 1x;
- Tritanium 2x;
- Zortrium 4x;
- Neutronium 6x;
- Adamantium 8x;
- Xentronium 10x, with additional AP interaction.

Shields:
- Class I: block 1 per hit, capacity 5 x hull size class;
- Class III: block 3, capacity 15 x;
- Class V: block 5, capacity 25 x;
- Class VII: block 7, capacity 35 x;
- Class X: block 10, capacity 50 x;
- shield recharge/facing are Tactical mechanics.

Drive combat-speed contribution is distinct from strategic FTL. Secondary technical evidence gives drive increments of +0/+2/+4/+6/+8/+10 on top of hull/load-dependent combat speed.

### Current Tactical runtime boundary

Current MOOX Tactical handoff intentionally accepts only the existing baseline:

- Frigate;
- Electronic Computer;
- Titanium Armor;
- no Shield;
- Standard Fuel;
- currently normalized Tactical drive subset;
- supported Laser baseline.

The Tactical ruleset currently has only:

- Frigate Titanium armor/structure baseline;
- Electronic Computer Beam Offense;
- Nuclear/Fusion tactical drive speed rules.

Therefore Gate 1 classifies broad Computer accuracy, Armor durability, Shield arcs/capacity/recharge and broader drive combat-speed effects as **downstream Tactical implementation**.

17.1 may freeze Designer-visible derived values and preserve component identity, but it must not fake unsupported Tactical battle behavior merely to make a richer design saveable.

## 9. Current MOOX runtime gaps

### Hulls

Catalog projects all six, but `SaveAvailable` is hard-limited to the supported Frigate.

`clearedMilitaryDesignSpec` rejects every non-Frigate hull.

### Technology locks

Titan/Doom Star locks are already projected correctly by key.

### Mandatory systems

The save/compiler path currently auto-selects:

- best known Drive;
- best known Computer;
- best known Armor;
- best known Fuel Cell;
- best known Shield if available.

It does not yet support the original-valid “choose a lesser Computer/Shield” surface.

### Costs/Space

The compiler uses normalized unminiaturized base tables only.

It does not yet apply full original miniaturization.

### Colony Construction

Every owned saved military design is currently offered at every idle Colony.

There is no Cruiser+ Star-Base-or-better legality check yet.

### Tactical

Broad mandatory-system effects are intentionally unsupported, and current validation rejects them rather than approximating them.

This rejection is correct until a downstream Tactical contract is frozen.

## 10. Gate-2 decisions now ready to freeze

Gate 2 should freeze all of the following explicitly:

1. hull availability:
   - Frigate/Destroyer/Cruiser/Battleship base availability;
   - Titan tech 186;
   - Doom Star tech 55;

2. producing-Colony infrastructure:
   - Cruiser+ requires Star Base/Battlestation/Star Fortress;
   - Frigate/Destroyer do not;

3. category-specific mandatory-system selection:
   - Drive automatic;
   - Fuel automatic;
   - Armor automatic best for authored/refitted design;
   - Computer default best but selectable lower;
   - Shield default best but selectable lower/none according to the frozen UI contract;

4. strategic semantics:
   - current Empire Drive/Fuel governs movement of existing ships;
   - design snapshot keeps build provenance;

5. CP:
   - derive from Hull size index + 1;

6. unminiaturized original base tables:
   - six hulls;
   - six Drives;
   - five Computers;
   - six Armors;
   - five Shields;
   - five Fuel Cells;

7. miniaturization:
   - freeze the directly verified successor-field/Hyper-Advanced level derivation;
   - freeze exact cost and Space integer factors/rounding from object-1 routines;
   - adopt the direct Computer and Shield golden vectors;

8. derived stat representation:
   - which values are projected/derived in Designer;
   - avoid duplicating mutable derived values in persisted `ShipDesignSpec` unless required for immutable build semantics;

9. Tactical honesty:
   - richer 17.1 designs may be valid strategic/constructible Ships;
   - Tactical launch remains rejected for unsupported equipment until the appropriate Tactical breadth is implemented.

## 11. Reference-lab acceptance fixtures for later gates

The durable 17.0 labs naturally provide:

- Baseline Triangle -> early mandatory systems;
- Mid-Tech Triangle -> Titan Construction plus mid-tier mandatory systems;
- All-Tech Triangle -> Doom Star plus all research-addressable mandatory technologies.

Gate 2 should freeze exact representative Designer/Construction fixtures using these existing worlds rather than creating new throwaway scenarios.

## Source register

### Primary executable re-audit (2026-09-19)

- current private path: `C:\ASH\Installation\mastori2\Orion2.exe`;
- SHA-256 `7AE2AC2E5904CA330009AF2827279D889906B0B9B7A8854C38EB707A56E955B5`;
- LE object-1 miniaturization routines re-audited at `0x5D048`, `0x5B519`, `0x5E60E`, `0x5E70A`, with direct call-site checks in `Design_Template_Costs_` / `Design_Template_Space_Requirements_`.

### Repository/direct evidence

- `data/rulesets/moo2-1.31/ship_hulls.json`
- `data/rulesets/moo2-1.31/technologies.json`
- `internal/ruleset/ship_hulls.go`
- `internal/game/military_designer_catalog.go`
- `internal/game/military_design.go`
- `internal/game/buildability.go`
- `internal/game/command_points.go`
- `internal/game/combat_fleet.go`
- `internal/game/tactical_combat.go`
- `internal/session/session.go`
- `docs/research/MILITARY_SHIP_CORE_DESIGN_BASELINE_2026-08-31.md`
- `docs/research/COMMAND_POINTS_SHIP_MAINTENANCE_2026-08-31.md`
- `docs/research/COMBAT_FLEET_MOVEMENT_MERGE_SPLIT_2026-08-31.md`

### Official manual corroboration

- MicroProse Master of Orion II manual mirror, Design Dock / Automatics / Refit:
  - https://manualzz.com/doc/66196384/microprose-master-of-orion-ii-n%C3%A1vod-k-pou%C5%BEit%C3%AD
- ManualShelf original manual, Star Base:
  - https://www.manualshelf.com/manual/games-pc/master-of-orion-ii-battle-at-antares/user-guide-english/page-69.html
- ManualShelf original manual, Design Dock:
  - https://www.manualshelf.com/manual/games-pc/master-of-orion-ii-battle-at-antares/user-guide-english/page-111.html

### Secondary technical corroboration

- StrategyWiki Warship Technologies:
  - https://strategywiki.org/wiki/Master_of_Orion_II%3A_Battle_at_Antares/Warship_technologies
- StrategyWiki Computers:
  - https://strategywiki.org/wiki/Master_of_Orion_II%3A_Battle_at_Antares/Computers
- StrategyWiki Force Fields:
  - https://strategywiki.org/wiki/Master_of_Orion_II%3A_Battle_at_Antares/Force_Fields
- GameFAQs Hronk strategy guide:
  - https://gamefaqs.gamespot.com/pc/197873-master-of-orion-ii-battle-at-antares/faqs/72470
- GameFAQs weapon-calculation technical guide:
  - https://gamefaqs.gamespot.com/pc/197873-master-of-orion-ii-battle-at-antares/faqs/16743
- humbe.no MOO2 reference hull table:
  - https://www.humbe.no/public/computergames/moo2/

## Gate 1 result

The audit is sufficient to proceed to Gate 2.

The major original-valid boundary is now clear:

- six authoritative hulls with CP 1..6;
- Cruiser+ requires an orbital dock;
- Titan/Doom Star add explicit technology locks;
- Computer and Shield are the user-adjustable mandatory families;
- Drive/Fuel are current-Empire strategic capabilities that upgrade existing ships;
- Armor is an automatic best-known design material;
- only Computer cost and Shield cost/Space currently require mandatory-family miniaturization work;
- broad Tactical effects remain downstream and must not be faked.

The previously missing miniaturization evidence has now been closed from the restored, hash-identical private 1.31 executable. Gate 1 has no remaining material evidence gap for mandatory-system cost/Space miniaturization; Gate 2 may freeze the direct-original contract without relying on secondary rounding assumptions.
