# OPEN - Outpost Ship, Outpost state and supply range - 2026-08-31

Status: **Gate 3 complete; Gate 4 pending**.

Planned specification: `docs/slices/PLANNED_02_OUTPOST_SHIP_SUPPLY_RANGE.md`

Permanent evidence: `docs/research/OUTPOST_SHIP_SUPPLY_RANGE_2026-08-31.md`

This is the single recovery marker for the active MOOX gameplay slice.

## Gate 1 - Checkup + original analysis / reverse engineering

- [x] Re-check Git status, `_OPEN_*.md`, Core schema and current strategic movement implementation.
- [x] Open exactly one active slice marker.
- [x] Reconfirm direct baseline evidence: Outpost Ship special class 4, Technology 109 and base cost 100 PP.
- [x] Verify government-adjusted production cost and installed-drive semantics.
- [x] Verify Outpost Ship strategic Fuel/supply-range rules and whether existing Outposts are supply origins.
- [x] Resolve original deployment legality and whether deployment targets a System or Planet.
- [x] Resolve authoritative Outpost representation and ownership state.
- [x] Verify when supply extension becomes active.
- [x] Verify Outpost-to-Colony replacement/conversion semantics.
- [x] Verify minimum blockade/hostility implications needed for this slice.
- [x] Update permanent evidence and `ACTIVE_RESEARCH.md` with Gate-1-complete findings.
- [x] Remove temporary/private executable extraction helpers.
- [x] Run `git diff --check` over intended Gate-1 docs.
- [x] Present Gate-1 findings and proposed Gate-2 contract; stop before gameplay implementation.

## Gate 2 - Implementation decision

Accepted by the user on 2026-08-31.

- [x] Use dedicated persisted `Outpost` state rather than fake zero-Population `core.Colony` records.
- [x] Advance Core to schema 17 with `GameState.Outposts` and reciprocal `Planet.OutpostID`.
- [x] Add fixed `outpost_ship` StrategicFleet identity and reuse/generalize only Colony/Outpost special-ship transit.
- [x] Add Tech-109 Outpost Ship Construction at 100 PP base / 67 PP current Feudal.
- [x] Persist installed best FTL speed at Outpost Ship completion.
- [x] Generalize supply origins to owned Colonies plus owned Outposts without blockade suppression.
- [x] Add Planet-level `fleet.deploy_outpost` and consume the successful Outpost Ship.
- [x] Support same-owner Outpost-to-Colony replacement through Colony Ship and Colony Base founding.
- [x] Keep legacy `star+0x37` hostile/control semantics deferred rather than inventing a diplomacy rule.
- [x] Require save/load, Observer isolation, deterministic replay and vertical lifecycle regressions.

## Gate 3 - Implementation

- [x] Core schema 17 Outpost entity, Planet link and reciprocal/occupancy validation.
- [x] Outpost Ship construction choice, queue command, completion event and installed-drive Fleet materialization.
- [x] Fixed Colony/Outpost special-ship Fuel-range movement and transit generalization.
- [x] Colony + Outpost authoritative supply-origin lookup.
- [x] Planet-level Outpost deployment, ID allocation, ship consumption and immediate supply effect.
- [x] Same-owner Outpost conversion/removal during Colony Ship founding.
- [x] Same-owner Outpost conversion/removal during Colony Base founding; foreign Outposts remain illegal targets.
- [x] Core schema-17 exact round-trip and invalid-link tests.
- [x] Focused Game tests for 100/67 PP cost, Tech 109, queue/completion, movement, deployment, blockaded supply and both conversion paths.
- [x] GameSession build -> move -> arrive -> deploy Observer flow, clone isolation, save/load and deterministic repeated-session replay.
- [x] `gofmt` affected Go files.
- [x] `go test ./internal/core ./internal/game -count=1` passed.
- [x] `go test ./internal/session -count=1` passed.
- [x] `go test ./... -count=1` passed.
- [x] working-tree `git diff --check` passed.
- [x] Temporary Gate-3 editor tooling removed.

## Gate 4 - QA + commit + close

Pending. Do not remove this marker or open Slice 03 before Gate 4 completes.

- [ ] Fresh repository/session conflict check.
- [ ] Final `gofmt` verification.
- [ ] `go test ./... -count=1`.
- [ ] `go vet ./...`.
- [ ] Focused Outpost/Colony Ship/Colony Base/strategic-range regressions.
- [ ] working-tree and staged `git diff --check`.
- [ ] Final documentation/HISTORY/project-status review.
- [ ] Gameplay implementation commit.
- [ ] Closing documentation commit deleting this `_OPEN_` marker.
