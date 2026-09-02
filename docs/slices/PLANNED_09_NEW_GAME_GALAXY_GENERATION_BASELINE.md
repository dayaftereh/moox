# Planned slice 09 - New Game and galaxy generation baseline

Status: **active; Gate 4 QA complete / implementation commit + close pending**.

Queue position: **9 of 12**.

This specification is active under `docs/slices/_OPEN_NEW_GAME_GALAXY_GENERATION_BASELINE_2026-09-02.md`. Gates 1-3 are complete: the frozen New Game contract is implemented and regression-green. Gate 4 final QA, commit and close are pending.

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

- [x] Re-check current fixture/state constructors and normalized galaxy/planet/start data.
- [x] Create permanent New Game/galaxy evidence document.
- [x] Identify original MOO2 1.31 new-game entry points, RNG ownership and the smallest reproducible galaxy/start fixture.
- [x] Verify supported galaxy size/star-count/coordinate constraints for the selected fixture.
- [x] Verify star/system/planet generation ordering and random-stream consumption needed for deterministic reproduction.
- [x] Verify Homeworld/start-player placement constraints and tie/spacing behavior.
- [x] Verify starting Empire/Colony/Population/Treasury/Research/Fleet inventory for the selected technology level.
- [x] Separate proven original behavior from deliberate MOOX modernization/settings simplification.
- [x] Present exact Gate-2 state/settings/generation contract before implementation.

## Gate 2 - Implementation decision

- [x] Accept NewGame settings schema/version and seed ownership.
- [x] Accept deterministic generation phases/order and stable ID allocation policy.
- [x] Accept first supported galaxy size/player-count/start-tech fixture.
- [x] Accept Homeworld/start placement and minimum fog/knowledge model.
- [x] Accept how generated starting fleets/designs reuse current Ship/Fleet state.
- [x] Accept server API response/error contract for New Game creation.
- [x] Freeze explicitly deferred galaxy/start families.

### Gate-2 accepted contract - 2026-09-02

- Rules/data: add strict `data/rulesets/moo2-1.31/new_game_galaxy.json` schema **1** with original-derived tables/provenance plus named MOOX algorithms. Economy schema remains 8, Core remains 21 and planet-classes schema is unchanged.
- Seed/RNG ownership: `EconomyRules.NewGame(seed uint64, settings)` creates `core.NewGameState(seed)`, owns one shared `state.RNG()` across all generation phases, passes that RNG into existing New Game technology initialization where required, and calls `CommitRNG` exactly once after final validation. The HTTP seed is a decimal or `0x` string; internal state remains `uint64`.
- Determinism: freeze the Gate-1 generation phase order and ID allocation order. No external Host/session registration occurs until the candidate `GameState` validates.
- First production fixture: Small / Normal / Average / Tactical only; exactly two strictly ascending non-zero seats; exactly one Human and one Darlok; non-empty unique Empire names. Random events, Antarans, custom races and AI are not accepted silently.
- Homeworld/materialization: 20 systems, `moox_grid_jitter_v1`, original-derived planet tables/order, transient spectral/body state, only body type 3 materialized, `farthest_pair_v1`, `uniform_unused_orbit_v1`, Home systems >=3 materialized planets and Homeworld Medium/Abundant/Normal-G/Terran.
- Knowledge boundary: Slice 09 adds **no new Core fog/knowledge schema**. Player projection remains the current authoritative own-Empire/own-Colony view plus public Seat metadata; full state remains observer-only when that server capability is explicitly enabled. Galaxy-map discovery/visibility is deferred rather than faked.
- Starting assets per Empire: Population 8 with MOOX jobs 4/2/2, Treasury 50, Freighters 0, exact Tactical Average technology set, no starting Buildings, one accepted unarmed Frigate/Nuclear/Electronic/Titanium/Standard-Fuel `Scout` design, two concrete Scouts in one combat fleet, and one existing special Colony Ship fleet at the Home system.
- Session seats created from New Game are `session.ControllerLocalHuman` in this slice. AI/controller selection is deferred.
- HTTP creation: `POST /api/v1/games`, request schema 1, same-origin mutation guard, `application/json`, existing 1 MiB body limit and strict unknown-field rejection. Success is **201 Created** with `{schema_version, game, players}` where `game` is the normal `app.GameSummary` and `players` are generated Seat/Empire/Race/name identities.
- HTTP errors: malformed JSON/content type/schema/seed/Game ID/settings => **400 `bad_request`**; duplicate Game ID => atomic **409 `game_exists`** with no replacement or partial registration; unexpected generator/invariant failures => **500 `internal_error`**. Add a dedicated app-level duplicate-game sentinel instead of leaking the current generic `Register` error.
- Startup: normal `cmd/moox-server` starts an empty production Host with New Game capability, not `core.NewSmallFixture`. Any legacy fixture bootstrap must be behind an explicit development-only option and never be the default.
- Deferrals are frozen exactly as listed in Gate-1 evidence: larger sizes, other ages/tech starts, Strategic Combat, custom races/trait permutations, original coordinate/Homeworld/orbit parity, persistent non-colonizable bodies, exact packed jobs/free buildings/Scout auto-design, Orion/monsters/wormholes/nebulae/specials/splinters/heroes, Random Events/Antarans, AI and broader fog/knowledge.
## Gate 3 - Implementation

- [x] Implement deterministic NewGame generator and settings validation.
- [x] Implement selected galaxy/star/planet/Homeworld generation baseline.
- [x] Materialize two legal starting Empires/Colonies/Population/Research/Treasury/Fleets.
- [x] Integrate with GameSession and Slice-08 create-game endpoint.
- [x] Add golden-seed exact-state and different-seed divergence tests.
- [x] Add save/load/replay/observer-isolation tests from generated games.
- [x] Add browser HMI New Game proof using server-authoritative generation.

## Gate 4 - Follow-up QA + commit + close

- [x] Re-run original/golden generation fixtures.
- [x] Verify identical seed/settings produce byte/semantic-identical state and IDs.
- [x] Verify generated games pass Core validation and can execute at least one real strategic turn.
- [x] Run full tests/vet/web integration checks and `git diff --check`.
- [ ] Update evidence/status/HISTORY and close marker.

## Exit criterion

A user can request one supported New Game configuration through the authoritative application API and receive a deterministic, validated, saveable two-Empire galaxy with real starting strategic state. The same seed/settings reproduce the same authoritative game without relying on `NewSmallFixture` or frontend-generated state.
