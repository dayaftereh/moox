# Slice 15.6 Gate3 - Tactical Wait / Done authoritative ship reordering

Status: **IMPLEMENTED / isolated current-battle browser QA green**
Date: 2026-09-13

## Trigger
Gate-3 usability testing showed that local ship selection was insufficient. A player could click an unfinished own ship, but the server still exposed legal actions only for one fixed `active_ship_id`. This made a persisted Laser Scout appear unavailable until its original initiative slot and made `Fertig` feel like the only way to advance.

The original Tactical research already records the intended distinction:

- `Wait` moves to another active ship while keeping the current ship in the active list;
- `Done` moves on and removes the current ship from the active list;
- movement is a per-round remaining-point pool;
- weapon readiness is explicit state.

## New authoritative contract
New command: `battle.wait_activation`.

Payload:

- `ship_id`: current authoritative active ship;
- optional `target_ship_id`: unfinished friendly ship to activate.

If `target_ship_id` is omitted, the server chooses the next unfinished same-seat ship deterministically from initiative order, wrapping as needed.

Validation is server-side:

- caller seat must control the current active ship;
- `ship_id` must equal `active_ship_id`;
- current ship must be alive and unfinished;
- target must be another alive, unfinished ship owned by the same seat;
- enemy, destroyed, completed or stale targets are rejected with zero mutation.

Accepted Wait:

- changes only `active_ship_id`;
- does **not** set `activation_complete`;
- does **not** refill movement;
- does **not** reload weapons;
- emits `activation_waited` with source/target ship IDs and current round;
- increments normal command/battle sequencing exactly like other mutating Tactical commands.

The Tactical view now projects `can_wait_activation` and `wait_target_ship_ids` from server authority.

## Done / round behavior
`battle.end_activation` remains the explicit `Done/Fertig` operation. It alone marks the current ship `activation_complete=true`.

Because Wait permits out-of-order friendly activation, end-activation progression was hardened: after Done the authority searches circular initiative order for any other alive unfinished ship. A new round starts only when **no** unfinished live ship remains anywhere. Round start continues to reset:

- `activation_complete=false` on all live ships;
- `movement_current=movement_max`;
- every weapon `ready=true`;
- initiative is recomputed and the first active ship is selected.

This prevents an out-of-order Wait/Done sequence from incorrectly starting a new round while an earlier ship is still unfinished.

## HMI behavior
A visible `Warten` / `Wait` button was added next to `Fertig`.

- pressing `Warten` submits `battle.wait_activation` without a target; server selects the next unfinished friendly ship;
- clicking any unfinished own ship while another own ship is active submits the same command with that ship as explicit target;
- the existing active-ship effects then select, focus and show the authoritative green legal-move grid for the new ship;
- clicking a completed ship remains inspection/focus only because it is absent from `wait_target_ship_ids`;
- `Fertig` remains the only operation that permanently removes a ship from the current round;
- the bottom `Noch verfügbare Schiffe` list continues to contain all alive unfinished own ships.

Android/Chrome native tap-highlight is disabled on Tactical ship/roster/commit controls so the HMI's own blue selected-cell state is the only intended touch selection feedback.

## Automated regression coverage
Battle tests prove:

1. move a ship, Wait to another friendly ship, and Wait back: movement remains consumed and neither ship becomes complete;
2. fire a weapon, Wait away, then Wait back: the fired weapon remains `ready=false` and is not re-projected as a legal fire action;
3. Wait without explicit target deterministically selects the next unfinished friendly ship;
4. Done on a waited-to ship completes only that ship and does not advance the round while the original waited ship is unfinished;
5. enemy Wait targets are rejected with exact zero battle mutation;
6. existing normal 2v2 activation order/round reset regression remains green.

## Isolated browser acceptance from the user's current Triangle Battle
The current canonical turn6/revision15 Battle #1 was exported read-only and restored to isolated 7182.

Starting isolated state was Tactical round1, active Human Laser Scout #35 at 19/20. Pressing `Fertig` completed the last round-1 ship and correctly started round2:

- Human Scout #23: 20/20, unfinished;
- Human Scout #24: 20/20, unfinished;
- Human Laser Scout #35: 20/20, unfinished, Laser `ready=true`;
- active #23;
- server Wait targets `[24,35]`.

Then:

1. Scout #23 moved from its start cell using a cost-2 legal move -> 18/20;
2. `Warten` generated authoritative `activation_waited` #23 -> #24; #23 remained unfinished at 18/20;
3. clicking Scout #23 directly generated authoritative `activation_waited` #24 -> #23; #23 returned active at the same 18/20;
4. `Fertig` on #23 marked only #23 complete; round stayed 2 and #24 became active;
5. clicking Laser Scout #35 switched authority #24 -> #35; server projected its ready Laser Cannon with two legal Darlok targets and the UI showed `Laser Cannon · 2 Ziel(e)`.

This is the requested Wait/Done contract: a ship can be revisited until the player explicitly presses Done/Fertig, and its spent movement/weapon state survives every Wait switch.
