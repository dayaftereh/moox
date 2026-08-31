# Planned slice 02 - Military Ship core and design baseline

Status: **planned / queued; not open**.

Queue position: **2 of 6**.

This file is a prepared slice specification, not an active `_OPEN_*.md` marker.

## Objective

Introduce the minimum canonical military Ship/Design state required to stop treating a combat Fleet as an abstract blockade marker.

The target is a headless strategic baseline in which an Empire can own a persisted military design, construct a concrete ship from it, and place that ship into authoritative strategic fleet composition. Detailed tactical weapon behavior is deliberately outside this slice.

## Existing baseline

- `data/rulesets/moo2-1.31/ship_hulls.json` normalizes the six player hull identities: Frigate, Destroyer, Cruiser, Battleship, Titan and Doom Star.
- Strategic/tactical ship artwork mappings and original Auto-Design picture logic already exist in the data/analyzer layer.
- Core schema 16 has strategic Fleet identity/role and special Colony Ship transit, but no canonical military ship instances or design composition.
- `internal/battle` provides deterministic battle-session infrastructure but not ship combat rules.

## Dependencies

- planned slice 01 should be closed first unless explicitly reprioritized;
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
- fleet movement (slice 03);
- Command-Point accounting (slice 04), except storing evidenced per-ship/hull values if dependency-safe.

## Gate 1 - Checkup + original analysis / reverse engineering

- [ ] Re-check repo state, schema 16 and normalized hull data.
- [ ] Create a permanent evidence document.
- [ ] Verify original hull base space/cost and any construction-cost boundary required by the first buildable military design.
- [ ] Verify original design-slot count/lifecycle and how built ships reference designs.
- [ ] Verify mandatory/default drive, computer, armor and other fields that must exist even before tactical combat is modeled.
- [ ] Verify Auto-Design/default-design behavior only as needed to create a deterministic legal initial design.
- [ ] Verify military ship construction government-cost adjustment and completion placement.
- [ ] Resolve whether a built ship copies design values or references a mutable design slot, especially across later refits/design replacement.
- [ ] Resolve the minimum original ship-instance fields needed for strategic ownership and future fleet movement.
- [ ] Investigate Command-Point value provenance but defer accounting to slice 04 unless it is structurally inseparable from hull state.
- [ ] Distinguish direct evidence from inferred/modernized semantic state.
- [ ] Present Gate-1 findings and proposed minimal Core/API shape; stop before implementation.

## Gate 2 - Implementation decision

- [ ] Decide Core `ShipDesign` identity and immutable/mutable fields.
- [ ] Decide Core military ship-instance identity and relation to design slots.
- [ ] Decide how `StrategicFleet` owns/references concrete ship instances.
- [ ] Decide schema migration/version bump.
- [ ] Decide first supported legal military design surface and what remains opaque/deferred.
- [ ] Decide Construction project identity and queue command/event surface.
- [ ] Decide technology/buildability validation ownership.
- [ ] Decide Observer/replay/save-load contract.
- [ ] Record deliberate MOOX semantic normalization where original packed design structures are not copied byte-for-byte.

## Gate 3 - Implementation

- [ ] Add normalized runtime hull parameters proven by Gate 1 if not already represented.
- [ ] Add Core design and ship-instance state plus validation.
- [ ] Add deterministic legal design creation/selection for the accepted minimal design scope.
- [ ] Add military ship Construction queue/progress/completion.
- [ ] Spawn a concrete military ship at the producing system and attach it to authoritative fleet composition.
- [ ] Preserve installed drive/basic required design state on the built ship.
- [ ] Add Observer/replay events and exact save/load coverage.
- [ ] Add deterministic construction/design validation tests.
- [ ] Add a vertical test from legal design -> queue -> completion -> concrete strategic ship/fleet.

## Gate 4 - Follow-up QA + commit + close

- [ ] `gofmt` changed Go files.
- [ ] `go test ./... -count=1`.
- [ ] `go vet ./...`.
- [ ] `git diff --check`.
- [ ] Run focused ship-design/construction/Fleet/save-load/Observer regressions.
- [ ] Verify existing Colony Ship/Outpost and blockade behavior remains intact.
- [ ] Update permanent evidence/status/HISTORY.
- [ ] Commit implementation, then closing docs; remove `_OPEN_` marker.

## Exit criterion

A deterministic persisted military design can be constructed through the normal strategic Construction path into a concrete owned ship represented in strategic Fleet composition, with exact save/load and Observer/replay semantics and without yet requiring tactical combat.