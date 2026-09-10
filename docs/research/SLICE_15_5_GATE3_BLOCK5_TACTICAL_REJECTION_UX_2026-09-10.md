# Slice 15.5 Gate 3 Block 5 - Tactical stale/illegal rejection UX and regression

Date: 2026-09-10

## Goal

Close the remaining Tactical mutation-rejection path before Gate 4: stale sequence, illegal movement and illegal fire must be visible, non-destructive, refetched when appropriate and never automatically resubmitted.

## UI hardening

`App.errorText` contained an unrelated recursion bug for non-API errors. It now returns the native Error message (or String fallback) instead of recursively calling itself.

`runBattleCommand` still records the shared App error and invokes the frozen Slice-15.4 409 refresh path, but now rethrows to the Tactical surface after handling. `TacticalBattlefield` catches that rejection locally and shows an in-context warning.

The warning states that no local Tactical change was kept, the rejected command was not sent again and server state remains authoritative. It does not falsely claim every rejection triggered a refetch because HTTP request-validation failures can be 400 rather than 409.

## Managed-browser rejection QA

QA used an isolated real supported 2v2 Tactical battle on port 7181. The normal UI generated a legal command; a one-shot browser fetch hook altered only the outbound test payload to force each rejection while counting Battle POSTs and snapshot GETs.

### Stale move

A legal movement click was sent with command sequence 99 while BattleSession expected sequence 1.

Observed:

- exactly 1 Battle POST;
- exactly 1 participant snapshot GET after rejection;
- HTTP/API error `session_rejected: command sequence 99, expected 1`;
- local `Taktisches Kommando abgelehnt` notice visible in the battlefield;
- active ship remained at `(10,9)` Facing 0;
- Tactical command sequence remained 1;
- no automatic resubmission.

### Illegal move

A legal movement click was test-tampered to destination `(14,9)`, occupied by enemy ship 60.

Observed:

- exactly 1 Battle POST;
- exactly 1 participant snapshot GET;
- server error identified the occupied destination/ship;
- local rejection notice remained visible;
- active ship position and Tactical sequence remained unchanged;
- no automatic resubmission.

### Illegal fire

A legal Laser-target click was test-tampered so target_ship_id equaled the firing Human ship ID.

Observed:

- exactly 1 Battle POST;
- exactly 1 participant snapshot GET;
- server error `target ship 57 is not an opponent`;
- local rejection notice visible;
- active ship and Tactical sequence unchanged;
- no automatic resubmission.

A preliminary sequence=0 probe was intentionally rejected earlier by the HTTP command validator as 400 before reaching BattleSession. That probe also produced one POST, zero mutation and a local rejection notice; the UI text was then generalized so it does not claim a refetch for every possible rejection class.

## Repeatable server/API regression

`TestBattleCommandEndpointRejectsStaleAndIllegalWithoutMutation` uses the existing real Tactical HTTP fixture and automatically verifies:

- stale positive sequence -> HTTP 409 `session_rejected`;
- occupied movement destination -> HTTP 409 `session_rejected`;
- own-ship fire target -> HTTP 409 `session_rejected`;
- ChangeSequence unchanged after every rejection;
- strategic revision unchanged;
- participant Battle projection deep-equal before/after each rejection.

This complements the Battle-core zero-mutation tests and the managed-browser UI evidence.

## Validation

Before block close:

- `go test ./internal/server` passed;
- `npm run build` passed.

The full repository Go tests/vet/web build and diff checks are run again before commit.

## Remaining work

Gate-3 implementation/evidence items are now complete. Gate 4 remains as the independent close pass with final full regression/build checks, tracker/history close marker and fresh canonical review reset.
