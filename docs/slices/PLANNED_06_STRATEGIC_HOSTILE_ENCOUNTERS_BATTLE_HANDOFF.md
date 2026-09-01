# Planned slice 06 - Strategic hostile encounters and BattleSession handoff

Status: **closed; Gates 1-4 complete**.

Queue position: **6 of 7**.

## Objective

Turn hostile strategic Fleet co-location into an authoritative deterministic encounter producer that creates the already-supported `BattleSession` boundary automatically.

This slice ends at the battle handoff/lifecycle boundary; it does not yet implement the internal tactical combat model.

## Existing baseline

- Directed Empire relations already include the hostility needed for blockade derivation.
- Concrete military Fleet movement/composition is expected from slices 02/03.
- `GameSession` already supports encounter preparation, deterministic battle IDs/seeds, parallel `battle.Session` children, completion and stable replay order.
- Current `BattleSession` results are externally supplied; strategic gameplay does not yet automatically generate encounters from Fleets.

## Dependencies

- slice 03 concrete military ships;
- slice 04 generic combat Fleet movement and arrival;
- slice 05 is economically adjacent but not necessarily a hard gameplay dependency.

## Scope guard

In scope:

- detection of directly evidenced hostile strategic co-location;
- deterministic encounter grouping/order;
- BattleSession creation and authoritative GameSession phase transition;
- participant/composition snapshot/reference contract;
- stable completion handoff boundary back to strategic resolution.

Defer:

- tactical movement/fire/damage (slice 07);
- AI combat decisions;
- planetary bombardment/invasion;
- retreat results beyond the minimum result vocabulary required for the handoff;
- complex multi-party diplomacy unless directly required to group an encounter.

## Gate 1 - Checkup + original analysis / reverse engineering

- [x] Re-check GameSession encounter/battle lifecycle, directed hostility and Fleet arrival timing.
- [x] Create permanent encounter-timing evidence document.
- [x] Verify original strategic encounter trigger timing relative to movement/arrival/blockade/colonization.
- [x] Verify whether stationary hostile Fleets at turn start also trigger combat and under what conditions.
- [x] Verify encounter grouping when more than two Fleets/Empires occupy one system.
- [x] Verify whether civilian special ships participate, evade, are captured/destroyed or remain outside the tactical battle.
- [x] Verify battle ownership/participant mapping and deterministic order if multiple systems generate encounters in one strategic turn.
- [x] Verify what strategic resolution is paused until combat returns a result.
- [x] Verify the minimum result semantics the strategic layer must consume after battle.
- [x] Resolve Colony/planet-defense participation only to the extent required to build a valid first battle handoff.
- [x] Present findings and proposed encounter producer; stop before implementation.

## Gate 2 - Implementation decision

- [x] Decide authoritative `EncounterSpec` production inputs and ordering.
- [x] Decide Fleet/ship composition snapshot vs reference semantics at battle creation.
- [x] Decide strategic turn phase where encounters are generated.
- [x] Decide how concurrent independent encounters are represented using existing parallel BattleSessions.
- [x] Decide what state is frozen/pending while battles are active.
- [x] Decide minimal battle result contract to re-enter strategic resolution safely.
- [x] Decide Observer/replay event ordering and save/recovery expectations.

## Gate 3 - Implementation

- [x] Derive encounters deterministically from hostile strategic state at the accepted timing boundary.
- [x] Generate stable encounter ordering and BattleSession IDs/seeds.
- [x] Include authoritative participant/Fleet/ship identity required by the next tactical slice.
- [x] Transition GameSession into battle handling without leaking wall-clock completion order.
- [x] Accept/commit minimal battle result(s) and resume the strategic phase deterministically.
- [x] Keep tactical result production itself stubbed/external as allowed by the accepted scope.
- [x] Add Observer/replay and recovery tests.
- [x] Add multi-system/multi-battle stable-order regressions.

## Gate 4 - Follow-up QA + commit + close

- [x] `gofmt` changed Go files.
- [x] `go test ./... -count=1`.
- [x] `go vet ./...`.
- [x] `git diff --check`.
- [x] Run focused strategic movement/hostility/blockade/encounter/BattleSession ordering regressions.
- [x] Verify parallel battle completion cannot alter authoritative replay order.
- [x] Update evidence/status/HISTORY and close marker.

## Exit criterion

When evidence-backed hostile Fleets meet, the strategic runtime automatically and deterministically creates one or more BattleSessions with stable participants/order/seeds, pauses/resumes correctly and records the handoff through Observer/replay without yet pretending tactical combat is implemented.
