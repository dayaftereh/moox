# Slice 16.3 Gate 4 - Galaxy visual selector close

Date: 2026-09-16
Status: **closed / accepted**

Gate-1 evidence: `docs/research/SLICE_16_3_GATE1_GALAXY_VISUAL_SELECTOR_AUDIT_2026-09-15.md`
Gate-2 freeze: `docs/research/SLICE_16_3_GATE2_GALAXY_VISUAL_SELECTOR_FREEZE_2026-09-15.md`
Gate-3 implementation: `docs/research/SLICE_16_3_GATE3_IMPLEMENTATION_2026-09-16.md`
Pre-Gate-4 user visual refinement: commit `32d879e`.

## User acceptance

Before Gate 4, the user explicitly accepted the refined Galaxy Age presentation, including the stronger Mineral Rich industrial/production language and Organic Rich biotech/pharma-inspired visual language, and requested Gate 4 closure.

This acceptance does not alter the frozen gameplay semantics: authoritative Galaxy Age behavior remains the server-owned mineral-resource / food-world bias contract.

## Equal-seed/settings deterministic acceptance

Gate 4 re-ran the permanent targeted generator tests rather than relying only on the earlier Gate-3 result.

`TestNewGameAllGalaxySizeAgeTuplesAreDeterministicAndBounded` passed for all 12 authoritative tuples:

```text
small  x mineral_rich / normal / organic_rich
medium x mineral_rich / normal / organic_rich
large  x mineral_rich / normal / organic_rich
huge   x mineral_rich / normal / organic_rich
```

For every tuple, the test proves:

- same seed + same complete settings -> byte-identical authoritative state;
- exact star-system count for the selected size;
- system coordinates stay inside the frozen row-major grid/jitter contract.

The targeted Gate-4 run also passed:

- `TestNewGameGalaxyAgeProfilesMatchFrozenGate2Data`;
- `TestGalaxyCatalogMatchesFrozenGate2Contract`;
- `TestHTTPGalaxyCatalogMatchesFrozenServerContract`.

Thus the selected Age IDs still resolve to the frozen spectral/climate profiles, including Organic Rich's audited `0x57CF` climate data.

## Live visual-selection -> wire-settings -> generator mapping

Gate 4 performed a real managed-Chrome New Game creation through the running external-review UI.

Chosen visible options:

```text
Galaxiegröße: Riesig
Galaxiealter: Organisch reich
Seed: 0x8009
Difficulty: Normal
```

The actual browser POST to `/api/v1/games` was captured at runtime and returned HTTP 201. Its authoritative settings body contained exactly:

```json
"galaxy_size": "huge",
"galaxy_age": "organic_rich"
```

No frontend-only alias was sent.

The resulting seat-1 authoritative snapshot was then fetched from the created game and contained exactly **71 systems**, matching the server catalog's `huge -> 71` contract. The permanent generator-profile tests above independently prove `organic_rich` selects the frozen Organic-Rich spectral/climate profile.

This closes the full mapping chain:

```text
visible selector label/art
  -> typed UI ID
  -> Create Game wire setting
  -> server validation/profile resolution
  -> deterministic authoritative generator result
```

## Final browser acceptance

The reusable real-Chrome New Game smoke passed after the final user-requested Age-art refinement.

Coverage includes:

- Difficulty + Galaxy Size + Galaxy Age;
- 12 bound authoritative options total;
- 390 px mobile layout with no horizontal overflow;
- real touch input;
- mouse interaction;
- ArrowLeft/ArrowRight/Home/End keyboard interaction;
- info-dialog focus and Escape return;
- desktop responsive layout;
- server-derived Size star-count facts;
- server-derived Age mineral/food bias facts;
- authoritative support state for all bound Size/Age options.

The accepted bound Galaxy set remains:

```text
Size: small / medium / large / huge
Age:  mineral_rich / normal / organic_rich
```

`tiny.svg` remains only a historical/prototype asset and is not a bound gameplay option.

## Final full regression

Gate 4 final closeout passed:

- `go test ./... -count=1`;
- `npm run build`;
- UTF-8/mojibake guard;
- deterministic 13-SVG New Game asset regeneration check;
- TypeScript + Vite production build;
- `npm run check:new-game-selector:browser -- http://127.0.0.1:7171`;
- `git diff --check`;
- `GET /healthz` via localhost;
- `GET /healthz` via NetBird `100.120.252.216:7171`;
- `GET /healthz` via LAN `192.168.5.27:7171`;
- live `GET /api/v1/new-game/galaxy` catalog verification.

The live Galaxy catalog remained:

```text
sizes: small=20, medium=36, large=54, huge=71
ages:
  mineral_rich -> mineral higher / food lower
  normal       -> baseline / baseline
  organic_rich -> mineral lower / food higher
```

## Gate-4 checklist result

- [x] Repeat equal-seed/settings generation and hash/equality acceptance.
- [x] Verify visual selection maps exactly to authoritative generator settings.
- [x] Desktop/mobile browser QA.
- [x] Full tests/build/diff checks.

## Exit criterion

**Met.** Galaxy Size and Galaxy Age can be selected through the shared image-led left/right selector grammar, their compact facts come from the authoritative server catalog, and the chosen settings produce the exact deterministic authoritative galaxy contract selected by the player.

Slice 16.3 is closed. Slice 16.4 is the next prepared Slice-16 sub-slice but is not opened by this closeout.
