# Slice 15.1 Gate 4 - independent QA / closure evidence

Date: **2026-09-04**

Status: **PASS; Slice 15.1 Gates 1-4 complete**.

## Scope

This Gate-4 pass independently re-verified the implemented Slice-15.1 browser shell rather than relying only on Gate-3 implementation evidence. It covered navigation, responsive behavior, DE/EN runtime localization/persistence, focus/loading/error states, deep-link recovery, live NetBird serving, authoritative-state preservation and full repository QA.

The accepted downstream gameplay direction from later planning discussions does not reopen 15.1: Galaxy/Colony/Fleet/Research/Diplomacy/Espionage feature depth remains Slice 15.2+, final visual identity remains 15.3, rich decisions/persistence remains 15.4, interactive Tactical Combat remains 15.5 and final integration remains 15.6.

## Browser/runtime QA

Managed browser profile/session: `moox-ui-review` / Chrome 152.

Live build: production Vite output served by the running `moox-server` preview on `0.0.0.0:7171`.

### Navigation / deep-link paths

Verified through the real browser runtime:

- Main Menu;
- New Game;
- direct hosted-game resume;
- Galaxy;
- Colonies;
- Fleets;
- Research;
- More;
- direct hash/deep-link recovery to `#/game/game-1/colonies` after a full navigation/reload.

The existing authoritative Colony projection remained intact during the localization/navigation checks: Farmers 4 / Workers 2 / Scientists 2.

### Desktop adaptive layout

Browser runtime reported:

- `innerWidth=1034`, `clientWidth=1019` (scrollbar excluded), `scrollWidth=1019`;
- no horizontal page overflow;
- desktop side rail `display:flex`;
- mobile bottom navigation hidden;
- German and English switching both worked without altering game state.

The primary 1440x900 contract remains a CSS/layout target; the managed Chrome window could not be programmatically resized above the host window bounds in this session, so Gate 4 independently re-verified the accepted desktop breakpoint/layout at the actual 1034px review viewport and retained Gate-3 structural proof for the desktop mode. No desktop-only workflow semantics exist.

### Mobile 390px

A real popup browser window was opened with requested width 390px:

- `innerWidth=390`;
- `clientWidth=375`;
- `scrollWidth=375`;
- no horizontal overflow;
- side rail hidden;
- five-item bottom nav visible;
- each nav target approximately 73x66px.

### Mobile 360px minimum functional target

A real popup browser window was opened with requested width 360px:

- `innerWidth=360`;
- `clientWidth=345`;
- `scrollWidth=345`;
- no horizontal overflow;
- five bottom-nav targets approximately 67x66px;
- German long-label More/session content remained usable.

### Mobile 320px catastrophic-overflow guard

A real popup browser window was opened with requested width 320px:

- `innerWidth=320`;
- `clientWidth=305` because the desktop Chrome scrollbar consumes 15px;
- `scrollWidth=320`, i.e. content did **not** extend beyond the actual 320px viewport;
- five bottom-nav targets approximately 59x66px;
- both German and English long labels remained inside the viewport;
- no required horizontal page scrolling beyond the 320px viewport.

This passes the Gate-2 320px catastrophic-overflow-prevention contract; 320px is not the preferred ergonomic target.

## Localization / persistence QA

Verified:

- DE -> EN switch on New Game;
- `localStorage['moox.locale']='en'` and `<html lang='en'>`;
- direct deep-link to an existing Colony view retained English and unchanged 4/2/2 authoritative population data;
- EN -> DE switch;
- full browser navigation/reload retained `localStorage['moox.locale']='de'` and `<html lang='de'>`;
- language changes remained presentation-only.

## Focus / loading / error-state QA

### Focus

Programmatic keyboard focus on the Research bottom-nav button produced a visible CSS focus outline:

- active element confirmed;
- outline style `solid`;
- outline width `2px`;
- outline offset `2px`.

### Loading

Snapshot fetch was deliberately delayed in the managed browser. The game route rendered the explicit loading state:

- `Synchronizing game`;
- `Waiting for the authoritative player snapshot.`.

### Error / stale-state finding and fix

Independent Gate-4 QA found one real defect before closure:

- navigating to a missing game correctly rendered a `role=alert` error;
- however, an old successful snapshot could still appear underneath because an already-started snapshot request could resolve after the route/game switch.

Gate 4 fixed this in `web/src/App.tsx` by:

- tracking the currently active game/seat in refs;
- discarding snapshot responses that no longer match the active game/seat;
- clearing prior errors on a game-route switch;
- suppressing the loading placeholder when a terminal fetch error is already shown.

Post-fix browser proof for `#/game/__missing_gate4__/galaxy` showed only the localized error alert and shell navigation, with no stale Galaxy/player snapshot underneath.

## Live NetBird preview

Current NetBird address resolved on LLO Desktop:

- `100.120.252.216`.

Gate-4 live checks:

- loopback `http://127.0.0.1:7171/healthz` -> `{"status":"ok"}`;
- self-request through NetBird address `http://100.120.252.216:7171/healthz` -> `{"status":"ok"}`;
- self-request through NetBird address to root -> HTTP 200;
- the user had already confirmed direct phone access to the same NetBird preview during Slice 15.1 review;
- no Windows Firewall rule was necessary.

Review URL remains:

`http://100.120.252.216:7171/`

The server remains intentionally unauthenticated and development-only on trusted LAN/NetBird.

## Final repository QA

After the Gate-4 stale-snapshot fix:

- `npm run build` - PASS;
  - TypeScript build PASS;
  - Vite 8.2.2 PASS;
  - 21 modules transformed;
  - final JS bundle `index-BtRtHOSp.js`;
- `go test ./...` - PASS across the full repository, including `internal/session`;
- `go vet ./...` - PASS;
- `git diff --check` - PASS before closure documentation.

## Closure conclusion

All Slice-15.1 Gate-4 acceptance items pass after the one stale-snapshot race fix. The shell is accepted as the structural/mobile-first bilingual navigation foundation, **not** as the final gameplay screen design or final visual identity.

Slice 15.1 is closed. The next prepared objective is **Slice 15.2 Gate 1 - functional strategic gameplay HMI audit**, using the accepted product-direction evidence in `docs/research/SLICE_15_2_STRATEGIC_HMI_PRODUCT_DIRECTION_2026-09-04.md` as binding downstream input.
