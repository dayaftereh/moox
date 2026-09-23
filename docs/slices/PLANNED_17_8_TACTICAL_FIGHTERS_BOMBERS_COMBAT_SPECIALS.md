# Planned Slice 17.8 - Tactical Fighters / Bombers / Combat Specials

Status: **planned; child of Slice 17; not open**.

## Objective

Turn the authoritative Fighter/Bomber/Interceptor bay and combat-special snapshots created by Slice 17.4 into supported server-authoritative Tactical behavior on top of the 17.5-17.7 combat foundation.

17.8 owns carrier/fighter combat actors and those ship-special effects whose authoritative consumer belongs in Tactical. Gate 1 must classify each supported 17.4 special precisely; no special is silently treated as working merely because it is legal in the Ship Designer.

## Inputs

17.8 opens only after:

- 17.4 Ship Specials and Fighter Systems is closed;
- 17.5 Tactical Combat Foundation is closed;
- 17.6 Beam Tactical combat is closed;
- 17.7 Ordnance Tactical combat is closed;
- immutable special/fighter snapshots remain lossless through built Ship/BattleSession.

## Ownership boundary

17.8 owns, where directly evidenced and frozen:

- Fighter/Bomber/Interceptor bay launch and availability;
- fighter-group or individual combat-actor representation;
- movement/engagement/attack/loss/recovery semantics;
- fighter interaction with Beam Point Defense, ordnance and defensive systems;
- Bomber attack behavior in its verified target context;
- carrier/bay readiness and remaining-craft state;
- participant-safe projections and Tactical controls for carrier actions;
- supported combat-special consumers that modify targeting, movement, defense, initiative, stealth, teleportation, scanning, damage or other Tactical behavior;
- fighter/bomber/special event-driven presentation and destruction feedback.

Strategic-only specials remain strategic. Any special whose original behavior is not directly evidenced remains explicitly unsupported/deferred rather than guessed.

## Gate 1 - fighter/special Tactical audit

- [ ] Re-audit Fighter/Bomber/Interceptor bay identities and capacity/readiness semantics.
- [ ] Re-audit fighter actor/group representation, launch, movement, targeting, attack and loss behavior.
- [ ] Re-audit Bomber target eligibility and damage context.
- [ ] Re-audit Point Defense/Beam/ordnance interactions with fighters where applicable.
- [ ] Classify every supported 17.4 special as strategic-only, Tactical-common, Beam-related, ordnance-related, fighter-related or unsupported.
- [ ] Re-audit exact Tactical effects and interaction ordering for every special accepted into this slice.
- [ ] Audit BattleSession/replay/save requirements for launched craft and temporary special states.
- [ ] Audit Tactical UI/action/presentation requirements for bays, fighter actors and specials.
- [ ] Produce deterministic original/reference vectors for representative carrier and special-effect cases.

## Gate 2 - fighter/special combat contract freeze

- [ ] Freeze the supported Fighter/Bomber/Interceptor Tactical matrix.
- [ ] Freeze launch/capacity/readiness/recovery/loss semantics.
- [ ] Freeze fighter movement/target/attack and defense interaction rules.
- [ ] Freeze Bomber combat-context rules.
- [ ] Freeze every accepted combat-special consumer, interaction order and RNG use.
- [ ] Freeze temporary Tactical state/event/replay schema required by fighters/specials.
- [ ] Freeze participant-safe UI/action projection contracts.
- [ ] Freeze event-driven fighter/special presentation semantics.
- [ ] Freeze explicit unsupported behavior for unproven specials.

## Gate 3 - implementation

- [ ] Implement authoritative fighter/bomber/interceptor actors or groups according to the frozen contract.
- [ ] Implement launch, movement, targeting, attack, losses and remaining-craft readiness.
- [ ] Implement Beam Point Defense/ordnance/defensive interactions with fighter actors where frozen.
- [ ] Implement verified Bomber behavior.
- [ ] Implement every accepted Tactical combat-special consumer.
- [ ] Extend Tactical action projection/UI for carrier and special actions without browser authority.
- [ ] Implement fighter/special presentation from authoritative events only.
- [ ] Preserve save/load/replay/Observer behavior for launched craft and temporary special state.
- [ ] Add deterministic carrier, fighter-loss and special-interaction regressions.

## Gate 4 - QA + close

- [ ] Supported fighter/bomber/interceptor systems can be used through normal Tactical UI.
- [ ] Launch/readiness/capacity/loss state matches frozen vectors.
- [ ] Fighter/Beam/Point-Defense/ordnance interactions match frozen ordering.
- [ ] Every accepted combat special produces its exact authoritative effect.
- [ ] Strategic-only or unsupported specials are not falsely presented as Tactical-working.
- [ ] Fighter/special animations correspond to authoritative actors/events and remain presentation-only.
- [ ] Save/load/replay remains deterministic with active fighter/special state.
- [ ] Desktop/mobile carrier/action controls remain usable.
- [ ] Full Go/vet/web/browser/persistence/replay/reference-game QA passes.

## Exit criterion

Slice 17.8 closes when supported Fighter/Bomber/Interceptor systems and every accepted combat-special consumer are real Tactical mechanics rather than passive design metadata, with server-authoritative actions/state/results and event-driven presentation.

After 17.8 closes, Slice 17.9 performs full Ship Designer + Tactical integration and acceptance for the Slice-17 parent program.
