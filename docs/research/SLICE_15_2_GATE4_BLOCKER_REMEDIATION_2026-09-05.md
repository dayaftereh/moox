# Slice 15.2 Gate 4 blocker remediation

Date: **2026-09-05**

Status: **G4-A complete; G4-B pending**.

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

Status: **pending**.

- Replace the current below-map system detail with an explicit star-system dialog/overlay opened by star selection.
- Make orbital bodies, fleets, colony/outpost status and legal actions immediately understandable.
- Reduce Colony text density and make current production/build queue prominent.
- Improve Population presentation toward direct Farmer/Worker/Scientist manipulation while preserving server-projected legality and Planning Preview authority.
- Keep richer 2D Colony/building artwork as later 15.x presentation work unless required for usability.

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
