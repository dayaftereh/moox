# Treasury settlement research - 2026-08-28

## Scope

This checkpoint records the directly verified Master of Orion II 1.31 Treasury/statistics path and the deliberately narrower Master of Orion X runtime ledger built from economy components that are already normalized and simulated.

MOOX keeps BC values as domain-native `float64`. The original 16/32-bit storage widths below are clean-room evidence for behavior and layout, not MOOX numeric-architecture requirements.

## Original executable evidence

Reference executable SHA-256:

`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`

### Persistent Treasury

`Apply_All_Player_Changes_` starts at VA `0xE4F49`.

For each active player it performs the following sequence:

1. calls `Player_Maintenance_` at VA `0xEE0B0`;
2. reads signed current-turn net BC from player `+0xB2`;
3. adds that value to the 32-bit player field at `+0x32` (`0xE4F8E..0xE4F95`);
4. calls `Update_Player_Stats_` at VA `0xE2710`;
5. then calls the research-breakthrough path at VA `0xE44E0`.

The player `+0x32` field is therefore the persistent Treasury / credit balance, and the BC apply occurs before Research breakthrough resolution in this turn path.

### Gross, Maintenance and Net

The end of `Update_Player_Stats_` proves the current-turn BC relationship:

- player `+0xAE`: gross/current income before Maintenance;
- `Compute_Player_Maintenance_` at VA `0xE2000` materializes Maintenance;
- player `+0xB4`: total Maintenance;
- player `+0xB2`: gross minus Maintenance.

Relevant instructions:

```text
0xE2A4F  store gross/current BC at player+0xAE
0xE2A56  call Compute_Player_Maintenance_
0xE2A5B  read player+0xB4
0xE2A62  subtract Maintenance from gross accumulator
0xE2A64  store result at player+0xB2
```

`Compute_Player_Maintenance_` independently materializes six component words at player `+0xB8 .. +0xC2`. Its final loop at `0xE2619..0xE2634` sums exactly those six components into the total at `+0xB4`.

This is important for MOOX: Freighter operating cost is not the original game's entire Maintenance system. A Treasury implementation must remain explicit about which Maintenance categories are currently modeled.

### Colony money values are materialized before player stats

`Pre_Import_Computing_` at VA `0xE1D59` performs the Colony-side current-turn calculations and calls:

- `Colony_BC_Production_` at VA `0xE03F1`;
- `Colony_BC_Maintenance_` at VA `0xE094F`.

Those values therefore exist before `Update_Player_Stats_` aggregates player income/Maintenance.

### New Game starts at 50 BC

Direct executable scanning finds the New Game initialization write at VA `0x12C86` inside the `Init_Players_` path:

```text
mov DWORD PTR [player+0x32], 0x32
```

So the directly verified original starting Treasury is **50 BC**.

## Original SAVE10.GAM cross-check

Private reference save:

- file: `SAVE10.GAM`
- SHA-256: `ECE2EB06D782078DD0A6F746020A05691355303CEB02BBFBBE2233E987272BE1`
- stardate: 3500.0
- documented player-array start: `0x1AA0F`
- stride: `0xEA9`

For the Treasury/stat fields in this save, the serialized locations are one byte earlier than their runtime-structure offsets. Reading the five active players with that mapping gives:

| Player | Treasury | Gross | Maintenance | Net |
| ---: | ---: | ---: | ---: | ---: |
| 0 | 50 | 11 | 3 | 8 |
| 1 | 50 | 10 | 3 | 7 |
| 2 | 50 | 8 | 3 | 5 |
| 3 | 50 | 10 | 3 | 7 |
| 4 | 50 | 13 | 3 | 10 |

Every active player independently satisfies `gross - maintenance = net`, and every Treasury is 50 BC. This cross-check is consistent with the executable path above.

## MOOX state model

Core `StateSchemaVersion = 9` adds a semantic per-Empire Treasury state:

```go
type EmpireTreasuryState struct {
    BalanceBC                 float64
    TaxIncomeBC               float64
    SurplusFoodIncomeBC       float64
    GrossIncomeBC             float64
    BuildingMaintenanceBC     float64
    FreighterOperatingCostBC  float64
    TotalModeledMaintenanceBC float64
    NetModeledIncomeBC        float64
}
```

`BalanceBC` may be negative and must remain finite. Materialized income/Maintenance components must be finite and non-negative; `NetModeledIncomeBC` may be negative.

`InitializeNewGameTreasury` is intentionally separate from technology initialization and applies the directly verified **50 BC** starting balance to each Empire. This keeps future New Game orchestration domain-composable instead of hiding economy initialization inside Research code.

## Current authoritative settlement

`EconomyResolver.settleTreasury` runs once per strategic resolution after:

1. current-turn Colony economy recalculation;
2. current-turn Food/Freighter logistics materialization;

and before:

3. Research;
4. Population Growth/Starvation;
5. Construction.

The current MOOX ledger includes only money categories whose underlying gameplay state is already modeled:

### Gross modeled income

```text
TaxIncomeBC
+ SurplusFoodIncomeBC
= GrossIncomeBC
```

`TaxIncomeBC` is the sum of current-turn `Colony.AdjustedEconomy.TaxBC` for the Empire. `SurplusFoodIncomeBC` comes from the already-materialized Empire Food/Freighter logistics snapshot.

### Modeled Maintenance

```text
BuildingMaintenanceBC
+ FreighterOperatingCostBC
= TotalModeledMaintenanceBC
```

Building Maintenance is summed from normalized owned Building definitions. Freighter operating cost comes from the current-turn Food/Freighter logistics snapshot.

### Settlement

```text
NetModeledIncomeBC = GrossIncomeBC - TotalModeledMaintenanceBC
BalanceBC += NetModeledIncomeBC
```

The resolver emits `empire.treasury_settled` with previous balance and the complete current semantic Treasury snapshot. Session Observer/replay history receives the event through the normal strategic resolution stream.

Because settlement uses the pre-growth/pre-construction snapshot, a Building completed later in the same turn does **not** retroactively charge Maintenance. Regression coverage proves that a Holo Simulator completed this turn costs 0 BC in the current settlement and 1 BC in the following turn.

## Deliberately deferred original categories

The original six-component Maintenance total contains categories beyond what MOOX currently simulates. This checkpoint does not guess or fabricate the missing categories.

Still deferred include, as applicable to the original six-component ledger:

- ship-related Maintenance / command consequences;
- leader/officer costs;
- spy/security-related costs;
- diplomatic/trade/tribute income and expense sources not yet normalized into the MOOX economy;
- the exact original deficit/scrapping policy reached through `Player_Maintenance_`.

The current state/event fields use the word `Modeled` for totals/net specifically so UI, AI and replay consumers can distinguish the implemented ledger from a claim of complete original-MOO2 finances.

## Verification

Implemented regression coverage includes:

- direct 50-BC New Game Treasury initialization;
- finite/negative Balance validation;
- exact schema-9 Treasury save/load round-trip;
- component arithmetic and settlement event payload;
- full `internal/game` regression suite with Treasury inserted before Research;
- Building-completion Maintenance starts only on the following turn;
- Session Observer history contains the authoritative Treasury settlement and resulting balance.

## Next economy investigation

The **exact original insufficient-Freighter import priority** is now resolved in `INSUFFICIENT_FREIGHTER_PRIORITY_2026-08-28.md`: constrained imports use original-style round-robin deficit allocation. The next Food/Freighter expansion can therefore move to blockade eligibility/effects and Population transport without carrying the old allocation ambiguity.