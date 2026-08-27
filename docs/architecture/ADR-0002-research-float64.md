# ADR-0002: Use float64 research points in the MOOX runtime

- Status: Superseded
- Superseded by: `ADR-0003-domain-native-float64.md`
- Date: 2026-08-27
- Scope: deterministic strategic research runtime and protocol/state serialization

## Context

The original Master of Orion II 1.31 executable stores several research values as integers because the 1996 implementation had very different memory and data-layout constraints. Reverse engineering those fields is valuable for understanding game behavior, but reproducing the old storage width is not itself a MOOX requirement.

MOOX has no meaningful memory pressure at this scale and benefits from domain values that are directly understandable by UI, AI, Observer tooling and save-game inspection. A value of `153.75 RP` should be represented as `153.75`, not as `153750` implicit milli-units or a legacy integer that discards the fraction.

## Decision

Research progress is represented in the simulation as RP-native `float64` values.

```go
ResearchState{
    ProgressRP: 153.75,
}
```

Technology-field base costs are also exposed to the game layer as `float64` RP.

At the time of this ADR the wider economy still used fixed-point milli-units. That broader numeric policy is now superseded by ADR-0003: Colony economy, Population allocation and Construction progress also use domain-native `float64`. The research examples below remain valid as the first step in that direction:

```text
Colony A = 1.9 RP
Colony B = 1.9 RP
Empire research = 3.8 RP
```

not:

```text
floor(1.9) + floor(1.9) = 2 RP
```

## Rounding policy

Do not round research merely because the original executable stored an integer.

Rounding is applied only at an explicit gameplay-rule boundary. The currently verified breakthrough rule still uses a discrete 1..100 roll, so MOOX computes the percentage from full floating-point RP and floors only the final percentage used for that roll:

```text
raw_chance = (projected_rp - base_cost_rp) * 100 / base_cost_rp
chance_percent = floor(raw_chance)
chance_percent = clamp(chance_percent, 1, 100) when projected_rp > base_cost_rp
```

This deliberately preserves the recognizable MOO2 breakthrough curve while removing the old integer-RP accumulation constraint.

Future rules may choose `floor`, `ceil`, nearest rounding, or no rounding at all. The rule must state the boundary explicitly.

## Determinism contract

Using `float64` does not relax deterministic simulation requirements.

- Only finite, non-negative Research RP is valid. NaN and +/-Inf are rejected by `GameState.Validate`.
- Aggregations run in deterministic game-state order; concurrent/network arrival order must never determine floating-point reduction order.
- Game decisions must not depend on locale-formatted decimal strings.
- JSON/save values use numeric RP (`progress_rp`), not formatted text.
- If a future rule becomes numerically sensitive enough that platform-independent bit identity cannot be maintained with ordinary Go `float64`, that specific rule gets an explicit canonicalization/quantization boundary rather than reverting the whole research model to legacy integers.

## State schema

Changing persistent research state from:

```text
progress_milli: int64
```

to:

```text
progress_rp: float64
```

is an intentional state-format change. `StateSchemaVersion` is therefore bumped from 1 to 2.

No compatibility shim is added for schema-1 development fixtures at this stage; the project is still pre-release and the new representation is the canonical architecture going forward.

## Consequences

Positive:

- UI, AI and Observer APIs receive human-scale RP directly.
- fractional bonuses can accumulate without silent loss,
- fewer legacy-storage assumptions leak into new systems,
- later modifiers can be expressed naturally,
- save/debug output is easier to inspect.

Trade-offs:

- MOOX will not reproduce bit-for-bit the original integer research accumulation stream,
- tests must use rule-aware comparisons when calculations stop being exactly representable,
- future simulation code must keep aggregation order deterministic.

## Non-goals

This ADR does not require converting Construction, BC, Food or Production to `float64`. Those systems may be reconsidered independently when there is a gameplay or architecture benefit.
