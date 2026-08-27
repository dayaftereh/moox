# Colony economy baseline - 2026-08-27

Status: historical economy-evidence checkpoint; runtime numeric representation updated by `docs/architecture/ADR-0003-domain-native-float64.md`.

This checkpoint establishes only the base population-job economy needed to exercise the command/session/observer architecture. It is deliberately smaller than the complete MOO2 colony economy.

## Implemented population model

A colony stores continuous Population allocation with exactly three current jobs:

- farmers,
- workers,
- scientists.

`PopulationState.Total`, `Farmers`, `Workers` and `Scientists` are now `float64`. The role allocations must sum to `Total` within deterministic tolerance, so fractional allocations such as 1.25 farmers or 1.5 workers are legal.

The first real strategic command is:

```text
colony.assign_population
```

Its payload contains colony ID plus farmer/worker/scientist allocation values. The server validates that:

- assignments are non-negative,
- the assigned total equals the colony's assignable population,
- the colony exists,
- the submitting seat controls the colony's empire.

Seat -> empire authority comes from `GameSession`, not from the client payload.

## Base output values

The current runtime uses domain-native `float64` values directly in Food, PP, RP and BC. Determinism is maintained through stable evaluation order, finite-value validation and explicit tolerance/rounding boundaries rather than global fixed-point scaling.

### Food per farmer

`planet_classes.json` remains the source for original-observed climate food per farmer:

```text
toxic     0
radiated  0
barren    0
desert    1
tundra    1
ocean     2
swamp     2
arid      1
terran    2
gaia      3
```

The values are directly backed by the original 1.31 executable `Generate_Food_Per_Farmer_` table already normalized in the planet-class checkpoint.

Race Designer farming deltas come from `race_traits.json`. A racial farming penalty cannot reduce a life-bearing climate below 1 food/farmer. Farming race bonuses do not make Toxic/Radiated/Barren farmable. Aquatic is handled as a separately normalized +1 planet coefficient on Tundra/Ocean/Terran; it likewise does not make a zero-food climate farmable.

### Industry per worker

Base worker production is a separate rule from the already-normalized original `base_extraction` table. The two datasets must not be conflated.

The base industry/worker values used by this checkpoint are:

```text
ultra_poor  1
poor        2
abundant    3
rich        5
ultra_rich  8
```

These values are stored in `economy.json`, with `secondary-reference` manual provenance. Race Designer industry deltas come from `race_traits.json`, with a floor of 1 production/worker.

### Research per scientist

The neutral baseline is 3 RP/scientist and is stored in `economy.json`. Race Designer science deltas are applied from `race_traits.json`, with a floor of 1 RP/scientist.

The local original `HELP.LBX` also provides useful consistency evidence: Android Scientists generate 3 research per turn without racial bonuses, and the Science help text records the -1/+1/+2 racial modifier family and the minimum-1 behavior.

### Tax BC per population

The neutral base taxable income is 1 BC/population and is stored in `economy.json`. The Race Designer money deltas (-0.5/+0.5/+1 BC) are read from `race_traits.json`. MOOX now preserves fractional base BC (for example 1.5 BC) instead of rounding merely because the original runtime used integer storage. Explicit tax-bonus rounding remains a separate named rule where normalized.

This is base population income only. The game's player-controlled tax rate converts production into cash and is a distinct mechanic not implemented in this slice.

## Current base-economy calculation

For each colony the current materialized `ColonyEconomy` snapshot contains:

```text
Food       = farmers    * effective food/farmer
Production = workers    * effective industry/worker
Research   = scientists * effective research/scientist
TaxBC      = population * effective base BC/population
```

The result is recalculated during strategic resolution and is visible through both the owning `PlayerView` and privileged `ObserverView`.

A successful population command emits:

```text
colony.population_assigned
```

with previous assignment, current assignment and resulting base-economy snapshot.

## Explicitly not implemented yet

The current snapshot is a base/gross role output. It does not yet apply:

- government multipliers (including Democracy/Unification/Feudal behavior),
- morale,
- gravity mismatch and gravity racial traits,
- buildings or technology bonuses,
- flat building production/research/food,
- pollution and pollution-control rules,
- blockade effects,
- colony leader bonuses,
- food consumption,
- freighter import/export logistics,
- surplus-food sale,
- tax-rate production conversion,
- maintenance/expenses,
- treaties/trade income,
- specials such as Lithovore/Cybernetic/Fantastic Traders,
- population growth and maximum-population rules.

Those effects must be added as separately proven layers rather than folded into an unexplained formula.

## Determinism and multiplayer guard

The economy resolver receives a detached game state and stable command batches. `GameSession` supplies trusted Seat -> Empire authority. Invalid ownership or assignment totals reject the entire strategic resolution transaction, leaving authoritative state/revision/events unchanged.

This is covered by both resolver-level and full GameSession integration tests.

## Next fidelity slice

The next economy work should normalize and test the contextual modifiers needed to turn base output into effective output. The likely first dependency chain is gravity compatibility and government effects, followed by morale/buildings/pollution. Exact ordering and rounding should be proven before implementation.