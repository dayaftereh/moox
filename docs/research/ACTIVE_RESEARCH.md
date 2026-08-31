# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: **none**.
- Latest completed gameplay slice: **Colony Ship production, strategic movement and colonization**.
- Implementation commit: `db9d01e` (`game: add colony ship colonization loop`).
- Core `StateSchemaVersion`: **16**.
- Economy ruleset schema: **7**.
- Permanent evidence: `docs/research/COLONY_SHIP_MOVEMENT_COLONIZATION_2026-08-30.md`.
- Previous closed evidence: `docs/research/TREASURY_MAINTENANCE_DEFICIT_2026-08-29.md`.
- Before starting new work, check `docs/slices/_OPEN_*.md`; resume any marker before opening another slice.

## Closed Colony Ship checkpoint

The first authoritative headless expansion loop is complete:

```text
home Colony
-> construct Tech-41 Colony Ship
-> create stationary civilian Colony Ship at producing System
-> issue Fuel-range-checked strategic move
-> advance deterministic ETA before current-turn Construction
-> arrive at target System
-> explicitly colonize an empty same-System Planet
-> consume Colony Ship
-> create/link second Colony with one assimilated founding Population unit
```

Direct MOO2 1.31 evidence and the accepted Gate-2 contract are preserved in `docs/research/COLONY_SHIP_MOVEMENT_COLONIZATION_2026-08-30.md`.

Runtime boundaries now include:

- base Colony Ship cost 500 PP, with the current Feudal observed-rule cost of 334 PP;
- original Fuel Cell strategic ranges: Standard 4 pc, Deuterium 6, Iridium 9, Urridium 12 and Thorium 255;
- persisted Colony Ship special identity, installed FTL speed, destination and ETA in Core schema 16;
- Colony-owned Systems as the currently authoritative supply origins; Outpost supply remains deferred until Outposts exist;
- Colony Ship transit before current-turn Construction, so a newly produced ship cannot also move that turn;
- explicit colonization only after arrival; no automatic same-step colonization;
- second-Colony Planet link, founding cohort/economy recalculation and Colony Ship consumption;
- exact Core save/load plus GameSession Observer/replay coverage.

Gate 4 passed `gofmt`, `go test ./... -count=1`, `go vet ./...`, focused Colony Ship/transit/blockade/Population-transfer regressions and `git diff --check`.

## Prepared next-slice queue

No gameplay slice is currently open. All previously started slices are closed, and the next six are prepared as planning specifications under `docs/slices/`.

Recommended order:

1. **Outpost Ship / Outpost state / supply range**
   `docs/slices/PLANNED_01_OUTPOST_SHIP_SUPPLY_RANGE.md`
2. **Military Ship core / design baseline**
   `docs/slices/PLANNED_02_MILITARY_SHIP_CORE_DESIGN_BASELINE.md`
3. **Generic combat Fleet movement / merge / split**
   `docs/slices/PLANNED_03_COMBAT_FLEET_MOVEMENT_MERGE_SPLIT.md`
4. **Command Points / ship overage Maintenance**
   `docs/slices/PLANNED_04_COMMAND_POINTS_SHIP_MAINTENANCE.md`
5. **Strategic hostile encounters -> BattleSession handoff**
   `docs/slices/PLANNED_05_STRATEGIC_HOSTILE_ENCOUNTERS_BATTLE_HANDOFF.md`
6. **Tactical ship combat baseline**
   `docs/slices/PLANNED_06_TACTICAL_SHIP_COMBAT_BASELINE.md`

The files above already contain Gate 1-4 checklists, dependency notes, scope guards and exit criteria, but **none is active**. Starting one requires a fresh Gate 1 and exactly one dated `_OPEN_*.md` marker. Do not reopen the completed Colony Ship slice unless new evidence exposes a concrete fidelity defect.

The order is deliberately progressive: Outpost extends the current special-ship/supply model; military Ship state then supplies concrete Fleet composition; generic movement enables real strategic combat Fleets; Command Points can then use canonical ships; hostile arrivals can generate BattleSessions; only then does the first tactical-combat rules slice have all required producers/consumers.

## Later deferred dependencies

- generic combat-Fleet movement/orders and tactical ship composition;
- Outpost Ship production, Outpost supply and Outpost->Colony conversion;
- Transport/troop movement and invasion;
- hostile-system engagement before colonization;
- Ship command-point capacity/usage and overage Maintenance;
- Spy state/Maintenance;
- trade/research/Tribute treaty economics;
- Officer/Leader state, Maintenance and economic bonuses;
- complete original staged Treasury-deficit liquidation;
- active conquest/occupation/automatic-assimilation progression;
- Android/Native population and persisted custom race designs;
- AI colony targeting/auto-colonize policy and broader diplomacy;
- special/non-player Fleet owners.

Original copyrighted assets remain private reference material and are not distributable MOOX content. Wails/network/MCP layers remain adapters and must not own gameplay legality or deterministic state transitions.