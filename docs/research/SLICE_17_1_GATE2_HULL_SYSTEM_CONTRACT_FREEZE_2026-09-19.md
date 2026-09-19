# Slice 17.1 Gate 2 - Hull/System Contract Freeze

Date: 2026-09-19
Slice: 17.1 - Hulls and Mandatory Ship Systems
Gate: 2 - contract freeze
Baseline: `bbc8795` (`docs(slice): verify 17.1 miniaturization from 1.31 exe`)
Gate-1 evidence: `docs/research/SLICE_17_1_GATE1_HULLS_MANDATORY_SYSTEMS_AUDIT_2026-09-19.md`

## Purpose

Freeze the exact 17.1 gameplay/API/QA contract before implementation.

Gate 3 must implement this contract without expanding into Slice 17.2 weapons, 17.3 ordnance, 17.4 specials/fighters or broad Tactical combat fidelity.

## 1. Hull availability contract

All six normalized military hulls are in 17.1:

| Hull | Size | Base PP | Base Space | CP | Designer tech lock |
| --- | ---: | ---: | ---: | ---: | --- |
| Frigate | 0 | 20 | 25 | 1 | none |
| Destroyer | 1 | 70 | 60 | 2 | none |
| Cruiser | 2 | 250 | 120 | 3 | none |
| Battleship | 3 | 600 | 250 | 4 | none |
| Titan | 4 | 1500 | 500 | 5 | tech 186 `titan_construction` |
| Doom Star | 5 | 4000 | 1200 | 6 | tech 55 `doom_star_construction` |

### Designer availability is not Colony buildability

A legal Empire may save a Cruiser/Battleship/Titan/Doom-Star design even when its current Colony has no orbital dock.

Designer save legality depends on:

- hull exists in the normalized six-hull catalog;
- hull technology is known when required;
- legal strategic picture for that hull;
- legal mandatory-system selection;
- legal current 17.1 weapon baseline;
- effective Space does not exceed hull Space.

It does **not** depend on a Colony being selected or having a Star Base.

### Colony buildability

Military-Ship construction is separately constrained:

- Frigate and Destroyer: no command-station requirement from the large-hull rule;
- Cruiser, Battleship, Titan and Doom Star: require command-station tier >= 1;
- tier 1 = `star_base`;
- tier 2 = `battlestation`;
- tier 3 = `star_fortress`.

The rule must be enforced in both:

1. `AvailableConstructionChoices` projection;
2. authoritative `queueMilitaryShip` validation.

A stale/forged queue command must never bypass the Colony rule.

## 2. Mandatory-system selection contract

17.1 does **not** model Drive, Computer, Armor, Shield and Fuel as five equivalent user-selectable slots.

### Warp Drive - automatic/current Empire capability

Normalized sequence:

`nuclear_drive -> fusion_drive -> ion_drive -> anti_matter_drive -> hyper_drive -> interphased_drive`

Technology IDs:

`120, 72, 96, 11, 88, 95`

Designer default/effective Drive is current best known.

Strategic movement continues to use the current Empire Drive, including existing Ships. A saved historical `WarpDriveID` / `FTLSpeed` remains provenance, not strategic movement authority.

The save payload does not accept a user-selected Drive ID in 17.1.

### Fuel Cell - automatic/current Empire capability

Normalized sequence:

`standard_fuel_cells -> deuterium_fuel_cells -> iridium_fuel_cells -> urridium_fuel_cells -> thorium_fuel_cells`

Technology IDs:

`167, 51, 98, 194, 184`

Strategic ranges:

`4, 6, 9, 12, unlimited(255)`

Designer default/effective Fuel is current best known.

Strategic movement continues to use the current Empire Fuel capability for existing Ships. Historical `FuelCellID` / `FuelRangeParsecs` in a built snapshot are provenance.

The save payload does not accept a user-selected Fuel Cell ID in 17.1.

### Armor - automatic at explicit design/refit save

Normalized sequence:

`titanium_armor -> tritanium_armor -> zortrium_armor -> neutronium_armor -> adamantium_armor -> xentronium_armor`

Technology IDs:

`187, 191, 203, 117, 2, 201`

A new design or explicit revision/save receives the best currently known Armor.

Armor is then part of the saved design selection/snapshot. Researching a newer Armor does **not** silently rewrite an existing saved design or already-built Ship. Re-saving/refitting the design applies the current best Armor.

The save payload does not accept a user-selected Armor ID in 17.1.

### Computer - explicit selectable family

Normalized sequence:

`electronic_computer -> optronic_computer -> positronic_computer -> cybertronic_computer -> moleculartronic_computer`

Technology IDs:

`58, 122, 143, 44, 110`

Rules:

- default for a new design = best currently known Computer;
- user may select any lower or higher Computer whose technology is known;
- unknown/unowned Computer IDs are rejected;
- chosen Computer remains selected on the saved design until an explicit edit changes it;
- Computer PP cost is miniaturized from current research state whenever an effective saved-design preview or a new construction snapshot is compiled.

### Shield - optional explicit selectable family

Normalized sequence:

`class_i_shield -> class_iii_shield -> class_v_shield -> class_vii_shield -> class_x_shield`

Technology IDs:

`33, 34, 35, 36, 37`

Rules:

- a new design defaults to the best currently known Shield;
- if no Shield technology is known, default is no Shield;
- user may select any known Shield;
- user may explicitly select **no Shield** even when Shield technology is available;
- selected Shield remains selected until an explicit edit changes it;
- Shield PP and Space are miniaturized from current research state whenever an effective saved-design preview or a new construction snapshot is compiled.

## 3. Save-command compatibility contract

The existing strict `SaveMilitaryDesignPayload` gains only the user-selectable mandatory fields needed by 17.1.

Frozen semantics:

- `computer_id`:
  - omitted/empty = compatibility default to best currently known Computer;
  - non-empty = exact requested Computer and must be known/legal.

- `shield_id`:
  - field absent = compatibility default to best currently known Shield, or none if no Shield is known;
  - field present with empty string = explicit **no Shield**;
  - field present with a non-empty ID = exact requested Shield and must be known/legal.

Implementation consequence: Go must preserve the absent-vs-present-empty distinction for `shield_id` (for example with a pointer/optional representation). JSON `null` is treated the same as absent/automatic; the HMI uses an explicit empty string for “None”.

No `warp_drive_id`, `armor_id` or `fuel_cell_id` selection field is accepted. Strict decoding continues to reject unknown payload fields.

When editing an existing design, the official HMI always submits its current explicit Computer and Shield selection so that an edit does not accidentally reset a deliberately downgraded component.

## 4. Catalog/projection contract

The authoritative Ship Designer catalog must expose enough server-derived data for the HMI to render all 17.1 choices without re-implementing research or miniaturization rules in TypeScript.

The catalog must expose:

- all six hulls with current technology availability and CP;
- all normalized Computers with:
  - technology ID;
  - known/available flag;
  - stable lock reason;
  - current miniaturization level;
  - effective PP by hull;
- all normalized Shields with:
  - technology ID;
  - known/available flag;
  - stable lock reason;
  - current miniaturization level;
  - effective PP by hull;
  - effective Space by hull;
- current automatic Drive;
- current automatic Fuel Cell;
- Armor that will be used on a new/refitted save;
- current production-cost government ratio.

The HMI may calculate a live preview by summing these **already server-derived effective values**, but it must not independently derive research-chain miniaturization.

The existing `current_design_scope` lock must disappear for the six 17.1 hulls. It may remain a generic future-scope reason elsewhere, but it is not a valid 17.1 hull lock after Gate 3.

## 5. Direct-original miniaturization contract

Primary evidence is the hash-identical MOO2 1.31 executable:

`C:\ASH\Installation\mastori2\Orion2.exe`

SHA-256:

`7AE2AC2E5904CA330009AF2827279D889906B0B9B7A8854C38EB707A56E955B5`

### Level calculation

For the selected component technology:

1. find its technology field;
2. follow that field's `NextID` chain;
3. for each ordinary successor field < 75 that is present in `Empire.KnownTechnologyFieldIDs`, add one level and continue;
4. stop on the first ordinary successor not known;
5. if the chain reaches Hyper-Advanced field 75..82, add that field's `Empire.HyperAdvancedResearch.CompletedLevels` and stop;
6. reduction helpers saturate at level 5+.

This normalized representation corresponds to the original field-state == 3 and Hyper-Advanced repeat-counter logic.

### Production-cost miniaturization

Exact factors:

| level | factor |
| ---: | ---: |
| 0 | 100% |
| 1 | 75% |
| 2 | 55% |
| 3 | 40% |
| 4 | 30% |
| 5+ | 25% |

For a positive component base cost:

`effective = max(1, floor(base * percent / 100))`

A zero base remains zero.

In 17.1 this directly affects Computer PP and Shield PP.

### Space miniaturization

The reusable original helper has two classes:

| level | class 0 | class 1 |
| ---: | ---: | ---: |
| 0 | 1000 | 1000 |
| 1 | 800 | 800 |
| 2 | 650 | 700 |
| 3 | 500 | 600 |
| 4 | 350 | 500 |
| 5+ | 250 | 400 |

For positive Space:

`effective = max(1, floor((base * factor + 500) / 1000))`

This is nearest-integer rounding with positive half values rounded upward.

The 17.1 Shield call site uses **class 0**, therefore Shield Space is:

`100 / 80 / 65 / 50 / 35 / 25 %`

Computer has no optional Space table. Current normalized Drive Space is zero.

### Application order

Miniaturization is applied **per component before summing** design cost/Space.

17.1 does not miniaturize the hull base cost.

17.1 does not introduce weapon miniaturization; that belongs to the later weapon slice.

After base design cost is assembled, the existing government production modifier remains authoritative. For Feudal this remains `ceil(2 * base / 3)`.

## 6. Saved design vs effective design vs construction snapshot

This distinction is frozen because it is required for both original behavior and deterministic persistence.

### Saved design selection

A saved design/revision preserves the explicit design choices:

- Hull;
- strategic picture;
- Computer;
- Armor chosen at that explicit save/refit;
- Shield or no Shield;
- current supported weapon mounts;
- visual genome/revision through the existing visual path.

Drive/Fuel historical fields may remain persisted for provenance/backward compatibility, but they are not strategic movement authority.

### Effective saved-design projection

Primary 1.31 code re-audit shows `Cost_Of_Design_` rebuilds a working design and recalculates its cost using current player technology.

Fresh object-1 verification for this Gate: `Cost_Of_Design_` at relative `0x5B577` copies the stored design selections into the working design and invokes the normal recalculation path; a colony-path call site at `0xD3AEC` calls `Cost_Of_Design_` and materializes the returned cost. The original saved design record therefore does not carry a forever-authoritative pre-miniaturized production-cost scalar.

Therefore current effective cost/Space for a saved design is **derived**, not permanently frozen at the research state of the last save.

Research progression that only changes miniaturization:

- does not increment the saved design revision;
- does not alter its Computer/Shield selection;
- can reduce its projected effective PP and Shield Space;
- can change its current effective Drive/Fuel projection via the existing global-auto semantics;
- does not mutate already-built Ships.

The server, not the HMI, owns this recompilation.

### Queue-time construction snapshot

When a Colony successfully queues a military design, the server compiles the design against the **current** Empire state and freezes a construction snapshot.

The queued project must persist enough information to make these values stable for the life of that construction:

- design ID;
- design revision;
- effective compiled `ShipDesignSpec`;
- effective `ProductionCostPP`.

Later research must not retroactively change the PP target of an already-running project.

This is also the deterministic persistence boundary: save/restore during construction must resume with the exact same target and completion Ship snapshot.

### Built Ship snapshot

Completion copies the queued compiled spec into the concrete `Ship`, together with existing source design/revision and visual provenance.

Later edits to the saved design, later miniaturization or later research do not mutate that built Ship snapshot.

Current strategic Drive/Fuel capability may still govern movement of the built Ship, as already accepted in the movement contract.

## 7. CP contract

No mutable Command-Point cost is persisted in `ShipDesignSpec`.

Military Ship CP usage is always derived:

`CP = hull.size_index + 1`

Therefore:

`Frigate..Doom Star = 1,2,3,4,5,6 CP`

Construction/project/UI projections may display CP, but state authority remains the hull ID/normalized hull size.

## 8. Stable lock/rejection reasons

The following machine reason tokens are frozen for 17.1 validation/catalog tests:

- `technology_required` - required hull/component technology is not known;
- `mandatory_systems_required` - no legal automatic mandatory baseline can be compiled;
- `invalid_component_selection` - supplied Computer/Shield ID is unknown or belongs to the wrong family;
- `invalid_strategic_picture` - picture is not legal for the selected hull;
- `space_exceeded` - effective design Space exceeds hull Space;
- `orbital_dock_required` - forced Cruiser+ construction at a Colony without Star Base/Battlestation/Star Fortress;
- `design_revision_in_use` - explicit design revision attempted while that design revision is under construction.

For HTTP immediate-command transport, the existing outer `session_rejected` envelope may remain. Gate 3 must keep these inner reason tokens deterministic in the domain error/validation path and tests; the HMI must prefer catalog availability/lock data over parsing human text.

Existing generic `current_design_scope` is not a valid post-17.1 hull rejection.

## 9. Reference-Lab fixture freeze

No new throwaway world is added. Slice 17.0 Baseline/Mid-Tech/All-Tech Triangle games remain the durable laboratories.

### Baseline - early game

Game:

`game-triangle-2pc` / profile `baseline`

Expected automatic baseline currently verified live:

- Nuclear Drive;
- Electronic Computer default;
- Titanium Armor;
- no Shield;
- Standard Fuel Cells.

Required Gate-3 acceptance:

- Frigate and Destroyer save normally;
- Cruiser/Battleship also save because they have no technology lock;
- stationless Colony does not offer/accept Cruiser+ construction;
- Frigate no-weapon design remains a valid narrow fixture.

Baseline Frigate no-weapon effective base:

- hull 20 PP;
- Electronic Computer 5 PP;
- no Shield;
- base design cost 25 PP;
- SpaceUsed 0;
- CP 1.

### Mid-Tech - unlock + non-saturated miniaturization

Game:

`game-triangle-mid-tech` / profile `mid_tech`

Live profile expectations already verified:

- Ion Drive;
- Computers through Positronic;
- Tritanium Armor;
- Shields through Class III;
- Iridium Fuel;
- Titan Construction known;
- Doom Star Construction not known.

Frozen deliberate downgrade vector:

**Cruiser + Electronic Computer + Class I Shield**

Research-chain mini levels:

- Electronic Computer = 3;
- Class I Shield = 2.

Expected mandatory-only effective values:

- hull = 250 PP / 120 Space;
- Electronic Computer = `floor(50 * 40 / 100) = 20 PP`;
- Class I Shield = `floor(10 * 55 / 100) = 5 PP`;
- Class I Shield Space = `floor((20 * 650 + 500) / 1000) = 13`;
- base design cost = 275 PP;
- SpaceUsed = 13;
- CP = 3.

Titan must be Designer-unlocked in Mid-Tech. Doom Star remains `technology_required`.

### All-Tech - late game + saturated miniaturization

Game:

`game-triangle-all-tech` / profile `all_tech`

Live profile expectations already verified:

- Interphased Drive;
- Computers through Moleculartronic;
- Xentronium Armor;
- Shields through Class X;
- Thorium Fuel;
- Titan and Doom Star Construction known.

Frozen saturated downgrade vector:

**Cruiser + Electronic Computer + Class I Shield**

Both component levels exceed the level-5 saturation boundary.

Expected mandatory-only effective values:

- Electronic Computer = `floor(50 * 25 / 100) = 12 PP`;
- Class I Shield = `floor(10 * 25 / 100) = 2 PP`;
- Class I Shield Space = `floor((20 * 250 + 500) / 1000) = 5`;
- base design cost = `250 + 12 + 2 = 264 PP`;
- SpaceUsed = 5;
- CP = 3.

Late-game breadth fixture:

**Doom Star + default top systems**

Expected component identity:

- Interphased Drive;
- Moleculartronic Computer;
- Xentronium Armor;
- Class X Shield by default;
- Thorium Fuel.

Doom Star must save at Turn 1 in All-Tech without research grinding. Construction still requires an orbital dock.

A second All-Tech edit must explicitly select **no Shield** and persist that choice.

## 10. Gate-3 implementation obligations

Gate 3 must implement, at minimum:

1. remove the Frigate-only save gate and enable all six frozen hulls subject to tech;
2. add Computer/Shield selection semantics to command + catalog + HMI;
3. preserve automatic Drive/Fuel and save-time automatic Armor semantics;
4. implement the direct-original miniaturization helper and golden vectors;
5. project effective saved-design costs from current research without changing revision;
6. create/persist queue-time compiled military construction snapshots;
7. enforce Cruiser+ orbital-dock legality in projection and queue authority;
8. preserve CP derivation and immutable built Ship snapshots;
9. keep the current Laser/no-weapon weapon baseline only; do not widen 17.2 scope;
10. keep Tactical rejection honest for currently unsupported late mandatory systems.

## 11. Gate-3 required tests

Focused tests must cover:

- six-hull catalog order/base values/CP;
- Titan and Doom Star technology locks;
- all Computer/Shield tech availability;
- absent Computer = best-known compatibility default;
- absent Shield = best-known compatibility default;
- explicit empty Shield = no Shield;
- unknown/unowned Computer/Shield rejection;
- Mid-Tech 275 PP / 13 Space Cruiser golden vector;
- All-Tech 264 PP / 5 Space Cruiser golden vector;
- miniaturization changes effective saved-design projection without revision increment;
- queue snapshot remains unchanged after later research;
- built Ship deep-copy remains unchanged after later design edit/research;
- Frigate/Destroyer construction without station;
- Cruiser+ rejection without station;
- Cruiser+ acceptance with Star Base, Battlestation and Star Fortress;
- persistence round trip of an in-progress military construction snapshot;
- Baseline/Mid/All reference profile acceptance;
- current Tactical unsupported-equipment rejection remains explicit.

## Gate 2 result

The 17.1 hull/system contract is frozen and ready for implementation.

No Gate-3 gameplay code is part of this Gate.

No Slice 17.2+ work is opened.
