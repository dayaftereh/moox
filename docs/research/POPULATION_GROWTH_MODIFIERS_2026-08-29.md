# Population growth building / medicine modifiers - 2026-08-29

**Open marker:** `docs/slices/_OPEN_POPULATION_GROWTH_MODIFIERS_2026-08-29.md`

## Goal

Determine original Master of Orion II 1.31 Population-growth behavior for Housing, Cloning Center and medicine/Technology modifiers before extending MOOX's existing classic capacity-limited growth curve.

## Gate 1 - Checkup + original analysis / reverse engineering

**Status:** complete. Gameplay implementation has not started.

### Recovery state

- Starting branch: `main`.
- Starting HEAD: `b873d28` (`docs: reconcile completed slice status`).
- Starting tree before this slice: clean, `main` ahead of `origin/main` by 6 commits.
- Previous gameplay slice: Population relocation through shared Freighters (`f816eaf`), closed.
- Core schema at slice start: 11.

## Existing MOOX baseline

The current runtime keeps the classic natural curve isolated in `EconomyRules.refreshPopulationProjection`:

```text
base_growth = sqrt(0.002 * population * (capacity - population) / capacity)
natural_growth = base_growth * race_growth_multiplier
```

Population is domain-native `float64`; MOOX intentionally does not reproduce the original integer/k-pop storage representation. Cybernetic Production sustenance is already materialized as:

```text
production_required  = population * 0.5
production_available = max(0, adjusted_production - production_required)
```

`cloning_center` already exists as a normalized Building identity and Technology ownership already exists authoritatively through `Empire.KnownTechnologyIDs`.

## Original executable provenance

Private original reference used for this gate:

```text
reference/original/support/files/Orion2.exe
SHA-256 7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5
```

The bound LE image has executable object 1 at relocation base `0x10000` and data object 2 at relocation base `0x178000`. Embedded original symbols identify the relevant paths directly:

```text
Colony_Pop_Grows_                VA 0xE1839
Apply_Colony_Pop_Growth_         VA 0xE2DCA
Housing_Is_Useless_              VA 0xE36B7
Planet_Max_Population_For_Player_ VA 0xE0B4F
Colony_Race_Pop_Limit_           VA 0xE0C1D
Colony_Officer_                  VA 0xDD9F2
Event_Check_Plague_              VA 0x234B8
Event_Check_Population_Boom_     VA 0x23509
Pre_Import_Computing_            VA 0xE1D59
Colony_Industry_Production_      VA 0xDEE1B
```

No original executable/data bytes are added to MOOX by this slice.

## Direct original growth formula

**Evidence level: original-observed / original-derived.**

`Colony_Pop_Grows_` computes growth per original Population/race cohort. For a single cohort, the core path is:

```text
curve_input = ((capacity - population) * population * 2000) / capacity
natural_kpop = integer_sqrt(curve_input) * growth_percent / 100
turn_kpop = natural_kpop + flat_additive_kpop
```

The helper at `0x134C92` is directly verified as an integer square-root implementation: it binary-searches for the largest integer whose square does not exceed the input.

`Apply_Colony_Pop_Growth_` then adds the turn delta to a per-cohort remainder. At `1000` (`0x3E8`) accumulated units it creates one Population entry and subtracts `1000`. This proves the original intermediate is 1/1000 Population.

Therefore MOOX's existing continuous formula is the algebraic direct-Population form of the original natural curve:

```text
integer original: sqrt(pop * (cap-pop) * 2000 / cap) / 1000
continuous MOOX : sqrt(0.002 * pop * (cap-pop) / cap)
```

MOOX's prior decision not to copy integer truncation around this curve remains consistent with the project's domain-native numeric architecture.

## Race Growth modifier

The original Race block byte is directly consumed as:

```text
player + 0x8A0 = signed Population-growth trait delta
base growth_percent = 100 + signed(player+0x8A0)
```

This matches the already normalized MOOX selections:

```text
Growth -50%  ->  50%
normal       -> 100%
Growth +50%  -> 150%
Growth +100% -> 200%
```

## Medicine / Technology growth modifiers

Original Technology ownership/status occupies 203 bytes beginning at `player+0x117`; status `3` is owned. `Colony_Pop_Grows_` tests two exact entries:

```text
player+0x182 = player+0x117+107
Technology 107 = Microbiotics
owned/status 3 -> growth_percent += 25

player+0x1D8 = player+0x117+193
Technology 193 = Universal Antidote
owned/status 3 -> growth_percent += 50
```

Both IDs and semantic keys already exist in committed `technologies.json`:

```text
107  microbiotics
193  universal_antidote
```

### Result

Microbiotics and Universal Antidote are **additive percentage points on the same natural-growth multiplier**. They are not flat Population additions and are not chained multiplicatively.

Examples:

```text
normal + Microbiotics                 = 1.25x natural growth
normal + both medicine Technologies   = 1.75x natural growth
Growth +50% race + both Technologies = 2.25x natural growth
```

## Housing

### Original identity

The special current-production sentinel is:

```text
colony + 0x115 == -3
```

The `-3` production path performs the same max-Population test as the separately named `Housing_Is_Useless_` routine: once current Population equals `Planet_Max_Population_For_Player_`, the special project is no longer useful and the current project is cleared. Together with the named routine, this establishes `-3` as the original Housing production mode.

Housing is therefore a **continuous colony Production project**, not a Building.

### Production term

While Housing is active, `Colony_Pop_Grows_` adds:

```text
growth_percent += ((colony[+0xE9] - colony[+0xF0]) * 40) / population
```

`colony+0xE9` is already directly established in `TURN_ORDER_2026-08-28.md` as the materialized current-turn Industry/PP snapshot.

`colony+0xF0` is now directly resolved from the pre-calculation path. The routine immediately before `Colony_Industry_Production_` scans Population entries, tests each race's `player+0x8B0` Cybernetic flag, and stores:

```text
colony[+0xF0] = ceil(cybernetic_population / 2)
```

It is therefore the original whole-PP Cybernetic Production sustenance requirement. Housing uses **Production remaining after Cybernetic sustenance**.

MOOX already owns the corresponding semantic value as `ColonyPopulationDynamics.ProductionAvailable`; its fractional `0.5 PP / Population` representation deliberately avoids the original whole-PP storage rounding.

### Result

For the current single-race MOOX model the Housing contribution is:

```text
housing_bonus_percentage_points = 40 * production_available / population
```

The original performs integer division at this percentage-point boundary. Gate 2 should explicitly choose whether to preserve that observable floor or continue MOOX's continuous-domain convention here. The recommended fidelity choice is to retain the original percentage-point floor while continuing to use MOOX's domain-native `production_available` input.

Housing changes the **natural-growth percentage**; it is not a flat +Population/turn effect.

## Cloning Center

### Original identity

`Add_Building_` stores normal Building ownership as:

```text
colony[0x136 + building_production_id] = 1
```

The original-derived normalized Building table already proves:

```text
cloning_center
production_id = 10
technology_id = 39
cost          = 100 PP
maintenance   = 2 BC
```

Therefore Cloning Center ownership is exactly `colony+0x140` (`0x136 + 10`). `Colony_Pop_Grows_` directly tests that byte.

### Original effect

If Cloning Center is present and the Colony still has room, its contribution is placed in the separate additive growth component, not in `growth_percent`:

```text
cohort_flat_kpop += cohort_population * 100 / total_population
```

Across all cohorts that is 100 k-pop growth units per Colony per turn. `Apply_Colony_Pop_Growth_` proves 1000 units = 1 Population, therefore:

```text
Cloning Center = +0.1 Population / turn
```

The flat Cloning contribution is capacity-limited, but it is **not multiplied** by Race Growth, Housing, Microbiotics or Universal Antidote.

## Stacking/order for the current single-race MOOX model

Ignoring unrelated original event/leader modifiers that are outside this slice, the directly supported shape is:

```text
base_growth = sqrt(0.002 * population * (capacity - population) / capacity)

natural_multiplier =
    race_growth_multiplier
  + microbiotics_bonus        // +0.25 if known
  + universal_antidote_bonus  // +0.50 if known
  + housing_bonus             // Production-dependent, if Housing active

natural_growth = base_growth * natural_multiplier
cloning_growth = 0.1 if Cloning Center exists and population < capacity else 0

gross_growth = natural_growth + cloning_growth
```

The existing MOOX starvation/sustenance layer remains applied to the resulting turn projection. Final positive growth remains capped to `capacity - population`.

### Turn-order invariant

The already-proven strategic order remains unchanged:

```text
1. materialize current-turn Economy / PopulationDynamics
2. Research consumes pre-growth RP
3. Population Growth/Starvation applies
4. Construction consumes pre-growth PP
5. recalculate next-state Economy / PopulationDynamics
```

Housing therefore uses the **pre-growth Production snapshot**. Newly grown Population cannot generate additional Housing Production, RP or PP retroactively in the same turn.

## Other original growth modifiers observed but outside this slice

The original routine contains additional orthogonal modifiers. They are deliberately not folded into this slice:

- `Event_Check_Population_Boom_` adds +100 percentage points;
- `Event_Check_Plague_` subtracts 200 percentage points;
- Colony Officer paths add level-dependent modifiers;
- a separate relocated helper contributes another +150 percentage points under a condition not required to identify Housing, Cloning Center, Microbiotics or Universal Antidote;
- the `0x21CB0` non-default game-setting/NPC layer is already known from prior Advanced-start research and is not a medicine modifier.

The unresolved +150 source is therefore recorded as a future modifier-research item, but it does **not** block this slice: all requested Housing/Cloning/medicine paths are independently and directly identified.

## Minimal normalized/runtime identities required

No new Building or Technology identity is required:

- `cloning_center` already exists in `buildings.json`;
- `microbiotics` / Technology 107 already exists;
- `universal_antidote` / Technology 193 already exists.

Housing does require a semantic runtime production identity because it is not a Building. The clean shape is a new persistent Construction project kind:

```text
project_kind = housing
project_id   = housing
```

This avoids encoding the original `-3` sentinel in MOOX state or protocol.

## Proposed Gate 2 implementation shape

No gameplay code has been changed yet. Proposed implementation:

1. **Ruleset**
   - add original-proven Population-growth effect metadata to `economy.json`;
   - Housing coefficient: `40` percentage points per available PP divided by Population;
   - Cloning Center flat growth: `0.1 Population/turn`;
   - Technology modifiers keyed semantically to `microbiotics` (+0.25) and `universal_antidote` (+0.50);
   - validate referenced Building/Technology identities during `LoadEconomyRules`.

2. **Core/state**
   - add `ConstructionProjectHousing = "housing"`;
   - use project ID `housing`;
   - because a Housing project can now be persisted, advance `StateSchemaVersion` from 11 to 12 and add exact round-trip/validation coverage.

3. **Authoritative command / choices**
   - add `colony.queue_housing`;
   - expose Housing through `AvailableConstructionChoices` only while the owned Colony has remaining Population capacity;
   - Housing has no finite PP cost, progress accumulator or completion event;
   - when Population reaches capacity, clear Housing automatically, mirroring the original full-colony behavior.

4. **Population projection**
   - compute medicine bonuses from authoritative Empire Technology ownership;
   - compute Cloning Center from authoritative Colony Building ownership;
   - compute Housing from the active authoritative Construction project and pre-growth `ProductionAvailable`;
   - add Race + medicine + Housing as percentage points on the natural curve;
   - add Cloning Center as flat +0.1 after the natural component;
   - preserve existing starvation and capacity clamping.

5. **Construction phase**
   - Housing consumes the Colony's production mode for the turn but never accumulates construction progress;
   - do not let the same PP simultaneously advance another project;
   - preserve the existing Research -> Population -> Construction ordering and pre-growth PP snapshot.

## Proposed deterministic regression coverage

### Core

- schema-12 Housing Construction state round-trips exactly;
- invalid Housing project identity is rejected.

### Game

- existing no-modifier natural-growth fixture remains unchanged;
- Microbiotics alone produces +25 percentage points;
- Universal Antidote alone produces +50 percentage points;
- both stack additively;
- Growth +50% race + both medicine Technologies produces 2.25x natural growth, not chained multiplication;
- Cloning Center adds exactly +0.1 and is not multiplied by Race/medicine/Housing;
- Cloning does not exceed capacity;
- Housing uses pre-growth Production after Cybernetic sustenance;
- Housing coefficient/rounding matches the accepted Gate 2 rule;
- Housing is unavailable/useless at full capacity and auto-clears when the Colony becomes full;
- combined Race + medicine + Housing + Cloning regression locks stacking order;
- fresh Population does not contribute RP/PP/Housing Production retroactively in the same turn.

### Session / Observer

- only the owning seat can queue Housing;
- construction choices expose Housing authoritatively;
- Observer/replay state shows the Housing project and resulting growth lifecycle without moving legality to adapters.

## Gate 1 conclusion

All five Gate 1 questions are closed for the requested scope. The implementation can proceed after Gate 2 acceptance of the semantic design above, especially the one deliberate numeric choice: preserve the original integer percentage-point floor for Housing versus using fully continuous MOOX arithmetic.

## Gate 2 - implementation decision

Accepted on 2026-08-29. The implementation keeps the original integer percentage-point floor for Housing while retaining MOOX domain-native `float64` Population/Production state.

## Gate 3 - implementation

**Status:** complete.

Implemented runtime shape:

- Core `StateSchemaVersion` advanced from 11 to 12.
- Added semantic persistent `ConstructionProjectHousing` / `project_id = "housing"`; Housing state cannot accumulate PP progress.
- Added `colony.queue_housing` with authoritative ownership/capacity validation.
- `AvailableConstructionChoices` exposes Housing only while the Colony has remaining Population capacity.
- Housing remains active across turns without accumulating Construction progress and automatically clears when the Colony reaches capacity.
- `economy.json` schema advanced to 6 and now stores direct-original Population-growth modifier metadata with `Orion2.exe` provenance.
- `LoadEconomyRules` cross-validates the normalized `cloning_center`, `microbiotics`, and `universal_antidote` identities against the committed Building/Technology catalogs.
- Population projection now uses authoritative Empire Technology ownership, Colony Building ownership and active Housing Construction state.
- Natural multiplier stacking is Race + Microbiotics + Universal Antidote + Housing percentage points.
- Cloning Center is applied afterwards as flat `+0.1 Population/turn` and is never multiplied by the natural-growth modifiers.
- Housing reads the materialized pre-growth `ProductionAvailable`, so Cybernetic Production sustenance is deducted before Housing and freshly grown Population cannot contribute PP to its own turn.
- Existing Research -> Population -> Construction ordering remains unchanged.

Deterministic regression coverage added:

- Core schema-12 Housing state round-trip and invalid Housing identity/progress rejection;
- Microbiotics +25 percentage points and Universal Antidote +50 percentage points;
- additive medicine stacking;
- original Housing floor plus Cybernetic Production-available input;
- Cloning Center flat-growth/non-multiplication behavior;
- combined Race + medicine + Housing + Cloning stacking and Population-capacity clamp;
- Housing legal-action, queue and automatic full-capacity stop;
- GameSession seat authority plus Observer-visible Housing lifecycle.

## Gate 4 - QA

The implementation passed:

```text
gofmt on all changed Go files
git diff --check
go test ./...
go vet ./...
```

No temporary original-executable extraction remains in the working environment.
