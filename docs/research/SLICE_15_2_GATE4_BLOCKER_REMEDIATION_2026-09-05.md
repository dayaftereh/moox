# Slice 15.2 Gate 4 blocker remediation

Date: **2026-09-05**

Status: **G4-A/G4-B and iterative G4-B2 A-J complete; G4-B2-K Galaxy framing in progress; G4-B2-L/B2-M/B2-N/B2-O system inspection refinements complete; G4-B2-P construction workspace refinement complete; G4-B2-Q colony/planet economy presentation complete; G4-B2-R Colony density cleanup complete; G4-B2-S Planet/Colony profile split complete; G4-B2-T signed growth/resource cleanup complete; G4-B2-U build-management density cleanup complete; G4-B2-V Colony live breakdown/progress complete; G4-C independent QA/closure remains open**.

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

#### G4-B2-O - Direct colony navigation from colonized planets

Status: **complete; direct user review pending**.

- Clicking an authoritative colonized planet in the System dialog now navigates directly to that colony detail view instead of only selecting the planet in the local body inspector.
- The former redundant `Open Colony` footer action was removed; the planet itself is the primary interaction target and its accessible/title text announces the colony action.
- Uncolonized planets, gas giants and asteroid belts keep the existing in-dialog inspection behavior.
- Browser smoke on Alpha verified clicking colonized `Alpha I` closed the System dialog and navigated to `#/game/demo/colonies/10`, rendering `Kolonie #10 / Alpha · Alpha I`.
- Regression smoke on uncolonized Beta verified clicking `Beta II` remained at `#/game/demo/galaxy/4`, kept the System dialog open and updated the inspector to Beta II (`Barren / Large / Poor / Heavy G / Unkolonisiert`).
- `npm run build` - **PASS**.

#### G4-B2-P - MOO2-inspired construction workspace + confirmed build cancellation

Status: **complete; direct user review pending**.

- Reworked the dedicated Colony build workspace using the supplied Master of Orion II build-management screenshot as an information-architecture reference without copying original artwork.
- Desktop at 1200px now uses three functional columns: authoritative build catalog, selected-project detail, and current build / ordered queue. A 791px-class viewport collapses to two columns with the queue beneath; 320px stacks Catalog -> Project -> Queue vertically.
- Catalog clicks now select/inspect first rather than immediately enqueueing. The center panel shows only authoritative choice metadata currently exposed by the server: project kind/name, PP cost, maintenance where present, freighters added where present, Ship Design name/revision where present, and authoritative remaining PP/ETA once the project is in the projected queue.
- Adding a selected choice still emits only the existing `colony.set_construction_queue` Planning draft. Non-repeatable building/transformation choices remain disabled when already queued.
- The current build has a dedicated progress block, authoritative ETA, and `Bau abbrechen`. Removing the queue head requires an explicit in-app `role=alertdialog` confirmation; queued future items remain removable without the destructive-current-build confirmation.
- Core audit confirmed cancellation/reordering semantics already preserve accumulated production: `setConstructionQueue` folds `Construction.ProgressPP + ConstructionReservePP` into `stockPP`, transfers it to the new head, or keeps it as `ConstructionReservePP` if the queue becomes empty. Existing `TestSetConstructionQueuePreservesAccumulatedProductionStock` passes. The confirmation text therefore truthfully states that accumulated PP are preserved.
- Browser smoke: selected Troop Transport -> `Einplanen` -> current block showed `0.0 / 100 PP`, authoritative `34 Runde(n)` ETA; opening Bauabbruch produced the confirmation text and cancelling left the queue untouched; reopening and confirming removed the current project and left a zero-item planned queue.
- Responsive smoke: 1200x720 measured three panels at about `275 / 419 / 316px` with no horizontal document overflow. 320x640 stacked the three 308px-wide panels; after planning Troop Transport, the confirmation dialog measured `296x182` and stayed entirely inside the viewport, with body width 320 and no horizontal overflow.
- The center project area reserves a visibly disabled `Schiff entwerfen` affordance for military ship/design choices. No fake route is exposed: the interactive Ship Designer remains deferred to Slice 17.
- `npm run build` - **PASS**.
- `go test ./internal/game -run TestSetConstructionQueuePreservesAccumulatedProductionStock -count=1` - **PASS**.

#### G4-B2-Q - MOO2-inspired Colony command view + authoritative planet potential

Status: **complete; direct user review pending**.

- Added an additive player-safe `PlanetPotential` projection to `PlayerDecisionView.strategic`. It is derived server-side from the same `EconomyRules` used by Colony economy/capacity resolution; React performs display formatting only.
- `PlanetPotential` exposes per planet: player-relative base Food/Farmer, Production/Worker and Research/Scientist, gravity penalty, ruleset climate habitability percentage, ruleset size base capacity, and Empire-relative population capacity. Government/morale and Colony-local buildings are deliberately excluded from the base job potential.
- Added game-level coverage for the Small Fixture Alpha I baseline (`2 Food/Farmer`, `3 PP/Worker`, `3 RP/Scientist`, `0%` gravity penalty, `80%` Terran habitability, Medium base capacity `15`, player capacity `12`) and DecisionView coverage requiring one valid potential for every projected planet.
- Uncolonized System planets now show a compact `Basis für dein Volk` block beneath their authoritative body facts. Browser smoke: Beta I projected `1.0 F / 5.0 PP / 3.0 RP`, `25%` habitability, `0%` gravity penalty and max population `3`; Beta II projected `0.0 F / 2.0 PP / 3.0 RP`, `25%` habitability, `-50%` Heavy-G penalty and max population `5`.
- Planet class tokens are now localized for climate, size, mineral class and gravity rather than exposing raw `terran / medium / abundant / normal_g` IDs.
- Reworked Colony Detail into a compact classic hierarchy without copying original artwork: left Planet Profile, center Population & Output, right current Construction summary, with the existing Building/Colony surface below.
- Planet Profile explains the current planet using authoritative potential/rules data. Demo Alpha I shows `Terran -> 2.0 Food/Farmer + 80% habitability`, `Medium -> base capacity 15`, `Abundant -> 3.0 PP/Worker`, `Normal-G -> no gravity penalty`, current `4 / 12` population and `6.0 BC` Colony tax contribution.
- The three existing Farmer/Worker/Scientist drag/drop bands now surface their actual authoritative adjusted totals directly beside the population tokens while also showing server-projected base output per Pop. Demo baseline displays Farmer `2.0 -> 4.0 F`, Worker `1.0 -> 3.0 PP`, Scientist `1.0 -> 4.5 RP`, with base Scientist potential `3.0 RP/Pop`.
- Authority regression: selecting one Farmer and moving it to Worker through the existing planning flow changed Farmer `2 -> 1`, Worker `1 -> 2`, Food `4 -> 2` and Production `3 -> 6`; the top Food indicator followed the authoritative Planning Preview to `-2.0`.
- `adjusted_economy.tax_bc` is surfaced as the Colony BC contribution; no new treasury/economy math was added to the client.
- Current Construction summary remains the right-hand Colony action and opens the B2-P manager. Buildings remain structural tiles; rich colony landscape/building art stays deferred to Slice 15.3.
- Responsive QA: 1200x720 produced three Colony command columns at about `242 / 480 / 273px`, equal `414px` height and no horizontal overflow. A 320x640 probe stacked all three command cards at `308px` width with no horizontal overflow and retained touch/tap population controls.
- Mobile System QA after adding potential data: 320x640 dialog `314x632`, circular orbital stage `312x329`, selected-body inspector `312x210`, potential block `312x91`, no horizontal overflow.
- `go test ./internal/game ./internal/session -count=1` - **PASS**.
- `npm run build` - **PASS**.

#### G4-B2-R - Colony density cleanup

Status: **complete; direct user review pending**.

- Replaced the large generic PageHeader (`Imperiumsverwaltung` + large Colony title/subtitle) with a compact Colony toolbar containing only `Kolonie #ID`, System/Planet context and a small Back action.
- Removed redundant population/capacity presentation from the Planet Profile heading and removed the separately repeated max-population row; current population/capacity is now shown once in the compact stats table.
- Removed redundant free-capacity display from Colony Detail because it is directly inferable from the single population/capacity value.
- `Kolonie-BC` is now shown only once in the Planet Profile stats; the duplicate BC badge in the population/output card was removed.
- Farmer/Worker/Scientist rows retain only the current authoritative Colony output totals (`F`, `PP`, `RP`). The repeated per-Pop base-output subtitles and separate base-research footer were removed; planet rule explanations remain once in the Planet Profile.
- Planet trait explanations were shortened (`2.0 F/Farmer · 80%`, `15 Pop Basis`, `3.0 PP/Arbeiter`, `0% Malus`) and converted to a dense table-like row layout rather than large stacked cards.
- Reduced card minimum heights, row padding, marker size and typography. The population drag/drop hint is hidden on Colony Detail while existing selection/tap/drag mechanics remain unchanged.
- The uncolonized-planet `Basis für dein Volk` block was also compressed from ~91px to ~62px at 320px width using a three-column value grid, with all six authoritative values retained.
- Responsive browser QA: 1200x720 Colony command row is now ~300px high (previous B2-Q ~414px), with the compact header ~40px and no horizontal overflow. At 320x640: Profile ~193px, Jobs ~212px, Build ~147px, all 308px wide and no horizontal overflow.
- Authority regression: moving one Farmer to Worker still changed Farmer `2 -> 1`, Worker `1 -> 2`, Food `4 -> 2`, Production `3 -> 6`; top Food followed Planning Preview to `-2.0`.
- `npm run build` - **PASS**.

#### G4-B2-S - Planet Profile / Colony Profile split

Status: **complete; direct user review pending**.

- Replaced the always-visible four-row planet trait block in Colony Detail with a compact `Planeten-Profil` line showing the planet name plus a tap/click `Planeten-Info` disclosure.
- `Planeten-Info` opens a small responsive 2x2 popover containing authoritative climate, size, mineral class and gravity plus their already-projected player-relative effects. A click/tap disclosure was chosen instead of hover-only tooltip behavior so the same interaction works on mobile.
- Added a distinct `Kolonie-Profil` directly underneath. Only Colony-state values remain permanently visible there: population/capacity, growth/starvation, next-Pop ETA, ground forces and Colony BC.
- The German labels now explicitly read `Planeten-Profil` and `Kolonie-Profil`; population uses `Bevölkerungsstatus` rather than the assignment-oriented label.
- No game rules or PlanetPotential math changed; this is presentation-only reuse of B2-Q authoritative data.
- Desktop smoke: popover is bounded inside the viewport (`240x105px`, no horizontal overflow); Colony profile occupies ~88px beneath the 34px Planet Profile line.
- 320x640 smoke: profile card reduced to ~127px total, Colony Profile ~79px, jobs ~212px, build ~147px. The `240x95px` Planet Info popover fits fully at x=65..305 with no horizontal overflow.
- `npm run build` - **PASS**.

#### G4-B2-T - Signed growth/resource presentation + Morale boundary audit

Status: **complete; direct user review pending**.

- Merged Colony `Wachstum` and `Nächste Population` into one compact row. Positive growth now renders as e.g. `+0.07 (14 Runde(n))`; the separate next-pop row is removed. When the colony is starving, the same row remains `Wachstum` but shows the authoritative negative rate without a misleading next-positive-pop ETA.
- Reused the existing signed resource tone convention everywhere touched by this refinement: positive values are green, negative values red, zero neutral. Colony growth now consumes the same `resource-text-positive / resource-text-danger / resource-text-neutral` presentation used by top-bar resource deltas.
- Browser smoke verified the positive baseline: `Wachstum +0.07 (14 Runde(n))`, BC `[+6]` and CP `[+5]` all compute to the positive green tone.
- Authority/negative smoke moved one Farmer to Worker through the existing planning flow. Food became `[-2.0]` with `resource-danger` red, and Colony growth became `-0.10` with `resource-text-danger` red; Farmer/Worker outputs still changed authoritatively from `2/1 -> 1/2`, Food `4 -> 2`, Production `3 -> 6`.
- Morale audit: the core already contains a Phase-1 local morale context (`ECONOMY_MORALE_2026-08-27.md`) including Feudal/Dictatorship missing-barracks penalties, Marine/Armor Barracks cancellation, Holo Simulator +20%, Pleasure Dome +30%, Unification morale immunity, and applied morale effects on colony output/money. The current player-safe browser projection does **not** expose `MoralePercent`, so Slice 15.2 continues to avoid fabricating a Morale UI row.
- Deferred morale breadth remains in the post-Slice-17 economy-fidelity backlog: Virtual Reality Network empire-wide morale, Telepathic Training, Capitol-loss morale, and conquest/assimilation morale. When a player-safe morale breakdown is projected, Colony Profile is the intended display location, using the same signed green/red convention and a detail disclosure for sources.
- `npm run build` - **PASS**.

#### G4-B2-U - Build-management density cleanup

Status: **complete; direct user review pending**.

- Replaced the large Build-management PageHeader (`Bau`, `Bauverwaltung · Kolonie #ID`, Planet subtitle, large return action) with the same compact toolbar pattern as Colony Detail: only `Bauverwaltung` + a small `Zurück` action.
- Split the right construction column into two non-duplicated concepts: the current build is rendered exactly once with its progress bar, PP progress, explicit percentage and ETA; the Queue below contains only future items.
- Queue badge/count now reports only future queued projects (`items.slice(1)`), not current + future combined.
- The current project is no longer repeated as queue position 1. Browser smoke after planning Troop Transport once showed `Aktuell / Troop Transport / 0.0 / 100 PP · 0% / 34 Runde(n)` and `Queue 0` with zero queue rows.
- Planning the same repeatable project a second time kept the first as Current and produced exactly one future Queue row (`Queue 1`, position 1, ETA 67 turns), proving the duplicate-list presentation is gone.
- Future Queue reordering is now constrained to the future tail: the first future item's Up action is disabled, so a queue reorder cannot silently replace the current build and bypass the explicit current-build abort confirmation. Current-build removal still uses the existing confirmed abort flow and PP-preservation semantics.
- Empty construction state renders only one `Kein Projekt eingeplant.` current-state message plus `Queue 0`; there is no duplicate empty/current queue row.
- Responsive smoke at 320x640 retained the same stacked Catalog -> Project -> Current/Queue order, with a 40px compact header and no horizontal overflow.
- `npm run build` - **PASS**.

#### G4-B2-V - Colony construction progress + live metric breakdown

Status: **complete; direct user review pending**.

- Colony Detail's compact Build card now mirrors the manager's authoritative current-project progress: current PP / cost PP, explicit percentage, progress bar and projected ETA. The queue badge is future-only, matching B2-U semantics.
- Browser smoke with Troop Transport: before resolution the Colony card showed `0.0 / 100 PP · 0%` and `34 Runde(n)`; after `Fertig` advanced to Turn 2 it showed `3.0 / 100 PP · 3%`, `32 Runde(n)`, and the progress fill measured 3% of the track.
- Added additive player-safe `PlanningColonyBreakdowns` to the Planning Preview. The breakdown is server-derived and travels with the same disposable draft state used for live Planning Preview; React only formats/localizes it.
- Growth breakdown components: natural/race growth, population-growth technology, Housing, Cloning Center, starvation and a guarded `other` residual. Housing's percentage calculation was extracted into one shared authoritative helper used by both actual population dynamics and the breakdown to prevent drift.
- BC breakdown components: population tax base, government modifier, morale contribution and guarded `other` residual. Components sum to the exact projected `adjusted_economy.tax_bc` total.
- Unit coverage verifies the breakdown sums to authoritative totals; Small Fixture baseline yields 4 BC population tax + 2 BC government = 6 BC. A Holo Simulator fixture produces a positive morale BC component. Cloning Center produces a positive growth component.
- Colony Profile now exposes a small tap/click `Kolonie-Info` disclosure containing Growth and Colony-BC totals plus their component rows. Positive component values use the existing green signed tone, negative values red.
- Live Housing smoke: planning Housing as the current project changed Colony growth from `+0.07 (14 turns)` to `+0.09 (11 turns)`, and `Kolonie-Info` changed from only `Natürlich / Volk +0.07` to `Natürlich / Volk +0.07` + `Housing +0.02`; BC remained 4 population tax + 2 government = 6.
- Responsive QA: desktop metric popover is bounded to `240px` at x=24..264 with no horizontal overflow. 320x640 smoke: Colony Profile `308x113px`, metric popover `240x127px` at x=65..305, Build card `308x147px`, no horizontal overflow.
- Important mechanics boundary: `cloning_center` and morale buildings are already implemented economy effects and therefore appear in authoritative breakdowns. `planetary_stock_exchange` / a Bank-style direct BC building bonus is catalogued but **not yet an implemented EconomyRules money modifier**, so Slice 15.2 does not fabricate that contribution; it remains economy/building fidelity work.
- `npm run build` - **PASS**.
- targeted Game breakdown tests - **PASS**.
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
