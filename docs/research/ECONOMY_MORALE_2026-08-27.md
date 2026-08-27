# Colony economy morale - 2026-08-27

Status: implemented Phase 1 morale-context checkpoint.

This checkpoint adds the first local colony morale layer on top of the previously implemented base, gravity and starting-government economy context.

## Original HELP.LBX evidence

The private MOO2 1.31 `HELP.LBX` reference is the primary evidence for the rules implemented here.

```text
HELP.LBX SHA-256:
78f8f0b6b6d05426de5b36116fbc8d1e4863305ad3fb659ae0ad3d3bbe83d7dd

block 0 SHA-256:
6fa9a4c1b844a4688c3fc01ad66c8b364c0291525a7c7098f6a918b7132399c9
```

The original help data establishes:

- Feudal colonies suffer -20% morale without Marine Barracks or Armor Barracks;
- Dictatorship colonies suffer -20% morale without Marine Barracks or Armor Barracks;
- Marine Barracks removes that government morale penalty;
- Armor Barracks removes that government morale penalty;
- Holo Simulator increases local colony morale by +20%;
- Pleasure Dome increases local colony morale by +30%;
- Pleasure Dome is cumulative with Holo Simulator;
- Unification ignores morale effects and morale buildings;
- morale display levels change total colony output.

The current implementation uses only those local-colony effects.

## Normalized rules

`economy.json` schema version 3 now includes:

```text
barracks penalty: -20%
applicable governments:
    government_feudal
    government_dictatorship

penalty-removing buildings:
    marine_barracks
    armor_barracks

local morale buildings:
    holo_simulator  +20%
    pleasure_dome   +30%
```

All referenced building IDs are cross-validated against committed `buildings.json` when the game economy rules are loaded.

## Colony state

`Colony` now carries a minimal building-inventory list:

```go
Buildings []string
```

The list is intentionally only state at this checkpoint. Construction cost, maintenance, technology prerequisites and general building effects are not yet driven by it.

Core validation rejects empty and duplicate building IDs. Ruleset-aware economy loading rejects morale-rule building IDs that do not exist in the normalized building catalog, and contextual calculation rejects unknown buildings on a colony.

## Morale context

`ColonyEconomyContext` now records:

```text
morale_barracks_penalty_percent
morale_building_bonus_percent
morale_percent
```

The split is deliberate. For example, a Unification colony with Holo Simulator + Pleasure Dome records a +50% building contribution for Observer/debugging, but the applied `morale_percent` is 0 because Unification ignores morale.

## Output composition

Food, production and research continue to be recomputed from the unmodified base snapshot:

```text
adjusted percent =
    100
  + government percent
  + applied morale percent
  - gravity penalty percent

adjusted output = ROUND(base output * adjusted percent / 100)
```

This avoids chained multiplication against already-rounded intermediate values.

Population income keeps separate rounded terms:

```text
adjusted income =
    base population income
  + rounded government money bonus
  + rounded signed morale money bonus
```

Democracy's government tax bonus still uses the already-normalized down-rounding rule. Morale's positive or negative share is independently rounded to whole BC before addition.

## Implemented examples

Regression fixtures cover:

- Psilon / Dictatorship / matching Low-G: 5 RP base -> 4 RP with missing-Barracks -20% morale;
- the same colony with Marine Barracks: morale penalty removed -> 5 RP;
- Human / Democracy with Holo Simulator + Pleasure Dome: +50% local morale is additive with +50% Democracy research;
- Unification with Holo Simulator + Pleasure Dome: the +50% building contribution is visible in context but not applied;
- unknown building IDs rejected by contextual economy;
- duplicate/empty colony building IDs rejected by core validation;
- full GameSession -> resolver -> domain event -> Observer path includes resolved morale context.

## Explicitly deferred morale sources

The original data also exposes additional morale systems which are not part of this narrow checkpoint:

- Virtual Reality Network empire-wide morale bonus;
- Telepathic Training empire-wide government-dependent morale bonus;
- Capitol-loss morale consequences;
- leaders;
- conquered population / assimilation morale;
- other empire-wide technology effects.

These require empire technology/state or race-aware population cohorts and should not be faked through local building state.

## Population-model limitation

Original MOO2 can retain population of multiple races after conquest. Morale, gravity and racial economy behavior can therefore differ by population cohort.

The current aggregate `PopulationState` still assumes the owning empire's preset race for all assigned population. Race-aware cohorts and persisted custom-race traits remain required before conquest/assimilation is implemented faithfully.

## Next narrow slice

The next useful runtime step is to connect normalized building/technology effects to colony output and construction/research progression while preserving the explicit base/context/adjusted split. Empire-wide morale technology and conquest should wait until the required empire-tech and population-cohort state exists.