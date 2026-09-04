# Planned slice 20 - Espionage / intelligence baseline

Status: **planned / queued; not open**.

Queue position: **reserved future Slice 20 by product decision**. Slices 18/19 are currently unassigned and need not be invented merely to fill numbering.

## Objective

Add a real server-authoritative MOO2-like Espionage/Intelligence gameplay baseline and connect it to the first-class Espionage product area reserved by Slice 15.2.

Slice 15.2 deliberately does **not** fake Spy controls because the current runtime has only Spying modifiers, not Spy gameplay state or commands.

## Original direction to preserve

The current local MOO2 reference states that Spies can be assigned to defense or offensive missions; offensive intelligence includes espionage and sabotage; race spying traits and computer-related technology materially affect the system.

Current normalized MOOX data already includes Spying bonuses such as the Darlok +20 modifier. Slice 20 must route these through one authoritative mechanics pipeline rather than adding HMI-only percentages.

## Dependencies

- strategic player-safe snapshot/DecisionView transport from Slice 15.2;
- first-class Espionage navigation/product location from Slice 15.2;
- normalized race/technology modifiers already present;
- deterministic turn/session/save infrastructure;
- built-in AI controller boundary.

## Scope guard

In scope for the first playable baseline:

- authoritative Spy/intelligence state per Empire;
- Spy production/training acquisition path backed by legal Construction choices where historically applicable;
- defensive/offensive Spy allocation;
- target-Empire selection;
- at least Espionage and Sabotage mission families;
- deterministic success/failure/detection resolution;
- captured/lost Spy consequences needed by the baseline;
- technology theft/transfer result for Espionage;
- a bounded Sabotage result family against legal targets;
- race SpyingBonus and relevant Computer/technology modifiers through server rules;
- player-safe intelligence projection: own Spy state, legal targets/actions, known results, no hidden enemy internals;
- event/report feedback;
- built-in AI policy sufficient to keep AI games progressing;
- save/load/replay determinism;
- DE/EN browser Espionage screen using the previously reserved product location.

Defer unless Gate 2 of Slice 20 explicitly broadens scope:

- complete original diplomacy framing/blame/false-flag personality interactions;
- every original sabotage target/result;
- advanced counter-intelligence personalities;
- leader-specific Spy breadth beyond existing generic modifier hooks;
- advanced treaty/tribute/technology-trade diplomacy dependencies;
- exact original probability tables if evidence remains incomplete.

## Gate 1 - evidence / mechanics audit

- [ ] Re-check original manual/help/executable/save evidence for Spy acquisition, allocation, mission timing and outcomes.
- [ ] Map normalized Spying/Computer modifiers to the original formulas or record explicit compatibility formulas.
- [ ] Identify exact state required for defensive/offensive Spy counts and target missions.
- [ ] Define player-safe information boundaries and what an opponent may learn on detection.
- [ ] Define minimum Espionage and Sabotage result families.
- [ ] Define deterministic AI policy and turn-resolution phase placement.
- [ ] Define save/replay/state-schema implications.
- [ ] Present Gate-2 authoritative contract.

## Gate 2 - authority freeze

- [ ] Freeze state schema and normalized rules data.
- [ ] Freeze Spy production/acquisition semantics.
- [ ] Freeze defensive/offensive assignment and target commands.
- [ ] Freeze success/detection/counter-intelligence formulas from evidence.
- [ ] Freeze technology-theft and Sabotage result boundaries.
- [ ] Freeze player-safe DecisionView/API projection.
- [ ] Freeze AI and event/report behavior.
- [ ] Freeze browser interaction contract and acceptance scenarios.

## Gate 3 - implementation

- [ ] Implement authoritative Spy state/rules/commands.
- [ ] Integrate race/technology modifiers.
- [ ] Implement deterministic mission resolution and events.
- [ ] Add legal decision catalogs and player-safe projection.
- [ ] Add built-in AI behavior.
- [ ] Add persistence/replay support.
- [ ] Implement the first-class browser Espionage workflow in DE/EN.
- [ ] Add deterministic regressions.

## Gate 4 - QA + close

- [ ] Execute Human-vs-AI Spy acquisition/assignment/mission flow through browser authority.
- [ ] Verify no hidden enemy state leaks.
- [ ] Verify save/reload/replay determinism around pending/completed missions.
- [ ] Run full Go tests/vet/web build and `git diff --check`.
- [ ] Update roadmap/evidence/HISTORY and close marker.

## Exit criterion

A Human player can acquire/manage Spies, assign defensive/offensive intelligence, choose a legal enemy target, execute the accepted Espionage/Sabotage baseline through normal authoritative turn resolution, receive clear result feedback, save/reload without divergence, and play against built-in AI without any Spy rules living in React.
