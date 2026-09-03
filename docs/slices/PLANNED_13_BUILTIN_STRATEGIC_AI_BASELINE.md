# Planned slice 13 - Built-in strategic AI baseline

Status: **planned / queued; not open**.

Queue position: **13 of 17**.

This is a prepared slice specification, not an `_OPEN_*.md` recovery marker. Gate 1 must be run fresh when the slice is explicitly started.

## Objective

Implement the first deterministic built-in AI controller that can legally drive the already-supported complete match lifecycle through the same authoritative Session/App command surfaces used by human seats.

## Dependencies

- Slices 01-12 closed.
- `ControllerBuiltinAI` already exists as a valid controller identity.
- Current supported New Game remains the narrow Slice-09 setup unless Gate 1 proves a small compatibility fix is required.

## Scope guard

In scope:

- deterministic AI decision loop at stable interactive boundaries;
- Research selection/completion decisions;
- Colony job/construction choices needed by the current lifecycle;
- Colony Ship/Outpost expansion and Fleet movement;
- war/peace baseline decisions;
- Encounter/Battle/Invasion choices required to keep the game progressing;
- use of public legality/projection/command surfaces only;
- AI-vs-AI and human-vs-AI deterministic headless regressions;
- explicit no-cheat boundary: no hidden BC/PP/RP, no direct state edits, no omniscient use of hidden player data.

Defer:

- exact original MOO2 AI behavior/personality fidelity;
- difficulty levels/cheat modifiers;
- stronger modern search/planning AI;
- broader New Game settings/races;
- diplomacy treaty/trade/espionage depth.

## Gate 1 - Checkup + evidence/architecture audit

- [ ] Re-check repo/session/controller/command surfaces and canonical Slice-12 match.
- [ ] Inventory every stable boundary that can require a seat decision.
- [ ] Inspect available original AI evidence and separate proven original behavior from MOOX baseline policy.
- [ ] Define deterministic AI input visibility and no-cheat contract.
- [ ] Define minimal legal-action/query gaps that must be exposed by Go authority rather than recomputed by AI.
- [ ] Define AI-vs-AI completion fixture/turn cap and failure diagnostics.
- [ ] Present exact Gate-2 AI contract before implementation.

## Gate 2 - Implementation decision

- [ ] Freeze AI determinism/no-cheat contract.
- [ ] Freeze supported decisions and intentionally deferred systems.
- [ ] Freeze AI lifecycle integration point and event/revision behavior.
- [ ] Freeze canonical AI-vs-AI and human-vs-AI regressions.

## Gate 3 - Implementation

- [ ] Implement built-in AI command producer/driver.
- [ ] Integrate all supported stable decision boundaries.
- [ ] Add deterministic legal-action and atomic-failure regressions.
- [ ] Prove at least one AI-driven game reaches authoritative completion without direct state mutation.

## Gate 4 - Follow-up QA + commit + close

- [ ] Repeat AI games and compare exact result/event history.
- [ ] Run full tests/vet/web checks and `git diff --check`.
- [ ] Update evidence/status/HISTORY and close marker.