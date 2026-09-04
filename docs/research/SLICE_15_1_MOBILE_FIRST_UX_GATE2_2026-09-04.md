# Slice 15.1 Gate 2 - Mobile-first interaction/design-system contract

Date: **2026-09-04**

Status: **frozen / approved; Gate 3 authorized**.

## Frozen information architecture

Top-level lifecycle:

1. **Main Menu**
2. **New Game**
3. **Running Game Shell**
4. **Blocking/priority decision surface** when authoritative phase requires it
5. **Completed / Result**

Running-game primary domains:

- **Galaxy**
- **Colonies**
- **Fleets**
- **Research**
- **More** (Diplomacy, game/development controls, language, diagnostics)

The route selects presentation only. It never becomes gameplay authority.

## Mobile shell - primary acceptance target

- design mobile-first;
- primary review widths around 390-412 px;
- minimum functional layout 360 px; prevent catastrophic overflow at 320 px;
- sticky compact top status bar with product identity, turn/phase and connection state;
- fixed five-destination bottom navigation: Galaxy / Colonies / Fleets / Research / More;
- one principal content task at a time;
- contextual actions/priority decisions use prominent cards/sheets rather than expanding the long-page dashboard;
- primary touch controls target at least ~44 px;
- no required horizontal page scrolling.

## Desktop adaptive shell

- same information architecture and semantics as mobile;
- primary QA at 1440x900, functional from ~1024 px upward;
- left navigation rail replaces mobile bottom navigation;
- sticky top global status bar;
- wider main content canvas;
- contextual/right-side information may use additional horizontal room but must not create desktop-only gameplay flows.

## Navigation contract

No router dependency is required for 15.1. Use a lightweight browser-history/hash route model that is refresh/deep-link recoverable and keeps server SPA fallback simple.

Frozen concepts:

- home/main menu;
- new game;
- game + Galaxy;
- game + Colonies;
- game + Fleets;
- game + Research;
- game + More;
- completed/result through the running game context;
- blocking authoritative decisions remain phase-driven overlays/cards rather than permanent navigation destinations.

## Interaction/lifecycle states

The shell must explicitly expose:

- discovering hosted games;
- no hosted game;
- creating game;
- loading/synchronizing snapshot;
- connected/live;
- reconnecting;
- stale/invalidated snapshot and refetch;
- authoritative command rejected/server error;
- pending blocking decision;
- completed game.

Diagnostics/raw JSON must move out of the primary player flow and live under More/developer diagnostics.

## Structural design-system freeze

15.1 establishes reusable structural primitives and semantic tokens, not final MOOX art direction. Slice 15.3 may replace visual values without changing interaction semantics.

Required primitives include:

- AppShell / TopBar / PrimaryNav;
- PageHeader;
- Card/Panel;
- Metric/ResourceBadge;
- ListRow;
- Button variants and fields/selects;
- Notice/Error/Empty/Loading state;
- PendingDecision surface;
- compact language switch;
- developer diagnostics surface.

Required semantic tokens include spacing, type scale, surfaces, text roles, border/accent/status roles, radii, focus/pressed/disabled states, layer/z-index roles and responsive breakpoints.

## German / English localization freeze

Initial locales: **German (`de`) and English (`en`)**.

- runtime switching without game restart or reload;
- persistent explicit client preference;
- first-run choice: stored preference -> supported browser language -> English fallback;
- language is presentation-only and must not enter authoritative GameState, command semantics, snapshots, saves or replay;
- structural/player-facing copy uses translation keys;
- canonical server IDs/enums remain language-neutral and are mapped to localized labels;
- proper/user-defined names remain unchanged unless a later content catalog provides explicit localized names;
- both languages are QA targets on mobile and desktop; German long-label overflow is explicit acceptance coverage;
- switch is always reachable in the shell/More area.

## Gate 3 implementation boundary

Gate 3 may refactor React presentation structure and local client UI state. It must not add/change gameplay rules, legality, costs, target selection or deterministic simulation.

Existing functional proof surfaces (New Game, snapshot loading, population assignment, Diplomacy, Invasion, completion and invalidation diagnostics) must remain reachable through the new shell. Galaxy/Fleet/Research breadth not yet projected by the current API may be represented by honest structural/empty states for later Slice 15.2 rather than fabricated gameplay data.

## Live acceptance environment

Frozen review endpoint:

- MOOX server bind: `0.0.0.0:7171`;
- current NetBird URL: `http://100.120.252.216:7171/`;
- user already confirmed successful phone access;
- managed Chrome profile/session is used for desktop navigation/visual QA;
- add Windows Firewall TCP 7171 rule only if later remote access is actually blocked.
