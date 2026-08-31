# Colony Base and same-system colonization - MOO2 1.31 Gate 1

Date: 2026-08-31

Status: **Gate 4 complete; slice closed**

Active marker: `docs/slices/_OPEN_COLONY_BASE_SAME_SYSTEM_COLONIZATION_2026-08-31.md`

Planned slice: `docs/slices/PLANNED_01_COLONY_BASE_SAME_SYSTEM_COLONIZATION.md`

## Purpose

Resolve the original Master of Orion II 1.31 Colony Base lifecycle before MOOX implements same-system colonization. The immediate fidelity issue is that MOOX currently treats `colony_base` only through the generic Building path, while the original gives Building ID 11 additional system-level colonization semantics.

Original copyrighted executable/data are private reference evidence only and are not distributable MOOX content.

## Reference identity and extraction

Private reference executable:

- `C:\ASH\Temp\mastori2\Orion2.exe`
- SHA-256 `7AE2AC2E5904CA330009AF2827279D889906B0B9B7A8854C38EB707A56E955B5`
- DOS/4GW-bound inner LE image
- LE header file offset `0x292E4`
- LE object 1 virtual base `0x10000`, virtual size `0x160695`
- LE object 2 virtual base `0x178000`, virtual size `0x5DCD0`

Gate 1 used the repository's pure-Go `internal/moo2exe` reader to extract LE object 1 temporarily and local `objdump` only for private disassembly. Temporary extracts/helpers are not permanent project artifacts and will be deleted before Gate 1 closes.

The executable contains Watcom debug-symbol records. Relevant object-relative symbol offsets map to code VAs by adding object-1 base `0x10000`.

## Existing normalized facts

Evidence level: **original-observed / already normalized**.

`data/rulesets/moo2-1.31/buildings.json` records:

- Building ID / production ID: `11`
- semantic key: `colony_base`
- Technology ID: `40`
- Production cost: `200 PP`
- Maintenance: `0 BC/turn`

The original `_buildings` table is LE object 2 offset `0x6B3D`, 49 records x `0x13` bytes. The cost field is record `+0x08`; Maintenance is `+0x0C`.

`Colony_Product_Cost_` object offset `0xD0DD6` / VA `0xE0DD6` reads ordinary positive Building products from this table. Building ID 11 does not enter any of its special-product cost branches, so Colony Base uses the ordinary table cost of 200 PP. The ship-government cost reductions are not applied to this Building path.

## Relevant original symbols

| Symbol | Object-1 offset | VA |
| --- | ---: | ---: |
| `Init_Colony_` | `0x02D75` | `0x12D75` |
| `Planet_Is_Colonizable_` | `0x614A1` | `0x714A1` |
| `Planet_Colonization_In_Main_Screen_` | `0x7B2DE` | `0x8B2DE` |
| `Colony_Can_Build_Product_` | `0xD11BC` | `0xE11BC` |
| `Apply_Production_` | `0xD36DF` | `0xE36DF` |
| `Make_New_Colony_Or_Outpost_` | `0xD5EB3` | `0xE5EB3` |
| `Make_New_Colony_` | `0xD6071` | `0xE6071` |
| `Star_N_Colonizable_Planets_For_Player_` | `0xD60C8` | `0xE60C8` |
| `AI_Colonize_` | `0xD65F8` | `0xE65F8` |
| `Player_Colony_Base_At_Star_` | `0xEDA3F` | `0xFDA3F` |
| `Colonization_` | `0xEDB01` | `0xFDB01` |
| `Display_Report_Aux_` | `0xEE63E` | `0xFE63E` |
| `Remove_Building_` | `0x045EA` | `0x145EA` |
| `New_Colony_Selection_Popup_` | `0xB8DB8` | `0xC8DB8` |
| `New_Outpost_Selection_Popup_` | `0xB8D49` | `0xC8D49` |

Embedded debug/local symbol names in the same report/colonization area also include `_really_trash_colony_base_fmt`, `_colonize_preselected_planet_fmt` and `_outpost_preselected_planet_fmt`, independently showing that Colony Base has dedicated colonization/report handling.

## Finding 1 - Colony Base buildability is same-system dependent

Evidence level: **direct original code**.

`Colony_Can_Build_Product_` VA `0xE11BC` performs the generic Building checks first:

1. product/building ID must be in the normal Building range;
2. required Technology must be known;
3. the Colony must not already own that Building.

For product ID `0x0B` / 11 it then enters a dedicated branch at `0xE1394`.

That branch:

- obtains the source Colony's Planet;
- obtains that Planet's Star/System;
- scans the five Planet slots of that same Star;
- ignores unused `-1` Planet slots;
- accepts normal planet-state value `3`;
- tests the candidate Planet's Colony link;
- returns the normal buildable state only if at least one such Planet is uncolonized (`Planet colony link == -1`).

Therefore Colony Base is not merely “Technology 40 known”. Normal construction legality requires an empty normal colonizable Planet in the **same StarSystem** as the constructing Colony.

The original function has differentiated non-buildable return states (`0` vs `2`) around this branch; their exact UI-reason semantics are not required for authoritative MOOX legality as long as only the proven normal-buildable result is exposed as legal.

## Finding 2 - target Planet is not bound when Colony Base is queued

Evidence level: **direct original code plus call-flow correlation**.

The Building-ID-11 branch in `Colony_Can_Build_Product_` only answers whether any eligible Planet exists in the source System. It stores no chosen Planet.

Later, `Colonization_` VA `0xFDB01` discovers completed Colony Bases and invokes `New_Colony_Selection_Popup_` VA `0xC8DB8` as part of the colonization/report flow. This establishes the human target choice after the Base exists, not as a target persisted with Construction progress.

MOOX therefore should not invent a queue-time target Planet unless Gate 2 deliberately chooses a modernization. The fidelity-first shape is a normal 200-PP Building project followed by an explicit colonize action from the completed Base.

## Finding 3 - a completed Colony Base is Building ID 11, not a separate ready entity

Evidence level: **direct original code**.

The Colony Building-ownership array starts at `colony+0x136`. `Colony_Can_Build_Product_` indexes it as `colony+0x136+productID` to reject already-owned Buildings.

For Colony Base, Building ID 11 therefore maps exactly to:

```text
colony + 0x136 + 0x0B = colony + 0x141
```

`Player_Colony_Base_At_Star_` VA `0xFDA3F` scans the five Planets in a Star, resolves their Colonies, matches Colony owner to the requested player and returns a Colony when `colony+0x141 > 0`.

Thus `+0x141` is not a separate “Colony Base ready flag”; it is the ordinary ownership slot for Building 11. A completed Base can exist as an owned Building while waiting for colonization.

This corrects the initial planning assumption that Colony Base should never be materialized in the Building list. In MOOX it should remain a Building, but one with consumable colonization semantics.

## Finding 4 - Colony Base is a first-class colonizer source

Evidence level: **direct original code**.

At the beginning of `Colonization_` VA `0xFDB01`, for each relevant Star the original resolves three possible colonization sources:

- `Player_Colony_Base_At_Star_` -> completed Colony Base;
- `Player_Ship_Type_At_Star_(..., 1)` -> Colony Ship;
- `Player_Ship_Type_At_Star_(..., 4)` -> Outpost Ship.

If none exists the Star is skipped. Colony Base is therefore explicitly parallel to the special colonizer ships at this strategic/report boundary.

`Display_Report_Aux_` VA `0xFE63E` is the sole direct caller of `Colonization_` found by the Gate-1 xref pass (`call 0xFDB01` at VA `0xFE710`). This places human Colony Base colonization in report handling after normal turn processing rather than inside the Production function itself.

## Finding 5 - successful Colony Base colonization consumes Building 11

Evidence level: **direct original code**.

After a normal-Colony target is selected, `Colonization_` calls `Make_New_Colony_` VA `0xE6071` at VA `0xFDE32`.

Immediately afterward the source type determines consumption:

- Colony Base path: compute the source Colony index, set product/building ID `0x0B`, call `Remove_Building_` VA `0x145EA`;
- ship path: resolve the ship index and call the ship-destruction path.

Therefore a Colony Base remains Building 11 until successful use, and is then **consumed by removing Building 11 from its source Colony**.

A cancelled/no-target colonization must not be modeled as successful consumption unless the remaining report-flow evidence proves otherwise.

## Finding 6 - destination Colony uses the normal Colony creation path

Evidence level: **direct original code; corroborated by the completed Colony Ship slice**.

`Colonization_` uses `Make_New_Colony_`, the same normal-Colony wrapper already traced for Colony Ship colonization. `Make_New_Colony_Or_Outpost_` calls `Init_Colony_`, links the Planet/Colony, initializes a normal Colony and creates the normal founding Population entry.

The completed success branch at VA `0xFDE1B..0xFDE80` makes this stronger than negative evidence alone:

1. target Planet and player are passed to `Make_New_Colony_` at `0xFDE32`;
2. `Make_New_Colony_` receives no source-Colony pointer and is the same generic wrapper used by Colony Ship colonization;
3. the only source-Colony-specific success mutation immediately afterward is `Remove_Building_(sourceColony, 11)` at `0xFDE51`;
4. no source Population count/cohort/job field is read or decremented in this source-Base branch.

Therefore the original Colony Base **does not transfer or consume Population from the source Colony**. The new Colony receives the normal one-unit founding Population from the generic Colony initialization path, while source Population remains unchanged.

## Finding 8 - Production completion materializes Building 11 before colonization resolution

Evidence level: **direct original code**.

`Apply_Production_` VA `0xE36DF` handles completed positive Building products through `Add_Building_` VA `0x13FD9`; the direct call is at VA `0xE391F`.

`Add_Building_` writes the Building ownership array at VA `0x14562`:

```text
colony + 0x136 + buildingID = 1
```

For Building 11 this is exactly `colony+0x141 = 1`.

Thus original turn semantics are two-stage:

```text
Production completes 200 PP Colony Base
-> Add_Building_(11)
-> post-turn/report Colonization_ discovers that completed Base
-> player resolves colonize-or-trash decision
```

The target is therefore not part of Construction progress and the Base is a real completed Building at the decision boundary.

## Finding 9 - cancelling Colony Base target selection leads to colonize-or-trash resolution

Evidence level: **direct original code**.

When `New_Colony_Selection_Popup_` returns no selected Planet while a Colony Base source exists, `Colonization_` enters a dedicated confirmation branch around VA `0xFDCC3..0xFDD41`.

That branch:

- reads data offset `0x6C16`, which is Building-11's original cost field (`0x6B3D + 11*0x13 + 0x08`) = **200**;
- divides it by two, yielding **100**;
- formats the Colony-Base-specific confirmation path (the embedded debug/local symbol area contains `_really_trash_colony_base_fmt`);
- if the player declines, calls `New_Colony_Selection_Popup_` again;
- if the player confirms, adds the 100 value to the player's Treasury field and clears `colony+0x141`, then searches for another Colony Base at that Star.

So the original user-facing resolution is effectively mandatory:

```text
completed Colony Base
-> choose a legal same-system Colony target
   OR
-> trash the Colony Base for 100 BC
```

It is **not** an ordinary Building that the player can simply leave unresolved and proceed indefinitely. Multiple Colony Bases are resolved sequentially by repeatedly searching the Star after one is consumed/trashed.

For MOOX, the closest headless semantic equivalent is a derived pending Colony-Base decision that must be resolved before the next strategic turn can advance. Whether this is represented as persisted pending state or derived purely from owned Building 11 is a Gate-2 design choice.

## Finding 10 - completion/decision timing is post-Production report handling

Evidence level: **direct original call flow**.

The Gate-1 xref pass found the sole direct call to `Colonization_` at VA `0xFE710`, inside `Display_Report_Aux_` VA `0xFE63E`. That report dispatcher invokes Colony Base/ship colonization after normal turn mechanics have produced reportable state.

This is distinct from `Apply_Production_`: Production completes and adds Building 11 first; colonization is subsequently resolved at the report boundary. In MOOX's headless/session architecture, a fidelity-preserving mapping is therefore:

1. current turn Production may complete Colony Base;
2. the resulting authoritative state exposes a pending Colony Base resolution;
3. normal next-turn advancement is rejected while such resolution remains;
4. a colonize or trash command resolves it;
5. ordinary strategic advancement may then continue.

This preserves original ordering without coupling gameplay authority to a UI popup.
## Finding 7 - in-progress Colony Base can become strategically pointless

Evidence level: **direct original AI/planner code; player cancellation policy not yet generalized**.

An original colonization-planning path around VA `0xD7881` scans Colonies whose current construction product is `0x0B`. It derives that Colony's System and checks a candidate colonizable-Planet list. If no appropriate same-system candidate remains, it changes the current product away from Colony Base (`0xFFFE` in that planner path).

This confirms that Colony Base validity depends on remaining same-system targets over time. It does **not** by itself prove that MOOX should automatically cancel a human player's in-progress Colony Base at the same moment; that policy remains separate from the proven buildability rule and report-time target choice.

## Current MOOX gap

Current MOOX behavior is only partially correct:

- normalized Tech/cost/Maintenance are correct;
- generic Building Construction can materialize `colony_base` in `Colony.Buildings`, which now matches the original completed-state representation;
- however generic `AvailableBuildingChoices` currently lacks the original same-system empty-Planet legality check;
- there is no explicit completed-Colony-Base colonization action;
- there is no successful-use consumption of Building 11;
- therefore same-system colonization through Colony Base is not implemented.

The required fix is not to stop Colony Base from ever being a Building. It is to make its buildability and post-completion use special while preserving the normal Building completion representation.

## Gate-1 boundary and unresolved policy detail

Gate 1 has enough direct evidence to define the gameplay contract. One original AI/planner detail remains intentionally **not generalized**: an AI path cancels an in-progress Colony Base when its reserved candidate list no longer contains a suitable same-system Planet. This does not prove an equivalent automatic human Construction cancellation rule.

For the first authoritative MOOX implementation, Gate 2 should therefore distinguish:

- **proven legality:** Colony Base may only be queued when a legal same-system empty normal Planet exists;
- **proven completion behavior:** normal Building 11 completes at 200 PP;
- **proven resolution behavior:** completed Building 11 must be resolved by colonization or trash/refund before normal progression continues;
- **deferred AI policy:** proactive cancellation/replanning while Construction is still in progress.

## Proposed Gate-2 contract

The fidelity-first implementation proposal is:

1. **Keep Colony Base as a normal persisted Building product.** `ConstructionProjectBuilding` with project ID `colony_base`; generic Production progression and `colony.building_completed` remain authoritative.
2. **Specialize buildability.** Offer/accept Colony Base only when Technology 40 is known, the source Colony does not already own Building 11, and at least one empty normal colonizable Planet exists in the source Colony's StarSystem.
3. **No queue-time target.** Do not add `TargetPlanetID` to Construction state. Target selection is post-completion.
4. **Derived pending resolution.** Any owned completed `colony_base` is an unresolved Colony-Base decision. Expose legal resolution actions and reject next-turn advancement while one or more such Bases remain unresolved. Because this can be derived from existing Building ownership, **no Core schema bump is required** solely for the pending state.
5. **Colonize command.** Add an authoritative command such as `colony.colonize_with_base { source_colony_id, target_planet_id }`; target must be an empty normal colonizable Planet in the same StarSystem and source must own Building 11.
6. **Normal destination initialization.** Reuse the already accepted Colony Ship normal-Colony initializer semantics: link target Planet, create new Colony, one assimilated founding Population unit, recalculate economy/capacity. **Do not subtract Population from the source Colony.**
7. **Consume Base on success.** Remove `colony_base` from the source Colony atomically after successful founding.
8. **Trash command.** Provide a Colony-Base-specific discard resolution (or a generic Building-scrap command if Gate 2 deliberately broadens that primitive) that removes Building 11 and credits exactly **100 BC** to the owning Empire Treasury.
9. **Multiple Bases.** Permit multiple unresolved Colony Bases across the Empire/System, but normal turn advancement stays blocked until all are resolved; resolution order must be deterministic for Observer/legal-action projection, while the player may choose which legal pending Base to resolve first unless original ordering becomes materially relevant.
10. **Events/replay.** Keep generic Building completion event, add dedicated Colony-Base colonization and trash/refund events with source Colony, target Planet/new Colony where applicable, and Treasury delta for trash.
11. **Save/load and Observer.** Pending resolution is derived from Building ownership, so exact schema-16 round-trip remains valid; legal actions/Observer projection must surface pending sources and legal targets without client-side inference.
12. **Scope guard.** Outpost targets/conversion, Transport/invasion, AI target selection/replanning and broader Building-scrap UX remain outside this slice unless required by the accepted command primitive.

Gate 2 was explicitly accepted on 2026-08-31. Gate 3 is complete; Gate 4 QA/commit/closure is pending.
## Gate 3 implementation result

Status: **Gate 3 complete**.

The accepted Gate-2 contract is now implemented without changing Core persistence schema.

### Authoritative runtime surface

- `internal/game/colony_base.go` defines the Colony Base semantic constants from the verified original evidence: Building/production ID 11, Technology 40, base cost 200 PP, and trash refund 100 BC.
- `AvailableBuildingChoices` only exposes `colony_base` when the source Colony has at least one currently empty Planet in its own StarSystem. Direct `colony.queue_building` validation enforces the same rule so a client cannot bypass legal-action projection.
- Construction remains the generic persisted `ConstructionProjectBuilding` path. No target Planet is stored in Construction state.
- `PendingColonyBaseResolutions` derives unresolved completed Bases directly from Colony Building ownership. It projects source Colony, System, currently legal same-system target Planet IDs, and the 100-BC trash refund. Targetless completed Bases remain pending so the original trash resolution remains possible.
- `GameSession.CompleteTurn` rejects strategic turn advancement while any Colony Base resolution is pending.
- `GameSession.ColonyBaseResolutions` provides the seat-scoped legal resolution surface; `PlayerView` exposes only that player's pending Bases, while `ObserverView` exposes all pending Bases.

### Post-resolution commands

Two commands resolve a completed Base during `PhasePostResolution`:

```text
colony.colonize_with_base
  source_colony_id
  planet_id

colony.trash_colony_base
  source_colony_id
```

`colony.colonize_with_base` requires a completed Base owned by the acting Empire and an empty Planet in the source Colony's StarSystem. It creates the normal destination Colony, consumes Building `colony_base`, and leaves the source Colony Population unchanged.

`colony.trash_colony_base` consumes Building `colony_base` and credits exactly 100 BC directly to the owning Empire's current Treasury balance. It is legal even when no target Planet remains, matching the original completed-Base fallback.

### Shared founding semantics

Colony Ship and Colony Base now call the same internal `prepareFoundedColony` helper. The shared initializer:

- allocates the next stable Colony ID;
- links the requested empty Planet at commit time;
- creates one assimilated Farmer founder for the owning Empire;
- recalculates Colony economy/capacity using the existing authoritative rules.

This removes duplicated founding logic and keeps the already accepted Colony Ship behavior and the new Colony Base behavior on one deterministic semantic path.

### Events

Successful Base founding emits, in deterministic order:

1. `empire.planet_colonized`;
2. `colony.colony_base_colonized`.

Trash resolution emits:

- `colony.colony_base_trashed`, including source Colony/System, Building ID, 100-BC refund, previous Treasury balance and current Treasury balance.

The existing generic `colony.building_completed` event remains the Production-completion signal before the post-resolution decision.

### Persistence and replay

Core `StateSchemaVersion` remains **16**. No pending-decision field was added to `GameState`; pending Colony Base decisions are fully reconstructed from persisted Building ownership plus Galaxy/Colony links.

Focused tests verify:

- Colony Base is unavailable without a same-system empty target and direct queue bypass is rejected;
- a completed Base projects deterministic legal targets;
- cross-system founding is rejected without state mutation;
- successful founding creates the second Colony, consumes the Base, and does not subtract source Population;
- targetless completed Bases can be trashed for exactly 100 BC;
- duplicate consumption is rejected;
- pending resolution survives an exact schema-16 marshal/unmarshal round trip;
- identical Colony Base replays produce identical DomainEvents and identical serialized Core state;
- GameSession blocks `CompleteTurn` until the Base is resolved;
- Player/Observer projections are authority-correct and clone-isolated;
- Observer event order, seat authority, command sequence and revision are stable;
- both colonize and trash resolutions release the turn guard.

Gate-3 QA completed successfully with:

- focused Colony Base Game tests;
- focused Colony Base GameSession tests;
- Colony Ship founding regressions after moving both colonizers to the shared founding helper;
- neighboring Construction, Building-choice, Treasury and deterministic-session regressions;
- `go test ./... -count=1`;
- `go vet ./...`.

Gate 4 remains responsible for final `gofmt`/`git diff --check`, permanent HISTORY/handoff closure, commits, and deletion of the active `_OPEN_` marker.
## Gate 4 closure

Status: **complete**.

- Final `gofmt` passed on all touched Go files.
- `go test ./... -count=1` passed across the repository.
- `go vet ./...` passed across the repository.
- Focused Colony Base / Colony Ship / Building-choice regressions passed.
- `git diff --check` and staged `git diff --cached --check` passed.
- Runtime/tests/evidence were committed as `261e418` (`game: add colony base colonization flow`).
- Core `StateSchemaVersion` remains **16**.
- The dated `_OPEN_` marker is removed by the closing documentation commit; no gameplay slice remains open.