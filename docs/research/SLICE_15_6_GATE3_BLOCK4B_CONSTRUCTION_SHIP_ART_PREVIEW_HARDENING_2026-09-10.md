# Slice 15.6 Gate 3 Block 4B - Construction ship art and planning-preview hardening

Date: 2026-09-10
Status: **implemented and isolated-browser QA passed; canonical service refresh pending**

## User feedback

- Colony Construction catalog showed saved Military Ship designs only as a generic ship icon instead of the exact procedural design artwork.
- The browser could surface a global `Planning preview unavailable` warning during lifecycle/revision churn.

## Changes

- Construction catalog Military Ship rows now resolve the exact player-safe `ship_design_id` + `ship_design_revision` against `decision.strategic.ship_designs` and render its `ProceduralShipGlyph` with the persisted visual genome, hull and weapon count.
- The selected Military Ship project hero uses the same exact design artwork rather than a generic icon.
- Building/housing/transformation rich artwork remains unchanged; non-rich non-military projects retain their normal icons.
- Planning preview now records the game/turn/revision that started the request. A result is discarded when the current snapshot has already moved beyond that identity.
- Lifecycle-only `session_rejected` preview conflicts (base revision/turn/submitted/phase moved) are treated as obsolete read-only previews rather than user-facing failures. Genuine invalid draft/internal preview errors remain visible.

## Validation before service refresh

- `git diff --check` passed.
-
pm run build` passed (`tsc -b` + Vite production build).
- Main JS remains below the frozen 350 KiB single-chunk guardrail.

## Browser QA evidence

Managed Chrome against isolated 7181 / game `qa-52d5f74` verified:

- baseline Scout catalog row renders an exact `ProceduralShipGlyph` at 68x48 with hull `frigate`;
- selected Scout hero renders the same design at 432x190;
- planning Scout shows 25.0 PP remaining / 5 turns with no Planning Preview warning;
- a new design was created entirely through visible Ship Designer UI as `Laser Test Mk I`, rerolled visually once, equipped with 1x Laser, and saved at 30 PP / 1 CP / 10 of 25 space;
- returning to Construction shows both Scout and Laser Test Mk I with distinct persisted SVG genomes (Scout `spear/chevron`, Laser Test Mk I `organic/needle` in this run);
- selecting Laser Test Mk I renders its exact persisted `organic/needle` hero;
- planning it shows 30.0 PP remaining / 5 turns and no `Planungsvorschau nicht verfügbar` warning.

The temporary browser was closed and the 7181 QA server was stopped.

## Remaining acceptance

Restart canonical 7171 from the resulting commit and verify in managed Chrome:

1. Construction catalog shows exact SVG thumbnails for both the baseline Scout and saved Laser design.
2. Selecting a Military Ship shows the same design in the project hero.
3. `Einplanen` produces authoritative remaining PP / ETA.
4. No `Planungsvorschau nicht verfügbar` warning appears in the normal route.
5. Other construction artwork remains intact.
