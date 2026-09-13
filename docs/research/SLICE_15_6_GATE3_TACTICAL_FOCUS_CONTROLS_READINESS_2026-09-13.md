# Slice 15.6 Gate3 - Tactical focus zoom, unified controls and readiness roster

Status: **IMPLEMENTED / isolated live-battle browser QA green**
Date: 2026-09-13

## User feedback
The Tactical battlefield opened too far away, making the actual engagement hard to read. The bottom HUD also spent a large independent area on Scan while the activation controls lived separately. The ship roster needed at-a-glance remaining movement and weapon readiness. A follow-up question asked how future multiple-weapon selection and split fire should work.

## Camera refinement
The Tactical camera now starts on the authoritative active ship when a Tactical state is already available, instead of first opening on the generic battlefield center. `FOCUS_ZOOM` increases from 2.2 to **3.2** and the camera follows authoritative active-ship changes at at least that zoom.

This deliberately does not try to fit every own ship into the viewport. The player can still pan and wheel/pinch zoom normally.

Isolated browser QA against a clone of the user's current Turn-6 Battle #2 showed a Tactical viewBox width of **13.75 world units** (`44 / 3.2`) and centered on the active ship from the first Tactical frame. When the authoritative active ship changed from ship 23 to ship 35, the camera recentered on ship 35 while retaining the 3.2 focus zoom.

## Unified bottom controls
Scan no longer owns a large standalone HUD column. The bottom HUD is now two conceptual areas:

- `Noch verfügbare Schiffe`
- one `tactical-hud-controls` card containing Scan, authoritative weapon-slot choices, Wait, Finish, Retreat and Back-to-Encounter.

Compact GameIcon glyphs were added to the controls. Desktop/tablet uses compact horizontal/grid placement; narrow mobile keeps touch targets usable and may wrap inside the single controls card instead of creating separate functional panels.

## Ship readiness roster
Each own roster card now derives its status only from `TacticalShipView`:

- movement: `movement_current / movement_max`
- weapon readiness: sum of weapon `count` whose authoritative `ready` flag is true / total weapon count.

The values are displayed compactly below the ship glyph and are also included in the button's accessible label/title. The current live-battle clone correctly rendered the two unarmed Scouts as `20/20 · 0/0` and the armed Scout as `20/20 · 1/1`.

## Weapon selection boundary
No new client combat authority was introduced. Existing Tactical authority already exposes `legal_fire_actions` with an explicit `weapon_slot`, and the UI already submits `battle.fire_beam` using the selected server-projected slot. The unified control card makes those slot choices more visible.

The requested future case `4 Laser -> 2 shots at Scout A + 2 shots at Scout B` is **not** implemented in this UI block. Current server semantics mark an entire weapon slot ready/not-ready and the current Slice-07 Laser fixture explicitly accepts only a Laser mount with `count == 1`. Partial consumption/splitting therefore needs a separate server-authority slice: define fire-group/ammo/count semantics, project legal partial-fire actions, resolve damage authoritatively, then add the corresponding selection UI.

## Verification
- `npm run build` green.
- `go test ./...` green.
- Isolated server `7186` restored from a read-only export of the user's current canonical Triangle Turn-6 / Encounters / revision-17 state.
- Real browser entered Battle #2 through the normal encounter flow.
- Initial active ship opened at 3.2 focus zoom.
- Active-ship change recentered at 3.2.
- Roster movement/weapon readiness values visible and accessibility labels correct.
- Combined control card contains Scan + selected Laser Cannon slot + Wait + Finish + Retreat + Back.
- Weapon slot button is visibly selected from authoritative `legal_fire_actions`.

Canonical 7171 was not mutated during this isolated QA block.
