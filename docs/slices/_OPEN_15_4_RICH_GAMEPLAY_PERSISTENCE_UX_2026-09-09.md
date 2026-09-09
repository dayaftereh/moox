# Open slice 15.4 - Rich gameplay decisions / persistence UX
Status: **open; Gates 1-2 complete; Gate 3 implementation active**.
Queue position: **15.4 of Slice-15 family**.
## Objective
Turn the functional strategic HMI and visual identity into coherent high-value presentation for non-tactical interactive resolution phases, Encounter/battle entry and return transitions, save/load/reconnect and exceptional states. The interactive 2D Tactical battlefield itself is deliberately separated into Slice 15.5.
## Dependencies
- Slice 15.1 UX/navigation/design-system shell.
- Slice 15.2 functional strategic HMI.
- Slice 15.3 visual identity/asset pipeline.
- Slice 14 live save/resume and invalidation semantics.
## Scope guard
In scope:
- Encounter detection, battle-entry/context presentation and post-battle return/summary transitions without implementing the interactive Tactical battlefield;
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
- **interactive Tactical movement/fire battlefield to Slice 15.5**;
- tactical mechanics beyond the Slice-07 baseline until 15.5 Gate 1/2 accepts their authoritative shape;
- cloud save slots/account sync;
- new diplomacy/research mechanics;
- full cinematics/native shell packaging.
## Gate 1 - Rich-state UX audit
- [x] Inventory every non-Planning interactive phase and current DecisionView/action projection.
- [x] Map Encounter handoff/Invasion/Colony-Base/Research workflows to accepted 15.1 navigation and define the explicit transition boundary into Slice-15.5 Tactical Combat.
- [x] Map all live persistence and WebSocket invalidation/reconnect paths to explicit UX states.
- [x] Identify required player-safe read additions and animation-safe state boundaries.
- [x] Define error/conflict/retry behavior without hiding authoritative rejection reasons.
- [x] Present Gate-2 rich-interaction contract and mockups.
## Gate-1 completion evidence (2026-09-09)

- Rich-state / phase / persistence audit: `docs/research/SLICE_15_4_GATE1_RICH_STATE_UX_AUDIT_2026-09-09.md`.
- Gate-2 interaction contract draft and textual mockups: `docs/research/SLICE_15_4_GATE2_RICH_INTERACTION_CONTRACT_DRAFT_2026-09-09.md`.
- Canonical review server remains persistence-enabled on port **7171** with exact restored `game-1`.
- Gate 2 is **not frozen** until explicit review/approval.

## Gate 2 - Rich presentation freeze
- [x] Freeze decision layouts plus Encounter -> Tactical entry and Tactical -> strategic return/summary transitions.
- [x] Freeze save/load/reconnect lifecycle and user-visible state machine.
- [x] Freeze invalidation/refetch/error/conflict presentation.
- [x] Freeze result/victory transition and major-event feedback.
- [x] Freeze accepted motion/audio hooks without gameplay timing authority.
## Gate-2 approval evidence (2026-09-09)

User approved the Gate-2 rich-interaction proposal without requested changes. The following are now frozen for Gate 3:

- blocking decision model for Encounter, Invasion and Colony Base;
- transition-summary model for Research breakthrough, Battle return, loaded game and Victory;
- stable Battle handoff route `#/game/<gameID>/battle/<battleID>`, with actual Tactical interaction owned by 15.5;
- local-file Save/Load contract using server export/import/atomic restore and explicit overwrite confirmation;
- stale last-known snapshot remains inspectable but read-only while reconnecting/refreshing;
- structured API errors and explicit lifecycle states;
- player-safe public identity and resolution-summary read additions only where the current server projection is insufficient;
- motion/audio remain presentation-only and never gate simulation or submit commands.

See `docs/research/SLICE_15_4_GATE2_RICH_INTERACTION_CONTRACT_DRAFT_2026-09-09.md`, now accepted as the Gate-2 contract.

## Gate 3 - Implementation
- [ ] Implement supported Encounter-entry/Invasion/Colony-Base/Research rich workflows, excluding the Slice-15.5 battlefield.
- [ ] Implement save/export/import/restore/reconnect UX.
- [ ] Implement snapshot invalidation/refetch/conflict/error presentation.
- [ ] Integrate 15.3 assets and accepted motion/audio hooks.
- [ ] Add browser/server regressions for critical decision, encounter-transition and persistence paths.
### Gate-3 implementation progress

- [x] Block 1 foundation: typed Colony Base/Battle contracts, structured `APIError`, explicit lifecycle state and synchronized-only direct mutation gate. Evidence: `docs/research/SLICE_15_4_GATE3_BLOCK1_CLIENT_AUTHORITY_LIFECYCLE_2026-09-09.md`.
- [x] Block 2 persistence UX: local Save, Home/In-Game Load, explicit Restore confirmation, new-ID Import and lifecycle-safe invalidation handling. Evidence: `docs/research/SLICE_15_4_GATE3_BLOCK2_SAVE_LOAD_RESTORE_UX_2026-09-09.md`.
- [x] Block 3 Colony Base: mandatory Post-Resolution legal-target/Scrap blocking decision, both real server outcomes, desktop + 320px QA. Evidence: `docs/research/SLICE_15_4_GATE3_BLOCK3_COLONY_BASE_UX_2026-09-09.md`.
- [x] Block 4 Invasion: player-safe public empire identity, contextual blocking decision and real Decline/Invade server outcomes. Evidence: `docs/research/SLICE_15_4_GATE3_BLOCK4_INVASION_UX_2026-09-09.md`.
- [x] Block 5 resolution summary / Research / result: stable player-safe event-derived transition summaries, automatic Research Breakthrough presentation and dedicated read-only Victory surface. Evidence: `docs/research/SLICE_15_4_GATE3_BLOCK5_RESOLUTION_SUMMARY_RESEARCH_RESULT_2026-09-09.md`.
- [x] Blocker remediation: browser New Game now sends explicit Human/Built-in-AI controllers; End Turn advances correctly and canonical `game-1` was recovered to Turn 2 with the submitted Research + construction choices applied. Evidence: `docs/research/SLICE_15_4_GATE3_ENDTURN_BUILTIN_AI_BLOCKER_2026-09-09.md`.

## Gate 4 - QA + close
- [ ] Exercise every supported non-tactical interactive phase through browser surfaces.
- [ ] Verify Encounter transitions preserve enough authoritative context for Slice-15.5 Tactical entry/return without React-owned battle rules.
- [ ] Verify save/reload/reconnect and stale-client invalidation behavior end to end.
- [ ] Verify rejected commands/saves remain clear and non-destructive in UX.
- [ ] Run full tests/vet/web build and `git diff --check`; update evidence/status/HISTORY and close marker.