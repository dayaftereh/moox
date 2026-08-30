# Open slice: Colony Ship production, strategic movement and colonization

Status: Gate 3 complete on 2026-08-30; Gate 4 QA + commit + close pending.

## Objective

Resolve and then implement the minimum original Master of Orion II 1.31 semantics required for an authoritative headless vertical loop from Colony Ship production through strategic movement to colonizing a legal unowned planet and creating a second Colony.

## Recovery

- Starting HEAD: `5976f7e` (`docs: close treasury maintenance slice`).
- Starting branch/tree: `main`, clean, ahead of `origin/main` by 16 commits.
- Core schema: 15.
- Economy ruleset schema: 7.
- Permanent Gate 1 evidence: `docs/research/COLONY_SHIP_MOVEMENT_COLONIZATION_2026-08-30.md`.

## Gates

- [x] Gate 1: repository checkup + original MOO2 1.31 reverse engineering
- [x] Gate 2: implementation decision
- [x] Gate 3: implementation
- [ ] Gate 4: QA + commit + close

## Gate 1 result

Direct original evidence now establishes:

- exact special-ship classes: `1=Colony Ship`, `2=Transport`, `4=Outpost Ship`;
- Colony Ship requires Technology 41 `colony_ship`;
- base Colony Ship cost is 500 PP before the directly observed government ship-cost reduction;
- the fixed Colony Ship design installs the best drive available when it is built;
- original Fuel Cell base ranges are 4/6/9/12/255 parsecs for Standard/Deuterium/Iridium/Urridium/Thorium;
- Construction completion creates the ship stationary at the producing Colony's System;
- original transit persists destination/movement state and a discrete remaining ETA; MOOX can replace the packed status/index/XY representation with semantic destination + remaining-turn fields;
- strategic ship movement happens before current-turn Colony Production, so a newly completed Colony Ship cannot move in the same resolution;
- a legal colonization requires an owned Colony Ship at the System plus a colonizable Planet;
- original normal Planets may be empty or contain an Outpost; current MOOX should initially implement only empty normal Planets because Outpost state does not yet exist;
- a normal newly founded Colony starts with one assimilated founding Population unit and no invented free Building;
- successful colonization consumes the Colony Ship.

## Proposed Gate 2 contract

- Core schema **15 -> 16**; Economy ruleset schema remains **7**.
- Add semantic Construction kind `colony_ship`, gated by Technology 41, base 500 PP plus relevant currently modeled government ship-cost reduction.
- Extend `StrategicFleet` with a special kind plus semantic transit fields (`DestinationSystemID`, `RemainingTurns`, persisted `FTLSpeed`) while preserving existing combat/civilian `Role`.
- Add original Fuel Cell range helper and reject out-of-range movement orders.
- Add an Empire-authorized Colony-Ship movement command; movement advances before Construction.
- Add an Empire-authorized colonize command requiring a stationary owned Colony Ship and an empty same-System Planet.
- Colonization atomically creates/links the second Colony with one assimilated founding Population unit, recalculates its authoritative state, and removes the consumed Colony Ship.
- Preserve exact save/load, Observer clone isolation and deterministic replay/events.
- Do not queue auto-colonize on arrival; a ship arriving during movement becomes colonizable in the next command phase.

## Scope guard / deferrals

Do not absorb tactical ship design/combat, generic combat-Fleet movement, Outpost production/conversion, Transports/invasion, Native special population, Extended Fuel Tanks/custom ship range, hostile engagement, AI colonization policy or broad diplomacy into this slice.

Gate 2 accepted the proposed contract without revision on 2026-08-30. Gate 3 is implementing only that accepted scope.
## Gate 3 implementation result

Accepted scope is implemented without broadening into deferred systems.

- Core schema advanced from 15 to **16**; Economy ruleset schema remains **7**.
- `ConstructionProjectColonyShip` / `colony.queue_colony_ship` uses Technology 41 and the accepted 500 PP / current Feudal 334 PP rule.
- Completion creates a civilian `StrategicFleet` with `SpecialKind=colony_ship`, producing-System presence and persisted installed `FTLSpeed`.
- `StrategicFleet` now carries semantic `DestinationSystemID`, `RemainingTurns` and `FTLSpeed`; Core validation enforces stationary/transit exclusivity and Colony-Ship-only semantic transit for schema 16.
- Original Fuel Cell ranges are authoritative: Standard 4 pc, Deuterium 6, Iridium 9, Urridium 12, Thorium 255. Range is checked against the nearest own Colony supply System; Outpost supply remains deferred because Outposts are not authoritative yet.
- `empire.move_fleet` is currently legal only for owned stationary Colony Ships and advances deterministic transit before current-turn Construction.
- `empire.colonize_planet` requires an owned stationary Colony Ship and an empty normal Planet in the same System; success creates/links a second Colony with one assimilated founding Farmer cohort, recalculates it and consumes the Colony Ship.
- No automatic same-step colonization occurs on arrival.
- Population-transfer ETA now reuses the same strategic coordinate-to-parsec helper while retaining its independently proven 15-turn cap.
- Core roundtrip/validation, Game legality/timing/range/colonization, full headless vertical loop and GameSession Observer clone/history regressions are present.
- `go test ./... -count=1` passes after the new regressions.

Gate 4 still owns final gofmt/vet/diff QA, implementation commit, closing documentation and marker removal.