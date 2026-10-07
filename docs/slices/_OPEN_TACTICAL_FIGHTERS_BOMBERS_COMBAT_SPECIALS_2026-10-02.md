# OPEN Slice 17.8 - Tactical Fighters / Bombers / Combat Specials

Reopened for user review: 2026-10-07

Formal Gate-4 close exists at 77854f8, but final user acceptance is still pending.
Slice 17.9 remains planned and must not be opened until explicit user acceptance.

## Current review request

Add a dedicated deterministic all-tech combat-review reference game without modifying the existing Triangle games:

- neutral central Fight System;
- current Falcon MK1 review loadout;
- three Falcons for Human and three identical Falcons for one built-in AI;
- Human and that AI already at war;
- normal strategic encounter/Tactical BattleSession path on the next resolution boundary.

No Slice-17.9 work belongs in this phase.

## Permanent reference-game decision

User decision on 2026-10-07: `game-triangle-combat-review` is a permanent member of the MOOX development Reference Games. It must remain registered whenever `-reference-games` is enabled. Later slices may evolve its combat payload, but must not silently remove its stable game/scenario identity or replace it with a disposable fixture.

## User-review fix - Point Defense vs Fighter/Bomber

- Combat Review Falcon keeps an explicit Point Defense Particle Beam and now uses a 360-degree firing arc so the permanent review fixture can exercise manual PD fire against hostile Fighter/Bomber groups deterministically.
- Global combat legality is unchanged: Standard and Point Defense Beams may target Fighter groups, Heavy Beams may not, and configured firing arcs remain authoritative.
- Tactical weapon controls expose PD/Heavy mount identity; PD also shows its arc, and legal Fighter/Bomber targets receive a stronger target highlight.

## User-review fix - moving-actor target layering

- Ships render below live projectile/Fighter actors on the Tactical battlefield so a ship sprite/hitbox cannot cover a legal missile/Fighter/Bomber target.
- Destroyed ships do not consume combat pointer input; scan mode can still inspect them.
- Projectile/Fighter hitboxes consume pointer input only while that actor is a legal target for the selected action, allowing clicks to fall through to ship targets otherwise.
- The authoritative combat rules and logical actor coordinates are unchanged; this is a browser interaction/layering correction only.

## User-review fix - contextual Tactical cursors

- Legal weapon targets (Ship, projectile, Fighter/Bomber group) use a target crosshair cursor.
- Only server-projected legal movement cells use the movement cursor; illegal battlefield cells retain the normal pan/grab cursor.
- Cursor affordance therefore mirrors the currently selected authoritative action without deriving legality in the browser.

## User-review fix - Fighter/Bomber ship-cell separation

- Outbound Fighter, Bomber and Heavy-Fighter groups targeting a Ship attack from a deterministic free adjacent attack station instead of entering the Ship's own tactical cell.
- Ship cells remain physically exclusive to Ships. The server chooses the nearest free neighboring attack station, spreads groups across free stations where possible, and only allows Fighter groups to share an adjacent station after all neighboring stations are occupied.
- Existing pre-impact defenses, payload cadence, retargeting and return/recovery semantics remain authoritative; only the ship-target arrival coordinate is refined.


## User-review cosmetic fix - Fighter/Bomber attack feedback

- The Tactical browser now consumes authoritative fighter_attack_resolved events for visible attack feedback.
- Fighter Beam payloads render a short beam pulse from the Fighter group to its authoritative target.
- Bomber Bomb payloads render a short bomb-run trail and impact pulse.
- Heavy Fighters can display both Beam and Bomb payload events from the same attack cycle; misses are rendered with subdued feedback.
- The animation does not derive combat results or timing; it only visualizes already-resolved server events.

## User-review cosmetic pass - miniature Fighter/Bomber formations

- Fighter-family groups keep their existing server multiplicity and large interaction hitbox, but the battlefield now renders one small aircraft silhouette per actor in the group instead of one oversized marker plus xN text.
- Fighter, Bomber and Heavy-Fighter silhouettes are visually distinct and auto-scale/pack into the tactical cell as group size grows.
- This first cosmetic checkpoint changes presentation only; targeting, group identity, movement and combat semantics remain unchanged.

## User-review cosmetic pass - miniature projectile salvos

- Tactical projectile groups now use the same in-cell mini-formation principle: one small missile silhouette per authoritative multiplicity, both at rest and during projectile flight animation.
- Visible projectile xN labels are removed; the large transparent projectile hitbox remains unchanged for usability.
- Projectile count, targeting, movement, interception and damage remain server-authoritative and unchanged.

## User-review cosmetic pass - sci-fi Fighter propulsion detail

- Miniature Fighter-family actors now layer the existing silhouette with a luminous cockpit/energy canopy, panel/wing accents, engine cores and animated exhaust flames.
- Bombers use twin engine cores and twin exhaust plumes; Fighter/Heavy-Fighter actors use a single central engine plume.
- Exact group multiplicity, actor identity, targeting and the large interaction hitbox remain unchanged.

## User-review cosmetic pass - sci-fi missile propulsion detail

- Miniature missile actors now add a luminous nose, cyan energy spine, bright engine core and animated orange/yellow exhaust plume while retaining the compact exact-multiplicity formation.
- Projectile flight animation carries the same nose/spine/engine language with a longer animated plume and bright flame core, making moving salvos read as powered sci-fi missiles rather than simple geometric markers.
- Browser contract coverage now requires the new static missile plume/energy detail and Fighter-family propulsion detail.

## User-review fix - direct diagonal moving-actor trajectory

- Shared Fighter/projectile partial movement now spends the full movement-field quantum first and chooses the straightest in-budget grid endpoint toward the locked target before using remaining distance as a tie-breaker.
- This prevents the previous axis-heavy behavior where a Fighter/projectile aimed several rows away advanced almost entirely on X before correcting Y on a later round.
- Diagonal movement remains governed by the existing ceil-Euclidean tactical field metric; movement budget, speed, impact/arrival, retargeting and scheduling are unchanged.

## User-review fix - projectile animation origin

- Projectile flight animation now anchors the SVG group statically at the authoritative from_x/from_y before any animation frame and animates only the relative delta toward to_x/to_y.
- This removes the transient render at SVG (0,0) that made freshly fired missiles appear to pull in from the upper-left corner.
- The animated flame now uses transform-box: fill-box with a local center origin so pulse scaling cannot shift the exhaust particle around the global SVG viewport.

## User-review cosmetic fix - event-driven combat camera

- Tactical presentation now queues authoritative combat events instead of collapsing a multi-event AI/moving-actor batch to the final animation state.
- The camera auto-frames the action span for Ship movement/Beam fire, Fighter-family movement/attack and projectile flight so both source and target remain visible where possible; low-value launch/retarget/impact/recovery bookkeeping does not create duplicate camera stops. Outbound Fighter movement is folded into the attack cue when that attack immediately follows in the same authoritative event batch, and reactive PD volleys use a shorter hold.
- Fighter movement receives a distinct cyan trail; returning Fighter movement uses an amber trail so carrier recovery is visually legible rather than appearing as an unexplained disappearance.
- Fighter lifecycle rules are unchanged: Fighter/Bomber/Heavy-Fighter attack stock remains 4/1/2 respectively; exhausted groups return to a functional carrier and recover, while intercepted groups are destroyed.

## User-review refinement - calmer combat camera and one-click targeting

- Combat presentation pacing is deliberately calmer: Beam, Fighter, movement and projectile cues remain event-driven but hold longer before the next camera cut.
- Combat framing keeps wider context around source and target and caps automatic combat zoom at 2.8, avoiding the previous close-up 4.4 zoom.
- Tactical mouse targeting now tolerates normal pointer jitter. Single-pointer motion must accumulate 4 CSS pixels before the battlefield is treated as a pan/drag, and the camera does not pan at all below that threshold.
- The target click handler remains unchanged and server-authoritative; the usability fix prevents tiny pointer motion from suppressing the first legal Ship/Fighter/projectile click.

## User-review refinement - combat presentation 20 percent slower

- The complete Tactical presentation cadence is slowed uniformly by factor 1.20 without changing authoritative combat timing or gameplay rules.
- Cue holds are now reactive PD 624 ms, Beam 1320 ms, Fighter attack 1500 ms, movement 840 ms and projectile 900 ms.
- Matching finite Beam/Fighter/movement/projectile animation durations and propulsion flicker cycles are also scaled by 1.20 so the visual motion stays synchronized with the calmer cue cadence.
- Combat zoom/framing and the 4 px one-click targeting drag tolerance remain unchanged from the preceding review fix.

## User-review work item - ship destruction explosion

- New review evidence shows that ship_destroyed has no dedicated ship-explosion presentation and that the original shared ship-explosion damage consumer is not yet present in MOOX production code.
- Review work is tracked in docs/research/SLICE_17_8_USER_REVIEW_SHIP_EXPLOSION_2026-10-07.md.
- The fix must preserve authoritative event/replay semantics, recover the original ordinary explosion-power/radius rules before implementation, consume the already-frozen Quantum Detonator x3 multiplier, and drive browser VFX only from authoritative events.
