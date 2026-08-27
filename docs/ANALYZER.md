# MOOX MOO2 Analyzer

`moox-analyze` is a read-only, pure-Go console tool for researching a legally owned Master of Orion II installation.

It is deliberately independent of Wails and must remain buildable with `CGO_ENABLED=0`.
## Current status

Current analyzer version: **0.21.0**.

The tool has grown from the original LBX inspector into the main read-only 1.31 research/normalization pipeline. It now covers installation inventory, LBX inspection/extraction, graphics/palette extraction, audio/text/support catalogs, evidence-based block classification, original DOS bound-MZ/LE reads, and normalized race/building/technology/ship-hull/semantic-asset generation.

Current committed normalized outputs are summarized in `docs/PROJECT_STATUS.md`. The one active decoder ticket is tracked in `docs/research/ACTIVE_RESEARCH.md` rather than in a broad open-ended decoder list.

## Build

```powershell
go test ./...
go build -o out/moox-analyze.exe ./cmd/moox-analyze
```

## Commands

### Inventory an installation

```powershell
out\moox-analyze.exe inventory `
  -out reference\catalogs\moo2-1.31-inventory.json `
  C:\ASH\Temp\mastori2
```

The catalog contains:

- relative file paths and sizes,
- SHA-256 per file,
- detected outer format,
- LBX block count,
- LBX block offsets and sizes,
- optional SHA-256 per LBX block with `-entry-hashes`.

The current local 1.31 reference produces:

- 420 total files,
- 338,042,423 bytes,
- 363 normal SimTex LBX containers,
- 10 Smacker `SMK2` videos stored with an `.LBX` extension,
- 47 other support/executable/configuration files.

Generated catalogs live below `reference/catalogs/` and are ignored by Git.

### Inspect an LBX

```powershell
out\moox-analyze.exe inspect C:\ASH\Temp\mastori2\RACESTUF.LBX
```

`RACESTUF.LBX` in the local 1.31 reference is recognized as 14 blocks with its first payload at offset `0x800`.

Use `-json` for machine-readable output.

### Scan strings

```powershell
out\moox-analyze.exe strings -block 0 -min 5 C:\ASH\Temp\mastori2\RACESTUF.LBX
```

This already exposes useful source labels such as the visible race-design modifiers. The scanner is intentionally an ASCII research aid, not yet a locale-aware MOO2 text decoder.

### Extract a raw block

```powershell
out\moox-analyze.exe extract `
  -out reference\extracted\RACESTUF.block0.bin `
  C:\ASH\Temp\mastori2\RACESTUF.LBX 0
```

Existing output files are not overwritten unless `-force` is passed.

## Package layout

```text
cmd/moox-analyze/        CLI
internal/lbx/            generic SimTex LBX parsing/block IO
internal/catalog/        installation inventory/provenance
internal/textscan/       printable binary-string scanning
internal/i18n/           stable language-file loading/merge/validation
internal/moo2data/       MOO2 1.31-specific normalization/decoders
internal/ruleset/        normalized runtime schemas/loaders/validation
internal/moo2exe/        bound MZ/LE original-executable reader
internal/moo2gfx/        MOO2 graphics/frame decoder
internal/graphiccatalog/ graphics inventory/export
internal/palettecatalog/ external palette catalog/export
internal/audiocatalog/   lossless RIFF/WAVE discovery/export
internal/textcatalog/    structured/private text reference catalog
internal/blockcatalog/   evidence-based LBX block classification
internal/rawextract/     full private original-data extraction
internal/supportcatalog/ non-LBX support snapshot
```

The generic parsers remain game-format focused; MOO2-specific identities/formulas belong in `internal/moo2data` and always carry explicit evidence/provenance boundaries.
## Safety / data boundary

The analyzer never modifies the source installation. Raw extracted game data belongs under ignored `reference/` directories and is not committed. Only independently authored parsers, factual normalized data, tests and research notes belong in Git.

## Current normalized coverage and next decoder work

Already committed:

1. Race Designer options/localization architecture.
2. 13 preset races.
3. 203 technology identities/IDs.
4. 48 building IDs + original technology links.
5. Six military ship hull identities.
6. Semantic race/UI/building/strategic-ship/tactical-ship asset mappings.
7. Private graphics/palette/audio/text/support/block-classification reference tooling.

Not yet complete as gameplay data:

- planet classes (active ticket),
- technology research fields/costs/effects,
- building costs/maintenance/effects,
- ship components/weapons/specials/full design rules,
- leaders/officers,
- diplomacy/events,
- save-game fixtures and runtime parity scenarios.

Do not treat this list as permission for broad parallel research. `docs/research/ACTIVE_RESEARCH.md` is the authoritative single next decoder/research objective.
## Normalize loadable ruleset data

The analyzer now has a normalization layer in addition to raw binary inspection.

```powershell
out\moox-analyze-windows-amd64.exe normalize race-traits `
  -out data\rulesets\moo2-1.31\race_traits.json `
  C:\ASH\Temp\mastori2
```

This command validates the expected English `RACESTUF.LBX` block layout before generating data. A mismatched archive/string order fails instead of silently producing a wrong ruleset.

The generated ruleset is runtime-oriented: stable IDs and numeric values are authoritative, while user-visible text is referenced by language keys. `-language-out` generates a separate language JSON (currently English from the original 1.31 block). Provenance deliberately distinguishes direct 1.31 observations from secondary-reference Pick costs and still-unverified behavioral formulas.





## Normalize ship hulls

```powershell
out\moox-analyze-windows-amd64.exe normalize ship-hulls `
  -out data\rulesets\moo2-1.31\ship_hulls.json `
  -languages-dir data\languages `
  C:\ASH\Temp\mastori2
```

The decoder combines the original six-name `TECHNAME.LBX` hull sequence with a SHA-256-verified `Auto_Design_Ship_` code range from `Orion2.exe`. Sizes 0..4 use eight picture IDs each (`size*8 + style`); size 5 uses picture ID 43. The ruleset now exposes separate strategic and tactical semantic asset keys. Tactical frame semantics are validated separately from the hull-name/picture-ID identity layer.
## Normalize technologies

```powershell
out\moox-analyze-windows-amd64.exe normalize technologies `
  -out data\rulesets\moo2-1.31\technologies.json `
  -languages-dir data\languages `
  C:\ASH\Temp\mastori2
```

The decoder identifies the bounded 203-name technology section directly in English `TECHNAME.LBX` block 0 (`No Tech` -> Achilles... through Zortrium Armor -> next `Biology` section). The resulting IDs 1..203 are therefore original-observed, not imported from a secondary tech table.
## Normalize colony buildings

```powershell
out\moox-analyze-windows-amd64.exe normalize buildings `
  -out data\rulesets\moo2-1.31\buildings.json `
  -languages-dir data\languages `
  C:\ASH\Temp\mastori2
```

This dataset contains 48 stable building IDs. Names are verified against original `HELP.LBX` / `TECHNAME.LBX`. Building IDs and technology IDs are read directly from the original `Orion2.exe` `_buildings` table through the pure-Go bound MZ/LE reader; the normalizer cross-checks each table technology ID against the normalized 203-entry original technology catalog. `Hydroponic Farms` is the original singular-name alias to tech 87 `Hydroponic Farm`; building ID 48 directly references tech 16 `Planet Construction`, while its display-name identity `Artificial Planet` remains explicitly distinguished.
## Normalize semantic assets

Generate tracked semantic asset metadata from the private original reference and normalized race order:

```powershell
out\moox-analyze-windows-amd64.exe normalize assets `
  -out data\rulesets\moo2-1.31\assets.json `
  -races data\rulesets\moo2-1.31\races.json `
  -buildings data\rulesets\moo2-1.31\buildings.json `
  -ship-hulls data\rulesets\moo2-1.31\ship_hulls.json `
  C:\ASH\Temp\mastori2
```

The semantic catalog validates the `RACESEL` portrait sequence and 13x13 `RACEICON` matrix, selected UI/production mappings, and all 48 standard building colony sets. Building graphics are generated only after `Orion2.exe` function hashes confirm the original `(building_id-1)/10`, `%10*36`, and 6x6 serpentine-frame formulas; every one of the 1,728 building variants is then revalidated against the local BLDG block dimensions and hash. Generic `race.<id>.icon` keys remain `pending` rather than arbitrarily selecting one of the 13 role/context variants. The catalog also contains six military hull strategic sets and Colony/Outpost/Transport sets: 352 `SHIPS.LBX` references generated from hash-verified original color/picture-ID logic and revalidated as one-frame 52-pixel strategic graphics. It additionally contains six tactical hull sets: `Load_Combat_Ship_` proves the shared picture ID, `CMBTSHP.LBX` uses `color*45+picture`, and the hash-verified draw paths establish 20 frames as 5 folded orientations x 4 animation phases, yielding 6,560 tactical references.

The generated `assets.json` is distributable metadata only; original graphics remain below ignored `reference/original/` paths.
## Decode graphics

Graphics with embedded/internal palettes can now be exported directly to PNG:

```powershell
out\moox-analyze-windows-amd64.exe image `
  -out reference\original\images\racesel\block_015.png `
  C:\ASH\Temp\mastori2\RACESEL.LBX 15
```

The pure-Go decoder handles image headers, internal palette shift/count, frame offsets, sparse pixel sequences, raw `NoCompression` frames and cumulative `Junction` animations. The single-image command still requires either an internal palette or an explicitly supplied external palette, so it never guesses colors. See `docs/research/MOO2_GRAPHICS.md`.
## Batch graphics catalog/export

Catalog every structurally recognized MOO2 graphic block and export every frame with sufficient palette context. Internal palettes are used directly; externally paletted blocks are exported only when the resolver has a confirmed context rule:

```powershell
out\moox-analyze-windows-amd64.exe graphics `
  -clean `
  -out reference\original\images `
  C:\ASH\Temp\mastori2
```

Use `-manifest-only` to inventory graphics without writing PNGs. Manifest schema v2 records archive/block/frame provenance, dimensions, flags, palette mode, hashes, per-frame decode status and palette-context status/source/evidence/confidence. The evidence-gated resolver now covers verified archive-wide and mixed contexts including `BLDG0..4`, `COUNCIL`, `CMBTMISL`, `SHIPS`, `RACEICON`, `DESIGN`, and `MAINMENU`. Each resolved block records its palette source, evidence and confidence; all other unresolved external/mixed contexts remain `external_palette_pending` rather than being color-guessed. See `docs/research/MOO2_PALETTES.md`.
## External palette extraction

Export the 13 `FONTS.LBX` and 4 `IFONTS.LBX` external palettes:

```powershell
out\moox-analyze-windows-amd64.exe palettes `
  -clean `
  -out reference\original\palettes `
  C:\ASH\Temp\mastori2
```

The `image` command can also be given `-palette-file` and `-palette-block` to decode a graphic with an explicitly known external palette. Mixed palettes preserve internal entries and fill only missing indices from the external base palette.
## Full raw LBX extraction

Extract every block from every normal LBX archive and copy `.LBX`-named Smacker movies as `.smk`:

```powershell
out\moox-analyze-windows-amd64.exe unpack `
  -clean `
  -out reference\original\unpacked `
  C:\ASH\Temp\mastori2
```

The private manifest keeps archive/block offsets, sizes and SHA-256 hashes so every later decoder can work from a stable raw block path while the source installation remains untouched.
## Evidence-based block classification

Classify every normal LBX block without assigning unsupported semantics:

```powershell
out\moox-analyze-windows-amd64.exe classify `
  -out reference\catalogs\block-classification.json `
  C:\ASH\Temp\mastori2
```

Current classes include graphics, known external palettes, RIFF/WAVE, structurally exact `fixed_record_v1` records, conservative string-table candidates, empty blocks and unknowns. See `docs/research/BLOCK_CLASSIFICATION_2026-08-26.md`.
## Lossless WAVE extraction

Copy every structurally verified RIFF/WAVE LBX block without transcoding:

```powershell
out\moox-analyze-windows-amd64.exe audio `
  -clean `
  -out reference\original\audio `
  C:\ASH\Temp\mastori2
```

The private manifest records WAVE/PCM metadata and hashes. See `docs/research/AUDIO_EXTRACTION_2026-08-26.md`.
## Private text catalog

Catalog fixed-size text/message records and strong string-table candidates while keeping high bytes undecoded:

```powershell
out\moox-analyze-windows-amd64.exe text `
  -clean `
  -out reference\original\text `
  C:\ASH\Temp\mastori2
```

ASCII previews are research aids only; raw bytes remain authoritative. Schema v2 additionally preserves fixed-array record index/offset/size/hash boundaries for HELP, TECHDESC, leader/name tables and related data. See `docs/research/TEXT_EXTRACTION_2026-08-26.md`.
## Non-LBX support snapshot

Copy and hash every non-LBX file while preserving relative paths:

```powershell
out\moox-analyze-windows-amd64.exe support `
  -clean `
  -out reference\original\support `
  C:\ASH\Temp\mastori2
```

This captures executables, audio drivers/configuration, README and the private save-game fixture without duplicating the LBX files handled by the other analyzers.
## Race Designer locales

`normalize race-traits -languages-dir data\languages` now writes EN/DE/FR/ES/IT language files with identical stable keys. The locale-specific MOO2 glyph substitutions are decoded to Unicode only where the byte-to-glyph mapping is evidenced by parallel source words. See `docs/research/LOCALIZATION_LAYOUT_2026-08-26.md`.
## Normalize preset races

Build the 13 standard-race definitions after `race_traits.json` exists:

```powershell
out\moox-analyze-windows-amd64.exe normalize races `
  -out data\rulesets\moo2-1.31\races.json `
  -race-traits data\rulesets\moo2-1.31\race_traits.json `
  -languages-dir data\languages `
  C:\ASH\Temp\mastori2
```

This verifies HELP record hashes, maps the documented preset characteristics to stable Race Designer trait IDs, calculates Pick totals from original-observed costs, and merges the canonical singular race names from `ESTRINGS.LBX` into English localization. It deliberately does not assign race artwork yet.