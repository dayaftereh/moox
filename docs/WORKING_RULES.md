# MOOX working and research rules

These rules exist to keep long reverse-engineering sessions convergent. They are mandatory for MOOX research work unless a task explicitly requires a different workflow.

## 1. One active objective

There is exactly one active research objective at a time.

The objective must be narrow enough to finish, block, or checkpoint independently. Good example:

> Determine the original `SHIPS.LBX` strategic block-index formula from the 1.31 executable.

Bad example:

> Understand ships.

The authoritative current objective is `docs/research/ACTIVE_RESEARCH.md`.

## 2. Start every research session with a preflight

Before opening a new investigation path:

1. read `docs/research/ACTIVE_RESEARCH.md`,
2. inspect `git status --short --branch`,
3. inspect the last relevant commits,
4. run `tools/research-preflight.ps1` when working on the local checkout,
5. confirm that the next action directly advances the active objective.

Do not rediscover a closed subsystem merely because the current conversation is long or has lost context.

## 3. Dirty-tree quarantine

Existing unrelated changes are not part of the current task unless `ACTIVE_RESEARCH.md` explicitly says so.

When unrelated dirty/untracked files exist:

- record them in `ACTIVE_RESEARCH.md`,
- do not modify them unless they are in scope,
- do not stage them,
- do not delete them as cleanup,
- do not include them in a checkpoint commit.

A checkpoint commit must list/stage only files belonging to its objective.

## 4. Exploration budget

An investigation path gets at most **10 consecutive search/inspection actions** without materialized progress.

Materialized progress means at least one of:

- a verified factual finding added to a research note,
- normalized data/code added or improved,
- a regression/unit test that captures the finding,
- a clearly documented blocker with the failed hypothesis and next alternative.

If 10 research actions produce none of these, stop that path. Do not continue with "one more broad search". Record the blocker and switch to a different, narrower method.

The counter resets when a real artifact/finding is materialized, not merely when a new command is run.

## 5. Partial knowledge is committed explicitly

Do not wait for an entire subsystem to be understood.

When only part of a mapping is proven:

- proven facts become `confirmed` / `original-observed` / equivalent,
- evidence-backed but indirect facts remain explicitly tagged as such,
- unknown values remain `pending`,
- no slot/name/formula is guessed simply to make a table complete.

A partially complete but well-proven dataset is preferable to an all-or-nothing research branch.

## 6. Evidence hierarchy must remain visible

Use the strongest available evidence and preserve its provenance.

Preferred order:

1. **original-observed**: directly measured in the local 1.31 data/executable,
2. **original-derived**: deterministic formula derived from hash-verified original code/data,
3. **secondary-verified**: external implementation/reference, independently cross-checked where possible,
4. **pending**: unresolved or insufficiently proven.

Never silently promote a secondary inference to an original fact.

## 7. Closed areas stay closed

A subsystem becomes **closed** when:

- the relevant normalized artifact/code exists,
- provenance is documented,
- tests pass,
- the state is committed.

A closed subsystem may only be reopened when:

- new evidence contradicts it,
- a downstream test exposes an inconsistency,
- the active objective explicitly requires an extension of it.

Do not revisit a closed area merely to regain context.

## 8. Drift detection

Before starting a new branch of investigation, ask one question:

> Does this directly advance `Next exact action` in `ACTIVE_RESEARCH.md`?

If not, do not do it yet. Add it to `Later / follow-up` instead.

When a useful but unrelated discovery appears, record it briefly and return to the active objective.

## 9. Checkpoint size

Prefer small evidence-complete commits over subsystem-sized commits.

Typical reverse-engineering checkpoints should look like:

- identify original table/layout,
- normalize identities,
- prove one index formula,
- map one asset family,
- add one behavioral formula,
- document one blocker.

Each stable checkpoint should include the relevant tests and documentation.

Do not wait for "all ships", "all planets", or another large subsystem before committing.

## 10. Required checkpoint gate

Before a normal checkpoint commit:

1. `gofmt` changed Go files,
2. run focused tests while developing,
3. run `go test ./...` before the checkpoint,
4. run `git diff --check`,
5. inspect `git status --short --branch`,
6. stage only in-scope files,
7. commit with a focused message,
8. update `ACTIVE_RESEARCH.md` to the next objective/action.

If the current work is intentionally blocked, commit documentation only if that blocker is useful and stable; otherwise leave the repo unchanged and record the blocker in `ACTIVE_RESEARCH.md` before switching approaches.

If go test ./... is prevented solely by **pre-existing quarantined dirty files that the current checkpoint must not modify**, record the exact failure in ACTIVE_RESEARCH.md, run the broadest unaffected package test set, and state the exception explicitly in the checkpoint. Never delete, stage or rewrite the unrelated files just to obtain a green global test.

## 11. ACTIVE_RESEARCH.md is the recovery contract

`docs/research/ACTIVE_RESEARCH.md` must always answer:

- What is the single active objective?
- What is explicitly in scope?
- What is explicitly out of scope?
- What facts are already proven?
- What is the next exact action?
- What is the exploration budget state?
- What dirty files already existed?
- What areas are closed and should not be revisited?
- What was the last good commit/checkpoint?

When a conversation is restarted or context is lost, this file is read before broad project exploration.

## 12. Blocker rule

A blocker is a valid result.

When blocked, write down:

- the exact missing fact,
- the attempted evidence path,
- why it failed,
- what was ruled out,
- the next narrower alternative.

Then stop repeating the failed path.

## 13. No hidden scope expansion

Research, normalized gameplay data, asset semantics, localization and engine implementation are separate deliverables.

Discovering a useful adjacent task does not automatically make it part of the active objective. Add it to follow-up work and keep moving on the current ticket.

## 14. Definition of done for a research ticket

A research ticket is done when one of these is true:

### Completed

- target fact/formula/mapping is proven to the stated evidence level,
- resulting code/data/docs/tests are materialized,
- tests/checks pass,
- a focused checkpoint exists.

### Blocked

- the missing evidence is precisely documented,
- repeated broad exploration has stopped,
- a next alternative is recorded,
- no unsupported conclusion is committed.

Either state is progress. Endless exploration is not.
