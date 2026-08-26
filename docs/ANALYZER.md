# MOOX MOO2 Analyzer

`moox-analyze` is a read-only, pure-Go console tool for researching a legally owned Master of Orion II installation.

It is deliberately independent of Wails and must remain buildable with `CGO_ENABLED=0`.

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
cmd/moox-analyze/     CLI only
internal/lbx/         generic SimTex LBX parsing and block IO
internal/catalog/     installation inventory/provenance catalog
internal/textscan/    generic binary string scanning
```

The next layer will be `internal/moo2data/`, containing archive-specific decoders for MOO2 1.31. The generic LBX parser should not gain game-specific knowledge.

## Safety / data boundary

The analyzer never modifies the source installation. Raw extracted game data belongs under ignored `reference/` directories and is not committed. Only independently authored parsers, factual normalized data, tests and research notes belong in Git.

## Planned decoder order

1. Race-design strings and data (`RACESTUF`, `RACEOPT`, related race archives).
2. Technologies/research (`SCIENCE` plus relevant structured tables/strings).
3. Colony buildings and economy data.
4. Ship hulls, weapons and specials.
5. Leaders/officers.
6. Planet/system data.
7. Events/diplomacy.
8. Save-game fixtures.
9. Graphics/palettes/audio only as private development reference tooling.

## Normalize loadable ruleset data

The analyzer now has a normalization layer in addition to raw binary inspection.

```powershell
out\moox-analyze-windows-amd64.exe normalize race-traits `
  -out data\rulesets\moo2-1.31\race_traits.json `
  C:\ASH\Temp\mastori2
```

This command validates the expected English `RACESTUF.LBX` block layout before generating data. A mismatched archive/string order fails instead of silently producing a wrong ruleset.

The generated ruleset is runtime-oriented: stable IDs and numeric values are authoritative, while user-visible text is referenced by language keys. `-language-out` generates a separate language JSON (currently English from the original 1.31 block). Provenance deliberately distinguishes direct 1.31 observations from secondary-reference Pick costs and still-unverified behavioral formulas.

## Decode graphics

Graphics with embedded/internal palettes can now be exported directly to PNG:

```powershell
out\moox-analyze-windows-amd64.exe image `
  -out reference\original\images\racesel\block_015.png `
  C:\ASH\Temp\mastori2\RACESEL.LBX 15
```

The pure-Go decoder handles image headers, internal palette shift/count, frame offsets and sparse pixel sequences. Graphics that rely only on external palettes are rejected for now so the analyzer does not emit knowingly wrong colors. See `docs/research/MOO2_GRAPHICS.md`.
## Batch graphics catalog/export

Catalog every structurally recognized MOO2 graphic block and export every frame whose block carries an internal palette:

```powershell
out\moox-analyze-windows-amd64.exe graphics `
  -clean `
  -out reference\original\images `
  C:\ASH\Temp\mastori2
```

Use `-manifest-only` to inventory graphics without writing PNGs. The generated private `manifest.json` records archive/block/frame provenance, dimensions, flags, palette mode, hashes and per-frame decode status. Graphics that require an external palette are cataloged as `external_palette_pending` rather than being color-guessed.
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