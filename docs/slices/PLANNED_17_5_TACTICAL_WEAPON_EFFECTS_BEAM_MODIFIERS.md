# Planned Slice 17.5 - Tactical Weapon Effects / Beam Modifiers

Status: **planned; child of Slice 17; not open**.

## Objective

Turn the authoritative Beam mount/modifier metadata created by Slice 17.2 into real server-authoritative Tactical behavior, with one frozen rule path shared by combat resolution and the Ship Designer Weapon Selector preview.

This slice owns the supported Beam Tactical effects for Standard, Heavy and Point Defense mounts plus the supported 17.2 Beam modifier set (including Autofire, Armor Piercing, Continuous, No Range Dissipation, Shield Piercing and Enveloping where the frozen compatibility matrix permits them). Gate 1 must re-verify the exact original behavior before any effect is enabled; no multiplier, range rule, hit rule or interaction may be guessed.

## Inputs

17.5 opens only after:

- 17.0 Reference/Test Harness is closed;
- 17.1 Hulls/Mandatory Systems is closed;
- 17.2 Beam Weapons and Beam Mount Modifiers is closed;
- 17.3 Ordnance is closed;
- 17.4 Specials/Fighters is closed;
- the immutable built-Ship/BattleSession Beam snapshot from 17.2 remains lossless and authoritative.

## Ownership boundary

17.5 owns **Beam Tactical weapon effects** and the corresponding Designer preview contract. It does not silently absorb the complete missile/torpedo/bomb projectile simulation, fighter/carrier combat, planetary bombardment or unrelated ship-special effects. Those remain with their explicit later Tactical owners unless Gate 1 proves a small dependency must be implemented here to make a Beam effect correct.

The same configured Beam row must mean the same thing in:

1. the Weapon Selector preview;
2. the saved ShipDesignSpec;
3. the immutable built Ship;
4. BattleSession/Tactical resolution;
5. Observer/replay/persistence projections.

No browser-only combat formula is allowed.

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

- [ ] Re-audit direct original evidence for Standard / Heavy / Point Defense Beam behavior.
- [ ] Re-audit each supported 17.2 Beam modifier's exact Tactical effect and legal compatibility.
- [ ] Resolve damage/range/hit/shot/defense interaction semantics and exact rounding/order boundaries.
- [ ] Identify Beam-specific effects already transported as metadata but not yet consumed by Tactical.
- [ ] Audit current Weapon Selector preview API/model for a single server-owned effective-stat path.
- [ ] Separate Beam effects owned here from missile/ordnance, fighter and unrelated ship-special Tactical work.
- [ ] Produce deterministic original/reference vectors for representative early-, mid- and late-game Beams.

## Gate 2 - Tactical Beam contract freeze

- [ ] Freeze the exact supported Tactical Beam effect matrix.
- [ ] Freeze Standard / Heavy / Point Defense semantics and incompatibilities.
- [ ] Freeze each supported modifier's rule, interaction order and rounding behavior.
- [ ] Freeze base-damage versus configured/effective-damage terminology and Selector preview fields.
- [ ] Freeze server API/snapshot fields required by Tactical and Weapon Selector preview.
- [ ] Freeze stable unsupported/rejection behavior for any effect still deferred.
- [ ] Freeze deterministic Combat/Designer Lab vectors proving Selector/Tactical parity.

## Gate 3 - implementation

- [ ] Implement the frozen Beam Tactical effect rules in server-authoritative code.
- [ ] Make Tactical consume the existing immutable Beam snapshot without translation hacks.
- [ ] Implement Heavy and Point Defense Tactical consumers to the frozen contract.
- [ ] Implement the frozen supported Beam modifier consumers.
- [ ] Feed the Weapon Selector configured/effective damage and other changed Tactical preview stats from the same authoritative rules.
- [ ] Preserve raw catalog base damage separately from configured/effective values.
- [ ] Preserve save/load/replay/Observer behavior for configured Beam rows.
- [ ] Add deterministic Standard/Heavy/Point-Defense and modifier-combination regressions.

## Gate 4 - QA + close

- [ ] Selector base/configured damage and other preview effects match frozen reference vectors.
- [ ] Heavy and Point Defense visibly produce the correct configured preview and the same Tactical result.
- [ ] Supported modifiers produce their exact frozen Tactical effects and combinations.
- [ ] Selector -> Save -> Build -> BattleSession -> Tactical preserves the exact configured Beam semantics.
- [ ] Mobile/desktop Weapon Selector remains readable when base and configured/effective stats are shown.
- [ ] Unsupported/deferred non-Beam Tactical mechanics remain explicit rather than silently simulated.
- [ ] Full Go/vet/web/browser/persistence/replay/reference-game QA passes.

## Exit criterion

Slice 17.5 closes only when every Beam mount/modifier effect accepted into its frozen matrix has a real authoritative Tactical consumer and the Weapon Selector previews the same configured result from the same rule source. In particular, Heavy/Point Defense and supported Beam modifiers may no longer appear as working configuration choices while their combat effect is merely deferred metadata.

After 17.5 closes, Slice 17.6 performs the full Ship Designer integration/acceptance closeout for the parent Slice-17 program.
