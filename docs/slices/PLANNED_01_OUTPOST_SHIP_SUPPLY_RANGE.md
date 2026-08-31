# Planned slice 01 - Outpost Ship, Outpost state and supply range

Status: **planned / queued; not open**.

Queue position: **1 of 6**.

This file is a prepared slice specification. It is not an `_OPEN_*.md` recovery marker and does not mean implementation has started.

## Objective

Extend the completed Colony Ship expansion loop with the smallest authoritative Outpost loop:

```text
Colony
-> construct Outpost Ship
-> move through strategic transit
-> deploy Outpost at a legal destination
-> persist Outpost ownership/state
-> use the Outpost as a strategic supply origin
```

The slice should also resolve the minimum original rule needed for later Outpost-to-Colony replacement/conversion, without absorbing combat, planetary defenses or AI.

## Existing baseline

- Core `StateSchemaVersion` is 16.
- Colony Ship Construction, fixed installed FTL speed, Fuel-range-checked movement and explicit colonization are implemented.
- Strategic range currently treats Empire-owned Colony systems as the only authoritative supply origins.
- Direct original evidence already maps special ship class `4` to **Outpost Ship**.
- Technology `109` is normalized as `outpost_ship`.
- The previously inspected special-ship cost helper reports an Outpost Ship base cost of **100 PP** before government ship-cost adjustment.

Permanent predecessor evidence: `docs/research/COLONY_SHIP_MOVEMENT_COLONIZATION_2026-08-30.md`.

## Dependencies

- completed Colony Ship / strategic transit slice (`db9d01e`);
- current semantic Construction project model;
- current system-coordinate/parsec distance and Fuel Cell range helpers.

## Scope guard

In scope only if directly evidenced/required:

- Outpost Ship buildability, cost and completion;
- installed-drive semantics;
- Outpost Ship strategic movement;
- legal Outpost destination/deployment;
- canonical Outpost ownership/state;
- supply-range participation;
- minimum Outpost-to-Colony replacement semantics needed to avoid contradictory state.

Explicitly defer:

- military fleet combat;
- star-base/battlestation defenses;
- AI Outpost targeting;
- diplomacy permissions beyond current ownership/hostility prerequisites;
- Transport/invasion;
- tactical combat.

## Gate 1 - Checkup + original analysis / reverse engineering

- [ ] Re-check Git status, `_OPEN_*.md`, Core schema and current strategic movement implementation.
- [ ] Create the permanent evidence document for this slice.
- [ ] Verify Outpost Ship technology gate and base/government-adjusted production cost in original MOO2 1.31.
- [ ] Verify whether the Outpost Ship installs the current best Warp Drive exactly like the Colony Ship or uses different design semantics.
- [ ] Verify strategic Fuel range and whether Outpost Ships can use existing Colony/Outpost supply while travelling.
- [ ] Resolve original Outpost placement legality: empty system vs planet target, system ownership restrictions and hostile presence handling.
- [ ] Resolve the authoritative original Outpost representation sufficiently to choose a MOOX semantic state model.
- [ ] Verify exactly how an Outpost extends supply range and when that effect becomes active in turn order.
- [ ] Verify what happens when a Colony is founded in a system containing the owner's Outpost: replacement, conversion or coexistence.
- [ ] Verify blockade/hostility implications required by this slice.
- [ ] Separate direct facts, original-derived interpretation and deliberate MOOX modernization.
- [ ] Present Gate-1 findings and a narrow implementation proposal; stop before gameplay implementation.

## Gate 2 - Implementation decision

Do not mark complete until discussed/accepted.

Candidate decisions to resolve:

- [ ] Core state shape for an Outpost (dedicated entity vs semantic system/planet ownership state).
- [ ] Schema-version impact.
- [ ] Construction project kind/ID and command/event names.
- [ ] Whether Outpost Ship reuses the existing Colony Ship transit fields or requires a generalized special-ship path first.
- [ ] Authoritative supply-origin helper shared by Colony and Outpost state.
- [ ] Deployment target identity and validation.
- [ ] Outpost-to-Colony transition semantics.
- [ ] Observer/replay/save-load contract.
- [ ] Accepted deliberate divergences and deferrals.

## Gate 3 - Implementation

Tentative implementation checklist; adjust only to the accepted Gate-2 contract.

- [ ] Add/extend Core state and validation for Outposts and/or Outpost Ship identity.
- [ ] Add Outpost Ship legal Construction choice and queue command.
- [ ] Produce a stationary Outpost Ship with persisted installed FTL speed.
- [ ] Reuse/generalize strategic movement and deterministic ETA/range validation.
- [ ] Add deployment command and authoritative Outpost creation/ship consumption.
- [ ] Include Outposts in Empire supply-origin range calculations at the evidenced timing boundary.
- [ ] Implement minimum Outpost-to-Colony replacement rule if Gate 1 proves it belongs here.
- [ ] Add exact save/load, Observer clone isolation and replay events.
- [ ] Add a vertical deterministic test: build -> move -> arrive -> deploy -> supply extension.
- [ ] Keep combat, AI and unrelated diplomacy out of scope.

## Gate 4 - Follow-up QA + commit + close

- [ ] `gofmt` all changed Go files.
- [ ] `go test ./... -count=1`.
- [ ] `go vet ./...`.
- [ ] `git diff --check`.
- [ ] Run focused Outpost/Colony Ship/strategic transit/supply/blockade regressions.
- [ ] Verify exact Core save/load and Session Observer/replay behavior.
- [ ] Update permanent evidence, README/status/HISTORY as required.
- [ ] Commit gameplay implementation.
- [ ] Commit closing documentation and delete the single `_OPEN_` marker.

## Exit criterion

An Empire can construct an Outpost Ship, move it legally, deploy an authoritative Outpost, and immediately/at the correctly evidenced turn boundary use that Outpost as a supply origin. The state round-trips exactly and no `_OPEN_` marker remains after closure.