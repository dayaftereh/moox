# Empire elimination / first headless victory loop - Gate 1 evidence

Date: 2026-09-03

Status: **Gate 1 complete; Gate 2 contract awaiting approval. No victory/game-over implementation has started.**

## Scope

Establish the first deterministic conquest-victory lifecycle from real New Game through public strategic/session commands to an authoritative completed GameSession. Alternative victory families remain deferred.

## Current MOOX baseline

- Slices 08-11 are closed and the repository entered Slice 12 at clean HEAD `e2979c1`.
- `GameSession` currently has `planning`, `strategic_resolution`, `encounters`, `invasion_decisions`, and `post_resolution`; there is no strategic `completed` phase and no authoritative winner/result field.
- `CompleteTurn()` currently always advances the turn, clears staged battle/invasion state and returns to `planning`.
- Player/observer projections expose phase/revision/state but no game result.
- Existing invasion continuation commits conquest before returning through encounter/post-resolution continuation, which is the natural victory-evaluation boundary.
- Fresh baseline `go test ./...` passed on 2026-09-03.

## Original MOO2 1.31 evidence

Primary source: private owned `reference/original/support/files/Orion2.exe` with embedded debug symbols. Read-only inspection used `build/research-disasm/main.go`; temporary xref/symbol helpers were research-only and removed after use.

### Elimination predicate

`Check_For_Eliminated_Players_` is at VA `0xE4EB3`.

The routine scans all Colony records (stride `0x169`) and marks the owner of each record only when:

- owner byte is a normal player `< 8`; and
- `colony+0x06 == 0`.

Prior direct MOOX evidence proves `colony+0x06 = 1` is the Outpost flag. Therefore **only real Colonies keep an Empire alive; Outposts do not**.

For every living player record not represented by at least one normal Colony, the routine calls:

- `Eliminate_Player_` at `0xE45FF`; then
- `Event_Race_Eliminated_` at `0x233AB`.

`Eliminate_Player_` includes Ship cleanup (`Kill_Ship_` calls) and related diplomatic/spy cleanup. Therefore surviving Fleets/Ships do **not** postpone elimination; they are cleaned up as a consequence of elimination. Outposts also do not postpone elimination.

Evidence level: **direct original code + previously proven Outpost flag mapping**.

### Ordering relative to combat/invasion

The only direct call xref to `Check_For_Eliminated_Players_` found in object 1 is at `0xEA269`, inside `Search_For_Battles_` (`0xE9D62`). The nearby sequence performs combat, retreat processing and system-discovery work before the elimination check.

`Invade_` (`0xED48B`) performs ground combat and on attacker victory calls `Change_Colony_Ownership_` before resolving invasion troops and ship cleanup. Thus the final Colony ownership change is committed before the later elimination scan.

This establishes the narrow original ordering required by Slice 12:

1. resolve space-combat/invasion continuation;
2. commit final Colony ownership and invasion cleanup;
3. evaluate Empire elimination;
4. then evaluate whether one living player remains and finish the game.

Evidence level: **direct original code**.

### Winner detection

`Next_Turn_` calls `Next_Turn_Calc_`, then `N_Living_Players_`. The code compares the living-player count to `1`; when the result is `<= 1`, it calls `Get_Winner_` (`0x139E8`) and enters the normal win/end-game path.

`Get_Winner_` selects the surviving non-eliminated player (with the separate Council winner override when present). Council victory is outside this Slice 12 conquest baseline.

For the first MOOX conquest fixture, the authoritative winner can therefore be defined as the sole non-eliminated Empire after the post-conquest elimination pass.

Evidence level: **direct original code**.

## Gate 1 conclusions

1. **Empire elimination predicate:** an Empire is eliminated when it owns zero normal Colonies. Outposts do not count.
2. **Surviving assets:** Fleets/Ships and Outposts do not delay elimination. Elimination cleanup removes remaining Ships/assets as applicable.
3. **Victory timing:** evaluate only after the final invasion/encounter continuation has committed Colony ownership and cleanup, before advancing into another planning turn.
4. **Victory predicate for Slice 12:** after applying elimination, if exactly one non-eliminated participating Empire remains, that Empire wins by conquest.
5. **No implicit draw semantics are introduced in Slice 12.** The supported deterministic fixture must produce exactly one surviving Empire.

## Proposed authoritative MOOX result model for Gate 2

Add strategic `PhaseCompleted = "completed"` and an immutable session result concept with at minimum:

- result kind: `conquest`;
- winner Empire ID;
- winner Seat ID;
- eliminated/losing Empire IDs (stable sorted order);
- completed turn;
- completed revision.

The result is projected identically in Status/PlayerView/ObserverView and persisted/replayed with the session/event history. A deterministic `game_completed` session event records the same result payload.

No Council/Orion/Antaran/score result kinds become executable in this slice; their enum space may remain reserved/documented only.

## Proposed completion boundary for Gate 2

After the invasion/encounter continuation reaches the point that would otherwise enter `post_resolution`, run one authoritative finalization step:

1. derive normal-Colony ownership per Empire;
2. mark zero-Colony Empires eliminated and perform deterministic cleanup of their remaining strategic assets required by current Core invariants;
3. recompute/validate resulting state as needed;
4. if exactly one participating Empire remains, increment the session revision once for the committed completion boundary, store immutable result, append elimination/result events in stable order and set phase `completed`;
5. do **not** call `AdvanceTurn()` and do not return to `planning`.

If more than one Empire remains, continue existing `post_resolution -> CompleteTurn -> planning` behavior unchanged.

## Proposed post-game behavior for Gate 2

After `completed`:

- all gameplay mutation commands (`SubmitTurn`, immediate diplomacy/invasion/base commands, battle commands, `CompleteTurn`, telemetry that mutates session workflow) reject atomically without revision/event changes;
- read-only Status, PlayerView, ObserverView, event history, save/export and replay/reconnect remain available;
- server HTTP/WebSocket projections expose the completed phase/result and reconnect returns the same immutable result;
- no automatic new turn is created.

## Deterministic public-command-only headless scenario

Fixture baseline:

- ruleset `moo2-1.31`;
- seed `0x8009`;
- Small galaxy, Normal age, Average technology;
- two explicit human-controlled seats: Human and Darlok;
- real `NewGame` generation followed by a real `GameSession`; no post-NewGame direct state mutation.

Scenario driver uses only public application/session command/query boundaries and deterministic legal-choice selection:

1. create the seeded game and verify both starting Empires/Colonies;
2. execute ordinary strategic turns and construction choices needed for the attacking Empire to expand using existing colonization path;
3. construct the required military/transport assets through normal construction commands;
4. move Fleets through normal strategic movement commands;
5. declare war through the Slice-10 immediate diplomacy command;
6. resolve any required hostile encounter/tactical public command path;
7. move the Troop Transport to the defender's final Colony;
8. accept/resolve the Slice-11 invasion opportunity through its public immediate command;
9. verify final Colony ownership commits, defender is eliminated despite any remaining Fleet/Outpost, cleanup occurs, and session enters `completed` with the attacker as sole conquest winner;
10. replay the exact seed + public command/event sequence and require byte/semantic equality of final result/state/event history; save/load/reconnect must preserve the same completed result.

The implementation test may use deterministic helper code to *choose* among public legal actions, but it must never directly mutate `GameState` after `NewGame`.

## Remaining systems after this milestone

Not blockers to a complete match lifecycle once Slice 12 closes:

- Galactic Council victory;
- Orion/Antaran victory paths;
- score/time-limit victory;
- surrender fidelity;
- competitive/autonomous AI;
- broader tactical weapon/combat depth;
- full Economy/pollution/race/customization depth;
- leaders/espionage completeness;
- polished victory UI/cinematics/audio;
- broader MOO2 event/endgame fidelity.

These are fidelity/depth or alternative-victory work, not blockers to the first complete deterministic headless 4X lifecycle.

## Exact Gate 2 decision requested

Approve/freeze the following contract before implementation:

- zero **normal Colonies** => eliminated; Outposts/Fleets do not save the Empire;
- elimination evaluated after final conquest/continuation commit and before turn advance;
- one surviving participating Empire => immutable `conquest` result;
- `GameSession` enters `completed`, does not advance another turn;
- result contains winner Seat/Empire, sorted losers, completion turn/revision;
- post-game gameplay writes reject atomically while reads/save/replay/reconnect remain legal;
- server/web project the same completed result;
- canonical end-to-end fixture is seed `0x8009`, two human seats, Small/Normal/Average, public-command-only after NewGame;
- Council/Orion/Antaran/score/surrender/competitive-AI remain frozen out of scope.
## Gate 2 approval - 2026-09-03

User approval received. The exact Gate-2 contract above is frozen for Gate 3 implementation without scope expansion.


## Gate 3 implementation result - 2026-09-03

Gate 3 implements the frozen contract without adding alternative victory families:

- `session.PhaseCompleted` is an authoritative terminal strategic phase.
- Immutable `session.Result` stores `conquest`, winner Empire/Seat, stable eliminated Empire IDs, completion Turn and completion Revision.
- A real `empire.colony_conquered` event arms elimination; evaluation is delayed until all staged invasion/encounter continuation is fully settled.
- Zero-normal-Colony Empires are eliminated deterministically. Their Outposts, Ships, Strategic Fleets, Population transfers and diplomacy/peace-offer edges are removed; Empire records remain as historical/race identities with Capital/Freighters cleared.
- Eliminated Seats remain readable/reconnectable but cannot submit gameplay turns or immediate mutation commands and do not block surviving Seats in games that continue after an intermediate elimination.
- Final conquest increments one authoritative completion revision, emits stable `empire_eliminated`, `game_completed` and `phase_changed(completed)` events, and does not advance another turn.
- Status, player, observer and hosted-game summary projections expose the same immutable result.
- Host/WebSocket mutation notifications use `reason="game_completed"`; reconnect over HTTP returns the identical completed projection.
- The React HMI renders an explicit conquest-completed panel with winner, completion boundary and eliminated Empires.
- Completed GameSession snapshots have a versioned strict JSON format with exact byte round-trip. Completed sessions reject post-game writes atomically while preserving read/snapshot access.

### Canonical public-command-only match

`TestCanonicalNewGamePublicCommandHeadlessConquestIsExactReplay` begins from real `rules.NewGame(0x8009, Small/Normal/Average, Human vs Darlok)` and performs no direct `GameState` mutation after New Game.

The pinned seed revealed a real reachability constraint: the Human home at system 4 has no planet-bearing supply anchor inside the initial 4/6/9 pc fuel ranges; the first usable system is system 10 at 12 pc. The scenario therefore plays the ordinary Research frontier through Deuterium -> the required intermediate fields -> Iridium -> Urridium Fuel Cells (12 pc) for both seats. It then:

1. moves Darlok's combat and Colony-Ship fleets to system 18 so surviving assets remain as an elimination proof;
2. moves the Human starting Colony Ship to system 10 and colonizes real planet 30;
3. builds, moves and deploys real Outpost Ships through systems 11 -> 12 -> 13 -> 17;
4. builds a real Troop Transport at the Human capital;
5. declares war through the public immediate diplomacy command;
6. moves the original Human combat fleet plus Transport to Darlok system 23;
7. resolves the projected invasion through the public invasion command;
8. verifies final Colony capture, Darlok elimination despite surviving off-system Fleets, cleanup and immutable Human conquest result;
9. serializes/restores the completed match exactly.

The test runs the complete scenario twice from the same seed and compares the full completed-session snapshot bytes, which include final state, result and authoritative event history. The second run is therefore also an exact replay/determinism regression rather than only a winner assertion.

### Long-run lifecycle blocker found and fixed

The first real match attempt exposed a pre-existing Population Cohort growth bug at turn 92: aggregate projected growth could exceed the tiny remaining Colony capacity near the cap, even though legacy flat-population growth already clamped at capacity. `population_cohort_dynamics.go` now proportionally clamps per-origin projected cohort growth to the aggregate remaining capacity before starvation/food projection. Focused Population/Cohort regressions and the full long-run match both pass after the fix.

## Gate 4 QA - 2026-09-03

Independent follow-up execution after the final lifecycle hardening passed:

- canonical NewGame -> Research -> Colony -> Outposts -> war -> Transport -> invasion -> completed conquest runs twice byte-identically inside the deterministic test;
- final Colony ownership, eliminated off-system Fleets and completion Turn/Revision are asserted;
- completed snapshot marshal -> unmarshal -> marshal is byte-identical;
- rejected completed-session gameplay writes leave revision/status/event history unchanged;
- intermediate eliminated Seats do not block active turn readiness and retain read access;
- WebSocket completion notification + HTTP reconnect projection is verified;
- `go test ./... -count=1` PASS (all packages; `internal/session` includes the ~42 s complete-match regression);
- `go vet ./...` PASS;
- `npm run build` in `web/` PASS;
- `git diff --check` PASS (only Git line-ending conversion warnings, no whitespace errors).

Gate 4 functional QA is complete. Remaining closure work is documentation/HISTORY synchronization, implementation/evidence commit, OPEN-marker removal and final clean-repository verification. No push is part of this closure.
