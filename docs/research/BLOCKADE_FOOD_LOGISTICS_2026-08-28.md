# Blockade effects on Food/Freighter logistics - MOO2 1.31 evidence

Date: 2026-08-28

## Result

Direct `Orion2.exe` 1.31 analysis resolves the Food-logistics meaning of a blockade strongly enough to model it without guessing:

- blockade state is materialized **per star system and target Empire**;
- a blockaded Colony is excluded from the Empire Food **import** pool;
- a blockaded Colony is also excluded from the Empire Food **surplus/export** pool;
- therefore its local Food surplus cannot feed other Colonies and is not part of Empire surplus-Food sale income;
- its local Food shortage remains unresolved by Freighters and can feed the normal starvation calculation;
- `Compute_Blockades_` runs between two Colony-calculation passes, so a newly computed blockade affects the second Food calculation in the same `Next_Turn_Calc_` sequence.

MOOX now models the smallest authoritative state needed by those proven rules: `StarSystem.BlockadedEmpireIDs`. The Food resolver consumes that state. Strategic fleets and diplomacy are not yet canonical MOOX state, so this slice deliberately does not invent a Fleet-derived blockade producer.

## Original executable

Reference executable:

- `C:\ASH\Temp\mastori2\Orion2.exe`
- MOO2 version: 1.31
- SHA-256: `7AE2AC2E5904CA330009AF2827279D889906B0B9B7A8854C38EB707A56E955B5`

Relevant original routines:

| Routine | VA | Finding |
| --- | ---: | --- |
| `Colony_Is_Blockaded_` | `0xDF8C1` | resolves Colony owner + Star System blockade bit |
| `Pass_Out_Imports_` | `0xDF8F0` | excludes blocked deficit and surplus Colonies from Food logistics |
| `Pre_Import_Computing_` | `0xE1D59` | materializes local Colony economy immediately before import distribution |
| `Do_Colony_Calculations_` | `0xE2B31` | runs Colony calculations that reach Food logistics |
| `Player_Is_Hostile_To_Player_` | `0xD8DE1` | independently confirms relation values `4..6` as hostile |
| `Compute_Blockades_` | `0xE5097` | rebuilds per-System target-Empire blockade masks from strategic presence |
| `Next_Turn_Calc_` | `0x136B3` | brackets `Compute_Blockades_` with two Colony-calculation passes |

## Blockade predicate

`Colony_Is_Blockaded_` at `0xDF8C1` follows this data path:

```text
Colony +0x02
  -> planet/system lookup
  -> Star System record (stride 0x71)
  -> byte +0x2A
  -> shift by Colony owner/player index
  -> bit & 1
```

The important semantic result is that original `system+0x2A` is a bitmask indexed by the **target Empire/player**. A Colony is blockaded when the bit belonging to its owner is set for that Colony's star system.

This is system state, not an independently stored per-Colony boolean.

## Import exclusion

Inside `Pass_Out_Imports_`, a Colony with local Food deficit reaches the blockade test before it is appended to the deficit/import list.

The relevant branch is around `0xDF991..0xDF9E0`:

1. calculate local `food production - food maintenance`;
2. for a negative result, resolve the Star System;
3. read the same `system+0x2A` target-Empire bit;
4. when the bit is set, jump to the next Colony;
5. only an unblocked deficit Colony is appended to the import list and contributes to the Empire import demand pool.

Therefore a blockaded deficit Colony cannot receive Food through the Empire Freighter network.

## Export/surplus exclusion

The non-negative branch repeats the same blockade predicate around `0xDF9E2..0xDFA21`.

The original writes the Colony-local surplus bookkeeping, then:

1. resolves the Star System;
2. tests the owner bit in `system+0x2A`;
3. skips the Empire surplus-pool addition when blockaded;
4. only an unblocked Colony contributes its positive Food delta to the shared Empire surplus pool.

Therefore blockade affects **both directions** of Food logistics.

A blocked local surplus is still a real local surplus, but it is not available as an Empire transfer source. Because the original omits it from the Empire surplus pool, MOOX also excludes it from `SurplusFoodSold` / surplus-Food BC income.

## Materialization timing

`Next_Turn_Calc_` proves the order directly:

```text
0x1376F  call Do_Colony_Calculations_ (0xE2B31)
0x13774  call Compute_Blockades_       (0xE5097)
0x1377E  call Do_Colony_Calculations_ (0xE2B31)
```

`Colony_Calculation_` at `0xE2A70` reaches:

```text
0xE2A89  call Pre_Import_Computing_ (0xE1D59)
0xE2A93  call Pass_Out_Imports_      (0xDF8F0)
```

So the new blockade mask is materialized between two Colony-economy passes and is consumed by the second pass during the same next-turn calculation.

MOOX currently receives the blockade as already-materialized authoritative `StarSystem` state. When strategic Fleet/Diplomacy resolution is implemented, its blockade producer must run before the economy snapshot that consumes this state.

## What `Compute_Blockades_` already proves

The current disassembly pass is sufficient to establish the shape of the future producer without pretending the full Fleet model is normalized:

- System `+0x2A` is rebuilt as the target-Empire blockade mask;
- the neighboring per-player bytes beginning at System `+0x2B` are also cleared/rebuilt as blockader-related masks;
- strategic fleet records use a `0x81`-byte stride in this routine;
- fleet owner/index is read from `+0x63`;
- system/star index is read from `+0x65`;
- the target Empire must have presence at the system before its blockade bit is considered;
- diplomacy relation byte `player + otherPlayer + 0x627` must be in `4..6` inclusive for the hostile path that sets blockade state.

`Player_Is_Hostile_To_Player_` at `0xD8DE1` independently checks the same relation byte and classifies `4..6` as hostile.

Some fleet-record status predicates used by `Compute_Blockades_` are visible (`+0x64`, `+0x11`, and related state), but their semantic identities are not normalized yet. MOOX therefore does not manufacture Fleet objects or diplomacy states merely to reproduce these bytes.

## MOOX state model

Core state schema 10 adds:

```go
StarSystem.BlockadedEmpireIDs []core.ID
```

The list is:

- authoritative simulation state;
- sorted and unique;
- validated against known Empires;
- saved/replayed with the rest of the deterministic state;
- deliberately semantic rather than an 8-bit copy of the original memory layout.

This keeps the original rule while following MOOX's architecture principle that legacy storage width is evidence, not a runtime requirement.

## Food resolver behavior

For each Empire, MOOX now partitions its Colonies into:

- **eligible** Colonies: system does not blockade that Empire;
- **blocked** Colonies: system contains that Empire in `BlockadedEmpireIDs`.

Only eligible Colonies contribute to:

- Empire transfer-source Food surplus;
- Empire import demand;
- `FreightersRequired`;
- `FreightersUsed`;
- surplus Food that may be sold for BC.

Blocked Colonies retain their local values. Observer/replay projection now includes a per-Colony `blockaded` flag plus aggregate:

- `blocked_food_surplus`;
- `blocked_food_shortage`.

`FoodUnmet` still includes blocked shortage, so starvation remains driven by the Colony's unresolved local Food deficit.

## Tests

The implementation locks these boundaries:

- System blockade state round-trips through Core schema 10;
- zero, duplicate, unsorted and unknown Empire references are rejected by state validation where applicable;
- a blockaded surplus Colony exports no Food and its surplus produces no sale income;
- a blockaded deficit Colony imports no Food and its shortage remains unmet/starvation-eligible;
- unblocked Colonies keep the existing original-style round-robin import allocation;
- the existing complete test suite remains green.

## Deliberately deferred

This checkpoint does **not** yet implement:

- canonical strategic Fleet state;
- canonical diplomacy/relationship state;
- automatic `Compute_Blockades_` derivation from fleets and relations;
- Population transport competition for the shared Freighter pool;
- exact mixed-population-cohort import priority passes.

The next Food/Freighter slice is Population transport through the shared Freighter pool. Full Fleet-derived blockade computation should be attached later when the strategic Fleet/Diplomacy state exists, using `StarSystem.BlockadedEmpireIDs` as its authoritative output.