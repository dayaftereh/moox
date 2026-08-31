# Outpost Ship, Outpost state and supply range - MOO2 1.31 Gate 1

Date: 2026-08-31

Status: **Gate 1 complete; Gate 2 decision pending**

Active marker: `docs/slices/_OPEN_OUTPOST_SHIP_SUPPLY_RANGE_2026-08-31.md`

Planned slice: `docs/slices/PLANNED_02_OUTPOST_SHIP_SUPPLY_RANGE.md`

## Purpose

Resolve the original Master of Orion II 1.31 Outpost Ship lifecycle and the smallest authoritative Outpost/supply model needed by MOOX before implementation.

Original copyrighted executable/data are private reference evidence only and are not distributable MOOX content.

## Reference identity and method

Private reference executable:

- `C:\ASH\Temp\mastori2\Orion2.exe`
- SHA-256 `7AE2AC2E5904CA330009AF2827279D889906B0B9B7A8854C38EB707A56E955B5`
- DOS/4GW-bound inner LE image
- LE header file offset `0x292E4`
- LE object 1 virtual base `0x10000`, virtual size `0x160695`

Gate 1 reused the repository's `internal/moo2exe` reader to extract LE object 1 temporarily and local `objdump` for private disassembly. The temporary extractor/source and object dump are removed before Gate 1 closes.

Evidence labels in this document:

- **direct original code**: observed directly in the 1.31 executable;
- **original-derived**: semantic interpretation supported by multiple direct paths;
- **MOOX proposal**: implementation architecture, not a claim about original storage layout.

## Relevant original symbols

| Symbol | Object-1 offset | VA |
| --- | ---: | ---: |
| `Load_Outpost_Ship_Design_` | `0x46586` | `0x56586` |
| `Best_Warp_Drive_` | `0x46726` | `0x56726` |
| `Can_Build_Outpost_` | `0x6022D` | `0x7022D` |
| `Set_Build_Outpost_Button_Flag_` | `0x61DD8` | `0x71DD8` |
| `Planet_Is_Colonizable_` | `0x614A1` | `0x714A1` |
| `Planet_Has_Outpost_` | `0x69C1D` | `0x79C1D` |
| `Player_Has_Outpost_` | `0x69D86` | `0x79D86` |
| `Star_Is_Outpost_Star_` | `0x6A133` | `0x7A133` |
| `Building_Outposts_In_Main_Screen_` | `0x7A97A` | `0x8A97A` |
| `Player_Has_Colony_Or_Outpost_At_Star_` | `0xB6EBF` | `0xC6EBF` |
| `New_Outpost_Selection_Popup_` | `0xB8D49` | `0xC8D49` |
| `Compute_Star_Colony_Stuff_` | `0xD5296` | `0xE5296` |
| `Make_New_Colony_Or_Outpost_` | `0xD5EB3` | `0xE5EB3` |
| `Make_New_Outpost_` | `0xD607F` | `0xE607F` |
| `Star_N_Outpostable_Planets_For_Player_` | `0xD6132` | `0xE6132` |
| `Colonization_` | `0xEDB01` | `0xFDB01` |
| `Ship_Range_` | `0xEF496` | `0xFF496` |
| `Star_In_Range_Of_Player_Aux1_` | `0xEF4E9` | `0xFF4E9` |
| `Star_In_Range_Of_Player_` | `0xEF5F8` | `0xFF5F8` |
| `Update_Player_Ship_Range_` | `0xF034D` | `0x10034D` |
| `Ship_Type_Cost_For_Player_` | `0xD0D98` | `0xE0D98` |
| `Colony_Can_Build_Product_` | `0xD11BC` | `0xE11BC` |

## Finding 1 - Outpost Ship identity, Technology and Production cost

Evidence level: **direct original code**.

`Load_Outpost_Ship_Design_` writes:

```text
design+0x11 = 4
```

The special-ship mapping remains:

```text
1 = Colony Ship
2 = Transport
4 = Outpost Ship
```

`Colony_Can_Build_Product_` dispatches special class `4` to the Technology-status byte at player `+0x184`. Technology status begins at player `+0x117`:

```text
0x184 - 0x117 = 0x6D = 109
```

The byte must equal the established owned-Technology state `3`. Normalized MOOX data identifies Technology 109 as `outpost_ship`.

`Ship_Type_Cost_For_Player_` directly assigns special class 4 a base Production cost of **100 PP** and then calls the same government ship-cost reducer used by Colony Ship.

For the currently modeled starting-government surface:

```text
base Outpost Ship = 100 PP
Feudal             = ceil(2 * 100 / 3) = 67 PP
other currently modeled starting governments = 100 PP
```

The original reduced enum corresponding to evolved Confederation would produce 34 PP, but evolved-government runtime remains outside this slice.

## Finding 2 - Outpost Ship installs the current best Warp Drive exactly like Colony Ship

Evidence level: **direct original code**.

`Load_Outpost_Ship_Design_` and `Load_Colony_Ship_Design_` have the same relevant fixed-special-ship design sequence:

- clear normal optional-special slots;
- set the special class byte (`4` for Outpost, `1` for Colony Ship);
- call `Best_Warp_Drive_` and store the selected drive in `design+0x13`;
- initialize the remaining fixed design fields;
- compute cost through `Ship_Type_Cost_For_Player_`.

Thus an Outpost Ship is built with the best Warp Drive available **at construction time**. The installed drive is part of the completed design/ship state; later drive research does not retroactively change an already-built special ship.

This is the same semantic boundary MOOX already uses for Colony Ship's persisted `FTLSpeed`.

## Finding 3 - Outpost Ship uses normal Fuel Cell strategic range

Evidence level: **direct original code plus previously normalized fuel table**.

`Ship_Range_` reads the owner's base strategic range from player `+0x324`. `Update_Player_Ship_Range_` derives that value from the best owned Fuel Cell technology.

The directly decoded ranges remain:

| Fuel technology | Range |
| --- | ---: |
| Standard Fuel Cells | 4 pc |
| Deuterium Fuel Cells | 6 pc |
| Iridium Fuel Cells | 9 pc |
| Urridium Fuel Cells | 12 pc |
| Thorium Fuel Cells | 255 pc |

`Load_Outpost_Ship_Design_` clears the normal optional-special slots just as Colony Ship does, so the fixed Outpost design does not silently gain an optional Extended Fuel Tanks modifier.

`Make_Ships_Move_To_` explicitly treats Outpost/Colony special ships in the strategic movement path. Therefore the existing MOOX special-ship transit model can be generalized rather than inventing a separate Outpost movement system.

## Finding 4 - Outpost deployment targets a Planet/body record, not merely a StarSystem

Evidence level: **direct original code**.

`Can_Build_Outpost_` receives a Planet index. It returns legal only when:

1. the Planet index resolves to a real Planet record;
2. the Planet's linked Colony-record index is `-1` (unoccupied);
3. `Planet_Has_Outpost_` is false.

`Star_N_Outpostable_Planets_For_Player_` independently scans all five Planet slots in a Star and counts every real Planet record whose Colony link is `-1`.

Importantly, neither of those direct Outpost predicates calls `Planet_Is_Colonizable_` or requires original planet-body state `planet+0x04 == 3`.

Therefore an Outpost target is **not restricted to a Colony-colonizable terrestrial body**. It is an unoccupied Planet/body record in the System. This is intentionally broader than Colony Ship target legality.

MOOX must therefore not implement Outpost deployment as only a system-level flag if it wants to preserve later conversion/planet occupancy semantics.

## Finding 5 - completed deployment creates a zero-Population Outpost using the legacy Colony record family

Evidence level: **direct original code**.

`Building_Outposts_In_Main_Screen_` calls `New_Outpost_Selection_Popup_`, locates an Outpost Ship (`design+0x11 == 4`), calls `Make_New_Outpost_`, then consumes/invalidates the used ship record.

`Make_New_Outpost_` is a thin wrapper over `Make_New_Colony_Or_Outpost_` with type `1`.

The common creation path calls `Init_Colony_`, links the chosen Planet to a `0x169`-byte Colony record and sets ownership. For type `1` it then writes:

```text
colony+0x06 = 1   // Outpost flag
colony+0x0A = 0   // zero Population entries
```

`Planet_Has_Outpost_`, `Player_Has_Outpost_` and `Star_Is_Outpost_Star_` all recognize Outposts by resolving Planet -> linked Colony record -> `colony+0x06 != 0`.

`Star_Is_Outpost_Star_` counts Outpost and normal-Colony records separately and returns true only when the System has at least one Outpost and **no normal Colony**. This also proves that the original data model permits mixed systems where Outposts and normal Colonies coexist on different planets.

### MOOX semantic implication

The original reuses a Colony storage record, but a MOOX `core.Colony` is an authoritative populated economic entity with Population/economy invariants. Treating a zero-Population Outpost as a fake Colony would leak special cases through Food, Growth, Construction, Treasury and relocation systems.

The fidelity-preserving semantic model is therefore likely a **dedicated persisted Outpost entity linked to one Planet**, even though the legacy executable stores it in the Colony record family.

## Finding 6 - Outposts are direct strategic supply origins

Evidence level: **direct original code**.

`Make_New_Colony_Or_Outpost_` ends by calling `Compute_Star_Colony_Stuff_` on the target Star.

`Compute_Star_Colony_Stuff_` clears and rebuilds the Star's player-presence bitfield at `star+0x38`. For every Planet with a linked Colony-family record, it reads that record's owner byte and ORs that player's bit into `star+0x38`.

Crucially, this presence calculation does **not** test `colony+0x06`. Normal Colonies and Outposts therefore contribute identically to the Star owner-presence bitfield.

`Star_In_Range_Of_Player_Aux1_` performs the strategic range test by scanning Stars whose `star+0x38` bit is set for the queried player and comparing squared map distance against that player's Fuel range.

Therefore:

```text
owned normal Colony at a Star -> supply origin
owned Outpost at a Star       -> supply origin
```

This is direct evidence, not merely the common gameplay description of Outposts.

## Finding 7 - supply extension becomes active immediately after successful Outpost deployment

Evidence level: **direct original call flow**.

The Outpost creation sequence is:

```text
Make_New_Outpost_
-> Make_New_Colony_Or_Outpost_(type=1)
-> Init_Colony_
-> set Outpost flag / zero Population
-> Compute_Star_Colony_Stuff_
-> return
```

Because `Compute_Star_Colony_Stuff_` rebuilds `star+0x38` before `Make_New_Outpost_` returns, subsequent range queries see the new Outpost immediately.

The supply effect is therefore not delayed until a later Growth/Construction/Treasury turn phase. A later movement/order decision can use the newly deployed Outpost as soon as deployment has successfully committed.

It cannot retroactively extend range for the movement that brought the Outpost Ship to the System; the ship must first arrive and then deploy.

## Finding 8 - blockade state does not remove an Outpost from the original range-origin predicate

Evidence level: **direct original code**.

The original blockade calculation uses strategic player presence and hostile ship state, but `Star_In_Range_Of_Player_Aux1_` itself selects supply origins from the `star+0x38` ownership/presence bitfield and does not consult the blockade state.

`Compute_Blockades_` does not clear the ownership/presence bit for a blockaded player. Thus a blockaded Colony/Outpost remains an ownership presence for the direct range calculation.

For this slice, MOOX should therefore **not suppress an Outpost supply origin merely because the Empire is currently blockaded at that System**.

This is separate from whether hostile encounter/control state should prevent the deployment action itself.

## Finding 9 - deployment UI has an additional legacy Star-control gate, but no new diplomatic rule is directly established here

Evidence level: **direct original code with unresolved packed-state semantic**.

`Set_Build_Outpost_Button_Flag_` directly requires:

- the selected strategic stack belongs to the current player;
- at least one ship in the stack is special class `4` Outpost Ship;
- at least one Planet in the Star passes `Can_Build_Outpost_`;
- a separate Star byte at `star+0x37` equals the current player.

The same `star+0x37` style gate also participates in the original Colony-Ship colonization UI path. It is **not** the Star owner display field (`star+0x14`) and is **not** the multi-player Colony/Outpost presence bitfield (`star+0x38`). Its complete combat/control lifecycle is coupled to strategic encounter state not yet modeled by MOOX.

No separate diplomacy relation check exists in `Can_Build_Outpost_`, `Make_New_Outpost_` or the planet target predicate itself.

### Gate-2 boundary

Do not invent a treaty/ownership rule in the Outpost slice. Preserve the same current headless boundary already accepted for Colony Ship and leave the legacy encounter/control-byte semantics for the planned strategic-hostile-encounter slice.

The minimum directly supported deployment authority is therefore:

```text
owned Outpost Ship
+ stationary at the target System
+ chosen unoccupied Planet/body in that same System
```

Current explicit MOOX hostility/blockade state remains observable but should not be silently treated as a new supply or deployment rule without additional direct evidence.

## Finding 10 - Outpost-to-Colony is replacement/conversion, not coexistence on one Planet

Evidence level: **direct original code**.

`Planet_Is_Colonizable_` accepts a normal Colony-Ship target only when the body itself has the original colonizable-body state and either:

- the Planet has no linked Colony-family record; or
- its linked record has `colony+0x06 != 0`, i.e. it is an Outpost.

`Make_New_Colony_Or_Outpost_` type `0` (`Make_New_Colony_`) detects an existing linked record and reuses/reinitializes it through `Init_Colony_`. The normal-Colony branch creates one founding Population entry and does not set the Outpost flag.

Thus a Colony founded on an Outpost Planet **replaces/converts** the Outpost on that Planet. The Outpost and Colony do not coexist on the same Planet afterward.

Two important boundaries follow:

1. an Outpost may be built on a body that is not Colony-colonizable;
2. such an Outpost cannot be converted by Colony Ship until/unless that body satisfies the normal `Planet_Is_Colonizable_` body predicate.

For the first MOOX implementation, same-owner Outpost conversion is sufficient. Cross-Empire Outpost conquest/destruction remains tied to future hostile encounter/combat rules.

## Finding 11 - Systems can contain multiple Empire presences

Evidence level: **direct original code**.

The Star presence field `star+0x38` is a bitfield, not a single owner. `Compute_Star_Colony_Stuff_` ORs one bit per owner found among the System's linked Colony/Outpost records.

Therefore Outpost state should remain attached to its owner and Planet, not represented as an exclusive `StarSystem.OwnerID`.

This matters for future diplomacy/combat and prevents a system-level Outpost flag from erasing legitimate multi-Empire presence on different planets.

## Current MOOX gap

Current Core schema 16 has:

- `GameState.Colonies []Colony`;
- `Planet.ColonyID`;
- `StrategicFleets` with Colony Ship special identity/transit;
- `nearestEmpireSupplyDistanceParsecs`, which currently considers only owned Colonies as supply origins;
- no persisted Outpost state and no `Planet.OutpostID`.

The existing `core.Colony` model is intentionally economic/populated, so forcing Outposts into `Colonies` would require broad special casing throughout already-authoritative systems.

## Proposed Gate-2 contract

The narrow fidelity-first implementation proposal is:

1. **Core schema 17 with dedicated Outpost state.** Add `GameState.Outposts []Outpost` and `Planet.OutpostID`, with semantic `Outpost { ID, EmpireID, PlanetID }`. Validation enforces one Planet link, no simultaneous Colony and Outpost on the same Planet, unique IDs and valid owner/Planet references.
2. **Outpost Ship special identity.** Add `StrategicFleetSpecialOutpostShip = "outpost_ship"`. Reuse the existing fixed special-ship Fleet state (`AtSystemID`, `DestinationSystemID`, `RemainingTurns`, `FTLSpeed`) rather than creating a second transit representation.
3. **Construction project.** Add semantic Construction project `outpost_ship`, Technology 109, base 100 PP, current Feudal 67 PP, and the same government ship-cost helper already used by Colony Ship.
4. **Installed drive.** Persist the best available FTL speed at completion exactly as Colony Ship does; later research does not mutate existing ships.
5. **Fuel range.** Generalize the fixed-special-ship movement validation so Colony Ship and Outpost Ship share the same Fuel Cell base-range table and current supply lookup.
6. **Supply origins.** Generalize Empire supply origins from `owned Colonies` to `owned Colonies + owned Outposts`. Do not suppress an Outpost merely because its System currently lists that Empire as blockaded.
7. **Planet-level deployment.** Add an authoritative command such as `fleet.deploy_outpost { fleet_id, planet_id }`. Require an owned stationary Outpost Ship, target Planet in that same System, `ColonyID == 0`, `OutpostID == 0`. Do **not** require Colony-style climate/body colonizability.
8. **Deployment consumption.** Successful deployment atomically allocates/links the Outpost, consumes the Outpost Ship and makes the System an immediate supply origin for subsequent commands.
9. **Conversion.** Extend Colony Ship colonization so a legal same-owner Outpost on a Colony-colonizable target Planet is atomically removed/replaced by the new Colony. Never persist Colony and Outpost simultaneously on one Planet.
10. **Same-system Colony Base conversion.** If a completed Colony Base targets a Planet containing the same Empire's Outpost and the body is otherwise a legal Colony target, use the same Outpost-replacement helper rather than creating contradictory occupancy. Gate 1 does not claim the original Colony Base UI path independently here; this is a consistency requirement once the semantic Outpost exists.
11. **Hostility scope.** Do not add new diplomacy/treaty permissions or pretend the legacy `star+0x37` control byte has been modeled. Strategic hostile encounter/control semantics remain deferred. Existing blockade state does not remove supply range.
12. **Observer/replay/save.** Persist Outposts in Core schema 17, expose them through normal state/Observer cloning, and add deterministic Construction/move/deploy/supply/conversion events and exact save/load/replay tests.
13. **Vertical test.** Prove the complete loop: Colony -> build Tech-109 Outpost Ship -> move within current Colony supply -> deploy on unoccupied same-System body -> consume ship -> new Outpost immediately extends supply -> another fixed special ship can legally reach a destination that was previously out of range.

No gameplay code is changed during Gate 1. Gate 2 is pending explicit acceptance or revision.

## Gate 2 decision - accepted 2026-08-31

The user accepted the proposed Gate-2 contract without modification. Gate 3 therefore implements the documented semantic boundary rather than reproducing the original packed Colony-record storage shape.

Accepted runtime contract:

- Core schema 17 owns a dedicated Planet-linked `Outpost` entity and reciprocal `Planet.OutpostID`;
- Outpost Ship is a fixed civilian StrategicFleet special kind with persisted installed FTL speed;
- Technology 109 unlocks `outpost_ship` Construction at 100 PP base and 67 PP under the currently modeled Feudal ship-cost rule;
- Colony Ship and Outpost Ship share only the fixed-special-ship strategic movement/range path; ordinary combat Fleet movement remains deferred;
- owned Colonies and owned Outposts are strategic supply origins and blockade alone does not suppress either origin in the range lookup;
- deployment is an explicit Planet-level action that consumes the Outpost Ship and makes the new Outpost authoritative immediately;
- a same-owner Outpost can be replaced atomically when a normal Colony is founded on that Planet, including the current Colony Ship and Colony Base founding paths;
- foreign-Outpost conquest/destruction, the legacy `star+0x37` control gate and broader encounter/diplomacy legality remain deferred;
- save/load, Observer cloning, deterministic replay and an end-to-end Outpost lifecycle are required regressions.

## Gate 3 implementation - complete

Gate 3 implements the accepted contract without expanding into military Ship design, generic combat Fleet movement, hostile encounter resolution or tactical combat.

### Core schema 17

`internal/core/state.go` now defines:

```text
GameState.Outposts []Outpost
Planet.OutpostID
Outpost { ID, EmpireID, PlanetID }
StateSchemaVersion = 17
```

Core validation now enforces:

- valid unique Outpost identity and valid Empire/Planet references;
- at most one Outpost per Planet;
- at most one Colony per Planet;
- no simultaneous Colony and Outpost on one Planet;
- reciprocal `Planet.ColonyID` and `Planet.OutpostID` links;
- `outpost_ship` as a valid semantic Construction project kind.

This deliberately keeps Outpost out of `core.Colony`, so zero-Population Outposts cannot leak into Population, Food, Growth, Construction, Treasury or relocation calculations.

### Outpost Ship Construction

Gate 3 adds:

```text
Technology:       109
Project ID:       outpost_ship
Project kind:     outpost_ship
Base cost:        100 PP
Current Feudal:   67 PP
Queue command:    colony.queue_outpost_ship
```

Completion creates one civilian `StrategicFleetSpecialOutpostShip` at the producing Colony's System. The current best strategic drive is resolved at completion through the same established best-drive helper used by Colony Ship and persisted as the Fleet's installed `FTLSpeed`.

Events:

```text
colony.outpost_ship_queued
colony.outpost_ship_completed
```

### Fixed special-ship movement and Fuel range

The existing semantic transit representation remains authoritative:

```text
AtSystemID
DestinationSystemID
RemainingTurns
FTLSpeed
```

Only the fixed Colony Ship and Outpost Ship families share this path in schema 17. Ordinary/combat Fleet transit remains explicitly unsupported until the planned generic combat-Fleet movement slice.

Both fixed special ships use the established original Fuel Cell range table. No Extended Fuel Tank or generic design-special behavior is introduced here.

### Supply origins

`nearestEmpireSupplyDistanceParsecs` now considers the Systems containing:

- an owned normal Colony; or
- an owned Outpost.

The lookup intentionally does not reject an origin because the Empire is listed in that System's current blockade state, matching the direct original range predicate established in Gate 1.

A focused regression places Alpha, Beta and Gamma at deterministic distances so Gamma is outside Standard-Fuel range from Alpha before deployment. An Outpost Ship reaches Beta, deploys, and the same unchanged Standard-Fuel Empire can immediately issue a legal move to Gamma because Beta is now the nearest 4-pc supply origin. The test keeps Beta blockaded for the owner while proving that direct supply result.

### Planet-level deployment

New command:

```text
fleet.deploy_outpost {
    fleet_id,
    planet_id
}
```

Authoritative legality requires:

- the Fleet exists and belongs to the issuing Empire;
- it is a stationary Outpost Ship;
- the target Planet exists in that same System;
- the Planet has no Colony and no Outpost occupancy/reference.

No Colony-style climate/body restriction or invented treaty rule is added.

On success the resolver atomically:

1. allocates the next deterministic Outpost ID;
2. appends `core.Outpost { ID, EmpireID, PlanetID }`;
3. writes `Planet.OutpostID`;
4. consumes the Outpost Ship Fleet;
5. exposes the Outpost immediately to subsequent supply-range queries.

Events:

```text
empire.outpost_deployed
empire.outpost_ship_consumed
```

### Outpost -> Colony replacement

A shared commit helper now owns normal-Colony occupancy mutation after the existing founding initializer has prepared a valid Colony.

For a same-owner Outpost target it atomically:

- validates the reciprocal Outpost relation;
- removes the Outpost entity;
- clears `Planet.OutpostID`;
- assigns `Planet.ColonyID`;
- appends the new Colony.

The Colony Ship path emits:

```text
empire.planet_colonized
empire.outpost_converted       // only when conversion occurs
empire.colony_ship_consumed
```

The Colony Base path uses the same replacement primitive. A same-owner Outpost becomes a legal same-System Base target; a foreign Outpost remains illegal. This avoids contradictory occupancy while preserving the existing Colony Base source-Population and Building-consumption rules.

### Save/load, Observer and deterministic replay

Schema-17 Core tests prove exact JSON byte round-trip for a linked Outpost and reject dangling, missing-reciprocal and Colony+Outpost collision states.

The GameSession regression runs the normal authoritative sequence:

```text
Colony
-> queue/build Tech-109 Outpost Ship
-> Observer sees stationary installed-drive Outpost Ship
-> move under Fuel range
-> deterministic transit/arrival
-> deploy on target Planet
-> Observer sees Planet-linked Outpost and consumed ship
```

The returned Observer state is deliberately mutated in the test and a fresh Observer read proves that no mutation leaks into authority. The final state survives exact Core save/load. Running the same full session scenario twice from the same deterministic fixture produces byte-identical final Core state and identical authoritative event history.

### Gate 3 regression result

Passed before documentation-only synchronization:

```text
gofmt affected Go files
go test ./internal/core ./internal/game -count=1
go test ./internal/session -count=1
go test ./... -count=1
git diff --check
```

The full repository test run is green. Gate 4 remains intentionally open for the final fresh check, `go vet ./...`, repeated focused/full QA, staged diff verification, commits, HISTORY/status closure and marker removal.

## Gate 4 closure - complete

Final Gate 4 was executed after a fresh repository/session conflict check. No foreign active writer existed, exactly one Outpost slice marker was present and temporary Gate-3 editor files were absent.

Final verification passed:

```text
gofmt -l over every modified/untracked Go file -> no output
go test ./... -count=1 -> PASS
go vet ./... -> PASS
go test ./internal/core ./internal/game ./internal/session -run "(Outpost|ColonyShip|ColonyBase|SupplyRange|StrategicFleet)" -count=1 -> PASS
git diff --check -> PASS
git diff --cached --check -> PASS
```

Gameplay/runtime/tests/evidence were committed as:

```text
7c7284e game: add outpost ship supply expansion
```

The slice is therefore closed. The `_OPEN_OUTPOST_SHIP_SUPPLY_RANGE_2026-08-31.md` recovery marker is removed by the closing documentation commit, and Slice 03 Military Ship core / design baseline becomes the next prepared objective.