# Race-aware Population cohorts - 2026-08-29

**Open marker:** `docs/slices/_OPEN_POPULATION_COHORTS_2026-08-29.md`

## Goal

Determine original Master of Orion II 1.31 Population cohort behavior for native/conquered/mixed-origin Colonies and the Food-logistics priority logic that depends on those cohorts before replacing MOOX's current aggregate single-origin `PopulationState`.

## Gate 1 - Checkup + original analysis / reverse engineering

**Status:** Gates 1-3 complete; Gate 4 follow-up QA + commit + close pending.

### Recovery state

- Starting branch: `main`.
- Starting HEAD: `5ce3c99` (`docs: close population capacity transition slice`).
- Starting tree: clean, `main` ahead of `origin/main` by 10 commits.
- Previous gameplay slice: Population capacity transitions (`25d05a9`), closed.
- Core `StateSchemaVersion`: 13.
- Economy ruleset schema: 7.

## Evidence / provenance

Private original executable used for direct reverse engineering:

```text
reference/original/support/files/Orion2.exe
SHA-256 7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5
```

Original text evidence used to resolve special-population semantics:

```text
reference/original/text/estrings/block_0000.ascii.txt
reference/original/text/help/block_0000.ascii.txt
reference/original/text/maintext/block_0006.ascii.txt
reference/original/text/rstring0/block_0000.ascii.txt
```

Important directly traced executable routines:

```text
Pops_Identical_                         VA 0xB9CE3
Get_Effective_Pop_Player_               VA 0xBADCC
Pop_Race_String_                        VA 0xBAE1A
Pop_To_Pop_State_                       VA 0xBCACF
Sum_Colonists_                          VA 0xBC972
Send_Cluster_                           VA 0xB9E94
Colony_Pop_Prod_Produced_               VA 0xDDFD3
Colony_Food2_Per_Farmer_                VA 0xDE03E
Colony_Empire_Base_Food2_Produced_      VA 0xDE0C6
Colony_Pop_Base_Prod_Produced_           VA 0xDE22C
Colony_Food_Production_                 VA 0xDE664
Colony_Food_Maintenance_                VA 0xDEB4B
Colony_Empire_Base_Industry_Produced_   VA 0xDED47
Colony_Industry_Production_             VA 0xDEE1B
Pass_Out_Imports_                       VA 0xDF8F0
Colony_Race_Pop_Limit_                  VA 0xE0C1D
Colony_Pop_Grows_                       VA 0xE1839
Apply_Colony_Pop_Growth_                VA 0xE2DCA
Apply_Assimilation_                     VA 0xE3456
population-over-cap trimming helper     VA 0xEC97C
Change_Pop_Ownership_                   VA 0xECBF7
Occupation_Policy_Popup_                VA 0xCD969
Random_                                 VA 0x1247A0
```

Temporary LE-object extracts/disassembly files are research scratch only and are removed at Gate 1 close.

## Correction of interrupted sampler findings

The interrupted analysis produced two contradictory intermediate mappings for special source codes 8/9. The cause was a provisional debug-symbol parser that sometimes associated an unnamed helper with the following symbol record. Those intermediate names are **not** evidence and must not be reused.

The consolidated result below is based on executable control flow plus original HELP/MAINTEXT evidence. In particular:

```text
source code 8 = Android
source code 9 = Native
```

This mapping is now strong enough for permanent documentation.

## Original Population storage model

**Evidence level: direct original executable.**

An original Colony stores:

```text
colony + 0x0A = number of Population entries
colony + 0x0C = first Population entry
entry size     = 4 bytes / 32 bits
```

Each full original Population entry represents one whole Population unit. MOOX deliberately uses continuous `float64` Population and should preserve the semantics without restoring the 1996 integer-per-million representation.

### Packed fields required by this slice

The directly decoded fields are:

```text
bits  0..3  source/origin player code
bits  4..6  loyalty / assimilation owner player code
bits  7..8  job
bit      10  conquered / not-yet-assimilated organic population
```

Job values are:

```text
0 = Farmer
1 = Worker
2 = Scientist
```

`Pops_Identical_` directly compares job, broad special-pop state, source code and conquered bit. The original can therefore contain two entries with the same source race but different assimilation status and/or job.

Higher packed bits include additional UI/gameplay statuses. Original strings prove separate Rebel and Expert states, but their exact bit positions are not required for the race/conquest/Food slice and are intentionally not guessed here.

## Origin identity is an Empire/player identity, not merely a species string

**Important architecture result.**

For normal source codes `0..7`, the source code indexes an original player slot. `Pop_Race_String_`, capacity/growth paths and production paths resolve traits through that source player. Captured Population therefore retains the racial/custom-trait identity of the Empire/player that originated it.

For MOOX the stable semantic equivalent is **`OriginEmpireID`**, not only `RaceID`.

`OriginEmpireID` can resolve the current normalized `RaceID` and later custom-race traits from that Empire. This also preserves the correct conceptual model if custom races are added later.

The loyalty field is a separate player identity. Its clean MOOX equivalent is **`LoyaltyEmpireID`**.

## Organic assimilation state

For normal organic source codes `< 8`:

- bit 10 clear = not currently marked conquered;
- bit 10 set = conquered / not yet assimilated;
- loyalty-owner bits record the Empire to which that population is assimilated/loyal.

`Change_Pop_Ownership_` proves that source and loyalty are independent:

1. source/current-owner population becomes loyal to the new owner and is not conquered;
2. foreign organic Population whose stored loyalty already matches the new owner is not conquered;
3. Telepathic conquest immediately rewrites loyalty to the new owner and clears the conquered bit;
4. otherwise foreign organic Population keeps its previous loyalty and receives the conquered bit.

This state is per Population entry, so one origin race can simultaneously have assimilated and conquered Population in the same Colony.

## Special source codes

### Source 8 = Android

**Evidence level: direct executable + original HELP.**

Executable evidence:

- `Pop_To_Pop_State_` gives source 8 its own special state;
- `Colony_Food_Maintenance_` gives source 8 **zero Food maintenance**;
- `Colony_Empire_Base_Food2_Produced_` returns `6 FOOD2 = 3 Food` for source 8;
- `Colony_Empire_Base_Industry_Produced_` returns `3 Production` for source 8;
- the capacity helper treats source 8 as full-habitability/Tolerant-like without normal race trait lookup.

Original HELP explicitly describes Android Farmer/Scientist/Worker:

- each produces 3 units in its fixed job;
- each requires **1 Production per turn instead of Food**;
- Androids do not receive normal racial bonuses;
- Androids are unaffected by morale;
- Androids do not generate normal population income.

Original UI text also says an Android cannot be reconfigured to another job.

### Source 9 = Native

**Evidence level: direct executable + original MAINTEXT/UI text.**

Executable evidence:

- `Pop_To_Pop_State_` gives source 9 its own special state;
- `Colony_Food_Maintenance_` adds `2 FOOD2 = 1 Food` per source-9 Population;
- `Colony_Empire_Base_Food2_Produced_` returns `4 FOOD2 = 2 Food` for source 9;
- source 9 does not receive the Android Industry base path;
- source 9 does not use normal racial Aquatic/Tolerant/Subterranean traits in the population-capacity helper.

`MAINTEXT.LBX` is particularly explicit:

```text
Humanoid life ... may only work the farms, and will not leave the planet to colonize elsewhere.
Furthermore, the natives do not gain the racial bonuses of your race.
```

This agrees with the executable's source-9 Food/Industry behavior. A separate ESTRINGS UI sentence says natives can only farm or mine; because that conflicts with MAINTEXT and the executable Industry base, the runtime evidence + MAINTEXT wording is treated as authoritative for this research checkpoint: **Native Population is farm-only and cannot be resettled.**

### Scope decision for Android/Native

Android and Native semantics are now sufficiently resolved for future implementation, but they are **not proposed as part of Gate 3 of this slice**. Their morale immunity, Production sustenance, tax behavior, fixed-job rules and dedicated production projects would turn this cohort/Food slice into a second large Economy slice.

The organic cohort architecture below deliberately leaves room for later `android` / `native` kinds without guessing them into current runtime state.

## Food maintenance cohort buckets

**Evidence level: direct original executable.**

`Colony_Food_Maintenance_` scans each original Population entry and builds half-Food (`FOOD2`) buckets:

```text
colony + 0xFC = current owner's origin population
colony + 0xFD = assimilated foreign organic population
colony + 0xFE = conquered/unassimilated foreign organic population
colony + 0xFF = Native population (source 9)
```

Android source 8 contributes no Food maintenance.

For normal organic Population the per-population Food requirement is source-race aware, preserving already-normalized Lithovore/Cybernetic/ordinary sustenance semantics.

The effective whole-Food requirement is:

```text
EF = ceil((FC + FD + FE + FF) / 2)
```

The routine also stores per-source maintenance, which the starvation calculation later uses to allocate foreign shortages back to origin races.

## Four-pass insufficient-Freighter Food priority - resolved

The earlier research in `docs/research/INSUFFICIENT_FREIGHTER_PRIORITY_2026-08-28.md` had already proved four round-robin thresholds but deliberately left the bucket meanings unresolved. They are now closed.

`Pass_Out_Imports_` uses:

```text
pass 1: 2 * (local_food + imports) < FC
pass 2: 2 * (local_food + imports) < FC + FD
pass 3: 2 * (local_food + imports) < FC + FD + FE
pass 4:     (local_food + imports) < EF
```

So the original shortage priority is:

1. feed the Colony owner's own organic Population;
2. then assimilated foreign organic Population;
3. then conquered/unassimilated foreign organic Population;
4. then the remaining rounded whole-Food requirement, including Native demand and half-Food rounding.

Each successful allocation still consumes exactly one Food and one Freighter, and each threshold uses the previously proven rotating round-robin Colony order. Androids never participate in Food imports because their sustenance is Production.

This is the exact missing semantic layer needed to replace MOOX's currently collapsed single round-robin pass once organic cohorts exist.

## Conquered organic productivity

**Evidence level: direct original executable.**

`Colony_Pop_Prod_Produced_` uses the normal per-population production path and then applies:

```text
conquered bit set   -> positive output * 3
conquered bit clear -> positive output * 4
```

All later common factors see the same scaled value, therefore conquered/unassimilated organic Population produces **75% of the otherwise equivalent assimilated Population output**.

The original Alien Management Center HELP says aliens work harder and assimilate faster. Its exact executable interaction with every government/advanced-government state is not fully normalized here and is intentionally parked rather than mixed into this slice.

## Mixed-origin Population growth

**Evidence level: direct original executable.**

`Colony_Pop_Grows_` first counts Population by origin/source and separately tracks conquered organic entries. Natural growth then runs for normal organic source/player codes `0..7`; Android and Native do not use this organic-growth loop.

For each represented organic origin:

```text
origin_capacity = Colony_Race_Pop_Limit_(colony, origin_player)
origin_population = all organic Population with that origin
colony_population = total Colony Population across all origins/types

curve_input = ((origin_capacity - colony_population) * origin_population * 2000) / origin_capacity
natural_kpop = integer_sqrt(curve_input) * origin_growth_percent / 100
```

Important consequences:

- capacity is source/origin-race specific;
- the cohort-size term is the Population of that origin;
- **crowding uses total Colony Population**, not only the origin's own Population;
- Growth traits come from the origin player/race;
- previously proven Medicine/Housing/other natural-growth modifiers layer onto this source-specific calculation.

### Occupation policy correction

An earlier interrupted finding incorrectly generalized a conquered-population growth suppression.

`Occupation_Policy_Popup_` plus HELP/ESTRINGS resolve Colony `+0x12F`:

- value `0` = Annihilate conquered occupants;
- value `3` = keep/Assimilate path.

The original multiplies positive foreign-race natural growth by the already-assimilated fraction **only while the Colony is on the Annihilate path**. Under the normal Assimilate/keep path, the full represented origin Population participates in natural growth.

MOOX does not yet have canonical conquest/occupation-policy gameplay. Gate 3 should therefore implement the normal assimilating mixed-origin case and leave the Annihilate policy for the future conquest slice instead of inventing a temporary occupation-policy command.

### Where new foreign-race growth lands

`Apply_Colony_Pop_Growth_` directly constructs a newly grown Population entry with:

```text
source/origin = the growing origin race
loyalty owner = current Colony owner
conquered bit = clear
```

Therefore **new Population of a foreign origin is born assimilated to the current Colony owner**, even if older Population of that same origin still contains conquered members.

MOOX should add positive origin-specific fractional growth to the matching **assimilated** cohort.

The original chooses a discrete job for each newly created whole Population unit. MOOX already modernizes Population as continuous fractional values and currently scales jobs continuously. This slice should retain that established continuous job-distribution choice rather than reintroducing one-million-person job RNG.

## Starvation / Population loss by origin

`Colony_Pop_Grows_` consumes available sustenance against the Food buckets in the same owner -> assimilated foreign -> conquered foreign -> Native order and accumulates losses by source/origin race.

For foreign organic shortages, per-origin maintenance is used to assign the shortage back to origin races. The final negative Population accumulator is therefore source/origin specific, not a single aggregate loss.

`Apply_Colony_Pop_Growth_` then removes Population of the affected origin. Inside that origin it prefers non-Farmers (Worker/Scientist) before Farmers; within those groups the original uses RNG over individual entries. It does not make assimilated-vs-conquered status the removal priority for starvation.

The clean continuous MOOX analogue is:

1. calculate starvation loss by `OriginEmpireID`;
2. remove loss proportionally across Worker/Scientist Population of that origin first;
3. then Farmers of that origin;
4. distribute across assimilated/conquered cohorts of that origin proportionally within the same job class.

This preserves the original semantic priority without reintroducing whole-Pop RNG into MOOX's continuous model.

## Heterogeneous capacity constraints and immediate trimming

**Evidence level: direct original executable.**

`Colony_Race_Pop_Limit_` calculates a limit for a given origin/source Population using that source's race traits plus owner-wide Colony effects such as Advanced City Planning and Biospheres.

The helper at VA `0xEC97C` is more sophisticated than a single aggregate clamp:

1. calculate each represented source's Population limit;
2. iterate distinct capacity thresholds in ascending order;
3. at threshold `C`, count all Population entries whose own source capacity is `<= C`;
4. if that count exceeds `C`, remove a candidate and repeat until the threshold is satisfied;
5. continue upward through higher capacity thresholds.

This means a low-capacity origin constrains the subset at or below that limit, while higher-capacity origins may occupy additional slots above it. A mixed Colony therefore cannot be represented by one owner's scalar capacity alone.

### Direct over-cap removal priority

When a candidate must be removed, the original preference is:

1. Native (source 9);
2. foreign organic not loyal/assimilated to the current owner;
3. foreign organic loyal/assimilated to the current owner;
4. Android (source 8);
5. current owner's own origin Population.

For the organic-only Gate-3 scope this reduces to:

```text
conquered foreign -> assimilated foreign -> owner's own origin
```

Within equal priority the original works on individual entries; MOOX can apply continuous proportional trimming within the selected semantic class.

## Ownership change

`Change_Pop_Ownership_` directly proves the following semantic transitions:

- `OriginEmpireID` is preserved;
- Population that originated from the new owner becomes assimilated/loyal to the new owner;
- foreign Population already loyal to the new owner becomes assimilated immediately;
- Telepathic new owners immediately assimilate foreign organic Population;
- other foreign organic Population becomes conquered while retaining its previous loyalty identity;
- capacity is re-evaluated and the heterogeneous trimming helper runs after ownership changes.

MOOX does not yet have canonical invasion/conquest state, so Gate 3 should implement the cohort representation and helpers required by these semantics without inventing an ownership-change strategic system.

## Assimilation

`Apply_Assimilation_` shows that assimilation is per-Population, not per-race as one indivisible block:

- an assimilation progress byte exists at Colony `+0x12E`;
- a full conversion occurs whenever progress reaches 240;
- one eligible conquered organic entry is selected using the authoritative RNG;
- the selected entry has its conquered bit cleared and loyalty rewritten to the Colony owner;
- multiple converted entries of the same origin can therefore coexist with still-conquered entries.

Original HELP gives the base-government headline rates:

```text
Feudal       1 Population / 8 turns
Dictatorship 1 Population / 8 turns
Democracy    1 Population / 4 turns
Unification  1 Population / 20 turns
```

HELP also states Feudal Population normally assimilates instantly when conquered by another race. Advanced-government and Alien Management Center interactions are not completely normalized in this checkpoint.

Because MOOX has no authoritative conquest/occupation lifecycle yet, **automatic assimilation ticking is intentionally deferred**. Gate 3 only needs a state shape that can later move continuous Population from a conquered cohort to the matching assimilated cohort without changing origin or jobs.

## Population transfer / resettlement implications

MOOX currently identifies a transfer only by source Colony, destination Colony and job. That becomes ambiguous as soon as two origins/statuses share the same job.

Original UI text directly proves:

- conquered Population cannot be resettled;
- Native Population cannot leave the planet;
- Androids cannot be reconfigured to another job;
- Expert status is lost on transfer (Expert itself is outside this slice).

For organic Gate-3 scope:

- only assimilated organic cohorts may be transferred;
- a transfer must carry the exact origin/loyalty cohort identity, not merely `Job`;
- arrival preserves `OriginEmpireID` and remains assimilated to the same owning Empire;
- existing one-Population-per-transfer MOOX behavior may remain;
- same-system and in-transit Freighter semantics remain unchanged.

## Existing MOOX architecture gap

Current Core state is aggregate:

```go
type PopulationState struct {
    Total      float64
    Farmers    float64
    Workers    float64
    Scientists float64
}
```

This loses every distinction required above:

- origin Empire/race;
- loyalty Empire;
- assimilated vs conquered;
- per-origin capacity/growth;
- four Food priority buckets;
- exact transfer identity.

Current `CalculateBaseEconomy` also applies one race's base Food/Production/Research values to all jobs, current Population Growth has one scalar capacity, and Population transfer chooses only a job.

The direct aggregate Population accesses are concentrated in Population, transfer, resolver and a limited set of tests, so a full authority migration is feasible in one slice. A permanent duplicated mutable aggregate cache is not justified.

## Gate 2 implementation decision

**Decision:** accepted on 2026-08-29 without scope changes. The following design is the binding Gate 3 implementation contract. No gameplay code had been changed when Gate 2 was accepted.

Accepted design:

### 1. Core schema 13 -> 14; organic cohorts become authoritative

Replace mutable aggregate Population fields with an authoritative ordered cohort list:

```text
PopulationState
  Cohorts []PopulationCohort

PopulationCohort
  OriginEmpireID
  LoyaltyEmpireID
  AssimilationState = assimilated | conquered
  Farmers
  Workers
  Scientists
```

For this slice all cohorts are organic. Android/Native are documented future kinds, not schema values yet.

Derived helpers provide:

```text
Total()
Farmers()
Workers()
Scientists()
TotalsByOrigin()
CohortBySemanticKey()
```

No second mutable aggregate truth is persisted.

Cohorts are normalized/merged and kept in deterministic semantic order. Validation rejects negative values, missing Empire identities, duplicate keys and impossible owner/conquered combinations.

Existing fixtures are converted to one assimilated organic cohort:

```text
OriginEmpireID  = Colony.EmpireID
LoyaltyEmpireID = Colony.EmpireID
jobs             = previous aggregate job values
```

### 2. Per-origin base Economy; owner-wide context remains owner-wide

Refactor base Food/Production/Research calculation to iterate cohorts.

For each organic cohort:

- resolve base racial production traits from `OriginEmpireID -> Empire.RaceID`;
- apply source-race gravity where the existing model supports gravity;
- if conquered, apply the direct original 0.75 positive-output factor;
- sum the cohort outputs.

Owner-wide Colony morale/government/building context remains applied through the owning Empire, preserving the current Economy layering rather than treating a captured population's old government as the Colony government.

Tax remains an owner/Colony aggregate unless separate original evidence later requires a source-specific split.

### 3. Food maintenance buckets + exact four-pass allocator

Materialize semantic Food2 demand per Colony:

```text
owner_origin_food2
assimilated_foreign_food2
conquered_foreign_food2
remaining_food2
whole_food_required
```

For organic Gate-3 scope `remaining_food2` is normally zero; pass 4 still matters for final half-Food rounding. The field shape leaves Native demand straightforward for the later special-pop slice.

Replace the current one collapsed round-robin pass with the four original thresholds while preserving:

- stable Colony-ID scan order;
- rotating start behavior already normalized;
- one Food + one Freighter per allocation step;
- existing continuous partial final load modernization.

### 4. Per-origin Population Growth

Calculate natural growth separately for each represented `OriginEmpireID`:

```text
origin_capacity = Colony capacity using origin race + owner ACP/Biospheres
origin_population = assimilated + conquered population of that origin
crowding_population = total Colony Population
```

Use the already-normalized continuous equivalent of the original square-root formula and the origin race's Growth modifier.

Under the current no-occupation-policy runtime, use the normal Assimilate/keep semantics: all Population of that origin contributes to cohort size.

Positive growth is added to the matching **assimilated** cohort of that origin. Keep MOOX's established continuous job-distribution modernization instead of reproducing discrete one-Pop job assignment RNG.

`ColonyPopulationDynamics` can retain aggregate display totals, but authoritative application must be backed by per-origin projected deltas rather than one scalar growth value.

### 5. Per-origin starvation application

Compute remaining Food shortage by the same semantic bucket order, map foreign shortages back to origin by maintenance burden, and apply loss per origin.

Continuous loss order within an origin:

```text
Worker + Scientist first, proportionally
then Farmer
```

Assimilated/conquered cohorts within the same origin/job class share that loss proportionally because the original starvation removal does not prefer one assimilation status.

### 6. Heterogeneous capacity + organic-only trim priority

Implement the original ascending-threshold capacity constraint across represented organic origins.

When trimming is required in this slice:

```text
conquered foreign
-> assimilated foreign
-> owner's own origin
```

Android/Native priority slots are added only when those special kinds become runtime state later.

Route Housing legality, Population Growth and Population-transfer destination checks through the mixed-origin capacity helper rather than one owner's scalar cap.

### 7. Population transfer selector becomes cohort-aware

Extend transfer command/state/event identity with an organic cohort selector:

```text
OriginEmpireID
LoyaltyEmpireID
AssimilationState
Job
```

Only assimilated organic Population is legal for resettlement. The in-transit record preserves the source origin identity and arrival adds to the corresponding assimilated destination cohort.

### 8. Explicitly deferred from this slice

Do not expand Gate 3 into:

- Android construction, Production sustenance, morale immunity or tax behavior;
- Native generation / planet-native lifecycle;
- Rebel/Expert packed status fidelity;
- active conquest/invasion/occupation-policy commands;
- automatic assimilation ticking / Alien Management Center exact rate table;
- Annihilate policy and one-pop-per-turn killing;
- Fleet/Diplomacy blockade production;
- unrelated Treasury/UI work.

Those are documented future extensions, not hidden assumptions.

## Proposed deterministic regression coverage

### Core

- schema-14 organic single-cohort state round-trips exactly;
- mixed origins/statuses round-trip in deterministic order;
- aggregate helper totals exactly equal cohort sums;
- duplicate/invalid cohort keys and impossible conquered owner state are rejected.

### Economy

- single-origin fixture reproduces every existing Food/PP/RP result exactly;
- two assimilated origins use their own race Food/PP/RP baselines;
- conquered organic output is exactly 75% of otherwise matching positive output;
- source-race gravity is applied per origin while owner-wide morale/government remains owner-wide.

### Food logistics

- four-pass demand order locks owner -> assimilated foreign -> conquered foreign -> final rounded demand;
- insufficient Freighters stop at the exact pass reached;
- round-robin fairness remains deterministic within each pass;
- Cybernetic half-Food demand proves pass-4 rounding even without Native runtime state;
- existing homogeneous one-pass behavior remains observationally identical when no mixed cohorts exist.

### Growth / starvation

- two origins use separate race capacity/Growth traits but shared total Colony crowding;
- positive foreign-origin growth lands in its assimilated cohort;
- conquered and assimilated members of the same origin are combined for normal Assimilate-path cohort size;
- starvation shortage is attributed by origin and removes non-Farmers before Farmers;
- no negative/duplicate cohorts remain after continuous mutations.

### Capacity

- low-cap and high-cap origins enforce ascending threshold constraints;
- organic trim priority is conquered foreign -> assimilated foreign -> owner origin;
- ACP/Biospheres remain owner-wide additions to each origin-specific capacity.

### Transfer / Session / Observer

- same-job mixed-origin Colony requires explicit cohort selector;
- conquered cohort transfer is rejected;
- assimilated foreign-origin transfer preserves origin on arrival;
- foreign seat cannot issue transfer commands;
- Observer/replay sees cohort identities, transfer identity and deterministic post-turn cohort state.

## Gate 1 conclusion

The original evidence is sufficient to close Gate 1 for an **organic race-aware cohort + four-pass Food priority** implementation.

The key resolved facts are:

1. Population identity is origin-player/Empire based, with separate loyalty and conquered state.
2. Jobs are per Population entry and can coexist across origin/status cohorts.
3. Food priority is owner -> assimilated foreign -> conquered foreign -> final rounded/Native demand.
4. Conquered positive output is 75% of assimilated output.
5. Growth is per origin race with shared total Colony crowding; new foreign-origin growth is born assimilated.
6. Starvation is attributed per origin and removes non-Farmers before Farmers.
7. Capacity is a heterogeneous ascending-threshold constraint, not one owner scalar.
8. Ownership change preserves origin, rewrites loyalty/conquered semantics, and can immediately trim over-cap Population.
9. Android is source 8 and Native is source 9; both are now sufficiently resolved but intentionally deferred from Gate 3.
10. Active conquest/occupation/assimilation progression is not required to make the organic cohort state and Food algorithm authoritative today.

Gate 2 is accepted. Gate 3 may now implement the contract above; any material scope or semantic change must be documented before implementation continues.

## Gate 3 - implementation result

**Status:** complete on 2026-08-29. The accepted Gate 2 design was implemented without expanding into Android, Native, active conquest/occupation, automatic assimilation, Fleet/Diplomacy blockade production or unrelated Treasury/UI work.

### Authoritative Core schema 14

`PopulationState` now persists only an ordered `[]PopulationCohort`. Each organic cohort carries `OriginEmpireID`, `LoyaltyEmpireID`, `AssimilationState` and Farmer/Worker/Scientist quantities. Aggregate totals are derived helpers rather than a second mutable truth.

Core normalization merges equal semantic keys, removes epsilon-empty cohorts and orders cohorts deterministically. Validation rejects unknown identities, duplicate keys, invalid assimilation states, impossible conquered owner-origin state and invalid/negative job quantities. Existing single-race fixtures are represented by one owner-loyal assimilated cohort.

Population transfer state now persists the exact cohort identity in addition to the job, so interstellar transport survives save/load and Session cloning without losing origin/status semantics.

### Race-aware Economy

The authoritative resolver now calculates job output per cohort using the cohort origin race. Source-race gravity is applied per cohort; government, morale and Colony-wide context remain owned by the current Empire. Conquered organic positive Food/Production/Research output receives the directly proven `0.75` factor. Tax remains owner-wide as accepted at Gate 2.

The legacy single-race Economy helper remains available for focused unit/API compatibility, but the strategic resolver uses the cohort-aware path.

### Exact Food2 maintenance and four-pass imports

`ColonyPopulationDynamics` now records the semantic Food2 buckets:

```text
OwnerOriginFood2
AssimilatedForeignFood2
ConqueredForeignFood2
RemainingFood2
WholeFoodRequired
```

Organic runtime state currently leaves `RemainingFood2` at zero; it is reserved for the later Native/special-pop extension. `WholeFoodRequired` uses the final half-Food rounding boundary.

The import allocator now runs the original semantic thresholds in stable Colony order:

1. owner-origin demand;
2. owner + assimilated foreign demand;
3. owner + assimilated foreign + conquered foreign demand;
4. final whole-Food rounded demand.

The existing MOOX continuous partial-final-load modernization remains in place while Freighter accounting stays deterministic.

### Per-origin growth and starvation

Population dynamics now persist deterministic `PopulationOriginDynamics` rows. Each represented origin uses its own race Growth modifier and race-specific capacity while all origins share total Colony crowding.

Positive growth for a foreign origin is added to the matching **assimilated** cohort, even when the existing population of that origin is currently conquered. Continuous job distribution remains the established MOOX modernization.

Food/Production starvation is attributed per origin. Within an origin, Worker and Scientist population is removed before Farmers; assimilated/conquered cohorts of the same origin share same-job loss proportionally. The existing minimum-survivor invariant remains authoritative.

### Heterogeneous capacity

Mixed Colony capacity is no longer treated as one owner-race scalar. The runtime evaluates the directly reconstructed ascending source-race capacity thresholds across represented origins. Capacity-reducing transitions and post-growth enforcement use the organic trim priority:

```text
conquered foreign
-> assimilated foreign
-> owner origin
```

Advanced City Planning and Biospheres remain owner-wide additions to each source-race capacity. Housing availability and Population-transfer destination legality now use the same heterogeneous-capacity model. Terraforming/Gaia completion trims through the same invariant.

### Cohort-aware Population transfer and Session/Observer

`transfer_population` accepts an optional explicit cohort selector. A legacy command without a selector is accepted only when exactly one owner-loyal assimilated cohort has the requested job. Same-job mixed-origin transfers are rejected as ambiguous until the caller supplies the selector. Conquered Population cannot be resettled.

Same-system and interstellar transfers preserve origin/loyalty/status. Started, transferred, arrived and lost events all carry the resolved cohort key. The authoritative Session clone path already round-trips through Core serialization, so cohort slices are detached correctly; a new integration regression proves foreign-origin identity reaches Observer state and replay events unchanged.

### Deterministic regression coverage added

New regressions cover:

- schema-14 mixed-cohort serialization, ordering, helpers and semantic validation;
- global assignment compatibility without losing origin totals;
- source-race Economy plus exact conquered `0.75` output;
- four Food import thresholds including final half-Food rounding;
- heterogeneous threshold enforcement and organic trim priority;
- source-race Growth with foreign growth landing assimilated;
- non-Farmer-first starvation across assimilation statuses;
- ambiguous/conquered transfer rejection and explicit foreign cohort transfer;
- foreign cohort identity through authoritative GameSession/Observer event flow.

Implementation-stage verification completed successfully:

```text
go test ./internal/core ./internal/game
go test ./internal/session
go test ./...
```

The repository-wide Go test suite is green. Gate 4 remains responsible for the formal final formatting/vet/diff review, permanent status reconciliation, commit and slice close.