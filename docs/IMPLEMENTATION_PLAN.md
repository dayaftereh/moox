# MOOX Implementation Plan

## Guiding architecture

Build the game rules as a deterministic, headless simulation core first. UI, rendering, audio and platform integration should depend on that core rather than contain game rules themselves.

Suggested conceptual layers:

```text
moox/
  core/          deterministic game state + turn processing
  rules/         data-driven rulesets and modifiers
  ai/            strategic/tactical decision systems
  persistence/   save/load and migrations
  net/           optional deterministic multiplayer protocol
  ui/            application/UI layer
  tools/         data import, original-game research, balancing
  data/          MOOX-owned normalized rules/content
```

Architecture decision ADR-0001 selects a Go-first core with Wails v3 as the planned application shell. Core simulation, persistence, analyzers and developer tools remain Wails-independent and target pure-Go builds with `CGO_ENABLED=0`.

## Phase 0 - Research baseline

- Capture MOO2 systems and vocabulary.
- Separate official 1.31 baseline from community mods.
- Catalog legal/public references.
- Establish policy for original copyrighted assets.
- When available, inventory a legally owned original installation.

Exit criterion: enough structured knowledge to write deterministic subsystem tests.

## Phase 1 - Simulation skeleton

Implement:

- seeded RNG service,
- stable IDs,
- galaxy/star/planet model,
- empire/race model,
- colony/population model,
- turn clock,
- serialization,
- event log.

No graphical UI required.

Exit criterion: generate a fixed small galaxy and round-trip the exact state through save/load.

## Phase 2 - Colony economy

Implement:

- farmers/workers/scientists,
- population growth,
- food and freighters,
- production,
- pollution,
- research points,
- BC economy,
- morale,
- buildings and maintenance,
- production queue/buyout.

Exit criterion: reference colony scenarios reproduce expected per-turn outputs.

## Phase 3 - Technology

Implement:

- eight research fields,
- research levels and choices,
- Creative/Uncreative behavior,
- tech acquisition from non-research sources,
- unlock/effect system,
- component miniaturization.

Exit criterion: complete research progression can run headlessly and deterministically.

## Phase 4 - Strategic expansion

Implement:

- ship range/strategic movement,
- fuel/range modifiers,
- colony/outpost ships,
- transports,
- fleet ownership/merging,
- colonization,
- blockade state,
- command points.

Exit criterion: one empire can expand from homeworld to a second colony solely through normal game rules.

## Phase 5 - Race designer and governments

Implement:

- pick budget,
- trait costs,
- incompatible traits,
- special abilities,
- government rule modules,
- preset race templates.

Exit criterion: preset templates and valid custom builds produce deterministic modifier sets.

## Phase 6 - Ship designer

Implement:

- military hulls,
- drives/computers/armor/shields,
- beams/missiles/bombs/fighters,
- specials,
- weapon modifications,
- space/cost calculations,
- refits and design-slot behavior.

Exit criterion: canonical MOO2-style designs can be represented and validated by the engine.

## Phase 7 - Tactical combat

Implement:

- tactical map,
- initiative/turn order,
- movement/facing as required by observed original behavior,
- weapon firing/hit/damage,
- shields/armor/structure/internal systems,
- missiles/fighters,
- boarding,
- retreat,
- planetary defenses.

Exit criterion: deterministic combat fixtures with fixed RNG seeds.

## Phase 8 - Diplomacy, spying, leaders

Implement:

- relation state/history,
- treaties,
- negotiations,
- technology exchange,
- war/peace,
- espionage/sabotage,
- leaders and skills,
- race personality/AI diplomacy hooks.

Exit criterion: two AI empires can conduct a complete diplomacy lifecycle.

## Phase 9 - Conquest and victory

Implement:

- troop transports,
- bombardment/invasion,
- conquered population/assimilation,
- Galactic Council,
- Orion/Guardian,
- Antaran attack system,
- dimensional gate/end battle,
- elimination victory.

Exit criterion: all main victory paths are executable in headless tests.

## Phase 10 - AI

Separate AI from simulation rules:

- empire personality,
- economy planner,
- research planner,
- colonization planner,
- diplomacy planner,
- ship designer,
- fleet strategy,
- tactical combat AI,
- explicit difficulty modifiers.

Support two goals:

1. original-like behavior/difficulty,
2. optional stronger modern behavior without economy cheats.

## Phase 11 - UI vertical slice

Only after the rules core is stable, implement the minimum recognizable loop:

- new game,
- galaxy map,
- colony screen,
- research screen,
- fleet movement,
- turn button,
- save/load.

Then add:

- race designer,
- diplomacy,
- ship design,
- tactical combat,
- leaders,
- reports/graphs.

## Phase 12 - Content and presentation

Use original MOO2 presentation only as interaction reference. Produce independent:

- MOOX branding,
- portraits,
- ship art,
- planet/star art,
- UI skin,
- music,
- sound effects,
- cinematics.

A private developer-only original-asset viewer may exist for comparison, but original assets should not be required by the distributable game unless a deliberate licensed-data compatibility mode is created.

## Recommended immediate next work

1. Decide implementation language/engine after evaluating desktop targets and UI needs.
2. Create core project skeleton and tests.
3. Implement deterministic RNG/state serialization.
4. Define normalized schemas for race, technology, buildings, planets and ship components.
5. Obtain/install a legitimate MOO2 copy locally and build a read-only archive inventory tool.
6. Start exact 1.31 data capture with provenance and parity fixtures.
