# Work slice protocol

This directory defines the recovery-safe workflow for focused Master of Orion X implementation slices.

Closed slices are audited in docs/slices/HISTORY.md; the current/next objective remains authoritative in docs/research/ACTIVE_RESEARCH.md.

## Open-slice marker

An unfinished slice is represented by a file named:

```text
docs/slices/_OPEN_<AREA>_<SLUG>_<YYYY-MM-DD>.md
```

The leading `_OPEN_` is intentional: it is the repository-visible signal that work was started but the slice is not closed yet.

By default there should be exactly one `_OPEN_*.md` marker. Do not start a later slice while an open marker exists unless parallel work was explicitly intended.

The marker is a compact recovery/checklist file. Long-lived evidence, reverse-engineering findings, architectural decisions and final results belong in a permanent domain document, for example `docs/research/<SLICE>_<DATE>.md`.
## Planned-slice specifications

Future work may be prepared without opening it. Prepared queue entries use:

```text
docs/slices/PLANNED_<NN>_<SLUG>.md
```

`PLANNED_*.md` files are **not** recovery markers. Multiple planned specifications may coexist and none of them counts as active work. They may contain the complete intended Gate 1-4 checklist, dependencies, scope guard and exit criterion so a later chat can start from an explicit contract rather than reconstructing the roadmap.

When a planned slice actually starts:

1. verify there is no existing `_OPEN_*.md` marker;
2. read the planned specification and current `docs/research/ACTIVE_RESEARCH.md`;
3. create exactly one dated `_OPEN_<AREA>_<SLUG>_<YYYY-MM-DD>.md` recovery marker referencing the planned specification;
4. create/update the permanent `docs/research/<SLICE>_<DATE>.md` evidence document during Gate 1;
5. follow the normal four-gate protocol below.

Do not mark Gate 1 complete merely because a `PLANNED_` checklist exists. Gate 1 always requires a fresh repository checkup and original-evidence review at the time work begins.

### Current active / prepared queue - 2026-09-03

Active:

- `PLANNED_14_LIVE_GAMESESSION_SAVE_RESUME_BASELINE.md` - **closed; Gates 1-4 complete**. Implementation: `a2c4ccd`; Gate-4 rejection coverage: `e6a5e69`.

Prepared queue:

1. `PLANNED_15_PLAYABLE_BROWSER_STRATEGIC_HMI_LOOP.md` - parent epic for the first saveable Human-vs-AI browser match.
   - `PLANNED_15_1_UX_NAVIGATION_DESIGN_SYSTEM.md` - UX architecture, navigation, responsive shell and design-system contract.
   - `PLANNED_15_2_FUNCTIONAL_STRATEGIC_GAMEPLAY_HMI.md` - functional authoritative galaxy/colony/research/fleet/diplomacy gameplay HMI.
   - `PLANNED_15_3_VISUAL_IDENTITY_GRAPHICS_ASSET_PIPELINE.md` - MOOX corporate/visual identity plus graphics/image/model asset pipeline.
   - `PLANNED_15_4_RICH_GAMEPLAY_PERSISTENCE_UX.md` - rich Battle/Invasion/decision presentation plus save/load/reconnect UX.
   - `PLANNED_15_5_FULL_BROWSER_VERTICAL_SLICE.md` - final complete-game browser integration, polish and end-to-end QA.
2. `PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md` - widen narrow New Game settings and preset-race support.
3. `PLANNED_17_MILITARY_SHIP_DESIGN_COMPONENT_WEAPON_BREADTH.md` - multi-hull/component/weapon design baseline for later tactical depth.

Milestone after Slice 15.5: a human can start, save/resume and finish the supported Human-vs-built-in-AI match from the browser through one coherent MOOX visual/interaction language. Re-audit after Slice 17 before numbering Tactical/Diplomacy/Spies/Leaders/alternative-victory depth.

The queue order is the current recommendation, not a hard dependency lock. Reprioritization is allowed before a later slice is opened if the dependency check still passes.
## Four gates

Every slice uses the same four gates.

### 1. Checkup + original analysis / reverse engineering

- inspect Git status, current runtime/data/docs and the previous checkpoint before changing behavior;
- inspect original MOO2 executable/data/save evidence where the slice concerns original behavior;
- distinguish directly proven behavior from inference and deliberate MOOX modernization;
- record findings, addresses/offsets/provenance, unknowns and blockers in the permanent slice document;
- update the `_OPEN_` marker;
- **stop before implementation and present the findings plus the proposed implementation shape for discussion.**

This gate combines the repository checkup and the original/reverse-engineering investigation.

### 2. Implementation decision

- discuss the implementation shape with the user;
- record the accepted state/data/API/rule design and any deliberate divergence from original MOO2;
- mark this gate complete only after that decision is clear.

No gameplay implementation should begin before this gate is complete.

### 3. Implementation

- implement only the agreed slice;
- add/update deterministic tests and normalized data where required;
- keep authoritative gameplay logic in core/game/session layers rather than UI/network adapters;
- keep the permanent slice document synchronized with material changes.

### 4. Follow-up QA + commit + close

Run the relevant final gates, normally including:

```text
gofmt on changed Go files
go test ./...
go vet ./...
git diff --check
```

Also run slice-specific generators/analyzers/regressions when applicable.

Then:

- record final behavior, QA and the commit subject/result in the permanent slice document;
- update `docs/research/ACTIVE_RESEARCH.md` to the next objective;
- delete the completed slice's `_OPEN_*.md` marker in the closing commit;
- create the next `_OPEN_*.md` marker only when the next slice actually starts.

A completed slice therefore leaves a permanent evidence/result document but no `_OPEN_` marker.

## Recovery rule

At the start of work:

1. inspect `git status` without discarding unknown changes;
2. check `docs/slices/_OPEN_*.md`;
3. if an open marker exists, resume that slice first and read its permanent document;
4. only if no open marker exists, use `docs/research/ACTIVE_RESEARCH.md` to choose the next slice.

This convention is specifically intended to make interrupted chats or agent handoffs obvious and recoverable.
