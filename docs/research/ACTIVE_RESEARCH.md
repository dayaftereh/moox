# Active research / implementation handoff

This file is the authoritative **live** handoff for the current Master of Orion X implementation sequence. Permanent reverse-engineering evidence and closed results live in `docs/research/`; the completed-slice ledger lives in `docs/slices/HISTORY.md`.

## Recovery state

- Branch: `main`.
- Open slice marker: none.
- Latest completed gameplay slice: **Command Points / ship Maintenance**.
- Implementation commit: `16afe0b` (`game: add command point maintenance`).
- Core `StateSchemaVersion`: **20**.
- Economy ruleset schema: **8**.
- Active permanent evidence: none.
- Latest closed evidence: `docs/research/COMMAND_POINTS_SHIP_MAINTENANCE_2026-08-31.md`.
- Before starting new work, check `docs/slices/_OPEN_*.md`; no gameplay slice is open. Slice 06 **Strategic hostile encounters / BattleSession handoff** is the next prepared objective.

## Closed Slice 05 - Command Points and ship Maintenance

**Gates 1-4 are complete.** Gate 4 passed fresh gofmt, focused original-accounting/compatibility regressions, full tests, vet and diff checks; implementation commit `16afe0b` is recorded and the recovery marker is removed.

Permanent evidence: `docs/research/COMMAND_POINTS_SHIP_MAINTENANCE_2026-08-31.md`.

Gate-1 conclusions:

- military Ship CP usage is authoritative hull `SizeIndex + 1` (Frigate 1 through Doom Star 6);
- fixed Colony/Outpost special ships consume 1 CP each, including while in normal transit; Population-transfer Settlers/Freighters are not original troop Transports and do not acquire fake CP usage;
- original generated capacity starts at 5, adds Star Base/Battlestation/Star Fortress +1/+2/+3, strongest known Communications +1/+2/+3 per station, Warlord +2 per owned Colony, then Imperium adds `floor(capacity/2)`;
- an original officer-derived capacity source exists but remains zero/deferred because MOOX has no Leader subsystem;
- player-facing overage Maintenance is exactly 10 BC per excess CP; the separate original NPC/difficulty `12 - global setting` branch remains deferred;
- original scrap/dead cleanup precedes Maintenance and current-turn ship Production completes after Maintenance, so a newly completed Ship first affects CP Maintenance on the following settlement;
- Gate 1 found that pre-Slice MOOX permitted multiple orbital station tiers on one Colony; Gate 3 now enforces the original Star Base -> Battlestation -> Star Fortress replacement/buildability relationship;
- implementation uses Core schema 20 + materialized Empire `{capacity, used}`, Treasury ship-command Maintenance, economy ruleset schema 8 Command-Point constants/producers, no ship-hulls schema bump, and no new player command.

Gate 4 re-ran the complete QA surface successfully, including `go vet ./...`, reviewed accounting/schema/station/Treasury diffs, and committed gameplay/evidence as `16afe0b`. Slice 05 is closed; Slice 06 is next.
## Closed Slice 04 - Combat Fleet movement, merge and split

**Gates 1-4 are complete.** Final Gate 4 gofmt, full tests, vet, focused compatibility regressions and diff checks passed; the implementation is committed and the recovery marker is removed.

Permanent evidence: `docs/research/COMBAT_FLEET_MOVEMENT_MERGE_SPLIT_2026-08-31.md`.

Gate-1 conclusions:

- original strategic movement operates on an arbitrary selected subset of co-located ships; MOOX must map that to explicit stable Fleet containers;
- normal in-hyperspace destination changes are technology-gated by Hyperspace Communications, so this slice can reject in-transit split/merge/reroute;
- current ordinary military movement should derive effective FTL speed from the Empire's current best Warp Drive plus Trans-Dimensional, rather than from stale build-snapshot speeds;
- current Fleet range similarly uses the Empire's best Fuel Cell and the already-proven Colony/Outpost supply origins; Extended Fuel Tanks remains a future per-Ship modifier boundary;
- existing `AtSystemID` / `DestinationSystemID` / `RemainingTurns` state can be generalized directly; no second transit model is needed;
- original selected-subset movement maps cleanly to `move_fleet` with optional `ship_ids` for atomic split+move, plus explicit stationary split/merge commands;
- Fleet movement/arrival precedes blockade recomputation, so departure removes and arrival can establish blockade in the same resolution;
- hostile arrival may temporarily materialize stationary co-location/blockade; Slice 06 owns encounter/BattleSession handoff.

Gate 2 accepted the schema-19/API/event contract and Gate 3 now implements it.

Gate-3 runtime result:

- Core schema 19 accepts ordinary concrete combat-Fleet transit and rejects empty/locationless/cached-FTL combat Fleet state;
- combat movement derives current Empire Warp Drive + Trans-Dimensional speed and current Fuel Cell range while validating every concrete member Ship;
- `empire.move_fleet` supports deterministic whole-Fleet or selected-subset movement; proper subsets allocate one new Fleet ID and emit `empire.fleet_split` before `empire.fleet_movement_started`;
- `empire.split_fleet` and `empire.merge_fleets` operate only on owned stationary ordinary combat Fleets and preserve sorted unique non-empty Ship composition;
- rejected subset movement validates before mutation and does not consume `NextID`;
- combat transit uses the existing destination/remaining-turn countdown, arrival materializes before blockade recomputation, and no in-transit reroute/split/merge is introduced;
- schema-19 save/load, Observer `ShipIDs` isolation, identical-session state/event replay, Colony/Outpost compatibility and full repository tests are green.

Gate 4 closure: fresh repository/session check found no foreign writer and exactly one expected OPEN marker. `gofmt` on changed Go files, `go test ./... -count=1`, `go vet ./...`, focused CombatFleet/Strategic/Military/ColonyShip/Outpost/Blockade/Observer/Save/Load/Replay regressions and `git diff --check` all passed. Gameplay/evidence commit: `f297480`. The OPEN marker is removed in the closing documentation commit. No push was performed. Next prepared objective: Slice 05 **Command Points / ship maintenance**.
## Closed Slice 03 - Military Ship core / design baseline

**Gates 1-4 are complete.** Final Gate 4 gofmt, full tests, vet, focused compatibility regressions and diff checks passed; the implementation is committed and the recovery marker is removed.

Current runtime boundaries:

- Core `StateSchemaVersion` is **18** with persisted `ShipDesigns`, concrete `Ships` and combat-Fleet `ShipIDs`;
- MOOX intentionally permits an arbitrary number of active designs per Empire; the original six `0x63` slots remain evidence only, with dedicated >6-design validation/round-trip coverage;
- current designs use stable IDs and monotonic revisions, while built Ships persist an independent source-design snapshot so later current-design edits cannot mutate existing Ships;
- `ship_hulls.json` schema 3 now normalizes the six direct original hull cost/space pairs plus the supported mandatory Drive/Computer/Armor/Shield/Fuel tables from `Orion2.exe`;
- the first supported authoritative design surface is a cleared/unarmed Frigate with a legal original Frigate picture and the best known supported mandatory strategic systems;
- the exact first baseline is 20 PP Frigate hull + 5 PP Electronic Computer = 25 PP, with current Feudal production cost 17 PP;
- every owned active design is surfaced as a distinct authoritative `military_ship` Construction choice; `colony.queue_military_ship` snapshots exact design ID/revision and design revision is blocked while that design is under Construction;
- normal Production completion creates an independent concrete Ship and a new stationary one-Ship combat Fleet at the producing Colony's System;
- generic combat-Fleet movement/merge/split remains Slice 04; Command Points/ship-overage Maintenance remains Slice 05; hostile encounter handoff and tactical weapons/combat remain Slices 06-07;
- detailed component miniaturization beyond the first exact baseline remains deferred rather than being falsely approximated as fully original-equivalent;
- Core save/load, GameSession legal actions, Observer clone isolation and identical-session deterministic replay are covered.

Gate-3 QA already passed:

```text
go test ./internal/game -run Military -count=1
go test ./internal/session -run Military -count=1
go test ./internal/core ./internal/ruleset ./internal/moo2data ./internal/game ./internal/session -count=1
go test ./... -count=1
git diff --check
```

All temporary Gate-3 reverse-engineering/editor artifacts were removed. Gate 4 then passed the fresh session/repository conflict check, final gofmt/test/vet/focused compatibility regressions and working/staged diff checks; implementation commit `2670d36` is recorded and the closing documentation removes the recovery marker.
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

Slice 03 **Military Ship core / design baseline** is closed. Slices 04-07 remain prepared planning specifications, with Slice 04 next.

Recommended later order:

- **Slice 04 - Generic combat Fleet movement / merge / split** - `docs/slices/PLANNED_04_COMBAT_FLEET_MOVEMENT_MERGE_SPLIT.md`
- **Slice 05 - Command Points / ship overage Maintenance** - `docs/slices/PLANNED_05_COMMAND_POINTS_SHIP_MAINTENANCE.md`
- **Slice 06 - Strategic hostile encounters -> BattleSession handoff** - `docs/slices/PLANNED_06_STRATEGIC_HOSTILE_ENCOUNTERS_BATTLE_HANDOFF.md`
- **Slice 07 - Tactical ship combat baseline** - `docs/slices/PLANNED_07_TACTICAL_SHIP_COMBAT_BASELINE.md`

No slice is currently open. Start Slice 04 only with a fresh repository/session check and a new dated `_OPEN_` marker.
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