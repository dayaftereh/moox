# Slice 17.5 Gate 1 - Tactical Combat Foundation Audit

Date: 2026-09-28
Status: **complete; broad combatant matrix, action/event boundary and representative vectors directly frozen for Gate 2**

## Purpose

Audit the current Tactical combatant admission/runtime foundation before replacing the accepted Slice-15.5 baseline-only path. Gate 1 is evidence/audit only; no protected guard is removed here.

## Current admission boundary

`internal/game/tactical_combat.go` still owns `baselineCombatantUnsupportedReason`.

It currently rejects any combatant that is not:

- Frigate hull;
- Electronic Computer;
- Titanium Armor;
- no primary Shield;
- Standard Fuel Cells.

It also requires a Tactical drive rule, validates at most eight ascending weapon slots, and rejects unknown weapon rows. Non-Beam rows are admitted only when their Tactical support mode is `metadata_only`.

## Hidden baseline substitutions still present

`baselineTacticalShip` copies the strategic Hull/Drive/Computer/Armor identities but still derives combat values from singleton baseline rules:

- Beam Offense always comes from the one `rules.Computer` row;
- ArmorMax always comes from the one `rules.Frigate.ArmorHits` row;
- StructureMax always comes from the one `rules.Frigate.Structure` row;
- Beam Defense is initialized to 0;
- Shield identity/state is not present in TacticalShipSpec;
- Fuel Cell identity is not present in TacticalShipSpec.

This means deleting only the Game-layer admission guard would silently admit richer designs while continuing to simulate Frigate/Electronic/Titanium durability/offense.

## BattleSession duplicate baseline guard

`internal/battle/tactical.go` independently validates the same Slice-15.5 combatant baseline.

`validateBaselineCombatant` currently requires:

- Frigate;
- Electronic Computer;
- Titanium Armor;
- ArmorMax = 4;
- StructureMax = 4;
- Fusion Drive pristine combat speed 22 or Nuclear Drive pristine combat speed 20.

Therefore broad admission requires coordinated Game-layer and Battle-layer replacement; changing only encounter projection is insufficient.

## Current Tactical ruleset shape

`data/rulesets/moo2-1.31/tactical_combat.json` and `internal/ruleset/tactical_combat.go` currently contain:

- one Frigate baseline row: Titanium / Standard Fuel / ArmorHits 4 / Structure 4;
- one Electronic Computer row: Beam Offense 25;
- only Nuclear and Fusion drive combat-speed rows;
- Beam/RNG/initiative rules plus the growing weapon metadata table.

The schema does not yet contain:

- all six hull durability baselines;
- the six Armor Tactical durability/effect rows;
- all five Computer Beam-Offense rows;
- Shield Tactical layer/capacity/block/recharge rules;
- all six Drive combat-speed rows;
- any Tactical Fuel-Cell behavior classification.

## Damage-layer gap

Current BattleSession damage resolution is Armor -> Structure only.

`TacticalShipState` tracks ArmorCurrent and StructureDamage. `TacticalShipSpec` tracks ArmorMax and StructureMax. There is no primary-Shield state/layer in the shared Tactical foundation yet.

Beam damage currently:

1. consumes Armor until zero;
2. applies remaining damage to Structure;
3. emits `battle_damage_applied` with armor/structure before/after payload;
4. destroys the ship when StructureDamage reaches StructureMax.

Shield behavior therefore requires an explicit Gate-2 contract before implementation; it cannot be represented truthfully by merely adding ShieldID metadata.

## Existing foundation that can be preserved

The audit also confirms several useful foundations already exist and should not be rewritten:

- BattleSession is the authoritative Tactical state owner;
- deterministic Tactical RNG state and ordered server-side roll consumption already exist;
- Tactical combat already supports N-vs-M ship lists rather than attacker/defender singletons;
- weapon slots, immutable weapon metadata and Special metadata are projected into TacticalShipSpec;
- movement/turning and event-driven browser presentation already exist;
- non-Beam metadata-only weapons can already be carried through the Tactical snapshot without executing their deferred behavior;
- Inertial Stabilizer/Nullifier already connect to the existing turning-mode consumer.

## Relationship to Slice 17.1 data

Slice 17.1 provides the complete legal design identities and economics for:

- six hulls;
- six drives;
- five computers;
- six armors;
- five shields plus no-shield;
- five fuel cells.

However `ship_hulls.json` intentionally contains design/runtime identity, Space/PP, FTL/range and provenance data, not the full Tactical combat values required by 17.5. Tactical hull durability, computer offense, armor durability, shield behavior and the full drive combat-speed matrix still require original-evidence freeze.


## Direct original broad-combatant matrix closure (2026-09-28)

Gate 1 was completed against the local 1.31 executable:

`reference/original/support/files/Orion2.exe`

SHA-256:

`7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5`

The values below are not extrapolated from the old Frigate fixture. They were decoded from the original object tables and the routines that consume those tables.

### Hull durability

`Get_Ship_Armor_Hits_` reads the hull base Armor word and `Get_Ship_Structure_` reads the paired base Structure word from the original hull table. The six combat hull rows are:

| Hull | base Armor | base Structure |
| --- | ---: | ---: |
| Frigate | 4 | 4 |
| Destroyer | 10 | 10 |
| Cruiser | 30 | 30 |
| Battleship | 50 | 50 |
| Titan | 80 | 80 |
| Doom Star | 150 | 150 |

This directly closes the Titan/Doom-Star values and replaces the hidden Frigate 4/4 substitution.

### Armor material durability

`Get_Ship_Armor_Hits_` reads the installed Armor material percentage from the original 15-byte Armor record, applies `base + floor(base * bonusPercent / 100)`, and only then reaches Special hooks such as Heavy Armor.

| Armor | original bonus | foundation multiplier |
| --- | ---: | ---: |
| Titanium | 0% | 1x |
| Tritanium | 100% | 2x |
| Zortrium | 300% | 4x |
| Neutronium | 500% | 6x |
| Adamantium | 700% | 8x |
| Xentronium | 900% | 10x |

Heavy Armor and Armor-Piercing interaction remain Special/weapon-family consumers owned by later slices.

### Computer Beam Offense

The five Computer records are contiguous 22-byte original rows. The technology sequence 58/122/143/44/110 directly identifies Electronic/Optronic/Positronic/Cybertronic/Moleculartronic; the final word of each record is the same Beam-Offense input already proven for Electronic in the baseline.

| Computer | Beam Offense |
| --- | ---: |
| Electronic | 25 |
| Optronic | 50 |
| Positronic | 75 |
| Cybertronic | 100 |
| Moleculartronic | 125 |

### Drive combat speed

`Current_Design_Min_Combat_Speed_` and `Current_Design_Max_Combat_Speed_` read the original drive/hull matrices directly. Values are Frigate / Destroyer / Cruiser / Battleship / Titan / Doom Star:

| Drive | minimum/damage floor | empty-design maximum |
| --- | --- | --- |
| Nuclear | 10 / 8 / 6 / 5 / 4 / 3 | 20 / 18 / 16 / 15 / 14 / 13 |
| Fusion | 12 / 10 / 8 / 7 / 6 / 5 | 22 / 20 / 18 / 17 / 16 / 15 |
| Ion | 14 / 12 / 10 / 9 / 8 / 7 | 24 / 22 / 20 / 19 / 18 / 17 |
| Anti-Matter | 16 / 14 / 12 / 11 / 10 / 9 | 26 / 24 / 22 / 21 / 20 / 19 |
| Hyper | 18 / 16 / 14 / 13 / 12 / 11 | 28 / 26 / 24 / 23 / 22 / 21 |
| Interphased | 20 / 18 / 16 / 15 / 14 / 13 | 30 / 28 / 26 / 25 / 24 / 23 |

These matrices reconcile exactly with the already accepted Slice-17.1 Designer expression:

`combatSpeed = hullBaseSpeed + driveBonus + ceil(10 * unusedSpace / hullSpace)`

where hull base speed is `10/8/6/5/4/3`, drive bonus is `0/2/4/6/8/10`, and the unused-Space term is clamped to the 0..10 range. The direct minimum is hull base + drive bonus; the direct maximum is minimum + 10. `Current_Design_Base_Combat_Speed_` then reduces that pristine design speed toward the minimum as engine damage accumulates.

The accepted original display/rules contract also gives:

`BeamDefense = CurrentCombatSpeed * 5`

before later ECM/Inertial/crew/race modifiers. The current Tactical `BeamDefense=0` projection is therefore another protected Slice-15.5 substitution to replace in Gate 3.

### Primary Shield foundation

The original shield-strength field is 1 / 3 / 5 / 7 / 10 for Class I / III / V / VII / X.

`Calc_Combat_Shield_Max_` proves four per-facing-arc pools. The hull factor is 1/2/3/4/5/7 for Frigate through Doom Star, so:

`shieldArcMax = shieldStrength * 5 * hullFactor`

| Shield | per-hit block | Frg | Des | Cru | BB | Titan | Doom Star |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| none | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| Class I | 1 | 5 | 10 | 15 | 20 | 25 | 35 |
| Class III | 3 | 15 | 30 | 45 | 60 | 75 | 105 |
| Class V | 5 | 25 | 50 | 75 | 100 | 125 | 175 |
| Class VII | 7 | 35 | 70 | 105 | 140 | 175 | 245 |
| Class X | 10 | 50 | 100 | 150 | 200 | 250 | 350 |

`Load_Combat_Ship_` writes the calculated maximum to the per-ship shield-max field and initializes all four 2-byte facing arcs to that maximum.

`Facing_Shield_` directly maps relative bearing into four 90-degree arcs:
- <=45 or >315 -> arc 0;
- 46..135 -> arc 3;
- 136..225 -> arc 2;
- 226..315 -> arc 1.

The Beam path computes relative bearing and `Facing_Shield_` before `Apply_Ship_Damage_`.

`Apply_Ship_Damage_` directly closes the normal shield sequence:
1. subtract installed Shield strength from incoming damage;
2. later Special hook Hard Shields can subtract an additional 3 (deferred to Slice 17.8);
3. clamp reduced damage at zero;
4. consume the facing arc pool by the remaining damage;
5. only facing-arc overflow continues to the lower damage layers.

`Recharge_Shields_` directly closes base recharge:
- ships without a shield or with a dead shield generator do not recharge;
- normal recharge budget is `floor(shieldArcMax * 3 / 10)` per turn;
- the routine repeatedly walks the four arcs in stable order and adds one point to deficient arcs until that budget is exhausted or all arcs are full;
- Shield Capacitors changes the factor from 3/10 to 10/10 and remains a Slice-17.8 Special consumer.

`Shield_Strength_` / `Current_Design_Shield_Absorption_` also expose later Special hooks (Multi-Phased Shields / Hard Shields). Those modifiers are not folded into the 17.5 foundation.

### Fuel Cells

The original tactical ship record contains Hull, Shield, Drive, Computer and Armor identities plus combat state, but no Fuel-Cell runtime field or Fuel-derived combat calculation was found in the tactical load/combat paths. Fuel Cells therefore remain legal design provenance and strategic range authority but are **Tactical admission-neutral** in Slice 17.5.

MOOX may carry `FuelCellID` into the immutable Tactical snapshot for truthful identity/Observer display; it must not change movement, initiative, damage or action availability.

## Current action / readiness / event audit

The current `BattleSession` remains the correct shared authority and already owns:
- one ordered mutable Tactical RNG stream;
- immutable Tactical weapon slots plus per-slot mutable readiness;
- movement, Beam fire, wait/end-activation and retreat commands;
- command sequence validation and atomic prepared-command commit;
- detached Tactical Observer snapshots;
- ordered Tactical events.

Current command handlers are Beam-specific only where expected. The current event envelope itself is already family-neutral:

`TacticalEvent{sequence, kind, seat_id?, command_sequence?, data?}`

Existing event kinds include `round_started`, `ship_moved`, `beam_fired`, `battle_damage_applied`, `beam_volley_resolved`, `ship_destroyed`, `winner_determined`, `activation_waited`, `activation_auto_ended`, `activation_ended` and `side_retreated`.

The browser consumes those server events for effects and feedback. It does not need authority-side hit/damage formulas.

The remaining admission problem is therefore not the event model: it is the Game/Battle baseline guard and the lack of a generic per-slot action-availability projection. Gate 2 can preserve the event envelope and replace whole-encounter rejection with per-action unavailability.

## Representative foundation vectors

Gate 1 freezes three mandatory-system tiers over every hull. The speed columns are the exact 0%-free / 100%-free boundaries; a concrete design lies between them using the accepted unused-Space formula.

### Early: Nuclear / Electronic / Titanium / no Shield / Standard Fuel

| Hull | Structure | Armor | Speed min..max | Beam Offense | Beam Defense range |
| --- | ---: | ---: | --- | ---: | --- |
| Frigate | 4 | 4 | 10..20 | 25 | 50..100 |
| Destroyer | 10 | 10 | 8..18 | 25 | 40..90 |
| Cruiser | 30 | 30 | 6..16 | 25 | 30..80 |
| Battleship | 50 | 50 | 5..15 | 25 | 25..75 |
| Titan | 80 | 80 | 4..14 | 25 | 20..70 |
| Doom Star | 150 | 150 | 3..13 | 25 | 15..65 |

### Mid: Ion / Positronic / Zortrium / Class III / Iridium Fuel

| Hull | Structure | Armor | Shield/arc | Base recharge | Speed min..max | Offense | Defense range |
| --- | ---: | ---: | ---: | ---: | --- | ---: | --- |
| Frigate | 4 | 16 | 15 | 4 | 14..24 | 75 | 70..120 |
| Destroyer | 10 | 40 | 30 | 9 | 12..22 | 75 | 60..110 |
| Cruiser | 30 | 120 | 45 | 13 | 10..20 | 75 | 50..100 |
| Battleship | 50 | 200 | 60 | 18 | 9..19 | 75 | 45..95 |
| Titan | 80 | 320 | 75 | 22 | 8..18 | 75 | 40..90 |
| Doom Star | 150 | 600 | 105 | 31 | 7..17 | 75 | 35..85 |

Class III block is 3 per hit.

### Late: Interphased / Moleculartronic / Xentronium / Class X / Thorium Fuel

| Hull | Structure | Armor | Shield/arc | Base recharge | Speed min..max | Offense | Defense range |
| --- | ---: | ---: | ---: | ---: | --- | ---: | --- |
| Frigate | 4 | 40 | 50 | 15 | 20..30 | 125 | 100..150 |
| Destroyer | 10 | 100 | 100 | 30 | 18..28 | 125 | 90..140 |
| Cruiser | 30 | 300 | 150 | 45 | 16..26 | 125 | 80..130 |
| Battleship | 50 | 500 | 200 | 60 | 15..25 | 125 | 75..125 |
| Titan | 80 | 800 | 250 | 75 | 14..24 | 125 | 70..120 |
| Doom Star | 150 | 1500 | 350 | 105 | 13..23 | 125 | 65..115 |

Class X block is 10 per hit. Thorium remains Tactical-neutral.

## Gate 1 conclusion

Gate 1 is **complete**.

The replacement surface is now directly evidenced:
- all six Hull base durability rows;
- all six Armor durability rows;
- all five Computer Beam-Offense rows;
- no-Shield plus all five primary Shield block/capacity/facing/recharge rows;
- all six Drive min/max matrices plus the accepted concrete-design unused-Space formula;
- base Beam Defense from current combat speed;
- Fuel Cell Tactical neutrality;
- current BattleSession RNG/action/readiness/event authority;
- explicit 18-vector early/mid/late hull coverage.

Gate 2 may now freeze the broad Tactical foundation. No Slice-15.5 admission guard is removed by this audit.
