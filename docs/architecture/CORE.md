# Deterministic simulation core

MOOX keeps the gameplay simulation independent from UI and platform code. The Phase 1 core lives in `internal/core` and establishes deterministic state ownership independently from UI, networking and combat. The first colony-economy state is now layered onto that foundation.

## Current foundation

- stable monotonic simulation IDs (`core.ID`),
- simulation-owned SplitMix64 RNG with explicit serializable state,
- minimal galaxy, star-system, planet, empire and colony state,
- continuous `PopulationState` (`Total`, farmers, workers, scientists) with deterministic assignment tolerance,
- domain-native `float64` Colony economy snapshots (`Food`, `Production`, `Research`, `TaxBC`) plus continuous Population allocation and PP construction progress,
- materialized `ColonyPopulationDynamics` for capacity, Food/Production sustenance and projected turn-end growth,
- explicit `ColonyEconomyContext` plus `AdjustedEconomy` for currently implemented Gravity/Government/local-Morale layers,
- minimal validated colony building IDs used by the first proven local morale effects,
- turn clock and event log,
- strict state validation including cross-reference checks,
- deterministic JSON serialization,
- atomic file save and load,
- a fixed three-system headless fixture used for regression tests.

`NewSmallFixture` is test scaffolding. Its star positions and planet classes are deliberately explicit and are not evidence for original MOO2 galaxy-generation probabilities.

## Determinism contract

The simulation must not depend on Go standard-library PRNG implementation details, map iteration order, wall-clock time or UI state. Random state is part of `GameState` and must be committed explicitly after use. Save/load must reproduce the exact logical state, and the current fixture also reproduces identical serialized bytes after a load/save cycle.

## Next runtime work

The next slice should keep the same headless boundary and add only the state needed for the first colony economy loop: population assignment, food/production/research/money accounting and a deterministic turn processor. Original-behavior values should come from normalized ruleset data where available; missing gameplay formulas should be added only when their simulation tests require them.