# Slice 16.1 Gate 2 - single-card visual selector review

Date: 2026-09-15
Status: **Gate 2 active / review candidate; not frozen yet**

## User review direction

The first Gate-1 prototype used previous/next arrows beside an inner selector card. User review changed the preferred reusable grammar before Gate 2 freeze:

- use **one selector card**, not an outer card plus a second inner card;
- put the setting heading (for example `Galaxiegröße`) above the visual;
- let the visual use as much of the available card width/height as practical;
- place the selected option name centered below the visual;
- place previous/next arrows left and right of that selected option name in the same control row;
- keep compact facts/status below without creating another card-like box;
- keep the component generic so Difficulty, Galaxy, Starting Tech, Race and Opponent Count can reuse the same navigation grammar with setting-specific content.

## Implemented review candidate

`web/src/components/VisualSelector.tsx` now owns a single visual hierarchy:

1. centered selector heading;
2. full-width visual;
3. control row: previous arrow / centered selected option / next arrow;
4. compact facts;
5. position dots/label.

The New Game page no longer renders the previous extra heading/hint block above the component and the selector no longer renders an inner `.visual-selector-card`.

The server-support contract remains unchanged: only `small` is currently accepted for galaxy size. Browsing planned options still locks Create Game.

## Responsive browser QA

Live QA was executed against `moox-server` on port 7171 after the layout revision.

### 390 px mobile

- viewport width: `390 px`;
- document scroll width: `390 px` -> no horizontal overflow;
- outer New Game visual card width: `362 px`;
- selector content width: `340 px`;
- visual: `340 x 250 px`;
- control row width: `340 px`;
- arrow buttons: `46 x 46 px`;
- heading: `Galaxiegröße` above the visual;
- selected option is centered between arrows;
- nested `.visual-selector-card` count: `0`.

Compared with the Gate-1 prototype, the mobile visual grows from `242 px` to `340 px` width because the arrows no longer consume the image row.

### Desktop 1280 px

- outer visual card width: `920 px`;
- selector content width: `882 px`;
- visual: `882 x 440 px`;
- centered control row: `640 px`;
- arrow buttons: `60 x 60 px`;
- no horizontal overflow observed.

## Validation

- `npm run build` passes after the revision.
- Keyboard navigation remains attached to the reusable selector root.
- Mouse/touch arrow controls remain first-class buttons.
- The old selector CSS implementation was removed rather than left as a competing override.

## Gate-2 note

This layout is the current freeze candidate, but Gate 2 remains intentionally **open** until user review of the revised live UI and the pending image/art discussion are complete.
## Compact-card follow-up review

A second user review found the first single-card candidate too visually dominant. The reference card was therefore reduced without changing its interaction grammar:

- desktop/tablet visual-card maximum width: **640 px** instead of filling the 920 px content column;
- desktop/tablet visual: approximately **606 x 341 px** at the current content width, down from 882 x 440 px;
- 390 px mobile visual: approximately **342 x 214 px**, down from 340 x 250 px by using a flatter 16:10 mobile ratio;
- desktop arrows: 52 px square; mobile arrows remain at the 44 px touch-size floor;
- headings/value typography and vertical gaps were reduced slightly so the card reads as one setting among several, not as a full-screen hero.

### Responsive settings-grid direction

The New Game visual settings now sit inside `.new-game-settings-grid`, using an auto-fit CSS grid with a 340 px minimum card track. This establishes the intended later Slice-16 composition:

- phone/narrow widths: one setting card per row, all settings stacked vertically;
- tablet/desktop widths: two or more setting cards may sit beside each other when space permits;
- a single currently implemented card remains centered by its own 640 px maximum width;
- each selector preserves the same internal heading -> visual -> previous/value/next -> facts grammar.

This is still a Gate-2 review candidate rather than a frozen contract until user visual review and image/art direction are complete.
