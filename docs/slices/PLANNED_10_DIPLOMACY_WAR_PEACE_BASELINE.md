# Planned slice 10 - Diplomacy, war and peace baseline

Status: **closed; Gates 1-4 complete**.

Queue position: **10 of 12**.

This file is a prepared slice specification. It is not an `_OPEN_*.md` recovery marker and does not mean implementation has started.

## Objective

Turn the current directed `neutral` / `hostile` stance input into the first authoritative diplomacy lifecycle that players can intentionally change through commands and that governs strategic attack authorization.

The first slice should support a directly evidenced minimal war/peace loop, not the complete MOO2 negotiation system.

## Existing baseline

- directed Empire relations already exist and drive blockade/hostility checks;
- Slice 06 encounters consume hostility deterministically;
- Seat authority, commands, events and revisioned sessions are implemented;
- generated real players/Empires are expected from Slice 09.

## Dependencies

- Slice 09 real New Game/two-Empire start;
- Slice 06 strategic encounter authorization;
- Slice 08 transport for real player command flow.

## Scope guard

In scope:

- persisted authoritative bilateral/directed diplomatic state sufficient for war/peace;
- declare-war command and event semantics;
- directly evidenced peace proposal/acceptance or equivalent minimum bilateral transition;
- attack authorization relationship between diplomatic state and strategic encounter generation;
- deterministic relation-history/revision data only if required for legality/replay;
- player/observer projections and web-command surfaces;
- exact behavior when a relation changes while Fleets already share or approach a System.

Defer:

- trade/research treaties;
- tribute/gifts/demands/threats;
- alliances/non-aggression pacts unless required by the minimal fixture;
- diplomatic AI/personality;
- diplomacy screen polish;
- spies/espionage;
- Galactic Council diplomacy;
- race/leader diplomacy bonuses.

## Gate 1 - Checkup + original analysis / reverse engineering

- [x] Re-check current directed relation/hostility/blockade/encounter implementation.
- [x] Create permanent diplomacy baseline evidence document.
- [x] Verify original MOO2 relation states and minimum war/peace transition path.
- [x] Verify declare-war timing relative to same-turn movement/encounter processing.
- [x] Verify peace timing and whether already-created battles/attacks are cancelled or remain committed.
- [x] Verify directed vs symmetric state updates for the selected transitions.
- [x] Verify minimum messages/events/history needed for deterministic replay and UI projection.
- [x] Identify original attack-authorization/sneak-attack distinction and decide what is included now.
- [x] Present exact Gate-2 relation/command/state contract before implementation.

## Gate 2 - Implementation decision

- [x] Accept diplomatic state schema and symmetric/directed invariants.
- [x] Accept declare-war and peace command payload/authority/sequence semantics.
- [x] Accept turn/phase boundary for relation changes.
- [x] Accept attack authorization mapping consumed by blockade/encounter logic.
- [x] Accept event/history/projection model.
- [x] Freeze deferred treaty/espionage/AI families.

## Gate 3 - Implementation

- [x] Implement accepted diplomatic state/validation changes.
- [x] Implement war/peace commands and deterministic events.
- [x] Integrate hostility/attack authorization into existing blockade and encounter derivation.
- [x] Add player/observer/API projections and web command proof.
- [x] Add deterministic transition, replay, same-system timing and invalid-authority tests.
- [x] Preserve Slice-06/07 encounter/tactical behavior when war is active.

## Gate 4 - Follow-up QA + commit + close

- [x] Run focused diplomacy/hostility/blockade/encounter regressions.
- [x] Verify peace/war state survives save/load/replay exactly.
- [x] Verify no hidden UI/network authority bypass exists.
- [x] Run full tests/vet/web integration checks and `git diff --check`.
- [x] Update evidence/status/HISTORY and close marker.

## Exit criterion

Two real Empires can authoritatively enter war and return to peace through deterministic player commands, and the resulting diplomatic state is the single input used by strategic attack/blockade/encounter authorization. Broader MOO2 negotiation, treaties and espionage remain explicit future work.
