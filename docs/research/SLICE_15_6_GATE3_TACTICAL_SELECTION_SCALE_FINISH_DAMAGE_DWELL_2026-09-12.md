# Slice 15.6 Gate3 - Tactical selection, hull scale, Finish and damage dwell

Status: **IMPLEMENTED / isolated real-battle browser QA green**
Date: 2026-09-12

## User intent
This follow-up tightens the Tactical HMI around direct ship selection rather than a large bottom roster. The requested visual contract is:

- click an own ship/cell to select and focus it;
- selected ship cell is a full blue square, not a ring;
- occupied ship cells are red outlines because they are blocked movement destinations;
- legal movement remains the crisp green server-projected grid;
- no permanent Tactical grid and no active/target/scan selection circles;
- small hulls stay visibly small inside one movement cell; Doom Star is the maximum one-cell visual;
- damage numbers remain readable for roughly four seconds;
- the current ship activation is completed with a concise `Fertig` action.

## Authority boundary
The current Battle authority has one `active_ship_id` and accepts `battle.end_activation` only for that exact ship. There is no authoritative `select_ship`, `wait_activation`, or arbitrary initiative switch command yet.

Therefore this UI does **not** fake ship activation in React:

- any live own ship can be selected and camera-focused locally;
- movement/status for the selected own ship is shown;
- the green movement grid appears only when the selected ship is also the server-active ship because only that ship has participant-safe authoritative `legal_moves`;
- clicking an inactive own ship does not create client-side legal moves and does not change initiative;
- when the server advances to the next own active ship, that ship is selected and focused automatically;
- `Fertig` is deliberately only a new HMI label for the existing authoritative `battle.end_activation` command. It ends the **current ship activation**, not the entire Tactical side/round.

Moves and beam commands remain immediate authoritative commands and are visible through the shared Battle state as before.

## Selection and cell presentation
`TacticalBattlefield` now keeps an explicit `selectedShipID` for own ships.

The exact selected cell is rendered as one 1x1 blue filled square. Every other non-destroyed ship position is rendered as a 1x1 red outlined occupied cell. The green reachable-grid path remains derived only from `tactical.legal_moves`.

The ship hit target is also exactly one unrotated 1x1 cell. Rotation is now applied only to the ship drawing in a nested facing group, so clicking the cell does not rotate or distort the interaction target with the ship facing.

All active/target/scan circle elements were removed from the Tactical DOM and their dead CSS was removed.

## Hull-relative SVG scale
Tactical rendering now resolves the visual hull from the persisted genome first:

`visual_genome.hull_id ?? tactical gameplay hull_id`

This matters for the fresh starting Scout: gameplay currently uses the Frigate rules baseline, while the persisted visual genome correctly identifies it as `scout`.

The existing shared hull footprint scale is now used directly as the Tactical one-cell visual envelope:

- Scout: `0.40`
- Frigate: `0.48`
- Destroyer: `0.58`
- Cruiser: `0.68`
- Battleship: `0.78`
- Titan: `0.90`
- Doom Star: `1.00`

The Doom Star therefore owns the full one-cell maximum envelope. A new `tightShipViewBox` derives the internal SVG viewBox from the persisted hull geometry, primitives, engines and hardpoints instead of using the old broad fixed minimum margins. This makes footprint represent visible ship scale rather than a mostly empty SVG container.

The whole 1x1 cell remains the click target regardless of hull visual size.

## Camera / remaining-ship HUD
Selecting an own ship focuses the Tactical camera on that cell with a minimum zoom of `2.2` rather than `1.45`.

The bottom roster is now `Noch verfügbare Schiffe` / `Ships still available` and contains only own ships that are alive and have not completed their current activation. Completing a ship with `Fertig` removes it from that strip and the next authoritative own active ship is automatically selected.

## Damage readability
The floating layer-damage number now uses a four-second animation and remains at full opacity for most of that interval. The containing beam-feedback state is kept for 4.2 seconds so React does not remove the damage number early.

Layer colors remain authority-derived:

- shield: cyan, future-ready until Tactical shield authority exists;
- armor: amber;
- structure: red.

The reference Triangle starting Scouts are unarmed, so this block does not claim a real browser-fired damage event. Runtime browser CSS QA proves the loaded damage-number style uses `tactical-damage-float` with `animation-duration: 4s`.

## Real Battle browser QA - isolated 7179
A fresh Triangle reference game was driven through the normal browser flow to Battle #1 at Human Home, Human 2 Scouts vs Darlok 2 Scouts.

Authoritative initial Human active Scout had `20/20` Nuclear movement and the participant view projected 515 legal moves.

Observed DOM/UI evidence:

- exactly 1 blue selected 1x1 cell;
- exactly 3 red outlined occupied 1x1 cells;
- exactly 1 deduplicated green reachable-grid path;
- 0 active/target/scan circle elements;
- Scout hit cell approximately 27.1 x 27.1 px and remains the full cell;
- persisted Human Scout uses `data-hull-id="scout"`, `data-footprint="0.40"`, Tactical SVG width/height `0.4` cell;
- tight-viewBox visible Human Scout envelope approximately 9.6 x 8.0 px inside the 27.1 px cell at the QA camera zoom;
- Tactical camera viewport is approximately 20 cells wide after focus;
- visible commit action is `Fertig`;
- runtime damage CSS reports `tactical-damage-float`, `4s`, font size `0.62px`.

Selection behavior was exercised:

1. selecting inactive own Scout 2 moved the blue cell and camera to Scout 2, status became `Scout 2 - 20/20`, and the green grid disappeared because Scout 2 was not authoritative active;
2. selecting active Scout 1 moved focus/blue cell back and restored its authoritative green grid;
3. pressing `Fertig` sent `battle.end_activation`; Scout 1 became activation-complete and disappeared from the remaining-ship strip; Scout 2 became server-active, automatically selected blue, and received its green legal-move grid.

This preserves Battle authority while providing the direct-selection HMI requested by the user.

## Deferred authority work
A future Battle authority block may add original-style `Wait` / explicit arbitrary activation selection semantics. Until that exists, the HMI must not make an inactive ship actionable merely because it is selected for inspection/focus.
