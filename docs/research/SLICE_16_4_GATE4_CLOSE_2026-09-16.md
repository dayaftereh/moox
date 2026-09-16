# Slice 16.4 Gate 4 - Preset race portraits / carousel close

Date: 2026-09-16
Status: **closed / accepted**

Gate-1 evidence: `docs/research/SLICE_16_4_GATE1_PRESET_RACE_AUDIT_2026-09-16.md`
Gate-2 freeze: `docs/research/SLICE_16_4_GATE2_FREEZE_2026-09-16.md`
Gate-3 implementation: `docs/research/SLICE_16_4_GATE3_IMPLEMENTATION_2026-09-16.md`

## Scope accepted

Slice 16.4 closes with the bounded Gate-2 race contract intact:

- all 13 normalized preset race identities are server-owned catalog/browse entries;
- local-player support is `human` + `klackon`;
- `darlok` remains the fixed built-in-AI baseline opponent;
- all other preset races remain visible but `planned` for local-player selection;
- broader race-mechanics completion and Custom Race Designer remain explicitly reserved for Slice 22.

No Gate-4 work broadened the supported-race set or introduced deferred race mechanics.

## Portrait acceptance / provenance

Gate 4 re-verified the three canonical runtime portraits:

```text
/assets/races/human/portrait.webp
/assets/races/klackon/portrait.webp
/assets/races/darlok/portrait.webp
```

All three:

- decode at exactly **1200 x 1500** (4:5);
- have different SHA-256 hashes;
- have retained deterministic SVG source files under `docs/research/prototypes/SLICE_16_4_RUNTIME_PORTRAITS_2026-09-16/`;
- contain no embedded `<image>` source, external raster reference or data URI in those SVG sources;
- are generated from `web/scripts/generate-race-portrait-sources.mjs`;
- have manifest provenance `original-procedural-mox` and explicit source notes;
- are validated by `npm run check:race-assets`.

The source generator contains no original/LBX/raster portrait input path. Runtime portraits are therefore original MOOX procedural composition assets, not copied/traced original-game portrait files.

## Accessibility hardening found during Gate 4

Gate 4 identified one presentation gap: the selected option's `supported/planned` text was available in the info dialog and data attributes but was not visible directly below the selected name.

Commit `71e32c0` closes that gap:

- `VisualSelector` now renders the current availability label as visible text under the selected option;
- `supported` and `planned` remain distinguishable without relying on color/border style;
- browser QA temporarily hides the current race portrait and proves race name, availability text, selector ARIA label and info-button ARIA identity remain usable;
- planned Alkari explicitly renders `Geplant / gesperrt` while Create Game remains locked.

The race portrait itself stays `aria-hidden=true`; authoritative textual identity comes from selector text/ARIA, not from alt text embedded in artwork.

## Live visual -> wire -> authoritative mapping

Gate 4 repeated a real managed-Chrome end-to-end creation through the running New Game UI.

Visible selection:

```text
Spielerrasse: Klackon
Status: Jetzt unterstützt
Player empire: Klackon
```

The actual captured POST to `/api/v1/games` returned HTTP 201 and contained:

```json
"players": [
  {"seat_id":1,"empire_name":"Klackon","race_id":"klackon"},
  {"seat_id":2,"empire_name":"Darlok","race_id":"darlok"}
]
```

The resulting authoritative seat-1 snapshot was then fetched and reported:

```text
race_id = klackon
empire name = Klackon
```

This closes the full race-selection chain:

```text
portrait / race carousel
  -> typed race ID
  -> Create Game wire settings
  -> server validation
  -> authoritative Empire race_id
```

## Determinism / server contract acceptance

Targeted Gate-4 tests passed:

- `TestPresetRaceCatalogMatchesFrozenGate2Contract`;
- `TestNewGameAllowsKlackonAgainstDarlokDeterministically`;
- `TestHTTPRaceCatalogMatchesFrozenServerContract`.

The catalog remains exactly 13 profiles in normalized order, with only Human and Klackon marked player-supported and Darlok fixed as opponent.

Same seed + same complete Klackon/Darlok settings continues to produce byte-identical `NewGameResult` state.

## Desktop / mobile browser acceptance

The real Chrome New Game smoke passed after the accessibility hardening.

Coverage includes:

- Difficulty + Galaxy Size + Galaxy Age + Player Race;
- 25 browsable options total;
- all 13 race catalog entries in canonical order;
- Human and Klackon supported;
- all other local-player races planned/locked;
- canonical Human/Klackon/Darlok portrait paths where available;
- no borrowed portrait for planned races without one;
- 390 px mobile layout without horizontal overflow;
- 4:5 race artwork limited to the frozen mobile target;
- desktop 400 x 500 race artwork target;
- mouse, real touch and keyboard Home/End/Arrow navigation;
- >=44 px mobile and >=52 px desktop selector controls;
- dialog focus/close behavior;
- server-curated Klackon facts;
- visible Supported/Planned status;
- text/ARIA fallback with the race portrait hidden.

## Final full regression

Final Gate-4 closeout passed:

- `go test ./... -count=1`;
- `npm run build`;
- UTF-8/mojibake guard;
- deterministic New Game selector guard;
- deterministic race-source/WebP asset guard;
- TypeScript + Vite production build;
- real-Chrome New Game selector smoke via direct Node runner;
- `git diff --check`;
- `/healthz` via localhost;
- `/healthz` via NetBird `100.120.252.216:7171`;
- `/healthz` via LAN `192.168.5.27:7171`;
- live `GET /api/v1/new-game/races` verification.

## Gate-4 checklist result

- [x] Every supported race has a non-placeholder portrait.
- [x] No copied/traced original-game art remains in runtime assets.
- [x] Portrait family passes desktop and real/mobile-width visual review.
- [x] Text/ARIA remains sufficient with images disabled.
- [x] Equal seed/settings/race tuples remain deterministic.
- [x] Full tests/build/diff checks.

## Exit criterion

**Met.** Preset race selection now behaves like choosing a civilization rather than selecting a database row: the supported local-player races have original reusable MOOX portrait identities, the current Darlok opponent has a canonical reusable portrait, the authoritative server catalog controls support/facts, planned races remain honestly locked, and the selected supported race maps deterministically into the authoritative New Game state.

Slice 16.4 is closed. Slice 16.5 - Starting Technology visual selector - is next prepared but is not opened by this closeout.
