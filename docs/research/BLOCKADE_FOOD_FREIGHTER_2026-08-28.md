# Blockade effect on Food/Freighter logistics - 2026-08-28

**Slice status:** OPEN - gate 1 pending

**Open marker:** `docs/slices/_OPEN_BLOCKADE_FOOD_FREIGHTER_2026-08-28.md`

**Workflow:** `docs/slices/README.md`

## Goal

Resolve the original MOO2 1.31 blockade eligibility/effect on Colony Food import/export before adding Population transport competition for the shared Freighter pool.

The slice must answer, from original evidence where practical:

1. which Colony/System blockade predicate is used by `Pass_Out_Imports_` and the surrounding resource-calculation path;
2. whether blockade prevents imports, exports, both, or changes Food production/maintenance separately;
3. when blockade state is materialized relative to `Pre_Import_Computing_` and `Pass_Out_Imports_`;
4. the smallest authoritative MOOX Colony/System state required for deterministic Food logistics;
5. which replay/Observer-visible effects are required without moving policy into UI/network adapters.

Population transport is explicitly outside this slice until Food blockade eligibility is independently proven and represented.

## Gate 1 - Checkup + original analysis / reverse engineering

**Status:** pending.

No new original blockade finding has been accepted for this slice yet. The first work step is evidence collection and documentation; implementation is intentionally blocked until the findings are reviewed.

### Evidence log

Pending.

### Known unknowns

- exact original blockade predicate and source state;
- import/export asymmetry, if any;
- materialization timing relative to Food/Freighter allocation;
- whether any effect currently conflated with blockade actually belongs to a separate production/maintenance rule.

## Gate 2 - Implementation decision

**Status:** blocked by gate 1 and user discussion.

The accepted implementation shape will be recorded here before code changes begin.

## Gate 3 - Implementation

**Status:** not started.

## Gate 4 - Follow-up QA + commit

**Status:** not started.

Final checks and commit result will be recorded here before the `_OPEN_` marker is removed.