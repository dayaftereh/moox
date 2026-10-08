# Slice 17.8 – Tactical VFX Pivot Chrome Audit

Date: 2026-10-08
Status: implemented for user review.

## Live Chrome reproduction, before patch

- A real Particle Beam impact drifted visibly away from its target while its scale animation ran.
- Chrome control: a reference SVG circle and an identically positioned tactical-beam-impact were compared with getBoundingClientRect(). The effect drifted approximately 19.6 px at scale .2 and 29.3 px at scale 2.2. Computed CSS transform-box was view-box instead of fill-box.
- tactical-move-arrival similarly drifted 13.15 px at scale .45 and 23.9 px at scale 2.0. tactical-projectile-event-pulse already remained anchored because it used fill-box.
- A debris fragment with SVG rotate(78) was forced to 0 degrees while its CSS keyframe animated transform: scale(...) on the same SVG line. The CSS transform overrode the SVG rotation.
- A real Zeon missile movement followed the expected event route from (55,9) to (39,9). During movement, the server snapshot may already contain the projectile at its final location, showing a second static rocket/engine on top of the flying rocket.
- The bright rocket flame independently animated its SVG center coordinate (cx from -0.56 to -0.92), shifting backward relative to the exhaust/nozzle.

## Presentation-only fix

- Beam impacts and movement arrival pulses now scale about their own SVG fill-box center.
- Explosion debris rotation is on a parent SVG g while the fragment scales locally along its own axis.
- While a projectile_moved flight cue is presented, the matching already-final static projectile is hidden; it returns automatically when the flight cue ends.
- The bright rocket flame remains fixed to the nozzle at local cx=-0.62; CSS pulsing handles its animation without moving its position.

## Chrome verification after patch

- The updated Web build on isolated 7173 has fill-box transforms on the Beam impact and movement-arrival pulse. Their measured position drift is 0 to 0.0001 px, matching the already-correct projectile pulse.
- An animated debris fragment retains its 78-degree direction.
- A real Beam volley event sequence 8 retained a 0 to 0.0001 px impact difference from a same-location reference SVG circle even while the camera viewBox changed.
- A real Zeon missile launch showed the flying rocket/exhaust while the static projectile count stayed zero during the movement cue. The static marker returned after playback; the new flame has no independently animated cx coordinate.
- The Web build, Tactical contract and git diff --check passed. No tactical rules or server event data changed.

## Scope

Only Web presentation CSS, SVG rendering and the Tactical contract test changed. The canonical reference game can be restarted for an untouched visual QA run. Keep Slice 17.8 open pending user acceptance. No push.
