# Planned slice 04 - Generic combat Fleet movement, merge and split

Status: **planned / queued; not open**.

Queue position: **4 of 7**.

## Objective

Generalize the strategic transit model from fixed Colony/Outpost special ships to concrete military fleets, including deterministic movement, fleet speed/range, composition-preserving merge/split and arrival semantics.

## Existing baseline

- Colony Ship transit already has semantic source/destination/ETA state and advances before current-turn Construction.
- System-coordinate/parsec distance and Fuel Cell base ranges exist.
- Strategic combat Fleets currently provide blockade inputs but are not yet backed by concrete ship composition until slice 02.
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
- hostile encounter creation beyond any minimum legality guard needed here (full handoff is slice 05);
- AI movement;
- transports/invasion;
- stealth/detection unless Gate 1 proves required for movement legality.

## Gate 1 - Checkup + original analysis / reverse engineering

- [ ] Re-check current special-ship movement, Fleet composition and blockade code.
- [ ] Create permanent evidence document.
- [ ] Verify original combat Fleet movement-order timing and command legality.
- [ ] Verify Fleet FTL speed aggregation, expected to depend on the slowest ship but not assumed until directly proven.
- [ ] Verify Fleet range aggregation and supply-origin rules for mixed designs, including Extended Fuel Tanks only if it materially affects the first generic implementation.
- [ ] Verify departure/arrival status semantics and ETA calculation against original 1.31.
- [ ] Verify merge conditions: same owner/location/destination/status and any original exceptions.
- [ ] Verify split semantics and whether movement progress/ETA is preserved, recomputed or constrained.
- [ ] Verify blockade recomputation timing for departing/arriving combat Fleets.
- [ ] Resolve what happens when movement would create a hostile co-location; isolate the boundary for slice 05.
- [ ] Present findings and proposed generalized transit API/state; stop before implementation.

## Gate 2 - Implementation decision

- [ ] Decide generalized Fleet transit state versus special-ship-only fields.
- [ ] Decide Fleet speed/range derivation from concrete ship composition.
- [ ] Decide move, merge and split command payloads and event names.
- [ ] Decide supply-origin lookup shared by Colony/Outpost/military movement.
- [ ] Decide hostile-destination boundary without prematurely implementing combat.
- [ ] Decide schema changes and invariants for stationary vs moving Fleets.
- [ ] Decide Observer/replay/save-load semantics.

## Gate 3 - Implementation

- [ ] Generalize movement legality to concrete military Fleets.
- [ ] Derive Fleet FTL speed/range from authoritative composition according to Gate-1 evidence.
- [ ] Implement deterministic transit progression/arrival.
- [ ] Implement legal merge and split without duplicating/losing ship instances.
- [ ] Recompute blockade state at the evidenced strategic timing boundaries.
- [ ] Preserve special Colony/Outpost behavior through shared helpers where safe.
- [ ] Add exact save/load and Observer clone/replay tests.
- [ ] Add focused fixtures for slowest-ship speed, range rejection, merge, split and arrival.

## Gate 4 - Follow-up QA + commit + close

- [ ] `gofmt` changed Go files.
- [ ] `go test ./... -count=1`.
- [ ] `go vet ./...`.
- [ ] `git diff --check`.
- [ ] Run focused combat-Fleet movement/merge/split/blockade/Colony Ship/Outpost regressions.
- [ ] Verify no ship-instance duplication or loss across save/load and merge/split.
- [ ] Update evidence/status/HISTORY and close marker through two-commit slice close.

## Exit criterion

A concrete military Fleet can move between systems under original-derived speed/range rules, arrive deterministically, merge and split without corrupting composition, and drive blockade state from its authoritative current location.