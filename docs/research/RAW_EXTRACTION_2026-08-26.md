# Full raw LBX extraction - checkpoint 2026-08-26

Source: local official Master of Orion II 1.31 reference at `C:\ASH\Temp\mastori2`.

## Result

`moox-analyze unpack` created a complete private raw decomposition at:

```text
reference/original/unpacked/
```

The current checkpoint contains:

- 363 normal SimTex LBX archives,
- 10,498 extracted LBX blocks,
- 278.02 MiB LBX payload bytes,
- 10 `.LBX`-named Smacker videos copied as `.smk`,
- 316.70 MiB total copied payload including the Smacker files,
- 10,509 files including the manifest,
- approximately 2.44 MiB `manifest.json`.

## Layout

```text
reference/original/unpacked/
  manifest.json
  lbx/
    racestuf/
      block_0000.bin
      ...
    science/
    ships/
    ...
  smacker/
    INTRO.smk
    ...
```

Every LBX block manifest record contains:

- source archive,
- archive SHA-256,
- block index,
- original offset,
- block size,
- block SHA-256,
- extracted relative path.

Every copied Smacker record contains source filename, size, SHA-256 and output path.

## Verification

A sample end-to-end hash check for `RACESTUF.LBX` block 0 matches exactly:

```text
5583d8d5744e3a075b2e925282ab7f98f0c40418e3c029e82718136ffd665970
```

All ten copied Smacker outputs retain the `SMK2` file signature.

## Why keep raw blocks

The LBX container itself does not reliably tag the type of each contained block. A block may be graphics, text, structured rules, palettes, fonts, sounds or some other game-specific binary structure. Keeping the complete raw decomposition gives later decoders a stable source path without repeatedly unpacking archives or modifying the owned installation.

The raw files are original copyrighted game data and remain under the ignored `reference/original/` tree. They are not committed.

## Next analysis layer

The next pass should classify the 10,498 raw blocks into recognizable families while retaining `unknown` for anything not proven:

- graphic,
- known palette,
- text/string table,
- sound/music signatures,
- font/cursor candidates,
- structured-data candidates,
- unknown.

This classification must be evidence-based; no block should be assigned a semantic type solely from its archive filename.
