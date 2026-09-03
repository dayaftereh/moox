# Planned slice 15 - Playable browser strategic HMI loop

Status: **planned / queued; not open**.

Queue position: **15 of 17**.

## Objective

Turn the transport-proof React HMI into the first genuinely playable browser strategic loop for the supported Human-vs-built-in-AI match while keeping all gameplay authority in Go.

## Dependencies

- Slice 08 server/web authority baseline.
- Slice 13 built-in AI baseline.
- Slice 14 live save/resume baseline.

## Scope guard

In scope:

- supported New Game flow with Human vs built-in AI;
- usable strategic galaxy/system/fleet presentation;
- Colony population/jobs/economy/construction interaction;
- Research selection/progress interaction;
- Fleet movement/colonization/outpost interaction;
- war/peace baseline UI;
- Encounter/Battle/Invasion decision UI for currently supported commands;
- turn progression, live save/load/resume and completed-game result;
- reconnect/invalidation behavior through existing HTTP/WebSocket contract;
- server-projected legal choices/costs/targets; React must not own rule logic.

Defer:

- polished MOO2-style art/layout;
- full race designer/ship designer screens;
- broader New Game options from Slice 16;
- tactical depth beyond currently authoritative commands;
- native/Wails packaging.

## Gate 1 - HMI workflow audit

- [ ] Map every command/decision required by the canonical supported match to an existing App/Server surface.
- [ ] Identify missing read/legal-action queries that must be added server-side.
- [ ] Define minimum galaxy/colony/research/fleet/battle information architecture.
- [ ] Define browser save/load/reconnect lifecycle.
- [ ] Define an end-to-end browser regression/smoke strategy.
- [ ] Present Gate-2 interaction contract and screenshots/wireframe-level structure before implementation.

## Gate 2 - Implementation decision

- [ ] Freeze page/panel/navigation model.
- [ ] Freeze authoritative query/command additions.
- [ ] Freeze error/invalidation/reconnect behavior.
- [ ] Freeze minimum visual acceptance criteria for a playable match.

## Gate 3 - Implementation

- [ ] Implement strategic game navigation/presentation.
- [ ] Implement all supported human command workflows.
- [ ] Integrate AI turn driving and save/resume.
- [ ] Add browser-facing regressions for critical commands and completion.

## Gate 4 - Follow-up QA + commit + close

- [ ] Play a complete supported Human-vs-AI game through browser surfaces.
- [ ] Verify save/reload/reconnect and completion projection.
- [ ] Run full tests/vet/web checks and `git diff --check`.
- [ ] Update evidence/status/HISTORY and close marker.