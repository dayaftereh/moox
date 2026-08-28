# Food, Freighter logistics and starvation - 2026-08-27

Status: implemented deterministic Economy checkpoint; exact original insufficient-Freighter allocation priority and global 1.31 turn ordering remain open research.

This slice extends the direct-`float64` Population/Sustenance model with empire-wide Food balancing, Freighter capacity/cost accounting, surplus-Food valuation and starvation.

## Verified / cross-checked mechanics used

The normalized rules are recorded in `data/rulesets/moo2-1.31/economy.json` and currently rely on secondary/manual-transcription evidence:

- one active Freighter transports one Food;
- a Freighter Fleet represents five Freighters;
- the Freighter Fleet production cost is 50 PP;
- Food balancing between colonies is automatic/immediate for the strategic turn;
- a Freighter actually used for Food transport costs 0.5 BC for the turn; an idle Freighter has no operating cost;
- remaining surplus Food is valued at 0.5 BC/Food normally;
- Fantastic Traders values remaining surplus Food at 1 BC/Food;
- ordinary starvation penalty is 0.05 Population per unresolved Food shortage in MOOX direct Population units;
- Cybernetic starvation uses 0.025 Population per unresolved Food shortage plus 0.025 per unresolved Production shortage;
- classic/default starvation does not eliminate the final Population unit.

The values above are gameplay-rule inputs. MOOX does not reintroduce the old thousand-population integer representation.

## State model

`StateSchemaVersion = 5` adds an empire-owned discrete Freighter pool and a materialized logistics snapshot:

```text
Empire.Freighters
Empire.FoodLogistics
```

`EmpireFoodLogistics` records:

```text
freighters_required
freighters_used
local_food_surplus
local_food_shortage
food_transferred
food_unmet
surplus_food_sold
freighter_operating_cost_bc
surplus_food_income_bc
```

The Colony `PopulationDynamics` snapshot now distinguishes local production from empire balancing:

```text
local_food_surplus
local_food_shortage
food_imported
food_exported
food_surplus      # remaining after logistics
food_shortage     # remaining after logistics
projected_growth
projected_starvation
```

This distinction is intentional for Player/Observer/AI explainability. A caller can tell whether a colony feeds itself, imports Food successfully or remains undersupplied because the empire lacks transport capacity.

## Deterministic logistics resolution

For every empire, colonies are considered in stable simulation-ID order. First, local Food supply/demand is materialized from the current-turn colony economy. The maximum transportable amount is:

```text
transfer_possible = min(total_local_surplus, total_local_shortage)
freighter_capacity = freighters * food_per_freighter
food_transferred   = min(transfer_possible, freighter_capacity)
```

A fractional amount of Food may use a Freighter because MOOX Food is continuous. Freighters used are therefore:

```text
ceil(food_transferred / food_per_freighter)
```

`freighters_required` similarly describes how many Freighters would be required to satisfy all transferable shortage.

### Insufficient-Freighter priority

Available references establish capacity and automatic transfer, but do not reliably establish the exact original priority when several colonies compete for too few Freighters. Some community observations even describe the algorithm as unclear.

MOOX therefore currently uses an explicit deterministic proportional-sharing policy across shortages and surpluses. This is a temporary MOOX policy, not an original-fidelity claim. It is isolated in `allocateFoodImports` / `allocateFoodExports` so a proven original priority can replace it without changing state, events or API shapes.

## Starvation and natural growth

Food/Production shortage is no longer modeled as merely disabling growth. The current formula uses the documented shape:

```text
natural_growth = base_growth * race_growth_multiplier
starvation_penalty = unresolved_shortage * starvation_factor
net_population_change = natural_growth - starvation_penalty
```

For ordinary population:

```text
starvation_penalty = food_shortage * 0.05
```

For Cybernetic population:

```text
starvation_penalty = food_shortage * 0.025
                   + production_shortage * 0.025
```

Positive net change becomes `projected_growth`. Negative net change becomes `projected_starvation`.

Starvation scales the aggregate farmer/worker/scientist allocation proportionally, preserving the Population assignment invariant. The final Population is clamped to at least `1.0` by the currently normalized classic/default starvation floor.

Positive change emits:

```text
colony.population_grew
```

Negative change emits:

```text
colony.population_starved
```

## Logistics event / Observer surface

A non-trivial logistics pass emits:

```text
empire.food_logistics_resolved
```

The event contains:

- Empire ID;
- the complete turn logistics snapshot;
- stable per-Colony import/export/surplus/shortage entries.

No event is emitted for an empire with no Food surplus, shortage or transfer, avoiding replay noise for a perfectly locally balanced empire.

The authoritative state still materializes the post-Population next-state Food logistics preview after growth/starvation, but that preview does not create a second authoritative logistics event for the same turn.

## BC accounting boundary

This slice calculates, but does **not** yet settle into a Treasury:

```text
freighter_operating_cost_bc
surplus_food_income_bc
```

That is deliberate. The runtime does not yet have the complete empire Treasury/maintenance ledger. Inventing a partial balance mutation here would mix Food logistics with the later Money/maintenance subsystem.

The values are authoritative materialized inputs ready for that ledger.

## Freighter acquisition boundary

The normalized data records:

```text
Freighter Fleet = 5 Freighters
Freighter Fleet cost = 50 PP
```

Freighter Fleet acquisition is now implemented through the shared Construction authority model. `GameSession.ConstructionChoices` exposes `project_kind = freighter_fleet` with the normalized 50 PP cost and +5 Freighters result; `colony.queue_freighter_fleet` is server-validated, progresses with the same pre-growth PP snapshot as buildings, and completion increments `Empire.Freighters` by five with Observer/replay events. MOOX currently gates this choice on normalized Technology 69 (`freighters`), which is the semantic unlock identity used by the runtime; this checkpoint does not claim that the separate original production-UI branch for that gate was newly isolated in disassembly.

## Blockades and population transport

Not implemented in this checkpoint:

- blockade prevention of Food transport;
- blockade production/Food penalties;
- using Freighters for Population transport;
- travel time for Population transport;
- competition between Food and Population transport for the same Freighter pool.

Those need strategic movement/blockade state before they can be represented correctly.

## Resolved MOO2 1.31 materialize/apply order

The previously open conflict is resolved by direct analysis of the original executable. Food/import calculations belong to the pre-apply materialized colony snapshot. MOOX therefore materializes local Economy and Food/Freighter logistics at the strategic-resolution boundary, then applies:

```text
Research using pre-growth RP
-> Population Growth/Starvation
-> Construction using pre-growth PP
-> next-state Economy/Food-logistics recalculation
```

This matches the original control/data-flow semantics even though MOOX does not continuously recalculate UI state between commands. See `TURN_ORDER_2026-08-28.md`.
## Ruleset schema

This slice introduces:

```text
EconomySchemaVersion = 5
StateSchemaVersion   = 5
```

`economy.json` adds a `food_logistics` block with explicit source IDs.

## Tests / regression coverage

Coverage now includes:

- two-colony Food export/import with enough Freighters;
- `2 Food -> 2 Freighters -> 1 BC` operating-cost example;
- insufficient Freighters leaving unresolved Food shortage;
- unresolved shortage producing deterministic starvation;
- last-Population starvation floor;
- Cybernetic Food + Production shortage penalty;
- surplus-Food valuation;
- Fantastic Traders doubled surplus-Food valuation;
- state rejection for negative Freighters, NaN logistics and impossible import amounts;
- prior fractional Economy/Construction/Research tests updated to distinguish command-time output from the later Population transition.

## Next exact work

Food/Freighter/Starvation and the strategic turn-order question are closed for this pass. Remaining Economy follow-ups stay parked: Treasury settlement, blockade/population transport, special growth/capacity layers and exact insufficient-Freighter allocation priority.

Active Research work moves to exact Uncreative application RNG/initialization timing where practical, then hyper-advanced repeated-field state/cost and Advanced-start ownership.