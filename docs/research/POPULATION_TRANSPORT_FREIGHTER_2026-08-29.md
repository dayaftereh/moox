# Population transport / shared Freighter pool - 2026-08-29

**Slice status:** COMPLETE

**Open marker:** `docs/slices/_OPEN_POPULATION_TRANSPORT_FREIGHTER_2026-08-29.md`

**Workflow:** `docs/slices/README.md`

## Goal

Resolve the original MOO2 1.31 Population relocation / shared-Freighter interaction and add the smallest deterministic authoritative MOOX state needed for Human UI, built-in AI and remote agents to use the same server rule.

## Gate 1 - Checkup + original analysis / reverse engineering

**Status:** complete.

The original feature is named **Settler**, not Transport. The executable has a separate troop-transport system (`Move_Transports_`, `Unload_Transports_`, fleet type 2) which must not be confused with Population relocation.

### Original Settler functions

Direct symbol/disassembly evidence from the private MOO2 1.31 `Orion2.exe` reference:

- `Settler_ETA_` - VA `0xFED3F`
- `Room_For_Another_` - VA `0xFED77`
- `Pop_Tries_To_Settle_` - VA `0xFEE4C`
- `Add_Settler_To_Colony_` - VA `0xFF015`
- `Settle_Pop_` - VA `0xFF116`
- `Move_Settlers_` - VA `0xFF212`
- `N_Freighters_Player_Can_Scrap_` - VA `0xFF477`

The player record stride is `0xEA9`.

### Shared Freighter capacity

The original player fields are:

```text
player+0x36 = total Freighters
player+0x40 = active interstellar Settler count
player+0x42 = first packed Settler record
```

A packed original Settler record is four bytes. MOOX does not reproduce that storage format.

`Pass_Out_Imports_` and `N_Freighters_Player_Can_Scrap_` independently establish the shared-capacity equation:

```text
population_reserved_freighters = active_settlers * 5
freighters_available_for_food = total_freighters - population_reserved_freighters
```

`Pop_Tries_To_Settle_` checks the candidate launch as:

```text
(active_settlers + 1) * 5 <= total_freighters
```

and rejects a 26th active Settler. Therefore:

- one active interstellar Population transfer reserves exactly 5 Freighters;
- Population reservation has priority over automatic Food imports;
- the original maximum is 25 active interstellar Settlers per Empire;
- Food consumes only the remaining Freighter pool.

This closes the previously deferred `player[+0x36] - 5 * player[+0x40]` term from the insufficient-Freighter investigation.

### Same-system transfer

`Settle_Pop_` compares source and destination Star Systems.

When both Colonies are in the same System it calls `Add_Settler_To_Colony_` directly. No Settler record is created and no five-Freighter reservation is made.

For an interstellar move, `Settle_Pop_` appends the four-byte Settler record and increments `player+0x40` before removing the selected Population record from the source Colony.

### Population identity preserved

The packed Settler record preserves the original Population job bits. `Add_Settler_To_Colony_` restores those bits into the destination Population entry.

MOOX currently has aggregate single-race Population rather than original per-Pop race cohorts, so this slice preserves the implemented semantic job only:

- Farmer
- Worker
- Scientist

Mixed-race Population transfer remains part of the later race-aware Population-cohort slice rather than being guessed here.

### Source and destination capacity rules

`Pop_Tries_To_Settle_` rejects moving the last Population entry from the source Colony.

`Room_For_Another_` checks destination capacity and, when requested by the transfer path, scans already in-flight Settlers destined for the same Colony/race. Thus incoming Settlers reserve destination Population capacity before arrival and the UI cannot legitimately overbook a destination with multiple pending transfers.

MOOX mirrors this with one remaining Population unit required at the source and current Population plus inbound transfers checked against the authoritative destination capacity.

### ETA and original strategic FTL speed

`Settler_ETA_` calls `Parsecs_Between_Stars_` / `Parsecs_Between_Points_` and then uses `player+0x5A0` strategic FTL speed.

The original coordinate-to-parsec conversion is equivalent to:

```text
parsecs = ceil(hypot(dx, dy) / 30)
```

`Settler_ETA_` then uses at least speed 2, computes ceiling division and caps ETA at 15 turns:

```text
effective_speed = max(player_ftl_speed, 2)
eta = min(15, ceil(parsecs / effective_speed))
```

`Calc_Player_FTL_Speed_` (`0x575D6`) obtains the best Warp Drive through `Best_Warp_Drive_` (`0x5679E`). Direct table reads establish:

| Technology ID | Stable key | Original FTL speed |
| ---: | --- | ---: |
| 120 | `nuclear_drive` | 2 |
| 72 | `fusion_drive` | 3 |
| 96 | `ion_drive` | 4 |
| 11 | `anti_matter_drive` | 5 |
| 88 | `hyper_drive` | 6 |
| 95 | `interphased_drive` | 7 |

The same function adds +2 when original player byte `+0x8BC` is set. Existing direct race reverse engineering already identifies `+0x8BC` as the semantic `trans_dimensional` trait. MOOX therefore applies that +2 through Race Trait data rather than persisting the original byte.

### Arrival and blockade behavior

`Move_Settlers_` is called from `Next_Turn_Calc_` and decrements the ETA field every turn.

When ETA resolves, the original checks the destination System blockade bitmask. If the destination is blockaded, no Population is added. It also requires a still-valid destination Colony/owner before calling `Add_Settler_To_Colony_`.

Whether arrival succeeds or fails, the resolved Settler is removed and `player+0x40` is decremented. The five reserved Freighters therefore return to the shared pool immediately when the Settler resolves.

MOOX records explicit loss reasons for:

- `destination_blockaded`
- `destination_unavailable`
- `destination_full`

instead of reproducing packed notification bytes.

### Turn order relative to Food

Direct `Next_Turn_Calc_` disassembly fixes this relevant sequence:

```text
0x1376F  All_Colony_Calculations_
0x13774  Compute_Blockades_
0x13779  Move_Settlers_
0x1377E  All_Colony_Calculations_
```

Consequences:

1. Settlers launched during planning reserve their five Freighters before the first next-turn Food allocation.
2. `Move_Settlers_` resolves arrivals/losses before the second/final Colony calculation.
3. A Settler that resolves in that cycle releases its five Freighters in time for the final Food snapshot.

MOOX maps this by processing transfer commands before initial Food materialization, then resolving Population transfers after the pre-growth Research/Population/Construction sequence but before the final Colony/Food recalculation.

### Cancellation evidence

The relevant direct writes to the original player `+0x40` Settler counter are:

- `0xE47D3`: initialization to zero;
- `0xFF1F9`: increment on interstellar Settler launch;
- `0xFF37A`: decrement when `Move_Settlers_` resolves the record.

No separate in-flight Settler cancellation write/action was found in this path. MOOX therefore does **not** invent a cancel command in this slice.

### Freighter operating-cost boundary

This slice directly proves capacity reservation but did not establish a separate original Maintenance charge for the five reserved Settler Freighters. Existing MOOX Treasury settlement therefore continues charging the already-modeled 0.5 BC only for Food Freighters actually used by Food logistics. A Settler-specific operating-cost addition is deliberately not invented.

## Gate 2 - Implementation decision

**Status:** complete and accepted for implementation.

MOOX uses semantic, domain-native state instead of the original packed bytes:

```go
type PopulationTransfer struct {
    ID                  ID
    EmpireID            ID
    SourceColonyID      ID
    DestinationColonyID ID
    Job                 PopulationJob
    RemainingTurns      int
}
```

`GameState.PopulationTransfers` is authoritative and serialized. `StateSchemaVersion` advances from 10 to 11.

The command is:

```text
colony.transfer_population
```

Client payload contains source Colony, destination Colony and semantic job only. Empire authority, capacity, Freighter reservation, ETA and outcome remain server-owned.

Each command transfers exactly 1.0 MOOX Population unit, matching one original discrete Pop entry while retaining MOOX's domain-native `float64` Population architecture.

Same-System movement is immediate and creates no in-flight object. Interstellar movement creates one authoritative transfer and reserves five Freighters by derivation from active transfer count.

## Gate 3 - Implementation

**Status:** complete.

Implemented:

- Core `PopulationJob` and `PopulationTransfer` state;
- Core schema 11 validation and exact JSON roundtrip coverage;
- `colony.transfer_population` strict command decoder and authority checks;
- source last-Pop guard;
- Farmer/Worker/Scientist preservation;
- destination capacity including inbound transfers;
- same-System immediate transfer without Freighters;
- interstellar 5-Freighter reservation and 25-transfer maximum;
- original parsec/ETA rule;
- original Warp Drive speed mapping;
- Trans Dimensional +2 FTL semantic trait;
- turn-by-turn transfer progression;
- arrival, blockade/unavailable/full loss and reservation release;
- Food logistics capacity reduced by active Population reservations;
- observer telemetry:
  - `population_transport_freighters_reserved`
  - `freighters_available_for_food`
- Observer/replay events:
  - `colony.population_transferred`
  - `empire.population_transfer_started`
  - `empire.population_transfer_progressed`
  - `empire.population_transfer_arrived`
  - `empire.population_transfer_lost`
- authoritative `GameSession` end-to-end coverage.

The Wails/network layers still contain no gameplay legality.

## Regression coverage

Tests cover:

- five Freighters reserved before Food and zero remaining Food capacity with exactly five total Freighters;
- reservation released on arrival;
- same-System transfer requires no Freighters;
- job preservation;
- blockaded destination loses the Settler and releases Freighters;
- inbound Settlers reserve destination capacity;
- Nuclear/Interphased Drive speeds;
- Trans Dimensional +2 FTL;
- ETA ceiling/cap behavior;
- Core schema-11 transfer roundtrip;
- Seat-authoritative session command and Observer start/arrival lifecycle.

## Gate 4 - Follow-up QA + commit

**Status:** complete.

Final checks passed:

```text
gofmt on changed Go files        PASS
go test ./...                    PASS
go vet ./...                     PASS
git diff --check                PASS
```

Closing commit subject:

```text
feat: add population Freighter transfers
```

The open-slice marker is removed in the same closing commit. No push is part of this slice.