# Slice 15.2 Gate 4 blocker remediation

Date: **2026-09-05**

Status: **G4-A and G4-B complete; G4-B2 interaction refinement accepted; G4-C independent QA/closure deferred until B2 is complete**.

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
