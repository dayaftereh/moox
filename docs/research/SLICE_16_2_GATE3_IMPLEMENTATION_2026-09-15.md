# Slice 16.2 Gate 3 - Difficulty implementation

Date: 2026-09-15
Status: **complete**

## Objective

Implement the Gate-2-frozen Difficulty contract end to end without pulling later AI-intelligence work into Slice 16.2.

Gate 3 delivers the authoritative server contract, persisted New Game selection, built-in-AI economic effects, deterministic fixtures, production assets, the shared visual selector and browser-visible server-derived facts.

## Implementation commits

- `be6ac99 feat: add authoritative difficulty contract`
  - stable `DifficultyID` values;
  - server-owned `DifficultyProfile` catalog;
  - New Game field/default/validation;
  - persisted game-state Difficulty;
  - schema-23-compatible legacy/fixture semantics where an omitted field means effective Normal;
  - read-only `GET /api/v1/new-game/difficulties` endpoint;
  - persisted `BuiltinAIControlled` empire marker derived from controller assignment.
- `a71fd57 feat: apply difficulty to builtin ai economy`
  - frozen additive per-role AI food/production/research/tax effects;
  - Difficulty-specific built-in-AI command-point deficit cost;
  - human empires remain on normal ruleset values;
  - deterministic all-five-level fixtures/tests.
- `80d4512 ui: add difficulty command crest assets`
  - five original deterministic 1200 x 675 SVG command-crest assets;
  - manifest/provenance/generator registration;
  - build-time deterministic asset contract checks.
- `94143e8 ui: integrate difficulty selector`
  - typed web Difficulty catalog/API contract;
  - `VisualSelector<DifficultyID>` integration;
  - Create Game submits selected `difficulty_id`;
  - localized fact text is calculated from server-owned profile values;
  - mobile/desktop live browser QA for Galaxy Size + Difficulty.

## Authoritative server contract

Supported wire/save IDs:

- `easy`
- `normal`
- `hard`
- `very_hard`
- `impossible`

Default/effective legacy value: `normal`.

The game-state schema remains version 23. Difficulty is an additive optional field for old persisted states; old/fixture states with an empty value resolve through `EffectiveDifficultyID()` to Normal. New Game persists an explicit supported ID.

Unknown non-empty IDs are rejected.

The server publishes the current contract from:

`GET /api/v1/new-game/difficulties`

The live endpoint was verified during Gate 3 and returned schema version 1, default `normal` and exactly the five frozen profiles.

## Built-in-AI ownership

Difficulty economy effects apply only to empires explicitly controlled by the built-in AI controller.

Controller assignments are transferred into the generated New Game settings before galaxy/economy generation. This ensures the AI marker is known while the initial colony economy is calculated; the first authoritative state therefore already contains the selected Difficulty effects.

Human-controlled empires do not receive direct Difficulty economy modifiers.

## Frozen gameplay effects implemented

Gate 3 implements the Gate-2 additive per-role model:

| Difficulty | Food / farmer | Production / worker | Research / scientist | Tax BC / population | AI CP deficit / excess CP |
| --- | ---: | ---: | ---: | ---: | ---: |
| Easy | -0.25 | -0.50 | -0.50 | -0.25 | 11 BC |
| Normal | 0 | 0 | 0 | 0 | 10 BC |
| Hard | +0.25 | +0.50 | +0.50 | +0.25 | 9 BC |
| Very Hard | +0.50 | +1.00 | +1.00 | +0.375 | 8.5 BC |
| Impossible | +0.75 | +1.50 | +1.50 | +0.50 | 8 BC |

Values are stored server-side as integer eighth-units and converted only at the calculation edge.

Application order:

1. calculate the normal race/source per-role output;
2. apply the built-in-AI Difficulty per-role delta;
3. multiply by assigned population/cohort and conquered-population factors;
4. apply contextual government/morale/gravity rules.

Food keeps genuinely non-farmable zero-output climates at zero rather than creating food from the generic role floor. Production/research preserve the existing >=1 role floor; tax preserves >=0.

Command-point overage uses the Difficulty profile only for built-in AI empires. Human empires keep the ruleset-standard 10 BC per excess command point regardless of selected Difficulty.

## Explicit non-goals retained

Gate 3 does not implement or claim active Difficulty effects for:

- diplomacy hostility/personality;
- strategic planning/search quality;
- tactical intelligence;
- spying strategy/bonuses;
- troop/marine bonuses;
- population-growth modifiers.

Later AI/system slices may consume `DifficultyID`, but they do not own or redefine the Slice-16.2 IDs or V1 economy contract.

## Deterministic fixtures

Automated Go coverage now proves:

- the five catalog profiles equal the frozen values;
- omitted New Game Difficulty defaults to Normal;
- every supported explicit ID persists exactly;
- unknown IDs are rejected;
- legacy/fixture empty Difficulty remains schema-23-compatible and resolves to Normal;
- only built-in-AI economy changes with Difficulty;
- human colony economy remains unchanged across all five levels;
- all five command-deficit rates are exact, including 8.5 BC on Very Hard;
- a human empire at Impossible still uses the standard 10 BC command overage;
- same seed + same complete New Game settings + same Difficulty produces byte-identical state for every supported level.

## Production artwork and asset contract

Five deterministic original MOOX command-crest SVGs are generated under:

`web/public/assets/new-game/difficulty/`

Runtime asset option IDs are kebab-case per the Slice-16.1 asset convention:

- `easy`
- `normal`
- `hard`
- `very-hard`
- `impossible`

The authoritative server ID remains `very_hard`. Gate-3 implementation exposed this cross-contract naming collision and resolved it by mapping server ID `very_hard` to asset option ID `very-hard`, rather than weakening the global semantic asset validator.

All assets are:

- 1200 x 675 / 16:9;
- SVG;
- original deterministic vector work;
- free of embedded localized text;
- registered with semantic manifest IDs and provenance;
- byte-reproducible from `web/scripts/generate-difficulty-art.mjs`.

The build-time New Game contract check now validates ten deterministic assets total: five Galaxy Size plus five Difficulty.

## Web selector and server-derived facts

The web client now fetches the authoritative Difficulty catalog and renders a second typed shared selector:

`VisualSelector<DifficultyID>`

The selector uses the five production command-crest assets and the shared Slice-16.1 interaction contract.

All five Difficulty choices are currently supported. With the currently supported Small galaxy selected, changing Difficulty never locks Create Game.

Detailed information is kept behind the artwork information control. The numeric facts are formatted from the server profile's eighth-unit fields; there is no separate frontend gameplay-number table.

If the authoritative Difficulty catalog is unavailable, Create Game remains locked instead of silently guessing a profile.

## Browser/live QA

The reusable Chrome browser smoke was extended to identify visual selectors by semantic `data-setting-id`, avoiding fragile "first selector on the page" assumptions.

390 px mobile QA covers:

- no horizontal overflow;
- both Galaxy Size and Difficulty selectors;
- 44 px minimum controls;
- real CDP touch interaction with off-screen cards scrolled into view first;
- keyboard Home/End/Arrow navigation;
- all five Difficulty labels/assets;
- every Difficulty marked supported;
- Create Game stays enabled for all five Difficulties while Galaxy Size is Small;
- Impossible info dialog exposes the live server-derived values, including +0.75 food/farmer, +1.5 production/research and 8 BC command deficit;
- modal/focus behavior remains intact.

Desktop QA covers the two-card responsive grid, no horizontal overflow, 20 px artwork radius and >=52 px arrow controls. With two cards side by side, artwork is approximately 419 px wide at the 1280 px test viewport, which is valid under the frozen multi-card responsive contract.

## Verification result

Final Gate-3 regression run passed:

- `go test ./...`;
- `npm run build`;
- UTF-8/mojibake guard;
- ten-asset deterministic New Game contract check;
- TypeScript + Vite production build;
- `npm run check:new-game-selector:browser -- http://127.0.0.1:7171`;
- `git diff --check`.

Live `/healthz` and `/api/v1/new-game/difficulties` were verified after restarting the backend on the Gate-3 code.

## Gate-3 checklist result

- [x] Add authoritative difficulty setting/catalog support.
- [x] Add the visual carousel selector.
- [x] Bind concise modifier facts from server/normalized data.
- [x] Add deterministic New Game fixtures for each supported level.

## Next

Gate 4 remains separate. It should perform final closure acceptance focused on deterministic equality, proving Difficulty changes are limited to the frozen intended effects, desktop/mobile visual review and final full tests/build/diff checks before Slice 16.2 is closed.
