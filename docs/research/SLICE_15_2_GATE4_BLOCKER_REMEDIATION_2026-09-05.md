# Slice 15.2 Gate 4 blocker remediation

Date: **2026-09-05**

Status: **G4-A/G4-B and iterative G4-B2 A-J complete; G4-B2-K Galaxy framing in progress; G4-B2-L/B2-M/B2-N system inspection refinements complete; G4-C independent QA/closure remains open**.

Gate 3 is technically complete, but direct mobile/desktop review exposed usability blockers that should be corrected before Slice 15.2 is independently closed. Visual identity, final iconography/artwork and richer presentation remain later 15.x work unless they are required for basic strategic usability.

## Accepted remediation split

### G4-A - Galaxy interaction and persistent actions

Status: **complete**.

- Replace button-only Galaxy panning with direct pointer/touch panning.
- Desktop: left-button drag pans; mouse wheel zooms.
- Mobile/touch: one-finger pan; two-finger pinch zoom plus midpoint movement.
- Retain +/- and reset controls as accessible fallbacks.
- Reduce coordinate/text clutter on Galaxy nodes; keep compact star labels.
- Give the map more viewport area and make it the visual focus of the Galaxy screen.
- Keep `Main` and `End Turn` in a persistent, predictable location on mobile and desktop.
- Keep End Turn server-authoritative through the existing planning submission path.

### G4-B - System dialog and Colony usability

Status: **complete**.

- Replace the current below-map system detail with an explicit star-system dialog/overlay opened by star selection.
- Make orbital bodies, fleets, colony/outpost status and legal actions immediately understandable.
- Reduce Colony text density and make current production/build queue prominent.
- Improve Population presentation toward direct Farmer/Worker/Scientist manipulation while preserving server-projected legality and Planning Preview authority.
- Keep richer 2D Colony/building artwork as later 15.x presentation work unless required for usability.


## G4-B validation evidence

- The selected star now opens an explicit `role="dialog"` system overlay rather than rendering system details below the Galaxy map.
- Dialog behavior was browser-smoke-tested for star selection, explicit Close and `Escape` return to the base Galaxy route.
- The dialog presents a compact orbital-body visual plus body type, Colony/Outpost status, legal Colonize/Outpost/Colony actions and Fleet/contact information without inventing mechanics.
- Population assignment no longer depends on source/destination dropdowns. Farmer/Worker/Scientist buckets render one visual marker per Population unit (including fractional Population); selecting a marker exposes direct destination-job actions.
- Browser authority smoke: moving one Farmer to Worker produced a Planning draft and authoritative preview changing Farmer `2.0 -> 1.0`, Worker `1.0 -> 2.0`, Food `4.0 -> 2.0`, Production `3.0 -> 6.0`, and the projected growth state to starvation. No economy/growth rule was copied into React.
- Construction is promoted beside Population at the top of Colony Detail. Browser smoke added `Housing` to the same Planning draft and the queue immediately projected it while retaining the Population change.
- Narrow/mobile Colony management now uses responsive Colony cards instead of requiring the desktop table to scroll horizontally; the desktop table remains available at larger widths.
- Built Buildings use a compact tile grid with neutral structural placeholders only; final building/icon artwork remains later 15.x work.
- `npm run build` in `web/` - **PASS** after the final G4-B changes.
- `go test ./internal/session ./internal/server -count=1` - **PASS** (`internal/session` 47.862s; `internal/server` 1.422s).
- `git diff --check` - **PASS** (Windows LF/CRLF notices only).
### G4-B2 - Interaction-density refinement after direct review

Status: **complete; direct user review pending before independent G4-C closure QA**.

This block captures direct hands-on review after G4-B. These items are treated as pre-closure interaction requirements rather than later artwork polish:

- **Colony overview remains table-first on both desktop and mobile.** Do not replace the strategic colony overview with large vertical cards as the primary representation. On narrow screens, use a compact horizontally scrollable/sticky-column table or similarly dense table treatment.
- **Population assignment should behave more like the original MOO interaction model.** Farmer/Worker/Scientist Population remains visible as individual figures/markers inside each job cell. Support direct drag-and-drop between Farmer, Worker and Scientist buckets on desktop and touch devices. Selecting multiple figures/units before dropping is desirable so, for example, two of three Farmers can be moved in one operation. The final icon artwork is later 15.x, but the interaction model belongs here.
- **Construction Queue becomes a dedicated Colony build-management route/page.** Colony Detail should summarize current build and queue compactly, with an explicit action opening the build page. That page should present available projects in dense side/top selection areas and the ordered active queue clearly separated below, with an obvious return to the Colony Detail. Planning authority remains unchanged.
- **Galaxy should read as a near-fullscreen strategic surface.** Reduce page chrome/headlines/descriptive copy around it; let the map consume most available viewport area while retaining pan/zoom and system selection.
- **Star-system presentation should become spatial/orbital rather than list-first.** Center the star and place planets/orbital bodies around it on visible orbit rings. Body details/actions may remain in supporting panels, but the first visual impression should be a larger orbital system view.
- **Global UI density should increase.** Reduce base font sizes, card/panel padding and redundant headings. Prefer concise labels plus contextual `?`/help affordances over long explanatory paragraphs. This is a product-wide 15.x direction, but Slice 15.2 should already avoid oversized primary controls and headline-heavy strategic screens.
- **End Turn should stop dominating the interface.** Move it toward a smaller persistent bottom-right action on desktop/mobile-safe layouts instead of a wide two-column action bar. Keep it predictable and always available during Planning, but visually secondary to the strategic surface.


#### G4-B2-A - Colony table + Population drag/drop

Status: **complete**.

- Colony overview is table-first again at every viewport width; mobile keeps the same dense table with horizontal scrolling and a sticky Colony column instead of switching to large cards.
- Population remains visible in three Farmer/Worker/Scientist buckets inside the table row and Colony Detail.
- Individual Population markers support multi-selection; selected units can be moved together.
- Desktop HTML drag/drop is supported between job buckets.
- Touch/pen pointer drag/drop is supported by pointer capture plus destination hit-testing, with tap-selection + destination-button fallback retained for accessibility.
- Authority evidence: selecting both starting Farmers and moving them to Worker projected Farmer `2.0 -> 0.0`, Worker `1.0 -> 3.0`, Food `4.0 -> 0.0` and Production `3.0 -> 9.0` through the existing Planning Preview. A subsequent desktop drag Worker -> Scientist and touch-style drag Scientist -> Farmer also updated the same preview path.
- `npm run build` - **PASS**.

#### G4-B2-B - Dedicated Colony build-management page

Status: **complete**.

- Added a typed `build` game subview so a Colony construction workspace has its own route, e.g. `#/game/demo/colonies/10/build`, while keeping Colonies as the active strategic section.
- Colony Detail now shows only a compact current-build/queue summary, authoritative ETA and a `Bau verwalten` action.
- The dedicated build page separates the available build catalog from the ordered queue, with the catalog first and queue below.
- Queue items remain editable with up/down/remove plus full draft reset, all using the existing `colony.set_construction_queue` Planning draft.
- Browser smoke added Housing and Troop Transport, reordered Troop Transport to position 1, then returned to Colony Detail. The compact summary projected `Troop Transport`, `2 in Queue`, `34 Runde(n)` and `Housing` as the next item.
- `npm run build` - **PASS**.

#### G4-B2-C - Spatial Galaxy/System and strategic density

Status: **complete; direct user review pending**.

- Galaxy is now a near-fullscreen strategic surface: the large page heading/explanatory copy were removed from the active map view, while a compact toolbar keeps zoom/reset plus contextual `?` help.
- The Galaxy map continues to support direct mouse/touch pan and zoom from G4-A. A low-height desktop smoke viewport (`791 x 605`) measured the final Galaxy card at top `175`, height `420`, bottom `595`, with `379px` of map height, keeping the strategic surface inside the visible shell rather than forcing vertical overflow.
- System presentation is now spatial/orbital: a central star, one visible orbit ring per body and deterministically positioned clickable planets/orbital bodies. Clicking a body links back to its authoritative detail/action row; Colony/Outpost/Colonize actions remain unchanged.
- Browser geometry smoke on Alpha confirmed one central star, one ring and one body (`Alpha I`) spatially separated on a `493 x 360` orbital stage.
- Strategic shell typography, cards, list rows, buttons, page headers and resource chips were compacted to increase information density without changing authority or data semantics.
- Persistent actions are now a small bottom-right cluster instead of a broad bar. In the same desktop smoke viewport the cluster measured about `154 x 40px`; on narrow/mobile layouts the redundant `Main` action is hidden and End Turn remains available without dominating the screen.
- Regression smoke after density changes confirmed the table-first Colony overview remains intact and the dedicated `/colonies/10/build` construction workspace still renders its build catalog and queue correctly.
- `npm run build` - **PASS** after final viewport-height correction.
- `git diff --check` - **PASS** (Windows LF/CRLF notices only).

#### G4-B2-D - Compact shell and full-viewport strategic layout

Status: **complete; direct user review pending**.

- Removed the separate resource strip and moved strategic resources directly into the compact top bar.
- The visible top bar now prioritizes Main/OX plus BC, Food, Freighters/Transporter and Command Points. Connection state remains available as a tiny status dot/tooltip rather than visible status text; DE/EN was removed from the top bar and remains in the left desktop Main/navigation area.
- Food in the top bar is display-only aggregation of the authoritative projected Colony `adjusted_economy.food` values. Browser authority smoke moved one Farmer to Worker and the top Food display changed `4.0 -> 2.0` together with the Planning Preview (`Food 4.0 -> 2.0`, Production `3.0 -> 6.0`). No food rule was duplicated in React.
- Mobile bottom navigation now shows all seven strategic sections in one row: Galaxy, Colonies, Fleets, Research, Diplomacy, Espionage and More.
- End Turn is integrated at the bottom-right edge of the command bar on mobile and remains a small persistent bottom-right action on desktop.
- The former 1180px content cap was removed for strategic game screens. A 1200px desktop probe measured the retained left navigation at `158px`, Colony content/table at `1030px`, and End Turn at about `98 x 34px` in the bottom-right corner.
- A 360px probe measured the compact top bar at `46px`, bottom command bar at `48px`, all seven tabs at about `39px` each and End Turn at about `70 x 44px`.
- A 320px probe confirmed icon-only Main/OX, no visible round/phase chip, all BC/Food/Tr/CP resources fitting without horizontal overflow, all seven tabs at about `33px` each, and a full-width Colony content area (`308px`) with horizontally scrollable table (`306px` viewport) and sticky Colony column.
- Galaxy now computes its height from the compact shell bars so the strategic surface uses essentially the full remaining viewport.
- `npm run build` - **PASS**.
- `git diff --check` - **PASS** (Windows LF/CRLF notices only).

#### G4-B2-E - Three-dot game menu and five-tab primary navigation

Status: **complete; direct user review pending**.

- Replaced the top-left OX/home shortcut with an explicit `⋯` game-menu trigger. Clicking it no longer leaves the game.
- The popover contains compact Settings/Language controls, Espionage, an Advanced entry for the existing session/diagnostic/development surface, and a clear Main-menu exit action.
- The MOOX/OX brand remains available on standalone/main-menu surfaces instead of consuming the in-game menu affordance.
- Removed `More` and `Espionage` from the primary navigation rail and mobile bottom bar. The five primary strategic destinations are now Galaxy, Colonies, Fleets, Research and Diplomacy; Espionage is reached from the `⋯` menu and remains intentionally non-actionable until Slice 20.
- Existing `More` functionality was not deleted: it is retained behind the `Advanced` menu entry so session/seat/diagnostic development utilities remain reachable without occupying primary gameplay navigation.
- The menu closes when navigating from it and supports outside-pointer/Escape dismissal via the shell event handling.
- Browser smoke switched DE -> EN directly inside the popover and immediately updated both menu and strategic-shell localization.
- Browser smoke opened Espionage from the popover and confirmed the menu closed while routing to `/espionage`.
- A 320px probe measured five bottom tabs at about `48px` each plus End Turn, while the popover measured about `294 x 321px` and remained fully inside a `320 x 640` viewport. The top resource row still fit without horizontal overflow.
- `npm run build` - **PASS** after the final removal of visible technical game-ID text.
- `git diff --check` - **PASS** (Windows LF/CRLF notices only).

#### G4-B2-F - Interactive resource HUD and detail popovers

Status: **complete; direct user review pending**.

- Expanded the compact top HUD to five clickable resources: BC, Food, Freighters, Command Points and Research. Current glyphs are structural placeholders for later final icon artwork.
- BC now shows treasury balance plus signed projected net income in square brackets, e.g. `0 [+6]`; positive delta is green and negative delta is red.
- Food now shows the authoritative empire-wide surplus/deficit derived from projected Colony `population_dynamics.food_surplus` / `food_shortage`, e.g. `[+x]` or `[-x]`, instead of raw food production.
- Browser authority smoke moved one Farmer to Worker: the HUD changed Food `0.0 -> -2.0` in red while the same Planning Preview changed Colony Food `4.0 -> 2.0` and Production `3.0 -> 6.0`.
- Command Points show total capacity plus signed available/over-capacity value in brackets, e.g. `5 [+5]`; the detail popover exposes capacity, used, available/overage and command maintenance.
- Research now shows authoritative RP/turn plus progress percent and ETA when a field is selected. Browser smoke selected tech field `29` (50 RP): HUD changed from `4.5 RP [—]` to `4.5 RP [0% · 12T]`, while the Research view independently showed `0 / 50` RP and ETA `12` turns.
- Research becomes yellow only for an active project whose authoritative ETA is <=1 turn (or remaining RP is within the current authoritative RP/turn). No-active-research state stays neutral `[—]`; a false yellow breakthrough state found during smoke testing was fixed.
- Every HUD resource is a button opening a compact detail popover. BC exposes balance/net and projected tax plus maintenance components; Food exposes produced/required/surplus/shortage/import/export; Freighters expose total/available/food/reserved logistics; Command exposes capacity/usage/overage; Research exposes rate/progress/cost/remaining/ETA/field/mode/breakthrough state.
- A 320px browser probe confirmed all five HUD resources fit without horizontal overflow (`clientWidth == scrollWidth == 242`) after ultra-narrow spacing, while the research detail popover remains fully inside the viewport.
- `npm run build` - **PASS** after final projection-consistency cleanup.
- `git diff --check` - **PASS** (Windows LF/CRLF notices only).

#### G4-B2-G - HUD-driven Research selection overlay

Status: **complete; direct user review pending**.

- Removed Research from the visible primary bottom/side navigation. The four primary gameplay tabs are now Galaxy, Colonies, Fleets and Diplomacy; the internal `/research` route remains available as a fallback/debug path.
- Clicking the Research resource in the top HUD now opens a large modal research-selection screen rather than the small resource-detail popover.
- The modal follows the original MOO2 interaction pattern without copying artwork: two-column desktop research grid, one panel per authoritative category, category header + RP cost, dark field body with green technology/application entries, active-field highlight, and Cancel/close behavior.
- No research-field names are fabricated. The current server projection exposes category and technology/application names but no independent display name for a tech field, so the overlay labels these honestly as `Technologiefeld #<id>` until richer metadata exists.
- `choose_one` research exposes each authoritative technology/application as its own selectable action. Other selection modes show their projected technology list plus one field-level Research action.
- Selecting any research option uses the existing `empire.select_research` draft and immediately closes the modal and routes back to Galaxy.
- Browser smoke from Colony opened the Research HUD overlay, selected Chemistry, and verified the route returned to Galaxy while the Research planning draft remained active.
- Browser smoke selecting Construction field #29 updated the top HUD from `4.5 RP [—]` to `4.5 RP [0% · 12T]` after closing back to Galaxy.
- Desktop smoke at `791 x 605` measured an 8-panel, two-column (`369px + 369px`) overlay inside `775 x 589px`; first panel displayed `Konstruktion`, `50 RP`, `Technologiefeld #29`, and its real technology list.
- A 320px probe confirmed the overlay fits inside the viewport (`312 x 632px`) with a vertically scrollable one-column grid and all eight panels, while the bottom navigation expands to four tabs at ~59px each.
- `npm run build` - **PASS** after final authoritative field-label adjustment.

#### G4-B2-H - Explicit research-technology selection state

Status: **complete; direct user review pending**.

- Clarified the difference between field-level and technology-level research selection directly in the overlay.
- The ruleset already exposes four authoritative selection modes: `all`, `choose_one`, `fixed_one`, and `repeat_field`. The UI now labels the mode in each panel instead of making non-clickable entries look broken.
- `choose_one` panels render every technology/application as an explicit button. The selected technology is derived from authoritative projected/current `research.technology_ids` and is shown bold, bright green, outlined, `aria-pressed=true`, and labeled `Aktuell`.
- `all` panels explicitly say that all technologies are researched; `fixed_one` says the technology is fixed; repeat-field mode is labeled separately. No client-side legality is invented.
- Current demo research choices explain the user's observation: five currently exposed fields resolve as `all`, while Sociology, Biology, and Force Fields resolve as `choose_one`. The server rule is: general research fields are always `all`; Creative races are also `all`; Uncreative races are `fixed_one`; otherwise non-general fields use `choose_one`.
- Browser smoke selected Biology / Biospheres, returned to Galaxy, reopened Research, and verified the active Biology field plus `Biospheres · AKTUELL` with bold/bright selected styling. Switching to Hydroponic Farm and reopening changed the selected marker to `Hydroponic Farm · AKTUELL`.
- `npm run build` - **PASS**.
- `git diff --check` - **PASS** (Windows LF/CRLF notices only).

#### G4-B2-I - Robust selected-technology visibility

Status: **complete; direct user review pending**.

- Fixed a timing hole in the Research overlay: selected technology highlighting previously depended primarily on the asynchronous Planning Preview, so an immediate reopen after choosing a technology could briefly show no bold/selected state.
- The overlay now prefers the current `research` DraftOrder (`tech_field_id` / `technology_id`) before falling back to Planning Preview/current empire research state. This keeps the displayed choice synchronized immediately with the user's pending authoritative command intent without duplicating research legality.
- The selected choose-one technology name now renders semantically as `<strong>` with an explicit `font-weight: 1000`, brighter text, a 2px green outline, stronger green background, and a `✓ AKTUELL` badge.
- Immediate browser smoke: select Biology / Biospheres -> overlay closes -> reopen Research immediately with no wait -> `BIOSPHERES · ✓ AKTUELL` rendered as `STRONG`, computed weight `1000`, `aria-pressed=true`.
- Immediate switch smoke: choose Hydroponic Farm -> reopen immediately -> marker moved to `HYDROPONIC FARM · ✓ AKTUELL`, computed weight `1000`.
- `npm run build` - **PASS** after restoring the internal fallback Research view signature.

#### G4-B2-J - HUD spacing and connection-state menu frame

Status: **complete; direct user review pending**.

- Increased visual separation between BC, Food, Freighters, Command Points and Research without increasing the overall top-bar height.
- Removed the separate connection-status dot from the far right of the top bar. The three-dot game-menu control now carries connection status directly as a 2px frame: green for `success`, red for `danger`, yellow for `warning`, neutral gray otherwise.
- The menu trigger keeps the connection-status text in its title/accessible label while the visible UI only uses the colored frame.
- Desktop browser smoke at `791px` measured a consistent `10px` visible gap between all five HUD resources with no resource overflow.
- A 320px probe measured `11px` visible gaps between all five HUD resources while `clientWidth == scrollWidth == 264`, so the more relaxed layout still fits without horizontal scrolling.
- The 320px and desktop probes confirmed the separate `.topbar-live`/`.connection-dot` indicator is gone.
- Computed-style smoke confirmed `success` renders a `2px` `rgb(98, 215, 124)` outline and forced `danger` renders a `2px` `rgb(240, 109, 123)` outline.
- `npm run build` - **PASS** after correcting the menu accessibility expression syntax.

#### G4-B2-K - Galaxy pan bounds and framing review

Status: **in progress; pan bounds complete, MOO2 reference review pending**.

- Added hard pan limits to the Galaxy map for both one-pointer/mouse drag and the pan portion of two-pointer touch gestures.
- The limit is based on the transformed full Galaxy layer rather than individual stars. At 100% zoom only a small edge offset is allowed; zooming in expands the legal pan range by exactly the extra scaled map area; zooming back out automatically clamps the current pan again.
- The small overscroll allowance is responsive: `6%` of the smaller map dimension, bounded to `18..42px`.
- Desktop smoke (`781 x 460` map) forced multi-thousand-pixel drags and clamped to `±27.6px` at 100%. At 208% zoom the positive bound was `x=449.34px`, `y=276px`, matching the scaled-layer calculation; returning to 100% automatically reclamped to `27.6px`.
- A 320px mobile probe (`310 x 495` map) forced a 3000px touch drag and clamped to `18.6px` at 100%, so the map cannot be dragged into an empty no-system viewport.
- `npm run build` - **PASS**.
- Removed all visible Galaxy-map chrome at the user's request: empire/title toolbar, help button, zoom `-`, zoom percentage, zoom `+`, and reset control are gone. The Galaxy card is now map-only.
- Direct gesture zoom remains: desktop mouse wheel and touch pinch still change the authoritative UI zoom state, while the bounded-pan behavior remains active.
- Desktop smoke measured the `501px` Galaxy card with the map occupying `499px`; no toolbar, zoom buttons, or zoom badge remained. Wheel smoke changed the layer from `scale(1)` to `scale(1.12)`.
- A 320px probe measured the `536px` card with `534px` of actual map and no toolbar, preserving the four bottom strategic tabs.
- Restored Espionage to the visible primary navigation now that Research no longer consumes a tab. The bottom/side primary set is Galaxy, Colonies, Fleets, Diplomacy and Espionage; the duplicate Espionage entry was removed from the three-dot game menu.
- Shortened the turn-submit label from Runde beenden / End turn to Fertig / Done to preserve horizontal space. A 320px probe measured five primary tabs at about 48px each plus a 66px Fertig button with no navigation overflow.- Next: review the user's supplied Master of Orion II Galaxy screenshot and refine Galaxy framing/density without copying original artwork.

#### G4-B2-L - MOO2-inspired interactive star-system inspection

Status: **complete; direct user review pending**.

- Rebuilt the System dialog using the user-supplied Master of Orion II system-view screenshot as an interaction/layout reference without copying original artwork.
- The dialog now centers on a large dark orbital scene with an authoritative star, elliptical orbit tracks and clickable orbital bodies. The selected body receives a visible targeting reticle and stronger label treatment.
- Clicking any body now updates a dedicated inspector rather than scrolling to a separate row. For authoritative planets the inspector shows body type, orbit, climate, size, mineral class, gravity and settlement status. Uncolonized planets are fully inspectable.
- Browser smoke on Beta verified `Beta I -> Planet / Orbit 1 / Desert / Small / Rich / Normal G / Unkolonisiert`, then selecting Beta II immediately changed the inspector to `Planet / Orbit 2 / Barren / Large / Poor / Heavy G / Unkolonisiert`.
- Gas giants and asteroid belts are rendered from authoritative `OrbitalBody.kind` with distinct structural visuals. They do not receive fabricated planet-only climate/size/mineral/gravity values; an unoccupied non-planet body reports neutral `Keine Siedlung` rather than implying colonizability.
- Legal actions remain server-projected: Open Colony, Colonize and Build Outpost only appear from the existing authoritative decision choices for the currently selected body.
- The prior fleet/contact information remains available as compact footer traffic badges when present. Close moved into the lower-right footer, matching the supplied reference's hierarchy more closely.
- Current core state exposes no authoritative crystal/Orion/artifact/special-resource field on Planet or OrbitalBody. The UI therefore does **not** invent such markers. The lower-left special-feature concept is reserved for a later core/data extension when actual feature IDs exist.
- Desktop smoke at 791x605 measured a 777x463 orbital stage filling the full dialog scene. A 320x640 probe kept the dialog inside the viewport with a 407px orbital scene plus a compact 132px selected-body inspector.
- `npm run build` - **PASS** after final layout/status cleanup.

#### G4-B2-M - Circular system orbits + authoritative fleet/ship inspection

Status: **complete; direct user review pending**.

- Replaced the perspective/stadium orbit geometry with a dedicated square orbital canvas. Planet positions now use the same X/Y radius and each orbit ring is rendered with equal width/height, so the visual tracks are true circles rather than ellipses.
- Desktop browser smoke on Beta measured a `430 x 430px` orbit canvas and circular rings (`235.1 x 235.1px`, `323.9 x 323.9px`).
- The permanent Slice 15.2 product direction already required the system dialog to show player fleets currently present plus the ships belonging to each visible player fleet. That requirement is now represented directly in the orbital scene rather than only as footer badges.
- Added a clickable `Systemorbit` fleet dock. Every own authoritative `StrategicFleet` whose `at_system_id` matches the open system becomes a fleet marker. Selecting the marker switches the inspector from the planet/body view to fleet details.
- Fleet inspection shows authoritative role, special kind where present, ship count and an explicit location statement: `Im Sternsystem · kein Planeten-Orbitanker`.
- If projected ship identities are available, the fleet inspector resolves `ship_ids` against `decision.strategic.ships`, renders each ship as an individual selectable row and shows its name, hull and source design/revision. No ship identity is fabricated when the projection omits it.
- Fleet markers are deliberately **not attached to a specific planet**. The current authoritative `core.StrategicFleet` contains `AtSystemID` / `DestinationSystemID` but no `AtBodyID` / planet-orbit location. Rendering a fleet beside a particular planet today would falsely imply a location the server does not know.
- If planet-local orbital positioning becomes gameplay-significant, add an explicit authoritative body anchor (for example `AtBodyID`) plus transition/stacking semantics before the HMI places fleets at individual planets.
- For browser QA only, the demo server was temporarily augmented with one valid authoritative combat fleet containing two valid ships (`Falcon`, `Raven`). The temporary backend change was inspected and fully restored before staging/commit. Smoke verified: fleet marker -> fleet inspector -> Falcon/Raven ship buttons -> selecting Raven updates the ship detail panel.
- 320px smoke with that temporary authoritative QA fleet kept the complete dialog inside the viewport, used a `262 x 262px` circular canvas, a compact `142px` fleet dock, a `188px` fleet inspector, and no horizontal document overflow. The displayed ring remained circular (`170 x 170px`).
- `npm run build` - **PASS**.

#### G4-B2-N - System Fleet/Ships dialog and technical ship drill-down

Status: **complete; direct user review pending**.

- Replaced the B2-M in-scene `Systemorbit` fleet dock with the clarified interaction: the orbital picture remains visually clean and a contextual `Flotten / Schiffe` button appears in the System dialog footer whenever at least one own authoritative fleet/special vessel is present in the system.
- The button count represents currently present ship-like units: concrete combat ships from `ship_ids`, plus one unit for each special civilian strategic vessel. It therefore works for ordinary fleets as well as Colony Ship, Outpost Ship and Troop Transport fleet objects.
- Clicking `Flotten / Schiffe` opens a nested system fleet dialog. The roster groups concrete ships under their fleet and lists special civilian vessels as their own selectable entries.
- Clicking a concrete projected `Ship` opens an authoritative technical detail panel with hull, warp drive, FTL speed, computer, armor, shield, fuel cell, fuel range, production cost, source design/revision and all projected weapon mounts/counts/slots.
- Special strategic vessels currently exist as `StrategicFleet` objects rather than concrete `Ship` objects. They remain visible/selectable but only expose the metadata the core actually owns (special kind, role, system location and FTL speed); the UI explicitly states that no component loadout exists for them yet rather than fabricating one.
- Strategic partial damage is a backend gap, not a presentation omission: `core.Ship` contains design/spec data but no persistent armor/structure damage fields. Tactical battle state does track `ArmorCurrent` and `StructureDamage` during an active battle, while strategic encounter resolution currently persists destroyed ship IDs rather than surviving partial damage. The ship panel therefore explains that persistent damage is not yet available instead of inventing a health percentage.
- Browser QA used a **temporary, restored-before-commit** authoritative demo augmentation containing one combat fleet with two concrete Frigates (`Falcon`, `Raven`, each with `2x laser_cannon`) plus one Colony Ship, Outpost Ship and Troop Transport in Alpha. The System dialog correctly showed `Flotten / Schiffe 5`; the nested dialog showed all four fleet/special-vessel groups; Raven selection exposed the full technical loadout and `2x Laser Cannon`; Colony Ship selection exposed only its real strategic metadata.
- 320x640 QA kept the outer system dialog at `314x632` and nested fleet dialog at `306x624`, with a `304x217` roster and `304x353` technical panel; no horizontal document overflow occurred.
- Temporary QA fleets/ships were removed completely from `cmd/moox-server/main.go` before staging.
- `npm run build` - **PASS**.
G4-C must not close Slice 15.2 until this B2 block has either been implemented and reviewed or explicitly re-scoped by product decision.
### G4-C - Independent QA and closure

Status: **pending**.

- Run the canonical strategic workflow on desktop and mobile form factors.
- Verify no browser interaction bypasses server legality/authority.
- Repeat full Go tests/vet/web build/diff checks.
- Synchronize status/HISTORY, remove the OPEN marker and close Slice 15.2 only if the workflow is acceptably usable.

## G4-A validation evidence

- `npm run build` in `web/` - **PASS** after direct pan/zoom and persistent-action changes.
- `git diff --check` - **PASS** (Windows LF/CRLF notices only).
- Browser smoke against the demo fixture:
  - persistent `Main` and `Runde beenden` controls are present;
  - synthetic wheel input changes Galaxy scale;
  - pointer-drag changes map translation;
  - simulated two-pointer pinch changes scale and midpoint translation;
  - star selection still opens the existing system-detail path, confirming the G4-B dialog conversion can remain a separate block.

No push is part of this remediation unless explicitly requested.
