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
