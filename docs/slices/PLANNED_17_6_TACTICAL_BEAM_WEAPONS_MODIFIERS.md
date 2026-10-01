# Planned Slice 17.6 - Tactical Beam Weapons & Beam Modifiers

Status: **open; Gates 1-2 complete; Gate 3 checkpoint 1 complete; checkpoint 2 ready (2026-10-01)**.

## Objective

Turn the authoritative Beam weapon, mount, modifier and firing-arc snapshots created by Slice 17.2 plus follow-up 17.2A into complete supported server-authoritative Tactical Beam combat on top of the 17.5 foundation.

17.6 owns Beam hit resolution, range, damage, mount behavior, supported Beam modifiers, Weapon Selector effective-stat parity and Beam-specific event-driven presentation. It extends the accepted Laser vertical slice to the full frozen Beam matrix without moving authority into React.

## Inputs

17.6 opens only after:

- 17.2 Beam Weapons and Beam Mount Modifiers is closed;
- 17.2A Beam Weapon Firing Arcs is closed;
- 17.5 Tactical Combat Foundation / Ship Combatant Breadth is closed;
- the immutable built-Ship/BattleSession Beam snapshot remains lossless and authoritative.

## Ownership boundary

17.6 owns:

- all supported Beam identities frozen by 17.2;
- all supported Beam firing-arc identities/cost/persistence metadata frozen by 17.2A;
- Standard, Heavy and Point Defense Beam mount semantics;
- supported Beam modifiers such as Autofire, Armor Piercing, Continuous, No Range Dissipation, Shield Piercing and Enveloping where allowed by the frozen compatibility matrix;
- server-authoritative Beam hit rolls, range, damage, firing-arc target legality and effect ordering;
- Beam readiness/shot consumption;
- Beam-side Point Defense targeting/interceptor semantics;
- Weapon Selector base-versus-configured/effective damage and affected-stat preview;
- Beam/impact/shield/hull presentation driven only by authoritative combat events.

17.6 does not own missile projectile flight, torpedo flight, bomb resolution or fighter/bomber actor simulation. Full missile-interception end-to-end behavior is completed in 17.7 using the frozen Beam Point Defense contract from 17.6.

## Weapon Selector configured/effective preview

The Weapon Selector must stop treating a static catalog damage badge as the complete configured result.

- The catalog weapon min/max remains authoritative **base damage / Basisschaden**.
- The selector displays **configured/effective damage** when mount/modifier choices change damage.
- If a choice changes range, hit chance, shot behavior, armor/shield interaction or another Tactical property rather than raw damage, the changed property is shown explicitly instead of inventing a damage delta.
- Base and configured/effective damage remain visually distinguishable when different.
- The preview is derived from the same server-owned frozen rule implementation used by Tactical combat, including identical rounding/order semantics.
- Editing an installed weapon preserves its installed count while recomputing mount/modifier preview effects.
- React contains no duplicate Beam combat formula.

## Tactical Beam presentation / VFX

- Beam fire is visualized from authoritative fire events using the persisted firing/target ship coordinates.
- Hits and misses are visibly distinct.
- Grouped/Autofire shots remain individually understandable rather than collapsing authoritative shots into one misleading visual.
- Shield impact, Armor/Structure impact and destruction feedback consume authoritative damage-layer/result events.
- Beam identity may have distinct MOOX-native visual treatment only after a presentation contract is frozen; visual style does not alter hit/damage authority.
- Animation lifetime and completion never gate BattleSession progression.

## Gate 1 - original Beam-effect audit

- [x] Re-audit direct original evidence for every supported Beam identity's Tactical range/hit/damage behavior.
- [x] Re-audit Standard / Heavy / Point Defense behavior and incompatibilities.
- [x] Re-audit each supported 17.2 Beam modifier's exact Tactical effect and legal compatibility.
- [x] Resolve hit-roll inputs, distance/range effects, damage rolls, shot counts, defense interaction and exact rounding/order boundaries.
- [x] Audit Beam-side Point Defense semantics and the handoff required for 17.7 projectile interception.
- [x] Audit current Weapon Selector preview API/model for one server-owned effective-stat path.
- [x] Audit existing `beam_fired`/damage events and current Laser animation/readiness behavior.
- [x] Produce deterministic original/reference vectors for representative early-, mid- and late-game Beams.

Evidence: docs/research/SLICE_17_6_GATE2_TACTICAL_BEAM_CONTRACT_FREEZE_2026-09-30.md - Gate 2 frozen/complete.

## Gate 2 - Tactical Beam contract freeze

- [x] Freeze the exact supported Tactical Beam effect matrix.
- [x] Freeze Standard / Heavy / Point Defense semantics and incompatibilities.
- [x] Freeze every supported modifier's rule, interaction order and rounding behavior.
- [x] Freeze hit-roll, range, damage-roll and readiness/shot-consumption authority.
- [x] Freeze Beam-side Point Defense interceptor hooks consumed later by 17.7.
- [x] Freeze base-damage versus configured/effective-damage terminology and Selector preview fields.
- [x] Freeze server API/snapshot/event fields required by Tactical, replay and Weapon Selector preview.
- [x] Freeze Beam event/VFX presentation semantics without making timing authoritative.
- [x] Freeze deterministic Combat/Designer Lab vectors proving Selector/Tactical parity.

## Gate 3 - implementation

- [ ] Implement the frozen Beam Tactical effect rules in server-authoritative code.
- [ ] Make Tactical consume the existing immutable Beam snapshot without translation hacks.
- [ ] Implement Standard, Heavy and Point Defense Beam consumers.
- [ ] Implement supported Beam modifier consumers.
- [ ] Implement server-authoritative Beam hit and damage rolls with exact frozen ordering.
- [ ] Feed Weapon Selector configured/effective damage and other changed Tactical preview stats from the same rules.
- [ ] Preserve raw catalog base damage separately from configured/effective values.
- [ ] Extend Beam fire/impact/shield/damage event payloads only as required by the frozen contract.
- [ ] Implement Beam-specific presentation from those authoritative events.
- [ ] Preserve save/load/replay/Observer behavior for configured Beam rows.
- [ ] Add deterministic Beam/mount/modifier combinations and repeated/grouped-fire regressions.

## Gate 4 - QA + close

- [ ] Every supported Beam can be selected in Tactical and resolves through server-projected legal actions.
- [ ] Hit/miss, range and damage rolls match frozen reference vectors.
- [ ] Heavy and Point Defense produce the correct configured preview and Tactical result.
- [ ] Supported modifiers produce exact frozen Tactical effects and combinations.
- [ ] Selector -> Save -> Build -> BattleSession -> Tactical preserves exact configured Beam semantics.
- [ ] Base/configured damage and affected stats match between Selector and Combat Lab.
- [ ] Beam/impact/shield/damage animations correspond to authoritative events and remain presentation-only.
- [ ] Desktop/mobile multi-weapon selection remains usable.
- [ ] Full Go/vet/web/browser/persistence/replay/reference-game QA passes.

## Exit criterion

Slice 17.6 closes when every Beam mount/modifier effect accepted into its frozen matrix has a real authoritative Tactical consumer, Beam hit/damage resolution is server-owned, the Weapon Selector previews the same configured result from the same rule source, and the Tactical UI visibly presents the authoritative Beam outcomes without client-side gameplay formulas.

After 17.6 closes, Slice 17.7 adds complete supported ordnance/projectile combat and end-to-end missile interception.
