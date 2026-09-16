# Slice 16.4 Gate 3 - Preset race catalog, portraits and carousel implementation

Date: 2026-09-16
Status: **complete / Gate 4 next**

Gate-1 audit: `docs/research/SLICE_16_4_GATE1_PRESET_RACE_AUDIT_2026-09-16.md`
Gate-2 freeze: `docs/research/SLICE_16_4_GATE2_FREEZE_2026-09-16.md`

## Objective

Implement the frozen Slice-16.4 contract without broadening race mechanics beyond the supported Human/Klackon player set and fixed Darlok opponent baseline.

Gate 3 delivers four linked runtime pieces:

1. a server-owned 13-race preset catalog;
2. narrow Human/Klackon local-player New Game support with fixed Darlok opponent;
3. canonical Human/Klackon/Darlok portrait assets and manifest;
4. a shared image-led race carousel that consumes server availability/facts and submits the selected supported race.

Gate 4 remains separate.

## Implementation commits

### `6a7c0c2 feat: add authoritative preset race catalog`

Adds the server/runtime contract:

- `internal/game/race_catalog.go`;
- canonical 13-race catalog derived from loaded normalized race data;
- Human as `default_player_race_id`;
- Darlok as `fixed_opponent_race_id`;
- `supported` player availability only for Human and Klackon;
- `planned` availability for the remaining 11 entries;
- server-curated compact `card_fact_trait_ids`;
- `GET /api/v1/new-game/races`;
- Host catalog accessor;
- New Game validation widened only from Human-vs-Darlok to Human-or-Klackon vs fixed Darlok;
- server/game tests for catalog shape, supported/planned policy and crafted unsupported-race rejection.

The existing two-player topology remains unchanged. Darlok is still the fixed built-in-AI baseline; this does not promote Darlok to local-player support.

## Live authoritative race catalog

The final Gate-3 live endpoint returns:

```text
schema_version = 1
default_player_race_id = human
fixed_opponent_race_id = darlok
profiles = 13
```

Canonical order and availability:

```text
0  alkari     planned
1  bulrathi   planned
2  darlok     planned
3  elerian    planned
4  gnolam     planned
5  human      supported
6  klackon    supported
7  meklar     planned
8  mrrshan    planned
9  psilon     planned
10 sakkra     planned
11 silicoid   planned
12 trilarian  planned
```

Compact card facts remain server-owned and match the Gate-2 frozen normalized trait IDs. Examples:

```text
human   -> government_democracy
klackon -> government_unification, farming_plus_1, industry_plus_1, uncreative
darlok  -> spying_plus_20, stealthy_ships, government_dictatorship
```

The browser does not own support state or authoritative trait identifiers.

## Narrow New Game race validation

Gate 3 keeps the deterministic early topology intentionally narrow:

- exactly two players;
- strictly ascending unique non-zero seat IDs;
- unique race IDs;
- exactly one Darlok fixed opponent;
- the other race must be server-supported for local-player use: Human or Klackon;
- every submitted race must resolve through loaded race rules;
- planned races are rejected server-side even if a client bypasses the browser.

The existing Human-vs-Darlok path remains a non-regression fixture. Klackon-vs-Darlok is the new playable breadth case and is covered by deterministic tests.

## Canonical race portrait assets

### `daecb8c ui: add canonical race portrait assets`

Adds original MOOX runtime identity assets for the three Gate-3-required races:

```text
/assets/races/human/portrait.webp
/assets/races/klackon/portrait.webp
/assets/races/darlok/portrait.webp
```

Source/generation support:

- deterministic SVG source generation for the three current portraits;
- WebP runtime outputs;
- `/assets/races/manifest.json`;
- race asset validation script;
- build integration so asset correctness is checked as part of normal web validation.

Runtime WebP sizes at this implementation checkpoint are approximately:

```text
human   26.5 KB
klackon 33.8 KB
darlok  28.2 KB
```

The assets follow the frozen 4:5 race portrait contract and are original MOOX compositions. They do not copy/trace/repaint original MOO2 portrait pixels.

The committed portrait art remains replaceable under the stable semantic path/manifest contract if later visual review requests refinements.

## Web race carousel integration

### `df52e63 ui: integrate preset race carousel`

The New Game page now:

- loads `GET /api/v1/new-game/races`;
- renders all 13 canonical race identities in server order;
- uses the shared typed `VisualSelector` interaction grammar;
- exposes the selector under the semantic `player-race` setting identity;
- defaults to Human from the server catalog;
- marks Human and Klackon supported;
- marks the other 11 races planned;
- uses canonical Human/Klackon/Darlok WebP identities where available;
- uses a deliberately neutral planned placeholder for catalog-visible races whose final portrait is not yet curated, rather than reusing another race's identity;
- formats compact race-card facts from server-provided `card_fact_trait_ids`;
- keeps availability visible as text, not color-only state;
- disables Create Game while a planned race is selected;
- disables Create Game if the authoritative race catalog is unavailable;
- sends the selected supported local-player `race_id` in the New Game request;
- keeps seat 2 fixed to Darlok;
- changes the former Human-specific empire-name label to race-neutral Player Empire wording;
- adds English/German race names and compact trait labels.

## Live UI -> POST -> authoritative state proof

Managed-browser Gate-3 acceptance created:

```text
gate3-klackon-20260916-1017
```

The visible UI selected Klackon as the local-player race. The captured actual Create Game POST returned HTTP 201 with:

```text
seat 1 race_id = klackon
seat 2 race_id = darlok
```

The resulting authoritative seat-1 snapshot reported:

```text
race_id = klackon
```

This proves the complete chain:

```text
race portrait / carousel selection
  -> typed UI race ID
  -> Create Game wire payload
  -> server validation
  -> authoritative generated Empire race identity
```

## Browser acceptance

The reusable New Game Chrome smoke was extended from the previous selector set to cover the race carousel as well.

Current browseable New Game option count covered by the smoke is 25 across the bound selector families.

Race-specific checks include:

- all 13 race identities present in canonical server order;
- only Human/Klackon supported for local-player selection;
- planned races visibly marked and unable to enable Create Game;
- canonical Human/Klackon/Darlok WebP dimensions/load behavior;
- neutral planned placeholder behavior for races without final runtime portraits;
- Klackon compact facts sourced from the server catalog;
- mobile and desktop layout;
- mouse, real touch and keyboard navigation;
- race selector integration with the existing Difficulty / Galaxy Size / Galaxy Age flows.

## Final Gate-3 technical regression

Before formal documentation closeout, the dedicated final QA session passed:

- `go test ./... -count=1`;
- web production build;
- New Game static/deterministic asset checks including the race asset validator;
- complete New Game browser smoke with the race carousel;
- live `/api/v1/new-game/races` verification;
- `/healthz` via localhost;
- `/healthz` via NetBird `100.120.252.216:7171`;
- `/healthz` via LAN `192.168.5.27:7171`;
- `git diff --check`.

No implementation fix was required after that closeout run.

## Gate-3 checklist result

- [x] Produce/curate one original portrait for every supported race.
- [x] Add race asset manifest and optional emblems.
- [x] Expose authoritative preset race catalog data to the New Game HMI.
- [x] Implement left/right race carousel using Slice 16.1 grammar.
- [x] Reuse the same race portrait identity for opponent composition where possible.
- [x] Add deterministic New Game fixtures for each supported race/start tuple.

Notes on checklist interpretation:

- "every supported race" means the frozen Gate-2 supported local-player set Human/Klackon; Darlok also receives a canonical portrait because it is the fixed current opponent and that identity is intended for reuse.
- optional emblems remain optional and do not block Gate 3.
- general opponent composition remains Slice 16.6; Gate 3 establishes reusable Darlok identity now rather than implementing that later screen early.

## Result

Gate 3 is technically and formally complete. Slice 16.4 remains open for Gate 4 closure acceptance only.

Gate 4 is **not opened** by this document.
