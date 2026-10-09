# Slice 17.8 — Tactical ship selection must survive exhaustion of legal moves

Date: 2026-10-09
Status: implemented and verified on the **canonical persisted user's game** and isolated Chrome fixture; manual User Review still pending.

## User report

Falcon MK1 1 (own ship 1) cannot be visibly focused any more, while Falcon MK1 2 and 3 remain selectable. Falcon 1 has one movement point and usable weapons left, so the user expects a red selection square when firing remains possible even without a visible movement grid.

## Verified canonical state BEFORE the fix (READ-ONLY Chrome on 7171)

- Persisted `game-triangle-combat-review`, battle 1, round 1; authoritative active ship ID **40** (Falcon MK1 1), at `(37,4)`, heading/facing **15**.
- `movement_current = 1`, `movement_max = 16`, `activation_complete = false`, not destroyed; next command sequence **26**.
- Server supplied **zero** `legal_moves` (omitted empty field from JSON) and **three** `weapon_actions` with `available = true` (the user's four count did not match this projection).
- React DOM contained `g.tactical-ship.own.active.selected` for Falcon 40, but **zero** `tactical-selected-cell` rectangles and **zero** `tactical-selected-ship-vitals` groups.
- Root cause: `movementSelectionActive` was used as the visibility gate for BOTH movement aids AND ship selection, and required `legalMoves.length > 0`. The active ship stayed selected internally but the user had no visible selection confirmation or status bars.
- Movement reasoning: server move cost is translation plus turning cost (`tacticalMoveCost` in `internal/battle/tactical.go`), so a lone movement point does not guarantee a legal destination when the ship must rotate. No movement rules were changed.

## Fix

- Introduced `selectionFocusActive` independent of movement-points/available move destinations, gated by active selected Human ship, unfinished activation, unlocked presentation, no in-flight command, and combat (not scan) mode.
- Kept `movementSelectionActive = selectionFocusActive && movement_current > 0 && legalMoves.length > 0`: green legal movement grid, occupied red squares, clickable move destinations, teleport markers and drag-to-move are still based only on **server-authorized legal moves**.
- The selected ship's focus square is shown if `selectionFocusActive`: **blue** when legal movement exists; **red** (`data-focus-kind="attack"`) when no move is legal but a weapon/boarding action is available; subdued gray (`idle`) when neither is available.
- The three inline blue/yellow/red condition microbars now follow `selectionFocusActive` rather than `movementSelectionActive`, so they stay with an attack-only focused ship. No extra floating HUD, game-state modification or camera change.
- During animations and non-Human turns all focus squares and status bars remain hidden.

## Real Chrome test evidence

**Canonical 7171, no tactical command sent:**

1. Reloaded the exact persisted battle after production web build. Active Falcon 40 retained one movement point, zero legal moves and three available weapon actions.
2. DOM showed exactly **one** selected-cell with class `tactical-selected-cell is-attack-ready`, `data-focus-kind="attack"`, computed red stroke `rgb(255,112,102)` and translucent red fill.
3. One inline status group was visible for ship 40, with authoritative Shield 100%, Armor 100%, Structure 100%, and no green movement-grid or occupied red cell overlays.
4. Clicking on Falcon MK1 1 via its own ship marker preserved the visible `attack` focus and three status strips. Server command sequence remained at **26**; no command or game reset was triggered.

**Isolated 7174 fixture:**

5. Fresh Falcon 40 with 16 movement points and 260 legal cells displayed the prior BLUE selection frame (`data-focus-kind="movement"`), five occupied red cells, green movement grid and three inline status strips.
6. Click-to-move `(25,9)` to legal `(26,9)` still issued authoritative `ship_moved` event sequence 2, left 15 movement points, and kept the focus BLUE and aligned to the new ship position.
7. Heavy Particle Beam shot at an actual Darlok ship: presentation/input lock hid the focus, red occupied cells, green movement grid and all health bars; after the lock ended the selection and movement grid returned.

## QA

- `web/npm run build` including TypeScript and Tactical browser contract: PASS.
- `git diff --check`: PASS.
- Targeted Go tactical/server tests: go test ./internal/battle ./internal/server -count=1 PASS.
- Canonical 7171 not reset or restarted; no unrequested push. Slice 17.8 OPEN; Slice 17.9 not opened.
