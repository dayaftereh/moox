# Planned slice 16.1 - New Game visual selection grammar

Status: **open; Gate 1 complete (2026-09-15), Gate 2 not started**.

Parent: `PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`.

## Objective

Replace the current form-like New Game setup with one coherent, image-led selection grammar that can be reused by the later Slice-16 settings: difficulty, galaxy settings, race selection, starting technology level and opponent/player composition.

The visual target is a **carousel / previous-next selector** rather than a dense wall of dropdowns. The user should always understand the currently selected option from a strong central visual, short title and a few concise facts, with clear left/right arrows to browse alternatives.

## Binding UX direction

- Desktop and mobile use the same conceptual selector.
- Current option is the visual focus in the center.
- Previous/next arrows are first-class controls and must work with mouse, touch and keyboard.
- Swipe may be added on mobile, but arrows remain visible and authoritative.
- Each selector has:
  - one primary visual;
  - one short title;
  - at most a few compact facts underneath;
  - a stable option position indicator where useful;
  - accessible text/ARIA that does not depend on the image.
- Changing an option must never silently mutate other settings unless the server contract explicitly requires it.
- The server remains authoritative for which option combinations are valid; React only presents and edits the draft.
- The visual selector must not make unsupported settings look enabled.

## Shared visual language

The New Game setup should feel like part of the same MOOX product as Galaxy, Colony, Ship Designer and Tactical:

- dark-space card surfaces;
- strong central illustrations;
- compact luminous selection frame;
- large previous/next chevrons inspired by the Ship Designer navigation controls;
- small factual strip below the image rather than long prose;
- no stock imagery and no copied Master of Orion II artwork.

## Asset policy

Slice 16 may introduce a dedicated New Game asset manifest. Every asset must have a stable semantic ID and deterministic repository path.

Use the asset type that best matches the content:

- **SVG**: frames, arrows, emblems, abstract setting icons, galaxy diagrams and simple symbolic illustrations.
- **PNG/WebP raster art**: painterly/organic race portraits and other artwork where a full SVG would be artificial or expensive.
- Prefer keeping an editable/high-resolution source or generation recipe in the repository workflow even when the shipped runtime asset is optimized.

Generated artwork must be original MOOX artwork. Genre/classic-space-4X inspiration is allowed; tracing, copying or reusing original MOO2 portraits/UI art is not.

## Layout baseline

Desktop candidate:

```text
< previous      [ large selected artwork ]      next >
                OPTION TITLE
         fact 1  |  fact 2  |  fact 3
```

Mobile candidate:

```text
      [ large artwork ]
<        OPTION        >
      compact facts
```

The selector must avoid horizontal page overflow at the established mobile QA width (390 px CSS viewport).

## Gate 1 - audit / prototype

- [x] Inventory the current New Game form and its server-authoritative settings schema.
- [x] Audit existing Ship Designer arrow/carousel interaction components for reuse.
- [x] Prototype one selector with placeholder/original art at desktop and 390 px mobile.
- [x] Verify keyboard, mouse and touch-size controls.
- [x] Define a stable New Game asset manifest/path convention.
- [x] Confirm which settings can share one generic selector and which require specialized cards.

## Gate 2 - freeze

- [ ] Freeze the selector component contract and responsive layout.
- [ ] Freeze accessibility/keyboard behavior.
- [ ] Freeze asset naming/runtime format conventions.
- [ ] Freeze unsupported/disabled-state presentation.

## Gate 3 - implementation

- [ ] Implement the shared visual selector foundation.
- [ ] Migrate Slice-16 setting screens to the common grammar as their sub-slices land.
- [ ] Add deterministic UI tests where practical and browser-visible QA.

## Gate 4 - close

- [ ] Desktop and 390 px mobile browser acceptance.
- [ ] Keyboard/mouse/touch interaction smoke.
- [ ] No horizontal overflow or hidden required facts.
- [ ] Full web build and repository diff checks.

## Exit criterion

Slice 16.1 is complete when New Game has one reusable, responsive and accessible image-led previous/next selection grammar that all later 16.x settings can use without inventing separate UX patterns.
