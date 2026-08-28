# Master of Orion X

**Master of Orion X** is a new game project whose gameplay reference is **Master of Orion II: Battle at Antares (1996)**.

The initial goal is a high-fidelity reimplementation of the systems that make MOO2 work: turn-based 4X galaxy strategy, colony management, custom race design, research, diplomacy/espionage, ship design, tactical combat, leaders, random events, Orion/Antaran end-game content and the original pacing/feedback loops.

## Product naming

The official user-facing project and product name is **Master of Orion X**.

This name must be used consistently for all user-visible branding, including:

- application and game name: **Master of Orion X**;
- Windows executable: **Master of Orion X.exe**;
- application/window title: **Master of Orion X**;
- launcher, splash-screen and main-menu titles: **Master of Orion X**;
- website/frontend page titles and primary headlines: **Master of Orion X**;
- release/package/display names: **Master of Orion X**.

The short identifier `moox` is reserved for internal technical names where a compact identifier is useful, such as the Git repository, Go module/package paths, command-line tooling (`moox-analyze`), internal directories and development-only identifiers. It must not replace **Master of Orion X** in user-facing application titles or branding.

## Project status

Current snapshot: **2026-08-27**.

Master of Orion X is transitioning from the advanced research/data-normalization phase into the deterministic headless runtime. The repository contains the pure-Go MOO2 1.31 analyzer (`moox-analyze 0.22.0`), normalized runtime datasets and the first simulation/session infrastructure.

The game is **not playable yet**, but the first strategic economy/construction/research path now runs through the authoritative session boundary. Population allocation, Food, Production, Research, BC and Construction progress use domain-native `float64` values; Population capacity, Food/Cybernetic sustenance, empire-wide Food/Freighter logistics, starvation and turn-end Population Growth are now materialized as well. Gravity, starting-government and local Morale context layer on top without implicit intermediate rounding. Technology ownership/buildability plus server-derived BuildingChoices and ResearchChoices are shared by future Human UI and AI controllers.

See `docs/PROJECT_STATUS.md` for the complete current-state snapshot and `docs/research/ACTIVE_RESEARCH.md` for the one active research objective.
### Latest runtime checkpoints

- `56b0096` - deterministic single-project colony construction using original MOO2 1.31 building PP costs and BC maintenance data.
- `17fc2c2` - persistent technology ownership, technology-gated building construction and authority-filtered building legal actions shared by Human UI and AI callers.
- `17791c1` - original technology-field/RP-cost normalization, deterministic Pre-Warp/Average new-game ownership and authoritative research-completion ownership transition.
- `30ee47b` - verified original standard-field breakthrough curve/RNG/overflow behavior and automatic strategic research resolution.

The active runtime follows a domain-native numeric architecture: continuous quantities use `float64` (`PopulationState.Total/Farmers/Workers/Scientists`, Food, PP, RP, BC and Construction progress), while genuinely discrete IDs/counts remain discrete. Population Growth now uses the classic curve directly in Population units, and freshly grown Population cannot retroactively contribute PP/RP to the turn that produced it. Rounding occurs only at explicit gameplay-rule boundaries. The Economy checkpoint includes Food/Freighter balancing and starvation. Research now models General `all`, ordinary `choose_one`, Creative `all`, and Uncreative server-fixed `fixed_one` application semantics. Active research can now be switched while preserving the complete accumulated RP pool. Direct original-executable analysis has resolved the 1.31 turn order: current RP/PP are materialized pre-growth; Research is applied first, then Population Growth/Starvation, then Construction using the pre-growth PP snapshot. Direct original-executable analysis has now also resolved Uncreative initial-selection timing/RNG ownership: fixed applications are generated during new-game player initialization from the shared New Game RNG. Hyper-Advanced repeated-field progression and Advanced-start randomized/race-aware technology ownership are now implemented from direct original-executable evidence. Advanced start uses the shared New Game RNG, cross-Empire competition weighting, and exactly 19 weighted extra grants after the Average baseline. The next Research work is the authoritative external Technology-grant path that applies the verified Uncreative repair transition.
## Research documents

- `docs/research/MOO2_GAME_REFERENCE.md` - gameplay/system reference and fidelity checklist.
- `docs/research/MOO2_DATA_AND_FILE_FORMATS.md` - original data formats, LBX research and extraction strategy.
- `docs/research/SOURCES.md` - curated source register with links and usage notes.
- `docs/WORKING_RULES.md` - mandatory convergence/checkpoint rules for long research sessions.
- `docs/research/ACTIVE_RESEARCH.md` - authoritative current objective, next exact action, blockers and closed milestones.
- `docs/research/ECONOMY_BASELINE_2026-08-27.md` - first population-command and base colony-economy fidelity checkpoint.
- `docs/research/ECONOMY_CONTEXT_2026-08-27.md` - Gravity/Government contextual economy checkpoint and current model limitations.
- `docs/research/ECONOMY_MORALE_2026-08-27.md` - local Morale/Barracks/building checkpoint and deferred empire-wide morale systems.
- `docs/research/POPULATION_GROWTH_SUSTENANCE_2026-08-27.md` - float-native Population capacity, sustenance, growth and turn-order checkpoint.
- `docs/research/FOOD_FREIGHTER_LOGISTICS_2026-08-27.md` - empire Food balancing, Freighter capacity/cost, surplus valuation and starvation checkpoint.
- `docs/research/BUILDING_CONSTRUCTION_2026-08-27.md` - original building costs/maintenance, deterministic construction and technology-gated buildability checkpoint.
- `docs/research/TECHNOLOGY_START_RESEARCH_2026-08-27.md` - original technology/tech-field tables, Pre-Warp/Average ownership initialization and research-completion ownership transition.
- `docs/research/RESEARCH_BREAKTHROUGH_2026-08-27.md` - original breakthrough evidence plus the documented MOOX float-RP divergence.
- `docs/research/RESEARCH_SELECTION_2026-08-27.md` - authority-filtered ResearchChoices and `empire.select_research`.
- `docs/research/RESEARCH_MULTI_TECH_2026-08-27.md` - General/ordinary/Creative/Uncreative multi-application research semantics and deterministic Uncreative choice plan.
- `docs/research/UNCREATIVE_INITIAL_SELECTION_2026-08-28.md` - direct MOO2 1.31 proof of Uncreative new-game selection timing, shared RNG ownership, field range and race/game-mode rejection rules.
- `docs/research/HYPER_ADVANCED_RESEARCH_2026-08-28.md` - direct MOO2 1.31 Hyper-Advanced counters, dynamic strategic cost, repeat completion semantics and the original UI preview off-by-one.
- `docs/research/RESEARCH_SWITCHING_2026-08-27.md` - active-project switching with complete accumulated-RP transfer and Observer/replay event.
- `docs/research/TURN_ORDER_2026-08-28.md` - direct MOO2 1.31 executable proof of pre-growth RP/PP snapshots and Research -> Population -> Construction apply order.
- `docs/architecture/ADR-0002-research-float64.md` - historical first RP-native `float64` decision.
- `docs/architecture/ADR-0003-domain-native-float64.md` - canonical continuous-quantity `float64` and explicit rounding-boundary architecture.
- `docs/IMPLEMENTATION_PLAN.md` - proposed clean-room development phases.
- docs/ANALYZER.md - pure-Go MOO2 console analyzer usage and architecture.
- docs/architecture/README.md - runtime architecture overview for parallel turns, authoritative sessions, battles, observer and AI.
- docs/architecture/ADR-0001-go-wails-v3.md - Go/Wails v3 portability decision and layer boundary.
- docs/architecture/CORE.md - deterministic headless simulation state, RNG and save/load contract.
- docs/architecture/SESSION_PROTOCOL.md - implemented session, command, observer and battle-boundary contract.
- `reference/README.md` - policy for locally supplied original-game files.

## Fidelity target

The target is behavioral fidelity rather than copying copyrighted material. We want to reproduce the *rules and interaction model* closely enough that an experienced MOO2 player recognizes the game immediately, while keeping the codebase, branding, art, audio and other expressive assets independently created unless we have explicit rights to use an asset.

Important fidelity areas:

1. New-game configuration and galaxy generation.
2. Star systems, planets, climate, size, gravity, mineral richness and specials.
3. Population, food, production, research, pollution, morale, money and freighters.
4. Race picks, governments and special racial abilities.
5. The eight research fields, technology choices, Creative/Uncreative behavior and miniaturization.
6. Diplomacy, treaties, relations, espionage and sabotage.
7. Colony/building construction queues and buyout behavior.
8. Ship hulls, components, weapon modifications, refits and command points.
9. Tactical space combat and planetary invasion.
10. Leaders and their colony/ship bonuses.
11. Random events, space monsters, Orion/Guardian and Antaran attacks.
12. Galactic Council, extermination and Antaran victory paths.
13. Save/load and deterministic simulation suitable for regression testing.

## Original Master of Orion II data

Do **not** commit original MOO2 game binaries, manuals, screenshots, music, sound, `.LBX` archives or other copyrighted assets into this repository unless their redistribution rights have been explicitly established.

A legitimate GOG/Steam/local installation can still be very useful as a private reference source. The planned tooling will be able to point at such an installation, catalog archive structures/hashes and extract information into local ignored directories. Original files and extracted copyrighted assets remain outside Git.

A local Master of Orion II 1.31 reference installation was subsequently found at `C:\ASH\Temp\mastori2`. It is treated as private research input and is not copied into Git; see `docs/research/LOCAL_REFERENCE_2026-08-26.md`.

## Public reverse-engineering references

Two especially useful starting points are:

- OpenMOO2 (`mimi1vx/openmoo2`): GPL-2.0 open-source MOO2 clone. It requires an original MOO2 release and consumes the original `.LBX` data files.
- ModdingWiki LBX documentation: documents the SimTex LBX container used by MOO, Master of Magic and MOO2.

These projects are references, not dependencies at this stage. If code is reused later, its license obligations must be handled explicitly.

## Next milestone

**Phase 1 - deterministic simulation skeleton is active.** The implemented runtime now covers:

- simulation-owned deterministic RNG and exact save/load,
- stable game-state IDs/types and the fixed headless fixture,
- versioned multiplayer command batches and transactional strategic resolution,
- parallel seat submissions plus Player/Observer projections,
- discrete farmer/worker/scientist assignment,
- the first real `colony.assign_population` strategic command,
- domain-native `float64` Population/Food/Production/Research/BC output from normalized ruleset data,
- explicit Gravity/Government/local-Morale economy context and adjusted output,
- original building PP costs and BC/turn maintenance normalized from the MOO2 1.31 `_buildings` table,
- single-project deterministic colony construction driven by adjusted production,
- persistent `Empire.KnownTechnologyIDs` and technology-gated `colony.queue_building`,
- authority-filtered `GameSession.BuildingChoices` as the shared legal-action surface for future UI and AI,
- server-owned Seat -> Empire authorization,
- deterministic construction/domain events and tactical BattleSession boundaries.

The research runtime uses RP-native `float64` progress. `GameSession.ResearchChoices` exposes race-aware `all` / `choose_one` / `fixed_one` / `repeat_field` legal actions. `empire.select_research` accepts `technology_id` only when an ordinary race must choose one application; Creative/General and Uncreative-fixed projects remain server-owned. Research/Observer events carry selection mode plus numeric IDs and speaking keys. Active-project switching with complete RP transfer is implemented. Uncreative initial-selection timing/RNG ownership and Hyper-Advanced repeated-field progression are directly verified. Hyper fields use server-owned `repeat_field` legal actions with dynamic strategic costs. Advanced start is implemented as six fixed Average fields plus exactly 19 weighted extra grants. The full-state generator uses normalized Technology AI classes/class weights/field-group values, semantic preference profiles, stable Empire order, one caller-owned New Game RNG, race-aware ordinary/Creative/Uncreative acquisition, and Strategic Combat filtering. The next Research slice is the authoritative external Technology-grant path with Uncreative fixed-choice repair.

Wails v3 is already the planned application shell; rendering/UI framework selection is no longer an open prerequisite.
## Runtime ruleset data

Normalized rules live under `data/rulesets/moo2-1.31/`; user-visible text is separate under `data/languages/` and rules reference stable translation keys.

Current committed coverage includes:

- 11 Race Designer groups / 53 options,
- 13 preset races,
- 203 technology identities,
- 48 building identities with original technology links, PP costs and BC/turn maintenance,
- 6 military ship hull identities,
- planet-class primitives: 5 sizes, 5 mineral classes, 3 gravity classes and 10 climates,
- 156 semantic asset records (143 confirmed / 13 intentionally pending), including 1,728 building-position variants, 352 strategic ship variants and 6,560 tactical ship-frame variants.

English currently contains 357 runtime keys. German/French/Spanish/Italian each contain the 64 verified Race Designer keys and fall back to English for other normalized names.

See `data/rulesets/moo2-1.31/README.md` for field-level provenance and `docs/PROJECT_STATUS.md` for the current completeness/gap table.
