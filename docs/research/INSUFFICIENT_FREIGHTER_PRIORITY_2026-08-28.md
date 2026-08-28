# Original insufficient-Freighter priority - 2026-08-28

Status: directly verified from the private Master of Orion II 1.31 executable and implemented for MOOX's current single-cohort Population model.

## Scope

This checkpoint answers the previously open question: when an Empire has transferable Food but too few available Freighters to satisfy every Colony shortage, which Colony receives the limited transport capacity first?

The authoritative original routine is `Pass_Out_Imports_` at VA `0xDF8F0`. Supporting evidence comes from `Colony_Food_Maintenance_` at `0xDEB4B`, `Pre_Import_Computing_` at `0xE1D59`, and `N_Freighters_Player_Can_Scrap_` at `0xFF477`.

## Available Freighter pool

`Pass_Out_Imports_` starts with the same free-Freighter calculation exposed by `N_Freighters_Player_Can_Scrap_`:

```text
available_freighters = player[+0x36] - 5 * player[+0x40]
```

The important fidelity point for this slice is that the import allocator receives a discrete available-Freighter budget and consumes one unit of that budget for each one-Food import increment.

MOOX already models the Empire Freighter pool explicitly. Population transport reservation is still deferred until the strategic transport model exists.

## Deficit list and signed import field

`Pass_Out_Imports_` scans the original global Colony array in increasing array-index order. Eligible Colonies belonging to the target player are classified by:

```text
local_delta = colony.food_production - colony.food_maintenance
```

Relevant original fields are:

```text
Colony +0xE7  current Food production
Colony +0xEF  whole-Food maintenance requirement
Colony +0xF3  signed Food import/export amount
```

For a deficit Colony (`local_delta < 0`), `+0xF3` starts at zero and the Colony index is appended to the deficit list. For a surplus Colony, `+0xF3` is initialized to the negative surplus and the surplus contributes to an Empire-wide Food pool.

If both Food surplus and available Freighters are sufficient, every deficit is filled directly. The interesting path is the constrained branch beginning around `0xDFA99`.

## Constrained allocation is round-robin, not proportional

The original does not divide the available Food proportionally between shortages. It repeatedly walks the deficit list in its existing order. Whenever the current Colony is below the current feeding threshold, the routine performs:

```text
colony.imported_food += 1
remaining_food_surplus -= 1
remaining_freighters -= 1
```

After reaching the end of the deficit list, it wraps to the first deficit Colony and repeats until either the current threshold is satisfied everywhere or Food/Freighters are exhausted.

For a simple homogeneous Empire with shortages `[3, 1, 2]` and only four transferable Food/Freighters, the resulting import allocation is therefore:

```text
round 1: [1, 1, 1]
round 2: [2, 1, 1]
result : [2, 1, 1]
```

This is observably different from proportional sharing.

## Why the original has four passes

`Colony_Food_Maintenance_` builds several half-Food (`food2`) maintenance buckets at Colony `+0xFC/+0xFD/+0xFE/+0xFF`, derived from the Colony's individual Population entries and their race/status properties. It then computes whole-Food maintenance at `+0xEF` as the rounded half-Food total.

`Pass_Out_Imports_` uses four progressively broader round-robin thresholds:

```text
pass 1: 2 * (production + imports) < FC
pass 2: 2 * (production + imports) < FC + FD
pass 3: 2 * (production + imports) < FC + FD + FE
pass 4:     (production + imports) < EF
```

Each successful step still consumes exactly one Food from the surplus pool and one available Freighter.

These extra stages matter when a Colony contains mixed population ownership/status cohorts. MOOX does not yet have race-aware Population cohorts; each Colony currently carries one aggregate `PopulationState` under one Empire/Race context. In that current model, the relevant demand collapses to a single round-robin deficit threshold.

The later original cohort stages are therefore documented but deliberately deferred until the Population cohort model exists. They are not guessed into aggregate state.

## MOOX implementation

`internal/game/food_logistics.go` now replaces the temporary proportional import allocator with deterministic original-style round-robin allocation:

```text
1. `coloniesForEmpire` provides stable Colony-ID order;
2. each pass gives a Colony at most one Freighter-load;
3. a Colony is skipped once its local shortage is satisfied;
4. the pass wraps until the authoritative transfer budget is exhausted.
```

MOOX uses stable monotonic Colony IDs as the clean deterministic analogue of the original mutable Colony-array index order.

Continuous Food remains a MOOX architecture choice. A final partial Freighter-load may therefore be fractional; no integer-only storage is reintroduced merely to copy the 1996 representation.

Source-Colony export attribution remains MOOX telemetry. The original constrained algorithm consumes an Empire-wide surplus pool rather than selecting a source route per imported unit, so this checkpoint changes starvation-relevant import priority only.

## Regression

`TestAllocateFoodImportsFollowsOriginalRoundRobinPriority` locks the `[3,1,2]` shortage / four-Freighter case to `[2,1,1]` imports. Existing starvation, fractional Food and Treasury/Freighter accounting tests remain unchanged.

## Remaining Freighter work

The exact insufficient-Freighter import priority is closed for the current aggregate Population model. Remaining strategic logistics work is structurally separate:

- blockade eligibility/effects;
- Population transport and reservation of the shared Freighter pool;
- mixed race/status Population cohorts, at which point the original four feeding passes can be represented completely.