# Slice 15.1 mobile-first UX / navigation / design-system Gate 1 evidence

Date: **2026-09-04**

Slice: **15.1 - UX architecture / navigation / design system contract**

Status: **Gate 1 audit complete; Gate 2 interaction/design-system freeze pending user approval**.

## User acceptance direction

The browser HMI is **mobile-first** because day-to-day development/review is primarily from a phone. Desktop remains a required second layout mode rather than an afterthought.

After Slice 15.1 implementation/QA, a review server must be started on the LLO Desktop machine and exposed only through its NetBird interface so the user can inspect the UI directly from the phone. The review endpoint must use the built MOOX web assets and the same-origin Go API/WebSocket server.

Current infrastructure facts found during Gate 1:

- `moox-server` default bind is `127.0.0.1:7171`.
- non-loopback bind requires explicit `-insecure-allow-nonloopback`; the user explicitly chose a broad development bind on `0.0.0.0:7171` so the same preview can be reached through NetBird, LAN or localhost. This exposes the unauthenticated development server on every local interface, so the port is for trusted-network development only;
- current NetBird interface is `wt0`, IPv4 **100.120.252.216/16**;
- port **7171** was checked and free before preview startup;
- preferred/frozen review port is **7171**; if it becomes occupied, stop and resolve the conflict rather than silently changing the documented development URL;
- the React API uses relative `/api/...` requests and derives WebSocket URLs from `window.location.host`, so static UI + API + WebSocket can be served from one NetBird URL;
- the Go server already serves SPA assets from `web/dist` and falls back to `index.html` for browser routes.

No review server is started in Gate 1. It is a post-implementation acceptance step after 15.1 Gate 3/4.

## Current browser baseline

The current frontend is deliberately small and transport-oriented:

- `web/src/api.ts`
- `web/src/App.tsx`
- `web/src/main.tsx`
- `web/src/styles.css`

There is no routing dependency and no reusable component/navigation layer yet. `App.tsx` is effectively one long technical page.

Current visible sections are:

1. `Authoritative Web HMI`
2. `New Game`
3. `Game completed`
4. active Empire heading
5. `Invasion`
6. `Diplomacy`
7. `Participant Battles`
8. `Concrete gameplay command`
9. `Latest invalidation`

The page already has a dark blue functional theme and one breakpoint at `max-width:760px`, but mobile behavior is only structural stacking:

- shell padding shrinks;
- hero becomes block layout;
- metrics collapse from five columns to two;
- two-column `.grid` becomes one column.

This is not yet a mobile interaction architecture: there is no persistent navigation, no contextual action surface, no touch-oriented hierarchy, no explicit compact status bar and no route/state model for moving among game areas.

## UX problems to solve in 15.1

### 1. Long-page dashboard does not scale to gameplay

As 15.2-15.5 add galaxy, colonies, research, fleets, diplomacy, decisions, save/load and the interactive Tactical battlefield, one vertically stacked page becomes unusable on a phone.

### 2. Technical state dominates player intent

`Latest invalidation`, raw snapshots and implementation-oriented command sections are useful diagnostics but should not occupy primary player navigation. Diagnostics need a deliberate developer/debug surface.

### 3. No stable global game context

The player needs turn, phase, connection state, active game and important pending decision visibility while moving between screens.

### 4. Mobile and desktop need different composition, not separate products

The same information architecture should adapt:

- phone: single main content column + bottom navigation + sheets/drawers for context/actions;
- desktop: navigation rail/sidebar + wider main canvas + optional contextual side panel.

### 5. Lifecycle/error/reconnect states need first-class UX

The shell must explicitly represent:

- no game / main menu;
- creating game;
- loading snapshot;
- connected;
- reconnecting;
- stale snapshot / invalidated;
- authoritative command rejected;
- game loaded and refetch required;
- completed game;
- fatal/unrecoverable session state.

## Canonical browser journey map

### Top-level lifecycle

1. **Main Menu**
   - New Game
   - Resume/Load entry point placeholder for later 15.4 functionality
   - current/recent game entry if available
2. **New Game**
   - current supported narrow configuration
   - create and enter game
3. **Game Shell**
   - persistent global game status
   - domain navigation
   - contextual pending-decision surface
4. **Interactive decision mode**
   - invasion / battle / post-resolution / research decisions can temporarily take priority
5. **Completed / Result**
   - winner/result summary
   - return to main menu / inspect game

### Primary game domains

The shell must reserve stable destinations for:

- **Galaxy** - strategic overview / systems / selected-system context
- **Colonies** - colony economy, population, construction
- **Fleets** - fleets, movement and later colonize/outpost actions
- **Research** - progress and choices
- **More** - diplomacy, save/load, game details/settings and diagnostics as appropriate

Battle/Invasion/other blocking decisions are not permanent bottom-nav destinations; they appear as prioritized contextual workflows when the authoritative session phase requires them.

## Recommended mobile-first information architecture for Gate 2

### Mobile primary target

Recommended acceptance target to freeze in Gate 2:

- minimum supported layout width: **360 px**;
- primary review sizes: approximately **390 x 844** and **412 x 915**;
- safe handling down to 320 px for catastrophic overflow prevention, without promising ideal density;
- touch targets should generally be at least ~44 px high/wide where controls are primary;
- no required horizontal page scrolling.

### Mobile shell

**Top app bar**

- compact MOOX mark/title;
- turn + phase summary;
- connection/reconnect indicator;
- overflow/menu affordance.

**Main content**

- exactly one principal task/screen at a time;
- cards/lists may stack, but diagnostics/raw JSON are hidden from default player flow.

**Bottom navigation**

Five stable positions recommended:

1. Galaxy
2. Colonies
3. Fleets
4. Research
5. More

Use labels plus icons eventually; icon art direction is finalized in 15.3.

**Contextual action sheet**

- bottom sheet/drawer for selected system/colony/fleet actions;
- authoritative legal actions only;
- blocking session decisions may expand to full-screen on small phones.

### Desktop secondary target

Recommended acceptance target to freeze in Gate 2:

- primary review at **1440 x 900**;
- functional support from ~1024 px width upward;
- left navigation rail/sidebar containing the same domain destinations as mobile bottom navigation;
- top global status strip;
- main content canvas;
- optional right contextual inspector/action panel where space permits;
- no separate desktop-only workflow semantics.

## Navigation / route recommendation

Because the Go server already provides SPA fallback and the project currently has no router dependency, 15.1 should keep navigation lightweight.

Recommended Gate-2 contract:

- browser history/state-backed internal route model;
- route identity must be deep-linkable/recoverable for the current game/domain;
- exact library choice remains implementation detail unless Gate 2 decides a router dependency is justified;
- minimum route concepts:
  - main menu
  - new game
  - game + galaxy
  - game + colonies
  - game + fleets
  - game + research
  - game + more/diplomacy
  - result
  - contextual blocking decision state.

The route must never become gameplay authority; it only selects a presentation surface.

## Baseline design-system contract to freeze in Gate 2

15.1 should establish reusable structural primitives but **must not finalize the MOOX art direction**, which belongs to 15.3.

Recommended primitives:

- `AppShell`
- `TopBar`
- `PrimaryNav`
- `PageHeader`
- `Panel/Card`
- `Metric/ResourceBadge`
- `ListRow`
- `Tabs/SegmentedControl`
- `PrimaryButton`, `SecondaryButton`, `DangerButton`, `IconButton`
- `Field`, `Select`, `NumberField`
- `Dialog`
- `BottomSheet/Drawer`
- `Toast/Notice`
- `ConnectionState`
- `LoadingState`
- `EmptyState`
- `ErrorState`
- `PendingDecisionBanner`

Baseline tokens should cover:

- spacing scale;
- type scale;
- semantic foreground/background/border/accent/danger/warning/success roles;
- border radii;
- elevation/layering;
- focus/hover/pressed/disabled states;
- z-index/layer contract;
- responsive breakpoints.

The existing dark-blue palette may be used as the temporary structural skin; 15.3 can replace visual values without rewriting layout semantics.

## Diagnostics policy

The current HMI exposes implementation/debug information directly. Gate 2 should freeze:

- player-facing UI by default;
- diagnostic information behind a developer/debug panel in `More` or an explicit development-only toggle;
- raw JSON snapshots should not be required for normal gameplay or mobile review.

## Mobile preview acceptance after implementation

After 15.1 Gate 3/4 is green:

1. build `web/dist` with `npm run build`;
2. re-check that port **7171** is free;
3. start `moox-server` on **`0.0.0.0:7171`** with `-insecure-allow-nonloopback` and built `web/dist` assets;
4. re-resolve the current NetBird interface/IP and verify `/healthz`, static UI and same-origin API/WebSocket reachability through that address;
5. keep the broad bind only for trusted development networks; if remote access is blocked, add a targeted Windows Firewall inbound rule for TCP 7171;
6. launch/maintain the agent-managed Chrome profile `moox-ui-review` against the running UI for desktop navigation, responsive checks and screenshot-based review;
7. keep the managed server process attached to the active ASH work session while review is in progress;
8. report the exact `http://<NetBird-IP>:7171/` URL to the user for phone testing.

The first live baseline preview was successfully started during Gate 1 on `0.0.0.0:7171`; both `http://127.0.0.1:7171/healthz` and `http://100.120.252.216:7171/healthz` returned OK, the NetBird UI returned HTTP 200, and managed Chrome opened the page with title `Master of Orion X`. Current phone URL: `http://100.120.252.216:7171/`.

## Multi-language requirement added before Gate 2

User requirement confirmed on 2026-09-04: the browser HMI must support **German and English from the first structural implementation**, with a runtime language switch. This is intentionally captured before Gate 2 so layout, component APIs and copy handling are localized from the start rather than retrofitted later.

Recommended Gate-2 i18n contract:

- supported locales initially: `de` and `en`;
- runtime switching must not require restarting/recreating a game and should update the visible shell immediately;
- language is a **client presentation preference only** and must not alter GameState, command payload semantics, deterministic simulation, live snapshots or completed snapshots;
- persist explicit user choice locally in the browser;
- first-run fallback: supported browser language (`de*` -> German, `en*` -> English), otherwise English;
- all structural/player-facing UI copy introduced by 15.1 uses translation keys, not hard-coded English/German strings inside screen components;
- canonical server IDs/enums/action kinds remain language-neutral; UI maps them to localized labels;
- game-defined proper names/user names are not automatically translated unless a later content catalog explicitly provides localized display names;
- German and English must both be tested on phone and desktop because German labels are often longer and can expose responsive overflow;
- language switch should live in an always-reachable settings/`More` surface, with an optional compact shortcut if Gate 2 wireframes show sufficient room.

This contract is inherited by Slices 15.2-15.6.

## Gate 1 conclusion

Gate 1 supports moving to Gate 2 with the following recommended freeze:

- **mobile-first**, desktop-adaptive single information architecture;
- phone uses top status bar + five-item bottom navigation + contextual sheets;
- desktop uses the same destinations through a left navigation rail and optional context panel;
- functional dark baseline skin in 15.1, final MOOX visual identity deferred to 15.3;
- explicit lifecycle/reconnect/error/pending-decision states;
- diagnostics removed from the primary player flow;
- live NetBird phone-preview server is a required 15.1 post-implementation acceptance step;
- German + English runtime localization is mandatory from 15.1 onward, with locale kept entirely outside authoritative game state.
