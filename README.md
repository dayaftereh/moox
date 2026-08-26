# MOOX

MOOX is a new game project whose gameplay reference is **Master of Orion II: Battle at Antares (1996)**.

The initial goal is a high-fidelity reimplementation of the systems that make MOO2 work: turn-based 4X galaxy strategy, colony management, custom race design, research, diplomacy/espionage, ship design, tactical combat, leaders, random events, Orion/Antaran end-game content and the original pacing/feedback loops.

## Project status

Research/bootstrap phase started 2026-08-26.

The repository was empty except for `.git` before this baseline. This first commit intentionally contains **research and project structure only**, not an implementation yet.

## Research documents

- `docs/research/MOO2_GAME_REFERENCE.md` - gameplay/system reference and fidelity checklist.
- `docs/research/MOO2_DATA_AND_FILE_FORMATS.md` - original data formats, LBX research and extraction strategy.
- `docs/research/SOURCES.md` - curated source register with links and usage notes.
- `docs/IMPLEMENTATION_PLAN.md` - proposed clean-room development phases.
- docs/ANALYZER.md - pure-Go MOO2 console analyzer usage and architecture.
- docs/architecture/ADR-0001-go-wails-v3.md - Go/Wails v3 portability decision and layer boundary.
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

Before picking the rendering/UI technology, build a **deterministic headless simulation specification** and normalize the reference data we need. The first playable vertical slice should be:

- generate a small galaxy,
- create one empire and homeworld,
- assign population to farmer/worker/scientist roles,
- run turns,
- produce food/industry/research/money,
- research one technology,
- build a colony ship,
- move it to another system,
- colonize a planet,
- save and reload the exact game state.

After this baseline is stable we can add race design, diplomacy, ship design and tactical combat without coupling game rules to the UI.

