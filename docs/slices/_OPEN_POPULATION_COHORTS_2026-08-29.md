# Open slice: Race-aware Population cohorts

Status: Gates 1-3 complete; Gate 4 follow-up QA + commit + close pending.

## Objective

Establish original MOO2 1.31 Population cohort semantics for native, conquered and mixed-race Colonies, including assignment, race-specific capacity, conquest/assimilation transitions and the remaining Food-priority passes that depend on cohort/status Population.

## Gate checklist

- [x] Gate 1: checkup + original analysis / reverse engineering
- [x] Gate 2: implementation decision
- [x] Gate 3: implementation
- [ ] Gate 4: follow-up QA + commit + close

## Gate 1 questions

1. Original Colony Population-entry/cohort storage and authoritative identity fields.
2. Relationship between race/status cohort entries and Farmer/Worker/Scientist assignment.
3. Growth/starvation/transfer behavior with multiple races/status groups in one Colony.
4. Race-specific capacity and removal priority when capacity falls.
5. Conquest/invasion/ownership/assimilation semantics that mutate cohort state.
6. Exact later original Food-import priority passes/thresholds that require cohort/status Population.
7. Minimal Core/Game/Session schema and deterministic test shape without absorbing unrelated Diplomacy/Fleet/UI/Treasury work.

## Recovery

Starting branch: `main`.
Starting HEAD: `5ce3c99` (`docs: close population capacity transition slice`).
Starting tree: clean, `main` ahead of `origin/main` by 10 commits.
Previous gameplay slice: Population capacity transitions (`25d05a9`), closed.
Core schema at slice start: 13.
Economy ruleset schema at slice start: 7.

Gate 1 is complete. Gate 2 was accepted on 2026-08-29 without scope changes. Gate 3 now implements that contract with Core schema 14 authoritative organic cohorts, race-aware Economy/Population dynamics, exact four-pass Food allocation, heterogeneous capacity and cohort-aware transfer/session/observer behavior. `go test ./...` passes; Gate 4 remains the formal final QA/commit/close step.
