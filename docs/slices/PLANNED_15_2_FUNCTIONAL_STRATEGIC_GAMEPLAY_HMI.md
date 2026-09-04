# Planned slice 15.2 - Functional strategic gameplay HMI

Status: **planned / queued; not open**.

Queue position: **15.2 of Slice-15 family**.

## Objective

Make the currently authoritative strategic game genuinely operable from the browser using server-projected state, legal actions, targets and costs, while visual styling remains intentionally functional rather than final.

## Dependencies

- Slice 15.1 UX/navigation shell.
- Slice 13 built-in AI.
- Slice 14 live save/resume baseline.

## Scope guard

In scope:

- supported Human-vs-built-in-AI New Game entry using current narrow settings;
- galaxy and star-system navigation/read state;
- Colony population/jobs/economy/construction interaction;
- Research selection/progress/completion interaction;
- Fleet inspection, movement, colonization and Outpost workflows;
- war/peace baseline interaction;
- turn submission/end-turn and bounded built-in-AI driving;
- server-side read/legal-action additions required by those workflows;
- player-safe projections only; no React-owned legality, costs, formulas or hidden AI state.

Defer:

- final visual identity/assets to 15.3;
- Tactical/Encounter/Invasion/persistence rich presentation to 15.4;
- final complete-game polish and E2E acceptance to 15.5;
- broader New Game/race options to Slice 16;
- broad Ship Designer to Slice 17.

## Gate 1 - Functional workflow audit

- [ ] Map every strategic command/decision in the canonical match to current App/HTTP surfaces.
- [ ] Identify missing player-safe reads, legal actions, target catalogs and command envelopes.
- [ ] Define minimum galaxy/system/colony/research/fleet/diplomacy data presentation.
- [ ] Define AI-turn driving boundaries and user-visible busy/ready states.
- [ ] Define functional browser regression strategy for each command family.
- [ ] Present Gate-2 authoritative HMI contract.

## Gate 2 - Functional authority freeze

- [ ] Freeze required server read/legal-action additions.
- [ ] Freeze browser command submission patterns and optimistic/non-optimistic behavior.
- [ ] Freeze strategic page/panel responsibilities inherited from 15.1.
- [ ] Freeze AI-turn driving and error recovery behavior.
- [ ] Freeze minimum functional acceptance scenarios.

## Gate 3 - Implementation

- [ ] Implement galaxy/system/fleet functional screens.
- [ ] Implement Colony jobs/economy/construction workflows.
- [ ] Implement Research workflows.
- [ ] Implement movement/colonization/Outpost and diplomacy workflows.
- [ ] Integrate end-turn and built-in-AI automation through normal Host authority.
- [ ] Add browser/server regressions for critical strategic commands.

## Gate 4 - QA + close

- [ ] Execute the canonical strategic Human-vs-AI workflow without direct API/dev-tool commands.
- [ ] Verify browser cannot bypass server legality/authority.
- [ ] Run full Go tests/vet/web build and `git diff --check`.
- [ ] Update evidence/status/HISTORY and close marker.
