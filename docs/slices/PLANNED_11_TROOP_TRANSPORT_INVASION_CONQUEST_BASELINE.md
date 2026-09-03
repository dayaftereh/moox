# Planned slice 11 - Troop Transport, invasion and conquest baseline

Status: **closed; Gates 1-4 complete**.

Queue position: **11 of 12**.

This file is a prepared slice specification. It is not an `_OPEN_*.md` recovery marker and does not mean implementation has started.

## Objective

Implement the first complete territorial-war path: build/move a Troop Transport, invade an enemy Colony, resolve the minimum directly evidenced ground-combat fixture, transfer Colony ownership and connect conquered organic Population to the existing loyalty/assimilation state.

This slice is the first one that allows military conflict to change territorial ownership.

## Existing baseline

- construction, concrete Ships/Fleets, strategic transit and hostile encounter timing exist;
- Colony ownership, organic Population cohorts, origin/loyalty/assimilation state and multi-Colony Economy are already authoritative;
- Slice 10 is intended to provide explicit war/attack authorization;
- tactical ship combat exists only as a narrow ship-vs-ship fixture and must not silently absorb ground combat.

## Dependencies

- Slice 10 war/peace/attack authorization;
- existing Population cohort/assimilation model;
- existing Fleet movement/encounter/Colony state.

## Scope guard

In scope:

- concrete Troop Transport identity/design/construction semantics required by the original baseline;
- transport capacity/troop payload and strategic movement;
- invasion eligibility against an enemy Colony after required space-combat conditions are satisfied;
- one narrow directly evidenced ground-combat resolution fixture;
- deterministic attacker/defender troop losses and Colony capture outcome;
- Colony Empire ownership transfer;
- conquered Population loyalty/assimilation initialization using existing cohort fields;
- cleanup/consumption/retreat semantics for the transport after the selected fixture;
- post-conquest recalculation of blockade, Economy, Food/Freighter and legal actions.

Defer:

- planetary bombardment and bombs;
- bio-weapons;
- full marine/ground-tech/race bonus matrix unless required by selected fixture;
- colony destruction rather than capture;
- multiple simultaneous invasions;
- tactical ground-combat UI/animation;
- advanced occupation/rebellion systems;
- AI invasion planning.

## Gate 1 - Checkup + original analysis / reverse engineering

- [x] Re-check Colony ownership, Population cohorts/assimilation, special Ship construction and Fleet transit.
- [x] Create permanent invasion/conquest evidence document.
- [x] Identify original Transport/troop representation, build cost/capacity and movement participation.
- [x] Verify invasion timing relative to space combat, blockade and turn settlement.
- [x] Verify minimum ground-combat attacker/defender strength, RNG and casualty semantics for a canonical fixture.
- [x] Verify Colony ownership-transfer sequence and immediate post-capture state changes.
- [x] Verify treatment of native/foreign/conquered Population cohorts and initial assimilation/loyalty state.
- [x] Verify Transport consumption/survival after successful/failed invasion.
- [x] List every bombardment/ground-tech/race modifier deferred from the baseline.
- [x] Present exact Gate-2 state/session/command contract before implementation.

## Gate 2 - Implementation decision

- [x] Accept Transport/troop persistent state and construction/movement identity.
- [x] Accept invasion command/automatic boundary and authority model.
- [x] Accept first ground-combat fixture, RNG ownership and result vocabulary.
- [x] Accept Colony ownership-transfer and Population assimilation initialization rules.
- [x] Accept interaction with existing space BattleSession/encounter continuation.
- [x] Accept post-conquest recalculation order and event model.
- [x] Freeze deferred bombardment/advanced ground-combat families.

## Gate 3 - Implementation

- [x] Implement Transport/troop state, construction and strategic movement.
- [x] Implement invasion legality and selected deterministic ground-combat resolution.
- [x] Implement casualties, Colony ownership transfer and conquered Population state.
- [x] Recalculate strategic/economic derived state after conquest.
- [x] Expose command/projection through existing server/web transport.
- [x] Add deterministic success/failure, retry/atomicity, save/load and replay tests.
- [x] Add integration coverage from declared war + arrival through Colony capture.

## Gate 4 - Follow-up QA + commit + close

- [x] Run focused Transport/Fleet/space-encounter/invasion/Population/Economy regressions.
- [x] Verify exact ownership and cohort state after conquest round-trips through save/load.
- [x] Verify failed/invalid invasion is atomic.
- [x] Run full tests/vet/web integration checks and `git diff --check`.
- [x] Update evidence/status/HISTORY and close marker.

## Exit criterion

During an authorized war, a player can construct and move a real Troop Transport to an enemy Colony, resolve one directly evidenced deterministic invasion fixture, capture the Colony, apply troop losses and initialize conquered Population/assimilation state without bypassing existing strategic session authority. Bombardment and broad ground-combat systems remain deferred.
