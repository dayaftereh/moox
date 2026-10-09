# Slice 17.8 - Tactical selection and movement grid correction (Chrome QA)

Date: 2026-10-09
Status: implemented and verified; user visual acceptance remains pending.

## User clarification

Commit a556207 had hidden red occupied squares and the green legal-move grid unless the player was physically dragging a ship; it had also removed the blue selected-ship square. This eliminated the visible click-to-move grid while an active ship was selected. The actual requested behavior is:

- Show the blue selected-ship square, red occupied-cell squares, and green legal-move grid while the active Human ship is selected, its activation is not completed, movement points remain, and the server exposes legal move destinations.
- Hide all these overlays while animations or pending commands disable player actions, when the player does not control the active ship, when there is no movement left, when no legal destination remains, or after the ship finishes its activation.
- Retain both normal click-to-move and optional drag-to-move. Do not make overlays depend on an ongoing drag gesture.
- Keep the normal procedural ship artwork without an additional blue glow (the blue selected-cell rectangle is restored).

## Browser implementation

In `web/src/TacticalBattlefield.tsx`, `movementSelectionActive` now requires all of:

- no presentation lock, command disable or in-flight command;
- Human own activation, combat mode, selected ship matches the active ship;
- active ship not `activation_complete`;
- `activeShip.movement_current > 0`;
- authoritative `legalMoves.length > 0`.

The selected blue square, red occupied squares, green legal-move path, click hit cells and teleport indicators use the same gate. Removed obsolete `movementDragVisible` and unused drag-indicator state; preserved the actual drag-to-move pointer handler. Restored `.tactical-selected-cell` CSS. Updated English and German hints and the Tactical browser contract. No server rules, movement legality or camera follow logic changed.

## Real Chrome QA on isolated reference server 7174

Permanent Combat Review game `game-triangle-combat-review` was launched from an isolated `-reference-games` server. Canonical port 7171, its browser/game state and in-memory battle were not reset or modified.

- Initial Human Falcon MK1 1 at (25, 9), 16/16 movement points, 260 legal moves: blue 1 / red 5 / green 1 / movement hitboxes 260. Blue selected cell was rendered with restored style.
- Human Zeon Missile fired at Darlok ship #45: during projectile animation and presentation input lock, maximum blue/red/green counts were 0/0/0. Once unlocked with 16 movement points, counts restored to 1/5/1.
- Clicked legal move cell (41, 9) using the actual viewport coordinate transform; authoritative `ship_moved` event sequence 4 moved Falcon MK1 1 from (25, 9) to (41, 9), using all 16 movement points. Its selection remained on the ship, but blue/red/green overlays were immediately 0/0/0 and `legal_moves` had no entries.
- Clicked the usual `Fertig` command: Human Falcon MK1 2 became active with 16 movement points and selection automatically changed to that ship. Blue/red/green returned to 1/5/1.
- The click-to-move and drag-to-move workflows are both preserved; this change restores discoverability of movement targets without interfering with camera panning.

## QA

- `web/npm run build` including encoding, TypeScript and Tactical ordnance contract: PASS.
- `git diff --check`: PASS.
- Additional targeted Go tests: go test ./internal/battle ./internal/server -count=1 PASS.
- Manual user approval: pending. Keep Slice 17.8 OPEN. Do not open Slice 17.9.
