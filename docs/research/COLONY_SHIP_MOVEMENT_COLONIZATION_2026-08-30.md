# Colony Ship production, strategic movement and colonization - 2026-08-30

**Open marker:** `docs/slices/_OPEN_COLONY_SHIP_MOVEMENT_COLONIZATION_2026-08-30.md`

## Goal

Establish from private original MOO2 1.31 executable/data evidence the minimum authoritative state and turn flow needed to build a Colony Ship, move it strategically, colonize a legal unowned planet, and create a second Colony in MOOX.

## Gate 1 - Checkup + original analysis / reverse engineering

**Status:** Gate 1 complete on 2026-08-30; Gate 2 implementation decision pending. Gameplay implementation has not started.

### Recovery state

- Starting branch: `main`.
- Starting HEAD: `5976f7e` (`docs: close treasury maintenance slice`).
- Starting tree: clean; `main` ahead of `origin/main` by 16 commits.
- Core `StateSchemaVersion`: 15.
- Economy ruleset schema: 7.
- Private executable SHA-256: `7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`.

### Existing MOOX boundary

MOOX already has:

- semantic Construction projects with authoritative PP progress/completion;
- Freighter Fleet production as a non-Building Construction project;
- minimal `StrategicFleet { ID, EmpireID, Role, AtSystemID }` state introduced solely for blockade production;
- original drive/ETA evidence and implementation for interstellar Population Settlers, but no generic Fleet transit state;
- deterministic Core save/load, GameSession command authorization, Observer projection and replay/event boundaries;
- authoritative Planet/Colony IDs and Planet `ColonyID` linkage, but no runtime Colony Ship colonization transition;
- race-aware Population cohorts suitable for representing a founding Population unit without restoring the original packed-entry representation.

### Evidence discipline

Direct executable/data observations below are labeled **original-observed**. Semantic interpretation supported by multiple direct paths is labeled **original-derived**. Proposed MOOX architecture is labeled **modernization**. Packed legacy encodings that are unnecessary to reproduce the behavior remain documentation only.

## Original special-ship class mapping

**Evidence level: original-observed.**

The prior blockade slice deliberately left the exact nonzero `fleet/design+0x11` mapping unresolved. The special-design loaders close that uncertainty directly:

```text
Load_Colony_Ship_Design_   VA 0x564CD -> design+0x11 = 1
Load_Transport_Ship_Design_            -> design+0x11 = 2
Load_Outpost_Ship_Design_              -> design+0x11 = 4
```

`Build_Ship_Type_Array_` independently branches on the same values for the three special ship families.

Therefore:

```text
+0x11 = 0  ordinary/combat ship
+0x11 = 1  Colony Ship
+0x11 = 2  Transport
+0x11 = 4  Outpost Ship
```

This slice needs only Colony Ship. Transport and Outpost remain separate later project kinds.

## Colony Ship technology requirement

**Evidence level: original-observed.**

`Colony_Can_Build_Product_` VA `0xE11BC` checks the current player's Technology-status byte for the `+0x11 == 1` special ship. Original technology status begins at player `+0x117`; the checked address is player `+0x140`:

```text
0x140 - 0x117 = 41
```

The byte must equal `3`, the existing directly established owned-Technology state.

Normalized MOOX data identifies Technology 41 as:

```text
technology_id = 41
id            = colony_ship
tech_field_id = 23
```

The same routine independently checks the matching technologies for Transport and Outpost Ship branches, corroborating the mapping.

**Gate 2 implication:** the Colony Ship Construction choice requires authoritative ownership of Technology 41 `colony_ship`.

## Colony Ship production cost

**Evidence level: original-observed for arithmetic; government-name mapping partly original-derived.**

`Ship_Type_Cost_For_Player_` VA `0xE0D98` assigns special base costs before the government modifier:

```text
Colony Ship  (+0x11 = 1): 500 PP
Transport    (+0x11 = 2): 100 PP
Outpost Ship (+0x11 = 4): 100 PP
```

`Load_Colony_Ship_Design_` calls this helper for special type 1 and stores the resulting cost in the design.

The helper then calls `Cost_Reduction_For_Govt_Type_` VA `0x6E1A0`. The directly observed arithmetic for the two reduced government enums is:

```text
enum 0: ceil(2 * cost / 3)
enum 1: ceil(cost / 3)
other:  cost unchanged
```

The enum-0 semantic is consistent with the original Feudal ship-production discount and enum-1 with its evolved Confederation form; this name mapping is kept **original-derived** here rather than pretending the raw enum byte alone names the government.

For the currently modeled ordinary starting governments the minimum useful Gate 2 rule is therefore:

```text
base Colony Ship cost = 500 PP
Feudal-derived ship cost = ceil(1000/3) = 334 PP
otherwise current modeled starting governments = 500 PP
```

Confederation/evolved-government runtime is not yet authoritative and need not be added solely for this slice.

## Colony Ship design installs the current drive

**Evidence level: original-observed.**

`Load_Colony_Ship_Design_` VA `0x564CD`:

- calls `Best_Warp_Drive_` VA `0x56726` for the owning player;
- stores the selected drive byte in the Colony Ship design at `+0x13`;
- derives drive-dependent design data;
- computes/stores the final ship cost.

A Colony Ship therefore has a concrete installed drive when built. The original does not merely ask for the player's newest drive afresh every turn after the ship exists.

Current MOOX already has directly established drive speeds used by Population-transfer ETA:

```text
Technology 120 Nuclear Drive       speed 2
Technology 72  Fusion Drive        speed 3
Technology 96  Ion Drive           speed 4
Technology 11  Anti-Matter Drive   speed 5
Technology 88  Hyper Drive         speed 6
Technology 95  Interphased Drive   speed 7
Trans-Dimensional race modifier    +2
```

**Gate 2 implication:** store the Colony Ship's effective strategic FTL speed on creation, instead of recomputing it from future Empire technology each turn. The existing drive helper should be generalized/reused rather than duplicated.

## Fuel range and movement legality

### Original player base range

**Evidence level: original-observed.**

`Update_Player_Ship_Range_` VA `0x10034D` selects the best owned Fuel Cell technology and writes the resulting base strategic range to player `+0x324`.

The six-record original Fuel table was read directly from the executable's LE data object. Its researched entries are:

| Technology | Normalized ID | Base range |
| --- | --- | ---: |
| Standard Fuel Cells | 167 | 4 parsecs |
| Deuterium Fuel Cells | 51 | 6 parsecs |
| Iridium Fuel Cells | 98 | 9 parsecs |
| Urridium Fuel Cells | 194 | 12 parsecs |
| Thorium Fuel Cells | 184 | 255 parsecs |

The leading zero record is the table sentinel/default record. `255` is effectively unlimited on the normal strategic map but remains an exact original byte value.

### Ship-specific range

**Evidence level: original-observed.**

`Ship_Range_` VA `0xFF496` reads the owner's base range from player `+0x324`. If the design contains the appropriate range-extending ship special, it applies an approximately 1.5x integer range multiplier. `Load_Colony_Ship_Design_` clears the normal optional-special slots for the special Colony Ship design, so the Colony Ship uses the player's Fuel Cell base range rather than an arbitrary optional Extended Fuel Tanks fit.

**Gate 2 implication:** a Colony Ship movement command must reject a destination farther from an in-range supply origin than the Colony Ship's Fuel Cell range. For the first headless vertical loop, the existing deterministic system-coordinate/parsec conversion should be reused. Do not silently make Colony Ships unlimited-range.

A later generic ship-design slice may model Extended Fuel Tanks and more complex per-design range; it is not needed for a fixed special Colony Ship.

## Construction completion creates a stationary strategic ship

**Evidence level: original-observed.**

`Apply_Production_` VA `0xE36DF` uses `Create_Ship_` VA `0xAF7B4` when a ship product completes.

`Create_Ship_` initializes the strategic record with:

```text
+0x63 = owner/player
+0x64 = 0              // stationary at star
+0x65 = producing Colony's star/system index
+0x67 = star X
+0x69 = star Y
```

It copies the built design into the strategic ship record and allocates/reuses the record deterministically.

`Apply_Production_` contains transient original post-production statuses while it finishes internal bookkeeping. These statuses are implementation artifacts, not additional persistent gameplay concepts MOOX needs to expose.

**Gate 2 implication:** successful Colony Ship Construction creates exactly one authoritative strategic Colony Ship, stationary at the producing Colony's System.

## Original strategic transit state

**Evidence level: original-observed.**

The strategic ship record has stride `0x81`. Movement routines directly establish:

```text
+0x64 = movement/location status
+0x65 = star index or encoded destination, depending on status
+0x67 = current strategic X
+0x69 = current strategic Y
+0x6B = movement-side transient/order byte
+0x6C = strategic movement step/speed input
+0x6D = remaining ETA in turns
```

`Ship_Destination_Star_` VA `0x1003F2` decodes destination as:

```text
status 1 -> +0x65 - 500
status 2 -> +0x65 - 1000
other    -> +0x65 directly
```

`Make_Ships_Move_To_` VA `0xFFD08` writes the packed destination, stores the movement ETA in `+0x6D`, updates the movement/order state and treats Colony/Outpost special ships explicitly.

`Update_Ship_ETAs_` VA `0x10041C` handles transit statuses 1/2 and repeatedly advances temporary coordinates using the movement-speed input until the destination coordinates are reached, writing the number of required turns to `+0x6D`.

`Move_All_Ships_Toward_Stars_` VA `0xFFEEA` advances status-1/2 ships, decrements ETA and calls `Make_Ship_Arrive_At_Star_` VA `0xFFDDA` when the destination is reached. Arrival normalizes the record back to stationary-at-star form.

### What Gate 1 does not claim

The exact gameplay semantic distinction between original packed status 1 and 2 is not required by this vertical slice and is not guessed. Likewise, interpolation X/Y exists for the original map animation/position model but need not become a second MOOX source of truth for headless movement.

### Proposed semantic replacement

**Evidence level: modernization preserving the observed behavior.**

For authoritative headless movement, replace the packed legacy encoding with explicit state:

```go
AtSystemID          ID
DestinationSystemID ID
RemainingTurns      int
FTLSpeed            int
```

Stationary and in-transit states should be mutually exclusive and validation-enforced. No encoded `500+star`/`1000+star` values are required.

## Turn-order timing

**Evidence level: original-observed.**

`Next_Turn_Calc_` VA `0x136B3` has the relevant sequence:

```text
All_AI_Colonize_
...
Move_All_Ships_Toward_Stars_
Resolve_Spies_
Apply_All_Player_Changes_
Apply_All_Colony_Changes_
...
Do_Colony_Calculations_
Compute_Blockades_
Move_Settlers_
Do_Colony_Calculations_
```

Existing permanent turn-order research already proves `Apply_All_Colony_Changes_` applies Population changes and then `Apply_Production_`.

Therefore strategic Fleet movement occurs **before current-turn Construction completion**.

Consequence:

- a Colony Ship that already existed in transit can move/arrive this turn;
- a Colony Ship completed by this turn's Production did not exist during that movement phase and cannot move until a later turn.

The original AI colonization pass also occurs before the movement pass, so an AI Colony Ship that arrives during movement is not immediately colonized by that earlier automatic colonization pass in the same sequence.

**Gate 2 implication:** MOOX should advance already-existing strategic Fleet transit before `advanceConstruction`. Do not create and move a newly completed Colony Ship in one resolution tick.

## Planet colonizability

**Evidence level: original-observed.**

`Planet_Is_Colonizable_` VA `0x714A1` accepts a target only when:

1. the Planet index is valid;
2. original `planet+0x04 == 3`, the real colonizable-planet body state rather than a non-colonizable system body;
3. and either:
   - the Planet has no Colony record (`planet+0x00 == -1`), or
   - its existing Colony record has `colony+0x06 != 0`.

The Colony/Outpost creation path independently proves `colony+0x06 = 1` is the Outpost flag. Thus original Colony Ships can colonize an empty normal Planet **or convert an existing Outpost**.

There is no climate/terraforming-technology gate in this direct predicate.

Current MOOX has semantic normal `Planet` records but no Outpost state. The dependency-safe first slice should therefore allow only an existing normal Planet with `Planet.ColonyID == 0`; Outpost conversion is explicitly deferred rather than simulated with a fake Colony.

## Colonize action requires an owned Colony Ship at the System

**Evidence level: original-observed.**

`Check_For_Colonization_Button_` / `Set_Colonize_Button_Flag_` VA `0x71F35` directly requires:

- the relevant Fleet/stack belongs to the current player;
- at least one strategic record in the selected stack has `+0x11 == 1` (Colony Ship);
- at least one Planet in the System passes `Planet_Is_Colonizable_`.

This directly supports the MOOX legality boundary:

```text
owned Colony Ship + stationary at target System + legal unowned Planet in that same System
```

A separate original system-state byte participates in UI flag calculation; it is not needed to restate the independently proven Fleet-owner/Planet legality rules and is not promoted into MOOX state here.

## Colony creation

### Colony versus Outpost

**Evidence level: original-observed.**

`Make_New_Colony_Or_Outpost_` VA `0xE5EB3` is the common path. Its wrappers establish:

```text
Make_New_Colony_  -> type 0
Make_New_Outpost_ -> type 1
```

The Outpost branch explicitly sets `colony+0x06 = 1`; the Colony branch initializes a normal Colony and does not set the Outpost flag.

`Init_Colony_` VA `0x12D75` clears/initializes the Colony record, links Planet and Colony, sets ownership and updates strategic presence/bookkeeping.

### Initial Population

**Evidence level: original-observed plus prior permanent Population decoding.**

The normal new-Colony path sets:

```text
colony+0x0A = 1
```

`POPULATION_COHORTS_2026-08-29.md` directly establishes:

```text
colony+0x0A = number of Population entries
colony+0x0C = first packed Population entry
```

The founding entry's source/origin and loyalty bits are initialized to the founding player and it is not marked conquered. Thus the normal new Colony begins with **one assimilated founding Population unit**.

For an ordinary organic food-consuming founder, the normal job is Farmer. The original path contains special race/planet branches that can initialize different job/special-pop cases and a separate Native-style multi-pop case. MOOX does not yet model Native population and should not invent that state in this slice.

For the first dependency-safe vertical slice, create one assimilated founding cohort with `OriginEmpireID == LoyaltyEmpireID == founding Empire`. Ordinary races use one Farmer. Existing authoritative race semantics can later extend special founding-job behavior where directly required without changing the identity model.

### Starting Buildings

No direct normal new-Colony path grants a persistent starting Building. The original initializes Colony bookkeeping and recalculates the Colony. Gate 2 should not invent free Buildings.

## Colony Ship consumption

**Evidence level: original-observed.**

`AI_Colonize_` VA `0xE65F8`:

1. finds a Colony Ship at the System;
2. chooses a valid unoccupied Planet;
3. calls the normal Colony creation path;
4. after successful Colony creation, calls `Kill_Ship_` VA `0xA163A` for the Colony Ship record.

Therefore colonization **consumes the Colony Ship**.

The semantic MOOX transition should atomically create/link the new Colony and remove the consumed Colony Ship Fleet from authoritative state.

## Proposed Gate 2 contract

This section is a proposal only. Gate 3 must not start until it is accepted or revised.

### Schema

Advance Core `StateSchemaVersion` from **15 to 16** because strategic Fleet transit and Colony Ship identity become persisted authoritative state.

Keep Economy ruleset schema **7** unless implementation needs a genuinely new normalized-data field. Technology IDs, system coordinates and existing race/government rules are already available.

### Colony Ship Construction project

Add a semantic Construction project kind:

```go
ConstructionProjectColonyShip = "colony_ship"
```

Expose a server-authoritative legal action/command consistent with existing Construction patterns, e.g. `colony.queue_colony_ship`.

Legality:

- Colony belongs to commanding Empire;
- Technology 41 `colony_ship` is known;
- one active Construction project at a time as today.

Cost:

- base 500 PP;
- apply the directly observed ship-government reduction for currently modeled relevant governments;
- no tactical ship-design choice is exposed for this fixed special ship.

Completion:

- create one strategic Fleet at the producing Colony's System;
- role = civilian;
- special kind = Colony Ship;
- persist the effective FTL speed installed at completion;
- clear/advance Construction through the existing deterministic completion model.

### Strategic Fleet identity and transit

Preserve the existing `Role` field because blockade logic already uses `combat` versus `civilian`. Add a semantic special kind rather than overloading Role:

```go
type StrategicFleetSpecialKind string

const (
    StrategicFleetSpecialNone       StrategicFleetSpecialKind = ""
    StrategicFleetSpecialColonyShip StrategicFleetSpecialKind = "colony_ship"
)
```

Extend `StrategicFleet` approximately as:

```go
type StrategicFleet struct {
    ID                  ID
    EmpireID            ID
    Role                StrategicFleetRole
    SpecialKind         StrategicFleetSpecialKind `json:"special_kind,omitempty"`
    AtSystemID          ID                        `json:"at_system_id,omitempty"`
    DestinationSystemID ID                        `json:"destination_system_id,omitempty"`
    RemainingTurns      int                       `json:"remaining_turns,omitempty"`
    FTLSpeed            int                       `json:"ftl_speed,omitempty"`
}
```

Validation should make stationary and transit states mutually exclusive. Existing combat Fleets can remain stationary-only with zero `FTLSpeed` until their own generic movement/composition slice; do not force fake drives onto blockade fixtures.

### Fuel range helper

Add/reuse an authoritative helper that maps known Fuel Cell Technology to the original Colony Ship base range:

```text
167 -> 4 pc
 51 -> 6 pc
 98 -> 9 pc
194 -> 12 pc
184 -> 255 pc
```

Use the existing StarSystem coordinate-to-parsec distance convention already shared by Population transfer. A movement order beyond range is rejected.

Do not implement Extended Fuel Tanks in this fixed special-ship slice.

### Movement command

Add an Empire-authorized strategic command such as:

```text
empire.move_fleet
```

Payload:

```text
fleet_id
destination_system_id
```

For this slice, authorize movement only for the new Colony Ship special kind. Broad combat-Fleet movement remains deferred.

Launch legality:

- Fleet belongs to commanding Empire;
- Fleet is a Colony Ship;
- Fleet is stationary at a System;
- destination is a different known System;
- destination is within Fuel range.

Transition:

```text
AtSystemID = 0
DestinationSystemID = target
RemainingTurns = deterministic ETA from system distance / persisted FTLSpeed
```

Advance transit once per strategic resolution **before Construction**. On arrival:

```text
AtSystemID = DestinationSystemID
DestinationSystemID = 0
RemainingTurns = 0
```

Emit deterministic started/progressed/arrived events. Do not persist original map-animation X/Y interpolation or packed 500/1000 destination encodings.

### Colonization command

Add an Empire-authorized command such as:

```text
empire.colonize_planet
```

Payload:

```text
fleet_id
planet_id
```

Legality:

- Fleet belongs to commanding Empire;
- Fleet is a Colony Ship;
- Fleet is stationary;
- target Planet exists in the same System;
- target is a normal currently representable Planet;
- `Planet.ColonyID == 0`.

Do not queue a colonize order for a Fleet that is still moving. This matches the original turn shape: colonization is decided from a ship already at a System, while movement occurs as a separate later turn phase. A Colony Ship arriving during movement becomes eligible for a colonize command on the next command phase.

Effect, atomically:

1. allocate a new Colony ID;
2. set `Planet.ColonyID`;
3. create Colony owned by the founding Empire;
4. create one assimilated founding Population cohort for that Empire (ordinary default: one Farmer);
5. recalculate authoritative Colony economy/population snapshots;
6. consume/remove the Colony Ship Fleet;
7. emit colonization/ship-consumption events visible in Observer/replay.

### Resolver timing

Insert strategic Fleet transit advancement before `advanceConstruction`, preserving the original fact that movement precedes current-turn production.

Do not automatically colonize on arrival in that same movement step.

Blockade recomputation continues to see only stationary combat Fleets; the new moving/civilian Colony Ship cannot accidentally produce a blockade.

### Session / Observer / replay

- Construction choices expose Colony Ship only when Tech 41 is known.
- Seat authorization applies to move/colonize commands.
- Observer full-state clone must preserve/isolate the new persisted Fleet transit/special fields and the resulting Colony.
- Save/load exact-roundtrip must include the new fields.
- Replay events must make build, movement, arrival and colonization deterministic and inspectable.

## Gate 3 regression plan if Gate 2 is accepted

At minimum:

### Construction

- no Tech 41 -> no Colony Ship legal choice / command rejected;
- Tech 41 -> legal choice;
- standard government cost 500 PP;
- Feudal current-model cost follows observed `ceil(2/3)` rule;
- completion creates one civilian Colony Ship at producing System;
- newly completed ship does not also advance movement in the same resolution.

### Fuel / movement

- Standard Fuel Cells = 4 pc, Deuterium = 6, Iridium = 9, Urridium = 12, Thorium = 255;
- out-of-range destination rejected;
- in-range order transitions to transit with exact deterministic ETA;
- transit decrements once per turn before Construction;
- arrival restores stationary target-System state;
- moving Colony Ship never contributes to blockade producer;
- save/load and Observer clone preserve transit exactly.

### Colonization

- wrong owner rejected;
- ordinary combat Fleet rejected;
- moving Colony Ship rejected;
- Planet in different System rejected;
- already-colonized Planet rejected;
- valid stationary Colony Ship + empty same-System Planet creates second Colony;
- Planet links to new Colony;
- founding Colony has one assimilated founding Population unit;
- no free Building is invented;
- consumed Colony Ship disappears;
- recalculated Economy/Population state is valid;
- Observer/replay reflects the transition.

### Vertical loop

A headless regression should demonstrate:

```text
home Colony
-> research/own Technology 41
-> queue/build Colony Ship
-> ship appears at home System
-> later command moves it to another in-range System
-> movement advances over turns
-> ship arrives
-> later command colonizes empty Planet
-> second Colony exists
-> save/load exact round trip
```

## Deliberate Gate 1 deferrals

- tactical ship design, weapons, shields, armor and combat composition;
- generic combat-Fleet movement/orders;
- original packed movement status 1-vs-2 semantics beyond what is required for equivalent headless transit;
- map-animation/interpolated ship X/Y as authoritative state;
- Extended Fuel Tanks and arbitrary per-design range customization;
- Outpost Ship production and Outpost->Colony conversion;
- Transport/troop movement and invasion;
- Native/special founding-population cases;
- hostile-system combat/engagement before colonization;
- AI colony targeting/auto-colonize policy;
- broad diplomacy and treaty permissions;
- special/non-player ship owners.

These items remain later evidence-driven slices rather than prerequisites invented into the first second-Colony vertical loop.
## Gate 2 - Accepted implementation decision

**Status:** accepted without revision on 2026-08-30.

The proposed Core schema 16, fixed Colony Ship Construction identity/cost, semantic Fleet transit, original Fuel Cell range, pre-Construction movement timing, explicit colonize command, second-Colony creation/ship consumption and Observer/save/replay boundaries were accepted as written. Deferred Outpost, Transport, tactical/combat Fleet and AI/diplomacy systems remain outside this slice.

## Gate 3 - Implementation

**Status:** complete on 2026-08-30; Gate 4 pending.

### Core schema 16

`StrategicFleet` preserves the existing combat/civilian `Role` and adds semantic special/transit state:

```text
SpecialKind
AtSystemID
DestinationSystemID
RemainingTurns
FTLSpeed
```

`colony_ship` is the only special kind and the only Fleet type allowed to use semantic transit in this schema. Validation enforces a civilian Colony Ship, installed `FTLSpeed >= 2`, known current/destination System references, positive ETA while moving and mutually exclusive stationary/transit state. Existing ordinary blockade Fleets remain compatible and do not receive fabricated drive/order data.

### Construction

Technology 41 exposes `ConstructionProjectColonyShip` / project `colony_ship`. Base cost is 500 PP; the currently modeled Feudal government uses the accepted observed `ceil(2*500/3)=334` ship cost. Completion allocates one civilian Colony Ship at the producing Colony's System and persists the effective FTL speed installed at construction time.

### Range and movement

The implementation uses the directly observed Fuel Cell table:

```text
Tech 167 Standard   4 pc
Tech 51  Deuterium  6 pc
Tech 98  Iridium    9 pc
Tech 194 Urridium  12 pc
Tech 184 Thorium  255 pc
```

Movement legality checks the destination against the nearest currently authoritative Empire supply System. In schema 16 those supply origins are Empire-owned Colonies; Outpost supply is deliberately deferred until Outposts exist as authoritative state. Travel ETA uses source-to-destination parsec distance and the Colony Ship's persisted FTL speed.

`empire.move_fleet` currently accepts only owned, stationary Colony Ships. Transit advances once per strategic resolution before Construction and emits started/progressed/arrived events. Arrival clears the destination/ETA and restores `AtSystemID`.

The common strategic parsec conversion is also reused by Population-transfer ETA; its separately proven 15-turn cap is retained only for Population transfers and is not imposed on generic Colony Ship travel.

### Colonization

`empire.colonize_planet` requires:

- owned Colony Ship;
- stationary at a System;
- target Planet in the same System;
- `Planet.ColonyID == 0` and no hidden duplicate Colony reference.

Success atomically allocates the new Colony identity, links the Planet, creates one assimilated founding Farmer cohort with founder origin/loyalty, recalculates the new Colony, removes the consumed Colony Ship and emits colonized/consumed events. No free Building is created. Outpost conversion remains deferred.

### Deterministic regressions

`internal/game/colony_ship_test.go` covers Technology/cost, queue state, installed drive, movement-before-Production timing, all five Fuel ranges, in/out-of-range movement, exact ETA progression/arrival, invalid colonization cases, founding Population/no-free-Building/ship consumption and the complete build -> move -> arrive -> second-Colony loop including save/load.

`internal/core/strategic_test.go` covers schema-16 Colony Ship transit roundtrip and invalid state combinations.

`internal/session/colony_ship_test.go` drives the real GameSession across build, movement, arrival and colonization, checks seat legal-action authority, verifies Observer transit/new-Colony state and event history, and proves mutations to the Observer clone do not alter authoritative Fleet state.

Fresh Gate 3 regression:

```text
go test ./... -count=1
```

passed across all packages.

Gate 4 remains pending for final gofmt, full test/vet/diff QA, implementation commit and slice-closing documentation/marker removal.