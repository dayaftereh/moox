# Slice 15.6 Gate 3 - Tactical UI refinement review
Date: 2026-09-12
Status: **CLOSED / accepted Slice-15.6 Tactical baseline; further polish deferred**
User workstream shorthand: **5A Tactical UI**
## Naming / history guard
The user refers to this resumed workstream as `5A Tactical UI`. Existing Slice 15.6 documentation already uses `Block5A` for the completed Galaxy knowledge/exploration authority block, so the repository history is not renumbered or rewritten. This document is the active Tactical UI refinement review within Slice 15.6 Gate 3.
## Entry condition now satisfied
Block5E is complete. The Triangle reference scenario now provides a real browser gameplay path from strategic play into Tactical combat, including the previously blocked larger-Fleet case:
- co-located known foreign Fleets can be inspected;
- `Angreifen` performs authoritative war declaration;
- encounter creation happens on the next `Fertig` strategic-resolution boundary;
- Tactical accepts 3v2+ concrete combat-Ship counts;
- unarmed Scouts are valid Tactical participants;
- strategic Colony/Outpost/Troop vessels do not materialize as Tactical ships;
- explicit `R├╝ckzug` exists for a no-weapon/retreat outcome;
- on 2026-09-12 the user played the Triangle scenario through the real encounter flow and successfully reached the Battle/Tactical UI.
This is the stable entry point required before judging or redesigning the Tactical presentation.
## Existing Tactical baseline to preserve while reviewing UI
Slice 15.5 remains the closed functional authority baseline. Current Tactical presentation already has:
- participant-safe server-projected SVG battlefield;
- authoritative active ship / initiative sequencing;
- server-projected legal movement points;
- server-projected legal fire actions/targets;
- read-only Scan/ship inspection;
- End Activation and Retreat commands;
- server-authoritative damage/readiness/state refresh;
- desktop wheel zoom + drag pan;
- touch one-finger pan + two-finger pinch zoom;
- explicit zoom/recenter controls;
- effectively unbounded camera presentation without a normal arena edge;
- responsive desktop/mobile layout and shared DE/EN localization;
- BattleSession -> strategic return contract.
These are behavior/authority contracts, not a freeze on visual design.
## Review scope
The next pass is deliberately UI/HMI-first and should be evaluated in the real Triangle Tactical encounter. Review and refine, without moving authority into React:
1. overall battlefield composition and visual hierarchy;
2. ship scale, orientation, selection state and active-ship emphasis;
3. friendly/enemy distinction and fleet readability for 3v2+ battles;
4. movement visualization and destination affordances;
5. attack/target visualization, weapon affordances and readiness feedback;
6. initiative/turn-order presentation;
7. ship health, armor/structure and damage feedback;
8. Scan/details presentation and information density;
9. action controls, mode switching, End Activation and Retreat placement;
10. camera controls, default framing, pan/zoom ergonomics and recenter behavior;
11. desktop use of available width/height;
12. mobile/touch layout, overlays/sheets and no-hover operation;
13. battle-entry, active-battle and battle-result/return transitions;
14. animation/polish only where it remains presentation-only and does not affect deterministic authority.
## Acceptance discipline
- Use the real `game-triangle-2pc` strategic path to reach Tactical whenever practical; do not substitute a React-only mock for final acceptance.
- Preserve server authority for movement legality, target legality, range, RNG, damage, readiness, initiative and terminal result.
- Support 3v2+ presentation; do not regress to assumptions that exactly two ships exist per side.
- Unarmed ships and Retreat must remain valid.
- Civilian strategic vessels must remain excluded from Tactical materialization.
- Desktop and mobile/touch are both required acceptance surfaces.
- Block5F scanner/deep-space Fleet intelligence remains queued until this Tactical UI review has a stable direction.
## Closeout decision (2026-09-14)
The user has now play-tested the real Triangle Tactical path through movement, activation controls, independent weapon slots and grouped Laser volleys and confirmed that the Lasers behave as expected. This Tactical review is therefore accepted as the Slice-15.6 playable baseline. It is intentionally **not** the final Tactical UX/mechanics version. Any further visual polish or deeper authority mechanics must not keep Slice 15.6 open; they move to the explicit later Tactical breadth backlog.

## Block 1 - immersive Tactical shell and command HUD
Status: **IMPLEMENTED / browser-QA green / user-accepted Slice-15.6 baseline**

The original MOO2 Tactical screen supplied by the user is used as interaction/layout reference, not as a pixel-for-pixel asset copy. The first refinement block establishes the same overall information hierarchy while preserving MOOX server authority:

- Battle/Tactical uses an immersive full-viewport shell instead of rendering inside the normal strategic application chrome.
- The existing game-menu trigger remains as a floating control at the upper-left so language, save/load and Main Menu remain available.
- Strategic resource chips, strategic turn/phase context, side navigation and the strategic bottom `Fertig` bar are hidden while the Battle route is active.
- Entering Tactical replaces the encounter-card body with the battlefield itself; Tactical is no longer nested inside the `battle-entry-card`.
- Battlefield owns the available viewport and keeps pan, wheel zoom, touch pan/pinch and explicit zoom/recenter controls.
- Movement presentation uses a 1x1 minor square grid (one smallest movement cell) with stronger 4x4 major grid lines. The camera stays effectively unbounded; there is no normal arena-edge presentation.
- Every server-projected legal movement destination is rendered as a green clickable 1x1 cell. React does not calculate reachability.
- Bottom Tactical HUD contains the own-Ship roster, active-Ship summary, Move/Fire/Scan controls, weapon-slot selection while in Fire mode, `Nächstes Schiff`, `Rückzug`, and encounter-overview return.
- Own-Ship roster entries show active/ready/complete state and can focus the camera. The server remains authoritative over which Ship is actually active.
- Scan mode opens a Fleet-`?`-style read-only detail popover after selecting a Ship, including identity, movement/facing, beam values, armor/structure and weapon/readiness information already present in the participant-safe Tactical projection.
- Desktop and compact/mobile HUD breakpoints are presentation-only; no hover-only Tactical action is required.

### Browser evidence
A clean parallel server was started on `127.0.0.1:7172` so the user's live 7171 Triangle 3v2 battle would not be reset. Browser QA replayed the Triangle path through contact, `Angreifen`, strategic `Fertig`, encounter creation and `Taktischen Kampf betreten`.

Inside the new Tactical screen the browser verified:

- `shell-immersive` is active;
- strategic resource bar, strategic context, side navigation and strategic bottom command bar are all hidden;
- Tactical battlefield and bottom HUD fill the viewport;
- both 1x1 minor and 4x4 major grid definitions are present;
- 515 server-projected legal movement cells were rendered for the active Scout in the QA encounter;
- the HUD exposed own Scouts, Move/Fire/Scan, `Nächstes Schiff`, `Rückzug` and encounter return;
- Scan mode on a Darlok Scout opened the detail popover with participant-safe Frigate / Nuclear Drive / position / facing / movement / beam / armor / structure / weapon information.

### Authority gap deliberately not faked
The user's preferred activation workflow is stronger than the current server contract: move/fire choices should be previewed/staged for the active Ship and committed together when `Nächstes Schiff`/finish-activation is chosen. Current MOOX Tactical commands (`battle.move_ship`, `battle.fire_beam`, `battle.end_activation`, `battle.retreat`) are submitted immediately one-by-one to the authoritative Battle endpoint; there is no Tactical preview/batch endpoint today.

Therefore Block 1 keeps honest current semantics: movement and fire still mutate authoritatively immediately, while `Nächstes Schiff` maps to the existing `battle.end_activation`. A later authority block must normalize and implement Tactical activation preview/batching before the UI may claim deferred commit semantics.

Likewise, the requested convenience rule “no weapon selected = fire all eligible weapons at the target” is not implemented cosmetically. Current authority fires one explicit weapon slot per command. Fire-all needs an authoritative command/transaction rule (including ordering, readiness, target validity, sequence/RNG behavior and failure semantics) before React exposes it.

## Next implementation blocks
1. **Tactical activation transaction authority:** research/freeze preview + staged move/fire + commit/end-activation semantics, save/reconnect/conflict behavior and opponent information timing.
2. **Weapon targeting contract:** explicit per-weapon selection plus authoritative `fire all eligible` behavior when no subset is selected.
3. **HUD/field polish after user visual review:** initiative/next-Ship presentation, active-Ship emphasis, target/range feedback, damage feedback, mobile spacing and final control wording.


### Final live 3v2 verification on the user's preserved 7171 state
After the Block1 commit candidate was built, the already-running 7171 reference server was **not restarted**. A cache-busted browser load picked up the new web assets while preserving the user's real Round-6 Triangle encounter with three Human Scouts versus two Darlok Scouts.

Entering Tactical performed no Battle command and confirmed the final UI against that preserved state:

- exactly five Tactical ship markers were rendered: Human `Scout 1`, `Scout 2`, newly built `Scout`, plus Darlok ship IDs 26 and 27;
- the bottom own-Ship roster showed all three Human Scouts simultaneously;
- the Tactical screen occupied the full browser viewport from top 0 to bottom 605 in the QA window with no document overflow;
- resource chips, strategic side navigation and strategic bottom command bar remained hidden;
- Scan on Darlok ship 26 opened the same participant-safe detail popover successfully;
- no move/fire/end-activation/retreat command was sent during this final 3v2 verification, so the user's live Battle state remained unchanged.

This is the preferred Block1 browser evidence because it uses the exact larger-Fleet state that originally blocked access to Tactical.


## Block 2 - implicit move/fire interaction and Galaxy-style camera
Status: **IMPLEMENTED / browser-QA green / user-accepted Slice-15.6 baseline**

User feedback after Block1 was to remove explicit Move/Fire mode selection, reduce the HUD footprint, make the tactical grid materially visible, and make camera interaction match the Galaxy-map mental model.

### Interaction contract now implemented
- Normal Tactical interaction is a single `combat` state rather than separate Move and Fire modes.
- Server-projected legal movement cells are always shown during the player's active Ship activation. Clicking a legal green cell immediately issues the existing authoritative `battle.move_ship` command.
- Server-projected legal enemy targets are simultaneously target-highlighted. Clicking a legal enemy Ship immediately fires the currently selected/available authoritative weapon slot through `battle.fire_beam`; no separate `Feuern` mode/button is required.
- Clicking an own Ship in normal combat interaction focuses the camera only; it does not attempt friendly fire or change server activation authority.
- `Scannen` remains the one explicit temporary interaction mode. While Scan is active, move cells and fire-target affordances are suppressed and Ship clicks open the read-only participant-safe details. Pressing Scan again returns immediately to normal combat interaction.
- Explicit Move and Fire action buttons were removed from the bottom HUD. The weapon-slot selector remains available only when authoritative legal fire actions exist.
- Existing authority remains honest: Tactical commands still submit immediately. Deferred activation batching/preview and no-selection=fire-all are still later server-authority work and are not simulated in React.

### Camera / gesture contract
- Removed visible `+`, `-` and Recenter controls.
- Mouse wheel zoom remains direct.
- Left-button drag pans.
- Right-button drag also pans; the browser context menu is suppressed on the battlefield.
- One-finger touch pans.
- Two-finger touch pinch pans/zooms around the gesture center.
- Click-vs-drag protection prevents a drag beginning on a legal move cell or Ship from becoming an accidental move/fire click when released. Pointer capture is used opportunistically but is not required for correctness.

### Grid and HUD visibility
- Minor 1x1 movement grid lines are now approximately 0.45 CSS px with stronger contrast.
- Major 4x4 grid lines are approximately 1 CSS px and significantly brighter.
- Green legal-cell outlines were increased to approximately 0.45 CSS px.
- At the 791x605 QA viewport the bottom HUD reduced from about 151 px in Block1 to about **94 px**, giving the battlefield roughly 57 additional vertical pixels.
- The separate active-Ship summary panel was removed; active/ready/complete state remains visible in the own-Ship roster.
- The compact tools area now contains only Scan plus weapon-slot chips when weapons are actually legal; `Nächstes Schiff`, `Rückzug` and encounter overview remain in the compact commit group.

### Browser QA evidence
The user's preserved Round-6 7171 Triangle 3v2 state was exported read-only (107,451-byte live snapshot) and restored onto isolated QA server `127.0.0.1:7173` with persistence enabled. The canonical 7171 Battle was not mutated.

The cloned real 3v2 state had Human `Scout 1`, `Scout 2`, newly built armed `Scout`, and Darlok ships 26/27. Browser QA proved:

1. **Implicit fire:** active Human Scout exposed `Laser Cannon` and two legal targets with no Fire mode. Clicking Darlok Ship #26 directly produced authoritative `beam_fired` command sequence 12, hit for 3 damage, and applied Armor 4 -> 1; the weapon became spent.
2. **Implicit movement:** with no Move mode/button, clicking legal green cell `(11,13)` directly produced authoritative `ship_moved` command sequence 13, from `(10,13)` to `(11,13)`, movement 20 -> 19.
3. **Scan toggle:** normal mode showed 514 move cells and two fire targets before the shot; Scan hid both and opened the Darlok detail popover; toggling Scan off restored normal combat affordances.
4. **Drag safety:** a synthetic drag followed immediately by a click did not add a second `ship_moved` event; the latest Battle event remained sequence 16 from the intentional move.
5. **Wheel zoom:** wheel input changed Tactical viewBox from 44x40 to approximately 36.75x33.41.
6. **Left drag:** left-button pointer drag changed the viewBox center.
7. **Right drag:** right-button pointer drag also changed the viewBox center, and a context-menu event was `defaultPrevented=true`.
8. **Touch pinch:** a two-pointer touch gesture changed viewBox from 44x40 to approximately 29.33x26.67.
9. **One-finger pan:** a single touch drag changed viewBox center while keeping the same zoom dimensions.
10. **Final grid CSS:** browser computed minor grid width 0.45 px, major grid width 1 px, green legal-cell outline 0.45 px; explicit camera-control count was zero.

The exact live 7171 user battle remains the acceptance surface for the user's visual review; all destructive move/fire QA was confined to the 7173 clone.


## Block 3 - persisted ship visuals and selective movement cells
Status: **IMPLEMENTED / browser-QA green / user-accepted Slice-15.6 baseline**

This block replaces the remaining Tactical-only ship-art identity and permanent-grid presentation with the same ship identity contract already used by Fleet/System/Ship Designer.

### Authoritative ship visual identity
- `TacticalShipSpec` and `TacticalShipView` now carry strategic source-design identity (`source_design_id`, design revision, strategic picture id) plus optional persisted `source_visual_revision` / `visual_genome`.
- Tactical materialization deep-clones a combat Ship's persisted visual genome from `core.Ship`; Battle validation rejects mismatched visual revision/genome pairs and validates the genome schema.
- Battle-session cloning and participant view projection deep-clone the visual genome again so UI/view mutation cannot alias authoritative state.
- React renders `ProceduralShipGlyph` with the decoded persisted genome when present. It no longer invents a Tactical-specific visual seed.
- Legacy ships with no persisted genome use the **same stable design fallback seed as Fleet/System**: empire + source design id + source design revision + strategic picture id. Thus old starting Scouts remain visually consistent with strategic Fleet presentation even though their old save records predate persisted visual genomes.
- The field ship art lives inside exactly one 1x1 Tactical cell (foreignObject -0.5..+0.5 on each axis with overflow hidden). The existing shared `shipHullFootprint` drives scale: Doom Star 1.0, Titan .9, Battleship .78, Cruiser .68, Destroyer .58, Frigate .48, Scout token .4. Current Slice15.5 combat authority still limits real battles to the accepted Frigate/Scout baseline, so browser proof here is Frigate footprint .48; broader hull combat will inherit the same scaling without Tactical-specific constants.
- The whole occupied 0.96x0.96 cell is an invisible click target, so a small Scout does not require pixel-perfect clicking on the SVG silhouette.

### Selective cell presentation
- The permanent minor/major Tactical grid is removed from the rendered battlefield. Empty space has no visible grid lines.
- Before selecting the active own Ship, no movement/selection/occupancy cells are shown.
- Clicking the active own Ship/its cell toggles movement selection. Its current cell becomes blue.
- Only server-projected `legal_moves` become green; React still does not calculate reachability.
- Every other non-destroyed Tactical Ship coordinate is shown as a red occupied/blocking cell while movement selection is open.
- Occupied coordinates are never simultaneously green because the server's existing `tacticalMoveOption` rejects occupied destinations.
- Occupancy blocks only the destination coordinate; diagonal/free movement around another Ship remains available whenever the server projects that destination as legal.
- Clicking the selected active Ship again closes the movement overlay and returns blue/green/red cell counts to zero.

### Regression and browser evidence
New Go regressions prove that strategic persisted visual identity is cloned into Tactical metadata and then safely projected through the Battle session without pointer/slice aliasing. Source design identity is also projected for exact Fleet/System legacy fallback. Existing movement authority tests already cover occupied-destination rejection and diagonal movement.

Final browser QA used isolated `127.0.0.1:7174`, restored from the user's preserved Round-6 3v2 snapshot. Because that Battle was originally created before Block3, the QA copy was enriched only with visual/source-design fields already present on the same snapshot's authoritative strategic Ship records; canonical 7171 was not modified. This bridge is only for old-snapshot UI QA; new Tactical materialization is covered by the server regressions above.

Observed in browser:
- before selection: 0 green move cells, 0 blue selected cells, 0 red occupied cells, 0 minor/major grid elements;
- active newly built Scout projected `source_visual_revision=1` and persisted seed `shipbuilder:catalog:20:1`; its SVG reported genome version 4 and Frigate footprint .48;
- the SVG container was exactly 1x1 with `overflow:hidden`, while the invisible click cell was .96x.96;
- selecting the active Scout produced one blue cell, 514 server-authoritative green move cells and four red occupied cells;
- red occupied coordinates were (14,12), (10,11), (14,9), (14,11), and **none** appeared in the green move set;
- diagonal legal destinations around blockers remained present, including (13,10), (13,12), (11,12) and (11,14);
- clicking the selected Ship cell again returned all overlay counts to zero.

Legacy starting Scouts in this existing save still have no persisted `visual_genome` because they were created before the persistent visual contract was assigned to them. Tactical now reproduces their Fleet/System design-seeded SVG exactly instead of using a Tactical-only seed. A future new-game/legacy-backfill task may assign persisted per-instance genomes to those historical starting Ships if desired; this is separate from Tactical rendering authority.


## Block 4 - automatic activation feedback and combat motion
Status: **IMPLEMENTED / browser-QA green / user-accepted Slice-15.6 baseline**

Block3 still required a second local click to expose movement overlays even though the Battle server had already selected the active Ship. User review correctly identified this as a broken-feeling activation model. Block4 removes that duplicate UI selection state: server `active_ship_id` is now the only activation authority.

### Activation and movement affordance
- During an own activation and outside Scan mode, the authoritative active Ship cell is immediately blue; no extra click/toggle is required.
- All current server-projected `legal_moves` are immediately shown green.
- Every other non-destroyed occupied Ship cell is immediately shown red. React still derives no reachability or blocking rules.
- Green/blue/red contrast and strokes are strengthened for readability.
- When `active_ship_id` changes, the camera automatically recenters on the new active Ship and keeps at least the Tactical focus zoom. The movement overlay follows the new active Ship immediately.
- Clicking own Ship SVGs remains useful for focus/Scan but no longer changes or hides the movement overlay.

### Ship-first presentation
- Battlefield text labels such as Scout 1/2/3 are removed; the playfield now presents the authoritative SVG silhouette centered inside its 1x1 movement cell.
- The persisted/shared SVG identity from Block3 remains unchanged; active Ship SVG receives a blue visibility glow without changing geometry.
- The bottom own-Ship strip is now icon-only. Ship names remain only in title/ARIA metadata for accessibility/inspection, not as visible Tactical chrome. Active, complete and destroyed states are indicated through border/background/state-dot treatment.
- Top status no longer repeats the Ship name; it reports activation state plus remaining movement.

### Authoritative combat motion
- Tactical now observes newly appended Battle events instead of synthesizing effects from click intent. Historical events are not replayed on first mount.
- New `beam_fired` events animate a visible beam from the authoritative firing Ship coordinate to the authoritative target coordinate. A hit uses a red glow/light core plus impact flash; misses use a dashed/weaker presentation. The effect carries event sequence/hit/damage metadata for QA and clears after ~620 ms.
- New `ship_moved` events animate a short movement trail from authoritative `from_x/from_y` to `to_x/to_y`, plus an arrival pulse, and clear after ~720 ms.
- Both animations are presentation-only consequences of authoritative events; they do not alter command/range/damage/movement rules.

### Browser evidence on isolated 7174
Final QA used the preserved 3v2 clone only; canonical 7171 remained untouched.
- On Tactical entry, with Ship 35 active, the UI immediately showed 1 blue active cell, 514 green legal destinations and 4 red occupied cells **without any extra ship click**.
- Field Ship labels were 0; the three own-Ship roster buttons had no visible text while preserving title/ARIA names. The active persisted Scout stayed Frigate footprint .48 with its existing organic SVG.
- Clicking an authoritative legal enemy target produced a real `beam_fired` event and, 220 ms later, a visible `tactical-beam-animation is-hit` with two beam lines, impact circle, `data-hit=true` and `data-damage=3`; it cleared automatically.
- Clicking an authoritative legal green destination moved the Ship from (10,13) to (11,13), produced a `ship_moved`-driven movement trail/arrival pulse, and recomputed the remaining legal move catalog from 514 to 442 cells.
- Clicking `Nächstes Schiff` completed the last Round-2 activation; the server advanced to Round 3 and changed `active_ship_id` to Ship 23 at (14,12). Without another click, the UI immediately recentered on (14,12), moved the blue active cell there, restored 514 green + 4 red cells, and moved the blue active roster treatment to the new Ship.

## Mobile performance / ship readability follow-up (2026-09-12)
Implemented in the dedicated follow-ups `SLICE_15_6_GATE3_TACTICAL_MOBILE_PERFORMANCE_DAMAGE_FEEDBACK_2026-09-12.md` and `SLICE_15_6_GATE3_STARTING_SCOUT_VISUALS_SPEED_CORRECTION_2026-09-12.md`: one-path reachable grid instead of per-cell rectangles, direct centered native Tactical ship SVGs with tight viewBox, slower authority-derived damage feedback, and persistent v4 visual genomes for fresh starting Scout designs. Original MOO2 drive evidence was rechecked and confirms pristine Frigate Nuclear/Fusion movement 20/22; 10/12 is the engine-damage minimum, so the temporary MinSpeed baseline was reverted.

## Direct selection / hull scale / Finish follow-up

Implemented and browser-QA green in SLICE_15_6_GATE3_TACTICAL_SELECTION_SCALE_FINISH_DAMAGE_DWELL_2026-09-12.md. Own ships are now directly selectable/focusable by their exact unrotated 1x1 cell. The selected cell is filled blue, all other occupied live-ship cells are red outlines, and only server-projected legal movement is green. Circles are removed. Persisted visual hull drives the shared Scout .40 -> Doom Star 1.00 one-cell scale with a tighter geometry-derived viewBox. The bottom strip is limited to unfinished own activations. Fertig submits the existing authoritative battle.end_activation for the current active ship; inactive selection remains inspection/focus only until a server Wait/select authority exists. Damage numbers now dwell for 4s (React feedback lifetime 4.2s).

## Authoritative Wait / Done follow-up (2026-09-13)

Gate-3 testing proved local focus-only selection was insufficient. attle.wait_activation is now authoritative: Warten or clicking another unfinished own ship changes active_ship_id without completing or refreshing the current ship; movement and fired-weapon state are preserved and the player can switch back until explicitly pressing Fertig. Tactical view projects allowed friendly wait targets. Done remains activation_complete and round refresh happens only after every live ship is explicitly complete. See SLICE_15_6_GATE3_TACTICAL_WAIT_DONE_REORDER_2026-09-13.md.
