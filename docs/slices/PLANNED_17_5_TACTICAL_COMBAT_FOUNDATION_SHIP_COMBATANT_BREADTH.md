# Planned Slice 17.5 - Tactical Combat Foundation / Ship Combatant Breadth

Status: **open; Gates 1-2 complete; Gate 3 Checkpoints 1-2 complete; Checkpoint 3 ready / not started**.

## Objective

Replace the remaining Slice-15.5 baseline-only Tactical combatant boundary with a broad, server-authoritative Tactical foundation that can admit the legal Slice-17.1 ship matrix without silently substituting Frigate/Electronic/Titanium constants.

17.5 is the common foundation for later weapon-family slices. It makes Tactical a real destination for broad Ship Designer output, standardizes shared combat authority/event/damage/presentation contracts, and ensures unsupported weapon families are isolated as explicit per-weapon capabilities instead of blocking an otherwise legal encounter.

17.5 does **not** implement the complete Beam, missile/torpedo/bomb or fighter/special rule sets. Those are owned by 17.6, 17.7 and 17.8.

## Inputs

17.5 opens only after:

- 17.0 Reference/Test Harness is closed;
- 17.1 Hulls/Mandatory Systems is closed;
- 17.2 Beam design/mount metadata is closed;
- 17.3 Ordnance design breadth is closed;
- 17.4 Specials/Fighter design breadth is closed;
- immutable built-Ship snapshots from normal construction remain lossless.

## Current blocker to replace

The current Tactical runtime still carries a Slice-15.5 fixture boundary: it admits only Frigate + Electronic Computer + Titanium Armor + no Shield + Standard Fuel and constructs Tactical durability/offense from old baseline constants.

That guard may not simply be deleted. 17.5 must first replace the protected hard-coded values with verified authoritative hull/component rules.

## Broad Tactical combatant contract

- Frigate, Destroyer, Cruiser, Battleship, Titan and Doom Star must enter Tactical with verified per-hull Structure/base durability semantics.
- Titanium, Tritanium, Zortrium, Neutronium, Adamantium and Xentronium must apply their verified Armor/Structure effects; no later Armor may merely carry an ID while using Titanium values.
- Supported Computers must supply verified Beam Offense / combat targeting inputs instead of always using the Electronic Computer baseline.
- Shield choices from no Shield through Class X must receive their verified Tactical damage-layer behavior. Facing, block, capacity and recharge semantics are frozen from direct evidence before implementation.
- Supported Drives must provide verified Tactical movement/combat-speed behavior for the accepted 17.1 matrix.
- Fuel Cell identity must not block Tactical admission when it has no Tactical effect; if direct evidence proves a combat effect, that behavior must be frozen and implemented.
- A legal ship must not be rejected merely because one installed later-family weapon lacks a Tactical consumer. The encounter can open, while that individual action remains explicitly unavailable until its owner slice closes.
- No accepted ship may enter Tactical with hidden Frigate/Electronic/Titanium substitutions.

## Shared Tactical authority foundation

17.5 freezes and implements the common substrate used by 17.6-17.8:

1. BattleSession remains the sole combat authority.
2. Weapon selection is slot/action based and server-projected; React never invents legal fire actions.
3. RNG consumption is server-authoritative, ordered and testable; the browser never rolls hit or damage.
4. Common damage-layer application (Shield / Armor / Structure and any directly evidenced shared layer behavior) is owned by server rules.
5. Combat events identify actor, source slot/action, target, sequence, result and damage/effect payload without requiring the UI to recalculate gameplay.
6. Tactical UI can present generic authoritative action availability for multiple weapon families without a Laser-only hard-coded path.
7. Presentation/animation is a consequence of authoritative events only. Animation completion never advances gameplay.
8. Weapon-family-specific hit formulas, projectile flight, fighter behavior and visual effects remain in their dedicated child slices.

The existing Laser vertical slice remains a regression baseline, not the architectural limit.

## Gate 1 - foundation audit

- [x] Re-audit the Slice-15.5 Tactical combatant guard and every hard-coded Frigate/Electronic/Titanium/no-Shield/Standard-Fuel value it protects.
- [x] Re-audit direct original evidence for all six hull durability baselines.
- [x] Re-audit Armor multipliers/effects, Computer combat targeting inputs, Shield behavior, drive combat speed and any Tactical relevance of Fuel Cells.
- [x] Audit the current BattleSession RNG, damage-layer, weapon-slot readiness, legal-fire projection and combat-event model.
- [x] Audit current Tactical UI assumptions that are Laser-specific versus already weapon-family-neutral.
- [x] Audit existing beam/movement animation event routing and identify the minimum generic presentation/event envelope needed by 17.6-17.8.
- [x] Produce representative early/mid/late combatant vectors for all six hull classes.
- [x] Define explicit unsupported-per-weapon behavior that does not reject an otherwise valid encounter.

## Gate 2 - foundation contract freeze

- [x] Freeze the accepted hull/Armor/Computer/Shield/Drive/Fuel Tactical matrix.
- [x] Freeze replacement semantics for the old baseline-only admission guard.
- [x] Freeze generic weapon-slot/action availability and unsupported-action projection.
- [x] Freeze authoritative RNG sequencing and shared damage-layer application boundaries.
- [x] Freeze the common combat-event envelope consumed by Tactical UI and replay/Observer projections.
- [x] Freeze generic presentation rules: events drive effects; effects never drive authority.
- [x] Freeze deterministic Combat Lab vectors proving broad admission and shared damage/event behavior.
- [x] Freeze compatibility boundaries that 17.6 Beam, 17.7 Ordnance and 17.8 Fighters/Specials must consume rather than fork.

## Gate 3 execution checkpoints

Execution plan: docs/research/SLICE_17_5_GATE3_IMPLEMENTATION_CHECKPOINT_PLAN_2026-09-29.md

Gate 3 is executed in five bounded checkpoints. After each checkpoint: focused QA, git diff --check, dedicated commit, clean-tree verification, durable ASH session checkpoint, user report, then stop until explicit continuation.

Current checkpoint: **1-2 complete; 3 - Shields and Shared Damage Foundation ready / not started**.

## Gate 3 - implementation

- [ ] Replace the baseline-only Tactical combatant path with the frozen hull/mandatory-system rules.
- [ ] Remove the exact Frigate/Electronic/Titanium/no-Shield/Standard-Fuel gate only after Tactical snapshots derive correct values.
- [ ] Keep Fuel Cells admission-neutral unless the frozen evidence contract assigns a Tactical effect.
- [ ] Generalize Tactical weapon/action projection so unsupported weapon families disable only their own actions rather than the entire encounter.
- [ ] Normalize shared RNG/result/damage event plumbing required by all later weapon-family slices.
- [ ] Extend Tactical UI selection/action rendering to consume generic server-projected weapon/action descriptors.
- [ ] Preserve the accepted Laser baseline and existing movement/selection/retreat behavior.
- [ ] Preserve save/load/replay/Observer behavior and deterministic event ordering.
- [ ] Add all-hull/mandatory-system Tactical admission and derived-stat regressions.

## Gate 4 - QA + close

- [ ] Every accepted 17.1 hull/mandatory-system representative enters Tactical with frozen durability/offense/defense/speed values.
- [ ] The old Slice-15.5 Frigate/Electronic/Titanium/no-shield/standard-fuel rejection no longer blocks otherwise legal richer designs.
- [ ] No hidden baseline substitution remains for Titan/Doom Star, later Armor, later Computers, Shields or Drives.
- [ ] A ship carrying a not-yet-implemented weapon family can still enter Tactical; only that specific action is explicitly unavailable.
- [ ] Existing Laser combat remains authoritative and regression-green.
- [ ] Tactical UI consumes generic action/event data without adding client-side hit/damage formulas.
- [ ] Generic event-driven presentation remains presentation-only.
- [ ] Full Go/vet/web/browser/persistence/replay/reference-game QA passes.

## Exit criterion

Slice 17.5 closes when the Tactical screen is a valid, truthful destination for the accepted Slice-17.1 ship matrix, shared combat authority/event/damage infrastructure is no longer tied to the old Frigate/Laser fixture, and later weapon-family slices can add their rules without reopening the combatant-admission foundation.

After 17.5 closes, Slice 17.6 adds complete supported Beam weapon and Beam-modifier Tactical behavior.
