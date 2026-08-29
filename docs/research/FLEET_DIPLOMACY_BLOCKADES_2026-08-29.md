# Canonical strategic Fleet/Diplomacy blockade production - 2026-08-29

**Open marker:** `docs/slices/_OPEN_FLEET_DIPLOMACY_BLOCKADES_2026-08-29.md`

## Goal

Determine the minimum original Master of Orion II 1.31 Fleet and diplomacy/relationship semantics needed to produce system-level blockade state automatically, while preserving the existing MOOX consumer contract `StarSystem.BlockadedEmpireIDs`.

## Gate 1 - Checkup + original analysis / reverse engineering

**Status:** Gates 1-3 complete on 2026-08-29; Gate 4 QA/commit/close pending.

### Recovery state

- Starting branch: `main`.
- Starting HEAD: `41bfe92` (`docs: close population cohort slice`).
- Starting tree: clean; `main` ahead of `origin/main` by 12 commits.
- Previous gameplay slice: Race-aware Population cohorts (`7c49e90`), closed.
- Core `StateSchemaVersion`: 14.
- Economy ruleset schema: 7.

### Existing MOOX consumer boundary

MOOX already persists sorted/unique `StarSystem.BlockadedEmpireIDs`. Food logistics excludes a Colony from Empire import/export/surplus-sale pools when its owner appears in the current system's blockade list. Population transfers arriving at a system also fail with `destination_blockaded` for that target Empire.

No current Core/Game subsystem derives `BlockadedEmpireIDs` from Fleet or diplomacy state. This slice must supply that producer without changing the proven consumer meaning unless contradictory original evidence is found.

### Existing original evidence carried forward

`docs/research/BLOCKADE_FOOD_LOGISTICS_2026-08-28.md` already directly establishes from the private MOO2 1.31 `Orion2.exe` reference:

- `Colony_Is_Blockaded_` VA `0xDF8C1` consumes a per-System target-Empire bit mask at System `+0x2A`.
- `Compute_Blockades_` VA `0xE5097` rebuilds this mask from strategic presence.
- strategic fleet records use stride `0x81` in this routine;
- fleet owner/player index is read from Fleet `+0x63`;
- system/star index is read from Fleet `+0x65`;
- diplomacy relation values `4..6` are hostile (`Player_Is_Hostile_To_Player_`, VA `0xD8DE1`);
- `Compute_Blockades_` also uses fleet/status fields around `+0x64`, `+0x11` and related bytes whose semantics were deliberately left unresolved;
- the target Empire must have some qualifying presence at the system before its blockade bit is considered;
- `Next_Turn_Calc_` executes `Do_Colony_Calculations_ -> Compute_Blockades_ -> Do_Colony_Calculations_`, so newly derived blockades affect the second same-turn Economy/Food pass.

The task for this Gate 1 is therefore to resolve the remaining fleet-presence/status semantics and turn them into a minimal semantic design rather than copying original packed bytes.

## Gate 1 target questions

1. Which fleet record/state represents stationary system presence, transit, combat or other states relevant to blockade production?
2. Which qualifying presence is required from the target Empire and which from a hostile blockader?
3. Is the hostile relation check directional, symmetric, or normalized elsewhere before `Compute_Blockades_`?
4. How do multiple fleet owners at the same system combine into `system+0x2A` and the neighboring per-player blockader bytes?
5. Which strategic lifecycle point should own recomputation in MOOX?
6. What minimum persisted Fleet/Relationship state is justified for Gate 2, and which tempting fields must remain deferred?

## Evidence discipline

Direct executable observations are labeled **original-observed**. Semantic names inferred from call structure/field behavior are labeled **original-derived**. MOOX-only simplifications are labeled **modernization**. Unknown packed-field meanings remain unknown rather than being assigned convenient names.


## Gate 1 findings

### Correct LE mapping and provenance

**Evidence level: original-observed.**

The private reference remains:

```text
C:\ASH\Temp\mastori2\Orion2.exe
SHA-256 7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5
```

The bound LE module contains two objects. For the code object used here:

```text
object 1 virtual base = 0x10000
object 1 virtual size = 0x160695
```

Therefore a documented code VA such as `0xE5097` maps to LE object-1 offset `0xD5097`; research extracts must not confuse VA with object offset. Gate 1 re-extracted object 1 with `internal/moo2exe` and disassembled it with a local `objdump`; all extracts are temporary/private and are removed at Gate 1 close.

### `Compute_Blockades_` normal-Empire producer

**Evidence level: original-observed.**

The central path at `Compute_Blockades_` VA `0xE5097` is now resolved more precisely:

```text
for each strategic fleet record (stride 0x81):
    require fleet +0x64 == 0
    require fleet +0x65 < 0x48
    require fleet +0x11 == 0

    system = systems[fleet.+0x65]
    fleet_owner = fleet.+0x63

    if fleet_owner < 8:
        for each target player != fleet_owner:
            require target bit in system +0x38
            relation = player[fleet_owner].relation[target] at +0x627
            require relation in 4..6 inclusive

            system +0x2A |= 1 << target
            system +(0x2B + target) |= 1 << fleet_owner
```

Relevant direct instructions include:

```text
0xE50F6  fleet +0x64 must equal 0
0xE5100  fleet +0x65 must be < 0x48
0xE510B  fleet +0x11 must equal 0
0xE5124  owner < 8 enters the normal-player path
0xE516C  tests target bit in system +0x38
0xE517A  reads owner -> target relation at player +0x627
0xE5181..0xE5189  accepts relation values 4..6
0xE5191  ORs target player into system +0x2A
0xE5197  ORs fleet owner into system +(0x2B + target)
```

Consequences:

- blockade hostility is read **directionally from fleet owner to target player** in this routine;
- the producer does not need a target fleet to exist;
- repeated fleets from the same blockader are idempotent for blockade identity;
- several hostile fleet owners can blockade the same target simultaneously;
- original `system+0x2A` is the target-Empire mask already represented by MOOX `BlockadedEmpireIDs`;
- original `system+(0x2B+target)` records the mask of blockader player identities for that target. No current MOOX consumer needs this second mask, so persisting it would expand the slice without a demonstrated runtime requirement.

### Target presence is Colony ownership, not target-fleet presence

**Evidence level: original-observed.**

A directly adjacent system-state rebuild routine at VA `0xE5296` clears `system+0x38` and then scans the five planet slots of each system. For each planet that has a Colony, it resolves the Colony owner and executes the semantic equivalent of:

```text
system +0x38 |= 1 << colony_owner
```

The chain is directly visible as:

```text
system +0x4A + 2*planet_slot -> planet index
planet record (stride 0x11) -> Colony index
Colony record (stride 0x169) -> owner/player byte
owner bit -> system +0x38
```

`Compute_Blockades_` then tests this `+0x38` bit before setting a target blockade bit. Therefore a normal player is a blockade target in a system because it **owns a Colony in that system**. A defending/target fleet is not required by this predicate.

For MOOX this is important: target presence should be derived from existing `Planet.ColonyID` / `Colony.EmpireID`; no duplicate persisted `EmpirePresence` or `SystemPresence` state is justified.

### Fleet `+0x11`: only the ordinary/combat class can blockade

**Direct fact:** `Compute_Blockades_` requires `fleet+0x11 == 0`.

**Original-derived semantic interpretation:** the same 0x81-byte record family has three explicitly initialized non-zero special classes at `+0x11 = 1`, `4`, and `2`, while many ordinary fleet initialization paths leave/write `+0x11 = 0`. The executable also contains separate Colony Ship, Outpost Ship and Transport stack logic/functions. The exact mapping of `1/2/4` to those three civilian categories is not required to establish blockade legality and is intentionally not guessed here.

Gate 2 therefore only needs the proven semantic boundary:

```text
blockade-capable ordinary/combat strategic fleet
vs.
non-blockading civilian/special strategic stack
```

Copying the original `1/2/4` packed enum into MOOX would add unexplained legacy storage rather than useful semantics.

### Fleet `+0x64`: stationary/system presence vs non-stationary state

**Direct facts:**

- `Compute_Blockades_` requires `fleet+0x64 == 0`;
- direct fleet creation at a star/system writes `+0x64 = 0`, `+0x65 = system index`, and `+0x63 = owner`;
- other fleet lifecycle paths write non-zero values to `+0x64` while also maintaining destination/order/timing fields.

**Original-derived semantic conclusion:** value zero is the state in which the fleet is physically present at the star system and can blockade; non-zero values include movement/other non-stationary states. The full original movement-state enum is outside this slice and is not normalized here.

MOOX should model this semantically as an explicit at-system location/presence condition rather than persisting the original status byte.

### Fleet `+0x65` and owner

**Evidence level: original-observed.**

For qualifying stationary records, `+0x65` is used directly as a star-system array index and `+0x63` as the owner/player index. The `+0x65 < 0x48` check is consistent with a valid normal star-system index boundary in this path.

Embedded executable symbols/functions independently expose semantic concepts such as `Ship_Stack_Owner_`, `Ship_Stack_Location_`, `Ship_Stack_Star_Id_`, `Find_Ship_Stacks_` and `Update_Ship_Stack_`. This supports an ordinary semantic Fleet owner + location model rather than packed byte compatibility.

### Diplomacy predicate

**Evidence level: original-observed for the blockade path.**

`Compute_Blockades_` does not call a broad inferred diplomacy helper. It directly reads the fleet owner's player record and indexes the target player's relation byte at `+0x627`; only values `4..6` set a normal-Empire blockade.

`Player_Is_Hostile_To_Player_` VA `0xD8DE1` independently begins with the same directional `player[from].relation[to]` read and recognizes `4..6` as hostile, although that broader helper also contains additional special-player handling not used by the normal `Compute_Blockades_` path.

For this slice the directly justified semantic is therefore a **directed blockade-hostile relationship**. Gate 2 should not invent names for the three legacy relation values before a full diplomacy slice resolves them.

### Special/non-player fleet owners

**Evidence level: original-observed storage behavior; semantic identity deferred.**

When `fleet+0x63 >= 8`, `Compute_Blockades_` takes a special path and copies the system Colony-owner presence mask (`system+0x38`) into the target blockade mask (`system+0x2A`). Thus this class can blockade all normal Empire Colonies present at the system without the normal relation test.

This gate does **not** assign a guessed identity such as Antaran/monster to every such owner code. MOOX currently has only normal `EmpireID` ownership, so special/non-player blockade producers should remain deferred until their canonical owner/entity model exists.

### Timing relative to Settlers and Food

**Evidence level: original-observed, confirmed by the prior Population-transfer slice.**

Original `Next_Turn_Calc_` runs:

```text
All_Colony_Calculations_
Compute_Blockades_
Move_Settlers_
All_Colony_Calculations_
```

MOOX currently performs an initial Economy/Food materialization, later resolves Population transfers, then recalculates the next-state Colony/Food snapshot. The original-compatible insertion point for the future blockade producer is therefore **immediately before `advancePopulationTransfers`**:

```text
initial Colony/Food snapshot
...
Research
Population
Construction
recompute blockades        <- new producer
advance Population transfers
final Colony/Food snapshot
```

That gives arriving Settlers the newly computed blockade state and gives the final Food snapshot the same state, while leaving the first snapshot on the previous materialized blockade exactly as the original two-pass sequence does.

### Secondary original blockader mask

After producing `system+0x2A` and the per-target blockader masks at `+0x2B+target`, the original routine has further logic that iterates those masks and performs additional side effects/random handling. Those side effects are not required for the already-proven Food/Settler blockade consumer contract and are not normalized in this slice. Gate 2 should not add them speculatively.

## Gate 1 design implications

The minimal authoritative inputs justified by the evidence are:

1. **Strategic Fleet identity**: stable Fleet ID and normal Empire owner.
2. **Blockade capability/role**: semantic combat/blockade-capable vs civilian/special; only the former participates.
3. **At-system location**: an explicit current StarSystem reference; fleets not currently at a system do not blockade.
4. **Directed Empire relationship**: a semantic relation sufficient to answer whether `from` is blockade-hostile to `to`.
5. **Target presence**: derived from existing Colony ownership in the StarSystem; not persisted again.
6. **Materialized blockade output**: retain existing sorted/unique `StarSystem.BlockadedEmpireIDs` as the deterministic derived state consumed by Food and Population transfer.

The evidence does **not** justify adding tactical ship composition, weapons, officers, movement ETA/path, diplomacy proposals/treaties, blockader masks, special/non-player fleet ownership or AI orders to this slice.

## Gate 2 implementation decision

Decision: accepted on 2026-08-29 without scope changes. The following schema/rule shape is the binding Gate 3 implementation contract.

### Core schema 14 -> 15

Add minimal strategic state:

```go
type StrategicFleetRole string

const (
    StrategicFleetRoleCombat   StrategicFleetRole = "combat"
    StrategicFleetRoleCivilian StrategicFleetRole = "civilian"
)

type StrategicFleet struct {
    ID         ID                `json:"id"`
    EmpireID   ID                `json:"empire_id"`
    Role       StrategicFleetRole `json:"role"`
    AtSystemID ID                `json:"at_system_id,omitempty"`
}

type DiplomaticStance string

const (
    DiplomaticStanceNeutral DiplomaticStance = "neutral"
    DiplomaticStanceHostile DiplomaticStance = "hostile"
)

type DiplomaticRelation struct {
    FromEmpireID ID                `json:"from_empire_id"`
    ToEmpireID   ID                `json:"to_empire_id"`
    Stance       DiplomaticStance `json:"stance"`
}
```

and persisted collections on `GameState`, with deterministic ordering and validation.

`AtSystemID == 0` means the Fleet is not currently stationary at a StarSystem and therefore cannot blockade. This slice deliberately does not invent transit destination/ETA state; the later strategic-movement slice can extend the same Fleet schema when its own evidence is established.

The relationship collection is directed. Missing entries are treated as non-hostile. The semantic `hostile` state represents the directly proven blockade predicate rather than claiming names for legacy relation values `4..6`; a later diplomacy slice may extend the enum without changing blockade callers.

### Deterministic blockade producer

Add a pure authoritative phase, conceptually:

```go
recomputeSystemBlockades(state *core.GameState) error
```

For every StarSystem:

1. derive target Empire IDs from Colonies on planets in that System;
2. scan stable-ID-ordered strategic Fleets with `Role == combat` and `AtSystemID == system.ID`;
3. for each target owner other than the Fleet owner, test directed `relation(fleet.EmpireID, target).Stance == hostile`;
4. collect target Empire IDs into a sorted/unique blockade list;
5. replace `system.BlockadedEmpireIDs` completely, including clearing stale entries when no qualifying Fleet remains.

Multiple Fleets/owners naturally collapse through set semantics. No combat-strength threshold is visible in `Compute_Blockades_`; one qualifying normal combat Fleet presence is sufficient to set the bit.

### Strategic lifecycle

Invoke blockade recomputation after Construction and before Population-transfer resolution, immediately before the existing final Colony/Food recalculation. No player command is introduced by this slice.

### Session / Observer

Strategic Fleets, directed relations and the resulting materialized blockade state are authoritative Core state and therefore must round-trip through save/load and GameSession clone/Observer projections. No UI-specific filtered model should own blockade legality.

### Gate 3 regression shape

Deterministic tests should cover at least:

- schema-15 Fleet/relation roundtrip and strict validation/order;
- combat Fleet at hostile Colony system -> target Empire blockaded;
- civilian Fleet -> no blockade;
- Fleet with `AtSystemID == 0` -> no blockade;
- hostile Fleet at a system without target Colony -> no target blockade;
- own Colony/own Fleet -> no self-blockade;
- directed relation `A -> B hostile`, `B -> A neutral` -> only B can be blockaded by A's Fleet;
- multiple hostile owners -> one sorted/unique target blockade result;
- removing/moving the last qualifying Fleet clears stale `BlockadedEmpireIDs`;
- recomputation occurs before Settler arrival, so current blockade can cause `destination_blockaded`;
- final Food snapshot uses the same newly recomputed blockade;
- GameSession/Observer/save-load preserve Fleet, relation and derived blockade state.

## Gate 1 deliberate deferrals

- exact `+0x11` mapping among Colony Ship / Outpost Ship / Transport values `1/2/4`;
- complete fleet movement/path/ETA semantics and the original `+0x64` enum;
- special/non-player owner codes `>=8`;
- original per-target blockader masks beyond the existing target blockade consumer;
- full diplomacy relation-state names, treaties, negotiations and AI diplomacy;
- tactical combat/ship composition/strength;
- secondary random/side-effect logic later in `Compute_Blockades_`.


## Gate 3 implementation result

**Status:** complete on 2026-08-29. Gate 4 QA/commit/close remains open.

### Core schema 15

`internal/core/state.go` now advances `StateSchemaVersion` from 14 to 15 and persists:

```go
StrategicFleets     []StrategicFleet
DiplomaticRelations []DiplomaticRelation
```

`internal/core/strategic.go` implements the accepted semantic contract:

```go
StrategicFleet { ID, EmpireID, Role, AtSystemID }
DiplomaticRelation { FromEmpireID, ToEmpireID, Stance }
```

Fleet roles are `combat|civilian`; diplomatic stances are `neutral|hostile`. Missing directed relations resolve to neutral. Validation requires:

- strategic Fleet IDs strictly ascending and globally unique;
- known Fleet owners;
- only accepted Fleet roles;
- `AtSystemID == 0` or a known StarSystem;
- non-zero, non-self directed diplomatic pairs;
- known source/target Empires;
- only accepted diplomatic stances;
- diplomatic relations strictly ascending by `(from_empire_id,to_empire_id)`.

The existing strict Core serializer therefore round-trips Fleet/relation state automatically. Schema-14 exact assertions in earlier transformation/cohort tests were advanced to schema 15.

### Authoritative blockade producer

`internal/game/blockade.go` adds the pure deterministic producer:

```go
recomputeSystemBlockades(state *core.GameState) error
```

For every StarSystem it:

1. derives target Empire presence from `Planet.ColonyID -> Colony.EmpireID`;
2. scans strategic Fleets and keeps only `role=combat` with `AtSystemID == system.ID`;
3. skips self-targets;
4. tests directed `DiplomaticStanceBetween(fleet.EmpireID, targetEmpireID) == hostile`;
5. collects targets with set semantics;
6. writes sorted/unique `StarSystem.BlockadedEmpireIDs`;
7. fully clears stale blockade state when no qualifying producer remains.

No target Fleet, Fleet-strength threshold, tactical ship state or duplicate system-presence table was introduced.

### Turn integration

`internal/game/economy_resolver.go` now recomputes blockade state after Construction and immediately before `advancePopulationTransfers`, preserving the direct original sequence:

```text
first Colony/Food snapshot
Research
Population
Construction
recompute blockades
Population-transfer arrivals
final Colony/Food snapshot
```

This means the first snapshot can still consume the previously materialized blockade state, while Settler/Population arrivals and the final Food snapshot consume the newly derived Fleet/Diplomacy state.

### Session / Observer

No duplicate session schema was needed. `GameSession` already owns and clones the complete `core.GameState` through strict Core serialization. `internal/session/strategic_blockade_test.go` now proves that Observer views preserve schema-15 Fleet, directed relation and materialized blockade state and that Observer mutations cannot leak back into authoritative session state.

### Deterministic regressions added

Core regressions cover:

- schema-15 Fleet/relation save-load roundtrip;
- missing reverse relation -> neutral;
- Fleet ordering, owner, role and system-reference validation;
- diplomatic source/target/self/stance/order/duplicate validation.

Game regressions cover:

- Colony target presence without a defending Fleet;
- combat Fleet + directed hostility -> blockade;
- civilian Fleet -> no blockade;
- non-stationary Fleet -> no blockade;
- own-system/self presence -> no self-blockade;
- directed hostility does not imply reverse hostility;
- multiple hostile blockaders collapse idempotently;
- multiple target Empires are sorted deterministically;
- stale blockade materialization is cleared;
- first Food pass keeps previous materialized state while final Food consumes the recomputed blockade;
- Population-transfer arrival sees the freshly recomputed blockade and resolves as `destination_blockaded`.

Development verification completed during Gate 3:

```text
go test ./internal/core     PASS
go test ./internal/game     PASS
go test ./internal/session  PASS
go test ./...               PASS
git diff --check           PASS
```

Formal Gate 4 still owns the final formatting/test/vet/diff QA, commit, permanent-history update and deletion of the `_OPEN_` marker.
