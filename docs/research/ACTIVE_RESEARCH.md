# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: `docs/slices/_OPEN_COLONY_BASE_SAME_SYSTEM_COLONIZATION_2026-08-31.md`.
- Latest completed gameplay slice: **Colony Ship production, strategic movement and colonization**.
- Implementation commit: `db9d01e` (`game: add colony ship colonization loop`).
- Core `StateSchemaVersion`: **16**.
- Economy ruleset schema: **7**.
- Active permanent evidence: `docs/research/COLONY_BASE_SAME_SYSTEM_COLONIZATION_2026-08-31.md`.
- Latest closed evidence: `docs/research/COLONY_SHIP_MOVEMENT_COLONIZATION_2026-08-30.md`.
- Before starting new work, check `docs/slices/_OPEN_*.md`; resume any marker before opening another slice.

## Active slice - Colony Base / same-system colonization

**Gate 3 is complete; Gate 4 is pending.** The accepted implementation is present in the working tree and has not yet been committed/closed.

Implemented runtime boundaries:

- Colony Base remains normal Building 11 / `colony_base`, Tech 40, 200 PP, 0 BC Maintenance;
- queue/building-choice legality requires at least one empty Planet in the constructing Colony's StarSystem;
- no target Planet is persisted with Construction;
- completed Bases are projected as derived `PendingColonyBaseResolutions` from existing Building ownership, so Core remains schema **16**;
- next-turn completion is blocked until every completed Base is resolved;
- `colony.colonize_with_base` creates a normal one-founder Colony on a legal same-system Planet, consumes the Base and leaves source Population unchanged;
- `colony.trash_colony_base` consumes the Base and refunds exactly 100 BC, including when no target remains;
- Colony Ship and Colony Base share the same normal-Colony founding helper;
- PlayerView, ObserverView and seat-scoped resolution projection expose authoritative pending decisions;
- dedicated Colony Base colonize/trash events are replay-stable.

Gate-3 verification is green: focused Colony Base/Colony Ship/Construction/Treasury/session regressions, exact schema-16 round-trip, deterministic replay, `go test ./... -count=1` and `go vet ./...`.

Next step is **Gate 4 only**: final format/diff QA, docs/HISTORY closure, implementation + closing-doc commits, then remove the `_OPEN_` marker. Do not start Slice 02 before this slice is closed.
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

Slice 01 **Colony Base / same-system colonization** is now open with Gate 1 complete and Gate 2 pending. The remaining six later slices stay prepared as planning specifications under `docs/slices/`.

Recommended order:

1. **Colony Base / same-system colonization** - `docs/slices/PLANNED_01_COLONY_BASE_SAME_SYSTEM_COLONIZATION.md`
2. **Outpost Ship / Outpost state / supply range** - `docs/slices/PLANNED_02_OUTPOST_SHIP_SUPPLY_RANGE.md`
3. **Military Ship core / design baseline** - `docs/slices/PLANNED_03_MILITARY_SHIP_CORE_DESIGN_BASELINE.md`
4. **Generic combat Fleet movement / merge / split** - `docs/slices/PLANNED_04_COMBAT_FLEET_MOVEMENT_MERGE_SPLIT.md`
5. **Command Points / ship overage Maintenance** - `docs/slices/PLANNED_05_COMMAND_POINTS_SHIP_MAINTENANCE.md`
6. **Strategic hostile encounters -> BattleSession handoff** - `docs/slices/PLANNED_06_STRATEGIC_HOSTILE_ENCOUNTERS_BATTLE_HANDOFF.md`
7. **Tactical ship combat baseline** - `docs/slices/PLANNED_07_TACTICAL_SHIP_COMBAT_BASELINE.md`

The queue files contain Gate 1-4 checklists, dependency notes, scope guards and exit criteria. Slice 01 is active through the single dated `_OPEN_*.md` marker; do not start slices 02-07 until the active slice is closed unless parallel work is explicitly intended.

The order is progressive: Colony Base first closes the currently incorrect generic-Building gap for same-system expansion; Outpost then extends special-ship/supply expansion; military Ship state supplies concrete Fleet composition; generic movement enables real strategic combat Fleets; Command Points can then use canonical ships; hostile arrivals can generate BattleSessions; only then does the first tactical-combat rules slice have all required producers/consumers.

Do not reopen the completed Colony Ship slice unless new evidence exposes a concrete fidelity defect.
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