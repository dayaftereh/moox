# Open slice 15.5 - Interactive 2D tactical combat
Status: **open; Gates 1-2 complete; Gate 3 implementation active**.
Queue position: **15.5 of Slice-15 family**.
## Objective
Turn the deliberately narrow Slice-07 fixed-position Laser BattleSession baseline into the first genuinely **interactive, turn-based 2D Tactical Combat** experience in the browser: the player selects an active ship, sees server-authoritative legal movement/target choices, moves on the tactical battlefield, fires supported weapons, ends activations/rounds and returns the deterministic battle result to the strategic game.
The UI should feel like a tactical game surface rather than a form. The exact original-valid coordinate/grid, movement-cost, initiative, range and facing semantics are Gate-1 research items and must not be invented in React.
## Dependencies
- Slice 07 tactical ship-combat baseline and Slice 06 strategic BattleSession handoff.
- Slice 14 active-game save/resume semantics.
- Slice 15.1 UX/navigation/i18n shell.
- Slice 15.2 functional strategic HMI and Encounter entry points.
- Slice 15.3 MOOX visual identity/ship/battle asset pipeline.
- Slice 15.4 rich decision/persistence UX and battle-entry/return transitions.
## Existing baseline entering this slice
Slice 07 already provides a deterministic authoritative Tactical BattleSession for one narrow fixture:
- BattleSession-local tactical state and deterministic RNG;
- initiative/activation order;
- fixed post-deployment X/Y positions;
- Armor/Structure damage and destruction;
- one supported standard Laser path;
- `battle.fire_beam` and `battle.end_activation` commands;
- terminal result reconciliation back into the strategic GameSession.
It explicitly does **not** yet provide player-controlled movement/turning, deployment, facing, broad shields/weapons, missiles/fighters/boarding/retreat or a general multi-ship tactical model.
## Product interaction target
Desktop and touch devices should expose the same authoritative battle semantics with different input ergonomics:
1. show the tactical battlefield as the primary surface;
2. select/indicate the currently active player ship;
3. display legal movement cells/points/positions projected by the server;
4. click/tap a legal destination to submit an authoritative movement command;
5. select a supported weapon/action;
6. display only player-safe legal targets projected by the server;
7. click/tap a target to submit the authoritative fire command;
8. show resulting damage/destruction and updated readiness/state;
9. end activation/advance initiative/round through authoritative commands;
10. when the battle resolves, transition back through the existing strategic encounter handoff.
Mobile target:
- battlefield consumes most of the viewport;
- touch pan/zoom where required by battlefield size;
- tap active ship / legal movement target / enemy target;
- ship/action details in a bottom sheet or compact overlay;
- no interaction may depend on hover.
Desktop target:
- battlefield remains primary;
- mouse selection and clear movement/target overlays;
- persistent ship/status/action panel where useful;
- keyboard shortcuts may be additive but never required.
## Authority boundary
React is presentation/orchestration only. The Go battle/session authority must own and project:
- active ship/seat and initiative order;
- legal movement destinations and movement costs;
- movement validation and committed tactical position;
- legal fire actions/weapon readiness;
- legal targets, range and attack legality;
- hit/damage/destruction resolution and RNG;
- round/activation progression;
- terminal battle result and strategic reconciliation.
The browser must never compute hidden movement legality, hit chance, damage, target legality or deterministic RNG outcomes independently.
## Scope guard
In scope:
- original-evidence-backed tactical coordinate/grid and movement baseline;
- authoritative move command plus player-safe legal-move projection;
- interactive 2D battlefield with movement/selection/target overlays and no normal visible arena edge;
- existing Laser fire and end-activation flows integrated into the battlefield UX;
- range/readiness/damage/destruction presentation for the accepted weapon baseline;
- read-only Tactical Scan/ship inspection for player-visible combatants;
- effectively unbounded Tactical coordinate/camera presentation without a normal fixed combat width/height;
- BattleSession -> strategic GameSession return flow;
- data/UI structures that do not assume the forever-final combat model is one ship versus one ship;
- Gate-1 decision on the minimum multi-ship breadth required for a credible first interactive Tactical slice;
- DE/EN battle UI through the shared 15.1 localization layer;
- representative desktop and phone/touch QA.
Gate 2 may pull in the minimum facing/turning/shield semantics only if original evidence proves they are inseparable from the accepted movement/fire baseline.
Defer unless explicitly accepted at Gate 2:
- broad deployment editor;
- broad multi-weapon/component/miniaturization mechanics;
- missiles/torpedoes/bombs/fighters/point defense;
- boarding/capture/stasis/self-destruct;
- retreat and advanced maneuver families;
- repair/regeneration/special systems;
- Colony/planet/station combat;
- leaders/crew/race tactical bonuses;
- monsters/Antarans;
- cinematics or timing-sensitive animation that influences gameplay authority.
## Gate 1 - Tactical mechanics / UX audit
- [x] Re-audit Slice-07 evidence and current BattleSession/Session/App/browser battle surfaces.
- [x] Research original MOO2 tactical coordinate/grid, movement, speed/cost, initiative, range and any facing/turning coupling required by the minimum interactive baseline.
- [x] Distinguish proven original behavior from deliberate MOOX UI modernization.
- [x] Define the minimum supported combatant/multi-ship breadth for the first credible interactive battle.
- [x] Identify required Core/Battle/Session player-safe reads, legal-action projections and command envelopes for movement and targeting.
- [x] Define desktop mouse and mobile touch battlefield interaction, including pan/zoom and no-hover requirements.
- [x] Define deterministic/browser regression strategy from strategic Encounter entry through battle result return.
- [x] Present the Gate-2 authoritative Tactical contract before implementation.
## Gate 2 - Interactive Tactical contract freeze
- [x] Freeze the accepted tactical coordinate/grid and movement-cost model.
- [x] Freeze BattleSession movement command semantics, sequencing and zero-mutation rejection behavior.
- [x] Freeze legal movement/fire/target projections and player-safe state shape.
- [x] Freeze minimum supported ship-count/battle-shape breadth and explicit unsupported boundaries.
- [x] Freeze battlefield selection/movement/weapon/target/end-activation interaction model for desktop and phone.
- [x] Freeze battle-entry/return, reconnect/stale-state and error behavior.
- [x] Freeze Tactical Scan/inspection visibility and read-only detail projection.
- [x] Freeze effectively unbounded battlefield/camera semantics with no normal fixed arena edge.
- [x] Freeze the minimum browser/server acceptance scenarios.
## Gate 3 - Implementation
- [ ] Implement authoritative tactical movement state/rules/commands and deterministic legal-action projection in Go.
- [ ] Integrate movement with initiative, range, existing Laser readiness/fire, damage/destruction and battle completion.
- [ ] Add player-safe BattleSession/App/server reads required by the browser without leaking hidden state.
- [ ] Implement the interactive 2D battlefield with active-ship, legal-move, weapon and legal-target overlays on an effectively unbounded coordinate plane.
- [ ] Implement Tactical Scan/ship inspection with participant-safe weapon/readiness and damage/status details.
- [ ] Implement responsive mouse/touch interaction and 15.1 DE/EN localization.
- [ ] Integrate 15.3 accepted ship/battle visuals with safe fallbacks.
- [ ] Add deterministic Go and browser regressions for movement, firing, rejection/rollback and strategic result handoff.
## Gate 4 - Independent Tactical QA + close
- [ ] Enter a supported Tactical battle through normal browser gameplay rather than direct dev/API injection.
- [ ] Move at least one player ship through server-projected legal movement targets.
- [ ] Fire a supported weapon at a server-projected legal target and observe authoritative damage/state refresh.
- [ ] Exercise activation/round progression through the browser until the battle resolves.
- [ ] Verify destroyed/surviving strategic ship identities reconcile correctly after returning to the strategic game.
- [ ] Verify rejected/stale/illegal moves and shots are non-destructive and understandable in the UI.
- [ ] Verify Tactical Scan can inspect friendly/enemy visible ships without consuming or mutating Tactical state.
- [ ] Verify representative desktop plus phone/touch behavior with no required hover, no catastrophic overflow and no visible artificial arena boundary.
- [ ] Run full Go tests/vet/web build plus `git diff --check`; update evidence/status/HISTORY and close marker.
Milestone on closure: **first interactive server-authoritative 2D Tactical Combat battle playable from the browser and reconciled back into the strategic match**.

## Gate-1 close marker (2026-09-09)

Gate 1 is **complete**. Evidence: docs/research/SLICE_15_5_GATE1_TACTICAL_MECHANICS_UX_AUDIT_2026-09-09.md. A proposed Gate-2 contract is available at docs/research/SLICE_15_5_GATE2_TACTICAL_CONTRACT_DRAFT_2026-09-09.md and remains **DRAFT / awaiting user acceptance**. No 15.5 Tactical implementation starts until Gate 2 is accepted/frozen.

## Gate-2 accepted freeze marker (2026-09-10)

Gate 2 is **complete and frozen** by explicit user acceptance. Frozen evidence: `docs/research/SLICE_15_5_GATE2_FREEZE_2026-09-10.md`. Gate 3 implementation is authorized within this boundary.
