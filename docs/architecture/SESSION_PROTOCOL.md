# Session, command and observer protocol

Status: implemented Phase 1 runtime boundary as of 2026-08-27.

This document describes the concrete contract currently implemented by `internal/protocol`, `internal/game`, `internal/session` and `internal/battle`. It refines the target architecture in `README.md`; it does not define unverified MOO2 gameplay formulas.

## Command envelope

Strategic player intent is represented by a versioned `protocol.Command` and submitted as a versioned `protocol.CommandBatch`.

A batch carries:

- `game_id`;
- `seat_id`;
- strategic `turn`;
- `base_revision`;
- an ordered `commands[]` list.

Command sequences are contiguous (`1..N`). Payloads are JSON so future Wails, HTTP/WebSocket and MCP adapters can preserve the same semantic envelope.

The current protocol layer validates schema/identity/order/JSON structure. Concrete gameplay command kinds and their domain legality are intentionally the next strategic-resolver slice.

## Strategic resolver boundary

`internal/game` defines the transport-independent resolver contract used after all seats have submitted:

```text
Resolve(detached GameState, stable CommandBatch[])
    -> resolved GameState
    -> ordered strategic DomainEvent[]
    -> Encounter[]
```

`GameSession.ResolveStrategic` is transactional. It validates the returned state, immutable seed/turn constraints, event references/JSON, and all encounter participants before committing anything. Invalid output leaves the prior authoritative state, revision, phase and event history unchanged.

The resolver receives detached state and command copies, and the returned state is cloned again before commit. A resolver therefore cannot retain a pointer and mutate authoritative session state after resolution.

A successful strategic resolution increments the authoritative revision, appends resolver events with session-owned event sequence numbers, and then enters either `encounters` or `post_resolution`.
## Authoritative GameSession

`session.GameSession` owns a detached copy of the supplied `core.GameState`. Callers receive projected/copied views rather than mutable access to that state.

The implemented strategic lifecycle is:

```text
planning
   |
   | all seats SubmitTurn
   v
strategic_resolution
   |
   | resolver determines zero or more encounters
   v
encounters (only when battles exist)
   |
   | all battle sessions complete
   v
post_resolution
   |
   | CompleteTurn
   v
planning / next turn + revision
```

If strategic resolution produces no encounter, `BeginEncounters(nil)` advances directly to `post_resolution`.

## Parallel planning

Seats submit independently while the session remains in `planning`.

A submission must match the current:

- `game_id`;
- `turn`;
- `base_revision`;
- known `seat_id`.

Duplicate submissions are rejected.

The observer can see immediately which seats have submitted and can inspect their submitted batches. This live readiness state is not itself added to the authoritative replay stream in transport-arrival order.

When the final required seat submits, the session records `turn_submitted` events in stable ascending seat order and then records the transition to `strategic_resolution`. This makes replay independent from network timing.

## Player projection

`PlayerView(seatID)` currently exposes only the minimum implemented private state:

- game/turn/revision/phase;
- public seat readiness;
- the requesting seat;
- the requesting seat's empire;
- colonies owned by that empire;
- the requesting seat's own submitted command batch, if already submitted.

It deliberately does not expose the complete `GameState` or other empires. Star/fleet/intelligence visibility will be added later as explicit projections rather than by handing clients the authoritative state.

## Observer projection

`ObserverView()` is privileged and contains:

- a detached complete `GameState` snapshot;
- all seats and current submissions;
- authoritative session/strategic/battle event history;
- all current battle-session views;
- optional non-authoritative draft telemetry.

Returned state, event payloads, command payloads and battle slices are copied so modifying a view cannot mutate the authoritative session.

## Draft and AI telemetry

`PublishDraftTelemetry` provides an optional observer-only channel during `planning`.

It is intended for development/AI inspection such as:

```text
AI seat 2
kind: ai.plan
summary: consider colony defense
data: {"goal":"defense"}
```

Telemetry is deliberately separate from authoritative events and therefore does not affect replay, state resolution or deterministic outcomes. It should contain explicit diagnostic metadata supplied for observation, not hidden model reasoning.

## Tactical BattleSession boundary

`internal/battle` currently implements only the deterministic session boundary, not tactical MOO2 combat rules.

A battle has:

- stable battle ID;
- parent game ID;
- strategic turn;
- sorted participant seat IDs;
- deterministic infrastructure seed;
- `pending -> active -> completed` lifecycle;
- a final result with winner seats and outcome.

`battle.DeriveSeed` is infrastructure for deterministic isolation between encounters. It is explicitly not claimed to reproduce the original MOO2 combat RNG.

## Parallel battle determinism

Multiple child battles may run concurrently. Their wall-clock completion order does not become authoritative ordering.

For example, if battle 2 finishes before battle 1, the observer can see battle 2's live completed state immediately. The authoritative `battle_completed` events are emitted only once all required battles are done, in stable battle-ID order.

This allows independent players to fight concurrently without making process scheduling or network latency part of the simulation result/replay contract.

## Current limitations / next slice

This checkpoint intentionally does not yet implement:

- real strategic command kinds and concrete economy resolver behavior;
- colony economy resolution;
- movement/conflict detection;
- application of battle results back into ships/fleets;
- tactical combat commands/mechanics;
- HTTP/WebSocket networking;
- Wails services;
- MCP tools/resources;
- authentication/reconnect persistence.

The next development slice is the first real strategic resolver: population assignment followed by verified food/production/research/money resolution, expressed as commands and domain events through this session boundary.
