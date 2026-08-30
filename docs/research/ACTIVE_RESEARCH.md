# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence.

## Recovery state

- Branch: `main`.
- Open slice marker: `docs/slices/_OPEN_COLONY_SHIP_MOVEMENT_COLONIZATION_2026-08-30.md`.
- Active slice: **Colony Ship production, strategic movement and colonization**.
- Current gate: **Gate 4 - QA + commit + close pending**.
- Starting HEAD for this slice: `5976f7e`.
- Core `StateSchemaVersion`: **16**.
- Economy ruleset schema: **7**.
- Permanent Gate 1 evidence: `docs/research/COLONY_SHIP_MOVEMENT_COLONIZATION_2026-08-30.md`.
- Previous closed evidence: `docs/research/TREASURY_MAINTENANCE_DEFICIT_2026-08-29.md`.
- Before starting new work, check `docs/slices/_OPEN_*.md`; resume this marker rather than opening a second slice.

## Gate 1 result

Direct MOO2 1.31 evidence establishes the minimum vertical loop:

- special class `1` is exactly Colony Ship (`2` Transport, `4` Outpost Ship);
- Technology 41 `colony_ship` unlocks it;
- base cost is 500 PP before the original government ship-cost reduction;
- the ship captures the best available Warp Drive at construction time;
- Fuel Cell range is Standard 4 pc, Deuterium 6 pc, Iridium 9 pc, Urridium 12 pc, Thorium 255 pc;
- ship completion creates a stationary strategic record at the producing Colony's System;
- transit has a destination and remaining ETA; original packed status/star-index/XY encoding need not be copied into MOOX;
- movement occurs before current-turn Production, so a ship built this turn cannot move this turn;
- colonization needs an owned Colony Ship at the System and a colonizable Planet;
- a normal new Colony starts with one assimilated founding Population unit;
- the Colony Ship is consumed when colonization succeeds.

## Proposed Gate 2 shape

If accepted:

1. Core schema 16 adds Colony Ship special identity and semantic Fleet transit fields while retaining `Role` for blockade compatibility.
2. A `colony_ship` Construction project is gated by Tech 41 and uses the observed ship-cost rule.
3. Colony Ship completion stores its FTL speed and starts stationary at the producing System.
4. Movement is Empire-authorized, Fuel-range-checked, deterministic and advanced before Construction.
5. Colonization is Empire-authorized and requires a stationary owned Colony Ship plus an empty same-System normal Planet.
6. Success creates/links a second Colony with one assimilated founding Population unit, recalculates it and consumes the ship.
7. Save/load, Observer and replay remain exact/deterministic.

Gate 2 was accepted without revision on 2026-08-30; Gate 3 is complete. The accepted Colony Ship build/move/colonize vertical loop is implemented and covered by Core/Game/Session regressions. Gate 4 QA + commit + close is pending.

## Deliberate deferrals

- generic combat-Fleet movement and tactical ship design/combat;
- Outpost Ship production and Outpost conversion;
- troop Transports/invasion;
- Extended Fuel Tanks/custom ship loadouts;
- Native/special founding-population cases;
- hostile-system engagement and AI colonization policy;
- broad diplomacy and special/non-player Fleet owners.

## Queue after this slice

Continue the headless vertical loop from the second Colony into the next independently evidenced strategic/economy dependency rather than broadening this slice opportunistically.