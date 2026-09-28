# Slice 17.5 Gate 1 - Tactical Combat Foundation Audit

Date: 2026-09-28
Status: **in progress; current Slice-15.5 admission/runtime boundary mapped; original broad combatant values still open**

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

## Gate 1 remaining evidence work

- recover/freeze Frigate, Destroyer, Cruiser, Battleship, Titan and Doom Star Structure/base durability values;
- recover/freeze Titanium through Xentronium Armor effects/multipliers and exact rounding/application;
- recover/freeze Electronic through Moleculartronic Computer Beam-Offense inputs;
- recover/freeze no-shield through Class X Shield damage-layer semantics, including capacity/block/facing/recharge only where directly evidenced;
- recover/freeze Nuclear through Interphased Drive combat speed / minimum speed semantics;
- determine whether Fuel Cell identity has any Tactical effect; otherwise freeze admission-neutral behavior;
- audit generic action/readiness projection and isolate unsupported per-weapon actions without rejecting the encounter;
- audit the common Tactical event envelope and UI assumptions that remain Laser-specific;
- produce early/mid/late representative vectors across all six hull classes.

## Gate 1 status

The code-boundary audit is complete enough to establish the replacement surface, but Gate 1 remains open until the original-value matrix and unsupported-action contract are proven.
