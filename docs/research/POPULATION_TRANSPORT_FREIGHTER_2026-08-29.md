# Population transport / shared Freighter pool - 2026-08-29

**Slice status:** OPEN - gate 1 pending

**Open marker:** `docs/slices/_OPEN_POPULATION_TRANSPORT_FREIGHTER_2026-08-29.md`

**Workflow:** `docs/slices/README.md`

## Goal

Resolve the original MOO2 1.31 Population transport / Freighter interaction before extending authoritative logistics state beyond Food transport.

This slice must determine from original executable/save evidence where practical:

1. which original command/state creates a Population transport and which state represents it in flight;
2. when Population transport reserves, consumes or returns Freighter capacity relative to Food imports;
3. whether Food and Population use one shared capacity counter directly, separate reservations, or a fixed settlement priority;
4. cancellation, arrival, ownership-change and blocked-route behavior only where evidence supports it;
5. the smallest deterministic MOOX transport state, legal-action surface and Observer/replay events required by those rules.

The already-proven Food blockade behavior must remain unchanged while this slice is investigated.

## Gate 1 - Checkup + original analysis / reverse engineering

**Status:** pending.

Start by locating the original Population transport creation, movement and arrival routines and their Freighter reads/writes. Do not add gameplay implementation before the original capacity/timing behavior is documented strongly enough to discuss an implementation shape.

### Evidence log

Pending.

### Known unknowns

- exact original Population transport record/state layout;
- Freighter reservation/consumption timing;
- Food-vs-Population capacity priority;
- cancellation and arrival capacity release;
- blockade/hostility interaction with Population transport routes;
- ownership semantics while a transport is in flight.

## Gate 2 - Implementation decision

**Status:** blocked by gate 1 and user discussion.

The accepted Core/Game/Session state and command/event shape will be recorded here before code changes begin.

## Gate 3 - Implementation

**Status:** not started.

## Gate 4 - Follow-up QA + commit

**Status:** not started.

Final checks and commit result will be recorded here before this `_OPEN_` marker is removed.
