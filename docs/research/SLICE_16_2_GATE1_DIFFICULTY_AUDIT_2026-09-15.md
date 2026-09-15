# Slice 16.2 Gate 1 - Difficulty audit

Date: 2026-09-15
Status: **Gate 1 complete; implementation intentionally not started**

Planned slice: `docs/slices/PLANNED_16_2_DIFFICULTY_VISUAL_SELECTOR.md`
Depends on: closed Slice 16.1 visual-selection grammar.

## Gate-1 objective

Audit what is actually known about Master of Orion II difficulty, compare it with the current MOOX New Game/runtime contract, decide how the preferred MOOX five-label vocabulary should relate to original naming, and prototype one coherent five-state icon progression without binding invented gameplay numbers.

## Repository / runtime baseline

The repository was clean at Gate-1 start on `main`, HEAD `4f9046d`, with no `_OPEN_` marker. Slice 16.1 was closed and the live 7171 preview/build watcher were healthy.

Current MOOX New Game/server contract has **no authoritative Difficulty field yet**.

Current authoritative New Game breadth still constrains the existing baseline to the previously implemented settings such as Small galaxy, Normal galaxy age, Average starting technology, current combat mode and the current player composition. `web/src/newGameAssets.ts` already reserves the semantic asset domain `difficulty`, but that is only an asset namespace and is not a gameplay contract.

Gate-1 consequence: the frontend must not display numeric difficulty modifiers as though they already exist in MOOX runtime data.

## Original MOO2 public difficulty vocabulary

The original Master of Orion II instruction manual exposes five difficulty settings:

1. `Tutor`
2. `Easy`
3. `Average`
4. `Hard`
5. `Impossible`

Manual-level behavior that is directly supported:

- Tutor is the beginner mode; enemy races are intentionally weak/easy to fight and custom races are unavailable.
- Easy makes other races' production/research proceed more slowly than the player's and makes encountered races substantially friendlier; custom races are unavailable.
- Average is the normal-development baseline for all races.
- Hard is described as the challenge setting, but the instruction-manual text does not expose an exact numeric modifier table.
- Impossible gives opponents significantly accelerated production/research and makes them hostile from the start.

Primary public references checked during Gate 1:

- Original MOO2 instruction-manual text as mirrored by Manualzz / Manualshelf / manuals.plus.
- Local `docs/research/MOO2_GAME_REFERENCE.md` and `BUILTIN_STRATEGIC_AI_BASELINE_2026-09-03.md`, which already record the same qualitative manual evidence.

## Existing local reverse-engineering evidence

The repository already contains original-executable research for a broader global game-setting byte at `0x21CB0`.

Directly documented local findings include:

- default-game setup sets `0x21CB0` to `0`;
- non-default values participate in an Advanced-start chooser filter;
- non-default values also affect NPC/personality-related weighting;
- the original NPC command-point deficit maintenance branch uses `12 - global[0x21CB0]` BC per excess CP rather than the fixed human rate;
- prior research deliberately left the public difficulty meaning/mapping of this broader byte unresolved rather than guessing it.

Relevant local evidence:

- `docs/research/ADVANCED_START_CHOOSER_2026-08-28.md`
- `docs/research/ADVANCED_START_RESEARCH_2026-08-28.md`
- `docs/research/COMMAND_POINTS_SHIP_MAINTENANCE_2026-08-31.md`
- `docs/research/BUILTIN_STRATEGIC_AI_BASELINE_2026-09-03.md`

The legally owned original executables remain available locally under `reference/original/support/files/Orion2.exe` / `ORION95.EXE` for future deeper binary confirmation. They are private reference material and must not be committed.

## Secondary/community numeric tables

The MOO2 1.50 modding documentation exposes a five-row AI bonus table for Tutor/Easy/Average/Hard/Impossible, including food/production/research, income, command-deficit, spy and marine-related values. It also exposes configurable `ai_productivity_bonus` and `ai_income_bonus` tables.

This is useful **secondary reverse-engineering/modding evidence**, but Gate 1 does not promote those values to stock-authoritative MOOX gameplay data because:

- it is a fan-patch/modding layer rather than direct stock-executable proof;
- parts of its numeric interpretation do not line up trivially with the simplified wording of the original instruction manual;
- the local MOOX binary research has not yet normalized the full original difficulty table.

Therefore no numeric modifier is frozen in Gate 1.

## MOOX display-name decision

The preferred MOOX presentation remains:

1. `Easy`
2. `Normal`
3. `Hard`
4. `Very Hard`
5. `Impossible`

Gate-1 decision: these are a **deliberate modern MOOX UX vocabulary**, not a claim that the original game used those five names.

Semantic anchors for later Gate-2 mapping:

- MOOX `Normal` should preserve the original `Average` concept as the no-special-advantage baseline.
- MOOX `Hard` should preserve the original `Hard` concept.
- MOOX `Impossible` should preserve the original `Impossible` top-end identity.
- MOOX `Easy` may use original Easy-like handicap semantics, but the original Tutor restrictions/behavior must not be silently folded into it without an explicit product decision.
- MOOX `Very Hard` has no directly corresponding original display label and must be an explicit MOOX modernization if retained. Its mechanics cannot be invented in the frontend.

This means the five MOOX display labels do **not** map 1:1 by name to the five original labels. Gate 2 must freeze the authoritative server IDs and exact mechanics before any modifier summary text appears in the selector.

## Visual prototype

Gate 1 includes an original five-state command-emblem prototype:

`docs/research/prototypes/SLICE_16_2_DIFFICULTY_ICON_PROGRESSION_2026-09-15.svg`

The prototype deliberately uses one coherent silhouette family and increases visual weight rather than swapping to unrelated symbols:

- Easy: calm open crest, one rank mark;
- Normal: centered balanced shield, two marks;
- Hard: armored/crowned crest, three marks;
- Very Hard: heavier aggressive crest, four marks;
- Impossible: elite alien/power crest with broken outer aura, five marks.

This is a research prototype only. It is not registered in the runtime manifest and is not yet a production asset set.

## Shared Slice-16.1 contract to reuse

When Gate 3 eventually implements the selector, Slice 16.2 must reuse the frozen Slice-16.1 foundation:

- typed `VisualSelector<TId>`;
- compact setting card;
- rounded image-led artwork;
- previous / large current selection / next controls;
- `?` information-on-demand modal/bottom-sheet;
- responsive one-column mobile and multi-card wider layout;
- keyboard/mouse/touch interaction contract;
- semantic New Game asset path/manifest conventions;
- server-authoritative supported/planned state.

## Gate-1 checklist result

- [x] Re-check original difficulty levels and exact gameplay effects.
  - Result: public qualitative semantics are verified; exact stock numeric tables remain partially unresolved and are explicitly documented as such rather than invented.
- [x] Map existing server/runtime difficulty model, if any.
  - Result: none exists yet; only the `difficulty` asset namespace is present.
- [x] Decide whether the five MOOX display labels map 1:1 or represent deliberate modern naming.
  - Result: deliberate modern naming, with Normal/Hard/Impossible semantic anchors and Easy/Very-Hard mechanics still requiring Gate-2 freeze.
- [x] Prototype all five icon/art states in one coherent progression.
  - Result: research SVG prototype created; no runtime integration.

## Gate-1 stop condition / proposal for Gate 2

Per repository workflow, stop before implementation and present the research outcome plus proposed freeze shape.

Recommended Gate-2 freeze topics:

1. Define canonical server IDs for the supported difficulty set.
2. Decide whether Tutor exists as a separate accessibility/tutorial setting or is intentionally omitted from the main five-level MOOX selector.
3. Define the mechanics of MOOX Very Hard explicitly; do not interpolate hidden numbers casually.
4. Confirm which stock-original numeric difficulty effects can be proven from the owned binary / trusted reverse-engineering sources.
5. Freeze concise player-facing consequence text only from authoritative server/catalog data.
6. Freeze the final emblem/icon motif and decide production SVG vs raster artwork.

No gameplay/server implementation was performed in Gate 1.
