# Planned slice 12 - Empire elimination and first headless victory loop

Status: **planned / queued; not open**.

Queue position: **12 of 12**.

This file is a prepared slice specification. It is not an `_OPEN_*.md` recovery marker and does not mean implementation has started.

## Objective

Connect the completed strategic subsystems into the first deterministic end-to-end match lifecycle: create a real game, execute strategic turns, expand, enter war, fight, conquer the opponent's final Colony and finish the GameSession with an authoritative elimination/conquest winner.

This is an integration milestone and a deliberately narrow first victory condition, not the complete Master of Orion II endgame.

## Existing baseline expected at slice start

- Slice 08: authoritative server/web transport;
- Slice 09: real deterministic New Game/galaxy generation;
- Slice 10: war/peace/attack authorization;
- Slice 11: Troop Transport/invasion/Colony conquest;
- existing Economy/Research/Construction/colonization/Fleet/space encounter/tactical vertical slices.

## Dependencies

- Slices 08-11.

## Scope guard

In scope:

- authoritative Empire elimination predicate for the selected conquest baseline;
- game-over/winner state on GameSession and revisioned player/observer projections;
- deterministic final-turn ordering when the last Colony/remaining recovery assets are removed or captured;
- command rejection/final read-only behavior after game completion;
- server/web game-over notification/projection;
- one scripted/headless integration scenario that begins with `NewGame` and ends with a validated winner using only public application/session commands;
- replay/save-load equality for the completed match;
- a milestone report enumerating which remaining systems are fidelity/depth work rather than blockers to a complete match lifecycle.

Defer:

- Galactic Council diplomatic victory;
- Orion/Antaran victory paths;
- score/time-limit victory;
- surrender unless Gate 1 proves it is required for elimination semantics;
- competitive AI capable of winning the scenario autonomously;
- polished UI/game-over cinematic;
- broad tactical/economy/diplomacy completeness beyond dependencies.

## Gate 1 - Checkup + original analysis / integration audit

- [x] Re-check completed Slice 08-11 contracts and current GameSession phase/state model.
- [x] Create permanent first-victory-loop evidence/integration document.
- [x] Verify original MOO2 Empire elimination/end-of-game trigger needed for the narrow conquest fixture.
- [x] Verify whether surviving Fleets/Outposts without Colonies delay elimination and under what conditions.
- [x] Verify final ownership/turn-settlement ordering relative to victory detection.
- [x] Verify minimum final result/winner/loser information needed for player/observer views and replay.
- [x] Define a deterministic public-command-only end-to-end scenario from New Game to final conquest.
- [x] Inventory remaining systems after this milestone and distinguish completeness/fidelity work from match-lifecycle blockers.
- [x] Present the exact Gate-2 game-over/session contract before implementation.

## Gate 2 - Implementation decision

- [x] Accept first supported victory/elimination predicate.
- [x] Accept GameSession completed/game-over state and final winner/result representation.
- [x] Accept when victory is evaluated in the strategic/invasion continuation order.
- [x] Accept post-game command/read/save/projection behavior.
- [x] Accept WebSocket/HTTP game-over notification/projection behavior.
- [x] Accept the deterministic end-to-end integration scenario and its fixture/settings.
- [x] Freeze Council/Orion/Antaran/score victory families.

## Gate 3 - Implementation

- [x] Implement authoritative elimination/victory state and validation.
- [x] Integrate victory detection after conquest/strategic settlement at the accepted boundary.
- [x] Prevent further gameplay mutation after completion while preserving observer/save/replay access.
- [x] Expose completed-game result through server/web transport.
- [x] Implement the public-command-only deterministic headless match scenario from New Game to winner.
- [x] Add save/load/replay and same-seed/same-command exact-result tests.
- [x] Record remaining post-milestone roadmap gaps.

## Gate 4 - Follow-up QA + commit + close

- [x] Re-run full end-to-end match scenario multiple times and compare exact result/event history.
- [x] Verify game-over timing and final Colony/Ship/Fleet state are deterministic.
- [x] Verify post-game invalid commands are atomic and do not advance revisions.
- [x] Verify server/browser projection and reconnect of a completed game.
- [x] Run full tests/vet/web checks and `git diff --check`.
- [ ] Update evidence/status/HISTORY and close marker.

## Exit criterion

MOOX can start one supported real deterministic game, play the required public strategic/application commands through expansion, war and conquest, eliminate the losing Empire and end the authoritative GameSession with a reproducible winner. This marks the first complete headless 4X match lifecycle; broader MOO2 fidelity and alternative victory systems remain follow-up work.