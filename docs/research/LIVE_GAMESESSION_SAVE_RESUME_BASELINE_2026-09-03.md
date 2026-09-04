# Live GameSession save / resume baseline - Gate 1 evidence

Date: **2026-09-03**

Slice: **14 - Live GameSession save / resume baseline**

Status: **closed; Gates 1-4 complete**.

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

## Ruleset and simulation compatibility are persistence dependencies

Gate 1 found that `EconomyRules` carries no explicit ruleset identity. Gate 2 confirms that exact resume needs **two** compatibility guards:

1. exact authoritative rules data; and
2. deterministic simulation-code compatibility.

A matching ruleset alone is insufficient if Go transition logic changes between save and load. Conversely, a build identifier is too strict because documentation/UI-only commits should not invalidate saves.

### Ruleset identity

Slice 14 should add a deterministic identity to rules loaded by `LoadEconomyRules`:

- `id`: normalized base directory name, currently `moo2-1.31`;
- `sha256`: digest over exactly the JSON files consumed by `LoadEconomyRules`.

Loaded authoritative rules files today are:

- `buildings.json`;
- `economy.json`;
- `new_game_galaxy.json`;
- `planet_classes.json`;
- `race_traits.json`;
- `races.json`;
- `ship_hulls.json`;
- `tactical_combat.json`;
- `technologies.json`.

Gate 2 refines the Gate-1 raw-file proposal so the fingerprint is stable across Windows/Linux checkout whitespace and line endings:

1. sort the loaded filenames ascending;
2. read each JSON file and run Go `json.Compact` over its bytes;
3. SHA-256 the compacted JSON bytes;
4. build UTF-8 lines `<filename>:<lowercase-file-sha256>\n`;
5. SHA-256 the complete manifest bytes.

For the current committed `data/rulesets/moo2-1.31` the reviewed algorithm produces:

`2e2e9760051845a99092457760cfe78d0e90e20381fecf1072002c2081844f5b`

This value is evidence only; Gate 3 computes it in Go. The earlier Gate-1 raw-byte probe `771e3486...` is superseded because raw bytes are needlessly sensitive to checkout whitespace/line endings.

`assets.json` and `README.md` are intentionally excluded because `LoadEconomyRules` does not consume them and they cannot affect authoritative continuation.

A live save is rejected unless ruleset ID **and** compacted rules fingerprint match the active Host rules.

### Simulation compatibility

LiveSnapshot v1 adds `simulation_compat_version = 1` independently from `schema_version = 1` and Core23.

This constant represents deterministic continuation semantics implemented in Go, including the built-in AI policy used for automatic controller continuation. Future code changes that can alter the same saved state/rules/controller input into different authoritative commands, RNG consumption, events, revisions, Battle results or final result must increment this compatibility version even if the JSON shape does not change.

A different simulation compatibility version is rejected rather than silently resumed. This avoids tying save compatibility to arbitrary Git commits while still preventing false exact-continuation claims across behavior-changing engine builds.

### Resolver and controller scope for v1

The Gate-1 statement that the production resolver is stateless is correct but must become an explicit v1 support guard:

- export/import/restore is supported only when the hosted strategic resolver is the production `*game.EconomyResolver` and its immediate resolver uses the same validated `EconomyRules` identity;
- custom/stateful Resolver implementations are rejected by v1 persistence rather than serialized or guessed;
- supported persisted controller types are `local_human`, `remote_human` and `builtin_ai`;
- `external_ai` and `mcp_ai` live saves are rejected in v1 because their external/private controller state and connection lifecycle are not part of the snapshot.

This keeps the exact-continuation guarantee honest. Supporting stateful custom resolvers or external/MCP controller checkpoints requires a later versioned persistence extension.

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
   - loader validation requires at least one pending Colony-Base resolution **and zero due Research completions under the active rules**, matching the Host boundary ordering;
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
- `simulation_compat_version = 1`;
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
- validate simulation compatibility plus ruleset ID/fingerprint before activation;
- validate Seat IDs/Empire mapping/controller values and reject ExternalAI/MCPAI in LiveSnapshot v1;
- validate all stored submissions against snapshot game/seat/turn and their recorded command schema;
- validate Session event and telemetry sequence monotonicity plus explicit next counters;
- validate Battle IDs/spec game/turn/participants, phase/result/tactical invariants, tactical revision/RNG/command/event sequences and next event counter;
- validate `next_battle_id` is strictly above every created/persisted Battle ID required by history;
- validate phase-specific Battle/Invasion/Result/continuation invariants, including PostResolution pending-Colony-Base plus zero-due-Research requirement;
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
  - require the configured production EconomyResolver/rules compatibility contract;
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

All three require persistence API enablement. `cmd/moox-server` exposes a separate `-enable-persistence` flag, default false.

The existing `maxJSONBodyBytes = 1 MiB` remains the command-envelope limit. Live snapshot import/restore uses a separate v1 request cap of **64 MiB** so normal current matches are not constrained by the command limit while malformed/unbounded uploads are still bounded. Oversize snapshot requests return HTTP `413` and do not reach Host mutation. A later browser Slice can provide file download/upload UX without changing the engine contract.

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

- ruleset `moo2-1.31` with reviewed compacted fingerprint `2e2e9760051845a99092457760cfe78d0e90e20381fecf1072002c2081844f5b`;
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

## Gate-2 contract - reviewed freeze candidate

1. Slice 14 adds a separate **LiveSnapshot v1**; existing CompletedSnapshot v1 remains compatible and unchanged as a public format.
2. LiveSnapshot embeds typed **Core23 GameState** plus all future-deterministic Session/Battle continuation state instead of copying strategic fields into a second model.
3. LiveSnapshot v1 records both `schema_version = 1` and **`simulation_compat_version = 1`**. A behavior-changing engine/built-in-AI change must bump simulation compatibility even when snapshot JSON shape is unchanged.
4. Loaded EconomyRules gain a rules identity: normalized directory-base ID plus SHA-256 over a sorted manifest of the **nine JSON files actually consumed by `LoadEconomyRules`**, hashing each after Go `json.Compact`. Current reviewed `moo2-1.31` fingerprint: `2e2e9760051845a99092457760cfe78d0e90e20381fecf1072002c2081844f5b`.
5. Load requires exact simulation compatibility, ruleset ID and rules fingerprint. No best-effort resume is allowed on mismatch.
6. Persistence v1 supports only hosted games using the production `*game.EconomyResolver` for strategic continuation with matching immediate EconomyRules. Custom/stateful Resolver implementations are rejected rather than falsely claimed resumable.
7. Persisted controller types supported by v1 are `local_human`, `remote_human` and `builtin_ai`. `external_ai` and `mcp_ai` saves are rejected because external/private controller state is not persisted.
8. Persist Seats/controllers/submissions, Session events/telemetry and explicit next counters, Battle snapshots/nextBattleID, encounter context, pending Invasion/handled keys, elimination/result and pending-elimination state.
9. Battle v1 persists Spec/Phase/Result plus Tactical State, RNG, command sequence, events, explicit next event sequence and Tactical revision. Tactical counters are cross-validated; active Battle children restore without recomputing their initial state from current rules.
10. Resolver/interface pointers, mutexes, Host subscriber state and network/process state are never serialized. The compatible production resolver is re-injected only after snapshot/rules validation.
11. Saveable v1 phases are Planning, Encounters, InvasionDecisions, interactive PostResolution with pending Colony-Base decision, and Completed. PostResolution restore additionally requires **zero due Research completions** under the active rules.
12. StrategicResolution, transient/non-interactive PostResolution and any point inside an in-flight Host mutation/resolver commit are explicitly unsaveable. Host locking makes in-flight states unobservable through supported export.
13. Live JSON uses compact deterministic struct encoding, stable slice ordering and strict `DisallowUnknownFields` plus single-document/trailing-data rejection. Slice 14 accepts version 1 only and implements no migrations.
14. Explicit `next_event_sequence`, `next_telemetry_sequence`, `next_battle_id` and Tactical next-event/revision counters are persisted and validated against their histories/invariants; they are not silently reset on load.
15. `Host.ExportLiveSnapshot(gameID)` runs under hosted-game lock and is read-only: no auto-drive, revision, event or Host ChangeSequence change.
16. `Host.ImportLiveSnapshot(data)` fully validates against a rules-configured Host, uses embedded GameID, conflicts on an existing ID, registers at transport ChangeSequence 1 and **does not auto-drive** the restored boundary so immediate re-export can be byte-exact.
17. `Host.RestoreLiveSnapshot(gameID,data)` fully validates before locking/replacing the existing Session, requires embedded/target GameID equality, preserves configured resolvers/subscribers, rebuilds Seat caches, increments existing ChangeSequence once and publishes one `snapshot_invalidated` notification with reason `game_loaded`. Rejection changes nothing.
18. Host ChangeSequence/subscriber IDs are transport state and are not persisted. Restored Session revision is the saved revision unchanged, including intentional rewind to an older save; clients refetch after `game_loaded`.
19. HTTP persistence is a separate privileged capability (`PersistenceEnabled`, default false; server flag `-enable-persistence`), never a Seat endpoint. Baseline routes are GET `/api/v1/games/{gameID}/live-snapshot`, POST `/api/v1/games/import`, and PUT `/api/v1/games/{gameID}/live-snapshot`.
20. Live import/restore uses a dedicated **64 MiB** request cap, separate from the existing 1 MiB command JSON limit. Oversize requests fail with 413 before Host mutation.
21. Built-in AI persists no private state. `baseline_v1` remains pure; controller identity plus restored authoritative state is sufficient **within the matching simulation compatibility version**.
22. Canonical exact proof is uninterrupted real `NewGame(0x8009)` AI-vs-AI versus save at Planning Turn 250 -> fresh compatible Host import -> exact live re-export -> continue; final existing CompletedSnapshot bytes must equal the uninterrupted run exactly.
23. Gate 3 also requires exact phase fixtures for partial Planning submission, active supported Tactical Encounter after at least one command/RNG transition, Invasion decision, interactive PostResolution and Completed.
24. Malformed/unknown-version/trailing-data, simulation/rules mismatch, unsupported controller/resolver, invalid counter/phase/Battle/Invasion/Result, duplicate import and target GameID mismatch all receive atomic rejection tests; rejected restore preserves Session bytes and Host ChangeSequence.
25. Server filesystem slots, cloud sync, browser file UX, original MOO2 SAV compatibility, future migrations, custom/stateful resolver persistence and External/MCP controller checkpointing remain deferred.

Gate 2 design review completed this 25-point contract. It was **explicitly frozen by user approval on 2026-09-04** before Gate 3 implementation.

## Gate 1 / Gate 2 status

Gate 1 audit was accepted when the user advanced to Gate 2 on 2026-09-03. Gate 2 refined the contract above and was explicitly frozen when the user advanced to Gate 3 on 2026-09-04.

## Gate 2 freeze and Gate 3 implementation - 2026-09-04

The user explicitly advanced from Gate 2 to Gate 3 on 2026-09-04, freezing the reviewed 25-point contract above. Gate 3 is implemented in commit `a2c4ccd` (`game: add live gamesession persistence`).

### Implemented authority/runtime surfaces

- `EconomyRules` now carries a loaded ruleset ID and compacted-JSON SHA-256 identity. The committed `moo2-1.31` fingerprint is `2e2e9760051845a99092457760cfe78d0e90e20381fecf1072002c2081844f5b`; a regression proves harmless JSON whitespace does not change it.
- `LiveSnapshotSchemaVersion = 1` and `SimulationCompatibilityVersion = 1` are separate guards. Rules/simulation mismatches are rejected before activation.
- Battle live snapshots preserve Spec/Phase/Result plus Tactical State, RNG, next command/event sequences, event history and Tactical revision. Active Tactical combat can save after an RNG-consuming Beam shot, roundtrip exactly and continue identically.
- GameSession live snapshots preserve Core23, revision/phase, Seats/controllers/submissions, events/telemetry and explicit counters, Battle children/nextBattleID, encounter context, Invasion/handled keys, eliminated Empires, Result and pending-elimination state. Historical `battle_created` events also constrain `next_battle_id` so removed completed Battles cannot allow ID reuse after load.
- Strict restore uses unknown-field and trailing-document rejection, Core validation, stable Seat/Battle/key ordering, phase-specific continuation validation and controller guards. Planning, Encounters, InvasionDecisions, interactive Colony-Base PostResolution and Completed are supported; StrategicResolution/transient PostResolution are rejected.
- Built-in AI requires no private save state. LiveSnapshot v1 permits `local_human`, `remote_human` and `builtin_ai`; ExternalAI/MCPAI and custom/stateful resolvers are explicitly rejected.
- `Host.ExportLiveSnapshot`, `Host.ImportLiveSnapshot` and `Host.RestoreLiveSnapshot` implement storage-neutral persistence. Import starts at Host ChangeSequence 1 without auto-drive. Restore validates off to the side, atomically replaces the Session, preserves subscribers/resolvers, increments ChangeSequence once and publishes `snapshot_invalidated` with reason `game_loaded`.
- HTTP persistence is separately privileged via `PersistenceEnabled` / `-enable-persistence` (default false): GET live snapshot, POST import and PUT restore. Import/restore use a separate 64 MiB body cap; the existing 1 MiB command limit is unchanged.

### Exact Gate 3 regressions

The committed tests cover:

- real canonical `NewGame(0x8009)` AI-vs-AI: uninterrupted completion versus save at **Planning Turn 250** -> fresh compatible Host import -> byte-identical live re-export -> continue; the existing final CompletedSnapshot bytes are exactly equal;
- Completed live snapshot -> fresh Host import/re-export -> existing CompletedSnapshot bytes remain exact;
- partial Human/BuiltinAI Planning submission plus draft telemetry -> exact roundtrip and identical continuation;
- active supported Tactical Encounter after a Beam command has consumed RNG -> exact Session/Battle roundtrip and identical Battle/strategic continuation;
- Invasion decision -> exact roundtrip and identical decline/next-Planning continuation;
- interactive Colony-Base PostResolution -> exact roundtrip and identical colonization/next-Planning continuation;
- invalid simulation/rules identity, trailing JSON, unsupported controller/custom resolver, StrategicResolution save, corrupt Tactical/Event counters, duplicate import and target/embedded GameID mismatch; rejected existing-game restores leave Session revision and Host ChangeSequence unchanged;
- HTTP persistence disabled-by-default behavior and enabled export/import/restore byte equality.

Gate 3 verification performed in the implementation session:

- targeted persistence matrix across game/battle/session/app/server: PASS;
- `go test ./... -count=1`: PASS, including `internal/session` in 45.174 s;
- `go vet ./...`: PASS;
- `git diff --check`: PASS before the implementation commit.

Gate 4 was executed independently on 2026-09-04 and passed; final evidence is recorded below.

## Gate 4 independent QA and closure - 2026-09-04

Gate 4 ran in a fresh ASH write session against clean HEAD `fc46a91`, with `main...origin/main [ahead 54]` and exactly one expected Slice-14 OPEN marker.

Independent QA repeated the frozen contract rather than relying on Gate-3 results:

- app persistence exactness/atomicity suite with `-count=2`: PASS, including the real canonical `NewGame(0x8009)` AI-vs-AI save at **Planning Turn 250**, fresh-Host import, byte-identical live re-export and byte-identical final existing CompletedSnapshot;
- Session/Battle `TestLiveSnapshot*` phase/RNG/continuation suite with `-count=2`: PASS for partial Planning submission, active Tactical after RNG-consuming Beam, Invasion decision, interactive Colony-Base PostResolution, Completed and Tactical-counter rejection;
- HTTP persistence suite with `-count=2`: PASS for default-disabled privilege and enabled export/import/restore behavior;
- the contract review found one missing explicit QA assertion: unknown `schema_version` and a partial snapshot missing `next_battle_id` were not directly exercised as existing-game atomic restore failures. Gate 4 added those two rejection assertions only, with no product-logic change, in commit `e6a5e69` (`test: harden live snapshot rejection coverage`); the targeted atomic-restore test passed twice;
- final `go test ./... -count=1`: PASS, including `internal/session` in **44.693 s**;
- final `go vet ./...`: PASS;
- final `npm run build` in `web`: PASS (TypeScript build + Vite 8.2.2 production build, 17 modules transformed, JS 204.00 kB / gzip 63.65 kB);
- final `git diff --check`: PASS.

All supported LiveSnapshot v1 phases, deterministic continuation claims and malformed/old/partial atomic-rejection requirements are now independently covered. No gameplay/runtime code changed in Gate 4. Slice 14 is closed; the OPEN marker is removed by the closure documentation commit. No push was performed. Next prepared objective: Slice 15 Gate 1 - Playable browser strategic HMI workflow audit.
