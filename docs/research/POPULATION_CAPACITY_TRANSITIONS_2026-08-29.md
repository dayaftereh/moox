# Population capacity transitions - 2026-08-29

**Open marker:** `docs/slices/_OPEN_POPULATION_CAPACITY_TRANSITIONS_2026-08-29.md`

## Goal

Determine original Master of Orion II 1.31 behavior for Biospheres, Advanced City Planning and terraforming/climate-driven Population-capacity transitions before extending MOOX's current race-aware capacity calculation.

## Gate 1 - Checkup + original analysis / reverse engineering

**Status:** in progress. Gameplay implementation has not started.

### Recovery state

- Starting branch: `main`.
- Starting HEAD: `f9f4792` (`docs: close population growth modifier slice`).
- Starting tree: clean, `main` ahead of `origin/main` by 8 commits.
- Previous gameplay slice: Population growth building / medicine modifiers (`e7f0d72`), closed.
- Core `StateSchemaVersion`: 12.
- Economy ruleset schema: 6.

### Gate 1 target questions

1. Biospheres capacity delta, applicability and ordering.
2. Advanced City Planning capacity delta and Technology semantics.
3. Terraforming/climate transitions and their effect on current/max Population.
4. Ordering/stacking with Aquatic, Tolerant and Subterranean capacity rules.
5. Behavior when a capacity change would place current Population above the new cap.
6. Minimal semantic runtime/data identities and deterministic test shape.

## Existing MOOX baseline

The current runtime owns physical planet climate as `core.Planet.ClimateID`, explicit one-based `Planet.Orbit`, normalized planet size/climate capacity data, authoritative Empire Technology ownership, and Colony Building ownership. The current `EconomyRules.PopulationCapacity(planet, raceID)` implements only the planet/race portion:

```text
physical climate
-> Aquatic capacity adjustment
-> Tolerant habitability adjustment
-> round(size capacity * habitability)
-> Subterranean flat size-class bonus
```

It does not yet include Advanced City Planning or Biospheres. `terraforming` and `gaia_transformation` are currently normalized production/building-table identities, but the generic MOOX Building construction path would incorrectly persist them in `Colony.Buildings` if used unchanged.

## Original executable provenance

Private original reference used for this gate:

```text
reference/original/support/files/Orion2.exe
SHA-256 7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5
```

Relevant embedded original symbols / directly traced routines:

```text
Size_Subterranean_Pop_Bonus_          VA 0xE0A14
Player_Effective_Climate_             VA 0xE0A49
planet/race capacity helper           VA 0xE0A93
Planet_Max_Population_For_Player_     VA 0xE0B4F
Colony_Race_Pop_Limit_                VA 0xE0C1D
Add_Building_                         VA 0x13FD9
Remove_Building_                      VA 0x145EA
building/product legality helper      VA 0xE11BC
population-over-cap trimming helper   VA 0xEC97C
Random_                               VA 0x1247A0
```

The two temporary LE object extracts used for disassembly are not repository artifacts and are deleted at the end of Gate 1.

## Direct original capacity order

**Evidence level: original-observed / original-derived.**

The original capacity helper takes planet size, physical climate and player/race state. Its directly observed order is:

```text
1. effective_climate = Player_Effective_Climate_(player, physical_climate)
2. habitability = climate_modifier[effective_climate]
3. if Tolerant: habitability = min(100, habitability + 25)
4. base_capacity = (size_base * habitability + 50) / 100
   // explicit nearest-integer rounding boundary
5. if Subterranean: base_capacity += 2 * (size_index + 1)
6. if player owns Advanced City Planning: base_capacity += 5
7. for an actual Colony, if Biospheres exists: base_capacity += 2
```

This directly matches MOOX's existing size/climate/Tolerant/Subterranean arithmetic and supplies the two missing additive layers.

### Aquatic effective-climate mapping

`Player_Effective_Climate_` directly maps physical climates for an Aquatic race as follows:

```text
Tundra -> Terran
Ocean  -> Gaia
Swamp  -> Terran
Terran -> Gaia
other climates unchanged
```

For capacity this is equivalent to the currently implemented MOOX Aquatic habitability rules. Importantly, the Aquatic mapping occurs **before** Tolerant and before the Subterranean / Advanced City Planning / Biospheres flat additions.

## Advanced City Planning

Normalized identity already committed:

```text
technology key: advanced_city_planning
Technology ID: 3
TechField ID: 42
```

Both `Planet_Max_Population_For_Player_` and `Colony_Race_Pop_Limit_` test:

```text
player + 0x11A = player + 0x117 + Technology ID 3
owned/status 3 -> capacity += 5
```

### Result

Advanced City Planning is a **global passive +5 Population capacity per planet/Colony for the owning Empire**. It is not a Building and it is not multiplied by climate, Aquatic, Tolerant or Subterranean modifiers; it is added after those layers.

The planet-level routine applies it even without a Colony/Biospheres context, while `Colony_Race_Pop_Limit_` applies the same owner-Technology bonus before the Colony-only Biospheres bonus.

### Research-turn timing

The already-proven original materialize/apply pipeline matters here. `Colony_Pop_Grows_` is part of the calculation/materialization phase. Research breakthrough/Technology ownership is later applied in `Apply_All_Player_Changes_`, before Population **application**, but the current-turn Population delta was already calculated.

Therefore a newly completed Advanced City Planning Technology does **not** retroactively enlarge the already-materialized Population-growth delta of the breakthrough turn. The post-apply `Do_Colony_Calculations_` refresh sees the new +5 capacity for the next state/turn. This is consistent with MOOX's current pre-apply Population projection architecture.

## Biospheres

Normalized identity already committed:

```text
building key: biospheres
production ID: 15
Technology ID: 61 (biospheres)
cost: 60 PP
maintenance: 1 BC
```

Original Building ownership is stored beginning at `colony+0x136`. `Colony_Race_Pop_Limit_` directly tests:

```text
colony + 0x145 = colony + 0x136 + production ID 15
non-zero -> capacity += 2
```

### Result

Biospheres is an ordinary persistent Building with **flat +2 Population capacity**. It is added after effective climate, Tolerant rounding, Subterranean and Advanced City Planning. It therefore does not get multiplied by any of those earlier layers.

Because Construction is applied after Population growth, completing Biospheres cannot increase Population growth retroactively in its completion turn. The next-state Colony recalculation sees the +2 capacity.

## Terraforming

Normalized production identity already committed:

```text
project key currently normalized as building: terraforming
production ID: 44
Technology ID: 183 (terraforming)
cost: 250 PP
maintenance: 0 BC
```

### Original legality

The original product-legality helper at `0xE11BC` allows production ID 44 only while the **physical** planet climate index is 2 through 7 inclusive:

```text
2 Barren
3 Desert
4 Tundra
5 Ocean
6 Swamp
7 Arid
```

It is therefore unavailable on Toxic, Radiated, Terran or Gaia planets.

### Original completion mapping

`Add_Building_` special-cases production ID 44 and does **not** persist a Building flag. It mutates the planet's physical climate:

```text
Barren -> Desert or Tundra
Desert -> Arid
Tundra -> Swamp
Ocean  -> Terran
Swamp  -> Terran
Arid   -> Terran
```

The Barren case uses the planet's slot in the original five-planet system array (`system+0x4A ... +0x52`):

```text
first two orbit slots -> Desert
last two orbit slots  -> Tundra
middle orbit slot     -> Random_(2): Desert or Tundra
```

MOOX already has one-based `Planet.Orbit` and authoritative deterministic RNG state, so no new structural concept is required for this decision.

The original also updates auxiliary planet bytes from the climate-table delta and increments a transformation-related planet byte. Those auxiliary original storage fields are not required to represent the Population-capacity result in MOOX because current Food/capacity rules derive semantically from `ClimateID`; they should not be copied as opaque bytes without a separate proven need.

### Capacity timing / trait interaction

Terraforming changes the **physical** `ClimateID`. The subsequent capacity calculation then re-runs the normal chain:

```text
new physical climate
-> Aquatic effective climate
-> Tolerant habitability
-> rounded size/climate capacity
-> Subterranean
-> Advanced City Planning
-> Biospheres
```

For Aquatic races some physical improvements therefore legitimately produce no capacity change (for example Ocean -> Terran is Gaia-effective both before and after), while others do.

Terraforming completes in the original Construction/Production apply phase, after current-turn Population growth. The increased capacity becomes part of the post-apply next-state calculation; it does not create additional Population in the same turn.

## Gaia Transformation

Normalized production identity already committed:

```text
project key currently normalized as building: gaia_transformation
production ID: 17
Technology ID: 74 (gaia_transformation)
cost: 500 PP
maintenance: 0 BC
```

The original product-legality helper permits production ID 17 only when the physical climate is exactly index 8 = Terran.

`Add_Building_` special-cases production ID 17:

```text
Terran -> Gaia
```

It then removes any queued Terraforming production ID 44 and, like Terraforming, does **not** persist a normal Building flag. Gaia Transformation is therefore a one-shot planetary transformation project, not a persistent Colony Building.

For Aquatic capacity, physical Terran is already Gaia-effective, so Terran -> Gaia can leave Population capacity unchanged even though the physical climate and other climate-derived gameplay values change.

## Capacity decreases and immediate Population loss

**Direct original answer: a capacity decrease can immediately remove Population.**

`Remove_Building_` clears the Building flag first. Production ID 15 = Biospheres has a dedicated path that then invokes the helper at `0xEC97C`.

That helper:

1. scans the Colony's original Population entries by race/cohort;
2. calls `Colony_Race_Pop_Limit_` for each represented race;
3. counts entries against the newly reduced limit;
4. decrements the Colony Population-entry count and removes entries until every cohort is at or below its limit.

The same trimming helper is also called from surrender/ownership/invasion paths, which is consistent with capacity changing when the owner/race/Technology context changes.

For today's aggregate single-race MOOX model the faithful semantic equivalent of such a future capacity-reducing transition is an **immediate clamp to the newly valid capacity**, not merely suppressing future growth. Mixed-race removal priority remains explicitly deferred to the later cohort slice.

The transformations added by this slice are upward/non-decreasing physical-climate transitions, and Biospheres/Advanced City Planning acquisition increase capacity; this slice does not need to invent a new capacity-decrease gameplay action. The original immediate-clamp rule should nevertheless be preserved as a reusable invariant for future Building destruction/conquest/climate-degradation transitions.

## Current MOOX architecture gap

The current generic Building construction path would treat both `terraforming` and `gaia_transformation` as persistent Buildings because both originated in the original production/building table. Direct executable evidence proves that would be incorrect.

Biospheres **is** a normal persistent Building. Terraforming and Gaia Transformation are **one-shot planetary transformation projects** whose completion mutates `Planet.ClimateID` and does not append to `Colony.Buildings`.

## Minimal normalized/runtime identities required

No new catalog identity is required:

- Building `biospheres` already exists;
- Technology `advanced_city_planning` / ID 3 already exists;
- production definitions `terraforming` / ID 44 and `gaia_transformation` / ID 17 already exist;
- `Planet.ClimateID`, one-based `Planet.Orbit`, and authoritative state RNG already exist.

The rules layer does need semantic capacity/transformation metadata so gameplay does not depend on numeric Production IDs.

## Proposed Gate 2 implementation shape

No gameplay code has been changed in this slice. Proposed implementation:

1. **Ruleset / validation**
   - add Population-capacity effect metadata to `economy.json`:
     - `advanced_city_planning` flat +5;
     - `biospheres` flat +2;
     - semantic Terraforming/Gaia transformation definitions and allowed/result climates;
   - cross-validate all referenced Technology/production identities against committed catalogs;
   - encode Barren Terraforming orbit rule semantically, not as original offsets.

2. **Capacity API split matching the original architecture**
   - keep a planet/race base-capacity helper for size + effective climate + Tolerant + Subterranean;
   - add a planet/Empire capacity helper that adds Advanced City Planning, mirroring `Planet_Max_Population_For_Player_`;
   - add a Colony capacity helper that additionally adds Biospheres, mirroring `Colony_Race_Pop_Limit_`;
   - route Housing legality, Population dynamics and Population-transfer destination capacity through the authoritative Colony helper.

3. **Biospheres / Advanced City Planning**
   - Biospheres remains a normal Building and its +2 is derived from authoritative Colony Building ownership;
   - Advanced City Planning is derived from authoritative Empire Technology ownership;
   - preserve current materialize/apply timing: neither a research breakthrough nor a Building completion retroactively changes the already-materialized growth for that turn.

4. **Planetary transformation Construction**
   - add semantic persistent `ConstructionProjectPlanetaryTransformation = "planetary_transformation"`;
   - project IDs: `terraforming`, `gaia_transformation`;
   - add an authoritative queue command/legal-action path and exclude these identities from ordinary persistent Building choices;
   - use their existing normalized PP costs/Technology requirements;
   - completion mutates `Planet.ClimateID`, emits a transformation event, and never appends to `Colony.Buildings`;
   - Terraforming on Barren uses one-based Orbit 1/2 -> Desert, Orbit 4/5 -> Tundra, Orbit 3 -> authoritative RNG 50/50;
   - because the new project kind can be saved mid-build, advance Core schema 12 -> 13.

5. **Capacity-decrease invariant**
   - document/provide a reusable authoritative clamp helper for future capacity-decreasing transitions;
   - do not expand this slice into Building destruction, conquest or mixed-race cohort priority;
   - when a future transition lowers capacity, current aggregate single-race Population must be reduced immediately to the new cap, with cohort-specific priority deferred until cohorts exist.

6. **Post-completion recalculation**
   - retain Research -> Population -> Construction order;
   - transformation/Biospheres completion is visible in the existing post-Construction next-state recalculation;
   - no same-turn retroactive growth or PP/RP contribution is introduced.

## Proposed deterministic regression coverage

### Rules / Game

- normal base capacity fixtures remain unchanged;
- Advanced City Planning adds exactly +5 after race/planet layers;
- Biospheres adds exactly +2 after Advanced City Planning;
- Aquatic/Tolerant/Subterranean + ACP + Biospheres combination locks the original order;
- Biospheres completion does not retroactively increase that turn's already-materialized growth;
- newly researched Advanced City Planning becomes visible in the post-turn capacity snapshot, not the already-materialized growth delta;
- Terraforming legal only on Barren through Arid;
- Gaia Transformation legal only on Terran;
- Terraforming mappings for Desert/Tundra/Ocean/Swamp/Arid are exact;
- Barren Orbit 1/2 -> Desert, 4/5 -> Tundra, Orbit 3 consumes authoritative RNG and deterministically selects Desert/Tundra;
- Gaia Transformation maps Terran -> Gaia;
- transformations never appear in `Colony.Buildings` after completion;
- transformation completion changes next-state capacity but not current-turn applied Population growth.

### Core

- schema-13 planetary-transformation Construction state round-trips exactly;
- invalid transformation project IDs are rejected.

### Session / Observer

- owning seat sees/queues the legal transformation action; foreign seat cannot;
- Observer/replay sees queue/progress/completion plus old/new climate;
- replay from the same RNG state reproduces a middle-orbit Barren Terraforming result exactly.

## Gate 1 conclusion

All requested Gate 1 questions are closed by direct original evidence or existing normalized identities. The only intentionally unmodeled original details are opaque auxiliary planet-storage bytes changed alongside climate; they are not required for the Population-capacity semantics of this slice and should not be copied without a separate gameplay need.

Gameplay implementation must stop here until Gate 2 accepts the semantic design above.

## Gate 2 - implementation decision

Accepted by the user on 2026-08-29. The semantic design above is authoritative for this slice: split capacity layers, persistent Biospheres, passive Advanced City Planning, one-shot planetary transformation projects, schema 13, and deterministic authoritative RNG for middle-orbit Barren Terraforming.

## Gate 3 - implementation

**Status:** complete.

Implemented runtime shape:

- Core `StateSchemaVersion` advanced from 12 to 13.
- Economy ruleset schema advanced from 6 to 7.
- Added semantic `ConstructionProjectPlanetaryTransformation = "planetary_transformation"` with valid project IDs `terraforming` and `gaia_transformation`.
- Added validated capacity/transformation metadata to `economy.json` with direct `Orion2.exe` provenance.
- Added original-style capacity layers:
  - race/planet capacity = effective climate -> Tolerant -> rounding -> Subterranean;
  - Empire capacity adds Advanced City Planning +5;
  - Colony capacity adds Biospheres +2.
- Population Dynamics and Housing legality now use the authoritative Colony capacity layer.
- Biospheres remains a normal persistent Building.
- Terraforming and Gaia Transformation are excluded from persistent Building choices and rejected through `colony.queue_building`.
- Added `colony.queue_planetary_transformation` plus authoritative ownership, Technology and physical-climate legality checks.
- Transformation construction uses the existing normalized PP costs and pre-growth Production snapshot.
- Completion mutates `Planet.ClimateID`, emits a transformation event and never appends the project to `Colony.Buildings`.
- Terraforming mappings implemented exactly for Desert/Tundra/Ocean/Swamp/Arid; Barren uses one-based Orbit 1/2 -> Desert, 4/5 -> Tundra, Orbit 3 -> deterministic authoritative RNG between Desert/Tundra.
- Gaia Transformation maps Terran -> Gaia.
- Added an aggregate single-race capacity-clamp helper for future proven capacity decreases; until cohort state exists it preserves current job proportions while immediately clamping total Population.
- Existing Research -> Population -> Construction ordering remains unchanged. ACP research and Biospheres/Gaia completion affect the post-turn next-state capacity and do not retroactively grow Population in the same turn.

Deterministic regression coverage added:

- schema-13 planetary-transformation state round-trip and invalid project rejection;
- Aquatic + Tolerant + Subterranean + ACP + Biospheres capacity stacking/order;
- Terraforming excluded from persistent Building choices and rejected by the Building queue command;
- physical-climate legality for Terraforming/Gaia;
- all direct Terraforming/Gaia climate mappings;
- deterministic Barren Orbit-3 RNG and no RNG consumption on deterministic inner/outer orbit paths;
- transformation queue/progress/completion with no persistent Building leakage;
- ACP breakthrough timing: no retroactive same-turn Population growth, +5 visible in next-state projection;
- Biospheres/Gaia completion timing: no retroactive same-turn Population growth, new capacity visible in next-state projection;
- aggregate capacity clamp preserves continuous job proportions;
- GameSession seat authority, Observer queue/progress visibility and deterministic completion/replay for Barren middle-orbit Terraforming.

## Gate 4 - QA

The completed implementation passed:

```text
gofmt on all changed Go files
git diff --check
go test ./...
go vet ./...
```

All original-executable temporary disassembly extracts from Gate 1 were removed before implementation. No original binary/object copy is part of the repository changes.
