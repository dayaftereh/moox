# Open slice: Population growth building / medicine modifiers

Status: Gate 3 complete; Gate 4 QA / commit / close in progress.

## Objective

Establish and implement original MOO2 1.31 behavior for Housing, Cloning Center and medicine/Technology Population-growth modifiers without disturbing the proven deterministic turn-order contract.

## Gate checklist

- [x] Gate 1: checkup + original analysis / reverse engineering
- [x] Gate 2: implementation decision
- [x] Gate 3: implementation
- [ ] Gate 4: follow-up QA + commit + close (in progress)

## Gate 1 result

Direct original evidence is materialized in `docs/research/POPULATION_GROWTH_MODIFIERS_2026-08-29.md`.

Closed findings:

- natural curve and Race Growth stacking;
- Housing identity, capacity guard and Production-dependent percentage-point formula;
- Cybernetic PP-sustenance term used by Housing;
- Microbiotics +25 percentage points;
- Universal Antidote +50 percentage points;
- Cloning Center flat +0.1 Population/turn, separate from the multiplier;
- original 1/1000 Population apply scale and capacity boundary;
- deterministic Core/Game/Session implementation/test shape.

## Gate 2 decision point

Proposed MOOX design:

- semantic persistent `housing` Construction project kind and authoritative queue command;
- normalized growth-effect metadata in economy rules referencing existing `cloning_center`, `microbiotics` and `universal_antidote` identities;
- Race + medicine + Housing add as percentage points to natural growth;
- Cloning Center adds flat +0.1 Population/turn after natural growth;
- use pre-growth `ProductionAvailable`, preserving Research -> Population -> Construction ordering;
- advance Core schema 11 -> 12 for persisted Housing project state;
- recommended fidelity choice: retain the original integer percentage-point floor in the Housing bonus while keeping MOOX Population/Production domain-native floats elsewhere.

Gate 2 accepted by the user on 2026-08-29. Implement the design above.

## Recovery

Resume Gate 3 implementation from `docs/research/POPULATION_GROWTH_MODIFIERS_2026-08-29.md`; then run deterministic Core/Game/Session tests and proceed to Gate 4 QA only after implementation is complete.
