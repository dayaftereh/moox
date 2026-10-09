# Slice 17.8 – Tactical movement overlay review (Chrome)

Date: 2026-10-09
Status: implemented for user review; Slice 17.8 manual acceptance remains open.

## User request

The Tactical battlefield must not show red collision boxes or a blue selected-ship rectangle at rest, when aiming, or while animations play. Red occupied cells should be visible only while the player is dragging the currently active ship to move it. The ship sprite should retain its normal appearance without an extra blue selection halo.

## Before fix: isolated real Chrome on 7174

- Opened the permanent 3-vs-3 Combat Review fixture on an isolated `-reference-games` instance; canonical 7171 remained untouched.
- During an idle Human activation, SVG contained **five `tactical-occupied-cell` red rectangles**, **one `tactical-selected-cell` blue rectangle**, **one `tactical-reachable-grid` green path**, and **260 invisible movement-cursor hit cells**.
- `movementSelectionActive` previously exposed all the move guides throughout a Human activation rather than only during a movement gesture; selected/occupied overlays were unconditionally rendered whenever a friendly ship was selected.

## Implementation

- Removed the blue selection cell entirely, together with its unused CSS, and removed selected/active blue SVG ship glows. The standard procedural ship sprite remains intact.
- Red occupied cells, the green legal-movement grid, teleport markers and move-cursor hit areas render only while a valid drag of the active friendly ship is underway. The gesture begins after the existing 4px drag threshold.
- Releasing the drag on an authorized destination issues the existing `tacticalMoveCommand`. Dragging the ship never pans the camera; dragging empty space continues to pan. A canceled/invalid drop issues no movement.
- A second pointer cancels ship-drag preview so two-finger pan/pinch remains available; touch drag is supported.
- `movementDragVisible` is gated by `movementSelectionActive`, so during combat animation playback and non-Human turns all move overlays remain hidden.
- Legacy click-to-move to server-authorized cells remains available without persistent visual clutter. No movement legality rules or server state were changed.
- Updated EN/DE movement instructions and Tactical ordnance browser contract.

## Isolated Chrome after fix

- Idle Human activation: blue **0**, red **0**, green **0**, movement cells **0**. Selected active Falcon retains only normal ship artwork (no added blue halo).
- Mouse drag of active Falcon: blue **0**, red **5**, green **1**, legal movement cursor cells **260**. On cancel: all overlays **0**.
- Actual server-authorized drag/drop: Falcon MK1 #40 moved from **(25, 9)** to **(26, 9)**; authoritative `ship_moved` event sequence **2**. Overlays disappeared on release.
- Empty-space pointer drag continued to pan the viewBox with no red/blue/green overlay.
- Synthetic touch drag activated movement guides; cancellation removed them without a movement command.
- Human Zeon missile shot generated a live projectile animation while the presentation lock was true. Animation samples recorded maximum red **0**, blue **0**, green **0** throughout.
- QA was run on separate reference-game server 7174; no gameplay input was issued to canonical 7171.

## Verification

- `web/npm run build`, including Tactical ordnance HMI contract and TypeScript: PASS.
- `git diff --check`: PASS.
- Manual user visual acceptance: pending. No Slice 17.9.
