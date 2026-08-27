# ADR-0003: Use domain-native float64 for continuous game quantities

- Status: Accepted
- Date: 2026-08-27
- Scope: strategic simulation state, economy, population allocation, construction progress, research and protocol/save representations
- Supersedes: the numeric-policy part of `ADR-0002-research-float64.md`

## Context

The clean-room project uses original Master of Orion II structures and arithmetic as behavioral evidence, not as a requirement to reproduce 1996 storage constraints. Earlier MOOX runtime slices used fixed-point `int64` values with an `EconomyScale` of 1000 for Food, Production, Research and BC, while Population assignments remained whole integers.

That representation is deterministic, but it leaks implementation mechanics into every API and silently creates intermediate rounding boundaries that are not inherently part of the game design. MOOX has no relevant memory pressure that requires narrow integer storage for these quantities.

## Decision

Continuous game quantities are represented directly in their domain units as `float64`.

Current canonical examples:

```go
PopulationState{
    Total:      4.0,
    Farmers:    1.25,
    Workers:    1.50,
    Scientists: 1.25,
}

ColonyEconomy{
    Food:       2.50,
    Production: 4.50, // PP
    Research:   3.75, // RP
    TaxBC:      4.00,
}

ConstructionState{
    BuildingID: "holo_simulator",
    ProgressPP: 3.75,
}
```

The same principle applies to future continuous stores and rates unless a rule requires discrete values.

## What stays discrete

Do not convert identity or count-like concepts merely for consistency. These remain integer/string/enum concepts where appropriate:

- entity IDs,
- TechField and Technology IDs,
- building ownership entries,
- ship counts where individual ships are represented,
- command sequence numbers,
- turn/revision counters,
- percentages that are currently normalized as integer rule parameters,
- enum/index/table identifiers.

The distinction is semantic: continuous quantities use `float64`; genuinely discrete concepts remain discrete.

## Rounding policy

Rounding is not an implementation side effect. It must be a named gameplay-rule boundary.

Therefore MOOX does not round after each modifier or role calculation. For example:

```text
1.25 farmers * 2 food/farmer = 2.5 food
1.50 workers * 3 PP/worker = 4.5 PP
1.25 scientists * 3 RP/scientist = 3.75 RP
```

Government, morale and gravity percentages are applied to the full floating-point value.

If a proven or intentionally designed rule requires a discrete result, that code performs an explicit `floor`, `ceil`, `round`, clamp or quantization at that rule boundary. Examples include the existing discrete 1..100 research breakthrough roll and normalized tax-bonus rounding policy.

UI formatting is not simulation rounding. A UI may display `3.8` while the authoritative state retains `3.75`.

## Population model

Population is continuous in MOOX. `PopulationState.Total` and the Farmer/Worker/Scientist allocations are `float64` and must sum to the total within deterministic numeric tolerance.

This intentionally permits allocations such as:

```text
Total       4.00
Farmers     1.25
Workers     1.50
Scientists  1.25
```

The server remains authoritative: `colony.assign_population` validates non-negative finite values and verifies the submitted allocations against the Colony total.

## Construction model

Building costs and construction progress are expressed directly in PP:

```text
production_cost_pp: 120
progress_pp: 37.5
```

No `* 1000` conversion exists in the runtime. Fractional production is applied without truncation. Completion uses a small deterministic numeric tolerance so floating-point residue cannot create a phantom extra construction turn.

## Economy model

`ColonyEconomy` stores direct Food / PP / RP / BC values. Base role output and contextual output preserve fractions through race, government, morale and gravity calculations.

The normalized ruleset may originate from integer or decimal reference values; the game layer converts those values directly to `float64` domain units rather than scaled integers.

## Determinism contract

Using `float64` does not permit nondeterministic aggregation.

- state values must be finite and non-negative where the domain requires it,
- NaN and +/-Inf are invalid,
- aggregation/evaluation order must be stable and independent of network arrival order,
- simulation comparisons that require equality use a documented tolerance where arithmetic can accumulate error,
- JSON stores numeric values, never locale-formatted decimal strings,
- authoritative simulation logic must not depend on display rounding,
- if a future cross-platform calculation proves too numerically sensitive, canonicalization is added at that specific rule boundary rather than reintroducing global fixed-point storage.

## Persistent state schema

The original Float migration bumped `StateSchemaVersion` from 2 to 3 with these canonical changes:

```text
population.units        -> population.total
food_milli              -> food
production_milli        -> production
research_milli          -> research
tax_bc_milli            -> tax_bc
construction.progress_milli -> construction.progress_pp
```

All affected values are JSON numbers backed by `float64` in Go.

The Population Growth/Sustenance slice subsequently bumps the development state to schema 4 by adding materialized `population_dynamics` (`capacity`, sustenance, available Production and projected Growth). This is an extension of the same Float policy, not a return to scaled integers.

Food/Freighter logistics and Starvation subsequently bump the development state to schema 5 by adding direct-float logistics materialization and projected starvation plus a discrete Freighter count. Multi-Technology Research subsequently bumps the development state to schema 6 by adding discrete `selection_mode` and persisted Uncreative field/application choices; this does not change the domain-native float policy. The project is still pre-release, so older development-schema migration shims are not retained; schema 6 is the canonical development state going forward.

## Relationship to original MOO2 behavior

Original integer fields, storage widths and rounding instructions remain valuable reverse-engineering evidence. They may still define a gameplay rule when that rule is important to MOOX.

They do not automatically define MOOX's storage representation.

This means the project can preserve recognizable rules such as breakthrough probability while allowing finer population allocation, fractional bonuses and continuous resource accumulation.

## Consequences

Positive:

- APIs expose human-scale values directly,
- fractional modifiers are not silently lost,
- UI and AI receive the same natural quantities,
- save/Observer output is easier to inspect,
- fewer legacy implementation details leak through the architecture,
- future growth/food/economy systems can evolve continuously.

Trade-offs:

- some results intentionally diverge from original intermediate rounding,
- exact float comparisons must be avoided after nontrivial arithmetic,
- older development schemas are no longer accepted without an explicit future migration layer,
- tests must state the intended rule boundary rather than assume whole-number output.
