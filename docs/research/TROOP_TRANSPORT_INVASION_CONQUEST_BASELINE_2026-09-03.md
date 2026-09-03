# Troop Transport / invasion / conquest baseline - original 1.31 evidence

Status: **Gate 4 independent QA complete; final commit/closure pending**

Date: **2026-09-03**

Slice: `docs/slices/PLANNED_11_TROOP_TRANSPORT_INVASION_CONQUEST_BASELINE.md`

Open marker: `docs/slices/_OPEN_TROOP_TRANSPORT_INVASION_CONQUEST_BASELINE_2026-09-03.md`

Starting implementation baseline: `f353a05` after closed Slice 10 / Core22.

## Research objective

Freeze only the minimum directly evidenced territorial-war path needed to build/move a Troop Transport, invade a hostile Colony after required space-combat conditions, resolve one deterministic ground-combat fixture, transfer Colony ownership, initialize conquered Population state and recalculate dependent strategic/economic state.

Gate 1 is research/checkup only. No Slice-11 gameplay code was changed.

## Sources and method

Primary original reference:

- `reference/original/support/files/Orion2.exe` - Master of Orion II 1.31 executable with embedded debug symbols.

Read-only executable inspection used the existing helper:

- `build/research-disasm/main.go`
- object-1 offset = function VA - `0x10000`.

Prior permanent MOOX evidence reused where it already established the surrounding architecture:

- `docs/research/STRATEGIC_HOSTILE_ENCOUNTERS_BATTLE_HANDOFF_2026-08-31.md`
- `docs/research/COMMAND_POINTS_SHIP_MAINTENANCE_2026-08-31.md`
- `docs/research/POPULATION_COHORTS_2026-08-29.md`
- `docs/research/POPULATION_TRANSPORT_FREIGHTER_2026-08-29.md`
- `docs/research/DIPLOMACY_WAR_PEACE_BASELINE_2026-09-02.md`
- `docs/research/MOO2_GAME_REFERENCE.md`

Temporary research-only xref helpers were created under `build/research-disasm` and deleted immediately after use; no research helper remains in the work tree.

## Current MOOX baseline

### Colony ownership and strategic settlement

`core.Colony` already owns the authoritative `EmpireID`, `PlanetID`, Population, Buildings, Economy materialization and current Construction state. Planet-to-Colony linking is independent of Empire ownership, so conquest does not require replacing the Colony ID or Planet link.

`core.Empire.Capital` is persistent. Current validation only proves that a referenced Capital Colony exists; conquest makes owner consistency important, so Slice 11 should strengthen the invariant to require a non-zero Capital to belong to that Empire.

The existing post-encounter continuation in `EconomyResolver.finishPostEncounter` already runs, in order:

1. recompute System blockades;
2. advance Population transfers;
3. recalculate every Colony's local/race-aware economy and Population dynamics;
4. rematerialize Food logistics.

A Population transfer whose destination Colony has changed owner is already lost with `destination_unavailable`, releasing its reserved Freighters. Therefore conquest must happen **before** `finishPostEncounter`; the existing continuation can then correctly settle captured-state Blockades, transfers, Economy and Food without a second authority path.

### Existing Population conquest vocabulary

`core.PopulationCohort` already persists:

```go
OriginEmpireID     ID
LoyaltyEmpireID    ID
AssimilationState  PopulationAssimilationState
```

with canonical states:

```text
assimilated
conquered
```

Current validation already enforces the exact narrow semantics needed by Slice 11:

- conquered Population cannot have `OriginEmpireID == current Colony owner`;
- conquered Population remains loyal to its Origin Empire;
- assimilated Population is loyal to the current Colony owner;
- Origin/Loyalty Empires must exist.

No production code currently advances `conquered -> assimilated`. Slice 11 therefore only needs to initialize the captured state; ongoing assimilation/rebellion progression remains a later feature.

### Existing fixed-special Fleet model

MOOX already represents Colony and Outpost ships as one fixed-purpose civilian `StrategicFleet` each:

- `Role == civilian`;
- fixed `SpecialKind`;
- no concrete tactical `ShipIDs`;
- normal strategic FTL transit fields;
- included as `CivilianFleetIDs` in strategic Encounter presence;
- unescorted civilian-only assets can be overrun/destroyed strategically;
- civilians on the losing side retreat/die with that side after Battle resolution.

This is the correct representation family for the original Troop Transport. It is **not** the existing Settler/Population-transfer system; that distinction was already frozen in prior Population-transfer research.

### Current Encounter / continuation boundary

`GameSession.ResolveStrategic` commits the pre-encounter `Resolution` and enters `PhaseEncounters` when BattleSessions are present.

When the final active Battle of a wave completes, Session calls:

```text
EncounterResolver.ResumeAfterEncounters(...)
```

Current `EconomyResolver.ResumeAfterEncounters`:

1. applies Encounter outcomes;
2. derives another Encounter wave if required;
3. if no Encounter remains, immediately calls `finishPostEncounter`.

That is one boundary too short for original-compatible invasion. Slice 11 needs an authoritative invasion decision/resolution **after orbital combat outcome application but before another encounter continuation / final post-encounter settlement**.

## Direct original 1.31 symbol map

Important executable symbols found directly in `Orion2.exe`:

| VA | Symbol |
| --- | --- |
| `0x5663E` | `Load_Transport_Ship_Design_` |
| `0x8B7A5` | `Unloading_Transports_In_Main_Screen_` |
| `0xCFF02` | `Colony_Is_Conquered_` |
| `0xD9F85` | `Intercept_Invaders_` |
| `0xDA99C` | `Move_Transports_` |
| `0xE8776` | `Player_Owns_Transports_` |
| `0xEC15C` | `Compute_Player_Ground_Combat_Bonuses_` |
| `0xEC3A0` | `Find_Troops_` |
| `0xEC3CE` | `Compute_Ground_Combat_Info_` |
| `0xEC4FE` | `Ground_Combat_Round_` |
| `0xEC601` | `Resolve_Ground_Combat_` |
| `0xEC61E` | `Colony_N_Militia_` |
| `0xEC831` | `Get_Invasion_Info_` |
| `0xEC97C` | `Enforce_Population_Limits_At_Colony_` |
| `0xECBF7` | `Change_Pop_Ownership_` |
| `0xECE05` | `Resolve_Invasion_Troops_` |
| `0xECF41` | `Change_Colony_Ownership_` |
| `0xED48B` | `Invade_` |
| `0xED59D` | `Colony_Infantry_Limit_` |
| `0xED6B7` | `Unload_Transports_` |
| `0xED713` | `Compute_Colony_Ground_Combat_Info_` |

## Finding 1 - Transport is fixed special Ship type 2, cost 100, 1 CP

Evidence level: **direct original code**.

`Load_Transport_Ship_Design_` clears a fixed special design and writes:

```text
design+0x11 = 2     // Transport special type
design+0x10 = 0     // therefore 1 Command Point in the original CP scan
design+0x5c = 0x2f  // fixed Transport name/string slot
```

It uses the current best Warp drive, an automatically selected civilian armor tier, no tactical weapons, and asks `Ship_Type_Cost_For_Player_` for type `2`.

`Ship_Type_Cost_For_Player_` gives:

```text
type 1  Colony Ship = 500 PP base
type 2  Transport   = 100 PP base
type 4  Outpost     = 100 PP base
```

and then applies the existing government cost reduction path. This matches the current MOOX Colony/Outpost fixed-special construction model and the prior direct CP evidence that all three fixed special Ships consume exactly 1 CP while stationary or in transit.

The `Add_To_Queue_` path maps its Transport auto-design sentinel to `Load_Transport_Ship_Design_` directly. Gate 1 did not find a separate Transport technology prerequisite in the fixed-design/queue path. Gate 2 therefore proposes a baseline-buildable 100-PP support project rather than inventing a technology ID; if a later direct unlock dependency is found it can be added without changing Transport identity/capacity.

## Finding 2 - one extant Transport always represents four Infantry

Evidence level: **direct original code**.

`Get_Invasion_Info_` iterates the selected invasion Ship IDs. For every Ship with:

```text
ship+0x11 == 2
```

it executes the equivalent of:

```text
attacker_ground_info.infantry += 4
```

The count is added to ground-info offset `+0x0c`, the same troop-type slot into which defender Colony standing Infantry (`colony+0x130`) is loaded.

`Unload_Transports_` independently proves the same capacity: unloading a Transport destroys that Ship and adds exactly **4 Infantry** to the friendly Colony garrison, subject to the original Infantry limit.

Important representation consequence for MOOX: the original Transport does **not** need a variable cargo field while it exists. A surviving Transport after invasion always corresponds to a complete four-Infantry group. Therefore Slice 11 should not persist a second mutable `TroopPayload` on the Fleet.

## Finding 3 - Transport uses ordinary strategic movement and remains noncombat

Evidence level: **direct original code + prior direct combat evidence**.

`Move_Transports_` calls normal Ship movement helpers including:

- `Ship_Destination_Star_`;
- `Ships_Try_To_Move_To_`;
- `Make_Ships_Move_To_`;
- `Move_Ships_With_Possible_Intermediate_`.

`Player_Owns_Transports_` scans owner Ships and recognizes special type `2`; it is a distinct strategic Ship class, not a Settler/Freighter reservation.

Prior direct `Search_For_Battles_` / HELP evidence proves Transport does not tactically fight. Like Colony/Outpost ships it is a civilian/noncombat asset: unescorted civilian assets can be destroyed, and civilian survivors follow their defeated side's retreat disposition.

Therefore the current fixed-special `StrategicFleet` movement + `CivilianFleetIDs` architecture is reusable. Slice 11 only needs to add Transport to the supported fixed-special set and existing 1-CP accounting.

## Finding 4 - invasion occurs after orbital combat and before Blockade/settlement

Evidence level: **direct original code**.

Correct `Next_Turn_Calc_` at VA `0x136B3` has the already-proven relevant order:

```text
Move_All_Ships_Toward_Stars_
...
Apply_All_Colony_Changes_
...
Search_For_Battles_
...
All_Colony_Calculations_
Compute_Blockades_
Move_Settlers_
All_Colony_Calculations_
...
```

Fresh xref work for Slice 11 proves `Invade_` is called from `Do_Attacker_Beat_Colony_Stuff_` VA `0xE87D2`.

`Do_1_Combat_` calls `Russ_Combat_`, handles noncombat destruction/retreat, and then calls `Do_Attacker_Beat_Colony_Stuff_` for the relevant defeated Colony side. That post-space-combat function owns the original bombard/mind-control/invasion choices and invokes `Invade_` on its invasion branches.

Therefore invasion is **not** a generic pre-combat planning mutation. It is a post-orbital-combat strategic decision/resolution and happens before the later Blockade/Settler/Colony settlement path.

MOOX architectural consequence: do not predeclare/auto-resolve invasion in the ordinary Planning CommandBatch. Add a narrow server-authoritative invasion-decision phase between Encounter continuation and `finishPostEncounter`.

## Finding 5 - minimum ground combat is deterministic d100 opposed rolls

Evidence level: **direct original code**.

`Compute_Player_Ground_Combat_Bonuses_` first zeroes its 0x13-byte bonus structure, then applies race/government/technology/leader modifiers. Slice 11 can therefore define an exact zero-modifier fixture without implementing those later families.

`Get_Invasion_Info_` represents four troop-type count slots at ground-info offsets:

```text
+0x0a  type 0
+0x0c  type 1  // Infantry; Transport contributes here
+0x0e  type 2  // Militia in the basic Colony path
+0x10  type 3
```

`Find_Troops_` advances the active type from 0 upward until it finds a nonzero count; active type `4` is the no-troops sentinel.

The `Compute_Ground_Combat_Info_` jump table was read directly:

```text
type 0 -> combat branch at 0xEC494
type 1 -> branch at 0xEC4AD
type 2 -> branch at 0xEC4C2
type 3 -> branch at 0xEC4BB
```

For the zero-modifier baseline:

- Infantry/type 1 receives no type penalty and has one-hit durability;
- Militia/type 2 receives **-10** combat score and has one-hit durability;
- advanced type-0/type-3 handling is outside the Slice-11 fixture.

`Ground_Combat_Round_` performs, in order:

```text
attacker_roll = Random_(100) + attacker_active_type_score
defender_roll = Random_(100) + defender_active_type_score
```

Then:

- if attacker roll <= defender roll, attacker takes one hit;
- if attacker roll >= defender roll, defender takes one hit;
- therefore an exact tie damages both sides;
- reaching the active troop type's hit threshold decrements that troop count and resets accumulated damage;
- combat advances troop types when a count reaches zero.

For the proposed first fixture, all supported units have threshold 1, so one scored hit removes one unit.

`Invade_` loops `Ground_Combat_Round_` until either side reaches the type-4 no-troops sentinel.

### Proposed canonical first fixture

Use only already-evidenced basic types and no modifier families:

- attacker: one selected Transport = 4 Infantry;
- defender: no standing advanced units; optionally standing Infantry plus derived Militia;
- no ground technologies, race ground bonus, officer/leader bonus or special government combat modifier;
- Infantry score 0;
- Militia score -10;
- one hit removes one supported unit;
- rolls consume the authoritative MOOX `GameState` RNG in attacker-then-defender order via `Intn(100)`.

The existing MOOX RNG checkpoint model is suitable: resolve on a cloned State, use `state.RNG()`, and `CommitRNG` only when the accepted invasion resolution commits.

## Finding 6 - basic Colony Militia is derived from Population / 5

Evidence level: **direct original code**.

`Colony_N_Militia_` counts eligible Colony Population records and returns integer division by `5`.

For the narrow MOOX organic-cohort baseline, Gate 2 proposes:

```text
Militia = floor(eligible non-conquered organic Colony Population / 5)
```

The direct `Colony_N_Militia_` loop accepts normal organic source codes `< 8` and rejects entries whose conquered bit is set. Current MOOX runtime cohorts are all organic, so the Slice-11 canonical derivation is the sum of `PopulationAssimilated` cohorts only, divided by five with floor. Conquered cohorts do **not** generate Militia. Android/Native population kinds are still not runtime schema values and remain deferred. Militia is derived at invasion time and is not persisted as a second Colony state field. Standing Infantry must be persisted because original successful invaders are installed as the new Colony garrison and surviving defender Infantry remains after a failed invasion.

## Finding 7 - failed invasion destroys every participating Transport

Evidence level: **direct original code**.

When the attacker runs out of troops, `Invade_`:

- keeps the Colony under the defender;
- writes surviving defender standing troop counts back to Colony ground-force fields;
- iterates the selected invasion Ship IDs;
- calls `Kill_Ship_` for every special type-2 Transport.

Therefore Slice 11 failure semantics are unambiguous:

- every **selected** participating Transport is consumed/destroyed;
- unselected co-located Transports are not part of that invasion and remain;
- Population/Colony owner remain unchanged;
- surviving standing defender Infantry remains; derived Militia is not a persistent second garrison pool.

## Finding 8 - successful invasion may preserve full four-troop Transports

Evidence level: **direct original code**.

When attacker troops survive, `Invade_` first calls:

1. `Change_Colony_Ownership_`;
2. `Resolve_Invasion_Troops_`.

`Resolve_Invasion_Troops_` performs a very useful survivor packing algorithm:

1. reserve one surviving Infantry as the conquered Colony's initial garrison;
2. iterate participating Transport Ships backwards;
3. for every complete remaining group of four Infantry, subtract four and keep that Transport alive;
4. otherwise destroy that Transport;
5. add the final remainder (<4) to the Colony standing Infantry, subject to the original Colony Infantry limit.

Thus an extant original Transport remains an implicit four-Infantry Transport; partial payload does not persist.

MOOX can reproduce this deterministically without a cargo field by sorting selected Transport Fleet IDs and processing them in reverse canonical order for survivor packing.

The exact original planet/government Infantry-capacity table is outside the first fixture. Gate 2 should keep Colony Infantry non-negative and choose a canonical fixture whose resulting 1..4 Infantry garrison does not require the advanced cap rule; the exact cap can be added with the broader ground-force layer later.

## Finding 9 - ownership transfer maps directly to existing MOOX conquered cohorts

Evidence level: **direct original code + existing Core invariants**.

`Change_Colony_Ownership_` calls `Change_Pop_Ownership_` before its later building/technology side effects.

`Change_Pop_Ownership_`:

- handles lost-Colony leader/homeworld bookkeeping;
- changes the Colony owner;
- iterates Population records;
- treats Population already native to the new owner differently from foreign Population;
- marks foreign Population conquered unless an original special-trait branch applies;
- recalculates Colony environmental/population state;
- clears the Colony construction queue;
- recomputes star/Colony state.

This matches current MOOX cohort invariants exactly. Gate 2 proposes, on capture:

```text
if cohort.OriginEmpireID == attacker:
    LoyaltyEmpireID = attacker
    AssimilationState = assimilated
else:
    LoyaltyEmpireID = cohort.OriginEmpireID
    AssimilationState = conquered
```

The Colony's `EmpireID` becomes the attacker. No automatic `conquered -> assimilated` progression is added in Slice 11.

Original Telepathic/mind-control immediate-control behavior is a special exception and remains explicitly deferred.

## Finding 10 - capture clears production and reassigns Capital deterministically

Evidence level: **direct original code**.

`Change_Pop_Ownership_` calls `Clear_Colony_Queue_`. MOOX therefore should clear the captured Colony's active `Construction` rather than allowing the conqueror to inherit the previous owner's unfinished project.

When the captured Colony was the defender's home planet, original `Assign_New_Home_Planet_` scans the defender's other Colonies and chooses the one with the **largest Population**; if no Colony remains, home planet becomes none.

The original also assigns the captured Colony as the new owner's home planet if the new owner currently has no home planet.

MOOX Gate-2 adaptation:

- if defender `Capital == capturedColony.ID`, choose the defender's remaining owned Colony with greatest total Population; break equal-population ties by lowest canonical Colony ID; if none remains set Capital=0;
- if attacker `Capital == 0`, set it to the captured Colony;
- strengthen Core validation so a nonzero `Empire.Capital` must reference a Colony owned by that Empire.

Empire elimination itself remains Slice 12.

## Finding 11 - original capture has additional side effects deliberately outside baseline

Evidence level: **direct original code**.

`Change_Colony_Ownership_` / `Change_Pop_Ownership_` also contain branches for:

- special building removal/destruction;
- captured-technology eligibility and random technology acquisition;
- leader handling;
- special government/race/Telepathic behavior;
- exact Infantry/Barracks capacity limits;
- additional original Colony status fields that do not yet have MOOX equivalents.

These are not necessary to prove the first territorial ownership loop. A Gate-3 canonical fixture can use no relevant special buildings/ground technology/leaders so omission is explicit rather than accidental.

## Gate-2 accepted / frozen contract

The following Slice-11 v1 contract was **accepted and frozen in Gate 2 on 2026-09-03**. Gate-2 review refinements later in this document are normative where they narrow earlier Gate-1 proposal wording.

### 1. Core schema 23 - Transport identity and standing Infantry

Advance Core `StateSchemaVersion` **22 -> 23**.

Add:

```go
const StrategicFleetSpecialTroopTransport StrategicFleetSpecialKind = "troop_transport"

type ColonyGroundForces struct {
    Infantry int `json:"infantry,omitempty"`
}
```

and to `Colony`:

```go
GroundForces ColonyGroundForces `json:"ground_forces,omitempty"`
```

Canonical invariants:

- `GroundForces.Infantry >= 0`;
- a Troop Transport is `Role == civilian`, `SpecialKind == troop_transport`, has no concrete tactical `ShipIDs`, and otherwise uses the existing fixed-special location/transit invariants;
- an extant Transport implicitly carries exactly **4 Infantry**; no cargo/payload integer is persisted;
- a nonzero Empire Capital must reference a Colony currently owned by that Empire.

New Game may leave standing Infantry at zero; basic Militia is derived from Population at invasion time.

### 2. Construction / movement / CP

Add construction kind/project:

```text
troop_transport
```

with:

- base production cost **100 PP**;
- same already-supported Feudal government cost reduction path as Colony/Outpost fixed ships;
- no invented technology prerequisite in Slice 11;
- one completed project creates one fixed-special civilian Troop Transport Fleet at the Colony System;
- its FTL speed uses the same supported current-best-drive derivation as other fixed specials;
- each extant Transport consumes **1 Command Point**;
- existing `empire.move_fleet` is its strategic movement command; no second transport movement authority is added.

Friendly manual Transport unloading/reinforcement is not required for the first conquest loop and remains deferred; invasion cleanup is in scope.

### 3. Encounter integration / orbital-clear eligibility

Include Troop Transport Fleet IDs in existing `EncounterSide.CivilianFleetIDs` exactly like Colony/Outpost fixed specials.

A narrow invasion opportunity may exist when all are true:

- attacker and Colony owner are at `war` under Slice-10 authority;
- target is an enemy Colony in the System;
- attacker has an authorized Seat in the current `ResolveContext`; unseated/AI invasion planning remains deferred;
- attacker has at least one stationary combat-capable Fleet remaining in the System;
- attacker has at least one stationary Troop Transport there;
- target owner has no stationary combat-capable Fleet remaining after the applicable Encounter outcome/retreat;
- attacker has uncontested modeled orbital control: no other stationary Empire Fleet in the System is currently war-hostile to the attacker under `MayAttackEmpire`;
- no defender-owned modeled orbital command-station building (Star Base / Battlestation / Star Fortress) remains anywhere in that target System; station-only tactical combat is not silently bypassed by the baseline;
- no ambiguous same-Colony multi-attacker/simultaneous-invasion case is present in the v1 fixture.

This deliberately does not model orbital station-only tactical defenses, bombardment, neutral sneak invasion or transport-only conquest without supported orbital control.

### 4. New Session phase: `invasion_decisions`

Add a transport-neutral Session phase between Encounters and PostResolution:

```text
planning
strategic_resolution
encounters
invasion_decisions
post_resolution
```

The strategic resolver must be able to return a pending `InvasionOpportunity` before `finishPostEncounter`.

Required order:

```text
movement / construction
-> encounter boundary
-> apply completed space-Battle outcome(s)
-> invasion opportunity/decision/resolution
-> any required next encounter continuation
-> recompute blockades
-> Population transfers
-> Colony Economy/Population recalc
-> Food logistics
-> post_resolution
```

This preserves the original post-space-combat/pre-Blockade ownership boundary.
Gate-2 timing refinement: MOOX v1 evaluates invasion opportunities at **prepared encounter-wave boundaries**. An already-prepared BattleSession wave completes and its outcomes/retreats commit first; then invasion opportunities may interpose before later post-encounter settlement. Exact original interleaving between individual battles inside one already-prepared parallel MOOX wave is explicitly deferred rather than rewriting the Slice-06/07 BattleSession architecture.

The phase graph may therefore legally cycle:

```text
encounters -> invasion_decisions -> encounters
```

before eventually reaching `post_resolution`.

### 5. Pending opportunity / projection

Minimum deterministic opportunity data:

```go
type InvasionOpportunity struct {
    SystemID                  core.ID
    ColonyID                  core.ID
    AttackerEmpireID          core.ID
    DefenderEmpireID          core.ID
    AttackerSeatID            protocol.SeatID
    EligibleTransportFleetIDs []core.ID // sorted unique
}
```

Pending opportunity is Session continuation state, analogous to active BattleSession state, not a second persistent diplomacy/gameplay authority inside `GameState`.
Exactly **one** opportunity is active at a time. Canonical choice is lowest `(SystemID, ColonyID, AttackerEmpireID)` among currently legal opportunities. Session also carries a sorted transient handled-key set `(ColonyID, AttackerEmpireID)` for the current turn. `decline`, failed invasion and successful invasion all mark that key handled before continuation, so the same attacker/Colony pair cannot be offered repeatedly in one turn even if unselected Transports remain. The handled set is cleared when the turn completes. Other independent opportunities can then be derived and handled sequentially; this is not simultaneous ground-combat resolution.

The active opportunity and handled keys are transient Session continuation state like active BattleSessions. Core save/load after a resolved invasion must round-trip exactly, but durable process-restart serialization of an in-progress `invasion_decisions` phase remains outside Slice 11 just as active BattleSession restart serialization remains deferred.

Player projection exposes an invasion opportunity only to the acting attacker Seat. Observer may expose the complete pending opportunity for debug/proof.

### 6. Immediate commands / authority

Reuse:

```text
POST /api/v1/games/{gameID}/immediate-commands
```

Keep Command/Event schema **1**.

Add command kinds:

```text
invasion.invade
invasion.decline
```

Strict payloads:

`invasion.invade`:

```json
{
  "colony_id": 42,
  "transport_fleet_ids": [101, 103]
}
```

`invasion.decline`:

```json
{"colony_id": 42}
```

Authority/transaction rules:

- only `PhaseInvasionDecisions`;
- exact current `base_revision`;
- exactly one immediate command, `Sequence == 1`;
- acting Empire derived from SeatID, never client supplied;
- Colony must match the pending opportunity;
- selected Transport IDs sorted/unique/nonempty and a subset of eligible stationary attacker Transports;
- war/orbital-control eligibility revalidated on the cloned authoritative State;
- rejection changes no State, Session revision, RNG, events or pending decision.

The web baseline may submit all eligible Transports by default while the protocol permits a legal subset.

### 7. Narrow ground-combat engine

For an accepted `invasion.invade`:

```text
attacker Infantry = 4 * number(selected Transports)
defender Infantry = Colony.GroundForces.Infantry
defender Militia  = floor(sum(assimilated organic Colony Population) / 5)
```

Slice-11 supported combat modifiers are exactly zero. Advanced ground technologies, race/government combat modifiers and leaders are not partially approximated.

Supported scores/durability:

```text
Infantry score = 0, one hit per unit
Militia  score = -10, one hit per unit
```

Active defender order follows original type order: standing Infantry before Militia when no type-0 unit exists.

Each round consumes authoritative RNG in exact order:

```text
A = rng.Intn(100) + attackerScore
D = rng.Intn(100) + defenderScore
if A <= D: attacker active unit loses one hit
if A >= D: defender active unit loses one hit
```

A tie therefore removes one supported unit from each side. Continue until attacker Infantry reaches zero or defender Infantry+Militia reaches zero. Commit RNG once as part of the accepted atomic command.

### 8. Failure result

If attacker Infantry reaches zero first:

- Colony owner unchanged;
- Population unchanged;
- active Construction unchanged;
- surviving defender standing Infantry becomes `Colony.GroundForces.Infantry`;
- derived Militia is not persisted;
- every selected Transport Fleet is removed/consumed;
- unselected Transports remain;
- Session resolves/clears the pending opportunity and continues strategic resolution.

### 9. Capture result / survivor packing

If attacker has `S >= 1` surviving Infantry:

1. `remaining = S - 1`;
2. reserve one Infantry for the conquered Colony;
3. process selected Transport Fleet IDs in reverse canonical ID order;
4. while `remaining >= 4`, subtract four and keep that Transport alive;
5. otherwise consume that Transport;
6. set captured Colony standing Infantry to `1 + remaining` for the first fixture;
7. change Colony `EmpireID` to attacker;
8. clear captured Colony `Construction`;
9. initialize every Population cohort:
   - origin == attacker -> assimilated, loyalty attacker;
   - origin != attacker -> conquered, loyalty origin;
10. if defender Capital was captured, choose remaining defender Colony with greatest total Population (tie lowest Colony ID), otherwise Capital=0 if none;
11. if attacker Capital is zero, captured Colony becomes attacker Capital;
12. keep diplomatic relation at war;
13. defer building destruction/captured-tech/Telepathic/special original side effects;
14. resume the existing post-encounter boundary so Blockades, transfer loss/arrival, Economy, Population dynamics and Food logistics materialize from the new owner state;
15. do **not** rerun Treasury settlement or current-turn Command-Point maintenance after capture. Those were already settled before strategic movement/combat. Transport destruction/survival and Colony/station ownership affect the next normal Treasury/CP settlement, avoiding double income or double maintenance charges.

Exact original Infantry-capacity rules are not generalized in Slice 11; the canonical fixture must not exceed the supported 1..4 conquered garrison range.

### 10. Events / history / web proof

Keep Event schema **1** and add minimum deterministic events:

```text
empire.invasion_declined
empire.invasion_resolved
empire.colony_conquered
```

`empire.invasion_resolved` should include enough deterministic summary data for replay/debug projection without storing wall-clock time:

- Colony/System IDs;
- attacker/defender Empire IDs;
- selected Transport IDs;
- initial attacker Infantry;
- initial defender Infantry/Militia;
- surviving attacker Infantry;
- surviving defender standing Infantry/Militia;
- consumed/surviving Transport IDs;
- captured bool.

The browser proof is intentionally small:

1. move combat Fleet + Transport to hostile Colony;
2. resolve/clear the orbital Battle when present;
3. attacker snapshot enters `invasion_decisions` and shows the pending Colony + eligible Transports;
4. choose **Invade** or **Decline**;
5. accepted invasion resolves immediately through the server;
6. capture refresh shows Colony under attacker, foreign cohorts conquered/origin-loyal and post-conquest strategic settlement materialized.

No tactical ground-combat screen is required.

## Gate-2 review refinements and acceptance record

Gate 2 reviewed the Gate-1 proposal against the live Slice-10/06/07 authority boundaries and the direct 1.31 disassembly. The following decisions are frozen for Gate 3:

1. **Core23 accepted.** Add `StrategicFleetSpecialTroopTransport`, `ColonyGroundForces.Infantry`, non-negative validation and Capital-owner validation. One extant Transport is always an implicit four-Infantry civilian special Fleet; no mutable cargo field. Economy ruleset schema remains **8** because the 100-PP fixed-support cost follows the existing hard-coded Colony/Outpost pattern rather than changing the external economy ruleset contract. Command/Event/API schema versions remain **1**.
2. **Transport identity accepted.** Base cost 100 PP, existing Feudal reduction path, no invented technology prerequisite, 1 CP at the next normal CP settlement, existing `empire.move_fleet`, existing fixed-special FTL derivation, and `CivilianFleetIDs` encounter participation. Friendly manual unload remains deferred.
3. **Invasion authority accepted with a one-opportunity Session continuation.** `PhaseInvasionDecisions` sits between encounter-wave continuation and post-resolution settlement. Only the attacker Seat may send revision-bound `invasion.invade` or `invasion.decline`; Sequence must be 1. One canonical opportunity is active at a time and a transient per-turn handled-key set prevents decline/failure from reoffering the same attacker/Colony pair. Independent opportunities may be processed sequentially. Same-Colony multi-attacker resolution remains deferred.
4. **Orbital-clear rule accepted and tightened.** The attacker must be at war, own a stationary combat Fleet and stationary Transport, have an authorized Seat, face no remaining modeled hostile combat Fleet in the target System, and must not bypass a defender-owned modeled Star Base/Battlestation/Star Fortress. This keeps station-only tactical defense explicitly unsupported instead of silently conquered.
5. **Ground-combat fixture accepted and corrected for Militia eligibility.** Attacker Infantry is `4 * selected Transport count`; defender standing Infantry is persisted; defender Militia is `floor(non-conquered organic Population / 5)`. In current MOOX that means `PopulationAssimilated` cohorts only. Infantry score 0, Militia score -10, one hit per supported unit, attacker `Intn(100)` then defender `Intn(100)`, and ties damage both. RNG commits only with the accepted atomic invade command.
6. **Failure/capture semantics accepted.** Failed invasion preserves Colony owner/Population/Construction, writes surviving defender standing Infantry, consumes every selected Transport and marks the pair handled. Capture performs original-style survivor packing, changes Colony owner, clears Construction, initializes attacker-origin cohorts assimilated/attacker-loyal and other cohorts conquered/origin-loyal under current Core semantics, reassigns a lost Capital to the defender's greatest-Population remaining Colony with lowest Colony-ID tie-break, assigns the captured Colony if attacker Capital is zero, and leaves war active. The pair is handled after success as well.
7. **Continuation/event model accepted.** Successful immediate command mutates State/Session exactly once, increments Session revision once and hosted `change_sequence` once. Event order is deterministic: `empire.invasion_resolved`, then `empire.colony_conquered` only on capture, followed by any later continuation events; decline emits `empire.invasion_declined`. No authoritative wall-clock timestamp and no separate opportunity event are required because opportunity is derivable Session projection. Blockade, Population transfers, Colony Economy/Population dynamics and Food rematerialize after all relevant invasion decisions; Treasury/CP settlement is **not** rerun in the same turn.

Transactional requirement: validation, Ground-Combat RNG use, State mutation, handled-key mutation, continuation derivation and any next BattleSession preparation all happen on candidates first. Any error leaves authoritative State, RNG, revision, events, pending opportunity, handled keys and host change sequence unchanged.

Gate 3 must implement only this frozen contract. The deferrals below are not implementation TODOs.
## Explicit Slice-11 v1 deferrals

The following are intentionally **not** Gate-3 TODOs:

- planetary bombardment and Bomb weapons;
- Bio-Weapons;
- Colony destruction instead of capture;
- Tanks/Armor, Battleoids and advanced troop types;
- full ground-combat technology matrix;
- racial/government/Warlord ground-combat modifiers;
- Ground Commander/Leader/Officer bonuses;
- exact original Barracks/Infantry capacity tables beyond the canonical fixture;
- friendly Transport unload/reinforcement UI;
- captured-technology rolls;
- capture-time special Building destruction/removal rules;
- Telepathic/mind-control immediate Population control;
- assimilation progression, occupation unrest and advanced rebellion;
- AI Transport/invasion planning and `Intercept_Invaders_` behavior;
- multiple simultaneous/multi-attacker invasion resolution;
- tactical ground-combat UI/animation;
- Empire elimination/victory declaration (Slice 12).

## Gate-1 conclusion

Gate 1 resolves all planned research questions without gameplay implementation.

The key architectural result is that Slice 11 should **not** bolt conquest onto the ordinary Planning CommandBatch or run it after `finishPostEncounter`. The original executable places invasion in the post-space-combat Colony-resolution path, and current MOOX already has the exact staged Encounter continuation point needed to insert a narrow `invasion_decisions` phase.

The fixed-special Fleet representation also remains sufficient: one extant Transport implicitly equals one four-Infantry support Ship, consumes 1 CP, moves through normal strategic transit and participates in Encounters only as a civilian asset. Variable cargo state is unnecessary because the original success cleanup preserves only complete four-Infantry Transports and moves the remainder into the captured Colony garrison.

Gate 2 accepted and froze the contract above with the recorded refinements. Gate 3 may now implement Core23 / Transport / invasion strictly within that boundary.

## Gate-3 implementation result - 2026-09-03

Gate 3 implemented the frozen Slice-11 v1 contract without expanding into the deferred ground-war families.

### Core23 and Troop Transport

- Core `StateSchemaVersion` is now **23**.
- `core.ColonyGroundForces.Infantry` persists standing Infantry and rejects negative values.
- A non-zero Empire Capital must reference a Colony currently owned by that Empire.
- `StrategicFleetSpecialTroopTransport = "troop_transport"` is the third fixed civilian support-Ship identity; it has no tactical `ShipIDs` and no mutable cargo field.
- Troop Transport construction is baseline-buildable with no invented Transport technology prerequisite, costs **100 PP** normally and **67 PP** under the existing Feudal 2/3-ceil support-Ship reduction, and completion uses the Empire's installed strategic drive.
- Each extant Transport contributes **1 Command Point** at the next ordinary CP/Treasury settlement, uses the existing fuel-range-aware `empire.move_fleet` path, and participates in strategic encounters only through `CivilianFleetIDs`.
- Core23 save/load tests cover standing Infantry, Transport identity and Capital-owner validation.

### Authoritative invasion boundary

- Session now includes `invasion_decisions` between prepared encounter-wave continuation and `post_resolution`.
- `game.Resolution` may return exactly one `InvasionOpportunity`; it may not return Encounters and an Invasion at the same boundary.
- The canonical opportunity remains the lowest `(SystemID, ColonyID, AttackerEmpireID)` legal candidate.
- Legal opportunity derivation enforces war, an authorized attacker Seat, stationary attacker combat Fleet plus stationary Troop Transport, no remaining modeled war-hostile combat Fleet, and no defender-owned modeled Star Base/Battlestation/Star Fortress in the target System.
- Same-Colony multi-attacker ambiguity remains deferred; such a Colony is not offered by the baseline.
- Session stores the pending opportunity and per-turn sorted handled `(ColonyID, AttackerEmpireID)` keys transiently. Decline, failure and capture all mark the pair handled, preventing a same-turn re-offer while allowing other independent opportunities to resolve sequentially.
- The handled set is cleared on normal turn completion.

### Immediate authority and atomicity

The existing endpoint remains authoritative:

`POST /api/v1/games/{gameID}/immediate-commands`

Implemented commands:

- `invasion.invade` with Colony ID plus a sorted-unique non-empty subset of the projected eligible Transport Fleet IDs;
- `invasion.decline` with the pending Colony ID.

Both retain Command/API schema 1, require command Sequence 1 and exact current `base_revision`, and derive the acting Empire from the authorized Seat. The Game resolver re-derives the current opportunity before resolution. Session performs State clone, Ground-Combat RNG work, handled-key mutation, continuation derivation and any next-Battle preparation entirely on candidates before committing. Rejected/stale/invalid commands leave State, RNG, Session revision, events, pending opportunity and handled keys unchanged.

A direct accepted Session invasion increments the Session revision exactly once. Through the hosted HTTP path, the same host mutation still has exactly one `change_sequence`; when invasion reaches `post_resolution` with no remaining immediate decision, the pre-existing host phase driver also performs the normal `CompleteTurn`, so the final hosted GameRevision may include both the invasion revision and the normal turn-advance revision. No extra host mutation or client authority is introduced.

### Ground Combat and conquest

The implemented narrow resolver matches the frozen original-1.31 fixture:

- attacker Infantry = `4 * selected Transport count`;
- defender standing Infantry = persisted `Colony.GroundForces.Infantry`;
- defender Militia = `floor(sum(PopulationAssimilated cohort Population) / 5)`; conquered cohorts do not generate Militia;
- Infantry score 0; Militia score -10;
- each supported unit takes one hit;
- each round consumes attacker `Intn(100)` then defender `Intn(100)`;
- `A <= D` damages attacker, `A >= D` damages defender, so ties damage both.

Failure preserves Colony ownership, Population and Construction, writes surviving defender standing Infantry, consumes every selected Transport, leaves unselected Transports intact and marks the pair handled.

Capture implements the original survivor-packing rule: reserve one attacker Infantry as garrison, process selected Transport IDs in reverse canonical order, preserve only complete remaining groups of four as full Transports, consume the other selected Transports and land the <4 remainder with the reserved Infantry. The Colony changes owner, Construction is cleared, attacker-origin cohorts become assimilated/attacker-loyal and other cohorts become conquered/origin-loyal. A captured defender Capital is replaced by the defender's greatest-Population remaining Colony with lowest-Colony-ID tie-break, or zero if none remains; a capital-less attacker adopts the captured Colony. Diplomatic war remains active.

After all relevant invasion decisions, the existing continuation rematerializes Blockades, Population transfers, Colony Economy/Population dynamics and Food logistics from the new ownership state. Treasury and Command-Point maintenance are **not** re-settled in the same turn.

### Server and browser proof

- Host dispatches invasion commands through the existing immediate-command transport and treats `invasion_decisions` as an interactive boundary.
- HTTP integration proves attacker-only projected opportunity, stale revision -> conflict, missing base revision -> bad request, accepted capture and the resulting authoritative snapshot.
- React/Vite exposes only the server-projected opportunity. The minimal Invasion panel can **Invade with all eligible Transports** or **Decline**; both use the current snapshot revision and reload the authoritative snapshot after success. The browser owns no Ground-Combat or Colony-ownership state.

### Deterministic regression proof

Gate-3 tests now cover:

- Core23 standing-Infantry/Transport save-load and invalid-state validation;
- 100-PP / Feudal-67-PP Transport build choice, queue, completion, 1-CP identity and strategic movement;
- opportunity allowed/blocked by orbital conditions;
- conquered-Population Militia exclusion;
- deterministic success and failure casualties;
- selected-only Transport destruction and successful full-four Transport survivor packing;
- captured cohort, Construction, Capital and war state;
- byte-identical Core23 save/load after conquest;
- identical seed/state/command invasion replay with byte-identical serialized State and DeepEqual event stream;
- invalid unsorted Transport IDs as State/RNG atomic no-op;
- Session stale-revision no-op, attacker-only projection, decline/failure no-reoffer and exactly-one direct Session invasion revision;
- full real Session flow from neutral relation -> `diplomacy.declare_war` -> ordinary combat/Transport transit arrival -> `invasion_decisions` -> Colony capture;
- HTTP immediate-command capture and browser build.

The current Core23 canonical New-Game seed `0x8009` full-state SHA-256 is `83614740b216409877b03c536a16328c7fa7031867f58dbeb70857a8953f5762`. Historical Core21 and Core22 fingerprints remain preserved in the New-Game evidence.

Final Gate-3 QA passed:

- `go test ./... -count=1`;
- `go vet ./...`;
- `npm --prefix web run build`;
- `git diff --check`;
- current-code schema-22 assertion scan: zero;
- exactly one Slice-11 `_OPEN_` marker remains.

Gate 3 is complete. Gate 4 owns independent final QA, HISTORY/status closure, commits and removal of the OPEN marker. No push has been performed.


## Gate-4 independent QA result - 2026-09-03

Gate 4 independently re-audited the frozen Slice-11 contract rather than relying only on Gate-3 results.

Additional Gate-4 regression proof was added for two closure-sensitive invariants:

- captured Colony ownership, standing Infantry, cleared Construction, exact canonical cohort states, attacker Capital adoption, defender replacement-Capital selection and continuing war stance all survive a Core23 save/load round-trip exactly;
- transient invasion handled keys clear on normal `CompleteTurn`, while capture continuation leaves already-settled Treasury and Command-Point state unchanged, proving there is no second same-turn Treasury/CP settlement after conquest.

The new round-trip QA initially exposed only fixture/test assumptions: `Population.Normalize()` canonically sorts cohorts, and the defender fixture owns another Colony, so a captured Capital correctly moves to that remaining Colony rather than to zero. The engine behavior matched the frozen contract; the QA expectations were corrected to the canonical order/replacement-capital rule.

Independent Gate-4 execution passed:

- focused Core/Game/Session/App/Server Transport, Invasion, Encounter, Command-Point, Population, Economy, Diplomacy and Tactical/Strategic Encounter regressions, repeated three times;
- focused captured ownership/cohort save-load, deterministic invasion replay and success/failure Ground-Combat regressions, repeated three times;
- focused Session capture/failure, declared-war transit-arrival capture and handled-key-reset regressions, repeated three times;
- `go test ./... -count=1`;
- `go vet ./...`;
- `npm --prefix web run build`;
- `git diff --check`.

Code audit reconfirmed candidate-only State/RNG/handled/continuation preparation before authoritative Session commit, assimilated-only Militia, station/remaining-hostile blockers, no attack-authority bypass and no second Treasury/CP settlement in `finishPostEncounter`.

Gate 4 QA is complete. The remaining closure step is documentation/HISTORY synchronization, implementation/evidence commit, OPEN-marker removal and final clean-repository verification. No push is part of this slice closure.
