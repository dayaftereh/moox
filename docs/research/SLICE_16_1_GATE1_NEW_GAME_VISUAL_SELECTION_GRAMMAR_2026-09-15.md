# Slice 16.1 Gate 1 - New Game visual selection grammar audit

Date: 2026-09-15
Status: **Gate 1 active**
Source specification: `docs/slices/PLANNED_16_1_NEW_GAME_VISUAL_SELECTION_GRAMMAR.md`

## Repository start state

- `main` synchronized with `origin/main` at `4720219`.
- Working tree clean at Gate-1 start.
- No pre-existing `_OPEN_*.md` marker.
- Slice 15.6 already closed; Slice 16.1 deliberately selected as the next implementation slice.

## 1. Current New Game UI inventory

The browser New Game surface is still a compact form in `web/src/App.tsx`.

Current editable fields:

- game ID;
- seed;
- Human empire name;
- Darlok empire name.

Current displayed-but-disabled facts:

- galaxy baseline;
- technology/combat baseline.

`submitNewGame` currently sends these fixed values:

- `galaxy_size: 'small'`;
- `galaxy_age: 'normal'`;
- `technology_level: 'average'`;
- `strategic_combat: false`;
- seat 1 `human` / local human;
- seat 2 `darlok` / built-in AI.

`web/src/api.ts` mirrors this narrow contract in `NewGameSettings`; the TypeScript unions currently permit only those exact settings and the Human/Darlok pair.

## 2. Server-authoritative contract inventory

`internal/game/new_game.go` defines the authoritative settings model and validates it before generation.

Gate-1 audit confirms the existing Slice-09 boundaries are still enforced server-side:

- galaxy size must be `small`;
- galaxy age must be `normal`;
- technology level must be `average`;
- strategic combat must be `false`;
- exactly two players are required;
- only `human` and `darlok` are accepted;
- each accepted race must occur exactly once.

Therefore Slice 16.1 must not make broader options look live merely because the visual selector can render them. Broader valid choices belong to later 16.x contract work.

## 3. Ship Designer previous/next audit

`web/src/ShipBuilderView.tsx` contains a local hull stepper:

- `stepHull(delta)` owns the selection change;
- previous/next controls are plain buttons;
- buttons have explicit ARIA labels;
- disabled edge states are already represented.

`web/src/styles.css` defines `.shipdesigner-arrow` with a 44 px minimum width and 60 px minimum height plus hover/focus-visible treatment. This is a useful visual/touch baseline.

There is **no shared carousel/selector component** to reuse directly. Reuse should therefore be conceptual/style-level initially, followed by extraction of a generic New Game visual-selector shell with its own neutral API.

## Gate-1 decisions so far

1. Server authority remains unchanged during the selector-foundation prototype.
2. The prototype must distinguish supported/live options from future/unsupported content.
3. The Ship Designer arrow treatment is an input to the new component, not a dependency that should be imported wholesale.
4. The reusable shell should own previous/next semantics, focusability, option position and accessible labels; setting-specific cards should own facts/art.

## Remaining Gate-1 work

- prototype one shared selector at desktop and 390 px;
- exercise keyboard, mouse and touch-size behavior;
- freeze the initial asset manifest/path convention;
- classify Slice-16 settings into generic-shell cards versus specialized card bodies.
## 4. Visual-selector prototype and live QA

The first shared selector prototype now exists in `web/src/components/VisualSelector.tsx` and is integrated into the New Game page as the Galaxy-size selector. The generic shell owns previous/next controls, keyboard stepping, option position, accessibility labels, availability presentation and the selected-card frame. The Galaxy prototype supplies setting-specific art/facts.

The prototype intentionally exposes Tiny/Small/Medium/Large/Huge for browsing while only `small` is marked supported. Selecting a planned size never changes the server payload and disables Create Game, so unsupported server capabilities are not presented as usable.

Browser-backed QA against the live `moox-server` on port 7171:

- production web build passed after the component integration;
- Go tests passed for `./cmd/moox-server`, `./internal/game` and `./internal/server`;
- 390 px mobile viewport: `innerWidth=390`, document/body `scrollWidth=390` (no horizontal overflow);
- mobile previous/next controls measure 44 CSS px wide and stretch to the selected card height;
- keyboard ArrowRight moved Small -> Medium and disabled Create Game because Medium is planned/locked;
- keyboard ArrowLeft returned Medium -> Small and re-enabled Create Game;
- button click moved Small -> Medium with the same locked-state behavior;
- 1280 px desktop viewport remained within width (document `scrollWidth=1265`) and uses 64 px side controls.

The placeholder galaxy visual is original CSS/vector-style geometry local to the application and introduces no external/copyrighted artwork.
