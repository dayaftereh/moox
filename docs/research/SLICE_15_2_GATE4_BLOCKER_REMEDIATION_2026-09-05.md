# Slice 15.2 Gate 4 blocker remediation

Date: **2026-09-05**

Status: **G4-A and G4-B complete; G4-C independent QA/closure pending**.

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
