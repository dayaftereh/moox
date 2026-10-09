# Slice 17.8 – Three concentric ship condition rings (Chrome QA)

Date: 2026-10-09
Status: implemented, real-Chrome tested, manual User Review still pending.

## User request

Keep the already-liked blue four-sector shield circle and show armor/hull (yellow) and structure (red) as two further concentric circles. Display the rings on **all ships**, both Human and enemy, without requiring scanning or focusing. Remove the old shield/armor/structure microbars inside the blue/red movement-selection square, but retain its selection and legal-movement behavior.

## Source of truth and semantics

- Outermost **blue shield** radius 0.95: actual independent shield charge in authoritative `shield_arc_current[0..3]`, relative to `shield_arc_max`. Sector directions preserve original 0 Front / 3 Right / 2 Rear / 1 Left; each arc rotates with ship facing. No shield ring if the ship lacks shields or all four sectors have zero remaining charge.
- Middle **yellow armor/hull** radius 0.78: authoritative `armor_current / armor_max`, repeated at the SAME fraction across four separated quadrants. There are not four separate armor hitpoint pools in this server protocol.
- Innermost **red structure** radius 0.61: authoritative `structure_current / structure_max`, also one global hitpoint pool repeated across four quadrants; do not misrepresent the four arcs as directional structure damage.
- Each ring contains a faint full quarter-track plus a charge path whose actual SVG arc length reflects the positive current fraction. When armor or structure reaches zero, the empty track remains, but the colored charge is absent. Destroyed ships display no status rings, even when old shield values survive in historical snapshots.
- The three rings are within each `g.tactical-ship`, rotated to its `facing * 22.5`, with `pointerEvents=none` and `aria-hidden`; normal ship selection/targeting and the player's red/blue selection square continue working.
- The old `TacticalSelectedShipVitals`, its SVG microbars, and related CSS are removed. Existing ship scanning still shows precise numerical values independently.
- No Tactical combat, server, event, damage, aiming, movement, or camera rules changed.

## Independent real-Chrome reference QA — port 7174

1. On a fresh 3 Humans versus 3 Darloks battle, Chrome showed 6 ship ring groups, 18 colored rings and 72 quarter-sector nodes (4 per layer per ship), with no legacy mini status bars.
2. Browser-computed colors: blue Shield rgb(104,216,255), yellow Armor rgb(255,196,94), red Structure rgb(255,90,87). The rings are physically nested (about 48.5 / 39.8 / 31.2 screen pixels across at tested zoom).
3. A real Particle Beam volley against Darlok ship #45 changed armor 500/500 -> 401/500. Browser reported the yellow ring fraction **0.802** on ALL four quadrants and their arc geometry visibly shortened to 80.2%, while the four blue directional shields and red structure remained at 100%.
4. A real legal click-to-move from (25,9) to (26,9) generated server `ship_moved`; the three rings followed the ship and normal selection/movement grid remained usable.
5. Ending three Human activations triggered actual Rocket/AI action. Darlok ship #45 subsequently had authoritative shield arcs [100,100,75,100] and the blue shield ring showed total 0.938 with the rear sector 0.75 while all three other sectors remained 1.00.
6. After AI killed Falcon #40, that destroyed ship had zero condition-ring groups (the five remaining live ships continued to display their rings).

## Read-only Chrome against canonical 7171

No game commands, restart or fixture reset. A READ-ONLY browser loaded the newly rebuilt Web bundle, with no change to existing tactical command sequence.

- All six live ships had three rings and the active ship's original selection rectangle still appeared.
- Enemy Darlok #45 had authoritative Shield [0,100,100,100] = 300/400, Armor 0/500, Structure 32/50. SVG inspection confirmed:
  - blue shield: Front sector 0.00 with no charge path, Right/Rear/Left 1.00, total 0.750;
  - yellow armor: all four global-pool quarters 0.00, dark tracks visible, zero yellow charge;
  - red structure: all four global-pool quarters 0.64, the measured SVG charged arc is ~0.49058 while the full red quarter-track is ~0.76646 (ratio 0.64).
- No selected-ship numeric microbars remained.

## Automated gates

- `web/npm run build` (including TypeScript, Tactical browser contracts, Vite production assets): PASS.
- `git diff --check`: PASS.
- `go test ./internal/battle ./internal/server -count=1`: PASS (focused tactical/server).
- `go vet ./...`: PASS.
- `go test ./... -count=1`: many packages PASSED, including internal/battle, internal/server and internal/app, but internal/session remained running for more than four minutes; the full run was canceled with no reported failure in completed packages. The full-suite gate is NOT claimed as passed.
- Explicit user acceptance of Slice 17.8 still pending, no push, no new Slice 17.9.
