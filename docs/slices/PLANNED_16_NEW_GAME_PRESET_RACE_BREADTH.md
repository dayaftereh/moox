# Planned slice 16 - New Game and preset-race breadth

Status: **planned / queued; not open**.

Queue position: **16 of 17**.

## Objective

Widen the deliberately narrow Slice-09 New Game contract using existing original/normalized evidence, with particular focus on preset races and settings that can now be exercised by built-in AI and the browser HMI.

## Existing narrow boundary

Current runtime explicitly accepts only Small galaxy, Normal age, Average technology, Tactical combat, exactly two players, exactly one Human and one Darlok. The normalized ruleset already contains 13 preset races.

## Dependencies

- Slice 13 built-in AI.
- Slice 15.5 completed browser vertical slice for the established user-facing shell; this slice then widens its New Game/race breadth.

## Scope guard

In scope candidates to freeze during Gate 1:

- additional preset races with correct starting modifiers/technologies/government;
- more than two seats where evidence/runtime invariants are ready;
- additional galaxy sizes/ages/technology levels;
- difficulty/controller assignment if it can be separated cleanly from AI quality;
- deterministic seed equality per supported settings tuple;
- HMI exposure for accepted options.

Defer:

- full custom Race Designer;
- unsupported trait interactions not required by preset races;
- Council/Antaran/Orion victory;
- broad AI personality fidelity.

## Gate 1 - Original/settings audit

- [ ] Re-check original New Game option tables and generator evidence.
- [ ] Inventory which of 13 normalized races are already runtime-compatible without new subsystem work.
- [ ] Map size/age/tech/player-count options to generator rules/evidence.
- [ ] Identify preset races blocked by missing trait/government/population mechanics.
- [ ] Choose a bounded expansion set rather than enabling unsupported options optimistically.
- [ ] Define seed/hash fixtures for each newly accepted setup.
- [ ] Present Gate-2 breadth contract.

## Gate 2 - Implementation decision

- [ ] Freeze newly supported settings/races/player counts.
- [ ] Freeze exact unsupported/deferred list.
- [ ] Freeze deterministic fixture matrix and HMI exposure.

## Gate 3 - Implementation

- [ ] Expand validated New Game settings/race support.
- [ ] Add required preset-race runtime modifiers only where evidence-backed.
- [ ] Extend server/HMI creation surfaces.
- [ ] Add deterministic multi-seed/settings regressions.

## Gate 4 - Follow-up QA + commit + close

- [ ] Repeat each supported settings fixture exactly.
- [ ] Verify AI and HMI can start/play newly accepted setups.
- [ ] Run full tests/vet/web checks and `git diff --check`.
- [ ] Update evidence/status/HISTORY and close marker.