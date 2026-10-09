# Slice 17.8 - Selected Ship Shield, Armor and Structure Bars

Date: 2026-10-09
Status: implementation and automated QA complete; manual User Review pending.

## User request

Show three small percentages above the blue-selected Human ship, vertically stacked: **blue shield**, **yellow armor/hull** and **red structure**. Keep the combat arena uncluttered for nonselected ships and during animations.

## Implementation

- `TacticalSelectedShipVitals` is a small non-interactive SVG group anchored at the currently selected ship coordinates. It appears just above the blue movement-selection square, not above every fleet ship.
- Visibility uses the existing `movementSelectionActive && selectedShip` gate, so the three bars follow the same rules as the blue selection square: active selected Human ship with movement points and server-authorized moves, no presentation lock or pending command, not scanning or disabled.
- The blue percentage is `sum(max(0, shield_arc_current[i])) / (shield_arc_max * 4) * 100`.
- Yellow and red percentages use authoritative `armor_current / armor_max` and `structure_current / structure_max`. The already-present clamped `meterPercent` helper rounds the display to whole percent; a max of zero gives 0%. The UI never infers these totals from effects or animation events.
- Each meter retains a dark track at zero, independently fills proportionally, and shows a numeric percentage. The floating UI uses a dark panel and `pointerEvents=none` to preserve movement, targeting and viewport pan interactions.
- Existing four-direction shield arcs, selection squares, camera behavior, combat outcomes and server rules were not changed.

## Isolated real-Chrome QA

QA server: fresh `-reference-games` fixture on 127.0.0.1:7174. Canonical 7171 in-memory game was not touched.

1. Fresh 3-vs-3 review, selected Falcon MK1 1 ship ID 40: exactly one blue selection rectangle and one three-bar SVG panel, displaying Shield 100%, Armor 100%, Structure 100%. Browser computed fills: blue `rgb(104,216,255)`, yellow `rgb(255,196,94)`, red `rgb(255,90,87)`.
2. DOM bounding geometry: status panel 73 x 41 screen pixels, centered 14 pixels above the blue selection square; no overlap with the ship frame.
3. Heavy Particle Beam fired: all status bars and the blue square disappeared throughout presentation input lock and returned after the animation finished.
4. Clicking Finish switched active ship from Falcon #40 to Falcon #41; exactly one health overlay moved with selection, without stale overlays.
5. Finishing the remaining Human activations let the real AI attack. When the next Human activation became available, authoritative Falcon #40 state was shield arcs `[0,100,100,100]` (300/400), armor 0/500, structure 12/50. The floating values automatically read **75% blue**, **0% yellow** (empty track, still readable `0%`) and **24% red**. Rendered fill widths were 1.95, 0 and 0.624 of a 2.6-unit track respectively.
6. The health panel and blue selection square were absent while the AI presentation queue was locked; after the queue drained, they reappeared together.

## QA gates

- `web/npm run build`: PASS, includes TypeScript production build and extended Tactical ordnance/HMI contract.
- `go test ./internal/battle ./internal/server -count=1`: PASS.
- `git diff --check`: PASS.
- No Git push, no reset of canonical 7171, no new slice. Slice 17.8 remains OPEN pending explicit User approval.
