# Open Slice 16.1 - New Game visual selection grammar

Opened: 2026-09-15
Planned specification: `docs/slices/PLANNED_16_1_NEW_GAME_VISUAL_SELECTION_GRAMMAR.md`
Permanent Gate-1 evidence: `docs/research/SLICE_16_1_GATE1_NEW_GAME_VISUAL_SELECTION_GRAMMAR_2026-09-15.md`
Permanent Gate-2 freeze evidence: `docs/research/SLICE_16_1_GATE2_FREEZE_2026-09-15.md`
Permanent Gate-3 implementation evidence: `docs/research/SLICE_16_1_GATE3_IMPLEMENTATION_2026-09-15.md`

## Recovery state

- Branch: `main`.
- Opened from clean synchronized HEAD `4720219`.
- Gate 1, Gate 2 and Gate 3 are complete; Gate 4 is next.
- No gameplay/server contract changes have been made yet.

## Gate 1 checklist

- [x] Inventory current New Game form and server-authoritative settings schema.
- [x] Audit existing Ship Designer arrow/carousel interaction components for reuse.
- [x] Prototype one selector with placeholder/original art at desktop and 390 px mobile.
- [x] Verify keyboard, mouse and touch-size controls.
- [x] Define a stable New Game asset manifest/path convention.
- [x] Confirm which settings can share one generic selector and which require specialized cards.

## Current findings

- The shared `VisualSelector` foundation now exists and is used by the Galaxy Size reference selector.
- The selector contract is compact and image-led: title -> rounded artwork with info control -> previous/current/next -> position dots.
- Detailed facts/status are on demand in the responsive info dialog rather than permanently expanding the card.
- The responsive settings grid stacks one card per row on mobile and allows multiple setting cards at wider widths.
- Semantic New Game asset IDs/paths and runtime format conventions are frozen; Galaxy Size currently ships five deterministic original SVG assets.
- Unsupported visual choices remain browseable but do not bypass the authoritative server contract; Create Game is disabled for unsupported selections.
- The current server still accepts only the existing Small/Normal/Average/two-player baseline, so broader visual options remain contract-honest previews until later implementation slices expand the backend.

## Next

Proceed to Gate 4 closure acceptance. Re-run build plus the reusable New Game selector browser smoke at desktop and 390 px mobile, confirm keyboard/mouse/touch behavior and preserve the frozen server-authoritative supported/planned contract.
