# Slice 15.6 Gate3 - Tactical automatic activation completion

Status: **IMPLEMENTED / isolated current-battle browser QA green**
Date: 2026-09-13

## Trigger
After authoritative Wait/Done ship switching was introduced, the user requested one further usability rule: when an active ship has completely consumed its per-round resources, it should not require an extra `Fertig` click.

## Authoritative rule
The server now auto-completes the current Tactical activation after a successful move or Beam fire when **both** conditions are true:

1. `movement_current <= 0`, and
2. no runtime weapon slot has `ready=true`.

This intentionally models **resource exhaustion**, not merely current target availability.

Consequences:

- unarmed ship: `movement_current == 0` is sufficient;
- armed ship at `movement_current == 0` with at least one ready weapon remains active;
- once the final ready weapon is fired while movement is already exhausted, activation auto-completes;
- a ready weapon that currently has no legal target still prevents auto-completion; the player can explicitly press `Fertig` in that case.

`Wait` semantics are unchanged. Movement and weapon readiness remain per-ship state and are not refreshed by switching.

## Event and progression contract
Auto-completion emits:

- kind: `activation_auto_ended`
- seat: active ship owner
- command sequence: the move/fire command that exhausted the final resource
- data: `ship_id`, `round`, `reason=resources_exhausted`

The normal activation progression path is shared with explicit `battle.end_activation`:

- current ship becomes `activation_complete=true`;
- server selects the next alive unfinished ship in circular initiative order;
- if no unfinished live ship remains, a new Tactical round starts;
- round start resets movement to max, reloads weapon readiness, clears activation-complete and recomputes initiative.

A battle-ending fire (victory) does not create a redundant auto-finish event after the winner has already been determined.

## Regression coverage
`internal/battle/tactical_test.go` now proves:

1. an unarmed ship using an exact remaining movement budget reaches 0 and auto-completes;
2. active authority moves to the next unfinished ship without an explicit Done command;
3. an armed ship at 0 movement does **not** auto-complete while its Laser remains ready;
4. firing that final ready Laser at 0 movement auto-completes the ship;
5. `activation_auto_ended` carries the triggering command sequence;
6. existing Wait/Done and normal round-reset tests remain green.

## Isolated browser acceptance from the current Triangle battle
The canonical `game-triangle-2pc` live snapshot was copied to isolated port 7183. Canonical 7171 was not mutated.

Starting snapshot:

- strategic turn 6 / encounters / revision 15;
- Battle #1 tactical revision 8;
- Tactical round 1;
- active Human Laser Scout #35;
- next command sequence 9;
- 10 Tactical events.

QA flow:

1. In the isolated copy, `Fertig` on #35 completed Tactical round 1.
2. Round 2 started with Human Scout #23 active at 20/20; #24 and Laser Scout #35 also reset to 20/20, and #35's Laser was ready.
3. The server projected legal moves whose `movement_remaining_after` was exactly 0.
4. A real battlefield click selected legal destination `(15,-2)`, move cost 20.
5. Server emitted `ship_moved` for #23 with `movement_remaining=0`.
6. The same command sequence then emitted `activation_auto_ended` with `reason=resources_exhausted`.
7. #23 became `activation_complete=true`, remained 0/20, and Scout #24 became authoritative active ship at 20/20 without any manual `Fertig` click.
8. Browser HUD immediately showed `Scout 2 · 20/20` as the active controlled ship.

This satisfies the requested convenience behavior while preserving server authority and the explicit Wait/Done contract whenever any resource remains.
