# Planned Slice 17.8A - Tactical GPU-Composited Camera Gestures

Status: **CONDITIONAL PLANNED / NOT OPEN / NO IMPLEMENTATION AUTHORIZED** (2026-10-11).

This is a prepared post-17.8 mobile-rendering follow-up, not a second currently active Slice-17 child. The sole live `_OPEN_` marker remains Slice 17.8 until the user's visual acceptance is resolved. Do not open 17.8A while another slice is open or silently rewrite 17.9.

## Why prepare this now

The user repeatedly reports slow Android Tactical 2D pan/pinch/zoom. Existing optimizations (commit `ce8451c`: requestAnimationFrame-batched SVG viewBox update, touch shadow suppression; commit `5819db1`: 904 -> 524 DOM nodes, 42% fewer) improve code structure and SVG cost, but **no actual Android frame-time baseline or confirmed frame-rate improvement exists**.

The likely next bottleneck is that changing the SVG `viewBox` every gesture frame can cause repeated SVG layout/raster/paint, even if React re-render is avoided. This is a hypothesis to profile, **not** a verified phone GPU diagnosis.

## Decision and ordering

1. **Immediately** document this candidate, without starting implementation and without creating a new `_OPEN_` marker. Keep current Slice 17.8 User Review as the only open slice.
2. **First** evaluate `5819db1` on the user's actual Android phone with the current representative Triangle Combat 3v3 scene. Record responsiveness, frame times/long frames where tooling permits, and whether selection, move, weapon targeting and health rings work.
3. If mobile navigation is **acceptable**: finish Slice 17.8 only upon user approval. Keep 17.8A parked. Proceed to prepared 17.9 as already planned; do not automatically build a new renderer.
4. If performance is **unacceptable and blocks 17.8 acceptance**: keep 17.8 open and agree a bounded follow-up decision. Do not open a second parallel slice. If the required redesign cannot safely fit the active review, explicitly sequence a single focused 17.8A before 17.9 once the active marker is resolved.
5. If 3v3 is acceptable but larger battles fail: treat 17.8A as conditional scalability work. Schedule it relative to 17.9 based on measured acceptance risk; do not quietly make it a new blanket prerequisite.

**Default recommendation:** design and measure now; implementation only if user testing or a concrete 17.9 browser-acceptance gate justifies it.

## Preferred technical approach - Stage A

Implement a **temporary CSS compositor-friendly camera preview during a gesture**; retain the same authoritative Tactical SVG and existing React camera contract:

- Separate an outer clipped camera viewport from a single inner movable Tactical world layer. HUD and fixed interaction controls must not be transformed with the world.
- On pointerdown, capture the committed `viewBox`, transform origin, starting pointer positions and the current camera matrix.
- On one-finger pan / two-finger pinch, update only the world's CSS `transform: translate3d(...) scale(...)` (or equivalent compatible matrix) through `requestAnimationFrame`; avoid recalculating every SVG child or changing `viewBox` on every touch frame.
- Apply `will-change: transform` **only during active interaction** where it improves actual compositor behavior; remove it on release to avoid permanently consuming GPU memory.
- Anchor pinch around the actual two-finger midpoint, not the SVG origin; handle scroll/zoom bounds and letterboxed aspect ratios correctly.
- On release, atomically commit the equivalent camera center/zoom to the normal React/viewBox state and remove the preview transform without a visual jump, and maintain user camera ownership for Human vs AI turns.
- Make pointer hit testing use the **effective transformed world matrix** while previewing, with correct rounding for legal movement and selection. Verify whether `getScreenCTM().inverse()` includes the CSS transform consistently on Chrome Android; if not, combine the preview matrix explicitly.
- Preserve ship SVG drawing, true four-way directional shield, global armor/structure status arcs, projectile animations, accessibility and screen-readable labels.
- Handle `pointercancel`, orientation/resize, finger-count transitions, presentation input locks, ship dragging, and browser lifecycle without a stale active preview.
- Maintain a clean fallback: existing non-composited SVG `viewBox` camera for browsers/devices where transform compositing does not improve results.

CSS transforms are **eligible** for hardware compositing, not a guarantee. Verify actual compositing in Android Chrome rather than promising that all devices will use the GPU.

## Alternative, only if Stage A fails

A separate Canvas 2D or WebGL world renderer can reduce SVG DOM/raster costs further but is significantly larger: sprites and health-ring presentation, z-order, hit tests, high-DPI sizing, accessibility mirror and animation parity must be reimplemented. **Do not choose Canvas/WebGL by default** or implement it inside this small follow-up without a new architecture decision and acceptance evidence.

## Gate 1 - on-device audit / decision (NOT STARTED)

- [ ] Recheck live repo, sole open marker, mobile build version and actual canonical vs isolated game state.
- [ ] Confirm the user's Android browser/device, viewport resolution/devicePixelRatio and which navigation gestures stutter; avoid recording personal/private device identifiers.
- [ ] Capture repeatable 3v3 pan, pinch and long-scene scroll traces via Android Chrome remote inspection or in-app frame-time instrumentation where available.
- [ ] Separate input dispatch, JavaScript, SVG layout/paint, compositing and GPU bottlenecks using actual device evidence.
- [ ] Test a representative denser fleet/ordnance scene where a safe fixture exists; no destructive changes to canonical 7171.
- [ ] Record baseline frame interval p50/p95, long frames, touch-to-visual latency, memory/battery concerns, and interaction correctness. Suggested **targets**, subject to baseline/device: p95 under ~33 ms for sustained 30 fps, aim for smooth 60 fps (~16.7 ms), and measurable improvement over the same-device baseline.
- [ ] Decide whether a compositor prototype is justified, or whether further SVG draw simplification is cheaper and sufficient.

## Gate 2 - rendering and input contract freeze (NOT STARTED)

- [ ] Freeze layer boundary, transform math, pinch anchor, camera clamp/restore, touch-action and event priorities.
- [ ] Freeze mapping between CSS preview matrix and authoritative SVG/legal move cell coordinates.
- [ ] Freeze cancellation, resize, active ship transition, animation and AI camera-follow behavior.
- [ ] Freeze compatibility/fallback strategy and performance instrumentation; no combat/server API changes.
- [ ] Write before/after scenario and screenshot parity matrix: 3v3, larger battle, arrows, shields, occupied tiles, click/drag move, scan/target, missiles and fighter effects.

## Gate 3 - staged implementation (NOT STARTED)

- [ ] Implement the camera layer behind an isolated feature switch in a separate 7174 test service; do not change canonical game commands/state.
- [ ] Prove compositing is actually enabled when the gesture is active; profile render cost while panning/pinching.
- [ ] Validate no drift/jump after release, correct tapped target and movement, and touchcancel handling.
- [ ] Retain existing `getScreenCTM` / authoritative target protections; preserve regression tests and data attributes.
- [ ] Promote only after QA, with clean Git and explicit user review.

## Gate 4 - mobile/desktop acceptance (NOT STARTED)

- [ ] Actual Android device side-by-side baseline and candidate tests, repeated under comparable conditions.
- [ ] No regressions in small and large fleets; no incorrect ship selection, targeting or movement.
- [ ] Desktop mouse wheel, trackpad, touch, zoom buttons, keyboard/assistive interactions and camera ownership remain functional.
- [ ] No excessive compositor layers, permanent `will-change` GPU memory usage, artifacts, flicker, clipped rings or blurry post-gesture labels.
- [ ] User explicitly approves touch responsiveness and visual result before closing the follow-up.

## Ownership / exclusions

Presentation/browser-only: `web/src/TacticalBattlefield.tsx`, `web/src/styles.css`, focused tests, accessibility and research QA. **Do not modify** authoritative Go tactical mechanics, RNG, legal move rules, game serialization, replay, shield/damage values, game save state or reference-fixture data. No push/reset of canonical 7171, no automatic opening of Slice 17.9.

## References

- `docs/research/SLICE_17_8_TACTICAL_MOBILE_CAMERA_PERFORMANCE_FOCUS_BADGE_2026-10-09.md`
- `docs/research/SLICE_17_8_TACTICAL_SVG_DOM_REDUCTION_2026-10-10.md`
- `docs/slices/PLANNED_17_9_FULL_SHIP_DESIGNER_TACTICAL_INTEGRATION_ACCEPTANCE.md`
- `docs/slices/README.md` - sole-open-slice protocol and planned specifications
