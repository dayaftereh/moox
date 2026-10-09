# Slice 17.8 - Mobile pan/pinch performance and unambiguous selected-ship badge

Date: 2026-10-09
Status: Browser QA complete; personal Android user review outstanding.

## User observation

The small arrow with number 1/2/3 appeared to indicate a flight direction and was unclear. Tactical Battlefield pan, pinch-to-zoom and scrolling were slow on the user's Android phone.

## Reproduced architectural hot path

- The Tactical scene contains approximately **904-905 SVG elements**, including around 335 paths in the isolated six-ship reference battle and **72 path drop-shadow effects** on a desktop-capability browser.
- Before the change, each pointermove while panning/pinching called React `setCamera`, reconciling the complete tactical React view and re-evaluating expensive multi-sector ring SVG paths. Rapid wheel events used the same expensive path.
- The agent's managed Chrome runs with `document.hidden === true`; browsers can suspend native `requestAnimationFrame` while hidden, so functional gesture QA used a local test-tab-only timeout-backed RAF shim. This is a correctness/architecture test, **not** an Android hardware frame-time benchmark.

## Implemented UI

- Removed directional triangle/arrow completely. The single selected, active own ship receives a clear **'Ship 1' / 'Schiff 1'** (2/3 as applicable) number plate below its three concentric condition rings. The plate uses cyan when movement is possible, coral for attack-only and muted gray when neither is possible.
- Labels and viewport hint translated EN/DE; hint explicitly states the label is a ship selection marker, not the heading.
- The same authoritative stable per-side ordinals are retained; no new ship IDs or combat mechanics.

## Implemented optimization

- Pointer pan and pinch keep the current live camera in `liveCameraRef`, update the SVG viewBox and background rect imperatively at **most once per requestAnimationFrame**, and commit the resulting React camera state once when the final pointer lifts.
- Rapid mousewheel / trackpad zoom uses the same live viewport scheduling and an idle 110ms commit rather than a state update for every wheel delta. The original button-zoom, automatic opponent-camera and manual camera restore remain.
- Memoize `TacticalShipConditionRings`, avoiding repeat geometric path creation when ship and facing are unchanged.
- On actual touch/coarse-pointer hardware, drop SVG drop-shadow filters from all three health-ring charge paths and procedural ship glyphs, preserving distinct crisp blue/yellow/red strokes. The override matches the original selectors' specificity so it wins in the cascade.
- Click-to-move and ship drag-to-move determine the target cell via `SVGSVGElement.getScreenCTM().inverse()` and `DOMPoint`, which is accurate under an imperatively moved viewport **and** under SVG preserveAspectRatio letterboxing on narrower screens.
- Added regression contracts rejecting direct React camera updates from pointermove and individual wheel events, verifying all three touch GPU shadow overrides and both SVG-matrix hit-test paths.

## Isolated real-Chrome browser QA (port 7174)

1. Fresh 3v3 Triangle Combat reference game, initial 904-905 SVG elements, 18 health rings, three layers on each of six ships. Default focus label showed 'Schiff 1' below own Falcon 1, with **zero** directional-arrow paths; label SVG text was visually contained within its plate.
2. Simulated one-finger touch pan shifted actual SVG viewBox and backing rectangle before pointerup; state committed on pointerup without mismatch. During the gesture, the component's last committed React camera remained unchanged while live SVG viewport moved.
3. Simulated two-pointer pinch changed viewBox from width 49 to width ~33.31, kept the background width aligned, and committed the camera after lifting the last finger.
4. Simulated rapid wheel zoom applied deltas to the live SVG viewport and committed only after the 110ms idle timeout. When verifying React hooks, alternate fiber state contained the committed zoom ratio; the managed hidden Chrome can lag state presentation in a stale fiber pointer.
5. After pan and pinch, a real authoritative move command succeeded by clicking the correctly projected legal destination using the SVG screen transformation: Falcon ship #41 moved from (25,11) to (26,11), `ship_moved`, movement 15.
6. Dragging that own ship to a second legal cell using pointerdown / pointermove / pointerup also succeeded: (26,11) -> (27,11), movement 14, without camera panning.
7. Finishing the prior ship activation changed the displayed plate to 'Schiff 2'. Existing three rings and focus state persisted.
8. CSSOM check found coarse-pointer media rule with all three explicit high-specificity shield/armor/structure shadow overrides, plus procedural ship glyph; desktop test browser correctly reports no coarse pointer, so we did not claim to have benchmarked Android GPU frames.

## Automated gates / deployment restrictions

- `npm.cmd run build` (encoding, Tactical/ordnance contract, TypeScript and Vite): PASS.
- `go test ./internal/battle ./internal/server -count=1`: PASS.
- `go vet ./...`: PASS.
- `git diff --check`: PASS.
- Canonical game on port 7171 must stay untouched (recently reset to original reference planning state on request); no unrequested turn advance, reset, restart, or Git push.
- Slice 17.8 remains the sole OPEN Slice. Manual Android performance and visibility acceptance is pending.
