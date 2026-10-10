# Slice 17.8 - Tactical 2D mobile SVG complexity reduction

Date: 2026-10-10
Status: isolated Chrome QA complete; **on-device Android FPS assessment and personal user acceptance pending**.

## User-reported problem

On Android phone, Tactical Combat 2D pan, pinch zoom and scrolling still feel slow, even after the previous RAF-batched camera commit work (ce8451c). User asked to inspect both the number of graphics and how they are redrawn.

## Live profiling, BEFORE this change

Repository verified clean at ce8451c; canonical reference server 7171, services 7172/7173 left untouched; isolated tactical reference game ran on temporary 7174.

The agent-managed desktop Chrome DOM snapshot from the original 3 vs 3 reference battle showed **904 SVG nodes**, including:

- **280 rects**, of which **260** were transparent `tactical-move-cursor-cell` hit areas, one per legal move target;
- **334 paths**, including 72 blue/yellow/red ring charge paths and 72 tracks;
- **166 groups** (`g`), including 4 sector groups for each health layer;
- **78 computed CSS filters** (72 ring per-path glow filters and 6 procedural ship glyph filters) on desktop. The existing coarse-pointer CSS already disables those on mobile; we did not assume desktop and mobile have identical GPU behavior;
- **332 `vector-effect: non-scaling-stroke` elements** at the captured desktop camera.

The 260 invisible movement rectangles were redundant: click-to-move already converts the client position through the current SVG screen transform (`tacticalViewportCell`) and checks the authoritative `legalMoveByCell` map.

Armor and structure do **not** have directional HP pools. Each ring's 4 quarter-sectors was independently represented by four groups, four tracks and four charge paths, even though each used exactly the same armor/structure fraction. The user sees one global four-segment armor ring and one global four-segment structure ring.

## Implementation

1. **Removed the 260 invisible move hitboxes.** One authoritative screen-transform hit test handles a click on the battlefield; ship sprites/targeting retain their own hit areas. The green legal movement border path and teleport indicators remain unchanged. Removed obsolete CSS `.tactical-move-cursor-cell`.
2. **Compound four-quarter armor and structure paths.** Each global ring renders one compound SVG track path and one compound charge path with FOUR distinct `M ... A` subpaths, maintaining precisely the same four visual gaps and color. The global current/max/fraction remains in the layer metadata, and `data-global-pool` marks the combined paths. The blue shield remains four true directional sector groups with four individually measured charge fractions.
3. Kept previous `requestAnimationFrame` batched pan/pinch and wheel camera updates, coarse mobile glow suppression, ship focus badge, red occupied backgrounds, and accurate SVG hit-cell projection, without touching authoritative movement or combat rules.
4. Tightened the Tactical browser source contract to reject reintroduction of per-move invisible SVG cells and to assert combined global armor/structure geometry and independent shield sectors.

## Live isolated Chrome QA on 7174 AFTER change

| Metric | Before | After | Change |
|---|---:|---:|---:|
| All SVG elements | 904 | **524** | **-380 / -42.0%** |
| Rectangles | 280 | **20** | -260 |
| SVG paths | 334 | **262** | -72 |
| SVG groups | 166 | **118** | -48 |

- Initial 6 ships still rendered 18 rings: 6 outer blue directional shield rings with **24** real shield sector groups, and 12 global armor/structure rings with **two paths each**.
- The compound armor and structure paths each had exactly **four independent `M` subpaths** for visually unchanged separated quarters.
- Actual Heavy Particle Beam combat shot reduced Darlok ship #45 armor **500 -> 401**, and the new yellow charge path measured 3.14477 units against the 3.92112-unit full track (ratio **0.802**), matching 401/500; shield and structure remained 100%. Thus real damage changes ring length correctly.
- Actual click-to-move sent a server-accepted `ship_moved` command from Falcon #40 (25,9) to (27,9) using an actual battlefield DOM target; this worked WITHOUT move cursor rectangles.
- Simulated one-finger pan and two-finger pinch in the managed hidden Chrome browser (QA-only RAF timer shim) updated live `viewBox` and backing space rectangle during gestures; the React state remained committed at the end, and all 18 health rings persisted.
- After the pan + pinch, actual battlefield click-to-move from Falcon #40 (27,9) to (29,9) succeeded (`ship_moved`, movement to 12). Touch pointer drag-to-move (29,9)->(30,9) also succeeded (`ship_moved`, movement to 11). Camera screen-transform hit testing stayed aligned.
- All tests were run in isolated fixture 7174, not on canonical 7171 and not on the physical Android device. Therefore DOM-complexity savings are measured, **phone FPS gain is not measured or claimed**.

## QA and lifecycle

- `npm.cmd run build` (encoding, UI source contracts, TypeScript, Vite) PASS.
- `git diff --check` PASS.
- Targeted Go tests (`go test ./internal/battle ./internal/server -count=1`): PASS.
- `go vet ./...`: PASS.
- Keep Slice 17.8 OPEN pending the user testing their Android hardware; do not push, start Slice 17.9, or reset the canonical 7171 game.

## Future escalation if real phone still janks

The camera still changes the SVG `viewBox` on each animation frame. Even with a smaller scene and mobile CSS shadow removal, that may force a complete GPU re-rasterization. If Android remains slow, investigate a dedicated composited drawing layer/canvas or a two-phase camera preview (CSS composited transform while finger is down, authoritative SVG transform on release), profiling touch rendering on the device before choosing the larger architectural change.
