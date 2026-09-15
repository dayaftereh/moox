# Planned slice 16.1 - New Game visual selection grammar

Status: **closed; Gates 1-4 complete (2026-09-15)**.

Parent: `PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`.

## Objective

Replace the current form-like New Game setup with one coherent, image-led selection grammar that can be reused by the later Slice-16 settings: difficulty, galaxy settings, race selection, starting technology level and opponent/player composition.

The visual target is a **carousel / previous-next selector** rather than a dense wall of dropdowns. The user should always understand the currently selected option from a strong central visual and short title, with clear left/right arrows to browse alternatives. Detailed facts belong behind the reusable artwork info control so the card stays compact.

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

- [x] Freeze the selector component contract and responsive layout.
- [x] Freeze accessibility/keyboard behavior.
- [x] Freeze asset naming/runtime format conventions.
- [x] Freeze unsupported/disabled-state presentation.

Permanent freeze evidence: `docs/research/SLICE_16_1_GATE2_FREEZE_2026-09-15.md`.

## Gate 3 - implementation

- [x] Implement the shared visual selector foundation.
- [x] Migrate the Slice-16 reference setting to the common grammar and bind later setting migrations to their sub-slices.
- [x] Add deterministic UI tests where practical and browser-visible QA.

Permanent implementation evidence: `docs/research/SLICE_16_1_GATE3_IMPLEMENTATION_2026-09-15.md`.

Downstream Difficulty/Galaxy Age/Race/Starting Tech/Player Composition breadth remains owned by Slice 16.2-16.6 and must reuse this foundation rather than being pre-built inside 16.1.

## Gate 4 - close

- [x] Desktop and 390 px mobile browser acceptance.
- [x] Keyboard/mouse/touch interaction smoke.
- [x] No horizontal overflow or hidden required facts.
- [x] Full web build and repository diff checks.

Permanent closure evidence: `docs/research/SLICE_16_1_GATE4_CLOSE_2026-09-15.md`.

## Exit criterion

Slice 16.1 is complete when New Game has one reusable, responsive and accessible image-led previous/next selection grammar that all later 16.x settings can use without inventing separate UX patterns.
