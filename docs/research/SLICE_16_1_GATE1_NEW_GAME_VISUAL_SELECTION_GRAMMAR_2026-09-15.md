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
## 5. Frozen Gate-1 asset convention

Slice 16.1 now reserves one stable manifest and path grammar:

- manifest URL: `/assets/new-game/manifest.json`;
- repository root: `web/public/assets/new-game/`;
- semantic ID: `new-game:<domain>:<option-id>`;
- runtime setting-art path: `/assets/new-game/<domain>/<option-id>.<svg|webp|png>`;
- domains: `difficulty`, `galaxy-size`, `galaxy-age`, `technology-level`, `player-count`;
- option IDs are lowercase kebab-case and stable across localized display copy.

`web/src/newGameAssets.ts` owns typed path helpers and validates semantic option/race IDs. The public manifest intentionally starts with no file entries because the Slice-16.1 prototype uses original CSS geometry rather than pretending missing runtime assets exist. Later sub-slices add entries only when corresponding files land.

Race portrait identity remains canonical at `/assets/races/<race-id>/portrait.webp` (or intentional PNG) plus `emblem.svg`, matching Slice 16.4; New Game references those assets rather than duplicating them.

## 6. Generic selector versus specialized-card classification

The shared `VisualSelector` owns navigation, focus/keyboard behavior, option position, availability/locked state and responsive frame. Setting-specific bodies own artwork and facts.

| Slice / setting | Shared selector shell | Specialized body / reason |
| --- | --- | --- |
| 16.2 Difficulty | yes | Setting-specific emblem/strength art and evidence-backed facts only. |
| 16.3 Galaxy size | yes | Galaxy diagram/illustration body; invalid player-count combinations can use shared locked state. |
| 16.3 Galaxy age | yes | Stellar/nebular illustration body; same shell. |
| 16.4 Preset race | yes | Specialized 4:5 portrait body with race emblem/facts and optional detail affordance; navigation/frame stay shared. |
| 16.5 Starting technology | yes | Standard image/fact body; a desktop three-card presentation is allowed only as a layout variant of the same selection contract. |
| 16.6 Opponent count | yes | Number art plus compact capacity facts; enabled maximum remains server-authoritative. |
| 16.6 Opponent/player composition | partial | Specialized multi-slot portrait/controller layout; embeds the generic opponent-count selector rather than forcing the whole composition into one card. |
| 16.6 Final New Game summary | no carousel requirement | Specialized visual summary composed from the already-selected assets. |

Game ID, seed and editable empire names remain ordinary form/text controls; forcing them into the image selector would reduce clarity without adding selection semantics.

## Gate 1 conclusion

All six Gate-1 items are complete. Gate 2 can now freeze the reusable selector component contract, responsive layout, accessibility/keyboard behavior, asset naming/runtime conventions and unsupported/disabled-state presentation before later 16.x implementation breadth depends on them.
