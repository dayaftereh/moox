# Planned slice 09 - New Game and galaxy generation baseline

Status: **planned / queued; not open**.

Queue position: **9 of 12**.

This file is a prepared slice specification. It is not an `_OPEN_*.md` recovery marker and does not mean implementation has started.

## Objective

Replace the current test-fixture-only game start with the first authoritative deterministic `NewGame(seed, settings)` path that generates a playable minimal galaxy, Empires, Homeworlds and starting state.

The first target is deliberately narrow: a directly evidenced two-Empire/small-galaxy baseline that is sufficient to start real strategic turns through the server/web application boundary from Slice 08.

## Existing baseline

- deterministic Core RNG/IDs/save-load and substantial Empire/Colony/Fleet/Research/Construction state already exist;
- `core.NewSmallFixture` and related test fixtures currently stand in for real game creation;
- normalized star/planet/race/technology/ship data is already available in the repository;
- Slice 08 is intended to expose a real remote New Game/application boundary.

## Dependencies

- Slice 08 application/server transport baseline;
- existing Core/Economy/Research/Construction/colonization/Fleet foundations.

## Scope guard

In scope:

- authoritative NewGame settings/value object and deterministic seed ownership;
- one canonical small-galaxy size/topology baseline directly derived from original evidence where possible;
- deterministic star coordinates/system identities and basic planet inventory generation;
- two Empire start slots/Homeworlds with legal non-overlapping placement;
- starting Colonies, Population, Treasury/Research/known-tech and starting Fleet/Ship state required by the chosen original start level;
- deterministic fog/knowledge projection only to the minimum needed for a real player view;
- exact save/replay equality for identical seed/settings;
- server/web create-game flow using the Slice-08 transport without moving generation into the frontend.

Defer:

- all galaxy sizes/shapes/difficulties at once;
- complete race designer/custom races;
- monsters/Orion/Antarans;
- wormholes/black holes/special systems unless essential to chosen fixture;
- full random event placement;
- advanced map-generation options;
- AI player strategy.

## Gate 1 - Checkup + original analysis / reverse engineering

- [ ] Re-check current fixture/state constructors and normalized galaxy/planet/start data.
- [ ] Create permanent New Game/galaxy evidence document.
- [ ] Identify original MOO2 1.31 new-game entry points, RNG ownership and the smallest reproducible galaxy/start fixture.
- [ ] Verify supported galaxy size/star-count/coordinate constraints for the selected fixture.
- [ ] Verify star/system/planet generation ordering and random-stream consumption needed for deterministic reproduction.
- [ ] Verify Homeworld/start-player placement constraints and tie/spacing behavior.
- [ ] Verify starting Empire/Colony/Population/Treasury/Research/Fleet inventory for the selected technology level.
- [ ] Separate proven original behavior from deliberate MOOX modernization/settings simplification.
- [ ] Present exact Gate-2 state/settings/generation contract before implementation.

## Gate 2 - Implementation decision

- [ ] Accept NewGame settings schema/version and seed ownership.
- [ ] Accept deterministic generation phases/order and stable ID allocation policy.
- [ ] Accept first supported galaxy size/player-count/start-tech fixture.
- [ ] Accept Homeworld/start placement and minimum fog/knowledge model.
- [ ] Accept how generated starting fleets/designs reuse current Ship/Fleet state.
- [ ] Accept server API response/error contract for New Game creation.
- [ ] Freeze explicitly deferred galaxy/start families.

## Gate 3 - Implementation

- [ ] Implement deterministic NewGame generator and settings validation.
- [ ] Implement selected galaxy/star/planet/Homeworld generation baseline.
- [ ] Materialize two legal starting Empires/Colonies/Population/Research/Treasury/Fleets.
- [ ] Integrate with GameSession and Slice-08 create-game endpoint.
- [ ] Add golden-seed exact-state and different-seed divergence tests.
- [ ] Add save/load/replay/observer-isolation tests from generated games.
- [ ] Add browser HMI New Game proof using server-authoritative generation.

## Gate 4 - Follow-up QA + commit + close

- [ ] Re-run original/golden generation fixtures.
- [ ] Verify identical seed/settings produce byte/semantic-identical state and IDs.
- [ ] Verify generated games pass Core validation and can execute at least one real strategic turn.
- [ ] Run full tests/vet/web integration checks and `git diff --check`.
- [ ] Update evidence/status/HISTORY and close marker.

## Exit criterion

A user can request one supported New Game configuration through the authoritative application API and receive a deterministic, validated, saveable two-Empire galaxy with real starting strategic state. The same seed/settings reproduce the same authoritative game without relying on `NewSmallFixture` or frontend-generated state.
