# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: `docs/slices/_OPEN_OUTPOST_SHIP_SUPPLY_RANGE_2026-08-31.md`.
- Latest completed gameplay slice: **Colony Base / same-system colonization**.
- Implementation commit: `261e418` (`game: add colony base colonization flow`).
- Core `StateSchemaVersion`: **17**.
- Economy ruleset schema: **7**.
- Active permanent evidence: `docs/research/OUTPOST_SHIP_SUPPLY_RANGE_2026-08-31.md`.
- Latest closed evidence: `docs/research/COLONY_BASE_SAME_SYSTEM_COLONIZATION_2026-08-31.md`.
- Before starting new work, check `docs/slices/_OPEN_*.md`; resume any marker before opening another slice.

## Active slice - Outpost Ship / Outpost state / supply range

**Gate 3 is complete; Gate 4 is pending.** The user accepted Gate 2 on 2026-08-31 and the accepted implementation contract is now present in the working tree. Do not open Slice 03 or remove the current `_OPEN_` marker before Gate 4.

Current runtime boundaries:

- Core `StateSchemaVersion` is **17** with dedicated `GameState.Outposts []Outpost` and reciprocal `Planet.OutpostID`; Outposts are not fake zero-Population `core.Colony` records;
- Core validation rejects invalid/dangling Outpost ownership/Planet references, multiple Outposts on one Planet, multiple Colonies on one Planet, missing reciprocal links and simultaneous Colony+Outpost occupancy;
- Outpost Ship is fixed civilian StrategicFleet special kind `outpost_ship`, Technology 109, 100 PP base / 67 PP current Feudal, with best available FTL speed persisted at completion;
- `colony.queue_outpost_ship` uses the normal single-project Construction path and materializes a stationary Outpost Ship at the producing System;
- Colony Ship and Outpost Ship share only the fixed-special-ship Fuel-range/transit path; generic military Fleet movement remains deferred;
- strategic supply origins are now owned Colonies **plus owned Outposts**; current blockade state does not suppress the direct range origin;
- `fleet.deploy_outpost` requires an owned stationary Outpost Ship and an unoccupied target Planet in that same System, consumes the ship, links a dedicated Outpost and exposes the new supply origin immediately;
- Colony Ship founding and Colony Base same-system founding atomically replace a same-owner Outpost on the target Planet; foreign Outposts remain illegal and hostile conquest/destruction stays deferred;
- the legacy original `star+0x37` control/UI gate remains deliberately unresolved for the later strategic-hostile-encounter slice rather than being represented by an invented diplomacy rule;
- schema-17 save/load, GameSession Observer clone isolation, full build -> move -> arrive -> deploy flow and identical-session deterministic replay are covered by dedicated regressions.

Gate-3 QA already passed:

```text
gofmt affected Go files
go test ./internal/core ./internal/game -count=1
go test ./internal/session -count=1
go test ./... -count=1
git diff --check
```

Gate 4 must still perform a fresh conflict/status check, final gofmt/test/vet/focused regressions, working/staged diff checks, documentation/HISTORY review, implementation commit and closing documentation commit that deletes the marker.
## Closed Colony Base checkpoint

Slice 01 **Colony Base / same-system colonization** is closed. Gate 4 passed and the runtime/tests/evidence are committed in `261e418`.

Closed runtime boundaries:

- Colony Base remains normal Building 11 / `colony_base`, Tech 40, 200 PP and 0 BC Maintenance;
- buildability and direct queue validation require at least one empty Planet in the source Colony's StarSystem;
- Construction persists no target Planet; completed Bases are derived from Building ownership;
- `colony.colonize_with_base` creates a normal same-system Colony, consumes the Base and leaves source Population unchanged;
- `colony.trash_colony_base` consumes the Base and refunds exactly 100 BC, including targetless completed Bases;
- `CompleteTurn` is blocked until all completed Colony Bases are resolved;
- Colony Ship and Colony Base share one deterministic normal-Colony founding helper;
- PlayerView/ObserverView expose authority-correct pending resolution state;
- Core schema remains **16**, with exact save/load reconstruction and deterministic replay coverage.

Gate 4 passed `gofmt`, `go test ./... -count=1`, `go vet ./...`, focused Colony Base/Colony Ship/Building-choice regressions and both working-tree/staged `git diff --check`.
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
- at that Colony-Ship checkpoint, Colony-owned Systems were the only authoritative supply origins; active Slice 02 now extends this to owned Outposts;
- Colony Ship transit before current-turn Construction, so a newly produced ship cannot also move that turn;
- explicit colonization only after arrival; no automatic same-step colonization;
- second-Colony Planet link, founding cohort/economy recalculation and Colony Ship consumption;
- exact Core save/load plus GameSession Observer/replay coverage.

Gate 4 passed `gofmt`, `go test ./... -count=1`, `go vet ./...`, focused Colony Ship/transit/blockade/Population-transfer regressions and `git diff --check`.

## Prepared next-slice queue

Slice 02 **Outpost Ship / Outpost state / supply range** is active with Gate 3 complete and Gate 4 pending. Slices 03-07 remain prepared planning specifications.

Recommended later order:

- **Slice 03 - Military Ship core / design baseline** - `docs/slices/PLANNED_03_MILITARY_SHIP_CORE_DESIGN_BASELINE.md`
- **Slice 04 - Generic combat Fleet movement / merge / split** - `docs/slices/PLANNED_04_COMBAT_FLEET_MOVEMENT_MERGE_SPLIT.md`
- **Slice 05 - Command Points / ship overage Maintenance** - `docs/slices/PLANNED_05_COMMAND_POINTS_SHIP_MAINTENANCE.md`
- **Slice 06 - Strategic hostile encounters -> BattleSession handoff** - `docs/slices/PLANNED_06_STRATEGIC_HOSTILE_ENCOUNTERS_BATTLE_HANDOFF.md`
- **Slice 07 - Tactical ship combat baseline** - `docs/slices/PLANNED_07_TACTICAL_SHIP_COMBAT_BASELINE.md`

Do not start slices 03-07 while Slice 02 remains active unless parallel work is explicitly intended.
## Later deferred dependencies

- generic combat-Fleet movement/orders and tactical ship composition;
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