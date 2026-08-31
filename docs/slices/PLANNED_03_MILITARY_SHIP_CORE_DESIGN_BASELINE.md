# Planned slice 03 - Military Ship core and design baseline

Status: **closed; Gates 1-4 complete**.

Queue position: **3 of 7**.

This file records the completed Slice 03 specification. Recovery marker removed at Gate 4 closure.

## Objective

Introduce the minimum canonical military Ship/Design state required to stop treating a combat Fleet as an abstract blockade marker.

The target is a headless strategic baseline in which an Empire can own a persisted military design, construct a concrete ship from it, and place that ship into authoritative strategic fleet composition. Detailed tactical weapon behavior is deliberately outside this slice.

## Existing baseline

- `data/rulesets/moo2-1.31/ship_hulls.json` normalizes the six player hull identities: Frigate, Destroyer, Cruiser, Battleship, Titan and Doom Star.
- Strategic/tactical ship artwork mappings and original Auto-Design picture logic already exist in the data/analyzer layer.
- Core schema 17 has strategic Fleet identity/role, fixed Colony/Outpost Ship transit and Planet-linked Outposts, but no canonical military ship instances or design composition.
- `internal/battle` provides deterministic battle-session infrastructure but not ship combat rules.

## Dependencies

- planned slice 02 should be closed first unless explicitly reprioritized;
- normalized hull identities and technology ownership;
- semantic Construction model;
- current strategic Fleet owner/location model.

## Scope guard

In scope:

- minimum original military hull/design identity;
- persisted design/ship ownership;
- authoritative design-slot semantics if required for normal construction;
- buildability and production cost inputs needed to construct a military ship;
- installed drive/basic mandatory component state required for later strategic movement;
- strategic Fleet composition containing concrete ship instances.

Defer:

- full beam/missile/bomb/fighter behavior;
- tactical hit/damage;
- refits unless Gate 1 proves unavoidable for design identity;
- detailed miniaturization beyond what is required to validate the first design;
- fleet movement (slice 04);
- Command-Point accounting (slice 05), except storing evidenced per-ship/hull values if dependency-safe.

## Gate 1 - Checkup + original analysis / reverse engineering

- [x] Re-check repo state, schema 17 and normalized hull data.
- [x] Create a permanent evidence document.
- [x] Verify original hull base space/cost and any construction-cost boundary required by the first buildable military design.
- [x] Verify original design-slot count/lifecycle and how built ships reference designs.
- [x] Verify mandatory/default drive, computer, armor and other fields that must exist even before tactical combat is modeled.
- [x] Verify Auto-Design/default-design behavior only as needed to create a deterministic legal initial design.
- [x] Verify military ship construction government-cost adjustment and completion placement.
- [x] Resolve whether a built ship copies design values or references a mutable design slot, especially across later refits/design replacement.
- [x] Resolve the minimum original ship-instance fields needed for strategic ownership and future fleet movement.
- [x] Investigate Command-Point value provenance but defer accounting to slice 05 unless it is structurally inseparable from hull state.
- [x] Distinguish direct evidence from inferred/modernized semantic state.
- [x] Present Gate-1 findings and proposed minimal Core/API shape; stop before implementation.

## Gate 2 - Implementation decision

- [x] Decide Core `ShipDesign` identity and immutable/mutable fields; stable ID/revision with no fixed MOOX design-count ceiling.
- [x] Decide Core military ship-instance identity and relation to designs; built Ship snapshots are independent of later design revisions.
- [x] Decide how `StrategicFleet` owns/references concrete ship instances via deterministic `ShipIDs`.
- [x] Decide schema migration/version bump: Core schema 18.
- [x] Decide first supported legal military design surface: cleared/unarmed Frigate with mandatory strategic systems; weapons/specials remain deferred.
- [x] Decide Construction project identity and queue command/event surface: `military_ship` referencing design ID/revision.
- [x] Decide technology/buildability validation ownership in authoritative game/rules logic; first vertical surface remains Frigate.
- [x] Decide Observer/replay/save-load contract with exact design/Ship/Fleet state round-trip.
- [x] Record deliberate MOOX semantic normalization: arbitrary active design count instead of the original six-slot storage ceiling, stable IDs/revisions, and one-Ship Fleet completion boundary.

## Gate 3 - Implementation

- [x] Add normalized runtime hull parameters proven by Gate 1 if not already represented.
- [x] Add Core design and ship-instance state plus validation.
- [x] Add deterministic legal design creation/selection for the accepted minimal design scope.
- [x] Add military ship Construction queue/progress/completion.
- [x] Spawn a concrete military ship at the producing system and attach it to authoritative fleet composition.
- [x] Preserve installed drive/basic required design state on the built ship.
- [x] Add Observer/replay events and exact save/load coverage.
- [x] Add deterministic construction/design validation tests.
- [x] Add a vertical test from legal design -> queue -> completion -> concrete strategic ship/fleet.

## Gate 4 - Follow-up QA + commit + close

- [x] `gofmt` changed Go files.
- [x] `go test ./... -count=1`.
- [x] `go vet ./...`.
- [x] `git diff --check`.
- [x] Run focused ship-design/construction/Fleet/save-load/Observer regressions.
- [x] Verify existing Colony Ship/Outpost and blockade behavior remains intact.
- [x] Update permanent evidence/status/HISTORY.
- [x] Commit implementation, then closing docs; remove `_OPEN_` marker.

## Exit criterion

A deterministic persisted military design can be constructed through the normal strategic Construction path into a concrete owned ship represented in strategic Fleet composition, with exact save/load and Observer/replay semantics and without yet requiring tactical combat.