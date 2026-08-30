# Open slice: Treasury Maintenance categories and deficit/scrap policy

Status: Gate 4 QA passed; implementation commit pending.

## Objective

Resolve the original Master of Orion II 1.31 strategic Treasury income/Maintenance categories and deficit/scrap behavior far enough to extend the existing MOOX Treasury ledger only where its dependent state is already authoritative.

## Recovery

- Starting branch: `main`.
- Starting HEAD: `d891c27` (`docs: close strategic blockade slice`).
- Starting tree: clean; `main` ahead of `origin/main` by 14 commits.
- Previous gameplay slice: strategic Fleet/Diplomacy blockade production (`d148502`), closed.
- Core `StateSchemaVersion`: 15.
- Economy ruleset schema: 7.
- Existing Treasury evidence: `docs/research/TREASURY_SETTLEMENT_2026-08-28.md`.
- Permanent Gate 1 evidence for this slice: `docs/research/TREASURY_MAINTENANCE_DEFICIT_2026-08-29.md`.

## Gates

- [x] Gate 1: repository checkup + original MOO2 1.31 reverse engineering
- [x] Gate 2: implementation decision
- [x] Gate 3: implementation
- [ ] Gate 4: QA + commit + close

## Gate 1 questions

1. What do the six original Maintenance component words at player `+0xB8..+0xC2` represent?
2. Which original gross-income components feed player `+0xAE`, and which of those have authoritative dependencies in current MOOX?
3. What exact arithmetic/rounding/order is used when gross income and Maintenance become current-turn net BC at `+0xB2`?
4. What does `Player_Maintenance_` do when Treasury plus current-turn net BC is insufficient, and what assets can it scrap/disband?
5. What ordering/tie-breaking protects determinism when multiple scrap candidates exist?
6. Which original categories must remain deferred because MOOX does not yet own canonical Ship/Leader/Spy/Trade/etc. state?
7. What is the smallest Core/Game/Session contract justified for Gate 2 without duplicating future subsystem truth?

## Scope guard

Do not implement gameplay during Gate 1. Do not fabricate Ship, Leader, Spy, diplomacy/trade or other state solely to complete original Treasury labels. A category may remain explicitly deferred until its source subsystem is authoritative.
## Gate 1 result

The original six Maintenance buckets are now resolved as Building/Colony, active Freighters, Ship command-point overage, Spies, outgoing Tribute, and Officers/Leaders. Gross BC additionally includes Colony BC production, Leader economic effects, treaty/tribute effects and surplus Food.

The full original negative-Treasury policy is staged and asset-aware across Buildings, Spies, Ships and Leaders. Current MOOX lacks canonical state for most of those classes, so a partial automatic scraper would be observably non-original and remains deferred.

Two corrections are dependency-safe now:

1. active Freighter Maintenance includes Food Freighters plus five-Freighter Population-transfer reservations and truncates aggregate cost to `floor(active/2)` BC;
2. surplus-Food income has a direct whole-BC boundary: normal `floor(surplus*0.5)`, Fantastic Traders `floor(surplus)` under MOOX's continuous Food representation.

## Proposed Gate 2 contract

- Keep Core schema 15 and Economy ruleset schema 7.
- Correct `FreighterOperatingCostBC` using existing `PopulationTransportFreightersReserved + FreightersUsed`, with original aggregate floor/truncation.
- Correct surplus-Food BC conversion at the directly proven whole-BC boundary while retaining continuous `SurplusFoodSold` telemetry.
- Keep existing Building Maintenance and Treasury settlement order.
- Do not add fake Ship/Spy/Tribute/Officer buckets or partial deficit liquidation.
- Preserve negative `BalanceBC` as an explicit modernization until all asset classes required by original `Player_Maintenance_` exist.

Gate 2 must accept or revise this contract before Gate 3 implementation begins.
## Gate 2 decision - 2026-08-30

Accepted without revision: keep Core schema 15 and Economy ruleset schema 7; correct active-Freighter Maintenance using existing reservation/usage state with aggregate whole-BC truncation; correct surplus-Food sale at the directly observed whole-BC boundary; do not add partial deficit liquidation or fake zero-valued Ship/Spy/Tribute/Officer buckets.

## Gate 3 implementation - 2026-08-30

Implemented the accepted narrow fidelity contract without schema changes:

- materializeFoodLogistics now derives active Freighters from owned Freighters minus post-allocation unused Freighters, so Population-transfer reservations and Food-transfer use share one original Maintenance bucket;
- Freighter operating cost materializes at the original aggregate whole-BC boundary via loor(active * configured_rate);
- surplus-Food telemetry remains continuous, while BC income now materializes at loor(SurplusFoodSold * saleRate);
- no partial deficit liquidation and no fake Ship/Spy/Tribute/Officer state were added.

Focused regressions cover one/three active Freighters, five reserved Settler Freighters, five reserved plus one Food Freighter, Treasury consumption of that bucket, normal 1/2/3/fractional surplus Food and Fantastic Traders.

## Gate 4 QA

- gofmt on changed Go files: pass
- go test ./...: pass
- go vet ./...: pass
- git diff --check: pass

The marker remains open until the implementation and closing documentation commits are complete.
