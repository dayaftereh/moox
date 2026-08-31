# Planned slice 04 - Generic combat Fleet movement, merge and split

Status: **closed; Gates 1-4 complete**.

Queue position: **4 of 7**.

## Objective

Generalize the strategic transit model from fixed Colony/Outpost special ships to concrete military fleets, including deterministic movement, fleet speed/range, composition-preserving merge/split and arrival semantics.

## Existing baseline

- Colony Ship transit already has semantic source/destination/ETA state and advances before current-turn Construction.
- System-coordinate/parsec distance and Fuel Cell base ranges exist.
- Core schema 18 now backs combat Fleets with concrete military `ShipIDs`; ordinary combat transit is intentionally still rejected.
- Directed hostile relations and deterministic blockade derivation already exist.

## Dependencies

- slice 03 concrete military Ship/Fleet composition;
- slice 02 supply-origin generalization if Outposts are implemented as planned.

## Scope guard

In scope:

- movement orders for owned military fleets;
- original Fleet speed and range aggregation rules;
- transit/arrival state;
- merge/split/recomposition at legal locations;
- deterministic blockade timing around departure/arrival;
- Observer/replay/save-load behavior.

Defer:

- tactical combat resolution;
- hostile encounter creation beyond the temporary strategic co-location boundary needed here (full handoff is slice 06);
- AI movement;
- transports/invasion;
- stealth/detection unless Gate 1 proves required for movement legality.

## Gate 1 - Checkup + original analysis / reverse engineering

- [x] Re-check current special-ship movement, Fleet composition and blockade code.
- [x] Create permanent evidence document.
- [x] Verify original combat Fleet movement-order timing and command legality.
- [x] Verify Fleet FTL speed behavior: for the currently supported normal military surface, original drives auto-upgrade Empire-wide and strategic speed uses current Empire drive plus race modifier; do not invent slowest build-snapshot aggregation.
- [x] Verify Fleet range and supply-origin rules: current military surface uses current Empire Fuel Cell base range and Colony/Outpost supply; Extended Fuel Tanks remains a future per-Ship modifier boundary.
- [x] Verify departure/arrival status semantics and ETA calculation against original 1.31.
- [x] Verify merge boundary: original stacks are co-location-derived; MOOX explicit merge is narrowed to same-owner stationary combat Fleets at one System.
- [x] Verify split semantics: original selected-Ship movement maps to stationary split/subset-move; in-transit mutation is technology-gated/deferred.
- [x] Verify blockade recomputation timing for departing/arriving combat Fleets.
- [x] Resolve hostile co-location boundary for slice 06: allow temporary stationary co-location/blockade here, defer encounter/BattleSession handoff.
- [x] Present findings and proposed generalized transit API/state; stop before implementation.

## Gate 2 - Implementation decision

- [x] Decide generalized Fleet transit state versus special-ship-only fields: schema 19 reuses `AtSystemID` / `DestinationSystemID` / `RemainingTurns`; ordinary combat `FTLSpeed` remains non-authoritative/zero.
- [x] Decide Fleet speed/range derivation: current Empire best Warp Drive + Trans-Dimensional; effective range through validated member hook, currently collapsing to current Empire best Fuel Cell.
- [x] Decide commands/events: `move_fleet` optional `ship_ids` atomic split+move; `split_fleet`; `merge_fleets`; `fleet_split`; `fleets_merged`; existing movement events retained.
- [x] Decide supply-origin lookup: reuse nearest owned Colony/Outpost helper; no military-only supply graph.
- [x] Decide hostile destination: allow temporary stationary hostile co-location/blockade; Slice 06 owns encounter/BattleSession handoff.
- [x] Decide schema/invariants: schema 19, strict mutually exclusive stationary/transit state, sorted unique non-empty combat `ShipIDs`, no in-transit split/merge/reroute.
- [x] Decide Observer/replay/save-load: command/resolver determinism, deterministic `NewID`, exact membership/transit round-trip, deep-cloned `ShipIDs`, split-before-move event order.

## Gate 3 - Implementation

- [x] Generalize movement legality to concrete military Fleets.
- [x] Derive Fleet FTL speed/range from authoritative composition according to Gate-1 evidence.
- [x] Implement deterministic transit progression/arrival.
- [x] Implement legal merge and split without duplicating/losing ship instances.
- [x] Recompute blockade state at the evidenced strategic timing boundaries.
- [x] Preserve special Colony/Outpost behavior through shared helpers where safe.
- [x] Add exact save/load and Observer clone/replay tests.
- [x] Add focused fixtures for current-Empire FTL/range derivation, Trans-Dimensional speed, range rejection, subset split+move, explicit merge/split and arrival.

## Gate 4 - Follow-up QA + commit + close

- [x] `gofmt` changed Go files.
- [x] `go test ./... -count=1`.
- [x] `go vet ./...`.
- [x] `git diff --check`.
- [x] Run focused combat-Fleet movement/merge/split/blockade/Colony Ship/Outpost regressions.
- [x] Verify no ship-instance duplication or loss across save/load and merge/split.
- [x] Update evidence/status/HISTORY and close marker through two-commit slice close.

## Exit criterion

A concrete military Fleet can move between systems under original-derived speed/range rules, arrive deterministically, merge and split without corrupting composition, and drive blockade state from its authoritative current location.