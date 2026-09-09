# Slice 15.5 Gate 3 Block 2 - Participant-safe Tactical projection / API command surface

Date: 2026-09-10

## Goal

Expose the Gate-3 Block-1 Tactical authority to participants and the browser without leaking future RNG state or another seat's current command catalog. Provide concrete read-only Scan data, server-projected movement/fire actions and typed browser command helpers before building the battlefield UI.

## Seat-specific Battle projection

`battle.Session.PlayerView(seatID)` now derives participant Battle state from the full authoritative Observer view.

For participant views:

- the current Tactical RNG state is redacted;
- the initial Tactical RNG state is redacted;
- player-safe Scan/status data for all combat ships in the visible supported Battle is retained;
- legal movement/fire actions and `can_end_activation` are retained only when the active ship belongs to that seat;
- a non-active participant receives no opponent command catalog;
- non-participant requests reject.

`GameSession.PlayerBattleViews` now uses this Battle-level player projection rather than returning the raw child `View()` to participants.

Observer/live authority still retains full RNG state for deterministic persistence/replay.

## Tactical Scan/status projection

`TacticalView.ships[]` now combines immutable supported ship configuration with current runtime state. Each participant-visible ship exposes:

- stable ship / empire / seat identity;
- hull, drive, computer and armor identity;
- current X/Y and Facing;
- MovementCurrent / MovementMax;
- activation-complete / destroyed state;
- Armor current/max;
- Structure current/max;
- Beam Offense / Beam Defense;
- supported weapon list with damage range and current readiness.

Structure current is derived from authoritative StructureMax minus StructureDamage. No internal-system health is fabricated because that subsystem is not yet normalized. No client-side aggregate combat-strength score is introduced.

This projection is deliberately suitable for the frozen read-only Tactical Scan UX in Block 3.

## Legal fire projection

The server now projects `legal_fire_actions[]` alongside `legal_moves[]` for the active controlling seat.

Each supported fire action includes:

- ship ID;
- weapon slot;
- weapon ID;
- player-safe target IDs;
- authoritative current range index for each target.

Only live enemy targets inside the evidenced Laser range table are projected. Readiness and current runtime positions are authoritative. The browser therefore does not need to recreate target/range legality.

## HTTP / browser command surface

The existing generic endpoint remains:

`POST /api/v1/games/<gameID>/battles/<battleID>/commands`

`web/src/api.ts` now has concrete Tactical types instead of `unknown` for Battle spec/view state, including Scan ships, legal moves, legal fire targets/actions and activation state.

Browser helpers added:

- `submitBattleCommand(...)`;
- `tacticalMoveCommand(...)`;
- `tacticalFireBeamCommand(...)`;
- `tacticalEndActivationCommand(...)`.

Command sequence comes from the latest authoritative Tactical view. These helpers only construct/transport intent; they do not calculate movement legality, resulting facing, range, hit or damage.

## End-to-end server regression

The HTTP Tactical command regression now checks the participant snapshot contract itself:

1. strategic turn submission creates a Tactical Battle;
2. participant snapshot contains Scan ships and active legal move/fire projection;
3. player snapshot exposes no current or initial RNG state;
4. a server-projected legal move is submitted through the real Battle command endpoint;
5. the next participant snapshot reflects authoritative X/Y/Facing/remaining movement;
6. the post-move server-projected Laser action is submitted with the next sequence;
7. change sequence and Tactical command sequence advance normally.

Battle-core regression separately confirms that the inactive participant receives Scan data but zero legal moves/fire actions and cannot end the active ship's activation.

## Validation

Passed after implementation:

- `go test ./internal/battle ./internal/session ./internal/server`;
- `npm run build`.

Full repository validation is run again before the Block-2 commit.

## Boundary to Block 3

Block 2 intentionally does not render the battlefield. Block 3 can now build the desktop/mobile Tactical UI entirely from typed participant-safe projection:

- ship sprites/markers from `tactical.ships`;
- active-ship highlighting;
- move overlay from `legal_moves`;
- target overlay from `legal_fire_actions`;
- Scan panel from current ship view;
- actions submitted through the typed helpers.

No Tactical simulation rule needs to be duplicated in React.
