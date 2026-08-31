# OPEN - Military Ship core and design baseline - 2026-08-31

Status: **Gate 3 complete; Gate 4 pending**.

Planned specification: `docs/slices/PLANNED_03_MILITARY_SHIP_CORE_DESIGN_BASELINE.md`

Permanent evidence: `docs/research/MILITARY_SHIP_CORE_DESIGN_BASELINE_2026-08-31.md`

This is the single recovery marker for the active MOOX gameplay slice.

## Gate 1 - Checkup + original analysis / reverse engineering

- [x] Re-check repository state, active sessions, `_OPEN_*.md`, Core schema 17 and normalized hull identities.
- [x] Create exactly one active recovery marker.
- [x] Verify original military hull base space/cost for all six player hulls.
- [x] Verify original design-slot count and lifecycle.
- [x] Verify how built ship instances relate to mutable design slots/design replacement.
- [x] Verify mandatory/default drive, computer, armor, shield and other baseline design fields.
- [x] Resolve the smallest deterministic original Auto-Design/default-design fixture suitable for first construction.
- [x] Verify military ship construction cost path, government modifier and completion placement.
- [x] Resolve minimum persisted ship-instance fields required before generic Fleet movement/tactical combat.
- [x] Investigate per-hull size-point provenance; keep Command-Point accounting deferred because the final CP link/timing is not structurally required here.
- [x] Distinguish direct original evidence from MOOX semantic normalization/inference.
- [x] Update permanent evidence, README/live handoff and planned-slice state.
- [x] Remove temporary/private reverse-engineering artifacts.
- [x] Run working and staged `git diff --check` for Gate-1 documentation.
- [x] Present Gate-1 findings and proposed Gate-2 Core/API contract; stop before implementation.

## Gate 2 - Implementation decision

- [x] Accept Core schema 18 with concrete `ShipDesign`, `Ship` and Fleet `ShipIDs` composition.
- [x] Keep the original six design slots as evidence only; MOOX permits an arbitrary number of active designs per Empire with stable IDs and deterministic ordering.
- [x] Accept design revision + built-Ship snapshot semantics so later design edits cannot mutate existing Ships.
- [x] Accept normalization of all six hull base values and only the mandatory component subset required by this slice.
- [x] Accept a cleared/unarmed Frigate as the first supported deterministic military design fixture.
- [x] Accept full supported-design PP cost plus the established government ship-cost adjustment.
- [x] Accept `military_ship` Construction referencing design ID/revision and deterministic protection against unresolved active-queue design mutation.
- [x] Accept completion as a concrete Ship at the producing System in a one-Ship combat Fleet.
- [x] Keep generic combat-Fleet movement, merge/split, weapons, refit, Command-Point accounting and tactical combat deferred to their later slices.
- [x] Accept exact save/load, Observer and replay coverage as part of Gate 3.

## Gate 3 - Implementation

- [x] Advance Core to schema 18 with persisted unbounded `ShipDesigns`, concrete `Ships` and combat-Fleet `ShipIDs` composition.
- [x] Validate design/Ship identity, ownership, revisions, immutable built-Ship snapshots and one-Fleet-per-Ship membership without a six-design ceiling.
- [x] Extend normalized `ship_hulls.json` to schema 3 with original hull cost/space and mandatory Drive/Computer/Armor/Shield/Fuel tables generated from `Orion2.exe`.
- [x] Add authoritative cleared/unarmed Frigate design create/update command with legal picture validation and best known mandatory strategic systems.
- [x] Preserve the original-evidence baseline cost: Frigate 20 PP + Electronic Computer 5 PP = 25 PP; current Feudal production cost = 17 PP.
- [x] Expose every owned active design through authoritative Construction choices; no original six-slot cap is imposed.
- [x] Add `colony.queue_military_ship` with exact design ID/revision lock and reject revisions while that design is under Construction.
- [x] Run normal Production progress/cost to completion and create an independent concrete Ship snapshot at the producing System.
- [x] Create a deterministic one-Ship combat `StrategicFleet` containing the concrete Ship; generic combat-Fleet movement remains deferred.
- [x] Verify later current-design revision does not mutate an already built Ship snapshot.
- [x] Add Core exact schema-18 round-trip and >6-design validation tests.
- [x] Add Game cost/design/queue/completion/unsupported-scope regressions.
- [x] Add GameSession legal-action, Observer isolation, exact save/load and deterministic repeated-session replay coverage.
- [x] `go test ./internal/core ./internal/ruleset ./internal/moo2data ./internal/game ./internal/session -count=1` passed.
- [x] `go test ./... -count=1` passed.
- [x] Temporary Gate-3 reverse-engineering/editor artifacts removed.
- [x] Working-tree `git diff --check` passed.

## Gate 4 - QA + commit + close

Pending. Do not remove this marker or open Slice 04 before Gate 4 completes.

- [ ] Fresh repository/session conflict check.
- [ ] Final `gofmt` verification.
- [ ] `go test ./... -count=1`.
- [ ] `go vet ./...`.
- [ ] Focused military design/Construction/Fleet/save-load/Observer regressions plus Colony Ship/Outpost/blockade compatibility checks.
- [ ] Working-tree and staged `git diff --check`.
- [ ] Final documentation/HISTORY/project-status review.
- [ ] Gameplay implementation commit.
- [ ] Closing documentation commit deleting this `_OPEN_` marker.
