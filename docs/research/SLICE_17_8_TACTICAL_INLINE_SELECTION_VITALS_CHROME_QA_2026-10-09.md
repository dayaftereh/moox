# Slice 17.8 – In-square selected ship condition microbars (Chrome QA)

Date: 2026-10-09
Status: implemented in User Review, manual approval pending.

## Request

The previous three large colored percentages were displayed on a floating backplate above the selected active ship and obscured the battlefield. Replace them with three **tiny charge strips completely inside** the existing blue 1×1 ship movement-selection square: a slightly stronger blue Shield strip along the **top** and thin **yellow Armor** and **red Structure** strips inside the **bottom** of the square. No visible percent numbers.

## Implementation

- Retains `TacticalSelectedShipVitals`, its authoritative calculations, accessible aria-label, per-row `data-percent`, and `movementSelectionActive && selectedShip` visibility gate.
- The bars all have x = -0.44 and width 0.88, within the square with local ship coordinates x/y ±0.5.
- **Shield:** y -0.455, height 0.105, computed from sum of all 4 `shield_arc_current` / (`shield_arc_max * 4`).
- **Armor/Hull:** y +0.318, height 0.055, `armor_current / armor_max`.
- **Structure:** y +0.404, height 0.055, `structure_current / structure_max`.
- `meterPercent` clamps 0..100 before composing the filled part. Full gray/dark tracks remain at zero. Colored widths are proportional to actual percentages.
- Removed large 2.86×1.57 backplate, floating placement and visible percentage text; accessibility percentages remain in the SVG aria-label.
- SVG pointer events remain disabled. Existing blue selection square, green legal movement grid, camera follow/ownership, shield sectors and battle rules are unchanged.

## Real Chrome verification (isolated 7174)

- Fresh Triangle combat review: exactly one 1×1 blue selection square and one three-strip group attached to own active Falcon MK1 #40; all three fill widths 0.88 at 100% shield/armor/structure; blue/yellow/red computed colors are rgb(104,216,255), rgb(255,196,94), rgb(255,90,87).
- The SVG overlay bounding box (approx. 22.47×23.34 screen px) lies **entirely inside** the blue square (approx. 25.54×25.54 screen px). All three colored stripes and their dark tracks fit between -0.5 and +0.5 in both axes. No `tactical-vitals-backplate` or SVG text nodes exist.
- Actual click-to-move from (25,9) to authorized (26,9) generated authoritative `ship_moved` event sequence 2, remaining movement 15. The three strips moved with Falcon #40, remained inside the selection square, and did not intercept pointer input.
- Firing a real Heavy Particle Beam on Darlok #45 produced presentation lock. While locked, both the blue selection square and all three strips were absent; they returned immediately after unlock.
- Finishing own ship #40 selected Falcon #41 and correctly moved the group to that new selected ship. Other own and enemy ships never showed extra status groups.
- After actual AI tactical attacks, the authoritative state of Falcon #40 was [0,100,100,100] shield (300/400), Armor 0/500, Structure 12/50. In animation lock the overlays were hidden; after unlock they should show 75%, 0%, 24% respectively in data fields, with visible colored widths of 0.66, 0, and 0.2112 (of 0.88 units). Real Chrome confirmed the unlocked state showed 75%, 0%, 24% in per-row data-percent attributes, visible fill widths **0.66**, **0** and **0.2112** (of 0.88), with all three rows fully within the same blue selected-cell boundary. No percentage text nodes were rendered.

## QA

- `web/npm run build` incl. TypeScript, Tactical HMI contract and Vite: PASS.
- `go test ./internal/battle ./internal/server -count=1`: PASS.
- `git diff --check`: PASS.
- New Slice 17.9: NOT opened. Slice 17.8 remains in User Review.
- Canonical 7171: Not reset/restarted; tests only used the isolated 7174 reference game.
