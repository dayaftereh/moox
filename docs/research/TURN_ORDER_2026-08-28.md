# Original MOO2 1.31 strategic turn ordering - 2026-08-28

Status: direct original-executable evidence resolves the Population/PP/RP turn-order question.

This checkpoint examines the private original MOO2 1.31 `Orion2.exe` and separates two concepts that secondary descriptions often collapse into one:

1. **calculation/materialization** of colony Food/Industry/Research/BC and projected Population change;
2. **application** of the already-materialized turn values during `Next_Turn_Calc_`.

That distinction matters because the original executable does not first mutate Population and then freshly calculate the RP/PP consumed by the same turn. Instead, the current-turn resource snapshot already exists when the end-turn apply dispatcher starts.

## Reference binary

```text
C:\ASH\Temp\mastori2\Orion2.exe
SHA-256 7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5
```

Bound LE facts already established elsewhere in this project:

```text
inner MZ module start: 0x26654
LE header:             0x292E4
code object:           object 1
object-1 base VA:      0x10000
```

For code symbols in object 1:

```text
VA = object_offset + 0x10000
```

## Debug-symbol calibration

The original executable contains Watcom debug symbol records. For the function records used here, the 32-bit object offset appears immediately before the record metadata/name.

Two independent calibration points match existing reverse engineering exactly:

```text
Cache_Load_Bldg_
object offset 0x9F6DC
```

and:

```text
Next_Turn_Calc_
object offset 0x36B3
VA            0x136B3
```

`0x136B3` is exactly the previously disassembled global turn dispatcher entry, so the symbol/address mapping is considered reliable for this investigation.

## Global `Next_Turn_Calc_` apply window

The important original call window is:

```text
0x13729 -> Make_Scrap_Ships_Dead_          (0xEDF92)
0x1372E -> Initialize_Reports_             (0xFD81C)
0x13733 -> Process_Trade_And_Research_Agreements_ (0x101E77)
0x13738 -> Move_All_Ships_Toward_Stars_    (0x100010)
0x1373D -> Resolve_Spies_                  (0x10192B)
0x13742 -> Apply_All_Player_Changes_       (0xE4F49)
0x13747 -> Apply_All_Colony_Changes_       (0xE3FDC)
0x1374C -> Do_Surrenders_                  (0xE4DC9)
...
0x1376F -> Do_Colony_Calculations_         (0xE2B31)
0x13774 -> Compute_Blockades_              (0xE5097)
0x13779 -> Move_Settlers_                  (0xFF212)
0x1377E -> Do_Colony_Calculations_         (0xE2B31)
```

The two calls to `Do_Colony_Calculations_` after the apply phases are especially important: the executable deliberately rematerializes colony/player calculation state after turn mutations and again after blockade/settler changes.

## Research is applied before Population growth

Inside `Apply_All_Player_Changes_` the relevant per-player order is:

```text
0xE4F89 -> Player_Maintenance_                    (0xEE0B0)
0xE4F9A -> Update_Player_Stats_                   (0xE2710)
0xE4FA1 -> Check_For_Research_Breakthrough_       (0xE44E0)
0xE4FA8 -> Compute_Player_Total_Known_Tech_Cost_  (0xE4535)
```

`Check_For_Research_Breakthrough_` is the already-verified original 1.31 research resolver. It:

```text
reads current-turn RP from player + 0xAC
adds it to accumulated RP at player + 0x1EB
rolls the breakthrough check
on success clears accumulated RP and calls Give_Player_Field_
```

Therefore Technology ownership can change inside `Apply_All_Player_Changes_`, before the later global `Apply_All_Colony_Changes_` call.

## Exact RP data flow is pre-growth

The direct data flow is now identified:

```text
Colony_Research_Production_ (VA 0xDFF74)
    writes colony + 0xEB

Update_Player_Stats_ (VA 0xE2710)
    sums colony + 0xEB
    stores result at player + 0xAC

Check_For_Research_Breakthrough_ (VA 0xE44E0)
    consumes player + 0xAC
```

Concrete instructions:

```text
0xE0050: mov WORD PTR [colony+0xEB], dx

0xE27F8: movsx eax, WORD PTR [colony+0xEB]
0xE27FF: add   [research accumulator], eax
...
0xE2A11: mov WORD PTR [player+0xAC], ax

0xE44FB: movsx eax, WORD PTR [player+0xAC]
0xE4502: add DWORD PTR [player+0x1EB], eax
```

This proves that the RP used by the current breakthrough pass is a materialized colony/player snapshot, not a value recalculated from Population after the turn's Population mutation.

## Population growth is calculated separately from applying it

The symbol table exposes both:

```text
Colony_Pop_Grows_          VA 0xE1839
Apply_Colony_Pop_Growth_   VA 0xE2DCA
```

`Do_Colony_Calculations_` calls `Colony_Pop_Grows_` as part of the calculation/projection phase.

The later mutation phase is explicit. `Apply_Colony_Changes_` at VA `0xE3F6E` calls, in order:

```text
0xE3FB3 -> Apply_Colony_Pop_Growth_  (0xE2DCA)
0xE3FBA -> Apply_Assimilation_       (0xE3456)
0xE3FC1 -> Produce_Ground_Military_  (0xE3616)
0xE3FC8 -> Apply_Production_         (0xE36DF)
```

And `Apply_All_Colony_Changes_` loops all eligible colonies through this function.

Because `Next_Turn_Calc_` calls `Apply_All_Player_Changes_` first, the direct original apply order relevant to MOOX is:

```text
Research breakthrough/ownership
-> Population growth/starvation apply
-> Production/construction apply
```

## Construction also consumes a pre-growth materialized PP snapshot

`Pre_Import_Computing_` (VA `0xE1D59`) calls the resource functions:

```text
Colony_Industry_Production_  (0xDEE1B)
Colony_Food_Production_      (0xDE664)
Colony_Research_Production_  (0xDFF74)
Colony_BC_Production_        (0xE03F1)
```

`Colony_Industry_Production_` writes its materialized Industry/PP result to:

```text
colony + 0xE9
```

including the final store at:

```text
0xDF0F0: mov WORD PTR [colony+0xE9], ax
```

Although `Apply_Colony_Pop_Growth_` is called immediately before `Apply_Production_`, `Apply_Production_` consumes the already-stored PP value. Its production accumulator reads `colony+0xE9` before the post-apply recalculation:

```text
0xE381A: movsx edx, WORD PTR [colony+0xE9]
...
production accumulator is updated / compared with product cost
```

Only later, after applying/completing the production item, `Apply_Production_` calls:

```text
0xE3960 -> Colony_Production_Calculation_ (0xE1E84)
0xE3978 -> Update_Player_Stats_            (0xE2710)
```

`Colony_Production_Calculation_` in turn starts the resource-calculation path again via `Pre_Import_Computing_` / imports.

So newly applied Population growth does not retroactively increase the PP already consumed by that same turn's construction. It affects the freshly materialized next-state production snapshot.

## Food/Freighter placement in MOOX

The original calculation phase includes:

```text
Pre_Import_Computing_
Pass_Out_Imports_
Colony_Pop_Grows_
Colony_Specialty_
Update_Player_Stats_
```

MOOX does not continuously recalculate colony UI state between commands like the original client. Instead it materializes local Economy and Food/Freighter logistics once at the strategic-resolution boundary. Semantically this corresponds to the original pre-apply calculation snapshot.

Therefore the current MOOX start-of-resolution materialization remains appropriate:

```text
local Economy / PopulationDynamics
Food/Freighter balancing
```

before applying Research/Population/Construction transitions.

## Resolved MOOX strategic apply order

The authoritative MOOX order is now:

```text
1. commands / legal selections
2. materialize pre-apply local Economy + PopulationDynamics
3. materialize Food/Freighter logistics
4. Research progress / breakthrough using pre-growth RP
5. Population growth / starvation apply
6. Construction apply using pre-growth PP
7. recalculate next-state Economy + PopulationDynamics
8. rematerialize next-state Food/Freighter preview without a second turn event
```

The key semantic invariant is more important than the storage representation:

```text
newly grown/starved Population does NOT retroactively alter
that same turn's consumed RP or PP
```

MOOX keeps continuous quantities as domain-native `float64`; only the ordering semantics are copied from the original.

## Why the StrategyWiki sequence looked contradictory

StrategyWiki's Calculations page says it was checked under MOO2 1.31 and describes a user-facing sequence roughly as:

```text
Population change
resource generation
building construction
research completion
```

That page remains useful for formulas and gameplay effects, but direct executable evidence shows that the implementation is a materialize/apply pipeline rather than a simple recompute-after-each-listed-step pipeline.

In particular:

- current RP already exists in `player+0xAC` before the breakthrough call;
- breakthrough/Technology ownership is applied before `Apply_All_Colony_Changes_`;
- Growth is then applied;
- Construction consumes already-materialized Industry;
- resource/player calculations are refreshed afterwards for the next state.

For MOOX implementation fidelity, direct original-executable control/data flow takes precedence over the simplified secondary sequence.

## Regression lock

`TestStrategicApplyOrderUsesPreGrowthResearchAndConstructionSnapshots` now verifies in one turn that:

```text
empire.research_progressed
colony.population_grew
colony.construction_progressed
```

occur in original apply order while:

```text
Research consumes exactly 4.5 pre-growth RP
Construction consumes exactly 3 pre-growth PP
```

and the final post-turn snapshot has the higher output produced by the larger Population.

## Schema impact

None.

```text
StateSchemaVersion   = 6
EconomySchemaVersion = 5
```

The change is resolver ordering plus documentation/tests; persistent state shapes remain unchanged.

## Next Research work

The turn-order conflict is closed. The next narrow Research topics are:

1. exact original Uncreative application RNG/initialization timing if practical from executable/save evidence;
2. hyper-advanced repeated-field level/cost state;
3. Advanced-start randomized/race-aware technology ownership using the existing `all` / `choose_one` / `fixed_one` policy.
