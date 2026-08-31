# Planned slice 05 - Command Points and ship overage Maintenance

Status: **planned / queued; not open**.

Queue position: **5 of 7**.

## Objective

Connect concrete military ships to the already researched Treasury Maintenance ledger by implementing authoritative Command-Point capacity/usage and the original overage Maintenance boundary.

## Existing baseline

- Treasury settlement and Maintenance snapshots are authoritative.
- Direct Treasury evidence already identifies **Ship command-point overage Maintenance** as one of the original six Maintenance buckets.
- That bucket was deliberately deferred because the runtime had no canonical ships or Command-Point capacity/usage.
- Slices 02/03 are intended to provide concrete ships and Fleet composition first.

Predecessor evidence: `docs/research/TREASURY_MAINTENANCE_DEFICIT_2026-08-29.md`.

## Dependencies

- slice 03 military ship instances/hulls;
- preferably slice 04 authoritative Fleet composition/movement, although accounting should depend on owned ships rather than location if original evidence says so.

## Scope guard

In scope:

- Command-Point cost/usage of owned ships;
- Empire Command-Point capacity from only the original sources proven and dependency-safe in Gate 1;
- over-capacity Maintenance calculation and rounding;
- integration into Treasury settlement and Observer/replay snapshots.

Defer:

- Spy, Tribute and Officer/Leader Maintenance;
- full staged negative-Treasury liquidation;
- combat effects of command infrastructure;
- unrelated base/defense behavior unless it is required solely as a Command-Point capacity producer.

## Gate 1 - Checkup + original analysis / reverse engineering

- [ ] Re-check Treasury evidence and current settlement order.
- [ ] Create permanent Command-Point evidence document.
- [ ] Verify original per-hull/per-ship Command-Point consumption.
- [ ] Verify Empire base Command-Point capacity.
- [ ] Verify capacity producers/modifiers: Star Base/Battlestation/Star Fortress, technologies, race/government effects and any difficulty modifiers, adding only directly proven dependencies.
- [ ] Verify whether special civilian ships consume Command Points.
- [ ] Verify exact overage Maintenance formula, sign, rounding/truncation and caps.
- [ ] Verify timing relative to Construction completion, ship loss/scrap and Treasury settlement.
- [ ] Verify whether ships in transit/reserve/special locations count differently.
- [ ] Resolve materialized Treasury snapshot fields/events needed for auditability.
- [ ] Present findings and narrow implementation proposal; stop before implementation.

## Gate 2 - Implementation decision

- [ ] Decide where per-ship Command-Point cost lives (hull/design/ship instance).
- [ ] Decide Empire capacity/usage derived vs persisted fields.
- [ ] Decide which proven capacity producers are safe to model in this slice.
- [ ] Decide Treasury snapshot extension and schema/ruleset version impact.
- [ ] Decide exact resolution timing within the existing strategic turn.
- [ ] Decide legal-action/Observer/replay exposure.
- [ ] Preserve explicit deferral of full deficit liquidation.

## Gate 3 - Implementation

- [ ] Add normalized Command-Point values/producers proven in Gate 1.
- [ ] Derive authoritative Empire capacity and usage.
- [ ] Compute overage Maintenance using exact original rounding/timing.
- [ ] Integrate the bucket into Treasury settlement/materialized snapshot.
- [ ] Add events/Observer projection as required.
- [ ] Add tests below/at/above capacity and around whole-BC boundaries.
- [ ] Add construction/combat-loss timing regressions if those transitions affect same-turn cost.
- [ ] Verify existing Food/Freighter/Building Maintenance remains unchanged.

## Gate 4 - Follow-up QA + commit + close

- [ ] `gofmt` changed Go files.
- [ ] `go test ./... -count=1`.
- [ ] `go vet ./...`.
- [ ] `git diff --check`.
- [ ] Run focused Treasury/Command-Point/ship construction/Fleet regressions.
- [ ] Compare fixed original-reference accounting fixtures at rounding boundaries.
- [ ] Update evidence/status/HISTORY and close marker.

## Exit criterion

Owned ships produce deterministic authoritative Command-Point usage, the Empire has evidence-backed capacity, and Treasury settlement charges the exact modeled ship-overage Maintenance bucket without inventing the still-missing Spy/Leader/treaty deficit systems.