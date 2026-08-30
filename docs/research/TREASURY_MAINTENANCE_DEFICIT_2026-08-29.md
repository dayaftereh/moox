# Treasury Maintenance categories and deficit/scrap policy - 2026-08-29

**Open marker:** `docs/slices/_OPEN_TREASURY_MAINTENANCE_DEFICIT_2026-08-29.md`

## Goal

Extend the earlier Treasury settlement evidence by resolving the original six Maintenance buckets, relevant gross-income sources and the exact deficit/scrap path, then separate what current MOOX can implement from categories whose source systems do not yet exist.

## Gate 1 - Checkup + original analysis / reverse engineering

**Status:** Gate 4 QA passed on 2026-08-30; closing commits pending.

### Recovery state

- Starting branch: `main`.
- Starting HEAD: `d891c27` (`docs: close strategic blockade slice`).
- Starting tree: clean; `main` ahead of `origin/main` by 14 commits.
- Core `StateSchemaVersion`: 15.
- Economy ruleset schema: 7.

### Existing proven Treasury boundary

`docs/research/TREASURY_SETTLEMENT_2026-08-28.md` already establishes directly from the private MOO2 1.31 executable:

- persistent Treasury/credits at player `+0x32`;
- gross/current income at player `+0xAE`;
- current-turn net BC at player `+0xB2`;
- total Maintenance at player `+0xB4`;
- exactly six Maintenance component words at player `+0xB8..+0xC2`;
- `Compute_Player_Maintenance_` at VA `0xE2000` materializes those six components and sums them into `+0xB4`;
- `Apply_All_Player_Changes_` calls `Player_Maintenance_` VA `0xEE0B0`, then applies signed `+0xB2` to Treasury `+0x32`, then updates stats and resolves Research;
- `Pre_Import_Computing_` materializes Colony BC production/maintenance before player aggregation;
- New Game starts at 50 BC.

Current MOOX intentionally models only already-authoritative sources: adjusted Colony Tax + surplus-Food income, Building Maintenance + active-Freighter operating cost. The `Modeled` field names deliberately avoid claiming original completeness.

### Evidence discipline

Direct executable/save/data observations are labeled **original-observed**. Semantic labels inferred from call/data flow are **original-derived**. MOOX-only design choices are **modernization**. Unknown component identities remain unknown until evidence resolves them.
## Gate 1 findings

### Original six-component Maintenance ledger

**Evidence level: original-observed for storage/data flow; original-derived for semantic labels.**

`Compute_Player_Maintenance_` VA `0xE2000` clears the 12 bytes beginning at player `+0xB8`, independently materializes six signed words, then sums exactly those six words into total Maintenance `player+0xB4` at `0xE2619..0xE2634`.

Gate 1 resolves the component identities as follows:

| Player field | Original semantic component | Direct basis | Current MOOX dependency |
| --- | --- | --- | --- |
| `+0xB8` | Colony/Building Maintenance | sums `colony+0xF2`; `Colony_BC_Maintenance_` writes `+0xF2` | authoritative now |
| `+0xBA` | active Freighter operating cost | `(total Freighters - free/unused Freighters) / 2` | authoritative now, but current charge is incomplete |
| `+0xBC` | Ship command-point overage Maintenance | zero at/under command capacity; excess command use multiplied by command-overage rate | no canonical Ship/command-point state yet |
| `+0xBE` | Spy Maintenance | sums low-six-bit spy counts stored per opponent/player slot | no Spy system yet |
| `+0xC0` | outgoing tribute payments | sums `Tribute_Received_From_(recipient,payer)` for active tribute relations | no Tribute treaty state yet |
| `+0xC2` | Officer/Leader Maintenance | scans original 67-entry Officer pool and sums Officer cost helper results | no Leader/Officer system yet |

The original Maintenance total is therefore not an undifferentiated cost. Four of its six buckets depend on strategic systems MOOX does not yet own. Gate 2 must not fabricate those source records merely to populate Treasury telemetry.

### `+0xB8`: Colony/Building Maintenance

**Evidence level: original-observed.**

Within the Colony scan of `Compute_Player_Maintenance_`:

```text
0xE2141..0xE214F  require Colony owner == player and active Colony
0xE2151           read byte Colony +0xF2
0xE2159           add it to player +0xB8
```

`Colony_BC_Maintenance_` VA `0xE094F` directly materializes `colony+0xF2`. This is the same semantic Building-Maintenance source that current MOOX already derives from normalized owned Building definitions.

### `+0xBA`: active Freighters, including Settler reservations

**Evidence level: original-observed.**

The exact component calculation at `0xE23A6..0xE23C2` is:

```text
if player+0x38 <= 0:
    active = player+0x36
else:
    active = player+0x36 - player+0x38
component_BA = active / 2
```

The division is signed integer division/truncation; for the non-negative active-Freighter count this is `floor(active / 2)` BC.

Prior direct evidence already identifies:

```text
player+0x36 = total Freighters
available Food Freighters = player+0x36 - 5 * player+0x40
```

`Pass_Out_Imports_` then writes the residual/free Freighter amount into `player+0x38` after constrained Food allocation. Therefore `total - free` includes both:

- five reserved Freighters for each active interstellar Settler/Population transport; and
- Freighters actually consumed by Food transfer.

This closes the deliberate uncertainty left in `POPULATION_TRANSPORT_FREIGHTER_2026-08-29.md`: reserved Settler Freighters **do participate in the original active-Freighter Maintenance bucket**.

Current MOOX already owns all required state (`PopulationTransportFreightersReserved` and `FreightersUsed`) but currently computes only:

```go
float64(FreightersUsed) * 0.5
```

That misses active Settler reservations and produces fractional 0.5-BC charges for odd active counts, while the directly observed original bucket truncates the aggregate `/2` to whole BC. This is an actionable fidelity correction with no new subsystem dependency.

### `+0xBC`: Ship command-point overage

**Evidence level: original-observed for arithmetic/fields; original-derived label corroborated by embedded symbols/messages.**

`Compute_Player_Maintenance_` scans ordinary strategic Fleet records and builds command-point usage in `player+0x3C`, adding `(fleet+0x10)+1` for qualifying records. It then compares this against available command points `player+0x3A`:

```text
used == 0                 -> component BC = 0
used <= available         -> component BC = 0
used > available          -> excess = used - available
component BC = excess * command-overage rate
```

The rate is 10 in one special player condition; otherwise it is `12 -` an original global discount byte. Embedded strings in this same executable include `_command_summary_msg`, `_total_command_points_used_msg`, `_total_command_points_msg`, `_ship_maintenance_costs_msg` and `_starting_command_points_msg`.

MOOX schema 15 has only minimal strategic Fleet owner/role/at-system identity. It has no canonical ships or command-point usage/capacity. This component must remain deferred rather than attaching fake command strength to `StrategicFleet`.

### `+0xBE`: Spy Maintenance

**Evidence level: original-observed.**

At `0xE25B0..0xE25C8`, the routine loops player slots and calls VA `0x1026CF`, accumulating each return into `+0xBE`. The helper indexes the current player's byte array at `player+0xE57+other`, masks `& 0x3F`, and returns that low-six-bit count.

The executable independently exposes the matching spy APIs:

```text
Get_My_Spy_Number_
Get_Their_Spy_Number_
Set_My_Spy_Number_
Set_Their_Spy_Number_
Add_My_Spy_
Take_My_Spy_
...
```

MOOX has no Spy state, so this bucket remains deferred.

### `+0xC0`: outgoing Tribute

**Evidence level: original-observed call/data flow; original-derived semantic name from the adjacent embedded symbol `Tribute_Received_From_`.**

At `0xE25D1..0xE2615`, the routine scans other players. When the treaty/tribute word at `player+0x63F+2*other` is non-zero, it calls VA `0xE1FC7` and adds the result to `+0xC0`.

VA `0xE1FC7` computes a non-negative payment from payer gross income and the payer/recipient relation percentage, dividing by 20 and clamping negative results to zero. `Update_Player_Stats_` separately uses the same helper to accumulate **received** tribute into player `+0xE78` and then adds it to gross BC.

Current directed `DiplomaticRelation` deliberately only models blockade-relevant `neutral|hostile`; it does not own trade/tribute treaty economics. Both paid and received Tribute remain deferred to a later diplomacy/economy slice.

### `+0xC2`: Officer/Leader Maintenance

**Evidence level: original-observed scan/cost accumulation; original-derived semantic label corroborated by executable symbols.**

The opening section of `Compute_Player_Maintenance_` scans the original 67-entry, `0x3B`-stride Officer/Leader pool, filters by owner/state, calls VA `0x94A9D` for the maintained Officer cost and accumulates the result into `player+0xC2` at `0xE20A5`.

The executable exposes `Officer_Maintenance_`, `Calculate_Officer_Cost_`, `Officer_Cost_`, hire/dismiss/assign functions and the Officer screen. MOOX has no canonical Leader/Officer state, so this bucket remains deferred.

## Original gross/current income sources

**Evidence level: original-observed.**

`Update_Player_Stats_` VA `0xE2710` builds current-turn player aggregates from already-materialized Colony and Empire state. The Colony scan directly shows:

```text
Colony +0xE7 -> Food aggregate
Colony +0xE9 -> Production aggregate
Colony +0xEB -> Research aggregate
Colony +0xED -> BC/gross-income accumulator
```

`Colony_BC_Production_` VA `0xE03F1` directly writes `colony+0xED`, confirming it is current Colony BC production.

The BC accumulator also receives original sources that are not all currently authoritative in MOOX:

- Leader financial/economic skill contributions;
- trade-treaty income;
- signed research-treaty economic effects;
- tribute received, accumulated through `Tribute_Received_From_`;
- positive surplus Food sale value.

The final BC accumulator is capped to `0x7FFF` and stored at `player+0xAE`; `Compute_Player_Maintenance_` is then called and total `+0xB4` is subtracted before signed net BC is stored at `player+0xB2`.

Current MOOX can authoritatively represent Colony BC/Tax and surplus Food. Leader/treaty/tribute sources must stay deferred until their canonical producer state exists.

### Surplus-Food BC rounding boundary

**Evidence level: original-observed.**

`Update_Player_Stats_` materializes signed whole-Food surplus at `player+0xB0`. Only a positive surplus contributes BC. For a normal player, the code performs a signed right-shift/divide-by-two before adding the result to gross income; the Fantastic Traders condition bypasses that halving.

For original non-negative integer surplus this is:

```text
normal:            floor(surplus_food / 2) BC
Fantastic Traders: surplus_food BC
```

Current MOOX uses continuous Food and currently applies `sellableSurplus * 0.5` (or `*1.0` for Fantastic Traders). The direct executable now establishes an explicit original whole-BC boundary. Gate 2 should decide whether to close this fidelity gap now by applying `floor(sellableSurplus * saleRate)` at the sale-income boundary while retaining continuous `SurplusFoodSold` telemetry.

## Forced deficit / `Player_Maintenance_`

### Trigger

**Evidence level: original-observed.**

At VA `0xEE0B0`:

```text
0xEE0C9  read signed current-turn net player+0xB2
0xEE0D0  add persistent Treasury player+0x32
0xEE0D3  test result
0xEE0D5  if result >= 0, skip forced-maintenance liquidation
```

The forced-maintenance path therefore begins only when:

```text
Treasury + current_turn_net_BC < 0
```

This happens **before** `Apply_All_Player_Changes_` later applies `+0xB2` to persistent Treasury.

### Asset-aware staged liquidation, not a simple negative-balance clamp

**Evidence level: original-observed control flow plus embedded function-symbol corroboration.**

The deficit path builds strategic candidate/threat data, scans the current player's ordinary strategic Fleets and scores them against eligible other players. Fleet score uses owner/state filters and strategic-combat weighting; candidate values are normalized by each Fleet's command-point use. The routine then runs a long deterministic sequence of selection passes with different criterion combinations.

A direct pressure check compares roughly `4/3 * available command points` against used command points before the ship-selection sequence. The surrounding executable module exposes the explicit staged helpers:

```text
Kill_Research_Buildings_
Building_Worth_
Phase1_Building_
Phase2_Building_
Phase3_Building_
Kill_Buildings_
Kill_Spy_
Ship_Is_At_Colony_
Ship_Scrap_Value_
Kill_Ships_
Make_Scrap_Ships_Dead_
Kill_Leaders_
Player_Maintenance_
```

This proves the original forced-deficit policy is aware of multiple asset classes and uses ordered valuation/selection logic; it is not equivalent to merely deleting the cheapest Building or allowing Treasury to remain negative.

### Freighters are not a forced-deficit liquidation class

**Evidence level: original-observed negative evidence within the complete `Player_Maintenance_` body, corroborated by module separation.**

The complete `Player_Maintenance_` body from `0xEE0B0` to the next function at `0xEE4A1` contains no direct references to the Freighter fields `player+0x36`, `+0x38` or Settler-reservation field `+0x40`, and no call to the separately exposed Fleet-screen `Scrap_Freighters_` / `N_Freighters_Player_Can_Scrap_` functions.

Manual/UI Freighter scrapping exists elsewhere, but the forced Treasury-deficit routine does not directly use it as one of its liquidation stages.

### Determinism / tie-breaking

**Evidence level: original-observed at the structural level.**

`Player_Maintenance_` uses fixed array scans and an explicitly ordered series of liquidation-selection passes. There is no evidence in the analyzed deficit body of RNG selecting among equivalent liquidation classes. Exact packed priorities inside every Ship/Building helper are not normalized here because current MOOX lacks the underlying Ship/Spy/Leader systems and reproducing those helper enums would create dead legacy-shaped API.

The fidelity requirement for a future complete deficit implementation is therefore: deterministic stable scans plus original staged asset-class/valuation rules, resolved when the dependent subsystems exist.

## Gate 1 scope conclusion

The requested investigation resolves the original accounting categories, but it also shows that a **complete original deficit/scrap implementation is not currently dependency-safe**. Four of six Maintenance buckets and several gross-income categories depend on systems MOOX has intentionally not modeled yet; the forced-deficit routine itself liquidates/scores Buildings, Spies, Ships and Leaders.

Implementing a partial automatic scraper now would be observably non-original because it would choose from only the asset classes that happen to exist in MOOX. Gate 1 therefore recommends preserving the current explicit modernization that permits negative `BalanceBC` until the complete required asset model exists.

Two directly proven Treasury/Freighter fidelity corrections **are** dependency-safe today:

1. active Freighter operating cost must include both Food Freighters and five-Freighter Settler reservations, and aggregate whole-BC cost is `floor(activeFreighters / 2)`;
2. surplus-Food sale has an original whole-BC boundary: normal `floor(surplus/2)`, Fantastic Traders `surplus` for original whole-Food surplus.

## Proposed Gate 2 contract

### Schema / architecture

- Keep Core `StateSchemaVersion = 15`; no new persisted source-of-truth state is required.
- Keep Economy ruleset schema 7 unless implementation discovers a data-shape change is actually necessary.
- Do **not** add Ship-command, Spy, Tribute or Officer Maintenance fields as fake zero-valued runtime categories.
- Do **not** add partial automatic deficit/scrap behavior. Negative `BalanceBC` remains the explicit MOOX modernization until canonical Ship/Spy/Leader and treaty state can support the original staged policy.

### Freighter Maintenance fidelity correction

Use already-authoritative runtime state:

```text
active_freighters = PopulationTransportFreightersReserved + FreightersUsed
freighter_operating_cost_bc = floor(active_freighters * 0.5)
```

Equivalently for the current original ruleset: `floor(active_freighters / 2)`.

Clamp/validate the active count against owned Freighters through existing transport/logistics invariants; do not persist a duplicate active-Freighter count.

`EmpireFoodLogistics.FreighterOperatingCostBC` and the existing Treasury field/event continue to carry this now-complete modeled Freighter bucket.

### Surplus-Food sale fidelity correction

Keep continuous Food quantities and `SurplusFoodSold` for MOOX simulation/telemetry, but apply the proven whole-BC conversion boundary when materializing income:

```text
normal:            floor(SurplusFoodSold * 0.5)
Fantastic Traders: floor(SurplusFoodSold * 1.0)
```

This is an explicit original gameplay-rule rounding boundary, consistent with MOOX's domain-native float architecture: continuous quantities remain continuous until a directly evidenced rule rounds/converts them.

### Gate 3 regressions if accepted

At minimum:

- one active Freighter -> 0 BC Maintenance;
- two active Freighters -> 1 BC;
- odd active count truncates after aggregate calculation;
- an active interstellar Population transfer with five reserved Freighters and no Food transfer -> 2 BC Maintenance;
- five reserved + one Food Freighter -> 3 BC;
- Treasury settlement consumes the corrected full active-Freighter cost;
- one unit of normal surplus Food -> 0 BC, two -> 1 BC, three -> 1 BC;
- Fantastic Traders preserves one BC per whole surplus Food;
- fractional MOOX surplus remains continuous in `SurplusFoodSold` but BC income follows the explicit floor boundary;
- existing blockade, Population-transfer reservation and Food-priority regressions remain green;
- no automatic Building/Ship/Spy/Leader liquidation occurs in this slice.

## Gate 1 deliberate deferrals

- Ship command-point capacity/usage and overage Maintenance;
- Spy counts/Maintenance;
- Tribute treaty payment/receipt runtime;
- Officer/Leader Maintenance and economic bonuses;
- full trade/research treaty income/effects;
- exact Building/Ship/Spy/Leader forced-liquidation priorities and value formulas;
- complete original Treasury-deficit asset scrapping until all required canonical asset classes exist.
## Gate 2 decision - 2026-08-30

The proposed narrow contract was accepted without revision. Gate 3 is limited to the active-Freighter Maintenance aggregation/rounding correction and the surplus-Food whole-BC conversion boundary. The full original deficit liquidation policy and currently dependency-missing Maintenance/income categories remain deferred.

## Gate 3 implementation result - 2026-08-30

The accepted dependency-safe corrections are implemented in internal/game/food_logistics.go without Core or ruleset schema changes.

### Active Freighter Maintenance

After Food allocation, MOOX now computes remaining unused Food-capable Freighters and derives ctiveFreighters = ownedFreighters - unusedFreighters. This is equivalent to the original 	otal - free model and naturally includes both already-reserved Population-transfer Freighters and Freighters consumed by Food logistics. The Maintenance bucket is materialized with loor(activeFreighters * FreighterOperatingCostBC); for the original 0.5 rate this is loor(active/2) BC.

### Surplus-Food BC conversion

SurplusFoodSold remains a continuous simulation/telemetry quantity. SurplusFoodIncomeBC now applies the directly observed whole-BC boundary with loor(SurplusFoodSold * saleRate), including Fantastic Traders.

### Scope preserved

Negative Treasury remains representable and no automatic Building/Ship/Spy/Leader liquidation was added. Missing original Maintenance/income categories remain deferred until their canonical producer systems exist.
