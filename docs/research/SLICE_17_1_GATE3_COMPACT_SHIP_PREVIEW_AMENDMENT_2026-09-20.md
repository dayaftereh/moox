# Slice 17.1 Gate 3 - Compact Ship Preview Amendment

Date: 2026-09-20
Slice: 17.1 - Hulls and Mandatory Ship Systems
Gate: 3 HMI amendment
Baseline: `e51cfa9` (`feat(ui): compact ship system information`)
Session: `ses-20260920T090827-205c7a6fbea5`

## User review

Desktop review found the Ship preview and hull/construction selector still too large relative to the newly compacted system-information surface.

The requested change was presentation-only:

- reduce the hull selector footprint;
- reduce the Ship image/preview footprint;
- preserve the complete system-information card directly below;
- preserve mobile usability and touch targets;
- do not change gameplay/domain rules.

## Implemented layout change

### Hull selector

Desktop:

- selector columns reduced from 48px arrows to 42px columns;
- arrow minimum width reduced from 44px to 40px;
- selector/arrow height reduced from 60px to 48px;
- hull title reduced from 1rem to .92rem;
- section/stepper spacing reduced.

Mobile:

- 44px touch targets are retained at <=360px;
- hull selector remains at least 44px high.

### Ship preview

The preview no longer relies only on a minimum height. It now has an explicit bounded block size so parent/grid/SVG layout cannot stretch it vertically.

Desktop:

- preview fixed to 184px;
- procedural Ship glyph reduced to 148px and max width 520px;
- preview top margin reduced to 8px;
- preview padding reset to zero.

<=620px:

- preview fixed to 168px;
- Ship glyph 136px.

<=360px:

- preview fixed to 150px;
- Ship glyph 120px.

The preview keeps the existing reroll affordance and clipping.

## Browser guard

The reference-lab browser smoke now measures the actual rendered layout.

At 390px mobile width it requires:

- Ship preview height <=185px;
- hull selector height <=52px;
- no horizontal overflow.

At 1280px desktop width it requires:

- Ship preview height <=205px;
- hull selector height <=52px;
- no horizontal overflow.

This specifically guards against the prior behavior where changing only `min-height` still allowed the preview to stretch far beyond the intended size.

## Verification

- `npm run build` PASS.
- `git diff --check` PASS.
- fresh Reference server on port 7195 health PASS.
- `npm run check:reference-lab:browser -- http://127.0.0.1:7195` PASS.

The combined browser smoke still covers:

- compact mandatory-system rows and detail popups;
- Computer/Shield rich selectors;
- explicit no-Shield save;
- All-Tech Doom Star flow;
- mobile and desktop layout;
- Development Tools;
- BC grant and normal construction buyout;
- ordinary-game isolation.

## Scope

No server/domain/gameplay rule changed.

Gate 4 remains not started.
Slice 17.2+ remain closed.


## Mobile SVG containment regression fix

A follow-up mobile review exposed that Titan and especially Doom Star could be clipped after the compact-preview change.

Root cause:

- the global rule `.procedural-ship-glyph:not(.tactical-ship-vector)` has higher CSS specificity than the original `.shipdesigner-preview-ship` sizing rule;
- therefore the intended fixed preview SVG height was ignored and the SVG fell back to `width: 100%; height: auto`;
- the larger Titan/Doom Star intrinsic canvas then exceeded the compact mobile preview and was clipped by the preview container's `overflow: hidden`.

Fix:

- the Ship Designer preview sizing rules now use `.procedural-ship-glyph.shipdesigner-preview-ship`, matching/exceeding the global SVG selector specificity;
- desktop, <=620px and <=360px preview SVG heights now apply deterministically;
- the browser smoke now advances to Doom Star at 390px and verifies that the actual SVG bounding box is fully contained and centered in the preview, rather than checking only the preview container height.

Fresh reference verification used port 7196 and passed the complete mobile/desktop Ship Designer smoke.


## Hull-size progression follow-up

A second mobile review confirmed that containment alone was not sufficient: after the clipping fix the SVG frame still grew by hull class, but procedural silhouette variance could make a Titan appear as large as or larger than a Doom Star.

The Ship Designer preview now separates two concerns:

- procedural morphology remains free to vary per reroll;
- perceived hull-class size follows a dedicated Designer ladder.

For the Designer preview only, the procedural silhouette is normalized around its generated visible bounds before the hull-class footprint is applied. A stronger preview-only progression is then used from Frigate through Doom Star, while other ship render contexts keep their existing scale contract.

At the 390px deterministic All-Tech reference sequence, the measured visible maximum silhouette extent is approximately:

- Frigate: 37px;
- Destroyer: 56px;
- Cruiser: 71px;
- Battleship: 81px;
- Titan: 88px;
- Doom Star: 97px.

The browser smoke now measures the rendered path/circle/ellipse union for every hull while stepping through the Designer and requires every successive hull to be visibly larger. The existing Doom Star containment and centering guard remains active.
