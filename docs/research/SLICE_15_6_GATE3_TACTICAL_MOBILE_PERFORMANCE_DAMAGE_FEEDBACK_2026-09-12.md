# Slice 15.6 Gate3 - Tactical mobile performance, ship readability and damage feedback

Status: **IMPLEMENTED / browser-QA green for movement + rendering**
Date: 2026-09-12

## User feedback addressed
The immersive Tactical screen was accepted as the right direction, with four concrete polish issues:

- phone pan/zoom felt janky, especially while hundreds of legal-move squares were visible;
- ships looked like badly centered rectangles instead of readable ship silhouettes;
- Scout movement range felt much too large;
- beam impact feedback was too fast and did not expose where the damage was actually applied.

## 1. Reachable movement is now one grid path, not hundreds of SVG rectangles
The old Tactical HMI materialized every server-projected legal move as its own interactive `<g><rect/></g>`. Previous QA had **515** legal move cells for one activation, so every camera update kept a very large SVG subtree alive. The rounded rectangles were not the grid itself; the small `rx` corners were part of those per-cell target rectangles.

The new renderer keeps server authority unchanged but changes presentation:

- React receives the exact same authoritative `legal_moves` list.
- It deduplicates the four cell edges into one SVG path string.
- One `.tactical-reachable-grid` path draws the reachable grid with a stronger green stroke and `shape-rendering: crispEdges`.
- There are no rounded per-cell polygons and no per-cell DOM listeners.
- A click on the battlefield is converted to the nearest Tactical integer cell and accepted only when that cell exists in the server-projected legal-move map. React still does not compute movement legality.

### Browser evidence
Isolated Triangle 2v2 Tactical on port 7177:

- `.tactical-move-cell`: **0**
- `.tactical-reachable-grid`: **1**
- Tactical ship groups: **4**
- total descendants inside the Tactical SVG viewport: **163**
- current Nuclear Scout legal moves at 10 movement points: **63** initially
- after clicking server-legal cell (11,9), ship #23 moved authoritatively from (10,9) to (11,9), movement 10 -> 9, and the next server projection contained **47** legal moves while the HMI still rendered one grid path.

The managed QA browser is ~791x605 and cannot emulate the user's exact physical phone GPU through the available browser controller, so this block does not claim a measured handset FPS number. The high-cost DOM structure responsible for the obvious pan/zoom pressure has nevertheless been removed.

## 2. Baseline Tactical movement corrected from drive maximum to drive minimum
Ruleset values are currently:

- Nuclear Drive: min speed **10**, max speed **20**
- Fusion Drive: min speed **12**, max speed **22**

The Slice 15.5 baseline had been materializing every ship at `drive.MaxSpeed`, even though MOOX does not yet model the extra engine/maneuver allocation that should justify reaching that ceiling. That made an ordinary Nuclear Scout start every activation at 20 movement points.

The baseline now starts at `drive.MinSpeed`:

- Nuclear baseline movement: **10**
- Fusion baseline movement: **12**

`MaxSpeed` remains available in the rules as a future design/engine ceiling. The hard-coded Battle validator and all Tactical movement/scan regression expectations were updated to the same 10/12 authority contract. An isolated browser replay confirmed the Human Nuclear Scout displays **10/10**, not 20/20.

During QA this exposed a stale Slice-15.5 validator that still required exactly Fusion=22/Nuclear=20 and caused encounter drive to stop in `strategic_resolution`. Updating that validator to the new authoritative baseline restored the normal browser flow: Round1 -> Darlok contact -> declare war -> Fertig -> Battle #1.

## 3. Ships are native Tactical SVGs, larger and centered
The Tactical board previously wrapped `ProceduralShipGlyph` in SVG `foreignObject`. Combined with the generic glyph footprint CSS transform, that made the silhouette very small and visually offset under pan/zoom.

The board now embeds the glyph as a native nested SVG:

- no Tactical `foreignObject`;
- explicit centered Tactical viewport around each ship;
- `preserveAspectRatio=xMidYMid meet`;
- a Tactical-only tight viewBox crops the large generic empty SVG margins without changing the generated hull geometry;
- no root CSS footprint transform in the Tactical nested-SVG case;
- active/target/scan rings remain independent of the art;
- hit target is larger than the visual hull for touch usability.

Final browser geometry at the QA zoom:

- four Tactical ships -> four `.tactical-ship-vector.procedural-ship-glyph` elements;
- Tactical ship `foreignObject` count: **0**;
- starter Frigate/Scout rendered silhouette about **32.3 x 24.3 px** at ~26 px/grid-cell scale;
- visual center differs from the Tactical ship origin by only about **1.8 px horizontally / 0 px vertically**, rather than the previous >200 px transformed offset.

### Persistence truth
The Tactical projection already preserves and projects `ship.visual_genome` and its source visual revision when a ship has a persistent design genome. Existing Go regressions cover that path. The current Triangle **starting Scouts do not have a stored visual genome**, because New Game creates the starting Scout designs before a player has saved visual customization. For those ships the HMI intentionally uses the deterministic procedural fallback seed. Newly produced/customized designs can carry the persistent genome. This block improves rendering of both paths; it does not pretend the starter ships already have persisted art.

## 4. Slower impact + authority-derived floating damage numbers
The beam visual lifetime was slowed from ~0.62 s to roughly 1.35-1.6 s so an impact can be read instead of flashing away.

For a hit, the HMI correlates `beam_fired` with the authoritative `battle_damage_applied` event from the same command sequence and derives actual layer deltas from the server's before/after values:

- **Shield**: cyan/blue floating number (future-ready; no Tactical shield authority exists yet)
- **Armor**: amber/yellow floating number
- **Structure**: red floating number

Numbers rise from the target and fade over about 1.6 s. If an older event has only a total hit damage and no layer breakdown, the UI falls back to one red total-damage number rather than inventing a layer.

Current server damage authority is Armor -> Structure. Therefore this block can browser-test the rendering pipeline and authority mapping in code, but it does **not** claim current Shield combat support. The reference Triangle starter Scouts are unarmed, so no fabricated browser laser shot was used as acceptance evidence.

## Authority / non-goals
- Legal moves, movement cost, facing, initiative and damage remain server-authoritative.
- This block does not implement staged Tactical activation commit/batching.
- This block does not implement no-selection fire-all semantics.
- This block does not introduce shields into the combat model.
- Persistent visual generation for untouched New-Game starting Scouts remains a separate visual-persistence seam.
