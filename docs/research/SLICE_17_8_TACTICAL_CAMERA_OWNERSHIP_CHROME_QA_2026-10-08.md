# Slice 17.8 – Tactical camera ownership / Chrome QA

Date: 2026-10-08
Status: implemented and verified for manual user review.

## User-reported problem

During a Human Tactical activation, targeting or firing Particle Beams, Zeon/Merculite missiles, and Fighter/Bomber bays caused the camera to frame the weapon path and finally jump/zoom back to the active Falcon, discarding the user's manually chosen battlefield overview. The requested behavior is **no automatic camera movement while the Human acts**, with automatic focus/follow only during non-Human activations.

## Isolated real-Chrome reproduction before fix

- The pre-fix reference fixture was reproduced on isolated port 7174, not canonical 7171.
- After manually zooming out, firing a Zeon missile changed the SVG battlefield `viewBox` from `-9.312931944775812 -14.80897318617097 68.62586388955162 47.61794637234194` to `19 -0.7142857142857135 28 19.428571428571427` during flight and then to `17.34375 3.6875 15.3125 10.625` centered on the active Falcon.
- Causes: unconditional `focusCombatSpan` during cue playback, automatic `FOCUS_ZOOM=3.2` camera return after the queue, active-ship autofocus, and viewport extents changing with moving actors/legal cells.

## Browser-only correction

- Human Tactical starts with battlefield overview, not active-ship 3.2x zoom.
- Manual pan, wheel zoom, pinch zoom, and intentional Scan focus remain tracked as the player's viewport.
- Human activation and own-ship selection no longer auto-focus.
- Each visual cue is tagged with camera-follow policy reconstructed from authoritative `round_started.initiative_order` and `activation_ended` / `activation_auto_ended` events. This matters when the server sends an entire AI activation with the next Human activation already active in its final snapshot.
- Cues execute and lock commands unchanged in both turns; only non-Human cues steer the camera.
- After AI playback has finished, the previously chosen Human camera center and zoom are restored if control has returned.
- A stable initial viewport extent prevents the SVG viewbox changing merely because fighters, projectiles or legal movement cells change bounds. Camera centering/zoom is still allowed for non-Human cues.
- No combat rules, damage, animation VFX geometry, protocol commands or server data were modified.

## Isolated Chrome verification after fix

A fresh 3-vs-3 Falcon Combat Review was run on isolated port 7176 with the rebuilt Web UI:

- Human Zeon missile fired: identical `viewBox` before, throughout lock/flight, and after unlocking.
- Human Heavy Particle Beam volley: identical `viewBox` before/during/after.
- Human Bomber Bays launch: identical `viewBox` before/during/after.
- When Human ended the three activations, the server sent AI event history already advanced to the next Human ship. Despite that final ownership, the camera *did* follow the AI combat cue path and moved repeatedly to relevant combat positions.
- After the AI presentation finished and user controls unlocked, the camera returned **exactly** to the user's manual pre-turn view: `-6.141910609567475 -21.976019606638655 89.28382121913495 61.95203921327731`. No automatic 3.2x snap.
- The presentation input interlock was visibly true while effects were playing and false when the queue finished.
- The isolated QA instance was not canonical 7171, and no canonical gameplay input was generated during these tests.

## QA

- `web/npm run build`: PASS (includes tactical ordnance contract and TypeScript production build).
- `go test ./internal/battle ./internal/server -count=1`: PASS.
- `git diff --check`: PASS.
- Live real-Chrome camera observation: PASS for Human rocket, Beam, Bomber, AI follow, and restoration.

Slice 17.8 stays OPEN pending explicit user visual acceptance. No new Slice 17.9, no unrequested push.
