# Local reference material

This directory is for **private development references** from a legally owned Master of Orion II installation.

Do not commit original game binaries, LBX archives, screenshots, music, video, manuals or extracted copyrighted assets unless redistribution rights have been explicitly confirmed.

Recommended local-only folders:

- `original/` - canonical local reference material derived 1:1 from the owned original. Decoded original images belong under `original/images/` with source archive/block/frame preserved.
- `original/unpacked/` - complete raw LBX block decomposition plus renamed Smacker files and provenance manifest.
- `extracted/` - temporary/intermediate decoder output that has not yet been promoted to canonical original reference material.
- `screenshots/` - locally captured UI comparison images.
- `catalogs/` - inventories/hashes that may contain original filenames or extracted text.
- `original/palettes/` - decoded external palette JSON/swatches and provenance, local only.
- `original/audio/` - 1:1 RIFF/WAVE block copies and audio metadata, local only.
- `original/text/` - structural text catalog and ASCII-only previews; raw bytes remain in `original/unpacked/`.
- `original/support/` - all non-LBX installation files copied 1:1 with path/hash provenance, local only.

Current original-image layout:

```text
reference/original/images/
  manifest.json
  racesel/
    block_015.png
    ...
  planets/
    block_000.png
    ...
```

`manifest.json` records source archive, block, frame, dimensions, palette/decode status and SHA-256 so a later MOOX asset can always be traced back to its visual reference. The current batch contains 15,145 PNG frames (~447.28 MiB), including 4,490 frames whose confirmed external or mixed palette context is resolved automatically. It has zero frame decode failures. This private tree is intentionally not versioned.

MOOX-owned observations and normalized factual rule data belong in `docs/` and `data/rulesets/`. New independently created artwork belongs under `assets/`, never under `reference/original/`.
