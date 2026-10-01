# Planned Slice 17.7 - Tactical Missiles / Torpedoes / Bombs

Status: **open; Gate 1 ordnance Tactical audit active (2026-10-01)**.

## Objective

Turn the authoritative missile, torpedo and bomb design snapshots created by Slice 17.3 into supported server-authoritative combat behavior on the 17.5 Tactical foundation, integrating the Beam Point Defense contract frozen by 17.6 where projectile interception applies.

17.7 owns ordnance launch/readiness/ammunition, projectile lifecycle, targeting, interception, impact/damage resolution and event-driven ordnance presentation. Bomb behavior is implemented only in the original-evidenced combat context; Gate 1 must determine whether each bomb mechanic belongs to ship Tactical, planetary bombardment, or another explicit combat path rather than forcing bombs into the ship battlefield.

## Inputs

17.7 opens only after:

- 17.3 Missiles, Torpedoes and Bomb Ordnance is closed;
- 17.5 Tactical Combat Foundation is closed;
- 17.6 Tactical Beam Weapons & Beam Modifiers is closed, including the frozen Beam Point Defense interceptor contract;
- immutable ordnance snapshots remain lossless from Ship Designer through built Ship/BattleSession.

## Ownership boundary

17.7 owns:

- supported missile and torpedo launch/action semantics;
- ammunition/salvo/readiness consumption;
- authoritative projectile actors/state where the original rules require flight over time;
- projectile movement/ETA/target tracking/retargeting only where directly evidenced;
- missile/torpedo hit and damage resolution;
- Point Defense/interceptor interaction with Beam-side contracts from 17.6;
- ordnance impact, interception and explosion events/VFX;
- bomb combat resolution in its verified original context;
- Tactical UI weapon/action selection for supported ordnance.

It does not own fighter/bomber squadron simulation or unrelated ship-special effects; those are 17.8.

## Gate 1 - ordnance Tactical audit

- [ ] Re-audit every supported 17.3 missile/torpedo/bomb identity against direct original evidence.
- [ ] Re-audit ammo, salvo, range, travel, guidance, hit and damage semantics.
- [ ] Determine exact projectile lifecycle/state requirements rather than approximating flight in the browser.
- [ ] Re-audit Point Defense/interceptor interactions, target eligibility and ordering.
- [ ] Re-audit defensive countermeasures relevant to missiles/torpedoes where represented by current design snapshots.
- [ ] Determine each bomb's correct combat context and planetary-bombardment ownership.
- [ ] Audit current BattleSession/event model for persistent projectiles and replay/save implications.
- [ ] Audit Tactical UI requirements for projectile markers, flight paths, interception and impact feedback.
- [ ] Produce deterministic original/reference vectors for representative missile, torpedo and bomb cases.

## Gate 2 - ordnance combat contract freeze

- [ ] Freeze the supported missile/torpedo/bomb Tactical matrix.
- [ ] Freeze ammo/salvo/readiness rules.
- [ ] Freeze projectile spawn, movement, target and impact state/event contracts.
- [ ] Freeze exact hit/damage/interception ordering and RNG consumption.
- [ ] Freeze Point Defense integration with 17.6 Beam semantics.
- [ ] Freeze explicit unsupported behavior for any unproven guidance/countermeasure mechanic.
- [ ] Freeze bomb-vs-planet versus ship-Tactical ownership.
- [ ] Freeze event-driven projectile/interception/impact/explosion presentation semantics.
- [ ] Freeze deterministic Combat Lab/replay vectors.

## Gate 3 - implementation

- [ ] Implement supported ordnance Tactical consumers in server-authoritative code.
- [ ] Materialize authoritative projectile state where the frozen contract requires it.
- [ ] Implement ammo/salvo/readiness consumption.
- [ ] Implement projectile movement/targeting and deterministic event sequencing.
- [ ] Implement missile/torpedo hit and damage resolution.
- [ ] Implement Point Defense/interceptor resolution using the frozen 17.6 contract.
- [ ] Implement bomb combat behavior in the verified context.
- [ ] Extend participant-safe Tactical projections with required projectile/action state.
- [ ] Implement projectile/interception/impact/explosion presentation from authoritative events only.
- [ ] Preserve save/load/replay/Observer state and event ordering.
- [ ] Add deterministic projectile, interception, impact and exhaustion regressions.

## Gate 4 - QA + close

- [ ] Supported missile/torpedo weapons can be selected and launched from normal Tactical UI.
- [ ] Ammo/readiness/salvo state remains authoritative and visible.
- [ ] Projectile flight/state matches frozen reference vectors.
- [ ] Point Defense interception and non-interception outcomes match frozen ordering/RNG.
- [ ] Impacts apply correct Shield/Armor/Structure damage.
- [ ] Bomb behavior is available only in its verified combat context.
- [ ] Projectile/interception/explosion animations reproduce authoritative outcomes without client-side simulation.
- [ ] Save/load/replay during active projectile state remains deterministic where supported.
- [ ] Full Go/vet/web/browser/persistence/replay/reference-game QA passes.

## Exit criterion

Slice 17.7 closes when the supported ordnance matrix is no longer merely design metadata: missiles/torpedoes can launch, travel/interact/impact according to frozen server rules, Point Defense works end to end, verified bomb behavior resolves in the correct combat context, and Tactical presentation reflects authoritative projectile events.

After 17.7 closes, Slice 17.8 adds Fighter/Bomber/Interceptor combat and the remaining supported combat-special consumers.
