# Planned slice 15.4 - Rich gameplay presentation / battle / persistence UX

Status: **planned / queued; not open**.

Queue position: **15.4 of Slice-15 family**.

## Objective

Turn the functional strategic HMI and visual identity into coherent high-value gameplay presentation for interactive resolution phases, save/load/reconnect and exceptional states.

## Dependencies

- Slice 15.1 UX/navigation/design-system shell.
- Slice 15.2 functional strategic HMI.
- Slice 15.3 visual identity/asset pipeline.
- Slice 14 live save/resume and invalidation semantics.

## Scope guard

In scope:

- supported Encounter/Tactical Battle presentation and command interaction;
- Invasion decision presentation;
- Colony-Base decision presentation;
- Research completion/selection presentation;
- diplomacy and major strategic event feedback where already authoritative;
- save/export, import/load and restore/reconnect user flows;
- `game_loaded` / snapshot invalidation / refetch behavior;
- loading, transient failure, stale-state, conflict and fatal-session presentation;
- result/victory transition presentation;
- animation/motion/audio hooks that do not affect simulation or command timing authority.

Defer:

- tactical mechanics beyond existing Go authority;
- cloud save slots/account sync;
- new diplomacy/research mechanics;
- full cinematics/native shell packaging.

## Gate 1 - Rich-state UX audit

- [ ] Inventory every non-Planning interactive phase and current DecisionView/action projection.
- [ ] Map Tactical/Encounter/Invasion/Colony-Base/Research workflows to accepted 15.1 navigation.
- [ ] Map all live persistence and WebSocket invalidation/reconnect paths to explicit UX states.
- [ ] Identify required player-safe read additions and animation-safe state boundaries.
- [ ] Define error/conflict/retry behavior without hiding authoritative rejection reasons.
- [ ] Present Gate-2 rich-interaction contract and mockups.

## Gate 2 - Rich presentation freeze

- [ ] Freeze battle/decision layouts and command interaction model.
- [ ] Freeze save/load/reconnect lifecycle and user-visible state machine.
- [ ] Freeze invalidation/refetch/error/conflict presentation.
- [ ] Freeze result/victory transition and major-event feedback.
- [ ] Freeze accepted motion/audio hooks without gameplay timing authority.

## Gate 3 - Implementation

- [ ] Implement supported Tactical/Encounter/Invasion/Colony-Base/Research rich workflows.
- [ ] Implement save/export/import/restore/reconnect UX.
- [ ] Implement snapshot invalidation/refetch/conflict/error presentation.
- [ ] Integrate 15.3 assets and accepted motion/audio hooks.
- [ ] Add browser/server regressions for critical decision and persistence paths.

## Gate 4 - QA + close

- [ ] Exercise every supported interactive phase through browser surfaces.
- [ ] Verify save/reload/reconnect and stale-client invalidation behavior end to end.
- [ ] Verify rejected commands/saves remain clear and non-destructive in UX.
- [ ] Run full tests/vet/web build and `git diff --check`; update evidence/status/HISTORY and close marker.
