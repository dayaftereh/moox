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
