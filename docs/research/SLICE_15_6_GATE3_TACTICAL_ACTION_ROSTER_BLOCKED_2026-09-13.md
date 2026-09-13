# Slice 15.6 Gate3 - Tactical action row, responsive roster and blocked-cell clarity

Status: **IMPLEMENTED / desktop + mobile browser QA green**
Date: 2026-09-13

## User feedback
The Tactical HUD still exposed `Zurück zur Begegnung`, which does not fit the combat contract: once Tactical combat is entered, the fight should resolve through victory/defeat or explicit retreat, not by returning to an unresolved encounter shell. The user also requested one clean action row, a more compact responsive own-ship roster, and clearer blocked battlefield cells.

## Action row
`Zurück zur Begegnung` was removed from `TacticalBattlefield` and from both BattleRouteView call sites. Tactical now exposes exactly four direct action buttons:

- Scannen
- Warten
- Fertig
- Rückzug

Scan moved into the same action row as the three activation controls. Server-projected weapon-slot buttons remain in the same controls card, separate from the four direct action buttons.

This is a UI-only removal; Tactical authority remains unchanged. A running battle still ends only through its existing authoritative combat result or retreat path.

## Responsive own-ship roster
The own-ship strip is now a CSS grid instead of a horizontal flex strip.

Desktop/tablet:
- columns are `auto-fill` with a 76 px minimum, so wider screens automatically fit more cards per row;
- maximum visible height is two 64 px rows plus the row gap;
- further ships scroll vertically.

Mobile (`<=720px`):
- exactly **4 columns**;
- 64 px row height;
- max two visible rows (`134px`);
- more than 8 ships scroll vertically.

Real browser QA:
- 390 px viewport -> computed grid `82.75px 82.75px 82.75px 82.75px` = exactly 4 columns;
- temporary DOM expansion to 9 cards -> `clientHeight=134`, `scrollHeight=204`, so the list is genuinely scrollable after two rows;
- 1400 px viewport -> computed roster grid uses 6 columns in the current HUD allocation, proving larger screens automatically use more horizontal capacity instead of staying locked to four.

## Blocked battlefield cells
Occupied/blocked Tactical cells keep the authoritative 1x1 occupied-cell geometry but now use a translucent red fill:

- fill: `rgba(255, 80, 77, .28)`
- stroke: `rgba(255, 96, 92, .98)`

This makes blocked space visually comparable to the filled blue selected cell while remaining distinguishable from the green legal-move grid.

Browser QA confirmed the computed occupied-cell fill is non-transparent and the old outline-only presentation is gone.

## Verification
- `npm run build` green.
- `go test ./...` green.
- `git diff --check` green.
- Isolated port 7187 restored from a read-only copy of the current canonical Triangle battle.
- Desktop and mobile popup viewports exercised in real Chrome.
- No `Zurück zur Begegnung` button present in Tactical.
- Action row contains exactly `Scannen / Warten / Fertig / Rückzug`.
- Mobile roster is exactly four columns and scrolls after two rows.
- Wider viewport automatically increases roster column count.
- Occupied cells are visibly filled translucent red.

Canonical 7171 was not mutated during isolated QA.
