# Slice 15.1 Gate 3 - Mobile-first browser shell / navigation / i18n implementation

Date: **2026-09-04**

Status: **Gate 3 implemented; Gate 4 subsequently passed and Slice 15.1 is closed.**

## Implemented structure

Gate 3 replaces the previous long technical single-page React HMI with the frozen Slice-15.1 information architecture while preserving the existing authoritative HTTP/WebSocket/game-command contract.

New frontend structure:

- `web/src/i18n.tsx`
  - shared German (`de`) / English (`en`) catalogs;
  - runtime language switching;
  - persisted `moox.locale` browser preference;
  - first-run supported-browser-language detection with English fallback;
  - `document.documentElement.lang` follows the active locale;
  - compile-time catalog-key parity through the `Catalog = Record<TranslationKey, string>` contract.
- `web/src/navigation.ts`
  - lightweight hash routing with no new router dependency;
  - recoverable routes for main menu, New Game and game sections;
  - presentation-only route state.
- `web/src/components/AppShell.tsx`
  - sticky game top bar;
  - mobile bottom navigation;
  - desktop left navigation rail;
  - connection state, turn/phase context and always-reachable language switching.
- `web/src/components/ui.tsx`
  - shared Card, PageHeader, Metric, Notice and EmptyState primitives.
- `web/src/App.tsx`
  - main menu and hosted-game resume surface;
  - New Game screen;
  - running-game shell with Galaxy / Colonies / Fleets / Research / More destinations;
  - global Invasion priority decision card;
  - global completed-game result card;
  - existing real population assignment, Diplomacy and Invasion actions preserved;
  - diagnostics/raw invalidation JSON moved under More rather than occupying the primary player flow;
  - unsupported Fleet/Research/Galaxy depth is represented honestly as structural empty/pending states for Slice 15.2 rather than invented data.
- `web/src/styles.css`
  - mobile-first semantic design tokens and structural dark skin;
  - minimum touch-sized controls;
  - fixed safe-area-aware five-item bottom navigation on mobile;
  - desktop rail from the accepted breakpoint;
  - responsive cards/forms/metrics;
  - focus-visible and reduced-motion behavior;
  - final MOOX art direction intentionally deferred to Slice 15.3.

## Preserved authoritative gameplay behavior

The refactor does not move gameplay authority into React. Existing API functions remain authoritative and are still used for:

- deterministic New Game creation;
- player snapshot retrieval;
- WebSocket invalidation/refetch;
- `colony.assign_population` turn submission;
- war/peace Diplomacy immediate commands;
- Invasion accept/decline immediate commands.

No gameplay rules, costs, targets, legality or deterministic simulation were added to the browser.

## Live Gate-3 proof game

A disposable live game `ui-gate3` was created through the real browser New Game workflow with seed `0x8009`.

Observed authoritative player snapshot in the new shell:

- Game: `ui-gate3`
- Turn: 1
- Phase: Planning
- Revision: 1
- Change sequence: 1
- Empire: Human / Empire 2 / Seat 1 `local_human`
- Colonies: 1
- first projected colony population: Farmers 4 / Workers 2 / Scientists 2
- active participant battles: 0
- diplomatic contacts: 1

The Colonies screen exposes the same real 4/2/2 population values and retains the real revision-bound Submit Turn action.

## Browser navigation / localization smoke

Managed Chrome profile/session: `moox-ui-review`.

Cache-fresh live build was exercised through `http://127.0.0.1:7171/?v=gate3-final`.

Verified:

- Main Menu renders through the new standalone shell.
- New Game route renders and creates the real `ui-gate3` game.
- Galaxy route loads the authoritative game/empire snapshot.
- Colonies route loads current colony/population state and the real population command form.
- Fleets route loads current participant-battle projection and explicitly states that the full strategic Fleet projection belongs to Slice 15.2.
- Research route explicitly states that current player-HMI research projection is not yet available and belongs to Slice 15.2.
- More route preserves language, session controls, real Diplomacy controls and developer diagnostics.
- Runtime EN -> DE switching updates visible copy immediately.
- after a real browser reload, `localStorage['moox.locale'] == 'de'`, `<html lang="de">`, and the German navigation returns without recreating/reloading game state.
- accessibility labels for language, primary navigation and game status are localized as well.

## Responsive proof

Desktop browser state before mobile emulation:

- browser `innerWidth`: 1034 px;
- document client width: 1019 px;
- document scroll width: 1019 px;
- horizontal overflow: **false**;
- desktop side navigation display: **flex**;
- mobile bottom navigation display: **none**.

Chrome DevTools mobile metrics were then set to the primary frozen phone viewport **390 x 844**.

Observed mobile state:

- browser `innerWidth`: 390 px;
- document client width: 375 px;
- document scroll width: 375 px;
- horizontal overflow: **false**;
- desktop side navigation display: **none**;
- mobile bottom navigation display: **grid**;
- all five German bottom-nav entries measured approximately 73 x 66 px each, exceeding the ~44 px touch-target baseline;
- mobile Bottom Nav successfully navigated from More to Galaxie.

Gate 4 should independently re-check the lower 360 px target and catastrophic-overflow guard at 320 px, plus final representative desktop width, but the primary phone viewport is already proven in Gate 3.

## Build / regression checks

Passed during Gate 3:

- `npm run build`
  - TypeScript build PASS
  - Vite production build PASS
  - 21 modules transformed
- `go test ./internal/server ./internal/app`
  - PASS
- `git diff --check`
  - PASS
- UTF-8 BOM scan on all newly written frontend source files
  - BOM-free, matching repository style
- JSX hard-coded-copy scan
  - only deliberate language tokens `DE`/`EN`, product mark `OX`, and translated JSX expressions remain; structural copy is catalog-driven.

## Live review environment

The existing managed development server remains intentionally running:

- bind: `0.0.0.0:7171`
- current NetBird address: `100.120.252.216`
- phone URL: `http://100.120.252.216:7171/`
- the user previously confirmed that the phone can access this endpoint.

The live `web/dist` now contains the Gate-3 shell. A cache-busting query such as `?v=gate3-final` can be used if a browser still holds the previous bundle.

## Gate-3 conclusion

Gate 3 is complete for the Slice-15.1 implementation boundary:

- mobile-first / desktop-adaptive shell implemented;
- recoverable five-domain navigation implemented;
- reusable structural primitives/tokens implemented;
- explicit lifecycle/error/invalidation surfaces implemented;
- German/English runtime localization implemented and persistent;
- existing authoritative gameplay proof controls preserved;
- unsupported 15.2 breadth remains explicit rather than fabricated.

Next: independent Gate-4 QA, final phone/desktop review, full repository checks, commit/close workflow.
