# Slice 10 research - Diplomacy, war and peace baseline

Status: **Gate 3 complete; Gate 4 QA/closure ready**
Opened: 2026-09-02
Starting HEAD: `9b2641d` (`docs: close deterministic new game slice`)

## Objective

Define the first authoritative deterministic diplomacy lifecycle needed to move the now-generated two-Empire game from neutral relations into intentional war and back to peace, while reusing the existing strategic hostility/blockade/encounter/tactical authority instead of creating a parallel combat path.

This Gate-1 result deliberately separates direct MOO2 1.31 evidence from the narrow MOOX contract proposed for Gate 2. No diplomacy gameplay implementation is part of Gate 1.

## Gate-1 repository checkup

At Slice-10 open:

- branch `main` was clean at HEAD `9b2641d`;
- Slice 09 was closed;
- no prior `_OPEN_*.md` marker existed;
- exactly one Slice-10 marker was created at `docs/slices/_OPEN_DIPLOMACY_WAR_PEACE_BASELINE_2026-09-02.md`;
- Core `StateSchemaVersion` remains 21;
- Command schema remains 1 and Event schema remains 1;
- focused current baseline tests pass:
  - `go test ./internal/core -run 'Diplomatic|Strategic' -count=1`;
  - `go test ./internal/game -run 'Blockade|Encounter|CombatFleet' -count=1`;
  - `go test ./internal/session -run 'StrategicBlockade|StrategicEncounter' -count=1`.

## Current MOOX diplomacy / hostility baseline

### Persisted relation shape

Current Core already contains:

```go
type DiplomaticStance string

const (
    DiplomaticStanceNeutral DiplomaticStance = "neutral"
    DiplomaticStanceHostile DiplomaticStance = "hostile"
)

type DiplomaticRelation struct {
    FromEmpireID ID
    ToEmpireID   ID
    Stance       DiplomaticStance
}
```

`GameState.DiplomaticRelations` is sorted by `(FromEmpireID, ToEmpireID)`. `DiplomaticStanceBetween(from,to)` is directional and returns `neutral` for missing/self relations.

Current validation already requires:

- non-zero, distinct Empire IDs;
- referenced Empires must exist;
- only the current `neutral|hostile` values;
- strict ascending `(from,to)` ordering.

### Current attack / blockade consumers

Current MOOX has one existing hostility source and two important consumers:

- `internal/game/blockade.go` checks `DiplomaticStanceBetween(fleet.EmpireID, targetEmpireID)` directionally;
- `internal/game/encounter.go` checks `DiplomaticStanceBetween(attackerID, defenderID)` directionally before civilian overrun or combat encounter creation;
- tactical BattleSession behavior is downstream of an already-created strategic encounter and does not independently decide diplomacy.

Therefore Slice 10 should evolve the existing relation source rather than add a second combat-authorization model.

Player projection currently exposes own Empire/Colonies and public seat data but **does not expose diplomatic relations**. Observer projection already includes the complete cloned `GameState`, including `DiplomaticRelations`.

### Current phase / transport constraints

`GameSession` phases are:

1. `planning`;
2. `strategic_resolution`;
3. `encounters`;
4. `post_resolution`;
5. back to `planning`.

The existing `/api/v1/games/{gameID}/immediate-commands` route is currently specialized to the post-resolution Colony Base command path. Its request is not revision-bound. Slice 10 needs a revision-bound planning immediate-command path if diplomacy is to behave like an interactive authoritative action rather than a delayed End-Turn command.

## Original MOO2 1.31 executable anchors

The committed reference executable is:

`reference/original/support/files/Orion2.exe`

Embedded debug symbols identify the relevant routines:

| VA | Symbol |
| --- | --- |
| `0x179F4` | `Diplomacy_Declare_War_Resolution_` |
| `0x17D5B` | `Diplomacy_Propose_Treaty_` |
| `0x1AFA0` | `Show_Sneak_Attack_Message_` |
| `0x4D78E` | `Init_Diplomatic_Relations_` |
| `0x4E3B5` | `Change_Relations_` |
| `0x4F59B` | `Sneak_Attack_Or_Declare_War_` |
| `0x51078` | `Declare_War_` |
| `0x5138E` | `Break_Treaties_` |
| `0x5232E` | `Start_Treaty_` |
| `0x524FB` | `Declare_Peace_` |
| `0xDB257` | `Launch_Sneak_Attack_` |
| `0xD8DE1` | `Player_Is_Hostile_To_Player_` |
| `0xE8231` | `Player_May_Attack_Player_` |
| `0xEF807` | `Add_Peace_Msg_` |
| `0xEF81F` | `Add_Peace_Ended_Msg_` |

These addresses are from object-1 virtual addresses used by the existing read-only research disassembler.

## Original relation-state evidence

### Directional storage

Prior blockade evidence already proved that the player record contains a per-target relation byte at `player + 0x627`.

`Compute_Blockades_` VA `0xE5097` reads `player[fleetOwner].relation[target]` directly. Only values `4..6` qualify as hostile for a normal-Empire blockade.

`Player_Is_Hostile_To_Player_` VA `0xD8DE1` independently begins with the same directional `player[from].relation[to]` lookup and recognizes values `4..6` as hostile.

So original storage is directional even though the selected normal war/peace transition routines write both directions symmetrically.

### Minimum relation values resolved for Slice 10

Direct executable evidence now resolves the minimum state path:

- relation `0`: no breakable treaty / neutral baseline for this slice. `Break_Treaties_` immediately exits when the relation is `0`;
- relation `1`: an ordinary treaty state handled by `Start_Treaty_`/`Break_Treaties_`; exact broader treaty semantics are deferred here;
- relation `2`: another ordinary treaty state handled by `Start_Treaty_`/`Break_Treaties_`; exact broader treaty semantics are deferred here;
- relation `3`: **peace**. `Declare_Peace_` writes `3` in both directions;
- relation `4`: **normal-Empire war**. `Declare_War_` writes `4` in both directions on the normal-player path;
- relation `5..6`: hostile special-case values selected by branches inside `Declare_War_`; they are not required for the frozen two-normal-Empire Slice-10 fixture.

`Break_Treaties_` further distinguishes relations `1`, `2` and `3` with treaty-break penalties and refuses the treaty-break path once the reciprocal relation is `>=4`. This reinforces the treaty/peace versus war boundary without requiring Slice 10 to implement the deferred treaty families.

### War transition is symmetric for normal Empires

`Declare_War_` VA `0x51078`:

1. checks the existing directed relation and exits the ordinary path if it is already `>=4`;
2. clears/breaks trade, research, tribute and treaty state through `Break_Trade_`, `Break_Research_`, `Break_Tribute_` and `Break_Treaties_`;
3. on the ordinary two-normal-Empire path writes relation value `4` from A -> B **and** B -> A;
4. updates diplomacy modifiers afterwards.

The special branches can select `5` or `6`, but Slice 10 does not need those branches.

### Peace transition is symmetric and agreement-gated

`Declare_Peace_` VA `0x524FB` directly writes relation value `3` from A -> B and B -> A and updates only diplomacy relation/modifier fields in that routine.

The human proposal flow in `Diplomacy_Propose_Treaty_` calls `Diplomacy_Determine_Treaty_Proposal_`; only the successful proposal branch (`result == 1` in the observed branch) calls `Declare_Peace_`.

Therefore the minimum semantic is:

- war declaration is unilateral;
- returning from war to peace requires acceptance/agreement rather than one side unilaterally forcing peace.

The original synchronous diplomacy UI does not provide a directly reusable network-persistent offer object, so MOOX needs a small explicit pending-offer representation for asynchronous/server-authoritative use.

## Original timing evidence

### Human war declaration takes effect immediately

`Diplomacy_Declare_War_Resolution_` VA `0x179F4` performs its diplomacy UI/message work and then directly calls:

```text
0x17B36 call 0x51078 ; Declare_War_
0x17B3F call 0x524C3 ; Adjust_Diplomat_Modifiers_
```

The relation mutation is therefore part of the interactive diplomacy resolution itself, not deferred into a later generic strategic-turn batch.

### Accepted peace takes effect immediately

The successful peace-proposal branch inside `Diplomacy_Propose_Treaty_` directly calls `Declare_Peace_` after proposal resolution. Again, the relation mutation is immediate once agreement is reached.

### Fleet orders are not cancelled by these transition routines

The directly inspected `Declare_War_` and `Declare_Peace_` bodies mutate diplomacy/treaty/modifier fields. Neither routine calls strategic ship-movement cancellation, fleet relocation or tactical battle teardown routines.

Therefore a relation transition does **not** erase already-issued fleet movement merely because war/peace changed. Downstream authorization determines whether those fleets fight/blockade when the strategic boundary is reached.

Exact original behavior for an already-instantiated tactical battle object after a later peace agreement is not proven by these routines. MOOX has a cleaner phase boundary: diplomacy is accepted only during `planning`, so once `encounters` has materialized a BattleSession, that battle is already committed and diplomacy cannot retroactively cancel it in Slice 10.

## Original attack authorization versus sneak attack

`Player_May_Attack_Player_` VA `0xE8231` is decisive:

- relation values `1..3` do **not** authorize ordinary attack;
- relation values `4..6` do authorize attack;
- relation `0` enters a separate path that checks additional flags/player properties and, on one observed branch, a `Random_(10)` result.

This means neutral/no-treaty attack permission is not equivalent to formal war. It belongs to the original sneak-attack/AI decision family.

The executable also exposes separate `Sneak_Attack_Or_Declare_War_`, `Show_Sneak_Attack_Message_`, `Sneak_Attack_War_Check_`, `NPC_Sneak_Attacks_`, `Sneak_Attack_Evaluations_` and `Launch_Sneak_Attack_` routines.

`Launch_Sneak_Attack_` eventually has a path that calls `Declare_War_`, confirming that sneak-attack planning/launch and the formal war relation transition are separate concepts even when one later leads to the other.

### Slice-10 inclusion decision

For the narrow normal-Empire baseline:

- **war** authorizes strategic attack/blockade/encounter;
- **neutral** does not;
- **peace** does not;
- neutral sneak attacks, AI chance/flags and implicit attack-triggered war declaration are deferred.

A player that wants to attack a neutral/peace Empire must first use the explicit authoritative declare-war command.

## Current strategic timing mapped onto Slice 10

Current `EconomyResolver.Resolve` applies submitted strategic commands, materializes economy/research/population, advances Fleet transit and Construction, then calls `prepareEncounterBoundary`. After encounter waves are exhausted, `finishPostEncounter` recomputes system blockades.

The proposed MOOX timing is therefore:

1. diplomacy commands mutate authoritative relation state immediately during `PhasePlanning`;
2. existing Fleet movement orders/submitted commands are not cancelled;
3. when strategic resolution later reaches `prepareEncounterBoundary`, the **current** war/peace state is the single encounter authorization input;
4. after encounters, existing blockade recomputation consumes the same current relation state;
5. `BlockadedEmpireIDs` is not eagerly rewritten by the diplomacy command itself; it remains a derived strategic value updated at the existing authoritative recomputation boundary;
6. once `PhaseEncounters` begins and BattleSessions exist, diplomacy commands are rejected until the game returns to planning, so created battles remain committed.

Consequences for same-turn tests:

- neutral fleets may already share a System without battle;
- declaring war during planning makes that same-system pair eligible when the upcoming encounter boundary is derived;
- accepting peace during planning before resolution suppresses creation of a new hostile encounter for that pair, but existing Fleet transit/orders remain;
- peace cannot be accepted during an already-running BattleSession in this baseline.

## Gate-2 proposed contract

This section is the exact contract to accept/freeze in Gate 2 before implementation.

### 1. Core schema

Advance Core `StateSchemaVersion` from **21 -> 22**.

Retain the existing directional storage shape but resolve the selected semantic values:

```go
type DiplomaticStance string

const (
    DiplomaticStanceNeutral DiplomaticStance = "neutral"
    DiplomaticStancePeace   DiplomaticStance = "peace"
    DiplomaticStanceWar     DiplomaticStance = "war"
)

type DiplomaticRelation struct {
    FromEmpireID ID               `json:"from_empire_id"`
    ToEmpireID   ID               `json:"to_empire_id"`
    Stance       DiplomaticStance `json:"stance"`
}

type DiplomaticPeaceOffer struct {
    FromEmpireID ID `json:"from_empire_id"`
    ToEmpireID   ID `json:"to_empire_id"`
}
```

Add:

```go
DiplomaticPeaceOffers []DiplomaticPeaceOffer `json:"diplomatic_peace_offers,omitempty"`
```

to `GameState`.

Canonical invariants:

- missing relation row means `neutral`;
- canonical persisted relation rows are only `peace` or `war`; explicit `neutral` rows are omitted/rejected;
- normal-Empire `peace` and `war` must exist as a reciprocal pair with identical stance;
- relation rows remain strictly sorted by `(from,to)`;
- peace offers are directional, strictly sorted/unique by `(from,to)`, non-self, reference existing Empires and are legal only while that pair is at `war`;
- at most one pending direction exists for a pair; successful peace acceptance removes any pair offers;
- New Game starts with no relation rows and no offers, therefore neutral.

`DiplomaticStanceBetween(from,to)` remains the authoritative semantic accessor. Existing generic `DiplomaticStanceHostile` usage is replaced by `DiplomaticStanceWar` for the normal-Empire path.

### 2. Attack authorization

Add/use one semantic helper for normal Empire combat authorization, conceptually:

```go
MayAttackEmpire(fromEmpireID, toEmpireID) bool
```

Slice-10 mapping:

- `war` -> true;
- `neutral` -> false;
- `peace` -> false.

`blockade.go` and `encounter.go` must consume this same helper/state. No UI, server or tactical code may maintain a second hostility flag.

Sneak attack / relation-0 exceptions are explicitly deferred.

### 3. Planning diplomacy commands

Add three command kinds, keeping `protocol.CommandSchemaVersion = 1`:

```text
diplomacy.declare_war
diplomacy.offer_peace
diplomacy.accept_peace
```

Payloads:

```json
{"target_empire_id": 3}
```

for `declare_war` and `offer_peace`, and:

```json
{"from_empire_id": 3}
```

for `accept_peace`.

All payloads use the existing strict JSON command-decoding convention: unknown fields/trailing JSON rejected.

Authority and legality:

- acting Empire is derived from `SeatID`; client cannot supply actor Empire ID;
- target/source must be a different existing normal Empire;
- commands are accepted only in `PhasePlanning`;
- **no seat may already have submitted the current End-Turn batch**; diplomacy closes for the turn at the first accepted turn submission;
- each immediate diplomacy request contains exactly one command with `Sequence == 1`;
- request must carry the exact current `base_revision`; stale revisions are rejected transactionally;
- rejected commands do not mutate state/revision/events/offers.

Command transitions:

**declare_war**

- legal from `neutral` or `peace`;
- illegal if already `war`;
- set A -> B and B -> A to `war` atomically;
- clear any pending peace offer for that pair;
- do not cancel Fleet orders or directly edit `BlockadedEmpireIDs`.

**offer_peace**

- legal only while pair is `war`;
- illegal duplicate if the same outgoing offer already exists;
- if an incoming B -> A offer already exists, the actor must use `accept_peace` rather than create a reciprocal offer;
- add directed offer A -> B;
- relation remains `war`.

**accept_peace**

- legal only while pair is `war` and an incoming B -> A offer exists;
- atomically set both directed relations to `peace`;
- clear pending pair offers;
- do not cancel Fleet orders or directly edit `BlockadedEmpireIDs`.

No reject/withdraw command is required in v1; an unaccepted offer has no attack-authorization effect and can remain pending across turns until the relation changes or it is accepted. A later slice may add richer proposal lifecycle. Slice 10 also does **not** implement the original automatic peace-ended/peace-duration behavior: MOOX `peace` remains stable until an explicit later war declaration.

### 4. Immediate transport / revision contract

Keep the existing endpoint name:

```text
POST /api/v1/games/{gameID}/immediate-commands
```

but split/upgrade its request contract to be revision-bound:

```json
{
  "schema_version": 1,
  "seat_id": 1,
  "base_revision": 7,
  "command": {
    "schema_version": 1,
    "sequence": 1,
    "kind": "diplomacy.declare_war",
    "payload": {"target_empire_id": 3}
  }
}
```

The existing post-resolution Colony Base immediate command is migrated to the same revision-bound request shape. Battle command transport remains separate.

Session/app authority should become a general immediate-command dispatcher rather than a Colony-Base-named Host entry point:

- planning diplomacy commands -> diplomacy resolver;
- post-resolution Colony Base command -> existing Colony Base resolver;
- wrong phase/kind -> reject.

A successful immediate command increments GameSession revision once and hosted `change_sequence` once, then normal WebSocket invalidation tells clients to resync over HTTP.

### 5. Events / replay audit

Keep Event schema version **1** and add deterministic strategic event kinds:

```text
diplomacy.war_declared
diplomacy.peace_offered
diplomacy.peace_accepted
```

Minimum payloads:

- `war_declared`: actor Empire, target Empire, previous stance;
- `peace_offered`: from Empire, to Empire;
- `peace_accepted`: accepting Empire, offering Empire, previous stance (`war`), current stance (`peace`).

Event `SeatID` is the acting seat and `CommandSequence` is `1` for the immediate request. State mutation, event append and revision increment are atomic.

Save/load/replay tests must include both relation rows and pending peace offers. No wall-clock timestamps enter authoritative state.

### 6. Player / Observer projection

Observer projection remains the detached complete `GameState` and therefore sees all relation rows/offers.

Add a compact player-safe diplomacy projection, one row per other Empire, conceptually:

```go
type DiplomacyView struct {
    OtherEmpireID       core.ID                `json:"other_empire_id"`
    Stance              core.DiplomaticStance `json:"stance"`
    IncomingPeaceOffer  bool                   `json:"incoming_peace_offer,omitempty"`
    OutgoingPeaceOffer  bool                   `json:"outgoing_peace_offer,omitempty"`
}
```

`PlayerView` exposes only diplomacy involving the player's own Empire. It does not reveal unrelated future Empire-to-Empire diplomacy.

Web proof for the current two-local-seat New Game fixture:

- select Human seat -> declare war on Darlok;
- snapshot refresh shows `war`;
- offer peace;
- select Darlok seat -> incoming offer visible -> accept;
- both seat snapshots show `peace`.

### 7. Same-system / movement / BattleSession behavior

Freeze these tests:

1. neutral opposing combat Fleets in one System -> no encounter;
2. same state, declare war during planning -> next strategic encounter boundary creates the existing Slice-06/07 encounter/BattleSession path;
3. war + hostile movement already ordered, peace accepted before strategic resolution -> movement remains, but no new hostile encounter is created for that now-peace pair;
4. war encounter already materialized in `PhaseEncounters` -> diplomacy command rejected; BattleSession remains committed;
5. post-battle blockade recomputation uses the same current war/peace state;
6. existing tactical behavior under `war` is byte/semantic-regression compatible except the renamed persisted stance/schema version.

### 8. Explicit deferrals

Still deferred after Slice 10:

- original relation values `1/2` treaty families as gameplay features;
- trade/research treaties;
- alliance / non-aggression UI and lifecycle;
- tribute, gifts, demands, threats;
- diplomatic AI/personality and original relation-score modifiers;
- contact/discovery gating;
- neutral sneak attacks and attack-triggered implicit war declaration;
- spies/espionage;
- Galactic Council diplomacy;
- race/leader diplomacy bonuses;
- exact original special hostile values `5/6`;
- exact original peace duration/peace-ended side mechanics not required for the accepted bilateral baseline.


## Gate-2 decisions - accepted 2026-09-03

Gate 2 reviewed the complete proposal against the current Session revision/submission model and accepts the contract with one deliberate tightening: planning diplomacy is available only **before the first End-Turn submission of the current turn**. This avoids mixed-revision submitted batches and prevents submission-state timing from becoming hidden gameplay authority.

### Decision 1 - diplomatic state schema and invariants: accepted

Freeze Core schema **22** for Slice 10.

- Missing relation row means `neutral`.
- Persisted normal-Empire relation rows are `peace` or `war`; explicit neutral rows are not canonical.
- `peace` and `war` are written and validated as reciprocal symmetric pairs for the normal-Empire baseline, while storage remains directionally keyed for future original-compatible expansion.
- `DiplomaticPeaceOffer` remains directional, sorted and unique, legal only while the pair is at war.
- Only one pending offer direction may exist per Empire pair. If B -> A already exists, A uses `accept_peace`; A -> B is not added as a reciprocal duplicate.
- New Game has no relation rows/offers and therefore begins neutral.
- Slice-10 `peace` does not auto-expire. Original peace-duration / `peace ended` mechanics remain deferred.

### Decision 2 - declare-war / peace command authority: accepted

Freeze Command schema **1** and these immediate command kinds:

```text
diplomacy.declare_war
diplomacy.offer_peace
diplomacy.accept_peace
```

Freeze strict one-field payloads from the Gate-1 proposal. The acting Empire is always derived from authenticated/authorized `SeatID`; actor Empire ID is never client-authoritative. Each immediate request contains exactly one command with `Sequence == 1` and exact current `base_revision`.

- `declare_war`: unilateral, from neutral/peace only, atomically writes reciprocal war and clears pair offers.
- `offer_peace`: war-only, creates one directed pending offer, does not alter attack authorization.
- `accept_peace`: requires an incoming offer, atomically writes reciprocal peace and clears pair offers.
- Invalid/stale/duplicate/wrong-authority requests are transactional no-ops.

### Decision 3 - turn/phase boundary: accepted with tightened submission rule

Diplomacy immediate commands are legal only when all of the following hold:

1. Session phase is `planning`;
2. request `base_revision` equals current Session revision;
3. **zero seats have submitted the current turn**.

The first accepted End-Turn submission closes diplomacy for that turn. Diplomacy becomes available again when the Session returns to the next `planning` phase.

This is intentionally stricter than the Gate-1 draft's actor-only submission check because current `SubmitTurn` validates `BaseRevision` only when a batch is accepted and then stores that batch. Allowing a later immediate mutation would otherwise leave already-stored and later-submitted batches based on different revisions.

Relation changes do not cancel existing persistent Fleet transit/orders from prior state. They only change authorization consumed at the next strategic encounter/blockade derivation boundary. Once `PhaseEncounters` begins, created BattleSessions are committed and diplomacy cannot retroactively cancel them.

### Decision 4 - attack authorization: accepted

Freeze one authoritative normal-Empire semantic helper (`MayAttackEmpire` or equivalent):

- `war` -> attack authorized;
- `neutral` -> not authorized;
- `peace` -> not authorized.

Both blockade derivation and strategic encounter generation must consume that same helper/state. No UI/server/tactical shadow hostility flag is allowed. Neutral sneak attacks, original relation-0 exceptions and special hostile values `5/6` remain deferred.

### Decision 5 - event/history/projection model: accepted

Keep Event schema **1** and freeze the three deterministic strategic event kinds from Gate 1:

```text
diplomacy.war_declared
diplomacy.peace_offered
diplomacy.peace_accepted
```

Events carry only deterministic authority/state transition data; no wall-clock timestamp enters authoritative state. Relation rows and pending peace offers are persisted in `GameState`; events are the deterministic audit/projection history, not a second diplomacy state store.

Observer projection continues to receive the detached complete `GameState`. Player projection adds only diplomacy involving that player's Empire: other Empire ID, stance, incoming offer and outgoing offer. Unrelated Empire-to-Empire diplomacy is not exposed through PlayerView.

The existing `/api/v1/games/{gameID}/immediate-commands` endpoint is retained and made revision-bound/general enough to dispatch planning diplomacy and the existing post-resolution Colony Base immediate command by phase/kind. Battle command transport remains separate.

### Decision 6 - deferred families: frozen

Slice 10 explicitly does not implement:

- relation `1/2` treaty families as gameplay;
- trade/research treaties, alliances/non-aggression, tribute/gifts/demands/threats;
- diplomatic AI/personality or original relation-score modifiers;
- contact/discovery gating;
- neutral sneak attacks or attack-triggered implicit war declaration;
- special original hostile values `5/6`;
- automatic original peace-duration / peace-ended behavior;
- spies/espionage, Galactic Council diplomacy or race/leader diplomacy bonuses.

These deferrals are not implementation TODOs inside Gate 3; they are out of Slice-10 scope.

### Gate-3 entry contract

Gate 3 may now implement only the accepted contract above. Its required proof remains:

- Core22 validation/roundtrip for relations and pending offers;
- revision-bound pre-submission planning diplomacy commands;
- deterministic war/offer/accept events;
- shared war-only attack authorization for blockade/encounter derivation;
- player/observer/API/web projections;
- neutral same-system -> declare war -> battle path;
- war movement/state -> accepted peace before turn submission/resolution -> movement retained, no new hostile encounter;
- diplomacy rejection after first turn submission and during an already-materialized encounter;
- preservation of Slice-06/07 tactical behavior when war is active.

Gate 2 is **6/6 complete**. No diplomacy gameplay code was changed during Gate 2.

## Gate-1 conclusion

Gate 1 is complete enough to make an implementation decision without guessing the war/peace core:

- original normal war is relation `4`, symmetric on declaration;
- original peace is relation `3`, symmetric after an accepted peace proposal;
- ordinary attack authorization is war-only (`4..6`), while neutral sneak attack is a distinct path;
- war/peace mutations are immediate diplomacy actions and do not cancel Fleet movement;
- current MOOX directed relation storage and downstream blockade/encounter authority can be evolved rather than replaced;
- the remaining choices are now explicit Gate-2 contract choices, not unresolved reverse-engineering blockers.

## Gate-3 implementation result - 2026-09-03

Gate 3 implemented the frozen Gate-2 contract without adding deferred diplomacy families.

### Authoritative state and attack mapping

Core `StateSchemaVersion` is now **22**. `GameState` persists sorted `DiplomaticRelations` and `DiplomaticPeaceOffers`. Canonical normal-Empire relation rows are reciprocal `peace` or `war`; missing rows are neutral. Validation rejects explicit neutral rows, asymmetric peace/war, offers outside war, unsorted/duplicate state and reciprocal simultaneous offers.

`GameState.MayAttackEmpire(from,to)` returns true only for `war`. Existing blockade and strategic encounter derivation now consume this helper, removing the old direct `hostile` comparison as a second semantic surface.

The schema-only New Game serialized fingerprint changed for canonical seed `0x8009` from the historical Slice-09 Core21 hash to current Core22 SHA-256 `8effff679109cd10a427f1c80b83047dc0d2fa8e5e4871a67235a486dbc2fc68`; deterministic generation semantics are otherwise unchanged and the golden regression remains pinned.

### Commands, events and session authority

`internal/game/diplomacy.go` implements:

- `diplomacy.declare_war` -> reciprocal war, pair offers cleared, `diplomacy.war_declared`;
- `diplomacy.offer_peace` -> directional pending offer while war remains active, `diplomacy.peace_offered`;
- `diplomacy.accept_peace` -> requires incoming offer, reciprocal peace and pair-offer removal, `diplomacy.peace_accepted`.

All payloads use the existing strict JSON decoder. Acting Empire comes from server Seat authority. Session immediate commands require sequence `1` and exact current revision. Diplomacy additionally requires `planning` and **zero accepted turn submissions**; the first submission closes diplomacy for the turn. Rejected commands are transactional no-ops.

The existing Colony Base immediate path now uses the same revision-bound request envelope. Battle command validation remains a separate transport path and is not made dependent on `base_revision`.

### Projection and web proof

`PlayerView` now contains one `DiplomacyView` per other Empire with stance and incoming/outgoing peace-offer booleans; it does not expose unrelated Empire-pair diplomacy. Observer continues to receive the detached complete GameState.

The React/Vite HMI consumes this projection and the revision-bound `/immediate-commands` endpoint. It renders the relation and offers exactly one appropriate action for the minimal lifecycle: Declare war from neutral/peace, Offer peace during war, or Accept peace when an incoming offer exists. Controls are disabled outside planning or once any seat has submitted.

A real HTTP server regression executes Human declare-war -> Human peace-offer -> Darlok accept-peace, verifies player snapshots, and also proves stale revision is rejected with conflict and missing `base_revision` with bad request.

### Strategic timing / Battle commitment proof

Regression coverage now proves:

1. opposing combat Fleets sharing a System while neutral create no encounter;
2. declaring war before the next boundary causes that same state to create the established encounter path;
3. accepting peace before the boundary leaves both Fleets in place but suppresses a new hostile encounter;
4. once tactical encounter/BattleSession state is materialized, diplomacy is rejected and revision/Battle/pending-offer state remains unchanged;
5. legacy Slice-06/07 tactical/encounter behavior remains green under reciprocal `war` fixtures;
6. replaying the identical declare-war -> offer-peace -> accept-peace command stream from identical starting state produces byte-identical serialized Core22 state and DeepEqual deterministic event sequences.

### Gate-3 verification

Passed after implementation:

```text
go test ./... -count=1
go vet ./...
npm --prefix web run build
git diff --check
```

Gate 3 is **6/6 complete**. Gate 4 owns the independent final QA, HISTORY/status closure, commit and removal of the single `_OPEN_` marker. No push has been performed.
## Gate-4 independent QA - 2026-09-03

Gate 4 independently repeated and extended the Gate-3 proof before any commit.

- Focused Core/Game/Session/Server diplomacy, blockade, encounter, tactical-battle and New-Game-golden coverage passed **5 consecutive runs**.
- Deterministic diplomacy lifecycle replay passed **10 consecutive runs**, producing byte-identical serialized Core22 state and DeepEqual deterministic event sequences from identical starting state/commands.
- Save/load coverage confirms reciprocal `war|peace` state and directional pending peace offers survive Core22 marshal/unmarshal exactly; invalid neutral rows, asymmetric relation state and invalid offers remain rejected.
- A production authority scan found **0** unexpected state touches outside the allowed Core state/query/validation and `internal/game/diplomacy.go` resolver surfaces. The only production diplomacy resolver chain is `internal/server` -> `internal/app` -> `internal/session` -> `internal/game`, and acting Empire remains Seat-derived.
- The React client has exactly one direct `fetch()` call, inside the shared `requestJSON` transport helper. The Diplomacy panel reads projected stance/offer fields and submits server commands; it does not mutate authoritative relations locally.
- Full repository QA passed again: `go test ./... -count=1`, `go vet ./...`, `npm --prefix web run build`, `git diff --check`.

No Gate-4 defect was found. The implementation is ready to commit; final Slice closure remains documentation/HISTORY/marker work only.