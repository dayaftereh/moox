# Planned slice 06 - Strategic hostile encounters and BattleSession handoff

Status: **planned / queued; not open**.

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

- [ ] Re-check GameSession encounter/battle lifecycle, directed hostility and Fleet arrival timing.
- [ ] Create permanent encounter-timing evidence document.
- [ ] Verify original strategic encounter trigger timing relative to movement/arrival/blockade/colonization.
- [ ] Verify whether stationary hostile Fleets at turn start also trigger combat and under what conditions.
- [ ] Verify encounter grouping when more than two Fleets/Empires occupy one system.
- [ ] Verify whether civilian special ships participate, evade, are captured/destroyed or remain outside the tactical battle.
- [ ] Verify battle ownership/participant mapping and deterministic order if multiple systems generate encounters in one strategic turn.
- [ ] Verify what strategic resolution is paused until combat returns a result.
- [ ] Verify the minimum result semantics the strategic layer must consume after battle.
- [ ] Resolve Colony/planet-defense participation only to the extent required to build a valid first battle handoff.
- [ ] Present findings and proposed encounter producer; stop before implementation.

## Gate 2 - Implementation decision

- [ ] Decide authoritative `EncounterSpec` production inputs and ordering.
- [ ] Decide Fleet/ship composition snapshot vs reference semantics at battle creation.
- [ ] Decide strategic turn phase where encounters are generated.
- [ ] Decide how concurrent independent encounters are represented using existing parallel BattleSessions.
- [ ] Decide what state is frozen/pending while battles are active.
- [ ] Decide minimal battle result contract to re-enter strategic resolution safely.
- [ ] Decide Observer/replay event ordering and save/recovery expectations.

## Gate 3 - Implementation

- [ ] Derive encounters deterministically from hostile strategic state at the accepted timing boundary.
- [ ] Generate stable encounter ordering and BattleSession IDs/seeds.
- [ ] Include authoritative participant/Fleet/ship identity required by the next tactical slice.
- [ ] Transition GameSession into battle handling without leaking wall-clock completion order.
- [ ] Accept/commit minimal battle result(s) and resume the strategic phase deterministically.
- [ ] Keep tactical result production itself stubbed/external as allowed by the accepted scope.
- [ ] Add Observer/replay and recovery tests.
- [ ] Add multi-system/multi-battle stable-order regressions.

## Gate 4 - Follow-up QA + commit + close

- [ ] `gofmt` changed Go files.
- [ ] `go test ./... -count=1`.
- [ ] `go vet ./...`.
- [ ] `git diff --check`.
- [ ] Run focused strategic movement/hostility/blockade/encounter/BattleSession ordering regressions.
- [ ] Verify parallel battle completion cannot alter authoritative replay order.
- [ ] Update evidence/status/HISTORY and close marker.

## Exit criterion

When evidence-backed hostile Fleets meet, the strategic runtime automatically and deterministically creates one or more BattleSessions with stable participants/order/seeds, pauses/resumes correctly and records the handoff through Observer/replay without yet pretending tactical combat is implemented.