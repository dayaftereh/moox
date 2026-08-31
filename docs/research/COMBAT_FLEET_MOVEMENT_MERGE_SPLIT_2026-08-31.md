# Combat Fleet movement, merge and split - Slice 04 evidence

Date: 2026-08-31

Status: **closed; Gates 1-4 complete**

Recovery marker removed at Gate 4 closure: `docs/slices/_OPEN_COMBAT_FLEET_MOVEMENT_MERGE_SPLIT_2026-08-31.md`

Planned slice: `docs/slices/PLANNED_04_COMBAT_FLEET_MOVEMENT_MERGE_SPLIT.md`

## Objective

Establish the original-evidence-backed strategic rules and the minimum deterministic Core/Game boundary for concrete military Fleet movement, arrival, merge and split without pulling tactical combat into this slice.

## Fresh repository check

- Starting commit: `7fd43f4` (`docs: close military ship design slice`).
- `main` was 26 commits ahead of `origin/main` at slice start; no push is part of Gate 1.
- Working tree was clean and no `_OPEN_*.md` marker existed before Slice 04 activation.
- No foreign active write session was present.
- Slice 03 is closed; Core schema 18 is the current baseline.
- Fresh baseline `go test ./... -count=1` passed after the Slice 04 documentation marker/evidence files were created; no gameplay code was changed in Gate 1.

## Evidence discipline

This document separates:

- **direct original evidence** - original 1.31 help/data/executable semantics;
- **original-derived inference** - semantic conclusions forced by several direct facts but not yet represented by one isolated original branch;
- **current MOOX state** - behavior already implemented in Core/Game;
- **proposed MOOX normalization** - the Gate-2 decision surface, not yet accepted or implemented.

The private original installation remains research-only and is not distributable MOOX content.

## Original-reference provenance and symbol correction

The private 1.31 DOS executable at `C:\ASH\Temp\mastori2\Orion2.exe` contains Watcom debug-symbol strings, including source-module names such as `D:\MOX\OTHERS\SHIPMOVE.C` and `D:\MOX\OTHERS\FLEETPOP.C` and strategic symbols such as `Ships_Try_To_Move_To_`, `Make_Ships_Move_To_`, `Make_Ship_Arrive_At_Star_`, `Move_All_Ships_Toward_Stars_`, `Update_Ship_ETAs_`, `Fltscrn_Move_Ships_`, `Get_Fleet_Box_Selected_Ship_Ids_`, `Calc_Player_FTL_Speed_` and `Get_FTL_Speed_`.

Gate 1 found an address-provenance issue in some older research notes: the Watcom symbol record stores the object offset **after the symbol name**. Earlier notes sometimes associated the address immediately before a name with that name, which is actually the previous symbol's address. For Slice 04, exact function-start claims use the embedded record trailer, not the preceding record.

Verified examples from the embedded symbol records, object 1 base `0x10000`:

| Symbol | object-1 offset | corrected VA |
| --- | ---: | ---: |
| `Best_Warp_Drive_` | `0x4679E` | `0x5679E` |
| `Calc_Player_FTL_Speed_` | `0x475D6` | `0x575D6` |
| `Get_FTL_Speed_` | `0x47617` | `0x57617` |
| `Ships_Try_To_Move_To_` | `0xEFBD6` | `0xFFBD6` |
| `Make_Ships_Move_To_` | `0xEFDC9` | `0xFFDC9` |
| `Ship_Is_Moving_` | `0xEFDDA` | `0xFFDDA` |
| `Make_Ship_Arrive_At_Star_` | `0xEFEEA` | `0xFFEEA` |
| `Move_All_Ships_Toward_Stars_` | `0xF0010` | `0x100010` |
| `Update_Player_Ship_Range_` | `0xF038C` | `0x10038C` |
| `Update_All_Ship_Ranges_` | `0xF03F2` | `0x1003F2` |
| `Ship_Destination_Star_` | `0xF041C` | `0x10041C` |
| `Update_Ship_ETAs_` | `0xF0519` | `0x100519` |

The behavioral conclusions carried from earlier slices remain useful, but future exact-address citations should use this corrected symbol interpretation or a freshly verified disassembly.

## Finding 1 - original movement operates on a selected subset of ships

Evidence level: **direct original help + direct executable symbol structure**.

Original `HELP.LBX` record 213, `Move Ships`, explicitly describes this flow:

1. select the ship/fleet icon at a System;
2. the UI displays the ships currently in orbit there;
3. individual ship pictures can be selected/deselected;
4. `ALL` selects/deselects all ships in that fleet/stack;
5. the selected ships are then sent to a destination System;
6. the destination line is green when in range and the popup shows ETA; red means out of range.

The executable independently exposes `FLEETPOP.C` symbols including `Get_Fleet_Box_Selected_Ship_Ids_`, `Fleet_Movement_Icon_Ship_Stack_`, `Fltscrn_Move_Ships_` and selection-status helpers.

Therefore an original strategic "split" is not primarily a persistent fleet-container edit. The original owns individual strategic Ship records and allows an arbitrary selected subset of a co-located stack to receive the movement order. The remainder stays behind.

This is the strongest original requirement Slice 04 must preserve semantically.

## Finding 2 - ordinary movement orders start from a stationary System; hyperspace retargeting is technology-gated

Evidence level: **direct original help**.

The normal `Move Ships` help text starts from ships "currently in orbit" around a System.

Original `HELP.LBX` record 91, `Hyperspace Communications`, explicitly states that the technology allows communication with a ship **already in hyperspace** so its destination can be changed.

Consequences:

- baseline movement does not need arbitrary in-transit rerouting;
- changing the destination of an already-moving Fleet is a separate technology-gated rule;
- Slice 04 can safely reject in-transit split/merge/reroute and defer those semantics until Hyperspace Communications is modeled.

This also removes the need in this slice to persist continuous route coordinates solely to support mid-flight re-routing.

## Finding 3 - current MOOX transit state is already the right semantic shape

Evidence level: **current MOOX state + earlier direct original movement evidence**.

Core schema 18 already persists on `StrategicFleet`:

- `AtSystemID`;
- `DestinationSystemID`;
- `RemainingTurns`;
- `FTLSpeed`;
- `ShipIDs`.

Current validation deliberately allows transit only for fixed Colony/Outpost Ship fleets and rejects ordinary combat-Fleet transit with `ordinary fleet cannot carry semantic transit state yet`.

`internal/game/strategic_movement.go` already provides deterministic:

- star-to-star parsec distance (`ceil(hypot(dx,dy)/30)`);
- ETA ceiling division;
- Colony/Outpost supply-origin lookup;
- movement-start/progress/arrival events;
- mutually exclusive stationary vs transit state.

The original strategic Ship record carries movement status/destination/ETA and transitions back to stationary-at-star state on arrival. Slice 04 therefore should generalize the existing semantic transit representation rather than introduce a second combat-only route model.

## Finding 4 - normal strategic drive speed is empire/current-technology behavior, not a justified persisted Fleet aggregate

Evidence level: **direct original help + direct prior original-code evidence; planned "slowest ship" expectation is not confirmed for the supported surface**.

Original drive help records state for Nuclear through Interphased Drive that the technology moves a ship at 2..7 parsecs/turn and that **all ship drives are automatically upgraded when the technology is discovered**.

Prior direct 1.31 analysis of `Calc_Player_FTL_Speed_` established the player strategic speed from the best Warp Drive and the `Trans-Dimensional` +2 parsecs/turn modifier. The corrected embedded symbol starts are `Calc_Player_FTL_Speed_` VA `0x575D6` and `Get_FTL_Speed_` VA `0x57617`.

Current MOOX already applies the same best-drive + Trans-Dimensional rule to Population transfer and to fixed Colony/Outpost Ship completion. By contrast, Slice 03 stores a military Ship's design/build snapshot `Spec.FTLSpeed` as provenance of its installed design at construction time.

For the **currently supported ordinary military Ship surface** there is no original basis to persist a separate Fleet speed or to take the minimum of stale build-snapshot speeds. Normal ship drives auto-upgrade with Empire technology, and all ships of one Empire receive the same race movement modifier.

Proposed semantic rule for Gate 2:

```text
combat_fleet_effective_ftl_speed = current empire strategic FTL speed
                                    = best known Warp Drive speed
                                    + Trans-Dimensional modifier
```

The movement-start event may record that derived value for replay/debug visibility, but ordinary combat `StrategicFleet.FTLSpeed` should not become an authoritative cached aggregate.

Future captured-ship, damaged-drive, special movement or other genuinely per-Ship strategic modifiers can introduce a minimum-per-Ship aggregation if original evidence requires it. Slice 04 should not invent that rule now.

### Deferred original modifier - wormholes

Original help states that a Wormhole increases speed through that connection by +3 parsecs/turn. MOOX currently has no authoritative Wormhole route state. That route-specific modifier is therefore evidence-backed but dependency-blocked and remains deferred rather than silently approximated.

## Finding 5 - normal Fleet range is current Empire fuel range; per-Ship range differences enter through ship specials

Evidence level: **direct original help + carried direct original-code evidence + original-derived group legality**.

Original help states that Fuel Cell technology affects strategic distance and that `Extended Fuel Tanks` increase a ship's overall range by +50%.

Earlier original-code analysis established the normal Fuel Cell ranges:

| Fuel Cell | parsecs |
| --- | ---: |
| Standard | 4 |
| Deuterium | 6 |
| Iridium | 9 |
| Urridium | 12 |
| Thorium | 255 |

The original `Ship_Range_` path uses the owning player's current base range and applies the per-Ship Extended Fuel Tanks multiplier when present. That means the design/build-time `Ship.Spec.FuelRangeParsecs` is useful provenance but is not sufficient as the permanent Empire movement limit after new Fuel Cell technology is discovered.

For a selected group, every selected Ship must be legal at the destination; therefore a future group containing range-modified Ships has the effective range of its most restrictive member. That minimum is an **original-derived inference** from per-Ship range plus selected-subset movement.

However, Slice 03 deliberately supports no ship specials. Consequently every currently supported military Ship uses the same current Empire base Fuel Cell range and no mixed-range special case exists yet.

Gate-2 proposal for this slice:

- ordinary combat Fleet base range = current Empire best Fuel Cell range;
- supply legality uses the existing nearest owned Colony/Outpost supply-origin helper;
- no Extended Fuel Tanks implementation is pulled into Slice 04;
- structure the range helper so a later per-Ship modifier can reduce/increase member range without changing the movement command contract.

## Finding 6 - destination range is measured from Empire supply, and Outposts are supply origins

Evidence level: **direct original-code evidence carried from Slice 02 + current MOOX implementation**.

The original UI help simplifies the out-of-range message as distance from the nearest Colony. Slice 02's direct Outpost analysis established that owned Outposts extend ship operational range in the original.

Current `nearestEmpireSupplyDistanceParsecs` already resolves the nearest owned Colony or owned Outpost to the requested destination System. Slice 04 should reuse this helper for combat Fleets; a separate military range map would be wrong and would duplicate proven semantics.

## Finding 7 - ETA and transit progression can reuse the existing deterministic model

Evidence level: **direct original movement structure + current MOOX semantic normalization**.

The original movement code stores a destination and ETA on strategic Ship records, advances moving ships during next-turn processing and normalizes them back to stationary System state on arrival.

Current MOOX already represents that as:

```text
stationary: AtSystemID != 0, DestinationSystemID = 0, RemainingTurns = 0
transit:    AtSystemID = 0, DestinationSystemID != 0, RemainingTurns > 0
arrival:    AtSystemID = destination, DestinationSystemID = 0, RemainingTurns = 0
```

For a normal route without Wormhole state, ETA remains:

```text
parsecs = ceil(hypot(dx, dy) / 30)
eta = ceil(parsecs / effective_ftl_speed)
```

Because Slice 04 defers in-transit rerouting/splitting, destination + remaining turns is sufficient persistent route state. Source System and the derived speed/range belong in movement-start Observer/replay events rather than being duplicated as authoritative mutable Fleet fields.

## Finding 8 - split is explicit in MOOX; merge is a semantic normalization of original stack co-location

Evidence level: **direct original selected-subset behavior + original-derived stack semantics + proposed MOOX normalization**.

The original model exposes individual Ship selection and stack/fleet UI grouping rather than MOOX-style stable `StrategicFleet` IDs. Therefore MOOX needs an explicit state transition when a selected subset leaves a persistent Fleet container.

Proposed stationary split semantics:

- only an owned stationary combat Fleet at a known System;
- selected Ship IDs must be sorted/unique, belong to that Fleet and form a non-empty proper subset;
- source Fleet keeps the unselected Ships;
- a new combat Fleet receives the selected Ships at the same System;
- the new Fleet ID is allocated through `GameState.NewID()`;
- both resulting `ShipIDs` lists remain sorted and non-empty.

For the original one-action UX, `move_fleet` should be able to carry an optional selected `ship_ids` subset and atomically create/move the new Fleet. This avoids requiring a client to predict the newly allocated Fleet ID before issuing the move.

A separate explicit stationary `split_fleet` command is still useful for organization without immediate movement.

### Merge

No direct original evidence was found for a persistent "merge Fleet object" command, which is expected because the original stack is derived from individual Ship co-location. MOOX nevertheless needs a semantic operation because its stable Fleet containers are authoritative.

The narrow evidence-safe merge is:

- same owner;
- combat role;
- both stationary;
- same `AtSystemID`;
- no fixed civilian special kind;
- combine sorted `ShipIDs` into one retained Fleet ID;
- delete the consumed Fleet.

In-transit merge is not required by the direct original selected-at-System flow and should be rejected in this slice.

## Finding 9 - movement/arrival precedes blockade recomputation

Evidence level: **direct original turn-order evidence + current MOOX implementation**.

Earlier direct `Next_Turn_Calc_` analysis places strategic ship movement before the later `Compute_Blockades_` pass.

Current `EconomyResolver.Resolve` already does:

```text
... Population
advanceStrategicFleetTransit
advanceConstruction
recomputeSystemBlockades
advancePopulationTransfers
...
```

`recomputeSystemBlockades` considers only stationary combat Fleets at the System (`fleet.Role == combat && fleet.AtSystemID == system.ID`) and hostile relations.

Consequences for Slice 04:

- a departing combat Fleet immediately ceases to blockade its source because `AtSystemID` becomes zero;
- a combat Fleet arriving during transit advancement can participate in the same turn's subsequent blockade recomputation;
- newly completed military Ships are also present before that same blockade recomputation;
- no additional blockade-specific movement state is needed.

## Finding 10 - hostile arrival can remain a temporary strategic co-location boundary until the encounter slice

Evidence level: **current proven blockade semantics + roadmap boundary; deliberate temporary MOOX scope choice proposed for Gate 2**.

The current blockade model deliberately permits a hostile stationary combat Fleet at a System containing another Empire's Colony and derives blockade from that co-location. Slice 04's exit criterion specifically requires movement/arrival to drive blockade from authoritative Fleet location.

Therefore Gate 2 should allow a combat Fleet to arrive at a hostile System and become stationary for now. Slice 04 must **not** invent tactical battle results or consume either Fleet.

Slice 06 (`Strategic hostile encounters -> BattleSession handoff`) will later intercept/resolve hostile co-location at the correct boundary. Until then, stationary hostile co-location is an explicit temporary strategic placeholder, not a claim that the original skipped combat.

## Current MOOX gap

Core schema 18 has all state primitives needed for the narrow slice but intentionally blocks ordinary combat transit:

- concrete military Ships exist and are assigned exactly once through `StrategicFleet.ShipIDs`;
- combat Fleets are currently stationary-only;
- `moveFleet` explicitly accepts only fixed Colony/Outpost Ship fleets;
- `advanceStrategicFleetTransit` explicitly rejects ordinary combat transit;
- no split/merge commands/events exist;
- combat Fleet movement speed/range derivation helpers do not exist yet;
- Observer/save/load already serialize the underlying Fleet/Ship state and only need new event/command coverage.

## Gate 2 accepted implementation contract

The following contract is **accepted for Gate 3 implementation**. Gate 2 changes documentation/decision state only; gameplay code remains untouched until Gate 3.

1. **Schema 19.** Bump Core schema because ordinary combat Fleets gain legal semantic transit and split/merge state transitions even though no new Fleet field is strictly required.
2. **Reuse `StrategicFleet` transit fields.** Stationary/transit remain mutually exclusive. Fixed special fleets continue using persisted `FTLSpeed`; ordinary combat Fleets derive speed when movement is ordered and do not cache it as authoritative Fleet state.
3. **Current-Empire movement profile.** Add one authoritative helper for combat Fleet effective FTL speed = best current Warp Drive + Trans-Dimensional bonus, and base range = best current Fuel Cell. Do not use stale construction snapshots as the Empire's current strategic movement capability.
4. **Range extension-ready boundary.** Validate every concrete Ship membership now; keep a per-member effective-range hook for future Extended Fuel Tanks/captured/special rules, but implement no unsupported ship special in this slice.
5. **Reuse supply lookup.** Destination legality is nearest owned Colony/Outpost supply distance <= effective Fleet range.
6. **Generalize `empire.move_fleet`.** Keep `fleet_id` + `destination_system_id`; add optional sorted/unique `ship_ids`. Omitted/all means move the whole Fleet. A proper subset performs a deterministic split and moves the new Fleet atomically. Fixed Colony/Outpost Ship movement rejects `ship_ids` selection.
7. **Add explicit stationary split.** `empire.split_fleet {fleet_id, ship_ids}` allocates a new Fleet with `GameState.NewID()` and emits `empire.fleet_split`.
8. **Add explicit stationary merge.** `empire.merge_fleets {target_fleet_id, source_fleet_id}` requires same owner/system, stationary combat Fleets; target survives, source is removed; emit `empire.fleets_merged`.
9. **Movement events.** Keep `empire.fleet_movement_started/progressed/arrived`; movement-start records source/destination, effective FTL speed, effective range and supply distance. A subset move emits split before movement-start deterministically.
10. **No in-transit mutation.** Reject split, merge and reroute while moving. Hyperspace Communications owns later reroute semantics.
11. **No Wormhole bonus yet.** Route-specific +3 is deferred until authoritative Wormhole connection state exists.
12. **Hostile arrival placeholder.** Arrival may produce hostile stationary co-location and therefore blockade; no battle handoff in this slice.
13. **Deterministic ordering/invariants.** `ShipIDs` sorted/unique/non-empty for combat Fleets; every concrete Ship assigned exactly once; Fleet slice remains ascending by ID after allocation/removal; no empty combat Fleet survives split/merge.
14. **Observer/replay/save-load.** New commands/events must reproduce exact Fleet IDs, membership, transit and arrival state across replay/save-load.
15. **Vertical regression.** Build at least two military Ships, merge or construct a multi-Ship Fleet, move a selected subset to force split, advance transit/arrival, verify source/moving Fleet membership, effective current-tech movement profile, supply legality and arrival-driven blockade.
16. **Compatibility.** Colony Ship/Outpost Ship transit, Population transfer, construction timing and existing blockade behavior remain unchanged.

## Gate 1 conclusion

Gate 1 is complete.

The key correction to the prepared expectation is that the supported original military movement surface does **not** justify a persisted or "slowest build-snapshot Ship" Fleet speed. Original drives auto-upgrade Empire-wide, strategic speed has an Empire/race component, and current MOOX military Ships do not yet carry per-Ship strategic movement specials. The evidence-safe first implementation derives current Empire movement capability at order time, while preserving a future per-Ship modifier boundary.

The original selected-subset movement behavior maps cleanly to MOOX as a deterministic split-plus-move operation on explicit Fleet containers. Stationary explicit split/merge complete the minimum management surface. In-transit rerouting and route-specific Wormhole effects remain correctly deferred.

**Gate 2 is accepted. Stop here before Gate 3 implementation.**

## Gate 2 final decisions

Gate 2 accepts the Gate-1 direction with the following exact implementation decisions.

### 1. Core schema and persisted Fleet state

Core `StateSchemaVersion` will become **19** in Gate 3.

No new `StrategicFleet` field is required for the first generic combat movement implementation. The existing fields remain authoritative:

```text
stationary Fleet:
  AtSystemID != 0
  DestinationSystemID == 0
  RemainingTurns == 0

moving Fleet:
  AtSystemID == 0
  DestinationSystemID != 0
  RemainingTurns > 0
```

For fixed Colony/Outpost special Fleets, persisted `FTLSpeed` remains their existing installed/derived special-Fleet value.

For ordinary combat Fleets, persisted `FTLSpeed` is **not authoritative movement state** and must remain zero. Combat movement speed is derived at order time from current Empire capability. Validation in schema 19 therefore permits ordinary combat transit while requiring `FTLSpeed == 0` for ordinary combat Fleets.

Rationale for the schema bump despite no new JSON field: schema 18 explicitly defines ordinary combat transit as invalid. Schema 19 changes the legal semantic state space and save/load validation contract, so retaining version 18 would make old and new validators disagree about the same serialized state.

### 2. Combat Fleet movement profile

Gate 3 will add one authoritative movement-profile helper for an owned ordinary combat Fleet:

```text
effectiveFTLSpeed = current Empire best known Warp Drive strategic speed
                  + Trans-Dimensional strategic bonus

effectiveRange = minimum effective range of every member Ship
```

For the currently supported military Ship surface, every member's effective range is the Empire's current best Fuel Cell range, so the minimum collapses to that value. The helper must nevertheless enumerate and validate concrete `ShipIDs`, establishing the future hook for Extended Fuel Tanks, captured Ships or other per-Ship modifiers without changing the command contract later.

Build-time `Ship.Spec.FTLSpeed` and `Ship.Spec.FuelRangeParsecs` remain immutable design provenance and are not used as current Empire strategic capability after technology advances.

A combat Fleet with no valid member Ships, a foreign Ship, duplicated Ship membership or unsupported member state is rejected by Core validation/command legality rather than approximated.

### 3. Shared supply lookup and route calculation

All ordinary combat movement reuses the existing semantic helpers:

```text
routeParsecs = ceil(hypot(destination.X-source.X, destination.Y-source.Y) / 30)
ETA          = ceil(routeParsecs / effectiveFTLSpeed)
supplyDistance = minimum distance from destination to an owned Colony or owned Outpost
legal if supplyDistance <= effectiveRange
```

No military-only supply graph is introduced.

The original Wormhole `+3 parsecs/turn` connection bonus remains deferred because MOOX does not yet persist authoritative Wormhole route state.

### 4. `empire.move_fleet` command

`MoveFleetPayload` becomes:

```go
type MoveFleetPayload struct {
    FleetID             core.ID   `json:"fleet_id"`
    DestinationSystemID core.ID   `json:"destination_system_id"`
    ShipIDs             []core.ID `json:"ship_ids,omitempty"`
}
```

Semantics:

- fixed Colony/Outpost special Fleet: `ShipIDs` must be omitted/empty and existing movement behavior is preserved;
- ordinary combat Fleet: omitted/empty `ShipIDs` means move the entire Fleet;
- an explicitly supplied list must be sorted, unique, non-zero and a subset of that Fleet's `ShipIDs`;
- supplying all member Ships is normalized to whole-Fleet movement; no pointless replacement Fleet is allocated;
- supplying a proper non-empty subset performs an atomic split+move;
- movement is legal only while the source combat Fleet is stationary at a valid System;
- destination must differ from source and pass range/supply legality;
- an already-moving combat Fleet is rejected; Hyperspace Communications rerouting remains deferred.

For a proper subset move, resolver order inside that one command is:

1. validate the complete command without mutating state;
2. allocate `newFleetID = state.NewID()`;
3. remove selected Ship IDs from the source Fleet;
4. create a new stationary combat Fleet at the same source System with the selected Ship IDs;
5. emit `empire.fleet_split`;
6. start transit on the new Fleet;
7. emit `empire.fleet_movement_started`.

This gives the original one-action selected-Ship movement UX while keeping stable MOOX Fleet identities deterministic.

### 5. Explicit `empire.split_fleet` command

New command:

```go
const CommandSplitFleet = "empire.split_fleet"

type SplitFleetPayload struct {
    FleetID core.ID   `json:"fleet_id"`
    ShipIDs []core.ID `json:"ship_ids"`
}
```

Legality:

- owned ordinary combat Fleet only;
- stationary at a valid System;
- `ShipIDs` sorted, unique, non-zero;
- selected Ships must all belong to source Fleet;
- selection must be a **proper non-empty subset**; selecting none or all is rejected;
- source and resulting Fleet must both remain non-empty.

Mutation:

- `NewFleetID = state.NewID()`;
- source retains unselected Ships;
- new Fleet receives selected Ships at the same System;
- both lists remain sorted;
- `StrategicFleets` is restored to ascending Fleet-ID order after insertion.

Event:

```text
empire.fleet_split
{
  empire_id,
  source_fleet_id,
  new_fleet_id,
  system_id,
  moved_ship_ids
}
```

`moved_ship_ids` is emitted sorted.

### 6. Explicit `empire.merge_fleets` command

New command:

```go
const CommandMergeFleets = "empire.merge_fleets"

type MergeFleetsPayload struct {
    TargetFleetID core.ID `json:"target_fleet_id"`
    SourceFleetID core.ID `json:"source_fleet_id"`
}
```

Legality:

- IDs must be non-zero and different;
- both Fleets owned by the issuing Empire;
- both ordinary combat Fleets;
- both stationary;
- same non-zero `AtSystemID`;
- no special Fleet kind;
- no duplicate concrete Ship membership may result.

Mutation:

- target Fleet ID survives;
- target receives the sorted union of both `ShipIDs`;
- source Fleet is removed;
- no new ID is allocated;
- global `StrategicFleets` remains sorted by Fleet ID.

Event:

```text
empire.fleets_merged
{
  empire_id,
  target_fleet_id,
  source_fleet_id,
  system_id,
  ship_ids
}
```

`ship_ids` is the final sorted target composition.

No automatic co-location merge occurs on arrival. Separate Fleet identities remain distinct until an explicit merge command. This is necessary because MOOX Fleet IDs are authoritative while original stack grouping was UI-derived.

### 7. Movement events

Existing event names remain stable:

```text
empire.fleet_movement_started
empire.fleet_movement_progressed
empire.fleet_arrived
```

`FleetMovementStartedEvent` remains the observability snapshot of the movement decision and will contain:

```text
fleet_id
empire_id
source_system_id
destination_system_id
remaining_turns
ftl_speed
fuel_range_parsecs
supply_distance_parsecs
```

For combat Fleets, `ftl_speed` and `fuel_range_parsecs` are the derived current-Empire/effective-Fleet values used to validate that command. They are **event facts**, not persisted combat-Fleet caches.

Progress and arrival events keep their existing minimal payloads. Combat transit progression uses persisted destination/remaining-turn state exactly like fixed special fleets but does not require persisted combat `FTLSpeed` because each turn merely decrements the already-established ETA. New research acquired after departure does not retroactively change an in-flight ETA in this slice.

### 8. Turn/blockade ordering

Gate 3 must preserve the existing resolver boundary:

```text
commands / movement orders
...
advanceStrategicFleetTransit
advanceConstruction
recomputeSystemBlockades
advancePopulationTransfers
```

Therefore:

- movement order immediately clears source `AtSystemID`, removing that Fleet from source blockade presence;
- transit arrival materializes destination `AtSystemID` before the same resolution's blockade recomputation;
- an arriving hostile combat Fleet can establish blockade in that same turn;
- no separate departure/arrival blockade event is required because blockade is derived state.

### 9. Hostile destination boundary

Slice 04 does not prohibit a hostile destination merely because hostile strategic presence/Colony exists there.

Arrival becomes ordinary stationary combat-Fleet state and can contribute to blockade. No Fleet is consumed and no battle outcome is fabricated.

This is an explicit temporary architecture boundary. Slice 06 will introduce the deterministic hostile encounter / `BattleSession` handoff and may interpose encounter state before long-lived hostile co-location becomes observable.

### 10. Observer, save/load and deterministic replay

No second event-sourced state reducer is introduced.

Authoritative determinism remains command/resolver based:

- a given starting state + ordered command batches must allocate the same Fleet IDs through `GameState.NewID()`;
- subset move must deterministically emit split before movement-start;
- explicit split/merge commands must reproduce exact membership and Fleet identities;
- JSON save/load under schema 19 must preserve stationary/transit Fleet state and exact Ship membership;
- Observer and PlayerView clones must deep-copy all `ShipIDs` slices so client mutation cannot alter authoritative state;
- identical-session replay tests must compare final state and domain events.

Because movement-start events carry derived movement facts, tests will also prove that replay with the same starting state/rules derives identical FTL/range/supply values.

### 11. Command transaction rule

The existing resolver mutates the working state while processing ordered commands. Gate 3 helpers for split+move and merge must therefore perform **all command-local validation before the first mutation that belongs to that command**. A rejected subset move must not leave a partially split Fleet or consume a `NextID`.

Focused tests must explicitly cover this failure atomicity.

### 12. Gate-3 implementation order

The accepted implementation order is:

1. schema-19 Core validation changes for legal ordinary combat transit and strict Fleet/Ship membership invariants;
2. shared current-Empire combat movement profile/range helper;
3. command payloads/decoders for optional subset move, split and merge;
4. pure validation/composition helpers;
5. atomic subset split+move and explicit split/merge mutations/events;
6. generalize transit advancement to ordinary combat Fleets;
7. session/legal-action and Observer surfaces where currently exposed;
8. focused Core/Game/Session replay/save-load/blockade compatibility tests;
9. full Gate-3 test pass before Gate 4.

## Gate 2 conclusion

Gate 2 is complete. The schema-19 state semantics, movement derivation, command payloads, event ordering, merge/split identity rules, supply lookup, hostile-destination boundary and replay/save-load requirements are now fixed for Gate 3.

No gameplay implementation is part of Gate 2. The next permitted action is Gate 3 implementation against this contract.

## Gate 3 implementation

Gate 3 implements the accepted Gate-2 contract without pulling hostile encounter resolution or tactical combat into this slice.

### Core schema 19

`StateSchemaVersion` is now **19**. Ordinary combat Fleets with `SpecialKind == none`:

- require sorted, unique, non-empty concrete `ShipIDs`;
- require exactly one semantic location mode: stationary `AtSystemID` or transit `DestinationSystemID + RemainingTurns`;
- keep persisted `FTLSpeed == 0`, because ordinary combat movement speed is derived from current Empire state;
- reject locationless combat Fleets and cached combat FTL speed;
- preserve the existing fixed Colony/Outpost special-Fleet installed `FTLSpeed` contract.

Schema-19 combat transit round-trips through `MarshalState` / `UnmarshalState` with exact Fleet membership, Ship state and `NextID`.

### Current-Empire combat movement profile

`combatFleetMovementProfile` validates each concrete member Ship and derives:

```text
effective FTL = best currently known Warp Drive + Trans-Dimensional (+2 when present)
effective range = minimum member effective range
```

The currently supported military surface has no per-Ship strategic range special, so each member uses the Empire's current best Fuel Cell range. Built `Ship.Spec.FTLSpeed` / `FuelRangeParsecs` remain immutable design provenance and do not pin later strategic capability.

Tests explicitly prove a Nuclear/Standard build snapshot moves with later Fusion/Deuterium capability and receives the Trans-Dimensional +2 speed modifier.

### Movement, split and merge commands

`empire.move_fleet` now accepts optional sorted/unique `ship_ids`:

- omitted/empty or an explicit all-member list moves the existing Fleet;
- a proper subset is validated completely before mutation, allocates exactly one new Fleet ID through `GameState.NewID()`, leaves the unselected Ships in the source Fleet, and moves the selected Ships in the new Fleet;
- the atomic subset path emits `empire.fleet_split` before `empire.fleet_movement_started`;
- a rejected subset move does not change Fleet composition and does not consume `NextID`;
- fixed Colony/Outpost special Fleets retain their prior whole-Fleet behavior and reject `ship_ids` selection.

New explicit commands are implemented:

- `empire.split_fleet {fleet_id, ship_ids}` - owned stationary ordinary combat Fleet, proper non-empty subset only;
- `empire.merge_fleets {target_fleet_id, source_fleet_id}` - owned stationary ordinary combat Fleets at the same System; target ID survives, source is removed.

Both operations preserve sorted unique non-empty Ship composition. In-transit split, merge and reroute remain rejected.

### Transit, arrival and blockade timing

Combat Fleets reuse the existing semantic destination / remaining-turn countdown. The movement-start event records the derived FTL speed, range and supply distance used to authorize the order, while the combat Fleet does not persist those derived values.

Transit progress decrements the already-established ETA; later technology changes do not retroactively alter an in-flight ETA in this slice. Arrival clears destination/remaining-turn state and materializes `AtSystemID` before the existing blockade recomputation. A focused resolver test proves an ETA-1 hostile arrival can establish blockade in that same resolution.

Hostile co-location is still a temporary strategic placeholder. Slice 06 owns automatic hostile encounter / `BattleSession` handoff.

### Save/load, Observer and replay

Gate-3 coverage proves:

- exact schema-19 combat transit save/load;
- Observer `ShipIDs` deep-clone isolation;
- two identical GameSessions produce identical final strategic state and event history;
- deterministic event order for subset movement: `fleet_split` -> `fleet_movement_started` -> `fleet_movement_progressed` when ETA remains;
- Colony Ship, Outpost Ship, blockade and previous military construction behavior remain compatible.

### Gate-3 QA

Passed before updating this document:

```text
go test ./internal/core ./internal/game ./internal/session -run "(CombatFleet|Strategic|Military|ColonyShip|Outpost|Blockade|Observer|Save|Load)" -count=1
go test ./... -count=1
gofmt -w <all changed Go files>
git diff --check
```

The final full repository run passed all packages. Gate 4 remains responsible for a fresh conflict check, final gofmt/test/vet/focused-regression/diff QA, commits and slice closure.

## Gate 4 closure

Fresh Gate-4 verification found no foreign writer, `main` still based on `7fd43f4` plus the Slice-04 work, and exactly one expected Slice-04 OPEN marker.

Final QA passed:

```text
gofmt -w <all changed Go files>
go test ./... -count=1
go vet ./...
go test ./internal/core ./internal/game ./internal/session -run "(CombatFleet|Strategic|Military|ColonyShip|Outpost|Blockade|Observer|Save|Load|Replay)" -count=1
git diff --check
```

The focused coverage includes composition-preserving split/merge, rejected subset-move atomicity, schema-19 save/load, Observer isolation and deterministic replay, so no Ship instance duplication/loss is left across the supported Slice-04 paths.

Gameplay, tests, permanent evidence and the Gate-3 recovery state are committed as:

```text
f297480 game: add combat fleet movement
```

The closing documentation commit removes the OPEN marker, marks this slice closed in live status/HISTORY and promotes Slice 05 **Command Points / ship maintenance** as the next prepared objective. No push was performed.