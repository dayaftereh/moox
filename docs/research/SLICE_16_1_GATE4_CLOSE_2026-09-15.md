# Slice 16.1 Gate 4 - Close

Date: 2026-09-15
Status: **complete / Slice 16.1 closed**

## Acceptance decision

The user explicitly approved closing Gate 4 after accepting the Galaxy Size selector/art direction and the shared visual-selection grammar. Slice 16.1 is therefore closed with Gates 1-4 complete.

## Final acceptance checks

### Full web build

`npm run build` passed in the closing run. The build includes:

1. UTF-8/mojibake guard;
2. deterministic New Game selector/asset contract check;
3. TypeScript project build;
4. Vite production build.

The deterministic selector check confirmed all five Galaxy Size SVG assets, semantic paths/manifest entries, reproducible generator output, generic typed selector contract, accessibility hooks, responsive-grid invariant and touch-target invariant.

### Desktop + 390 px browser acceptance

`npm run check:new-game-selector:browser -- http://127.0.0.1:7171` passed in the closing run against the live MOX application.

The reusable headless-Chrome smoke verified:

- 390 px mobile has no horizontal overflow;
- desktop has no horizontal overflow;
- compact artwork stays at the accepted responsive size/radius;
- all five Galaxy Size options load the correct 1200 x 675 SVG;
- Small remains the only currently server-supported Galaxy Size selection;
- unsupported/planned selections keep Create Game disabled;
- mouse previous/next interaction works;
- a real Chrome DevTools touch tap on the mobile next arrow works;
- keyboard ArrowLeft/ArrowRight/Home/End navigation works;
- the artwork `?` information dialog opens with required detailed facts/status available on demand;
- modal focus enters the close control;
- Escape closes the dialog and returns focus to the information control.

### Required information

The compact card intentionally does not show long fact blocks permanently. This is accepted behavior rather than hidden required information: details and supported/planned status remain directly reachable from the visible artwork `?` control using mouse, touch or keyboard focus.

### Repository checks

- `git diff --check` passes before the closing commit.
- Exactly one Slice-16.1 `_OPEN_` marker existed before closure.
- The marker is deleted in the closing commit as required by `docs/slices/README.md`.
- No new Slice-16.2 marker is created during this closure; the next slice is only opened when work on it actually starts.

## Final frozen result

Slice 16.1 leaves the repository with a reusable New Game visual-selection foundation:

- generic typed `VisualSelector<TId>`;
- compact responsive card/grid grammar;
- rounded image-led setting art;
- large current selection between previous/next arrows;
- information-on-demand modal/bottom sheet;
- keyboard/mouse/touch interaction baseline;
- semantic New Game asset IDs/paths and format policy;
- server-authoritative supported/planned presentation;
- deterministic Galaxy Size reference SVG set and generator;
- deterministic build-time contract checker;
- reusable real-browser acceptance smoke.

The final depth of informational copy is not frozen; later sub-slices may enrich content while preserving this interaction/asset contract.

## Downstream ownership

The following remain separate prepared slices and are **not** hidden unfinished Slice-16.1 work:

- Slice 16.2 - Difficulty visual selector;
- Slice 16.3 - Galaxy Size/Age breadth;
- Slice 16.4 - Preset Race portraits/carousel;
- Slice 16.5 - Starting Technology visual selector;
- Slice 16.6 - Player/opponent composition and final New Game integration.

Closing commit subject: `docs: close slice 16.1 visual selection grammar`.
