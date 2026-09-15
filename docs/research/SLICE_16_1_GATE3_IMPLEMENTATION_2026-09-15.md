# Slice 16.1 Gate 3 - Shared visual-selector implementation

Date: 2026-09-15
Status: **complete**

## Objective

Turn the Gate-2-frozen visual-selection grammar into a reusable implementation foundation that later Slice-16 settings can consume without re-inventing layout, interaction, asset conventions or supported/planned semantics.

Gate 3 deliberately does **not** implement the gameplay/settings breadth owned by Slice 16.2-16.6. Galaxy Size is the reference consumer; downstream settings must migrate onto the same foundation when their own sub-slices land.

## Shared component foundation

`web/src/components/VisualSelector.tsx` is the common implementation.

Gate-3 hardening makes the option ID generic:

- `VisualSelectorOption<TId extends string>` carries the setting-specific option ID type;
- `VisualSelectorProps<TId>` keeps `selectedId`, `options` and `onChange` on the same type;
- consumers can instantiate `VisualSelector<SpecificOptionId>` without unsafe casts;
- the Galaxy Size reference consumer now uses `VisualSelector<GalaxyPrototypeSize>` and no longer casts `id as GalaxyPrototypeSize`.

This is important for later Difficulty, Galaxy Age, Race, Starting Tech and other selectors because invalid setting IDs can now be caught at TypeScript compile time rather than leaking through a generic string callback.

## Reference migration

Galaxy Size is the first migrated Slice-16 setting and proves the frozen grammar end to end:

- responsive setting card/grid;
- title above rounded image-led artwork;
- typed previous/current/next selection;
- keyboard navigation;
- position dots;
- artwork `?` information control;
- responsive modal/bottom-sheet details;
- semantic runtime asset paths;
- supported/planned state kept honest against the current server contract.

Game ID, seed and empire names remain normal text controls. Slice 16.2-16.6 own the migration of their visual settings and specialized bodies as they land; Gate 3 does not invent their gameplay values early.

## Deterministic contract check

New command:

```text
npm run check:new-game-selector
```

Implemented by `web/scripts/check-new-game-selector.mjs` and now run automatically by `npm run build` after the UTF-8 check.

It verifies:

- exactly five Galaxy Size manifest entries;
- semantic IDs and runtime paths;
- SVG format and original-procedural provenance;
- committed SVG existence;
- frozen 1200 x 675 native dimensions and 16:9 viewBox;
- title/description metadata;
- no embedded localized `<text>` in artwork;
- typed `VisualSelector<GalaxyPrototypeSize>` use and absence of the old unsafe cast;
- required ArrowLeft/ArrowRight/Home/End behavior in the shared selector;
- modal/carousel accessibility contract;
- responsive grid and minimum touch-target invariants;
- deterministic art generation by regenerating all five SVGs into an isolated temporary directory and comparing SHA-256 content to the committed runtime assets.

## Reusable browser smoke

New command:

```text
npm run check:new-game-selector:browser -- http://127.0.0.1:7171
```

Implemented by `web/scripts/check-new-game-selector-browser.mjs`.

The script locates Chrome/Chromium (or accepts `CHROME_PATH`), launches an isolated headless profile and verifies the real 7171 application through Chrome DevTools Protocol.

### 390 px mobile coverage

- German locale loads;
- Small/Klein is the initial option;
- Small SVG loads at native 1200 x 675;
- no horizontal overflow;
- info and arrow touch targets remain at least 44 px;
- artwork keeps the frozen 16 px radius;
- five position dots render;
- Small keeps Create Game enabled;
- previous-arrow navigation reaches Tiny/Winzig and disables Create Game;
- ArrowRight returns to Small;
- End selects Huge/Riesig;
- Home selects Tiny/Winzig;
- every option loads its matching SVG;
- only Small remains server-supported;
- the info dialog opens with detailed copy, focuses its close control and applies modal body state;
- Escape closes the dialog and returns focus to the artwork info control.

### Desktop coverage

At 1280 x 900 the smoke verifies:

- no horizontal overflow;
- compact artwork remains within the frozen ~606 px reference width;
- artwork uses the frozen 20 px radius;
- previous/next controls remain at least 52 px.

The browser smoke completed successfully on the LLO Desktop Agent against the live 7171 server.

## Build result

`npm run build` passes with this sequence:

1. UTF-8/mojibake guard;
2. deterministic New Game selector contract check;
3. TypeScript project build;
4. Vite production build.

No server/gameplay contract was changed.

## Gate-3 interpretation of downstream migration

The planned Gate-3 item "Migrate Slice-16 setting screens to the common grammar as their sub-slices land" is satisfied here as an implementation contract plus the Galaxy Size reference migration. It does **not** mean pre-building Slice 16.2-16.6 inside Slice 16.1.

Downstream ownership remains:

- Slice 16.2: Difficulty;
- Slice 16.3: Galaxy Size/Age breadth;
- Slice 16.4: Preset Race portraits/carousel;
- Slice 16.5: Starting Technology;
- Slice 16.6: Player/opponent composition and final integration.

Each must reuse the frozen/shared grammar unless a documented specialized body is required.

## Next

Gate 4 is next: final Slice-16.1 desktop/mobile acceptance plus keyboard/mouse/touch smoke and closure evidence. The Gate-3 browser smoke already provides reusable automation for that acceptance, but Gate 4 remains a separate closure decision.
