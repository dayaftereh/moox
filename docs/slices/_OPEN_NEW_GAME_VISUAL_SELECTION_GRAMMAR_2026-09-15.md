# Open Slice 16.1 - New Game visual selection grammar

Opened: 2026-09-15
Planned specification: `docs/slices/PLANNED_16_1_NEW_GAME_VISUAL_SELECTION_GRAMMAR.md`
Permanent Gate-1 evidence: `docs/research/SLICE_16_1_GATE1_NEW_GAME_VISUAL_SELECTION_GRAMMAR_2026-09-15.md`

## Recovery state

- Branch: `main`.
- Opened from clean synchronized HEAD `4720219`.
- Gate 1 is complete; Gate 2 is active / review candidate.
- No gameplay/server contract changes have been made yet.

## Gate 1 checklist

- [x] Inventory current New Game form and server-authoritative settings schema.
- [x] Audit existing Ship Designer arrow/carousel interaction components for reuse.
- [x] Prototype one selector with placeholder/original art at desktop and 390 px mobile.
- [x] Verify keyboard, mouse and touch-size controls.
- [x] Define a stable New Game asset manifest/path convention.
- [x] Confirm which settings can share one generic selector and which require specialized cards.

## Current findings

- `web/src/App.tsx` still presents New Game as a form and submits fixed Slice-09 values.
- `web/src/api.ts` narrows the client request type to the same fixed baseline.
- `internal/game/new_game.go` remains authoritative and rejects unsupported galaxy size/age, technology level, strategic combat and player/race combinations.
- Ship Designer navigation uses local `stepHull` logic plus `.shipdesigner-arrow`; there is no reusable generic carousel component yet.

## Next

Prototype the shared visual-selector shell without pretending unsupported server choices are enabled. Reuse the proven arrow sizing/focus language, but extract a New-Game-neutral component contract rather than coupling to Ship Designer state.
