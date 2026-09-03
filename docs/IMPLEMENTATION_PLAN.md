# MOOX Implementation Plan

## Current position - 2026-09-01

MOOX is actively inside the deterministic headless runtime. The authoritative `GameSession` boundary now executes a substantial first strategic Economy/Research/Construction path rather than only infrastructure.

Implemented/established runtime baseline includes:

- deterministic pure-Go Core state, seeded serializable RNG, stable IDs and strict schema-21 JSON round trips;
- authoritative Session/Command/Event/Observer architecture with Human/AI-shared legal-action projections;
- domain-native `float64` cohort Population quantities, Food, PP, RP, BC and Construction progress (`ADR-0003`);
- base/contextual Colony Economy with Gravity, starting-government and local Morale layers;
- race-aware organic Population cohorts with heterogeneous capacity, Food/Cybernetic sustenance, per-origin starvation and the classic capacity-limited turn-end Growth curve;
- Empire-wide Food/Freighter balancing with exact four-pass mixed-cohort priority, verified constrained round-robin allocation and system-blockade exclusion;
- minimal strategic Fleet location/role plus directed Empire relations, with deterministic Colony-presence/hostile-combat-Fleet blockade production before Settler arrival/final Food;
- Colony Ship production through semantic Construction, installed-drive persistence, original Fuel Cell ranges, deterministic strategic transit before Construction, explicit colonization and authoritative second-Colony creation/ship consumption;
- semantic Population relocation / Settlers sharing the same Freighter pool, including same-system movement, interstellar reservation, ETA and resolution;
- generic single-project Construction with technology-gated Buildings and Freighter Fleet production;
- modeled Treasury settlement before Research;
- normalized Technology Fields/RP costs, Pre-Warp/Average/Advanced starts, breakthrough resolution, race-aware multi-application research, project switching and Hyper-Advanced repeat fields;
- authoritative external Technology grants with verified Uncreative fixed-choice repair.

The first supported deterministic headless match lifecycle is now complete through Slice 12: authoritative server/web transport, deterministic New Game, war/peace, Troop Transport/invasion/Colony conquest and zero-Colony conquest victory are connected end-to-end. Broader tactical depth, complete Economy/pollution, race/customization extensions, leaders/espionage, AI, alternative victory families and endgame fidelity remain important post-milestone work.

Slice 14 **Live GameSession save/resume baseline** is currently open. Gate 2 design review is complete. The reviewed LiveSnapshot v1 freeze candidate adds full Session/Battle continuation state over Core23, simulation compatibility, cross-platform compacted ruleset fingerprinting, production-resolver/controller scope, stable interactive save boundaries and atomic Host import/restore semantics; approval is pending and no persistence implementation has started. Slices 15-17 remain prepared and unopened. `docs/research/ACTIVE_RESEARCH.md` remains the authoritative handoff.
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

ADR-0004 selects an authoritative Go game server plus browser-first HTTP/WebSocket HMI as the primary application topology. Wails v3 is optional native packaging only and must consume the same server contract rather than expose gameplay bypasses. ADR-0001's pure-Go/Wails-independent core boundary remains valid; Core simulation, persistence, analyzers and developer tools continue to target pure-Go builds with `CGO_ENABLED=0`.

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

The post-Slice-12 audit changes the immediate objective from "finish enough rules for a complete match" to "make the proven lifecycle independently playable". Prepared queue:

1. **Slice 13 - Built-in strategic AI baseline.** Deterministic, rules-legal, no-cheat AI consuming the same Session/App command surfaces as human seats.
2. **Slice 14 - Live GameSession save/resume baseline.** Versioned in-progress persistence at stable interactive boundaries, exact continuation and server import/export surfaces.
3. **Slice 15 - Playable browser strategic HMI loop.** Human-vs-built-in-AI browser match covering the currently supported strategic commands, live save/load and final result.

Milestone after Slice 15: a human can start, save/resume and finish the supported deterministic single-player game from the browser.

Then widen fidelity/depth:

4. **Slice 16 - New Game + preset-race breadth.** Expand the current Small/Normal/Average/exactly-two/Human+Darlok contract using evidence-backed settings and the 13 normalized preset races.
5. **Slice 17 - Military ship design component/weapon breadth.** Expand the current one-hull/one-Laser minimal design into a canonical multi-hull/component/multi-weapon model before deeper Tactical Combat.

Keep Tactical Combat depth, treaty/trade diplomacy, Espionage/Leaders, Economy/Population fidelity, full custom Race Designer, Council/Orion/Antaran victories, random events, multiplayer/external AI and presentation unnumbered until the next audit after Slice 17.

Do not reopen closed Research/Economy slices listed in `docs/slices/HISTORY.md` unless contradictory original evidence or a concrete runtime regression requires it. ADR-0004 makes the server/web contract primary; browser, optional Wails wrapper, AI and other transports must continue to consume the same canonical Session/legal-action surfaces rather than owning gameplay rules.