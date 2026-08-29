# Open slice: Canonical strategic Fleet/Diplomacy blockade production

Status: Gates 1-3 complete; Gate 4 QA + commit + close pending.

## Objective

Establish the minimum original-faithful strategic Fleet and diplomacy/relationship state required to derive `StarSystem.BlockadedEmpireIDs` deterministically for the existing Food-logistics and Population-transfer consumers, without absorbing tactical combat or broad diplomacy scope.

## Recovery

- Starting branch: `main`.
- Starting HEAD: `41bfe92` (`docs: close population cohort slice`).
- Starting tree: clean; `main` ahead of `origin/main` by 12 commits.
- Previous gameplay slice: Race-aware Population cohorts (`7c49e90`), closed.
- Core `StateSchemaVersion`: 14.
- Economy ruleset schema: 7.
- Existing consumer evidence: `docs/research/BLOCKADE_FOOD_LOGISTICS_2026-08-28.md`.
- Permanent Gate 1 evidence: `docs/research/FLEET_DIPLOMACY_BLOCKADES_2026-08-29.md`.

## Gates

- [x] Gate 1: repository checkup + original MOO2 1.31 reverse engineering
- [x] Gate 2: implementation decision
- [x] Gate 3: implementation
- [ ] Gate 4: QA + commit + close

## Gate 1 result

Direct MOO2 1.31 evidence resolves the minimum producer boundary:

- target presence is Colony-owner presence at the System;
- only stationary ordinary/combat strategic Fleets participate;
- Fleet owner/system are sufficient blockade-producer identities;
- the normal relation test is directed from Fleet owner to target Colony owner and accepts original relation values 4..6;
- multiple qualifying hostile Fleet owners combine by bit/set semantics;
- existing `StarSystem.BlockadedEmpireIDs` is the correct semantic target-state output;
- original timing places blockade recomputation between the first Colony/Food pass and Settler movement/final Colony/Food pass.

## Proposed Gate 2 contract

- Core schema 14 -> 15.
- Add semantic `StrategicFleet { ID, EmpireID, Role, AtSystemID }`.
- Add directed semantic `DiplomaticRelation { FromEmpireID, ToEmpireID, Stance }` with extensible `neutral|hostile` stance for this slice.
- Derive target Empire presence from existing Colony/Planet state; persist no duplicate presence table.
- Add deterministic `recomputeSystemBlockades` that fully replaces sorted/unique `BlockadedEmpireIDs` from current Fleet/relation/Colony state.
- Run it after Construction and before Population-transfer resolution/final Food recalculation.
- No player command, fleet strength, tactical composition, broad diplomacy, full movement model or special/non-player blockade producer in this slice.

Gate 2 accepted this contract on 2026-08-29 without scope changes. Gate 3 implementation is authorized.

## Scope guard

Do not implement gameplay during Gate 1/2 discussion. Do not absorb tactical combat, ship-design detail, fleet strength, full movement/ETA, diplomacy negotiations/treaties, special/non-player fleets, AI strategy or UI work.
