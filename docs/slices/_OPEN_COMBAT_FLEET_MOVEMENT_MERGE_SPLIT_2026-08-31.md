# Open slice 04 - Combat Fleet movement, merge and split

Status: **Gates 1-3 complete; Gate 4 pending**.

Planned specification: `PLANNED_04_COMBAT_FLEET_MOVEMENT_MERGE_SPLIT.md`.

Permanent evidence: `docs/research/COMBAT_FLEET_MOVEMENT_MERGE_SPLIT_2026-08-31.md`.

## Completed gates

- [x] Gate 1: fresh repository/session check and original 1.31 evidence review.
- [x] Gate 2: schema-19 / command / event / replay contract accepted.
- [x] Gate 3: implementation and focused/full regression coverage complete.
- [ ] Gate 4: fresh QA, commits, documentation close and marker removal.

## Gate 3 implemented state

- Core `StateSchemaVersion` = **19**.
- Ordinary concrete combat Fleets use existing `AtSystemID` / `DestinationSystemID` / `RemainingTurns` semantic transit.
- Ordinary combat `FTLSpeed` remains zero/non-authoritative; movement speed derives from current Empire best Warp Drive plus Trans-Dimensional.
- Combat effective range derives through validated concrete Ship membership and currently resolves to current Empire best Fuel Cell range.
- Colony/Outpost supply origins are reused for combat movement legality.
- `empire.move_fleet` supports whole-Fleet movement and optional sorted/unique `ship_ids` subset movement.
- Proper subset movement is atomic: validate -> allocate one deterministic Fleet ID -> emit `empire.fleet_split` -> emit `empire.fleet_movement_started`.
- Rejected subset movement does not mutate composition or consume `NextID`.
- `empire.split_fleet` and `empire.merge_fleets` are implemented for owned stationary ordinary combat Fleets.
- In-transit split/merge/reroute remains rejected; Hyperspace Communications is deferred.
- Combat arrival materializes before blockade recomputation; hostile arrival may temporarily create stationary co-location/blockade until Slice 06.
- Schema-19 save/load, Observer `ShipIDs` isolation and identical-session deterministic state/event replay are covered.

## Gate 3 QA passed

```text
go test ./internal/core ./internal/game ./internal/session -run "(CombatFleet|Strategic|Military|ColonyShip|Outpost|Blockade|Observer|Save|Load)" -count=1
go test ./... -count=1
gofmt -w <all changed Go files>
git diff --check
```

All repository packages passed the final Gate-3 full test run.

## Current resume point

Do **not** begin Slice 05 and do **not** delete this marker yet. Resume Slice 04 at **Gate 4** with a fresh session/repository conflict check, then run final gofmt/full tests/`go vet`/focused regressions/diff checks, create the gameplay/data/research commit and closing-docs commit, update HISTORY/status, remove this marker and close the slice. No commit or push has been performed during Gates 1-3.
