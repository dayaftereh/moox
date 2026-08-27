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

The game is **not playable yet**, but the first strategic economy/construction path now runs through the authoritative session boundary. It includes server-authorized population commands, fixed-point base output, Gravity, starting-government and local Morale context, single-project deterministic building construction, original building PP cost/BC maintenance data, persistent technology ownership and technology-gated legal building choices. The same server-derived building-choice projection is available to future Human UI and AI controllers; starting technologies are deliberately not guessed and are the active research target.

See `docs/PROJECT_STATUS.md` for the complete current-state snapshot and `docs/research/ACTIVE_RESEARCH.md` for the one active research objective.
### Latest runtime checkpoints

- `56b0096` - deterministic single-project colony construction using original MOO2 1.31 building PP costs and BC maintenance data.
- `17fc2c2` - persistent technology ownership, technology-gated building construction and authority-filtered building legal actions shared by Human UI and AI callers.
- `17791c1` - original technology-field/RP-cost normalization, deterministic Pre-Warp/Average new-game ownership and authoritative research-completion ownership transition.
- `30ee47b` - verified original standard-field breakthrough curve/RNG/overflow behavior and automatic strategic research resolution.

The active implementation sequence has moved past breakthrough research: RP-native `float64` progress is now the MOOX architecture, `ResearchChoices` exposes the server-authoritative frontier, and `empire.select_research` lets Human UI and AI submit only a TechField while the server materializes its Technology set.
## Research documents

- `docs/research/MOO2_GAME_REFERENCE.md` - gameplay/system reference and fidelity checklist.
- `docs/research/MOO2_DATA_AND_FILE_FORMATS.md` - original data formats, LBX research and extraction strategy.
- `docs/research/SOURCES.md` - curated source register with links and usage notes.
- `docs/WORKING_RULES.md` - mandatory convergence/checkpoint rules for long research sessions.
- `docs/research/ACTIVE_RESEARCH.md` - authoritative current objective, next exact action, blockers and closed milestones.
- `docs/research/ECONOMY_BASELINE_2026-08-27.md` - first population-command and base colony-economy fidelity checkpoint.
- `docs/research/ECONOMY_CONTEXT_2026-08-27.md` - Gravity/Government contextual economy checkpoint and current model limitations.
- `docs/research/ECONOMY_MORALE_2026-08-27.md` - local Morale/Barracks/building checkpoint and deferred empire-wide morale systems.
- `docs/research/BUILDING_CONSTRUCTION_2026-08-27.md` - original building costs/maintenance, deterministic construction and technology-gated buildability checkpoint.
- `docs/research/TECHNOLOGY_START_RESEARCH_2026-08-27.md` - original technology/tech-field tables, Pre-Warp/Average ownership initialization and research-completion ownership transition.
- `docs/research/RESEARCH_BREAKTHROUGH_2026-08-27.md` - original breakthrough evidence plus the documented MOOX float-RP divergence.
- `docs/research/RESEARCH_SELECTION_2026-08-27.md` - authority-filtered ResearchChoices and `empire.select_research`.
- `docs/architecture/ADR-0002-research-float64.md` - accepted RP-native `float64` architecture and explicit rounding policy.
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
- fixed-point base food/production/research/tax colony output from normalized ruleset data,
- explicit Gravity/Government/local-Morale economy context and adjusted output,
- original building PP costs and BC/turn maintenance normalized from the MOO2 1.31 `_buildings` table,
- single-project deterministic colony construction driven by adjusted production,
- persistent `Empire.KnownTechnologyIDs` and technology-gated `colony.queue_building`,
- authority-filtered `GameSession.BuildingChoices` as the shared legal-action surface for future UI and AI,
- server-owned Seat -> Empire authorization,
- deterministic construction/domain events and tactical BattleSession boundaries.

The research runtime now uses RP-native `float64` progress (`progress_rp`) rather than reproducing 1996 integer storage constraints. Fractional Colony research is preserved through Empire aggregation; rounding happens only at explicit rule boundaries such as the discrete breakthrough-percent roll. `GameSession.ResearchChoices` exposes the authority-filtered research frontier, while `empire.select_research` accepts only a TechField ID and materializes Technology IDs/keys on the server. Research/Observer events carry both original numeric IDs and speaking keys such as `Technology 155 (research_laboratory)`. Creative/Uncreative acquisition semantics, active-project switching and hyper-advanced repeated-field costs remain later research slices.

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
