# Open slice: Population capacity transitions

Status: Gate 3 complete; Gate 4 QA / commit / close in progress.

## Objective

Establish original MOO2 1.31 Population-capacity behavior for Biospheres, Advanced City Planning and terraforming/climate transitions before changing authoritative gameplay code.

## Gate checklist

- [x] Gate 1: checkup + original analysis / reverse engineering
- [x] Gate 2: implementation decision
- [x] Gate 3: implementation
- [ ] Gate 4: follow-up QA + commit + close (in progress)

## Gate 1 questions

1. Exact capacity effect and application order for Biospheres.
2. Exact Advanced City Planning capacity effect and ownership/application semantics.
3. Exact terraforming/climate-transition effect on current/max Population.
4. Interaction with Aquatic, Tolerant and Subterranean capacity rules already normalized.
5. Whether capacity decreases can force immediate Population loss or only clamp future growth.
6. Minimal normalized Building/Technology/runtime identities and deterministic Core/Game/Session test shape.

## Recovery

Starting branch: `main`.
Starting HEAD: `f9f4792` (`docs: close population growth modifier slice`).
Starting tree: clean, `main` ahead of `origin/main` by 8 commits.
Previous gameplay slice: Population growth building / medicine modifiers (`e7f0d72`), closed.
Core schema at slice start: 12.

Gate 1 is complete. Review `docs/research/POPULATION_CAPACITY_TRANSITIONS_2026-08-29.md` and accept/revise the proposed Gate 2 implementation shape before gameplay implementation.

## Gate 2 decision

Accepted by the user on 2026-08-29. Proceed with the proposed semantic design from docs/research/POPULATION_CAPACITY_TRANSITIONS_2026-08-29.md.
