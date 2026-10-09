# Slice 17.8 - Directional Shield Segments & Shield-Piercing HMI

Date: 2026-10-09
Status: browser QA passed; explicit manual user approval still pending.

## Request

Expose the natural shield-piercing nature of Particle Beams, and change the single all-or-nothing cyan shield ring into four directional sectors whose actual charged fractions can be recognized at a glance.

## Source-of-truth mapping

- Four server-authoritative shield_arc_current entries with per-sector shield_arc_max are used; there is no client-side damage simulation or inferred hitpoint depletion.
- Sector 0 front at 0 degrees, sector 3 right at 90 degrees, sector 2 rear at 180 degrees, sector 1 left at 270 degrees, aligned to the same relative-angle signs as internal/battle/tactical.go tacticalShieldImpactArc.
- The ring group rotates by ship.facing * 22.5 degrees, following the sprite heading.
- Each quarter has a faint whole-sector track and a cyan charged arc with SVG geometric sweep proportional to current/max. Four separated spans have a 18-degree total gap between neighboring quadrants.
- An all-zero shield or a destroyed ship has no ring. Fully depleted individual sectors retain only their dim background track while other sectors still have charge.
- Scan overlay exposes numerics for Front, Right, Rear, Left from the same state, with English and German locale strings.

## Shield-piercing status

- The active ship's weapon strip labels Particle Beam and Beam mounts with shield_piercing modifier or natural_shield_piercing innate effect as Shield-piercing / Schilddurchdringend.
- Beam presentation reads actual beam_fired event weapon_id and modifier_ids, plus the source weapon's innate effects. It draws the cyan Shield-piercing hit text only on an actual successful ship hit; not while shooting projectiles or missing.
- If the attacked ship currently reports a functional hard_shields special, the on-hit piercing message is suppressed, consistent with the battle engine's ShieldPiercing && !HardShields gate. The weapon strip still describes the intrinsic weapon capability.
- No server, weapon, damage, shielding or protocol rules are changed.

## Isolated real-Chrome QA

Instance 127.0.0.1:7174, independent Falcon 3 versus Darlok 3 reference game. No canonical 7171 tactical action was issued.

1. Fresh game: six segment groups (one per ship), 24 quadrants, four charged paths on each fully charged ship, all ratios 1.00. Three Particle Beam slot badges displayed Schilddurchdringend.
2. Human Heavy Particle Beam ×3 fired at Darlok ship 45: the browser observed a successful particle_beam hit with the Schilddurchdringend message, computed cyan text color rgb(143,223,255). Actual battle_damage_applied events showed armor losses of 33 each without any shield-charge loss.
3. Scan inspection of target ship displayed Vorne 100/100, Rechts 100/100, Hinten 100/100, Links 100/100.
4. After actual Zeon Rocket launch and the following AI round, target ship 45 had authoritative shield arcs [100,100,75,100] and the rear sector rendered with data-current 75, data-fraction 0.75; all three other sectors remained full.
5. Real SVG getTotalLength measured charged rear arc 0.725676 versus full rear track 0.967660, ratio 0.749929 (75% geometrically). Computed stroke rgb(104,216,255), width 2.35px.
6. Post-battle smoke test reloaded the final built Web asset. One ship was destroyed, leaving exactly five visible shield rings, 20 total quarter segments, and the expected partial rear segment. Human movement HUD and animation camera gates remained operational.

## Verification

- npm run build (Tactical browser contract, TypeScript, Vite production output): PASS.
- go test ./internal/battle ./internal/server -count=1: PASS.
- git diff --check: PASS.
- Git push: none. Slice 17.8 remains OPEN. Slice 17.9 not opened.
- Canonical 7171 was intentionally not restarted or reset for this review change; the server dynamically serves the fresh Web assets.

## Canonical 7171 read-only verification

- The rebuilt frontend is served by canonical port 7171 with HTTP 200, and its index bytes match the local Web production build.
- Read-only Chrome inspection showed the current persisted in-memory review encounter still in Round 1 with active Human ship 40. Next command sequence was already 6 at observation time, reflecting pre-existing game interactions; this work did not reset or submit a Tactical command to the canonical game.
- The canonical frontend rendered 24 shield-sector nodes and the three Particle Beam piercing capability labels.
- No 7171 server restart was performed. The user can continue reviewing their current battle.
