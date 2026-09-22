# Planned Slice 17.5 - Tactical Weapon Effects / Beam Modifiers

Status: **planned; child of Slice 17; not open**.

## Objective

Replace the remaining Slice-15.5 baseline-only Tactical combatant boundary with the verified Slice-17.1 hull/mandatory-system breadth, then turn the authoritative Beam mount/modifier metadata created by Slice 17.2 into real server-authoritative Tactical behavior. Combat resolution and the Ship Designer Weapon Selector preview must share one frozen rule path.

This slice first owns the Tactical combatant breadth required to admit the designs already legal under Slice 17.1: Frigate through Doom Star plus their supported Computers, Armors, Shields, Drives and Fuel Cells. It then owns the supported Beam Tactical effects for Standard, Heavy and Point Defense mounts plus the supported 17.2 Beam modifier set (including Autofire, Armor Piercing, Continuous, No Range Dissipation, Shield Piercing and Enveloping where the frozen compatibility matrix permits them). Gate 1 must re-verify the exact original behavior before any effect is enabled; no hull durability, Beam Offense, Shield, combat-speed, damage, range, hit or interaction rule may be guessed.

## Inputs

17.5 opens only after:

- 17.0 Reference/Test Harness is closed;
- 17.1 Hulls/Mandatory Systems is closed;
- 17.2 Beam Weapons and Beam Mount Modifiers is closed;
- 17.3 Ordnance is closed;
- 17.4 Specials/Fighters is closed;
- the immutable built-Ship/BattleSession Beam snapshot from 17.2 remains lossless and authoritative.

## Ownership boundary

17.5 owns **broad Tactical combatant admission for the Slice-17.1 hull/mandatory-system matrix**, **Beam Tactical weapon effects**, and the corresponding Designer preview contract. It must remove the old exact Frigate/Electronic/Titanium/no-Shield/Standard-Fuel admission assumption only by replacing its hard-coded Tactical values with verified authoritative hull/component rules. It does not silently absorb the complete missile/torpedo/bomb projectile simulation, fighter/carrier combat, planetary bombardment or unrelated ship-special effects. Those remain with their explicit later Tactical owners unless Gate 1 proves a small dependency must be implemented here to make a Beam effect correct.

The same configured Beam row must mean the same thing in:

1. the Weapon Selector preview;
2. the saved ShipDesignSpec;
3. the immutable built Ship;
4. BattleSession/Tactical resolution;
5. Observer/replay/persistence projections.

No browser-only combat formula is allowed.

## Broad Tactical combatant prerequisite

The current Tactical runtime still carries a Slice-15.5 fixture boundary: it admits only Frigate + Electronic Computer + Titanium Armor + no Shield + Standard Fuel and then constructs Tactical durability/offense from the old Frigate/Electronic constants. Slice 17.5 must replace that fixture path before broader Beam effects can be accepted.

- Frigate, Destroyer, Cruiser, Battleship, Titan and Doom Star must enter Tactical only with verified per-hull Structure/base durability semantics.
- Titanium, Tritanium, Zortrium, Neutronium, Adamantium and Xentronium must apply their verified Armor/Structure effects; no late Armor may merely carry an ID while using Titanium hit points.
- Supported Computers must supply their verified Beam Offense rather than always using the Electronic Computer value.
- Shield choices from no Shield through Class X must receive their verified Tactical block/capacity behavior; facing/recharge details are frozen from direct evidence before implementation.
- Supported Drives must provide verified combat-speed behavior for the legal 17.1 drive matrix rather than the old Nuclear/Fusion-only subset.
- Fuel Cell identity must not block Tactical admission when it has no Tactical effect; if direct evidence proves a combat effect, that effect must be frozen and implemented instead.
- The admission gate must reject only a specifically unsupported mechanic, not an otherwise legal ship merely because it differs from the old Slice-15.5 fixture.
- No guard may simply be deleted while the Tactical snapshot still substitutes Frigate/Electronic/Titanium constants.

## Weapon Selector damage / effect preview

The Weapon Selector must stop treating a static catalog damage badge as though it were the complete configured result.

- The catalog weapon's frozen min/max value remains the authoritative **base damage / Basisschaden**.
- The selector must calculate and display the **configured/effective damage** correctly for the currently selected mount and modifiers whenever those choices change Tactical damage.
- Heavy, Point Defense and every supported modifier must update the preview fields they actually affect (damage, range, hit behavior, shot behavior, armor/shield interaction or other frozen Beam effect).
- If a choice does not alter raw damage but changes another Tactical property, the UI must keep base damage truthful and show the changed property instead of inventing a damage delta.
- Base damage and configured/effective damage must be visually distinguishable when they differ.
- The preview must be returned/derived from the same server-owned frozen rule implementation used by Tactical combat, including the same rounding/order rules.
- Editing an already installed weapon must preserve its installed count while recomputing preview effects for changed mount/modifiers.

## Gate 1 - original-effect audit

- [ ] Re-audit the Slice-15.5 Tactical combatant guard and every hard-coded Frigate/Electronic/Titanium/no-Shield/Standard-Fuel value it protects.
- [ ] Re-audit direct original evidence for all six hull durability baselines, Armor multipliers/effects, Computer Beam Offense, Shield block/capacity/facing/recharge, drive combat speed and any Tactical relevance of Fuel Cells.
- [ ] Re-audit direct original evidence for Standard / Heavy / Point Defense Beam behavior.
- [ ] Re-audit each supported 17.2 Beam modifier's exact Tactical effect and legal compatibility.
- [ ] Resolve damage/range/hit/shot/defense interaction semantics and exact rounding/order boundaries.
- [ ] Identify Beam-specific effects already transported as metadata but not yet consumed by Tactical.
- [ ] Audit current Weapon Selector preview API/model for a single server-owned effective-stat path.
- [ ] Separate broad hull/mandatory-system + Beam effects owned here from missile/ordnance, fighter and unrelated ship-special Tactical work.
- [ ] Produce deterministic original/reference vectors for Frigate/Destroyer/Cruiser/Battleship/Titan/Doom Star with representative early/mid/late mandatory systems.
- [ ] Produce deterministic original/reference vectors for representative early-, mid- and late-game Beams.

## Gate 2 - Tactical Beam contract freeze

- [ ] Freeze the exact broad Tactical combatant matrix: hull durability, Armor, Computer, Shield, Drive and Fuel admission/effect behavior.
- [ ] Freeze replacement semantics for the old baseline-only admission guard so no legal 17.1 design inherits Frigate/Electronic/Titanium constants.
- [ ] Freeze the exact supported Tactical Beam effect matrix.
- [ ] Freeze Standard / Heavy / Point Defense semantics and incompatibilities.
- [ ] Freeze each supported modifier's rule, interaction order and rounding behavior.
- [ ] Freeze base-damage versus configured/effective-damage terminology and Selector preview fields.
- [ ] Freeze server API/snapshot fields required by Tactical and Weapon Selector preview.
- [ ] Freeze stable unsupported/rejection behavior for any effect still deferred.
- [ ] Freeze deterministic Combat/Designer Lab vectors proving hull/component Tactical values and Selector/Tactical Beam parity.

## Gate 3 - implementation

- [ ] Replace the baseline-only Tactical combatant path with the frozen hull/mandatory-system Tactical rules.
- [ ] Remove the exact Frigate/Electronic/Titanium/no-Shield/Standard-Fuel gate only after Tactical snapshots derive correct hull durability, Armor, Computer, Shield and Drive values.
- [ ] Keep Fuel Cells admission-neutral unless the frozen evidence contract assigns them a Tactical effect.
- [ ] Implement the frozen Beam Tactical effect rules in server-authoritative code.
- [ ] Make Tactical consume the existing immutable Beam snapshot without translation hacks.
- [ ] Implement Heavy and Point Defense Tactical consumers to the frozen contract.
- [ ] Implement the frozen supported Beam modifier consumers.
- [ ] Feed the Weapon Selector configured/effective damage and other changed Tactical preview stats from the same authoritative rules.
- [ ] Preserve raw catalog base damage separately from configured/effective values.
- [ ] Preserve save/load/replay/Observer behavior for configured Beam rows.
- [ ] Add deterministic all-hull/mandatory-system Tactical admission and derived-stat regressions.
- [ ] Add deterministic Standard/Heavy/Point-Defense and modifier-combination regressions.

## Gate 4 - QA + close

- [ ] Every legal 17.1 hull/mandatory-system representative enters Tactical with its frozen durability/offense/defense/speed values; no hidden Frigate/Electronic/Titanium substitution remains.
- [ ] Previously observed richer-design encounters no longer fail merely with the Slice-15.5 Frigate/Electronic/Titanium/no-shield/standard-fuel message.
- [ ] Selector base/configured damage and other preview effects match frozen reference vectors.
- [ ] Heavy and Point Defense visibly produce the correct configured preview and the same Tactical result.
- [ ] Supported modifiers produce their exact frozen Tactical effects and combinations.
- [ ] Selector -> Save -> Build -> BattleSession -> Tactical preserves the exact configured Beam semantics.
- [ ] Mobile/desktop Weapon Selector remains readable when base and configured/effective stats are shown.
- [ ] Unsupported/deferred non-Beam Tactical mechanics remain explicit rather than silently simulated.
- [ ] Full Go/vet/web/browser/persistence/replay/reference-game QA passes.

## Exit criterion

Slice 17.5 closes only when the old Slice-15.5 baseline-only combatant restriction has been replaced by verified Tactical support for the accepted Slice-17.1 hull/mandatory-system matrix, every Beam mount/modifier effect accepted into its frozen matrix has a real authoritative Tactical consumer, and the Weapon Selector previews the same configured result from the same rule source. A Titan or shielded/late-system ship must never be admitted with hidden Frigate/Electronic/Titanium values, and Heavy/Point Defense or supported Beam modifiers may no longer appear as working configuration choices while their combat effect is merely deferred metadata.

After 17.5 closes, Slice 17.6 performs the full Ship Designer integration/acceptance closeout for the parent Slice-17 program.
