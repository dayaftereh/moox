# Strategic hostile encounters and BattleSession handoff - Slice 06 evidence

Date: 2026-08-31

Status: **closed; Gates 1-4 complete**

Planned slice: `docs/slices/PLANNED_06_STRATEGIC_HOSTILE_ENCOUNTERS_BATTLE_HANDOFF.md`

Recovery marker: **removed at Gate 4 closure**.

## Objective

Determine the original MOO2 strategic encounter trigger/order and the narrow authoritative MOOX handoff needed to turn hostile strategic Fleet co-location into deterministic `BattleSession` children, while keeping tactical movement/fire/damage in Slice 07.

## Repository baseline at Gate-1 start

- starting HEAD: `936fb16` (`docs: close command point maintenance slice`);
- branch `main` was 30 commits ahead of `origin/main`;
- working tree clean;
- zero pre-existing `_OPEN_*.md` markers;
- Core schema 20 / economy ruleset schema 8;
- Slice 05 closed; Slice 06 was the next prepared objective.

## Current MOOX infrastructure baseline

### Existing strategic resolver

`internal/game/resolver.go` already has:

```go
type Encounter struct {
    Participants []protocol.SeatID
}

type Resolution struct {
    State      *core.GameState
    Events     []DomainEvent
    Encounters []Encounter
}
```

but the current `EconomyResolver.Resolve` returns no encounters.

Its currently relevant turn order after strategic commands is:

```text
recalculate current Colony economy
Food logistics
Treasury
Research
Population
Strategic Fleet transit
Construction
recompute Blockades
Population transfers
next-state Colony/Food materialization
return Resolution
```

Therefore the narrow insertion point that matches the directly proven original ordering is naturally **after Fleet transit + Construction and before blockade/population-transfer processing**.

### Existing GameSession / BattleSession boundary

`GameSession.ResolveStrategic` currently asks one resolver call for a completely resolved strategic State, validates it, prepares all returned encounters atomically and only then commits State/events plus child battles.

Existing battle lifecycle strengths:

- stable sequential Battle IDs;
- deterministic `battle.DeriveSeed(gameSeed, strategicTurn, battleID)`;
- battle children are prepared before authoritative strategic State is committed;
- `PhaseEncounters` blocks the normal turn flow while battles are active;
- independent child battles can complete in arbitrary wall-clock order;
- authoritative `battle_completed` events are emitted only after all active children are complete and in stored Battle-ID order;
- then the session enters `PhasePostResolution`.

Current battle model is intentionally minimal:

```go
type battle.Spec struct {
    ID            uint64
    GameID        string
    StrategicTurn uint64
    Participants  []protocol.SeatID
    Seed          uint64
}

type battle.Result struct {
    WinnerSeats []protocol.SeatID
    Outcome     string
}
```

This is enough for a lifecycle demo but not enough to resume the strategic MOO2 order correctly after an actual hostile encounter.

`NewGameSession` enforces at most one Seat per Empire, but does not require every Core Empire to have a Seat. The current `battle.Session` participant model is Seat-based, so Slice 06 should not silently invent an unseated/NPC battle participant model.

There is currently no `GameSession` marshal/unmarshal or active-session persistence API. Core save/load is not an active-BattleSession recovery format.

## Original-reference provenance

Private local reference executable:

`C:\ASH\Temp\mastori2\Orion2.exe`

Original HELP text:

`reference/original/text/help/block_0000.ascii.txt`

The executable is evidence-only and must not be distributed with MOOX.

## Important Watcom debug-symbol address correction

Gate 1 found and verified an address-decoding mistake that can affect older research notes.

For the Watcom symbol records used here, the current function's object offset is the **four bytes immediately before the current symbol record header**, followed by the two-byte object number. The four bytes immediately following the symbol name belong to the **next** symbol record.

Fresh disassembly proves this directly:

- correct `Next_Turn_Calc_`: object offset `0x36B3`, VA `0x136B3`; it contains the full next-turn call chain;
- `0x3822`, previously easy to associate with the preceding name when reading the trailing value, is actually `Do_Autosave_` and contains autosave cadence code;
- correct `Next_Turn_`: `0x3870`, VA `0x13870`, and it calls both `Do_Autosave_` and `Next_Turn_Calc_`;
- correct `Move_All_Ships_Toward_Stars_`: object offset `0xEFEEA`, VA `0xFFEEA`;
- the following `0xF0010` routine is `Create_Nonplayer_Ship_`, not movement.

Historical documents are not rewritten in Gate 1, but future executable work must use the corrected record interpretation.

Correct key COMBAT.C / movement symbols used by this evidence:

| Function | object offset | VA |
| --- | ---: | ---: |
| `Get_Combat_Ships_` | `0xD6A0C` | `0xE6A0C` |
| `Update_Player_Needs_To_Attack_` | `0xD6AA2` | `0xE6AA2` |
| `Kill_Noncombat_Ships_` | `0xD6B44` | `0xE6B44` |
| `Closest_Other_Colony_Star_` | `0xD6BCD` | `0xE6BCD` |
| `Determine_Retreat_Ships_` | `0xD6CAA` | `0xE6CAA` |
| `Process_Retreating_Ships_` | `0xD6E52` | `0xE6E52` |
| `Human_Confirms_Attack_` | `0xD7B79` | `0xE7B79` |
| `Get_Local_Human_Target_` | `0xD7CDB` | `0xE7CDB` |
| `Get_AI_Target_` | `0xD7DCA` | `0xE7DCA` |
| `Get_NPC_Target_` | `0xD8029` | `0xE8029` |
| `Pick_Attacker_At_Star_` | `0xD8194` | `0xE8194` |
| `Player_May_Attack_Player_` | `0xD8231` | `0xE8231` |
| `Collect_Targets_For_Combat_` | `0xD82B4` | `0xE82B4` |
| `Some_Star_Player_Is_Busy_` | `0xD841A` | `0xE841A` |
| `Pick_Star_And_Attacker_` | `0xD84A5` | `0xE84A5` |
| `Do_Attacker_Beat_Colony_Stuff_` | `0xD87D2` | `0xE87D2` |
| `Reset_Combat_Ships_At_Current_Star_Has_Player_` | `0xD9358` | `0xE9358` |
| `Do_1_Combat_` | `0xD938C` | `0xE938C` |
| `Search_For_Battles_` | `0xD9D62` | `0xE9D62` |
| `Client_Combat_Loop_` | `0xDA2F1` | `0xEA2F1` |
| `Make_Ship_Arrive_At_Star_` | `0xEFDDA` | `0xFFDDA` |
| `Move_All_Ships_Toward_Stars_` | `0xEFEEA` | `0xFFEEA` |

## Gate-1 original findings

### 1. Exact turn-order boundary: movement and production before combat; blockade/transfers after combat

Correctly mapped `Next_Turn_Calc_` at `0x136B3` directly calls, in relevant order:

```text
Make_Scrap_Ships_Dead_
Initialize_Reports_
Process_Trade_And_Research_Agreements_
Move_All_Ships_Toward_Stars_
Resolve_Spies_
Apply_All_Player_Changes_
Apply_All_Colony_Changes_
Do_Surrenders_
Deallocate_AI_Data_
Search_For_Battles_
Determine_Event_
Do_All_Ships_XP_Check_
Event_Twiddle_
Lose_Out_Of_Range_Ships_
All_Colony_Calculations_
Compute_Blockades_
Move_Settlers_
All_Colony_Calculations_
Allocate_New_Ship_Slots_
...
Allocate_New_Colony_Slots_
...
```

Consequences:

- normal ship movement/arrival is complete before encounter search;
- current-turn Colony changes/Production are complete before encounter search, so a newly completed military Ship can participate in same-turn combat;
- combat is resolved before Blockade recomputation;
- combat is resolved before original Settler/population-transfer movement;
- later Colony/Food calculations see the post-combat strategic position.

Current MOOX can preserve the important supported boundary by inserting encounter derivation **after `advanceStrategicFleetTransit` and `advanceConstruction`, before `recomputeSystemBlockades` and population-transfer advancement**. Slice-05 Treasury remains earlier and is unaffected.

### 2. Arrival is not a special-only trigger; every stationary hostile co-location can be searched

`Move_All_Ships_Toward_Stars_` scans normal moving Ship statuses and calls `Make_Ship_Arrive_At_Star_` when appropriate.

`Search_For_Battles_` then scans the complete Ship table. Normal combat candidates are stationary Ships with:

```text
ship+0x64 == 0
ship+0x65 < 0x48   // real star/system index
```

There is no "arrived this turn" requirement in the combat scan.

Therefore:

- a Ship that arrived during the preceding movement phase is eligible once its arrival materializes stationary state;
- hostile Ships that were already stationary together at the start of the turn are also eligible;
- in-transit Ships are not tactical participants until arrival.

This maps cleanly to MOOX `StrategicFleet.AtSystemID != 0` plus no active destination/ETA.

### 3. Original sides are Empire aggregates, not Fleet-container pairs

`Get_Combat_Ships_` scans all Ship records and appends every stationary Ship at the selected star whose owner matches either selected player.

Original MOO2 has no MOOX-style persistent strategic Fleet container boundary for this selection. All co-located Ships of one player become one combat side.

MOOX implication:

- if one Empire has several combat `StrategicFleet` containers at the same System, an encounter side must aggregate all of those Fleet IDs and all of their concrete Ship IDs;
- do **not** create one BattleSession for every Fleet-pair cross product;
- Fleet containers remain useful reconciliation identities after battle but are not separate original tactical sides.

### 4. Three-or-more-party systems are resolved as sequential pairwise combats

`Search_For_Battles_` does not construct one all-party battle.

It:

1. builds a list of candidate battle stars;
2. `Pick_Star_And_Attacker_` chooses a star and one attacker;
3. `Collect_Targets_For_Combat_` builds attackable colony/player target lists for that attacker;
4. local-human, AI or NPC target selection chooses one target;
5. `Do_1_Combat_` resolves that attacker/target pair;
6. `Update_Player_Needs_To_Attack_` removes sides that no longer have relevant stationary presence;
7. the search loops again.

Thus a system containing A, B and C can produce multiple **pairwise, sequential** combats. The state after an earlier combat changes eligibility for the next pair.

MOOX implication: overlapping BattleSessions for the **same System** must not be launched in parallel. At most one attacker/defender pair per System can be active in an encounter wave; after its result is applied, that System must be re-evaluated.

Independent Systems can still use the existing parallel child-BattleSession machinery safely if result application is committed in a stable order.

### 5. Original battle-system and attacker order is RNG-driven

After `_battle_stars` is built, `Search_For_Battles_` calls `Shuffle_Char_` on the battle-star list.

`Pick_Attacker_At_Star_` uses repeated `Random_` calls/reservoir-style selection among eligible attacker bits.

For local-human attackers, `Defense_Colony_Selection_Popup_` lets the player choose a target and `Human_Confirms_Attack_` confirms it. AI/NPC attackers use separate target selectors.

Therefore original MOO2 does **not** have a simple star-ID / player-ID deterministic order. Its order is deterministic only through the original game RNG plus human target choices.

Gate-1 recommendation for Slice 06: do not copy only the RNG fragment while omitting original human target choice and sneak-attack diplomacy. With current MOOX `hostile` as already-authorized attack intent, prefer a deliberate canonical order (System ID, directed attacker Empire ID, defender Empire ID) for the handoff slice and record this as a modernization. A later diplomacy/attack-choice slice can restore original random/human selection semantics as a complete feature rather than pseudo-fidelity.

### 6. Original attack eligibility is broader than current MOOX hostility

`Player_May_Attack_Player_` is directional and consults original player/diplomacy state; it contains several diplomacy/status gates and even a `Random_` branch.

`Russ_Combat_` contains `Show_Sneak_Attack_Message_` / `Player_Attacked_Popup_` paths, proving combat can also be coupled to original sneak-attack/war-state transitions.

Current MOOX deliberately has only:

```text
neutral
hostile
```

and Blockade logic already treats `fromEmpire -> targetEmpire == hostile` as attack-authorized strategic hostility.

Gate-1 recommendation: Slice 06 should auto-produce encounters only from **already-hostile directed MOOX relations**. Neutral sneak attacks, declarations of war, attack confirmation and richer original diplomacy stay deferred. This is narrower than original but evidence-safe.

If only A -> B is hostile, A is the authorized attacker. If both directions are hostile, canonical encounter ordering chooses the first pair; participant sorting must not erase explicit attacker/defender identity.

### 7. Civilian Colony/Transport/Outpost ships do not tactically fight

Original HELP directly states:

- Colony Ship: will not engage in combat and is destroyed when attacked if not escorted by a military ship;
- troop Transport: does not engage in combat and is destroyed when attacked if not escorted by military ships;
- Outpost Ship: unarmed and destroyed if not escorted by military ships.

The executable is consistent with those records:

- `Search_For_Battles_` tracks Ship special-kind flags separately from ordinary kind 0 combat presence;
- ordinary kind 0 contributes the normal player combat/attacker bit;
- special kinds are tracked as noncombat presence;
- `Do_1_Combat_` calls `Kill_Noncombat_Ships_` when a target has no combat defense and no colony-defense target;
- `Kill_Noncombat_Ships_` iterates the selected target Ship-ID list and calls `Kill_Ship_` for every entry, then clears that player's combat-presence flag.

If a side loses a real combat, `Determine_Retreat_Ships_` marks all remaining stationary Ships of that player at that star for retreat, so surviving civilian ships leave with the defeated side rather than acting as separate tactical units.

MOOX currently represents Colony/Outpost Ships as fixed-special `StrategicFleet`s rather than concrete military `Ship`s. Recommended handoff representation:

- combat Fleet IDs + concrete Ship IDs are tactical side composition;
- co-located Colony/Outpost special Fleet IDs are attached separately as civilian assets;
- civilian assets do not make a side combat-capable or an attacker;
- a hostile combat side versus an unescorted civilian-only side with no modeled colony defense can be resolved as an immediate strategic civilian overrun, not a fake tactical battle;
- civilians accompanying a combat side follow that side's post-battle retreat/destruction disposition.

Original troop Transport remains deferred because MOOX Population transfers are not that Ship type.

### 8. Losing surviving ships retreat before Blockade is computed

`Do_1_Combat_` receives the combat result from `Russ_Combat_` and branches by attacker/defender result.

`Determine_Retreat_Ships_` directly scans remaining stationary Ships of the losing player at the battle star and normally writes:

```text
ship+0x64 = 9   // temporary retreat state
```

for every remaining Ship of that player at that star. Under the original Hyperspace Flux condition it instead kills those Ships.

After the local combat-search loop finishes, `Search_For_Battles_` calls `Process_Retreating_Ships_` before returning to `Next_Turn_Calc_`.

`Process_Retreating_Ships_`:

- groups retreating Ships by owner/current star;
- calls `Closest_Other_Colony_Star_`;
- that function searches the player's other Colony stars and chooses the smallest squared star-coordinate distance;
- tries to move the group using the normal ship movement helpers;
- if there is no valid other Colony destination or the retreat movement cannot be established, the relevant Ships are destroyed;
- successful retreats are converted back into normal movement state.

Only after this does `Next_Turn_Calc_` later call `Compute_Blockades_`.

This is the decisive reason the current `battle.Result{WinnerSeats, Outcome}` is insufficient for a correct strategic resume: the strategic state must know which side lost and must establish the losing survivors' post-battle location before Blockade/population-transfer processing continues.

### 9. Minimum battle result vocabulary for Slice 06

Original tactical/strategic combat mutates damage/death internally before the strategic retreat step. Current MOOX `Ship` intentionally has no tactical HP/damage fields yet.

Therefore Gate 1 recommends a deliberately small handoff contract:

- a normal two-side strategic BattleSession must return **exactly one winning side**;
- optional `DestroyedShipIDs` may be accepted as an externally supplied subset of the encounter's concrete Ship snapshot, so Slice 07 can grow into the contract without changing strategic identity again;
- after removing any explicitly destroyed Ships, the strategic layer, not tactical code, applies the directly evidenced minimum loser disposition: all surviving losing-side Fleets/civilians retreat from the System toward the closest other owned Colony System using existing strategic transit machinery; if no valid retreat can be established, they are removed/destroyed;
- do not model tactical damage, retreat choice, individual tactical retreat timing, officer effects or Hyperspace Flux in this slice beyond the minimum supported strategic disposition;
- civilian-only overrun can remove the unescorted civilian special Fleet(s) immediately without starting a BattleSession.

With an external/stub battle result that supplies no tactical casualties, the temporary Slice-06 behavior is therefore "winner remains; all losing combat Ships survive tactically but retreat strategically". This is explicit scaffolding, not a claim that tactical combat has been implemented.

Gate 2 must accept/revise the exact result field shape before implementation.

### 10. Colony / orbital defense is an original combat target, but current tactical representation is incomplete

`Collect_Targets_For_Combat_` first scans colony/planet targets and applies `Player_May_Attack_Player_`; it separately collects player/ship targets.

`Do_1_Combat_` carries a colony/planet target through `Russ_Combat_`, and after an attacker victory can call `Do_Attacker_Beat_Colony_Stuff_`.

Original HELP also documents Star Base/Battlestation/Star Fortress combat effects.

Current MOOX has station buildings but no tactical station/planet-defense unit model. Slice 06 must not invent tactical station stats.

Recommended scope:

- automatic BattleSession creation requires a combat-capable Fleet side on both Empires;
- if the battle occurs over a defender Colony, carry optional Colony ID / station-building context in the EncounterSpec for future Slice 07 use, but do not synthesize a station Ship;
- a hostile Fleet facing only a Colony/station with no defending combat Fleet remains outside the first automatic BattleSession producer; current blockade state still represents strategic pressure;
- planetary bombardment/invasion remains explicitly deferred.

### 11. Colonization timing: newly arriving Colony/Outpost ships are exposed to combat before new colonization opportunity materialization

Correct `Next_Turn_Calc_` places `Search_For_Battles_` well before `Allocate_New_Colony_Slots_`.

`Allocate_New_Colony_Slots_` calls `Find_Colonize_Players_`.

Fresh `Find_Colonize_Players_` disassembly shows it scans stationary Ships (`ship+0x64 == 0`) and explicitly tests special kinds:

```text
ship+0x11 == 1  // Colony Ship
ship+0x11 == 4  // Outpost Ship
```

before `Star_N_Colonizable_Planets_For_Player_` / `Star_N_Outpostable_Planets_For_Player_`.

Thus a **newly arriving** Colony/Outpost Ship reaches stationary state, passes through combat search, and only later contributes to the normal human colonization/outpost opportunity discovery.

`All_AI_Colonize_` is separately called earlier in the turn before movement, so original AI already-stationary colonization has a different ordering path. Slice 06 does not need to duplicate that AI asymmetry.

Current MOOX commands submitted from the planning state cannot anticipate a same-turn arrival, so inserting encounters after transit naturally preserves the important new-arrival boundary. Pre-existing stationary Colony/Outpost commands remain governed by their existing legal command timing.

### 12. Original battle order is synchronous inside the turn; current MOOX commits encounters too late

In original single-player `Search_For_Battles_`, each `Do_1_Combat_` returns before the function continues to the next pair, retreats are processed, and only then does `Next_Turn_Calc_` continue into Blockades/Settlers/etc.

Current `GameSession.ResolveStrategic` instead asks its resolver for a completely finished strategic State and then enters `PhaseEncounters`.

That architecture is too late for real encounters: if Blockades and Population transfers have already been resolved, battle results cannot correctly affect them without rollback/recomputation.

Gate-2 architectural requirement:

```text
pre-encounter strategic resolution
    -> Fleet transit
    -> Construction
    -> encounter detection / BattleSession wave
PAUSE while battles active
apply results in stable Battle-ID order
re-evaluate same-System pairwise encounters as necessary
repeat encounter waves until none remain
post-encounter strategic resolution
    -> Blockades
    -> Population transfers
    -> next-state Colony/Food materialization
    -> PhasePostResolution
```

The authoritative Core State exposed during `PhaseEncounters` is the validated **pre-post-resolution** state. No normal strategic commands are accepted while it is frozen at this boundary.

### 13. Parallelism: independent Systems can run together; one System must remain sequential

Original processing is globally sequential, but current MOOX already intentionally supports parallel child BattleSessions.

Evidence permits a safe modernization:

- launch at most one attacker/defender pair per System in one encounter wave;
- Battles from different Systems are independent under Slice-06's already-hostile/no-diplomacy-change scope and may run in parallel;
- apply completed results in stable Battle-ID order, never wall-clock completion order;
- after the wave, re-run encounter detection on the updated State;
- if a three-party System still has a hostile eligible pair, create its next BattleSession in the next wave;
- only when no encounter remains may post-encounter Blockade/transfer processing run.

This preserves the original same-System dependency while retaining MOOX's deterministic parallel architecture.

### 14. Participant mapping must preserve direction as well as Seat identity

Original combat distinguishes attacker and target. Existing `battle.Session` sorts its `Participants`, so `Participants []SeatID` cannot carry attack direction.

Recommended Encounter/Battle handoff needs, at minimum:

```text
SystemID
AttackerEmpireID + AttackerSeatID
DefenderEmpireID + DefenderSeatID
AttackerCombatFleetIDs + AttackerShipIDs
DefenderCombatFleetIDs + DefenderShipIDs
AttackerCivilianFleetIDs / DefenderCivilianFleetIDs
optional defended ColonyID/context
```

Battle Session participants remain the two Seats, but attacker/defender identity is explicit metadata.

Regular AI Empires already use Seat controllers and fit this model. Unseated/NPC/monster participants are explicitly deferred rather than represented by invented Seat 0 semantics.

### 15. Save/recovery boundary

There is no current durable `GameSession` persistence format. Core save/load alone cannot reconstruct:

- active Battle IDs;
- BattleSession phases/results;
- a pending pre/post strategic continuation;
- which encounter wave is active.

Slice 06 should therefore not pretend active-battle disk saves exist.

Minimum accepted recovery guarantee for this slice should be deterministic replay:

- same pre-turn Core State + same submitted commands + same external battle results reproduce the same encounter specs, Battle IDs/seeds, result-application order, post-resolution State and events;
- failed/invalid battle-result application must leave the authoritative session at the same encounter boundary without partial strategic mutation;
- Observer exposes enough encounter/battle metadata to reconstruct what is pending for clients.

Durable mid-BattleSession process-restart serialization is a separate future session-persistence concern unless Gate 2 explicitly expands scope.

## Gate-1 proposed Gate-2 implementation contract

This is a proposal, **not yet accepted or implemented**.

### A. No Core schema bump required by encounter metadata alone

Keep Core schema 20 if the implementation uses existing Ship/Fleet transit fields for post-battle retreat and does not persist a new Core retreat/encounter state.

Battle/Encounter/session structs can evolve independently of Core State schema.

### B. Split strategic resolution at the original combat boundary

Refactor the resolver/session interaction into:

1. pre-encounter resolution through Fleet transit + Construction;
2. derive one encounter pair per contested System;
3. freeze authoritative State in `PhaseEncounters`;
4. after every active wave is complete, atomically apply battle results in Battle-ID order;
5. re-derive encounter pairs from the updated State;
6. if none remain, run Blockades + Population transfers + next-state materialization and enter PostResolution.

No post-encounter phase may run before all encounter waves are exhausted.

### C. Encounter side identity aggregates same-Empire Fleets

Per System/Empire side, collect all stationary combat Fleet IDs and their concrete Ship IDs. Keep fixed Colony/Outpost special Fleet IDs as separate civilian context.

A Fleet container never creates a separate tactical side merely because its Fleet ID differs.

### D. Only current directed `hostile` relationships auto-attack

Do not implement neutral sneak attacks, war declarations or human attack confirmation in Slice 06.

Use directed hostility to determine attacker authorization. When multiple already-hostile pairs exist, use a canonical stable System/Empire pair order as a deliberate temporary modernization rather than partially copying original RNG/human-selection behavior.

### E. Pairwise encounter waves

At most one BattleSession per System per wave. Independent Systems may be parallel.

After stable result application, re-evaluate the System before creating a second pair there.

### F. Battle Spec grows authoritative strategic context

Extend Battle/Encounter spec to preserve System, directed Empire/Seat sides and stable Fleet/Ship/civilian IDs. Core State is frozen during the active wave, so IDs are authoritative references; tactical Slice 07 may build its own immutable tactical snapshot from those IDs later.

### G. Minimum result semantics

Require exactly one winning side for the supported two-side BattleSession.

Recommended future-proof minimum:

- existing winner/outcome;
- optional `DestroyedShipIDs` subset of the encounter snapshot;
- strategic resolver removes those IDs, then retreats all remaining losing-side Fleet/civilian assets to the closest other owned Colony System using existing strategic movement state;
- if no valid retreat exists, remove/destroy those losing assets as the original does;
- tactical HP/damage/fire/retreat decisions remain external/deferred.

Civilian-only overrun with no colony defense is an immediate strategic outcome and does not create a fake tactical BattleSession.

### H. Colony-only defense remains deferred

Carry Colony context when a normal Fleet-vs-Fleet battle occurs at a Colony, but do not generate a station/colony-only tactical battle until there is an authoritative tactical station/planet-defense representation.

### I. Observer/replay

Battle-created/completed events include the enhanced encounter context. Parallel completions continue to commit in Battle-ID order.

Identical runs must reproduce encounter-wave order, Battle IDs/seeds, post-battle retreat/removal effects, Blockades/transfers and final State/events.

No active-session disk persistence is claimed in Slice 06.

## Gate-2 accepted implementation contract

Gate 2 was explicitly approved by the user. The following contract is authoritative for Gate 3 and supersedes the earlier "proposed Gate-2" wording where this section is more specific.

### 1. Schema/ruleset boundary

- Core `StateSchemaVersion` remains **20**.
- Economy ruleset schema remains **8**.
- No new persisted Core encounter/retreat object is introduced in Slice 06.
- Post-battle movement/removal uses the existing `Ship` / `StrategicFleet` identity and transit fields.
- Battle/Session/Resolver structs may evolve without a Core schema bump because active encounter lifecycle is session state, not persisted Core state.

### 2. Resolver continuation contract

Keep the existing one-shot `game.Resolver` interface for compatibility with resolvers that return a complete strategic resolution without encounters.

Add a staged extension for resolvers that return real encounters, conceptually:

```go
type EncounterResolver interface {
    Resolver
    ResumeAfterEncounters(
        ctx ResolveContext,
        state *core.GameState,
        outcomes []EncounterOutcome,
    ) (Resolution, error)
}
```

`EconomyResolver` implements this extension.

Accepted behavior:

- `Resolve(...)` executes commands and the normal turn only through Fleet transit + Construction.
- It performs deterministic civilian-overrun handling and encounter derivation at that boundary.
- If no BattleSession encounter remains, it immediately executes the post-encounter stages and returns a normal complete `Resolution`.
- If BattleSessions are required, it returns the frozen pre-post-resolution State/events plus the current encounter wave.
- A resolver that returns non-empty `Resolution.Encounters` but does not implement the staged extension is rejected **before authoritative session mutation**.
- `GameSession` may retain the staged resolver/continuation **in memory only** while `PhaseEncounters` is active. Durable serialization of that reference is explicitly out of scope.
- `ResumeAfterEncounters(...)` applies one completed wave on a cloned State, re-derives same-System follow-up encounters and either returns the next wave or, when none remain, executes Blockades + Population transfers + final Colony/Food materialization.

### 3. Authoritative phase split

Accepted supported order:

```text
commands / normal pre-existing turn work
Treasury
Research
Population
Strategic Fleet transit / arrivals
Construction
civilian-only hostile overruns
encounter-wave derivation
PAUSE: PhaseEncounters
  -> BattleSession wave
  -> stable result application
  -> same-System re-evaluation
  -> repeat waves while needed
post-encounter continuation
Blockades
Population transfers
next-state Colony/Food materialization
PhasePostResolution
```

No Blockade, Population-transfer or final next-state economy materialization is allowed before all encounter waves have been exhausted.

Treasury remains in its established pre-Construction position; this slice does not reopen Slice-05 accounting timing.

### 4. Encounter side identity and ordering

A side is an **Empire aggregate at one System**, never a single Fleet container.

The accepted encounter context contains at minimum:

```text
SystemID
AttackerEmpireID
AttackerSeatID
AttackerCombatFleetIDs[]
AttackerShipIDs[]
AttackerCivilianFleetIDs[]
DefenderEmpireID
DefenderSeatID
DefenderCombatFleetIDs[]
DefenderShipIDs[]
DefenderCivilianFleetIDs[]
DefenderColonyIDs[]   // context only; no station tactical unit is synthesized
```

All ID lists are unique and sorted ascending before BattleSession creation.

Only stationary assets at `AtSystemID == SystemID` participate. In-transit assets are excluded until arrival.

Encounter candidate ordering is the deliberate MOOX modernization:

1. `SystemID` ascending;
2. directed attacker `EmpireID` ascending;
3. defender `EmpireID` ascending.

Only current directed `DiplomaticStanceHostile` authorizes an attacker. A reciprocal hostile relation therefore produces two directed candidates, but only the first canonical pair is started for that System in the current wave; after its result the System is re-evaluated from new State.

The attacker must own at least one combat-capable stationary combat Fleet with concrete Ship IDs. Civilian assets alone never initiate combat.

All participating Empires must resolve to a real non-zero Session Seat. A hostile co-location that otherwise requires an unseated/NPC/monster participant is an explicit unsupported error in Slice 06 rather than a silent skip or invented Seat 0.

### 5. Fleet/Ship snapshot vs reference semantics

Battle creation freezes **identity lists**, not a duplicate Core Ship model:

- Battle Spec stores stable System/Empire/Seat/Fleet/Ship/civilian/Colony IDs.
- The authoritative Core State is frozen against normal strategic commands while `PhaseEncounters` is active.
- Result validation is against the Battle Spec identity snapshot.
- Gate 3 does not copy Ship design/equipment/HP into the Battle Spec.
- Slice 07 may derive an immutable tactical snapshot from those frozen Ship IDs when tactical stats are introduced.

This avoids duplicate mutable strategic/tactical Ship state while preserving deterministic battle identity.

### 6. One BattleSession per System per wave

- At most one attacker/defender BattleSession may exist for a given System in one wave.
- Independent Systems may have BattleSessions active in parallel.
- Current-wave Battle IDs remain globally monotonic and seeds continue to use `battle.DeriveSeed(gameSeed, strategicTurn, battleID)`.
- Wall-clock completion order is never authoritative.
- After all Battles in a wave have valid results, results are processed in **Battle-ID ascending** order and contested Systems are re-derived.
- A second pair in a three-party System is therefore a new BattleSession in the next wave, with a new Battle ID/seed.
- The Session remains `PhaseEncounters` between waves; it does not briefly enter `PhasePostResolution`.

`GameSession.battles` may represent the current wave only. Prior-wave BattleSpecs/results remain authoritative through the event log; `nextBattleID` remains monotonic.

### 7. Battle Spec direction is explicit

The current sorted `Participants []SeatID` cannot encode attack direction.

Gate 3 must make attacker/defender side metadata explicit in the Battle Spec. A sorted participant list may be retained only as a compatibility/derived view if validation guarantees it is exactly the two side Seat IDs; it is not the source of attack direction.

The supported Slice-06 battle is exactly two-sided.

### 8. Minimum Battle result contract

Normal supported BattleSession completion requires exactly **one** winner.

The accepted semantic result is conceptually:

```go
type Result struct {
    WinnerSeat       protocol.SeatID
    Outcome          string
    DestroyedShipIDs []core.ID
}
```

Requirements:

- `WinnerSeat` must be exactly the attacker or defender Seat from the Battle Spec.
- `Outcome` remains non-empty descriptive metadata; strategic correctness must not parse free-form Outcome text.
- `DestroyedShipIDs` is optional, unique, sorted and must be a subset of the Battle Spec's concrete attacker/defender `ShipIDs`.
- Fixed-special Colony/Outpost civilian Fleets are not encoded as fake destroyed Ship IDs.
- Gate 3 may preserve a compatibility `WinnerSeats` representation internally only if it validates to exactly one winner; the authoritative supported semantics are singular.

### 9. Atomic final-result / continuation preparation

The last result of a wave must not leave a half-applied session if post-battle strategic validation fails.

Accepted orchestration:

- a pure Battle result validator validates a candidate result against its Battle Spec without mutating the child session;
- non-final child results may be committed to their child BattleSession because they do not mutate strategic State;
- for the final outstanding child, GameSession builds the complete Battle-ID-ordered candidate outcome set using prior completed children plus the still-uncommitted candidate;
- `ResumeAfterEncounters` runs on a **cloned** Core State;
- returned State/events/next-wave encounters are fully validated and any next BattleSessions are prepared before authoritative mutation;
- only after all preparation succeeds is the final child result committed and the wave State/events/new children committed atomically;
- if continuation preparation fails, the final child remains active and the authoritative Core State remains at the same encounter boundary, allowing a corrected retry.

### 10. Applying tactical casualties

For every Battle outcome in Battle-ID order:

1. validate all destroyed IDs against that Battle's frozen Ship IDs;
2. remove those concrete Ships from `GameState.Ships`;
3. remove the IDs from their owning combat `StrategicFleet.ShipIDs`;
4. remove any combat Fleet that becomes empty.

No tactical HP/damage value is persisted in Slice 06.

### 11. Losing-side strategic retreat

After explicit destroyed Ship IDs are removed, every **remaining** losing-side asset at the battle System receives the minimum original-backed strategic disposition.

Destination selection:

- choose the closest **other** System containing an owned Colony of the losing Empire;
- distance metric is squared Galaxy coordinate distance, matching the original `Closest_Other_Colony_Star_` evidence;
- deterministic MOOX tie-break: lowest `SystemID`;
- do not select the battle System itself.

Application:

- surviving losing combat Fleet containers remain separate stable MOOX Fleet IDs but all attempt the same chosen retreat destination;
- losing fixed-special Colony/Outpost civilian Fleets at that System retreat with the losing side;
- establish retreat using existing strategic transit/movement capability rather than adding a persisted `retreating` Core state;
- if no valid other owned Colony System exists, remove/destroy the losing assets;
- if an individual losing Fleet/civilian asset cannot establish the accepted retreat movement to that chosen destination under the existing supported movement capability, remove/destroy that asset rather than searching a second destination.

This is the Slice-06 mapping of the original temporary retreat status + `Process_Retreating_Ships_`; Hyperspace Flux and detailed tactical retreat choice remain deferred.

### 12. Civilian-only overrun

A BattleSession is not created merely to kill unescorted noncombat special ships.

Accepted immediate strategic overrun requires:

- a directed-hostile attacker with a combat-capable side at the System;
- the defender has fixed-special Colony/Outpost civilian Fleet assets there;
- the defender has **no** combat-capable Fleet side there;
- the defender has **no Colony in that System** whose unmodeled orbital/planet defense would make the encounter ambiguous.

Then the defender's civilian special Fleets at that System are removed atomically and a deterministic strategic overrun/destruction event is emitted. Encounter derivation then continues from the updated pre-encounter State.

If a defender Colony is present, Slice 06 does **not** pretend the Colony/station is defenseless; colony/station-only combat remains deferred and no civilian-only auto-overrun is performed through that Colony-defense boundary.

### 13. Colony/orbital-defense context

For a normal Fleet-vs-Fleet Battle at a System containing defender Colonies:

- include sorted `DefenderColonyIDs` as context in the encounter/battle spec;
- do not create synthetic station Ship IDs;
- do not add Star Base/Battlestation/Star Fortress tactical stats;
- do not apply bombardment, invasion or Colony ownership changes from the Slice-06 battle result.

A hostile Fleet versus Colony/station only, with no defender combat Fleet, does not create a BattleSession in this slice. Existing Blockade behavior remains the supported strategic pressure representation after encounter processing.

### 14. Event and revision ordering

Accepted authoritative ordering:

- pre-encounter strategic events keep their existing deterministic resolver order;
- immediate civilian-overrun events use canonical System/attacker/defender order;
- `battle_created` events are emitted in assigned Battle-ID order;
- each completed wave emits `battle_completed` events in Battle-ID order only when the entire wave can be committed;
- strategic casualty/retreat/removal events for that wave follow Battle-ID order, and IDs inside one event payload are sorted;
- next-wave `battle_created` events follow those result events without a phase change out of `PhaseEncounters`;
- after the final wave, post-encounter Blockade/transfer/economy events are emitted, then the Session changes to `PhasePostResolution`.

Revision policy:

- no-encounter turn: one normal authoritative strategic commit/revision;
- encounter turn: one revision for the initial frozen pre-encounter commit, then one revision per atomically applied completed encounter wave;
- the final wave's result application and post-encounter continuation may commit as the same final revision because they are prepared/validated together.

### 15. Observer/replay/recovery contract

Observer-visible current Battle views include the enhanced Battle Spec context and results.

Deterministic replay requirement:

```text
same starting Core State
+ same submitted command batches
+ same externally supplied Battle results
=> same encounter waves
=> same Battle IDs/seeds/specs
=> same civilian overruns/casualty removals/retreats
=> same Blockades/transfers/final State
=> same authoritative event order
```

Active encounter persistence across process restart is **not** claimed. The in-memory staged resolver continuation is cleared after final post-resolution/turn completion. Durable mid-battle GameSession serialization is deferred to a future session-persistence slice.

### 16. Gate-3 minimum regression matrix

Gate 3 must cover at least:

- no-hostile/no-encounter turn remains behaviorally equivalent to current resolution;
- already-stationary hostile Fleet-vs-Fleet encounter;
- same-turn Fleet arrival encounter before Blockade;
- newly constructed combat Ship can join same-turn encounter;
- multiple same-Empire Fleet containers aggregate into one side/spec;
- directed hostility only; reverse-neutral direction does not invent reverse attack;
- reciprocal hostility canonical pair selection;
- two independent Systems create parallel same-wave Battles with stable IDs/seeds;
- three-Empire one-System case produces sequential waves, not parallel overlapping battles;
- wall-clock result completion order does not change authoritative event/result application order;
- invalid final result or invalid destroyed Ship ID leaves final child active and strategic State unchanged;
- destroyed Ship removal + empty Fleet cleanup;
- losing surviving combat/civilian Fleets retreat to closest other owned Colony System;
- deterministic equal-distance retreat tie by System ID;
- no retreat destination / retreat establishment failure destroys losing assets;
- unescorted civilian-only hostile overrun without defender Colony creates no BattleSession;
- defender Colony suppresses civilian auto-overrun/colony-only fake battle;
- Fleet-vs-Fleet over defender Colony carries sorted Colony context but no station tactical unit;
- Blockade and Population transfers execute only after final encounter wave;
- Observer isolation/current-wave Battle view;
- identical-session deterministic replay.
## Explicit deferrals after Gate 1

- tactical movement, firing, damage and ship-system combat internals (Slice 07);
- original human attack-target UI and neutral sneak-attack/war-declaration diplomacy;
- NPC/monster/unseated battle participants;
- colony/station-only space combat and exact orbital defense tactical stats;
- bombardment, invasion and ground combat;
- original Hyperspace Flux retreat destruction special case;
- detailed tactical retreat choices/timing and officer combat effects;
- durable process-restart serialization of an active GameSession/BattleSession.

## Gate 1 conclusion

Gate 1 now has direct evidence for the original trigger boundary, stationary/new-arrival behavior, Empire-side aggregation, pairwise multi-party sequencing, RNG/target selection, civilian handling, losing-side retreat, colony-target context and the strategic work that must wait until combat returns.

The key architecture correction is that MOOX cannot continue to resolve Blockades/transfers before awaiting BattleSession results. The existing child-battle lifecycle is reusable, but the strategic resolver must become a pre-encounter / encounter-wave / post-encounter continuation.

**Gates 1-4 are complete. Slice 06 is closed.**

## Gate 3 implementation

Gate 3 was implemented on 2026-09-01 against the accepted Gate-2 contract. Core remains schema 20 and the economy ruleset remains schema 8; no persisted tactical/encounter state was added to Core.

### Staged strategic resolver

`internal/game/resolver.go` now carries the explicit encounter identity and continuation contract:

```go
type EncounterSide struct {
    EmpireID         core.ID
    SeatID           protocol.SeatID
    CombatFleetIDs   []core.ID
    ShipIDs          []core.ID
    CivilianFleetIDs []core.ID
}

type Encounter struct {
    SystemID          core.ID
    Attacker          EncounterSide
    Defender          EncounterSide
    DefenderColonyIDs []core.ID
    Participants      []protocol.SeatID
}

type EncounterOutcome struct {
    BattleID         uint64
    Encounter        Encounter
    WinnerSeat       protocol.SeatID
    Outcome          string
    DestroyedShipIDs []core.ID
}

type EncounterResolver interface {
    Resolver
    ResumeAfterEncounters(ResolveContext, *core.GameState, []EncounterOutcome) (Resolution, error)
}
```

The old one-shot `Resolver` remains supported. If a resolver returns encounters without implementing `EncounterResolver`, `GameSession` rejects the resolution before authoritative State mutation.

`EconomyResolver.Resolve` now stops after strategic Fleet transit/arrival and Construction, applies deterministic civilian-only overrun processing and derives the current encounter wave. If no BattleSession is needed it immediately executes the post-encounter continuation. Otherwise it returns the frozen pre-post-resolution State plus encounters.

`EconomyResolver.ResumeAfterEncounters` applies a completed wave, re-derives contested Systems and either returns the next wave or performs the deferred post-encounter stages:

```text
recompute Blockades
advance Population transfers
recalculate next-state Colony economy
rematerialize Food logistics
```

Therefore Blockades and Population transfers cannot observe pre-battle Fleet positions on an encounter turn.

### Encounter derivation

`internal/game/encounter.go` derives authoritative sides from stationary strategic State:

- all same-Empire combat Fleet containers at a System are aggregated;
- all concrete `ShipIDs` from those Fleets are aggregated and sorted;
- fixed Colony/Outpost special Fleets are carried separately as civilian Fleet IDs;
- defender Colony IDs are attached as context only;
- only `AtSystemID != 0` stationary assets participate;
- only directed `DiplomaticStanceHostile` authorizes the attacker;
- candidate order is System ID, attacker Empire ID, defender Empire ID;
- at most one BattleSession-producing pair is returned for one System in a wave;
- all Battle-required Empires must map to non-zero Session Seats.

Three-party Systems therefore naturally produce sequential waves after the preceding result changes strategic State. Different Systems may appear in the same wave.

### Civilian-only overrun

An already-hostile combat side may immediately overrun fixed Colony/Outpost civilian Fleets only if the target Empire has no combat Fleet and no Colony at that System. The deterministic event is:

```text
empire.civilian_fleets_overrun
```

A defender Colony suppresses this shortcut. Slice 06 does not synthesize orbital stations or pretend Colony defenses are absent.

### Battle Spec and Result

`internal/battle/session.go` now has explicit strategic side metadata while retaining generic legacy BattleSession support.

A strategic `battle.Spec` includes:

- System ID;
- explicit attacker and defender `battle.Side` with Empire/Seat/Fleet/Ship/civilian IDs;
- sorted defender Colony IDs;
- participant Seats;
- stable Battle ID and deterministic seed.

Strategic specs require exactly two distinct sides and strictly ascending non-zero identity lists. The sorted participant list is no longer the source of attack direction.

`battle.Result` now supports singular strategic semantics:

```go
type Result struct {
    WinnerSeat       protocol.SeatID
    WinnerSeats      []protocol.SeatID // compatibility path, normalized to one
    Outcome          string
    DestroyedShipIDs []core.ID
}
```

Exactly one winner is required. `Outcome` remains descriptive metadata only. Destroyed Ship IDs are sorted/unique and must belong to the frozen attacker/defender Ship snapshot.

`battle.Session.ValidateResult` performs pure validation/normalization without changing Battle phase. This is used for the final child in a strategic wave so the strategic continuation can be proven valid before committing that final result.

### Atomic encounter-wave commit

`GameSession` now retains the staged resolver and ResolveContext in memory only while a real strategic encounter is active.

For non-final children in a parallel wave, valid results may complete the child but produce no authoritative `battle_completed` event yet.

For the final active child:

1. validate the candidate result without completing it;
2. collect the whole wave's normalized outcomes in Battle-ID order;
3. clone the frozen Core State;
4. call `ResumeAfterEncounters` on the clone;
5. validate the returned strategic State/events;
6. prepare every next-wave BattleSession before mutation;
7. only then complete the final child and atomically commit strategic State, one new revision, all `battle_completed` events in Battle-ID order, strategic casualty/retreat/overrun events and any next-wave `battle_created` events.

If result validation, continuation or next-wave preparation fails, the final child remains `active`, the authoritative strategic State/revision remain unchanged and the caller may retry with a corrected result.

A following wave in the same System stays in `PhaseEncounters`; no transient `PhasePostResolution` event is emitted. Battle IDs remain monotonic across waves and seeds remain derived from game seed + strategic turn + Battle ID.

The legacy manual `BeginEncounters` path remains supported and deliberately clears any staged strategic continuation. `CompleteTurn` also clears encounter continuation state.

### Casualties and strategic retreat

Explicit destroyed concrete Ships are removed from `GameState.Ships`, from owning combat Fleet `ShipIDs`, and empty combat Fleets are removed. The event is:

```text
empire.battle_casualties_applied
```

Every surviving losing-side combat Fleet and fixed Colony/Outpost civilian Fleet still at the battle System receives the minimum strategic retreat behavior proven in Gate 1:

- find the closest other System containing a Colony owned by the losing Empire;
- use squared Galaxy coordinate distance;
- equal distance ties resolve to lower System ID;
- keep separate MOOX Fleet IDs rather than merging them;
- establish existing strategic transit fields toward that destination;
- if no destination exists, or the current supported movement profile cannot establish transit, remove/destroy that asset without trying a second destination.

Successful retreat emits:

```text
empire.fleet_retreated_after_battle
```

Retreat failure/destruction emits:

```text
empire.fleet_destroyed_after_battle
```

with deterministic reason `no_retreat_destination` or `retreat_movement_unavailable`.

No tactical HP/damage/fire model, Hyperspace Flux rule or tactical retreat-choice system was added.

### Strategic timing proven by vertical tests

Gate-3 tests now prove the accepted timing, not merely the helper functions:

- a combat Fleet whose ETA reaches zero during current Fleet transit is present in the same turn's BattleSpec;
- a military Ship completed by current-turn Construction creates its one-Ship Fleet before encounter derivation and participates immediately;
- Blockade state remains unmaterialized at the frozen encounter boundary and is recomputed only after battle result/retreat;
- a Population transfer with one turn remaining does not advance while the BattleSession is pending; after the final wave it sees the newly computed Blockade and is deterministically lost as `destination_blockaded`;
- a real `GameSession` + real `EconomyResolver` encounter pauses in `PhaseEncounters`, exposes detached BattleSpec data through Observer, retreats the loser on completion, then materializes Blockade and enters `PhasePostResolution`.

### Gate-3 deterministic regression coverage

Focused coverage now includes:

- no-hostile/no-encounter real EconomyResolver compatibility;
- already-stationary Fleet-vs-Fleet handoff;
- same-turn strategic Fleet arrival;
- same-turn military construction completion;
- aggregation of multiple same-Empire Fleet containers into one side;
- directed hostility and reciprocal-hostility canonical direction;
- two independent Systems in one encounter wave;
- three-Empire same-System sequential waves;
- explicit non-zero Seat requirement for every Battle participant;
- stable Battle IDs/seeds and preservation of attacker/defender direction;
- wall-clock completion independence: Battle 2 may finish before Battle 1 but authoritative completion commits 1 then 2;
- invalid out-of-snapshot casualty result leaves the final child active and does not invoke the strategic continuation;
- continuation failure after a valid final result leaves State/revision/final child unchanged and retryable;
- destroyed Ship removal and empty Fleet cleanup;
- escorted civilian Fleet retreat with a losing combat side;
- closest-Colony retreat and equal-distance System-ID tie break;
- no retreat destination destruction;
- movement-profile-unavailable destruction even when a Colony destination exists;
- civilian-only overrun and defender-Colony suppression;
- Colony context in Fleet-vs-Fleet BattleSpec;
- Population-transfer and Blockade work delayed until the final encounter wave;
- Observer BattleSpec isolation;
- identical real-session State/Event/Battle replay.

Gate-3 QA passed:

```text
gofmt changed/untracked Go files
go test ./internal/battle ./internal/game ./internal/session -run '(StrategicBattle|Encounter|HostileEncounter|PopulationTransferWaits|StagedEncounter|EconomyResolverSessionEncounter|EconomyResolverNoEncounter)' -count=1
go test ./... -count=1
git diff --check
```

`go vet ./...`, final staged-diff review, commits, HISTORY/status closure and OPEN-marker removal remain Gate 4 work.

### Deferrals preserved

Gate 3 did not add tactical movement/fire/damage, neutral sneak-attack/war-declaration UI/semantics, NPC/monster unseated combat, station/Colony-only space battle, bombardment/invasion, ground combat, Hyperspace Flux retreat destruction, detailed tactical retreat choice/officers, or durable active-GameSession process-restart serialization.

## Gate 4 closure

Gate 4 was completed on 2026-09-01.

- Gameplay/evidence commit: `096fd0a` (`game: add strategic encounter battle handoff`).
- Core `StateSchemaVersion` remains **20**.
- Economy ruleset schema remains **8**.
- No tactical damage/weapon/bombardment/invasion implementation was added.
- Focused CombatFleet/hostility/Blockade/Encounter/BattleSession and parallel replay-order regressions passed.
- `go test ./... -count=1` passed.
- `go vet ./...` passed.
- `git diff --check` and tracked/untracked whitespace checks passed.
- The Slice-06 `_OPEN_` recovery marker is removed by the closing documentation commit.
- Next prepared objective: Slice 07 **Tactical ship combat baseline**.