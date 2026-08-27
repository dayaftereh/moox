# Colony economy context - gravity and government - 2026-08-27

Status: implemented Phase 1 contextual-economy checkpoint.

This checkpoint layers the first verified contextual modifiers over the base colony role output described in `ECONOMY_BASELINE_2026-08-27.md`.

The runtime now deliberately keeps three separate representations on every colony:

```text
Economy          base role output
EconomyContext   currently applied contextual coefficients
AdjustedEconomy  result recomputed from Economy + EconomyContext
```

This separation is important. Later morale, leaders, buildings and other bonuses must be added to the context and recomputed from the base snapshot rather than multiplied sequentially onto already-rounded results.

## Original HELP.LBX evidence

The local private MOO2 1.31 reference `HELP.LBX` was inspected directly.

Verified source hashes:

```text
HELP.LBX SHA-256:
78f8f0b6b6d05426de5b36116fbc8d1e4863305ad3fb659ae0ad3d3bbe83d7dd

block 0 SHA-256:
6fa9a4c1b844a4688c3fc01ad66c8b364c0291525a7c7098f6a918b7132399c9
```

Original help text establishes the starting-government economy effects used here:

- Feudal: scientist research reduced by 50%;
- Dictatorship: no direct food/industry/research/tax percentage in this slice;
- Democracy: scientist research +50%, tax revenues +50%;
- Unification: food +50%, industrial production +50%, morale effects/buildings ignored.

The same original help data establishes gravity compatibility qualitatively:

- Low-G race: Low-G no penalty, Normal-G gets half the Heavy-G penalty;
- Heavy-G race: Normal-G and Heavy-G no penalty;
- Gravity affects colony worker productivity.

The exact effective 25%/50% matrix is cross-checked against StrategyWiki and stored as secondary-reference evidence until a direct executable parity fixture is added.

## Gravity matrix

The normalized matrix in `economy.json` is:

```text
race gravity   planet Low-G   planet Normal-G   planet Heavy-G
----------------------------------------------------------------
Low-G               0%              -25%              -50%
Normal-G           -25%               0%              -50%
Heavy-G            -25%               0%                0%
```

The penalty applies to farmer, worker and scientist role output. It does not apply to base population tax income.

The High-G-on-Low-G value is intentionally 25%. Community/reference evidence notes that one planet UI display can misleadingly show 50%, while effective production behaves as 25%.

### Gravity Generator limitation

MOO2's Gravity Generator removes wrong-gravity penalties. Master of Orion X does not yet have colony building inventory/effects, so this override is not applied yet. Once building state exists, the gravity layer must first determine whether an active Gravity Generator nullifies the penalty.

## Starting-government context

`economy.json` schema 2 contains the four starting governments:

```text
Feudal
    research     -50%

Dictatorship
    no direct percentage modifiers in this slice

Democracy
    research     +50%
    tax revenue  +50%

Unification
    food         +50%
    production   +50%
    ignores morale
```

Only the effects needed by the current colony-economy pipeline are normalized. Ship-cost, espionage, assimilation and other government behavior remains outside this slice.

### Percentage composition

For food, production and research the current contextual layer follows the documented additive structure:

```text
adjusted percent = 100 + government percent - gravity penalty
adjusted role output = ROUND(base role output * adjusted percent / 100)
```

The base role output already includes planet coefficient and racial role delta. Later morale and leader percentages belong in this same additive context rather than as chained multiplications.

### Democracy tax rounding

Population income is first materialized as base whole BC. Democracy then adds its +50% government bonus as a separate money bonus, rounded down to whole BC before being added to base income.

The current implementation covers only base population income. MOO2 also includes special colony income (for example gold/gems) in the government tax-bonus base; planet specials are not yet represented in this economy state and therefore remain a known gap.

## Race context derived from normalized presets

The runtime derives a preset race's economy context from `races.json` + `race_traits.json`:

- no explicit gravity trait -> `normal_g`;
- `low_g_world` -> `low_g`;
- `high_g_world` -> `heavy_g`;
- one of the four normalized starting-government abilities;
- Aquatic and numeric farming/industry/science/money traits continue to feed the base layer.

The context is not accepted from a player command.

## Materialized colony state

`ColonyEconomyContext` currently records:

```text
race_gravity_id
planet_gravity_id
gravity_penalty_percent
government_trait_id
government_food_percent
government_production_percent
government_research_percent
government_tax_percent
government_ignores_morale
```

`AdjustedEconomy` stores the current result after these context layers. `Economy` remains the pre-context base snapshot.

A successful `colony.assign_population` event includes all three pieces so the Observer can explain why the effective output differs from the base output.

## Tested examples

Regression coverage includes:

- Human / Democracy / Normal-G: research and tax +50%;
- Psilon / Low-G on Normal-G and Heavy-G: 25% / 50% penalties;
- Bulrathi / Heavy-G on Low-G: 25% penalty;
- Klackon / Unification: +50% food and industry, morale-ignore flag;
- Sakkra / Feudal: -50% research;
- Player/Observer end-to-end projection of base, context and adjusted economy.

## Important model limitations

### Mixed-race colonies

Original MOO2 colonies can contain conquered population of different races. Gravity, racial production and other traits can therefore differ by population cohort.

The current Phase 1 `PopulationState` is aggregate and assumes the owning empire's preset race for all assigned workers. A future fidelity step needs population cohorts with race identity before conquest/assimilation can be modeled correctly.

### Custom races

`Empire` currently stores a preset `RaceID`. The original game supports custom race designs. A future empire/race configuration must persist the actual selected trait set so economy context is not tied only to the 13 preset races.

### Morale and buildings

Non-Unification morale is not yet applied. Unification already records `government_ignores_morale=true` so the later morale layer has an explicit bypass.

Buildings/technology, pollution, blockade, leaders, food consumption/logistics and maintenance remain outside this context checkpoint.

## Next narrow slice

The next economy layer should normalize morale calculation and the first building/technology coefficients needed to test it, without mixing in pollution or logistics yet. Before adding conquered/mixed population behavior, the colony population model should be split into race-aware cohorts.