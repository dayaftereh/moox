# Slice 17.8 - Unobtrusive tactical focus marker and occupancy underlay

Date: 2026-10-09
Status: implementation / automated QA complete, explicit visual User Review pending.

## User request and design decision

With three concentric condition rings (outer shield, middle armor, inner structure), thick red occupied-field outlines and blue/red 1×1 selection-frame outlines visibly intersect the health circles. The User approved the following redesign for Slice 17.8 without increasing grid cell size:

- Keep the three concentric rings on every live ship untouched.
- On move-enabled activations, show occupied ship tiles as **very faint red backgrounds without any stroke/border**, drawn behind all ships and rings.
- Keep the green legal-move grid and click/drag-to-move interactions exactly as they are.
- Focus the active own selected ship using a **small colored pointer below its outer ring plus its stable own-ship number (1, 2, 3...)**, not a thick selection box.
- The selected ship's 1×1 cell may retain a very subtle blue / red / neutral fill behind the rings but no visible outline.
- Cyan pointer = active and legal movement available; coral/red pointer = selected, no legal movement, but combat actions available; muted pointer = selected and no action available.
- Hide selection overlays and pointer during animation/presentation lock or when ownership is not the current player. Keep state/camera, combat, shield and damage rules unchanged.

## Code changes

- `web/src/TacticalBattlefield.tsx`: compute the selected own ship's stable ordinal by sorting all own ships by ship_id (destroyed ships still count). Add a single `tactical-ship-focus-marker` SVG group **after all ship sprites/rings** with `data-ship-id`, `data-ship-number`, `data-focus-kind`, pointer triangle, dark numeric plate, and non-interactive pointer events.
- `web/src/TacticalBattlefield.tsx`: preserve the existing focus and movement gates. Selected tile retains subtle background only; occupied rectangles now have slight corner rounding. Occupied tile rendering remains before ship and health-ring rendering.
- `web/src/styles.css`: selected-cell and occupied-cell outlines changed to `stroke: none`; occupied cell uses rgba(255,87,85,.14). Pointer / number marker receives state color using CSS classes `is-movement`, `is-attack`, `is-idle`, with no hitbox.
- `web/src/i18n.tsx`: English/German viewport instructions now describe the pointer+number and subtle occupied fields. UTF8 encoding check preserved.
- `web/scripts/check-tactical-ordnance.mjs`: contract assertions updated to reject heavy tile outlines and assert numbered active focus, stable ordinal and subdued occupied backgrounds.

## Isolated real Chrome QA (port 7174)

Game fixture: `game-triangle-combat-review`; spawned independent reference-game instance, untouched canonical game.

1. At initial Round 1 Falcon MK1 #40, the DOM showed exactly **one** selected-ship pointer: ship=40, index=1, mode=movement, cyan rgb(130,226,255). Six ships retained **18 rings** in total, with the original green legal-move grid and 260 legal destinations.
2. Five occupied-cell underlays had computed fill rgba(255,87,85,0.14), `stroke=none`; the selected cell had fill rgba(73,157,233,0.09) and `stroke=none`. The selected ship's health rings did not acquire additional borders.
3. The selected focus marker bounding rectangle started below the outer shield-ring rectangle; marker displayed only a tiny triangle and the ship number, with `pointerEvents=none`.
4. Actual click-to-move from (25,9) to the server-authorized (26,9) generated authoritative `ship_moved` event seq 2. Falcon #40 moved to (26,9), movement fell from 16 to 15 and the marker remained on the ship with index=1 and kind=movement.
5. A second actual move from (26,9) to (41,9) cost 15 points, leaving movement=0 and **zero** legal move cells. Falcon #40 still had 7 available projected weapons; its marker remained visible and became `data-focus-kind=attack`, with coral/red rgb(255,129,119). No green movement grid, no occupied-cell overlays, no selection box outline.
6. Actual Heavy Particle Beam firing remained possible with movement=0. During the resulting presentation/input lock the marker was absent; after the presentation queue drained it reappeared in attack mode.
7. Clicking Finish transitioned the active own ship to Falcon #41, focus number **2** (cyan). The next Finish transitioned to Falcon #42, focus number **3**. No stale marker was left on the prior ship.

## Canonical read-only review (port 7171)

A separate managed Chrome session opened the already-running canonical instance; entry to the Tactical battle was navigation only. **No tactical commands, moves, attacks, reset or server restart** occurred.

- Existing canonical battle was at round 1, active Human ship #42, next command sequence 49.
- New web asset loaded (`main-DvC5QKdA.js`). A single focus marker on ship42 displayed own-ship index **3**, cyan movement mode.
- Five occupied underlays still used faint red rgba(255,87,85,0.14) with `stroke=none`; selected cell used transparent blue with `stroke=none`. All 18 condition rings remained on six ships.

## QA gates and restrictions

- `web/npm run build`: PASS (UTF8/mojibake check, Tactical browser contract, TypeScript and Vite).
- `go test ./internal/battle ./internal/server -count=1`: PASS.
- `git diff --check`: PASS before staging.
- No Canonical 7171 reset/restart, no push, no Slice 17.9. Slice 17.8 remains OPEN until personal user visual acceptance.
