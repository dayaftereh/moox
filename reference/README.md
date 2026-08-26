# Local reference material

This directory is for **private development references** from a legally owned Master of Orion II installation.

Do not commit original game binaries, LBX archives, screenshots, music, video, manuals or extracted copyrighted assets unless redistribution rights have been explicitly confirmed.

Recommended local-only folders:

- `original/` - canonical local reference material derived 1:1 from the owned original. Decoded original images belong under `original/images/` with source archive/block/frame preserved.
- `extracted/` - temporary/intermediate decoder output that has not yet been promoted to canonical original reference material.
- `screenshots/` - locally captured UI comparison images.
- `catalogs/` - inventories/hashes that may contain original filenames or extracted text.`r`n- `original/palettes/` - decoded external palette JSON/swatches and provenance, local only.

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

`manifest.json` records source archive, block, frame, dimensions, palette/decode status and SHA-256 so a later MOOX asset can always be traced back to its visual reference. The current batch contains 10,646 PNG frames; this private tree is intentionally not versioned.

MOOX-owned observations and normalized factual rule data belong in `docs/` and `data/rulesets/`. New independently created artwork belongs under `assets/`, never under `reference/original/`.
