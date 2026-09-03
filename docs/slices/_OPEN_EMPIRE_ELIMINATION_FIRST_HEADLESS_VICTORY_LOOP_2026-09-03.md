# Open slice 12 - Empire elimination and first headless victory loop

Status: **Gate 3 complete; Gate 4 QA complete; implementation/evidence commit + final closure pending**.

Opened: **2026-09-03**

Planned specification: `docs/slices/PLANNED_12_EMPIRE_ELIMINATION_FIRST_HEADLESS_VICTORY_LOOP.md`

Permanent evidence: `docs/research/EMPIRE_ELIMINATION_FIRST_HEADLESS_VICTORY_LOOP_2026-09-03.md`

Starting HEAD: `e2979c1` (`docs: close troop transport invasion conquest slice`)

## Gate state

- [x] Gate 1: current-code checkup + original MOO2 elimination/victory research + deterministic public-command scenario contract.
- [x] Gate 2: accept/freeze exact elimination/game-over/session/transport contract.
- [x] Gate 3: implementation + deterministic headless match regressions.
- [ ] Gate 4: independent QA + commits + close.

## Scope guard

Gate 1 is research/integration analysis only. Do not implement victory/game-over gameplay before Gate 2 freezes the contract.

In scope for eventual implementation: authoritative conquest-elimination predicate; completed GameSession result; final-turn ordering; read-only post-game behavior; player/observer/server/web projection; public-command-only NewGame-to-winner scenario; save/load/replay equality.

Deferred: Galactic Council, Orion/Antaran and score/time-limit victory; surrender unless required by original elimination semantics; competitive AI; polished victory UI/cinematics.

## Gate 1 checklist

- [x] Re-check Slice 08-11 contracts and GameSession phase/state model.
- [x] Create/update permanent first-victory-loop evidence document.
- [x] Verify original MOO2 Empire elimination/end-of-game trigger.
- [x] Verify whether surviving Fleets/Outposts without Colonies delay elimination.
- [x] Verify final ownership/turn-settlement ordering relative to victory detection.
- [x] Verify minimum final winner/loser/result representation for projections/replay.
- [x] Define deterministic public-command-only NewGame -> final conquest scenario.
- [x] Inventory remaining systems as fidelity/depth vs match-lifecycle blockers.
- [x] Present exact Gate-2 contract before implementation.
