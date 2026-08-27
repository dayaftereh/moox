# Population growth and sustenance - 2026-08-27

Status: implemented deterministic Phase 1 population-growth/sustenance slice.

This checkpoint extends the domain-native `float64` economy from job output into Population capacity, Food/Production sustenance and turn-end Population growth. It intentionally does not implement the complete MOO2 population/logistics system yet.

## Evidence used

The normalized implementation is based on three secondary references, all recorded in `data/rulesets/moo2-1.31/economy.json`:

- `moo2-150-population-growth` - MOO2 1.50 technical manual, documenting the classic square-root Population growth curve, factor 2000 in the original population-thousand representation, race growth modifiers, Housing and Cloning hooks.
- `moo2-online-population-capacity` - MOO2 formula reference for planet-size base capacities, climate habitability, Aquatic/Tolerant reinterpretation and Subterranean capacity bonus.
- `strategywiki-moo2-sustenance` - cross-check for normal Food consumption, Cybernetic Food/Production consumption and Lithovore Food independence.

The original-observed race-trait normalization already provides the `population_growth_multiplier` selections and the Aquatic, Tolerant, Subterranean, Cybernetic and Lithovore abilities.

## Domain-native MOOX representation

MOOX does not reproduce the classic integer/k-pop growth intermediate. Population remains direct `float64` state:

```text
Population.Total = 4.08
```

not:

```text
Population = 4080
```

The classic curve is algebraically expressed directly in Population units:

```text
base_growth = sqrt(0.002 * population * (capacity - population) / capacity)
projected_growth = min(capacity - population, base_growth * race_growth_multiplier)
```

`0.002` is the direct-Population-unit counterpart of the classic `FACTOR1 = 2000` formulation after converting its growth result from thousands of population into full Population units. No integer truncation is introduced around the curve.

Race growth factors currently materialize from normalized race traits:

- normal: `1.0`
- Growth -50%: `0.5`
- Growth +50%: `1.5`
- Growth +100%: `2.0`

## Population capacity

The current normalized base capacity by planet size is:

```text
tiny     5
small   10
medium  15
large   20
huge    25
```

Climate habitability fractions are:

```text
toxic     0.25
radiated  0.25
barren    0.25
desert    0.25
tundra    0.25
ocean     0.25
swamp     0.40
arid      0.60
terran    0.80
gaia      1.00
```

Capacity is the one place in this slice where classic whole-Population rounding is deliberately retained as an explicit gameplay-rule boundary:

```text
capacity = round(size_capacity * habitability)
```

Examples:

- Human / Medium / Terran: `round(15 * 0.8) = 12`
- Human / Medium / Barren: `round(15 * 0.25) = 4`

Race effects currently implemented:

- Aquatic: Tundra/Swamp become at least 80%; Ocean/Terran/Gaia become 100%.
- Tolerant: +25 percentage points habitability, capped at 100%.
- Subterranean: +2 Population per size class: Tiny +2 through Huge +10.

The old implementation ceiling is not copied into MOOX. Capacity is determined by gameplay rules, not by a legacy storage limit.

## Sustenance

The materialized `ColonyPopulationDynamics` contains:

```text
capacity
food_required
food_surplus
food_shortage
production_required
production_shortage
production_available
base_growth
growth_multiplier
projected_growth
```

Current sustenance rules:

- normal population: `1 Food / Population`
- Lithovore: `0 Food / Population`
- Cybernetic: `0.5 Food + 0.5 PP / Population`

For Cybernetic colonies the Production requirement is removed before Construction sees available PP:

```text
production_available = max(0, adjusted_production - production_required)
```

This preserves the authoritative separation between gross/adjusted colony output and the PP that is actually available to Construction.

## Turn ordering

The current strategic resolution order is intentionally:

```text
1. Commands / population assignment / research selection / construction selection
2. Recalculate current-turn Colony economy + PopulationDynamics
3. Construction consumes current-turn production_available
4. Research consumes current-turn Research
5. Population growth uses current-turn sustenance and projected_growth
6. Recalculate the post-turn snapshot from the new Population
```

This prevents newly grown Population from retroactively producing PP or RP in the same turn.

A dedicated regression test proves that an exactly-fed Human colony can grow at turn end while its selected Research project still receives only the pre-growth `4.5 RP`; the larger scientist output appears only in the post-turn snapshot.

## Population job allocation after growth

Population growth is continuous, so new Population is not left as an unassigned whole worker icon. The current aggregate model preserves job shares proportionally:

```text
before:
Total       4.00
Farmers     1.25
Workers     1.50
Scientists  1.25

+0.08 growth

after:
Total       4.08
jobs scaled proportionally
```

The final scientist value is calculated as the remainder after Farmers/Workers so the authoritative invariant remains:

```text
Farmers + Workers + Scientists == Total
```

within deterministic floating-point tolerance.

This proportional policy is a MOOX choice for the current aggregate-population model, not a claim that the original UI assigned fractional worker icons this way.

## Event / Observer surface

Turn-end positive growth emits:

```text
colony.population_grew
```

with:

- Colony ID,
- previous Population state,
- current Population state,
- applied growth,
- the PopulationDynamics snapshot that produced the growth.

`PopulationDynamics` is persisted on the Colony and therefore visible through the authoritative Observer state and the owning Player projection wherever the Colony itself is exposed.

## State and ruleset schema

This slice introduces:

- `EconomySchemaVersion = 4` for the normalized population rules.
- `StateSchemaVersion = 4` for persisted `population_dynamics`.

The domain-native Float decision from ADR-0003 remains unchanged; schema 4 extends the state with derived Population dynamics rather than reintroducing scaled integers.

## Deliberately deferred

The following remain separate research/implementation slices:

- starvation Population loss and the exact original starvation quirks,
- empire Food logistics / Freighters,
- sale/disposal of Food surplus,
- Housing production-to-growth conversion,
- Cloning Center bonus,
- medicine-technology growth modifiers,
- Biospheres / Advanced City Planning / other capacity buildings,
- terraforming capacity transitions,
- multi-race Population cohorts, conquest and assimilation,
- custom-race runtime composition beyond the currently normalized preset-race traits.

These should be added one proven rule at a time rather than folded into the base growth equation implicitly.

## Research lane remains active

This Economy checkpoint does not replace the active Research backlog. The next Research-specific slice remains the multi-Technology TechField policy:

1. determine Creative/Uncreative acquisition/choice semantics,
2. encode that policy in server-authoritative `ResearchChoice` / legal actions,
3. keep clients from selecting arbitrary Technology IDs,
4. then address active-project switching and hyper-advanced repeated fields.

Human UI, built-in AI and future external agents must continue to use the same authority-filtered Research action surface.


## Superseding logistics checkpoint

Starvation, empire Food/Freighter logistics and surplus-Food valuation were implemented in FOOD_FREIGHTER_LOGISTICS_2026-08-27.md, which bumps the current development state/ruleset to schema 5. Schema 4 above remains the historical boundary introduced by this Population Growth checkpoint.
