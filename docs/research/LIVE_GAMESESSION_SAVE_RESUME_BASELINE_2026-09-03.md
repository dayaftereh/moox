# Live GameSession save / resume baseline - Gate 1 evidence

Date: **2026-09-03**

Slice: **14 - Live GameSession save / resume baseline**

Status: **Gate 1 complete; Gate 2 contract proposed; no persistence implementation yet**.

## Gate 1 conclusion

MOOX already has the essential deterministic payload needed for live persistence, but the current persistence surface stops at two narrower layers:

- Core schema **23** serializes the strategic `core.GameState`, including `Seed`, strategic `RNGState`, `Turn`, `NextID`, galaxy, Empires, Colonies, Outposts, ShipDesigns, Ships, StrategicFleets, Diplomacy, PopulationTransfers and Core event history.
- Slice 12 adds `CompletedSnapshotSchemaVersion = 1`, but only for terminal `session.PhaseCompleted` sessions.

A live save must therefore preserve the **session state above Core23** and the **battle state below the session phase**, while deliberately excluding transport/process objects that can be reconstructed after load.

The correct Slice-14 boundary is:

> Persist every authoritative value that can change future game/result/event bytes; re-inject stateless rules/resolver dependencies; do not persist sockets, subscriber channels, process-local locks or host delivery counters.

The current built-in AI is compatible with this design because Slice 13 `baseline_v1` is a pure deterministic policy over `PlayerDecisionView`/`DecisionCatalog`. It has no private mutable AI memory or RNG to serialize.

## Existing persistence and strictness

### Core23

`internal/core/state.go` defines `StateSchemaVersion = 23`.

`core.GameState` already owns:

- `SchemaVersion`;
- `Seed`;
- `RNGState`;
- `Turn`;
- `NextID`;
- complete strategic entities/relations/transfers;
- Core event history.

`internal/core/serialization.go` provides strict `DisallowUnknownFields()` decode plus `GameState.Validate()`.

Slice 14 should embed a typed Core23 `GameState` in the live save rather than introduce a second copy of strategic fields.

### Completed-session snapshot

`internal/session/completed_snapshot.go` currently stores:

- schema version;
- game ID;
- revision;
- completed phase;
- Core23 state;
- Seats and optional submissions;
- Session domain events;
- draft telemetry;
- eliminated Empire IDs;
- immutable Result.

On restore it reuses `NewGameSession`, validates the completed result boundary and reconstructs the next Session event/telemetry sequence from history.

This is a useful validation/reference implementation, but it does **not** contain the live continuation fields required by Encounters/Invasion/PostResolution and it intentionally rejects non-completed phases.

Slice 14 should keep the existing completed-snapshot format/API valid. The live format is a **separate schema**. A completed session may be exported as a live save, and a restored completed live save must still produce the same existing completed-snapshot bytes.

## Mutable GameSession state inventory

Current `GameSession` determinism-relevant fields are:

- `gameID string`;
- `revision uint64`;
- `phase Phase`;
- `state *core.GameState`;
- `seats []seatState`:
  - immutable Seat identity/controller metadata;
  - optional revision-bound `protocol.CommandBatch` submission;
- `events []protocol.DomainEvent`;
- `nextEventSequence uint64`;
- `telemetry []protocol.DraftTelemetry`;
- `nextTelemetrySeq uint64`;
- `battles []*battle.Session`;
- `nextBattleID uint64`;
- `encounterResolver game.EncounterResolver`;
- `encounterContext game.ResolveContext`;
- `invasion *game.InvasionOpportunity`;
- `handledInvasions []game.InvasionHandledKey`;
- `eliminatedEmpires []core.ID`;
- `result *Result`;
- `pendingEliminationCheck bool`.

### Persist directly

The live snapshot must contain all of the above **except the resolver interface object itself**.

The three next-sequence values should be persisted explicitly and cross-validated against history where possible:

- `next_event_sequence`;
- `next_telemetry_sequence`;
- `next_battle_id`.

This is deliberately stricter than relying on reconstruction. Session event/telemetry counters can currently be derived from complete history, but storing them makes corruption visible. `nextBattleID` cannot safely be reconstructed from the current `s.battles` list alone because completed battle children can be discarded while future battle IDs must remain monotonic.

### Re-inject, do not persist

`encounterResolver` is a runtime dependency, not authoritative game data.

The production continuation resolver is `*game.EconomyResolver`, whose mutable struct state is only a pointer to `EconomyRules`. All staged continuation-specific data is already in `encounterContext`, Core state, Battle results and Invasion state.

A restored session at `encounters` or `invasion_decisions` therefore receives the host's validated continuation resolver after ruleset identity validation. A future stateful/custom resolver would need a later snapshot extension; Slice 14 does not serialize Go interfaces/function pointers.

## Battle child state inventory

`battle.Session` contains:

- `Spec`;
- Battle `Phase`;
- optional Battle `Result`;
- optional tactical runtime;
- `tacticalRevision`.

The tactical runtime contains:

- `TacticalState`:
  - round;
  - initiative order;
  - active Ship ID;
  - next command sequence;
  - tactical RNG state;
  - per-Ship armor/structure/weapon-ready/destroyed state;
- Tactical event history;
- next Tactical event sequence.

A live Battle snapshot therefore needs at minimum:

- Spec;
- Phase;
- Result;
- Tactical State;
- Tactical Events;
- next Tactical event sequence;
- Tactical revision.

Battle children must be serialized in ascending Battle ID order. Active tactical state must preserve RNG and command/event sequences exactly; otherwise save/load can change future hit/damage/result bytes.

## Encounter/Invasion continuation state

`game.ResolveContext` contains:

- active Seat-to-Empire authority mapping;
- sorted handled-Invasion keys.

The Session keeps both `encounterContext` and canonical `handledInvasions`. They matter because a strategic turn can pause through multiple Battle/Invasion continuations after the original turn batch has already been committed.

`pendingEliminationCheck` is also future-relevant: conquest may have happened, but elimination/victory intentionally waits until encounter/invasion continuation has completed.

These values must survive live load exactly.

## Ruleset identity is a required persistence dependency

Gate 1 found a persistence gap not represented in the original Slice-14 draft: `EconomyRules` currently carries **no ruleset ID/version/fingerprint**. `LoadEconomyRules(rulesetDir)` receives a path and loads normalized JSON, but that identity is discarded.

Restoring the same Core23 state against modified rules under the same folder name could change Research, Economy, construction, movement or Tactical continuation while still decoding successfully.

### Proposed rules identity

Slice 14 should add a deterministic rules identity to loaded `EconomyRules`:

- `id`: normalized base directory name, currently `moo2-1.31`;
- `sha256`: digest of the actual JSON files consumed by `LoadEconomyRules`.

Loaded rules files today are:

- `buildings.json`;
- `economy.json`;
- `new_game_galaxy.json`;
- `planet_classes.json`;
- `race_traits.json`;
- `races.json`;
- `ship_hulls.json`;
- `tactical_combat.json`;
- `technologies.json`.

Proposed fingerprint algorithm:

1. sort those loaded filenames ascending;
2. SHA-256 each raw file;
3. build UTF-8 lines `<filename>:<lowercase-file-sha256>\n`;
4. SHA-256 the complete manifest bytes.

For the current committed `data/rulesets/moo2-1.31` loaded-rule files the Gate-1 probe produced:

`771e34863008fb70bf947ee583b2a1a872f96feca564b28501d6ea726fc8bd8c`

This is evidence only; Gate 3 must compute it in Go rather than hard-code it.

A live save must be rejected if either active ruleset ID or fingerprint differs.

`assets.json` is intentionally not in this digest because `LoadEconomyRules` does not consume it and it cannot affect authoritative continuation.

## Stable save boundaries

The Host already drives a mutation until the next interactive boundary under `hostedGame.mu`. That lock is the correct read/save serialization boundary: export waits for an in-flight mutation to finish and cannot observe a half-committed resolver transaction.

### Saveable

1. **`planning`**
   - partial Seat submissions are valid and must survive;
   - Built-in AI Seat submissions/controller identity survive;
   - no resolver continuation object is required.

2. **`encounters`**
   - Battle children, completed/active child phases, Battle/tactical RNG/revisions/events and `encounterContext` are persisted;
   - a compatible host `EncounterResolver` must be re-injected on load.

3. **`invasion_decisions`**
   - pending `InvasionOpportunity`, continuation context, handled keys and pending elimination state are persisted;
   - compatible `EncounterResolver` + `InvasionResolver` capability must be available on load.

4. **`post_resolution` only when an interactive Colony-Base decision is pending**
   - this is the only PostResolution state the current Host deliberately exposes as an interactive stop;
   - ordinary PostResolution with no pending human decision is auto-driven to CompleteTurn and is not part of the v1 public save contract.

5. **`completed`**
   - accepted by live save;
   - immutable Result/elimination invariants apply;
   - existing `MarshalCompletedSnapshot` remains supported separately.

### Explicitly unsaveable

- **`strategic_resolution`**;
- transient PostResolution with no interactive Colony-Base decision;
- any point while a Host mutation/resolver commit is in flight.

The Host lock makes the last case unobservable through the supported API. Session-level marshal must still reject unsupported phases explicitly.

## Proposed live snapshot v1 shape

The exact Go field names can be finalized in Gate 2, but the authoritative JSON content should be equivalent to:

- `schema_version = 1`;
- `ruleset`:
  - `id`;
  - `sha256`;
- `game_id`;
- `revision`;
- `phase`;
- `state` (typed Core23 `GameState`);
- `seats[]`:
  - `seat` including `ControllerType`;
  - optional `submission`;
- `events[]`;
- `next_event_sequence`;
- optional `telemetry[]`;
- `next_telemetry_sequence`;
- optional `battles[]` using a versioned Battle snapshot payload;
- `next_battle_id`;
- `encounter_context`;
- optional `invasion`;
- `handled_invasions[]`;
- `eliminated_empire_ids[]`;
- optional `result`;
- `pending_elimination_check`.

No timestamps, filesystem paths, filenames, subscriber IDs, socket IDs or process-local identifiers belong in authoritative save bytes.

## Canonical bytes and validation policy

### Encoding

- compact `encoding/json` struct encoding, matching the completed-session snapshot approach;
- fixed struct field order;
- canonical sorted slices where the runtime contract already treats ordering as identity-independent:
  - Seats by Seat ID;
  - Battles by Battle ID;
  - handled Invasion keys by Colony/Attacker Empire;
  - eliminated Empire IDs ascending;
- existing event/telemetry/tactical-event sequence order preserved exactly;
- Core23 state embedded as a typed object, not as a quoted JSON blob.

### Decode

Live decode must:

- use `DisallowUnknownFields()`;
- reject trailing non-whitespace JSON after the single snapshot object;
- accept **only** `LiveSnapshotSchemaVersion == 1` in Slice 14;
- reject old/new versions rather than guessing or silently migrating;
- call Core23 `GameState.Validate()`;
- validate ruleset ID/fingerprint before activation;
- validate Seat IDs/Empire mapping/controller values;
- validate all stored submissions against snapshot game/seat/turn and their recorded command schema;
- validate Session event and telemetry sequence monotonicity plus explicit next counters;
- validate Battle IDs/spec game/turn/participants, phase/result/tactical invariants, tactical revision/RNG/command/event sequences and next event counter;
- validate `next_battle_id` is strictly above every created/persisted Battle ID required by history;
- validate phase-specific Battle/Invasion/Result/continuation invariants;
- validate sorted unique handled/eliminated IDs and Result consistency;
- build the replacement `GameSession` completely off to the side before any Host mutation.

There are **no migrations in Slice 14**. Future format changes increment the version and add explicit migration code later.

## Host state: persist vs reconstruct

Current `hostedGame` owns:

- Session pointer;
- strategic/immediate resolver pointers;
- Seat ID list and Seat info cache;
- `changeSequence`;
- `nextSubscriberID`;
- subscriber channels.

### Reconstruct / preserve outside save

Do not persist:

- resolver pointers;
- Seat cache/list (rebuild from restored Session Seats);
- `changeSequence` as gameplay state;
- subscriber IDs/channels;
- mutexes/process-local state.

### ChangeSequence semantics

`changeSequence` is transport invalidation state, not game determinism.

- Import into a fresh Host registers the restored Game with `changeSequence = 1`.
- Replacing an existing hosted Game preserves the hosted transport object/subscribers, increments the existing `changeSequence` exactly once and publishes one `snapshot_invalidated` notification with reason `game_loaded` and the restored Game revision.
- Restored Session `revision` is **not rewritten** to match Host transport history; save/load may intentionally rewind to an older authoritative revision.
- Clients must refetch after `game_loaded` before issuing further commands.

This keeps exact game continuation independent from server connection history.

## Proposed App/Host authority surface

Storage-neutral App methods should be equivalent to:

- `Host.ExportLiveSnapshot(gameID) ([]byte, error)`
  - lock hosted game;
  - require saveable stable boundary;
  - read-only; no revision/changeSequence/event mutation.

- `Host.ImportLiveSnapshot(data []byte) (GameSummary, error)`
  - available only on a Host configured with authoritative rules/resolvers;
  - decode/validate against active ruleset before touching `h.games`;
  - use embedded GameID;
  - conflict if that GameID already exists;
  - register with changeSequence 1;
  - do not auto-drive after load.

- `Host.RestoreLiveSnapshot(gameID, data []byte) (Receipt, error)`
  - decode/validate completely first;
  - embedded GameID must equal route/target GameID;
  - atomically replace only the Session + Seat caches under hosted-game lock;
  - preserve current configured rules/resolvers and subscriber channels;
  - increment ChangeSequence once and notify `game_loaded`;
  - invalid input leaves Session, ChangeSequence and subscribers untouched.

No server-side save directory or filename policy is part of Slice 14.

## HTTP authority proposal

A live save contains hidden data for every Empire and can replace authoritative state. It must **not** be exposed as a Seat endpoint.

Current server config already treats Observer access as privileged and disabled by default. Slice 14 should add a separate privileged switch, proposed `PersistenceEnabled bool`, also disabled by default.

Proposed storage-neutral endpoints:

- `GET /api/v1/games/{gameID}/live-snapshot` - export raw JSON bytes;
- `POST /api/v1/games/import` - import a save as a new hosted Game using embedded GameID;
- `PUT /api/v1/games/{gameID}/live-snapshot` - atomically restore/replace that existing Game.

All three require persistence API enablement. A later browser Slice can provide file download/upload UX without changing the engine contract.

Suggested error families:

- persistence disabled -> `403 forbidden`;
- malformed/unknown-version/invariant-invalid save -> `400 invalid_save`;
- wrong active ruleset -> `409 ruleset_mismatch`;
- duplicate imported GameID -> `409 game_exists`;
- route/embedded GameID mismatch -> `409 game_id_mismatch`.

## Built-in AI resume contract

No AI-private state is persisted in Slice 14.

Persisted Seat controller identity plus the restored player-safe state/catalog inputs are sufficient because `internal/ai baseline_v1`:

- has no private RNG;
- has no memory between decisions;
- cannot read ObserverView/mutable GameState;
- is deterministic for the same DecisionView.

After live load the Host simply resumes its normal controller orchestration at the restored interactive boundary.

## Exact continuation regression plan

### A. Canonical mid-match AI-vs-AI continuation

Use the existing real canonical fixture:

- ruleset `moo2-1.31`;
- seed `0x8009`;
- Small / Normal / Average / Tactical;
- Human + Darlok content;
- both Seats `ControllerBuiltinAI`.

Run 1 uninterrupted to `completed` and capture existing `MarshalCompletedSnapshot()` bytes.

Run 2:

1. start the identical match;
2. advance through `Host.AdvanceAutomation` to deterministic **Planning Turn 250**;
3. export live snapshot;
4. import into a **fresh rules-configured Host**;
5. immediately re-export and require exact live snapshot bytes;
6. continue the restored game to completion;
7. require final completed-session snapshot bytes to equal Run 1 exactly.

This proves State RNG, IDs, Session revisions/events, AI decisions and future result are preserved with no hidden AI state.

### B. Partial Human Planning submission

Two human-controller Seats:

- submit Seat 1 only;
- save at Planning while Seat 2 remains pending;
- restore;
- verify Seat/controller/submission projection exactly;
- apply identical Seat-2 command batch on original/restored branches and require exact continuation.

### C. Active Tactical Encounter

Use a supported Slice-07 1-vs-1 Tactical fixture:

- enter `encounters`;
- execute at least one tactical command so Tactical RNG/event/command/revision counters have moved;
- save/restore/re-export exact;
- continue with identical commands and require exact Battle result + Session continuation/event bytes.

### D. Invasion decision

Save at `invasion_decisions`, restore, issue identical invade/decline command and require exact state/events/result continuation.

### E. PostResolution Colony-Base decision

Save with a pending human Colony-Base decision, restore, issue the same choice and require exact next Planning boundary.

### F. Completed

Live-save a completed session, restore/re-export exact, then require its existing `MarshalCompletedSnapshot()` bytes to remain exact.

### G. Atomic rejection

For each malformed class, snapshot/ChangeSequence before the load attempt must equal after the rejected attempt.

At minimum cover:

- wrong live schema;
- wrong Core schema;
- unknown JSON field/trailing document;
- wrong ruleset ID/digest;
- invalid Seat/controller/submission;
- invalid next sequence counters;
- invalid Battle/tactical RNG/revision/sequence state;
- invalid phase/Battle/Invasion/Result combination;
- duplicate import GameID;
- mismatched restore GameID;
- unsupported `strategic_resolution` save.

## Gate-2 contract proposed for approval

1. Slice 14 adds a separate **LiveSnapshot v1**; the existing CompletedSnapshot v1 remains compatible and unchanged as a public format.
2. LiveSnapshot embeds typed **Core23 GameState** and all future-deterministic Session continuation state, not a parallel copy of strategic data.
3. Persist Seats/controllers/submissions, Session events/telemetry and explicit next counters, Battle snapshots/nextBattleID, encounter context, pending Invasion/handled keys, elimination/result and pending-elimination state.
4. Battle v1 persists Spec/Phase/Result plus Tactical State, RNG, command sequence, events, next event sequence and Tactical revision.
5. Resolver/interface pointers, mutexes, Host subscriber state and network/process state are never serialized; production `EconomyResolver` is re-injected after rules validation.
6. Add rules identity to loaded EconomyRules: directory-base ID plus deterministic SHA-256 over exactly the authoritative JSON files consumed by `LoadEconomyRules`.
7. A save loads only when **both ruleset ID and fingerprint match** the active Host rules.
8. Saveable v1 phases are Planning, Encounters, InvasionDecisions, interactive PostResolution with pending Colony-Base decision, and Completed. StrategicResolution/transient non-interactive PostResolution/in-flight mutation are rejected.
9. Live JSON is compact/canonical, strictly decoded with unknown-field + trailing-data rejection; Slice 14 accepts version 1 only and implements no migrations.
10. `ExportLiveSnapshot` is read-only and runs under hosted-game lock; it changes neither Session revision/events nor Host ChangeSequence.
11. `ImportLiveSnapshot` creates a new hosted Game from embedded GameID on a rules-configured Host and starts transport ChangeSequence at 1; it does not auto-drive the restored boundary.
12. `RestoreLiveSnapshot` validates fully before an atomic existing-game replacement, preserves configured resolvers/subscribers, increments ChangeSequence once and publishes `snapshot_invalidated` reason `game_loaded`; rejected loads change nothing.
13. Host ChangeSequence/subscriber IDs are transport state and are not persisted; exact Session revision from the save is restored unchanged.
14. HTTP persistence is a separate privileged capability (`PersistenceEnabled`, default false), not a Seat API. Baseline routes are GET live snapshot, POST import and PUT restore as described above.
15. Built-in AI persists no private state; controller identity + restored authoritative player-safe state is sufficient for deterministic resume.
16. Canonical exact proof is uninterrupted `0x8009` AI-vs-AI versus save at Planning Turn 250 -> fresh-Host import -> exact re-export -> continue; final existing completed-snapshot bytes must match uninterrupted bytes exactly.
17. Gate 3 also requires exact phase fixtures for partial Planning submission, active Tactical Encounter, Invasion decision, interactive PostResolution and Completed, plus malformed/old/rules-mismatch atomic rejection regressions.
18. Server filesystem save slots, cloud sync, browser file UX, original MOO2 SAV compatibility, future migrations and stateful external/MCP controller connection persistence remain deferred.

## Gate 1 status

All Gate-1 audit items are complete. The 18-point contract above is the **Gate-2 freeze candidate**. No live-persistence implementation is authorized until Gate 2 is explicitly approved.
