# MOOX Implementation Plan

## Current position - 2026-08-27

MOOX is now actively inside the deterministic headless runtime. Phase 1 infrastructure is established and the first Phase-2/Phase-3 Economy/Research slices already execute through the authoritative `GameSession` boundary.

Implemented/established runtime baseline:

- deterministic pure-Go core state, seeded serializable RNG, stable IDs and exact JSON round trips,
- authoritative Session/Command/Event/Observer architecture with Human/AI-shared legal-action projections,
- domain-native `float64` Population, Food, PP, RP, BC and Construction progress (`ADR-0003`),
- base/contextual colony Economy with Gravity, starting-government and local Morale layers,
- technology-gated BuildingChoices plus deterministic single-project Construction,
- normalized Technology Fields/RP costs, Pre-Warp/Average ownership, breakthrough resolution and authority-filtered `ResearchChoices`,
- server-side `empire.select_research` materialization of Technology IDs/keys,
- Population capacity, Food/Cybernetic sustenance and direct-float turn-end Population Growth with explicit pre-growth Construction/Research ordering.

Still missing major engine systems include strategic movement/colonization, complete Economy logistics/pollution/maintenance, complete Technology effects and Creative/Uncreative semantics, ship design/combat rules, diplomacy and playable UI.

Current work is deliberately two-lane: Economy/Population is the immediate implementation focus, while Research multi-Technology TechField behavior remains the next Research-specific slice. `docs/research/ACTIVE_RESEARCH.md` is the authoritative handoff for both lanes.
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

1. Model original Hyper-Advanced repeated-field level/cost state and remove the current explicit runtime rejection only after the repeated-level state is verified.
2. Implement Advanced-start randomized/race-aware technology ownership using the completed General/ordinary/Creative/Uncreative policy.
3. Later implement Uncreative external-acquisition replacement using `Ensure_Uncreative_Field_OK_` and resolve the remaining Dimensional Portal eligibility gate when needed.
4. Then continue the first headless vertical slice toward colony ship production, movement and a second colony.
Do not reopen closed race/building/ship graphics research unless contradictory evidence or a concrete runtime requirement demands it. Wails v3 remains the planned application shell; transports/UI must continue to consume the same canonical Session/legal-action APIs rather than contain gameplay rules.