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
