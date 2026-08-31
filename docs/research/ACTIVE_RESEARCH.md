# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: **none**.
- Latest completed gameplay slice: **Outpost Ship / Outpost state / supply range**.
- Implementation commit: `7c7284e` (`game: add outpost ship supply expansion`).
- Core `StateSchemaVersion`: **17**.
- Economy ruleset schema: **7**.
- Latest closed evidence: `docs/research/OUTPOST_SHIP_SUPPLY_RANGE_2026-08-31.md`.
- Before starting new work, check `docs/slices/_OPEN_*.md`; if none exists, Slice 03 is the next prepared objective and must begin with a fresh Gate 1 plus exactly one dated marker.

## Closed Outpost Ship / supply-range checkpoint

Slice 02 **Outpost Ship / Outpost state / supply range** is closed. Gate 4 passed and the runtime/tests/evidence are committed in `7c7284e`.

Closed runtime boundaries:

- Core schema 17 owns dedicated Planet-linked `Outpost` state rather than fake zero-Population Colonies;
- Outpost Ship is Technology 109, 100 PP base / 67 PP current Feudal, with installed best FTL speed persisted at completion;
- Colony Ship and Outpost Ship share the fixed-special-ship Fuel-range/transit path while generic military Fleet movement remains deferred;
- strategic supply origins are owned Colonies plus owned Outposts, and blockade alone does not suppress the range origin;
- `fleet.deploy_outpost` deploys on an unoccupied same-System Planet/body, consumes the ship and exposes the new supply origin immediately;
- Colony Ship and Colony Base founding atomically replace a same-owner Outpost; foreign Outpost conquest/destruction stays deferred;
- the unresolved original `star+0x37` control/UI gate remains deferred to the strategic hostile-encounter/control work rather than being invented here;
- schema-17 save/load, invalid-link validation, GameSession Observer isolation and deterministic repeated-session replay are covered.

Gate 4 passed fresh conflict/status inspection, `gofmt` verification, `go test ./... -count=1`, `go vet ./...`, focused Outpost/Colony Ship/Colony Base/Supply-range regressions and working/staged `git diff --check`.
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

No gameplay slice is currently open. Recommended next order:

- **Slice 03 - Military Ship core / design baseline** - `docs/slices/PLANNED_03_MILITARY_SHIP_CORE_DESIGN_BASELINE.md`
- **Slice 04 - Generic combat Fleet movement / merge / split** - `docs/slices/PLANNED_04_COMBAT_FLEET_MOVEMENT_MERGE_SPLIT.md`
- **Slice 05 - Command Points / ship overage Maintenance** - `docs/slices/PLANNED_05_COMMAND_POINTS_SHIP_MAINTENANCE.md`
- **Slice 06 - Strategic hostile encounters -> BattleSession handoff** - `docs/slices/PLANNED_06_STRATEGIC_HOSTILE_ENCOUNTERS_BATTLE_HANDOFF.md`
- **Slice 07 - Tactical ship combat baseline** - `docs/slices/PLANNED_07_TACTICAL_SHIP_COMBAT_BASELINE.md`

Start Slice 03 only after a fresh repository/session check and creation of exactly one dated `_OPEN_*.md` marker.
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