# Tactical ship combat baseline - Slice 07 evidence

Date: 2026-09-01

Status: **Gates 1-3 complete; Gate 4 pending**

Planned slice: `docs/slices/PLANNED_07_TACTICAL_SHIP_COMBAT_BASELINE.md`

Recovery marker: `docs/slices/_OPEN_TACTICAL_SHIP_COMBAT_BASELINE_2026-09-01.md`

## Objective

Establish one narrow, deterministic and directly evidenced Master of Orion II 1.31 tactical ship-combat fixture that can produce an authoritative `BattleSession` result and reconcile survivors/losses through the completed Slice-06 strategic handoff without pretending the broader tactical combat system is already implemented.

## Repository baseline at Gate-1 start

- starting HEAD: `2d7ac6e` (`docs: close strategic encounter battle handoff slice`);
- branch `main` was 32 commits ahead of `origin/main`;
- working tree clean;
- zero pre-existing `_OPEN_*.md` markers;
- Core schema 20 / economy ruleset schema 8 / ship-hulls ruleset schema 3;
- Slice 06 closed with gameplay/evidence commit `096fd0a`;
- Slice 07 was the next prepared objective.

## Current MOOX tactical infrastructure baseline

`internal/battle` currently implements BattleSession lifecycle, strategic side identity and result handoff only. It does **not** yet implement tactical movement, initiative, firing, damage or hit points.

The current strategic `core.ShipDesignSpec` / built `core.Ship.Spec` snapshot includes hull, strategic picture, warp drive, FTL speed, computer, armor, shield, fuel and cost/space values, but **no weapon mounts** and no tactical state.

Slice 03 intentionally established an original-valid cleared/unarmed Frigate baseline. Its evidence explicitly notes that the original design contains weapon/special arrays while MOOX deferred those arrays until tactical combat was researched. The current `SaveMilitaryDesignPayload` likewise exposes only design ID/name/hull/picture and still creates the cleared/unarmed baseline.

Slice 06 now provides the strategic boundary needed by this slice:

- BattleSpec freezes directed attacker/defender Empire/Seat/Fleet/Ship/civilian/Colony identities;
- strategic State is frozen during active encounter waves;
- the supported result has exactly one winner plus optional concrete `DestroyedShipIDs`;
- destroyed strategic Ships are removed and surviving losers retreat before Blockades/transfers.

Therefore Slice 07 needs to produce a trustworthy tactical result; it does not need a second strategic reconciliation mechanism.

## Original-reference provenance

Private local reference executable:

`C:\ASH\Temp\mastori2\Orion2.exe`

Original HELP text:

`reference/original/text/help/block_0000.ascii.txt`

The executable is evidence-only and must not be distributed with MOOX.

Future executable work uses the corrected Watcom record interpretation established in Slice 06: the current function object offset is stored in the four bytes before the current symbol-record header; the value after a symbol name belongs to the next record.

## Canonical first tactical fixture

Gate 1 converged on the following intentionally narrow fixture.

### Attacker

- Frigate;
- Fusion Drive;
- Electronic Computer;
- Titanium Armor;
- no shield;
- standard fuel;
- exactly **one standard Laser Cannon**;
- no beam mount/modifier flags;
- no specials;
- crew level 0;
- no leader/race/fleet combat bonus;
- explicit post-deployment tactical position `(10,10)`.

### Defender

- Frigate;
- Nuclear Drive;
- Electronic Computer;
- Titanium Armor;
- no shield;
- standard fuel;
- **unarmed**;
- no specials;
- crew level 0;
- no leader/race/fleet combat bonus;
- explicit post-deployment tactical position `(11,10)`.

### Battle options / fixture constraints

- tactical initiative **enabled**;
- no planet/station defense;
- no tactical movement in the first fixture;
- no retreat/capture/stasis;
- no shields, so facing does not affect the supported damage path;
- attacker fires its single standard Laser on its activation;
- defender ends its activation without firing;
- next round attacker fires the Laser a second time;
- original RNG state is pinned to the canonical trace documented below.

This is a **baseline fixture**, not a claim that all armed ships or all tactical battles are now supported.

## Direct HELP evidence

### Laser Cannon

Original HELP describes the Laser Cannon as a beam weapon that inflicts **1-4 damage**. HELP lists optional modifications such as Autofire, Armor Piercing, Heavy Mount, Continuous, Point Defense and No Range Dissipation. The baseline uses none of them.

### Electronic Computer

Original HELP gives the Electronic Computer a **+25 chance-to-hit bonus with beam weapons**. The baseline uses a pristine computer with no damage.

### Titanium Armor

Original HELP identifies Titanium Armor as the standard ship armor. No Reinforced Hull or Heavy Armor special is used.

### Combat speed and initiative

Original HELP states that combat speed determines how far a ship can travel in one combat round.

With tactical initiative enabled, HELP states that ships act from highest to lowest initiative and gives the initiative expression as:

```text
modified beam offense / 10 + current combat speed
```

The executable independently proves the same arithmetic.

### Shields

HELP documents four 90-degree shield arcs and shield absorption before internal damage. The baseline deliberately uses **no shield** so shield-facing and recharge semantics remain outside the first fixture.

## Correct tactical executable symbols used by Gate 1

| Function | object offset | VA |
| --- | ---: | ---: |
| `Range_To_Ship_` | `0x1885A` | `0x2885A` |
| `Apply_Internal_Damage_` | `0x25251` | `0x35251` |
| `Select_Internal_` | `0x26810` | `0x36810` |
| `Total_Internal_Space_` | `0x26B64` | `0x36B64` |
| `Is_Combat_Ship_Dead_` | `0x2897A` | `0x3897A` |
| `Fire_Ship_` | `0x28B5E` | `0x38B5E` |
| `Get_Beam_Range_To_Hit_Bonus_` | `0x29406` | `0x39406` |
| `Get_Beam_Weapon_Modifiers_` | `0x29434` | `0x39434` |
| `Fire_Beam_Weapon_` | `0x294F7` | `0x394F7` |
| `Apply_Ship_Damage_` | `0x29985` | `0x39985` |
| `Facing_Shield_` | `0x29DE0` | `0x39DE0` |
| `Destroy_Ship_` | `0x29E15` | `0x39E15` |
| `Weapon_In_Range_` | `0x29F1D` | `0x39F1D` |
| `Init_Ship_For_Start_Of_Turn_` | `0x32B70` | `0x42B70` |
| `Set_Ocv_Dcv_For_All_` | `0x32DE8` | `0x42DE8` |
| `Ship_Initiative_Value_` | `0x32E66` | `0x42E66` |
| `qsort_ship_initiative_` | `0x32E9C` | `0x42E9C` |
| `Do_Combat_Turn_` | `0x32F7F` | `0x42F7F` |
| `Calc_Current_Speed_` | `0x3528F` | `0x4528F` |
| `Check_For_Winner_` | `0x3545B` | `0x4545B` |
| `Load_Combat_Ship_` | `0x3954A` | `0x4954A` |
| `Calc_Combat_Shield_Max_` | `0x394A8` | `0x494A8` |
| `Get_Ship_Combat_Bonuses_` | `0x44E5B` | `0x54E5B` |
| `Get_Ship_Structure_` | `0x48387` | `0x58387` |
| `Get_Ship_Armor_Hits_` | `0x48425` | `0x58425` |
| `Current_Design_Min_Combat_Speed_` | `0x5B82A` | `0x6B82A` |
| `Current_Design_Max_Combat_Speed_` | `0x5B84A` | `0x6B84A` |
| `Current_Design_Base_Combat_Speed_` | `0x5B86A` | `0x6B86A` |
| `Weapon_Cost_` | `0x5EC74` | `0x6EC74` |
| `One_Weapon_Space_` | `0x5EDC6` | `0x6EDC6` |
| `Weapon_Space_` | `0x5EE8E` | `0x6EE8E` |
| `Random_` | `0x1147A0` | `0x1247A0` |
| `Set_Random_Seed_` | `0x114820` | `0x124820` |
| `Get_Random_Seed_` | `0x11484C` | `0x12484C` |
| `Range_` | `0x124637` | `0x134637` |

## Gate-1 findings

### 1. Original tactical combat has a concrete per-ship runtime record

`Load_Combat_Ship_` materializes a tactical ship record of size `0x139` from the strategic ship/design data.

Fields directly relevant to the baseline include:

```text
+0x20 owner/player
+0x21 tactical X
+0x22 tactical Y
+0x24 tactical destroyed/absent state
+0x25 hull-size band used by center-offset logic
+0x26 shield type
+0x31 warp drive
+0x32 computer
+0x33 armor
+0x34 beam offense
+0x36 beam defense
+0x38 missile defense
+0x3A base combat speed
+0x3B current combat speed
+0x52... weapon slots, 0x0B stride, 8 slots
+0xAA maximum structure
+0xAC crew level
+0xC0 accumulated structure damage
+0xC2 current armor hits
+0x100 engine/internal health
+0x102 computer/internal health
+0x134 shield-generator state
```

MOOX currently has none of these mutable tactical fields. Gate 2 therefore needs a BattleSession-local tactical snapshot rather than mutating strategic Core Ships during combat.

### 2. Initiative is directly proven and the canonical fixture has no tie

`Ship_Initiative_Value_` returns exactly:

```text
current combat speed (+0x3B) + beam offense (+0x34) / 10
```

with integer division.

`qsort_ship_initiative_` orders descending. `Do_Combat_Turn_` builds the active ship list and uses this sort when the tactical initiative option is enabled.

`Set_Ocv_Dcv_For_All_` stores:

- `Offensive_Combat_Bonus_` at `+0x34`;
- beam `Defensive_Combat_Bonus_` at `+0x36`;
- missile defense at `+0x38`.

Crew-level table evidence gives crew level 0 no offense/defense bonus. Electronic Computer component index 1 gives **25 beam offense**.

Drive tables are directly decoded per hull:

```text
Nuclear Drive minimum combat speed: 10,8,6,5,4,3
Nuclear Drive pristine/max speed:    20,18,16,15,14,13

Fusion Drive minimum combat speed:   12,10,8,7,6,5
Fusion Drive pristine/max speed:     22,20,18,17,16,15
```

Values are Frigate through Doom Star.

`Current_Design_Base_Combat_Speed_` linearly reduces the max toward the min as engine damage accumulates; a pristine ship uses the max.

Canonical initiatives are therefore:

```text
attacker Fusion Frigate: 22 + 25/10 = 24
defender Nuclear Frigate: 20 + 25/10 = 22
```

The attacker acts first without requiring any race, leader, crew, special-system or MOOX tie-break rule.

### 3. The first fixture can use fixed post-deployment positions and defer movement/deployment

`Range_To_Ship_` reads tactical X/Y, applies `Ship_Center_Offsets_`, then calls `Range_` and returns `(rawRange + 2) / 3` using integer division.

`Range_` computes:

```text
max(abs(dx), abs(dy)) + floor(min(abs(dx), abs(dy)) / 2)
```

For equal-size Frigates the center offsets cancel. Explicit post-deployment positions `(10,10)` and `(11,10)` therefore produce raw range 1 and Beam range index 1.

The original `Deploy_Ships_` machinery is substantial and not required to prove firing/damage. Gate 1 recommends that the baseline BattleSpec carries explicit post-deployment positions and that tactical movement/deployment commands are unsupported in Slice 07.

### 4. Normal Beam range tables are closed

`Get_Beam_Range_To_Hit_Bonus_` uses the signed table:

```text
range index:       0   1   2   3   4   5   6   7   8
To-Hit modifier:   0   0 -10 -20 -30 -40 -55 -70 -85
```

`Get_Beam_Weapon_Modifiers_` starts at a 100% damage multiplier and applies the normal signed range dissipation table:

```text
range index:       0   1   2   3   4   5   6   7   8
damage modifier:   0   0 -10 -20 -30 -40 -50 -60 -65
```

At canonical range index 1 both modifiers are zero, so the fixture uses full To-Hit and damage without needing a range-dissipation special case.

### 5. The exact standard Laser table entry is identified

The executable weapon table has `0x1C`-byte records. Weapon entry index 3 has:

```text
weapon kind: 0 (beam)
technology ID: 100
base space: 10
base cost: 5
minimum damage: 1
maximum damage: 4
```

Normalized `technologies.json` independently maps technology ID 100 to `laser_cannon`.

Thus Gate 1 can identify an authoritative first weapon rather than hard-coding only HELP prose.

The base Space/Cost values are sufficient for the canonical no-modifier/no-miniaturization fixture. Full weapon miniaturization and modifier pricing remain outside Gate 1.

### 6. Beam hit roll and threshold are directly closed

The simple ship-target path in `Fire_Beam_Weapon_`:

1. computes tactical range;
2. applies `Get_Beam_Weapon_Modifiers_`;
3. forms attacker beam offense minus defender beam defense;
4. calls `Random_(100)`;
5. if raw roll is at most 95, adds the offense-defense differential and caps at 100; raw values above 95 are forced to 100;
6. computes the hit threshold as `min(95, 40 - rangeToHitModifier)`;
7. misses if the effective roll is below that threshold.

For the canonical fixture:

```text
range index = 1
range To-Hit modifier = 0
threshold = 40
attacker beam offense = 25
defender beam defense = 0
```

The baseline does not need a probability claim; it uses a fixed original RNG trace. A raw `Random(100)` result of 100 yields effective roll 100 and a hit.

### 7. Beam damage mapping is directly closed for the fixture

For the simple Beam path, the scaled minimum/maximum damage are derived from the weapon table and the range damage multiplier.

When max damage exceeds min damage, the hit quality maps into damage using the integer expression represented by:

```text
damage = min + ((effectiveRoll - threshold) * (max-min+1)) / (100-threshold)
```

and is capped to the weapon maximum.

At canonical Laser range 1:

```text
min = 1
max = 4
threshold = 40
effective roll = 100
```

The intermediate result reaches 5 and is capped to **4 damage**. Therefore the pinned raw roll 100 gives an original maximum Laser hit of 4.

### 8. Frigate Titanium Armor and Structure are both 4

The original Frigate hull table gives:

```text
base Armor hits = 4
base Structure = 4
```

`Get_Ship_Armor_Hits_` starts from the hull Armor value and applies the armor-material modifier. Titanium Armor is the baseline material and adds no percentage bonus, so the canonical Frigate has **4 current Armor**.

`Get_Ship_Structure_` gives the Frigate **4 maximum Structure** when no Reinforced Hull/special modifier is present.

### 9. The supported damage trace can avoid the broader internal-system engine

`Select_Internal_` always consumes `Random_(100)` first.

For ordinary non-Armor-Piercing damage while Armor remains, it then selects **Armor** (`internal type 7`) deterministically regardless of that consumed roll.

`Apply_Internal_Damage_` type 7 subtracts from current Armor `+0xC2`; if damage does not exceed remaining Armor, no further internal damage occurs.

Once Armor reaches zero, `Select_Internal_` computes a weighted internal-vs-Structure decision. With positive remaining Structure, the internal threshold is strictly below 100, so a `Random(100)` value of **100 guarantees Structure** (`internal type 0`).

`Apply_Internal_Damage_` type 0 adds damage to accumulated Structure damage `+0xC0`, capped at maximum Structure `+0xAA`.

`Is_Combat_Ship_Dead_` marks an ordinary ship dead when:

```text
structure damage >= maximum structure
```

It also contains engine-death/special cases that are not part of the canonical fixture.

Consequently the baseline can exercise a real destruction path without implementing computer, engine, weapon-slot or other subsystem damage:

```text
first max Laser hit:  Armor 4 -> 0
second max Laser hit: Structure Damage 0 -> 4
=> ship dead
```

Any noncanonical internal-system selection remains a hard unsupported boundary rather than being approximated.

### 10. Destruction and victory are directly closed

`Fire_Ship_` invokes `Is_Combat_Ship_Dead_` during firing and calls `Destroy_Ship_` for dead targets.

`Destroy_Ship_` puts a normal destroyed tactical ship in state `+0x24 = 5` and clears/retargets relevant continuing effects.

`Check_For_Winner_` ignores tactical ships in destroyed state 5 while counting remaining sides. With exactly one attacker Frigate, one defender Frigate and no Colony defense, destruction of the defender immediately leaves the attacker as the tactical winner.

The Slice-06 strategic result is then naturally:

```text
WinnerSeat = attacker seat
DestroyedShipIDs = [defender strategic ShipID]
Outcome = stable tactical-victory metadata
```

No new strategic post-battle mechanism is needed.

### 11. The exact original random-number generator is closed

`Random_` uses a mutable 32-bit global state. Each advance is:

```text
state = state * 0x41C64E6D + 0x3039   // modulo 2^32
```

For `Random(N)` it uses rejection sampling:

```text
q = floor(0xFFFFFFFF / N)
cutoff = q * N
advance until state < cutoff
return floor(state / q) + 1
```

Therefore `Random(N)` returns values **1..N inclusive** without simple modulo bias.

`Set_Random_Seed_` and `Get_Random_Seed_` directly expose the mutable state, and `Do_Combat_Turn_` contains combat synchronization around this seed. Tactical replay therefore needs one ordered mutable RNG stream, not independent per-shot random values.

The existing MOOX `battle.DeriveSeed(...)` is explicitly infrastructure seed logic, not an emulation of the original MOO2 battle-seed derivation.

Gate-1 recommendation:

- keep the existing 64-bit Battle ID/seed for BattleSession identity;
- add a BattleSession-local 32-bit tactical RNG state;
- initialize it deterministically from the BattleSpec using a documented MOOX mapping, for example low 32 bits of `battle.Spec.Seed` or an explicit fixture override;
- once initialized, use the exact original LCG and exact supported call order;
- expose/replay the current tactical RNG state;
- do **not** claim the initial mapping is the original MOO2 strategic-to-tactical seed derivation.

### 12. Canonical fixed original RNG trace

A temporary helper implementing the exact original `Random_` algorithm found the seed:

```text
initial RNG state = 0x001500BD
```

The first four `Random(100)` results are:

```text
100, 19, 100, 100
```

and the final state after those four outputs is:

```text
0xFD95EBB9
```

Applied to the canonical fixture:

1. first Laser `Fire_Beam_Weapon_` hit roll = **100** -> max 4 damage;
2. first `Select_Internal_` consumes **19**, but remaining Armor forces Armor -> Armor 4 to 0;
3. second Laser hit roll = **100** -> max 4 damage;
4. second `Select_Internal_` = **100** -> Structure guaranteed -> Structure Damage 0 to 4 -> dead.

The defender is unarmed, so it introduces no firing RNG calls between the two attacker shots.

This gives Gate 3 a byte-for-byte reproducible original-LCG fixture rather than mocked hit/damage outcomes.

### 13. Weapon readiness is round-local

`Init_Ship_For_Start_Of_Turn_` iterates tactical weapon slots and resets supported ready-state fields at the start of the combat turn. `Fire_Ship_` consumes the firing availability for the weapon slot.

The first baseline therefore needs only:

- one standard Laser weapon slot;
- one fire command during the attacker activation;
- an end-activation command;
- a new round that reinitializes the single Laser as ready;
- a second fire command.

Autofire, continuous fire, multiple batteries and other weapon scheduling remain deferred.

### 14. A persisted strategic weapon snapshot is required for an authoritative handoff

Slice 06 freezes strategic `ShipIDs`; it intentionally does not invent tactical equipment. The current built `core.Ship.Spec` is an immutable snapshot copied from `ShipDesignSpec`, which is the correct ownership model for adding the first weapon identity.

It would be incorrect for BattleSession to silently grant a Laser to an unarmed Slice-03 strategic Ship merely because the tactical fixture expects one.

Therefore Gate 2 must choose a minimal persisted weapon-mount representation in `ShipDesignSpec`/built `Ship.Spec` **or** explicitly limit the baseline to pre-seeded fixture Ships that carry an equivalent authoritative weapon snapshot. A tactical-only fake weapon not backed by the strategic Ship snapshot is rejected by Gate 1.

A likely minimal representation is conceptually:

```text
ShipWeaponMount {
    WeaponID = "laser_cannon"
    Count = 1
    Modifiers = none
}
```

Gate 1 does not yet approve a Core schema bump or command/API shape; that is Gate 2 work.

### 15. The first fixture does not need full deployment, movement or facing

The original has substantial deployment, movement, bearing and facing machinery. Implementing it before the first hit/damage/result chain would widen the slice dramatically.

The canonical positions are therefore treated as an explicit **post-deployment tactical snapshot**. Movement commands are outside the baseline. Because neither ship has shields and no supported weapon/special depends on firing arc/facing in this fixture, facing can remain absent from the first state model or fixed/noninteractive.

This is a deliberate supported-surface decision, not a claim that original MOO2 tactical combat lacks deployment/movement/facing.

### 16. Strategic reconciliation is already solved by Slice 06

The tactical combatant snapshot must retain a stable mapping back to the strategic `core.Ship.ID` from the Slice-06 BattleSpec.

The tactical layer mutates only BattleSession-local combat state while active. When the defender is destroyed it returns the strategic defender Ship ID in `DestroyedShipIDs`.

The already implemented Slice-06 continuation then:

- removes that concrete strategic Ship/Fleet identity;
- determines the winner from the singular WinnerSeat;
- handles any losing survivors' retreat;
- recomputes Blockades/transfers only after the battle result.

No tactical code should directly mutate the strategic Core State.

## Proposed Gate-2 implementation shape

This section is a proposal only; it is **not accepted or implemented yet**.

### A. Exact supported fixture

Gate 2 should approve the canonical Fusion-Frigate-with-one-standard-Laser versus unarmed Nuclear-Frigate fixture above, with initiative enabled, fixed adjacent post-deployment positions, no shields/specials/race/crew/leader bonuses and the pinned original-LCG trace for the golden regression.

### B. Tactical state remains BattleSession-local

Add an immutable tactical combatant snapshot plus mutable BattleSession-local state, conceptually containing:

```text
TacticalShipID / StrategicShipID
EmpireID / SeatID
X / Y
BaseCombatSpeed / CurrentCombatSpeed
BeamOffense / BeamDefense
ArmorMax / ArmorCurrent
StructureMax / StructureDamage
one supported Laser mount / ready state
Destroyed
```

No tactical HP/position/RNG field is persisted in Core during an active battle.

### C. Strategic weapon snapshot decision

Gate 2 should decide whether to:

1. bump Core schema and add a general `Weapons []ShipWeaponMount` snapshot to `ShipDesignSpec`/`Ship.Spec`, while Gate 3 supports only one standard Laser; or
2. keep Core schema 20 but restrict Slice 07 to a non-user-facing canonical pre-seeded Battle fixture whose authoritative Ship snapshot is extended outside the persistent design model.

Gate 1 recommends **option 1** because silently arming an otherwise unarmed strategic Ship is not an authoritative handoff. A likely Core schema is therefore 21, but Gate 2 owns the decision.

The existing player `SaveMilitaryDesign` command need not necessarily expose weapons in this slice. Gate 2 may keep armed designs fixture/pre-seeded only while the general design editor remains deferred.

### D. Tactical rules data

Gate 2 should decide whether to bump `ship_hulls.json` schema 3 or add a dedicated `tactical_combat.json` schema 1.

Gate 1 prefers a small dedicated tactical rules file referencing normalized IDs, because the new facts are a coherent tactical domain:

- standard Laser: tech100, beam, base space10, base cost5, damage1-4;
- normal Beam To-Hit range table;
- normal Beam damage-dissipation table;
- Frigate Armor4 / Structure4;
- Nuclear/Fusion Frigate combat-speed values needed by the fixture;
- Electronic Computer Beam Offense25;
- original RNG multiplier/increment/rejection behavior.

Existing strategic hull/component files remain sources for IDs and construction identity.

### E. Battle commands / authority

Minimal commands should be something like:

```text
battle.fire_beam(active tactical/strategic ship, target ship, weapon slot)
battle.end_activation(active ship)
```

Only the currently active ship's controlling Seat may submit commands. Target and weapon identity are validated against the frozen BattleSpec/tactical snapshot. Normal strategic commands remain blocked by `PhaseEncounters` as in Slice 06.

### F. Round / initiative

At round start:

- supported pristine fixture derives current speed and OCV;
- sort live ships by original initiative descending;
- canonical fixture has no tie (24 vs 22);
- if general code requires a tie-break, use TacticalShipID/StrategicShipID ascending as a documented deterministic MOOX tie-break, not an original claim;
- initialize supported weapon ready state;
- advance active ship only through explicit end-activation or destruction.

### G. Exact supported firing path

Gate 3 should implement only:

- standard Laser Cannon;
- normal Beam range 0..8 table logic;
- normal beam offense-defense hit calculation;
- exact original `Random_(100)` stream;
- standard damage scaling/cap;
- no shield;
- ordinary Armor then supported Structure damage;
- ship destruction and victory.

If post-Armor `Select_Internal_` chooses an internal subsystem outside the baseline, the command/battle must fail with an explicit unsupported-mechanic error **without partial mutation**. The golden pinned seed avoids that path and proves the complete supported chain.

This is intentionally a fixture engine, not yet a general tactical-combat simulator.

### H. RNG / replay

BattleSession owns one `uint32` tactical RNG state and advances it exactly in supported original call order. Tactical events/Observer state expose enough information to reproduce the state transition deterministically.

The initial mapping from current 64-bit BattleSpec seed to uint32 original state is a MOOX infrastructure decision unless Gate 2 pins an explicit fixture seed. The golden test must reproduce:

```text
initial 0x001500BD
rolls 100,19,100,100
final 0xFD95EBB9
```

### I. Result handoff

When the defender dies and `Check_For_Winner_` equivalent leaves only the attacker side:

```text
WinnerSeat = attacker SeatID
DestroyedShipIDs = [defender StrategicShipID]
Outcome = stable tactical-victory metadata
```

Feed that result through the existing Slice-06 atomic strategic continuation. Tactical code does not remove strategic Ships itself.

### J. Observer / replay / persistence

Observer must return detached tactical snapshots, round/active ship, weapon-ready/HP state and RNG state. Identical BattleSpec + commands + seed must reproduce identical tactical events/result and identical Slice-06 strategic reconciliation.

Durable process-restart serialization of an active BattleSession remains deferred, consistent with Slice 06.

## Explicit hard deferrals after Gate 1

- original `Deploy_Ships_` and player-selectable tactical deployment;
- tactical movement, turning and maneuver cost;
- facing-dependent weapon arcs;
- shields and shield recharge;
- missiles, torpedoes, bombs, fighters and point defense;
- beam mount/modifier families including Autofire, Armor Piercing, Heavy Mount, Continuous, Point Defense and No Range Dissipation;
- multiple weapon batteries / full eight-slot design surface;
- internal subsystem damage other than the canonical Armor-then-Structure trace;
- engine/computer/weapon damage and resulting speed/OCV degradation outside the fixed fixture;
- repair systems and regeneration;
- specials, cloaking, stasis, capture/boarding, retreat and self-destruct;
- colony/station/planet tactical combat, bombardment and invasion;
- leaders, race combat modifiers, nonzero crew levels and fleet combat bonuses;
- general military weapon-design UI, miniaturization and modifier pricing;
- exact original strategic-to-tactical initial RNG seed derivation;
- NPC/monster/Antaran tactical differences;
- active mid-battle process-restart persistence.

## Gate 1 conclusion

Gate 1 has a directly evidenced, deterministic and deliberately narrow tactical baseline:

```text
Fusion Frigate + Electronic Computer + Titanium Armor + 1 standard Laser
vs
Nuclear Frigate + Electronic Computer + Titanium Armor + no weapon

initiative 24 vs 22
fixed range index 1
original RNG state 0x001500BD
rolls 100,19,100,100
hit #1: 4 damage -> Armor 4 to 0
hit #2: 4 damage -> Structure Damage 0 to 4
defender destroyed
attacker wins
Slice-06 result destroys defender strategic ShipID
```

The evidence is strong enough for Gate 2 to decide the first tactical state/weapon/RNG/command contract without implementing broader MOO2 tactical combat.

**Stop here for Gate 2 acceptance/revision before tactical gameplay implementation.**
## Gate 2 accepted implementation contract

Status: **Gates 1-3 complete; Gate 4 pending**

Gate 2 accepts the Gate-1 canonical fixture and fixes the following implementation contract. These are architecture decisions for MOOX; original-MOO2 facts remain those evidenced above.

### Decision 1 - Core schema 21 carries authoritative weapon mounts

`core.StateSchemaVersion` advances **20 -> 21** because built strategic Ships must carry the weapon identity that tactical combat is allowed to use.

Add the structural snapshot:

```go
type ShipWeaponMount struct {
    Slot     int    `json:"slot"`
    WeaponID string `json:"weapon_id"`
    Count    int    `json:"count"`
}
```

and:

```go
Weapons []ShipWeaponMount `json:"weapons,omitempty"`
```

on `ShipDesignSpec`. Built `Ship.Spec` already copies the whole design spec, so the built ship remains independent from later design revision exactly as Slice 03 requires.

Core validation is structural, not ruleset-aware:

- at most 8 mounts;
- `Slot` in `0..7`;
- slots unique and stored in ascending slot order;
- non-empty `WeaponID`;
- positive `Count`.

Gate 3 does **not** add modifier/mount-mode fields. Autofire, Armor Piercing, Heavy Mount, Continuous, Point Defense, No Range Dissipation and future modifier data stay deferred instead of being represented as inert fields.

No migration from schema 20 is added in this slice; MOOX continues its current strict schema-version policy. Existing current-version persistence fixtures move to schema 21 while preserving their gameplay assertions.

### Decision 2 - the existing design command may save zero or one baseline Laser

Extend `SaveMilitaryDesignPayload` with an optional weapon list using the same structural concepts (`slot`, `weapon_id`, `count`). The generic protocol command envelope/schema does **not** change.

Slice-07 accepted player design surfaces are exactly:

```text
[]
```

or:

```text
[{slot:0, weapon_id:"laser_cannon", count:1}]
```

Anything else is rejected by `empire.save_military_design` in this slice.

For the one-Laser design:

- the Empire must know technology ID 100 (`laser_cannon`);
- `SpaceUsed += 10`;
- `BaseDesignCostPP += 5`;
- the existing military production-cost government reduction is re-applied to the new total base cost;
- hull-space overflow is rejected before state mutation.

The existing best-known mandatory component selection remains unchanged. Therefore an Empire can still save an armed design whose currently best shield/drive/etc. makes it outside the tactical baseline; the weapon snapshot is valid strategic state, while tactical support is decided separately and explicitly.

The existing cleared/unarmed Slice-03 design remains legal and unchanged when the weapon list is omitted.

### Decision 3 - add `tactical_combat.json` schema 1; do not change economy/ship-hulls schemas

Add:

```text
data/rulesets/moo2-1.31/tactical_combat.json
```

with `TacticalCombatSchemaVersion = 1` and a strict loader/validator in `internal/ruleset`.

It carries only the directly evidenced facts needed by this slice:

- initiative Beam-Offense divisor `10`;
- exact original RNG multiplier `0x41C64E6D`, increment `0x3039` and rejection-sampling semantics;
- Beam base hit threshold `40` and maximum threshold/effective-roll cap `95/100` semantics evidenced in Gate 1;
- Beam To-Hit range modifiers `0,0,-10,-20,-30,-40,-55,-70,-85`;
- Beam damage range modifiers `0,0,-10,-20,-30,-40,-50,-60,-65`;
- `laser_cannon`: technology ID100, beam kind, base space10, base cost5, damage1-4;
- Frigate Armor hits4 and Structure4;
- Nuclear-Drive Frigate min/max combat speed10/20;
- Fusion-Drive Frigate min/max combat speed12/22;
- Electronic Computer Beam Offense25.

The loader cross-validates referenced normalized IDs against existing technology/ship-component data where possible.

`economy.json` remains schema8. `ship_hulls.json` remains schema3. The current `EconomyRules` aggregate may carry the normalized tactical rule set even though its historical name is broader than pure economy; renaming that aggregate is explicitly outside this slice.

### Decision 4 - `battle.Spec` gets an optional self-contained tactical snapshot

`battle.Spec` remains the frozen authoritative battle boundary and gains an optional tactical specification plus an explicit unsupported reason for auto-created strategic encounters.

Conceptually:

```go
type TacticalSpec struct {
    Rules             TacticalRulesSnapshot `json:"rules"`
    InitiativeEnabled bool                  `json:"initiative_enabled"`
    InitialRNGState   uint32                `json:"initial_rng_state"`
    Ships             []TacticalShipSpec    `json:"ships"`
}
```

`TacticalRulesSnapshot` contains the exact schema/version and the small numeric rule tables/constants needed by the running BattleSession. Battle runtime must not read the filesystem or the mutable strategic GameState.

Each `TacticalShipSpec` is immutable and contains exactly the baseline facts needed at battle start:

```text
ShipID (the strategic core.Ship.ID; no second tactical ID in Slice 07)
EmpireID
SeatID
X / Y
HullID
WarpDriveID
ComputerID
ArmorID
weapon specs by slot
CurrentCombatSpeed
BeamOffense
BeamDefense
ArmorMax
StructureMax
```

Each tactical weapon spec carries slot, weapon ID, count and min/max damage copied from the validated tactical rules snapshot.

No facing field is introduced in Slice 07. The accepted fixture has no shield and no facing-dependent weapon rule, so introducing a meaningless facing value would falsely widen the supported surface.

### Decision 5 - tactical mutable state belongs only to `battle.Session`

For a tactical-enabled BattleSession, mutable state is BattleSession-local and conceptually contains:

```text
Round
InitiativeOrder []ShipID
ActiveShipID
NextCommandSequence
RNGState
per-ship ArmorCurrent
per-ship StructureDamage
per-weapon Ready state
Destroyed
local Battle Events
```

Fixed X/Y remains in the immutable TacticalSpec because movement is unsupported.

Round 1 starts when the BattleSession enters `PhaseActive`. The accepted fixture sorts to attacker ShipID first by the original initiative formula. The canonical fixture has no tie. If defensive code needs a stable tie fallback, use ascending strategic ShipID and document it as a deterministic **MOOX tie-break**, not an original-MOO2 claim.

At the start of every round all supported live Laser mounts become ready. `battle.end_activation` advances to the next live ship; ending the last live activation starts the next round, recomputes the stable initiative order and refreshes supported weapon readiness.

### Decision 6 - exact canonical auto-tactical support predicate

The strategic encounter preparer attaches a TacticalSpec only when the auto-created encounter is exactly inside the accepted Slice-07 surface:

- exactly one attacker combat Ship and one defender combat Ship;
- no defender Colony/planet defense and no civilian Fleet identities in the tactical fixture;
- attacker Ship: Frigate, Fusion Drive, Electronic Computer, Titanium Armor, no shield, standard fuel, exactly one mount `{slot:0, laser_cannon, count:1}`;
- defender Ship: Frigate, Nuclear Drive, Electronic Computer, Titanium Armor, no shield, standard fuel, zero weapon mounts;
- side/Ship Empire ownership and frozen strategic IDs all match the Slice-06 encounter identity.

The preparer assigns the fixed post-deployment positions:

```text
attacker (10,10)
defender (11,10)
```

which Gate 1 proved gives Beam range index1 for equal Frigates.

If an auto-created encounter is outside that exact surface, strategic encounter creation does **not** fail and no mechanic is approximated. The BattleSpec carries no TacticalSpec and carries an explicit `TacticalUnsupportedReason`. The existing Slice-06 lifecycle/manual-result path remains available for such battles. Calling the new tactical command API on one of them returns that explicit unsupported error.

Legacy/manual BattleSessions created by existing tests/APIs may also remain lifecycle-only.

### Decision 7 - the Slice-07 tactical baseline pins the Gate-1 golden RNG state

Gate 1 provisionally suggested mapping the existing 64-bit infrastructure battle seed to a 32-bit original RNG state. Gate 2 intentionally **does not accept that mapping yet**.

Reason: arbitrary seeds can make a standard Laser select deferred computer/engine/weapon internal-system damage after Armor is gone. That would make the supposedly accepted baseline nondeterministically leave its supported surface.

For the exact auto-tactical Slice-07 fixture only:

```text
InitialRNGState = 0x001500BD
```

is part of TacticalSpec.

The running RNG then uses the exact original LCG/rejection algorithm. The golden command trace must produce:

```text
Random(100): 100,19,100,100
final state: 0xFD95EBB9
```

This fixed initial state is a **MOOX baseline-fixture decision**, not a claim about original strategic-to-tactical seed derivation. `battle.Spec.Seed` remains the existing 64-bit battle identity/infrastructure seed and is not silently redefined.

A later tactical slice may replace the fixed initial tactical state with a documented derivation once arbitrary standard internal-damage paths are supported and evidenced.

### Decision 8 - use the existing `protocol.Command` envelope with a battle-local global sequence

Add two command kinds:

```text
battle.fire_beam
battle.end_activation
```

with payloads conceptually:

```go
type FireBeamPayload struct {
    ShipID       core.ID `json:"ship_id"`
    TargetShipID core.ID `json:"target_ship_id"`
    WeaponSlot   int     `json:"weapon_slot"`
}

type EndActivationPayload struct {
    ShipID core.ID `json:"ship_id"`
}
```

GameSession exposes one authoritative route conceptually equivalent to:

```text
SubmitBattleCommand(battleID, seatID, protocol.Command)
```

The caller's SeatID is authority metadata and is not trusted from command payload.

Each tactical BattleSession owns one **global battle command sequence**, starting at1 and exposed in its View. `protocol.Command.Validate(expectedSequence)` is reused. Successful commands consume exactly one command sequence. Rejected commands consume none.

Authority/legality:

- Seat must be a participant and control the current `ActiveShipID`;
- payload `ShipID` must equal that active ship;
- `fire_beam` target must be a live opposing tactical ship;
- weapon slot must exist, be ready and be the supported standard Laser mount;
- `end_activation` may only end the caller's current live ship activation.

A tactical-enabled BattleSession rejects public/manual `CompleteBattle` result injection. Its result must come from tactical commands. Lifecycle-only/unsupported battles preserve the existing manual `CompleteBattle` behavior.

### Decision 9 - command execution is transactional, including terminal strategic reconciliation

Every tactical command is first simulated/prepared against a copy of:

- tactical mutable state;
- RNG state;
- weapon readiness;
- local event sequence.

Any validation failure or unsupported mechanic returns an error with **zero mutation**: no RNG consumption, no ready-state change, no command-sequence advance and no events.

This includes a Beam hit that would select a deferred internal subsystem. The original selection may be computed on the copy, but if the selected result is outside the supported Armor/Structure baseline, the whole command is rejected atomically.

Non-terminal prepared commands can then commit directly to the child BattleSession.

A terminal command prepares, but does not prematurely commit, the candidate tactical result. GameSession reuses the Slice-06 staged continuation model:

- if other battles in the current wave are still active, the terminal tactical transition/result may commit to that child and strategic continuation waits;
- if this is the final active battle, build encounter outcomes with the candidate result, run `ResumeAfterEncounters` on a cloned strategic state, validate it and prepare any next-wave BattleSessions **before** committing the terminal tactical transition/result;
- only after all of that succeeds may the child tactical state/result, authoritative strategic state and next encounter wave mutate together.

This preserves the existing final-wave retry guarantee.

### Decision 10 - tactical snapshots must be prepared from the validated pre-encounter `resolution.State`

Current `GameSession.ResolveStrategic` prepares child BattleSessions before assigning `s.state = committed`. Once tactical snapshots depend on concrete Ship specs, reading the old `s.state` would be wrong.

Gate 3 must refactor encounter preparation to accept the explicit validated **pre-encounter committed state** (`resolution.State` clone) used for that encounter wave.

The same rule applies when `ResumeAfterEncounters` returns another encounter wave: prepare the next wave from that returned/validated state clone, never from the prior authoritative state.

The staged resolver/economy rules own the tactical rule data and provide the session layer with a fully frozen BattleSpec/TacticalSpec. `internal/battle` itself does not reach back into `internal/game`, ruleset files or mutable Core state.

### Decision 11 - exact supported fire/damage path

For `battle.fire_beam` the accepted engine implements only the direct Gate-1 path:

1. compute fixed-position range using original `Range_` / `Range_To_Ship_` arithmetic;
2. require range index within the evidenced normal Beam table;
3. apply normal To-Hit and damage range modifiers;
4. compute attacker Beam Offense minus defender Beam Defense;
5. consume exact original `Random(100)` hit roll;
6. apply original threshold/effective-roll logic;
7. on hit, derive Laser damage1-4 using the evidenced integer mapping and cap;
8. consume the original `Select_Internal_` RNG call;
9. while Armor exists, apply ordinary damage to Armor;
10. once Armor is zero, only a Structure selection is supported;
11. StructureDamage reaching StructureMax marks the ship destroyed.

The canonical pinned RNG trace is therefore guaranteed to remain inside the supported surface:

```text
round1 attacker fire -> max4 -> Armor4->0
attacker end activation
defender end activation
round2 attacker fire -> max4 -> StructureDamage0->4 -> destroyed -> victory
```

No second fire in the same round is legal because the slot is no longer ready.

### Decision 12 - tactical result is singular and uses the existing Slice-06 handoff

Destroying the defender in the accepted fixture produces the candidate result:

```text
WinnerSeat = attacker SeatID
Outcome = "tactical_victory"
DestroyedShipIDs = [defender strategic ShipID]
```

The existing result normalizer may mirror the singular winner into `WinnerSeats` for compatibility, but Slice-06 strategic continuation continues to use singular WinnerSeat semantics.

Tactical code never removes a strategic Ship/Fleet directly. The existing Slice-06 continuation owns concrete Ship removal, empty Fleet cleanup, loser retreat and later Blockade/transfer recomputation.

### Decision 13 - local Battle events are deterministic and do not leak cross-battle wall-clock ordering

A tactical BattleSession keeps its own local deterministic event stream in `battle.View`; it is **not** appended command-by-command into GameSession's global DomainEvent stream, because independent parallel battles could otherwise leak wall-clock submission order into authoritative global history.

Minimum local event kinds:

```text
round_started
beam_fired
battle_damage_applied
activation_ended
ship_destroyed
winner_determined
```

The local event has a Battle-local sequence plus SeatID/CommandSequence as applicable. `beam_fired` records range index, raw/effective hit roll, threshold, hit and damage. Damage events record layer and before/after Armor/Structure values plus the consumed selection roll where applicable.

`battle.View` exposes detached copies of:

- TacticalSpec;
- round / initiative order / active ship;
- NextCommandSequence;
- current RNG state;
- per-ship HP/destroyed/weapon-ready state;
- local tactical events;
- final result when completed.

GameSession keeps its existing stable `battle_created` / final `battle_completed` global events; the final completed Battle View contains the local tactical event history.

### Decision 14 - explicit hard boundaries

Tactical commands hard-reject rather than approximate:

- deployment changes;
- movement or turning;
- facing-dependent rules;
- any shielded tactical fixture;
- any weapon other than one standard Laser in attacker slot0;
- multiple batteries/count>1;
- all Beam modifiers/mount modes;
- missiles, torpedoes, bombs, fighters and point defense;
- internal-system damage to engine/computer/weapons or any other subsystem;
- repair/regeneration;
- specials;
- retreat, boarding/capture, self-destruct or stasis;
- Colony/planet/station tactical combat;
- nonzero crew/race/leader/fleet tactical bonuses;
- monsters/Antarans/NPC tactical differences;
- active mid-battle process-restart persistence.

Auto-created strategic encounters outside the exact baseline remain explicit lifecycle-only BattleSessions with a tactical-unsupported reason; this preserves Slice-06 functionality without claiming those battles are tactically simulated.

## Gate-3 acceptance tests fixed by Gate 2

Gate 3 must at minimum prove:

1. `tactical_combat.json` schema1 loads, validates and cross-references the committed normalized IDs.
2. Core schema21 round-trips empty and one-Laser weapon mounts; malformed slot/count/order is rejected.
3. `empire.save_military_design` preserves the unarmed baseline and can save exactly one legal Laser, enforcing tech100, +10 space and +5 base cost/production-cost recalculation.
4. A built Ship snapshots its weapon mount independently from later design revision.
5. Encounter preparation uses the passed committed pre-encounter state, not stale `s.state`.
6. Exact supported strategic encounter builds TacticalSpec with the frozen two strategic ShipIDs, positions, stats/rules and `InitialRNGState=0x001500BD`.
7. Nearby but unsupported encounters receive an explicit tactical-unsupported reason and remain legacy-manually completable.
8. Round1 initiative is attacker24 then defender22; round/ready/active state is deterministic.
9. Golden commands with sequences1..4 reproduce RNG `100,19,100,100`, final `0xFD95EBB9`, Armor4->0, Structure0->4, defender destruction and `tactical_victory`.
10. Wrong Seat, wrong command sequence, wrong active Ship, wrong target/slot, double fire and unsupported internal branch reject with byte-equivalent tactical state/RNG/events before vs. after.
11. Tactical-enabled battles reject external/manual result injection.
12. Observer/Battle View mutations do not alias authoritative TacticalSpec/state/events.
13. Same BattleSpec + same commands produces identical local events, View and Result.
14. GameSession integration applies defender strategic Ship destruction through Slice06 and preserves final-wave atomic retry if strategic continuation/preparation fails.
15. Multiple independent battles keep local event streams deterministic without wall-clock tactical command ordering in the GameSession global event log.
16. Full existing Slice03-06 military/Fleet/CommandPoint/encounter tests remain behaviorally intact except the expected Core schema21 fixture updates.

## Gate 2 conclusion

Gate 2 accepts a **real but deliberately tiny** tactical vertical slice: one strategically persisted Laser mount, one self-contained tactical BattleSpec, exact original Beam/RNG/Armor/Structure mechanics for the golden Frigate-vs-Frigate fixture, authoritative battle commands, and atomic result handoff through Slice 06.

It explicitly does **not** accept a fake tactical-only weapon, arbitrary RNG initialization, silent system-damage approximation or tactical mutation of strategic Core state.

**Stop here. Gate 3 may now implement exactly this contract and nothing broader.**

## Gate 3 implementation evidence

Status: **complete on 2026-09-01; Gate 4 final QA/commit/closure pending**.

Gate 3 implemented exactly the accepted Gate-2 vertical slice without adding any deferred tactical family.

### Persisted strategic weapon identity

- Core `StateSchemaVersion` is **21**.
- `core.ShipWeaponMount{Slot, WeaponID, Count}` is persisted on `ShipDesignSpec.Weapons` and snapshots into built `Ship.Spec`.
- Core validation enforces at most eight mounts, slots0..7, strict ascending/unique slots, nonempty weapon ID and positive count.
- Existing current-schema persistence fixtures were advanced to schema21; there is intentionally no schema20 migration.
- Built Ship weapon slices are deep-copied. A regression revises the source design to unarmed after construction and proves the built Ship remains armed at source revision1.

### Tactical rules and Laser design surface

- Added `data/rulesets/moo2-1.31/tactical_combat.json`, tactical schema **1**, plus strict loader/validator and normalized-ID cross-reference checks.
- `economy.json` remains schema **8**; `ship_hulls.json` remains schema **3**.
- Tactical schema1 contains only the Gate-1 evidenced initiative divisor, exact original RNG multiplier/increment, Beam hit/range tables, standard Laser tech100/space10/cost5/damage1-4, Frigate Armor4/Structure4, Nuclear/Fusion Frigate combat speeds and Electronic Computer Beam Offense25.
- `SaveMilitaryDesignPayload` now accepts either no weapons or exactly one slot0 `laser_cannon` count1. Laser tech100 is required; +10 space/+5 base cost is applied before the existing production-cost modifier.
- Wrong slot/weapon/count/multiple mounts and missing Laser technology reject without creating a design or consuming a state ID.

### Frozen TacticalSpec and exact support predicate

- `game.Encounter`, Session EncounterSpec and `battle.Spec` carry an optional frozen TacticalSpec plus `TacticalUnsupportedReason`.
- Auto TacticalSpec is produced only for one attacker Ship vs one defender Ship, no civilian Fleet context and no defender Colony, with the exact Gate-2 equipment surface: Fusion/Electronic/Titanium/no-shield/standard-fuel/slot0-Laser attacker and Nuclear/Electronic/Titanium/no-shield/standard-fuel/unarmed defender.
- Fixed post-deployment positions remain attacker(10,10) and defender(11,10). No movement, facing or shield state was introduced.
- The strategic resolver constructs the snapshot from its pre-encounter state. GameSession then validates Tactical Ship/equipment/weapon identity against the explicit state passed into encounter preparation.
- A regression changes the attacker only inside the resolver-returned state and proves child Battle preparation uses that returned `resolution.State`, not the stale pre-resolution `s.state`.
- Nearby unsupported strategic encounters still create lifecycle BattleSessions, carry an explicit unsupported reason, reject Tactical commands and remain manually completable through the existing Slice-06 result path.

### BattleSession tactical state, commands and RNG

- Tactical-enabled BattleSession owns round, initiative order, active strategic ShipID, Battle-global next command sequence, uint32 RNG state, ArmorCurrent, StructureDamage, Laser ready state, destroyed state and deterministic local tactical events.
- Round1 uses the evidenced 24-vs-22 initiative ordering. ShipID ascending exists only as a deterministic MOOX tie fallback.
- Public commands are `battle.fire_beam` and `battle.end_activation` using the existing `protocol.Command` envelope plus authoritative caller SeatID.
- Commands are prepared on detached state. Wrong authority/sequence/active Ship/target/slot, double fire or a deferred internal-subsystem selection returns an error with no authoritative RNG, ready-state, event or sequence mutation.
- Tactical-enabled BattleSessions reject external/manual result injection; lifecycle-only unsupported Battles preserve manual completion.
- The accepted fixture pins `InitialRNGState=0x001500BD`. The runtime uses the exact original uint32 LCG/rejection algorithm while `battle.Spec.Seed` retains its separate 64-bit infrastructure meaning.
- The golden event regression asserts the actual consumed values `100,19,100,100`, final RNG `0xFD95EBB9`, first 4-damage hit Armor4->0, second 4-damage Structure0->4, defender destruction and singular `tactical_victory`.
- Local deterministic event kinds are `round_started`, `beam_fired`, `battle_damage_applied`, `activation_ended`, `ship_destroyed` and `winner_determined`; Observer/Battle Views deep-clone Tactical spec/state/events.

### Slice-06 strategic reconciliation and parallelism

- Added `GameSession.SubmitBattleCommand` and generalized the staged final-battle candidate commit path so external lifecycle results and prepared terminal Tactical results share the same atomic strategic continuation boundary.
- Nonterminal Tactical commands mutate only the child BattleSession.
- If a terminal Tactical result completes one of multiple independent Battles, the child can complete while strategic continuation waits for the entire encounter wave.
- For the final active Battle, GameSession first builds outcomes, runs `ResumeAfterEncounters` on a cloned strategic state, validates the returned state and prepares any next-wave Battles from that returned state. Only then can the terminal Tactical child and authoritative strategic state commit.
- A forced continuation-failure regression proves terminal command seq4 leaves Battle/strategic state unchanged and can be retried with the same command sequence. The successful retry removes the concrete defender Ship and empty Fleet through the existing Slice-06 handoff.
- A two-system real-EconomyResolver regression finishes Battle2 wall-clock before Battle1. No local Tactical events leak into GameSession's global event stream, and final global `battle_completed` events remain Battle-ID order1,2 while each completed Battle View retains its own local Tactical history.

### Gate-3 verification

New/updated tests cover the full accepted Gate-3 matrix across Core, Ruleset, Game, Battle and Session: rules schema/value validation; schema21 weapon roundtrip/validation; Laser design tech/cost/snapshotting; exact TacticalSpec and unsupported fallback; initiative/round/readiness; exact four-roll golden trace; command authority and rollback; manual-result rejection; View isolation/replay; returned-state preparation; real strategic casualty handoff/retry; and parallel stable ordering.

Pre-handoff verification passed:

- `go test ./... -count=1` - **PASS all packages**.
- `go vet ./...` - **PASS**.

### Scope still deliberately deferred

Gate 3 does **not** implement deployment changes, movement/turning, facing, shields/recharge, non-Laser or multiple/modified weapon batteries, missiles/torpedoes/bombs/fighters/point-defense, generic internal subsystem damage, repair/regeneration, specials, retreat/capture/self-destruct/stasis, Colony/planet/station combat, race/leader/nonzero-crew tactical bonuses, NPC/monster/Antaran tactical variants, full weapon-design/miniaturization or active mid-battle process-restart persistence.

**Gate 3 is complete. Stop before Gate 4 final QA, commits and slice closure.**
