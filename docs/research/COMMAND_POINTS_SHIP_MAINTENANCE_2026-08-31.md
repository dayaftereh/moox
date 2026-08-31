# Command Points and ship Maintenance - Slice 05 evidence

Date: 2026-08-31

Status: **closed; Gates 1-4 complete**

Open marker: `docs/slices/_OPEN_COMMAND_POINTS_SHIP_MAINTENANCE_2026-08-31.md`

Planned slice: `docs/slices/PLANNED_05_COMMAND_POINTS_SHIP_MAINTENANCE.md`

## Objective

Establish original MOO2 1.31 Command-Point capacity/usage and ship-overage Maintenance semantics, then define the narrow deterministic MOOX integration boundary without pulling Spy/Leader/treaty Maintenance or full deficit liquidation into this slice.

## Fresh repository check

- Starting commit: `25cba83` (`docs: close combat fleet movement slice`).
- `main` was 28 commits ahead of `origin/main` at slice start.
- The working tree was clean and there were no existing `_OPEN_*.md` markers.
- Core `StateSchemaVersion` is 19; economy ruleset schema is 7; ship-hulls schema is 3.
- Slice 05 created exactly one recovery marker and this permanent evidence document.
- No gameplay/ruleset code has changed during Gate 1. No commit or push is part of Gate 1.

## Reference material and confidence labels

Private reference executable:

```text
C:\ASH\Temp\mastori2\Orion2.exe
```

Do not distribute the private reference binary.

Primary text evidence:

```text
reference/original/text/help/block_0000.ascii.txt
```

Fresh executable inspection used `internal/moo2exe.ReadObject` plus a temporary `x86asm` helper outside the repository. The helper is deleted before Gate-1 closure.

Evidence labels below:

- **direct original help** - wording/table in original `HELP.LBX`;
- **direct original code** - fresh 1.31 executable disassembly/data flow;
- **original-derived semantic interpretation** - semantic name inferred only where multiple direct original observations make the mapping unambiguous;
- **MOOX boundary** - deliberate current implementation limit, not claimed as original behavior.

## Finding 1 - the original Command Point window defines reserve, total capacity and the monetary penalty

Evidence level: **direct original help**.

`HELP.LBX` record 288 (`Command Points Window`) states:

- every ship removes points from the reserve;
- every station adds points to the reserve;
- Warlord gains 2 Command Points per Colony;
- Imperium increases generated Command Points by 50%;
- the first displayed number is the current reserve;
- the second number in parentheses is total Command Points;
- a negative reserve costs **10 BC per negative point per turn** for the player-facing rule.

This directly establishes the strategic UI semantics:

```text
reserve = capacity - used
negative reserve => overage = used - capacity
overage Maintenance = overage * 10 BC/turn      // player-facing rule
```

No fractional Command Points or rounding rule is implied by this surface.

## Finding 2 - hull Command-Point usage is exactly size index + 1

Evidence level: **direct original help + direct original code**.

`HELP.LBX` record 289 (`Command_Point_Table`) gives:

```text
Frigate     -1
Destroyer   -2
Cruiser     -3
Battleship  -4
Titan       -5
Doom Star   -6
```

Fresh `Compute_Player_Maintenance_` disassembly at VA `0xE2000` scans concrete `0x81`-byte Ship records. For every qualifying Ship it executes the semantic equivalent of:

```text
used += unsigned(ship+0x10) + 1
```

The same original hull byte is already proven by the Military Ship slice as the 0..5 hull-size index. Therefore the authoritative per-warship CP usage is:

```text
commandPoints(ship) = hull.SizeIndex + 1
```

MOOX should derive this from the authoritative hull definition / built Ship `HullID`. It should **not** persist a second mutable `CommandPointCost` on Ship or ShipDesign.

## Finding 3 - Colony Ship, Outpost Ship and original troop Transport each consume 1 CP

Evidence level: **direct original code**.

The CP usage scan adds `ship+0x10 + 1` regardless of whether `ship+0x11` is zero or a fixed special type. The special byte is used only to separate ordinary versus special counts in the Command Point summary UI.

Prior direct original evidence fixes:

```text
ship/design+0x11 = 0  ordinary/combat ship
ship/design+0x11 = 1  Colony Ship
ship/design+0x11 = 2  Transport
ship/design+0x11 = 4  Outpost Ship
```

Fresh disassembly of the fixed special-design loaders confirms all three set:

```text
design+0x10 = 0
```

Therefore:

```text
Colony Ship  = 1 CP
Transport    = 1 CP
Outpost Ship = 1 CP
```

Current MOOX has authoritative fixed Colony/Outpost special Fleets but no troop-Transport subsystem. Gate 3 can count each extant Colony/Outpost special Fleet as 1 CP. It must **not** treat Population-transfer Settlers/Freighter reservations as the original troop Transport merely to manufacture CP usage.

## Finding 4 - stationary and in-transit ships consume the same CP

Evidence level: **direct original code**.

The original usage scan accepts qualifying Ship records with:

```text
owner == player
ship+0x64 <= 2
ship+0x80 == 0
```

Earlier movement evidence establishes:

- `ship+0x64 == 0` = stationary at a Star;
- statuses 1/2 are the normal strategic movement/transit states advanced by `Move_All_Ships_Toward_Stars_`.

Therefore supported stationary and in-transit Ships count identically. Command Point usage is an Empire fleet-size constraint, not a location/range constraint.

The exact semantic name of legacy `ship+0x80` is not required for the current MOOX surface; extant validated MOOX Ship/special-Fleet entities are the supported equivalent. Do not invent a new persisted byte/state for it.

## Finding 5 - the original base capacity is 5 CP

Evidence level: **direct original code**.

`Compute_Player_Maintenance_` builds the capacity accumulator and contains the direct constant step:

```text
mov eax, starBaseCount
add eax, 5
```

The final accumulator adds station, Communications, Warlord and officer/government terms around this constant. Thus every Empire begins this calculation with:

```text
base command-point capacity = 5
```

## Finding 6 - orbital stations provide +1 / +2 / +3 CP

Evidence level: **direct original help + direct original code**.

`HELP.LBX` record 289 gives:

```text
Star Base      +1
Battlestation  +2
Star Fortress  +3
```

Fresh Colony-record scanning inside `Compute_Player_Maintenance_` checks the original building-ownership bytes corresponding to the already-normalized production IDs:

```text
colony+0x15E = building flag 40 = Star Base
colony+0x13E = building flag 8  = Battlestation
colony+0x15F = building flag 41 = Star Fortress
```

The arithmetic is exactly:

```text
+ 1 * starBaseCount
+ 2 * battlestationCount
+ 3 * starFortressCount
```

The original branch structure expects these station tiers to be mutually exclusive on a Colony.

### Dependency note - MOOX station replacement is not yet normalized

Direct original HELP descriptions state:

- Battlestation **replaces a Star Base**;
- Star Fortress **replaces Battlestations and Star Bases**.

Current MOOX generic building completion only appends the completed building and buildability filters only the exact already-owned building. It can therefore currently create an invalid Colony containing multiple station tiers.

For CP fidelity, Gate 2 should accept a narrow orbital-station-family replacement dependency: higher station completion removes lower station tiers and lower/redundant station choices are suppressed when a higher tier exists. No station weapons, tactical defense, scan bonus or unrelated combat behavior belongs in Slice 05.

## Finding 7 - Communications technologies add +1 / +2 / +3 per station

Evidence level: **direct original help + direct original code**.

Original help states:

```text
Tachyon Communications     +1 CP / base
Subspace Communications    +2 CP / base
Hyperspace Communications  +3 CP / base
```

Fresh executable checks the exact Technology-status bytes:

```text
Tech 91  Hyperspace Communications -> 3
else Tech 176 Sub Space Communications -> 2
else Tech 180 Tachyon Communications -> 1
else -> 0
```

The selected Communications value is multiplied by the total number of owned orbital command stations, independent of station tier:

```text
communicationsBonus = commPointsPerStation * (starBases + battlestations + starFortresses)
```

If several of these Technologies are known, the original priority above selects the strongest supported bonus.

## Finding 8 - Warlord provides +2 CP per owned active Colony

Evidence level: **direct original help + direct original code**.

Original Warlord help explicitly says:

```text
Each colony produces 2 command points.
```

The original race block mapping already proves `player+0x8BD` is the Warlord flag. During the owned-active-Colony scan, `Compute_Player_Maintenance_` adds 2 for every Colony when that flag is set.

Current MOOX race-trait normalization already contains the semantic `warlord` ability, but `RaceEconomyModifiers` does not yet expose it. Slice 05 can add a derived `Warlord` boolean to the rules projection without changing race-trait data.

## Finding 9 - Imperium adds 50% after the other generated-CP sources, with integer truncation

Evidence level: **direct original help + direct original code**.

Original help states Imperium increases total generated Command Points by 50%.

Fresh code proves the CP routine checks materialized government byte:

```text
player+0x89F == 3
```

and then performs the signed-integer equivalent of:

```text
capacity += floor(capacity / 2)
```

for the nonnegative generated-CP domain.

The government provenance is also closed by fresh inspection:

- `player+0x89F` is initialized from the starting government as an even-valued materialized government state;
- the government-evolution grant path increments this byte;
- existing research evidence maps Government evolution eligibility, including **Imperium Technology 92 only for Dictatorship**.

Thus value 3 is the evolved Dictatorship/Imperium state. For current MOOX, the safe semantic derivation is:

```text
Imperium active = base government trait == government_dictatorship
               && Empire knows Technology 92
```

No freely mutable persisted legacy-government enum is needed solely for Slice 05.

## Finding 10 - original capacity also contains an officer-derived CP term, but Leaders are out of scope

Evidence level: **direct original code for presence of the term; exact leader-skill semantics intentionally deferred**.

Before Colony capacity calculation, `Compute_Player_Maintenance_` scans original officer records, derives a Command-related value from eligible owned officers and keeps a maximum term. That term is added to generated CP before the Imperium +50% step.

MOOX has no authoritative Leader/Officer subsystem yet. Therefore current supported capacity must use zero for this source and leave a clear future hook. Slice 05 must not invent fake Leader state or broaden into Officer Maintenance.

Because Imperium applies after this term in original code, a future Leader implementation must insert its CP bonus before the +50% Imperium step.

## Finding 11 - exact supported capacity formula

Evidence level: **direct original code, with unsupported Leader term isolated**.

For the current supported MOOX surface:

```text
capacityBeforeImperium =
    5
  + 1 * starBaseCount
  + 2 * battlestationCount
  + 3 * starFortressCount
  + communicationsPerStation * stationCount
  + (Warlord ? 2 * ownedActiveColonyCount : 0)
  + leaderCommandBonus                    // 0 until Leader subsystem exists

capacity = capacityBeforeImperium
if Imperium:
    capacity += floor(capacityBeforeImperium / 2)
```

`stationCount` is the count of Colonies with one legal Star Base/Battlestation/Star Fortress station after replacement normalization.

## Finding 12 - human/player overage is exactly 10 BC per excess CP; original AI has a separate difficulty branch

Evidence level: **direct original help + direct original code**.

Fresh code at the overage branch:

```text
if used == 0 or used <= capacity:
    shipMaintenance = 0
else:
    excess = used - capacity
    if player+0x28 == 100:
        rate = 10
    else:
        rate = 12 - globalByte[0x21CB0]
    shipMaintenance = excess * rate
```

Prior original research proves `player+0x28 == 100` is the human-player sentinel. It also proves global `0x21CB0` belongs to a broader non-default AI/game-setting layer and defaults to 0, but its full public difficulty semantics are not yet normalized in MOOX.

Therefore:

- the directly supported player-facing rate is **10 BC per excess CP**;
- no rounding is involved: CP, excess and rate are integers;
- no separate cap was observed before the signed-word Maintenance bucket is stored;
- the original AI `12 - difficulty/global setting` branch should remain explicit/deferred until AI/difficulty semantics are authoritative.

Current pure Game arithmetic must not depend on Session controller type merely to guess this original NPC branch.

## Finding 13 - CP Maintenance timing is before current-turn ship Construction completion

Evidence level: **direct original turn-order code + current MOOX ordering**.

Original `Next_Turn_Calc_` order already established:

```text
Make_Scrap_Ships_Dead_
...
Move_All_Ships_Toward_Stars_
...
Apply_All_Player_Changes_
    Player_Maintenance_
    Update_Player_Stats_
    Research breakthrough
Apply_All_Colony_Changes_
    ...
    Apply_Production_
```

Consequences:

1. Ships in original scrap/dead statuses 7/8 are removed by `Make_Scrap_Ships_Dead_` before Maintenance.
2. Normal transit movement occurs before Maintenance, but stationary/transit Ships both consume CP, so movement itself does not change usage.
3. Current-turn Construction completion happens **after** Player Maintenance. A ship completed this turn first affects CP Maintenance on the following turn.

Current MOOX already settles Treasury before Research/Population/Fleet-transit/Construction. Keeping ship-overage settlement in the existing `settleTreasury` phase therefore preserves the important Construction boundary without moving the resolver phase.

MOOX currently advances strategic Fleet transit after Treasury rather than before it; for Slice 05 this is behaviorally neutral because supported stationary and in-transit Ships have identical CP usage. Future battle destruction/encounter slices must re-check this boundary when they add ship loss between phases.

## Finding 14 - current MOOX data/state surfaces and gaps

Current authoritative inputs already available:

- military `Ship.Spec.HullID` and normalized hull `SizeIndex`;
- fixed Colony/Outpost special `StrategicFleet` identities;
- Colony ownership/buildings;
- Star Base/Battlestation/Star Fortress normalized building IDs and Technology IDs;
- Empire `KnownTechnologyIDs` for Communications and Imperium;
- semantic base race government trait;
- semantic normalized Warlord race ability data;
- Treasury settlement before Construction.

Missing pieces needed by this slice:

- a materialized Command Point Empire snapshot for Observer/save/load/UI auditing;
- ship-overage Maintenance in Treasury;
- normalized Command Point economy constants/producers;
- Warlord projection in `RaceEconomyModifiers`;
- narrow station-tier replacement normalization.

Leader/Officer CP, original troop Transport, NPC/difficulty overage rates and automatic deficit ship liquidation remain dependency-unsafe/out of scope.

## Gate 2 accepted implementation contract

The following Gate-1 recommendation was **accepted in Gate 2 on 2026-08-31**. Gate 3 must implement this contract without silently broadening the deferred boundaries.

### 1. Core schema 20

Add a materialized Empire Command Point snapshot:

```go
type EmpireCommandPoints struct {
    Capacity int `json:"capacity"`
    Used     int `json:"used"`
}
```

under `Empire.CommandPoints`.

Do not persist redundant reserve/overage truth. Derive:

```text
reserve = capacity - used
overage = max(0, used - capacity)
```

Add to `EmpireTreasuryState`:

```text
ShipCommandMaintenanceBC float64
```

and include it in `TotalModeledMaintenanceBC` / `NetModeledIncomeBC`.

Schema bump is justified because Empire/Treasury persisted semantic state changes.

### 2. Economy ruleset schema 8; no ship-hulls schema bump

Add a normalized `command_points` rule section to `economy.json` containing original-proven constants/semantic IDs:

```text
base_capacity = 5
fixed_special_ship_points = 1
warlord_points_per_colony = 2
station building ids + base points: star_base 1, battlestation 2, star_fortress 3
communications technology ids + per-station points: 180->1, 176->2, 91->3
Imperium technology id 92 + required base government dictatorship + 50%, floor
standard/player overage rate = 10 BC per point
```

Keep military per-Ship usage derived from existing `ShipHull.SizeIndex + 1`; do not duplicate it in `ship_hulls.json` and do not bump ship-hulls schema solely for CP.

The original NPC `12 - global setting` rule remains documented but is not normalized as active gameplay until that setting/controller boundary exists.

### 3. Authoritative CP derivation helper

One pure helper should calculate for an Empire from authoritative state/rules:

- military Ships by concrete `Ship` ownership and hull definition;
- extant Colony special Fleets = 1 CP each;
- extant Outpost special Fleets = 1 CP each;
- no Population-transfer/Freighter CP;
- no location/transit modifier;
- station capacity from owned active Colonies;
- Communications highest-known bonus;
- Warlord Colony bonus;
- Imperium +50% last;
- Leader bonus = 0 with explicit future hook.

It should reject unknown Hull IDs or structurally invalid ownership rather than silently approximating.

### 4. Narrow orbital station replacement dependency

To keep station capacity and Building Maintenance consistent with original legal state:

- completing Battlestation removes Star Base;
- completing Star Fortress removes Star Base and Battlestation;
- buildability suppresses lower/redundant station tiers when an equal/higher tier is already owned;
- no tactical station combat/scanner behavior is added.

This is included only because original CP generation assumes one station tier per Colony and HELP directly proves the replacement relationship.

### 5. Treasury settlement timing

Keep CP materialization/ship-overage charge in the existing pre-Research/pre-Construction `settleTreasury` phase.

Each settlement:

1. derive CP Capacity/Used from current extant state;
2. write `Empire.CommandPoints`;
3. compute `overage = max(0, Used-Capacity)`;
4. compute `ShipCommandMaintenanceBC = overage * 10` for the supported standard/player rule;
5. include it in Treasury total/net fields;
6. emit the existing Treasury settlement event with the expanded snapshot.

A ship completed later in that same resolution therefore starts consuming/charging on the next settlement, matching the original order.

### 6. No new player command is required

Command Points are derived Empire state. Slice 05 needs no new strategic order merely to calculate CP.

### 7. Observer/save-load/replay

- schema-20 save/load preserves exact last-materialized CP/Treasury snapshot;
- Observer receives CP through the cloned Empire state;
- Treasury events expose the same expanded snapshot;
- identical sessions must reproduce Capacity, Used, ship-overage BC and final Treasury exactly.

### 8. Vertical Gate-3 tests

At minimum prove:

- empty Empire starts with Capacity 5 / Used 0;
- military hulls consume 1..6 CP from HullID;
- Colony/Outpost special ships each consume 1 CP, stationary and in transit;
- Warlord +2 per Colony;
- each station tier base CP;
- Tachyon/Subspace/Hyperspace Communications +1/+2/+3 per station and strongest-known selection;
- Dictatorship + known Imperium gets `capacity += floor(capacity/2)` after other supported producers;
- overage 0 at/below capacity and exactly 10 BC/point above it;
- station replacement avoids double station capacity/maintenance;
- a military ship completed this turn affects CP Maintenance only on the following settlement;
- split/merge/movement do not alter Used CP;
- consuming a Colony/Outpost special ship before settlement removes its 1-CP usage;
- save/load, Observer isolation and deterministic replay.

### 9. Explicitly deferred

- Leader/Officer Command Point bonuses and Officer Maintenance;
- original troop Transport CP until an authoritative troop-Transport subsystem exists;
- AI/difficulty `12 - global setting` overage rate;
- automatic deficit ship scrapping/liquidation;
- station tactical combat, weapons, scanner and battle behavior;
- hostile battle destruction timing beyond the current Slice-04 strategic surface.

## Gate 2 final decisions

Gate 2 accepts the Gate-1 proposal with the following exact implementation decisions and clarifications.

### 1. Core schema 20 and Command-Point snapshot semantics

Gate 3 bumps `core.StateSchemaVersion` from 19 to **20** and adds:

```go
type EmpireCommandPoints struct {
    Capacity int `json:"capacity"`
    Used     int `json:"used"`
}
```

as `Empire.CommandPoints`.

`Capacity` and `Used` are a **materialized last-settlement snapshot**, not a second independently editable gameplay truth. The authoritative inputs remain Ships/Fleets, Colonies/buildings, Technology ownership and race/government rules. Core validation requires both snapshot values to be non-negative, but must **not** require them to equal a freshly recomputed live value at every arbitrary state boundary: current-turn Production occurs after Treasury settlement, so a newly completed Ship deliberately does not alter this snapshot or its Maintenance charge until the next settlement.

Do not persist redundant `Reserve` or `Overage` fields. Derive them where needed:

```text
reserve = Capacity - Used
overage = max(0, Used - Capacity)
```

This stale-until-next-settlement behavior is intentional for Slice 05 and is covered by a vertical test.

### 2. Treasury schema extension

Add exactly one modeled Maintenance category:

```go
ShipCommandMaintenanceBC float64 `json:"ship_command_maintenance_bc"`
```

`EmpireTreasuryState.TotalModeledMaintenanceBC` becomes:

```text
BuildingMaintenanceBC
+ FreighterOperatingCostBC
+ ShipCommandMaintenanceBC
```

and `NetModeledIncomeBC` / `BalanceBC` use that expanded total.

No fake zero-valued Spy, Leader/Officer, Tribute or other deferred categories are added.

`empire.treasury_settled` remains the settlement event and carries the expanded `EmpireTreasuryState`; no separate Command-Point settlement event is required in this slice.

### 3. Economy ruleset schema 8

Gate 3 bumps `ruleset.EconomySchemaVersion` and `data/rulesets/moo2-1.31/economy.json` from 7 to **8** and adds a required `command_points` rule block with original-proven semantic inputs:

```text
base_capacity = 5
fixed_special_ship_points = 1
warlord_points_per_colony = 2
standard_overage_bc_per_point = 10

station points:
  star_base      = 1
  battlestation  = 2
  star_fortress  = 3

communications, strongest known wins:
  technology 180 Tachyon Communications    = +1 / station
  technology 176 Sub Space Communications  = +2 / station
  technology 91  Hyperspace Communications = +3 / station

Imperium:
  technology 92
  required base government = government_dictatorship
  bonus numerator/denominator = 1/2
  rounding = floor for the non-negative generated-capacity domain
```

The exact Go rule-struct names may follow existing `internal/ruleset/economy.go` conventions, but the semantic fields above are mandatory and validated. Provenance/source IDs for the new block must point to the Slice-05 original HELP/executable evidence rather than introducing unsupported secondary values.

Do **not** bump `ship_hulls.json`: military Ship CP cost is derived from the existing authoritative `ShipHull.SizeIndex + 1`.

The original NPC rate `12 - global[0x21CB0]` is documented evidence only and is not an active schema-8 gameplay option.

### 4. Authoritative usage derivation

Add one pure authoritative helper that derives `(Capacity, Used)` for one Empire before settlement.

Military usage:

```text
for every concrete Ship owned by Empire:
    resolve Ship.Spec.HullID through normalized hull rules
    add hull.SizeIndex + 1
```

The helper counts each concrete Ship exactly once regardless of stationary/transit Fleet state. It rejects unknown Hull IDs or structurally invalid ownership instead of approximating.

Fixed-special usage:

```text
each extant Colony special StrategicFleet = +1
each extant Outpost special StrategicFleet = +1
```

Current MOOX fixed special Fleets do not also contain concrete `ShipIDs`, so this does not double-count concrete military Ships. Population transfers, Food freighters and generic Settler reservations are **not** original troop Transports and consume no CP.

Split, merge and strategic movement therefore cannot change `Used`; creating/consuming an actual supported Ship/special Fleet can.

### 5. Authoritative capacity derivation and ordering

For every owned active Colony, inspect the legal orbital command-station family and race/technology modifiers. The supported formula is exactly:

```text
stationCount = starBaseCount + battlestationCount + starFortressCount
communicationsPerStation = strongest known of {0, 1, 2, 3}

capacityBeforeImperium =
    5
  + starBaseCount
  + 2 * battlestationCount
  + 3 * starFortressCount
  + communicationsPerStation * stationCount
  + (Warlord ? 2 * ownedActiveColonyCount : 0)
  + leaderCommandBonus

leaderCommandBonus = 0 // until authoritative Leader subsystem exists

capacity = capacityBeforeImperium
if base government is government_dictatorship and Technology 92 is known:
    capacity += floor(capacityBeforeImperium / 2)
```

The Imperium bonus is last among all currently supported producers. A future Leader implementation must insert its original-proven bonus **before** this +50% step.

### 6. Orbital station family replacement is accepted as a narrow Slice-05 dependency

Gate 3 normalizes only the station-family ownership/buildability semantics required for correct CP and Building Maintenance:

```text
Star Base < Battlestation < Star Fortress
```

Rules:

- completing Battlestation removes Star Base from that Colony;
- completing Star Fortress removes Star Base and Battlestation;
- a Colony owning Battlestation cannot offer/build Star Base;
- a Colony owning Star Fortress cannot offer/build Star Base or Battlestation;
- the exact already-owned tier remains unavailable as today;
- CP derivation rejects an impossible state containing more than one station-family tier instead of silently summing it.

Because Production completion is after Treasury settlement, the turn in which an upgrade completes still pays the lower station's current-turn Building Maintenance and uses its current-turn CP capacity; the replacement tier affects the following settlement. This matches the accepted pre-Construction settlement boundary.

No station weapons, tactical stats, scanners, BattleSession behavior or combat resolution are included.

### 7. Treasury transaction rule

Current `settleTreasury` mutates each Empire while iterating, and can currently return an error later (for example on an unknown building). Slice 05 adds new error-producing hull/station validation, so Gate 3 must avoid introducing a new partial-mutation path.

Before the first Treasury balance/snapshot mutation, precompute/validate all per-Empire settlement facts needed by this slice, including Command Points and Building Maintenance. Only after every Empire's settlement input is valid may the resolver apply Treasury/CommandPoint snapshots and emit events in deterministic Empire-ID order.

A rejected settlement must not leave an earlier Empire with an updated Balance/CommandPoints while a later Empire remains unchanged.

This transaction hardening is part of Slice 05 because the new CP helper adds authoritative validation failures directly inside settlement.

### 8. Resolution timing

Keep the existing MOOX strategic order. Command Point materialization and ship-overage Maintenance occur in the existing Treasury settlement phase **before** Research, Population changes, Fleet transit advancement and Construction completion.

For supported Slice-05 state this preserves the proven original boundary that current-turn Production does not retroactively change the Maintenance charge already settled.

Consequences to test:

- a military Ship completed later in the resolution affects `Used` and overage Maintenance on the **next** settlement;
- station replacement completed later in the resolution affects capacity/Building Maintenance on the **next** settlement;
- a Colony/Outpost special Ship consumed by a legal command **before** settlement is absent from that settlement's usage;
- split/merge/movement before settlement do not alter usage.

No separate post-Construction CP refresh is added in Slice 05.

### 9. Legal-action and command surface

No new player command is introduced for Command Points or ship Maintenance.

CP is a derived/materialized Empire economy state. Existing strategic commands remain legal according to their own rules even if they push the Empire into negative CP reserve; the consequence is the 10 BC/point Maintenance charge, not command rejection.

The only legal-action change is the accepted orbital-station buildability suppression from the station replacement family. Existing `AvailableConstructionChoices` / Session legal-action projection must reflect that automatically.

Do not add a fleet-size hard cap based on CP.

### 10. Observer, save/load and deterministic replay

Observer/PlayerView must expose the materialized `Empire.CommandPoints` through its normal isolated Empire clone. Because the snapshot contains only scalar integers, no new slice deep-copy hazard is introduced, but focused Observer coverage is still required.

Schema-20 roundtrip must preserve exactly:

- `CommandPoints.Capacity`;
- `CommandPoints.Used`;
- `Treasury.ShipCommandMaintenanceBC`;
- expanded Treasury totals/balance;
- station-family ownership after a completed replacement.

Identical-session replay tests compare final State and DomainEvents, including the expanded Treasury settlement payload.

No separate event-sourced reducer is introduced.

### 11. Gate-3 implementation order

Gate 3 should proceed in this order:

1. schema-20 Core fields/validation and schema-8 ruleset structures/data/provenance;
2. Warlord projection and Command-Point rules loading/validation;
3. pure military/special usage + capacity derivation helper;
4. station-family validation/buildability/replacement behavior;
5. transactional Treasury precompute/apply split + CP snapshot/ship-overage charge;
6. Observer/save-load/replay surfaces;
7. focused vertical tests, then full repository QA.

Every new error-producing helper used by settlement must be exercised by a no-partial-mutation regression.

### 12. Gate-3 required test matrix

At minimum:

- empty Empire -> Capacity 5 / Used 0;
- all six military hulls -> 1..6 CP from `SizeIndex + 1`;
- Colony and Outpost special Fleets -> 1 CP each, stationary and transit;
- Population-transfer/Freighter state -> no CP;
- Star Base/Battlestation/Star Fortress -> +1/+2/+3;
- Tachyon/Subspace/Hyperspace -> +1/+2/+3 per station, strongest known only;
- Warlord -> +2 per owned active Colony;
- Dictatorship + Technology 92 -> `capacity += floor(capacity/2)` after other supported sources, including odd-capacity truncation;
- non-Dictatorship knowing Technology 92 does not receive Imperium CP bonus;
- no overage charge at/below capacity; exactly 10 BC per excess CP above capacity;
- movement/split/merge preserve Used;
- unknown Hull or invalid multiple station tiers rejects settlement without any Empire balance/CP partial mutation;
- station upgrade removes lower tiers and prevents lower build choices;
- completed Ship and completed station upgrade affect CP/Treasury only on following settlement;
- existing Building/Freighter Maintenance regressions remain exact;
- schema-20 save/load, Observer isolation and identical-session replay/event equality.

### 13. Explicit deferrals remain binding

Gate 3 must not pull in:

- Leader/Officer Command-Point bonuses or Officer Maintenance;
- original troop Transport CP until a real troop-Transport subsystem exists;
- NPC/difficulty `12 - global[0x21CB0]` rate;
- automatic deficit ship liquidation/scrapping;
- tactical station combat/scanner behavior;
- hostile battle destruction timing beyond the current strategic surface.

These are not omissions from the accepted contract; they are explicit later dependencies.
## Gate 1 conclusion

Gates 1-2 are complete. The original 1.31 evidence is sufficient and the implementation contract above is accepted. Gate 3 is the next step.

**Gates 1-4 are complete. Slice 05 is closed.**

## Gate 3 implementation

Gate 3 implements the accepted Gate-2 contract without broadening the explicit deferrals.

### Core schema 20

`core.StateSchemaVersion` is now **20**. `Empire` persists the last-settlement snapshot:

```go
type EmpireCommandPoints struct {
    Capacity int `json:"capacity"`
    Used     int `json:"used"`
}
```

`EmpireTreasuryState` now includes:

```go
ShipCommandMaintenanceBC float64 `json:"ship_command_maintenance_bc"`
```

Core validation requires non-negative Command-Point Capacity/Used and a finite ship-command Maintenance value. It intentionally does not require the snapshot to equal a fresh live derivation outside settlement, preserving the accepted pre-Construction timing boundary.

### Economy ruleset schema 8

`economy.json` and `ruleset.EconomySchemaVersion` are now **8**. The committed `command_points` block records the original-derived values accepted in Gate 2:

- base capacity 5;
- fixed supported special Ship 1 CP;
- Warlord +2 per owned Colony;
- standard/player overage 10 BC per point;
- Star Base/Battlestation/Star Fortress +1/+2/+3;
- Tachyon/Sub Space/Hyperspace Communications Technology 180/176/91 -> +1/+2/+3 per station, strongest known value only;
- Imperium Technology 92 for base `government_dictatorship`, +1/2 with floor semantics.

The new rule source `moo2-1.31-command-points-ship-maintenance` points to the direct HELP/executable Slice-05 evidence. `ship_hulls.json` is unchanged: concrete military Ship usage is still derived from the authoritative existing hull `SizeIndex + 1`.

`RaceEconomyModifiers` now projects the already-normalized `warlord` ability for the CP calculator.

### Authoritative Command-Point derivation

`internal/game/command_points.go` provides one pure derivation path for the supported surface.

Usage:

- every concrete owned military `Ship` resolves its `HullID` through normalized hull rules and consumes `SizeIndex + 1`;
- every extant Colony Ship or Outpost Ship special `StrategicFleet` consumes 1 CP;
- stationary/transit Fleet location does not alter usage;
- split, merge and strategic movement alter Fleet containers, not concrete Ship count, and therefore preserve usage;
- legal Outpost deployment consumes/removes its special Fleet and therefore removes its 1 CP from subsequent derivation/settlement;
- Population transfers and freighters do not consume CP;
- unknown hulls or invalid structural station state reject derivation rather than approximating.

Capacity follows the accepted ordering:

```text
5
+ station base points
+ strongest Communications bonus * station count
+ Warlord 2 * owned Colony count
+ future Leader term (currently 0)
then Imperium: capacity += floor(capacity / 2)
```

The Imperium bonus is applied only when the base government trait is Dictatorship and Technology 92 is known.

### Orbital command-station family

The original replacement relationship is normalized as the narrow Slice-05 dependency:

```text
Star Base < Battlestation < Star Fortress
```

- Battlestation completion removes Star Base;
- Star Fortress completion removes either lower tier;
- buildability suppresses an equal/lower station tier once a higher tier exists;
- multiple station-family tiers in an invalid state reject CP derivation/settlement;
- no tactical station weapons/scanners/combat behavior was added.

Because Treasury settles before Construction, the lower station remains the current-turn CP/Building-Maintenance input; a newly completed upgrade affects the following settlement.

### Transactional Treasury settlement

`settleTreasury` is now a two-phase operation:

1. derive and validate every Empire's Command Points, Building Maintenance and complete Treasury/event plan without mutating State;
2. only after all plans succeed, apply `Empire.CommandPoints`, Treasury snapshots and events in deterministic Empire-ID order.

The ship-command bucket is:

```text
overage = max(0, Used - Capacity)
ShipCommandMaintenanceBC = overage * 10
```

and `TotalModeledMaintenanceBC` is now Building + Freighter + Ship-command Maintenance.

This also hardens the existing settlement path: an unknown Hull, unknown building or impossible station-family state for a later Empire cannot leave an earlier Empire partially settled.

### Timing

The existing strategic resolver order remains unchanged. Treasury/CP settlement occurs before Research, Population changes, Fleet transit advancement and Construction completion.

Vertical tests prove:

- a sixth Frigate completed after a 5/5 settlement does not retroactively change that snapshot/charge; next settlement becomes 6/5 and charges 10 BC;
- a station upgrade completed after settlement likewise changes capacity and Building Maintenance only on the next settlement.

No post-Construction CP refresh or CP fleet-size hard cap was introduced.

### Save/load, Observer and replay

Schema-20 round-trip tests preserve CommandPoints and the expanded Treasury snapshot exactly. Session coverage proves Observer mutation cannot leak into authoritative CP/Treasury State, and two identical sessions reproduce identical final State and DomainEvents including the expanded `empire.treasury_settled` payload.

### Gate-3 test coverage

New focused coverage includes:

- exact committed CP rule values;
- all six hull CP costs 1..6;
- Colony/Outpost special usage including transit;
- no Population-transfer/Freighter CP;
- base/station/Communications/Warlord/Imperium capacity including odd-value truncation and non-Dictatorship negative case;
- at/below/above capacity ship Maintenance;
- split/merge/movement usage invariance;
- legal Outpost Ship consumption removes 1 CP;
- station replacement/buildability and invalid multiple-tier rejection;
- no-partial-mutation settlement errors across multiple Empires;
- current-turn Ship/station completion timing;
- schema-20 Core validation/save-load;
- Observer isolation and deterministic replay.

Gate-3 QA passed:

```text
go test ./internal/core ./internal/ruleset ./internal/game ./internal/session -count=1
go test ./... -count=1
```

A final fresh formatting/focused/full-suite/diff pass is run again before Gate-3 handoff. Gate 4 still owns `go vet ./...`, final diff review, commits and slice closure.

### Deferrals preserved

No Leader/Officer CP or Officer Maintenance, original troop Transport CP, NPC/difficulty `12 - global[0x21CB0]` rate, automatic deficit liquidation/scrapping, tactical station behavior or hostile-battle destruction timing was implemented.

## Gate 4 closure

Fresh Gate-4 verification passed on 2026-08-31:

```text
gofmt changed/untracked Go files
go test ./internal/core ./internal/ruleset ./internal/game ./internal/session -run '(CommandPoint|CommandStation|Treasury|Military|CombatFleet|ColonyShip|Outpost|Construction|Observer|Save|Load|Replay)' -count=1
go test ./... -count=1
go vet ./...
git diff --check
```

The focused original-accounting fixtures also passed explicitly for committed CP values, station/Communications/Warlord/Imperium capacity including odd-value floor behavior, 10 BC/excess-CP Treasury accounting, pre-Construction Ship/station timing, Fleet split/merge/movement invariance and legal Outpost Ship consumption. The schema-20 compatibility diff was reviewed to ensure older Slice-03/04 persistence tests only changed their expected schema number/name and retained their gameplay assertions.

Gameplay/data/evidence commit: `16afe0b` (`game: add command point maintenance`).

Final composition/integrity review confirms:

- every concrete owned Ship contributes exactly one hull-derived CP term independent of Fleet container operations;
- fixed Colony/Outpost special Fleets contribute exactly 1 CP while extant and drop out when legally consumed;
- Population transfers/Freighters remain excluded;
- invalid hull or impossible multiple-station-tier state aborts all-Empire Treasury settlement before the first mutation;
- Star Base/Battlestation/Star Fortress replacement leaves at most one command-station tier after legal completion;
- current-turn Construction still cannot retroactively alter the already-settled CP/Treasury snapshot.

No push was performed. Leader/Officer CP, original troop Transport, NPC/difficulty overage rate, deficit liquidation, tactical station behavior and hostile-battle destruction timing remain deferred.